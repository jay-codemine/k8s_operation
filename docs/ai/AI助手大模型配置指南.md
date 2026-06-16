# AI 助手大模型配置指南

> 本文档说明如何在 K8sOperation 平台中配置 AI 大模型提供商，支持 OpenAI、DeepSeek、智谱GLM、通义千问、Moonshot、NVIDIA NIM 等兼容 OpenAI API 协议的模型。

---

## 一、配置文件位置

```
configs/config.yaml → AIAssistant 配置块
```

---

## 二、必填参数与可选参数速查表

### 2.1 全局配置（AIAssistant 顶层）

| 参数 | 是否必填 | 类型 | 默认值 | 说明 |
|------|:--------:|------|--------|------|
| `Enabled` | **必填** | bool | `false` | 是否启用 AI 助手，`true` = 启用 |
| `DefaultProvider` | **必填** | string | 无 | 默认提供商 ID，必须与某个 Provider 的 ID 一致 |
| `SystemPrompt` | 可选 | string | 内置默认 | 全局 System Prompt，自定义 AI 人设 |
| `ApprovalExpire` | 可选 | int | `30` | 高危操作审批过期时间（分钟） |
| `MaxHistoryRound` | 可选 | int | `20` | 会话最大历史轮数 |

### 2.2 提供商配置（Providers 列表中的每一项）

| 参数 | 是否必填 | 类型 | 默认值 | 说明 |
|------|:--------:|------|--------|------|
| `ID` | **必填** | string | 无 | 提供商唯一标识，如 `openai`、`deepseek`、`nvidia` |
| `Name` | **必填** | string | 无 | 前端显示名称，如 `"OpenAI"`、`"DeepSeek"` |
| `APIKey` | **必填** | string | 无 | API 密钥/Token（**没有此字段则该提供商不会加载**） |
| `BaseURL` | **视情况** | string | `""` | API 地址（见下方说明） |
| `Icon` | 可选 | string | `""` | 前端图标标识 |
| `MaxTokens` | 可选 | int | `4096` | 默认最大 Token 数 |
| `Temperature` | 可选 | float | `0.7` | 生成温度（0~1，越高越随机） |
| `Models` | **必填** | 列表 | 无 | 至少配置一个模型 |

### 2.3 模型配置（Models 列表中的每一项）

| 参数 | 是否必填 | 类型 | 默认值 | 说明 |
|------|:--------:|------|--------|------|
| `ID` | **必填** | string | 无 | 模型 ID，传给 API 的标识（如 `gpt-4o-mini`、`deepseek-chat`） |
| `Name` | **必填** | string | 无 | 前端显示名称 |
| `Description` | 可选 | string | `""` | 模型描述，前端展示用 |
| `Capability` | 可选 | string | `"chat"` | 能力标签：`chat` / `reasoning` / `code` / `vision` |
| `MaxTokens` | 可选 | int | 继承提供商 | 可覆盖提供商级别的 MaxTokens |

---

## 三、BaseURL 配置规则

| 提供商 | BaseURL 是否必填 | 填写值 |
|--------|:----------------:|--------|
| OpenAI（官方直连） | 不填 | 留空 `""` 即走官方 `api.openai.com` |
| OpenAI（代理/中转） | **必填** | 代理地址，如 `https://your-proxy.com/v1` |
| DeepSeek | **必填** | `https://api.deepseek.com/v1` |
| 智谱 GLM | **必填** | `https://open.bigmodel.cn/api/paas/v4` |
| 通义千问 | **必填** | `https://dashscope.aliyuncs.com/compatible-mode/v1` |
| Moonshot (Kimi) | **必填** | `https://api.moonshot.cn/v1` |
| NVIDIA NIM | **必填** | `https://integrate.api.nvidia.com/v1` |

> **规则：除 OpenAI 官方直连外，所有其他提供商都必须填写 BaseURL。**

---

## 四、各提供商 API Key 获取方式

| 提供商 | 申请地址 | Key 格式示例 |
|--------|----------|-------------|
| OpenAI | https://platform.openai.com/api-keys | `sk-proj-xxxxx` |
| DeepSeek | https://platform.deepseek.com/api_keys | `sk-xxxxx` |
| 智谱 AI | https://open.bigmodel.cn/usercenter/apikeys | `xxx.xxx`（两段式） |
| 通义千问 | https://dashscope.console.aliyun.com/apiKey | `sk-xxxxx` |
| Moonshot | https://platform.moonshot.cn/console/api-keys | `sk-xxxxx` |
| NVIDIA NIM | https://build.nvidia.com/explore/discover | `nvapi-xxxxx` |

---

## 五、最小配置示例

### 5.1 只配一个提供商（最简配置）

只需 **4 个必填参数**：`Enabled` + `DefaultProvider` + `APIKey` + `Models[].ID`

```yaml
AIAssistant:
  Enabled: true
  DefaultProvider: "deepseek"

  Providers:
    - ID: "deepseek"
      Name: "DeepSeek"
      APIKey: "sk-你的APIKey"            # ← 必填
      BaseURL: "https://api.deepseek.com/v1"  # ← 国产模型必填
      Models:
        - ID: "deepseek-chat"            # ← 必填：模型 ID
          Name: "DeepSeek V3"            # ← 必填：显示名称
```

### 5.2 配多个提供商

```yaml
AIAssistant:
  Enabled: true
  DefaultProvider: "nvidia"              # 指向默认使用的提供商

  Providers:
    # 提供商 1
    - ID: "nvidia"
      Name: "NVIDIA NIM"
      APIKey: "nvapi-你的APIKey"
      BaseURL: "https://integrate.api.nvidia.com/v1"
      Models:
        - ID: "minimaxai/minimax-m2.7"
          Name: "MiniMax M2.7"

    # 提供商 2
    - ID: "deepseek"
      Name: "DeepSeek"
      APIKey: "sk-你的APIKey"
      BaseURL: "https://api.deepseek.com/v1"
      Models:
        - ID: "deepseek-chat"
          Name: "DeepSeek V3"
        - ID: "deepseek-reasoner"
          Name: "DeepSeek R1"

    # 提供商 3（没有 APIKey 的会被自动跳过，不影响系统）
    - ID: "openai"
      Name: "OpenAI"
      APIKey: ""                         # 空的会被跳过
      Models:
        - ID: "gpt-4o-mini"
          Name: "GPT-4o Mini"
```

---

## 六、配置校验规则

系统启动时会自动校验配置，规则如下：

| 校验规则 | 结果 |
|----------|------|
| `Enabled: false` | AI 助手完全关闭，不加载任何提供商 |
| Provider 的 `APIKey` 为空 | **该提供商被自动跳过**，不影响其他提供商 |
| 所有 Provider 的 APIKey 都为空 | 日志警告，AI 助手自动禁用 |
| `DefaultProvider` 指向的提供商 APIKey 为空 | 系统自动回退到第一个有效提供商 |
| `Models` 列表为空 | 该提供商不会被加载 |

> **重点：只有配置了有效 APIKey 的提供商才会被加载。** 你可以放心地保留多个提供商配置，暂时不用的只需清空 APIKey 即可。

---

## 七、配置完成后的验证步骤

### 7.1 重启后端

```bash
# 编译
go build -o bin/k8soperation.exe ./cmd/k8soperation/
# 或使用 Makefile
make build

# 启动（Windows）
.\bin\k8soperation.exe

# 启动（Linux/Mac）
./bin/k8soperation
```

### 7.2 查看启动日志

成功加载时会输出：

```
[AIAssistant] AI 助手已启用: 2 个提供商, 4 个模型, 默认: nvidia/minimaxai/minimax-m2.7
```

如果所有 APIKey 都为空：

```
[AIAssistant] 警告: 无有效的 AI 提供商配置（检查 APIKey），AI 助手功能将不可用
```

### 7.3 API 验证

```bash
# 检查 AI 状态
curl http://localhost:8080/api/v1/ai/status \
  -H "Authorization: Bearer <你的JWT Token>"

# 期望返回：
# {
#   "data": {
#     "enabled": true,
#     "provider_count": 2,
#     "default_provider": "nvidia",
#     "default_model": "minimaxai/minimax-m2.7"
#   }
# }

# 查看所有可用模型
curl http://localhost:8080/api/v1/ai/models \
  -H "Authorization: Bearer <你的JWT Token>"

# 测试对话
curl -X POST http://localhost:8080/api/v1/ai/quick-ask \
  -H "Authorization: Bearer <你的JWT Token>" \
  -H "Content-Type: application/json" \
  -d '{"message":"你好，请介绍一下你自己"}'
```

---

## 八、常见问题

### Q1: 我只想用一个国产模型，最少需要填什么？

**3 个必填：** `APIKey` + `BaseURL` + `Models[].ID`

```yaml
AIAssistant:
  Enabled: true
  DefaultProvider: "deepseek"
  Providers:
    - ID: "deepseek"
      Name: "DeepSeek"
      APIKey: "sk-你的key"
      BaseURL: "https://api.deepseek.com/v1"
      Models:
        - ID: "deepseek-chat"
          Name: "DeepSeek V3"
```

### Q2: 提供商 APIKey 填错了会怎样？

不影响系统启动。系统会正常加载该提供商，但用户使用时 AI 对话会返回认证错误。其他提供商和平台功能不受影响。

### Q3: 可以随时新增提供商吗？

可以。在 `Providers` 列表中添加新的提供商配置，重启后端即可生效。只要兼容 OpenAI API 协议，填写 `APIKey` + `BaseURL` + `Models` 即可接入。

### Q4: MaxTokens 和 Temperature 不填会怎样？

系统有默认值：`MaxTokens = 4096`、`Temperature = 0.7`。不填就用默认值，大多数场景够用。

### Q5: 如何让某个提供商暂时下线？

把该提供商的 `APIKey` 清空即可，系统启动时会自动跳过。

---

## 九、参数总览图

```
AIAssistant:
  Enabled: true                    ← [必填] 启用开关
  DefaultProvider: "nvidia"        ← [必填] 默认提供商 ID
  SystemPrompt: "..."              ← [可选] 自定义 AI 人设
  ApprovalExpire: 30               ← [可选] 审批过期时间(分)
  MaxHistoryRound: 20              ← [可选] 历史轮数上限

  Providers:
    - ID: "nvidia"                 ← [必填] 提供商唯一标识
      Name: "NVIDIA NIM"           ← [必填] 显示名称
      Icon: "nvidia"               ← [可选] 图标标识
      APIKey: "nvapi-xxx"          ← [必填] API 密钥 ★ 最关键
      BaseURL: "https://..."       ← [视情况] 国产模型必填
      MaxTokens: 4096              ← [可选] 默认 4096
      Temperature: 0.7             ← [可选] 默认 0.7
      Models:
        - ID: "minimaxai/minimax-m2.7"  ← [必填] 模型 ID
          Name: "MiniMax M2.7"          ← [必填] 显示名称
          Description: "..."            ← [可选] 模型描述
          Capability: "chat"            ← [可选] 能力标签
          MaxTokens: 8192               ← [可选] 覆盖提供商默认值
```
