<template>
  <div class="resource-page">
    <div class="page-header">
      <h1>CRD</h1>
    </div>

    <el-card class="info-card">
      <el-empty description="功能开发中" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
</script>

<style scoped>
.resource-page {
  padding: 20px;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: var(--gap-4);
  background: #ffffff;
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-card);
}
.page-header h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: #2c3e50;
}
.info-card {
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-card);
}
</style>
