<template>
  <div class="cr-instances-page">
    <!-- 页头区域 -->
    <div class="page-header-pro">
      <div class="header-content">
        <div class="header-left">
          <div class="title-group">
            <div class="title-badge">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/>
                <polyline points="14 2 14 8 20 8"/>
                <line x1="16" y1="13" x2="8" y2="13"/>
                <line x1="16" y1="17" x2="8" y2="17"/>
              </svg>
            </div>
            <div>
              <h1 class="page-title-pro">CR 实例管理</h1>
              <p class="page-subtitle">跨 CRD 浏览和管理所有自定义资源实例</p>
            </div>
          </div>
        </div>
        <div class="header-right">
          <button class="header-btn refresh-btn" @click="loadCRDList" :disabled="loadingCRDs" title="刷新 CRD 列表">
            <svg :class="{ spinning: loadingCRDs }" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="23 4 23 10 17 10"/><polyline points="1 20 1 14 7 14"/>
              <path d="M3.51 9a9 9 0 0114.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0020.49 15"/>
            </svg>
          </button>
        </div>
      </div>
    </div>

    <!-- 资源选择器面板 -->
    <div class="selector-panel-pro">
      <div class="selector-grid">
        <!-- CRD 类型选择 -->
        <div class="selector-col crd-col">
          <label class="field-label">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M3 9h18"/><path d="M9 21V9"/></svg>
            CRD 资源类型
          </label>
          <div class="crd-search-select" ref="crdSelectRef">
            <div class="crd-search-box" :class="{ focused: crdDropdownOpen }">
              <svg class="search-prefix-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
              <input
                v-model="crdSearchQuery"
                placeholder="搜索 CRD（Kind / Group）..."
                class="crd-search-input"
                @focus="crdDropdownOpen = true"
                @input="crdDropdownOpen = true"
              />
              <span v-if="crdSearchQuery" class="search-clear" @click.stop="crdSearchQuery = ''; crdDropdownOpen = true">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              </span>
            </div>
            <transition name="dropdown">
              <div v-if="crdDropdownOpen && (filteredCRDs.length > 0 || crdSearchQuery)" class="crd-dropdown-pro">
                <div class="dropdown-header" v-if="filteredCRDs.length > 0">
                  <span class="dropdown-count">{{ filteredCRDs.length }} 个资源类型</span>
                </div>
                <div class="dropdown-list" v-if="filteredCRDs.length > 0">
                  <div
                    v-for="crd in filteredCRDs"
                    :key="crd.name"
                    class="crd-option"
                    :class="{ active: crd.name === selectedCRD }"
                    @mousedown.prevent="selectCRD(crd)"
                  >
                    <div class="crd-option-left">
                      <span class="crd-kind-badge">{{ crd.kind?.charAt(0) }}</span>
                      <div class="crd-option-info">
                        <span class="crd-option-kind">{{ crd.kind }}</span>
                        <span class="crd-option-group">{{ crd.group }}/{{ crd.version }}</span>
                      </div>
                    </div>
                    <span class="crd-scope-mini" :class="crd.scope?.toLowerCase()">{{ crd.scope === 'Namespaced' ? 'NS' : 'C' }}</span>
                  </div>
                </div>
                <div v-else class="dropdown-empty">
                  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="10"/><path d="M8 15s1.5-2 4-2 4 2 4 2"/><line x1="9" y1="9" x2="9.01" y2="9"/><line x1="15" y1="9" x2="15.01" y2="9"/></svg>
                  <span>无匹配的 CRD</span>
                </div>
              </div>
            </transition>
          </div>
        </div>

        <!-- 命名空间 -->
        <div class="selector-col ns-col" v-if="currentCRD?.scope === 'Namespaced'">
          <label class="field-label">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z"/></svg>
            命名空间
          </label>
          <select v-model="namespace" @change="fetchInstances" class="select-pro">
            <option value="">所有命名空间</option>
            <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
          </select>
        </div>

        <!-- 搜索 -->
        <div class="selector-col search-col">
          <label class="field-label">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
            实例搜索
          </label>
          <div class="search-box-pro">
            <input v-model="searchQuery" placeholder="按名称筛选..." class="input-pro" />
          </div>
        </div>

        <!-- 操作 -->
        <div class="selector-col action-col">
          <label class="field-label">&nbsp;</label>
          <button class="btn-create-pro" @click="openCreate" :disabled="!selectedCRD">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            创建实例
          </button>
        </div>
      </div>

      <!-- 当前选中 CRD 信息条 -->
      <transition name="slide-fade">
        <div class="crd-meta-bar" v-if="currentCRD">
          <div class="meta-chips">
            <span class="meta-chip">
              <span class="chip-label">Kind</span>
              <span class="chip-value">{{ currentCRD.kind }}</span>
            </span>
            <span class="meta-chip">
              <span class="chip-label">Group</span>
              <span class="chip-value">{{ currentCRD.group }}</span>
            </span>
            <span class="meta-chip">
              <span class="chip-label">Version</span>
              <span class="chip-value">{{ currentCRD.version }}</span>
            </span>
            <span class="meta-chip">
              <span class="chip-label">Scope</span>
              <span class="chip-value scope" :class="currentCRD.scope?.toLowerCase()">{{ currentCRD.scope }}</span>
            </span>
            <span class="meta-chip">
              <span class="chip-label">Resource</span>
              <span class="chip-value mono">{{ currentCRD.resource }}</span>
            </span>
          </div>
          <div class="meta-stat" v-if="instances.length > 0">
            <span class="stat-num">{{ filteredInstances.length }}</span> 个实例
          </div>
        </div>
      </transition>
    </div>

    <!-- 未选择 CRD 引导 -->
    <div v-if="!selectedCRD" class="guide-state">
      <div class="guide-card">
        <div class="guide-illustration">
          <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1">
            <rect x="3" y="3" width="18" height="18" rx="2"/>
            <path d="M3 9h18"/><path d="M9 21V9"/>
          </svg>
        </div>
        <h3 class="guide-title">选择 CRD 资源类型</h3>
        <p class="guide-desc">从上方搜索框中选择一个 CustomResourceDefinition，<br/>即可浏览和管理其所有 CR 实例</p>
        <div class="guide-tips">
          <div class="tip-item">
            <span class="tip-num">1</span>
            <span>在搜索框中输入 Kind 或 Group 关键词</span>
          </div>
          <div class="tip-item">
            <span class="tip-num">2</span>
            <span>从下拉列表中选择目标 CRD</span>
          </div>
          <div class="tip-item">
            <span class="tip-num">3</span>
            <span>查看、创建、编辑或删除 CR 实例</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 加载中 -->
    <div v-else-if="loading" class="loading-state-pro">
      <div class="loader-ring"></div>
      <p class="loading-text">正在加载 <strong>{{ currentCRD?.kind }}</strong> 实例...</p>
    </div>

    <!-- 批量操作栏 -->
    <div v-else-if="filteredInstances.length > 0" class="batch-bar" :class="{ visible: selectedItems.length > 0 }">
      <div class="batch-left" v-if="selectedItems.length > 0">
        <span class="batch-count">已选择 <strong>{{ selectedItems.length }}</strong> 项</span>
        <button class="batch-btn danger" @click="batchDelete">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/></svg>
          批量删除
        </button>
        <button class="batch-btn ghost" @click="clearSelection">取消选择</button>
      </div>
    </div>

    <!-- 实例表格 -->
    <div v-if="selectedCRD && !loading && filteredInstances.length > 0" class="table-container">
      <table class="data-table-pro">
        <thead>
          <tr>
            <th class="col-check">
              <label class="checkbox-pro">
                <input type="checkbox" :checked="isAllSelected" @change="toggleSelectAll" />
                <span class="checkmark"></span>
              </label>
            </th>
            <th class="col-name">名称</th>
            <th v-if="currentCRD?.scope === 'Namespaced'" class="col-ns">命名空间</th>
            <th class="col-time">创建时间</th>
            <th class="col-labels">标签</th>
            <th class="col-actions">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in paginatedInstances" :key="item.uid || item.name" :class="{ selected: selectedItems.includes(item.name) }">
            <td class="col-check">
              <label class="checkbox-pro">
                <input type="checkbox" :checked="selectedItems.includes(item.name)" @change="toggleSelect(item)" />
                <span class="checkmark"></span>
              </label>
            </td>
            <td class="col-name">
              <div class="name-cell-pro">
                <div class="name-avatar">{{ item.name?.charAt(0)?.toUpperCase() }}</div>
                <div class="name-info">
                  <span class="name-primary">{{ item.name }}</span>
                  <span class="name-uid" v-if="item.uid">{{ item.uid?.slice(0, 8) }}</span>
                </div>
              </div>
            </td>
            <td v-if="currentCRD?.scope === 'Namespaced'" class="col-ns">
              <span class="ns-tag">{{ item.namespace }}</span>
            </td>
            <td class="col-time">
              <span class="time-text">{{ formatTime(item.createdAt) }}</span>
            </td>
            <td class="col-labels">
              <div class="label-group">
                <span v-for="(val, key) in limitLabels(item.labels)" :key="key" class="label-chip">
                  <span class="label-key">{{ key }}</span>
                  <span class="label-val">{{ val }}</span>
                </span>
                <span v-if="!item.labels || Object.keys(item.labels).length === 0" class="no-label">—</span>
                <span v-if="Object.keys(item.labels || {}).length > 3" class="label-overflow">+{{ Object.keys(item.labels).length - 3 }}</span>
              </div>
            </td>
            <td class="col-actions" @click.stop>
              <div class="actions-pro">
                <button class="act-btn act-yaml" @click="viewYaml(item)" title="查看 YAML">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
                  YAML
                </button>
                <button class="act-btn act-edit" @click="editInstance(item)" title="编辑">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                  编辑
                </button>
                <button class="act-btn act-del" @click="confirmDelete(item)" title="删除">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/></svg>
                  删除
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 空实例 -->
    <div v-if="selectedCRD && !loading && filteredInstances.length === 0" class="empty-state-pro">
      <div class="empty-visual">
        <svg width="56" height="56" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1">
          <rect x="2" y="3" width="20" height="14" rx="2"/>
          <line x1="8" y1="21" x2="16" y2="21"/>
          <line x1="12" y1="17" x2="12" y2="21"/>
        </svg>
      </div>
      <h3 class="empty-title">暂无 {{ currentCRD?.kind }} 实例</h3>
      <p class="empty-desc">{{ searchQuery ? '没有匹配的实例，请调整搜索条件' : '当前尚未创建任何实例，点击下方按钮开始' }}</p>
      <button v-if="!searchQuery" class="btn-create-pro sm" @click="openCreate">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
        创建第一个实例
      </button>
    </div>

    <!-- 分页栏 -->
    <div v-if="filteredInstances.length > pageSize" class="pagination-pro">
      <div class="page-left">
        <span class="page-total">共 <strong>{{ filteredInstances.length }}</strong> 条记录</span>
      </div>
      <div class="page-center">
        <button class="page-arrow" :disabled="currentPage <= 1" @click="currentPage--">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="15 18 9 12 15 6"/></svg>
        </button>
        <template v-for="p in displayPages" :key="p">
          <span v-if="p === '...'" class="page-ellipsis">...</span>
          <button v-else class="page-num" :class="{ active: p === currentPage }" @click="currentPage = p">{{ p }}</button>
        </template>
        <button class="page-arrow" :disabled="currentPage >= totalPages" @click="currentPage++">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="9 18 15 12 9 6"/></svg>
        </button>
      </div>
      <div class="page-right">
        <select v-model="pageSize" class="page-size-sel" @change="currentPage = 1">
          <option :value="10">10 条/页</option>
          <option :value="15">15 条/页</option>
          <option :value="30">30 条/页</option>
          <option :value="50">50 条/页</option>
        </select>
      </div>
    </div>

    <!-- YAML 抽屉 -->
    <Teleport to="body">
      <transition name="drawer">
        <div v-if="showDrawer" class="drawer-overlay-pro" @click="showDrawer = false">
          <div class="drawer-panel-pro" @click.stop>
            <div class="drawer-header-pro">
              <div class="drawer-title-area">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
                <h3>{{ drawerTitle }}</h3>
              </div>
              <button class="drawer-close-pro" @click="showDrawer = false">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              </button>
            </div>
            <div class="drawer-body-pro">
              <YamlHighlight :content="drawerYaml" :title="drawerTitle" :showLineNumbers="true" maxHeight="calc(100vh - 120px)" />
            </div>
          </div>
        </div>
      </transition>
    </Teleport>

    <!-- 编辑器模态框 -->
    <Teleport to="body">
      <transition name="modal">
        <div v-if="showEditor" class="modal-overlay-pro" @click="showEditor = false">
          <div class="editor-modal-pro" @click.stop>
            <div class="editor-topbar">
              <div class="editor-topbar-left">
                <span class="editor-mode-badge" :class="editMode">
                  {{ editMode === 'create' ? 'CREATE' : 'EDIT' }}
                </span>
                <h3 class="editor-title-pro">{{ editMode === 'create' ? `创建 ${currentCRD?.kind}` : `编辑 ${editTarget?.name}` }}</h3>
              </div>
              <button class="editor-close-btn" @click="showEditor = false">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              </button>
            </div>
            <div class="editor-info-bar">
              <span class="info-tag" v-if="currentCRD"><span class="info-tag-label">Kind</span><span class="info-tag-value">{{ currentCRD.kind }}</span></span>
              <span class="info-tag" v-if="currentCRD"><span class="info-tag-label">Group</span><span class="info-tag-value">{{ currentCRD.group }}</span></span>
              <span class="info-tag" v-if="currentCRD"><span class="info-tag-label">Version</span><span class="info-tag-value">{{ currentCRD.version }}</span></span>
            </div>
            <div class="editor-toolbar">
              <div class="editor-toolbar-left">
                <button class="toolbar-btn" @click="doDryRun" :disabled="dryRunning">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 3H5a2 2 0 00-2 2v4m6-6h10a2 2 0 012 2v4M9 3v18m0 0h10a2 2 0 002-2v-4M9 21H5a2 2 0 01-2-2v-4"/></svg>
                  {{ dryRunning ? '校验中...' : 'DryRun 预校验' }}
                </button>
                <button class="toolbar-btn secondary" @click="formatEditorYaml" title="格式化">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="21" y1="10" x2="3" y2="10"/><line x1="21" y1="6" x2="3" y2="6"/><line x1="21" y1="14" x2="3" y2="14"/><line x1="21" y1="18" x2="3" y2="18"/></svg>
                  格式化
                </button>
                <button class="toolbar-btn secondary" @click="copyEditorYaml" title="复制">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>
                  复制
                </button>
                <button class="toolbar-btn danger" @click="clearEditorYaml" title="清空">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/></svg>
                  清空
                </button>
              </div>
              <span class="toolbar-stat">{{ editorLineCount }} 行 · {{ editorYaml.length }} 字符</span>
            </div>
            <transition name="slide-down">
              <div v-if="dryRunResult" class="dryrun-banner" :class="dryRunResult.success ? 'success' : 'error'">
                <span>{{ dryRunResult.message }}</span>
                <button @click="dryRunResult = null">✕</button>
              </div>
            </transition>
            <div class="editor-workspace">
              <div class="editor-gutter">
                <div class="gutter-line" v-for="n in editorLineCount" :key="n">{{ n }}</div>
              </div>
              <textarea
                v-model="editorYaml"
                class="yaml-editor-pro"
                spellcheck="false"
                placeholder="# 在此输入 YAML..."
                @scroll="syncGutter"
                @keydown.tab.prevent="insertTab"
              ></textarea>
            </div>
            <div class="editor-footer-pro">
              <span class="footer-mode-tag">YAML</span>
              <div class="editor-footer-right">
                <button class="btn-cancel-pro" @click="showEditor = false">取消</button>
                <button class="btn-submit-pro" @click="submitEdit" :disabled="submitting">
                  {{ submitting ? '提交中...' : editMode === 'create' ? '确认创建' : '确认更新' }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </transition>
    </Teleport>

    <!-- 删除确认 -->
    <Teleport to="body">
      <transition name="modal">
        <div v-if="showDeleteDialog" class="modal-overlay-pro" @click="showDeleteDialog = false">
          <div class="delete-dialog-pro" @click.stop>
            <div class="delete-icon-pro">
              <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z"/>
                <line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/>
              </svg>
            </div>
            <h3 class="delete-title-pro">{{ batchMode ? '批量删除确认' : '确认删除' }}</h3>
            <p class="delete-desc-pro" v-if="!batchMode">确定要删除 <code>{{ deleteTarget?.name }}</code> 吗？</p>
            <p class="delete-desc-pro" v-else>确定要删除选中的 <strong>{{ selectedItems.length }}</strong> 个实例吗？</p>
            <p class="delete-warn-pro">此操作不可恢复，请谨慎操作。</p>
            <div class="delete-btns">
              <button class="dbtn cancel" @click="showDeleteDialog = false">取消</button>
              <button class="dbtn danger" @click="executeDelete" :disabled="deleting">
                {{ deleting ? '删除中...' : '确认删除' }}
              </button>
            </div>
          </div>
        </div>
      </transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'
import crdApi from '@/api/cluster/extensions/crd'
import YamlHighlight from '@/components/YamlHighlight.vue'

// ===== CRD 列表 =====
const crdList = ref([])
const loadingCRDs = ref(false)
const selectedCRD = ref('')
const crdSearchQuery = ref('')
const crdDropdownOpen = ref(false)
const crdSelectRef = ref(null)
const currentCRD = computed(() => crdList.value.find(c => c.name === selectedCRD.value))
const filteredCRDs = computed(() => {
  if (!crdSearchQuery.value) return crdList.value
  const q = crdSearchQuery.value.toLowerCase()
  return crdList.value.filter(c =>
    c.kind?.toLowerCase().includes(q) ||
    c.group?.toLowerCase().includes(q) ||
    c.name?.toLowerCase().includes(q) ||
    c.resource?.toLowerCase().includes(q)
  )
})
const selectCRD = (crd) => {
  selectedCRD.value = crd.name
  crdSearchQuery.value = `${crd.kind} (${crd.group}/${crd.version})`
  crdDropdownOpen.value = false
  selectedItems.value = []
  onCRDChange()
}

// 点击外部关闭下拉
const handleClickOutside = (e) => {
  if (crdSelectRef.value && !crdSelectRef.value.contains(e.target)) {
    crdDropdownOpen.value = false
  }
}
onMounted(() => { document.addEventListener('click', handleClickOutside) })
onBeforeUnmount(() => { document.removeEventListener('click', handleClickOutside) })

// ===== CR 实例 =====
const instances = ref([])
const loading = ref(false)
const namespace = ref('')
const namespaces = ref([])
const searchQuery = ref('')
const currentPage = ref(1)
const pageSize = ref(15)

// ===== 批量选择 =====
const selectedItems = ref([])
const batchMode = ref(false)
const isAllSelected = computed(() => {
  if (paginatedInstances.value.length === 0) return false
  return paginatedInstances.value.every(i => selectedItems.value.includes(i.name))
})
const toggleSelectAll = () => {
  if (isAllSelected.value) {
    const names = paginatedInstances.value.map(i => i.name)
    selectedItems.value = selectedItems.value.filter(n => !names.includes(n))
  } else {
    const names = paginatedInstances.value.map(i => i.name)
    selectedItems.value = [...new Set([...selectedItems.value, ...names])]
  }
}
const toggleSelect = (item) => {
  const idx = selectedItems.value.indexOf(item.name)
  if (idx >= 0) selectedItems.value.splice(idx, 1)
  else selectedItems.value.push(item.name)
}
const clearSelection = () => { selectedItems.value = [] }
const batchDelete = () => {
  batchMode.value = true
  showDeleteDialog.value = true
}

// ===== 编辑器 =====
const showEditor = ref(false)
const editMode = ref('create')
const editTarget = ref(null)
const editorYaml = ref('')
const submitting = ref(false)
const dryRunning = ref(false)
const dryRunResult = ref(null)

// ===== 抽屉 =====
const showDrawer = ref(false)
const drawerTitle = ref('')
const drawerYaml = ref('')

// ===== 删除 =====
const showDeleteDialog = ref(false)
const deleteTarget = ref(null)
const deleting = ref(false)

// ===== 计算属性 =====
const filteredInstances = computed(() => {
  if (!searchQuery.value) return instances.value
  return instances.value.filter(i => i.name?.toLowerCase().includes(searchQuery.value.toLowerCase()))
})
const totalPages = computed(() => Math.ceil(filteredInstances.value.length / pageSize.value))
const paginatedInstances = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  return filteredInstances.value.slice(start, start + pageSize.value)
})
const displayPages = computed(() => {
  const total = totalPages.value
  const current = currentPage.value
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1)
  const pages = []
  pages.push(1)
  if (current > 3) pages.push('...')
  for (let i = Math.max(2, current - 1); i <= Math.min(total - 1, current + 1); i++) pages.push(i)
  if (current < total - 2) pages.push('...')
  pages.push(total)
  return pages
})
const editorLineCount = computed(() => Math.max((editorYaml.value || '').split('\n').length, 1))

// ===== 加载 CRD 列表 =====
const loadCRDList = async () => {
  loadingCRDs.value = true
  try {
    const res = await crdApi.listCRDs({})
    if (res?.code === 0) {
      const items = res.data?.list || res.data?.items || []
      crdList.value = items.map(item => ({
        name: item.name,
        group: item.group,
        version: item.version,
        kind: item.kind,
        scope: item.scope,
        resource: item.resource || item.kind?.toLowerCase() + 's'
      }))
    }
  } catch (e) { console.error(e) }
  finally { loadingCRDs.value = false }
}

// ===== 选择 CRD =====
const onCRDChange = () => {
  namespace.value = ''
  instances.value = []
  namespaces.value = []
  currentPage.value = 1
  if (selectedCRD.value) fetchInstances()
}

// ===== 加载实例 =====
const fetchInstances = async () => {
  if (!currentCRD.value) return
  loading.value = true
  try {
    const params = {
      group: currentCRD.value.group,
      version: currentCRD.value.version,
      resource: currentCRD.value.resource
    }
    if (namespace.value) params.namespace = namespace.value
    const res = await crdApi.listCRs(params)
    if (res?.code === 0) {
      const items = res.data?.list || res.data?.items || []
      instances.value = items.map(item => {
        const meta = item.metadata || {}
        return {
          name: item.name || meta.name || '',
          namespace: item.namespace || meta.namespace || '',
          createdAt: item.created_at || meta.creationTimestamp || '-',
          labels: item.labels || meta.labels || {},
          uid: item.uid || meta.uid || '',
          resourceVersion: item.resource_version || meta.resourceVersion || ''
        }
      })
      const nsSet = new Set(instances.value.map(i => i.namespace).filter(Boolean))
      namespaces.value = [...nsSet].sort()
    }
  } catch (e) { console.error(e) }
  finally { loading.value = false }
}

// ===== YAML 查看 =====
const viewYaml = async (item) => {
  drawerTitle.value = item.name
  drawerYaml.value = '# 加载中...'
  showDrawer.value = true
  try {
    const res = await crdApi.getCRYaml({
      group: currentCRD.value.group,
      version: currentCRD.value.version,
      resource: currentCRD.value.resource,
      namespace: item.namespace,
      name: item.name
    })
    drawerYaml.value = res?.code === 0 ? (res.data?.yaml || '') : `# 获取失败: ${res?.msg}`
  } catch (e) { drawerYaml.value = `# 错误: ${e?.message}` }
}

// ===== 创建/编辑 =====
const openCreate = () => {
  if (!currentCRD.value) return
  editMode.value = 'create'
  editTarget.value = null
  dryRunResult.value = null
  const plural = currentCRD.value.resource
  editorYaml.value = `apiVersion: ${currentCRD.value.group}/${currentCRD.value.version}
kind: ${currentCRD.value.kind}
metadata:
  name: my-${plural.slice(0, -1) || 'resource'}-01
  namespace: default
spec:
  # 在此定义资源规格
`
  showEditor.value = true
}

const editInstance = async (item) => {
  editMode.value = 'edit'
  editTarget.value = item
  dryRunResult.value = null
  editorYaml.value = '# 加载中...'
  showEditor.value = true
  try {
    const res = await crdApi.getCRYaml({
      group: currentCRD.value.group,
      version: currentCRD.value.version,
      resource: currentCRD.value.resource,
      namespace: item.namespace,
      name: item.name
    })
    editorYaml.value = res?.code === 0 ? (res.data?.yaml || '') : `# 获取失败: ${res?.msg}`
  } catch (e) { editorYaml.value = `# 错误: ${e?.message}` }
}

const submitEdit = async () => {
  if (!editorYaml.value.trim()) return
  submitting.value = true
  try {
    let res
    if (editMode.value === 'create') {
      res = await crdApi.createCR({
        group: currentCRD.value.group,
        version: currentCRD.value.version,
        resource: currentCRD.value.resource,
        namespace: namespace.value || undefined,
        yaml: editorYaml.value
      })
    } else {
      res = await crdApi.updateCR({
        group: currentCRD.value.group,
        version: currentCRD.value.version,
        resource: currentCRD.value.resource,
        namespace: editTarget.value?.namespace,
        name: editTarget.value?.name,
        yaml: editorYaml.value
      })
    }
    if (res?.code === 0) {
      showEditor.value = false
      await fetchInstances()
    } else {
      dryRunResult.value = { success: false, message: res?.msg || '操作失败' }
    }
  } catch (e) {
    const details = Array.isArray(e?.details) ? e.details[0] : null
    dryRunResult.value = { success: false, message: details || e?.msg || e?.response?.data?.msg || e?.message || '操作失败' }
  } finally { submitting.value = false }
}

// ===== DryRun =====
const doDryRun = async () => {
  if (!editorYaml.value.trim()) return
  dryRunning.value = true
  dryRunResult.value = null
  try {
    const res = await crdApi.dryRun({
      group: currentCRD.value.group,
      version: currentCRD.value.version,
      resource: currentCRD.value.resource,
      namespace: editTarget.value?.namespace || namespace.value || undefined,
      name: editTarget.value?.name,
      yaml: editorYaml.value,
      is_update: editMode.value === 'edit'
    })
    if (res?.code === 0) {
      const dr = res.data
      if (dr?.valid) {
        dryRunResult.value = { success: true, message: dr.message || '✓ 校验通过，可安全提交' }
      } else {
        const errMsg = (dr?.errors && dr.errors.length > 0) ? dr.errors.join('; ') : (dr?.message || '校验失败')
        dryRunResult.value = { success: false, message: errMsg }
      }
    } else {
      dryRunResult.value = { success: false, message: res?.msg || '校验失败' }
    }
  } catch (e) {
    const details = Array.isArray(e?.details) ? e.details[0] : null
    dryRunResult.value = { success: false, message: details || e?.msg || e?.response?.data?.msg || e?.message || '校验失败' }
  } finally { dryRunning.value = false }
}

// ===== 删除 =====
const confirmDelete = (item) => {
  deleteTarget.value = item
  batchMode.value = false
  showDeleteDialog.value = true
}

const executeDelete = async () => {
  deleting.value = true
  try {
    if (batchMode.value) {
      // 批量删除
      for (const name of selectedItems.value) {
        const item = instances.value.find(i => i.name === name)
        if (!item) continue
        await crdApi.deleteCR({
          group: currentCRD.value.group,
          version: currentCRD.value.version,
          resource: currentCRD.value.resource,
          namespace: item.namespace,
          name: item.name
        })
      }
      selectedItems.value = []
      showDeleteDialog.value = false
      await fetchInstances()
    } else {
      if (!deleteTarget.value) return
      const res = await crdApi.deleteCR({
        group: currentCRD.value.group,
        version: currentCRD.value.version,
        resource: currentCRD.value.resource,
        namespace: deleteTarget.value.namespace,
        name: deleteTarget.value.name
      })
      if (res?.code === 0) { showDeleteDialog.value = false; await fetchInstances() }
      else alert(res?.msg || '删除失败')
    }
  } catch (e) { alert(e?.message || '删除失败') }
  finally { deleting.value = false }
}

// ===== 工具函数 =====
const formatTime = (t) => {
  if (!t || t === '-') return '-'
  try { return new Date(t).toLocaleString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' }) }
  catch { return t }
}
const limitLabels = (labels) => { if (!labels) return {}; const keys = Object.keys(labels).slice(0, 3); const out = {}; keys.forEach(k => out[k] = labels[k]); return out }
const syncGutter = (e) => { const g = e.target?.parentElement?.querySelector('.editor-gutter'); if (g) g.scrollTop = e.target.scrollTop }
const insertTab = (e) => { const ta = e.target; const s = ta.selectionStart; editorYaml.value = editorYaml.value.substring(0, s) + '  ' + editorYaml.value.substring(ta.selectionEnd); nextTick(() => { ta.selectionStart = ta.selectionEnd = s + 2 }) }

// ===== 编辑器工具按钮 =====
const formatEditorYaml = () => {
  editorYaml.value = editorYaml.value.split('\n').map(l => l.trimEnd()).join('\n').trim() + '\n'
}
const copyEditorYaml = async () => {
  try {
    await navigator.clipboard.writeText(editorYaml.value)
    dryRunResult.value = { success: true, message: '已复制到剪贴板' }
    setTimeout(() => { if (dryRunResult.value?.message === '已复制到剪贴板') dryRunResult.value = null }, 2000)
  } catch {
    dryRunResult.value = { success: false, message: '复制失败，请手动选择复制' }
  }
}
const clearEditorYaml = () => {
  editorYaml.value = ''
  dryRunResult.value = null
}

onMounted(() => { loadCRDList() })
</script>

<style scoped>
.cr-instances-page {
  --primary: #6366f1;
  --primary-light: #818cf8;
  --primary-bg: rgba(99,102,241,0.06);
  --primary-border: rgba(99,102,241,0.2);
  --danger: #ef4444;
  --danger-bg: rgba(239,68,68,0.06);
  --success: #10b981;
  --success-bg: rgba(16,185,129,0.06);
  --bg-page: #f8fafc;
  --bg-card: #ffffff;
  --bg-hover: #f1f5f9;
  --border: #e2e8f0;
  --border-light: #f1f5f9;
  --text-primary: #0f172a;
  --text-secondary: #475569;
  --text-muted: #94a3b8;
  --shadow-xs: 0 1px 2px rgba(0,0,0,0.03);
  --shadow-sm: 0 1px 3px rgba(0,0,0,0.04), 0 1px 2px rgba(0,0,0,0.02);
  --shadow-md: 0 4px 6px -1px rgba(0,0,0,0.05), 0 2px 4px -2px rgba(0,0,0,0.03);
  --shadow-lg: 0 10px 15px -3px rgba(0,0,0,0.06), 0 4px 6px -4px rgba(0,0,0,0.04);
  --radius: 12px;
  --radius-sm: 8px;
  --radius-xs: 6px;
  padding: 24px 28px;
  min-height: 100%;
  background: var(--bg-page);
}

/* ===== 页头 ===== */
.page-header-pro { margin-bottom: 20px; }
.header-content { display: flex; justify-content: space-between; align-items: center; }
.title-group { display: flex; align-items: center; gap: 14px; }
.title-badge { width: 42px; height: 42px; border-radius: var(--radius-sm); background: linear-gradient(135deg, var(--primary), #8b5cf6); display: flex; align-items: center; justify-content: center; color: #fff; box-shadow: 0 4px 12px rgba(99,102,241,0.25); }
.page-title-pro { font-size: 20px; font-weight: 700; color: var(--text-primary); margin: 0; letter-spacing: -0.3px; }
.page-subtitle { font-size: 13px; color: var(--text-muted); margin: 2px 0 0; }
.header-btn { width: 36px; height: 36px; border: 1px solid var(--border); border-radius: var(--radius-xs); background: var(--bg-card); cursor: pointer; display: flex; align-items: center; justify-content: center; color: var(--text-secondary); transition: all 0.2s; }
.header-btn:hover { border-color: var(--primary-border); color: var(--primary); background: var(--primary-bg); }
.spinning { animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

/* ===== 选择器面板 ===== */
.selector-panel-pro { background: var(--bg-card); border-radius: var(--radius); padding: 20px 24px; box-shadow: var(--shadow-sm); border: 1px solid var(--border-light); margin-bottom: 16px; }
.selector-grid { display: grid; grid-template-columns: 1.5fr auto 1fr auto; gap: 16px; align-items: end; }
.field-label { display: flex; align-items: center; gap: 5px; font-size: 11px; font-weight: 600; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.5px; margin-bottom: 6px; }

/* CRD 搜索框 */
.crd-search-select { position: relative; }
.crd-search-box { display: flex; align-items: center; border: 1px solid var(--border); border-radius: var(--radius-xs); background: var(--bg-card); padding: 0 12px; transition: all 0.2s; }
.crd-search-box.focused { border-color: var(--primary); box-shadow: 0 0 0 3px rgba(99,102,241,0.08); }
.search-prefix-icon { color: var(--text-muted); flex-shrink: 0; }
.crd-search-input { flex: 1; border: none; outline: none; padding: 10px 8px; font-size: 13px; background: transparent; color: var(--text-primary); min-width: 200px; }
.crd-search-input::placeholder { color: var(--text-muted); }
.search-clear { cursor: pointer; color: var(--text-muted); padding: 4px; border-radius: 4px; display: flex; }
.search-clear:hover { background: var(--bg-hover); color: var(--text-secondary); }

/* 下拉面板 */
.crd-dropdown-pro { position: absolute; top: calc(100% + 4px); left: 0; right: 0; max-height: 360px; background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-sm); box-shadow: var(--shadow-lg); z-index: 200; overflow: hidden; display: flex; flex-direction: column; }
.dropdown-header { padding: 8px 14px; border-bottom: 1px solid var(--border-light); }
.dropdown-count { font-size: 11px; color: var(--text-muted); font-weight: 500; }
.dropdown-list { overflow-y: auto; max-height: 320px; padding: 4px; }
.crd-option { display: flex; justify-content: space-between; align-items: center; padding: 10px 12px; border-radius: var(--radius-xs); cursor: pointer; transition: all 0.15s; margin: 1px 0; }
.crd-option:hover { background: var(--primary-bg); }
.crd-option.active { background: var(--primary-bg); border-left: 3px solid var(--primary); }
.crd-option-left { display: flex; align-items: center; gap: 10px; }
.crd-kind-badge { width: 28px; height: 28px; border-radius: 6px; background: linear-gradient(135deg, #e0e7ff, #c7d2fe); color: var(--primary); font-size: 12px; font-weight: 700; display: flex; align-items: center; justify-content: center; }
.crd-option-info { display: flex; flex-direction: column; }
.crd-option-kind { font-size: 13px; font-weight: 600; color: var(--text-primary); }
.crd-option-group { font-size: 11px; color: var(--text-muted); font-family: 'JetBrains Mono', monospace; }
.crd-scope-mini { font-size: 10px; font-weight: 600; padding: 2px 6px; border-radius: 4px; }
.crd-scope-mini.namespaced { background: #dbeafe; color: #2563eb; }
.crd-scope-mini.cluster { background: #f3e8ff; color: #7c3aed; }
.dropdown-empty { padding: 24px; text-align: center; display: flex; flex-direction: column; align-items: center; gap: 8px; color: var(--text-muted); font-size: 12px; }

/* 其他选择器 */
.select-pro { padding: 10px 14px; border: 1px solid var(--border); border-radius: var(--radius-xs); font-size: 13px; background: var(--bg-card); color: var(--text-primary); min-width: 150px; cursor: pointer; transition: all 0.2s; }
.select-pro:focus { outline: none; border-color: var(--primary); box-shadow: 0 0 0 3px rgba(99,102,241,0.08); }
.input-pro { width: 100%; padding: 10px 14px; border: 1px solid var(--border); border-radius: var(--radius-xs); font-size: 13px; background: var(--bg-card); color: var(--text-primary); transition: all 0.2s; box-sizing: border-box; }
.input-pro:focus { outline: none; border-color: var(--primary); box-shadow: 0 0 0 3px rgba(99,102,241,0.08); }
.input-pro::placeholder { color: var(--text-muted); }

/* 创建按钮 */
.btn-create-pro { display: inline-flex; align-items: center; gap: 6px; padding: 10px 18px; background: linear-gradient(135deg, var(--primary), #8b5cf6); color: #fff; border: none; border-radius: var(--radius-xs); font-size: 13px; font-weight: 600; cursor: pointer; transition: all 0.2s; box-shadow: 0 2px 8px rgba(99,102,241,0.25); white-space: nowrap; }
.btn-create-pro:hover:not(:disabled) { transform: translateY(-1px); box-shadow: 0 4px 12px rgba(99,102,241,0.35); }
.btn-create-pro:disabled { opacity: 0.5; cursor: not-allowed; transform: none; }
.btn-create-pro.sm { padding: 8px 16px; font-size: 12px; }

/* CRD 信息条 */
.crd-meta-bar { display: flex; justify-content: space-between; align-items: center; margin-top: 16px; padding-top: 16px; border-top: 1px solid var(--border-light); }
.meta-chips { display: flex; gap: 10px; flex-wrap: wrap; }
.meta-chip { display: inline-flex; align-items: center; gap: 6px; }
.chip-label { font-size: 10px; font-weight: 600; color: var(--text-muted); text-transform: uppercase; }
.chip-value { font-size: 12px; color: var(--text-primary); background: var(--bg-hover); padding: 3px 8px; border-radius: 4px; font-weight: 500; }
.chip-value.scope.namespaced { background: #dbeafe; color: #2563eb; }
.chip-value.scope.cluster { background: #f3e8ff; color: #7c3aed; }
.chip-value.mono { font-family: 'JetBrains Mono', monospace; font-size: 11px; }
.meta-stat { font-size: 12px; color: var(--text-muted); }
.stat-num { font-weight: 700; color: var(--primary); font-size: 14px; }

/* ===== 引导状态 ===== */
.guide-state { display: flex; justify-content: center; padding: 48px 0; }
.guide-card { text-align: center; max-width: 420px; }
.guide-illustration { margin-bottom: 20px; color: var(--text-muted); opacity: 0.4; }
.guide-title { font-size: 18px; font-weight: 700; color: var(--text-primary); margin: 0 0 8px; }
.guide-desc { font-size: 13px; color: var(--text-secondary); line-height: 1.7; margin: 0 0 24px; }
.guide-tips { display: flex; flex-direction: column; gap: 12px; text-align: left; }
.tip-item { display: flex; align-items: center; gap: 12px; font-size: 13px; color: var(--text-secondary); }
.tip-num { width: 24px; height: 24px; border-radius: 50%; background: var(--primary-bg); color: var(--primary); font-size: 11px; font-weight: 700; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }

/* ===== 加载状态 ===== */
.loading-state-pro { text-align: center; padding: 60px 20px; }
.loader-ring { width: 36px; height: 36px; border: 3px solid var(--border); border-top-color: var(--primary); border-radius: 50%; animation: spin 0.8s linear infinite; margin: 0 auto 16px; }
.loading-text { font-size: 13px; color: var(--text-secondary); }

/* ===== 批量操作栏 ===== */
.batch-bar { display: flex; align-items: center; padding: 0; margin-bottom: 12px; min-height: 40px; transition: all 0.3s; }
.batch-bar.visible .batch-left { display: flex; }
.batch-left { display: none; align-items: center; gap: 12px; padding: 8px 16px; background: var(--primary-bg); border: 1px solid var(--primary-border); border-radius: var(--radius-xs); }
.batch-count { font-size: 13px; color: var(--text-secondary); }
.batch-btn { display: inline-flex; align-items: center; gap: 5px; padding: 6px 12px; border-radius: var(--radius-xs); font-size: 12px; font-weight: 500; cursor: pointer; border: none; transition: all 0.15s; }
.batch-btn.danger { background: var(--danger); color: #fff; }
.batch-btn.danger:hover { background: #dc2626; }
.batch-btn.ghost { background: transparent; color: var(--text-secondary); border: 1px solid var(--border); }
.batch-btn.ghost:hover { background: var(--bg-hover); }

/* ===== 数据表格 ===== */
.table-container { background: var(--bg-card); border-radius: var(--radius); overflow: hidden; box-shadow: var(--shadow-sm); border: 1px solid var(--border-light); }
.data-table-pro { width: 100%; border-collapse: collapse; font-size: 13px; }
.data-table-pro th { padding: 12px 16px; text-align: left; font-size: 11px; font-weight: 600; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.5px; background: #f8fafc; border-bottom: 1px solid var(--border); }
.data-table-pro td { padding: 14px 16px; border-bottom: 1px solid var(--border-light); color: var(--text-primary); transition: background 0.15s; }
.data-table-pro tr:hover td { background: #fafbff; }
.data-table-pro tr.selected td { background: var(--primary-bg); }
.data-table-pro tr:last-child td { border-bottom: none; }

/* Checkbox */
.checkbox-pro { position: relative; display: flex; align-items: center; cursor: pointer; }
.checkbox-pro input { position: absolute; opacity: 0; width: 0; height: 0; }
.checkmark { width: 16px; height: 16px; border: 1.5px solid var(--border); border-radius: 4px; display: flex; align-items: center; justify-content: center; transition: all 0.2s; background: var(--bg-card); }
.checkbox-pro input:checked + .checkmark { background: var(--primary); border-color: var(--primary); }
.checkbox-pro input:checked + .checkmark::after { content: ''; width: 4px; height: 7px; border: solid #fff; border-width: 0 2px 2px 0; transform: rotate(45deg) translateY(-1px); }

/* 名称单元格 */
.name-cell-pro { display: flex; align-items: center; gap: 10px; }
.name-avatar { width: 30px; height: 30px; border-radius: 6px; background: linear-gradient(135deg, #e0e7ff, #c7d2fe); color: var(--primary); font-size: 12px; font-weight: 700; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.name-info { display: flex; flex-direction: column; }
.name-primary { font-weight: 500; color: var(--text-primary); }
.name-uid { font-size: 10px; color: var(--text-muted); font-family: monospace; }
.ns-tag { display: inline-block; padding: 3px 8px; background: #eff6ff; color: #2563eb; border-radius: 4px; font-size: 11px; font-weight: 500; }
.time-text { font-size: 12px; color: var(--text-secondary); }

/* 标签 */
.label-group { display: flex; flex-wrap: wrap; gap: 4px; }
.label-chip { display: inline-flex; font-size: 10px; border-radius: 4px; overflow: hidden; border: 1px solid var(--border-light); }
.label-key { padding: 2px 5px; background: #f1f5f9; color: var(--text-secondary); font-weight: 500; }
.label-val { padding: 2px 5px; color: var(--text-primary); }
.no-label { color: var(--text-muted); font-size: 12px; }
.label-overflow { font-size: 10px; color: var(--primary); font-weight: 600; padding: 2px 6px; background: var(--primary-bg); border-radius: 4px; }

/* 操作按钮 */
.actions-pro { display: flex; gap: 4px; }
.act-btn { display: inline-flex; align-items: center; gap: 4px; padding: 5px 10px; border: 1px solid transparent; border-radius: var(--radius-xs); font-size: 12px; cursor: pointer; transition: all 0.15s; background: transparent; color: var(--text-secondary); }
.act-btn:hover { background: var(--bg-hover); }
.act-btn.act-yaml { color: var(--primary); }
.act-btn.act-yaml:hover { background: var(--primary-bg); border-color: var(--primary-border); }
.act-btn.act-edit { color: #059669; }
.act-btn.act-edit:hover { background: rgba(5,150,105,0.06); border-color: rgba(5,150,105,0.2); }
.act-btn.act-del { color: var(--danger); }
.act-btn.act-del:hover { background: var(--danger-bg); border-color: rgba(239,68,68,0.2); }

/* ===== 空状态 ===== */
.empty-state-pro { text-align: center; padding: 60px 20px; }
.empty-visual { margin-bottom: 16px; color: var(--text-muted); opacity: 0.3; }
.empty-title { font-size: 16px; font-weight: 600; color: var(--text-primary); margin: 0 0 8px; }
.empty-desc { font-size: 13px; color: var(--text-secondary); margin: 0 0 20px; }

/* ===== 分页 ===== */
.pagination-pro { display: flex; justify-content: space-between; align-items: center; padding: 16px 0; }
.page-total { font-size: 13px; color: var(--text-secondary); }
.page-center { display: flex; align-items: center; gap: 4px; }
.page-arrow { width: 32px; height: 32px; border: 1px solid var(--border); border-radius: var(--radius-xs); background: var(--bg-card); display: flex; align-items: center; justify-content: center; cursor: pointer; color: var(--text-secondary); transition: all 0.15s; }
.page-arrow:hover:not(:disabled) { border-color: var(--primary-border); color: var(--primary); }
.page-arrow:disabled { opacity: 0.4; cursor: not-allowed; }
.page-num { min-width: 32px; height: 32px; border: 1px solid var(--border); border-radius: var(--radius-xs); background: var(--bg-card); font-size: 13px; color: var(--text-primary); cursor: pointer; display: flex; align-items: center; justify-content: center; transition: all 0.15s; }
.page-num:hover { border-color: var(--primary-border); color: var(--primary); }
.page-num.active { background: var(--primary); color: #fff; border-color: var(--primary); }
.page-ellipsis { width: 32px; height: 32px; display: flex; align-items: center; justify-content: center; color: var(--text-muted); }
.page-size-sel { padding: 6px 10px; border: 1px solid var(--border); border-radius: var(--radius-xs); font-size: 12px; background: var(--bg-card); color: var(--text-secondary); cursor: pointer; }

/* ===== 抽屉 ===== */
.drawer-overlay-pro { position: fixed; inset: 0; background: rgba(15,23,42,0.4); backdrop-filter: blur(2px); z-index: 2000; display: flex; justify-content: flex-end; }
.drawer-panel-pro { width: 56%; max-width: 740px; min-width: 400px; background: #1e1e2e; display: flex; flex-direction: column; box-shadow: -12px 0 40px rgba(0,0,0,0.3); }
.drawer-header-pro { display: flex; justify-content: space-between; align-items: center; padding: 16px 24px; border-bottom: 1px solid rgba(255,255,255,0.06); }
.drawer-title-area { display: flex; align-items: center; gap: 10px; color: #e2e8f0; }
.drawer-title-area h3 { margin: 0; font-size: 15px; font-weight: 600; }
.drawer-close-pro { width: 32px; height: 32px; border: none; background: rgba(255,255,255,0.06); border-radius: 6px; color: #94a3b8; cursor: pointer; display: flex; align-items: center; justify-content: center; transition: all 0.15s; }
.drawer-close-pro:hover { background: rgba(255,255,255,0.1); color: #fff; }
.drawer-body-pro { flex: 1; overflow: auto; }

/* ===== 编辑器 ===== */
.modal-overlay-pro { position: fixed; inset: 0; background: rgba(15,23,42,0.5); backdrop-filter: blur(4px); z-index: 2000; display: flex; align-items: center; justify-content: center; }
.editor-modal-pro { width: 82%; max-width: 1020px; height: 88vh; background: #1b1d2e; border-radius: 16px; display: flex; flex-direction: column; box-shadow: 0 25px 80px rgba(0,0,0,0.5); overflow: hidden; }
.editor-topbar { display: flex; justify-content: space-between; align-items: center; padding: 14px 20px; background: #252840; border-bottom: 1px solid rgba(255,255,255,0.06); }
.editor-topbar-left { display: flex; align-items: center; gap: 12px; }
.editor-mode-badge { padding: 4px 10px; border-radius: 4px; font-size: 11px; font-weight: 700; letter-spacing: 0.5px; color: #fff; }
.editor-mode-badge.create { background: linear-gradient(135deg, #10b981, #059669); }
.editor-mode-badge.edit { background: linear-gradient(135deg, #f59e0b, #d97706); }
.editor-title-pro { margin: 0; font-size: 15px; font-weight: 600; color: #e8eaf6; }
.editor-close-btn { width: 32px; height: 32px; border: none; background: rgba(255,255,255,0.06); border-radius: 6px; color: #8b95b0; cursor: pointer; display: flex; align-items: center; justify-content: center; }
.editor-close-btn:hover { background: rgba(255,255,255,0.12); color: #fff; }
.editor-info-bar { padding: 10px 20px; background: #1e2035; border-bottom: 1px solid rgba(255,255,255,0.04); display: flex; gap: 12px; }
.info-tag { display: inline-flex; align-items: center; gap: 6px; font-size: 12px; }
.info-tag-label { color: #6b7394; font-weight: 500; }
.info-tag-value { color: #a5b4fc; background: rgba(165,180,252,0.1); padding: 2px 8px; border-radius: 4px; font-family: monospace; font-size: 11px; }
.editor-toolbar { display: flex; justify-content: space-between; align-items: center; padding: 10px 20px; background: #22253a; border-bottom: 1px solid rgba(255,255,255,0.04); }
.editor-toolbar-left { display: flex; align-items: center; gap: 8px; }
.toolbar-btn { display: inline-flex; align-items: center; gap: 6px; padding: 6px 14px; border: 1px solid rgba(255,255,255,0.12); border-radius: 6px; background: transparent; color: #8b95b0; font-size: 12px; cursor: pointer; transition: all 0.15s; }
.toolbar-btn:hover:not(:disabled) { background: rgba(255,255,255,0.06); color: #e0e6f0; }
.toolbar-btn:disabled { opacity: 0.5; }
.toolbar-btn.secondary { border-color: rgba(255,255,255,0.08); }
.toolbar-btn.secondary:hover { background: rgba(255,255,255,0.06); color: #c8cce0; border-color: rgba(255,255,255,0.15); }
.toolbar-btn.danger { color: #fca5a5; border-color: rgba(239,68,68,0.3); }
.toolbar-btn.danger:hover { background: rgba(239,68,68,0.12); border-color: rgba(239,68,68,0.5); }
.toolbar-stat { font-size: 11px; color: #6b7394; }
.dryrun-banner { display: flex; align-items: center; justify-content: space-between; padding: 10px 20px; font-size: 13px; }
.dryrun-banner.success { background: rgba(16,185,129,0.12); color: #6ee7b7; }
.dryrun-banner.error { background: rgba(239,68,68,0.12); color: #fca5a5; }
.dryrun-banner button { background: none; border: none; color: inherit; opacity: 0.6; cursor: pointer; }
.editor-workspace { flex: 1; display: flex; overflow: hidden; }
.editor-gutter { width: 48px; min-width: 48px; padding: 16px 0; background: #181a2a; border-right: 1px solid rgba(255,255,255,0.04); overflow: hidden; text-align: right; user-select: none; }
.gutter-line { height: 22px; line-height: 22px; padding-right: 12px; font-family: 'JetBrains Mono', monospace; font-size: 12px; color: #4a5073; }
.yaml-editor-pro { flex: 1; width: 100%; padding: 16px 20px; border: none; resize: none; font-family: 'JetBrains Mono', monospace; font-size: 13px; line-height: 22px; background: #1e2030; color: #cdd6f4; outline: none; tab-size: 2; white-space: pre; overflow: auto; }
.yaml-editor-pro::placeholder { color: #4a5073; }
.editor-footer-pro { display: flex; justify-content: space-between; align-items: center; padding: 12px 20px; background: #252840; border-top: 1px solid rgba(255,255,255,0.06); }
.footer-mode-tag { padding: 3px 8px; background: rgba(99,102,241,0.15); color: #a5b4fc; border-radius: 4px; font-weight: 600; font-size: 10px; }
.editor-footer-right { display: flex; gap: 10px; }
.btn-cancel-pro { padding: 8px 18px; border: 1px solid rgba(255,255,255,0.15); border-radius: 8px; background: transparent; color: #8b95b0; font-size: 13px; cursor: pointer; }
.btn-cancel-pro:hover { background: rgba(255,255,255,0.06); color: #c8cce0; }
.btn-submit-pro { padding: 8px 22px; border: none; border-radius: 8px; background: linear-gradient(135deg, #6366f1, #8b5cf6); color: #fff; font-size: 13px; font-weight: 600; cursor: pointer; transition: all 0.2s; }
.btn-submit-pro:hover:not(:disabled) { box-shadow: 0 6px 16px rgba(99,102,241,0.35); }
.btn-submit-pro:disabled { opacity: 0.5; cursor: not-allowed; }

/* ===== 删除对话框 ===== */
.delete-dialog-pro { width: 420px; max-width: 90vw; background: var(--bg-card); border-radius: 16px; padding: 32px 28px 24px; text-align: center; box-shadow: 0 20px 60px rgba(0,0,0,0.15); }
.delete-icon-pro { width: 52px; height: 52px; margin: 0 auto 16px; border-radius: 50%; background: #fef2f2; color: var(--danger); display: flex; align-items: center; justify-content: center; }
.delete-title-pro { font-size: 17px; font-weight: 700; color: var(--text-primary); margin: 0 0 8px; }
.delete-desc-pro { font-size: 13px; color: var(--text-secondary); margin: 0 0 8px; }
.delete-desc-pro code { background: #f1f5f9; padding: 2px 6px; border-radius: 4px; font-size: 12px; }
.delete-warn-pro { font-size: 12px; color: var(--danger); margin: 0 0 20px; font-weight: 500; }
.delete-btns { display: flex; gap: 10px; justify-content: center; }
.dbtn { padding: 9px 22px; border-radius: 8px; font-size: 13px; font-weight: 500; cursor: pointer; transition: all 0.15s; }
.dbtn.cancel { border: 1px solid var(--border); background: var(--bg-card); color: var(--text-secondary); }
.dbtn.cancel:hover { background: var(--bg-hover); }
.dbtn.danger { border: none; background: var(--danger); color: #fff; font-weight: 600; }
.dbtn.danger:hover { background: #dc2626; }
.dbtn.danger:disabled { opacity: 0.6; }

/* ===== 动画 ===== */
.drawer-enter-active { transition: all 0.3s ease; }
.drawer-leave-active { transition: all 0.25s ease; }
.drawer-enter-from .drawer-panel-pro { transform: translateX(100%); }
.drawer-leave-to .drawer-panel-pro { transform: translateX(100%); }
.drawer-enter-from, .drawer-leave-to { opacity: 0; }
.modal-enter-active { transition: all 0.25s ease; }
.modal-leave-active { transition: all 0.2s ease; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
.modal-enter-from .editor-modal-pro, .modal-enter-from .delete-dialog-pro { transform: scale(0.95); opacity: 0; }
.modal-leave-to .editor-modal-pro, .modal-leave-to .delete-dialog-pro { transform: scale(0.95); opacity: 0; }
.slide-fade-enter-active { transition: all 0.3s ease; }
.slide-fade-leave-active { transition: all 0.2s ease; }
.slide-fade-enter-from, .slide-fade-leave-to { opacity: 0; transform: translateY(-8px); }
.slide-down-enter-active { transition: all 0.3s ease; }
.slide-down-leave-active { transition: all 0.2s ease; }
.slide-down-enter-from, .slide-down-leave-to { opacity: 0; transform: translateY(-8px); }
.dropdown-enter-active { transition: all 0.2s ease; }
.dropdown-leave-active { transition: all 0.15s ease; }
.dropdown-enter-from, .dropdown-leave-to { opacity: 0; transform: translateY(-4px); }

/* ===== 响应式 ===== */
@media (max-width: 1200px) {
  .selector-grid { grid-template-columns: 1fr 1fr; }
}
@media (max-width: 768px) {
  .selector-grid { grid-template-columns: 1fr; }
  .cr-instances-page { padding: 16px; }
}
</style>
