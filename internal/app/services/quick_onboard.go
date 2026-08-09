package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8soperation/global"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/requests"
)

// ============================================================
// 快速接入：一键创建 K8s 资源 + 接入流水线 + 可选首次部署
// 支持 5 种工作负载 + ConfigMap/Secret/PVC
// ============================================================

type QuickOnboardResult struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	WorkloadKind string `json:"workload_kind"`
	WorkloadName string `json:"workload_name"`
	Namespace    string `json:"namespace"`
	ServiceName  string `json:"service_name,omitempty"`
	PipelineID   int64  `json:"pipeline_id,omitempty"`
	ReleaseID    int64  `json:"release_id,omitempty"`
}

func (s *Services) QuickOnboard(ctx context.Context, req *requests.QuickOnboardRequest, userID int64) (*QuickOnboardResult, error) {
	result := &QuickOnboardResult{WorkloadKind: req.WorkloadKind, Namespace: req.Namespace}

	if req.WorkloadName == "" {
		req.WorkloadName = req.AppName
	}
	if req.ContainerName == "" {
		req.ContainerName = req.WorkloadName
	}
	if req.Replicas <= 0 {
		req.Replicas = 1
	}
	result.WorkloadName = req.WorkloadName

	factory := NewClusterClientFactory(s)
	cli, err := factory.Get(ctx, req.ClusterID)
	if err != nil {
		result.Message = fmt.Sprintf("连接集群失败: %v", err)
		return result, err
	}

	if err := s.ensureNamespace(ctx, cli, req.Namespace); err != nil {
		global.Logger.Warn("[快速接入] 创建NS失败（可能已存在）", zap.String("namespace", req.Namespace), zap.Error(err))
	}

	labels := map[string]string{
		"app.kubernetes.io/name":       req.AppName,
		"app.kubernetes.io/managed-by": "k8s-operation-quick-onboard",
	}
	annotations := map[string]string{
		"k8s-operation/onboarded-at": time.Now().Format(time.RFC3339),
		"k8s-operation/app-name":     req.AppName,
	}

	// 先创建 ConfigMap / Secret / PVC，收集 volumes/volumeMounts
	volumes, volumeMounts := s.createExtraResources(ctx, cli, req)

	// 创建工作负载（已存在时视为接入成功，支持从 K8s 导入已有应用）
	workloadExisted := false
	if err := s.createWorkloadByKind(ctx, cli, req, labels, annotations, volumes, volumeMounts); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			result.Message = fmt.Sprintf("创建 %s 失败: %v", req.WorkloadKind, err)
			return result, err
		}
		workloadExisted = true
		global.Logger.Info("[快速接入] 工作负载已存在，跳过创建直接接入",
			zap.String("kind", req.WorkloadKind), zap.String("name", req.WorkloadName))
	}

	// 按需创建 Service（已存在时跳过）
	if req.ServiceType != "" {
		svcName, svcErr := s.createServiceForOnboard(ctx, cli, req, labels, annotations)
		if svcErr != nil && !k8serrors.IsAlreadyExists(svcErr) {
			global.Logger.Warn("[快速接入] 创建Service失败", zap.Error(svcErr))
		} else {
			result.ServiceName = svcName
		}
	}

	// 可选：创建 Pipeline
	if req.GitRepo != "" {
		pipelineID, pipeErr := s.createPipelineForOnboard(ctx, req, userID)
		if pipeErr != nil {
			global.Logger.Warn("[快速接入] 创建流水线失败", zap.Error(pipeErr))
		} else {
			result.PipelineID = pipelineID
		}
	}

	// 可选：立即部署
	if req.AutoDeploy {
		releaseID, depErr := s.quickDeploy(ctx, req, userID)
		if depErr != nil {
			global.Logger.Warn("[快速接入] 触发部署失败", zap.Error(depErr))
			result.Message = fmt.Sprintf("%s 创建成功（部署失败: %v）", req.WorkloadKind, depErr)
		} else {
			result.ReleaseID = releaseID
			result.Message = fmt.Sprintf("%s 创建成功并已触发部署", req.WorkloadKind)
		}
	} else if workloadExisted {
		result.Message = fmt.Sprintf("%s 已存在，接入成功", req.WorkloadKind)
	} else {
		result.Message = fmt.Sprintf("%s 创建成功", req.WorkloadKind)
	}

	result.Success = true
	return result, nil
}

// ==================== ConfigMap / Secret / PVC ====================

func (s *Services) createExtraResources(ctx context.Context, cli *K8sClients, req *requests.QuickOnboardRequest) ([]corev1.Volume, []corev1.VolumeMount) {
	var volumes []corev1.Volume
	var mounts []corev1.VolumeMount

	// ConfigMap
	if len(req.ConfigMapData) > 0 {
		cmName := req.WorkloadName + "-config"
		cmData := make(map[string]string)
		for _, kv := range req.ConfigMapData {
			cmData[kv.Key] = kv.Value
		}
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: cmName, Namespace: req.Namespace},
			Data:       cmData,
		}
		if _, err := cli.Kube.CoreV1().ConfigMaps(req.Namespace).Create(ctx, cm, metav1.CreateOptions{}); err == nil || k8serrors.IsAlreadyExists(err) {
			mountPath := req.ConfigMapMountPath
			if mountPath == "" {
				mountPath = "/etc/config"
			}
			volumes = append(volumes, corev1.Volume{
				Name: "config",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: cmName},
					},
				},
			})
			mounts = append(mounts, corev1.VolumeMount{Name: "config", MountPath: mountPath})
		}
	}

	// Secret
	if len(req.SecretData) > 0 {
		secName := req.WorkloadName + "-secret"
		secData := make(map[string][]byte)
		for _, kv := range req.SecretData {
			secData[kv.Key] = []byte(kv.Value)
		}
		sec := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: secName, Namespace: req.Namespace},
			Data:       secData,
		}
		if _, err := cli.Kube.CoreV1().Secrets(req.Namespace).Create(ctx, sec, metav1.CreateOptions{}); err == nil || k8serrors.IsAlreadyExists(err) {
			mountPath := req.SecretMountPath
			if mountPath == "" {
				mountPath = "/etc/secrets"
			}
			volumes = append(volumes, corev1.Volume{
				Name: "secret",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: secName,
					},
				},
			})
			mounts = append(mounts, corev1.VolumeMount{Name: "secret", MountPath: mountPath, ReadOnly: true})
		}
	}

	// PVC
	if req.PVCSize != "" {
		pvcName := req.PVCName
		if pvcName == "" {
			pvcName = req.WorkloadName + "-data"
		}
		accessMode := corev1.ReadWriteOnce
		switch req.PVCAccessMode {
		case "ReadWriteMany":
			accessMode = corev1.ReadWriteMany
		case "ReadOnlyMany":
			accessMode = corev1.ReadOnlyMany
		}
		storageClassName := req.PVCStorageClass
		var scName *string
		if storageClassName != "" {
			scName = &storageClassName
		}
		pvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{Name: pvcName, Namespace: req.Namespace},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{accessMode},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse(req.PVCSize),
					},
				},
				StorageClassName: scName,
			},
		}
		if _, err := cli.Kube.CoreV1().PersistentVolumeClaims(req.Namespace).Create(ctx, pvc, metav1.CreateOptions{}); err == nil || k8serrors.IsAlreadyExists(err) {
			mountPath := req.PVCMountPath
			if mountPath == "" {
				mountPath = "/data"
			}
			volumes = append(volumes, corev1.Volume{
				Name: "data",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: pvcName},
				},
			})
			mounts = append(mounts, corev1.VolumeMount{Name: "data", MountPath: mountPath})
		}
	}

	return volumes, mounts
}

// ==================== 工作负载创建（5 种） ====================

func (s *Services) createWorkloadByKind(ctx context.Context, cli *K8sClients, req *requests.QuickOnboardRequest, labels, annotations map[string]string, volumes []corev1.Volume, volumeMounts []corev1.VolumeMount) error {
	podSpec := s.buildPodSpec(req, volumes, volumeMounts)
	podLabels := copyMap(labels)

	switch req.WorkloadKind {
	case "Deployment":
		replicas := req.Replicas
		if replicas <= 0 { replicas = 1 }
		deploy := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: req.WorkloadName, Namespace: req.Namespace, Labels: labels, Annotations: annotations},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
				Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app.kubernetes.io/name": req.AppName}},
				Template: corev1.PodTemplateSpec{ObjectMeta: metav1.ObjectMeta{Labels: podLabels}, Spec: podSpec},
			},
		}
		_, err := cli.Kube.AppsV1().Deployments(req.Namespace).Create(ctx, deploy, metav1.CreateOptions{})
		return err

	case "StatefulSet":
		replicas := req.Replicas
		if replicas <= 0 { replicas = 1 }
		sts := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{Name: req.WorkloadName, Namespace: req.Namespace, Labels: labels, Annotations: annotations},
			Spec: appsv1.StatefulSetSpec{
				Replicas: &replicas, ServiceName: req.WorkloadName,
				Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app.kubernetes.io/name": req.AppName}},
				Template: corev1.PodTemplateSpec{ObjectMeta: metav1.ObjectMeta{Labels: podLabels}, Spec: podSpec},
			},
		}
		_, err := cli.Kube.AppsV1().StatefulSets(req.Namespace).Create(ctx, sts, metav1.CreateOptions{})
		return err

	case "DaemonSet":
		ds := &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{Name: req.WorkloadName, Namespace: req.Namespace, Labels: labels, Annotations: annotations},
			Spec: appsv1.DaemonSetSpec{
				Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app.kubernetes.io/name": req.AppName}},
				Template: corev1.PodTemplateSpec{ObjectMeta: metav1.ObjectMeta{Labels: podLabels}, Spec: podSpec},
			},
		}
		_, err := cli.Kube.AppsV1().DaemonSets(req.Namespace).Create(ctx, ds, metav1.CreateOptions{})
		return err

	case "CronJob":
		schedule := req.CronSchedule
		if schedule == "" { schedule = "*/5 * * * *" }
		concurrency := batchv1.AllowConcurrent
		if req.CronConcurrencyPolicy == "Forbid" { concurrency = batchv1.ForbidConcurrent
		} else if req.CronConcurrencyPolicy == "Replace" { concurrency = batchv1.ReplaceConcurrent }
		podSpec.RestartPolicy = corev1.RestartPolicyOnFailure
		cj := &batchv1.CronJob{
			ObjectMeta: metav1.ObjectMeta{Name: req.WorkloadName, Namespace: req.Namespace, Labels: labels, Annotations: annotations},
			Spec: batchv1.CronJobSpec{
				Schedule: schedule, ConcurrencyPolicy: concurrency,
				JobTemplate: batchv1.JobTemplateSpec{
					Spec: batchv1.JobSpec{Template: corev1.PodTemplateSpec{ObjectMeta: metav1.ObjectMeta{Labels: podLabels}, Spec: podSpec}},
				},
			},
		}
		_, err := cli.Kube.BatchV1().CronJobs(req.Namespace).Create(ctx, cj, metav1.CreateOptions{})
		return err

	case "Job":
		podSpec.RestartPolicy = corev1.RestartPolicyNever
		job := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{Name: req.WorkloadName, Namespace: req.Namespace, Labels: labels, Annotations: annotations},
			Spec: batchv1.JobSpec{Template: corev1.PodTemplateSpec{ObjectMeta: metav1.ObjectMeta{Labels: podLabels}, Spec: podSpec}},
		}
		if req.JobCompletions != nil { job.Spec.Completions = req.JobCompletions }
		if req.JobParallelism != nil { job.Spec.Parallelism = req.JobParallelism }
		if req.JobBackoffLimit != nil { job.Spec.BackoffLimit = req.JobBackoffLimit }
		if req.JobTTLSecondsAfterFinished != nil { job.Spec.TTLSecondsAfterFinished = req.JobTTLSecondsAfterFinished }
		_, err := cli.Kube.BatchV1().Jobs(req.Namespace).Create(ctx, job, metav1.CreateOptions{})
		return err

	default:
		return fmt.Errorf("不支持的工作负载类型: %s", req.WorkloadKind)
	}
}

// ==================== Service ====================

func (s *Services) createServiceForOnboard(ctx context.Context, cli *K8sClients, req *requests.QuickOnboardRequest, labels, annotations map[string]string) (string, error) {
	svcName := req.WorkloadName
	ports := s.buildServicePorts(req)
	if len(ports) == 0 {
		ports = []corev1.ServicePort{{Name: "http", Port: 80, TargetPort: intstr.FromInt(80), Protocol: corev1.ProtocolTCP}}
	}
	svcType := corev1.ServiceTypeClusterIP
	if req.ServiceType == "NodePort" { svcType = corev1.ServiceTypeNodePort
	} else if req.ServiceType == "LoadBalancer" { svcType = corev1.ServiceTypeLoadBalancer }
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: svcName, Namespace: req.Namespace, Labels: labels, Annotations: annotations},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app.kubernetes.io/name": req.AppName},
			Ports:    ports, Type: svcType,
		},
	}
	_, err := cli.Kube.CoreV1().Services(req.Namespace).Create(ctx, svc, metav1.CreateOptions{})
	return svcName, err
}

// ==================== Pipeline + Deploy ====================

func (s *Services) createPipelineForOnboard(ctx context.Context, req *requests.QuickOnboardRequest, userID int64) (int64, error) {
	existing, _ := s.cicdSvc().PipelineGetByName(ctx, req.AppName)
	if existing != nil {
		return existing.ID, nil
	}
	branch := req.GitBranch
	if branch == "" { branch = global.DefaultBranch() }
	pipeline := &models.CicdPipeline{
		Name: req.AppName, Description: fmt.Sprintf("快速接入: %s (%s)", req.AppName, req.WorkloadKind),
		GitRepo: req.GitRepo, GitBranch: branch, Status: models.PipelineStatusIdle,
		AutoDeploy: req.AutoDeploy, TargetClusterID: int64(req.ClusterID),
		TargetNamespace: req.Namespace, TargetWorkloadKind: req.WorkloadKind,
		TargetWorkloadName: req.WorkloadName, TargetContainer: req.ContainerName,
		CreatedUserID: userID,
	}
	if global.JenkinsSetting != nil && global.JenkinsSetting.URL != "" {
		pipeline.JenkinsURL = global.JenkinsSetting.URL
	}
	if err := s.cicdSvc().PipelineCreate(ctx, pipeline); err != nil { return 0, err }
	return pipeline.ID, nil
}

func (s *Services) quickDeploy(ctx context.Context, req *requests.QuickOnboardRequest, userID int64) (int64, error) {
	imageRepo, imageTag := parseImage(req.Image)
	return s.CicdReleaseCreate(ctx, &requests.CicdReleaseCreateRequest{
		AppName: req.AppName, Namespace: req.Namespace, WorkloadKind: req.WorkloadKind,
		WorkloadName: req.WorkloadName, ContainerName: req.ContainerName,
		Strategy: "rolling", TimeoutSec: 300, Concurrency: 1,
		ImageRepo: imageRepo, ImageTag: imageTag,
		ClusterIDs: []int64{int64(req.ClusterID)},
		Message:    "快速接入: 首次测试部署",
	}, userID)
}

// ==================== 工具函数 ====================

func (s *Services) buildPodSpec(req *requests.QuickOnboardRequest, volumes []corev1.Volume, mounts []corev1.VolumeMount) corev1.PodSpec {
	container := corev1.Container{Name: req.ContainerName, Image: req.Image, VolumeMounts: mounts}
	for _, p := range req.Ports {
		proto := corev1.ProtocolTCP
		if strings.EqualFold(p.Protocol, "UDP") { proto = corev1.ProtocolUDP }
		container.Ports = append(container.Ports, corev1.ContainerPort{Name: p.Name, ContainerPort: p.Port, Protocol: proto})
	}
	for _, ev := range req.EnvVars {
		container.Env = append(container.Env, corev1.EnvVar{Name: ev.Name, Value: ev.Value})
	}
	if req.CPUReq != "" || req.MemReq != "" || req.CPULim != "" || req.MemLim != "" {
		container.Resources = corev1.ResourceRequirements{Requests: corev1.ResourceList{}, Limits: corev1.ResourceList{}}
		if req.CPUReq != "" { container.Resources.Requests[corev1.ResourceCPU] = resource.MustParse(req.CPUReq) }
		if req.MemReq != "" { container.Resources.Requests[corev1.ResourceMemory] = resource.MustParse(req.MemReq) }
		if req.CPULim != "" { container.Resources.Limits[corev1.ResourceCPU] = resource.MustParse(req.CPULim) }
		if req.MemLim != "" { container.Resources.Limits[corev1.ResourceMemory] = resource.MustParse(req.MemLim) }
	}
	return corev1.PodSpec{Containers: []corev1.Container{container}, Volumes: volumes}
}

func (s *Services) buildServicePorts(req *requests.QuickOnboardRequest) []corev1.ServicePort {
	sourcePorts := req.Ports
	if len(req.ServicePorts) > 0 { sourcePorts = req.ServicePorts }
	ports := make([]corev1.ServicePort, 0, len(sourcePorts))
	for _, p := range sourcePorts {
		proto := corev1.ProtocolTCP
		if strings.EqualFold(p.Protocol, "UDP") { proto = corev1.ProtocolUDP }
		name := p.Name
		if name == "" { name = fmt.Sprintf("port-%d", p.Port) }
		targetPort := p.TargetPort
		if targetPort == 0 { targetPort = p.Port }
		sp := corev1.ServicePort{Name: name, Port: p.Port, TargetPort: intstr.FromInt(int(targetPort)), Protocol: proto}
		if p.NodePort > 0 { sp.NodePort = p.NodePort }
		ports = append(ports, sp)
	}
	return ports
}

func copyMap(src map[string]string) map[string]string {
	dst := make(map[string]string, len(src))
	for k, v := range src { dst[k] = v }
	return dst
}
