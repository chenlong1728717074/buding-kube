<template>
  <div>
    <div v-for="(v, i) in volumes" :key="'vol' + i" class="wl-volume">
      <div class="wl-volume__head">
        <span class="wl-volume__title">{{ v.name || `存储卷 ${i + 1}` }}</span>
        <el-tag size="small">{{ volumeTypeText(v.type) }}</el-tag>
        <el-button link type="danger" @click="volumes.splice(i, 1)"><el-icon><Delete /></el-icon>删除</el-button>
      </div>
      <div class="fd-choice-grid wl-volume-types">
        <button v-for="t in VOLUME_TYPES" :key="t.value" type="button" :class="['fd-choice', { 'is-active': v.type === t.value }]" @click="setType(v, t.value)">
          <span class="fd-choice__body">
            <span class="fd-choice__title"><span>{{ t.label }}</span></span>
            <span class="fd-choice__desc">{{ t.desc }}</span>
          </span>
          <el-icon class="fd-choice__check"><CircleCheckFilled /></el-icon>
        </button>
      </div>
      <el-form :model="v" label-width="90px" class="wl-volume__form">
        <el-form-item label="名称"><el-input v-model="v.name" placeholder="例如 config" /></el-form-item>
        <el-form-item v-if="v.type === 'pvc'" label="持久卷声明">
          <el-select v-model="v.source" filterable placeholder="选择持久卷声明" style="width: 100%" :loading="loading">
            <el-option v-for="o in pvcOptions" :key="o" :label="o" :value="o" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="v.type === 'pvc'" label="只读"><el-switch v-model="v.readOnly" /></el-form-item>
        <el-form-item v-if="v.type === 'configMap'" label="ConfigMap">
          <el-select v-model="v.source" filterable placeholder="选择 ConfigMap" style="width: 100%" :loading="loading">
            <el-option v-for="o in configMapOptions" :key="o" :label="o" :value="o" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="v.type === 'secret'" label="Secret">
          <el-select v-model="v.source" filterable placeholder="选择 Secret" style="width: 100%" :loading="loading">
            <el-option v-for="o in secretOptions" :key="o" :label="o" :value="o" />
          </el-select>
        </el-form-item>
        <template v-if="v.type === 'emptyDir'">
          <el-form-item label="介质">
            <el-radio-group v-model="v.medium">
              <el-radio-button value="">节点磁盘</el-radio-button>
              <el-radio-button value="Memory">内存</el-radio-button>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="容量上限"><el-input v-model="v.sizeLimit" placeholder="可选，如 1Gi" /></el-form-item>
        </template>
        <template v-if="v.type === 'hostPath'">
          <el-form-item label="主机路径"><el-input v-model="v.source" placeholder="例如 /var/log" class="mono-input" /></el-form-item>
          <el-form-item label="路径类型">
            <el-select v-model="v.hostPathType" style="width: 100%">
              <el-option label="不检查" value="" />
              <el-option label="目录，不存在时创建（DirectoryOrCreate）" value="DirectoryOrCreate" />
              <el-option label="已存在的目录（Directory）" value="Directory" />
              <el-option label="文件，不存在时创建（FileOrCreate）" value="FileOrCreate" />
              <el-option label="已存在的文件（File）" value="File" />
              <el-option label="Socket" value="Socket" />
            </el-select>
          </el-form-item>
        </template>
      </el-form>
    </div>
    <el-button class="wl-add" @click="volumes.push(createVolumeForm(volumes.length))">
      <el-icon><Plus /></el-icon>
      添加存储卷
    </el-button>
  </div>
</template>

<script setup lang="ts">
import { CircleCheckFilled, Delete, Plus } from '@element-plus/icons-vue'
import { VOLUME_TYPES, createVolumeForm, volumeTypeText, type VolumeForm } from '@/utils/workload'

// 存储卷列表编辑, 工作负载和容器组创建共用; volumes 由父组件创建, 这里直接增删和编辑
defineProps<{
  volumes: VolumeForm[]
  pvcOptions: string[]
  configMapOptions: string[]
  secretOptions: string[]
  loading?: boolean
}>()

const setType = (v: VolumeForm, type: VolumeForm['type']) => {
  if (v.type !== type) v.source = ''
  v.type = type
}
</script>

<style scoped>
.wl-volume {
  margin-bottom: 14px;
  padding: 14px 16px 2px;
  border: 1px solid #e3eaf4;
  border-radius: 12px;
  background: #fbfcfe;
}

.wl-volume__head {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}

.wl-volume__head .el-button {
  margin-left: auto;
}

.wl-volume__title {
  color: #1e293b;
  font-weight: 600;
}

.wl-volume__form {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 16px;
}

.wl-volume-types {
  grid-template-columns: repeat(5, minmax(0, 1fr));
  margin-bottom: 14px;
}

.wl-volume-types .fd-choice {
  padding: 10px 12px;
}

.wl-add {
  width: 100%;
  margin-bottom: 18px;
  border-style: dashed;
}

.mono-input :deep(input) {
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
}
</style>
