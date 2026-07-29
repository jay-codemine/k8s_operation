// license-gen 平台 License 签发工具（仅软件作者持有，私钥绝不外发）
//
// 用法：
//
//	1) 初始化密钥对（只需一次，私钥务必妥善保管、不要提交 git）：
//	   go run ./cmd/license-gen -init-keys -key-dir configs/license-keys
//	   输出 private.key（私钥，保密）与 public.key（公钥，内容需写入 pkg/license/key.go）
//
//	2) 查看本机机器码（给客户的机器生成时，让客户在其服务器上执行平台的 -machine-id）：
//	   go run ./cmd/license-gen -machine-id
//
//	3) 签发 License：
//	   go run ./cmd/license-gen -key configs/license-keys/private.key ^
//	     -licensee "客户名称" -machine "XXXX-XXXX-XXXX-XXXX" ^
//	     -expire 2027-12-31 -edition enterprise -out customer.lic
package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"k8soperation/pkg/license"
)

func main() {
	var (
		initKeys  = flag.Bool("init-keys", false, "生成 Ed25519 密钥对")
		keyDir    = flag.String("key-dir", "configs/license-keys", "密钥对输出目录（配合 -init-keys）")
		machineID = flag.Bool("machine-id", false, "输出本机机器码")

		keyFile  = flag.String("key", "", "私钥文件路径（签发时必填）")
		licensee = flag.String("licensee", "", "被授权方（客户名称）")
		machine  = flag.String("machine", "", "绑定的机器码")
		expire   = flag.String("expire", "", "到期日期，格式 2027-12-31（当天 23:59:59 过期）")
		edition  = flag.String("edition", "enterprise", "版本: enterprise/standard/trial")
		out      = flag.String("out", "", "License 输出文件（缺省打印到控制台）")
	)
	flag.Parse()

	switch {
	case *initKeys:
		doInitKeys(*keyDir)
	case *machineID:
		fmt.Println(license.MachineID())
	default:
		doSign(*keyFile, *licensee, *machine, *expire, *edition, *out)
	}
}

// doInitKeys 生成 Ed25519 密钥对
func doInitKeys(dir string) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	fatalIf(err, "生成密钥对失败")

	fatalIf(os.MkdirAll(dir, 0o700), "创建密钥目录失败")

	privPath := filepath.Join(dir, "private.key")
	pubPath := filepath.Join(dir, "public.key")

	// 私钥以 hex(seed) 形式保存（32 字节 seed 足以恢复完整私钥）
	fatalIf(os.WriteFile(privPath, []byte(hex.EncodeToString(priv.Seed())+"\n"), 0o600), "写入私钥失败")
	fatalIf(os.WriteFile(pubPath, []byte(hex.EncodeToString(pub)+"\n"), 0o644), "写入公钥失败")

	fmt.Println("✅ 密钥对已生成：")
	fmt.Println("   私钥（保密，勿提交 git）:", privPath)
	fmt.Println("   公钥:", pubPath)
	fmt.Println()
	fmt.Println("公钥 hex（请写入 pkg/license/key.go 的 embeddedPublicKeyHex）：")
	fmt.Println(hex.EncodeToString(pub))
}

// doSign 签发 License
func doSign(keyFile, licensee, machine, expire, edition, out string) {
	if keyFile == "" || licensee == "" || machine == "" || expire == "" {
		fmt.Println("缺少参数。签发 License 需要: -key -licensee -machine -expire")
		flag.Usage()
		os.Exit(1)
	}

	// 读取私钥（hex seed）
	data, err := os.ReadFile(keyFile)
	fatalIf(err, "读取私钥失败")
	seed, err := hex.DecodeString(strings.TrimSpace(string(data)))
	fatalIf(err, "私钥格式错误（应为 hex）")
	if len(seed) != ed25519.SeedSize {
		fatal(fmt.Sprintf("私钥长度错误: 期望 %d 字节 seed，实际 %d", ed25519.SeedSize, len(seed)))
	}
	priv := ed25519.NewKeyFromSeed(seed)

	// 解析到期日期（当天 23:59:59，本地时区）
	expDay, err := time.ParseInLocation("2006-01-02", expire, time.Local)
	fatalIf(err, "到期日期格式错误，应为 2027-12-31")
	expireAt := expDay.Add(24*time.Hour - time.Second)

	payload := license.Payload{
		Licensee:  licensee,
		Edition:   edition,
		MachineID: strings.ToUpper(strings.TrimSpace(machine)),
		IssuedAt:  time.Now().Unix(),
		ExpireAt:  expireAt.Unix(),
	}
	raw, err := json.Marshal(payload)
	fatalIf(err, "序列化载荷失败")

	sig := ed25519.Sign(priv, raw)
	text := license.Encode(raw, sig)

	fmt.Println("✅ License 签发成功：")
	fmt.Println("   被授权方:", payload.Licensee)
	fmt.Println("   版本:", payload.Edition)
	fmt.Println("   机器码:", payload.MachineID)
	fmt.Println("   到期时间:", expireAt.Format("2006-01-02 15:04:05"))
	fmt.Println()

	if out != "" {
		fatalIf(os.WriteFile(out, []byte(text+"\n"), 0o644), "写入 License 文件失败")
		fmt.Println("License 已写入:", out)
	} else {
		fmt.Println(text)
	}
}

func fatalIf(err error, msg string) {
	if err != nil {
		fatal(msg + ": " + err.Error())
	}
}

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, "❌ "+msg)
	os.Exit(1)
}
