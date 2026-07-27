<template>
  <div class="resource-view">
    <div class="view-header">
      <h1>DaemonSet 管理</h1>
      <p>Kubernetes 集群中的 DaemonSet 列表</p>
    </div>
    <div class="action-bar">
      <div class="search-box">
        <input type="text" v-model="searchQuery" placeholder="搜索 DaemonSet 名称..." @input="onSearchInput" />
      </div>

      <div class="filter-buttons">
        <button class="btn btn-filter" :class="{ active: statusFilter === 'all' }" @click="setStatusFilter('all')">全部</button>
        <button class="btn btn-filter" :class="{ active: statusFilter === 'Running' }" @click="setStatusFilter('Running')">Running</button>
        <button class="btn btn-filter" :class="{ active: statusFilter === 'Updating' }" @click="setStatusFilter('Updating')">Updating</button>
        <button class="btn btn-filter" :class="{ active: statusFilter === 'Failed' }" @click="setStatusFilter('Failed')">Failed</button>
      </div>

      <div class="filter-dropdown">
        <select v-model="namespaceFilter">
          <option value="">所有命名空间</option>
          <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
        </select>
      </div>

      <div class="action-buttons">
        <button v-if="canOperate && !batchMode" class="btn btn-batch" @click="enterBatchMode" title="进入批量操作模式">☑️ 批量操作</button>
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
        <button v-if="canOperate" class="btn btn-primary" @click="showCreateModal = true">创建 DaemonSet</button>
        <button class="btn btn-secondary" @click="refreshList" :disabled="loading">{{ loading ? '加载中...' : '🔄 刷新' }}</button>
      </div>
    </div>

    <div v-if="errorMsg" class="error-box">{{ errorMsg }}</div>

    <!-- 批量操作浮动栏 -->
    <div v-if="batchMode && selectedDaemonsets.length > 0" class="batch-action-bar">
      <div class="batch-info">
        <span class="batch-count">已选择 {{ selectedDaemonsets.length }} 个 DaemonSet</span>
        <button class="batch-clear" @click="clearSelection">清空选择</button>
      </div>
      <div class="batch-actions">
        <button class="batch-btn" @click="batchRestart" title="批量重启">🔄 批量重启</button>
        <button class="batch-btn danger" @click="batchDelete" title="批量删除">🗑️ 批量删除</button>
      </div>
    </div>
    
    <!-- 表格视图 -->
    <div v-if="viewMode === 'table'" class="table-container">
      <table class="resource-table">
        <thead>
          <tr>
            <th v-if="batchMode" style="width: 40px;">
              <input type="checkbox" :checked="isAllSelected" @change="toggleSelectAll" title="全选/取消全选" />
            </th>
            <th style="width: 100px;">状态</th>
            <th style="min-width: 180px;">名称</th>
            <th style="width: 130px;">命名空间</th>
            <th style="width: 150px;">节点数</th>
            <th style="min-width: 200px;">镜像</th>
            <th style="min-width: 180px;">选择器</th>
            <th style="width: 120px;">更新策略</th>
            <th style="width: 170px;">创建时间</th>
            <th style="width: 100px;">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="ds in paginatedDaemonsets" :key="ds.name" :class="{ 'row-selected': isDaemonsetSelected(ds) }">
            <td v-if="batchMode">
              <input type="checkbox" :checked="isDaemonsetSelected(ds)" @change="toggleDaemonsetSelection(ds)" />
            </td>
            <td>
              <span class="status-indicator" :class="(ds.status || 'unknown').toLowerCase()">{{ ds.status || 'Unknown' }}</span>
            </td>
            <td>
              <div class="daemonset-name">
                <span class="icon">🔄</span>
                <span>{{ ds.name }}</span>
              </div>
            </td>
            <td><span class="namespace-badge">{{ ds.namespace }}</span></td>
            <td>
              <div class="replicas-info">
                <div class="replicas-display">
                  <span class="ready-replicas">{{ ds.numberReady || 0 }}</span>
                  <span class="replicas-sep">/</span>
                  <span class="desired-replicas">{{ ds.desiredNumberScheduled || 0 }}</span>
                </div>
                <div class="replicas-bar">
                  <div class="replicas-fill" :style="{ width: `${(ds.numberReady / Math.max(ds.desiredNumberScheduled, 1)) * 100}%` }"></div>
                </div>
              </div>
            </td>
            <td>
              <div class="image-text" :class="{ clickable: canOperate }" @click="canOperate && startInlineImage(ds)" :title="canOperate ? '点击修改镜像' : ds.image">
                <span class="image-name">{{ ds.image || '-' }}</span>
                <span v-if="canOperate" class="edit-icon">✏️</span>
              </div>
            </td>
            <td>
              <div class="selector-tags">
                <span v-for="(value, key) in ds.selector" :key="key" class="selector-tag">{{ key }}={{ value }}</span>
              </div>
            </td>
            <td><span class="strategy-badge">{{ ds.updateStrategy }}</span></td>
            <td style="white-space: nowrap;">{{ ds.createdAt }}</td>
            <td>
              <div class="action-icons">
                <button class="action-btn" @click="viewDaemonset(ds)" title="DaemonSet 详情 (describe)">📋 详情</button>
                <button class="action-btn terminal" @click="openTerminalForDaemonSet(ds)" title="打开容器终端">>_ 终端</button>
                <button class="action-btn primary" @click="viewPods(ds)" title="查看 Pod">📦 Pod</button>
                <div class="more-btn">
                  <button class="icon-btn" @click="toggleMoreOptions(ds, $event)">⋮ 更多</button>
                  <div v-if="showMoreOptions && selectedDaemonset === ds" class="more-menu" :style="menuStyle">
                    <button class="menu-item" @click="viewDaemonsetLogs(ds)"><span class="menu-icon">📄</span><span>查看日志</span></button>
                    <button class="menu-item" @click="viewHistory(ds)"><span class="menu-icon">📜</span><span>版本记录</span></button>
                    <div v-if="canOperate" class="menu-divider"></div>
                    <button v-if="canOperate" class="menu-item" @click="restartDaemonset(ds)"><span class="menu-icon">🔄</span><span>重启</span></button>
                    <button v-if="canOperate" class="menu-item" @click="openUpdateImage(ds)"><span class="menu-icon">🔧</span><span>更新镜像</span></button>
                    <button v-if="canOperate" class="menu-item" @click="openResourceModal(ds)"><span class="menu-icon">⚡</span><span>资源配置</span></button>
                    <button v-if="canOperate" class="menu-item" @click="openRollback(ds)"><span class="menu-icon">⏪</span><span>回滚</span></button>
                    <div class="menu-divider"></div>
                    <button class="menu-item" @click="openEvents(ds)"><span class="menu-icon">📡</span><span>查看事件</span></button>
                    <button class="menu-item" @click="openYamlPreview(ds)"><span class="menu-icon">📝</span><span>查看/编辑 YAML</span></button>
                    <div v-if="canOperate" class="menu-divider"></div>
                    <button v-if="canOperate" class="menu-item danger" @click="deleteDaemonset(ds)"><span class="menu-icon">🗑️</span><span>删除</span></button>
                  </div>
                </div>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-if="filteredDaemonsets.length === 0" class="empty-state">
        <div class="empty-icon">🔍</div>
        <div class="empty-text">没有找到匹配的 DaemonSet</div>
      </div>
      <Pagination v-if="filteredDaemonsets.length > 0" v-model:currentPage="currentPage" :totalItems="filteredDaemonsets.length" :itemsPerPage="itemsPerPage" />
    </div>

    <!-- 卡片视图 -->
    <div v-if="viewMode === 'card'" class="cards-container">
      <div v-if="filteredDaemonsets.length === 0" class="empty-state">
        <div class="empty-icon">🔍</div>
        <div class="empty-text">没有找到匹配的 DaemonSet</div>
      </div>
      <div class="cards-grid">
        <div v-for="ds in paginatedDaemonsets" :key="ds.name" class="daemonset-card" :class="{ 'card-selected': isDaemonsetSelected(ds) }">
          <div v-if="batchMode" class="card-checkbox">
            <input type="checkbox" :checked="isDaemonsetSelected(ds)" @change="toggleDaemonsetSelection(ds)" />
          </div>
          <div class="card-header">
            <div class="card-title-row">
              <span class="card-icon">🔄</span>
              <h3 class="card-title">{{ ds.name }}</h3>
              <span class="status-indicator" :class="(ds.status || 'unknown').toLowerCase()">{{ ds.status || 'Unknown' }}</span>
            </div>
            <span class="namespace-badge">{{ ds.namespace }}</span>
          </div>
          <div class="card-body">
            <div class="card-section">
              <div class="section-label">节点数</div>
              <div class="replicas-info">
                <div class="replicas-display">
                  <span class="ready-replicas">{{ ds.numberReady || 0 }}</span>
                  <span class="replicas-sep">/</span>
                  <span class="desired-replicas">{{ ds.desiredNumberScheduled || 0 }}</span>
                </div>
                <div class="replicas-bar">
                  <div class="replicas-fill" :style="{ width: `${(ds.numberReady / Math.max(ds.desiredNumberScheduled, 1)) * 100}%` }"></div>
                </div>
              </div>
            </div>
            <div class="card-section">
              <div class="section-label">镜像</div>
              <div class="image-text" :class="{ clickable: canOperate }" @click="canOperate && startInlineImage(ds)" :title="canOperate ? '点击修改镜像' : ds.image">
                <span class="image-name">{{ ds.image || '-' }}</span>
                <span v-if="canOperate" class="edit-icon">✏️</span>
              </div>
            </div>
            <div class="card-section">
              <div class="section-label">选择器</div>
              <div class="selector-tags">
                <span v-for="(value, key) in ds.selector" :key="key" class="selector-tag">{{ key }}={{ value }}</span>
              </div>
            </div>
            <div class="card-section card-section-row">
              <div class="card-meta-item">
                <div class="meta-label">更新策略</div>
                <span class="strategy-badge">{{ ds.updateStrategy }}</span>
              </div>
              <div class="card-meta-item">
                <div class="meta-label">创建时间</div>
                <div class="meta-value">{{ ds.createdAt }}</div>
              </div>
            </div>
          </div>
          <div class="card-footer">
            <button class="card-action-btn primary" @click="viewPods(ds)" title="查看 Pod">📦 Pod</button>
            <button class="card-action-btn" @click="viewDaemonsetLogs(ds)" title="查看日志">📄 日志</button>
            <button class="card-action-btn" @click="viewHistory(ds)" title="版本记录">📜 版本</button>
            <button class="card-action-btn warning" @click="openResourceModal(ds)" title="快速调整 CPU/Memory（防 OOM）">⚡ 资源</button>
            <button class="card-action-btn" @click="openYamlPreview(ds)" title="查看/编辑 YAML">📝 YAML</button>
            <button v-if="canOperate" class="card-action-btn danger" @click="deleteDaemonset(ds)" title="删除">🗑️ 删除</button>
          </div>
        </div>
      </div>
      <Pagination v-if="filteredDaemonsets.length > 0" v-model:currentPage="currentPage" :totalItems="filteredDaemonsets.length" :itemsPerPage="itemsPerPage" />
    </div>

    <!-- 创建 DaemonSet 模态框 -->
    <div v-if="showCreateModal" class="modal-overlay" @click.self="showCreateModal = false">
      <div class="modal-content modal-create-daemonset">
        <div class="modal-header">
          <h2>🔄 创建 DaemonSet</h2>
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
          <form @submit.prevent="createDaemonset">
            <div class="form-section">
              <div class="section-header"><span class="section-icon">📋</span><h3>基本信息</h3></div>
              <div class="section-body">
                <div class="form-group">
                  <label for="dsName">名称 <span class="required">*</span></label>
                  <input type="text" id="dsName" v-model="daemonsetForm.name" class="form-input" required placeholder="输入 DaemonSet 名称" />
                </div>
                <div class="form-group">
                  <label for="dsNamespace">命名空间 <span class="required">*</span></label>
                  <div class="namespace-selector">
                    <select v-if="!showNamespaceInput" id="dsNamespace" v-model="daemonsetForm.namespace" class="form-select" required style="flex: 1;">
                      <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
                    </select>
                    <span v-if="!showNamespaceInput" class="namespace-or">或</span>
                    <div v-if="showNamespaceInput" class="namespace-create">
                      <input v-model="newNamespace" type="text" class="form-input" placeholder="新命名空间名称" :disabled="creatingNamespace" />
                      <button type="button" class="btn btn-secondary btn-sm" @click="createNewNamespace" :disabled="!newNamespace || creatingNamespace">
                        {{ creatingNamespace ? '创建中...' : '创建' }}
                      </button>
                      <button type="button" class="btn btn-secondary btn-sm" @click="cancelCreateNamespace">取消</button>
                    </div>
                    <button v-if="!showNamespaceInput" type="button" class="btn btn-secondary btn-sm" @click="showNamespaceInput = true">➕ 创建新命名空间</button>
                  </div>
                </div>
              </div>
            </div>
            <div class="form-section">
              <div class="section-header"><span class="section-icon">🐳</span><h3>容器镜像</h3></div>
              <div class="section-body">
                <div class="form-group">
                  <label for="dsImage">镜像地址 <span class="required">*</span></label>
                  <input type="text" id="dsImage" v-model="daemonsetForm.container_image" class="form-input" required placeholder="例如: nginx:latest" />
                </div>
              </div>
            </div>
            <div class="form-section">
              <div class="section-header"><span class="section-icon">🏷️</span><h3>标签</h3></div>
              <div class="section-body">
                <div class="form-group">
                  <label>Pod 标签 (Labels) <span class="required">*</span></label>
                  <div class="labels-editor">
                    <div v-for="(label, index) in daemonsetForm.labels" :key="index" class="label-row">
                      <input v-model="label.key" class="label-input label-key" placeholder="键，例如: app" />
                      <span class="label-separator">=</span>
                      <input v-model="label.value" class="label-input label-value" placeholder="值，例如: nginx" />
                      <button type="button" class="btn-icon btn-remove" @click="removeLabel(index)" :disabled="daemonsetForm.labels.length === 1">🗑️</button>
                    </div>
                    <button type="button" class="btn btn-secondary btn-sm" @click="addLabel">➕ 添加标签</button>
                  </div>
                </div>
              </div>
            </div>
            <div class="form-section">
              <div class="section-header"><span class="section-icon">🎯</span><h3>节点选择器</h3></div>
              <div class="section-body">
                <div class="form-hint">指定 DaemonSet 应该在哪些节点上运行 Pod</div>
                <div class="form-group">
                  <div class="labels-editor">
                    <div v-for="(ns, index) in daemonsetForm.nodeSelector" :key="index" class="label-row">
                      <input v-model="ns.key" class="label-input label-key" placeholder="键，例如: node-role" />
                      <span class="label-separator">=</span>
                      <input v-model="ns.value" class="label-input label-value" placeholder="值，例如: worker" />
                      <button type="button" class="btn-icon btn-remove" @click="removeNodeSelector(index)">🗑️</button>
                    </div>
                    <button type="button" class="btn btn-secondary btn-sm" @click="addNodeSelector">➕ 添加节点选择器</button>
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
              <p class="yaml-hint">✨ 支持多资源 YAML 创建（用 <code>---</code> 分隔），可同时创建 ConfigMap、Service 等依赖资源</p>
              <div class="yaml-header-buttons">
                <button class="load-template-btn" @click="loadDaemonSetYamlTemplate">
                  📑 加载模板（DaemonSet）
                </button>
                <button class="copy-yaml-btn" @click="copyYamlContent">
                  📋 复制
                </button>
                <button class="reset-yaml-btn" @click="resetYamlContent">
                  🔄 重置
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
                  <li>支持完整的 Kubernetes DaemonSet 配置</li>
                  <li>可以通过“加载模板”获取示例 YAML</li>
                  <li>创建前会验证 YAML 格式的正确性</li>
                  <li>DaemonSet 会在每个符合条件的节点上运行一个 Pod</li>
                </ul>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <!-- 表单模式按钮 -->
          <template v-if="createMode === 'form'">
            <button type="button" class="btn btn-secondary" @click="showCreateModal = false">取消</button>
            <button type="button" class="btn btn-primary" @click="createDaemonset" :disabled="creating">{{ creating ? '创建中...' : '创建' }}</button>
          </template>
          
          <!-- YAML 模式按钮 -->
          <template v-if="createMode === 'yaml'">
            <button type="button" class="btn btn-secondary" @click="showCreateModal = false">取消</button>
            <button 
              type="button"
              class="btn btn-primary" 
              @click="createDaemonSetFromYaml"
              :disabled="!yamlContent"
            >
              <span class="btn-icon">🔄</span>从 YAML 创建
            </button>
          </template>
        </div>
      </div>
    </div>

    <!-- Describe 抽屉 (DDP Style) -->
    <transition name="ddp-slide">
      <div v-if="showViewModal" class="ddp-overlay" @click.self="showViewModal = false">
        <div class="ddp-panel">
          <div class="ddp-hd">
            <div class="ddp-hd-body">
              <div class="ddp-hd-row1">
                <span class="ddp-status-pill" :class="ddpStatusClass(detailData)"><span class="ddp-status-led"></span>{{ ddpStatusText(detailData) }}</span>
                <span class="ddp-ns-pill">{{ detailData?.metadata?.namespace || '-' }}</span>
              </div>
              <h2 class="ddp-name">{{ detailData?.metadata?.name || '-' }}</h2>
              <div class="ddp-hd-meta" v-if="detailData?.status">
                <span>就绪 {{ detailData.status?.numberReady || 0 }}/{{ detailData.status?.desiredNumberScheduled || 0 }}</span>
                <span class="ddp-dot">·</span>
                <span>策略 {{ detailData.spec?.updateStrategy?.type || '-' }}</span>
                <span class="ddp-dot">·</span>
                <span>{{ fmtTime(detailData.metadata?.creationTimestamp) }}</span>
              </div>
            </div>
            <button class="ddp-close-btn" @click="showViewModal = false"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M18 6L6 18M6 6l12 12"/></svg></button>
          </div>
          <div v-if="loadingDetail" class="ddp-loading"><div class="ddp-spin"></div><span>正在获取 DaemonSet 详情...</span></div>
          <div v-else-if="detailData" class="ddp-body">
            <div class="ddp-sec">
              <div class="ddp-sec-title"><span class="ddp-si ddp-si-blue">⚡</span>基本信息</div>
              <div class="ddp-kv-grid">
                <div class="ddp-kv"><span class="ddp-k">名称</span><span class="ddp-v mono">{{ detailData.metadata?.name }}</span></div>
                <div class="ddp-kv"><span class="ddp-k">命名空间</span><span class="ddp-v">{{ detailData.metadata?.namespace }}</span></div>
                <div class="ddp-kv"><span class="ddp-k">创建时间</span><span class="ddp-v">{{ fmtTime(detailData.metadata?.creationTimestamp) }}</span></div>
                <div class="ddp-kv"><span class="ddp-k">UID</span><span class="ddp-v mono" style="font-size:11px">{{ detailData.metadata?.uid || '-' }}</span></div>
                <div class="ddp-kv"><span class="ddp-k">节点数</span><span class="ddp-v">期望 {{ detailData.status?.desiredNumberScheduled || 0 }} | 就绪 {{ detailData.status?.numberReady || 0 }} | 可用 {{ detailData.status?.numberAvailable || 0 }} | 已更新 {{ detailData.status?.updatedNumberScheduled || 0 }}</span></div>
                <div class="ddp-kv"><span class="ddp-k">更新策略</span><span class="ddp-v">{{ detailData.spec?.updateStrategy?.type || '-' }}<span v-if="detailData.spec?.updateStrategy?.rollingUpdate"> (maxUnavailable: {{ detailData.spec.updateStrategy.rollingUpdate.maxUnavailable || 1 }})</span></span></div>
                <div class="ddp-kv"><span class="ddp-k">历史版本数</span><span class="ddp-v">{{ detailData.spec?.revisionHistoryLimit ?? 10 }}</span></div>
              </div>
            </div>
            <div class="ddp-sec" v-if="(detailData.status?.conditions || []).length">
              <div class="ddp-sec-title"><span class="ddp-si ddp-si-green">✅</span>状态条件 Conditions</div>
              <div class="ddp-conds">
                <div v-for="c in detailData.status.conditions" :key="c.type" class="ddp-cond-chip" :class="c.status === 'True' ? 'cond-ok' : 'cond-fail'">
                  <span class="cond-led"></span><span class="cond-name">{{ c.type }}</span><span class="cond-reason" v-if="c.reason">{{ c.reason }}</span>
                </div>
              </div>
            </div>
            <div class="ddp-sec" v-if="detailData.spec?.selector?.matchLabels">
              <div class="ddp-sec-title"><span class="ddp-si ddp-si-purple">🎯</span>选择器 Selector</div>
              <div class="ddp-tags ddp-tags-pad">
                <span v-for="(v, k) in detailData.spec.selector.matchLabels" :key="k" class="ddp-label-tag"><b>{{ k }}</b>=<span>{{ v }}</span></span>
              </div>
            </div>
            <div class="ddp-sec" v-if="Object.keys(detailData.metadata?.labels || {}).length">
              <div class="ddp-sec-title"><span class="ddp-si ddp-si-teal">🏷</span>标签 Labels</div>
              <div class="ddp-tags ddp-tags-pad">
                <span v-for="(v, k) in detailData.metadata.labels" :key="k" class="ddp-label-tag"><b>{{ k }}</b>=<span>{{ v }}</span></span>
              </div>
            </div>
            <div class="ddp-sec" v-if="Object.keys(detailData.metadata?.annotations || {}).length">
              <div class="ddp-sec-title ddp-sec-title-clickable" @click="ddpShowAnno = !ddpShowAnno">
                <span class="ddp-si ddp-si-orange">📝</span>注解 Annotations
                <span class="ddp-cnt-badge">{{ Object.keys(detailData.metadata.annotations).length }}</span>
                <span class="ddp-fold-arrow">{{ ddpShowAnno ? '▲' : '▼' }}</span>
              </div>
              <div v-if="ddpShowAnno" class="ddp-anno-list">
                <div v-for="(v, k) in detailData.metadata.annotations" :key="k" class="ddp-anno-row"><span class="ddp-anno-k">{{ k }}</span><span class="ddp-anno-v">{{ v }}</span></div>
              </div>
            </div>
            <div class="ddp-sec" v-if="(detailData.spec?.template?.spec?.containers || []).length">
              <div class="ddp-sec-title"><span class="ddp-si ddp-si-blue">📦</span>Pod 模板容器<span class="ddp-cnt-badge">{{ detailData.spec.template.spec.containers.length }}</span></div>
              <div class="ddp-ctr-list">
                <div v-for="c in detailData.spec.template.spec.containers" :key="c.name" class="ddp-ctr-card">
                  <div class="ddp-ctr-hd"><span class="ddp-ctr-name">{{ c.name }}</span></div>
                  <div class="ddp-ctr-img">{{ c.image }}</div>
                  <div class="ddp-ctr-detail">
                    <div class="ddp-ctr-row" v-if="(c.command||[]).length||(c.args||[]).length"><span class="ddp-rk">命令</span><div class="pdp-cmd-blk"><div v-if="(c.command||[]).length" class="pdp-cmd-line"><span class="pdp-cmd-label">command</span>{{ c.command.join(' ') }}</div><div v-if="(c.args||[]).length" class="pdp-cmd-line"><span class="pdp-cmd-label">args</span>{{ c.args.join(' ') }}</div></div></div>
                    <div class="ddp-ctr-row" v-if="(c.ports||[]).length"><span class="ddp-rk">端口</span><span v-for="p in c.ports" :key="p.containerPort" class="ddp-port-chip">{{ p.containerPort }}/{{ p.protocol||'TCP' }}</span></div>
                    <div class="ddp-ctr-row" v-if="c.resources?.requests||c.resources?.limits"><span class="ddp-rk">资源</span><div class="pdp-res-lines"><div v-if="c.resources?.requests" class="pdp-res-line"><span class="pdp-res-lbl req">requests</span><span v-for="(v,k) in c.resources.requests" :key="k" class="pdp-res-tag">{{ k }}:{{ v }}</span></div><div v-if="c.resources?.limits" class="pdp-res-line"><span class="pdp-res-lbl lim">limits</span><span v-for="(v,k) in c.resources.limits" :key="k" class="pdp-res-tag">{{ k }}:{{ v }}</span></div></div></div>
                    <div class="ddp-ctr-row" v-if="(c.volumeMounts||[]).length"><span class="ddp-rk">挂载</span><div class="pdp-mnt-tbl"><div v-for="m in c.volumeMounts" :key="m.name" class="pdp-mnt-row"><span class="pdp-mn">{{ m.name }}</span><span class="pdp-mp">{{ m.mountPath }}</span><span v-if="m.readOnly" class="pdp-mro">ro</span></div></div></div>
                  </div>
                </div>
              </div>
            </div>
            <div class="ddp-sec ddp-sec-events">
              <div class="ddp-sec-title"><span class="ddp-si ddp-si-orange">⚡</span>事件 Events<span v-if="ddpEventsData.length" class="ddp-cnt-badge">{{ ddpEventsData.length }}</span><button class="pdp-ev-refresh-btn" @click="ddpRefreshEvents" :disabled="ddpEventsLoading" title="刷新事件"><svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M23 4v6h-6M1 20v-6h6"/><path d="M3.51 9a9 9 0 0114.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0020.49 15"/></svg></button></div>
              <div v-if="ddpEventsLoading" class="pdp-ev-empty"><div class="pdp-ev-spin"></div>加载事件中...</div>
              <div v-else-if="!ddpEventsData.length" class="pdp-ev-empty">暂无事件记录</div>
              <div v-else class="pdp-ev-list">
                <div v-for="(ev, i) in ddpEventsData" :key="i" class="pdp-ev-item" :class="ev.type === 'Warning' ? 'ev-warn' : 'ev-norm'">
                  <div class="pdp-ev-hd"><span class="pdp-ev-type-dot" :class="ev.type === 'Warning' ? 'ev-dot-warn' : 'ev-dot-norm'"></span><span class="pdp-ev-reason">{{ ev.reason }}</span><span class="pdp-ev-count" v-if="ev.count > 1">×{{ ev.count }}</span><span class="pdp-ev-age">{{ ddpFmtEvTime(ev.lastTimestamp || ev.last_timestamp || ev.event_time) }}</span></div>
                  <div class="pdp-ev-msg">{{ ev.message }}</div>
                  <div class="pdp-ev-src" v-if="ev.source?.component || ev.reporting_component">来源: {{ ev.source?.component || ev.reporting_component }}</div>
                </div>
              </div>
            </div>
          </div>
          <div v-else class="ddp-empty"><div style="font-size:48px;margin-bottom:12px">📭</div><p>暂无 DaemonSet 数据</p></div>
        </div>
      </div>
    </transition>

    <!-- 更新镜像弹窗 -->
    <div v-if="showUpdateImageModal" class="modal-overlay" @click.self="showUpdateImageModal = false">
      <div class="modal-content" style="max-width: 520px;">
        <div class="modal-header">
          <h3>🔧 更新镜像</h3>
          <button class="close-btn" @click="showUpdateImageModal = false">×</button>
        </div>
        <div class="modal-body">
          <div class="info-box">
            <div><strong>DaemonSet:</strong> {{ updateImageForm.name }}</div>
            <div><strong>命名空间:</strong> {{ updateImageForm.namespace }}</div>
          </div>
          <div class="form-group" v-if="containerList.length > 1">
            <label>选择容器</label>
            <select v-model="updateImageForm.container" class="form-select" @change="onUpdateImageContainerChange">
              <option value="" disabled>请选择容器</option>
              <option v-for="c in containerList" :key="c" :value="c">{{ c }}</option>
            </select>
          </div>
          <div class="form-group" v-else-if="containerList.length === 1">
            <label>容器</label>
            <div class="form-static">{{ containerList[0] }}</div>
          </div>
          <div class="form-group" v-else>
            <label>容器名称</label>
            <input type="text" v-model="updateImageForm.container" class="form-input" placeholder="容器名称" />
          </div>
          <div class="form-group">
            <label>当前镜像</label>
            <div class="current-image-display" :title="updateImageForm.currentImage">
              <span class="current-image-text">{{ updateImageForm.currentImage || '-' }}</span>
            </div>
          </div>
          <div class="form-group">
            <label>新镜像地址</label>
            <input type="text" v-model="updateImageForm.image" class="form-input" :placeholder="updateImageForm.currentImage || '例如: nginx:1.25'" />
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

    <!-- 快速资源调整弹窗（防 OOM） -->
    <div v-if="showResourceModal" class="modal-overlay" @click.self="showResourceModal = false">
      <div class="modal-content" style="max-width: 560px;">
        <div class="modal-header">
          <h3>⚡ 快速调整资源配置</h3>
          <button class="close-btn" @click="showResourceModal = false">&times;</button>
        </div>
        <div class="modal-body">
          <div class="info-box">
            <div><strong>DaemonSet:</strong> {{ resourceForm.name }}</div>
            <div><strong>命名空间:</strong> {{ resourceForm.namespace }}</div>
          </div>

          <!-- 容器选择 -->
          <div class="form-group" v-if="resourceContainerList.length > 1">
            <label>选择容器</label>
            <select v-model="resourceForm.container" class="form-select" @change="onResourceContainerChange">
              <option value="" disabled>请选择容器</option>
              <option v-for="c in resourceContainerList" :key="c" :value="c">{{ c }}</option>
            </select>
          </div>
          <div class="form-group" v-else-if="resourceContainerList.length === 1">
            <label>容器</label>
            <div class="form-static">{{ resourceContainerList[0] }}</div>
          </div>

          <!-- 当前资源配置展示 -->
          <div v-if="currentResources" class="resource-current-info">
            <div class="section-label" style="margin-bottom: 8px;">📊 当前配置</div>
            <div class="resource-grid-display">
              <div class="resource-item-display">
                <span class="resource-key">CPU Request:</span>
                <span class="resource-val">{{ currentResources.cpu_request || '未设置' }}</span>
              </div>
              <div class="resource-item-display">
                <span class="resource-key">CPU Limit:</span>
                <span class="resource-val">{{ currentResources.cpu_limit || '未设置' }}</span>
              </div>
              <div class="resource-item-display">
                <span class="resource-key">Memory Request:</span>
                <span class="resource-val">{{ currentResources.memory_request || '未设置' }}</span>
              </div>
              <div class="resource-item-display">
                <span class="resource-key">Memory Limit:</span>
                <span class="resource-val" :class="{ 'text-danger': !currentResources.memory_limit }">{{ currentResources.memory_limit || '⚠️ 未设置' }}</span>
              </div>
            </div>
          </div>
          <div v-else class="resource-current-info">
            <div class="form-hint">⚠️ 当前容器未设置资源配置，建议设置以防 OOM</div>
          </div>

          <!-- 资源编辑表单 -->
          <div class="resource-edit-section">
            <div class="section-label" style="margin-bottom: 8px;">✏️ 新配置</div>
            <div class="resource-form-grid">
              <div class="form-group">
                <label>Memory Limit <span class="required">*</span></label>
                <input v-model="resourceForm.memory_limit" type="text" class="form-input" placeholder="例: 512Mi, 1Gi, 2Gi" />
                <div class="form-hint">必填，防止 OOM 必须设置内存上限</div>
              </div>
              <div class="form-group">
                <label>Memory Request</label>
                <input v-model="resourceForm.memory_request" type="text" class="form-input" placeholder="例: 256Mi, 512Mi" />
                <div class="form-hint">建议设置为 Limit 的 50%-80%</div>
              </div>
              <div class="form-group">
                <label>CPU Limit</label>
                <input v-model="resourceForm.cpu_limit" type="text" class="form-input" placeholder="例: 500m, 1000m, 2" />
              </div>
              <div class="form-group">
                <label>CPU Request</label>
                <input v-model="resourceForm.cpu_request" type="text" class="form-input" placeholder="例: 100m, 250m" />
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showResourceModal = false">取消</button>
          <button class="btn btn-primary" @click="submitUpdateResources" :disabled="updatingResources || !resourceForm.memory_limit">
            {{ updatingResources ? '更新中...' : '确认修改' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 事件弹窗 -->
    <div v-if="showEventsModal" class="modal-overlay" @click.self="showEventsModal = false">
      <div class="modal-content">
        <div class="modal-header">
          <h3>📡 DaemonSet 事件</h3>
          <button class="close-btn" @click="showEventsModal = false">×</button>
        </div>
        <div class="modal-body">
          <div v-if="loadingEvents" class="loading-state">加载中...</div>
          <div v-else-if="eventsData.length > 0">
            <div v-for="(event, idx) in eventsData" :key="idx" class="event-item">
              <div class="event-type" :class="(event.type || 'normal').toLowerCase()">{{ event.type || 'Normal' }}</div>
              <div class="event-content">
                <div class="event-reason">{{ event.reason }}</div>
                <div class="event-message">{{ event.message }}</div>
                <div class="event-time">{{ fmtTime(event.event_time || event.lastTimestamp) }}</div>
              </div>
            </div>
          </div>
          <div v-else class="empty-state-small">
            <div class="empty-text">暂无事件</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 版本历史弹窗 -->
    <div v-if="showHistoryModal" class="modal-overlay" @click.self="showHistoryModal = false">
      <div class="modal-content modal-lg">
        <div class="modal-header">
          <h3>📜 版本记录</h3>
          <button class="close-btn" @click="showHistoryModal = false">×</button>
        </div>
        <div class="modal-body">
          <div v-if="historyDaemonset" class="info-box" style="margin-bottom: 16px;">
            <div><strong>DaemonSet:</strong> {{ historyDaemonset.name }}</div>
            <div><strong>命名空间:</strong> {{ historyDaemonset.namespace }}</div>
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
                  <td><span class="version-badge">{{ rev.revision }}</span></td>
                  <td class="mono">{{ rev.name }}</td>
                  <td>{{ fmtTime(rev.creation_time) }}</td>
                  <td>
                    <button class="btn btn-sm btn-primary" @click="rollbackToVersion(rev)" :disabled="rollingBack">回滚到此版本</button>
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
    <RollbackDialog
      :visible="showRollbackModal"
      :loading="loadingRollbackHistory"
      :confirming="rollingBack"
      :resource-name="rollbackForm.name"
      :current-info="dsRollbackCurrentInfo"
      :revisions="rollbackHistoryList"
      :selected-revision="(rollbackHistoryList || []).find(h => h.name === rollbackForm.revision_name)"
      @close="showRollbackModal = false"
      @select="(rev) => rollbackForm.revision_name = rev.name"
      @confirm="submitRollback"
    />

    <!-- Pod 关联弹窗 -->
    <div v-if="showPodsModal" class="modal-overlay" @click.self="showPodsModal = false">
      <div class="modal-content modal-lg">
        <div class="modal-header">
          <h3>📦 关联 Pods</h3>
          <button class="close-btn" @click="showPodsModal = false">×</button>
        </div>
        <div class="modal-body">
          <div class="info-box">
            <div><strong>DaemonSet:</strong> {{ podsDaemonset?.name }}</div>
            <div><strong>命名空间:</strong> {{ podsDaemonset?.namespace }}</div>
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
                    <span class="status-indicator" :class="(pod.status || 'unknown').toLowerCase()">{{ pod.status }}</span>
                  </td>
                  <td>{{ pod.node || '-' }}</td>
                  <td style="text-align: center;">{{ pod.restartCount || 0 }}</td>
                  <td>{{ pod.createdAt || '-' }}</td>
                  <td>
                    <div class="pod-actions">
                      <button class="icon-btn" title="查看日志" @click="openPodLogs(pod)">📄 日志</button>
                      <button class="icon-btn" title="容器终端" @click="openTerminal(pod)">&#62;_ 终端</button>
                      <div class="more-btn">
                        <button class="icon-btn" @click="togglePodMoreOptions(pod, $event)" title="更多操作">⋮ 更多</button>
                        <div v-if="showPodMoreOptions && selectedPodForAction === pod" class="more-menu" :style="podMenuStyle">
                          <button class="menu-item" @click="restartPodFromList(pod)"><span class="menu-icon">🔄</span><span>重启 Pod</span></button>
                          <button class="menu-item" @click="openPodDetail(pod)"><span class="menu-icon">📋</span><span>查看详情</span></button>
                          <button class="menu-item" @click="openPodEvents(pod)"><span class="menu-icon">📡</span><span>查看事件</span></button>
                          <div class="menu-divider"></div>
                          <button class="menu-item danger" @click="deletePodFromList(pod)"><span class="menu-icon">🗑️</span><span>优雅删除</span></button>
                          <button class="menu-item danger" @click="forceDeletePodFromList(pod)"><span class="menu-icon">💥</span><span>强制删除</span></button>
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

    <!-- Pod 日志弹窗（增强版 - 实时日志、高亮） -->
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
              </select>
            </div>
            <div class="control-actions">
              <button 
                class="btn btn-primary btn-sm" 
                @click="fetchPodLogsContent" 
                :disabled="loadingPodLogs || (!podLogsForm.container && podContainerList.length > 1)"
              >
                {{ loadingPodLogs ? '加载中...' : (podLogsForm.follow ? '获取实时日志' : '获取日志') }}
              </button>
              <!-- 终止按钮 -->
              <button 
                v-if="loadingPodLogs || isStreamingPodLogs" 
                class="btn btn-danger btn-sm" 
                @click="stopPodLogLoading"
              >
                终止
              </button>
              <button class="btn btn-secondary btn-sm" @click="clearPodLogs" :disabled="!podLogsContent || loadingPodLogs">清除</button>
            </div>
          </div>
          <div v-if="podLogsError" class="error-box">{{ podLogsError }}</div>
          <div class="logs-content-wrapper">
            <pre class="logs-content" ref="podLogsContentRef" v-html="highlightedPodLogs"></pre>
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
          <pre class="detail-json">{{ JSON.stringify(selectedPodForAction?.raw || selectedPodForAction, null, 2) }}</pre>
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
          <div v-else-if="podEventsList.length > 0">
            <div v-for="(event, idx) in podEventsList" :key="idx" class="event-item">
              <div class="event-type" :class="(event.type || 'normal').toLowerCase()">{{ event.type || 'Normal' }}</div>
              <div class="event-content">
                <div class="event-reason">{{ event.reason }}</div>
                <div class="event-message">{{ event.message }}</div>
                <div class="event-time">{{ fmtTime(event.event_time || event.lastTimestamp) }}</div>
              </div>
            </div>
          </div>
          <div v-else class="empty-state-small">
            <div class="empty-text">暂无事件</div>
          </div>
        </div>
      </div>
    </div>

    <!-- DaemonSet 日志弹窗（增强版 - 实时日志、高亮） -->
    <div v-if="showDaemonsetLogsModal" class="modal-overlay" @click.self="closeDaemonsetLogsModal">
      <div class="modal-content logs-modal">
        <div class="modal-header">
          <h3>📄 DaemonSet 日志</h3>
          <button class="close-btn" @click="closeDaemonsetLogsModal">×</button>
        </div>
        <div class="modal-body">
          <div v-if="logsDaemonset" class="pod-info-bar">
            <span><strong>命名空间：</strong>{{ logsDaemonset.namespace }}</span>
            <span><strong>DaemonSet：</strong>{{ logsDaemonset.name }}</span>
          </div>
          <div class="logs-controls">
            <div class="control-item" v-if="daemonsetPodList.length > 0">
              <label>选择 Pod</label>
              <select v-model="daemonsetLogsForm.pod" class="form-select" @change="onDaemonsetPodChange">
                <option value="">全部 Pod</option>
                <option v-for="p in daemonsetPodList" :key="p.name" :value="p.name">{{ p.name }}</option>
              </select>
            </div>
            <div class="control-item" v-if="daemonsetLogsForm.pod && daemonsetContainerList.length > 0">
              <label>容器</label>
              <select v-model="daemonsetLogsForm.container" class="form-select">
                <option value="" disabled>选择容器</option>
                <option v-for="c in daemonsetContainerList" :key="c" :value="c">{{ c }}</option>
              </select>
            </div>
            <div class="control-item">
              <label>行数</label>
              <select v-model="daemonsetLogsForm.tail" class="form-select">
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
                <input type="checkbox" v-model="daemonsetLogsForm.follow" />
                <span>实时日志</span>
                <span v-if="daemonsetLogsForm.follow && isStreamingDaemonsetLogs" class="streaming-indicator">●</span>
              </label>
            </div>
            <div class="control-actions">
              <button 
                class="btn btn-primary btn-sm" 
                @click="fetchDaemonsetLogs" 
                :disabled="loadingDaemonsetLogs || isStreamingDaemonsetLogs"
              >
                {{ loadingDaemonsetLogs ? '获取中...' : (daemonsetLogsForm.follow ? '获取实时日志' : '获取日志') }}
              </button>
              <!-- 终止按钮 -->
              <button 
                v-if="loadingDaemonsetLogs || isStreamingDaemonsetLogs" 
                class="btn btn-danger btn-sm" 
                @click="stopDaemonsetLogStream"
              >
                终止
              </button>
              <button class="btn btn-secondary btn-sm" @click="clearDaemonsetLogs" :disabled="!daemonsetLogsContent">清除</button>
            </div>
          </div>
          <div v-if="loadingDaemonsetLogs && !isStreamingDaemonsetLogs" class="loading-state">加载中...</div>
          <div v-else-if="daemonsetLogsError" class="error-box">{{ daemonsetLogsError }}</div>
          <div v-else class="logs-content-wrapper">
            <pre class="logs-content" ref="daemonsetLogsContentRef" v-html="highlightedDaemonsetLogs"></pre>
          </div>
        </div>
      </div>
    </div>

    <!-- YAML 查看/编辑模态框 -->
    <div v-if="showYamlModal" class="modal-overlay" @click.self="closeYamlModal">
      <div class="modal-content yaml-modal">
        <div class="modal-header">
          <h3>📝 {{ yamlEditMode ? '编辑 YAML' : '查看 YAML' }} - {{ selectedYamlDaemonset?.name }}</h3>
          <div class="yaml-header-actions">
            <button v-if="!yamlEditMode" class="btn btn-sm btn-secondary" @click="downloadYaml">📄 下载</button>
            <button v-if="!yamlEditMode" class="btn btn-sm btn-secondary" @click="yamlEditMode = true">✈️ 编辑模式</button>
            <button v-if="yamlEditMode" class="btn btn-sm btn-secondary" @click="yamlEditMode = false">👁️ 预览模式</button>
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

    <!-- 容器终端 -->
    <KubeTerminal
      :visible="showTerminal"
      :namespace="terminalPod.namespace"
      :pod-name="terminalPod.name"
      :container-name="terminalPod.container"
      @close="closeTerminal"
    />

    <!-- 资源状态监听浮窗 -->
    <transition name="watcher-slide">
      <div v-if="watchingStatus" class="resource-watcher-panel">
        <div class="watcher-header">
          <span class="watcher-title">{{ phaseIcon(watchPhase) }} Rollout 监听</span>
          <span class="watcher-elapsed">{{ formatElapsed(watchElapsed) }}</span>
          <button class="watcher-close" @click="stopWatching" title="停止监听">×</button>
        </div>
        <div class="watcher-body">
          <div class="watcher-progress">
            <div class="watcher-progress-bar" :style="{ width: watchProgress + '%', background: phaseColor(watchPhase) }"></div>
          </div>
          <div class="watcher-phase" :style="{ color: phaseColor(watchPhase) }">{{ watchPhase }} ({{ watchProgress }}%)</div>
          <div class="watcher-events" v-if="watchEvents.length > 0">
            <div v-for="(ev, i) in watchEvents.slice(0, 8)" :key="i" class="watcher-event" :class="{ warning: ev.type === 'Warning' }">
              <span class="ev-type">{{ ev.type === 'Warning' ? '⚠' : 'ℹ️' }}</span>
              <span class="ev-reason">{{ ev.reason }}</span>
              <span class="ev-msg">{{ ev.message }}</span>
            </div>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch, watchEffect } from 'vue'
import { Message } from '@arco-design/web-vue'
import Pagination from '@/components/Pagination.vue'
import KubeTerminal from '@/components/KubeTerminal.vue'
import RollbackDialog from '@/components/RollbackDialog.vue'
import daemonsetsApi from '@/api/cluster/workloads/daemonsets'
import podsApi from '@/api/cluster/workloads/pods'
import namespaceApi from '@/api/cluster/namespaces'
import { useClusterStore } from '@/stores/cluster'
import { useResourceWatcher } from '@/composables/useResourceWatcher'
import { useConfirmDialog } from '@/composables/useConfirmDialog'
import permissionStore from '@/stores/permission'

const { confirm } = useConfirmDialog()

// ===== 容器终端 =====
const showTerminal = ref(false)
const terminalPod = ref({ namespace: '', name: '', container: '' })

const openTerminal = (pod) => {
  terminalPod.value = {
    namespace: pod.namespace || podsDaemonset.value?.namespace || '',
    name: pod.name,
    container: pod.containers?.[0] || pod.containerName || '',
  }
  showTerminal.value = true
}

const openTerminalForDaemonSet = async (ds) => {
  try {
    const res = await daemonsetsApi.pods({ namespace: ds.namespace, name: ds.name })
    const list = res.code === 0 ? (res.data?.pods || res.data?.items || res.data?.list || res.data || []) : []
    const pods = list.map(p => ({
      name: p.metadata?.name || p.name,
      namespace: p.metadata?.namespace || ds.namespace,
      status: p.status?.phase || p.status || 'Unknown',
      containers: p.spec?.containers?.map(c => c.name) || p.containers || [],
    }))
    const runningPod = pods.find(p => p.status === 'Running') || pods[0]
    if (runningPod) {
      terminalPod.value = { namespace: runningPod.namespace, name: runningPod.name, container: runningPod.containers?.[0] || '' }
      showTerminal.value = true
    } else {
      Message.warning({ content: '该 DaemonSet 没有可用的 Pod，无法打开终端', duration: 2000 })
    }
  } catch (e) {
    Message.error({ content: '查找 Pod 失败: ' + (e.message || e), duration: 2000 })
  }
}

const closeTerminal = () => {
  showTerminal.value = false
}

// ===== 资源状态监听 =====
const {
  watching: watchingStatus,
  watchPhase,
  watchProgress,
  watchEvents,
  watchElapsed,
  startWatching,
  stopWatching,
  formatElapsed,
  phaseColor,
  phaseIcon,
} = useResourceWatcher()

const startDaemonSetWatcher = (ds) => {
  startWatching(
    { namespace: ds.namespace, name: ds.name, kind: 'DaemonSet' },
    {
      getStatus: async () => {
        try {
          const res = await daemonsetsApi.detail({ namespace: ds.namespace, name: ds.name })
          const statusStr = res?.data?.status || 'Unknown'
          const raw = res?.data?.data || {}
          const s = raw.status || {}
          return {
            status: statusStr,
            desiredReplicas: s.desiredNumberScheduled || 0,
            readyReplicas: s.numberReady || 0,
            updatedReplicas: s.updatedNumberScheduled || 0,
            availableReplicas: s.numberAvailable || 0,
          }
        } catch { return null }
      },
      getEvents: async () => {
        try {
          const res = await daemonsetsApi.events({ namespace: ds.namespace, name: ds.name, limit: 20, since_seconds: 300 })
          return res?.data?.events || res?.data?.items || []
        } catch { return [] }
      },
      onComplete: ({ success, elapsed }) => {
        if (success) {
          Message.success({ content: `DaemonSet ${ds.name} 已就绪（耗时 ${elapsed}s）`, duration: 3000 })
        }
        refreshList()
      },
      pollInterval: 2000,
      eventInterval: 4000,
      timeout: 300000,
    },
  )
}

// ===== 操作权限控制 =====
// viewer 角色只能查看，不能执行任何修改操作
const canOperate = computed(() => {
  if (permissionStore.state.isSuperAdmin) return true
  const roleTypes = permissionStore.roleTypes.value
  // viewer 角色无操作权限
  if (roleTypes.length === 1 && roleTypes.includes('viewer')) return false
  // 其他角色有操作权限
  return roleTypes.some(r => ['super_admin', 'platform_admin', 'cluster_admin', 'developer', 'cicd_admin'].includes(r))
})

// ===== 获取认证头 =====
const getAuthHeaders = () => {
  const headers = { 'Content-Type': 'application/json' }
  const token = localStorage.getItem('token') || sessionStorage.getItem('token')
  if (token) headers['Authorization'] = `Bearer ${token}`
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

// =========================
// 状态变量
// =========================
const loading = ref(true)
const errorMsg = ref('')
const searchQuery = ref('')
const statusFilter = ref('all')
const namespaceFilter = ref('')
const currentPage = ref(1)
const itemsPerPage = ref(10)
const namespaces = ref([])
const daemonsets = ref([])
const viewMode = ref('table')

// ========== 批量操作 ==========
const batchMode = ref(false)
const selectedDaemonsets = ref([])
const deleteConfirmText = ref('')
const restartConfirmText = ref('')
const batchExecuting = ref(false)

const enterBatchMode = () => {
  batchMode.value = true
  selectedDaemonsets.value = [...paginatedDaemonsets.value]
}

const exitBatchMode = () => {
  batchMode.value = false
  selectedDaemonsets.value = []
}

const clearSelection = () => {
  selectedDaemonsets.value = []
}

const isDaemonsetSelected = (ds) => {
  return selectedDaemonsets.value.some(d => d.name === ds.name && d.namespace === ds.namespace)
}

const toggleDaemonsetSelection = (ds) => {
  const index = selectedDaemonsets.value.findIndex(d => d.name === ds.name && d.namespace === ds.namespace)
  if (index >= 0) selectedDaemonsets.value.splice(index, 1)
  else selectedDaemonsets.value.push(ds)
}

const isAllSelected = computed(() => {
  return paginatedDaemonsets.value.length > 0 &&
         paginatedDaemonsets.value.every(ds => isDaemonsetSelected(ds))
})

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    paginatedDaemonsets.value.forEach(ds => {
      const index = selectedDaemonsets.value.findIndex(d => d.name === ds.name && d.namespace === ds.namespace)
      if (index >= 0) selectedDaemonsets.value.splice(index, 1)
    })
  } else {
    paginatedDaemonsets.value.forEach(ds => {
      if (!isDaemonsetSelected(ds)) selectedDaemonsets.value.push(ds)
    })
  }
}

// 批量重启
const batchRestart = async () => {
  if (selectedDaemonsets.value.length === 0) return
  const ok1 = await confirm({
    title: '批量重启确认',
    content: `确定要重启选中的 ${selectedDaemonsets.value.length} 个 DaemonSet 吗？`,
    type: 'warning',
    details: [{ label: '操作数量', value: `${selectedDaemonsets.value.length} 个 DaemonSet` }],
    confirmText: '确认重启'
  })
  if (!ok1) return
  
  batchExecuting.value = true
  let successCount = 0, failCount = 0
  for (const ds of selectedDaemonsets.value) {
    try {
      await daemonsetsApi.restart({ namespace: ds.namespace, name: ds.name })
      successCount++
    } catch { failCount++ }
  }
  batchExecuting.value = false
  
  if (failCount === 0) {
    Message.success({ content: `成功重启 ${successCount} 个 DaemonSet`, duration: 2200 })
  } else {
    Message.warning({ content: `成功 ${successCount} 个，失败 ${failCount} 个`, duration: 2200 })
  }
  exitBatchMode()
  refreshList()
}

// 批量删除
const batchDelete = async () => {
  if (selectedDaemonsets.value.length === 0) return
  const ok2 = await confirm({
    title: '批量删除确认',
    content: `确定要删除选中的 ${selectedDaemonsets.value.length} 个 DaemonSet 吗？`,
    type: 'danger',
    details: [{ label: '操作数量', value: `${selectedDaemonsets.value.length} 个 DaemonSet`, danger: true }],
    tip: '此操作不可恢复！',
    confirmText: '确认删除'
  })
  if (!ok2) return
  
  batchExecuting.value = true
  let successCount = 0, failCount = 0
  for (const ds of selectedDaemonsets.value) {
    try {
      await daemonsetsApi.delete({ namespace: ds.namespace, name: ds.name })
      successCount++
    } catch { failCount++ }
  }
  batchExecuting.value = false
  
  if (failCount === 0) {
    Message.success({ content: `成功删除 ${successCount} 个 DaemonSet`, duration: 2200 })
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
const selectedDaemonset = ref(null)
const menuStyle = ref({})

// Modal 状态
const showCreateModal = ref(false)
const showViewModal = ref(false)
const showEventsModal = ref(false)
const showUpdateImageModal = ref(false)
const showRollbackModal = ref(false)
const showPodsModal = ref(false)
const showHistoryModal = ref(false)
const showDaemonsetLogsModal = ref(false)

// ========== YAML 查看/编辑相关 ==========
const showYamlModal = ref(false)
const selectedYamlDaemonset = ref(null)
const yamlContent = ref('')
const yamlEditMode = ref(false)
const loadingYaml = ref(false)
const savingYaml = ref(false)
const yamlError = ref('')

// YAML 创建相关
const createMode = ref('form') // 'form' | 'yaml'

// 监听 createMode 变化，切换到 YAML 模式时如果内容为空则自动加载模板
watch(createMode, (newMode) => {
  if (newMode === 'yaml' && !yamlContent.value.trim()) {
    loadDaemonSetYamlTemplate()
  }
})

// 创建命名空间相关
const showNamespaceInput = ref(false)
const newNamespace = ref('')
const creatingNamespace = ref(false)

// 加载状态
const loadingDetail = ref(false)
const loadingEvents = ref(false)
const loadingHistory = ref(false)
const loadingPods = ref(false)
const updatingImage = ref(false)
const updatingResources = ref(false)  // 资源调整中
const rollingBack = ref(false)
const creating = ref(false)

// 数据
const detailData = ref(null)
const ddpEventsData = ref([])
const ddpEventsLoading = ref(false)
const ddpShowAnno = ref(false)
const eventsData = ref([])
const historyList = ref([])
const historyDaemonset = ref(null)
const podsList = ref([])
const podsDaemonset = ref(null)

// Pod 操作相关
const showPodLogsModal = ref(false)
const showPodDetailModal = ref(false)
const showPodEventsModal = ref(false)
const showPodMoreOptions = ref(false)
const podMenuStyle = ref({})
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
const podEventsList = ref([])
const loadingPodEvents = ref(false)

// DaemonSet 日志相关
const logsDaemonset = ref(null)
const daemonsetPodList = ref([])
const daemonsetContainerList = ref([])
const daemonsetLogsForm = ref({ pod: '', container: '', tail: 100, follow: false })
const daemonsetLogsContent = ref('')
const daemonsetLogsError = ref('')
const loadingDaemonsetLogs = ref(false)
const isStreamingDaemonsetLogs = ref(false)
const daemonsetLogsContentRef = ref(null)
let daemonsetLogAbortController = null

// 回滚相关
const loadingRollbackHistory = ref(false)
const rollbackHistoryList = ref([])
const rollbackForm = ref({ namespace: '', name: '', revision_name: '' })
const dsRollbackCurrentInfo = computed(() => {
  const ds = daemonsets.value.find(d => d.name === rollbackForm.value.name && d.namespace === rollbackForm.value.namespace)
  if (!ds) return null
  // DaemonSet 无副本数概念，展示 就绪/期望 调度数；本地对象字段为 numberReady/desiredNumberScheduled/createdAt
  return {
    image: ds.image || '—',
    replicas: `${ds.numberReady ?? 0}/${ds.desiredNumberScheduled ?? 0}`,
    deployedAt: ds.createdAt ? fmtTime(ds.createdAt) : '—'
  }
})

// 更新镜像表单
const updateImageForm = ref({ namespace: '', name: '', container: '', image: '', currentImage: '' })
const containerList = ref([])
const imagesList = ref([])

// 资源快速修改表单
const showResourceModal = ref(false)
const resourceForm = ref({
  namespace: '',
  name: '',
  container: '',
  cpu_request: '',
  cpu_limit: '',
  memory_request: '',
  memory_limit: '',
})
const resourceContainerList = ref([])  // 资源弹窗的容器列表
const currentResources = ref(null)     // 当前容器的资源配置

// 创建表单
const daemonsetForm = ref({
  name: '',
  namespace: 'default',
  container_image: '',
  labels: [{ key: 'app', value: '' }],
  nodeSelector: []
})

// =========================
// 生命周期
// =========================
onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('scroll', handleScroll, true)
  document.addEventListener('click', handlePodClickOutside)
  document.addEventListener('scroll', handlePodScroll, true)
  fetchNamespaces().then(fetchDaemonsets)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('scroll', handleScroll, true)
  document.removeEventListener('click', handlePodClickOutside)
  document.removeEventListener('scroll', handlePodScroll, true)
  stopAutoRefresh()
})

// =========================
// 更多菜单控制
// =========================
const toggleMoreOptions = (ds, event) => {
  if (selectedDaemonset.value === ds && showMoreOptions.value) {
    showMoreOptions.value = false
    selectedDaemonset.value = null
  } else {
    selectedDaemonset.value = ds
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
    selectedDaemonset.value = null
  }
}

const handleScroll = () => {
  if (showMoreOptions.value) {
    showMoreOptions.value = false
    selectedDaemonset.value = null
  }
}

// =========================
// 自动刷新
// =========================
const startAutoRefresh = () => {
  if (autoRefreshTimer) return
  autoRefreshTimer = setInterval(() => {
    if (!loading.value) fetchDaemonsets()
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
  fetchDaemonsets()
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

const fetchDaemonsets = async () => {
  loading.value = true
  errorMsg.value = ''
  try {
    const params = { namespace: namespaceFilter.value || '', page: 1, limit: 200 }
    if (searchQuery.value) params.name = searchQuery.value
    const res = await daemonsetsApi.list(params)
    const list = res.data?.list || res.data?.items || []
    if (res.code === 0 && list.length > 0) {
      daemonsets.value = list.map(item => ({
        name: item.name,
        namespace: item.namespace,
        status: item.status || 'Unknown',
        desiredNumberScheduled: item.desired_number_scheduled || item.desiredNumberScheduled || 0,
        numberReady: item.number_ready || item.numberReady || 0,
        image: item.image || (item.images && item.images[0]) || '',
        containers: item.containers || [],
        images: item.images || [],
        selector: item.selector || {},
        updateStrategy: item.update_strategy || item.updateStrategy || 'RollingUpdate',
        createdAt: item.created_at || item.createdAt || ''
      }))
    } else {
      daemonsets.value = []
    }
  } catch (e) {
    console.error('获取 DaemonSet 列表失败:', e)
    errorMsg.value = e?.msg || e?.message || '获取 DaemonSet 列表失败'
    daemonsets.value = []
  } finally {
    loading.value = false
  }
}

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

const refreshList = () => fetchDaemonsets()

// =========================
// 计算属性
// =========================
const filteredDaemonsets = computed(() => {
  let result = daemonsets.value
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

const paginatedDaemonsets = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage.value
  return filteredDaemonsets.value.slice(start, start + itemsPerPage.value)
})

// =========================
// 创建 DaemonSet
// =========================
const addLabel = () => daemonsetForm.value.labels.push({ key: '', value: '' })
const removeLabel = (index) => daemonsetForm.value.labels.splice(index, 1)
const addNodeSelector = () => daemonsetForm.value.nodeSelector.push({ key: '', value: '' })
const removeNodeSelector = (index) => daemonsetForm.value.nodeSelector.splice(index, 1)

const createNewNamespace = async () => {
  if (!newNamespace.value) return
  creatingNamespace.value = true
  try {
    await namespaceApi.create({ name: newNamespace.value })
    Message.success({ content: `命名空间 ${newNamespace.value} 创建成功` })
    await fetchNamespaces()
    daemonsetForm.value.namespace = newNamespace.value
    showNamespaceInput.value = false
    newNamespace.value = ''
  } catch (e) {
    Message.error({ content: e?.msg || '创建命名空间失败' })
  } finally {
    creatingNamespace.value = false
  }
}

const cancelCreateNamespace = () => {
  showNamespaceInput.value = false
  newNamespace.value = ''
}

const createDaemonset = async () => {
  if (!daemonsetForm.value.name || !daemonsetForm.value.container_image) {
    Message.error({ content: '请填写名称和镜像' })
    return
  }
  
  // 验证标签
  const validLabels = daemonsetForm.value.labels.filter(l => l.key && l.value)
  if (validLabels.length === 0) {
    Message.error({ content: '至少需要一个有效的标签（键值对）' })
    return
  }
  
  creating.value = true
  try {
    // 构建节点选择器（如果有）
    const validNodeSelectors = daemonsetForm.value.nodeSelector.filter(s => s.key && s.value)
    const nodeSelectorMap = {}
    validNodeSelectors.forEach(s => {
      nodeSelectorMap[s.key] = s.value
    })
    
    const payload = {
      namespace: daemonsetForm.value.namespace,
      name: daemonsetForm.value.name,
      container_image: daemonsetForm.value.container_image,
      labels: validLabels,  // 直接发送数组格式 [{key, value}]
      node_selector: Object.keys(nodeSelectorMap).length > 0 ? nodeSelectorMap : undefined
    }
    const res = await daemonsetsApi.create(payload)
    if (res.code === 0) {
      Message.success({ content: 'DaemonSet 创建成功' })
      showCreateModal.value = false
      resetDaemonsetForm()
      refreshList()
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

const resetDaemonsetForm = () => {
  daemonsetForm.value = {
    name: '',
    namespace: 'default',
    container_image: '',
    labels: [{ key: 'app', value: '' }],
    nodeSelector: []
  }
  // 重置 YAML 创建状态
  createMode.value = 'form'
  yamlContent.value = ''
  yamlError.value = ''
}

// YAML 创建相关函数
const loadDaemonSetYamlTemplate = () => {
  yamlContent.value = `apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: example-daemonset
  namespace: default
  labels:
    app: example
spec:
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
      nodeSelector:
        node-role: worker`
  yamlError.value = ''
  Message.success({ content: '已加载 YAML 模板，请修改后创建' })
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
const resetYamlContent = async () => {
  if (yamlContent.value.trim()) {
    const ok = await confirm({
      title: '重置 YAML 内容',
      content: '确定要清空当前编辑的 YAML 内容吗？',
      type: 'warning',
      confirmText: '确认重置',
      cancelText: '取消',
    })
    if (!ok) return
  }
  yamlContent.value = ''
  yamlError.value = ''
  Message.success({ content: 'YAML 内容已重置' })
}

const createDaemonSetFromYaml = async () => {
  if (!yamlContent.value.trim()) {
    Message.error({ content: '请输入 YAML 内容' })
    return
  }
  
  // 简单验证
  if (!yamlContent.value.includes('kind: DaemonSet')) {
    yamlError.value = 'YAML 中必须包含 "kind: DaemonSet"'
    return
  }
  if (!yamlContent.value.includes('apiVersion: apps/v1')) {
    yamlError.value = 'YAML 中必须包含 "apiVersion: apps/v1"'
    return
  }
  
  yamlError.value = ''
  
  try {
    const res = await daemonsetsApi.createFromYaml({ yaml: yamlContent.value })
    if (res.code === 0) {
      const msg = res.data?.message || 'DaemonSet 创建成功'
      Message.success({ content: msg, duration: 5000 })
      showCreateModal.value = false
      resetDaemonsetForm()
      await fetchDaemonsets()
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

// =========================
// 详情
// =========================
const viewDaemonset = async (ds) => {
  showMoreOptions.value = false
  loadingDetail.value = true
  detailData.value = null
  ddpEventsData.value = []
  ddpShowAnno.value = false
  showViewModal.value = true
  ddpEventsLoading.value = true
  try {
    const [detailRes, eventsRes] = await Promise.all([
      daemonsetsApi.detail({ namespace: ds.namespace, name: ds.name }),
      daemonsetsApi.events({ namespace: ds.namespace, name: ds.name })
    ])
    detailData.value = detailRes.code === 0 ? (detailRes.data?.data || detailRes.data) : ds
    ddpEventsData.value = eventsRes.code === 0 ? (eventsRes.data?.events || eventsRes.data?.items || eventsRes.data || []) : []
  } catch (e) {
    detailData.value = ds
    ddpEventsData.value = []
  } finally {
    loadingDetail.value = false
    ddpEventsLoading.value = false
  }
}
const ddpStatusText = (d) => {
  if (!d?.status) return 'UNKNOWN'
  const ready = d.status?.numberReady || 0
  const desired = d.status?.desiredNumberScheduled || 0
  if (ready >= desired && desired > 0) return 'READY'
  if (ready > 0) return 'PROGRESSING'
  return 'NOT READY'
}
const ddpStatusClass = (d) => {
  const t = ddpStatusText(d)
  if (t === 'READY') return 'ddp-st-ok'
  if (t === 'PROGRESSING') return 'ddp-st-prog'
  return 'ddp-st-err'
}
const ddpRefreshEvents = async () => {
  if (!detailData.value?.metadata) return
  ddpEventsLoading.value = true
  try {
    const res = await daemonsetsApi.events({ namespace: detailData.value.metadata.namespace, name: detailData.value.metadata.name })
    ddpEventsData.value = res.code === 0 ? (res.data?.events || res.data?.items || res.data || []) : []
  } catch { ddpEventsData.value = [] }
  finally { ddpEventsLoading.value = false }
}
const ddpFmtEvTime = (ts) => {
  if (!ts) return '-'
  const d = new Date(ts)
  const diff = Math.floor((Date.now() - d.getTime()) / 1000)
  if (diff < 60) return `${diff}s ago`
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`
  return `${Math.floor(diff / 86400)}d ago`
}

// =========================
// 事件
// =========================
const openEvents = async (ds) => {
  showMoreOptions.value = false
  loadingEvents.value = true
  eventsData.value = []
  showEventsModal.value = true
  try {
    const res = await daemonsetsApi.events({ namespace: ds.namespace, name: ds.name })
    eventsData.value = res.code === 0 ? (res.data?.events || res.data || []) : []
  } catch (e) { console.error('获取事件失败:', e) }
  finally { loadingEvents.value = false }
}

// =========================
// 更新镜像（内联）
// =========================
const startInlineImage = (ds) => {
  containerList.value = ds.containers || []
  imagesList.value = ds.images || []
  const firstContainer = containerList.value[0] || ''
  const firstImage = imagesList.value[0] || ds.image || ''
  updateImageForm.value = {
    namespace: ds.namespace,
    name: ds.name,
    container: firstContainer,
    image: '',
    currentImage: firstImage
  }
  showUpdateImageModal.value = true
}

const openUpdateImage = (ds) => {
  showMoreOptions.value = false
  startInlineImage(ds)
}

// 容器切换时更新当前镜像显示
const onUpdateImageContainerChange = () => {
  const idx = containerList.value.indexOf(updateImageForm.value.container)
  updateImageForm.value.currentImage = idx >= 0 ? (imagesList.value[idx] || '') : ''
}

const submitUpdateImage = async () => {
  if (!updateImageForm.value.image) {
    Message.error({ content: '请输入镜像地址' })
    return
  }
  // 二次确认
  const okImg = await confirm({
    title: '镜像更新确认',
    content: '此操作将触发滚动更新，请确认以下信息',
    type: 'warning',
    details: [
      { label: 'DaemonSet', value: `${updateImageForm.value.namespace}/${updateImageForm.value.name}` },
      { label: '容器', value: updateImageForm.value.container },
      { label: '当前镜像', value: updateImageForm.value.currentImage || '未知' },
      { label: '新镜像', value: updateImageForm.value.image, highlight: true }
    ],
    confirmText: '确认更新'
  })
  if (!okImg) return
  updatingImage.value = true
  try {
    const res = await daemonsetsApi.updateImage({
      namespace: updateImageForm.value.namespace,
      name: updateImageForm.value.name,
      container: updateImageForm.value.container,
      image: updateImageForm.value.image
    })
    if (res.code === 0) {
      Message.success({ content: '镜像更新成功，开始监听 Rollout 状态...' })
      showUpdateImageModal.value = false
      autoRefresh.value = true
      setTimeout(() => { autoRefresh.value = false }, 15000)
      refreshList()
      startDaemonSetWatcher({ namespace: updateImageForm.value.namespace, name: updateImageForm.value.name })
    } else {
      Message.error({ content: res.msg || '更新失败' })
    }
  } catch (e) {
    Message.error({ content: e?.msg || '更新失败' })
  } finally {
    updatingImage.value = false
  }
}

// =========================
// 资源快速调整
// =========================
const openResourceModal = async (ds) => {
  showMoreOptions.value = false
  resourceContainerList.value = ds.containers || []
  const firstContainer = resourceContainerList.value[0] || ''
  resourceForm.value = {
    namespace: ds.namespace,
    name: ds.name,
    container: firstContainer,
    cpu_request: '',
    cpu_limit: '',
    memory_request: '',
    memory_limit: '',
  }
  currentResources.value = null

  // 获取当前容器的资源配置
  if (firstContainer) {
    await loadContainerResources(ds, firstContainer)
  }
  showResourceModal.value = true
}

const loadContainerResources = async (ds, containerName) => {
  try {
    const res = await daemonsetsApi.detail({
      namespace: ds.namespace,
      name: ds.name
    })
    if (res.code === 0 && res.data) {
      const d = res.data?.data || res.data
      const containers = d.spec?.template?.spec?.containers || []
      const target = containers.find(c => c.name === containerName)
      if (target && target.resources) {
        const r = target.resources
        currentResources.value = {
          cpu_request: r.requests?.cpu || '',
          cpu_limit: r.limits?.cpu || '',
          memory_request: r.requests?.memory || '',
          memory_limit: r.limits?.memory || '',
        }
        resourceForm.value.cpu_request = currentResources.value.cpu_request
        resourceForm.value.cpu_limit = currentResources.value.cpu_limit
        resourceForm.value.memory_request = currentResources.value.memory_request
        resourceForm.value.memory_limit = currentResources.value.memory_limit
      } else {
        currentResources.value = null
      }
    }
  } catch (e) {
    console.warn('获取容器资源配置失败:', e)
  }
}

const onResourceContainerChange = async () => {
  const ds = { namespace: resourceForm.value.namespace, name: resourceForm.value.name }
  await loadContainerResources(ds, resourceForm.value.container)
}

const submitUpdateResources = async () => {
  if (!resourceForm.value.memory_limit) {
    Message.error({ content: '内存上限（Memory Limit）必须设置，防止 OOM' })
    return
  }

  const ok = await confirm({
    title: '确认修改资源配置',
    content: '此操作将触发滚动更新，修改容器的 CPU/Memory 配置。',
    type: 'warning',
    details: [
      { label: 'DaemonSet', value: `${resourceForm.value.namespace}/${resourceForm.value.name}` },
      { label: '容器', value: resourceForm.value.container },
      { label: 'Memory Limit', value: resourceForm.value.memory_limit, highlight: true },
      { label: 'Memory Request', value: resourceForm.value.memory_request || '未设置' },
      { label: 'CPU Limit', value: resourceForm.value.cpu_limit || '未设置' },
      { label: 'CPU Request', value: resourceForm.value.cpu_request || '未设置' },
    ],
    tip: '资源修改会触发滚动更新，期间服务保持可用',
    confirmText: '确认修改',
  })
  if (!ok) return

  updatingResources.value = true
  try {
    const res = await daemonsetsApi.updateResources(resourceForm.value)
    if (res.code === 0) {
      Message.success({ content: '资源配置更新成功' })
      showResourceModal.value = false
      autoRefresh.value = true
      setTimeout(() => { autoRefresh.value = false }, 15000)
      refreshList()
    } else {
      Message.error({ content: res.msg || '资源配置更新失败' })
    }
  } catch (e) {
    Message.error({ content: e?.msg || e?.message || '资源配置更新失败' })
  } finally {
    updatingResources.value = false
  }
}

// =========================
// 重启
// =========================
const restartDaemonset = async (ds) => {
  showMoreOptions.value = false
  const okR = await confirm({
    title: '重启确认',
    content: `确定重启 DaemonSet: ${ds.name}？`,
    type: 'warning',
    confirmText: '确认重启'
  })
  if (!okR) return
  try {
    const res = await daemonsetsApi.restart({ namespace: ds.namespace, name: ds.name })
    if (res.code === 0) {
      Message.success({ content: '重启成功，正在滚动更新' })
      autoRefresh.value = true
      setTimeout(() => { autoRefresh.value = false }, 15000)
      refreshList()
    } else {
      Message.error({ content: res.msg || '重启失败' })
    }
  } catch (e) {
    Message.error({ content: e?.msg || '重启失败' })
  }
}

// =========================
// 删除
// =========================
const deleteDaemonset = async (ds) => {
  showMoreOptions.value = false
  const okD = await confirm({
    title: '删除确认',
    content: `确定删除 DaemonSet: ${ds.name}？`,
    type: 'danger',
    tip: '此操作不可恢复！',
    confirmText: '确认删除'
  })
  if (!okD) return
  try {
    const res = await daemonsetsApi.delete({ namespace: ds.namespace, name: ds.name })
    if (res.code === 0) {
      Message.success({ content: '删除成功' })
      refreshList()
    } else {
      Message.error({ content: res.msg || '删除失败' })
    }
  } catch (e) {
    Message.error({ content: e?.msg || '删除失败' })
  }
}

// =========================
// 版本历史
// =========================
const viewHistory = async (ds) => {
  showMoreOptions.value = false
  historyDaemonset.value = ds
  loadingHistory.value = true
  historyList.value = []
  showHistoryModal.value = true
  try {
    const res = await daemonsetsApi.history({ namespace: ds.namespace, name: ds.name })
    // 后端返回 snake_case（creation_time），映射为组件所需的 createdAt 并格式化
    historyList.value = res.code === 0 ? (res.data || []).map(r => ({ ...r, createdAt: fmtTime(r.creation_time || r.createdAt) })) : []
  } catch (e) { console.error('获取历史失败:', e) }
  finally { loadingHistory.value = false }
}

// 从版本记录直接回滚
const rollbackToVersion = async (rev) => {
  if (!historyDaemonset.value) return
  const okRb = await confirm({
    title: '回滚确认',
    content: `确定回滚到版本 ${rev.revision}？`,
    type: 'warning',
    details: [{ label: '目标版本', value: `Revision ${rev.revision}`, highlight: true }],
    confirmText: '确认回滚'
  })
  if (!okRb) return
  rollingBack.value = true
  try {
    const res = await daemonsetsApi.rollback({
      namespace: historyDaemonset.value.namespace,
      name: historyDaemonset.value.name,
      revision_name: rev.name
    })
    if (res.code === 0) {
      Message.success({ content: '回滚成功' })
      showHistoryModal.value = false
      autoRefresh.value = true
      setTimeout(() => { autoRefresh.value = false }, 15000)
      refreshList()
    } else {
      Message.error({ content: res.msg || '回滚失败' })
    }
  } catch (e) {
    Message.error({ content: e?.msg || '回滚失败' })
  } finally {
    rollingBack.value = false
  }
}

// =========================
// 回滚弹窗
// =========================
const openRollback = async (ds) => {
  showMoreOptions.value = false
  rollbackForm.value = { namespace: ds.namespace, name: ds.name, revision_name: '' }
  loadingRollbackHistory.value = true
  rollbackHistoryList.value = []
  showRollbackModal.value = true
  try {
    const res = await daemonsetsApi.history({ namespace: ds.namespace, name: ds.name })
    // 后端返回 snake_case（creation_time），映射为回滚弹窗所需的 createdAt 并格式化
    rollbackHistoryList.value = res.code === 0 ? (res.data || []).map(r => ({ ...r, createdAt: fmtTime(r.creation_time || r.createdAt) })) : []
  } catch (e) { console.error('获取历史失败:', e) }
  finally { loadingRollbackHistory.value = false }
}

const submitRollback = async () => {
  if (!rollbackForm.value.revision_name) {
    Message.error({ content: '请选择版本' })
    return
  }
  rollingBack.value = true
  try {
    const res = await daemonsetsApi.rollback({
      namespace: rollbackForm.value.namespace,
      name: rollbackForm.value.name,
      revision_name: rollbackForm.value.revision_name
    })
    if (res.code === 0) {
      Message.success({ content: '回滚成功' })
      showRollbackModal.value = false
      autoRefresh.value = true
      setTimeout(() => { autoRefresh.value = false }, 15000)
      refreshList()
    } else {
      Message.error({ content: res.msg || '回滚失败' })
    }
  } catch (e) {
    Message.error({ content: e?.msg || '回滚失败' })
  } finally {
    rollingBack.value = false
  }
}

// =========================
// Pods 关联
// =========================
const viewPods = async (ds) => {
  podsDaemonset.value = ds
  loadingPods.value = true
  podsList.value = []
  showPodsModal.value = true
  try {
    const res = await daemonsetsApi.pods({ namespace: ds.namespace, name: ds.name })
    const list = res.code === 0 ? (res.data?.pods || res.data || []) : []
    podsList.value = list.map(p => {
      const metadata = p.metadata || {}
      const spec = p.spec || {}
      const status = p.status || {}
      return {
        name: metadata.name || p.name,
        namespace: metadata.namespace || ds.namespace,
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

// =========================
// Pod 更多菜单
// =========================
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
      const menuHeight = 280
      let style = { position: 'fixed', left: rect.left + 'px' }
      if (rect.bottom + menuHeight > viewportHeight) {
        style.bottom = (viewportHeight - rect.top + 4) + 'px'
      } else {
        style.top = (rect.bottom + 4) + 'px'
      }
      podMenuStyle.value = style
    }
  }
}

const handlePodClickOutside = (event) => {
  if (showPodMoreOptions.value && !event.target.closest('.more-btn')) {
    showPodMoreOptions.value = false
    selectedPodForAction.value = null
  }
}

const handlePodScroll = () => {
  if (showPodMoreOptions.value) {
    showPodMoreOptions.value = false
    selectedPodForAction.value = null
  }
}

// =========================
// Pod 日志
// =========================
const openPodLogs = (pod) => {
  showPodMoreOptions.value = false
  selectedPodForAction.value = pod
  podContainerList.value = pod.containers || []
  podLogsForm.value = {
    container: podContainerList.value.length === 1 ? podContainerList.value[0] : '',
    tail: 100,
    follow: false,
    duration: 0
  }
  podLogsContent.value = ''
  podLogsError.value = ''
  showPodLogsModal.value = true
}

const fetchPodLogsContent = async () => {
  if (!selectedPodForAction.value) return
  let container = podLogsForm.value.container || ''
  if (podContainerList.value.length === 1) container = podContainerList.value[0]
  if (!container) {
    podLogsError.value = '请选择容器'
    return
  }
  loadingPodLogs.value = true
  podLogsError.value = ''
  podLogsContent.value = ''
  
  try {
    if (podLogsForm.value.follow) {
      await fetchPodStreamLogs(container)
    } else {
      await fetchPodStaticLogs(container)
    }
  } catch (e) {
    if (e.name !== 'AbortError') {
      podLogsError.value = e?.message || '获取日志失败'
      loadingPodLogs.value = false
    }
  }
}

const fetchPodStaticLogs = async (container) => {
  const params = new URLSearchParams({
    namespace: selectedPodForAction.value.namespace,
    name: selectedPodForAction.value.name,
    container: container
  })
  if (podLogsForm.value.tail != null) {
    params.set('tail', podLogsForm.value.tail)
  }
  
  podLogAbortController = new AbortController()
  const response = await fetch(`/api/v1/k8s/pod/container_logs?${params}`, {
    signal: podLogAbortController.signal,
    headers: getAuthHeaders()
  })
  
  if (!response.ok) throw new Error(`HTTP ${response.status}`)
  
  const res = await response.json()
  podLogsContent.value = res?.data?.log || '暂无日志'
  loadingPodLogs.value = false
}

const fetchPodStreamLogs = async (container) => {
  isStreamingPodLogs.value = true
  podLogAbortController = new AbortController()
  
  const params = new URLSearchParams({
    namespace: selectedPodForAction.value.namespace,
    name: selectedPodForAction.value.name,
    container: container,
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
    headers: getAuthHeaders()
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

const closePodLogsModal = () => {
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
  
  const escapeHtml = (str) => {
    return str
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;');
  }
  
  const escaped = escapeHtml(podLogsContent.value)
  
  return escaped.split('\n').map(line => {
    let highlighted = line.replace(
      /(\d{4}-\d{2}-\d{2}[T\s]\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:?\d{2})?)/g,
      '<span class="log-timestamp">$1</span>'
    )
    
    if (/\b(ERROR|FATAL|PANIC|error|fatal|panic|失败|错误)\b/i.test(line)) {
      highlighted = `<span class="log-error">${highlighted}</span>`
    } else if (/\b(WARN|WARNING|warn|warning|警告)\b/i.test(line)) {
      highlighted = `<span class="log-warn">${highlighted}</span>`
    } else if (/\b(INFO|info)\b/i.test(line)) {
      highlighted = `<span class="log-info">${highlighted}</span>`
    } else if (/\b(DEBUG|debug)\b/i.test(line)) {
      highlighted = `<span class="log-debug">${highlighted}</span>`
    } else if (/^=+ Pod:/.test(line)) {
      highlighted = `<span class="log-separator">${highlighted}</span>`
    }
    
    return highlighted
  }).join('\n')
})

// =========================
// Pod 详情
// =========================
const openPodDetail = (pod) => {
  showPodMoreOptions.value = false
  selectedPodForAction.value = pod
  showPodDetailModal.value = true
}

// =========================
// Pod 事件
// =========================
const openPodEvents = async (pod) => {
  showPodMoreOptions.value = false
  selectedPodForAction.value = pod
  loadingPodEvents.value = true
  podEventsList.value = []
  showPodEventsModal.value = true
  try {
    const res = await podsApi.events({ namespace: pod.namespace, name: pod.name })
    podEventsList.value = res.code === 0 ? (res.data?.events || res.data || []) : []
  } catch (e) { console.error('获取 Pod 事件失败:', e) }
  finally { loadingPodEvents.value = false }
}

// =========================
// Pod 重启/删除
// =========================
const restartPodFromList = async (pod) => {
  showPodMoreOptions.value = false
  const okPr = await confirm({
    title: 'Pod 重启确认',
    content: `确认重启 Pod: ${pod.name}？`,
    type: 'warning',
    confirmText: '确认重启'
  })
  if (!okPr) return
  try {
    await podsApi.graceDelete({ namespace: pod.namespace, name: pod.name })
    Message.success({ content: 'Pod 重启中...' })
    if (podsDaemonset.value) await viewPods(podsDaemonset.value)
  } catch (e) {
    Message.error({ content: e?.msg || '重启失败' })
  }
}

const deletePodFromList = async (pod) => {
  showPodMoreOptions.value = false
  const okPd = await confirm({
    title: 'Pod 删除确认',
    content: `确认删除 Pod: ${pod.name}？`,
    type: 'danger',
    confirmText: '确认删除'
  })
  if (!okPd) return
  try {
    await podsApi.graceDelete({ namespace: pod.namespace, name: pod.name })
    Message.success({ content: 'Pod 已删除' })
    if (podsDaemonset.value) await viewPods(podsDaemonset.value)
  } catch (e) {
    Message.error({ content: e?.msg || '删除失败' })
  }
}

const forceDeletePodFromList = async (pod) => {
  showPodMoreOptions.value = false
  const okFd = await confirm({
    title: '强制删除确认',
    content: `确认强制删除 Pod: ${pod.name}？`,
    type: 'danger',
    tip: '此操作会立即终止 Pod，不会等待优雅终止！',
    confirmText: '强制删除'
  })
  if (!okFd) return
  try {
    await podsApi.forceDelete({ namespace: pod.namespace, name: pod.name })
    Message.success({ content: 'Pod 已强制删除' })
    if (podsDaemonset.value) await viewPods(podsDaemonset.value)
  } catch (e) {
    Message.error({ content: e?.msg || '强制删除失败' })
  }
}

// =========================
// DaemonSet 日志
// =========================
const viewDaemonsetLogs = async (ds) => {
  showMoreOptions.value = false
  logsDaemonset.value = ds
  daemonsetLogsContent.value = ''
  daemonsetLogsError.value = ''
  daemonsetPodList.value = []
  daemonsetContainerList.value = []
  daemonsetLogsForm.value = { pod: '', container: '', tail: 100, follow: false }
  showDaemonsetLogsModal.value = true
  
  // 获取 Pod 列表
  try {
    const res = await daemonsetsApi.pods({ namespace: ds.namespace, name: ds.name })
    const list = res.code === 0 ? (res.data?.pods || res.data || []) : []
    daemonsetPodList.value = list.map(p => ({
      name: p.metadata?.name || p.name,
      containers: p.spec?.containers?.map(c => c.name) || []
    }))
    // 不自动选择第一个 Pod，让用户可以选择"全部 Pod"
  } catch (e) {
    console.error('获取 Pod 列表失败:', e)
  }
}

const onDaemonsetPodChange = () => {
  const selected = daemonsetPodList.value.find(p => p.name === daemonsetLogsForm.value.pod)
  daemonsetContainerList.value = selected?.containers || []
  daemonsetLogsForm.value.container = daemonsetContainerList.value.length > 0 ? daemonsetContainerList.value[0] : ''
}

const fetchDaemonsetLogs = async () => {
  loadingDaemonsetLogs.value = true
  daemonsetLogsError.value = ''
  daemonsetLogsContent.value = ''
  
  try {
    if (daemonsetLogsForm.value.follow) {
      // 实时日志需要选择单个 Pod
      if (!daemonsetLogsForm.value.pod) {
        daemonsetLogsError.value = '实时日志需要选择单个 Pod'
        loadingDaemonsetLogs.value = false
        return
      }
      await fetchDaemonsetStreamLogs()
    } else {
      await fetchDaemonsetStaticLogs()
    }
  } catch (e) {
    if (e.name !== 'AbortError') {
      daemonsetLogsError.value = e?.message || '获取日志失败'
      loadingDaemonsetLogs.value = false
    }
  }
}

const fetchDaemonsetStaticLogs = async () => {
  // 如果选择了特定 Pod
  if (daemonsetLogsForm.value.pod) {
    if (!daemonsetLogsForm.value.container) {
      daemonsetLogsError.value = '请选择容器'
      loadingDaemonsetLogs.value = false
      return
    }
    
    const params = new URLSearchParams({
      namespace: logsDaemonset.value.namespace,
      name: daemonsetLogsForm.value.pod,
      container: daemonsetLogsForm.value.container
    })
    if (daemonsetLogsForm.value.tail != null) {
      params.set('tail', daemonsetLogsForm.value.tail)
    }
    
    const response = await fetch(`/api/v1/k8s/pod/container_logs?${params}`, {
      headers: getAuthHeaders()
    })
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const res = await response.json()
    daemonsetLogsContent.value = res?.data?.log || '暂无日志'
    loadingDaemonsetLogs.value = false
  } else {
    // 获取所有 Pod 的日志
    const pods = daemonsetPodList.value
    if (pods.length === 0) {
      daemonsetLogsError.value = '暂无 Pod'
      loadingDaemonsetLogs.value = false
      return
    }
    
    const logsArray = []
    
    for (const pod of pods) {
      const container = pod.containers[0] || ''
      if (!container) continue
      
      const params = new URLSearchParams({
        namespace: logsDaemonset.value.namespace,
        name: pod.name,
        container
      })
      if (daemonsetLogsForm.value.tail != null) {
        params.set('tail', daemonsetLogsForm.value.tail)
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
    
    daemonsetLogsContent.value = logsArray.join('\n')
    loadingDaemonsetLogs.value = false
  }
}

const fetchDaemonsetStreamLogs = async () => {
  if (!daemonsetLogsForm.value.container) {
    daemonsetLogsError.value = '请选择容器'
    loadingDaemonsetLogs.value = false
    return
  }
  
  isStreamingDaemonsetLogs.value = true
  daemonsetLogAbortController = new AbortController()
  
  const params = new URLSearchParams({
    namespace: logsDaemonset.value.namespace,
    name: daemonsetLogsForm.value.pod,
    container: daemonsetLogsForm.value.container,
    follow: 'true'
  })
  if (daemonsetLogsForm.value.tail != null) {
    params.set('tail', daemonsetLogsForm.value.tail)
  }
  
  const response = await fetch(`/api/v1/k8s/pod/container_log?${params}`, {
    signal: daemonsetLogAbortController.signal,
    headers: getAuthHeaders()
  })
  
  if (!response.ok) throw new Error(`HTTP ${response.status}`)
  
  const reader = response.body.getReader()
  const decoder = new TextDecoder('utf-8')
  
  while (true) {
    const { done, value } = await reader.read()
    if (done) break
    
    const chunk = decoder.decode(value, { stream: true })
    daemonsetLogsContent.value += chunk
    
    // 自动滚动到底部
    if (daemonsetLogsContentRef.value) {
      daemonsetLogsContentRef.value.scrollTop = daemonsetLogsContentRef.value.scrollHeight
    }
  }
  
  isStreamingDaemonsetLogs.value = false
  loadingDaemonsetLogs.value = false
}

const stopDaemonsetLogStream = () => {
  if (daemonsetLogAbortController) {
    daemonsetLogAbortController.abort()
    daemonsetLogAbortController = null
  }
  isStreamingDaemonsetLogs.value = false
  loadingDaemonsetLogs.value = false
}

const clearDaemonsetLogs = () => {
  daemonsetLogsContent.value = ''
  daemonsetLogsError.value = ''
}

const closeDaemonsetLogsModal = () => {
  stopDaemonsetLogStream()
  showDaemonsetLogsModal.value = false
  daemonsetLogsContent.value = ''
  daemonsetLogsError.value = ''
  daemonsetPodList.value = []
  daemonsetContainerList.value = []
}

// DaemonSet 日志高亮处理
const highlightedDaemonsetLogs = computed(() => {
  if (!daemonsetLogsContent.value) {
    return '<span class="log-placeholder">暂无日志，请点击"获取日志"按钮</span>'
  }
  
  const escapeHtml = (str) => {
    return str
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;');
  }
  
  const escaped = escapeHtml(daemonsetLogsContent.value)
  
  return escaped.split('\n').map(line => {
    let highlighted = line.replace(
      /(\d{4}-\d{2}-\d{2}[T\s]\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:?\d{2})?)/g,
      '<span class="log-timestamp">$1</span>'
    )
    
    if (/\b(ERROR|FATAL|PANIC|error|fatal|panic|失败|错误)\b/i.test(line)) {
      highlighted = `<span class="log-error">${highlighted}</span>`
    } else if (/\b(WARN|WARNING|warn|warning|警告)\b/i.test(line)) {
      highlighted = `<span class="log-warn">${highlighted}</span>`
    } else if (/\b(INFO|info)\b/i.test(line)) {
      highlighted = `<span class="log-info">${highlighted}</span>`
    } else if (/\b(DEBUG|debug)\b/i.test(line)) {
      highlighted = `<span class="log-debug">${highlighted}</span>`
    } else if (/^=+ Pod:/.test(line)) {
      highlighted = `<span class="log-separator">${highlighted}</span>`
    }
    
    return highlighted
  }).join('\n')
})

// ========== YAML 查看/编辑功能 ==========
const openYamlPreview = async (ds) => {
  showMoreOptions.value = false
  selectedYamlDaemonset.value = ds
  yamlContent.value = ''
  yamlError.value = ''
  yamlEditMode.value = false
  loadingYaml.value = true
  showYamlModal.value = true

  try {
    const res = await daemonsetsApi.yaml({ namespace: ds.namespace, name: ds.name })
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
  selectedYamlDaemonset.value = null
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
    const res = await daemonsetsApi.applyYaml({
      namespace: selectedYamlDaemonset.value.namespace,
      name: selectedYamlDaemonset.value.name,
      yaml: yamlContent.value
    })
    if (res.code === 0) {
      Message.success({ content: 'YAML 应用成功' })
      closeYamlModal()
      refreshList()
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

// 下载 YAML
const downloadYaml = () => {
  if (!yamlContent.value || !selectedYamlDaemonset.value) {
    Message.warning({ content: '没有可下载的 YAML 内容' })
    return
  }
  
  try {
    const blob = new Blob([yamlContent.value], { type: 'text/yaml;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `${selectedYamlDaemonset.value.name}-daemonset.yaml`
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
.resource-view {
  width: 100%;
  display: flex;
  flex-direction: column;
  min-height: 0;
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

.btn-sm {
  padding: 6px 12px;
  font-size: 13px;
}

.error-box {
  background: #fff5f5;
  border: 1px solid #feb2b2;
  color: #c53030;
  padding: 12px 16px;
  border-radius: 8px;
  margin-bottom: 16px;
}

/* 批量操作浮动栏 */
.batch-action-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 20px;
  background: linear-gradient(135deg, #4338ca 0%, #3730a3 100%);
  color: white;
  border-radius: 10px;
  margin-bottom: 20px;
  box-shadow: 0 4px 12px rgba(67, 56, 202, 0.3);
}

.batch-info {
  display: flex;
  align-items: center;
  gap: 16px;
}

.batch-count {
  font-weight: 600;
}

.batch-clear {
  background: rgba(255, 255, 255, 0.2);
  border: none;
  color: white;
  padding: 4px 10px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
}

.batch-actions {
  display: flex;
  gap: 10px;
}

.batch-btn {
  background: rgba(255, 255, 255, 0.15);
  border: 1px solid rgba(255, 255, 255, 0.3);
  color: white;
  padding: 8px 14px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  transition: all 0.2s;
}

.batch-btn:hover {
  background: rgba(255, 255, 255, 0.25);
}

.batch-btn.danger {
  background: rgba(239, 68, 68, 0.3);
  border-color: rgba(239, 68, 68, 0.5);
}

.batch-btn.danger:hover {
  background: rgba(239, 68, 68, 0.5);
}

/* 视图切换 */
.view-toggle {
  display: flex;
  gap: 4px;
}

.btn-view {
  padding: 8px 12px;
  background: #e2e8f0;
  border: none;
  border-radius: 6px;
  cursor: pointer;
}

.btn-view.active {
  background: #326ce5;
}

.btn-batch {
  background: #e2e8f0;
  color: #4a5568;
}

/* 表格 */
.table-container {
  background: white;
  border-radius: 12px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.05);
  overflow-x: auto;
}

.resource-table {
  width: 100%;
  border-collapse: collapse;
  min-width: 1200px;
}

.resource-table th {
  background-color: #f7fafc;
  text-align: left;
  padding: 14px 12px;
  font-size: 14px;
  font-weight: 600;
  color: #4a5568;
  border-bottom: 1px solid #e2e8f0;
  white-space: nowrap;
}

.resource-table td {
  padding: 14px 12px;
  font-size: 14px;
  color: #2d3748;
  border-bottom: 1px solid #f7fafc;
}

.resource-table tbody tr:hover {
  background-color: #f7fafc;
}

.row-selected {
  background-color: rgba(50, 108, 229, 0.08) !important;
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

.status-indicator.unknown {
  background-color: rgba(107, 114, 128, 0.1);
  color: #6b7280;
}

.daemonset-name {
  display: flex;
  align-items: center;
  gap: 8px;
}

.daemonset-name .icon {
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

.replicas-display {
  display: flex;
  align-items: center;
  gap: 2px;
  font-size: 14px;
  font-weight: 500;
}

.replicas-sep {
  color: #a0aec0;
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
  max-width: 320px;
  position: relative;
}

.image-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: inline-block;
  max-width: 280px;
  vertical-align: middle;
  cursor: pointer;
}

.selector-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  max-width: 220px;
}

.selector-tag {
  display: inline-block;
  padding: 3px 8px;
  background-color: #edf2f7;
  color: #4a5568;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
  max-width: 180px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  cursor: default;
}

.selector-tag:hover {
  background-color: #e2e8f0;
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

/* 操作按钮 */
.action-icons {
  display: flex;
  gap: 6px;
  align-items: center;
  flex-wrap: nowrap;
}

.action-btn {
  border: none;
  font-size: 12px;
  cursor: pointer;
  padding: 5px 8px;
  border-radius: 6px;
  transition: all 0.2s;
  white-space: nowrap;
  font-weight: 500;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.action-btn.primary {
  background: linear-gradient(135deg, #326ce5 0%, #2558c9 100%);
  color: white;
  box-shadow: 0 2px 4px rgba(50, 108, 229, 0.2);
}

.action-btn.primary:hover {
  background: linear-gradient(135deg, #2558c9 0%, #1e45a0 100%);
  transform: translateY(-1px);
}

.action-btn.terminal {
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
  color: #00d4aa;
  font-family: 'Consolas', 'Monaco', monospace;
  letter-spacing: 0.5px;
}

.action-btn.terminal:hover {
  background: linear-gradient(135deg, #16213e 0%, #0f3460 100%);
  color: #00ffcc;
  box-shadow: 0 4px 8px rgba(0, 212, 170, 0.3);
  transform: translateY(-1px);
}

.icon-btn {
  background: none;
  border: 1px solid #e2e8f0;
  font-size: 12px;
  cursor: pointer;
  padding: 5px 8px;
  border-radius: 6px;
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

/* 卡片视图 */
.cards-container {
  padding: 0;
}

.cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(380px, 1fr));
  gap: 20px;
}

.daemonset-card {
  background: white;
  border-radius: 12px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.05);
  overflow: hidden;
  transition: all 0.3s;
  position: relative;
}

.daemonset-card:hover {
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.daemonset-card.card-selected {
  border: 2px solid #326ce5;
}

.card-checkbox {
  position: absolute;
  top: 12px;
  right: 12px;
  z-index: 1;
}

.card-header {
  padding: 16px 20px;
  border-bottom: 1px solid #f0f0f0;
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
  font-size: 16px;
  font-weight: 600;
  color: #2d3748;
  margin: 0;
  flex: 1;
}

.card-body {
  padding: 16px 20px;
}

.card-section {
  margin-bottom: 14px;
}

.card-section:last-child {
  margin-bottom: 0;
}

.section-label {
  font-size: 12px;
  font-weight: 500;
  color: #718096;
  margin-bottom: 6px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.card-section-row {
  display: flex;
  gap: 20px;
}

.card-meta-item {
  flex: 1;
}

.meta-label {
  font-size: 11px;
  color: #a0aec0;
  margin-bottom: 4px;
}

.meta-value {
  font-size: 13px;
  color: #4a5568;
}

.card-footer {
  padding: 12px 20px;
  background: #f8fafc;
  border-top: 1px solid #f0f0f0;
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.card-action-btn {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  background: white;
  color: #4a5568;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.card-action-btn:hover {
  border-color: #326ce5;
  color: #326ce5;
}

.card-action-btn.primary {
  background: #326ce5;
  border-color: #326ce5;
  color: white;
}

.card-action-btn.primary:hover {
  background: #2554c7;
}

.card-action-btn.danger {
  color: #e53e3e;
}

.card-action-btn.danger:hover {
  background: #fff5f5;
  border-color: #e53e3e;
}

/* Modal */
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

.modal-header h2, .modal-header h3 {
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

/* 表单 */
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

.form-hint {
  font-size: 12px;
  color: #6b7280;
  margin-top: 6px;
}

.required {
  color: #ef4444;
}

/* 创建表单 */
.modal-create-daemonset {
  max-width: 1000px;
}

.modal-create-daemonset .modal-body {
  max-height: 75vh;
  overflow-y: auto;
}

.form-section {
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  margin-bottom: 20px;
  overflow: hidden;
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
}

.section-icon {
  font-size: 20px;
}

.section-body {
  padding: 18px;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding-top: 16px;
  border-top: 1px solid #e2e8f0;
  margin-top: 16px;
}

/* 命名空间选择器 */
.namespace-selector {
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
}

.namespace-or {
  color: #718096;
  font-size: 13px;
}

.namespace-create {
  display: flex;
  gap: 8px;
  align-items: center;
  flex: 1;
}

.namespace-create .form-input {
  flex: 1;
  min-width: 150px;
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

/* 事件样式 */
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

/* 详情 */
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

/* 简单表格 */
.simple-table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 12px;
}

.simple-table th, .simple-table td {
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

.version-badge {
  display: inline-block;
  padding: 4px 10px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
}

.mono {
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 12px;
}

/* Pod 操作样式 */
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

/* 日志弹窗 */
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

.control-actions {
  display: flex;
  gap: 8px;
  align-items: flex-end;
}

.single-container {
  padding: 6px 10px;
  background: linear-gradient(135deg, #dbeafe 0%, #bfdbfe 100%);
  border-radius: 6px;
  font-size: 13px;
  color: #1e40af;
  font-weight: 500;
}

.logs-content-wrapper {
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
}

/* 日志高亮样式 */
.logs-content :deep(.log-error) {
  color: #f87171;
  font-weight: 500;
}

.logs-content :deep(.log-warn) {
  color: #fbbf24;
}

.logs-content :deep(.log-info) {
  color: #60a5fa;
}

.logs-content :deep(.log-debug) {
  color: #9ca3af;
}

.logs-content :deep(.log-timestamp) {
  color: #34d399;
  font-weight: 500;
}

.logs-content :deep(.log-separator) {
  color: #22d3ee;
  font-weight: 600;
  background: rgba(34, 211, 238, 0.1);
  display: block;
  padding: 4px 8px;
  margin: 8px 0;
  border-radius: 4px;
}

.logs-content :deep(.log-placeholder) {
  color: #64748b;
  font-style: italic;
}

/* 实时日志开关 */
.follow-toggle {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  font-size: 13px;
}

.follow-toggle input[type="checkbox"] {
  accent-color: #326ce5;
  cursor: pointer;
}

.streaming-indicator {
  color: #34d399;
  animation: pulse 1s infinite;
  font-size: 12px;
}

/* 空状态 */
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

/* 响应式 */
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
  .filter-buttons {
    flex-wrap: wrap;
  }
  .cards-grid {
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
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
  background: linear-gradient(135deg, #4f46e5 0%, #4338ca 100%);
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.3s ease;
  box-shadow: 0 2px 8px rgba(79, 70, 229, 0.3);
}

.load-template-btn:hover {
  background: linear-gradient(135deg, #4338ca 0%, #3730a3 100%);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(79, 70, 229, 0.4);
}

.copy-yaml-btn {
  padding: 10px 20px;
  background: linear-gradient(135deg, #059669 0%, #047857 100%);
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.3s ease;
  box-shadow: 0 2px 8px rgba(5, 150, 105, 0.3);
}

.copy-yaml-btn:hover {
  background: linear-gradient(135deg, #047857 0%, #065f46 100%);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(5, 150, 105, 0.4);
}

.clear-yaml-btn,
.reset-yaml-btn {
  padding: 10px 20px;
  background: linear-gradient(135deg, #64748b 0%, #475569 100%);
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.3s ease;
  box-shadow: 0 2px 8px rgba(100, 116, 139, 0.3);
}

.clear-yaml-btn:hover,
.reset-yaml-btn:hover {
  background: linear-gradient(135deg, #475569 0%, #334155 100%);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(100, 116, 139, 0.4);
}

.yaml-editor {
  width: 100%;
  min-height: 400px;
  padding: 20px;
  font-family: 'JetBrains Mono', 'Fira Code', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
  border: 2px solid #334155;
  border-radius: 12px;
  background: #1e293b;
  color: #e2e8f0;
  resize: vertical;
  tab-size: 2;
  transition: all 0.2s;
}

.yaml-editor:focus {
  outline: none;
  border-color: #4f46e5;
  box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.15);
}

.yaml-error {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: #fee2e2;
  border: 1px solid #fca5a5;
  border-radius: 8px;
  color: #dc2626;
  font-size: 14px;
}

.error-icon {
  font-size: 18px;
}

.yaml-editor-footer {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
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

/* 资源监听浮窗 */
.resource-watcher-panel {
  position: fixed;
  bottom: 24px;
  right: 24px;
  width: 380px;
  background: #1a1b26;
  border: 1px solid #2f3549;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0,0,0,0.4);
  z-index: 1000;
  overflow: hidden;
  color: #c0caf5;
}
.watcher-header {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  background: #24283b;
  border-bottom: 1px solid #2f3549;
}
.watcher-title { flex: 1; font-weight: 600; font-size: 14px; }
.watcher-elapsed { font-size: 12px; color: #565f89; margin-right: 8px; font-family: monospace; }
.watcher-close { background: none; border: none; color: #565f89; font-size: 18px; cursor: pointer; padding: 0 4px; }
.watcher-close:hover { color: #f7768e; }
.watcher-body { padding: 12px 16px; }
.watcher-progress { height: 4px; background: #2f3549; border-radius: 2px; overflow: hidden; margin-bottom: 8px; }
.watcher-progress-bar { height: 100%; border-radius: 2px; transition: width 0.5s ease; }
.watcher-phase { font-size: 13px; font-weight: 500; margin-bottom: 8px; }
.watcher-events { max-height: 160px; overflow-y: auto; }
.watcher-event { display: flex; gap: 6px; font-size: 12px; padding: 3px 0; border-bottom: 1px solid #2f3549; }
.watcher-event.warning { color: #e0af68; }
.ev-type { flex-shrink: 0; }
.ev-reason { flex-shrink: 0; font-weight: 500; color: #7aa2f7; }
.ev-msg { color: #9aa5ce; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.watcher-slide-enter-active, .watcher-slide-leave-active { transition: all 0.3s ease; }
.watcher-slide-enter-from, .watcher-slide-leave-to { opacity: 0; transform: translateY(20px); }

.current-image-display {
  padding: 10px 12px;
  background: #f0f4f8;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 13px;
  color: #2d3748;
  word-break: break-all;
  line-height: 1.5;
}
.current-image-text { opacity: 0.85; }
/* ===== DDP Describe Drawer ===== */
.ddp-overlay { position: fixed; inset: 0; z-index: 9000; background: rgba(15,23,42,.45); backdrop-filter: blur(2px); }
.ddp-panel { position: absolute; right: 0; top: 0; height: 100%; width: 880px; max-width: 92vw; display: flex; flex-direction: column; background: #fff; box-shadow: -8px 0 40px rgba(0,0,0,.18); }
.ddp-hd { display: flex; align-items: flex-start; justify-content: space-between; padding: 24px 28px 18px; background: linear-gradient(135deg, #0f172a 0%, #1e3a5f 100%); color: #fff; flex-shrink: 0; }
.ddp-hd-body { flex: 1; min-width: 0; } .ddp-hd-row1 { display: flex; align-items: center; gap: 8px; margin-bottom: 6px; }
.ddp-status-pill { display: inline-flex; align-items: center; gap: 5px; font-size: 11px; font-weight: 700; padding: 3px 10px; border-radius: 12px; text-transform: uppercase; }
.ddp-st-ok { background: rgba(16,185,129,.2); color: #6ee7b7; } .ddp-st-prog { background: rgba(250,204,21,.2); color: #fde68a; } .ddp-st-err { background: rgba(239,68,68,.2); color: #fca5a5; }
.ddp-status-led { width: 7px; height: 7px; border-radius: 50%; background: currentColor; }
.ddp-ns-pill { background: rgba(255,255,255,.12); padding: 2px 9px; border-radius: 10px; font-size: 11px; font-weight: 600; }
.ddp-name { font-size: 20px; font-weight: 800; margin: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.ddp-hd-meta { margin-top: 6px; font-size: 12px; color: rgba(255,255,255,.7); display: flex; gap: 6px; align-items: center; }
.ddp-dot { color: rgba(255,255,255,.35); }
.ddp-close-btn { background: rgba(255,255,255,.1); border: none; color: #fff; cursor: pointer; border-radius: 8px; padding: 8px; }
.ddp-loading { flex: 1; display: flex; align-items: center; justify-content: center; gap: 10px; color: #64748b; font-size: 14px; }
.ddp-spin { width: 22px; height: 22px; border-radius: 50%; border: 3px solid #e2e8f0; border-top-color: #6366f1; animation: ddp-spin .7s linear infinite; }
@keyframes ddp-spin { to { transform: rotate(360deg); } }
.ddp-body { flex: 1; overflow-y: auto; } .ddp-sec { border-bottom: 1px solid #f1f5f9; }
.ddp-sec-title { display: flex; align-items: center; gap: 8px; padding: 12px 20px; font-size: 13px; font-weight: 700; color: #1e293b; }
.ddp-sec-title-clickable { cursor: pointer; user-select: none; } .ddp-sec-title-clickable:hover { background: rgba(99,102,241,.04); }
.ddp-si { font-size: 14px; } .ddp-si-blue { color: #3b82f6; } .ddp-si-green { color: #10b981; } .ddp-si-purple { color: #8b5cf6; } .ddp-si-orange { color: #f59e0b; } .ddp-si-teal { color: #06b6d4; }
.ddp-cnt-badge { font-size: 11px; background: #e0e7ff; color: #4338ca; border-radius: 8px; padding: 1px 7px; font-weight: 700; }
.ddp-fold-arrow { margin-left: auto; color: #94a3b8; font-size: 11px; }
.ddp-kv-grid { display: grid; grid-template-columns: 1fr 1fr; padding: 4px 0; }
.ddp-kv { display: flex; align-items: baseline; gap: 8px; padding: 6px 20px; border-bottom: 1px solid #fafafa; }
.ddp-k { font-size: 12px; color: #64748b; min-width: 85px; flex-shrink: 0; } .ddp-v { font-size: 12px; color: #1e293b; word-break: break-all; } .ddp-v.mono { font-family: monospace; }
.ddp-conds { padding: 8px 20px 12px; display: flex; flex-wrap: wrap; gap: 8px; }
.ddp-cond-chip { display: inline-flex; align-items: center; gap: 5px; padding: 4px 10px; border-radius: 6px; font-size: 12px; font-weight: 600; }
.ddp-cond-chip.cond-ok { background: #ecfdf5; color: #065f46; } .ddp-cond-chip.cond-fail { background: #fef2f2; color: #991b1b; }
.ddp-cond-chip .cond-led { width: 6px; height: 6px; border-radius: 50%; } .ddp-cond-chip.cond-ok .cond-led { background: #10b981; } .ddp-cond-chip.cond-fail .cond-led { background: #ef4444; }
.ddp-cond-chip .cond-name { font-weight: 700; } .ddp-cond-chip .cond-reason { font-weight: 400; color: #64748b; font-size: 11px; }
.ddp-tags { display: flex; flex-wrap: wrap; gap: 6px; } .ddp-tags-pad { padding: 8px 20px 12px; }
.ddp-label-tag { display: inline-flex; align-items: center; gap: 2px; background: #f1f5f9; border-radius: 4px; padding: 3px 8px; font-size: 11px; font-family: monospace; color: #334155; } .ddp-label-tag b { color: #4338ca; }
.ddp-anno-list { padding: 0 20px 12px; display: flex; flex-direction: column; }
.ddp-anno-row { display: grid; grid-template-columns: 260px 1fr; gap: 8px; padding: 5px 0; border-bottom: 1px solid #f8fafc; font-size: 12px; }
.ddp-anno-k { font-family: monospace; color: #4338ca; font-weight: 600; word-break: break-all; } .ddp-anno-v { font-family: monospace; color: #475569; word-break: break-all; white-space: pre-wrap; max-height: 80px; overflow-y: auto; }
.ddp-ctr-list { padding: 4px 12px 12px; display: flex; flex-direction: column; gap: 8px; }
.ddp-ctr-card { border: 1px solid #e2e8f0; border-radius: 10px; overflow: hidden; border-left: 3px solid #6366f1; }
.ddp-ctr-hd { padding: 10px 14px 4px; } .ddp-ctr-name { font-size: 13px; font-weight: 700; color: #1e293b; }
.ddp-ctr-img { padding: 0 14px 8px; font-size: 12px; font-family: monospace; color: #6366f1; }
.ddp-ctr-detail { padding: 0 14px 10px; display: flex; flex-direction: column; gap: 6px; }
.ddp-ctr-row { display: flex; flex-wrap: wrap; align-items: flex-start; gap: 6px; font-size: 12px; }
.ddp-rk { flex-shrink: 0; font-weight: 600; color: #475569; min-width: 60px; }
.ddp-port-chip { background: #ede9fe; color: #5b21b6; border-radius: 4px; padding: 2px 7px; font-size: 11px; font-weight: 600; }
.ddp-sec-events { border-bottom: none; } .ddp-empty { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; color: #94a3b8; }
.ddp-slide-enter-active { transition: transform 0.32s cubic-bezier(0.4,0,0.2,1); } .ddp-slide-leave-active { transition: transform 0.22s cubic-bezier(0.4,0,1,1); }
.ddp-slide-enter-from, .ddp-slide-leave-to { transform: translateX(100%); }
.pdp-ev-refresh-btn { margin-left: auto; background: none; border: 1px solid #e2e8f0; border-radius: 6px; padding: 4px 8px; cursor: pointer; color: #64748b; display: flex; align-items: center; } .pdp-ev-refresh-btn:hover { background: #f1f5f9; color: #4338ca; }
.pdp-ev-empty { padding: 20px; text-align: center; color: #94a3b8; font-size: 13px; display: flex; align-items: center; justify-content: center; gap: 8px; }
.pdp-ev-spin { width: 16px; height: 16px; border: 2px solid #e2e8f0; border-top-color: #6366f1; border-radius: 50%; animation: ddp-spin .7s linear infinite; }
.pdp-ev-list { padding: 8px 16px 16px; display: flex; flex-direction: column; gap: 6px; }
.pdp-ev-item { padding: 8px 12px; border-radius: 8px; border-left: 3px solid #10b981; background: #f8fafc; } .pdp-ev-item.ev-warn { border-left-color: #f59e0b; background: #fffbeb; }
.pdp-ev-hd { display: flex; align-items: center; gap: 6px; font-size: 12px; }
.pdp-ev-type-dot { width: 7px; height: 7px; border-radius: 50%; } .pdp-ev-type-dot.ev-dot-norm { background: #10b981; } .pdp-ev-type-dot.ev-dot-warn { background: #f59e0b; }
.pdp-ev-reason { font-weight: 700; color: #1e293b; } .pdp-ev-count { font-size: 11px; color: #64748b; } .pdp-ev-age { margin-left: auto; font-size: 11px; color: #94a3b8; }
.pdp-ev-msg { font-size: 12px; color: #475569; margin-top: 4px; } .pdp-ev-src { font-size: 11px; color: #94a3b8; margin-top: 2px; }
.pdp-cmd-blk { background: #1e293b; border-radius: 6px; padding: 6px 10px; font-family: monospace; font-size: 11px; color: #e2e8f0; flex: 1; min-width: 0; }
.pdp-cmd-line { white-space: pre-wrap; word-break: break-all; } .pdp-cmd-label { color: #94a3b8; margin-right: 8px; }
.pdp-res-lines { display: flex; flex-direction: column; gap: 3px; } .pdp-res-line { display: flex; align-items: center; gap: 6px; flex-wrap: wrap; }
.pdp-res-lbl { font-size: 10px; padding: 1px 5px; border-radius: 3px; font-weight: 600; } .pdp-res-lbl.req { background: #dbeafe; color: #1d4ed8; } .pdp-res-lbl.lim { background: #fef3c7; color: #92400e; }
.pdp-res-tag { font-size: 11px; color: #334155; font-family: monospace; }
.pdp-mnt-tbl { display: flex; flex-direction: column; gap: 2px; flex: 1; } .pdp-mnt-row { display: flex; gap: 8px; font-size: 11px; font-family: monospace; }
.pdp-mn { color: #4338ca; font-weight: 600; } .pdp-mp { color: #475569; } .pdp-mro { color: #dc2626; font-size: 10px; font-weight: 700; }
</style>
