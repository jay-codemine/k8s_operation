<template>
  <div class="resource-view">
    <div class="view-header">
      <h1>Job管理</h1>
      <p>Kubernetes集群中的Job列表</p>
    </div>
    <div class="action-bar">
      <div class="search-box">
        <input type="text" v-model="searchQuery" placeholder="搜索Job名称..." @input="onSearchInput" />
      </div>

      <div class="filter-buttons">
        <button class="btn btn-filter" :class="{ active: statusFilter === 'all' }" @click="setStatusFilter('all')">
          全部
        </button>
        <button class="btn btn-filter" :class="{ active: statusFilter === 'Complete' }" @click="setStatusFilter('Complete')">
          Complete
        </button>
        <button class="btn btn-filter" :class="{ active: statusFilter === 'Running' }" @click="setStatusFilter('Running')">
          Running
        </button>
        <button class="btn btn-filter" :class="{ active: statusFilter === 'Failed' }" @click="setStatusFilter('Failed')">
          Failed
        </button>
      </div>

      <div class="filter-dropdown">
        <select v-model="namespaceFilter">
          <option value="">所有命名空间</option>
          <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
        </select>
      </div>

      <div class="action-buttons">
        <!-- 批量操作按钮 -->
        <button 
          v-if="!batchMode" 
          class="btn btn-batch" 
          @click="enterBatchMode"
          title="进入批量操作模式"
        >
          ☑️ 批量操作
        </button>
        <button 
          v-if="batchMode" 
          class="btn btn-secondary" 
          @click="exitBatchMode"
        >
          ✖️ 退出批量
        </button>

        <!-- 视图切换按钮 -->
        <div class="view-toggle">
          <button 
            class="btn btn-view" 
            :class="{ active: viewMode === 'table' }" 
            @click="viewMode = 'table'"
            title="表格视图"
          >
            📋
          </button>
          <button 
            class="btn btn-view" 
            :class="{ active: viewMode === 'card' }" 
            @click="viewMode = 'card'"
            title="卡片视图"
          >
            🗂️
          </button>
        </div>
        
        <label class="auto-refresh-toggle">
          <input type="checkbox" v-model="autoRefresh" />
          <span>自动刷新</span>
          <span v-if="autoRefresh" class="refresh-indicator">●</span>
        </label>
        <button class="btn btn-primary" @click="showCreateModal = true">创建Job</button>
        <button class="btn btn-secondary" @click="refreshList" :disabled="loading">
          {{ loading ? '加载中...' : '🔄 刷新' }}
        </button>
      </div>
    </div>

    <div v-if="errorMsg" class="error-box">{{ errorMsg }}</div>

    <!-- 批量操作浮动栏 -->
    <div v-if="batchMode && selectedJobs.length > 0" class="batch-action-bar">
      <div class="batch-info">
        <span class="batch-count">已选择 {{ selectedJobs.length }} 个 Job</span>
        <button class="batch-clear" @click="clearSelection">清空选择</button>
      </div>
      <div class="batch-actions">
        <button class="batch-btn danger" @click="openBatchDeletePreview" title="批量删除">
          🗑️ 批量删除
        </button>
      </div>
    </div>

    <!-- 数据统计信息 -->
    <div v-if="!loading && jobs.length > 0" class="stats-bar">
      <div class="stat-item">
        <span class="stat-label">总计:</span>
        <span class="stat-value">{{ total }}</span>
      </div>
      <div class="stat-item">
        <span class="stat-label">✅ Complete:</span>
        <span class="stat-value success">{{ getStatusCount('Complete') }}</span>
      </div>
      <div class="stat-item">
        <span class="stat-label">▶️ Running:</span>
        <span class="stat-value running">{{ getStatusCount('Running') }}</span>
      </div>
      <div class="stat-item">
        <span class="stat-label">❌ Failed:</span>
        <span class="stat-value failed">{{ getStatusCount('Failed') }}</span>
      </div>
      <div class="stat-item">
        <span class="stat-label">当前页:</span>
        <span class="stat-value">{{ paginatedJobs.length }}</span>
      </div>
    </div>

    <!-- 表格视图 -->
    <div v-if="viewMode === 'table'" class="table-container">
      <table class="resource-table">
        <thead>
          <tr>
            <th v-if="batchMode" style="width: 40px;">
              <input 
                type="checkbox" 
                :checked="isAllSelected" 
                @change="toggleSelectAll"
                title="全选/取消全选"
                ref="selectAllCheckbox"
              />
            </th>
            <th style="width: 100px;">状态</th>
            <th style="min-width: 180px;">名称</th>
            <th style="width: 130px;">命名空间</th>
            <th style="width: 150px;">完成/总数</th>
            <th style="width: 100px;">并行度</th>
            <th style="min-width: 200px;">镜像</th>
            <th style="min-width: 180px;">选择器</th>
            <th style="width: 100px;">退避限制</th>
            <th style="width: 170px; white-space: nowrap;">开始时间</th>
            <th style="width: 180px;">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="job in paginatedJobs" :key="job.name + job.namespace" :class="{ 'row-selected': isJobSelected(job) }">
            <td v-if="batchMode">
              <input 
                type="checkbox" 
                :checked="isJobSelected(job)" 
                @change="toggleJobSelection(job)"
              />
            </td>
            <td>
              <div class="status-cell">
                <span class="status-indicator" :class="job.status.toLowerCase()">
                  <span class="status-dot"></span>
                  {{ job.status }}
                </span>
                <!-- 状态详细信息提示 -->
                <div class="status-tooltip">
                  <div>活跃: {{ job.active || 0 }}</div>
                  <div>成功: {{ job.succeeded || 0 }}</div>
                  <div>失败: {{ job.failed || 0 }}</div>
                </div>
              </div>
            </td>
            <td>
              <div class="job-name">
                <span class="icon">⚙️</span>
                <span>{{ job.name }}</span>
                <!-- Kuboard风格：显示创建时长 -->
                <span class="age-badge" :title="job.created_at">{{ getAge(job.created_at) }}</span>
              </div>
            </td>
            <td>
              <span class="namespace-badge">{{ job.namespace }}</span>
            </td>
            <td>
              <div class="completion-info">
                <div class="completion-stats">
                  <span class="completion-text">{{ job.succeeded || 0 }} / {{ job.completions || 1 }}</span>
                  <span class="completion-percent">{{ getCompletionPercentage(job).toFixed(0) }}%</span>
                </div>
                <div class="completion-bar" :title="`进度: ${getCompletionPercentage(job).toFixed(1)}%`">
                  <div class="completion-fill" :class="getProgressClass(job)" :style="{ width: `${getCompletionPercentage(job)}%` }">
                    <span class="progress-shimmer"></span>
                  </div>
                </div>
              </div>
            </td>
            <td>
              <div class="parallelism-info">
                <span>{{ job.parallelism || 1 }}</span>
                <span v-if="job.active > 0" class="active-indicator" title="当前活跃Pod数">⚡{{ job.active }}</span>
              </div>
            </td>
            <td>
              <!-- Job 的 template 不可变，只显示镜像信息，不支持内联编辑 -->
              <div class="image-text" :title="`${job.image || '-'}\n\n⚠️ Job 创建后镜像不可修改\n如需更新镜像，请删除后重新创建`">
                <span class="image-name">{{ job.image || '-' }}</span>
                <span class="image-hint" title="Job 的 Pod 模板创建后不可修改">🔒</span>
              </div>
            </td>
            <td>
              <div class="selector-tags">
                <span v-for="(value, key) in job.selector" :key="key" class="selector-tag" :title="`${key}=${value}`">
                  {{ key }}={{ value }}
                </span>
                <span v-if="Object.keys(job.selector || {}).length === 0" class="selector-empty">-</span>
              </div>
            </td>
            <td>
              <span :class="{ 'warning-text': (job.failed || 0) >= (job.backoff_limit || 6) }">
                {{ job.backoff_limit || 6 }}
              </span>
            </td>
            <td style="white-space: nowrap;">
              <div class="time-info">
                <div>{{ job.start_time || '-' }}</div>
                <div v-if="job.completion_time" class="time-duration">耗时: {{ getDuration(job.start_time, job.completion_time) }}</div>
              </div>
            </td>
            <td>
              <div class="action-icons">
                <!-- Rancher风格：快速操作按钮 -->
                <button 
                  class="action-btn icon-only" 
                  @click="viewJobDetail(job)" 
                  title="查看详情"
                >
                  📋
                </button>
                <button 
                  v-if="job.status !== 'Complete'"
                  class="action-btn icon-only"
                  @click="restartJob(job)"
                  title="重启Job"
                >
                  🔄
                </button>
                
                <!-- 更多按钮 -->
                <div class="more-btn">
                  <button class="icon-btn" @click="toggleMoreOptions(job, $event)">
                    ⋮
                  </button>

                  <!-- 更多菜单 -->
                  <div v-if="showMoreOptions && selectedJob === job" class="more-menu" :style="menuStyle">
                    <button class="menu-item" @click="viewJobLogs(job)">
                      <span class="menu-icon">📄</span>
                      <span>查看日志</span>
                    </button>
                    <button class="menu-item" @click="viewJobYaml(job)">
                      <span class="menu-icon">📝</span>
                      <span>查看YAML</span>
                    </button>
                    <button class="menu-item" @click="suspendJob(job)" v-if="job.status !== 'Complete' && job.status !== 'Failed'">
                      <span class="menu-icon">⏸️</span>
                      <span>{{ job.suspend ? '恢复' : '暂停' }}</span>
                    </button>
                    <div class="menu-divider"></div>
                    <button class="menu-item" @click="restartJob(job)">
                      <span class="menu-icon">🔄</span>
                      <span>重新运行</span>
                    </button>
                    <div class="menu-divider"></div>
                    <button class="menu-item danger" @click="deleteJob(job)">
                      <span class="menu-icon">🗑️</span>
                      <span>删除</span>
                    </button>
                  </div>
                </div>
              </div>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-if="loading" class="loading-indicator">
        <div class="loading-spinner"></div>
        <p>加载中...</p>
      </div>
      <div v-if="!loading && filteredJobs.length === 0" class="empty-state">
        <div class="empty-icon">📁</div>
        <p class="empty-title">暂无Job数据</p>
        <p class="empty-desc">当前筛选条件下没有找到Job，试试调整筛选条件或创建新的Job</p>
        <button class="btn btn-primary" @click="showCreateModal = true" style="margin-top: 16px;">
          创建第一个Job
        </button>
      </div>
    </div>

    <!-- 卡片视图 -->
    <div v-else class="card-grid">
      <div v-for="job in paginatedJobs" :key="job.name + job.namespace" class="resource-card" :class="{ 'card-selected': isJobSelected(job) }">
        <!-- 批量选择复选框 -->
        <div v-if="batchMode" class="card-checkbox">
          <input 
            type="checkbox" 
            :checked="isJobSelected(job)" 
            @change="toggleJobSelection(job)"
          />
        </div>

        <div class="card-header" :class="`status-${job.status.toLowerCase()}`">
          <div class="card-title">
            <span class="card-icon">⚙️</span>
            <h3>{{ job.name }}</h3>
          </div>
          <span class="status-indicator" :class="job.status.toLowerCase()">
            {{ job.status }}
          </span>
        </div>

        <div class="card-body">
          <div class="card-meta">
            <div class="card-meta-item">
              <span class="label">命名空间:</span>
              <span class="namespace-badge">{{ job.namespace }}</span>
            </div>
            <div class="card-meta-item">
              <span class="label">完成数:</span>
              <span>{{ job.succeeded || 0 }} / {{ job.completions || 1 }}</span>
            </div>
            <div class="card-meta-item">
              <span class="label">并行度:</span>
              <span>{{ job.parallelism || 1 }}</span>
            </div>
            <div class="card-meta-item">
              <span class="label">镜像:</span>
              <span class="image-name">{{ job.image || '-' }}</span>
            </div>
            <div class="card-meta-item">
              <span class="label">退避限制:</span>
              <span>{{ job.backoff_limit || 6 }}</span>
            </div>
            <div class="card-meta-item">
              <span class="label">开始时间:</span>
              <span>{{ job.start_time || '-' }}</span>
            </div>
          </div>

          <div class="completion-bar-container">
            <div class="completion-bar">
              <div class="completion-fill" :style="{ width: `${getCompletionPercentage(job)}%` }"></div>
            </div>
            <span class="completion-text">{{ getCompletionPercentage(job).toFixed(0) }}%</span>
          </div>
        </div>

        <div class="card-footer">
          <button class="card-btn primary" @click="viewJobDetail(job)">📋 详情</button>
          <button class="card-btn" @click="restartJob(job)">🔄 重启</button>
          <button class="card-btn danger" @click="deleteJob(job)">🗑️ 删除</button>
        </div>
      </div>

      <div v-if="loading" class="loading-indicator">
        <div class="loading-spinner"></div>
        <p>加载中...</p>
      </div>
      <div v-if="!loading && filteredJobs.length === 0" class="empty-state">
        <div class="empty-icon">📁</div>
        <p class="empty-title">暂无Job数据</p>
        <p class="empty-desc">当前筛选条件下没有找到Job，试试调整筛选条件或创建新的Job</p>
        <button class="btn btn-primary" @click="showCreateModal = true" style="margin-top: 16px;">
          创建第一个Job
        </button>
      </div>
    </div>

    <!-- 分页组件 -->
    <div v-if="total > 0" class="pagination-container">
      <div class="pagination-info">
        <span class="total-info">共 {{ total }} 条数据</span>
      </div>
      <Pagination
        v-model:currentPage="currentPage"
        :totalItems="total"
        :itemsPerPage="itemsPerPage"
      />
      <div class="pagination-summary">
        <span>显示 {{ ((currentPage - 1) * itemsPerPage) + 1 }} - {{ Math.min(currentPage * itemsPerPage, total) }} 条</span>
      </div>
    </div>

    <!-- 创建Job模态框 - Rancher/Kuboard 风格 -->
    <div v-if="showCreateModal" class="modal-overlay" @click="showCreateModal = false">
      <div class="modal-content advanced" @click.stop>
        <div class="modal-header">
          <h2>📦 创建 Job</h2>
          <!-- 视图切换按钮 -->
          <div class="view-toggle-buttons">
            <button 
              class="view-toggle-btn" 
              :class="{ active: createMode === 'form' }"
              @click="createMode = 'form'"
              title="表单模式"
            >
              📝 表单
            </button>
            <button 
              class="view-toggle-btn" 
              :class="{ active: createMode === 'yaml' }"
              @click="createMode = 'yaml'"
              title="YAML模式"
            >
              📄 YAML
            </button>
          </div>
          <button @click="showCreateModal = false" class="close-btn">&times;</button>
        </div>
        
        <!-- 步骤指示器 -->
        <div class="steps-indicator">
          <div 
            class="step-item" 
            :class="{ active: currentStep === 1, completed: currentStep > 1 }"
            @click="currentStep = 1"
          >
            <div class="step-number">1</div>
            <div class="step-label">基础配置</div>
          </div>
          <div class="step-line" :class="{ active: currentStep > 1 }"></div>
          <div 
            class="step-item" 
            :class="{ active: currentStep === 2, completed: currentStep > 2 }"
            @click="currentStep = 2"
          >
            <div class="step-number">2</div>
            <div class="step-label">容器配置</div>
          </div>
          <div class="step-line" :class="{ active: currentStep > 2 }"></div>
          <div 
            class="step-item" 
            :class="{ active: currentStep === 3, completed: currentStep > 3 }"
            @click="currentStep = 3"
          >
            <div class="step-number">3</div>
            <div class="step-label">高级选项</div>
          </div>
        </div>

        <div class="modal-body scrollable">
          <!-- 表单模式 -->
          <div v-if="createMode === 'form'">
          <!-- 步骤 1: 基础配置 -->
          <div v-show="currentStep === 1" class="step-content">
            <div class="form-section">
              <div class="section-header">
                <h3>🔖 基础信息</h3>
                <p class="section-desc">配置 Job 的基本标识信息</p>
              </div>
              
              <div class="form-row">
                <div class="form-group">
                  <label class="required">命名空间</label>
                  <div class="namespace-selector">
                    <select 
                      v-if="!showNamespaceInput" 
                      v-model="jobForm.namespace" 
                      class="form-select" 
                      required
                      style="flex: 1;"
                    >
                      <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
                    </select>
                    <span v-if="!showNamespaceInput" class="namespace-or">或</span>
                    <div v-if="showNamespaceInput" class="namespace-create">
                      <input 
                        v-model="newNamespace" 
                        type="text" 
                        class="form-input"
                        placeholder="新命名空间名称"
                        :disabled="creatingNamespace"
                      />
                      <button 
                        type="button" 
                        class="btn btn-secondary btn-sm" 
                        @click="createNewNamespace"
                        :disabled="!newNamespace || creatingNamespace"
                      >
                        {{ creatingNamespace ? '创建中...' : '创建' }}
                      </button>
                      <button 
                        type="button" 
                        class="btn btn-secondary btn-sm" 
                        @click="cancelCreateNamespace"
                      >
                        取消
                      </button>
                    </div>
                    <button 
                      v-if="!showNamespaceInput"
                      type="button" 
                      class="btn btn-secondary btn-sm" 
                      @click="showNamespaceInput = true"
                    >
                      ➕ 创建新命名空间
                    </button>
                  </div>
                  <span class="form-hint">Job 将在此命名空间中创建</span>
                </div>
                
                <div class="form-group">
                  <label class="required">Job 名称</label>
                  <input 
                    type="text" 
                    v-model="jobForm.name" 
                    required 
                    placeholder="例如: data-processing-job"
                    @input="validateJobName"
                  />
                  <span class="form-hint" :class="{ 'error-hint': jobNameError }">
                    {{ jobNameError || '只能包含小写字母、数字和连字符' }}
                  </span>
                </div>
              </div>

              <div class="form-group">
                <label>描述（可选）</label>
                <textarea 
                  v-model="jobForm.description" 
                  placeholder="描述此 Job 的用途，例如：每日数据清理任务"
                  rows="3"
                ></textarea>
              </div>
            </div>

            <div class="form-section">
              <div class="section-header">
                <h3>⚙️ 执行策略</h3>
                <p class="section-desc">定义 Job 的执行模式和完成条件</p>
              </div>
              
              <div class="form-row">
                <div class="form-group">
                  <label>完成数 <span class="info-icon" title="需要成功完成的 Pod 数量">ℹ️</span></label>
                  <input 
                    type="number" 
                    v-model.number="jobForm.completions" 
                    min="1" 
                    max="1000"
                  />
                  <span class="form-hint">需要成功完成的 Pod 数量（默认 1）</span>
                </div>
                
                <div class="form-group">
                  <label>并行度 <span class="info-icon" title="同时运行的 Pod 数量">ℹ️</span></label>
                  <input 
                    type="number" 
                    v-model.number="jobForm.parallelism" 
                    min="1" 
                    :max="jobForm.completions || 100"
                  />
                  <span class="form-hint">同时运行的 Pod 数量（默认 1）</span>
                </div>
              </div>

              <div class="form-row">
                <div class="form-group">
                  <label>退避限制 <span class="info-icon" title="失败重试次数">ℹ️</span></label>
                  <input 
                    type="number" 
                    v-model.number="jobForm.backoff_limit" 
                    min="0" 
                    max="100"
                  />
                  <span class="form-hint">失败后的最大重试次数（默认 6）</span>
                </div>
                
                <div class="form-group">
                  <label>超时时间（秒） <span class="info-icon" title="Job 最长执行时间">ℹ️</span></label>
                  <input 
                    type="number" 
                    v-model.number="jobForm.active_deadline_seconds" 
                    min="0" 
                    placeholder="不限制"
                  />
                  <span class="form-hint">Job 执行的最长时间，超时后自动终止</span>
                </div>
              </div>

              <div class="form-group">
                <label>完成后保留时间（秒） <span class="info-icon" title="Job 完成后自动删除的时间">ℹ️</span></label>
                <input 
                  type="number" 
                  v-model.number="jobForm.ttl_seconds_after_finished" 
                  min="0" 
                  placeholder="永久保留"
                />
                <span class="form-hint">Job 完成后保留的时间，到期后自动删除（0=立即删除，空=永久保留）</span>
              </div>
            </div>
          </div>

          <!-- 步骤 2: 容器配置 -->
          <div v-show="currentStep === 2" class="step-content">
            <div class="form-section">
              <div class="section-header">
                <h3>🐳 容器镜像</h3>
                <p class="section-desc">配置容器的镜像和启动命令</p>
              </div>
              
              <div class="form-group">
                <label class="required">容器镜像</label>
                <input 
                  type="text" 
                  v-model="jobForm.container_image" 
                  required 
                  placeholder="例如: busybox:latest 或 nginx:1.21"
                />
                <span class="form-hint">指定容器的镜像地址，支持 Docker Hub、私有仓库等</span>
              </div>

              <div class="form-group">
                <label>容器名称</label>
                <input 
                  type="text" 
                  v-model="jobForm.container_name" 
                  placeholder="默认使用 Job 名称"
                />
                <span class="form-hint">容器的名称，留空则使用 Job 名称</span>
              </div>

              <div class="form-group">
                <label>启动命令 (Command) <span class="info-icon" title="覆盖镜像的 ENTRYPOINT">ℹ️</span></label>
                <input 
                  type="text" 
                  v-model="jobForm.container_command" 
                  placeholder="例如: /bin/sh"
                />
                <span class="form-hint">容器的启动命令，会覆盖镜像的 ENTRYPOINT</span>
              </div>

              <div class="form-group">
                <label>命令参数 (Args) <span class="info-icon" title="传递给启动命令的参数">ℹ️</span></label>
                <input 
                  type="text" 
                  v-model="jobForm.container_command_args" 
                  placeholder="例如: -c echo hello world"
                />
                <span class="form-hint">传递给启动命令的参数，多个参数用空格分隔</span>
              </div>
            </div>

            <div class="form-section">
              <div class="section-header">
                <h3>🔐 镜像拉取策略</h3>
              </div>
              
              <div class="form-row">
                <div class="form-group">
                  <label>镜像拉取策略</label>
                  <select v-model="jobForm.image_pull_policy">
                    <option value="">默认（IfNotPresent）</option>
                    <option value="Always">Always - 总是拉取</option>
                    <option value="IfNotPresent">IfNotPresent - 本地不存在时拉取</option>
                    <option value="Never">Never - 从不拉取</option>
                  </select>
                </div>
                
                <div class="form-group">
                  <label>镜像拉取密钥</label>
                  <input 
                    type="text" 
                    v-model="jobForm.image_pull_secret" 
                    placeholder="例如: my-registry-secret"
                  />
                  <span class="form-hint">私有镜像仓库的认证密钥名称</span>
                </div>
              </div>
            </div>

            <div class="form-section">
              <div class="section-header">
                <h3>🌍 环境变量</h3>
                <button type="button" class="add-btn" @click="addEnvVar">
                  ➕ 添加环境变量
                </button>
              </div>
              
              <div v-for="(env, index) in jobForm.variables" :key="index" class="env-var-row">
                <input 
                  type="text" 
                  v-model="env.name" 
                  placeholder="变量名"
                  class="env-name"
                />
                <input 
                  type="text" 
                  v-model="env.value" 
                  placeholder="变量值"
                  class="env-value"
                />
                <button type="button" class="remove-btn" @click="removeEnvVar(index)">
                  🗑️
                </button>
              </div>
              <div v-if="jobForm.variables.length === 0" class="empty-hint">
                暂无环境变量，点击上方按钮添加
              </div>
            </div>
          </div>

          <!-- 步骤 3: 高级选项 -->
          <div v-show="currentStep === 3" class="step-content">
            <div class="form-section">
              <div class="section-header">
                <h3>🔄 重启策略</h3>
                <p class="section-desc">定义 Pod 失败后的重启行为</p>
              </div>
              
              <div class="form-group">
                <label>重启策略</label>
                <select v-model="jobForm.restart_policy">
                  <option value="OnFailure">OnFailure - 失败时重启（推荐）</option>
                  <option value="Never">Never - 从不重启</option>
                </select>
                <span class="form-hint">OnFailure 适合大多数场景，Never 适合一次性任务</span>
              </div>
            </div>

            <div class="form-section">
              <div class="section-header">
                <h3>📊 资源配置</h3>
                <p class="section-desc">限制容器的 CPU 和内存使用</p>
              </div>
              
              <div class="form-row">
                <div class="form-group">
                  <label>CPU 请求</label>
                  <input 
                    type="text" 
                    v-model="jobForm.cpu_requirement" 
                    placeholder="例如: 100m 或 0.5"
                  />
                  <span class="form-hint">最小 CPU 需求（100m = 0.1核）</span>
                </div>
                
                <div class="form-group">
                  <label>内存请求</label>
                  <input 
                    type="text" 
                    v-model="jobForm.memory_requirement" 
                    placeholder="例如: 128Mi 或 1Gi"
                  />
                  <span class="form-hint">最小内存需求</span>
                </div>
              </div>
            </div>

            <div class="form-section">
              <div class="section-header">
                <h3>🏷️ 标签与注解</h3>
                <button type="button" class="add-btn" @click="addLabel">
                  ➕ 添加标签
                </button>
              </div>
              
              <div v-for="(label, index) in jobForm.labels" :key="index" class="env-var-row">
                <input 
                  type="text" 
                  v-model="label.key" 
                  placeholder="标签键"
                  class="env-name"
                />
                <input 
                  type="text" 
                  v-model="label.value" 
                  placeholder="标签值"
                  class="env-value"
                />
                <button type="button" class="remove-btn" @click="removeLabel(index)">
                  🗑️
                </button>
              </div>
              <div v-if="jobForm.labels.length === 0" class="empty-hint">
                暂无标签，点击上方按钮添加
              </div>
            </div>

            <div class="form-section">
              <div class="section-header">
                <h3>🎯 节点选择器</h3>
                <button type="button" class="add-btn" @click="addNodeSelector">
                  ➕ 添加节点选择器
                </button>
              </div>
              
              <div v-for="(selector, index) in jobForm.node_selectors" :key="index" class="env-var-row">
                <input 
                  type="text" 
                  v-model="selector.key" 
                  placeholder="节点标签键"
                  class="env-name"
                />
                <input 
                  type="text" 
                  v-model="selector.value" 
                  placeholder="节点标签值"
                  class="env-value"
                />
                <button type="button" class="remove-btn" @click="removeNodeSelector(index)">
                  🗑️
                </button>
              </div>
              <div v-if="jobForm.node_selectors.length === 0" class="empty-hint">
                暂无节点选择器，Job 将在任意节点运行
              </div>
            </div>
          </div>
          </div>

          <!-- YAML 模式 -->
          <div v-if="createMode === 'yaml'" class="yaml-editor-container">
            <div class="yaml-editor-header">
              <h3>📄 YAML 配置</h3>
              <p class="yaml-hint">✨ 支持多资源 YAML 创建（用 <code>---</code> 分隔），可同时创建 ConfigMap、Secret、PVC 等依赖资源</p>
              <div class="yaml-header-buttons">
                <button class="load-template-btn" @click="loadMultiResourceYamlTemplate">
                  📑 加载多资源模板
                </button>
                <button class="load-template-btn" @click="loadJobYamlTemplate">
                  📄 Job 模板
                </button>
                <button class="copy-yaml-btn" @click="copyYamlContent">
                  📋 复制
                </button>
                <button class="reset-yaml-btn" @click="resetYamlContent">
                  🔄 重置
                </button>
              </div>

              <!-- 多资源预览面板 -->
              <div v-if="multiResourcePreview.resources.length > 0" class="multi-resource-preview">
                <div class="preview-header">
                  <h4>🔍 资源预览 ({{ multiResourcePreview.resources.length }} 个资源)</h4>
                  <button
                    v-if="multiResourcePreview.errors.length === 0"
                    class="btn btn-sm btn-primary"
                    @click="applyMultiResourceYaml"
                    :disabled="creating"
                  >
                    {{ creating ? '创建中...' : '🚀 一键创建所有资源' }}
                  </button>
                </div>

                <!-- 依赖关系图 -->
                <div v-if="multiResourcePreview.dependencies.length > 0" class="dependencies-info">
                  <h5>🔗 依赖关系</h5>
                  <div class="dependency-graph">
                    <div
                      v-for="dep in multiResourcePreview.dependencies"
                      :key="dep.id"
                      class="dependency-item"
                    >
                      <span class="dep-source">{{ dep.source.kind }}/{{ dep.source.name }}</span>
                      <span class="dep-arrow">→</span>
                      <span class="dep-target">{{ dep.target.kind }}/{{ dep.target.name }}</span>
                    </div>
                  </div>
                </div>

                <!-- 资源列表 -->
                <div class="resource-list">
                  <div
                    v-for="resource in multiResourcePreview.resources"
                    :key="`${resource.kind}-${resource.name}`"
                    class="resource-item"
                    :class="{
                      'has-warnings': resource.warnings?.length > 0,
                      'has-errors': multiResourcePreview.errors.some(e => e.includes(resource.name))
                    }"
                  >
                    <div class="resource-header">
                      <span class="resource-kind" :class="getResourceKindClass(resource.kind)">
                        {{ getResourceIcon(resource.kind) }} {{ resource.kind }}
                      </span>
                      <span class="resource-name">{{ resource.name }}</span>
                      <span class="resource-namespace">({{ resource.namespace }})</span>
                      <span class="resource-order">#{{ resource.order }}</span>
                    </div>

                    <div v-if="resource.warnings?.length > 0" class="resource-warnings">
                      <div v-for="warning in resource.warnings" :key="warning" class="warning-item">
                        ⚠️ {{ warning }}
                      </div>
                    </div>
                  </div>
                </div>

                <!-- 错误信息 -->
                <div v-if="multiResourcePreview.errors.length > 0" class="preview-errors">
                  <h5>❌ 解析错误</h5>
                  <div class="error-list">
                    <div
                      v-for="error in multiResourcePreview.errors"
                      :key="error"
                      class="error-item"
                    >
                      {{ error }}
                    </div>
                  </div>
                </div>
              </div>
            </div>
            
            <textarea 
              v-model="yamlContent" 
              class="yaml-editor"
              placeholder="输入或粘贴 YAML 内容..."
              spellcheck="false"
            ></textarea>
            
            <div v-if="yamlError" class="yaml-error">
              <span class="error-icon">⚠️</span>
              {{ yamlError }}
            </div>
            
            <div class="yaml-editor-footer">
              <div class="yaml-tips">
                <strong>💡 提示：</strong>
                <ul>
                  <li>支持完整的 Kubernetes Job 配置</li>
                  <li>可以通过“加载模板”获取示例 YAML</li>
                  <li>创建前会验证 YAML 格式的正确性</li>
                </ul>
              </div>
            </div>
          </div>
        </div>

        <div class="modal-footer">
          <!-- 表单模式按钮 -->
          <template v-if="createMode === 'form'">
            <button 
              v-if="currentStep > 1" 
              @click="currentStep--" 
              class="nav-btn secondary"
            >
              ⬅️ 上一步
            </button>
            <button @click="showCreateModal = false" class="cancel-btn">取消</button>
            <button 
              v-if="currentStep < 3" 
              @click="currentStep++" 
              class="nav-btn primary"
            >
              下一步 ➡️
            </button>
            <button 
              type="button"
              v-if="currentStep === 3"
              @click="createJob" 
              class="submit-btn" 
              :disabled="creating || !isFormValid"
            >
              {{ creating ? '创建中...' : '✅ 创建 Job' }}
            </button>
          </template>
          
          <!-- YAML 模式按钮 -->
          <template v-if="createMode === 'yaml'">
            <button type="button" @click="showCreateModal = false" class="cancel-btn">取消</button>
            <button 
              type="button"
              @click="createJobFromYaml" 
              class="submit-btn"
              :disabled="creating || !yamlContent"
            >
              {{ creating ? '创建中...' : '✅ 创建 Job' }}
            </button>
          </template>
        </div>
      </div>
    </div>

    <!-- 批量删除预览弹窗（高危操作） -->
    <div v-if="showBatchDeleteModal" class="modal-overlay" @click.self="closeBatchDeleteModal">
      <div class="modal-content modal-batch-preview modal-danger">
        <div class="modal-header danger-header">
          <h3>🗑️ 批量删除预览（高危操作）</h3>
          <button class="close-btn" @click="closeBatchDeleteModal">×</button>
        </div>
        <div class="modal-body">
          <!-- 警告提示 -->
          <div class="danger-warning">
            <div class="warning-icon-large">⚠️</div>
            <div class="warning-content">
              <div class="warning-title">将删除以下 Job</div>
              <ul class="warning-list">
                <li>Job 及其创建的所有 Pod 将被删除</li>
                <li>此操作不可撤销！</li>
              </ul>
            </div>
          </div>
          
          <!-- 受影响 Job 列表 -->
          <div class="preview-section">
            <div class="section-title">受影响 Job ({{ selectedJobs.length }})</div>
            <div class="affected-jobs-detail">
              <div v-for="job in selectedJobs" :key="job.name" class="affected-job-card">
                <div class="job-info">
                  <span class="job-name">⚙️ {{ job.name }}</span>
                  <span class="job-namespace">{{ job.namespace }}</span>
                </div>
                <div class="job-stats">
                  <span class="status-indicator" :class="job.status.toLowerCase()">
                    {{ job.status }}
                  </span>
                  <span class="completion-tag">{{ job.succeeded || 0 }}/{{ job.completions || 1 }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- 二次确认 -->
          <div class="confirm-section">
            <div class="section-title">请输入 "DELETE" 确认操作</div>
            <input 
              v-model="deleteConfirmText" 
              placeholder="请输入 DELETE" 
              class="confirm-input"
              :class="{ valid: deleteConfirmText === 'DELETE' }"
            />
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeBatchDeleteModal">取消</button>
          <button 
            class="btn btn-danger" 
            @click="executeBatchDelete" 
            :disabled="deleteConfirmText !== 'DELETE' || batchExecuting"
          >
            {{ batchExecuting ? '执行中...' : '确认删除' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 详情模态框 -->
    <div v-if="showDetailModal" class="modal-overlay" @click="showDetailModal = false">
      <div class="modal-content large" @click.stop>
        <div class="modal-header">
          <h2>Job详情</h2>
          <button @click="showDetailModal = false" class="close-btn">&times;</button>
        </div>
        <div class="modal-body">
          <div v-if="loadingDetail" class="loading-indicator">加载中...</div>
          <div v-else-if="detailData" class="detail-content">
            <div class="detail-section">
              <h3>基本信息</h3>
              <div class="detail-row">
                <span class="detail-label">名称:</span>
                <span>{{ detailData.name }}</span>
              </div>
              <div class="detail-row">
                <span class="detail-label">命名空间:</span>
                <span>{{ detailData.namespace }}</span>
              </div>
              <div class="detail-row">
                <span class="detail-label">状态:</span>
                <span class="status-indicator" :class="detailData.status?.toLowerCase()">
                  {{ detailData.status }}
                </span>
              </div>
              <div class="detail-row">
                <span class="detail-label">镜像:</span>
                <span>{{ detailData.image || '-' }}</span>
              </div>
            </div>

            <div class="detail-section">
              <h3>执行信息</h3>
              <div class="detail-row">
                <span class="detail-label">完成数:</span>
                <span>{{ detailData.succeeded || 0 }} / {{ detailData.completions || 1 }}</span>
              </div>
              <div class="detail-row">
                <span class="detail-label">并行度:</span>
                <span>{{ detailData.parallelism || 1 }}</span>
              </div>
              <div class="detail-row">
                <span class="detail-label">退避限制:</span>
                <span>{{ detailData.backoff_limit || 6 }}</span>
              </div>
              <div class="detail-row">
                <span class="detail-label">活跃Pod:</span>
                <span>{{ detailData.active || 0 }}</span>
              </div>
              <div class="detail-row">
                <span class="detail-label">失败次数:</span>
                <span>{{ detailData.failed || 0 }}</span>
              </div>
            </div>

            <div class="detail-section">
              <h3>时间信息</h3>
              <div class="detail-row">
                <span class="detail-label">开始时间:</span>
                <span>{{ detailData.start_time || '-' }}</span>
              </div>
              <div class="detail-row">
                <span class="detail-label">完成时间:</span>
                <span>{{ detailData.completion_time || '-' }}</span>
              </div>
            </div>

            <div class="detail-section" v-if="detailData.selector && Object.keys(detailData.selector).length > 0">
              <h3>选择器</h3>
              <div class="selector-tags">
                <span v-for="(value, key) in detailData.selector" :key="key" class="selector-tag">
                  {{ key }}={{ value }}
                </span>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button @click="showDetailModal = false" class="cancel-btn">关闭</button>
        </div>
      </div>
    </div>

    <!-- YAML 查看/编辑模态框 -->
    <div v-if="showYamlModal" class="modal-overlay" @click.self="closeYamlModal">
      <div class="modal-content yaml-modal">
        <div class="modal-header">
          <h3>📝 {{ yamlEditMode ? '编辑 YAML' : '查看 YAML' }} - {{ selectedYamlJob?.name }}</h3>
          <div class="yaml-header-actions">
            <button v-if="!yamlEditMode" class="btn btn-sm btn-secondary" @click="downloadYaml" :disabled="loadingYaml">
              📄 下载
            </button>
            <button v-if="!yamlEditMode" class="btn btn-sm btn-secondary" @click="yamlEditMode = true">
              ✈️ 编辑模式
            </button>
            <button v-if="yamlEditMode" class="btn btn-sm btn-secondary" @click="yamlEditMode = false">
              👁️ 预览模式
            </button>
            <button class="close-btn" @click="closeYamlModal">×</button>
          </div>
        </div>
        <div class="modal-body yaml-modal-body">
          <div v-if="loadingYaml" class="loading-state">加载 YAML...</div>
          <div v-else-if="yamlViewError" class="error-box">{{ yamlViewError }}</div>
          <div v-else class="yaml-editor-wrapper">
            <textarea v-if="yamlEditMode" v-model="yamlViewContent" class="yaml-editor" spellcheck="false"></textarea>
            <pre v-else class="yaml-content">{{ yamlViewContent }}</pre>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeYamlModal">取消</button>
          <button v-if="yamlEditMode" class="btn btn-primary" @click="applyYamlChanges" :disabled="savingYaml">
            {{ savingYaml ? '保存中...' : '应用更改' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Job 日志模态框 -->
    <div v-if="showJobLogsModal" class="modal-overlay" @click.self="closeJobLogs">
      <div class="modal-content logs-modal">
        <div class="modal-header">
          <h3>📄 Job 日志</h3>
          <button class="close-btn" @click="closeJobLogs">×</button>
        </div>
        <div class="modal-body">
          <div v-if="selectedJobForLogs" class="info-box" style="margin-bottom: 16px;">
            <div><strong>Job:</strong> {{ selectedJobForLogs.name }}</div>
            <div><strong>命名空间:</strong> {{ selectedJobForLogs.namespace }}</div>
          </div>

          <!-- 日志控制面板 -->
          <div class="logs-controls">
            <div class="control-item">
              <label>Pod 选择</label>
              <select v-model="jobLogsForm.selectedPod" class="form-select" @change="onJobPodChange">
                <option value="">全部 Pod</option>
                <option v-for="pod in jobPodsList" :key="pod.name" :value="pod.name">
                  {{ pod.name }}
                </option>
              </select>
            </div>

            <div class="control-item" v-if="jobLogsForm.selectedPod">
              <label>容器</label>
              <select v-model="jobLogsForm.container" class="form-select">
                <option value="" disabled>选择容器</option>
                <option v-for="c in jobContainerList" :key="c" :value="c">
                  {{ c }}
                </option>
              </select>
            </div>

            <div class="control-item">
              <label>行数</label>
              <select v-model="jobLogsForm.tail" class="form-select">
                <option :value="null">全部</option>
                <option :value="10">10 行</option>
                <option :value="50">50 行</option>
                <option :value="100">100 行</option>
                <option :value="200">200 行</option>
                <option :value="500">500 行</option>
                <option :value="1000">1000 行</option>
              </select>
            </div>

            <div class="control-item">
              <label class="follow-toggle">
                <input type="checkbox" v-model="jobLogsForm.follow" />
                实时日志
                <span v-if="jobLogsForm.follow && isStreamingJobLogs" class="streaming-indicator">●</span>
              </label>
            </div>

            <div class="control-actions">
              <button 
                class="btn btn-primary btn-sm" 
                @click="fetchJobLogs" 
                :disabled="loadingJobLogs || isStreamingJobLogs"
              >
                {{ loadingJobLogs ? '获取中...' : (jobLogsForm.follow ? '获取实时日志' : '获取日志') }}
              </button>
              <button 
                v-if="loadingJobLogs || isStreamingJobLogs" 
                class="btn btn-danger btn-sm" 
                @click="stopJobLogStream"
              >
                终止
              </button>
              <button 
                class="btn btn-secondary btn-sm" 
                @click="clearJobLogs"
                :disabled="!jobLogsContent"
              >
                清除
              </button>
            </div>
          </div>

          <div v-if="loadingJobLogs && !isStreamingJobLogs" class="loading-state">加载中...</div>
          
          <div v-else-if="jobLogsError" class="error-box">
            {{ jobLogsError }}
          </div>

          <!-- 日志内容 -->
          <div v-else class="logs-viewer">
            <pre class="logs-content" ref="jobLogsContentRef" v-html="highlightedJobLogs"></pre>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, watch, watchEffect } from 'vue'
import { Message } from '@arco-design/web-vue'
import Pagination from '@/components/Pagination.vue'
import jobsApi from '@/api/cluster/workloads/jobs'
import namespacesApi from '@/api/cluster/namespaces'
import namespaceApi from '@/api/cluster/config/namespace'

// =========================
// 状态管理
// =========================
const loading = ref(false)
const errorMsg = ref('')
const jobs = ref([])
const total = ref(0)
const namespaces = ref([])

// 视图模式
const viewMode = ref('table') // 'table' | 'card'

// ========== 批量操作相关 ==========
const batchMode = ref(false)
const selectedJobs = ref([])
const showBatchDeleteModal = ref(false)
const deleteConfirmText = ref('')
const batchExecuting = ref(false)

// 搜索和筛选
const searchQuery = ref('')
const namespaceFilter = ref('')
const statusFilter = ref('all')

// 分页
const currentPage = ref(1)
const itemsPerPage = ref(10)

// 自动刷新
const autoRefresh = ref(false)
let refreshTimer = null

// 更多菜单
const showMoreOptions = ref(false)
const selectedJob = ref(null)
const menuStyle = ref({})

// 模态框
const showCreateModal = ref(false)
const showDetailModal = ref(false)
const creating = ref(false)
const loadingDetail = ref(false)
const detailData = ref(null)

// YAML 查看/编辑模态框
const showYamlModal = ref(false)
const selectedYamlJob = ref(null)
const yamlViewContent = ref('') // 用于 YAML 查看/编辑模态框
const yamlEditMode = ref(false)
const loadingYaml = ref(false)
const savingYaml = ref(false)
const yamlViewError = ref('')

// ========== Job 日志相关 ==========
const showJobLogsModal = ref(false)
const selectedJobForLogs = ref(null)
const jobPodsList = ref([])           // Job 关联的 Pod 列表
const jobContainerList = ref([])      // 当前 Pod 的容器列表
const jobLogsForm = ref({
  selectedPod: '',
  container: '',
  tail: 100,
  follow: false
})
const loadingJobLogs = ref(false)
const jobLogsContent = ref('')
const jobLogsError = ref('')
const isStreamingJobLogs = ref(false)
let jobLogAbortController = null
const jobLogsContentRef = ref(null)

// 创建表单
const currentStep = ref(1) // 当前步骤
const jobNameError = ref('') // 名称验证错误
const createMode = ref('form') // 'form' | 'yaml' - 创建模式
const yamlContent = ref('') // YAML 内容
const yamlError = ref('') // YAML 验证错误

// 多资源预览相关状态
const multiResourcePreview = ref({
  resources: [],
  dependencies: [],
  errors: []
})

// 创建命名空间相关
const showNamespaceInput = ref(false)
const newNamespace = ref('')
const creatingNamespace = ref(false)

// =========================
// 内联编辑状态
// =========================
const inlineEdit = ref({
  key: '',      // 当前编辑的 key，如 'image-job-name-namespace'
  value: null,  // 编辑的值
  original: null, // 原始值，用于取消时恢复
  job: null,    // 当前编辑的 Job 对象
  container: '' // 容器名称
})

// 防抖函数 - 必须在 watch 使用之前定义
const debounce = (func, wait) => {
  let timeout
  return function executedFunction(...args) {
    const later = () => {
      clearTimeout(timeout)
      func(...args)
    }
    clearTimeout(timeout)
    timeout = setTimeout(later, wait)
  }
}

// 获取认证头 - 必须在 parseYamlPreview 之前定义
const getAuthHeaders = () => {
  const token = localStorage.getItem('token')
  return token ? { 'Authorization': `Bearer ${token}` } : {}
}

// 实时解析 YAML 并预览多资源 - 必须在 watch 使用之前定义
const parseYamlPreview = async () => {
  if (!yamlContent.value.trim()) {
    multiResourcePreview.value = {
      resources: [],
      dependencies: [],
      errors: []
    }
    return
  }

  try {
    // 调用后端 API 解析 YAML
    const res = await fetch('/api/v1/k8s/multi-resource/parse-yaml', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...getAuthHeaders()
      },
      body: JSON.stringify({
        yaml: yamlContent.value,
        cluster_id: 1 // 默认集群ID
      })
    })

    if (res.ok) {
      const data = await res.json()
      if (data.code === 0) {
        multiResourcePreview.value = {
          resources: data.data.resources || [],
          dependencies: data.data.dependencies || [],
          errors: data.data.errors || []
        }
        yamlError.value = ''
      } else {
        multiResourcePreview.value = {
          resources: [],
          dependencies: [],
          errors: [data.msg || '解析失败']
        }
        yamlError.value = data.msg || '解析失败'
      }
    } else {
      const errorData = await res.json().catch(() => ({}))
      multiResourcePreview.value = {
        resources: [],
        dependencies: [],
        errors: [errorData.msg || `HTTP ${res.status}: ${res.statusText}`]
      }
      yamlError.value = errorData.msg || `解析失败: HTTP ${res.status}`
    }
  } catch (e) {
    console.error('解析 YAML 失败:', e)
    multiResourcePreview.value = {
      resources: [],
      dependencies: [],
      errors: [`网络错误: ${e.message}`]
    }
    yamlError.value = `解析失败: ${e.message}`
  }
}

// 监听 createMode 变化，切换到 YAML 模式时如果内容为空则自动加载模板
watch(createMode, (newMode) => {
  if (newMode === 'yaml' && !yamlContent.value.trim()) {
    loadJobYamlTemplate()
  }
})

// 监听 YAML 内容变化，实现实时预览
watch(yamlContent, debounce(parseYamlPreview, 800))

const jobForm = ref({
  namespace: 'default',
  name: '',
  description: '',
  container_image: '',
  container_name: '',
  container_command: null,
  container_command_args: null,
  completions: 1,
  parallelism: 1,
  backoff_limit: 6,
  active_deadline_seconds: null,
  ttl_seconds_after_finished: null,
  restart_policy: 'OnFailure',
  image_pull_policy: '',
  image_pull_secret: null,
  cpu_requirement: null,
  memory_requirement: null,
  variables: [],
  labels: [],
  node_selectors: []
})

// 表单验证
const isFormValid = computed(() => {
  return jobForm.value.name && 
         jobForm.value.container_image && 
         !jobNameError.value
})

// Job 名称验证
const validateJobName = () => {
  const name = jobForm.value.name
  if (!name) {
    jobNameError.value = ''
    return
  }
  
  // Kubernetes 命名规则：小写字母、数字、连字符
  const regex = /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/
  if (!regex.test(name)) {
    jobNameError.value = '只能包含小写字母、数字和连字符，且不能以连字符开头或结尾'
  } else if (name.length > 63) {
    jobNameError.value = '名称长度不能超过63个字符'
  } else {
    jobNameError.value = ''
  }
}

// 环境变量管理
const addEnvVar = () => {
  jobForm.value.variables.push({ name: '', value: '' })
}

const removeEnvVar = (index) => {
  jobForm.value.variables.splice(index, 1)
}

// 标签管理
const addLabel = () => {
  jobForm.value.labels.push({ key: '', value: '' })
}

const removeLabel = (index) => {
  jobForm.value.labels.splice(index, 1)
}

// 节点选择器管理
const addNodeSelector = () => {
  jobForm.value.node_selectors.push({ key: '', value: '' })
}

const removeNodeSelector = (index) => {
  jobForm.value.node_selectors.splice(index, 1)
}

// =========================
// 批量操作函数
// =========================
const enterBatchMode = () => {
  batchMode.value = true
  // 默认全选当前页，用户可取消不需要的项
  selectedJobs.value = [...paginatedJobs.value]
}

const exitBatchMode = () => {
  batchMode.value = false
  selectedJobs.value = []
}

const clearSelection = () => {
  selectedJobs.value = []
}

const isJobSelected = (job) => {
  return selectedJobs.value.some(j => j.name === job.name && j.namespace === job.namespace)
}

const toggleJobSelection = (job) => {
  const index = selectedJobs.value.findIndex(j => j.name === job.name && j.namespace === job.namespace)
  if (index >= 0) {
    selectedJobs.value.splice(index, 1)
  } else {
    selectedJobs.value.push(job)
  }
}

const isAllSelected = computed(() => {
  return paginatedJobs.value.length > 0 && 
         paginatedJobs.value.every(job => isJobSelected(job))
})

// 部分选中状态
const isPartialSelected = computed(() => {
  if (paginatedJobs.value.length === 0) return false
  const selectedCount = paginatedJobs.value.filter(job => isJobSelected(job)).length
  return selectedCount > 0 && selectedCount < paginatedJobs.value.length
})

// 全选复选框 ref
const selectAllCheckbox = ref(null)

// 设置 indeterminate 状态
watchEffect(() => {
  if (selectAllCheckbox.value) {
    selectAllCheckbox.value.indeterminate = isPartialSelected.value
  }
})

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    paginatedJobs.value.forEach(job => {
      const index = selectedJobs.value.findIndex(j => j.name === job.name && j.namespace === job.namespace)
      if (index >= 0) selectedJobs.value.splice(index, 1)
    })
  } else {
    paginatedJobs.value.forEach(job => {
      if (!isJobSelected(job)) {
        selectedJobs.value.push(job)
      }
    })
  }
}

// 批量删除预览
const openBatchDeletePreview = () => {
  deleteConfirmText.value = ''
  showBatchDeleteModal.value = true
}

const closeBatchDeleteModal = () => {
  showBatchDeleteModal.value = false
  deleteConfirmText.value = ''
}

// 执行批量删除
const executeBatchDelete = async () => {
  if (deleteConfirmText.value !== 'DELETE') return
  
  batchExecuting.value = true
  let successCount = 0
  let failCount = 0
  
  for (const job of selectedJobs.value) {
    try {
      await jobsApi.delete({
        name: job.name,
        namespace: job.namespace
      })
      successCount++
    } catch (e) {
      console.error(`Failed to delete ${job.name}:`, e)
      failCount++
    }
  }
  
  batchExecuting.value = false
  showBatchDeleteModal.value = false
  deleteConfirmText.value = ''
  
  if (failCount === 0) {
    Message.success({ content: `成功删除 ${successCount} 个 Job`, duration: 2200 })
  } else {
    Message.warning({ content: `成功 ${successCount} 个，失败 ${failCount} 个`, duration: 2200 })
  }
  
  exitBatchMode()
  refreshList()
}

// =========================
// 生命周期
// =========================
onMounted(async () => {
  await fetchNamespaces()
  await fetchJobs()
  setupAutoRefresh()
})

onBeforeUnmount(() => {
  clearAutoRefresh()
})

// =========================
// 数据获取
// =========================
const fetchNamespaces = async () => {
  try {
    const res = await namespacesApi.list({ page: 1, limit: 1000 })
    const list = res?.data?.list || res?.data?.items || []
    namespaces.value = (Array.isArray(list) ? list : []).map(ns =>
      typeof ns === 'string' ? ns : (ns?.metadata?.name || ns?.name || ns)
    ).filter(Boolean)
    
    // 如果没有获取到命名空间，使用默认值
    if (namespaces.value.length === 0) {
      namespaces.value = ['default', 'kube-system']
    }
  } catch (e) {
    console.error('获取命名空间列表失败:', e)
    // 失败时使用默认命名空间
    namespaces.value = ['default', 'kube-system']
  }
}

// 创建新命名空间
const createNewNamespace = async () => {
  if (!newNamespace.value || !newNamespace.value.trim()) {
    Message.error({ content: '请输入命名空间名称' })
    return
  }
  
  creatingNamespace.value = true
  try {
    const res = await namespaceApi.create({
      name: newNamespace.value.trim()
    })
    
    if (res.code === 0) {
      Message.success({ content: '命名空间创建成功' })
      // 刷新命名空间列表
      await fetchNamespaces()
      // 自动选择新创建的命名空间
      jobForm.value.namespace = newNamespace.value.trim()
      // 切换回选择模式
      showNamespaceInput.value = false
      newNamespace.value = ''
    } else {
      Message.error({ content: res.msg || '创建命名空间失败' })
    }
  } catch (e) {
    console.error('创建命名空间失败:', e)
    Message.error({ content: e?.msg || e?.message || '创建命名空间失败' })
  } finally {
    creatingNamespace.value = false
  }
}

const cancelCreateNamespace = () => {
  showNamespaceInput.value = false
  newNamespace.value = ''
}

const fetchJobs = async () => {
  loading.value = true
  errorMsg.value = ''
  
  try {
    const params = {
      page: currentPage.value,
      limit: itemsPerPage.value,
      namespace: namespaceFilter.value || '', // 空字符串表示所有命名空间
      name: searchQuery.value.trim()
    }

    const res = await jobsApi.list(params)
    
    if (res.code === 0) {
      // 格式化返回数据（后端返回 data.list，不是 data.items）
      const list = res.data?.list || res.data?.items || []
      jobs.value = list.map(item => ({
        name: item.name,
        namespace: item.namespace,
        status: item.status,
        image: item.image || (item.images && item.images[0]) || '',
        containers: item.containers || [],
        selector: item.selector || {},
        completions: item.completions,
        parallelism: item.parallelism,
        backoff_limit: item.backoff_limit,
        active: item.active,
        succeeded: item.succeeded,
        failed: item.failed,
        start_time: item.start_time,
        completion_time: item.completion_time,
        suspend: item.suspend || false
      }))
      total.value = res.data?.total || 0
    } else {
      errorMsg.value = res.msg || '获取Job列表失败'
      jobs.value = []
      total.value = 0
    }
  } catch (e) {
    console.error('获取 Job 列表失败:', e)
    errorMsg.value = e?.msg || e?.message || '获取Job列表失败'
    jobs.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

// =========================
// 筛选与搜索
// =========================
const setStatusFilter = (status) => {
  statusFilter.value = status
  currentPage.value = 1
  fetchJobs()
}

let searchDebounceTimer = null
const onSearchInput = () => {
  if (searchDebounceTimer) clearTimeout(searchDebounceTimer)
  searchDebounceTimer = setTimeout(() => {
    currentPage.value = 1
    fetchJobs()
  }, 500)
}

const refreshList = () => fetchJobs()

// 命名空间切换时重新加载
watch(namespaceFilter, () => {
  currentPage.value = 1
  fetchJobs()
})

// 分页变化时重新加载
watch(currentPage, () => {
  fetchJobs()
})

// =========================
// 计算属性
// =========================
const filteredJobs = computed(() => {
  let result = jobs.value
  
  // 状态过滤
  if (statusFilter.value !== 'all') {
    result = result.filter(job => job.status === statusFilter.value)
  }
  
  return result
})

const paginatedJobs = computed(() => {
  return filteredJobs.value
})

// 计算完成百分比
const getCompletionPercentage = (job) => {
  if (!job.completions || job.completions === 0) return 0
  return Math.min(((job.succeeded || 0) / job.completions) * 100, 100)
}

// 统计状态数量
const getStatusCount = (status) => {
  return jobs.value.filter(job => job.status === status).length
}

// =========================
// 自动刷新
// =========================
const setupAutoRefresh = () => {
  watch(autoRefresh, (enabled) => {
    if (enabled) {
      refreshTimer = setInterval(() => {
        fetchJobs()
      }, 90000) // 每90秒刷新一次
    } else {
      clearAutoRefresh()
    }
  })
}

const clearAutoRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

// =========================
// 更多菜单
// =========================
const toggleMoreOptions = (job, event) => {
  if (showMoreOptions.value && selectedJob.value === job) {
    showMoreOptions.value = false
    selectedJob.value = null
  } else {
    selectedJob.value = job
    showMoreOptions.value = true
    
    // 计算菜单位置
    const rect = event.target.getBoundingClientRect()
    menuStyle.value = {
      top: `${rect.bottom + window.scrollY + 5}px`,
      left: `${rect.left + window.scrollX - 150}px`
    }
  }
}

// 点击外部关闭菜单
onMounted(() => {
  document.addEventListener('click', (e) => {
    if (showMoreOptions.value && !e.target.closest('.more-btn')) {
      showMoreOptions.value = false
      selectedJob.value = null
    }
  })
})

// =========================
// Job 操作
// =========================
const createJob = async () => {
  if (!jobForm.value.name || !jobForm.value.container_image) {
    Message.error({ content: '请填写名称和镜像' })
    return
  }
  
  creating.value = true
  try {
    const res = await jobsApi.create(jobForm.value)
    if (res.code === 0) {
      Message.success({ content: 'Job 创建成功' })
      showCreateModal.value = false
      resetJobForm()
      await fetchJobs()
    } else {
      const errorMsg = res.details ? `${res.msg}: ${res.details}` : (res.msg || '创建失败')
      Message.error({ content: errorMsg, duration: 5000 })
    }
  } catch (e) {
    const errorMsg = e?.details ? `${e?.msg || '创建失败'}: ${e.details}` : (e?.msg || '创建失败')
    Message.error({ content: errorMsg, duration: 5000 })
  } finally {
    creating.value = false
  }
}

const resetJobForm = () => {
  currentStep.value = 1
  jobNameError.value = ''
  createMode.value = 'form' // 重置为表单模式
  yamlContent.value = '' // 清空 YAML 内容
  yamlError.value = '' // 清空 YAML 错误
  showNamespaceInput.value = false
  newNamespace.value = ''
  jobForm.value = {
    namespace: 'default',
    name: '',
    description: '',
    container_image: '',
    container_name: '',
    container_command: null,
    container_command_args: null,
    completions: 1,
    parallelism: 1,
    backoff_limit: 6,
    active_deadline_seconds: null,
    ttl_seconds_after_finished: null,
    restart_policy: 'OnFailure',
    image_pull_policy: '',
    image_pull_secret: null,
    cpu_requirement: null,
    memory_requirement: null,
    variables: [],
    labels: [],
    node_selectors: []
  }
}

const viewJobDetail = async (job) => {
  showMoreOptions.value = false
  loadingDetail.value = true
  showDetailModal.value = true
  
  try {
    const res = await jobsApi.detail({ namespace: job.namespace, name: job.name })
    detailData.value = res.code === 0 ? res.data : job
  } catch (e) {
    console.error('获取详情失败:', e)
    detailData.value = job
  } finally {
    loadingDetail.value = false
  }
}

const restartJob = async (job) => {
  showMoreOptions.value = false
  
  if (!confirm(`确定要重启 Job "${job.name}" 吗？这将创建一个新的Job实例。`)) {
    return
  }
  
  try {
    const res = await jobsApi.restart({ namespace: job.namespace, name: job.name })
    if (res.code === 0) {
      Message.success({ content: `Job 已重启为: ${res.data.newJob || '新Job'}` })
      await fetchJobs()
    } else {
      Message.error({ content: res.msg || '重启失败' })
    }
  } catch (e) {
    Message.error({ content: e?.msg || '重启失败' })
  }
}

const suspendJob = async (job) => {
  showMoreOptions.value = false
  
  const action = job.suspend ? '恢复' : '暂停'
  
  try {
    const res = await jobsApi.suspend({
      namespace: job.namespace,
      name: job.name,
      suspend: !job.suspend
    })
    if (res.code === 0) {
      Message.success({ content: `Job 已${action}` })
      await fetchJobs()
    } else {
      Message.error({ content: res.msg || `${action}失败` })
    }
  } catch (e) {
    Message.error({ content: e?.msg || `${action}失败` })
  }
}

const deleteJob = async (job) => {
  showMoreOptions.value = false
  
  if (!confirm(`确定要删除 Job "${job.name}" 吗？此操作无法撤销。`)) {
    return
  }
  
  try {
    const res = await jobsApi.delete({ namespace: job.namespace, name: job.name })
    if (res.code === 0) {
      Message.success({ content: 'Job 删除成功' })
      await fetchJobs()
    } else {
      Message.error({ content: res.msg || '删除失败' })
    }
  } catch (e) {
    Message.error({ content: e?.msg || '删除失败' })
  }
}

const fmtTime = (ts) => {
  if (!ts) return '-'
  try {
    const d = new Date(ts)
    return d.toLocaleString('zh-CN')
  } catch {
    return ts
  }
}

// =========================
// Rancher/Kuboard 风格增强功能
// =========================

// 计算创建时长（类似 Kuboard）
const getAge = (createdAt) => {
  if (!createdAt) return '-'
  try {
    const created = new Date(createdAt)
    const now = new Date()
    const diffMs = now - created
    const diffMins = Math.floor(diffMs / 60000)
    const diffHours = Math.floor(diffMs / 3600000)
    const diffDays = Math.floor(diffMs / 86400000)
    
    if (diffMins < 1) return '刚刚'
    if (diffMins < 60) return `${diffMins}分钟前`
    if (diffHours < 24) return `${diffHours}小时前`
    if (diffDays < 30) return `${diffDays}天前`
    return `${Math.floor(diffDays / 30)}个月前`
  } catch {
    return createdAt
  }
}

// 计算执行耗时
const getDuration = (startTime, endTime) => {
  if (!startTime || !endTime) return '-'
  try {
    const start = new Date(startTime)
    const end = new Date(endTime)
    const diffMs = end - start
    const diffSecs = Math.floor(diffMs / 1000)
    const diffMins = Math.floor(diffMs / 60000)
    const diffHours = Math.floor(diffMs / 3600000)
    
    if (diffSecs < 60) return `${diffSecs}秒`
    if (diffMins < 60) return `${diffMins}分${diffSecs % 60}秒`
    return `${diffHours}小时${diffMins % 60}分`
  } catch {
    return '-'
  }
}

// 获取进度条颜色类（Rancher 风格）
const getProgressClass = (job) => {
  const percent = getCompletionPercentage(job)
  if (job.status === 'Complete') return 'progress-success'
  if (job.status === 'Failed') return 'progress-failed'
  if (percent > 75) return 'progress-high'
  if (percent > 25) return 'progress-medium'
  return 'progress-low'
}

// 查看 Job 日志（实现）
const viewJobLogs = async (job) => {
  showMoreOptions.value = false
  selectedJobForLogs.value = job
  jobPodsList.value = []
  jobContainerList.value = []
  jobLogsContent.value = ''
  jobLogsError.value = ''
  jobLogsForm.value = {
    selectedPod: '',
    container: '',
    tail: 100,
    follow: false
  }
  showJobLogsModal.value = true
  
  // 获取 Job 关联的 Pod 列表
  try {
    const res = await fetch(`/api/v1/k8s/pod/list?namespace=${job.namespace}&limit=100`, {
      headers: getAuthHeaders()
    })
    if (res.ok) {
      const data = await res.json()
      // 过滤出属于该 Job 的 Pod（通过 Job 名称前缀匹配）
      const jobPods = (data.data?.list || []).filter(pod => 
        pod.name.startsWith(job.name + '-')
      )
      jobPodsList.value = jobPods
      
      // 如果有 Pod，默认选中第一个
      if (jobPods.length > 0) {
        jobLogsForm.value.selectedPod = jobPods[0].name
        jobContainerList.value = jobPods[0].containers || []
        if (jobContainerList.value.length === 1) {
          jobLogsForm.value.container = jobContainerList.value[0]
        }
      }
    }
  } catch (e) {
    console.error('获取 Pod 列表失败:', e)
  }
}

// Pod 选择变化
const onJobPodChange = () => {
  const selectedPod = jobPodsList.value.find(p => p.name === jobLogsForm.value.selectedPod)
  if (selectedPod) {
    jobContainerList.value = selectedPod.containers || []
    // 单容器自动选中
    if (jobContainerList.value.length === 1) {
      jobLogsForm.value.container = jobContainerList.value[0]
    } else {
      jobLogsForm.value.container = ''
    }
  } else {
    jobContainerList.value = []
    jobLogsForm.value.container = ''
  }
}

// 获取 Job 日志
const fetchJobLogs = async () => {
  if (!selectedJobForLogs.value) return
  
  // 如果选择了单个 Pod，需要验证容器
  if (jobLogsForm.value.selectedPod) {
    let container = jobLogsForm.value.container || ''
    
    if (jobContainerList.value.length === 1) {
      container = jobContainerList.value[0]
      jobLogsForm.value.container = container
    }
    
    if (jobContainerList.value.length > 1 && !container) {
      jobLogsError.value = '请选择容器'
      return
    }
    
    if (!container) {
      jobLogsError.value = '无法确定容器名称'
      return
    }
  }
  
  stopJobLogStream()
  loadingJobLogs.value = true
  jobLogsError.value = ''
  jobLogsContent.value = ''
  
  try {
    if (jobLogsForm.value.follow) {
      await fetchJobStreamLogs()
    } else {
      await fetchJobStaticLogs()
    }
  } catch (e) {
    if (e.name !== 'AbortError') {
      jobLogsError.value = e?.msg || e?.message || '获取日志失败'
    }
  } finally {
    if (!jobLogsForm.value.follow) {
      loadingJobLogs.value = false
    }
  }
}

// 获取静态日志
const fetchJobStaticLogs = async () => {
  if (jobLogsForm.value.selectedPod) {
    // 获取单个 Pod 的日志
    const params = new URLSearchParams({
      namespace: selectedJobForLogs.value.namespace,
      name: jobLogsForm.value.selectedPod,
      container: jobLogsForm.value.container
    })
    if (jobLogsForm.value.tail != null) {
      params.set('tail', jobLogsForm.value.tail)
    }
    
    jobLogAbortController = new AbortController()
    const response = await fetch(`/api/v1/k8s/pod/container_logs?${params}`, {
      signal: jobLogAbortController.signal,
      headers: getAuthHeaders()
    })
    
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    
    const res = await response.json()
    jobLogsContent.value = res?.data?.log || '暂无日志'
  } else {
    // 获取所有 Pod 的日志（汇总）
    const pods = jobPodsList.value.slice(0, 5)
    const logsArray = []
    
    for (const pod of pods) {
      const container = pod.containers?.[0] || ''
      const params = new URLSearchParams({
        namespace: selectedJobForLogs.value.namespace,
        name: pod.name,
        container
      })
      if (jobLogsForm.value.tail != null) {
        params.set('tail', jobLogsForm.value.tail)
      }
      
      try {
        const response = await fetch(`/api/v1/k8s/pod/container_logs?${params}`, {
          headers: getAuthHeaders()
        })
        if (response.ok) {
          const res = await response.json()
          logsArray.push(`\n========== Pod: ${pod.name} | 容器: ${container} ==========\n${res?.data?.log || '（无日志）'}`)
        }
      } catch (e) {
        logsArray.push(`\n========== Pod: ${pod.name} ==========\n获取日志失败: ${e.message}`)
      }
    }
    
    jobLogsContent.value = logsArray.join('\n') || '没有找到关联的 Pod'
  }
}

// 获取流式日志
const fetchJobStreamLogs = async () => {
  if (!jobLogsForm.value.selectedPod) {
    jobLogsError.value = '实时日志需要选择单个 Pod'
    loadingJobLogs.value = false
    return
  }
  
  isStreamingJobLogs.value = true
  jobLogAbortController = new AbortController()
  
  const params = new URLSearchParams({
    namespace: selectedJobForLogs.value.namespace,
    name: jobLogsForm.value.selectedPod,
    container: jobLogsForm.value.container,
    follow: 'true'
  })
  if (jobLogsForm.value.tail != null) {
    params.set('tail', jobLogsForm.value.tail)
  }
  
  const response = await fetch(`/api/v1/k8s/pod/container_log?${params}`, {
    signal: jobLogAbortController.signal,
    headers: getAuthHeaders()
  })
  
  if (!response.ok) throw new Error(`HTTP ${response.status}`)
  
  const reader = response.body.getReader()
  const decoder = new TextDecoder('utf-8')
  
  while (true) {
    const { done, value } = await reader.read()
    if (done) break
    
    const chunk = decoder.decode(value, { stream: true })
    jobLogsContent.value += chunk
    
    // 自动滚动到底部
    if (jobLogsContentRef.value) {
      jobLogsContentRef.value.scrollTop = jobLogsContentRef.value.scrollHeight
    }
  }
  
  isStreamingJobLogs.value = false
  loadingJobLogs.value = false
}

// 停止日志流
const stopJobLogStream = () => {
  if (jobLogAbortController) {
    jobLogAbortController.abort()
    jobLogAbortController = null
  }
  isStreamingJobLogs.value = false
  loadingJobLogs.value = false
}

// 清除日志
const clearJobLogs = () => {
  jobLogsContent.value = ''
  jobLogsError.value = ''
}

// 关闭日志模态框
const closeJobLogs = () => {
  stopJobLogStream()
  showJobLogsModal.value = false
  jobLogsContent.value = ''
  jobLogsError.value = ''
  jobPodsList.value = []
  jobContainerList.value = []
}

// Job 日志高亮处理
const highlightedJobLogs = computed(() => {
  if (!jobLogsContent.value) {
    return '<span class="log-placeholder">暂无日志，请点击"获取日志"按钮</span>'
  }
  
  // 转义 HTML 特殊字符
  const escapeHtml = (str) => {
    return str
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
  }
  
  const escaped = escapeHtml(jobLogsContent.value)
  
  // 按行处理
  return escaped.split('\n').map(line => {
    // 时间戳高亮
    let highlighted = line.replace(
      /(\d{4}-\d{2}-\d{2}[T\s]\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:?\d{2})?)/g,
      '<span class="log-timestamp">$1</span>'
    )
    
    // ERROR / FATAL / PANIC 高亮（红色）
    if (/\b(ERROR|FATAL|PANIC|error|fatal|panic)\b/.test(highlighted)) {
      return `<span class="log-error">${highlighted}</span>`
    }
    
    // WARN 高亮（橙色）
    if (/\b(WARN|WARNING|warn|warning)\b/.test(highlighted)) {
      return `<span class="log-warn">${highlighted}</span>`
    }
    
    // INFO 高亮（蓝色）
    if (/\b(INFO|info)\b/.test(highlighted)) {
      return `<span class="log-info">${highlighted}</span>`
    }
    
    // DEBUG 高亮（灰色）
    if (/\b(DEBUG|debug)\b/.test(highlighted)) {
      return `<span class="log-debug">${highlighted}</span>`
    }
    
    return highlighted
  }).join('\n')
})

// 查看 Job YAML（新增）
const viewJobYaml = async (job) => {
  showMoreOptions.value = false
  selectedYamlJob.value = job
  yamlViewContent.value = ''
  yamlViewError.value = ''
  yamlEditMode.value = false
  showYamlModal.value = true
  loadingYaml.value = true
  
  try {
    const res = await jobsApi.getYaml({
      namespace: job.namespace,
      name: job.name
    })
    if (res.code === 0 && res.data?.yaml) {
      yamlViewContent.value = res.data.yaml
    } else {
      yamlViewError.value = res.msg || '获取 YAML 失败'
    }
  } catch (e) {
    yamlViewError.value = e?.msg || e?.message || '获取 YAML 失败'
  } finally {
    loadingYaml.value = false
  }
}

// 关闭 YAML 模态框
const closeYamlModal = () => {
  showYamlModal.value = false
  selectedYamlJob.value = null
  yamlViewContent.value = ''
  yamlViewError.value = ''
  yamlEditMode.value = false
}

// 下载 YAML
const downloadYaml = () => {
  if (!yamlViewContent.value || !selectedYamlJob.value) {
    Message.warning({ content: '没有可下载的 YAML 内容' })
    return
  }
  
  try {
    const blob = new Blob([yamlViewContent.value], { type: 'text/yaml;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `${selectedYamlJob.value.name}-job.yaml`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
    Message.success({ content: 'YAML 文件已下载' })
  } catch (e) {
    Message.error({ content: '下载失败' })
  }
}

// 应用 YAML 修改
const applyYamlChanges = async () => {
  if (!yamlViewContent.value?.trim()) {
    Message.error({ content: '请输入 YAML 内容' })
    return
  }
  savingYaml.value = true
  try {
    const res = await jobsApi.applyYaml({
      namespace: selectedYamlJob.value.namespace,
      name: selectedYamlJob.value.name,
      yaml: yamlViewContent.value
    })
    if (res.code === 0) {
      Message.success({ content: 'YAML 应用成功' })
      closeYamlModal()
      fetchJobs()
    } else {
      Message.error({ content: res.msg || '应用 YAML 失败' })
    }
  } catch (e) {
    Message.error({ content: e?.msg || e?.message || '应用 YAML 失败' })
  } finally {
    savingYaml.value = false
  }
}

// =========================
// 内联编辑功能
// =========================

// 内联编辑 - 开始编辑镜像
const startInlineImage = (job) => {
  inlineEdit.value = {
    key: `image-${job.name}-${job.namespace}`,
    value: job.image,
    original: job.image,
    job: job,
    container: job.containers?.[0] || ''  // containers 是字符串数组，不是对象数组
  }
}

// 内联编辑 - 保存镜像
const saveInlineImage = async (job) => {
  const newImage = inlineEdit.value.value?.trim()
  const oldImage = inlineEdit.value.original
  const container = inlineEdit.value.container
  
  // 如果没有变化，直接取消编辑
  if (newImage === oldImage || !newImage) {
    cancelInlineEdit()
    return
  }
  
  // 二次确认 - 镜像更新是高危操作
  if (!confirm(`⚠️ 确认更新镜像？

Job: ${job.namespace}/${job.name}
容器: ${container}
旧镜像: ${oldImage}
新镜像: ${newImage}

⚠️ 注意: Job 不支持滚动更新！
更新镜像只会修改模板，已运行的 Pod 不会自动重启。
如需使用新镜像，请手动删除 Pod 或重新创建 Job。`)) {
    cancelInlineEdit()
    return
  }
  
  try {
    const res = await jobsApi.updateImage({
      namespace: job.namespace,
      name: job.name,
      container: container,
      image: newImage
    })
    if (res.code === 0) {
      Message.success({ content: '镜像已更新（需手动重启 Pod 才能生效）' })
      cancelInlineEdit()
      await fetchJobs()
    } else {
      Message.error({ content: res.msg || '更新镜像失败' })
    }
  } catch (e) {
    console.error('更新镜像失败:', e)
    Message.error({ content: e?.msg || e?.message || '更新镜像失败' })
  }
}

// 内联编辑 - 取消编辑
const cancelInlineEdit = () => {
  inlineEdit.value = { key: '', value: null, original: null, job: null, container: '' }
}

// =========================
// YAML 创建相关功能
// =========================

// 加载 Job YAML 模板
const loadJobYamlTemplate = () => {
  yamlContent.value = `apiVersion: batch/v1
kind: Job
metadata:
  name: example-job
  namespace: default
  labels:
    app: example
spec:
  completions: 1
  parallelism: 1
  backoffLimit: 6
  template:
    metadata:
      labels:
        app: example
    spec:
      restartPolicy: OnFailure
      containers:
      - name: example-container
        image: busybox:latest
        command: ["/bin/sh"]
        args: ["-c", "echo Hello from Job && sleep 30"]
        env:
        - name: EXAMPLE_ENV
          value: "example-value"
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "128Mi"
            cpu: "200m"`
  yamlError.value = ''
  Message.success({ content: '已加载 YAML 模板，请修改后创建' })
}

// 加载多资源 YAML 模板
const loadMultiResourceYamlTemplate = () => {
  yamlContent.value = `# 多资源 Job 示例
# 使用 --- 分隔不同的资源对象

# 1. ConfigMap - 存储配置数据
apiVersion: v1
kind: ConfigMap
metadata:
  name: job-config
  namespace: default
  labels:
    app: example-job
    component: config
    version: v1.0
    tier: backend
    environment: production
data:
  config.properties: |
    # Job 配置文件
    job.name=example-job
    job.timeout=300
    log.level=INFO
    database.url=jdbc:mysql://mysql-service:3306/mydb
    database.username=user
    database.password=password
  startup.sh: |
    #!/bin/bash
    echo "Starting job with config..."
    cat /config/config.properties
    echo "Job processing started at $(date)"
    # 模拟工作负载
    sleep 30
    echo "Job completed successfully at $(date)"
---
# 2. Secret - 敏感数据
apiVersion: v1
kind: Secret
metadata:
  name: job-secret
  namespace: default
  labels:
    app: example-job
    component: secret
    version: v1.0
    tier: backend
    environment: production
type: Opaque
data:
  # echo -n 'mysql-password' | base64
  database-password: bXlzcWwtcGFzc3dvcmQ=
  # echo -n 'api-key-12345' | base64
  api-key: YXBpLWtleS0xMjM0NQ==
---
# 3. PVC - 持久化存储
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: job-storage
  namespace: default
  labels:
    app: example-job
    component: storage
    version: v1.0
    tier: backend
    environment: production
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
  storageClassName: local-path
---
# 4. Service - 内部服务发现
apiVersion: v1
kind: Service
metadata:
  name: job-metrics-service
  namespace: default
  labels:
    app: example-job
    component: metrics
    version: v1.0
    tier: backend
    environment: production
spec:
  selector:
    app: example-job
    component: worker
  ports:
    - protocol: TCP
      port: 8080
      targetPort: metrics
      name: metrics-port
  type: ClusterIP
---
# 5. Job - 主要工作负载
apiVersion: batch/v1
kind: Job
metadata:
  name: example-job
  namespace: default
  labels:
    app: example-job
    component: worker
    version: v1.0
    tier: backend
    environment: production
spec:
  completions: 1
  parallelism: 1
  backoffLimit: 3
  activeDeadlineSeconds: 600
  ttlSecondsAfterFinished: 3600
  template:
    metadata:
      labels:
        app: example-job
        component: worker
        version: v1.0
        tier: backend
        environment: production
    spec:
      restartPolicy: OnFailure
      volumes:
        - name: config-volume
          configMap:
            name: job-config
        - name: storage-volume
          persistentVolumeClaim:
            claimName: job-storage
      containers:
        - name: job-container
          image: busybox:latest
          command: ["/bin/sh"]
          args: ["-c", "/config/startup.sh"]
          volumeMounts:
            - name: config-volume
              mountPath: /config
            - name: storage-volume
              mountPath: /data
          env:
            - name: CONFIG_PATH
              value: "/config/config.properties"
            - name: DATABASE_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: job-secret
                  key: database-password
            - name: API_KEY
              valueFrom:
                secretKeyRef:
                  name: job-secret
                  key: api-key
          resources:
            requests:
              memory: "128Mi"
              cpu: "200m"
            limits:
              memory: "256Mi"
              cpu: "500m"`
  yamlError.value = ''
  Message.success({ content: '已加载多资源 YAML 模板，请修改后创建' })
}

// 复制 YAML 内容到剪贴板
const copyYamlContent = async () => {
  if (!yamlContent.value.trim()) {
    Message.warning({ content: '没有内容可复制' })
    return
  }
  try {
    await navigator.clipboard.writeText(yamlContent.value)
    Message.success({ content: 'YAML 内容已复制到剪贴板' })
  } catch (err) {
    console.error('复制失败:', err)
    Message.error({ content: '复制失败，请手动复制' })
  }
}

// 重置 YAML 内容
const resetYamlContent = () => {
  if (yamlContent.value.trim() && !confirm('确定要重置 YAML 内容吗？')) {
    return
  }
  yamlContent.value = ''
  yamlError.value = ''
  // 重置多资源预览
  multiResourcePreview.value = {
    resources: [],
    dependencies: [],
    errors: []
  }
  Message.success({ content: 'YAML 内容已重置' })
}

// 一键创建所有多资源
const applyMultiResourceYaml = async () => {
  if (!yamlContent.value.trim()) {
    Message.error({ content: '请输入 YAML 内容' })
    return
  }

  if (multiResourcePreview.value.errors.length > 0) {
    Message.error({ content: '存在解析错误，请修正后再创建' })
    return
  }

  if (multiResourcePreview.value.resources.length === 0) {
    Message.error({ content: '未检测到有效的资源' })
    return
  }

  creating.value = true
  try {
    const res = await fetch('/api/v1/k8s/multi-resource/apply-yaml', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...getAuthHeaders()
      },
      body: JSON.stringify({
        yaml: yamlContent.value,
        cluster_id: 1 // 默认集群ID
      })
    })

    if (res.ok) {
      const data = await res.json()
      if (data.code === 0) {
        Message.success({ 
          content: `成功创建 ${data.data.created_count} 个资源，跳过 ${data.data.skipped_count} 个已存在的资源` 
        })
        showCreateModal.value = false
        resetJobForm()
        await fetchJobs()
      } else {
        Message.error({ content: data.msg || '创建失败' })
        yamlError.value = data.msg || '创建失败'
      }
    } else {
      const errorData = await res.json().catch(() => ({}))
      Message.error({ content: errorData.msg || `创建失败: HTTP ${res.status}` })
      yamlError.value = errorData.msg || `创建失败: HTTP ${res.status}`
    }
  } catch (e) {
    console.error('创建多资源失败:', e)
    Message.error({ content: `创建失败: ${e.message}` })
    yamlError.value = `创建失败: ${e.message}`
  } finally {
    creating.value = false
  }
}

// 获取资源图标
const getResourceIcon = (kind) => {
  const icons = {
    'Job': '⚙️',
    'ConfigMap': '📋',
    'Secret': '🔒',
    'PersistentVolumeClaim': '💾',
    'Service': '🌐',
    'Deployment': '🚢',
    'StatefulSet': '🔢',
    'DaemonSet': '👾',
    'Ingress': '🚪',
    'Namespace': '📂'
  }
  return icons[kind] || '📄'
}

// 获取资源种类样式类
const getResourceKindClass = (kind) => {
  const classes = {
    'Job': 'job-resource',
    'ConfigMap': 'configmap-resource',
    'Secret': 'secret-resource',
    'PersistentVolumeClaim': 'pvc-resource',
    'Service': 'service-resource',
    'Deployment': 'deployment-resource',
    'StatefulSet': 'statefulset-resource',
    'DaemonSet': 'daemonset-resource'
  }
  return classes[kind] || 'default-resource'
}

// 从 YAML 创建 Job
const createJobFromYaml = async () => {
  if (!yamlContent.value.trim()) {
    Message.error({ content: '请输入 YAML 内容' })
    return
  }
  
  // 简单验证 YAML 格式
  try {
    // 检查必要字段
    if (!yamlContent.value.includes('kind: Job')) {
      yamlError.value = 'YAML 中必须包含 "kind: Job"'
      return
    }
    if (!yamlContent.value.includes('apiVersion: batch/v1')) {
      yamlError.value = 'YAML 中必须包含 "apiVersion: batch/v1"'
      return
    }
    
    yamlError.value = ''
  } catch (e) {
    yamlError.value = `YAML 格式错误: ${e.message}`
    return
  }
  
  creating.value = true
  try {
    const res = await jobsApi.createFromYaml({ yaml: yamlContent.value })
    if (res.code === 0) {
      Message.success({ content: 'Job 创建成功' })
      showCreateModal.value = false
      resetJobForm()
      await fetchJobs()
    } else {
      const errorMsg = res.details ? `${res.msg}: ${res.details}` : (res.msg || '创建失败')
      Message.error({ content: errorMsg, duration: 5000 })
      yamlError.value = errorMsg
    }
  } catch (e) {
    const errorMsg = e?.details ? `${e?.msg || '创建失败'}: ${e.details}` : (e?.msg || e?.message || '创建失败')
    Message.error({ content: errorMsg, duration: 5000 })
    yamlError.value = errorMsg
  } finally {
    creating.value = false
  }
}
</script>

<style scoped>
/* 与 Deployments.vue 相同的样式，这里只列出关键样式 */
.resource-view {
  padding: 24px;
  background: linear-gradient(135deg, #f5f7fa 0%, #e8eef5 100%);
  min-height: 100vh;
}

.view-header {
  margin-bottom: 24px;
}

.view-header h1 {
  font-size: 28px;
  font-weight: 700;
  color: #2c3e50;
  margin: 0 0 8px 0;
}

.view-header p {
  color: #7f8c8d;
  font-size: 14px;
  margin: 0;
}

.action-bar {
  display: flex;
  gap: 16px;
  align-items: center;
  margin-bottom: 24px;
  flex-wrap: wrap;
  background: white;
  padding: 20px;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.search-box {
  flex: 1;
  min-width: 250px;
}

.search-box input {
  width: 100%;
  padding: 10px 16px;
  border: 2px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  transition: all 0.2s;
}

.search-box input:focus {
  outline: none;
  border-color: #326ce5;
  box-shadow: 0 0 0 3px rgba(50, 108, 229, 0.1);
}

.filter-buttons {
  display: flex;
  gap: 8px;
}

.btn {
  padding: 8px 16px;
  border: 2px solid transparent;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-filter {
  background: #f1f5f9;
  color: #64748b;
}

.btn-filter:hover {
  background: #e2e8f0;
  color: #475569;
}

.btn-filter.active {
  background: #326ce5;
  color: white;
  border-color: #326ce5;
}

.filter-dropdown select {
  padding: 8px 16px;
  border: 2px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  cursor: pointer;
  background: white;
  transition: all 0.2s;
}

.filter-dropdown select:focus {
  outline: none;
  border-color: #326ce5;
}

.action-buttons {
  display: flex;
  gap: 12px;
  align-items: center;
}

.view-toggle {
  display: flex;
  gap: 4px;
  background: #f1f5f9;
  padding: 4px;
  border-radius: 8px;
}

.btn-view {
  padding: 6px 12px;
  background: transparent;
  border: none;
  border-radius: 6px;
  font-size: 16px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-view.active {
  background: white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.auto-refresh-toggle {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  background: #f1f5f9;
  border-radius: 8px;
  cursor: pointer;
  user-select: none;
}

.auto-refresh-toggle input[type="checkbox"] {
  cursor: pointer;
}

.refresh-indicator {
  color: #22c55e;
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.btn-primary {
  background: #326ce5;
  color: white;
  padding: 10px 20px;
}

.btn-primary:hover {
  background: #2557c7;
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(50, 108, 229, 0.3);
}

.btn-secondary {
  background: #64748b;
  color: white;
  padding: 10px 20px;
}

.btn-secondary:hover {
  background: #475569;
}

.btn-secondary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.error-box {
  background: #fee2e2;
  border: 1px solid #fca5a5;
  color: #991b1b;
  padding: 12px 16px;
  border-radius: 8px;
  margin-bottom: 16px;
}

/* 数据统计栏 */
.stats-bar {
  display: flex;
  gap: 24px;
  background: white;
  padding: 16px 24px;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.stat-label {
  font-size: 14px;
  color: #64748b;
  font-weight: 500;
}

.stat-value {
  font-size: 18px;
  font-weight: 700;
  color: #1e293b;
}

.stat-value.success {
  color: #16a34a;
}

.stat-value.running {
  color: #2563eb;
}

.stat-value.failed {
  color: #dc2626;
}

.table-container {
  background: white;
  border-radius: 12px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.05);
  overflow-x: auto; /* 水平滚动 */
  overflow-y: auto; /* 垂直滚动 */
  max-height: 600px; /* 最大高度，超过后显示垂直滚动条 */
  position: relative;
}

/* 表格滚动条样式（水平 + 垂直） */
.table-container::-webkit-scrollbar {
  width: 12px; /* 垂直滚动条宽度 */
  height: 12px; /* 水平滚动条高度 */
}

.table-container::-webkit-scrollbar-track {
  background: #f1f5f9;
  border-radius: 6px;
  margin: 4px;
}

.table-container::-webkit-scrollbar-thumb {
  background: linear-gradient(135deg, #cbd5e1 0%, #94a3b8 100%);
  border-radius: 6px;
  border: 2px solid #f1f5f9;
  transition: all 0.3s;
}

.table-container::-webkit-scrollbar-thumb:hover {
  background: linear-gradient(135deg, #94a3b8 0%, #64748b 100%);
  border-color: #e2e8f0;
}

/* 滚动条交叉区域 */
.table-container::-webkit-scrollbar-corner {
  background: #f1f5f9;
  border-radius: 6px;
}

.resource-table {
  width: 100%;
  border-collapse: collapse;
  min-width: 1600px;
  table-layout: auto;
}

.resource-table th {
  background-color: #f7fafc;
  text-align: left;
  padding: 16px 20px;
  font-size: 14px;
  font-weight: 600;
  color: #4a5568;
  border-bottom: 1px solid #e2e8f0;
}

.resource-table td {
  padding: 16px 20px;
  font-size: 14px;
  color: #2d3748;
  border-bottom: 1px solid #f7fafc;
  vertical-align: middle;
}

.resource-table tbody tr:hover {
  background-color: #f9fafb;
  transition: background-color 0.2s;
  cursor: pointer;
}

.resource-table tbody tr {
  transition: all 0.2s;
}

.resource-table tbody tr:hover .action-btn {
  transform: scale(1.05);
}

.status-indicator {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
}

/* Rancher 风格：状态点指示器 */
.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: currentColor;
  animation: statusPulse 2s ease-in-out infinite;
}

@keyframes statusPulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.6; }
}

/* 状态单元格悬停提示 */
.status-cell {
  position: relative;
}

.status-tooltip {
  display: none;
  position: absolute;
  top: 100%;
  left: 0;
  z-index: 100;
  background: rgba(0, 0, 0, 0.9);
  color: white;
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 12px;
  white-space: nowrap;
  margin-top: 4px;
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
}

.status-cell:hover .status-tooltip {
  display: block;
}

.status-tooltip > div {
  padding: 2px 0;
}

.status-indicator.complete {
  background: #d1fae5;
  color: #065f46;
}

.status-indicator.running {
  background: #dbeafe;
  color: #1e40af;
}

.status-indicator.failed {
  background: #fee2e2;
  color: #991b1b;
}

.job-name {
  display: flex;
  align-items: center;
  gap: 8px;
}

.job-name .icon {
  font-size: 18px;
}

.namespace-badge {
  display: inline-block;
  padding: 4px 10px;
  background: #f1f5f9;
  border-radius: 6px;
  font-size: 12px;
  color: #475569;
  font-weight: 500;
}

.completion-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.completion-text {
  font-size: 13px;
  color: #64748b;
}

.completion-bar {
  width: 100%;
  height: 6px;
  background: #e2e8f0;
  border-radius: 3px;
  overflow: hidden;
  position: relative;
}

.completion-fill {
  height: 100%;
  background: linear-gradient(90deg, #10b981 0%, #059669 100%);
  border-radius: 3px;
  transition: width 0.6s ease;
  position: relative;
  overflow: hidden;
}

/* Rancher 风格：进度条颜色分级 */
.completion-fill.progress-success {
  background: linear-gradient(90deg, #10b981 0%, #059669 100%);
}

.completion-fill.progress-failed {
  background: linear-gradient(90deg, #ef4444 0%, #dc2626 100%);
}

.completion-fill.progress-high {
  background: linear-gradient(90deg, #3b82f6 0%, #2563eb 100%);
}

.completion-fill.progress-medium {
  background: linear-gradient(90deg, #f59e0b 0%, #d97706 100%);
}

.completion-fill.progress-low {
  background: linear-gradient(90deg, #94a3b8 0%, #64748b 100%);
}

.progress-shimmer {
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(
    90deg,
    transparent 0%,
    rgba(255, 255, 255, 0.4) 50%,
    transparent 100%
  );
  animation: shimmer 2s infinite;
}

.completion-fill::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(
    90deg,
    transparent 0%,
    rgba(255, 255, 255, 0.3) 50%,
    transparent 100%
  );
  animation: shimmer 2s infinite;
}

@keyframes shimmer {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(100%); }
}

.image-text {
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.image-name {
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 12px;
  color: #6366f1;
}

.selector-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.selector-tag {
  display: inline-block;
  padding: 3px 8px;
  background: #fef3c7;
  border: 1px solid #fbbf24;
  border-radius: 4px;
  font-size: 11px;
  font-family: 'Monaco', 'Menlo', monospace;
  color: #92400e;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.selector-empty {
  color: #94a3b8;
  font-size: 12px;
}

/* Kuboard 风格：时间徽章 */
.age-badge {
  display: inline-block;
  margin-left: 8px;
  padding: 2px 6px;
  background: #e0e7ff;
  color: #4338ca;
  border-radius: 4px;
  font-size: 10px;
  font-weight: 500;
}

/* 并行度活跃指示器 */
.parallelism-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.active-indicator {
  display: inline-flex;
  align-items: center;
  padding: 2px 6px;
  background: #fef3c7;
  color: #d97706;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
}

/* 警告文本 */
.warning-text {
  color: #dc2626;
  font-weight: 600;
}

/* 时间信息 */
.time-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.time-duration {
  font-size: 11px;
  color: #64748b;
  font-style: italic;
}

/* 完成信息增强 */
.completion-stats {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.completion-percent {
  font-size: 11px;
  color: #64748b;
  font-weight: 600;
}

.action-icons {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: nowrap;
}

.action-btn {
  padding: 6px 12px;
  border: none;
  border-radius: 6px;
  font-size: 12px;
  cursor: pointer;
  background: #e2e8f0;
  color: #4a5568;
  transition: all 0.2s;
  white-space: nowrap;
}

.action-btn.primary {
  background: #326ce5;
  color: white;
}

/* Rancher 风格：图标按钮 */
.action-btn.icon-only {
  padding: 6px 10px;
  font-size: 16px;
  min-width: unset;
  background: transparent;
  border: 1px solid #e2e8f0;
}

.action-btn.icon-only:hover {
  background: #f1f5f9;
  border-color: #326ce5;
  color: #326ce5;
  transform: scale(1.1);
}

.action-btn:hover {
  transform: translateY(-1px);
}

.icon-btn {
  background: none;
  border: none;
  font-size: 14px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  white-space: nowrap;
}

.icon-btn:hover {
  background-color: #e2e8f0;
}

.more-btn {
  position: relative;
}

.more-menu {
  position: fixed;
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  padding: 6px;
  min-width: 180px;
  z-index: 1000;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 10px 14px;
  border: none;
  background: none;
  text-align: left;
  cursor: pointer;
  border-radius: 6px;
  font-size: 14px;
  color: #374151;
  transition: all 0.2s;
}

.menu-item:hover {
  background: #f3f4f6;
}

.menu-item.danger {
  color: #dc2626;
}

.menu-item.danger:hover {
  background: #fee2e2;
}

.menu-icon {
  font-size: 16px;
}

.menu-divider {
  height: 1px;
  background: #e5e7eb;
  margin: 4px 0;
}

.loading-indicator {
  text-align: center;
  padding: 60px 20px;
  color: #64748b;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.loading-spinner {
  width: 48px;
  height: 48px;
  border: 4px solid #e2e8f0;
  border-top-color: #326ce5;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.loading-indicator p {
  margin: 0;
  font-size: 14px;
  color: #64748b;
}

.empty-state {
  text-align: center;
  padding: 80px 20px;
  color: #94a3b8;
}

.empty-icon {
  font-size: 64px;
  margin-bottom: 16px;
  opacity: 0.6;
}

.empty-title {
  font-size: 18px;
  font-weight: 600;
  color: #64748b;
  margin: 0 0 8px 0;
}

.empty-desc {
  font-size: 14px;
  color: #94a3b8;
  margin: 0;
  max-width: 400px;
  margin-left: auto;
  margin-right: auto;
}

/* 卡片视图样式 */
.card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(380px, 1fr));
  gap: 20px;
  padding: 4px;
}

.resource-card {
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  overflow: hidden;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border: 2px solid transparent;
}

.resource-card:hover {
  transform: translateY(-6px);
  box-shadow: 0 12px 24px rgba(0, 0, 0, 0.15);
  border-color: #326ce5;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.card-header.status-complete {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
}

.card-header.status-running {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
}

.card-header.status-failed {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
}

.card-title {
  display: flex;
  align-items: center;
  gap: 10px;
}

.card-icon {
  font-size: 24px;
}

.card-title h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.card-body {
  padding: 20px;
}

.card-meta {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 16px;
}

.card-meta-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 13px;
}

.card-meta-item .label {
  color: #64748b;
  font-weight: 500;
}

.completion-bar-container {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 16px;
}

.completion-bar-container .completion-bar {
  flex: 1;
  height: 8px;
}

.completion-bar-container .completion-text {
  font-size: 12px;
  font-weight: 600;
  color: #10b981;
  min-width: 45px;
  text-align: right;
}

.card-footer {
  display: flex;
  gap: 8px;
  padding: 16px 20px;
  background: #f9fafb;
  border-top: 1px solid #e5e7eb;
}

.card-btn {
  flex: 1;
  padding: 8px 12px;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
  background: #f1f5f9;
  color: #475569;
}

.card-btn.primary {
  background: #326ce5;
  color: white;
}

.card-btn.danger {
  background: #fee2e2;
  color: #dc2626;
}

.card-btn:hover {
  transform: translateY(-2px);
  filter: brightness(1.1);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
}

.card-btn:active {
  transform: translateY(0);
}

/* 分页样式 */
.pagination-container {
  margin-top: 24px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  gap: 20px;
  flex-wrap: wrap;
}

.pagination-info {
  display: flex;
  align-items: center;
  gap: 20px;
  color: #64748b;
  font-size: 14px;
}

.total-info {
  font-weight: 500;
}

.items-per-page {
  display: flex;
  align-items: center;
  gap: 8px;
}

.items-per-page label {
  font-weight: 500;
  color: #64748b;
}

.items-per-page select {
  padding: 6px 12px;
  border: 2px solid #e2e8f0;
  border-radius: 8px;
  background: white;
  color: #475569;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 14px;
}

.items-per-page select:hover {
  border-color: #cbd5e1;
}

.items-per-page select:focus {
  outline: none;
  border-color: #326ce5;
  box-shadow: 0 0 0 3px rgba(50, 108, 229, 0.1);
}

.pagination-summary {
  color: #64748b;
  font-size: 14px;
  font-weight: 500;
}

/* 分页组件内部样式优化 */
.pagination-container :deep(.pagination) {
  display: flex;
  gap: 8px;
  align-items: center;
}

.pagination-container :deep(.page-btn) {
  padding: 8px 12px;
  border: 2px solid #e2e8f0;
  border-radius: 8px;
  background: white;
  color: #64748b;
  cursor: pointer;
  transition: all 0.2s;
  font-weight: 500;
}

.pagination-container :deep(.page-btn:hover) {
  border-color: #326ce5;
  color: #326ce5;
  background: #f0f7ff;
}

.pagination-container :deep(.page-btn.active) {
  background: #326ce5;
  color: white;
  border-color: #326ce5;
}

.pagination-container :deep(.page-btn:disabled) {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 模态框样式 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
  backdrop-filter: blur(4px);
}

.modal-content {
  background: white;
  border-radius: 16px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  max-width: 600px;
  width: 90%;
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.modal-content.large {
  max-width: 900px;
}

/* 高级模态框（Rancher/Kuboard 风格） */
.modal-content.advanced {
  max-width: 1000px;
  height: 85vh;
}

/* 步骤指示器 */
.steps-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px 32px;
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  border-bottom: 1px solid #e2e8f0;
}

.step-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  transition: all 0.3s;
  opacity: 0.5;
}

.step-item.active {
  opacity: 1;
}

.step-item.completed {
  opacity: 0.8;
}

.step-number {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #e2e8f0;
  color: #64748b;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 16px;
  transition: all 0.3s;
}

.step-item.active .step-number {
  background: linear-gradient(135deg, #326ce5 0%, #2557c7 100%);
  color: white;
  box-shadow: 0 4px 12px rgba(50, 108, 229, 0.4);
  transform: scale(1.1);
}

.step-item.completed .step-number {
  background: #10b981;
  color: white;
}

.step-label {
  font-size: 13px;
  font-weight: 600;
  color: #64748b;
  white-space: nowrap;
}

.step-item.active .step-label {
  color: #326ce5;
}

.step-item.completed .step-label {
  color: #10b981;
}

.step-line {
  height: 2px;
  width: 80px;
  background: #e2e8f0;
  margin: 0 16px;
  transition: all 0.3s;
}

.step-line.active {
  background: linear-gradient(90deg, #10b981 0%, #326ce5 100%);
}

/* 模态框主体 */
.modal-body.scrollable {
  flex: 1;
  overflow-y: auto;
  padding: 32px;
  background: #ffffff;
}

.modal-body.scrollable::-webkit-scrollbar {
  width: 8px;
}

.modal-body.scrollable::-webkit-scrollbar-track {
  background: #f1f5f9;
  border-radius: 4px;
}

.modal-body.scrollable::-webkit-scrollbar-thumb {
  background: #cbd5e1;
  border-radius: 4px;
}

.modal-body.scrollable::-webkit-scrollbar-thumb:hover {
  background: #94a3b8;
}

/* 步骤内容 */
.step-content {
  display: flex;
  flex-direction: column;
  gap: 28px;
}

/* 表单分区 */
.form-section {
  background: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 24px;
  transition: all 0.3s;
}

.form-section:hover {
  border-color: #cbd5e1;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 2px solid #e5e7eb;
}

.section-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 700;
  color: #1f2937;
  display: flex;
  align-items: center;
  gap: 8px;
}

.section-desc {
  margin: 8px 0 0 0;
  font-size: 13px;
  color: #6b7280;
  font-weight: 400;
}

/* 表单行 */
.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

@media (max-width: 768px) {
  .form-row {
    grid-template-columns: 1fr;
  }
}

/* 表单组 */
.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group label {
  font-size: 14px;
  font-weight: 600;
  color: #374151;
  display: flex;
  align-items: center;
  gap: 6px;
}

.form-group label.required::after {
  content: '*';
  color: #ef4444;
  font-size: 14px;
}

.form-group input,
.form-group select,
.form-group textarea {
  padding: 10px 14px;
  border: 2px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  color: #1f2937;
  transition: all 0.2s;
  background: white;
}

.form-group input:focus,
.form-group select:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #326ce5;
  box-shadow: 0 0 0 3px rgba(50, 108, 229, 0.1);
}

.form-group input:hover,
.form-group select:hover,
.form-group textarea:hover {
  border-color: #cbd5e1;
}

.form-group textarea {
  resize: vertical;
  min-height: 80px;
  font-family: inherit;
}

.form-hint {
  font-size: 12px;
  color: #6b7280;
  line-height: 1.4;
}

.form-hint.error-hint {
  color: #ef4444;
  font-weight: 500;
}

/* 命名空间选择器 */
.namespace-selector {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.namespace-selector select.form-select {
  padding: 10px 14px;
  border: 2px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  color: #1f2937;
  transition: all 0.2s;
  background: white;
}

.namespace-or {
  font-size: 14px;
  color: #9ca3af;
  font-weight: 500;
}

.namespace-create {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
}

.namespace-create .form-input {
  flex: 1;
  min-width: 200px;
  padding: 8px 12px;
  border: 2px solid #e2e8f0;
  border-radius: 6px;
  font-size: 14px;
}

.btn-sm {
  padding: 8px 16px;
  font-size: 13px;
  border-radius: 6px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.btn.btn-secondary.btn-sm {
  background: #f1f5f9;
  color: #64748b;
  border: 2px solid #e2e8f0;
}

.btn.btn-secondary.btn-sm:hover:not(:disabled) {
  background: #e2e8f0;
  border-color: #cbd5e1;
  transform: translateY(-1px);
}

.btn.btn-secondary.btn-sm:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.info-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: #dbeafe;
  color: #3b82f6;
  font-size: 11px;
  cursor: help;
  transition: all 0.2s;
}

.info-icon:hover {
  background: #3b82f6;
  color: white;
  transform: scale(1.1);
}

/* 环境变量/标签/选择器行 */
.env-var-row {
  display: grid;
  grid-template-columns: 1fr 1fr auto;
  gap: 12px;
  align-items: center;
  padding: 12px;
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  margin-bottom: 12px;
  transition: all 0.2s;
}

.env-var-row:hover {
  border-color: #cbd5e1;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.env-name,
.env-value {
  padding: 8px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-size: 13px;
  transition: all 0.2s;
}

.env-name:focus,
.env-value:focus {
  outline: none;
  border-color: #326ce5;
  box-shadow: 0 0 0 2px rgba(50, 108, 229, 0.1);
}

.add-btn {
  padding: 6px 14px;
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 2px 4px rgba(16, 185, 129, 0.2);
}

.add-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(16, 185, 129, 0.3);
}

.add-btn:active {
  transform: translateY(0);
}

.remove-btn {
  padding: 6px 10px;
  background: #fee2e2;
  color: #dc2626;
  border: 1px solid #fca5a5;
  border-radius: 6px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.remove-btn:hover {
  background: #dc2626;
  color: white;
  border-color: #dc2626;
  transform: scale(1.05);
}

.empty-hint {
  text-align: center;
  padding: 24px;
  color: #9ca3af;
  font-size: 13px;
  background: #f9fafb;
  border: 2px dashed #e5e7eb;
  border-radius: 8px;
  font-style: italic;
}

/* 模态框底部 */
.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 20px 24px;
  border-top: 1px solid #e5e7eb;
  background: #f9fafb;
}

.nav-btn {
  padding: 10px 24px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 8px;
}

.nav-btn.primary {
  background: linear-gradient(135deg, #326ce5 0%, #2557c7 100%);
  color: white;
  box-shadow: 0 2px 8px rgba(50, 108, 229, 0.3);
}

.nav-btn.primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(50, 108, 229, 0.4);
}

.nav-btn.secondary {
  background: #f1f5f9;
  color: #64748b;
  border: 2px solid #e2e8f0;
}

.nav-btn.secondary:hover {
  background: #e2e8f0;
  border-color: #cbd5e1;
}

.cancel-btn {
  padding: 10px 24px;
  background: white;
  border: 2px solid #e2e8f0;
  border-radius: 8px;
  color: #64748b;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.cancel-btn:hover {
  background: #f9fafb;
  border-color: #cbd5e1;
  color: #475569;
}

.submit-btn {
  padding: 10px 32px;
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  border: none;
  border-radius: 8px;
  color: white;
  font-size: 14px;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.3);
  display: flex;
  align-items: center;
  gap: 8px;
}

.submit-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.4);
}

.submit-btn:active:not(:disabled) {
  transform: translateY(0);
}

.submit-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  background: #94a3b8;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #e5e7eb;
}

.modal-header h2 {
  margin: 0;
  font-size: 20px;
  color: #1f2937;
}

.close-btn {
  background: none;
  border: none;
  font-size: 28px;
  color: #9ca3af;
  cursor: pointer;
  padding: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  transition: all 0.2s;
}

.close-btn:hover {
  background: #f3f4f6;
  color: #4b5563;
}

.modal-body {
  padding: 24px;
  overflow-y: auto;
  flex: 1;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-size: 14px;
  font-weight: 500;
  color: #374151;
}

.form-group input,
.form-group select {
  width: 100%;
  padding: 10px 14px;
  border: 2px solid #e5e7eb;
  border-radius: 8px;
  font-size: 14px;
  transition: all 0.2s;
}

.form-group input:focus,
.form-group select:focus {
  outline: none;
  border-color: #326ce5;
  box-shadow: 0 0 0 3px rgba(50, 108, 229, 0.1);
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 24px;
  border-top: 1px solid #e5e7eb;
  background: #f9fafb;
}

.cancel-btn {
  padding: 10px 20px;
  border: 2px solid #e5e7eb;
  border-radius: 8px;
  background: white;
  color: #6b7280;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s;
}

.cancel-btn:hover {
  background: #f3f4f6;
  border-color: #d1d5db;
}

.submit-btn {
  padding: 10px 20px;
  border: none;
  border-radius: 8px;
  background: #326ce5;
  color: white;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s;
}

.submit-btn:hover:not(:disabled) {
  background: #2557c7;
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(50, 108, 229, 0.3);
}

.submit-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 详情样式 */
.detail-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.detail-section {
  background: #f9fafb;
  padding: 16px;
  border-radius: 8px;
}

.detail-section h3 {
  margin: 0 0 16px 0;
  font-size: 16px;
  color: #374151;
  padding-bottom: 12px;
  border-bottom: 2px solid #e5e7eb;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px solid #e5e7eb;
}

.detail-row:last-child {
  border-bottom: none;
}

.detail-label {
  font-weight: 500;
  color: #6b7280;
  min-width: 140px;
}

/* 模态框滚动条样式 */
.modal-body::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

.modal-body::-webkit-scrollbar-track {
  background: #f8fafc;
  border-radius: 4px;
}

.modal-body::-webkit-scrollbar-thumb {
  background: linear-gradient(135deg, #cbd5e1 0%, #94a3b8 100%);
  border-radius: 4px;
  transition: all 0.3s;
}

.modal-body::-webkit-scrollbar-thumb:hover {
  background: linear-gradient(135deg, #94a3b8 0%, #64748b 100%);
}

/* 响应式设计 */
@media (max-width: 1200px) {
  .card-grid {
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  }
  
  .stats-bar {
    gap: 16px;
  }
}

@media (max-width: 768px) {
  .resource-view {
    padding: 16px;
  }
  
  .action-bar {
    flex-direction: column;
    align-items: stretch;
  }
  
  .search-box {
    min-width: 100%;
  }
  
  .filter-buttons {
    flex-wrap: wrap;
  }
  
  .action-buttons {
    flex-wrap: wrap;
    width: 100%;
  }
  
  .card-grid {
    grid-template-columns: 1fr;
  }
  
  .table-container {
    overflow-x: auto;
  }
  
  .stats-bar {
    flex-direction: column;
    gap: 12px;
  }
}

/* YAML 编辑器样式 */
.view-toggle-buttons {
  display: flex;
  gap: 8px;
  margin: 0 auto;
}

.view-toggle-btn {
  padding: 8px 16px;
  border: 2px solid #e2e8f0;
  background: white;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  color: #64748b;
}

.view-toggle-btn:hover {
  border-color: #326ce5;
  color: #326ce5;
}

.view-toggle-btn.active {
  background: #326ce5;
  border-color: #326ce5;
  color: white;
}

.yaml-editor-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding: 20px;
}

.yaml-editor-header {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.yaml-editor-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 700;
  color: #2c3e50;
}

.yaml-hint {
  color: #7f8c8d;
  font-size: 14px;
  margin: 0;
}

.yaml-header-buttons {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.load-template-btn {
  padding: 10px 20px;
  background: linear-gradient(135deg, #1e293b 0%, #0f172a 100%);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 2px 8px rgba(15, 23, 42, 0.3);
}

.load-template-btn:hover {
  background: linear-gradient(135deg, #0f172a 0%, #020617 100%);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(15, 23, 42, 0.5);
}

.load-template-btn:active {
  background: #020617;
  transform: translateY(0);
}

.clear-yaml-btn,
.copy-yaml-btn,
.reset-yaml-btn {
  padding: 10px 20px;
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.3s ease;
}

.copy-yaml-btn {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.3);
}

.copy-yaml-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.4);
}

.clear-yaml-btn,
.reset-yaml-btn {
  background: linear-gradient(135deg, #94a3b8 0%, #64748b 100%);
  box-shadow: 0 2px 8px rgba(100, 116, 139, 0.3);
}

.clear-yaml-btn:hover,
.reset-yaml-btn:hover {
  background: linear-gradient(135deg, #64748b 0%, #475569 100%);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(100, 116, 139, 0.4);
}

.yaml-editor {
  width: 100%;
  min-height: 400px;
  max-height: 500px;
  padding: 16px;
  border: 2px solid #334155;
  border-radius: 12px;
  font-family: 'Monaco', 'Menlo', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
  resize: vertical;
  background: #1e1e1e;
  color: #d4d4d4;
  transition: all 0.2s;
}

.yaml-editor:focus {
  outline: none;
  border-color: #326ce5;
  background: #1e1e1e;
  box-shadow: 0 0 0 3px rgba(50, 108, 229, 0.3);
}

.yaml-editor::placeholder {
  color: #94a3b8;
}

.yaml-error {
  padding: 12px 16px;
  background: #fee2e2;
  border: 1px solid #fca5a5;
  border-radius: 8px;
  color: #dc2626;
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.error-icon {
  font-size: 18px;
}

.yaml-editor-footer {
  padding: 16px;
  background: #f8f9fa;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
}

.yaml-tips {
  color: #64748b;
  font-size: 13px;
}

.yaml-tips strong {
  color: #2c3e50;
  font-weight: 600;
}

.yaml-tips ul {
  margin: 8px 0 0 0;
  padding-left: 20px;
}

.yaml-tips li {
  margin: 4px 0;
  line-height: 1.6;
}

/* 批量操作相关样式 */
.btn-batch {
  background-color: #8b5cf6;
  color: white;
}

.btn-batch:hover {
  background-color: #7c3aed;
}

.batch-action-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 12px 20px;
  border-radius: 8px;
  margin-bottom: 16px;
}

.batch-info {
  display: flex;
  align-items: center;
  gap: 16px;
  color: white;
}

.batch-count {
  font-weight: 600;
}

.batch-clear {
  background: rgba(255,255,255,0.2);
  border: none;
  color: white;
  padding: 4px 12px;
  border-radius: 4px;
  cursor: pointer;
}

.batch-actions {
  display: flex;
  gap: 8px;
}

.batch-btn {
  background: rgba(255,255,255,0.2);
  border: none;
  color: white;
  padding: 8px 16px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.batch-btn:hover {
  background: rgba(255,255,255,0.3);
}

.batch-btn.danger {
  background: rgba(239, 68, 68, 0.8);
}

.row-selected {
  background-color: #ebf5ff !important;
}

.card-selected {
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.2);
}

.card-checkbox {
  position: absolute;
  top: 12px;
  left: 12px;
  z-index: 10;
}

.card-checkbox input[type="checkbox"] {
  width: 18px;
  height: 18px;
  cursor: pointer;
}

/* 批量删除弹窗 */
.modal-batch-preview {
  max-width: 700px;
}

.modal-danger .modal-header {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: white;
}

.danger-header h3 {
  color: white;
}

.danger-warning {
  display: flex;
  gap: 16px;
  padding: 20px;
  background: #fef2f2;
  border: 2px solid #fecaca;
  border-radius: 12px;
  margin-bottom: 24px;
}

.warning-icon-large {
  font-size: 48px;
  flex-shrink: 0;
}

.warning-content {
  flex: 1;
}

.warning-title {
  font-size: 16px;
  font-weight: 600;
  color: #dc2626;
  margin-bottom: 12px;
}

.warning-list {
  margin: 0;
  padding-left: 20px;
  color: #991b1b;
}

.warning-list li {
  margin: 6px 0;
}

.preview-section {
  margin-bottom: 24px;
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: #374151;
  margin-bottom: 12px;
}

.affected-jobs-detail {
  display: grid;
  gap: 12px;
  max-height: 300px;
  overflow-y: auto;
}

.affected-job-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
}

.job-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.job-name {
  font-weight: 600;
  color: #1f2937;
}

.job-namespace {
  font-size: 12px;
  color: #6b7280;
  background: #e5e7eb;
  padding: 2px 8px;
  border-radius: 4px;
  width: fit-content;
}

.job-stats {
  display: flex;
  gap: 8px;
  align-items: center;
}

.completion-tag {
  font-size: 12px;
  color: #6b7280;
  background: #f3f4f6;
  padding: 4px 8px;
  border-radius: 4px;
}

.confirm-section {
  margin-top: 24px;
}

.confirm-input {
  width: 100%;
  padding: 12px;
  border: 2px solid #d1d5db;
  border-radius: 8px;
  font-size: 14px;
  margin-top: 8px;
  transition: all 0.2s;
}

.confirm-input:focus {
  outline: none;
  border-color: #3b82f6;
}

.confirm-input.valid {
  border-color: #10b981;
  background: #f0fdf4;
}

/* YAML 查看/编辑模态框样式 */
.yaml-modal {
  max-width: 900px;
  width: 90%;
  max-height: 90vh;
}

.yaml-header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.yaml-modal-body {
  padding: 0;
  max-height: 70vh;
  overflow-y: auto;
}

.yaml-editor-wrapper {
  width: 100%;
  height: 100%;
}

.yaml-editor,
.yaml-content {
  width: 100%;
  min-height: 500px;
  padding: 20px;
  font-family: 'Monaco', 'Menlo', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
  background: #1e1e1e;
  color: #d4d4d4;
  border: none;
  resize: vertical;
}

.yaml-content {
  margin: 0;
  overflow-x: auto;
  white-space: pre;
}

.yaml-editor:focus {
  outline: none;
  box-shadow: inset 0 0 0 2px #326ce5;
}

.loading-state {
  text-align: center;
  padding: 40px;
  color: #718096;
  font-size: 14px;
}

.error-box {
  padding: 16px 20px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  color: #dc2626;
  border-radius: 8px;
  margin: 20px;
  font-size: 14px;
}

/* =========================
   内联编辑样式
   ========================= */
/* 镜像显示与编辑 */
.image-text {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  border-radius: 6px;
  transition: all 0.2s;
  font-size: 13px;
  max-width: 300px;
}

.image-text.clickable {
  cursor: pointer;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
}

.image-text.clickable:hover {
  background: #eff6ff;
  border-color: #3b82f6;
  transform: scale(1.02);
}

.image-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #1e40af;
  font-weight: 500;
}

.image-hint {
  font-size: 12px;
  opacity: 0.6;
  cursor: help;
}

.edit-icon {
  font-size: 14px;
  opacity: 0.6;
  transition: opacity 0.2s;
}

.image-text.clickable:hover .edit-icon {
  opacity: 1;
}

/* 内联编辑输入框 */
.inline-edit-wrapper {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px;
  background: #fffbeb;
  border: 2px solid #fbbf24;
  border-radius: 6px;
  animation: fadeIn 0.2s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: scale(0.95);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

.inline-input {
  flex: 1;
  min-width: 250px;
  padding: 6px 10px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 13px;
  font-family: 'Monaco', 'Menlo', 'Courier New', monospace;
  transition: all 0.2s;
  background: white;
}

.inline-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.inline-hint {
  font-size: 11px;
  color: #92400e;
  font-weight: 500;
  white-space: nowrap;
  background: #fef3c7;
  padding: 2px 8px;
  border-radius: 4px;
}

/* ========== Job 日志样式 ========== */
.logs-modal {
  max-width: 900px;
  width: 95%;
  max-height: 85vh;
}

.logs-controls {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  align-items: flex-end;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 8px;
  margin-bottom: 16px;
}

.logs-controls .control-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.logs-controls .control-item label {
  font-size: 12px;
  color: #6b7280;
  font-weight: 500;
}

.logs-controls .form-select {
  padding: 6px 10px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 13px;
  min-width: 120px;
}

.logs-controls .control-actions {
  display: flex;
  gap: 8px;
  margin-left: auto;
}

.follow-toggle {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
}

.follow-toggle input {
  cursor: pointer;
}

.streaming-indicator {
  color: #10b981;
  font-size: 10px;
  animation: pulse 1s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.logs-viewer {
  background: #1e1e1e;
  border-radius: 8px;
  overflow: hidden;
}

.logs-content {
  max-height: 450px;
  overflow-y: auto;
  padding: 16px;
  font-family: 'Monaco', 'Menlo', 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.6;
  color: #d4d4d4;
  background: #1e1e1e;
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
}

.logs-content::-webkit-scrollbar {
  width: 8px;
}

.logs-content::-webkit-scrollbar-track {
  background: #2d2d2d;
}

.logs-content::-webkit-scrollbar-thumb {
  background: #555;
  border-radius: 4px;
}

.logs-content::-webkit-scrollbar-thumb:hover {
  background: #666;
}

/* 日志高亮样式 */
.log-timestamp {
  color: #888;
}

.log-error {
  color: #f87171;
  background: rgba(248, 113, 113, 0.1);
}

.log-warn {
  color: #fbbf24;
}

.log-info {
  color: #60a5fa;
}

.log-debug {
  color: #9ca3af;
}

.log-placeholder {
  color: #6b7280;
  font-style: italic;
}

.info-box {
  background: #f0f9ff;
  border: 1px solid #bae6fd;
  border-radius: 6px;
  padding: 12px;
  font-size: 13px;
}

.info-box div {
  margin-bottom: 4px;
}

.info-box div:last-child {
  margin-bottom: 0;
}
</style>
