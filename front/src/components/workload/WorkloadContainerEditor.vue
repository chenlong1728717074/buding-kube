<template>
  <el-form :model="c" label-width="100px" label-position="right" class="wc-form">
    <div class="wc-row">
      <el-form-item label="容器名称" required>
        <el-input v-model="c.name" placeholder="例如 nginx" maxlength="63" />
      </el-form-item>
      <el-form-item label="容器类型">
        <el-radio-group v-model="c.init">
          <el-radio-button :value="false">工作容器</el-radio-button>
          <el-radio-button :value="true">初始化容器</el-radio-button>
        </el-radio-group>
      </el-form-item>
    </div>
    <el-form-item label="镜像" required>
      <el-input v-model="c.image" placeholder="例如 nginx:1.27、registry.example.com/team/app:v1" class="mono-input" />
      <div v-if="c.init" class="field-hint">初始化容器在工作容器启动前按顺序运行，运行成功后才会启动工作容器</div>
    </el-form-item>
    <el-form-item label="拉取策略">
      <el-radio-group v-model="c.imagePullPolicy">
        <el-radio-button value="IfNotPresent">本地没有时拉取</el-radio-button>
        <el-radio-button value="Always">每次都拉取</el-radio-button>
        <el-radio-button value="Never">只使用本地镜像</el-radio-button>
      </el-radio-group>
    </el-form-item>

    <p class="fd-subtitle">端口设置<small>容器对外提供服务的端口，创建服务时会用到</small></p>
    <div class="wc-table">
      <div v-for="(p, i) in c.ports" :key="i" class="wc-table__row">
        <el-select v-model="p.protocol" style="width: 100px">
          <el-option label="TCP" value="TCP" />
          <el-option label="UDP" value="UDP" />
          <el-option label="SCTP" value="SCTP" />
        </el-select>
        <el-input v-model="p.name" placeholder="名称（可选），如 http" maxlength="15" style="width: 200px" />
        <el-input-number v-model="p.containerPort" :min="1" :max="65535" placeholder="容器端口" controls-position="right" style="width: 160px" />
        <el-button link type="danger" @click="c.ports.splice(i, 1)"><el-icon><Delete /></el-icon></el-button>
      </div>
      <el-button size="small" @click="c.ports.push({ name: '', containerPort: undefined, protocol: 'TCP' })">
        <el-icon><Plus /></el-icon>
        添加端口
      </el-button>
    </div>

    <p class="fd-subtitle">启动命令</p>
    <el-form-item label="命令方式">
      <el-radio-group v-model="c.commandMode">
        <el-radio-button value="default">使用镜像默认</el-radio-button>
        <el-radio-button value="shell">Shell 脚本</el-radio-button>
        <el-radio-button value="exec">命令与参数</el-radio-button>
      </el-radio-group>
    </el-form-item>
    <el-form-item v-if="c.commandMode === 'shell'" label="脚本">
      <el-input v-model="c.script" type="textarea" :rows="3" placeholder="以 /bin/sh -c 执行" class="mono-input" />
    </el-form-item>
    <template v-else-if="c.commandMode === 'exec'">
      <el-form-item label="命令">
        <el-input v-model="c.commandText" type="textarea" :rows="2" placeholder="每行一项" class="mono-input" />
      </el-form-item>
      <el-form-item label="参数">
        <el-input v-model="c.argsText" type="textarea" :rows="2" placeholder="每行一项" class="mono-input" />
      </el-form-item>
    </template>
    <el-form-item label="工作目录">
      <el-input v-model="c.workingDir" placeholder="可选，例如 /app" class="mono-input" />
    </el-form-item>

    <p class="fd-subtitle">资源配额<small>预留是调度时保证的资源，上限是最多可使用的资源；留空表示不限制</small></p>
    <div class="wc-row wc-row--4">
      <el-form-item label="CPU 预留"><el-input v-model="c.cpuRequest" placeholder="如 100m" /></el-form-item>
      <el-form-item label="CPU 上限"><el-input v-model="c.cpuLimit" placeholder="如 1" /></el-form-item>
      <el-form-item label="内存预留"><el-input v-model="c.memoryRequest" placeholder="如 128Mi" /></el-form-item>
      <el-form-item label="内存上限"><el-input v-model="c.memoryLimit" placeholder="如 512Mi" /></el-form-item>
    </div>

    <KVRowsTable v-model="c.env" title="环境变量" add-text="添加环境变量" empty-text="暂无环境变量" key-placeholder="变量名" value-placeholder="值" />

    <template v-if="!c.init">
      <p class="fd-subtitle">健康检查</p>
      <ProbeEditor :probe="c.readiness" title="就绪检查" desc="检查通过后才会接收服务流量" allow-success :port-options="portNumbers" />
      <ProbeEditor :probe="c.liveness" title="存活检查" desc="检查失败时重启容器" :port-options="portNumbers" />
      <ProbeEditor :probe="c.startup" title="启动检查" desc="启动较慢的应用使用，通过前不执行其他检查" :port-options="portNumbers" />
    </template>

    <p class="fd-subtitle">
      存储挂载
      <small>{{ volumeNames.length ? '把「存储设置」中的存储卷挂载到容器的目录' : '先在「存储设置」中添加存储卷' }}</small>
    </p>
    <div class="wc-table">
      <div v-for="(m, i) in c.mounts" :key="i" class="wc-table__row">
        <el-select v-model="m.volume" placeholder="存储卷" style="width: 180px">
          <el-option v-for="v in volumeNames" :key="v" :label="v" :value="v" />
        </el-select>
        <el-input v-model="m.mountPath" placeholder="挂载路径，如 /data" class="mono-input" style="width: 220px" />
        <el-input v-model="m.subPath" placeholder="子路径（可选）" class="mono-input" style="width: 160px" />
        <el-checkbox v-model="m.readOnly">只读</el-checkbox>
        <el-button link type="danger" @click="c.mounts.splice(i, 1)"><el-icon><Delete /></el-icon></el-button>
      </div>
      <el-button size="small" :disabled="!volumeNames.length" @click="c.mounts.push({ volume: volumeNames[0] || '', mountPath: '', subPath: '', readOnly: false })">
        <el-icon><Plus /></el-icon>
        添加挂载
      </el-button>
    </div>
  </el-form>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Delete, Plus } from '@element-plus/icons-vue'
import KVRowsTable from '@/components/KVRowsTable.vue'
import ProbeEditor from '@/components/workload/ProbeEditor.vue'
import type { WorkloadContainerForm } from '@/utils/workload'

// 表单对象由父组件创建, 这里直接编辑它的字段
const props = defineProps<{
  container: WorkloadContainerForm
  volumeNames: string[]
}>()

const c = computed(() => props.container)
const portNumbers = computed(() => props.container.ports.map(p => p.containerPort).filter((p): p is number => !!p))
</script>

<style scoped>
.wc-row {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 16px;
}

.wc-row--4 {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.wc-row--4 :deep(.el-form-item__label) {
  width: 80px !important;
}

.wc-table {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 8px;
  margin: 0 0 18px 100px;
}

.wc-table__row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.wc-form :deep(.fd-subtitle) {
  margin-top: 6px;
}

.mono-input :deep(input),
.mono-input :deep(textarea) {
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
}
</style>
