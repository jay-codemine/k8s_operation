# K8sOperation CI/CD 完整流程图

## 一、总体架构流程

```mermaid
graph TB
    subgraph 开发者
        A[代码提交到 Git]
    end

    subgraph K8sOperation平台
        B[创建/配置流水线]
        C[触发构建]
        D[参数注入]
        E[接收阶段回调]
        F[接收最终回调]
        G{需要审批?}
        H[创建审批记录]
        I{审批通过?}
        J[自动部署到 K8s]
        K[同步发布记录]
        L[发送通知]
    end

    subgraph Jenkins
        M[接收构建请求]
        N[创建 K8s Pod Agent]
        O[执行 Pipeline Stages]
        P[阶段回调到平台]
        Q[最终回调到平台]
    end

    subgraph K8s集群
        R[更新 Deployment 镜像]
        S[Rollout 滚动更新]
        T[健康检查通过]
    end

    subgraph 通知渠道
        U[钉钉/飞书通知]
        V[飞书审批卡片]
    end

    A --> B
    B --> C
    C --> D
    D --> M
    M --> N
    N --> O
    O --> P
    P --> E
    O --> Q
    Q --> F
    F --> G
    G -->|是| H
    H --> V
    V --> I
    I -->|通过| J
    I -->|拒绝| L
    G -->|否,AutoDeploy=true| J
    G -->|否,AutoDeploy=false| K
    J --> R
    R --> S
    S --> T
    T --> K
    K --> L
    L --> U
```

---

## 二、Jenkins Pipeline 阶段详细流程

```mermaid
graph LR
    subgraph Pipeline阶段
        A[Clean Workspace] --> B[Checkout]
        B --> C[Compile/Test]
        C --> D{SonarQube?}
        D -->|启用| E[代码质量扫描]
        D -->|跳过| F[Build & Push Image]
        E --> F
        F --> G[制品上传]
        G --> H[最终回调]
    end

    subgraph 阶段回调
        A1[stageCallback checkout]
        A2[stageCallback compile]
        A3[stageCallback sonar]
        A4[stageCallback build]
        A5[stageCallback push]
        A6[callbackPlatform SUCCESS/FAILURE]
    end

    B --> A1
    C --> A2
    E --> A3
    F --> A4
    F --> A5
    H --> A6
```

---

## 三、Java 项目构建详细流程

```mermaid
graph TB
    A[平台触发构建] --> B[Jenkins 创建 Pod]

    subgraph Pod Agent 容器
        C[maven 容器<br/>动态 JDK 版本]
        D[kaniko 容器<br/>镜像构建]
        E[jnlp 容器<br/>Agent 通信]
    end

    B --> C
    B --> D
    B --> E

    C --> F[git checkout]
    F --> G[自动检测 pom.xml 位置]
    G --> H[mvn clean package]
    H --> I{测试?}
    I -->|跳过| J[生成 Dockerfile]
    I -->|执行| K[mvn test]
    K --> J
    J --> L[Kaniko 构建镜像]
    L --> M[推送到 ACR]
    M --> N[回调平台]
    N --> O[Pod 自动销毁]
```

---

## 四、发布审批流程（双路径）

```mermaid
graph TB
    subgraph 路径一：流水线自动发布
        A1[流水线构建成功] --> B1{RequireApproval?}
        B1 -->|是| C1[创建审批记录]
        C1 --> D1[发送飞书审批卡片]
        D1 --> E1{审批结果}
        E1 -->|通过| F1[执行自动部署]
        E1 -->|拒绝| G1[标记失败,通知]
        B1 -->|否,AutoDeploy=true| F1
        B1 -->|否,AutoDeploy=false| H1[仅记录,不部署]
    end

    subgraph 路径二：手动创建发布单
        A2[创建发布单] --> B2[强制创建审批]
        B2 --> C2{多级审批}
        C2 --> D2[第1级审批]
        D2 -->|通过| E2{还有下级?}
        E2 -->|是| F2[第N级审批]
        F2 --> E2
        E2 -->|全部通过| G2[入队执行部署]
        D2 -->|拒绝| H2[标记失败]
    end

    subgraph 部署执行
        F1 --> I[Patch K8s Workload]
        G2 --> I
        I --> J[等待 Rollout 完成]
        J --> K{部署成功?}
        K -->|是| L[发送成功通知]
        K -->|否| M[发送失败通知]
    end
```

---

## 五、多语言构建容器选择

```mermaid
graph TB
    A[平台触发构建] --> B{语言类型}

    B -->|Java| C[maven 容器<br/>3.9.9-eclipse-temurin-JDK-OS<br/>JDK: 8/11/17/21]
    B -->|Go| D[golang 容器<br/>golang:1.24-alpine]
    B -->|Frontend| E[node 容器<br/>node:22-alpine]
    B -->|Python| F[python 容器<br/>python:3.11-slim]

    C --> G[kaniko 构建镜像]
    D --> G
    E --> G
    F --> G

    G --> H[推送到阿里云 ACR]
```

---

## 六、回调与通知机制

```mermaid
graph TB
    subgraph Jenkins构建中
        A[每个 Stage 完成] --> B[stageCallback]
        C[Pipeline 结束] --> D[callbackPlatform]
    end

    subgraph 平台接收
        B --> E[POST /stage/callback<br/>实时更新阶段状态]
        D --> F[POST /pipeline/callback<br/>更新最终构建结果]
    end

    subgraph HMAC签名验证
        G[X-Signature Header]
        H[hmacSha256 job:build:stage]
    end

    B --> G
    D --> G
    G --> H

    subgraph 通知分发
        F --> I{构建结果}
        I -->|成功+需审批| J[飞书审批卡片]
        I -->|成功+自动部署| K[部署开始通知]
        I -->|成功+无部署| L[构建成功通知]
        I -->|失败| M[构建失败通知]
        K --> N{部署结果}
        N -->|成功| O[部署成功通知]
        N -->|失败| P[部署失败通知]
    end

    subgraph 通知渠道
        J --> Q[飞书 Webhook]
        L --> R[钉钉 Webhook]
        O --> R
        P --> R
    end
```

---

## 七、数据流全景

```
┌──────────────┐     触发构建(API)      ┌──────────────┐     buildWithParameters     ┌──────────────┐
│              │ ──────────────────────▶ │              │ ──────────────────────────▶ │              │
│   前端 UI    │                         │   后端 API   │                             │   Jenkins    │
│  (Vue3)      │ ◀────────────────────── │   (Go/Gin)   │ ◀──────────────────────────  │              │
│              │     WebSocket/轮询       │              │   HTTP 回调(HMAC签名)       │              │
└──────────────┘                         └──────────────┘                             └──────────────┘
                                               │                                            │
                                               │ K8s API                                    │ K8s API
                                               ▼                                            ▼
                                         ┌──────────────┐                             ┌──────────────┐
                                         │  K8s 集群    │                             │  动态 Pod    │
                                         │  Deployment  │                             │  Build Agent │
                                         │  更新镜像    │                             │  (临时创建)  │
                                         └──────────────┘                             └──────────────┘
```

---

## 八、关键 API 端点

| 端点 | 方法 | 用途 |
|------|------|------|
| `/api/v1/k8s/cicd/pipeline/run` | POST | 触发流水线构建 |
| `/api/v1/k8s/cicd/pipeline/callback` | POST | Jenkins 最终构建回调 |
| `/api/v1/k8s/cicd/stage/callback` | POST | Jenkins 阶段实时回调 |
| `/api/v1/k8s/cicd/pipeline/sonar-callback` | POST | SonarQube 扫描结果回调 |
| `/api/v1/k8s/cicd/artifact/upload` | POST | 制品上传回调 |
| `/api/v1/k8s/cicd/approval/feishu-callback` | GET/POST | 飞书审批操作回调 |
| `/api/v1/k8s/cicd/callback/build` | POST | 发布单构建回调 |
| `/api/v1/k8s/cicd/release` | POST | 创建发布单 |
| `/api/v1/k8s/cicd/approval/:id` | PUT | 审批操作(通过/拒绝) |
