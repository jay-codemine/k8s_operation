<template>
  <div>
    <h2 style="margin:0 0 12px;">审计日志</h2>
    <table class="table">
      <thead>
      <tr>
        <th>时间</th>
        <th>用户</th>
        <th>动作</th>
        <th>对象</th>
        <th>结果</th>
      </tr>
      </thead>
      <tbody>
      <tr v-for="x in rows" :key="x.time + x.action">
        <td>{{ x.time }}</td>
        <td>{{ x.user }}</td>
        <td>{{ x.action }}</td>
        <td>{{ x.object }}</td>
        <td><span :class="['pill', x.ok ? 'ok' : 'bad']">{{ x.ok ? 'SUCCESS' : 'FAILED' }}</span>
        </td>
      </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
const rows = [
  {time: '2026-01-20 20:58', user: 'admin', action: 'LOGIN', object: '-', ok: true},
  {time: '2026-01-20 21:01', user: 'admin', action: 'CREATE_CLUSTER', object: 'prod-k8s', ok: true},
  {
    time: '2026-01-20 21:05',
    user: 'admin',
    action: 'DELETE_POD',
    object: 'c/1 default/nginx',
    ok: false
  },
]
</script>

<style scoped>
.table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
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

.ok {
  background: #dcfce7;
  color: #16a34a;
}

.bad {
  background: #fee2e2;
  color: #ef4444;
}
</style>
