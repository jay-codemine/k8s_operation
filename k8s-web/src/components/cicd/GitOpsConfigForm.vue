<template>
  <div class="gitops-config-form">
    <a-divider orientation="left">
      <icon-gitlab /> GitOps 配置 (ArgoCD + Argo Workflows)
    </a-divider>

    <a-form-item field="gitops_config.argo_app_name" label="ArgoCD 应用名称" required>
      <a-input
        v-model="form.argo_app_name"
        placeholder="例如: my-app-prod"
        allow-clear
      />
      <template #extra>
        在 ArgoCD 中创建的 Application 名称，用于管理同步状态
      </template>
    </a-form-item>

    <a-row :gutter="16">
      <a-col :span="12">
        <a-form-item field="gitops_config.git_manifest_repo" label="Manifest Git 仓库">
          <a-input
            v-model="form.git_manifest_repo"
            placeholder="例如: https://git.example.com/manifests.git"
            allow-clear
          />
        </a-form-item>
      </a-col>
      <a-col :span="12">
        <a-form-item field="gitops_config.manifest_path" label="Manifest 路径">
          <a-input
            v-model="form.manifest_path"
            placeholder="例如: manifests/overlays/prod"
            allow-clear
          />
        </a-form-item>
      </a-col>
    </a-row>

    <a-row :gutter="16">
      <a-col :span="8">
        <a-form-item field="gitops_config.argo_project" label="ArgoCD Project">
          <a-input
            v-model="form.argo_project"
            placeholder="默认: default"
            allow-clear
          />
        </a-form-item>
      </a-col>
      <a-col :span="8">
        <a-form-item field="gitops_config.target_revision" label="目标分支/Tag">
          <a-input
            v-model="form.target_revision"
            placeholder="默认: main"
            allow-clear
          />
        </a-form-item>
      </a-col>
      <a-col :span="8">
        <a-form-item field="gitops_config.workflow_namespace" label="Workflow 命名空间">
          <a-input
            v-model="form.workflow_namespace"
            placeholder="默认: argo"
            allow-clear
          />
        </a-form-item>
      </a-col>
    </a-row>

    <a-divider orientation="left">镜像构建配置</a-divider>

    <a-row :gutter="16">
      <a-col :span="12">
        <a-form-item field="gitops_config.image_registry" label="镜像仓库地址" required>
          <a-input
            v-model="form.image_registry"
            placeholder="例如: registry.example.com/myproject"
            allow-clear
          />
        </a-form-item>
      </a-col>
      <a-col :span="12">
        <a-form-item field="gitops_config.image_repo" label="镜像名称" required>
          <a-input
            v-model="form.image_repo"
            placeholder="例如: myapp"
            allow-clear
          />
        </a-form-item>
      </a-col>
    </a-row>

    <a-row :gutter="16">
      <a-col :span="12">
        <a-form-item field="gitops_config.dockerfile_path" label="Dockerfile 路径">
          <a-input
            v-model="form.dockerfile_path"
            placeholder="默认: Dockerfile"
            allow-clear
          />
        </a-form-item>
      </a-col>
      <a-col :span="12">
        <a-form-item field="gitops_config.build_context" label="构建上下文">
          <a-input
            v-model="form.build_context"
            placeholder="默认: . (项目根目录)"
            allow-clear
          />
        </a-form-item>
      </a-col>
    </a-row>

    <a-divider orientation="left">同步策略</a-divider>

    <a-row :gutter="16">
      <a-col :span="8">
        <a-form-item field="gitops_config.auto_sync" label="自动同步">
          <a-switch v-model="form.auto_sync" />
          <template #extra>
            启用后 ArgoCD 自动同步 Git 仓库中的变更
          </template>
        </a-form-item>
      </a-col>
      <a-col :span="8">
        <a-form-item field="gitops_config.prune_resource" label="资源清理">
          <a-switch v-model="form.prune_resource" />
          <template #extra>
            自动删除 Git 中不存在但 K8s 中存在的资源
          </template>
        </a-form-item>
      </a-col>
    </a-row>
  </div>
</template>

<script setup>
import { reactive, watch, onMounted } from 'vue'

const props = defineProps({
  modelValue: {
    type: Object,
    default: () => ({})
  }
})

const emit = defineEmits(['update:modelValue'])

const form = reactive({
  argo_app_name: '',
  git_manifest_repo: '',
  manifest_path: 'manifests',
  argo_project: 'default',
  target_revision: 'main',
  workflow_namespace: 'argo',
  image_registry: '',
  image_repo: '',
  dockerfile_path: 'Dockerfile',
  build_context: '.',
  auto_sync: true,
  prune_resource: false
})

// Initialize from props
onMounted(() => {
  if (props.modelValue && Object.keys(props.modelValue).length > 0) {
    Object.assign(form, props.modelValue)
  }
})

// Watch form changes and emit
watch(form, (val) => {
  emit('update:modelValue', { ...val })
}, { deep: true })

// Watch external model changes
watch(() => props.modelValue, (val) => {
  if (val && Object.keys(val).length > 0) {
    Object.assign(form, val)
  }
}, { deep: true })
</script>

<style scoped>
.gitops-config-form {
  padding: 8px 0;
}
</style>
