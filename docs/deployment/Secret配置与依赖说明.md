# K8sOperation Secret 配置与依赖关系说明

> 本文档说明 `deploy/backend/secret.yaml` 中各配置项的用途、是否为强依赖，以及 ServiceAccount 自动创建机制。

---

## 一、ServiceAccount（SA）是否需要手动创建？

**不需要手动创建。**

`deploy/backend/service.yaml` 中已定义完整的 RBAC 资源，执行 `kubectl apply -k deploy/` 时会自动创建：

| 资源类型 | 名称 | 说明 |
|----------|------|------|
| ServiceAccount | `k8soperation` | 后端 Pod 使用的服务账号 |
| ClusterRole | `k8soperation` | 跨命名空间操作 K8s 资源的权限规则 |
| ClusterRoleBinding | `k8soperation` | 将 SA 绑定到 ClusterRole |

部署命令 `kubectl apply -k deploy/` 会按 kustomization.yaml 中的顺序依次创建所有资源，SA 和 RBAC 会自动就绪。

---

## 二、Secret 配置项总览

文件路径：`deploy/backend/secret.yaml`

```yaml
stringData:
  DB_PASSWORD: "123456"                              # MySQL 密码
  REDIS_PASSWORD: "123456"                           # Redis 密码
  JWT_SIGNING_KEY: "eoNB0%bv5M7995F1"               # JWT 签名密钥
  JENKINS_URL: "http://8.211.45.178:30080/"          # Jenkins 地址
  JENKINS_USERNAME: "devops"                         # Jenkins 用户名
  JENKINS_API_TOKEN: "1140e22ae3b1db1af5242b72213a530acb"  # Jenkins API Token
  HMAC_SECRET: "changeme"                            # Jenkins 回调签名
  KUBECONFIG_ENCRYPT_KEY: "changemp"                 # KubeConfig 加密密钥
  CACHE_SECRET: "k8smana"                            # Session 加密密钥
  DINGTALK_WEBHOOK: ""                               # 钉钉通知（可选）
  PLATFORM_FRONTEND_URL: "http://localhost:5173"     # 前端地址（可选）
```

---

## 三、依赖级别分类

### 3.1 强依赖（缺失 → Pod 无法启动）

如果 Secret 中缺少以下 key，Pod 会卡在 `CreateContainerConfigError` 状态，**完全无法启动**：

| Key | 用途 | 写错的后果 |
|-----|------|-----------|
| `DB_PASSWORD` | MySQL 连接密码 | 连接 MySQL 失败 → 应用崩溃退出 → **CrashLoopBackOff** |
| `REDIS_PASSWORD` | Redis 连接密码 | 连接 Redis 失败 → 应用崩溃退出 → **CrashLoopBackOff** |
| `JWT_SIGNING_KEY` | JWT Token 签名 | 应用能启动，但所有用户 Token 失效，无法登录 |
| `JENKINS_URL` | Jenkins 服务地址 | 应用能启动，CI/CD 流水线功能不可用 |
| `JENKINS_USERNAME` | Jenkins 用户名 | 同上 |
| `JENKINS_API_TOKEN` | Jenkins API 令牌 | 同上 |
| `HMAC_SECRET` | Webhook 签名验证 | 应用能启动，Jenkins 回调验证失败 |
| `KUBECONFIG_ENCRYPT_KEY` | AES-256 加密密钥 | 应用能启动，已保存的集群配置解密失败 |

> **注意**：以上 key 在 deployment.yaml 中**没有** `optional: true` 标记，K8s 会强制要求 Secret 中必须存在这些 key，否则 Pod 调度后创建容器时直接失败。

### 3.2 弱依赖（缺失 → Pod 正常启动）

以下 key 标记了 `optional: true`，即使不存在也不影响 Pod 启动：

| Key | 用途 | 缺失影响 |
|-----|------|---------|
| `DINGTALK_WEBHOOK` | 钉钉通知 Webhook URL | 构建/发布通知不会推送到钉钉 |
| `PLATFORM_FRONTEND_URL` | 前端公网地址 | 钉钉消息中的跳转链接为空 |

### 3.3 仅通过 ConfigMap 引用（不在 env 中）

以下 key 存在于 Secret 中，但通过 configmap.yaml 的 `${VAR}` 占位符引用：

| Key | 用途 | 说明 |
|-----|------|------|
| `CACHE_SECRET` | Session 加密密钥 | 通过 configmap 中 `Secret: "${CACHE_SECRET}"` 引用 |

---

## 四、依赖关系图

```
Pod 启动流程：
┌─────────────────────────────────────────────────────┐
│ 1. K8s 调度 Pod                                      │
│ 2. 检查 Secret key 是否存在                           │
│    ├── 缺少非 optional key → CreateContainerConfigError │
│    └── 全部存在 → 继续                                │
│ 3. 创建容器，注入环境变量                              │
│ 4. 应用启动                                           │
│    ├── 连接 MySQL (DB_PASSWORD)                       │
│    │   └── 失败 → 退出 → CrashLoopBackOff            │
│    ├── 连接 Redis (REDIS_PASSWORD)                    │
│    │   └── 失败 → 退出 → CrashLoopBackOff            │
│    └── 两者都成功 → 应用就绪 ✅                        │
│ 5. 健康检查探针                                       │
│    ├── startupProbe: /healthz/live (最多等 150s)      │
│    ├── readinessProbe: /healthz/ready (检查 DB)       │
│    └── livenessProbe: /healthz/live (每 30s)          │
└─────────────────────────────────────────────────────┘
```

---

## 五、实际影响分级

| 级别 | 配置项 | 写错后果 | 修复方式 |
|:----:|--------|---------|---------|
| 🔴 致命 | `DB_PASSWORD` | Pod CrashLoop，完全不可用 | 改正后重新 apply Secret |
| 🔴 致命 | `REDIS_PASSWORD` | Pod CrashLoop，完全不可用 | 同上 |
| 🟡 严重 | `JWT_SIGNING_KEY` | 所有用户无法登录 | 改正后重启 Pod |
| 🟡 严重 | `KUBECONFIG_ENCRYPT_KEY` | 已有集群配置不可用 | 改正后重启 Pod（需与加密时一致） |
| 🟢 一般 | `JENKINS_*` | CI/CD 功能不可用 | 改正后重启 Pod |
| 🟢 一般 | `HMAC_SECRET` | Jenkins 回调失败 | 改正后重启 Pod |
| ⚪ 可选 | `DINGTALK_WEBHOOK` | 无通知 | 随时补填 |
| ⚪ 可选 | `PLATFORM_FRONTEND_URL` | 通知无跳转链接 | 随时补填 |

---

## 六、修复 Secret 的方法

```bash
# 方法 1：直接编辑（推荐）
kubectl edit secret k8soperation-secret -n k8soperation

# 方法 2：删除重建
kubectl delete secret k8soperation-secret -n k8soperation
kubectl apply -f deploy/backend/secret.yaml

# 方法 3：用 patch 更新单个值
kubectl patch secret k8soperation-secret -n k8soperation \
  -p '{"stringData":{"DB_PASSWORD":"new-password"}}'

# 修改 Secret 后需要重启 Pod 才能生效
kubectl rollout restart deployment/k8soperation -n k8soperation
```

---

## 七、各配置项的获取方式

| 配置项 | 值从哪里来 | 获取/生成方法 |
|--------|-----------|-------------|
| `DB_PASSWORD` | MySQL 数据库密码 | 你部署 MySQL 时设置的 root 密码（当前：`123456`） |
| `REDIS_PASSWORD` | Redis 密码 | 你部署 Redis 时设置的 `requirepass`（当前：`123456`） |
| `JWT_SIGNING_KEY` | 自定义密钥 | 自己生成：`openssl rand -base64 16` |
| `JENKINS_URL` | Jenkins 服务地址 | 浏览器打开 Jenkins 首页的 URL（含端口） |
| `JENKINS_USERNAME` | Jenkins 账号 | Jenkins 管理后台的登录用户名 |
| `JENKINS_API_TOKEN` | Jenkins API 令牌 | Jenkins → 右上角用户名 → Configure → API Token → Add new Token |
| `HMAC_SECRET` | 自定义密钥 | 自己生成：`openssl rand -hex 16` |
| `KUBECONFIG_ENCRYPT_KEY` | AES-256 密钥（32位） | 自己生成：`openssl rand -hex 16`（输出 32 字符） |
| `CACHE_SECRET` | 自定义密钥 | 自己定义一个字符串，用于 Session 加密 |
| `DINGTALK_WEBHOOK` | 钉钉机器人 | 钉钉群 → 群设置 → 智能群助手 → 添加机器人 → 复制 Webhook URL |
| `PLATFORM_FRONTEND_URL` | 前端访问地址 | 部署完成后的前端 URL，如 `http://1.117.227.207:30081` |

### 快速生成密钥命令

```bash
# JWT 签名密钥（16 字节 Base64）
openssl rand -base64 16
# 输出示例：eoNB0%bv5M7995F1

# HMAC / KUBECONFIG_ENCRYPT_KEY（32 字符 Hex）
openssl rand -hex 16
# 输出示例：a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6
```

### Jenkins API Token 获取步骤

1. 登录 Jenkins（`http://8.211.45.178:30080/`）
2. 点击右上角用户名 → 「Configure」
3. 找到「API Token」区域
4. 点击「Add new Token」→ 输入名称 → 「Generate」
5. 复制生成的 Token 填入 `JENKINS_API_TOKEN`

### 钉钉 Webhook 获取步骤（可选）

1. 打开钉钉群 → 群设置 → 智能群助手
2. 添加机器人 → 选择「自定义」
3. 安全设置选「加签」或「关键词」
4. 复制 Webhook URL 填入 `DINGTALK_WEBHOOK`

---

## 八、当前配置确认

| 配置项 | 当前值 | 状态 |
|--------|--------|:----:|
| DB_PASSWORD | `123456` | ✅ 正确（匹配 1.117.227.207 MySQL） |
| REDIS_PASSWORD | `123456` | ✅ 正确（匹配 1.117.227.207 Redis） |
| JWT_SIGNING_KEY | `eoNB0%bv5M7995F1` | ✅ 可用 |
| JENKINS_URL | `http://8.211.45.178:30080/` | ⚠️ 确认 Jenkins 可达 |
| JENKINS_USERNAME | `devops` | ⚠️ 确认用户名正确 |
| JENKINS_API_TOKEN | `1140e22ae3b1db1af5242b72213a530acb` | ⚠️ 确认 Token 有效 |
| HMAC_SECRET | `changeme` | ⚠️ 建议生产环境修改 |
| KUBECONFIG_ENCRYPT_KEY | `changemp` | ⚠️ 建议改为 32 位随机字符串 |
| CACHE_SECRET | `k8smana` | ✅ 可用 |
| DINGTALK_WEBHOOK | 空 | ✅ 可选，无影响 |
| PLATFORM_FRONTEND_URL | `http://localhost:5173` | ⚠️ 部署后改为 `http://1.117.227.207:30081` |
