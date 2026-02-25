<template>
  <div>
    <h2 style="margin:0 0 12px;">平台健康</h2>

    <div class="grid">
      <div class="card">
        <div class="label">平台状态</div>
        <div class="value ok">正常</div>
        <div class="desc">最近探测：{{ health.lastCheck }}</div>
      </div>

      <div class="card">
        <div class="label">集群在线</div>
        <div class="value">{{ health.clustersOnline }}/{{ health.clustersTotal }}</div>
        <div class="desc">异常：{{ health.clustersAbnormal }}</div>
      </div>

      <div class="card">
        <div class="label">告警（24h）</div>
        <div class="value">{{ health.alerts24h }}</div>
        <div class="desc">严重：{{ health.criticalAlerts }}</div>
      </div>

      <div class="card">
        <div class="label">任务队列</div>
        <div class="value">{{ health.queue }}</div>
        <div class="desc">平均延迟：{{ health.queueLag }}</div>
      </div>
    </div>

    <h3 style="margin:16px 0 8px;">组件状态（静态）</h3>
    <table class="table">
      <thead>
      <tr>
        <th>组件</th>
        <th>状态</th>
        <th>说明</th>
      </tr>
      </thead>
      <tbody>
      <tr v-for="x in components" :key="x.name">
        <td>{{ x.name }}</td>
        <td><span :class="['pill', x.ok ? 'pill-ok' : 'pill-bad']">{{ x.ok ? 'OK' : 'DOWN' }}</span>
        </td>
        <td style="color:#64748b;">{{ x.note }}</td>
      </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
const health = {
  lastCheck: '2026-01-20 21:10',
  clustersOnline: 2,
  clustersTotal: 2,
  clustersAbnormal: 0,
  alerts24h: 3,
  criticalAlerts: 0,
  queue: 12,
  queueLag: '120ms',
}

const components = [
  {name: 'API Server', ok: true, note: '请求延迟 28ms'},
  {name: 'Auth Service', ok: true, note: 'Token 校验正常'},
  {name: 'Cluster Agent', ok: true, note: '2/2 在线'},
  {name: 'Event Collector', ok: true, note: '事件接收正常'},
]
</script>

<style scoped>
.grid {
  display: grid;
  grid-template-columns:repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.card {
  background: #fff;
  border: 1px solid #e6eaf2;
  border-radius: 14px;
  padding: 12px;
}

.label {
  color: #64748b;
  font-size: 12px;
}

.value {
  font-size: 22px;
  font-weight: 800;
  margin-top: 6px;
}

.ok {
  color: #16a34a;
}

.desc {
  margin-top: 6px;
  color: #64748b;
  font-size: 12px;
}

.table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  margin-top: 8px;
  background: #fff;
  border: 1px solid #e6eaf2;
  border-radius: 14px;
  overflow: hidden;
}

.table th, .table td {
  padding: 10px 12px;
  border-bottom: 1px solid #eef2f7;
  text-align: left;
}

.table tr:last-child td {
  border-bottom: 0;
}

.pill {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 12px;
}

.pill-ok {
  background: #dcfce7;
  color: #16a34a;
}

.pill-bad {
  background: #fee2e2;
  color: #ef4444;
}
</style>
