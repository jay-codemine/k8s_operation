<template>
  <div class="notfound">
    <h2>404 - 路由不存在</h2>
    <p>当前访问路径：</p>
    <pre>{{ $route.fullPath }}</pre>

    <div class="actions">
      <button @click="$router.back()">返回上一页</button>
      <button @click="$router.push('/clusters')">去集群列表</button>
      <button @click="$router.push('/dashboard')">去 Dashboard</button>
    </div>
  </div>
</template>

<script setup>
</script>

<style scoped>
.notfound {
  padding: 24px;
}

pre {
  background: #0b1220;
  color: #e2e8f0;
  padding: 12px;
  border-radius: 8px;
  overflow: auto;
}

.actions {
  margin-top: 12px;
  display: flex;
  gap: 10px;
}

button {
  padding: 8px 12px;
  border: 0;
  border-radius: 8px;
  cursor: pointer;
}
</style>
