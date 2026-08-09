package models

import dm "k8soperation/internal/domain/appstore"

// ===== 类型别名 =====

type (
	AppStoreApp                         = dm.AppStoreApp
	AppStoreComponent                   = dm.AppStoreComponent
	AppStoreComponentRequest            = dm.AppStoreComponentRequest
	AppStoreComponentBatchDeleteRequest = dm.AppStoreComponentBatchDeleteRequest
	AppStoreComponentSortRequest        = dm.AppStoreComponentSortRequest
	ComponentSortItem                   = dm.ComponentSortItem
	AppStoreListRequest                 = dm.AppStoreListRequest
	AppStoreCreateRequest               = dm.AppStoreCreateRequest
	AppStoreUpdateRequest               = dm.AppStoreUpdateRequest
	AppStoreInstallRequest              = dm.AppStoreInstallRequest
	AppStoreInstall                     = dm.AppStoreInstall
	AppStoreInstallListRequest          = dm.AppStoreInstallListRequest
	AppStoreCategoryCount               = dm.AppStoreCategoryCount
	AppInstallStatusResponse            = dm.AppInstallStatusResponse
	NamespaceOverview                   = dm.NamespaceOverview
	DeploymentStatusInfo                = dm.DeploymentStatusInfo
	ServiceStatusInfo                   = dm.ServiceStatusInfo
	K8sEventInfo                        = dm.K8sEventInfo
	PodStatusInfo                       = dm.PodStatusInfo
	ContainerStatusInfo                 = dm.ContainerStatusInfo
	ConfigMapStatusInfo                 = dm.ConfigMapStatusInfo
	AppStoreInstallUpdateRequest        = dm.AppStoreInstallUpdateRequest
)

// ===== 安装状态常量 =====

const (
	InstallStatusInstalling   = dm.InstallStatusInstalling
	InstallStatusInstalled    = dm.InstallStatusInstalled
	InstallStatusFailed       = dm.InstallStatusFailed
	InstallStatusUninstalling = dm.InstallStatusUninstalling
	InstallStatusUninstalled  = dm.InstallStatusUninstalled
	InstallStatusPartialReady = dm.InstallStatusPartialReady
)
