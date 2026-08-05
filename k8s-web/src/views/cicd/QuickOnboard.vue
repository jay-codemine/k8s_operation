<template>
  <div class="quick-onboard-page">
    <!-- 顶部标题栏 -->
    <div class="page-header">
      <div class="header-content">
        <div class="header-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/>
          </svg>
        </div>
        <div class="header-text">
          <h1>快速接入应用</h1>
          <p>一键创建 K8s 工作负载并接入 CI/CD，支持 5 种工作负载类型</p>
        </div>
      </div>
      <div class="header-actions">
        <button class="btn-import" @click="openImportModal">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/>
          </svg>
          从 K8s 导入
        </button>
        <button class="btn-back" @click="$router.push('/cicd/pipelines')">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16">
            <line x1="19" y1="12" x2="5" y2="12"/><polyline points="12 19 5 12 12 5"/>
          </svg>
          返回列表
        </button>
      </div>
    </div>

    <div class="form-container">
      <!-- 工作负载类型选择器 -->
      <div class="workload-type-bar">
        <div class="type-label">选择工作负载类型：</div>
        <div class="type-options">
          <div
            v-for="wt in workloadTypes"
            :key="wt.value"
            :class="['type-chip', { active: form.workload_kind === wt.value }]"
            @click="selectWorkloadType(wt.value)"
          >
            <span class="type-icon" v-html="wt.icon"></span>
            <span class="type-name">{{ wt.label }}</span>
            <span class="type-desc">{{ wt.desc }}</span>
          </div>
        </div>
      </div>

      <form @submit.prevent="submit" class="onboard-form">
        <!-- === 基础信息 === -->
        <div class="form-section">
          <div class="section-title">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18">
              <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
            </svg>
            基础信息
          </div>
          <div class="form-grid cols-3">
            <div class="form-group">
              <label class="form-label">应用名称 <span class="required">*</span></label>
              <input v-model="form.app_name" class="form-input" placeholder="如 my-java-service" required />
            </div>
            <div class="form-group">
              <label class="form-label">工作负载名称</label>
              <input v-model="form.workload_name" class="form-input" :placeholder="form.app_name || '与应用名称相同'" />
            </div>
            <div class="form-group">
              <label class="form-label">容器名称</label>
              <input v-model="form.container_name" class="form-input" :placeholder="form.workload_name || form.app_name || 'app'" />
            </div>
          </div>
          <div class="form-grid cols-3">
            <div class="form-group">
              <label class="form-label">命名空间 <span class="required">*</span></label>
              <input v-model="form.namespace" class="form-input" placeholder="default" required />
            </div>
            <div class="form-group">
              <label class="form-label">目标集群 <span class="required">*</span></label>
              <select v-model.number="form.cluster_id" class="form-input" required>
                <option :value="0" disabled>请选择集群</option>
                <option v-for="c in clusters" :key="c.id" :value="c.id">{{ c.cluster_name || c.name || '集群-' + c.id }}</option>
              </select>
            </div>
            <div class="form-group">
              <label class="form-label">镜像地址 <span class="required">*</span></label>
              <input v-model="form.image" class="form-input" placeholder="registry.example.com/my-app:v1.0" required />
            </div>
          </div>
        </div>

        <!-- === 副本数（Deployment / StatefulSet） === -->
        <div class="form-section" v-if="['Deployment','StatefulSet'].includes(form.workload_kind)">
          <div class="section-title">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18">
              <rect x="2" y="2" width="20" height="20" rx="2"/><line x1="8" y1="2" x2="8" y2="22"/><line x1="16" y1="2" x2="16" y2="22"/>
            </svg>
            副本配置
          </div>
          <div class="form-grid cols-3">
            <div class="form-group">
              <label class="form-label">副本数</label>
              <input v-model.number="form.replicas" type="number" min="1" max="100" class="form-input" />
            </div>
          </div>
        </div>

        <!-- === CronJob 专属 === -->
        <div class="form-section" v-if="form.workload_kind === 'CronJob'">
          <div class="section-title">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18">
              <circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/>
            </svg>
            CronJob 配置
          </div>
          <div class="form-grid cols-3">
            <div class="form-group">
              <label class="form-label">调度表达式 <span class="required">*</span></label>
              <input v-model="form.cron_schedule" class="form-input" placeholder="*/5 * * * *" required />
              <div class="input-hint">Cron 表达式：分 时 日 月 周</div>
            </div>
            <div class="form-group">
              <label class="form-label">并发策略</label>
              <select v-model="form.cron_concurrency_policy" class="form-input">
                <option value="Allow">Allow - 允许并发</option>
                <option value="Forbid">Forbid - 禁止并发</option>
                <option value="Replace">Replace - 替换旧Job</option>
              </select>
            </div>
          </div>
        </div>

        <!-- === Job 专属 === -->
        <div class="form-section" v-if="form.workload_kind === 'Job'">
          <div class="section-title">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18">
              <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/>
            </svg>
            Job 配置
          </div>
          <div class="form-grid cols-4">
            <div class="form-group">
              <label class="form-label">完成次数</label>
              <input v-model.number="form.job_completions" type="number" min="1" class="form-input" placeholder="1" />
            </div>
            <div class="form-group">
              <label class="form-label">并行度</label>
              <input v-model.number="form.job_parallelism" type="number" min="1" class="form-input" placeholder="1" />
            </div>
            <div class="form-group">
              <label class="form-label">重试次数</label>
              <input v-model.number="form.job_backoff_limit" type="number" min="0" class="form-input" placeholder="6" />
            </div>
            <div class="form-group">
              <label class="form-label">完成后保留(秒)</label>
              <input v-model.number="form.job_ttl" type="number" min="0" class="form-input" placeholder="3600" />
            </div>
          </div>
        </div>

        <!-- === 端口配置 === -->
        <div class="form-section">
          <div class="section-title">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18">
              <polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/>
            </svg>
            端口配置
          </div>
          <div class="ports-editor">
            <div v-for="(port, idx) in form.ports" :key="idx" class="port-row">
              <input v-model="port.name" class="form-input port-name" placeholder="http" />
              <input v-model.number="port.port" type="number" class="form-input port-num" placeholder="8080" />
              <select v-model="port.protocol" class="form-input port-proto">
                <option value="TCP">TCP</option>
                <option value="UDP">UDP</option>
              </select>
              <button type="button" class="btn-icon-sm" @click="form.ports.splice(idx, 1)" title="删除">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14">
                  <polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6"/>
                </svg>
              </button>
            </div>
            <button type="button" class="btn-add-port" @click="form.ports.push({name:'http',port:8080,protocol:'TCP'})">
              + 添加端口
            </button>
          </div>
        </div>

        <!-- === 资源配置 === -->
        <div class="form-section">
          <div class="section-title">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18">
              <rect x="2" y="3" width="20" height="14" rx="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/>
            </svg>
            资源配置
          </div>
          <div class="form-grid cols-4">
            <div class="form-group">
              <label class="form-label">CPU Request</label>
              <input v-model="form.cpu_req" class="form-input" placeholder="100m" />
            </div>
            <div class="form-group">
              <label class="form-label">CPU Limit</label>
              <input v-model="form.cpu_lim" class="form-input" placeholder="500m" />
            </div>
            <div class="form-group">
              <label class="form-label">Memory Request</label>
              <input v-model="form.mem_req" class="form-input" placeholder="128Mi" />
            </div>
            <div class="form-group">
              <label class="form-label">Memory Limit</label>
              <input v-model="form.mem_lim" class="form-input" placeholder="256Mi" />
            </div>
          </div>
        </div>

        <!-- === 环境变量 === -->
        <div class="form-section">
          <div class="section-title">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18">
              <path d="M4 7h16M4 12h16M4 17h10"/>
            </svg>
            环境变量
          </div>
          <div class="env-editor">
            <div v-for="(ev, idx) in form.env_vars" :key="idx" class="env-row">
              <input v-model="ev.name" class="form-input" placeholder="变量名" />
              <span class="env-eq">=</span>
              <input v-model="ev.value" class="form-input" placeholder="变量值" />
              <button type="button" class="btn-icon-sm" @click="form.env_vars.splice(idx, 1)" title="删除">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14">
                  <polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6"/>
                </svg>
              </button>
            </div>
            <button type="button" class="btn-add" @click="form.env_vars.push({name:'',value:''})">
              + 添加环境变量
            </button>
          </div>
        </div>

        <!-- === ConfigMap === -->
        <div class="form-section">
          <div class="section-title">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18"><rect x="2" y="2" width="20" height="20" rx="2"/><line x1="8" y1="2" x2="8" y2="22"/><line x1="16" y1="2" x2="16" y2="22"/></svg>
            ConfigMap（配置文件挂载）
          </div>
          <div class="form-grid cols-2">
            <div class="form-group">
              <label class="form-label">挂载路径</label>
              <input v-model="form.configmap_mount_path" class="form-input" placeholder="/etc/config" />
            </div>
          </div>
          <div class="env-editor">
            <div v-for="(kv, idx) in form.configmap_data" :key="idx" class="env-row">
              <input v-model="kv.key" class="form-input" placeholder="文件名/键" />
              <span class="env-eq">=</span>
              <input v-model="kv.value" class="form-input" placeholder="内容" />
              <button type="button" class="btn-icon-sm" @click="form.configmap_data.splice(idx,1)">✕</button>
            </div>
            <button type="button" class="btn-add" @click="form.configmap_data.push({key:'',value:''})">+ 添加配置项</button>
          </div>
        </div>

        <!-- === Secret === -->
        <div class="form-section">
          <div class="section-title">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18"><rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
            Secret（敏感信息）
          </div>
          <div class="form-grid cols-2">
            <div class="form-group">
              <label class="form-label">挂载路径</label>
              <input v-model="form.secret_mount_path" class="form-input" placeholder="/etc/secrets" />
            </div>
          </div>
          <div class="env-editor">
            <div v-for="(kv, idx) in form.secret_data" :key="idx" class="env-row">
              <input v-model="kv.key" class="form-input" placeholder="键名" />
              <span class="env-eq">=</span>
              <input v-model="kv.value" class="form-input" type="password" placeholder="值" />
              <button type="button" class="btn-icon-sm" @click="form.secret_data.splice(idx,1)">✕</button>
            </div>
            <button type="button" class="btn-add" @click="form.secret_data.push({key:'',value:''})">+ 添加 Secret</button>
          </div>
        </div>

        <!-- === PVC 存储 === -->
        <div class="form-section">
          <div class="section-title">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg>
            持久化存储 (PVC)
          </div>
          <div class="form-grid cols-4">
            <div class="form-group">
              <label class="form-label">PVC 名称</label>
              <input v-model="form.pvc_name" class="form-input" :placeholder="(form.workload_name||form.app_name||'app')+'-data'" />
            </div>
            <div class="form-group">
              <label class="form-label">容量</label>
              <input v-model="form.pvc_size" class="form-input" placeholder="10Gi" />
            </div>
            <div class="form-group">
              <label class="form-label">StorageClass</label>
              <input v-model="form.pvc_storage_class" class="form-input" placeholder="默认" />
            </div>
            <div class="form-group">
              <label class="form-label">访问模式</label>
              <select v-model="form.pvc_access_mode" class="form-input">
                <option value="">ReadWriteOnce</option>
                <option value="ReadWriteOnce">ReadWriteOnce</option>
                <option value="ReadWriteMany">ReadWriteMany</option>
                <option value="ReadOnlyMany">ReadOnlyMany</option>
              </select>
            </div>
          </div>
          <div class="form-grid cols-2" style="margin-top:12px;">
            <div class="form-group">
              <label class="form-label">挂载路径</label>
              <input v-model="form.pvc_mount_path" class="form-input" placeholder="/data" />
            </div>
          </div>
        </div>

        <!-- === Service 配置 === -->
        <div class="form-section">
          <div class="section-title">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18">
              <circle cx="12" cy="12" r="10"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10A15.3 15.3 0 0 1 12 2z"/>
            </svg>
            Service 配置
          </div>
          <div class="form-grid cols-2">
            <div class="form-group">
              <label class="form-label">Service 类型</label>
              <select v-model="form.service_type" class="form-input">
                <option value="">不创建 Service</option>
                <option value="ClusterIP">ClusterIP</option>
                <option value="NodePort">NodePort</option>
                <option value="LoadBalancer">LoadBalancer</option>
              </select>
            </div>
          </div>
          <!-- 服务端口映射（仅在选择 Service 类型时显示） -->
          <div v-if="form.service_type" class="service-ports-editor" style="margin-top:14px;">
            <div class="service-ports-header">
              <span class="sp-label">服务端口映射</span>
              <span class="sp-hint">不填则使用上方"端口配置"作为服务端口</span>
            </div>
            <div v-for="(sp, idx) in form.service_ports" :key="idx" class="sp-row">
              <input v-model="sp.name" class="form-input sp-name" placeholder="端口名（如 http）" />
              <div class="sp-field">
                <label class="sp-field-label">服务端口</label>
                <input v-model.number="sp.port" type="number" class="form-input" placeholder="80" />
              </div>
              <div class="sp-field">
                <label class="sp-field-label">目标端口</label>
                <input v-model.number="sp.target_port" type="number" class="form-input" placeholder="容器端口" />
              </div>
              <div class="sp-field" v-if="form.service_type === 'NodePort'">
                <label class="sp-field-label">NodePort</label>
                <input v-model.number="sp.node_port" type="number" class="form-input" min="30000" max="32767" placeholder="30000-32767" />
              </div>
              <select v-model="sp.protocol" class="form-input sp-proto">
                <option value="TCP">TCP</option>
                <option value="UDP">UDP</option>
              </select>
              <button type="button" class="btn-icon-sm" @click="form.service_ports.splice(idx,1)" title="删除">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14">
                  <polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6"/>
                </svg>
              </button>
            </div>
            <button type="button" class="btn-add-port" @click="form.service_ports.push({name:'',port:80,target_port:0,node_port:0,protocol:'TCP'})">
              + 添加服务端口
            </button>
          </div>
        </div>

        <!-- === CI/CD === -->
        <div class="form-section">
          <div class="section-title">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18">
              <path d="M12 2L2 7l10 5 10-5-10-5z"/><path d="M2 17l10 5 10-5"/><path d="M2 12l10 5 10-5"/>
            </svg>
            CI/CD 接入（可选）
          </div>
          <div class="form-grid cols-2">
            <div class="form-group">
              <label class="form-label">Git 仓库地址</label>
              <input v-model="form.git_repo" class="form-input" placeholder="https://gitlab.com/team/my-app.git" />
              <div class="input-hint">填写后将自动创建 CI/CD 流水线</div>
            </div>
            <div class="form-group">
              <label class="form-label">Git 分支</label>
              <input v-model="form.git_branch" class="form-input" placeholder="main（留空使用默认分支）" />
            </div>
          </div>
          <div class="toggle-row" style="margin-top:16px;">
            <div class="toggle-info">
              <label class="form-label" style="margin:0;">创建后立即部署</label>
              <p class="toggle-desc">K8s 资源创建完成后自动触发一次测试发布（若目标环境配置了审批，将进入待审批状态）</p>
            </div>
            <label class="toggle-switch">
              <input type="checkbox" v-model="form.auto_deploy" />
              <span class="toggle-slider"></span>
            </label>
          </div>
        </div>

        <!-- === 提交按钮 === -->
        <div class="form-actions">
          <button type="button" class="btn-cancel" @click="$router.push('/cicd/pipelines')">取消</button>
          <button type="submit" class="btn-submit" :disabled="submitting">
            <svg v-if="!submitting" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16">
              <polyline points="20 6 9 17 4 12"/>
            </svg>
            <span v-if="submitting" class="spinner-sm"></span>
            {{ submitting ? '接入中...' : '一键接入' }}
          </button>
        </div>
      </form>
    </div>

    <!-- 从 K8s 导入弹窗 -->
    <transition name="fade">
      <div v-if="showImport" class="result-overlay" @click.self="showImport = false">
        <div class="result-card" style="width:760px;text-align:left;">
          <div class="result-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="#667eea" stroke-width="2" width="36" height="36"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
          </div>
          <h2 style="margin:0 0 16px;font-size:17px;">从 K8s 集群导入应用</h2>
          <div style="display:flex;gap:12px;margin-bottom:16px;">
            <select v-model.number="importClusterId" class="form-input" style="flex:1">
              <option :value="0" disabled>选择集群</option>
              <option v-for="c in clusters" :key="c.id" :value="c.id">{{ c.cluster_name || c.name || '集群-'+c.id }}</option>
            </select>
            <input v-model="importNamespace" class="form-input" style="flex:1" placeholder="命名空间，如 default" />
            <button class="btn-submit" @click="discoverK8sWorkloads" :disabled="importLoading" style="white-space:nowrap;padding:9px 18px;">
              <span v-if="importLoading" class="spinner-sm"></span>
              {{ importLoading ? '扫描中...' : '扫描应用' }}
            </button>
          </div>
          <div v-if="discoveredApps.length > 0">
            <input v-model="importSearch" class="form-input" style="width:100%;margin-bottom:12px;" placeholder="🔍 搜索应用名 / 命名空间 / 工作负载名..." />
            <div style="max-height:400px;overflow-y:auto;">
            <div v-for="(app, ai) in filteredDiscoveredApps" :key="discoveredApps.indexOf(app)" style="border:1px solid #e8eaf0;border-radius:10px;padding:14px;margin-bottom:10px;" :style="{background: importSelected.includes(discoveredApps.indexOf(app))?'#f0f0ff':''}">
              <div style="display:flex;align-items:center;gap:10px;margin-bottom:8px;">
                <input type="checkbox" :checked="importSelected.includes(discoveredApps.indexOf(app))" @change="toggleSelect(discoveredApps.indexOf(app))" />
                <strong style="font-size:14px;">{{ app.app_name }}</strong>
                <span style="font-size:11px;color:#9ca3af;">{{ app.namespace }}</span>
                <span style="flex:1;"></span>
                <span v-for="wl in app.workloads" :key="wl.name" :class="'type-chip-sm type-'+wl.kind.toLowerCase()" style="margin-left:4px;">{{ wl.kind }}</span>
              </div>
              <!-- 工作负载 -->
              <div v-if="app.workloads.length" style="font-size:12px;color:#6b7280;margin-left:28px;">
                <span v-for="(wl,wi) in app.workloads" :key="wi">
                  <strong>{{ wl.kind }}</strong> {{ wl.name }} ({{ wl.image.split('/').pop() }}){{ wi < app.workloads.length-1 ? '、' : '' }}
                </span>
              </div>
              <!-- 关联资源 badges -->
              <div style="margin-left:28px;margin-top:4px;display:flex;gap:6px;flex-wrap:wrap;">
                <span v-if="app.configmaps.length" class="res-badge res-cm" :title="app.configmaps.map(c=>c.name+'['+c.keys.join(',')+']').join('\n')">📋 ConfigMap ×{{ app.configmaps.length }}</span>
                <span v-if="app.secrets.length" class="res-badge res-sec" :title="app.secrets.map(s=>s.name).join('\n')">🔒 Secret ×{{ app.secrets.length }}</span>
                <span v-if="app.services.length" class="res-badge res-svc" :title="app.services.map(s=>s.name+':'+s.type).join('\n')">🌐 Service ×{{ app.services.length }}</span>
                <span v-if="app.pvcs.length" class="res-badge res-pvc" :title="app.pvcs.map(p=>p.name+' '+p.size).join('\n')">💾 PVC ×{{ app.pvcs.length }}</span>
              </div>
            </div>
            </div>
          </div>
          <div v-if="discoveredApps.length === 0 && importSearched" style="text-align:center;padding:24px;color:#9ca3af;">
            该命名空间下未发现应用
          </div>
          <div class="result-actions" style="margin-top:16px;">
            <button class="btn-cancel" @click="showImport = false">取消</button>
            <button class="btn-submit" @click="batchOnboard" :disabled="importSelected.length === 0 || batchLoading">
              <span v-if="batchLoading" class="spinner-sm"></span>
              {{ batchLoading ? '接入中...' : `批量接入 (${importSelected.length})` }}
            </button>
          </div>
        </div>
      </div>
    </transition>

    <!-- 结果弹窗 -->
    <transition name="fade">
      <div v-if="showResult" class="result-overlay" @click.self="showResult = false">
        <div class="result-card" :class="{ success: result.success, error: !result.success }">
          <div class="result-icon">
            <svg v-if="result.success" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="48" height="48">
              <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/>
            </svg>
            <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="48" height="48">
              <circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/>
            </svg>
          </div>
          <h2 class="result-title">{{ result.success ? '接入成功！' : '接入失败' }}</h2>
          <p class="result-msg">{{ result.message }}</p>
          <div v-if="result.success" class="result-details">
            <div class="detail-row">
              <span class="detail-label">工作负载类型</span>
              <span class="detail-value">{{ result.workload_kind }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">工作负载名称</span>
              <span class="detail-value">{{ result.workload_name }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">命名空间</span>
              <span class="detail-value">{{ result.namespace }}</span>
            </div>
            <div v-if="result.service_name" class="detail-row">
              <span class="detail-label">Service</span>
              <span class="detail-value">{{ result.service_name }}</span>
            </div>
            <div v-if="result.pipeline_id" class="detail-row">
              <span class="detail-label">流水线 ID</span>
              <span class="detail-value">{{ result.pipeline_id }}</span>
            </div>
            <div v-if="result.release_id" class="detail-row">
              <span class="detail-label">发布单 ID</span>
              <span class="detail-value">{{ result.release_id }}</span>
            </div>
          </div>
          <div class="result-actions">
            <button class="btn-primary" @click="showResult = false">关闭</button>
            <button v-if="result.success && result.pipeline_id" class="btn-secondary" @click="goToPipeline">查看流水线</button>
            <button v-if="result.success" class="btn-secondary" @click="resetForm">继续接入</button>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script>
import { quickOnboard, discoverWorkloads, discoverApps } from '@/api/cicd'
import { getClusterList } from '@/api/cluster'

export default {
  name: 'QuickOnboard',

  data() {
    return {
      submitting: false,
      showResult: false,
      result: { success: false, message: '' },
      // 导入弹窗
      showImport: false,
      importClusterId: 0,
      importNamespace: 'default',
      importLoading: false,
      importSearched: false,
      importSelected: [],
      batchLoading: false,
      importSearch: '',
      discoveredApps: [],
      discoveredWorkloads: [],
      clusters: [],

      workloadTypes: [
        { value: 'Deployment', label: 'Deployment', desc: '无状态服务', icon: '<rect x="2" y="2" width="20" height="20" rx="2"/><line x1="8" y1="2" x2="8" y2="22"/><line x1="16" y1="2" x2="16" y2="22"/>' },
        { value: 'StatefulSet', label: 'StatefulSet', desc: '有状态服务', icon: '<rect x="2" y="2" width="20" height="20" rx="2"/><line x1="8" y1="2" x2="8" y2="22"/><line x1="16" y1="2" x2="16" y2="22"/><circle cx="12" cy="12" r="3"/>' },
        { value: 'DaemonSet', label: 'DaemonSet', desc: '每节点一个', icon: '<rect x="2" y="2" width="20" height="20" rx="2"/><line x1="8" y1="8" x2="8" y2="16"/><line x1="12" y1="8" x2="12" y2="16"/><line x1="16" y1="8" x2="16" y2="16"/>' },
        { value: 'CronJob', label: 'CronJob', desc: '定时任务', icon: '<circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/>' },
        { value: 'Job', label: 'Job', desc: '一次性任务', icon: '<path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/>' },
      ],

      form: {
        app_name: '',
        workload_name: '',
        container_name: '',
        namespace: 'default',
        workload_kind: 'Deployment',
        image: '',
        replicas: 1,
        ports: [],
        cpu_req: '100m',
        cpu_lim: '500m',
        mem_req: '128Mi',
        mem_lim: '256Mi',
        env_vars: [],
        cron_schedule: '',
        cron_concurrency_policy: 'Allow',
        job_completions: null,
        job_parallelism: null,
        job_backoff_limit: null,
        job_ttl: null,
        service_type: '',
        service_ports: [],
        cluster_id: 0,
        auto_deploy: false,
        git_repo: '',
        git_branch: '',
        configmap_data: [],
        configmap_mount_path: '',
        secret_data: [],
        secret_mount_path: '',
        pvc_name: '',
        pvc_size: '',
        pvc_storage_class: '',
        pvc_access_mode: '',
        pvc_mount_path: '',
      },
    }
  },

  computed: {
    filteredDiscoveredApps() {
      const q = (this.importSearch || '').toLowerCase().trim()
      if (!q) return this.discoveredApps
      return this.discoveredApps.filter(app => {
        const name = (app.app_name || '').toLowerCase()
        const ns = (app.namespace || '').toLowerCase()
        const wlNames = (app.workloads || []).map(w => (w.name || '').toLowerCase()).join(' ')
        return name.includes(q) || ns.includes(q) || wlNames.includes(q)
      })
    }
  },

  async created() {
    this.fetchClusters()
  },

  methods: {
    async fetchClusters() {
      try {
        const res = await getClusterList({ page: 1, page_size: 100 })
        // res = { code: 0, msg: "OK", data: { list: [...], total: N } }
        this.clusters = res?.data?.list || []
      } catch (e) {
        console.error('获取集群列表失败', e)
      }
    },

    selectWorkloadType(type) {
      this.form.workload_kind = type
      // 清理不相关字段
      if (type !== 'CronJob') {
        this.form.cron_schedule = ''
        this.form.cron_concurrency_policy = 'Allow'
      }
      if (type !== 'Job') {
        this.form.job_completions = null
        this.form.job_parallelism = null
        this.form.job_backoff_limit = null
        this.form.job_ttl = null
      }
    },

    async submit() {
      if (!this.form.app_name || !this.form.image || !this.form.cluster_id) {
        this.$message?.error?.({ content: '请填写必填项：应用名称、镜像地址、目标集群' }) ||
          alert('请填写必填项：应用名称、镜像地址、目标集群')
        return
      }

      if (this.form.workload_kind === 'CronJob' && !this.form.cron_schedule) {
        this.$message?.error?.({ content: 'CronJob 必须设置调度表达式' }) ||
          alert('CronJob 必须设置调度表达式')
        return
      }

      this.submitting = true
      try {
        const payload = this.buildPayload()
        const res = await quickOnboard(payload)
        // 兼容后端业务错误码（HTTP 200 但 code 非 0）
        if (res?.code !== undefined && res.code !== 0) {
          throw new Error(res?.msg || '接入失败')
        }
        this.result = {
          success: true,
          ...(res?.data || res),
        }
      } catch (e) {
        this.result = {
          success: false,
          message: e?.message || e?.msg || '接入失败，请检查参数后重试',
        }
      } finally {
        this.submitting = false
        this.showResult = true
      }
    },

    buildPayload() {
      const p = { ...this.form }

      // 清理空字符串
      if (!p.workload_name) delete p.workload_name
      if (!p.container_name) delete p.container_name
      if (!p.git_repo) delete p.git_repo
      if (!p.git_branch) delete p.git_branch
      if (!p.service_type) { p.service_type = '' }
      if (!p.pvc_name) delete p.pvc_name
      if (!p.pvc_size) delete p.pvc_size
      if (!p.pvc_storage_class) delete p.pvc_storage_class
      if (!p.pvc_mount_path) delete p.pvc_mount_path
      // 清理空 ConfigMap/Secret 条目
      p.configmap_data = (p.configmap_data || []).filter(kv => kv.key)
      p.secret_data = (p.secret_data || []).filter(kv => kv.key)
      if (!p.configmap_data.length) delete p.configmap_data
      if (!p.secret_data.length) delete p.secret_data

      // 清理不适用字段
      if (!['Deployment', 'StatefulSet'].includes(p.workload_kind)) {
        delete p.replicas
      }
      if (p.workload_kind !== 'CronJob') {
        delete p.cron_schedule
        delete p.cron_concurrency_policy
      }
      if (p.workload_kind !== 'Job') {
        delete p.job_completions
        delete p.job_parallelism
        delete p.job_backoff_limit
        delete p.job_ttl
      } else {
        // 转换 Job TTL 字段名
        if (p.job_ttl) {
          p.job_ttl_seconds_after_finished = p.job_ttl
        }
        delete p.job_ttl
      }

      // 清理空端口（只要求端口号，名字缺失时后端自动生成）
      p.ports = (p.ports || []).filter(po => po.port)
      // 服务端口：仅在选择 Service 类型且有端口时发送
      p.service_ports = (p.service_ports || []).filter(sp => sp.port)
      if (!p.service_ports.length) {
        delete p.service_ports
      } else {
        // node_port=0 时后端自动分配，target_port=0 时回退为 port
        p.service_ports = p.service_ports.map(sp => ({
          name: sp.name, port: sp.port,
          target_port: sp.target_port || 0,
          node_port: sp.node_port || 0,
          protocol: sp.protocol || 'TCP',
        }))
      }
      // 清理空环境变量
      p.env_vars = (p.env_vars || []).filter(ev => ev.name)

      return p
    },

    resetForm() {
      this.showResult = false
      this.form = {
        app_name: '',
        workload_name: '',
        container_name: '',
        namespace: 'default',
        workload_kind: 'Deployment',
        image: '',
        replicas: 1,
        ports: [],
        cpu_req: '100m',
        cpu_lim: '500m',
        mem_req: '128Mi',
        mem_lim: '256Mi',
        env_vars: [],
        cron_schedule: '',
        cron_concurrency_policy: 'Allow',
        job_completions: null,
        job_parallelism: null,
        job_backoff_limit: null,
        job_ttl: null,
        service_type: '',
        service_ports: [],
        cluster_id: 0,
        auto_deploy: false,
        git_repo: '',
        git_branch: '',
        configmap_data: [],
        configmap_mount_path: '',
        secret_data: [],
        secret_mount_path: '',
        pvc_name: '',
        pvc_size: '',
        pvc_storage_class: '',
        pvc_access_mode: '',
        pvc_mount_path: '',
      }
    },

    // ========== K8s 导入 ==========
    openImportModal() {
      this.showImport = true
      this.importSelected = []
      this.discoveredApps = []
      this.importSearched = false
      if (this.clusters.length > 0 && !this.importClusterId) {
        this.importClusterId = this.form.cluster_id || this.clusters[0].id
      }
    },

    async discoverK8sWorkloads() {
      if (!this.importClusterId || !this.importNamespace) {
        alert('请选择集群和命名空间')
        return
      }
      this.importLoading = true
      this.importSearched = true
      try {
        const res = await discoverApps({ cluster_id: this.importClusterId, namespace: this.importNamespace })
        const data = res?.data || res || {}
        this.discoveredApps = data.apps || []
        this.importSelected = []
      } catch (e) {
        alert('扫描失败: ' + (e?.message || e?.msg || '集群连接异常'))
        this.discoveredApps = []
      } finally {
        this.importLoading = false
      }
    },

    toggleSelectAll(e) {
      if (e.target.checked) {
        this.importSelected = this.discoveredApps.map((_, i) => i)
      } else {
        this.importSelected = []
      }
    },

    toggleSelect(i) {
      const idx = this.importSelected.indexOf(i)
      if (idx >= 0) this.importSelected.splice(idx, 1)
      else this.importSelected.push(i)
    },

    async batchOnboard() {
      if (this.importSelected.length === 0) return
      this.batchLoading = true
      let success = 0; let fail = 0
      for (const i of this.importSelected) {
        const app = this.discoveredApps[i]
        if (!app.workloads.length) continue
        const wl = app.workloads[0]
        try {
          // 构建 payload：工作负载 + 关联资源
          const payload = {
            app_name: app.app_name,
            workload_name: wl.name,
            container_name: wl.container_name,
            namespace: this.importNamespace,
            workload_kind: wl.kind,
            image: wl.image,
            replicas: wl.replicas || 1,
            cron_schedule: wl.schedule || undefined,
            cluster_id: this.importClusterId,
            service_type: app.services.length > 0 ? (app.services[0].type || 'ClusterIP') : '',
            auto_deploy: false,
          }
          // 已有 ConfigMap/Secret/PVC 保留在集群中，无需重复创建（后端对已存在资源自动跳过）
          const res = await quickOnboard(payload)
          if (res?.code !== undefined && res.code !== 0) {
            throw new Error(res?.msg || '接入失败')
          }
          success++
        } catch (e) {
          fail++
          console.error(`接入失败: ${app.app_name}`, e)
        }
      }
      this.batchLoading = false
      this.showImport = false
      this.result = {
        success: fail === 0,
        message: `批量接入完成: 成功 ${success} / 共 ${success + fail}`,
        workload_kind: '批量',
        workload_name: `${success} 个应用`,
        namespace: this.importNamespace,
      }
      this.showResult = true
    },

    goToPipeline() {
      this.showResult = false
      if (this.result.pipeline_id) {
        this.$router.push(`/cicd/pipelines/${this.result.pipeline_id}?tab=stages`)
      } else {
        this.$router.push('/cicd/pipelines')
      }
    },
  },
}
</script>

<style scoped>
.quick-onboard-page {
  max-width: 960px;
  margin: 0 auto;
  padding: 20px 24px 40px;
}

/* 顶部标题栏 */
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
}
.header-content {
  display: flex;
  align-items: center;
  gap: 14px;
}
.header-icon {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
}
.header-icon svg {
  width: 22px;
  height: 22px;
}
.header-text h1 {
  font-size: 20px;
  font-weight: 700;
  color: #1f2937;
  margin: 0;
}
.header-text p {
  font-size: 13px;
  color: #6b7280;
  margin: 4px 0 0;
}
.btn-import {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border: 1px solid #667eea;
  border-radius: 8px;
  background: linear-gradient(135deg, #667eea10 0%, #764ba210 100%);
  color: #667eea;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  margin-right: 8px;
}
.btn-import:hover {
  background: linear-gradient(135deg, #667eea20 0%, #764ba220 100%);
}
.btn-back {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  background: #fff;
  color: #374151;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-back:hover {
  background: #f3f4f6;
}

/* 工作负载类型小标签 */
.type-chip-sm {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
}
.type-chip-sm.type-deployment { background: #dbeafe; color: #1d4ed8; }
.type-chip-sm.type-statefulset { background: #fef3c7; color: #b45309; }
.type-chip-sm.type-daemonset { background: #ede9fe; color: #6d28d9; }
.type-chip-sm.type-cronjob { background: #d1fae5; color: #065f46; }
.type-chip-sm.type-job { background: #fce7f3; color: #9d174d; }

.res-badge {
  display: inline-block;
  padding: 1px 8px;
  border-radius: 10px;
  font-size: 11px;
  cursor: default;
}
.res-badge.res-cm { background: #dbeafe; color: #1d4ed8; }
.res-badge.res-sec { background: #fee2e2; color: #991b1b; }
.res-badge.res-svc { background: #d1fae5; color: #065f46; }
.res-badge.res-pvc { background: #fef3c7; color: #b45309; }

/* 工作负载类型选择器 */
.workload-type-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 18px;
  background: #f8f9fb;
  border: 1px solid #e8eaf0;
  border-radius: 12px;
  margin-bottom: 24px;
  flex-wrap: wrap;
}
.type-label {
  font-size: 13px;
  font-weight: 600;
  color: #6b7280;
  white-space: nowrap;
}
.type-options {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.type-chip {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border: 2px solid #e5e7eb;
  border-radius: 10px;
  background: #fff;
  cursor: pointer;
  transition: all 0.2s;
  user-select: none;
}
.type-chip:hover {
  border-color: #667eea;
  background: #f0f0ff;
}
.type-chip.active {
  border-color: #667eea;
  background: linear-gradient(135deg, #667eea10 0%, #764ba210 100%);
  box-shadow: 0 0 0 2px rgba(102,126,234,0.15);
}
.type-icon {
  width: 18px;
  height: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #667eea;
}
.type-icon svg {
  width: 16px;
  height: 16px;
}
.type-name {
  font-size: 13px;
  font-weight: 600;
  color: #1f2937;
}
.type-desc {
  font-size: 11px;
  color: #9ca3af;
}

/* 表单容器 */
.form-container {
  background: #fff;
  border: 1px solid #e8eaf0;
  border-radius: 16px;
  padding: 24px;
}
.onboard-form {
  display: flex;
  flex-direction: column;
  gap: 0;
}

/* 表单区块 */
.form-section {
  padding: 20px 0;
  border-bottom: 1px solid #f0f1f3;
}
.form-section:last-child {
  border-bottom: none;
}
.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 16px;
}
.section-title svg {
  color: #667eea;
}

/* 表单网格 */
.form-grid {
  display: grid;
  gap: 16px;
}
.form-grid.cols-2 { grid-template-columns: repeat(2, 1fr); }
.form-grid.cols-3 { grid-template-columns: repeat(3, 1fr); }
.form-grid.cols-4 { grid-template-columns: repeat(4, 1fr); }

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.form-label {
  font-size: 13px;
  font-weight: 500;
  color: #374151;
}
.required {
  color: #ef4444;
}

.form-input {
  padding: 9px 12px;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 13px;
  color: #1f2937;
  background: #fff;
  transition: border-color 0.2s;
  width: 100%;
  box-sizing: border-box;
}
.form-input:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102,126,234,0.1);
}
.input-hint {
  font-size: 11px;
  color: #9ca3af;
  margin-top: 2px;
}

/* 端口编辑器 */
.ports-editor {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.port-row {
  display: flex;
  gap: 8px;
  align-items: center;
}
.port-name { width: 100px; flex-shrink: 0; }
.port-num { width: 90px; flex-shrink: 0; }
.port-proto { width: 100px; flex-shrink: 0; }
.btn-icon-sm {
  width: 30px;
  height: 30px;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  background: #fff;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #9ca3af;
  transition: all 0.2s;
  flex-shrink: 0;
}
.btn-icon-sm:hover {
  color: #ef4444;
  border-color: #fca5a5;
  background: #fef2f2;
}
.btn-add-port, .btn-add {
  padding: 6px 12px;
  border: 1px dashed #d1d5db;
  border-radius: 8px;
  background: #fafbfc;
  color: #667eea;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
  margin-top: 4px;
  width: fit-content;
}
.btn-add-port:hover, .btn-add:hover {
  border-color: #667eea;
  background: #f0f0ff;
}

/* 环境变量编辑器 */
.env-editor {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.env-row {
  display: flex;
  gap: 8px;
  align-items: center;
}
.env-eq {
  color: #9ca3af;
  font-size: 14px;
  font-weight: 600;
  flex-shrink: 0;
}

/* 服务端口编辑器 */
.service-ports-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}
.sp-label {
  font-size: 13px;
  font-weight: 500;
  color: #374151;
}
.sp-hint {
  font-size: 11px;
  color: #9ca3af;
}
.sp-row {
  display: flex;
  gap: 6px;
  align-items: flex-end;
  margin-bottom: 8px;
  flex-wrap: wrap;
}
.sp-name { width: 100px; flex-shrink: 0; }
.sp-proto { width: 90px; flex-shrink: 0; }
.sp-field {
  display: flex;
  flex-direction: column;
  gap: 3px;
  width: 100px;
  flex-shrink: 0;
}
.sp-field-label {
  font-size: 10px;
  color: #9ca3af;
  font-weight: 500;
}

/* Toggle */
.toggle-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: #fafbfc;
  border: 1px solid #e8eaf0;
  border-radius: 10px;
}
.toggle-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.toggle-desc {
  font-size: 12px;
  color: #9ca3af;
  margin: 0;
}
.toggle-switch {
  position: relative;
  width: 44px;
  height: 24px;
  cursor: pointer;
  flex-shrink: 0;
}
.toggle-switch input {
  opacity: 0;
  width: 0;
  height: 0;
}
.toggle-slider {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: #d1d5db;
  border-radius: 12px;
  transition: 0.3s;
}
.toggle-slider::before {
  content: '';
  position: absolute;
  height: 18px;
  width: 18px;
  left: 3px;
  bottom: 3px;
  background: #fff;
  border-radius: 50%;
  transition: 0.3s;
}
.toggle-switch input:checked + .toggle-slider {
  background: #667eea;
}
.toggle-switch input:checked + .toggle-slider::before {
  transform: translateX(20px);
}

/* 提交按钮 */
.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding-top: 24px;
  border-top: 1px solid #f0f1f3;
  margin-top: 8px;
}
.btn-cancel {
  padding: 10px 24px;
  border: 1px solid #d1d5db;
  border-radius: 10px;
  background: #fff;
  color: #6b7280;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-cancel:hover {
  background: #f3f4f6;
}
.btn-submit {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 28px;
  border: none;
  border-radius: 10px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-submit:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(102,126,234,0.35);
}
.btn-submit:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.spinner-sm {
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255,255,255,0.3);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
  display: inline-block;
}
@keyframes spin { to { transform: rotate(360deg); } }

/* 结果弹窗 */
.result-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0,0,0,0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}
.result-card {
  background: #fff;
  border-radius: 20px;
  padding: 36px 32px;
  width: 460px;
  max-width: 90vw;
  text-align: center;
  box-shadow: 0 20px 60px rgba(0,0,0,0.15);
}
.result-icon {
  margin-bottom: 16px;
}
.result-card.success .result-icon { color: #10b981; }
.result-card.error .result-icon { color: #ef4444; }
.result-title {
  font-size: 20px;
  font-weight: 700;
  color: #1f2937;
  margin: 0 0 8px;
}
.result-msg {
  font-size: 14px;
  color: #6b7280;
  margin: 0 0 20px;
  line-height: 1.5;
}
.result-details {
  background: #f9fafb;
  border-radius: 10px;
  padding: 16px;
  margin-bottom: 20px;
  text-align: left;
}
.detail-row {
  display: flex;
  justify-content: space-between;
  padding: 6px 0;
  border-bottom: 1px solid #f0f1f3;
  font-size: 13px;
}
.detail-row:last-child {
  border-bottom: none;
}
.detail-label {
  color: #9ca3af;
}
.detail-value {
  color: #1f2937;
  font-weight: 500;
}
.result-actions {
  display: flex;
  gap: 10px;
  justify-content: center;
}
.btn-primary {
  padding: 10px 24px;
  border: none;
  border-radius: 10px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
}
.btn-secondary {
  padding: 10px 24px;
  border: 1px solid #d1d5db;
  border-radius: 10px;
  background: #fff;
  color: #374151;
  font-size: 14px;
  cursor: pointer;
}

.fade-enter-active, .fade-leave-active {
  transition: opacity 0.25s ease;
}
.fade-enter-from, .fade-leave-to {
  opacity: 0;
}

@media (max-width: 768px) {
  .form-grid.cols-3, .form-grid.cols-4 { grid-template-columns: 1fr; }
  .form-grid.cols-2 { grid-template-columns: 1fr; }
  .workload-type-bar { flex-direction: column; align-items: flex-start; }
  .type-options { width: 100%; }
  .type-chip { flex: 1; }
}
</style>
