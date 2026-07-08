<template>
  <div class="env-management">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <div class="header-icon">
          <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M12 2L2 7L12 12L22 7L12 2Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <path d="M2 17L12 22L22 17" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <path d="M2 12L12 17L22 12" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </div>
        <div class="header-text">
          <h1>K8s 环境管理</h1>
          <p class="header-desc">管理和监控多环境 Kubernetes 集群配置</p>
        </div>
      </div>
      <div class="header-actions">
        <button class="btn-icon" @click="loadEnvironments" title="刷新">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M23 4v6h-6M1 20v-6h6"/>
            <path d="M3.51 9a9 9 0 0114.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0020.49 15"/>
          </svg>
        </button>
        <button class="btn-primary-gradient" @click="showAddModal = true">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <line x1="12" y1="8" x2="12" y2="16"/>
            <line x1="8" y1="12" x2="16" y2="12"/>
          </svg>
          创建环境
        </button>
      </div>
    </div>

    <!-- 统计卡片区域 -->
    <div class="stats-grid">
      <div class="stat-card stat-total">
        <div class="stat-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="2" y="3" width="20" height="14" rx="2" ry="2"/>
            <line x1="8" y1="21" x2="16" y2="21"/>
            <line x1="12" y1="17" x2="12" y2="21"/>
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ environments.length }}</div>
          <div class="stat-label">环境总数</div>
        </div>
        <div class="stat-trend up">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="23 6 13.5 15.5 8.5 10.5 1 18"/>
            <polyline points="17 6 23 6 23 12"/>
          </svg>
        </div>
      </div>
      
      <div class="stat-card stat-online">
        <div class="stat-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M22 11.08V12a10 10 0 11-5.93-9.14"/>
            <polyline points="22 4 12 14.01 9 11.01"/>
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ connectedCount }}</div>
          <div class="stat-label">在线环境</div>
        </div>
        <div class="stat-badge success">运行中</div>
      </div>
      
      <div class="stat-card stat-offline">
        <div class="stat-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <line x1="15" y1="9" x2="9" y2="15"/>
            <line x1="9" y1="9" x2="15" y2="15"/>
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ disconnectedCount }}</div>
          <div class="stat-label">离线环境</div>
        </div>
        <div class="stat-badge warning">需关注</div>
      </div>
      
      <div class="stat-card stat-prod">
        <div class="stat-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ prodCount }}</div>
          <div class="stat-label">生产环境</div>
        </div>
        <div class="stat-badge danger">高优先级</div>
      </div>
    </div>

    <!-- 工具栏 -->
    <div class="toolbar-card">
      <div class="toolbar-left">
        <div class="search-wrapper">
          <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="8"/>
            <line x1="21" y1="21" x2="16.65" y2="16.65"/>
          </svg>
          <input
            v-model="searchQuery"
            placeholder="搜索环境名称、集群、命名空间..."
            class="search-input"
          />
          <button v-if="searchQuery" class="search-clear" @click="searchQuery = ''">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"/>
              <line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>
        
        <div class="filter-group">
          <select v-model="filterType" class="filter-select">
            <option value="">全部类型</option>
            <option value="development">开发环境</option>
            <option value="testing">测试环境</option>
            <option value="staging">预发布环境</option>
            <option value="production">生产环境</option>
          </select>
          
          <select v-model="filterStatus" class="filter-select">
            <option value="">全部状态</option>
            <option value="connected">在线</option>
            <option value="disconnected">离线</option>
          </select>
        </div>
      </div>
      
      <div class="toolbar-right">
        <div class="view-toggle">
          <button 
            class="toggle-btn" 
            :class="{ active: viewMode === 'table' }" 
            @click="viewMode = 'table'"
            title="表格视图"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="8" y1="6" x2="21" y2="6"/>
              <line x1="8" y1="12" x2="21" y2="12"/>
              <line x1="8" y1="18" x2="21" y2="18"/>
              <line x1="3" y1="6" x2="3.01" y2="6"/>
              <line x1="3" y1="12" x2="3.01" y2="12"/>
              <line x1="3" y1="18" x2="3.01" y2="18"/>
            </svg>
          </button>
          <button 
            class="toggle-btn" 
            :class="{ active: viewMode === 'card' }" 
            @click="viewMode = 'card'"
            title="卡片视图"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="3" width="7" height="7"/>
              <rect x="14" y="3" width="7" height="7"/>
              <rect x="14" y="14" width="7" height="7"/>
              <rect x="3" y="14" width="7" height="7"/>
            </svg>
          </button>
        </div>
      </div>
    </div>

    <!-- 表格视图 -->
    <div v-if="viewMode === 'table'" class="table-container">
      <table class="modern-table">
        <thead>
          <tr>
            <th class="th-checkbox">
              <input type="checkbox" class="custom-checkbox" />
            </th>
            <th>环境信息</th>
            <th>环境类型</th>
            <th>集群配置</th>
            <th>命名空间</th>
            <th>状态</th>
            <th>更新时间</th>
            <th class="th-actions">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="env in paginatedEnvironments" :key="env.id" class="table-row">
            <td class="td-checkbox">
              <input type="checkbox" class="custom-checkbox" />
            </td>
            <td>
              <div class="env-info">
                <div class="env-avatar" :class="`avatar-${env.type || env.envType || 'development'}`">
                  {{ (env.name || '').charAt(0).toUpperCase() }}
                </div>
                <div class="env-details">
                  <div class="env-name">{{ env.name }}</div>
                  <div class="env-desc">{{ env.description || '暂无描述' }}</div>
                </div>
              </div>
            </td>
            <td>
              <span class="type-badge" :class="`type-${env.type || env.envType || 'development'}`">
                <span class="type-dot"></span>
                {{ getEnvTypeName(env.type || env.envType) }}
              </span>
            </td>
            <td>
              <div class="cluster-info">
                <div class="cluster-name">{{ env.clusterName }}</div>
                <div class="cluster-url">{{ env.apiUrl }}</div>
              </div>
            </td>
            <td>
              <code class="namespace-tag">{{ env.namespace }}</code>
            </td>
            <td>
              <div class="status-wrapper">
                <span class="status-indicator" :class="`status-${env.status}`"></span>
                <span class="status-text" :class="`text-${env.status}`">
                  {{ env.status === 'connected' ? '在线' : '离线' }}
                </span>
              </div>
            </td>
            <td>
              <div class="time-info">
                <span class="time-relative">{{ formatTimeAgo(env.updatedAt) }}</span>
              </div>
            </td>
            <td>
              <div class="action-group">
                <button class="action-btn action-view" @click="handleView(env)" title="查看详情">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
                    <circle cx="12" cy="12" r="3"/>
                  </svg>
                </button>
                <button class="action-btn action-edit" @click="handleEdit(env)" title="编辑">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7"/>
                    <path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z"/>
                  </svg>
                </button>
                <button class="action-btn action-connect" @click="handleCheckConnection(env)" title="测试连接">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M23 4v6h-6M1 20v-6h6"/>
                    <path d="M3.51 9a9 9 0 0114.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0020.49 15"/>
                  </svg>
                </button>
                <button class="action-btn action-delete" @click="handleDelete(env)" title="删除">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="3 6 5 6 21 6"/>
                    <path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/>
                    <line x1="10" y1="11" x2="10" y2="17"/>
                    <line x1="14" y1="11" x2="14" y2="17"/>
                  </svg>
                </button>
              </div>
            </td>
          </tr>
          <tr v-if="paginatedEnvironments.length === 0">
            <td colspan="8" class="empty-state">
              <div class="empty-content">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                  <path d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>
                </svg>
                <p>暂无环境数据</p>
                <button class="btn-create-empty" @click="showAddModal = true">创建第一个环境</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 卡片视图 -->
    <div v-else class="cards-grid">
      <div v-for="env in paginatedEnvironments" :key="env.id" class="env-card" :class="`card-${env.type || env.envType || 'development'}`">
        <div class="card-header">
          <div class="card-avatar" :class="`avatar-${env.type || env.envType || 'development'}`">
            {{ (env.name || '').charAt(0).toUpperCase() }}
          </div>
          <div class="card-status">
            <span class="status-dot" :class="`dot-${env.status}`"></span>
            {{ env.status === 'connected' ? '在线' : '离线' }}
          </div>
        </div>
        <div class="card-body">
          <h3 class="card-title">{{ env.name }}</h3>
          <p class="card-desc">{{ env.description || '暂无描述' }}</p>
          <div class="card-meta">
            <div class="meta-item">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="2" y="3" width="20" height="14" rx="2" ry="2"/>
                <line x1="8" y1="21" x2="16" y2="21"/>
                <line x1="12" y1="17" x2="12" y2="21"/>
              </svg>
              <span>{{ env.clusterName }}</span>
            </div>
            <div class="meta-item">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z"/>
              </svg>
              <span>{{ env.namespace }}</span>
            </div>
          </div>
          <span class="card-type-tag" :class="`tag-${env.type || env.envType || 'development'}`">
            {{ getEnvTypeName(env.type || env.envType) }}
          </span>
        </div>
        <div class="card-footer">
          <button class="card-btn" @click="handleView(env)">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
              <circle cx="12" cy="12" r="3"/>
            </svg>
            查看
          </button>
          <button class="card-btn" @click="handleEdit(env)">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7"/>
              <path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z"/>
            </svg>
            编辑
          </button>
          <button class="card-btn danger" @click="handleDelete(env)">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="3 6 5 6 21 6"/>
              <path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/>
            </svg>
            删除
          </button>
        </div>
      </div>
      
      <!-- 空状态 -->
      <div v-if="paginatedEnvironments.length === 0" class="empty-card">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M12 2L2 7l10 5 10-5-10-5z"/>
          <path d="M2 17l10 5 10-5"/>
          <path d="M2 12l10 5 10-5"/>
        </svg>
        <h3>暂无环境</h3>
        <p>点击上方按钮创建第一个 K8s 环境</p>
      </div>
    </div>

    <!-- 分页 -->
    <div class="pagination-wrapper" v-if="filteredEnvironments.length > 0">
      <div class="pagination-info">
        共 <strong>{{ filteredEnvironments.length }}</strong> 条记录，
        当前第 <strong>{{ currentPage }}</strong> / <strong>{{ totalPages }}</strong> 页
      </div>
      <div class="pagination-controls">
        <button 
          class="page-btn" 
          :disabled="currentPage === 1" 
          @click="currentPage = 1"
          title="首页"
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="11 17 6 12 11 7"/>
            <polyline points="18 17 13 12 18 7"/>
          </svg>
        </button>
        <button 
          class="page-btn" 
          :disabled="currentPage === 1" 
          @click="currentPage--"
          title="上一页"
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="15 18 9 12 15 6"/>
          </svg>
        </button>
        
        <div class="page-numbers">
          <button 
            v-for="page in visiblePages" 
            :key="page" 
            class="page-num"
            :class="{ active: page === currentPage }"
            @click="currentPage = page"
          >
            {{ page }}
          </button>
        </div>
        
        <button 
          class="page-btn" 
          :disabled="currentPage === totalPages" 
          @click="currentPage++"
          title="下一页"
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="9 18 15 12 9 6"/>
          </svg>
        </button>
        <button 
          class="page-btn" 
          :disabled="currentPage === totalPages" 
          @click="currentPage = totalPages"
          title="末页"
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="13 17 18 12 13 7"/>
            <polyline points="6 17 11 12 6 7"/>
          </svg>
        </button>
        
        <select v-model="pageSize" class="page-size-select">
          <option :value="5">5 条/页</option>
          <option :value="10">10 条/页</option>
          <option :value="20">20 条/页</option>
          <option :value="50">50 条/页</option>
        </select>
      </div>
    </div>

    <!-- 创建/编辑模态框 -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showAddModal || showEditModal" class="modal-overlay" @click="closeModal">
          <div class="modal-container" @click.stop>
            <div class="modal-header">
              <div class="modal-title-group">
                <div class="modal-icon">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M12 2L2 7l10 5 10-5-10-5z"/>
                    <path d="M2 17l10 5 10-5"/>
                    <path d="M2 12l10 5 10-5"/>
                  </svg>
                </div>
                <div>
                  <h3>{{ showAddModal ? '创建 K8s 环境' : '编辑 K8s 环境' }}</h3>
                  <p>{{ showAddModal ? '配置新的 Kubernetes 环境连接' : '修改环境配置信息' }}</p>
                </div>
              </div>
              <button class="modal-close" @click="closeModal">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <line x1="18" y1="6" x2="6" y2="18"/>
                  <line x1="6" y1="6" x2="18" y2="18"/>
                </svg>
              </button>
            </div>
            
            <div class="modal-body">
              <form @submit.prevent="submitEnvironment" class="env-form">
                <div class="form-section">
                  <h4 class="section-title">基本信息</h4>
                  <div class="form-grid">
                    <div class="form-group">
                      <label class="form-label required">环境名称</label>
                      <input
                        type="text"
                        v-model="envForm.name"
                        required
                        class="form-input"
                        placeholder="例如: production-cluster"
                      />
                    </div>
                    <div class="form-group">
                      <label class="form-label required">环境类型</label>
                      <select v-model="envForm.type" required class="form-select">
                        <option value="development">开发环境</option>
                        <option value="testing">测试环境</option>
                        <option value="staging">预发布环境</option>
                        <option value="production">生产环境</option>
                      </select>
                    </div>
                  </div>
                  <div class="form-group">
                    <label class="form-label">描述</label>
                    <textarea
                      v-model="envForm.description"
                      rows="2"
                      class="form-textarea"
                      placeholder="简要描述此环境的用途..."
                    ></textarea>
                  </div>
                </div>
                
                <div class="form-section">
                  <h4 class="section-title">集群配置</h4>
                  <div class="form-grid">
                    <div class="form-group">
                      <label class="form-label required">集群名称</label>
                      <input
                        type="text"
                        v-model="envForm.clusterName"
                        required
                        class="form-input"
                        placeholder="例如: k8s-prod-cluster"
                      />
                    </div>
                    <div class="form-group">
                      <label class="form-label required">命名空间</label>
                      <input
                        type="text"
                        v-model="envForm.namespace"
                        required
                        class="form-input"
                        placeholder="例如: default"
                      />
                    </div>
                  </div>
                  <div class="form-group">
                    <label class="form-label required">API Server URL</label>
                    <input
                      type="url"
                      v-model="envForm.apiUrl"
                      required
                      class="form-input"
                      placeholder="https://kubernetes.example.com:6443"
                    />
                  </div>
                </div>
                
                <div class="form-actions">
                  <button type="button" class="btn-cancel" @click="closeModal">取消</button>
                  <button type="submit" class="btn-submit" :disabled="submitting">
                    <span v-if="submitting" class="spinner"></span>
                    {{ submitting ? '提交中...' : (showAddModal ? '创建环境' : '保存修改') }}
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- 查看详情模态框 -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showViewModal" class="modal-overlay" @click="closeModal">
          <div class="modal-container modal-lg" @click.stop>
            <div class="modal-header detail-header" :class="`header-${viewEnv.type || viewEnv.envType || 'development'}`">
              <div class="detail-header-content">
                <div class="detail-avatar" :class="`avatar-${viewEnv.type || viewEnv.envType || 'development'}`">
                  {{ (viewEnv.name || '').charAt(0).toUpperCase() }}
                </div>
                <div class="detail-title">
                  <h3>{{ viewEnv.name }}</h3>
                  <div class="detail-status">
                    <span class="status-dot" :class="`dot-${viewEnv.status}`"></span>
                    {{ viewEnv.status === 'connected' ? '在线运行中' : '离线' }}
                  </div>
                </div>
              </div>
              <button class="modal-close light" @click="closeModal">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <line x1="18" y1="6" x2="6" y2="18"/>
                  <line x1="6" y1="6" x2="18" y2="18"/>
                </svg>
              </button>
            </div>
            
            <div class="modal-body detail-body">
              <div class="detail-grid">
                <div class="detail-section">
                  <h4>
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <circle cx="12" cy="12" r="10"/>
                      <line x1="12" y1="16" x2="12" y2="12"/>
                      <line x1="12" y1="8" x2="12.01" y2="8"/>
                    </svg>
                    基本信息
                  </h4>
                  <div class="detail-items">
                    <div class="detail-item">
                      <span class="item-label">环境名称</span>
                      <span class="item-value">{{ viewEnv.name }}</span>
                    </div>
                    <div class="detail-item">
                      <span class="item-label">环境类型</span>
                      <span class="item-value">
                        <span class="type-badge" :class="`type-${viewEnv.type || viewEnv.envType || 'development'}`">
                          <span class="type-dot"></span>
                          {{ getEnvTypeName(viewEnv.type || viewEnv.envType) }}
                        </span>
                      </span>
                    </div>
                    <div class="detail-item full">
                      <span class="item-label">描述</span>
                      <span class="item-value">{{ viewEnv.description || '暂无描述' }}</span>
                    </div>
                  </div>
                </div>
                
                <div class="detail-section">
                  <h4>
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <rect x="2" y="3" width="20" height="14" rx="2" ry="2"/>
                      <line x1="8" y1="21" x2="16" y2="21"/>
                      <line x1="12" y1="17" x2="12" y2="21"/>
                    </svg>
                    集群配置
                  </h4>
                  <div class="detail-items">
                    <div class="detail-item">
                      <span class="item-label">集群名称</span>
                      <span class="item-value">{{ viewEnv.clusterName }}</span>
                    </div>
                    <div class="detail-item">
                      <span class="item-label">命名空间</span>
                      <span class="item-value"><code>{{ viewEnv.namespace }}</code></span>
                    </div>
                    <div class="detail-item full">
                      <span class="item-label">API Server</span>
                      <span class="item-value url">{{ viewEnv.apiUrl }}</span>
                    </div>
                  </div>
                </div>
                
                <div class="detail-section" v-if="viewEnv.createdAt || viewEnv.updatedAt">
                  <h4>
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <circle cx="12" cy="12" r="10"/>
                      <polyline points="12 6 12 12 16 14"/>
                    </svg>
                    时间信息
                  </h4>
                  <div class="detail-items">
                    <div class="detail-item">
                      <span class="item-label">创建时间</span>
                      <span class="item-value">{{ formatDate(viewEnv.createdAt) }}</span>
                    </div>
                    <div class="detail-item">
                      <span class="item-label">更新时间</span>
                      <span class="item-value">{{ formatDate(viewEnv.updatedAt) }}</span>
                    </div>
                  </div>
                </div>
              </div>
              
              <div class="detail-actions">
                <button class="btn-action edit" @click="closeModal(); handleEdit(viewEnv)">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7"/>
                    <path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z"/>
                  </svg>
                  编辑环境
                </button>
                <button class="btn-action connect" @click="handleCheckConnection(viewEnv)">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M23 4v6h-6M1 20v-6h6"/>
                    <path d="M3.51 9a9 9 0 0114.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0020.49 15"/>
                  </svg>
                  测试连接
                </button>
                <button class="btn-action close" @click="closeModal">关闭</button>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- 删除确认模态框 -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showDeleteModal" class="modal-overlay" @click="showDeleteModal = false">
          <div class="modal-container modal-sm" @click.stop>
            <div class="delete-modal-content">
              <div class="delete-icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <circle cx="12" cy="12" r="10"/>
                  <line x1="12" y1="8" x2="12" y2="12"/>
                  <line x1="12" y1="16" x2="12.01" y2="16"/>
                </svg>
              </div>
              <h3>确认删除</h3>
              <p>确定要删除环境 <strong>{{ deleteTarget?.name }}</strong> 吗？</p>
              <p class="delete-warning">此操作不可恢复，请谨慎操作。</p>
              <div class="delete-actions">
                <button class="btn-cancel" @click="showDeleteModal = false">取消</button>
                <button class="btn-delete" @click="confirmDelete" :disabled="deleting">
                  <span v-if="deleting" class="spinner"></span>
                  {{ deleting ? '删除中...' : '确认删除' }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Toast 通知 -->
    <Teleport to="body">
      <Transition name="toast">
        <div v-if="toast.show" class="toast" :class="`toast-${toast.type}`">
          <svg v-if="toast.type === 'success'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M22 11.08V12a10 10 0 11-5.93-9.14"/>
            <polyline points="22 4 12 14.01 9 11.01"/>
          </svg>
          <svg v-else-if="toast.type === 'error'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <line x1="15" y1="9" x2="9" y2="15"/>
            <line x1="9" y1="9" x2="15" y2="15"/>
          </svg>
          <span>{{ toast.message }}</span>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import {
  getK8sEnvironments,
  deleteK8sEnvironment,
  createK8sEnvironment,
  updateK8sEnvironment,
  getK8sEnvironmentDetail
} from '@/api/cicd.js'

export default {
  name: 'K8sEnvironments',
  setup() {
    const environments = ref([])
    const searchQuery = ref('')
    const filterType = ref('')
    const filterStatus = ref('')
    const currentPage = ref(1)
    const pageSize = ref(10)
    const viewMode = ref('table')

    // 模态框状态
    const showAddModal = ref(false)
    const showEditModal = ref(false)
    const showViewModal = ref(false)
    const showDeleteModal = ref(false)
    const submitting = ref(false)
    const deleting = ref(false)
    const deleteTarget = ref(null)

    // Toast
    const toast = ref({ show: false, type: 'success', message: '' })

    const showToast = (type, message) => {
      toast.value = { show: true, type, message }
      setTimeout(() => { toast.value.show = false }, 3000)
    }

    // 表单数据
    const envForm = ref({
      id: null,
      name: '',
      description: '',
      clusterName: '',
      apiUrl: '',
      namespace: 'default',
      type: 'development',
      status: 'disconnected'
    })

    const viewEnv = ref({})

    // 计算属性
    const connectedCount = computed(() => 
      environments.value.filter(e => e.status === 'connected').length
    )
    const disconnectedCount = computed(() => 
      environments.value.filter(e => e.status === 'disconnected').length
    )
    const prodCount = computed(() => 
      environments.value.filter(e => (e.type || e.envType) === 'production').length
    )

    const filteredEnvironments = computed(() => {
      return environments.value.filter(env => {
        const matchSearch = !searchQuery.value || 
          env.name?.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
          env.description?.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
          env.clusterName?.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
          env.namespace?.toLowerCase().includes(searchQuery.value.toLowerCase())
        const matchType = !filterType.value || (env.type || env.envType) === filterType.value
        const matchStatus = !filterStatus.value || env.status === filterStatus.value
        return matchSearch && matchType && matchStatus
      })
    })

    const totalPages = computed(() => Math.ceil(filteredEnvironments.value.length / pageSize.value) || 1)
    
    const visiblePages = computed(() => {
      const pages = []
      const total = totalPages.value
      const current = currentPage.value
      let start = Math.max(1, current - 2)
      let end = Math.min(total, start + 4)
      if (end - start < 4) start = Math.max(1, end - 4)
      for (let i = start; i <= end; i++) pages.push(i)
      return pages
    })

    const paginatedEnvironments = computed(() => {
      const start = (currentPage.value - 1) * pageSize.value
      return filteredEnvironments.value.slice(start, start + pageSize.value)
    })

    // 方法
    const loadEnvironments = async () => {
      try {
        const response = await getK8sEnvironments()
        if (response.code === 0) {
          environments.value = response.data || []
        } else {
          showToast('error', response.msg || '获取环境列表失败')
        }
      } catch (error) {
        showToast('error', '获取环境列表失败')
      }
    }

    const getEnvTypeName = (type) => {
      const typeMap = {
        production: '生产环境',
        staging: '预发布',
        'pre-production': '预发布',
        testing: '测试环境',
        development: '开发环境'
      }
      return typeMap[type] || type || '未知'
    }

    const formatDate = (dateString) => {
      if (!dateString) return '-'
      return new Date(dateString).toLocaleString('zh-CN')
    }

    const formatTimeAgo = (dateString) => {
      if (!dateString) return '-'
      const now = new Date()
      const date = new Date(dateString)
      const diff = Math.floor((now - date) / 1000)
      if (diff < 60) return '刚刚'
      if (diff < 3600) return `${Math.floor(diff / 60)} 分钟前`
      if (diff < 86400) return `${Math.floor(diff / 3600)} 小时前`
      if (diff < 2592000) return `${Math.floor(diff / 86400)} 天前`
      return formatDate(dateString)
    }

    const closeModal = () => {
      showAddModal.value = false
      showEditModal.value = false
      showViewModal.value = false
      envForm.value = {
        id: null, name: '', description: '', clusterName: '',
        apiUrl: '', namespace: 'default', type: 'development', status: 'disconnected'
      }
    }

    const handleView = async (env) => {
      try {
        const response = await getK8sEnvironmentDetail(env.id)
        if (response.code === 0) {
          viewEnv.value = response.data
          showViewModal.value = true
        } else {
          showToast('error', response.msg || '获取详情失败')
        }
      } catch (error) {
        showToast('error', '获取详情失败')
      }
    }

    const handleEdit = (env) => {
      envForm.value = { ...env, type: env.type || env.envType || 'development' }
      showEditModal.value = true
      showAddModal.value = false
    }

    const handleDelete = (env) => {
      deleteTarget.value = env
      showDeleteModal.value = true
    }

    const confirmDelete = async () => {
      if (!deleteTarget.value) return
      try {
        deleting.value = true
        const response = await deleteK8sEnvironment(deleteTarget.value.id)
        if (response.code === 0) {
          showToast('success', '删除成功')
          loadEnvironments()
          showDeleteModal.value = false
        } else {
          showToast('error', response.msg || '删除失败')
        }
      } catch (error) {
        showToast('error', '删除失败')
      } finally {
        deleting.value = false
      }
    }

    const handleCheckConnection = async (env) => {
      showToast('success', `正在测试连接 ${env.name}...`)
      // TODO: 实现连接测试逻辑
    }

    const submitEnvironment = async () => {
      try {
        submitting.value = true
        let response
        if (showAddModal.value) {
          response = await createK8sEnvironment(envForm.value)
        } else if (showEditModal.value && envForm.value.id) {
          response = await updateK8sEnvironment(envForm.value.id, envForm.value)
        }
        if (response?.code === 0) {
          showToast('success', showAddModal.value ? '创建成功' : '更新成功')
          loadEnvironments()
          closeModal()
        } else {
          showToast('error', response?.msg || '操作失败')
        }
      } catch (error) {
        showToast('error', '操作失败')
      } finally {
        submitting.value = false
      }
    }

    onMounted(() => loadEnvironments())

    return {
      environments, searchQuery, filterType, filterStatus,
      currentPage, pageSize, totalPages, visiblePages, viewMode,
      filteredEnvironments, paginatedEnvironments,
      connectedCount, disconnectedCount, prodCount,
      showAddModal, showEditModal, showViewModal, showDeleteModal,
      envForm, viewEnv, deleteTarget, submitting, deleting, toast,
      loadEnvironments, getEnvTypeName, formatDate, formatTimeAgo,
      closeModal, handleView, handleEdit, handleDelete, confirmDelete,
      handleCheckConnection, submitEnvironment
    }
  }
}
</script>

<style scoped>
/* ============ 基础变量 ============ */
.env-management {
  --primary: #6366f1;
  --primary-light: #818cf8;
  --primary-dark: #4f46e5;
  --success: #10b981;
  --warning: #f59e0b;
  --danger: #ef4444;
  --info: #3b82f6;
  
  --gray-50: #f9fafb;
  --gray-100: #f3f4f6;
  --gray-200: #e5e7eb;
  --gray-300: #d1d5db;
  --gray-400: #9ca3af;
  --gray-500: #6b7280;
  --gray-600: #4b5563;
  --gray-700: #374151;
  --gray-800: #1f2937;
  --gray-900: #111827;
  
  --dev-color: #10b981;
  --test-color: #3b82f6;
  --staging-color: #f59e0b;
  --prod-color: #ef4444;
  
  padding: 24px;
  background: var(--gray-50);
  min-height: 100vh;
}

/* ============ 页面头部 ============ */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding: 24px 28px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 16px;
  box-shadow: 0 10px 40px rgba(102, 126, 234, 0.3);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-icon {
  width: 56px;
  height: 56px;
  background: rgba(255,255,255,0.2);
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  backdrop-filter: blur(10px);
}

.header-icon svg {
  width: 32px;
  height: 32px;
  color: white;
}

.header-text h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 700;
  color: white;
  letter-spacing: -0.5px;
}

.header-desc {
  margin: 4px 0 0;
  font-size: 14px;
  color: rgba(255,255,255,0.8);
}

.header-actions {
  display: flex;
  gap: 12px;
}

.btn-icon {
  width: 44px;
  height: 44px;
  border: none;
  border-radius: 12px;
  background: rgba(255,255,255,0.2);
  color: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  backdrop-filter: blur(10px);
}

.btn-icon:hover {
  background: rgba(255,255,255,0.3);
  transform: translateY(-2px);
}

.btn-icon svg {
  width: 20px;
  height: 20px;
}

.btn-primary-gradient {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  border: none;
  border-radius: 12px;
  background: white;
  color: var(--primary);
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 4px 15px rgba(0,0,0,0.1);
}

.btn-primary-gradient:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(0,0,0,0.15);
}

.btn-primary-gradient svg {
  width: 18px;
  height: 18px;
}

/* ============ 统计卡片 ============ */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  margin-bottom: 24px;
}

.stat-card {
  background: white;
  border-radius: 16px;
  padding: 24px;
  display: flex;
  align-items: center;
  gap: 16px;
  box-shadow: 0 4px 20px rgba(0,0,0,0.05);
  transition: all 0.3s;
  position: relative;
  overflow: hidden;
}

.stat-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 4px;
  height: 100%;
}

.stat-total::before { background: var(--primary); }
.stat-online::before { background: var(--success); }
.stat-offline::before { background: var(--warning); }
.stat-prod::before { background: var(--danger); }

.stat-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 40px rgba(0,0,0,0.1);
}

.stat-icon {
  width: 52px;
  height: 52px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.stat-total .stat-icon { background: rgba(99,102,241,0.1); color: var(--primary); }
.stat-online .stat-icon { background: rgba(16,185,129,0.1); color: var(--success); }
.stat-offline .stat-icon { background: rgba(245,158,11,0.1); color: var(--warning); }
.stat-prod .stat-icon { background: rgba(239,68,68,0.1); color: var(--danger); }

.stat-icon svg {
  width: 26px;
  height: 26px;
}

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 32px;
  font-weight: 700;
  color: var(--gray-800);
  line-height: 1;
}

.stat-label {
  font-size: 13px;
  color: var(--gray-500);
  margin-top: 6px;
}

.stat-badge {
  font-size: 11px;
  padding: 4px 10px;
  border-radius: 20px;
  font-weight: 600;
}

.stat-badge.success { background: rgba(16,185,129,0.1); color: var(--success); }
.stat-badge.warning { background: rgba(245,158,11,0.1); color: var(--warning); }
.stat-badge.danger { background: rgba(239,68,68,0.1); color: var(--danger); }

/* ============ 工具栏 ============ */
.toolbar-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  background: white;
  border-radius: 12px;
  margin-bottom: 20px;
  box-shadow: 0 2px 12px rgba(0,0,0,0.04);
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.search-wrapper {
  position: relative;
  width: 320px;
}

.search-icon {
  position: absolute;
  left: 14px;
  top: 50%;
  transform: translateY(-50%);
  width: 18px;
  height: 18px;
  color: var(--gray-400);
  pointer-events: none;
}

.search-input {
  width: 100%;
  padding: 10px 40px 10px 42px;
  border: 2px solid var(--gray-200);
  border-radius: 10px;
  font-size: 14px;
  transition: all 0.2s;
  background: var(--gray-50);
}

.search-input:focus {
  outline: none;
  border-color: var(--primary);
  background: white;
  box-shadow: 0 0 0 4px rgba(99,102,241,0.1);
}

.search-clear {
  position: absolute;
  right: 10px;
  top: 50%;
  transform: translateY(-50%);
  width: 24px;
  height: 24px;
  border: none;
  background: var(--gray-200);
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--gray-500);
}

.search-clear svg {
  width: 14px;
  height: 14px;
}

.filter-group {
  display: flex;
  gap: 10px;
}

.filter-select {
  padding: 10px 32px 10px 14px;
  border: 2px solid var(--gray-200);
  border-radius: 10px;
  font-size: 14px;
  background: var(--gray-50) url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='%236b7280' stroke-width='2'%3E%3Cpolyline points='6 9 12 15 18 9'/%3E%3C/svg%3E") no-repeat right 12px center;
  appearance: none;
  cursor: pointer;
  transition: all 0.2s;
}

.filter-select:focus {
  outline: none;
  border-color: var(--primary);
}

.view-toggle {
  display: flex;
  background: var(--gray-100);
  border-radius: 8px;
  padding: 4px;
}

.toggle-btn {
  width: 36px;
  height: 36px;
  border: none;
  border-radius: 6px;
  background: transparent;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--gray-500);
  transition: all 0.2s;
}

.toggle-btn.active {
  background: white;
  color: var(--primary);
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
}

.toggle-btn svg {
  width: 18px;
  height: 18px;
}

/* ============ 表格 ============ */
.table-container {
  background: white;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(0,0,0,0.05);
}

.modern-table {
  width: 100%;
  border-collapse: collapse;
}

.modern-table thead {
  background: linear-gradient(135deg, var(--gray-800) 0%, var(--gray-700) 100%);
}

.modern-table th {
  padding: 16px 20px;
  text-align: left;
  font-size: 13px;
  font-weight: 600;
  color: white;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.th-checkbox, .td-checkbox {
  width: 50px;
  text-align: center !important;
}

.th-actions {
  width: 180px;
  text-align: center !important;
}

.custom-checkbox {
  width: 18px;
  height: 18px;
  cursor: pointer;
  accent-color: var(--primary);
}

.table-row {
  border-bottom: 1px solid var(--gray-100);
  transition: all 0.2s;
}

.table-row:hover {
  background: var(--gray-50);
}

.modern-table td {
  padding: 16px 20px;
  font-size: 14px;
  color: var(--gray-700);
  vertical-align: middle;
}

.env-info {
  display: flex;
  align-items: center;
  gap: 14px;
}

.env-avatar {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  font-weight: 700;
  color: white;
  flex-shrink: 0;
}

.avatar-development { background: linear-gradient(135deg, #10b981 0%, #059669 100%); }
.avatar-testing { background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%); }
.avatar-staging { background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%); }
.avatar-production { background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%); }

.env-details {
  min-width: 0;
}

.env-name {
  font-weight: 600;
  color: var(--gray-800);
  margin-bottom: 2px;
}

.env-desc {
  font-size: 12px;
  color: var(--gray-500);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 180px;
}

.type-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
}

.type-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.type-development { background: rgba(16,185,129,0.1); color: var(--dev-color); }
.type-development .type-dot { background: var(--dev-color); }
.type-testing { background: rgba(59,130,246,0.1); color: var(--test-color); }
.type-testing .type-dot { background: var(--test-color); }
.type-staging { background: rgba(245,158,11,0.1); color: var(--staging-color); }
.type-staging .type-dot { background: var(--staging-color); }
.type-production { background: rgba(239,68,68,0.1); color: var(--prod-color); }
.type-production .type-dot { background: var(--prod-color); }

.cluster-info {
  max-width: 200px;
}

.cluster-name {
  font-weight: 500;
  color: var(--gray-800);
  margin-bottom: 2px;
}

.cluster-url {
  font-size: 11px;
  color: var(--gray-400);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.namespace-tag {
  display: inline-block;
  padding: 4px 10px;
  background: var(--gray-100);
  border-radius: 6px;
  font-size: 12px;
  color: var(--gray-600);
  font-family: 'SF Mono', Monaco, monospace;
}

.status-wrapper {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-indicator {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  animation: pulse 2s infinite;
}

.status-connected { background: var(--success); }
.status-disconnected { background: var(--gray-400); animation: none; }

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.status-text {
  font-size: 13px;
  font-weight: 500;
}

.text-connected { color: var(--success); }
.text-disconnected { color: var(--gray-500); }

.time-info {
  font-size: 13px;
  color: var(--gray-500);
}

.action-group {
  display: flex;
  justify-content: center;
  gap: 6px;
}

.action-btn {
  width: 34px;
  height: 34px;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.action-btn svg {
  width: 16px;
  height: 16px;
}

.action-view { background: rgba(59,130,246,0.1); color: var(--info); }
.action-view:hover { background: var(--info); color: white; }
.action-edit { background: rgba(245,158,11,0.1); color: var(--warning); }
.action-edit:hover { background: var(--warning); color: white; }
.action-connect { background: rgba(16,185,129,0.1); color: var(--success); }
.action-connect:hover { background: var(--success); color: white; }
.action-delete { background: rgba(239,68,68,0.1); color: var(--danger); }
.action-delete:hover { background: var(--danger); color: white; }

.empty-state {
  text-align: center;
  padding: 60px 20px;
}

.empty-content {
  color: var(--gray-400);
}

.empty-content svg {
  width: 64px;
  height: 64px;
  margin-bottom: 16px;
}

.empty-content p {
  font-size: 16px;
  margin-bottom: 20px;
}

.btn-create-empty {
  padding: 10px 24px;
  background: var(--primary);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-create-empty:hover {
  background: var(--primary-dark);
}

/* ============ 卡片视图 ============ */
.cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 20px;
}

.env-card {
  background: white;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(0,0,0,0.05);
  transition: all 0.3s;
  border-top: 4px solid;
}

.card-development { border-top-color: var(--dev-color); }
.card-testing { border-top-color: var(--test-color); }
.card-staging { border-top-color: var(--staging-color); }
.card-production { border-top-color: var(--prod-color); }

.env-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 40px rgba(0,0,0,0.12);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 20px 0;
}

.card-avatar {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  font-weight: 700;
  color: white;
}

.card-status {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--gray-600);
}

.status-dot, .dot-connected, .dot-disconnected {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.dot-connected { background: var(--success); }
.dot-disconnected { background: var(--gray-400); }

.card-body {
  padding: 16px 20px;
}

.card-title {
  margin: 0 0 6px;
  font-size: 18px;
  font-weight: 600;
  color: var(--gray-800);
}

.card-desc {
  margin: 0 0 16px;
  font-size: 13px;
  color: var(--gray-500);
  line-height: 1.5;
}

.card-meta {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 16px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--gray-600);
}

.meta-item svg {
  width: 16px;
  height: 16px;
  color: var(--gray-400);
}

.card-type-tag {
  display: inline-block;
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
}

.tag-development { background: rgba(16,185,129,0.1); color: var(--dev-color); }
.tag-testing { background: rgba(59,130,246,0.1); color: var(--test-color); }
.tag-staging { background: rgba(245,158,11,0.1); color: var(--staging-color); }
.tag-production { background: rgba(239,68,68,0.1); color: var(--prod-color); }

.card-footer {
  display: flex;
  border-top: 1px solid var(--gray-100);
  padding: 12px;
  gap: 8px;
}

.card-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 10px;
  border: none;
  border-radius: 8px;
  background: var(--gray-100);
  color: var(--gray-700);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.card-btn:hover {
  background: var(--gray-200);
}

.card-btn.danger:hover {
  background: rgba(239,68,68,0.1);
  color: var(--danger);
}

.card-btn svg {
  width: 16px;
  height: 16px;
}

.empty-card {
  grid-column: 1 / -1;
  text-align: center;
  padding: 60px;
  background: white;
  border-radius: 16px;
  color: var(--gray-400);
}

.empty-card svg {
  width: 64px;
  height: 64px;
  margin-bottom: 16px;
}

.empty-card h3 {
  color: var(--gray-600);
  margin-bottom: 8px;
}

/* ============ 分页 ============ */
.pagination-wrapper {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  background: white;
  border-radius: 12px;
  margin-top: 20px;
  box-shadow: 0 2px 12px rgba(0,0,0,0.04);
}

.pagination-info {
  font-size: 14px;
  color: var(--gray-600);
}

.pagination-controls {
  display: flex;
  align-items: center;
  gap: 8px;
}

.page-btn {
  width: 36px;
  height: 36px;
  border: 1px solid var(--gray-200);
  border-radius: 8px;
  background: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--gray-600);
  transition: all 0.2s;
}

.page-btn:hover:not(:disabled) {
  border-color: var(--primary);
  color: var(--primary);
}

.page-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-btn svg {
  width: 16px;
  height: 16px;
}

.page-numbers {
  display: flex;
  gap: 4px;
}

.page-num {
  width: 36px;
  height: 36px;
  border: 1px solid var(--gray-200);
  border-radius: 8px;
  background: white;
  cursor: pointer;
  font-size: 14px;
  color: var(--gray-600);
  transition: all 0.2s;
}

.page-num:hover {
  border-color: var(--primary);
  color: var(--primary);
}

.page-num.active {
  background: var(--primary);
  border-color: var(--primary);
  color: white;
}

.page-size-select {
  padding: 8px 12px;
  border: 1px solid var(--gray-200);
  border-radius: 8px;
  font-size: 14px;
  margin-left: 12px;
  cursor: pointer;
}

/* ============ 模态框 ============ */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.5);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
}

.modal-container {
  background: white;
  border-radius: 20px;
  width: 100%;
  max-width: 560px;
  max-height: 90vh;
  overflow: hidden;
  box-shadow: 0 25px 80px rgba(0,0,0,0.25);
}

.modal-lg { max-width: 680px; }
.modal-sm { max-width: 420px; }

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24px 28px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.modal-title-group {
  display: flex;
  align-items: center;
  gap: 16px;
}

.modal-icon {
  width: 48px;
  height: 48px;
  background: rgba(255,255,255,0.2);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal-icon svg {
  width: 24px;
  height: 24px;
}

.modal-title-group h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.modal-title-group p {
  margin: 4px 0 0;
  font-size: 13px;
  opacity: 0.8;
}

.modal-close {
  width: 36px;
  height: 36px;
  border: none;
  border-radius: 10px;
  background: rgba(255,255,255,0.2);
  color: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.modal-close:hover {
  background: rgba(255,255,255,0.3);
  transform: rotate(90deg);
}

.modal-close svg {
  width: 20px;
  height: 20px;
}

.modal-close.light {
  background: rgba(0,0,0,0.2);
}

.modal-body {
  padding: 24px 28px;
  overflow-y: auto;
  max-height: calc(90vh - 180px);
}

/* 表单样式 */
.env-form {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.form-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.section-title {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--gray-800);
  padding-bottom: 12px;
  border-bottom: 1px solid var(--gray-100);
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--gray-700);
}

.form-label.required::after {
  content: ' *';
  color: var(--danger);
}

.form-input, .form-select, .form-textarea {
  padding: 12px 16px;
  border: 2px solid var(--gray-200);
  border-radius: 10px;
  font-size: 14px;
  transition: all 0.2s;
  background: var(--gray-50);
}

.form-input:focus, .form-select:focus, .form-textarea:focus {
  outline: none;
  border-color: var(--primary);
  background: white;
  box-shadow: 0 0 0 4px rgba(99,102,241,0.1);
}

.form-textarea {
  resize: vertical;
  min-height: 80px;
  font-family: inherit;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding-top: 16px;
  border-top: 1px solid var(--gray-100);
}

.btn-cancel {
  padding: 12px 24px;
  border: none;
  border-radius: 10px;
  background: var(--gray-100);
  color: var(--gray-700);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-cancel:hover {
  background: var(--gray-200);
}

.btn-submit {
  padding: 12px 28px;
  border: none;
  border-radius: 10px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn-submit:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(102,126,234,0.4);
}

.btn-submit:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.spinner {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255,255,255,0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* 详情模态框 */
.detail-header {
  padding: 32px 28px;
}

.header-development { background: linear-gradient(135deg, #10b981 0%, #059669 100%); }
.header-testing { background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%); }
.header-staging { background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%); }
.header-production { background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%); }

.detail-header-content {
  display: flex;
  align-items: center;
  gap: 20px;
}

.detail-avatar {
  width: 64px;
  height: 64px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  font-weight: 700;
  color: white;
  background: rgba(255,255,255,0.2);
}

.detail-title h3 {
  margin: 0;
  font-size: 22px;
  font-weight: 700;
}

.detail-status {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
  font-size: 14px;
  opacity: 0.9;
}

.detail-body {
  padding: 24px 28px;
}

.detail-grid {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.detail-section h4 {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 0 0 16px;
  font-size: 15px;
  font-weight: 600;
  color: var(--gray-800);
  padding-bottom: 12px;
  border-bottom: 1px solid var(--gray-100);
}

.detail-section h4 svg {
  width: 18px;
  height: 18px;
  color: var(--primary);
}

.detail-items {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.detail-item.full {
  grid-column: 1 / -1;
}

.item-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--gray-500);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.item-value {
  font-size: 14px;
  color: var(--gray-800);
}

.item-value code {
  padding: 4px 10px;
  background: var(--gray-100);
  border-radius: 6px;
  font-size: 13px;
}

.item-value.url {
  word-break: break-all;
  color: var(--primary);
}

.detail-actions {
  display: flex;
  gap: 12px;
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid var(--gray-100);
}

.btn-action {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 12px;
  border: none;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-action svg {
  width: 18px;
  height: 18px;
}

.btn-action.edit {
  background: rgba(245,158,11,0.1);
  color: var(--warning);
}

.btn-action.edit:hover {
  background: var(--warning);
  color: white;
}

.btn-action.connect {
  background: rgba(16,185,129,0.1);
  color: var(--success);
}

.btn-action.connect:hover {
  background: var(--success);
  color: white;
}

.btn-action.close {
  background: var(--gray-100);
  color: var(--gray-700);
}

.btn-action.close:hover {
  background: var(--gray-200);
}

/* 删除确认模态框 */
.delete-modal-content {
  text-align: center;
  padding: 40px 32px;
}

.delete-icon {
  width: 72px;
  height: 72px;
  margin: 0 auto 20px;
  background: rgba(239,68,68,0.1);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.delete-icon svg {
  width: 36px;
  height: 36px;
  color: var(--danger);
}

.delete-modal-content h3 {
  margin: 0 0 12px;
  font-size: 20px;
  color: var(--gray-800);
}

.delete-modal-content p {
  margin: 0;
  color: var(--gray-600);
  line-height: 1.6;
}

.delete-warning {
  margin-top: 12px !important;
  font-size: 13px;
  color: var(--danger) !important;
}

.delete-actions {
  display: flex;
  gap: 12px;
  margin-top: 28px;
}

.btn-delete {
  flex: 1;
  padding: 12px;
  border: none;
  border-radius: 10px;
  background: var(--danger);
  color: white;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  transition: all 0.2s;
}

.btn-delete:hover:not(:disabled) {
  background: #dc2626;
}

.btn-delete:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

/* ============ Toast ============ */
.toast {
  position: fixed;
  bottom: 32px;
  right: 32px;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 24px;
  border-radius: 12px;
  font-size: 14px;
  font-weight: 500;
  box-shadow: 0 10px 40px rgba(0,0,0,0.2);
  z-index: 2000;
}

.toast svg {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.toast-success {
  background: var(--success);
  color: white;
}

.toast-error {
  background: var(--danger);
  color: white;
}

/* ============ 动画 ============ */
.modal-enter-active,
.modal-leave-active {
  transition: all 0.3s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .modal-container,
.modal-leave-to .modal-container {
  transform: scale(0.95) translateY(-20px);
}

.toast-enter-active,
.toast-leave-active {
  transition: all 0.3s ease;
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateX(100px);
}

/* ============ 响应式 ============ */
@media (max-width: 1200px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .env-management {
    padding: 16px;
  }
  
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }
  
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .toolbar-card {
    flex-direction: column;
    gap: 16px;
  }
  
  .toolbar-left {
    flex-direction: column;
    width: 100%;
  }
  
  .search-wrapper {
    width: 100%;
  }
  
  .filter-group {
    width: 100%;
  }
  
  .filter-select {
    flex: 1;
  }
  
  .form-grid {
    grid-template-columns: 1fr;
  }
  
  .detail-items {
    grid-template-columns: 1fr;
  }
  
  .pagination-wrapper {
    flex-direction: column;
    gap: 16px;
  }
}
</style>
