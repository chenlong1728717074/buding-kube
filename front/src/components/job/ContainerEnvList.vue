<template>
  <el-empty v-if="!containers.length" description="没有容器" :image-size="72" />
  <div v-else>
    <div v-for="c in containers" :key="(c.init ? 'init-' : '') + c.name" class="jd-section">
      <h4 class="jd-section__title">
        {{ c.name }}
        <el-tag v-if="c.init" size="small" type="warning">初始化容器</el-tag>
      </h4>
      <el-table :data="c.envs" style="width: 100%" :header-cell-class-name="tableHeaderClass">
        <el-table-column label="变量名" min-width="220">
          <template #default="{ row }"><span class="mono">{{ row.name }}</span></template>
        </el-table-column>
        <el-table-column label="值" min-width="320">
          <template #default="{ row }">
            <span v-if="row.source" class="jd-muted">{{ row.source }}</span>
            <span v-else class="mono" style="white-space: pre-wrap; word-break: break-all">{{ row.value === '' ? '""' : row.value }}</span>
          </template>
        </el-table-column>
        <template #empty>
          <span class="jd-muted">该容器没有设置环境变量</span>
        </template>
      </el-table>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ContainerTemplateVO } from '@/api/job'
import { tableHeaderClass } from '@/utils/table'

defineProps<{ containers: ContainerTemplateVO[] }>()
</script>
