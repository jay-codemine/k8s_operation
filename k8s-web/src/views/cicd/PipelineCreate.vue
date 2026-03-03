<template>
  <div class="pipeline-wizard">
    <!-- 顶部标题栏 -->
    <div class="wizard-header">
      <div class="header-content">
        <div class="header-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12 2L2 7l10 5 10-5-10-5z"/>
            <path d="M2 17l10 5 10-5"/>
            <path d="M2 12l10 5 10-5"/>
          </svg>
        </div>
        <div class="header-text">
          <h1>{{ isEdit ? '编辑流水线' : '创建流水线' }}</h1>
          <p>配置 CI/CD 流水线，实现代码自动构建和部署</p>
        </div>
      </div>
      <div class="header-actions">
        <button class="btn-icon" @click="cancel" title="返回列表">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M6 18L18 6M6 6l12 12"/>
          </svg>
        </button>
      </div>
    </div>

    <div class="wizard-body">
      <!-- 左侧步骤导航 -->
      <div class="wizard-sidebar">
        <div class="steps-container">
          <div
            v-for="(step, index) in steps"
            :key="step.id"
            :class="['step-item', { active: currentStep === index, completed: index < currentStep }]"
            @click="goToStep(index)"
          >
            <div class="step-indicator">
              <span v-if="index < currentStep" class="check-icon">✓</span>
              <span v-else>{{ index + 1 }}</span>
            </div>
            <div class="step-content">
              <div class="step-title">{{ step.title }}</div>
              <div class="step-desc">{{ step.description }}</div>
            </div>
          </div>
        </div>

        <!-- 快速模板选择 -->
        <div class="template-selector">
          <div class="template-label">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/>
              <line x1="3" y1="9" x2="21" y2="9"/>
              <line x1="9" y1="21" x2="9" y2="9"/>
            </svg>
            快速模板
          </div>
          <select v-model="selectedTemplateId" @change="handleTemplateChange" class="template-select">
            <option value="">不使用模板</option>
            <option v-for="template in templates" :key="template.id" :value="template.id">
              {{ template.name }}
            </option>
          </select>
        </div>
      </div>

      <!-- 右侧表单内容 -->
      <div class="wizard-content">
        <form @submit.prevent="submit">
          <!-- Step 1: 基本信息 -->
          <div v-show="currentStep === 0" class="step-panel">
            <div class="panel-header">
              <div class="panel-icon basic">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <circle cx="12" cy="12" r="10"/>
                  <line x1="12" y1="8" x2="12" y2="12"/>
                  <line x1="12" y1="16" x2="12.01" y2="16"/>
                </svg>
              </div>
              <div>
                <h2>基本信息</h2>
                <p>设置流水线的名称和描述信息</p>
              </div>
            </div>

            <div class="form-card">
              <div class="form-group">
                <label class="form-label">
                  流水线名称
                  <span class="required">*</span>
                </label>
                <div class="input-wrapper">
                  <svg class="input-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M12 2L2 7l10 5 10-5-10-5z"/>
                    <path d="M2 17l10 5 10-5"/>
                    <path d="M2 12l10 5 10-5"/>
                  </svg>
                  <input
                    type="text"
                    v-model="pipelineData.name"
                    class="form-input with-icon"
                    placeholder="例如：frontend-deploy-pipeline"
                    required
                  />
                </div>
                <div class="input-hint">使用小写字母、数字和连字符，例如 my-app-pipeline</div>
              </div>

              <div class="form-group">
                <label class="form-label">描述</label>
                <textarea
                  v-model="pipelineData.description"
                  class="form-textarea"
                  placeholder="简要描述此流水线的用途..."
                  rows="3"
                ></textarea>
              </div>
            </div>
          </div>

          <!-- Step 2: 代码仓库配置 -->
          <div v-show="currentStep === 1" class="step-panel">
            <div class="panel-header">
              <div class="panel-icon git">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <circle cx="12" cy="18" r="3"/>
                  <circle cx="6" cy="6" r="3"/>
                  <circle cx="18" cy="6" r="3"/>
                  <path d="M18 9a9 9 0 0 1-9 9"/>
                  <path d="M6 9a9 9 0 0 0 9 9"/>
                </svg>
              </div>
              <div>
                <h2>代码仓库</h2>
                <p>配置 Git 代码仓库和分支信息</p>
              </div>
            </div>

            <div class="form-card">
              <div class="form-group">
                <label class="form-label">
                  Git 仓库地址
                  <span class="required">*</span>
                </label>
                <div class="input-with-action">
                  <div class="input-wrapper flex-1">
                    <svg class="input-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/>
                      <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/>
                    </svg>
                    <input
                      type="url"
                      v-model="pipelineData.git_repo"
                      class="form-input with-icon"
                      placeholder="https://github.com/your-org/your-repo.git"
                      required
                      @blur="onRepoUrlChange"
                    />
                  </div>
                  <button
                    type="button"
                    class="btn-fetch"
                    @click="fetchBranches"
                    :disabled="!pipelineData.git_repo || fetchingBranches"
                    :title="fetchingBranches ? '获取中...' : '获取分支列表'"
                  >
                    <svg v-if="!fetchingBranches" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M23 4v6h-6"/>
                      <path d="M1 20v-6h6"/>
                      <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/>
                    </svg>
                    <span v-else class="loading-spinner-sm"></span>
                    {{ fetchingBranches ? '' : '获取分支' }}
                  </button>
                </div>
              </div>

              <div class="form-group">
                <label class="form-label">
                  分支
                  <span class="required">*</span>
                  <span v-if="branches.length > 0" class="branch-count">（共 {{ branches.length }} 个分支）</span>
                </label>
                
                <!-- 有分支列表时显示下拉选择 -->
                <div v-if="branches.length > 0" class="branch-selector">
                  <div class="branch-search">
                    <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <circle cx="11" cy="11" r="8"/>
                      <line x1="21" y1="21" x2="16.65" y2="16.65"/>
                    </svg>
                    <input
                      type="text"
                      v-model="branchSearch"
                      class="branch-search-input"
                      placeholder="搜索分支..."
                    />
                  </div>
                  <div class="branch-list">
                    <div
                      v-for="branch in filteredBranches"
                      :key="branch.name"
                      :class="['branch-item', { selected: pipelineData.git_branch === branch.name, default: branch.isDefault }]"
                      @click="selectBranch(branch.name)"
                    >
                      <div class="branch-info">
                        <svg class="branch-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                          <line x1="6" y1="3" x2="6" y2="15"/>
                          <circle cx="18" cy="6" r="3"/>
                          <circle cx="6" cy="18" r="3"/>
                          <path d="M18 9a9 9 0 0 1-9 9"/>
                        </svg>
                        <span class="branch-name">{{ branch.name }}</span>
                        <span v-if="branch.isDefault" class="default-badge">default</span>
                      </div>
                      <div v-if="pipelineData.git_branch === branch.name" class="branch-check">
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
                          <polyline points="20 6 9 17 4 12"/>
                        </svg>
                      </div>
                    </div>
                    <div v-if="filteredBranches.length === 0" class="no-branches">
                      没有找到匹配的分支
                    </div>
                  </div>
                </div>

                <!-- 没有分支列表时显示输入框 -->
                <div v-else class="input-wrapper">
                  <svg class="input-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <line x1="6" y1="3" x2="6" y2="15"/>
                    <circle cx="18" cy="6" r="3"/>
                    <circle cx="6" cy="18" r="3"/>
                    <path d="M18 9a9 9 0 0 1-9 9"/>
                  </svg>
                  <input
                    type="text"
                    v-model="pipelineData.git_branch"
                    class="form-input with-icon"
                    placeholder="main"
                    required
                  />
                </div>
                <div class="input-hint">
                  <span v-if="branches.length === 0">输入分支名称，或点击上方"获取分支"按钮自动获取</span>
                  <span v-else>已选择：<strong>{{ pipelineData.git_branch }}</strong></span>
                </div>
              </div>

              <!-- Git 配置提示卡片 -->
              <div class="info-card">
                <div class="info-icon">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="12" cy="12" r="10"/>
                    <line x1="12" y1="16" x2="12" y2="12"/>
                    <line x1="12" y1="8" x2="12.01" y2="8"/>
                  </svg>
                </div>
                <div class="info-content">
                  <div class="info-title">提示</div>
                  <div class="info-text">确保仓库地址可访问，并已在 Jenkins 中配置好相应的凭证</div>
                </div>
              </div>
            </div>
          </div>

          <!-- Step 3: Jenkins 配置 -->
          <div v-show="currentStep === 2" class="step-panel">
            <div class="panel-header">
              <div class="panel-icon jenkins">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <rect x="2" y="3" width="20" height="14" rx="2" ry="2"/>
                  <line x1="8" y1="21" x2="16" y2="21"/>
                  <line x1="12" y1="17" x2="12" y2="21"/>
                </svg>
              </div>
              <div>
                <h2>Jenkins 配置</h2>
                <p>配置 Jenkins 服务器和构建任务</p>
              </div>
            </div>

            <div class="form-card">
              <div class="form-group">
                <label class="form-label">Jenkins 服务器地址</label>
                <div class="input-wrapper">
                  <svg class="input-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <rect x="2" y="2" width="20" height="8" rx="2" ry="2"/>
                    <rect x="2" y="14" width="20" height="8" rx="2" ry="2"/>
                    <line x1="6" y1="6" x2="6.01" y2="6"/>
                    <line x1="6" y1="18" x2="6.01" y2="18"/>
                  </svg>
                  <input
                    type="url"
                    v-model="pipelineData.jenkins_url"
                    class="form-input with-icon"
                    placeholder="http://jenkins.example.com:8080"
                  />
                </div>
                <div class="input-hint">留空则使用系统默认 Jenkins 服务器</div>
              </div>

              <div class="form-group">
                <label class="form-label">
                  Jenkins Job 名称
                  <span class="required">*</span>
                </label>
                <div class="input-wrapper">
                  <svg class="input-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/>
                  </svg>
                  <input
                    type="text"
                    v-model="pipelineData.jenkins_job"
                    class="form-input with-icon"
                    placeholder="my-app-build"
                    required
                  />
                </div>
                <div class="input-hint">Jenkins 中已创建的 Job 名称</div>
              </div>

              <!-- 环境变量配置 -->
              <div class="env-section">
                <div class="section-header" @click="toggleEnvVars">
                  <div class="section-title">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <rect x="1" y="4" width="22" height="16" rx="2" ry="2"/>
                      <line x1="1" y1="10" x2="23" y2="10"/>
                    </svg>
                    环境变量
                    <span class="badge">{{ pipelineData.env_vars.length }}</span>
                  </div>
                  <svg :class="['chevron', { expanded: showEnvVars }]" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="6 9 12 15 18 9"/>
                  </svg>
                </div>

                <div v-show="showEnvVars" class="env-vars-container">
                  <div v-for="(envVar, index) in pipelineData.env_vars" :key="index" class="env-var-row">
                    <input
                      type="text"
                      v-model="envVar.name"
                      class="form-input env-name"
                      placeholder="变量名"
                    />
                    <span class="env-separator">=</span>
                    <input
                      type="text"
                      v-model="envVar.value"
                      class="form-input env-value"
                      placeholder="变量值"
                    />
                    <button type="button" class="btn-icon-sm danger" @click="removeEnvVar(index)">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <line x1="18" y1="6" x2="6" y2="18"/>
                        <line x1="6" y1="6" x2="18" y2="18"/>
                      </svg>
                    </button>
                  </div>

                  <button type="button" class="btn-add-env" @click="addEnvVar">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <line x1="12" y1="5" x2="12" y2="19"/>
                      <line x1="5" y1="12" x2="19" y2="12"/>
                    </svg>
                    添加环境变量
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- Step 4: 部署策略 -->
          <div v-show="currentStep === 3" class="step-panel">
            <div class="panel-header">
              <div class="panel-icon deploy">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/>
                  <polyline points="22,6 12,13 2,6"/>
                </svg>
              </div>
              <div>
                <h2>部署策略</h2>
                <p>配置滚动更新参数</p>
              </div>
            </div>

            <div class="form-card">
              <!-- 部署策略卡片 -->
              <div class="strategy-cards">
                <div
                  v-for="strategy in deployStrategies"
                  :key="strategy.value"
                  :class="['strategy-card', { selected: pipelineData.deploy_config.strategy === strategy.value }]"
                  @click="pipelineData.deploy_config.strategy = strategy.value"
                >
                  <div class="strategy-icon" v-html="strategy.icon"></div>
                  <div class="strategy-name">{{ strategy.name }}</div>
                  <div class="strategy-desc">{{ strategy.description }}</div>
                  <div v-if="pipelineData.deploy_config.strategy === strategy.value" class="strategy-check">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
                      <polyline points="20 6 9 17 4 12"/>
                    </svg>
                  </div>
                </div>
              </div>

              <!-- 副本数配置 -->
              <div class="form-group">
                <label class="form-label">副本数</label>
                <div class="replica-control">
                  <button type="button" class="replica-btn" @click="decreaseReplicas" :disabled="pipelineData.deploy_config.replicas <= 1">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <line x1="5" y1="12" x2="19" y2="12"/>
                    </svg>
                  </button>
                  <input
                    type="number"
                    v-model.number="pipelineData.deploy_config.replicas"
                    class="replica-input"
                    min="1"
                    max="100"
                  />
                  <button type="button" class="replica-btn" @click="increaseReplicas">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <line x1="12" y1="5" x2="12" y2="19"/>
                      <line x1="5" y1="12" x2="19" y2="12"/>
                    </svg>
                  </button>
                </div>
              </div>

              <!-- 资源配置 -->
              <div class="resources-section">
                <div class="section-header" @click="toggleResources">
                  <div class="section-title">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <rect x="2" y="3" width="20" height="14" rx="2" ry="2"/>
                      <line x1="8" y1="21" x2="16" y2="21"/>
                      <line x1="12" y1="17" x2="12" y2="21"/>
                    </svg>
                    资源配置
                  </div>
                  <svg :class="['chevron', { expanded: showResources }]" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="6 9 12 15 18 9"/>
                  </svg>
                </div>

                <div v-show="showResources" class="resources-container">
                  <div class="resource-group">
                    <div class="resource-label">
                      <span class="resource-type limits">Limits</span>
                      资源上限
                    </div>
                    <div class="resource-inputs">
                      <div class="resource-input-group">
                        <label>CPU</label>
                        <input
                          type="text"
                          v-model="pipelineData.deploy_config.resources.limits.cpu"
                          class="form-input"
                          placeholder="500m"
                        />
                      </div>
                      <div class="resource-input-group">
                        <label>内存</label>
                        <input
                          type="text"
                          v-model="pipelineData.deploy_config.resources.limits.memory"
                          class="form-input"
                          placeholder="512Mi"
                        />
                      </div>
                    </div>
                  </div>

                  <div class="resource-group">
                    <div class="resource-label">
                      <span class="resource-type requests">Requests</span>
                      资源请求
                    </div>
                    <div class="resource-inputs">
                      <div class="resource-input-group">
                        <label>CPU</label>
                        <input
                          type="text"
                          v-model="pipelineData.deploy_config.resources.requests.cpu"
                          class="form-input"
                          placeholder="200m"
                        />
                      </div>
                      <div class="resource-input-group">
                        <label>内存</label>
                        <input
                          type="text"
                          v-model="pipelineData.deploy_config.resources.requests.memory"
                          class="form-input"
                          placeholder="256Mi"
                        />
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Step 5: 自动部署配置 -->
          <div v-show="currentStep === 4" class="step-panel">
            <div class="panel-header">
              <div class="panel-icon auto-deploy">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <circle cx="12" cy="12" r="10"/>
                  <path d="M12 6v6l4 2"/>
                </svg>
              </div>
              <div>
                <h2>自动部署配置</h2>
                <p>构建成功后自动部署到 Kubernetes 集群</p>
              </div>
            </div>

            <div class="form-card">
              <!-- 自动部署开关 -->
              <div class="form-group">
                <div class="toggle-row">
                  <div class="toggle-info">
                    <label class="form-label">启用自动部署</label>
                    <p class="toggle-desc">构建成功后自动更新 K8s 工作负载的镜像</p>
                  </div>
                  <label class="toggle-switch">
                    <input type="checkbox" v-model="pipelineData.auto_deploy" @change="onAutoDeployChange" />
                    <span class="toggle-slider"></span>
                  </label>
                </div>
              </div>

              <!-- 自动部署配置详情 -->
              <div v-if="pipelineData.auto_deploy" class="auto-deploy-config">
                <!-- 部署环境选择 -->
                <div class="form-group">
                  <label class="form-label">
                    部署环境
                    <span class="required">*</span>
                  </label>
                  <div class="env-selector">
                    <div
                      v-for="env in deployEnvOptions"
                      :key="env.value"
                      :class="['env-card', { selected: pipelineData.deploy_env === env.value }]"
                      @click="selectDeployEnv(env.value)"
                    >
                      <div class="env-indicator" :style="{ backgroundColor: env.color }"></div>
                      <div class="env-info">
                        <span class="env-name">{{ env.label }}</span>
                        <span class="env-desc">{{ env.description }}</span>
                      </div>
                      <div v-if="pipelineData.deploy_env === env.value" class="env-check">
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
                          <polyline points="20 6 9 17 4 12"/>
                        </svg>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- 部署审批开关（所有环境可配置） -->
                <div class="form-group">
                  <div :class="['toggle-row', { warning: pipelineData.deploy_env === 'prod' }]">
                    <div class="toggle-info">
                      <label class="form-label">
                        部署前审批
                        <span v-if="pipelineData.deploy_env === 'prod'" class="env-tag prod">生产环境建议开启</span>
                      </label>
                      <p class="toggle-desc">部署前需要人工审批确认，防止误操作</p>
                    </div>
                    <label class="toggle-switch">
                      <input type="checkbox" v-model="pipelineData.require_approval" />
                      <span :class="['toggle-slider', { warning: pipelineData.deploy_env === 'prod' }]"></span>
                    </label>
                  </div>
                </div>

                <!-- 目标集群选择 -->
                <div class="form-group">
                  <label class="form-label">
                    目标集群
                    <span class="required">*</span>
                  </label>
                  <div class="select-wrapper">
                    <select 
                      v-model="pipelineData.target_cluster_id" 
                      class="form-select"
                      @change="onClusterChange"
                      :disabled="loadingClusters"
                    >
                      <option :value="0">请选择集群</option>
                      <option v-for="cluster in clusters" :key="cluster.id" :value="cluster.id">
                        {{ cluster.cluster_name }} (ID: {{ cluster.id }})
                      </option>
                    </select>
                    <button type="button" class="btn-refresh" @click="loadClusters" :disabled="loadingClusters">
                      <svg v-if="!loadingClusters" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <polyline points="23 4 23 10 17 10"/>
                        <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
                      </svg>
                      <span v-else class="loading-spinner-sm"></span>
                    </button>
                  </div>
                </div>

                <!-- 目标命名空间 -->
                <div class="form-group">
                  <label class="form-label">
                    目标命名空间
                    <span class="required">*</span>
                  </label>
                  <div class="select-wrapper">
                    <select 
                      v-model="pipelineData.target_namespace" 
                      class="form-select"
                      @change="onNamespaceChange"
                      :disabled="loadingNamespaces || !pipelineData.target_cluster_id"
                    >
                      <option value="">请选择命名空间</option>
                      <option v-for="ns in namespaces" :key="ns.name" :value="ns.name">
                        {{ ns.name }}
                      </option>
                    </select>
                    <button 
                      type="button" 
                      class="btn-refresh" 
                      @click="loadNamespaces" 
                      :disabled="loadingNamespaces || !pipelineData.target_cluster_id"
                    >
                      <svg v-if="!loadingNamespaces" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <polyline points="23 4 23 10 17 10"/>
                        <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
                      </svg>
                      <span v-else class="loading-spinner-sm"></span>
                    </button>
                  </div>
                </div>

                <!-- 工作负载类型 -->
                <div class="form-group">
                  <label class="form-label">
                    工作负载类型
                    <span class="required">*</span>
                  </label>
                  <div class="workload-kind-selector">
                    <div
                      v-for="kind in workloadKindOptions"
                      :key="kind.value"
                      :class="['kind-card', { selected: pipelineData.target_workload_kind === kind.value }]"
                      @click="selectWorkloadKind(kind.value)"
                    >
                      <span class="kind-name">{{ kind.label }}</span>
                      <span class="kind-desc">{{ kind.description }}</span>
                    </div>
                  </div>
                </div>

                <!-- 工作负载名称 -->
                <div class="form-group">
                  <label class="form-label">
                    工作负载名称
                    <span class="required">*</span>
                  </label>
                  <div class="select-wrapper">
                    <select 
                      v-if="workloads.length > 0"
                      v-model="pipelineData.target_workload_name" 
                      class="form-select"
                      @change="onWorkloadChange"
                    >
                      <option value="">请选择工作负载</option>
                      <option v-for="w in workloads" :key="w.name" :value="w.name">
                        {{ w.name }}
                      </option>
                    </select>
                    <input 
                      v-else
                      type="text" 
                      v-model="pipelineData.target_workload_name" 
                      class="form-input"
                      placeholder="输入工作负载名称"
                    />
                    <button 
                      type="button" 
                      class="btn-refresh" 
                      @click="loadWorkloads" 
                      :disabled="loadingWorkloads || !pipelineData.target_namespace"
                    >
                      <svg v-if="!loadingWorkloads" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <polyline points="23 4 23 10 17 10"/>
                        <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
                      </svg>
                      <span v-else class="loading-spinner-sm"></span>
                    </button>
                  </div>
                  <div class="input-hint">将更新该工作负载的容器镜像</div>
                </div>

                <!-- 容器名称 -->
                <div class="form-group">
                  <label class="form-label">容器名称</label>
                  <input 
                    type="text" 
                    v-model="pipelineData.target_container" 
                    class="form-input"
                    placeholder="输入要更新的容器名称，留空则更新第一个容器"
                  />
                  <div class="input-hint">指定要更新镜像的容器，留空则更新第一个容器</div>
                </div>

                <!-- 配置摘要 -->
                <div class="config-summary">
                  <div class="summary-title">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <circle cx="12" cy="12" r="10"/>
                      <line x1="12" y1="16" x2="12" y2="12"/>
                      <line x1="12" y1="8" x2="12.01" y2="8"/>
                    </svg>
                    配置摘要
                  </div>
                  <div class="summary-content">
                    <div class="summary-item">
                      <span class="summary-label">部署目标:</span>
                      <span class="summary-value">
                        {{ getClusterName(pipelineData.target_cluster_id) }} / 
                        {{ pipelineData.target_namespace || '-' }} / 
                        {{ pipelineData.target_workload_kind }} / 
                        {{ pipelineData.target_workload_name || '-' }}
                      </span>
                    </div>
                    <div class="summary-item">
                      <span class="summary-label">部署环境:</span>
                      <span class="summary-value" :class="`env-${pipelineData.deploy_env}`">
                        {{ getEnvLabel(pipelineData.deploy_env) }}
                        <span v-if="pipelineData.require_approval" class="approval-badge">需审批</span>
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div class="wizard-footer">
            <button
              type="button"
              class="btn btn-secondary"
              @click="previousStep"
              :disabled="currentStep === 0"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="19" y1="12" x2="5" y2="12"/>
                <polyline points="12 19 5 12 12 5"/>
              </svg>
              上一步
            </button>

            <div class="footer-info">
              步骤 {{ currentStep + 1 }} / {{ steps.length }}
            </div>

            <button
              v-if="currentStep < steps.length - 1"
              type="button"
              class="btn btn-primary"
              @click="nextStep"
            >
              下一步
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="5" y1="12" x2="19" y2="12"/>
                <polyline points="12 5 19 12 12 19"/>
              </svg>
            </button>

            <button
              v-else
              type="submit"
              class="btn btn-success"
              :disabled="submitting"
            >
              <svg v-if="!submitting" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="20 6 9 17 4 12"/>
              </svg>
              <span v-if="submitting" class="loading-spinner"></span>
              {{ submitting ? '提交中...' : (isEdit ? '保存修改' : '创建流水线') }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  createPipeline,
  updatePipeline,
  getPipelineDetail,
  getPipelineTemplates,
  getGitBranches
} from '@/api/cicd.js'
import { getClusterList } from '@/api/cluster.js'
import namespaceApi from '@/api/cluster/config/namespace'
import deploymentsApi from '@/api/cluster/workloads/deployments'
import statefulsetsApi from '@/api/cluster/workloads/statefulsets'
import daemonsetsApi from '@/api/cluster/workloads/daemonsets'
import { useClusterStore } from '@/stores/cluster'

export default {
  name: 'PipelineCreate',
  setup() {
    const router = useRouter()
    const route = useRoute()
    const pipelineId = route.params.id
    const isEdit = !!pipelineId

    // 步骤定义
    const steps = ref([
      { id: 'basic', title: '基本信息', description: '名称和描述' },
      { id: 'git', title: '代码仓库', description: 'Git 配置' },
      { id: 'jenkins', title: 'Jenkins', description: '构建配置' },
      { id: 'deploy', title: '部署策略', description: '滚动更新参数' },
      { id: 'auto-deploy', title: '自动部署', description: 'K8s 目标配置' }
    ])

    const currentStep = ref(0)
    const templates = ref([])
    const selectedTemplateId = ref('')
    const showEnvVars = ref(true)
    const showResources = ref(true)

    // Git 分支相关
    const branches = ref([])
    const branchSearch = ref('')
    const fetchingBranches = ref(false)
    const lastFetchedRepo = ref('')

    // 过滤后的分支列表
    const filteredBranches = computed(() => {
      if (!branchSearch.value.trim()) {
        return branches.value
      }
      const keyword = branchSearch.value.toLowerCase()
      return branches.value.filter(b => 
        b.name.toLowerCase().includes(keyword)
      )
    })

    // 部署策略选项
    const deployStrategies = ref([
      {
        value: 'rollingUpdate',
        name: '滚动更新',
        description: '逐步替换旧版本',
        icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M23 4v6h-6"/><path d="M1 20v-6h6"/><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/></svg>'
      },
      {
        value: 'recreate',
        name: '重新创建',
        description: '停止后再启动',
        icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="1 4 1 10 7 10"/><path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/></svg>'
      },
      {
        value: 'blueGreen',
        name: '蓝绿部署',
        description: '零停机切换',
        icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="2" width="9" height="9" rx="2"/><rect x="13" y="13" width="9" height="9" rx="2"/><path d="M9 13l6-6"/></svg>'
      },
      {
        value: 'canary',
        name: '金丝雀',
        description: '渐进式发布',
        icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 20V10"/><path d="M12 20V4"/><path d="M6 20v-6"/></svg>'
      }
    ])

    // 集群列表
    const clusters = ref([])
    const loadingClusters = ref(false)
    
    // 命名空间列表
    const namespaces = ref([])
    const loadingNamespaces = ref(false)
    
    // 工作负载列表
    const workloads = ref([])
    const loadingWorkloads = ref(false)
    
    // 部署环境选项
    const deployEnvOptions = ref([
      { value: 'dev', label: '开发环境', color: '#52c41a', description: '用于开发调试' },
      { value: 'test', label: '测试环境', color: '#1890ff', description: '用于单元测试' },
      { value: 'staging', label: '预发环境', color: '#faad14', description: '用于集成测试' },
      { value: 'prod', label: '生产环境', color: '#ff4d4f', description: '需要审批流程' }
    ])
    
    // 工作负载类型选项
    const workloadKindOptions = ref([
      { value: 'Deployment', label: 'Deployment', description: '无状态应用' },
      { value: 'StatefulSet', label: 'StatefulSet', description: '有状态应用' },
      { value: 'DaemonSet', label: 'DaemonSet', description: '守护进程' }
    ])

    // 表单数据
    const pipelineData = ref({
      name: '',
      description: '',
      git_repo: '',
      git_branch: 'main',
      jenkins_url: '',
      jenkins_job: '',
      env_vars: [],
      deploy_config: {
        replicas: 3,
        strategy: 'rollingUpdate',
        resources: {
          limits: { cpu: '500m', memory: '512Mi' },
          requests: { cpu: '200m', memory: '256Mi' }
        }
      },
      // 自动部署配置
      auto_deploy: false,
      target_cluster_id: 0,
      target_namespace: '',
      target_workload_kind: 'Deployment',
      target_workload_name: '',
      target_container: '',
      deploy_env: 'dev',
      require_approval: true
    })

    const submitting = ref(false)

    // 步骤导航
    const nextStep = () => {
      if (validateCurrentStep()) {
        currentStep.value = Math.min(currentStep.value + 1, steps.value.length - 1)
      }
    }

    const previousStep = () => {
      currentStep.value = Math.max(currentStep.value - 1, 0)
    }

    const goToStep = (index) => {
      if (index <= currentStep.value || validateCurrentStep()) {
        currentStep.value = index
      }
    }

    // 验证当前步骤
    const validateCurrentStep = () => {
      switch (currentStep.value) {
        case 0:
          if (!pipelineData.value.name.trim()) {
            alert('请输入流水线名称')
            return false
          }
          break
        case 1:
          if (!pipelineData.value.git_repo.trim()) {
            alert('请输入 Git 仓库地址')
            return false
          }
          if (!pipelineData.value.git_branch.trim()) {
            alert('请输入分支名称')
            return false
          }
          break
        case 2:
          if (!pipelineData.value.jenkins_job.trim()) {
            alert('请输入 Jenkins Job 名称')
            return false
          }
          break
      }
      return true
    }

    // 环境变量操作
    const toggleEnvVars = () => {
      showEnvVars.value = !showEnvVars.value
    }

    const addEnvVar = () => {
      pipelineData.value.env_vars.push({ name: '', value: '' })
    }

    const removeEnvVar = (index) => {
      pipelineData.value.env_vars.splice(index, 1)
    }

    // 资源配置
    const toggleResources = () => {
      showResources.value = !showResources.value
    }

    // 副本数控制
    const increaseReplicas = () => {
      pipelineData.value.deploy_config.replicas++
    }

    const decreaseReplicas = () => {
      if (pipelineData.value.deploy_config.replicas > 1) {
        pipelineData.value.deploy_config.replicas--
      }
    }

    // Git 分支获取
    const fetchBranches = async () => {
      const repoUrl = pipelineData.value.git_repo.trim()
      if (!repoUrl) {
        alert('请先输入 Git 仓库地址')
        return
      }

      // 如果已经获取过相同仓库的分支，不重复获取
      if (lastFetchedRepo.value === repoUrl && branches.value.length > 0) {
        return
      }

      fetchingBranches.value = true
      branchSearch.value = ''

      try {
        const response = await getGitBranches(repoUrl)
        
        if (response.code === 0 && response.data) {
          // 后端返回的分支列表格式: [{ name: 'main', isDefault: true }, ...]
          branches.value = response.data.branches || response.data || []
          lastFetchedRepo.value = repoUrl
          
          // 如果当前没有选择分支，自动选择默认分支
          if (!pipelineData.value.git_branch || !branches.value.find(b => b.name === pipelineData.value.git_branch)) {
            const defaultBranch = branches.value.find(b => b.isDefault)
            if (defaultBranch) {
              pipelineData.value.git_branch = defaultBranch.name
            } else if (branches.value.length > 0) {
              // 优先选择 main 或 master
              const mainBranch = branches.value.find(b => b.name === 'main' || b.name === 'master')
              pipelineData.value.git_branch = mainBranch ? mainBranch.name : branches.value[0].name
            }
          }
        } else {
          // 后端接口未实现或返回错误，使用模拟数据
          console.warn('获取分支失败，使用默认分支列表')
          branches.value = generateMockBranches(repoUrl)
          lastFetchedRepo.value = repoUrl
        }
      } catch (error) {
        console.error('获取分支失败:', error)
        // 接口调用失败时，使用模拟分支数据
        branches.value = generateMockBranches(repoUrl)
        lastFetchedRepo.value = repoUrl
      } finally {
        fetchingBranches.value = false
      }
    }

    // 生成模拟分支数据（后端接口未实现时使用）
    const generateMockBranches = (repoUrl) => {
      // 根据仓库类型生成常见分支
      const commonBranches = [
        { name: 'main', isDefault: true },
        { name: 'master', isDefault: false },
        { name: 'develop', isDefault: false },
        { name: 'release', isDefault: false },
        { name: 'feature/new-feature', isDefault: false },
        { name: 'hotfix/bug-fix', isDefault: false }
      ]
      return commonBranches
    }

    // 选择分支
    const selectBranch = (branchName) => {
      pipelineData.value.git_branch = branchName
    }

    // 仓库地址变化时清空分支列表
    const onRepoUrlChange = () => {
      if (pipelineData.value.git_repo !== lastFetchedRepo.value) {
        branches.value = []
        branchSearch.value = ''
      }
    }

    // 模板处理
    const loadTemplates = async () => {
      try {
        const response = await getPipelineTemplates()
        if (response.code === 0) {
          templates.value = response.data || []
        }
      } catch (error) {
        console.error('获取模板失败:', error)
      }
    }

    const handleTemplateChange = () => {
      if (selectedTemplateId.value) {
        const template = templates.value.find(t => t.id === parseInt(selectedTemplateId.value))
        if (template) {
          pipelineData.value.env_vars = JSON.parse(JSON.stringify(template.defaultEnvVars || []))
          pipelineData.value.deploy_config = JSON.parse(JSON.stringify(template.defaultDeploymentConfig || pipelineData.value.deploy_config))
        }
      }
    }

    // 加载编辑数据
    const loadPipelineData = async () => {
      if (isEdit) {
        try {
          const response = await getPipelineDetail(pipelineId)
          if (response.code === 0) {
            const data = response.data?.pipeline || response.data
            pipelineData.value = {
              name: data.name || '',
              description: data.description || '',
              git_repo: data.git_repo || '',
              git_branch: data.git_branch || 'main',
              jenkins_url: data.jenkins_url || '',
              jenkins_job: data.jenkins_job || '',
              env_vars: data.env_vars || [],
              deploy_config: data.deploy_config || pipelineData.value.deploy_config,
              // 自动部署配置
              auto_deploy: data.auto_deploy || false,
              target_cluster_id: data.target_cluster_id || 0,
              target_namespace: data.target_namespace || '',
              target_workload_kind: data.target_workload_kind || 'Deployment',
              target_workload_name: data.target_workload_name || '',
              target_container: data.target_container || '',
              deploy_env: data.deploy_env || 'dev',
              require_approval: data.require_approval || false
            }
            // 如果有自动部署配置，加载相关数据
            if (data.auto_deploy && data.target_cluster_id) {
              await loadClusters()
              await loadNamespaces()
              await loadWorkloads()
            }
          }
        } catch (error) {
          alert('获取流水线详情失败')
        }
      }
    }

    // 提交表单
    const submit = async () => {
      if (!validateCurrentStep()) return

      try {
        submitting.value = true
        let response

        if (isEdit) {
          response = await updatePipeline({
            id: parseInt(pipelineId),
            ...pipelineData.value
          })
        } else {
          response = await createPipeline(pipelineData.value)
        }

        if (response.code === 0) {
          alert(isEdit ? '更新流水线成功' : '创建流水线成功')
          router.push('/cicd/pipelines')
        } else {
          alert(response.msg || '操作失败')
        }
      } catch (error) {
        console.error('提交失败:', error)
        alert(error.msg || (isEdit ? '更新流水线失败' : '创建流水线失败'))
      } finally {
        submitting.value = false
      }
    }

    const cancel = () => {
      router.push('/cicd/pipelines')
    }

    // ==================== 自动部署相关方法 ====================
    const clusterStore = useClusterStore()
    
    // 加载集群列表
    const loadClusters = async () => {
      loadingClusters.value = true
      try {
        const res = await getClusterList({ page: 1, limit: 100 })
        if (res.code === 0 && res.data) {
          clusters.value = res.data.list || []
        }
      } catch (error) {
        console.error('加载集群失败:', error)
      } finally {
        loadingClusters.value = false
      }
    }
    
    // 加载命名空间列表
    const loadNamespaces = async () => {
      if (!pipelineData.value.target_cluster_id) {
        namespaces.value = []
        return
      }
      
      loadingNamespaces.value = true
      try {
        // 设置当前集群
        const cluster = clusters.value.find(c => c.id === pipelineData.value.target_cluster_id)
        if (cluster) {
          clusterStore.setCurrent(cluster)
        }
        
        const res = await namespaceApi.list({ page: 1, limit: 1000 })
        if (res.code === 0 && res.data) {
          namespaces.value = res.data.list || res.data || []
        }
      } catch (error) {
        console.error('加载命名空间失败:', error)
        // 回退到常用命名空间
        namespaces.value = [
          { name: 'default' },
          { name: 'kube-system' },
          { name: 'kube-public' }
        ]
      } finally {
        loadingNamespaces.value = false
      }
    }
    
    // 加载工作负载列表
    const loadWorkloads = async () => {
      if (!pipelineData.value.target_namespace) {
        workloads.value = []
        return
      }
      
      loadingWorkloads.value = true
      try {
        let res
        const kind = pipelineData.value.target_workload_kind
        const ns = pipelineData.value.target_namespace
        
        switch (kind) {
          case 'StatefulSet':
            res = await statefulsetsApi.list({ namespace: ns, page: 1, limit: 1000 })
            break
          case 'DaemonSet':
            res = await daemonsetsApi.list({ namespace: ns, page: 1, limit: 1000 })
            break
          default:
            res = await deploymentsApi.list({ namespace: ns, page: 1, limit: 1000 })
        }
        
        if (res.code === 0 && res.data) {
          workloads.value = res.data.list || res.data || []
        }
      } catch (error) {
        console.error('加载工作负载失败:', error)
        workloads.value = []
      } finally {
        loadingWorkloads.value = false
      }
    }
    
    // 自动部署开关变化
    const onAutoDeployChange = () => {
      if (pipelineData.value.auto_deploy && clusters.value.length === 0) {
        loadClusters()
      }
    }
    
    // 集群变化
    const onClusterChange = () => {
      pipelineData.value.target_namespace = ''
      pipelineData.value.target_workload_name = ''
      namespaces.value = []
      workloads.value = []
      if (pipelineData.value.target_cluster_id) {
        loadNamespaces()
      }
    }
    
    // 命名空间变化
    const onNamespaceChange = () => {
      pipelineData.value.target_workload_name = ''
      workloads.value = []
      if (pipelineData.value.target_namespace) {
        loadWorkloads()
      }
    }
    
    // 工作负载变化
    const onWorkloadChange = () => {
      // 可以在这里自动填充容器名称
    }
    
    // 选择部署环境
    const selectDeployEnv = (env) => {
      pipelineData.value.deploy_env = env
      // 生产环境默认需要审批
      if (env === 'prod') {
        pipelineData.value.require_approval = true
      }
    }
    
    // 选择工作负载类型
    const selectWorkloadKind = (kind) => {
      pipelineData.value.target_workload_kind = kind
      pipelineData.value.target_workload_name = ''
      workloads.value = []
      if (pipelineData.value.target_namespace) {
        loadWorkloads()
      }
    }
    
    // 获取集群名称
    const getClusterName = (clusterId) => {
      if (!clusterId) return '-'
      const cluster = clusters.value.find(c => c.id === clusterId)
      return cluster ? cluster.cluster_name : '-'
    }
    
    // 获取环境标签
    const getEnvLabel = (env) => {
      const option = deployEnvOptions.value.find(o => o.value === env)
      return option ? option.label : env
    }

    onMounted(() => {
      loadTemplates()
      loadPipelineData()
    })

    return {
      isEdit,
      steps,
      currentStep,
      templates,
      selectedTemplateId,
      pipelineData,
      submitting,
      showEnvVars,
      showResources,
      deployStrategies,
      // Git 分支相关
      branches,
      branchSearch,
      fetchingBranches,
      filteredBranches,
      fetchBranches,
      selectBranch,
      onRepoUrlChange,
      // 自动部署相关
      clusters,
      loadingClusters,
      namespaces,
      loadingNamespaces,
      workloads,
      loadingWorkloads,
      deployEnvOptions,
      workloadKindOptions,
      loadClusters,
      loadNamespaces,
      loadWorkloads,
      onAutoDeployChange,
      onClusterChange,
      onNamespaceChange,
      onWorkloadChange,
      selectDeployEnv,
      selectWorkloadKind,
      getClusterName,
      getEnvLabel,
      // 方法
      nextStep,
      previousStep,
      goToStep,
      toggleEnvVars,
      addEnvVar,
      removeEnvVar,
      toggleResources,
      increaseReplicas,
      decreaseReplicas,
      handleTemplateChange,
      submit,
      cancel
    }
  }
}
</script>

<style scoped>
/* ==================== 整体布局 ==================== */
.pipeline-wizard {
  min-height: 100vh;
  background: linear-gradient(135deg, #f5f7fa 0%, #e4e8ec 100%);
}

/* ==================== 顶部标题栏 ==================== */
.wizard-header {
  background: linear-gradient(135deg, #1e3a5f 0%, #2c5282 100%);
  color: white;
  padding: 20px 32px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
}

.header-content {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-icon {
  width: 48px;
  height: 48px;
  background: rgba(255, 255, 255, 0.15);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.header-icon svg {
  width: 28px;
  height: 28px;
}

.header-text h1 {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
}

.header-text p {
  margin: 4px 0 0;
  opacity: 0.8;
  font-size: 14px;
}

.btn-icon {
  width: 40px;
  height: 40px;
  border: none;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  color: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s;
}

.btn-icon:hover {
  background: rgba(255, 255, 255, 0.2);
}

.btn-icon svg {
  width: 20px;
  height: 20px;
}

/* ==================== 主体布局 ==================== */
.wizard-body {
  display: flex;
  max-width: 1400px;
  margin: 0 auto;
  padding: 24px;
  gap: 24px;
}

/* ==================== 左侧步骤导航 ==================== */
.wizard-sidebar {
  width: 280px;
  flex-shrink: 0;
}

.steps-container {
  background: white;
  border-radius: 16px;
  padding: 24px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
}

.step-item {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  padding: 16px;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.3s;
  margin-bottom: 8px;
}

.step-item:last-child {
  margin-bottom: 0;
}

.step-item:hover {
  background: #f7fafc;
}

.step-item.active {
  background: linear-gradient(135deg, #ebf4ff 0%, #e6fffa 100%);
}

.step-item.completed .step-indicator {
  background: linear-gradient(135deg, #48bb78 0%, #38a169 100%);
  color: white;
}

.step-indicator {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: #e2e8f0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 14px;
  color: #718096;
  flex-shrink: 0;
  transition: all 0.3s;
}

.step-item.active .step-indicator {
  background: linear-gradient(135deg, #4299e1 0%, #3182ce 100%);
  color: white;
  box-shadow: 0 4px 12px rgba(66, 153, 225, 0.4);
}

.check-icon {
  font-size: 16px;
}

.step-content {
  flex: 1;
}

.step-title {
  font-weight: 600;
  color: #2d3748;
  font-size: 14px;
  margin-bottom: 4px;
}

.step-desc {
  font-size: 12px;
  color: #a0aec0;
}

/* 模板选择器 */
.template-selector {
  background: white;
  border-radius: 16px;
  padding: 20px;
  margin-top: 16px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
}

.template-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 600;
  color: #4a5568;
  margin-bottom: 12px;
}

.template-label svg {
  width: 16px;
  height: 16px;
}

.template-select {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 13px;
  color: #4a5568;
  background: white;
  cursor: pointer;
  transition: all 0.3s;
}

.template-select:focus {
  outline: none;
  border-color: #4299e1;
  box-shadow: 0 0 0 3px rgba(66, 153, 225, 0.15);
}

/* ==================== 右侧表单内容 ==================== */
.wizard-content {
  flex: 1;
  min-width: 0;
}

.step-panel {
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.panel-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 24px;
}

.panel-icon {
  width: 56px;
  height: 56px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.panel-icon svg {
  width: 28px;
  height: 28px;
  color: white;
}

.panel-icon.basic {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.panel-icon.git {
  background: linear-gradient(135deg, #f6ad55 0%, #ed8936 100%);
}

.panel-icon.jenkins {
  background: linear-gradient(135deg, #fc8181 0%, #f56565 100%);
}

.panel-icon.deploy {
  background: linear-gradient(135deg, #4fd1c5 0%, #38b2ac 100%);
}

.panel-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #1a202c;
}

.panel-header p {
  margin: 4px 0 0;
  font-size: 14px;
  color: #718096;
}

/* ==================== 表单卡片 ==================== */
.form-card {
  background: white;
  border-radius: 16px;
  padding: 28px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
}

.form-group {
  margin-bottom: 24px;
}

.form-group:last-child {
  margin-bottom: 0;
}

.form-label {
  display: block;
  font-size: 14px;
  font-weight: 600;
  color: #2d3748;
  margin-bottom: 8px;
}

.required {
  color: #e53e3e;
}

.input-wrapper {
  position: relative;
}

.input-icon {
  position: absolute;
  left: 14px;
  top: 50%;
  transform: translateY(-50%);
  width: 18px;
  height: 18px;
  color: #a0aec0;
}

.form-input {
  width: 100%;
  padding: 12px 16px;
  border: 2px solid #e2e8f0;
  border-radius: 10px;
  font-size: 14px;
  color: #2d3748;
  transition: all 0.3s;
  background: #f7fafc;
}

.form-input.with-icon {
  padding-left: 44px;
}

.form-input:focus {
  outline: none;
  border-color: #4299e1;
  background: white;
  box-shadow: 0 0 0 4px rgba(66, 153, 225, 0.1);
}

.form-input::placeholder {
  color: #a0aec0;
}

.form-textarea {
  width: 100%;
  padding: 12px 16px;
  border: 2px solid #e2e8f0;
  border-radius: 10px;
  font-size: 14px;
  color: #2d3748;
  transition: all 0.3s;
  background: #f7fafc;
  resize: vertical;
  min-height: 100px;
}

.form-textarea:focus {
  outline: none;
  border-color: #4299e1;
  background: white;
  box-shadow: 0 0 0 4px rgba(66, 153, 225, 0.1);
}

.input-hint {
  font-size: 12px;
  color: #a0aec0;
  margin-top: 6px;
}

.input-hint strong {
  color: #4299e1;
  font-weight: 600;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

/* ==================== Git 仓库和分支选择器 ==================== */
.input-with-action {
  display: flex;
  gap: 12px;
  align-items: flex-start;
}

.input-with-action .flex-1 {
  flex: 1;
}

.btn-fetch {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 12px 16px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  border-radius: 10px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
  white-space: nowrap;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.btn-fetch:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(102, 126, 234, 0.4);
}

.btn-fetch:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none;
}

.btn-fetch svg {
  width: 16px;
  height: 16px;
}

.loading-spinner-sm {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.branch-count {
  font-size: 12px;
  font-weight: 400;
  color: #718096;
  margin-left: 4px;
}

/* 分支选择器 */
.branch-selector {
  border: 2px solid #e2e8f0;
  border-radius: 12px;
  overflow: hidden;
  background: #f7fafc;
}

.branch-search {
  position: relative;
  padding: 12px;
  background: white;
  border-bottom: 1px solid #e2e8f0;
}

.search-icon {
  position: absolute;
  left: 24px;
  top: 50%;
  transform: translateY(-50%);
  width: 16px;
  height: 16px;
  color: #a0aec0;
}

.branch-search-input {
  width: 100%;
  padding: 10px 12px 10px 36px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 13px;
  color: #2d3748;
  background: #f7fafc;
  transition: all 0.3s;
}

.branch-search-input:focus {
  outline: none;
  border-color: #4299e1;
  background: white;
}

.branch-list {
  max-height: 280px;
  overflow-y: auto;
}

.branch-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  cursor: pointer;
  transition: all 0.2s;
  border-bottom: 1px solid #edf2f7;
}

.branch-item:last-child {
  border-bottom: none;
}

.branch-item:hover {
  background: #edf2f7;
}

.branch-item.selected {
  background: linear-gradient(135deg, #ebf8ff 0%, #e6fffa 100%);
}

.branch-item.default {
  font-weight: 500;
}

.branch-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.branch-icon {
  width: 18px;
  height: 18px;
  color: #718096;
}

.branch-item.selected .branch-icon {
  color: #4299e1;
}

.branch-name {
  font-size: 14px;
  color: #2d3748;
}

.branch-item.selected .branch-name {
  color: #2b6cb0;
  font-weight: 600;
}

.default-badge {
  padding: 2px 8px;
  background: #c6f6d5;
  color: #276749;
  font-size: 10px;
  font-weight: 700;
  border-radius: 10px;
  text-transform: uppercase;
}

.branch-check {
  width: 22px;
  height: 22px;
  background: #4299e1;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.branch-check svg {
  width: 12px;
  height: 12px;
}

.no-branches {
  padding: 24px;
  text-align: center;
  color: #a0aec0;
  font-size: 13px;
  margin-top: 6px;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

/* ==================== 信息卡片 ==================== */
.info-card {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 16px;
  background: linear-gradient(135deg, #ebf8ff 0%, #e6fffa 100%);
  border-radius: 12px;
  border-left: 4px solid #4299e1;
  margin-top: 20px;
}

.info-icon {
  width: 24px;
  height: 24px;
  color: #4299e1;
  flex-shrink: 0;
}

.info-icon svg {
  width: 100%;
  height: 100%;
}

.info-title {
  font-weight: 600;
  color: #2b6cb0;
  font-size: 13px;
  margin-bottom: 4px;
}

.info-text {
  font-size: 13px;
  color: #4a5568;
  line-height: 1.5;
}

/* ==================== 环境变量配置 ==================== */
.env-section, .resources-section {
  margin-top: 24px;
  border-top: 1px solid #e2e8f0;
  padding-top: 24px;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  cursor: pointer;
  padding: 12px 16px;
  background: #f7fafc;
  border-radius: 10px;
  transition: all 0.3s;
}

.section-header:hover {
  background: #edf2f7;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: 600;
  color: #2d3748;
  font-size: 14px;
}

.section-title svg {
  width: 18px;
  height: 18px;
  color: #718096;
}

.badge {
  background: #4299e1;
  color: white;
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 11px;
  font-weight: 600;
}

.chevron {
  width: 20px;
  height: 20px;
  color: #718096;
  transition: transform 0.3s;
}

.chevron.expanded {
  transform: rotate(180deg);
}

.env-vars-container, .resources-container {
  margin-top: 16px;
  animation: slideDown 0.3s ease;
}

@keyframes slideDown {
  from { opacity: 0; transform: translateY(-10px); }
  to { opacity: 1; transform: translateY(0); }
}

.env-var-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.env-name {
  width: 180px;
  flex-shrink: 0;
}

.env-separator {
  color: #a0aec0;
  font-weight: 600;
  font-size: 16px;
}

.env-value {
  flex: 1;
}

.btn-icon-sm {
  width: 36px;
  height: 36px;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s;
  flex-shrink: 0;
}

.btn-icon-sm svg {
  width: 16px;
  height: 16px;
}

.btn-icon-sm.danger {
  background: #fff5f5;
  color: #e53e3e;
}

.btn-icon-sm.danger:hover {
  background: #fed7d7;
}

.btn-add-env {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  width: 100%;
  padding: 12px;
  border: 2px dashed #cbd5e0;
  border-radius: 10px;
  background: transparent;
  color: #718096;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s;
}

.btn-add-env:hover {
  border-color: #4299e1;
  color: #4299e1;
  background: #ebf8ff;
}

.btn-add-env svg {
  width: 18px;
  height: 18px;
}

/* ==================== 部署策略卡片 ==================== */
.strategy-cards {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  margin-bottom: 28px;
}

.strategy-card {
  position: relative;
  padding: 20px;
  border: 2px solid #e2e8f0;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.3s;
  background: white;
}

.strategy-card:hover {
  border-color: #4299e1;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(66, 153, 225, 0.15);
}

.strategy-card.selected {
  border-color: #4299e1;
  background: linear-gradient(135deg, #ebf8ff 0%, #e6fffa 100%);
}

.strategy-icon {
  width: 40px;
  height: 40px;
  margin-bottom: 12px;
  color: #4299e1;
}

.strategy-icon :deep(svg) {
  width: 100%;
  height: 100%;
}

.strategy-name {
  font-weight: 600;
  color: #2d3748;
  font-size: 14px;
  margin-bottom: 4px;
}

.strategy-desc {
  font-size: 12px;
  color: #718096;
}

.strategy-check {
  position: absolute;
  top: 12px;
  right: 12px;
  width: 24px;
  height: 24px;
  background: #4299e1;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.strategy-check svg {
  width: 14px;
  height: 14px;
}

/* ==================== 副本数控制 ==================== */
.replica-control {
  display: flex;
  align-items: center;
  gap: 12px;
}

.replica-btn {
  width: 44px;
  height: 44px;
  border: 2px solid #e2e8f0;
  border-radius: 10px;
  background: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s;
}

.replica-btn:hover:not(:disabled) {
  border-color: #4299e1;
  color: #4299e1;
}

.replica-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.replica-btn svg {
  width: 20px;
  height: 20px;
}

.replica-input {
  width: 80px;
  text-align: center;
  font-size: 20px;
  font-weight: 600;
  padding: 10px;
  border: 2px solid #e2e8f0;
  border-radius: 10px;
  background: #f7fafc;
}

.replica-input:focus {
  outline: none;
  border-color: #4299e1;
  background: white;
}

/* ==================== 资源配置 ==================== */
.resource-group {
  padding: 16px;
  background: #f7fafc;
  border-radius: 10px;
  margin-bottom: 12px;
}

.resource-group:last-child {
  margin-bottom: 0;
}

.resource-label {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
  font-weight: 600;
  color: #4a5568;
  margin-bottom: 12px;
}

.resource-type {
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
}

.resource-type.limits {
  background: #fed7d7;
  color: #c53030;
}

.resource-type.requests {
  background: #c6f6d5;
  color: #276749;
}

.resource-inputs {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.resource-input-group label {
  display: block;
  font-size: 12px;
  color: #718096;
  margin-bottom: 6px;
}

/* ==================== 底部操作栏 ==================== */
.wizard-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 24px;
  padding: 20px 28px;
  background: white;
  border-radius: 16px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
}

.footer-info {
  font-size: 14px;
  color: #718096;
  font-weight: 500;
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  border: none;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
}

.btn svg {
  width: 18px;
  height: 18px;
}

.btn-secondary {
  background: #edf2f7;
  color: #4a5568;
}

.btn-secondary:hover:not(:disabled) {
  background: #e2e8f0;
}

.btn-secondary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-primary {
  background: linear-gradient(135deg, #4299e1 0%, #3182ce 100%);
  color: white;
  box-shadow: 0 4px 12px rgba(66, 153, 225, 0.4);
}

.btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(66, 153, 225, 0.5);
}

.btn-success {
  background: linear-gradient(135deg, #48bb78 0%, #38a169 100%);
  color: white;
  box-shadow: 0 4px 12px rgba(72, 187, 120, 0.4);
}

.btn-success:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(72, 187, 120, 0.5);
}

.btn-success:disabled {
  opacity: 0.7;
  cursor: not-allowed;
  transform: none;
}

.loading-spinner {
  width: 18px;
  height: 18px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* ==================== 响应式适配 ==================== */
@media (max-width: 1024px) {
  .wizard-body {
    flex-direction: column;
  }
  
  .wizard-sidebar {
    width: 100%;
  }
  
  .steps-container {
    display: flex;
    overflow-x: auto;
    gap: 8px;
    padding: 16px;
  }
  
  .step-item {
    flex-shrink: 0;
    margin-bottom: 0;
  }
  
  .strategy-cards {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .wizard-header {
    padding: 16px;
  }
  
  .header-icon {
    display: none;
  }
  
  .wizard-body {
    padding: 16px;
  }
  
  .form-row {
    grid-template-columns: 1fr;
  }
  
  .resource-inputs {
    grid-template-columns: 1fr;
  }
}

/* ==================== 自动部署配置样式 ==================== */
.panel-icon.auto-deploy {
  background: linear-gradient(135deg, #805ad5 0%, #6b46c1 100%);
}

.toggle-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  background: #f7fafc;
  border-radius: 12px;
  border: 2px solid #e2e8f0;
}

.toggle-row.warning {
  background: #fffbeb;
  border-color: #fcd34d;
}

.toggle-info {
  flex: 1;
}

.toggle-info .form-label {
  margin-bottom: 4px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.env-tag {
  display: inline-block;
  padding: 2px 8px;
  font-size: 11px;
  font-weight: 500;
  border-radius: 4px;
}

.env-tag.prod {
  background: #fed7d7;
  color: #c53030;
}

.toggle-desc {
  font-size: 13px;
  color: #718096;
  margin: 0;
}

.toggle-switch {
  position: relative;
  width: 52px;
  height: 28px;
  cursor: pointer;
}

.toggle-switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.toggle-slider {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: #cbd5e0;
  border-radius: 28px;
  transition: all 0.3s;
}

.toggle-slider::before {
  position: absolute;
  content: "";
  height: 22px;
  width: 22px;
  left: 3px;
  bottom: 3px;
  background: white;
  border-radius: 50%;
  transition: all 0.3s;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.toggle-switch input:checked + .toggle-slider {
  background: linear-gradient(135deg, #48bb78 0%, #38a169 100%);
}

.toggle-switch input:checked + .toggle-slider.warning {
  background: linear-gradient(135deg, #f6ad55 0%, #ed8936 100%);
}

.toggle-switch input:checked + .toggle-slider::before {
  transform: translateX(24px);
}

.auto-deploy-config {
  margin-top: 24px;
  padding-top: 24px;
  border-top: 2px dashed #e2e8f0;
}

/* 部署环境选择器 */
.env-selector {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.env-card {
  position: relative;
  padding: 16px;
  border: 2px solid #e2e8f0;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.3s;
  display: flex;
  align-items: center;
  gap: 12px;
}

.env-card:hover {
  border-color: #a0aec0;
  transform: translateY(-2px);
}

.env-card.selected {
  border-color: #4299e1;
  background: #ebf8ff;
}

.env-indicator {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  flex-shrink: 0;
}

.env-info {
  flex: 1;
}

.env-name {
  display: block;
  font-weight: 600;
  color: #2d3748;
  font-size: 14px;
}

.env-desc {
  display: block;
  font-size: 12px;
  color: #718096;
  margin-top: 2px;
}

.env-check {
  width: 20px;
  height: 20px;
  background: #4299e1;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.env-check svg {
  width: 12px;
  height: 12px;
}

/* 工作负载类型选择器 */
.workload-kind-selector {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.kind-card {
  padding: 14px 16px;
  border: 2px solid #e2e8f0;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.3s;
  text-align: center;
}

.kind-card:hover {
  border-color: #a0aec0;
}

.kind-card.selected {
  border-color: #4299e1;
  background: #ebf8ff;
}

.kind-name {
  display: block;
  font-weight: 600;
  color: #2d3748;
  font-size: 14px;
}

.kind-desc {
  display: block;
  font-size: 12px;
  color: #718096;
  margin-top: 2px;
}

/* 下拉选择器包装 */
.select-wrapper {
  display: flex;
  gap: 8px;
}

.form-select {
  flex: 1;
  padding: 12px 16px;
  border: 2px solid #e2e8f0;
  border-radius: 10px;
  font-size: 14px;
  color: #2d3748;
  background: #f7fafc;
  cursor: pointer;
  transition: all 0.3s;
}

.form-select:focus {
  outline: none;
  border-color: #4299e1;
  background: white;
}

.form-select:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-refresh {
  width: 44px;
  height: 44px;
  border: 2px solid #e2e8f0;
  border-radius: 10px;
  background: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s;
  flex-shrink: 0;
}

.btn-refresh:hover:not(:disabled) {
  border-color: #4299e1;
  color: #4299e1;
}

.btn-refresh:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-refresh svg {
  width: 18px;
  height: 18px;
}

.loading-spinner-sm {
  width: 16px;
  height: 16px;
  border: 2px solid #e2e8f0;
  border-top-color: #4299e1;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

/* 配置摘要 */
.config-summary {
  margin-top: 24px;
  padding: 16px 20px;
  background: linear-gradient(135deg, #f0f9ff 0%, #e0f2fe 100%);
  border-radius: 12px;
  border: 1px solid #bae6fd;
}

.summary-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: #0369a1;
  font-size: 14px;
  margin-bottom: 12px;
}

.summary-title svg {
  width: 18px;
  height: 18px;
}

.summary-content {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.summary-item {
  display: flex;
  gap: 8px;
  font-size: 13px;
}

.summary-label {
  color: #64748b;
  flex-shrink: 0;
}

.summary-value {
  color: #1e293b;
  font-weight: 500;
  word-break: break-all;
}

.summary-value.env-dev {
  color: #16a34a;
}

.summary-value.env-staging {
  color: #d97706;
}

.summary-value.env-prod {
  color: #dc2626;
}

.approval-badge {
  display: inline-block;
  padding: 2px 8px;
  background: #fef3c7;
  color: #92400e;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  margin-left: 8px;
}

@media (max-width: 768px) {
  .env-selector,
  .workload-kind-selector {
    grid-template-columns: 1fr;
  }
}
</style>
