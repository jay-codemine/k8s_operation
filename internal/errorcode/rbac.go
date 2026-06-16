package errorcode

var (
	// 角色相关错误
	ErrorRoleCreateFail   *Error
	ErrorRoleUpdateFail   *Error
	ErrorRoleDeleteFail   *Error
	ErrorRoleNotFound     *Error
	ErrorRoleListFail     *Error
	ErrorRoleSystemLocked *Error

	// 权限相关错误
	ErrorPermissionNotFound   *Error
	ErrorPermissionListFail   *Error
	ErrorPermissionAssignFail *Error

	// 用户角色相关错误
	ErrorUserRoleAssignFail *Error
	ErrorUserRoleRemoveFail *Error
	ErrorUserRoleListFail   *Error

	// 集群权限相关错误
	ErrorClusterPermissionCreateFail *Error
	ErrorClusterPermissionUpdateFail *Error
	ErrorClusterPermissionDeleteFail *Error
	ErrorClusterPermissionListFail   *Error
	ErrorClusterPermissionNotFound   *Error

	// RBAC通用错误
	ErrorRBACAccessDenied      *Error
	ErrorRBACInvalidPermission *Error
)

func registerRBAC() {
	// 角色 300xxx
	ErrorRoleCreateFail = NewError(300001, "创建角色失败")
	ErrorRoleUpdateFail = NewError(300002, "更新角色失败")
	ErrorRoleDeleteFail = NewError(300003, "删除角色失败")
	ErrorRoleNotFound = NewError(300004, "角色不存在")
	ErrorRoleListFail = NewError(300005, "获取角色列表失败")
	ErrorRoleSystemLocked = NewError(300006, "系统内置角色不可修改")

	// 权限 301xxx
	ErrorPermissionNotFound = NewError(301001, "权限不存在")
	ErrorPermissionListFail = NewError(301002, "获取权限列表失败")
	ErrorPermissionAssignFail = NewError(301003, "分配权限失败")

	// 用户角色 302xxx
	ErrorUserRoleAssignFail = NewError(302001, "分配用户角色失败")
	ErrorUserRoleRemoveFail = NewError(302002, "移除用户角色失败")
	ErrorUserRoleListFail = NewError(302003, "获取用户角色列表失败")

	// 集群权限 303xxx
	ErrorClusterPermissionCreateFail = NewError(303001, "创建集群权限失败")
	ErrorClusterPermissionUpdateFail = NewError(303002, "更新集群权限失败")
	ErrorClusterPermissionDeleteFail = NewError(303003, "删除集群权限失败")
	ErrorClusterPermissionListFail = NewError(303004, "获取集群权限列表失败")
	ErrorClusterPermissionNotFound = NewError(303005, "集群权限不存在")

	// RBAC通用 309xxx
	ErrorRBACAccessDenied = NewError(309001, "访问被拒绝，权限不足")
	ErrorRBACInvalidPermission = NewError(309002, "无效的权限配置")
}
