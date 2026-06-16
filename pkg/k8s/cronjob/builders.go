package cronjob

import (
	"github.com/gin-gonic/gin"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/job"
	"strings"
)

// 建议放在 cronjob 包顶部的默认值
var (
	defaultPolicy               = batchv1.ForbidConcurrent
	defaultSuccessHistory int32 = 3
	defaultFailedHistory  int32 = 1
)

func BuildCronJobFromCreateReq(req *requests.KubeCronJobCreateRequest) *batchv1.CronJob {
	// --- 基本参数 ---
	ns := strings.TrimSpace(req.Namespace)
	name := strings.TrimSpace(req.Name)
	schedule := strings.TrimSpace(req.Schedule)

	// --- RestartPolicy 默认值 ---
	rp := corev1.RestartPolicyOnFailure
	if req.RestartPolicy == string(corev1.RestartPolicyNever) {
		rp = corev1.RestartPolicyNever
	}

	// --- 构造容器 ---
	var containers []corev1.Container
	if len(req.Containers) > 0 {
		containers = req.Containers
	} else {
		cn := strings.TrimSpace(req.ContainerName)
		if cn == "" {
			cn = name
		}
		containers = []corev1.Container{{
			Name:    cn,
			Image:   strings.TrimSpace(req.ContainerImage),
			Command: buildCommand(req.ContainerCommand),
			Args:    buildArgs(req.ContainerCommandArgs),
		}}
	}

	// --- 系统标签（CronJob + Job + Pod 都会继承） ---
	systemLabels := map[string]string{
		"app.k8soperation.io/name":       name,
		"app.k8soperation.io/managed-by": "k8soperation",
		"app.k8soperation.io/cronjob":    name,
		"app.k8soperation.io/created-by": "form",
	}

	// --- 构造 JobSpec ---
	jobSpec := batchv1.JobSpec{
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: systemLabels, // Pod 标签
			},
			Spec: job.BuildJobPodSpec(&requests.KubeJobCreateRequest{
				RestartPolicy:    string(rp),
				ImagePullSecrets: req.ImagePullSecrets,
				ServiceAccount:   req.ServiceAccount,
				NodeSelector:     req.NodeSelector,
				Tolerations:      req.Tolerations,
				Affinity:         req.Affinity,
			}, containers),
		},
	}

	// --- 构造 CronJob ---
	cj := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
			Labels:    systemLabels, // CronJob 自身标签
		},
		Spec: batchv1.CronJobSpec{
			Schedule:      schedule,
			Suspend:       req.Suspend,
			JobTemplate: batchv1.JobTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: systemLabels, // Job 标签（继承到创建的 Job）
				},
				Spec: jobSpec,
			},
			ConcurrencyPolicy:          batchv1.ConcurrencyPolicy(req.ConcurrencyPolicy),
			StartingDeadlineSeconds:    req.StartingDeadlineSeconds,
			SuccessfulJobsHistoryLimit: req.SuccessfulJobsHistoryLimit,
			FailedJobsHistoryLimit:     req.FailedJobsHistoryLimit,
		},
	}

	// --- 时区 ---
	if tz := strings.TrimSpace(req.TimeZone); tz != "" {
		cj.Spec.TimeZone = &tz
	}

	return cj
}

func buildCommand(cmd []string) []string {
	if len(cmd) == 0 {
		return nil
	}
	return cmd
}

func buildArgs(args []string) []string {
	if len(args) == 0 {
		return nil
	}
	return args
}

// 小工具
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// BuildCronJobResponse 构建 CronJob 的返回结构
func BuildCronJobResponse(cj *batchv1.CronJob, req *requests.KubeCronJobCreateRequest) gin.H {
	// 取第一个容器镜像（兼容旧前端单镜像展示）
	firstImage := ""
	cs := cj.Spec.JobTemplate.Spec.Template.Spec.Containers
	if len(cs) > 0 {
		firstImage = cs[0].Image
	}

	// 同时把所有容器返回给前端（推荐）
	outContainers := make([]gin.H, 0, len(cs))
	for _, c := range cs {
		outContainers = append(outContainers, gin.H{
			"name":    c.Name,
			"image":   c.Image,
			"command": c.Command,
			"args":    c.Args,
		})
	}

	spec := gin.H{
		"schedule":                   cj.Spec.Schedule,
		"concurrencyPolicy":          string(cj.Spec.ConcurrencyPolicy),
		"suspend":                    cj.Spec.Suspend != nil && *cj.Spec.Suspend,
		"successfulJobsHistoryLimit": cj.Spec.SuccessfulJobsHistoryLimit,
		"failedJobsHistoryLimit":     cj.Spec.FailedJobsHistoryLimit,
	}

	// 可选：把 JobTemplate 里的一些字段也回传
	js := cj.Spec.JobTemplate.Spec
	spec["jobSpec"] = gin.H{
		"restartPolicy":           string(cj.Spec.JobTemplate.Spec.Template.Spec.RestartPolicy),
		"parallelism":             js.Parallelism,
		"completions":             js.Completions,
		"backoffLimit":            js.BackoffLimit,
		"ttlSecondsAfterFinished": js.TTLSecondsAfterFinished,
		"activeDeadlineSeconds":   js.ActiveDeadlineSeconds,
	}

	status := gin.H{
		"active": len(cj.Status.Active),
		"lastScheduleTime": func() interface{} {
			if cj.Status.LastScheduleTime != nil {
				return cj.Status.LastScheduleTime.Time
			}
			return nil
		}(),
		"lastSuccessfulTime": func() interface{} {
			if cj.Status.LastSuccessfulTime != nil {
				return cj.Status.LastSuccessfulTime.Time
			}
			return nil
		}(),
	}

	return gin.H{
		"cronjob": gin.H{
			"name":            cj.Name,
			"namespace":       cj.Namespace,
			"labels":          cj.Labels,
			"uid":             string(cj.UID),
			"resourceVersion": cj.ResourceVersion,
			"image":           firstImage,    // ★ 兼容老字段
			"containers":      outContainers, // ★ 新增容器数组
			"spec":            spec,
			"status":          status,
		},
	}
}

// CronJobListItem 列表项响应结构
type CronJobListItem struct {
	Name                       string   `json:"name"`
	Namespace                  string   `json:"namespace"`
	Status                     string   `json:"status"`                          // Active/Suspended
	Schedule                   string   `json:"schedule"`                         // cron 表达式
	Suspend                    bool     `json:"suspend"`                          // 是否暂停
	LastScheduleTime           string   `json:"last_schedule_time"`               // 最后调度时间
	Active                     int      `json:"active"`                           // 活跃 Job 数
	SuccessfulJobsHistoryLimit *int32   `json:"successful_jobs_history_limit"`
	FailedJobsHistoryLimit     *int32   `json:"failed_jobs_history_limit"`
	Image                      string   `json:"image"`      // 第一个容器镜像
	Containers                 []string `json:"containers"` // 所有容器名称
	CreatedAt                  string   `json:"created_at"` // 创建时间
	// Job 执行统计 (Rancher/Kuboard 风格)
	JobStats                   *JobStats `json:"job_stats,omitempty"` // Job 统计
}

// JobStats CronJob 关联的 Job 统计信息
type JobStats struct {
	Total     int `json:"total"`     // 总执行次数
	Succeeded int `json:"succeeded"` // 成功次数
	Failed    int `json:"failed"`    // 失败次数
	Running   int `json:"running"`   // 运行中
}

// BuildCronJobListResponse 将 CronJob 列表转换为响应格式
// jobs 参数为当前命名空间下的所有 Job，用于统计关联的 Job 执行情况
func BuildCronJobListResponse(cronjobs []batchv1.CronJob, jobs []batchv1.Job) []CronJobListItem {
	result := make([]CronJobListItem, 0, len(cronjobs))

	// 构建 CronJob UID 到 Job 统计的映射
	jobStatsMap := buildJobStatsMap(cronjobs, jobs)

	for _, cj := range cronjobs {
		// 提取容器信息
		containers := []string{}
		var firstImage string

		if len(cj.Spec.JobTemplate.Spec.Template.Spec.Containers) > 0 {
			for _, c := range cj.Spec.JobTemplate.Spec.Template.Spec.Containers {
				containers = append(containers, c.Name)
			}
			firstImage = cj.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Image
		}

		// 获取状态
		suspend := cj.Spec.Suspend != nil && *cj.Spec.Suspend
		status := "Active"
		if suspend {
			status = "Suspended"
		}

		// 格式化时间
		lastScheduleTime := ""
		if cj.Status.LastScheduleTime != nil {
			lastScheduleTime = cj.Status.LastScheduleTime.Format("2006-01-02 15:04:05")
		}

		item := CronJobListItem{
			Name:                       cj.Name,
			Namespace:                  cj.Namespace,
			Status:                     status,
			Schedule:                   cj.Spec.Schedule,
			Suspend:                    suspend,
			LastScheduleTime:           lastScheduleTime,
			Active:                     len(cj.Status.Active),
			SuccessfulJobsHistoryLimit: cj.Spec.SuccessfulJobsHistoryLimit,
			FailedJobsHistoryLimit:     cj.Spec.FailedJobsHistoryLimit,
			Image:                      firstImage,
			Containers:                 containers,
			CreatedAt:                  cj.CreationTimestamp.Format("2006-01-02 15:04:05"),
			JobStats:                   jobStatsMap[string(cj.UID)],
		}

		result = append(result, item)
	}

	return result
}

// buildJobStatsMap 构建 CronJob UID 到 Job 统计的映射
func buildJobStatsMap(cronjobs []batchv1.CronJob, jobs []batchv1.Job) map[string]*JobStats {
	statsMap := make(map[string]*JobStats)

	// 初始化所有 CronJob 的统计
	for _, cj := range cronjobs {
		statsMap[string(cj.UID)] = &JobStats{}
	}

	// 遍历所有 Job，通过 ownerReferences 匹配 CronJob
	for _, job := range jobs {
		for _, owner := range job.OwnerReferences {
			if owner.Kind == "CronJob" {
				if stats, ok := statsMap[string(owner.UID)]; ok {
					stats.Total++
					// 判断 Job 状态
					if job.Status.Succeeded > 0 {
						stats.Succeeded++
					} else if job.Status.Failed > 0 {
						stats.Failed++
					} else if job.Status.Active > 0 {
						stats.Running++
					}
				}
				break
			}
		}
	}

	return statsMap
}
