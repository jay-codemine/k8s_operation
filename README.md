# 🚀 K8sOperation · 企业级 Kubernetes 多集群管理平台

一个对标 **Rancher/KubeSphere** 的企业级 K8s 管理平台，基于 **Go + Gin + GORM + Vue3 + client-go + AI** 构建。

解决中大型企业 **多集群管理分散、权限隔离困难、运维效率低下** 的核心痛点，并集成 **AI 智能助手** 实现自然语言运维。

------

## 🎯 平台定位与核心价值

| 痛点 | 解决方案 |
|------|----------|
| 多集群管理分散 | 统一控制台管理 N 个 K8s 集群（开发/测试/生产） |
| 权限隔离困难 | 三层 RBAC 模型（平台→集群→命名空间） |
| kubectl 门槛高 | 可视化操作，降低使用门槛 |
| 发布不可追溯 | CI/CD 流水线 + 审计日志 |
| 代码质量无管控 | SonarQube 集成 + 质量门禁自动拦截 |
| 构建产物无追踪 | 制品库全生命周期管理（上传/下载/版本/统计） |
| 镜像管理混乱 | 多仓库接入 + 自动清理策略 |
| 可观测性集成复杂 | 构建探针管理，上传即生效，全自动注入 Agent |
| kubectl 操作不便 | 容器 Web 终端，浏览器内直接进入容器 Shell |
| 基础组件部署繁琐 | 应用商城一键部署开源组件 |
| 运维门槛高 | AI 智能助手，自然语言操作平台 |

**核心能力**：
- 🌐 **多集群资源治理** - 统一管理开发/测试/生产环境
- 🔐 **RBAC 精细化权限** - 细粒度权限隔离，满足审计要求
- 🔄 **CI/CD 14 阶段闭环** - 集成 Jenkins，14 阶段全链路流水线，支持审批/回滚/批量发布
- 🔍 **SonarQube 代码质量** - 代码扫描 + 质量门禁，4 语言统一集成（Go/Java/Python/前端）
- 📦 **制品库管理** - 构建产物全生命周期管理，支持上传/下载/版本追踪/统计
- 🏗️ **镜像仓库管理** - 支持 Harbor/ACR/Docker Registry
- 🔭 **构建探针管理** - 可观测性 Agent 全自动注入，上传即生效，无需改流水线
- 🖥️ **容器 Web 终端** - 浏览器内 kubectl exec，WebSocket + xterm.js 交互式 Shell
- 🏪 **应用商城** - 内置应用市场，一键部署开源组件与自有应用
- ⚙ **平台运维监控** - 健康检查、ETCD/核心组件监控、审计日志、系统配置
- 🤖 **AI 智能助手** - 自然语言操作 K8s，多模型支持，高危自动审批

------

## 🏗 技术架构

```
┌─────────────────────────────────────────────────────────────────────┐
│                    前端层 (Vue3 + Vite + Pinia)                      │
│  企业级 UI · 动态权限菜单 · AI 助手面板 · 流水线可视化       │
└─────────────────────────────────────────────────────────────────────┘
                                │ RESTful API + JWT
┌─────────────────────────────────────────────────────────────────────┐
│                      后端层 (Go + Gin + GORM)                        │
│  JWT 认证 · RBAC 鉴权 · AI 服务 · CI/CD 引擎 · 制品管理         │
└─────────────────────────────────────────────────────────────────────┘
      │             │              │              │              │
┌────┴────┐  ┌────┴─────┐  ┌─────┴─────┐  ┌────┴─────┐  ┌────┴─────┐
│  MySQL  │  │   Redis   │  │  K8s API  │  │  Jenkins  │  │  AI LLM  │
│ 持久化   │  │ 会话/缓存 │  │  多集群   │  │  CI/CD   │  │ 多提供商  │
│ 审计日志 │  │ Token管理 │  │ client-go │  │ SonarQube│  │ OpenAI.. │
└─────────┘  └──────────┘  └───────────┘  └──────────┘  └──────────┘
```

### 技术选型说明

| 层面 | 技术 | 选型理由 |
|------|------|----------|
| 后端框架 | Go + Gin | 高并发、云原生标准语言、性能优异 |
| ORM | GORM | 功能完善、支持多数据库、开发效率高 |
| 前端框架 | Vue3 + Vite | 响应式、Composition API、热更新快 |
| 状态管理 | Pinia | 轻量、TypeScript 友好 |
| 认证方案 | JWT + Redis | 无状态、可横向扩展、支持主动失效 |
| K8s 客户端 | client-go | 官方 SDK、功能完整、版本兼容好 |
| AI 引擎 | OpenAI Function Calling | 多模型提供商支持、工具调用、意图识别 |
| CI/CD 引擎 | Jenkins + HMAC | 多语言模板 + 阶段回调 + 签名校验 |
| 代码质量 | SonarQube | Bug/漏洞/异味/覆盖率/重复率 + 质量门禁 |
| 制品库 | 内置制品管理 | SHA256 校验、多类型支持、全生命周期管理 |
| 构建探针 | Build Agent 管理 | 全自动注入可观测性探针（OTEL/SkyWalking/Arthas 等） |
| 容器终端 | xterm.js + WebSocket | 浏览器内交互式 Shell，SPDY 桥接 K8s exec |
| 应用商城 | 内置 App Store | 一键部署开源组件，组件化安装管理 |
| 日志系统 | Zap | 高性能、结构化日志、三日志分类（系统/业务/AI） |

------

## 🔐 三层 RBAC 权限架构（核心亮点）

```
┌─────────────────────────────────────────────────────────────┐
│                    平台角色 (Platform Role)                  │
│         super_admin · platform_admin · developer            │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    集群权限 (Cluster Permission)             │
│              用户可以管理哪些 K8s 集群                        │
│         cluster_admin · cluster_viewer · none               │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                 命名空间权限 (Namespace Permission)           │
│              用户在集群内可以操作哪些 namespace               │
│                    精确到 CRUD 粒度                          │
└─────────────────────────────────────────────────────────────┘
```

### 权限实现机制

**后端**：
- 每个 API 请求携带 `x-cluster-id` 头部
- 中间件校验用户对目标集群的权限
- K8s 操作使用 SubjectAccessReview 二次校验

**前端**：
- 登录时获取完整权限树，缓存到 Pinia Store
- 动态菜单：根据角色显示/隐藏菜单项
- 路由守卫：无权限页面自动拦截跳转
- `v-permission` 指令：按钮级权限控制

```vue
<!-- 按钮级权限控制示例 -->
<button v-permission="'cluster:delete'">删除集群</button>
```

------

## 🔗 项目地址

- Gitee（主仓库）：https://gitee.com/jay-kim/k8s_operation
- GitHub（镜像仓库）：https://github.com/jay-codemine/k8s_operation

> 📦 **配套 AppConfig Operator（Kubebuilder 项目）请访问：**
>  👉 https://gitee.com/jay-kim/appconfig-operator

系统支持多集群管理、事件聚合、滚动升级、镜像更新、扩缩容、Pod 日志流、节点驱逐/隔离、PVC 扩容等能力。

------

## 📦 Kubernetes / client-go 版本兼容性

本项目基于：
- `k8s.io/client-go v0.34.2`

根据官方版本映射规则：
- client-go v0.34.x 对应 Kubernetes v1.34.x

由于 Kubernetes 对客户端有向后兼容策略，旧版本 Kubernetes 也可以正常访问绝大多数 API：
- ✅ 推荐：Kubernetes v1.34.x
- 👍 支持：Kubernetes v1.28.x ~ v1.33.x（大多数功能均正常）
- ⚠ 低于 v1.25 可能存在部分 API 不支持或弃用问题

> 建议生产环境尽量使用与 client-go 主版本号一致的 Kubernetes 版本，以获得最佳兼容性。

------

## 🖥️ 多架构支持（amd64 / arm64）

本项目 **完整支持 ARM64 架构部署**，后端基于 Go 纯静态编译（`CGO_ENABLED=0`），所有依赖均为纯 Go 实现，无 C 库依赖。

### 支持的目标架构

| 架构 | 说明 | 典型场景 |
|------|------|----------|
| `linux/amd64` | x86_64，默认架构 | 传统云服务器、PC 开发机 |
| `linux/arm64` | AArch64 | 华为鲲鹏、AWS Graviton、Apple M 系列、树莓派 4/5 |

### 构建方式

**方式一：本地交叉编译（Go 原生支持）**

```bash
# 编译 arm64 二进制
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o bin/k8soperation-arm64 ./cmd/k8soperation

# 编译 amd64 二进制（默认）
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o bin/k8soperation-amd64 ./cmd/k8soperation
```

**方式二：Docker 单架构构建**

```bash
# 构建 arm64 镜像
docker build --build-arg TARGETARCH=arm64 -t k8soperation:arm64 .

# 构建 amd64 镜像（默认）
docker build -t k8soperation:amd64 .
```

**方式三：Docker Buildx 多架构镜像（推荐生产使用）**

```bash
# 一次构建同时生成 amd64 + arm64 镜像并推送到仓库
make docker-buildx IMAGE=registry.example.com/k8soperation:latest

# 或手动执行
docker buildx build --platform linux/amd64,linux/arm64 \
  -t registry.example.com/k8soperation:latest . --push
```

### 技术保障

- **纯 Go 编译**：`CGO_ENABLED=0`，无 C 依赖，交叉编译零配置
- **依赖全兼容**：`client-go`、`gin`、`gorm`、`go-openai` 等均为纯 Go 库，天然支持多架构
- **Dockerfile 适配**：使用 `BUILDPLATFORM` + `TARGETARCH` 实现 BuildKit 多架构构建
- **前端无架构依赖**：Vue3 + Vite 构建产物为静态 HTML/JS/CSS，任何架构通用

## 🖥️ 系统界面展示

> 以下为真实系统运行截图

---

### 🔹 核心界面预览

<p align="center">
  <img src="docs/images/1.png" width="900"/>
</p>

<p align="center">
  <img src="docs/images/2.png" width="900"/>
</p>

<p align="center">
  <img src="docs/images/3.png" width="900"/>
</p>

---

<details>
<summary>📂 点击展开查看全部界面截图 </summary>

<br/>

<p align="center">
  <img src="docs/images/4.png" width="900"/>
</p>

<p align="center">
  <img src="docs/images/5.png" width="900"/>
</p>

<p align="center">
  <img src="docs/images/6.png" width="900"/>
</p>

<p align="center">
  <img src="docs/images/7.png" width="900"/>
</p>

<p align="center">
  <img src="docs/images/8.png" width="900"/>
</p>

<p align="center">
  <img src="docs/images/9.png" width="900"/>
</p>

<p align="center">
  <img src="docs/images/10.png" width="900"/>
</p>

<p align="center">
  <img src="docs/images/11.png" width="900"/>
</p>

<p align="center">
  <img src="docs/images/12.png" width="900"/>
</p>

<p align="center">
  <img src="docs/images/13.png" width="900"/>
</p>

<p align="center">
  <img src="docs/images/14.png" width="900"/>
</p>

<p align="center">
  <img src="docs/images/15.png" width="900"/>
</p>

<p align="center">
  <img src="docs/images/16.png" width="900"/>
</p>

<p align="center">
  <img src="docs/images/17.png" width="900"/>
</p>

<p align="center">
  <img src="docs/images/18.png" width="900"/>
</p>


<p align="center">
  <img src="docs/images/19.png" width="900"/>
</p>


<p align="center">
  <img src="docs/images/20.png" width="900"/>
</p>

<p align="center">
  <img src="docs/images/21.png" width="900"/>
</p>

<p align="center">
  <img src="docs/images/22.png" width="900"/>
</p>

<p align="center">
  <img src="docs/images/23.png" width="900"/>
</p>

<p align="center">
  <img src="docs/images/24.png" width="900"/>
</p>

<p align="center">
  <img src="docs/images/25.png" width="900"/>
</p>
<p align="center">
  <img src="docs/images/26.png" width="900"/>
</p>
<p align="center">
  <img src="docs/images/26.png" width="900"/>
</p>

</details>



## ✨ 核心特性

### 🚦 CI/CD 14 阶段全链路闭环

平台内置 **14 阶段 CI/CD 闭环流水线**，覆盖从代码检出到生产部署的完整生命周期，将 CI/CD 从脚本编排升级为 **可控、可追踪、可回滚** 的企业级发布体系。

#### 14 阶段执行矩阵

```
┌──────┬──────────────────┬──────────────────────────────────┬──────────┐
│ 序号 │ 阶段             │ 说明                             │ 可选     │
├──────┼──────────────────┼──────────────────────────────────┼──────────┤
│  1   │ Clean Workspace  │ 清理 Jenkins 工作空间            │          │
│  2   │ Checkout         │ 拉取代码 + 提取 Git 信息         │          │
│  3   │ Dependencies     │ 依赖下载（go mod/maven/npm/pip） │          │
│  4   │ Compile          │ 编译检查                         │          │
│  5   │ Test             │ 单元测试 + 覆盖率报告            │          │
│  6   │ Lint             │ 静态分析（golangci-lint/flake8）  │          │
│  7   │ SonarQube Scan   │ 代码质量扫描（Bug/漏洞/异味）    │ ✅ 按需  │
│  8   │ Quality Gate     │ 质量门禁（不通过则中断流水线）   │ ✅ 按需  │
│  9   │ Build Binary     │ 构建二进制/打包制品              │ ✅ 按需  │
│ 10   │ Upload Artifact  │ 上传制品到制品库                 │ ✅ 按需  │
│ 11   │ Build Image      │ nerdctl/BuildKit 构建容器镜像     │          │
│ 12   │ Push Image       │ 推送镜像到 Harbor（并发 8 层）    │          │
│ 13   │ Approval         │ 人工审批（生产环境强制）         │ ✅ 按需  │
│ 14   │ Deploy           │ 滚动更新到目标集群               │ ✅ 按需  │
└──────┴──────────────────┴──────────────────────────────────┴──────────┘
```

#### 多语言模板支持

| 语言 | 模板文件 | 特性 |
|------|----------|------|
| **Go** | `go-pipeline.groovy` | go test + golangci-lint + SonarQube + 制品上传 + nerdctl build |
| **Java** | `java-spring-pipeline.groovy` | Maven + SonarQube + 质量门禁 + 制品上传 |
| **Python** | `python-pipeline.groovy` | pip + flake8 + pytest + SonarQube + 制品上传 |
| **前端** | `frontend-pipeline.groovy` | npm ci + SonarQube + 制品上传 + Nginx 镜像 |

> 所有模板统一支持 **BuildKit 层缓存** + **nerdctl push --concurrency 8 并发推送**，二次构建速度提升 50%+

#### 发布流程
1. 创建发布单（Release）
2. 触发构建（CI）：Jenkins 自动执行 14 阶段流水线
3. 代码扫描 + 质量门禁自动拦截（可选）
4. 构建制品上传到制品库（可选）
5. 构建镜像 + 推送到 Harbor
6. 进入待审批（PENDING_APPROVAL）
7. 人工审批通过（APPROVED）后进入部署（DEPLOYING）
8. 滚动更新（RollingUpdate）并持续采集状态
9. 发布成功（SUCCEEDED）或失败（FAILED）
10. 支持一键回滚（ROLLBACKING -> ROLLED_BACK）

#### 能力说明
- ✅ **SonarQube 代码扫描**：Bug/漏洞/代码异味/覆盖率/重复率，扫描报告自动回传平台
- ✅ **质量门禁**：扫描不通过自动中断流水线，确保代码质量底线
- ✅ **制品库管理**：构建产物自动上传，支持下载/版本追踪/统计分析
- ✅ **人工审批**：审批人/审批时间/审批意见记录，可做权限控制与审计
- ✅ **发布状态机**：状态迁移可控，避免"脚本式发布不可追踪"
- ✅ **HMAC-SHA256 签名校验**：防伪造回调，保障 Jenkins ↔ 平台通信安全
- ✅ **发布日志**：记录构建/部署/回滚全过程关键事件
- ✅ **回滚能力**：
    - Deployment：基于 ReplicaSet 历史版本回滚（或指定历史版本）
    - StatefulSet/DaemonSet：基于 ControllerRevision 回滚
- ✅ **发布过程可观测**：14 阶段实时状态 + 滚动更新进度 + Pod 状态 + 事件聚合
- ✅ **批量发布**：支持多个发布单同时触发构建，提升发布效率
- ✅ **批量回滚**：支持多个发布单一键批量回滚到上一个稳定版本
- ✅ **部署失败重试**：失败的发布单支持一键重试，自动创建新的发布单

### 🤖 AI 智能助手（核心亮点）

平台内置 AI 智能助手，支持用自然语言操作 K8s 集群，大幅降低运维门槛。

#### AI 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                      用户自然语言输入                          │
│  "查看 default 命名空间的 Pod"  "扩容 nginx 到 5 个副本"   │
└─────────────────────────┴───────────────────────────────────┘
                              │
                    ┌───────┴────────┐
                    │  LLM 意图识别  │  ← OpenAI Function Calling
                    └───────┬────────┘
                            │
                ┌────────┼────────┐
                │           │           │
          ┌────┴────┐ ┌──┴─────┐ ┌──┴──────┐
          │ 只读查询 │ │ 写操作  │ │ 高危操作  │
          │ 直接执行 │ │ 确认执行 │ │ 审批流程  │
          └─────────┘ └────────┘ └─────────┘
```

#### 核心能力

| 能力 | 说明 |
|------|------|
| 🗣️ 自然语言操作 | 用中/英文描述运维意图，AI 自动转化为 K8s API 调用 |
| 🧠 意图识别 | 基于 OpenAI Function Calling，自动识别查询/扩缩容/删除等意图 |
| 🛡️ 高危操作审批 | 删除/扩容等危险操作自动进入审批流程，防止误操作 |
| 🔄 多模型支持 | 支持 NVIDIA NIM、OpenAI、国产大模型等多提供商，可热切换 |
| 💬 多轮对话 | 支持上下文连续对话，理解复杂运维场景 |
| 📊 意图标签 | 前端可视化展示操作意图（🔍 查询 / ⚙️ 执行 / ⚠️ 审批） |
| 📄 AI 日志 | 独立 AI 日志系统，全链路追踪，支持 API 查询 |

#### 支持的工具调用（Function Calling）

```
✔ list_pods          - 查看 Pod 列表
✔ list_deployments   - 查看 Deployment 列表
✔ list_services      - 查看 Service 列表
✔ list_namespaces    - 查看命名空间
✔ get_pod_detail     - 获取 Pod 详情
✔ get_pod_logs       - 获取 Pod 日志
✔ scale_deployment   - 扩缩容（需审批）
✔ delete_pod         - 删除 Pod（需审批）
✔ restart_deployment - 重启 Deployment（需审批）
... 更多工具持续扩展中
```

#### 多模型提供商支持

```yaml
# config.yaml 配置示例
ai:
  enabled: true
  default_provider: nvidia           # 默认提供商
  providers:
    - id: nvidia
      name: "NVIDIA NIM"
      base_url: "https://integrate.api.nvidia.com/v1"
      api_key: "nvapi-xxx"
      model: "minimaxai/minimax-m2.7"
    - id: openai
      name: "OpenAI"
      base_url: "https://api.openai.com/v1"
      api_key: "sk-xxx"
      model: "gpt-4o"
```

> 💡 支持任何兼容 OpenAI API 协议的提供商（NVIDIA NIM、通义千问、智谱 AI、DeepSeek 等）

#### AI 日志与排查

AI 助手拥有独立的日志系统（`storage/logs/ai.log`），方便排查大模型调用问题：

```
─ 全链路追踪：Registry 解析 → API 请求 → 工具执行 → 审批流程 → 响应完成
─ 记录开销：每次调用的 latency、prompt_tokens、completion_tokens
─ 错误定位：模型拒绝/超时/调用失败自动记录 Error 级别日志
─ API 查询：GET /api/v1/ai/logs?level=error&keyword=超时
```

------

### 🔍 SonarQube 代码质量扫描

平台深度集成 SonarQube，为 **Go/Java/Python/前端** 四种语言提供统一的代码质量管控能力。

#### 扫描指标

| 指标 | 说明 |
|------|------|
| Bugs | 代码缺陷数量 |
| Vulnerabilities | 安全漏洞数量 |
| Code Smells | 代码异味（可维护性问题） |
| Coverage | 单元测试覆盖率 (%) |
| Duplications | 重复代码率 (%) |
| Security Hotspots | 安全热点数量 |
| Reliability/Security/Maintainability Rating | A~E 五级评级 |

#### 工作流程

```
Jenkins 构建 → SonarQube Scanner 执行扫描
     → 质量门禁检查（不通过则中断流水线）
     → 扫描报告自动回传平台（sonar-callback）
     → 前端展示完整质量报告
```

- 每种语言自动注入对应的 `SONAR_SOURCES` 和 `SONAR_EXCLUSIONS`
- 通过 `pipeline.EnableSonar` 一键开关，无需修改 Jenkinsfile
- 质量门禁状态：`OK`（通过） / `WARN`（警告） / `ERROR`（未通过）

### 📦 制品库管理

平台内置制品库，对 CI/CD 构建产物进行全生命周期管理。

#### 支持的制品类型

| 类型 | 语言 | 说明 |
|------|------|------|
| `jar` / `war` | Java | Maven 构建产物 |
| `binary` | Go | 编译后的二进制文件 |
| `dist` | 前端 | 前端构建产物（dist.tar.gz） |
| `wheel` | Python | Python wheel 包 |
| `image` | 通用 | Docker 镜像引用（仅记录元数据） |
| `archive` | 通用 | 通用压缩包 |

#### 核心能力

- ✅ **Jenkins 自动上传**：构建完成后自动 gzip 压缩 + curl multipart 上传
- ✅ **手动上传**：支持通过平台界面手动上传制品
- ✅ **SHA256 校验**：上传时自动计算文件哈希，确保完整性
- ✅ **版本追踪**：关联流水线 + 运行记录 + Git 信息（repo/branch/commit）
- ✅ **下载管理**：支持 Token 鉴权下载 + 下载计数统计
- ✅ **批量操作**：支持批量删除 + 制品统计（按类型分组）
- ✅ **生命周期**：uploading → ready → expired → deleted 状态流转

```
API 接口一览：
├── GET  /artifact/list          # 制品列表（分页 + 多维筛选）
├── GET  /artifact/detail        # 制品详情
├── GET  /artifact/by-run        # 某次运行的制品列表
├── POST /artifact/upload        # 上传制品（Jenkins 回调 / 手动）
├── POST /artifact/create        # 创建制品记录（镜像类型）
├── POST /artifact/attach        # 为已有制品补传/替换文件
├── POST /artifact/update        # 更新制品信息
├── GET  /artifact/download      # 下载制品文件
├── POST /artifact/delete        # 删除制品
├── POST /artifact/batch-delete  # 批量删除
└── GET  /artifact/stats         # 制品统计（按类型分组）
```

### 🔭 构建探针管理（Build Agent Management）

平台内置构建探针管理系统，实现 **上传即生效** 的全自动 Agent 注入，无需修改流水线或 Dockerfile。

#### 支持的探针分类

| 分类 | 说明 | 典型探针 |
|------|------|----------|
| 可观测性 | 分布式追踪/指标采集 | OpenTelemetry Java Agent、SkyWalking |
| 诊断工具 | 在线诊断/性能分析 | Arthas、JProfiler Agent |
| 安全扫描 | 运行时安全防护 | RASP、Falco Probe |
| 自定义 | 任意自定义 Agent | 用户上传的任意 JAR/Binary |

#### 核心能力

- ✅ **全自动注入**：Jenkins 流水线 `Prepare Build Agents` 阶段自动从平台拉取所有已启用探针
- ✅ **动态 Dockerfile**：Build Image 阶段自动生成 COPY + ENV 指令，零人工干预
- ✅ **三级降级策略**：平台 API → 项目本地 → Maven 中央仓库，确保构建不中断
- ✅ **SHA256 校验**：上传时自动计算文件哈希，确保探针完整性
- ✅ **多语言支持**：Java / Go / Python / 通用，按语言范围自动匹配
- ✅ **版本管理**：支持版本追踪、启用/停用、安装包上传与下载

### 🖥️ 容器 Web 终端（Container Terminal）

类似 `kubectl exec -it` 的浏览器内交互式终端，基于 **WebSocket + SPDY** 双协议桥接。

#### 核心特性

| 特性 | 说明 |
|------|------|
| Tokyo Night 主题 | 暗色终端配色，参考 VS Code Tokyo Night 主题 |
| Shell 自动检测 | 按优先级探测 bash → sh → zsh → ash |
| 窗口自适应 | ResizeObserver + FitAddon 自动适配终端尺寸 |
| 心跳保活 | 双端 ping/pong（前端 25s / 后端 30s） |
| 全屏模式 | 一键切换全屏/窗口模式 |
| 拖拽支持 | 标题栏可拖拽移动窗口位置 |
| Distroless 友好 | 无 Shell 容器自动检测并给出调试提示 |

### 📡 工作负载实时状态监听（Resource Watcher）

参考 KubeSphere 的状态追踪机制，当用户执行镜像更新/重启等操作后，自动开启快速轮询：

- 实时展示 `Updating → Progressing → Running` 完整状态变化
- 自动拉取关联 Events 显示在右下角浮窗
- 覆盖全部 5 类工作负载（Deployment / StatefulSet / DaemonSet / Job / CronJob）

### 🏪 应用商城（App Store）

内置应用市场，支持一键部署开源组件与自有应用到 K8s 集群。

- ✅ **应用目录管理**：分类浏览、搜索、版本管理
- ✅ **一键安装**：选择集群 + 命名空间 + 自定义 Values，自动创建 Deployment/Service
- ✅ **组件化部署**：每个应用支持多组件（如 Prometheus + Grafana + AlertManager）
- ✅ **安装状态追踪**：实时监控 Pod Ready 状态，部分就绪降级运行
- ✅ **卸载与管理**：一键卸载，清理所有关联资源

### 🧩 系统通用能力

- 配置化加载（YAML / ENV）
- JWT 鉴权 + 刷新机制
- Zap 三日志系统（系统日志 / 业务日志 / AI 日志）
- Swagger 在线 API 文档（支持 Standalone）
- 健康检查与优雅关闭（含 ETCD / Controller Manager / Scheduler / CoreDNS 核心组件监控）
- 标准化控制器 / 服务 / DAO 分层
- 全局异常拦截（中间件）

------

## ☸ Kubernetes 高级能力（全部已实现）

### Deployment 管理

- CRUD、扩缩容、镜像更新、滚动升级
- 滚动重启、基于 ReplicaSet 的版本回滚
- Pods 列表、事件聚合、历史版本查询

### Pod 管理

- 列表、详情、日志（流式/非流）
- 镜像 Patch、事件查询、强制删除

### StatefulSet / DaemonSet

- CRUD、扩缩容、镜像更新
- ControllerRevision 回滚

### Service / Ingress

- CRUD
- Strategic / JSON Merge Patch
- TLS 配置、事件聚合

### Job / CronJob

- Job：创建 / 删除 / 状态查询
- CronJob：启停、删除、历史 Job 查询

### Secret / PVC / PV / ConfigMap / StorageClass

- 全生命周期管理
- PVC 扩容、PV ReclaimPolicy 修改
- ConfigMap Patch、StorageClass CRUD

### Node 高级管理

- Cordon / Uncordon
- Drain（驱逐可驱逐 Pod）
- Pod Evict（支持 gracePeriod）
- 节点 Metrics、Pods 列表

### Event 事件聚合

- Pod / Deploy / StatefulSet / Node 等资源
- 支持排障快速定位（Backoff、PullError、Unschedulable）

------

## 🧩 多集群管理

- 动态添加/切换多个 K8s 集群
- **kubeconfig 加密存储**（AES-256-GCM）
- TLS 证书动态信任（解决 x509 证书问题）
- 连通性检测与健康状态监控
- 多集群 clientset 连接池管理

适合企业多集群统一管控场景。

------

## 💡 技术亮点

### 1. 敏感数据加密存储

```go
// kubeconfig 使用 AES-256-GCM 加密
// 数据库存储格式：ENC:base64(nonce+ciphertext)
func EncryptKubeConfig(plain, key string) string {
    // AEAD 加密，防篡改，安全性高
}
```

### 2. 统一错误码体系

```go
// 前端可根据错误码精准处理不同场景
var (
    InvalidParams    = NewError(10001, "参数错误")
    Unauthorized     = NewError(10002, "认证失败")
    ClusterNotFound  = NewError(20001, "集群不存在")
    PermissionDenied = NewError(20002, "权限不足")
)
```

### 3. K8s 客户端连接池

```go
// 多集群场景，每个集群维护独立 clientset
// 支持动态切换，避免重复创建连接
type ClusterClientManager struct {
    clients sync.Map // clusterId -> *kubernetes.Clientset
}
```

### 4. 配置热更新

```
系统设置存储在数据库，修改后立即生效
├── 基础设置（默认页面、语言、时区）
├── 安全设置（会话超时、密码策略）
├── 告警设置（CPU/内存/磁盘阈值）
└── 通知设置（邮件/钉钉/Webhook）
```

### 5. AI 智能助手引擎

```go
// 多模型提供商注册中心（Provider Registry）
// 支持 NVIDIA NIM、OpenAI、国产大模型等，可热切换
type ProviderRegistry struct {
    providers sync.Map // providerID -> *Client
    defaults  struct { providerID, modelID string }
}

// Function Calling 工具注册 + 高危操作审批
func (s *AIService) AIChat(msg string) {
    // 1. LLM 意图识别（Function Calling）
    // 2. 工具执行（list_pods / scale_deployment ...)
    // 3. 高危操作 → 自动进入审批流程
    // 4. 全链路 AI 日志记录
}
```

------

## 💡 项目难点与解决方案

| 难点 | 解决方案 |
|------|----------|
| 多集群 TLS 证书信任 | 解析 kubeconfig 中的 CA，动态构建 TLS Config |
| 权限隔离粒度细 | 三层模型 + 前后端双重校验 |
| 配置热更新 | 数据库存储 + 内存缓存，修改后立即生效 |
| 空集群启动 | 降级策略，无集群时跳过 K8s 初始化 |
| 错误码统一 | 全局错误码体系 + 中间件统一处理 |
| 前后端类型一致性 | DAO 层强制返回空切片而非 nil |
| AI 多模型兼容性 | Provider Registry 抽象层 + 统一 OpenAI 协议适配 |
| AI 调用问题排查 | 独立 AI 日志 + 全链路追踪 + API 查询接口 |
| CI/CD 14 阶段闭环 | DefaultStageDefinitions 声明式阶段 + Jenkins 回调协议 + 动态阶段推断 |
| 代码质量管控 | SonarQube 4 语言统一集成 + 质量门禁自动拦截 + 扫描报告回传 |
| 制品全生命周期 | SHA256 完整性校验 + 1MB 大缓冲区加速 I/O + 流式下载 |
| 构建性能优化 | BuildKit 本地层缓存 + nerdctl push --concurrency 8 并发推送 |
| 构建探针全自动注入 | 平台 API 自动拉取 + 动态 Dockerfile 生成 + 三级降级策略 |
| 容器终端桥接 | WebSocket ↔ SPDY 双协议桥接 + Shell 自动检测 + 心跳保活 |
| 应用商城组件化部署 | 多组件自动创建 + Pod Ready 等待 + 部分就绪降级 |
| Jenkins 高并发优化 | Client 连接池化 + Job 信息 5min TTL 缓存 + PollWorker 5 并行 |

------



------

## 📈 项目收益

通过这个平台，运维团队可以：

- **效率提升**：多集群统一管理，减少 80% 切换成本
- **安全合规**：细粒度权限隔离，满足审计要求
- **降低门槛**：开发人员无需学习 kubectl，可视化操作
- **AI 赋能**：自然语言操作 K8s，零门槛运维，高危自动拦截审批
- **质量保障**：SonarQube 代码扫描 + 质量门禁，从源头把控代码质量
- **制品管理**：构建产物全链路追踪，版本可查、文件可下载、统计可视
- **探针管理**：构建探针上传即生效，全自动注入到 Docker 镜像，零运维成本
- **终端体验**：浏览器内直接进入容器 Shell，免去 kubectl 配置
- **应用商城**：一键部署开源组件，降低基础设施搭建成本

------

## 📦 项目结构（真实仓库对应）

```bash
k8soperation/
├── cmd/k8soperation/          # 程序入口
├── configs/                   # 配置文件（config.yaml / k8s.yaml）
├── docs/                      # 部署文档 / SQL / Swagger
├── global/                    # 全局变量（DB/Redis/Logger/AILogger）
├── initialize/                # 初始化（日志/数据库/路由/集群）
├── internal/
│   ├── app/
│   │   ├── controllers/       # API 控制器（含 AI / CI/CD / 制品库 / 探针 / 终端 / 应用商城）
│   │   ├── services/          # 业务服务（含 AI / 流水线 / 制品 / SonarQube / 探针 / 应用商城）
│   │   ├── models/            # 数据模型（含 CicdArtifact / SonarQube / BuildAgent / AppStore）
│   │   ├── dao/               # 数据访问层
│   │   └── routers/           # 路由注册
│   ├── bootstrap/             # 启动编排
│   └── errorcode/             # 统一错误码
├── pkg/
│   ├── k8s/                   # K8s 客户端封装（含容器终端 WebSocket 桥接）
│   ├── openai/                # AI 多模型 Provider Registry
│   ├── jwt/                   # JWT 认证
│   └── jenkins/               # Jenkins CI 集成（连接池化 + 缓存优化）
├── configs/
│   ├── jenkins-templates/     # 多语言 Jenkins 流水线模板
│   │   ├── go-pipeline.groovy
│   │   ├── java-spring-pipeline.groovy
│   │   ├── python-pipeline.groovy
│   │   └── frontend-pipeline.groovy
│   └── dockerfile-templates/  # 多语言 Dockerfile 模板
├── k8s-web/                   # Vue3 前端项目
│   └── src/components/AiAssistant.vue  # AI 助手前端组件
├── build/                     # Docker / Containerd 构建
├── storage/
│   ├── logs/                  # 日志（app.log / biz.log / ai.log）
│   └── artifacts/             # 制品存储（pipeline_id/日期/文件）
└── docs/                      # 部署文档 / SQL / Swagger / 发布流程
```

------

## ⚙️ 快速启动

> 📖 **完整操作手册**：[QUICK_START.md](QUICK_START.md) — 包含详细的环境准备、配置说明、Docker 部署、常见问题等

### 🚀 一键启动（推荐）

项目提供一键脚本，自动完成 **环境检查 → 数据库初始化 → 配置生成 → 编译 → 启动**：

```bash
# Linux / macOS
git clone https://gitee.com/jay-kim/k8s_operation.git
cd k8s_operation
chmod +x scripts/*.sh
bash scripts/quick-start.sh
```

```powershell
# Windows (PowerShell)
git clone https://gitee.com/jay-kim/k8s_operation.git
cd k8s_operation
powershell -ExecutionPolicy Bypass -File scripts\quick-start.ps1
```

> 💡 所有配置均可通过环境变量覆盖：`DB_HOST`、`DB_PASS`、`REDIS_HOST`、`REDIS_PASS` 等，详见 [QUICK_START.md](QUICK_START.md#七脚本工具一览)

### 📋 环境要求

| 组件 | 最低版本 | 说明 |
|------|---------|------|
| **Go** | 1.21+ | 后端编译（必须） |
| **MySQL** | 5.7+ | 主数据库（必须） |
| **Redis** | 5.0+ | Session/缓存（必须） |
| Node.js | 20+ | 前端编译（可选） |
| Docker | 20+ | 容器化部署（可选） |

### 🔧 手动分步启动

```bash
# 1. 克隆仓库
git clone https://gitee.com/jay-kim/k8s_operation.git
cd k8s_operation

# 2. 初始化数据库（MySQL 8.0+，含 34 张表 + RBAC 权限 + CICD 模板）
mysql -u root -padmin123 --default-character-set=utf8mb4 < docs/sql/k8s_platform_full_init.sql
# 或使用脚本：bash scripts/init-db.sh

# 3. 生成配置文件
cp configs/config.yaml.example configs/config.yaml
# 编辑 configs/config.yaml 修改数据库/Redis 连接信息

# 4. 编译 & 启动后端
make build
./bin/k8soperation
# 后端运行在 http://localhost:8080

# 5. 编译 & 启动前端
cd k8s-web
npm install
npm run dev
# 前端运行在 http://localhost:5173
```

### 🐳 Docker 一键部署

```bash
# 单架构构建
make docker-build && make docker-run

# 多架构构建（amd64 + arm64）
make docker-buildx IMAGE=registry.example.com/k8soperation:latest
```

### 🔑 访问系统

| 服务 | 地址 |
|------|------|
| 前端界面 | http://localhost:5173 |
| 后端 API | http://localhost:8080 |
| Swagger 文档 | http://localhost:8080/swagger |
| 默认账号 | `admin` / `admin123` |

### 🛠 脚本工具

| 脚本 | 平台 | 用途 |
|------|------|------|
| `scripts/quick-start.sh` / `.ps1` | Linux/Mac / Windows | 一键启动（全自动） |
| `scripts/check-env.sh` / `.ps1` | Linux/Mac / Windows | 仅检查环境依赖 |
| `scripts/init-db.sh` / `.ps1` | Linux/Mac / Windows | 独立数据库初始化 |

------

## 📄 部署文档（强烈推荐阅读）

官方部署说明文档（包括 **后端服务** 与 **前端管理界面** 的部署方式）：

👉 **K8sOperation 后台系统部署文档（后端）**  
https://gitee.com/jay-kim/k8s_operation/blob/master/docs/📄%20K8sOperation%20后台系统部署文档.md

👉 **前端管理系统部署文档（k8s-web）**  
https://gitee.com/jay-kim/k8s_operation/blob/master/docs/%E5%89%8D%E7%AB%AF%E7%AE%A1%E7%90%86%E7%B3%BB%E7%BB%9F%E9%83%A8%E7%BD%B2%E6%96%87%E6%A1%A3.md

---

### 后端部署文档内容包括：

- 构建后端二进制
- Docker / Containerd 镜像构建（支持 **amd64 / arm64** 多架构）
- 使用 Systemd 管理服务
- Kubernetes Deployment / Service 部署示例
- 参数说明与优化建议
- 生产环境目录规划

---

### 前端部署文档内容包括：

- 前端项目构建（Vite）
- 环境变量配置（API_BASE）
- Nginx 部署（SPA 路由支持）
- Docker 部署
- Kubernetes（Deployment / Service / Ingress）
- 与后端 API 对接说明

## 🔗 关联项目（推荐配套使用）

### 📘 AppConfig Operator

（Kubebuilder 开发，用于管理自定义资源 AppConfig）
👉 https://gitee.com/jay-kim/appconfig-operator

Operator → 管理 AppConfig CRD
k8soperation → 提供 HTTP API/Web 后台

两者解耦，便于独立演进。

------

## ⭐ Star / Watch / Fork

如果本项目对你有帮助，非常欢迎：

- ⭐ **Star**
- 👀 **Watch**
- 🍴 **Fork**

你的支持是我持续完善的最大动力！

------

## 📜 License

Apache-2.0