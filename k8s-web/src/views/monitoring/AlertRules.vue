<template>
  <div class="alert-rules-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h3>告警规则</h3>
        <span class="header-desc">配置基于 PromQL 的告警规则，支持多级别通知</span>
      </div>
      <div class="header-actions">
        <button class="btn-action btn-yaml-import" @click="yamlImportVisible = true" title="YAML 批量导入">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
          导入
        </button>
        <button class="btn-action btn-yaml-export" @click="handleExportYAML" title="导出为 YAML">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
          导出
        </button>
        <button class="btn-primary" @click="openDialog()">
          <span>+</span> 新增规则
        </button>
      </div>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar">
      <input v-model="filters.keyword" placeholder="搜索规则名称..." class="search-input" @input="debouncedLoad" />
      <select v-model="filters.severity" @change="loadList" class="filter-select">
        <option value="">全部级别</option>
        <option value="critical">Critical</option>
        <option value="warning">Warning</option>
        <option value="info">Info</option>
      </select>
      <select v-model="filters.group" @change="loadList" class="filter-select">
        <option value="">全部分组</option>
        <option v-for="g in groups" :key="g" :value="g">{{ g }}</option>
      </select>
    </div>

    <!-- 规则列表 -->
    <div class="rules-table-wrapper" v-if="list.length">
      <table class="data-table">
        <thead>
          <tr>
            <th>规则名称</th>
            <th>分组</th>
            <th>级别</th>
            <th>PromQL 表达式</th>
            <th>持续时间</th>
            <th>评估状态</th>
            <th>启用</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="rule in list" :key="rule.id" :class="{ disabled: !rule.enabled }">
            <td class="rule-name">{{ rule.name }}</td>
            <td><span class="group-tag">{{ rule.group }}</span></td>
            <td><span class="severity-badge" :class="rule.severity">{{ rule.severity }}</span></td>
            <td class="expr-cell"><code>{{ truncateExpr(rule.expr) }}</code></td>
            <td>{{ rule.duration }}</td>
            <td>
              <span class="eval-status" :class="rule.last_eval_result || 'unknown'">
                {{ evalStatusMap[rule.last_eval_result] || '未评估' }}
              </span>
            </td>
            <td>
              <label class="toggle-switch">
                <input type="checkbox" :checked="rule.enabled" @change="handleToggle(rule)" />
                <span class="toggle-slider"></span>
              </label>
            </td>
            <td class="action-cell">
              <button class="btn-icon" @click="openDialog(rule)" title="编辑">✏️</button>
              <button class="btn-icon danger" @click="confirmDelete(rule)" title="删除">🗑️</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 空状态 -->
    <div class="empty-state" v-else>
      <div class="empty-icon">📋</div>
      <h3>暂无告警规则</h3>
      <p>创建告警规则后，系统将自动检测异常指标并触发通知</p>
      <div class="template-section">
        <h4>快速创建（常用模板）</h4>
        <div class="template-grid">
          <button class="template-card" v-for="tpl in templates" :key="tpl.name" @click="openDialogFromTemplate(tpl)">
            <span class="tpl-icon">{{ tpl.icon }}</span>
            <span class="tpl-name">{{ tpl.name }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- 规则模板快速创建区（有数据时也显示） -->
    <div class="template-section" v-if="list.length">
      <h4>📌 快速添加常用规则</h4>
      <div class="template-grid">
        <button class="template-card" v-for="tpl in templates" :key="tpl.name" @click="openDialogFromTemplate(tpl)">
          <span class="tpl-icon">{{ tpl.icon }}</span>
          <span class="tpl-name">{{ tpl.name }}</span>
        </button>
      </div>
    </div>

    <!-- 新增/编辑弹窗 -->
    <div class="modal-overlay" v-if="dialogVisible" @click.self="dialogVisible = false">
      <div class="modal-dialog modal-xl">

        <!-- ===== Header ===== -->
        <div class="modal-hd">
          <div class="modal-hd-bar"></div>
          <div class="modal-hd-inner">
            <div class="modal-hd-icon">
              <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#4f46e5" stroke-width="2" stroke-linecap="round"><circle cx="12" cy="12" r="10"/><path d="M12 6v6l4 2"/></svg>
            </div>
            <div>
              <h3 class="modal-hd-title">{{ editingId ? '编辑告警规则' : '新增告警规则' }}</h3>
              <p class="modal-hd-sub">配置 PromQL 触发条件与多渠道通知策略</p>
            </div>
          </div>
          <button class="modal-close-btn" @click="dialogVisible = false">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><path d="M18 6L6 18M6 6l12 12"/></svg>
          </button>
        </div>

        <!-- ===== Body ===== -->
        <div class="modal-bd">

          <!-- 01 基本信息 -->
          <div class="rule-section">
            <div class="rule-section-hd">
              <span class="sec-num">01</span>
              <span class="sec-title">基本信息</span>
            </div>
            <div class="field-grid-2">
              <div class="field-item col-2">
                <label class="field-lbl">规则名称 <em class="req-star">*</em></label>
                <input class="field-ctrl" v-model="form.name" placeholder="如: 节点 CPU 使用率过高" />
              </div>
              <div class="field-item">
                <label class="field-lbl">所属分组</label>
                <input class="field-ctrl" v-model="form.group" placeholder="default" />
              </div>
              <div class="field-item">
                <label class="field-lbl">持续时间 <span class="hint-tag">for</span></label>
                <input class="field-ctrl" v-model="form.duration" placeholder="5m" />
              </div>
            </div>
            <!-- 严重级别卡片 -->
            <div class="field-item" style="margin-top:14px">
              <label class="field-lbl">严重级别 <em class="req-star">*</em></label>
              <div class="sev-picker">
                <div class="sev-card sev-critical" :class="{ 'sev-active': form.severity === 'critical' }" @click="form.severity = 'critical'">
                  <div class="sev-stripe"></div>
                  <div class="sev-content">
                    <div class="sev-top"><span class="sev-dot"></span><span class="sev-name">Critical</span></div>
                    <span class="sev-sub-label">P0 · 立即响应</span>
                  </div>
                  <div class="sev-tick" v-show="form.severity === 'critical'">✓</div>
                </div>
                <div class="sev-card sev-warning" :class="{ 'sev-active': form.severity === 'warning' }" @click="form.severity = 'warning'">
                  <div class="sev-stripe"></div>
                  <div class="sev-content">
                    <div class="sev-top"><span class="sev-dot"></span><span class="sev-name">Warning</span></div>
                    <span class="sev-sub-label">P1 · 关注处理</span>
                  </div>
                  <div class="sev-tick" v-show="form.severity === 'warning'">✓</div>
                </div>
                <div class="sev-card sev-info" :class="{ 'sev-active': form.severity === 'info' }" @click="form.severity = 'info'">
                  <div class="sev-stripe"></div>
                  <div class="sev-content">
                    <div class="sev-top"><span class="sev-dot"></span><span class="sev-name">Info</span></div>
                    <span class="sev-sub-label">P2 · 记录参考</span>
                  </div>
                  <div class="sev-tick" v-show="form.severity === 'info'">✓</div>
                </div>
              </div>
            </div>
          </div>

          <!-- 02 PromQL 表达式 -->
          <div class="rule-section">
            <div class="rule-section-hd">
              <span class="sec-num">02</span>
              <span class="sec-title">PromQL 表达式 <em class="req-star">*</em></span>
            </div>
            <div class="code-editor-wrap">
              <div class="code-editor-topbar">
                <div class="code-win-dots">
                  <span class="dot dot-red"></span><span class="dot dot-yellow"></span><span class="dot dot-green"></span>
                </div>
                <span class="code-lang-badge">PromQL</span>
                <span class="code-editor-tip">支持聚合函数、标签过滤、算术运算</span>
              </div>
              <textarea class="code-textarea" v-model="form.expr" rows="4"
                placeholder='avg(100 - (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > 80'></textarea>
            </div>
          </div>

          <!-- 03 告警消息 -->
          <div class="rule-section">
            <div class="rule-section-hd">
              <span class="sec-num">03</span>
              <span class="sec-title">告警消息</span>
            </div>
            <div class="field-grid-2">
              <div class="field-item col-2">
                <label class="field-lbl">摘要 <span v-pre class="hint-tag">支持 {{ $labels.instance }} 模板变量</span></label>
                <input class="field-ctrl" v-model="form.summary" :placeholder="'如: 节点 {{ $labels.instance }} CPU 使用率 {{ $value }}%'" />
              </div>
              <div class="field-item col-2">
                <label class="field-lbl">描述 <span class="hint-tag">可选</span></label>
                <textarea class="field-textarea" v-model="form.description" rows="2" placeholder="告警详情描述，可包含排查建议..."></textarea>
              </div>
            </div>
          </div>

          <!-- 04 通知与调度 -->
          <div class="rule-section">
            <div class="rule-section-hd">
              <span class="sec-num">04</span>
              <span class="sec-title">通知与调度</span>
            </div>
            <div class="field-grid-2">
              <div class="field-item">
                <label class="field-lbl">通知渠道</label>
                <div class="ch-picker-grid" v-if="notifyChannelList.length">
                  <label v-for="ch in notifyChannelList" :key="ch.id"
                    class="ch-card" :class="{ 'ch-selected': selectedNotifyChannels.includes(channelValue(ch)) }">
                    <input type="checkbox" :value="channelValue(ch)" v-model="selectedNotifyChannels" style="display:none" />
                    <span class="ch-icon">{{ getChannelIcon(ch.type) }}</span>
                    <span class="ch-meta">
                      <b>{{ ch.name }}</b>
                      <small>{{ getChannelLabel(ch.type) }}</small>
                    </span>
                    <span class="ch-check-mark">✓</span>
                  </label>
                </div>
                <div class="ch-empty-tip" v-else>
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#94a3b8" stroke-width="2"><path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9M13.73 21a2 2 0 0 1-3.46 0"/></svg>
                  暂无已启用的通知渠道
                </div>
                <div class="ch-quick-btns">
                  <button class="ch-quick-btn" type="button" @click="openQuickChannel('dingtalk')">+ 钉钉</button>
                  <button class="ch-quick-btn" type="button" @click="openQuickChannel('feishu')">飞书</button>
                  <button class="ch-quick-btn" type="button" @click="openQuickChannel('wechat')">企业微信</button>
                  <button class="ch-quick-btn" type="button" @click="openQuickChannel('webhook')">Webhook</button>
                  <button class="ch-quick-btn ch-quick-muted" type="button" @click="goNotifyChannels">更多配置 →</button>
                </div>
              </div>
              <div class="field-item">
                <label class="field-lbl">评估间隔 <span class="hint-tag">秒/次</span></label>
                <div class="num-field-wrap">
                  <input class="field-ctrl" v-model.number="form.eval_interval" type="number" min="10" max="600" />
                  <span class="num-field-unit">s</span>
                </div>
                <p class="field-hint-text">每隔此时间对 PromQL 执行一次评估</p>
              </div>
            </div>

            <!-- 快速添加渠道面板 -->
            <div class="qc-panel" v-if="quickChannelVisible">
              <div class="qc-panel-bar">
                <div class="qc-bar-left">
                  <span class="qc-tag">快速添加</span>
                  <strong class="qc-title">{{ getChannelLabel(quickChannel.type) }}渠道</strong>
                </div>
                <button type="button" class="qc-dismiss" @click="quickChannelVisible = false">×</button>
              </div>
              <div class="field-grid-2">
                <div class="field-item">
                  <label class="field-lbl">渠道名称 <em class="req-star">*</em></label>
                  <input class="field-ctrl" v-model="quickChannel.name" placeholder="如: 运维钉钉群" />
                </div>
                <div class="field-item">
                  <label class="field-lbl">Webhook URL <em class="req-star">*</em></label>
                  <input class="field-ctrl" v-model="quickChannel.webhook_url" placeholder="https://oapi.dingtalk.com/robot/..." />
                </div>
              </div>
              <div class="field-grid-2" v-if="quickChannel.type === 'dingtalk'">
                <div class="field-item">
                  <label class="field-lbl">加签密钥</label>
                  <input class="field-ctrl" v-model="quickChannel.secret" type="password" placeholder="SECxxx..." />
                </div>
                <div class="field-item">
                  <label class="field-lbl">安全关键字</label>
                  <input class="field-ctrl" v-model="quickChannel.security_keyword" placeholder="如: prom（逗号分隔多个）" />
                </div>
              </div>
              <div class="field-grid-2" v-if="quickChannel.type === 'dingtalk'">
                <div class="field-item">
                  <label class="field-lbl">@手机号</label>
                  <input class="field-ctrl" v-model="quickChannel.at_mobiles" placeholder="138xxx,139xxx（逗号分隔）" />
                </div>
              </div>
              <div class="qc-panel-footer">
                <label class="mini-chk-label" v-if="quickChannel.type === 'dingtalk'">
                  <input type="checkbox" v-model="quickChannel.at_all" /><span>@所有人</span>
                </label>
                <button class="btn-qc-add" type="button" @click="createQuickChannel" :disabled="quickChannelSaving">
                  {{ quickChannelSaving ? '添加中...' : '✓ 添加并选中' }}
                </button>
              </div>
            </div>

            <!-- 立即启用 -->
            <div class="enable-row">
              <label class="enable-label">
                <div class="toggle-sw">
                  <input type="checkbox" v-model="form.enabled" />
                  <span class="toggle-track"></span>
                </div>
                <div class="enable-text">
                  <span class="enable-title">立即启用</span>
                  <span class="enable-sub">保存后规则将立即参与告警评估周期</span>
                </div>
              </label>
            </div>
          </div>

        </div><!-- end modal-bd -->

        <!-- ===== Footer ===== -->
        <div class="modal-ft">
          <span class="modal-ft-meta" v-if="editingId">规则 ID: #{{ editingId }}</span>
          <span v-else></span>
          <div class="modal-ft-btns">
            <button class="btn-ft-cancel" @click="dialogVisible = false">取消</button>
            <button class="btn-ft-save" @click="submitForm" :disabled="submitting">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M19 21H5a2 2 0 01-2-2V5a2 2 0 012-2h11l5 5v11a2 2 0 01-2 2z"/><path d="M17 21v-8H7v8M7 3v5h8"/></svg>
              {{ submitting ? '保存中...' : '保存规则' }}
            </button>
          </div>
        </div>

      </div>
    </div>

    <!-- 删除确认 -->
    <div class="modal-overlay" v-if="deleteTarget" @click.self="deleteTarget = null">
      <div class="modal-dialog modal-sm">
        <div class="modal-header"><h3>确认删除</h3><button class="modal-close" @click="deleteTarget = null">×</button></div>
        <div class="modal-body">
          <p>确定删除告警规则 <b>{{ deleteTarget.name }}</b>？</p>
        </div>
        <div class="modal-footer">
          <button class="btn-outline" @click="deleteTarget = null">取消</button>
          <button class="btn-danger" @click="doDelete">确认删除</button>
        </div>
      </div>
    </div>

    <!-- YAML 批量导入弹窗 -->
    <div class="modal-overlay" v-if="yamlImportVisible" @click.self="yamlImportVisible = false">
      <div class="modal-dialog modal-xl yaml-modal">
        <div class="yaml-modal-header">
          <div class="yaml-header-left">
            <div class="yaml-icon-wrap import">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
            </div>
            <div>
              <h3>YAML 批量导入告警规则</h3>
              <p class="yaml-header-desc">支持 PrometheusRule 格式，一次导入多条规则</p>
            </div>
          </div>
          <button class="modal-close-btn" @click="yamlImportVisible = false">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><path d="M18 6L6 18M6 6l12 12"/></svg>
          </button>
        </div>
        <div class="yaml-modal-body">
          <div class="yaml-toolbar">
            <div class="yaml-toolbar-left">
              <label class="yaml-overwrite-chk">
                <input type="checkbox" v-model="yamlImportOverwrite" />
                <span>同名规则覆盖更新</span>
              </label>
            </div>
            <div class="yaml-toolbar-right">
              <button class="yaml-tpl-btn" @click="fillImportTemplate">填入示例模板</button>
            </div>
          </div>
          <div class="yaml-editor-wrap">
            <div class="yaml-editor-topbar">
              <div class="code-win-dots"><span class="dot dot-red"></span><span class="dot dot-yellow"></span><span class="dot dot-green"></span></div>
              <span class="code-lang-badge">YAML</span>
              <span class="yaml-editor-hint">粘贴 PrometheusRule spec 格式</span>
            </div>
            <textarea class="yaml-textarea" v-model="yamlImportContent" rows="18" placeholder="groups:
  - name: infrastructure
    rules:
      - alert: NodeCPUHigh
        expr: 100 - (avg by(instance)(rate(node_cpu_seconds_total{mode=&quot;idle&quot;}[5m])) * 100) > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: 节点CPU使用率过高"></textarea>
          </div>
          <div class="yaml-import-result" v-if="yamlImportResult">
            <div class="import-result-header">
              <span class="import-result-icon">📊</span>
              <span class="import-result-title">导入结果</span>
            </div>
            <div class="import-result-stats">
              <span class="stat-item success">新建 {{ yamlImportResult.created }}</span>
              <span class="stat-item update">更新 {{ yamlImportResult.updated }}</span>
              <span class="stat-item skip">跳过 {{ yamlImportResult.skipped }}</span>
              <span class="stat-item fail" v-if="yamlImportResult.failed">失败 {{ yamlImportResult.failed }}</span>
            </div>
            <div class="import-result-errors" v-if="yamlImportResult.errors?.length">
              <div v-for="(err, i) in yamlImportResult.errors" :key="i" class="error-line">{{ err }}</div>
            </div>
          </div>
        </div>
        <div class="yaml-modal-footer">
          <button class="btn-outline" @click="yamlImportVisible = false">关闭</button>
          <button class="btn-primary" @click="handleImportYAML" :disabled="yamlImporting || !yamlImportContent.trim()">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
            {{ yamlImporting ? '导入中...' : '执行导入' }}
          </button>
        </div>
      </div>
    </div>

    <!-- YAML 导出/预览弹窗 -->
    <div class="modal-overlay" v-if="yamlExportVisible" @click.self="yamlExportVisible = false">
      <div class="modal-dialog modal-xl yaml-modal">
        <div class="yaml-modal-header">
          <div class="yaml-header-left">
            <div class="yaml-icon-wrap export">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
            </div>
            <div>
              <h3>导出告警规则 YAML</h3>
              <p class="yaml-header-desc">PrometheusRule 兼容格式，可直接用于 K8s 集群或重新导入</p>
            </div>
          </div>
          <button class="modal-close-btn" @click="yamlExportVisible = false">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><path d="M18 6L6 18M6 6l12 12"/></svg>
          </button>
        </div>
        <div class="yaml-modal-body">
          <div class="yaml-editor-wrap export-view">
            <div class="yaml-editor-topbar">
              <div class="code-win-dots"><span class="dot dot-red"></span><span class="dot dot-yellow"></span><span class="dot dot-green"></span></div>
              <span class="code-lang-badge">YAML</span>
              <span class="yaml-editor-hint">可编辑后重新导入</span>
            </div>
            <textarea class="yaml-textarea" v-model="yamlExportContent" rows="22" readonly></textarea>
          </div>
        </div>
        <div class="yaml-modal-footer">
          <button class="btn-outline" @click="yamlExportVisible = false">关闭</button>
          <button class="btn-action btn-yaml-copy" @click="handleCopyYAML">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>
            复制
          </button>
          <button class="btn-primary" @click="handleDownloadYAML">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
            下载 YAML
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Message } from '@arco-design/web-vue'
import {
  listAlertRules, createAlertRule, updateAlertRule, deleteAlertRule,
  toggleAlertRule, getAlertRuleGroups, listNotifyChannels, createNotifyChannel,
  importAlertRulesYAML, exportAlertRulesYAML,
} from '@/api/monitoring'

const router = useRouter()
const list = ref([])
const groups = ref([])
const notifyChannelList = ref([])
const selectedNotifyChannels = ref([])
const dialogVisible = ref(false)
const editingId = ref(null)
const submitting = ref(false)
const deleteTarget = ref(null)
const filters = reactive({ keyword: '', severity: '', group: '' })
const quickChannelVisible = ref(false)
const quickChannelSaving = ref(false)

const channelTypes = [
  { value: 'dingtalk', label: '钉钉', icon: '🔷' },
  { value: 'feishu', label: '飞书', icon: '🟦' },
  { value: 'wechat', label: '企业微信', icon: '🟩' },
  { value: 'webhook', label: 'Webhook', icon: '🔗' },
  { value: 'email', label: '邮件', icon: '📧' },
]

const evalStatusMap = { normal: '正常', firing: '告警中', pending: '待触发', error: '异常', unknown: '未评估' }

const templates = [
  { name: 'CPU > 80%', icon: '💻', expr: '100 - (avg(rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > 80', severity: 'warning', duration: '5m', summary: '集群CPU使用率超过80%' },
  { name: '内存 > 85%', icon: '🧠', expr: 'avg(100 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes * 100)) > 85', severity: 'warning', duration: '5m', summary: '集群内存使用率超过85%' },
  { name: '磁盘 > 90%', icon: '💾', expr: 'avg(100 - (node_filesystem_avail_bytes{mountpoint="/"} / node_filesystem_size_bytes{mountpoint="/"} * 100)) > 90', severity: 'critical', duration: '10m', summary: '磁盘使用率超过90%' },
  { name: 'Pod 重启', icon: '🔄', expr: 'increase(kube_pod_container_status_restarts_total[1h]) > 3', severity: 'warning', duration: '1m', summary: 'Pod 1小时内重启超过3次' },
  { name: '节点宕机', icon: '🖥️', expr: 'up{job="node-exporter"} == 0', severity: 'critical', duration: '2m', summary: '节点 {{ $labels.instance }} 不可达' },
  { name: 'API 5xx', icon: '🌐', expr: 'sum(rate(http_requests_total{status=~"5.."}[5m])) > 1', severity: 'critical', duration: '3m', summary: 'HTTP 5xx 错误率升高' },
]

const defaultForm = () => ({
  name: '', group: 'default', severity: 'warning', expr: '', duration: '5m',
  summary: '', description: '', labels: '', annotations: '', enabled: true,
  notify_channels: '', notify_url: '', eval_interval: 60, datasource_id: 0,
})
const form = reactive(defaultForm())
const quickChannel = reactive({
  name: '',
  type: 'dingtalk',
  description: '',
  webhook_url: '',
  secret: '',
  at_mobiles: '',
  at_all: false,
  enabled: true,
  send_resolved: true,
  rate_limit: 10,
})

let debounceTimer = null
const debouncedLoad = () => { clearTimeout(debounceTimer); debounceTimer = setTimeout(loadList, 300) }

const truncateExpr = (expr) => expr?.length > 60 ? expr.slice(0, 60) + '...' : expr

async function loadList() {
  try {
    const [res, gRes, chRes] = await Promise.all([
      listAlertRules({ page: 1, size: 50, ...filters }),
      getAlertRuleGroups(),
      listNotifyChannels({ size: 100 }),
    ])
    if (res?.code === 0) list.value = res.data?.items || []
    if (gRes?.code === 0) groups.value = gRes.data || []
    if (chRes?.code === 0) notifyChannelList.value = (chRes.data?.items || []).filter(c => c.enabled)
  } catch {}
}

function openDialog(rule = null) {
  if (rule) {
    editingId.value = rule.id
    Object.assign(form, { ...rule })
    selectedNotifyChannels.value = parseNotifyChannels(rule.notify_channels)
  } else {
    editingId.value = null
    Object.assign(form, defaultForm())
    selectedNotifyChannels.value = []
  }
  quickChannelVisible.value = false
  dialogVisible.value = true
}

function openDialogFromTemplate(tpl) {
  editingId.value = null
  Object.assign(form, { ...defaultForm(), name: tpl.name, expr: tpl.expr, severity: tpl.severity, duration: tpl.duration, summary: tpl.summary })
  selectedNotifyChannels.value = []
  quickChannelVisible.value = false
  dialogVisible.value = true
}

async function submitForm() {
  if (!form.name || !form.expr || !form.severity) return alert('请填写必填字段')
  submitting.value = true
  try {
    form.notify_channels = selectedNotifyChannels.value.join(',')
    if (editingId.value) {
      await updateAlertRule(editingId.value, form)
    } else {
      await createAlertRule(form)
    }
    dialogVisible.value = false
    loadList()
  } catch (e) {
    alert('保存失败: ' + (e?.msg || e?.message || ''))
  } finally { submitting.value = false }
}

async function handleToggle(rule) {
  try {
    await toggleAlertRule(rule.id, !rule.enabled)
    loadList()
  } catch {}
}

function confirmDelete(rule) { deleteTarget.value = rule }

async function doDelete() {
  try {
    await deleteAlertRule(deleteTarget.value.id)
    deleteTarget.value = null
    loadList()
  } catch {}
}

onMounted(loadList)

// ==================== YAML 批量导入/导出 ====================
const yamlImportVisible = ref(false)
const yamlExportVisible = ref(false)
const yamlImportContent = ref('')
const yamlExportContent = ref('')
const yamlImportOverwrite = ref(false)
const yamlImporting = ref(false)
const yamlImportResult = ref(null)

const IMPORT_TEMPLATE = `groups:
  - name: infrastructure
    rules:
      - alert: NodeCPUHigh
        expr: 100 - (avg by(instance)(rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > 80
        for: 5m
        labels:
          severity: warning
          monitor_type: node
        annotations:
          summary: "节点 {{ $labels.instance }} CPU使用率过高"
          description: "当前值: {{ $value }}%"

      - alert: NodeMemoryHigh
        expr: (1 - node_memory_MemAvailable_bytes/node_memory_MemTotal_bytes) * 100 > 85
        for: 5m
        labels:
          severity: warning
          monitor_type: node
        annotations:
          summary: "节点内存使用率过高"

      - alert: NodeDiskFull
        expr: (1 - node_filesystem_avail_bytes{mountpoint="/"}/node_filesystem_size_bytes{mountpoint="/"}) * 100 > 90
        for: 10m
        labels:
          severity: critical
          monitor_type: node
        annotations:
          summary: "磁盘使用率超过90%"

  - name: kubernetes
    rules:
      - alert: PodCrashLooping
        expr: increase(kube_pod_container_status_restarts_total[1h]) > 3
        for: 1m
        labels:
          severity: warning
          monitor_type: pod
        annotations:
          summary: "Pod {{ $labels.namespace }}/{{ $labels.pod }} 频繁重启"

      - alert: DeploymentReplicasMismatch
        expr: kube_deployment_status_replicas_unavailable > 0
        for: 5m
        labels:
          severity: critical
          monitor_type: deployment
        annotations:
          summary: "{{ $labels.namespace }}/{{ $labels.deployment }} 副本不足"
`

function fillImportTemplate() {
  yamlImportContent.value = IMPORT_TEMPLATE
  yamlImportResult.value = null
}

async function handleImportYAML() {
  if (!yamlImportContent.value.trim()) return
  yamlImporting.value = true
  yamlImportResult.value = null
  try {
    const res = await importAlertRulesYAML({
      yaml: yamlImportContent.value,
      datasource_id: 1,
      overwrite: yamlImportOverwrite.value,
    })
    if (res?.code === 0) {
      yamlImportResult.value = res.data
      loadList()
      if (res.data.failed === 0) {
        Message.success(res.msg || '导入成功')
      } else {
        Message.warning(res.msg || '部分导入失败')
      }
    } else {
      Message.error(res?.msg || '导入失败')
    }
  } catch (e) {
    Message.error('导入异常: ' + (e?.msg || e?.message || ''))
  } finally {
    yamlImporting.value = false
  }
}

async function handleExportYAML() {
  try {
    const params = {}
    if (filters.group) params.group = filters.group
    const res = await exportAlertRulesYAML(params)
    if (res?.code === 0 && res.data?.yaml) {
      yamlExportContent.value = res.data.yaml
      yamlExportVisible.value = true
    } else {
      Message.warning(res?.msg || '没有可导出的规则')
    }
  } catch (e) {
    Message.error('导出失败: ' + (e?.msg || e?.message || ''))
  }
}

function handleCopyYAML() {
  navigator.clipboard.writeText(yamlExportContent.value).then(() => {
    Message.success('已复制到剪贴板')
  }).catch(() => {
    Message.warning('复制失败，请手动选择复制')
  })
}

function handleDownloadYAML() {
  const blob = new Blob([yamlExportContent.value], { type: 'application/x-yaml;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `alert-rules-${new Date().toISOString().slice(0,10)}.yaml`
  a.click()
  URL.revokeObjectURL(url)
  Message.success('下载已开始')
}

function channelValue(ch) {
  return `${ch.type}:${ch.id}`
}

function parseNotifyChannels(value = '') {
  return String(value)
    .split(',')
    .map(v => v.trim())
    .filter(Boolean)
}

function getChannelIcon(type) {
  return channelTypes.find(t => t.value === type)?.icon || '📡'
}

function getChannelLabel(type) {
  return channelTypes.find(t => t.value === type)?.label || type
}

function openQuickChannel(type) {
  Object.assign(quickChannel, {
    name: '',
    type,
    description: '从告警规则快速创建',
    webhook_url: '',
    secret: '',
    security_keyword: '',
    at_mobiles: '',
    at_all: false,
    enabled: true,
    send_resolved: true,
    rate_limit: 10,
  })
  quickChannel.name = `默认${getChannelLabel(type)}告警`
  quickChannelVisible.value = true
}

async function createQuickChannel() {
  if (!quickChannel.name || !quickChannel.webhook_url) {
    Message.warning('请填写渠道名称和 Webhook URL')
    return
  }
  quickChannelSaving.value = true
  try {
    const res = await createNotifyChannel(quickChannel)
    const created = res?.data
    await loadList()
    if (created?.id) {
      const value = `${created.type}:${created.id}`
      if (!selectedNotifyChannels.value.includes(value)) {
        selectedNotifyChannels.value.push(value)
      }
    }
    quickChannelVisible.value = false
    Message.success('通知渠道已添加')
  } catch (e) {
    Message.error(e?.msg || e?.message || '添加失败')
  } finally {
    quickChannelSaving.value = false
  }
}

function goNotifyChannels() {
  dialogVisible.value = false
  router.push('/monitoring/notify-channels')
}
</script>

<style scoped>
.alert-rules-page { padding: 24px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.header-left h3 { margin: 0; font-size: 20px; color: #1a202c; }
.header-desc { font-size: 13px; color: #718096; }
.btn-primary { background: #4f46e5; color: #fff; border: none; padding: 8px 18px; border-radius: 8px; font-size: 14px; cursor: pointer; font-weight: 500; display: flex; align-items: center; gap: 4px; }
.btn-primary:hover { background: #4338ca; }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }

.filter-bar { display: flex; gap: 12px; margin-bottom: 20px; }
.search-input { flex: 1; max-width: 280px; padding: 8px 14px; border: 1px solid #e2e8f0; border-radius: 8px; font-size: 14px; outline: none; }
.search-input:focus { border-color: #4f46e5; }
.filter-select { padding: 8px 14px; border: 1px solid #e2e8f0; border-radius: 8px; font-size: 14px; background: #fff; }

.rules-table-wrapper { background: #fff; border-radius: 12px; border: 1px solid #e8ecf0; overflow: hidden; }
.data-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.data-table th { background: #f7f8fa; padding: 12px 16px; text-align: left; font-weight: 600; color: #4a5568; border-bottom: 1px solid #e8ecf0; }
.data-table td { padding: 12px 16px; border-bottom: 1px solid #f1f5f9; }
.data-table tr:hover { background: #f7fafc; }
.data-table tr.disabled { opacity: 0.5; }
.rule-name { font-weight: 500; color: #1a202c; }
.group-tag { background: #edf2f7; color: #4a5568; padding: 2px 8px; border-radius: 4px; font-size: 12px; }
.severity-badge { padding: 3px 10px; border-radius: 12px; font-size: 11px; font-weight: 600; text-transform: uppercase; }
.severity-badge.critical { background: #fef2f2; color: #dc2626; }
.severity-badge.warning { background: #fffbeb; color: #d97706; }
.severity-badge.info { background: #eff6ff; color: #2563eb; }
.expr-cell code { font-size: 12px; background: #f7f8fa; padding: 2px 6px; border-radius: 4px; color: #6b7280; }
.eval-status { padding: 3px 8px; border-radius: 4px; font-size: 11px; font-weight: 500; }
.eval-status.normal { background: #ecfdf5; color: #059669; }
.eval-status.firing { background: #fef2f2; color: #dc2626; }
.eval-status.pending { background: #fffbeb; color: #d97706; }
.eval-status.error { background: #fef2f2; color: #dc2626; }
.eval-status.unknown { background: #f3f4f6; color: #6b7280; }

.toggle-switch { position: relative; display: inline-block; width: 36px; height: 20px; }
.toggle-switch input { opacity: 0; width: 0; height: 0; }
.toggle-slider { position: absolute; cursor: pointer; inset: 0; background: #cbd5e1; border-radius: 20px; transition: 0.3s; }
.toggle-slider::before { content: ''; position: absolute; width: 16px; height: 16px; left: 2px; bottom: 2px; background: #fff; border-radius: 50%; transition: 0.3s; }
.toggle-switch input:checked + .toggle-slider { background: #4f46e5; }
.toggle-switch input:checked + .toggle-slider::before { transform: translateX(16px); }

.action-cell { white-space: nowrap; }
.btn-icon { background: none; border: none; font-size: 16px; cursor: pointer; padding: 4px 6px; border-radius: 4px; }
.btn-icon:hover { background: #f1f5f9; }
.btn-icon.danger:hover { background: #fef2f2; }

.template-section { margin-top: 24px; padding: 20px; background: #fff; border-radius: 12px; border: 1px solid #e8ecf0; }
.template-section h4 { margin: 0 0 12px; font-size: 14px; color: #4a5568; }
.template-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(160px, 1fr)); gap: 10px; }
.template-card { display: flex; align-items: center; gap: 8px; padding: 10px 14px; border: 1px solid #e2e8f0; border-radius: 8px; background: #fff; cursor: pointer; font-size: 13px; transition: all 0.15s; }
.template-card:hover { border-color: #4f46e5; background: #f5f3ff; }
.tpl-icon { font-size: 18px; }
.tpl-name { color: #4a5568; font-weight: 500; }

.empty-state { text-align: center; padding: 60px 20px; }
.empty-icon { font-size: 56px; margin-bottom: 12px; }
.empty-state h3 { color: #1a202c; margin: 0 0 8px; }
.empty-state p { color: #718096; margin: 0 0 24px; }

/* Modal */
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 1000; }
.modal-dialog { background: #fff; border-radius: 14px; width: 560px; max-height: 85vh; overflow-y: auto; box-shadow: 0 20px 60px rgba(0,0,0,0.15); }
.modal-lg { width: 680px; }
.modal-sm { width: 420px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 20px 24px; border-bottom: 1px solid #e8ecf0; }
.modal-header h3 { margin: 0; font-size: 18px; }
.modal-close { background: none; border: none; font-size: 24px; cursor: pointer; color: #a0aec0; }
.modal-body { padding: 24px; }
.modal-footer { display: flex; justify-content: flex-end; gap: 10px; padding: 16px 24px; border-top: 1px solid #e8ecf0; }
.form-row { margin-bottom: 16px; }
.form-row label { display: block; font-size: 13px; font-weight: 500; color: #4a5568; margin-bottom: 6px; }
.form-row input, .form-row select, .form-row textarea { width: 100%; padding: 9px 14px; border: 1px solid #e2e8f0; border-radius: 8px; font-size: 14px; outline: none; box-sizing: border-box; font-family: inherit; }
.form-row textarea { resize: vertical; font-family: 'Fira Code', monospace, inherit; }
.form-row input:focus, .form-row select:focus, .form-row textarea:focus { border-color: #4f46e5; box-shadow: 0 0 0 3px rgba(79,70,229,0.1); }
.form-row-group { display: flex; gap: 16px; }
.form-row.half { flex: 1; }
.form-row-inline { display: flex; gap: 24px; margin-top: 8px; }
.checkbox-label { display: flex; align-items: center; gap: 6px; font-size: 14px; color: #4a5568; cursor: pointer; }
.required { color: #ef4444; }
.channel-picker { display: grid; grid-template-columns: repeat(auto-fill, minmax(180px, 1fr)); gap: 8px; }
.channel-option { display: flex; align-items: center; gap: 8px; min-height: 50px; padding: 9px 10px; border: 1px solid #e2e8f0; border-radius: 8px; background: #fff; cursor: pointer; transition: all 0.15s; }
.channel-option:hover { border-color: #a5b4fc; background: #f8fafc; }
.channel-option.selected { border-color: #4f46e5; background: #f5f3ff; box-shadow: 0 0 0 3px rgba(79,70,229,0.08); }
.channel-option input { width: auto; flex-shrink: 0; }
.channel-icon { font-size: 18px; }
.channel-text { min-width: 0; display: flex; flex-direction: column; gap: 2px; }
.channel-text b { font-size: 13px; color: #1f2937; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.channel-text small { color: #718096; font-size: 12px; }
.channel-empty { padding: 12px 14px; border: 1px dashed #cbd5e1; border-radius: 8px; color: #64748b; background: #f8fafc; font-size: 13px; }
.channel-actions { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 8px; }
.link-btn { border: none; background: transparent; color: #4f46e5; padding: 0; font-size: 13px; cursor: pointer; }
.link-btn:hover { text-decoration: underline; }
.link-btn.muted { color: #64748b; }
.quick-channel-panel { margin: 2px 0 16px; padding: 14px; border: 1px solid #c7d2fe; border-radius: 10px; background: #f8f7ff; }
.quick-channel-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px; color: #312e81; }
.quick-close { border: none; background: transparent; color: #64748b; cursor: pointer; font-size: 20px; line-height: 1; }
.quick-channel-footer { display: flex; justify-content: space-between; align-items: center; gap: 12px; }
.btn-outline { padding: 8px 16px; border: 1px solid #e2e8f0; border-radius: 8px; background: #fff; cursor: pointer; font-size: 14px; }
.btn-outline:hover { border-color: #4f46e5; color: #4f46e5; }
.btn-danger { background: #dc2626; color: #fff; border: none; padding: 8px 16px; border-radius: 8px; cursor: pointer; font-size: 14px; }
.btn-danger:hover { background: #b91c1c; }

/* =====================================================
   企业级告警规则弹窗 - 完整重设计
   ===================================================== */

/* 弹窗容器 */
.modal-xl {
  width: 740px;
  max-height: 88vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border-radius: 16px;
  box-shadow: 0 24px 64px rgba(0, 0, 0, 0.18), 0 0 0 1px rgba(0,0,0,0.04);
}

/* ---- Header ---- */
.modal-hd {
  display: flex;
  align-items: center;
  border-bottom: 1px solid #f1f5f9;
  flex-shrink: 0;
  position: relative;
  background: #fff;
  border-radius: 16px 16px 0 0;
}
.modal-hd-bar {
  position: absolute;
  left: 0; top: 0; bottom: 0;
  width: 4px;
  background: linear-gradient(180deg, #4f46e5 0%, #818cf8 100%);
  border-radius: 16px 0 0 0;
}
.modal-hd-inner {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 18px 20px 18px 28px;
  flex: 1;
  min-width: 0;
}
.modal-hd-icon {
  width: 46px;
  height: 46px;
  background: linear-gradient(135deg, #ede9fe 0%, #ddd6fe 100%);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 2px 8px rgba(79,70,229,0.15);
}
.modal-hd-title {
  font-size: 16px;
  font-weight: 700;
  color: #111827;
  margin: 0 0 3px;
  letter-spacing: -0.2px;
}
.modal-hd-sub {
  font-size: 12px;
  color: #94a3b8;
  margin: 0;
}
.modal-close-btn {
  margin-right: 18px;
  width: 32px;
  height: 32px;
  border: none;
  background: #f8fafc;
  border-radius: 8px;
  color: #64748b;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
  flex-shrink: 0;
}
.modal-close-btn:hover { background: #fee2e2; color: #ef4444; }

/* ---- Body ---- */
.modal-bd {
  flex: 1;
  overflow-y: auto;
  padding: 0 24px 4px;
  background: #fff;
}
.modal-bd::-webkit-scrollbar { width: 4px; }
.modal-bd::-webkit-scrollbar-thumb { background: #e2e8f0; border-radius: 2px; }

/* ---- Sections ---- */
.rule-section {
  padding: 20px 0;
  border-bottom: 1px solid #f1f5f9;
}
.rule-section:last-child { border-bottom: none; }
.rule-section-hd {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
}
.sec-num {
  font-size: 10px;
  font-weight: 800;
  color: #4f46e5;
  background: #ede9fe;
  padding: 2px 7px;
  border-radius: 10px;
  letter-spacing: 0.5px;
}
.sec-title {
  font-size: 13px;
  font-weight: 700;
  color: #1e293b;
}

/* ---- Fields ---- */
.field-grid-2 {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px 16px;
}
.field-item { display: flex; flex-direction: column; gap: 5px; }
.field-item.col-2 { grid-column: span 2; }
.field-lbl {
  font-size: 12px;
  font-weight: 600;
  color: #475569;
  display: flex;
  align-items: center;
  gap: 5px;
}
.req-star { color: #ef4444; font-style: normal; font-size: 13px; }
.hint-tag {
  font-size: 11px;
  font-weight: 400;
  color: #94a3b8;
  background: #f1f5f9;
  padding: 1px 6px;
  border-radius: 4px;
}
.field-ctrl {
  width: 100%;
  padding: 9px 12px;
  border: 1.5px solid #e2e8f0;
  border-radius: 8px;
  font-size: 13px;
  outline: none;
  box-sizing: border-box;
  color: #1e293b;
  background: #fafbfd;
  transition: border-color 0.15s, box-shadow 0.15s, background 0.15s;
  font-family: inherit;
}
.field-ctrl:focus {
  border-color: #4f46e5;
  box-shadow: 0 0 0 3px rgba(79,70,229,0.1);
  background: #fff;
}
.field-textarea {
  width: 100%;
  padding: 9px 12px;
  border: 1.5px solid #e2e8f0;
  border-radius: 8px;
  font-size: 13px;
  outline: none;
  box-sizing: border-box;
  color: #1e293b;
  background: #fafbfd;
  resize: vertical;
  font-family: inherit;
  transition: border-color 0.15s, box-shadow 0.15s;
}
.field-textarea:focus {
  border-color: #4f46e5;
  box-shadow: 0 0 0 3px rgba(79,70,229,0.1);
  background: #fff;
}
.field-hint-text { font-size: 11px; color: #94a3b8; margin: 4px 0 0; }

/* 数字输入 */
.num-field-wrap { position: relative; }
.num-field-unit {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 12px;
  color: #94a3b8;
  pointer-events: none;
  font-weight: 500;
}

/* ---- 严重级别卡片 ---- */
.sev-picker {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
}
.sev-card {
  position: relative;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 14px;
  border: 2px solid #e2e8f0;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
  background: #fafbfd;
  overflow: hidden;
  user-select: none;
}
.sev-card:hover { border-color: #a5b4fc; background: #f8f7ff; }
.sev-stripe {
  position: absolute;
  left: 0; top: 0; bottom: 0;
  width: 3px;
}
.sev-critical .sev-stripe { background: #ef4444; }
.sev-warning  .sev-stripe { background: #f59e0b; }
.sev-info     .sev-stripe { background: #3b82f6; }
.sev-content { flex: 1; }
.sev-top { display: flex; align-items: center; gap: 6px; margin-bottom: 3px; }
.sev-dot {
  width: 8px; height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}
.sev-critical .sev-dot { background: #ef4444; box-shadow: 0 0 4px rgba(239,68,68,0.4); }
.sev-warning  .sev-dot { background: #f59e0b; box-shadow: 0 0 4px rgba(245,158,11,0.4); }
.sev-info     .sev-dot { background: #3b82f6; box-shadow: 0 0 4px rgba(59,130,246,0.4); }
.sev-name { font-size: 13px; font-weight: 700; color: #1e293b; }
.sev-sub-label { font-size: 11px; color: #94a3b8; }
.sev-tick {
  width: 20px; height: 20px;
  border-radius: 50%;
  background: #4f46e5;
  color: #fff;
  font-size: 10px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.sev-critical.sev-active { border-color: #fca5a5; background: #fff5f5; box-shadow: 0 2px 8px rgba(239,68,68,0.1); }
.sev-warning.sev-active  { border-color: #fcd34d; background: #fffbf0; box-shadow: 0 2px 8px rgba(245,158,11,0.1); }
.sev-info.sev-active     { border-color: #93c5fd; background: #eff6ff; box-shadow: 0 2px 8px rgba(59,130,246,0.1); }

/* ---- 代码编辑器 ---- */
.code-editor-wrap {
  border: 1px solid #2d2d3f;
  border-radius: 10px;
  overflow: hidden;
  box-shadow: 0 4px 16px rgba(0,0,0,0.08);
}
.code-editor-topbar {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 14px;
  background: #1e1e2e;
}
.code-win-dots { display: flex; gap: 5px; }
.dot { width: 10px; height: 10px; border-radius: 50%; }
.dot-red    { background: #ff5f56; }
.dot-yellow { background: #ffbd2e; }
.dot-green  { background: #27c93f; }
.code-lang-badge {
  font-size: 10px;
  font-weight: 700;
  color: #a78bfa;
  background: rgba(167,139,250,0.15);
  padding: 2px 8px;
  border-radius: 4px;
  letter-spacing: 0.5px;
}
.code-editor-tip {
  flex: 1;
  font-size: 11px;
  color: #4b5563;
  text-align: right;
}
.code-textarea {
  width: 100%;
  padding: 14px 16px;
  border: none;
  font-family: 'Fira Code', 'Cascadia Code', 'Consolas', monospace;
  font-size: 13px;
  color: #e2e8f0;
  background: #252535;
  outline: none;
  resize: vertical;
  min-height: 90px;
  box-sizing: border-box;
  line-height: 1.7;
}
.code-textarea::placeholder { color: #4b5563; }
.code-textarea:focus { background: #1e1e2e; }

/* ---- 通知渠道卡片 ---- */
.ch-picker-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(155px, 1fr));
  gap: 8px;
}
.ch-card {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 9px 11px;
  border: 1.5px solid #e2e8f0;
  border-radius: 8px;
  background: #fafbfd;
  cursor: pointer;
  transition: all 0.15s;
}
.ch-card:hover { border-color: #a5b4fc; background: #f8f7ff; }
.ch-card.ch-selected {
  border-color: #4f46e5;
  background: #f5f3ff;
  box-shadow: 0 0 0 3px rgba(79,70,229,0.08);
}
.ch-icon { font-size: 18px; flex-shrink: 0; }
.ch-meta { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 1px; }
.ch-meta b { font-size: 12px; color: #1e293b; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.ch-meta small { font-size: 11px; color: #94a3b8; }
.ch-check-mark {
  width: 16px; height: 16px;
  border-radius: 50%;
  background: #4f46e5;
  color: #fff;
  font-size: 9px;
  font-weight: 700;
  display: none;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.ch-card.ch-selected .ch-check-mark { display: flex; }
.ch-empty-tip {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 12px 14px;
  border: 1px dashed #cbd5e1;
  border-radius: 8px;
  color: #94a3b8;
  font-size: 12px;
  background: #f8fafc;
}
.ch-quick-btns { display: flex; flex-wrap: wrap; gap: 6px; margin-top: 10px; }
.ch-quick-btn {
  border: 1.5px solid #e2e8f0;
  background: #fff;
  color: #4f46e5;
  padding: 4px 10px;
  font-size: 12px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s;
  font-weight: 500;
}
.ch-quick-btn:hover { border-color: #4f46e5; background: #f5f3ff; }
.ch-quick-muted { color: #64748b; }
.ch-quick-muted:hover { color: #4f46e5; border-color: #4f46e5; }

/* ---- 快速添加渠道面板 ---- */
.qc-panel {
  margin-top: 14px;
  padding: 16px;
  border: 1.5px solid #c7d2fe;
  border-radius: 12px;
  background: linear-gradient(135deg, #f5f3ff 0%, #eff6ff 100%);
}
.qc-panel-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}
.qc-bar-left { display: flex; align-items: center; gap: 8px; }
.qc-tag {
  font-size: 10px;
  font-weight: 700;
  color: #4f46e5;
  background: #ede9fe;
  padding: 2px 7px;
  border-radius: 4px;
  letter-spacing: 0.5px;
}
.qc-title { font-size: 13px; font-weight: 600; color: #1e1e3f; }
.qc-dismiss { border: none; background: transparent; color: #64748b; cursor: pointer; font-size: 20px; line-height: 1; padding: 0; }
.qc-panel-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 12px;
}
.mini-chk-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #4b5563;
  cursor: pointer;
}
.btn-qc-add {
  padding: 7px 18px;
  background: linear-gradient(135deg, #4f46e5, #7c3aed);
  color: #fff;
  border: none;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s;
}
.btn-qc-add:hover { opacity: 0.9; transform: translateY(-1px); }
.btn-qc-add:disabled { opacity: 0.5; cursor: not-allowed; transform: none; }

/* ---- 立即启用开关 ---- */
.enable-row {
  margin-top: 14px;
  padding: 14px 16px;
  background: #f8fafc;
  border-radius: 10px;
  border: 1px solid #f1f5f9;
}
.enable-label {
  display: flex;
  align-items: center;
  gap: 14px;
  cursor: pointer;
}
.toggle-sw { position: relative; width: 40px; height: 22px; flex-shrink: 0; }
.toggle-sw input { opacity: 0; width: 0; height: 0; }
.toggle-track {
  position: absolute;
  cursor: pointer;
  inset: 0;
  background: #cbd5e1;
  border-radius: 22px;
  transition: 0.3s;
}
.toggle-track::before {
  content: '';
  position: absolute;
  width: 18px; height: 18px;
  left: 2px; bottom: 2px;
  background: #fff;
  border-radius: 50%;
  transition: 0.3s;
  box-shadow: 0 1px 3px rgba(0,0,0,0.2);
}
.toggle-sw input:checked + .toggle-track { background: #4f46e5; }
.toggle-sw input:checked + .toggle-track::before { transform: translateX(18px); }
.enable-text { display: flex; flex-direction: column; gap: 2px; }
.enable-title { font-size: 14px; font-weight: 500; color: #1e293b; }
.enable-sub { font-size: 12px; color: #94a3b8; }

/* ---- Footer ---- */
.modal-ft {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 24px;
  border-top: 1px solid #f1f5f9;
  flex-shrink: 0;
  background: #fafbfd;
  border-radius: 0 0 16px 16px;
}
.modal-ft-meta { font-size: 12px; color: #94a3b8; font-family: monospace; }
.modal-ft-btns { display: flex; gap: 10px; }
.btn-ft-cancel {
  padding: 9px 20px;
  border: 1.5px solid #e2e8f0;
  border-radius: 9px;
  background: #fff;
  color: #4b5563;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}
.btn-ft-cancel:hover { border-color: #4f46e5; color: #4f46e5; }
.btn-ft-save {
  display: flex;
  align-items: center;
  gap: 7px;
  padding: 9px 22px;
  border: none;
  border-radius: 9px;
  background: linear-gradient(135deg, #4f46e5 0%, #7c3aed 100%);
  color: #fff;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  box-shadow: 0 2px 10px rgba(79,70,229,0.35);
  transition: all 0.15s;
  letter-spacing: 0.2px;
}
.btn-ft-save:hover {
  background: linear-gradient(135deg, #4338ca 0%, #6d28d9 100%);
  box-shadow: 0 4px 14px rgba(79,70,229,0.45);
  transform: translateY(-1px);
}
.btn-ft-save:disabled { opacity: 0.6; cursor: not-allowed; transform: none; }

/* ===== Header Actions ===== */
.header-actions { display: flex; align-items: center; gap: 10px; }
.btn-action {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 8px 14px; border: 1.5px solid #e2e8f0;
  border-radius: 8px; background: #fff; color: #4b5563;
  font-size: 13px; font-weight: 500; cursor: pointer; transition: all 0.15s;
}
.btn-action:hover { border-color: #4f46e5; color: #4f46e5; background: #f5f3ff; }
.btn-yaml-import:hover { border-color: #059669; color: #059669; background: #ecfdf5; }
.btn-yaml-export:hover { border-color: #d97706; color: #d97706; background: #fffbeb; }

/* ===== YAML Modal ===== */
.yaml-modal { max-width: 780px; }
.yaml-modal-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 20px 24px; border-bottom: 1px solid #f1f5f9;
}
.yaml-header-left { display: flex; align-items: center; gap: 14px; }
.yaml-header-left h3 { margin: 0; font-size: 16px; font-weight: 600; color: #1e293b; }
.yaml-header-desc { margin: 2px 0 0; font-size: 12px; color: #94a3b8; }
.yaml-icon-wrap {
  width: 42px; height: 42px; border-radius: 12px;
  display: flex; align-items: center; justify-content: center;
}
.yaml-icon-wrap.import { background: linear-gradient(135deg, #ecfdf5, #d1fae5); color: #059669; }
.yaml-icon-wrap.export { background: linear-gradient(135deg, #fffbeb, #fef3c7); color: #d97706; }
.yaml-modal-body { padding: 20px 24px; overflow-y: auto; max-height: 65vh; }
.yaml-modal-footer {
  display: flex; align-items: center; justify-content: flex-end; gap: 10px;
  padding: 14px 24px; border-top: 1px solid #f1f5f9; background: #fafbfd;
  border-radius: 0 0 16px 16px;
}
.yaml-toolbar {
  display: flex; align-items: center; justify-content: space-between;
  margin-bottom: 12px;
}
.yaml-overwrite-chk {
  display: flex; align-items: center; gap: 8px;
  font-size: 13px; color: #4b5563; cursor: pointer;
}
.yaml-overwrite-chk input { accent-color: #4f46e5; }
.yaml-tpl-btn {
  border: 1.5px solid #e2e8f0; background: #fff;
  color: #4f46e5; padding: 5px 12px; font-size: 12px;
  border-radius: 6px; cursor: pointer; font-weight: 500; transition: all 0.15s;
}
.yaml-tpl-btn:hover { background: #f5f3ff; border-color: #4f46e5; }

.yaml-editor-wrap {
  border: 1.5px solid #e2e8f0; border-radius: 12px;
  overflow: hidden; background: #1e1e2e;
}
.yaml-editor-wrap.export-view { background: #0f172a; }
.yaml-editor-topbar {
  display: flex; align-items: center; gap: 10px;
  padding: 8px 14px; background: #2d2d3f; border-bottom: 1px solid #3d3d5c;
}
.yaml-editor-hint { font-size: 11px; color: #a5b4fc; margin-left: auto; }
.yaml-textarea {
  width: 100%; border: none; outline: none; resize: vertical;
  padding: 16px; font-family: 'JetBrains Mono', 'Fira Code', 'Cascadia Code', monospace;
  font-size: 13px; line-height: 1.6; color: #e2e8f0;
  background: transparent; min-height: 260px;
}
.yaml-textarea::placeholder { color: #64748b; }
.yaml-textarea:focus { box-shadow: inset 0 0 0 1px #4f46e5; }

/* YAML 导入结果 */
.yaml-import-result {
  margin-top: 16px; padding: 16px;
  border: 1.5px solid #e2e8f0; border-radius: 10px; background: #f8fafc;
}
.import-result-header { display: flex; align-items: center; gap: 8px; margin-bottom: 10px; }
.import-result-icon { font-size: 16px; }
.import-result-title { font-size: 14px; font-weight: 600; color: #1e293b; }
.import-result-stats { display: flex; flex-wrap: wrap; gap: 10px; }
.stat-item {
  padding: 4px 12px; border-radius: 6px; font-size: 12px; font-weight: 600;
}
.stat-item.success { background: #d1fae5; color: #065f46; }
.stat-item.update { background: #dbeafe; color: #1e40af; }
.stat-item.skip { background: #f1f5f9; color: #64748b; }
.stat-item.fail { background: #fee2e2; color: #991b1b; }
.import-result-errors { margin-top: 10px; padding: 10px; background: #fef2f2; border-radius: 6px; }
.error-line { font-size: 12px; color: #991b1b; padding: 2px 0; font-family: monospace; }

/* Copy button */
.btn-yaml-copy { border-color: #4f46e5; color: #4f46e5; }
.btn-yaml-copy:hover { background: #ede9fe; }

</style>
