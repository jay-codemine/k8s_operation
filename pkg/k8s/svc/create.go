package svc

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
)

/**
 * 创建Kubernetes Service的函数
 * @param ctx 上下文信息，用于控制请求的超时和取消
 * @param req 创建Service的请求参数，包含Service的配置信息
 * @return 创建成功的Service对象和可能的错误
 */
func CreateService(ctx context.Context, Kube kubernetes.Interface, req *requests.KubeServiceCreateRequest) (*corev1.Service, error) {
	// 根据请求参数构建Service对象
	svc := BuildServiceFromSvcReq(req)

	// 调用Kubernetes API创建Service
	createdSvc, err := Kube.CoreV1().
		Services(req.Namespace).
		Create(ctx, svc, metav1.CreateOptions{})
	// 处理可能的错误
	if err != nil {
		// 如果错误类型为Service已存在，返回自定义错误信息
		if apierrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("service %q already exists in namespace %q", svc.Name, svc.Namespace)
		}
		// 记录创建失败的错误日志
		global.Logger.Errorf("create service failed: %v", err)
		return nil, err
	}

	// 记录Service创建成功的日志
	global.Logger.Infof("service %q created successfully", createdSvc.Name)
	return createdSvc, nil
}
