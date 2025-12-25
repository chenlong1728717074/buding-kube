<template>
  <div class="resource-page">
    <div class="page-header">
      <h1>CRD</h1>
    </div>

    <el-card class="info-card">
      <div class="coming-soon">
        <el-result
          icon="info"
          title="功能开发中"
          sub-title="CRD 管理功能正在开发中，敬请期待！"
        >
          <template #extra>
            <el-button type="primary" @click="$router.go(-1)">返回</el-button>
          </template>
        </el-result>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
</script>

<style scoped>
.resource-page {
  padding: 20px;
}
.info-card {
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-card);
}

.coming-soon {
  padding: 50px;
  text-align: center;
}
</style>
