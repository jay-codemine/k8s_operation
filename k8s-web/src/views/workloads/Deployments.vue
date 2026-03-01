<template>
  <div class="resource-view">
    <div class="view-header">
      <h1>部署管理</h1>
      <p>Kubernetes集群中的部署列表</p>
    </div>
    <div class="action-bar">
      <div class="search-box">
        <input type="text" v-model="searchQuery" placeholder="搜索部署名称..." @input="onSearchInput" />
      </div>

      <div class="filter-buttons">
        <button class="btn btn-filter" :class="{ active: statusFilter === 'all' }" @click="setStatusFilter('all')">
          全部
        </button>
        <button class="btn btn-filter" :class="{ active: statusFilter === 'Running' }" @click="setStatusFilter('Running')">
          Running
        </button>
        <button class="btn btn-filter" :class="{ active: statusFilter === 'Updating' }" @click="setStatusFilter('Updating')">
          Updating
        </button>
        <button class="btn btn-filter" :class="{ active: statusFilter === 'Stopped' }" @click="setStatusFilter('Stopped')">
          Stopped
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
        <button class="btn btn-primary" @click="showCreateModal = true">创建部署</button>
        <button class="btn btn-secondary" @click="refreshList" :disabled="loading">
          {{ loading ? '加载中...' : '🔄 刷新' }}
        </button>
      </div>
    </div>

    <div v-if="errorMsg" class="error-box">{{ errorMsg }}</div>

    <!-- 批量操作浮动栏 -->
    <div v-if="batchMode && selectedDeployments.length > 0" class="batch-action-bar">
      <div class="batch-info">
        <span class="batch-count">已选择 {{ selectedDeployments.length }} 个 Deployment</span>
        <button class="batch-clear" @click="clearSelection">清空选择</button>
      </div>
      <div class="batch-actions">
        <button class="batch-btn" @click="openBatchRestartPreview" title="批量重启">
          🔄 批量重启
        </button>
        <button class="batch-btn" @click="openBatchScaleModal" title="批量扩缩容">
          📊 批量扩缩容
        </button>
        <button class="batch-btn danger" @click="openBatchDeletePreview" title="批量删除">
          🗑️ 批量删除
        </button>
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
            <th style="width: 200px;">副本数</th>
            <th style="min-width: 200px;">镜像</th>
            <th style="min-width: 180px;">选择器</th>
            <th style="width: 120px;">更新策略</th>
            <th style="width: 170px; white-space: nowrap;">创建时间</th>
            <th style="width: 100px;">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="deployment in paginatedDeployments" :key="deployment.name" :class="{ 'row-selected': isDeploymentSelected(deployment) }">
            <td v-if="batchMode">
              <input 
                type="checkbox" 
                :checked="isDeploymentSelected(deployment)" 
                @change="toggleDeploymentSelection(deployment)"
              />
            </td>
            <td>
              <span class="status-indicator" :class="deployment.status.toLowerCase()">
                {{ deployment.status }}
              </span>
            </td>
            <td>
              <div class="deployment-name">
                <span class="icon">🚀</span>
                <span>{{ deployment.name }}</span>
              </div>
            </td>
            <td>
              <span class="namespace-badge">{{ deployment.namespace }}</span>
            </td>
            <td>
              <div class="replicas-info">
                <!-- 副本数 +/- 控制 -->
                <div class="replicas-control">
                  <button 
                    class="replica-btn minus" 
                    @click="decreaseReplicas(deployment)"
                    :disabled="deployment.desiredReplicas <= 0 || scalingMap[deployment.name]"
                    title="减少副本"
                  >−</button>
                  
                  <!-- 内联编辑副本数 -->
                  <div v-if="inlineEdit.key === `replicas-${deployment.name}`" class="replicas-edit-wrapper">
                    <input 
                      type="number" 
                      v-model="inlineEdit.value" 
                      class="replicas-input"
                      min="0"
                      @blur="saveInlineReplicas(deployment)"
                      @keyup.enter="saveInlineReplicas(deployment)"
                      @keyup.escape="cancelInlineEdit"
                      placeholder="副本数"
                      autofocus
                    />
                    <span class="inline-hint-small">回车</span>
                  </div>
                  
                  <!-- 显示副本数（可点击编辑） -->
                  <div v-else class="replicas-display clickable" 
                       :class="{ updating: scalingMap[deployment.name] }"
                       @click="startInlineReplicas(deployment)"
                       title="点击修改副本数">
                    <span class="ready-replicas">{{ deployment.readyReplicas }}</span>
                    <span class="replicas-sep">/</span>
                    <span class="desired-replicas">{{ deployment.desiredReplicas }}</span>
                    <span v-if="scalingMap[deployment.name]" class="scaling-indicator">⏳</span>
                    <span class="edit-icon-small">✏️</span>
                  </div>
                  
                  <button 
                    class="replica-btn plus" 
                    @click="increaseReplicas(deployment)"
                    :disabled="scalingMap[deployment.name]"
                    title="增加副本"
                  >+</button>
                  <!-- 停服按钮 -->
                  <button 
                    class="replica-btn stop" 
                    @click="stopService(deployment)"
                    :disabled="deployment.desiredReplicas === 0 || scalingMap[deployment.name]"
                    title="停服（副本数调为0）"
                  >⏸️</button>
                </div>
                <div class="replicas-bar">
                  <div class="replicas-fill" :style="{ width: `${(deployment.readyReplicas / Math.max(deployment.desiredReplicas, 1)) * 100}%` }"></div>
                </div>
              </div>
            </td>
            <td>
              <!-- 内联编辑镜像 -->
              <div class="inline-edit-wrapper" v-if="inlineEdit.key === `image-${deployment.name}`">
                <input 
                  type="text" 
                  v-model="inlineEdit.value" 
                  class="inline-input"
                  @blur="saveInlineImage(deployment)"
                  @keyup.enter="saveInlineImage(deployment)"
                  @keyup.escape="cancelInlineEdit"
                  placeholder="输入新镜像地址"
                  autofocus
                />
                <span class="inline-hint">按 Enter 保存</span>
              </div>
              <div v-else class="image-text clickable" @click="startInlineImage(deployment)" title="点击修改镜像">
                <span class="image-name">{{ deployment.image || '-' }}</span>
                <span class="edit-icon">✏️</span>
              </div>
            </td>
            <td>
              <div class="selector-tags">
                <span v-for="(value, key) in deployment.selector" :key="key" class="selector-tag">
                  {{ key }}={{ value }}
                </span>
              </div>
            </td>
            <td>
              <span class="strategy-badge">{{ deployment.updateStrategy }}</span>
            </td>
            <td style="white-space: nowrap;">{{ deployment.createdAt }}</td>
            <td>
              <div class="action-icons">
                <!-- Pod 关联按钮（独立显示） -->
                <button class="action-btn primary" @click="viewPods(deployment)" title="查看此 Deployment 管理的所有 Pod">
                  📦 Pod 关联
                </button>
                
                <!-- 更多按钮 -->
                <div class="more-btn">
                  <button class="icon-btn" @click="toggleMoreOptions(deployment, $event)">
                    ⋮ 更多
                  </button>

                  <!-- 更多菜单 -->
                  <div v-if="showMoreOptions && selectedDeployment === deployment" class="more-menu" :style="menuStyle">
                    <button class="menu-item" @click="viewDeploymentLogs(deployment)">
                      <span class="menu-icon">📄</span>
                      <span>查看日志</span>
                    </button>
                    <button class="menu-item" @click="viewHistory(deployment)">
                      <span class="menu-icon">📜</span>
                      <span>版本记录</span>
                    </button>
                    <div class="menu-divider"></div>
                    <button class="menu-item" @click="restartDeployment(deployment)">
                      <span class="menu-icon">🔄</span>
                      <span>重启</span>
                    </button>
                    <button class="menu-item" @click="openUpdateImage(deployment)">
                      <span class="menu-icon">🔧</span>
                      <span>更新镜像</span>
                    </button>
                    <button class="menu-item" @click="openRollback(deployment)">
                      <span class="menu-icon">⏪</span>
                      <span>回滚</span>
                    </button>
                    <div class="menu-divider"></div>
                    <button class="menu-item" @click="viewDeployment(deployment)">
                      <span class="menu-icon">📋</span>
                      <span>查看详情</span>
                    </button>
                    <button class="menu-item" @click="viewRelatedServices(deployment)">
                      <span class="menu-icon">🔌</span>
                      <span>关联 Service</span>
                    </button>
                    <button class="menu-item" @click="openEvents(deployment)">
                      <span class="menu-icon">📡</span>
                      <span>查看事件</span>
                    </button>
                    <button class="menu-item" @click="openYamlPreview(deployment)">
                      <span class="menu-icon">📝</span>
                      <span>查看/编辑 YAML</span>
                    </button>
                    <div class="menu-divider"></div>
                    <button class="menu-item danger" @click="deleteDeployment(deployment)">
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
      <div v-if="filteredDeployments.length === 0" class="empty-state">
        <div class="empty-icon">🔍</div>
        <div class="empty-text">没有找到匹配的部署</div>
      </div>
      <Pagination v-if="filteredDeployments.length > 0" v-model:currentPage="currentPage" :totalItems="filteredDeployments.length" :itemsPerPage="itemsPerPage" />
    </div>

    <!-- 卡片视图 -->
    <div v-if="viewMode === 'card'" class="cards-container">
      <div v-if="filteredDeployments.length === 0" class="empty-state">
        <div class="empty-icon">🔍</div>
        <div class="empty-text">没有找到匹配的部署</div>
      </div>
      
      <div class="cards-grid">
        <div v-for="deployment in paginatedDeployments" :key="deployment.name" class="deployment-card" :class="{ 'card-selected': isDeploymentSelected(deployment) }">
          <!-- 批量选择复选框 -->
          <div v-if="batchMode" class="card-checkbox">
            <input 
              type="checkbox" 
              :checked="isDeploymentSelected(deployment)" 
              @change="toggleDeploymentSelection(deployment)"
            />
          </div>
          
          <!-- 卡片头部 -->
          <div class="card-header">
            <div class="card-title-row">
              <span class="card-icon">🚀</span>
              <h3 class="card-title">{{ deployment.name }}</h3>
              <span class="status-indicator" :class="deployment.status.toLowerCase()">
                {{ deployment.status }}
              </span>
            </div>
            <span class="namespace-badge">{{ deployment.namespace }}</span>
          </div>

          <!-- 卡片主体 -->
          <div class="card-body">
            <!-- 副本数控制 -->
            <div class="card-section">
              <div class="section-label">副本数</div>
              <div class="replicas-info">
                <div class="replicas-control">
                  <button 
                    class="replica-btn minus" 
                    @click="decreaseReplicas(deployment)"
                    :disabled="deployment.desiredReplicas <= 0 || scalingMap[deployment.name]"
                    title="减少副本"
                  >−</button>
                  
                  <div v-if="inlineEdit.key === `replicas-${deployment.name}`" class="replicas-edit-wrapper">
                    <input 
                      type="number" 
                      v-model="inlineEdit.value" 
                      class="replicas-input"
                      min="0"
                      @blur="saveInlineReplicas(deployment)"
                      @keyup.enter="saveInlineReplicas(deployment)"
                      @keyup.escape="cancelInlineEdit"
                      placeholder="副本数"
                      autofocus
                    />
                    <span class="inline-hint-small">回车</span>
                  </div>
                  
                  <div v-else class="replicas-display clickable" 
                       :class="{ updating: scalingMap[deployment.name] }"
                       @click="startInlineReplicas(deployment)"
                       title="点击修改副本数">
                    <span class="ready-replicas">{{ deployment.readyReplicas }}</span>
                    <span class="replicas-sep">/</span>
                    <span class="desired-replicas">{{ deployment.desiredReplicas }}</span>
                    <span v-if="scalingMap[deployment.name]" class="scaling-indicator">⏳</span>
                    <span class="edit-icon-small">✏️</span>
                  </div>
                  
                  <button 
                    class="replica-btn plus" 
                    @click="increaseReplicas(deployment)"
                    :disabled="scalingMap[deployment.name]"
                    title="增加副本"
                  >+</button>
                  
                  <button 
                    class="replica-btn stop" 
                    @click="stopService(deployment)"
                    :disabled="deployment.desiredReplicas === 0 || scalingMap[deployment.name]"
                    title="停服（副本数调为0）"
                  >⏸️</button>
                </div>
                <div class="replicas-bar">
                  <div class="replicas-fill" :style="{ width: `${(deployment.readyReplicas / Math.max(deployment.desiredReplicas, 1)) * 100}%` }"></div>
                </div>
              </div>
            </div>

            <!-- 镜像 -->
            <div class="card-section">
              <div class="section-label">镜像</div>
              <div class="inline-edit-wrapper" v-if="inlineEdit.key === `image-${deployment.name}`">
                <input 
                  type="text" 
                  v-model="inlineEdit.value" 
                  class="inline-input"
                  @blur="saveInlineImage(deployment)"
                  @keyup.enter="saveInlineImage(deployment)"
                  @keyup.escape="cancelInlineEdit"
                  placeholder="输入新镜像地址"
                  autofocus
                />
                <span class="inline-hint">按 Enter 保存</span>
              </div>
              <div v-else class="image-text clickable" @click="startInlineImage(deployment)" title="点击修改镜像">
                <span class="image-name">{{ deployment.image || '-' }}</span>
                <span class="edit-icon">✏️</span>
              </div>
            </div>

            <!-- 选择器 -->
            <div class="card-section">
              <div class="section-label">选择器</div>
              <div class="selector-tags">
                <span v-for="(value, key) in deployment.selector" :key="key" class="selector-tag">
                  {{ key }}={{ value }}
                </span>
              </div>
            </div>

            <!-- 更新策略 & 创建时间 -->
            <div class="card-section card-section-row">
              <div class="card-meta-item">
                <div class="meta-label">更新策略</div>
                <span class="strategy-badge">{{ deployment.updateStrategy }}</span>
              </div>
              <div class="card-meta-item">
                <div class="meta-label">创建时间</div>
                <div class="meta-value">{{ deployment.createdAt }}</div>
              </div>
            </div>

            <!-- 资源消耗（Pod 聚合） -->
            <div class="card-section">
              <div class="section-label">
                资源使用
                <span v-if="deployment.metrics">（{{ deployment.metrics.podCount }} 个 Pod）</span>
              </div>
              <!-- 有 metrics 数据时显示具体值 -->
              <div v-if="deployment.metrics" class="metrics-summary">
                <div class="metric-item">
                  <span class="metric-icon">⚡</span>
                  <span class="metric-label">CPU:</span>
                  <span class="metric-value">{{ deployment.metrics.totalCpu }}</span>
                </div>
                <div class="metric-item">
                  <span class="metric-icon">💾</span>
                  <span class="metric-label">内存:</span>
                  <span class="metric-value">{{ deployment.metrics.totalMemory }}</span>
                </div>
              </div>
              <!-- 无 metrics 数据时显示原因提示 -->
              <div v-else class="metrics-summary metrics-unavailable">
                <div class="metric-item">
                  <span class="metric-icon">⚡</span>
                  <span class="metric-label">CPU:</span>
                  <span class="metric-value muted">-</span>
                </div>
                <div class="metric-item">
                  <span class="metric-icon">💾</span>
                  <span class="metric-label">内存:</span>
                  <span class="metric-value muted">-</span>
                </div>
                <div class="metrics-hint">
                  <span v-if="deployment.status === 'Stopped' || deployment.desiredReplicas === 0">已停服</span>
                  <span v-else-if="deployment.readyReplicas === 0">无就绪 Pod</span>
                  <span v-else>暂无数据</span>
                </div>
              </div>
            </div>
          </div>

          <!-- 卡片操作按钮 -->
          <div class="card-footer">
            <button class="card-action-btn primary" @click="viewPods(deployment)" title="查看此 Deployment 管理的所有 Pod">
              📦 Pod 关联
            </button>
            <button class="card-action-btn primary" @click="viewRelatedServices(deployment)" title="查看关联的 Service">
              🔌 Service
            </button>
            <button class="card-action-btn" @click="viewDeploymentLogs(deployment)" title="查看日志">
              📄 日志
            </button>
            <button class="card-action-btn" @click="viewHistory(deployment)" title="版本记录">
              📜 版本
            </button>
            <button class="card-action-btn" @click="openYamlPreview(deployment)" title="查看/编辑 YAML">
              📝 YAML
            </button>
            <button class="card-action-btn danger" @click="deleteDeployment(deployment)" title="删除">
              🗑️ 删除
            </button>
          </div>
        </div>
      </div>
      
      <Pagination v-if="filteredDeployments.length > 0" v-model:currentPage="currentPage" :totalItems="filteredDeployments.length" :itemsPerPage="itemsPerPage" />
    </div>

    <!-- 创建部署模态框（Rancher 风格 - 增强版） -->
    <div v-if="showCreateModal" class="modal-overlay" @click.self="showCreateModal = false">
      <div 
        ref="createModalRef"
        class="modal-content modal-create-deployment resizable-modal"
        :style="createModalStyle"
      >
        <!-- 8个拖拽手柄 -->
        <div class="resize-handle resize-handle-top" @mousedown="createStartResize($event, 'top')"></div>
        <div class="resize-handle resize-handle-bottom" @mousedown="createStartResize($event, 'bottom')"></div>
        <div class="resize-handle resize-handle-left" @mousedown="createStartResize($event, 'left')"></div>
        <div class="resize-handle resize-handle-right" @mousedown="createStartResize($event, 'right')"></div>
        <div class="resize-handle resize-handle-top-left" @mousedown="createStartResize($event, 'top-left')"></div>
        <div class="resize-handle resize-handle-top-right" @mousedown="createStartResize($event, 'top-right')"></div>
        <div class="resize-handle resize-handle-bottom-left" @mousedown="createStartResize($event, 'bottom-left')"></div>
        <div class="resize-handle resize-handle-bottom-right" @mousedown="createStartResize($event, 'bottom-right')"></div>
        
        <div class="modal-header">
          <h2>🚀 创建新部署</h2>
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
          <button class="close-btn" @click="showCreateModal = false">&times;</button>
        </div>
        <div class="modal-body">
          <!-- 表单模式 -->
          <div v-if="createMode === 'form'">
          <form @submit.prevent="createDeployment">
            <!-- 基本信息卡片 -->
            <div class="form-section">
              <div class="section-header">
                <span class="section-icon">📋</span>
                <h3>基本信息</h3>
              </div>
              <div class="section-body">
                <div class="form-group">
                  <label for="deploymentName">
                    部署名称 <span class="required">*</span>
                  </label>
                  <input 
                    type="text" 
                    id="deploymentName" 
                    v-model="deploymentForm.name" 
                    class="form-input"
                    required 
                    placeholder="输入部署名称，例如: my-app"
                  />
                  <div class="form-hint">用于标识此部署，建议使用小写字母和连字符</div>
                </div>
                <div class="form-group">
                  <label for="deploymentNamespace">
                    命名空间 <span class="required">*</span>
                  </label>
                  <div class="namespace-selector">
                    <select 
                      v-if="!showNamespaceInput" 
                      id="deploymentNamespace" 
                      v-model="deploymentForm.namespace" 
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
                </div>
                <div class="form-row">
                  <div class="form-group">
                    <label for="deploymentReplicas">
                      副本数 <span class="required">*</span>
                    </label>
                    <input 
                      type="number" 
                      id="deploymentReplicas" 
                      v-model="deploymentForm.replicas" 
                      class="form-input"
                      required 
                      min="0" 
                      placeholder="1"
                    />
                  </div>
                  <div class="form-group">
                    <label for="deploymentStrategy">更新策略</label>
                    <select id="deploymentStrategy" v-model="deploymentForm.updateStrategy" class="form-select" required>
                      <option value="RollingUpdate">滚动更新 (RollingUpdate)</option>
                      <option value="Recreate">重建 (Recreate)</option>
                    </select>
                  </div>
                </div>
              </div>
            </div>

            <!-- 容器镜像卡片 -->
            <div class="form-section">
              <div class="section-header">
                <span class="section-icon">🐳</span>
                <h3>容器镜像</h3>
              </div>
              <div class="section-body">
                <div class="form-group">
                  <label for="deploymentImage">
                    镜像地址 <span class="required">*</span>
                  </label>
                  <input 
                    type="text" 
                    id="deploymentImage" 
                    v-model="deploymentForm.image" 
                    class="form-input"
                    required 
                    placeholder="例如: nginx:latest 或 registry.example.com/app:v1.0.0" 
                  />
                  <div class="form-hint">支持 Docker Hub、私有仓库等镜像源</div>
                </div>
              </div>
            </div>

            <!-- 资源配置卡片 (Resources) -->
            <div class="form-section">
              <div class="section-header clickable" @click="toggleSection('resources')">
                <span class="section-icon">📊</span>
                <h3>资源限制 (Resources)</h3>
                <span class="toggle-indicator" :class="{ expanded: expandedSections.resources }">▼</span>
              </div>
              <div v-show="expandedSections.resources" class="section-body">
                <div class="form-hint mb-16">设置容器的 CPU 和内存资源限制，确保应用稳定运行</div>
                
                <!-- CPU 配置 -->
                <div class="resource-group">
                  <h4 class="resource-title">💻 CPU</h4>
                  <div class="resource-row">
                    <div class="form-group">
                      <label>CPU Request (请求)</label>
                      <div class="input-with-unit">
                        <input 
                          v-model="deploymentForm.resources.cpuRequest" 
                          type="text"
                          class="form-input"
                          placeholder="例如: 100m, 0.5, 1"
                        />
                        <span class="unit-hint">cores</span>
                      </div>
                      <div class="form-hint">CPU 最小保证值，100m = 0.1 cores</div>
                    </div>
                    <div class="form-group">
                      <label>CPU Limit (限制)</label>
                      <div class="input-with-unit">
                        <input 
                          v-model="deploymentForm.resources.cpuLimit" 
                          type="text"
                          class="form-input"
                          placeholder="例如: 200m, 1, 2"
                        />
                        <span class="unit-hint">cores</span>
                      </div>
                      <div class="form-hint">CPU 最大使用值，超过后会被限流</div>
                    </div>
                  </div>
                </div>

                <!-- 内存配置 -->
                <div class="resource-group">
                  <h4 class="resource-title">💾 内存</h4>
                  <div class="resource-row">
                    <div class="form-group">
                      <label>Memory Request (请求)</label>
                      <div class="input-with-unit">
                        <input 
                          v-model="deploymentForm.resources.memoryRequest" 
                          type="text"
                          class="form-input"
                          placeholder="例如: 64Mi, 128Mi, 1Gi"
                        />
                        <span class="unit-hint">bytes</span>
                      </div>
                      <div class="form-hint">内存最小保证值</div>
                    </div>
                    <div class="form-group">
                      <label>Memory Limit (限制)</label>
                      <div class="input-with-unit">
                        <input 
                          v-model="deploymentForm.resources.memoryLimit" 
                          type="text"
                          class="form-input"
                          placeholder="例如: 128Mi, 256Mi, 2Gi"
                        />
                        <span class="unit-hint">bytes</span>
                      </div>
                      <div class="form-hint">内存最大使用值，超过后 Pod 会被 OOM Kill</div>
                    </div>
                  </div>
                </div>

                <div class="resource-tips">
                  <span class="tip-icon">💡</span>
                  <strong>提示：</strong>
                  <ul>
                    <li>Request 决定 Pod 调度到节点的条件，Limit 防止资源过度使用</li>
                    <li>CPU 单位：1 = 1 core，100m = 0.1 core</li>
                    <li>内存单位：Mi (兆字节), Gi (吉字节)</li>
                  </ul>
                </div>
              </div>
            </div>

            <!-- 探针配置卡片 (Probes) -->
            <div class="form-section">
              <div class="section-header clickable" @click="toggleSection('probes')">
                <span class="section-icon">🔍</span>
                <h3>健康检查 (Health Probes)</h3>
                <span class="toggle-indicator" :class="{ expanded: expandedSections.probes }">▼</span>
              </div>
              <div v-show="expandedSections.probes" class="section-body">
                <div class="form-hint mb-16">配置容器的健康检查探针，Kubernetes 根据探针结果自动管理 Pod 生命周期</div>
                
                <!-- Liveness Probe -->
                <div class="probe-config">
                  <div class="probe-header">
                    <label class="toggle-switch">
                      <input type="checkbox" v-model="deploymentForm.probes.enableLiveness" />
                      <span class="toggle-slider"></span>
                    </label>
                    <h4 class="probe-title">❤️ Liveness Probe (存活探针)</h4>
                  </div>
                  <div class="form-hint">检测容器是否正常运行，失败后 K8s 会重启容器</div>
                  <div v-if="deploymentForm.probes.enableLiveness" class="probe-body">
                    <div class="form-group">
                      <label>探测类型</label>
                      <select v-model="deploymentForm.probes.livenessProbe.type" class="form-select">
                        <option value="HTTP">HTTP GET</option>
                        <option value="TCP">TCP Socket</option>
                        <option value="Command">Exec Command</option>
                      </select>
                    </div>
                    
                    <div v-if="deploymentForm.probes.livenessProbe.type === 'HTTP'" class="probe-http-config">
                      <div class="form-row">
                        <div class="form-group">
                          <label>端口</label>
                          <input v-model.number="deploymentForm.probes.livenessProbe.port" type="number" class="form-input" placeholder="8080" />
                        </div>
                        <div class="form-group">
                          <label>路径</label>
                          <input v-model="deploymentForm.probes.livenessProbe.path" type="text" class="form-input" placeholder="/healthz" />
                        </div>
                      </div>
                    </div>
                    
                    <div v-if="deploymentForm.probes.livenessProbe.type === 'TCP'" class="probe-tcp-config">
                      <div class="form-group">
                        <label>端口</label>
                        <input v-model.number="deploymentForm.probes.livenessProbe.port" type="number" class="form-input" placeholder="8080" />
                      </div>
                    </div>
                    
                    <div v-if="deploymentForm.probes.livenessProbe.type === 'Command'" class="probe-command-config">
                      <div class="form-group">
                        <label>执行命令</label>
                        <input v-model="deploymentForm.probes.livenessProbe.command" type="text" class="form-input" placeholder="cat /tmp/healthy" />
                      </div>
                    </div>
                    
                    <div class="probe-timing-config">
                      <div class="form-row">
                        <div class="form-group">
                          <label>初始延迟(秒)</label>
                          <input v-model.number="deploymentForm.probes.livenessProbe.initialDelaySeconds" type="number" class="form-input" placeholder="10" />
                        </div>
                        <div class="form-group">
                          <label>检测间隔(秒)</label>
                          <input v-model.number="deploymentForm.probes.livenessProbe.periodSeconds" type="number" class="form-input" placeholder="10" />
                        </div>
                        <div class="form-group">
                          <label>超时时间(秒)</label>
                          <input v-model.number="deploymentForm.probes.livenessProbe.timeoutSeconds" type="number" class="form-input" placeholder="1" />
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- Readiness Probe -->
                <div class="probe-config">
                  <div class="probe-header">
                    <label class="toggle-switch">
                      <input type="checkbox" v-model="deploymentForm.probes.enableReadiness" />
                      <span class="toggle-slider"></span>
                    </label>
                    <h4 class="probe-title">✅ Readiness Probe (就绪探针)</h4>
                  </div>
                  <div class="form-hint">检测容器是否准备好接收流量，失败后 Service 不会路由流量到此 Pod</div>
                  <div v-if="deploymentForm.probes.enableReadiness" class="probe-body">
                    <div class="form-group">
                      <label>探测类型</label>
                      <select v-model="deploymentForm.probes.readinessProbe.type" class="form-select">
                        <option value="HTTP">HTTP GET</option>
                        <option value="TCP">TCP Socket</option>
                        <option value="Command">Exec Command</option>
                      </select>
                    </div>
                    
                    <div v-if="deploymentForm.probes.readinessProbe.type === 'HTTP'" class="probe-http-config">
                      <div class="form-row">
                        <div class="form-group">
                          <label>端口</label>
                          <input v-model.number="deploymentForm.probes.readinessProbe.port" type="number" class="form-input" placeholder="8080" />
                        </div>
                        <div class="form-group">
                          <label>路径</label>
                          <input v-model="deploymentForm.probes.readinessProbe.path" type="text" class="form-input" placeholder="/ready" />
                        </div>
                      </div>
                    </div>
                    
                    <div v-if="deploymentForm.probes.readinessProbe.type === 'TCP'" class="probe-tcp-config">
                      <div class="form-group">
                        <label>端口</label>
                        <input v-model.number="deploymentForm.probes.readinessProbe.port" type="number" class="form-input" placeholder="8080" />
                      </div>
                    </div>
                    
                    <div v-if="deploymentForm.probes.readinessProbe.type === 'Command'" class="probe-command-config">
                      <div class="form-group">
                        <label>执行命令</label>
                        <input v-model="deploymentForm.probes.readinessProbe.command" type="text" class="form-input" placeholder="cat /tmp/ready" />
                      </div>
                    </div>
                    
                    <div class="probe-timing-config">
                      <div class="form-row">
                        <div class="form-group">
                          <label>初始延迟(秒)</label>
                          <input v-model.number="deploymentForm.probes.readinessProbe.initialDelaySeconds" type="number" class="form-input" placeholder="5" />
                        </div>
                        <div class="form-group">
                          <label>检测间隔(秒)</label>
                          <input v-model.number="deploymentForm.probes.readinessProbe.periodSeconds" type="number" class="form-input" placeholder="10" />
                        </div>
                        <div class="form-group">
                          <label>超时时间(秒)</label>
                          <input v-model.number="deploymentForm.probes.readinessProbe.timeoutSeconds" type="number" class="form-input" placeholder="1" />
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- Startup Probe -->
                <div class="probe-config">
                  <div class="probe-header">
                    <label class="toggle-switch">
                      <input type="checkbox" v-model="deploymentForm.probes.enableStartup" />
                      <span class="toggle-slider"></span>
                    </label>
                    <h4 class="probe-title">🚀 Startup Probe (启动探针)</h4>
                  </div>
                  <div class="form-hint">检测容器是否已启动，在此期间 Liveness/Readiness 不会执行（适用于启动慢的容器）</div>
                  <div v-if="deploymentForm.probes.enableStartup" class="probe-body">
                    <div class="form-group">
                      <label>探测类型</label>
                      <select v-model="deploymentForm.probes.startupProbe.type" class="form-select">
                        <option value="HTTP">HTTP GET</option>
                        <option value="TCP">TCP Socket</option>
                        <option value="Command">Exec Command</option>
                      </select>
                    </div>
                    
                    <div v-if="deploymentForm.probes.startupProbe.type === 'HTTP'" class="probe-http-config">
                      <div class="form-row">
                        <div class="form-group">
                          <label>端口</label>
                          <input v-model.number="deploymentForm.probes.startupProbe.port" type="number" class="form-input" placeholder="8080" />
                        </div>
                        <div class="form-group">
                          <label>路径</label>
                          <input v-model="deploymentForm.probes.startupProbe.path" type="text" class="form-input" placeholder="/startup" />
                        </div>
                      </div>
                    </div>
                    
                    <div v-if="deploymentForm.probes.startupProbe.type === 'TCP'" class="probe-tcp-config">
                      <div class="form-group">
                        <label>端口</label>
                        <input v-model.number="deploymentForm.probes.startupProbe.port" type="number" class="form-input" placeholder="8080" />
                      </div>
                    </div>
                    
                    <div v-if="deploymentForm.probes.startupProbe.type === 'Command'" class="probe-command-config">
                      <div class="form-group">
                        <label>执行命令</label>
                        <input v-model="deploymentForm.probes.startupProbe.command" type="text" class="form-input" placeholder="cat /tmp/started" />
                      </div>
                    </div>
                    
                    <div class="probe-timing-config">
                      <div class="form-row">
                        <div class="form-group">
                          <label>初始延迟(秒)</label>
                          <input v-model.number="deploymentForm.probes.startupProbe.initialDelaySeconds" type="number" class="form-input" placeholder="0" />
                        </div>
                        <div class="form-group">
                          <label>检测间隔(秒)</label>
                          <input v-model.number="deploymentForm.probes.startupProbe.periodSeconds" type="number" class="form-input" placeholder="10" />
                        </div>
                        <div class="form-group">
                          <label>失败阈值</label>
                          <input v-model.number="deploymentForm.probes.startupProbe.failureThreshold" type="number" class="form-input" placeholder="30" />
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- 标签选择器卡片 -->
            <div class="form-section">
              <div class="section-header">
                <span class="section-icon">🏷️</span>
                <h3>标签与选择器</h3>
              </div>
              <div class="section-body">
                <div class="form-group">
                  <label>
                    Pod 标签 (Labels) <span class="required">*</span>
                  </label>
                  <div class="labels-editor">
                    <div v-for="(label, index) in deploymentForm.labels" :key="index" class="label-row">
                      <input 
                        v-model="label.key" 
                        class="label-input label-key"
                        placeholder="键，例如: app"
                      />
                      <span class="label-separator">=</span>
                      <input 
                        v-model="label.value" 
                        class="label-input label-value"
                        placeholder="值，例如: nginx"
                      />
                      <button 
                        type="button" 
                        class="btn-icon btn-remove" 
                        @click="removeLabel(index)"
                        :disabled="deploymentForm.labels.length === 1"
                      >
                        🗑️
                      </button>
                    </div>
                    <button type="button" class="btn btn-secondary btn-sm" @click="addLabel">
                      ➕ 添加标签
                    </button>
                  </div>
                  <div class="form-hint">标签用于识别和选择 Pod，至少需要一个标签</div>
                </div>
              </div>
            </div>
            
            <!-- Service 配置卡片 -->
            <div class="form-section">
              <div class="section-header">
                <span class="section-icon">🌐</span>
                <h3>服务暴露 (Service)</h3>
                <label class="toggle-switch">
                  <input type="checkbox" v-model="deploymentForm.createService" />
                  <span class="toggle-slider"></span>
                </label>
              </div>
              
              <!-- Service 配置（当勾选时显示） -->
              <div v-if="deploymentForm.createService" class="section-body service-config">
                <div class="form-group">
                  <label for="serviceName">Service 名称</label>
                  <input 
                    type="text" 
                    id="serviceName" 
                    v-model="deploymentForm.serviceName" 
                    class="form-input"
                    placeholder="留空则自动使用 Deployment 名称"
                  />
                </div>
                
                <div class="form-group">
                  <label for="serviceType">
                    服务类型 <span class="required">*</span>
                  </label>
                  <div class="service-type-selector">
                    <label 
                      v-for="type in serviceTypes" 
                      :key="type.value" 
                      class="service-type-card"
                      :class="{ active: deploymentForm.serviceType === type.value }"
                    >
                      <input 
                        type="radio" 
                        :value="type.value" 
                        v-model="deploymentForm.serviceType"
                        style="display: none;"
                      />
                      <div class="type-icon">{{ type.icon }}</div>
                      <div class="type-name">{{ type.label }}</div>
                      <div class="type-desc">{{ type.description }}</div>
                    </label>
                  </div>
                </div>

                <!-- 端口配置 -->
                <div v-if="deploymentForm.serviceType !== 'None'" class="form-group">
                  <label>
                    端口映射 <span class="required">*</span>
                  </label>
                  <div class="port-mapping-editor">
                    <div v-for="(port, index) in deploymentForm.servicePorts" :key="index" class="port-row">
                      <div class="port-field">
                        <label>容器端口</label>
                        <input 
                          v-model.number="port.targetPort" 
                          type="number"
                          class="form-input"
                          placeholder="80"
                          min="1"
                          max="65535"
                        />
                      </div>
                      <span class="port-arrow">→</span>
                      <div class="port-field">
                        <label>服务端口</label>
                        <input 
                          v-model.number="port.port" 
                          type="number"
                          class="form-input"
                          placeholder="80"
                          min="1"
                          max="65535"
                        />
                      </div>
                      <div v-if="deploymentForm.serviceType === 'NodePort'" class="port-field">
                        <label>节点端口 (可选)</label>
                        <input 
                          v-model.number="port.nodePort" 
                          type="number"
                          class="form-input"
                          placeholder="30000-32767"
                          min="30000"
                          max="32767"
                        />
                      </div>
                      <button 
                        type="button" 
                        class="btn-icon btn-remove" 
                        @click="removeServicePort(index)"
                        :disabled="deploymentForm.servicePorts.length === 1"
                      >
                        🗑️
                      </button>
                    </div>
                    <button type="button" class="btn btn-secondary btn-sm" @click="addServicePort">
                      ➕ 添加端口
                    </button>
                  </div>
                  <div class="form-hint" v-if="deploymentForm.serviceType === 'NodePort'">
                    NodePort 范围: 30000-32767，留空则自动分配
                  </div>
                </div>
              </div>
            </div>

            <!-- 调度规则卡片 -->
            <div class="form-section">
              <div class="section-header">
                <span class="section-icon">📍</span>
                <h3>容器组调度规则</h3>
              </div>
              <div class="section-body">
                <div class="form-hint mb-12">设置容器组副本调度到节点的规则</div>
                
                <!-- 调度策略选择 -->
                <div class="scheduling-policy-selector">
                  <label 
                    v-for="policy in schedulingPolicies" 
                    :key="policy.value" 
                    class="scheduling-policy-card"
                    :class="{ active: deploymentForm.schedulingPolicy === policy.value }"
                  >
                    <input 
                      type="radio" 
                      :value="policy.value" 
                      v-model="deploymentForm.schedulingPolicy"
                      style="display: none;"
                    />
                    <div class="policy-title">{{ policy.label }}</div>
                    <div class="policy-desc">{{ policy.description }}</div>
                  </label>
                </div>

                <!-- 自定义规则配置 -->
                <div v-if="deploymentForm.schedulingPolicy === 'custom'" class="custom-scheduling-config">
                  <!-- 节点选择器 -->
                  <div class="form-group">
                    <label>节点选择器 (Node Selector)</label>
                    <div class="form-hint">指定 Pod 只能调度到带有特定标签的节点</div>
                    <div class="labels-editor">
                      <div v-for="(selector, index) in deploymentForm.nodeSelector" :key="index" class="label-row">
                        <input 
                          v-model="selector.key" 
                          class="label-input label-key"
                          placeholder="键，例如: kubernetes.io/hostname"
                        />
                        <span class="label-separator">=</span>
                        <input 
                          v-model="selector.value" 
                          class="label-input label-value"
                          placeholder="值，例如: node-1"
                        />
                        <button 
                          type="button" 
                          class="btn-icon btn-remove" 
                          @click="removeNodeSelector(index)"
                        >
                          🗑️
                        </button>
                      </div>
                      <button type="button" class="btn btn-secondary btn-sm" @click="addNodeSelector">
                        ➕ 添加选择器
                      </button>
                    </div>
                  </div>

                  <!-- 容忍配置 -->
                  <div class="form-group">
                    <label>容忍 (Tolerations)</label>
                    <div class="form-hint">允许 Pod 调度到带有特定污点的节点</div>
                    <div class="tolerations-editor">
                      <div v-for="(toleration, index) in deploymentForm.tolerations" :key="index" class="toleration-row">
                        <input 
                          v-model="toleration.key" 
                          class="toleration-input"
                          placeholder="污点键"
                        />
                        <select v-model="toleration.operator" class="toleration-select">
                          <option value="Equal">Equal</option>
                          <option value="Exists">Exists</option>
                        </select>
                        <input 
                          v-if="toleration.operator === 'Equal'"
                          v-model="toleration.value" 
                          class="toleration-input"
                          placeholder="污点值"
                        />
                        <select v-model="toleration.effect" class="toleration-select">
                          <option value="">所有效果</option>
                          <option value="NoSchedule">NoSchedule</option>
                          <option value="PreferNoSchedule">PreferNoSchedule</option>
                          <option value="NoExecute">NoExecute</option>
                        </select>
                        <input 
                          v-if="toleration.effect === 'NoExecute'"
                          v-model.number="toleration.tolerationSeconds" 
                          class="toleration-input toleration-seconds"
                          type="number"
                          placeholder="容忍时间(秒)"
                          min="0"
                        />
                        <button 
                          type="button" 
                          class="btn-icon btn-remove" 
                          @click="removeToleration(index)"
                        >
                          🗑️
                        </button>
                      </div>
                      <button type="button" class="btn btn-secondary btn-sm" @click="addToleration">
                        ➕ 添加容忍
                      </button>
                    </div>
                  </div>

                  <!-- 节点亲和性 -->
                  <div class="form-group">
                    <label>节点亲和性 (Node Affinity)</label>
                    <div class="form-hint">根据节点标签设置调度偏好或要求</div>
                    <div class="affinity-editor">
                      <div v-for="(rule, index) in deploymentForm.nodeAffinityRules" :key="index" class="affinity-row">
                        <input 
                          v-model="rule.key" 
                          class="affinity-input"
                          placeholder="节点标签键"
                        />
                        <select v-model="rule.operator" class="affinity-select">
                          <option value="In">In (包含)</option>
                          <option value="NotIn">NotIn (不包含)</option>
                          <option value="Exists">Exists (存在)</option>
                          <option value="DoesNotExist">DoesNotExist (不存在)</option>
                          <option value="Gt">Gt (大于)</option>
                          <option value="Lt">Lt (小于)</option>
                        </select>
                        <input 
                          v-if="rule.operator !== 'Exists' && rule.operator !== 'DoesNotExist'"
                          v-model="rule.values" 
                          class="affinity-input affinity-values"
                          placeholder="值(多个用逗号分隔)"
                        />
                        <label class="affinity-required-label">
                          <input type="checkbox" v-model="rule.required" />
                          <span>硬性要求</span>
                        </label>
                        <input 
                          v-if="!rule.required"
                          v-model.number="rule.weight" 
                          class="affinity-input affinity-weight"
                          type="number"
                          placeholder="权重(1-100)"
                          min="1"
                          max="100"
                        />
                        <button 
                          type="button" 
                          class="btn-icon btn-remove" 
                          @click="removeNodeAffinityRule(index)"
                        >
                          🗑️
                        </button>
                      </div>
                      <button type="button" class="btn btn-secondary btn-sm" @click="addNodeAffinityRule">
                        ➕ 添加节点亲和性规则
                      </button>
                    </div>
                  </div>

                  <!-- 拓扑分布约束 -->
                  <div class="form-group">
                    <label>拓扑分布约束 (Topology Spread)</label>
                    <div class="form-hint">跨可用区/节点均匀分布 Pod 副本</div>
                    <div class="topology-editor">
                      <div v-for="(spread, index) in deploymentForm.topologySpreadConfigs" :key="index" class="topology-row">
                        <select v-model="spread.topologyKey" class="topology-select">
                          <option value="kubernetes.io/hostname">按节点分布</option>
                          <option value="topology.kubernetes.io/zone">按可用区分布</option>
                          <option value="topology.kubernetes.io/region">按区域分布</option>
                        </select>
                        <div class="topology-field">
                          <label>最大偏差</label>
                          <input 
                            v-model.number="spread.maxSkew" 
                            type="number"
                            class="topology-input"
                            min="1"
                            placeholder="1"
                          />
                        </div>
                        <select v-model="spread.whenUnsatisfiable" class="topology-select">
                          <option value="ScheduleAnyway">软性约束 (ScheduleAnyway)</option>
                          <option value="DoNotSchedule">硬性约束 (DoNotSchedule)</option>
                        </select>
                        <button 
                          type="button" 
                          class="btn-icon btn-remove" 
                          @click="removeTopologySpread(index)"
                        >
                          🗑️
                        </button>
                      </div>
                      <button type="button" class="btn btn-secondary btn-sm" @click="addTopologySpread">
                        ➕ 添加拓扑分布约束
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </form>
          </div>

          <!-- YAML 模式 -->
          <div v-if="createMode === 'yaml'" class="yaml-editor-container">
            <div class="yaml-editor-header">
              <h3>📄 YAML 配置</h3>
              <p class="yaml-hint">✨ 支持多资源 YAML 创建（用 <code>---</code> 分隔），可同时创建 PVC、ConfigMap、Service 等依赖资源</p>
              <button class="load-template-btn" @click="loadDeploymentYamlTemplate">
                📑 加载模板（PVC + Deployment）
              </button>
              <button class="copy-yaml-btn" @click="copyYamlContent">
                📋 复制
              </button>
              <button class="reset-yaml-btn" @click="resetYamlContent">
                🔄 重置
              </button>
            </div>
            
            <textarea 
              v-model="yamlContent" 
              class="yaml-editor"
              placeholder="输入或粘贴 YAML 内容...&#10;&#10;支持多资源 YAML 创建示例：&#10;apiVersion: v1&#10;kind: PersistentVolumeClaim&#10;metadata:&#10;  name: my-pvc&#10;---&#10;apiVersion: apps/v1&#10;kind: Deployment&#10;metadata:&#10;  name: my-app&#10;..."
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
                  <li>✅ 支持单资源或多资源 YAML（用 <code>---</code> 分隔）</li>
                  <li>🔗 可同时创建：PVC + Deployment / ConfigMap + Deployment / Secret + Deployment</li>
                  <li>📦 资源创建顺序：PVC/ConfigMap/Secret → Service → Deployment</li>
                  <li>🚀 点击"加载模板"获取完整的多资源示例</li>
                </ul>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <!-- 表单模式按钮 -->
          <template v-if="createMode === 'form'">
            <button type="button" class="btn btn-secondary" @click="showCreateModal = false">取消</button>
            <button type="button" class="btn btn-primary" @click="createDeployment">
              <span class="btn-icon">🚀</span>
              创建部署
            </button>
          </template>
          
          <!-- YAML 模式按钮 -->
          <template v-if="createMode === 'yaml'">
            <button type="button" class="btn btn-secondary" @click="showCreateModal = false">取消</button>
            <button 
              type="button"
              class="btn btn-primary" 
              @click="createDeploymentFromYaml"
              :disabled="!yamlContent"
            >
              <span class="btn-icon">🚀</span>
              创建部署
            </button>
          </template>
        </div>
      </div>
    </div>

    <!-- 编辑部署模态框 -->
    <div v-if="showEditModal" class="modal-overlay" @click.self="showEditModal = false">
      <div 
        ref="editModalRef"
        class="modal-content resizable-modal"
        :style="editModalStyle"
      >
        <!-- 8个拖拽手柄 -->
        <div class="resize-handle resize-handle-top" @mousedown="editStartResize($event, 'top')"></div>
        <div class="resize-handle resize-handle-bottom" @mousedown="editStartResize($event, 'bottom')"></div>
        <div class="resize-handle resize-handle-left" @mousedown="editStartResize($event, 'left')"></div>
        <div class="resize-handle resize-handle-right" @mousedown="editStartResize($event, 'right')"></div>
        <div class="resize-handle resize-handle-top-left" @mousedown="editStartResize($event, 'top-left')"></div>
        <div class="resize-handle resize-handle-top-right" @mousedown="editStartResize($event, 'top-right')"></div>
        <div class="resize-handle resize-handle-bottom-left" @mousedown="editStartResize($event, 'bottom-left')"></div>
        <div class="resize-handle resize-handle-bottom-right" @mousedown="editStartResize($event, 'bottom-right')"></div>
        
        <div class="modal-header">
          <h2>编辑部署</h2>
          <button class="close-btn" @click="showEditModal = false">&times;</button>
        </div>
        <div class="modal-body">
          <form @submit.prevent="updateDeployment">
            <div class="form-group">
              <label for="editDeploymentName">部署名称</label>
              <input type="text" id="editDeploymentName" v-model="editForm.name" required />
            </div>
            <div class="form-group">
              <label for="editDeploymentNamespace">命名空间</label>
              <select id="editDeploymentNamespace" v-model="editForm.namespace" required disabled>
                <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
              </select>
            </div>
            <div class="form-group">
              <label for="editDeploymentReplicas">副本数</label>
              <input type="number" id="editDeploymentReplicas" v-model="editForm.replicas" required min="1" />
            </div>
            <div class="form-group">
              <label for="editDeploymentImage">镜像</label>
              <input type="text" id="editDeploymentImage" v-model="editForm.image" required />
            </div>
            <div class="form-group">
              <label for="editDeploymentSelector">选择器 (key=value)</label>
              <input type="text" id="editDeploymentSelector" v-model="editSelectorInput" required />
            </div>
            <div class="form-group">
              <label for="editDeploymentStrategy">更新策略</label>
              <select id="editDeploymentStrategy" v-model="editForm.updateStrategy" required>
                <option value="RollingUpdate">RollingUpdate</option>
                <option value="Recreate">Recreate</option>
              </select>
            </div>
          </form>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showEditModal = false">取消</button>
          <button class="btn btn-primary" @click="updateDeployment">保存</button>
        </div>
      </div>
    </div>

    <!-- 删除确认模态框 -->
    <div v-if="showDeleteModal" class="modal-overlay" @click.self="showDeleteModal = false">
      <div class="modal-content">
        <div class="modal-header">
          <h2>确认删除</h2>
          <button class="close-btn" @click="showDeleteModal = false">&times;</button>
        </div>
        <div class="modal-body">
          <p>您确定要删除部署 <strong>{{ deploymentToDelete }}</strong> 吗？此操作无法撤销。</p>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showDeleteModal = false">取消</button>
          <button class="btn btn-danger" @click="confirmDelete">删除</button>
        </div>
      </div>
    </div>

    <!-- 查看部署详情模态框 -->
    <div v-if="showViewModal" class="modal-overlay" @click.self="showViewModal = false">
      <div class="modal-content">
        <div class="modal-header">
          <h3>📋 Deployment 详情</h3>
          <button class="close-btn" @click="showViewModal = false">×</button>
        </div>
        <div class="modal-body">
          <div v-if="loadingDetail" class="loading-state">加载中...</div>
          <div v-else-if="detailData">
            <pre class="detail-json">{{ JSON.stringify(detailData, null, 2) }}</pre>
          </div>
        </div>
      </div>
    </div>

    <!-- 扩缩容模态框 -->
    <div v-if="showScaleModal" class="modal-overlay" @click.self="showScaleModal = false">
      <div class="modal-content" style="max-width: 480px;">
        <div class="modal-header">
          <h3>🔢 扩缩容</h3>
          <button class="close-btn" @click="showScaleModal = false">×</button>
        </div>
        <div class="modal-body">
          <div class="info-box">
            <div><strong>部署:</strong> {{ scaleForm.name }}</div>
            <div><strong>命名空间:</strong> {{ scaleForm.namespace }}</div>
          </div>
          <div class="form-group">
            <label>目标副本数</label>
            <input type="number" v-model="scaleForm.replicas" min="0" class="form-input" />
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showScaleModal = false">取消</button>
          <button class="btn btn-primary" @click="scaleDeployment" :disabled="scaling">
            {{ scaling ? '处理中...' : '确认' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 事件弹窗 -->
    <div v-if="showEventsModal" class="modal-overlay" @click.self="showEventsModal = false">
      <div class="modal-content">
        <div class="modal-header">
          <h3>📡 Deployment 事件</h3>
          <button class="close-btn" @click="showEventsModal = false">×</button>
        </div>
        <div class="modal-body">
          <div v-if="loadingEvents" class="loading-state">加载中...</div>
          <div v-else-if="eventsData.length > 0">
            <div v-for="(event, idx) in eventsData" :key="idx" class="event-item">
              <div class="event-type" :class="(event.type || 'normal').toLowerCase()">
                {{ event.type || 'Normal' }}
              </div>
              <div class="event-content">
                <div class="event-reason">{{ event.reason }}</div>
                <div class="event-message">{{ event.message }}</div>
                <div class="event-time">
                  {{ fmtTime(event.event_time || event.lastTimestamp) }}
                  <span v-if="event.count > 1" class="event-count">(发生 {{ event.count }} 次)</span>
                </div>
              </div>
            </div>
          </div>
          <div v-else class="empty-state-small">
            <div class="empty-icon">📬</div>
            <div class="empty-text">暂无事件记录</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 更新镜像弹窗 -->
    <div v-if="showUpdateImageModal" class="modal-overlay" @click.self="showUpdateImageModal = false">
      <div class="modal-content" style="max-width: 520px;">
        <div class="modal-header">
          <h3>🔧 更新容器镜像</h3>
          <button class="close-btn" @click="showUpdateImageModal = false">×</button>
        </div>
        <div class="modal-body">
          <div class="info-box">
            <div><strong>部署:</strong> {{ updateImageForm.name }}</div>
            <div><strong>命名空间:</strong> {{ updateImageForm.namespace }}</div>
          </div>
          <div class="form-group" v-if="containerList.length > 1">
            <label>选择容器</label>
            <select v-model="updateImageForm.container" class="form-select">
              <option value="" disabled>请选择容器</option>
              <option v-for="c in containerList" :key="c" :value="c">{{ c }}</option>
            </select>
          </div>
          <div class="form-group" v-else-if="containerList.length === 1">
            <label>容器</label>
            <div class="form-static">{{ containerList[0] }}</div>
          </div>
          <div class="form-group">
            <label>新镜像地址</label>
            <input v-model="updateImageForm.image" type="text" class="form-input" placeholder="例如: nginx:1.28" />
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showUpdateImageModal = false">取消</button>
          <button class="btn btn-primary" @click="submitUpdateImage" :disabled="updatingImage || !updateImageForm.image">
            {{ updatingImage ? '更新中...' : '确认更新' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 版本记录弹窗 -->
    <div v-if="showHistoryModal" class="modal-overlay" @click.self="showHistoryModal = false">
      <div class="modal-content modal-lg">
        <div class="modal-header">
          <h3>📜 版本记录</h3>
          <button class="close-btn" @click="showHistoryModal = false">×</button>
        </div>
        <div class="modal-body">
          <div v-if="historyDeployment" class="info-box" style="margin-bottom: 16px;">
            <div><strong>部署:</strong> {{ historyDeployment.name }}</div>
            <div><strong>命名空间:</strong> {{ historyDeployment.namespace }}</div>
            <div><strong>当前版本:</strong> {{ historyDeployment.revision || '-' }}</div>
          </div>
          
          <div v-if="loadingHistory" class="loading-state">加载版本历史...</div>
          <div v-else-if="historyList.length > 0">
            <table class="simple-table">
              <thead>
                <tr>
                  <th style="width: 100px;">版本号</th>
                  <th>ReplicaSet 名称</th>
                  <th style="width: 100px;">副本数</th>
                  <th style="width: 180px;">创建时间</th>
                  <th style="width: 150px;">操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="h in historyList" :key="h.name">
                  <td>
                    <span class="version-badge">
                      {{ h.revision }}
                      <span v-if="h.revision === historyDeployment?.revision" class="current-tag">当前</span>
                    </span>
                  </td>
                  <td class="mono">{{ h.name }}</td>
                  <td style="text-align: center;">{{ h.replicas }}</td>
                  <td>{{ h.createdAt }}</td>
                  <td>
                    <button 
                      v-if="h.revision !== historyDeployment?.revision"
                      class="btn btn-sm btn-primary" 
                      @click="rollbackToVersion(h)"
                    >
                      回滚到此版本
                    </button>
                    <span v-else class="current-version-text">当前运行版本</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          <div v-else class="empty-state-small">
            <div class="empty-icon">📜</div>
            <div class="empty-text">暂无版本历史记录</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 回滚弹窗 -->
    <!-- 回滚弹窗 -->
    <div v-if="showRollbackModal" class="modal-overlay" @click.self="showRollbackModal = false">
      <div class="modal-content">
        <div class="modal-header">
          <h3>⏪ 回滚 Deployment</h3>
          <button class="close-btn" @click="showRollbackModal = false">×</button>
        </div>
        <div class="modal-body">
          <div class="info-box">
            <div><strong>部署:</strong> {{ rollbackForm.name }}</div>
            <div><strong>命名空间:</strong> {{ rollbackForm.namespace }}</div>
          </div>
          <div v-if="loadingHistory" class="loading-state">加载历史版本...</div>
          <div v-else-if="historyList.length > 0">
            <div class="form-group">
              <label>选择 ReplicaSet（历史版本）</label>
              <select v-model="rollbackForm.replica_set" class="form-select">
                <option value="">请选择版本</option>
                <option v-for="h in historyList" :key="h.name" :value="h.name">
                  {{ h.name }} (版本 {{ h.revision }}, 副本数: {{ h.replicas }}, {{ h.createdAt }})
                </option>
              </select>
            </div>
            <div class="form-hint">
              ℹ️ 提示：选择一个 ReplicaSet 回滚，Deployment 将恢复到该版本的配置
            </div>
          </div>
          <div v-else class="empty-state-small">
            <div class="empty-text">暂无历史版本</div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showRollbackModal = false">取消</button>
          <button class="btn btn-warning" @click="submitRollback" :disabled="rollingBack || !rollbackForm.replica_set">
            {{ rollingBack ? '回滚中...' : '确认回滚' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Pods 列表弹窗 -->
    <div v-if="showPodsModal" class="modal-overlay" @click.self="showPodsModal = false">
      <div class="modal-content modal-lg">
        <div class="modal-header">
          <h3>📦 关联 Pods</h3>
          <button class="close-btn" @click="showPodsModal = false">×</button>
        </div>
        <div class="modal-body">
          <div class="info-box">
            <div><strong>部署:</strong> {{ podsDeployment?.name }}</div>
            <div><strong>命名空间:</strong> {{ podsDeployment?.namespace }}</div>
            <div><strong>Pod 数量:</strong> {{ podsList.length }}</div>
          </div>
          <div v-if="loadingPods" class="loading-state">加载 Pods...</div>
          <div v-else-if="podsList.length > 0">
            <table class="simple-table">
              <thead>
                <tr>
                  <th>名称</th>
                  <th>状态</th>
                  <th>节点</th>
                  <th>重启</th>
                  <th>创建时间</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="pod in podsList" :key="pod.name">
                  <td class="pod-name-cell" :title="pod.name">{{ pod.name }}</td>
                  <td>
                    <span class="status-indicator" :class="(pod.status || 'unknown').toLowerCase()">
                      {{ pod.status }}
                    </span>
                  </td>
                  <td>{{ pod.node || '-' }}</td>
                  <td style="text-align: center;">{{ pod.restartCount || 0 }}</td>
                  <td>{{ pod.createdAt || '-' }}</td>
                  <td>
                    <div class="pod-actions">
                      <!-- 日志按钮（常用，直接显示） -->
                      <button class="icon-btn" title="查看日志" @click="openPodLogs(pod)">
                        📄 日志
                      </button>
                      
                      <!-- 更多菜单 -->
                      <div class="more-btn">
                        <button class="icon-btn" @click="togglePodMoreOptions(pod, $event)" title="更多操作">
                          ⋮ 更多
                        </button>
                        <div v-if="showPodMoreOptions && selectedPodForAction === pod" class="more-menu" :style="podMenuStyle">
                          <button class="menu-item" @click="restartPodFromList(pod)">
                            <span class="menu-icon">🔄</span>
                            <span>重启 Pod</span>
                          </button>
                          <button class="menu-item" @click="openPodDetail(pod)">
                            <span class="menu-icon">📋</span>
                            <span>查看详情</span>
                          </button>
                          <button class="menu-item" @click="openPodEvents(pod)">
                            <span class="menu-icon">📡</span>
                            <span>查看事件</span>
                          </button>
                          <button class="menu-item" @click="openPodUpdateImage(pod)">
                            <span class="menu-icon">🔧</span>
                            <span>更新镜像</span>
                          </button>
                          <button class="menu-item" @click="evictPodFromList(pod)">
                            <span class="menu-icon">⚠️</span>
                            <span>驱逐Pod</span>
                          </button>
                          <div class="menu-divider"></div>
                          <button class="menu-item danger" @click="deletePodFromList(pod)">
                            <span class="menu-icon">🗑️</span>
                            <span>优雅删除</span>
                          </button>
                          <button class="menu-item danger" @click="forceDeletePodFromList(pod)">
                            <span class="menu-icon">💥</span>
                            <span>强制删除</span>
                          </button>
                        </div>
                      </div>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          <div v-else class="empty-state-small">
            <div class="empty-icon">📦</div>
            <div class="empty-text">暂无关联 Pods</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Services 关联弹窗 -->
    <div v-if="showServicesModal" class="modal-overlay" @click.self="showServicesModal = false">
      <div class="modal-content modal-lg">
        <div class="modal-header">
          <h3>🔌 关联 Services</h3>
          <button class="close-btn" @click="showServicesModal = false">×</button>
        </div>
        <div class="modal-body">
          <div class="info-box">
            <div><strong>部署:</strong> {{ servicesDeployment?.name }}</div>
            <div><strong>命名空间:</strong> {{ servicesDeployment?.namespace }}</div>
            <div><strong>Service 数量:</strong> {{ servicesList.length }}</div>
            <div><strong>Selector:</strong> 
              <span class="selector-tags">
                <span v-for="(v, k) in servicesDeployment?.selector" :key="k" class="selector-tag">
                  {{ k }}={{ v }}
                </span>
              </span>
            </div>
          </div>
          <div v-if="loadingServices" class="loading-state">加载 Services...</div>
          <div v-else-if="servicesList.length > 0">
            <table class="simple-table">
              <thead>
                <tr>
                  <th>名称</th>
                  <th>类型</th>
                  <th>Cluster IP</th>
                  <th>External IP</th>
                  <th>端口</th>
                  <th>创建时间</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="svc in servicesList" :key="svc.name">
                  <td>
                    <div class="service-name">
                      <span class="icon">🔌</span>
                      <span>{{ svc.name }}</span>
                    </div>
                  </td>
                  <td>
                    <span class="service-type-badge" :class="(svc.type || 'ClusterIP').toLowerCase()">
                      {{ svc.type || 'ClusterIP' }}
                    </span>
                  </td>
                  <td>{{ svc.clusterIP || '-' }}</td>
                  <td>{{ svc.externalIP || '-' }}</td>
                  <td>{{ svc.ports || '-' }}</td>
                  <td>{{ svc.createdAt || '-' }}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <div v-else class="empty-state-small">
            <div class="empty-icon">🔌</div>
            <div class="empty-text">暂无关联 Services</div>
            <div class="empty-hint">Service 的 selector 需要匹配 Deployment 的 labels</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Pod 日志弹窗（增强版 - 实时日志、高亮、美化样式） -->
    <div v-if="showPodLogsModal" class="modal-overlay" @click.self="closePodLogs">
      <div class="modal-content logs-modal">
        <div class="modal-header">
          <h3>📄 Pod 日志</h3>
          <button class="close-btn" @click="closePodLogs">×</button>
        </div>
        <div class="modal-body">
          <!-- Pod 信息 -->
          <div v-if="selectedPodForAction" class="pod-info-bar">
            <span><strong>命名空间：</strong>{{ selectedPodForAction.namespace }}</span>
            <span><strong>Pod：</strong>{{ selectedPodForAction.name }}</span>
          </div>
          
          <!-- 日志控制栏 -->
          <div class="logs-control-bar">
            <!-- 容器选择 -->
            <div class="control-item" v-if="podContainerList.length > 1">
              <label>容器</label>
              <select v-model="podLogsForm.container" class="form-select">
                <option value="">请选择容器</option>
                <option v-for="c in podContainerList" :key="c" :value="c">{{ c }}</option>
              </select>
            </div>
            <div class="control-item" v-else-if="podContainerList.length === 1">
              <label>容器</label>
              <span class="single-container">{{ podContainerList[0] }}</span>
            </div>
            
            <!-- Tail 行数 -->
            <div class="control-item">
              <label>显示行数</label>
              <select v-model="podLogsForm.tail" class="form-select">
                <option :value="null">全部</option>
                <option :value="10">10 行</option>
                <option :value="50">50 行</option>
                <option :value="100">100 行</option>
                <option :value="200">200 行</option>
                <option :value="500">500 行</option>
                <option :value="1000">1000 行</option>
              </select>
            </div>
            
            <!-- 实时日志开关 -->
            <div class="control-item">
              <label class="follow-toggle">
                <input type="checkbox" v-model="podLogsForm.follow" />
                <span>实时日志</span>
                <span v-if="podLogsForm.follow && isStreamingPodLogs" class="streaming-indicator">●</span>
              </label>
            </div>
            
            <!-- 保持时长（仅实时日志时显示） -->
            <div class="control-item" v-if="podLogsForm.follow">
              <label>保持时长</label>
              <select v-model="podLogsForm.duration" class="form-select">
                <option :value="0">不限制</option>
                <option :value="30">30 秒</option>
                <option :value="60">1 分钟</option>
                <option :value="180">3 分钟</option>
                <option :value="300">5 分钟</option>
                <option :value="600">10 分钟</option>
              </select>
            </div>
            
            <!-- 获取日志按钮 -->
            <button 
              class="btn btn-primary btn-sm" 
              @click="fetchPodLogs" 
              :disabled="loadingPodLogs || (!podLogsForm.container && podContainerList.length > 1)"
            >
              {{ loadingPodLogs ? '加载中...' : '获取日志' }}
            </button>
            
            <!-- 终止加载按钮 -->
            <button 
              v-if="loadingPodLogs || isStreamingPodLogs" 
              class="btn btn-danger btn-sm" 
              @click="stopPodLogLoading"
            >
              终止
            </button>
            
            <!-- 清除日志按钮 -->
            <button 
              class="btn btn-secondary btn-sm" 
              @click="clearPodLogs"
              :disabled="!podLogsContent || loadingPodLogs"
            >
              清除
            </button>
          </div>
          
          <!-- 日志内容区 -->
          <div class="logs-content-wrapper">
            <div v-if="podLogsError" class="error-box">{{ podLogsError }}</div>
            <pre v-else class="logs-content" ref="podLogsContentRef" v-html="highlightedPodLogs"></pre>
          </div>
        </div>
      </div>
    </div>

    <!-- Pod 详情弹窗 -->
    <div v-if="showPodDetailModal" class="modal-overlay" @click.self="showPodDetailModal = false">
      <div class="modal-content">
        <div class="modal-header">
          <h3>📋 Pod 详情</h3>
          <button class="close-btn" @click="showPodDetailModal = false">×</button>
        </div>
        <div class="modal-body">
          <div v-if="loadingPodDetail" class="loading-state">加载中...</div>
          <div v-else-if="podDetailData">
            <pre class="detail-json">{{ JSON.stringify(podDetailData, null, 2) }}</pre>
          </div>
        </div>
      </div>
    </div>

    <!-- Pod 事件弹窗 -->
    <div v-if="showPodEventsModal" class="modal-overlay" @click.self="showPodEventsModal = false">
      <div class="modal-content">
        <div class="modal-header">
          <h3>📡 Pod 事件</h3>
          <button class="close-btn" @click="showPodEventsModal = false">×</button>
        </div>
        <div class="modal-body">
          <div v-if="loadingPodEvents" class="loading-state">加载中...</div>
          <div v-else-if="podEventsData.length > 0">
            <div v-for="(event, idx) in podEventsData" :key="idx" class="event-item">
              <div class="event-type" :class="(event.type || 'normal').toLowerCase()">
                {{ event.type || 'Normal' }}
              </div>
              <div class="event-content">
                <div class="event-reason">{{ event.reason }}</div>
                <div class="event-message">{{ event.message }}</div>
                <div class="event-time">{{ fmtTime(event.event_time || event.lastTimestamp) }}</div>
              </div>
            </div>
          </div>
          <div v-else class="empty-state-small">
            <div class="empty-icon">📬</div>
            <div class="empty-text">暂无事件</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Pod 更新镜像弹窗 -->
    <div v-if="showPodUpdateImageModal" class="modal-overlay" @click.self="closePodUpdateImage">
      <div class="modal-content" style="max-width: 520px;">
        <div class="modal-header">
          <h3>🔧 更新容器镜像</h3>
          <button class="close-btn" @click="closePodUpdateImage">×</button>
        </div>
        <div class="modal-body">
          <div v-if="selectedPodForAction" class="info-box">
            <div><strong>命名空间:</strong> {{ selectedPodForAction.namespace }}</div>
            <div><strong>Pod:</strong> {{ selectedPodForAction.name }}</div>
          </div>

          <div class="form-group" v-if="podContainerListForUpdate.length > 1">
            <label>选择容器</label>
            <select v-model="podUpdateImageForm.container" class="form-select" :disabled="patchingPodImage">
              <option value="" disabled>请选择容器</option>
              <option v-for="c in podContainerListForUpdate" :key="c" :value="c">
                {{ c }}
              </option>
            </select>
          </div>
          <div v-else-if="podContainerListForUpdate.length === 1" class="form-group">
            <label>容器</label>
            <div class="form-static">{{ podContainerListForUpdate[0] }}</div>
          </div>

          <div class="form-group">
            <label>新镜像地址</label>
            <input
              v-model="podUpdateImageForm.image"
              type="text"
              class="form-input"
              :disabled="patchingPodImage"
              placeholder="例如: nginx:1.28 或 registry.example.com/app:v2"
            />
          </div>

          <div v-if="podUpdateImageError" class="error-box" style="margin-top: 12px;">
            {{ podUpdateImageError }}
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closePodUpdateImage" :disabled="patchingPodImage">
            取消
          </button>
          <button class="btn btn-primary" @click="submitPodUpdateImage" :disabled="patchingPodImage || !podUpdateImageForm.container || !podUpdateImageForm.image">
            {{ patchingPodImage ? '更新中...' : '确认更新' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Pod 强制删除确认弹窗 -->
    <div v-if="showPodForceDeleteModal" class="modal-overlay" @click.self="showPodForceDeleteModal = false">
      <div class="modal-content" style="max-width: 480px;">
        <div class="modal-header">
          <h3>💥 强制删除 Pod</h3>
          <button class="close-btn" @click="showPodForceDeleteModal = false">×</button>
        </div>
        <div class="modal-body">
          <p>确认强制删除 Pod：</p>
          <div class="info-box">
            <div><strong>命名空间:</strong> {{ selectedPodForAction?.namespace }}</div>
            <div><strong>名称:</strong> {{ selectedPodForAction?.name }}</div>
          </div>
          <p style="color: #e53e3e; font-size: 13px;">
            ⚠️ 强制删除会立即终止 Pod，不会等待优雅终止期，可能导致数据丢失！
          </p>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showPodForceDeleteModal = false" :disabled="deletingPod">
            取消
          </button>
          <button class="btn btn-danger" @click="confirmForceDeletePod" :disabled="deletingPod">
            {{ deletingPod ? '删除中...' : '强制删除' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Pod 删除确认弹窗 -->
    <div v-if="showPodDeleteModal" class="modal-overlay" @click.self="showPodDeleteModal = false">
      <div class="modal-content" style="max-width: 480px;">
        <div class="modal-header">
          <h3>🗑️ 删除 Pod</h3>
          <button class="close-btn" @click="showPodDeleteModal = false">×</button>
        </div>
        <div class="modal-body">
          <p>确认删除 Pod：</p>
          <div class="info-box">
            <div><strong>命名空间:</strong> {{ selectedPodForAction?.namespace }}</div>
            <div><strong>名称:</strong> {{ selectedPodForAction?.name }}</div>
          </div>
          <p style="color: #718096; font-size: 13px;">删除后，Deployment 控制器会自动创建新的 Pod 副本。</p>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showPodDeleteModal = false">取消</button>
          <button class="btn btn-danger" @click="confirmDeletePod" :disabled="deletingPod">
            {{ deletingPod ? '删除中...' : '确认删除' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Deployment 日志弹窗（增强版 - 支持行数选择和实时日志） -->
    <div v-if="showDeploymentLogsModal" class="modal-overlay" @click.self="closeDeploymentLogs">
      <div class="modal-content logs-modal">
        <div class="modal-header">
          <h3>📄 Deployment 日志</h3>
          <button class="close-btn" @click="closeDeploymentLogs">×</button>
        </div>
        <div class="modal-body">
          <div v-if="selectedDeploymentForLogs" class="info-box" style="margin-bottom: 16px;">
            <div><strong>部署:</strong> {{ selectedDeploymentForLogs.name }}</div>
            <div><strong>命名空间:</strong> {{ selectedDeploymentForLogs.namespace }}</div>
          </div>

          <!-- 日志控制面板 -->
          <div class="logs-controls">
            <div class="control-item">
              <label>Pod 选择</label>
              <select v-model="deploymentLogsForm.selectedPod" class="form-select" @change="onPodChange">
                <option value="">全部 Pod</option>
                <option v-for="pod in deploymentPodsList" :key="pod.name" :value="pod.name">
                  {{ pod.name }}
                </option>
              </select>
            </div>

            <div class="control-item" v-if="deploymentLogsForm.selectedPod">
              <label>容器</label>
              <select v-model="deploymentLogsForm.container" class="form-select">
                <option value="" disabled>选择容器</option>
                <option v-for="c in deploymentContainerList" :key="c" :value="c">
                  {{ c }}
                </option>
              </select>
            </div>

            <div class="control-item">
              <label>行数</label>
              <select v-model="deploymentLogsForm.tail" class="form-select">
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
                <input type="checkbox" v-model="deploymentLogsForm.follow" />
                实时日志
                <span v-if="deploymentLogsForm.follow && isStreamingDeploymentLogs" class="streaming-indicator">●</span>
              </label>
            </div>

            <div class="control-actions">
              <button 
                class="btn btn-primary btn-sm" 
                @click="fetchDeploymentLogs" 
                :disabled="loadingDeploymentLogs || isStreamingDeploymentLogs"
              >
                {{ loadingDeploymentLogs ? '获取中...' : (deploymentLogsForm.follow ? '获取实时日志' : '获取日志') }}
              </button>
              <button 
                v-if="loadingDeploymentLogs || isStreamingDeploymentLogs" 
                class="btn btn-danger btn-sm" 
                @click="stopDeploymentLogStream"
              >
                终止
              </button>
              <button 
                class="btn btn-secondary btn-sm" 
                @click="clearDeploymentLogs"
                :disabled="!deploymentLogsContent"
              >
                清除
              </button>
            </div>
          </div>

          <div v-if="loadingDeploymentLogs && !isStreamingDeploymentLogs" class="loading-state">加载中...</div>
          
          <div v-else-if="deploymentLogsError" class="error-box">
            {{ deploymentLogsError }}
          </div>

          <!-- 日志内容 -->
          <div v-else class="logs-viewer">
            <pre class="logs-content" ref="deploymentLogsContentRef" v-html="highlightedDeploymentLogs"></pre>
          </div>
        </div>
      </div>
    </div>

    <!-- 批量重启预览弹窗（高危操作） -->
    <div v-if="showBatchRestartModal" class="modal-overlay" @click.self="closeBatchRestartModal">
      <div class="modal-content modal-batch-preview modal-warning">
        <div class="modal-header warning-header">
          <h3>🔄 批量重启预览（高危）</h3>
          <button class="close-btn" @click="closeBatchRestartModal">×</button>
        </div>
        <div class="modal-body">
          <!-- 警告提示 -->
          <div class="warning-box">
            <div class="warning-icon-large">⚠️</div>
            <div class="warning-content">
              <div class="warning-title">将重启以下 Deployment</div>
              <ul class="warning-list">
                <li>重启会滚动更新所有 Pod</li>
                <li>可能导致服务短暂不可用</li>
                <li>请确认业务可以承受重启影响</li>
              </ul>
            </div>
          </div>
          
          <!-- 受影响 Deployment 列表 -->
          <div class="preview-section">
            <div class="section-title">受影响 Deployment ({{ selectedDeployments.length }})</div>
            <div class="affected-deployments-detail">
              <div v-for="dep in selectedDeployments" :key="dep.name" class="affected-dep-card">
                <div class="dep-info">
                  <span class="dep-name">🚀 {{ dep.name }}</span>
                  <span class="dep-namespace">{{ dep.namespace }}</span>
                </div>
                <div class="dep-stats">
                  <span class="status-indicator" :class="dep.status.toLowerCase()">
                    {{ dep.status }}
                  </span>
                  <span class="replicas-tag">{{ dep.readyReplicas }}/{{ dep.desiredReplicas }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- 二次确认 -->
          <div class="confirm-section">
            <div class="section-title">请输入 "RESTART" 确认操作</div>
            <input 
              v-model="restartConfirmText" 
              placeholder="请输入 RESTART" 
              class="confirm-input"
              :class="{ valid: restartConfirmText === 'RESTART' }"
            />
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeBatchRestartModal">取消</button>
          <button 
            class="btn btn-warning" 
            @click="executeBatchRestart" 
            :disabled="restartConfirmText !== 'RESTART' || batchExecuting"
          >
            {{ batchExecuting ? '执行中...' : '确认重启' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 批量扩缩容弹窗 -->
    <div v-if="showBatchScaleModal" class="modal-overlay" @click.self="closeBatchScaleModal">
      <div class="modal-content modal-batch-preview">
        <div class="modal-header">
          <h3>📊 批量扩缩容预览</h3>
          <button class="close-btn" @click="closeBatchScaleModal">×</button>
        </div>
        <div class="modal-body">
          <!-- 受影响 Deployment 列表 -->
          <div class="preview-section">
            <div class="section-title">受影响 Deployment ({{ selectedDeployments.length }})</div>
            <div class="affected-deployments">
              <div v-for="dep in selectedDeployments" :key="dep.name" class="affected-item">
                <span class="dep-name">🚀 {{ dep.name }}</span>
                <span class="dep-replicas">当前: {{ dep.readyReplicas }}/{{ dep.desiredReplicas }}</span>
              </div>
            </div>
          </div>
          
          <!-- 目标副本数 -->
          <div class="preview-section">
            <div class="section-title">目标副本数</div>
            <input 
              type="number" 
              v-model="batchScaleReplicas" 
              min="0" 
              class="form-input"
              style="width: 120px;"
            />
          </div>

          <!-- 变更预览 -->
          <div class="preview-section">
            <div class="section-title">变更预览</div>
            <div class="change-preview">
              <div v-for="dep in selectedDeployments" :key="dep.name" class="change-item">
                <span>{{ dep.name }}:</span>
                <span class="old-value">{{ dep.desiredReplicas }}</span>
                <span class="arrow">→</span>
                <span class="new-value">{{ batchScaleReplicas }}</span>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeBatchScaleModal">取消</button>
          <button 
            class="btn btn-primary" 
            @click="executeBatchScale" 
            :disabled="batchExecuting"
          >
            {{ batchExecuting ? '执行中...' : '确认扩缩容' }}
          </button>
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
              <div class="warning-title">将删除以下 Deployment 及其所有 Pod</div>
              <ul class="warning-list">
                <li>所有关联的 ReplicaSet 和 Pod 将被删除</li>
                <li>此操作不可撤销！</li>
              </ul>
            </div>
          </div>
          
          <!-- 受影响 Deployment 列表 -->
          <div class="preview-section">
            <div class="section-title">受影响 Deployment ({{ selectedDeployments.length }})</div>
            <div class="affected-deployments-detail">
              <div v-for="dep in selectedDeployments" :key="dep.name" class="affected-dep-card">
                <div class="dep-info">
                  <span class="dep-name">🚀 {{ dep.name }}</span>
                  <span class="dep-namespace">{{ dep.namespace }}</span>
                </div>
                <div class="dep-stats">
                  <span class="status-indicator" :class="dep.status.toLowerCase()">
                    {{ dep.status }}
                  </span>
                  <span class="replicas-tag">{{ dep.readyReplicas }}/{{ dep.desiredReplicas }}</span>
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

    <!-- YAML 查看/编辑模态框 -->
    <div v-if="showYamlModal" class="modal-overlay" @click.self="closeYamlModal">
      <div 
        ref="yamlModalRef"
        class="modal-content yaml-modal resizable-modal"
        :style="yamlModalStyle"
      >
        <!-- 8个拖拽手柄 -->
        <div class="resize-handle resize-handle-top" @mousedown="yamlStartResize($event, 'top')"></div>
        <div class="resize-handle resize-handle-bottom" @mousedown="yamlStartResize($event, 'bottom')"></div>
        <div class="resize-handle resize-handle-left" @mousedown="yamlStartResize($event, 'left')"></div>
        <div class="resize-handle resize-handle-right" @mousedown="yamlStartResize($event, 'right')"></div>
        <div class="resize-handle resize-handle-top-left" @mousedown="yamlStartResize($event, 'top-left')"></div>
        <div class="resize-handle resize-handle-top-right" @mousedown="yamlStartResize($event, 'top-right')"></div>
        <div class="resize-handle resize-handle-bottom-left" @mousedown="yamlStartResize($event, 'bottom-left')"></div>
        <div class="resize-handle resize-handle-bottom-right" @mousedown="yamlStartResize($event, 'bottom-right')"></div>
        
        <div class="modal-header">
          <h3>📝 {{ yamlEditMode ? '编辑 YAML' : '查看 YAML' }} - {{ selectedYamlDeployment?.name }}</h3>
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
          <div v-if="loadingYaml" class="loading-state">Loading YAML...</div>
          <div v-else-if="yamlError" class="error-box">{{ yamlError }}</div>
          <div v-else class="yaml-editor-wrapper">
            <textarea v-if="yamlEditMode" v-model="yamlContent" class="yaml-editor" spellcheck="false"></textarea>
            <pre v-else class="yaml-content">{{ yamlContent }}</pre>
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
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch, watchEffect } from 'vue'
import { Message } from '@arco-design/web-vue'
import Pagination from '@/components/Pagination.vue'
import deploymentsApi from '@/api/cluster/workloads/deployments'
import podsApi from '@/api/cluster/workloads/pods'
import serviceApi from '@/api/cluster/networking/service'
import namespaceApi from '@/api/cluster/config/namespace'
import { useClusterStore } from '@/stores/cluster'
import { useResizableModal } from '@/composables/useResizableModal'

// ===== 获取认证头（复用 http.js 逻辑） =====
const getAuthHeaders = () => {
  const headers = { 'Content-Type': 'application/json' }
  
  // 1) JWT Token
  const token = localStorage.getItem('token') || sessionStorage.getItem('token')
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }
  
  // 2) X-Cluster-ID
  const clusterStore = useClusterStore()
  const getClusterIdFromPath = () => {
    try {
      const m = window.location.pathname.match(/\/c\/([^/]+)/)
      return m ? decodeURIComponent(m[1]) : ''
    } catch {
      return ''
    }
  }
  const cid = clusterStore.current?.id ?? getClusterIdFromPath()
  if (cid !== undefined && cid !== null && cid !== '') {
    headers['X-Cluster-ID'] = String(cid)
  }
  
  return headers
}

// =========================
// 状态变量
// =========================
const loading = ref(false)
const errorMsg = ref('')
const searchQuery = ref('')
const statusFilter = ref('all')
const namespaceFilter = ref('')
const currentPage = ref(1)
const itemsPerPage = ref(10)
const namespaces = ref([])
const deployments = ref([])

// 视图模式：table（表格） 或 card（卡片）
const viewMode = ref('table')

// ========== 批量操作相关 ==========
const batchMode = ref(false)
const selectedDeployments = ref([])
const showBatchScaleModal = ref(false)
const showBatchDeleteModal = ref(false)
const showBatchRestartModal = ref(false)
const batchScaleReplicas = ref(1)
const deleteConfirmText = ref('')
const restartConfirmText = ref('')
const batchExecuting = ref(false)

// ========== YAML 查看/编辑相关 ==========
const showYamlModal = ref(false)
const selectedYamlDeployment = ref(null)
const yamlContent = ref('')
const yamlEditMode = ref(false)
const loadingYaml = ref(false)
const savingYaml = ref(false)
const yamlError = ref('')

const enterBatchMode = () => {
  batchMode.value = true
  // 默认全选当前页，用户可取消不需要的项
  selectedDeployments.value = [...paginatedDeployments.value]
}

const exitBatchMode = () => {
  batchMode.value = false
  selectedDeployments.value = []
}

const clearSelection = () => {
  selectedDeployments.value = []
}

const isDeploymentSelected = (dep) => {
  return selectedDeployments.value.some(d => d.name === dep.name && d.namespace === dep.namespace)
}

const toggleDeploymentSelection = (dep) => {
  const index = selectedDeployments.value.findIndex(d => d.name === dep.name && d.namespace === dep.namespace)
  if (index >= 0) {
    selectedDeployments.value.splice(index, 1)
  } else {
    selectedDeployments.value.push(dep)
  }
}

const isAllSelected = computed(() => {
  return paginatedDeployments.value.length > 0 && 
         paginatedDeployments.value.every(dep => isDeploymentSelected(dep))
})

// 部分选中状态
const isPartialSelected = computed(() => {
  if (paginatedDeployments.value.length === 0) return false
  const selectedCount = paginatedDeployments.value.filter(dep => isDeploymentSelected(dep)).length
  return selectedCount > 0 && selectedCount < paginatedDeployments.value.length
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
    paginatedDeployments.value.forEach(dep => {
      const index = selectedDeployments.value.findIndex(d => d.name === dep.name && d.namespace === dep.namespace)
      if (index >= 0) selectedDeployments.value.splice(index, 1)
    })
  } else {
    paginatedDeployments.value.forEach(dep => {
      if (!isDeploymentSelected(dep)) {
        selectedDeployments.value.push(dep)
      }
    })
  }
}

// 批量重启预览
const openBatchRestartPreview = () => {
  restartConfirmText.value = ''
  showBatchRestartModal.value = true
}

const closeBatchRestartModal = () => {
  showBatchRestartModal.value = false
  restartConfirmText.value = ''
}

// 执行批量重启
const executeBatchRestart = async () => {
  if (restartConfirmText.value !== 'RESTART') return
  
  batchExecuting.value = true
  let successCount = 0
  let failCount = 0
  
  for (const dep of selectedDeployments.value) {
    try {
      await deploymentsApi.restart({
        name: dep.name,
        namespace: dep.namespace
      })
      successCount++
    } catch (e) {
      console.error(`Failed to restart ${dep.name}:`, e)
      failCount++
    }
  }
  
  batchExecuting.value = false
  showBatchRestartModal.value = false
  restartConfirmText.value = ''
  
  if (failCount === 0) {
    Message.success({ content: `成功重启 ${successCount} 个 Deployment`, duration: 2200 })
  } else {
    Message.warning({ content: `成功 ${successCount} 个，失败 ${failCount} 个`, duration: 2200 })
  }
  
  exitBatchMode()
  refreshList()
}

// 批量扩缩容
const openBatchScaleModal = () => {
  batchScaleReplicas.value = 1
  showBatchScaleModal.value = true
}

const closeBatchScaleModal = () => {
  showBatchScaleModal.value = false
}

const executeBatchScale = async () => {
  batchExecuting.value = true
  let successCount = 0
  let failCount = 0
  
  for (const dep of selectedDeployments.value) {
    try {
      await deploymentsApi.scale({
        name: dep.name,
        namespace: dep.namespace,
        scale_num: parseInt(batchScaleReplicas.value)  // 后端字段名是 scale_num
      })
      successCount++
    } catch (e) {
      console.error(`Failed to scale ${dep.name}:`, e)
      failCount++
    }
  }
  
  batchExecuting.value = false
  showBatchScaleModal.value = false
  
  if (failCount === 0) {
    Message.success({ content: `成功扩缩容 ${successCount} 个 Deployment`, duration: 2200 })
  } else {
    Message.warning({ content: `成功 ${successCount} 个，失败 ${failCount} 个`, duration: 2200 })
  }
  
  refreshList()
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
  
  for (const dep of selectedDeployments.value) {
    try {
      await deploymentsApi.delete({
        name: dep.name,
        namespace: dep.namespace
      })
      successCount++
    } catch (e) {
      console.error(`Failed to delete ${dep.name}:`, e)
      failCount++
    }
  }
  
  batchExecuting.value = false
  showBatchDeleteModal.value = false
  deleteConfirmText.value = ''
  
  if (failCount === 0) {
    Message.success({ content: `成功删除 ${successCount} 个 Deployment`, duration: 2200 })
  } else {
    Message.warning({ content: `成功 ${successCount} 个，失败 ${failCount} 个`, duration: 2200 })
  }
  
  exitBatchMode()
  refreshList()
}

// 自动刷新
const autoRefresh = ref(false)
let autoRefreshTimer = null
const AUTO_REFRESH_INTERVAL = 90000 // 90秒

// 更多菜单
const showMoreOptions = ref(false)
const selectedDeployment = ref(null)
const menuStyle = ref({})

// Modal states
const showCreateModal = ref(false)
const showEditModal = ref(false)
const showDeleteModal = ref(false)
const showViewModal = ref(false)
const showScaleModal = ref(false)
const showEventsModal = ref(false)
const showUpdateImageModal = ref(false)
const showRollbackModal = ref(false)
const showPodsModal = ref(false)
const showServicesModal = ref(false)        // Service 关联弹窗
const showDeploymentLogsModal = ref(false)  // Deployment 日志弹窗
const showHistoryModal = ref(false)          // 版本记录弹窗

// ========== 可拖拽调整大小的模态框 ==========
// 创建模态框
const {
  modalRef: createModalRef,
  modalStyle: createModalStyle,
  startResize: createStartResize
} = useResizableModal({ initialWidth: '1200px', initialHeight: '80vh' })

// 编辑模态框
const {
  modalRef: editModalRef,
  modalStyle: editModalStyle,
  startResize: editStartResize
} = useResizableModal({ initialWidth: '900px', initialHeight: '70vh' })

// YAML模态框
const {
  modalRef: yamlModalRef,
  modalStyle: yamlModalStyle,
  startResize: yamlStartResize
} = useResizableModal({ initialWidth: '1100px', initialHeight: '80vh' })

// YAML 创建相关
const createMode = ref('form') // 'form' | 'yaml'

// 监听 createMode 变化，切换到 YAML 模式时如果内容为空则自动加载模板
watch(createMode, (newMode) => {
  if (newMode === 'yaml' && !yamlContent.value.trim()) {
    loadDeploymentYamlTemplate()
  }
})

// 创建命名空间相关
const showNamespaceInput = ref(false)
const newNamespace = ref('')
const creatingNamespace = ref(false)

// Loading states
const loadingDetail = ref(false)
const loadingEvents = ref(false)
const loadingHistory = ref(false)
const loadingPods = ref(false)
const scaling = ref(false)
const updatingImage = ref(false)
const rollingBack = ref(false)
const deleting = ref(false)

// Data
const detailData = ref(null)
const eventsData = ref([])
const historyList = ref([])
const historyDeployment = ref(null)  // 用于版本记录弹窗
const podsList = ref([])
const containerList = ref([])
const podsDeployment = ref(null)
const servicesList = ref([])                // 关联 Service 列表
const servicesDeployment = ref(null)        // 查看 Service 的 Deployment
const loadingServices = ref(false)          // 加载 Service 状态

// Pod 操作相关状态
const showPodLogsModal = ref(false)
const showPodDetailModal = ref(false)
const showPodEventsModal = ref(false)
const showPodDeleteModal = ref(false)
const showPodForceDeleteModal = ref(false)  // 强制删除弹窗
const showPodUpdateImageModal = ref(false)  // 更新镜像弹窗
const showPodMoreOptions = ref(false)       // 更多菜单
const podMenuStyle = ref({})                // 菜单位置样式
const selectedPodForAction = ref(null)
const podContainerList = ref([])
const podLogsForm = ref({ container: '', tail: 100, follow: false, duration: 0 })
const podLogsContent = ref('')
const podLogsError = ref('')
const loadingPodLogs = ref(false)
const isStreamingPodLogs = ref(false)
const podLogsContentRef = ref(null)
let podLogAbortController = null
let podLogTimeoutId = null
const podDetailData = ref(null)
const loadingPodDetail = ref(false)
const podEventsData = ref([])
const loadingPodEvents = ref(false)
const deletingPod = ref(false)

// 更新镜像相关
const podUpdateImageForm = ref({ container: '', image: '' })
const podContainerListForUpdate = ref([])
const patchingPodImage = ref(false)
const podUpdateImageError = ref('')

// Deployment 日志相关
const deploymentLogsContent = ref('')  // 日志内容
const deploymentPodsList = ref([])  // Deployment 的 Pod 列表
const deploymentContainerList = ref([])  // 选中 Pod 的容器列表
const deploymentLogsForm = ref({
  selectedPod: '',      // 选中的 Pod
  container: '',        // 选中的容器
  tail: 100,            // 显示行数
  follow: false         // 是否实时日志
})
const loadingDeploymentLogs = ref(false)
const deploymentLogsError = ref('')
const selectedDeploymentForLogs = ref(null)
const isStreamingDeploymentLogs = ref(false)
const deploymentLogsContentRef = ref(null)
let deploymentLogAbortController = null

// Forms
const deploymentForm = ref({
  name: '',
  namespace: 'default',
  replicas: 1,
  image: '',
  updateStrategy: 'RollingUpdate',
  labels: [{ key: 'app', value: '' }],  // 标签数组
  createService: false,
  serviceName: '',
  serviceType: 'ClusterIP',
  servicePorts: [{ port: 80, targetPort: 80, nodePort: null }],
  // 调度规则
  schedulingPolicy: 'default', // default/spread/pack/custom
  nodeSelector: [],  // [{key: '', value: ''}]
  tolerations: [],    // [{key: '', operator: '', value: '', effect: '', tolerationSeconds: null}]
  nodeAffinityRules: [], // [{key: '', operator: 'In', values: '', required: false, weight: 50}]
  topologySpreadConfigs: [], // [{topologyKey: '', maxSkew: 1, whenUnsatisfiable: 'ScheduleAnyway'}]
  // 资源配置 (Resources)
  resources: {
    cpuRequest: '',
    cpuLimit: '',
    memoryRequest: '',
    memoryLimit: ''
  },
  // 探针配置 (Probes)
  probes: {
    enableLiveness: false,
    livenessProbe: {
      type: 'HTTP',
      port: 8080,
      path: '/healthz',
      command: '',
      initialDelaySeconds: 10,
      periodSeconds: 10,
      timeoutSeconds: 1,
      successThreshold: 1,
      failureThreshold: 3
    },
    enableReadiness: false,
    readinessProbe: {
      type: 'HTTP',
      port: 8080,
      path: '/ready',
      command: '',
      initialDelaySeconds: 5,
      periodSeconds: 10,
      timeoutSeconds: 1,
      successThreshold: 1,
      failureThreshold: 3
    },
    enableStartup: false,
    startupProbe: {
      type: 'HTTP',
      port: 8080,
      path: '/startup',
      command: '',
      initialDelaySeconds: 0,
      periodSeconds: 10,
      timeoutSeconds: 1,
      successThreshold: 1,
      failureThreshold: 30
    }
  }
})

// 折叠面板状态
const expandedSections = ref({
  resources: false,
  probes: false
})

// 切换折叠面板
const toggleSection = (section) => {
  expandedSections.value[section] = !expandedSections.value[section]
}

// Service 类型配置
const serviceTypes = [
  { 
    value: 'None', 
    label: 'None', 
    icon: '🚫',
    description: '不创建 Service'
  },
  { 
    value: 'ClusterIP', 
    label: 'ClusterIP', 
    icon: '🏢',
    description: '集群内部访问'
  },
  { 
    value: 'NodePort', 
    label: 'NodePort', 
    icon: '🌐',
    description: '节点端口访问'
  },
  { 
    value: 'LoadBalancer', 
    label: 'LoadBalancer', 
    icon: '⚖️',
    description: '负载均衡器（云环境）'
  }
]

// 调度策略配置
const schedulingPolicies = [
  {
    value: 'default',
    label: '默认规则',
    description: '按照默认的规则将容器组副本调度到节点'
  },
  {
    value: 'spread',
    label: '分散调度',
    description: '尽可能将容器组副本调度到不同的节点上'
  },
  {
    value: 'pack',
    label: '集中调度',
    description: '尽可能将容器组副本调度到同一节点上'
  },
  {
    value: 'custom',
    label: '自定义规则',
    description: '按照自定义的规则将容器组调度到节点'
  }
]

const editForm = ref({
  name: '',
  namespace: 'default',
  replicas: 1,
  image: '',
  container: '',
  updateStrategy: 'RollingUpdate'
})

const scaleForm = ref({ namespace: '', name: '', replicas: 1 })
const updateImageForm = ref({ namespace: '', name: '', container: '', image: '' })
const rollbackForm = ref({ namespace: '', name: '', revision: 0 })
const selectorInput = ref('')
const editSelectorInput = ref('')
const deploymentToDelete = ref('')
const deploymentNamespaceToDelete = ref('')

// =========================
// 内联编辑状态
// =========================
const inlineEdit = ref({
  key: '',      // 当前编辑的 key，如 'replicas-nginx' 或 'image-nginx'
  value: null,  // 编辑的值
  original: null // 原始值，用于取消时恢复
})
const inlineInputRef = ref(null)

// 副本数扩缩容状态追踪（每个 deployment 独立）
const scalingMap = ref({})

// =========================
// 生命周期
// =========================
onMounted(() => {
  document.addEventListener('click', handlePodClickOutside)
  document.addEventListener('scroll', handlePodScroll, true)
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('scroll', handleScroll, true)
  fetchNamespaces().then(fetchDeployments)
})

onUnmounted(() => {
  document.removeEventListener('click', handlePodClickOutside)
  document.removeEventListener('scroll', handlePodScroll, true)
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('scroll', handleScroll, true)
  stopAutoRefresh()
})

// =========================
// 更多菜单控制
// =========================
const toggleMoreOptions = (deployment, event) => {
  if (selectedDeployment.value === deployment && showMoreOptions.value) {
    showMoreOptions.value = false
    selectedDeployment.value = null
  } else {
    selectedDeployment.value = deployment
    showMoreOptions.value = true
    const button = event.target.closest('.more-btn')
    if (button) {
      const rect = button.getBoundingClientRect()
      const viewportHeight = window.innerHeight
      const menuHeight = 280
      let style = { position: 'fixed', left: rect.left + 'px' }
      if (rect.bottom + menuHeight > viewportHeight) {
        style.bottom = (viewportHeight - rect.top + 4) + 'px'
      } else {
        style.top = (rect.bottom + 4) + 'px'
      }
      menuStyle.value = style
    }
  }
}

const handleClickOutside = (event) => {
  if (showMoreOptions.value && !event.target.closest('.more-btn')) {
    showMoreOptions.value = false
    selectedDeployment.value = null
  }
}

const handleScroll = () => {
  if (showMoreOptions.value) {
    showMoreOptions.value = false
    selectedDeployment.value = null
  }
}

// =========================
// 自动刷新
// =========================
const startAutoRefresh = () => {
  if (autoRefreshTimer) return
  autoRefreshTimer = setInterval(() => {
    if (!loading.value) fetchDeployments()
  }, AUTO_REFRESH_INTERVAL)
}

const stopAutoRefresh = () => {
  if (autoRefreshTimer) {
    clearInterval(autoRefreshTimer)
    autoRefreshTimer = null
  }
}

watch(autoRefresh, (val) => {
  if (val) startAutoRefresh()
  else stopAutoRefresh()
})

watch(namespaceFilter, () => {
  currentPage.value = 1
  fetchDeployments()
})

// =========================
// API 调用
// =========================
const fetchNamespaces = async () => {
  try {
    const res = await namespaceApi.list({ page: 1, limit: 1000 })
    const list = res?.data?.list || res?.data?.items || []
    namespaces.value = (Array.isArray(list) ? list : []).map(ns =>
      typeof ns === 'string' ? ns : (ns?.metadata?.name || ns?.name || ns)
    ).filter(Boolean)
  } catch (e) {
    console.error('获取命名空间失败:', e)
    namespaces.value = ['default', 'kube-system']
  }
}

const fetchDeployments = async () => {
  loading.value = true
  errorMsg.value = ''
  try {
    const params = { namespace: namespaceFilter.value || '', page: 1, limit: 1000 }
    const res = await deploymentsApi.list(params)
    // 后端返回 data.list，已包含处理后的状态信息
    const list = res.data?.list || res.data?.items || []
    if (res.code === 0 && list.length > 0) {
      deployments.value = list.map(item => ({
        name: item.name,
        namespace: item.namespace,
        status: item.status,  // 直接使用后端计算的状态
        statusReason: item.status_reason || '',  // 状态原因
        desiredReplicas: item.replicas || 0,
        readyReplicas: item.ready_replicas || 0,
        availableReplicas: item.available_replicas || 0,
        updatedReplicas: item.updated_replicas || 0,
        image: item.image || (item.images && item.images[0]) || '',
        selector: item.selector || {},
        updateStrategy: item.update_strategy || 'RollingUpdate',
        createdAt: item.created_at || '',
        containers: item.containers || [],
        metrics: null  // 初始化 metrics 字段
      }))
      
      // 异步获取 metrics（不阻塞主流程）
      fetchDeploymentsMetrics()
    } else {
      deployments.value = []
    }
  } catch (e) {
    console.error('获取部署列表失败:', e)
    errorMsg.value = e?.msg || e?.message || '获取部署列表失败'
    deployments.value = []
  } finally {
    loading.value = false
  }
}

// 获取 Deployments 的资源消耗（通过聚合其管理的 Pod metrics）
const fetchDeploymentsMetrics = async () => {
  if (!namespaceFilter.value) return  // 需要指定 namespace
  
  try {
    // 批量获取该命名空间下所有 Pod 的 metrics
    const res = await podsApi.metricsList({ namespace: namespaceFilter.value })
    const metricsMap = res?.data || {}
    
    // 为每个 Deployment 聚合其 Pod 的 metrics
    deployments.value.forEach(deployment => {
      // 根据 selector 匹配 Pod（简化版：通过名称前缀匹配）
      const deploymentPodMetrics = []
      let totalCpuMillicores = 0
      let totalMemoryBytes = 0
      
      Object.keys(metricsMap).forEach(podName => {
        // 简单匹配：Pod 名称通常以 deployment-name 开头
        if (podName.startsWith(deployment.name + '-')) {
          const podMetric = metricsMap[podName]
          deploymentPodMetrics.push(podMetric)
          
          // 解析 CPU（去掉 'm' 后缀）
          if (podMetric.total_cpu) {
            const cpuValue = podMetric.total_cpu.replace('m', '')
            totalCpuMillicores += parseFloat(cpuValue) || 0
          }
          
          // 解析内存（转换为 bytes）
          if (podMetric.total_memory) {
            totalMemoryBytes += parseMemoryToBytes(podMetric.total_memory)
          }
        }
      })
      
      // 保存聚合后的 metrics
      if (deploymentPodMetrics.length > 0) {
        deployment.metrics = {
          podCount: deploymentPodMetrics.length,
          totalCpu: formatCpu(totalCpuMillicores),
          totalMemory: formatMemory(totalMemoryBytes),
          pods: deploymentPodMetrics
        }
      }
    })
  } catch (e) {
    console.warn('获取 Deployment metrics 失败（可能 metrics-server 未安装）:', e)
  }
}

// 解析内存字符串为 bytes
const parseMemoryToBytes = (memStr) => {
  const units = {
    'Ki': 1024,
    'Mi': 1024 * 1024,
    'Gi': 1024 * 1024 * 1024,
    'Ti': 1024 * 1024 * 1024 * 1024
  }
  
  for (const [unit, multiplier] of Object.entries(units)) {
    if (memStr.endsWith(unit)) {
      return parseFloat(memStr.replace(unit, '')) * multiplier
    }
  }
  
  // 纯数字视为 bytes
  return parseFloat(memStr) || 0
}

// 格式化 CPU（millicores）
const formatCpu = (millicores) => {
  if (millicores >= 1000) {
    return (millicores / 1000).toFixed(2)
  }
  return millicores.toFixed(0) + 'm'
}

// 格式化内存（bytes）
const formatMemory = (bytes) => {
  const units = [
    { unit: 'Ti', divisor: 1024 * 1024 * 1024 * 1024 },
    { unit: 'Gi', divisor: 1024 * 1024 * 1024 },
    { unit: 'Mi', divisor: 1024 * 1024 },
    { unit: 'Ki', divisor: 1024 }
  ]
  
  for (const { unit, divisor } of units) {
    if (bytes >= divisor) {
      return (bytes / divisor).toFixed(2) + unit
    }
  }
  
  return bytes.toFixed(0) + 'B'
}

// 时间格式化工具函数
const fmtTime = (ts) => {
  if (!ts) return '-'
  try {
    const d = new Date(ts)
    return d.toLocaleString('zh-CN')
  } catch { return ts }
}

// =========================
// 筛选与搜索
// =========================
const setStatusFilter = (status) => {
  statusFilter.value = status
  currentPage.value = 1
}

let searchDebounceTimer = null
const onSearchInput = () => {
  if (searchDebounceTimer) clearTimeout(searchDebounceTimer)
  searchDebounceTimer = setTimeout(() => { currentPage.value = 1 }, 300)
}

const refreshList = () => fetchDeployments()

// =========================
// 计算属性
// =========================
const filteredDeployments = computed(() => {
  let result = deployments.value
  const q = searchQuery.value.trim().toLowerCase()
  if (q) {
    result = result.filter(d =>
      d.name.toLowerCase().includes(q) ||
      d.namespace.toLowerCase().includes(q) ||
      d.image.toLowerCase().includes(q)
    )
  }
  if (statusFilter.value !== 'all') {
    result = result.filter(d => d.status === statusFilter.value)
  }
  return result
})

const paginatedDeployments = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage.value
  return filteredDeployments.value.slice(start, start + itemsPerPage.value)
})

// =========================
// 详情
// =========================
const viewDeployment = async (deployment) => {
  showMoreOptions.value = false
  loadingDetail.value = true
  showViewModal.value = true
  try {
    const res = await deploymentsApi.detail({ namespace: deployment.namespace, name: deployment.name })
    detailData.value = res.code === 0 ? res.data : deployment
  } catch (e) {
    detailData.value = deployment
  } finally {
    loadingDetail.value = false
  }
}

// =========================
// 事件
// =========================
const openEvents = async (deployment) => {
  showMoreOptions.value = false
  loadingEvents.value = true
  eventsData.value = []
  showEventsModal.value = true
  try {
    const res = await deploymentsApi.events({ namespace: deployment.namespace, name: deployment.name })
    // 后端返回 res.data.events
    eventsData.value = res.code === 0 ? (res.data?.events || res.data?.items || res.data || []) : []
  } catch (e) {
    console.error('获取事件失败:', e)
  } finally {
    loadingEvents.value = false
  }
}

// =========================
// Pods
// =========================
const viewPods = async (deployment) => {
  podsDeployment.value = deployment
  loadingPods.value = true
  podsList.value = []
  showPodsModal.value = true
  try {
    const res = await deploymentsApi.pods({ namespace: deployment.namespace, name: deployment.name })
    // 后端返回 res.data.pods
    const list = res.code === 0 ? (res.data?.pods || res.data?.items || res.data?.list || res.data || []) : []
    podsList.value = list.map(p => {
      const metadata = p.metadata || {}
      const spec = p.spec || {}
      const status = p.status || {}
      return {
        name: metadata.name || p.name,
        namespace: metadata.namespace || deployment.namespace,
        status: status.phase || p.status || 'Unknown',
        node: spec.nodeName || p.node || '-',
        restartCount: status.containerStatuses?.[0]?.restartCount || p.restartCount || 0,
        createdAt: fmtTime(metadata.creationTimestamp || p.created_at),
        containers: spec.containers?.map(c => c.name) || p.containers || [],
        raw: p
      }
    })
  } catch (e) {
    console.error('获取 Pods 失败:', e)
  } finally {
    loadingPods.value = false
  }
}

// =========================
// Services 关联
// =========================
const viewRelatedServices = async (deployment) => {
  showMoreOptions.value = false
  servicesDeployment.value = deployment
  loadingServices.value = true
  servicesList.value = []
  showServicesModal.value = true
  
  try {
    // 获取所有 Service 列表，然后在前端过滤
    const res = await serviceApi.list({ 
      namespace: deployment.namespace,
      page: 1,
      limit: 1000  // 获取足够多的数据进行过滤
    })
    
    if (res.code !== 0) {
      Message.error({ content: '获取 Service 列表失败' })
      return
    }
    
    const allServices = res.data?.list || res.data || []
    
    // 过滤出匹配 Deployment selector 的 Service
    const deploymentLabels = deployment.selector || {}
    
    servicesList.value = allServices.filter(svc => {
      const svcSelector = svc.selector || {}
      
      // 如果 Service 没有 selector，不匹配
      if (Object.keys(svcSelector).length === 0) {
        return false
      }
      
      // 检查 Service 的 selector 是否匹配 Deployment 的 labels
      // Service 的每一个 selector 键值对都必须在 Deployment labels 中存在
      return Object.entries(svcSelector).every(([key, value]) => {
        return deploymentLabels[key] === value
      })
    }).map(svc => {
      // 转换为前端需要的格式
      return {
        name: svc.name,
        namespace: svc.namespace,
        type: svc.type,
        clusterIP: svc.cluster_ip || svc.clusterIP,
        externalIP: svc.external_ip || svc.externalIP,
        ports: svc.ports,
        targetPort: svc.target_port || svc.targetPort,
        selector: svc.selector,
        createdAt: svc.created_at || svc.createdAt,
        raw: svc
      }
    })
    
    if (servicesList.value.length === 0) {
      Message.info({ 
        content: `没有找到匹配的 Service，请确认 Service 的 selector 与 Deployment 的 labels 一致`,
        duration: 3000
      })
    }
  } catch (e) {
    console.error('获取关联 Service 失败:', e)
    Message.error({ content: '获取关联 Service 失败' })
  } finally {
    loadingServices.value = false
  }
}

// =========================
// Pod 操作：更多菜单控制
// =========================
const togglePodMoreOptions = (pod, event) => {
  if (selectedPodForAction.value === pod && showPodMoreOptions.value) {
    showPodMoreOptions.value = false
    selectedPodForAction.value = null
  } else {
    selectedPodForAction.value = pod
    showPodMoreOptions.value = true
    
    // 计算菜单位置，使用 fixed 定位防止被父容器裁剪
    const button = event.target.closest('.more-btn')
    if (button) {
      const rect = button.getBoundingClientRect()
      const viewportHeight = window.innerHeight
      const viewportWidth = window.innerWidth
      const menuHeight = 280 // 预估菜单高度
      const menuWidth = 180 // 预估菜单宽度

      let style = {
        position: 'fixed',
      }

      // 垂直方向：如果下方空间不足，向上展开
      if (rect.bottom + menuHeight > viewportHeight) {
        style.bottom = (viewportHeight - rect.top + 4) + 'px'
      } else {
        style.top = (rect.bottom + 4) + 'px'
      }

      // 水平方向：如果右侧空间不足，向左展开
      if (rect.right + menuWidth > viewportWidth) {
        style.right = (viewportWidth - rect.right) + 'px'
      } else {
        style.left = rect.left + 'px'
      }

      podMenuStyle.value = style
    }
  }
}

// 点击外部关闭菜单
const handlePodClickOutside = (event) => {
  if (showPodMoreOptions.value && !event.target.closest('.more-btn')) {
    showPodMoreOptions.value = false
    selectedPodForAction.value = null
  }
}

// 滚动时关闭菜单（因为使用 fixed 定位）
const handlePodScroll = () => {
  if (showPodMoreOptions.value) {
    showPodMoreOptions.value = false
    selectedPodForAction.value = null
  }
}

// =========================
// Pod 操作：日志
// =========================
const openPodLogs = (pod) => {
  selectedPodForAction.value = pod
  podContainerList.value = pod.containers || []
  
  // ✅ 单容器自动选中，多容器清空（让用户选择）
  if (podContainerList.value.length === 1) {
    podLogsForm.value = {
      container: podContainerList.value[0],
      tail: 100,
      follow: false,
      duration: 0
    }
  } else {
    podLogsForm.value = {
      container: '',  // 多容器时清空，强制用户选择
      tail: 100,
      follow: false,
      duration: 0
    }
  }
  
  podLogsContent.value = ''
  podLogsError.value = ''
  showPodLogsModal.value = true
}

const fetchPodLogs = async () => {
  if (!selectedPodForAction.value) return
  
  // 确定容器名称
  let container = podLogsForm.value.container || ''
  
  // 如果只有一个容器，自动使用
  if (podContainerList.value.length === 1) {
    container = podContainerList.value[0]
  }
  
  // 多容器时必须选择容器
  if (podContainerList.value.length > 1 && !container) {
    podLogsError.value = '请选择容器'
    return
  }
  
  // 确保容器名不为空
  if (!container) {
    podLogsError.value = '无法确定容器名称'
    return
  }
  
  stopPodLogLoading()
  loadingPodLogs.value = true
  podLogsError.value = ''
  podLogsContent.value = ''
  
  try {
    if (podLogsForm.value.follow) {
      // 实时日志
      await fetchPodStreamLogs()
    } else {
      // 静态日志
      await fetchPodStaticLogs()
    }
  } catch (e) {
    if (e.name !== 'AbortError') {
      podLogsError.value = e?.msg || e?.message || '获取日志失败'
    }
  } finally {
    if (!podLogsForm.value.follow) {
      loadingPodLogs.value = false
    }
  }
}

const fetchPodStaticLogs = async () => {
  const params = new URLSearchParams({
    namespace: selectedPodForAction.value.namespace,
    name: selectedPodForAction.value.name,
    container: podLogsForm.value.container
  })
  if (podLogsForm.value.tail != null) {
    params.set('tail', podLogsForm.value.tail)
  }
  
  podLogAbortController = new AbortController()
  const response = await fetch(`/api/v1/k8s/pod/container_logs?${params}`, {
    signal: podLogAbortController.signal,
    headers: getAuthHeaders()  // ✅ 使用认证头
  })
  
  if (!response.ok) throw new Error(`HTTP ${response.status}`)
  
  const res = await response.json()
  podLogsContent.value = res?.data?.log || '暂无日志'
  loadingPodLogs.value = false
}

const fetchPodStreamLogs = async () => {
  isStreamingPodLogs.value = true
  podLogAbortController = new AbortController()
  
  const params = new URLSearchParams({
    namespace: selectedPodForAction.value.namespace,
    name: selectedPodForAction.value.name,
    container: podLogsForm.value.container,
    follow: 'true'
  })
  if (podLogsForm.value.tail != null) {
    params.set('tail', podLogsForm.value.tail)
  }
  
  // 设置超时（如果有保持时长）
  if (podLogsForm.value.duration > 0) {
    podLogTimeoutId = setTimeout(() => {
      stopPodLogLoading()
    }, podLogsForm.value.duration * 1000)
  }
  
  const response = await fetch(`/api/v1/k8s/pod/container_log?${params}`, {
    signal: podLogAbortController.signal,
    headers: getAuthHeaders()  // ✅ 使用认证头
  })
  
  if (!response.ok) throw new Error(`HTTP ${response.status}`)
  
  const reader = response.body.getReader()
  const decoder = new TextDecoder('utf-8')
  
  while (true) {
    const { done, value } = await reader.read()
    if (done) break
    
    const chunk = decoder.decode(value, { stream: true })
    podLogsContent.value += chunk
    
    // 自动滚动到底部
    if (podLogsContentRef.value) {
      podLogsContentRef.value.scrollTop = podLogsContentRef.value.scrollHeight
    }
  }
  
  isStreamingPodLogs.value = false
  loadingPodLogs.value = false
}

const stopPodLogLoading = () => {
  if (podLogAbortController) {
    podLogAbortController.abort()
    podLogAbortController = null
  }
  if (podLogTimeoutId) {
    clearTimeout(podLogTimeoutId)
    podLogTimeoutId = null
  }
  isStreamingPodLogs.value = false
  loadingPodLogs.value = false
}

const clearPodLogs = () => {
  podLogsContent.value = ''
  podLogsError.value = ''
}

const closePodLogs = () => {
  stopPodLogLoading()
  showPodLogsModal.value = false
  podLogsContent.value = ''
  podLogsError.value = ''
}

// Pod 日志高亮处理
const highlightedPodLogs = computed(() => {
  if (!podLogsContent.value) {
    return '<span class="log-placeholder">暂无日志，请点击"获取日志"按钮</span>'
  }
  
  // 转义 HTML 特殊字符
  const escapeHtml = (str) => {
    return str
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;');
  }
  
  const escaped = escapeHtml(podLogsContent.value)
  
  // 按行处理
  return escaped.split('\n').map(line => {
    // 时间戳高亮 (ISO 格式或常见日志时间格式)
    let highlighted = line.replace(
      /(\d{4}-\d{2}-\d{2}[T\s]\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:?\d{2})?)/g,
      '<span class="log-timestamp">$1</span>'
    )
    
    // ERROR / FATAL / PANIC 高亮（红色）
    if (/\b(ERROR|FATAL|PANIC|error|fatal|panic|失败|错误)\b/i.test(line)) {
      highlighted = `<span class="log-error">${highlighted}</span>`
    }
    // WARN / WARNING 高亮（黄色）
    else if (/\b(WARN|WARNING|warn|warning|警告)\b/i.test(line)) {
      highlighted = `<span class="log-warn">${highlighted}</span>`
    }
    // INFO 高亮（蓝色）
    else if (/\b(INFO|info)\b/i.test(line)) {
      highlighted = `<span class="log-info">${highlighted}</span>`
    }
    // DEBUG 高亮（灰色）
    else if (/\b(DEBUG|debug)\b/i.test(line)) {
      highlighted = `<span class="log-debug">${highlighted}</span>`
    }
    // Pod 名称分隔符高亮（绿色）
    else if (/^=+ Pod:/.test(line)) {
      highlighted = `<span class="log-separator">${highlighted}</span>`
    }
    
    return highlighted
  }).join('\n')
})

// =========================
// Pod 操作：详情
// =========================
const openPodDetail = async (pod) => {
  selectedPodForAction.value = pod
  loadingPodDetail.value = true
  podDetailData.value = null
  showPodDetailModal.value = true
  try {
    const res = await podsApi.detail({
      namespace: pod.namespace,
      name: pod.name
    })
    podDetailData.value = res.code === 0 ? res.data : pod.raw
  } catch (e) {
    podDetailData.value = pod.raw
  } finally {
    loadingPodDetail.value = false
  }
}

// =========================
// Pod 操作：事件
// =========================
const openPodEvents = async (pod) => {
  selectedPodForAction.value = pod
  loadingPodEvents.value = true
  podEventsData.value = []
  showPodEventsModal.value = true
  try {
    const res = await podsApi.events({
      namespace: pod.namespace,
      name: pod.name
    })
    podEventsData.value = res.code === 0 ? (res.data?.events || res.data || []) : []
  } catch (e) {
    console.error('获取 Pod 事件失败:', e)
  } finally {
    loadingPodEvents.value = false
  }
}

// =========================
// Pod 操作：重启
// =========================
const restartPodFromList = async (pod) => {
  showPodMoreOptions.value = false
  if (!confirm(`确认重启（删除重建）Pod：${pod.namespace}/${pod.name} ?`)) return
  try {
    await podsApi.graceDelete({namespace: pod.namespace, name: pod.name})
    Message.success({ content: 'Pod 重启中...' })
    // 刷新 Pods 列表
    if (podsDeployment.value) {
      await viewPods(podsDeployment.value)
    }
    // 开启自动刷新
    autoRefresh.value = true
    setTimeout(() => { autoRefresh.value = false }, 15000)
  } catch (e) {
    Message.error({ content: e?.msg || e?.message || '重启失败' })
  }
}

// =========================
// Pod 操作：更新镜像
// =========================
const openPodUpdateImage = (pod) => {
  showPodMoreOptions.value = false
  selectedPodForAction.value = pod
  podContainerListForUpdate.value = pod.containers || []
  podUpdateImageForm.value = {
    container: podContainerListForUpdate.value[0] || '',
    image: ''
  }
  podUpdateImageError.value = ''
  showPodUpdateImageModal.value = true
}

const submitPodUpdateImage = async () => {
  if (!selectedPodForAction.value) return
  patchingPodImage.value = true
  podUpdateImageError.value = ''
  try {
    const res = await podsApi.updateImage({
      namespace: selectedPodForAction.value.namespace,
      name: selectedPodForAction.value.name,
      container: podUpdateImageForm.value.container,
      image: podUpdateImageForm.value.image
    })
    if (res.code === 0) {
      Message.success({ content: '镜像更新成功，Pod 重启中...' })
      showPodUpdateImageModal.value = false
      // 刷新 Pods 列表
      if (podsDeployment.value) {
        await viewPods(podsDeployment.value)
      }
      // 开启自动刷新
      autoRefresh.value = true
      setTimeout(() => { autoRefresh.value = false }, 15000)
    } else {
      podUpdateImageError.value = res.msg || '更新镜像失败'
    }
  } catch (e) {
    podUpdateImageError.value = e?.msg || e?.message || '更新镜像失败'
  } finally {
    patchingPodImage.value = false
  }
}

const closePodUpdateImage = () => {
  showPodUpdateImageModal.value = false
  podUpdateImageError.value = ''
}

// =========================
// Pod 操作：驱逐
// =========================
const evictPodFromList = async (pod) => {
  showPodMoreOptions.value = false
  if (!confirm(`确认驱逐 Pod：${pod.namespace}/${pod.name}？\n\n驱逐会受 PDB（Pod Disruption Budget）约束，相比直接删除更安全。`)) return
  try {
    await podsApi.evict({
      namespace: pod.namespace,
      podName: pod.name,
    })
    Message.success({ content: 'Pod 驱逐中...' })
    // 刷新 Pods 列表
    if (podsDeployment.value) {
      await viewPods(podsDeployment.value)
    }
    // 开启自动刷新
    autoRefresh.value = true
    setTimeout(() => { autoRefresh.value = false }, 15000)
  } catch (e) {
    Message.error({ content: e?.msg || e?.message || '驱逐失败' })
  }
}

// =========================
// Pod 操作：删除
// =========================
const deletePodFromList = (pod) => {
  selectedPodForAction.value = pod
  showPodDeleteModal.value = true
}

const confirmDeletePod = async () => {
  if (!selectedPodForAction.value) return
  deletingPod.value = true
  try {
    const res = await podsApi.graceDelete({
      namespace: selectedPodForAction.value.namespace,
      name: selectedPodForAction.value.name
    })
    if (res.code === 0) {
      Message.success({ content: 'Pod 已删除，Deployment 将自动重建' })
      showPodDeleteModal.value = false
      // 刷新 Pods 列表
      if (podsDeployment.value) {
        await viewPods(podsDeployment.value)
      }
      // 开启自动刷新
      autoRefresh.value = true
      setTimeout(() => { autoRefresh.value = false }, 15000)
    } else {
      Message.error({ content: res.msg || '删除失败' })
    }
  } catch (e) {
    Message.error({ content: e?.msg || '删除失败' })
  } finally {
    deletingPod.value = false
  }
}

// =========================
// Pod 操作：强制删除
// =========================
const forceDeletePodFromList = (pod) => {
  selectedPodForAction.value = pod
  showPodForceDeleteModal.value = true
}

const confirmForceDeletePod = async () => {
  if (!selectedPodForAction.value) return
  deletingPod.value = true
  try {
    const res = await podsApi.forceDelete({
      namespace: selectedPodForAction.value.namespace,
      name: selectedPodForAction.value.name
    })
    if (res.code === 0) {
      Message.success({ content: 'Pod 已强制删除，Deployment 将自动重建' })
      showPodForceDeleteModal.value = false
      // 刷新 Pods 列表
      if (podsDeployment.value) {
        await viewPods(podsDeployment.value)
      }
      // 开启自动刷新
      autoRefresh.value = true
      setTimeout(() => { autoRefresh.value = false }, 15000)
    } else {
      Message.error({ content: res.msg || '强制删除失败' })
    }
  } catch (e) {
    Message.error({ content: e?.msg || '强制删除失败' })
  } finally {
    deletingPod.value = false
  }
}

// =========================
// Deployment 操作：查看日志（增强版 - 支持行数选择和实时日志）
// =========================
const viewDeploymentLogs = async (deployment) => {
  showMoreOptions.value = false
  selectedDeploymentForLogs.value = deployment
  deploymentLogsContent.value = ''
  deploymentLogsError.value = ''
  deploymentPodsList.value = []
  deploymentContainerList.value = []
  deploymentLogsForm.value = {
    selectedPod: '',
    container: '',
    tail: 100,
    follow: false
  }
  showDeploymentLogsModal.value = true

  // 加载 Pod 列表
  try {
    const podsRes = await deploymentsApi.pods({
      namespace: deployment.namespace,
      name: deployment.name
    })
    const pods = podsRes.data?.pods || podsRes.data || []
    deploymentPodsList.value = (Array.isArray(pods) ? pods : []).map(pod => ({
      name: pod.name || pod.metadata?.name,
      containers: pod.containers || pod.spec?.containers?.map(c => c.name) || []
    }))
  } catch (e) {
    deploymentLogsError.value = '获取 Pod 列表失败'
  }
}

const onPodChange = () => {
  // 当选择 Pod 时，更新容器列表
  const pod = deploymentPodsList.value.find(p => p.name === deploymentLogsForm.value.selectedPod)
  if (pod) {
    deploymentContainerList.value = pod.containers || []
    // ✅ 单容器自动选中，多容器清空（让用户选择）
    if (deploymentContainerList.value.length === 1) {
      deploymentLogsForm.value.container = deploymentContainerList.value[0]
    } else if (deploymentContainerList.value.length > 1) {
      deploymentLogsForm.value.container = ''  // 多容器时清空，强制用户选择
    } else {
      deploymentLogsForm.value.container = ''  // 无容器时清空
    }
  } else {
    deploymentContainerList.value = []
    deploymentLogsForm.value.container = ''
  }
  deploymentLogsContent.value = ''
}

const fetchDeploymentLogs = async () => {
  if (!selectedDeploymentForLogs.value) return

  // 如果选择了单个 Pod，需要验证容器
  if (deploymentLogsForm.value.selectedPod) {
    let container = deploymentLogsForm.value.container || ''
    
    // 如果只有一个容器，自动使用
    if (deploymentContainerList.value.length === 1) {
      container = deploymentContainerList.value[0]
      deploymentLogsForm.value.container = container  // 更新表单值
    }
    
    // 多容器时必须选择容器
    if (deploymentContainerList.value.length > 1 && !container) {
      deploymentLogsError.value = '请选择容器'
      return
    }
    
    // 确保容器名不为空
    if (!container) {
      deploymentLogsError.value = '无法确定容器名称'
      return
    }
  }

  stopDeploymentLogStream()
  loadingDeploymentLogs.value = true
  deploymentLogsError.value = ''
  deploymentLogsContent.value = ''

  try {
    if (deploymentLogsForm.value.follow) {
      // 实时日志
      await fetchDeploymentStreamLogs()
    } else {
      // 单次日志
      await fetchDeploymentStaticLogs()
    }
  } catch (e) {
    if (e.name !== 'AbortError') {
      deploymentLogsError.value = e?.msg || e?.message || '获取日志失败'
    }
  } finally {
    if (!deploymentLogsForm.value.follow) {
      loadingDeploymentLogs.value = false
    }
  }
}

const fetchDeploymentStaticLogs = async () => {
  if (deploymentLogsForm.value.selectedPod) {
    // 获取单个 Pod 的日志
    const params = new URLSearchParams({
      namespace: selectedDeploymentForLogs.value.namespace,
      name: deploymentLogsForm.value.selectedPod,
      container: deploymentLogsForm.value.container
    })
    if (deploymentLogsForm.value.tail != null) {
      params.set('tail', deploymentLogsForm.value.tail)
    }

    deploymentLogAbortController = new AbortController()
    const response = await fetch(`/api/v1/k8s/pod/container_logs?${params}`, {
      signal: deploymentLogAbortController.signal,
      headers: getAuthHeaders()  // ✅ 使用认证头
    })

    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    
    const res = await response.json()
    deploymentLogsContent.value = res?.data?.log || '暂无日志'
  } else {
    // 获取所有 Pod 的日志（汇总）
    const pods = deploymentPodsList.value.slice(0, 5)  // 最多5个
    const logsArray = []

    for (const pod of pods) {
      const container = pod.containers[0] || ''
      const params = new URLSearchParams({
        namespace: selectedDeploymentForLogs.value.namespace,
        name: pod.name,
        container
      })
      if (deploymentLogsForm.value.tail != null) {
        params.set('tail', deploymentLogsForm.value.tail)
      }

      try {
        const response = await fetch(`/api/v1/k8s/pod/container_logs?${params}`, {
          headers: getAuthHeaders()  // ✅ 使用认证头
        })
        if (response.ok) {
          const res = await response.json()
          logsArray.push(`\n========== Pod: ${pod.name} | 容器: ${container} ==========\n${res?.data?.log || '（无日志）'}`)
        }
      } catch (e) {
        logsArray.push(`\n========== Pod: ${pod.name} ==========\n获取日志失败: ${e.message}`)
      }
    }

    deploymentLogsContent.value = logsArray.join('\n')
  }
}

const fetchDeploymentStreamLogs = async () => {
  if (!deploymentLogsForm.value.selectedPod) {
    deploymentLogsError.value = '实时日志需要选择单个 Pod'
    return
  }

  isStreamingDeploymentLogs.value = true
  deploymentLogAbortController = new AbortController()

  const params = new URLSearchParams({
    namespace: selectedDeploymentForLogs.value.namespace,
    name: deploymentLogsForm.value.selectedPod,
    container: deploymentLogsForm.value.container,
    follow: 'true'
  })
  if (deploymentLogsForm.value.tail != null) {
    params.set('tail', deploymentLogsForm.value.tail)
  }

  const response = await fetch(`/api/v1/k8s/pod/container_log?${params}`, {
    signal: deploymentLogAbortController.signal,
    headers: getAuthHeaders()  // ✅ 使用认证头
  })

  if (!response.ok) throw new Error(`HTTP ${response.status}`)

  const reader = response.body.getReader()
  const decoder = new TextDecoder('utf-8')

  while (true) {
    const { done, value } = await reader.read()
    if (done) break

    const chunk = decoder.decode(value, { stream: true })
    deploymentLogsContent.value += chunk

    // 自动滚动到底部
    if (deploymentLogsContentRef.value) {
      deploymentLogsContentRef.value.scrollTop = deploymentLogsContentRef.value.scrollHeight
    }
  }

  isStreamingDeploymentLogs.value = false
  loadingDeploymentLogs.value = false
}

const stopDeploymentLogStream = () => {
  if (deploymentLogAbortController) {
    deploymentLogAbortController.abort()
    deploymentLogAbortController = null
  }
  isStreamingDeploymentLogs.value = false
  loadingDeploymentLogs.value = false
}

const clearDeploymentLogs = () => {
  deploymentLogsContent.value = ''
}

const closeDeploymentLogs = () => {
  stopDeploymentLogStream()
  showDeploymentLogsModal.value = false
  deploymentLogsContent.value = ''
  deploymentLogsError.value = ''
  deploymentPodsList.value = []
  deploymentContainerList.value = []
}

// Deployment 日志高亮处理
const highlightedDeploymentLogs = computed(() => {
  if (!deploymentLogsContent.value) {
    return '<span class="log-placeholder">暂无日志，请点击"获取日志"按钮</span>'
  }
  
  // 转义 HTML 特殊字符
  const escapeHtml = (str) => {
    return str
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;');
  }
  
  const escaped = escapeHtml(deploymentLogsContent.value)
  
  // 按行处理
  return escaped.split('\n').map(line => {
    // 时间戳高亮 (ISO 格式或常见日志时间格式)
    let highlighted = line.replace(
      /(\d{4}-\d{2}-\d{2}[T\s]\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:?\d{2})?)/g,
      '<span class="log-timestamp">$1</span>'
    )
    
    // ERROR / FATAL / PANIC 高亮（红色）
    if (/\b(ERROR|FATAL|PANIC|error|fatal|panic|失败|错误)\b/i.test(line)) {
      highlighted = `<span class="log-error">${highlighted}</span>`
    }
    // WARN / WARNING 高亮（黄色）
    else if (/\b(WARN|WARNING|warn|warning|警告)\b/i.test(line)) {
      highlighted = `<span class="log-warn">${highlighted}</span>`
    }
    // INFO 高亮（蓝色）
    else if (/\b(INFO|info)\b/i.test(line)) {
      highlighted = `<span class="log-info">${highlighted}</span>`
    }
    // DEBUG 高亮（灰色）
    else if (/\b(DEBUG|debug)\b/i.test(line)) {
      highlighted = `<span class="log-debug">${highlighted}</span>`
    }
    // Pod 名称分隔符高亮（青色 - 更醒目的绿色）
    else if (/^=+ Pod:/.test(line)) {
      highlighted = `<span class="log-separator">${highlighted}</span>`
    }
    
    return highlighted
  }).join('\n')
})

// =========================
// 扩缩容
// =========================

// 增加副本数（+1）
const increaseReplicas = async (deployment) => {
  await updateReplicaCount(deployment, deployment.desiredReplicas + 1)
}

// 减少副本数（-1）
const decreaseReplicas = async (deployment) => {
  if (deployment.desiredReplicas <= 0) return
  await updateReplicaCount(deployment, deployment.desiredReplicas - 1)
}

// 停服（将副本数调为0）
const stopService = async (deployment) => {
  if (deployment.desiredReplicas === 0) return
  
  if (!confirm(`确认停服 ${deployment.namespace}/${deployment.name}？\n\n将副本数调整为 0，所有 Pod 将被终止。`)) {
    return
  }
  
  await updateReplicaCount(deployment, 0)
}

// 更新副本数（通用方法）
const updateReplicaCount = async (deployment, newReplicas) => {
  if (scalingMap.value[deployment.name]) return // 防止重复点击
  
  scalingMap.value[deployment.name] = true
  
  try {
    const res = await deploymentsApi.scale({
      namespace: deployment.namespace,
      name: deployment.name,
      scale_num: newReplicas  // 后端字段名是 scale_num
    })
    if (res.code === 0) {
      const action = newReplicas === 0 ? '已停服' : `副本数已更新为 ${newReplicas}`
      Message.success({ content: action })
      
      // 立即刷新数据，确保显示最新状态
      await fetchDeployments()
      
      // 开启自动刷新追踪状态变化（持续监控直到稳定）
      autoRefresh.value = true
      setTimeout(() => { 
        autoRefresh.value = false 
      }, 15000)  // 延长到 15 秒，确保状态稳定
    } else {
      Message.error({ content: res.msg || '扩缩容失败' })
    }
  } catch (e) {
    Message.error({ content: '扩缩容失败' })
  } finally {
    // 延迟解除锁定，防止连续快速点击
    setTimeout(() => {
      scalingMap.value[deployment.name] = false
    }, 500)
  }
}

// 内联编辑 - 开始编辑镜像
const startInlineImage = (deployment) => {
  inlineEdit.value = {
    key: `image-${deployment.name}`,
    value: deployment.image,
    original: deployment.image,
    deployment,
    container: deployment.containers?.[0] || ''  // containers 是字符串数组
  }
}

// 内联编辑 - 保存镜像
const saveInlineImage = async (deployment) => {
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

Deployment: ${deployment.namespace}/${deployment.name}
容器: ${container}
旧镜像: ${oldImage}
新镜像: ${newImage}

此操作将触发滚动更新，请确认！`)) {
    return
  }
  
  try {
    const res = await deploymentsApi.updateImage({
      namespace: deployment.namespace,
      name: deployment.name,
      container: container,
      image: newImage
    })
    if (res.code === 0) {
      Message.success({ content: '镜像已更新，正在滚动更新...' })
      cancelInlineEdit()
      autoRefresh.value = true
      setTimeout(() => { autoRefresh.value = false }, 15000)
      await fetchDeployments()
    } else {
      Message.error({ content: res.msg || '更新镜像失败' })
    }
  } catch (e) {
    Message.error({ content: '更新镜像失败' })
  }
}

// 内联编辑 - 取消编辑
const cancelInlineEdit = () => {
  inlineEdit.value = { key: '', value: null, original: null }
}

// 内联编辑副本数 - 开始编辑
const startInlineReplicas = (deployment) => {
  if (scalingMap.value[deployment.name]) return // 防止在扩缩容中编辑
  
  inlineEdit.value = {
    key: `replicas-${deployment.name}`,
    value: deployment.desiredReplicas,
    original: deployment.desiredReplicas,
    deployment
  }
}

// 内联编辑副本数 - 保存
const saveInlineReplicas = async (deployment) => {
  const newReplicas = parseInt(inlineEdit.value.value, 10)
  const oldReplicas = inlineEdit.value.original
  
  // 验证输入
  if (isNaN(newReplicas) || newReplicas < 0) {
    Message.error({ content: '副本数必须是 0 或正整数' })
    cancelInlineEdit()
    return
  }
  
  // 如果没有变化，直接取消编辑
  if (newReplicas === oldReplicas) {
    cancelInlineEdit()
    return
  }
  
  // 调用通用的更新副本数方法
  cancelInlineEdit()
  await updateReplicaCount(deployment, newReplicas)
}


const openScaleModal = (deployment) => {
  showMoreOptions.value = false
  scaleForm.value = { namespace: deployment.namespace, name: deployment.name, replicas: deployment.desiredReplicas }
  showScaleModal.value = true
}

const scaleDeployment = async () => {
  scaling.value = true
  try {
    const res = await deploymentsApi.scale({
      namespace: scaleForm.value.namespace,
      name: scaleForm.value.name,
      scale_num: scaleForm.value.replicas  // 后端字段名是 scale_num
    })
    if (res.code === 0) {
      Message.success({ content: '扩缩容成功' })
      showScaleModal.value = false
      
      // 立即刷新数据
      await fetchDeployments()
      
      // 开启自动刷新追踪状态变化
      autoRefresh.value = true
      setTimeout(() => { autoRefresh.value = false }, 15000)
    } else {
      Message.error({ content: res.msg || '扩缩容失败' })
    }
  } catch (e) {
    Message.error({ content: '扩缩容失败' })
  } finally {
    scaling.value = false
  }
}

// =========================
// 更新镜像
// =========================
const openUpdateImage = (deployment) => {
  showMoreOptions.value = false
  updateImageForm.value = {
    namespace: deployment.namespace,
    name: deployment.name,
    container: deployment.containers?.[0] || '',  // containers 是字符串数组
    image: ''
  }
  containerList.value = deployment.containers || []  // 直接使用字符串数组
  if (containerList.value.length === 1) {
    updateImageForm.value.container = containerList.value[0]
  }
  showUpdateImageModal.value = true
}

const submitUpdateImage = async () => {
  // 二次确认 - 镜像更新是高危操作
  if (!confirm(`⚠️ 确认更新镜像？

Deployment: ${updateImageForm.value.namespace}/${updateImageForm.value.name}
容器: ${updateImageForm.value.container}
新镜像: ${updateImageForm.value.image}

此操作将触发滚动更新，请确认！`)) {
    return
  }
  
  updatingImage.value = true
  try {
    const res = await deploymentsApi.updateImage(updateImageForm.value)
    if (res.code === 0) {
      Message.success({ content: '镜像更新成功，正在滚动更新...' })
      showUpdateImageModal.value = false
      
      // 立即刷新数据
      await fetchDeployments()
      
      // 开启自动刷新追踪滚动更新状态
      autoRefresh.value = true
      setTimeout(() => { autoRefresh.value = false }, 15000)
    } else {
      Message.error({ content: res.msg || '更新镜像失败' })
    }
  } catch (e) {
    Message.error({ content: '更新镜像失败' })
  } finally {
    updatingImage.value = false
  }
}

// =========================
// 重启
// =========================
const restartDeployment = async (deployment) => {
  showMoreOptions.value = false
  
  // 二次确认 - 重启是高危操作
  if (!confirm(`⚠️ 确认重启 Deployment？

Deployment: ${deployment.namespace}/${deployment.name}
副本数: ${deployment.desiredReplicas}

此操作将滚动重启所有 Pod，请确认！`)) {
    return
  }
  
  try {
    const res = await deploymentsApi.restart({ namespace: deployment.namespace, name: deployment.name })
    if (res.code === 0) {
      Message.success({ content: '重启成功，正在滚动重启...' })
      autoRefresh.value = true
      setTimeout(() => { autoRefresh.value = false }, 15000)
      await fetchDeployments()
    } else {
      Message.error({ content: res.msg || '重启失败' })
    }
  } catch (e) {
    Message.error({ content: '重启失败' })
  }
}

// =========================
// 查看版本记录
// =========================
const viewHistory = async (deployment) => {
  showMoreOptions.value = false
  historyDeployment.value = deployment
  loadingHistory.value = true
  historyList.value = []
  showHistoryModal.value = true
  
  try {
    const res = await deploymentsApi.history({ 
      namespace: deployment.namespace, 
      name: deployment.name 
    })
    const list = res.code === 0 ? (res.data?.list || res.data || []) : []
    historyList.value = list.map(rs => ({
      name: rs.metadata?.name || rs.name,
      revision: rs.metadata?.annotations?.['deployment.kubernetes.io/revision'] || '-',
      replicas: rs.spec?.replicas || 0,
      createdAt: fmtTime(rs.metadata?.creationTimestamp),
      raw: rs
    })).sort((a, b) => {
      // 按版本号降序排列（最新的在前）
      const revA = parseInt(a.revision) || 0
      const revB = parseInt(b.revision) || 0
      return revB - revA
    })
  } catch (e) {
    console.error('获取版本历史失败:', e)
    Message.error({ content: '获取版本历史失败' })
  } finally {
    loadingHistory.value = false
  }
}

// 从版本记录弹窗回滚到指定版本
const rollbackToVersion = async (history) => {
  if (!historyDeployment.value || !history.name) return
  
  const confirmed = confirm(`确认回滚到版本 ${history.revision} 吗？

ReplicaSet: ${history.name}
副本数: ${history.replicas}
创建时间: ${history.createdAt}`)
  if (!confirmed) return
  
  rollingBack.value = true
  try {
    const res = await deploymentsApi.rollback({
      namespace: historyDeployment.value.namespace,
      name: historyDeployment.value.name,
      replica_set: history.name
    })
    
    if (res.code === 0) {
      Message.success({ content: `回滚到版本 ${history.revision} 成功` })
      showHistoryModal.value = false
      await fetchDeployments()
      // 开启自动刷新
      autoRefresh.value = true
      setTimeout(() => { autoRefresh.value = false }, 15000)
    } else {
      Message.error({ content: res.msg || '回滚失败' })
    }
  } catch (e) {
    console.error('回滚失败:', e)
    Message.error({ content: e?.msg || e?.message || '回滚失败' })
  } finally {
    rollingBack.value = false
  }
}

// =========================
// 回滚
// =========================
const openRollback = async (deployment) => {
  showMoreOptions.value = false
  rollbackForm.value = { namespace: deployment.namespace, name: deployment.name, replica_set: '' }
  loadingHistory.value = true
  historyList.value = []
  showRollbackModal.value = true
  try {
    const res = await deploymentsApi.history({ namespace: deployment.namespace, name: deployment.name })
    // 后端返回 res.data.list
    const list = res.code === 0 ? (res.data?.list || res.data || []) : []
    historyList.value = list.map(rs => ({
      name: rs.metadata?.name || rs.name,
      revision: rs.metadata?.annotations?.['deployment.kubernetes.io/revision'] || '-',
      replicas: rs.spec?.replicas || 0,
      createdAt: fmtTime(rs.metadata?.creationTimestamp),
      raw: rs
    }))
  } catch (e) {
    console.error('获取历史版本失败:', e)
  } finally {
    loadingHistory.value = false
  }
}

const submitRollback = async () => {
  // 二次确认 - 回滚是高危操作
  const selectedRS = historyList.value.find(h => h.name === rollbackForm.value.replica_set)
  if (!confirm(`⚠️ 确认回滚 Deployment？

Deployment: ${rollbackForm.value.namespace}/${rollbackForm.value.name}
回滚到版本: ${selectedRS?.revision || '-'}
ReplicaSet: ${rollbackForm.value.replica_set}

此操作将触发滚动更新到历史版本，请确认！`)) {
    return
  }
  
  rollingBack.value = true
  try {
    const res = await deploymentsApi.rollback(rollbackForm.value)
    if (res.code === 0) {
      Message.success({ content: '回滚成功，正在滚动更新...' })
      showRollbackModal.value = false
      autoRefresh.value = true
      setTimeout(() => { autoRefresh.value = false }, 15000)
      await fetchDeployments()
    } else {
      Message.error({ content: res.msg || '回滚失败' })
    }
  } catch (e) {
    Message.error({ content: '回滚失败' })
  } finally {
    rollingBack.value = false
  }
}

// =========================
// 删除
// =========================
const deleteDeployment = (deployment) => {
  showMoreOptions.value = false
  deploymentToDelete.value = deployment.name
  deploymentNamespaceToDelete.value = deployment.namespace
  showDeleteModal.value = true
}

const confirmDelete = async () => {
  deleting.value = true
  try {
    const res = await deploymentsApi.delete({ namespace: deploymentNamespaceToDelete.value, name: deploymentToDelete.value })
    if (res.code === 0) {
      Message.success({ content: '删除成功' })
      showDeleteModal.value = false
      await fetchDeployments()
    } else {
      Message.error({ content: res.msg || '删除失败' })
    }
  } catch (e) {
    Message.error({ content: '删除失败' })
  } finally {
    deleting.value = false
  }
}

// =========================
// 创建
// =========================
// 标签管理
const addLabel = () => {
  deploymentForm.value.labels.push({ key: '', value: '' })
}

const removeLabel = (index) => {
  if (deploymentForm.value.labels.length > 1) {
    deploymentForm.value.labels.splice(index, 1)
  }
}

// 端口管理
const addServicePort = () => {
  deploymentForm.value.servicePorts.push({ port: 80, targetPort: 80, nodePort: null })
}

const removeServicePort = (index) => {
  if (deploymentForm.value.servicePorts.length > 1) {
    deploymentForm.value.servicePorts.splice(index, 1)
  }
}

// 节点选择器管理
const addNodeSelector = () => {
  deploymentForm.value.nodeSelector.push({ key: '', value: '' })
}

const removeNodeSelector = (index) => {
  deploymentForm.value.nodeSelector.splice(index, 1)
}

// 容忍管理
const addToleration = () => {
  deploymentForm.value.tolerations.push({ key: '', operator: 'Equal', value: '', effect: '', tolerationSeconds: null })
}

const removeToleration = (index) => {
  deploymentForm.value.tolerations.splice(index, 1)
}

// 节点亲和性规则管理
const addNodeAffinityRule = () => {
  deploymentForm.value.nodeAffinityRules.push({ key: '', operator: 'In', values: '', required: false, weight: 50 })
}

const removeNodeAffinityRule = (index) => {
  deploymentForm.value.nodeAffinityRules.splice(index, 1)
}

// 拓扑分布约束管理
const addTopologySpread = () => {
  deploymentForm.value.topologySpreadConfigs.push({ topologyKey: 'kubernetes.io/hostname', maxSkew: 1, whenUnsatisfiable: 'ScheduleAnyway' })
}

const removeTopologySpread = (index) => {
  deploymentForm.value.topologySpreadConfigs.splice(index, 1)
}

const createDeployment = async () => {
  try {
    // 验证标签
    const validLabels = deploymentForm.value.labels.filter(label => label.key && label.value)
    
    if (validLabels.length === 0) {
      Message.error({ content: '至少需要一个有效的标签（键值对）' })
      return
    }
    
    const data = {
      namespace: deploymentForm.value.namespace,
      name: deploymentForm.value.name,
      container_image: deploymentForm.value.image,  // 注意：后端字段是 container_image
      replicas: parseInt(deploymentForm.value.replicas),
      labels: validLabels,  // 直接发送标签数组，后端会解析
      port_mappings: [],  // 暂时空数组，后续可扩展
      variables: [],  // 暂时空数组
      is_create_service: deploymentForm.value.createService && deploymentForm.value.serviceType !== 'None',
      service_type: deploymentForm.value.createService ? deploymentForm.value.serviceType : '',
      service_name: deploymentForm.value.createService ? (deploymentForm.value.serviceName || deploymentForm.value.name) : '',
      run_as_privileged: false,
      is_readiness_enable: false,
      is_liveness_enable: false,
      readiness_probe: {},
      liveness_probe: {},
      // 调度规则
      scheduling_policy: deploymentForm.value.schedulingPolicy || 'default'
    }
    
    // 添加资源配置 (Resources)
    if (deploymentForm.value.resources.cpuRequest || deploymentForm.value.resources.cpuLimit ||
        deploymentForm.value.resources.memoryRequest || deploymentForm.value.resources.memoryLimit) {
      data.resources = {
        cpu_request: deploymentForm.value.resources.cpuRequest || null,
        cpu_limit: deploymentForm.value.resources.cpuLimit || null,
        memory_request: deploymentForm.value.resources.memoryRequest || null,
        memory_limit: deploymentForm.value.resources.memoryLimit || null
      }
    }
    
    // 添加探针配置 (Probes)
    if (deploymentForm.value.probes.enableLiveness || deploymentForm.value.probes.enableReadiness || deploymentForm.value.probes.enableStartup) {
      data.probes = {
        enable_liveness: deploymentForm.value.probes.enableLiveness,
        enable_readiness: deploymentForm.value.probes.enableReadiness,
        enable_startup: deploymentForm.value.probes.enableStartup
      }
      
      // Liveness Probe
      if (deploymentForm.value.probes.enableLiveness) {
        data.probes.liveness_probe = buildProbeData(deploymentForm.value.probes.livenessProbe)
      }
      
      // Readiness Probe
      if (deploymentForm.value.probes.enableReadiness) {
        data.probes.readiness_probe = buildProbeData(deploymentForm.value.probes.readinessProbe)
      }
      
      // Startup Probe
      if (deploymentForm.value.probes.enableStartup) {
        data.probes.startup_probe = buildProbeData(deploymentForm.value.probes.startupProbe)
      }
    }
    
    // 添加节点选择器（custom 模式）
    if (deploymentForm.value.schedulingPolicy === 'custom') {
      const validSelectors = deploymentForm.value.nodeSelector.filter(s => s.key && s.value)
      if (validSelectors.length > 0) {
        const nodeSelectorMap = {}
        validSelectors.forEach(s => {
          nodeSelectorMap[s.key] = s.value
        })
        data.node_selector = nodeSelectorMap
      }
      
      // 添加容忍配置
      const validTolerations = deploymentForm.value.tolerations.filter(t => t.key)
      if (validTolerations.length > 0) {
        data.tolerations = validTolerations.map(t => ({
          key: t.key,
          operator: t.operator || 'Equal',
          value: t.operator === 'Equal' ? t.value : undefined,
          effect: t.effect || undefined,
          tolerationSeconds: t.effect === 'NoExecute' && t.tolerationSeconds ? t.tolerationSeconds : undefined
        }))
      }
      
      // 添加节点亲和性规则
      const validAffinityRules = deploymentForm.value.nodeAffinityRules.filter(r => r.key)
      if (validAffinityRules.length > 0) {
        data.node_affinity_rules = validAffinityRules.map(r => ({
          key: r.key,
          operator: r.operator || 'In',
          values: r.values || '',
          required: r.required || false,
          weight: r.required ? 0 : (r.weight || 50)
        }))
      }
      
      // 添加拓扑分布约束
      const validTopologySpread = deploymentForm.value.topologySpreadConfigs.filter(s => s.topologyKey)
      if (validTopologySpread.length > 0) {
        data.topology_spread_configs = validTopologySpread.map(s => ({
          topology_key: s.topologyKey,
          max_skew: s.maxSkew || 1,
          when_unsatisfiable: s.whenUnsatisfiable || 'ScheduleAnyway'
        }))
      }
    }
    
    // 添加端口配置（如果创建 Service）
    if (data.is_create_service && deploymentForm.value.serviceType !== 'None') {
      data.port_mappings = deploymentForm.value.servicePorts.map(p => ({
        port: parseInt(p.port),
        target_port: parseInt(p.targetPort),
        protocol: 'TCP'
      })).filter(p => p.port && p.target_port)
      
      if (data.port_mappings.length === 0) {
        Message.error({ content: '至少需要配置一个有效的端口映射' })
        return
      }
    }
    
    const res = await deploymentsApi.create(data)
    if (res.code === 0) {
      let successMsg = '创建成功'
      if (deploymentForm.value.createService && deploymentForm.value.serviceType !== 'None') {
        successMsg += `（包含 ${deploymentForm.value.serviceType} Service）`
      }
      Message.success({ content: successMsg, duration: 2000 })
      showCreateModal.value = false
      resetForm()
      await fetchDeployments()
      // 创建后启动自动刷新 15 秒，观察 Pod 启动状态
      autoRefresh.value = true
      setTimeout(() => { autoRefresh.value = false }, 15000)
    } else {
      const errorMsg = res.details ? `${res.msg}: ${res.details}` : (res.msg || '创建失败')
      Message.error({ content: errorMsg, duration: 5000 })
    }
  } catch (e) {
    console.error('创建失败:', e)
    const errorMsg = e?.details ? `${e?.msg || '创建失败'}: ${e.details}` : (e?.msg || e?.message || '创建失败')
    Message.error({ content: errorMsg, duration: 5000 })
  }
}

const resetForm = () => {
  deploymentForm.value = { 
    name: '', 
    namespace: 'default', 
    replicas: 1, 
    image: '', 
    updateStrategy: 'RollingUpdate', 
    labels: [{ key: 'app', value: '' }],
    createService: false,
    serviceType: 'ClusterIP',
    serviceName: '',
    servicePorts: [{ port: 80, targetPort: 80, nodePort: null }],
    // 调度规则
    schedulingPolicy: 'default',
    nodeSelector: [],
    tolerations: [],
    nodeAffinityRules: [],
    topologySpreadConfigs: [],
    // 资源配置
    resources: {
      cpuRequest: '',
      cpuLimit: '',
      memoryRequest: '',
      memoryLimit: ''
    },
    // 探针配置
    probes: {
      enableLiveness: false,
      livenessProbe: {
        type: 'HTTP',
        port: 8080,
        path: '/healthz',
        command: '',
        initialDelaySeconds: 10,
        periodSeconds: 10,
        timeoutSeconds: 1,
        successThreshold: 1,
        failureThreshold: 3
      },
      enableReadiness: false,
      readinessProbe: {
        type: 'HTTP',
        port: 8080,
        path: '/ready',
        command: '',
        initialDelaySeconds: 5,
        periodSeconds: 10,
        timeoutSeconds: 1,
        successThreshold: 1,
        failureThreshold: 3
      },
      enableStartup: false,
      startupProbe: {
        type: 'HTTP',
        port: 8080,
        path: '/startup',
        command: '',
        initialDelaySeconds: 0,
        periodSeconds: 10,
        timeoutSeconds: 1,
        successThreshold: 1,
        failureThreshold: 30
      }
    }
  }
  showNamespaceInput.value = false
  newNamespace.value = ''
  // 重置 YAML 相关状态
  createMode.value = 'form'
  yamlContent.value = ''
  yamlError.value = ''
  // 重置折叠面板
  expandedSections.value = {
    resources: false,
    probes: false
  }
}

// 构建探针数据（转换为后端格式）
const buildProbeData = (probe) => {
  const data = {
    type: probe.type,
    initial_delay_seconds: probe.initialDelaySeconds || 0,
    period_seconds: probe.periodSeconds || 10,
    timeout_seconds: probe.timeoutSeconds || 1,
    success_threshold: probe.successThreshold || 1,
    failure_threshold: probe.failureThreshold || 3
  }
  
  if (probe.type === 'HTTP') {
    data.protocol = 'HTTP'
    data.port = probe.port || 8080
    data.path = probe.path || '/'
  } else if (probe.type === 'TCP') {
    data.protocol = 'TCP'
    data.port = probe.port || 8080
  } else if (probe.type === 'Command') {
    data.command = probe.command || ''
  }
  
  return data
}

// =========================
// YAML 创建相关功能
// =========================

// 加载 Deployment YAML 模板
const loadDeploymentYamlTemplate = () => {
  yamlContent.value = `# 支持多资源 YAML 创建（用 --- 分隔）
# 示例：PVC + Deployment 组合创建
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: nginx-pvc
  namespace: default
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: nfs-client
  resources:
    requests:
      storage: 1Gi
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-with-storage
  namespace: default
  labels:
    app: nginx-with-storage
spec:
  replicas: 2
  selector:
    matchLabels:
      app: nginx-with-storage
  template:
    metadata:
      labels:
        app: nginx-with-storage
    spec:
      containers:
      - name: nginx
        image: nginx:1.26
        ports:
        - containerPort: 80
        volumeMounts:
        - name: nginx-data
          mountPath: /usr/share/nginx/html
        resources:
          requests:
            cpu: "100m"
            memory: "128Mi"
          limits:
            cpu: "500m"
            memory: "256Mi"
        startupProbe:
          httpGet:
            path: /
            port: 80
          failureThreshold: 30
          periodSeconds: 5
        readinessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: nginx-data
        persistentVolumeClaim:
          claimName: nginx-pvc`
  yamlError.value = ''
  Message.success({ content: '已加载多资源 YAML 模板（PVC + Deployment），请修改后创建' })
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
  Message.success({ content: 'YAML 内容已重置' })
}

// 从 YAML 创建 Deployment
const createDeploymentFromYaml = async () => {
  if (!yamlContent.value.trim()) {
    Message.error({ content: '请输入 YAML 内容' })
    return
  }
  
  // 简单验证 YAML 格式
  try {
    if (!yamlContent.value.includes('kind: Deployment')) {
      yamlError.value = 'YAML 中必须包含 "kind: Deployment"'
      return
    }
    if (!yamlContent.value.includes('apiVersion: apps/v1')) {
      yamlError.value = 'YAML 中必须包含 "apiVersion: apps/v1"'
      return
    }
    
    yamlError.value = ''
  } catch (e) {
    yamlError.value = `YAML 格式错误: ${e.message}`
    return
  }
  
  try {
    const res = await deploymentsApi.createFromYaml({ yaml: yamlContent.value })
    if (res.code === 0) {
      Message.success({ content: 'Deployment 创建成功' })
      showCreateModal.value = false
      resetForm()
      await fetchDeployments()
      autoRefresh.value = true
      setTimeout(() => { autoRefresh.value = false }, 15000)
    } else {
      // 组合 msg 和 details 展示完整错误信息
      const errorMsg = res.details ? `${res.msg}: ${res.details}` : (res.msg || '创建失败')
      Message.error({ content: errorMsg, duration: 5000 })
      yamlError.value = errorMsg
    }
  } catch (e) {
    // 处理异常情况，同样组合 msg 和 details
    const errorMsg = e?.details ? `${e?.msg || '创建失败'}: ${e.details}` : (e?.msg || e?.message || '创建失败')
    Message.error({ content: errorMsg, duration: 5000 })
    yamlError.value = errorMsg
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
      deploymentForm.value.namespace = newNamespace.value.trim()
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

// 编辑（保留但简化）
const editDeployment = (deployment) => {
  editForm.value = {
    name: deployment.name,
    namespace: deployment.namespace,
    replicas: deployment.desiredReplicas,
    image: deployment.image,
    container: deployment.containers?.[0]?.name || '',
    updateStrategy: deployment.updateStrategy
  }
  editSelectorInput.value = Object.entries(deployment.selector || {}).map(([k, v]) => `${k}=${v}`).join(', ')
  showEditModal.value = true
}

const updateDeployment = async () => {
  try {
    const res = await deploymentsApi.updateImage({
      namespace: editForm.value.namespace,
      name: editForm.value.name,
      container: editForm.value.container,
      image: editForm.value.image
    })
    if (res.code === 0) {
      Message.success({ content: '更新成功' })
      showEditModal.value = false
      await fetchDeployments()
    } else {
      Message.error({ content: res.msg || '更新失败' })
    }
  } catch (e) {
    Message.error({ content: '更新失败' })
  }
}

// ========== YAML 查看/编辑功能 ==========
const openYamlPreview = async (deployment) => {
  showMoreOptions.value = false
  selectedYamlDeployment.value = deployment
  yamlContent.value = ''
  yamlError.value = ''
  yamlEditMode.value = false
  loadingYaml.value = true
  showYamlModal.value = true

  try {
    const res = await deploymentsApi.yaml({ namespace: deployment.namespace, name: deployment.name })
    if (res.code === 0) {
      yamlContent.value = res.data?.yaml || res.data || '# YAML 内容为空'
    } else {
      yamlError.value = res.msg || '获取 YAML 失败'
    }
  } catch (e) {
    yamlError.value = e?.msg || e?.message || '获取 YAML 失败'
  } finally {
    loadingYaml.value = false
  }
}

const closeYamlModal = () => {
  showYamlModal.value = false
  selectedYamlDeployment.value = null
  yamlContent.value = ''
  yamlError.value = ''
  yamlEditMode.value = false
}

const applyYamlChanges = async () => {
  if (!yamlContent.value?.trim()) {
    Message.error({ content: '请输入 YAML 内容' })
    return
  }
  savingYaml.value = true
  try {
    const res = await deploymentsApi.applyYaml({
      namespace: selectedYamlDeployment.value.namespace,
      name: selectedYamlDeployment.value.name,
      yaml: yamlContent.value
    })
    if (res.code === 0) {
      Message.success({ content: 'YAML 应用成功' })
      closeYamlModal()
      fetchDeployments()
      autoRefresh.value = true
      setTimeout(() => { autoRefresh.value = false }, 15000)
    } else {
      const errorMsg = res.details ? `${res.msg}: ${res.details}` : (res.msg || '应用 YAML 失败')
      Message.error({ content: errorMsg, duration: 5000 })
    }
  } catch (e) {
    const errorMsg = e?.details ? `${e?.msg || '应用 YAML 失败'}: ${e.details}` : (e?.msg || e?.message || '应用 YAML 失败')
    Message.error({ content: errorMsg, duration: 5000 })
  } finally {
    savingYaml.value = false
  }
}

// 下载 YAML
const downloadYaml = () => {
  if (!yamlContent.value || !selectedYamlDeployment.value) {
    Message.warning({ content: '没有可下载的 YAML 内容' })
    return
  }
  
  try {
    const blob = new Blob([yamlContent.value], { type: 'text/yaml;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `${selectedYamlDeployment.value.name}-deployment.yaml`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
    Message.success({ content: 'YAML 文件已下载' })
  } catch (e) {
    Message.error({ content: '下载失败' })
  }
}
</script>

<style scoped>
@import '@/assets/styles/resizable-modal.css';

.resource-view {
  max-width: 1400px;
  margin: 0 auto;
}

.view-header {
  margin-bottom: 24px;
}

.view-header h1 {
  font-size: 32px;
  font-weight: 700;
  color: #2d3748;
  margin-bottom: 8px;
}

.view-header p {
  font-size: 16px;
  color: #718096;
}

.action-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  flex-wrap: wrap;
  gap: 16px;
}

.search-box input {
  padding: 10px 16px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  width: 280px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.filter-buttons {
  display: flex;
  gap: 8px;
}

.btn-filter {
  padding: 8px 16px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  background: white;
  color: #4a5568;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-filter:hover {
  border-color: #326ce5;
  color: #326ce5;
}

.btn-filter.active {
  background: #326ce5;
  border-color: #326ce5;
  color: white;
}

.filter-dropdown select {
  padding: 10px 16px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  background-color: white;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.action-buttons {
  display: flex;
  gap: 12px;
  align-items: center;
}

.auto-refresh-toggle {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #4a5568;
  cursor: pointer;
}

.auto-refresh-toggle input {
  cursor: pointer;
}

.refresh-indicator {
  color: #34d399;
  animation: pulse 1s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.btn-primary {
  background-color: #326ce5;
  color: white;
}

.btn-primary:hover {
  background-color: #2554c7;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(50, 108, 229, 0.3);
}

.btn-secondary {
  background-color: #e2e8f0;
  color: #4a5568;
}

.btn-secondary:hover {
  background-color: #cbd5e0;
}

.btn-danger {
  background-color: #e53e3e;
  color: white;
}

.btn-warning {
  background-color: #f59e0b;
  color: white;
}

.error-box {
  background: #fff5f5;
  border: 1px solid #feb2b2;
  color: #c53030;
  padding: 12px 16px;
  border-radius: 8px;
  margin-bottom: 16px;
}

.table-container {
  background: white;
  border-radius: 12px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.05);
  overflow-x: auto; /* 支持横向滚动 */
  overflow-y: visible;
}

.resource-table {
  width: 100%;
  border-collapse: collapse;
  min-width: 1400px; /* 确保表格有足够宽度避免挤压 */
  table-layout: auto; /* 根据内容自动调整 */
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
}

.resource-table tbody tr:hover {
  background-color: #f7fafc;
}

.status-indicator {
  display: inline-flex;
  align-items: center;
  padding: 6px 12px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
}

.status-indicator.running {
  background-color: rgba(52, 211, 153, 0.1);
  color: #34d399;
}

.status-indicator.failed {
  background-color: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.status-indicator.updating {
  background-color: rgba(245, 158, 11, 0.1);
  color: #f59e0b;
}

.status-indicator.stopped {
  background-color: rgba(107, 114, 128, 0.1);
  color: #6b7280;
}

.status-indicator.pending {
  background-color: rgba(156, 163, 175, 0.1);
  color: #9ca3af;
}

.deployment-name {
  display: flex;
  align-items: center;
  gap: 8px;
}

.deployment-name .icon {
  font-size: 18px;
}

.namespace-badge {
  display: inline-block;
  padding: 4px 8px;
  background-color: rgba(50, 108, 229, 0.1);
  color: #326ce5;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.replicas-info {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

/* 副本数 +/- 控制 */
.replicas-control {
  display: flex;
  align-items: center;
  gap: 4px;
}

.replica-btn {
  width: 26px;
  height: 26px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  background: white;
  color: #4a5568;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  line-height: 1;
}

.replica-btn:hover:not(:disabled) {
  border-color: #326ce5;
  color: #326ce5;
  background: rgba(50, 108, 229, 0.05);
}

.replica-btn:active:not(:disabled) {
  transform: scale(0.95);
}

.replica-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.replica-btn.plus:hover:not(:disabled) {
  border-color: #34d399;
  color: #34d399;
  background: rgba(52, 211, 153, 0.05);
}

.replica-btn.minus:hover:not(:disabled) {
  border-color: #f59e0b;
  color: #f59e0b;
  background: rgba(245, 158, 11, 0.05);
}

.replica-btn.stop {
  font-size: 14px;
  padding: 2px;
}

.replica-btn.stop:hover:not(:disabled) {
  border-color: #ef4444;
  color: #ef4444;
  background: rgba(239, 68, 68, 0.05);
}


.replicas-display {
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 4px 8px;
  min-width: 50px;
  justify-content: center;
  font-size: 14px;
  font-weight: 500;
  border-radius: 4px;
  transition: background 0.2s;
}

.replicas-display.updating {
  background: rgba(245, 158, 11, 0.1);
}

.replicas-sep {
  color: #a0aec0;
}

.scaling-indicator {
  margin-left: 4px;
  animation: pulse 1s infinite;
}

.replicas-text {
  font-size: 14px;
  font-weight: 500;
}

.ready-replicas {
  color: #34d399;
  font-weight: 600;
}

.desired-replicas {
  color: #718096;
}

.replicas-bar {
  width: 100%;
  height: 6px;
  background-color: #e2e8f0;
  border-radius: 3px;
  overflow: hidden;
}

.replicas-fill {
  height: 100%;
  background-color: #326ce5;
  transition: width 0.3s ease;
}

/* 内联编辑样式 */
.inline-edit-wrapper {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.inline-input {
  padding: 6px 10px;
  border: 2px solid #326ce5;
  border-radius: 6px;
  font-size: 13px;
  width: 100%;
  min-width: 180px;
  outline: none;
  background: #fff;
  box-shadow: 0 2px 8px rgba(50, 108, 229, 0.2);
}

.inline-input-sm {
  min-width: 80px;
  width: 80px;
  text-align: center;
}

.inline-input:focus {
  border-color: #2554c7;
}

.inline-hint {
  font-size: 11px;
  color: #718096;
}

/* 副本数内联编辑样式 */
.replicas-edit-wrapper {
  display: flex;
  align-items: center;
  gap: 4px;
}

.replicas-input {
  padding: 4px 8px;
  border: 2px solid #326ce5;
  border-radius: 6px;
  font-size: 13px;
  width: 60px;
  text-align: center;
  outline: none;
  background: #fff;
  box-shadow: 0 2px 8px rgba(50, 108, 229, 0.2);
}

.replicas-input:focus {
  border-color: #2554c7;
}

.inline-hint-small {
  font-size: 10px;
  color: #718096;
  white-space: nowrap;
}

.edit-icon-small {
  opacity: 0;
  font-size: 11px;
  margin-left: 4px;
  transition: opacity 0.2s;
}

.replicas-display.clickable:hover .edit-icon-small {
  opacity: 1;
}


.clickable {
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  margin: -4px -8px;
  border-radius: 6px;
  transition: all 0.2s;
}

.clickable:hover {
  background-color: rgba(50, 108, 229, 0.08);
}

.clickable .edit-icon {
  opacity: 0;
  font-size: 12px;
  transition: opacity 0.2s;
}

.clickable:hover .edit-icon {
  opacity: 1;
}

.image-text {
  max-width: 280px;
}

.image-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: inline-block;
  max-width: 240px;
  vertical-align: middle;
}

.selector-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.selector-tag {
  display: inline-block;
  padding: 3px 6px;
  background-color: #edf2f7;
  color: #4a5568;
  border-radius: 3px;
  font-size: 11px;
  font-weight: 500;
}

.strategy-badge {
  display: inline-block;
  padding: 4px 8px;
  background-color: rgba(139, 92, 246, 0.1);
  color: #8b5cf6;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.action-icons {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

/* Pod 关联按钮 */
.action-btn {
  border: none;
  font-size: 0.8125rem;
  cursor: pointer;
  padding: 0.5rem 0.75rem;
  border-radius: 0.375rem;
  transition: all 0.2s;
  white-space: nowrap;
  font-weight: 500;
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
}

.action-btn.primary {
  background: linear-gradient(135deg, #326ce5 0%, #2558c9 100%);
  color: white;
  box-shadow: 0 2px 4px rgba(50, 108, 229, 0.2);
}

.action-btn.primary:hover {
  background: linear-gradient(135deg, #2558c9 0%, #1e45a0 100%);
  box-shadow: 0 4px 8px rgba(50, 108, 229, 0.3);
  transform: translateY(-1px);
}

.action-btn.primary:active {
  transform: translateY(0);
}

.icon-btn {
  background: none;
  border: 1px solid #e2e8f0;
  font-size: 0.8125rem;
  cursor: pointer;
  padding: 0.375rem 0.625rem;
  border-radius: 0.375rem;
  color: #4a5568;
  transition: all 0.2s;
  white-space: nowrap;
}

.icon-btn:hover {
  background-color: #edf2f7;
  border-color: #326ce5;
  color: #326ce5;
}

.more-btn {
  position: relative;
}

.more-menu {
  position: fixed;
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.15);
  min-width: 160px;
  z-index: 1000;
  padding: 6px 0;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 10px 16px;
  border: none;
  background: none;
  font-size: 13px;
  color: #2d3748;
  cursor: pointer;
  text-align: left;
  transition: background 0.15s;
}

.menu-item:hover {
  background: #f7fafc;
}

.menu-item.danger {
  color: #e53e3e;
}

.menu-item.danger:hover {
  background: #fff5f5;
}

.menu-icon {
  font-size: 14px;
}

.menu-divider {
  height: 1px;
  background: #e2e8f0;
  margin: 6px 0;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 20px;
  color: #a0aec0;
}

.empty-state-small {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 40px 20px;
  color: #a0aec0;
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.empty-text {
  font-size: 16px;
}

.loading-state {
  text-align: center;
  padding: 40px;
  color: #718096;
}

/* Modal styles */
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
  z-index: 1000;
}

.modal-content {
  background: white;
  border-radius: 12px;
  width: 90%;
  max-width: 600px;
  max-height: 85vh;
  overflow-y: auto;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.2);
}

.modal-content.modal-lg {
  max-width: 900px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #e2e8f0;
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
  color: #2d3748;
}

.close-btn {
  background: none;
  border: none;
  font-size: 24px;
  color: #718096;
  cursor: pointer;
  padding: 0;
  line-height: 1;
}

.close-btn:hover {
  color: #2d3748;
}

.modal-body {
  padding: 24px;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 24px;
  border-top: 1px solid #e2e8f0;
}

.info-box {
  background: #f7fafc;
  padding: 12px 16px;
  border-radius: 8px;
  margin-bottom: 16px;
}

.info-box div {
  margin-bottom: 4px;
}

.info-box div:last-child {
  margin-bottom: 0;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  font-size: 14px;
  font-weight: 500;
  color: #4a5568;
  margin-bottom: 6px;
}

.form-input, .form-select {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-size: 14px;
  transition: border-color 0.2s;
}

.form-input:focus, .form-select:focus {
  outline: none;
  border-color: #326ce5;
  box-shadow: 0 0 0 3px rgba(50, 108, 229, 0.1);
}

.form-static {
  padding: 10px 12px;
  background: #f7fafc;
  border-radius: 6px;
  font-size: 14px;
  color: #4a5568;
}

/* Event styles */
.event-item {
  display: flex;
  gap: 12px;
  padding: 12px;
  border-bottom: 1px solid #f7fafc;
}

.event-item:last-child {
  border-bottom: none;
}

.event-type {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  flex-shrink: 0;
}

.event-type.normal {
  background: rgba(52, 211, 153, 0.1);
  color: #34d399;
}

.event-type.warning {
  background: rgba(245, 158, 11, 0.1);
  color: #f59e0b;
}

.event-content {
  flex: 1;
}

.event-reason {
  font-weight: 600;
  color: #2d3748;
  margin-bottom: 4px;
}

.event-message {
  font-size: 13px;
  color: #4a5568;
  margin-bottom: 4px;
}

.event-time {
  font-size: 12px;
  color: #a0aec0;
}

.event-count {
  color: #f59e0b;
}

/* Detail JSON */
.detail-json {
  background: #1a202c;
  color: #68d391;
  padding: 16px;
  border-radius: 8px;
  font-size: 12px;
  font-family: 'Monaco', 'Menlo', monospace;
  overflow-x: auto;
  max-height: 400px;
  white-space: pre-wrap;
  word-break: break-all;
}

/* Simple table */
.simple-table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 12px;
}

.simple-table th,
.simple-table td {
  padding: 10px 12px;
  text-align: left;
  border-bottom: 1px solid #e2e8f0;
  font-size: 13px;
}

.simple-table th {
  background: #f7fafc;
  font-weight: 600;
  color: #4a5568;
}

.simple-table tbody tr:hover {
  background: #f7fafc;
}

/* Pod 操作相关样式 */
.pod-name-cell {
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pod-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.action-link {
  background: none;
  border: none;
  font-size: 16px;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  transition: all 0.2s;
  opacity: 0.7;
}

.action-link:hover {
  opacity: 1;
  background: rgba(50, 108, 229, 0.1);
}

.action-link.danger:hover {
  background: rgba(229, 62, 62, 0.1);
}

/* 日志控制栏 */
.logs-control-bar {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px 16px;
  background: #f7fafc;
  border-radius: 8px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.control-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.control-item label {
  font-size: 13px;
  color: #4a5568;
  white-space: nowrap;
}

.form-select-sm {
  padding: 6px 10px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-size: 13px;
  background: white;
}

.btn-sm {
  padding: 6px 12px;
  font-size: 13px;
}

/* 日志内容 */
.logs-content {
  background: #1a202c;
  color: #68d391;
  padding: 16px;
  border-radius: 8px;
  font-size: 12px;
  font-family: 'Monaco', 'Menlo', monospace;
  overflow-x: auto;
  max-height: 500px;
  overflow-y: auto;
  white-space: pre-wrap;
  word-break: break-all;
  min-height: 200px;
}

/* 日志控制面板 */
.logs-controls {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 16px;
  padding: 16px;
  background: #f7fafc;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
}

.control-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.control-item label {
  font-size: 13px;
  font-weight: 500;
  color: #4a5568;
}

.control-item .form-select {
  min-width: 150px;
  padding: 8px 12px;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 13px;
}

.follow-toggle {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  user-select: none;
}

.follow-toggle input[type="checkbox"] {
  cursor: pointer;
}

.streaming-indicator {
  color: #ef4444;
  animation: pulse 1s infinite;
}

.control-actions {
  display: flex;
  gap: 8px;
  align-items: flex-end;
}

.logs-viewer {
  margin-top: 16px;
}

/* Service 配置区域 */
.service-config-section {
  margin-top: 12px;
  padding: 16px;
  background: #f0f9ff;
  border: 1px solid #bfdbfe;
  border-radius: 8px;
}

/* Deployment 日志样式 */
.logs-modal .modal-content {
  max-width: 90vw;
  width: 1200px;
  max-height: 90vh;
}

.logs-modal .modal-body {
  max-height: 70vh;
  overflow-y: auto;
}

.deployment-logs-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.pod-log-section {
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  overflow: hidden;
  background: #f8fafc;
}

.pod-log-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  font-weight: 600;
  font-size: 14px;
}

.pod-name {
  font-size: 14px;
}

.container-name {
  font-size: 12px;
  opacity: 0.9;
}

.pod-log-section .logs-content {
  background: #1e293b;
  color: #e2e8f0;
  padding: 16px;
  font-size: 12px;
  line-height: 1.6;
  overflow-x: auto;
  font-family: 'Consolas', 'Monaco', ui-monospace, monospace;
  margin: 0;
  max-height: 400px;
  overflow-y: auto;
  white-space: pre-wrap;
  word-break: break-word;
}

.pod-log-section .error-box {
  margin: 12px;
}

@media (max-width: 1200px) {
  .resource-table {
    display: block;
    overflow-x: auto;
  }

  .search-box input {
    width: 200px;
  }

  .action-bar {
    flex-direction: column;
    align-items: stretch;
  }

  .action-buttons {
    justify-content: flex-end;
  }

  .filter-buttons {
    flex-wrap: wrap;
  }
}

/* ========== 日志弹窗样式（增强版 - 参考 Pods.vue） ========== */
.logs-modal {
  width: 90vw;
  max-width: 1000px;
  height: 80vh;
  display: flex;
  flex-direction: column;
}

.logs-modal .modal-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding: 16px;
}

.pod-info-bar {
  display: flex;
  gap: 24px;
  padding: 10px 14px;
  background: linear-gradient(135deg, #f8fafc 0%, #e2e8f0 100%);
  border-radius: 8px;
  margin-bottom: 12px;
  font-size: 13px;
  color: #475569;
  border-left: 4px solid #3b82f6;
}

.logs-control-bar {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px 0;
  border-bottom: 2px solid #e2e8f0;
  margin-bottom: 12px;
  flex-wrap: wrap;
}

.logs-control-bar .control-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.logs-control-bar label {
  font-size: 13px;
  color: #64748b;
  white-space: nowrap;
  font-weight: 500;
}

.logs-control-bar .form-select {
  padding: 6px 10px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-size: 13px;
  background: #fff;
  cursor: pointer;
  transition: all 0.2s;
}

.logs-control-bar .form-select:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.single-container {
  padding: 6px 10px;
  background: linear-gradient(135deg, #dbeafe 0%, #bfdbfe 100%);
  border-radius: 6px;
  font-size: 13px;
  color: #1e40af;
  font-weight: 500;
}

.follow-toggle {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  padding: 6px 10px;
  background: #f1f5f9;
  border-radius: 6px;
  transition: all 0.2s;
  user-select: none;
}

.follow-toggle:hover {
  background: #e2e8f0;
}

.follow-toggle input[type="checkbox"] {
  cursor: pointer;
}

.follow-toggle input:checked + span {
  color: #3b82f6;
  font-weight: 600;
}

.streaming-indicator {
  color: #22c55e;
  animation: pulse 1.5s infinite;
  font-size: 16px;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

.btn-sm {
  padding: 6px 14px;
  font-size: 13px;
}

.btn-danger {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: #fff;
  border: none;
}

.btn-danger:hover {
  background: linear-gradient(135deg, #dc2626 0%, #b91c1c 100%);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);
}

.logs-content-wrapper,
.logs-viewer {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.logs-content {
  flex: 1;
  margin: 0;
  padding: 14px;
  background: linear-gradient(135deg, #1e293b 0%, #0f172a 100%);
  color: #e2e8f0;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', monospace;
  font-size: 12px;
  line-height: 1.6;
  border-radius: 8px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
  box-shadow: inset 0 2px 8px rgba(0, 0, 0, 0.3);
}

.logs-content::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

.logs-content::-webkit-scrollbar-track {
  background: #334155;
  border-radius: 4px;
}

.logs-content::-webkit-scrollbar-thumb {
  background: linear-gradient(135deg, #64748b 0%, #475569 100%);
  border-radius: 4px;
}

.logs-content::-webkit-scrollbar-thumb:hover {
  background: linear-gradient(135deg, #94a3b8 0%, #64748b 100%);
}

/* 日志高亮样式（精美配色） */
.logs-content :deep(.log-placeholder) {
  color: #64748b;
  font-style: italic;
  font-size: 13px;
}

.logs-content :deep(.log-timestamp) {
  color: #94a3b8;
  font-weight: 500;
}

.logs-content :deep(.log-error) {
  color: #fca5a5;
  background: linear-gradient(90deg, rgba(248, 113, 113, 0.15) 0%, rgba(239, 68, 68, 0.1) 100%);
  display: block;
  margin: 0 -14px;
  padding: 0 14px;
  border-left: 3px solid #f87171;
  font-weight: 500;
}

.logs-content :deep(.log-warn) {
  color: #fcd34d;
  background: linear-gradient(90deg, rgba(251, 191, 36, 0.12) 0%, rgba(245, 158, 11, 0.08) 100%);
  display: block;
  margin: 0 -14px;
  padding: 0 14px;
  border-left: 3px solid #fbbf24;
  font-weight: 500;
}

.logs-content :deep(.log-info) {
  color: #60a5fa;
  font-weight: 400;
}

.logs-content :deep(.log-debug) {
  color: #9ca3af;
  opacity: 0.8;
}

.logs-content :deep(.log-separator) {
  color: #34d399;
  background: linear-gradient(90deg, rgba(52, 211, 153, 0.15) 0%, rgba(16, 185, 129, 0.1) 100%);
  display: block;
  margin: 8px -14px;
  padding: 6px 14px;
  border-left: 3px solid #10b981;
  font-weight: 600;
  letter-spacing: 0.5px;
}

/* 控制面板样式优化 */
.logs-controls {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  padding: 12px;
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  border-radius: 8px;
  margin-bottom: 12px;
  border: 1px solid #e2e8f0;
}

.logs-controls .control-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.logs-controls label {
  font-size: 13px;
  font-weight: 500;
  color: #475569;
}

.control-actions {
  display: flex;
  gap: 8px;
  margin-left: auto;
}

/* ========== Rancher 风格创建表单样式 ========== */
.modal-create-deployment {
  max-width: 900px;
  max-height: 90vh;
}

.modal-create-deployment .modal-body {
  max-height: 75vh;
  overflow-y: auto;
  padding: 24px 28px;
}

/* 表单分组卡片 */
.form-section {
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  margin-bottom: 20px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.section-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 18px;
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  border-bottom: 1px solid #e2e8f0;
}

.section-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #1e293b;
  flex: 1;
}

.section-icon {
  font-size: 20px;
}

.section-body {
  padding: 18px;
}

/* 表单元素 */
.form-group {
  margin-bottom: 18px;
}

.form-group:last-child {
  margin-bottom: 0;
}

.form-group label {
  display: block;
  font-size: 14px;
  font-weight: 500;
  color: #334155;
  margin-bottom: 8px;
}

.required {
  color: #ef4444;
  margin-left: 2px;
}

.form-input,
.form-select {
  width: 100%;
  padding: 10px 14px;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 14px;
  background: #fff;
  transition: all 0.2s;
  box-sizing: border-box;
}

.form-input:focus,
.form-select:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-input::placeholder {
  color: #9ca3af;
}

.form-hint {
  font-size: 12px;
  color: #6b7280;
  margin-top: 6px;
}

.mb-12 {
  margin-bottom: 12px;
}

.mb-16 {
  margin-bottom: 16px;
}

/* 折叠面板 */
.section-header.clickable {
  cursor: pointer;
  user-select: none;
}

.section-header.clickable:hover {
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
}

.toggle-indicator {
  margin-left: auto;
  font-size: 12px;
  color: #64748b;
  transition: transform 0.2s ease;
}

.toggle-indicator.expanded {
  transform: rotate(-180deg);
}

/* Resources 配置 */
.resource-group {
  margin-bottom: 20px;
}

.resource-group:last-child {
  margin-bottom: 0;
}

.resource-title {
  font-size: 15px;
  font-weight: 600;
  color: #1e293b;
  margin: 0 0 12px 0;
  display: flex;
  align-items: center;
  gap: 6px;
}

.resource-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.input-with-unit {
  position: relative;
  display: flex;
  align-items: center;
}

.input-with-unit .form-input {
  flex: 1;
  padding-right: 60px;
}

.unit-hint {
  position: absolute;
  right: 14px;
  font-size: 12px;
  color: #9ca3af;
  font-weight: 500;
  pointer-events: none;
}

.resource-tips {
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
  border-left: 3px solid #3b82f6;
  padding: 12px 16px;
  border-radius: 8px;
  font-size: 13px;
  color: #1e40af;
  margin-top: 16px;
}

.resource-tips .tip-icon {
  font-size: 16px;
  margin-right: 6px;
}

.resource-tips ul {
  margin: 8px 0 0 0;
  padding-left: 20px;
}

.resource-tips li {
  margin: 4px 0;
}

/* Probes 配置 */
.probe-config {
  background: linear-gradient(135deg, #fafbfc 0%, #f8fafc 100%);
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
}

.probe-config:last-child {
  margin-bottom: 0;
}

.probe-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.probe-title {
  font-size: 15px;
  font-weight: 600;
  color: #1e293b;
  margin: 0;
}

.probe-body {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #e2e8f0;
}

.probe-timing-config {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px dashed #e2e8f0;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

/* 标签编辑器 */
.labels-editor {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.label-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.label-input {
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 13px;
  transition: all 0.2s;
}

.label-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

.label-key {
  flex: 1;
  max-width: 200px;
}

.label-value {
  flex: 2;
}

.label-separator {
  color: #6b7280;
  font-weight: 600;
}

.btn-icon {
  padding: 6px 10px;
  border: none;
  background: none;
  cursor: pointer;
  font-size: 16px;
  border-radius: 6px;
  transition: all 0.2s;
}

.btn-remove {
  color: #dc2626;
}

.btn-remove:hover {
  background: #fee2e2;
}

.btn-remove:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Toggle 开关 */
.toggle-switch {
  position: relative;
  display: inline-block;
  width: 48px;
  height: 24px;
  margin-left: auto;
}

.toggle-switch input[type="checkbox"] {
  opacity: 0;
  width: 0;
  height: 0;
}

.toggle-slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: #cbd5e1;
  transition: 0.3s;
  border-radius: 24px;
}

.toggle-slider:before {
  position: absolute;
  content: "";
  height: 18px;
  width: 18px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: 0.3s;
  border-radius: 50%;
}

.toggle-switch input:checked + .toggle-slider {
  background-color: #3b82f6;
}

.toggle-switch input:checked + .toggle-slider:before {
  transform: translateX(24px);
}

/* Service 类型选择器 */
.service-type-selector {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 12px;
}

.service-type-card {
  position: relative;
  padding: 16px 12px;
  border: 2px solid #e2e8f0;
  border-radius: 10px;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s;
  background: #fff;
}

.service-type-card:hover {
  border-color: #3b82f6;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.15);
}

.service-type-card.active {
  border-color: #3b82f6;
  background: linear-gradient(135deg, #dbeafe 0%, #eff6ff 100%);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.type-icon {
  font-size: 32px;
  margin-bottom: 8px;
}

.type-name {
  font-weight: 600;
  color: #1e293b;
  margin-bottom: 4px;
  font-size: 14px;
}

.type-desc {
  font-size: 11px;
  color: #64748b;
}

/* 调度策略选择器 */
.scheduling-policy-selector {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 12px;
  margin-bottom: 16px;
}

.scheduling-policy-card {
  position: relative;
  padding: 16px;
  border: 2px solid #e2e8f0;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
  background: #fff;
}

.scheduling-policy-card:hover {
  border-color: #3b82f6;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.15);
}

.scheduling-policy-card.active {
  border-color: #3b82f6;
  background: linear-gradient(135deg, #dbeafe 0%, #eff6ff 100%);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.policy-title {
  font-weight: 600;
  color: #1e293b;
  margin-bottom: 4px;
  font-size: 14px;
}

.policy-desc {
  font-size: 12px;
  color: #64748b;
  line-height: 1.4;
}

/* 自定义调度规则配置 */
.custom-scheduling-config {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #e2e8f0;
  animation: slideDown 0.3s ease-out;
}

/* 容忍编辑器 */
.tolerations-editor {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.toleration-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px;
  background: #f8fafc;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
  flex-wrap: wrap;
}

.toleration-input {
  flex: 1;
  min-width: 100px;
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 13px;
  transition: border-color 0.2s;
}

.toleration-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.toleration-select {
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 13px;
  background: #fff;
  cursor: pointer;
  min-width: 100px;
}

.toleration-select:focus {
  outline: none;
  border-color: #3b82f6;
}

.mb-12 {
  margin-bottom: 12px;
}

/* 容忍时间输入 */
.toleration-seconds {
  max-width: 120px;
}

/* 节点亲和性编辑器 */
.affinity-editor {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.affinity-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px;
  background: #f8fafc;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
  flex-wrap: wrap;
}

.affinity-input {
  flex: 1;
  min-width: 100px;
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 13px;
  transition: border-color 0.2s;
}

.affinity-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.affinity-values {
  min-width: 150px;
}

.affinity-weight {
  max-width: 100px;
}

.affinity-select {
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 13px;
  background: #fff;
  cursor: pointer;
  min-width: 120px;
}

.affinity-required-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #475569;
  cursor: pointer;
  white-space: nowrap;
}

.affinity-required-label input[type="checkbox"] {
  width: 16px;
  height: 16px;
  cursor: pointer;
}

/* 拓扑分布约束编辑器 */
.topology-editor {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.topology-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px;
  background: #f8fafc;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
  flex-wrap: wrap;
}

.topology-select {
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 13px;
  background: #fff;
  cursor: pointer;
  min-width: 160px;
}

.topology-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.topology-field label {
  font-size: 11px;
  color: #64748b;
  font-weight: 500;
}

.topology-input {
  width: 80px;
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 13px;
}

/* 端口映射编辑器 */
.port-mapping-editor {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.port-row {
  display: flex;
  align-items: flex-end;
  gap: 10px;
  padding: 12px;
  background: #f8fafc;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
}

.port-field {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.port-field label {
  font-size: 12px;
  font-weight: 500;
  color: #64748b;
  margin-bottom: 6px;
}

.port-field .form-input {
  padding: 8px 10px;
  font-size: 13px;
}

.port-arrow {
  color: #3b82f6;
  font-weight: 600;
  font-size: 18px;
  margin-bottom: 10px;
}

/* 服务配置区域动画 */
.service-config {
  animation: slideDown 0.3s ease-out;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 按钮样式增强 */
.btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 10px 20px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  color: white;
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.3);
}

.btn-primary:hover {
  background: linear-gradient(135deg, #2563eb 0%, #1d4ed8 100%);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
  transform: translateY(-1px);
}

.btn-secondary {
  background: #f1f5f9;
  color: #475569;
  border: 1px solid #e2e8f0;
}

.btn-secondary:hover {
  background: #e2e8f0;
  border-color: #cbd5e1;
}

.btn-icon {
  font-size: 16px;
}

/* 响应式优化 */
@media (max-width: 768px) {
  .modal-create-deployment {
    max-width: 95vw;
  }

  .form-row {
    grid-template-columns: 1fr;
  }

  .service-type-selector {
    grid-template-columns: 1fr 1fr;
  }

  .port-row {
    flex-wrap: wrap;
  }

  .port-field {
    min-width: 45%;
  }
}

/* 命名空间选择器样式 */
.namespace-selector {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.namespace-or {
  color: #6b7280;
  font-size: 13px;
  font-weight: 500;
}

.namespace-create {
  display: flex;
  gap: 8px;
  flex: 1;
  min-width: 200px;
}

.namespace-create .form-input {
  flex: 1;
}

.namespace-create .btn-sm {
  white-space: nowrap;
}

/* ==================== */
/* 版本记录弹窗样式 */
/* ==================== */
.version-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  background: linear-gradient(135deg, #dbeafe 0%, #eff6ff 100%);
  border: 1px solid #bfdbfe;
  border-radius: 6px;
  font-weight: 600;
  font-size: 13px;
  color: #1e40af;
}

.current-tag {
  padding: 2px 6px;
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  box-shadow: 0 2px 4px rgba(16, 185, 129, 0.3);
}

.current-version-text {
  color: #10b981;
  font-weight: 500;
  font-size: 13px;
}

.mono {
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 12px;
  color: #475569;
  background: #f8fafc;
  padding: 4px 8px;
  border-radius: 4px;
}

.simple-table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 10px;
}

.simple-table th,
.simple-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #e2e8f0;
}

.simple-table th {
  background: #f8fafc;
  font-weight: 600;
  color: #475569;
  font-size: 13px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.simple-table tbody tr:hover {
  background: #f8fafc;
}

.simple-table tbody tr:last-child td {
  border-bottom: none;
}

.btn-sm {
  padding: 6px 12px;
  font-size: 13px;
}

.modal-lg {
  max-width: 1000px;
}

/* ==================== */
/* 视图切换按钮样式 */
/* ==================== */
.view-toggle {
  display: flex;
  gap: 0;
  border-radius: 6px;
  overflow: hidden;
  border: 1px solid #d1d5db;
}

.btn-view {
  padding: 8px 16px;
  background: white;
  border: none;
  font-size: 18px;
  cursor: pointer;
  transition: all 0.2s;
  border-right: 1px solid #d1d5db;
}

.btn-view:last-child {
  border-right: none;
}

.btn-view:hover {
  background: #f3f4f6;
}

.btn-view.active {
  background: #3b82f6;
  color: white;
}

/* ==================== */
/* 卡片视图样式 */
/* ==================== */
.cards-container {
  padding: 0;
}

.cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(380px, 1fr));
  gap: 20px;
  margin-bottom: 20px;
}

.deployment-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  overflow: hidden;
  transition: all 0.3s ease;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.deployment-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}

/* 卡片头部 */
.card-header {
  padding: 16px;
  background: linear-gradient(135deg, #f8fafc 0%, #e2e8f0 100%);
  border-bottom: 1px solid #e2e8f0;
}

.card-title-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.card-icon {
  font-size: 24px;
}

.card-title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #1e293b;
  flex: 1;
}

/* 卡片主体 */
.card-body {
  padding: 16px;
}

.card-section {
  margin-bottom: 16px;
}

.card-section:last-child {
  margin-bottom: 0;
}

.section-label {
  font-size: 12px;
  font-weight: 600;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 8px;
}

.card-section-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.card-meta-item {
  background: #f8fafc;
  padding: 12px;
  border-radius: 8px;
}

.meta-label {
  font-size: 11px;
  font-weight: 600;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 6px;
}

.meta-value {
  font-size: 13px;
  color: #1e293b;
  word-break: break-all;
}

/* 卡片底部按钮 */
.card-footer {
  display: flex;
  gap: 8px;
  padding: 12px 16px;
  background: #f8fafc;
  border-top: 1px solid #e2e8f0;
  flex-wrap: wrap;
}

.card-action-btn {
  flex: 1;
  min-width: 4.375rem;
  padding: 0.5rem 0.75rem;
  background: white;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
  font-size: 0.8125rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.card-action-btn:hover {
  background: #f3f4f6;
  border-color: #9ca3af;
}

.card-action-btn.primary {
  background: linear-gradient(135deg, #326ce5 0%, #2558c9 100%);
  color: white;
  border-color: #326ce5;
  box-shadow: 0 2px 4px rgba(50, 108, 229, 0.2);
}

.card-action-btn.primary:hover {
  background: linear-gradient(135deg, #2558c9 0%, #1e45a0 100%);
  box-shadow: 0 4px 8px rgba(50, 108, 229, 0.3);
  transform: translateY(-1px);
}

.card-action-btn.danger {
  color: #dc2626;
}

.card-action-btn.danger:hover {
  background: #fef2f2;
  border-color: #fca5a5;
}

/* ==================== */
/* 资源消耗样式 */
/* ==================== */
.metrics-summary {
  display: flex;
  gap: 12px;
}

.metric-item {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 8px 10px;
  background: linear-gradient(135deg, #f0f9ff 0%, #e0f2fe 100%);
  border: 1px solid #bae6fd;
  border-radius: 6px;
  font-size: 12px;
}

.metric-icon {
  font-size: 14px;
}

.metric-label {
  font-weight: 600;
  color: #0369a1;
}

.metric-value {
  font-weight: 700;
  color: #0c4a6e;
  font-family: 'Monaco', 'Menlo', monospace;
}

/* 无资源数据时的样式 */
.metrics-unavailable {
  flex-wrap: wrap;
}

.metrics-unavailable .metric-item {
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  border-color: #e2e8f0;
}

.metrics-unavailable .metric-value.muted {
  color: #94a3b8;
  font-weight: 500;
}

.metrics-hint {
  width: 100%;
  text-align: center;
  font-size: 11px;
  color: #94a3b8;
  margin-top: 4px;
  font-style: italic;
}

/* 响应式 - 小屏幕单列 */
@media (max-width: 768px) {
  .cards-grid {
    grid-template-columns: 1fr;
  }
  
  .card-footer {
    flex-direction: column;
  }
  
  .card-action-btn {
    width: 100%;
  }
  
  .metrics-summary {
    flex-direction: column;
    gap: 8px;
  }
}

/* ========== 批量操作样式 ========== */
.btn-batch {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
}

.btn-batch:hover {
  opacity: 0.9;
}

.batch-action-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 12px 20px;
  border-radius: 8px;
  margin-bottom: 16px;
}

.batch-info {
  display: flex;
  align-items: center;
  gap: 16px;
}

.batch-count {
  font-weight: 600;
  font-size: 14px;
}

.batch-clear {
  background: rgba(255,255,255,0.2);
  border: none;
  color: white;
  padding: 4px 12px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
}

.batch-clear:hover {
  background: rgba(255,255,255,0.3);
}

.batch-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.batch-btn {
  padding: 8px 16px;
  background: white;
  color: #4a5568;
  border: none;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.batch-btn:hover {
  background: #f7fafc;
}

.batch-btn.danger {
  background: #fed7d7;
  color: #c53030;
}

.batch-btn.danger:hover {
  background: #feb2b2;
}

.row-selected {
  background: rgba(102, 126, 234, 0.1) !important;
}

/* 批量预览弹窗 */
.modal-batch-preview {
  max-width: 700px;
}

.preview-section {
  margin-bottom: 20px;
}

.preview-section .section-title {
  font-weight: 600;
  color: #2d3748;
  margin-bottom: 12px;
  font-size: 14px;
}

.affected-deployments, .affected-deployments-detail {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 200px;
  overflow-y: auto;
}

.affected-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #f7fafc;
  padding: 8px 12px;
  border-radius: 6px;
}

.affected-item .dep-name {
  font-weight: 500;
  color: #2d3748;
}

.affected-item .dep-replicas {
  font-size: 12px;
  color: #718096;
}

.affected-dep-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #f7fafc;
  padding: 12px 16px;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
}

.dep-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.dep-info .dep-name {
  font-weight: 600;
  color: #2d3748;
}

.dep-info .dep-namespace {
  font-size: 12px;
  color: #718096;
}

.dep-stats {
  display: flex;
  gap: 12px;
  align-items: center;
}

.replicas-tag {
  font-size: 12px;
  color: #718096;
  background: #edf2f7;
  padding: 2px 8px;
  border-radius: 4px;
}

.change-preview {
  background: #f7fafc;
  border-radius: 8px;
  padding: 12px;
}

.change-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 0;
  font-size: 13px;
}

.change-item .old-value {
  color: #718096;
  text-decoration: line-through;
}

.change-item .arrow {
  color: #a0aec0;
}

.change-item .new-value {
  color: #38a169;
  font-weight: 600;
}

.modal-danger .danger-header {
  background: linear-gradient(135deg, #fc8181 0%, #f56565 100%);
}

.modal-warning .warning-header {
  background: linear-gradient(135deg, #fbbf24 0%, #f59e0b 100%);
}

.warning-box {
  display: flex;
  gap: 16px;
  background: #fffbeb;
  border: 1px solid #fcd34d;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 20px;
}

.btn-warning {
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
  color: #fff;
  border: none;
  box-shadow: 0 2px 8px rgba(245, 158, 11, 0.25);
}

.btn-warning:hover:not(:disabled) {
  box-shadow: 0 4px 12px rgba(245, 158, 11, 0.35);
}

.btn-warning:disabled {
  background: #fcd34d;
  cursor: not-allowed;
}

.danger-warning {
  display: flex;
  gap: 16px;
  background: #fff5f5;
  border: 1px solid #feb2b2;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 20px;
}

.warning-icon-large {
  font-size: 32px;
}

.warning-title {
  font-weight: 600;
  color: #c53030;
  margin-bottom: 8px;
}

.warning-list {
  margin: 0;
  padding-left: 20px;
  color: #742a2a;
  font-size: 13px;
}

.warning-list li {
  margin-bottom: 4px;
}

.confirm-section {
  margin-top: 20px;
}

.confirm-input {
  width: 100%;
  padding: 12px 16px;
  border: 2px solid #e2e8f0;
  border-radius: 8px;
  font-size: 16px;
  text-align: center;
  font-weight: 600;
  letter-spacing: 2px;
}

.confirm-input:focus {
  outline: none;
  border-color: #fc8181;
}

.confirm-input.valid {
  border-color: #48bb78;
  background: #f0fff4;
}

/* 卡片批量选择样式 */
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
  accent-color: #667eea;
}

.deployment-card {
  position: relative;
}

.card-selected {
  border: 2px solid #667eea !important;
  background: rgba(102, 126, 234, 0.05) !important;
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

.load-template-btn {
  align-self: flex-start;
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
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
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

</style>

