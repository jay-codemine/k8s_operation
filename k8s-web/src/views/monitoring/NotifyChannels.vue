<template>
  <div class="notify-channels-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h3>通知渠道</h3>
        <span class="header-desc">配置钉钉、飞书、企业微信等告警通知渠道，支持分级推送与限流</span>
      </div>
      <div class="header-actions">
        <button class="btn-action btn-batch-del" :disabled="selectedIds.length === 0" @click="batchDeleteVisible = true">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/></svg>
          批量删除<span v-if="selectedIds.length">({{ selectedIds.length }})</span>
        </button>
        <button class="btn-action btn-batch-update" :disabled="selectedIds.length === 0" @click="batchUpdateVisible = true">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
          批量更新<span v-if="selectedIds.length">({{ selectedIds.length }})</span>
        </button>
        <button class="btn-primary" @click="openDialog()">
          <span>+</span> 新增渠道
        </button>
      </div>
    </div>

    <!-- 渠道类型快捷入口 -->
    <div class="channel-type-bar">
      <button v-for="t in channelTypes" :key="t.value"
        :class="['type-chip', { active: filters.type === t.value }]"
        @click="filters.type = filters.type === t.value ? '' : t.value; loadList()">
        <span class="chip-icon">{{ t.icon }}</span>
        <span>{{ t.label }}</span>
        <span class="chip-count" v-if="typeCounts[t.value]">{{ typeCounts[t.value] }}</span>
      </button>
    </div>

    <!-- 渠道卡片列表 -->
    <div class="channels-grid" v-if="list.length">
      <div class="channel-card" v-for="ch in paginatedList" :key="ch.id" :class="{ disabled: !ch.enabled, selected: selectedIds.includes(ch.id) }">
        <div class="card-select">
          <label class="table-checkbox" @click.stop>
            <input type="checkbox" :checked="selectedIds.includes(ch.id)" @change="toggleSelect(ch.id)" />
            <span class="checkmark"></span>
          </label>
        </div>
        <div class="card-header">
          <div class="card-type-icon" :class="ch.type">
            {{ getTypeIcon(ch.type) }}
          </div>
          <div class="card-title-area">
            <h4 class="card-title">{{ ch.name }}</h4>
            <span class="card-type-label">{{ getTypeLabel(ch.type) }}</span>
          </div>
          <div class="card-actions">
            <label class="toggle-switch mini">
              <input type="checkbox" :checked="ch.enabled" @change="toggleEnabled(ch)" />
              <span class="toggle-slider"></span>
            </label>
          </div>
        </div>
        <div class="card-body">
          <p class="card-desc" v-if="ch.description">{{ ch.description }}</p>
          <div class="card-meta">
            <span class="meta-item" v-if="ch.webhook_url">
              <span class="meta-icon">🔗</span>
              {{ truncateURL(ch.webhook_url) }}
            </span>
            <span class="meta-item" v-if="ch.at_mobiles">
              <span class="meta-icon">📱</span>
              @{{ ch.at_mobiles.split(',').length }}人
            </span>
            <span class="meta-item" v-if="ch.type === 'dingtalk' && ch.security_keyword">
              <span class="meta-icon">🔑</span>
              {{ ch.security_keyword.split(',').filter(k => k.trim()).length > 1
                ? `关键字 ×${ch.security_keyword.split(',').filter(k => k.trim()).length}`
                : `关键字: ${ch.security_keyword.trim()}` }}
            </span>
            <span class="meta-item" v-if="ch.rate_limit">
              <span class="meta-icon">⚡</span>
              限流{{ ch.rate_limit }}条/分
            </span>
            <span class="meta-item" v-if="ch.send_resolved">
              <span class="meta-icon">✅</span>
              恢复通知
            </span>
          </div>
        </div>
        <div class="card-footer">
          <button class="btn-sm btn-test" @click="doTest(ch)" :disabled="testing === ch.id">
            {{ testing === ch.id ? '发送中...' : '🔔 测试' }}
          </button>
          <button class="btn-sm" @click="openDialog(ch)">✏️ 编辑</button>
          <button class="btn-sm btn-danger" @click="confirmDelete(ch)">🗑️ 删除</button>
        </div>
      </div>
    </div>

    <!-- 分页 -->
    <div class="pagination-bar" v-if="list.length > 0">
      <div class="pagination-info">共 <b>{{ list.length }}</b> 个渠道，当前第 <b>{{ channelPage }}</b> / {{ channelTotalPages }} 页</div>
      <div class="pagination-controls">
        <button class="page-btn" :disabled="channelPage <= 1" @click="channelPage--">‹</button>
        <template v-for="p in channelVisiblePages" :key="p">
          <button v-if="p === '...'" class="page-btn" disabled>...</button>
          <button v-else class="page-btn" :class="{ active: p === channelPage }" @click="channelPage = p">{{ p }}</button>
        </template>
        <button class="page-btn" :disabled="channelPage >= channelTotalPages" @click="channelPage++">›</button>
        <select class="page-size-select" v-model="channelPageSize" @change="channelPage = 1">
          <option :value="6">6条/页</option>
          <option :value="12">12条/页</option>
          <option :value="24">24条/页</option>
        </select>
      </div>
    </div>

    <!-- 批量删除确认 -->
    <div class="modal-overlay" v-if="batchDeleteVisible" @click.self="batchDeleteVisible = false">
      <div class="modal-dialog modal-sm">
        <div class="modal-header"><h3>⚠️ 批量删除确认</h3><button class="modal-close" @click="batchDeleteVisible = false">×</button></div>
        <div class="modal-body">
          <p>确定删除选中的 <b style="color:#dc2626">{{ selectedIds.length }}</b> 个通知渠道？关联的告警规则将不再收到通知。</p>
        </div>
        <div class="modal-footer">
          <button class="btn-outline" @click="batchDeleteVisible = false">取消</button>
          <button class="btn-danger" @click="handleBatchDelete" :disabled="batchDeleting">{{ batchDeleting ? '删除中...' : '确认删除' }}</button>
        </div>
      </div>
    </div>

    <!-- 批量更新弹窗 -->
    <div class="modal-overlay" v-if="batchUpdateVisible" @click.self="batchUpdateVisible = false">
      <div class="modal-dialog modal-sm">
        <div class="modal-header">
          <h3>📝 批量更新通知渠道</h3>
          <button class="modal-close" @click="batchUpdateVisible = false">×</button>
        </div>
        <div class="modal-body">
          <p class="batch-update-hint">对选中的 <b>{{ selectedIds.length }}</b> 个渠道统一修改以下属性（留空不修改）：</p>
          <div class="form-section" style="margin-top: 16px;">
            <div class="form-row">
              <label>启用状态</label>
              <select v-model="batchUpdateForm.enabled" class="batch-field">
                <option value="">不修改</option>
                <option value="true">全部启用</option>
                <option value="false">全部禁用</option>
              </select>
            </div>
            <div class="form-row">
              <label>限流 (条/分钟)</label>
              <input v-model.number="batchUpdateForm.rate_limit" type="number" min="0" max="1000" placeholder="留空不修改" class="batch-field" />
              <span class="form-hint">设为 0 表示不限流</span>
            </div>
            <div class="form-row">
              <label>恢复通知</label>
              <select v-model="batchUpdateForm.send_resolved" class="batch-field">
                <option value="">不修改</option>
                <option value="true">开启恢复通知</option>
                <option value="false">关闭恢复通知</option>
              </select>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-outline" @click="batchUpdateVisible = false">取消</button>
          <button class="btn-primary" @click="handleBatchUpdate" :disabled="batchUpdating">
            {{ batchUpdating ? '更新中...' : '确认更新' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <div class="empty-state" v-if="!list.length">
      <div class="empty-icon">📡</div>
      <h3>暂无通知渠道</h3>
      <p>配置通知渠道后，告警触发时将自动推送到钉钉、飞书等平台</p>
      <div class="quick-add-grid">
        <button class="quick-add-card" v-for="t in channelTypes" :key="t.value" @click="openDialogWithType(t.value)">
          <span class="quick-icon">{{ t.icon }}</span>
          <span class="quick-label">{{ t.label }}</span>
        </button>
      </div>
    </div>

    <!-- 新增/编辑弹窗 -->
    <div class="modal-overlay" v-if="dialogVisible" @click.self="dialogVisible = false">
      <div class="modal-dialog modal-lg">
        <div class="modal-header">
          <h3>{{ editingId ? '编辑通知渠道' : '新增通知渠道' }}</h3>
          <button class="modal-close" @click="dialogVisible = false">×</button>
        </div>
        <div class="modal-body">
          <!-- 基础信息 -->
          <div class="form-section">
            <h4 class="section-title">基础配置</h4>
            <div class="form-row-group">
              <div class="form-row half">
                <label>渠道名称 <span class="required">*</span></label>
                <input v-model="form.name" placeholder="如: 运维钉钉群" />
              </div>
              <div class="form-row half">
                <label>渠道类型 <span class="required">*</span></label>
                <select v-model="form.type" @change="onTypeChange">
                  <option value="">请选择</option>
                  <option v-for="t in channelTypes" :key="t.value" :value="t.value">{{ t.icon }} {{ t.label }}</option>
                </select>
              </div>
            </div>
            <div class="form-row">
              <label>描述</label>
              <input v-model="form.description" placeholder="渠道用途说明" />
            </div>
          </div>

          <!-- Webhook 配置 (钉钉/飞书/企业微信/通用Webhook) -->
          <div class="form-section" v-if="['dingtalk','feishu','webhook','wechat'].includes(form.type)">
            <h4 class="section-title">
              {{ form.type === 'dingtalk' ? '钉钉机器人' : form.type === 'feishu' ? '飞书机器人' : form.type === 'wechat' ? '企业微信' : 'Webhook' }} 配置
            </h4>
            <div class="form-row">
              <label>Webhook URL <span class="required">*</span></label>
              <input v-model="form.webhook_url" :placeholder="form.type === 'feishu' ? 'https://open.feishu.cn/open-apis/bot/v2/hook/xxxxxx' : form.type === 'dingtalk' ? 'https://oapi.dingtalk.com/robot/send?access_token=...' : 'https://...'" />
            </div>
            <!-- 钉钉安全配置 -->
            <div class="form-row" v-if="form.type === 'dingtalk'">
              <label>加签密钥 (Secret)</label>
              <input v-model="form.secret" type="password" placeholder="SEC..." />
              <span class="form-hint">钉钉机器人安全设置中的加签密钥</span>
            </div>
            <div class="form-row" v-if="form.type === 'dingtalk'">
              <label>安全关键字</label>
              <input v-model="form.security_keyword" placeholder="如: prom（多个用英文逗号分隔，如: prom,K8s告警）" />
              <span class="form-hint">
                对应钉钉机器人「自定义关键词」安全设置，可填最多10个（逗号分隔）。
                发送时只要消息包含其中任意一个关键字即可通过校验；若消息中均未出现，系统会自动将第一个注入到消息头部。
              </span>
            </div>
            <div class="form-row-group" v-if="form.type === 'dingtalk'">
              <div class="form-row half">
                <label>@手机号 (逗号分隔)</label>
                <input v-model="form.at_mobiles" placeholder="138xxx,139xxx" />
              </div>
              <div class="form-row half">
                <label class="checkbox-label" style="margin-top: 28px">
                  <input type="checkbox" v-model="form.at_all" /> @所有人
                </label>
              </div>
            </div>
            <!-- 飞书安全配置 -->
            <div class="form-row" v-if="form.type === 'feishu'">
              <label>签名校验密钥 (Secret)</label>
              <input v-model="form.secret" type="password" placeholder="如: abcdefg123..." />
              <span class="form-hint">飞书机器人「安全设置 → 签名校验」中的密钥。启用后发送消息时会附带签名，飞书将验证签名合法性。</span>
            </div>
            <div class="form-row" v-if="form.type === 'feishu'">
              <label>自定义关键词</label>
              <input v-model="form.security_keyword" placeholder="如: 飞书（多个用英文逗号分隔）" />
              <span class="form-hint">
                对应飞书机器人「安全设置 → 自定义关键词」，可填多个（逗号分隔）。
                发送时消息需包含至少一个关键词才能通过安全校验；若消息不含关键词，系统会自动将第一个注入到消息中。
              </span>
            </div>
          </div>

          <!-- 邮件配置 -->
          <div class="form-section" v-if="form.type === 'email'">
            <h4 class="section-title">SMTP 邮件配置</h4>
            <div class="form-row-group">
              <div class="form-row half">
                <label>SMTP 主机 <span class="required">*</span></label>
                <input v-model="form.smtp_host" placeholder="smtp.163.com" />
              </div>
              <div class="form-row half">
                <label>端口</label>
                <input v-model.number="form.smtp_port" type="number" placeholder="465" />
              </div>
            </div>
            <div class="form-row-group">
              <div class="form-row half">
                <label>用户名</label>
                <input v-model="form.smtp_user" placeholder="alert@company.com" />
              </div>
              <div class="form-row half">
                <label>密码</label>
                <input v-model="form.smtp_pass" type="password" placeholder="授权码" />
              </div>
            </div>
            <div class="form-row">
              <label>收件人 (逗号分隔)</label>
              <input v-model="form.smtp_to" placeholder="admin@company.com,ops@company.com" />
            </div>
          </div>

          <!-- 高级配置 -->
          <div class="form-section">
            <h4 class="section-title">高级配置</h4>
            <div class="form-row-group">
              <div class="form-row half">
                <label>限流 (条/分钟)</label>
                <input v-model.number="form.rate_limit" type="number" min="1" max="100" />
                <span class="form-hint">防止告警风暴导致消息轰炸</span>
              </div>
              <div class="form-row half" style="display:flex;flex-direction:column;gap:8px;padding-top:24px">
                <label class="checkbox-label">
                  <input type="checkbox" v-model="form.enabled" /> 立即启用
                </label>
                <label class="checkbox-label">
                  <input type="checkbox" v-model="form.send_resolved" /> 发送恢复通知
                </label>
              </div>
            </div>
            <div class="form-row">
              <div class="template-label-row">
                <label>Prometheus 告警消息模板</label>
                <button class="btn-sm" type="button" @click="useDefaultMsgTemplate">使用推荐模板</button>
              </div>
              <textarea v-model="form.msg_template" rows="8" class="msg-template-editor"
                :placeholder="'留空使用系统默认文案。支持变量: ' + '{' + '{.RuleName}' + '} {' + '{.SeverityText}' + '} {' + '{.StatusText}' + '} {' + '{.Summary}' + '} {' + '{.Description}' + '} {' + '{.Value}' + '} {' + '{.FiredAt}' + '} {' + '{.LabelsText}' + '}'"></textarea>
              <div class="template-vars">
                <span v-for="v in templateVars" :key="v" @click="insertTemplateVar(v)">{{ v }}</span>
              </div>
              <span class="form-hint">可按团队习惯调整标题、字段顺序、标签展示；钉钉关键字会在发送时自动补进消息，避免被机器人安全策略拦截。</span>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-outline" @click="dialogVisible = false">取消</button>
          <button class="btn-primary" @click="submitForm" :disabled="submitting">
            {{ submitting ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 通知模板管理区 -->
    <div class="template-section">
      <div class="section-header">
        <div class="section-header-left">
          <h3>📋 通知模板管理</h3>
          <span class="section-desc">自定义各渠道告警通知的消息格式，支持变量插值与实时预览</span>
        </div>
        <button class="btn-primary" @click="openTplDialog()">
          <span>+</span> 新增模板
        </button>
      </div>

      <!-- 模板类型筛选 -->
      <div class="tpl-filter-bar">
        <button v-for="t in tplTypeFilters" :key="t.value"
          :class="['type-chip', { active: tplFilter.type === t.value }]"
          @click="tplFilter.type = tplFilter.type === t.value ? '' : t.value; loadTemplates()">
          <span class="chip-icon">{{ t.icon }}</span>
          <span>{{ t.label }}</span>
        </button>
      </div>

      <!-- 模板卡片列表 -->
      <div class="template-grid" v-if="templates.length">
        <div class="template-card" :class="tpl.type" v-for="tpl in templates" :key="tpl.id">
          <div class="tpl-card-header">
            <span class="tpl-icon">{{ getTypeIcon(tpl.type) }}</span>
            <span class="tpl-title">{{ tpl.name }}</span>
            <span class="tpl-badge" v-if="tpl.is_default">默认</span>
            <span class="tpl-badge scene-badge">{{ getSceneLabel(tpl.scene) }}</span>
          </div>
          <div class="tpl-preview">
            <div class="msg-bubble" :class="tpl.type + '-bubble'">
              <pre class="tpl-content-preview">{{ truncateContent(tpl.content) }}</pre>
            </div>
          </div>
          <div class="tpl-card-footer">
            <div class="tpl-meta">
              <span class="tpl-meta-item" v-if="tpl.description">{{ tpl.description }}</span>
            </div>
            <div class="tpl-actions">
              <button class="btn-sm btn-test" @click="previewTemplate(tpl)">👁️ 预览</button>
              <button class="btn-sm" @click="openTplDialog(tpl)">✏️ 编辑</button>
              <button class="btn-sm" @click="setDefault(tpl)" v-if="!tpl.is_default">⭐ 设为默认</button>
              <button class="btn-sm btn-danger" @click="deleteTpl(tpl)">🗑️</button>
            </div>
          </div>
        </div>
      </div>

      <!-- 模板空状态 -->
      <div class="tpl-empty" v-else>
        <div class="empty-icon">📝</div>
        <h4>还没有通知模板</h4>
        <p>创建自定义模板，为不同渠道定制专属的告警消息格式</p>
        <button class="btn-primary" @click="openTplDialog()">创建第一个模板</button>
      </div>
    </div>

    <!-- 模板新增/编辑弹窗 -->
    <div class="modal-overlay" v-if="tplDialogVisible" @click.self="tplDialogVisible = false">
      <div class="modal-dialog modal-xl">
        <div class="modal-header">
          <h3>{{ tplEditingId ? '编辑通知模板' : '新增通知模板' }}</h3>
          <button class="modal-close" @click="tplDialogVisible = false">×</button>
        </div>
        <div class="modal-body tpl-editor-body">
          <div class="tpl-editor-left">
            <div class="form-section">
              <h4 class="section-title">基础信息</h4>
              <div class="form-row-group">
                <div class="form-row half">
                  <label>模板名称 <span class="required">*</span></label>
                  <input v-model="tplForm.name" placeholder="如: 钉钉告警模板" />
                </div>
                <div class="form-row half">
                  <label>渠道类型 <span class="required">*</span></label>
                  <select v-model="tplForm.type">
                    <option value="">请选择</option>
                    <option v-for="t in channelTypes" :key="t.value" :value="t.value">{{ t.icon }} {{ t.label }}</option>
                  </select>
                </div>
              </div>
              <div class="form-row-group">
                <div class="form-row half">
                  <label>场景</label>
                  <select v-model="tplForm.scene">
                    <option value="alert">🔥 告警触发</option>
                    <option value="resolved">✅ 告警恢复</option>
                    <option value="test">🔔 测试通知</option>
                  </select>
                </div>
                <div class="form-row half">
                  <label>模板标题</label>
                  <input v-model="tplForm.title" placeholder="可用变量: {' + '{.RuleName}' + '}" />
                </div>
              </div>
              <div class="form-row">
                <label>描述</label>
                <input v-model="tplForm.description" placeholder="模板用途说明" />
              </div>
            </div>
            <div class="form-section">
              <h4 class="section-title">模板内容 <span class="required">*</span></h4>
              <div class="form-row">
                <textarea v-model="tplForm.content" rows="12" class="tpl-content-editor"
                  placeholder="支持 Markdown 格式，可用变量:
{' + '{.RuleName}' + '}  - 规则名称
{' + '{.Severity}' + '}  - 级别 (critical/warning/info)
{' + '{.SeverityText}' + '} - 级别文本 (🟡 P1-Warning)
{' + '{.Status}' + '}    - 状态 (firing/resolved)
{' + '{.StatusText}' + '} - 状态文本
{' + '{.Summary}' + '}   - 摘要
{' + '{.Value}' + '}     - 触发值
{' + '{.FiredAt}' + '}   - 触发时间
{' + '{.Labels}' + '}    - 标签列表
{' + '{.Platform}' + '}  - 平台名称"></textarea>
              </div>
              <div class="form-row-group" style="margin-top:8px">
                <label class="checkbox-label">
                  <input type="checkbox" v-model="tplForm.enabled" /> 启用
                </label>
                <label class="checkbox-label">
                  <input type="checkbox" v-model="tplForm.is_default" /> 设为默认
                </label>
              </div>
            </div>
          </div>
          <div class="tpl-editor-right">
            <div class="preview-panel">
              <div class="preview-header">
                <span>👁️ 实时预览</span>
                <button class="btn-sm" @click="doPreview">刷新预览</button>
              </div>
              <div class="preview-content">
                <div class="msg-bubble" :class="(tplForm.type || 'dingtalk') + '-bubble'">
                  <pre class="preview-rendered">{{ tplPreviewContent || '编辑模板内容后点击"刷新预览"查看效果' }}</pre>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-outline" @click="tplDialogVisible = false">取消</button>
          <button class="btn-primary" @click="submitTplForm" :disabled="tplSubmitting">
            {{ tplSubmitting ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 模板预览弹窗 -->
    <div class="modal-overlay" v-if="tplPreviewVisible" @click.self="tplPreviewVisible = false">
      <div class="modal-dialog modal-lg">
        <div class="modal-header">
          <h3>👁️ 模板预览 - {{ previewingTpl?.name }}</h3>
          <button class="modal-close" @click="tplPreviewVisible = false">×</button>
        </div>
        <div class="modal-body">
          <div class="msg-bubble" :class="(previewingTpl?.type || 'dingtalk') + '-bubble'" style="max-width:100%">
            <pre class="preview-rendered full">{{ tplPreviewResult }}</pre>
          </div>
        </div>
      </div>
    </div>

    <!-- 删除确认 -->
    <div class="modal-overlay" v-if="deleteTarget" @click.self="deleteTarget = null">
      <div class="modal-dialog modal-sm">
        <div class="modal-header"><h3>确认删除</h3><button class="modal-close" @click="deleteTarget = null">×</button></div>
        <div class="modal-body">
          <p>确定删除通知渠道 <b>{{ deleteTarget.name }}</b>？关联的告警规则将不再收到通知。</p>
        </div>
        <div class="modal-footer">
          <button class="btn-outline" @click="deleteTarget = null">取消</button>
          <button class="btn-danger" @click="doDelete">确认删除</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { Message } from '@arco-design/web-vue'
import { useConfirmDialog } from '@/composables/useConfirmDialog'

const { confirm: showConfirm } = useConfirmDialog()
import {
  listNotifyChannels, createNotifyChannel, updateNotifyChannel,
  deleteNotifyChannel, testNotifyChannel, batchDeleteNotifyChannels,
  batchUpdateNotifyChannels,
  listNotifyTemplates, createNotifyTemplate, updateNotifyTemplate,
  deleteNotifyTemplate, previewNotifyTemplate, setDefaultNotifyTemplate,
} from '@/api/monitoring'

const list = ref([])
const dialogVisible = ref(false)
const editingId = ref(null)
const submitting = ref(false)
const deleteTarget = ref(null)
const testing = ref(null)
const filters = reactive({ type: '', keyword: '' })

// 分页
const channelPage = ref(1)
const channelPageSize = ref(6)
const channelTotalPages = computed(() => Math.ceil(list.value.length / channelPageSize.value) || 1)
const paginatedList = computed(() => {
  const start = (channelPage.value - 1) * channelPageSize.value
  return list.value.slice(start, start + channelPageSize.value)
})
const channelVisiblePages = computed(() => {
  const tp = channelTotalPages.value
  if (tp <= 7) return Array.from({ length: tp }, (_, i) => i + 1)
  const curr = channelPage.value
  const pages = [1]
  if (curr > 3) pages.push('...')
  for (let i = Math.max(2, curr - 1); i <= Math.min(tp - 1, curr + 1); i++) pages.push(i)
  if (curr < tp - 2) pages.push('...')
  pages.push(tp)
  return pages
})

// 选择
const selectedIds = ref([])
function toggleSelect(id) {
  const idx = selectedIds.value.indexOf(id)
  if (idx >= 0) selectedIds.value.splice(idx, 1)
  else selectedIds.value.push(id)
}

// 批量删除
const batchDeleteVisible = ref(false)
const batchDeleting = ref(false)
async function handleBatchDelete() {
  if (!selectedIds.value.length) return
  batchDeleting.value = true
  try {
    const res = await batchDeleteNotifyChannels({ ids: selectedIds.value })
    if (res?.code === 0) {
      Message.success(`成功删除 ${res.data?.success || selectedIds.value.length} 个渠道`)
      selectedIds.value = []
      batchDeleteVisible.value = false
      loadList()
    } else {
      Message.error(res?.msg || '批量删除失败')
    }
  } catch (e) {
    Message.error(e?.msg || '操作失败')
  } finally {
    batchDeleting.value = false
  }
}

// 批量更新
const batchUpdateVisible = ref(false)
const batchUpdating = ref(false)
const batchUpdateForm = reactive({ enabled: '', rate_limit: '', send_resolved: '' })
async function handleBatchUpdate() {
  if (!selectedIds.value.length) return
  const payload = { ids: selectedIds.value }
  let hasField = false
  if (batchUpdateForm.enabled !== '') {
    payload.enabled = batchUpdateForm.enabled === 'true'
    hasField = true
  }
  if (batchUpdateForm.rate_limit !== '' && batchUpdateForm.rate_limit !== null) {
    payload.rate_limit = Number(batchUpdateForm.rate_limit)
    hasField = true
  }
  if (batchUpdateForm.send_resolved !== '') {
    payload.send_resolved = batchUpdateForm.send_resolved === 'true'
    hasField = true
  }
  if (!hasField) {
    Message.warning('请至少选择一个要修改的属性')
    return
  }
  batchUpdating.value = true
  try {
    const res = await batchUpdateNotifyChannels(payload)
    if (res?.code === 0) {
      Message.success(`批量更新完成: 成功 ${res.data?.success || 0} 个`)
      selectedIds.value = []
      batchUpdateVisible.value = false
      Object.assign(batchUpdateForm, { enabled: '', rate_limit: '', send_resolved: '' })
      loadList()
    } else {
      Message.error(res?.msg || '批量更新失败')
    }
  } catch (e) {
    Message.error(e?.msg || '操作失败')
  } finally {
    batchUpdating.value = false
  }
}

const channelTypes = [
  { value: 'dingtalk', label: '钉钉', icon: '🔷' },
  { value: 'feishu', label: '飞书', icon: '🟦' },
  { value: 'wechat', label: '企业微信', icon: '🟩' },
  { value: 'webhook', label: 'Webhook', icon: '🔗' },
  { value: 'email', label: '邮件', icon: '📧' },
]

const templateVars = [
  '{{.RuleName}}',
  '{{.StatusText}}',
  '{{.SeverityText}}',
  '{{.Summary}}',
  '{{.Description}}',
  '{{.Value}}',
  '{{.FiredAt}}',
  '{{.ResolvedAt}}',
  '{{.LabelsText}}',
]

const defaultMsgTemplate = `## 🔔 {{.StatusText}} | {{.RuleName}}

> **{{.SeverityText}}**　·　{{.FiredAt}}

---

**📋 摘要**
> {{.Summary}}

{{if .Description}}**📝 详情**
> {{.Description}}

{{end}}{{if .Value}}**📊 监控指标**
- **当前值** \`{{.Value}}\`

{{end}}{{if .LabelsText}}---

{{.LabelsText}}
{{end}}{{if .ResolvedAt}}---
⏱ 触发: {{.FiredAt}}　✅ 恢复: {{.ResolvedAt}}

{{end}}---
🛡 {{.Platform}} 监控平台　·　⏰ {{.FiredAt}}`

const typeCounts = computed(() => {
  const counts = {}
  list.value.forEach(ch => { counts[ch.type] = (counts[ch.type] || 0) + 1 })
  return counts
})

const defaultForm = () => ({
  name: '', type: '', description: '', webhook_url: '', secret: '', security_keyword: '',
  at_mobiles: '', at_all: false, smtp_host: '', smtp_port: 465,
  smtp_user: '', smtp_pass: '', smtp_to: '', msg_template: '',
  enabled: true, send_resolved: true, rate_limit: 10,
})
const form = reactive(defaultForm())

function getTypeIcon(type) {
  return channelTypes.find(t => t.value === type)?.icon || '📡'
}
function getTypeLabel(type) {
  return channelTypes.find(t => t.value === type)?.label || type
}
function truncateURL(url) {
  if (url.length > 50) return url.slice(0, 50) + '...'
  return url
}

function onTypeChange() { /* 切换类型时可清空不相关字段 */ }

function useDefaultMsgTemplate() {
  form.msg_template = defaultMsgTemplate
}

function insertTemplateVar(value) {
  const current = form.msg_template || ''
  form.msg_template = current ? `${current}${value}` : value
}

function openDialog(ch = null) {
  Object.assign(form, defaultForm())
  editingId.value = null
  if (ch) {
    editingId.value = ch.id
    Object.assign(form, ch)
  }
  dialogVisible.value = true
}

function openDialogWithType(type) {
  Object.assign(form, defaultForm())
  form.type = type
  editingId.value = null
  dialogVisible.value = true
}

function confirmDelete(ch) { deleteTarget.value = ch }

async function loadList() {
  try {
    const res = await listNotifyChannels({ ...filters })
    if (res?.code === 0) list.value = res.data?.items || []
  } catch {}
}

async function submitForm() {
  if (!form.name || !form.name.trim()) {
    Message.warning('请填写渠道名称')
    return
  }
  if (!form.type) {
    Message.warning('请选择渠道类型')
    return
  }
  submitting.value = true
  try {
    let res
    if (editingId.value) {
      res = await updateNotifyChannel(editingId.value, form)
    } else {
      res = await createNotifyChannel(form)
    }
    // 检查业务响应码（code 非 0 表示后端返回了错误）
    if (res && res.code !== 0) {
      Message.error(res.msg || (editingId.value ? '更新失败' : '创建失败'))
      return
    }
    Message.success(editingId.value ? '更新成功' : '创建成功')
    dialogVisible.value = false
    loadList()
  } catch (e) {
    // e?.msg: 后端 JSON 错误体；e?.message: axios/网络错误
    Message.error(e?.msg || e?.message || '操作失败，请检查网络或联系管理员')
  } finally {
    submitting.value = false
  }
}

async function doDelete() {
  try {
    await deleteNotifyChannel(deleteTarget.value.id)
    Message.success('删除成功')
    deleteTarget.value = null
    loadList()
  } catch {}
}

async function doTest(ch) {
  testing.value = ch.id
  try {
    const res = await testNotifyChannel(ch.id)
    if (res?.data?.success) {
      Message.success(res.data.message || '测试消息发送成功')
    } else {
      const errMsg = res?.data?.message || '发送失败'
      // 如果是关键字相关错误，给出更明确的提示
      if (errMsg.includes('310000') || errMsg.includes('安全验证') || errMsg.includes('关键字')) {
        Message.error(`钉钉安全校验失败：请确认渠道配置的「安全关键字」与机器人安全设置一致。${errMsg}`)
      } else {
        Message.error(errMsg)
      }
    }
  } catch (e) {
    Message.error('测试请求失败，请检查网络或后端服务')
  } finally {
    testing.value = null
  }
}

async function toggleEnabled(ch) {
  try {
    await updateNotifyChannel(ch.id, { ...ch, enabled: !ch.enabled })
    ch.enabled = !ch.enabled
  } catch {}
}

onMounted(() => { loadList(); loadTemplates() })

// ===== 模板管理 =====
const templates = ref([])
const tplDialogVisible = ref(false)
const tplEditingId = ref(null)
const tplSubmitting = ref(false)
const tplPreviewContent = ref('')
const tplPreviewVisible = ref(false)
const tplPreviewResult = ref('')
const previewingTpl = ref(null)
const tplFilter = reactive({ type: '' })

const tplTypeFilters = [
  { value: '', icon: '📚', label: '全部' },
  { value: 'dingtalk', icon: '🔷', label: '钉钉' },
  { value: 'feishu', icon: '🟦', label: '飞书' },
  { value: 'wechat', icon: '🟩', label: '企业微信' },
  { value: 'webhook', icon: '🔗', label: 'Webhook' },
  { value: 'email', icon: '📧', label: '邮件' },
]

const defaultTplForm = () => ({
  name: '', type: 'dingtalk', scene: 'alert', title: '',
  content: '', description: '', is_default: false, enabled: true,
})
const tplForm = reactive(defaultTplForm())

function getSceneLabel(scene) {
  const map = { alert: '🔥 告警', resolved: '✅ 恢复', test: '🔔 测试' }
  return map[scene] || scene
}

function truncateContent(content) {
  if (!content) return ''
  return content.length > 120 ? content.slice(0, 120) + '...' : content
}

async function loadTemplates() {
  try {
    const res = await listNotifyTemplates({ ...tplFilter, size: 50 })
    if (res?.code === 0) templates.value = res.data?.items || []
  } catch {}
}

function openTplDialog(tpl = null) {
  Object.assign(tplForm, defaultTplForm())
  tplEditingId.value = null
  tplPreviewContent.value = ''
  if (tpl) {
    tplEditingId.value = tpl.id
    Object.assign(tplForm, tpl)
  }
  tplDialogVisible.value = true
}

async function submitTplForm() {
  if (!tplForm.name || !tplForm.type || !tplForm.content) {
    Message.warning('请填写模板名称、类型和内容')
    return
  }
  tplSubmitting.value = true
  try {
    if (tplEditingId.value) {
      await updateNotifyTemplate(tplEditingId.value, tplForm)
      Message.success('更新成功')
    } else {
      await createNotifyTemplate(tplForm)
      Message.success('创建成功')
    }
    tplDialogVisible.value = false
    loadTemplates()
  } catch (e) {
    Message.error(e?.msg || '操作失败')
  } finally {
    tplSubmitting.value = false
  }
}

async function deleteTpl(tpl) {
  const ok = await showConfirm({
    title: '确认删除模板',
    type: 'danger',
    details: [
      { label: '模板名称', value: tpl.name },
    ],
    confirmText: '确认删除',
    cancelText: '取消',
  })
  if (!ok) return
  try {
    await deleteNotifyTemplate(tpl.id)
    Message.success('删除成功')
    loadTemplates()
  } catch {}
}

async function setDefault(tpl) {
  try {
    await setDefaultNotifyTemplate(tpl.id)
    Message.success('已设为默认模板')
    loadTemplates()
  } catch {}
}

async function doPreview() {
  try {
    const res = await previewNotifyTemplate(tplForm)
    if (res?.code === 0) tplPreviewContent.value = res.data?.rendered || ''
  } catch {
    tplPreviewContent.value = '预览失败'
  }
}

async function previewTemplate(tpl) {
  previewingTpl.value = tpl
  try {
    const res = await previewNotifyTemplate(tpl)
    tplPreviewResult.value = res?.data?.rendered || tpl.content
  } catch {
    tplPreviewResult.value = tpl.content
  }
  tplPreviewVisible.value = true
}
</script>

<style scoped>
.notify-channels-page { padding: 24px; }

.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.page-header h3 { font-size: 18px; font-weight: 700; color: #1f2937; margin: 0; }
.header-left { display: flex; align-items: baseline; gap: 12px; }
.header-desc { font-size: 13px; color: #9ca3af; }
.header-actions { display: flex; gap: 10px; align-items: center; }
.btn-action { display: inline-flex; align-items: center; gap: 6px; padding: 8px 14px; border-radius: 8px; font-size: 13px; font-weight: 500; border: 1.5px solid #e2e8f0; background: #fff; cursor: pointer; transition: all 0.2s; }
.btn-action:disabled { opacity: 0.45; cursor: not-allowed; }
.btn-batch-del { border-color: #dc2626; color: #dc2626; }
.btn-batch-del:hover:not(:disabled) { background: #fef2f2; }
.btn-batch-update { border-color: #d97706; color: #d97706; }
.btn-batch-update:hover:not(:disabled) { background: #fffbeb; }
.btn-primary { padding: 8px 18px; background: #4f46e5; color: #fff; border: none; border-radius: 8px; font-size: 13px; font-weight: 500; cursor: pointer; transition: background 0.2s; }
.btn-primary:hover { background: #4338ca; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }

/* Card selection */
.card-select { position: absolute; top: 12px; right: 12px; z-index: 2; }
.channel-card { position: relative; }
.channel-card.selected { border-color: #4f46e5; box-shadow: 0 0 0 2px rgba(79,70,229,0.12); }
.table-checkbox { position: relative; display: inline-flex; cursor: pointer; }
.table-checkbox input { position: absolute; opacity: 0; width: 0; height: 0; }
.checkmark { width: 16px; height: 16px; border: 1.5px solid #d1d5db; border-radius: 4px; background: #fff; display: flex; align-items: center; justify-content: center; transition: all 0.15s; }
.table-checkbox input:checked + .checkmark { background: linear-gradient(135deg, #4f46e5, #7c3aed); border-color: transparent; }
.table-checkbox input:checked + .checkmark::after { content: '✓'; color: #fff; font-size: 11px; font-weight: 700; }

/* Pagination */
.pagination-bar { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; background: #fff; border-radius: 12px; border: 1px solid #e8ecf0; margin-top: 16px; }
.pagination-info { font-size: 13px; color: #6b7280; }
.pagination-info b { color: #1e293b; }
.pagination-controls { display: flex; align-items: center; gap: 4px; }
.page-btn { min-width: 32px; height: 32px; padding: 0 8px; border: 1px solid #e5e7eb; border-radius: 6px; background: #fff; font-size: 13px; color: #374151; cursor: pointer; display: flex; align-items: center; justify-content: center; transition: all 0.15s; }
.page-btn:hover:not(:disabled):not(.active) { border-color: #4f46e5; color: #4f46e5; }
.page-btn.active { background: linear-gradient(135deg, #4f46e5, #6d28d9); color: #fff; border-color: transparent; font-weight: 600; }
.page-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.page-size-select { margin-left: 12px; padding: 6px 10px; border: 1px solid #e5e7eb; border-radius: 6px; font-size: 13px; background: #fff; }

/* 类型筛选条 */
.channel-type-bar { display: flex; gap: 8px; margin-bottom: 20px; flex-wrap: wrap; }
.type-chip { display: flex; align-items: center; gap: 6px; padding: 6px 14px; border: 1px solid #e5e7eb; border-radius: 20px; background: #fff; font-size: 13px; color: #4b5563; cursor: pointer; transition: all 0.2s; }
.type-chip:hover { border-color: #4f46e5; color: #4f46e5; }
.type-chip.active { border-color: #4f46e5; background: #eef2ff; color: #4f46e5; font-weight: 500; }
.chip-icon { font-size: 16px; }
.chip-count { background: #e5e7eb; padding: 1px 6px; border-radius: 10px; font-size: 11px; font-weight: 600; }
.type-chip.active .chip-count { background: #c7d2fe; color: #4f46e5; }

/* 渠道卡片网格 */
.channels-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(360px, 1fr)); gap: 16px; }
.channel-card { background: #fff; border-radius: 12px; border: 1px solid #e8ecf0; padding: 20px; transition: all 0.2s; display: flex; flex-direction: column; }
.channel-card:hover { box-shadow: 0 4px 16px rgba(0,0,0,0.08); transform: translateY(-1px); }
.channel-card.disabled { opacity: 0.6; }
.card-header { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; }
.card-type-icon { width: 40px; height: 40px; border-radius: 10px; display: flex; align-items: center; justify-content: center; font-size: 20px; background: #f3f4f6; }
.card-type-icon.dingtalk { background: #e6f4ff; }
.card-type-icon.feishu { background: #e6f7ff; }
.card-type-icon.wechat { background: #f0fdf4; }
.card-type-icon.webhook { background: #fef3c7; }
.card-type-icon.email { background: #fce7f3; }
.card-title-area { flex: 1; }
.card-title { font-size: 15px; font-weight: 600; color: #1f2937; margin: 0; }
.card-type-label { font-size: 12px; color: #9ca3af; }
.card-body { flex: 1; margin-bottom: 14px; }
.card-desc { font-size: 13px; color: #6b7280; margin: 0 0 10px; }
.card-meta { display: flex; flex-wrap: wrap; gap: 10px; }
.meta-item { font-size: 12px; color: #6b7280; display: flex; align-items: center; gap: 4px; }
.meta-icon { font-size: 14px; }
.card-footer { display: flex; gap: 8px; border-top: 1px solid #f3f4f6; padding-top: 14px; }
.btn-sm { padding: 5px 12px; border: 1px solid #e5e7eb; border-radius: 6px; background: #fff; font-size: 12px; color: #4b5563; cursor: pointer; transition: all 0.2s; }
.btn-sm:hover { border-color: #4f46e5; color: #4f46e5; }
.btn-sm:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-test { border-color: #10b981; color: #10b981; }
.btn-test:hover { background: #f0fdf4; }
.btn-sm.btn-danger { color: #ef4444; border-color: #fca5a5; }
.btn-sm.btn-danger:hover { background: #fef2f2; }

/* Toggle 开关 */
.toggle-switch { position: relative; display: inline-block; width: 36px; height: 20px; }
.toggle-switch.mini { width: 32px; height: 18px; }
.toggle-switch input { opacity: 0; width: 0; height: 0; }
.toggle-slider { position: absolute; cursor: pointer; inset: 0; background: #d1d5db; border-radius: 20px; transition: 0.2s; }
.toggle-slider::before { content: ''; position: absolute; width: 14px; height: 14px; bottom: 3px; left: 3px; background: #fff; border-radius: 50%; transition: 0.2s; }
.toggle-switch.mini .toggle-slider::before { width: 12px; height: 12px; }
.toggle-switch input:checked + .toggle-slider { background: #4f46e5; }
.toggle-switch input:checked + .toggle-slider::before { transform: translateX(16px); }
.toggle-switch.mini input:checked + .toggle-slider::before { transform: translateX(14px); }

/* 空状态 */
.empty-state { text-align: center; padding: 60px 20px; }
.empty-icon { font-size: 48px; margin-bottom: 16px; }
.empty-state h3 { font-size: 18px; color: #374151; margin-bottom: 8px; }
.empty-state p { font-size: 14px; color: #9ca3af; margin-bottom: 24px; }
.quick-add-grid { display: flex; gap: 12px; justify-content: center; flex-wrap: wrap; }
.quick-add-card { display: flex; flex-direction: column; align-items: center; gap: 8px; padding: 16px 24px; border: 2px dashed #e5e7eb; border-radius: 12px; background: #fff; cursor: pointer; transition: all 0.2s; }
.quick-add-card:hover { border-color: #4f46e5; background: #f5f3ff; }
.quick-icon { font-size: 28px; }
.quick-label { font-size: 13px; color: #4b5563; font-weight: 500; }

/* 弹窗 */
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 1000; }
.modal-dialog { background: #fff; border-radius: 16px; width: 90%; max-height: 85vh; display: flex; flex-direction: column; box-shadow: 0 20px 60px rgba(0,0,0,0.2); }
.modal-lg { max-width: 640px; }
.modal-sm { max-width: 420px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 20px 24px; border-bottom: 1px solid #f3f4f6; }
.modal-header h3 { font-size: 16px; font-weight: 600; margin: 0; }
.modal-close { width: 32px; height: 32px; border: none; background: #f3f4f6; border-radius: 8px; font-size: 18px; cursor: pointer; display: flex; align-items: center; justify-content: center; }
.modal-close:hover { background: #e5e7eb; }
.modal-body { padding: 24px; overflow-y: auto; flex: 1; }
.modal-footer { display: flex; justify-content: flex-end; gap: 12px; padding: 16px 24px; border-top: 1px solid #f3f4f6; }
.btn-outline { padding: 8px 18px; border: 1px solid #d1d5db; border-radius: 8px; background: #fff; font-size: 13px; color: #4b5563; cursor: pointer; }
.btn-outline:hover { border-color: #9ca3af; }
.btn-danger { padding: 8px 18px; border: none; border-radius: 8px; background: #ef4444; color: #fff; font-size: 13px; cursor: pointer; }
.btn-danger:hover { background: #dc2626; }

/* 表单 */
.form-section { margin-bottom: 20px; }
.section-title { font-size: 14px; font-weight: 600; color: #374151; margin: 0 0 12px; padding-bottom: 8px; border-bottom: 1px solid #f3f4f6; }
.form-row { margin-bottom: 14px; }
.form-row label { display: block; font-size: 13px; font-weight: 500; color: #374151; margin-bottom: 5px; }
.form-row input, .form-row select, .form-row textarea { width: 100%; padding: 8px 12px; border: 1px solid #d1d5db; border-radius: 8px; font-size: 13px; background: #fff; transition: border-color 0.2s; box-sizing: border-box; }
.form-row input:focus, .form-row select:focus, .form-row textarea:focus { outline: none; border-color: #4f46e5; box-shadow: 0 0 0 3px rgba(79,70,229,0.1); }
.form-row textarea { resize: vertical; font-family: 'SF Mono', monospace; }
.form-row-group { display: flex; gap: 16px; }
.form-row.half { flex: 1; }
.form-hint { font-size: 12px; color: #9ca3af; margin-top: 4px; display: block; }
.required { color: #ef4444; }
.checkbox-label { display: flex; align-items: center; gap: 6px; font-size: 13px; color: #4b5563; cursor: pointer; }
.checkbox-label input { width: auto; }
.template-label-row { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 6px; }
.template-label-row label { margin-bottom: 0; }
.msg-template-editor { min-height: 180px; line-height: 1.65; font-size: 12.5px; }
.template-vars { display: flex; flex-wrap: wrap; gap: 6px; margin-top: 8px; }
.template-vars span { padding: 3px 8px; border: 1px solid #dbe2ea; border-radius: 6px; background: #f8fafc; color: #475569; font-family: 'SF Mono', 'Fira Code', monospace; font-size: 11px; cursor: pointer; }
.template-vars span:hover { border-color: #4f46e5; color: #4f46e5; background: #eef2ff; }

/* 通知模板管理区 */
.template-section { margin-top: 32px; border-top: 1px solid #e5e7eb; padding-top: 28px; }
.section-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 16px; }
.section-header-left { flex: 1; }
.section-header h3 { font-size: 17px; font-weight: 700; color: #1f2937; margin: 0 0 4px; }
.section-desc { font-size: 13px; color: #9ca3af; }
.tpl-filter-bar { display: flex; gap: 8px; margin-bottom: 16px; flex-wrap: wrap; }
.template-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(380px, 1fr)); gap: 16px; }
.template-card { background: #fff; border-radius: 14px; border: 1px solid #e8ecf0; overflow: hidden; transition: all 0.2s; display: flex; flex-direction: column; }
.template-card:hover { box-shadow: 0 6px 24px rgba(0,0,0,0.08); transform: translateY(-2px); }
.template-card.dingtalk { border-top: 3px solid #2196f3; }
.template-card.feishu { border-top: 3px solid #3370ff; }
.template-card.wechat { border-top: 3px solid #07c160; }
.template-card.webhook { border-top: 3px solid #f59e0b; }
.template-card.email { border-top: 3px solid #ec4899; }
.tpl-card-header { display: flex; align-items: center; gap: 8px; padding: 16px 20px 10px; }
.tpl-icon { font-size: 20px; }
.tpl-title { font-size: 14px; font-weight: 600; color: #1f2937; flex: 1; }
.tpl-badge { padding: 2px 10px; border-radius: 12px; font-size: 11px; font-weight: 500; background: #f3f4f6; color: #6b7280; }
.tpl-badge.scene-badge { background: #eef2ff; color: #4f46e5; }
.tpl-preview { padding: 0 20px 12px; flex: 1; }
.msg-bubble { border-radius: 10px; padding: 14px; font-size: 12px; line-height: 1.7; }
.dingtalk-bubble { background: linear-gradient(135deg, #f0f9ff 0%, #e6f7ff 100%); border: 1px solid #bae7ff; }
.feishu-bubble { background: linear-gradient(135deg, #f5f7ff 0%, #eef1ff 100%); border: 1px solid #d6e4ff; }
.wechat-bubble { background: linear-gradient(135deg, #f0fff4 0%, #e6ffed 100%); border: 1px solid #b7eb8f; }
.webhook-bubble { background: linear-gradient(135deg, #fffbeb 0%, #fef3c7 100%); border: 1px solid #fde68a; }
.email-bubble { background: linear-gradient(135deg, #fdf2f8 0%, #fce7f3 100%); border: 1px solid #fbcfe8; }
.tpl-content-preview { margin: 0; white-space: pre-wrap; word-break: break-all; font-size: 12px; color: #374151; font-family: 'SF Mono', 'Fira Code', monospace; max-height: 100px; overflow: hidden; }
.tpl-card-footer { padding: 12px 20px 16px; border-top: 1px solid #f3f4f6; }
.tpl-meta { margin-bottom: 8px; }
.tpl-meta-item { font-size: 12px; color: #9ca3af; }
.tpl-actions { display: flex; gap: 6px; flex-wrap: wrap; }
.tpl-empty { text-align: center; padding: 48px 20px; background: #fafbfc; border-radius: 12px; border: 2px dashed #e5e7eb; }
.tpl-empty h4 { font-size: 16px; color: #374151; margin: 12px 0 6px; }
.tpl-empty p { font-size: 13px; color: #9ca3af; margin-bottom: 16px; }

/* 模板编辑器弹窗 */
.modal-xl { max-width: 960px; }
.tpl-editor-body { display: flex; gap: 20px; }
.tpl-editor-left { flex: 1; min-width: 0; }
.tpl-editor-right { width: 340px; flex-shrink: 0; }
.tpl-content-editor { font-family: 'SF Mono', 'Fira Code', monospace; font-size: 12.5px; line-height: 1.6; min-height: 200px; }
.preview-panel { background: #f9fafb; border-radius: 12px; border: 1px solid #e5e7eb; overflow: hidden; height: 100%; display: flex; flex-direction: column; }
.preview-header { display: flex; justify-content: space-between; align-items: center; padding: 12px 16px; background: #fff; border-bottom: 1px solid #e5e7eb; font-size: 13px; font-weight: 600; color: #374151; }
.preview-content { padding: 16px; flex: 1; overflow-y: auto; }
.preview-rendered { margin: 0; white-space: pre-wrap; word-break: break-all; font-size: 12px; color: #374151; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; line-height: 1.8; }
.preview-rendered.full { font-size: 13px; }

/* 批量更新 */
.batch-update-hint { margin: 0; font-size: 14px; color: #4b5563; }
.batch-update-hint b { color: #4f46e5; }
.batch-field { width: 100%; padding: 8px 12px; border: 1px solid #d1d5db; border-radius: 8px; font-size: 13px; background: #fff; transition: border-color 0.2s; box-sizing: border-box; }
.batch-field:focus { outline: none; border-color: #4f46e5; box-shadow: 0 0 0 3px rgba(79,70,229,0.1); }
</style>
