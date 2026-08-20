<template>
  <!-- 悬浮按钮 -->
  <div class="ai-fab" :class="{ active: isOpen }" @click="togglePanel">
    <div class="ai-fab-inner">
      <transition name="fab-icon" mode="out-in">
        <svg v-if="!isOpen" key="ai" viewBox="0 0 24 24" fill="none" class="fab-svg">
          <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2z" fill="url(#fab-grad)"/>
          <defs><linearGradient id="fab-grad" x1="0" y1="0" x2="24" y2="24">
            <stop offset="0%" stop-color="#818cf8"/><stop offset="100%" stop-color="#6366f1"/>
          </linearGradient></defs>
          <circle cx="8.5" cy="11" r="1.2" fill="#fff"/>
          <circle cx="12" cy="11" r="1.2" fill="#fff"/>
          <circle cx="15.5" cy="11" r="1.2" fill="#fff"/>
          <path d="M7 15.5c1.1 1.3 2.9 2 5 2s3.9-.7 5-2" stroke="#fff" stroke-width="1.5" stroke-linecap="round" fill="none"/>
        </svg>
        <svg v-else key="close" viewBox="0 0 24 24" fill="none" class="fab-svg">
          <line x1="6" y1="6" x2="18" y2="18" stroke="#fff" stroke-width="2.5" stroke-linecap="round"/>
          <line x1="18" y1="6" x2="6" y2="18" stroke="#fff" stroke-width="2.5" stroke-linecap="round"/>
        </svg>
      </transition>
    </div>
    <span v-if="pendingCount > 0 && !isOpen" class="fab-badge">{{ pendingCount }}</span>
  </div>

  <!-- 聊天面板 -->
  <transition name="panel-slide">
    <div
      v-if="isOpen"
      class="ai-panel"
      :class="['mode-' + panelMode]"
      :style="panelStyle"
    >
      <!-- 头部 -->
      <div class="panel-header">
        <div class="header-left">
          <div class="ai-avatar-sm">
            <svg viewBox="0 0 24 24" fill="none">
              <circle cx="12" cy="12" r="10" fill="url(#head-grad)"/>
              <defs><linearGradient id="head-grad" x1="0" y1="0" x2="24" y2="24">
                <stop offset="0%" stop-color="#818cf8"/><stop offset="100%" stop-color="#6366f1"/>
              </linearGradient></defs>
              <path d="M8 14s1.5 2 4 2 4-2 4-2" stroke="#fff" stroke-width="1.5" stroke-linecap="round" fill="none"/>
              <circle cx="9" cy="10" r="1" fill="#fff"/><circle cx="15" cy="10" r="1" fill="#fff"/>
            </svg>
          </div>
          <div class="header-info">
            <span class="header-title">K8s AI 助手</span>
            <span class="header-status" :class="{ online: aiEnabled }">
              {{ aiEnabled ? '在线' : '离线' }}
            </span>
          </div>
        </div>
        <div class="header-actions">
          <!-- 模型选择器 -->
          <div class="model-selector-wrap" v-if="aiEnabled && providers.length > 0">
            <button class="header-btn model-trigger" @click.stop="showModelPicker = !showModelPicker" :title="currentModelLabel">
              <span class="model-icon-mini">{{ currentProviderIcon }}</span>
            </button>
            <transition name="picker-fade">
              <div v-if="showModelPicker" class="model-picker" @click.stop>
                <div class="picker-header">
                  <span class="picker-title">选择 AI 模型</span>
                  <button class="picker-close" @click="showModelPicker = false">&times;</button>
                </div>
                <div class="picker-body">
                  <div v-for="provider in providers" :key="provider.id" class="provider-group">
                    <div class="provider-label">
                      <span class="provider-icon">{{ providerIconMap[provider.icon] || '✨' }}</span>
                      <span class="provider-name">{{ provider.name }}</span>
                    </div>
                    <div
                      v-for="model in provider.models" :key="model.id"
                      :class="['model-option', { active: selectedProvider === provider.id && selectedModel === model.id }]"
                      @click="selectModel(provider.id, model.id, model.name, provider.icon)"
                    >
                      <div class="model-main">
                        <span class="model-name">{{ model.name }}</span>
                        <span v-if="model.capability" :class="['cap-tag', model.capability]">{{ capabilityLabel[model.capability] || model.capability }}</span>
                      </div>
                      <div class="model-desc" v-if="model.description">{{ model.description }}</div>
                    </div>
                  </div>
                </div>
              </div>
            </transition>
          </div>
          <button v-if="pendingCount > 0" class="header-btn approval-btn" @click="activeTab = 'approvals'" title="待审批">
            <svg viewBox="0 0 20 20" fill="currentColor"><path d="M9 2a1 1 0 000 2h2a1 1 0 100-2H9z"/><path fill-rule="evenodd" d="M4 5a2 2 0 012-2 3 3 0 003 3h2a3 3 0 003-3 2 2 0 012 2v11a2 2 0 01-2 2H6a2 2 0 01-2-2V5zm9.707 5.707a1 1 0 00-1.414-1.414L9 12.586l-1.293-1.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"/></svg>
            <span class="btn-badge">{{ pendingCount }}</span>
          </button>
          <button class="header-btn" @click="startNewChat" title="新对话">
            <svg viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clip-rule="evenodd"/></svg>
          </button>
          <button class="header-btn" @click="activeTab = activeTab === 'history' ? 'chat' : 'history'" title="历史记录">
            <svg viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clip-rule="evenodd"/></svg>
          </button>
          <!-- 窗口控制按钮（3 个：最小化 / 最大化 / 关闭） -->
          <span class="header-divider"></span>
          <button class="header-btn win-ctrl" @click="toggleMinimize" :title="panelMode === 'minimized' ? '展开' : '最小化'">
            <svg v-if="panelMode !== 'minimized'" viewBox="0 0 20 20" fill="currentColor"><rect x="4" y="13" width="12" height="2" rx="1"/></svg>
            <svg v-else viewBox="0 0 20 20" fill="currentColor"><path d="M5 8l5 5 5-5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" fill="none"/></svg>
          </button>
          <button class="header-btn win-ctrl" @click="toggleMaximize" :title="panelMode === 'maximized' ? '还原' : '最大化'">
            <svg v-if="panelMode !== 'maximized'" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.8"><rect x="4.5" y="4.5" width="11" height="11" rx="1.5"/></svg>
            <svg v-else viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.8"><rect x="7" y="4" width="9" height="9" rx="1.2"/><rect x="4" y="7" width="9" height="9" rx="1.2" fill="#fff"/></svg>
          </button>
          <button class="header-btn win-ctrl close" @click="isOpen = false" title="关闭">
            <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="6" y1="6" x2="14" y2="14"/><line x1="14" y1="6" x2="6" y2="14"/></svg>
          </button>
        </div>
      </div>

      <!-- Tab 切换区 -->
      <div class="panel-tabs" v-if="activeTab !== 'chat'">
        <button :class="['tab-btn', { active: activeTab === 'history' }]" @click="activeTab = 'history'">历史会话</button>
        <button :class="['tab-btn', { active: activeTab === 'approvals' }]" @click="activeTab = 'approvals'">
          {{ isApprovalAdmin ? '待审批' : '我的申请' }}
          <span v-if="pendingCount > 0" class="tab-badge">{{ pendingCount }}</span>
        </button>
      </div>

      <!-- 多轮对话状态条（仅 chat tab 且有消息时显示） -->
      <div v-if="activeTab === 'chat' && messages.length > 0" class="context-bar">
        <div class="context-bar-left">
          <span class="ctx-dot"></span>
          <span class="ctx-text">
            <strong>多轮对话</strong>
            · 第 <b class="ctx-round">{{ contextRound || Math.ceil(messages.length / 2) }}</b> 轮
            <span class="ctx-sep">|</span>
            已记忆 <b>{{ historyCount }}</b> 条上下文
          </span>
        </div>
        <button class="ctx-clear-btn" @click="startNewChat" title="清空上下文，开启全新对话">
          <svg viewBox="0 0 16 16" fill="currentColor"><path d="M2.5 1a1 1 0 00-1 1v1a1 1 0 001 1H3v9a2 2 0 002 2h6a2 2 0 002-2V4h.5a1 1 0 001-1V2a1 1 0 00-1-1H10a1 1 0 00-1-1H7a1 1 0 00-1 1H2.5zm3 4a.5.5 0 01.5.5v7a.5.5 0 01-1 0v-7a.5.5 0 01.5-.5zM8 5a.5.5 0 01.5.5v7a.5.5 0 01-1 0v-7A.5.5 0 018 5zm3 .5v7a.5.5 0 01-1 0v-7a.5.5 0 011 0z"/></svg>
          <span>新对话</span>
        </button>
      </div>

      <!-- 聊天内容区 -->
      <div v-show="activeTab === 'chat'" class="panel-body" ref="chatBody">
        <!-- 欢迎消息 -->
        <div v-if="messages.length === 0" class="welcome-area">
          <div class="welcome-icon">
            <svg viewBox="0 0 48 48" fill="none">
              <circle cx="24" cy="24" r="22" fill="url(#welcome-grad)" opacity="0.1"/>
              <circle cx="24" cy="24" r="16" fill="url(#welcome-grad)" opacity="0.15"/>
              <defs><linearGradient id="welcome-grad" x1="0" y1="0" x2="48" y2="48">
                <stop offset="0%" stop-color="#818cf8"/><stop offset="100%" stop-color="#6366f1"/>
              </linearGradient></defs>
              <text x="24" y="30" text-anchor="middle" font-size="20">🤖</text>
            </svg>
          </div>
          <h3 class="welcome-title">Hi，我是 K8s AI 助手</h3>
          <p class="welcome-desc">我可以回答任何问题，也能直接操作 K8s 平台。<br/>高危操作会自动触发审批流程。</p>
          <div class="welcome-features">
            <span class="feat-chip">💬 多轮对话</span>
            <span class="feat-chip">🧠 上下文记忆</span>
            <span class="feat-chip">🔧 Function Calling</span>
            <span class="feat-chip">⚠️ 高危审批</span>
          </div>
          <!-- 分组快捷按钮 -->
          <div class="quick-group">
            <div class="quick-group-label"><span class="qg-icon">💬</span> 常规问答</div>
            <div class="quick-actions">
              <button v-for="q in quickGeneral" :key="q.text" class="quick-btn general" @click="sendQuickQuestion(q.text)">
                {{ q.icon }} {{ q.text }}
              </button>
            </div>
          </div>
          <div class="quick-group">
            <div class="quick-group-label"><span class="qg-icon">🖥️</span> 平台操作</div>
            <div class="quick-actions">
              <button v-for="q in quickPlatform" :key="q.text" class="quick-btn platform" @click="sendQuickQuestion(q.text)">
                {{ q.icon }} {{ q.text }}
              </button>
            </div>
          </div>
        </div>

        <!-- 消息列表 -->
        <div v-for="(msg, idx) in messages" :key="idx" :class="['msg-row', msg.role]">
          <div v-if="msg.role === 'assistant'" class="msg-avatar">
            <svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="10" fill="url(#msg-grad)"/>
              <defs><linearGradient id="msg-grad" x1="0" y1="0" x2="24" y2="24"><stop offset="0%" stop-color="#818cf8"/><stop offset="100%" stop-color="#6366f1"/></linearGradient></defs>
              <text x="12" y="16" text-anchor="middle" font-size="10" fill="#fff">AI</text>
            </svg>
          </div>
          <div :class="['msg-bubble', msg.role, { 'msg-error': msg.isError }]">
            <!-- 意图标签 -->
            <div v-if="msg.role === 'assistant' && msg.intentTag" :class="['intent-tag', msg.intentTag.type]">
              <span class="intent-icon">{{ msg.intentTag.icon }}</span>
              <span class="intent-label">{{ msg.intentTag.label }}</span>
              <span v-if="msg.toolsCalled && msg.toolsCalled.length" class="intent-tools">
                {{ msg.toolsCalled.join(', ') }}
              </span>
            </div>
            <div class="msg-text" v-html="formatMessage(msg.content)"></div>
            <!-- 错误详情折叠 -->
            <div v-if="msg.isError && msg.errorInfo" class="error-detail">
              <div class="error-detail-header" @click="msg._showDetail = !msg._showDetail">
                <span class="error-detail-icon">{{ msg._showDetail ? '▼' : '▶' }}</span>
                <span>排查信息</span>
              </div>
              <div v-if="msg._showDetail" class="error-detail-body">
                <div v-if="msg.errorInfo.code" class="error-field">
                  <span class="error-label">错误码:</span>
                  <span class="error-value">{{ msg.errorInfo.code }}</span>
                </div>
                <div v-if="msg.errorInfo.detail" class="error-field">
                  <span class="error-label">详情:</span>
                  <span class="error-value">{{ msg.errorInfo.detail }}</span>
                </div>
                <div class="error-field">
                  <span class="error-label">时间:</span>
                  <span class="error-value">{{ msg.errorInfo.timestamp }}</span>
                </div>
                <div class="error-tip">查看 AI 日志: storage/logs/ai.log</div>
              </div>
            </div>
            <!-- 审批提示 -->
            <div v-if="msg.pendingTools && msg.pendingTools.length" class="approval-notice">
              <div class="approval-icon">⚠️</div>
              <div class="approval-info">
                <span class="approval-label">需要审批</span>
                <div v-for="pt in msg.pendingTools" :key="pt.approval_id" class="approval-item">
                  <div class="approval-item-header">
                    <span class="tool-name">{{ pt.tool_name }}</span>
                    <span :class="['risk-tag', pt.risk_level]">{{ pt.risk_level }}</span>
                  </div>
                  <div v-if="pt.tool_args" class="approval-params">
                    <div v-if="pt.tool_args.namespace" class="param-row">
                      <span class="param-label">命名空间:</span>
                      <span class="param-value">{{ pt.tool_args.namespace }}</span>
                    </div>
                    <div v-if="pt.resource_name" class="param-row">
                      <span class="param-label">资源名:</span>
                      <span class="param-value">{{ pt.resource_name }}</span>
                    </div>
                    <div v-if="pt.cluster_id" class="param-row">
                      <span class="param-label">集群ID:</span>
                      <span class="param-value">{{ pt.cluster_id }}</span>
                    </div>
                    <div v-if="pt.tool_args.replicas != null" class="param-row">
                      <span class="param-label">目标副本数:</span>
                      <span class="param-value">{{ pt.tool_args.replicas }}</span>
                    </div>
                    <div v-if="pt.tool_args.image" class="param-row">
                      <span class="param-label">目标镜像:</span>
                      <span class="param-value">{{ pt.tool_args.image }}</span>
                    </div>
                  </div>
                  <div class="approval-item-id">审批ID: {{ pt.approval_id }}</div>
                </div>
              </div>
            </div>
            <span class="msg-time">{{ formatTime(msg.time) }}</span>
          </div>
        </div>

        <!-- 加载动画 -->
        <div v-if="loading" class="msg-row assistant">
          <div class="msg-avatar">
            <svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="10" fill="url(#load-grad)"/>
              <defs><linearGradient id="load-grad" x1="0" y1="0" x2="24" y2="24"><stop offset="0%" stop-color="#818cf8"/><stop offset="100%" stop-color="#6366f1"/></linearGradient></defs>
              <text x="12" y="16" text-anchor="middle" font-size="10" fill="#fff">AI</text>
            </svg>
          </div>
          <div class="msg-bubble assistant typing">
            <div class="typing-dots">
              <span></span><span></span><span></span>
            </div>
          </div>
        </div>
      </div>

      <!-- 历史会话列表 -->
      <div v-show="activeTab === 'history'" class="panel-body history-list">
        <div v-if="conversations.length === 0" class="empty-state">
          <span class="empty-icon">💬</span>
          <span>暂无历史会话</span>
        </div>
        <div v-for="conv in conversations" :key="conv.id" class="history-item" @click="loadConversation(conv.id)">
          <div class="history-title">{{ conv.title || '新对话' }}</div>
          <div class="history-meta">
            <span>{{ conv.message_count || 0 }} 条消息</span>
            <span>{{ formatDate(conv.updated_at) }}</span>
          </div>
          <button class="history-delete" @click.stop="removeConversation(conv.id)">
            <svg viewBox="0 0 16 16" fill="currentColor"><path d="M5.5 5.5A.5.5 0 016 6v6a.5.5 0 01-1 0V6a.5.5 0 01.5-.5zm2.5 0a.5.5 0 01.5.5v6a.5.5 0 01-1 0V6a.5.5 0 01.5-.5zm3 .5a.5.5 0 00-1 0v6a.5.5 0 001 0V6z"/><path fill-rule="evenodd" d="M14.5 3a1 1 0 01-1 1H13v9a2 2 0 01-2 2H5a2 2 0 01-2-2V4h-.5a1 1 0 01-1-1V2a1 1 0 011-1H5a1 1 0 011-1h4a1 1 0 011 1h2.5a1 1 0 011 1v1zM4.118 4L4 4.059V13a1 1 0 001 1h6a1 1 0 001-1V4.059L11.882 4H4.118zM2.5 3V2h11v1h-11z"/></svg>
          </button>
        </div>
      </div>

      <!-- 审批列表 -->
      <div v-show="activeTab === 'approvals'" class="panel-body approval-list">
        <!-- 角色标识条 -->
        <div class="approval-role-bar">
          <div :class="['role-indicator', isApprovalAdmin ? 'admin' : 'user']">
            <span class="role-dot"></span>
            <span class="role-text">{{ isApprovalAdmin ? '审批管理员' : '我的申请' }}</span>
          </div>
          <span class="role-hint" v-if="!isApprovalAdmin">只有管理员角色可审批操作</span>
        </div>

        <div v-if="approvals.length === 0" class="empty-state">
          <span class="empty-icon">{{ isApprovalAdmin ? '✅' : '📋' }}</span>
          <span>{{ isApprovalAdmin ? '暂无待审批操作' : '暂无待处理的申请' }}</span>
        </div>
        <div v-for="ap in approvals" :key="ap.id" class="approval-card">
          <div class="approval-card-header">
            <span class="tool-name">{{ ap.tool_name || ap.action }}</span>
            <span :class="['risk-tag', ap.risk_level || 'write']">{{ ap.risk_level || 'write' }}</span>
            <span v-if="ap.can_approve" class="can-approve-badge">可审批</span>
          </div>
          <div class="approval-card-desc">{{ ap.summary || ap.description }}</div>
          <!-- 申请人信息 -->
          <div class="approval-requester">
            <div class="requester-avatar">
              <svg viewBox="0 0 24 24" fill="none">
                <circle cx="12" cy="8" r="4" fill="#94a3b8"/>
                <path d="M4 20c0-4.418 3.582-8 8-8s8 3.582 8 8" fill="#94a3b8" opacity="0.3"/>
              </svg>
            </div>
            <div class="requester-info">
              <span class="requester-name">{{ ap.request_user_name || '未知用户' }}</span>
              <span class="requester-label">申请人</span>
            </div>
            <span class="requester-time">{{ formatApprovalTime(ap.created_at) }}</span>
          </div>
          <div class="approval-card-detail">
            <div v-if="ap.namespace" class="detail-row">
              <span class="detail-label">命名空间:</span>
              <span class="detail-value">{{ ap.namespace }}</span>
            </div>
            <div v-if="ap.resource_name" class="detail-row">
              <span class="detail-label">资源名:</span>
              <span class="detail-value">{{ ap.resource_name }}</span>
            </div>
            <div v-if="ap.cluster_id" class="detail-row">
              <span class="detail-label">集群ID:</span>
              <span class="detail-value">{{ ap.cluster_id }}</span>
            </div>
            <div v-if="ap.operation_json" class="detail-row">
              <span class="detail-label">操作参数:</span>
              <span class="detail-value param-json">{{ formatToolArgs(ap.operation_json) }}</span>
            </div>
          </div>
          <!-- 操作按钮：管理员且非自己提交的才能审批 -->
          <div v-if="ap.can_approve" class="approval-card-actions">
            <button class="approve-btn" @click="handleApprove(ap.id)">
              <svg viewBox="0 0 16 16" fill="currentColor" class="action-icon"><path d="M12.736 3.97a.733.733 0 011.047 0c.286.289.29.756.01 1.05L7.88 12.01a.733.733 0 01-1.065.02L3.217 8.384a.757.757 0 010-1.06.733.733 0 011.047 0l3.052 3.093 5.4-6.425z"/></svg>
              通过
            </button>
            <button class="reject-btn" @click="handleReject(ap.id)">
              <svg viewBox="0 0 16 16" fill="currentColor" class="action-icon"><path d="M4.646 4.646a.5.5 0 01.708 0L8 7.293l2.646-2.647a.5.5 0 01.708.708L8.707 8l2.647 2.646a.5.5 0 01-.708.708L8 8.707l-2.646 2.647a.5.5 0 01-.708-.708L7.293 8 4.646 5.354a.5.5 0 010-.708z"/></svg>
              拒绝
            </button>
          </div>
          <!-- 普通用户视角：只能取消自己的申请 -->
          <div v-else-if="!isApprovalAdmin" class="approval-card-actions">
            <div class="approval-status-hint">
              <svg viewBox="0 0 16 16" fill="currentColor" class="status-icon"><path d="M8 16A8 8 0 108 0a8 8 0 000 16zm.93-9.412l-1 4.705c-.07.34.029.533.304.533.194 0 .487-.07.686-.246l-.088.416c-.287.346-.92.598-1.465.598-.703 0-1.002-.422-.808-1.319l.738-3.468c.064-.293.006-.399-.287-.399l-.244.003.028-.307 2.136-.235zm.375-3.38a.77.77 0 11-1.54 0 .77.77 0 011.54 0z"/></svg>
              等待管理员审批
            </div>
            <button class="cancel-btn" @click="handleCancel(ap.id)">取消申请</button>
          </div>
          <!-- 管理员但是自己的申请 -->
          <div v-else class="approval-card-actions">
            <div class="approval-status-hint self-hint">
              <svg viewBox="0 0 16 16" fill="currentColor" class="status-icon"><path d="M8 15A7 7 0 118 1a7 7 0 010 14zm0 1A8 8 0 108 0a8 8 0 000 16z"/><path d="M7.002 11a1 1 0 112 0 1 1 0 01-2 0zM7.1 4.995a.905.905 0 111.8 0l-.35 3.507a.552.552 0 01-1.1 0L7.1 4.995z"/></svg>
              不可自审
            </div>
            <button class="cancel-btn" @click="handleCancel(ap.id)">取消申请</button>
          </div>
        </div>
      </div>

      <!-- 输入区 -->
      <div class="panel-footer" v-show="activeTab === 'chat'">
        <div class="input-wrapper">
          <textarea
            ref="inputRef"
            v-model="inputText"
            class="chat-input"
            placeholder="输入消息，如：查看所有 Pod、扩容 Deployment..."
            rows="1"
            @keydown.enter.exact.prevent="sendMessage"
            @input="autoResize"
          ></textarea>
          <button class="send-btn" :disabled="!inputText.trim() || loading" @click="sendMessage">
            <svg viewBox="0 0 20 20" fill="currentColor">
              <path d="M10.894 2.553a1 1 0 00-1.788 0l-7 14a1 1 0 001.169 1.409l5-1.429A1 1 0 009 15.571V11a1 1 0 112 0v4.571a1 1 0 00.725.962l5 1.428a1 1 0 001.17-1.408l-7-14z"/>
            </svg>
          </button>
        </div>
        <div class="input-hint">
          <span v-if="currentModelLabel !== 'AI 模型'" class="current-model-hint">{{ currentProviderIcon }} {{ currentModelLabel }}</span>
          <span v-else>Enter 发送</span>
          <span class="hint-sep">·</span>
          <span class="hint-general">💬 常规问答</span>
          <span class="hint-sep">·</span>
          <span class="hint-platform">🖥️ 平台操作</span>
          <span class="hint-sep">·</span>
          <span>AI 自动识别</span>
        </div>
      </div>

      <!-- 调整大小的拖拽手柄（左上角，大厂风格） -->
      <div
        v-if="panelMode === 'normal'"
        class="resize-handle resize-tl"
        @mousedown.prevent="startResize($event, 'tl')"
        title="拖动调整大小"
      ></div>
      <div
        v-if="panelMode === 'normal'"
        class="resize-handle resize-t"
        @mousedown.prevent="startResize($event, 't')"
      ></div>
      <div
        v-if="panelMode === 'normal'"
        class="resize-handle resize-l"
        @mousedown.prevent="startResize($event, 'l')"
      ></div>
    </div>
  </transition>
</template>

<script setup>
import { ref, reactive, computed, onMounted, nextTick, watch, onBeforeUnmount } from 'vue'
import { aiChat, getAIStatus, getAIModels, getConversations, getConversationMessages, deleteConversation, getApprovals, approveApproval, rejectApproval, cancelApproval, getPendingApprovalCount } from '@/api/ai'
import { Message } from '@arco-design/web-vue'
import MarkdownIt from 'markdown-it'
import permissionStore from '@/stores/permission'

const isOpen = ref(false)
const activeTab = ref('chat')
const loading = ref(false)
const inputText = ref('')
const inputRef = ref(null)
const chatBody = ref(null)
const aiEnabled = ref(false)
const conversationId = ref(null)
const contextRound = ref(0)   // 当前对话轮次（后端返回）
const historyCount = ref(0)  // 已记忆上下文条数（后端返回）
const pendingCount = ref(0)
const isApprovalAdmin = ref(false)  // 后端返回的审批管理员标识

// 窗口状态：normal | maximized | minimized
const PANEL_STORAGE_KEY = 'ai-panel-state-v1'
const MIN_WIDTH = 360
const MIN_HEIGHT = 420
const DEFAULT_WIDTH = 416   // 26rem
const DEFAULT_HEIGHT = 608  // 38rem

const panelMode = ref('normal')
const panelSize = reactive({ width: DEFAULT_WIDTH, height: DEFAULT_HEIGHT })

// 从 localStorage 恢复
const _saved = (() => {
  try { return JSON.parse(localStorage.getItem(PANEL_STORAGE_KEY) || '{}') } catch { return {} }
})()
if (_saved.width && _saved.height) {
  panelSize.width = Math.max(MIN_WIDTH, Math.min(_saved.width, window.innerWidth - 32))
  panelSize.height = Math.max(MIN_HEIGHT, Math.min(_saved.height, window.innerHeight - 32))
}
if (_saved.mode === 'maximized') panelMode.value = 'maximized'

const panelStyle = computed(() => {
  if (panelMode.value === 'maximized') {
    return {
      width: 'calc(100vw - 4rem)',
      height: 'calc(100vh - 4rem)',
      bottom: '2rem',
      right: '2rem',
      maxHeight: 'calc(100vh - 4rem)',
    }
  }
  if (panelMode.value === 'minimized') {
    return {
      width: panelSize.width + 'px',
      height: 'auto',
      maxHeight: 'none',
    }
  }
  return {
    width: panelSize.width + 'px',
    height: panelSize.height + 'px',
  }
})

const saveState = () => {
  try {
    localStorage.setItem(PANEL_STORAGE_KEY, JSON.stringify({
      width: panelSize.width,
      height: panelSize.height,
      mode: panelMode.value === 'minimized' ? 'normal' : panelMode.value,
    }))
  } catch {}
}

const toggleMaximize = () => {
  panelMode.value = panelMode.value === 'maximized' ? 'normal' : 'maximized'
  saveState()
}
const toggleMinimize = () => {
  panelMode.value = panelMode.value === 'minimized' ? 'normal' : 'minimized'
}

// 拖拽调整大小（面板错右下角固定，拖动左上顶调整宽高）
let _resize = null
const startResize = (e, dir) => {
  if (panelMode.value !== 'normal') return
  _resize = {
    dir,
    startX: e.clientX,
    startY: e.clientY,
    startW: panelSize.width,
    startH: panelSize.height,
  }
  document.body.style.userSelect = 'none'
  document.body.style.cursor = dir === 'tl' ? 'nwse-resize' : (dir === 't' ? 'ns-resize' : 'ew-resize')
  window.addEventListener('mousemove', onResize)
  window.addEventListener('mouseup', stopResize)
}
const onResize = (e) => {
  if (!_resize) return
  const dx = _resize.startX - e.clientX  // 左侧拖：向左为正
  const dy = _resize.startY - e.clientY  // 顶部拖：向上为正
  if (_resize.dir === 'tl' || _resize.dir === 'l') {
    const newW = Math.max(MIN_WIDTH, Math.min(_resize.startW + dx, window.innerWidth - 32))
    panelSize.width = newW
  }
  if (_resize.dir === 'tl' || _resize.dir === 't') {
    const newH = Math.max(MIN_HEIGHT, Math.min(_resize.startH + dy, window.innerHeight - 32))
    panelSize.height = newH
  }
}
const stopResize = () => {
  if (!_resize) return
  _resize = null
  document.body.style.userSelect = ''
  document.body.style.cursor = ''
  window.removeEventListener('mousemove', onResize)
  window.removeEventListener('mouseup', stopResize)
  saveState()
}

// 多模型选择器状态
const showModelPicker = ref(false)
const providers = reactive([])
const selectedProvider = ref('')
const selectedModel = ref('')
const currentModelLabel = ref('AI 模型')
const currentProviderIcon = ref('✨')

const providerIconMap = {
  openai: '🟢',
  deepseek: '🔵',
  zhipu: '🟣',
  qwen: '🟠',
  moonshot: '🌙',
}
const capabilityLabel = {
  chat: '对话',
  reasoning: '推理',
  code: '代码',
  vision: '视觉',
}

// Markdown 渲染器
const md = new MarkdownIt({
  html: false,
  linkify: true,
  breaks: true,
  typographer: true,
})

const messages = reactive([])
const conversations = reactive([])
const approvals = reactive([])

// 分组快捷问题
const quickGeneral = [
  { icon: '💡', text: '什么是 Kubernetes？' },
  { icon: '📝', text: '帮我写个 Nginx Deployment YAML' },
  { icon: '🔧', text: '如何排查 Pod CrashLoopBackOff？' },
]
const quickPlatform = [
  { icon: '📋', text: '查看所有集群' },
  { icon: '🔍', text: '列出 default 命名空间的 Pod' },
  { icon: '📊', text: '集群健康检查' },
  { icon: '🚀', text: '查看 Deployment 状态' },
]

const togglePanel = () => {
  isOpen.value = !isOpen.value
  if (isOpen.value) {
    checkAIStatus()
    loadModels()
    loadPendingCount()
    nextTick(() => inputRef.value?.focus())
  }
}

const checkAIStatus = async () => {
  try {
    const res = await getAIStatus()
    aiEnabled.value = res?.data?.enabled ?? false
    if (res?.data?.default_provider && !selectedProvider.value) {
      selectedProvider.value = res.data.default_provider
    }
    if (res?.data?.default_model && !selectedModel.value) {
      selectedModel.value = res.data.default_model
    }
  } catch { aiEnabled.value = false }
}

const loadModels = async () => {
  try {
    const res = await getAIModels()
    const data = res?.data
    providers.length = 0
    ;(data?.providers || []).forEach(p => providers.push(p))
    if (data?.default_provider && !selectedProvider.value) {
      selectedProvider.value = data.default_provider
    }
    if (data?.default_model && !selectedModel.value) {
      selectedModel.value = data.default_model
    }
    // 设置当前模型显示
    updateCurrentModelLabel()
  } catch { /* ignore */ }
}

const selectModel = (providerId, modelId, modelName, icon) => {
  selectedProvider.value = providerId
  selectedModel.value = modelId
  currentModelLabel.value = modelName
  currentProviderIcon.value = providerIconMap[icon] || '✨'
  showModelPicker.value = false
}

const updateCurrentModelLabel = () => {
  for (const p of providers) {
    for (const m of p.models) {
      if (p.id === selectedProvider.value && m.id === selectedModel.value) {
        currentModelLabel.value = m.name
        currentProviderIcon.value = providerIconMap[p.icon] || '✨'
        return
      }
    }
  }
}

// 点击外部关闭模型选择器
const closePickerOnOutside = (e) => {
  if (showModelPicker.value && !e.target.closest('.model-selector-wrap')) {
    showModelPicker.value = false
  }
}

const loadPendingCount = async () => {
  try {
    const res = await getPendingApprovalCount()
    pendingCount.value = res?.data?.count ?? 0
    isApprovalAdmin.value = res?.data?.is_admin ?? false
  } catch (err) {
    // 静默处理：token 过期、后端未启动、网络错误等都忽略
    pendingCount.value = 0
    isApprovalAdmin.value = false
  }
}

const startNewChat = () => {
  messages.length = 0
  conversationId.value = null
  contextRound.value = 0
  historyCount.value = 0
  activeTab.value = 'chat'
}

const sendQuickQuestion = (q) => {
  inputText.value = q
  sendMessage()
}

const sendMessage = async () => {
  const text = inputText.value.trim()
  if (!text || loading.value) return

  messages.push({ role: 'user', content: text, time: new Date() })
  inputText.value = ''
  autoResize()
  loading.value = true
  scrollToBottom()

  try {
    const res = await aiChat({
      message: text,
      conversation_id: conversationId.value || undefined,
      provider_id: selectedProvider.value || undefined,
      model_id: selectedModel.value || undefined,
    })
    const data = res?.data || res
    conversationId.value = data.conversation_id || conversationId.value
    // 同步多轮对话状态
    if (typeof data.context_round === 'number') contextRound.value = data.context_round
    if (typeof data.history_count === 'number') historyCount.value = data.history_count

    const toolsCalled = data.tools_called || []
    const pendingTools = data.pending_tools || []
    const intentTag = getIntentTag(toolsCalled, pendingTools)

    messages.push({
      role: 'assistant',
      content: data.reply || '抱歉，我暂时无法回答。',
      time: new Date(),
      pendingTools,
      toolsCalled,
      intentTag,
    })

    if (data.pending_tools?.length) {
      loadPendingCount()
    }
  } catch (e) {
    const errDetail = e?.data?.details || e?.msg || e?.message || ''
    const errCode = e?.code || e?.status || ''
    messages.push({
      role: 'assistant',
      content: '请求失败，请检查网络或 AI 服务配置。' + (errDetail ? `\n错误: ${errDetail}` : ''),
      time: new Date(),
      isError: true,
      errorInfo: {
        code: errCode,
        detail: errDetail,
        timestamp: new Date().toISOString(),
      },
      intentTag: { type: 'error', icon: '❌', label: 'AI 异常' },
    })
  } finally {
    loading.value = false
    scrollToBottom()
  }
}

const loadConversation = async (id) => {
  try {
    const res = await getConversationMessages(id)
    messages.length = 0
    conversationId.value = id
    const list = res?.data || []
    list.forEach(m => {
      messages.push({ role: m.role, content: m.content, time: new Date(m.created_at) })
    })
    // 恢复多轮对话轮次统计（仅计 user/assistant）
    const valid = list.filter(m => m.role === 'user' || (m.role === 'assistant' && m.content))
    contextRound.value = valid.filter(m => m.role === 'user').length
    historyCount.value = Math.max(0, valid.length - 1)
    activeTab.value = 'chat'
    scrollToBottom()
  } catch { Message.error('加载会话失败') }
}

const loadHistory = async () => {
  try {
    const res = await getConversations()
    conversations.length = 0
    ;(res?.data || []).forEach(c => conversations.push(c))
  } catch { /* ignore */ }
}

const removeConversation = async (id) => {
  try {
    await deleteConversation(id)
    const idx = conversations.findIndex(c => c.id === id)
    if (idx >= 0) conversations.splice(idx, 1)
    if (conversationId.value === id) startNewChat()
  } catch { Message.error('删除失败') }
}

const loadApprovals = async () => {
  try {
    const res = await getApprovals({ status: 'pending' })
    approvals.length = 0
    const data = res?.data
    // 后端新格式: { list: [...], total: N, is_admin: bool }
    const list = data?.list || data || []
    list.forEach(a => approvals.push(a))
    if (data?.is_admin !== undefined) {
      isApprovalAdmin.value = data.is_admin
    }
  } catch { /* ignore */ }
}

const handleApprove = async (id) => {
  try {
    await approveApproval(id, { comment: '通过' })
    Message.success('已通过')
    loadApprovals()
    loadPendingCount()
  } catch (e) {
    const detail = e?.data?.details || e?.msg || '操作失败'
    Message.error(detail)
  }
}

const handleReject = async (id) => {
  try {
    await rejectApproval(id, { comment: '拒绝' })
    Message.success('已拒绝')
    loadApprovals()
    loadPendingCount()
  } catch (e) {
    const detail = e?.data?.details || e?.msg || '操作失败'
    Message.error(detail)
  }
}

const handleCancel = async (id) => {
  try {
    await cancelApproval(id)
    Message.success('已取消')
    loadApprovals()
    loadPendingCount()
  } catch (e) {
    const detail = e?.data?.details || e?.msg || '操作失败'
    Message.error(detail)
  }
}

const scrollToBottom = () => {
  nextTick(() => {
    if (chatBody.value) chatBody.value.scrollTop = chatBody.value.scrollHeight
  })
}

const autoResize = () => {
  nextTick(() => {
    if (!inputRef.value) return
    inputRef.value.style.height = 'auto'
    inputRef.value.style.height = Math.min(inputRef.value.scrollHeight, 120) + 'px'
  })
}

// 根据工具调用情况判断意图标签
const getIntentTag = (toolsCalled, pendingTools) => {
  if (pendingTools && pendingTools.length > 0) {
    const hasCritical = pendingTools.some(t => t.risk_level === 'critical' || t.risk_level === 'danger')
    return {
      type: hasCritical ? 'danger' : 'approval',
      icon: hasCritical ? '🛑' : '⚠️',
      label: hasCritical ? '高危操作 · 需审批' : '写操作 · 需审批',
    }
  }
  if (toolsCalled && toolsCalled.length > 0) {
    return {
      type: 'platform',
      icon: '🖥️',
      label: '平台数据',
    }
  }
  return {
    type: 'general',
    icon: '💬',
    label: '常规回答',
  }
}

const formatMessage = (text) => {
  if (!text) return ''
  try {
    return md.render(text)
  } catch {
    // 回退到简单格式化
    return text
      .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
      .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
      .replace(/`([^`]+)`/g, '<code>$1</code>')
      .replace(/\n/g, '<br/>')
  }
}

const formatTime = (d) => {
  if (!d) return ''
  const date = d instanceof Date ? d : new Date(d)
  return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

const formatDate = (d) => {
  if (!d) return ''
  const date = new Date(d)
  const now = new Date()
  if (date.toDateString() === now.toDateString()) return '今天'
  return date.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' })
}

const formatToolArgs = (jsonStr) => {
  try {
    const obj = typeof jsonStr === 'string' ? JSON.parse(jsonStr) : jsonStr
    return Object.entries(obj).map(([k, v]) => `${k}: ${v}`).join(', ')
  } catch {
    return jsonStr
  }
}

const formatApprovalTime = (ts) => {
  if (!ts) return ''
  const date = new Date(ts * 1000)
  const now = new Date()
  const diffMs = now - date
  const diffMin = Math.floor(diffMs / 60000)
  if (diffMin < 1) return '刚刚'
  if (diffMin < 60) return `${diffMin} 分钟前`
  const diffHour = Math.floor(diffMin / 60)
  if (diffHour < 24) return `${diffHour} 小时前`
  return date.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

watch(activeTab, (tab) => {
  if (tab === 'history') loadHistory()
  if (tab === 'approvals') loadApprovals()
})

onMounted(() => {
  loadPendingCount()
  document.addEventListener('click', closePickerOnOutside)
})
onBeforeUnmount(() => {
  document.removeEventListener('click', closePickerOnOutside)
})
</script>

<style scoped>
/* ===== 模型选择器 ===== */
.model-selector-wrap { position: relative; }
.model-trigger { position: relative; }
.model-icon-mini { font-size: 1rem; line-height: 1; }
.model-picker {
  position: absolute;
  top: 2.5rem; right: 0;
  width: 18rem;
  max-height: 22rem;
  background: #fff;
  border-radius: 0.75rem;
  box-shadow: 0 20px 50px -12px rgba(0,0,0,0.2), 0 0 0 1px rgba(0,0,0,0.05);
  z-index: 10000;
  display: flex; flex-direction: column;
  overflow: hidden;
}
.picker-fade-enter-active { transition: all 0.2s ease-out; }
.picker-fade-leave-active { transition: all 0.15s ease-in; }
.picker-fade-enter-from { opacity: 0; transform: translateY(-8px) scale(0.95); }
.picker-fade-leave-to { opacity: 0; transform: translateY(-4px); }
.picker-header {
  display: flex; justify-content: space-between; align-items: center;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid #f1f5f9;
}
.picker-title { font-size: 0.8125rem; font-weight: 700; color: #1e293b; }
.picker-close {
  border: none; background: none; color: #94a3b8; font-size: 1.25rem;
  cursor: pointer; line-height: 1; padding: 0;
}
.picker-close:hover { color: #475569; }
.picker-body { overflow-y: auto; padding: 0.5rem; }
.provider-group { margin-bottom: 0.5rem; }
.provider-group:last-child { margin-bottom: 0; }
.provider-label {
  display: flex; align-items: center; gap: 0.375rem;
  padding: 0.375rem 0.5rem; font-size: 0.6875rem;
  color: #94a3b8; font-weight: 600; text-transform: uppercase;
  letter-spacing: 0.5px;
}
.provider-icon { font-size: 0.875rem; }
.model-option {
  padding: 0.5rem 0.75rem; border-radius: 0.5rem;
  cursor: pointer; transition: all 0.15s;
  border: 1.5px solid transparent;
}
.model-option:hover { background: #f8fafc; border-color: #e2e8f0; }
.model-option.active {
  background: linear-gradient(135deg, #eef2ff, #e0e7ff);
  border-color: #818cf8;
}
.model-main { display: flex; align-items: center; gap: 0.5rem; }
.model-name { font-size: 0.8125rem; font-weight: 600; color: #1e293b; }
.cap-tag {
  padding: 0.0625rem 0.375rem; border-radius: 0.25rem;
  font-size: 0.5625rem; font-weight: 600;
}
.cap-tag.chat { background: #dbeafe; color: #2563eb; }
.cap-tag.reasoning { background: #fef3c7; color: #d97706; }
.cap-tag.code { background: #dcfce7; color: #16a34a; }
.cap-tag.vision { background: #f3e8ff; color: #9333ea; }
.model-desc {
  font-size: 0.6875rem; color: #94a3b8; margin-top: 0.125rem;
  padding-left: 0;
}

/* ===== 悬浮按钮 ===== */
.ai-fab {
  position: fixed;
  bottom: 2rem;
  right: 2rem;
  width: 3.5rem;
  height: 3.5rem;
  z-index: 9999;
  cursor: pointer;
  filter: drop-shadow(0 4px 12px rgba(99, 102, 241, 0.4));
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}
.ai-fab:hover { transform: scale(1.1); filter: drop-shadow(0 6px 20px rgba(99, 102, 241, 0.5)); }
.ai-fab.active { transform: scale(0.95); }
.ai-fab-inner {
  width: 100%; height: 100%;
  background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
  border-radius: 1rem;
  display: flex; align-items: center; justify-content: center;
  transition: border-radius 0.3s ease;
}
.ai-fab.active .ai-fab-inner { border-radius: 50%; background: linear-gradient(135deg, #4f46e5 0%, #7c3aed 100%); }
.fab-svg { width: 1.75rem; height: 1.75rem; }
.fab-badge {
  position: absolute; top: -0.25rem; right: -0.25rem;
  min-width: 1.25rem; height: 1.25rem;
  background: linear-gradient(135deg, #ef4444, #f97316);
  color: #fff; font-size: 0.7rem; font-weight: 700;
  border-radius: 0.625rem;
  display: flex; align-items: center; justify-content: center;
  border: 2px solid #fff;
}
.fab-icon-enter-active, .fab-icon-leave-active { transition: all 0.2s ease; }
.fab-icon-enter-from { opacity: 0; transform: rotate(-90deg) scale(0.5); }
.fab-icon-leave-to { opacity: 0; transform: rotate(90deg) scale(0.5); }

/* ===== 聊天面板 ===== */
.ai-panel {
  position: fixed;
  bottom: 6.5rem; right: 2rem;
  width: 26rem; height: 38rem;
  max-height: calc(100vh - 8rem);
  background: #ffffff;
  border-radius: 1.25rem;
  box-shadow: 0 25px 60px -12px rgba(0, 0, 0, 0.15), 0 0 0 1px rgba(0, 0, 0, 0.05);
  display: flex; flex-direction: column;
  z-index: 9998;
  overflow: hidden;
  transition: width 0.25s cubic-bezier(0.4, 0, 0.2, 1),
              height 0.25s cubic-bezier(0.4, 0, 0.2, 1),
              bottom 0.25s cubic-bezier(0.4, 0, 0.2, 1),
              right 0.25s cubic-bezier(0.4, 0, 0.2, 1),
              border-radius 0.25s ease;
}
.ai-panel.mode-maximized {
  border-radius: 0.75rem;
  box-shadow: 0 30px 80px -10px rgba(0, 0, 0, 0.25), 0 0 0 1px rgba(0, 0, 0, 0.06);
}
.ai-panel.mode-minimized {
  height: auto !important;
}
.ai-panel.mode-minimized .panel-tabs,
.ai-panel.mode-minimized .panel-body,
.ai-panel.mode-minimized .panel-footer,
.ai-panel.mode-minimized .resize-handle {
  display: none !important;
}
.ai-panel.mode-minimized .panel-header {
  border-radius: 1.25rem;
}

/* 调整大小的拖拽手柄（仅 normal 模式可见） */
.resize-handle {
  position: absolute;
  z-index: 10;
  background: transparent;
  transition: background 0.15s;
}
.resize-handle.resize-tl {
  top: 0; left: 0;
  width: 18px; height: 18px;
  cursor: nwse-resize;
  border-top-left-radius: 1.25rem;
}
.resize-handle.resize-tl::before {
  content: '';
  position: absolute;
  top: 6px; left: 6px;
  width: 8px; height: 8px;
  background:
    linear-gradient(135deg, transparent 40%, rgba(99,102,241,0.5) 40%, rgba(99,102,241,0.5) 50%, transparent 50%, transparent 70%, rgba(99,102,241,0.5) 70%, rgba(99,102,241,0.5) 80%, transparent 80%);
  opacity: 0.6;
  transition: opacity 0.2s;
  border-radius: 2px;
}
.resize-handle.resize-tl:hover::before { opacity: 1; }
.resize-handle.resize-t {
  top: 0; left: 18px; right: 0;
  height: 5px;
  cursor: ns-resize;
}
.resize-handle.resize-l {
  top: 18px; left: 0; bottom: 0;
  width: 5px;
  cursor: ew-resize;
}
.resize-handle.resize-t:hover,
.resize-handle.resize-l:hover {
  background: rgba(99, 102, 241, 0.12);
}
.panel-slide-enter-active { transition: all 0.35s cubic-bezier(0.34, 1.56, 0.64, 1); }
.panel-slide-leave-active { transition: all 0.2s ease-in; }
.panel-slide-enter-from { opacity: 0; transform: translateY(20px) scale(0.95); }
.panel-slide-leave-to { opacity: 0; transform: translateY(10px) scale(0.98); }

/* ===== 头部 ===== */
.panel-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 1rem 1.25rem;
  background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 50%, #a78bfa 100%);
  color: #fff;
}
.header-left { display: flex; align-items: center; gap: 0.75rem; }
.ai-avatar-sm { width: 2.25rem; height: 2.25rem; }
.ai-avatar-sm svg { width: 100%; height: 100%; }
.header-info { display: flex; flex-direction: column; }
.header-title { font-size: 0.9375rem; font-weight: 700; letter-spacing: 0.3px; }
.header-status { font-size: 0.6875rem; opacity: 0.8; display: flex; align-items: center; gap: 0.3rem; }
.header-status::before { content: ''; width: 6px; height: 6px; border-radius: 50%; background: #94a3b8; }
.header-status.online::before { background: #4ade80; box-shadow: 0 0 6px #4ade80; }
.header-actions { display: flex; gap: 0.25rem; align-items: center; }
.header-divider {
  display: inline-block;
  width: 1px; height: 1.1rem;
  background: rgba(255,255,255,0.3);
  margin: 0 0.25rem;
}
.header-btn {
  width: 2rem; height: 2rem;
  background: rgba(255,255,255,0.15); border: none; border-radius: 0.5rem;
  color: #fff; cursor: pointer;
  display: flex; align-items: center; justify-content: center;
  transition: background 0.2s;
  position: relative;
}
.header-btn:hover { background: rgba(255,255,255,0.25); }
.header-btn.win-ctrl {
  width: 1.75rem; height: 1.75rem;
  background: rgba(255,255,255,0.1);
  border-radius: 0.4rem;
}
.header-btn.win-ctrl:hover { background: rgba(255,255,255,0.22); transform: scale(1.05); }
.header-btn.win-ctrl.close:hover { background: rgba(239, 68, 68, 0.85); }
.header-btn svg { width: 1rem; height: 1rem; }
.btn-badge {
  position: absolute; top: -0.2rem; right: -0.2rem;
  min-width: 1rem; height: 1rem;
  background: #ef4444; color: #fff; font-size: 0.6rem;
  border-radius: 0.5rem; display: flex; align-items: center; justify-content: center;
}

/* ===== Tabs ===== */
.panel-tabs {
  display: flex; border-bottom: 1px solid #e5e7eb; background: #fafbfc;
}
.tab-btn {
  flex: 1; padding: 0.625rem; border: none; background: none;
  font-size: 0.8125rem; font-weight: 500; color: #6b7280; cursor: pointer;
  position: relative; transition: color 0.2s;
}
.tab-btn.active { color: #6366f1; }
.tab-btn.active::after {
  content: ''; position: absolute; bottom: 0; left: 20%; right: 20%;
  height: 2px; background: #6366f1; border-radius: 1px;
}
.tab-badge {
  margin-left: 0.25rem; padding: 0 0.35rem;
  background: #ef4444; color: #fff; font-size: 0.625rem; border-radius: 0.5rem;
}

/* ===== 消息区 ===== */
.panel-body {
  flex: 1; overflow-y: auto; padding: 1rem;
  scroll-behavior: smooth;
  background: linear-gradient(180deg, #f8f9ff 0%, #ffffff 100%);
}

/* ===== 欢迎区 ===== */
.welcome-area {
  display: flex; flex-direction: column; align-items: center;
  padding: 1.5rem 1rem; text-align: center;
}
.welcome-icon { width: 4rem; height: 4rem; margin-bottom: 1rem; }
.welcome-icon svg { width: 100%; height: 100%; }
.welcome-title { font-size: 1.125rem; font-weight: 700; color: #1e293b; margin-bottom: 0.5rem; }
.welcome-desc { font-size: 0.8125rem; color: #64748b; line-height: 1.6; margin-bottom: 1rem; }

/* ===== 欢迎区能力胶囊 ===== */
.welcome-features {
  display: flex; flex-wrap: wrap; gap: 0.375rem;
  justify-content: center; margin-bottom: 1rem;
}
.feat-chip {
  padding: 0.25rem 0.625rem;
  background: linear-gradient(135deg, #eef2ff, #f5f3ff);
  border: 1px solid #e0e7ff;
  border-radius: 1rem;
  font-size: 0.6875rem; font-weight: 600;
  color: #4f46e5;
  letter-spacing: 0.2px;
}

/* ===== 多轮对话状态条 ===== */
.context-bar {
  display: flex; align-items: center; justify-content: space-between;
  padding: 0.375rem 0.75rem;
  background: linear-gradient(90deg, rgba(99, 102, 241, 0.06), rgba(139, 92, 246, 0.04));
  border-bottom: 1px solid rgba(99, 102, 241, 0.12);
  font-size: 0.6875rem; color: #475569;
  flex-shrink: 0;
}
.context-bar-left {
  display: flex; align-items: center; gap: 0.4375rem;
  min-width: 0; flex: 1;
}
.ctx-dot {
  width: 0.4375rem; height: 0.4375rem; border-radius: 50%;
  background: #10b981; flex-shrink: 0;
  box-shadow: 0 0 0 2px rgba(16, 185, 129, 0.18);
  animation: ctx-pulse 2.4s ease-in-out infinite;
}
@keyframes ctx-pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.65; transform: scale(0.85); }
}
.ctx-text { font-size: 0.6875rem; color: #475569; line-height: 1.4; }
.ctx-text strong { color: #4f46e5; font-weight: 700; }
.ctx-text b { color: #1e293b; font-weight: 700; padding: 0 0.0625rem; }
.ctx-text .ctx-round { color: #6366f1; }
.ctx-sep { color: #cbd5e1; margin: 0 0.25rem; }
.ctx-clear-btn {
  display: flex; align-items: center; gap: 0.25rem;
  padding: 0.25rem 0.5rem;
  background: transparent; border: 1px solid transparent;
  border-radius: 0.375rem;
  font-size: 0.6875rem; color: #64748b; cursor: pointer;
  transition: all 0.18s;
  flex-shrink: 0;
}
.ctx-clear-btn:hover {
  background: rgba(239, 68, 68, 0.08);
  border-color: rgba(239, 68, 68, 0.2);
  color: #dc2626;
}
.ctx-clear-btn svg { width: 0.75rem; height: 0.75rem; }
.quick-group {
  width: 100%; text-align: left; margin-bottom: 0.75rem;
}
.quick-group-label {
  display: flex; align-items: center; gap: 0.375rem;
  font-size: 0.6875rem; font-weight: 700; color: #64748b;
  margin-bottom: 0.375rem; padding-left: 0.125rem;
  text-transform: uppercase; letter-spacing: 0.5px;
}
.qg-icon { font-size: 0.8125rem; }
.quick-actions { display: flex; flex-wrap: wrap; gap: 0.375rem; }
.quick-btn {
  padding: 0.4375rem 0.75rem; border: 1px solid #e2e8f0; border-radius: 1.25rem;
  background: #fff; color: #475569; font-size: 0.6875rem; cursor: pointer;
  transition: all 0.2s; white-space: nowrap;
}
.quick-btn:hover {
  transform: translateY(-1px); box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}
.quick-btn.general:hover { border-color: #6366f1; color: #6366f1; background: #eef2ff; }
.quick-btn.platform:hover { border-color: #0ea5e9; color: #0369a1; background: #e0f2fe; }
.quick-btn.platform { border-color: #bae6fd; }

/* ===== 消息行 ===== */
.msg-row { display: flex; gap: 0.5rem; margin-bottom: 1rem; max-width: 100%; }
.msg-row.user { flex-direction: row-reverse; }
.msg-avatar { width: 1.75rem; height: 1.75rem; flex-shrink: 0; margin-top: 0.125rem; }
.msg-avatar svg { width: 100%; height: 100%; }
.msg-bubble {
  max-width: 88%; padding: 0.75rem 1rem;
  border-radius: 1rem; position: relative;
  font-size: 0.8125rem; line-height: 1.6;
  word-break: break-word;
}
.msg-bubble.assistant {
  background: #ffffff; color: #1e293b;
  border-bottom-left-radius: 0.25rem;
  border: 1px solid #f1f5f9;
  box-shadow: 0 1px 4px rgba(0,0,0,0.04);
}
.msg-bubble.user {
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: #fff;
  border-bottom-right-radius: 0.25rem;
}
.msg-text :deep(code) {
  background: rgba(99, 102, 241, 0.08); padding: 0.125rem 0.375rem;
  border-radius: 0.25rem; font-size: 0.75rem; font-family: 'Fira Code', 'Cascadia Code', monospace;
  color: #6366f1;
}
.msg-bubble.user .msg-text :deep(code) { background: rgba(255,255,255,0.2); color: #fff; }

/* === Markdown 渲染样式（大厂风格） === */
.msg-text :deep(p) { margin: 0 0 0.5rem 0; line-height: 1.6; }
.msg-text :deep(p:last-child) { margin-bottom: 0; }

/* 表格：企业级数据展示 */
.msg-text :deep(table) {
  width: 100%; border-collapse: separate; border-spacing: 0;
  margin: 0.625rem 0; font-size: 0.75rem;
  border: 1px solid #e2e8f0; border-radius: 0.5rem;
  overflow: hidden;
}
.msg-text :deep(thead) {
  background: linear-gradient(135deg, #f8fafc, #f1f5f9);
}
.msg-text :deep(th) {
  padding: 0.5rem 0.75rem; text-align: left;
  font-weight: 700; color: #475569; font-size: 0.6875rem;
  text-transform: uppercase; letter-spacing: 0.3px;
  border-bottom: 2px solid #e2e8f0;
}
.msg-text :deep(td) {
  padding: 0.4375rem 0.75rem; color: #334155;
  border-bottom: 1px solid #f1f5f9;
}
.msg-text :deep(tr:last-child td) { border-bottom: none; }
.msg-text :deep(tr:hover td) { background: #f8fafc; }
.msg-text :deep(tbody tr:nth-child(even) td) { background: #fafbfc; }

/* 代码块：暗色主题 */
.msg-text :deep(pre) {
  margin: 0.625rem 0; padding: 0.875rem 1rem;
  background: #1e293b; color: #e2e8f0;
  border-radius: 0.625rem; overflow-x: auto;
  font-size: 0.75rem; line-height: 1.7;
  font-family: 'Fira Code', 'Cascadia Code', 'Consolas', monospace;
  border: 1px solid #334155;
}
.msg-text :deep(pre code) {
  background: transparent !important; color: inherit;
  padding: 0; font-size: inherit;
}
.msg-bubble.user .msg-text :deep(pre) {
  background: rgba(255,255,255,0.12); border-color: rgba(255,255,255,0.2);
  color: #f1f5f9;
}

/* 列表 */
.msg-text :deep(ul), .msg-text :deep(ol) {
  margin: 0.375rem 0; padding-left: 1.25rem;
}
.msg-text :deep(li) {
  margin-bottom: 0.25rem; line-height: 1.6; color: #334155;
}
.msg-text :deep(li::marker) { color: #6366f1; }

/* 引用块 */
.msg-text :deep(blockquote) {
  margin: 0.5rem 0; padding: 0.5rem 0.75rem;
  border-left: 3px solid #6366f1;
  background: #f8fafc; border-radius: 0 0.375rem 0.375rem 0;
  color: #64748b; font-size: 0.8125rem;
}

/* 分割线 */
.msg-text :deep(hr) {
  border: none; border-top: 1px solid #e2e8f0;
  margin: 0.75rem 0;
}

/* 链接 */
.msg-text :deep(a) {
  color: #6366f1; text-decoration: none;
  border-bottom: 1px dashed #6366f1;
}
.msg-text :deep(a:hover) { border-bottom-style: solid; }

/* 标题 */
.msg-text :deep(h1), .msg-text :deep(h2), .msg-text :deep(h3),
.msg-text :deep(h4), .msg-text :deep(h5), .msg-text :deep(h6) {
  margin: 0.75rem 0 0.375rem 0; font-weight: 700; color: #1e293b; line-height: 1.3;
}
.msg-text :deep(h1) { font-size: 1rem; }
.msg-text :deep(h2) { font-size: 0.9375rem; }
.msg-text :deep(h3) { font-size: 0.875rem; }

/* 加粗 */
.msg-text :deep(strong) { color: #1e293b; font-weight: 700; }
.msg-bubble.user .msg-text :deep(strong) { color: #fff; }
.msg-time {
  display: block; font-size: 0.625rem; opacity: 0.5; margin-top: 0.375rem; text-align: right;
}

/* ===== 审批提醒 ===== */
.approval-notice {
  display: flex; gap: 0.5rem; margin-top: 0.75rem;
  padding: 0.625rem; background: #fffbeb; border: 1px solid #fde68a;
  border-radius: 0.5rem;
}
.approval-icon { font-size: 1rem; }
.approval-label { font-size: 0.75rem; font-weight: 600; color: #92400e; }
.approval-item {
  display: flex; flex-direction: column; gap: 0.25rem; margin-top: 0.375rem;
  padding: 0.375rem 0;
  border-top: 1px dashed #fde68a;
}
.approval-item:first-child { border-top: none; padding-top: 0; }
.approval-item-header {
  display: flex; align-items: center; gap: 0.375rem;
}
.approval-params {
  padding: 0.25rem 0;
}
.param-row {
  display: flex; gap: 0.25rem; font-size: 0.6875rem; line-height: 1.5;
}
.param-label { color: #92400e; font-weight: 500; flex-shrink: 0; }
.param-value { color: #78716c; font-family: 'Fira Code', monospace; font-size: 0.6875rem; word-break: break-all; }
.approval-item-id { font-size: 0.625rem; color: #b8860b; font-style: italic; }
.tool-name { font-size: 0.75rem; color: #78716c; font-family: monospace; }
.risk-tag {
  padding: 0.0625rem 0.375rem; border-radius: 0.25rem;
  font-size: 0.625rem; font-weight: 600; text-transform: uppercase;
}
.risk-tag.write { background: #dbeafe; color: #1d4ed8; }
.risk-tag.danger { background: #fee2e2; color: #dc2626; }
.risk-tag.critical { background: #fecaca; color: #991b1b; }

/* ===== 打字动画 ===== */
.typing-dots { display: flex; gap: 0.3rem; padding: 0.25rem 0; }
.typing-dots span {
  width: 0.5rem; height: 0.5rem;
  background: #94a3b8; border-radius: 50%;
  animation: dot-bounce 1.4s infinite ease-in-out both;
}
.typing-dots span:nth-child(1) { animation-delay: -0.32s; }
.typing-dots span:nth-child(2) { animation-delay: -0.16s; }
@keyframes dot-bounce {
  0%, 80%, 100% { transform: scale(0.6); opacity: 0.4; }
  40% { transform: scale(1); opacity: 1; }
}

/* ===== 历史会话 ===== */
.history-list, .approval-list { padding: 0.75rem; }
.empty-state {
  display: flex; flex-direction: column; align-items: center;
  padding: 3rem 1rem; color: #94a3b8; font-size: 0.8125rem; gap: 0.5rem;
}
.empty-icon { font-size: 2rem; }
.history-item {
  padding: 0.75rem; border-radius: 0.75rem;
  cursor: pointer; transition: background 0.2s; position: relative;
  border: 1px solid transparent;
}
.history-item:hover { background: #f8fafc; border-color: #e2e8f0; }
.history-title { font-size: 0.8125rem; font-weight: 600; color: #1e293b; margin-bottom: 0.25rem; }
.history-meta { font-size: 0.6875rem; color: #94a3b8; display: flex; gap: 0.75rem; }
.history-delete {
  position: absolute; right: 0.5rem; top: 50%; transform: translateY(-50%);
  width: 1.5rem; height: 1.5rem; border: none; background: transparent;
  color: #cbd5e1; cursor: pointer; opacity: 0; transition: all 0.2s;
  display: flex; align-items: center; justify-content: center;
}
.history-item:hover .history-delete { opacity: 1; }
.history-delete:hover { color: #ef4444; }
.history-delete svg { width: 0.875rem; height: 0.875rem; }

/* ===== 审批卡片 ===== */
.approval-role-bar {
  display: flex; align-items: center; justify-content: space-between;
  padding: 0.5rem 0.75rem; margin-bottom: 0.75rem;
  background: linear-gradient(135deg, #f8fafc, #f1f5f9);
  border-radius: 0.625rem;
  border: 1px solid #e2e8f0;
}
.role-indicator {
  display: flex; align-items: center; gap: 0.375rem;
}
.role-dot {
  width: 0.5rem; height: 0.5rem; border-radius: 50%;
  animation: role-pulse 2s infinite;
}
.role-indicator.admin .role-dot { background: #10b981; box-shadow: 0 0 6px rgba(16, 185, 129, 0.4); }
.role-indicator.user .role-dot { background: #6366f1; box-shadow: 0 0 6px rgba(99, 102, 241, 0.4); }
@keyframes role-pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.6; transform: scale(0.85); }
}
.role-text {
  font-size: 0.6875rem; font-weight: 700;
  letter-spacing: 0.5px; text-transform: uppercase;
}
.role-indicator.admin .role-text { color: #059669; }
.role-indicator.user .role-text { color: #6366f1; }
.role-hint {
  font-size: 0.5625rem; color: #94a3b8; font-style: italic;
}
.can-approve-badge {
  padding: 0.0625rem 0.375rem; border-radius: 0.25rem;
  font-size: 0.5625rem; font-weight: 600;
  background: linear-gradient(135deg, #dcfce7, #bbf7d0); color: #15803d;
  border: 1px solid rgba(34, 197, 94, 0.2);
}
.approval-requester {
  display: flex; align-items: center; gap: 0.5rem;
  padding: 0.5rem 0.625rem; margin: 0.375rem 0;
  background: #fafbfc; border-radius: 0.5rem;
  border: 1px solid #f3f4f6;
}
.requester-avatar {
  width: 1.75rem; height: 1.75rem; flex-shrink: 0;
  background: #f1f5f9; border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  overflow: hidden;
}
.requester-avatar svg { width: 1.25rem; height: 1.25rem; }
.requester-info {
  display: flex; flex-direction: column; flex: 1; gap: 0.0625rem;
}
.requester-name {
  font-size: 0.75rem; font-weight: 600; color: #1e293b;
}
.requester-label {
  font-size: 0.5625rem; color: #94a3b8; text-transform: uppercase;
  letter-spacing: 0.3px;
}
.requester-time {
  font-size: 0.5625rem; color: #94a3b8; flex-shrink: 0;
}
.approval-card {
  padding: 0.875rem; border: 1px solid #e5e7eb; border-radius: 0.75rem;
  margin-bottom: 0.625rem; background: #fff;
}
.approval-card-header { display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.375rem; }
.approval-card-desc { font-size: 0.75rem; color: #6b7280; margin-bottom: 0.5rem; }
.approval-card-detail {
  padding: 0.5rem; margin-bottom: 0.625rem;
  background: #f9fafb; border-radius: 0.375rem;
  border: 1px solid #f3f4f6;
}
.detail-row {
  display: flex; gap: 0.375rem; font-size: 0.6875rem; line-height: 1.6;
}
.detail-label { color: #6b7280; font-weight: 500; flex-shrink: 0; min-width: 4rem; }
.detail-value { color: #1f2937; font-family: 'Fira Code', monospace; font-size: 0.6875rem; word-break: break-all; }
.detail-value.param-json { font-size: 0.625rem; color: #6366f1; }
.approval-card-actions { display: flex; gap: 0.5rem; }
.approve-btn, .reject-btn {
  flex: 1; padding: 0.4rem; border: none; border-radius: 0.5rem;
  font-size: 0.75rem; font-weight: 600; cursor: pointer; transition: all 0.2s;
  display: flex; align-items: center; justify-content: center; gap: 0.25rem;
}
.action-icon { width: 0.75rem; height: 0.75rem; }
.approve-btn { background: #dcfce7; color: #16a34a; }
.approve-btn:hover { background: #16a34a; color: #fff; }
.reject-btn { background: #fee2e2; color: #dc2626; }
.reject-btn:hover { background: #dc2626; color: #fff; }
.cancel-btn {
  padding: 0.375rem 0.75rem; border: 1px solid #e2e8f0; border-radius: 0.5rem;
  background: #fff; color: #6b7280; font-size: 0.6875rem; font-weight: 500;
  cursor: pointer; transition: all 0.2s;
}
.cancel-btn:hover { border-color: #f87171; color: #dc2626; background: #fff5f5; }
.approval-status-hint {
  flex: 1; display: flex; align-items: center; gap: 0.25rem;
  font-size: 0.6875rem; color: #94a3b8; font-style: italic;
}
.approval-status-hint.self-hint { color: #f59e0b; }
.status-icon { width: 0.75rem; height: 0.75rem; }

/* ===== 输入区 ===== */
.panel-footer { padding: 0.75rem 1rem; border-top: 1px solid #f1f5f9; background: #fff; }
.input-wrapper {
  display: flex; align-items: flex-end; gap: 0.5rem;
  background: #f8fafc; border: 1.5px solid #e2e8f0; border-radius: 0.875rem;
  padding: 0.5rem 0.5rem 0.5rem 1rem;
  transition: border-color 0.2s, box-shadow 0.2s;
}
.input-wrapper:focus-within {
  border-color: #6366f1;
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
}
.chat-input {
  flex: 1; border: none; background: transparent; outline: none;
  font-size: 0.8125rem; line-height: 1.5; color: #1e293b;
  resize: none; max-height: 7.5rem; font-family: inherit;
}
.chat-input::placeholder { color: #94a3b8; }
.send-btn {
  width: 2.25rem; height: 2.25rem; flex-shrink: 0;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  border: none; border-radius: 0.625rem; color: #fff;
  cursor: pointer; display: flex; align-items: center; justify-content: center;
  transition: all 0.2s;
}
.send-btn:hover:not(:disabled) { transform: scale(1.05); box-shadow: 0 2px 8px rgba(99, 102, 241, 0.4); }
.send-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.send-btn svg { width: 1rem; height: 1rem; }
.input-hint {
  font-size: 0.625rem; color: #94a3b8; text-align: center; margin-top: 0.375rem;
  display: flex; align-items: center; justify-content: center; gap: 0.25rem;
}
.hint-sep { color: #d1d5db; }
.hint-general { color: #6366f1; }
.hint-platform { color: #0ea5e9; }
.current-model-hint {
  color: #6366f1; font-weight: 600;
}

/* ===== 意图标签（大厂风格） ===== */
.intent-tag {
  display: inline-flex; align-items: center; gap: 0.3rem;
  padding: 0.25rem 0.625rem; border-radius: 1rem;
  font-size: 0.625rem; font-weight: 600;
  margin-bottom: 0.5rem; line-height: 1;
  backdrop-filter: blur(8px);
  box-shadow: 0 1px 3px rgba(0,0,0,0.06);
}
.intent-icon { font-size: 0.75rem; line-height: 1; }
.intent-label { white-space: nowrap; letter-spacing: 0.3px; }
.intent-tools {
  font-weight: 400; opacity: 0.7; font-family: 'Fira Code', monospace;
  font-size: 0.5625rem; max-width: 10rem; overflow: hidden;
  text-overflow: ellipsis; white-space: nowrap;
}
.intent-tag.general {
  background: linear-gradient(135deg, #f1f5f9, #e2e8f0); color: #475569;
}
.intent-tag.platform {
  background: linear-gradient(135deg, #dbeafe, #bfdbfe); color: #1e40af;
  border: 1px solid rgba(59, 130, 246, 0.15);
}
.intent-tag.approval {
  background: linear-gradient(135deg, #fef3c7, #fde68a); color: #92400e;
  border: 1px solid rgba(245, 158, 11, 0.2);
}
.intent-tag.danger {
  background: linear-gradient(135deg, #fee2e2, #fecaca); color: #991b1b;
  border: 1px solid rgba(239, 68, 68, 0.2);
}
.intent-tag.error {
  background: linear-gradient(135deg, #fee2e2, #fecaca); color: #dc2626;
  border: 1px solid rgba(239, 68, 68, 0.2);
}

/* ===== 错误消息样式 ===== */
.msg-bubble.msg-error {
  border-left: 3px solid #ef4444;
}
.error-detail {
  margin-top: 0.5rem;
  border-top: 1px dashed #e5e7eb;
  padding-top: 0.375rem;
}
.error-detail-header {
  display: flex; align-items: center; gap: 0.25rem;
  cursor: pointer; font-size: 0.6875rem; color: #6b7280;
  user-select: none;
}
.error-detail-header:hover { color: #374151; }
.error-detail-icon { font-size: 0.5rem; }
.error-detail-body {
  margin-top: 0.25rem;
  padding: 0.375rem 0.5rem;
  background: #fef2f2;
  border-radius: 0.375rem;
  font-size: 0.625rem;
  font-family: 'Fira Code', monospace;
}
.error-field {
  display: flex; gap: 0.375rem; margin-bottom: 0.125rem;
}
.error-label {
  color: #9ca3af; flex-shrink: 0;
}
.error-value {
  color: #374151; word-break: break-all;
}
.error-tip {
  margin-top: 0.25rem;
  color: #9ca3af;
  font-style: italic;
  font-size: 0.5625rem;
}

/* ===== 滚动条美化 ===== */
.panel-body::-webkit-scrollbar { width: 4px; }
.panel-body::-webkit-scrollbar-track { background: transparent; }
.panel-body::-webkit-scrollbar-thumb { background: #cbd5e1; border-radius: 2px; }
.panel-body::-webkit-scrollbar-thumb:hover { background: #94a3b8; }

/* ===== 响应式 ===== */
@media (max-width: 768px) {
  .ai-panel { width: calc(100vw - 2rem); right: 1rem; bottom: 5.5rem; height: 70vh; }
  .ai-fab { bottom: 1rem; right: 1rem; }
}
</style>
