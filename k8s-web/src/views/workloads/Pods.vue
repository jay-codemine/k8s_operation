<template>
  <div class="resource-view">
    <div class="view-header">
      <h1>Pod管理</h1>
      <p>Kubernetes集群中的Pod列表</p>
    </div>

    <div class="action-bar">
      <div class="search-box">
        <input
          type="text"
          v-model="searchQuery"
          placeholder="模糊搜索 Pod 名称..."
          @input="onSearchInput"
          @keyup.enter="refresh"
        />
      </div>

      <div class="filter-buttons">
        <button class="btn btn-filter" :class="{ active: statusFilter === 'all' }"
                @click="setStatusFilter('all')">
          全部
        </button>
        <button class="btn btn-filter" :class="{ active: statusFilter === 'Running' }"
                @click="setStatusFilter('Running')">
          Running
        </button>
        <button class="btn btn-filter" :class="{ active: statusFilter === 'Pending' }"
                @click="setStatusFilter('Pending')">
          Pending
        </button>
        <button class="btn btn-filter" :class="{ active: statusFilter === 'Succeeded' }"
                @click="setStatusFilter('Succeeded')">
          Succeeded
        </button>
        <button class="btn btn-filter" :class="{ active: statusFilter === 'CrashLoopBackOff' }"
                @click="setStatusFilter('CrashLoopBackOff')">
          CrashLoopBackOff
        </button>
        <button class="btn btn-filter" :class="{ active: statusFilter === 'ImagePullBackOff' }"
                @click="setStatusFilter('ImagePullBackOff')">
          ImagePullBackOff
        </button>
        <button class="btn btn-filter" :class="{ active: statusFilter === 'Failed' }"
                @click="setStatusFilter('Failed')">
          Failed
        </button>
      </div>

      <div class="filter-dropdown">
        <select v-model="namespaceFilter" @change="onNamespaceChange">
          <option value="">所有名称空间</option>
          <option v-for="ns in namespaces" :key="ns" :value="ns">
            {{ ns }}
          </option>
        </select>
      </div>

      <div class="action-buttons">
        <!-- 批量操作按钮 -->
        <button 
          v-if="canOperate && !batchMode" 
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
        <button v-if="canOperate" class="btn btn-primary" @click="openCreate">
          创建Pod
        </button>
        <button class="btn btn-secondary" @click="refresh" :disabled="loading">
          {{ loading ? '加载中...' : '🔄 刷新' }}
        </button>
      </div>
    </div>

    <div v-if="errorMsg" class="error-box">
      {{ errorMsg }}
    </div>

    <!-- 批量操作浮动栏 -->
    <div v-if="batchMode && selectedPods.length > 0" class="batch-action-bar">
      <div class="batch-info">
        <span class="batch-count">已选择 {{ selectedPods.length }} 个 Pod</span>
        <button class="batch-clear" @click="clearSelection">清空选择</button>
      </div>
      <div class="batch-actions">
        <button class="batch-btn" @click="openBatchEvictPreview" title="批量驱逐">
          ⚡ 批量驱逐
        </button>
        <button class="batch-btn danger" @click="openBatchDeletePreview(false)" title="批量优雅删除">
          🗑️ 批量删除
        </button>
        <button class="batch-btn danger" @click="openBatchDeletePreview(true)" title="批量强制删除">
          💥 批量强删
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
              :indeterminate="isPartialSelected"
              @change="toggleSelectAll"
              title="全选/取消全选"
              ref="selectAllCheckbox"
            />
          </th>
          <th>状态</th>
          <th>名称</th>
          <th>名称空间</th>
          <th>节点</th>
          <th>镜像</th>
          <th style="text-align: center;">重启次数</th>
          <th>关联服务</th>
          <th>创建时间</th>
          <th>操作</th>
        </tr>
        </thead>

        <tbody>
        <tr 
          v-for="pod in paginatedPods" 
          :key="pod.uid || pod.name" 
          :class="{ 'row-selected': isPodSelected(pod), 'row-clickable': batchMode }"
          @click="batchMode && handleRowClick(pod, $event)"
        >
          <td v-if="batchMode">
            <input 
              type="checkbox" 
              :checked="isPodSelected(pod)" 
              @change="togglePodSelection(pod)"
              @click.stop
            />
          </td>
          <td>
              <div class="status-cell">
                <span class="status-indicator" :class="(pod.status || 'unknown').toLowerCase()">
                  <span class="status-dot"></span>
                  {{ pod.status || 'Unknown' }}
                </span>
                <!-- 状态详细信息提示 -->
                <div class="status-tooltip">
                  <div>Pod IP: {{ pod.podIP || '-' }}</div>
                  <div>Host IP: {{ pod.hostIP || '-' }}</div>
                  <div>重启次数: {{ pod.restartCount ?? 0 }}</div>
                  <div v-if="pod.statusReason">原因: {{ pod.statusReason }}</div>
                </div>
              </div>
          </td>

          <td>
            <div class="pod-name">
              <a class="pod-name-link" :title="pod.name" @click="openDetail(pod)">{{ pod.name }}</a>
            </div>
          </td>

          <td>
            <span class="namespace-badge">{{ pod.namespace }}</span>
          </td>

          <td>{{ pod.node || '-' }}</td>
          <td class="mono">{{ pod.image || '-' }}</td>

          <td style="text-align: center;">{{ pod.restartCount ?? 0 }}</td>

          <td>
            <span v-if="pod.service" class="service-tag">{{ pod.service }}</span>
            <span v-else class="no-service">无</span>
          </td>

          <td class="mono">{{ pod.createdAt || '-' }}</td>

          <td>
            <div class="action-icons">
              <!-- 详情按钮（直接显示） -->
              <button class="action-btn primary" title="查看详情" @click="openDetail(pod)">
                📋 详情
              </button>
              <button class="icon-btn" title="查看日志" @click="openLogs(pod)">
                📄 日志
              </button>
              <button class="action-btn terminal" title="容器终端" @click="openTerminal(pod)">
                >_ 终端
              </button>

              <!-- 更多按钮 -->
              <div class="more-btn" ref="moreBtn">
                <button class="icon-btn" @click="toggleMoreOptions(pod, $event)">
                  ⋮ 更多
                </button>

                <!-- 更多菜单 -->
                <div v-if="showMoreOptions && selectedPod === pod" class="more-menu" :style="menuStyle">
                  <button class="menu-item" title="查看事件" @click="openEvents(pod)">
                    <span class="menu-icon">📡</span>
                    <span>查看事件</span>
                  </button>
                  <button v-if="canOperate" class="menu-item" title="更新镜像" @click="openPatchImage(pod)">
                    <span class="menu-icon">🔧</span>
                    <span>更新镜像</span>
                  </button>
                  <button v-if="canOperate" class="menu-item" title="驱逐Pod" @click="evictPod(pod)">
                    <span class="menu-icon">⚠️</span>
                    <span>驱逐Pod</span>
                  </button>
                  <button class="menu-item" title="查看/编辑YAML" @click="openYamlPreview(pod)">
                    <span class="menu-icon">📝</span>
                    <span>查看/编辑YAML</span>
                  </button>
                  <div v-if="canOperate" class="menu-divider"></div>
                  <button v-if="canOperate" class="menu-item danger" title="优雅删除" @click="deletePod(pod, false)">
                    <span class="menu-icon">🗑️</span>
                    <span>优雅删除</span>
                  </button>
                  <button v-if="canOperate" class="menu-item danger" title="强制删除" @click="deletePod(pod, true)">
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

      <div v-if="!loading && paginatedPods.length === 0" class="empty-state">
        <div class="empty-icon">🔍</div>
        <div class="empty-text">没有找到匹配的Pod</div>
      </div>

      <!-- 分页组件 -->
      <Pagination
        v-if="total > 0"
        v-model:currentPage="currentPage"
        :totalItems="total"
        :itemsPerPage="itemsPerPage"
      />
    </div>

    <!-- 卡片视图 -->
    <div v-if="viewMode === 'card'" class="cards-container">
      <div v-if="!loading && paginatedPods.length === 0" class="empty-state">
        <div class="empty-icon">🔍</div>
        <div class="empty-text">没有找到匹配的Pod</div>
      </div>
      
      <div class="cards-grid">
        <div 
          v-for="pod in paginatedPods" 
          :key="pod.uid || pod.name" 
          class="pod-card" 
          :class="{ 'card-selected': isPodSelected(pod), 'card-clickable': batchMode }"
          @click="batchMode && handleCardClick(pod, $event)"
        >
          <!-- 批量选择复选框 -->
          <div v-if="batchMode" class="card-checkbox" @click.stop>
            <input 
              type="checkbox" 
              :checked="isPodSelected(pod)" 
              @change="togglePodSelection(pod)"
            />
          </div>
          
          <!-- 卡片头部 -->
          <div class="card-header">
            <div class="card-title-row">
              <span class="card-icon">📦</span>
              <h3 class="card-title">
                <a class="card-title-link" :title="pod.name" @click="openDetail(pod)">{{ pod.name }}</a>
              </h3>
              <span class="status-indicator" :class="(pod.status || 'unknown').toLowerCase()">
                {{ pod.status || 'Unknown' }}
              </span>
            </div>
            <span class="namespace-badge">{{ pod.namespace }}</span>
          </div>

          <!-- 卡片主体 -->
          <div class="card-body">
            <!-- 节点信息 -->
            <div class="card-section">
              <div class="section-label">节点</div>
              <div class="meta-value">{{ pod.node || '-' }}</div>
            </div>

            <!-- 镜像 -->
            <div class="card-section">
              <div class="section-label">镜像</div>
              <div class="meta-value mono">{{ pod.image || '-' }}</div>
            </div>

            <!-- 重启次数和关联服务 -->
            <div class="card-section card-section-row">
              <div class="card-meta-item">
                <div class="meta-label">重启次数</div>
                <div class="meta-value">{{ pod.restartCount ?? 0 }}</div>
              </div>
              <div class="card-meta-item">
                <div class="meta-label">关联服务</div>
                <div class="meta-value">
                  <span v-if="pod.service" class="service-tag">{{ pod.service }}</span>
                  <span v-else class="no-service">无</span>
                </div>
              </div>
            </div>

            <!-- 创建时间 -->
            <div class="card-section">
              <div class="section-label">创建时间</div>
              <div class="meta-value mono">{{ pod.createdAt || '-' }}</div>
            </div>

            <!-- 资源消耗 -->
            <div class="card-section">
              <div class="section-label">资源使用</div>
              <!-- 有 metrics 数据时显示具体值 -->
              <div v-if="pod.metrics" class="metrics-summary">
                <div class="metric-item">
                  <span class="metric-icon">⚡</span>
                  <span class="metric-label">CPU:</span>
                  <span class="metric-value">{{ pod.metrics.total_cpu }}</span>
                </div>
                <div class="metric-item">
                  <span class="metric-icon">💾</span>
                  <span class="metric-label">内存:</span>
                  <span class="metric-value">{{ pod.metrics.total_memory }}</span>
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
                  <span v-if="['Succeeded', 'Completed', 'Failed'].includes(pod.status)">容器已停止</span>
                  <span v-else-if="pod.status === 'Pending'">容器未启动</span>
                  <span v-else>暂无数据</span>
                </div>
              </div>
              <!-- 容器级别详情 -->
              <div class="container-metrics" v-if="pod.metrics && pod.metrics.containers && pod.metrics.containers.length > 1">
                <div class="container-metrics-item" v-for="c in pod.metrics.containers" :key="c.name">
                  <div class="container-metrics-name">{{ c.name }}</div>
                  <div class="container-metrics-values">
                    <span class="container-cpu">CPU: {{ c.cpu }}</span>
                    <span class="container-memory">内存: {{ c.memory }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- 卡片底部按钮 -->
          <div class="card-footer">
            <button class="card-action-btn" @click="openLogs(pod)" title="查看日志">
              📄 日志
            </button>
            <button class="card-action-btn" @click="openTerminal(pod)" title="容器终端">
              >_ 终端
            </button>
            <button class="card-action-btn" @click="openDetail(pod)" title="查看详情">
              📋 详情
            </button>
            <button class="card-action-btn" @click="openEvents(pod)" title="查看事件">
              📡 事件
            </button>
            <button class="card-action-btn" @click="openPatchImage(pod)" title="更新镜像">
              🔧 更新
            </button>
            <button class="card-action-btn" @click="openYamlPreview(pod)" title="查看/编辑YAML">
              📝 YAML
            </button>
            <button class="card-action-btn danger" @click="deletePod(pod, false)" title="删除">
              🗑️ 删除
            </button>
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

    <!-- 删除确认弹窗 -->
    <div v-if="showDeleteModal" class="modal-overlay" @click.self="showDeleteModal = false">
      <div class="modal-content" style="max-width: 480px;">
        <div class="modal-header">
          <h3>{{ isForceDelete ? '⚠️ 强制删除 Pod' : '🗑️ 删除 Pod' }}</h3>
          <button class="close-btn" @click="showDeleteModal = false">×</button>
        </div>
        <div class="modal-body">
          <p v-if="podToDelete" style="margin-bottom: 16px;">
            确认{{ isForceDelete ? '强制' : '' }}删除 Pod：
          </p>
          <div v-if="podToDelete" style="background: #f7fafc; padding: 12px; border-radius: 8px; margin-bottom: 16px;">
            <div><strong>名称空间：</strong>{{ podToDelete.namespace }}</div>
            <div><strong>名称：</strong>{{ podToDelete.name }}</div>
            <div><strong>状态：</strong>{{ podToDelete.status || '-' }}</div>
            <div><strong>节点：</strong>{{ podToDelete.node || '-' }}</div>
          </div>
          <div v-if="!isForceDelete" style="background: #fff3cd; padding: 12px; border-radius: 8px; margin-bottom: 12px; border-left: 3px solid #ffc107;">
            <div style="color: #856404; font-size: 13px; line-height: 1.6;">
              ⚠️ 优雅删除：Pod 将在优雅终止期内被删除，容器有时间处理清理工作。
            </div>
          </div>
          <div v-if="isForceDelete" style="background: #f8d7da; padding: 12px; border-radius: 8px; margin-bottom: 12px; border-left: 3px solid #dc3545;">
            <div style="color: #721c24; font-size: 13px; line-height: 1.6; margin-bottom: 8px;">
              <strong>⚠️ 强制删除警告：</strong>
            </div>
            <ul style="margin: 0; padding-left: 20px; color: #721c24; font-size: 13px; line-height: 1.6;">
              <li>立即终止 Pod，不等待优雅终止期</li>
              <li>可能导致数据丢失或事务中断</li>
              <li>绕过 Pod Disruption Budget (PDB) 保护</li>
              <li>仅在 Pod 卡住无法正常删除时使用</li>
            </ul>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showDeleteModal = false" :disabled="deleting">
            取消
          </button>
          <button class="btn btn-danger" @click="confirmDelete" :disabled="deleting">
            {{ deleting ? '删除中...' : (isForceDelete ? '强制删除' : '确认删除') }}
          </button>
        </div>
      </div>
    </div>

    <!-- ========== DDP Describe 抽屉 ========== -->
    <Teleport to="body">
      <div v-if="showDetailModal" class="ddp-overlay" @click="closeDetail">
        <Transition name="ddp-slide">
          <div v-if="showDetailModal" class="ddp-panel" @click.stop>
            <!-- Header -->
            <div class="ddp-hd">
              <div class="ddp-hd-top">
                <span class="ddp-hd-icon">📦</span>
                <h2 class="ddp-hd-title">{{ detailData?.metadata?.name || selectedPod?.name }}</h2>
                <span class="ddp-hd-ns">{{ detailData?.metadata?.namespace || selectedPod?.namespace }}</span>
                <span v-if="detailData" class="ddp-hd-status" :class="ddpStatusClass(detailData)">
                  <i class="ddp-led"></i>{{ ddpStatusText(detailData) }}
                </span>
              </div>
              <button class="ddp-close" @click="closeDetail">&times;</button>
            </div>
            <!-- Body -->
            <div class="ddp-body">
              <div v-if="loadingDetail" class="ddp-loading">
                <div class="loading-spinner"></div><span>加载中…</span>
              </div>
              <template v-else-if="detailData">
                <!-- 基本信息 -->
                <section class="ddp-sec">
                  <h3 class="ddp-sec-title">基本信息</h3>
                  <div class="ddp-kv-grid">
                    <div class="ddp-kv"><span class="ddp-k">名称</span><span class="ddp-v">{{ detailData.metadata?.name }}</span></div>
                    <div class="ddp-kv"><span class="ddp-k">命名空间</span><span class="ddp-v">{{ detailData.metadata?.namespace }}</span></div>
                    <div class="ddp-kv"><span class="ddp-k">UID</span><span class="ddp-v mono">{{ detailData.metadata?.uid || '-' }}</span></div>
                    <div class="ddp-kv"><span class="ddp-k">创建时间</span><span class="ddp-v">{{ detailData.metadata?.creationTimestamp || '-' }}</span></div>
                    <div class="ddp-kv"><span class="ddp-k">Node</span><span class="ddp-v">{{ detailData.spec?.nodeName || '-' }}</span></div>
                    <div class="ddp-kv"><span class="ddp-k">Pod IP</span><span class="ddp-v mono">{{ detailData.status?.podIP || '-' }}</span></div>
                    <div class="ddp-kv"><span class="ddp-k">Host IP</span><span class="ddp-v mono">{{ detailData.status?.hostIP || '-' }}</span></div>
                    <div class="ddp-kv"><span class="ddp-k">Phase</span><span class="ddp-v">{{ detailData.status?.phase || '-' }}</span></div>
                    <div class="ddp-kv"><span class="ddp-k">QoS Class</span><span class="ddp-v">{{ detailData.status?.qosClass || '-' }}</span></div>
                    <div class="ddp-kv"><span class="ddp-k">Restart Policy</span><span class="ddp-v">{{ detailData.spec?.restartPolicy || '-' }}</span></div>
                    <div class="ddp-kv"><span class="ddp-k">Service Account</span><span class="ddp-v">{{ detailData.spec?.serviceAccountName || '-' }}</span></div>
                    <div class="ddp-kv" v-if="detailData.spec?.priorityClassName"><span class="ddp-k">Priority Class</span><span class="ddp-v">{{ detailData.spec.priorityClassName }}</span></div>
                  </div>
                </section>

                <!-- Conditions -->
                <section class="ddp-sec" v-if="detailData.status?.conditions?.length">
                  <h3 class="ddp-sec-title">Conditions</h3>
                  <div class="ddp-chip-list">
                    <span class="ddp-chip" :class="cond.status === 'True' ? '' : 'anno'" v-for="cond in detailData.status.conditions" :key="cond.type" :title="cond.message || ''">
                      {{ cond.type }}: {{ cond.status }}
                    </span>
                  </div>
                </section>

                <!-- Containers -->
                <section class="ddp-sec" v-if="detailData.spec?.containers?.length">
                  <h3 class="ddp-sec-title">Containers ({{ detailData.spec.containers.length }})</h3>
                  <div v-for="c in detailData.spec.containers" :key="c.name" class="ddp-container-card">
                    <div class="ddp-container-name">
                      📦 {{ c.name }}
                      <span v-if="ddpContainerStatus(detailData, c.name)" class="ddp-cstatus" :class="ddpContainerStatus(detailData, c.name).ready ? 'ready' : 'not-ready'">
                        {{ ddpContainerStatus(detailData, c.name).ready ? '✔ Ready' : '✘ Not Ready' }}
                      </span>
                    </div>
                    <div class="ddp-kv-grid">
                      <div class="ddp-kv"><span class="ddp-k">镜像</span><span class="ddp-v mono">{{ c.image }}</span></div>
                      <div class="ddp-kv" v-if="c.command"><span class="ddp-k">Command</span><span class="ddp-v mono">{{ c.command.join(' ') }}</span></div>
                      <div class="ddp-kv" v-if="c.args"><span class="ddp-k">Args</span><span class="ddp-v mono">{{ c.args.join(' ') }}</span></div>
                      <div class="ddp-kv" v-if="c.ports?.length"><span class="ddp-k">Ports</span><span class="ddp-v">{{ c.ports.map(p => p.containerPort + '/' + (p.protocol || 'TCP')).join(', ') }}</span></div>
                      <div class="ddp-kv" v-if="c.resources?.requests"><span class="ddp-k">Requests</span><span class="ddp-v">CPU: {{ c.resources.requests.cpu || '-' }} / Mem: {{ c.resources.requests.memory || '-' }}</span></div>
                      <div class="ddp-kv" v-if="c.resources?.limits"><span class="ddp-k">Limits</span><span class="ddp-v">CPU: {{ c.resources.limits.cpu || '-' }} / Mem: {{ c.resources.limits.memory || '-' }}</span></div>
                      <div class="ddp-kv" v-if="ddpContainerStatus(detailData, c.name)?.restartCount != null"><span class="ddp-k">Restarts</span><span class="ddp-v">{{ ddpContainerStatus(detailData, c.name).restartCount }}</span></div>
                    </div>
                    <!-- Env -->
                    <div v-if="c.env?.length" style="margin-top:8px;">
                      <div class="ddp-k" style="margin-bottom:4px">环境变量 ({{ c.env.length }})</div>
                      <div class="ddp-chip-list">
                        <span class="ddp-chip" v-for="e in c.env.slice(0, 10)" :key="e.name" :title="e.value || (e.valueFrom ? 'from ref' : '')">{{ e.name }}={{ e.value || '(ref)' }}</span>
                        <span v-if="c.env.length > 10" class="ddp-chip anno">+{{ c.env.length - 10 }} more</span>
                      </div>
                    </div>
                  </div>
                </section>

                <!-- Init Containers -->
                <section class="ddp-sec" v-if="detailData.spec?.initContainers?.length">
                  <h3 class="ddp-sec-title">Init Containers ({{ detailData.spec.initContainers.length }})</h3>
                  <div v-for="c in detailData.spec.initContainers" :key="c.name" class="ddp-container-card">
                    <div class="ddp-container-name">⚙️ {{ c.name }}</div>
                    <div class="ddp-kv-grid">
                      <div class="ddp-kv"><span class="ddp-k">镜像</span><span class="ddp-v mono">{{ c.image }}</span></div>
                      <div class="ddp-kv" v-if="c.command"><span class="ddp-k">Command</span><span class="ddp-v mono">{{ c.command.join(' ') }}</span></div>
                    </div>
                  </div>
                </section>

                <!-- Node Selectors -->
                <section class="ddp-sec" v-if="detailData.spec?.nodeSelector && Object.keys(detailData.spec.nodeSelector).length">
                  <h3 class="ddp-sec-title">Node Selector</h3>
                  <div class="ddp-chip-list">
                    <span class="ddp-chip" v-for="(v, k) in detailData.spec.nodeSelector" :key="k">{{ k }}={{ v }}</span>
                  </div>
                </section>

                <!-- Tolerations -->
                <section class="ddp-sec" v-if="detailData.spec?.tolerations?.length">
                  <h3 class="ddp-sec-title">Tolerations ({{ detailData.spec.tolerations.length }})</h3>
                  <div class="ddp-chip-list">
                    <span class="ddp-chip" v-for="(t, idx) in detailData.spec.tolerations" :key="idx">{{ t.key || '*' }}{{ t.operator === 'Exists' ? ' (Exists)' : '=' + (t.value || '') }} : {{ t.effect || 'All' }}</span>
                  </div>
                </section>

                <!-- Labels -->
                <section class="ddp-sec" v-if="detailData.metadata?.labels && Object.keys(detailData.metadata.labels).length">
                  <h3 class="ddp-sec-title">Labels</h3>
                  <div class="ddp-chip-list">
                    <span class="ddp-chip" v-for="(v, k) in detailData.metadata.labels" :key="k">{{ k }}={{ v }}</span>
                  </div>
                </section>

                <!-- Annotations -->
                <section class="ddp-sec" v-if="detailData.metadata?.annotations && Object.keys(detailData.metadata.annotations).length">
                  <h3 class="ddp-sec-title ddp-fold-title" @click="ddpShowAnno = !ddpShowAnno">
                    Annotations ({{ Object.keys(detailData.metadata.annotations).length }})
                    <span class="ddp-fold-icon">{{ ddpShowAnno ? '▾' : '▸' }}</span>
                  </h3>
                  <div v-if="ddpShowAnno" class="ddp-chip-list">
                    <span class="ddp-chip anno" v-for="(v, k) in detailData.metadata.annotations" :key="k" :title="v">{{ k }}={{ v?.substring?.(0, 60) }}</span>
                  </div>
                </section>

                <!-- Volumes -->
                <section class="ddp-sec" v-if="detailData.spec?.volumes?.length">
                  <h3 class="ddp-sec-title ddp-fold-title" @click="ddpShowVols = !ddpShowVols">
                    Volumes ({{ detailData.spec.volumes.length }})
                    <span class="ddp-fold-icon">{{ ddpShowVols ? '▾' : '▸' }}</span>
                  </h3>
                  <div v-if="ddpShowVols" class="ddp-chip-list">
                    <span class="ddp-chip" v-for="vol in detailData.spec.volumes" :key="vol.name" :title="JSON.stringify(vol)">{{ vol.name }}</span>
                  </div>
                </section>

                <!-- Events -->
                <section class="ddp-sec">
                  <h3 class="ddp-sec-title">Events</h3>
                  <div v-if="ddpEventsLoading" class="ddp-loading"><span>加载事件...</span></div>
                  <div v-else-if="ddpEventsData.length === 0" class="ddp-empty">暂无事件</div>
                  <div v-else class="ddp-events-list">
                    <div v-for="(ev, idx) in ddpEventsData" :key="idx" class="ddp-event-row" :class="ev.type === 'Warning' ? 'warn' : ''">
                      <span class="ddp-ev-type" :class="ev.type === 'Warning' ? 'warn' : 'normal'">{{ ev.type }}</span>
                      <span class="ddp-ev-reason">{{ ev.reason }}</span>
                      <span class="ddp-ev-msg">{{ ev.message }}</span>
                      <span class="ddp-ev-age">{{ ddpEvAge(ev.lastTimestamp || ev.event_time || ev.eventTime) }}</span>
                    </div>
                  </div>
                </section>
              </template>
            </div>
          </div>
        </Transition>
      </div>
    </Teleport>

    <!-- 事件弹窗 -->
    <div v-if="showEventsModal" class="modal-overlay" @click.self="closeEvents">
      <div class="modal-content">
        <div class="modal-header">
          <h3>📡 Pod 事件</h3>
          <button class="close-btn" @click="closeEvents">×</button>
        </div>
        <div class="modal-body">
          <div v-if="loadingEvents" style="text-align: center; padding: 40px;">
            <div class="loading-spinner">加载中...</div>
          </div>
          <div v-else-if="eventsData.length > 0">
            <div v-for="(event, idx) in eventsData" :key="idx" class="event-item">
              <div class="event-type" :class="(event.type || 'normal').toLowerCase()">
                {{ event.type || 'Normal' }}
              </div>
              <div class="event-content">
                <div class="event-reason">{{ event.reason }}</div>
                <div class="event-message">{{ event.message }}</div>
                <div class="event-time">
                  {{ fmtTime(event.event_time) }}
                  <span v-if="event.count > 1" class="event-count">(发生 {{ event.count }} 次)</span>
                </div>
              </div>
            </div>
          </div>
          <div v-else class="empty-state">
            <div class="empty-icon">📬</div>
            <div class="empty-text">暂无事件记录</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 更新镜像弹窗 -->
    <div v-if="showPatchImageModal" class="modal-overlay" @click.self="closePatchImage">
      <div class="modal-content" style="max-width: 520px;">
        <div class="modal-header">
          <h3>🔧 更新容器镜像</h3>
          <button class="close-btn" @click="closePatchImage">×</button>
        </div>
        <div class="modal-body">
          <div v-if="selectedPod" style="background: #f7fafc; padding: 12px; border-radius: 8px; margin-bottom: 16px;">
            <div><strong>名称空间：</strong>{{ selectedPod.namespace }}</div>
            <div><strong>Pod 名称：</strong>{{ selectedPod.name }}</div>
          </div>

          <div class="form-group" v-if="containerList.length > 1">
            <label>选择容器</label>
            <select v-model="patchImageForm.container" class="form-select" :disabled="loadingContainers">
              <option value="" disabled>{{ loadingContainers ? '加载中...' : '请选择容器' }}</option>
              <option v-for="c in containerList" :key="c.name" :value="c.name">
                {{ c.name }} ({{ c.image }})
              </option>
            </select>
          </div>
          <div v-else-if="containerList.length === 1" class="form-group">
            <label>容器</label>
            <div class="form-static">{{ containerList[0].name }} ({{ containerList[0].image }})</div>
          </div>

          <div class="form-group">
            <label>新镜像地址</label>
            <input
              v-model="patchImageForm.newImage"
              type="text"
              class="form-input"
              placeholder="例如: nginx:1.28 或 registry.example.com/app:v2"
            />
          </div>

          <div v-if="patchImageError" class="error-box" style="margin-top: 12px;">
            {{ patchImageError }}
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closePatchImage" :disabled="patchingImage">
            取消
          </button>
          <button class="btn btn-primary" @click="submitPatchImage" :disabled="patchingImage || !patchImageForm.container || !patchImageForm.newImage">
            {{ patchingImage ? '更新中...' : '确认更新' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 创建 Pod 弹窗 -->
    <div v-if="showCreateModal" class="modal-overlay" @click.self="closeCreate">
      <div class="modal-content modal-create" style="max-width: 1000px; max-height: 85vh;">
        <div class="modal-header">
          <h3>➕ 创建 Pod</h3>
          <button class="close-btn" @click="closeCreate">×</button>
        </div>
        
        <!-- 模式切换标签 -->
        <div class="mode-tabs">
          <button 
            :class="['mode-tab', { active: createMode === 'form' }]"
            @click="createMode = 'form'"
          >
            <span class="tab-icon">📋</span>
            表单模式
          </button>
          <button 
            :class="['mode-tab', { active: createMode === 'yaml' }]"
            @click="createMode = 'yaml'"
          >
            <span class="tab-icon">📄</span>
            YAML
          </button>
        </div>

        <div class="modal-body">
          <!-- 表单模式 -->
          <div v-if="createMode === 'form'" class="form-mode">
            <div class="form-group">
              <label>名称空间 <span class="required">*</span></label>
              <div class="namespace-selector">
                <select v-model="createForm.namespace" class="form-select" :disabled="createForm.isCreatingNs">
                  <option value="" disabled>请选择名称空间</option>
                  <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
                </select>
                <span class="namespace-or">或</span>
                <div class="namespace-create">
                  <input 
                    v-model="createForm.newNamespace" 
                    type="text" 
                    class="form-input" 
                    placeholder="新名称空间名称"
                    :disabled="createForm.isCreatingNs"
                  />
                  <button 
                    type="button" 
                    class="btn btn-sm btn-secondary" 
                    @click="createNamespace"
                    :disabled="!createForm.newNamespace || createForm.isCreatingNs"
                  >
                    {{ createForm.isCreatingNs ? '创建中...' : '创建' }}
                  </button>
                </div>
              </div>
              <div class="form-hint">选择已有名称空间，或创建新的名称空间后选择</div>
            </div>
            <div class="form-group">
              <label>Pod 名称 <span class="required">*</span></label>
              <input v-model="createForm.name" type="text" class="form-input" placeholder="例如: my-nginx-pod" />
            </div>
            <div class="form-group">
              <label>容器名称 <span class="required">*</span></label>
              <input v-model="createForm.containerName" type="text" class="form-input" placeholder="例如: nginx 或留空使用 Pod 名称" />
              <div class="form-hint">留空则默认使用 Pod 名称作为容器名</div>
            </div>
            <div class="form-group">
              <label>容器镜像 <span class="required">*</span></label>
              <input v-model="createForm.image" type="text" class="form-input" placeholder="例如: nginx:1.27" />
            </div>
            <div class="form-group">
              <label>标签（可选，每行一个 key=value）</label>
              <textarea v-model="createForm.labelsText" class="form-textarea" rows="3" placeholder="app=demo&#10;env=test"></textarea>
            </div>
            <div v-if="createError" class="error-box" style="margin-top: 12px;">{{ createError }}</div>
          </div>

          <!-- YAML 模式 -->
          <div v-if="createMode === 'yaml'" class="yaml-mode">
            <div class="yaml-header">
              <h4>📜 YAML 配置</h4>
              <div class="yaml-header-buttons">
                <button class="btn btn-sm btn-secondary" @click="loadPodYamlTemplate">
                  <span class="btn-icon">📄</span>加载模板
                </button>
                <button class="btn btn-sm btn-clear" @click="clearYamlContent">
                  <span class="btn-icon">🗑️</span>清除
                </button>
              </div>
            </div>
            <div class="form-hint">✨ 支持多资源 YAML 创建（用 <code>---</code> 分隔），可同时创建 PVC、ConfigMap 等依赖资源</div>
            <textarea 
              v-model="yamlContent" 
              class="yaml-editor" 
              rows="20"
              placeholder="apiVersion: v1&#10;kind: Pod&#10;metadata:&#10;  name: example-pod&#10;  namespace: default&#10;spec:&#10;  containers:&#10;  - name: nginx&#10;    image: nginx:latest"
            ></textarea>
            <div v-if="yamlError" class="error-box">{{ yamlError }}</div>
            <div class="yaml-tips">
              <span class="tip-icon">💡</span>
              <strong>提示：</strong>
              <ul>
                <li>支持完整的 Kubernetes Pod 配置</li>
                <li>可以通过“加载模板”获取示例 YAML</li>
              </ul>
            </div>
          </div>
        </div>
        
        <div class="modal-footer">
          <button type="button" class="btn btn-secondary" @click="closeCreate" :disabled="creating">取消</button>
          <button 
            type="button"
            v-if="createMode === 'form'"
            class="btn btn-primary" 
            @click="submitCreate" 
            :disabled="creating || !createForm.namespace || !createForm.name || !createForm.image"
          >
            {{ creating ? '创建中...' : '✅ 创建 Pod' }}
          </button>
          <button 
            type="button"
            v-if="createMode === 'yaml'"
            class="btn btn-primary" 
            @click="createPodFromYaml" 
            :disabled="creating || !yamlContent"
          >
            {{ creating ? '创建中...' : '✅ 创建 Pod' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 日志弹窗 -->
    <div v-if="showLogsModal" class="modal-overlay" @click.self="closeLogs">
      <div class="modal-content logs-modal">
        <div class="modal-header">
          <h3>📄 容器日志</h3>
          <button class="close-btn" @click="closeLogs">×</button>
        </div>
        <div class="modal-body">
          <!-- Pod 信息 -->
          <div v-if="logsPod" class="pod-info-bar">
            <span><strong>名称空间：</strong>{{ logsPod.namespace }}</span>
            <span><strong>Pod：</strong>{{ logsPod.name }}</span>
            <span><strong>容器数量：</strong>{{ logsContainerList.length }}</span>
          </div>

          <!-- 日志控制栏 -->
          <div class="logs-control-bar">
            <!-- 容器选择 - 多容器时显示下拉框 -->
            <div class="control-item" v-if="logsContainerList.length > 1">
              <label>容器 <span style="color: #f56c6c; font-size: 12px;">*</span></label>
              <select v-model="logsForm.container" class="form-select">
                <option value="">请选择容器（共 {{ logsContainerList.length }} 个）</option>
                <option v-for="c in logsContainerList" :key="c" :value="c">{{ c }}</option>
              </select>
            </div>
            <!-- 容器选择 - 单容器时显示固定值 -->
            <div class="control-item" v-else-if="logsContainerList.length === 1">
              <label>容器</label>
              <span class="single-container">{{ logsContainerList[0] }} <span style="color: #67c23a; font-size: 12px;">(已自动选中)</span></span>
            </div>

            <!-- Tail 行数 -->
            <div class="control-item">
              <label>显示行数</label>
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

            <!-- 实时日志开关 -->
            <div class="control-item">
              <label class="follow-toggle">
                <input type="checkbox" v-model="logsForm.follow" />
                <span>实时日志</span>
                <span v-if="logsForm.follow && isStreaming" class="streaming-indicator">●</span>
              </label>
            </div>

            <!-- 保持时长（仅实时日志时显示） -->
            <div class="control-item" v-if="logsForm.follow">
              <label>保持时长</label>
              <select v-model="logsForm.duration" class="form-select">
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
              @click="fetchLogs" 
              :disabled="loadingLogs || (!logsForm.container && logsContainerList.length > 1)"
            >
              {{ loadingLogs ? '加载中...' : '获取日志' }}
            </button>

            <!-- 终止加载按钮 -->
            <button 
              v-if="loadingLogs || isStreaming" 
              class="btn btn-danger btn-sm" 
              @click="stopLogLoading"
            >
              终止
            </button>

            <!-- 清除日志按钮 -->
            <button 
              class="btn btn-secondary btn-sm" 
              @click="clearLogs"
              :disabled="!logsContent || loadingLogs"
            >
              清除
            </button>
          </div>

          <!-- 日志内容区 -->
          <div class="logs-content-wrapper">
            <div v-if="logsError" class="error-box">{{ logsError }}</div>
            <pre v-else class="logs-content" ref="logsContentRef" v-html="highlightedLogs"></pre>
          </div>
        </div>
      </div>
    </div>

    <!-- 批量删除预览弹窗（高危操作） -->
    <div v-if="showBatchDeleteModal" class="modal-overlay" @click.self="closeBatchDeleteModal">
      <div class="modal-content modal-batch-preview modal-danger">
        <div class="modal-header danger-header">
          <h3>{{ batchDeleteForce ? '💥 批量强制删除预览（高危）' : '🗑️ 批量删除预览（高危）' }}</h3>
          <button class="close-btn" @click="closeBatchDeleteModal">×</button>
        </div>
        <div class="modal-body">
          <!-- 警告提示 -->
          <div class="danger-warning">
            <div class="warning-icon-large">⚠️</div>
            <div class="warning-content">
              <div class="warning-title">
                {{ batchDeleteForce ? '将强制删除以下 Pod（不等待优雅终止）' : '将删除以下 Pod' }}
              </div>
              <ul class="warning-list">
                <li v-if="batchDeleteForce">强制删除会立即终止 Pod，可能导致数据丢失</li>
                <li v-else>Pod 将进入优雅终止状态</li>
                <li>此操作不可撤销！</li>
              </ul>
            </div>
          </div>
          
          <!-- 受影响 Pod 列表 -->
          <div class="preview-section">
            <div class="section-title">受影响 Pod ({{ selectedPods.length }})</div>
            <div class="affected-pods-detail">
              <div v-for="pod in selectedPods" :key="pod.uid" class="affected-pod-card">
                <div class="pod-info">
                  <span class="pod-name">📦 {{ pod.name }}</span>
                  <span class="pod-namespace">{{ pod.namespace }}</span>
                </div>
                <div class="pod-stats">
                  <span class="status-indicator" :class="(pod.status || 'unknown').toLowerCase()">
                    {{ pod.status }}
                  </span>
                  <span class="node-tag">{{ pod.node || '-' }}</span>
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

    <!-- 批量驱逐预览弹窗（高危操作） -->
    <div v-if="showBatchEvictModal" class="modal-overlay" @click.self="closeBatchEvictModal">
      <div class="modal-content modal-batch-preview modal-warning">
        <div class="modal-header warning-header">
          <h3>⚡ 批量驱逐预览（高危）</h3>
          <button class="close-btn" @click="closeBatchEvictModal">×</button>
        </div>
        <div class="modal-body">
          <!-- 警告提示 -->
          <div class="warning-box">
            <div class="warning-icon-large">⚠️</div>
            <div class="warning-content">
              <div class="warning-title">将驱逐以下 Pod</div>
              <ul class="warning-list">
                <li>驱逐会受 PDB（Pod Disruption Budget）约束</li>
                <li>相比直接删除更安全，但仍会中断服务</li>
                <li>Pod 将被从当前节点驱逐</li>
              </ul>
            </div>
          </div>
          
          <!-- 受影响 Pod 列表 -->
          <div class="preview-section">
            <div class="section-title">受影响 Pod ({{ selectedPods.length }})</div>
            <div class="affected-pods-detail">
              <div v-for="pod in selectedPods" :key="pod.uid" class="affected-pod-card">
                <div class="pod-info">
                  <span class="pod-name">📦 {{ pod.name }}</span>
                  <span class="pod-namespace">{{ pod.namespace }}</span>
                </div>
                <div class="pod-stats">
                  <span class="status-indicator" :class="(pod.status || 'unknown').toLowerCase()">
                    {{ pod.status }}
                  </span>
                  <span class="node-tag">{{ pod.node || '-' }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- 二次确认 -->
          <div class="confirm-section">
            <div class="section-title">请输入 "EVICT" 确认操作</div>
            <input 
              v-model="evictConfirmText" 
              placeholder="请输入 EVICT" 
              class="confirm-input"
              :class="{ valid: evictConfirmText === 'EVICT' }"
            />
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeBatchEvictModal">取消</button>
          <button 
            class="btn btn-warning" 
            @click="executeBatchEvict" 
            :disabled="evictConfirmText !== 'EVICT' || batchExecuting"
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
          <h3>📝 {{ yamlEditMode ? '编辑 YAML' : '查看 YAML' }} - {{ selectedYamlPod?.name }}</h3>
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
  </div>
</template>

<script setup>
import {computed, onMounted, onUnmounted, ref, watch, watchEffect} from 'vue';
import podsApi from '@/api/cluster/workloads/pods';
import namespaceApi from '@/api/cluster/namespaces';
import {useClusterStore} from '@/stores/cluster';
import Pagination from '@/components/Pagination.vue';
import KubeTerminal from '@/components/KubeTerminal.vue'
import { Message } from '@arco-design/web-vue'
import { useFilteredNamespaces } from '@/composables/useFilteredNamespaces'
import permissionStore from '@/stores/permission'
import { useConfirmDialog } from '@/composables/useConfirmDialog'

const { confirm: showConfirm } = useConfirmDialog()

// ===== 容器终端 =====
const showTerminal = ref(false)
const terminalPod = ref({ namespace: '', name: '', container: '' })

const openTerminal = (pod) => {
  terminalPod.value = {
    namespace: pod.namespace || '',
    name: pod.name,
    container: pod.containers?.[0] || '',
  }
  showTerminal.value = true
}

const closeTerminal = () => {
  showTerminal.value = false
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

// 获取认证 headers（与 http.js 拦截器逻辑一致）
const getAuthHeaders = () => {
  const headers = {};
  
  // 1) JWT Token（优先 localStorage，其次 sessionStorage）
  const token = localStorage.getItem('token') || sessionStorage.getItem('token');
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  
  // 2) X-Cluster-ID（从 store 获取，URL 兜底）
  const clusterStore = useClusterStore();
  let cid = clusterStore.current?.id;
  if (!cid) {
    // URL 兜底：/c/:clusterId/...
    const m = window.location.pathname.match(/\/c\/([^/]+)/);
    cid = m ? decodeURIComponent(m[1]) : '';
  }
  if (cid) {
    headers['X-Cluster-ID'] = String(cid);
  }
  
  return headers;
};

// 控制更多按钮显示
const showMoreOptions = ref(false);
const selectedPod = ref(null); // 选中的 pod
const menuStyle = ref({}); // 菜单动态样式

const toggleMoreOptions = (pod, event) => {
  if (selectedPod.value === pod && showMoreOptions.value) {
    showMoreOptions.value = false;
    selectedPod.value = null;
  } else {
    selectedPod.value = pod;
    showMoreOptions.value = true;

    // 计算菜单位置，使用 fixed 定位防止被父容器裁剪
    const button = event.target.closest('.more-btn');
    if (button) {
      const rect = button.getBoundingClientRect();
      const viewportHeight = window.innerHeight;
      const viewportWidth = window.innerWidth;
      const menuHeight = 280; // 预估菜单高度
      const menuWidth = 180; // 预估菜单宽度

      let style = {
        position: 'fixed',
      };

      // 垂直方向：如果下方空间不足，向上展开
      if (rect.bottom + menuHeight > viewportHeight) {
        style.bottom = (viewportHeight - rect.top + 4) + 'px';
      } else {
        style.top = (rect.bottom + 4) + 'px';
      }

      // 水平方向：如果右侧空间不足，向左展开
      if (rect.right + menuWidth > viewportWidth) {
        style.right = (viewportWidth - rect.right) + 'px';
      } else {
        style.left = rect.left + 'px';
      }

      menuStyle.value = style;
    }
  }
};

// 点击外部关闭菜单
const handleClickOutside = (event) => {
  if (showMoreOptions.value && !event.target.closest('.more-btn')) {
    showMoreOptions.value = false;
    selectedPod.value = null;
  }
};

// 滚动时关闭菜单（因为使用 fixed 定位）
const handleScroll = () => {
  if (showMoreOptions.value) {
    showMoreOptions.value = false;
    selectedPod.value = null;
  }
};

onMounted(() => {
  document.addEventListener('click', handleClickOutside);
  // 监听滚动事件
  document.addEventListener('scroll', handleScroll, true);
});

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside);
  document.removeEventListener('scroll', handleScroll, true);
  stopAutoRefresh(); // 清理自动刷新定时器
});

const loading = ref(false);
const errorMsg = ref('');
const searchQuery = ref('');
const statusFilter = ref('all');  // all, Running, Pending, Failed
let searchDebounceTimer = null;
const namespaceFilter = ref('');

// 使用命名空间权限过滤 Hook
const {
  filteredNamespaces: namespaces,
  selectedNamespace: namespaceFilterRef,
  loadNamespaces: fetchNamespaces,
  permissionHint: nsPermissionHint,
  hasFullAccess: nsHasFullAccess,
  loading: nsLoading
} = useFilteredNamespaces({ autoLoad: false })

// 同步 namespaceFilter 与权限过滤的 namespaceFilterRef
watch(namespaceFilterRef, (val) => {
  namespaceFilter.value = val
})
watch(namespaceFilter, (val) => {
  namespaceFilterRef.value = val
})
const currentPage = ref(1);
const itemsPerPage = ref(10);

// 自动刷新
const autoRefresh = ref(false);
let autoRefreshTimer = null;
const AUTO_REFRESH_INTERVAL = 90000; // 90秒

// 启动自动刷新
const startAutoRefresh = () => {
  if (autoRefreshTimer) return;
  autoRefreshTimer = setInterval(() => {
    if (!loading.value) {
      refresh();  // 使用 refresh() 而不是 fetchPods()，确保逻辑一致
    }
  }, AUTO_REFRESH_INTERVAL);
};

// 停止自动刷新
const stopAutoRefresh = () => {
  if (autoRefreshTimer) {
    clearInterval(autoRefreshTimer);
    autoRefreshTimer = null;
  }
};

// 监听 autoRefresh 开关
watch(autoRefresh, (val) => {
  if (val) {
    startAutoRefresh();
  } else {
    stopAutoRefresh();
  }
});

const pods = ref([]);

// 视图模式：table（表格） 或 card（卡片）
const viewMode = ref('table');

// ========== 批量操作相关 ==========
const batchMode = ref(false)
const selectedPods = ref([])
const showBatchDeleteModal = ref(false)
const batchDeleteForce = ref(false)
const deleteConfirmText = ref('')
const batchExecuting = ref(false)

// 批量驱逐相关
const showBatchEvictModal = ref(false)
const evictConfirmText = ref('')

// ========== YAML 查看/编辑相关 ==========
const showYamlModal = ref(false)
const selectedYamlPod = ref(null)
const yamlContent = ref('')
const yamlEditMode = ref(false)
const loadingYaml = ref(false)
const savingYaml = ref(false)
const yamlError = ref('')

const enterBatchMode = () => {
  batchMode.value = true
  // 默认全选当前页，用户可取消不需要的项
  selectedPods.value = [...paginatedPods.value]
}

const exitBatchMode = () => {
  batchMode.value = false
  selectedPods.value = []
}

const clearSelection = () => {
  selectedPods.value = []
}

const isPodSelected = (pod) => {
  // 只使用 name + namespace 比较，避免 uid 为空时的匹配问题
  return selectedPods.value.some(p => p.name === pod.name && p.namespace === pod.namespace)
}

const togglePodSelection = (pod) => {
  const index = selectedPods.value.findIndex(p => p.name === pod.name && p.namespace === pod.namespace)
  if (index >= 0) {
    selectedPods.value.splice(index, 1)
  } else {
    selectedPods.value.push(pod)
  }
}

const isAllSelected = computed(() => {
  return paginatedPods.value.length > 0 && 
         paginatedPods.value.every(pod => isPodSelected(pod))
})

// 部分选中状态（当前页有选中但不是全选）
const isPartialSelected = computed(() => {
  if (paginatedPods.value.length === 0) return false
  const selectedCount = paginatedPods.value.filter(pod => isPodSelected(pod)).length
  return selectedCount > 0 && selectedCount < paginatedPods.value.length
})

// 全选复选框 ref
const selectAllCheckbox = ref(null)

// 设置 indeterminate 状态（DOM 属性需要通过 JS 设置）
watchEffect(() => {
  if (selectAllCheckbox.value) {
    selectAllCheckbox.value.indeterminate = isPartialSelected.value
  }
})

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    paginatedPods.value.forEach(pod => {
      const index = selectedPods.value.findIndex(p => p.name === pod.name && p.namespace === pod.namespace)
      if (index >= 0) selectedPods.value.splice(index, 1)
    })
  } else {
    paginatedPods.value.forEach(pod => {
      if (!isPodSelected(pod)) {
        selectedPods.value.push(pod)
      }
    })
  }
}

// 点击表格行切换选中（批量模式下）
const handleRowClick = (pod, event) => {
  // 排除点击操作按钮区域
  if (event.target.closest('.action-icons') || 
      event.target.closest('.icon-btn') || 
      event.target.closest('.more-btn') ||
      event.target.closest('.more-menu')) {
    return
  }
  togglePodSelection(pod)
}

// 点击卡片切换选中（批量模式下）
const handleCardClick = (pod, event) => {
  // 排除点击按钮区域
  if (event.target.closest('.card-footer') || 
      event.target.closest('.card-action-btn') ||
      event.target.closest('.card-checkbox')) {
    return
  }
  togglePodSelection(pod)
}

// 打开批量驱逐预览弹窗
const openBatchEvictPreview = () => {
  evictConfirmText.value = ''
  showBatchEvictModal.value = true
}

const closeBatchEvictModal = () => {
  showBatchEvictModal.value = false
  evictConfirmText.value = ''
}

// 执行批量驱逐
const executeBatchEvict = async () => {
  if (evictConfirmText.value !== 'EVICT') return
  
  batchExecuting.value = true
  let successCount = 0
  let failCount = 0
  
  for (const pod of selectedPods.value) {
    try {
      await podsApi.evict({
        name: pod.name,
        namespace: pod.namespace
      })
      successCount++
    } catch (e) {
      console.error(`Failed to evict ${pod.name}:`, e)
      failCount++
    }
  }
  
  batchExecuting.value = false
  showBatchEvictModal.value = false
  evictConfirmText.value = ''
  
  if (failCount === 0) {
    alert(`成功驱逐 ${successCount} 个 Pod`)
  } else {
    alert(`成功 ${successCount} 个，失败 ${failCount} 个`)
  }
  
  exitBatchMode()
  refresh()
}

// 批量删除预览
const openBatchDeletePreview = (force) => {
  batchDeleteForce.value = force
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
  
  for (const pod of selectedPods.value) {
    try {
      await podsApi.delete({
        name: pod.name,
        namespace: pod.namespace,
        force: batchDeleteForce.value
      })
      successCount++
    } catch (e) {
      console.error(`Failed to delete ${pod.name}:`, e)
      failCount++
    }
  }
  
  batchExecuting.value = false
  showBatchDeleteModal.value = false
  deleteConfirmText.value = ''
  
  if (failCount === 0) {
    alert(`成功删除 ${successCount} 个 Pod`)
  } else {
    alert(`成功 ${successCount} 个，失败 ${failCount} 个`)
  }
  
  exitBatchMode()
  refresh()
}

// ========== 获取名称空间列表（已迁移到 useFilteredNamespaces Hook）==========
// fetchNamespaces 现在由 useFilteredNamespaces 提供


const showCreateModal = ref(false);
const showDeleteModal = ref(false);
const showDetailModal = ref(false);
const showEventsModal = ref(false);
const showLogsModal = ref(false);
const showPatchImageModal = ref(false);

// 创建 Pod
const creating = ref(false);
const createError = ref('');
const createMode = ref('form'); // 'form' | 'yaml'
const createForm = ref({
  namespace: 'default',
  newNamespace: '',
  isCreatingNs: false,
  name: '',
  containerName: '',
  image: '',
  labelsText: '',
});

// 监听 createMode 变化，切换到 YAML 模式时如果内容为空则自动加载模板
watch(createMode, (newMode) => {
  if (newMode === 'yaml' && !yamlContent.value.trim()) {
    loadPodYamlTemplate()
  }
})

const openCreate = () => {
  createForm.value = {
    namespace: namespaceFilter.value || 'default',
    newNamespace: '',
    isCreatingNs: false,
    name: '',
    containerName: '',
    image: '',
    labelsText: '',
  };
  createError.value = '';
  showCreateModal.value = true;
};

// 创建名称空间
const createNamespace = async () => {
  if (!createForm.value.newNamespace) return;
  createForm.value.isCreatingNs = true;
  try {
    await namespaceApi.create({ name: createForm.value.newNamespace.trim() });
    // 创建成功，刷新名称空间列表并自动选中
    await fetchNamespaces();
    createForm.value.namespace = createForm.value.newNamespace.trim();
    createForm.value.newNamespace = '';
  } catch (e) {
    createError.value = e?.msg || e?.message || '创建名称空间失败';
  } finally {
    createForm.value.isCreatingNs = false;
  }
};

const closeCreate = () => {
  showCreateModal.value = false;
  createError.value = '';
};

const submitCreate = async () => {
  creating.value = true;
  createError.value = '';
  try {
    // 解析标签（支持换行或逗号分隔）
    const labels = {};
    if (createForm.value.labelsText) {
      const lines = createForm.value.labelsText.split(/[\n,]/).map(s => s.trim()).filter(s => s);
      lines.forEach(line => {
        const [key, val] = line.split('=').map(s => s.trim());
        if (key && val) labels[key] = val;
      });
    }

    if (!createForm.value.namespace) {
      createError.value = '请选择名称空间';
      creating.value = false;
      return;
    }

    await podsApi.create({
      namespace: createForm.value.namespace,
      name: createForm.value.name,
      containerName: createForm.value.containerName || createForm.value.name,
      image: createForm.value.image,
      labels: Object.keys(labels).length > 0 ? labels : undefined,
    });

    closeCreate();
    await refresh();
    startTrackingStatus();
  } catch (e) {
    console.error(e);
    const details = e?.data?.details || e?.details;
    const msg = e?.msg || e?.message || '创建 Pod 失败';
    createError.value = details ? `${msg}: ${Array.isArray(details) ? details.join(', ') : details}` : msg;
  } finally {
    creating.value = false;
  }
};

// 加载 Pod YAML 模板
const loadPodYamlTemplate = () => {
  yamlContent.value = `apiVersion: v1
kind: Pod
metadata:
  name: example-pod
  namespace: default
  labels:
    app: example
spec:
  restartPolicy: Always
  containers:
  - name: nginx
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
        cpu: "200m"`
  yamlError.value = ''
  Message.success({ content: '已加载 YAML 模板，请修改后创建' })
}

// 清除 YAML 内容
const clearYamlContent = () => {
  yamlContent.value = ''
  yamlError.value = ''
  Message.success({ content: 'YAML 内容已清除' })
}

// 从 YAML 创建 Pod
const createPodFromYaml = async () => {
  if (!yamlContent.value.trim()) {
    Message.error({ content: '请输入 YAML 内容' })
    return
  }
  
  // 简单验证 YAML 格式
  try {
    if (!yamlContent.value.includes('kind: Pod')) {
      yamlError.value = 'YAML 中必须包含 "kind: Pod"'
      return
    }
    if (!yamlContent.value.includes('apiVersion: v1')) {
      yamlError.value = 'YAML 中必须包含 "apiVersion: v1"'
      return
    }
    
    yamlError.value = ''
  } catch (e) {
    yamlError.value = `YAML 格式错误: ${e.message}`
    return
  }
  
  creating.value = true
  try {
    const res = await podsApi.createFromYaml({ yaml: yamlContent.value })
    if (res.code === 0) {
      const msg = res.data?.message || 'Pod 创建成功'
      Message.success({ content: msg, duration: 5000 })
      showCreateModal.value = false
      yamlContent.value = ''
      yamlError.value = ''
      await refresh()
      startTrackingStatus()
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

const podToDelete = ref(null);
const deleting = ref(false);
const logsPod = ref(null);

// 时间格式
const fmtTime = (ts) => {
  if (!ts) return '-';
  try {
    const d = new Date(ts);
    const yyyy = d.getFullYear();
    const mm = String(d.getMonth() + 1).padStart(2, '0');
    const dd = String(d.getDate()).padStart(2, '0');
    const hh = String(d.getHours()).padStart(2, '0');
    const mi = String(d.getMinutes()).padStart(2, '0');
    const ss = String(d.getSeconds()).padStart(2, '0');
    return `${yyyy}-${mm}-${dd} ${hh}:${mi}:${ss}`;
  } catch {
    return ts;
  }
};

// Pod 结构适配 - 直接使用后端返回的状态
const mapPod = (item) => {
  return {
    uid: item.uid || '',
    name: item.name || '-',
    namespace: item.namespace || '-',
    node: item.node || '-',
    status: item.status || 'Unknown',  // 直接使用后端计算的状态
    statusReason: item.status_reason || '',  // 状态原因
    restartCount: item.restart_count || 0,
    createdAt: item.created_at || '-',
    image: item.image || '-',
    podIP: item.pod_ip || '-',
    hostIP: item.host_ip || '-',
    service: '', // 这里可以根据需要补充
    raw: item,
    metrics: null, // 资源消耗数据（异步获取）
  };
};

// 资源消耗数据（Map：podName -> metrics）
const metricsMap = ref({});

const fetchPods = async () => {
  loading.value = true;
  errorMsg.value = '';
  try {
    const params = {
      // 前端分页：获取所有数据，不传 page/limit 或传大值
      limit: 1000,
    };
    // 如果选择了特定名称空间，则传递；否则不传（获取所有）
    if (namespaceFilter.value) {
      params.namespace = namespaceFilter.value;
    }

    const res = await podsApi.list(params);

    let list = res?.data?.list || [];
    pods.value = (Array.isArray(list) ? list : []).map(mapPod);
    
    // 异步获取 metrics（不阻塞主流程）
    fetchPodsMetrics();
  } catch (e) {
    console.error(e);
    errorMsg.value = e?.msg || e?.message || '获取Pod列表失败';
    pods.value = [];
  } finally {
    loading.value = false;
  }
};

// 批量获取 Pod metrics（异步，不阻塞主流程）
const fetchPodsMetrics = async () => {
  // 切换命名空间时先清空旧数据
  metricsMap.value = {};
  pods.value.forEach(pod => { pod.metrics = null; });
  
  if (!namespaceFilter.value) return; // 需要指定 namespace
  if (pods.value.length === 0) return; // 没有 Pod 不需要请求
  
  try {
    const res = await podsApi.metricsList({ namespace: namespaceFilter.value });
    const data = res?.data || {};
    
    // 更新 metricsMap
    metricsMap.value = data;
    
    // 为每个 Pod 关联 metrics
    pods.value.forEach(pod => {
      if (data[pod.name]) {
        pod.metrics = data[pod.name];
      }
    });
  } catch (e) {
    // metrics 获取失败不影响主流程，静默处理
    console.warn('获取 Pod metrics 失败（可能 metrics-server 未安装）:', e);
  }
};

// 前端过滤：名称 + 状态
const filteredPods = computed(() => {
  const q = searchQuery.value.trim().toLowerCase();
  return pods.value.filter(pod => {
    // 名称模糊匹配
    const hitName = !q || pod.name.toLowerCase().includes(q);
    // 状态过滤
    const hitStatus = statusFilter.value === 'all' || 
      (pod.status || '').toLowerCase() === statusFilter.value.toLowerCase();
    return hitName && hitStatus;
  });
});

// 前端分页：切片当前页数据
const paginatedPods = computed(() => {
  const startIndex = (currentPage.value - 1) * itemsPerPage.value;
  const endIndex = startIndex + itemsPerPage.value;
  return filteredPods.value.slice(startIndex, endIndex);
});

// 总数 = 过滤后的数量
const total = computed(() => filteredPods.value.length);

const refresh = async () => {
  if (currentPage.value < 1) currentPage.value = 1;
  await fetchPods();
};

const onNamespaceChange = () => {
  currentPage.value = 1;
  refresh(); // 名称空间变化需要重新获取数据
};

// 状态过滤（前端过滤，不需要重新请求）
const setStatusFilter = (status) => {
  statusFilter.value = status;
  currentPage.value = 1;
};

// 输入时防抖（前端过滤，不需要重新请求）
const onSearchInput = () => {
  if (searchDebounceTimer) clearTimeout(searchDebounceTimer);
  searchDebounceTimer = setTimeout(() => {
    currentPage.value = 1;
  }, 300);
};

// 操作后自动追踪状态变化（10秒后自动停止）
let autoStopTimer = null;
const startTrackingStatus = () => {
  autoRefresh.value = true;
  if (autoStopTimer) clearTimeout(autoStopTimer);
  autoStopTimer = setTimeout(() => {
    autoRefresh.value = false;
  }, 15000); // 15秒后自动停止
};

// 删除
const isForceDelete = ref(false);
const deletePod = (pod, force = false) => {
  podToDelete.value = pod;
  isForceDelete.value = force;
  showDeleteModal.value = true;
  showMoreOptions.value = false;
};

const confirmDelete = async () => {
  if (!podToDelete.value) return;
  deleting.value = true;
  try {
    if (isForceDelete.value) {
      await podsApi.forceDelete({
        namespace: podToDelete.value.namespace,
        name: podToDelete.value.name,
      });
    } else {
      await podsApi.graceDelete({
        namespace: podToDelete.value.namespace,
        name: podToDelete.value.name,
      });
    }
    showDeleteModal.value = false;
    podToDelete.value = null;
    await refresh();
    startTrackingStatus(); // 开始追踪状态变化
  } catch (e) {
    console.error(e);
    alert(e?.msg || e?.message || '删除失败');
  } finally {
    deleting.value = false;
  }
};

// 重启
// 查看详情
const detailData = ref(null);
const loadingDetail = ref(false);
const ddpEventsData = ref([]);
const ddpEventsLoading = ref(false);
const ddpShowAnno = ref(false);
const ddpShowVols = ref(false);
const openDetail = async (pod) => {
  showMoreOptions.value = false;
  selectedPod.value = pod;
  showDetailModal.value = true;
  loadingDetail.value = true;
  detailData.value = null;
  ddpEventsData.value = [];
  ddpShowAnno.value = false;
  ddpShowVols.value = false;
  ddpEventsLoading.value = true;
  try {
    const [detailRes, eventsRes] = await Promise.all([
      podsApi.detail({ namespace: pod.namespace, name: pod.name }),
      podsApi.events({ namespace: pod.namespace, name: pod.name, limit: 50 })
    ]);
    detailData.value = detailRes?.data || pod.raw;
    ddpEventsData.value = eventsRes?.data?.events || eventsRes?.data?.items || [];
  } catch (e) {
    console.error(e);
    detailData.value = pod.raw;
    ddpEventsData.value = [];
  } finally {
    loadingDetail.value = false;
    ddpEventsLoading.value = false;
  }
};

const closeDetail = () => {
  showDetailModal.value = false;
  detailData.value = null;
};

// DDP helpers for Pod
const ddpStatusText = (d) => {
  if (!d?.status) return 'UNKNOWN';
  const phase = d.status.phase;
  // 严格按容器就绪状态判定
  const containers = d.status.containerStatuses || [];
  if (containers.length > 0 && containers.every(c => c.ready)) return 'RUNNING';
  if (phase === 'Succeeded') return 'SUCCEEDED';
  if (phase === 'Failed') return 'FAILED';
  if (phase === 'Pending') return 'PENDING';
  return phase?.toUpperCase() || 'UNKNOWN';
};
const ddpStatusClass = (d) => {
  const s = ddpStatusText(d);
  if (s === 'RUNNING' || s === 'SUCCEEDED') return 'status-ok';
  if (s === 'PENDING') return 'status-warn';
  if (s === 'FAILED') return 'status-err';
  return 'status-warn';
};
const ddpContainerStatus = (d, name) => {
  const statuses = d?.status?.containerStatuses || [];
  return statuses.find(s => s.name === name);
};
const ddpEvAge = (ts) => {
  if (!ts) return '';
  const diff = Math.floor((Date.now() - new Date(ts).getTime()) / 1000);
  if (diff < 60) return diff + 's ago';
  if (diff < 3600) return Math.floor(diff / 60) + 'm ago';
  if (diff < 86400) return Math.floor(diff / 3600) + 'h ago';
  return Math.floor(diff / 86400) + 'd ago';
};

// 查看事件
const eventsData = ref([]);
const loadingEvents = ref(false);
const openEvents = async (pod) => {
  showMoreOptions.value = false;
  selectedPod.value = pod;
  showEventsModal.value = true;
  loadingEvents.value = true;
  try {
    const res = await podsApi.events({
      namespace: pod.namespace,
      name: pod.name,
    });
    // 后端返回格式: { events: [...], next: '...', message: '...' }
    eventsData.value = res?.data?.events || [];
  } catch (e) {
    console.error(e);
    alert(e?.msg || e?.message || '获取事件失败');
    eventsData.value = [];
  } finally {
    loadingEvents.value = false;
  }
};

const closeEvents = () => {
  showEventsModal.value = false;
  eventsData.value = [];
};

// 驱逐Pod
const evictPod = async (pod) => {
  showMoreOptions.value = false;
  
  // 二次确认 - 驱逐是高危操作
  const ok = await showConfirm({
    title: '确认驱逐 Pod',
    content: '驱逐会受 PDB（Pod Disruption Budget）约束，相比直接删除更安全。',
    type: 'warning',
    details: [
      { label: 'Pod', value: `${pod.namespace}/${pod.name}`, mono: true },
      { label: '状态', value: pod.status || '-' },
      { label: '节点', value: pod.node || '-' },
    ],
    confirmText: '确认驱逐',
    cancelText: '取消',
  })
  if (!ok) return;
  
  try {
    // 注意：后端字段名必须为 podName
    await podsApi.evict({
      namespace: pod.namespace,
      podName: pod.name,
    });
    await refresh();
    startTrackingStatus(); // 开始追踪状态变化
  } catch (e) {
    console.error(e);
    alert(e?.msg || e?.message || '驱逐失败');
  }
};

// ========== YAML 查看/编辑功能 ==========
const openYamlPreview = async (pod) => {
  showMoreOptions.value = false
  selectedYamlPod.value = pod
  yamlContent.value = ''
  yamlError.value = ''
  yamlEditMode.value = false
  loadingYaml.value = true
  showYamlModal.value = true

  try {
    const res = await podsApi.yaml({ namespace: pod.namespace, name: pod.name })
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
  selectedYamlPod.value = null
  yamlContent.value = ''
  yamlError.value = ''
  yamlEditMode.value = false
}

const applyYamlChanges = async () => {
  if (!yamlContent.value?.trim()) {
    alert('请输入 YAML 内容')
    return
  }
  savingYaml.value = true
  try {
    const res = await podsApi.applyYaml({
      namespace: selectedYamlPod.value.namespace,
      name: selectedYamlPod.value.name,
      yaml: yamlContent.value
    })
    if (res.code === 0) {
      alert('YAML 应用成功')
      closeYamlModal()
      refresh()
      autoRefresh.value = true
      setTimeout(() => { autoRefresh.value = false }, 15000)
    } else {
      alert(res.msg || '应用 YAML 失败')
    }
  } catch (e) {
    alert(e?.msg || e?.message || '应用 YAML 失败')
  } finally {
    savingYaml.value = false
  }
}

// 下载 YAML
const downloadYaml = () => {
  if (!yamlContent.value || !selectedYamlPod.value) {
    alert('没有可下载的 YAML 内容')
    return
  }
  
  try {
    const blob = new Blob([yamlContent.value], { type: 'text/yaml;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `${selectedYamlPod.value.name}-pod.yaml`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
    alert('YAML 文件已下载')
  } catch (e) {
    alert('下载失败')
  }
}

// 日志相关
const logsContainerList = ref([]);
const loadingLogs = ref(false);
const logsContent = ref('');
const logsError = ref('');
const isStreaming = ref(false);
const logsContentRef = ref(null);
let logAbortController = null;
let logDurationTimer = null;  // 保持时长定时器

const logsForm = ref({
  container: '',
  tail: null,  // null 表示获取全部日志
  follow: false,
  duration: 0,  // 保持时长（秒），0 表示不限制
});

const openLogs = (pod) => {
  showMoreOptions.value = false;
  logsPod.value = pod;
  logsContent.value = '';
  logsError.value = '';
  isStreaming.value = false;
  
  // 从后端返回的 containers 字段获取容器名列表
  logsContainerList.value = pod.raw?.containers || [];
  
  // ✅ 单容器自动选中，多容器清空（让用户选择）
  if (logsContainerList.value.length === 1) {
    logsForm.value.container = logsContainerList.value[0];
  } else if (logsContainerList.value.length > 1) {
    logsForm.value.container = '';  // 多容器时清空，强制用户选择
  } else {
    logsForm.value.container = '';  // 无容器时清空
  }
  
  showLogsModal.value = true;
};

const closeLogs = () => {
  stopLogStream();
  showLogsModal.value = false;
  logsPod.value = null;
  logsContent.value = '';
  logsError.value = '';
  logsContainerList.value = [];
};

// 停止流式日志
const stopLogStream = () => {
  // 清除保持时长定时器
  if (logDurationTimer) {
    clearTimeout(logDurationTimer);
    logDurationTimer = null;
  }
  if (logAbortController) {
    logAbortController.abort();
    logAbortController = null;
  }
  isStreaming.value = false;
};

// 终止所有日志加载（包括一次性和流式）
const stopLogLoading = () => {
  stopLogStream();
  loadingLogs.value = false;
};

// 获取日志
const fetchLogs = async () => {
  if (!logsPod.value) return;
  
  // 确定容器名称
  let container = logsForm.value.container || '';
  
  // 如果只有一个容器，自动使用
  if (logsContainerList.value.length === 1) {
    container = logsContainerList.value[0];
  }
  
  // 多容器时必须选择容器
  if (logsContainerList.value.length > 1 && !container) {
    logsError.value = '请选择容器';
    return;
  }
  
  // 确保容器名不为空
  if (!container) {
    logsError.value = '无法确定容器名称';
    return;
  }
  
  // 停止之前的流
  stopLogStream();
  
  loadingLogs.value = true;
  logsError.value = '';
  logsContent.value = '';
  
  // 创建新的 AbortController
  logAbortController = new AbortController();
  
  try {
    if (logsForm.value.follow) {
      // 实时日志（流式）
      await fetchStreamLogs(container);
    } else {
      // 一次性获取日志（也支持取消）
      const params = new URLSearchParams({
        namespace: logsPod.value.namespace,
        name: logsPod.value.name,
        container: container,
      });
      if (logsForm.value.tail != null) {
        params.set('tail', logsForm.value.tail);
      }
      
      const response = await fetch(`/api/v1/k8s/pod/container_logs?${params}`, {
        signal: logAbortController.signal,
        headers: getAuthHeaders(),
      });
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }
      
      const res = await response.json();
      logsContent.value = res?.data?.log || '暂无日志';
    }
  } catch (e) {
    if (e.name === 'AbortError') {
      console.log('日志请求已取消');
    } else {
      console.error('获取日志失败:', e);
      logsError.value = e?.msg || e?.message || '获取日志失败';
    }
  } finally {
    loadingLogs.value = false;
    if (!logsForm.value.follow) {
      logAbortController = null;
    }
  }
};

// 流式获取日志
const fetchStreamLogs = async (container) => {
  isStreaming.value = true;
  // AbortController 已在 fetchLogs 中创建
  
  // 设置保持时长定时器（如果有设置）
  if (logsForm.value.duration > 0) {
    logDurationTimer = setTimeout(() => {
      console.log(`实时日志已达到保持时长 ${logsForm.value.duration} 秒，自动停止`);
      stopLogLoading();
    }, logsForm.value.duration * 1000);
  }
  
  try {
    const params = new URLSearchParams({
      namespace: logsPod.value.namespace,
      name: logsPod.value.name,
      container: container,
      follow: 'true',
    });
    // 仅当 tail 有值时才传递（null 表示获取全部）
    if (logsForm.value.tail != null) {
      params.set('tail', logsForm.value.tail);
    }
    
    const response = await fetch(`/api/v1/k8s/pod/container_log?${params}`, {
      signal: logAbortController.signal,
      headers: getAuthHeaders(),
    });
    
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }
    
    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      
      const text = decoder.decode(value, { stream: true });
      logsContent.value += text;
      
      // 自动滚动到底部
      if (logsContentRef.value) {
        logsContentRef.value.scrollTop = logsContentRef.value.scrollHeight;
      }
    }
  } catch (e) {
    if (e.name === 'AbortError') {
      console.log('日志流已停止');
    } else {
      console.error('流式日志错误:', e);
      logsError.value = e?.message || '流式日志获取失败';
    }
  } finally {
    isStreaming.value = false;
    logAbortController = null;
  }
};

// 清除日志
const clearLogs = () => {
  logsContent.value = '';
  logsError.value = '';
};

// 日志高亮处理
const highlightedLogs = computed(() => {
  if (!logsContent.value) {
    return '<span class="log-placeholder">暂无日志，请点击"获取日志"按钮</span>';
  }
  
  // 转义 HTML 特殊字符
  const escapeHtml = (str) => {
    return str
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;');
  };
  
  const escaped = escapeHtml(logsContent.value);
  
  // 按行处理
  return escaped.split('\n').map(line => {
    // 时间戳高亮 (ISO 格式或常见日志时间格式)
    let highlighted = line.replace(
      /(\d{4}-\d{2}-\d{2}[T\s]\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:?\d{2})?)/g,
      '<span class="log-timestamp">$1</span>'
    );
    
    // ERROR / FATAL / PANIC 高亮（红色）
    if (/\b(ERROR|FATAL|PANIC|error|fatal|panic)\b/i.test(line)) {
      highlighted = `<span class="log-error">${highlighted}</span>`;
    }
    // WARN / WARNING 高亮（黄色）
    else if (/\b(WARN|WARNING|warn|warning)\b/i.test(line)) {
      highlighted = `<span class="log-warn">${highlighted}</span>`;
    }
    // INFO 高亮（蓝色）
    else if (/\b(INFO|info)\b/i.test(line)) {
      highlighted = `<span class="log-info">${highlighted}</span>`;
    }
    // DEBUG 高亮（灰色）
    else if (/\b(DEBUG|debug)\b/i.test(line)) {
      highlighted = `<span class="log-debug">${highlighted}</span>`;
    }
    
    return highlighted;
  }).join('\n');
});

// patch 镜像
const containerList = ref([]);
const loadingContainers = ref(false);
const patchingImage = ref(false);
const patchImageError = ref('');
const patchImageForm = ref({
  container: '',
  newImage: '',
});

const openPatchImage = (pod) => {
  showMoreOptions.value = false;
  selectedPod.value = pod;
  showPatchImageModal.value = true;
  patchImageForm.value = { container: '', newImage: '' };
  patchImageError.value = '';

  // 从后端返回的 container_statuses 提取容器信息（包含名称和镜像）
  const containerStatuses = pod.raw?.container_statuses || [];
  containerList.value = containerStatuses.map(cs => ({
    name: cs.name,
    image: cs.image || '-'
  }));

  // 如果 container_statuses 为空，尝试从 containers 字段获取（仅容器名）
  if (containerList.value.length === 0 && pod.raw?.containers) {
    containerList.value = pod.raw.containers.map(name => ({
      name: name,
      image: '-'
    }));
  }

  // 单容器自动选中
  if (containerList.value.length === 1) {
    patchImageForm.value.container = containerList.value[0].name;
  }
};

const closePatchImage = () => {
  showPatchImageModal.value = false;
  containerList.value = [];
  patchImageForm.value = { container: '', newImage: '' };
  patchImageError.value = '';
};

const submitPatchImage = async () => {
  if (!selectedPod.value || !patchImageForm.value.container || !patchImageForm.value.newImage) {
    patchImageError.value = '请填写完整信息';
    return;
  }
  
  // 获取当前镜像
  const currentContainer = containerList.value.find(c => c.name === patchImageForm.value.container);
  const oldImage = currentContainer?.image || '-';

  // 二次确认 - 镜像更新是高危操作
  const okImg = await showConfirm({
    title: '确认更新 Pod 镜像',
    content: 'Pod 镜像更新后，Pod 将被删除重建，可能导致服务中断。',
    type: 'danger',
    details: [
      { label: 'Pod', value: `${selectedPod.value.namespace}/${selectedPod.value.name}`, mono: true },
      { label: '容器', value: patchImageForm.value.container },
      { label: '旧镜像', value: oldImage, mono: true },
      { label: '新镜像', value: patchImageForm.value.newImage, highlight: true, mono: true },
    ],
    tip: '建议通过 Deployment/StatefulSet 更新镜像以实现滚动更新。',
    confirmText: '确认更新',
    cancelText: '取消',
  })
  if (!okImg) {
    return;
  }

  patchingImage.value = true;
  patchImageError.value = '';

  try {
    await podsApi.patchImage({
      namespace: selectedPod.value.namespace,
      name: selectedPod.value.name,
      container: patchImageForm.value.container,
      new_image: patchImageForm.value.newImage,
    });
    closePatchImage();
    await refresh();
    startTrackingStatus(); // 开始追踪状态变化
    alert('镜像更新成功！Pod 将被重建...');
  } catch (e) {
    console.error(e);
    patchImageError.value = e?.msg || e?.message || '更新镜像失败';
  } finally {
    patchingImage.value = false;
  }
};

onMounted(async () => {
  // 并行获取名称空间列表和 Pod 列表
  await Promise.all([fetchNamespaces(), fetchPods()]);
  document.addEventListener('click', handleClickOutside);
});

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside);
});
</script>


<style scoped>
/* ✅ 主容器：使用 Flex 布局填满可用空间 */
.resource-view {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  flex: 1; /* ✅ 填充父容器剩余空间 */
  min-height: 0; /* ✅ 关键：让 flex 子元素可以缩小 */
  box-sizing: border-box;
}

.view-header {
  flex-shrink: 0;
  margin-bottom: clamp(16px, 2vw, 24px);
}

.view-header h1 {
  font-size: clamp(24px, 3vw, 32px);
  font-weight: 700;
  color: #1a202c;
  margin-bottom: 6px;
  letter-spacing: -0.02em;
}

.view-header p {
  font-size: clamp(14px, 1.5vw, 16px);
  color: #718096;
  margin: 0;
}

.action-bar {
  display: flex;
  justify-content: flex-start;
  align-items: center;
  margin-bottom: clamp(14px, 2vw, 20px);
  flex-wrap: wrap;
  gap: clamp(8px, 1.5vw, 14px);
  flex-shrink: 0;
}

.search-box input {
  padding: clamp(8px, 1vw, 12px) clamp(12px, 1.5vw, 16px);
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  font-size: 14px;
  width: clamp(180px, 20vw, 320px);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
  transition: all 0.2s ease;
}

.search-box input:focus {
  outline: none;
  border-color: #326ce5;
  box-shadow: 0 0 0 3px rgba(50, 108, 229, 0.12);
}

.filter-buttons {
  display: flex;
  gap: 6px;
}

.btn-filter {
  padding: 8px 14px;
  font-size: 13px;
  border: 1px solid #e2e8f0;
  background: #fff;
  color: #64748b;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.btn-filter:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
}

.btn-filter.active {
  background: #3b82f6;
  border-color: #3b82f6;
  color: #fff;
}

.filter-dropdown select,
.filter-dropdown input {
  padding: clamp(8px, 1vw, 12px) clamp(10px, 1vw, 14px);
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  font-size: 14px;
  background-color: white;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
  transition: all 0.2s ease;
  cursor: pointer;
}

.filter-dropdown select:focus,
.filter-dropdown input:focus {
  outline: none;
  border-color: #326ce5;
  box-shadow: 0 0 0 3px rgba(50, 108, 229, 0.12);
}

.page-input {
  width: clamp(70px, 8vw, 120px);
}

.action-buttons {
  display: flex;
  gap: clamp(8px, 1vw, 12px);
  margin-left: auto;
  align-items: center;
}

.auto-refresh-toggle {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  cursor: pointer;
  font-size: 13px;
  color: #64748b;
  transition: all 0.2s;
}

.auto-refresh-toggle:hover {
  background: #f1f5f9;
  border-color: #cbd5e1;
}

.auto-refresh-toggle input {
  cursor: pointer;
}

.auto-refresh-toggle input:checked + span {
  color: #3b82f6;
  font-weight: 500;
}

.refresh-indicator {
  color: #22c55e;
  animation: pulse 1s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

.btn {
  padding: clamp(8px, 1vw, 12px) clamp(14px, 1.5vw, 20px);
  border: none;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
}

.btn:hover {
  transform: translateY(-1px);
}

.btn:active {
  transform: translateY(0);
}

.btn-primary {
  background: linear-gradient(135deg, #326ce5 0%, #2553b9 100%);
  color: white;
  box-shadow: 0 2px 8px rgba(50, 108, 229, 0.25);
}

.btn-primary:hover {
  box-shadow: 0 4px 12px rgba(50, 108, 229, 0.35);
}

.btn-secondary {
  background-color: #f1f5f9;
  color: #475569;
}

.btn-secondary:hover {
  background-color: #e2e8f0;
}

.btn-danger {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: #fff;
  box-shadow: 0 2px 8px rgba(239, 68, 68, 0.25);
}

.btn-danger:hover {
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.35);
}

/* ✅ 表格容器：自动填充剩余高度 */
.table-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: white;
  border-radius: clamp(10px, 1.5vw, 16px);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.06);
  min-height: 300px;
  min-width: 0; /* ✅ 防止内容撑开 */
}

/* ✅ 表格内容区域可滚动 */
.table-wrapper {
  flex: 1;
  overflow: auto;
  min-height: 0; /* ✅ 关键：让 flex 子元素可以缩小 */
  border-radius: clamp(10px, 1.5vw, 16px) clamp(10px, 1.5vw, 16px) 0 0;
}

.resource-table {
  width: 100%;
  border-collapse: collapse;
  min-width: 1200px;
  table-layout: auto; /* 改为自动布局，让内容决定列宽 */
}

/* ✅ 表格列样式 - 自动布局，内容决定宽度 */
.resource-table th,
.resource-table td {
  white-space: nowrap; /* 不换行，完整显示 */
}

.resource-table th {
  background: linear-gradient(180deg, #f8fafc 0%, #f1f5f9 100%);
  text-align: left;
  padding: clamp(12px, 1.5vw, 16px) clamp(12px, 1.5vw, 18px);
  font-size: clamp(12px, 1.2vw, 14px);
  font-weight: 600;
  color: #475569;
  border-bottom: 1px solid #e2e8f0;
  position: sticky;
  top: 0;
  z-index: 10;
}

.resource-table td {
  padding: clamp(12px, 1.5vw, 16px) clamp(12px, 1.5vw, 18px);
  font-size: clamp(12px, 1.2vw, 14px);
  color: #334155;
  border-bottom: 1px solid #f1f5f9;
  vertical-align: middle;
  transition: background-color 0.15s ease;
}

.resource-table tbody tr:hover {
  background-color: #f8fafc;
}

.resource-table tbody tr:hover td {
  color: #1e293b;
}

/* 状态单元格 - Jobs 风格 */
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

.status-indicator {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: clamp(4px, 0.5vw, 6px) clamp(10px, 1vw, 14px);
  border-radius: 20px;
  font-size: clamp(11px, 1vw, 12px);
  font-weight: 600;
  letter-spacing: 0.02em;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.status-indicator.running {
  background: linear-gradient(135deg, rgba(52, 211, 153, 0.15) 0%, rgba(16, 185, 129, 0.1) 100%);
  color: #059669;
}

.status-indicator.succeeded,
.status-indicator.success {
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.15) 0%, rgba(37, 99, 235, 0.1) 100%);
  color: #2563eb;
}

.status-indicator.pending {
  background: linear-gradient(135deg, rgba(245, 158, 11, 0.15) 0%, rgba(217, 119, 6, 0.1) 100%);
  color: #d97706;
}

.status-indicator.failed {
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.15) 0%, rgba(220, 38, 38, 0.1) 100%);
  color: #dc2626;
}

.status-indicator.crashloopbackoff {
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.2) 0%, rgba(220, 38, 38, 0.12) 100%);
  color: #dc2626;
  font-weight: 700;
}

.status-indicator.imagepullbackoff {
  background: linear-gradient(135deg, rgba(251, 146, 60, 0.15) 0%, rgba(249, 115, 22, 0.1) 100%);
  color: #ea580c;
}

.status-indicator.notready {
  background: linear-gradient(135deg, rgba(156, 163, 175, 0.15) 0%, rgba(107, 114, 128, 0.1) 100%);
  color: #6b7280;
}

.status-indicator.unknown {
  background: linear-gradient(135deg, rgba(148, 163, 184, 0.2) 0%, rgba(100, 116, 139, 0.12) 100%);
  color: #64748b;
}

.pod-name {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

/* 可点击的 Pod 名称链接 */
.pod-name-link {
  color: #326ce5;
  cursor: pointer;
  text-decoration: none;
  font-weight: 500;
  transition: all 0.15s ease;
}

.pod-name-link:hover {
  color: #1a47b8;
  text-decoration: underline;
}

.pod-name .icon {
  font-size: 18px;
}

.namespace-badge {
  display: inline-block;
  padding: clamp(3px, 0.4vw, 5px) clamp(8px, 1vw, 12px);
  background: linear-gradient(135deg, rgba(50, 108, 229, 0.12) 0%, rgba(37, 83, 185, 0.08) 100%);
  color: #2553b9;
  border-radius: 8px;
  font-size: clamp(11px, 1vw, 12px);
  font-weight: 600;
}

.resource-usage {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}


.usage-low {
  background-color: #34d399;
}

.usage-medium {
  background-color: #f59e0b;
}

.usage-high {
  background-color: #ef4444;
}

.usage-text {
  font-size: 12px;
  font-weight: 500;
  color: #718096;
  min-width: 18px;
}

.service-tag {
  display: inline-block;
  padding: clamp(3px, 0.4vw, 5px) clamp(8px, 1vw, 12px);
  background: linear-gradient(135deg, rgba(139, 92, 246, 0.12) 0%, rgba(109, 40, 217, 0.08) 100%);
  color: #7c3aed;
  border-radius: 8px;
  font-size: clamp(11px, 1vw, 12px);
  font-weight: 600;
}

.no-service {
  color: #94a3b8;
  font-size: clamp(11px, 1vw, 12px);
}

.action-icons {
  display: flex;
  gap: 6px;
  flex-wrap: nowrap;
  justify-content: flex-start;
  align-items: center;
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
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  font-size: 12px;
  cursor: pointer;
  padding: 6px 10px;
  border-radius: 6px;
  color: #475569;
  transition: all 0.2s ease;
  white-space: nowrap;
  font-weight: 500;
  min-width: 60px;
  text-align: center;
}

.icon-btn:hover {
  background-color: #f1f5f9;
  border-color: #e2e8f0;
  color: #326ce5;
  transform: translateY(-1px);
}

.icon-btn.danger:hover {
  background-color: rgba(239, 68, 68, 0.08);
  border-color: rgba(239, 68, 68, 0.2);
  color: #dc2626;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: clamp(40px, 6vw, 80px) 20px;
  color: #94a3b8;
}

.empty-icon {
  font-size: clamp(36px, 5vw, 56px);
  margin-bottom: clamp(10px, 1.5vw, 16px);
  opacity: 0.8;
}

.empty-text {
  font-size: clamp(13px, 1.2vw, 15px);
}

/* ✅ 底部分页栏优化 */
.pager {
  flex-shrink: 0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: clamp(10px, 1.2vw, 14px) clamp(14px, 1.5vw, 20px);
  font-size: clamp(12px, 1.1vw, 13px);
  color: #64748b;
  background: linear-gradient(180deg, #fafbfc 0%, #f8fafc 100%);
  border-top: 1px solid #e2e8f0;
  border-radius: 0 0 clamp(10px, 1.5vw, 16px) clamp(10px, 1.5vw, 16px);
}

/* ✅ Pagination 组件样式（确保底部固定显示） */
.table-container :deep(.pagination-wrapper) {
  flex-shrink: 0;
  background: linear-gradient(180deg, #fafbfc 0%, #f8fafc 100%);
  border-top: 1px solid #e2e8f0;
  border-radius: 0 0 clamp(10px, 1.5vw, 16px) clamp(10px, 1.5vw, 16px);
}

.pager > div {
  display: flex;
  align-items: center;
  gap: 8px;
}

.error-box {
  flex-shrink: 0;
  margin-bottom: clamp(10px, 1.5vw, 16px);
  padding: clamp(10px, 1.2vw, 14px) clamp(12px, 1.5vw, 16px);
  border-radius: 10px;
  color: #b91c1c;
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.08) 0%, rgba(220, 38, 38, 0.04) 100%);
  border: 1px solid rgba(239, 68, 68, 0.2);
  font-size: clamp(13px, 1.2vw, 14px);
}

/* ✅ 表单样式 */
.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  font-size: 14px;
  font-weight: 600;
  color: #334155;
  margin-bottom: 8px;
}

.form-input,
.form-select {
  width: 100%;
  padding: 12px 14px;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  font-size: 14px;
  background-color: white;
  transition: all 0.2s ease;
  box-sizing: border-box;
}

.form-input:focus,
.form-select:focus {
  outline: none;
  border-color: #326ce5;
  box-shadow: 0 0 0 3px rgba(50, 108, 229, 0.12);
}

.form-input::placeholder {
  color: #94a3b8;
}

.form-textarea {
  width: 100%;
  padding: 12px 14px;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  font-size: 14px;
  font-family: ui-monospace, monospace;
  resize: vertical;
  min-height: 80px;
  box-sizing: border-box;
  transition: all 0.2s ease;
}

.form-textarea:focus {
  outline: none;
  border-color: #326ce5;
  box-shadow: 0 0 0 3px rgba(50, 108, 229, 0.12);
}

.required {
  color: #ef4444;
}

.form-hint {
  font-size: 12px;
  color: #94a3b8;
  margin-top: 4px;
}

.input-with-select {
  display: flex;
  align-items: center;
  gap: 8px;
}

.input-divider {
  color: #94a3b8;
  font-size: 12px;
  flex-shrink: 0;
}

.namespace-selector {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.namespace-selector .form-select {
  flex: 1;
  min-width: 150px;
}

.namespace-or {
  color: #94a3b8;
  font-size: 12px;
  flex-shrink: 0;
}

.namespace-create {
  display: flex;
  gap: 6px;
  flex: 1;
  min-width: 200px;
}

.namespace-create .form-input {
  flex: 1;
}

.btn-sm {
  padding: 8px 12px;
  font-size: 12px;
  white-space: nowrap;
}

.form-select {
  cursor: pointer;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='16' height='16' viewBox='0 0 24 24' fill='none' stroke='%2364748b' stroke-width='2'%3E%3Cpolyline points='6 9 12 15 18 9'%3E%3C/polyline%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 12px center;
  padding-right: 40px;
}

.form-select:disabled {
  background-color: #f1f5f9;
  color: #94a3b8;
  cursor: not-allowed;
}

.form-static {
  padding: 12px 14px;
  background-color: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  font-size: 14px;
  color: #334155;
  font-family: ui-monospace, monospace;
}

.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
  font-size: clamp(11px, 1vw, 13px);
}

.muted {
  color: #94a3b8;
  font-size: clamp(11px, 1vw, 12px);
}

.ellipsis {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 100%;
  display: block;
}

/* ✅ 表格单元格内容自适应 */
.resource-table td {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* modal */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.5);
  backdrop-filter: blur(4px);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 999;
  animation: overlayFadeIn 0.2s ease-out;
}

@keyframes overlayFadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.modal-content {
  background: #fff;
  border-radius: clamp(12px, 1.5vw, 16px);
  width: min(92vw, 900px);
  max-height: 85vh;
  overflow: hidden;
  box-shadow: 0 25px 50px rgba(0, 0, 0, 0.2);
  animation: modalSlideIn 0.25s ease-out;
}

@keyframes modalSlideIn {
  from {
    opacity: 0;
    transform: translateY(-20px) scale(0.98);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: clamp(12px, 1.5vw, 18px) clamp(16px, 2vw, 24px);
  border-bottom: 1px solid #e2e8f0;
  background: linear-gradient(180deg, #fafbfc 0%, #fff 100%);
}

.modal-header h3 {
  font-size: clamp(16px, 1.5vw, 18px);
  font-weight: 600;
  color: #1e293b;
  margin: 0;
}

.close-btn {
  border: none;
  background: #f1f5f9;
  font-size: clamp(18px, 2vw, 22px);
  cursor: pointer;
  width: 32px;
  height: 32px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #64748b;
  transition: all 0.2s ease;
}

.close-btn:hover {
  background: #e2e8f0;
  color: #334155;
}

.modal-body {
  padding: clamp(14px, 2vw, 20px) clamp(16px, 2vw, 24px);
  max-height: 60vh;
  overflow-y: auto;
}

/* 模式切换标签 - Rancher/Kuboard 风格 */
.mode-tabs {
  display: flex;
  gap: 0;
  padding: 0 24px;
  border-bottom: 2px solid #e2e8f0;
  background: linear-gradient(180deg, #fafbfc 0%, #f8fafc 100%);
}

.mode-tab {
  flex: 1;
  padding: 14px 20px;
  border: none;
  background: transparent;
  font-size: 14px;
  font-weight: 500;
  color: #64748b;
  cursor: pointer;
  position: relative;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.mode-tab:hover {
  color: #326ce5;
  background: rgba(50, 108, 229, 0.05);
}

.mode-tab.active {
  color: #326ce5;
  font-weight: 600;
}

.mode-tab.active::after {
  content: '';
  position: absolute;
  bottom: -2px;
  left: 0;
  right: 0;
  height: 2px;
  background: #326ce5;
}

.tab-icon {
  font-size: 16px;
}

/* YAML 模式 */
.yaml-mode {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.yaml-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.yaml-header h4 {
  font-size: 15px;
  font-weight: 600;
  color: #1e293b;
  margin: 0;
}

.yaml-header-buttons {
  display: flex;
  gap: 8px;
}

.btn-clear {
  background: linear-gradient(135deg, #94a3b8 0%, #64748b 100%) !important;
  color: white !important;
  transition: all 0.3s ease !important;
}

.btn-clear:hover {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%) !important;
}

.yaml-editor {
  width: 100%;
  padding: 14px;
  border: 2px solid #334155;
  border-radius: 8px;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
  resize: vertical;
  background: #1e1e1e;
  color: #d4d4d4;
  transition: all 0.2s ease;
}

.yaml-editor:focus {
  outline: none;
  border-color: #326ce5;
  box-shadow: 0 0 0 3px rgba(50, 108, 229, 0.3);
  background: #1e1e1e;
}

.yaml-tips {
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
  border-left: 3px solid #3b82f6;
  padding: 12px 16px;
  border-radius: 8px;
  font-size: 13px;
  color: #1e40af;
}

.yaml-tips .tip-icon {
  font-size: 16px;
  margin-right: 6px;
}

.yaml-tips ul {
  margin: 8px 0 0 0;
  padding-left: 20px;
}

.yaml-tips li {
  margin: 4px 0;
}

.modal-footer {
  padding: clamp(12px, 1.5vw, 18px) clamp(16px, 2vw, 24px);
  border-top: 1px solid #e2e8f0;
  display: flex;
  justify-content: flex-end;
  gap: clamp(8px, 1vw, 12px);
  background: linear-gradient(180deg, #fff 0%, #fafbfc 100%);
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: clamp(10px, 1.5vw, 16px);
}

.detail-item label {
  font-size: clamp(11px, 1vw, 12px);
  color: #64748b;
  display: block;
  margin-bottom: 4px;
  font-weight: 500;
}

.detail-json {
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  padding: clamp(12px, 1.5vw, 18px);
  border-radius: 10px;
  font-size: clamp(11px, 1vw, 12px);
  line-height: 1.7;
  overflow-x: auto;
  color: #334155;
  font-family: 'Consolas', 'Monaco', ui-monospace, monospace;
  border: 1px solid #e2e8f0;
}

.event-item {
  display: flex;
  gap: clamp(10px, 1.5vw, 16px);
  padding: clamp(10px, 1.5vw, 14px);
  border-bottom: 1px solid #f1f5f9;
  transition: background-color 0.15s ease;
}

.event-item:hover {
  background-color: #fafbfc;
}

.event-item:last-child {
  border-bottom: none;
}

.event-type {
  padding: clamp(3px, 0.4vw, 5px) clamp(8px, 1vw, 10px);
  border-radius: 6px;
  font-size: clamp(10px, 0.9vw, 11px);
  font-weight: 600;
  text-transform: uppercase;
  flex-shrink: 0;
  height: fit-content;
  letter-spacing: 0.03em;
}

.event-type.normal {
  background: linear-gradient(135deg, rgba(52, 211, 153, 0.15) 0%, rgba(16, 185, 129, 0.1) 100%);
  color: #059669;
}

.event-type.warning {
  background: linear-gradient(135deg, rgba(245, 158, 11, 0.15) 0%, rgba(217, 119, 6, 0.1) 100%);
  color: #d97706;
}

.event-content {
  flex: 1;
  min-width: 0;
}

.event-reason {
  font-weight: 600;
  color: #1e293b;
  margin-bottom: 4px;
  font-size: clamp(13px, 1.2vw, 14px);
}

.event-message {
  font-size: clamp(12px, 1.1vw, 13px);
  color: #475569;
  margin-bottom: 6px;
  line-height: 1.5;
}

.event-time {
  font-size: clamp(10px, 0.9vw, 11px);
  color: #94a3b8;
  font-family: ui-monospace, monospace;
}

.event-count {
  margin-left: 8px;
  color: #64748b;
  font-weight: 500;
}

.loading-spinner {
  color: #326ce5;
  font-weight: 500;
  font-size: clamp(13px, 1.2vw, 14px);
}

.more-btn {
  position: static; /* 改为 static，菜单使用 fixed 定位 */
  display: inline-block;
}

.more-btn .icon-btn {
  font-size: clamp(14px, 1.5vw, 18px);
  padding: clamp(5px, 0.6vw, 8px) clamp(6px, 0.8vw, 10px);
}

/* ✅ 更多菜单 - 使用 fixed 定位防止被裁剪 */
.more-menu {
  position: fixed; /* 改为 fixed 定位 */
  background-color: white;
  border: 1px solid #e2e8f0;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.15);
  border-radius: 10px;
  z-index: 9999; /* 确保在最上层 */
  min-width: clamp(160px, 15vw, 200px);
  padding: 6px 0;
  animation: menuFadeIn 0.15s ease-out;
}

@keyframes menuFadeIn {
  from {
    opacity: 0;
    transform: translateY(-8px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.menu-item {
  display: flex;
  align-items: center;
  gap: clamp(8px, 1vw, 12px);
  width: 100%;
  padding: clamp(8px, 1vw, 12px) clamp(12px, 1.5vw, 18px);
  border: none;
  background: none;
  cursor: pointer;
  font-size: clamp(13px, 1.2vw, 14px);
  color: #334155;
  text-align: left;
  transition: all 0.15s ease;
  font-weight: 500;
}

.menu-item:hover {
  background-color: #f1f5f9;
  color: #326ce5;
}

.menu-item.danger {
  color: #dc2626;
}

.menu-item.danger:hover {
  background-color: rgba(239, 68, 68, 0.08);
  color: #b91c1c;
}

.menu-icon {
  font-size: clamp(14px, 1.3vw, 16px);
  width: clamp(18px, 1.8vw, 22px);
  text-align: center;
  flex-shrink: 0;
}

.menu-divider {
  height: 1px;
  background: linear-gradient(90deg, transparent 0%, #e2e8f0 10%, #e2e8f0 90%, transparent 100%);
  margin: 6px 0;
}

/* ✅ 响应式设计 - 更全面 */

/* 超大屏幕 */
@media (min-width: 1920px) {
  .resource-view {
    padding: 32px;
  }

  .view-header h1 {
    font-size: 36px;
  }

  .resource-table th,
  .resource-table td {
    padding: 18px 20px;
    font-size: 15px;
  }

  .btn {
    padding: 14px 24px;
    font-size: 15px;
  }
}

/* 大屏幕 */
@media (min-width: 1440px) and (max-width: 1919px) {
  .resource-view {
    padding: 28px;
  }

  .resource-table th,
  .resource-table td {
    padding: 16px 18px;
  }
}

@media (max-width: 1400px) {
  .action-bar {
    justify-content: flex-start;
  }
}

@media (max-width: 1024px) {
  .action-buttons {
    margin-left: 0;
    width: 100%;
    order: 10;
  }

  .action-buttons .btn {
    flex: 1;
  }
}

@media (max-width: 768px) {
  .view-header h1 {
    font-size: clamp(20px, 4vw, 24px);
  }

  .action-bar {
    gap: 8px;
  }

  .search-box {
    width: 100%;
    order: 1;
  }

  .search-box input {
    width: 100%;
    max-width: 100%;
  }

  .filter-dropdown {
    flex: 1;
    min-width: 0;
  }

  .filter-dropdown select,
  .filter-dropdown input {
    width: 100%;
  }

  .table-container {
    border-radius: clamp(8px, 1.5vw, 12px);
    min-height: 250px;
  }

  .resource-table {
    font-size: 12px;
  }

  .resource-table th,
  .resource-table td {
    padding: 10px 8px;
  }

  .icon-btn {
    padding: 5px 6px;
    font-size: 11px;
  }

  .modal-content {
    width: 96vw;
    max-height: 90vh;
    border-radius: clamp(10px, 2vw, 14px);
  }

  .pager {
    flex-direction: column;
    gap: 8px;
    text-align: center;
  }
}

@media (max-width: 480px) {
  .resource-view {
    padding: 10px;
  }

  .filter-dropdown:nth-child(n+3) {
    display: none;
  }

  .resource-table th,
  .resource-table td {
    padding: 8px 6px;
  }

  /* 小屏幕隐藏部分列 */
  .resource-table th:nth-child(4),
  .resource-table td:nth-child(4),
  .resource-table th:nth-child(7),
  .resource-table td:nth-child(7) {
    display: none;
  }
}

/* ✅ 打印样式 */
@media print {
  .action-bar,
  .action-icons,
  .modal-overlay,
  .pager {
    display: none !important;
  }

  .resource-view {
    height: auto;
    overflow: visible;
  }

  .table-container {
    box-shadow: none;
    border: 1px solid #e2e8f0;
    overflow: visible;
  }

  .table-wrapper {
    overflow: visible;
  }

  .resource-table {
    font-size: 10px;
  }
}

/* ========== 日志弹窗样式 ========== */
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
  background: #f8fafc;
  border-radius: 8px;
  margin-bottom: 12px;
  font-size: 13px;
  color: #475569;
}

.logs-control-bar {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px 0;
  border-bottom: 1px solid #e2e8f0;
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
}

.logs-control-bar .form-select {
  padding: 6px 10px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-size: 13px;
  background: #fff;
  cursor: pointer;
}

.logs-control-bar .form-select:focus {
  outline: none;
  border-color: #3b82f6;
}

.single-container {
  padding: 6px 10px;
  background: #e2e8f0;
  border-radius: 6px;
  font-size: 13px;
  color: #334155;
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
}

.follow-toggle:hover {
  background: #e2e8f0;
}

.follow-toggle input:checked + span {
  color: #3b82f6;
  font-weight: 500;
}

.streaming-indicator {
  color: #22c55e;
  animation: pulse 1s infinite;
}

.btn-sm {
  padding: 6px 14px;
  font-size: 13px;
}

.btn-danger {
  background: #ef4444;
  color: #fff;
}

.btn-danger:hover {
  background: #dc2626;
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
  background: #1e293b;
  color: #e2e8f0;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
  line-height: 1.6;
  border-radius: 8px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
}

.logs-content::-webkit-scrollbar {
  width: 8px;
}

.logs-content::-webkit-scrollbar-track {
  background: #334155;
  border-radius: 4px;
}

.logs-content::-webkit-scrollbar-thumb {
  background: #64748b;
  border-radius: 4px;
}

.logs-content::-webkit-scrollbar-thumb:hover {
  background: #94a3b8;
}

/* 日志高亮样式 */
.logs-content :deep(.log-placeholder) {
  color: #64748b;
  font-style: italic;
}

.logs-content :deep(.log-timestamp) {
  color: #94a3b8;
}

.logs-content :deep(.log-error) {
  color: #f87171;
  background: rgba(248, 113, 113, 0.1);
  display: block;
  margin: 0 -14px;
  padding: 0 14px;
}

.logs-content :deep(.log-warn) {
  color: #fbbf24;
  background: rgba(251, 191, 36, 0.08);
  display: block;
  margin: 0 -14px;
  padding: 0 14px;
}

.logs-content :deep(.log-info) {
  color: #60a5fa;
}

.logs-content :deep(.log-debug) {
  color: #9ca3af;
}

/* 行号显示（可选） */
.logs-content {
  counter-reset: line;
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
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 20px;
  margin-bottom: 20px;
}

.pod-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  overflow: hidden;
  transition: all 0.3s ease;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.pod-card:hover {
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
  font-size: 16px;
  font-weight: 600;
  color: #1e293b;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 卡片视图可点击的 Pod 名称链接 */
.card-title-link {
  color: inherit;
  text-decoration: none;
  cursor: pointer;
  transition: color 0.15s ease;
}

.card-title-link:hover {
  color: #326ce5;
  text-decoration: underline;
}

/* 卡片主体 */
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
  font-weight: 600;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 6px;
}

.card-section-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.card-meta-item {
  background: #f8fafc;
  padding: 10px;
  border-radius: 8px;
}

.meta-label {
  font-size: 10px;
  font-weight: 600;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 4px;
}

.meta-value {
  font-size: 13px;
  color: #1e293b;
  word-break: break-all;
}

/* 卡片底部按钮 */
.card-footer {
  display: flex;
  gap: 6px;
  padding: 12px 16px;
  background: #f8fafc;
  border-top: 1px solid #e2e8f0;
  flex-wrap: wrap;
}

.card-action-btn {
  flex: 1;
  min-width: 65px;
  padding: 7px 10px;
  background: white;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.card-action-btn:hover {
  background: #f3f4f6;
  border-color: #9ca3af;
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
  margin-bottom: 8px;
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

.container-metrics {
  margin-top: 8px;
  padding: 8px;
  background: #f8fafc;
  border-radius: 6px;
  border: 1px solid #e2e8f0;
}

.container-metrics-item {
  padding: 6px 8px;
  border-bottom: 1px solid #e2e8f0;
}

.container-metrics-item:last-child {
  border-bottom: none;
}

.container-metrics-name {
  font-size: 11px;
  font-weight: 600;
  color: #64748b;
  margin-bottom: 4px;
}

.container-metrics-values {
  display: flex;
  gap: 12px;
  font-size: 11px;
}

.container-cpu,
.container-memory {
  color: #475569;
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

.row-clickable {
  cursor: pointer;
  user-select: none;
}

.row-clickable:hover {
  background: rgba(102, 126, 234, 0.05);
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

.affected-pods-detail {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 200px;
  overflow-y: auto;
}

.affected-pod-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #f7fafc;
  padding: 12px 16px;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
}

.pod-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.pod-info .pod-name {
  font-weight: 600;
  color: #2d3748;
}

.pod-info .pod-namespace {
  font-size: 12px;
  color: #718096;
}

.pod-stats {
  display: flex;
  gap: 12px;
  align-items: center;
}

.node-tag {
  font-size: 12px;
  color: #718096;
  background: #edf2f7;
  padding: 2px 8px;
  border-radius: 4px;
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

.btn-danger {
  background: #e53e3e;
  color: white;
  border: none;
}

.btn-danger:hover:not(:disabled) {
  background: #c53030;
}

.btn-danger:disabled {
  background: #feb2b2;
  cursor: not-allowed;
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

.pod-card {
  position: relative;
}

.card-selected {
  border: 2px solid #667eea !important;
  background: rgba(102, 126, 234, 0.05) !important;
}

.card-clickable {
  cursor: pointer;
  user-select: none;
}

.card-clickable:hover {
  box-shadow: 0 4px 16px rgba(102, 126, 234, 0.2);
}

/* ========== YAML 模态框样式 ========== */
.yaml-modal {
  width: 90%;
  max-width: 900px;
  height: 80vh;
  display: flex;
  flex-direction: column;
}

.yaml-modal-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}

.yaml-header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.yaml-editor-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}

.yaml-editor {
  flex: 1;
  width: 100%;
  min-height: 300px;
  padding: 16px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', monospace;
  font-size: 13px;
  line-height: 1.6;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: #1e1e1e;
  color: #d4d4d4;
  resize: none;
  box-sizing: border-box;
}

.yaml-editor:focus {
  outline: none;
  border-color: #326ce5;
}

.yaml-content {
  flex: 1;
  margin: 0;
  padding: 16px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', monospace;
  font-size: 13px;
  line-height: 1.6;
  background: #1e1e1e;
  color: #d4d4d4;
  border-radius: 8px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
}

.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: #718096;
  font-size: 14px;
}

.btn-sm {
  padding: 6px 12px;
  font-size: 12px;
}

/* ===== DDP Describe 抽屉 ===== */
.ddp-overlay{position:fixed;inset:0;z-index:9000;background:rgba(15,23,42,.45);backdrop-filter:blur(2px)}
.ddp-panel{position:absolute;right:0;top:0;height:100%;width:880px;max-width:92vw;display:flex;flex-direction:column;background:#f8fafc;box-shadow:-4px 0 24px rgba(0,0,0,.18)}
.ddp-hd{padding:18px 28px;background:linear-gradient(135deg,#0f172a 0%,#1e3a5f 100%);display:flex;align-items:center;justify-content:space-between}
.ddp-hd-top{display:flex;align-items:center;gap:10px;min-width:0}
.ddp-hd-icon{font-size:22px}
.ddp-hd-title{color:#fff;font-size:18px;font-weight:700;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
.ddp-hd-ns{background:rgba(255,255,255,.12);color:#93c5fd;padding:2px 10px;border-radius:20px;font-size:12px;white-space:nowrap}
.ddp-hd-status{display:flex;align-items:center;gap:5px;padding:3px 12px;border-radius:20px;font-size:12px;font-weight:600;color:#fff}
.ddp-hd-status.status-ok{background:rgba(34,197,94,.2);color:#4ade80}
.ddp-hd-status.status-warn{background:rgba(234,179,8,.2);color:#facc15}
.ddp-hd-status.status-err{background:rgba(239,68,68,.2);color:#f87171}
.ddp-led{width:7px;height:7px;border-radius:50%;margin-right:4px;display:inline-block}
.status-ok .ddp-led{background:#4ade80;box-shadow:0 0 6px #4ade80}
.status-warn .ddp-led{background:#facc15;box-shadow:0 0 6px #facc15}
.status-err .ddp-led{background:#f87171;box-shadow:0 0 6px #f87171}
.ddp-close{background:none;border:none;color:rgba(255,255,255,.7);font-size:28px;cursor:pointer;line-height:1;padding:0 4px}
.ddp-close:hover{color:#fff}
.ddp-body{flex:1;overflow-y:auto;padding:24px 28px}
.ddp-loading{display:flex;align-items:center;gap:10px;justify-content:center;padding:40px 0;color:#64748b}
.ddp-empty{color:#94a3b8;text-align:center;padding:18px 0;font-size:13px}
.ddp-sec{margin-bottom:22px}
.ddp-sec-title{font-size:14px;font-weight:700;color:#1e293b;margin-bottom:10px;padding-bottom:6px;border-bottom:1px solid #e2e8f0}
.ddp-fold-title{cursor:pointer;user-select:none}
.ddp-fold-icon{float:right;font-size:12px;color:#94a3b8}
.ddp-kv-grid{display:grid;grid-template-columns:1fr 1fr;gap:6px 18px}
.ddp-kv{display:flex;gap:8px;font-size:13px;padding:4px 0;border-bottom:1px dashed #f1f5f9}
.ddp-k{color:#64748b;min-width:100px;flex-shrink:0}
.ddp-v{color:#1e293b;word-break:break-all}
.ddp-v.mono{font-family:monospace;font-size:12px}
.ddp-chip-list{display:flex;flex-wrap:wrap;gap:6px}
.ddp-chip{background:#e0e7ff;color:#3730a3;padding:2px 10px;border-radius:12px;font-size:12px;max-width:340px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.ddp-chip.anno{background:#fef3c7;color:#92400e}
.ddp-container-card{background:#fff;border:1px solid #e2e8f0;border-radius:8px;padding:12px 16px;margin-bottom:10px}
.ddp-container-name{font-weight:600;font-size:13px;color:#0f172a;margin-bottom:8px;display:flex;align-items:center;gap:8px}
.ddp-cstatus{font-size:11px;padding:2px 8px;border-radius:10px;font-weight:500}
.ddp-cstatus.ready{background:rgba(34,197,94,.1);color:#16a34a}
.ddp-cstatus.not-ready{background:rgba(239,68,68,.1);color:#dc2626}
.ddp-events-list{max-height:280px;overflow-y:auto;border:1px solid #e2e8f0;border-radius:8px;background:#fff}
.ddp-event-row{display:grid;grid-template-columns:60px 100px 1fr 70px;align-items:center;gap:8px;padding:7px 12px;font-size:12px;border-bottom:1px solid #f1f5f9}
.ddp-event-row.warn{background:#fefce8}
.ddp-ev-type{padding:1px 6px;border-radius:4px;font-weight:600;text-align:center}
.ddp-ev-type.normal{background:rgba(34,197,94,.1);color:#16a34a}
.ddp-ev-type.warn{background:rgba(234,179,8,.12);color:#a16207}
.ddp-ev-reason{font-weight:600;color:#334155;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
.ddp-ev-msg{color:#64748b;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
.ddp-ev-age{color:#94a3b8;text-align:right;white-space:nowrap}
.ddp-slide-enter-active{transition:transform .32s cubic-bezier(.4,0,.2,1)}
.ddp-slide-leave-active{transition:transform .22s cubic-bezier(.4,0,1,1)}
.ddp-slide-enter-from,.ddp-slide-leave-to{transform:translateX(100%)}
.action-btn.primary{background:rgba(59,130,246,.08);color:#2563eb;border:1px solid rgba(59,130,246,.2);padding:4px 10px;border-radius:6px;cursor:pointer;font-size:12px;white-space:nowrap}
.action-btn.primary:hover{background:rgba(59,130,246,.15)}

</style>
