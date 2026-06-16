package k8srbac

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1/k8srbac"
)

type K8sRBACRouter struct{}

func NewK8sRBACRouter() *K8sRBACRouter {
	return &K8sRBACRouter{}
}

func (r *K8sRBACRouter) Inject(router *gin.RouterGroup) {
	ctl := v1.NewK8sRBACController()

	g := router.Group("/rbac")
	{
		// ServiceAccount
		g.GET("/serviceaccounts", ctl.ListServiceAccounts)
		g.GET("/serviceaccount", ctl.GetServiceAccount)
		g.POST("/serviceaccount", ctl.CreateServiceAccount)
		g.DELETE("/serviceaccount", ctl.DeleteServiceAccount)

		// Role / ClusterRole
		g.GET("/roles", ctl.ListRoles)
		g.POST("/role", ctl.CreateRole)
		g.DELETE("/role", ctl.DeleteRole)

		// RoleBinding / ClusterRoleBinding
		g.GET("/rolebindings", ctl.ListRoleBindings)
		g.POST("/rolebinding", ctl.CreateRoleBinding)
		g.DELETE("/rolebinding", ctl.DeleteRoleBinding)

		// SubjectAccessReview - 权限检查
		g.POST("/check", ctl.CheckSubjectAccess)
		g.POST("/check/batch", ctl.BatchCheckSubjectAccess)
	}
}
