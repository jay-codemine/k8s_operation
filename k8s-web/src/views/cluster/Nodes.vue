<template>
  <div class="resource-view">
    <!-- 页面头部 -->
    <div class="view-header">
      <h1>节点管理</h1>
      <p>Kubernetes集群中的节点列表</p>
    </div>

    <!-- 操作栏 -->
    <div class="action-bar">
      <div class="search-box">
        <input
          type="text"
          v-model="searchQuery"
          placeholder="模糊搜索节点名称..."
          @input="onSearchInput"
        />
      </div>

      <!-- 状态筛选 -->
      <div class="filter-buttons">
        <button class="btn btn-filter" :class="{ active: statusFilter === 'all' }" @click="setStatusFilter('all')">
          全部
        </button>
        <button class="btn btn-filter" :class="{ active: statusFilter === 'Ready' }" @click="setStatusFilter('Ready')">
          Ready
        </button>
        <button class="btn btn-filter" :class="{ active: statusFilter === 'NotReady' }" @click="setStatusFilter('NotReady')">
          NotReady
        </button>
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

        <button class="btn btn-secondary" @click="refresh" :disabled="loading">
          {{ loading ? '加载中...' : '🔄 刷新' }}
        </button>
      </div>
    </div>

    <!-- 批量操作浮动栏 -->
    <div v-if="batchMode && selectedNodes.length > 0" class="batch-action-bar">
      <div class="batch-info">
        <span class="batch-count">已选择 {{ selectedNodes.length }} 个节点</span>
        <button class="batch-clear" @click="clearSelection">清空选择</button>
      </div>
      <div class="batch-actions">
        <button class="batch-btn" @click="openBatchLabelsModal" title="批量修改标签">
          🏷️ 批量标签
        </button>
        <button class="batch-btn" @click="openBatchTaintsModal" title="批量修改污点">
          ⚡ 批量污点
        </button>
        <button class="batch-btn" @click="batchCordon(true)" title="批量标记不可调度">
          🚫 批量 Cordon
        </button>
        <button class="batch-btn" @click="batchCordon(false)" title="批量取消不可调度">
          ✅ 批量 Uncordon
        </button>
        <button class="batch-btn danger" @click="openBatchDrainPreview" title="批量驱逐（高危）">
          ⚠️ 批量 Drain
        </button>
      </div>
    </div>

    <div v-if="errorMsg" class="error-box">
      {{ errorMsg }}
    </div>

    <!-- metrics 不可用提示 -->
    <div v-if="metricsUnavailable && !loading" class="metrics-warning">
      <span class="warning-icon">⚠️</span>
      <span class="warning-text">
        CPU/内存使用率数据不可用。请确保集群已安装 
        <a href="https://github.com/kubernetes-sigs/metrics-server" target="_blank">metrics-server</a>
      </span>
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
            <th>状态</th>
            <th>名称</th>
            <th>IP地址</th>
            <th>角色</th>
            <th>CPU使用</th>
            <th>内存使用</th>
            <th>Pod数量</th>
            <th>版本</th>
            <th>创建时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="node in paginatedNodes" :key="node.name" :class="{ 'row-selected': isNodeSelected(node) }">
            <td v-if="batchMode">
              <input 
                type="checkbox" 
                :checked="isNodeSelected(node)" 
                @change="toggleNodeSelection(node)"
              />
            </td>
            <td>
              <div class="status-cell">
                <span class="status-indicator" :class="node.status.toLowerCase()">
                  {{ node.status }}
                </span>
                <span v-if="node.unschedulable" class="unschedulable-badge-small">不可调度</span>
              </div>
            </td>
            <td>
              <div class="node-name">
                <span class="icon">🖥️</span>
                <span :title="node.name">{{ node.name }}</span>
              </div>
            </td>
            <td>{{ node.ip }}</td>
            <td>{{ node.role }}</td>
            <td>
              <div class="resource-usage">
                <div class="usage-bar">
                  <div class="usage-fill" :style="{ width: node.cpuUsage }" :class="getUsageClass(node.cpuUsage)"></div>
                </div>
                <span class="usage-text">{{ node.cpuUsage }}</span>
              </div>
            </td>
            <td>
              <div class="resource-usage">
                <div class="usage-bar">
                  <div class="usage-fill" :style="{ width: node.memoryUsage }" :class="getUsageClass(node.memoryUsage)"></div>
                </div>
                <span class="usage-text">{{ node.memoryUsage }}</span>
              </div>
            </td>
            <td>{{ node.podCount }}</td>
            <td>{{ node.version }}</td>
            <td class="mono">{{ node.createdAt }}</td>
            <td>
              <div class="action-icons">
                <button class="icon-btn" title="查看详情" @click="viewDetail(node)">
                  👁️ 详情
                </button>
                <div class="more-actions-wrapper">
                  <button class="icon-btn" @click="toggleMoreActions(node.name)" title="更多操作">
                    ⚙️ 更多 ▼
                  </button>
                  <div v-if="activeMoreMenu === node.name" class="more-actions-dropdown table-dropdown">
                    <button class="dropdown-item" @click="viewNodePods(node); activeMoreMenu = null">
                      📦 Pods
                    </button>
                    <button class="dropdown-item" @click="viewEvents(node); activeMoreMenu = null">
                      📋 事件
                    </button>
                    <button class="dropdown-item" @click="openLabelsModal(node); activeMoreMenu = null">
                      🏷️ 标签管理
                    </button>
                    <button class="dropdown-item" @click="openTaintsModal(node); activeMoreMenu = null">
                      ⚡ 污点管理
                    </button>
                    <button class="dropdown-item" @click="toggleCordon(node); activeMoreMenu = null">
                      {{ node.unschedulable ? '✅ Uncordon' : '🚫 Cordon' }}
                    </button>
                    <button class="dropdown-item danger" @click="confirmDrain(node); activeMoreMenu = null">
                      ⚠️ Drain
                    </button>
                    <button class="dropdown-item" @click="openYamlPreview(node); activeMoreMenu = null">
                      📝 查看/编辑 YAML
                    </button>
                  </div>
                </div>
              </div>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-if="!loading && paginatedNodes.length === 0" class="empty-state">
        <div class="empty-icon">🔍</div>
        <div class="empty-text">没有找到匹配的节点</div>
      </div>

      <Pagination
        v-if="total > 0"
        v-model:currentPage="currentPage"
        :totalItems="total"
        :itemsPerPage="itemsPerPage"
      />
    </div>

    <!-- 卡片视图 -->
    <div v-if="viewMode === 'card'" class="cards-container">
      <div v-if="!loading && paginatedNodes.length === 0" class="empty-state">
        <div class="empty-icon">🔍</div>
        <div class="empty-text">没有找到匹配的节点</div>
      </div>

      <div class="cards-grid">
        <div v-for="node in paginatedNodes" :key="node.name" class="node-card" :class="{ 'card-selected': isNodeSelected(node) }">
          <!-- 批量选择复选框 -->
          <div v-if="batchMode" class="card-checkbox">
            <input 
              type="checkbox" 
              :checked="isNodeSelected(node)" 
              @change="toggleNodeSelection(node)"
            />
          </div>
          
          <!-- 卡片头部 -->
          <div class="card-header">
            <div class="card-title-row">
              <span class="card-icon">🖥️</span>
              <h3 class="card-title" :title="node.name">{{ node.name }}</h3>
              <span class="status-indicator" :class="node.status.toLowerCase()">
                {{ node.status }}
              </span>
            </div>
            <div class="card-badges">
              <span class="role-badge" :class="node.role.toLowerCase().replace(' ', '-')">{{ node.role }}</span>
              <span v-if="node.unschedulable" class="unschedulable-badge">不可调度</span>
            </div>
          </div>

          <!-- 卡片主体 -->
          <div class="card-body">
            <!-- IP地址 -->
            <div class="card-section">
              <div class="section-label">IP 地址</div>
              <div class="meta-value mono">{{ node.ip }}</div>
            </div>

            <!-- 版本 -->
            <div class="card-section">
              <div class="section-label">Kubernetes 版本</div>
              <div class="meta-value">{{ node.version }}</div>
            </div>

            <!-- 资源使用情况 -->
            <div class="card-section">
              <div class="section-label">资源使用</div>
              <div class="resource-bars">
                <div class="resource-bar-item">
                  <div class="bar-label">
                    <span>⚡ CPU</span>
                    <span>{{ node.cpuUsage }}</span>
                  </div>
                  <div class="bar-track">
                    <div class="bar-fill" :style="{ width: node.cpuUsage }" :class="getUsageClass(node.cpuUsage)"></div>
                  </div>
                </div>
                <div class="resource-bar-item">
                  <div class="bar-label">
                    <span>💾 内存</span>
                    <span>{{ node.memoryUsage }}</span>
                  </div>
                  <div class="bar-track">
                    <div class="bar-fill" :style="{ width: node.memoryUsage }" :class="getUsageClass(node.memoryUsage)"></div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Pod数量 -->
            <div class="card-section">
              <div class="section-label">Pod 数量</div>
              <div class="meta-value">
                <span class="pod-count">{{ node.podCount }}</span>
              </div>
            </div>

            <!-- 创建时间 -->
            <div class="card-section">
              <div class="section-label">创建时间</div>
              <div class="meta-value mono">{{ node.createdAt }}</div>
            </div>
          </div>

          <!-- 卡片底部按钮 -->
          <div class="card-footer">
            <button class="card-action-btn" @click="viewDetail(node)" title="查看详情">
              👁️ 详情
            </button>
            <button class="card-action-btn" @click="viewNodePods(node)" title="查看Pod列表">
              📦 Pods
            </button>
            <button class="card-action-btn" @click="viewEvents(node)" title="查看事件">
              📋 事件
            </button>
            <div class="more-actions-wrapper">
              <button class="card-action-btn" @click="toggleMoreActions(node.name)" title="更多操作">
                ⚙️ 更多 ▼
              </button>
              <div v-if="activeMoreMenu === node.name" class="more-actions-dropdown">
                <button class="dropdown-item" @click="openLabelsModal(node); activeMoreMenu = null">
                  🏷️ 标签管理
                </button>
                <button class="dropdown-item" @click="openTaintsModal(node); activeMoreMenu = null">
                  ⚡ 污点管理
                </button>
                <button class="dropdown-item" @click="toggleCordon(node); activeMoreMenu = null">
                  {{ node.unschedulable ? '✅ Uncordon' : '🚫 Cordon' }}
                </button>
                <button class="dropdown-item danger" @click="confirmDrain(node); activeMoreMenu = null">
                  ⚠️ Drain
                </button>
                <button class="dropdown-item" @click="openYamlPreview(node); activeMoreMenu = null">
                  📝 查看/编辑 YAML
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <Pagination
        v-if="total > 0"
        v-model:currentPage="currentPage"
        :totalItems="total"
        :itemsPerPage="itemsPerPage"
      />
    </div>

    <!-- 详情弹窗 -->
    <div v-if="showDetailModal" class="modal-overlay" @click.self="closeDetail">
      <div class="modal-content">
        <div class="modal-header">
          <h3>🖥️ 节点详情</h3>
          <button class="close-btn" @click="closeDetail">×</button>
        </div>
        <div class="modal-body">
          <div v-if="loadingDetail" style="text-align: center; padding: 40px;">
            <div class="loading-spinner">加载中...</div>
          </div>
          <div v-else-if="detailData">
            <pre class="detail-json">{{ JSON.stringify(detailData, null, 2) }}</pre>
          </div>
        </div>
      </div>
    </div>

    <!-- Drain 确认弹窗 -->
    <div v-if="showDrainModal" class="modal-overlay" @click.self="showDrainModal = false">
      <div class="modal-content" style="max-width: 480px;">
        <div class="modal-header">
          <h3>⚠️ 驱逐节点 Pod</h3>
          <button class="close-btn" @click="showDrainModal = false">×</button>
        </div>
        <div class="modal-body">
          <p v-if="nodeToDrain" style="margin-bottom: 16px;">
            确认驱逐节点上的所有可驱逐 Pod：
          </p>
          <div v-if="nodeToDrain" style="background: #f7fafc; padding: 12px; border-radius: 8px; margin-bottom: 16px;">
            <div><strong>节点名称：</strong>{{ nodeToDrain.name }}</div>
            <div><strong>Pod 数量：</strong>{{ nodeToDrain.podCount }}</div>
          </div>
          <div style="background: #fff3cd; padding: 12px; border-radius: 8px; margin-bottom: 12px; border-left: 3px solid #ffc107;">
            <div style="color: #856404; font-size: 13px; line-height: 1.6;">
              ⚠️ 此操作将：<br/>
              1. 将节点标记为不可调度（Cordon）<br/>
              2. 驱逐节点上所有可驱逐的 Pod<br/>
              3. DaemonSet 和静态 Pod 不会被驱逐
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showDrainModal = false" :disabled="draining">
            取消
          </button>
          <button class="btn btn-danger" @click="confirmDrain" :disabled="draining">
            {{ draining ? '驱逐中...' : '确认驱逐' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 节点 Pod 列表弹窗 -->
    <div v-if="showPodsModal" class="modal-overlay" @click.self="closePodsModal">
      <div class="modal-content modal-pods">
        <div class="modal-header">
          <h3>📦 节点 Pod 列表</h3>
          <button class="close-btn" @click="closePodsModal">×</button>
        </div>
        <div class="modal-body">
          <div v-if="selectedNodeForPods" class="node-info-bar">
            <div class="node-info-item">
              <span class="info-label">节点名称</span>
              <span class="info-value">{{ selectedNodeForPods.name }}</span>
            </div>
            <div class="node-info-item">
              <span class="info-label">IP 地址</span>
              <span class="info-value">{{ selectedNodeForPods.ip }}</span>
            </div>
            <div class="node-info-item">
              <span class="info-label">Pod 数量</span>
              <span class="info-value pod-count">{{ nodePods.length }}</span>
            </div>
          </div>

          <div v-if="loadingPods" class="loading-state">
            <div class="loading-spinner">加载中...</div>
          </div>

          <div v-else-if="nodePods.length === 0" class="empty-state">
            <div class="empty-icon">📭</div>
            <div class="empty-text">该节点上暂无 Pod</div>
          </div>

          <div v-else class="pods-list">
            <div v-for="pod in nodePods" :key="pod.name" class="pod-item">
              <div class="pod-item-header">
                <span class="pod-icon">📦</span>
                <span class="pod-name">{{ pod.name }}</span>
                <span class="pod-status" :class="pod.status.toLowerCase()">{{ pod.status }}</span>
              </div>
              <div class="pod-item-meta">
                <span class="meta-item">
                  <span class="meta-label">命名空间:</span>
                  <span class="meta-value namespace-badge">{{ pod.namespace }}</span>
                </span>
                <span class="meta-item">
                  <span class="meta-label">IP:</span>
                  <span class="meta-value mono">{{ pod.ip }}</span>
                </span>
                <span class="meta-item">
                  <span class="meta-label">创建时间:</span>
                  <span class="meta-value mono">{{ pod.createdAt }}</span>
                </span>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closePodsModal">关闭</button>
        </div>
      </div>
    </div>

    <!-- 节点事件弹窗 -->
    <div v-if="showEventsModal" class="modal-overlay" @click.self="closeEventsModal">
      <div class="modal-content modal-events">
        <div class="modal-header">
          <h3>📋 节点事件 - {{ selectedNodeForEvents?.name }}</h3>
          <button class="close-btn" @click="closeEventsModal">×</button>
        </div>
        <div class="modal-body">
          <div v-if="loadingEvents" class="loading-state">
            <div class="loading-spinner">加载中...</div>
          </div>
          <div v-else-if="nodeEvents.length === 0" class="empty-state">
            <div class="empty-icon">📭</div>
            <div class="empty-text">暂无事件记录</div>
          </div>
          <div v-else class="events-list">
            <div v-for="(event, index) in nodeEvents" :key="index" class="event-item" :class="event.type.toLowerCase()">
              <div class="event-header">
                <span class="event-type" :class="event.type.toLowerCase()">{{ event.type }}</span>
                <span class="event-reason">{{ event.reason }}</span>
                <span class="event-count" v-if="event.count > 1">×{{ event.count }}</span>
              </div>
              <div class="event-message">{{ event.message }}</div>
              <div class="event-meta">
                <span>来源: {{ event.source }}</span>
                <span>最后发生: {{ formatEventTime(event.last_timestamp) }}</span>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeEventsModal">关闭</button>
        </div>
      </div>
    </div>

    <!-- 标签管理弹窗 -->
    <div v-if="showLabelsModal" class="modal-overlay" @click.self="closeLabelsModal">
      <div class="modal-content modal-labels">
        <div class="modal-header">
          <h3>🏷️ 标签管理 - {{ selectedNodeForLabels?.name }}</h3>
          <button class="close-btn" @click="closeLabelsModal">×</button>
        </div>
        <div class="modal-body">
          <!-- 现有标签 -->
          <div class="labels-section">
            <div class="section-title">现有标签</div>
            <div class="labels-list" v-if="currentLabels && Object.keys(currentLabels).length">
              <div v-for="(value, key) in currentLabels" :key="key" class="label-item">
                <span class="label-key">{{ key }}</span>
                <span class="label-value">{{ value }}</span>
                <button class="label-remove-btn" @click="removeLabel(key)" :disabled="isSystemLabel(key)">×</button>
              </div>
            </div>
            <div v-else class="empty-labels">暂无标签</div>
          </div>
          <!-- 添加标签 -->
          <div class="add-label-section">
            <div class="section-title">添加标签</div>
            <div class="add-label-form">
              <input v-model="newLabelKey" placeholder="键 (Key)" class="label-input" />
              <input v-model="newLabelValue" placeholder="值 (Value)" class="label-input" />
              <button class="btn btn-primary btn-sm" @click="addLabel" :disabled="!newLabelKey || savingLabels">
                添加
              </button>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeLabelsModal">关闭</button>
        </div>
      </div>
    </div>

    <!-- 污点管理弹窗 -->
    <div v-if="showTaintsModal" class="modal-overlay" @click.self="closeTaintsModal">
      <div class="modal-content modal-taints">
        <div class="modal-header">
          <h3>⚡ 污点管理 - {{ selectedNodeForTaints?.name }}</h3>
          <button class="close-btn" @click="closeTaintsModal">×</button>
        </div>
        <div class="modal-body">
          <!-- 现有污点 -->
          <div class="taints-section">
            <div class="section-title">现有污点</div>
            <div class="taints-list" v-if="currentTaints && currentTaints.length">
              <div v-for="(taint, index) in currentTaints" :key="index" class="taint-item">
                <span class="taint-key">{{ taint.key }}</span>
                <span class="taint-operator">=</span>
                <span class="taint-value">{{ taint.value || '(空)' }}</span>
                <span class="taint-effect" :class="taint.effect.toLowerCase()">{{ taint.effect }}</span>
                <button class="taint-remove-btn" @click="removeTaint(taint.key)">×</button>
              </div>
            </div>
            <div v-else class="empty-taints">暂无污点</div>
          </div>
          <!-- 添加污点 -->
          <div class="add-taint-section">
            <div class="section-title">添加污点</div>
            <div class="add-taint-form">
              <input v-model="newTaintKey" placeholder="键 (Key)" class="taint-input" />
              <input v-model="newTaintValue" placeholder="值 (Value, 可选)" class="taint-input" />
              <select v-model="newTaintEffect" class="taint-select">
                <option value="NoSchedule">NoSchedule</option>
                <option value="PreferNoSchedule">PreferNoSchedule</option>
                <option value="NoExecute">NoExecute</option>
              </select>
              <button class="btn btn-primary btn-sm" @click="addTaint" :disabled="!newTaintKey || savingTaints">
                添加
              </button>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeTaintsModal">关闭</button>
        </div>
      </div>
    </div>

    <!-- 批量标签预览弹窗 -->
    <div v-if="showBatchLabelsModal" class="modal-overlay" @click.self="closeBatchLabelsModal">
      <div class="modal-content modal-batch-preview">
        <div class="modal-header">
          <h3>🏷️ 批量修改标签预览</h3>
          <button class="close-btn" @click="closeBatchLabelsModal">×</button>
        </div>
        <div class="modal-body">
          <!-- 受影响节点列表 -->
          <div class="preview-section">
            <div class="section-title">受影响节点 ({{ selectedNodes.length }})</div>
            <div class="affected-nodes">
              <span v-for="node in selectedNodes" :key="node.name" class="affected-node-tag">
                {{ node.name }}
              </span>
            </div>
          </div>
          
          <!-- 添加标签 -->
          <div class="preview-section">
            <div class="section-title">添加标签</div>
            <div class="batch-label-form">
              <input v-model="batchLabelKey" placeholder="键 (Key)" class="label-input" />
              <input v-model="batchLabelValue" placeholder="值 (Value)" class="label-input" />
              <button class="btn btn-primary btn-sm" @click="addBatchLabel" :disabled="!batchLabelKey">
                添加
              </button>
            </div>
            <div class="pending-labels" v-if="batchLabelsToAdd.length">
              <div v-for="(label, index) in batchLabelsToAdd" :key="index" class="pending-label-item">
                <span class="label-key">{{ label.key }}</span>
                <span class="label-value">{{ label.value }}</span>
                <button class="label-remove-btn" @click="removeBatchLabel(index)">×</button>
              </div>
            </div>
          </div>

          <!-- 预览变更 -->
          <div v-if="batchLabelsToAdd.length" class="preview-section">
            <div class="section-title">变更预览</div>
            <div class="change-preview">
              <div class="change-item add" v-for="label in batchLabelsToAdd" :key="label.key">
                <span class="change-icon">+</span>
                <span>{{ label.key }}={{ label.value }}</span>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeBatchLabelsModal">取消</button>
          <button 
            class="btn btn-primary" 
            @click="executeBatchLabels" 
            :disabled="batchLabelsToAdd.length === 0 || batchExecuting"
          >
            {{ batchExecuting ? '执行中...' : '确认执行' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 批量污点预览弹窗 -->
    <div v-if="showBatchTaintsModal" class="modal-overlay" @click.self="closeBatchTaintsModal">
      <div class="modal-content modal-batch-preview">
        <div class="modal-header">
          <h3>⚡ 批量修改污点预览</h3>
          <button class="close-btn" @click="closeBatchTaintsModal">×</button>
        </div>
        <div class="modal-body">
          <!-- 受影响节点列表 -->
          <div class="preview-section">
            <div class="section-title">受影响节点 ({{ selectedNodes.length }})</div>
            <div class="affected-nodes">
              <span v-for="node in selectedNodes" :key="node.name" class="affected-node-tag">
                {{ node.name }}
              </span>
            </div>
          </div>
          
          <!-- 添加污点 -->
          <div class="preview-section">
            <div class="section-title">添加污点</div>
            <div class="batch-taint-form">
              <input v-model="batchTaintKey" placeholder="键 (Key)" class="taint-input" />
              <input v-model="batchTaintValue" placeholder="值 (可选)" class="taint-input" />
              <select v-model="batchTaintEffect" class="taint-select">
                <option value="NoSchedule">NoSchedule</option>
                <option value="PreferNoSchedule">PreferNoSchedule</option>
                <option value="NoExecute">NoExecute</option>
              </select>
              <button class="btn btn-primary btn-sm" @click="addBatchTaint" :disabled="!batchTaintKey">
                添加
              </button>
            </div>
            <div class="pending-taints" v-if="batchTaintsToAdd.length">
              <div v-for="(taint, index) in batchTaintsToAdd" :key="index" class="pending-taint-item">
                <span class="taint-key">{{ taint.key }}</span>
                <span class="taint-value">{{ taint.value || '(空)' }}</span>
                <span class="taint-effect" :class="taint.effect.toLowerCase()">{{ taint.effect }}</span>
                <button class="taint-remove-btn" @click="removeBatchTaint(index)">×</button>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeBatchTaintsModal">取消</button>
          <button 
            class="btn btn-primary" 
            @click="executeBatchTaints" 
            :disabled="batchTaintsToAdd.length === 0 || batchExecuting"
          >
            {{ batchExecuting ? '执行中...' : '确认执行' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 批量Drain预览弹窗（高危操作） -->
    <div v-if="showBatchDrainModal" class="modal-overlay" @click.self="closeBatchDrainModal">
      <div class="modal-content modal-batch-preview modal-danger">
        <div class="modal-header danger-header">
          <h3>⚠️ 批量 Drain 预览（高危操作）</h3>
          <button class="close-btn" @click="closeBatchDrainModal">×</button>
        </div>
        <div class="modal-body">
          <!-- 警告提示 -->
          <div class="danger-warning">
            <div class="warning-icon-large">⚠️</div>
            <div class="warning-content">
              <div class="warning-title">此操作将驱逐以下节点上的所有可驱逐 Pod</div>
              <ul class="warning-list">
                <li>节点将被标记为不可调度（Cordon）</li>
                <li>所有非 DaemonSet、非静态 Pod 将被驱逐</li>
                <li>此操作不可撤销，请确认！</li>
              </ul>
            </div>
          </div>
          
          <!-- 受影响节点及Pod统计 -->
          <div class="preview-section">
            <div class="section-title">受影响节点 ({{ selectedNodes.length }})</div>
            <div class="affected-nodes-detail">
              <div v-for="node in selectedNodes" :key="node.name" class="affected-node-card">
                <div class="node-info">
                  <span class="node-name">🖥️ {{ node.name }}</span>
                  <span class="node-ip">{{ node.ip }}</span>
                </div>
                <div class="node-stats">
                  <span class="stat-item">📦 {{ node.podCount }} Pods</span>
                  <span v-if="node.unschedulable" class="already-cordoned">已 Cordon</span>
                </div>
              </div>
            </div>
          </div>

          <!-- 二次确认 -->
          <div class="confirm-section">
            <div class="section-title">请输入 "DRAIN" 确认操作</div>
            <input 
              v-model="drainConfirmText" 
              placeholder="请输入 DRAIN" 
              class="confirm-input"
              :class="{ valid: drainConfirmText === 'DRAIN' }"
            />
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeBatchDrainModal">取消</button>
          <button 
            class="btn btn-danger" 
            @click="executeBatchDrain" 
            :disabled="drainConfirmText !== 'DRAIN' || batchExecuting"
          >
            {{ batchExecuting ? '执行中...' : '确认驱逐' }}
          </button>
        </div>
      </div>
    </div>

    <!-- YAML 查看/编辑模态框 -->
    <div v-if="showYamlModal" class="modal-overlay" @click.self="closeYamlModal">
      <div class="modal-content yaml-modal">
        <div class="modal-header">
          <h3>📝 {{ yamlEditMode ? '编辑 YAML' : '查看 YAML' }} - {{ selectedYamlNode?.name }}</h3>
          <div class="yaml-header-actions">
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
          <div v-if="loadingYaml" class="loading-state">
            <div class="loading-spinner">Loading YAML...</div>
          </div>
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
import nodesApi from '@/api/cluster/nodes'

// ========== 状态变量 ==========
const searchQuery = ref('')
const currentPage = ref(1)
const itemsPerPage = ref(10)
const loading = ref(false)
const viewMode = ref('card') // 'table' | 'card'
const statusFilter = ref('all')
const errorMsg = ref('')
let searchDebounceTimer = null

// 模态框状态
const showDetailModal = ref(false)
const showDrainModal = ref(false)
const showPodsModal = ref(false)
const showEventsModal = ref(false)
const showLabelsModal = ref(false)
const showTaintsModal = ref(false)
const loadingDetail = ref(false)
const loadingPods = ref(false)
const loadingEvents = ref(false)
const draining = ref(false)
const savingLabels = ref(false)
const savingTaints = ref(false)

// 数据
const nodes = ref([])
const nodeToDrain = ref(null)
const detailData = ref(null)
const nodePods = ref([])
const selectedNodeForPods = ref(null)
const metricsUnavailable = ref(false) // metrics 不可用标志

// 事件相关
const nodeEvents = ref([])
const selectedNodeForEvents = ref(null)

// 标签相关
const selectedNodeForLabels = ref(null)
const currentLabels = ref({})
const newLabelKey = ref('')
const newLabelValue = ref('')

// 污点相关
const selectedNodeForTaints = ref(null)
const currentTaints = ref([])
const newTaintKey = ref('')
const newTaintValue = ref('')
const newTaintEffect = ref('NoSchedule')

// ========== YAML 查看/编辑相关 ==========
const showYamlModal = ref(false)
const selectedYamlNode = ref(null)
const yamlContent = ref('')
const yamlEditMode = ref(false)
const loadingYaml = ref(false)
const savingYaml = ref(false)
const yamlError = ref('')

// ========== 批量操作相关 ==========
const batchMode = ref(false)
const selectedNodes = ref([])
const showBatchLabelsModal = ref(false)
const showBatchTaintsModal = ref(false)
const showBatchDrainModal = ref(false)
const batchExecuting = ref(false)

// 批量标签
const batchLabelKey = ref('')
const batchLabelValue = ref('')
const batchLabelsToAdd = ref([])

// 批量污点
const batchTaintKey = ref('')
const batchTaintValue = ref('')
const batchTaintEffect = ref('NoSchedule')
const batchTaintsToAdd = ref([])

// 批量Drain确认
const drainConfirmText = ref('')

// 批量操作方法
const enterBatchMode = () => {
  batchMode.value = true
  // 默认全选当前页，用户可取消不需要的项
  selectedNodes.value = [...paginatedNodes.value]
}

const exitBatchMode = () => {
  batchMode.value = false
  selectedNodes.value = []
}

const clearSelection = () => {
  selectedNodes.value = []
}

const isNodeSelected = (node) => {
  return selectedNodes.value.some(n => n.name === node.name)
}

const toggleNodeSelection = (node) => {
  const index = selectedNodes.value.findIndex(n => n.name === node.name)
  if (index >= 0) {
    selectedNodes.value.splice(index, 1)
  } else {
    selectedNodes.value.push(node)
  }
}

const isAllSelected = computed(() => {
  return paginatedNodes.value.length > 0 && 
         paginatedNodes.value.every(node => isNodeSelected(node))
})

// 部分选中状态
const isPartialSelected = computed(() => {
  if (paginatedNodes.value.length === 0) return false
  const selectedCount = paginatedNodes.value.filter(node => isNodeSelected(node)).length
  return selectedCount > 0 && selectedCount < paginatedNodes.value.length
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
    // 取消选择当前页的所有节点
    paginatedNodes.value.forEach(node => {
      const index = selectedNodes.value.findIndex(n => n.name === node.name)
      if (index >= 0) selectedNodes.value.splice(index, 1)
    })
  } else {
    // 选择当前页的所有节点
    paginatedNodes.value.forEach(node => {
      if (!isNodeSelected(node)) {
        selectedNodes.value.push(node)
      }
    })
  }
}

// 批量标签
const openBatchLabelsModal = () => {
  batchLabelsToAdd.value = []
  batchLabelKey.value = ''
  batchLabelValue.value = ''
  showBatchLabelsModal.value = true
}

const closeBatchLabelsModal = () => {
  showBatchLabelsModal.value = false
}

const addBatchLabel = () => {
  if (!batchLabelKey.value) return
  batchLabelsToAdd.value.push({
    key: batchLabelKey.value,
    value: batchLabelValue.value
  })
  batchLabelKey.value = ''
  batchLabelValue.value = ''
}

const removeBatchLabel = (index) => {
  batchLabelsToAdd.value.splice(index, 1)
}

const executeBatchLabels = async () => {
  batchExecuting.value = true
  let successCount = 0
  let failCount = 0
  
  for (const node of selectedNodes.value) {
    try {
      const addLabels = {}
      batchLabelsToAdd.value.forEach(l => {
        addLabels[l.key] = l.value
      })
      await nodesApi.patchLabels({
        name: node.name,
        add: addLabels,
        remove: []
      })
      successCount++
    } catch (e) {
      console.error(`Failed to patch labels for ${node.name}:`, e)
      failCount++
    }
  }
  
  batchExecuting.value = false
  showBatchLabelsModal.value = false
  
  if (failCount === 0) {
    Message.success({ content: `成功为 ${successCount} 个节点添加标签`, duration: 2200 })
  } else {
    Message.warning({ content: `成功 ${successCount} 个，失败 ${failCount} 个`, duration: 2200 })
  }
  
  refresh()
}

// 批量污点
const openBatchTaintsModal = () => {
  batchTaintsToAdd.value = []
  batchTaintKey.value = ''
  batchTaintValue.value = ''
  batchTaintEffect.value = 'NoSchedule'
  showBatchTaintsModal.value = true
}

const closeBatchTaintsModal = () => {
  showBatchTaintsModal.value = false
}

const addBatchTaint = () => {
  if (!batchTaintKey.value) return
  batchTaintsToAdd.value.push({
    key: batchTaintKey.value,
    value: batchTaintValue.value,
    effect: batchTaintEffect.value
  })
  batchTaintKey.value = ''
  batchTaintValue.value = ''
}

const removeBatchTaint = (index) => {
  batchTaintsToAdd.value.splice(index, 1)
}

const executeBatchTaints = async () => {
  batchExecuting.value = true
  let successCount = 0
  let failCount = 0
  
  for (const node of selectedNodes.value) {
    try {
      await nodesApi.patchTaints({
        name: node.name,
        add: batchTaintsToAdd.value,
        removeKeys: []
      })
      successCount++
    } catch (e) {
      console.error(`Failed to patch taints for ${node.name}:`, e)
      failCount++
    }
  }
  
  batchExecuting.value = false
  showBatchTaintsModal.value = false
  
  if (failCount === 0) {
    Message.success({ content: `成功为 ${successCount} 个节点添加污点`, duration: 2200 })
  } else {
    Message.warning({ content: `成功 ${successCount} 个，失败 ${failCount} 个`, duration: 2200 })
  }
  
  refresh()
}

// 批量 Cordon
const batchCordon = async (unschedulable) => {
  const action = unschedulable ? 'Cordon' : 'Uncordon'
  
  // 二次确认
  const nodeNames = selectedNodes.value.map(n => n.name).join('\n  - ')
  const confirmMsg = unschedulable
    ? `⚠️ 确认将以下 ${selectedNodes.value.length} 个节点标记为不可调度？

  - ${nodeNames}

标记后，新的 Pod 将不会被调度到这些节点。`
    : `确认取消以下 ${selectedNodes.value.length} 个节点的不可调度状态？

  - ${nodeNames}

取消后，新的 Pod 可以被调度到这些节点。`
  
  if (!confirm(confirmMsg)) return
  
  let successCount = 0
  let failCount = 0
  
  for (const node of selectedNodes.value) {
    try {
      await nodesApi.cordon({
        nodeName: node.name,
        unschedulable: unschedulable
      })
      successCount++
    } catch (e) {
      console.error(`Failed to ${action} ${node.name}:`, e)
      failCount++
    }
  }
  
  if (failCount === 0) {
    Message.success({ content: `成功 ${action} ${successCount} 个节点`, duration: 2200 })
  } else {
    Message.warning({ content: `成功 ${successCount} 个，失败 ${failCount} 个`, duration: 2200 })
  }
  
  refresh()
}

// 批量 Drain
const openBatchDrainPreview = () => {
  drainConfirmText.value = ''
  showBatchDrainModal.value = true
}

const closeBatchDrainModal = () => {
  showBatchDrainModal.value = false
  drainConfirmText.value = ''
}

const executeBatchDrain = async () => {
  if (drainConfirmText.value !== 'DRAIN') return
  
  batchExecuting.value = true
  let successCount = 0
  let failCount = 0
  
  for (const node of selectedNodes.value) {
    try {
      await nodesApi.drain({
        nodeName: node.name
      })
      successCount++
    } catch (e) {
      console.error(`Failed to drain ${node.name}:`, e)
      failCount++
    }
  }
  
  batchExecuting.value = false
  showBatchDrainModal.value = false
  drainConfirmText.value = ''
  
  if (failCount === 0) {
    Message.success({ content: `成功 Drain ${successCount} 个节点`, duration: 2200 })
  } else {
    Message.warning({ content: `成功 ${successCount} 个，失败 ${failCount} 个`, duration: 2200 })
  }
  
  exitBatchMode()
  refresh()
}

// ========== 计算属性 ==========
const filteredNodes = computed(() => {
  let result = nodes.value

  // 状态筛选
  if (statusFilter.value !== 'all') {
    result = result.filter(node => node.status === statusFilter.value)
  }

  // 搜索筛选
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(node =>
      node.name.toLowerCase().includes(query) ||
      node.ip.includes(query)
    )
  }

  return result
})

const paginatedNodes = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage.value
  return filteredNodes.value.slice(start, start + itemsPerPage.value)
})

const total = computed(() => filteredNodes.value.length)

// ========== 生命周期 ==========
onMounted(() => {
  fetchNodes()
})

onUnmounted(() => {
  if (searchDebounceTimer) clearTimeout(searchDebounceTimer)
  stopAutoRefresh()
})

// ========== 自动刷新 ==========
const autoRefresh = ref(false)
let autoRefreshTimer = null
const AUTO_REFRESH_INTERVAL = 90000 // 90秒

const startAutoRefresh = () => {
  if (autoRefreshTimer) return
  autoRefreshTimer = setInterval(() => {
    if (!loading.value) {
      fetchNodes()
    }
  }, AUTO_REFRESH_INTERVAL)
}

const stopAutoRefresh = () => {
  if (autoRefreshTimer) {
    clearInterval(autoRefreshTimer)
    autoRefreshTimer = null
  }
}

watch(autoRefresh, (val) => {
  if (val) {
    startAutoRefresh()
  } else {
    stopAutoRefresh()
  }
})

// ========== 方法 ==========
const fetchNodes = async () => {
  loading.value = true
  errorMsg.value = ''
  metricsUnavailable.value = false
  try {
    const response = await nodesApi.list({
      name: '',
      page: 1,
      limit: 1000,
    })

    const list = response?.data?.list || response?.data?.items || response?.list || response?.items || []

    // 并行获取所有节点的指标数据
    let metricsList = []
    try {
      const metricsResponse = await nodesApi.metrics({ name: '' })
      metricsList = metricsResponse?.data?.list || []
      if (metricsList.length === 0) {
        metricsUnavailable.value = true
      }
    } catch (err) {
      metricsUnavailable.value = true
      console.warn('获取节点指标失败:', err?.response?.data?.message || err)
    }

    const metricsMap = {}
    metricsList.forEach(m => {
      // 后端返回的字段是 name
      metricsMap[m.name] = m
    })

    nodes.value = Array.isArray(list) ? list.map(node => {
      const metrics = metricsMap[node.name] || {}
      // 格式化 CPU/内存使用率（后端返回的是数字，需要转成百分比字符串）
      const cpuPercent = metrics.cpu_usage_percent
      const memPercent = metrics.mem_usage_percent
      
      return {
        name: node.name || '',
        status: node.status || 'Unknown',
        ip: node.ip || '-',
        role: node.role || 'Worker',
        cpuUsage: cpuPercent ? `${cpuPercent.toFixed(1)}%` : '-',
        memoryUsage: memPercent ? `${memPercent.toFixed(1)}%` : '-',
        podCount: node.pod_count || 0,
        version: node.version || '-',
        createdAt: formatDate(node.creation_timestamp),
        unschedulable: node.unschedulable || false,
        labels: node.labels || {},
        cpuCapacity: node.cpu_capacity || '-',
        memoryCapacity: node.memory_capacity || '-',
      }
    }) : []

  } catch (error) {
    console.error('Failed to fetch nodes:', error)
    errorMsg.value = '获取节点列表失败'
    Message.error({ content: '获取节点列表失败', duration: 2200 })
    nodes.value = []
  } finally {
    loading.value = false
  }
}

const refresh = () => {
  fetchNodes()
}

const setStatusFilter = (status) => {
  statusFilter.value = status
  currentPage.value = 1
}

const onSearchInput = () => {
  if (searchDebounceTimer) clearTimeout(searchDebounceTimer)
  searchDebounceTimer = setTimeout(() => {
    currentPage.value = 1
  }, 300)
}

// ========== 更多操作下拉菜单 ==========
const activeMoreMenu = ref(null)

const toggleMoreActions = (nodeName) => {
  if (activeMoreMenu.value === nodeName) {
    activeMoreMenu.value = null
  } else {
    activeMoreMenu.value = nodeName
  }
}

// 点击外部关闭菜单
onMounted(() => {
  document.addEventListener('click', (e) => {
    if (!e.target.closest('.more-actions-wrapper')) {
      activeMoreMenu.value = null
    }
  })
})

const getUsageClass = (usage) => {
  const percentage = parseInt(usage)
  if (percentage > 80) return 'usage-high'
  if (percentage > 50) return 'usage-medium'
  return 'usage-low'
}

// 格式化时间
const formatDate = (timestamp) => {
  if (!timestamp) return '-'
  const date = new Date(timestamp)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

// 查看详情
const viewDetail = async (node) => {
  showDetailModal.value = true
  loadingDetail.value = true
  try {
    const response = await nodesApi.detail({ name: node.name })
    detailData.value = response?.data || response
  } catch (error) {
    console.error('Failed to fetch node detail:', error)
    Message.error({ content: '获取节点详情失败', duration: 2200 })
    detailData.value = null
  } finally {
    loadingDetail.value = false
  }
}

const closeDetail = () => {
  showDetailModal.value = false
  detailData.value = null
}

// 查看节点上的 Pods
const viewNodePods = async (node) => {
  selectedNodeForPods.value = node
  showPodsModal.value = true
  loadingPods.value = true
  nodePods.value = []
  
  try {
    const response = await nodesApi.listPods({
      name: node.name,
      page: 1,
      limit: 100,
    })
    
    const list = response?.data?.list || response?.data?.items || response?.list || []
    nodePods.value = Array.isArray(list) ? list.map(pod => ({
      name: pod.name || pod.metadata?.name || '',
      namespace: pod.namespace || pod.metadata?.namespace || '',
      status: pod.status?.phase || 'Unknown',
      ip: pod.status?.podIP || '-',
      createdAt: formatDate(pod.metadata?.creationTimestamp),
    })) : []
  } catch (error) {
    console.error('Failed to fetch node pods:', error)
    Message.error({ content: '获取节点 Pod 列表失败', duration: 2200 })
    nodePods.value = []
  } finally {
    loadingPods.value = false
  }
}

const closePodsModal = () => {
  showPodsModal.value = false
  nodePods.value = []
  selectedNodeForPods.value = null
}

// Cordon/Uncordon
const toggleCordon = async (node) => {
  const action = node.unschedulable ? 'Uncordon' : 'Cordon'
  const newState = !node.unschedulable
  
  // 二次确认
  const confirmMsg = node.unschedulable 
    ? `确认取消节点 "${node.name}" 的不可调度状态？\n\n取消后，新的 Pod 可以被调度到此节点。`
    : `⚠️ 确认将节点 "${node.name}" 标记为不可调度？\n\n标记后，新的 Pod 将不会被调度到此节点，但已有 Pod 不受影响。`
  
  if (!confirm(confirmMsg)) return

  try {
    await nodesApi.cordon({
      nodeName: node.name,
      unschedulable: newState,
    })

    Message.success({ content: `${action} 成功`, duration: 2200 })
    // 立即更新本地状态
    node.unschedulable = newState
    // 3秒后自动刷新获取最新数据
    setTimeout(() => refresh(), 3000)
  } catch (error) {
    console.error(`Failed to ${action} node:`, error)
    Message.error({ content: `${action} 失败`, duration: 2200 })
  }
}

// Drain
const confirmDrain = (node) => {
  if (typeof node === 'object') {
    nodeToDrain.value = node
    showDrainModal.value = true
  } else {
    // 点击确认按钮
    performDrain()
  }
}

const performDrain = async () => {
  if (!nodeToDrain.value) return

  draining.value = true
  try {
    await nodesApi.drain({
      nodeName: nodeToDrain.value.name,
    })

    Message.success({ content: 'Drain 成功', duration: 2200 })
    showDrainModal.value = false
    nodeToDrain.value = null
    refresh()
  } catch (error) {
    console.error('Failed to drain node:', error)
    Message.error({ content: 'Drain 失败', duration: 2200 })
  } finally {
    draining.value = false
  }
}

// ========== 事件管理 ==========
const viewEvents = async (node) => {
  selectedNodeForEvents.value = node
  showEventsModal.value = true
  loadingEvents.value = true
  nodeEvents.value = []
  
  try {
    const response = await nodesApi.events({ name: node.name, limit: 50 })
    nodeEvents.value = response?.data?.list || []
  } catch (error) {
    console.error('Failed to fetch node events:', error)
    Message.error({ content: '获取节点事件失败', duration: 2200 })
  } finally {
    loadingEvents.value = false
  }
}

const closeEventsModal = () => {
  showEventsModal.value = false
  nodeEvents.value = []
  selectedNodeForEvents.value = null
}

const formatEventTime = (timestamp) => {
  if (!timestamp) return '-'
  const date = new Date(timestamp)
  return date.toLocaleString('zh-CN')
}

// ========== 标签管理 ==========
const openLabelsModal = async (node) => {
  selectedNodeForLabels.value = node
  showLabelsModal.value = true
  newLabelKey.value = ''
  newLabelValue.value = ''
  
  // 获取当前标签
  try {
    const response = await nodesApi.detail({ name: node.name })
    currentLabels.value = response?.data?.labels || {}
  } catch (error) {
    console.error('Failed to fetch node labels:', error)
    currentLabels.value = node.labels || {}
  }
}

const closeLabelsModal = () => {
  showLabelsModal.value = false
  selectedNodeForLabels.value = null
  currentLabels.value = {}
}

const isSystemLabel = (key) => {
  // 系统标签不允许删除
  return key.startsWith('kubernetes.io/') || 
         key.startsWith('node-role.kubernetes.io/') ||
         key.startsWith('node.kubernetes.io/')
}

const addLabel = async () => {
  if (!newLabelKey.value || !selectedNodeForLabels.value) return
  
  savingLabels.value = true
  try {
    await nodesApi.patchLabels({
      name: selectedNodeForLabels.value.name,
      add: { [newLabelKey.value]: newLabelValue.value },
      remove: [],
    })
    
    currentLabels.value[newLabelKey.value] = newLabelValue.value
    newLabelKey.value = ''
    newLabelValue.value = ''
    Message.success({ content: '标签添加成功', duration: 2200 })
  } catch (error) {
    console.error('Failed to add label:', error)
    Message.error({ content: '添加标签失败', duration: 2200 })
  } finally {
    savingLabels.value = false
  }
}

const removeLabel = async (key) => {
  if (!selectedNodeForLabels.value || isSystemLabel(key)) return
  
  savingLabels.value = true
  try {
    await nodesApi.patchLabels({
      name: selectedNodeForLabels.value.name,
      add: {},
      remove: [key],
    })
    
    delete currentLabels.value[key]
    Message.success({ content: '标签删除成功', duration: 2200 })
  } catch (error) {
    console.error('Failed to remove label:', error)
    Message.error({ content: '删除标签失败', duration: 2200 })
  } finally {
    savingLabels.value = false
  }
}

// ========== 污点管理 ==========
const openTaintsModal = async (node) => {
  selectedNodeForTaints.value = node
  showTaintsModal.value = true
  newTaintKey.value = ''
  newTaintValue.value = ''
  newTaintEffect.value = 'NoSchedule'
  
  // 获取当前污点
  try {
    const response = await nodesApi.detail({ name: node.name })
    currentTaints.value = response?.data?.taints || []
  } catch (error) {
    console.error('Failed to fetch node taints:', error)
    currentTaints.value = []
  }
}

const closeTaintsModal = () => {
  showTaintsModal.value = false
  selectedNodeForTaints.value = null
  currentTaints.value = []
}

const addTaint = async () => {
  if (!newTaintKey.value || !selectedNodeForTaints.value) return
  
  savingTaints.value = true
  try {
    await nodesApi.patchTaints({
      name: selectedNodeForTaints.value.name,
      add: [{
        key: newTaintKey.value,
        value: newTaintValue.value,
        effect: newTaintEffect.value,
      }],
      removeKeys: [],
    })
    
    currentTaints.value.push({
      key: newTaintKey.value,
      value: newTaintValue.value,
      effect: newTaintEffect.value,
    })
    newTaintKey.value = ''
    newTaintValue.value = ''
    Message.success({ content: '污点添加成功', duration: 2200 })
  } catch (error) {
    console.error('Failed to add taint:', error)
    Message.error({ content: '添加污点失败', duration: 2200 })
  } finally {
    savingTaints.value = false
  }
}

const removeTaint = async (key) => {
  if (!selectedNodeForTaints.value) return
  
  savingTaints.value = true
  try {
    await nodesApi.patchTaints({
      name: selectedNodeForTaints.value.name,
      add: [],
      removeKeys: [key],
    })
    
    currentTaints.value = currentTaints.value.filter(t => t.key !== key)
    Message.success({ content: '污点删除成功', duration: 2200 })
  } catch (error) {
    console.error('Failed to remove taint:', error)
    Message.error({ content: '删除污点失败', duration: 2200 })
  } finally {
    savingTaints.value = false
  }
}

// ========== YAML 查看/编辑功能 ==========
const openYamlPreview = async (node) => {
  activeMoreMenu.value = null
  selectedYamlNode.value = node
  yamlContent.value = ''
  yamlError.value = ''
  yamlEditMode.value = false
  loadingYaml.value = true
  showYamlModal.value = true

  try {
    const res = await nodesApi.yaml({ name: node.name })
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
  selectedYamlNode.value = null
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
    const res = await nodesApi.applyYaml({
      name: selectedYamlNode.value.name,
      yaml: yamlContent.value
    })
    if (res.code === 0) {
      Message.success({ content: 'YAML 应用成功' })
      closeYamlModal()
      refresh()
      autoRefresh.value = true
      setTimeout(() => { autoRefresh.value = false }, 15000)
    } else {
      Message.error({ content: res.msg || '应用 YAML 失败' })
    }
  } catch (e) {
    Message.error({ content: e?.msg || e?.message || '应用 YAML 失败' })
  } finally {
    savingYaml.value = false
  }
}
</script>

<style scoped>
/* 参考 Pods.vue 样式，部分简化 */
.resource-view {
  max-width: 1400px;
  margin: 0 auto;
  padding: 20px;
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
  gap: 16px;
  align-items: center;
  margin-bottom: 24px;
  flex-wrap: wrap;
}

.search-box input {
  padding: 10px 16px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  width: 260px;
}

.filter-buttons {
  display: flex;
  gap: 8px;
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-filter {
  padding: 8px 16px;
  background: white;
  border: 1px solid #e2e8f0;
  color: #4a5568;
}

.btn-filter.active {
  background: #667eea;
  color: white;
  border-color: #667eea;
}

.action-buttons {
  display: flex;
  gap: 12px;
  margin-left: auto;
}

.view-toggle {
  display: flex;
  gap: 4px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  overflow: hidden;
}

.btn-view {
  padding: 8px 12px;
  background: white;
  border: none;
  font-size: 16px;
}

.btn-view.active {
  background: #667eea;
}

/* 自动刷新开关 */
.auto-refresh-toggle {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  color: #4a5568;
  cursor: pointer;
  transition: all 0.2s;
}

.auto-refresh-toggle:hover {
  border-color: #667eea;
}

.auto-refresh-toggle input[type="checkbox"] {
  width: 16px;
  height: 16px;
  cursor: pointer;
  accent-color: #667eea;
}

.refresh-indicator {
  color: #34d399;
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

.btn-secondary {
  background-color: #e2e8f0;
  color: #4a5568;
}

.btn-secondary:hover:not(:disabled) {
  background-color: #cbd5e0;
}

.btn-secondary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.error-box {
  padding: 12px;
  background: #fee;
  color: #c33;
  border-radius: 8px;
  margin-bottom: 16px;
}

.metrics-warning {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  background: linear-gradient(135deg, #fff3cd 0%, #ffeeba 100%);
  border: 1px solid #ffc107;
  border-radius: 8px;
  margin-bottom: 16px;
  font-size: 14px;
}

.metrics-warning .warning-icon {
  font-size: 18px;
}

.metrics-warning .warning-text {
  color: #856404;
}

.metrics-warning a {
  color: #0056b3;
  text-decoration: underline;
}

/* 表格视图 */
.table-container {
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.05);
  overflow: hidden;
}

.resource-table {
  width: 100%;
  border-collapse: collapse;
  table-layout: auto;
}

.resource-table th {
  background: #f7fafc;
  text-align: left;
  padding: 16px 20px;
  font-size: 13px;
  font-weight: 600;
  color: #4a5568;
  border-bottom: 1px solid #e2e8f0;
  white-space: nowrap;
}

.resource-table td {
  padding: 16px 20px;
  font-size: 14px;
  color: #2d3748;
  border-bottom: 1px solid #f7fafc;
}

.resource-table tbody tr:hover {
  background: #f7fafc;
}

.status-indicator {
  display: inline-flex;
  align-items: center;
  padding: 6px 12px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
}

.status-indicator.ready {
  background: rgba(52, 211, 153, 0.1);
  color: #34d399;
}

.status-indicator.notready {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.status-indicator.unknown {
  background: rgba(160, 174, 192, 0.1);
  color: #a0aec0;
}

/* 表格状态列布局 */
.status-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
  align-items: flex-start;
}

.unschedulable-badge-small {
  padding: 2px 6px;
  border-radius: 8px;
  font-size: 10px;
  font-weight: 600;
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.node-name {
  display: flex;
  align-items: center;
  gap: 8px;
}

.node-name .icon {
  font-size: 18px;
}

.resource-usage {
  display: flex;
  align-items: center;
  gap: 12px;
}

.usage-bar {
  width: 80px;
  height: 8px;
  background: #f7fafc;
  border-radius: 4px;
  overflow: hidden;
}

.usage-fill {
  height: 100%;
  border-radius: 4px;
  transition: width 0.3s;
}

.usage-low {
  background: #34d399;
}

.usage-medium {
  background: #f59e0b;
}

.usage-high {
  background: #ef4444;
}

.usage-text {
  font-size: 12px;
  font-weight: 500;
  color: #718096;
  min-width: 40px;
}

.mono {
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 13px;
}

.action-icons {
  display: flex;
  gap: 8px;
  align-items: center;
}

.icon-btn {
  padding: 6px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-size: 11px;
  cursor: pointer;
  background: white;
  color: #4a5568;
  transition: all 0.2s;
  white-space: nowrap;
  font-weight: 500;
}

.icon-btn:hover {
  background: #667eea;
  color: white;
  border-color: #667eea;
}

/* 卡片视图 */
.cards-container {
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.05);
}

.cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: 20px;
  margin-bottom: 24px;
}

.node-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  overflow: hidden;
  transition: all 0.3s;
}

.node-card:hover {
  box-shadow: 0 8px 16px rgba(0,0,0,0.1);
  transform: translateY(-2px);
}

.card-header {
  padding: 16px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.card-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.card-icon {
  font-size: 20px;
}

.card-title {
  flex: 1;
  font-size: 16px;
  font-weight: 600;
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-badges {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}

.role-badge {
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 600;
  background: rgba(255,255,255,0.2);
}

.unschedulable-badge {
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 600;
  background: rgba(239, 68, 68, 0.8);
  color: white;
}

.card-body {
  padding: 16px;
}

.card-section {
  margin-bottom: 14px;
}

.card-section:last-child {
  margin-bottom: 0;
}

.section-label {
  font-size: 11px;
  color: #718096;
  text-transform: uppercase;
  font-weight: 600;
  margin-bottom: 6px;
}

.meta-value {
  font-size: 14px;
  color: #2d3748;
  font-weight: 500;
}

.pod-count {
  font-size: 16px;
  font-weight: 700;
  color: #667eea;
}

.resource-bars {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.resource-bar-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.bar-label {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #4a5568;
}

.bar-track {
  height: 6px;
  background: #f7fafc;
  border-radius: 3px;
  overflow: hidden;
}

.bar-fill {
  height: 100%;
  border-radius: 3px;
  transition: width 0.3s;
}

.card-footer {
  display: flex;
  gap: 6px;
  padding: 10px 12px;
  background: #f7fafc;
  border-top: 1px solid #e2e8f0;
  flex-wrap: wrap;
}

.card-action-btn {
  flex: 0 1 auto;
  padding: 6px 10px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-size: 11px;
  font-weight: 500;
  cursor: pointer;
  background: white;
  color: #4a5568;
  transition: all 0.2s;
  white-space: nowrap;
  min-width: 0;
}

.card-action-btn:hover {
  background: #667eea;
  color: white;
  border-color: #667eea;
}

.card-action-btn.danger:hover {
  background: #ef4444;
  border-color: #ef4444;
}

/* 更多操作下拉菜单 */
.more-actions-wrapper {
  position: relative;
}

.more-actions-dropdown {
  position: absolute;
  bottom: 100%;
  right: 0;
  margin-bottom: 8px;
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  min-width: 150px;
  z-index: 100;
  overflow: hidden;
}

.dropdown-item {
  display: block;
  width: 100%;
  padding: 10px 16px;
  border: none;
  background: white;
  color: #4a5568;
  font-size: 12px;
  font-weight: 500;
  text-align: left;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.dropdown-item:hover {
  background: #f7fafc;
  color: #667eea;
}

.dropdown-item.danger {
  color: #ef4444;
}

.dropdown-item.danger:hover {
  background: #fef2f2;
  color: #dc2626;
}

/* 表格视图下拉菜单定位调整 */
.table-dropdown {
  bottom: auto;
  top: 100%;
  margin-top: 8px;
  margin-bottom: 0;
}

/* 空状态 */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: #a0aec0;
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.empty-text {
  font-size: 16px;
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
  z-index: 1000;
}

.modal-content {
  background: white;
  border-radius: 12px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.1);
  width: 90%;
  max-width: 600px;
  max-height: 80vh;
  overflow-y: auto;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #e2e8f0;
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.close-btn {
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
  color: #a0aec0;
  padding: 0;
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
}

.close-btn:hover {
  background: #f7fafc;
  color: #4a5568;
}

.modal-body {
  padding: 24px;
}

.detail-json {
  background: #f7fafc;
  padding: 16px;
  border-radius: 8px;
  overflow-x: auto;
  font-size: 12px;
  font-family: 'Monaco', 'Menlo', monospace;
}

.loading-spinner {
  font-size: 14px;
  color: #667eea;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 24px;
  border-top: 1px solid #e2e8f0;
}

.btn-danger {
  background: #ef4444;
  color: white;
}

.btn-danger:hover:not(:disabled) {
  background: #dc2626;
}

.btn-danger:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Pod 列表弹窗 */
.modal-pods {
  max-width: 800px;
}

.node-info-bar {
  display: flex;
  gap: 24px;
  padding: 16px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 8px;
  margin-bottom: 20px;
  color: white;
}

.node-info-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.info-label {
  font-size: 11px;
  opacity: 0.8;
  text-transform: uppercase;
}

.info-value {
  font-size: 15px;
  font-weight: 600;
}

.info-value.pod-count {
  font-size: 20px;
  color: #fbbf24;
}

.pods-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  max-height: 400px;
  overflow-y: auto;
}

.pod-item {
  padding: 12px 16px;
  background: #f8fafc;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
  transition: all 0.2s;
}

.pod-item:hover {
  border-color: #667eea;
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.15);
}

.pod-item-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.pod-icon {
  font-size: 16px;
}

.pod-name {
  flex: 1;
  font-weight: 600;
  color: #2d3748;
  font-size: 14px;
}

.pod-status {
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 600;
}

.pod-status.running {
  background: rgba(52, 211, 153, 0.1);
  color: #34d399;
}

.pod-status.pending {
  background: rgba(245, 158, 11, 0.1);
  color: #f59e0b;
}

.pod-status.succeeded, .pod-status.completed {
  background: rgba(59, 130, 246, 0.1);
  color: #3b82f6;
}

.pod-status.failed {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.pod-item-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
}

.meta-label {
  color: #718096;
}

.meta-value {
  color: #2d3748;
}

.namespace-badge {
  padding: 2px 8px;
  background: rgba(102, 126, 234, 0.1);
  color: #667eea;
  border-radius: 4px;
  font-weight: 500;
}

.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px;
}

/* 事件弹窗样式 */
.modal-events {
  max-width: 800px;
}

.events-list {
  max-height: 500px;
  overflow-y: auto;
}

.event-item {
  padding: 12px;
  border-radius: 8px;
  margin-bottom: 12px;
  border-left: 4px solid #e2e8f0;
}

.event-item.normal {
  background: #f0fff4;
  border-left-color: #48bb78;
}

.event-item.warning {
  background: #fffaf0;
  border-left-color: #ed8936;
}

.event-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.event-type {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.event-type.normal {
  background: #c6f6d5;
  color: #276749;
}

.event-type.warning {
  background: #feebc8;
  color: #c05621;
}

.event-reason {
  font-weight: 600;
  color: #2d3748;
}

.event-count {
  color: #718096;
  font-size: 12px;
}

.event-message {
  color: #4a5568;
  font-size: 14px;
  line-height: 1.5;
  margin-bottom: 8px;
}

.event-meta {
  display: flex;
  gap: 16px;
  font-size: 12px;
  color: #718096;
}

/* 标签弹窗样式 */
.modal-labels {
  max-width: 700px;
}

.labels-section, .add-label-section {
  margin-bottom: 20px;
}

.section-title {
  font-weight: 600;
  color: #2d3748;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid #e2e8f0;
}

.labels-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.label-item {
  display: inline-flex;
  align-items: center;
  background: #edf2f7;
  border-radius: 4px;
  padding: 4px 8px;
  font-size: 13px;
}

.label-key {
  color: #667eea;
  font-weight: 500;
}

.label-value {
  color: #4a5568;
  margin-left: 4px;
}

.label-value::before {
  content: "=";
  color: #a0aec0;
  margin-right: 4px;
}

.label-remove-btn {
  margin-left: 8px;
  background: none;
  border: none;
  color: #e53e3e;
  cursor: pointer;
  font-size: 16px;
  padding: 0 4px;
}

.label-remove-btn:disabled {
  color: #cbd5e0;
  cursor: not-allowed;
}

.empty-labels, .empty-taints {
  color: #a0aec0;
  font-style: italic;
  padding: 12px;
  text-align: center;
}

.add-label-form, .add-taint-form {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.label-input, .taint-input {
  padding: 8px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-size: 14px;
  flex: 1;
  min-width: 120px;
}

.taint-select {
  padding: 8px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-size: 14px;
  background: white;
}

.btn-sm {
  padding: 8px 16px;
  font-size: 13px;
}

.btn-primary {
  background: #667eea;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #5a67d8;
}

/* 污点弹窗样式 */
.modal-taints {
  max-width: 700px;
}

.taints-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.taint-item {
  display: flex;
  align-items: center;
  gap: 8px;
  background: #fef3c7;
  border-radius: 6px;
  padding: 8px 12px;
  font-size: 13px;
}

.taint-key {
  color: #92400e;
  font-weight: 600;
}

.taint-operator {
  color: #a0aec0;
}

.taint-value {
  color: #78350f;
}

.taint-effect {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  margin-left: auto;
}

.taint-effect.noschedule {
  background: #fed7d7;
  color: #c53030;
}

.taint-effect.prefernoschedule {
  background: #feebc8;
  color: #c05621;
}

.taint-effect.noexecute {
  background: #feb2b2;
  color: #9b2c2c;
}

.taint-remove-btn {
  background: none;
  border: none;
  color: #e53e3e;
  cursor: pointer;
  font-size: 16px;
  padding: 0 4px;
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

.affected-nodes {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.affected-node-tag {
  background: #e2e8f0;
  color: #4a5568;
  padding: 4px 12px;
  border-radius: 16px;
  font-size: 12px;
}

.batch-label-form, .batch-taint-form {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 12px;
}

.pending-labels, .pending-taints {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.pending-label-item, .pending-taint-item {
  display: flex;
  align-items: center;
  gap: 8px;
  background: #f0fff4;
  border: 1px solid #9ae6b4;
  padding: 8px 12px;
  border-radius: 6px;
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
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 13px;
}

.change-item.add {
  color: #38a169;
}

.change-icon {
  width: 20px;
  font-weight: 700;
}

/* 高危操作预览 */
.modal-danger .danger-header {
  background: linear-gradient(135deg, #fc8181 0%, #f56565 100%);
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

.affected-nodes-detail {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 200px;
  overflow-y: auto;
}

.affected-node-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #f7fafc;
  padding: 12px 16px;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
}

.node-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.node-info .node-name {
  font-weight: 600;
  color: #2d3748;
}

.node-info .node-ip {
  font-size: 12px;
  color: #718096;
}

.node-stats {
  display: flex;
  gap: 12px;
  align-items: center;
}

.stat-item {
  font-size: 12px;
  color: #4a5568;
}

.already-cordoned {
  background: #feebc8;
  color: #c05621;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
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

.node-card {
  position: relative;
}

.card-selected {
  border: 2px solid #667eea !important;
  background: rgba(102, 126, 234, 0.05) !important;
}

@media (max-width: 768px) {
  .cards-grid {
    grid-template-columns: 1fr;
  }

  .action-bar {
    flex-direction: column;
    align-items: stretch;
  }

  .search-box input {
    width: 100%;
  }

  .action-buttons {
    width: 100%;
    justify-content: stretch;
    margin-left: 0;
  }
}

/* ========== YAML 模态框样式 ========== */
.yaml-modal {
  width: 90%;
  max-width: 1000px;
  max-height: 90vh;
}

.yaml-header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.yaml-modal-body {
  padding: 0;
  max-height: 60vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.yaml-editor-wrapper {
  flex: 1;
  overflow: auto;
  min-height: 400px;
  max-height: 60vh;
}

.yaml-editor {
  width: 100%;
  height: 100%;
  min-height: 400px;
  padding: 16px;
  border: none;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
  line-height: 1.5;
  resize: none;
  background: #1e1e1e;
  color: #d4d4d4;
}

.yaml-editor:focus {
  outline: none;
}

.yaml-content {
  margin: 0;
  padding: 16px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-all;
  background: #f7f8fa;
  color: #2d3748;
  overflow: auto;
  min-height: 400px;
}

.btn-sm {
  padding: 6px 12px;
  font-size: 13px;
}
</style>
