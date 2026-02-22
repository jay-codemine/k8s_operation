<template>
  <div class="resource-view">
    <div class="view-header">
      <h1>PersistentVolume 管理</h1>
      <p>Kubernetes 集群中的持久卷列表</p>
    </div>

    <div class="action-bar">
      <div class="search-box">
        <input type="text" v-model="searchQuery" placeholder="搜索 PV 名称..." @input="onSearchInput" />
      </div>

      <div class="filter-buttons">
        <button class="btn btn-filter" :class="{ active: statusFilter === 'all' }" @click="setStatusFilter('all')">
          全部
        </button>
        <button class="btn btn-filter" :class="{ active: statusFilter === 'Available' }" @click="setStatusFilter('Available')">
          Available
        </button>
        <button class="btn btn-filter" :class="{ active: statusFilter === 'Bound' }" @click="setStatusFilter('Bound')">
          Bound
        </button>
        <button class="btn btn-filter" :class="{ active: statusFilter === 'Released' }" @click="setStatusFilter('Released')">
          Released
        </button>
        <button class="btn btn-filter" :class="{ active: statusFilter === 'Failed' }" @click="setStatusFilter('Failed')">
          Failed
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
        <button class="btn btn-primary" @click="showCreateModal = true">创建 PV</button>
        <button class="btn btn-secondary" @click="refreshList" :disabled="loading">
          {{ loading ? '加载中...' : '🔄 刷新' }}
        </button>
      </div>
    </div>

    <div v-if="errorMsg" class="error-box">{{ errorMsg }}</div>

    <!-- 批量操作浮动栏 -->
    <div v-if="batchMode && selectedPVs.length > 0" class="batch-action-bar">
      <div class="batch-info">
        <span class="batch-count">已选择 {{ selectedPVs.length }} 个 PV</span>
        <button class="batch-clear" @click="clearSelection">清空选择</button>
      </div>
      <div class="batch-actions">
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
              />
            </th>
            <th style="width: 100px;">状态</th>
            <th style="min-width: 180px;">名称</th>
            <th style="width: 100px;">容量</th>
            <th style="width: 150px;">访问模式</th>
            <th style="width: 120px;">回收策略</th>
            <th style="min-width: 150px;">StorageClass</th>
            <th style="min-width: 150px;">绑定的 PVC</th>
            <th style="width: 170px; white-space: nowrap;">创建时间</th>
            <th style="width: 200px;">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="pv in paginatedPVs" :key="pv.name" :class="{ 'row-selected': isPVSelected(pv) }">
            <td v-if="batchMode">
              <input 
                type="checkbox" 
                :checked="isPVSelected(pv)" 
                @change="togglePVSelection(pv)"
              />
            </td>
            <td>
              <span class="status-indicator" :class="getStatusClass(pv.status)">
                {{ pv.status }}
              </span>
            </td>
            <td>
              <div class="pv-name">
                <span class="icon">💾</span>
                <span>{{ pv.name }}</span>
              </div>
            </td>
            <td>{{ pv.capacity }}</td>
            <td>
              <span class="access-mode-badge">{{ pv.accessModes }}</span>
            </td>
            <td>
              <span class="reclaim-badge" :class="getReclaimClass(pv.reclaimPolicy)">
                {{ pv.reclaimPolicy }}
              </span>
            </td>
            <td>
              <span class="storage-class-badge">{{ pv.storageClassName || '-' }}</span>
            </td>
            <td>
              <span class="claim-ref">{{ pv.claimRef || '-' }}</span>
            </td>
            <td style="white-space: nowrap;">{{ pv.createdAt }}</td>
            <td>
              <div class="action-icons">
                <button class="action-btn icon-only" @click="viewDetail(pv)" title="查看详情">
                  📋
                </button>
                <button class="action-btn icon-only" @click="viewYaml(pv)" title="查看YAML">
                  📝
                </button>
                <button class="action-btn icon-only danger" @click="confirmDelete(pv)" title="删除">
                  🗑️
                </button>
                
                <!-- 更多菜单 -->
                <div class="more-actions-wrapper">
                  <button 
                    class="action-btn icon-only more-btn" 
                    @click="toggleMoreMenu(pv)"
                    title="更多操作"
                  >
                    ⋮
                  </button>
                  <div 
                    v-if="activeMoreMenu === pv.name" 
                    class="more-menu"
                    @click.stop
                  >
                    <button class="menu-item" @click="downloadYaml(pv)">
                      <span class="menu-icon">💾</span>
                      <span>下载 YAML</span>
                    </button>
                    <button 
                      v-if="pv.reclaimPolicy !== 'Retain'"
                      class="menu-item" 
                      @click="changeReclaimPolicy(pv)"
                    >
                      <span class="menu-icon">🔒</span>
                      <span>改为 Retain</span>
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

      <div v-if="!loading && filteredPVs.length === 0" class="empty-state">
        <div class="empty-icon">📦</div>
        <div class="empty-text">暂无 PersistentVolume 数据</div>
        <button class="btn btn-primary" @click="showCreateModal = true">创建第一个 PV</button>
      </div>
    </div>

    <!-- 卡片视图 -->
    <div v-if="viewMode === 'card'" class="card-container">
      <div v-for="pv in paginatedPVs" :key="pv.name" class="resource-card" :class="{ 'card-selected': isPVSelected(pv) }">
        <div class="card-header">
          <div class="card-title">
            <input 
              v-if="batchMode"
              type="checkbox" 
              :checked="isPVSelected(pv)" 
              @change="togglePVSelection(pv)"
              class="card-checkbox"
            />
            <span class="icon">💾</span>
            <span class="name">{{ pv.name }}</span>
          </div>
          <span class="status-indicator" :class="getStatusClass(pv.status)">
            {{ pv.status }}
          </span>
        </div>
        <div class="card-body">
          <div class="card-info">
            <div class="info-item">
              <span class="label">容量:</span>
              <span class="value">{{ pv.capacity }}</span>
            </div>
            <div class="info-item">
              <span class="label">访问模式:</span>
              <span class="value">{{ pv.accessModes }}</span>
            </div>
            <div class="info-item">
              <span class="label">回收策略:</span>
              <span class="reclaim-badge" :class="getReclaimClass(pv.reclaimPolicy)">
                {{ pv.reclaimPolicy }}
              </span>
            </div>
            <div class="info-item">
              <span class="label">StorageClass:</span>
              <span class="value">{{ pv.storageClassName || '-' }}</span>
            </div>
            <div class="info-item">
              <span class="label">绑定的 PVC:</span>
              <span class="value">{{ pv.claimRef || '-' }}</span>
            </div>
            <div class="info-item">
              <span class="label">创建时间:</span>
              <span class="value">{{ pv.createdAt }}</span>
            </div>
          </div>
        </div>
        <div class="card-footer">
          <button class="card-action-btn" @click="viewDetail(pv)">📋 详情</button>
          <button class="card-action-btn" @click="viewYaml(pv)">📝 YAML</button>
          <button class="card-action-btn" @click="downloadYaml(pv)">💾 下载</button>
          <button class="card-action-btn danger" @click="confirmDelete(pv)">🗑️ 删除</button>
        </div>
      </div>

      <div v-if="loading" class="loading-indicator">
        <div class="loading-spinner"></div>
        <p>加载中...</p>
      </div>

      <div v-if="!loading && filteredPVs.length === 0" class="empty-state">
        <div class="empty-icon">📦</div>
        <div class="empty-text">暂无 PersistentVolume 数据</div>
        <button class="btn btn-primary" @click="showCreateModal = true">创建第一个 PV</button>
      </div>
    </div>

    <!-- 分页（现代化三段式布局） -->
    <div v-if="filteredPVs.length > 0" class="pagination-wrapper">
      <div class="pagination-left">
        <span class="pagination-summary">共 <strong>{{ filteredPVs.length }}</strong> 条</span>
      </div>
      <div class="pagination-center">
        <button class="pagination-btn" @click="goToPage(1)" :disabled="currentPage === 1" title="首页">«</button>
        <button class="pagination-btn" @click="goToPage(currentPage - 1)" :disabled="currentPage === 1" title="上一页">‹</button>
        
        <template v-for="page in visiblePages" :key="page">
          <button v-if="typeof page === 'number'" class="pagination-btn page-number" :class="{ active: currentPage === page }" @click="goToPage(page)">
            {{ page }}
          </button>
          <span v-else class="pagination-ellipsis">...</span>
        </template>
        
        <button class="pagination-btn" @click="goToPage(currentPage + 1)" :disabled="currentPage === totalPages" title="下一页">›</button>
        <button class="pagination-btn" @click="goToPage(totalPages)" :disabled="currentPage === totalPages" title="尾页">»</button>
      </div>
      <div class="pagination-right">
        <select v-model.number="itemsPerPage" @change="onPageSizeChange" class="page-size-select">
          <option :value="10">10 条/页</option>
          <option :value="20">20 条/页</option>
          <option :value="50">50 条/页</option>
          <option :value="100">100 条/页</option>
        </select>
        <span class="pagination-goto">前往</span>
        <input v-model.number="jumpPage" type="number" min="1" :max="totalPages" class="page-jump-input" @keyup.enter="jumpToPage" />
      </div>
    </div>

    <!-- 创建 PV 模态框 -->
    <div v-if="showCreateModal" class="modal-overlay" @click="showCreateModal = false">
      <div 
        ref="createModalRef"
        class="modal-content resizable-modal" 
        @click.stop
        :style="createModalStyle"
      >
        <!-- 拖拽调整大小的手柄 -->
        <div class="resize-handle resize-handle-top" @mousedown="startResize($event, 'top')"></div>
        <div class="resize-handle resize-handle-bottom" @mousedown="startResize($event, 'bottom')"></div>
        <div class="resize-handle resize-handle-left" @mousedown="startResize($event, 'left')"></div>
        <div class="resize-handle resize-handle-right" @mousedown="startResize($event, 'right')"></div>
        <div class="resize-handle resize-handle-top-left" @mousedown="startResize($event, 'top-left')"></div>
        <div class="resize-handle resize-handle-top-right" @mousedown="startResize($event, 'top-right')"></div>
        <div class="resize-handle resize-handle-bottom-left" @mousedown="startResize($event, 'bottom-left')"></div>
        <div class="resize-handle resize-handle-bottom-right" @mousedown="startResize($event, 'bottom-right')"></div>
        
        <div class="modal-header">
          <h2>创建 PersistentVolume</h2>
          <button class="close-btn" @click="showCreateModal = false">&times;</button>
        </div>

        <!-- 创建模式切换 -->
        <div class="mode-tabs">
          <button 
            class="mode-tab" 
            :class="{ active: createMode === 'form' }" 
            @click="createMode = 'form'"
          >
            📝 表单模式
          </button>
          <button 
            class="mode-tab" 
            :class="{ active: createMode === 'yaml' }" 
            @click="createMode = 'yaml'"
          >
            📄 YAML 模式
          </button>
        </div>

        <!-- 表单模式 -->
        <div v-if="createMode === 'form'" class="modal-body">
          <div class="form-group">
            <label class="required">PV 名称</label>
            <input type="text" v-model="pvForm.name" class="form-input" placeholder="pv-001" required>
          </div>

          <div class="form-group">
            <label class="required">容量</label>
            <input type="text" v-model="pvForm.capacity" class="form-input" placeholder="10Gi" required>
            <span class="form-hint">示例: 1Gi, 10Gi, 100Gi</span>
          </div>

          <div class="form-group">
            <label class="required">访问模式</label>
            <div class="checkbox-group">
              <label class="checkbox-label">
                <input type="checkbox" value="ReadWriteOnce" v-model="pvForm.accessModes">
                ReadWriteOnce (RWO)
              </label>
              <label class="checkbox-label">
                <input type="checkbox" value="ReadOnlyMany" v-model="pvForm.accessModes">
                ReadOnlyMany (ROX)
              </label>
              <label class="checkbox-label">
                <input type="checkbox" value="ReadWriteMany" v-model="pvForm.accessModes">
                ReadWriteMany (RWX)
              </label>
            </div>
          </div>

          <div class="form-group">
            <label class="required">回收策略</label>
            <select v-model="pvForm.reclaimPolicy" class="form-select">
              <option value="Retain">Retain (保留)</option>
              <option value="Delete">Delete (删除)</option>
              <option value="Recycle">Recycle (回收)</option>
            </select>
            <span class="form-hint warning">⚠️ 推荐使用 Retain，避免数据丢失</span>
          </div>

          <div class="form-group">
            <label>StorageClass</label>
            <input type="text" v-model="pvForm.storageClassName" class="form-input" placeholder="留空表示不使用 StorageClass">
          </div>

          <div class="form-group">
            <label class="required">卷来源类型</label>
            <select v-model="pvForm.volumeSource.type" class="form-select">
              <option value="hostPath">HostPath (本地路径)</option>
              <option value="nfs">NFS (网络文件系统)</option>
            </select>
          </div>

          <div v-if="pvForm.volumeSource.type === 'hostPath'" class="form-group">
            <label class="required">HostPath 路径</label>
            <input type="text" v-model="pvForm.volumeSource.hostPath" class="form-input" placeholder="/mnt/data">
          </div>

          <div v-if="pvForm.volumeSource.type === 'nfs'">
            <div class="form-group">
              <label class="required">NFS 服务器</label>
              <input type="text" v-model="pvForm.volumeSource.nfsServer" class="form-input" placeholder="192.168.1.100">
            </div>
            <div class="form-group">
              <label class="required">NFS 路径</label>
              <input type="text" v-model="pvForm.volumeSource.nfsPath" class="form-input" placeholder="/exports/data">
            </div>
          </div>
        </div>

        <!-- YAML 模式 -->
        <div v-if="createMode === 'yaml'" class="modal-body">
          <!-- YAML 编辑器头部 -->
          <div class="yaml-editor-header">
            <h3>📄 YAML 配置</h3>
            <div class="yaml-header-buttons">
              <button class="load-template-btn" @click="loadPVYamlTemplate">
                📂 加载模板
              </button>
              <button class="clear-yaml-btn" @click="clearCreateYaml">
                🗑️ 清除
              </button>
            </div>
          </div>
          
          <div class="yaml-editor-wrapper">
            <textarea 
              v-model="createYamlContent" 
              class="yaml-editor"
              placeholder="请输入 PersistentVolume YAML 配置..."
              rows="20"
            ></textarea>
          </div>
          
          <div v-if="createYamlError" class="error-message">
            {{ createYamlError }}
          </div>
          
          <!-- YAML 预览区域 -->
          <div v-if="createYamlContent.trim()" class="yaml-preview-section">
            <div class="preview-header">
              <h4>🔍 YAML 预览</h4>
            </div>
            <div class="preview-content">
              <div class="preview-item">
                <span class="preview-label">📝 内容长度:</span>
                <span class="preview-value">{{ createYamlContent.length }} 字符</span>
              </div>
              <div class="preview-item">
                <span class="preview-label">📋 行数:</span>
                <span class="preview-value">{{ createYamlContent.split('\n').length }} 行</span>
              </div>
              <div class="preview-item">
                <span class="preview-label">✅ 基本检查:</span>
                <span class="preview-value" :class="validateYamlBasic() ? 'valid' : 'invalid'">
                  {{ validateYamlBasic() ? '包含必要字段' : '缺少必要字段' }}
                </span>
              </div>
            </div>
          </div>
          
          <!-- YAML 提示 -->
          <div class="yaml-editor-footer">
            <div class="yaml-tips">
              <strong>💡 提示：</strong>
              <ul>
                <li>支持完整的 Kubernetes PersistentVolume 配置</li>
                <li>可以通过“加载模板”获取示例 YAML</li>
                <li>创建前会验证 YAML 格式的正确性</li>
                <li>支持 HostPath、NFS、Ceph RBD 等多种存储类型</li>
              </ul>
            </div>
          </div>
        </div>

        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showCreateModal = false">取消</button>
          <button class="btn btn-primary" @click="createPV" :disabled="creating">
            {{ creating ? '创建中...' : '创建' }}
          </button>
        </div>
      </div>
    </div>

    <!-- YAML 查看/编辑模态框 -->
    <div v-if="showYamlModal" class="modal-overlay" @click="showYamlModal = false">
      <div class="modal-content yaml-modal" @click.stop>
        <div class="modal-header">
          <h2>YAML - {{ selectedPV?.name }}</h2>
          <div class="header-actions">
            <button 
              v-if="yamlMode === 'view'" 
              class="btn btn-sm btn-primary" 
              @click="yamlMode = 'edit'"
            >
              ✏️ 编辑
            </button>
            <button 
              v-if="yamlMode === 'edit'" 
              class="btn btn-sm btn-secondary" 
              @click="cancelYamlEdit"
            >
              取消
            </button>
            <button class="close-btn" @click="closeYamlModal">&times;</button>
          </div>
        </div>
        <div class="modal-body">
          <div class="yaml-viewer-wrapper">
            <textarea 
              v-if="yamlMode === 'edit'"
              v-model="editedYaml" 
              class="yaml-editor"
              rows="25"
            ></textarea>
            <pre v-else class="yaml-viewer"><code>{{ currentYaml }}</code></pre>
          </div>
          <div v-if="yamlError" class="error-message">
            {{ yamlError }}
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeYamlModal">关闭</button>
          <button 
            v-if="yamlMode === 'edit'" 
            class="btn btn-primary" 
            @click="applyYaml"
            :disabled="yamlSaving"
          >
            {{ yamlSaving ? '应用中...' : '应用' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 详情模态框 -->
    <div v-if="showDetailModal" class="modal-overlay" @click="showDetailModal = false">
      <div class="modal-content large" @click.stop>
        <div class="modal-header">
          <h2>PV 详情 - {{ selectedPV?.name }}</h2>
          <button class="close-btn" @click="showDetailModal = false">&times;</button>
        </div>
        <div class="modal-body">
          <div class="detail-section">
            <h3>基本信息</h3>
            <div class="detail-grid">
              <div class="detail-item">
                <span class="label">名称:</span>
                <span class="value">{{ selectedPV?.name }}</span>
              </div>
              <div class="detail-item">
                <span class="label">状态:</span>
                <span class="status-indicator" :class="getStatusClass(selectedPV?.status)">
                  {{ selectedPV?.status }}
                </span>
              </div>
              <div class="detail-item">
                <span class="label">容量:</span>
                <span class="value">{{ selectedPV?.capacity }}</span>
              </div>
              <div class="detail-item">
                <span class="label">访问模式:</span>
                <span class="value">{{ selectedPV?.accessModes }}</span>
              </div>
              <div class="detail-item">
                <span class="label">回收策略:</span>
                <span class="reclaim-badge" :class="getReclaimClass(selectedPV?.reclaimPolicy)">
                  {{ selectedPV?.reclaimPolicy }}
                </span>
              </div>
              <div class="detail-item">
                <span class="label">StorageClass:</span>
                <span class="value">{{ selectedPV?.storageClassName || '-' }}</span>
              </div>
              <div class="detail-item">
                <span class="label">绑定的 PVC:</span>
                <span class="value">{{ selectedPV?.claimRef || '-' }}</span>
              </div>
              <div class="detail-item">
                <span class="label">创建时间:</span>
                <span class="value">{{ selectedPV?.createdAt }}</span>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showDetailModal = false">关闭</button>
        </div>
      </div>
    </div>

    <!-- 删除确认模态框 -->
    <div v-if="showDeleteModal" class="modal-overlay" @click="showDeleteModal = false">
      <div class="modal-content modal-danger" @click.stop>
        <div class="modal-header">
          <h2>确认删除</h2>
          <button class="close-btn" @click="showDeleteModal = false">&times;</button>
        </div>
        <div class="modal-body">
          <div class="warning-content">
            <div class="warning-icon">⚠️</div>
            <p>确定要删除 PersistentVolume <strong>{{ selectedPV?.name }}</strong> 吗？</p>
            <p class="warning-text">此操作不可恢复！如果回收策略为 Delete，底层存储也会被删除。</p>
            <p v-if="selectedPV?.reclaimPolicy === 'Retain'" class="info-text">
              ✅ 当前回收策略为 Retain，底层存储数据将被保留。
            </p>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showDeleteModal = false">取消</button>
          <button class="btn btn-danger" @click="deletePV">确认删除</button>
        </div>
      </div>
    </div>

    <!-- 批量删除预览模态框 -->
    <div v-if="showBatchDeleteModal" class="modal-overlay" @click="showBatchDeleteModal = false">
      <div class="modal-content modal-danger" @click.stop>
        <div class="modal-header">
          <h2>批量删除确认</h2>
          <button class="close-btn" @click="showBatchDeleteModal = false">&times;</button>
        </div>
        <div class="modal-body">
          <div class="warning-content">
            <div class="warning-icon">⚠️</div>
            <p>即将删除 <strong>{{ selectedPVs.length }}</strong> 个 PersistentVolume：</p>
            <ul class="delete-list">
              <li v-for="pv in selectedPVs" :key="pv.name">
                {{ pv.name }}
                <span class="reclaim-badge" :class="getReclaimClass(pv.reclaimPolicy)">
                  {{ pv.reclaimPolicy }}
                </span>
              </li>
            </ul>
            <p class="warning-text">此操作不可恢复！请确认后再继续。</p>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showBatchDeleteModal = false">取消</button>
          <button class="btn btn-danger" @click="batchDeletePVs">确认批量删除</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import pvApi from '@/api/cluster/storage/pv'

// =============== 状态管理 ===============
const pvs = ref([])
const loading = ref(false)
const errorMsg = ref('')

// 搜索和过滤
const searchQuery = ref('')
const debouncedSearchQuery = ref('')
const statusFilter = ref('all')
const storageClassFilter = ref('')

// 视图模式
const viewMode = ref('table')

// 分页
const currentPage = ref(1)
const itemsPerPage = ref(10)
const jumpPage = ref(1)

// 批量操作
const batchMode = ref(false)
const selectedPVs = ref([])

// 自动刷新
const autoRefresh = ref(false)
const refreshTimer = ref(null)

// 模态框状态
const showCreateModal = ref(false)
const showYamlModal = ref(false)
const showDetailModal = ref(false)
const showDeleteModal = ref(false)
const showBatchDeleteModal = ref(false)
const creating = ref(false)
const yamlSaving = ref(false)

// 创建相关
const createMode = ref('form')
const createYamlContent = ref('')
const createYamlError = ref('')

// 选中的 PV
const selectedPV = ref(null)

// 创建表单
const pvForm = ref({
  name: '',
  capacity: '10Gi',
  accessModes: ['ReadWriteOnce'],
  reclaimPolicy: 'Retain',
  storageClassName: '',
  volumeSource: {
    type: 'hostPath',
    hostPath: '/mnt/data',
    nfsServer: '',
    nfsPath: '/'
  }
})

// YAML 查看/编辑
const currentYaml = ref('')
const editedYaml = ref('')
const yamlMode = ref('view')
const yamlError = ref('')

// 更多菜单状态
const activeMoreMenu = ref(null)

// =============== 模态框调整大小 ===============
const createModalRef = ref(null)
const createModalStyle = ref({
  width: '800px',
  maxWidth: '90vw',
  height: 'auto',
  maxHeight: '90vh'
})

const resizing = ref({
  isResizing: false,
  direction: '',
  startX: 0,
  startY: 0,
  startWidth: 0,
  startHeight: 0
})

const startResize = (event, direction) => {
  event.preventDefault()
  event.stopPropagation()
  
  const modal = createModalRef.value
  if (!modal) return
  
  const rect = modal.getBoundingClientRect()
  
  resizing.value = {
    isResizing: true,
    direction,
    startX: event.clientX,
    startY: event.clientY,
    startWidth: rect.width,
    startHeight: rect.height,
    startLeft: rect.left,
    startTop: rect.top
  }
  
  document.addEventListener('mousemove', handleResize)
  document.addEventListener('mouseup', stopResize)
  document.body.style.userSelect = 'none'
  document.body.style.cursor = getCursor(direction)
}

const handleResize = (event) => {
  if (!resizing.value.isResizing) return
  
  const deltaX = event.clientX - resizing.value.startX
  const deltaY = event.clientY - resizing.value.startY
  const direction = resizing.value.direction
  
  let newWidth = resizing.value.startWidth
  let newHeight = resizing.value.startHeight
  
  // 最小宽高限制
  const minWidth = 400
  const minHeight = 300
  const maxWidth = window.innerWidth * 0.95
  const maxHeight = window.innerHeight * 0.95
  
  // 根据拖拽方向调整大小
  if (direction.includes('right')) {
    newWidth = Math.min(Math.max(resizing.value.startWidth + deltaX, minWidth), maxWidth)
  }
  if (direction.includes('left')) {
    newWidth = Math.min(Math.max(resizing.value.startWidth - deltaX, minWidth), maxWidth)
  }
  if (direction.includes('bottom')) {
    newHeight = Math.min(Math.max(resizing.value.startHeight + deltaY, minHeight), maxHeight)
  }
  if (direction.includes('top')) {
    newHeight = Math.min(Math.max(resizing.value.startHeight - deltaY, minHeight), maxHeight)
  }
  
  createModalStyle.value = {
    width: `${newWidth}px`,
    height: `${newHeight}px`,
    maxWidth: '95vw',
    maxHeight: '95vh'
  }
}

const stopResize = () => {
  resizing.value.isResizing = false
  document.removeEventListener('mousemove', handleResize)
  document.removeEventListener('mouseup', stopResize)
  document.body.style.userSelect = ''
  document.body.style.cursor = ''
}

const getCursor = (direction) => {
  const cursors = {
    'top': 'n-resize',
    'bottom': 's-resize',
    'left': 'w-resize',
    'right': 'e-resize',
    'top-left': 'nw-resize',
    'top-right': 'ne-resize',
    'bottom-left': 'sw-resize',
    'bottom-right': 'se-resize'
  }
  return cursors[direction] || 'default'
}

// =============== 计算属性 ===============
const totalPages = computed(() => Math.ceil(filteredPVs.value.length / itemsPerPage.value) || 1)

// 智能页码显示
const visiblePages = computed(() => {
  const total = totalPages.value
  const current = currentPage.value
  const pages = []

  if (total <= 7) {
    for (let i = 1; i <= total; i++) {
      pages.push(i)
    }
  } else {
    if (current <= 4) {
      for (let i = 1; i <= 5; i++) {
        pages.push(i)
      }
      pages.push('...')
      pages.push(total)
    } else if (current >= total - 3) {
      pages.push(1)
      pages.push('...')
      for (let i = total - 4; i <= total; i++) {
        pages.push(i)
      }
    } else {
      pages.push(1)
      pages.push('...')
      for (let i = current - 1; i <= current + 1; i++) {
        pages.push(i)
      }
      pages.push('...')
      pages.push(total)
    }
  }

  return pages
})

const filteredPVs = computed(() => {
  let result = pvs.value

  // 搜索过滤
  if (debouncedSearchQuery.value) {
    const query = debouncedSearchQuery.value.toLowerCase()
    result = result.filter(pv => 
      pv.name.toLowerCase().includes(query) ||
      (pv.storageClassName && pv.storageClassName.toLowerCase().includes(query)) ||
      (pv.claimRef && pv.claimRef.toLowerCase().includes(query))
    )
  }

  // 状态过滤
  if (statusFilter.value && statusFilter.value !== 'all') {
    result = result.filter(pv => pv.status === statusFilter.value)
  }

  return result
})

const paginatedPVs = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage.value
  const end = start + itemsPerPage.value
  return filteredPVs.value.slice(start, end)
})

const isAllSelected = computed(() => {
  return paginatedPVs.value.length > 0 && 
         paginatedPVs.value.every(pv => selectedPVs.value.some(s => s.name === pv.name))
})

// =============== 搜索防抖 ===============
let searchTimeout = null
const onSearchInput = () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    debouncedSearchQuery.value = searchQuery.value
    currentPage.value = 1
  }, 300)
}

// =============== API 调用 ===============
const fetchList = async () => {
  loading.value = true
  errorMsg.value = ''
  try {
    const res = await pvApi.list({
      page: 1,
      limit: 1000
    })
    if (res.code === 0) {
      const list = res.data.list || res.data.items || []
      // 后端返回原生 K8s PV 对象，需要映射成前端所需的结构
      pvs.value = list.map(item => ({
        name: item.metadata?.name || item.name || '',
        status: item.status?.phase || 'Unknown',
        capacity: item.spec?.capacity?.storage || '-',
        accessModes: Array.isArray(item.spec?.accessModes) 
          ? item.spec.accessModes.join(', ') 
          : (item.spec?.accessModes || '-'),
        reclaimPolicy: item.spec?.persistentVolumeReclaimPolicy || 'Delete',
        storageClassName: item.spec?.storageClassName || '',
        claimRef: item.spec?.claimRef 
          ? `${item.spec.claimRef.namespace}/${item.spec.claimRef.name}` 
          : '',
        createdAt: item.metadata?.creationTimestamp 
          ? new Date(item.metadata.creationTimestamp).toLocaleString('zh-CN', {
              year: 'numeric',
              month: '2-digit',
              day: '2-digit',
              hour: '2-digit',
              minute: '2-digit',
              second: '2-digit'
            })
          : '-'
      }))
    } else {
      errorMsg.value = res.message || '获取 PV 列表失败'
    }
  } catch (error) {
    console.error('获取 PV 列表失败:', error)
    errorMsg.value = error.message || '网络错误'
  } finally {
    loading.value = false
  }
}

const refreshList = () => {
  fetchList()
}

// =============== 批量操作 ===============
const enterBatchMode = () => {
  batchMode.value = true
  selectedPVs.value = []
}

const exitBatchMode = () => {
  batchMode.value = false
  selectedPVs.value = []
}

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    selectedPVs.value = selectedPVs.value.filter(
      s => !paginatedPVs.value.some(pv => pv.name === s.name)
    )
  } else {
    paginatedPVs.value.forEach(pv => {
      if (!selectedPVs.value.some(s => s.name === pv.name)) {
        selectedPVs.value.push(pv)
      }
    })
  }
}

const isPVSelected = (pv) => {
  return selectedPVs.value.some(s => s.name === pv.name)
}

const togglePVSelection = (pv) => {
  const index = selectedPVs.value.findIndex(s => s.name === pv.name)
  if (index > -1) {
    selectedPVs.value.splice(index, 1)
  } else {
    selectedPVs.value.push(pv)
  }
}

const clearSelection = () => {
  selectedPVs.value = []
}

const openBatchDeletePreview = () => {
  showBatchDeleteModal.value = true
}

const batchDeletePVs = async () => {
  try {
    await pvApi.batchDelete(selectedPVs.value.map(pv => pv.name))
    showBatchDeleteModal.value = false
    selectedPVs.value = []
    batchMode.value = false
    refreshList()
  } catch (error) {
    console.error('批量删除失败:', error)
    alert('批量删除失败: ' + error.message)
  }
}

// =============== 过滤器 ===============
const setStatusFilter = (status) => {
  statusFilter.value = status
  currentPage.value = 1
}

// =============== 分页处理 ===============
const onPageSizeChange = () => {
  currentPage.value = 1
}

const goToPage = (page) => {
  if (page < 1 || page > totalPages.value) return
  currentPage.value = page
  jumpPage.value = page
}

const jumpToPage = () => {
  const page = parseInt(jumpPage.value)
  if (page >= 1 && page <= totalPages.value) {
    currentPage.value = page
  }
}

// =============== CRUD 操作 ===============
const createPV = async () => {
  creating.value = true
  createYamlError.value = ''
  try {
    if (createMode.value === 'form') {
      // 表单模式
      if (!pvForm.value.name || !pvForm.value.capacity || pvForm.value.accessModes.length === 0) {
        alert('请填写必填项')
        return
      }
      await pvApi.create(pvForm.value)
    } else {
      // YAML 模式
      if (!createYamlContent.value.trim()) {
        createYamlError.value = '请输入 YAML 内容'
        return
      }
      await pvApi.createFromYaml({ yaml: createYamlContent.value })
    }
    showCreateModal.value = false
    resetCreateForm()
    refreshList()
  } catch (error) {
    console.error('创建 PV 失败:', error)
    if (createMode.value === 'yaml') {
      createYamlError.value = error.message || '创建失败'
    } else {
      alert('创建失败: ' + error.message)
    }
  } finally {
    creating.value = false
  }
}

const resetCreateForm = () => {
  pvForm.value = {
    name: '',
    capacity: '10Gi',
    accessModes: ['ReadWriteOnce'],
    reclaimPolicy: 'Retain',
    storageClassName: '',
    volumeSource: {
      type: 'hostPath',
      hostPath: '/mnt/data',
      nfsServer: '',
      nfsPath: '/'
    }
  }
  createYamlContent.value = ''
  createYamlError.value = ''
}

const viewDetail = (pv) => {
  selectedPV.value = pv
  showDetailModal.value = true
}

const viewYaml = async (pv) => {
  try {
    const res = await pvApi.yaml({ name: pv.name })
    if (res.code === 0) {
      currentYaml.value = res.data.yaml
      editedYaml.value = res.data.yaml
      selectedPV.value = pv
      yamlMode.value = 'view'
      showYamlModal.value = true
    }
  } catch (error) {
    console.error('获取 YAML 失败:', error)
    alert('获取 YAML 失败: ' + error.message)
  }
}

const closeYamlModal = () => {
  showYamlModal.value = false
  yamlMode.value = 'view'
  yamlError.value = ''
}

const cancelYamlEdit = () => {
  yamlMode.value = 'view'
  editedYaml.value = currentYaml.value
  yamlError.value = ''
}

const applyYaml = async () => {
  yamlSaving.value = true
  yamlError.value = ''
  try {
    await pvApi.applyYaml({ yaml: editedYaml.value })
    showYamlModal.value = false
    refreshList()
  } catch (error) {
    console.error('应用 YAML 失败:', error)
    yamlError.value = error.message || '应用失败'
  } finally {
    yamlSaving.value = false
  }
}

const downloadYaml = async (pv) => {
  try {
    await pvApi.downloadYaml(pv.name)
  } catch (error) {
    console.error('下载 YAML 失败:', error)
    alert('下载失败: ' + error.message)
  }
}

const changeReclaimPolicy = async (pv) => {
  if (confirm(`确定要将 ${pv.name} 的回收策略改为 Retain 吗？`)) {
    try {
      await pvApi.reclaim({ name: pv.name, reclaimPolicy: 'Retain' })
      refreshList()
    } catch (error) {
      console.error('修改回收策略失败:', error)
      alert('修改失败: ' + error.message)
    }
  }
}

const confirmDelete = (pv) => {
  selectedPV.value = pv
  showDeleteModal.value = true
}

const deletePV = async () => {
  try {
    await pvApi.delete({ name: selectedPV.value.name })
    showDeleteModal.value = false
    selectedPV.value = null
    refreshList()
  } catch (error) {
    console.error('删除 PV 失败:', error)
    alert('删除失败: ' + error.message)
  }
}

// =============== 更多菜单 ===============
const toggleMoreMenu = (pv) => {
  if (activeMoreMenu.value === pv.name) {
    activeMoreMenu.value = null
  } else {
    activeMoreMenu.value = pv.name
  }
}

// 点击其他区域关闭菜单
const closeMoreMenu = () => {
  activeMoreMenu.value = null
}

// 监听全局点击事件关闭菜单
onMounted(() => {
  document.addEventListener('click', closeMoreMenu)
})

onUnmounted(() => {
  document.removeEventListener('click', closeMoreMenu)
})

// =============== 辅助方法 ===============
const getStatusClass = (status) => {
  if (!status) return ''
  const statusStr = typeof status === 'string' ? status : String(status)
  return statusStr.toLowerCase()
}

const getReclaimClass = (policy) => {
  const map = {
    'Retain': 'retain',
    'Delete': 'delete',
    'Recycle': 'recycle'
  }
  return map[policy] || ''
}

// =============== 自动刷新 ===============
watch(autoRefresh, (newVal) => {
  if (newVal) {
    refreshTimer.value = setInterval(() => {
      refreshList()
    }, 90000) // 90秒
  } else {
    if (refreshTimer.value) {
      clearInterval(refreshTimer.value)
      refreshTimer.value = null
    }
  }
})

// =============== 生命周期 ===============
onMounted(() => {
  fetchList()
  document.addEventListener('click', closeMoreMenu)
})

onUnmounted(() => {
  if (refreshTimer.value) {
    clearInterval(refreshTimer.value)
  }
  document.removeEventListener('click', closeMoreMenu)
})
</script>

<style scoped>
/* 基础布局 */
.resource-view {
  padding: 20px;
  background: #f5f7fa;
  min-height: 100vh;
}

.view-header {
  margin-bottom: 24px;
}

.view-header h1 {
  font-size: 28px;
  font-weight: 700;
  color: #2d3748;
  margin-bottom: 8px;
}

.view-header p {
  color: #718096;
  font-size: 14px;
}

/* 操作栏 */
.action-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 16px;
  background: white;
  padding: 16px;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.search-box {
  flex: 0 0 auto;
}

.search-box input {
  padding: 10px 16px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  width: 300px;
  transition: all 0.3s;
}

.search-box input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.filter-buttons {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.btn-filter {
  padding: 8px 16px;
  border: 1px solid #e2e8f0;
  background: white;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  transition: all 0.2s;
  color: #4a5568;
}

.btn-filter:hover {
  background: #f7fafc;
  border-color: #cbd5e0;
}

.btn-filter.active {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  color: white;
  border-color: #3b82f6;
  font-weight: 600;
}

.action-buttons {
  display: flex;
  gap: 12px;
  align-items: center;
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s;
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.btn-primary {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  color: white;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
}

.btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(59, 130, 246, 0.5);
}

.btn-secondary {
  background: white;
  color: #4a5568;
  border: 1px solid #e2e8f0;
}

.btn-secondary:hover {
  background: #f7fafc;
  border-color: #cbd5e0;
}

.btn-batch {
  background: linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%);
  color: white;
}

.btn-danger {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: white;
}

.view-toggle {
  display: flex;
  gap: 4px;
  background: #f7fafc;
  padding: 4px;
  border-radius: 6px;
}

.btn-view {
  padding: 6px 12px;
  background: transparent;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 16px;
  transition: all 0.2s;
}

.btn-view.active {
  background: white;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.auto-refresh-toggle {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  user-select: none;
}

.refresh-indicator {
  color: #10b981;
  animation: blink 1.5s infinite;
}

@keyframes blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}

/* 批量操作栏 */
.batch-action-bar {
  position: sticky;
  top: 0;
  z-index: 10;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 16px 20px;
  border-radius: 8px;
  margin-bottom: 20px;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
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
  padding: 4px 12px;
  background: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 4px;
  color: white;
  cursor: pointer;
  font-size: 12px;
}

.batch-actions {
  display: flex;
  gap: 12px;
}

.batch-btn {
  padding: 8px 16px;
  background: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 6px;
  color: white;
  cursor: pointer;
  font-size: 13px;
  transition: all 0.2s;
}

.batch-btn:hover {
  background: rgba(255, 255, 255, 0.3);
}

/* 表格视图 */
.table-container {
  background: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.resource-table {
  width: 100%;
  border-collapse: collapse;
}

.resource-table thead {
  background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
}

.resource-table th {
  padding: 16px;
  text-align: left;
  font-size: 12px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: #495057;
  border-bottom: 2px solid rgba(59, 130, 246, 0.2);
}

.resource-table td {
  padding: 16px;
  border-bottom: 1px solid #f1f3f5;
}

.resource-table tbody tr {
  transition: all 0.2s;
}

.resource-table tbody tr:hover {
  background: rgba(59, 130, 246, 0.03);
}

.row-selected {
  background: rgba(59, 130, 246, 0.08) !important;
}

/* 状态指示器 */
.status-indicator {
  display: inline-block;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.status-indicator.available {
  background: linear-gradient(135deg, #84fab0 0%, #8fd3f4 100%);
  color: white;
}

.status-indicator.bound {
  background: linear-gradient(135deg, #a8edea 0%, #fed6e3 100%);
  color: #0c5460;
}

.status-indicator.released {
  background: linear-gradient(135deg, #ffeaa7 0%, #fdcb6e 100%);
  color: #856404;
}

.status-indicator.failed {
  background: linear-gradient(135deg, #fa709a 0%, #fee140 100%);
  color: white;
}

.pv-name {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
}

.pv-name .icon {
  font-size: 18px;
}

.access-mode-badge,
.storage-class-badge {
  display: inline-block;
  padding: 2px 8px;
  background: #e2e8f0;
  color: #4a5568;
  border-radius: 4px;
  font-size: 12px;
}

.reclaim-badge {
  display: inline-block;
  padding: 4px 10px;
  border-radius: 8px;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
}

.reclaim-badge.retain {
  background: #d1fae5;
  color: #065f46;
}

.reclaim-badge.delete {
  background: #fee2e2;
  color: #991b1b;
}

.reclaim-badge.recycle {
  background: #dbeafe;
  color: #1e40af;
}

.claim-ref {
  color: #6b7280;
  font-size: 13px;
}

/* 操作按钮 */
.action-icons {
  display: flex;
  gap: 8px;
}

.action-btn {
  padding: 6px 12px;
  border: 1px solid #e2e8f0;
  background: white;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s;
}

.action-btn:hover {
  background: #f7fafc;
  border-color: #cbd5e0;
  transform: translateY(-2px);
}

.action-btn.danger:hover {
  background: #fee2e2;
  border-color: #fca5a5;
  color: #dc2626;
}

.action-btn.expand {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
  border: none;
}

.action-btn.expand:hover {
  background: linear-gradient(135deg, #059669 0%, #047857 100%);
  transform: translateY(-2px);
}

/* 更多菜单样式 */
.more-actions-wrapper {
  position: relative;
}

.more-btn {
  font-size: 20px;
  font-weight: bold;
  padding: 2px 8px;
}

.more-menu {
  position: absolute;
  right: 0;
  top: 100%;
  margin-top: 4px;
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  min-width: 150px;
  z-index: 1000;
  overflow: hidden;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 10px 16px;
  border: none;
  background: white;
  color: #2d3748;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
  text-align: left;
}

.menu-item:hover {
  background-color: #f7fafc;
  color: #326ce5;
}

.menu-icon {
  font-size: 16px;
}

/* 卡片视图 */
.card-container {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 20px;
}

.resource-card {
  background: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  transition: all 0.3s;
  border: 2px solid transparent;
}

.resource-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.card-selected {
  border-color: #3b82f6;
  background: rgba(59, 130, 246, 0.03);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid #f1f3f5;
  background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
}

.card-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
}

.card-checkbox {
  margin-right: 8px;
}

.card-body {
  padding: 16px;
}

.card-info {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.info-item .label {
  color: #6b7280;
  font-size: 13px;
}

.info-item .value {
  font-weight: 500;
  color: #2d3748;
}

.card-footer {
  display: flex;
  gap: 8px;
  padding: 12px 16px;
  border-top: 1px solid #f1f3f5;
  background: #f9fafb;
}

.card-action-btn {
  flex: 1;
  padding: 8px;
  border: 1px solid #e2e8f0;
  background: white;
  border-radius: 6px;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.2s;
}

.card-action-btn:hover {
  background: #f7fafc;
  transform: translateY(-1px);
}

.card-action-btn.danger:hover {
  background: #fee2e2;
  color: #dc2626;
}

/* 分页 */
.pagination-wrapper {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 20px;
  padding: 16px 20px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.pagination-btn {
  min-width: 36px;
  height: 36px;
  padding: 0 12px;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  background: white;
  cursor: pointer;
  transition: all 0.2s;
}

.pagination-btn:hover:not(:disabled) {
  background: #f3f4f6;
}

.pagination-btn.page-number.active {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  color: white;
  border: none;
}

.page-size-select {
  padding: 6px 10px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
}

.page-jump-input {
  width: 60px;
  padding: 6px 10px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  text-align: center;
}

/* 模态框 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  border-radius: 12px;
  width: 90%;
  max-width: 600px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.2);
}

/* 可调整大小的模态框 */
.modal-content.resizable-modal {
  position: relative;
  overflow: visible;
}

.modal-content.resizable-modal .modal-body {
  overflow-y: auto;
  max-height: calc(100% - 180px);
}

/* 拖拽手柄样式 */
.resize-handle {
  position: absolute;
  z-index: 10;
  background: transparent;
  transition: background 0.2s;
}

/* 上下左右边缘手柄 */
.resize-handle-top {
  top: 0;
  left: 0;
  right: 0;
  height: 6px;
  cursor: n-resize;
}

.resize-handle-bottom {
  bottom: 0;
  left: 0;
  right: 0;
  height: 6px;
  cursor: s-resize;
}

.resize-handle-left {
  top: 0;
  left: 0;
  bottom: 0;
  width: 6px;
  cursor: w-resize;
}

.resize-handle-right {
  top: 0;
  right: 0;
  bottom: 0;
  width: 6px;
  cursor: e-resize;
}

/* 四个角手柄 */
.resize-handle-top-left {
  top: 0;
  left: 0;
  width: 12px;
  height: 12px;
  cursor: nw-resize;
}

.resize-handle-top-right {
  top: 0;
  right: 0;
  width: 12px;
  height: 12px;
  cursor: ne-resize;
}

.resize-handle-bottom-left {
  bottom: 0;
  left: 0;
  width: 12px;
  height: 12px;
  cursor: sw-resize;
}

.resize-handle-bottom-right {
  bottom: 0;
  right: 0;
  width: 12px;
  height: 12px;
  cursor: se-resize;
}

/* 悬停高亮效果 */
.resize-handle:hover {
  background: rgba(59, 130, 246, 0.3);
}

.resize-handle-top:hover,
.resize-handle-bottom:hover {
  background: linear-gradient(to bottom, rgba(59, 130, 246, 0.5), rgba(59, 130, 246, 0.2));
}

.resize-handle-left:hover,
.resize-handle-right:hover {
  background: linear-gradient(to right, rgba(59, 130, 246, 0.5), rgba(59, 130, 246, 0.2));
}

.resize-handle-top-left:hover,
.resize-handle-top-right:hover,
.resize-handle-bottom-left:hover,
.resize-handle-bottom-right:hover {
  background: radial-gradient(circle, rgba(59, 130, 246, 0.6), transparent);
}

.modal-content.large {
  max-width: 900px;
}

.modal-content.yaml-modal {
  max-width: 1000px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #e5e7eb;
  background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
}

.modal-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 700;
}

.header-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.close-btn {
  background: none;
  border: none;
  font-size: 28px;
  cursor: pointer;
  color: #6b7280;
  line-height: 1;
}

.close-btn:hover {
  color: #374151;
}

.mode-tabs {
  display: flex;
  gap: 8px;
  padding: 16px 24px 0;
  border-bottom: 1px solid #e5e7eb;
}

.mode-tab {
  padding: 10px 20px;
  border: none;
  background: transparent;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  color: #6b7280;
  border-bottom: 2px solid transparent;
  transition: all 0.2s;
}

.mode-tab.active {
  color: #3b82f6;
  border-bottom-color: #3b82f6;
}

.modal-body {
  padding: 24px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 600;
  color: #374151;
  font-size: 14px;
}

.form-group label.required::after {
  content: ' *';
  color: #ef4444;
}

.form-input,
.form-select {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  transition: all 0.2s;
}

.form-input:focus,
.form-select:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-hint {
  display: block;
  margin-top: 6px;
  font-size: 12px;
  color: #6b7280;
}

.form-hint.warning {
  color: #f59e0b;
}

.checkbox-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-weight: normal;
}

.yaml-editor-wrapper,
.yaml-viewer-wrapper {
  border: 1px solid #d1d5db;
  border-radius: 6px;
  overflow: hidden;
}

.yaml-editor {
  width: 100%;
  padding: 16px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
  line-height: 1.6;
  border: none;
  resize: vertical;
}

.yaml-viewer {
  margin: 0;
  padding: 16px;
  background: #f8f9fa;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
  line-height: 1.6;
  overflow-x: auto;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 24px;
  border-top: 1px solid #e5e7eb;
}

/* 详情模态框 */
.detail-section {
  margin-bottom: 24px;
}

.detail-section h3 {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 16px;
  color: #2d3748;
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.detail-item .label {
  font-size: 12px;
  color: #6b7280;
  font-weight: 500;
}

.detail-item .value {
  font-size: 14px;
  color: #2d3748;
  font-weight: 600;
}

/* 删除确认 */
.modal-danger .modal-header {
  background: linear-gradient(135deg, #fee2e2 0%, #fecaca 100%);
}

.warning-content {
  text-align: center;
}

.warning-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.warning-text {
  color: #dc2626;
  font-weight: 600;
  margin-top: 12px;
}

.info-text {
  color: #10b981;
  font-weight: 500;
  margin-top: 12px;
}

.delete-list {
  margin: 16px 0;
  padding: 16px;
  background: #f9fafb;
  border-radius: 6px;
  max-height: 200px;
  overflow-y: auto;
}

.delete-list li {
  padding: 8px;
  border-bottom: 1px solid #e5e7eb;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

/* Info Box */
.info-box {
  background: #eff6ff;
  border-left: 4px solid #3b82f6;
  padding: 16px;
  border-radius: 6px;
  font-size: 14px;
  line-height: 1.8;
}

.info-box strong {
  color: #1e40af;
}

/* Warning Box */
.warning-box {
  background: #fef3c7;
  border-left: 4px solid #f59e0b;
  padding: 16px;
  border-radius: 6px;
  display: flex;
  gap: 12px;
  font-size: 14px;
  line-height: 1.6;
}

.warning-box .warning-icon {
  font-size: 24px;
  flex-shrink: 0;
}

.warning-box code {
  background: rgba(0, 0, 0, 0.1);
  padding: 2px 6px;
  border-radius: 3px;
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 13px;
}

/* 容量预览框 */
.capacity-preview-box {
  background: linear-gradient(135deg, #f0f9ff 0%, #e0f2fe 100%);
  border: 2px solid #3b82f6;
  border-radius: 8px;
  padding: 16px;
}

.capacity-header {
  font-size: 15px;
  font-weight: 600;
  color: #1e40af;
  margin-bottom: 16px;
  text-align: center;
}

.capacity-comparison {
  display: flex;
  align-items: center;
  justify-content: space-around;
  gap: 16px;
  margin-bottom: 12px;
}

.capacity-item {
  flex: 1;
  text-align: center;
  padding: 12px;
  border-radius: 6px;
  transition: all 0.3s;
}

.capacity-item.current {
  background: rgba(107, 114, 128, 0.1);
  border: 2px solid #9ca3af;
}

.capacity-item.new {
  background: linear-gradient(135deg, #d1fae5 0%, #a7f3d0 100%);
  border: 2px solid #10b981;
}

.capacity-label {
  font-size: 12px;
  color: #6b7280;
  font-weight: 600;
  text-transform: uppercase;
  margin-bottom: 8px;
  letter-spacing: 0.5px;
}

.capacity-value {
  font-size: 24px;
  font-weight: 700;
  color: #1f2937;
}

.capacity-value.placeholder {
  color: #9ca3af;
  font-style: italic;
  font-size: 18px;
}

.capacity-arrow {
  font-size: 32px;
  color: #10b981;
  font-weight: bold;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.6; transform: scale(1.1); }
}

.capacity-diff {
  text-align: center;
  padding: 10px;
  background: linear-gradient(135deg, #d1fae5 0%, #a7f3d0 100%);
  border-radius: 6px;
  border: 1px solid #10b981;
}

.diff-label {
  font-size: 13px;
  color: #065f46;
  font-weight: 600;
  margin-right: 8px;
}

.diff-value {
  font-size: 16px;
  font-weight: 700;
  color: #047857;
}

/* 错误消息 */
.error-box {
  background: #fee2e2;
  color: #991b1b;
  padding: 12px 16px;
  border-radius: 6px;
  margin-bottom: 16px;
  border-left: 4px solid #dc2626;
}

.error-message {
  background: #fee2e2;
  color: #991b1b;
  padding: 12px;
  border-radius: 6px;
  margin-top: 12px;
  font-size: 13px;
}

/* 加载和空状态 */
.loading-indicator {
  text-align: center;
  padding: 40px;
}

.loading-spinner {
  border: 3px solid rgba(59, 130, 246, 0.1);
  border-top-color: #3b82f6;
  border-radius: 50%;
  width: 40px;
  height: 40px;
  animation: spin 0.8s linear infinite;
  margin: 0 auto 16px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
}

.empty-icon {
  font-size: 64px;
  margin-bottom: 16px;
  opacity: 0.5;
}

.empty-text {
  font-size: 16px;
  color: #6b7280;
  margin-bottom: 24px;
}

/* 响应式 */
@media (max-width: 768px) {
  .action-bar {
    flex-direction: column;
    align-items: stretch;
  }

  .search-box input {
    width: 100%;
  }

  .filter-buttons {
    justify-content: center;
  }

  .action-buttons {
    justify-content: center;
    flex-wrap: wrap;
  }

  .table-container {
    overflow-x: auto;
  }

  .card-container {
    grid-template-columns: 1fr;
  }

  .detail-grid {
    grid-template-columns: 1fr;
  }
}
</style>
