# Prometheus 告警体系架构与使用指南

## 一、整体架构概览

本平台存在 **两条独立的告警通知路径**，各自独立工作，互不冲突：

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        告警通知双链路架构                                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  【链路A】Kubernetes 原生链路 (Alertmanager)                                 │
│  ────────────────────────────────────────────                               │
│  PrometheusRule CRD → Prometheus 评估 → Alertmanager 路由 → DingTalk Webhook│
│                                                                             │
│  配置位置: kube-prometheus-stack Helm values.yaml                            │
│  适用场景: 集群基础设施告警（节点宕机、etcd异常、kubelet不可用等）               │
│  特点: K8s 原生、高可靠、与平台无关                                           │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  【链路B】平台 AlertEvalWorker 链路                                          │
│  ────────────────────────────────────                                       │
│  平台DB规则 → Worker查询Prometheus → 状态机评估 → 通知渠道/路由策略 → 发送    │
│                                                                             │
│  配置位置: 平台 Web UI / API                                                 │
│  适用场景: 业务告警、自定义 PromQL、多渠道分发（钉钉/飞书/邮件/Webhook）        │
│  特点: 灵活路由、UI 可视化管理、支持 YAML 批量导入                             │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 二、链路A：Alertmanager 原生告警（集群层面）

### 2.1 组件版本

| 组件 | 版本 | 镜像 |
|------|------|------|
| Prometheus Operator | v0.90.1 | `registry.cn-hangzhou.aliyuncs.com/k8s-gos/prom:v0.90.1` |
| Prometheus | v3.11.3-distroless | `registry.cn-hangzhou.aliyuncs.com/k8s-gos/prom:v3.11.3-distroless` |
| Alertmanager | v0.32.1 | `registry.cn-hangzhou.aliyuncs.com/k8s-gos/prom:v0.32.1` |
| Grafana | v13.0.1-security-01 | `registry.cn-hangzhou.aliyuncs.com/k8s-gos/prom:13.0.1-security-01` |
| kube-state-metrics | v2.18.0 | `registry.cn-hangzhou.aliyuncs.com/k8s-gos/prom:v2.18.0` |
| node-exporter | v1.11.1 | `registry.cn-hangzhou.aliyuncs.com/k8s-gos/prom:v1.11.1` |

### 2.2 Alertmanager 路由配置（你的 values.yaml）

```yaml
alertmanager:
  config:
    global:
      resolve_timeout: 5m

    route:
      group_by: [namespace, alertname]
      group_wait: 30s          # 首次告警等待
      group_interval: 5m       # 同组告警间隔
      repeat_interval: 1h      # 重复告警间隔
      receiver: dingtalk       # 默认接收器

    receivers:
      - name: dingtalk
        webhook_configs:
          - url: http://prometheus-webhook-dingtalk.monitoring.svc.cluster.local:8060/dingtalk/ops/send
            send_resolved: true

    inhibit_rules:
      - source_matchers: [severity = critical]
        target_matchers: [severity =~ warning|info]
        equal: [namespace, alertname]
```

### 2.3 链路A 工作流程

```
PrometheusRule(CRD) → Prometheus 定时评估(1min)
                         ↓ 触发
                    Alertmanager
                         ↓ 路由匹配
                    receiver: dingtalk
                         ↓
              prometheus-webhook-dingtalk
                         ↓
                    钉钉群消息
```

**关键点**：
- 这条链路的告警规则是 **K8s CRD**（`PrometheusRule`），部署在集群中
- `kube-prometheus-stack` 自带大量内置规则（节点/etcd/kubelet/apiserver 等）
- 所有告警默认路由到 `dingtalk` 接收器
- 如需分流（如 critical 发钉钉、warning 发邮件），修改 `routes` 子路由

### 2.4 如何添加自定义 PrometheusRule

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: my-app-rules
  namespace: monitoring
  labels:
    release: prometheus   # 必须有此 label 才被 Prometheus 发现
spec:
  groups:
    - name: my-app
      rules:
        - alert: MyAppDown
          expr: up{job="my-app"} == 0
          for: 5m
          labels:
            severity: critical
          annotations:
            summary: "应用 {{ $labels.instance }} 宕机"
```

---

## 三、链路B：平台 AlertEvalWorker（应用层面）

### 3.1 工作机制

```
┌───────────────────────────────────────────────────────┐
│              AlertEvalWorker (30s 轮询)                 │
├───────────────────────────────────────────────────────┤
│  1. 从 DB 加载 enabled=1 的告警规则                     │
│  2. 对每条规则执行 PromQL 查询 Prometheus               │
│  3. 解析结果，驱动状态机:                               │
│     normal → pending → firing → resolved              │
│  4. 状态变更时创建告警事件、发送通知                     │
└───────────────────────────────────────────────────────┘
```

### 3.2 状态机流转

```
条件首次满足          持续 ≥ duration        条件恢复
  normal ──────→ pending ──────────→ firing ──────→ resolved
     ↑                    条件恢复      ↑              │
     │                   ┌──────────────┘              │
     │                   │                             │
     └───────────────────┴─────────────────────────────┘
                    重新进入 normal
```

### 3.3 通知发送决策流程

```
sendNotification(rule, event)
    │
    ├─ rule.NotifyChannels 不为空？
    │       ├─ 是 → 使用规则直接绑定的渠道
    │       └─ 否 → 进入路由策略匹配
    │
    ├─ ResolveRoutePolicyChannels(rule)
    │       ├─ 遍历 priority ASC 的路由策略
    │       │     └─ 按 severity/group/labels 匹配
    │       │         匹配成功 → 返回该策略的 channel_ids
    │       │
    │       └─ 所有非默认策略均不匹配
    │             └─ 回退到 IsDefault=true 的兜底策略 ← 你的钉钉在这里
    │                   └─ 返回兜底策略的 channel_ids
    │
    └─ 根据 channel_ids 查找渠道记录 → 调用对应发送器
```

**这就是为什么你"没有配置钉钉却收到钉钉通知"的原因**：
数据库中有一条 `is_default=1` 的路由策略，其 `channel_ids` 指向了钉钉渠道。

---

## 四、通知渠道与路由策略

### 4.1 数据模型

```
┌──────────────────────┐     ┌────────────────────────────┐
│ MonitorAlertRule      │     │ MonitorNotifyChannel       │
├──────────────────────┤     ├────────────────────────────┤
│ notify_channels: ""  │     │ id: 1                      │
│ (空=走路由策略)       │     │ name: "运维钉钉群"         │
│                      │     │ type: "dingtalk"           │
│ OR                   │     │ webhook_url: "https://..."  │
│                      │     │ enabled: true              │
│ notify_channels:"1,2"│     └────────────────────────────┘
│ (直接绑定渠道ID)     │
└──────────────────────┘
         │
         │ 未绑定时
         ▼
┌───────────────────────────────────────┐
│ MonitorNotifyRoutePolicy              │
├───────────────────────────────────────┤
│ name: "默认钉钉兜底"                   │
│ is_default: true                      │
│ channel_ids: "1"                      │
│ priority: 999                         │
├───────────────────────────────────────┤
│ name: "critical走飞书"                 │
│ is_default: false                     │
│ severities: "critical"               │
│ channel_ids: "2"                      │
│ priority: 10                          │
└───────────────────────────────────────┘
```

### 4.2 路由策略匹配优先级

1. **规则直接绑定** (`rule.notify_channels != ""`) → 最高优先
2. **非默认路由策略** (按 `priority ASC` 顺序匹配 severity/group/labels)
3. **默认兜底策略** (`is_default=true`) → 兜底，确保不漏发

---

## 五、YAML 批量导入告警规则

### 5.1 API 接口

```
POST /api/v1/monitoring/alert-rule/import-yaml
```

### 5.2 请求参数

```json
{
  "yaml": "<PrometheusRule 格式 YAML 内容>",
  "datasource_id": 1,
  "overwrite": false,

  // ========== 通知渠道绑定方式（4选1或组合使用） ==========

  // 方式1: 全局默认（所有导入规则统一绑定同一渠道）
  "default_notify_channels": "1",

  // 方式2: 按组指定不同渠道 ⭐推荐
  "group_channels": {
    "infra":    "1",       // 基础设施组 → 钉钉渠道(id=1)
    "app":      "2,3",     // 应用组 → 飞书(id=2) + 邮件(id=3)
    "database": "1,2"      // 数据库组 → 钉钉+飞书
  },

  // 方式3: 启用路由策略自动匹配
  "auto_route": true,

  // 方式4: 不指定任何渠道，运行时由路由策略兜底
  // (所有字段不填，告警触发时走 sendNotification 的路由策略匹配)
}
```

### 5.3 渠道决策优先级（从高到低）

| 优先级 | 来源 | 说明 |
|--------|------|------|
| 1 | 规则 annotations 中的 `notify_channels` | YAML 中 per-rule 指定 |
| 2 | 请求参数 `group_channels[组名]` | per-group 指定 |
| 3 | 请求参数 `default_notify_channels` | 全局默认 |
| 4 | `auto_route` + 路由策略匹配 | 自动匹配 |
| 5 | 空（不绑定） | 运行时由 `sendNotification` 路由策略兜底 |

### 5.4 YAML 示例（带 per-rule 渠道指定）

```yaml
groups:
  - name: infra
    rules:
      - alert: NodeDown
        expr: up{job="node-exporter"} == 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "节点 {{ $labels.instance }} 宕机"
          description: "节点已超过 5 分钟无响应"
          notify_channels: "1,2"    # ← 这条规则发 渠道1 + 渠道2

      - alert: NodeHighCPU
        expr: 100 - (avg by(instance)(rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > 85
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "节点 {{ $labels.instance }} CPU 使用率 {{ $value }}%"
          notify_channels: "1"      # ← 这条规则只发 渠道1

  - name: app
    rules:
      - alert: PodCrashLooping
        expr: rate(kube_pod_container_status_restarts_total[15m]) > 0
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Pod {{ $labels.namespace }}/{{ $labels.pod }} 频繁重启"
          # 不指定 notify_channels → 走 group_channels["app"] 或 兜底策略
```

### 5.5 导入后批量追加/修改渠道

```
POST /api/v1/monitoring/alert-rule/batch-bind-channels
```

```json
{
  "rule_ids": [1, 2, 3, 4, 5],
  "channel_ids": "1,2",
  "mode": "append"     // append=追加, replace=覆盖
}
```

---

## 六、两条链路对比与选择

| 维度 | 链路A (Alertmanager) | 链路B (平台 Worker) |
|------|---------------------|-------------------|
| 规则存储 | K8s CRD (PrometheusRule) | 平台数据库 |
| 评估引擎 | Prometheus 原生 | AlertEvalWorker (Go) |
| 通知方式 | Alertmanager → webhook-dingtalk | 平台通知渠道（多种类型） |
| 管理方式 | kubectl / Helm values | Web UI / API |
| 适合场景 | 基础设施监控 | 业务告警 + 灵活分发 |
| 规则来源 | Helm chart 自带 + 手动 CRD | 平台 UI 创建 / YAML 导入 |
| 通知分流 | Alertmanager routes 配置 | 路由策略 + 多渠道 |

### 推荐使用方式

- **基础设施告警**（节点/etcd/apiserver）：保持链路A，Alertmanager 原生处理
- **业务/自定义告警**：通过平台导入 YAML，利用路由策略灵活分发到不同群

---

## 七、常见问题

### Q1: 为什么没配置钉钉通知却收到了？

**答**：平台有一条 `is_default=true` 的路由策略作为兜底，会把所有未绑定渠道的告警自动发送到默认钉钉渠道。这是**设计行为**，确保告警不会被静默丢失。

如需取消某条规则的通知：在规则上设置 `notify_channels = "none"` 或禁用该规则。

### Q2: 告警事件摘要显示 `{{ $labels.instance }}` 没有替换？

**答**：这是历史旧事件（在模板渲染功能上线前创建的）。系统已添加后处理渲染逻辑，刷新告警事件列表页即可看到渲染后的值。新创建的事件不会有此问题。

### Q3: 平台规则和 Alertmanager 规则会冲突吗？

**答**：不会。两条链路完全独立：
- 平台规则存在 MySQL 中，由 `AlertEvalWorker` 评估
- Alertmanager 规则存在 K8s CRD 中，由 Prometheus 评估
- 如果同一条 PromQL 在两边都配了，会收到两次通知

### Q4: 如何查看当前生效的路由策略和通知渠道？

```
# 路由策略列表
GET /api/v1/monitoring/notify-route

# 通知渠道列表
GET /api/v1/monitoring/notify-channel
```

### Q5: values.yaml 中的钉钉 webhook 和平台的钉钉渠道是什么关系？

| | values.yaml 中的 webhook | 平台通知渠道 |
|--|--|--|
| 服务 | prometheus-webhook-dingtalk (K8s Service) | 平台后端 Go 代码直接调用 |
| 用途 | Alertmanager → 钉钉 | AlertEvalWorker → 钉钉 |
| 配置方式 | Helm values | 平台 UI "通知渠道管理" |
| Webhook 地址 | 写在 values.yaml receivers 里 | 写在 DB notify_channel 表里 |

两者可以指向**同一个钉钉群机器人**，也可以指向**不同群**。

---

## 八、快速操作指南

### 8.1 添加新的钉钉通知群

1. 在钉钉群创建自定义机器人，获取 Webhook URL
2. 平台 UI → 监控告警 → 通知渠道 → 新建钉钉渠道 → 填入 Webhook URL
3. 记录渠道 ID（如 `id=3`）

### 8.2 批量导入规则并指定渠道

```bash
curl -X POST http://localhost:8080/api/v1/monitoring/alert-rule/import-yaml \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "yaml": "<你的 PrometheusRule YAML>",
    "datasource_id": 1,
    "group_channels": {
      "infra": "1",
      "app": "3"
    }
  }'
```

### 8.3 修改默认兜底策略的目标渠道

```bash
# 查看当前路由策略
curl http://localhost:8080/api/v1/monitoring/notify-route \
  -H "Authorization: Bearer <token>"

# 更新默认策略的 channel_ids
curl -X PUT http://localhost:8080/api/v1/monitoring/notify-route/<id> \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "channel_ids": "1,3",
    "is_default": true
  }'
```

---

## 九、API 接口一览

| 模块 | 方法 | 路径 | 说明 |
|------|------|------|------|
| 数据源 | GET | `/api/v1/monitoring/datasource` | 列表 |
| 数据源 | POST | `/api/v1/monitoring/datasource/test` | 连通性测试 |
| 告警规则 | GET | `/api/v1/monitoring/alert-rule` | 列表 |
| 告警规则 | POST | `/api/v1/monitoring/alert-rule` | 创建 |
| 告警规则 | POST | `/api/v1/monitoring/alert-rule/import-yaml` | YAML批量导入 |
| 告警规则 | GET | `/api/v1/monitoring/alert-rule/export-yaml` | YAML导出 |
| 告警规则 | POST | `/api/v1/monitoring/alert-rule/batch-bind-channels` | 批量绑定渠道 |
| 告警事件 | GET | `/api/v1/monitoring/alert-event` | 事件列表 |
| 告警事件 | GET | `/api/v1/monitoring/alert-event/stats` | 统计 |
| 通知渠道 | GET | `/api/v1/monitoring/notify-channel` | 渠道列表 |
| 通知渠道 | POST | `/api/v1/monitoring/notify-channel/:id/test` | 测试发送 |
| 路由策略 | GET | `/api/v1/monitoring/notify-route` | 策略列表 |
| 路由策略 | POST | `/api/v1/monitoring/notify-route` | 创建策略 |
| 静默规则 | GET | `/api/v1/monitoring/silence-rule` | 静默列表 |
| 抑制规则 | GET | `/api/v1/monitoring/inhibit-rule` | 抑制列表 |
