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
          <h1>{{ isEdit ? '编辑流水线' : '快速创建流水线' }}</h1>
          <p>配置 CI/CD 流水线，实现代码自动构建和部署</p>
        </div>
      </div>
      <div class="header-actions">
        <!-- 从K8s导入按钮 -->
        <button
          v-if="!isEdit"
          type="button"
          class="btn-header-import"
          @click="showK8sImportModal = true"
          title="从K8s集群导入已有应用配置"
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
            <polyline points="7 10 12 15 17 10"/>
            <line x1="12" y1="15" x2="12" y2="3"/>
          </svg>
          从 K8s 导入
        </button>
        <!-- 保存按钮（编辑模式） -->
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

    <!-- 快速模板栏（顶部） -->
    <div v-if="!isEdit" class="template-bar">
      <div class="template-bar-content">
        <div class="template-bar-label">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/>
            <line x1="3" y1="9" x2="21" y2="9"/>
            <line x1="9" y1="21" x2="9" y2="9"/>
          </svg>
          选择模板快速开始：
        </div>
        <div class="template-bar-buttons">
          <button
            v-for="tpl in quickTemplates"
            :key="tpl.value"
            :class="['template-btn', { active: pipelineData.language_type === tpl.value }]"
            @click="applyQuickTemplate(tpl.value)"
          >
            <span class="tpl-dot" :style="{ backgroundColor: tpl.color }"></span>
            <span class="tpl-name">{{ tpl.label }}</span>
            <span v-if="tpl.badge" :class="['tpl-badge', tpl.badgeClass]">{{ tpl.badge }}</span>
          </button>
        </div>
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

        <!-- 快速模板选择（侧边栏保留） -->
        <div class="template-selector">
          <div class="template-label">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/>
              <line x1="3" y1="9" x2="21" y2="9"/>
              <line x1="9" y1="21" x2="9" y2="9"/>
            </svg>
            流水线模板
          </div>
          <select v-model="selectedTemplateId" @change="handleTemplateChange" class="template-select">
            <option value="">不使用模板</option>
            <option v-for="template in templates" :key="template.id" :value="template.id">
              {{ template.name }}
            </option>
          </select>
        </div>
      </div>

      <!-- 中间表单内容 -->
      <div class="wizard-content">
        <form @submit.prevent="submit">
          <!-- Step 1: 应用信息（基础信息 + 构建类型 + Git仓库 + 分支） -->
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
                <h2>应用信息</h2>
                <p>设置应用名称、代码仓库和构建类型</p>
              </div>
            </div>
                    
            <div class="form-card">
              <!-- 应用基础信息区块 -->
              <div class="form-section">
                <div class="section-title">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:16px;height:16px;">
                    <path d="M12 2L2 7l10 5 10-5-10-5z"/>
                    <path d="M2 17l10 5 10-5"/>
                    <path d="M2 12l10 5 10-5"/>
                  </svg>
                  应用基础信息
                </div>
          
                <!-- 应用名称 -->
                <div class="form-group">
                  <label class="form-label">
                    应用名称
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
                      placeholder="例如：user-service 或 springboot-hello"
                      required
                      @blur="checkName"
                    />
                  </div>
                  <div v-if="nameChecking" class="input-hint" style="color:#94a3b8">… 检查中</div>
                  <div v-else-if="nameAvailable === false" class="input-hint" style="color:#ef4444">❌ {{ nameCheckMsg }}</div>
                  <div v-else-if="nameAvailable === true" class="input-hint" style="color:#22c55e">✅ 名称可用</div>
                  <div v-else class="input-hint">建议使用小写字母和连字符，同时作为 K8s 工作负载名称</div>
                </div>
                        
                <!-- Git 仓库地址 -->
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
                        placeholder="https://gitee.com/your-org/your-repo.git"
                        required
                        @blur="onRepoUrlChange"
                      />
                    </div>
                    <button
                      type="button"
                      class="btn-detect-repo"
                      @click="detectRepo"
                      :disabled="!pipelineData.git_repo || detectingRepo"
                      :title="detectingRepo ? '检测中...' : '检测仓库信息'"
                    >
                      <svg v-if="!detectingRepo" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <circle cx="11" cy="11" r="8"/>
                        <path d="M21 21l-4.35-4.35"/>
                        <path d="M11 8v6M8 11h6"/>
                      </svg>
                      <span v-else class="loading-spinner-sm"></span>
                      {{ detectingRepo ? '检测中' : '检测仓库' }}
                    </button>
                  </div>
                            
                  <!-- 仓库检测结果 -->
                  <div v-if="repoDetectionResult" class="repo-detection-result">
                    <div class="result-header">
                      仓库检测结果
                    </div>
                    <div class="result-grid">
                      <div class="result-item">
                        <span class="result-label">仓库类型</span>
                        <span class="result-value" :class="repoDetectionResult.repoType.toLowerCase()">{{ repoDetectionResult.repoType }}</span>
                      </div>
                      <div class="result-item">
                        <span class="result-label">默认分支</span>
                        <span class="result-value branch">{{ repoDetectionResult.defaultBranch }}</span>
                      </div>
                      <div class="result-item">
                        <span class="result-label">语言</span>
                        <span class="result-value language">{{ repoDetectionResult.language }}</span>
                      </div>
                      <div class="result-item">
                        <span class="result-label">构建工具</span>
                        <span class="result-value">{{ repoDetectionResult.buildTool }}</span>
                      </div>
                      <div class="result-item">
                        <span class="result-label">Dockerfile</span>
                        <span class="result-value" :class="repoDetectionResult.hasDockerfile ? 'found' : 'not-found'">
                          {{ repoDetectionResult.hasDockerfile ? '✓ 已发现' : '✗ 未发现' }}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
                        
                <!-- 分支 -->
                <div class="form-group">
                  <label class="form-label">
                    代码分支
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
                    <span v-if="branches.length === 0">输入分支名称，或点击上方"检测仓库"按钮自动获取</span>
                    <span v-else>已选择：<strong>{{ pipelineData.git_branch }}</strong></span>
                  </div>
                </div>
              </div>
          
              <!-- 构建类型区块 -->
              <div class="form-section">
                <div class="section-title">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:16px;height:16px;">
                    <polyline points="16 18 22 12 16 6"/>
                    <polyline points="8 6 2 12 8 18"/>
                  </svg>
                  构建类型
                </div>
          
                <!-- 语言/框架类型 -->
                <div class="form-group">
                  <label class="form-label">
                    语言/框架类型
                    <span class="required">*</span>
                  </label>
                  <div class="quick-lang-selector">
                    <div
                      v-for="opt in serviceTypeOptions"
                      :key="opt.value"
                      :class="['lang-chip', { selected: pipelineData.language_type === opt.value }]"
                      @click="pipelineData.language_type = opt.value; selectedServiceType = opt.value"
                    >
                      <span class="lang-dot" :style="{ backgroundColor: opt.color }"></span>
                      {{ opt.label }}
                      <span v-if="opt.badge" :class="['lang-badge', opt.badgeClass]">{{ opt.badge }}</span>
                    </div>
                  </div>
                  <div class="input-hint">选择语言后自动匹配 Jenkins 构建模板，无需手动配置</div>
                </div>
              </div>
                        
              <!-- 描述（可折叠） -->
              <div class="form-group">
                <label class="form-label" style="cursor:pointer;display:flex;align-items:center;gap:6px;" @click="showDescription = !showDescription">
                  描述
                  <span style="font-size:11px;color:#94a3b8;font-weight:400;">(选填)</span>
                  <svg :style="{transform:showDescription?'rotate(180deg)':'rotate(0)',transition:'0.2s'}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="12" height="12"><polyline points="6 9 12 15 18 9"/></svg>
                </label>
                <textarea
                  v-show="showDescription"
                  v-model="pipelineData.description"
                  class="form-textarea"
                  placeholder="简要描述此应用的用途..."
                  rows="2"
                ></textarea>
              </div>
            </div>
          </div>
          
          <!-- Step 2: 构建配置 -->
          <div v-show="currentStep === 1" class="step-panel">
            <div class="panel-header">
              <div class="panel-icon jenkins">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <rect x="2" y="3" width="20" height="14" rx="2" ry="2"/>
                  <line x1="8" y1="21" x2="16" y2="21"/>
                  <line x1="12" y1="17" x2="12" y2="21"/>
                </svg>
              </div>
              <div>
                <h2>构建配置</h2>
                <p>镜像仓库和 Jenkins 构建参数</p>
              </div>
            </div>

            <div class="form-card">
              <!-- 默认顯示：构建功能开关 -->
              <!-- Jenkins 高级配置已自动处理，隐藏不显示 -->
              <div class="form-group" v-show="false">
                <div
                  class="advanced-toggle"
                  @click="showJenkinsAdvanced = !showJenkinsAdvanced"
                  style="display:flex;align-items:center;gap:6px;cursor:pointer;padding:8px 12px;background:#f8f9fb;border:1px solid #e8eaf0;border-radius:8px;margin-bottom:4px;"
                >
                  <svg :style="{transform: showJenkinsAdvanced?'rotate(90deg)':'rotate(0deg)',transition:'0.2s'}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><polyline points="9 18 15 12 9 6"/></svg>
                  <span style="font-size:13px;font-weight:500;color:#374151;">高级配置（Jenkins 地址 / Job 名称）</span>
                  <span style="margin-left:auto;font-size:11px;color:#8c8c8c;">留空即自动推导，一般无需修改</span>
                </div>
                <div v-show="showJenkinsAdvanced">
                  <div class="form-group" style="margin-top:12px;">
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
                  <div class="form-group" style="margin-top:12px;">
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
                          <span style="color:#52c41a;">&#10004;</span> 使用自定义 Job <strong>{{ pipelineData.jenkins_job }}</strong>，平台会自动注入 <code style="color:#1890ff;">LANGUAGE_TYPE={{ pipelineData.language_type }}</code> 参数
                        </template>
                        <template v-else>
                          留空将自动使用平台内置 Job: <strong>k8s-builder-{{ pipelineData.language_type }}</strong>（推荐）
                        </template>
                      </template>
                      <template v-else>自定义类型必须填写 Jenkins Job 名称（Jenkins 上已创建的 Job）</template>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Java 版本选择器（仅 Java 项目显示） -->
              <div v-if="pipelineData.language_type === 'java'" class="form-group">
                <label class="form-label">
                  Java 版本
                  <span class="env-tag" style="background:#1890ff;color:#fff;font-size:11px;padding:1px 6px;border-radius:3px;margin-left:6px;">JDK</span>
                </label>
                <div class="java-version-selector">
                  <div
                    v-for="ver in ['17', '11']"
                    :key="ver"
                    :class="['java-ver-chip', { selected: pipelineData.java_version === ver }]"
                    @click="pipelineData.java_version = ver"
                  >
                    <span class="ver-number">{{ ver }}</span>
                    <span v-if="ver === '17'" class="ver-badge">推荐</span>
                  </div>
                </div>
                <div class="input-hint">决定构建环境 JDK 版本（maven:3.9-eclipse-temurin-<strong>{{ pipelineData.java_version }}</strong>）和运行时基础镜像</div>
              </div>

              <!-- SonarQube 代码质量扫描开关（仅管理员可见） -->
              <div v-if="canApprove" class="form-group">
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

                <!-- SonarQube 质量门禁配置面板（展开式） -->
                <transition name="sonar-expand">
                  <div v-if="pipelineData.enable_sonar" class="sonar-config-panel">
                    <div class="sonar-panel-header">
                      <div class="sonar-panel-title">
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                          <path d="M12 20V10M18 20V4M6 20v-4"/>
                        </svg>
                        质量门禁规则配置
                      </div>
                      <span class="sonar-panel-badge">Quality Gate</span>
                    </div>

                    <div class="sonar-metrics-grid">
                      <!-- 代码覆盖率 -->
                      <div class="sonar-metric-card">
                        <div class="metric-header">
                          <div class="metric-icon coverage">
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
                            </svg>
                          </div>
                          <div class="metric-info">
                            <div class="metric-name">代码覆盖率</div>
                            <div class="metric-desc">新增代码单测覆盖率阈值</div>
                          </div>
                        </div>
                        <div class="metric-input-row">
                          <div class="metric-slider-wrap">
                            <input type="range" v-model.number="sonarConfig.coverage" min="0" max="100" step="5" class="metric-slider" />
                            <div class="metric-slider-track">
                              <div class="metric-slider-fill" :style="{ width: sonarConfig.coverage + '%' }"></div>
                            </div>
                          </div>
                          <div class="metric-value-box">
                            <input type="number" v-model.number="sonarConfig.coverage" min="0" max="100" class="metric-num-input" />
                            <span class="metric-unit">%</span>
                          </div>
                        </div>
                      </div>

                      <!-- 新代码 Bug -->
                      <div class="sonar-metric-card">
                        <div class="metric-header">
                          <div class="metric-icon bugs">
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                              <circle cx="12" cy="12" r="10"/>
                              <line x1="12" y1="8" x2="12" y2="12"/>
                              <line x1="12" y1="16" x2="12.01" y2="16"/>
                            </svg>
                          </div>
                          <div class="metric-info">
                            <div class="metric-name">新增 Bug</div>
                            <div class="metric-desc">允许的最大新增 Bug 数量</div>
                          </div>
                        </div>
                        <div class="metric-input-row">
                          <div class="metric-presets">
                            <button v-for="v in [0, 1, 3, 5, 10]" :key="v"
                              :class="['preset-btn', { active: sonarConfig.newBugs === v }]"
                              @click="sonarConfig.newBugs = v" type="button"
                            >{{ v }}</button>
                          </div>
                          <div class="metric-value-box">
                            <input type="number" v-model.number="sonarConfig.newBugs" min="0" max="999" class="metric-num-input" />
                            <span class="metric-unit">个</span>
                          </div>
                        </div>
                      </div>

                      <!-- 代码异味 -->
                      <div class="sonar-metric-card">
                        <div class="metric-header">
                          <div class="metric-icon smells">
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                              <path d="M8 12.5c0 1.38-1.12 2.5-2.5 2.5S3 13.88 3 12.5 4.12 10 5.5 10 8 11.12 8 12.5z"/>
                              <path d="M14.5 7c1.38 0 2.5 1.12 2.5 2.5S15.88 12 14.5 12 12 10.88 12 9.5 13.12 7 14.5 7z"/>
                              <path d="M16 19.5c0-1.38 1.12-2.5 2.5-2.5s2.5 1.12 2.5 2.5-1.12 2.5-2.5 2.5-2.5-1.12-2.5-2.5z"/>
                            </svg>
                          </div>
                          <div class="metric-info">
                            <div class="metric-name">代码异味</div>
                            <div class="metric-desc">允许的最大 Code Smells 数</div>
                          </div>
                        </div>
                        <div class="metric-input-row">
                          <div class="metric-presets">
                            <button v-for="v in [0, 5, 10, 20, 50]" :key="v"
                              :class="['preset-btn', { active: sonarConfig.codeSmells === v }]"
                              @click="sonarConfig.codeSmells = v" type="button"
                            >{{ v }}</button>
                          </div>
                          <div class="metric-value-box">
                            <input type="number" v-model.number="sonarConfig.codeSmells" min="0" max="999" class="metric-num-input" />
                            <span class="metric-unit">个</span>
                          </div>
                        </div>
                      </div>

                      <!-- 安全漏洞 -->
                      <div class="sonar-metric-card">
                        <div class="metric-header">
                          <div class="metric-icon vulnerabilities">
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                              <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/>
                              <path d="M7 11V7a5 5 0 0 1 10 0v4"/>
                            </svg>
                          </div>
                          <div class="metric-info">
                            <div class="metric-name">安全漏洞</div>
                            <div class="metric-desc">允许的最大安全漏洞数</div>
                          </div>
                        </div>
                        <div class="metric-input-row">
                          <div class="metric-presets">
                            <button v-for="v in [0, 1, 3, 5]" :key="v"
                              :class="['preset-btn', { active: sonarConfig.vulnerabilities === v }]"
                              @click="sonarConfig.vulnerabilities = v" type="button"
                            >{{ v }}</button>
                          </div>
                          <div class="metric-value-box">
                            <input type="number" v-model.number="sonarConfig.vulnerabilities" min="0" max="999" class="metric-num-input" />
                            <span class="metric-unit">个</span>
                          </div>
                        </div>
                      </div>

                      <!-- 重复率 -->
                      <div class="sonar-metric-card">
                        <div class="metric-header">
                          <div class="metric-icon duplications">
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                              <rect x="9" y="9" width="13" height="13" rx="2" ry="2"/>
                              <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
                            </svg>
                          </div>
                          <div class="metric-info">
                            <div class="metric-name">代码重复率</div>
                            <div class="metric-desc">新增代码最大重复率阈值</div>
                          </div>
                        </div>
                        <div class="metric-input-row">
                          <div class="metric-slider-wrap">
                            <input type="range" v-model.number="sonarConfig.duplications" min="0" max="50" step="1" class="metric-slider" />
                            <div class="metric-slider-track">
                              <div class="metric-slider-fill" :style="{ width: (sonarConfig.duplications / 50 * 100) + '%' }"></div>
                            </div>
                          </div>
                          <div class="metric-value-box">
                            <input type="number" v-model.number="sonarConfig.duplications" min="0" max="50" class="metric-num-input" />
                            <span class="metric-unit">%</span>
                          </div>
                        </div>
                      </div>

                      <!-- 质量门禁阻断策略 -->
                      <div class="sonar-metric-card full-width">
                        <div class="metric-header">
                          <div class="metric-icon gate">
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                              <polygon points="12 2 2 7 12 12 22 7 12 2"/>
                              <polyline points="2 17 12 22 22 17"/>
                              <polyline points="2 12 12 17 22 12"/>
                            </svg>
                          </div>
                          <div class="metric-info">
                            <div class="metric-name">门禁失败策略</div>
                            <div class="metric-desc">质量门禁未通过时的处理方式</div>
                          </div>
                        </div>
                        <div class="gate-strategy-selector">
                          <div :class="['gate-option', { active: sonarConfig.gateAction === 'block' }]"
                               @click="sonarConfig.gateAction = 'block'">
                            <div class="gate-option-icon block">
                              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <circle cx="12" cy="12" r="10"/>
                                <line x1="4.93" y1="4.93" x2="19.07" y2="19.07"/>
                              </svg>
                            </div>
                            <div>
                              <div class="gate-option-title">阻断构建</div>
                              <div class="gate-option-desc">门禁失败则标记构建失败</div>
                            </div>
                          </div>
                          <div :class="['gate-option', { active: sonarConfig.gateAction === 'warn' }]"
                               @click="sonarConfig.gateAction = 'warn'">
                            <div class="gate-option-icon warn">
                              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>
                                <line x1="12" y1="9" x2="12" y2="13"/>
                                <line x1="12" y1="17" x2="12.01" y2="17"/>
                              </svg>
                            </div>
                            <div>
                              <div class="gate-option-title">仅告警</div>
                              <div class="gate-option-desc">门禁失败仅发通知，不阻断</div>
                            </div>
                          </div>
                          <div :class="['gate-option', { active: sonarConfig.gateAction === 'skip' }]"
                               @click="sonarConfig.gateAction = 'skip'">
                            <div class="gate-option-icon skip">
                              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <polyline points="4 17 10 11 4 5"/>
                                <line x1="12" y1="19" x2="20" y2="19"/>
                              </svg>
                            </div>
                            <div>
                              <div class="gate-option-title">跳过门禁</div>
                              <div class="gate-option-desc">仅执行扫描，不判定门禁</div>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>

                    <div class="sonar-panel-footer">
                      <div class="sonar-tip">
                        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                          <circle cx="12" cy="12" r="10"/>
                          <path d="M12 16v-4"/>
                          <path d="M12 8h.01"/>
                        </svg>
                        以上参数将通过环境变量注入 Jenkins，由 SonarQube Scanner 执行检查
                      </div>
                    </div>
                  </div>
                </transition>
              </div>

              <!-- 制品上传到平台制品库开关（仅管理员可见） -->
              <div v-if="canApprove" class="form-group">
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
                      :placeholder="defaultImageRegistry ? `${defaultImageRegistry}/应用名` : 'harbor.example.com/project/app-name'"
                      required
                    />
                  </div>
                  <div class="input-hint">{{ defaultImageRegistry ? `已配置默认前缀: ${defaultImageRegistry}，填写应用名后自动拼接` : 'Jenkins 构建后将镜像推送到此地址，格式：registry/project/app' }}</div>
                </div>

                <!-- 镜像标签策略 -->
                <div class="form-group">
                  <label class="form-label">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:15px;height:15px;vertical-align:middle;margin-right:4px;">
                      <path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"/>
                      <line x1="7" y1="7" x2="7.01" y2="7"/>
                    </svg>
                    镜像标签
                  </label>
                  <div class="tag-strategy-selector">
                    <div
                      class="tag-strategy-btn"
                      :class="{ active: tagStrategy === 'latest' }"
                      @click="setTagStrategy('latest')"
                    >
                      <span class="tag-strategy-icon">🏷️</span>
                      <span class="tag-strategy-label">latest</span>
                      <span class="tag-strategy-desc">固定覆盖</span>
                    </div>
                    <div
                      class="tag-strategy-btn"
                      :class="{ active: tagStrategy === 'auto' }"
                      @click="setTagStrategy('auto')"
                    >
                      <span class="tag-strategy-icon">🔄</span>
                      <span class="tag-strategy-label">自动生成</span>
                      <span class="tag-strategy-desc">commit-时间戳</span>
                    </div>
                    <div
                      class="tag-strategy-btn"
                      :class="{ active: tagStrategy === 'custom' }"
                      @click="setTagStrategy('custom')"
                    >
                      <span class="tag-strategy-icon">✏️</span>
                      <span class="tag-strategy-label">自定义</span>
                      <span class="tag-strategy-desc">固定版本号</span>
                    </div>
                  </div>
                  <div v-if="tagStrategy === 'custom'" class="input-wrapper" style="margin-top: 8px;">
                    <svg class="input-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"/>
                      <line x1="7" y1="7" x2="7.01" y2="7"/>
                    </svg>
                    <input
                      type="text"
                      v-model="pipelineData.image_tag"
                      class="form-input with-icon"
                      placeholder="输入固定标签，如 v1.0.0"
                    />
                  </div>
                  <div class="input-hint">
                    <template v-if="tagStrategy === 'latest'">每次构建推送 latest 标签，适合覆盖式部署（如 harbor.maitian-yun.com/foxess/wms:latest）</template>
                    <template v-else-if="tagStrategy === 'auto'">Jenkins 自动生成 commit-timestamp 格式标签，适合版本追溯</template>
                    <template v-else>固定标签如 v1.0.0，适合手动控制版本号</template>
                  </div>
                </div>

                <!-- 镜像实时预览 -->
                <div v-if="pipelineData.image_repo" class="form-group">
                  <label class="form-label">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:15px;height:15px;vertical-align:middle;margin-right:4px;">
                      <circle cx="12" cy="12" r="10"/>
                      <line x1="12" y1="8" x2="12" y2="12"/>
                      <line x1="12" y1="16" x2="12.01" y2="16"/>
                    </svg>
                    镜像预览
                    <span class="preview-badge">实时生成</span>
                  </label>
                  <div class="image-preview-panel">
                    <div class="preview-row">
                      <span class="preview-label">最终镜像：</span>
                      <div class="preview-image-full">
                        <span class="image-registry">{{ imagePreview.registry }}</span><span class="image-path">/{{ imagePreview.project }}</span><span class="image-name">/{{ imagePreview.app }}</span><span class="image-tag">:{{ imagePreview.tag }}</span>
                      </div>
                    </div>
                    <div class="preview-tags">
                      <span class="tag-hint">标签规则：</span>
                      <code v-if="tagStrategy === 'latest'">固定标签 → latest</code>
                      <code v-else-if="tagStrategy === 'custom' && pipelineData.image_tag">固定标签 → {{ pipelineData.image_tag }}</code>
                      <code v-else>{{ pipelineData.git_branch || 'main' }} → {{ imagePreview.tag }}</code>
                    </div>
                  </div>
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
                      :class="['df-mode-card', 'active']"
                    >
                      <div class="df-mode-icon platform">
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                          <rect x="2" y="3" width="20" height="14" rx="2" ry="2"/>
                          <line x1="8" y1="21" x2="16" y2="21"/>
                          <line x1="12" y1="17" x2="12" y2="21"/>
                        </svg>
                      </div>
                      <div class="df-mode-info">
                        <div class="df-mode-title">平台统一生成<span class="df-badge recommend">推荐</span></div>
                        <div class="df-mode-desc">由平台生成生产级最优 Dockerfile，忽略项目自带文件</div>
                      </div>
                      <div class="df-mode-check">&#10003;</div>
                    </div>
                  </div>

                  <!-- 策略说明面板 -->
                  <div class="df-info-panel">
                    <div class="df-info-content">
                      <div class="df-info-title">&#129302; 平台统一生成说明</div>
                      <div class="df-info-steps">
                        <div class="df-step"><span class="df-step-num">1</span>平台根据 {{ dockerfileLangLabel }} 语言类型自动生成生产级 Dockerfile</div>
                        <div class="df-step"><span class="df-step-num">2</span>自动注入「构建探针管理」中已启用的 Agent（如 SkyWalking / OpenTelemetry）</div>
                        <div class="df-step"><span class="df-step-num">3</span>阿里云镜像源加速 &bull; 非 root 用户 &bull; 最小化镜像层 &bull; 生产级 JVM 参数</div>
                      </div>
                    </div>
                  </div>
                </div>
                
                <!-- Dockerfile 预览 -->
                <div class="form-group">
                  <label class="form-label">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:15px;height:15px;vertical-align:middle;margin-right:4px;">
                      <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                      <polyline points="14 2 14 8 20 8"/>
                    </svg>
                    Dockerfile 预览
                    <button type="button" class="btn-view-dockerfile" @click="showDockerfilePreview = !showDockerfilePreview">
                      {{ showDockerfilePreview ? '收起' : '查看' }}
                    </button>
                  </label>
                                  
                  <transition name="slide-fade">
                    <div v-if="showDockerfilePreview" class="dockerfile-preview-panel">
                      <div class="dockerfile-header">
                        <div class="dockerfile-type">
                          <span class="type-badge" :class="dockerfileMode">{{ dockerfileModeLabel }}</span>
                          <span class="type-lang">{{ dockerfileLangLabel }}</span>
                        </div>
                        <button type="button" class="btn-copy-dockerfile" @click="copyDockerfile" title="复制 Dockerfile 内容">
                          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <rect x="9" y="9" width="13" height="13" rx="2" ry="2"/>
                            <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
                          </svg>
                          {{ copiedDockerfile ? '已复制' : '复制' }}
                        </button>
                      </div>
                      <div class="dockerfile-content">
                        <pre><code>{{ dockerfileContent }}</code></pre>
                      </div>
                    </div>
                  </transition>
                </div>

                <!-- 构建目录（仅 Java 多模块项目时显示） -->
                <div v-if="pipelineData.language_type === 'java'" class="form-group">
                  <label class="form-label">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:15px;height:15px;vertical-align:middle;margin-right:4px;">
                      <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
                    </svg>
                    构建目录
                    <span class="optional-badge">可选</span>
                    <span class="env-tag" style="background:#fa8c16;color:#fff;font-size:11px;padding:1px 6px;border-radius:3px;margin-left:6px;">多模块</span>
                  </label>
                  <div class="input-wrapper">
                    <svg class="input-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
                    </svg>
                    <input
                      type="text"
                      v-model="pipelineData.build_dir"
                      class="form-input with-icon"
                      placeholder="留空则自动检测，如 foxess-writer"
                    />
                  </div>
                  <div class="input-hint">多模块 Maven 项目必填：指定包含 spring-boot-maven-plugin 的子模块目录名，避免构建产物选错模块</div>
                </div>

                <!-- 私有 Maven 仓库地址 -->
                <div class="form-group">
                  <label class="form-label">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:15px;height:15px;vertical-align:middle;margin-right:4px;">
                      <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
                      <polyline points="3.27 6.96 12 12.01 20.73 6.96"/>
                      <line x1="12" y1="22.08" x2="12" y2="12"/>
                    </svg>
                    私有 Maven 仓库
                    <span class="optional-badge">可选</span>
                  </label>
                  <div class="input-wrapper">
                    <svg class="input-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
                    </svg>
                    <input
                      type="text"
                      v-model="pipelineData.maven_private_repo_url"
                      class="form-input with-icon"
                      placeholder="留空则仅使用阿里云公共仓库，如 https://nexus.example.com/repository/maven-releases/"
                    />
                  </div>
                  <div class="input-hint">拉取公司内部依赖包时必填（Nexus/GitLab Maven Registry 等），需在 Jenkins 配置名为 maven-private-repo 的 Username/Password 凭证</div>
                </div>

                <!-- Git 凭证 ID 已自动配置，仅保留跳过测试 -->
                <div class="form-row">
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

          <!-- Step 3 (部署策略) 已移除，使用默认配置 -->
          <div v-show="false" class="step-panel">
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
                  <!-- 语言类型自动继承，显示当前选择，无需重复配置 -->
                  <div class="form-group half">
                    <label class="form-label">语言类型</label>
                    <div class="form-input" style="display:flex;align-items:center;gap:8px;background:#f8f9fb;cursor:default;">
                      <span class="env-dot" :style="{ backgroundColor: serviceTypeOptions.find(o=>o.value===pipelineData.language_type)?.color || '#8c8c8c' }" style="width:10px;height:10px;border-radius:50%;flex-shrink:0;"></span>
                      <span style="font-weight:500;">{{ serviceTypeOptions.find(o=>o.value===pipelineData.language_type)?.label || pipelineData.language_type }}</span>
                      <span style="margin-left:auto;font-size:11px;color:#8c8c8c;">继承自步骤1</span>
                    </div>
                    <div class="input-hint">资源模板自动按语言类型匹配</div>
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

          <!-- Step 3: 自动部署配置 (now step index 2) -->
          <div v-show="currentStep === 2" class="step-panel">
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
              <!-- 自动部署开关（仅管理员可见） -->
              <div v-if="canApprove" class="form-group">
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
                  <div class="input-hint">预发环境和生产环境将自动开启审批流程</div>
                </div>

                <!-- 审批开关（仅管理员可见） -->
                <div v-if="canApprove" class="form-group">
                  <div class="toggle-row">
                    <div class="toggle-info">
                      <label class="form-label">发布审批</label>
                      <p class="toggle-desc">开启后，构建成功需审批通过才能部署到 K8s</p>
                    </div>
                    <label class="toggle-switch">
                      <input type="checkbox" v-model="pipelineData.require_approval" />
                      <span class="toggle-slider"></span>
                    </label>
                  </div>
                  <div v-if="pipelineData.require_approval" class="approval-hint">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:14px;height:14px;flex-shrink:0;">
                      <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
                    </svg>
                    <span>审批策略可在「CI/CD → 审批策略」中配置多级审批链</span>
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
                      v-for="opt in workloadKindOptions"
                      :key="opt.value"
                      class="kind-card"
                      :class="{ selected: pipelineData.target_workload_kind === opt.value }"
                      @click="selectWorkloadKind(opt.value)"
                    >
                      <span class="kind-name">{{ opt.label }}</span>
                      <span class="kind-desc">{{ opt.description }}</span>
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

                <!-- 容器名称 - 默认与工作负载名称一致，可自定义修改 -->
                <div class="form-group">
                  <label class="form-label">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:15px;height:15px;vertical-align:middle;margin-right:4px;">
                      <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
                    </svg>
                    Pod 容器名称
                    <span class="optional-badge">可选</span>
                  </label>
                  <div class="input-wrapper">
                    <svg class="input-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
                    </svg>
                    <input
                      type="text"
                      v-model="pipelineData.target_container"
                      class="form-input with-icon"
                      :placeholder="pipelineData.target_workload_name || '默认与工作负载名称一致'"
                    />
                  </div>
                  <div class="input-hint">部署时更新该容器的镜像，默认与工作负载名称一致，多容器 Pod 需指定</div>
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

      <!-- 右侧部署预览面板 -->
      <div v-if="!isEdit && !showSuccessTopology" class="deploy-preview-panel">
        <div class="preview-card">
          <div class="preview-title">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
              <circle cx="12" cy="12" r="3"/>
            </svg>
            部署预览
          </div>
          <div class="preview-list">
            <div class="preview-item">
              <span class="preview-label">集群</span>
              <span class="preview-value">{{ deployPreview.cluster }}</span>
            </div>
            <div class="preview-item">
              <span class="preview-label">命名空间</span>
              <span class="preview-value">{{ deployPreview.namespace }}</span>
            </div>
            <div class="preview-item">
              <span class="preview-label">工作负载</span>
              <span class="preview-value">{{ deployPreview.workload }}</span>
            </div>
            <div class="preview-item">
              <span class="preview-label">副本数</span>
              <span class="preview-value">{{ deployPreview.replicas }}</span>
            </div>
            <div class="preview-item highlight">
              <span class="preview-label">即将发布镜像</span>
              <div v-if="deployPreview.newImage" class="preview-image">{{ deployPreview.newImage }}</div>
              <span v-else class="preview-value">-</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 确认创建弹窗 -->
    <transition name="slide-fade">
      <div v-if="showConfirmStep" class="confirm-overlay" @click.self="backFromConfirm">
        <div class="confirm-modal">
          <div class="confirm-header">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"/>
            </svg>
            <h3>确认创建流水线</h3>
          </div>
          <div class="confirm-grid">
            <div class="confirm-item">
              <span class="confirm-label">应用名称</span>
              <span class="confirm-value">{{ pipelineData.name }}</span>
            </div>
            <div class="confirm-item">
              <span class="confirm-label">Git 仓库</span>
              <span class="confirm-value">{{ pipelineData.git_repo }}</span>
            </div>
            <div class="confirm-item">
              <span class="confirm-label">分支</span>
              <span class="confirm-value">{{ pipelineData.git_branch }}</span>
            </div>
            <div class="confirm-item">
              <span class="confirm-label">语言类型</span>
              <span class="confirm-value">{{ serviceTypeOptions.find(o => o.value === pipelineData.language_type)?.label || pipelineData.language_type }}</span>
            </div>
            <div class="confirm-item">
              <span class="confirm-label">镜像仓库</span>
              <span class="confirm-value">{{ pipelineData.image_repo || '-' }}</span>
            </div>
            <div v-if="pipelineData.auto_deploy" class="confirm-item">
              <span class="confirm-label">部署目标</span>
              <span class="confirm-value">{{ getClusterName(pipelineData.target_cluster_id) }} / {{ pipelineData.target_namespace }} / {{ pipelineData.target_workload_name }}</span>
            </div>
          </div>
          <div class="confirm-actions">
            <button class="btn-confirm-back" @click="backFromConfirm">返回修改</button>
            <button class="btn-confirm-submit" @click="submit" :disabled="submitting">
              <span v-if="submitting" class="loading-spinner-sm"></span>
              {{ submitting ? '创建中...' : '确认创建' }}
            </button>
          </div>
        </div>
      </div>
    </transition>

    <!-- K8s 导入弹窗 -->
    <transition name="slide-fade">
      <div v-if="showK8sImportModal" class="k8s-import-overlay" @click.self="showK8sImportModal = false">
        <div class="k8s-import-modal">
          <div class="k8s-import-header">
            <h3>
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
                <polyline points="7 10 12 15 17 10"/>
                <line x1="12" y1="15" x2="12" y2="3"/>
              </svg>
              从 K8s 导入
            </h3>
            <button class="k8s-import-close" @click="showK8sImportModal = false">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M6 18L18 6M6 6l12 12"/>
              </svg>
            </button>
          </div>
          <div class="k8s-import-form">
            <div class="form-group">
              <label class="form-label">集群</label>
              <select v-model="k8sImportForm.cluster_id" @change="onK8sImportClusterChange" class="form-select">
                <option :value="0">请选择集群</option>
                <option v-for="c in k8sImportClusters" :key="c.id" :value="c.id">{{ c.cluster_name }}</option>
              </select>
            </div>
            <div class="form-group">
              <label class="form-label">命名空间</label>
              <select v-model="k8sImportForm.namespace" @change="onK8sImportNamespaceChange" class="form-select" :disabled="!k8sImportForm.cluster_id">
                <option value="">请选择命名空间</option>
                <option v-for="ns in k8sImportNamespaces" :key="ns.name" :value="ns.name">{{ ns.name }}</option>
              </select>
            </div>
            <div class="form-group">
              <label class="form-label">Deployment</label>
              <select v-model="k8sImportForm.deployment" class="form-select" :disabled="!k8sImportForm.namespace">
                <option value="">请选择 Deployment</option>
                <option v-for="d in k8sImportDeployments" :key="d.name" :value="d.name">{{ d.name }}</option>
              </select>
            </div>
          </div>
          <div class="k8s-import-actions">
            <button class="btn-import-cancel" @click="showK8sImportModal = false">取消</button>
            <button class="btn-import-submit" @click="importFromK8s" :disabled="importingFromK8s">
              <span v-if="importingFromK8s" class="loading-spinner-sm"></span>
              {{ importingFromK8s ? '导入中...' : '确认导入' }}
            </button>
          </div>
        </div>
      </div>
    </transition>

    <!-- 创建成功拓扑页 -->
    <transition name="slide-fade">
      <div v-if="showSuccessTopology" class="success-topology">
        <div class="topology-card">
          <div class="topology-success-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
              <polyline points="20 6 9 17 4 12"/>
            </svg>
          </div>
          <h2>流水线创建成功！</h2>
          <p>已自动配置完整的 CI/CD 流程</p>

          <div v-if="createWarnings.length" class="topology-warnings">
            <h4>⚠️ 注意事项</h4>
            <ul>
              <li v-for="(w, i) in createWarnings" :key="i">{{ w }}</li>
            </ul>
          </div>

          <div class="topology-flow">
            <div class="topology-node">
              <div class="topology-node-icon git">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/>
                  <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/>
                </svg>
              </div>
              <span class="topology-node-label">Git 仓库</span>
              <span class="topology-node-value">{{ pipelineData.git_repo ? pipelineData.git_repo.split('/').pop().replace('.git', '') : '-' }}</span>
            </div>

            <span class="topology-arrow">→</span>

            <div class="topology-node">
              <div class="topology-node-icon jenkins">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <rect x="2" y="3" width="20" height="14" rx="2" ry="2"/>
                  <line x1="8" y1="21" x2="16" y2="21"/>
                  <line x1="12" y1="17" x2="12" y2="21"/>
                </svg>
              </div>
              <span class="topology-node-label">Jenkins 构建</span>
              <span class="topology-node-value">{{ serviceTypeOptions.find(o => o.value === pipelineData.language_type)?.label || '-' }}</span>
            </div>

            <span class="topology-arrow">→</span>

            <div class="topology-node">
              <div class="topology-node-icon registry">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <rect x="2" y="2" width="20" height="8" rx="2" ry="2"/>
                  <rect x="2" y="14" width="20" height="8" rx="2" ry="2"/>
                  <line x1="6" y1="6" x2="6.01" y2="6"/>
                  <line x1="6" y1="18" x2="6.01" y2="18"/>
                </svg>
              </div>
              <span class="topology-node-label">镜像仓库</span>
              <span class="topology-node-value">{{ pipelineData.image_repo ? pipelineData.image_repo.split('/').pop() : '-' }}</span>
            </div>

            <span class="topology-arrow">→</span>

            <div class="topology-node">
              <div class="topology-node-icon k8s">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M12 2L2 7l10 5 10-5-10-5z"/>
                  <path d="M2 17l10 5 10-5"/>
                  <path d="M2 12l10 5 10-5"/>
                </svg>
              </div>
              <span class="topology-node-label">K8s 部署</span>
              <span class="topology-node-value">{{ pipelineData.target_namespace || '-' }}</span>
            </div>
          </div>

          <div class="topology-actions">
            <button class="btn-topology-secondary" @click="goToPipelineList">返回列表</button>
            <button class="btn-topology-primary" @click="viewPipelineDetail">查看流水线详情</button>
          </div>
        </div>
      </div>
    </transition>
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
  validateResourceConfig,
  discoverFromK8s
} from '@/api/cicd.js'
import { checkPipelineName, getJenkinsConfig } from '@/api/platform/pipeline.js'
import { getClusterList } from '@/api/cluster.js'
import { getNamespaces } from '@/api/namespace.js'
import namespaceApi from '@/api/cluster/config/namespace'
import deploymentsApi from '@/api/cluster/workloads/deployments'
import statefulsetsApi from '@/api/cluster/workloads/statefulsets'
import daemonsetsApi from '@/api/cluster/workloads/daemonsets'
import cronjobsApi from '@/api/cluster/workloads/cronjobs'
import jobsApi from '@/api/cluster/workloads/jobs'
import podsApi from '@/api/cluster/workloads/pods'
import { useClusterStore } from '@/stores/cluster'
import permissionStore from '@/stores/permission'

export default {
  name: 'PipelineCreate',
  setup() {
    const router = useRouter()
    const route = useRoute()
    const pipelineId = route.params.id
    const isEdit = !!pipelineId

    // 步骤定义：快速发布 3 步（应用信息 + 构建配置 + 自动部署）
    const steps = ref([
      { id: 'basic', title: '应用信息', description: '名称、语言和仓库' },
      { id: 'jenkins', title: '构建配置', description: '镜像仓库和构建参数' },
      { id: 'auto-deploy', title: '自动部署', description: 'K8s 目标配置' }
    ])

    const currentStep = ref(0)
    const templates = ref([])
    const selectedTemplateId = ref('')
    const showEnvVars = ref(true)
    const showResources = ref(false)
    const quickMode = ref(false) // 已废弃，保留变量避免引用报错
    const showJenkinsAdvanced = ref(false) // Jenkins高级配置默认折叠
    const showDescription = ref(false) // 描述字段默认折叠

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
      { value: 'java', label: 'Java', color: '#f89820', badge: '推荐', badgeClass: 'rec' },
      { value: 'go', label: 'Go', color: '#00add8', badge: '简单', badgeClass: 'easy' },
      { value: 'frontend', label: 'Node.js', color: '#339933' },
      { value: 'python', label: 'Python', color: '#3776ab', badge: '简单', badgeClass: 'easy' },
      { value: 'custom', label: '自定义', color: '#8c8c8c' }
    ])
    const selectedServiceType = ref('go')
    
    // 快速模板（顶部栏使用）
    const quickTemplates = computed(() => serviceTypeOptions.value)
    
    // 仓库检测相关
    const detectingRepo = ref(false)
    const repoDetectionResult = ref(null)
    
    // Dockerfile 预览相关
    const showDockerfilePreview = ref(false)
    const copiedDockerfile = ref(false)
    
    // 镜像预览计算属性
    const imagePreview = computed(() => {
      if (!pipelineData.value.image_repo) {
        return { registry: '', project: '', app: '', tag: '' }
      }
      
      // 解析镜像仓库地址
      const parts = pipelineData.value.image_repo.split('/')
      let registry = ''
      let project = ''
      let app = ''
      
      if (parts.length >= 3) {
        registry = parts[0]
        project = parts[1]
        app = parts[2]
      } else if (parts.length === 2) {
        registry = 'registry.hub.docker.com'
        project = parts[0]
        app = parts[1]
      } else {
        registry = 'registry.hub.docker.com'
        project = 'library'
        app = parts[0]
      }
      
      // 生成镜像标签
      let tag = ''
      if (pipelineData.value.image_tag) {
        tag = pipelineData.value.image_tag
      } else {
        const branch = pipelineData.value.git_branch || 'main'
        const date = new Date().toISOString().slice(0, 10).replace(/-/g, '')
        const shortHash = Math.random().toString(36).slice(2, 8)
        tag = `${date}-${branch}-${shortHash}`
      }
      
      return { registry, project, app, tag }
    })
    
    // Dockerfile 模式标签
    const dockerfileModeLabel = computed(() => {
      return '平台统一生成'
    })
    
    // Dockerfile 内容生成
    const dockerfileContent = computed(() => {
      const lang = pipelineData.value.language_type
      
      // 统一使用平台生成的 Dockerfile
      // 根据语言生成Dockerfile
      const dockerfiles = {
        java: `# 多阶段构建：Java Spring Boot 应用\n# 阶段1：构建\nFROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/maven:3.9-eclipse-temurin-17 AS builder\nWORKDIR /app\nCOPY pom.xml .\nRUN mvn dependency:go-offline -B\nCOPY src ./src\nRUN mvn clean package -DskipTests\n\n# 阶段2：运行（jammy 基底，兼容 snappy-java/Netty 等原生库）\nFROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/library/eclipse-temurin:17-jre-jammy\nWORKDIR /app\nCOPY --from=builder /app/target/*.jar app.jar\nRUN groupadd -r appgroup && useradd -r -g appgroup -d /app appuser\nRUN mkdir -p /app/logs && chown -R appuser:appgroup /app\nUSER appuser\nEXPOSE 8080\nENV JAVA_OPTS="-XX:MaxRAMPercentage=75.0 -XX:+UseG1GC -XX:+HeapDumpOnOutOfMemoryError -XX:HeapDumpPath=/app/logs -Djava.security.egd=file:/dev/./urandom"\nENTRYPOINT ["sh", "-c", "java $JAVA_OPTS -jar app.jar"]`,
        go: `# 多阶段构建：Go 应用\n# 阶段1：构建\nFROM golang:1.24-alpine AS builder\nWORKDIR /app\nCOPY go.mod go.sum ./\nRUN go mod download\nCOPY . .\nRUN CGO_ENABLED=0 GOOS=linux go build -o /app/main .\n\n# 阶段2：运行\nFROM alpine:3.19\nWORKDIR /app\nCOPY --from=builder /app/main .\nRUN addgroup -S appgroup && adduser -S appuser -G appgroup\nUSER appuser\nEXPOSE 8080\nENTRYPOINT ["./main"]`,
        frontend: `# Node.js 前端应用\n# 阶段1：构建\nFROM node:18-alpine AS builder\nWORKDIR /app\nCOPY package*.json ./\nRUN npm ci\nCOPY . .\nRUN npm run build\n\n# 阶段2：Nginx 运行\nFROM nginx:1.25-alpine\nCOPY --from=builder /app/dist /usr/share/nginx/html\nCOPY nginx.conf /etc/nginx/conf.d/default.conf\nEXPOSE 80\nCMD ["nginx", "-g", "daemon off;"]`,
        python: `# Python 应用\nFROM python:3.11-slim\nWORKDIR /app\nCOPY requirements.txt .\nRUN pip install --no-cache-dir -r requirements.txt\nCOPY . .\nRUN addgroup --system appgroup && adduser --system --group appuser\nUSER appuser\nEXPOSE 8000\nCMD ["gunicorn", "--bind", "0.0.0.0:8000", "app:app"]`,
        custom: `# 自定义应用\n# 请根据实际需求修改此 Dockerfile\nFROM ubuntu:22.04\nWORKDIR /app\nCOPY . .\nRUN echo "请添加构建命令"\nEXPOSE 8080\nCMD ["echo", "请替换为实际启动命令"]`
      }
      
      return dockerfiles[lang] || dockerfiles.custom
    })
    
    // 从 K8s 导入弹窗
    const showK8sImportModal = ref(false)
    const k8sImportForm = ref({ cluster_id: 0, namespace: '', deployment: '' })
    const k8sImportClusters = ref([])
    const k8sImportNamespaces = ref([])
    const k8sImportDeployments = ref([])
    const importingFromK8s = ref(false)

    // 默认镜像仓库前缀（从 Jenkins 配置加载）
    const defaultImageRegistry = ref('')
    
    // 确认步骤
    const showConfirmStep = ref(false)
    
    // 创建成功拓扑
    const showSuccessTopology = ref(false)
    const createdPipelineId = ref(0)
    
    // Dockerfile 构建策略模式：统一使用平台生成
    const dockerfileMode = ref('platform')
    
    // 镜像标签策略：latest / auto / custom
    const tagStrategy = ref('latest')
    const setTagStrategy = (strategy) => {
      tagStrategy.value = strategy
      if (strategy === 'latest') {
        pipelineData.value.image_tag = 'latest'
      } else if (strategy === 'auto') {
        pipelineData.value.image_tag = ''
      }
      // custom 策略不自动清空，用户手动输入
    }
    
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
      { value: 'DaemonSet', label: 'DaemonSet', description: '守护进程' },
      { value: 'CronJob', label: 'CronJob', description: '定时任务' },
      { value: 'Job', label: 'Job', description: '一次性任务' },
      { value: 'Pod', label: 'Pod', description: '独立 Pod' }
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
      image_tag: 'latest',  // 镜像标签，默认 latest；留空则 Jenkins 自动生成
      java_version: '17',   // Java 版本选择（仅 language_type=java 时生效）
      build_dir: '',          // 构建目录（多模块 Maven 项目指定子模块目录，仅 Java 生效）
      maven_private_repo_url: '', // 私有 Maven 仓库地址（用于拉取公司内部依赖包）
      skip_tests: false,    // 跳过单元测试
      dockerfile_path: '',  // Dockerfile 路径（空则自动生成）
      git_credential_id: '',  // Jenkins 中配置的 Git 凭证 ID
      env_vars: [],
      enable_sonar: false,  // SonarQube 代码质量扫描开关（Java 项目默认启用）
      enable_artifact_upload: false,  // 制品上传到平台制品库开关

      deploy_config: {
        replicas: 1,
        strategy: 'rollingUpdate',
        resources: {
          limits: { cpu: '200m', memory: '256Mi' },
          requests: { cpu: '100m', memory: '128Mi' }
        }
      },
      // 自动部署配置
      auto_deploy: true,
      target_cluster_id: 0,
      target_namespace: '',
      target_workload_kind: 'Deployment',
      target_workload_name: '',
      target_container: '',
      deploy_env: 'dev',
      require_approval: false,
      // 发布联动告警静默
      enable_deploy_silence: false,
      silence_buffer_minutes: 10,
      silence_severities: 'warning,info'
    })

    const submitting = ref(false)

    // 应用名称实时校验状态
    const nameChecking = ref(false)      // 正在检查中
    const nameAvailable = ref(null)      // null=未检查, true=可用, false=已占用
    const nameCheckMsg = ref('')         // 提示信息
    const createWarnings = ref([])       // 创建成功后的警告信息

    // 应用名称可用性检查（模板：name blur 时触发）
    const checkName = async () => {
      const name = pipelineData.value.name.trim()
      if (!name) { nameAvailable.value = null; nameCheckMsg.value = ''; return }
      nameChecking.value = true
      nameAvailable.value = null
      try {
        const res = await checkPipelineName(name, isEdit ? Number(pipelineId) : 0)
        if (res.code === 0) {
          nameAvailable.value = res.data?.available !== false
          nameCheckMsg.value = nameAvailable.value ? '' : '该应用名称已存在，请更换一个'
        }
      } catch (e) {
        console.warn('名称检查失败:', e)
      } finally {
        nameChecking.value = false
      }
    }

    // SonarQube 质量门禁配置
    const sonarConfig = ref({
      coverage: 80,          // 代码覆盖率阈值 %
      newBugs: 0,            // 新增 Bug 最大允许数
      codeSmells: 10,        // 代码异味最大允许数
      vulnerabilities: 0,    // 安全漏洞最大允许数
      duplications: 3,       // 代码重复率阈值 %
      gateAction: 'block',   // 门禁失败策略: block | warn | skip
    })

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
            alert('请输入应用名称')
            return false
          }
          if (nameAvailable.value === false) {
            alert('应用名称已存在，请更换一个名称')
            return false
          }
          if (!pipelineData.value.git_repo.trim()) {
            alert('请输入 Git 仓库地址')
            return false
          }
          if (!pipelineData.value.git_branch.trim()) {
            alert('请输入分支名称')
            return false
          }
          break
        case 1:
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
        case 2:
          // 自动部署配置校验
          if (pipelineData.value.auto_deploy) {
            if (!pipelineData.value.target_cluster_id) {
              alert('请选择目标集群')
              return false
            }
            if (!pipelineData.value.target_namespace) {
              alert('请选择目标命名空间')
              return false
            }
            if (!pipelineData.value.target_workload_name) {
              alert('请填写或选择工作负载名称（即 K8s 中的 Deployment 名称）')
              return false
            }
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

    // 生成默认分支列表（后端API未实现时显示空列表，不使用假数据）
    const generateMockBranches = (repoUrl) => {
      // 返回空列表，用户可手动输入分支名
      return []
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
      const promotedKeys = ['IMAGE_REPO', 'IMAGE_TAG', 'GIT_CREDENTIAL_ID', 'SKIP_TESTS', 'DOCKERFILE_PATH', 'JAVA_VERSION', 'BUILD_DIR', 'MAVEN_PRIVATE_REPO_URL']
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
            const promotedKeys = ['IMAGE_REPO', 'IMAGE_TAG', 'SKIP_TESTS', 'DOCKERFILE_PATH', 'GIT_CREDENTIAL_ID', 'JAVA_VERSION', 'BUILD_DIR', 'MAVEN_PRIVATE_REPO_URL']
            const filteredEnvVars = envArr.filter(e => !promotedKeys.includes(e.name))
            // 回显 Dockerfile 策略模式（统一使用平台生成）
            dockerfileMode.value = 'platform'
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
              image_tag: getEnv('IMAGE_TAG', ''),
              java_version: getEnv('JAVA_VERSION', '17'),
              build_dir: getEnv('BUILD_DIR', ''),
              maven_private_repo_url: getEnv('MAVEN_PRIVATE_REPO_URL', ''),
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
              require_approval: data.require_approval || false,
              // 发布联动告警静默
              enable_deploy_silence: data.enable_deploy_silence || false,
              silence_buffer_minutes: data.silence_buffer_minutes || 10,
              silence_severities: data.silence_severities || 'warning,info'
            }
            // 如果有自动部署配置，加载相关数据
            if (data.auto_deploy && data.target_cluster_id) {
              await loadClusters()
              await loadNamespaces()
              await loadWorkloads()
            }
            // 回显标签策略
            const loadedTag = pipelineData.value.image_tag
            if (loadedTag === 'latest') {
              tagStrategy.value = 'latest'
            } else if (!loadedTag) {
              tagStrategy.value = 'auto'
            } else {
              tagStrategy.value = 'custom'
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

    // 极简模式快速提交
    const quickSubmit = async () => {
      if (!pipelineData.value.name || !pipelineData.value.git_repo) {
        alert('请填写应用名称和 Git 仓库地址')
        return
      }
      if (nameAvailable.value === false) {
        alert('应用名称已存在，请更换一个名称')
        return
      }
      // 自动部署开启时，检查部署配置完整性
      if (pipelineData.value.auto_deploy) {
        if (!pipelineData.value.target_cluster_id) {
          alert('请选择目标集群')
          return
        }
        if (!pipelineData.value.target_namespace) {
          alert('请选择目标命名空间')
          return
        }
        if (!pipelineData.value.target_workload_name) {
          alert('请填写或选择工作负载名称')
          return
        }
      }
      try {
        submitting.value = true
        // 构建最小化提交数据，后端自动推导其余配置
        const submitData = {
          name: pipelineData.value.name,
          git_repo: pipelineData.value.git_repo,
          git_branch: pipelineData.value.git_branch || 'main',
          language_type: pipelineData.value.language_type || 'go',
          auto_deploy: pipelineData.value.auto_deploy,
          target_cluster_id: pipelineData.value.target_cluster_id || 0,
          target_namespace: pipelineData.value.target_namespace || '',
          target_workload_kind: pipelineData.value.target_workload_kind || 'Deployment',
          target_workload_name: pipelineData.value.target_workload_name || '',
          target_container: pipelineData.value.target_container || pipelineData.value.target_workload_name || '',
          env_vars: []
        }
        // 注入 IMAGE_REPO
        if (pipelineData.value.image_repo) {
          submitData.env_vars.push({ name: 'IMAGE_REPO', value: pipelineData.value.image_repo })
        }
        // 注入 IMAGE_TAG（留空则 Jenkins 自动生成）
        if (pipelineData.value.image_tag) {
          submitData.env_vars.push({ name: 'IMAGE_TAG', value: pipelineData.value.image_tag })
        }
        // 注入 JAVA_VERSION（仅 Java 项目）
        if (pipelineData.value.language_type === 'java') {
          submitData.env_vars.push({ name: 'JAVA_VERSION', value: pipelineData.value.java_version || '17' })
          // 注入 BUILD_DIR（多模块 Maven 项目）
          if (pipelineData.value.build_dir) {
            submitData.env_vars.push({ name: 'BUILD_DIR', value: pipelineData.value.build_dir })
          }
          // 注入私有 Maven 仓库地址
          if (pipelineData.value.maven_private_repo_url) {
            submitData.env_vars.push({ name: 'MAVEN_PRIVATE_REPO_URL', value: pipelineData.value.maven_private_repo_url })
          }
        }
        const response = await createPipeline(submitData)
        if (response.code === 0) {
          // 处理警告信息
          if (response.data?.warnings?.length) {
            const warnMsg = '应用创建成功！但有以下注意事项：\n\n' +
              response.data.warnings.map((w, i) => `${i + 1}. ${w}`).join('\n') +
              '\n\n建议确认后再进行构建。'
            alert(warnMsg)
          } else {
            alert('应用创建成功！')
          }
          router.push('/cicd/pipelines')
        } else {
          alert(response.msg || '创建失败')
        }
      } catch (error) {
        console.error('快速创建失败:', error)
        alert(error.msg || '创建流水线失败')
      } finally {
        submitting.value = false
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
        if (submitData.image_tag) injectEnv('IMAGE_TAG', submitData.image_tag)
        injectEnv('SKIP_TESTS', submitData.skip_tests ? 'true' : 'false')
        // Java 版本注入（仅 Java 项目）
        if (submitData.language_type === 'java') {
          injectEnv('JAVA_VERSION', submitData.java_version || '17')
          // BUILD_DIR 注入（多模块 Maven 项目指定子模块）
          if (submitData.build_dir) {
            injectEnv('BUILD_DIR', submitData.build_dir)
          }
          // 私有 Maven 仓库地址注入
          if (submitData.maven_private_repo_url) {
            injectEnv('MAVEN_PRIVATE_REPO_URL', submitData.maven_private_repo_url)
          }
        }
        // Dockerfile 策略：统一使用平台生成
        injectEnv('DOCKERFILE_PATH', '__PLATFORM_GENERATE__')
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
            envVars[gateIdx].value = sonarConfig.value.gateAction === 'skip' ? 'false' : 'true'
          } else {
            envVars.push({ name: 'SONAR_QUALITY_GATE', value: sonarConfig.value.gateAction === 'skip' ? 'false' : 'true' })
          }
          // 注入质量门禁参数
          injectEnv('SONAR_COVERAGE_THRESHOLD', String(sonarConfig.value.coverage))
          injectEnv('SONAR_NEW_BUGS_MAX', String(sonarConfig.value.newBugs))
          injectEnv('SONAR_CODE_SMELLS_MAX', String(sonarConfig.value.codeSmells))
          injectEnv('SONAR_VULNERABILITIES_MAX', String(sonarConfig.value.vulnerabilities))
          injectEnv('SONAR_DUPLICATIONS_MAX', String(sonarConfig.value.duplications))
          injectEnv('SONAR_GATE_ACTION', sonarConfig.value.gateAction)
        } else {
          // 关闭 SonarQube：移除相关环境变量，避免残留导致 Jenkins 仍执行扫描
          const sonarKeys = ['ENABLE_SONAR', 'SONAR_QUALITY_GATE', 'SONAR_COVERAGE_THRESHOLD', 'SONAR_NEW_BUGS_MAX', 'SONAR_CODE_SMELLS_MAX', 'SONAR_VULNERABILITIES_MAX', 'SONAR_DUPLICATIONS_MAX', 'SONAR_GATE_ACTION']
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
        delete submitData.image_tag
        delete submitData.skip_tests
        delete submitData.dockerfile_path
        delete submitData.git_credential_id
        delete submitData.java_version
        delete submitData.build_dir
        delete submitData.maven_private_repo_url

        // 确保容器名称有值（自动部署时必须）
        if (submitData.auto_deploy && !submitData.target_container && submitData.target_workload_name) {
          submitData.target_container = submitData.target_workload_name
        }

        if (isEdit) {
          response = await updatePipeline({
            id: parseInt(pipelineId),
            ...submitData
          })
        } else {
          response = await createPipeline(submitData)
        }

        if (response.code === 0) {
          if (!isEdit) {
            // 新建成功：记录 ID 并显示拓扑页
            createdPipelineId.value = response.data?.id || 0
            if (response.data?.warnings?.length) {
              createWarnings.value = response.data.warnings
            }
            showTopology()
          } else {
            // 编辑模式：保存成功后跳转到流水线详情的执行阶段页
            if (response.data?.warnings?.length) {
              const warnMsg = '更新成功！但有以下注意事项：\n\n' +
                response.data.warnings.map((w, i) => `${i + 1}. ${w}`).join('\n') +
                '\n\n建议确认后再进行构建。'
              alert(warnMsg)
            } else {
              alert('更新流水线成功')
            }
            router.push(`/cicd/pipelines/${pipelineId}?tab=stages`)
          }
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
          // 后端已做权限过滤，直接展示所有返回的集群
          clusters.value = res.data.list || []
          // 快速模式：自动选择第一个集群（开发环境只有一个集群）
          if (!isEdit && clusters.value.length > 0 && !pipelineData.value.target_cluster_id) {
            pipelineData.value.target_cluster_id = clusters.value[0].id
            // 等待命名空间加载完成
            await loadNamespaces()
          }
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
        // 设置当前集群（必须在 API 请求前完成，确保 X-Cluster-ID Header 正确注入）
        const cluster = clusters.value.find(c => c.id === pipelineData.value.target_cluster_id)
        if (cluster) {
          clusterStore.setCurrent(cluster)
        }
        
        const res = await namespaceApi.list({ page: 1, limit: 1000 })
        if (res.code === 0 && res.data) {
          // 直接展示所有命名空间，所有用户（开发、运维、管理员）都可以选择
          // 部署权限由 K8s 和审批流程控制，前端不做限制
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
          case 'CronJob':
            res = await cronjobsApi.list({ namespace: ns, page: 1, limit: 1000 })
            break
          case 'Job':
            res = await jobsApi.list({ namespace: ns, page: 1, limit: 1000 })
            break
          case 'Pod':
            res = await podsApi.list({ namespace: ns, page: 1, limit: 1000 })
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
    
    // 工作负载变化 - 自动填充容器名称
    const onWorkloadChange = () => {
      const selectedName = pipelineData.value.target_workload_name
      if (!selectedName) {
        pipelineData.value.target_container = ''
        return
      }
      // 从已加载的工作负载列表中找到选中项，获取第一个容器名称
      const workload = workloads.value.find(w => w.name === selectedName)
      if (workload && workload.containers && workload.containers.length > 0) {
        // 使用工作负载的第一个容器名称
        pipelineData.value.target_container = workload.containers[0]
      } else {
        // 回退：容器名默认使用工作负载名称（与后端智能默认值一致）
        pipelineData.value.target_container = selectedName
      }
    }
    
    // 选择部署环境
    const selectDeployEnv = (env) => {
      pipelineData.value.deploy_env = env
      // 预发环境和生产环境强制开启审批（与审批策略配置一致）
      if (env === 'prod' || env === 'staging') {
        pipelineData.value.require_approval = true
      }
      // dev/test 不强制重置，保留管理员的手动设置（管理员可自行决定开发/测试环境是否需要审批）
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

    // 快速模式：应用名称自动同步到工作负载名称 + 镜像地址（仅新建时）
    watch(() => pipelineData.value.name, (newName) => {
      if (!isEdit && newName) {
        // 自动填充工作负载名称（仅当用户未手动修改过时）
        if (!pipelineData.value.target_workload_name || pipelineData.value.target_workload_name === pipelineData.value._lastAutoName) {
          pipelineData.value.target_workload_name = newName
          pipelineData.value._lastAutoName = newName
        }
        // 自动拼接镜像仓库地址（仅当用户未手动修改过时）
        if (defaultImageRegistry.value) {
          const expectedOld = pipelineData.value._lastAutoImageRepo || ''
          if (!pipelineData.value.image_repo || pipelineData.value.image_repo === expectedOld) {
            const autoRepo = `${defaultImageRegistry.value}/${newName}`
            pipelineData.value.image_repo = autoRepo
            pipelineData.value._lastAutoImageRepo = autoRepo
          }
        }
      }
    })

    // ==================== 新增功能方法 ====================

    // 快速模板应用
    const applyQuickTemplate = (type) => {
      pipelineData.value.language_type = type
      selectedServiceType.value = type
      pipelineData.value.enable_sonar = (type === 'java')
      selectedResourceTemplate.value = ''
      loadResourceTemplates()

      // 自动填充推荐环境变量
      const defaults = languageEnvDefaults[type] || languageEnvDefaults.custom
      const allDefaultKeys = new Set()
      Object.values(languageEnvDefaults).forEach(arr => arr.forEach(d => allDefaultKeys.add(d.name)))
      const userCustom = pipelineData.value.env_vars.filter(e => !allDefaultKeys.has(e.name))
      const existingImageRepo = pipelineData.value.env_vars.find(e => e.name === 'IMAGE_REPO')
      if (existingImageRepo && existingImageRepo.value && existingImageRepo.value !== 'harbor.example.com/project/app-name') {
        pipelineData.value.image_repo = existingImageRepo.value
      }
      const promotedKeys = ['IMAGE_REPO', 'IMAGE_TAG', 'GIT_CREDENTIAL_ID', 'SKIP_TESTS', 'DOCKERFILE_PATH', 'JAVA_VERSION', 'BUILD_DIR', 'MAVEN_PRIVATE_REPO_URL']
      const newEnvVars = defaults.filter(d => !promotedKeys.includes(d.name)).map(d => ({ name: d.name, value: d.value }))
      pipelineData.value.env_vars = [...newEnvVars, ...userCustom]
    }

    // 仓库检测（替代原来的获取分支）
    const detectRepo = async () => {
      const repoUrl = pipelineData.value.git_repo.trim()
      if (!repoUrl) {
        alert('请先输入 Git 仓库地址')
        return
      }

      detectingRepo.value = true
      repoDetectionResult.value = null

      try {
        // 调用分支获取接口
        const response = await getGitBranches(repoUrl)
        if (response.code === 0 && response.data) {
          branches.value = response.data.branches || response.data || []
          lastFetchedRepo.value = repoUrl

          // 自动选择默认分支
          if (!pipelineData.value.git_branch || !branches.value.find(b => b.name === pipelineData.value.git_branch)) {
            const defaultBranch = branches.value.find(b => b.isDefault)
            if (defaultBranch) {
              pipelineData.value.git_branch = defaultBranch.name
            } else if (branches.value.length > 0) {
              const mainBranch = branches.value.find(b => b.name === 'main' || b.name === 'master')
              pipelineData.value.git_branch = mainBranch ? mainBranch.name : branches.value[0].name
            }
          }

          // 检测仓库类型
          let repoType = 'Git'
          if (repoUrl.includes('gitlab')) repoType = 'GitLab'
          else if (repoUrl.includes('github')) repoType = 'GitHub'
          else if (repoUrl.includes('gitee')) repoType = 'Gitee'
          else if (repoUrl.includes('bitbucket')) repoType = 'Bitbucket'

          // 检测语言和构建工具（根据仓库 URL 中的项目名推测）
          const repoName = repoUrl.split('/').pop().replace('.git', '').toLowerCase()
          let language = '未知'
          let buildTool = '未知'

          if (repoName.includes('spring') || repoName.includes('java')) {
            language = 'Java'
            buildTool = 'Maven'
            pipelineData.value.language_type = 'java'
            selectedServiceType.value = 'java'
          } else if (repoName.includes('go') || repoName.includes('golang')) {
            language = 'Go'
            buildTool = 'Go Modules'
            pipelineData.value.language_type = 'go'
            selectedServiceType.value = 'go'
          } else if (repoName.includes('node') || repoName.includes('vue') || repoName.includes('react') || repoName.includes('frontend')) {
            language = 'Node.js'
            buildTool = 'NPM'
            pipelineData.value.language_type = 'frontend'
            selectedServiceType.value = 'frontend'
          } else if (repoName.includes('python') || repoName.includes('flask') || repoName.includes('django')) {
            language = 'Python'
            buildTool = 'pip'
            pipelineData.value.language_type = 'python'
            selectedServiceType.value = 'python'
          }

          // Dockerfile 检测结果（后端未实现检测时默认为未知）
          const hasDockerfile = false

          repoDetectionResult.value = {
            repoType,
            defaultBranch: pipelineData.value.git_branch,
            language,
            buildTool,
            hasDockerfile
          }
        } else {
          console.warn('获取分支失败，使用默认分支列表')
          branches.value = generateMockBranches(repoUrl)
          lastFetchedRepo.value = repoUrl

          repoDetectionResult.value = {
            repoType: repoUrl.includes('gitlab') ? 'GitLab' : repoUrl.includes('github') ? 'GitHub' : 'Git',
            defaultBranch: 'main',
            language: '未知',
            buildTool: '未知',
            hasDockerfile: false
          }
        }
      } catch (error) {
        console.error('检测仓库失败:', error)
        branches.value = generateMockBranches(repoUrl)
        lastFetchedRepo.value = repoUrl

        repoDetectionResult.value = {
          repoType: 'Git',
          defaultBranch: 'main',
          language: '未知',
          buildTool: '未知',
          hasDockerfile: false
        }
      } finally {
        detectingRepo.value = false
      }
    }

    // 复制 Dockerfile 内容
    const copyDockerfile = () => {
      navigator.clipboard.writeText(dockerfileContent.value).then(() => {
        copiedDockerfile.value = true
        setTimeout(() => {
          copiedDockerfile.value = false
        }, 2000)
      }).catch(err => {
        console.error('复制失败:', err)
        alert('复制失败，请手动选择复制')
      })
    }

    // 从 K8s 导入
    const importFromK8s = async () => {
      if (!k8sImportForm.value.cluster_id || !k8sImportForm.value.namespace || !k8sImportForm.value.deployment) {
        alert('请选择集群、命名空间和 Deployment')
        return
      }

      importingFromK8s.value = true
      try {
        // 调用 Discover API 获取完整信息
        const res = await discoverFromK8s({
          cluster_id: k8sImportForm.value.cluster_id,
          namespace: k8sImportForm.value.namespace,
          deployment: k8sImportForm.value.deployment
        })

        if (res.code === 0 && res.data) {
          const data = res.data
          // 自动填充流水线数据
          pipelineData.value.name = data.deployment_name
          pipelineData.value.target_cluster_id = k8sImportForm.value.cluster_id
          pipelineData.value.target_namespace = data.namespace
          pipelineData.value.target_workload_name = data.deployment_name
          pipelineData.value.target_workload_kind = 'Deployment'
          pipelineData.value.target_container = data.primary_container || data.deployment_name
          pipelineData.value.image_repo = data.primary_image_repo || ''
          pipelineData.value.image_tag = data.primary_image_tag || ''
          pipelineData.value.auto_deploy = true

          // 加载集群和命名空间数据（用于展示）
          await loadClusters()
          await loadNamespaces()
          await loadWorkloads()

          showK8sImportModal.value = false
          alert('导入成功！已自动填充应用名称、集群、命名空间、工作负载和镜像仓库信息')
        } else {
          alert('发现失败：' + (res.msg || '未知错误'))
        }
      } catch (error) {
        console.error('从 K8s 导入失败:', error)
        alert('导入失败：' + (error.msg || error.message || '未知错误'))
      } finally {
        importingFromK8s.value = false
      }
    }

    // 加载 K8s 导入所需的集群列表
    const loadK8sImportClusters = async () => {
      if (k8sImportClusters.value.length > 0) return
      try {
        const res = await getClusterList()
        if (res.code === 0) {
          k8sImportClusters.value = res.data.list || res.data || []
        }
      } catch (e) {
        console.error('加载集群列表失败:', e)
      }
    }

    // K8s 导入：集群变化
    const onK8sImportClusterChange = async () => {
      k8sImportForm.value.namespace = ''
      k8sImportForm.value.deployment = ''
      k8sImportNamespaces.value = []
      k8sImportDeployments.value = []
      if (k8sImportForm.value.cluster_id) {
        try {
          // 显式传递 cluster_id，不依赖全局 store
          const res = await getNamespaces(k8sImportForm.value.cluster_id, { page: 1, limit: 1000 })
          if (res.code === 0 && res.data) {
            const list = res.data.list || res.data || []
            k8sImportNamespaces.value = list
          }
        } catch (e) {
          console.error('加载命名空间失败:', e)
        }
      }
    }

    // K8s 导入：命名空间变化
    const onK8sImportNamespaceChange = async () => {
      k8sImportForm.value.deployment = ''
      k8sImportDeployments.value = []
      if (k8sImportForm.value.namespace) {
        try {
          // 临时设置集群上下文，确保 X-Cluster-ID 正确
          const cluster = k8sImportClusters.value.find(c => c.id === k8sImportForm.value.cluster_id)
          if (cluster) {
            clusterStore.setCurrent(cluster)
          }
          const res = await deploymentsApi.list({
            namespace: k8sImportForm.value.namespace,
            page: 1,
            limit: 1000
          })
          if (res.code === 0) {
            k8sImportDeployments.value = res.data.list || res.data || []
          }
        } catch (e) {
          console.error('加载 Deployment 列表失败:', e)
        }
      }
    }

    // 确认创建
    const goToConfirmStep = () => {
      if (validateCurrentStep()) {
        showConfirmStep.value = true
      }
    }

    const backFromConfirm = () => {
      showConfirmStep.value = false
    }

    // 创建成功拓扑
    const showTopology = () => {
      showConfirmStep.value = false
      showSuccessTopology.value = true
    }

    const goToPipelineList = () => {
      router.push('/cicd/pipelines')
    }

    const viewPipelineDetail = () => {
      router.push(`/cicd/pipelines/${createdPipelineId.value}`)
    }

    // 部署预览 computed
    const deployPreview = computed(() => {
      const cluster = clusters.value.find(c => c.id === pipelineData.value.target_cluster_id)
      const imageTag = pipelineData.value.git_branch ?
        `${new Date().toISOString().slice(0, 10).replace(/-/g, '')}-${pipelineData.value.git_branch}-${Math.random().toString(36).slice(2, 8)}` :
        'latest'
      const fullImage = pipelineData.value.image_repo ?
        `${pipelineData.value.image_repo}:${imageTag}` : ''

      return {
        cluster: cluster?.cluster_name || '-',
        namespace: pipelineData.value.target_namespace || '-',
        workload: pipelineData.value.target_workload_name || '-',
        replicas: pipelineData.value.deploy_config?.replicas || 1,
        currentImage: '-',
        newImage: fullImage,
        env: pipelineData.value.deploy_env || 'dev'
      }
    })

    // 切换到第3步（自动部署配置）时，若命名空间列表为空则重新加载（安全兼容）
    watch(currentStep, (newStep) => {
      if (newStep === 2 && namespaces.value.length === 0 && pipelineData.value.target_cluster_id) {
        loadNamespaces()
      }
    })

    onMounted(async () => {
      loadTemplates()
      await loadClusters() // 加载集群列表并自动选择默认集群 + 自动加载命名空间
      // 加载默认镜像仓库前缀
      try {
        const configRes = await getJenkinsConfig()
        if (configRes.code === 0 && configRes.data?.default_image_registry) {
          defaultImageRegistry.value = configRes.data.default_image_registry
        }
      } catch (e) {
        console.warn('加载 Jenkins 配置失败:', e)
      }
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
      sonarConfig,
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
      quickTemplates,
      dockerfileMode,
      dockerfileLangLabel,
      templateFileMap,
      tagStrategy,
      setTagStrategy,
      onServiceTypeChange,
      applyQuickTemplate,
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
      // 极简模式
      quickMode,
      quickSubmit,
      showJenkinsAdvanced,
      showDescription,
      // 应用名称实时校验
      nameChecking,
      nameAvailable,
      nameCheckMsg,
      checkName,
      // 新增功能
      detectingRepo,
      repoDetectionResult,
      detectRepo,
      showDockerfilePreview,
      copiedDockerfile,
      dockerfileModeLabel,
      dockerfileContent,
      imagePreview,
      copyDockerfile,
      showK8sImportModal,
      k8sImportForm,
      k8sImportClusters,
      k8sImportNamespaces,
      k8sImportDeployments,
      importingFromK8s,
      importFromK8s,
      loadK8sImportClusters,
      onK8sImportClusterChange,
      onK8sImportNamespaceChange,
      defaultImageRegistry,
      showConfirmStep,
      goToConfirmStep,
      backFromConfirm,
      showSuccessTopology,
      createdPipelineId,
      showTopology,
      goToPipelineList,
      viewPipelineDetail,
      deployPreview,
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

.approval-hint {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
  padding: 10px 14px;
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
  border: 1px solid #93c5fd;
  border-radius: 8px;
  font-size: 12px;
  color: #1e40af;
}

.approval-hint svg {
  stroke: #3b82f6;
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
/* ==================== 镜像标签策略选择器 ==================== */
.tag-strategy-selector {
  display: flex;
  gap: 10px;
  margin-top: 6px;
}

.tag-strategy-btn {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  padding: 10px 12px;
  border: 2px solid #e2e8f0;
  border-radius: 10px;
  background: white;
  cursor: pointer;
  transition: all 0.2s ease;
  text-align: center;
}

.tag-strategy-btn:hover {
  border-color: #a0c4e8;
  background: #f7fafc;
  transform: translateY(-1px);
  box-shadow: 0 3px 10px rgba(66, 153, 225, 0.1);
}

.tag-strategy-btn.active {
  border-color: #4299e1;
  background: linear-gradient(135deg, #ebf8ff 0%, #f0f9ff 100%);
  box-shadow: 0 3px 12px rgba(66, 153, 225, 0.18);
}

.tag-strategy-icon {
  font-size: 18px;
  line-height: 1;
}

.tag-strategy-label {
  font-size: 13px;
  font-weight: 600;
  color: #2d3748;
}

.tag-strategy-btn.active .tag-strategy-label {
  color: #2b6cb0;
}

.tag-strategy-desc {
  font-size: 11px;
  color: #a0aec0;
  line-height: 1.3;
}

.tag-strategy-btn.active .tag-strategy-desc {
  color: #4a90d9;
}

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

/* ==================== SonarQube 质量门禁配置面板 ==================== */
.sonar-expand-enter-active,
.sonar-expand-leave-active {
  transition: all 0.3s ease;
  overflow: hidden;
}
.sonar-expand-enter-from,
.sonar-expand-leave-to {
  opacity: 0;
  max-height: 0;
  transform: translateY(-8px);
}
.sonar-expand-enter-to,
.sonar-expand-leave-from {
  opacity: 1;
  max-height: 1200px;
}

.sonar-config-panel {
  margin-top: 14px;
  border: 1.5px solid #e0e7ff;
  border-radius: 14px;
  background: linear-gradient(135deg, #fafbff 0%, #f5f7ff 100%);
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(99, 102, 241, 0.06);
}

.silence-config-panel {
  margin-top: 12px;
  padding: 16px 18px;
  border: 1.5px solid #fde68a;
  border-radius: 12px;
  background: linear-gradient(135deg, #fffdf7 0%, #fefce8 100%);
  box-shadow: 0 2px 8px rgba(217, 119, 6, 0.06);
}

.silence-config-panel .form-row {
  display: flex;
  gap: 16px;
}

.silence-config-panel .flex-1 {
  flex: 1;
}

.env-tag.info {
  font-size: 10px;
  font-weight: 600;
  background: #dbeafe;
  color: #1d4ed8;
  padding: 2px 6px;
  border-radius: 4px;
  margin-left: 6px;
}

.sonar-panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 18px;
  background: linear-gradient(135deg, #eef2ff 0%, #e0e7ff 100%);
  border-bottom: 1px solid #e0e7ff;
}

.sonar-panel-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 700;
  color: #3730a3;
}

.sonar-panel-badge {
  font-size: 10px;
  font-weight: 700;
  color: #4f46e5;
  background: rgba(79, 70, 229, 0.1);
  padding: 3px 8px;
  border-radius: 6px;
  letter-spacing: 0.5px;
}

.sonar-metrics-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  padding: 16px 18px;
}

.sonar-metric-card {
  background: white;
  border: 1px solid #e8ecf2;
  border-radius: 12px;
  padding: 14px 16px;
  transition: all 0.2s ease;
}

.sonar-metric-card:hover {
  border-color: #c7d2fe;
  box-shadow: 0 2px 8px rgba(99, 102, 241, 0.08);
}

.sonar-metric-card.full-width {
  grid-column: 1 / -1;
}

.metric-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}

.metric-icon {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.metric-icon.coverage {
  background: linear-gradient(135deg, #ecfdf5 0%, #d1fae5 100%);
  color: #059669;
}

.metric-icon.bugs {
  background: linear-gradient(135deg, #fef2f2 0%, #fee2e2 100%);
  color: #dc2626;
}

.metric-icon.smells {
  background: linear-gradient(135deg, #fffbeb 0%, #fef3c7 100%);
  color: #d97706;
}

.metric-icon.vulnerabilities {
  background: linear-gradient(135deg, #fdf2f8 0%, #fce7f3 100%);
  color: #db2777;
}

.metric-icon.duplications {
  background: linear-gradient(135deg, #f5f3ff 0%, #ede9fe 100%);
  color: #7c3aed;
}

.metric-icon.gate {
  background: linear-gradient(135deg, #eef2ff 0%, #e0e7ff 100%);
  color: #4f46e5;
}

.metric-info {
  min-width: 0;
}

.metric-name {
  font-size: 13px;
  font-weight: 600;
  color: #1e293b;
}

.metric-desc {
  font-size: 11px;
  color: #94a3b8;
  margin-top: 1px;
}

.metric-input-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.metric-slider-wrap {
  flex: 1;
  position: relative;
  height: 24px;
  display: flex;
  align-items: center;
}

.metric-slider {
  width: 100%;
  height: 6px;
  -webkit-appearance: none;
  appearance: none;
  background: transparent;
  outline: none;
  position: relative;
  z-index: 2;
  cursor: pointer;
}

.metric-slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: linear-gradient(135deg, #4f46e5 0%, #6366f1 100%);
  border: 2px solid white;
  box-shadow: 0 2px 6px rgba(79, 70, 229, 0.3);
  cursor: pointer;
}

.metric-slider-track {
  position: absolute;
  left: 0;
  right: 0;
  top: 50%;
  transform: translateY(-50%);
  height: 6px;
  background: #e2e8f0;
  border-radius: 3px;
  overflow: hidden;
  z-index: 1;
}

.metric-slider-fill {
  height: 100%;
  background: linear-gradient(90deg, #6366f1, #818cf8);
  border-radius: 3px;
  transition: width 0.15s ease;
}

.metric-value-box {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
}

.metric-num-input {
  width: 52px;
  height: 32px;
  border: 1.5px solid #e2e8f0;
  border-radius: 8px;
  text-align: center;
  font-size: 13px;
  font-weight: 600;
  color: #1e293b;
  outline: none;
  transition: all 0.15s;
}

.metric-num-input:focus {
  border-color: #6366f1;
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
}

.metric-unit {
  font-size: 12px;
  color: #94a3b8;
  font-weight: 500;
}

.metric-presets {
  display: flex;
  gap: 4px;
  flex: 1;
}

.preset-btn {
  flex: 1;
  padding: 6px 0;
  border: 1.5px solid #e2e8f0;
  border-radius: 7px;
  background: white;
  font-size: 12px;
  font-weight: 600;
  color: #64748b;
  cursor: pointer;
  transition: all 0.15s;
  text-align: center;
}

.preset-btn:hover {
  border-color: #c7d2fe;
  color: #4f46e5;
}

.preset-btn.active {
  background: linear-gradient(135deg, #4f46e5 0%, #6366f1 100%);
  color: white;
  border-color: transparent;
  box-shadow: 0 2px 6px rgba(79, 70, 229, 0.25);
}

/* 门禁策略选择器 */
.gate-strategy-selector {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
  margin-top: 4px;
}

.gate-option {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 14px;
  border: 1.5px solid #e2e8f0;
  border-radius: 10px;
  background: white;
  cursor: pointer;
  transition: all 0.2s ease;
}

.gate-option:hover {
  border-color: #c7d2fe;
  background: #fafbff;
}

.gate-option.active {
  border-color: #6366f1;
  background: linear-gradient(135deg, #eef2ff 0%, #e0e7ff 100%);
  box-shadow: 0 2px 8px rgba(99, 102, 241, 0.12);
}

.gate-option-icon {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.gate-option-icon.block {
  background: #fef2f2;
  color: #dc2626;
}

.gate-option-icon.warn {
  background: #fffbeb;
  color: #d97706;
}

.gate-option-icon.skip {
  background: #f0fdf4;
  color: #16a34a;
}

.gate-option-title {
  font-size: 12px;
  font-weight: 600;
  color: #1e293b;
}

.gate-option-desc {
  font-size: 10.5px;
  color: #94a3b8;
  margin-top: 1px;
}

.sonar-panel-footer {
  padding: 12px 18px;
  border-top: 1px solid #e8ecf2;
  background: #f8fafc;
}

.sonar-tip {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11.5px;
  color: #64748b;
}

@media (max-width: 768px) {
  .sonar-metrics-grid {
    grid-template-columns: 1fr;
  }
  .gate-strategy-selector {
    grid-template-columns: 1fr;
  }
}

/* ==================== 极简模式样式 ==================== */
.quick-mode-toggle {
  display: flex;
  gap: 4px;
  padding: 4px;
  background: #f1f5f9;
  border-radius: 10px;
  margin: 16px 32px 0;
  width: fit-content;
}

.mode-btn {
  padding: 8px 20px;
  border: none;
  background: transparent;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  color: #64748b;
  cursor: pointer;
  transition: all 0.3s;
}

.mode-btn.active {
  background: white;
  color: #1e3a5f;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.quick-create-panel {
  max-width: 680px;
  margin: 24px auto;
  background: white;
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.06);
  padding: 36px 40px;
}

.quick-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 32px;
}

.quick-icon {
  width: 48px;
  height: 48px;
  background: linear-gradient(135deg, #fbbf24, #f59e0b);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.quick-icon svg {
  width: 26px;
  height: 26px;
  color: white;
}

.quick-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 700;
  color: #1e293b;
}

.quick-header p {
  margin: 4px 0 0;
  font-size: 13px;
  color: #94a3b8;
}

.quick-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.quick-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.quick-field label {
  font-size: 13px;
  font-weight: 600;
  color: #475569;
}

.quick-field label .required {
  color: #ef4444;
}

.quick-field label .optional {
  color: #94a3b8;
  font-weight: 400;
}

.quick-field .form-input {
  padding: 10px 14px;
  border: 1.5px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  transition: all 0.2s;
  outline: none;
}

.quick-field .form-input:focus {
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.input-hint {
  font-size: 11.5px;
  color: #94a3b8;
}

.quick-lang-selector {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.lang-chip {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border: 1.5px solid #e2e8f0;
  border-radius: 8px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  color: #475569;
  transition: all 0.2s;
}

.lang-chip:hover {
  border-color: #94a3b8;
}

.lang-chip.selected {
  border-color: #3b82f6;
  background: #eff6ff;
  color: #1d4ed8;
}

/* Java 版本选择器 */
.java-version-selector {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.java-ver-chip {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 18px;
  border: 1.5px solid #e2e8f0;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
  background: #fff;
  position: relative;
}

.java-ver-chip:hover {
  border-color: #f59e0b;
  background: #fffbeb;
}

.java-ver-chip.selected {
  border-color: #f59e0b;
  background: linear-gradient(135deg, #fffbeb, #fef3c7);
  box-shadow: 0 2px 8px rgba(245, 158, 11, 0.15);
}

.java-ver-chip .ver-number {
  font-size: 15px;
  font-weight: 700;
  color: #475569;
}

.java-ver-chip.selected .ver-number {
  color: #d97706;
}

.java-ver-chip .ver-badge {
  font-size: 10px;
  font-weight: 600;
  padding: 1px 5px;
  border-radius: 4px;
  background: #dcfce7;
  color: #16a34a;
}

.java-ver-chip .ver-badge.new {
  background: #dbeafe;
  color: #2563eb;
}

.lang-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.quick-deploy-section {
  padding: 16px;
  background: #f8fafc;
  border-radius: 10px;
  border: 1px solid #e2e8f0;
}

.quick-toggle-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 500;
  color: #475569;
  cursor: pointer;
}

.quick-toggle-label input[type="checkbox"] {
  width: 16px;
  height: 16px;
  accent-color: #3b82f6;
}

.quick-deploy-fields {
  margin-top: 14px;
  padding-top: 14px;
  border-top: 1px solid #e2e8f0;
}

.quick-field-row {
  display: flex;
  gap: 12px;
}

.quick-field.half {
  flex: 1;
}

.quick-auto-hint {
  margin-top: 10px;
  font-size: 11.5px;
  color: #64748b;
  background: #eff6ff;
  padding: 8px 12px;
  border-radius: 6px;
}

.quick-empty-cluster {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 16px;
  background: #fffbeb;
  border: 1px solid #fde68a;
  border-radius: 8px;
  font-size: 13px;
  color: #92400e;
}

.quick-empty-cluster svg {
  flex-shrink: 0;
  color: #d97706;
}

.quick-empty-cluster strong {
  color: #2563eb;
  cursor: pointer;
}

.quick-footer {
  margin-top: 32px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.quick-auto-list {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: #64748b;
}

.auto-badge {
  padding: 2px 8px;
  background: #dcfce7;
  color: #16a34a;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
}

.btn-quick-submit {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 28px;
  background: linear-gradient(135deg, #3b82f6, #2563eb);
  color: white;
  border: none;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
}

.btn-quick-submit:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 6px 20px rgba(59, 130, 246, 0.4);
}

.btn-quick-submit:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-quick-submit svg {
  width: 18px;
  height: 18px;
}

/* ==================== 新增样式：快速模板栏 ==================== */
.template-bar {
  background: white;
  border-bottom: 1px solid #e2e8f0;
  padding: 12px 32px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.04);
}

.template-bar-content {
  display: flex;
  align-items: center;
  gap: 16px;
  max-width: 1400px;
  margin: 0 auto;
}

.template-bar-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 600;
  color: #64748b;
  flex-shrink: 0;
}

.template-bar-label svg {
  width: 16px;
  height: 16px;
}

.template-bar-buttons {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.template-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border: 2px solid #e2e8f0;
  border-radius: 8px;
  background: white;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 13px;
  font-weight: 500;
  color: #475569;
}

.template-btn:hover {
  border-color: #a0c4e8;
  background: #f7fafc;
  transform: translateY(-1px);
}

.template-btn.active {
  border-color: #4299e1;
  background: linear-gradient(135deg, #ebf8ff 0%, #f0f9ff 100%);
  color: #2b6cb0;
  box-shadow: 0 2px 8px rgba(66, 153, 225, 0.15);
}

.tpl-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.tpl-name {
  font-weight: 600;
}

.tpl-badge {
  padding: 1px 6px;
  font-size: 10px;
  font-weight: 700;
  border-radius: 4px;
  letter-spacing: 0.5px;
}

.tpl-badge.rec {
  background: linear-gradient(135deg, #48bb78 0%, #38a169 100%);
  color: white;
}

.tpl-badge.easy {
  background: linear-gradient(135deg, #63b3ed 0%, #4299e1 100%);
  color: white;
}

/* ==================== 新增样式：从 K8s 导入按钮 ==================== */
.btn-header-import {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s;
}

.btn-header-import:hover:not(:disabled) {
  background: linear-gradient(135deg, #5a67d8 0%, #6b46c1 100%);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.btn-header-import svg {
  width: 16px;
  height: 16px;
}

/* ==================== 新增样式：仓库检测按钮 ==================== */
.btn-detect-repo {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 12px 16px;
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
  border: none;
  border-radius: 10px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
  white-space: nowrap;
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
}

.btn-detect-repo:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(16, 185, 129, 0.4);
}

.btn-detect-repo:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none;
}

.btn-detect-repo svg {
  width: 16px;
  height: 16px;
}

/* ==================== 新增样式：仓库检测结果 ==================== */
.repo-detection-result {
  margin-top: 16px;
  padding: 16px;
  background: linear-gradient(135deg, #ecfdf5 0%, #d1fae5 100%);
  border: 1px solid #6ee7b7;
  border-radius: 12px;
  animation: fadeIn 0.3s ease;
}

.repo-result-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 600;
  color: #065f46;
  margin-bottom: 12px;
}

.repo-result-header svg {
  width: 20px;
  height: 20px;
  stroke: #10b981;
}

.repo-result-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 10px;
}

.repo-result-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 10px 12px;
  background: rgba(255,255,255,0.7);
  border-radius: 8px;
}

.repo-result-label {
  font-size: 11px;
  color: #6b7280;
  font-weight: 500;
}

.repo-result-value {
  font-size: 13px;
  color: #1f2937;
  font-weight: 600;
}

.repo-result-value.success {
  color: #059669;
}

.repo-result-value.warning {
  color: #d97706;
}

/* ==================== 新增样式：表单分组区块 ==================== */
.form-section {
  margin-bottom: 24px;
  padding: 20px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 700;
  color: #1e293b;
  margin-bottom: 16px;
  padding-bottom: 8px;
  border-bottom: 2px solid #e2e8f0;
}

.section-title svg {
  stroke: #3b82f6;
}

/* ==================== 新增样式：镜像预览面板 ==================== */
.image-preview-panel {
  padding: 16px;
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
  border: 1px solid #93c5fd;
  border-radius: 12px;
  animation: fadeIn 0.3s ease;
}

.preview-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 12px;
}

.preview-label {
  font-size: 12px;
  font-weight: 600;
  color: #1e40af;
  flex-shrink: 0;
  margin-top: 2px;
}

.preview-image-full {
  font-family: 'SF Mono', 'Monaco', 'Inconsolata', monospace;
  font-size: 13px;
  color: #0f172a;
  background: rgba(255, 255, 255, 0.8);
  padding: 10px 14px;
  border-radius: 8px;
  border: 1px solid #bfdbfe;
  word-break: break-all;
  line-height: 1.6;
}

.image-registry {
  color: #7c3aed;
  font-weight: 600;
}

.image-path {
  color: #059669;
}

.image-name {
  color: #dc2626;
  font-weight: 600;
}

.image-tag {
  color: #d97706;
  font-weight: 600;
}

.preview-tags {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: rgba(255, 255, 255, 0.6);
  border-radius: 6px;
  font-size: 12px;
}

.tag-hint {
  color: #64748b;
  font-weight: 500;
}

.preview-tags code {
  font-family: 'SF Mono', 'Monaco', 'Inconsolata', monospace;
  color: #1e40af;
  font-weight: 600;
  background: rgba(255, 255, 255, 0.8);
  padding: 2px 8px;
  border-radius: 4px;
}

.preview-badge {
  display: inline-block;
  padding: 2px 8px;
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  color: white;
  font-size: 10px;
  font-weight: 700;
  border-radius: 4px;
  margin-left: 8px;
  letter-spacing: 0.5px;
}

.optional-badge {
  display: inline-block;
  padding: 2px 8px;
  background: #f1f5f9;
  color: #64748b;
  font-size: 10px;
  font-weight: 600;
  border-radius: 4px;
  margin-left: 8px;
  letter-spacing: 0.3px;
}

/* ==================== 新增样式：Dockerfile 预览面板 ==================== */
.btn-view-dockerfile {
  margin-left: auto;
  padding: 4px 12px;
  background: #f1f5f9;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  color: #475569;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-view-dockerfile:hover {
  background: #e2e8f0;
  border-color: #94a3b8;
  color: #1e293b;
}

.dockerfile-preview-panel {
  margin-top: 12px;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  overflow: hidden;
  background: white;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.dockerfile-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: #f8fafc;
  border-bottom: 1px solid #e2e8f0;
}

.dockerfile-type {
  display: flex;
  align-items: center;
  gap: 8px;
}

.type-badge {
  padding: 3px 10px;
  font-size: 11px;
  font-weight: 700;
  border-radius: 6px;
  letter-spacing: 0.5px;
}

.type-badge.auto {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
}

.type-badge.project {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  color: white;
}

.type-badge.platform {
  background: linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%);
  color: white;
}

.type-lang {
  font-size: 12px;
  font-weight: 600;
  color: #475569;
}

.btn-copy-dockerfile {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: white;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  color: #475569;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-copy-dockerfile:hover {
  background: #f1f5f9;
  border-color: #94a3b8;
  color: #1e293b;
}

.btn-copy-dockerfile svg {
  width: 14px;
  height: 14px;
}

.dockerfile-content {
  padding: 16px;
  background: #1e293b;
  max-height: 400px;
  overflow-y: auto;
}

.dockerfile-content pre {
  margin: 0;
  font-family: 'SF Mono', 'Monaco', 'Inconsolata', monospace;
  font-size: 12.5px;
  line-height: 1.7;
  color: #e2e8f0;
  white-space: pre-wrap;
  word-break: break-all;
}

.dockerfile-content code {
  font-family: inherit;
}

/* ==================== 新增样式：section 分割线 ==================== */
.section-divider {
  display: flex;
  align-items: center;
  margin: 24px 0 20px;
  padding-bottom: 8px;
  border-bottom: 2px solid #f1f5f9;
}

.section-divider:first-child {
  margin-top: 0;
}

.divider-text {
  display: flex;
  align-items: center;
  font-size: 14px;
  font-weight: 700;
  color: #334155;
  letter-spacing: 0.3px;
}

/* ==================== 新增样式：语言标签 Badge ==================== */
.lang-badge {
  padding: 1px 6px;
  font-size: 10px;
  font-weight: 700;
  border-radius: 4px;
  letter-spacing: 0.5px;
  margin-left: 4px;
}

.lang-badge.rec {
  background: linear-gradient(135deg, #48bb78 0%, #38a169 100%);
  color: white;
}

.lang-badge.easy {
  background: linear-gradient(135deg, #63b3ed 0%, #4299e1 100%);
  color: white;
}

/* ==================== 新增样式：确认步骤 ==================== */
.confirm-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  animation: fadeIn 0.2s ease;
}

.confirm-modal {
  background: white;
  border-radius: 16px;
  padding: 32px;
  max-width: 600px;
  width: 90%;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 20px 60px rgba(0,0,0,0.3);
}

.confirm-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 24px;
}

.confirm-header svg {
  width: 32px;
  height: 32px;
  color: #f59e0b;
}

.confirm-header h3 {
  margin: 0;
  font-size: 20px;
  font-weight: 700;
  color: #1e293b;
}

.confirm-grid {
  display: grid;
  gap: 12px;
}

.confirm-item {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 12px 16px;
  background: #f8fafc;
  border-radius: 8px;
  border-left: 3px solid #4299e1;
}

.confirm-label {
  font-size: 12px;
  color: #64748b;
  font-weight: 500;
}

.confirm-value {
  font-size: 14px;
  color: #1e293b;
  font-weight: 600;
  text-align: right;
  word-break: break-all;
}

.confirm-actions {
  display: flex;
  gap: 12px;
  margin-top: 24px;
  justify-content: flex-end;
}

.btn-confirm-back {
  padding: 10px 24px;
  background: #f1f5f9;
  color: #475569;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-confirm-back:hover {
  background: #e2e8f0;
}

.btn-confirm-submit {
  padding: 10px 24px;
  background: linear-gradient(135deg, #3b82f6, #2563eb);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
}

.btn-confirm-submit:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 6px 20px rgba(59, 130, 246, 0.4);
}

.btn-confirm-submit:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* ==================== 新增样式：成功拓扑 ==================== */
.success-topology {
  min-height: 100vh;
  background: linear-gradient(135deg, #f5f7fa 0%, #e4e8ec 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px;
}

.topology-card {
  background: white;
  border-radius: 20px;
  padding: 40px;
  max-width: 800px;
  width: 100%;
  box-shadow: 0 20px 60px rgba(0,0,0,0.1);
  text-align: center;
}

.topology-success-icon {
  width: 80px;
  height: 80px;
  margin: 0 auto 24px;
  background: linear-gradient(135deg, #48bb78 0%, #38a169 100%);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  animation: popIn 0.4s ease;
}

.topology-success-icon svg {
  width: 40px;
  height: 40px;
  color: white;
}

.topology-card h2 {
  margin: 0 0 8px;
  font-size: 24px;
  font-weight: 700;
  color: #1e293b;
}

.topology-card > p {
  margin: 0 0 32px;
  color: #64748b;
  font-size: 14px;
}

.topology-warnings {
  background: #fffbeb;
  border: 1px solid #fde68a;
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 24px;
  text-align: left;
}

.topology-warnings h4 {
  margin: 0 0 8px;
  font-size: 14px;
  color: #92400e;
}

.topology-warnings ul {
  margin: 0;
  padding-left: 20px;
  font-size: 13px;
  color: #a16207;
}

.topology-flow {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  margin: 32px 0;
  flex-wrap: wrap;
}

.topology-node {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 20px 24px;
  background: #f8fafc;
  border: 2px solid #e2e8f0;
  border-radius: 12px;
  min-width: 120px;
  transition: all 0.3s;
}

.topology-node:hover {
  border-color: #a0c4e8;
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0,0,0,0.1);
}

.topology-node-icon {
  width: 48px;
  height: 48px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.topology-node-icon.git { background: linear-gradient(135deg, #f6ad55 0%, #ed8936 100%); color: white; }
.topology-node-icon.jenkins { background: linear-gradient(135deg, #fc8181 0%, #f56565 100%); color: white; }
.topology-node-icon.registry { background: linear-gradient(135deg, #63b3ed 0%, #4299e1 100%); color: white; }
.topology-node-icon.k8s { background: linear-gradient(135deg, #68d391 0%, #48bb78 100%); color: white; }

.topology-node-icon svg {
  width: 24px;
  height: 24px;
}

.topology-node-label {
  font-size: 12px;
  color: #64748b;
  font-weight: 500;
}

.topology-node-value {
  font-size: 13px;
  color: #1e293b;
  font-weight: 600;
}

.topology-arrow {
  font-size: 24px;
  color: #a0aec0;
  font-weight: 300;
}

.topology-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
  margin-top: 24px;
}

.btn-topology-primary {
  padding: 12px 32px;
  background: linear-gradient(135deg, #3b82f6, #2563eb);
  color: white;
  border: none;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
}

.btn-topology-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(59, 130, 246, 0.4);
}

.btn-topology-secondary {
  padding: 12px 32px;
  background: #f1f5f9;
  color: #475569;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-topology-secondary:hover {
  background: #e2e8f0;
}

/* ==================== 新增样式：K8s 导入弹窗 ==================== */
.k8s-import-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  animation: fadeIn 0.2s ease;
}

.k8s-import-modal {
  background: white;
  border-radius: 16px;
  padding: 32px;
  max-width: 500px;
  width: 90%;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 20px 60px rgba(0,0,0,0.3);
}

.k8s-import-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
}

.k8s-import-header h3 {
  margin: 0;
  font-size: 20px;
  font-weight: 700;
  color: #1e293b;
  display: flex;
  align-items: center;
  gap: 10px;
}

.k8s-import-header h3 svg {
  width: 24px;
  height: 24px;
  color: #667eea;
}

.k8s-import-close {
  width: 36px;
  height: 36px;
  border: none;
  background: #f1f5f9;
  border-radius: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.k8s-import-close:hover {
  background: #e2e8f0;
}

.k8s-import-close svg {
  width: 18px;
  height: 18px;
  color: #64748b;
}

.k8s-import-form .form-group {
  margin-bottom: 16px;
}

.k8s-import-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 24px;
}

.btn-import-cancel {
  padding: 10px 24px;
  background: #f1f5f9;
  color: #475569;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-import-cancel:hover {
  background: #e2e8f0;
}

.btn-import-submit {
  padding: 10px 24px;
  background: linear-gradient(135deg, #667eea, #764ba2);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.btn-import-submit:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 6px 20px rgba(102, 126, 234, 0.4);
}

.btn-import-submit:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* ==================== 新增样式：部署预览面板 ==================== */
.deploy-preview-panel {
  width: 320px;
  flex-shrink: 0;
}

.preview-card {
  background: white;
  border-radius: 16px;
  padding: 24px;
  box-shadow: 0 4px 20px rgba(0,0,0,0.08);
  position: sticky;
  top: 24px;
}

.preview-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 700;
  color: #1e293b;
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 2px solid #f1f5f9;
}

.preview-title svg {
  width: 20px;
  height: 20px;
  color: #4299e1;
}

.preview-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.preview-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 12px;
  background: #f8fafc;
  border-radius: 8px;
  border-left: 3px solid #e2e8f0;
}

.preview-item.highlight {
  background: linear-gradient(135deg, #ebf8ff 0%, #f0f9ff 100%);
  border-left-color: #4299e1;
}

.preview-label {
  font-size: 11px;
  color: #64748b;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.preview-value {
  font-size: 13px;
  color: #1e293b;
  font-weight: 600;
  word-break: break-all;
}

.preview-image {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
  background: #1e293b;
  color: #6ee7b7;
  padding: 8px 12px;
  border-radius: 6px;
  margin-top: 8px;
  word-break: break-all;
}

/* 动画 */
.slide-fade-enter-active,
.slide-fade-leave-active {
  transition: all 0.3s ease;
}

.slide-fade-enter-from,
.slide-fade-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}

@keyframes popIn {
  from { transform: scale(0); }
  to { transform: scale(1); }
}

@media (max-width: 1200px) {
  .deploy-preview-panel {
    display: none;
  }
}
</style>
