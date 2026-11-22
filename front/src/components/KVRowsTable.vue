<template>
  <div class="form-section">
    <div class="section-header-row">
      <div class="section-title">{{ title }}</div>
      <el-button size="small" @click="add">{{ addText }}</el-button>
    </div>
    <el-table :data="modelValue" style="width: 100%">
      <el-table-column label="键" width="320">
        <template #default="{ row }">
          <el-input v-model="row.key" :placeholder="keyPlaceholder" />
        </template>
      </el-table-column>
      <el-table-column label="值">
        <template #default="{ row }">
          <el-input v-model="row.value" :placeholder="valuePlaceholder" />
        </template>
      </el-table-column>
      <el-table-column label="操作" width="90" align="center">
        <template #default="{ $index }">
          <el-button link type="danger" @click="remove($index)">删除</el-button>
        </template>
      </el-table-column>
      <template #empty>
        <span>{{ emptyText }}</span>
      </template>
    </el-table>
  </div>
</template>

<script setup lang="ts">
// 表单弹窗里的键值编辑表格(标签、注解、参数等)
export interface KVRow {
  key: string
  value: string
}

const props = withDefaults(
  defineProps<{
    modelValue: KVRow[]
    title: string
    addText?: string
    emptyText?: string
    keyPlaceholder?: string
    valuePlaceholder?: string
  }>(),
  {
    addText: '添加',
    emptyText: '暂无数据',
    keyPlaceholder: 'key',
    valuePlaceholder: 'value'
  }
)

const emit = defineEmits<{ 'update:modelValue': [rows: KVRow[]] }>()

const add = () => emit('update:modelValue', [...props.modelValue, { key: '', value: '' }])
const remove = (index: number) => emit('update:modelValue', props.modelValue.filter((_, i) => i !== index))
</script>
