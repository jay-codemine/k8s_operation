# License 授权机制使用说明

## 概述

平台采用 **Ed25519 离线验签** + **机器码绑定** + **过期锁定** 的 License 授权机制。

- 私钥仅软件作者持有，签发 License
- 公钥内置于平台代码中，启动时自动校验
- License 绑定机器码，不可跨机器使用
- 到期自动锁定，需重新签发

---

## 首次初始化

### 1. 生成密钥对（只需一次）

```bash
go run cmd/license-gen/main.go -init-keys -key-dir configs/license-keys
```

输出：
```
configs/license-keys/private.key   ← 私钥（保密，绝不外发）
configs/license-keys/public.key    ← 公钥（已内置于代码中）
```

> ⚠️ **私钥 `private.key` 绝对不能提交 Git、不能外传。丢失后无法恢复。**

### 2. 将公钥写入代码（新项目初始化时）

公钥 hex 输出后，写入 `pkg/license/key.go` 的 `embeddedPublicKeyHex` 常量。

---

## 签发 License

### 1. 获取目标机器的机器码

在目标服务器上执行：

```bash
go run cmd/license-gen/main.go -machine-id
```

输出示例：`A670-1E2E-BCC8-C7BF`

### 2. 签发

```bash
go run cmd/license-gen/main.go \
  -key configs/license-keys/private.key \
  -licensee "客户名称" \
  -machine "机器码" \
  -expire 2027-12-31 \
  -edition enterprise
```

参数说明：

| 参数 | 必填 | 说明 |
|------|:---:|------|
| `-key` | ✅ | 私钥文件路径 |
| `-licensee` | ✅ | 被授权方名称 |
| `-machine` | ✅ | 绑定的机器码 |
| `-expire` | ✅ | 到期日期，格式 `YYYY-MM-DD`，当天 23:59:59 过期 |
| `-edition` | ❌ | 版本：`enterprise`（企业版）/ `standard`（标准版）/ `trial`（试用版），默认 `enterprise` |
| `-out` | ❌ | 输出文件路径，留空则打印到控制台 |

### 3. 激活

将输出的 `K8SOP-LICENSE.xxxxx` 文本完整复制给客户，在登录页粘贴激活。

---

## License 格式

```
K8SOP-LICENSE.{Base64JSON载荷}.{Base64Ed25519签名}
```

- **载荷（Payload）**：JSON 格式，包含客户名称、版本、机器码、签发时间、到期时间
- **签名（Signature）**：Ed25519 对载荷的签名，平台启动时用内置公钥验签

---

## 校验逻辑（自动）

平台启动时自动执行：

1. 解码 License 文本
2. 用内置公钥验证 Ed25519 签名
3. 检查机器码是否匹配
4. 检查是否过期

任一校验失败 → 平台进入未激活状态，仅开放激活页面。

---

## 安全提醒

1. **私钥 `private.key` 绝不外传**，仅软件作者持有
2. 每台机器一个 License，不可共享
3. License 到期后需重新签发
4. 更换服务器机器码会变，需重新签发
