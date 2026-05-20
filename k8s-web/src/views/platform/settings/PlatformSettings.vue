<template>
  <div class="settings-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-content">
        <div class="header-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="3"/>
            <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/>
          </svg>
        </div>
        <div class="header-text">
          <h1>系统设置</h1>
          <p>管理平台的全局配置、安全策略和通知设置</p>
        </div>
      </div>
      <div class="header-actions">
        <button class="btn-secondary" @click="resetSettings">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="1 4 1 10 7 10"/>
            <path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/>
          </svg>
          重置默认
        </button>
        <button class="btn-primary" @click="saveSettings" :disabled="!hasChanges">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"/>
            <polyline points="17 21 17 13 7 13 7 21"/>
            <polyline points="7 3 7 8 15 8"/>
          </svg>
          保存设置
        </button>
      </div>
    </div>

    <!-- 主体内容 -->
    <div class="settings-body">
      <!-- 左侧导航 -->
      <div class="settings-nav">
        <div 
          v-for="section in sections" 
          :key="section.id"
          class="nav-item"
          :class="{ active: activeSection === section.id }"
          @click="activeSection = section.id"
        >
          <div class="nav-icon" v-html="section.icon"></div>
          <div class="nav-content">
            <span class="nav-title">{{ section.title }}</span>
            <span class="nav-desc">{{ section.desc }}</span>
          </div>
          <div class="nav-arrow">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="9 18 15 12 9 6"/>
            </svg>
          </div>
        </div>
      </div>

      <!-- 右侧设置区域 -->
      <div class="settings-content">
        <!-- 基础设置 -->
        <div v-show="activeSection === 'basic'" class="settings-section">
          <div class="section-header">
            <h2>基础设置</h2>
            <p>配置平台的基本行为和默认选项</p>
          </div>
          
          <div class="setting-group">
            <div class="setting-item">
              <div class="setting-info">
                <div class="setting-icon blue">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>
                    <polyline points="9 22 9 12 15 12 15 22"/>
                  </svg>
                </div>
                <div class="setting-text">
                  <label>默认进入页</label>
                  <span>用户登录后默认跳转的页面</span>
                </div>
              </div>
              <div class="setting-control">
                <select v-model="settings.defaultPage">
                  <option value="/dashboard">仪表盘</option>
                  <option value="/clusters">集群管理</option>
                  <option value="/platform/health">平台健康</option>
                </select>
              </div>
            </div>

            <div class="setting-item">
              <div class="setting-info">
                <div class="setting-icon purple">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/>
                    <line x1="3" y1="9" x2="21" y2="9"/>
                    <line x1="9" y1="21" x2="9" y2="9"/>
                  </svg>
                </div>
                <div class="setting-text">
                  <label>默认集群</label>
                  <span>进入集群相关页面时的默认选择</span>
                </div>
              </div>
              <div class="setting-control">
                <select v-model="settings.defaultCluster">
                  <option value="auto">自动（上次使用）</option>
                  <option value="first">第一个可用集群</option>
                </select>
              </div>
            </div>

            <div class="setting-item">
              <div class="setting-info">
                <div class="setting-icon green">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="12" cy="12" r="10"/>
                    <line x1="2" y1="12" x2="22" y2="12"/>
                    <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/>
                  </svg>
                </div>
                <div class="setting-text">
                  <label>界面语言</label>
                  <span>选择平台显示语言</span>
                </div>
              </div>
              <div class="setting-control">
                <select v-model="settings.language">
                  <option value="zh-CN">简体中文</option>
                  <option value="en-US">English</option>
                </select>
              </div>
            </div>

            <div class="setting-item">
              <div class="setting-info">
                <div class="setting-icon orange">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="12" cy="12" r="10"/>
                    <polyline points="12 6 12 12 16 14"/>
                  </svg>
                </div>
                <div class="setting-text">
                  <label>时区设置</label>
                  <span>影响日志和告警的时间显示</span>
                </div>
              </div>
              <div class="setting-control">
                <select v-model="settings.timezone">
                  <option value="Asia/Shanghai">Asia/Shanghai (UTC+8)</option>
                  <option value="UTC">UTC (UTC+0)</option>
                  <option value="America/New_York">America/New_York (UTC-5)</option>
                </select>
              </div>
            </div>
          </div>
        </div>

        <!-- 安全设置 -->
        <div v-show="activeSection === 'security'" class="settings-section">
          <div class="section-header">
            <h2>安全设置</h2>
            <p>配置账户安全策略和访问控制</p>
          </div>

          <div class="setting-group">
            <div class="setting-item">
              <div class="setting-info">
                <div class="setting-icon red">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/>
                    <path d="M7 11V7a5 5 0 0 1 10 0v4"/>
                  </svg>
                </div>
                <div class="setting-text">
                  <label>会话超时</label>
                  <span>用户无操作后自动登出的时间</span>
                </div>
              </div>
              <div class="setting-control">
                <select v-model="settings.sessionTimeout">
                  <option value="30">30 分钟</option>
                  <option value="60">1 小时</option>
                  <option value="120">2 小时</option>
                  <option value="480">8 小时</option>
                  <option value="1440">24 小时</option>
                </select>
              </div>
            </div>

            <div class="setting-item">
              <div class="setting-info">
                <div class="setting-icon blue">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
                  </svg>
                </div>
                <div class="setting-text">
                  <label>双因素认证</label>
                  <span>强制用户使用 2FA 登录</span>
                </div>
              </div>
              <div class="setting-control">
                <label class="toggle-switch">
                  <input type="checkbox" v-model="settings.enable2FA">
                  <span class="toggle-slider"></span>
                </label>
              </div>
            </div>

            <div class="setting-item">
              <div class="setting-info">
                <div class="setting-icon purple">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
                    <circle cx="9" cy="7" r="4"/>
                    <path d="M23 21v-2a4 4 0 0 0-3-3.87"/>
                    <path d="M16 3.13a4 4 0 0 1 0 7.75"/>
                  </svg>
                </div>
                <div class="setting-text">
                  <label>密码强度要求</label>
                  <span>设置密码复杂度规则</span>
                </div>
              </div>
              <div class="setting-control">
                <select v-model="settings.passwordPolicy">
                  <option value="low">低（至少 6 位）</option>
                  <option value="medium">中（至少 8 位，含字母数字）</option>
                  <option value="high">高（至少 12 位，含特殊字符）</option>
                </select>
              </div>
            </div>

            <div class="setting-item">
              <div class="setting-info">
                <div class="setting-icon green">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                    <polyline points="14 2 14 8 20 8"/>
                    <line x1="16" y1="13" x2="8" y2="13"/>
                    <line x1="16" y1="17" x2="8" y2="17"/>
                    <polyline points="10 9 9 9 8 9"/>
                  </svg>
                </div>
                <div class="setting-text">
                  <label>审计日志保留</label>
                  <span>审计日志的保留时间</span>
                </div>
              </div>
              <div class="setting-control">
                <select v-model="settings.auditRetention">
                  <option value="7">7 天</option>
                  <option value="30">30 天</option>
                  <option value="90">90 天</option>
                  <option value="365">1 年</option>
                </select>
              </div>
            </div>
          </div>
        </div>

        <!-- 告警设置 -->
        <div v-show="activeSection === 'alert'" class="settings-section">
          <div class="section-header">
            <h2>告警设置</h2>
            <p>配置资源使用告警阈值</p>
          </div>

          <div class="setting-group">
            <div class="setting-item threshold-item">
              <div class="setting-info">
                <div class="setting-icon orange">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <rect x="4" y="4" width="16" height="16" rx="2" ry="2"/>
                    <rect x="9" y="9" width="6" height="6"/>
                    <line x1="9" y1="1" x2="9" y2="4"/>
                    <line x1="15" y1="1" x2="15" y2="4"/>
                    <line x1="9" y1="20" x2="9" y2="23"/>
                    <line x1="15" y1="20" x2="15" y2="23"/>
                    <line x1="20" y1="9" x2="23" y2="9"/>
                    <line x1="20" y1="14" x2="23" y2="14"/>
                    <line x1="1" y1="9" x2="4" y2="9"/>
                    <line x1="1" y1="14" x2="4" y2="14"/>
                  </svg>
                </div>
                <div class="setting-text">
                  <label>CPU 使用率告警</label>
                  <span>超过阈值时触发告警</span>
                </div>
              </div>
              <div class="setting-control threshold-control">
                <div class="threshold-input">
                  <input type="range" v-model="settings.cpuThreshold" min="50" max="100" step="5">
                  <span class="threshold-value" :class="getThresholdClass(settings.cpuThreshold)">
                    {{ settings.cpuThreshold }}%
                  </span>
                </div>
                <div class="threshold-bar">
                  <div class="threshold-fill" :style="{ width: settings.cpuThreshold + '%' }" :class="getThresholdClass(settings.cpuThreshold)"></div>
                </div>
              </div>
            </div>

            <div class="setting-item threshold-item">
              <div class="setting-info">
                <div class="setting-icon blue">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <rect x="2" y="7" width="20" height="15" rx="2" ry="2"/>
                    <polyline points="17 2 12 7 7 2"/>
                  </svg>
                </div>
                <div class="setting-text">
                  <label>内存使用率告警</label>
                  <span>超过阈值时触发告警</span>
                </div>
              </div>
              <div class="setting-control threshold-control">
                <div class="threshold-input">
                  <input type="range" v-model="settings.memThreshold" min="50" max="100" step="5">
                  <span class="threshold-value" :class="getThresholdClass(settings.memThreshold)">
                    {{ settings.memThreshold }}%
                  </span>
                </div>
                <div class="threshold-bar">
                  <div class="threshold-fill" :style="{ width: settings.memThreshold + '%' }" :class="getThresholdClass(settings.memThreshold)"></div>
                </div>
              </div>
            </div>

            <div class="setting-item threshold-item">
              <div class="setting-info">
                <div class="setting-icon purple">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <ellipse cx="12" cy="5" rx="9" ry="3"/>
                    <path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"/>
                    <path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"/>
                  </svg>
                </div>
                <div class="setting-text">
                  <label>磁盘使用率告警</label>
                  <span>超过阈值时触发告警</span>
                </div>
              </div>
              <div class="setting-control threshold-control">
                <div class="threshold-input">
                  <input type="range" v-model="settings.diskThreshold" min="50" max="100" step="5">
                  <span class="threshold-value" :class="getThresholdClass(settings.diskThreshold)">
                    {{ settings.diskThreshold }}%
                  </span>
                </div>
                <div class="threshold-bar">
                  <div class="threshold-fill" :style="{ width: settings.diskThreshold + '%' }" :class="getThresholdClass(settings.diskThreshold)"></div>
                </div>
              </div>
            </div>

            <div class="setting-item">
              <div class="setting-info">
                <div class="setting-icon green">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M22 17H2a3 3 0 0 0 3-3V9a7 7 0 0 1 14 0v5a3 3 0 0 0 3 3zm-8.27 4a2 2 0 0 1-3.46 0"/>
                  </svg>
                </div>
                <div class="setting-text">
                  <label>告警静默期</label>
                  <span>相同告警的重复通知间隔</span>
                </div>
              </div>
              <div class="setting-control">
                <select v-model="settings.alertSilence">
                  <option value="5">5 分钟</option>
                  <option value="15">15 分钟</option>
                  <option value="30">30 分钟</option>
                  <option value="60">1 小时</option>
                </select>
              </div>
            </div>
          </div>
        </div>

        <!-- 通知设置 -->
        <div v-show="activeSection === 'notification'" class="settings-section">
          <div class="section-header">
            <h2>通知设置</h2>
            <p>配置告警和事件的通知方式</p>
          </div>

          <div class="setting-group">
            <div class="setting-item">
              <div class="setting-info">
                <div class="setting-icon blue">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/>
                    <polyline points="22,6 12,13 2,6"/>
                  </svg>
                </div>
                <div class="setting-text">
                  <label>邮件通知</label>
                  <span>通过邮件发送告警通知</span>
                </div>
              </div>
              <div class="setting-control">
                <label class="toggle-switch">
                  <input type="checkbox" v-model="settings.enableEmail">
                  <span class="toggle-slider"></span>
                </label>
              </div>
            </div>

            <div class="setting-item sub-item" v-show="settings.enableEmail">
              <div class="setting-info">
                <div class="setting-text">
                  <label>SMTP 服务器</label>
                </div>
              </div>
              <div class="setting-control">
                <input type="text" v-model="settings.smtpServer" placeholder="smtp.example.com:587">
              </div>
            </div>

            <div class="setting-item">
              <div class="setting-info">
                <div class="setting-icon green">
                  <svg viewBox="0 0 24 24" fill="currentColor">
                    <path d="M20.94 11.06c-.51-3.32-3.37-5.94-6.94-5.94-2.95 0-5.47 1.75-6.61 4.24C4.58 9.79 2.5 12.14 2.5 15c0 3.31 2.69 6 6 6h11c2.76 0 5-2.24 5-5 0-2.64-2.05-4.78-4.56-4.94z"/>
                  </svg>
                </div>
                <div class="setting-text">
                  <label>钉钉通知</label>
                  <span>通过钉钉机器人推送告警</span>
                </div>
              </div>
              <div class="setting-control">
                <label class="toggle-switch">
                  <input type="checkbox" v-model="settings.enableDingTalk">
                  <span class="toggle-slider"></span>
                </label>
              </div>
            </div>

            <div class="setting-item sub-item" v-show="settings.enableDingTalk">
              <div class="setting-info">
                <div class="setting-text">
                  <label>Webhook URL</label>
                </div>
              </div>
              <div class="setting-control">
                <input type="text" v-model="settings.dingTalkWebhook" placeholder="https://oapi.dingtalk.com/robot/send?access_token=xxx">
              </div>
            </div>

            <div class="setting-item">
              <div class="setting-info">
                <div class="setting-icon purple">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/>
                    <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/>
                  </svg>
                </div>
                <div class="setting-text">
                  <label>Webhook 通知</label>
                  <span>发送到自定义 Webhook 端点</span>
                </div>
              </div>
              <div class="setting-control">
                <label class="toggle-switch">
                  <input type="checkbox" v-model="settings.enableWebhook">
                  <span class="toggle-slider"></span>
                </label>
              </div>
            </div>
          </div>
        </div>

        <!-- 关于 -->
        <div v-show="activeSection === 'about'" class="settings-section">
          <div class="section-header">
            <h2>关于平台</h2>
            <p>系统信息和版本详情</p>
          </div>

          <div class="about-card">
            <div class="about-logo">
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64">
                <defs>
                  <linearGradient id="about-gradient" x1="0%" y1="0%" x2="100%" y2="100%">
                    <stop offset="0%" style="stop-color:#326CE5"/>
                    <stop offset="100%" style="stop-color:#54A3FF"/>
                  </linearGradient>
                </defs>
                <circle cx="32" cy="32" r="30" fill="url(#about-gradient)"/>
                <g fill="#fff" transform="translate(12,12) scale(1.25)">
                  <polygon points="16,0 20,12 32,12 22,20 26,32 16,24 6,32 10,20 0,12 12,12"/>
                </g>
              </svg>
            </div>
            <div class="about-info">
              <h3>K8s Operation Platform</h3>
              <p class="about-desc">企业级多集群 Kubernetes 管理平台</p>
              <div class="about-version">
                <span class="version-badge">v2.0.0</span>
                <span class="version-date">2026-03-04</span>
              </div>
            </div>
          </div>

          <div class="info-grid">
            <div class="info-card">
              <div class="info-label">前端版本</div>
              <div class="info-value">Vue 3.5.13</div>
            </div>
            <div class="info-card">
              <div class="info-label">后端版本</div>
              <div class="info-value">Go 1.21</div>
            </div>
            <div class="info-card">
              <div class="info-label">数据库</div>
              <div class="info-value">MySQL 8.0</div>
            </div>
            <div class="info-card">
              <div class="info-label">K8s 支持</div>
              <div class="info-value">v1.25+</div>
            </div>
          </div>

          <div class="about-links">
            <a href="https://github.com" target="_blank" class="about-link">
              <svg viewBox="0 0 24 24" fill="currentColor">
                <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
              </svg>
              GitHub
            </a>
            <a href="#" class="about-link">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                <polyline points="14 2 14 8 20 8"/>
                <line x1="16" y1="13" x2="8" y2="13"/>
                <line x1="16" y1="17" x2="8" y2="17"/>
              </svg>
              文档
            </a>
            <a href="#" class="about-link">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"/>
                <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"/>
                <line x1="12" y1="17" x2="12.01" y2="17"/>
              </svg>
              帮助
            </a>
          </div>
        </div>
      </div>
    </div>

    <!-- 保存成功提示 -->
    <div class="toast" :class="{ show: showToast, error: toastType === 'error' }">
      <svg v-if="toastType === 'success'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
        <polyline points="22 4 12 14.01 9 11.01"/>
      </svg>
      <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="12" cy="12" r="10"/>
        <line x1="15" y1="9" x2="9" y2="15"/>
        <line x1="9" y1="9" x2="15" y2="15"/>
      </svg>
      {{ toastMessage }}
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import http from '@/api/http'
import { useConfirmDialog } from '@/composables/useConfirmDialog'

const { confirm: showConfirm } = useConfirmDialog()

const activeSection = ref('basic')
const showToast = ref(false)
const loading = ref(false)
const toastMessage = ref('设置已保存')
const toastType = ref('success') // success, error

const sections = [
  {
    id: 'basic',
    title: '基础设置',
    desc: '默认页面、语言、时区',
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>'
  },
  {
    id: 'security',
    title: '安全设置',
    desc: '密码策略、会话管理',
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>'
  },
  {
    id: 'alert',
    title: '告警设置',
    desc: '资源阈值、告警规则',
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 17H2a3 3 0 0 0 3-3V9a7 7 0 0 1 14 0v5a3 3 0 0 0 3 3zm-8.27 4a2 2 0 0 1-3.46 0"/></svg>'
  },
  {
    id: 'notification',
    title: '通知设置',
    desc: '邮件、钉钉、Webhook',
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/></svg>'
  },
  {
    id: 'about',
    title: '关于平台',
    desc: '版本信息、帮助文档',
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>'
  }
]

// 默认设置结构（与后端 API 结构对应）
const defaultSettings = {
  // 基础设置
  defaultPage: '/clusters',
  defaultCluster: 'auto',
  language: 'zh-CN',
  timezone: 'Asia/Shanghai',
  // 安全设置
  sessionTimeout: 120,
  enable2FA: false,
  passwordPolicy: 'medium',
  auditRetention: 30,
  // 告警设置
  cpuThreshold: 80,
  memThreshold: 80,
  diskThreshold: 85,
  alertSilence: 15,
  // 通知设置
  enableEmail: false,
  smtpServer: '',
  enableDingTalk: false,
  dingTalkWebhook: '',
  enableWebhook: false,
  webhookUrl: ''
}

const settings = ref({ ...defaultSettings })
const originalSettings = ref({ ...defaultSettings })

const hasChanges = computed(() => {
  return JSON.stringify(settings.value) !== JSON.stringify(originalSettings.value)
})

const getThresholdClass = (value) => {
  if (value >= 90) return 'danger'
  if (value >= 80) return 'warning'
  return 'normal'
}

// 从后端响应转换为前端设置格式
const apiToSettings = (apiData) => {
  return {
    // 基础设置
    defaultPage: apiData.basic?.default_page || defaultSettings.defaultPage,
    defaultCluster: apiData.basic?.default_cluster || defaultSettings.defaultCluster,
    language: apiData.basic?.language || defaultSettings.language,
    timezone: apiData.basic?.timezone || defaultSettings.timezone,
    // 安全设置
    sessionTimeout: apiData.security?.session_timeout || defaultSettings.sessionTimeout,
    enable2FA: apiData.security?.enable_2fa || defaultSettings.enable2FA,
    passwordPolicy: apiData.security?.password_policy || defaultSettings.passwordPolicy,
    auditRetention: apiData.security?.audit_retention || defaultSettings.auditRetention,
    // 告警设置
    cpuThreshold: apiData.alert?.cpu_threshold || defaultSettings.cpuThreshold,
    memThreshold: apiData.alert?.mem_threshold || defaultSettings.memThreshold,
    diskThreshold: apiData.alert?.disk_threshold || defaultSettings.diskThreshold,
    alertSilence: apiData.alert?.alert_silence || defaultSettings.alertSilence,
    // 通知设置
    enableEmail: apiData.notification?.enable_email || defaultSettings.enableEmail,
    smtpServer: apiData.notification?.smtp_server || defaultSettings.smtpServer,
    enableDingTalk: apiData.notification?.enable_dingtalk || defaultSettings.enableDingTalk,
    dingTalkWebhook: apiData.notification?.dingtalk_webhook || defaultSettings.dingTalkWebhook,
    enableWebhook: apiData.notification?.enable_webhook || defaultSettings.enableWebhook,
    webhookUrl: apiData.notification?.webhook_url || defaultSettings.webhookUrl
  }
}

// 从前端设置转换为后端 API 格式
const settingsToApi = (s) => {
  return {
    basic: {
      default_page: s.defaultPage,
      default_cluster: s.defaultCluster,
      language: s.language,
      timezone: s.timezone
    },
    security: {
      session_timeout: Number(s.sessionTimeout),
      enable_2fa: s.enable2FA,
      password_policy: s.passwordPolicy,
      audit_retention: Number(s.auditRetention)
    },
    alert: {
      cpu_threshold: Number(s.cpuThreshold),
      mem_threshold: Number(s.memThreshold),
      disk_threshold: Number(s.diskThreshold),
      alert_silence: Number(s.alertSilence)
    },
    notification: {
      enable_email: s.enableEmail,
      smtp_server: s.smtpServer,
      enable_dingtalk: s.enableDingTalk,
      dingtalk_webhook: s.dingTalkWebhook,
      enable_webhook: s.enableWebhook,
      webhook_url: s.webhookUrl
    },
    about: {} // about 不需要保存
  }
}

// 加载设置
const loadSettings = async () => {
  loading.value = true
  try {
    const res = await http.get('/api/v1/platform/settings')
    if (res.code === 0 && res.data) {
      settings.value = apiToSettings(res.data)
      originalSettings.value = { ...settings.value }
    }
  } catch (e) {
    console.error('加载设置失败', e)
    showNotification('加载设置失败', 'error')
  } finally {
    loading.value = false
  }
}

// 保存设置
const saveSettings = async () => {
  loading.value = true
  try {
    const payload = settingsToApi(settings.value)
    const res = await http.put('/api/v1/platform/settings', payload)
    if (res.code === 0) {
      originalSettings.value = { ...settings.value }
      showNotification('设置已保存', 'success')
    } else {
      showNotification(res.msg || '保存失败', 'error')
    }
  } catch (e) {
    console.error('保存设置失败', e)
    showNotification('保存设置失败', 'error')
  } finally {
    loading.value = false
  }
}

// 重置设置
const resetSettings = async () => {
  const ok = await showConfirm({
    title: '确认重置设置',
    content: '所有配置将恢复为默认值，当前修改将丢失。',
    type: 'warning',
    confirmText: '确认重置',
    cancelText: '取消',
  })
  if (!ok) return
  
  loading.value = true
  try {
    const res = await http.post('/api/v1/platform/settings/reset')
    if (res.code === 0 && res.data) {
      settings.value = apiToSettings(res.data)
      originalSettings.value = { ...settings.value }
      showNotification('已重置为默认设置', 'success')
    } else {
      showNotification(res.msg || '重置失败', 'error')
    }
  } catch (e) {
    console.error('重置设置失败', e)
    showNotification('重置设置失败', 'error')
  } finally {
    loading.value = false
  }
}

// 显示通知
const showNotification = (message, type = 'success') => {
  toastMessage.value = message
  toastType.value = type
  showToast.value = true
  setTimeout(() => showToast.value = false, 2000)
}

onMounted(() => {
  loadSettings()
})
</script>

<style scoped>
.settings-page {
  min-height: 100%;
  background: #f8fafc;
}

/* 页面头部 */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem 2rem;
  background: linear-gradient(135deg, #1e3a5f 0%, #2d4a6f 100%);
  color: #fff;
}

.header-content {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.header-icon {
  width: 3rem;
  height: 3rem;
  background: rgba(255, 255, 255, 0.15);
  border-radius: 0.75rem;
  display: flex;
  align-items: center;
  justify-content: center;
}

.header-icon svg {
  width: 1.75rem;
  height: 1.75rem;
}

.header-text h1 {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 600;
}

.header-text p {
  margin: 0.25rem 0 0;
  font-size: 0.875rem;
  opacity: 0.8;
}

.header-actions {
  display: flex;
  gap: 0.75rem;
}

.btn-primary, .btn-secondary {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.625rem 1.25rem;
  border: none;
  border-radius: 0.5rem;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-primary {
  background: linear-gradient(135deg, #326ce5 0%, #4a85f0 100%);
  color: #fff;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(50, 108, 229, 0.4);
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-secondary {
  background: rgba(255, 255, 255, 0.15);
  color: #fff;
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.btn-secondary:hover {
  background: rgba(255, 255, 255, 0.25);
}

.btn-primary svg, .btn-secondary svg {
  width: 1rem;
  height: 1rem;
}

/* 主体内容 */
.settings-body {
  display: flex;
  gap: 1.5rem;
  padding: 1.5rem 2rem;
  max-width: 1400px;
  margin: 0 auto;
}

/* 左侧导航 */
.settings-nav {
  width: 280px;
  flex-shrink: 0;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 0.875rem;
  padding: 1rem;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 0.75rem;
  margin-bottom: 0.5rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.nav-item:hover {
  border-color: #326ce5;
  box-shadow: 0 2px 8px rgba(50, 108, 229, 0.1);
}

.nav-item.active {
  background: linear-gradient(135deg, #326ce5 0%, #4a85f0 100%);
  border-color: transparent;
  color: #fff;
}

.nav-icon {
  width: 2.5rem;
  height: 2.5rem;
  background: #f1f5f9;
  border-radius: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.nav-item.active .nav-icon {
  background: rgba(255, 255, 255, 0.2);
}

.nav-icon :deep(svg) {
  width: 1.25rem;
  height: 1.25rem;
  color: #64748b;
}

.nav-item.active .nav-icon :deep(svg) {
  color: #fff;
}

.nav-content {
  flex: 1;
  min-width: 0;
}

.nav-title {
  display: block;
  font-weight: 600;
  font-size: 0.9375rem;
  color: #1e293b;
}

.nav-item.active .nav-title {
  color: #fff;
}

.nav-desc {
  display: block;
  font-size: 0.75rem;
  color: #94a3b8;
  margin-top: 2px;
}

.nav-item.active .nav-desc {
  color: rgba(255, 255, 255, 0.8);
}

.nav-arrow {
  width: 1.25rem;
  height: 1.25rem;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.nav-item:hover .nav-arrow,
.nav-item.active .nav-arrow {
  opacity: 1;
}

.nav-arrow svg {
  width: 100%;
  height: 100%;
  color: #94a3b8;
}

.nav-item.active .nav-arrow svg {
  color: #fff;
}

/* 右侧设置区域 */
.settings-content {
  flex: 1;
  min-width: 0;
}

.settings-section {
  background: #fff;
  border-radius: 1rem;
  border: 1px solid #e2e8f0;
  overflow: hidden;
}

.section-header {
  padding: 1.5rem;
  border-bottom: 1px solid #e2e8f0;
  background: linear-gradient(180deg, #f8fafc 0%, #fff 100%);
}

.section-header h2 {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 600;
  color: #1e293b;
}

.section-header p {
  margin: 0.25rem 0 0;
  font-size: 0.8125rem;
  color: #64748b;
}

.setting-group {
  padding: 0.5rem 0;
}

.setting-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid #f1f5f9;
  transition: background 0.2s ease;
}

.setting-item:hover {
  background: #f8fafc;
}

.setting-item:last-child {
  border-bottom: none;
}

.setting-item.sub-item {
  padding-left: 4.5rem;
  background: #f8fafc;
}

.setting-info {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.setting-icon {
  width: 2.5rem;
  height: 2.5rem;
  border-radius: 0.625rem;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.setting-icon svg {
  width: 1.25rem;
  height: 1.25rem;
}

.setting-icon.blue {
  background: linear-gradient(135deg, #e0f2fe 0%, #bae6fd 100%);
  color: #0284c7;
}

.setting-icon.purple {
  background: linear-gradient(135deg, #f3e8ff 0%, #e9d5ff 100%);
  color: #9333ea;
}

.setting-icon.green {
  background: linear-gradient(135deg, #dcfce7 0%, #bbf7d0 100%);
  color: #16a34a;
}

.setting-icon.orange {
  background: linear-gradient(135deg, #ffedd5 0%, #fed7aa 100%);
  color: #ea580c;
}

.setting-icon.red {
  background: linear-gradient(135deg, #fee2e2 0%, #fecaca 100%);
  color: #dc2626;
}

.setting-text label {
  display: block;
  font-weight: 600;
  font-size: 0.875rem;
  color: #1e293b;
}

.setting-text span {
  display: block;
  font-size: 0.75rem;
  color: #64748b;
  margin-top: 2px;
}

.setting-control select,
.setting-control input[type="text"] {
  min-width: 200px;
  padding: 0.5rem 0.75rem;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  font-size: 0.875rem;
  color: #1e293b;
  background: #fff;
  transition: all 0.2s ease;
}

.setting-control select:focus,
.setting-control input[type="text"]:focus {
  outline: none;
  border-color: #326ce5;
  box-shadow: 0 0 0 3px rgba(50, 108, 229, 0.1);
}

/* Toggle Switch */
.toggle-switch {
  position: relative;
  display: inline-block;
  width: 48px;
  height: 26px;
}

.toggle-switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.toggle-slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: #cbd5e1;
  border-radius: 26px;
  transition: 0.3s;
}

.toggle-slider:before {
  position: absolute;
  content: "";
  height: 20px;
  width: 20px;
  left: 3px;
  bottom: 3px;
  background: #fff;
  border-radius: 50%;
  transition: 0.3s;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.toggle-switch input:checked + .toggle-slider {
  background: linear-gradient(135deg, #326ce5 0%, #4a85f0 100%);
}

.toggle-switch input:checked + .toggle-slider:before {
  transform: translateX(22px);
}

/* Threshold控件 */
.threshold-item .setting-control {
  min-width: 280px;
}

.threshold-control {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.threshold-input {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.threshold-input input[type="range"] {
  flex: 1;
  -webkit-appearance: none;
  height: 6px;
  background: #e2e8f0;
  border-radius: 3px;
  outline: none;
}

.threshold-input input[type="range"]::-webkit-slider-thumb {
  -webkit-appearance: none;
  width: 18px;
  height: 18px;
  background: #326ce5;
  border-radius: 50%;
  cursor: pointer;
  box-shadow: 0 2px 6px rgba(50, 108, 229, 0.4);
}

.threshold-value {
  font-weight: 700;
  font-size: 0.9375rem;
  min-width: 50px;
  text-align: right;
}

.threshold-value.normal { color: #16a34a; }
.threshold-value.warning { color: #ea580c; }
.threshold-value.danger { color: #dc2626; }

.threshold-bar {
  height: 4px;
  background: #e2e8f0;
  border-radius: 2px;
  overflow: hidden;
}

.threshold-fill {
  height: 100%;
  border-radius: 2px;
  transition: width 0.3s ease;
}

.threshold-fill.normal { background: linear-gradient(90deg, #22c55e, #4ade80); }
.threshold-fill.warning { background: linear-gradient(90deg, #f97316, #fb923c); }
.threshold-fill.danger { background: linear-gradient(90deg, #ef4444, #f87171); }

/* 关于页面 */
.about-card {
  display: flex;
  align-items: center;
  gap: 1.5rem;
  padding: 2rem;
  background: linear-gradient(135deg, #f0f9ff 0%, #e0f2fe 100%);
  border-bottom: 1px solid #e2e8f0;
}

.about-logo svg {
  width: 80px;
  height: 80px;
}

.about-info h3 {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 700;
  color: #1e293b;
}

.about-desc {
  margin: 0.25rem 0 0.75rem;
  color: #64748b;
  font-size: 0.9375rem;
}

.about-version {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.version-badge {
  padding: 0.25rem 0.75rem;
  background: linear-gradient(135deg, #326ce5 0%, #4a85f0 100%);
  color: #fff;
  font-size: 0.8125rem;
  font-weight: 600;
  border-radius: 1rem;
}

.version-date {
  font-size: 0.8125rem;
  color: #64748b;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1rem;
  padding: 1.5rem;
}

.info-card {
  padding: 1rem;
  background: #f8fafc;
  border-radius: 0.75rem;
  text-align: center;
}

.info-label {
  font-size: 0.75rem;
  color: #64748b;
  margin-bottom: 0.25rem;
}

.info-value {
  font-size: 1rem;
  font-weight: 600;
  color: #1e293b;
}

.about-links {
  display: flex;
  gap: 1rem;
  padding: 1.5rem;
  border-top: 1px solid #e2e8f0;
}

.about-link {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.625rem 1rem;
  background: #f1f5f9;
  border-radius: 0.5rem;
  color: #475569;
  text-decoration: none;
  font-size: 0.875rem;
  font-weight: 500;
  transition: all 0.2s ease;
}

.about-link:hover {
  background: #e2e8f0;
  color: #1e293b;
}

.about-link svg {
  width: 1rem;
  height: 1rem;
}

/* Toast 提示 */
.toast {
  position: fixed;
  bottom: 2rem;
  left: 50%;
  transform: translateX(-50%) translateY(100px);
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem 1.5rem;
  background: linear-gradient(135deg, #22c55e 0%, #16a34a 100%);
  color: #fff;
  border-radius: 0.75rem;
  box-shadow: 0 10px 40px rgba(34, 197, 94, 0.4);
  opacity: 0;
  transition: all 0.3s ease;
  z-index: 1000;
}

.toast.show {
  transform: translateX(-50%) translateY(0);
  opacity: 1;
}

.toast svg {
  width: 1.25rem;
  height: 1.25rem;
}

.toast.error {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  box-shadow: 0 10px 40px rgba(239, 68, 68, 0.4);
}

/* 响应式 */
@media (max-width: 1024px) {
  .settings-body {
    flex-direction: column;
  }
  
  .settings-nav {
    width: 100%;
    display: flex;
    gap: 0.5rem;
    overflow-x: auto;
    padding-bottom: 0.5rem;
  }
  
  .nav-item {
    flex-shrink: 0;
    margin-bottom: 0;
  }
  
  .nav-desc,
  .nav-arrow {
    display: none;
  }
  
  .info-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 640px) {
  .page-header {
    flex-direction: column;
    gap: 1rem;
    text-align: center;
  }
  
  .header-actions {
    width: 100%;
    justify-content: center;
  }
  
  .setting-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }
  
  .setting-control {
    width: 100%;
  }
  
  .setting-control select,
  .setting-control input[type="text"] {
    width: 100%;
  }
  
  .threshold-item .setting-control {
    min-width: 100%;
  }
}
</style>
