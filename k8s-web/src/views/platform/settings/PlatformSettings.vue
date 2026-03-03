<template>
  <div>
    <h2 style="margin:0 0 12px;">系统设置</h2>
    <div class="card">
      <div class="row">
        <div class="label">默认进入页</div>
        <div class="value">/clusters</div>
      </div>
      <div class="row">
        <div class="label">默认集群</div>
        <div class="value">activeClusterId（localStorage）</div>
      </div>
      <div class="row">
        <div class="label">告警阈值</div>
        <div class="value">CPU &gt; 80% / Mem &gt; 80%</div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.card{background:#fff;border:1px solid #e6eaf2;border-radius:14px;padding:12px;}
.row{display:flex;justify-content:space-between;padding:10px 0;border-bottom:1px solid #eef2f7;}
.row:last-child{border-bottom:0;}
.label{color:#64748b;}
.value{font-weight:700;color:#0f172a;}
</style>
