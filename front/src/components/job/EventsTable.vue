<template>
  <div>
    <el-table :data="rows" style="width: 100%" :header-cell-class-name="tableHeaderClass" :cell-class-name="tableCellClass">
      <el-table-column label="类型" width="100">
        <template #default="{ row }">
          <el-tag :type="row.type === 'Warning' ? 'warning' : 'success'" size="small">{{ row.type === 'Warning' ? '警告' : '正常' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="原因" min-width="150" prop="reason" show-overflow-tooltip />
      <el-table-column label="发生时间" width="200">
        <template #default="{ row }">
          {{ formatDateTime(row.lastTime || row.firstTime) }}
          <span v-if="row.count > 1" class="jd-muted">（{{ row.count }} 次）</span>
        </template>
      </el-table-column>
      <el-table-column label="来源" min-width="150" show-overflow-tooltip>
        <template #default="{ row }">{{ row.source || '-' }}</template>
      </el-table-column>
      <el-table-column label="消息" min-width="360">
        <template #default="{ row }"><span style="word-break: break-all">{{ row.message }}</span></template>
      </el-table-column>
      <template #empty>
        <el-empty description="暂无事件（Kubernetes 默认只保留 1 小时内的事件）" :image-size="72" />
      </template>
    </el-table>
    <ListPagination
      v-model:current-page="page"
      v-model:page-size="pageSize"
      class="jd-pager"
      :total="total"
      :page-sizes="[10, 20, 50]"
    />
  </div>
</template>

<script setup lang="ts">
import { toRef } from 'vue'
import ListPagination from '@/components/ListPagination.vue'
import type { EventVO } from '@/api/job'
import { useClientPage } from '@/composables/useClientPage'
import { tableHeaderClass, tableCellClass } from '@/utils/table'
import { formatDateTime } from '@/utils/job'

const props = defineProps<{ events: EventVO[] }>()
const { page, pageSize, total, rows } = useClientPage(toRef(props, 'events'))
</script>
