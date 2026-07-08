# 项目凭证与 Token 配置说明

本文档汇总 K8sOperation 平台所有敏感凭证、Token 和密钥的配置位置及用途，方便部署和排查认证问题。

---

## 一、凭证总览表

| 类别 | 名称 | 配置文件/位置 | 作用 | 备注 |
|---|---|---|---|---|
| **Jenkins** | APIToken | `configs/config.yaml` → `Jenkins.APIToken` | 后端调用 Jenkins API 触发构建、获取日志 | **必须去 Jenkins UI 手动生成** |
| **Jenkins** | 登录密码 | `deploy/jenkins/secret.yaml` → `admin-password` | `ops-dev` 用户登录 Jenkins | 默认 `admin123` |
| **Jenkins** | HMAC Secret | `deploy/jenkins/secret.yaml` → `hmac-secret`<br>`configs/config.yaml` → `Jenkins.HMACSecret` | Jenkins 回调平台时做签名验证 | 两边必须保持一致 |
| **镜像仓库** | Registry 用户名 | `deploy/jenkins/secret.yaml` → `registry-username` | Jenkins 推送镜像时认证 | |
| **镜像仓库** | Registry 密码 | `deploy/jenkins/secret.yaml` → `registry-password` | Jenkins 推送镜像时认证 | |
| **Git** | Gitee 用户名 | `deploy/jenkins/secret.yaml` → `gitee-username` | Jenkins 拉取 Gitee 代码 | 固定为 `oauth2` |
| **Git** | Gitee Token | `deploy/jenkins/secret.yaml` → `gitee-password` | Jenkins 拉取 Gitee 代码 | Gitee Personal Access Token |
| **Maven** | Maven 用户名 | `deploy/jenkins/secret.yaml` → `maven-repo-username` | 拉取私有 Maven 依赖 | |
| **Maven** | Maven 密码 | `deploy/jenkins/secret.yaml` → `maven-repo-password` | 拉取私有 Maven 依赖 | 默认占位，需替换 |
| **平台安全** | JWT 签名密钥 | `configs/config.yaml` → `App.JWTSigningKey` | 用户登录后 Token 签名 | |
| **平台安全** | KubeConfig 加密密钥 | `configs/config.yaml` → `Security.KubeConfigEncryptKey` | 加密数据库中的 kubeconfig | |
| **数据库** | MySQL 密码 | `configs/config.yaml` → `Database.Password` | 连接 MySQL | |
| **缓存** | Redis 密码 | `configs/config.yaml` → `Cache.Password` | 连接 Redis | |
| **可选功能** | SonarQube Token | `deploy/jenkins/secret.yaml` → `sonarqube-token` | 代码质量扫描 | |
| **可选功能** | AI API Key | `configs/config.yaml` → `AIAssistant.APIKey` | AI 助手调用大模型 | |
| **通知** | 钉钉 Webhook | `configs/config.yaml` → `Jenkins.DingTalkWebhook` | 构建结果通知 | |
| **通知** | 飞书 Webhook | `configs/config.yaml` → `Jenkins.FeishuWebhook` | 构建结果通知 | |

---

## 二、Jenkins 认证配置详解

### 2.1 后端连接 Jenkins 需要的凭证

后端服务通过 `configs/config.yaml` 中的 `Jenkins` 配置块连接 Jenkins：

```yaml
Jenkins:
  URL: "http://127.0.0.1:30080/"
  Username: "ops-dev"
  APIToken: "这里填写 Jenkins 生成的 API Token"
```

- **URL**：Jenkins 服务地址，本地开发通常为 `http://127.0.0.1:30080/`
- **Username**：Jenkins 用户名，JCasC 默认创建为 `ops-dev`
- **APIToken**：Jenkins 用户 API Token，**不是登录密码**

### 2.2 如何获取 Jenkins API Token

1. 使用浏览器登录 Jenkins：`http://127.0.0.1:30080/`
   - 用户名：`ops-dev`
   - 密码：`admin123`（如未修改，见 `deploy/jenkins/secret.yaml` 中的 `admin-password`）

2. 点击右上角用户名 → **设置（Configure）**

3. 找到 **API Token** 区域，点击 **Add new Token**

4. 输入 Token 名称，例如 `k8s-operation`，点击 **Generate**

5. 复制生成的 Token 字符串，粘贴到 `configs/config.yaml` 的 `Jenkins.APIToken` 中

6. 重启后端服务使配置生效

### 2.3 快速验证 Token 是否有效

```bash
# 替换为你的 Token
curl -u ops-dev:你的Token http://127.0.0.1:30080/api/json
```

- 返回 JSON：Token 有效
- 返回 401：Token 错误或用户无权限

---

## 三、Jenkins 部署时注入的凭证

Jenkins 通过 JCasC（Configuration as Code）和 K8s Secret 注入初始用户和全局凭证。

### 3.1 初始用户配置

文件：`deploy/jenkins/configmap.yaml`

```yaml
securityRealm:
  local:
    allowsSignup: false
    users:
      - id: "ops-dev"
        password: "${JENKINS_ADMIN_PASSWORD}"

authorizationStrategy:
  loggedInUsersCanDoAnything:
    allowAnonymousRead: false
```

- 用户 `ops-dev` 的密码来自 Secret：`deploy/jenkins/secret.yaml` → `admin-password`
- 登录用户拥有所有权限

### 3.2 全局凭证配置

文件：`deploy/jenkins/configmap.yaml`

| 凭证 ID | 类型 | 用途 |
|---|---|---|
| `hmac-secret` | Secret text | 平台回调签名密钥 |
| `sonarqube-token` | Secret text | SonarQube 代码扫描 |
| `harbor-registry` | Username with password | 镜像仓库推送认证 |
| `gitee-id` | Username with password | Git 仓库拉取认证 |
| `maven-private-repo` | Username with password | 私有 Maven 仓库认证 |

这些凭证的具体值来自 `deploy/jenkins/secret.yaml`，部署前需要替换为实际值。

---

## 四、平台自身安全密钥

### 4.1 JWT 签名密钥

文件：`configs/config.yaml`

```yaml
App:
  JWTSigningKey: "local-dev-jwt-signing-key-32ch"
```

- 用于签发和校验用户登录后的 JWT Token
- 生产环境应替换为长度足够、随机性强的字符串

### 4.2 KubeConfig 加密密钥

文件：`configs/config.yaml`

```yaml
Security:
  KubeConfigEncryptKey: "k8s-operation-default-secret-key"
```

- 用于加密存储在数据库中的 Kubernetes kubeconfig 文件内容
- 生产环境必须替换为强密钥，且不可泄露

---

## 五、常见问题

### 5.1 触发流水线报错：HTTP 401

**现象**：

```
[流水线] Jenkins 构建触发失败 {"error": "获取Job信息失败: HTTP 401"}
```

**原因**：`configs/config.yaml` 中的 `Jenkins.APIToken` 无效、过期或与 `ops-dev` 用户不匹配。

**解决**：按本文档 **2.2 节** 重新生成 API Token 并更新配置。

### 5.2 Jenkins 回调平台失败

**原因**：`deploy/jenkins/secret.yaml` 中的 `hmac-secret` 与 `configs/config.yaml` 中的 `Jenkins.HMACSecret` 不一致，导致回调签名验证失败。

**解决**：确保两边配置相同，并重新部署 Jenkins 或更新后端配置。

---

## 六、安全建议

1. **生产环境务必修改所有默认密码和密钥**，特别是：
   - `admin-password`
   - `JWTSigningKey`
   - `KubeConfigEncryptKey`
   - `Cache.Password`
   - `Database.Password`

2. **不要将 secret.yaml 中的 base64 明文密码提交到仓库**，建议：
   - 使用 `.gitignore` 忽略本地 secret 文件
   - 通过 CI/CD 环境变量或 KMS 注入

3. **定期轮换 Token 和密钥**，尤其是 Jenkins API Token 和镜像仓库凭证。

4. **AI API Key、钉钉/飞书 Webhook 属于敏感信息**，避免泄露到前端或日志中。

---

## 七、相关文件路径

- `configs/config.yaml` — 后端平台核心配置
- `deploy/jenkins/configmap.yaml` — Jenkins JCasC 配置
- `deploy/jenkins/secret.yaml` — Jenkins 初始凭证 Secret
- `deploy/jenkins/secret.yaml.example` — Secret 模板示例
- `pkg/jenkins/client.go` — 后端 Jenkins API 客户端实现
