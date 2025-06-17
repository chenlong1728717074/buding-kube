<template>
  <div class="coming-soon">
    <el-result
      icon="info"
      title="功能开发中"
      sub-title="Deployment 管理功能正在开发中，敬请期待！"
    >
      <template #extra>
        <el-button type="primary" @click="$router.go(-1)">返回</el-button>
      </template>
    </el-result>
  </div>
</template>

<script setup lang="ts">
// Deployment 管理页面 - 暂未实现
</script>

<style scoped>
.coming-soon {
  padding: 50px;
  text-align: center;
}
</style>