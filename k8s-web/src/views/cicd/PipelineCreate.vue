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
        <button
          v-if="isEdit"
          type="button"
          class="btn-header-save"
          @click="submit"
          :disabled="submitting"
        >
          <svg v-if="!submitting" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="20 6 9 17 4 12"/>
          </svg>
          <span v-if="submitting" class="loading-spinner-sm"></span>
          {{ submitting ? '保存中...' : '保存修改' }}
        </button>
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
                  <span v-if="pipelineData.language_type === 'custom'" class="required">*</span>
                </label>
                <div class="input-wrapper">
                  <svg class="input-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/>
                  </svg>
                  <input
                    type="text"
                    v-model="pipelineData.jenkins_job"
                    class="form-input with-icon"
                    :placeholder="pipelineData.language_type === 'custom' ? 'my-app-build（必填）' : `k8s-builder-${pipelineData.language_type}（留空自动推导）`"
                    :required="pipelineData.language_type === 'custom'"
                  />
                </div>
                <div class="input-hint">
                  <template v-if="pipelineData.language_type !== 'custom'">
                    <template v-if="pipelineData.jenkins_job && pipelineData.jenkins_job.trim()">
                      <span style="color:#52c41a;">&#10004;</span> 使用自定义 Job <strong>{{ pipelineData.jenkins_job }}</strong>，平台会自动注入 <code style="color:#1890ff;">LANGUAGE_TYPE={{ pipelineData.language_type }}</code> 参数，Jenkinsfile 分发器会自动加载对应模板。
                    </template>
                    <template v-else>
                      留空将自动使用平台内置 Job: <strong>k8s-builder-{{ pipelineData.language_type }}</strong>（推荐）
                    </template>
                  </template>
                  <template v-else>自定义类型必须填写 Jenkins Job 名称（Jenkins 上已创建的 Job）</template>
                </div>
              </div>

              <!-- SonarQube 代码质量扫描开关 -->
              <div class="form-group">
                <div :class="['toggle-row', { highlight: pipelineData.language_type === 'java' }]">
                  <div class="toggle-info">
                    <label class="form-label">
                      SonarQube 代码质量扫描
                      <span v-if="pipelineData.language_type === 'java'" class="env-tag" style="background:#52c41a;color:#fff;font-size:11px;padding:1px 6px;border-radius:3px;margin-left:6px;">Java 推荐</span>
                    </label>
                    <p class="toggle-desc">启用后构建时自动进行代码质量扫描和质量门禁检查</p>
                  </div>
                  <label class="toggle-switch">
                    <input type="checkbox" v-model="pipelineData.enable_sonar" />
                    <span class="toggle-slider"></span>
                  </label>
                </div>
              </div>

              <!-- 制品上传到平台制品库开关 -->
              <div class="form-group">
                <div class="toggle-row">
                  <div class="toggle-info">
                    <label class="form-label">
                      制品上传到平台制品库
                      <span class="env-tag" style="background:#1890ff;color:#fff;font-size:11px;padding:1px 6px;border-radius:3px;margin-left:6px;">推荐</span>
                    </label>
                    <p class="toggle-desc">启用后构建完成自动将制品（JAR/二进制/dist）上传到平台制品库，支持版本溯源与下载</p>
                  </div>
                  <label class="toggle-switch">
                    <input type="checkbox" v-model="pipelineData.enable_artifact_upload" />
                    <span class="toggle-slider"></span>
                  </label>
                </div>
              </div>

              <!-- ==================== 构建核心参数 ==================== -->
              <div class="build-params-section">
                <div class="section-divider">
                  <span class="divider-text">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:14px;height:14px;vertical-align:middle;margin-right:4px;">
                      <path d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4"/>
                    </svg>
                    构建参数
                  </span>
                </div>

                <!-- 镜像仓库地址（必填） -->
                <div class="form-group">
                  <label class="form-label">
                    镜像仓库地址
                    <span class="required">*</span>
                  </label>
                  <div class="input-wrapper">
                    <svg class="input-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <rect x="2" y="2" width="20" height="8" rx="2" ry="2"/>
                      <rect x="2" y="14" width="20" height="8" rx="2" ry="2"/>
                      <line x1="6" y1="6" x2="6.01" y2="6"/>
                      <line x1="6" y1="18" x2="6.01" y2="18"/>
                    </svg>
                    <input
                      type="text"
                      v-model="pipelineData.image_repo"
                      class="form-input with-icon"
                      placeholder="harbor.example.com/project/app-name"
                      required
                    />
                  </div>
                  <div class="input-hint">Jenkins 构建后将镜像推送到此地址，格式：registry/project/app</div>
                </div>

                <!-- Dockerfile 构建策略 -->
                <div class="form-group">
                  <label class="form-label">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:15px;height:15px;vertical-align:middle;margin-right:4px;">
                      <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                      <polyline points="14 2 14 8 20 8"/>
                    </svg>
                    Dockerfile 策略
                  </label>
                  <div class="dockerfile-mode-selector">
                    <div
                      :class="['df-mode-card', { active: dockerfileMode === 'auto' }]"
                      @click="dockerfileMode = 'auto'"
                    >
                      <div class="df-mode-icon auto">
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                          <circle cx="11" cy="11" r="8"/>
                          <line x1="21" y1="21" x2="16.65" y2="16.65"/>
                        </svg>
                      </div>
                      <div class="df-mode-info">
                        <div class="df-mode-title">智能检测<span class="df-badge recommend">推荐</span></div>
                        <div class="df-mode-desc">自动检测项目 Dockerfile，未找到则平台生成</div>
                      </div>
                      <div class="df-mode-check" v-if="dockerfileMode === 'auto'">&#10003;</div>
                    </div>

                    <div
                      :class="['df-mode-card', { active: dockerfileMode === 'project' }]"
                      @click="dockerfileMode = 'project'"
                    >
                      <div class="df-mode-icon project">
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                          <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
                        </svg>
                      </div>
                      <div class="df-mode-info">
                        <div class="df-mode-title">项目自带</div>
                        <div class="df-mode-desc">使用仓库中已定义的 Dockerfile</div>
                      </div>
                      <div class="df-mode-check" v-if="dockerfileMode === 'project'">&#10003;</div>
                    </div>

                    <div
                      :class="['df-mode-card', { active: dockerfileMode === 'platform' }]"
                      @click="dockerfileMode = 'platform'"
                    >
                      <div class="df-mode-icon platform">
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                          <rect x="2" y="3" width="20" height="14" rx="2" ry="2"/>
                          <line x1="8" y1="21" x2="16" y2="21"/>
                          <line x1="12" y1="17" x2="12" y2="21"/>
                        </svg>
                      </div>
                      <div class="df-mode-info">
                        <div class="df-mode-title">平台生成</div>
                        <div class="df-mode-desc">忽略项目文件，由平台生成最优 Dockerfile</div>
                      </div>
                      <div class="df-mode-check" v-if="dockerfileMode === 'platform'">&#10003;</div>
                    </div>
                  </div>

                  <!-- 项目自带模式：路径输入 -->
                  <div v-if="dockerfileMode === 'project'" class="df-path-input">
                    <div class="input-wrapper">
                      <svg class="input-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                        <polyline points="14 2 14 8 20 8"/>
                      </svg>
                      <input
                        type="text"
                        v-model="pipelineData.dockerfile_path"
                        class="form-input with-icon"
                        placeholder="Dockerfile"
                      />
                    </div>
                    <div class="input-hint">相对于项目根目录的路径，例如 Dockerfile、docker/Dockerfile.prod</div>
                  </div>

                  <!-- 策略说明面板 -->
                  <div class="df-info-panel">
                    <div v-if="dockerfileMode === 'auto'" class="df-info-content">
                      <div class="df-info-title">&#9889; 智能检测流程</div>
                      <div class="df-info-steps">
                        <div class="df-step"><span class="df-step-num">1</span>检查项目根目录是否存在 Dockerfile</div>
                        <div class="df-step"><span class="df-step-num">2</span>存在则直接使用项目 Dockerfile 构建镜像</div>
                        <div class="df-step"><span class="df-step-num">3</span>不存在则根据 {{ dockerfileLangLabel }} 自动生成纯运行时 Dockerfile</div>
                      </div>
                    </div>
                    <div v-else-if="dockerfileMode === 'project'" class="df-info-content">
                      <div class="df-info-title">&#128193; 项目 Dockerfile 说明</div>
                      <div class="df-info-text">
                        直接使用项目仓库中的 Dockerfile，适合已自定义好构建逻辑的项目。
                        <br/>Jenkins 编译产物（如 <code>target/*.jar</code>、<code>bin/</code>）在同一工作目录，Dockerfile 可直接 COPY。
                      </div>
                    </div>
                    <div v-else class="df-info-content">
                      <div class="df-info-title">&#129302; 平台生成说明</div>
                      <div class="df-info-text">
                        平台根据 <strong>{{ dockerfileLangLabel }}</strong> 语言类型，自动生成生产级纯运行时 Dockerfile：
                        <br/>&#8226; 阿里云镜像源加速 &#8226; 非 root 用户 &#8226; 最小化镜像层 &#8226; 生产 JVM / Runtime 参数
                      </div>
                    </div>
                  </div>
                </div>

                <!-- Git 凭证 ID + 跳过测试 -->
                <div class="form-row">
                  <div class="form-group half">
                    <label class="form-label">Git 凭证 ID</label>
                    <input
                      type="text"
                      v-model="pipelineData.git_credential_id"
                      class="form-input"
                      placeholder="gitee-id"
                    />
                    <div class="input-hint">Jenkins 中配置的 Git 凭证 ID</div>
                  </div>
                  <div class="form-group half">
                    <label class="form-label">跳过单元测试</label>
                    <div class="toggle-row compact">
                      <span class="toggle-desc">构建时跳过测试阶段</span>
                      <label class="toggle-switch">
                        <input type="checkbox" v-model="pipelineData.skip_tests" />
                        <span class="toggle-slider"></span>
                      </label>
                    </div>
                  </div>
                </div>
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
              <!-- 部署环境选择（在Step 4直接选择） -->
              <div class="form-group">
                <label class="form-label">部署环境</label>
                <div class="env-selector-inline">
                  <div
                    v-for="env in deployEnvOptions"
                    :key="env.value"
                    :class="['env-chip', { selected: pipelineData.deploy_env === env.value }]"
                    @click="selectDeployEnv(env.value)"
                  >
                    <span class="env-dot" :style="{ backgroundColor: env.color }"></span>
                    <span>{{ env.label }}</span>
                  </div>
                </div>
              </div>

              <!-- 服务类型和资源模板选择 -->
              <div class="resource-template-section">
                <div class="form-row">
                  <div class="form-group half">
                    <label class="form-label">服务类型</label>
                    <div class="service-type-selector">
                      <div
                        v-for="svc in serviceTypeOptions"
                        :key="svc.value"
                        :class="['service-type-card', { selected: selectedServiceType === svc.value }]"
                        @click="onServiceTypeChange(svc.value)"
                      >
                        <span class="svc-dot" :style="{ backgroundColor: svc.color }"></span>
                        <span class="svc-name">{{ svc.label }}</span>
                      </div>
                    </div>
                  </div>
                  <div class="form-group half">
                    <label class="form-label">资源模板</label>
                    <select v-model="selectedResourceTemplate" @change="onResourceTemplateChange" class="form-select" :disabled="loadingResourceTemplates">
                      <option value="">自定义配置</option>
                      <option v-for="tpl in resourceTemplates" :key="tpl.id" :value="tpl.id">
                        {{ tpl.name }} - {{ tpl.description || tpl.name }}
                      </option>
                    </select>
                  </div>
                </div>
                <!-- 服务类型一致性提示 -->
                <div class="input-hint" style="margin-top:6px;padding:6px 10px;background:#fffbe6;border:1px solid #ffe58f;border-radius:4px;">
                  <span style="color:#faad14;">&#9888;</span>
                  <strong>注意：</strong>此处「服务类型」决定资源模板和部署参数，请确保与上方「语言/框架类型」(<strong>{{ pipelineData.language_type }}</strong>) 保持一致。
                  <template v-if="selectedServiceType !== pipelineData.language_type && pipelineData.language_type !== 'custom'">
                    <br/><span style="color:#ff4d4f;">&#10060; 当前不一致：语言类型为 <strong>{{ pipelineData.language_type }}</strong>，服务类型为 <strong>{{ selectedServiceType }}</strong>，可能导致资源模板不匹配！</span>
                  </template>
                </div>
              </div>

              <!-- 资源校验提示 -->
              <div v-if="resourceValidation" :class="['validation-result', resourceValidation.valid ? 'success' : 'error', resourceValidation.risk_level]">
                <div class="validation-header">
                  <svg v-if="resourceValidation.valid" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
                    <polyline points="22 4 12 14.01 9 11.01"/>
                  </svg>
                  <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="12" cy="12" r="10"/>
                    <line x1="12" y1="8" x2="12" y2="12"/>
                    <line x1="12" y1="16" x2="12.01" y2="16"/>
                  </svg>
                  <span>{{ resourceValidation.valid ? '配置校验通过' : '配置校验失败' }}</span>
                  <span v-if="resourceValidation.risk_level === 'high'" class="risk-badge high">高风险</span>
                  <span v-else-if="resourceValidation.risk_level === 'medium'" class="risk-badge medium">中风险</span>
                </div>
                <ul v-if="resourceValidation.errors && resourceValidation.errors.length" class="validation-errors">
                  <li v-for="(err, i) in resourceValidation.errors" :key="i">{{ err }}</li>
                </ul>
                <ul v-if="resourceValidation.warnings && resourceValidation.warnings.length" class="validation-warnings">
                  <li v-for="(warn, i) in resourceValidation.warnings" :key="i">{{ warn }}</li>
                </ul>
                
                <!-- 审批提示区域（大厂风格） -->
                <div v-if="resourceValidation.need_approval" class="approval-card">
                  <div class="approval-card-header">
                    <div class="approval-icon">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"/>
                      </svg>
                    </div>
                    <div class="approval-info">
                      <div class="approval-title">生产环境审批</div>
                      <div class="approval-desc">此配置需要 <strong>{{ resourceValidation.approval_role?.toUpperCase() || 'SRE' }}</strong> 角色审批后方可部署</div>
                    </div>
                  </div>
                  
                  <!-- 有审批权限：显示操作按钮 -->
                  <div v-if="canApprove" class="approval-actions">
                    <div class="approval-status approved">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
                        <polyline points="22 4 12 14.01 9 11.01"/>
                      </svg>
                      <span>你拥有审批权限，可直接部署</span>
                    </div>
                  </div>
                  
                  <!-- 无审批权限：显示等待审批提示 -->
                  <div v-else class="approval-pending">
                    <div class="pending-info">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <circle cx="12" cy="12" r="10"/>
                        <polyline points="12 6 12 12 16 14"/>
                      </svg>
                      <span>提交后将进入审批流程，请等待审批人处理</span>
                    </div>
                  </div>
                </div>
                
                <div v-if="resourceValidation.suggestion" class="validation-suggestion">
                  <strong>建议：</strong>{{ resourceValidation.suggestion }}
                </div>
              </div>

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
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  createPipeline,
  updatePipeline,
  getPipelineDetail,
  getPipelineTemplates,
  getGitBranches,
  getResourceTemplates,
  validateResourceConfig
} from '@/api/cicd.js'
import { getClusterList } from '@/api/cluster.js'
import namespaceApi from '@/api/cluster/config/namespace'
import deploymentsApi from '@/api/cluster/workloads/deployments'
import statefulsetsApi from '@/api/cluster/workloads/statefulsets'
import daemonsetsApi from '@/api/cluster/workloads/daemonsets'
import { useClusterStore } from '@/stores/cluster'
import permissionStore from '@/stores/permission'

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
    
    // 资源模板相关
    const resourceTemplates = ref([])
    const selectedResourceTemplate = ref('')
    const loadingResourceTemplates = ref(false)
    const resourceValidation = ref(null)
    const validatingResource = ref(false)
    
    // 审批权限判断（大厂风格：无权限则不显示审批按钮）
    const canApprove = computed(() => {
      // 超级管理员有所有权限
      if (permissionStore.state.isSuperAdmin) return true
      
      // 检查角色：platform_admin / cluster_admin / sre 可以审批
      const roleTypes = permissionStore.roleTypes?.value || []
      const approvalRoles = ['super_admin', 'platform_admin', 'cluster_admin', 'sre']
      return approvalRoles.some(role => roleTypes.includes(role))
    })
    
    // 服务类型选项（value 必须与后端 language_type 一致: go/java/frontend/python/custom）
    const serviceTypeOptions = ref([
      { value: 'java', label: 'Java', color: '#f89820' },
      { value: 'go', label: 'Go', color: '#00add8' },
      { value: 'frontend', label: 'Node.js', color: '#339933' },
      { value: 'python', label: 'Python', color: '#3776ab' },
      { value: 'custom', label: '自定义', color: '#8c8c8c' }
    ])
    const selectedServiceType = ref('go')
    
    // Dockerfile 构建策略模式：'auto' | 'project' | 'platform'
    const dockerfileMode = ref('auto')
    
    // 语言类型显示名称（用于 Dockerfile 策略面板）
    const dockerfileLangLabel = computed(() => {
      const langMap = { java: 'Java', go: 'Go', frontend: 'Node.js', python: 'Python', custom: '自定义' }
      return langMap[selectedServiceType.value] || selectedServiceType.value
    })

    // 语言类型 → Jenkins 模板文件名映射（前端提示用）
    const templateFileMap = {
      java: 'java-spring-pipeline.groovy',
      go: 'go-pipeline.groovy',
      frontend: 'frontend-pipeline.groovy',
      python: 'python-pipeline.groovy'
    }
    
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

    // ==================== 语言类型 → 推荐环境变量默认值映射 ====================
    const languageEnvDefaults = {
      java: [
        { name: 'IMAGE_REPO', value: 'harbor.example.com/project/app-name', _hint: '镜像仓库地址（必填）' },
        { name: 'JAVA_VERSION', value: '17', _hint: 'Java 版本' },
        { name: 'MAVEN_GOALS', value: 'clean package -DskipTests -B', _hint: 'Maven 构建命令' },
        { name: 'GIT_CREDENTIAL_ID', value: 'gitee-id', _hint: 'Git 凭证 ID' },
      ],
      go: [
        { name: 'IMAGE_REPO', value: 'harbor.example.com/project/app-name', _hint: '镜像仓库地址（必填）' },
        { name: 'GO_VERSION', value: '1.24', _hint: 'Go 版本' },
        { name: 'GIT_CREDENTIAL_ID', value: 'gitee-id', _hint: 'Git 凭证 ID' },
      ],
      frontend: [
        { name: 'IMAGE_REPO', value: 'harbor.example.com/project/app-name', _hint: '镜像仓库地址（必填）' },
        { name: 'NODE_VERSION', value: '18', _hint: 'Node.js 版本' },
        { name: 'BUILD_COMMAND', value: 'npm run build', _hint: '构建命令' },
        { name: 'BUILD_OUTPUT_DIR', value: 'dist', _hint: '构建产物目录' },
        { name: 'GIT_CREDENTIAL_ID', value: 'gitee-id', _hint: 'Git 凭证 ID' },
      ],
      python: [
        { name: 'IMAGE_REPO', value: 'harbor.example.com/project/app-name', _hint: '镜像仓库地址（必填）' },
        { name: 'PYTHON_VERSION', value: '3.11', _hint: 'Python 版本' },
        { name: 'GIT_CREDENTIAL_ID', value: 'gitee-id', _hint: 'Git 凭证 ID' },
      ],
      custom: [
        { name: 'IMAGE_REPO', value: 'harbor.example.com/project/app-name', _hint: '镜像仓库地址（必填）' },
        { name: 'GIT_CREDENTIAL_ID', value: 'gitee-id', _hint: 'Git 凭证 ID' },
      ]
    }

    // 表单数据
    const pipelineData = ref({
      name: '',
      description: '',
      git_repo: '',
      git_branch: 'main',
      jenkins_url: '',
      jenkins_job: '',
      language_type: 'go',  // 与 selectedServiceType 联动，后端据此自动推导 jenkins_job
      // 构建核心参数（独立字段，不混入 env_vars）
      image_repo: '',       // 镜像仓库地址（必填），如 harbor.example.com/project/app
      skip_tests: false,    // 跳过单元测试
      dockerfile_path: '',  // Dockerfile 路径（空则自动生成）
      git_credential_id: 'gitee-id',  // Git 凭证 ID
      env_vars: [],
      enable_sonar: false,  // SonarQube 代码质量扫描开关（Java 项目默认启用）
      enable_artifact_upload: false,  // 制品上传到平台制品库开关
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
          // jenkins_job 仅在 language_type 为 custom 时必填，其他语言类型由后端自动推导
          if (pipelineData.value.language_type === 'custom' && !pipelineData.value.jenkins_job.trim()) {
            alert('自定义类型必须填写 Jenkins Job 名称')
            return false
          }
          if (!pipelineData.value.image_repo.trim()) {
            alert('请填写镜像仓库地址（IMAGE_REPO）')
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

    // ==================== 资源模板相关方法 ====================
    
    // 加载资源模板
    const loadResourceTemplates = async () => {
      loadingResourceTemplates.value = true
      try {
        const res = await getResourceTemplates({
          env: pipelineData.value.deploy_env || 'dev',
          service_type: selectedServiceType.value
        })
        if (res.code === 0 && res.data) {
          // 后端返回 { list: [...], total: x }
          resourceTemplates.value = res.data.list || res.data || []
          // 找到默认模板
          const defaultTpl = resourceTemplates.value.find(t => t.is_default)
          if (defaultTpl && !selectedResourceTemplate.value) {
            selectedResourceTemplate.value = defaultTpl.id
            applyResourceTemplate(defaultTpl)
          }
        }
      } catch (error) {
        console.error('加载资源模板失败:', error)
      } finally {
        loadingResourceTemplates.value = false
      }
    }
    
    // 服务类型变化 — 同步更新 language_type 并联动 SonarQube 开关 + 自动填充推荐环境变量
    const onServiceTypeChange = (type) => {
      selectedServiceType.value = type
      pipelineData.value.language_type = type
      // Java 项目默认启用 SonarQube 代码质量扫描
      pipelineData.value.enable_sonar = (type === 'java')
      selectedResourceTemplate.value = ''
      loadResourceTemplates()

      // 自动填充语言类型对应的推荐环境变量（保留用户已自定义的变量）
      const defaults = languageEnvDefaults[type] || languageEnvDefaults.custom
      // 收集已知的默认 key——确保不会重复添加
      const allDefaultKeys = new Set()
      Object.values(languageEnvDefaults).forEach(arr => arr.forEach(d => allDefaultKeys.add(d.name)))
      // 保留用户自定义的（不在任何默认列表中的）
      const userCustom = pipelineData.value.env_vars.filter(e => !allDefaultKeys.has(e.name))
      // 将 IMAGE_REPO 提取到独立字段（如果之前在 env_vars 里）
      const existingImageRepo = pipelineData.value.env_vars.find(e => e.name === 'IMAGE_REPO')
      if (existingImageRepo && existingImageRepo.value && existingImageRepo.value !== 'harbor.example.com/project/app-name') {
        pipelineData.value.image_repo = existingImageRepo.value
      }
      // 重组 env_vars：默认推荐变量（排除 IMAGE_REPO 等已有独立字段的） + 用户自定义
      const promotedKeys = ['IMAGE_REPO', 'GIT_CREDENTIAL_ID', 'SKIP_TESTS', 'DOCKERFILE_PATH']
      const newEnvVars = defaults
        .filter(d => !promotedKeys.includes(d.name))
        .map(d => ({ name: d.name, value: d.value }))
      pipelineData.value.env_vars = [...newEnvVars, ...userCustom]
    }
    
    // 资源模板变化
    const onResourceTemplateChange = () => {
      if (selectedResourceTemplate.value) {
        const tpl = resourceTemplates.value.find(t => t.id === parseInt(selectedResourceTemplate.value))
        if (tpl) {
          applyResourceTemplate(tpl)
        }
      }
      doValidateResource()
    }
    
    // 应用资源模板配置
    const applyResourceTemplate = (tpl) => {
      pipelineData.value.deploy_config.replicas = tpl.replicas_default || 1
      pipelineData.value.deploy_config.resources = {
        limits: {
          cpu: tpl.cpu_limit || '500m',
          memory: tpl.memory_limit || '512Mi'
        },
        requests: {
          cpu: tpl.cpu_request || '200m',
          memory: tpl.memory_request || '256Mi'
        }
      }
    }
    
    // 校验资源配置
    const doValidateResource = async () => {
      validatingResource.value = true
      try {
        const res = await validateResourceConfig({
          env: pipelineData.value.deploy_env || 'dev',
          service_type: selectedServiceType.value,
          config: {
            replicas: pipelineData.value.deploy_config.replicas,
            strategy: pipelineData.value.deploy_config.strategy,
            resources: pipelineData.value.deploy_config.resources
          }
        })
        if (res.code === 0 && res.data) {
          resourceValidation.value = res.data
        }
      } catch (error) {
        console.error('资源校验失败:', error)
        resourceValidation.value = null
      } finally {
        validatingResource.value = false
      }
    }

    // 加载编辑数据
    const loadPipelineData = async () => {
      if (isEdit) {
        try {
          const response = await getPipelineDetail(pipelineId)
          if (response.code === 0) {
            const data = response.data?.pipeline || response.data
            // 回显语言类型和 SonarQube 开关
            const langType = data.language_type || 'custom'
            const hasSonar = (data.env_vars || []).some(e => e.name === 'ENABLE_SONAR' && e.value === 'true')
            // 从 env_vars 提取独立字段
            const envArr = data.env_vars || []
            const getEnv = (key, def) => {
              const found = envArr.find(e => e.name === key)
              return found ? found.value : def
            }
            const promotedKeys = ['IMAGE_REPO', 'SKIP_TESTS', 'DOCKERFILE_PATH', 'GIT_CREDENTIAL_ID']
            const filteredEnvVars = envArr.filter(e => !promotedKeys.includes(e.name))
            // 回显 Dockerfile 策略模式
            const savedDfPath = getEnv('DOCKERFILE_PATH', '')
            if (savedDfPath === '__PLATFORM_GENERATE__') {
              dockerfileMode.value = 'platform'
            } else if (savedDfPath) {
              dockerfileMode.value = 'project'
            } else {
              dockerfileMode.value = 'auto'
            }
            selectedServiceType.value = langType
            pipelineData.value = {
              name: data.name || '',
              description: data.description || '',
              git_repo: data.git_repo || '',
              git_branch: data.git_branch || 'main',
              jenkins_url: data.jenkins_url || '',
              jenkins_job: data.jenkins_job || '',
              language_type: langType,
              // 构建核心参数（从 env_vars 提取到独立字段）
              image_repo: getEnv('IMAGE_REPO', ''),
              skip_tests: getEnv('SKIP_TESTS', 'false') === 'true',
              dockerfile_path: getEnv('DOCKERFILE_PATH', ''),
              git_credential_id: getEnv('GIT_CREDENTIAL_ID', 'gitee-id'),
              env_vars: filteredEnvVars,
              enable_sonar: hasSonar,
              enable_artifact_upload: data.enable_artifact_upload || false,
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
            // 编辑回显后，按实际语言类型+环境重新加载资源模板
            selectedResourceTemplate.value = ''
            await loadResourceTemplates()
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

        // 构建提交数据：将独立字段注入 env_vars + 根据 enable_sonar 开关同步
        const submitData = { ...pipelineData.value }
        const envVars = [...(submitData.env_vars || [])]

        // 注入构建核心参数到 env_vars
        const injectEnv = (key, val) => {
          if (!val && val !== 'true' && val !== 'false') return
          const idx = envVars.findIndex(e => e.name === key)
          if (idx >= 0) { envVars[idx].value = String(val) }
          else { envVars.push({ name: key, value: String(val) }) }
        }
        injectEnv('IMAGE_REPO', submitData.image_repo)
        injectEnv('SKIP_TESTS', submitData.skip_tests ? 'true' : 'false')
        // Dockerfile 策略映射：
        //   auto     → 空（Jenkins 智能检测项目 Dockerfile → 回退平台生成）
        //   project  → 用户指定的路径，默认 'Dockerfile'
        //   platform → '__PLATFORM_GENERATE__'  强制平台生成
        if (dockerfileMode.value === 'project') {
          injectEnv('DOCKERFILE_PATH', submitData.dockerfile_path || 'Dockerfile')
        } else if (dockerfileMode.value === 'platform') {
          injectEnv('DOCKERFILE_PATH', '__PLATFORM_GENERATE__')
        }
        // auto 模式不发送 DOCKERFILE_PATH，由 Jenkins 智能检测
        if (submitData.git_credential_id) injectEnv('GIT_CREDENTIAL_ID', submitData.git_credential_id)

        // SonarQube 开关同步
        if (submitData.enable_sonar) {
          // 启用 SonarQube：确保 env_vars 中有 ENABLE_SONAR=true
          const idx = envVars.findIndex(e => e.name === 'ENABLE_SONAR')
          if (idx >= 0) {
            envVars[idx].value = 'true'
          } else {
            envVars.push({ name: 'ENABLE_SONAR', value: 'true' })
          }
          const gateIdx = envVars.findIndex(e => e.name === 'SONAR_QUALITY_GATE')
          if (gateIdx >= 0) {
            envVars[gateIdx].value = 'true'
          } else {
            envVars.push({ name: 'SONAR_QUALITY_GATE', value: 'true' })
          }
        } else {
          // 关闭 SonarQube：移除相关环境变量，避免残留导致 Jenkins 仍执行扫描
          const sonarKeys = ['ENABLE_SONAR', 'SONAR_QUALITY_GATE']
          for (let i = envVars.length - 1; i >= 0; i--) {
            if (sonarKeys.includes(envVars[i].name)) {
              envVars.splice(i, 1)
            }
          }
        }
        submitData.env_vars = envVars
        delete submitData.enable_sonar  // 后端不需要此字段（通过 env_vars 传递 ENABLE_SONAR）
        // 注意：enable_artifact_upload 是后端独立字段，保留传递
        // 清理前端独立字段，后端不需要
        delete submitData.image_repo
        delete submitData.skip_tests
        delete submitData.dockerfile_path
        delete submitData.git_credential_id

        if (isEdit) {
          response = await updatePipeline({
            id: parseInt(pipelineId),
            ...submitData
          })
        } else {
          response = await createPipeline(submitData)
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
          // 权限过滤：只显示用户有权限访问的集群
          const list = res.data.list || []
          clusters.value = list.filter(c => 
            permissionStore.state.isSuperAdmin ||
            permissionStore.state.accessibleClusterIds.includes(c.id)
          )
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
          let nsList = res.data.list || res.data || []
          
          // 权限过滤：只显示用户有权限访问的命名空间
          if (!permissionStore.state.isSuperAdmin) {
            const clusterId = pipelineData.value.target_cluster_id
            const accessibleNs = permissionStore.getAccessibleNamespaces(clusterId)
            if (accessibleNs.length > 0 && !accessibleNs.includes('*') && !accessibleNs.includes('__none__')) {
              nsList = nsList.filter(ns => accessibleNs.includes(ns.name || ns.metadata?.name))
            } else if (accessibleNs.includes('__none__')) {
              nsList = []
            }
          }
          
          namespaces.value = nsList
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
      // 切换环境后重新加载资源模板和校验
      selectedResourceTemplate.value = ''
      loadResourceTemplates()
      doValidateResource()
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

    onMounted(async () => {
      loadTemplates()
      if (isEdit) {
        // 编辑模式：先加载流水线数据（含语言类型、环境），再按实际参数加载资源模板
        await loadPipelineData()
      }
      await loadResourceTemplates()
      // 初始化时触发一次校验
      setTimeout(() => doValidateResource(), 500)
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
      // 资源模板相关
      resourceTemplates,
      selectedResourceTemplate,
      loadingResourceTemplates,
      resourceValidation,
      validatingResource,
      canApprove,
      serviceTypeOptions,
      selectedServiceType,
      dockerfileMode,
      dockerfileLangLabel,
      templateFileMap,
      onServiceTypeChange,
      onResourceTemplateChange,
      doValidateResource,
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

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
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

.btn-header-save {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: linear-gradient(135deg, #48bb78 0%, #38a169 100%);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s;
}

.btn-header-save:hover:not(:disabled) {
  background: linear-gradient(135deg, #38a169 0%, #2f855a 100%);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(72, 187, 120, 0.4);
}

.btn-header-save:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.btn-header-save svg {
  width: 16px;
  height: 16px;
}

/* ==================== 主体布局 ==================== */
.wizard-body {
  display: flex;
  max-width: 100%;
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

/* 构建参数分区 */
.build-params-section {
  margin-top: 20px;
  padding-top: 4px;
}

.section-divider {
  display: flex;
  align-items: center;
  margin-bottom: 20px;
}

.section-divider::before,
.section-divider::after {
  content: '';
  flex: 1;
  height: 1px;
  background: #e2e8f0;
}

.divider-text {
  padding: 0 14px;
  font-size: 13px;
  font-weight: 600;
  color: #4a5568;
  white-space: nowrap;
}

.form-row {
  display: flex;
  gap: 16px;
}

.form-group.half {
  flex: 1;
  min-width: 0;
}

.toggle-row.compact {
  padding: 10px 14px;
  margin-top: 6px;
}

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

/* ==================== 资源模板选择器 ==================== */
.resource-template-section {
  margin-bottom: 24px;
  padding-bottom: 24px;
  border-bottom: 1px solid #e2e8f0;
}

.form-row {
  display: flex;
  gap: 20px;
}

.form-group.half {
  flex: 1;
}

/* Step 4 部署环境选择器 */
.env-selector-inline {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.env-chip {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border: 2px solid #e2e8f0;
  border-radius: 20px;
  cursor: pointer;
  transition: all 0.2s;
  background: white;
  font-size: 13px;
  font-weight: 500;
  color: #64748b;
}

.env-chip:hover {
  border-color: #cbd5e1;
  background: #f8fafc;
}

.env-chip.selected {
  border-color: #3b82f6;
  background: #eff6ff;
  color: #1e40af;
}

.env-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.service-type-selector {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.service-type-card {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  border: 2px solid #e2e8f0;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  background: white;
}

.service-type-card:hover {
  border-color: #cbd5e1;
}

.service-type-card.selected {
  border-color: #3b82f6;
  background: #eff6ff;
}

.svc-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}

.svc-name {
  font-size: 13px;
  font-weight: 500;
  color: #334155;
}

.form-select {
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

.form-select:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.15);
}

/* ==================== 校验结果 ==================== */
.validation-result {
  padding: 16px;
  border-radius: 10px;
  margin-bottom: 20px;
}

.validation-result.success {
  background: #f0fdf4;
  border: 1px solid #86efac;
}

.validation-result.error {
  background: #fef2f2;
  border: 1px solid #fca5a5;
}

.validation-result.high {
  background: #fef2f2;
  border-color: #f87171;
}

.validation-result.medium {
  background: #fffbeb;
  border-color: #fbbf24;
}

.validation-header {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: 600;
  margin-bottom: 8px;
}

.validation-header svg {
  width: 20px;
  height: 20px;
}

.validation-result.success .validation-header {
  color: #16a34a;
}

.validation-result.success .validation-header svg {
  stroke: #16a34a;
}

.validation-result.error .validation-header {
  color: #dc2626;
}

.validation-result.error .validation-header svg {
  stroke: #dc2626;
}

.risk-badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
}

.risk-badge.high {
  background: #fee2e2;
  color: #b91c1c;
}

.risk-badge.medium {
  background: #fef3c7;
  color: #92400e;
}

.validation-errors,
.validation-warnings {
  margin: 8px 0;
  padding-left: 20px;
  font-size: 13px;
}

.validation-errors li {
  color: #dc2626;
  margin-bottom: 4px;
}

.validation-warnings li {
  color: #d97706;
  margin-bottom: 4px;
}

/* ==================== 审批卡片（大厂风格） ==================== */
.approval-card {
  margin-top: 16px;
  background: linear-gradient(135deg, #fffbeb 0%, #fef3c7 100%);
  border: 1px solid #fbbf24;
  border-radius: 12px;
  overflow: hidden;
}

.approval-card-header {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 16px 20px;
  background: rgba(251, 191, 36, 0.1);
  border-bottom: 1px solid rgba(251, 191, 36, 0.2);
}

.approval-icon {
  width: 44px;
  height: 44px;
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.approval-icon svg {
  width: 24px;
  height: 24px;
  stroke: white;
}

.approval-info {
  flex: 1;
}

.approval-title {
  font-size: 15px;
  font-weight: 600;
  color: #92400e;
  margin-bottom: 4px;
}

.approval-desc {
  font-size: 13px;
  color: #a16207;
}

.approval-desc strong {
  color: #92400e;
  font-weight: 600;
}

/* 有审批权限 */
.approval-actions {
  padding: 14px 20px;
}

.approval-status {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
}

.approval-status.approved {
  background: linear-gradient(135deg, #d1fae5 0%, #a7f3d0 100%);
  color: #065f46;
}

.approval-status.approved svg {
  width: 20px;
  height: 20px;
  stroke: #059669;
}

/* 无审批权限 */
.approval-pending {
  padding: 14px 20px;
}

.pending-info {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  background: rgba(251, 191, 36, 0.15);
  border-radius: 8px;
  font-size: 13px;
  color: #92400e;
}

.pending-info svg {
  width: 20px;
  height: 20px;
  stroke: #d97706;
  flex-shrink: 0;
}

/* 兼容旧样式 */
.approval-notice {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  background: #fef3c7;
  border-radius: 6px;
  margin-top: 10px;
  font-size: 13px;
  color: #92400e;
}

.approval-notice svg {
  width: 18px;
  height: 18px;
  stroke: #92400e;
  flex-shrink: 0;
}

.validation-suggestion {
  margin-top: 10px;
  padding: 10px 14px;
  background: #f0f9ff;
  border-radius: 6px;
  font-size: 13px;
  color: #0369a1;
}

/* ==================== Dockerfile 策略选择器 ==================== */
.dockerfile-mode-selector {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-bottom: 12px;
}

.df-mode-card {
  position: relative;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  border: 2px solid #e2e8f0;
  border-radius: 12px;
  background: white;
  cursor: pointer;
  transition: all 0.25s ease;
}

.df-mode-card:hover {
  border-color: #a0c4e8;
  background: #f7fafc;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(66, 153, 225, 0.1);
}

.df-mode-card.active {
  border-color: #4299e1;
  background: linear-gradient(135deg, #ebf8ff 0%, #f0f9ff 100%);
  box-shadow: 0 4px 16px rgba(66, 153, 225, 0.2);
}

.df-mode-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.df-mode-icon svg {
  width: 20px;
  height: 20px;
}

.df-mode-icon.auto {
  background: linear-gradient(135deg, #e6fffa 0%, #b2f5ea 100%);
  color: #234e52;
}

.df-mode-icon.project {
  background: linear-gradient(135deg, #fefcbf 0%, #fef08a 100%);
  color: #744210;
}

.df-mode-icon.platform {
  background: linear-gradient(135deg, #e9d8fd 0%, #d6bcfa 100%);
  color: #44337a;
}

.df-mode-info {
  flex: 1;
  min-width: 0;
}

.df-mode-title {
  font-size: 13px;
  font-weight: 600;
  color: #2d3748;
  display: flex;
  align-items: center;
  gap: 6px;
}

.df-mode-desc {
  font-size: 11px;
  color: #a0aec0;
  margin-top: 2px;
  line-height: 1.4;
}

.df-badge.recommend {
  display: inline-block;
  padding: 1px 6px;
  font-size: 10px;
  font-weight: 700;
  color: white;
  background: linear-gradient(135deg, #48bb78 0%, #38a169 100%);
  border-radius: 4px;
  letter-spacing: 0.5px;
}

.df-mode-check {
  position: absolute;
  top: 8px;
  right: 10px;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: linear-gradient(135deg, #4299e1 0%, #3182ce 100%);
  color: white;
  font-size: 12px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  animation: popIn 0.2s ease;
}

@keyframes popIn {
  from { transform: scale(0); }
  to { transform: scale(1); }
}

.df-path-input {
  margin-top: 12px;
  animation: fadeIn 0.2s ease;
}

/* Dockerfile 策略说明面板 */
.df-info-panel {
  margin-top: 12px;
  border-radius: 10px;
  overflow: hidden;
  animation: fadeIn 0.2s ease;
}

.df-info-content {
  padding: 14px 16px;
  background: #f7fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  font-size: 12px;
  line-height: 1.6;
  color: #4a5568;
}

.df-info-title {
  font-weight: 600;
  font-size: 13px;
  color: #2d3748;
  margin-bottom: 8px;
}

.df-info-steps {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.df-step {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: #4a5568;
}

.df-step-num {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: linear-gradient(135deg, #4299e1 0%, #3182ce 100%);
  color: white;
  font-size: 11px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.df-info-text {
  font-size: 12px;
  color: #4a5568;
  line-height: 1.7;
}

.df-info-text code {
  padding: 1px 5px;
  background: #edf2f7;
  border-radius: 4px;
  font-size: 11px;
  color: #e53e3e;
  font-family: 'Consolas', 'Monaco', monospace;
}

@media (max-width: 768px) {
  .dockerfile-mode-selector {
    grid-template-columns: 1fr;
  }
}
</style>
