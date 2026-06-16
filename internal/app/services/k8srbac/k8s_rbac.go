package k8srbac

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"
	authv1 "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8soperation/global"
	"k8soperation/pkg/utils"
)

// K8sRBACService K8s RBAC 服务
type K8sRBACService struct{}

func NewK8sRBACService() *K8sRBACService {
	return &K8sRBACService{}
}

// ==================== ServiceAccount ====================

// ServiceAccountItem ServiceAccount 列表项
type ServiceAccountItem struct {
	Name           string            `json:"name"`
	Namespace      string            `json:"namespace"`
	Secrets        []SecretRef       `json:"secrets"`
	AutomountToken *bool             `json:"automount_token"`
	Labels         map[string]string `json:"labels"`
	CreatedAt      string            `json:"created_at"`
}

type SecretRef struct {
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
}

// ListServiceAccounts 获取 ServiceAccount 列表
func (s *K8sRBACService) ListServiceAccounts(ctx context.Context, cli *kubernetes.Clientset, namespace string) ([]ServiceAccountItem, error) {
	var list *corev1.ServiceAccountList
	var err error

	if namespace == "" {
		list, err = cli.CoreV1().ServiceAccounts("").List(ctx, metav1.ListOptions{})
	} else {
		list, err = cli.CoreV1().ServiceAccounts(namespace).List(ctx, metav1.ListOptions{})
	}
	if err != nil {
		global.Logger.Error("获取 ServiceAccount 列表失败", zap.Error(err))
		return nil, err
	}

	items := make([]ServiceAccountItem, 0, len(list.Items))
	for _, sa := range list.Items {
		secrets := make([]SecretRef, 0, len(sa.Secrets))
		for _, sec := range sa.Secrets {
			secrets = append(secrets, SecretRef{Name: sec.Name})
		}

		items = append(items, ServiceAccountItem{
			Name:           sa.Name,
			Namespace:      sa.Namespace,
			Secrets:        secrets,
			AutomountToken: sa.AutomountServiceAccountToken,
			Labels:         sa.Labels,
			CreatedAt:      sa.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
		})
	}

	return items, nil
}

// GetServiceAccount 获取 ServiceAccount 详情
func (s *K8sRBACService) GetServiceAccount(ctx context.Context, cli *kubernetes.Clientset, namespace, name string) (*ServiceAccountItem, error) {
	sa, err := cli.CoreV1().ServiceAccounts(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		global.Logger.Error("获取 ServiceAccount 详情失败", zap.Error(err))
		return nil, err
	}

	secrets := make([]SecretRef, 0, len(sa.Secrets))
	for _, sec := range sa.Secrets {
		secrets = append(secrets, SecretRef{Name: sec.Name})
	}

	return &ServiceAccountItem{
		Name:           sa.Name,
		Namespace:      sa.Namespace,
		Secrets:        secrets,
		AutomountToken: sa.AutomountServiceAccountToken,
		Labels:         sa.Labels,
		CreatedAt:      sa.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
	}, nil
}

// CreateServiceAccount 创建 ServiceAccount
func (s *K8sRBACService) CreateServiceAccount(ctx context.Context, cli *kubernetes.Clientset, namespace, name string, labels map[string]string, autoMount bool) (*corev1.ServiceAccount, error) {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		AutomountServiceAccountToken: &autoMount,
	}

	created, err := cli.CoreV1().ServiceAccounts(namespace).Create(ctx, sa, metav1.CreateOptions{})
	if err != nil {
		global.Logger.Error("创建 ServiceAccount 失败", zap.Error(err))
		return nil, err
	}

	return created, nil
}

// DeleteServiceAccount 删除 ServiceAccount
func (s *K8sRBACService) DeleteServiceAccount(ctx context.Context, cli *kubernetes.Clientset, namespace, name string) error {
	err := cli.CoreV1().ServiceAccounts(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		global.Logger.Error("删除 ServiceAccount 失败", zap.Error(err))
		return err
	}
	return nil
}

// ==================== Role / ClusterRole ====================

// RoleItem Role 列表项
type RoleItem struct {
	Name      string              `json:"name"`
	Type      string              `json:"type"` // "Role" or "ClusterRole"
	Namespace string              `json:"namespace,omitempty"`
	Rules     []rbacv1.PolicyRule `json:"rules"`
	CreatedAt string              `json:"created_at"`
}

// ListRoles 获取 Role 列表（并行优化 + panic 保护）
func (s *K8sRBACService) ListRoles(ctx context.Context, cli *kubernetes.Clientset, namespace string) ([]RoleItem, error) {
	var (
		clusterRoles *rbacv1.ClusterRoleList
		roles        *rbacv1.RoleList
		crErr        error
		rErr         error
		wg           sync.WaitGroup
	)

	// 并行获取 ClusterRoles 和 Roles（带 panic 保护）
	wg.Add(2)

	utils.SafeGoWithWaitGroup(&wg, func() {
		clusterRoles, crErr = cli.RbacV1().ClusterRoles().List(ctx, metav1.ListOptions{})
	}, func(r interface{}) {
		crErr = fmt.Errorf("panic: %v", r)
	})

	utils.SafeGoWithWaitGroup(&wg, func() {
		if namespace == "" {
			roles, rErr = cli.RbacV1().Roles("").List(ctx, metav1.ListOptions{})
		} else {
			roles, rErr = cli.RbacV1().Roles(namespace).List(ctx, metav1.ListOptions{})
		}
	}, func(r interface{}) {
		rErr = fmt.Errorf("panic: %v", r)
	})

	wg.Wait()

	// 检查错误
	if crErr != nil {
		global.Logger.Error("获取 ClusterRole 列表失败", zap.Error(crErr))
		return nil, crErr
	}
	if rErr != nil {
		global.Logger.Error("获取 Role 列表失败", zap.Error(rErr))
		return nil, rErr
	}

	// 组装结果
	items := make([]RoleItem, 0, len(clusterRoles.Items)+len(roles.Items))

	for _, cr := range clusterRoles.Items {
		items = append(items, RoleItem{
			Name:      cr.Name,
			Type:      "ClusterRole",
			Rules:     cr.Rules,
			CreatedAt: cr.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
		})
	}

	for _, r := range roles.Items {
		items = append(items, RoleItem{
			Name:      r.Name,
			Type:      "Role",
			Namespace: r.Namespace,
			Rules:     r.Rules,
			CreatedAt: r.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
		})
	}

	return items, nil
}

// CreateRole 创建 Role
func (s *K8sRBACService) CreateRole(ctx context.Context, cli *kubernetes.Clientset, roleType, namespace, name string, rules []rbacv1.PolicyRule) error {
	if roleType == "ClusterRole" {
		cr := &rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{Name: name},
			Rules:      rules,
		}
		_, err := cli.RbacV1().ClusterRoles().Create(ctx, cr, metav1.CreateOptions{})
		return err
	}

	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Rules:      rules,
	}
	_, err := cli.RbacV1().Roles(namespace).Create(ctx, role, metav1.CreateOptions{})
	return err
}

// DeleteRole 删除 Role
func (s *K8sRBACService) DeleteRole(ctx context.Context, cli *kubernetes.Clientset, roleType, namespace, name string) error {
	if roleType == "ClusterRole" {
		return cli.RbacV1().ClusterRoles().Delete(ctx, name, metav1.DeleteOptions{})
	}
	return cli.RbacV1().Roles(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// ==================== RoleBinding / ClusterRoleBinding ====================

// RoleBindingItem RoleBinding 列表项
type RoleBindingItem struct {
	Name      string           `json:"name"`
	Type      string           `json:"type"` // "RoleBinding" or "ClusterRoleBinding"
	Namespace string           `json:"namespace,omitempty"`
	RoleRef   RoleRef          `json:"role_ref"`
	Subjects  []rbacv1.Subject `json:"subjects"`
	CreatedAt string           `json:"created_at"`
}

type RoleRef struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

// ListRoleBindings 获取 RoleBinding 列表（并行优化 + panic 保护）
func (s *K8sRBACService) ListRoleBindings(ctx context.Context, cli *kubernetes.Clientset, namespace string) ([]RoleBindingItem, error) {
	var (
		crbs   *rbacv1.ClusterRoleBindingList
		rbs    *rbacv1.RoleBindingList
		crbErr error
		rbErr  error
		wg     sync.WaitGroup
	)

	// 并行获取 ClusterRoleBindings 和 RoleBindings（带 panic 保护）
	wg.Add(2)

	utils.SafeGoWithWaitGroup(&wg, func() {
		crbs, crbErr = cli.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
	}, func(r interface{}) {
		crbErr = fmt.Errorf("panic: %v", r)
	})

	utils.SafeGoWithWaitGroup(&wg, func() {
		if namespace == "" {
			rbs, rbErr = cli.RbacV1().RoleBindings("").List(ctx, metav1.ListOptions{})
		} else {
			rbs, rbErr = cli.RbacV1().RoleBindings(namespace).List(ctx, metav1.ListOptions{})
		}
	}, func(r interface{}) {
		rbErr = fmt.Errorf("panic: %v", r)
	})

	wg.Wait()

	// 检查错误
	if crbErr != nil {
		global.Logger.Error("获取 ClusterRoleBinding 列表失败", zap.Error(crbErr))
		return nil, crbErr
	}
	if rbErr != nil {
		global.Logger.Error("获取 RoleBinding 列表失败", zap.Error(rbErr))
		return nil, rbErr
	}

	// 组装结果
	items := make([]RoleBindingItem, 0, len(crbs.Items)+len(rbs.Items))

	for _, crb := range crbs.Items {
		items = append(items, RoleBindingItem{
			Name: crb.Name,
			Type: "ClusterRoleBinding",
			RoleRef: RoleRef{
				Kind: crb.RoleRef.Kind,
				Name: crb.RoleRef.Name,
			},
			Subjects:  crb.Subjects,
			CreatedAt: crb.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
		})
	}

	for _, rb := range rbs.Items {
		items = append(items, RoleBindingItem{
			Name:      rb.Name,
			Type:      "RoleBinding",
			Namespace: rb.Namespace,
			RoleRef: RoleRef{
				Kind: rb.RoleRef.Kind,
				Name: rb.RoleRef.Name,
			},
			Subjects:  rb.Subjects,
			CreatedAt: rb.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
		})
	}

	return items, nil
}

// CreateRoleBinding 创建 RoleBinding
func (s *K8sRBACService) CreateRoleBinding(ctx context.Context, cli *kubernetes.Clientset, bindingType, namespace, name string, roleRef RoleRef, subjects []rbacv1.Subject) error {
	if bindingType == "ClusterRoleBinding" {
		crb := &rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{Name: name},
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     roleRef.Kind,
				Name:     roleRef.Name,
			},
			Subjects: subjects,
		}
		_, err := cli.RbacV1().ClusterRoleBindings().Create(ctx, crb, metav1.CreateOptions{})
		return err
	}

	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     roleRef.Kind,
			Name:     roleRef.Name,
		},
		Subjects: subjects,
	}
	_, err := cli.RbacV1().RoleBindings(namespace).Create(ctx, rb, metav1.CreateOptions{})
	return err
}

// DeleteRoleBinding 删除 RoleBinding
func (s *K8sRBACService) DeleteRoleBinding(ctx context.Context, cli *kubernetes.Clientset, bindingType, namespace, name string) error {
	if bindingType == "ClusterRoleBinding" {
		return cli.RbacV1().ClusterRoleBindings().Delete(ctx, name, metav1.DeleteOptions{})
	}
	return cli.RbacV1().RoleBindings(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// ==================== SubjectAccessReview ====================

// SubjectAccessCheckRequest 权限检查请求
type SubjectAccessCheckRequest struct {
	SubjectType string `json:"subject_type"` // User, ServiceAccount
	Username    string `json:"username,omitempty"`
	SANamespace string `json:"sa_namespace,omitempty"`
	SAName      string `json:"sa_name,omitempty"`
	Namespace   string `json:"namespace"`
	APIGroup    string `json:"api_group"`
	Resource    string `json:"resource"`
	ResourceName string `json:"resource_name,omitempty"`
	Verb        string `json:"verb"`
}

// SubjectAccessCheckResult 权限检查结果
type SubjectAccessCheckResult struct {
	Allowed     bool     `json:"allowed"`
	Reason      string   `json:"reason,omitempty"`
	MatchedRoles []string `json:"matched_roles,omitempty"`
}

// CheckSubjectAccess 检查主体权限（使用 K8s SubjectAccessReview）
func (s *K8sRBACService) CheckSubjectAccess(ctx context.Context, cli *kubernetes.Clientset, req SubjectAccessCheckRequest) (*SubjectAccessCheckResult, error) {
	// 构建 SubjectAccessReview 请求
	sar := &authv1.SubjectAccessReview{
		Spec: authv1.SubjectAccessReviewSpec{
			ResourceAttributes: &authv1.ResourceAttributes{
				Namespace: req.Namespace,
				Verb:      req.Verb,
				Group:     req.APIGroup,
				Resource:  req.Resource,
				Name:      req.ResourceName,
			},
		},
	}

	// 设置主体
	if req.SubjectType == "ServiceAccount" {
		// ServiceAccount 格式：system:serviceaccount:<namespace>:<name>
		sar.Spec.User = "system:serviceaccount:" + req.SANamespace + ":" + req.SAName
	} else {
		sar.Spec.User = req.Username
	}

	// 执行权限检查
	review, err := cli.AuthorizationV1().SubjectAccessReviews().Create(ctx, sar, metav1.CreateOptions{})
	if err != nil {
		global.Logger.Error("权限检查失败", zap.Error(err))
		return nil, err
	}

	result := &SubjectAccessCheckResult{
		Allowed: review.Status.Allowed,
		Reason:  review.Status.Reason,
	}

	// 如果被拒绝，尝试获取原因
	if !review.Status.Allowed {
		if review.Status.Reason != "" {
			result.Reason = review.Status.Reason
		} else if review.Status.EvaluationError != "" {
			result.Reason = review.Status.EvaluationError
		} else {
			result.Reason = "没有匹配的 RoleBinding 授予此权限"
		}
	} else {
		result.Reason = "权限校验通过"
	}

	return result, nil
}

// BatchCheckSubjectAccess 批量检查权限
func (s *K8sRBACService) BatchCheckSubjectAccess(ctx context.Context, cli *kubernetes.Clientset, checks []SubjectAccessCheckRequest) ([]SubjectAccessCheckResult, error) {
	results := make([]SubjectAccessCheckResult, 0, len(checks))
	for _, check := range checks {
		result, err := s.CheckSubjectAccess(ctx, cli, check)
		if err != nil {
			results = append(results, SubjectAccessCheckResult{
				Allowed: false,
				Reason:  "检查失败: " + err.Error(),
			})
		} else {
			results = append(results, *result)
		}
	}
	return results, nil
}
