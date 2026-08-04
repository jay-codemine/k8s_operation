// kubeconfig-keytool 校验并轮换 kube_cluster.kube_config 的 AES 加密密钥。
//
// 背景：kube_config 以 ENC:base64(AES-256-GCM) 落库，密钥来自配置项
// Security.KubeConfigEncryptKey。当本地开发与集群部署使用了不同的密钥时，
// 同一份数据在一端可解、另一端报 "decryption failed"。本工具用候选密钥逐行
// 试解密，定位每行数据实际使用的密钥，并可将全部数据统一重加密为目标密钥。
//
// 用法（先只读体检，确认无误后再加 -apply 落库）：
//
//	go run ./cmd/kubeconfig-keytool -dsn "root:pwd@tcp(127.0.0.1:3306)/k8s-platform" \
//	    -keys "key-a,key-b"
//	go run ./cmd/kubeconfig-keytool -dsn "..." -keys "key-a,key-b" -rekey-to "key-a" -apply
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"k8soperation/pkg/utils"
)

type row struct {
	id      int64
	name    string
	isDel   uint8
	payload string
}

func main() {
	dsn := flag.String("dsn", "", "MySQL DSN，例如 root:pwd@tcp(127.0.0.1:3306)/k8s-platform")
	keys := flag.String("keys", "", "候选密钥列表，逗号分隔，按顺序尝试解密")
	rekeyTo := flag.String("rekey-to", "", "目标密钥：将所有可解密的行统一重加密为该密钥（须同时出现在 -keys 中或直接可用）")
	apply := flag.Bool("apply", false, "真正写库；不加此参数只做只读体检")
	flag.Parse()

	if *dsn == "" || *keys == "" {
		fmt.Fprintln(os.Stderr, "必须提供 -dsn 与 -keys")
		flag.Usage()
		os.Exit(2)
	}

	candidates := splitNonEmpty(*keys)
	if len(candidates) == 0 {
		fmt.Fprintln(os.Stderr, "-keys 未解析出任何有效密钥")
		os.Exit(2)
	}

	db, err := sql.Open("mysql", *dsn)
	if err != nil {
		fatal("连接数据库失败: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		fatal("数据库不可达: %v", err)
	}

	rows, err := loadRows(db)
	if err != nil {
		fatal("读取 kube_cluster 失败: %v", err)
	}
	fmt.Printf("共读取 %d 行 kube_cluster\n\n", len(rows))

	// 只读体检：定位每行实际使用的密钥
	matched := make(map[int64]string, len(rows))
	for _, r := range rows {
		label := fmt.Sprintf("id=%d name=%q is_del=%d", r.id, r.name, r.isDel)

		if !utils.IsEncrypted(r.payload) {
			fmt.Printf("%s → 明文（未加密），任何密钥均可读\n", label)
			matched[r.id] = ""
			continue
		}

		hit := ""
		for _, k := range candidates {
			plain, derr := utils.NewCryptoService(k).DecryptKubeConfig(r.payload)
			if derr == nil && looksLikeKubeConfig(plain) {
				hit = k
				break
			}
		}
		if hit == "" {
			fmt.Printf("%s → ✗ 候选密钥均无法解密（数据可能由第三把密钥加密或已损坏）\n", label)
			continue
		}
		matched[r.id] = hit
		fmt.Printf("%s → ✓ 由密钥 %s 加密\n", label, mask(hit))
	}

	if *rekeyTo == "" {
		fmt.Println("\n未指定 -rekey-to，仅体检，未做任何修改。")
		return
	}

	// 轮换：统一重加密为目标密钥
	fmt.Printf("\n目标密钥: %s\n", mask(*rekeyTo))
	target := utils.NewCryptoService(*rekeyTo)
	changed, skipped, failed := 0, 0, 0

	for _, r := range rows {
		srcKey, ok := matched[r.id]
		if !ok {
			fmt.Printf("id=%d → 跳过（无法解密，不敢覆盖）\n", r.id)
			failed++
			continue
		}
		if srcKey == *rekeyTo {
			fmt.Printf("id=%d → 跳过（已是目标密钥）\n", r.id)
			skipped++
			continue
		}

		var plain string
		if srcKey == "" {
			plain = r.payload // 原本是明文
		} else {
			plain, err = utils.NewCryptoService(srcKey).DecryptKubeConfig(r.payload)
			if err != nil {
				fmt.Printf("id=%d → 解密失败，跳过: %v\n", r.id, err)
				failed++
				continue
			}
		}

		encrypted, err := target.EncryptKubeConfig(plain)
		if err != nil {
			fmt.Printf("id=%d → 重加密失败，跳过: %v\n", r.id, err)
			failed++
			continue
		}
		// 自校验：确保新密文能用目标密钥还原出同样的明文，避免写入不可读数据
		if back, verr := target.DecryptKubeConfig(encrypted); verr != nil || back != plain {
			fmt.Printf("id=%d → 重加密自校验未通过，跳过\n", r.id)
			failed++
			continue
		}

		if !*apply {
			fmt.Printf("id=%d → [dry-run] 将由 %s 重加密为目标密钥\n", r.id, mask(srcKey))
			changed++
			continue
		}
		if _, err := db.Exec("UPDATE kube_cluster SET kube_config = ? WHERE id = ?", encrypted, r.id); err != nil {
			fmt.Printf("id=%d → 写库失败: %v\n", r.id, err)
			failed++
			continue
		}
		fmt.Printf("id=%d → 已重加密并写库\n", r.id)
		changed++
	}

	mode := "dry-run（未写库）"
	if *apply {
		mode = "已写库"
	}
	fmt.Printf("\n完成[%s]：变更 %d，跳过 %d，失败 %d\n", mode, changed, skipped, failed)
	if failed > 0 {
		os.Exit(1)
	}
}

func loadRows(db *sql.DB) ([]row, error) {
	rs, err := db.Query("SELECT id, cluster_name, is_del, kube_config FROM kube_cluster ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rs.Close()

	var out []row
	for rs.Next() {
		var r row
		if err := rs.Scan(&r.id, &r.name, &r.isDel, &r.payload); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rs.Err()
}

// looksLikeKubeConfig 用于区分「解密成功」与「AES-GCM 恰好通过但内容是乱码」。
// GCM 自带认证，错误密钥几乎必然报错，这里作为二次确认。
func looksLikeKubeConfig(s string) bool {
	return strings.Contains(s, "apiVersion") && strings.Contains(s, "clusters")
}

func splitNonEmpty(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// mask 避免把完整密钥打进日志
func mask(k string) string {
	if len(k) <= 4 {
		return fmt.Sprintf("****(len=%d)", len(k))
	}
	return fmt.Sprintf("%s****(len=%d)", k[:4], len(k))
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
