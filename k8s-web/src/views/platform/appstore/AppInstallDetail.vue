<template>
  <div class="install-detail-page">
    <!-- ====== 顶部面包屑导航 ====== -->
    <div class="page-breadcrumb">
      <a-breadcrumb>
        <a-breadcrumb-item>
          <router-link to="/platform/appstore" class="breadcrumb-link">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="3" width="7" height="7" rx="1.5"/><rect x="14" y="3" width="7" height="7" rx="1.5"/>
              <rect x="3" y="14" width="7" height="7" rx="1.5"/><rect x="14" y="14" width="7" height="7" rx="1.5"/>
            </svg>
            应用商城
          </router-link>
        </a-breadcrumb-item>
        <a-breadcrumb-item>部署详情</a-breadcrumb-item>
      </a-breadcrumb>
    </div>

    <!-- ====== 加载状态 ====== -->
    <div v-if="pageLoading" class="page-loading">
      <div class="loading-spinner">
        <svg viewBox="0 0 50 50"><circle cx="25" cy="25" r="20" fill="none" stroke-width="3"/></svg>
      </div>
      <span>加载部署信息...</span>
    </div>

    <!-- ====== 错误状态 ====== -->
    <div v-else-if="loadError" class="page-error">
      <svg width="48" height="48" viewBox="0 0 48 48" fill="none">
        <circle cx="24" cy="24" r="22" stroke="#f53f3f" stroke-width="2"/>
        <line x1="16" y1="16" x2="32" y2="32" stroke="#f53f3f" stroke-width="3" stroke-linecap="round"/>
        <line x1="32" y1="16" x2="16" y2="32" stroke="#f53f3f" stroke-width="3" stroke-linecap="round"/>
      </svg>
      <div class="error-title">加载失败</div>
      <div class="error-msg">{{ loadError }}</div>
      <a-button type="primary" @click="loadData">重试</a-button>
    </div>

    <!-- ====== 主内容 ====== -->
    <template v-else-if="installRecord">
      <!-- 顶部应用信息横幅 -->
      <div class="app-banner" :class="bannerStatusClass">
        <div class="banner-left">
          <div class="banner-icon" :class="getIconClass(installRecord.app_name)">
            {{ getIconLetter(installRecord.app_name) }}
          </div>
          <div class="banner-info">
            <div class="banner-title">{{ installRecord.app_name }}</div>
            <div class="banner-meta">
              <a-tag size="small" color="arcoblue">v{{ installRecord.version }}</a-tag>
              <a-tag size="small" :color="installStatusColor(installRecord.status)">
                {{ installStatusLabel(installRecord.status) }}
              </a-tag>
              <span class="banner-time" v-if="installRecord.created_at">
                {{ formatTimestamp(installRecord.created_at) }}
              </span>
            </div>
          </div>
        </div>
        <div class="banner-actions">
          <a-button size="small" @click="refreshAll" :loading="refreshing">
            <template #icon><icon-refresh /></template>
            刷新
          </a-button>
          <a-button
            v-if="installRecord.status === 2"
            size="small" type="primary"
            @click="openEditModal"
          >
            <template #icon><icon-edit /></template>
            编辑
          </a-button>
          <a-button
            v-if="installRecord.status === 2 || installRecord.status === 3"
            size="small" type="outline" status="danger"
            @click="handleUninstall"
            :loading="uninstalling"
          >
            <template #icon><icon-delete /></template>
            卸载
          </a-button>
          <a-button size="small" @click="$router.push('/platform/appstore/records')">
            <template #icon><icon-arrow-left /></template>
            返回列表
          </a-button>
        </div>
      </div>

      <!-- 部署概览卡片网格 -->
      <div class="overview-grid">
        <div class="overview-card">
          <div class="ov-label">集群</div>
          <div class="ov-value">
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none" stroke="#326ce5" stroke-width="1.5"><circle cx="7" cy="7" r="5.5"/><circle cx="7" cy="7" r="2"/></svg>
            {{ installRecord.cluster_name }}
          </div>
        </div>
        <div class="overview-card">
          <div class="ov-label">命名空间</div>
          <div class="ov-value">
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none" stroke="#326ce5" stroke-width="1.5"><rect x="1.5" y="3.5" width="11" height="7.5" rx="1.5"/><line x1="1.5" y1="6.5" x2="12.5" y2="6.5"/></svg>
            <code>{{ installRecord.namespace }}</code>
          </div>
        </div>
        <div class="overview-card">
          <div class="ov-label">Release</div>
          <div class="ov-value">
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none" stroke="#326ce5" stroke-width="1.5"><path d="M2.5,11 L7,2.5 L11.5,11 Z"/></svg>
            <code>{{ installRecord.release_name }}</code>
          </div>
        </div>
        <div class="overview-card">
          <div class="ov-label">数据库状态</div>
          <div class="ov-value">
            <span class="status-dot" :class="'dot-s' + installRecord.status"></span>
            {{ installStatusLabel(installRecord.status) }}
          </div>
        </div>
      </div>

      <!-- 数据库消息 -->
      <div v-if="installRecord.message" class="db-message" :class="{ error: installRecord.status === 3 }">
        <svg width="14" height="14" viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="7" cy="7" r="5.5"/><line x1="7" y1="4.5" x2="7" y2="7.5"/><circle cx="7" cy="9.5" r="0.5" fill="currentColor"/></svg>
        {{ installRecord.message }}
      </div>

      <!-- ====== K8s 集群实时状态 ====== -->
      <div v-if="statusData" class="k8s-status-section">
        <div class="section-header">
          <h3 class="section-title">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#326ce5" stroke-width="2"><circle cx="12" cy="12" r="10"/><polyline points="12,6 12,12 16,14"/></svg>
            集群资源状态
          </h3>
          <div class="section-actions">
            <span class="auto-refresh-badge" v-if="autoRefreshEnabled">
              <span class="pulse-dot-mini"></span> 自动刷新
            </span>
            <a-switch v-model="autoRefreshEnabled" size="small" @change="toggleAutoRefresh">
              <template #checked>ON</template>
              <template #unchecked>OFF</template>
            </a-switch>
          </div>
        </div>

        <!-- 集群连接失败 -->
        <a-alert v-if="!statusData.cluster_reachable" type="error" class="cluster-alert">
          <template #title>集群连接失败</template>
          {{ statusData.cluster_error }}
        </a-alert>

        <template v-else>
          <!-- Deployment 资源卡片 -->
          <div class="resource-card">
            <div class="resource-header">
              <div class="resource-icon deploy-icon">
                <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="2" y="2" width="16" height="16" rx="3"/><line x1="6" y1="7" x2="14" y2="7"/><line x1="6" y1="10" x2="14" y2="10"/><line x1="6" y1="13" x2="11" y2="13"/></svg>
              </div>
              <div class="resource-title">
                <span>Deployment</span>
                <code>{{ statusData.release_name }}</code>
              </div>
              <a-tag :color="deployStatusColor(statusData.deployment_status)" size="small">
                {{ statusData.deployment_status || 'Unknown' }}
              </a-tag>
            </div>
            <div v-if="statusData.deployment_message" class="resource-message">
              {{ statusData.deployment_message }}
            </div>
            <!-- 副本进度条 -->
            <div class="replica-bar-section" v-if="statusData.deployment_status !== 'NotFound'">
              <div class="replica-labels">
                <span>副本就绪</span>
                <span class="replica-nums">{{ statusData.ready_replicas || 0 }} / {{ statusData.desired_replicas || 0 }}</span>
              </div>
              <div class="replica-bar">
                <div class="replica-fill" :style="{ width: replicaPercent + '%' }"></div>
              </div>
              <div class="replica-detail-row">
                <div class="replica-chip">
                  <span class="chip-dot" style="background:#00b42a"></span>
                  Ready: {{ statusData.ready_replicas || 0 }}
                </div>
                <div class="replica-chip">
                  <span class="chip-dot" style="background:#326ce5"></span>
                  Updated: {{ statusData.updated_replicas || 0 }}
                </div>
                <div class="replica-chip">
                  <span class="chip-dot" style="background:#ff7d00"></span>
                  Available: {{ statusData.available_replicas || 0 }}
                </div>
              </div>
            </div>
          </div>

          <!-- Service 资源卡片 -->
          <div class="resource-card">
            <div class="resource-header">
              <div class="resource-icon svc-icon">
                <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="10" cy="10" r="7"/><circle cx="10" cy="10" r="3"/></svg>
              </div>
              <div class="resource-title">
                <span>Service</span>
                <code>{{ statusData.release_name }}</code>
              </div>
              <a-tag :color="statusData.service_status === 'Active' ? 'green' : 'gray'" size="small">
                {{ statusData.service_status || 'Unknown' }}
              </a-tag>
            </div>
            <div v-if="statusData.service_status === 'Active'" class="svc-detail-grid">
              <div class="svc-item">
                <span class="svc-label">Type</span>
                <span class="svc-value">{{ statusData.service_type }}</span>
              </div>
              <div class="svc-item">
                <span class="svc-label">ClusterIP</span>
                <span class="svc-value"><code>{{ statusData.cluster_ip }}</code></span>
              </div>
              <div class="svc-item" v-if="statusData.service_ports?.length">
                <span class="svc-label">Ports</span>
                <span class="svc-value">
                  <a-tag v-for="p in statusData.service_ports" :key="p" size="small" color="arcoblue">{{ p }}</a-tag>
                </span>
              </div>
            </div>
            <div v-else class="resource-message">Service 未创建或已被删除</div>
          </div>

          <!-- Pod 列表卡片 -->
          <div class="resource-card pods-card">
            <div class="resource-header">
              <div class="resource-icon pod-icon">
                <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="3" y="5" width="14" height="10" rx="2"/><circle cx="7" cy="10" r="1.5" fill="currentColor"/><circle cx="13" cy="10" r="1.5" fill="currentColor"/></svg>
              </div>
              <div class="resource-title">
                <span>Pods</span>
                <span class="pod-count-badge">{{ statusData.pods?.length || 0 }}</span>
              </div>
            </div>

            <div v-if="!statusData.pods?.length" class="no-pods">
              <svg width="36" height="36" viewBox="0 0 36 36" fill="none" stroke="#c0c4cc" stroke-width="1.5"><rect x="4" y="8" width="28" height="20" rx="4"/><line x1="12" y1="18" x2="24" y2="18" stroke-dasharray="3 2"/></svg>
              <span>暂无 Pod 实例</span>
            </div>

            <!-- 每个 Pod -->
            <div v-for="pod in statusData.pods" :key="pod.name" class="pod-item" :class="'phase-' + pod.phase?.toLowerCase()">
              <div class="pod-header">
                <span class="pod-phase-dot" :class="'phase-' + pod.phase?.toLowerCase()"></span>
                <span class="pod-name">{{ pod.name }}</span>
                <a-tag :color="podPhaseColor(pod.phase)" size="small">{{ pod.phase }}</a-tag>
              </div>
              <div class="pod-info-grid">
                <div class="pod-info-item">
                  <span class="pi-label">Node</span>
                  <span class="pi-value">{{ pod.node_name || '-' }}</span>
                </div>
                <div class="pod-info-item">
                  <span class="pi-label">Pod IP</span>
                  <span class="pi-value"><code>{{ pod.pod_ip || '-' }}</code></span>
                </div>
                <div class="pod-info-item">
                  <span class="pi-label">启动时间</span>
                  <span class="pi-value">{{ formatTime(pod.start_time) }}</span>
                </div>
                <div class="pod-info-item">
                  <span class="pi-label">重启次数</span>
                  <span class="pi-value" :class="{ 'restart-warn': pod.restarts > 0 }">{{ pod.restarts }}</span>
                </div>
              </div>

              <!-- 容器详情 -->
              <div v-if="pod.containers?.length" class="container-list">
                <div class="container-list-title">容器 ({{ pod.containers.length }})</div>
                <div v-for="c in pod.containers" :key="c.name" class="container-row">
                  <div class="container-header">
                    <span class="container-ready-dot" :class="{ ready: c.ready }"></span>
                    <span class="container-name">{{ c.name }}</span>
                    <a-tag :color="containerStateColor(c.state)" size="mini">{{ c.state || 'Unknown' }}</a-tag>
                    <span v-if="c.restart_count > 0" class="restart-badge">{{ c.restart_count }}x restart</span>
                  </div>
                  <div class="container-detail">
                    <span class="cd-label">Image:</span>
                    <code class="cd-image">{{ c.image }}</code>
                  </div>
                  <div v-if="c.reason" class="container-detail warn">
                    <span class="cd-label">Reason:</span>
                    <span>{{ c.reason }}</span>
                    <span v-if="c.message" class="cd-msg">- {{ c.message }}</span>
                  </div>
                  <div v-if="c.started_at" class="container-detail">
                    <span class="cd-label">Started:</span>
                    <span>{{ formatTime(c.started_at) }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- ConfigMap 资源卡片 -->
          <div v-if="statusData.configmaps?.length" class="resource-card">
            <div class="resource-header">
              <div class="resource-icon cm-icon">
                <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="3" y="2" width="14" height="16" rx="2"/><line x1="7" y1="6" x2="13" y2="6"/><line x1="7" y1="9" x2="13" y2="9"/><line x1="7" y1="12" x2="11" y2="12"/></svg>
              </div>
              <div class="resource-title">
                <span>ConfigMaps</span>
                <span class="pod-count-badge" style="background:#722ed1">{{ statusData.configmaps.length }}</span>
              </div>
            </div>
            <div v-for="cm in statusData.configmaps" :key="cm.name" class="cm-item">
              <div class="cm-header">
                <code class="cm-name">{{ cm.name }}</code>
                <span v-if="cm.created_at" class="cm-time">{{ formatTime(cm.created_at) }}</span>
              </div>
              <div v-if="cm.data" class="cm-data-grid">
                <div v-for="(val, key) in cm.data" :key="key" class="cm-data-row">
                  <span class="cm-key">{{ key }}</span>
                  <span class="cm-val">{{ val }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- ====== 命名空间资源概览 ====== -->
          <div v-if="statusData.namespace_overview" class="ns-overview-section">
            <div class="section-header" style="margin-top:24px">
              <h3 class="section-title">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#165dff" stroke-width="2"><rect x="3" y="3" width="7" height="7" rx="1.5"/><rect x="14" y="3" width="7" height="7" rx="1.5"/><rect x="3" y="14" width="7" height="7" rx="1.5"/><rect x="14" y="14" width="7" height="7" rx="1.5"/></svg>
                命名空间资源概览
                <code style="font-size:12px;font-weight:400;margin-left:4px">{{ statusData.namespace }}</code>
              </h3>
            </div>
            <div class="ns-stats-grid">
              <div class="ns-stat-card">
                <div class="ns-stat-num">{{ statusData.namespace_overview.total_deployments }}</div>
                <div class="ns-stat-label">Deployments</div>
              </div>
              <div class="ns-stat-card">
                <div class="ns-stat-num">{{ statusData.namespace_overview.total_services }}</div>
                <div class="ns-stat-label">Services</div>
              </div>
              <div class="ns-stat-card">
                <div class="ns-stat-num">
                  <span style="color:#00b42a">{{ statusData.namespace_overview.running_pods }}</span>
                  <span style="color:#c0c4cc;font-size:14px"> / {{ statusData.namespace_overview.total_pods }}</span>
                </div>
                <div class="ns-stat-label">Pods (Running/Total)</div>
              </div>
              <div class="ns-stat-card" v-if="statusData.namespace_overview.pending_pods > 0 || statusData.namespace_overview.failed_pods > 0">
                <div class="ns-stat-num">
                  <span v-if="statusData.namespace_overview.pending_pods" style="color:#ff7d00">{{ statusData.namespace_overview.pending_pods }} Pending</span>
                  <span v-if="statusData.namespace_overview.failed_pods" style="color:#f53f3f;margin-left:8px">{{ statusData.namespace_overview.failed_pods }} Failed</span>
                </div>
                <div class="ns-stat-label">异常 Pod</div>
              </div>
              <div class="ns-stat-card" v-else>
                <div class="ns-stat-num">{{ statusData.namespace_overview.total_configmaps }}</div>
                <div class="ns-stat-label">ConfigMaps</div>
              </div>
            </div>
          </div>

          <!-- ====== 命名空间所有 Deployments ====== -->
          <div v-if="statusData.all_deployments?.length" class="resource-card" style="margin-top:16px">
            <div class="resource-header">
              <div class="resource-icon deploy-icon">
                <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="2" y="2" width="16" height="16" rx="3"/><line x1="6" y1="7" x2="14" y2="7"/><line x1="6" y1="10" x2="14" y2="10"/><line x1="6" y1="13" x2="11" y2="13"/></svg>
              </div>
              <div class="resource-title">
                <span>All Deployments</span>
                <span class="pod-count-badge" style="background:#326ce5">{{ statusData.all_deployments.length }}</span>
              </div>
            </div>
            <div class="ns-table">
              <div class="ns-table-header">
                <span class="ns-col-name">名称</span>
                <span class="ns-col-status">状态</span>
                <span class="ns-col-replicas">副本</span>
                <span class="ns-col-image">镜像</span>
                <span class="ns-col-age">创建时间</span>
              </div>
              <div v-for="d in statusData.all_deployments" :key="d.name" class="ns-table-row" :class="{'row-highlight': d.name === statusData.release_name}">
                <span class="ns-col-name">
                  <span class="ns-row-dot" :class="'dot-' + d.status?.toLowerCase()"></span>
                  <code>{{ d.name }}</code>
                  <a-tag v-if="d.name === statusData.release_name" size="mini" color="arcoblue" style="margin-left:4px">当前</a-tag>
                </span>
                <span class="ns-col-status">
                  <a-tag :color="deployStatusColor(d.status)" size="small">{{ d.status }}</a-tag>
                </span>
                <span class="ns-col-replicas">{{ d.ready_replicas }}/{{ d.replicas }}</span>
                <span class="ns-col-image"><code class="img-code">{{ shortenImage(d.image) }}</code></span>
                <span class="ns-col-age">{{ formatTime(d.created_at) }}</span>
              </div>
            </div>
          </div>

          <!-- ====== 命名空间所有 Services ====== -->
          <div v-if="statusData.all_services?.length" class="resource-card">
            <div class="resource-header">
              <div class="resource-icon svc-icon">
                <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="10" cy="10" r="7"/><circle cx="10" cy="10" r="3"/></svg>
              </div>
              <div class="resource-title">
                <span>All Services</span>
                <span class="pod-count-badge" style="background:#00b42a">{{ statusData.all_services.length }}</span>
              </div>
            </div>
            <div class="ns-table">
              <div class="ns-table-header">
                <span class="ns-col-name">名称</span>
                <span class="ns-col-status">类型</span>
                <span class="ns-col-replicas">ClusterIP</span>
                <span class="ns-col-image">端口</span>
                <span class="ns-col-age">创建时间</span>
              </div>
              <div v-for="svc in statusData.all_services" :key="svc.name" class="ns-table-row" :class="{'row-highlight': svc.name === statusData.release_name}">
                <span class="ns-col-name">
                  <span class="ns-row-dot dot-running"></span>
                  <code>{{ svc.name }}</code>
                  <a-tag v-if="svc.name === statusData.release_name" size="mini" color="green" style="margin-left:4px">当前</a-tag>
                </span>
                <span class="ns-col-status"><a-tag size="small" color="arcoblue">{{ svc.type }}</a-tag></span>
                <span class="ns-col-replicas"><code>{{ svc.cluster_ip }}</code></span>
                <span class="ns-col-image">
                  <a-tag v-for="p in svc.ports" :key="p" size="mini" color="arcoblue" style="margin:1px">{{ p }}</a-tag>
                </span>
                <span class="ns-col-age">{{ formatTime(svc.created_at) }}</span>
              </div>
            </div>
          </div>

          <!-- ====== Events 事件列表 ====== -->
          <div v-if="statusData.events?.length" class="resource-card events-card">
            <div class="resource-header">
              <div class="resource-icon event-icon">
                <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M11 2L4 12h6l-1 6 7-10h-6l1-6z"/></svg>
              </div>
              <div class="resource-title">
                <span>Events</span>
                <span class="pod-count-badge" style="background:#ff7d00">{{ statusData.events.length }}</span>
              </div>
            </div>
            <div class="events-list">
              <div v-for="(ev, idx) in statusData.events" :key="idx" class="event-row" :class="{ 'event-warning': ev.type === 'Warning' }">
                <span class="ev-type-badge" :class="ev.type?.toLowerCase()">{{ ev.type }}</span>
                <span class="ev-reason">{{ ev.reason }}</span>
                <span class="ev-object"><code>{{ ev.object }}</code></span>
                <span class="ev-message">{{ ev.message }}</span>
                <span class="ev-count" v-if="ev.count > 1">x{{ ev.count }}</span>
                <span class="ev-time">{{ formatTime(ev.last_time) }}</span>
              </div>
            </div>
          </div>
        </template>
      </div>

      <!-- 加载状态中 -->
      <div v-else-if="statusLoading" class="status-loading-section">
        <a-spin size="24" />
        <span>正在查询集群资源状态...</span>
      </div>

      <!-- 无状态数据 -->
      <div v-else class="no-status-section">
        <svg width="40" height="40" viewBox="0 0 40 40" fill="none" stroke="#c0c4cc" stroke-width="1.5"><circle cx="20" cy="20" r="16"/><line x1="20" y1="12" x2="20" y2="22"/><circle cx="20" cy="26" r="1" fill="#c0c4cc"/></svg>
        <span>暂无集群资源状态</span>
        <a-button size="small" type="primary" @click="fetchStatus">查询状态</a-button>
      </div>
    </template>

    <!-- ====== 编辑弹窗 ====== -->
    <a-modal
      v-model:visible="showEditModal"
      title="编辑 Deployment"
      :ok-loading="editSaving"
      @ok="handleEditSubmit"
      @cancel="showEditModal = false"
      :width="560"
    >
      <a-form :model="editForm" layout="vertical">
        <a-form-item label="副本数 (Replicas)">
          <a-input-number v-model="editForm.replicas" :min="0" :max="20" style="width:100%" placeholder="副本数量" />
        </a-form-item>
        <a-form-item label="容器镜像 (Image)">
          <a-input v-model="editForm.image" placeholder="例如: nginx:1.25-alpine" />
        </a-form-item>
        <a-divider>资源配置</a-divider>
        <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px">
          <a-form-item label="CPU Request">
            <a-input v-model="editForm.cpu_request" placeholder="50m" />
          </a-form-item>
          <a-form-item label="CPU Limit">
            <a-input v-model="editForm.cpu_limit" placeholder="200m" />
          </a-form-item>
          <a-form-item label="Memory Request">
            <a-input v-model="editForm.memory_request" placeholder="64Mi" />
          </a-form-item>
          <a-form-item label="Memory Limit">
            <a-input v-model="editForm.memory_limit" placeholder="256Mi" />
          </a-form-item>
        </div>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Message, Modal } from '@arco-design/web-vue'
import {
  IconRefresh, IconDelete, IconArrowLeft, IconEdit
} from '@arco-design/web-vue/es/icon'
import { getInstallDetail, getInstallStatus, uninstallApp, updateInstall } from '@/api/platform/appstore.js'

const route = useRoute()
const router = useRouter()

// ====== 状态 ======
const pageLoading = ref(true)
const loadError = ref('')
const installRecord = ref(null)
const statusData = ref(null)
const statusLoading = ref(false)
const refreshing = ref(false)
const uninstalling = ref(false)
const autoRefreshEnabled = ref(true)
let autoRefreshTimer = null

const installId = computed(() => {
  return parseInt(route.params.id) || 0
})

const bannerStatusClass = computed(() => {
  if (!installRecord.value) return ''
  const s = installRecord.value.status
  if (s === 2) return 'banner-success'
  if (s === 3) return 'banner-error'
  if (s === 1 || s === 4) return 'banner-progress'
  return 'banner-default'
})

const replicaPercent = computed(() => {
  if (!statusData.value || !statusData.value.desired_replicas) return 0
  return Math.round((statusData.value.ready_replicas / statusData.value.desired_replicas) * 100)
})

// ====== 数据加载 ======
async function loadData() {
  pageLoading.value = true
  loadError.value = ''
  try {
    const res = await getInstallDetail(installId.value)
    installRecord.value = res?.data
    if (!installRecord.value) {
      loadError.value = '安装记录不存在'
      return
    }
    // 加载 K8s 状态
    await fetchStatus()
  } catch (e) {
    loadError.value = e?.msg || e?.message || '加载失败'
  } finally {
    pageLoading.value = false
  }
}

async function fetchStatus() {
  if (!installId.value) return
  statusLoading.value = true
  try {
    const res = await getInstallStatus(installId.value)
    statusData.value = res?.data || null
  } catch (e) {
    console.error('查询状态失败:', e)
  } finally {
    statusLoading.value = false
  }
}

async function refreshAll() {
  refreshing.value = true
  try {
    const [recordRes] = await Promise.all([
      getInstallDetail(installId.value),
      fetchStatus()
    ])
    installRecord.value = recordRes?.data || installRecord.value
  } catch (e) {
    console.error('刷新失败:', e)
  } finally {
    refreshing.value = false
  }
}

// ====== 自动刷新 ======
function toggleAutoRefresh(val) {
  if (val) {
    startAutoRefresh()
  } else {
    stopAutoRefresh()
  }
}

function startAutoRefresh() {
  stopAutoRefresh()
  autoRefreshTimer = setInterval(async () => {
    try {
      const res = await getInstallStatus(installId.value)
      statusData.value = res?.data || null
    } catch { /* ignore */ }
  }, 5000)
}

function stopAutoRefresh() {
  if (autoRefreshTimer) {
    clearInterval(autoRefreshTimer)
    autoRefreshTimer = null
  }
}

// ====== 编辑 ======
const showEditModal = ref(false)
const editSaving = ref(false)
const editForm = reactive({
  replicas: null,
  image: '',
  cpu_request: '',
  cpu_limit: '',
  memory_request: '',
  memory_limit: '',
})

function openEditModal() {
  // 预填充当前值
  if (statusData.value) {
    editForm.replicas = statusData.value.desired_replicas || 1
    // 从第一个 pod 的第一个容器获取当前镜像
    const firstPod = statusData.value.pods?.[0]
    const firstContainer = firstPod?.containers?.[0]
    editForm.image = firstContainer?.image || ''
  }
  editForm.cpu_request = ''
  editForm.cpu_limit = ''
  editForm.memory_request = ''
  editForm.memory_limit = ''
  showEditModal.value = true
}

async function handleEditSubmit() {
  editSaving.value = true
  try {
    const payload = {}
    if (editForm.replicas !== null && editForm.replicas !== undefined) {
      payload.replicas = editForm.replicas
    }
    if (editForm.image) payload.image = editForm.image
    if (editForm.cpu_request) payload.cpu_request = editForm.cpu_request
    if (editForm.cpu_limit) payload.cpu_limit = editForm.cpu_limit
    if (editForm.memory_request) payload.memory_request = editForm.memory_request
    if (editForm.memory_limit) payload.memory_limit = editForm.memory_limit

    await updateInstall(installId.value, payload)
    Message.success('更新成功，Deployment 正在滚动更新')
    showEditModal.value = false
    // 延迟刷新状态
    setTimeout(() => refreshAll(), 2000)
  } catch (e) {
    Message.error('更新失败: ' + (e?.msg || e?.message || ''))
  } finally {
    editSaving.value = false
  }
}

// ====== 卸载 ======
function handleUninstall() {
  if (!installRecord.value) return
  const r = installRecord.value
  Modal.warning({
    title: '确认卸载',
    content: `确定要卸载「${r.app_name}」(${r.release_name} @ ${r.cluster_name}/${r.namespace})？将删除 Deployment、Service 和 ConfigMap。`,
    okText: '确认卸载',
    cancelText: '取消',
    hideCancel: false,
    onOk: async () => {
      uninstalling.value = true
      try {
        await uninstallApp(r.id)
        Message.success('卸载成功')
        // 刷新数据
        await refreshAll()
      } catch (e) {
        Message.error('卸载失败: ' + (e?.msg || e?.message || ''))
      } finally {
        uninstalling.value = false
      }
    }
  })
}

// ====== 工具方法 ======
function installStatusLabel(s) {
  const m = { 1: '安装中', 2: '运行中', 3: '安装失败', 4: '卸载中', 5: '已卸载' }
  return m[s] || '未知'
}
function installStatusColor(s) {
  const m = { 1: 'arcoblue', 2: 'green', 3: 'red', 4: 'orangered', 5: 'gray' }
  return m[s] || 'gray'
}
function deployStatusColor(status) {
  const m = { Running: 'green', PartialReady: 'orange', Pending: 'arcoblue', Failed: 'red', NotFound: 'gray', Error: 'red' }
  return m[status] || 'gray'
}
function podPhaseColor(phase) {
  const m = { Running: 'green', Succeeded: 'green', Pending: 'orange', Failed: 'red', Unknown: 'gray' }
  return m[phase] || 'gray'
}
function containerStateColor(state) {
  const m = { Running: 'green', Waiting: 'orange', Terminated: 'red' }
  return m[state] || 'gray'
}
function formatTime(isoStr) {
  if (!isoStr) return '-'
  try {
    const d = new Date(isoStr)
    return d.toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit' })
  } catch { return isoStr }
}
function formatTimestamp(ts) {
  if (!ts) return ''
  try {
    const d = new Date(ts * 1000)
    return d.toLocaleString('zh-CN')
  } catch { return '' }
}

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

function shortenImage(img) {
  if (!img) return '-'
  // 截取最后的 image:tag 部分
  const parts = img.split('/')
  return parts[parts.length - 1] || img
}

// ====== 生命周期 ======
onMounted(() => {
  loadData()
  if (autoRefreshEnabled.value) {
    startAutoRefresh()
  }
})

onBeforeUnmount(() => {
  stopAutoRefresh()
})

watch(() => route.params.id, (newId) => {
  if (newId) {
    stopAutoRefresh()
    loadData()
    if (autoRefreshEnabled.value) {
      startAutoRefresh()
    }
  }
})
</script>

<style scoped>
/* ====== 页面容器 ====== */
.install-detail-page { padding: 0; max-width: 960px; }

/* ====== 面包屑 ====== */
.page-breadcrumb { margin-bottom: 20px; }
.breadcrumb-link {
  display: inline-flex; align-items: center; gap: 6px; color: #4e5969; text-decoration: none;
}
.breadcrumb-link:hover { color: #326ce5; }

/* ====== 加载 / 错误 ====== */
.page-loading, .page-error {
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  padding: 100px 0; gap: 16px; color: #86909c;
}
.loading-spinner { width: 48px; height: 48px; }
.loading-spinner svg {
  width: 48px; height: 48px; animation: spin 1.2s linear infinite;
}
.loading-spinner circle {
  stroke: #326ce5; stroke-dasharray: 100; stroke-dashoffset: 30; stroke-linecap: round;
}
@keyframes spin { to { transform: rotate(360deg); } }
.error-title { font-size: 18px; font-weight: 600; color: #1d2129; }
.error-msg { font-size: 13px; color: #86909c; }

/* ====== 应用横幅 ====== */
.app-banner {
  display: flex; align-items: center; justify-content: space-between;
  padding: 20px 24px; border-radius: 12px; margin-bottom: 20px;
  background: linear-gradient(135deg, #f6f8fd, #eef3ff); border: 1px solid #e5e6eb;
}
.app-banner.banner-success { background: linear-gradient(135deg, #f0faf2, #e8f7ea); border-color: #b8e6c0; }
.app-banner.banner-error { background: linear-gradient(135deg, #fff5f3, #ffece8); border-color: #ffc4b8; }
.app-banner.banner-progress { background: linear-gradient(135deg, #f0f5ff, #e8f0ff); border-color: #b8d4ff; }
.banner-left { display: flex; align-items: center; gap: 16px; }
.banner-icon {
  width: 52px; height: 52px; border-radius: 14px; display: flex; align-items: center;
  justify-content: center; font-size: 24px; font-weight: 700; color: #fff; flex-shrink: 0;
}
.banner-title { font-size: 20px; font-weight: 700; color: #1d2129; margin-bottom: 6px; }
.banner-meta { display: flex; align-items: center; gap: 8px; }
.banner-time { font-size: 12px; color: #86909c; }
.banner-actions { display: flex; gap: 8px; }

/* icon colors */
.icon-blue { background: linear-gradient(135deg, #326ce5, #5b8ff9); }
.icon-green { background: linear-gradient(135deg, #00b42a, #23c343); }
.icon-orange { background: linear-gradient(135deg, #e8740c, #f59e0b); }
.icon-purple { background: linear-gradient(135deg, #722ed1, #9254de); }
.icon-teal { background: linear-gradient(135deg, #0fc6c2, #14c9c9); }
.icon-red { background: linear-gradient(135deg, #f53f3f, #f76560); }
.icon-yellow { background: linear-gradient(135deg, #e8b900, #fadb14); }
.icon-indigo { background: linear-gradient(135deg, #3730a3, #6366f1); }
.icon-dark { background: linear-gradient(135deg, #374151, #6b7280); }

/* ====== 概览网格 ====== */
.overview-grid {
  display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; margin-bottom: 16px;
}
.overview-card {
  background: #fff; border: 1px solid #e5e6eb; border-radius: 10px; padding: 16px;
}
.ov-label { font-size: 12px; color: #86909c; margin-bottom: 8px; text-transform: uppercase; letter-spacing: 0.5px; }
.ov-value {
  font-size: 14px; font-weight: 600; color: #1d2129;
  display: flex; align-items: center; gap: 6px;
}
.ov-value code { font-size: 13px; background: #f2f3f5; padding: 2px 6px; border-radius: 4px; font-family: monospace; font-weight: 500; }
.status-dot {
  width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0;
}
.dot-s1 { background: #326ce5; animation: pulse-fade 1.5s infinite; }
.dot-s2 { background: #00b42a; }
.dot-s3 { background: #f53f3f; }
.dot-s4 { background: #ff7d00; animation: pulse-fade 1.5s infinite; }
.dot-s5 { background: #86909c; }
@keyframes pulse-fade { 0%,100% { opacity:1; } 50% { opacity:0.3; } }

/* ====== 数据库消息 ====== */
.db-message {
  display: flex; align-items: center; gap: 8px; padding: 12px 16px;
  background: #f7f8fa; border: 1px solid #e5e6eb; border-radius: 8px;
  font-size: 13px; color: #4e5969; margin-bottom: 20px;
}
.db-message.error { background: #fff2f0; border-color: #ffc4b8; color: #f53f3f; }

/* ====== K8s 状态区 ====== */
.k8s-status-section { margin-bottom: 24px; }
.section-header {
  display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px;
}
.section-title {
  display: flex; align-items: center; gap: 8px;
  font-size: 16px; font-weight: 600; color: #1d2129; margin: 0;
}
.section-actions { display: flex; align-items: center; gap: 10px; }
.auto-refresh-badge {
  display: inline-flex; align-items: center; gap: 4px; font-size: 12px; color: #00b42a;
  background: #e8f7ea; padding: 2px 10px; border-radius: 10px;
}
.pulse-dot-mini {
  width: 6px; height: 6px; border-radius: 50%; background: #00b42a;
  animation: pulse-fade 1.5s infinite;
}
.cluster-alert { margin-bottom: 16px; }

/* ====== 资源卡片 ====== */
.resource-card {
  background: #fff; border: 1px solid #e5e6eb; border-radius: 12px;
  padding: 20px; margin-bottom: 16px; transition: border-color 0.2s;
}
.resource-card:hover { border-color: #c9cdd4; }
.resource-header {
  display: flex; align-items: center; gap: 12px; margin-bottom: 12px;
}
.resource-icon {
  width: 36px; height: 36px; border-radius: 10px; display: flex; align-items: center;
  justify-content: center; flex-shrink: 0;
}
.deploy-icon { background: #e8f3ff; color: #326ce5; }
.svc-icon { background: #e8f7ea; color: #00b42a; }
.pod-icon { background: #fff3e8; color: #ff7d00; }
.cm-icon { background: #f3e8ff; color: #722ed1; }
.resource-title {
  flex: 1; display: flex; align-items: center; gap: 8px;
  font-size: 15px; font-weight: 600; color: #1d2129;
}
.resource-title code {
  font-size: 13px; font-weight: 400; background: #f2f3f5; padding: 2px 8px; border-radius: 4px; color: #4e5969;
}
.resource-message {
  font-size: 13px; color: #86909c; padding: 8px 12px; background: #f7f8fa;
  border-radius: 6px; margin-bottom: 12px;
}

/* 副本进度条 */
.replica-bar-section { margin-top: 4px; }
.replica-labels {
  display: flex; justify-content: space-between; align-items: center;
  font-size: 12px; color: #86909c; margin-bottom: 6px;
}
.replica-nums { font-weight: 600; color: #1d2129; }
.replica-bar {
  height: 8px; background: #f2f3f5; border-radius: 4px; overflow: hidden;
}
.replica-fill {
  height: 100%; background: linear-gradient(90deg, #326ce5, #00b42a);
  border-radius: 4px; transition: width 0.6s ease;
}
.replica-detail-row {
  display: flex; gap: 16px; margin-top: 10px;
}
.replica-chip {
  display: flex; align-items: center; gap: 4px; font-size: 12px; color: #4e5969;
}
.chip-dot {
  width: 6px; height: 6px; border-radius: 50%;
}

/* Service 详情 */
.svc-detail-grid {
  display: grid; grid-template-columns: 1fr 1fr; gap: 10px; margin-top: 4px;
}
.svc-item { display: flex; flex-direction: column; gap: 4px; }
.svc-label { font-size: 12px; color: #86909c; }
.svc-value { font-size: 13px; color: #1d2129; font-weight: 500; display: flex; gap: 4px; align-items: center; flex-wrap: wrap; }
.svc-value code { font-size: 12px; background: #f2f3f5; padding: 2px 6px; border-radius: 4px; }

/* ====== Pod 列表 ====== */
.pod-count-badge {
  display: inline-flex; align-items: center; justify-content: center;
  min-width: 20px; height: 20px; border-radius: 10px; background: #ff7d00;
  color: #fff; font-size: 11px; font-weight: 600; padding: 0 6px;
}
.no-pods {
  display: flex; flex-direction: column; align-items: center; gap: 8px;
  padding: 32px 0; color: #c0c4cc; font-size: 13px;
}

.pod-item {
  border: 1px solid #e5e6eb; border-radius: 10px; padding: 16px; margin-top: 12px;
  transition: border-color 0.2s;
}
.pod-item:hover { border-color: #c9cdd4; }
.pod-item.phase-running { border-left: 3px solid #00b42a; }
.pod-item.phase-pending { border-left: 3px solid #ff7d00; }
.pod-item.phase-failed { border-left: 3px solid #f53f3f; }
.pod-item.phase-succeeded { border-left: 3px solid #00b42a; }

.pod-header {
  display: flex; align-items: center; gap: 8px; margin-bottom: 10px;
}
.pod-phase-dot {
  width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0;
}
.pod-phase-dot.phase-running { background: #00b42a; }
.pod-phase-dot.phase-pending { background: #ff7d00; animation: pulse-fade 1.5s infinite; }
.pod-phase-dot.phase-failed { background: #f53f3f; }
.pod-phase-dot.phase-succeeded { background: #00b42a; }
.pod-phase-dot.phase-unknown { background: #86909c; }
.pod-name { font-size: 14px; font-weight: 600; color: #1d2129; font-family: monospace; }

.pod-info-grid {
  display: grid; grid-template-columns: repeat(4, 1fr); gap: 8px; margin-bottom: 8px;
}
.pod-info-item { display: flex; flex-direction: column; gap: 2px; }
.pi-label { font-size: 11px; color: #86909c; text-transform: uppercase; }
.pi-value { font-size: 13px; color: #1d2129; font-weight: 500; }
.pi-value code { font-size: 12px; background: #f2f3f5; padding: 1px 4px; border-radius: 3px; }
.restart-warn { color: #ff7d00; font-weight: 600; }

/* 容器 */
.container-list {
  border-top: 1px solid #f2f3f5; padding-top: 10px; margin-top: 10px;
}
.container-list-title {
  font-size: 12px; font-weight: 600; color: #86909c; text-transform: uppercase;
  letter-spacing: 0.5px; margin-bottom: 8px;
}
.container-row {
  padding: 8px 12px; background: #fafbfc; border-radius: 6px; margin-bottom: 6px;
}
.container-header {
  display: flex; align-items: center; gap: 6px; margin-bottom: 4px;
}
.container-ready-dot {
  width: 7px; height: 7px; border-radius: 50%; background: #c0c4cc; flex-shrink: 0;
}
.container-ready-dot.ready { background: #00b42a; }
.container-name { font-size: 13px; font-weight: 600; color: #1d2129; }
.restart-badge {
  font-size: 11px; color: #ff7d00; background: #fff3e8; padding: 1px 6px;
  border-radius: 4px; margin-left: auto;
}
.container-detail {
  font-size: 12px; color: #4e5969; display: flex; align-items: center; gap: 4px; margin-top: 2px;
}
.container-detail.warn { color: #ff7d00; }
.cd-label { color: #86909c; font-weight: 500; }
.cd-image {
  font-size: 11px; background: #f2f3f5; padding: 1px 6px; border-radius: 3px; word-break: break-all;
}
.cd-msg { color: #86909c; }

/* ====== 加载/空状态 ====== */
.status-loading-section, .no-status-section {
  display: flex; flex-direction: column; align-items: center; gap: 12px;
  padding: 48px 0; color: #86909c; font-size: 13px;
}

/* ====== ConfigMap ====== */
.cm-item {
  border: 1px solid #f2f3f5; border-radius: 8px; padding: 12px; margin-top: 10px;
}
.cm-header {
  display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px;
}
.cm-name { font-size: 13px; background: #f3e8ff; padding: 2px 8px; border-radius: 4px; color: #722ed1; }
.cm-time { font-size: 12px; color: #86909c; }
.cm-data-grid { display: flex; flex-direction: column; gap: 4px; }
.cm-data-row {
  display: grid; grid-template-columns: 140px 1fr; gap: 8px; font-size: 12px;
  padding: 4px 8px; border-radius: 4px;
}
.cm-data-row:nth-child(odd) { background: #fafbfc; }
.cm-key { color: #86909c; font-weight: 500; word-break: break-all; }
.cm-val { color: #1d2129; word-break: break-all; }

/* 响应式 */
@media (max-width: 768px) {
  .overview-grid { grid-template-columns: repeat(2, 1fr); }
  .pod-info-grid { grid-template-columns: repeat(2, 1fr); }
  .app-banner { flex-direction: column; gap: 12px; align-items: flex-start; }
  .banner-actions { width: 100%; justify-content: flex-end; }
  .ns-stats-grid { grid-template-columns: repeat(2, 1fr); }
  .ns-table-row, .ns-table-header { font-size: 11px; }
}

/* ====== 命名空间概览 ====== */
.ns-overview-section { margin-bottom: 8px; }
.ns-stats-grid {
  display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; margin-bottom: 16px;
}
.ns-stat-card {
  background: #fff; border: 1px solid #e5e6eb; border-radius: 10px;
  padding: 16px; text-align: center;
}
.ns-stat-num {
  font-size: 24px; font-weight: 700; color: #1d2129; margin-bottom: 4px;
}
.ns-stat-label {
  font-size: 12px; color: #86909c; text-transform: uppercase; letter-spacing: 0.5px;
}

/* ====== 命名空间资源表格 ====== */
.ns-table { margin-top: 8px; }
.ns-table-header, .ns-table-row {
  display: grid;
  grid-template-columns: 2fr 100px 100px 1.5fr 120px;
  gap: 8px; padding: 8px 12px; align-items: center; font-size: 13px;
}
.ns-table-header {
  font-size: 11px; color: #86909c; text-transform: uppercase; letter-spacing: 0.5px;
  border-bottom: 1px solid #f2f3f5; font-weight: 600;
}
.ns-table-row {
  border-bottom: 1px solid #f7f8fa; color: #1d2129; transition: background 0.15s;
}
.ns-table-row:hover { background: #f7f8fa; }
.ns-table-row:last-child { border-bottom: none; }
.row-highlight { background: #f0f5ff !important; }
.ns-col-name {
  display: flex; align-items: center; gap: 6px; font-weight: 500;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.ns-col-name code { font-size: 12px; }
.ns-col-status { }
.ns-col-replicas { font-weight: 600; font-family: monospace; }
.ns-col-image { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.img-code { font-size: 11px; background: #f2f3f5; padding: 1px 6px; border-radius: 3px; }
.ns-col-age { font-size: 12px; color: #86909c; }
.ns-row-dot {
  width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0;
}
.dot-running { background: #00b42a; }
.dot-partialready { background: #ff7d00; }
.dot-pending { background: #ff7d00; animation: pulse-fade 1.5s infinite; }
.dot-failed { background: #f53f3f; }
.dot-scaled0 { background: #86909c; }

/* ====== Events 事件 ====== */
.event-icon { background: #fff3e8; color: #ff7d00; }
.events-list { max-height: 360px; overflow-y: auto; }
.event-row {
  display: flex; align-items: flex-start; gap: 8px; padding: 8px 12px;
  font-size: 12px; border-bottom: 1px solid #f7f8fa; line-height: 1.5;
}
.event-row:last-child { border-bottom: none; }
.event-warning { background: #fffbf0; }
.ev-type-badge {
  flex-shrink: 0; padding: 1px 8px; border-radius: 3px; font-size: 11px; font-weight: 600;
}
.ev-type-badge.normal { background: #e8f7ea; color: #00b42a; }
.ev-type-badge.warning { background: #fff3e8; color: #ff7d00; }
.ev-reason { font-weight: 600; color: #1d2129; flex-shrink: 0; min-width: 80px; }
.ev-object { flex-shrink: 0; }
.ev-object code { font-size: 11px; background: #f2f3f5; padding: 1px 4px; border-radius: 3px; color: #4e5969; }
.ev-message { flex: 1; color: #4e5969; word-break: break-word; }
.ev-count { flex-shrink: 0; font-size: 11px; color: #ff7d00; font-weight: 600; }
.ev-time { flex-shrink: 0; color: #86909c; font-size: 11px; min-width: 80px; text-align: right; }
</style>
