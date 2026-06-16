<template>
  <div class="appstore-container">
    <!-- ====== 顶部标题区 ====== -->
    <div class="appstore-header">
      <div class="header-left">
        <h2 class="page-title">
          <span class="title-icon-wrap">
            <svg class="title-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="3" width="7" height="7" rx="1.5"/>
              <rect x="14" y="3" width="7" height="7" rx="1.5"/>
              <rect x="3" y="14" width="7" height="7" rx="1.5"/>
              <rect x="14" y="14" width="7" height="7" rx="1.5"/>
            </svg>
          </span>
          应用商城
        </h2>
        <span class="app-count" v-if="total > 0">共 {{ total }} 个应用</span>
      </div>
      <div class="header-actions">
        <a-input-search
          v-model="searchKeyword"
          placeholder="搜索应用名称、描述、标签..."
          style="width: 280px"
          allow-clear
          @search="handleSearch"
          @clear="handleSearch"
          @press-enter="handleSearch"
        />
        <a-radio-group v-model="viewMode" type="button" size="small">
          <a-radio value="card"><icon-apps /> 卡片</a-radio>
          <a-radio value="table"><icon-list /> 列表</a-radio>
        </a-radio-group>
        <a-button class="history-btn" @click="router.push('/platform/appstore/records')">
          <template #icon><icon-history /></template>
          已部署应用
          <span v-if="installedCount > 0" class="badge-dot">{{ installedCount }}</span>
        </a-button>
        <a-button type="primary" @click="showCreateModal = true">
          <template #icon><icon-plus /></template>
          添加应用
        </a-button>
      </div>
    </div>

    <!-- ====== 分类导航栏 ====== -->
    <div class="category-bar">
      <div
        class="category-tag"
        :class="{ active: activeCategory === '' }"
        @click="switchCategory('')"
      >
        <span class="cat-label">全部</span>
        <span class="cat-count">{{ total }}</span>
      </div>
      <div
        v-for="cat in categories"
        :key="cat.category"
        class="category-tag"
        :class="{ active: activeCategory === cat.category }"
        @click="switchCategory(cat.category)"
      >
        <span class="cat-icon">{{ getCategoryIcon(cat.category) }}</span>
        <span class="cat-label">{{ cat.category }}</span>
        <span class="cat-count">{{ cat.count }}</span>
      </div>
    </div>

    <!-- ====== 加载状态 ====== -->
    <div v-if="loading" class="loading-wrap">
      <a-spin size="32" />
      <span class="loading-text">加载应用列表...</span>
    </div>

    <!-- ====== 空状态 ====== -->
    <div v-else-if="apps.length === 0" class="empty-wrap">
      <a-empty description="暂无应用">
        <template #image>
          <svg viewBox="0 0 80 80" width="80" height="80" fill="none" stroke="#c0c4cc" stroke-width="1.5">
            <rect x="8" y="8" width="28" height="28" rx="4"/>
            <rect x="44" y="8" width="28" height="28" rx="4"/>
            <rect x="8" y="44" width="28" height="28" rx="4"/>
            <rect x="44" y="44" width="28" height="28" rx="4" stroke-dasharray="4 2"/>
          </svg>
        </template>
        <a-button type="primary" size="small" @click="showCreateModal = true">添加第一个应用</a-button>
      </a-empty>
    </div>

    <!-- ====== 应用卡片网格 ====== -->
    <div v-else-if="viewMode === 'card'" class="app-grid">
      <div
        v-for="app in apps"
        :key="app.id"
        class="app-card"
        :class="{ featured: app.featured === 1, installed: isAppInstalled(app.id) }"
      >
        <!-- 推荐角标 -->
        <div v-if="app.featured === 1" class="featured-badge">
          <svg width="10" height="10" viewBox="0 0 10 10" fill="currentColor"><polygon points="5,0 6.18,3.82 10,3.82 6.91,6.18 8.09,10 5,7.64 1.91,10 3.09,6.18 0,3.82 3.82,3.82"/></svg>
          推荐
        </div>

        <!-- 已安装角标 -->
        <div v-if="isAppInstalled(app.id)" class="installed-badge">
          <svg width="12" height="12" viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="2"><polyline points="2.5,6.5 5,9 9.5,3.5"/></svg>
          已部署
        </div>

        <!-- 卡片头部 -->
        <div class="card-header">
          <div class="app-icon" :class="getIconClass(app.icon || app.name)">
            {{ getIconLetter(app.display_name || app.name) }}
          </div>
          <div class="app-meta">
            <div class="app-name">{{ app.display_name || app.name }}</div>
            <div class="app-version">
              <span class="version-tag">v{{ app.version }}</span>
              <span class="provider-tag" :class="'provider-' + app.provider">
                {{ providerLabel(app.provider) }}
              </span>
            </div>
          </div>
          <a-dropdown trigger="hover" position="br">
            <a-button type="text" size="small" class="more-btn">
              <icon-more />
            </a-button>
            <template #content>
              <a-doption @click="handleEdit(app)">
                <icon-edit /> 编辑
              </a-doption>
              <a-doption @click="handleDelete(app)" class="danger-opt">
                <icon-delete /> 删除
              </a-doption>
            </template>
          </a-dropdown>
        </div>

        <!-- 分类 + 标签 -->
        <div class="card-tags">
          <span class="category-badge">{{ app.category }}</span>
          <span v-if="app.min_k8s" class="k8s-badge">{{ app.min_k8s }}</span>
          <template v-if="app.tags">
            <span
              v-for="tag in parseTags(app.tags).slice(0, 3)"
              :key="tag"
              class="tag-item"
            >{{ tag }}</span>
          </template>
        </div>

        <!-- 描述 -->
        <div class="card-desc">{{ app.description || '暂无描述' }}</div>

        <!-- 底部操作 -->
        <div class="card-footer">
          <a-button size="small" @click="handleDetail(app)">
            <template #icon><icon-info-circle /></template>
            查看详情
          </a-button>
          <a-button
            :type="isAppInstalled(app.id) ? 'outline' : 'primary'"
            size="small"
            @click="handleInstall(app)"
          >
            <template #icon><icon-download /></template>
            {{ isAppInstalled(app.id) ? '重新安装' : '安装' }}
          </a-button>
        </div>
      </div>
    </div>

    <!-- ====== 表格管理视图 ====== -->
    <div v-else-if="viewMode === 'table' && apps.length > 0" class="table-view">
      <div class="table-toolbar" v-if="selectedRowKeys.length > 0">
        <span class="selected-info">已选择 {{ selectedRowKeys.length }} 项</span>
        <a-button size="small" status="danger" @click="handleBatchDeleteApps">
          <template #icon><icon-delete /></template>批量删除
        </a-button>
        <a-button size="small" @click="selectedRowKeys = []">取消选择</a-button>
      </div>
      <a-table
        :data="apps"
        :row-selection="{ type: 'checkbox', showCheckedAll: true }"
        v-model:selectedKeys="selectedRowKeys"
        :pagination="false"
        row-key="id"
        :bordered="false"
        stripe
        size="medium"
      >
        <template #columns>
          <a-table-column title="应用" :width="240">
            <template #cell="{ record }">
              <div style="display:flex;align-items:center;gap:10px;">
                <div class="app-icon" :class="getIconClass(record.icon || record.name)" style="width:32px;height:32px;font-size:14px;border-radius:8px;flex-shrink:0;">
                  {{ getIconLetter(record.display_name || record.name) }}
                </div>
                <div>
                  <div style="font-weight:600;color:#1d2129;font-size:13px;">{{ record.display_name || record.name }}</div>
                  <div style="font-size:11px;color:#86909c;">{{ record.name }}</div>
                </div>
              </div>
            </template>
          </a-table-column>
          <a-table-column title="分类" :width="90">
            <template #cell="{ record }">
              <span class="category-badge">{{ record.category }}</span>
            </template>
          </a-table-column>
          <a-table-column title="版本" :width="90">
            <template #cell="{ record }">
              <span class="version-tag">v{{ record.version }}</span>
            </template>
          </a-table-column>
          <a-table-column title="提供方" :width="80">
            <template #cell="{ record }">
              <a-tag size="small" :color="record.provider === 'official' ? 'blue' : 'green'">{{ providerLabel(record.provider) }}</a-tag>
            </template>
          </a-table-column>
          <a-table-column title="状态" :width="80">
            <template #cell="{ record }">
              <a-badge :status="record.status === 1 ? 'success' : record.status === 2 ? 'warning' : 'danger'" :text="statusLabel(record.status)" />
            </template>
          </a-table-column>
          <a-table-column title="推荐" :width="70" align="center">
            <template #cell="{ record }">
              <icon-star-fill v-if="record.featured === 1" style="color:#fadb14;font-size:16px;" />
              <span v-else style="color:#c0c4cc;">-</span>
            </template>
          </a-table-column>
          <a-table-column title="部署" :width="70" align="center">
            <template #cell="{ record }">
              <a-badge v-if="isAppInstalled(record.id)" status="success" text="已部署" />
              <span v-else style="color:#c0c4cc;font-size:12px;">未部署</span>
            </template>
          </a-table-column>
          <a-table-column title="排序" :width="70" align="center" data-index="sort_order" />
          <a-table-column title="操作" :width="220" align="center" fixed="right">
            <template #cell="{ record }">
              <a-space size="mini">
                <a-button type="text" size="mini" @click="handleDetail(record)">
                  <template #icon><icon-eye /></template>详情
                </a-button>
                <a-button type="text" size="mini" @click="handleInstall(record)">
                  <template #icon><icon-download /></template>安装
                </a-button>
                <a-button type="text" size="mini" @click="handleEdit(record)">
                  <template #icon><icon-edit /></template>编辑
                </a-button>
                <a-popconfirm content="确定删除该应用？" @ok="handleDeleteDirect(record.id)">
                  <a-button type="text" size="mini" status="danger">
                    <template #icon><icon-delete /></template>删除
                  </a-button>
                </a-popconfirm>
              </a-space>
            </template>
          </a-table-column>
        </template>
      </a-table>
    </div>

    <!-- ====== 分页 ====== -->
    <div v-if="total > pageSize" class="pagination-wrap">
      <a-pagination
        v-model:current="currentPage"
        :total="total"
        :page-size="pageSize"
        show-total
        show-jumper
        @change="handlePageChange"
      />
    </div>

    <!-- ====== 详情抽屉 (Rancher 风格) ====== -->
    <a-drawer
      :visible="showDetail"
      :width="580"
      @cancel="showDetail = false"
      unmount-on-close
      :footer="false"
      class="detail-drawer"
    >
      <template #title>
        <div class="drawer-title-custom" v-if="detailApp">
          <div class="app-icon" :class="getIconClass(detailApp.icon || detailApp.name)" style="width:32px;height:32px;font-size:15px;border-radius:8px;">
            {{ getIconLetter(detailApp.display_name || detailApp.name) }}
          </div>
          <div>
            <div style="font-weight:600;font-size:15px;color:#1d2129;">{{ detailApp.display_name || detailApp.name }}</div>
            <div style="font-size:12px;color:#86909c;">v{{ detailApp.version }} · {{ detailApp.category }}</div>
          </div>
        </div>
        <span v-else>应用详情</span>
      </template>
      <template v-if="detailLoading">
        <div class="detail-loading"><a-spin size="24" /><span>加载中...</span></div>
      </template>
      <template v-else-if="detailApp">
        <div class="detail-content">
          <!-- 快捷操作 -->
          <div class="detail-action-bar">
            <a-button type="primary" size="large" long @click="handleInstall(detailApp)">
              <template #icon><icon-download /></template>
              安装到集群
            </a-button>
            <a-button v-if="detailApp.doc_url" size="large" long @click="openDocUrl(detailApp.doc_url)">
              <template #icon><icon-book /></template>
              查看文档
            </a-button>
          </div>

          <!-- 基本信息 -->
          <div class="detail-section">
            <h4 class="section-title">基本信息</h4>
            <div class="info-grid">
              <div class="info-item">
                <span class="info-label">应用名称</span>
                <span class="info-value">{{ detailApp.name }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">版本</span>
                <span class="info-value"><a-tag size="small" color="arcoblue">v{{ detailApp.version }}</a-tag></span>
              </div>
              <div class="info-item">
                <span class="info-label">分类</span>
                <span class="info-value">{{ detailApp.category }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">提供方</span>
                <span class="info-value">
                  <a-tag size="small" :color="detailApp.provider === 'official' ? 'blue' : 'green'">{{ providerLabel(detailApp.provider) }}</a-tag>
                </span>
              </div>
              <div class="info-item">
                <span class="info-label">最低K8s</span>
                <span class="info-value">{{ detailApp.min_k8s || '-' }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">命名空间</span>
                <span class="info-value"><code>{{ detailApp.namespace || 'default' }}</code></span>
              </div>
            </div>
          </div>

          <!-- 描述 -->
          <div class="detail-section">
            <h4 class="section-title">描述</h4>
            <p class="detail-desc">{{ detailApp.description || '暂无描述' }}</p>
          </div>

          <!-- 标签 -->
          <div v-if="detailApp.tags" class="detail-section">
            <h4 class="section-title">标签</h4>
            <div class="detail-tags">
              <a-tag v-for="t in parseTags(detailApp.tags)" :key="t" color="arcoblue">{{ t }}</a-tag>
            </div>
          </div>

          <!-- 链接 -->
          <div class="detail-section">
            <h4 class="section-title">链接</h4>
            <div class="link-cards">
              <a v-if="detailApp.chart_url" :href="detailApp.chart_url" target="_blank" class="link-card">
                <icon-code /> Chart 仓库
                <icon-launch class="link-arrow" />
              </a>
              <a v-if="detailApp.doc_url" :href="detailApp.doc_url" target="_blank" class="link-card">
                <icon-book /> 官方文档
                <icon-launch class="link-arrow" />
              </a>
            </div>
          </div>

          <!-- 组件列表（动态配置） -->
          <div class="detail-section">
            <h4 class="section-title">
              安装组件
              <span class="comp-count" v-if="detailComponents.length">{{ detailComponents.length }} 个</span>
              <div style="margin-left:auto;display:flex;gap:4px;">
                <a-popconfirm v-if="selectedCompIds.length > 0" :content="`确定删除选中的 ${selectedCompIds.length} 个组件？`" @ok="handleBatchDeleteComps">
                  <a-button size="mini" type="text" status="danger">
                    <template #icon><icon-delete /></template>批量删除({{ selectedCompIds.length }})
                  </a-button>
                </a-popconfirm>
                <a-button size="mini" type="text" @click="showAddCompModal = true">
                  <template #icon><icon-plus /></template>添加
                </a-button>
              </div>
            </h4>
            <div v-if="compLoading" style="text-align:center;padding:16px;">
              <a-spin size="16" />
            </div>
            <div v-else-if="detailComponents.length === 0" class="comp-empty">
              暂无组件配置，安装时将使用内置默认组件
            </div>
            <div v-else class="comp-list">
              <div v-for="(comp, idx) in detailComponents" :key="comp.id" class="comp-item">
                <div class="comp-header">
                  <a-checkbox :model-value="selectedCompIds.includes(comp.id)" @change="(v) => { if(v) selectedCompIds.push(comp.id); else selectedCompIds = selectedCompIds.filter(id => id !== comp.id) }" />
                  <span class="comp-name">{{ comp.name }}</span>
                  <span class="comp-replicas">x{{ comp.replicas }}</span>
                  <span class="comp-sort-badge" v-if="comp.sort_order">{{ comp.sort_order }}</span>
                  <div class="comp-actions">
                    <a-tooltip content="上移">
                      <a-button size="mini" type="text" :disabled="idx === 0" @click="moveComp(idx, -1)">
                        <template #icon><icon-arrow-up /></template>
                      </a-button>
                    </a-tooltip>
                    <a-tooltip content="下移">
                      <a-button size="mini" type="text" :disabled="idx === detailComponents.length - 1" @click="moveComp(idx, 1)">
                        <template #icon><icon-arrow-down /></template>
                      </a-button>
                    </a-tooltip>
                    <a-button size="mini" type="text" @click="editComp(comp)">
                      <template #icon><icon-edit /></template>
                    </a-button>
                    <a-popconfirm content="确定删除该组件？" @ok="deleteComp(comp.id)">
                      <a-button size="mini" type="text" status="danger">
                        <template #icon><icon-delete /></template>
                      </a-button>
                    </a-popconfirm>
                  </div>
                </div>
                <div class="comp-image">{{ comp.image }}</div>
                <div class="comp-resources">
                  <span>CPU: {{ comp.cpu_req }}/{{ comp.cpu_lim }}</span>
                  <span>Mem: {{ comp.mem_req }}/{{ comp.mem_lim }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- 该应用的安装记录快捷入口 -->
          <div class="detail-section" v-if="detailApp">
            <h4 class="section-title">部署记录</h4>
            <a-button long @click="router.push(`/platform/appstore/records?app_id=${detailApp.id}`); showDetail = false">
              <template #icon><icon-history /></template>
              查看该应用的部署记录
            </a-button>
          </div>

          <!-- Values YAML -->
          <div v-if="detailApp.values_yaml" class="detail-section">
            <h4 class="section-title">默认 Values</h4>
            <pre class="yaml-block">{{ detailApp.values_yaml }}</pre>
          </div>
        </div>
      </template>
    </a-drawer>

    <!-- ====== 创建/编辑弹窗 ====== -->
    <a-modal
      v-model:visible="showCreateModal"
      :title="editingApp ? '编辑应用' : '添加应用'"
      :width="640"
      @ok="handleCreateSubmit"
      @cancel="resetForm"
      unmount-on-close
    >
      <a-form :model="formData" layout="vertical">
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="应用名称" required>
              <a-input v-model="formData.name" placeholder="如: ingress-nginx" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="显示名称">
              <a-input v-model="formData.display_name" placeholder="如: Ingress NGINX" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="16">
          <a-col :span="8">
            <a-form-item label="分类" required>
              <a-select v-model="formData.category" placeholder="选择分类" allow-create>
                <a-option v-for="c in categoryOptions" :key="c" :value="c">{{ c }}</a-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item label="版本" required>
              <a-input v-model="formData.version" placeholder="1.0.0" />
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item label="提供方">
              <a-select v-model="formData.provider" placeholder="选择">
                <a-option value="official">官方</a-option>
                <a-option value="community">社区</a-option>
                <a-option value="third-party">第三方</a-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="描述">
          <a-textarea v-model="formData.description" placeholder="应用功能描述" :max-length="500" show-word-limit :auto-size="{ minRows: 2, maxRows: 4 }" />
        </a-form-item>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="Chart 地址">
              <a-input v-model="formData.chart_url" placeholder="Helm Chart URL" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="文档地址">
              <a-input v-model="formData.doc_url" placeholder="Documentation URL" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="16">
          <a-col :span="8">
            <a-form-item label="最低K8s版本">
              <a-input v-model="formData.min_k8s" placeholder="v1.22+" />
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item label="默认命名空间">
              <a-input v-model="formData.namespace" placeholder="default" />
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item label="标签">
              <a-input v-model="formData.tags" placeholder="逗号分隔" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="16">
          <a-col :span="8">
            <a-form-item label="状态">
              <a-select v-model="formData.status">
                <a-option :value="1">可用</a-option>
                <a-option :value="2">维护中</a-option>
                <a-option :value="3">已下架</a-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item label="推荐">
              <a-switch v-model="formFeatured" />
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item label="排序权重">
              <a-input-number v-model="formData.sort_order" :min="0" :max="999" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="默认 Values YAML">
          <a-textarea v-model="formData.values_yaml" placeholder="values.yaml 内容" :auto-size="{ minRows: 3, maxRows: 8 }" />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- ====== 组件添加/编辑弹窗 ====== -->
    <a-modal
      v-model:visible="showAddCompModal"
      :title="editingComp ? '编辑组件' : '添加组件'"
      :width="520"
      @ok="handleCompSubmit"
      @cancel="resetCompForm"
      unmount-on-close
    >
      <a-form :model="compForm" layout="vertical">
        <a-row :gutter="12">
          <a-col :span="8">
            <a-form-item label="组件名称" required>
              <a-input v-model="compForm.name" placeholder="如: prometheus" />
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item label="副本数">
              <a-input-number v-model="compForm.replicas" :min="1" :max="10" />
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item label="排序权重">
              <a-input-number v-model="compForm.sort_order" :min="0" :max="999" placeholder="越大越靠前" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="容器镜像" required>
          <a-input v-model="compForm.image" placeholder="如: prom/prometheus:v2.51.2" />
        </a-form-item>
        <a-form-item label="端口 (JSON)">
          <a-input v-model="compForm.ports" placeholder='[{"name":"http","port":9090}]' />
        </a-form-item>
        <a-form-item label="启动参数 (JSON)">
          <a-input v-model="compForm.args" placeholder='["--config.file=..."]' />
        </a-form-item>
        <a-row :gutter="12">
          <a-col :span="6"><a-form-item label="CPU Req"><a-input v-model="compForm.cpu_req" placeholder="50m" /></a-form-item></a-col>
          <a-col :span="6"><a-form-item label="CPU Lim"><a-input v-model="compForm.cpu_lim" placeholder="200m" /></a-form-item></a-col>
          <a-col :span="6"><a-form-item label="Mem Req"><a-input v-model="compForm.mem_req" placeholder="64Mi" /></a-form-item></a-col>
          <a-col :span="6"><a-form-item label="Mem Lim"><a-input v-model="compForm.mem_lim" placeholder="256Mi" /></a-form-item></a-col>
        </a-row>
      </a-form>
    </a-modal>

    <!-- ====== 安装弹窗 (Rancher 风格步骤式) ====== -->
    <a-modal
      v-model:visible="showInstallModal"
      :title="null"
      :width="600"
      :footer="installPhase === 'form' ? undefined : false"
      :ok-text="'开始安装'"
      :ok-loading="installLoading"
      :mask-closable="installPhase === 'form'"
      :closable="installPhase === 'form' || installPhase === 'done'"
      @ok="handleInstallSubmit"
      @cancel="handleInstallCancel"
      unmount-on-close
      class="install-modal"
    >
      <!-- 阶段一：配置表单 -->
      <div v-if="installPhase === 'form'">
        <div class="install-modal-header">
          <div v-if="installTargetApp" class="install-app-banner">
            <div class="app-icon" :class="getIconClass(installTargetApp.icon || installTargetApp.name)" style="width:48px;height:48px;font-size:22px;border-radius:12px;">
              {{ getIconLetter(installTargetApp.display_name || installTargetApp.name) }}
            </div>
            <div class="install-app-info">
              <div class="install-app-name">{{ installTargetApp.display_name || installTargetApp.name }}</div>
              <div class="install-app-meta">
                <a-tag size="small" color="arcoblue">v{{ installTargetApp.version }}</a-tag>
                <a-tag size="small" :color="installTargetApp.provider === 'official' ? 'blue' : 'green'">{{ providerLabel(installTargetApp.provider) }}</a-tag>
                <span class="install-app-cat">{{ installTargetApp.category }}</span>
              </div>
              <div class="install-app-desc">{{ installTargetApp.description }}</div>
            </div>
          </div>
        </div>
        <a-divider style="margin:16px 0" />
        <a-form :model="installForm" layout="vertical">
          <a-form-item label="目标集群" required>
            <a-select v-model="installForm.cluster_id" placeholder="选择要安装到的集群" allow-search>
              <a-option v-for="c in clusterList" :key="c.id" :value="c.id">
                <div style="display:flex;align-items:center;gap:8px;">
                  <span style="width:6px;height:6px;border-radius:50%;background:#00b42a;display:inline-block;"></span>
                  {{ c.cluster_name }}
                  <span style="color:#86909c;font-size:12px;">({{ c.cluster_version || '-' }})</span>
                </div>
              </a-option>
            </a-select>
          </a-form-item>
          <a-row :gutter="12">
            <a-col :span="12">
              <a-form-item label="Release 名称" required>
                <a-input v-model="installForm.release_name" :placeholder="installTargetApp?.name || 'my-release'" />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="命名空间" required>
                <a-input v-model="installForm.namespace" :placeholder="installTargetApp?.namespace || 'default'" />
              </a-form-item>
            </a-col>
          </a-row>
          <a-form-item label="自定义 Values (YAML)">
            <a-textarea v-model="installForm.values" :placeholder="'# 自定义覆盖配置'" :auto-size="{ minRows: 3, maxRows: 8 }" class="values-editor" />
          </a-form-item>
        </a-form>
      </div>

      <!-- 阶段二：安装进度 -->
      <div v-else-if="installPhase === 'progress'" class="install-progress-panel">
        <div class="progress-header">
          <div class="progress-spinner">
            <svg class="spinner-ring" viewBox="0 0 50 50"><circle cx="25" cy="25" r="20" fill="none" stroke-width="3"/></svg>
          </div>
          <div class="progress-title">正在安装 {{ installTargetApp?.display_name || '' }}</div>
          <div class="progress-subtitle">请稍候，正在将应用部署到目标集群...</div>
        </div>
        <div class="install-steps">
          <div v-for="(step, idx) in installSteps" :key="idx" class="step-item" :class="step.status">
            <div class="step-indicator">
              <div class="step-dot">
                <svg v-if="step.status === 'done'" width="14" height="14" viewBox="0 0 14 14" fill="none"><polyline points="3,7 6,10 11,4" stroke="#fff" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
                <svg v-else-if="step.status === 'error'" width="14" height="14" viewBox="0 0 14 14" fill="none"><line x1="4" y1="4" x2="10" y2="10" stroke="#fff" stroke-width="2" stroke-linecap="round"/><line x1="10" y1="4" x2="4" y2="10" stroke="#fff" stroke-width="2" stroke-linecap="round"/></svg>
                <span v-else-if="step.status === 'active'" class="dot-pulse"></span>
                <span v-else class="dot-num">{{ idx + 1 }}</span>
              </div>
              <div v-if="idx < installSteps.length - 1" class="step-line" :class="{ filled: step.status === 'done' }"></div>
            </div>
            <div class="step-content">
              <div class="step-label">{{ step.label }}</div>
              <div class="step-desc">{{ step.desc }}</div>
            </div>
          </div>
        </div>
      </div>

      <!-- 阶段三：安装结果 -->
      <div v-else-if="installPhase === 'done'" class="install-result-panel">
        <div class="result-icon" :class="installResult.success ? 'success' : 'error'">
          <svg v-if="installResult.success" width="48" height="48" viewBox="0 0 48 48" fill="none">
            <circle cx="24" cy="24" r="22" stroke="currentColor" stroke-width="2.5"/>
            <polyline points="14,24 21,32 34,17" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
          <svg v-else width="48" height="48" viewBox="0 0 48 48" fill="none">
            <circle cx="24" cy="24" r="22" stroke="currentColor" stroke-width="2.5"/>
            <line x1="16" y1="16" x2="32" y2="32" stroke="currentColor" stroke-width="3" stroke-linecap="round"/>
            <line x1="32" y1="16" x2="16" y2="32" stroke="currentColor" stroke-width="3" stroke-linecap="round"/>
          </svg>
        </div>
        <div class="result-title">{{ installResult.success ? '安装成功' : '安装失败' }}</div>
        <div class="result-message">{{ installResult.message }}</div>
        <div v-if="installResult.success" class="result-summary">
          <div class="summary-row"><span class="summary-label">应用</span><span class="summary-value">{{ installTargetApp?.display_name }}</span></div>
          <div class="summary-row"><span class="summary-label">集群</span><span class="summary-value">{{ installResult.clusterName || '-' }}</span></div>
          <div class="summary-row"><span class="summary-label">命名空间</span><span class="summary-value"><code>{{ installForm.namespace }}</code></span></div>
          <div class="summary-row"><span class="summary-label">Release</span><span class="summary-value"><code>{{ installForm.release_name }}</code></span></div>
        </div>
        <div class="result-actions">
          <a-button @click="handleInstallCancel">关闭</a-button>
          <a-button v-if="!installResult.success" type="primary" @click="installPhase = 'form'">重试</a-button>
          <a-button v-if="installResult.success && installResult.installId" type="primary" @click="goToInstallDetail(installResult.installId)">查看部署详情</a-button>
          <a-button v-else-if="installResult.success" type="primary" @click="handleInstallCancel(); router.push('/platform/appstore/records')">查看已部署应用</a-button>
        </div>
      </div>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { Message, Modal } from '@arco-design/web-vue'
import {
  IconPlus, IconMore, IconEdit, IconDelete,
  IconInfoCircle, IconDownload, IconHistory,
  IconBook, IconCode, IconLaunch, IconEye,
  IconApps, IconList, IconStarFill, IconArrowUp, IconArrowDown
} from '@arco-design/web-vue/es/icon'
import {
  getAppStoreList,
  getAppStoreDetail,
  getAppStoreCategories,
  createAppStoreApp,
  updateAppStoreApp,
  deleteAppStoreApp,
  installApp as installAppApi,
  getInstallList,
  getInstallDetail,
  getComponentList,
  createComponent,
  updateComponent,
  deleteComponent,
  batchDeleteComponents,
  sortComponents,
} from '@/api/platform/appstore.js'
import { getK8sClusterList } from '@/api/platform/cluster.js'

const router = useRouter()

// ====== 状态 ======
const loading = ref(false)
const apps = ref([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = 12
const searchKeyword = ref('')
const activeCategory = ref('')
const categories = ref([])

// 管理模式（表格视图）
const viewMode = ref('card') // card | table
const selectedRowKeys = ref([])

// 详情抽屉
const showDetail = ref(false)
const detailLoading = ref(false)
const detailApp = ref(null)

// 创建/编辑
const showCreateModal = ref(false)
const editingApp = ref(null)
const formData = reactive({
  name: '', display_name: '', category: '', version: '',
  icon: '', description: '', provider: 'official',
  chart_url: '', doc_url: '', status: 1, featured: 0,
  sort_order: 0, tags: '', min_k8s: '', namespace: '', values_yaml: ''
})

const formFeatured = computed({
  get: () => formData.featured === 1,
  set: (v) => { formData.featured = v ? 1 : 0 }
})

// 安装弹窗（步骤式）
const showInstallModal = ref(false)
const installTargetApp = ref(null)
const installLoading = ref(false)
const installPhase = ref('form') // form | progress | done
const installForm = reactive({
  cluster_id: null, release_name: '', namespace: '', values: ''
})
const clusterList = ref([])
const installSteps = ref([
  { label: '校验应用信息', desc: '检查应用配置与依赖', status: 'pending' },
  { label: '连接目标集群', desc: '建立集群安全连接', status: 'pending' },
  { label: '创建命名空间', desc: '准备部署环境', status: 'pending' },
  { label: '部署应用组件', desc: '安装到目标集群', status: 'pending' },
])
const installResult = reactive({ success: false, message: '', clusterName: '', installId: null })
let installPollTimer = null

// 安装记录

// 已安装应用 ID 集合
const installedAppIds = ref(new Set())
const installedCount = ref(0)

// 分类选项
const categoryOptions = computed(() => {
  const preset = ['网络', '监控', '日志', '存储', '安全', 'GitOps', '数据库', '消息队列']
  const dynamic = categories.value.map(c => c.category)
  return [...new Set([...preset, ...dynamic])]
})

function isAppInstalled(appId) {
  return installedAppIds.value.has(appId)
}

// ====== 数据加载 ======
async function fetchApps() {
  loading.value = true
  try {
    const res = await getAppStoreList({
      category: activeCategory.value,
      keyword: searchKeyword.value,
      page: currentPage.value,
      page_size: pageSize,
      status: 1
    })
    apps.value = res?.data?.list || []
    total.value = res?.data?.total || 0
  } catch (e) {
    console.error('获取应用列表失败:', e)
    apps.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

async function fetchCategories() {
  try {
    const res = await getAppStoreCategories()
    categories.value = res?.data || []
  } catch (e) {
    console.error('获取分类失败:', e)
  }
}

async function fetchInstalledApps() {
  try {
    const res = await getInstallList({ page: 1, page_size: 200, status: 2 })
    const list = res?.data?.list || []
    const ids = new Set()
    list.forEach(r => { if (r.status === 2) ids.add(r.app_id) })
    installedAppIds.value = ids
    installedCount.value = ids.size
  } catch { /* ignore */ }
}

// ====== 事件处理 ======
function handleSearch() {
  currentPage.value = 1
  fetchApps()
}

function switchCategory(cat) {
  activeCategory.value = cat
  currentPage.value = 1
  fetchApps()
}

function handlePageChange(page) {
  currentPage.value = page
  fetchApps()
}

async function handleDetail(app) {
  showDetail.value = true
  detailLoading.value = true
  detailComponents.value = []
  try {
    const res = await getAppStoreDetail(app.id)
    detailApp.value = res?.data || app
    fetchComponents(app.id)
  } catch {
    detailApp.value = app
  } finally {
    detailLoading.value = false
  }
}

function handleEdit(app) {
  editingApp.value = app
  Object.assign(formData, {
    name: app.name, display_name: app.display_name,
    category: app.category, version: app.version,
    icon: app.icon, description: app.description,
    provider: app.provider, chart_url: app.chart_url,
    doc_url: app.doc_url, status: app.status,
    featured: app.featured, sort_order: app.sort_order,
    tags: app.tags, min_k8s: app.min_k8s,
    namespace: app.namespace, values_yaml: app.values_yaml
  })
  showCreateModal.value = true
}

function handleDelete(app) {
  Modal.warning({
    title: '确认删除',
    content: `确定要删除应用「${app.display_name || app.name}」吗？此操作不可恢复。`,
    okText: '删除',
    cancelText: '取消',
    hideCancel: false,
    onOk: async () => {
      try {
        await deleteAppStoreApp(app.id)
        Message.success('删除成功')
        fetchApps()
        fetchCategories()
      } catch (e) {
        Message.error('删除失败: ' + (e?.msg || e?.message || ''))
      }
    }
  })
}

async function handleDeleteDirect(id) {
  try {
    await deleteAppStoreApp(id)
    Message.success('删除成功')
    fetchApps()
    fetchCategories()
  } catch (e) {
    Message.error('删除失败: ' + (e?.msg || e?.message || ''))
  }
}

async function handleCreateSubmit() {
  if (!formData.name || !formData.category || !formData.version) {
    Message.warning('请填写必填项：应用名称、分类、版本')
    return
  }
  try {
    if (editingApp.value) {
      await updateAppStoreApp({ id: editingApp.value.id, ...formData })
      Message.success('更新成功')
    } else {
      await createAppStoreApp({ ...formData })
      Message.success('创建成功')
    }
    showCreateModal.value = false
    resetForm()
    fetchApps()
    fetchCategories()
  } catch (e) {
    Message.error((editingApp.value ? '更新' : '创建') + '失败: ' + (e?.msg || e?.message || ''))
  }
}

function resetForm() {
  editingApp.value = null
  Object.assign(formData, {
    name: '', display_name: '', category: '', version: '',
    icon: '', description: '', provider: 'official',
    chart_url: '', doc_url: '', status: 1, featured: 0,
    sort_order: 0, tags: '', min_k8s: '', namespace: '', values_yaml: ''
  })
}

// ====== 组件管理 ======
const detailComponents = ref([])
const compLoading = ref(false)
const showAddCompModal = ref(false)
const editingComp = ref(null)
const compForm = reactive({
  name: '', image: '', replicas: 1, ports: '', args: '',
  cpu_req: '50m', cpu_lim: '200m', mem_req: '64Mi', mem_lim: '256Mi',
  sort_order: 100
})
const selectedCompIds = ref([])

async function fetchComponents(appId) {
  compLoading.value = true
  try {
    const res = await getComponentList(appId)
    detailComponents.value = res?.data || []
  } catch {
    detailComponents.value = []
  } finally {
    compLoading.value = false
  }
}

function editComp(comp) {
  editingComp.value = comp
  Object.assign(compForm, {
    name: comp.name, image: comp.image, replicas: comp.replicas,
    ports: comp.ports, args: comp.args,
    cpu_req: comp.cpu_req, cpu_lim: comp.cpu_lim,
    mem_req: comp.mem_req, mem_lim: comp.mem_lim,
    sort_order: comp.sort_order || 0
  })
  showAddCompModal.value = true
}

function resetCompForm() {
  editingComp.value = null
  Object.assign(compForm, {
    name: '', image: '', replicas: 1, ports: '', args: '',
    cpu_req: '50m', cpu_lim: '200m', mem_req: '64Mi', mem_lim: '256Mi',
    sort_order: 100
  })
}

async function handleCompSubmit() {
  if (!compForm.name || !compForm.image) {
    Message.warning('请填写组件名称和镜像')
    return
  }
  try {
    if (editingComp.value) {
      await updateComponent({ id: editingComp.value.id, app_id: detailApp.value.id, ...compForm, sort_order: compForm.sort_order })
      Message.success('组件更新成功')
    } else {
      await createComponent({ app_id: detailApp.value.id, ...compForm, sort_order: compForm.sort_order })
      Message.success('组件添加成功')
    }
    showAddCompModal.value = false
    resetCompForm()
    fetchComponents(detailApp.value.id)
  } catch (e) {
    Message.error('操作失败: ' + (e?.msg || e?.message || ''))
  }
}

async function deleteComp(compId) {
  try {
    await deleteComponent(compId)
    Message.success('组件已删除')
    fetchComponents(detailApp.value.id)
  } catch (e) {
    Message.error('删除失败: ' + (e?.msg || e?.message || ''))
  }
}

async function handleBatchDeleteComps() {
  if (selectedCompIds.value.length === 0) {
    Message.warning('请选择要删除的组件')
    return
  }
  try {
    await batchDeleteComponents(selectedCompIds.value)
    Message.success(`已删除 ${selectedCompIds.value.length} 个组件`)
    selectedCompIds.value = []
    fetchComponents(detailApp.value.id)
  } catch (e) {
    Message.error('批量删除失败: ' + (e?.msg || e?.message || ''))
  }
}

function moveComp(index, direction) {
  const list = [...detailComponents.value]
  const targetIdx = index + direction
  if (targetIdx < 0 || targetIdx >= list.length) return
  ;[list[index], list[targetIdx]] = [list[targetIdx], list[index]]
  // 重新计算排序值（倒序：最靠前的值最大）
  const items = list.map((comp, i) => ({
    id: comp.id,
    sort_order: (list.length - i) * 10
  }))
  sortComponents(items).then(() => {
    fetchComponents(detailApp.value.id)
  }).catch(e => {
    Message.error('排序失败: ' + (e?.msg || e?.message || ''))
  })
}

// 批量删除应用
async function handleBatchDeleteApps() {
  if (selectedRowKeys.value.length === 0) {
    Message.warning('请选择要删除的应用')
    return
  }
  Modal.warning({
    title: '批量删除',
    content: `确定要删除选中的 ${selectedRowKeys.value.length} 个应用吗？此操作不可恢复。`,
    okText: '删除',
    cancelText: '取消',
    hideCancel: false,
    onOk: async () => {
      try {
        for (const id of selectedRowKeys.value) {
          await deleteAppStoreApp(id)
        }
        Message.success(`已删除 ${selectedRowKeys.value.length} 个应用`)
        selectedRowKeys.value = []
        fetchApps()
        fetchCategories()
      } catch (e) {
        Message.error('批量删除失败: ' + (e?.msg || e?.message || ''))
      }
    }
  })
}

// ====== 安装流程（步骤式） ======
function handleInstall(app) {
  installTargetApp.value = app
  installForm.cluster_id = null
  installForm.release_name = app.name || ''
  installForm.namespace = app.namespace || 'default'
  installForm.values = app.values_yaml || ''
  installPhase.value = 'form'
  resetInstallSteps()
  showInstallModal.value = true
  fetchClusters()
}

function handleInstallCancel() {
  showInstallModal.value = false
  if (installPollTimer) { clearInterval(installPollTimer); installPollTimer = null }
  fetchInstalledApps()
}

function resetInstallSteps() {
  installSteps.value = [
    { label: '校验应用信息', desc: '检查应用配置与依赖', status: 'pending' },
    { label: '连接目标集群', desc: '建立集群安全连接', status: 'pending' },
    { label: '创建命名空间', desc: '准备部署环境', status: 'pending' },
    { label: '部署应用组件', desc: '安装到目标集群', status: 'pending' },
  ]
}

async function fetchClusters() {
  try {
    const res = await getK8sClusterList({ page: 1, limit: 100 })
    clusterList.value = res?.data?.list || []
  } catch (e) {
    console.error('获取集群列表失败:', e)
    clusterList.value = []
  }
}

async function handleInstallSubmit() {
  if (!installForm.cluster_id) { Message.warning('请选择目标集群'); return }
  if (!installForm.release_name || !installForm.namespace) { Message.warning('请填写 Release 名称和命名空间'); return }

  installLoading.value = true
  installPhase.value = 'progress'

  // 找到集群名称
  const cluster = clusterList.value.find(c => c.id === installForm.cluster_id)
  installResult.clusterName = cluster?.cluster_name || '-'

  // 模拟步骤推进 + 真实API
  await animateStep(0, 400)  // 校验应用
  await animateStep(1, 600)  // 连接集群

  try {
    const payload = {
      app_id: installTargetApp.value.id,
      cluster_id: installForm.cluster_id,
      namespace: installForm.namespace,
      release_name: installForm.release_name,
      values: installForm.values
    }
    await animateStep(2, 500) // 创建命名空间
    setStepActive(3)

    const res = await installAppApi(payload)
    const installData = res?.data

    if (installData?.id) {
      installResult.installId = installData.id
      // 轮询等待安装结果
      await waitForInstallResult(installData.id)
    } else {
      // 没有返回ID，认为直接完成
      setStepDone(3)
      installResult.success = true
      installResult.message = '应用安装任务已提交'
      installPhase.value = 'done'
    }
  } catch (e) {
    setStepError(getCurrentActiveStep())
    installResult.success = false
    installResult.message = e?.msg || e?.message || '安装请求失败，请检查集群连接'
    installPhase.value = 'done'
  } finally {
    installLoading.value = false
  }
}

function getCurrentActiveStep() {
  return installSteps.value.findIndex(s => s.status === 'active')
}

function setStepActive(idx) {
  if (idx >= 0 && idx < installSteps.value.length) {
    installSteps.value[idx].status = 'active'
  }
}
function setStepDone(idx) {
  if (idx >= 0 && idx < installSteps.value.length) {
    installSteps.value[idx].status = 'done'
  }
}
function setStepError(idx) {
  if (idx >= 0 && idx < installSteps.value.length) {
    installSteps.value[idx].status = 'error'
  }
}

function animateStep(idx, delay) {
  return new Promise(resolve => {
    setStepActive(idx)
    setTimeout(() => {
      setStepDone(idx)
      resolve()
    }, delay)
  })
}

async function waitForInstallResult(installId) {
  return new Promise((resolve) => {
    let attempts = 0
    const maxAttempts = 20
    installPollTimer = setInterval(async () => {
      attempts++
      try {
        const res = await getInstallDetail(installId)
        const record = res?.data
        if (record) {
          if (record.status === 2) {
            clearInterval(installPollTimer); installPollTimer = null
            setStepDone(3)
            installResult.success = true
            installResult.message = record.message || '应用安装成功，已部署到目标集群'
            installPhase.value = 'done'
            resolve()
            return
          } else if (record.status === 3) {
            clearInterval(installPollTimer); installPollTimer = null
            setStepError(3)
            installResult.success = false
            installResult.message = record.message || '安装过程中发生错误'
            installPhase.value = 'done'
            resolve()
            return
          }
        }
      } catch { /* ignore poll error */ }
      if (attempts >= maxAttempts) {
        clearInterval(installPollTimer); installPollTimer = null
        setStepDone(3)
        installResult.success = true
        installResult.message = '安装任务已提交，后台正在处理中'
        installPhase.value = 'done'
        resolve()
      }
    }, 1500)
  })
}



async function handleViewStatus(record) {
  router.push(`/platform/appstore/install/${record.id}`)
}

function goToInstallDetail(installId) {
  handleInstallCancel()
  router.push(`/platform/appstore/install/${installId}`)
}


// ====== 工具方法 ======
function openDocUrl(url) {
  if (url) window.open(url, '_blank')
}

function parseTags(tags) {
  if (!tags) return []
  return tags.split(',').map(t => t.trim()).filter(Boolean)
}
function providerLabel(p) {
  const map = { official: '官方', community: '社区', 'third-party': '第三方' }
  return map[p] || p || '官方'
}
function statusLabel(s) {
  const map = { 1: '可用', 2: '维护中', 3: '已下架' }
  return map[s] || '未知'
}
function statusColor(s) {
  const map = { 1: 'green', 2: 'orange', 3: 'red' }
  return map[s] || 'gray'
}
const categoryIconMap = {
  '网络': '🌐', '监控': '📊', '日志': '📝', '存储': '💾',
  '安全': '🔒', 'GitOps': '🔄', '数据库': '🗄️', '消息队列': '📨'
}
function getCategoryIcon(cat) { return categoryIconMap[cat] || '📦' }

const iconColorMap = {
  'ingress': 'icon-blue', 'nginx': 'icon-green', 'prometheus': 'icon-orange',
  'grafana': 'icon-purple', 'argocd': 'icon-teal', 'elasticsearch': 'icon-yellow',
  'cert-manager': 'icon-red', 'metallb': 'icon-indigo', 'redis': 'icon-red',
  'mysql': 'icon-blue', 'kafka': 'icon-dark', 'harbor': 'icon-teal',
  'loki': 'icon-orange', 'longhorn': 'icon-green'
}
function getIconClass(name) {
  if (!name) return 'icon-blue'
  const n = name.toLowerCase()
  for (const [key, cls] of Object.entries(iconColorMap)) {
    if (n.includes(key)) return cls
  }
  const colors = ['icon-blue', 'icon-green', 'icon-orange', 'icon-purple', 'icon-teal', 'icon-red']
  let hash = 0
  for (let i = 0; i < n.length; i++) hash = n.charCodeAt(i) + ((hash << 5) - hash)
  return colors[Math.abs(hash) % colors.length]
}
function getIconLetter(name) {
  if (!name) return '?'
  return name.charAt(0).toUpperCase()
}

// ====== 生命周期 ======
onMounted(() => {
  fetchApps()
  fetchCategories()
  fetchInstalledApps()
})
onBeforeUnmount(() => {
  if (installPollTimer) { clearInterval(installPollTimer); installPollTimer = null }
})
</script>

<style scoped>
/* ====== 容器 ====== */
.appstore-container { padding: 0; }

/* ====== 顶部标题 ====== */
.appstore-header {
  display: flex; align-items: center; justify-content: space-between; margin-bottom: 20px;
}
.header-left { display: flex; align-items: baseline; gap: 12px; }
.page-title {
  margin: 0; font-size: 20px; font-weight: 700; color: #1d2129;
  display: flex; align-items: center; gap: 8px;
}
.title-icon-wrap {
  width: 32px; height: 32px; border-radius: 8px; display: flex; align-items: center; justify-content: center;
  background: linear-gradient(135deg, #326ce5, #5b8ff9);
}
.title-icon { width: 18px; height: 18px; color: #fff; }
.app-count { font-size: 13px; color: #86909c; }
.header-actions { display: flex; align-items: center; gap: 12px; }
.history-btn { position: relative; }
.badge-dot {
  position: absolute; top: -6px; right: -8px; min-width: 18px; height: 18px; line-height: 18px;
  text-align: center; border-radius: 9px; background: #f53f3f; color: #fff;
  font-size: 11px; font-weight: 600; padding: 0 4px;
}

/* ====== 分类导航 ====== */
.category-bar {
  display: flex; flex-wrap: wrap; gap: 8px; margin-bottom: 20px;
  padding: 12px 16px; background: #f7f8fa; border-radius: 8px;
}
.category-tag {
  display: inline-flex; align-items: center; gap: 6px; padding: 6px 14px;
  border-radius: 16px; font-size: 13px; color: #4e5969;
  background: #fff; border: 1px solid #e5e6eb; cursor: pointer;
  transition: all 0.2s; user-select: none;
}
.category-tag:hover { color: #326ce5; border-color: #326ce5; }
.category-tag.active { color: #fff; background: #326ce5; border-color: #326ce5; }
.category-tag.active .cat-count { background: rgba(255,255,255,0.25); color: #fff; }
.cat-icon { font-size: 14px; }
.cat-count {
  font-size: 11px; padding: 1px 6px; border-radius: 10px; background: #f2f3f5; color: #86909c;
}

/* ====== 加载/空状态 ====== */
.loading-wrap, .empty-wrap {
  display: flex; flex-direction: column; align-items: center;
  justify-content: center; padding: 80px 0; gap: 12px; color: #86909c;
}
.loading-text { font-size: 13px; }

/* ====== 应用卡片网格 ====== */
.app-grid {
  display: grid; grid-template-columns: repeat(auto-fill, minmax(340px, 1fr)); gap: 16px;
}
.app-card {
  position: relative; background: #fff; border: 1px solid #e5e6eb;
  border-radius: 12px; padding: 20px; transition: all 0.25s;
  display: flex; flex-direction: column;
}
.app-card:hover {
  border-color: #c9cdd4; box-shadow: 0 4px 20px rgba(0,0,0,0.08); transform: translateY(-2px);
}
.app-card.featured { border-color: #326ce5; border-width: 1.5px; }
.app-card.installed { border-left: 3px solid #00b42a; }

/* 推荐角标 */
.featured-badge {
  position: absolute; top: 12px; right: 12px; font-size: 11px; padding: 2px 8px;
  border-radius: 4px; background: linear-gradient(135deg, #326ce5, #5b8ff9);
  color: #fff; font-weight: 500; display: flex; align-items: center; gap: 3px;
}
/* 已安装角标 */
.installed-badge {
  position: absolute; top: 12px; right: 12px; font-size: 11px; padding: 2px 8px;
  border-radius: 4px; background: linear-gradient(135deg, #00b42a, #23c343);
  color: #fff; font-weight: 500; display: flex; align-items: center; gap: 3px;
}
.app-card.featured .installed-badge { top: 36px; }

/* 卡片头部 */
.card-header { display: flex; align-items: flex-start; gap: 12px; margin-bottom: 12px; }
.app-icon {
  width: 44px; height: 44px; border-radius: 10px; display: flex; align-items: center;
  justify-content: center; font-size: 20px; font-weight: 700; color: #fff; flex-shrink: 0;
}
.icon-blue { background: linear-gradient(135deg, #326ce5, #5b8ff9); }
.icon-green { background: linear-gradient(135deg, #00b42a, #23c343); }
.icon-orange { background: linear-gradient(135deg, #e8740c, #f59e0b); }
.icon-purple { background: linear-gradient(135deg, #722ed1, #9254de); }
.icon-teal { background: linear-gradient(135deg, #0fc6c2, #14c9c9); }
.icon-red { background: linear-gradient(135deg, #f53f3f, #f76560); }
.icon-yellow { background: linear-gradient(135deg, #e8b900, #fadb14); }
.icon-indigo { background: linear-gradient(135deg, #3730a3, #6366f1); }
.icon-dark { background: linear-gradient(135deg, #374151, #6b7280); }
.app-meta { flex: 1; min-width: 0; }
.app-name {
  font-size: 15px; font-weight: 600; color: #1d2129; line-height: 1.4;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.app-version { display: flex; align-items: center; gap: 6px; margin-top: 4px; }
.version-tag { font-size: 12px; color: #86909c; padding: 1px 6px; background: #f2f3f5; border-radius: 4px; }
.provider-tag { font-size: 11px; padding: 1px 6px; border-radius: 4px; }
.provider-official { background: #e8f3ff; color: #326ce5; }
.provider-community { background: #e8f7ea; color: #00b42a; }
.provider-third-party { background: #fff7e8; color: #e8740c; }
.more-btn { margin-left: auto; color: #86909c; }

/* 标签区 */
.card-tags { display: flex; flex-wrap: wrap; gap: 6px; margin-bottom: 10px; }
.category-badge { font-size: 11px; padding: 2px 8px; border-radius: 4px; background: #f0f1f5; color: #4e5969; font-weight: 500; }
.k8s-badge { font-size: 11px; padding: 2px 8px; border-radius: 4px; background: #e8f3ff; color: #326ce5; }
.tag-item { font-size: 11px; padding: 2px 6px; border-radius: 4px; background: #f7f8fa; color: #86909c; }

/* 描述 */
.card-desc {
  font-size: 13px; color: #4e5969; line-height: 1.6; flex: 1; margin-bottom: 14px;
  display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden;
}

/* 卡片底部 */
.card-footer {
  display: flex; justify-content: flex-end; gap: 8px; padding-top: 12px; border-top: 1px solid #f2f3f5;
}

/* ====== 分页 ====== */
.pagination-wrap { display: flex; justify-content: center; margin-top: 24px; }

/* ====== 详情抽屉 ====== */
.drawer-title-custom { display: flex; align-items: center; gap: 10px; }
.detail-loading {
  display: flex; align-items: center; gap: 8px; padding: 40px 0; justify-content: center; color: #86909c;
}
.detail-action-bar { display: flex; gap: 12px; margin-bottom: 24px; }
.detail-action-bar .arco-btn { flex: 1; }
.detail-section { margin-bottom: 24px; }
.section-title {
  font-size: 13px; font-weight: 600; color: #86909c; text-transform: uppercase;
  letter-spacing: 0.5px; margin: 0 0 12px; padding-bottom: 8px; border-bottom: 1px solid #f2f3f5;
}
.info-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
.info-item { display: flex; flex-direction: column; gap: 4px; }
.info-label { font-size: 12px; color: #86909c; }
.info-value { font-size: 14px; color: #1d2129; font-weight: 500; }
.info-value code { font-size: 13px; background: #f2f3f5; padding: 2px 6px; border-radius: 4px; font-family: monospace; }
.detail-desc { font-size: 14px; color: #4e5969; line-height: 1.8; margin: 0; }
.detail-tags { display: flex; flex-wrap: wrap; gap: 6px; }
.link-cards { display: flex; flex-direction: column; gap: 8px; }
.link-card {
  display: flex; align-items: center; gap: 8px; padding: 10px 14px; border-radius: 8px;
  background: #f7f8fa; color: #4e5969; font-size: 13px; text-decoration: none;
  transition: all 0.2s; border: 1px solid transparent;
}
.link-card:hover { background: #e8f3ff; color: #326ce5; border-color: #bedaff; }
.link-arrow { margin-left: auto; color: #c0c4cc; }
.yaml-block {
  background: #1e1e2e; color: #cdd6f4; padding: 14px; border-radius: 8px;
  font-family: 'JetBrains Mono', 'Fira Code', monospace; font-size: 12px;
  line-height: 1.6; overflow-x: auto; white-space: pre-wrap; margin: 0;
}

/* ====== 安装弹窗 ====== */
.install-modal-header { margin-bottom: 0; }
.install-app-banner {
  display: flex; align-items: flex-start; gap: 16px; padding: 16px;
  background: linear-gradient(135deg, #f6f8fd, #eef3ff); border-radius: 12px;
}
.install-app-info { flex: 1; min-width: 0; }
.install-app-name { font-size: 18px; font-weight: 700; color: #1d2129; margin-bottom: 6px; }
.install-app-meta { display: flex; align-items: center; gap: 6px; margin-bottom: 8px; }
.install-app-cat { font-size: 12px; color: #86909c; }
.install-app-desc {
  font-size: 13px; color: #86909c; line-height: 1.5;
  display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden;
}
.values-editor {
  font-family: 'JetBrains Mono', 'Fira Code', monospace !important; font-size: 12px !important;
}

/* ====== 安装进度面板 ====== */
.install-progress-panel { padding: 20px 0; }
.progress-header { text-align: center; margin-bottom: 32px; }
.progress-spinner { margin: 0 auto 16px; width: 56px; height: 56px; }
.spinner-ring {
  width: 56px; height: 56px; animation: spin 1.2s linear infinite;
}
.spinner-ring circle {
  stroke: #326ce5; stroke-dasharray: 100; stroke-dashoffset: 30; stroke-linecap: round;
}
@keyframes spin { to { transform: rotate(360deg); } }
.progress-title { font-size: 18px; font-weight: 600; color: #1d2129; margin-bottom: 4px; }
.progress-subtitle { font-size: 13px; color: #86909c; }

/* 步骤条 */
.install-steps { padding: 0 20px; }
.step-item { display: flex; gap: 14px; min-height: 56px; }
.step-indicator { display: flex; flex-direction: column; align-items: center; }
.step-dot {
  width: 28px; height: 28px; border-radius: 50%; display: flex; align-items: center;
  justify-content: center; flex-shrink: 0; transition: all 0.3s;
  background: #f2f3f5; border: 2px solid #e5e6eb;
}
.step-item.done .step-dot { background: #00b42a; border-color: #00b42a; }
.step-item.active .step-dot { background: #326ce5; border-color: #326ce5; animation: pulse-ring 1.5s infinite; }
.step-item.error .step-dot { background: #f53f3f; border-color: #f53f3f; }
.dot-num { font-size: 12px; font-weight: 600; color: #86909c; }
.dot-pulse {
  width: 8px; height: 8px; border-radius: 50%; background: #fff;
  animation: dot-blink 1s ease-in-out infinite;
}
@keyframes dot-blink { 0%, 100% { opacity: 1; } 50% { opacity: 0.3; } }
@keyframes pulse-ring {
  0% { box-shadow: 0 0 0 0 rgba(50,108,229,0.4); }
  70% { box-shadow: 0 0 0 8px rgba(50,108,229,0); }
  100% { box-shadow: 0 0 0 0 rgba(50,108,229,0); }
}
.step-line {
  width: 2px; flex: 1; min-height: 20px; background: #e5e6eb; margin: 4px 0; transition: background 0.3s;
}
.step-line.filled { background: #00b42a; }
.step-content { padding-top: 3px; padding-bottom: 12px; }
.step-label { font-size: 14px; font-weight: 500; color: #1d2129; }
.step-item.pending .step-label { color: #c0c4cc; }
.step-item.active .step-label { color: #326ce5; }
.step-item.error .step-label { color: #f53f3f; }
.step-desc { font-size: 12px; color: #86909c; margin-top: 2px; }
.step-item.pending .step-desc { color: #c0c4cc; }

/* ====== 安装结果面板 ====== */
.install-result-panel { text-align: center; padding: 24px 0; }
.result-icon { margin: 0 auto 16px; width: 64px; height: 64px; display: flex; align-items: center; justify-content: center; }
.result-icon.success { color: #00b42a; animation: result-pop 0.5s cubic-bezier(0.175,0.885,0.32,1.275); }
.result-icon.error { color: #f53f3f; animation: result-pop 0.5s cubic-bezier(0.175,0.885,0.32,1.275); }
@keyframes result-pop { 0% { transform: scale(0); } 100% { transform: scale(1); } }
.result-title { font-size: 20px; font-weight: 700; color: #1d2129; margin-bottom: 6px; }
.result-message { font-size: 13px; color: #86909c; margin-bottom: 20px; }
.result-summary {
  background: #f7f8fa; border-radius: 10px; padding: 16px 20px;
  text-align: left; margin: 0 auto 20px; max-width: 380px;
}
.summary-row {
  display: flex; justify-content: space-between; align-items: center;
  padding: 6px 0; font-size: 13px;
}
.summary-row + .summary-row { border-top: 1px solid #f2f3f5; }
.summary-label { color: #86909c; }
.summary-value { color: #1d2129; font-weight: 500; }
.summary-value code { background: #e8f3ff; padding: 2px 6px; border-radius: 4px; font-family: monospace; font-size: 12px; }
.result-actions { display: flex; gap: 10px; justify-content: center; }

/* 下拉危险选项 */
.danger-opt { color: #f53f3f !important; }

/* ====== 组件列表 ====== */
.section-title {
  display: flex; align-items: center;
}
.comp-count {
  font-size: 11px; padding: 1px 8px; border-radius: 10px;
  background: #e8f3ff; color: #326ce5; margin-left: 8px; font-weight: 500;
}
.comp-empty {
  text-align: center; padding: 20px; color: #c0c4cc; font-size: 13px;
  background: #fafafa; border-radius: 8px; border: 1px dashed #e5e6eb;
}
.comp-list { display: flex; flex-direction: column; gap: 8px; }
.comp-item {
  padding: 10px 14px; border-radius: 8px; background: #f7f8fa;
  border: 1px solid #e5e6eb; transition: border-color 0.2s;
}
.comp-item:hover { border-color: #c9cdd4; }
.comp-header {
  display: flex; align-items: center; gap: 8px; margin-bottom: 4px;
}
.comp-name { font-size: 14px; font-weight: 600; color: #1d2129; }
.comp-replicas {
  font-size: 11px; padding: 1px 6px; border-radius: 4px;
  background: #e8f7ea; color: #00b42a; font-weight: 500;
}
.comp-actions { margin-left: auto; display: flex; gap: 2px; }
.comp-image {
  font-size: 12px; color: #326ce5; font-family: 'JetBrains Mono', monospace;
  margin-bottom: 4px; word-break: break-all;
}
.comp-resources {
  display: flex; gap: 16px; font-size: 11px; color: #86909c;
}

/* ====== 表格视图 ====== */
.table-view {
  background: #fff; border-radius: 12px; border: 1px solid #e5e6eb;
  overflow: hidden;
}
.table-toolbar {
  display: flex; align-items: center; gap: 10px; padding: 10px 16px;
  background: #e8f3ff; border-bottom: 1px solid #bedaff;
}
.selected-info {
  font-size: 13px; color: #326ce5; font-weight: 500;
}

/* ====== 组件排序标记 ====== */
.comp-sort-badge {
  font-size: 10px; padding: 1px 5px; border-radius: 3px;
  background: #f2f3f5; color: #86909c; font-family: monospace;
}

</style>
