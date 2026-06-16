package builder

import (
	"strings"

	"github.com/google/uuid"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/utils"
)

func BuildCicdRelease(
	req *requests.CicdReleaseCreateRequest,
	userID int64,
	now uint64,
	imageRepo string,
	imageTag string,
	imageDigest string,
) *models.CicdRelease {

	return &models.CicdRelease{
		AppName:       req.AppName,
		Namespace:     req.Namespace,
		WorkloadKind:  req.WorkloadKind,
		WorkloadName:  req.WorkloadName,
		ContainerName: req.ContainerName,

		Strategy:    req.Strategy,
		TimeoutSec:  req.TimeoutSec,
		Concurrency: req.Concurrency,

		Status:  models.CicdReleaseStatusPending,
		Message: "",

		CreatedUserID: userID,
		RequestID:     resolveRequestID(req.RequestID),
		ImageRepo:     imageRepo,
		ImageTag:      imageTag,
		ImageDigest:   utils.NullString(imageDigest),

		CreatedAt:  now,
		ModifiedAt: now,
	}
}

// resolveRequestID 处理 request_id：空值时自动生成 UUID，避免唯一索引冲突
func resolveRequestID(reqID string) string {
	if id := strings.TrimSpace(reqID); id != "" {
		return id
	}
	return uuid.NewString()
}
