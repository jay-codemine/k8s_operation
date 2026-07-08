<template>
  <div class="gitops-create-view">
    <!-- 顶部面包屑 -->
    <div class="breadcrumb">
      <router-link to="/cicd/pipelines" class="breadcrumb-link">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16">
          <polyline points="15 18 9 12 15 6"/>
        </svg>
        流水线列表
      </router-link>
      <span class="separator">/</span>
      <span class="current">{{ isEdit ? '编辑 GitOps 流水线' : '创建 GitOps 流水线' }}</span>
    </div>

    <!-- 页面标题 -->
    <div class="page-header">
      <div class="header-content">
        <div class="header-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="16 3 21 3 21 8"/>
            <line x1="4" y1="20" x2="21" y2="3"/>
            <polyline points="21 16 21 21 16 21"/>
            <line x1="15" y1="15" x2="21" y2="21"/>
            <line x1="4" y1="4" x2="9" y2="9"/>
          </svg>
        </div>
        <div class="header-text">
          <h1>{{ isEdit ? '编辑 GitOps 流水线' : '创建 GitOps 流水线' }}</h1>
          <p>使用 ArgoCD + Argo Workflows 实现 Pull-based 部署</p>
        </div>
      </div>
      <div class="header-actions">
        <button
          type="button"
          class="btn-header-save"
          @click="submit"
          :disabled="submitting"
        >
          <span v-if="submitting" class="loading-spinner-sm"></span>
          <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16">
            <polyline points="20 6 9 17 4 12"/>
          </svg>
          {{ submitting ? '保存中...' : (isEdit ? '保存修改' : '创建流水线') }}
        </button>
        <button class="btn-icon" @click="cancel" title="返回列表">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18">
            <path d="M6 18L18 6M6 6l12 12"/>
          </svg>
        </button>
      </div>
    </div>

    <!-- 表单 -->
    <div class="form-container">
      <form @submit.prevent="submit">
        <!-- 基本信息卡片 -->
        <div class="form-card">
          <div class="card-title">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16">
              <path d="M12 2L2 7l10 5 10-5-10-5z"/>
              <path d="M2 17l10 5 10-5"/>
              <path d="M2 12l10 5 10-5"/>
            </svg>
            基本信息
          </div>

          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item field="name" label="应用名称" required>
                <a-input v-model="form.name" placeholder="例如: my-app" @blur="checkName" allow-clear />
                <template #extra>
                  <span v-if="nameChecking" style="color:#94a3b8">检查中...</span>
                  <span v-else-if="nameAvailable === false" style="color:#ef4444">名称已存在</span>
                  <span v-else-if="nameAvailable === true" style="color:#22c55e">名称可用</span>
                  <span v-else>小写字母和连字符</span>
                </template>
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item field="language_type" label="语言类型" required>
                <a-select v-model="form.language_type" placeholder="选择语言类型">
                  <a-option value="java">Java</a-option>
                  <a-option value="go">Go</a-option>
                  <a-option value="frontend">Node.js (前端)</a-option>
                  <a-option value="python">Python</a-option>
                  <a-option value="custom">自定义</a-option>
                </a-select>
                <template #extra>选择语言后自动匹配 Argo Workflow 构建模板</template>
              </a-form-item>
            </a-col>
          </a-row>

          <a-form-item field="description" label="描述">
            <a-textarea v-model="form.description" placeholder="简要描述此应用..." :max-length="200" :auto-size="{ minRows: 2 }" allow-clear />
          </a-form-item>

          <a-row :gutter="16">
            <a-col :span="16">
              <a-form-item field="git_repo" label="Git 仓库地址" required>
                <a-input v-model="form.git_repo" placeholder="https://gitee.com/org/repo.git" allow-clear />
              </a-form-item>
            </a-col>
            <a-col :span="8">
              <a-form-item field="git_branch" label="分支" required>
                <a-input v-model="form.git_branch" placeholder="main" allow-clear />
              </a-form-item>
            </a-col>
          </a-row>
        </div>

        <!-- GitOps 配置卡片 -->
        <div class="form-card">
          <GitOpsConfigForm v-model="form.gitops_config" />
        </div>

        <!-- 自动部署配置（可选） -->
        <div class="form-card">
          <div class="card-title">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16">
              <circle cx="12" cy="12" r="10"/>
              <path d="M12 6v6l4 2"/>
            </svg>
            自动部署配置
            <span style="font-size:12px;color:#94a3b8;font-weight:400;margin-left:8px;">(可选)</span>
          </div>

          <a-form-item field="auto_deploy" label="启用自动部署">
            <a-switch v-model="form.auto_deploy" />
            <template #extra>ArgoCD 同步成功后自动更新 K8s 工作负载镜像</template>
          </a-form-item>

          <template v-if="form.auto_deploy">
            <a-row :gutter="16">
              <a-col :span="8">
                <a-form-item field="target_cluster_id" label="目标集群" required>
                  <a-select v-model="form.target_cluster_id" placeholder="选择集群" @change="onClusterChange" :loading="loadingClusters" allow-clear>
                    <a-option v-for="c in clusters" :key="c.id" :value="c.id">{{ c.name }}</a-option>
                  </a-select>
                </a-form-item>
              </a-col>
              <a-col :span="8">
                <a-form-item field="target_namespace" label="命名空间" required>
                  <a-select v-model="form.target_namespace" placeholder="选择命名空间" :loading="loadingNamespaces" allow-clear @change="onNamespaceChange">
                    <a-option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</a-option>
                  </a-select>
                </a-form-item>
              </a-col>
              <a-col :span="8">
                <a-form-item field="target_workload_kind" label="工作负载类型">
                  <a-select v-model="form.target_workload_kind" placeholder="Deployment">
                    <a-option value="Deployment">Deployment</a-option>
                    <a-option value="StatefulSet">StatefulSet</a-option>
                    <a-option value="DaemonSet">DaemonSet</a-option>
                  </a-select>
                </a-form-item>
              </a-col>
            </a-row>

            <a-row :gutter="16">
              <a-col :span="12">
                <a-form-item field="target_workload_name" label="工作负载名称" required>
                  <a-select v-model="form.target_workload_name" placeholder="选择工作负载" :loading="loadingWorkloads" allow-clear>
                    <a-option v-for="wl in workloads" :key="wl.name" :value="wl.name">{{ wl.name }}</a-option>
                  </a-select>
                </a-form-item>
              </a-col>
              <a-col :span="12">
                <a-form-item field="target_container" label="容器名称">
                  <a-input v-model="form.target_container" placeholder="默认取工作负载名称" allow-clear />
                </a-form-item>
              </a-col>
            </a-row>

            <a-row :gutter="16">
              <a-col :span="8">
                <a-form-item field="deploy_env" label="部署环境">
                  <a-select v-model="form.deploy_env" placeholder="dev">
                    <a-option value="dev">开发环境</a-option>
                    <a-option value="test">测试环境</a-option>
                    <a-option value="staging">预发环境</a-option>
                    <a-option value="prod">生产环境</a-option>
                  </a-select>
                </a-form-item>
              </a-col>
              <a-col :span="8">
                <a-form-item field="require_approval" label="发布审批">
                  <a-switch v-model="form.require_approval" @change="onApprovalToggle" />
                  <template #extra>{{ form.require_approval ? '已开启，审批人可在审批策略页面配置' : '生产环境建议开启' }}</template>
                </a-form-item>
              </a-col>
            </a-row>

            <!-- 审批级别配置（开启审批时显示） -->
            <a-row v-if="form.require_approval" :gutter="16">
              <a-col :span="24">
                <div class="approval-config-section">
                  <div class="approval-config-header">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16">
                      <path d="M9 11l3 3L22 4"/>
                      <path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/>
                    </svg>
                    <span>多级审批配置</span>
                    <router-link to="/cicd/approval-policy" class="approval-policy-link">
                      全局策略设置 →
                    </router-link>
                  </div>
                  <div class="approval-levels-list" v-if="form.approval_levels && form.approval_levels.length > 0">
                    <div v-for="(level, idx) in form.approval_levels" :key="idx" class="approval-level-item">
                      <span class="level-badge">第{{ level.level }}级</span>
                      <span class="level-label">{{ level.label }}</span>
                      <span class="level-role">{{ level.role }}</span>
                      <button type="button" class="btn-remove-level" @click="removeApprovalLevel(idx)" title="移除此级">
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                      </button>
                    </div>
                  </div>
                  <div v-else class="approval-empty-hint">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/></svg>
                    尚未配置审批级别，将使用环境默认审批策略。可在「审批策略」页面统一配置。
                  </div>
                  <button type="button" class="btn-add-level" @click="addApprovalLevel">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
                    添加审批级别
                  </button>
                </div>
              </a-col>
            </a-row>
          </template>
        </div>

        <!-- 提交 -->
        <div class="form-actions">
          <a-button type="primary" @click="submit" :loading="submitting" size="large">
            {{ isEdit ? '保存修改' : '创建流水线' }}
          </a-button>
          <a-button @click="cancel" size="large">取消</a-button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { Message } from '@arco-design/web-vue'
import { createPipeline, updatePipeline, getPipelineDetail } from '@/api/cicd.js'
import { checkPipelineName } from '@/api/platform/pipeline.js'
import { getClusterList } from '@/api/cluster.js'
import { getNamespaces } from '@/api/namespace.js'
import deploymentsApi from '@/api/cluster/workloads/deployments'
import statefulsetsApi from '@/api/cluster/workloads/statefulsets'
import daemonsetsApi from '@/api/cluster/workloads/daemonsets'
import { GitOpsConfigForm } from '@/components/cicd'
import { useClusterStore } from '@/stores/cluster'

const router = useRouter()
const route = useRoute()
const clusterStore = useClusterStore()

const isEdit = computed(() => !!route.params.id)
const pipelineId = computed(() => route.params.id)
const submitting = ref(false)

const form = reactive({
  name: '',
  description: '',
  git_repo: '',
  git_branch: 'main',
  language_type: 'java',
  deploy_mode: 'gitops',
  gitops_config: {
    argo_app_name: '',
    git_manifest_repo: '',
    manifest_path: 'manifests',
    argo_project: 'default',
    target_revision: 'main',
    workflow_template: '',
    workflow_namespace: 'argo',
    image_registry: '',
    image_repo: '',
    dockerfile_path: 'Dockerfile',
    build_context: '.',
    auto_sync: true,
    prune_resource: false
  },
  auto_deploy: false,
  target_cluster_id: null,
  target_namespace: '',
  target_workload_kind: 'Deployment',
  target_workload_name: '',
  target_container: '',
  deploy_env: 'dev',
  require_approval: false,
  approval_levels: []  // 多级审批配置 [{ level, role, label }]
})

// 名称检查
const nameChecking = ref(false)
const nameAvailable = ref(null)
let nameCheckTimer = null
const checkName = async () => {
  if (!form.name) { nameAvailable.value = null; return }
  nameChecking.value = true
  try {
    const res = await checkPipelineName(form.name, isEdit.value ? parseInt(pipelineId.value) : 0)
    nameAvailable.value = res.data?.available ?? res.available
  } catch { nameAvailable.value = null }
  finally { nameChecking.value = false }
}

// 集群
const clusters = ref([])
const loadingClusters = ref(false)
const loadClusters = async () => {
  loadingClusters.value = true
  try {
    const res = await getClusterList()
    clusters.value = res.data?.clusters || res.data || []
  } catch { clusters.value = [] }
  finally { loadingClusters.value = false }
}

// 命名空间
const namespaces = ref([])
const loadingNamespaces = ref(false)
const loadNamespaces = async () => {
  if (!form.target_cluster_id) { namespaces.value = []; return }
  loadingNamespaces.value = true
  try {
    // 设置集群上下文到 store，确保后续 API Header 中的 X-Cluster-ID 正确
    const cluster = clusters.value.find(c => c.id === form.target_cluster_id)
    if (cluster) clusterStore.setCurrent(cluster)
    const res = await getNamespaces(form.target_cluster_id)
    namespaces.value = res.data?.namespaces || res.data?.list || res.data || []
  } catch { namespaces.value = [] }
  finally { loadingNamespaces.value = false }
}
const onClusterChange = () => {
  form.target_namespace = ''
  form.target_workload_name = ''
  workloads.value = []
  loadNamespaces()
}
const onNamespaceChange = () => {
  form.target_workload_name = ''
  loadWorkloads()
}

// 审批级别配置
const onApprovalToggle = (val) => {
  if (val && form.approval_levels.length === 0) {
    // 默认添加一级审批
    form.approval_levels.push({ level: 1, role: 'ops_lead', label: '运维负责人' })
  }
}
const addApprovalLevel = () => {
  const nextLevel = form.approval_levels.length + 1
  const defaults = [
    { level: 1, role: 'dev_lead', label: '开发负责人' },
    { level: 2, role: 'test_lead', label: '测试负责人' },
    { level: 3, role: 'ops_lead', label: '运维负责人' },
  ]
  const def = defaults.find(d => d.level === nextLevel) || { level: nextLevel, role: `level_${nextLevel}`, label: `第${nextLevel}级审批` }
  form.approval_levels.push({ ...def })
}
const removeApprovalLevel = (idx) => {
  form.approval_levels.splice(idx, 1)
  // 重新编号
  form.approval_levels.forEach((l, i) => l.level = i + 1)
}

// 工作负载
const workloads = ref([])
const loadingWorkloads = ref(false)
const loadWorkloads = async () => {
  if (!form.target_cluster_id || !form.target_namespace) { workloads.value = []; return }
  loadingWorkloads.value = true
  try {
    // 确保集群上下文已设置到 store
    const cluster = clusters.value.find(c => c.id === form.target_cluster_id)
    if (cluster) clusterStore.setCurrent(cluster)
    const params = { namespace: form.target_namespace }
    let res
    switch (form.target_workload_kind) {
      case 'StatefulSet': res = await statefulsetsApi.list(params); break
      case 'DaemonSet': res = await daemonsetsApi.list(params); break
      default: res = await deploymentsApi.list(params)
    }
    workloads.value = (res.data?.items || res.data?.list || res.data || []).map(w => ({ name: w.metadata?.name || w.name }))
  } catch { workloads.value = [] }
  finally { loadingWorkloads.value = false }
}

// 加载已有流水线（编辑模式）
const loadPipelineData = async () => {
  if (!isEdit.value) return
  try {
    const res = await getPipelineDetail(pipelineId.value)
    const data = res.data?.pipeline || res.data
    if (!data) return
    form.name = data.name || ''
    form.description = data.description || ''
    form.git_repo = data.git_repo || ''
    form.git_branch = data.git_branch || 'main'
    form.language_type = data.language_type || 'java'
    if (data.gitops_config) {
      Object.assign(form.gitops_config, data.gitops_config)
    }
    form.auto_deploy = data.auto_deploy || false
    form.target_cluster_id = data.target_cluster_id || null
    form.target_namespace = data.target_namespace || ''
    form.target_workload_kind = data.target_workload_kind || 'Deployment'
    form.target_workload_name = data.target_workload_name || ''
    form.target_container = data.target_container || ''
    form.deploy_env = data.deploy_env || 'dev'
    form.require_approval = data.require_approval || false
    form.approval_levels = data.approval_levels || []
    if (data.target_cluster_id) {
      await loadClusters()
      await loadNamespaces()
      await loadWorkloads()
      // 回显工作负载
      if (data.target_workload_name) {
        form.target_workload_name = data.target_workload_name
      }
    }
  } catch (err) {
    Message.error('加载流水线数据失败')
  }
}

// 提交
const submit = async () => {
  if (!form.name) { Message.warning('请输入应用名称'); return }
  if (!form.git_repo) { Message.warning('请输入 Git 仓库地址'); return }
  if (!form.git_branch) { Message.warning('请输入 Git 分支'); return }
  if (!form.gitops_config.argo_app_name) { Message.warning('请输入 ArgoCD 应用名称'); return }

  if (nameAvailable.value === false) { Message.warning('应用名称已存在'); return }

  if (form.auto_deploy) {
    if (!form.target_cluster_id) { Message.warning('请选择目标集群'); return }
    if (!form.target_namespace) { Message.warning('请选择命名空间'); return }
    if (!form.target_workload_name) { Message.warning('请选择工作负载'); return }
  }

  submitting.value = true
  try {
    const submitData = {
      name: form.name,
      description: form.description,
      git_repo: form.git_repo,
      git_branch: form.git_branch,
      language_type: form.language_type,
      deploy_mode: 'gitops',
      gitops_config: { ...form.gitops_config },
      auto_deploy: form.auto_deploy,
      target_cluster_id: form.target_cluster_id || 0,
      target_namespace: form.target_namespace || '',
      target_workload_kind: form.target_workload_kind || 'Deployment',
      target_workload_name: form.target_workload_name || '',
      target_container: form.target_container || form.target_workload_name || '',
      deploy_env: form.deploy_env || 'dev',
      require_approval: form.require_approval || false,
      approval_levels: form.require_approval ? form.approval_levels : []
    }

    let response
    if (isEdit.value) {
      response = await updatePipeline({ id: parseInt(pipelineId.value), ...submitData })
    } else {
      response = await createPipeline(submitData)
    }

    if (response.code === 0) {
      Message.success(isEdit.value ? '更新成功' : '创建成功')
      if (response.data?.warnings?.length) {
        Message.warning(response.data.warnings.join('; '))
      }
      const newId = response.data?.id || response.data?.pipeline_id || pipelineId.value
      router.push(`/cicd/pipelines/${newId}`)
    } else {
      Message.error(response.msg || '操作失败')
    }
  } catch (err) {
    Message.error(err.msg || '操作失败')
  } finally {
    submitting.value = false
  }
}

const cancel = () => router.push('/cicd/pipelines')

onMounted(() => {
  loadClusters()
  loadPipelineData()
})
</script>

<style scoped>
.gitops-create-view {
  min-height: 100vh;
  background: linear-gradient(135deg, #f5f7fa 0%, #e4e8ec 100%);
  padding: 20px 24px;
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
  font-size: 13px;
  color: #6b7280;
}
.breadcrumb-link {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #3b82f6;
  text-decoration: none;
}
.breadcrumb-link:hover { text-decoration: underline; }
.separator { color: #d1d5db; }
.current { color: #374151; font-weight: 500; }

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #fff;
  border-radius: 12px;
  padding: 20px 24px;
  margin-bottom: 20px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.06);
}
.header-content {
  display: flex;
  align-items: center;
  gap: 14px;
}
.header-icon {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  background: linear-gradient(135deg, #ccfbf1, #99f6e4);
  color: #0d9488;
  display: flex;
  align-items: center;
  justify-content: center;
}
.header-icon svg { width: 24px; height: 24px; }
.header-text h1 { font-size: 20px; font-weight: 700; color: #1f2937; margin: 0; }
.header-text p { font-size: 13px; color: #6b7280; margin: 4px 0 0; }
.header-actions { display: flex; align-items: center; gap: 8px; }

.btn-header-save {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 20px;
  background: #0d9488;
  color: #fff;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
}
.btn-header-save:hover { background: #0f766e; }
.btn-header-save:disabled { opacity: 0.6; cursor: not-allowed; }

.btn-icon {
  width: 36px;
  height: 36px;
  border: 1px solid #e5e7eb;
  background: #fff;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: #6b7280;
}
.btn-icon:hover { background: #f3f4f6; }

.form-container {
  max-width: 900px;
}
.form-card {
  background: #fff;
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 16px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.06);
}
.card-title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 15px;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 1px solid #f3f4f6;
}
.card-title svg { color: #0d9488; }

.form-actions {
  display: flex;
  gap: 12px;
  padding: 16px 0;
}

.loading-spinner-sm {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255,255,255,0.3);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
  display: inline-block;
}
@keyframes spin { to { transform: rotate(360deg); } }

/* 审批配置区域 */
.approval-config-section {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 16px;
  margin-top: 4px;
  margin-bottom: 8px;
}

.approval-config-header {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 600;
  color: #374151;
  margin-bottom: 12px;
}

.approval-policy-link {
  margin-left: auto;
  font-size: 12px;
  font-weight: 400;
  color: #3b82f6;
  text-decoration: none;
}
.approval-policy-link:hover { text-decoration: underline; }

.approval-levels-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 12px;
}

.approval-level-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
}

.level-badge {
  background: #3b82f6;
  color: #fff;
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 10px;
  white-space: nowrap;
}

.level-label {
  font-size: 13px;
  font-weight: 500;
  color: #374151;
}

.level-role {
  font-size: 11px;
  color: #94a3b8;
  font-family: monospace;
}

.btn-remove-level {
  margin-left: auto;
  background: none;
  border: none;
  cursor: pointer;
  color: #94a3b8;
  padding: 2px;
}
.btn-remove-level:hover { color: #ef4444; }

.approval-empty-hint {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #94a3b8;
  padding: 8px 0;
  margin-bottom: 12px;
}

.btn-add-level {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  background: #fff;
  border: 1px dashed #cbd5e1;
  border-radius: 6px;
  font-size: 12px;
  color: #64748b;
  cursor: pointer;
}
.btn-add-level:hover { border-color: #3b82f6; color: #3b82f6; }
</style>
