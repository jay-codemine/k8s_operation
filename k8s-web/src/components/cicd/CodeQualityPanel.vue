<template>
  <div class="code-quality-panel">
    <!-- 顶部状态栏 -->
    <div class="quality-header">
      <div class="header-left">
        <div :class="['gate-badge', `gate-${gateStatus}`]">
          <svg v-if="gateStatus === 'OK'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
            <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
            <polyline points="22 4 12 14.01 9 11.01"/>
          </svg>
          <svg v-else-if="gateStatus === 'ERROR'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
            <circle cx="12" cy="12" r="10"/>
            <line x1="15" y1="9" x2="9" y2="15"/>
            <line x1="9" y1="9" x2="15" y2="15"/>
          </svg>
          <svg v-else-if="gateStatus === 'WARN'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
            <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>
            <line x1="12" y1="9" x2="12" y2="13"/>
            <line x1="12" y1="17" x2="12.01" y2="17"/>
          </svg>
          <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <line x1="12" y1="8" x2="12" y2="12"/>
            <line x1="12" y1="16" x2="12.01" y2="16"/>
          </svg>
        </div>
        <div class="gate-info">
          <h3 class="gate-title">Quality Gate</h3>
          <span :class="['gate-status-text', `gate-${gateStatus}`]">{{ gateStatusText }}</span>
        </div>
      </div>
      <div class="header-right">
        <a v-if="report.dashboard_url" :href="report.dashboard_url" target="_blank" class="sonar-link">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/>
            <polyline points="15 3 21 3 21 9"/>
            <line x1="10" y1="14" x2="21" y2="3"/>
          </svg>
          SonarQube Dashboard
        </a>
        <button class="refresh-btn" @click="$emit('refresh')" :disabled="loading">
          <svg :class="{ spinning: loading }" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="23 4 23 10 17 10"/>
            <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
          </svg>
        </button>
      </div>
    </div>

    <!-- 三大评级卡片 -->
    <div class="rating-cards">
      <div class="rating-card">
        <div class="rating-header">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
          </svg>
          <span>可靠性</span>
        </div>
        <div :class="['rating-grade', `grade-${report.reliability_rating || 'A'}`]">
          {{ report.reliability_rating || 'A' }}
        </div>
        <div class="rating-metric">
          <span class="metric-value bug-color">{{ report.bugs || 0 }}</span>
          <span class="metric-label">Bugs</span>
        </div>
      </div>

      <div class="rating-card">
        <div class="rating-header">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/>
            <path d="M7 11V7a5 5 0 0 1 10 0v4"/>
          </svg>
          <span>安全性</span>
        </div>
        <div :class="['rating-grade', `grade-${report.security_rating || 'A'}`]">
          {{ report.security_rating || 'A' }}
        </div>
        <div class="rating-metric">
          <span class="metric-value vuln-color">{{ report.vulnerabilities || 0 }}</span>
          <span class="metric-label">漏洞</span>
        </div>
      </div>

      <div class="rating-card">
        <div class="rating-header">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="16 18 22 12 16 6"/>
            <polyline points="8 6 2 12 8 18"/>
          </svg>
          <span>可维护性</span>
        </div>
        <div :class="['rating-grade', `grade-${report.maintainability_rating || 'A'}`]">
          {{ report.maintainability_rating || 'A' }}
        </div>
        <div class="rating-metric">
          <span class="metric-value smell-color">{{ report.code_smells || 0 }}</span>
          <span class="metric-label">异味</span>
        </div>
      </div>
    </div>

    <!-- 关键指标仪表盘 -->
    <div class="metrics-dashboard">
      <div class="metric-card coverage-card">
        <div class="metric-ring-container">
          <svg viewBox="0 0 120 120" class="metric-ring">
            <circle cx="60" cy="60" r="52" fill="none" stroke="rgba(255,255,255,0.08)" stroke-width="10"/>
            <circle cx="60" cy="60" r="52" fill="none" :stroke="coverageColor"
              stroke-width="10" stroke-linecap="round"
              :stroke-dasharray="`${(report.coverage || 0) * 3.267} 326.7`"
              transform="rotate(-90 60 60)"
              class="progress-ring"
            />
          </svg>
          <div class="ring-value">
            <span class="value-number">{{ (report.coverage || 0).toFixed(1) }}</span>
            <span class="value-unit">%</span>
          </div>
        </div>
        <div class="metric-detail">
          <span class="metric-title">代码覆盖率</span>
          <span class="metric-sub" v-if="report.new_coverage">
            新代码: {{ report.new_coverage.toFixed(1) }}%
          </span>
        </div>
      </div>

      <div class="metric-card duplications-card">
        <div class="metric-ring-container">
          <svg viewBox="0 0 120 120" class="metric-ring">
            <circle cx="60" cy="60" r="52" fill="none" stroke="rgba(255,255,255,0.08)" stroke-width="10"/>
            <circle cx="60" cy="60" r="52" fill="none" :stroke="duplicationsColor"
              stroke-width="10" stroke-linecap="round"
              :stroke-dasharray="`${(report.duplications || 0) * 3.267} 326.7`"
              transform="rotate(-90 60 60)"
              class="progress-ring"
            />
          </svg>
          <div class="ring-value">
            <span class="value-number">{{ (report.duplications || 0).toFixed(1) }}</span>
            <span class="value-unit">%</span>
          </div>
        </div>
        <div class="metric-detail">
          <span class="metric-title">重复代码率</span>
          <span class="metric-sub">
            {{ formatNumber(report.lines_of_code || 0) }} 行代码
          </span>
        </div>
      </div>

      <div class="metric-card hotspot-card">
        <div class="hotspot-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12 2L2 7l10 5 10-5-10-5z"/>
            <path d="M2 17l10 5 10-5"/>
            <path d="M2 12l10 5 10-5"/>
          </svg>
        </div>
        <div class="hotspot-value">{{ report.security_hotspots || 0 }}</div>
        <div class="hotspot-label">安全热点</div>
      </div>

      <div class="metric-card loc-card">
        <div class="loc-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
            <polyline points="14 2 14 8 20 8"/>
            <line x1="16" y1="13" x2="8" y2="13"/>
            <line x1="16" y1="17" x2="8" y2="17"/>
          </svg>
        </div>
        <div class="loc-value">{{ formatNumber(report.lines_of_code || 0) }}</div>
        <div class="loc-label">代码行数</div>
      </div>
    </div>

    <!-- 新代码变更（增量分析） -->
    <div v-if="hasNewCodeMetrics" class="new-code-section">
      <h4 class="section-label">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="23 6 13.5 15.5 8.5 10.5 1 18"/>
          <polyline points="17 6 23 6 23 12"/>
        </svg>
        新代码变更分析
      </h4>
      <div class="new-code-metrics">
        <div class="new-metric" v-if="report.new_bugs > 0">
          <span class="new-metric-value bug-color">+{{ report.new_bugs }}</span>
          <span class="new-metric-label">新增 Bug</span>
        </div>
        <div class="new-metric" v-if="report.new_vulnerabilities > 0">
          <span class="new-metric-value vuln-color">+{{ report.new_vulnerabilities }}</span>
          <span class="new-metric-label">新增漏洞</span>
        </div>
        <div class="new-metric" v-if="report.new_code_smells > 0">
          <span class="new-metric-value smell-color">+{{ report.new_code_smells }}</span>
          <span class="new-metric-label">新增异味</span>
        </div>
        <div class="new-metric" v-if="!report.new_bugs && !report.new_vulnerabilities && !report.new_code_smells">
          <span class="new-metric-value clean-color">0</span>
          <span class="new-metric-label">新增问题</span>
        </div>
      </div>
    </div>

    <!-- 无数据提示 -->
    <div v-if="report.message && gateStatus === 'NONE'" class="no-data-hint">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
        <circle cx="12" cy="12" r="10"/>
        <line x1="12" y1="8" x2="12" y2="12"/>
        <line x1="12" y1="16" x2="12.01" y2="16"/>
      </svg>
      <span>{{ report.message }}</span>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  report: { type: Object, default: () => ({}) },
  loading: { type: Boolean, default: false }
})

defineEmits(['refresh'])

const gateStatus = computed(() => props.report.quality_gate || 'NONE')

const gateStatusText = computed(() => {
  const map = { OK: 'Passed', WARN: 'Warning', ERROR: 'Failed', NONE: '未扫描' }
  return map[gateStatus.value] || '未知'
})

const coverageColor = computed(() => {
  const v = props.report.coverage || 0
  if (v >= 80) return '#10b981'
  if (v >= 50) return '#f59e0b'
  return '#ef4444'
})

const duplicationsColor = computed(() => {
  const v = props.report.duplications || 0
  if (v <= 3) return '#10b981'
  if (v <= 10) return '#f59e0b'
  return '#ef4444'
})

const hasNewCodeMetrics = computed(() => {
  return props.report.new_bugs > 0 || props.report.new_vulnerabilities > 0 || props.report.new_code_smells > 0
})

const formatNumber = (n) => {
  if (n >= 1000000) return (n / 1000000).toFixed(1) + 'M'
  if (n >= 1000) return (n / 1000).toFixed(1) + 'K'
  return n.toString()
}
</script>

<style scoped>
.code-quality-panel {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* ============ 顶部状态栏 ============ */
.quality-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  background: linear-gradient(135deg, rgba(30, 41, 59, 0.95), rgba(15, 23, 42, 0.98));
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 12px;
  backdrop-filter: blur(12px);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.gate-badge {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s;
}

.gate-badge svg {
  width: 24px;
  height: 24px;
}

.gate-badge.gate-OK {
  background: rgba(16, 185, 129, 0.15);
  border: 1px solid rgba(16, 185, 129, 0.3);
  color: #10b981;
}

.gate-badge.gate-ERROR {
  background: rgba(239, 68, 68, 0.15);
  border: 1px solid rgba(239, 68, 68, 0.3);
  color: #ef4444;
}

.gate-badge.gate-WARN {
  background: rgba(245, 158, 11, 0.15);
  border: 1px solid rgba(245, 158, 11, 0.3);
  color: #f59e0b;
}

.gate-badge.gate-NONE {
  background: rgba(148, 163, 184, 0.1);
  border: 1px solid rgba(148, 163, 184, 0.15);
  color: #94a3b8;
}

.gate-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.gate-title {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 1.2px;
  color: #94a3b8;
  font-weight: 600;
  margin: 0;
}

.gate-status-text {
  font-size: 20px;
  font-weight: 700;
  letter-spacing: -0.3px;
}

.gate-status-text.gate-OK { color: #10b981; }
.gate-status-text.gate-ERROR { color: #ef4444; }
.gate-status-text.gate-WARN { color: #f59e0b; }
.gate-status-text.gate-NONE { color: #64748b; }

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.sonar-link {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border-radius: 8px;
  background: rgba(99, 102, 241, 0.1);
  border: 1px solid rgba(99, 102, 241, 0.2);
  color: #818cf8;
  text-decoration: none;
  font-size: 13px;
  font-weight: 500;
  transition: all 0.2s;
}

.sonar-link:hover {
  background: rgba(99, 102, 241, 0.2);
  border-color: rgba(99, 102, 241, 0.4);
}

.sonar-link svg {
  width: 14px;
  height: 14px;
}

.refresh-btn {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  border: 1px solid rgba(148, 163, 184, 0.15);
  background: rgba(255, 255, 255, 0.04);
  color: #94a3b8;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.refresh-btn:hover {
  background: rgba(255, 255, 255, 0.08);
  color: #e2e8f0;
}

.refresh-btn svg {
  width: 16px;
  height: 16px;
}

.refresh-btn svg.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* ============ 三大评级卡片 ============ */
.rating-cards {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}

.rating-card {
  background: linear-gradient(135deg, rgba(30, 41, 59, 0.9), rgba(15, 23, 42, 0.95));
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 12px;
  padding: 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  transition: all 0.3s;
}

.rating-card:hover {
  border-color: rgba(148, 163, 184, 0.2);
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.2);
}

.rating-header {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #94a3b8;
  font-size: 13px;
  font-weight: 500;
}

.rating-header svg {
  width: 16px;
  height: 16px;
}

.rating-grade {
  width: 56px;
  height: 56px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  font-weight: 800;
  letter-spacing: -1px;
}

.grade-A { background: rgba(16, 185, 129, 0.15); color: #10b981; border: 2px solid rgba(16, 185, 129, 0.3); }
.grade-B { background: rgba(132, 204, 22, 0.15); color: #84cc16; border: 2px solid rgba(132, 204, 22, 0.3); }
.grade-C { background: rgba(245, 158, 11, 0.15); color: #f59e0b; border: 2px solid rgba(245, 158, 11, 0.3); }
.grade-D { background: rgba(249, 115, 22, 0.15); color: #f97316; border: 2px solid rgba(249, 115, 22, 0.3); }
.grade-E { background: rgba(239, 68, 68, 0.15); color: #ef4444; border: 2px solid rgba(239, 68, 68, 0.3); }

.rating-metric {
  display: flex;
  align-items: baseline;
  gap: 6px;
}

.metric-value {
  font-size: 22px;
  font-weight: 700;
}

.metric-label {
  font-size: 12px;
  color: #64748b;
}

.bug-color { color: #ef4444; }
.vuln-color { color: #f97316; }
.smell-color { color: #eab308; }
.clean-color { color: #10b981; }

/* ============ 关键指标仪表盘 ============ */
.metrics-dashboard {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.metric-card {
  background: linear-gradient(135deg, rgba(30, 41, 59, 0.9), rgba(15, 23, 42, 0.95));
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 12px;
  padding: 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  transition: all 0.3s;
}

.metric-card:hover {
  border-color: rgba(148, 163, 184, 0.2);
}

.metric-ring-container {
  position: relative;
  width: 100px;
  height: 100px;
}

.metric-ring {
  width: 100%;
  height: 100%;
}

.progress-ring {
  transition: stroke-dasharray 1s ease;
}

.ring-value {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  text-align: center;
}

.value-number {
  font-size: 22px;
  font-weight: 700;
  color: #e2e8f0;
}

.value-unit {
  font-size: 12px;
  color: #94a3b8;
  margin-left: 1px;
}

.metric-detail {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}

.metric-title {
  font-size: 13px;
  color: #94a3b8;
  font-weight: 500;
}

.metric-sub {
  font-size: 11px;
  color: #64748b;
}

.hotspot-card, .loc-card {
  justify-content: center;
}

.hotspot-icon svg, .loc-icon svg {
  width: 32px;
  height: 32px;
  color: #f59e0b;
}

.loc-icon svg {
  color: #6366f1;
}

.hotspot-value, .loc-value {
  font-size: 28px;
  font-weight: 700;
  color: #e2e8f0;
}

.hotspot-label, .loc-label {
  font-size: 12px;
  color: #64748b;
  font-weight: 500;
}

/* ============ 新代码变更 ============ */
.new-code-section {
  background: linear-gradient(135deg, rgba(30, 41, 59, 0.9), rgba(15, 23, 42, 0.95));
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 12px;
  padding: 16px 20px;
}

.section-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: #94a3b8;
  font-weight: 600;
  margin: 0 0 12px 0;
}

.section-label svg {
  width: 16px;
  height: 16px;
  color: #6366f1;
}

.new-code-metrics {
  display: flex;
  gap: 24px;
}

.new-metric {
  display: flex;
  align-items: baseline;
  gap: 6px;
}

.new-metric-value {
  font-size: 18px;
  font-weight: 700;
}

.new-metric-label {
  font-size: 12px;
  color: #64748b;
}

/* ============ 无数据提示 ============ */
.no-data-hint {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 24px;
  background: rgba(30, 41, 59, 0.5);
  border: 1px dashed rgba(148, 163, 184, 0.15);
  border-radius: 12px;
  color: #64748b;
  font-size: 13px;
}

.no-data-hint svg {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
}

/* ============ 响应式 ============ */
@media (max-width: 1200px) {
  .metrics-dashboard {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .rating-cards {
    grid-template-columns: 1fr;
  }
  .metrics-dashboard {
    grid-template-columns: 1fr;
  }
}
</style>
