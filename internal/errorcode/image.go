package errorcode

// 镜像管理相关错误码 (50001-50099)
var (
	ErrorGetImageRegistryListFail  = NewError(50001, "获取镜像仓库列表失败")
	ErrorImageRegistryNotFound     = NewError(50002, "镜像仓库不存在")
	ErrorCreateImageRegistryFail   = NewError(50003, "创建镜像仓库失败")
	ErrorUpdateImageRegistryFail   = NewError(50004, "更新镜像仓库失败")
	ErrorDeleteImageRegistryFail   = NewError(50005, "删除镜像仓库失败")
	ErrorCheckImageRegistryFail    = NewError(50006, "检测仓库连接失败")
	ErrorGetImageRegistryStatsFail = NewError(50007, "获取仓库统计失败")
	ErrorSetDefaultRegistryFail    = NewError(50008, "设置默认仓库失败")
	ErrorImageRegistryNameExists   = NewError(50009, "仓库名称已存在")
	ErrorListImagesFail            = NewError(50010, "获取镜像列表失败")
	ErrorGetImageTagsFail          = NewError(50011, "获取镜像标签失败")
)
