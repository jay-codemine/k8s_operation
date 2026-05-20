//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dsn := "root:123456@tcp(localhost:3306)/k8s-platform?charset=utf8&parseTime=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("连接失败:", err)
	}
	defer db.Close()

	fmt.Println("========== 诊断：kube_cluster 表数据 ==========")
	rows, err := db.Query("SELECT id, cluster_name, status, is_del, deleted_at FROM kube_cluster ORDER BY id")
	if err != nil {
		log.Fatal("查询失败:", err)
	}
	defer rows.Close()

	fmt.Printf("%-5s %-25s %-8s %-8s %-15s\n", "ID", "ClusterName", "Status", "IsDel", "DeletedAt")
	fmt.Println("-----------------------------------------------------------")

	var needFix []int64
	for rows.Next() {
		var id int64
		var name string
		var status, isDel int
		var deletedAt int64
		if err := rows.Scan(&id, &name, &status, &isDel, &deletedAt); err != nil {
			log.Fatal(err)
		}
		marker := ""
		if isDel == 0 && deletedAt != 0 {
			marker = " <-- 异常! is_del=0 但 deleted_at!=0"
			needFix = append(needFix, id)
		}
		fmt.Printf("%-5d %-25s %-8d %-8d %-15d%s\n", id, name, status, isDel, deletedAt, marker)
	}

	fmt.Println()
	if len(needFix) == 0 {
		fmt.Println("✅ 数据一致，无需修复。")
		fmt.Println("如果健康中心仍只显示1个集群，请重新编译部署后端（已将查询条件从 deleted_at=0 改为 is_del=0）。")
	} else {
		fmt.Printf("⚠️  发现 %d 条异常数据，正在自动修复...\n", len(needFix))
		result, err := db.Exec("UPDATE kube_cluster SET deleted_at = 0 WHERE is_del = 0 AND deleted_at != 0")
		if err != nil {
			log.Fatal("修复失败:", err)
		}
		affected, _ := result.RowsAffected()
		fmt.Printf("✅ 修复完成！影响 %d 行。重启后端即可生效。\n", affected)
	}
}
