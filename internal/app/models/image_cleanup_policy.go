package models

import (
	dm "k8soperation/internal/domain/image"
)

// ===== 类型别名（领域定义在 domain/image） =====

type (
	ImageCleanupPolicy = dm.ImageCleanupPolicy
	ImageCleanupLog    = dm.ImageCleanupLog
)
