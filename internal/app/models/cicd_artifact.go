package models

import dm "k8soperation/internal/domain/cicd"

// ===== 类型别名 =====
type CicdArtifact = dm.CicdArtifact

// ===== 常量别名 =====
const (
	ArtifactTypeJar     = dm.ArtifactTypeJar
	ArtifactTypeWar     = dm.ArtifactTypeWar
	ArtifactTypeBinary  = dm.ArtifactTypeBinary
	ArtifactTypeDist    = dm.ArtifactTypeDist
	ArtifactTypeWheel   = dm.ArtifactTypeWheel
	ArtifactTypeImage   = dm.ArtifactTypeImage
	ArtifactTypeArchive = dm.ArtifactTypeArchive

	ArtifactStatusUploading = dm.ArtifactStatusUploading
	ArtifactStatusReady     = dm.ArtifactStatusReady
	ArtifactStatusExpired   = dm.ArtifactStatusExpired
	ArtifactStatusDeleted   = dm.ArtifactStatusDeleted
)

// ===== 变量别名 =====
var ArtifactTypeByLanguage = dm.ArtifactTypeByLanguage
