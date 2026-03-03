<template>
  <div class="resource-view">
    <div class="view-header">
      <h1>CronJob管理</h1>
      <p>Kubernetes集群中的CronJob列表</p>
    </div>

    <!-- 操作栏 -->
    <div class="action-bar">
      <div class="search-box">
        <input type="text" v-model="searchQuery" placeholder="搜索CronJob名称..." @input="onSearchInput" />
      </div>

      <div class="filter-buttons">
        <button class="btn btn-filter" :class="{ active: statusFilter === 'all' }" @click="setStatusFilter('all')">
          全部
        </button>
        <button class="btn btn-filter" :class="{ active: statusFilter === 'Active' }" @click="setStatusFilter('Active')">
          Active
        </button>
        <button class="btn btn-filter" :class="{ active: statusFilter === 'Suspended' }" @click="setStatusFilter('Suspended')">
          Suspended
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
        <button class="btn btn-primary" @click="showCreateModal = true">创建CronJob</button>
        <button class="btn btn-secondary" @click="refreshList" :disabled="loading">
          {{ loading ? '加载中...' : '🔄 刷新' }}
        </button>
      </div>
    </div>

    <div v-if="errorMsg" class="error-box">{{ errorMsg }}</div>

    <!-- 批量操作浮动栏 -->
    <div v-if="batchMode && selectedCronjobs.length > 0" class="batch-action-bar">
      <div class="batch-info">
        <span class="batch-count">已选择 {{ selectedCronjobs.length }} 个 CronJob</span>
        <button class="batch-clear" @click="clearSelection">清空选择</button>
      </div>
      <div class="batch-actions">
        <button class="batch-btn warning" @click="batchSuspend(true)" title="批量暂停">
          ⏸️ 批量暂停
        </button>
        <button class="batch-btn success" @click="batchSuspend(false)" title="批量恢复">
          ▶️ 批量恢复
        </button>
        <button class="batch-btn danger" @click="openBatchDeletePreview" title="批量删除">
          🗑️ 批量删除
        </button>
      </div>
    </div>

    <!-- 数据统计 -->
    <div v-if="!loading && cronjobs.length > 0" class="stats-bar">
      <div class="stat-item">
        <span class="stat-label">总计:</span>
        <span class="stat-value">{{ total }}</span>
      </div>
      <div class="stat-item">
        <span class="stat-label">✅ Active:</span>
        <span class="stat-value success">{{ getStatusCount('Active') }}</span>
      </div>
      <div class="stat-item">
        <span class="stat-label">⏸️ Suspended:</span>
        <span class="stat-value warning">{{ getStatusCount('Suspended') }}</span>
      </div>
      <div class="stat-item">
        <span class="stat-label">当前页:</span>
        <span class="stat-value">{{ paginatedCronjobs.length }}</span>
      </div>
    </div>

    <!-- CronJob 表格 -->
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
              />
            </th>
            <th style="width: 100px;">状态</th>
            <th style="min-width: 180px;">名称</th>
            <th style="width: 130px;">命名空间</th>
            <th style="min-width: 150px;">调度表达式</th>
            <th style="width: 80px;">挂起</th>
            <th style="width: 170px;">最后调度</th>
            <th style="width: 140px;">执行统计</th>
            <th style="min-width: 200px;">镜像</th>
            <th style="width: 170px;">创建时间</th>
            <th style="width: 220px;">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading">
            <td colspan="10" style="text-align: center; padding: 40px;">
              <div class="loading-spinner"></div>
              <div style="margin-top: 10px;">加载中...</div>
            </td>
          </tr>
          <tr v-else-if="paginatedCronjobs.length === 0">
            <td colspan="10" style="text-align: center; padding: 40px; color: #999;">
              {{ searchQuery || namespaceFilter || statusFilter !== 'all' ? '没有匹配的CronJob' : '暂无CronJob' }}
            </td>
          </tr>
          <tr v-for="cj in paginatedCronjobs" :key="cj.name + cj.namespace" :class="{ 'row-selected': isCronjobSelected(cj) }">
            <td v-if="batchMode">
              <input
                type="checkbox"
                :checked="isCronjobSelected(cj)"
                @change="toggleCronjobSelection(cj)"
              />
            </td>
            <td>
              <div class="status-cell">
                <span class="status-indicator" :class="cj.status.toLowerCase()">
                  <span class="status-dot"></span>
                  {{ cj.status }}
                </span>
              </div>
            </td>
            <td>
              <div class="job-name">
                <span class="icon">⏰</span>
                <span>{{ cj.name }}</span>
                <span class="age-badge" :title="cj.created_at">{{ getAge(cj.created_at) }}</span>
              </div>
            </td>
            <td>
              <span class="namespace-badge">{{ cj.namespace }}</span>
            </td>
            <td>
              <code class="schedule-text">{{ cj.schedule }}</code>
            </td>
            <td>
              <span :class="['suspend-badge', cj.suspend ? 'suspended' : 'active']">
                {{ cj.suspend ? '是' : '否' }}
              </span>
            </td>
            <td>
              <span class="time-text">{{ cj.last_schedule_time || '-' }}</span>
            </td>
            <td>
              <div class="job-stats">
                <span class="stat-item success" :title="`成功: ${cj.job_stats?.succeeded || 0}`">✔ {{ cj.job_stats?.succeeded || 0 }}</span>
                <span class="stat-item danger" :title="`失败: ${cj.job_stats?.failed || 0}`">✘ {{ cj.job_stats?.failed || 0 }}</span>
                <span class="stat-item running" :title="`运行中: ${cj.job_stats?.running || 0}`">▶ {{ cj.job_stats?.running || 0 }}</span>
              </div>
            </td>
            <td>
              <div class="image-text" :title="cj.image || '-'">
                <span class="image-name">{{ cj.image || '-' }}</span>
              </div>
            </td>
            <td>
              <span class="time-text">{{ cj.created_at }}</span>
            </td>
            <td class="action-cell">
              <div class="action-icons">
                <!-- 日志按钮（直接显示） -->
                <button class="action-btn primary" @click="viewJobs(cj)" title="查看此 CronJob 创建的所有 Job">
                  📦 Job 列表
                </button>
                
                <!-- 更多按钮 -->
                <div class="more-btn">
                  <button class="icon-btn" @click="toggleMoreOptions(cj, $event)">
                    ⋮ 更多
                  </button>

                  <!-- 更多菜单 -->
                  <div v-if="showMoreOptions && selectedCronjob === cj" class="more-menu" :style="menuStyle">
                    <button class="menu-item" @click="viewDetail(cj)">
                      <span class="menu-icon">📋</span>
                      <span>查看详情</span>
                    </button>
                    <button class="menu-item" @click="viewYaml(cj)">
                      <span class="menu-icon">📝</span>
                      <span>查看/编辑 YAML</span>
                    </button>
                    <div class="menu-divider"></div>
                    <button v-if="!cj.suspend" class="menu-item" @click="suspendCronJob(cj, true)">
                      <span class="menu-icon">⏸️</span>
                      <span>暂停</span>
                    </button>
                    <button v-else class="menu-item" @click="suspendCronJob(cj, false)">
                      <span class="menu-icon">▶️</span>
                      <span>恢复</span>
                    </button>
                    <button class="menu-item" @click="triggerCronJob(cj)">
                      <span class="menu-icon">🚀</span>
                      <span>手动触发</span>
                    </button>
                    <div class="menu-divider"></div>
                    <button class="menu-item danger" @click="confirmDelete(cj)">
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
    </div>

    <!-- 卡片视图 -->
    <div v-else class="card-grid">
      <div v-for="cj in paginatedCronjobs" :key="cj.name + cj.namespace" class="resource-card" :class="{ 'card-selected': isCronjobSelected(cj) }">
        <!-- 批量选择复选框 -->
        <div v-if="batchMode" class="card-checkbox">
          <input
            type="checkbox"
            :checked="isCronjobSelected(cj)"
            @change="toggleCronjobSelection(cj)"
          />
        </div>

        <div class="card-header" :class="`status-${cj.status.toLowerCase()}`">
          <div class="card-title">
            <span class="card-icon">⏰</span>
            <h3>{{ cj.name }}</h3>
          </div>
          <span class="status-indicator" :class="cj.status.toLowerCase()">
            {{ cj.status }}
          </span>
        </div>

        <div class="card-body">
          <div class="card-meta">
            <div class="card-meta-item">
              <span class="label">命名空间:</span>
              <span class="namespace-badge">{{ cj.namespace }}</span>
            </div>
            <div class="card-meta-item">
              <span class="label">调度:</span>
              <code class="schedule-text">{{ cj.schedule }}</code>
            </div>
            <div class="card-meta-item">
              <span class="label">暂停:</span>
              <span :class="['suspend-badge', cj.suspend ? 'suspended' : 'active']">
                {{ cj.suspend ? '是' : '否' }}
              </span>
            </div>
            <div class="card-meta-item">
              <span class="label">执行统计:</span>
              <div class="job-stats">
                <span class="stat-item success">✔ {{ cj.job_stats?.succeeded || 0 }}</span>
                <span class="stat-item danger">✘ {{ cj.job_stats?.failed || 0 }}</span>
                <span class="stat-item running">▶ {{ cj.job_stats?.running || 0 }}</span>
              </div>
            </div>
            <div class="card-meta-item">
              <span class="label">镜像:</span>
              <span class="image-name">{{ cj.image || '-' }}</span>
            </div>
            <div class="card-meta-item">
              <span class="label">最后调度:</span>
              <span class="time-text">{{ cj.last_schedule_time || '-' }}</span>
            </div>
            <div class="card-meta-item">
              <span class="label">创建时间:</span>
              <span class="time-text">{{ cj.created_at }}</span>
            </div>
          </div>
        </div>

        <div class="card-footer">
          <button
            v-if="!cj.suspend"
            @click="suspendCronJob(cj, true)"
            class="card-btn warning"
          >
            ⏸️ 暂停
          </button>
          <button
            v-else
            @click="suspendCronJob(cj, false)"
            class="card-btn success"
          >
            ▶️ 恢复
          </button>
          <button @click="viewDetail(cj)" class="card-btn primary">📋 详情</button>
          <button @click="viewYaml(cj)" class="card-btn">📝 YAML</button>
          <button @click="confirmDelete(cj)" class="card-btn danger">🗑️ 删除</button>
        </div>
      </div>

      <div v-if="loading" class="loading-indicator">
        <div class="loading-spinner"></div>
        <p>加载中...</p>
      </div>
      <div v-if="!loading && filteredCronjobs.length === 0" class="empty-state">
        <div class="empty-icon">📁</div>
        <p class="empty-title">暂无CronJob数据</p>
        <p class="empty-desc">当前筛选条件下没有找到CronJob，试试调整筛选条件或创建新的CronJob</p>
        <button class="btn btn-primary" @click="showCreateModal = true" style="margin-top: 16px;">
          创建第一个CronJob
        </button>
      </div>
    </div>

    <!-- 分页 -->
    <div v-if="!loading && cronjobs.length > 0" class="pagination-container">
      <Pagination
        v-model:currentPage="currentPage"
        :totalItems="filteredCronjobs.length"
        :itemsPerPage="itemsPerPage"
      />
    </div>

    <!-- 创建 CronJob 模态框 - Rancher/Kuboard 风格 -->
    <div v-if="showCreateModal" class="modal-overlay" @click="closeCreateModal">
      <div class="modal-content advanced" @click.stop>
        <div class="modal-header">
          <h2>⏰ 创建 CronJob</h2>
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
          <button @click="closeCreateModal" class="close-btn">&times;</button>
        </div>
        
        <div class="modal-body scrollable">
          <!-- 表单模式 -->
          <div v-if="createMode === 'form'">
          <div class="form-group">
            <label>名称 <span class="required">*</span></label>
            <input type="text" v-model="cronjobForm.name" class="form-input" placeholder="例如: backup-db" required>
          </div>

          <div class="form-group">
            <label>命名空间 <span class="required">*</span></label>
            <div class="namespace-selector">
              <select v-if="!showNamespaceInput" v-model="cronjobForm.namespace" class="form-select">
                <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
              </select>
              <span v-if="!showNamespaceInput" class="namespace-or">或</span>
              <button v-if="!showNamespaceInput" @click="showNamespaceInput = true" class="btn-create-ns">
                ➕ 创建新命名空间
              </button>
              <div v-if="showNamespaceInput" class="new-namespace-input">
                <input
                  type="text"
                  v-model="newNamespace"
                  class="form-input"
                  placeholder="输入新命名空间名称"
                  @keyup.enter="createNewNamespace"
                />
                <button @click="createNewNamespace" class="btn-confirm" :disabled="creatingNamespace">
                  {{ creatingNamespace ? '创建中...' : '✓' }}
                </button>
                <button @click="cancelCreateNamespace" class="btn-cancel">✗</button>
              </div>
            </div>
          </div>

          <div class="form-group">
            <label>调度表达式 (Cron) <span class="required">*</span></label>
            <input type="text" v-model="cronjobForm.schedule" class="form-input" placeholder="例如: 0 * * * * (每小时)" required>
            <small class="help-text">格式: 分 时 日 月 周 (例如: "0 2 * * *" 表示每天凌晨2点)</small>
          </div>

          <div class="form-group">
            <label>容器镜像 <span class="required">*</span></label>
            <input type="text" v-model="cronjobForm.container_image" class="form-input" placeholder="例如: busybox:latest" required>
          </div>

          <!-- 容器命令配置 (Rancher/Kuboard 风格) -->
          <div class="form-section-header">
            <h4>⚙️ 容器命令配置</h4>
            <p class="section-desc">配置容器启动时执行的命令和参数</p>
          </div>

          <div class="form-group">
            <label>容器名称</label>
            <input type="text" v-model="cronjobForm.container_name" class="form-input" placeholder="留空则使用 CronJob 名称">
            <small class="help-text">留空则默认使用 CronJob 名称作为容器名</small>
          </div>

          <div class="form-group">
            <label>命令 (Command) <span class="info-icon" title="覆盖容器的默认 ENTRYPOINT">ℹ️</span></label>
            <div class="command-input-wrapper">
              <textarea 
                v-model="cronjobForm.container_command" 
                class="form-textarea" 
                placeholder="每行一个命令参数，例如：&#10;/bin/sh&#10;-c"
                rows="3"
              ></textarea>
              <small class="help-text">每行一个参数，用于覆盖镜像的默认 ENTRYPOINT</small>
            </div>
          </div>

          <div class="form-group">
            <label>参数 (Args) <span class="info-icon" title="传递给命令的参数">ℹ️</span></label>
            <div class="command-input-wrapper">
              <textarea 
                v-model="cronjobForm.container_args" 
                class="form-textarea" 
                placeholder="每行一个参数，例如：&#10;echo Hello World && sleep 30"
                rows="3"
              ></textarea>
              <small class="help-text">每行一个参数，传递给命令或镜像的默认 CMD</small>
            </div>
          </div>

          <!-- 快速命令模板 -->
          <div class="command-templates">
            <span class="template-label">快速模板:</span>
            <button type="button" class="template-btn" @click="applyCommandTemplate('shell')">🐚 Shell 脚本</button>
            <button type="button" class="template-btn" @click="applyCommandTemplate('echo')">📢 Echo 测试</button>
            <button type="button" class="template-btn" @click="applyCommandTemplate('curl')">🌐 Curl 请求</button>
            <button type="button" class="template-btn" @click="applyCommandTemplate('backup')">💾 数据备份</button>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label>成功历史限制</label>
              <input type="number" v-model.number="cronjobForm.successful_jobs_history_limit" class="form-input" min="0">
            </div>
            <div class="form-group">
              <label>失败历史限制</label>
              <input type="number" v-model.number="cronjobForm.failed_jobs_history_limit" class="form-input" min="0">
            </div>
          </div>

          <div class="form-group">
            <label class="checkbox-label">
              <input type="checkbox" v-model="cronjobForm.suspend" class="form-checkbox">
              <span>创建后立即挂起</span>
            </label>
          </div>
          </div>

          <!-- YAML 模式 -->
          <div v-if="createMode === 'yaml'" class="yaml-editor-container">
            <div class="yaml-editor-header">
              <h3>📄 YAML 配置</h3>
              <p class="yaml-hint">✨ 支持直接粘贴 CronJob YAML 配置文件</p>
              <button type="button" class="load-template-btn" @click="loadCronjobYamlTemplate">
                📑 加载模板
              </button>
              <button type="button" class="copy-yaml-btn" @click="copyYamlContent">
                📋 复制
              </button>
              <button type="button" class="reset-yaml-btn" @click="resetYamlContent">
                🔄 重置
              </button>
            </div>
            
            <textarea 
              v-model="createYamlContent" 
              class="yaml-editor"
              placeholder="输入或粘贴 CronJob YAML 内容...&#10;&#10;apiVersion: batch/v1&#10;kind: CronJob&#10;metadata:&#10;  name: my-cronjob&#10;  namespace: default&#10;spec:&#10;  schedule: '*/5 * * * *'&#10;  jobTemplate:&#10;    spec:&#10;      template:&#10;        spec:&#10;          containers:&#10;          - name: hello&#10;            image: busybox:latest&#10;            command: ['/bin/sh', '-c', 'echo Hello']&#10;          restartPolicy: OnFailure"
              spellcheck="false"
            ></textarea>
            
            <div v-if="createYamlError" class="yaml-error">
              <span class="error-icon">⚠️</span>
              {{ createYamlError }}
            </div>
            
            <div class="yaml-editor-footer">
              <div class="yaml-tips">
                <strong>💡 提示：</strong>
                <ul>
                  <li>✅ 支持标准 Kubernetes CronJob YAML 格式</li>
                  <li>📅 schedule 字段使用 Cron 表达式（分 时 日 月 周）</li>
                  <li>🚀 点击“加载模板”获取完整的示例配置</li>
                </ul>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <!-- 表单模式按钮 -->
          <template v-if="createMode === 'form'">
            <button @click="closeCreateModal" class="btn btn-secondary">取消</button>
            <button @click="createCronjob" class="btn btn-primary" :disabled="creating">
              <span class="btn-icon">🚀</span>
              {{ creating ? '创建中...' : '创建 CronJob' }}
            </button>
          </template>
          
          <!-- YAML 模式按钮 -->
          <template v-if="createMode === 'yaml'">
            <button @click="closeCreateModal" class="btn btn-secondary">取消</button>
            <button 
              @click="createCronjobFromYaml" 
              class="btn btn-primary" 
              :disabled="!createYamlContent || creatingFromYaml"
            >
              <span class="btn-icon">🚀</span>
              {{ creatingFromYaml ? '创建中...' : '创建 CronJob' }}
            </button>
          </template>
        </div>
      </div>
    </div>

    <!-- 详情模态框 -->
    <div v-if="showDetailModal" class="modal-overlay" @click="showDetailModal = false">
      <div class="modal-content modal-large" @click.stop>
        <div class="modal-header">
          <h3>CronJob 详情: {{ selectedCronjob?.name }}</h3>
          <button @click="showDetailModal = false" class="close-btn">×</button>
        </div>
        <div class="modal-body">
          <div v-if="loadingDetail" style="text-align: center; padding: 40px;">
            <div class="loading-spinner"></div>
            <div style="margin-top: 10px;">加载详情中...</div>
          </div>
          <div v-else-if="detailData">
            <div class="detail-section">
              <h4>基本信息</h4>
              <table class="detail-table">
                <tbody>
                  <tr>
                    <td class="label">名称:</td>
                    <td>{{ detailData.cronjob.name }}</td>
                  </tr>
                  <tr>
                    <td class="label">命名空间:</td>
                    <td>{{ detailData.cronjob.namespace }}</td>
                  </tr>
                  <tr>
                    <td class="label">调度表达式:</td>
                    <td><code>{{ detailData.cronjob.schedule }}</code></td>
                  </tr>
                  <tr>
                    <td class="label">挂起状态:</td>
                    <td>{{ detailData.cronjob.suspend ? '是' : '否' }}</td>
                  </tr>
                  <tr>
                    <td class="label">并发策略:</td>
                    <td>{{ detailData.cronjob.concurrencyPolicy }}</td>
                  </tr>
                  <tr>
                    <td class="label">最后调度时间:</td>
                    <td>{{ detailData.cronjob.lastScheduleTime || '-' }}</td>
                  </tr>
                  <tr>
                    <td class="label">最后成功时间:</td>
                    <td>{{ detailData.cronjob.lastSuccessfulTime || '-' }}</td>
                  </tr>
                </tbody>
              </table>
            </div>

            <div class="detail-section" v-if="detailData.jobs && detailData.jobs.length > 0">
              <h4>历史任务 ({{ detailData.jobs.length }})</h4>
              <table class="detail-table">
                <thead>
                  <tr>
                    <th>任务名称</th>
                    <th>状态</th>
                    <th>开始时间</th>
                    <th>完成时间</th>
                    <th>活跃/成功/失败</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="job in detailData.jobs" :key="job.name">
                    <td>{{ job.name }}</td>
                    <td>
                      <span class="status-badge" :class="job.phase.toLowerCase()">{{ job.phase }}</span>
                    </td>
                    <td>{{ formatTime(job.startTime) }}</td>
                    <td>{{ formatTime(job.completionTime) }}</td>
                    <td>{{ job.active || 0 }} / {{ job.succeeded || 0 }} / {{ job.failed || 0 }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div v-else class="detail-section">
              <p style="color: #999; text-align: center; padding: 20px;">暂无历史任务</p>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button @click="showDetailModal = false" class="btn btn-secondary">关闭</button>
        </div>
      </div>
    </div>

    <!-- 删除确认模态框 -->
    <div v-if="showDeleteModal" class="modal-overlay" @click="showDeleteModal = false">
      <div class="modal-content delete-confirm" @click.stop>
        <div class="modal-header">
          <h3>⚠️ 确认删除</h3>
          <button @click="showDeleteModal = false" class="close-btn">×</button>
        </div>
        <div class="modal-body">
          <p>确定要删除 CronJob <strong>{{ selectedCronjob?.name }}</strong> 吗？</p>
          <p class="warning-text">删除后将停止所有定时任务，此操作不可恢复。</p>
        </div>
        <div class="modal-footer">
          <button @click="showDeleteModal = false" class="btn btn-secondary">取消</button>
          <button @click="deleteCronjob" class="btn btn-danger" :disabled="deleting">
            {{ deleting ? '删除中...' : '确认删除' }}
          </button>
        </div>
      </div>
    </div>

    <!-- YAML 查看/编辑模态框 -->
    <div v-if="showYamlModal" class="modal-overlay" @click="closeYamlModal">
      <div class="modal-content yaml-modal" @click.stop>
        <div class="modal-header">
          <h3>📝 {{ yamlEditMode ? '编辑 YAML' : '查看 YAML' }} - {{ selectedYamlCronjob?.name }}</h3>
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

    <!-- 批量删除模态框 -->
    <div v-if="showBatchDeleteModal" class="modal-overlay" @click="closeBatchDeleteModal">
      <div class="modal-content batch-delete-modal" @click.stop>
        <div class="modal-header">
          <h3>⚠️ 确认批量删除</h3>
          <button @click="closeBatchDeleteModal" class="close-btn">×</button>
        </div>
        <div class="modal-body">
          <p class="warning-text" style="margin-bottom: 16px;">
            您即将删除以下 {{ selectedCronjobs.length }} 个 CronJob，此操作<strong>不可恢复</strong>！
          </p>
          <div class="batch-delete-list">
            <div v-for="cj in selectedCronjobs" :key="cj.name + cj.namespace" class="batch-delete-item">
              🗑️ {{ cj.namespace }}/{{ cj.name }}
            </div>
          </div>
          <div class="confirm-input-group">
            <label>请输入 <code>DELETE</code> 确认删除：</label>
            <input
              type="text"
              v-model="deleteConfirmText"
              class="form-input"
              placeholder="DELETE"
              @keyup.enter="executeBatchDelete"
            />
          </div>
        </div>
        <div class="modal-footer">
          <button @click="closeBatchDeleteModal" class="btn btn-secondary">取消</button>
          <button
            @click="executeBatchDelete"
            class="btn btn-danger"
            :disabled="deleteConfirmText !== 'DELETE' || batchExecuting"
          >
            {{ batchExecuting ? '删除中...' : `确认删除 ${selectedCronjobs.length} 个` }}
          </button>
        </div>
      </div>
    </div>

    <!-- Job 列表模态框 -->
    <div v-if="showJobsModal" class="modal-overlay" @click="closeJobsModal">
      <div class="modal-content jobs-list-modal" @click.stop>
        <div class="modal-header">
          <h3>📆 CronJob 关联的 Job 列表 - {{ selectedCronjobForJobs?.name }}</h3>
          <button @click="closeJobsModal" class="close-btn">×</button>
        </div>
        <div class="modal-body">
          <div v-if="loadingJobs" class="loading-state">加载中...</div>
          <div v-else-if="cronjobJobs.length === 0" class="empty-state">
            <p style="color: #999; text-align: center; padding: 40px;">暂无关联的 Job</p>
          </div>
          <div v-else>
            <table class="detail-table">
              <thead>
                <tr>
                  <th>Job 名称</th>
                  <th>状态</th>
                  <th>开始时间</th>
                  <th>完成时间</th>
                  <th>活跃/成功/失败</th>
                  <th>耗时</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="job in cronjobJobs" :key="job.name">
                  <td>{{ job.name }}</td>
                  <td>
                    <span class="status-badge" :class="getJobStatusClass(job.phase)">{{ job.phase }}</span>
                  </td>
                  <td>{{ formatTime(job.startTime) }}</td>
                  <td>{{ formatTime(job.completionTime) }}</td>
                  <td>{{ job.active || 0 }} / {{ job.succeeded || 0 }} / {{ job.failed || 0 }}</td>
                  <td>{{ calculateDuration(job.startTime, job.completionTime) }}</td>
                  <td>
                    <div class="job-action-buttons">
                      <button class="btn-icon" @click="viewJobPods(job)" title="查看 Pod">
                        📦
                      </button>
                      <button class="btn-icon" @click="viewJobLogs(job)" title="查看日志">
                        📄
                      </button>
                      <button class="btn-icon" @click="viewJobDetail(job)" title="查看详情">
                        ℹ️
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
        <div class="modal-footer">
          <button @click="closeJobsModal" class="btn btn-secondary">关闭</button>
        </div>
      </div>
    </div>

    <!-- Job Pods 列表模态框 -->
    <div v-if="showJobPodsModal" class="modal-overlay" @click="closeJobPodsModal">
      <div class="modal-content jobs-list-modal" @click.stop>
        <div class="modal-header">
          <h3>📦 Job 关联的 Pod 列表 - {{ selectedJobForPods?.name }}</h3>
          <button @click="closeJobPodsModal" class="close-btn">×</button>
        </div>
        <div class="modal-body">
          <div v-if="loadingJobPods" class="loading-state">加载中...</div>
          <div v-else-if="jobPodsList.length === 0" class="empty-state">
            <p style="color: #999; text-align: center; padding: 40px;">暂无关联的 Pod</p>
          </div>
          <div v-else>
            <table class="detail-table">
              <thead>
                <tr>
                  <th>Pod 名称</th>
                  <th>状态</th>
                  <th>IP</th>
                  <th>节点</th>
                  <th>容器数</th>
                  <th>重启次数</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="pod in jobPodsList" :key="pod.name">
                  <td>{{ pod.name }}</td>
                  <td>
                    <span class="status-badge" :class="pod.status?.toLowerCase()">{{ pod.status }}</span>
                  </td>
                  <td>{{ pod.pod_ip || '-' }}</td>
                  <td>{{ pod.node_name || '-' }}</td>
                  <td>{{ pod.containers?.length || 0 }}</td>
                  <td>{{ pod.restart_count || 0 }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
        <div class="modal-footer">
          <button @click="closeJobPodsModal" class="btn btn-secondary">关闭</button>
        </div>
      </div>
    </div>

    <!-- Job 日志模态框 -->
    <div v-if="showJobLogsModal" class="modal-overlay" @click.self="closeJobLogs">
      <div class="modal-content logs-modal" @click.stop>
        <div class="modal-header">
          <h3>📄 Job 日志 - {{ selectedJobForLogs?.name }}</h3>
          <button class="close-btn" @click="closeJobLogs">×</button>
        </div>
        <div class="modal-body">
          <div class="logs-controls">
            <div class="control-item">
              <label>Pod 选择</label>
              <select v-model="jobLogsForm.selectedPod" class="form-select" @change="onJobPodChange">
                <option value="">选择 Pod</option>
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

          <div v-if="jobLogsError" class="error-box">{{ jobLogsError }}</div>
          
          <div class="logs-content-wrapper">
            <pre class="logs-content">{{ jobLogsContent || '请选择 Pod 和容器后获取日志' }}</pre>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeJobLogs">关闭</button>
        </div>
      </div>
    </div>

    <!-- Job 详情模态框 -->
    <div v-if="showJobDetailModal" class="modal-overlay" @click="closeJobDetailModal">
      <div class="modal-content large" @click.stop>
        <div class="modal-header">
          <h3>ℹ️ Job 详情 - {{ selectedJobForDetail?.name }}</h3>
          <button @click="closeJobDetailModal" class="close-btn">×</button>
        </div>
        <div class="modal-body">
          <div v-if="loadingJobDetail" class="loading-state">加载中...</div>
          <div v-else-if="jobDetailData" class="detail-content">
            <div class="detail-section">
              <h4>基本信息</h4>
              <table class="detail-table">
                <tbody>
                  <tr>
                    <td class="label">名称:</td>
                    <td>{{ jobDetailData.name }}</td>
                  </tr>
                  <tr>
                    <td class="label">命名空间:</td>
                    <td>{{ jobDetailData.namespace }}</td>
                  </tr>
                  <tr>
                    <td class="label">状态:</td>
                    <td>
                      <span class="status-badge" :class="jobDetailData.status?.toLowerCase()">
                        {{ jobDetailData.status }}
                      </span>
                    </td>
                  </tr>
                  <tr>
                    <td class="label">镜像:</td>
                    <td>{{ jobDetailData.image || '-' }}</td>
                  </tr>
                  <tr>
                    <td class="label">开始时间:</td>
                    <td>{{ jobDetailData.start_time || '-' }}</td>
                  </tr>
                  <tr>
                    <td class="label">完成时间:</td>
                    <td>{{ jobDetailData.completion_time || '-' }}</td>
                  </tr>
                </tbody>
              </table>
            </div>

            <div class="detail-section">
              <h4>执行信息</h4>
              <table class="detail-table">
                <tbody>
                  <tr>
                    <td class="label">完成数:</td>
                    <td>{{ jobDetailData.succeeded || 0 }} / {{ jobDetailData.completions || 1 }}</td>
                  </tr>
                  <tr>
                    <td class="label">并行度:</td>
                    <td>{{ jobDetailData.parallelism || 1 }}</td>
                  </tr>
                  <tr>
                    <td class="label">退避限制:</td>
                    <td>{{ jobDetailData.backoff_limit || 6 }}</td>
                  </tr>
                  <tr>
                    <td class="label">活跃 Pod:</td>
                    <td>{{ jobDetailData.active || 0 }}</td>
                  </tr>
                  <tr>
                    <td class="label">失败次数:</td>
                    <td>{{ jobDetailData.failed || 0 }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button @click="closeJobDetailModal" class="btn btn-secondary">关闭</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import Pagination from '@/components/Pagination.vue'
import cronjobsApi from '@/api/cluster/workloads/cronjobs'
import namespaceApi from '@/api/cluster/config/namespace'

// ==================== 状态管理 ====================
const loading = ref(false)
const errorMsg = ref('')
const cronjobs = ref([])
const total = ref(0)

// 视图模式
const viewMode = ref('table') // 'table' | 'card'

// 批量操作
const batchMode = ref(false)
const selectedCronjobs = ref([])
const showBatchDeleteModal = ref(false)
const deleteConfirmText = ref('')
const batchExecuting = ref(false)

// 搜索和过滤
const searchQuery = ref('')
const namespaceFilter = ref('')
const statusFilter = ref('all')

// 分页
const currentPage = ref(1)
const itemsPerPage = ref(10)

// 自动刷新
const autoRefresh = ref(false)
let refreshTimer = null

// 模态框状态
const showCreateModal = ref(false)
const showDetailModal = ref(false)
const showDeleteModal = ref(false)
const showJobsModal = ref(false)  // Job 列表模态框
const showJobPodsModal = ref(false)  // Job Pods 列表模态框
const showJobLogsModal = ref(false)  // Job 日志模态框
const showJobDetailModal = ref(false)  // Job 详情模态框

// 操作状态
const creating = ref(false)
const deleting = ref(false)
const loadingDetail = ref(false)
const loadingJobs = ref(false)    // 加载 Job 列表

// 选中的 CronJob
const selectedCronjob = ref(null)
const selectedCronjobForJobs = ref(null)  // 用于 Job 列表查看
const detailData = ref(null)
const cronjobJobs = ref([])  // CronJob 关联的 Job 列表

// Job 操作相关
const selectedJobForDetail = ref(null)  // 选中的 Job (查看详情)
const selectedJobForPods = ref(null)  // 选中的 Job (查看 Pods)
const selectedJobForLogs = ref(null)  // 选中的 Job (查看日志)
const jobPodsList = ref([])  // Job 关联的 Pod 列表
const jobDetailData = ref(null)  // Job 详情数据
const loadingJobDetail = ref(false)  // 加载 Job 详情
const loadingJobPods = ref(false)  // 加载 Job Pods

// Job 日志相关
const jobLogsContent = ref('')
const jobLogsError = ref('')
const loadingJobLogs = ref(false)
const isStreamingJobLogs = ref(false)
let jobLogAbortController = null
const jobLogsForm = ref({
  selectedPod: '',
  container: '',
  tail: 100,
  follow: false
})
const jobContainerList = ref([])

// 更多菜单
const showMoreOptions = ref(false)
const menuStyle = ref({})

// YAML 查看/编辑
const showYamlModal = ref(false)
const selectedYamlCronjob = ref(null)
const yamlViewContent = ref('')
const yamlEditMode = ref(false)
const loadingYaml = ref(false)
const savingYaml = ref(false)
const yamlViewError = ref('')

// 创建模式切换
const createMode = ref('form') // 'form' | 'yaml'
const createYamlContent = ref('')
const createYamlError = ref('')
const creatingFromYaml = ref(false)

// 命名空间相关
const namespaces = ref(['default', 'kube-system', 'kube-public'])
const showNamespaceInput = ref(false)
const newNamespace = ref('')
const creatingNamespace = ref(false)

// 创建表单
const cronjobForm = ref({
  name: '',
  namespace: 'default',
  schedule: '0 * * * *',
  container_image: '',
  container_name: '',
  container_command: '',
  container_args: '',
  successful_jobs_history_limit: 3,
  failed_jobs_history_limit: 1,
  suspend: false
})

// ==================== 计算属性 ====================
const filteredCronjobs = computed(() => {
  let result = cronjobs.value

  // 命名空间过滤
  if (namespaceFilter.value) {
    result = result.filter(cj => cj.namespace === namespaceFilter.value)
  }

  // 状态过滤
  if (statusFilter.value !== 'all') {
    result = result.filter(cj => cj.status === statusFilter.value)
  }

  // 搜索过滤
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(cj =>
      cj.name.toLowerCase().includes(query) ||
      cj.namespace.toLowerCase().includes(query)
    )
  }

  return result
})

const paginatedCronjobs = computed(() => {
  const startIndex = (currentPage.value - 1) * itemsPerPage.value
  const endIndex = startIndex + itemsPerPage.value
  return filteredCronjobs.value.slice(startIndex, endIndex)
})

// ==================== 方法 ====================
// 获取 CronJob 列表
const fetchCronjobs = async () => {
  try {
    loading.value = true
    errorMsg.value = ''

    const res = await cronjobsApi.list({
      namespace: '',  // 查询所有命名空间
      page: 1,
      limit: 1000  // 前端分页
    })

    if (res.code === 0 && res.data) {
      const list = res.data.list || []
      cronjobs.value = list.map(item => ({
        name: item.name,
        namespace: item.namespace,
        status: item.status,
        schedule: item.schedule,
        suspend: item.suspend,
        last_schedule_time: item.last_schedule_time,
        active: item.active,
        successful_jobs_history_limit: item.successful_jobs_history_limit,
        failed_jobs_history_limit: item.failed_jobs_history_limit,
        image: item.image || '',
        containers: item.containers || [],
        created_at: item.created_at
      }))
      total.value = res.data.total || list.length
    }
  } catch (error) {
    console.error('获取CronJob列表失败:', error)
    errorMsg.value = error.kube_message_error || error.message || '获取CronJob列表失败'
  } finally {
    loading.value = false
  }
}

// 获取命名空间列表
const fetchNamespaces = async () => {
  try {
    const res = await namespaceApi.list({ page: 1, limit: 1000 })
    const list = res?.data?.list || res?.data?.items || []
    namespaces.value = (Array.isArray(list) ? list : []).map(ns =>
      typeof ns === 'string' ? ns : (ns?.metadata?.name || ns?.name || ns)
    ).filter(Boolean)
    
    // 如果没有获取到命名空间，使用默认值
    if (namespaces.value.length === 0) {
      namespaces.value = ['default', 'kube-system', 'kube-public']
    }
  } catch (e) {
    console.error('获取命名空间列表失败:', e)
    // 失败时使用默认命名空间
    namespaces.value = ['default', 'kube-system', 'kube-public']
  }
}

// 刷新列表
const refreshList = () => {
  fetchCronjobs()
}

// 搜索输入处理
const onSearchInput = () => {
  currentPage.value = 1
}

// 设置状态过滤
const setStatusFilter = (status) => {
  statusFilter.value = status
  currentPage.value = 1
}

// 获取状态统计
const getStatusCount = (status) => {
  return cronjobs.value.filter(cj => cj.status === status).length
}

// 格式化时间（相对时间）
const getAge = (timeStr) => {
  if (!timeStr) return '-'
  try {
    const time = new Date(timeStr)
    const now = new Date()
    const diff = now - time
    const minutes = Math.floor(diff / 60000)
    const hours = Math.floor(minutes / 60)
    const days = Math.floor(hours / 24)

    if (days > 0) return `${days}d`
    if (hours > 0) return `${hours}h`
    if (minutes > 0) return `${minutes}m`
    return '刚刚'
  } catch {
    return timeStr
  }
}

// 格式化时间（完整时间）
const formatTime = (time) => {
  if (!time) return '-'
  try {
    return new Date(time).toLocaleString('zh-CN')
  } catch {
    return time
  }
}

// 创建命名空间
const createNewNamespace = async () => {
  const ns = newNamespace.value.trim()
  if (!ns) {
    alert('请输入命名空间名称')
    return
  }

  try {
    creatingNamespace.value = true
    await namespaceApi.create({ name: ns })

    // 重新获取命名空间列表
    await fetchNamespaces()
    
    // 设置为当前选中的命名空间
    cronjobForm.value.namespace = ns
    showNamespaceInput.value = false
    newNamespace.value = ''
    alert(`命名空间 ${ns} 创建成功`)
  } catch (error) {
    console.error('创建命名空间失败:', error)
    alert(error.kube_message_error || error.message || '创建命名空间失败')
  } finally {
    creatingNamespace.value = false
  }
}

const cancelCreateNamespace = () => {
  showNamespaceInput.value = false
  newNamespace.value = ''
}

// 创建 CronJob
const createCronjob = async () => {
  if (!cronjobForm.value.name || !cronjobForm.value.schedule || !cronjobForm.value.container_image) {
    alert('请填写必填项')
    return
  }

  try {
    creating.value = true
    
    // 构建请求数据，处理命令参数
    const requestData = {
      ...cronjobForm.value,
      // 将多行文本转换为数组
      container_command: cronjobForm.value.container_command 
        ? cronjobForm.value.container_command.split('\n').map(s => s.trim()).filter(Boolean)
        : null,
      container_command_args: cronjobForm.value.container_args
        ? cronjobForm.value.container_args.split('\n').map(s => s.trim()).filter(Boolean)
        : null
    }
    
    // 删除前端用的字段
    delete requestData.container_args
    
    await cronjobsApi.create(requestData)

    closeCreateModal()
    // 重置表单
    resetCronjobForm()

    fetchCronjobs()
  } catch (error) {
    console.error('创建CronJob失败:', error)
    alert(error.kube_message_error || error.message || '创建CronJob失败')
  } finally {
    creating.value = false
  }
}

// 重置表单
const resetCronjobForm = () => {
  cronjobForm.value = {
    name: '',
    namespace: 'default',
    schedule: '0 * * * *',
    container_image: '',
    container_name: '',
    container_command: '',
    container_args: '',
    successful_jobs_history_limit: 3,
    failed_jobs_history_limit: 1,
    suspend: false
  }
}

// 关闭创建模态框
const closeCreateModal = () => {
  showCreateModal.value = false
  showNamespaceInput.value = false
  newNamespace.value = ''
  createMode.value = 'form'
  createYamlContent.value = ''
  createYamlError.value = ''
}

// 加载 CronJob YAML 模板
const loadCronjobYamlTemplate = () => {
  createYamlContent.value = `apiVersion: batch/v1
kind: CronJob
metadata:
  name: my-cronjob
  namespace: default
spec:
  schedule: "*/5 * * * *"  # 每5分钟执行一次
  concurrencyPolicy: Forbid  # 禁止并发执行
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 1
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: hello
            image: busybox:latest
            imagePullPolicy: IfNotPresent
            command:
            - /bin/sh
            - -c
            args:
            - echo "Hello from CronJob at $(date)" && sleep 10
          restartPolicy: OnFailure`
  createYamlError.value = ''
}

// 复制 YAML 内容到剪贴板
const copyYamlContent = async () => {
  if (!createYamlContent.value.trim()) {
    alert('没有内容可复制')
    return
  }
  try {
    await navigator.clipboard.writeText(createYamlContent.value)
    alert('YAML 内容已复制到剪贴板')
  } catch (err) {
    console.error('复制失败:', err)
    alert('复制失败，请手动复制')
  }
}

// 重置 YAML 内容
const resetYamlContent = () => {
  if (createYamlContent.value.trim() && !confirm('确定要重置 YAML 内容吗？')) {
    return
  }
  createYamlContent.value = ''
  createYamlError.value = ''
}

// 从 YAML 创建 CronJob
const createCronjobFromYaml = async () => {
  if (!createYamlContent.value.trim()) {
    createYamlError.value = '请输入 YAML 内容'
    return
  }

  // 简单验证 YAML 格式
  if (!createYamlContent.value.includes('kind: CronJob')) {
    createYamlError.value = 'YAML 中必须包含 "kind: CronJob"'
    return
  }
  if (!createYamlContent.value.includes('apiVersion: batch/v1')) {
    createYamlError.value = 'YAML 中必须包含 "apiVersion: batch/v1"'
    return
  }

  createYamlError.value = ''
  creatingFromYaml.value = true

  try {
    const res = await cronjobsApi.createFromYaml({ yaml: createYamlContent.value })
    
    if (res.code === 0) {
      alert(`CronJob ${res.data?.name || ''} 创建成功`)
      closeCreateModal()
      fetchCronjobs()
    } else {
      createYamlError.value = res.msg || '创建失败'
    }
  } catch (error) {
    console.error('YAML 创建失败:', error)
    createYamlError.value = error.kube_message_error || error.message || 'YAML 创建失败'
  } finally {
    creatingFromYaml.value = false
  }
}

// 命令模板
const applyCommandTemplate = (template) => {
  const templates = {
    shell: {
      image: 'busybox:latest',
      command: '/bin/sh\n-c',
      args: 'echo "Hello from CronJob" && date && sleep 5'
    },
    echo: {
      image: 'busybox:latest',
      command: '/bin/sh\n-c',
      args: 'echo "CronJob executed at $(date)"'
    },
    curl: {
      image: 'curlimages/curl:latest',
      command: '/bin/sh\n-c',
      args: 'curl -s https://httpbin.org/get | head -20'
    },
    backup: {
      image: 'busybox:latest',
      command: '/bin/sh\n-c',
      args: 'echo "Starting backup..." && tar -czf /backup/data-$(date +%Y%m%d).tar.gz /data 2>/dev/null || echo "Backup simulation completed"'
    }
  }

  const tpl = templates[template]
  if (tpl) {
    cronjobForm.value.container_image = tpl.image
    cronjobForm.value.container_command = tpl.command
    cronjobForm.value.container_args = tpl.args
  }
}

// 暂停/恢复 CronJob
const suspendCronJob = async (cj, suspend) => {
  showMoreOptions.value = false
  
  const action = suspend ? '暂停' : '恢复'
  if (!confirm(`确定要${action} CronJob ${cj.namespace}/${cj.name} 吗？`)) {
    return
  }

  try {
    await cronjobsApi.suspend({
      namespace: cj.namespace,
      name: cj.name,
      suspend: suspend
    })

    fetchCronjobs()
  } catch (error) {
    console.error(`${action}CronJob失败:`, error)
    alert(error.kube_message_error || error.message || `${action}CronJob失败`)
  }
}

// 查看 Job 列表
const viewJobs = async (cj) => {
  showMoreOptions.value = false
  selectedCronjobForJobs.value = cj
  showJobsModal.value = true
  loadingJobs.value = true
  cronjobJobs.value = []
  
  try {
    // 获取 CronJob 详情（包含 Job 列表）
    const res = await cronjobsApi.detail({
      namespace: cj.namespace,
      name: cj.name
    })
    
    if (res.code === 0 && res.data && res.data.jobs) {
      cronjobJobs.value = res.data.jobs
    }
  } catch (error) {
    console.error('获取 Job 列表失败:', error)
  } finally {
    loadingJobs.value = false
  }
}

// 关闭 Job 列表模态框
const closeJobsModal = () => {
  showJobsModal.value = false
  selectedCronjobForJobs.value = null
  cronjobJobs.value = []
}

// 获取 Job 状态类名
const getJobStatusClass = (phase) => {
  const lowerPhase = (phase || '').toLowerCase()
  if (lowerPhase === 'complete' || lowerPhase === 'succeeded') return 'success'
  if (lowerPhase === 'failed') return 'danger'
  if (lowerPhase === 'running') return 'running'
  return 'pending'
}

// 计算耗时
const calculateDuration = (startTime, endTime) => {
  if (!startTime) return '-'
  
  const start = new Date(startTime)
  const end = endTime ? new Date(endTime) : new Date()
  const diffMs = end - start
  
  if (diffMs < 0) return '-'
  
  const seconds = Math.floor(diffMs / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)
  
  if (days > 0) return `${days}天${hours % 24}小时`
  if (hours > 0) return `${hours}小时${minutes % 60}分`
  if (minutes > 0) return `${minutes}分${seconds % 60}秒`
  return `${seconds}秒`
}

// ========== Job 操作函数 ==========
// 查看 Job 的 Pods
const viewJobPods = async (job) => {
  selectedJobForPods.value = job
  showJobPodsModal.value = true
  loadingJobPods.value = true
  jobPodsList.value = []
  
  try {
    const res = await fetch(`/api/v1/k8s/pod/list?namespace=${job.namespace}&limit=100`, {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })
    
    if (res.ok) {
      const data = await res.json()
      // 过滤出属于该 Job 的 Pod（通过 Job 名称前缀匹配）
      const jobPods = (data.data?.list || []).filter(pod => 
        pod.name.startsWith(job.name + '-')
      )
      jobPodsList.value = jobPods
    }
  } catch (error) {
    console.error('获取 Pod 列表失败:', error)
  } finally {
    loadingJobPods.value = false
  }
}

// 关闭 Job Pods 模态框
const closeJobPodsModal = () => {
  showJobPodsModal.value = false
  selectedJobForPods.value = null
  jobPodsList.value = []
}

// 查看 Job 日志
const viewJobLogs = async (job) => {
  selectedJobForLogs.value = job
  showJobLogsModal.value = true
  jobLogsContent.value = ''
  jobLogsError.value = ''
  jobLogsForm.value = {
    selectedPod: '',
    container: '',
    tail: 100,
    follow: false
  }
  
  // 首先获取 Job 关联的 Pod 列表
  try {
    const res = await fetch(`/api/v1/k8s/pod/list?namespace=${job.namespace}&limit=100`, {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })
    
    if (res.ok) {
      const data = await res.json()
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
  } catch (error) {
    console.error('获取 Pod 列表失败:', error)
  }
}

// Pod 选择变化
const onJobPodChange = () => {
  const pod = jobPodsList.value.find(p => p.name === jobLogsForm.value.selectedPod)
  if (pod) {
    jobContainerList.value = pod.containers || []
    // 如果只有一个容器，自动选中
    if (jobContainerList.value.length === 1) {
      jobLogsForm.value.container = jobContainerList.value[0]
    } else {
      jobLogsForm.value.container = ''
    }
  }
}

// 获取 Job 日志
const fetchJobLogs = async () => {
  if (!jobLogsForm.value.selectedPod || !jobLogsForm.value.container) {
    jobLogsError.value = '请选择 Pod 和容器'
    return
  }
  
  jobLogsError.value = ''
  jobLogsContent.value = ''
  
  const params = new URLSearchParams({
    namespace: selectedJobForLogs.value.namespace,
    pod_name: jobLogsForm.value.selectedPod,
    container: jobLogsForm.value.container
  })
  
  if (jobLogsForm.value.tail !== null) {
    params.append('tail', jobLogsForm.value.tail)
  }
  
  if (jobLogsForm.value.follow) {
    params.append('follow', 'true')
    // 流式日志
    jobLogAbortController = new AbortController()
    isStreamingJobLogs.value = true
    loadingJobLogs.value = true
    
    try {
      const response = await fetch(`/api/v1/k8s/pod/container_logs?${params.toString()}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        signal: jobLogAbortController.signal
      })
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`)
      }
      
      const reader = response.body.getReader()
      const decoder = new TextDecoder()
      
      while (true) {
        const { done, value } = await reader.read()
        if (done) break
        
        const chunk = decoder.decode(value, { stream: true })
        jobLogsContent.value += chunk
      }
    } catch (error) {
      if (error.name !== 'AbortError') {
        console.error('流式日志错误:', error)
        jobLogsError.value = error.message || '获取日志失败'
      }
    } finally {
      isStreamingJobLogs.value = false
      loadingJobLogs.value = false
      jobLogAbortController = null
    }
  } else {
    // 静态日志
    loadingJobLogs.value = true
    
    try {
      const response = await fetch(`/api/v1/k8s/pod/container_logs?${params.toString()}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      })
      
      const data = await response.json()
      
      if (data.code === 0) {
        jobLogsContent.value = data.data?.logs || '无日志内容'
      } else {
        jobLogsError.value = data.message || '获取日志失败'
      }
    } catch (error) {
      console.error('获取日志失败:', error)
      jobLogsError.value = error.message || '获取日志失败'
    } finally {
      loadingJobLogs.value = false
    }
  }
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

// 关闭 Job 日志模态框
const closeJobLogs = () => {
  stopJobLogStream()
  showJobLogsModal.value = false
  selectedJobForLogs.value = null
  jobLogsContent.value = ''
  jobLogsError.value = ''
}

// 查看 Job 详情
const viewJobDetail = async (job) => {
  selectedJobForDetail.value = job
  showJobDetailModal.value = true
  loadingJobDetail.value = true
  jobDetailData.value = null
  
  try {
    const res = await fetch(`/api/v1/k8s/job/detail?namespace=${job.namespace}&name=${job.name}`, {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })
    
    const data = await res.json()
    if (data.code === 0) {
      jobDetailData.value = data.data || data
    } else {
      alert(data.message || '获取详情失败')
    }
  } catch (error) {
    console.error('获取 Job 详情失败:', error)
    alert(error.message || '获取 Job 详情失败')
  } finally {
    loadingJobDetail.value = false
  }
}

// 关闭 Job 详情模态框
const closeJobDetailModal = () => {
  showJobDetailModal.value = false
  selectedJobForDetail.value = null
  jobDetailData.value = null
}

// 切换更多菜单
const toggleMoreOptions = (cj, event) => {
  if (showMoreOptions.value && selectedCronjob.value === cj) {
    showMoreOptions.value = false
    return
  }

  selectedCronjob.value = cj
  showMoreOptions.value = true

  // 计算菜单位置
  const rect = event.target.getBoundingClientRect()
  menuStyle.value = {
    position: 'fixed',
    top: `${rect.bottom + 5}px`,
    right: `${window.innerWidth - rect.right}px`,
    zIndex: 1000
  }

  // 点击其他地方关闭菜单
  const closeMenu = (e) => {
    if (!e.target.closest('.more-btn')) {
      showMoreOptions.value = false
      document.removeEventListener('click', closeMenu)
    }
  }
  setTimeout(() => {
    document.addEventListener('click', closeMenu)
  }, 0)
}

// 手动触发 CronJob
const triggerCronJob = async (cj) => {
  showMoreOptions.value = false
  
  if (!confirm(`确定要立即触发 CronJob ${cj.namespace}/${cj.name} 吗？\n\n这将创建一个新的 Job 立即执行。`)) {
    return
  }

  try {
    const res = await cronjobsApi.trigger({
      namespace: cj.namespace,
      name: cj.name
    })

    if (res.code === 0) {
      alert(`CronJob 已手动触发\n\n创建的 Job: ${res.data?.job_name || 'unknown'}`)
      fetchCronjobs() // 刷新列表
    } else {
      alert(res.message || '触发失败')
    }
  } catch (error) {
    console.error('手动触发失败:', error)
    alert(error.kube_message_error || error.message || '手动触发失败')
  }
}

// 查看详情
const viewDetail = async (cj) => {
  showMoreOptions.value = false
  
  selectedCronjob.value = cj
  showDetailModal.value = true
  loadingDetail.value = true
  detailData.value = null

  try {
    const res = await cronjobsApi.detail({
      namespace: cj.namespace,
      name: cj.name
    })

    if (res.code === 0 && res.data) {
      detailData.value = res.data
    }
  } catch (error) {
    console.error('获取详情失败:', error)
    alert(error.kube_message_error || error.message || '获取详情失败')
    showDetailModal.value = false
  } finally {
    loadingDetail.value = false
  }
}

// 确认删除
const confirmDelete = (cj) => {
  showMoreOptions.value = false
  
  selectedCronjob.value = cj
  showDeleteModal.value = true
}

// 删除 CronJob
const deleteCronjob = async () => {
  try {
    deleting.value = true
    await cronjobsApi.delete({
      namespace: selectedCronjob.value.namespace,
      name: selectedCronjob.value.name
    })

    showDeleteModal.value = false
    selectedCronjob.value = null
    fetchCronjobs()
  } catch (error) {
    console.error('删除CronJob失败:', error)
    alert(error.kube_message_error || error.message || '删除CronJob失败')
  } finally {
    deleting.value = false
  }
}

// ==================== 批量操作 ====================
const enterBatchMode = () => {
  batchMode.value = true
  // 默认全选当前页，用户可取消不需要的项
  selectedCronjobs.value = [...paginatedCronjobs.value]
}

const exitBatchMode = () => {
  batchMode.value = false
  selectedCronjobs.value = []
}

const clearSelection = () => {
  selectedCronjobs.value = []
}

const isCronjobSelected = (cj) => {
  return selectedCronjobs.value.some(c => c.name === cj.name && c.namespace === cj.namespace)
}

const toggleCronjobSelection = (cj) => {
  const index = selectedCronjobs.value.findIndex(c => c.name === cj.name && c.namespace === cj.namespace)
  if (index >= 0) {
    selectedCronjobs.value.splice(index, 1)
  } else {
    selectedCronjobs.value.push(cj)
  }
}

const isAllSelected = computed(() => {
  return paginatedCronjobs.value.length > 0 &&
         paginatedCronjobs.value.every(cj => isCronjobSelected(cj))
})

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    paginatedCronjobs.value.forEach(cj => {
      const index = selectedCronjobs.value.findIndex(c => c.name === cj.name && c.namespace === cj.namespace)
      if (index >= 0) selectedCronjobs.value.splice(index, 1)
    })
  } else {
    paginatedCronjobs.value.forEach(cj => {
      if (!isCronjobSelected(cj)) {
        selectedCronjobs.value.push(cj)
      }
    })
  }
}

// 批量暂停/恢复
const batchSuspend = async (suspend) => {
  const action = suspend ? '暂停' : '恢复'
  if (!confirm(`确定要${action} ${selectedCronjobs.value.length} 个 CronJob 吗？`)) {
    return
  }

  batchExecuting.value = true
  let successCount = 0
  let failCount = 0

  for (const cj of selectedCronjobs.value) {
    try {
      await cronjobsApi.suspend({
        namespace: cj.namespace,
        name: cj.name,
        suspend: suspend
      })
      successCount++
    } catch (e) {
      console.error(`Failed to ${action} ${cj.name}:`, e)
      failCount++
    }
  }

  batchExecuting.value = false

  if (failCount === 0) {
    alert(`成功${action} ${successCount} 个 CronJob`)
  } else {
    alert(`成功 ${successCount} 个，失败 ${failCount} 个`)
  }

  exitBatchMode()
  fetchCronjobs()
}

// 批量删除
const openBatchDeletePreview = () => {
  deleteConfirmText.value = ''
  showBatchDeleteModal.value = true
}

const closeBatchDeleteModal = () => {
  showBatchDeleteModal.value = false
  deleteConfirmText.value = ''
}

const executeBatchDelete = async () => {
  if (deleteConfirmText.value !== 'DELETE') return

  batchExecuting.value = true
  let successCount = 0
  let failCount = 0

  for (const cj of selectedCronjobs.value) {
    try {
      await cronjobsApi.delete({
        namespace: cj.namespace,
        name: cj.name
      })
      successCount++
    } catch (e) {
      console.error(`Failed to delete ${cj.name}:`, e)
      failCount++
    }
  }

  batchExecuting.value = false
  showBatchDeleteModal.value = false
  deleteConfirmText.value = ''

  if (failCount === 0) {
    alert(`成功删除 ${successCount} 个 CronJob`)
  } else {
    alert(`成功 ${successCount} 个，失败 ${failCount} 个`)
  }

  exitBatchMode()
  fetchCronjobs()
}

// ==================== YAML 操作 ====================
const viewYaml = async (cj) => {
  showMoreOptions.value = false
  
  selectedYamlCronjob.value = cj
  showYamlModal.value = true
  loadingYaml.value = true
  yamlViewError.value = ''
  yamlViewContent.value = ''
  yamlEditMode.value = false

  try {
    // 注意：CronJob 后端需要实现 getYaml API
    // 这里假设后端有类似 Job 的 yaml 接口
    const res = await cronjobsApi.detail({
      namespace: cj.namespace,
      name: cj.name
    })

    if (res.code === 0 && res.data) {
      // 如果后端返回 YAML，直接使用
      // 否则简单展示提示
      yamlViewContent.value = `# CronJob YAML 查看功能需要后端支持
# 当前仅显示部分信息

apiVersion: batch/v1
kind: CronJob
metadata:
  name: ${cj.name}
  namespace: ${cj.namespace}
spec:
  schedule: "${cj.schedule}"
  suspend: ${cj.suspend}
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: ${cj.containers?.[0] || 'container'}
            image: ${cj.image || 'busybox'}
          restartPolicy: OnFailure
`
    }
  } catch (error) {
    console.error('获取 YAML 失败:', error)
    yamlViewError.value = error.kube_message_error || error.message || '获取 YAML 失败'
  } finally {
    loadingYaml.value = false
  }
}

const closeYamlModal = () => {
  showYamlModal.value = false
  selectedYamlCronjob.value = null
  yamlViewContent.value = ''
  yamlEditMode.value = false
  yamlViewError.value = ''
}

const downloadYaml = () => {
  const blob = new Blob([yamlViewContent.value], { type: 'text/yaml' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${selectedYamlCronjob.value.namespace}-${selectedYamlCronjob.value.name}.yaml`
  a.click()
  URL.revokeObjectURL(url)
}

const applyYamlChanges = async () => {
  if (!yamlViewContent.value.trim()) {
    alert('请输入 YAML 内容')
    return
  }

  if (!confirm('确定要应用 YAML 更改吗？')) {
    return
  }

  try {
    savingYaml.value = true
    // 注意：需要后端实现 applyYaml API
    alert('YAML 编辑功能需要后端支持 apply-yaml 接口')
    closeYamlModal()
  } catch (error) {
    console.error('应用 YAML 失败:', error)
    alert(error.kube_message_error || error.message || '应用 YAML 失败')
  } finally {
    savingYaml.value = false
  }
}

// 自动刷新逻辑
watch(autoRefresh, (newVal) => {
  if (newVal) {
    refreshTimer = setInterval(fetchCronjobs, 90000) // 90秒刷新
  } else {
    if (refreshTimer) {
      clearInterval(refreshTimer)
      refreshTimer = null
    }
  }
})

// ==================== 生命周期 ====================
onMounted(() => {
  fetchNamespaces()  // 先获取命名空间列表
  fetchCronjobs()    // 再获取 CronJob 列表
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})
</script>

<style scoped>
/* 基础布局 */
.resource-view {
  padding: 20px;
  background-color: #f5f7fa;
  min-height: 100vh;
}

.view-header {
  margin-bottom: 24px;
}

.view-header h1 {
  font-size: 28px;
  font-weight: 600;
  color: #2c3e50;
  margin: 0 0 8px 0;
}

.view-header p {
  color: #7f8c8d;
  margin: 0;
  font-size: 14px;
}

/* 操作栏 */
.action-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
  align-items: center;
  flex-wrap: wrap;
}

.search-box input {
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  width: 250px;
  font-size: 14px;
}

.filter-buttons {
  display: flex;
  gap: 8px;
}

.btn-filter {
  padding: 8px 16px;
  border: 1px solid #ddd;
  background: white;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
  transition: all 0.2s;
}

.btn-filter:hover {
  background: #f8f9fa;
}

.btn-filter.active {
  background: #326ce5;
  color: white;
  border-color: #326ce5;
}

.filter-dropdown select {
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  background: white;
  font-size: 14px;
}

.action-buttons {
  display: flex;
  gap: 8px;
  margin-left: auto;
  align-items: center;
}

.auto-refresh-toggle {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  font-size: 14px;
  user-select: none;
}

.refresh-indicator {
  color: #4caf50;
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s;
}

.btn-primary {
  background: #326ce5;
  color: white;
}

.btn-primary:hover {
  background: #2554c7;
}

.btn-secondary {
  background: #6c757d;
  color: white;
}

.btn-secondary:hover {
  background: #5a6268;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 错误提示 */
.error-box {
  background: #fee;
  color: #c33;
  padding: 12px;
  border-radius: 4px;
  margin-bottom: 16px;
  border-left: 4px solid #c33;
}

/* 统计栏 */
.stats-bar {
  display: flex;
  gap: 24px;
  background: white;
  padding: 16px 20px;
  border-radius: 8px;
  margin-bottom: 16px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.stat-label {
  color: #666;
  font-size: 14px;
}

.stat-value {
  font-weight: 600;
  font-size: 18px;
  color: #2c3e50;
}

.stat-value.success {
  color: #4caf50;
}

.stat-value.warning {
  color: #ff9800;
}

/* 表格 */
.table-container {
  background: white;
  border-radius: 8px;
  overflow-x: auto;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.resource-table {
  width: 100%;
  border-collapse: collapse;
  min-width: 1400px;
}

.resource-table thead {
  background: #f8f9fa;
  border-bottom: 2px solid #dee2e6;
}

.resource-table th {
  padding: 12px;
  text-align: left;
  font-weight: 600;
  font-size: 13px;
  color: #495057;
  white-space: nowrap;
}

.resource-table td {
  padding: 12px;
  border-bottom: 1px solid #f0f0f0;
  font-size: 14px;
}

.resource-table tbody tr:hover {
  background: #f8f9fa;
}

/* 状态指示器 */
.status-cell {
  position: relative;
}

.status-indicator {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
  white-space: nowrap;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  display: inline-block;
}

.status-indicator.active {
  background: #d4edda;
  color: #155724;
}

.status-indicator.active .status-dot {
  background: #28a745;
}

.status-indicator.suspended {
  background: #fff3cd;
  color: #856404;
}

.status-indicator.suspended .status-dot {
  background: #ffc107;
}

/* Job名称 */
.job-name {
  display: flex;
  align-items: center;
  gap: 8px;
}

.job-name .icon {
  font-size: 16px;
}

.age-badge {
  background: #e9ecef;
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 11px;
  color: #6c757d;
}

/* 命名空间 */
.namespace-badge {
  background: #326ce5;
  color: white;
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

/* 调度表达式 */
.schedule-text {
  background: #f8f9fa;
  padding: 4px 8px;
  border-radius: 4px;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  color: #495057;
}

/* 挂起状态 */
.suspend-badge {
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.suspend-badge.suspended {
  background: #fff3cd;
  color: #856404;
}

.suspend-badge.active {
  background: #d4edda;
  color: #155724;
}

/* 活跃任务数 */
.active-count {
  font-weight: 600;
  color: #326ce5;
}

/* Job 执行统计 (Rancher/Kuboard 风格) */
.job-stats {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.job-stats .stat-item {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.job-stats .stat-item.success {
  background: #d4edda;
  color: #155724;
}

.job-stats .stat-item.danger {
  background: #f8d7da;
  color: #721c24;
}

.job-stats .stat-item.running {
  background: #cce5ff;
  color: #004085;
}

/* 时间文本 */
.time-text {
  color: #666;
  font-size: 13px;
}

/* 镜像 */
.image-text {
  max-width: 280px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.image-name {
  color: #1e40af;
  font-weight: 500;
  font-size: 13px;
}

/* 操作按钮 */
.action-cell {
  white-space: nowrap;
}

.action-buttons-cell {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

/* 操作图标布局 - 参考 Deployment.vue */
.action-icons {
  display: flex;
  align-items: center;
  gap: 8px;
}

.action-btn {
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.2s;
  white-space: nowrap;
}

.action-btn.primary {
  background: #326ce5;
  color: white;
}

.action-btn.primary:hover {
  background: #2355b8;
}

.icon-btn {
  padding: 6px 12px;
  background: #f8f9fa;
  border: 1px solid #ddd;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.2s;
  white-space: nowrap;
}

.icon-btn:hover {
  background: #e9ecef;
  border-color: #ccc;
}

.more-btn {
  position: relative;
}

.more-menu {
  background: white;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  min-width: 160px;
  overflow: hidden;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 10px 16px;
  border: none;
  background: none;
  cursor: pointer;
  font-size: 13px;
  color: #333;
  text-align: left;
  transition: background 0.2s;
}

.menu-item:hover {
  background: #f5f5f5;
}

.menu-item.danger {
  color: #dc3545;
}

.menu-item.danger:hover {
  background: #fff5f5;
}

.menu-icon {
  font-size: 14px;
}

.menu-divider {
  height: 1px;
  background: #eee;
  margin: 4px 0;
}

.btn-action {
  padding: 4px 10px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.2s;
  white-space: nowrap;
}

.btn-action.btn-warning {
  background: #ffc107;
  color: #333;
}

.btn-action.btn-warning:hover {
  background: #e0a800;
}

.btn-action.btn-success {
  background: #28a745;
  color: white;
}

.btn-action.btn-success:hover {
  background: #218838;
}

.btn-action.btn-info {
  background: #17a2b8;
  color: white;
}

.btn-action.btn-info:hover {
  background: #138496;
}

.btn-action.btn-danger {
  background: #dc3545;
  color: white;
}

.btn-action.btn-danger:hover {
  background: #c82333;
}

/* 加载动画 */
.loading-spinner {
  width: 40px;
  height: 40px;
  border: 4px solid #f3f3f3;
  border-top: 4px solid #326ce5;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

/* 分页 */
.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: center;
}

/* 模态框 */
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
}

.modal-content {
  background: white;
  border-radius: 12px;
  width: 90%;
  max-width: 600px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.2);
}

.modal-content.modal-large {
  max-width: 900px;
}

/* 高级模态框（Rancher/Kuboard 风格） */
.modal-content.advanced {
  max-width: 1000px;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
}

.modal-body.scrollable {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
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

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #e5e7eb;
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
}

.modal-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 700;
  color: #1f2937;
  display: flex;
  align-items: center;
  gap: 8px;
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
  color: #2c3e50;
}

.close-btn {
  background: none;
  border: none;
  font-size: 28px;
  cursor: pointer;
  color: #9ca3af;
  line-height: 1;
  padding: 0;
  width: 36px;
  height: 36px;
  border-radius: 8px;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.close-btn:hover {
  color: #374151;
  background: #e5e7eb;
}

.modal-body {
  padding: 20px;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 20px 24px;
  border-top: 1px solid #e5e7eb;
  background: #f9fafb;
}

/* 表单 */
.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 6px;
  font-weight: 500;
  font-size: 14px;
  color: #555;
}

.required {
  color: #dc3545;
}

.form-input,
.form-select {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.form-input:focus,
.form-select:focus {
  outline: none;
  border-color: #326ce5;
}

.help-text {
  display: block;
  margin-top: 4px;
  font-size: 12px;
  color: #6c757d;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-weight: normal;
}

.form-checkbox {
  cursor: pointer;
}

/* 命名空间选择器 */
.namespace-selector {
  display: flex;
  align-items: center;
  gap: 10px;
}

.namespace-or {
  color: #999;
  font-size: 13px;
}

.btn-create-ns {
  padding: 8px 12px;
  background: #28a745;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
  white-space: nowrap;
}

.btn-create-ns:hover {
  background: #218838;
}

.new-namespace-input {
  display: flex;
  gap: 8px;
  flex: 1;
}

.btn-confirm,
.btn-cancel {
  padding: 8px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  font-weight: bold;
}

.btn-confirm {
  background: #28a745;
  color: white;
}

.btn-confirm:hover {
  background: #218838;
}

.btn-cancel {
  background: #dc3545;
  color: white;
}

.btn-cancel:hover {
  background: #c82333;
}

/* 删除确认框 */
.delete-confirm .modal-body {
  text-align: center;
  padding: 30px 20px;
}

.warning-text {
  color: #dc3545;
  font-weight: 500;
  margin-top: 12px;
}

/* 详情表格 */
.detail-section {
  margin-bottom: 24px;
}

.detail-section h4 {
  margin: 0 0 12px 0;
  font-size: 16px;
  color: #2c3e50;
  border-bottom: 2px solid #326ce5;
  padding-bottom: 8px;
}

.detail-table {
  width: 100%;
  border-collapse: collapse;
}

.detail-table td {
  padding: 8px 12px;
  border-bottom: 1px solid #f0f0f0;
}

.detail-table td.label {
  font-weight: 500;
  color: #666;
  width: 180px;
}

.detail-table thead {
  background: #f8f9fa;
}

.detail-table th {
  padding: 10px 12px;
  text-align: left;
  font-weight: 600;
  font-size: 13px;
  color: #495057;
  border-bottom: 2px solid #dee2e6;
}

.status-badge {
  display: inline-block;
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.status-badge.complete {
  background: #d4edda;
  color: #155724;
}

.status-badge.running {
  background: #cce5ff;
  color: #004085;
}

.status-badge.failed {
  background: #f8d7da;
  color: #721c24;
}

.status-badge.pending {
  background: #fff3cd;
  color: #856404;
}

/* ====================  批量操作样式 ==================== */
.btn-batch {
  background: #17a2b8;
  color: white;
}

.btn-batch:hover {
  background: #138496;
}

.view-toggle {
  display: flex;
  gap: 4px;
  border: 1px solid #ddd;
  border-radius: 4px;
  overflow: hidden;
}

.btn-view {
  padding: 8px 12px;
  background: white;
  border: none;
  cursor: pointer;
  font-size: 16px;
  transition: all 0.2s;
}

.btn-view:hover {
  background: #f8f9fa;
}

.btn-view.active {
  background: #326ce5;
}

.batch-action-bar {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 12px 20px;
  border-radius: 8px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  box-shadow: 0 4px 6px rgba(0,0,0,0.1);
}

.batch-info {
  display: flex;
  align-items: center;
  gap: 16px;
}

.batch-count {
  font-weight: 600;
  font-size: 16px;
}

.batch-clear {
  background: rgba(255,255,255,0.2);
  color: white;
  border: 1px solid rgba(255,255,255,0.3);
  padding: 4px 12px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
}

.batch-clear:hover {
  background: rgba(255,255,255,0.3);
}

.batch-actions {
  display: flex;
  gap: 8px;
}

.batch-btn {
  padding: 6px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  transition: all 0.2s;
  color: white;
}

.batch-btn.warning {
  background: #ffc107;
  color: #333;
}

.batch-btn.warning:hover {
  background: #e0a800;
}

.batch-btn.success {
  background: #28a745;
}

.batch-btn.success:hover {
  background: #218838;
}

.batch-btn.danger {
  background: #dc3545;
}

.batch-btn.danger:hover {
  background: #c82333;
}

.row-selected {
  background: #e3f2fd !important;
}

/* ====================  卡片视图样式 ==================== */
.card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(380px, 1fr));
  gap: 20px;
  margin-bottom: 20px;
}

.resource-card {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
  transition: all 0.3s;
  position: relative;
  overflow: hidden;
}

.resource-card:hover {
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  transform: translateY(-2px);
}

.resource-card.card-selected {
  border: 2px solid #326ce5;
  box-shadow: 0 0 0 3px rgba(50,108,229,0.1);
}

.card-checkbox {
  position: absolute;
  top: 12px;
  left: 12px;
  z-index: 10;
}

.card-checkbox input[type="checkbox"] {
  width: 20px;
  height: 20px;
  cursor: pointer;
}

.card-header {
  padding: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 2px solid #f0f0f0;
}

.card-header.status-active {
  background: linear-gradient(135deg, #d4edda 0%, #c3e6cb 100%);
}

.card-header.status-suspended {
  background: linear-gradient(135deg, #fff3cd 0%, #ffeaa7 100%);
}

.card-title {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
}

.card-icon {
  font-size: 24px;
}

.card-title h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #2c3e50;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-body {
  padding: 16px;
}

.card-meta {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.card-meta-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 13px;
}

.card-meta-item .label {
  color: #666;
  font-weight: 500;
}

.card-footer {
  padding: 12px 16px;
  background: #f8f9fa;
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  border-top: 1px solid #e9ecef;
}

.card-btn {
  flex: 1;
  min-width: 80px;
  padding: 6px 12px;
  border: 1px solid #ddd;
  background: white;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.2s;
  white-space: nowrap;
}

.card-btn:hover {
  background: #f8f9fa;
  transform: translateY(-1px);
}

.card-btn.primary {
  background: #326ce5;
  color: white;
  border-color: #326ce5;
}

.card-btn.primary:hover {
  background: #2554c7;
}

.card-btn.warning {
  background: #ffc107;
  color: #333;
  border-color: #ffc107;
}

.card-btn.warning:hover {
  background: #e0a800;
}

.card-btn.success {
  background: #28a745;
  color: white;
  border-color: #28a745;
}

.card-btn.success:hover {
  background: #218838;
}

.card-btn.danger {
  background: #dc3545;
  color: white;
  border-color: #dc3545;
}

.card-btn.danger:hover {
  background: #c82333;
}

.loading-indicator,
.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: #666;
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.empty-title {
  font-size: 18px;
  font-weight: 600;
  color: #2c3e50;
  margin-bottom: 8px;
}

.empty-desc {
  font-size: 14px;
  color: #999;
}

/* ====================  YAML 模态框 ==================== */
.yaml-modal {
  max-width: 900px;
  max-height: 85vh;
}

.yaml-header-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.btn-sm {
  padding: 4px 12px;
  font-size: 13px;
}

.yaml-modal-body {
  padding: 0;
  max-height: calc(85vh - 150px);
  overflow: hidden;
}

.loading-state {
  text-align: center;
  padding: 40px;
  color: #666;
}

.yaml-editor-wrapper {
  height: 100%;
  overflow: auto;
}

.yaml-editor {
  width: 100%;
  min-height: 500px;
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

.yaml-content {
  padding: 16px;
  margin: 0;
  background: #f8f9fa;
  color: #2c3e50;
  font-family: 'Courier New', Consolas, monospace;
  font-size: 13px;
  line-height: 1.5;
  overflow-x: auto;
  white-space: pre;
}

/* ====================  批量删除模态框 ==================== */
.batch-delete-modal {
  max-width: 600px;
}

.batch-delete-list {
  max-height: 300px;
  overflow-y: auto;
  background: #f8f9fa;
  border: 1px solid #dee2e6;
  border-radius: 4px;
  padding: 12px;
  margin-bottom: 16px;
}

.batch-delete-item {
  padding: 8px;
  border-bottom: 1px solid #e9ecef;
  font-family: monospace;
  font-size: 13px;
}

.batch-delete-item:last-child {
  border-bottom: none;
}

.confirm-input-group {
  margin-top: 16px;
}

.confirm-input-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: #495057;
}

.confirm-input-group code {
  background: #f8d7da;
  color: #721c24;
  padding: 2px 6px;
  border-radius: 3px;
  font-weight: 600;
}

/* ====================  Job 列表模态框 ==================== */
.jobs-list-modal {
  max-width: 1000px;
  width: 95%;
}

.jobs-list-modal .modal-body {
  max-height: 600px;
  overflow-y: auto;
}

.jobs-list-modal .detail-table {
  width: 100%;
  margin: 0;
}

.jobs-list-modal .status-badge.success {
  background: #d4edda;
  color: #155724;
}

.jobs-list-modal .status-badge.danger {
  background: #f8d7da;
  color: #721c24;
}

.jobs-list-modal .status-badge.running {
  background: #cce5ff;
  color: #004085;
}

.jobs-list-modal .status-badge.pending {
  background: #fff3cd;
  color: #856404;
}

/* ====================  命令配置样式 ==================== */
.form-section-header {
  margin: 24px 0 16px 0;
  padding-bottom: 8px;
  border-bottom: 2px solid #e9ecef;
}

.form-section-header h4 {
  margin: 0 0 4px 0;
  font-size: 15px;
  font-weight: 600;
  color: #2c3e50;
}

.section-desc {
  margin: 0;
  font-size: 13px;
  color: #6c757d;
}

.form-textarea {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-family: 'Courier New', Consolas, monospace;
  font-size: 13px;
  line-height: 1.5;
  resize: vertical;
  background: #f8f9fa;
}

.form-textarea:focus {
  outline: none;
  border-color: #326ce5;
  background: white;
}

.command-input-wrapper {
  position: relative;
}

.info-icon {
  cursor: help;
  color: #6c757d;
  font-size: 14px;
}

.command-templates {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 16px 0;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 6px;
  flex-wrap: wrap;
}

.template-label {
  font-size: 13px;
  font-weight: 500;
  color: #495057;
}

.template-btn {
  padding: 6px 12px;
  background: white;
  border: 1px solid #dee2e6;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.2s;
  white-space: nowrap;
}

.template-btn:hover {
  background: #326ce5;
  color: white;
  border-color: #326ce5;
  transform: translateY(-1px);
}

/* 视图切换按钮 */
.view-toggle-buttons {
  display: flex;
  gap: 8px;
  margin-right: auto;
  margin-left: 16px;
}

.view-toggle-btn {
  padding: 8px 16px;
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  color: #64748b;
  transition: all 0.2s;
}

.view-toggle-btn:hover {
  background: #e2e8f0;
  color: #475569;
}

.view-toggle-btn.active {
  background: #326ce5;
  color: white;
  border-color: #326ce5;
}

/* YAML 编辑器容器 */
.yaml-editor-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.yaml-editor-header {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.yaml-editor-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
}

.yaml-hint {
  font-size: 13px;
  color: #6b7280;
  margin: 0;
}

.load-template-btn,
.copy-yaml-btn,
.reset-yaml-btn {
  padding: 8px 16px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  transition: all 0.2s;
}

.load-template-btn {
  background: linear-gradient(135deg, #1e293b 0%, #0f172a 100%);
  color: white;
  border: none;
}

.load-template-btn:hover {
  background: linear-gradient(135deg, #0f172a 0%, #020617 100%);
  box-shadow: 0 4px 12px rgba(15, 23, 42, 0.5);
  transform: translateY(-1px);
}

.load-template-btn:active {
  background: #020617;
  transform: translateY(0);
}

.copy-yaml-btn {
  background: #10b981;
  color: white;
  border: none;
}

.copy-yaml-btn:hover {
  background: #059669;
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.4);
  transform: translateY(-1px);
}

.reset-yaml-btn {
  background: #f1f5f9;
  color: #64748b;
  border: 1px solid #e2e8f0;
}

.reset-yaml-btn:hover {
  background: #e2e8f0;
  color: #475569;
}

.yaml-editor {
  width: 100%;
  min-height: 350px;
  padding: 16px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
  line-height: 1.6;
  background: #1e293b;
  color: #e2e8f0;
  border: 1px solid #334155;
  border-radius: 8px;
  resize: vertical;
}

.yaml-editor:focus {
  outline: none;
  border-color: #326ce5;
  box-shadow: 0 0 0 3px rgba(50, 108, 229, 0.2);
}

.yaml-error {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 8px;
  color: #dc2626;
  font-size: 13px;
}

.error-icon {
  font-size: 16px;
}

.yaml-editor-footer {
  background: #f8fafc;
  border-radius: 8px;
  padding: 16px;
}

.yaml-tips {
  font-size: 13px;
  color: #475569;
}

.yaml-tips strong {
  display: block;
  margin-bottom: 8px;
  color: #1f2937;
}

.yaml-tips ul {
  margin: 0;
  padding-left: 20px;
}

.yaml-tips li {
  margin-bottom: 4px;
}

.yaml-tips code {
  background: #e2e8f0;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 12px;
}

/* 按钮图标 */
.btn-icon {
  margin-right: 6px;
}

/* Job 操作按钮 */
.job-action-buttons {
  display: flex;
  gap: 8px;
  justify-content: center;
}

.job-action-buttons .btn-icon {
  padding: 4px 8px;
  font-size: 16px;
  border: 1px solid #e5e7eb;
  border-radius: 4px;
  background: white;
  cursor: pointer;
  transition: all 0.2s;
  margin: 0;
}

.job-action-buttons .btn-icon:hover {
  background: #f3f4f6;
  border-color: #326ce5;
  transform: translateY(-1px);
}

/* 日志模态框 */
.logs-modal {
  max-width: 1200px;
  width: 95%;
  max-height: 90vh;
}

.logs-controls {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  padding: 16px;
  background: #f8fafc;
  border-radius: 8px;
  margin-bottom: 16px;
}

.control-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 150px;
}

.control-item label {
  font-size: 12px;
  font-weight: 500;
  color: #475569;
}

.form-select {
  padding: 6px 10px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 13px;
  background: white;
}

.follow-toggle {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
}

.streaming-indicator {
  color: #22c55e;
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}

.control-actions {
  display: flex;
  gap: 8px;
  align-items: flex-end;
}

.logs-content-wrapper {
  max-height: 500px;
  overflow-y: auto;
  background: #1e293b;
  border-radius: 8px;
  padding: 16px;
}

.logs-content {
  color: #e2e8f0;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
}

.error-box {
  padding: 12px 16px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 8px;
  color: #dc2626;
  font-size: 13px;
  margin-bottom: 16px;
}

.loading-state,
.empty-state {
  text-align: center;
  padding: 40px;
  color: #64748b;
}
</style>
