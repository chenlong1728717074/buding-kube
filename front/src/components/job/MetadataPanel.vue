<template>
  <el-row :gutter="18">
    <el-col :span="12">
      <h4 class="jd-section__title">标签 <el-tag size="small" type="info">{{ labelRows.length }}</el-tag></h4>
      <div class="labels-content">
        <el-empty v-if="!labelRows.length" description="没有标签" :image-size="60" />
        <div v-for="row in labelRows" v-else :key="row.key" class="label-item">
          <div class="label-key">{{ row.key }}</div>
          <div class="label-value">{{ row.value || '""' }}</div>
        </div>
      </div>
    </el-col>
    <el-col :span="12">
      <h4 class="jd-section__title">注解 <el-tag size="small" type="info">{{ annotationRows.length }}</el-tag></h4>
      <div class="labels-content">
        <el-empty v-if="!annotationRows.length" description="没有注解" :image-size="60" />
        <div v-for="row in annotationRows" v-else :key="row.key" class="label-item">
          <div class="label-key">{{ row.key }}</div>
          <div class="label-value">{{ row.value || '""' }}</div>
        </div>
      </div>
    </el-col>
  </el-row>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  labels?: Record<string, string> | null
  annotations?: Record<string, string> | null
}>()

// kubectl 的 last-applied 注解是整份 JSON, 太长且和 YAML 重复, 不展示
const HIDDEN = new Set(['kubectl.kubernetes.io/last-applied-configuration'])

const toRows = (m?: Record<string, string> | null) =>
  Object.entries(m || {})
    .filter(([k]) => !HIDDEN.has(k))
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([key, value]) => ({ key, value }))

const labelRows = computed(() => toRows(props.labels))
const annotationRows = computed(() => toRows(props.annotations))
</script>
