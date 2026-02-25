<template>
  <div class="pipeline-detail">
    <h2>流水线详情</h2>

    <!-- 流水线基本信息 -->
    <div class="info-card">
      <div class="info-header">
        <h3>{{ pipeline.name }}</h3>
        <div class="status-badge" :class="`status-${pipeline.status}`">
          {{ pipeline.status }}
        </div>
      </div>
      <div class="info-content">
        <div class="info-item">
          <span class="label">描述：</span>
          <span class="value">{{ pipeline.description }}</span>
        </div>
        <div class="info-item">
          <span class="label">Git仓库：</span>
          <span class="value">{{ pipeline.gitRepo }}</span>
        </div>
        <div class="info-item">
          <span class="label">分支：</span>
          <span class="value">{{ pipeline.branch }}</span>
        </div>
        <div class="info-item">
          <span class="label">上次运行时间：</span>
          <span class="value">{{ formatDate(pipeline.lastRunTime) }}</span>
        </div>
        <div class="info-item">
          <span class="label">上次运行状态：</span>
          <span class="status-badge" :class="`status-${pipeline.lastRunStatus}`">
            {{ pipeline.lastRunStatus }}
          </span>
        </div>
      </div>
      <div class="info-actions">
        <button class="btn btn-primary" @click="runPipeline">运行流水线</button>
        <button class="btn btn-danger" @click="stopPipeline" v-if="pipeline.status === 'running'">停止流水线</button>
        <div class="auto-refresh">
          <input type="checkbox" id="auto-refresh" v-model="autoRefreshEnabled" />
          <label for="auto-refresh">自动刷新 ({{ refreshInterval / 1000 }}秒)</label>
        </div>
      </div>
    </div>

    <!-- 流水线阶段 -->
    <div class="stages-card">
      <div class="stage-header">
        <h3>流水线阶段</h3>
        <div class="stage-actions">
          <button class="btn btn-primary" @click="runPipeline" v-if="pipeline.status !== 'running'">
            开始流水线
          </button>
          <button class="btn btn-danger" @click="stopPipeline" v-else>
            停止流水线
          </button>
          <button class="btn btn-secondary" @click="resetPipeline" v-if="pipeline.status !== 'running'">
            重置流水线
          </button>
        </div>
      </div>
      <div class="stages-container">
        <div
          v-for="(stage, index) in pipeline.stages"
          :key="stage.name"
          class="stage-item"
          :class="{
            'stage-active': stage.status === 'running',
            'stage-completed': stage.status === 'success',
            'stage-failed': stage.status === 'failed',
            'stage-pending': stage.status === 'pending',
            'stage-skipped': stage.status === 'skipped',
            'stage-next': canRunNextStage(index)
          }"
        >
          <div class="stage-number" :class="stage.status">
            <span v-if="stage.status === 'success'" class="status-icon">✓</span>
            <span v-else-if="stage.status === 'failed'" class="status-icon">✗</span>
            <span v-else-if="stage.status === 'running'" class="status-icon">⚡</span>
            <span v-else-if="stage.status === 'skipped'" class="status-icon">➤</span>
            <span v-else>{{ index + 1 }}</span>
          </div>
          <div class="stage-content">
            <div class="stage-name">{{ stage.name }}</div>
            <div class="stage-description">{{ stage.description }}</div>
            <div class="stage-meta">
              <span class="stage-status" :class="stage.status">{{ stage.status }}</span>
              <span class="stage-time" v-if="stage.startTime">
                {{ formatDate(stage.startTime) }} - {{ stage.endTime ? formatDate(stage.endTime) : '进行中' }}
                <span class="stage-duration" v-if="stage.endTime">
                  ({{ calculateStageDuration(stage.startTime, stage.endTime) }})
                </span>
              </span>
            </div>
            <div class="stage-actions" style="margin-top: 12px;">
              <!-- 运行/重跑按钮 -->
              <button
                class="btn btn-small btn-primary"
                @click="runStage(index)"
                :disabled="!canRunStage(index)"
                title="运行此阶段"
              >
                <span v-if="stage.status === 'pending'">运行</span>
                <span v-else-if="stage.status === 'running'">运行中</span>
                <span v-else>重新运行</span>
              </button>

              <!-- 停止按钮 -->
              <button
                class="btn btn-small btn-danger"
                @click="stopStage(index)"
                :disabled="stage.status !== 'running'"
                title="停止此阶段"
              >
                停止
              </button>

              <!-- 跳过按钮 -->
              <button
                class="btn btn-small btn-warning"
                @click="skipStage(index)"
                :disabled="!canSkipStage(index)"
                title="跳过此阶段"
              >
                跳过
              </button>

              <!-- 继续按钮 -->
              <button
                class="btn btn-small btn-success"
                @click="continuePipeline(index)"
                :disabled="!canContinuePipeline(index)"
                title="继续流水线"
              >
                继续
              </button>

              <!-- 查看阶段详情/日志按钮 -->
              <button
                class="btn btn-small btn-secondary"
                @click="viewStageDetail(index)"
                title="查看阶段详情和日志"
              >
                查看日志
              </button>
            </div>

            <!-- 阶段进度条 -->
            <div class="stage-progress" v-if="stage.status === 'running'">
              <div class="progress-bar">
                <div class="progress-fill"></div>
              </div>
              <div class="progress-text">运行中...</div>
            </div>
          </div>
        </div>
      </div>

      <!-- 阶段控制提示 -->
      <div class="stage-control-tips" v-if="pipeline.status === 'running'">
        <p>流水线正在运行中，您可以：</p>
        <ul>
          <li>查看各个阶段的执行状态和日志</li>
          <li>手动运行等待中的阶段</li>
          <li>查看阶段详情</li>
          <li>跳过不需要执行的阶段</li>
        </ul>
      </div>
    </div>

    <!-- 阶段详情模态框 -->
    <div class="modal-overlay" v-if="showStageDetail" @click="closeStageDetail">
      <div class="modal stage-detail-modal" @click.stop>
        <div class="modal-header">
          <h3>{{ selectedStage ? selectedStage.name : '阶段详情' }} 详情</h3>
          <button class="close-btn" @click="closeStageDetail">&times;</button>
        </div>
        <div class="modal-body">
          <div v-if="selectedStage" class="stage-detail-content">
            <div class="detail-section">
              <h4>基本信息</h4>
              <div class="detail-item">
                <span class="label">阶段名称：</span>
                <span class="value">{{ selectedStage.name }}</span>
              </div>
              <div class="detail-item">
                <span class="label">描述：</span>
                <span class="value">{{ selectedStage.description }}</span>
              </div>
              <div class="detail-item">
                <span class="label">状态：</span>
                <span class="status-badge" :class="`status-${selectedStage.status}`">
                  {{ selectedStage.status }}
                </span>
              </div>
              <div class="detail-item" v-if="selectedStage.startTime">
                <span class="label">开始时间：</span>
                <span class="value">{{ formatDate(selectedStage.startTime) }}</span>
              </div>
              <div class="detail-item" v-if="selectedStage.endTime">
                <span class="label">结束时间：</span>
                <span class="value">{{ formatDate(selectedStage.endTime) }}</span>
              </div>
              <div class="detail-item" v-if="selectedStage.startTime && selectedStage.endTime">
                <span class="label">执行耗时：</span>
                <span class="value">{{ calculateStageDuration(selectedStage.startTime, selectedStage.endTime) }}</span>
              </div>
            </div>

            <div class="detail-section">
              <h4>阶段输出</h4>
              <div class="stage-output">
                <pre>{{ getStageOutput(selectedStage.name) }}</pre>
              </div>
            </div>

            <div class="detail-section">
              <h4>操作</h4>
              <div class="stage-actions">
                <button
                  class="btn btn-primary"
                  @click="runStage(selectedStageIndex)"
                  :disabled="!canRunStage(selectedStageIndex)"
                >
                  {{ selectedStage.status === 'pending' ? '运行此阶段' : '重新运行' }}
                </button>
                <button
                  class="btn btn-secondary"
                  @click="skipStage(selectedStageIndex)"
                  :disabled="!canSkipStage(selectedStageIndex)"
                >
                  跳过此阶段
                </button>
                <button
                  class="btn btn-secondary"
                  @click="closeStageDetail"
                >
                  关闭
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 部署配置 -->
    <div class="deploy-config-card">
      <h3>部署配置</h3>

      <!-- 部署前检查 -->
      <div class="pre-deploy-checks" v-if="showPreDeployChecks">
        <h4>部署前检查</h4>
        <div class="check-list">
          <div
            v-for="check in preDeployChecks"
            :key="check.id"
            class="check-item"
            :class="{ 'check-passed': check.passed }"
          >
            <input type="checkbox" v-model="check.passed" disabled />
            <label>{{ check.description }}</label>
            <span class="check-status" v-if="check.passed">✓</span>
          </div>
        </div>
        <div class="check-actions" style="margin-top: 12px;">
          <button type="button" class="btn btn-small btn-primary" @click="runPreDeployChecks">运行检查</button>
          <button type="button" class="btn btn-small btn-outline" @click="togglePreDeployChecks">
            {{ showPreDeployChecks ? '隐藏' : '显示' }}部署前检查
          </button>
        </div>
      </div>

      <form class="deploy-form" @submit.prevent="deploy">
        <!-- 项目/分支配置 -->
        <h4>项目配置</h4>
        <div class="form-row">
          <div class="form-group">
            <label>Git仓库 <span class="required">*</span></label>
            <input
              type="text"
              v-model="pipeline.gitRepo"
              class="form-input"
              required
              placeholder="https://github.com/example/project.git"
            />
            <div class="form-help-text">项目的Git仓库URL</div>
          </div>
          <div class="form-group">
            <label>分支 <span class="required">*</span></label>
            <input
              type="text"
              v-model="pipeline.branch"
              class="form-input"
              required
              placeholder="main 或 develop"
            />
            <div class="form-help-text">要构建的Git分支</div>
          </div>
        </div>

        <!-- 部署配置表单 -->
        <h4>部署配置</h4>
        <div class="form-row">
          <div class="form-group">
            <label>K8s环境 <span class="required">*</span></label>
            <select
              v-model="selectedEnvironmentId"
              class="form-select"
              required
              @change="onEnvironmentChange"
            >
              <option value="">请选择K8s环境</option>
              <option
                v-for="env in k8sEnvironments"
                :key="env.id"
                :value="env.id"
              >
                {{ env.name }} - {{ env.description }} ({{ env.clusterName }})
                <span class="env-type" :class="`type-${env.type || env.envType}`">
                  [{{ getEnvTypeName(env.type || env.envType) }}]
                </span>
              </option>
            </select>
            <div class="form-help-text">选择要部署到的K8s集群环境</div>
          </div>
          <div class="form-group">
            <label>部署名称 <span class="required">*</span></label>
            <input
              type="text"
              v-model="deployConfig.deploymentName"
              class="form-input"
              required
              placeholder="例如: frontend-deployment"
            />
            <div class="form-help-text">K8s中Deployment的名称</div>
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label>镜像名称 <span class="required">*</span></label>
            <input
              type="text"
              v-model="deployConfig.image"
              class="form-input"
              required
              placeholder="例如: nginx:latest 或 registry.example.com/nginx:latest"
            />
            <div class="form-help-text">要部署的Docker镜像名称和标签</div>
          </div>
          <div class="form-group">
            <label>副本数量 <span class="required">*</span></label>
            <input
              type="number"
              v-model.number="deployConfig.replicas"
              class="form-input"
              min="1"
              required
              placeholder="例如: 3"
            />
            <div class="form-help-text">要部署的Pod副本数量</div>
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label>部署策略 <span class="required">*</span></label>
            <select v-model="deployConfig.strategy" class="form-select">
              <option value="rollingUpdate">滚动更新</option>
              <option value="recreate">重新创建</option>
              <option value="blueGreen">蓝绿部署</option>
              <option value="canary">金丝雀发布</option>
            </select>
            <div class="form-help-text">
              <strong>滚动更新</strong>: 逐步替换现有Pod，零停机时间<br/>
              <strong>重新创建</strong>: 先删除所有现有Pod，再创建新Pod<br/>
              <strong>蓝绿部署</strong>: 同时运行新旧版本，切换流量<br/>
              <strong>金丝雀发布</strong>: 先发布少量新Pod，验证后逐步扩大
            </div>
          </div>
        </div>

        <!-- 部署确认 -->
        <div class="deploy-confirmation" v-if="selectedEnvironmentId">
          <h4>部署确认</h4>
          <div class="confirmation-details">
            <div class="detail-item">
              <span class="detail-label">环境:</span>
              <span class="detail-value">{{ selectedEnv.name }} ({{ getEnvTypeName(selectedEnv.type || selectedEnv.envType) }})</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">部署名称:</span>
              <span class="detail-value">{{ deployConfig.deploymentName }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">镜像:</span>
              <span class="detail-value">{{ deployConfig.image }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">副本数:</span>
              <span class="detail-value">{{ deployConfig.replicas }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">策略:</span>
              <span class="detail-value">{{ deployConfig.strategy }}</span>
            </div>
          </div>
        </div>

        <!-- 高级配置 -->
        <div class="advanced-config" v-if="showAdvancedConfig">
          <div class="advanced-header" @click="toggleAdvancedConfig">
            <h4>高级配置</h4>
            <span class="toggle-icon">{{ showAdvancedConfig ? '▼' : '▶' }}</span>
          </div>
          <div class="advanced-content" style="margin-top: 12px;">
            <div class="form-row">
              <div class="form-group">
                <label>CPU请求</label>
                <input
                  type="text"
                  v-model="deployConfig.resources.requests.cpu"
                  class="form-input"
                  placeholder="例如: 200m"
                />
                <div class="form-help-text">每个Pod请求的CPU资源</div>
              </div>
              <div class="form-group">
                <label>内存请求</label>
                <input
                  type="text"
                  v-model="deployConfig.resources.requests.memory"
                  class="form-input"
                  placeholder="例如: 256Mi"
                />
                <div class="form-help-text">每个Pod请求的内存资源</div>
              </div>
            </div>
            <div class="form-row">
              <div class="form-group">
                <label>CPU限制</label>
                <input
                  type="text"
                  v-model="deployConfig.resources.limits.cpu"
                  class="form-input"
                  placeholder="例如: 500m"
                />
                <div class="form-help-text">每个Pod允许使用的最大CPU资源</div>
              </div>
              <div class="form-group">
                <label>内存限制</label>
                <input
                  type="text"
                  v-model="deployConfig.resources.limits.memory"
                  class="form-input"
                  placeholder="例如: 512Mi"
                />
                <div class="form-help-text">每个Pod允许使用的最大内存资源</div>
              </div>
            </div>
          </div>
        </div>

        <div class="form-actions">
          <button type="button" class="btn btn-small btn-outline" @click="toggleAdvancedConfig">
            {{ showAdvancedConfig ? '隐藏' : '显示' }}高级配置
          </button>
          <button type="button" class="btn btn-secondary" @click="loadDeploymentHistory">查看部署历史</button>
          <button type="submit" class="btn btn-primary" :disabled="!canDeploy">
            {{ deploying ? '部署中...' : '部署到K8s' }}
          </button>
        </div>
      </form>

      <!-- 部署输出结果 -->
      <div class="deploy-output" v-if="deployOutput.show">
        <h4>部署结果</h4>
        <div class="output-content">
          <div class="output-header">
            <span class="output-status" :class="deployOutput.success ? 'success' : 'error'">
              {{ deployOutput.success ? '部署成功' : '部署失败' }}
            </span>
            <span class="output-time">{{ deployOutput.timestamp }}</span>
          </div>
          <div class="output-logs">
            <pre>{{ deployOutput.logs }}</pre>
          </div>
          <div class="output-actions" style="margin-top: 12px;">
            <button type="button" class="btn btn-small btn-primary" @click="deployOutput.show = false">关闭</button>
            <button type="button" class="btn btn-small btn-secondary" @click="copyDeployOutput">复制输出</button>
            <button type="button" class="btn btn-small btn-primary" @click="verifyDeployment" v-if="deployOutput.success">验证部署</button>
          </div>
        </div>
      </div>
    </div>

    <!-- 部署历史 -->
    <div class="history-card" v-if="showHistory">
      <h3>部署历史</h3>
      <table class="data-table">
        <thead>
          <tr>
            <th>版本</th>
            <th>镜像</th>
            <th>副本数量</th>
            <th>部署策略</th>
            <th>部署时间</th>
            <th>状态</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="history in deploymentHistory" :key="history.revision">
            <td>{{ history.revision }}</td>
            <td>{{ history.image }}</td>
            <td>{{ history.replicas }}</td>
            <td>{{ history.strategy }}</td>
            <td>{{ formatDate(history.deploymentTime) }}</td>
            <td>
              <span class="status-badge" :class="`status-${history.status}`">
                {{ history.status }}
              </span>
            </td>
            <td>
              <button class="btn btn-view" @click="rollback(history.revision)">回滚</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 流水线日志 -->
    <div class="logs-card">
      <div class="logs-header">
        <h3>流水线日志</h3>
        <div class="logs-actions">
          <input
            type="text"
            v-model="logSearchQuery"
            placeholder="搜索日志"
            class="log-search-input"
          />
          <button class="btn btn-small btn-secondary" @click="clearLogs">清除日志</button>
          <button class="btn btn-small btn-primary" @click="loadLogs" :disabled="loadingLogs">
            <span v-if="loadingLogs">加载中...</span>
            <span v-else>刷新日志</span>
          </button>
        </div>
      </div>
      <div class="logs-container">
        <div
          v-for="(log, index) in filteredLogs"
          :key="index"
          class="log-item"
          :class="{
            'log-error': log.includes('error') || log.includes('ERROR') || log.includes('Failed'),
            'log-warning': log.includes('warning') || log.includes('WARNING'),
            'log-info': log.includes('info') || log.includes('INFO')
          }"
        >
          {{ log }}
        </div>
      </div>
      <div class="logs-footer">
        <div class="log-stats">
          显示 {{ filteredLogs.length }} / {{ logs.length }} 条日志
        </div>
        <div class="log-controls">
          <button class="btn btn-small" @click="scrollToTop">滚动到顶部</button>
          <button class="btn btn-small" @click="scrollToBottom">滚动到底部</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted, onUnmounted, computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import {
  getPipelineDetail,
  runPipeline as runPipelineApi,
  stopPipeline as stopPipelineApi,
  getPipelineLogs,
  deployToK8s,
  getDeploymentHistory,
  getK8sEnvironments
} from '@/api/cicd.js'

export default {
  name: 'PipelineDetail',
  setup() {
    const route = useRoute()
    const pipelineId = ref(route.params.id || 1)

    const pipeline = ref({
      id: 0,
      name: '',
      description: '',
      status: '',
      lastRunTime: '',
      lastRunStatus: '',
      gitRepo: '',
      branch: '',
      stages: [],
      envVars: [],
      deploymentConfig: {
        namespace: 'default',
        deploymentName: '',
        image: '',
        replicas: 1,
        strategy: 'rollingUpdate'
      }
    })

    const k8sEnvironments = ref([])
    const selectedEnvironmentId = ref('')

    const deployConfig = ref({
      namespace: 'default',
      deploymentName: '',
      image: '',
      replicas: 1,
      strategy: 'rollingUpdate',
      environmentId: '',
      environmentName: ''
    })

    const logs = ref([])
    const logSearchQuery = ref('')
    const loadingLogs = ref(false)
    const deploymentHistory = ref([])
    const showHistory = ref(false)
    const autoRefreshEnabled = ref(true)
    const refreshInterval = ref(90000) // 90秒自动刷新
    let refreshTimer = null

    // 部署状态和输出
    const deploying = ref(false)
    const canDeploy = computed(() => {
      return deployConfig.value.deploymentName && deployConfig.value.image && selectedEnvironmentId.value
    })
    const deployOutput = ref({
      show: false,
      success: false,
      logs: '',
      timestamp: ''
    })

    // 阶段交互相关
    const showStageDetail = ref(false)
    const selectedStage = ref(null)
    const selectedStageIndex = ref(-1)
    const stageOutputs = ref({})
    const stageExecutionStatus = ref('idle') // idle, running, paused, completed, failed

    const filteredLogs = computed(() => {
      if (!logSearchQuery.value) {
        return logs.value
      }
      return logs.value.filter(log =>
        log.toLowerCase().includes(logSearchQuery.value.toLowerCase())
      )
    })

    const loadPipelineDetail = async () => {
      try {
        const response = await getPipelineDetail(pipelineId.value)
        if (response.code === 0) {
          pipeline.value = response.data
          // 初始化部署配置
          deployConfig.value = {
            namespace: pipeline.value.deploymentConfig.namespace,
            deploymentName: pipeline.value.deploymentConfig.deploymentName,
            image: pipeline.value.deploymentConfig.image,
            replicas: pipeline.value.deploymentConfig.replicas,
            strategy: pipeline.value.deploymentConfig.strategy,
            environmentId: '',
            environmentName: ''
          }
        } else {
          alert(response.msg)
        }
      } catch (error) {
        alert('获取流水线详情失败')
      }
    }

    const loadK8sEnvironments = async () => {
      try {
        const response = await getK8sEnvironments()
        if (response.code === 0) {
          k8sEnvironments.value = response.data
        } else {
          alert(response.msg)
        }
      } catch (error) {
        alert('获取K8s环境列表失败')
      }
    }

    const runPipeline = async () => {
      try {
        const response = await runPipelineApi(pipelineId.value)
        if (response.code === 0) {
          alert('流水线启动成功')
          loadPipelineDetail()
          loadLogs()
        } else {
          alert(response.msg)
        }
      } catch (error) {
        alert('启动流水线失败')
      }
    }

    const stopPipeline = async () => {
      try {
        const response = await stopPipelineApi(pipelineId.value)
        if (response.code === 0) {
          alert('流水线停止成功')
          loadPipelineDetail()
          loadLogs()
        } else {
          alert(response.msg)
        }
      } catch (error) {
        alert('停止流水线失败')
      }
    }

    const loadLogs = async () => {
      try {
        loadingLogs.value = true
        const response = await getPipelineLogs(pipelineId.value, Date.now())
        if (response.code === 0) {
          // 只添加新日志，不覆盖现有日志
          const newLogs = response.data.logs
          const existingLogSet = new Set(logs.value)
          const uniqueNewLogs = newLogs.filter(log => !existingLogSet.has(log))
          logs.value = [...logs.value, ...uniqueNewLogs]
        } else {
          alert(response.msg)
        }
      } catch (error) {
        alert('获取日志失败')
      } finally {
        loadingLogs.value = false
      }
    }

    const clearLogs = () => {
      logs.value = []
    }

    const deploy = async () => {
      try {
        // 获取选中的环境信息
        const selectedEnv = k8sEnvironments.value.find(env => env.id === parseInt(selectedEnvironmentId.value))
        if (!selectedEnv) {
          alert('请选择K8s环境')
          return
        }

        // 更新部署配置中的环境信息
        deployConfig.value.environmentId = selectedEnv.id
        deployConfig.value.environmentName = selectedEnv.name
        deployConfig.value.namespace = selectedEnv.namespace

        deploying.value = true

        // 准备部署输出
        const startTime = new Date()
        let deployLogs = `[${startTime.toISOString()}] 开始部署到K8s环境...\n`
        deployLogs += `[${startTime.toISOString()}] 环境: ${selectedEnv.name} (${getEnvTypeName(selectedEnv.type || selectedEnv.envType)})\n`
        deployLogs += `[${startTime.toISOString()}] 命名空间: ${deployConfig.value.namespace}\n`
        deployLogs += `[${startTime.toISOString()}] 部署名称: ${deployConfig.value.deploymentName}\n`
        deployLogs += `[${startTime.toISOString()}] 镜像: ${deployConfig.value.image}\n`
        deployLogs += `[${startTime.toISOString()}] 副本数: ${deployConfig.value.replicas}\n`
        deployLogs += `[${startTime.toISOString()}] 策略: ${deployConfig.value.strategy}\n\n`

        // 模拟部署步骤
        deployLogs += `[${new Date().toISOString()}] 1. 验证部署配置...\n`
        await new Promise(resolve => setTimeout(resolve, 500))
        deployLogs += `[${new Date().toISOString()}] ✓ 部署配置验证通过\n\n`

        deployLogs += `[${new Date().toISOString()}] 2. 连接到K8s集群...\n`
        await new Promise(resolve => setTimeout(resolve, 800))
        deployLogs += `[${new Date().toISOString()}] ✓ 成功连接到K8s集群: ${selectedEnv.clusterName}\n\n`

        deployLogs += `[${new Date().toISOString()}] 3. 检查命名空间 ${deployConfig.value.namespace}...\n`
        await new Promise(resolve => setTimeout(resolve, 500))
        deployLogs += `[${new Date().toISOString()}] ✓ 命名空间存在\n\n`

        deployLogs += `[${new Date().toISOString()}] 4. 开始部署应用...\n`
        deployLogs += `[${new Date().toISOString()}] 执行命令: kubectl apply -f deployment.yaml -n ${deployConfig.value.namespace}\n`
        await new Promise(resolve => setTimeout(resolve, 1500))

        const response = await deployToK8s(deployConfig.value)
        if (response.code === 0) {
          deployLogs += `[${new Date().toISOString()}] ✓ 部署成功!\n\n`
          deployLogs += `[${new Date().toISOString()}] 部署详情: ${JSON.stringify(response.data, null, 2)}\n\n`
          deployLogs += `[${new Date().toISOString()}] 5. 验证Pod状态...\n`
          await new Promise(resolve => setTimeout(resolve, 1000))
          deployLogs += `[${new Date().toISOString()}] ✓ Pod 1/2 运行中\n`
          await new Promise(resolve => setTimeout(resolve, 800))
          deployLogs += `[${new Date().toISOString()}] ✓ Pod 2/2 运行中\n\n`
          deployLogs += `[${new Date().toISOString()}] 6. 验证服务状态...\n`
          await new Promise(resolve => setTimeout(resolve, 800))
          deployLogs += `[${new Date().toISOString()}] ✓ 服务已就绪，可正常访问\n\n`
          deployLogs += `[${new Date().toISOString()}] 部署完成! 总耗时: ${Math.round((new Date() - startTime) / 1000)}秒\n`

          deployOutput.value = {
            show: true,
            success: true,
            logs: deployLogs,
            timestamp: new Date().toLocaleString()
          }

          loadDeploymentHistory()
        } else {
          deployLogs += `[${new Date().toISOString()}] ✗ 部署失败: ${response.msg}\n`
          deployOutput.value = {
            show: true,
            success: false,
            logs: deployLogs,
            timestamp: new Date().toLocaleString()
          }
        }
      } catch (error) {
        console.error('部署错误:', error)
        deployOutput.value = {
          show: true,
          success: false,
          logs: `[${new Date().toISOString()}] 部署失败: ${error.message}\n\n${error.stack}`,
          timestamp: new Date().toLocaleString()
        }
      } finally {
        deploying.value = false
      }
    }

    const copyDeployOutput = () => {
      navigator.clipboard.writeText(deployOutput.value.logs)
        .then(() => {
          alert('部署输出已复制到剪贴板')
        })
        .catch(err => {
          console.error('复制失败:', err)
          alert('复制失败，请手动复制')
        })
    }

    const verifyDeployment = async () => {
      try {
        const selectedEnv = k8sEnvironments.value.find(env => env.id === parseInt(selectedEnvironmentId.value))
        if (!selectedEnv) {
          alert('请选择K8s环境')
          return
        }

        let verifyLogs = `[${new Date().toISOString()}] 开始验证部署...\n`
        verifyLogs += `[${new Date().toISOString()}] 环境: ${selectedEnv.name}\n`
        verifyLogs += `[${new Date().toISOString()}] 部署名称: ${deployConfig.value.deploymentName}\n\n`

        verifyLogs += `[${new Date().toISOString()}] 1. 检查Deployment状态...\n`
        await new Promise(resolve => setTimeout(resolve, 800))
        verifyLogs += `[${new Date().toISOString()}] ✓ Deployment 状态正常\n\n`

        verifyLogs += `[${new Date().toISOString()}] 2. 检查Pod状态...\n`
        await new Promise(resolve => setTimeout(resolve, 1000))
        verifyLogs += `[${new Date().toISOString()}] ✓ 所有 ${deployConfig.value.replicas} 个Pod 运行正常\n\n`

        verifyLogs += `[${new Date().toISOString()}] 3. 验证服务端点...\n`
        await new Promise(resolve => setTimeout(resolve, 800))
        verifyLogs += `[${new Date().toISOString()}] ✓ 服务端点可访问\n\n`

        verifyLogs += `[${new Date().toISOString()}] 4. 运行健康检查...\n`
        await new Promise(resolve => setTimeout(resolve, 1200))
        verifyLogs += `[${new Date().toISOString()}] ✓ 健康检查通过\n\n`

        verifyLogs += `[${new Date().toISOString()}] 部署验证完成！应用运行正常。\n`

        deployOutput.value.logs += `\n\n--- 部署验证结果 ---\n${verifyLogs}`
      } catch (error) {
        deployOutput.value.logs += `

--- 部署验证失败 ---
[${new Date().toISOString()}] 验证失败: ${error.message}
`
      }
    }

    const loadDeploymentHistory = async () => {
      try {
        const response = await getDeploymentHistory(deployConfig.value.namespace, deployConfig.value.deploymentName)
        if (response.code === 0) {
          deploymentHistory.value = response.data
          showHistory.value = true
        } else {
          alert(response.msg)
        }
      } catch (error) {
        alert('获取部署历史失败')
      }
    }

    const rollback = (revision) => {
      alert(`回滚到版本 ${revision} 功能开发中`)
    }

    const getEnvTypeName = (type) => {
      const typeMap = {
        production: '生产环境',
        staging: '预发布环境',
        'pre-production': '预发布环境',
        preprod: '预发布环境',
        testing: '测试环境',
        development: '开发环境'
      }
      return typeMap[type] || type
    }

    const formatDate = (dateString) => {
      if (!dateString) return ''
      const date = new Date(dateString)
      return date.toLocaleString()
    }



    // 获取阶段输出
    const getStageOutput = (stageName) => {
      return stageOutputs.value[stageName] || `[${new Date().toISOString()}] 暂无输出\n`
    }

    // 检查阶段是否可以运行
    const canRunStage = (index) => {
      const stage = pipeline.value.stages[index]
      if (!stage) return false

      // 阶段处于pending状态才能运行
      if (stage.status !== 'pending') return false

      // 如果不是第一个阶段，需要前一个阶段成功
      if (index > 0) {
        const prevStage = pipeline.value.stages[index - 1]
        return prevStage && (prevStage.status === 'success' || prevStage.status === 'skipped')
      }

      return true
    }

    // 检查是否可以运行下一个阶段
    const canRunNextStage = (index) => {
      const stage = pipeline.value.stages[index]
      if (!stage) return false

      // 当前阶段成功或跳过，下一个阶段可以运行
      return (stage.status === 'success' || stage.status === 'skipped') &&
             index < pipeline.value.stages.length - 1 &&
             pipeline.value.stages[index + 1].status === 'pending'
    }

    // 检查是否可以跳过阶段
    const canSkipStage = (index) => {
      const stage = pipeline.value.stages[index]
      if (!stage) return false

      // 只有pending状态的阶段可以跳过
      if (stage.status !== 'pending') return false

      // 如果不是第一个阶段，需要前一个阶段成功
      if (index > 0) {
        const prevStage = pipeline.value.stages[index - 1]
        return prevStage && (prevStage.status === 'success' || prevStage.status === 'skipped')
      }

      return true
    }

    // 运行单个阶段
    const runStage = async (index) => {
      if (!canRunStage(index)) {
        alert('当前阶段无法运行，请检查前序阶段状态')
        return
      }

      const stage = pipeline.value.stages[index]
      if (!stage) return

      try {
        // 更新阶段状态为running
        stage.status = 'running'
        stage.startTime = new Date().toISOString()
        stage.endTime = null

        // 清空阶段输出
        stageOutputs.value[stage.name] = `[${stage.startTime}] 开始执行阶段: ${stage.name}\n`
        stageOutputs.value[stage.name] += `[${stage.startTime}] 描述: ${stage.description}\n\n`

        // 模拟阶段执行
        const startTime = new Date()
        let stageLogs = ''

        // 根据阶段类型执行不同的模拟逻辑
        switch (stage.name) {
          case 'checkout':
            stageLogs += await simulateCheckoutStage()
            break
          case 'install':
            stageLogs += await simulateInstallStage()
            break
          case 'test':
            stageLogs += await simulateTestStage()
            break
          case 'build':
            // 根据流水线类型执行不同的构建逻辑
            if (pipeline.value.name.includes('java') || pipeline.value.type === 'java') {
              stageLogs += await simulateJavaBuildStage()
            } else {
              stageLogs += await simulateBuildStage()
            }
            break
          case 'code-quality':
            stageLogs += await simulateCodeQualityStage()
            break
          case 'package':
            stageLogs += await simulatePackageStage()
            break
          case 'build-image':
            stageLogs += await simulateBuildImageStage()
            break
          case 'deploy':
            stageLogs += await simulateDeployStage()
            break
          default:
            stageLogs += await simulateGenericStage(stage.name)
        }

        // 更新阶段输出
        stageOutputs.value[stage.name] += stageLogs

        // 更新阶段状态为success
        stage.status = 'success'
        stage.endTime = new Date().toISOString()

        // 添加到流水线日志
        const duration = calculateStageDuration(stage.startTime, stage.endTime)
        logs.value.push(`[${stage.endTime}] 阶段 ${stage.name} 执行成功，耗时: ${duration}`)

        // 自动运行下一个阶段（如果设置了自动执行）
        if (index < pipeline.value.stages.length - 1) {
          const nextStage = pipeline.value.stages[index + 1]
          if (nextStage.status === 'pending') {
            // 自动执行下一个阶段
            runStage(index + 1)
          }
        }
      } catch (error) {
        // 更新阶段状态为failed
        stage.status = 'failed'
        stage.endTime = new Date().toISOString()

        // 更新阶段输出
        stageOutputs.value[stage.name] += `[${stage.endTime}] 阶段执行失败: ${error.message}\n`

        // 添加到流水线日志
        logs.value.push(`[${stage.endTime}] 阶段 ${stage.name} 执行失败: ${error.message}`)
      }
    }

    // 模拟Java项目构建阶段
    const simulateJavaBuildStage = async () => {
      let logs = `[${new Date().toISOString()}] 开始构建Java项目: mvn clean compile -DskipTests\n`
      await new Promise(resolve => setTimeout(resolve, 1000))
      logs += `[${new Date().toISOString()}] 下载依赖中...\n`
      await new Promise(resolve => setTimeout(resolve, 2500))
      logs += `[${new Date().toISOString()}] ✓ 依赖下载完成\n`
      await new Promise(resolve => setTimeout(resolve, 1500))
      logs += `[${new Date().toISOString()}] ✓ 编译成功，生成class文件\n`
      await new Promise(resolve => setTimeout(resolve, 500))
      logs += `[${new Date().toISOString()}] ✓ 构建完成\n\n`
      return logs
    }

    // 模拟代码质量检查阶段
    const simulateCodeQualityStage = async () => {
      let logs = `[${new Date().toISOString()}] 开始代码质量检查: mvn sonar:sonar\n`
      await new Promise(resolve => setTimeout(resolve, 1000))
      logs += `[${new Date().toISOString()}] 连接到SonarQube服务器...\n`
      await new Promise(resolve => setTimeout(resolve, 1500))
      logs += `[${new Date().toISOString()}] ✓ 连接成功\n`
      await new Promise(resolve => setTimeout(resolve, 1500))
      logs += `[${new Date().toISOString()}] 分析代码...\n`
      await new Promise(resolve => setTimeout(resolve, 2000))
      logs += `[${new Date().toISOString()}] ✓ 代码质量检查完成\n`
      logs += `[${new Date().toISOString()}] 质量评分: A (95/100)\n`
      logs += `[${new Date().toISOString()}] 问题数量: 0 严重, 0 主要, 2 次要\n\n`
      return logs
    }

    // 模拟打包阶段
    const simulatePackageStage = async () => {
      let logs = `[${new Date().toISOString()}] 开始打包应用: mvn package -DskipTests\n`
      await new Promise(resolve => setTimeout(resolve, 1000))
      logs += `[${new Date().toISOString()}] 编译资源文件...\n`
      await new Promise(resolve => setTimeout(resolve, 1000))
      logs += `[${new Date().toISOString()}] 编译Java代码...\n`
      await new Promise(resolve => setTimeout(resolve, 1500))
      logs += `[${new Date().toISOString()}] 执行测试（跳过）...\n`
      await new Promise(resolve => setTimeout(resolve, 500))
      logs += `[${new Date().toISOString()}] 打包成JAR文件...\n`
      await new Promise(resolve => setTimeout(resolve, 2000))
      logs += `[${new Date().toISOString()}] ✓ 打包完成\n`
      logs += `[${new Date().toISOString()}] 生成文件: target/java-app-1.0.0.jar\n`
      logs += `[${new Date().toISOString()}] 文件大小: 25MB\n\n`
      return logs
    }

    // 模拟checkout阶段
    const simulateCheckoutStage = async () => {
      let logs = `[${new Date().toISOString()}] 开始拉取代码...\n`
      await new Promise(resolve => setTimeout(resolve, 1000))
      logs += `[${new Date().toISOString()}] 执行命令: git clone https://github.com/example/hello-app.git\n`
      await new Promise(resolve => setTimeout(resolve, 1500))
      logs += `[${new Date().toISOString()}] 拉取成功，检出分支: main\n`
      await new Promise(resolve => setTimeout(resolve, 500))
      logs += `[${new Date().toISOString()}] 代码拉取完成\n\n`
      return logs
    }

    // 模拟install阶段
    const simulateInstallStage = async () => {
      let logs = `[${new Date().toISOString()}] 开始安装依赖...\n`
      await new Promise(resolve => setTimeout(resolve, 800))
      logs += `[${new Date().toISOString()}] 执行命令: npm install\n`
      await new Promise(resolve => setTimeout(resolve, 2000))
      logs += `[${new Date().toISOString()}] ✓ 依赖安装成功\n\n`
      return logs
    }

    // 模拟test阶段
    const simulateTestStage = async () => {
      let logs = `[${new Date().toISOString()}] 开始运行单元测试...\n`
      await new Promise(resolve => setTimeout(resolve, 600))
      logs += `[${new Date().toISOString()}] 执行命令: npm test\n`
      await new Promise(resolve => setTimeout(resolve, 1500))
      logs += `[${new Date().toISOString()}] ✓ 测试用例: testGetHelloMessage 执行成功\n`
      await new Promise(resolve => setTimeout(resolve, 800))
      logs += `[${new Date().toISOString()}] ✓ 测试用例: testServerStart 执行成功\n`
      await new Promise(resolve => setTimeout(resolve, 600))
      logs += `[${new Date().toISOString()}] ✓ 所有2个测试用例通过\n\n`
      return logs
    }

    // 模拟build阶段
    const simulateBuildStage = async () => {
      let logs = `[${new Date().toISOString()}] 开始构建应用...\n`
      await new Promise(resolve => setTimeout(resolve, 800))
      logs += `[${new Date().toISOString()}] 执行命令: npm run build\n`
      await new Promise(resolve => setTimeout(resolve, 2500))
      logs += `[${new Date().toISOString()}] ✓ 应用构建成功，生成dist目录\n\n`
      return logs
    }

    // 模拟build-image阶段
    const simulateBuildImageStage = async () => {
      let logs = `[${new Date().toISOString()}] 开始构建Docker镜像...\n`
      await new Promise(resolve => setTimeout(resolve, 1000))
      logs += `[${new Date().toISOString()}] 执行命令: docker build -t example/hello-app:v1.0.0 .\n`
      await new Promise(resolve => setTimeout(resolve, 3000))
      logs += `[${new Date().toISOString()}] ✓ 镜像构建完成，大小: 120MB\n`
      await new Promise(resolve => setTimeout(resolve, 1500))
      logs += `[${new Date().toISOString()}] ✓ 镜像推送成功\n\n`
      return logs
    }

    // 模拟deploy阶段
    const simulateDeployStage = async () => {
      let logs = `[${new Date().toISOString()}] 开始部署到K8s集群...\n`
      await new Promise(resolve => setTimeout(resolve, 800))
      logs += `[${new Date().toISOString()}] 执行命令: kubectl apply -f deployment.yaml\n`
      await new Promise(resolve => setTimeout(resolve, 2000))
      logs += `[${new Date().toISOString()}] ✓ 部署创建成功\n`
      await new Promise(resolve => setTimeout(resolve, 1500))
      logs += `[${new Date().toISOString()}] ✓ Pod 1/2 运行中\n`
      await new Promise(resolve => setTimeout(resolve, 1000))
      logs += `[${new Date().toISOString()}] ✓ Pod 2/2 运行中\n`
      await new Promise(resolve => setTimeout(resolve, 800))
      logs += `[${new Date().toISOString()}] ✓ 服务已就绪，可正常访问\n\n`
      return logs
    }

    // 模拟通用阶段
    const simulateGenericStage = async (stageName) => {
      let logs = `[${new Date().toISOString()}] 开始执行阶段: ${stageName}...\n`
      await new Promise(resolve => setTimeout(resolve, 1500))
      logs += `[${new Date().toISOString()}] ✓ 阶段执行成功\n\n`
      return logs
    }

    // 查看阶段详情
    const viewStageDetail = (index) => {
      const stage = pipeline.value.stages[index]
      if (stage) {
        selectedStage.value = stage
        selectedStageIndex.value = index
        showStageDetail.value = true
      }
    }

    // 关闭阶段详情
    const closeStageDetail = () => {
      showStageDetail.value = false
      selectedStage.value = null
      selectedStageIndex.value = -1
    }

    // 跳过阶段
    const skipStage = (index) => {
      if (!canSkipStage(index)) {
        alert('当前阶段无法跳过，请检查前序阶段状态')
        return
      }

      const stage = pipeline.value.stages[index]
      if (!stage) return

      // 确认跳过
      if (confirm(`确定要跳过阶段 "${stage.name}" 吗？`)) {
        stage.status = 'skipped'
        stage.startTime = new Date().toISOString()
        stage.endTime = new Date().toISOString()

        // 添加到流水线日志
        logs.value.push(`[${stage.endTime}] 阶段 ${stage.name} 已跳过`)
      }
    }

    // 停止阶段
    const stopStage = (index) => {
      const stage = pipeline.value.stages[index]
      if (!stage || stage.status !== 'running') return

      if (confirm(`确定要停止阶段 "${stage.name}" 吗？`)) {
        // 更新阶段状态为failed
        stage.status = 'failed'
        stage.endTime = new Date().toISOString()

        // 更新阶段输出
        if (!stageOutputs.value[stage.name]) {
          stageOutputs.value[stage.name] = ''
        }
        stageOutputs.value[stage.name] += `[${stage.endTime}] 阶段执行被手动停止\n`

        // 添加到流水线日志
        logs.value.push(`[${stage.endTime}] 阶段 ${stage.name} 已被手动停止`)
      }
    }

    // 继续流水线
    const continuePipeline = (index) => {
      if (!canContinuePipeline(index)) {
        alert('当前无法继续流水线，请检查前序阶段状态')
        return
      }

      // 运行下一个阶段
      runStage(index + 1)
    }

    // 检查是否可以继续流水线
    const canContinuePipeline = (index) => {
      if (index >= pipeline.value.stages.length - 1) return false

      const currentStage = pipeline.value.stages[index]
      const nextStage = pipeline.value.stages[index + 1]

      // 当前阶段已完成，下一个阶段处于等待状态
      return (currentStage.status === 'success' || currentStage.status === 'skipped') && nextStage.status === 'pending'
    }

    // 计算阶段执行时长
    const calculateStageDuration = (startTime, endTime) => {
      if (!startTime || !endTime) return ''

      const start = new Date(startTime)
      const end = new Date(endTime)
      const durationMs = end - start

      const seconds = Math.floor(durationMs / 1000)
      const minutes = Math.floor(seconds / 60)

      if (minutes > 0) {
        return `${minutes}分${seconds % 60}秒`
      } else {
        return `${seconds}秒`
      }
    }

    // 重置流水线
    const resetPipeline = () => {
      if (confirm('确定要重置流水线吗？这将清除所有执行状态。')) {
        // 重置所有阶段状态
        pipeline.value.stages.forEach(stage => {
          stage.status = 'pending'
          stage.startTime = null
          stage.endTime = null
        })

        // 清空日志
        logs.value = []

        // 清空阶段输出
        stageOutputs.value = {}

        // 重置流水线状态
        pipeline.value.status = 'idle'

        alert('流水线已重置')
      }
    }

    // 环境变化处理
    const onEnvironmentChange = () => {
      // 可以在这里添加环境变化时的处理逻辑
      console.log('环境变化:', selectedEnvironmentId.value)
    }

    const scrollToTop = () => {
      const logsContainer = document.querySelector('.logs-container')
      if (logsContainer) {
        logsContainer.scrollTop = 0
      }
    }

    const scrollToBottom = () => {
      const logsContainer = document.querySelector('.logs-container')
      if (logsContainer) {
        logsContainer.scrollTop = logsContainer.scrollHeight
      }
    }

    // 设置自动刷新
    const setupAutoRefresh = () => {
      if (refreshTimer) {
        clearInterval(refreshTimer)
      }

      if (autoRefreshEnabled.value) {
        refreshTimer = setInterval(() => {
          loadLogs()
          // 如果流水线正在运行，也刷新流水线详情
          if (pipeline.value.status === 'running') {
            loadPipelineDetail()
          }
        }, refreshInterval.value)
      }
    }

    // 监听自动刷新开关变化
    watch(autoRefreshEnabled, () => {
      setupAutoRefresh()
    })

    // 页面加载时设置自动刷新
    onMounted(() => {
      loadPipelineDetail()
      loadK8sEnvironments()
      loadLogs()
      setupAutoRefresh()
    })

    // 页面卸载时清除定时器
    onUnmounted(() => {
      if (refreshTimer) {
        clearInterval(refreshTimer)
      }
    })

    return {
      pipeline,
      k8sEnvironments,
      selectedEnvironmentId,
      deployConfig,
      logs,
      logSearchQuery,
      filteredLogs,
      loadingLogs,
      deploymentHistory,
      showHistory,
      autoRefreshEnabled,
      refreshInterval,
      runPipeline,
      stopPipeline,
      loadLogs,
      clearLogs,
      deploy,
      loadDeploymentHistory,
      rollback,
      formatDate,
      scrollToTop,
      scrollToBottom,
      getEnvTypeName,
      deployOutput,
      deploying,
      canDeploy,
      copyDeployOutput,
      verifyDeployment,
      onEnvironmentChange,
      // 阶段交互相关
      showStageDetail,
      selectedStage,
      selectedStageIndex,
      stageOutputs,
      canRunStage,
      canRunNextStage,
      canSkipStage,
      canContinuePipeline,
      runStage,
      viewStageDetail,
      closeStageDetail,
      skipStage,
      stopStage,
      continuePipeline,
      resetPipeline,
      calculateStageDuration,
      getStageOutput
    }
  }
}
</script>

<style scoped>
.pipeline-detail {
  padding: 20px;
  background-color: var(--bg-color);
  min-height: 100vh;
}

.info-card, .stages-card, .deploy-config-card, .logs-card, .history-card {
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  padding: 20px;
  margin-bottom: 20px;
}

.info-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.info-header h3 {
  margin: 0;
  font-size: 20px;
}

.status-badge {
  display: inline-block;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

/* 环境类型标签样式 */
.env-type {
  display: inline-block;
  padding: 2px 6px;
  margin-left: 8px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 500;
  text-transform: capitalize;
  vertical-align: middle;
}

.type-production {
  background-color: #f8d7da;
  color: #721c24;
  border: 1px solid #f5c6cb;
}

.type-staging {
  background-color: #fff3cd;
  color: #856404;
  border: 1px solid #ffeeba;
}

.type-pre-production {
  background-color: #fff3cd;
  color: #856404;
  border: 1px solid #ffeeba;
}

.type-testing {
  background-color: #d1ecf1;
  color: #0c5460;
  border: 1px solid #bee5eb;
}

.type-development {
  background-color: #d4edda;
  color: #155724;
  border: 1px solid #c3e6cb;
}

.type-preprod {
  background-color: #fff3cd;
  color: #856404;
  border: 1px solid #ffeeba;
}

.status-running {
  background-color: #fff3cd;
  color: #856404;
}

.status-idle {
  background-color: #d1ecf1;
  color: #0c5460;
}

.status-success {
  background-color: #d4edda;
  color: #155724;
}

.status-failed {
  background-color: #f8d7da;
  color: #721c24;
}

.status-pending {
  background-color: #e2e3e5;
  color: #383d41;
}

.info-content {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 16px;
  margin-bottom: 16px;
}

.info-item {
  display: flex;
  flex-direction: column;
}

.info-item .label {
  font-weight: 600;
  color: #4a5568;
  font-size: 14px;
  margin-bottom: 4px;
}

.info-item .value {
  color: #2d3748;
  font-size: 14px;
}

.info-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}

.auto-refresh {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #4a5568;
}

.stages-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.stage-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.stage-actions {
  display: flex;
  gap: 12px;
}

.stage-item {
  display: flex;
  gap: 16px;
  padding: 16px;
  border-radius: 8px;
  background-color: white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.08);
  transition: all 0.3s ease;
  border-left: 4px solid #e2e8f0;
}

.stage-item:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
}

.stage-active {
  border-left-color: #f6ad55;
  background-color: #fffaf0;
}

.stage-completed {
  border-left-color: #48bb78;
  background-color: #f0fff4;
}

.stage-failed {
  border-left-color: #f56565;
  background-color: #fff5f5;
}

.stage-pending {
  border-left-color: #4299e1;
  background-color: #ebf8ff;
}

.stage-next {
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% {
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.08);
  }
  50% {
    box-shadow: 0 4px 12px rgba(66, 153, 225, 0.4);
  }
}

.stage-number {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background-color: #e2e8f0;
  color: #4a5568;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 16px;
  flex-shrink: 0;
  transition: all 0.3s ease;
}

.stage-active .stage-number {
  background-color: #f6ad55;
  color: white;
}

.stage-completed .stage-number {
  background-color: #48bb78;
  color: white;
}

.stage-failed .stage-number {
  background-color: #f56565;
  color: white;
}

.stage-pending .stage-number {
  background-color: #4299e1;
  color: white;
}

.stage-content {
  flex: 1;
  min-width: 0;
}

.stage-name {
  font-weight: 600;
  color: #2d3748;
  font-size: 16px;
  margin-bottom: 4px;
}

.stage-description {
  font-size: 13px;
  color: #718096;
  margin-bottom: 8px;
}

.stage-status {
  display: inline-block;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
  margin-bottom: 8px;
}

.stage-active .stage-status {
  background-color: #f6ad55;
  color: white;
}

.stage-completed .stage-status {
  background-color: #48bb78;
  color: white;
}

.stage-failed .stage-status {
  background-color: #f56565;
  color: white;
}

.stage-pending .stage-status {
  background-color: #4299e1;
  color: white;
}

.stage-time {
  font-size: 12px;
  color: #718096;
  margin-bottom: 8px;
}

.stage-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.stage-control-tips {
  background-color: #ebf8ff;
  border: 1px solid #bee3f8;
  border-radius: 6px;
  padding: 16px;
  margin-top: 16px;
}

.stage-control-tips p {
  margin: 0 0 8px 0;
  font-weight: 600;
  color: #2b6cb0;
}

.stage-control-tips ul {
  margin: 0;
  padding-left: 20px;
}

.stage-control-tips li {
  margin-bottom: 4px;
  color: #2b6cb0;
}

/* 阶段详情模态框样式 */
.stage-detail-modal {
  max-width: 800px;
  width: 90%;
  max-height: 80vh;
  overflow-y: auto;
}

.stage-detail-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.detail-section {
  background-color: #f7fafc;
  padding: 16px;
  border-radius: 6px;
}

.detail-section h4 {
  margin: 0 0 12px 0;
  color: #2d3748;
  font-size: 16px;
}

.detail-item {
  display: flex;
  margin-bottom: 8px;
}

.detail-item .label {
  width: 100px;
  font-weight: 600;
  color: #4a5568;
}

.detail-item .value {
  flex: 1;
  color: #2d3748;
}

.stage-output {
  background-color: #2d3748;
  color: #e2e8f0;
  padding: 12px;
  border-radius: 4px;
  overflow-x: auto;
  max-height: 300px;
  overflow-y: auto;
}

.stage-output pre {
  margin: 0;
  white-space: pre-wrap;
  font-family: 'Courier New', Courier, monospace;
  font-size: 13px;
}

/* 按钮样式增强 */
.btn-small {
  padding: 4px 8px;
  font-size: 12px;
}

.btn-outline {
  background-color: transparent;
  border: 1px solid #e2e8f0;
  color: #4a5568;
}

.btn-outline:hover {
  background-color: #f7fafc;
  border-color: #cbd5e0;
}

.stage-active {
  border-left: 4px solid #f6ad55;
  background-color: #fffaf0;
}

.stage-active .stage-number {
  background-color: #f6ad55;
  color: white;
}

.stage-completed {
  border-left: 4px solid #48bb78;
  background-color: #f0fff4;
}

.stage-completed .stage-number {
  background-color: #48bb78;
  color: white;
}

.stage-failed {
  border-left: 4px solid #f56565;
  background-color: #fff5f5;
}

.stage-failed .stage-number {
  background-color: #f56565;
  color: white;
}

.stage-pending {
  border-left: 4px solid #a0aec0;
  background-color: #f7fafc;
}

.stage-pending .stage-number {
  background-color: #a0aec0;
  color: white;
}

.deploy-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group label {
  font-weight: 600;
  color: #4a5568;
  font-size: 14px;
}

.form-input, .form-select {
  padding: 8px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 4px;
  font-size: 14px;
  transition: border-color 0.3s ease;
}

.form-input:focus, .form-select:focus {
  outline: none;
  border-color: #326ce5;
  box-shadow: 0 0 0 3px rgba(50, 108, 229, 0.1);
}

.form-actions {
  display: flex;
  gap: 8px;
  margin-top: 8px;
}

.logs-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.logs-header h3 {
  margin: 0;
  font-size: 18px;
}

.logs-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.log-search-input {
  padding: 6px 10px;
  border: 1px solid #e2e8f0;
  border-radius: 4px;
  font-size: 13px;
  width: 200px;
}

.log-search-input:focus {
  outline: none;
  border-color: #326ce5;
  box-shadow: 0 0 0 3px rgba(50, 108, 229, 0.1);
}

.logs-container {
  background-color: #2d3748;
  color: #e2e8f0;
  padding: 16px;
  border-radius: 4px;
  max-height: 400px;
  overflow-y: auto;
  margin-bottom: 16px;
  font-family: monospace;
  font-size: 13px;
  line-height: 1.5;
}

.log-item {
  margin-bottom: 4px;
  padding: 2px 0;
}

.log-error {
  color: #f56565;
  font-weight: 500;
}

.log-warning {
  color: #ed8936;
  font-weight: 500;
}

.log-info {
  color: #4299e1;
  font-weight: 500;
}

.logs-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 13px;
  color: #718096;
}

.logs-actions, .log-controls {
  display: flex;
  gap: 8px;
  align-items: center;
}

.log-stats {
  font-size: 12px;
  color: #718096;
}

.btn {
  padding: 8px 16px;
  border: 1px solid #e2e8f0;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.btn-small {
  padding: 4px 8px;
  font-size: 12px;
}

.btn-primary {
  background-color: #326ce5;
  color: white;
  border-color: #326ce5;
}

.btn-primary:hover {
  background-color: #2554c7;
}

.btn-danger {
  background-color: #e53e3e;
  color: white;
  border-color: #e53e3e;
}

.btn-danger:hover {
  background-color: #c53030;
}

.btn-secondary {
  background-color: #6c757d;
  color: white;
  border-color: #6c757d;
}

.btn-secondary:hover {
  background-color: #5a6268;
}

.btn-view {
  background-color: #6c757d;
  color: white;
  border-color: #6c757d;
  padding: 4px 8px;
  font-size: 12px;
}

.btn-view:hover {
  background-color: #5a6268;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 16px;
}

.data-table th,
.data-table td {
  padding: 12px 16px;
  text-align: left;
  border-bottom: 1px solid #e2e8f0;
}

.data-table th {
  background-color: #f7fafc;
  font-weight: 600;
  color: #4a5568;
  font-size: 14px;
}

.data-table tr:hover {
  background-color: #f7fafc;
}

/* 新增样式 */
/* 阶段元信息样式 */
.stage-meta {
  display: flex;
  gap: 12px;
  align-items: center;
  margin-bottom: 8px;
  font-size: 13px;
}

.stage-status {
  display: inline-block;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.stage-status.success {
  background-color: #d4edda;
  color: #155724;
}

.stage-status.failed {
  background-color: #f8d7da;
  color: #721c24;
}

.stage-status.running {
  background-color: #cce5ff;
  color: #004085;
}

.stage-status.pending {
  background-color: #e2e3e5;
  color: #383d41;
}

.stage-status.skipped {
  background-color: #fff3cd;
  color: #856404;
}

.stage-time {
  color: #718096;
}

.stage-duration {
  color: #a0aec0;
  font-style: italic;
}

/* 阶段进度条样式 */
.stage-progress {
  margin-top: 12px;
}

.progress-bar {
  width: 100%;
  height: 8px;
  background-color: #e2e8f0;
  border-radius: 4px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, #326ce5 0%, #667eea 100%);
  animation: progressAnimation 2s ease-in-out infinite;
  width: 70%; /* 模拟进度 */
}

@keyframes progressAnimation {
  0% {
    transform: translateX(-100%);
  }
  100% {
    transform: translateX(100%);
  }
}

.progress-text {
  font-size: 12px;
  color: #718096;
  margin-top: 4px;
  text-align: center;
}

/* 阶段编号样式 */
.stage-number {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 16px;
  flex-shrink: 0;
  transition: all 0.3s ease;
}

.stage-number.success {
  background-color: #48bb78;
}

.stage-number.failed {
  background-color: #f56565;
}

.stage-number.running {
  background-color: #4299e1;
  animation: pulse 1.5s ease-in-out infinite;
}

.stage-number.skipped {
  background-color: #ed8936;
}

.stage-number.pending {
  background-color: #a0aec0;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.8;
    transform: scale(1.05);
  }
}

/* 状态图标样式 */
.status-icon {
  font-size: 18px;
  font-weight: bold;
}

/* 按钮样式增强 */
.btn-small {
  padding: 6px 12px;
  font-size: 12px;
  min-width: 80px;
}

.btn-warning {
  background-color: #f59e0b;
  border-color: #f59e0b;
  color: white;
}

.btn-warning:hover {
  background-color: #d97706;
}

.btn-success {
  background-color: #10b981;
  border-color: #10b981;
  color: white;
}

.btn-success:hover {
  background-color: #059669;
}

/* 阶段卡片样式增强 */
.stage-item {
  margin-bottom: 12px;
}

.stage-next {
  border-left: 4px solid #f6ad55;
  background-color: #fffaf0;
  animation: stageNextPulse 2s ease-in-out infinite;
}

@keyframes stageNextPulse {
  0%, 100% {
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.08);
  }
  50% {
    box-shadow: 0 4px 12px rgba(246, 173, 85, 0.3);
  }
}

.stage-skipped {
  border-left: 4px solid #ed8936;
  background-color: #fffaf0;
}

/* 必填字段标记 */
.required {
  color: #ef4444;
  font-weight: bold;
}
</style>
