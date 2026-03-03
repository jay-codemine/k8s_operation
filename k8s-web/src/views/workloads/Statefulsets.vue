<template>
  <div class="resource-view">
    <div class="view-header">
      <h1>有状态集管理</h1>
      <p>Kubernetes集群中的有状态集列表</p>
    </div>
    <div class="action-bar">
      <div class="search-box">
        <input type="text" v-model="searchQuery" placeholder="搜索有状态集名称..." @input="onSearchInput" />
      </div>

      <div class="filter-buttons">
        <button class="btn btn-filter" :class="{ active: statusFilter === 'all' }" @click="setStatusFilter('all')">全部</button>
        <button class="btn btn-filter" :class="{ active: statusFilter === 'Running' }" @click="setStatusFilter('Running')">Running</button>
        <button class="btn btn-filter" :class="{ active: statusFilter === 'Updating' }" @click="setStatusFilter('Updating')">Updating</button>
        <button class="btn btn-filter" :class="{ active: statusFilter === 'Stopped' }" @click="setStatusFilter('Stopped')">Stopped</button>
        <button class="btn btn-filter" :class="{ active: statusFilter === 'Failed' }" @click="setStatusFilter('Failed')">Failed</button>
      </div>

      <div class="filter-dropdown">
        <select v-model="namespaceFilter">
          <option value="">所有命名空间</option>
          <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
        </select>
      </div>

      <div class="action-buttons">
        <button v-if="!batchMode" class="btn btn-batch" @click="enterBatchMode" title="进入批量操作模式">☑️ 批量操作</button>
        <button v-if="batchMode" class="btn btn-secondary" @click="exitBatchMode">✖️ 退出批量</button>

        <div class="view-toggle">
          <button class="btn btn-view" :class="{ active: viewMode === 'table' }" @click="viewMode = 'table'" title="表格视图">📋</button>
          <button class="btn btn-view" :class="{ active: viewMode === 'card' }" @click="viewMode = 'card'" title="卡片视图">🗂️</button>
        </div>
        
        <label class="auto-refresh-toggle">
          <input type="checkbox" v-model="autoRefresh" />
          <span>自动刷新</span>
          <span v-if="autoRefresh" class="refresh-indicator">●</span>
        </label>
        <button class="btn btn-primary" @click="showCreateModal = true">创建有状态集</button>
        <button class="btn btn-secondary" @click="refreshList" :disabled="loading">{{ loading ? '加载中...' : '🔄 刷新' }}</button>
      </div>
    </div>

    <div v-if="errorMsg" class="error-box">{{ errorMsg }}</div>

    <!-- 批量操作浮动栏 -->
    <div v-if="batchMode && selectedStatefulsets.length > 0" class="batch-action-bar">
      <div class="batch-info">
        <span class="batch-count">已选择 {{ selectedStatefulsets.length }} 个 StatefulSet</span>
        <button class="batch-clear" @click="clearSelection">清空选择</button>
      </div>
      <div class="batch-actions">
        <button class="batch-btn" @click="openBatchRestartPreview" title="批量重启">🔄 批量重启</button>
        <button class="batch-btn" @click="openBatchScaleModal" title="批量扩缩容">📊 批量扩缩容</button>
        <button class="batch-btn danger" @click="openBatchDeletePreview" title="批量删除">🗑️ 批量删除</button>
      </div>
    </div>
    
    <!-- 表格视图 -->
    <div v-if="viewMode === 'table'" class="table-container">
      <table class="resource-table">
        <thead>
          <tr>
            <th v-if="batchMode" style="width: 40px;">
              <input type="checkbox" :checked="isAllSelected" @change="toggleSelectAll" title="全选/取消全选" ref="selectAllCheckbox" />
            </th>
            <th style="width: 100px;">状态</th>
            <th style="min-width: 180px;">名称</th>
            <th style="width: 130px;">命名空间</th>
            <th style="width: 200px;">副本数</th>
            <th style="min-width: 200px;">镜像</th>
            <th style="min-width: 180px;">选择器</th>
            <th style="width: 120px;">更新策略</th>
            <th style="width: 170px; white-space: nowrap;">创建时间</th>
            <th style="width: 180px;">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="sts in paginatedStatefulsets" :key="sts.name" :class="{ 'row-selected': isStatefulsetSelected(sts) }">
            <td v-if="batchMode">
              <input type="checkbox" :checked="isStatefulsetSelected(sts)" @change="toggleStatefulsetSelection(sts)" />
            </td>
            <td>
              <span class="status-indicator" :class="sts.status.toLowerCase()">{{ sts.status }}</span>
            </td>
            <td>
              <div class="statefulset-name">
                <span class="icon">🏛️</span>
                <span>{{ sts.name }}</span>
              </div>
            </td>
            <td>
              <span class="namespace-badge">{{ sts.namespace }}</span>
            </td>
            <td>
              <div class="replicas-info">
                <div class="replicas-control">
                  <button class="replica-btn minus" @click="decreaseReplicas(sts)" :disabled="sts.desiredReplicas <= 0 || scalingMap[sts.name]" title="减少副本">−</button>
                  <div v-if="inlineEdit.key === `replicas-${sts.name}`" class="replicas-edit-wrapper">
                    <input type="number" v-model="inlineEdit.value" class="replicas-input" min="0" @blur="saveInlineReplicas(sts)" @keyup.enter="saveInlineReplicas(sts)" @keyup.escape="cancelInlineEdit" placeholder="副本数" autofocus />
                    <span class="inline-hint-small">回车</span>
                  </div>
                  <div v-else class="replicas-display clickable" :class="{ updating: scalingMap[sts.name] }" @click="startInlineReplicas(sts)" title="点击修改副本数">
                    <span class="ready-replicas">{{ sts.readyReplicas }}</span>
                    <span class="replicas-sep">/</span>
                    <span class="desired-replicas">{{ sts.desiredReplicas }}</span>
                    <span v-if="scalingMap[sts.name]" class="scaling-indicator">⏳</span>
                    <span class="edit-icon-small">✏️</span>
                  </div>
                  <button class="replica-btn plus" @click="increaseReplicas(sts)" :disabled="scalingMap[sts.name]" title="增加副本">+</button>
                  <button class="replica-btn stop" @click="stopService(sts)" :disabled="sts.desiredReplicas === 0 || scalingMap[sts.name]" title="停服（副本数调为0）">⏸️</button>
                </div>
                <div class="replicas-bar">
                  <div class="replicas-fill" :style="{ width: `${(sts.readyReplicas / Math.max(sts.desiredReplicas, 1)) * 100}%` }"></div>
                </div>
              </div>
            </td>
            <td>
              <div class="inline-edit-wrapper" v-if="inlineEdit.key === `image-${sts.name}`">
                <input type="text" v-model="inlineEdit.value" class="inline-input" @blur="saveInlineImage(sts)" @keyup.enter="saveInlineImage(sts)" @keyup.escape="cancelInlineEdit" placeholder="输入新镜像地址" autofocus />
                <span class="inline-hint">按 Enter 保存</span>
              </div>
              <div v-else class="image-text clickable" @click="startInlineImage(sts)" title="点击修改镜像">
                <span class="image-name">{{ sts.image || '-' }}</span>
                <span class="edit-icon">✏️</span>
              </div>
            </td>
            <td>
              <div class="selector-tags">
                <span v-for="(value, key) in sts.selector" :key="key" class="selector-tag">{{ key }}={{ value }}</span>
              </div>
            </td>
            <td>
              <span class="strategy-badge">{{ sts.updateStrategy }}</span>
            </td>
            <td style="white-space: nowrap;">{{ sts.createdAt }}</td>
            <td>
              <div class="action-icons">
                <!-- Pod 关联按钮（独立显示） -->
                <button class="action-btn primary" @click="viewPods(sts)" title="查看此 StatefulSet 管理的所有 Pod">
                  📦 Pod 关联
                </button>
                
                <!-- 更多按钮 -->
                <div class="more-btn">
                  <button class="icon-btn" @click="toggleMoreOptions(sts, $event)">⋮ 更多</button>
                  
                  <!-- 更多菜单 -->
                  <div v-if="showMoreOptions && selectedStatefulset === sts" class="more-menu" :style="menuStyle">
                    <button class="menu-item" @click="viewStatefulsetLogs(sts)">
                      <span class="menu-icon">📄</span>
                      <span>查看日志</span>
                    </button>
                    <button class="menu-item" @click="viewHistory(sts)">
                      <span class="menu-icon">📜</span>
                      <span>版本记录</span>
                    </button>
                    <div class="menu-divider"></div>
                    <button class="menu-item" @click="restartStatefulset(sts)">
                      <span class="menu-icon">🔄</span>
                      <span>重启</span>
                    </button>
                    <button class="menu-item" @click="openUpdateImage(sts)">
                      <span class="menu-icon">🔧</span>
                      <span>更新镜像</span>
                    </button>
                    <div class="menu-divider"></div>
                    <button class="menu-item" @click="viewStatefulset(sts)">
                      <span class="menu-icon">📋</span>
                      <span>查看详情</span>
                    </button>
                    <button class="menu-item" @click="openEvents(sts)">
                      <span class="menu-icon">📡</span>
                      <span>查看事件</span>
                    </button>
                    <button class="menu-item" @click="openYamlPreview(sts)">
                      <span class="menu-icon">📝</span>
                      <span>查看/编辑 YAML</span>
                    </button>
                    <div class="menu-divider"></div>
                    <button class="menu-item danger" @click="deleteStatefulset(sts)">
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
      <div v-if="filteredStatefulsets.length === 0" class="empty-state">
        <div class="empty-icon">🔍</div>
        <div class="empty-text">没有找到匹配的有状态集</div>
      </div>
      <!-- 分页 + 跳转 -->
      <div v-if="filteredStatefulsets.length > 0" class="pagination-wrapper">
        <Pagination v-model:currentPage="currentPage" :totalItems="filteredStatefulsets.length" :itemsPerPage="itemsPerPage" />
        <div class="jump-page">
          <span>跳至</span>
          <input type="number" v-model.number="jumpPage" min="1" :max="totalPages" @keyup.enter="handleJump" />
          <span>页</span>
          <button class="btn btn-sm btn-secondary" @click="handleJump">GO</button>
        </div>
      </div>
    </div>

    <!-- 卡片视图 -->
    <div v-if="viewMode === 'card'" class="cards-container">
      <div v-if="filteredStatefulsets.length === 0" class="empty-state">
        <div class="empty-icon">🔍</div>
        <div class="empty-text">没有找到匹配的有状态集</div>
      </div>
      <div class="cards-grid">
        <div v-for="sts in paginatedStatefulsets" :key="sts.name" class="statefulset-card" :class="{ 'card-selected': isStatefulsetSelected(sts) }">
          <div v-if="batchMode" class="card-checkbox">
            <input type="checkbox" :checked="isStatefulsetSelected(sts)" @change="toggleStatefulsetSelection(sts)" />
          </div>
          <div class="card-header">
            <div class="card-title-row">
              <span class="card-icon">🏛️</span>
              <h3 class="card-title">{{ sts.name }}</h3>
              <span class="status-indicator" :class="sts.status.toLowerCase()">{{ sts.status }}</span>
            </div>
            <span class="namespace-badge">{{ sts.namespace }}</span>
          </div>
          <div class="card-body">
            <div class="card-section">
              <div class="section-label">副本数</div>
              <div class="replicas-info">
                <div class="replicas-control">
                  <button class="replica-btn minus" @click="decreaseReplicas(sts)" :disabled="sts.desiredReplicas <= 0 || scalingMap[sts.name]">−</button>
                  
                  <!-- 内联编辑副本数 -->
                  <div v-if="inlineEdit.key === `replicas-${sts.name}`" class="replicas-edit-wrapper">
                    <input 
                      type="number" 
                      v-model="inlineEdit.value" 
                      class="replicas-input"
                      min="0"
                      @blur="saveInlineReplicas(sts)"
                      @keyup.enter="saveInlineReplicas(sts)"
                      @keyup.escape="cancelInlineEdit"
                      placeholder="副本数"
                      autofocus
                    />
                    <span class="inline-hint-small">回车</span>
                  </div>
                  
                  <!-- 显示副本数（可点击编辑） -->
                  <div v-else class="replicas-display clickable" 
                       :class="{ updating: scalingMap[sts.name] }" 
                       @click="startInlineReplicas(sts)"
                       title="点击修改副本数">
                    <span class="ready-replicas">{{ sts.readyReplicas }}</span>
                    <span class="replicas-sep">/</span>
                    <span class="desired-replicas">{{ sts.desiredReplicas }}</span>
                    <span v-if="scalingMap[sts.name]" class="scaling-indicator">⏳</span>
                    <span class="edit-icon-small">✏️</span>
                  </div>
                  
                  <button class="replica-btn plus" @click="increaseReplicas(sts)" :disabled="scalingMap[sts.name]">+</button>
                  <button class="replica-btn stop" @click="stopService(sts)" :disabled="sts.desiredReplicas === 0 || scalingMap[sts.name]">⏸️</button>
                </div>
                <div class="replicas-bar">
                  <div class="replicas-fill" :style="{ width: `${(sts.readyReplicas / Math.max(sts.desiredReplicas, 1)) * 100}%` }"></div>
                </div>
              </div>
            </div>
            <div class="card-section">
              <div class="section-label">镜像</div>
              <div class="image-text clickable" @click="startInlineImage(sts)">
                <span class="image-name">{{ sts.image || '-' }}</span>
                <span class="edit-icon">✏️</span>
              </div>
            </div>
            <div class="card-section">
              <div class="section-label">选择器</div>
              <div class="selector-tags">
                <span v-for="(value, key) in sts.selector" :key="key" class="selector-tag">{{ key }}={{ value }}</span>
              </div>
            </div>
            <div class="card-section card-section-row">
              <div class="card-meta-item">
                <div class="meta-label">更新策略</div>
                <span class="strategy-badge">{{ sts.updateStrategy }}</span>
              </div>
              <div class="card-meta-item">
                <div class="meta-label">创建时间</div>
                <div class="meta-value">{{ sts.createdAt }}</div>
              </div>
            </div>
            <!-- 资源消耗（Pod 聚合） -->
            <div class="card-section">
              <div class="section-label">
                资源使用
                <span v-if="sts.metrics">（{{ sts.metrics.podCount }} 个 Pod）</span>
              </div>
              <!-- 有 metrics 数据时显示具体值 -->
              <div v-if="sts.metrics" class="metrics-summary">
                <div class="metric-item">
                  <span class="metric-icon">⚡</span>
                  <span class="metric-label">CPU:</span>
                  <span class="metric-value">{{ sts.metrics.totalCpu }}</span>
                </div>
                <div class="metric-item">
                  <span class="metric-icon">💾</span>
                  <span class="metric-label">内存:</span>
                  <span class="metric-value">{{ sts.metrics.totalMemory }}</span>
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
                  <span v-if="sts.status === 'Stopped' || sts.desiredReplicas === 0">已停服</span>
                  <span v-else-if="sts.readyReplicas === 0">无就绪 Pod</span>
                  <span v-else>暂无数据</span>
                </div>
              </div>
            </div>
          </div>
          <div class="card-footer">
            <button class="card-action-btn primary" @click="viewPods(sts)" title="查看此 StatefulSet 管理的所有 Pod">
              📦 Pod 关联
            </button>
            <button class="card-action-btn" @click="viewStatefulsetLogs(sts)" title="查看日志">
              📄 日志
            </button>
            <button class="card-action-btn" @click="viewHistory(sts)" title="版本记录">
              📜 版本
            </button>
            <button class="card-action-btn" @click="openYamlPreview(sts)" title="查看/编辑 YAML">
              📝 YAML
            </button>
            <button class="card-action-btn danger" @click="deleteStatefulset(sts)" title="删除">
              🗑️ 删除
            </button>
          </div>
        </div>
      </div>
      <!-- 分页 + 跳转 -->
      <div v-if="filteredStatefulsets.length > 0" class="pagination-wrapper">
        <Pagination v-model:currentPage="currentPage" :totalItems="filteredStatefulsets.length" :itemsPerPage="itemsPerPage" />
        <div class="jump-page">
          <span>跳至</span>
          <input type="number" v-model.number="jumpPage" min="1" :max="totalPages" @keyup.enter="handleJump" />
          <span>页</span>
          <button class="btn btn-sm btn-secondary" @click="handleJump">GO</button>
        </div>
      </div>
    </div>

    <!-- 创建有状态集模态框 -->
    <div v-if="showCreateModal" class="modal-overlay" @click.self="showCreateModal = false">
      <div 
        ref="createModalRef"
        class="modal-content modal-create-statefulset resizable-modal"
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
          <h2>🏛️ 创建新有状态集</h2>
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
          <form @submit.prevent="createStatefulset">
            <div class="form-section">
              <div class="section-header"><span class="section-icon">📋</span><h3>基本信息</h3></div>
              <div class="section-body">
                <div class="form-group">
                  <label for="stsName">有状态集名称 <span class="required">*</span></label>
                  <input type="text" id="stsName" v-model="statefulsetForm.name" class="form-input" required placeholder="输入有状态集名称" />
                </div>
                <div class="form-group">
                  <label for="stsNamespace">命名空间 <span class="required">*</span></label>
                  <div class="namespace-selector">
                    <select 
                      v-if="!showNamespaceInput" 
                      id="stsNamespace" 
                      v-model="statefulsetForm.namespace" 
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
                    <label for="stsReplicas">副本数 <span class="required">*</span></label>
                    <input type="number" id="stsReplicas" v-model="statefulsetForm.replicas" class="form-input" required min="0" placeholder="1" />
                  </div>
                  <div class="form-group">
                    <label for="stsStrategy">更新策略</label>
                    <select id="stsStrategy" v-model="statefulsetForm.updateStrategy" class="form-select">
                      <option value="RollingUpdate">滚动更新 (RollingUpdate)</option>
                      <option value="OnDelete">手动触发 (OnDelete)</option>
                    </select>
                  </div>
                </div>
              </div>
            </div>
            <div class="form-section">
              <div class="section-header"><span class="section-icon">🐳</span><h3>容器镜像</h3></div>
              <div class="section-body">
                <div class="form-group">
                  <label for="stsImage">镜像地址 <span class="required">*</span></label>
                  <input type="text" id="stsImage" v-model="statefulsetForm.image" class="form-input" required placeholder="例如: nginx:latest" />
                </div>
              </div>
            </div>
            <div class="form-section">
              <div class="section-header"><span class="section-icon">🏷️</span><h3>标签与选择器</h3></div>
              <div class="section-body">
                <div class="form-group">
                  <label>Pod 标签 (Labels) <span class="required">*</span></label>
                  <div class="labels-editor">
                    <div v-for="(label, index) in statefulsetForm.labels" :key="index" class="label-row">
                      <input v-model="label.key" class="label-input label-key" placeholder="键，例如: app" />
                      <span class="label-separator">=</span>
                      <input v-model="label.value" class="label-input label-value" placeholder="值，例如: nginx" />
                      <button type="button" class="btn-icon btn-remove" @click="removeLabel(index)" :disabled="statefulsetForm.labels.length === 1">🗑️</button>
                    </div>
                    <button type="button" class="btn btn-secondary btn-sm" @click="addLabel">➕ 添加标签</button>
                  </div>
                </div>
              </div>
            </div>
            <div class="form-section">
              <div class="section-header"><span class="section-icon">🌐</span><h3>Headless Service</h3></div>
              <div class="section-body">
                <div class="form-group">
                  <label class="checkbox-label">
                    <input type="checkbox" v-model="statefulsetForm.createService" />
                    <span>同时创建 Headless Service</span>
                  </label>
                </div>
                <div v-if="statefulsetForm.createService" class="form-group">
                  <label for="stsServiceName">Service 名称</label>
                  <input type="text" id="stsServiceName" v-model="statefulsetForm.serviceName" class="form-input" placeholder="留空则使用 StatefulSet 名称" />
                </div>
              </div>
            </div>
            
            <!-- VolumeClaimTemplate 存储配置（StatefulSet 特有） -->
            <div class="form-section">
              <div class="section-header">
                <span class="section-icon">💾</span>
                <h3>持久化存储 (VolumeClaimTemplates)</h3>
              </div>
              <div class="section-body">
                <div class="form-hint storage-hint">
                  StatefulSet 支持为每个 Pod 自动创建独立的 PVC，适用于数据库等有状态应用。支持 Ceph、NFS、AWS EBS/EFS、S3 等多种存储后端。
                </div>
                
                <div v-if="statefulsetForm.volumeClaimTemplates.length === 0" class="empty-volumes">
                  <span class="empty-text">未配置存储卷</span>
                  <button type="button" class="btn btn-secondary btn-sm" @click="addVolumeClaimTemplate">
                    ➕ 添加存储卷
                  </button>
                </div>
                
                <div v-for="(vct, idx) in statefulsetForm.volumeClaimTemplates" :key="idx" class="volume-claim-item">
                  <div class="volume-header">
                    <span class="volume-index">存储卷 #{{ idx + 1 }}</span>
                    <button type="button" class="btn-icon btn-remove" @click="removeVolumeClaimTemplate(idx)">🗑️</button>
                  </div>
                  <div class="volume-body">
                    <!-- 存储类型选择器（卡片式） -->
                    <div class="form-group">
                      <label>存储类型</label>
                      <div class="storage-type-selector">
                        <label 
                          v-for="stype in storageTypeOptions" 
                          :key="stype.value" 
                          class="storage-type-card"
                          :class="{ active: vct.storageType === stype.value }"
                        >
                          <input 
                            type="radio" 
                            :value="stype.value" 
                            v-model="vct.storageType"
                            @change="onStorageTypeChange(vct)"
                            style="display: none;"
                          />
                          <div class="stype-icon">{{ stype.icon }}</div>
                          <div class="stype-name">{{ stype.label }}</div>
                        </label>
                      </div>
                      <div class="storage-type-desc" v-if="vct.storageType">
                        {{ storageTypeOptions.find(t => t.value === vct.storageType)?.desc }}
                      </div>
                    </div>
                    
                    <div class="form-row">
                      <div class="form-group">
                        <label>卷名称 <span class="required">*</span></label>
                        <input type="text" v-model="vct.name" class="form-input" placeholder="例如: data, logs" />
                      </div>
                      <div class="form-group">
                        <label>挂载路径 <span class="required">*</span></label>
                        <input type="text" v-model="vct.mountPath" class="form-input" placeholder="例如: /var/lib/mysql" />
                      </div>
                    </div>
                    <div class="form-row">
                      <div class="form-group">
                        <label>存储类 (StorageClass)</label>
                        <select v-model="vct.storageClass" class="form-select">
                          <option value="">使用默认存储类</option>
                          <option v-for="sc in storageClasses" :key="sc" :value="sc">{{ sc }}</option>
                        </select>
                        <div class="form-hint" v-if="storageClasses.length === 0">
                          提示：未检测到 StorageClass，请确保集群已配置存储供应商
                        </div>
                      </div>
                      <div class="form-group">
                        <label>访问模式</label>
                        <select v-model="vct.accessMode" class="form-select">
                          <option 
                            v-for="am in getRecommendedAccessModes(vct.storageType)" 
                            :key="am.value" 
                            :value="am.value"
                          >{{ am.label }}</option>
                        </select>
                      </div>
                    </div>
                    <div class="form-row">
                      <div class="form-group storage-size-group">
                        <label>存储大小 <span class="required">*</span></label>
                        <div class="storage-size-input">
                          <input type="number" v-model="vct.storageSize" class="form-input" min="1" placeholder="1" />
                          <select v-model="vct.storageSizeUnit" class="form-select size-unit">
                            <option value="Mi">Mi</option>
                            <option value="Gi">Gi</option>
                            <option value="Ti">Ti</option>
                          </select>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
                
                <button v-if="statefulsetForm.volumeClaimTemplates.length > 0" type="button" class="btn btn-secondary btn-sm" @click="addVolumeClaimTemplate">
                  ➕ 添加更多存储卷
                </button>
              </div>
            </div>
          </form>
          </div>
          
          <!-- YAML 模式 -->
          <div v-if="createMode === 'yaml'" class="yaml-editor-container">
            <div class="yaml-editor-header">
              <h3>📄 YAML 配置</h3>
              <p class="yaml-hint">✨ 支持多资源 YAML 创建（用 <code>---</code> 分隔），可同时创建 PVC、ConfigMap、Service 等依赖资源</p>
              <div class="yaml-header-buttons">
                <button class="load-template-btn" @click="loadStatefulSetYamlTemplate">
                  📑 加载模板（PVC + StatefulSet）
                </button>
                <button class="clear-yaml-btn" @click="clearYamlContent">
                  🗑️ 清除
                </button>
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
                  <li>支持完整的 Kubernetes StatefulSet 配置</li>
                  <li>可以通过“加载模板”获取示例 YAML</li>
                  <li>创建前会验证 YAML 格式的正确性</li>
                  <li>StatefulSet 支持 volumeClaimTemplates 和 headless service</li>
                </ul>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <!-- 表单模式按钮 -->
          <template v-if="createMode === 'form'">
            <button type="button" class="btn btn-secondary" @click="showCreateModal = false">取消</button>
            <button type="button" class="btn btn-primary" @click="createStatefulset"><span class="btn-icon">🏛️</span>创建有状态集</button>
          </template>
            
          <!-- YAML 模式按钮 -->
          <template v-if="createMode === 'yaml'">
            <button type="button" class="btn btn-secondary" @click="showCreateModal = false">取消</button>
            <button 
              type="button"
              class="btn btn-primary" 
              @click="createStatefulSetFromYaml"
              :disabled="!yamlContent"
            >
              <span class="btn-icon">🏛️</span>从 YAML 创建
            </button>
          </template>
        </div>
      </div>
    </div>

    <!-- 详情模态框 -->
    <div v-if="showViewModal" class="modal-overlay" @click.self="showViewModal = false">
      <div class="modal-content">
        <div class="modal-header">
          <h2>有状态集详情</h2>
          <button class="close-btn" @click="showViewModal = false">&times;</button>
        </div>
        <div class="modal-body">
          <div v-if="loadingDetail" class="loading-state">加载中...</div>
          <div v-else-if="detailData">
            <div class="detail-group"><label>名称:</label><span>{{ detailData.name }}</span></div>
            <div class="detail-group"><label>命名空间:</label><span>{{ detailData.namespace }}</span></div>
            <div class="detail-group"><label>状态:</label><span>{{ detailData.status }}</span></div>
            <div class="detail-group"><label>副本数:</label><span>{{ detailData.ready_replicas || 0 }}/{{ detailData.replicas || 0 }}</span></div>
            <div class="detail-group"><label>镜像:</label><span>{{ detailData.image || '-' }}</span></div>
            <div class="detail-group"><label>更新策略:</label><span>{{ detailData.update_strategy }}</span></div>
            <div class="detail-group"><label>服务名称:</label><span>{{ detailData.service_name || '-' }}</span></div>
            <div class="detail-group"><label>创建时间:</label><span>{{ detailData.created_at }}</span></div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-primary" @click="showViewModal = false">关闭</button>
        </div>
      </div>
    </div>

    <!-- 事件模态框 -->
    <div v-if="showEventsModal" class="modal-overlay" @click.self="showEventsModal = false">
      <div class="modal-content modal-large">
        <div class="modal-header">
          <h2>事件列表</h2>
          <button class="close-btn" @click="showEventsModal = false">&times;</button>
        </div>
        <div class="modal-body">
          <div v-if="loadingEvents" class="loading-state">加载中...</div>
          <div v-else-if="eventsData.length === 0" class="empty-state">暂无事件</div>
          <table v-else class="events-table">
            <thead>
              <tr><th>类型</th><th>原因</th><th>对象</th><th>消息</th><th>时间</th></tr>
            </thead>
            <tbody>
              <tr v-for="(event, idx) in eventsData" :key="idx">
                <td><span class="event-type" :class="event.type?.toLowerCase()">{{ event.type }}</span></td>
                <td>{{ event.reason }}</td>
                <td>{{ event.involved_object_name || event.object }}</td>
                <td>{{ event.message }}</td>
                <td>{{ fmtTime(event.last_timestamp || event.time) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <div class="modal-footer">
          <button class="btn btn-primary" @click="showEventsModal = false">关闭</button>
        </div>
      </div>
    </div>

    <!-- 历史版本模态框 -->
    <div v-if="showHistoryModal" class="modal-overlay" @click.self="showHistoryModal = false">
      <div class="modal-content modal-lg">
        <div class="modal-header">
          <h3>📜 版本记录</h3>
          <button class="close-btn" @click="showHistoryModal = false">×</button>
        </div>
        <div class="modal-body">
          <div v-if="historyStatefulset" class="info-box" style="margin-bottom: 16px;">
            <div><strong>StatefulSet:</strong> {{ historyStatefulset.name }}</div>
            <div><strong>命名空间:</strong> {{ historyStatefulset.namespace }}</div>
            <div><strong>当前版本:</strong> {{ historyStatefulset.revision || '-' }}</div>
          </div>
          
          <div v-if="loadingHistory" class="loading-state">加载版本历史...</div>
          <div v-else-if="historyList.length > 0">
            <table class="simple-table">
              <thead>
                <tr>
                  <th style="width: 100px;">版本号</th>
                  <th>ControllerRevision 名称</th>
                  <th style="width: 180px;">创建时间</th>
                  <th style="width: 150px;">操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="rev in historyList" :key="rev.name">
                  <td>
                    <span class="version-badge">
                      {{ rev.revision }}
                      <span v-if="rev.revision === historyStatefulset?.revision" class="current-tag">当前</span>
                    </span>
                  </td>
                  <td class="mono">{{ rev.name }}</td>
                  <td>{{ fmtTime(rev.creation_time) }}</td>
                  <td>
                    <span v-if="rev.revision === historyStatefulset?.revision" class="current-version-text">当前运行版本</span>
                    <span v-else>-</span>
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

    <!-- Pod 关联模态框 -->
    <div v-if="showPodsModal" class="modal-overlay" @click.self="showPodsModal = false">
      <div class="modal-content modal-lg">
        <div class="modal-header">
          <h3>📦 关联 Pods</h3>
          <button class="close-btn" @click="showPodsModal = false">×</button>
        </div>
        <div class="modal-body">
          <div class="info-box">
            <div><strong>StatefulSet:</strong> {{ podsStatefulset?.name }}</div>
            <div><strong>命名空间:</strong> {{ podsStatefulset?.namespace }}</div>
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

    <!-- Pod 日志弹窗 -->
    <div v-if="showPodLogsModal" class="modal-overlay" @click.self="closePodLogsModal">
      <div class="modal-content logs-modal">
        <div class="modal-header">
          <h3>📄 Pod 日志</h3>
          <button class="close-btn" @click="closePodLogsModal">×</button>
        </div>
        <div class="modal-body">
          <div v-if="selectedPodForAction" class="pod-info-bar">
            <span><strong>命名空间：</strong>{{ selectedPodForAction.namespace }}</span>
            <span><strong>Pod：</strong>{{ selectedPodForAction.name }}</span>
          </div>
          <div class="logs-controls">
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
            <div class="control-item">
              <label>行数</label>
              <select v-model="podLogsForm.tail" class="form-select">
                <option :value="100">100 行</option>
                <option :value="200">200 行</option>
                <option :value="500">500 行</option>
                <option :value="1000">1000 行</option>
              </select>
            </div>
            <div class="control-actions">
              <button class="btn btn-primary btn-sm" @click="fetchPodLogsContent" :disabled="loadingPodLogs">
                {{ loadingPodLogs ? '加载中...' : '获取日志' }}
              </button>
              <button class="btn btn-secondary btn-sm" @click="podLogsContent = ''" :disabled="!podLogsContent">清除</button>
            </div>
          </div>
          <div class="logs-content-wrapper">
            <div v-if="podLogsError" class="error-box">{{ podLogsError }}</div>
            <pre v-else class="logs-content">{{ podLogsContent }}</pre>
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
          <div v-if="selectedPodForAction" class="detail-content">
            <div class="detail-group"><label>名称:</label><span>{{ selectedPodForAction.name }}</span></div>
            <div class="detail-group"><label>命名空间:</label><span>{{ selectedPodForAction.namespace }}</span></div>
            <div class="detail-group"><label>状态:</label><span class="status-indicator" :class="selectedPodForAction.status?.toLowerCase()">{{ selectedPodForAction.status }}</span></div>
            <div class="detail-group"><label>节点:</label><span>{{ selectedPodForAction.node || '-' }}</span></div>
            <div class="detail-group"><label>重启次数:</label><span>{{ selectedPodForAction.restartCount || 0 }}</span></div>
            <div class="detail-group"><label>创建时间:</label><span>{{ selectedPodForAction.createdAt }}</span></div>
            <div class="detail-group"><label>容器:</label><span>{{ selectedPodForAction.containers?.join(', ') || '-' }}</span></div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-primary" @click="showPodDetailModal = false">关闭</button>
        </div>
      </div>
    </div>

    <!-- Pod 事件弹窗 -->
    <div v-if="showPodEventsModal" class="modal-overlay" @click.self="showPodEventsModal = false">
      <div class="modal-content modal-large">
        <div class="modal-header">
          <h3>📡 Pod 事件</h3>
          <button class="close-btn" @click="showPodEventsModal = false">×</button>
        </div>
        <div class="modal-body">
          <div v-if="selectedPodForAction" class="info-box" style="margin-bottom: 16px;">
            <div><strong>Pod:</strong> {{ selectedPodForAction.name }}</div>
            <div><strong>命名空间:</strong> {{ selectedPodForAction.namespace }}</div>
          </div>
          <div v-if="loadingPodEvents" class="loading-state">加载事件...</div>
          <div v-else-if="podEventsList.length > 0">
            <table class="events-table">
              <thead>
                <tr><th>类型</th><th>原因</th><th>消息</th><th>时间</th></tr>
              </thead>
              <tbody>
                <tr v-for="(evt, idx) in podEventsList" :key="idx">
                  <td><span class="event-type" :class="evt.type?.toLowerCase()">{{ evt.type }}</span></td>
                  <td>{{ evt.reason }}</td>
                  <td>{{ evt.message }}</td>
                  <td style="white-space: nowrap;">{{ evt.lastTimestamp || evt.time }}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <div v-else class="empty-state-small">
            <div class="empty-text">暂无事件</div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-primary" @click="showPodEventsModal = false">关闭</button>
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
            <div><strong>有状态集:</strong> {{ updateImageForm.name }}</div>
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

    <!-- 删除确认模态框 -->
    <div v-if="showDeleteModal" class="modal-overlay" @click.self="showDeleteModal = false">
      <div class="modal-content">
        <div class="modal-header">
          <h2>确认删除</h2>
          <button class="close-btn" @click="showDeleteModal = false">&times;</button>
        </div>
        <div class="modal-body">
          <p>您确定要删除有状态集 <strong>{{ stsToDelete?.name }}</strong> 吗？此操作无法撤销。</p>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showDeleteModal = false">取消</button>
          <button class="btn btn-danger" @click="confirmDelete" :disabled="deleting">{{ deleting ? '删除中...' : '删除' }}</button>
        </div>
      </div>
    </div>

    <!-- 批量重启预览弹窗 -->
    <div v-if="showBatchRestartModal" class="modal-overlay" @click.self="closeBatchRestartModal">
      <div class="modal-content modal-batch-preview modal-warning">
        <div class="modal-header warning-header">
          <h3>🔄 批量重启预览（高危）</h3>
          <button class="close-btn" @click="closeBatchRestartModal">×</button>
        </div>
        <div class="modal-body">
          <div class="warning-box">
            <div class="warning-icon-large">⚠️</div>
            <div class="warning-content">
              <div class="warning-title">将重启以下 StatefulSet</div>
              <ul class="warning-list">
                <li>重启会滚动更新所有 Pod</li>
                <li>可能导致服务短暂不可用</li>
                <li>请确认业务可以承受重启影响</li>
              </ul>
            </div>
          </div>
          <div class="preview-section">
            <div class="section-title">受影响 StatefulSet ({{ selectedStatefulsets.length }})</div>
            <div class="affected-deployments-detail">
              <div v-for="sts in selectedStatefulsets" :key="sts.name" class="affected-dep-card">
                <div class="dep-info">
                  <span class="dep-name">🏛️ {{ sts.name }}</span>
                  <span class="dep-namespace">{{ sts.namespace }}</span>
                </div>
                <div class="dep-stats">
                  <span class="status-indicator" :class="sts.status.toLowerCase()">{{ sts.status }}</span>
                  <span class="replicas-tag">{{ sts.readyReplicas }}/{{ sts.desiredReplicas }}</span>
                </div>
              </div>
            </div>
          </div>
          <div class="confirm-section">
            <div class="section-title">请输入 "RESTART" 确认操作</div>
            <input v-model="restartConfirmText" placeholder="请输入 RESTART" class="confirm-input" :class="{ valid: restartConfirmText === 'RESTART' }" />
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeBatchRestartModal">取消</button>
          <button class="btn btn-warning" @click="executeBatchRestart" :disabled="restartConfirmText !== 'RESTART' || batchExecuting">{{ batchExecuting ? '执行中...' : '确认重启' }}</button>
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
          <div class="preview-section">
            <div class="section-title">受影响 StatefulSet ({{ selectedStatefulsets.length }})</div>
            <div class="affected-deployments">
              <div v-for="sts in selectedStatefulsets" :key="sts.name" class="affected-item">
                <span class="dep-name">🏛️ {{ sts.name }}</span>
                <span class="dep-replicas">当前: {{ sts.readyReplicas }}/{{ sts.desiredReplicas }}</span>
              </div>
            </div>
          </div>
          <div class="preview-section">
            <div class="section-title">目标副本数</div>
            <input type="number" v-model="batchScaleReplicas" min="0" class="form-input" style="width: 120px;" />
          </div>
          <div class="preview-section">
            <div class="section-title">变更预览</div>
            <div class="change-preview">
              <div v-for="sts in selectedStatefulsets" :key="sts.name" class="change-item">
                <span>{{ sts.name }}:</span>
                <span class="old-value">{{ sts.desiredReplicas }}</span>
                <span class="arrow">→</span>
                <span class="new-value">{{ batchScaleReplicas }}</span>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeBatchScaleModal">取消</button>
          <button class="btn btn-primary" @click="executeBatchScale" :disabled="batchExecuting">{{ batchExecuting ? '执行中...' : '确认扩缩容' }}</button>
        </div>
      </div>
    </div>

    <!-- 批量删除预览弹窗 -->
    <div v-if="showBatchDeleteModal" class="modal-overlay" @click.self="closeBatchDeleteModal">
      <div class="modal-content modal-batch-preview modal-danger">
        <div class="modal-header danger-header">
          <h3>🗑️ 批量删除预览（高危操作）</h3>
          <button class="close-btn" @click="closeBatchDeleteModal">×</button>
        </div>
        <div class="modal-body">
          <div class="danger-warning">
            <div class="warning-icon-large">⚠️</div>
            <div class="warning-content">
              <div class="warning-title">将删除以下 StatefulSet 及其所有 Pod</div>
              <ul class="warning-list">
                <li>所有关联的 Pod 和 PVC 将被删除</li>
                <li>此操作不可撤销！</li>
              </ul>
            </div>
          </div>
          <div class="preview-section">
            <div class="section-title">受影响 StatefulSet ({{ selectedStatefulsets.length }})</div>
            <div class="affected-deployments-detail">
              <div v-for="sts in selectedStatefulsets" :key="sts.name" class="affected-dep-card">
                <div class="dep-info">
                  <span class="dep-name">🏛️ {{ sts.name }}</span>
                  <span class="dep-namespace">{{ sts.namespace }}</span>
                </div>
                <div class="dep-stats">
                  <span class="status-indicator" :class="sts.status.toLowerCase()">{{ sts.status }}</span>
                  <span class="replicas-tag">{{ sts.readyReplicas }}/{{ sts.desiredReplicas }}</span>
                </div>
              </div>
            </div>
          </div>
          <div class="confirm-section">
            <div class="section-title">请输入 "DELETE" 确认操作</div>
            <input v-model="deleteConfirmText" placeholder="请输入 DELETE" class="confirm-input" :class="{ valid: deleteConfirmText === 'DELETE' }" />
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeBatchDeleteModal">取消</button>
          <button class="btn btn-danger" @click="executeBatchDelete" :disabled="deleteConfirmText !== 'DELETE' || batchExecuting">{{ batchExecuting ? '执行中...' : '确认删除' }}</button>
        </div>
      </div>
    </div>

    <!-- StatefulSet 日志弹窗 -->
    <div v-if="showLogsModal" class="modal-overlay" @click.self="closeLogsModal">
      <div class="modal-content logs-modal">
        <div class="modal-header">
          <h3>📄 StatefulSet 日志</h3>
          <button class="close-btn" @click="closeLogsModal">×</button>
        </div>
        <div class="modal-body">
          <div v-if="selectedStatefulsetForLogs" class="info-box" style="margin-bottom: 16px;">
            <div><strong>StatefulSet:</strong> {{ selectedStatefulsetForLogs.name }}</div>
            <div><strong>命名空间:</strong> {{ selectedStatefulsetForLogs.namespace }}</div>
          </div>

          <!-- 日志控制面板 -->
          <div class="logs-controls">
            <div class="control-item">
              <label>Pod 选择</label>
              <select v-model="logsForm.selectedPod" class="form-select" @change="onLogsPodChange">
                <option value="">全部 Pod</option>
                <option v-for="pod in logsPodsList" :key="pod.name" :value="pod.name">
                  {{ pod.name }}
                </option>
              </select>
            </div>

            <div class="control-item" v-if="logsForm.selectedPod">
              <label>容器</label>
              <select v-model="logsForm.container" class="form-select">
                <option value="" disabled>选择容器</option>
                <option v-for="c in logsContainerList" :key="c" :value="c">
                  {{ c }}
                </option>
              </select>
            </div>

            <div class="control-item">
              <label>行数</label>
              <select v-model="logsForm.tail" class="form-select">
                <option :value="null">全部</option>
                <option :value="10">10 行</option>
                <option :value="50">50 行</option>
                <option :value="100">100 行</option>
                <option :value="200">200 行</option>
                <option :value="500">500 行</option>
                <option :value="1000">1000 行</option>
              </select>
            </div>

            <div class="control-actions">
              <button 
                class="btn btn-primary btn-sm" 
                @click="fetchLogs" 
                :disabled="loadingLogs"
              >
                {{ loadingLogs ? '获取中...' : '获取日志' }}
              </button>
              <button 
                v-if="loadingLogs" 
                class="btn btn-danger btn-sm" 
                @click="stopLogLoading"
              >
                终止
              </button>
              <button 
                class="btn btn-secondary btn-sm" 
                @click="clearLogs"
                :disabled="!logsContent"
              >
                清除
              </button>
            </div>
          </div>

          <!-- 日志内容区 -->
          <div class="logs-content-wrapper">
            <div v-if="logsError" class="error-box">{{ logsError }}</div>
            <pre v-else class="logs-content">{{ logsContent }}</pre>
          </div>
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
          <h3>📝 {{ yamlEditMode ? '编辑 YAML' : '查看 YAML' }}</h3>
          <div class="yaml-header-actions">
            <button v-if="!yamlEditMode" class="btn btn-sm btn-secondary" @click="downloadYaml" :disabled="loadingYaml">
              📄 下载
            </button>
            <button v-if="!yamlEditMode" class="btn btn-sm btn-secondary" @click="yamlEditMode = true" :disabled="loadingYaml">
              ✈️ 编辑模式
            </button>
            <button v-if="yamlEditMode" class="btn btn-sm btn-secondary" @click="yamlEditMode = false">
              👁️ 预览模式
            </button>
            <button class="close-btn" @click="closeYamlModal">×</button>
          </div>
        </div>
        <div class="modal-body yaml-modal-body">
          <div v-if="selectedYamlStatefulset" class="info-box" style="margin-bottom: 16px;">
            <div><strong>StatefulSet:</strong> {{ selectedYamlStatefulset.name }}</div>
            <div><strong>命名空间:</strong> {{ selectedYamlStatefulset.namespace }}</div>
          </div>
          
          <div v-if="loadingYaml" class="loading-state">Loading YAML...</div>
          <div v-else-if="yamlError" class="error-box">{{ yamlError }}</div>
          <div v-else class="yaml-editor-wrapper">
            <textarea 
              v-if="yamlEditMode"
              v-model="yamlContent" 
              class="yaml-editor"
              spellcheck="false"
              placeholder="YAML 内容..."
            ></textarea>
            <pre v-else class="yaml-content">{{ yamlContent }}</pre>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeYamlModal">取消</button>
          <button v-if="yamlEditMode" class="btn btn-primary" @click="applyYamlChanges" :disabled="savingYaml || !yamlContent">
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
import statefulsetsApi from '@/api/cluster/workloads/statefulsets'
import podsApi from '@/api/cluster/workloads/pods'
import namespaceApi from '@/api/cluster/config/namespace'
import storageclassApi from '@/api/cluster/storage/storageclass'
import { useClusterStore } from '@/stores/cluster'
import { useResizableModal } from '@/composables/useResizableModal'

// ===== 获取认证头 =====
const getAuthHeaders = () => {
  const headers = { 'Content-Type': 'application/json' }
  const token = localStorage.getItem('token') || sessionStorage.getItem('token')
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }
  const clusterStore = useClusterStore()
  const getClusterIdFromPath = () => {
    try {
      const m = window.location.pathname.match(/\/c\/([^/]+)/)
      return m ? decodeURIComponent(m[1]) : ''
    } catch { return '' }
  }
  const cid = clusterStore.current?.id ?? getClusterIdFromPath()
  if (cid !== undefined && cid !== null && cid !== '') {
    headers['X-Cluster-ID'] = String(cid)
  }
  return headers
}

// 状态变量
const loading = ref(false)
const errorMsg = ref('')
const searchQuery = ref('')
const statusFilter = ref('all')
const namespaceFilter = ref('')
const currentPage = ref(1)
const itemsPerPage = ref(10)
const namespaces = ref([])
const statefulsets = ref([])
const totalFromServer = ref(0)  // 后端返回的总数
const viewMode = ref('table')

// 批量操作相关
const batchMode = ref(false)
const selectedStatefulsets = ref([])
const showBatchScaleModal = ref(false)
const showBatchDeleteModal = ref(false)
const showBatchRestartModal = ref(false)
const batchScaleReplicas = ref(1)
const deleteConfirmText = ref('')
const restartConfirmText = ref('')
const batchExecuting = ref(false)

// 自动刷新
const autoRefresh = ref(false)
let autoRefreshTimer = null
const AUTO_REFRESH_INTERVAL = 90000 // 90秒

// 更多菜单
const showMoreOptions = ref(false)
const selectedStatefulset = ref(null)
const menuStyle = ref({})

// 卡片视图更多菜单
const showCardMoreOptions = ref(false)
const selectedCardStatefulset = ref(null)
const cardMenuStyle = ref({})

// Modal states
const showCreateModal = ref(false)
const showDeleteModal = ref(false)
const showViewModal = ref(false)
const showEventsModal = ref(false)
const showUpdateImageModal = ref(false)
const showHistoryModal = ref(false)
const showPodsModal = ref(false)
const showLogsModal = ref(false)
const showYamlModal = ref(false)

// ========== 可拖拽调整大小的模态框 ==========
const {
  modalRef: createModalRef,
  modalStyle: createModalStyle,
  startResize: createStartResize
} = useResizableModal({ initialWidth: '1200px', initialHeight: '80vh' })

const {
  modalRef: yamlModalRef,
  modalStyle: yamlModalStyle,
  startResize: yamlStartResize
} = useResizableModal({ initialWidth: '1100px', initialHeight: '80vh' })

// Loading states
const loadingDetail = ref(false)
const loadingEvents = ref(false)
const loadingHistory = ref(false)
const loadingPods = ref(false)
const updatingImage = ref(false)
const deleting = ref(false)
const loadingLogs = ref(false)
const loadingYaml = ref(false)
const savingYaml = ref(false)

// Data
const detailData = ref(null)
const eventsData = ref([])
const historyList = ref([])
const historyStatefulset = ref(null)
const podsList = ref([])
const containerList = ref([])
const podsStatefulset = ref(null)

// Pod 操作相关
const showPodMoreOptions = ref(false)
const selectedPodForAction = ref(null)
const podMenuStyle = ref({})
const showPodLogsModal = ref(false)
const showPodDetailModal = ref(false)
const showPodEventsModal = ref(false)
const loadingPodLogs = ref(false)
const loadingPodEvents = ref(false)
const podLogsContent = ref('')
const podLogsError = ref('')
const podEventsList = ref([])
const podContainerList = ref([])
const podLogsForm = ref({ container: '', tail: 100 })

// 日志相关
const selectedStatefulsetForLogs = ref(null)
const logsPodsList = ref([])
const logsContainerList = ref([])
const logsContent = ref('')
const logsError = ref('')
let logAbortController = null
let logStreamingTimer = null
const isStreamingLogs = ref(false)
const logsForm = ref({
  selectedPod: '',
  container: '',
  tail: 100,
  follow: false,        // 实时日志开关
  duration: 60          // 实时日志保持时长（秒）
})

// YAML 编辑相关
const selectedYamlStatefulset = ref(null)
const yamlContent = ref('')
const yamlError = ref('')
const yamlEditMode = ref(false)

// YAML 创建相关
const createMode = ref('form') // 'form' | 'yaml'

// 监听 createMode 变化，切换到 YAML 模式时如果内容为空则自动加载模板
watch(createMode, (newMode) => {
  if (newMode === 'yaml' && !yamlContent.value.trim()) {
    loadStatefulSetYamlTemplate()
  }
})

// 命名空间创建相关
const showNamespaceInput = ref(false)
const newNamespace = ref('')
const creatingNamespace = ref(false)

// 存储类相关
const storageClasses = ref([])
const accessModeOptions = [
  { value: 'ReadWriteOnce', label: 'ReadWriteOnce (单节点读写)' },
  { value: 'ReadOnlyMany', label: 'ReadOnlyMany (多节点只读)' },
  { value: 'ReadWriteMany', label: 'ReadWriteMany (多节点读写)' }
]

// 存储类型定义（参考 Rancher/Kuboard）
const storageTypeOptions = [
  { 
    value: 'default', 
    label: '默认存储', 
    icon: '💾',
    desc: '使用集群默认 StorageClass',
    accessModes: ['ReadWriteOnce'],
    provisioners: []
  },
  { 
    value: 'ceph-rbd', 
    label: 'Ceph RBD', 
    icon: '🔴',
    desc: '高性能块存储，适合数据库、虚拟机',
    accessModes: ['ReadWriteOnce'],
    provisioners: ['rbd.csi.ceph.com', 'kubernetes.io/rbd', 'ceph.com/rbd']
  },
  { 
    value: 'ceph-fs', 
    label: 'CephFS', 
    icon: '🟠',
    desc: '分布式文件系统，支持多节点共享读写',
    accessModes: ['ReadWriteOnce', 'ReadWriteMany'],
    provisioners: ['cephfs.csi.ceph.com', 'kubernetes.io/cephfs']
  },
  { 
    value: 'nfs', 
    label: 'NFS', 
    icon: '📁',
    desc: '网络文件系统，广泛兼容，适合共享存储',
    accessModes: ['ReadWriteOnce', 'ReadWriteMany', 'ReadOnlyMany'],
    provisioners: ['nfs.csi.k8s.io', 'kubernetes.io/nfs', 'nfs-client']
  },
  { 
    value: 'aws-ebs', 
    label: 'AWS EBS', 
    icon: '☁️',
    desc: 'Amazon 弹性块存储，适合 AWS 云环境',
    accessModes: ['ReadWriteOnce'],
    provisioners: ['ebs.csi.aws.com', 'kubernetes.io/aws-ebs']
  },
  { 
    value: 'aws-efs', 
    label: 'AWS EFS', 
    icon: '☁️',
    desc: 'Amazon 弹性文件系统，支持多可用区共享',
    accessModes: ['ReadWriteMany'],
    provisioners: ['efs.csi.aws.com']
  },
  { 
    value: 's3', 
    label: 'S3 对象存储', 
    icon: '🪣',
    desc: '对象存储挂载（需 s3fs/goofys），适合静态资源',
    accessModes: ['ReadWriteMany', 'ReadOnlyMany'],
    provisioners: ['s3.csi.aws.com', 'ch.ctrox.csi.s3-driver']
  },
  { 
    value: 'local', 
    label: '本地存储', 
    icon: '💽',
    desc: '节点本地磁盘，性能最佳但无高可用',
    accessModes: ['ReadWriteOnce'],
    provisioners: ['kubernetes.io/no-provisioner', 'local.csi.k8s.io']
  },
  { 
    value: 'hostpath', 
    label: 'HostPath', 
    icon: '📂',
    desc: '宿主机路径挂载，仅用于测试环境',
    accessModes: ['ReadWriteOnce'],
    provisioners: ['hostpath.csi.k8s.io', 'docker.io/hostpath']
  }
]

// Forms
const statefulsetForm = ref({
  name: '', namespace: 'default', replicas: 1, image: '',
  updateStrategy: 'RollingUpdate',
  labels: [{ key: 'app', value: '' }],
  createService: true, serviceName: '',
  // VolumeClaimTemplates - StatefulSet 特有的存储配置
  volumeClaimTemplates: []
})

const updateImageForm = ref({ namespace: '', name: '', container: '', image: '' })
const stsToDelete = ref(null)

// 内联编辑状态
const inlineEdit = ref({ key: '', value: null, original: null })
const scalingMap = ref({})

// 生命周期
onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('scroll', handleScroll, true)
  fetchNamespaces().then(fetchStatefulsets)
  fetchStorageClasses()
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('scroll', handleScroll, true)
  stopAutoRefresh()
})

// 更多菜单控制
const toggleMoreOptions = (sts, event) => {
  if (selectedStatefulset.value === sts && showMoreOptions.value) {
    showMoreOptions.value = false
    selectedStatefulset.value = null
  } else {
    selectedStatefulset.value = sts
    showMoreOptions.value = true
    const button = event.target.closest('.more-btn')
    if (button) {
      const rect = button.getBoundingClientRect()
      const viewportHeight = window.innerHeight
      const menuHeight = 320
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

// 卡片视图更多菜单控制
const toggleCardMoreOptions = (sts, event) => {
  if (selectedCardStatefulset.value === sts && showCardMoreOptions.value) {
    showCardMoreOptions.value = false
    selectedCardStatefulset.value = null
  } else {
    selectedCardStatefulset.value = sts
    showCardMoreOptions.value = true
    const button = event.target.closest('.card-more-btn')
    if (button) {
      const rect = button.getBoundingClientRect()
      const viewportHeight = window.innerHeight
      const menuHeight = 320
      let style = { position: 'fixed', left: rect.left + 'px', zIndex: 2000 }
      if (rect.bottom + menuHeight > viewportHeight) {
        style.bottom = (viewportHeight - rect.top + 4) + 'px'
      } else {
        style.top = (rect.bottom + 4) + 'px'
      }
      cardMenuStyle.value = style
    }
  }
}

const handleClickOutside = (event) => {
  if (showMoreOptions.value && !event.target.closest('.more-btn')) {
    showMoreOptions.value = false
    selectedStatefulset.value = null
  }
  if (showCardMoreOptions.value && !event.target.closest('.card-more-btn')) {
    showCardMoreOptions.value = false
    selectedCardStatefulset.value = null
  }
}

const handleScroll = () => {
  if (showMoreOptions.value) {
    showMoreOptions.value = false
    selectedStatefulset.value = null
  }
  if (showCardMoreOptions.value) {
    showCardMoreOptions.value = false
    selectedCardStatefulset.value = null
  }
}

// 自动刷新
const startAutoRefresh = () => {
  if (autoRefreshTimer) return
  autoRefreshTimer = setInterval(() => {
    if (!loading.value) fetchStatefulsets()
  }, AUTO_REFRESH_INTERVAL)
}

const stopAutoRefresh = () => {
  if (autoRefreshTimer) {
    clearInterval(autoRefreshTimer)
    autoRefreshTimer = null
  }
}

watch(autoRefresh, (val) => val ? startAutoRefresh() : stopAutoRefresh())
watch(namespaceFilter, () => { currentPage.value = 1; fetchStatefulsets() })

// API 调用
const fetchNamespaces = async () => {
  try {
    const res = await namespaceApi.list({ page: 1, limit: 1000 })
    const list = res?.data?.list || res?.data?.items || []
    namespaces.value = (Array.isArray(list) ? list : []).map(ns =>
      typeof ns === 'string' ? ns : (ns?.metadata?.name || ns?.name || ns)
    ).filter(Boolean)
    if (namespaces.value.length > 0 && !statefulsetForm.value.namespace) {
      statefulsetForm.value.namespace = namespaces.value[0]
    }
  } catch (e) {
    console.error('获取命名空间失败:', e)
    namespaces.value = ['default', 'kube-system']
  }
}

// 获取 StorageClass 列表
const fetchStorageClasses = async () => {
  try {
    const res = await storageclassApi.list({ page: 1, limit: 100 })
    const list = res?.data?.list || res?.data?.items || []
    storageClasses.value = (Array.isArray(list) ? list : []).map(sc => 
      typeof sc === 'string' ? sc : (sc?.metadata?.name || sc?.name || sc)
    ).filter(Boolean)
  } catch (e) {
    console.error('获取 StorageClass 失败:', e)
    storageClasses.value = []
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
      await fetchNamespaces()
      statefulsetForm.value.namespace = newNamespace.value.trim()
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

const fetchStatefulsets = async () => {
  loading.value = true
  errorMsg.value = ''
  try {
    // 使用后端分页和模糊查询，limit 限制为 200
    const params = { 
      namespace: namespaceFilter.value || '', 
      name: searchQuery.value.trim() || '',  // 模糊查询参数
      page: 1, 
      limit: 200 
    }
    const res = await statefulsetsApi.list(params)
    const list = res.data?.list || res.data?.items || []
    totalFromServer.value = res.data?.total || list.length  // 保存后端总数
    if (res.code === 0 && list.length > 0) {
      statefulsets.value = list.map(item => ({
        name: item.name,
        namespace: item.namespace,
        status: item.status || 'Unknown',
        desiredReplicas: item.replicas || 0,
        readyReplicas: item.ready_replicas || 0,
        image: item.image || (item.images && item.images[0]) || '',
        selector: item.selector || {},
        updateStrategy: item.update_strategy || 'RollingUpdate',
        serviceName: item.service_name || '',
        createdAt: item.created_at || '',
        containers: item.containers || []
      }))
    } else {
      statefulsets.value = []
    }
  } catch (e) {
    console.error('获取有状态集列表失败:', e)
    errorMsg.value = e?.msg || e?.message || '获取有状态集列表失败'
    statefulsets.value = []
  } finally {
    loading.value = false
  }
}

// 时间格式化
const fmtTime = (ts) => {
  if (!ts) return '-'
  try { return new Date(ts).toLocaleString('zh-CN') } catch { return ts }
}

// 筛选与搜索
const setStatusFilter = (status) => { statusFilter.value = status; currentPage.value = 1 }
let searchDebounceTimer = null
const onSearchInput = () => {
  if (searchDebounceTimer) clearTimeout(searchDebounceTimer)
  searchDebounceTimer = setTimeout(() => { 
    currentPage.value = 1
    fetchStatefulsets()  // 触发后端模糊查询
  }, 300)
}
const refreshList = () => fetchStatefulsets()

// 计算属性
const filteredStatefulsets = computed(() => {
  let result = statefulsets.value
  const q = searchQuery.value.trim().toLowerCase()
  if (q) {
    result = result.filter(d => d.name.toLowerCase().includes(q) || d.namespace.toLowerCase().includes(q) || d.image.toLowerCase().includes(q))
  }
  if (statusFilter.value !== 'all') {
    result = result.filter(d => d.status === statusFilter.value)
  }
  return result
})

const paginatedStatefulsets = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage.value
  return filteredStatefulsets.value.slice(start, start + itemsPerPage.value)
})

// 分页跳转
const jumpPage = ref(1)
const totalPages = computed(() => Math.ceil(filteredStatefulsets.value.length / itemsPerPage.value))
const handleJump = () => {
  const page = parseInt(jumpPage.value)
  if (page >= 1 && page <= totalPages.value) {
    currentPage.value = page
  } else if (page > totalPages.value) {
    currentPage.value = totalPages.value
    jumpPage.value = totalPages.value
  } else {
    currentPage.value = 1
    jumpPage.value = 1
  }
}

// 批量操作
const enterBatchMode = () => {
  batchMode.value = true
  selectedStatefulsets.value = [...paginatedStatefulsets.value]
}
const exitBatchMode = () => { batchMode.value = false; selectedStatefulsets.value = [] }
const clearSelection = () => { selectedStatefulsets.value = [] }
const isStatefulsetSelected = (sts) => selectedStatefulsets.value.some(d => d.name === sts.name && d.namespace === sts.namespace)
const toggleStatefulsetSelection = (sts) => {
  const index = selectedStatefulsets.value.findIndex(d => d.name === sts.name && d.namespace === sts.namespace)
  if (index >= 0) selectedStatefulsets.value.splice(index, 1)
  else selectedStatefulsets.value.push(sts)
}
const isAllSelected = computed(() => paginatedStatefulsets.value.length > 0 && paginatedStatefulsets.value.every(sts => isStatefulsetSelected(sts)))
const isPartialSelected = computed(() => {
  if (paginatedStatefulsets.value.length === 0) return false
  const selectedCount = paginatedStatefulsets.value.filter(sts => isStatefulsetSelected(sts)).length
  return selectedCount > 0 && selectedCount < paginatedStatefulsets.value.length
})
const selectAllCheckbox = ref(null)
watchEffect(() => { if (selectAllCheckbox.value) selectAllCheckbox.value.indeterminate = isPartialSelected.value })
const toggleSelectAll = () => {
  if (isAllSelected.value) {
    paginatedStatefulsets.value.forEach(sts => {
      const index = selectedStatefulsets.value.findIndex(d => d.name === sts.name && d.namespace === sts.namespace)
      if (index >= 0) selectedStatefulsets.value.splice(index, 1)
    })
  } else {
    paginatedStatefulsets.value.forEach(sts => { if (!isStatefulsetSelected(sts)) selectedStatefulsets.value.push(sts) })
  }
}

// 批量重启
const openBatchRestartPreview = () => { restartConfirmText.value = ''; showBatchRestartModal.value = true }
const closeBatchRestartModal = () => { showBatchRestartModal.value = false; restartConfirmText.value = '' }
const executeBatchRestart = async () => {
  if (restartConfirmText.value !== 'RESTART') return
  batchExecuting.value = true
  let successCount = 0, failCount = 0
  for (const sts of selectedStatefulsets.value) {
    try { await statefulsetsApi.restart({ name: sts.name, namespace: sts.namespace }); successCount++ }
    catch (e) { console.error(`Failed to restart ${sts.name}:`, e); failCount++ }
  }
  batchExecuting.value = false
  showBatchRestartModal.value = false
  restartConfirmText.value = ''
  if (failCount === 0) Message.success({ content: `成功重启 ${successCount} 个 StatefulSet`, duration: 2200 })
  else Message.warning({ content: `成功 ${successCount} 个，失败 ${failCount} 个`, duration: 2200 })
  exitBatchMode(); refreshList()
}

// 批量扩缩容
const openBatchScaleModal = () => { batchScaleReplicas.value = 1; showBatchScaleModal.value = true }
const closeBatchScaleModal = () => { showBatchScaleModal.value = false }
const executeBatchScale = async () => {
  batchExecuting.value = true
  let successCount = 0, failCount = 0
  for (const sts of selectedStatefulsets.value) {
    try { await statefulsetsApi.scale({ name: sts.name, namespace: sts.namespace, scale_num: parseInt(batchScaleReplicas.value) }); successCount++ }
    catch (e) { console.error(`Failed to scale ${sts.name}:`, e); failCount++ }
  }
  batchExecuting.value = false
  showBatchScaleModal.value = false
  if (failCount === 0) Message.success({ content: `成功扩缩容 ${successCount} 个 StatefulSet`, duration: 2200 })
  else Message.warning({ content: `成功 ${successCount} 个，失败 ${failCount} 个`, duration: 2200 })
  refreshList()
  // 批量扩缩容后启动自动刷新
  autoRefresh.value = true
  setTimeout(() => { autoRefresh.value = false }, 15000)
}

// 批量删除
const openBatchDeletePreview = () => { deleteConfirmText.value = ''; showBatchDeleteModal.value = true }
const closeBatchDeleteModal = () => { showBatchDeleteModal.value = false; deleteConfirmText.value = '' }
const executeBatchDelete = async () => {
  if (deleteConfirmText.value !== 'DELETE') return
  batchExecuting.value = true
  let successCount = 0, failCount = 0
  for (const sts of selectedStatefulsets.value) {
    try { await statefulsetsApi.delete({ name: sts.name, namespace: sts.namespace }); successCount++ }
    catch (e) { console.error(`Failed to delete ${sts.name}:`, e); failCount++ }
  }
  batchExecuting.value = false
  showBatchDeleteModal.value = false
  deleteConfirmText.value = ''
  if (failCount === 0) Message.success({ content: `成功删除 ${successCount} 个 StatefulSet`, duration: 2200 })
  else Message.warning({ content: `成功 ${successCount} 个，失败 ${failCount} 个`, duration: 2200 })
  exitBatchMode(); refreshList()
}

// 详情
const viewStatefulset = async (sts) => {
  showMoreOptions.value = false
  loadingDetail.value = true
  showViewModal.value = true
  try {
    const res = await statefulsetsApi.detail({ namespace: sts.namespace, name: sts.name })
    detailData.value = res.code === 0 ? res.data : sts
  } catch (e) { detailData.value = sts }
  finally { loadingDetail.value = false }
}

// 事件
const openEvents = async (sts) => {
  showMoreOptions.value = false
  loadingEvents.value = true
  eventsData.value = []
  showEventsModal.value = true
  try {
    const res = await statefulsetsApi.events({ namespace: sts.namespace, name: sts.name })
    eventsData.value = res.code === 0 ? (res.data?.events || res.data?.items || res.data || []) : []
  } catch (e) { console.error('获取事件失败:', e) }
  finally { loadingEvents.value = false }
}

// 历史版本
const viewHistory = async (sts) => {
  showMoreOptions.value = false
  historyStatefulset.value = sts
  loadingHistory.value = true
  historyList.value = []
  showHistoryModal.value = true
  try {
    const res = await statefulsetsApi.history({ namespace: sts.namespace, name: sts.name })
    historyList.value = res.code === 0 ? (res.data || []) : []
  } catch (e) { console.error('获取历史版本失败:', e) }
  finally { loadingHistory.value = false }
}

// Pods
const viewPods = async (sts) => {
  podsStatefulset.value = sts
  loadingPods.value = true
  podsList.value = []
  showPodsModal.value = true
  try {
    const res = await statefulsetsApi.pods({ namespace: sts.namespace, name: sts.name })
    const list = res.code === 0 ? (res.data?.pods || res.data?.items || res.data?.list || res.data || []) : []
    podsList.value = list.map(p => {
      const metadata = p.metadata || {}
      const spec = p.spec || {}
      const status = p.status || {}
      return {
        name: metadata.name || p.name,
        namespace: metadata.namespace || sts.namespace,
        status: status.phase || p.status || 'Unknown',
        node: spec.nodeName || p.node || '-',
        restartCount: status.containerStatuses?.[0]?.restartCount || p.restartCount || 0,
        createdAt: fmtTime(metadata.creationTimestamp || p.created_at),
        containers: spec.containers?.map(c => c.name) || p.containers || [],
        raw: p
      }
    })
  } catch (e) { console.error('获取 Pods 失败:', e) }
  finally { loadingPods.value = false }
}

const deletePod = async (pod) => {
  if (!confirm(`确认删除 Pod：${pod.name}？`)) return
  try {
    await podsApi.graceDelete({ namespace: pod.namespace, name: pod.name })
    Message.success({ content: 'Pod 已删除' })
    if (podsStatefulset.value) await viewPods(podsStatefulset.value)
  } catch (e) { Message.error({ content: e?.msg || '删除失败' }) }
}

// Pod 更多菜单控制
const togglePodMoreOptions = (pod, event) => {
  if (selectedPodForAction.value === pod && showPodMoreOptions.value) {
    showPodMoreOptions.value = false
    selectedPodForAction.value = null
  } else {
    selectedPodForAction.value = pod
    showPodMoreOptions.value = true
    const button = event.target.closest('.more-btn')
    if (button) {
      const rect = button.getBoundingClientRect()
      const viewportHeight = window.innerHeight
      const menuHeight = 200
      let style = { position: 'fixed', left: rect.left + 'px', zIndex: 2000 }
      if (rect.bottom + menuHeight > viewportHeight) {
        style.bottom = (viewportHeight - rect.top + 4) + 'px'
      } else {
        style.top = (rect.bottom + 4) + 'px'
      }
      podMenuStyle.value = style
    }
  }
}

// Pod 日志
const openPodLogs = (pod) => {
  showPodMoreOptions.value = false
  selectedPodForAction.value = pod
  podContainerList.value = pod.containers || []
  podLogsForm.value.container = podContainerList.value.length === 1 ? podContainerList.value[0] : ''
  podLogsForm.value.tail = 100
  podLogsContent.value = ''
  podLogsError.value = ''
  showPodLogsModal.value = true
}

const closePodLogsModal = () => {
  showPodLogsModal.value = false
  selectedPodForAction.value = null
}

const fetchPodLogsContent = async () => {
  if (!selectedPodForAction.value) return
  loadingPodLogs.value = true
  podLogsError.value = ''
  try {
    const container = podLogsForm.value.container || (podContainerList.value.length === 1 ? podContainerList.value[0] : '')
    const res = await podsApi.logs({
      namespace: selectedPodForAction.value.namespace,
      name: selectedPodForAction.value.name,
      container: container,
      tail: podLogsForm.value.tail
    })
    if (res.code === 0) {
      podLogsContent.value = res.data?.logs || res.data || '无日志内容'
    } else {
      podLogsError.value = res.msg || '获取日志失败'
    }
  } catch (e) {
    podLogsError.value = e?.msg || '获取日志失败'
  }
  finally { loadingPodLogs.value = false }
}

// Pod 详情
const openPodDetail = (pod) => {
  showPodMoreOptions.value = false
  selectedPodForAction.value = pod
  showPodDetailModal.value = true
}

// Pod 事件
const openPodEvents = async (pod) => {
  showPodMoreOptions.value = false
  selectedPodForAction.value = pod
  loadingPodEvents.value = true
  podEventsList.value = []
  showPodEventsModal.value = true
  try {
    const res = await podsApi.events({
      namespace: pod.namespace,
      name: pod.name,
      kind: 'Pod'
    })
    podEventsList.value = res.code === 0 ? (res.data?.events || res.data || []) : []
  } catch (e) {
    console.error('获取 Pod 事件失败:', e)
  }
  finally { loadingPodEvents.value = false }
}

// 重启 Pod
const restartPodFromList = async (pod) => {
  showPodMoreOptions.value = false
  if (!confirm(`确认重启 Pod：${pod.name}？\n\n注意：StatefulSet 的 Pod 将被删除并由控制器重建。`)) return
  try {
    await podsApi.graceDelete({ namespace: pod.namespace, name: pod.name })
    Message.success({ content: 'Pod 重启中（已删除，等待重建）' })
    if (podsStatefulset.value) {
      setTimeout(() => viewPods(podsStatefulset.value), 2000)
    }
  } catch (e) { Message.error({ content: e?.msg || '重启失败' }) }
}

// 优雅删除 Pod
const deletePodFromList = async (pod) => {
  showPodMoreOptions.value = false
  if (!confirm(`确认优雅删除 Pod：${pod.name}？`)) return
  try {
    await podsApi.graceDelete({ namespace: pod.namespace, name: pod.name })
    Message.success({ content: 'Pod 已删除' })
    if (podsStatefulset.value) await viewPods(podsStatefulset.value)
  } catch (e) { Message.error({ content: e?.msg || '删除失败' }) }
}

// 强制删除 Pod
const forceDeletePodFromList = async (pod) => {
  showPodMoreOptions.value = false
  if (!confirm(`确认强制删除 Pod：${pod.name}？\n\n警告：强制删除可能导致资源未完全清理！`)) return
  try {
    await podsApi.forceDelete({ namespace: pod.namespace, name: pod.name })
    Message.success({ content: 'Pod 已强制删除' })
    if (podsStatefulset.value) await viewPods(podsStatefulset.value)
  } catch (e) { Message.error({ content: e?.msg || '强制删除失败' }) }
}

// 内联编辑
const startInlineReplicas = (sts) => {
  inlineEdit.value = { key: `replicas-${sts.name}`, value: sts.desiredReplicas, original: sts.desiredReplicas }
}
const saveInlineReplicas = async (sts) => {
  const newVal = parseInt(inlineEdit.value.value)
  if (isNaN(newVal) || newVal < 0 || newVal === inlineEdit.value.original) { cancelInlineEdit(); return }
  scalingMap.value[sts.name] = true
  cancelInlineEdit()
  try {
    await statefulsetsApi.scale({ namespace: sts.namespace, name: sts.name, scale_num: newVal })
    Message.success({ content: `副本数已调整为 ${newVal}` })
    refreshList()
    // 扩缩容后启动自动刷新，追踪 Pod 状态
    autoRefresh.value = true
    setTimeout(() => { autoRefresh.value = false }, 15000)
  } catch (e) { Message.error({ content: e?.msg || '扩缩容失败' }) }
  finally { scalingMap.value[sts.name] = false }
}

const startInlineImage = (sts) => {
  inlineEdit.value = { key: `image-${sts.name}`, value: sts.image, original: sts.image }
  containerList.value = sts.containers || []
}
const saveInlineImage = async (sts) => {
  const newVal = inlineEdit.value.value?.trim()
  if (!newVal || newVal === inlineEdit.value.original) { cancelInlineEdit(); return }
  cancelInlineEdit()
  try {
    await statefulsetsApi.updateImage({ namespace: sts.namespace, name: sts.name, container: containerList.value[0] || '', image: newVal })
    Message.success({ content: '镜像更新成功' })
    refreshList()
  } catch (e) { Message.error({ content: e?.msg || '更新镜像失败' }) }
}
const cancelInlineEdit = () => { inlineEdit.value = { key: '', value: null, original: null } }

// 副本数控制
const increaseReplicas = async (sts) => {
  scalingMap.value[sts.name] = true
  try {
    await statefulsetsApi.scale({ namespace: sts.namespace, name: sts.name, scale_num: sts.desiredReplicas + 1 })
    Message.success({ content: `副本数已增加到 ${sts.desiredReplicas + 1}` })
    refreshList()
    // 扩容后启动自动刷新
    autoRefresh.value = true
    setTimeout(() => { autoRefresh.value = false }, 15000)
  } catch (e) { Message.error({ content: e?.msg || '扩容失败' }) }
  finally { scalingMap.value[sts.name] = false }
}

const decreaseReplicas = async (sts) => {
  if (sts.desiredReplicas <= 0) return
  scalingMap.value[sts.name] = true
  try {
    await statefulsetsApi.scale({ namespace: sts.namespace, name: sts.name, scale_num: sts.desiredReplicas - 1 })
    Message.success({ content: `副本数已减少到 ${sts.desiredReplicas - 1}` })
    refreshList()
    // 缩容后启动自动刷新
    autoRefresh.value = true
    setTimeout(() => { autoRefresh.value = false }, 15000)
  } catch (e) { Message.error({ content: e?.msg || '缩容失败' }) }
  finally { scalingMap.value[sts.name] = false }
}

const stopService = async (sts) => {
  if (!confirm(`确认停服（副本数调为0）：${sts.name}？`)) return
  scalingMap.value[sts.name] = true
  try {
    await statefulsetsApi.scale({ namespace: sts.namespace, name: sts.name, scale_num: 0 })
    Message.success({ content: '已停服' })
    refreshList()
    // 停服后启动自动刷新
    autoRefresh.value = true
    setTimeout(() => { autoRefresh.value = false }, 15000)
  } catch (e) { Message.error({ content: e?.msg || '停服失败' }) }
  finally { scalingMap.value[sts.name] = false }
}

// 重启
const restartStatefulset = async (sts) => {
  showMoreOptions.value = false
  if (!confirm(`确认重启 StatefulSet：${sts.name}？`)) return
  try {
    await statefulsetsApi.restart({ namespace: sts.namespace, name: sts.name })
    Message.success({ content: '重启中...' })
    autoRefresh.value = true
    setTimeout(() => { autoRefresh.value = false }, 15000)
  } catch (e) { Message.error({ content: e?.msg || '重启失败' }) }
}

// ========== StatefulSet 日志功能 ==========
const viewStatefulsetLogs = async (sts) => {
  showMoreOptions.value = false
  selectedStatefulsetForLogs.value = sts
  logsContent.value = ''
  logsError.value = ''
  logsPodsList.value = []
  logsContainerList.value = []
  logsForm.value = {
    selectedPod: '',
    container: '',
    tail: 100
  }
  showLogsModal.value = true

  // 加载 Pod 列表
  try {
    const res = await statefulsetsApi.pods({
      namespace: sts.namespace,
      name: sts.name
    })
    const pods = res.data?.pods || res.data?.list || res.data || []
    logsPodsList.value = (Array.isArray(pods) ? pods : []).map(pod => ({
      name: pod.name || pod.metadata?.name,
      containers: pod.containers || pod.spec?.containers?.map(c => c.name) || []
    }))
  } catch (e) {
    logsError.value = '获取 Pod 列表失败'
  }
}

const onLogsPodChange = () => {
  const pod = logsPodsList.value.find(p => p.name === logsForm.value.selectedPod)
  if (pod) {
    logsContainerList.value = pod.containers || []
    if (logsContainerList.value.length === 1) {
      logsForm.value.container = logsContainerList.value[0]
    } else if (logsContainerList.value.length > 1) {
      logsForm.value.container = ''
    } else {
      logsForm.value.container = ''
    }
  } else {
    logsContainerList.value = []
    logsForm.value.container = ''
  }
  logsContent.value = ''
}

const fetchLogs = async () => {
  if (!selectedStatefulsetForLogs.value) return

  // 验证容器
  if (logsForm.value.selectedPod) {
    let container = logsForm.value.container || ''
    if (logsContainerList.value.length === 1) {
      container = logsContainerList.value[0]
      logsForm.value.container = container
    }
    if (logsContainerList.value.length > 1 && !container) {
      logsError.value = '请选择容器'
      return
    }
    if (!container) {
      logsError.value = '无法确定容器名称'
      return
    }
  }

  stopLogLoading()
  loadingLogs.value = true
  logsError.value = ''
  logsContent.value = ''

  try {
    if (logsForm.value.selectedPod) {
      // 获取单个 Pod 的日志
      const params = new URLSearchParams({
        namespace: selectedStatefulsetForLogs.value.namespace,
        name: logsForm.value.selectedPod,
        container: logsForm.value.container
      })
      if (logsForm.value.tail != null) {
        params.set('tail', logsForm.value.tail)
      }

      logAbortController = new AbortController()
      const response = await fetch(`/api/v1/k8s/pod/container_logs?${params}`, {
        signal: logAbortController.signal,
        headers: getAuthHeaders()
      })

      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      
      const res = await response.json()
      logsContent.value = res?.data?.log || '暂无日志'
    } else {
      // 获取所有 Pod 的日志（汇总）
      const pods = logsPodsList.value.slice(0, 5)
      const logsArray = []

      for (const pod of pods) {
        const container = pod.containers[0] || ''
        const params = new URLSearchParams({
          namespace: selectedStatefulsetForLogs.value.namespace,
          name: pod.name,
          container
        })
        if (logsForm.value.tail != null) {
          params.set('tail', logsForm.value.tail)
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

      logsContent.value = logsArray.join('\n')
    }
  } catch (e) {
    if (e.name !== 'AbortError') {
      logsError.value = e?.msg || e?.message || '获取日志失败'
    }
  } finally {
    loadingLogs.value = false
  }
}

const stopLogLoading = () => {
  if (logAbortController) {
    logAbortController.abort()
    logAbortController = null
  }
  loadingLogs.value = false
}

const clearLogs = () => {
  logsContent.value = ''
  logsError.value = ''
}

const closeLogsModal = () => {
  stopLogLoading()
  showLogsModal.value = false
  selectedStatefulsetForLogs.value = null
}

// ========== YAML 查看/编辑功能 ==========
const openYamlPreview = async (sts) => {
  showMoreOptions.value = false
  showCardMoreOptions.value = false
  selectedYamlStatefulset.value = sts
  yamlContent.value = ''
  yamlError.value = ''
  yamlEditMode.value = false
  loadingYaml.value = true
  showYamlModal.value = true

  try {
    const res = await statefulsetsApi.yaml({ namespace: sts.namespace, name: sts.name })
    if (res.code === 0) {
      yamlContent.value = res.data?.yaml || res.data || '# YAML 内容为空'
    } else {
      yamlError.value = res.msg || '获取 YAML 失败'
    }
  } catch (e) {
    console.error('获取 YAML 失败:', e)
    yamlError.value = e?.msg || e?.message || '获取 YAML 失败'
  } finally {
    loadingYaml.value = false
  }
}

const closeYamlModal = () => {
  showYamlModal.value = false
  selectedYamlStatefulset.value = null
  yamlContent.value = ''
  yamlError.value = ''
  yamlEditMode.value = false
}

const applyYamlChanges = async () => {
  if (!yamlContent.value?.trim()) {
    Message.error({ content: '请输入 YAML 内容' })
    return
  }
  if (!selectedYamlStatefulset.value) return

  savingYaml.value = true
  try {
    const res = await statefulsetsApi.applyYaml({
      namespace: selectedYamlStatefulset.value.namespace,
      name: selectedYamlStatefulset.value.name,
      yaml: yamlContent.value
    })
    if (res.code === 0) {
      Message.success({ content: 'YAML 应用成功' })
      closeYamlModal()
      refreshList()
      // 应用后启动自动刷新
      autoRefresh.value = true
      setTimeout(() => { autoRefresh.value = false }, 15000)
    } else {
      Message.error({ content: res.msg || '应用 YAML 失败' })
    }
  } catch (e) {
    console.error('应用 YAML 失败:', e)
    Message.error({ content: e?.msg || e?.message || '应用 YAML 失败' })
  } finally {
    savingYaml.value = false
  }
}

// 下载 YAML
const downloadYaml = () => {
  if (!yamlContent.value || !selectedYamlStatefulset.value) {
    Message.warning({ content: '没有可下载的 YAML 内容' })
    return
  }
  
  try {
    const blob = new Blob([yamlContent.value], { type: 'text/yaml;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `${selectedYamlStatefulset.value.name}-statefulset.yaml`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
    Message.success({ content: 'YAML 文件已下载' })
  } catch (e) {
    Message.error({ content: '下载失败' })
  }
}

// 更新镜像
const openUpdateImage = (sts) => {
  showMoreOptions.value = false
  containerList.value = sts.containers || []
  updateImageForm.value = {
    namespace: sts.namespace,
    name: sts.name,
    container: containerList.value[0] || '',
    image: ''
  }
  showUpdateImageModal.value = true
}

const submitUpdateImage = async () => {
  if (!updateImageForm.value.image?.trim()) {
    Message.error({ content: '请输入镜像地址' })
    return
  }
  updatingImage.value = true
  try {
    await statefulsetsApi.updateImage({
      namespace: updateImageForm.value.namespace,
      name: updateImageForm.value.name,
      container: updateImageForm.value.container,
      image: updateImageForm.value.image.trim()
    })
    Message.success({ content: '镜像更新成功' })
    showUpdateImageModal.value = false
    refreshList()
  } catch (e) { Message.error({ content: e?.msg || '更新镜像失败' }) }
  finally { updatingImage.value = false }
}

// 删除
const deleteStatefulset = (sts) => {
  showMoreOptions.value = false
  stsToDelete.value = sts
  showDeleteModal.value = true
}

const confirmDelete = async () => {
  if (!stsToDelete.value) return
  deleting.value = true
  try {
    await statefulsetsApi.delete({ namespace: stsToDelete.value.namespace, name: stsToDelete.value.name })
    Message.success({ content: '删除成功' })
    showDeleteModal.value = false
    refreshList()
    // 删除后启动自动刷新，观察关联资源清理
    autoRefresh.value = true
    setTimeout(() => { autoRefresh.value = false }, 15000)
  } catch (e) { Message.error({ content: e?.msg || '删除失败' }) }
  finally { deleting.value = false }
}

// 创建
const addLabel = () => { statefulsetForm.value.labels.push({ key: '', value: '' }) }
const removeLabel = (index) => { if (statefulsetForm.value.labels.length > 1) statefulsetForm.value.labels.splice(index, 1) }

// VolumeClaimTemplate 操作
const addVolumeClaimTemplate = () => {
  statefulsetForm.value.volumeClaimTemplates.push({
    name: '',
    storageType: 'default',  // 存储类型
    storageClass: storageClasses.value[0] || '',
    accessMode: 'ReadWriteOnce',
    storageSize: '1',
    storageSizeUnit: 'Gi',
    mountPath: ''
  })
}

// 根据存储类型获取推荐的访问模式
const getRecommendedAccessModes = (storageType) => {
  const typeConfig = storageTypeOptions.find(t => t.value === storageType)
  if (!typeConfig) return accessModeOptions
  return accessModeOptions.filter(am => typeConfig.accessModes.includes(am.value))
}

// 存储类型变更时自动调整访问模式
const onStorageTypeChange = (vct) => {
  const typeConfig = storageTypeOptions.find(t => t.value === vct.storageType)
  if (typeConfig && typeConfig.accessModes.length > 0) {
    // 如果当前访问模式不在推荐列表中，自动切换到第一个推荐模式
    if (!typeConfig.accessModes.includes(vct.accessMode)) {
      vct.accessMode = typeConfig.accessModes[0]
    }
  }
}

const removeVolumeClaimTemplate = (index) => {
  statefulsetForm.value.volumeClaimTemplates.splice(index, 1)
}

const createStatefulset = async () => {
  const form = statefulsetForm.value
  if (!form.name || !form.namespace || !form.image) {
    Message.error({ content: '请填写必填字段' })
    return
  }
  
  // 验证标签（与 Deployment 保持一致，发送数组格式）
  const validLabels = form.labels.filter(label => label.key && label.value)
  if (validLabels.length === 0) {
    Message.error({ content: '请至少添加一个标签' })
    return
  }
  
  // 构建 VolumeClaimTemplates
  const volumeClaimTemplates = form.volumeClaimTemplates
    .filter(vct => vct.name && vct.mountPath && vct.storageSize)
    .map(vct => ({
      name: vct.name,
      storage_class: vct.storageClass || '',
      access_mode: vct.accessMode || 'ReadWriteOnce',
      storage_size: `${vct.storageSize}${vct.storageSizeUnit || 'Gi'}`,
      mount_path: vct.mountPath
    }))
  
  try {
    const res = await statefulsetsApi.create({
      namespace: form.namespace,
      name: form.name,
      container_image: form.image,
      replicas: parseInt(form.replicas) || 1,
      labels: validLabels,  // 发送数组格式，与 Deployment 一致
      is_create_service: form.createService,
      service_name: form.serviceName || form.name,
      volume_claim_templates: volumeClaimTemplates
    })
    if (res.code === 0) {
      Message.success({ content: '创建成功' })
      showCreateModal.value = false
      resetForm()
      refreshList()
      // 创建后启动自动刷新 15 秒，观察 Pod 启动状态
      autoRefresh.value = true
      setTimeout(() => { autoRefresh.value = false }, 15000)
    } else {
      const errorMsg = res.details ? `${res.msg}: ${res.details}` : (res.msg || '创建失败')
      Message.error({ content: errorMsg, duration: 5000 })
    }
  } catch (e) {
    const errorMsg = e?.details ? `${e?.msg || '创建失败'}: ${e.details}` : (e?.msg || '创建失败')
    Message.error({ content: errorMsg, duration: 5000 })
  }
}

const resetForm = () => {
  statefulsetForm.value = {
    name: '', namespace: namespaces.value[0] || 'default', replicas: 1, image: '',
    updateStrategy: 'RollingUpdate',
    labels: [{ key: 'app', value: '' }],
    createService: true, serviceName: '',
    volumeClaimTemplates: []
  }
  showNamespaceInput.value = false
  newNamespace.value = ''
  // 重置 YAML 创建状态
  createMode.value = 'form'
  yamlContent.value = ''
  yamlError.value = ''
}

// YAML 创建相关函数
const loadStatefulSetYamlTemplate = () => {
  yamlContent.value = `apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: example-statefulset
  namespace: default
  labels:
    app: example
spec:
  serviceName: "example-service"
  replicas: 3
  selector:
    matchLabels:
      app: example
  template:
    metadata:
      labels:
        app: example
    spec:
      containers:
      - name: example-container
        image: nginx:latest
        ports:
        - containerPort: 80
          name: web
        env:
        - name: EXAMPLE_ENV
          value: "example-value"
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "128Mi"
            cpu: "200m"
        volumeMounts:
        - name: data
          mountPath: /usr/share/nginx/html
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: [ "ReadWriteOnce" ]
      resources:
        requests:
          storage: 1Gi`
  yamlError.value = ''
  Message.success({ content: '已加载 YAML 模板，请修改后创建' })
}

// 清除 YAML 内容
const clearYamlContent = () => {
  yamlContent.value = ''
  yamlError.value = ''
  Message.success({ content: 'YAML 内容已清除' })
}

const createStatefulSetFromYaml = async () => {
  if (!yamlContent.value.trim()) {
    Message.error({ content: '请输入 YAML 内容' })
    return
  }
  
  // 简单验证
  if (!yamlContent.value.includes('kind: StatefulSet')) {
    yamlError.value = 'YAML 中必须包含 "kind: StatefulSet"'
    return
  }
  if (!yamlContent.value.includes('apiVersion: apps/v1')) {
    yamlError.value = 'YAML 中必须包含 "apiVersion: apps/v1"'
    return
  }
  
  yamlError.value = ''
  
  try {
    const res = await statefulsetsApi.createFromYaml({ yaml: yamlContent.value })
    if (res.code === 0) {
      Message.success({ content: 'StatefulSet 创建成功' })
      showCreateModal.value = false
      resetForm()
      await fetchStatefulsets()
    } else {
      const errorMsg = res.details ? `${res.msg}: ${res.details}` : (res.msg || '创建失败')
      Message.error({ content: errorMsg, duration: 5000 })
      yamlError.value = errorMsg
    }
  } catch (e) {
    const errorMsg = e?.details ? `${e?.msg || '创建失败'}: ${e.details}` : (e?.msg || e?.message || '创建失败')
    Message.error({ content: errorMsg, duration: 5000 })
    yamlError.value = errorMsg
  }
}
</script>

<style scoped>
@import '@/assets/styles/resizable-modal.css';

.resource-view { max-width: 1400px; margin: 0 auto; }
.view-header { margin-bottom: 24px; }
.view-header h1 { font-size: 32px; font-weight: 700; color: #2d3748; margin-bottom: 8px; }
.view-header p { font-size: 16px; color: #718096; }

.action-bar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; flex-wrap: wrap; gap: 16px; }
.search-box input { padding: 10px 16px; border: 1px solid #e2e8f0; border-radius: 8px; font-size: 14px; width: 300px; }
.filter-dropdown select { padding: 10px 16px; border: 1px solid #e2e8f0; border-radius: 8px; font-size: 14px; background-color: white; }
.action-buttons { display: flex; gap: 12px; align-items: center; }

.filter-buttons { display: flex; gap: 8px; }
.btn-filter { padding: 8px 16px; border: 1px solid #e2e8f0; border-radius: 6px; background: white; cursor: pointer; transition: all 0.2s; }
.btn-filter.active { background: #326ce5; color: white; border-color: #326ce5; }
.btn-filter:hover:not(.active) { background: #f7fafc; }

.btn { padding: 10px 20px; border: none; border-radius: 8px; font-size: 14px; font-weight: 500; cursor: pointer; transition: all 0.3s ease; }
.btn-primary { background-color: #326ce5; color: white; }
.btn-primary:hover { background-color: #2554c7; transform: translateY(-1px); box-shadow: 0 4px 12px rgba(50, 108, 229, 0.3); }
.btn-secondary { background-color: #e2e8f0; color: #4a5568; }
.btn-secondary:hover { background-color: #cbd5e0; }
.btn-danger { background-color: #ef4444; color: white; }
.btn-danger:hover { background-color: #dc2626; }
.btn-warning { background-color: #f59e0b; color: white; }
.btn-batch { background-color: #8b5cf6; color: white; }
.btn-sm { padding: 6px 12px; font-size: 12px; }

.view-toggle { display: flex; gap: 4px; }
.btn-view { padding: 8px 12px; background: #e2e8f0; border: none; border-radius: 6px; cursor: pointer; }
.btn-view.active { background: #326ce5; color: white; }

.auto-refresh-toggle { display: flex; align-items: center; gap: 6px; font-size: 14px; cursor: pointer; }
.refresh-indicator { color: #34d399; animation: pulse 1s infinite; }
@keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.5; } }

.error-box { background: #fef2f2; border: 1px solid #fecaca; color: #dc2626; padding: 12px 16px; border-radius: 8px; margin-bottom: 16px; }

.batch-action-bar { display: flex; justify-content: space-between; align-items: center; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding: 12px 20px; border-radius: 8px; margin-bottom: 16px; }
.batch-info { display: flex; align-items: center; gap: 16px; color: white; }
.batch-count { font-weight: 600; }
.batch-clear { background: rgba(255,255,255,0.2); border: none; color: white; padding: 4px 12px; border-radius: 4px; cursor: pointer; }
.batch-actions { display: flex; gap: 8px; }
.batch-btn { background: rgba(255,255,255,0.2); border: none; color: white; padding: 8px 16px; border-radius: 6px; cursor: pointer; transition: all 0.2s; }
.batch-btn:hover { background: rgba(255,255,255,0.3); }
.batch-btn.danger { background: rgba(239, 68, 68, 0.8); }

.table-container { background: white; border-radius: 12px; box-shadow: 0 4px 6px rgba(0, 0, 0, 0.05); overflow-x: auto; /* 支持横向滚动 */ overflow-y: visible; }
.resource-table { width: 100%; border-collapse: collapse; min-width: 1600px; /* 确保表格有足够宽度避免挤压 */ table-layout: auto; /* 根据内容自动调整列宽 */ }
.resource-table th { background-color: #f7fafc; text-align: left; padding: 16px 20px; font-size: 14px; font-weight: 600; color: #4a5568; border-bottom: 1px solid #e2e8f0; }
.resource-table td { padding: 16px 20px; font-size: 14px; color: #2d3748; border-bottom: 1px solid #f7fafc; vertical-align: middle; }
.resource-table tbody tr:hover { background-color: #f7fafc; }
.row-selected { background-color: #ebf5ff !important; }

.status-indicator { display: inline-flex; align-items: center; padding: 6px 12px; border-radius: 20px; font-size: 12px; font-weight: 600; }
.status-indicator.running { background-color: rgba(52, 211, 153, 0.1); color: #34d399; }
.status-indicator.failed { background-color: rgba(239, 68, 68, 0.1); color: #ef4444; }
.status-indicator.updating { background-color: rgba(245, 158, 11, 0.1); color: #f59e0b; }
.status-indicator.stopped { background-color: rgba(107, 114, 128, 0.1); color: #6b7280; }
.status-indicator.pending { background-color: rgba(59, 130, 246, 0.1); color: #3b82f6; }
.status-indicator.unknown { background-color: rgba(107, 114, 128, 0.1); color: #6b7280; }

.statefulset-name { display: flex; align-items: center; gap: 8px; }
.namespace-badge { background-color: #e2e8f0; color: #4a5568; padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: 500; }

.replicas-info { display: flex; flex-direction: column; gap: 4px; }
.replicas-control { display: flex; align-items: center; gap: 4px; }
.replica-btn { width: 24px; height: 24px; border: 1px solid #e2e8f0; background: white; border-radius: 4px; cursor: pointer; font-size: 14px; display: flex; align-items: center; justify-content: center; }
.replica-btn:hover:not(:disabled) { background: #f7fafc; border-color: #326ce5; }
.replica-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.replica-btn.stop { font-size: 12px; }
.replicas-display { display: flex; align-items: center; padding: 4px 8px; background: #f7fafc; border-radius: 4px; min-width: 50px; justify-content: center; }
.replicas-display.clickable { cursor: pointer; }
.replicas-display.clickable:hover { background: #e2e8f0; }
.replicas-display.updating { background: #fef3c7; }
.ready-replicas { color: #34d399; font-weight: 600; }
.replicas-sep { color: #718096; margin: 0 2px; }
.desired-replicas { color: #718096; }
.scaling-indicator { margin-left: 4px; }
.edit-icon-small { font-size: 10px; margin-left: 4px; opacity: 0; transition: opacity 0.2s; }
.replicas-display:hover .edit-icon-small { opacity: 1; }
.replicas-bar { width: 60px; height: 4px; background-color: #e2e8f0; border-radius: 2px; overflow: hidden; }
.replicas-fill { height: 100%; background-color: #34d399; border-radius: 2px; transition: width 0.3s; }
.replicas-edit-wrapper { display: flex; align-items: center; gap: 4px; }
.replicas-input { width: 60px; padding: 4px 8px; border: 1px solid #326ce5; border-radius: 4px; font-size: 14px; }
.inline-hint-small { font-size: 10px; color: #718096; }

.inline-edit-wrapper { display: flex; flex-direction: column; gap: 4px; }
.inline-input { width: 100%; padding: 6px 10px; border: 1px solid #326ce5; border-radius: 4px; font-size: 14px; }
.inline-hint { font-size: 10px; color: #718096; }
.image-text { display: flex; align-items: center; gap: 8px; max-width: 320px; position: relative; }
.image-text.clickable { cursor: pointer; }
.image-text.clickable:hover { color: #326ce5; }
.image-name { max-width: 280px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.edit-icon { font-size: 12px; opacity: 0; transition: opacity 0.2s; }
.image-text:hover .edit-icon { opacity: 1; }

.selector-tags { display: flex; flex-wrap: wrap; gap: 4px; max-width: 220px; }
.selector-tag { background-color: rgba(50, 108, 229, 0.1); color: #326ce5; padding: 2px 8px; border-radius: 4px; font-size: 11px; font-weight: 500; max-width: 180px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.selector-tag:hover { background-color: rgba(50, 108, 229, 0.15); }
.strategy-badge { background-color: rgba(59, 130, 246, 0.1); color: #3b82f6; padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: 500; }

.action-icons { display: flex; gap: 8px; align-items: center; flex-wrap: nowrap; }
.action-btn { padding: 6px 12px; border: none; border-radius: 6px; font-size: 12px; cursor: pointer; background: #e2e8f0; color: #4a5568; transition: all 0.2s; white-space: nowrap; }
.action-btn.primary { background: #326ce5; color: white; }
.action-btn:hover { transform: translateY(-1px); }
.icon-btn { background: none; border: none; font-size: 14px; cursor: pointer; padding: 4px 8px; border-radius: 4px; white-space: nowrap; }
.icon-btn:hover { background-color: #e2e8f0; }

.more-btn { position: relative; }
.more-menu { position: fixed; background: white; border-radius: 8px; box-shadow: 0 10px 40px rgba(0,0,0,0.15); min-width: 160px; z-index: 1000; padding: 8px 0; }
.menu-item { display: flex; align-items: center; gap: 10px; width: 100%; padding: 10px 16px; border: none; background: none; cursor: pointer; font-size: 14px; color: #2d3748; }
.menu-item:hover { background: #f7fafc; }
.menu-item.danger { color: #ef4444; }
.menu-item.danger:hover { background: #fef2f2; }
.menu-icon { font-size: 16px; }
.menu-divider { height: 1px; background: #e2e8f0; margin: 4px 0; }

.empty-state { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 64px 24px; color: #718096; }
.empty-icon { font-size: 48px; margin-bottom: 16px; }
.empty-text { font-size: 16px; }

/* 卡片视图 */
.cards-container { }
.cards-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(380px, 1fr)); gap: 20px; margin-bottom: 20px; }
.statefulset-card { background: white; border-radius: 12px; box-shadow: 0 4px 6px rgba(0, 0, 0, 0.05); overflow: hidden; position: relative; transition: all 0.2s; }
.statefulset-card:hover { box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1); }
.card-selected { border: 2px solid #326ce5; }
.card-checkbox { position: absolute; top: 12px; right: 12px; z-index: 10; }
.card-header { padding: 16px; border-bottom: 1px solid #f7fafc; }
.card-title-row { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; }
.card-icon { font-size: 24px; }
.card-title { font-size: 16px; font-weight: 600; color: #2d3748; margin: 0; flex: 1; }
.card-body { padding: 16px; }
.card-section { margin-bottom: 16px; }
.card-section:last-child { margin-bottom: 0; }
.section-label { font-size: 12px; color: #718096; margin-bottom: 6px; }
.card-section-row { display: flex; gap: 20px; }
.card-meta-item { flex: 1; }
.meta-label { font-size: 12px; color: #718096; margin-bottom: 4px; }
.meta-value { font-size: 14px; color: #2d3748; }
.card-footer { display: flex; gap: 8px; padding: 12px 16px; border-top: 1px solid #f7fafc; background: #f7fafc; flex-wrap: wrap; }
.card-action-btn { padding: 6px 12px; border: none; border-radius: 6px; font-size: 12px; cursor: pointer; background: white; color: #4a5568; transition: all 0.2s; }
.card-action-btn.primary { background: #326ce5; color: white; }
.card-action-btn.danger { color: #ef4444; }
.card-action-btn:hover { transform: translateY(-1px); }

/* 资源使用样式 */
.metrics-summary { display: flex; gap: 12px; }
.metric-item { flex: 1; display: flex; align-items: center; gap: 4px; padding: 8px 10px; background: linear-gradient(135deg, #f0f9ff 0%, #e0f2fe 100%); border: 1px solid #bae6fd; border-radius: 6px; font-size: 12px; }
.metric-icon { font-size: 14px; }
.metric-label { font-weight: 600; color: #0369a1; }
.metric-value { font-weight: 700; color: #0c4a6e; font-family: 'Monaco', 'Menlo', monospace; }
.metrics-unavailable { flex-wrap: wrap; }
.metrics-unavailable .metric-item { background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%); border-color: #e2e8f0; }
.metrics-unavailable .metric-value.muted { color: #94a3b8; font-weight: 500; }
.metrics-hint { width: 100%; text-align: center; font-size: 11px; color: #94a3b8; margin-top: 4px; font-style: italic; }

/* Modal */
.modal-overlay { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background-color: rgba(0, 0, 0, 0.5); display: flex; align-items: center; justify-content: center; z-index: 1000; }
.modal-content { background-color: white; border-radius: 12px; box-shadow: 0 10px 25px rgba(0, 0, 0, 0.1); width: 90%; max-width: 600px; max-height: 90vh; overflow-y: auto; }
.modal-large { max-width: 900px; }
.modal-create-statefulset { max-width: 700px; }
.modal-batch-preview { max-width: 600px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 20px; border-bottom: 1px solid #e2e8f0; }
.modal-header h2, .modal-header h3 { font-size: 20px; font-weight: 600; color: #2d3748; margin: 0; }
.warning-header { background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%); color: white; }
.warning-header h3 { color: white; }
.danger-header { background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%); color: white; }
.danger-header h3 { color: white; }
.close-btn { background: none; border: none; font-size: 24px; cursor: pointer; color: #718096; padding: 0; width: 32px; height: 32px; display: flex; align-items: center; justify-content: center; border-radius: 4px; }
.close-btn:hover { background-color: #e2e8f0; }
.modal-body { padding: 20px; }
.modal-footer { display: flex; justify-content: flex-end; gap: 12px; padding: 20px; border-top: 1px solid #e2e8f0; }

.loading-state { text-align: center; padding: 40px; color: #718096; }
.detail-group { display: flex; margin-bottom: 12px; }
.detail-group label { font-weight: 600; width: 120px; color: #4a5568; }
.detail-group span { color: #2d3748; }

/* Form */
.form-section { background: #f7fafc; border-radius: 8px; margin-bottom: 16px; overflow: hidden; }
.section-header { display: flex; align-items: center; gap: 10px; padding: 12px 16px; background: #e2e8f0; }
.section-header h3 { margin: 0; font-size: 14px; font-weight: 600; color: #2d3748; }
.section-icon { font-size: 18px; }
.section-body { padding: 16px; }
.form-group { margin-bottom: 16px; }
.form-group:last-child { margin-bottom: 0; }
.form-group label { display: block; font-size: 14px; font-weight: 500; color: #4a5568; margin-bottom: 6px; }
.form-input, .form-select { width: 100%; padding: 10px 12px; border: 1px solid #e2e8f0; border-radius: 8px; font-size: 14px; }
.form-input:focus, .form-select:focus { outline: none; border-color: #326ce5; box-shadow: 0 0 0 3px rgba(50, 108, 229, 0.1); }
.form-row { display: flex; gap: 16px; }
.form-row .form-group { flex: 1; }
.required { color: #ef4444; }
.checkbox-label { display: flex; align-items: center; gap: 8px; cursor: pointer; }
.labels-editor { }
.label-row { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; }
.label-input { padding: 8px 12px; border: 1px solid #e2e8f0; border-radius: 6px; font-size: 14px; }
.label-key { flex: 1; }
.label-value { flex: 1; }
.label-separator { color: #718096; }
.btn-icon { background: none; border: none; cursor: pointer; padding: 4px; font-size: 16px; }
.btn-remove { color: #ef4444; }

/* Tables */
.events-table, .history-table, .pods-table { width: 100%; border-collapse: collapse; }
.events-table th, .history-table th, .pods-table th { background: #f7fafc; padding: 12px; text-align: left; font-size: 13px; font-weight: 600; color: #4a5568; }
.events-table td, .history-table td, .pods-table td { padding: 12px; border-bottom: 1px solid #f7fafc; font-size: 13px; }
.event-type { padding: 4px 8px; border-radius: 4px; font-size: 11px; font-weight: 600; }
.event-type.normal { background: rgba(52, 211, 153, 0.1); color: #34d399; }
.event-type.warning { background: rgba(245, 158, 11, 0.1); color: #f59e0b; }
.revision-badge { background: #326ce5; color: white; padding: 4px 10px; border-radius: 12px; font-size: 12px; font-weight: 600; }

/* Batch Preview */
.warning-box, .danger-warning { display: flex; gap: 16px; padding: 16px; background: #fef3c7; border-radius: 8px; margin-bottom: 20px; }
.danger-warning { background: #fef2f2; }
.warning-icon-large { font-size: 32px; }
.warning-content { flex: 1; }
.warning-title { font-weight: 600; color: #92400e; margin-bottom: 8px; }
.danger-warning .warning-title { color: #991b1b; }
.warning-list { margin: 0; padding-left: 20px; color: #92400e; font-size: 14px; }
.danger-warning .warning-list { color: #991b1b; }
.preview-section { margin-bottom: 20px; }
.section-title { font-size: 14px; font-weight: 600; color: #4a5568; margin-bottom: 12px; }
.affected-deployments { max-height: 200px; overflow-y: auto; }
.affected-item { display: flex; justify-content: space-between; padding: 8px 12px; background: #f7fafc; border-radius: 6px; margin-bottom: 8px; }
.dep-name { font-weight: 500; }
.dep-replicas { color: #718096; font-size: 13px; }
.affected-deployments-detail { max-height: 300px; overflow-y: auto; }
.affected-dep-card { display: flex; justify-content: space-between; align-items: center; padding: 12px; background: #f7fafc; border-radius: 8px; margin-bottom: 8px; }
.dep-info { display: flex; flex-direction: column; gap: 4px; }
.dep-namespace { font-size: 12px; color: #718096; }
.dep-stats { display: flex; align-items: center; gap: 8px; }
.replicas-tag { font-size: 12px; color: #718096; background: #e2e8f0; padding: 2px 8px; border-radius: 4px; }
.change-preview { }
.change-item { display: flex; align-items: center; gap: 8px; padding: 8px 12px; background: #f7fafc; border-radius: 6px; margin-bottom: 8px; }
.old-value { color: #ef4444; text-decoration: line-through; }
.arrow { color: #718096; }
.new-value { color: #34d399; font-weight: 600; }
.confirm-section { }
.confirm-input { width: 100%; padding: 12px; border: 2px solid #e2e8f0; border-radius: 8px; font-size: 16px; text-align: center; }
.confirm-input.valid { border-color: #34d399; background: #f0fdf4; }

/* 分页跳转 */
.pagination-wrapper { display: flex; justify-content: center; align-items: center; gap: 20px; margin-top: 20px; flex-wrap: wrap; }
.jump-page { display: flex; align-items: center; gap: 8px; font-size: 14px; color: #4a5568; }
.jump-page input { width: 60px; padding: 6px 8px; border: 1px solid #e2e8f0; border-radius: 6px; text-align: center; font-size: 14px; }
.jump-page input:focus { outline: none; border-color: #326ce5; }

/* 命名空间创建 */
.namespace-selector { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.namespace-or { color: #718096; font-size: 12px; }
.namespace-create { display: flex; align-items: center; gap: 8px; flex: 1; }
.namespace-create .form-input { flex: 1; min-width: 120px; }

/* 存储配置样式 */
.storage-hint { margin-bottom: 16px; padding: 10px 14px; background: #e2e8f0; border-radius: 6px; color: #4a5568; font-size: 13px; }
.empty-volumes { display: flex; align-items: center; gap: 12px; padding: 16px; background: white; border: 2px dashed #e2e8f0; border-radius: 8px; }
.empty-volumes .empty-text { color: #718096; font-size: 14px; }
.volume-claim-item { background: white; border: 1px solid #e2e8f0; border-radius: 8px; margin-bottom: 12px; overflow: hidden; }
.volume-header { display: flex; justify-content: space-between; align-items: center; padding: 10px 14px; background: #f7fafc; border-bottom: 1px solid #e2e8f0; }
.volume-index { font-size: 13px; font-weight: 600; color: #4a5568; }
.volume-body { padding: 14px; }
.storage-size-group { max-width: 200px; }
.storage-size-input { display: flex; gap: 8px; }
.storage-size-input .form-input { flex: 1; }
.storage-size-input .size-unit { width: 80px; flex-shrink: 0; }
.form-hint { font-size: 12px; color: #718096; margin-top: 4px; }

/* 存储类型选择器（卡片式 - 参考 Rancher/Kuboard） */
.storage-type-selector { display: flex; flex-wrap: wrap; gap: 8px; margin-bottom: 8px; }
.storage-type-card { display: flex; flex-direction: column; align-items: center; justify-content: center; min-width: 72px; padding: 8px 10px; border: 2px solid #e2e8f0; border-radius: 8px; cursor: pointer; transition: all 0.2s; background: white; }
.storage-type-card:hover { border-color: #326ce5; background: #f7fafc; }
.storage-type-card.active { border-color: #326ce5; background: rgba(50, 108, 229, 0.1); }
.stype-icon { font-size: 20px; margin-bottom: 2px; }
.stype-name { font-size: 11px; font-weight: 500; color: #4a5568; text-align: center; white-space: nowrap; }
.storage-type-card.active .stype-name { color: #326ce5; font-weight: 600; }
.storage-type-desc { font-size: 12px; color: #718096; padding: 8px 12px; background: #f0f4f8; border-radius: 6px; margin-top: 4px; }

/* 日志弹窗 */
.logs-modal { max-width: 900px; width: 95%; }
.logs-controls { display: flex; flex-wrap: wrap; gap: 12px; align-items: flex-end; margin-bottom: 16px; padding: 12px; background: #f7fafc; border-radius: 8px; }
.control-item { display: flex; flex-direction: column; gap: 4px; }
.control-item label { font-size: 12px; color: #4a5568; font-weight: 500; }
.control-item .form-select { min-width: 120px; padding: 8px 10px; font-size: 13px; }
.control-actions { display: flex; gap: 8px; align-items: flex-end; }
.logs-content-wrapper { background: #1a1a2e; border-radius: 8px; min-height: 300px; max-height: 500px; overflow: auto; }
.logs-content { margin: 0; padding: 16px; color: #e2e8f0; font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace; font-size: 12px; line-height: 1.6; white-space: pre-wrap; word-break: break-all; }
.info-box { padding: 12px 16px; background: #f0f4f8; border-radius: 8px; font-size: 14px; color: #4a5568; margin-bottom: 16px; }
.info-box div { margin-bottom: 4px; }
.info-box div:last-child { margin-bottom: 0; }

/* Pod 关联弹窗 */
.modal-lg { max-width: 1000px; }
.simple-table { width: 100%; border-collapse: collapse; }
.simple-table th { background: #f7fafc; padding: 12px; text-align: left; font-size: 13px; font-weight: 600; color: #4a5568; border-bottom: 2px solid #e2e8f0; }
.simple-table td { padding: 12px; border-bottom: 1px solid #f7fafc; font-size: 13px; }
.pod-name-cell { max-width: 300px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.pod-actions { display: flex; gap: 8px; align-items: center; }
.pod-info-bar { display: flex; gap: 20px; padding: 12px 16px; background: #f0f4f8; border-radius: 8px; font-size: 14px; margin-bottom: 16px; }
.single-container { padding: 4px 8px; background: #e2e8f0; border-radius: 4px; font-size: 13px; }
.empty-state-small { text-align: center; padding: 40px; color: #718096; }
.empty-state-small .empty-icon { font-size: 32px; margin-bottom: 8px; }
.empty-state-small .empty-text { font-size: 14px; }
.error-box { padding: 12px; background: #fed7d7; color: #c53030; border-radius: 6px; font-size: 13px; }
.detail-content { }
.detail-content .detail-group { display: flex; margin-bottom: 12px; }
.detail-content .detail-group label { font-weight: 600; width: 100px; color: #4a5568; }
.detail-content .detail-group span { color: #2d3748; }

/* 版本历史样式 */
.version-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  background: linear-gradient(135deg, #f7941d 0%, #f26b3a 100%);
  color: white;
  border-radius: 12px;
  font-size: 13px;
  font-weight: 600;
}
.current-tag {
  display: inline-block;
  padding: 2px 6px;
  background: rgba(255, 255, 255, 0.3);
  border-radius: 8px;
  font-size: 11px;
  font-weight: 500;
}
.current-version-text {
  color: #38a169;
  font-size: 13px;
  font-weight: 500;
}
.mono {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
  color: #4a5568;
}

/* 卡片视图更多菜单 */
.card-more-btn { position: relative; }
.card-more-menu { position: fixed; background: white; border-radius: 8px; box-shadow: 0 10px 40px rgba(0,0,0,0.15); min-width: 160px; z-index: 2000; padding: 8px 0; }

/* YAML 编辑器样式 */
.yaml-modal { max-width: 900px; width: 95%; }
.yaml-modal-body { min-height: 400px; }
.yaml-header-actions { display: flex; align-items: center; gap: 12px; }
.yaml-editor-wrapper { background: #1a1a2e; border-radius: 8px; min-height: 400px; max-height: 600px; overflow: auto; }
.yaml-editor {
  width: 100%;
  min-height: 400px;
  padding: 16px;
  background: #1a1a2e;
  color: #e2e8f0;
  border: none;
  border-radius: 8px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
  line-height: 1.6;
  resize: vertical;
  outline: none;
}
.yaml-editor:focus {
  outline: none;
  box-shadow: 0 0 0 2px rgba(50, 108, 229, 0.3);
}
.yaml-content {
  margin: 0;
  padding: 16px;
  color: #e2e8f0;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
}

/* 滚动条美化样式 */
/* 表格容器滚动条 */
.table-container::-webkit-scrollbar {
  width: 10px;
  height: 10px;
}

.table-container::-webkit-scrollbar-track {
  background: #f1f5f9;
  border-radius: 5px;
}

.table-container::-webkit-scrollbar-thumb {
  background: linear-gradient(135deg, #cbd5e1 0%, #94a3b8 100%);
  border-radius: 5px;
  border: 2px solid #f1f5f9;
}

.table-container::-webkit-scrollbar-thumb:hover {
  background: linear-gradient(135deg, #94a3b8 0%, #64748b 100%);
}

/* 模态框内容滚动条 */
.modal-content::-webkit-scrollbar {
  width: 8px;
}

.modal-content::-webkit-scrollbar-track {
  background: #f8fafc;
  border-radius: 4px;
}

.modal-content::-webkit-scrollbar-thumb {
  background: linear-gradient(135deg, #cbd5e1 0%, #94a3b8 100%);
  border-radius: 4px;
}

.modal-content::-webkit-scrollbar-thumb:hover {
  background: linear-gradient(135deg, #94a3b8 0%, #64748b 100%);
}

/* 通用滚动区域滚动条 */
.modal-body::-webkit-scrollbar,
.logs-content::-webkit-scrollbar,
.yaml-preview::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

.modal-body::-webkit-scrollbar-track,
.logs-content::-webkit-scrollbar-track,
.yaml-preview::-webkit-scrollbar-track {
  background: #334155;
  border-radius: 4px;
}

.modal-body::-webkit-scrollbar-thumb,
.logs-content::-webkit-scrollbar-thumb,
.yaml-preview::-webkit-scrollbar-thumb {
  background: linear-gradient(135deg, #64748b 0%, #475569 100%);
  border-radius: 4px;
}

.modal-body::-webkit-scrollbar-thumb:hover,
.logs-content::-webkit-scrollbar-thumb:hover,
.yaml-preview::-webkit-scrollbar-thumb:hover {
  background: linear-gradient(135deg, #94a3b8 0%, #64748b 100%);
}

/* YAML 编辑器样式 */
.view-toggle-buttons {
  display: flex;
  gap: 8px;
  margin: 0 auto;
}

.view-toggle-btn {
  padding: 8px 16px;
  border: 1px solid #e2e8f0;
  background: white;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  color: #4a5568;
  transition: all 0.2s;
}

.view-toggle-btn.active {
  background: #326ce5;
  color: white;
  border-color: #326ce5;
}

.view-toggle-btn:hover:not(.active) {
  background: #f7fafc;
  border-color: #326ce5;
}

.yaml-editor-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.yaml-editor-header {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.yaml-editor-header h3 {
  font-size: 16px;
  font-weight: 600;
  color: #2d3748;
  margin: 0;
}

.yaml-hint {
  font-size: 14px;
  color: #718096;
}

.yaml-header-buttons {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.load-template-btn {
  padding: 10px 20px;
  background: linear-gradient(135deg, #326ce5 0%, #2554c7 100%);
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.3s ease;
  box-shadow: 0 2px 8px rgba(50, 108, 229, 0.3);
}

.load-template-btn:hover {
  background: linear-gradient(135deg, #2554c7 0%, #1e429f 100%);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(50, 108, 229, 0.4);
}

.clear-yaml-btn {
  padding: 10px 20px;
  background: linear-gradient(135deg, #94a3b8 0%, #64748b 100%);
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.3s ease;
  box-shadow: 0 2px 8px rgba(100, 116, 139, 0.3);
}

.clear-yaml-btn:hover {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.4);
}

.yaml-editor {
  width: 100%;
  min-height: 400px;
  max-height: 500px;
  padding: 16px;
  border: 1px solid #334155;
  border-radius: 8px;
  font-family: 'Monaco', 'Menlo', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
  resize: vertical;
  background: #1e1e1e;
  color: #d4d4d4;
}

.yaml-editor:focus {
  outline: none;
  border-color: #326ce5;
  box-shadow: 0 0 0 3px rgba(50, 108, 229, 0.3);
  background: #1e1e1e;
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
  font-size: 14px;
}

.error-icon {
  font-size: 18px;
}

.yaml-editor-footer {
  background: #f7fafc;
  padding: 16px;
  border-radius: 8px;
}

.yaml-tips {
  font-size: 13px;
  color: #4a5568;
}

.yaml-tips strong {
  display: block;
  margin-bottom: 8px;
  color: #2d3748;
}

.yaml-tips ul {
  margin: 0;
  padding-left: 20px;
}

.yaml-tips li {
  margin-bottom: 4px;
  line-height: 1.6;
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none;
}

.btn-primary:disabled:hover {
  background-color: #326ce5;
  transform: none;
  box-shadow: none;
}
</style>