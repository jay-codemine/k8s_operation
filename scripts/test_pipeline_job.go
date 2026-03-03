// 测试 Jenkins Pipeline Job 的支持情况
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"k8soperation/pkg/jenkins"
)

func main() {
	// 从环境变量或配置文件获取 Jenkins 配置
	jenkinsURL := os.Getenv("JENKINS_URL")
	if jenkinsURL == "" {
		jenkinsURL = "http://1.117.227.207:8080" // 默认 URL
	}
	username := os.Getenv("JENKINS_USERNAME")
	if username == "" {
		username = "admin"
	}
	apiToken := os.Getenv("JENKINS_API_TOKEN")
	if apiToken == "" {
		log.Fatal("请设置 JENKINS_API_TOKEN 环境变量")
	}

	// 创建 Jenkins 客户端
	client := jenkins.NewClient(jenkinsURL, username, apiToken)

	ctx := context.Background()

	// 测试连接
	if err := client.CheckConnection(ctx); err != nil {
		log.Fatalf("Jenkins 连接失败: %v", err)
	}
	fmt.Println("✅ Jenkins 连接成功")

	// 获取 Job 信息
	jobName := os.Args[1]
	if jobName == "" {
		log.Fatal("请提供 Job 名称作为参数")
	}

	jobInfo, err := client.GetJobInfo(ctx, jobName)
	if err != nil {
		log.Fatalf("获取 Job 信息失败: %v", err)
	}

	fmt.Printf("📋 Job 信息:\n")
	fmt.Printf("  名称: %s\n", jobInfo.Name)
	fmt.Printf("  URL: %s\n", jobInfo.URL)
	fmt.Printf("  类型: %s\n", jobInfo.Class)
	fmt.Printf("  可构建: %t\n", jobInfo.Buildable)
	fmt.Printf("  最新构建号: %d\n", jobInfo.NextBuildNumber-1)

	// 检查是否为参数化 Job
	isParameterized := false
	if jobInfo.Property != nil {
		for _, prop := range jobInfo.Property {
			if prop.Class == "hudson.model.ParametersDefinitionProperty" {
				isParameterized = true
				fmt.Printf("  参数化属性: %s\n", prop.Class)
				break
			}
		}
	}
	fmt.Printf("  是否参数化: %t\n", isParameterized)

	// 检查是否为 Pipeline Job
	isPipelineJob := false
	if jobInfo.Class != "" {
		isPipelineJob = containsPipelineJobClass(jobInfo.Class)
	}
	fmt.Printf("  是否 Pipeline Job: %t\n", isPipelineJob)

	// 测试触发构建（不带参数）
	fmt.Println("\n🧪 测试触发无参数构建...")
	result, err := client.TriggerBuild(ctx, jobName, nil)
	if err != nil {
		fmt.Printf("❌ 无参数构建失败: %v\n", err)
	} else {
		fmt.Printf("✅ 无参数构建触发成功: QueueID=%d, URL=%s\n", result.QueueID, result.QueueURL)
	}

	// 测试触发构建（带参数）
	fmt.Println("\n🧪 测试触发带参数构建...")
	params := map[string]string{
		"BRANCH": "main",
		"ENV":    "dev",
	}
	result, err = client.TriggerBuild(ctx, jobName, params)
	if err != nil {
		fmt.Printf("❌ 带参数构建失败: %v\n", err)
	} else {
		fmt.Printf("✅ 带参数构建触发成功: QueueID=%d, URL=%s\n", result.QueueID, result.QueueURL)
	}

	fmt.Println("\n🎉 测试完成!")
}

func containsPipelineJobClass(class string) bool {
	return containsAny(class, []string{
		"WorkflowJob",
		"org.jenkinsci.plugins.workflow.job.WorkflowJob",
		"org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject",
	})
}

func containsAny(s string, substrings []string) bool {
	for _, sub := range substrings {
		if contains(s, sub) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findStringIndex(s, substr) != -1
}

func findStringIndex(s, substr string) int {
	n := len(substr)
	if n == 0 {
		return 0
	}
	for i := 0; i <= len(s)-n; i++ {
		if s[i:i+n] == substr {
			return i
		}
	}
	return -1
}