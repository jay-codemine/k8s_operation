<template>
  <div>
    <h2 style="margin:0 0 12px;">应用商城</h2>
    <p style="color:#64748b;margin:0 0 12px;">先用静态数据占位：后续可接 Helm Chart / 插件市场 /
      模版库</p>

    <div class="grid">
      <div class="app" v-for="a in apps" :key="a.name">
        <div class="title">{{ a.name }}</div>
        <div class="meta">{{ a.category }} · v{{ a.version }}</div>
        <div class="desc">{{ a.desc }}</div>
        <div class="actions">
          <button class="btn">查看详情</button>
          <button class="btn primary">安装</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
const apps = [
  {
    name: 'Ingress-NGINX',
    category: '网络',
    version: '1.10.1',
    desc: 'Kubernetes Ingress Controller（示例）'
  },
  {
    name: 'Prometheus Stack',
    category: '监控',
    version: '56.6.2',
    desc: 'Prometheus + Grafana + Alertmanager（示例）'
  },
  {name: 'Argo CD', category: 'GitOps', version: '2.11.3', desc: '持续交付与应用同步（示例）'},
  {
    name: 'EFK',
    category: '日志',
    version: '0.1.0',
    desc: 'Elasticsearch + Fluentbit + Kibana（示例）'
  },
]
</script>

<style scoped>
.grid {
  display: grid;
  grid-template-columns:repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.app {
  background: #fff;
  border: 1px solid #e6eaf2;
  border-radius: 14px;
  padding: 12px;
}

.title {
  font-weight: 900;
  font-size: 16px;
}

.meta {
  color: #64748b;
  font-size: 12px;
  margin-top: 4px;
}

.desc {
  color: #334155;
  margin-top: 8px;
  min-height: 40px;
}

.actions {
  display: flex;
  gap: 10px;
  margin-top: 10px;
}

.btn {
  border: 1px solid #e6eaf2;
  background: #fff;
  border-radius: 10px;
  padding: 8px 12px;
  cursor: pointer;
}

.primary {
  background: #326ce5;
  color: #fff;
  border-color: #326ce5;
}
</style>
