<template>
  <div class="page-hero">
    <!-- 右侧装饰插画: 纯装饰, 不参与交互 -->
    <svg class="page-hero__deco" viewBox="0 0 560 140" preserveAspectRatio="xMidYMid slice" aria-hidden="true" focusable="false">
      <defs>
        <linearGradient id="pageHeroHex" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stop-color="#a9cdff" />
          <stop offset="100%" stop-color="#4a88f5" />
        </linearGradient>
      </defs>
      <path d="M120 0 L270 0 L150 140 L0 140Z" fill="#ffffff" opacity="0.45" />
      <path d="M300 0 L350 0 L230 140 L180 140Z" fill="#ffffff" opacity="0.3" />
      <g opacity="0.6">
        <path d="M70 26 L82 32.0 L70 38 L58 32.0Z" fill="#f2f7ff" />
        <path d="M58 32.0 L70 38 L70 50 L58 44.0Z" fill="#cfe0fc" />
        <path d="M70 38 L82 32.0 L82 44.0 L70 50Z" fill="#b5cdf8" />
      </g>
      <g opacity="0.55">
        <path d="M118 74 L134 82.0 L118 90 L102 82.0Z" fill="#f2f7ff" />
        <path d="M102 82.0 L118 90 L118 106 L102 98.0Z" fill="#cfe0fc" />
        <path d="M118 90 L134 82.0 L134 98.0 L118 106Z" fill="#b5cdf8" />
      </g>
      <g opacity="0.45">
        <path d="M168 14 L178 19.0 L168 24 L158 19.0Z" fill="#f2f7ff" />
        <path d="M158 19.0 L168 24 L168 34 L158 29.0Z" fill="#cfe0fc" />
        <path d="M168 24 L178 19.0 L178 29.0 L168 34Z" fill="#b5cdf8" />
      </g>
      <g opacity="0.4">
        <path d="M214 96 L223 100.5 L214 105 L205 100.5Z" fill="#f2f7ff" />
        <path d="M205 100.5 L214 105 L214 114 L205 109.5Z" fill="#cfe0fc" />
        <path d="M214 105 L223 100.5 L223 109.5 L214 114Z" fill="#b5cdf8" />
      </g>
      <g opacity="0.55">
        <path d="M452 18 L466 25.0 L452 32 L438 25.0Z" fill="#f2f7ff" />
        <path d="M438 25.0 L452 32 L452 46 L438 39.0Z" fill="#cfe0fc" />
        <path d="M452 32 L466 25.0 L466 39.0 L452 46Z" fill="#b5cdf8" />
      </g>
      <g opacity="0.6">
        <path d="M500 70 L518 79.0 L500 88 L482 79.0Z" fill="#f2f7ff" />
        <path d="M482 79.0 L500 88 L500 106 L482 97.0Z" fill="#cfe0fc" />
        <path d="M500 88 L518 79.0 L518 97.0 L500 106Z" fill="#b5cdf8" />
      </g>
      <g opacity="0.45">
        <path d="M424 98 L434 103.0 L424 108 L414 103.0Z" fill="#f2f7ff" />
        <path d="M414 103.0 L424 108 L424 118 L414 113.0Z" fill="#cfe0fc" />
        <path d="M424 108 L434 103.0 L434 113.0 L424 118Z" fill="#b5cdf8" />
      </g>
      <polygon points="340.0,26.0 385.0,52.0 385.0,104.0 340.0,130.0 295.0,104.0 295.0,52.0" fill="#8fb8fb" opacity="0.35" />
      <polygon points="330.0,16.0 375.0,42.0 375.0,94.0 330.0,120.0 285.0,94.0 285.0,42.0" fill="url(#pageHeroHex)" opacity="0.7" />
      <g fill="none" stroke="#ffffff" stroke-width="5" stroke-linecap="round" opacity="0.9">
        <circle cx="330" cy="68" r="22" />
        <line x1="330.0" y1="60.0" x2="330.0" y2="37.0" /><line x1="336.3" y1="63.0" x2="354.2" y2="48.7" /><line x1="337.8" y1="69.8" x2="360.2" y2="74.9" /><line x1="333.5" y1="75.2" x2="343.5" y2="95.9" /><line x1="326.5" y1="75.2" x2="316.5" y2="95.9" /><line x1="322.2" y1="69.8" x2="299.8" y2="74.9" /><line x1="323.7" y1="63.0" x2="305.8" y2="48.7" />
      </g>
      <circle cx="330" cy="68" r="6" fill="#ffffff" opacity="0.9" />
    </svg>

    <div class="page-hero__main">
      <button v-if="back" type="button" class="page-hero__back" title="返回" aria-label="返回" @click="emit('back')">
        <el-icon><ArrowLeft /></el-icon>
      </button>
      <div class="page-hero__icon">
        <K8sIcon :name="icon" :size="56" />
      </div>
      <div class="page-hero__text">
        <div class="page-hero__title-row">
          <h1 class="page-hero__title" :class="{ 'is-deprecated': deprecated }">{{ title }}</h1>
          <div v-if="$slots.meta" class="page-hero__meta">
            <slot name="meta" />
          </div>
        </div>
        <p v-if="subtitle" class="page-hero__subtitle">{{ subtitle }}</p>
      </div>
    </div>

    <div v-if="$slots.default" class="page-hero__actions">
      <slot />
    </div>

    <div v-if="$slots.extra" class="page-hero__extra">
      <slot name="extra" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ArrowLeft } from '@element-plus/icons-vue'
import K8sIcon from '@/components/K8sIcon.vue'

// 列表页 / 详情页统一页头: 左侧(返回按钮) + 资源图标 + 标题/副标题, 中部装饰插画, 右侧操作按钮(默认插槽)
// 插槽: default 操作按钮; meta 标题右侧的状态标签; extra 页头下方的提示信息
withDefaults(
  defineProps<{
    title: string
    subtitle?: string
    icon?: any
    deprecated?: boolean
    back?: boolean
  }>(),
  {
    subtitle: '',
    icon: 'k8s',
    deprecated: false,
    back: false
  }
)

const emit = defineEmits<{ back: [] }>()
</script>

<style scoped>
.page-hero {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 16px 24px;
  min-height: 112px;
  margin-bottom: 18px;
  padding: 24px 32px;
  border: 1px solid #e3ecfb;
  border-radius: 18px;
  background: linear-gradient(105deg, #ffffff 0%, #f6f9ff 42%, #e9f1ff 100%);
  box-shadow: 0 6px 24px rgba(37, 99, 235, 0.06);
  overflow: hidden;
}

.page-hero__deco {
  position: absolute;
  top: 0;
  right: 220px;
  width: 560px;
  height: 100%;
  pointer-events: none;
  -webkit-mask-image: linear-gradient(90deg, transparent 0, #000 18%, #000 82%, transparent 100%);
  mask-image: linear-gradient(90deg, transparent 0, #000 18%, #000 82%, transparent 100%);
}

.page-hero__main {
  position: relative;
  z-index: 1;
  display: flex;
  flex: 1 1 320px;
  align-items: center;
  gap: 18px;
  min-width: 0;
}

.page-hero__back {
  display: grid;
  place-items: center;
  flex: none;
  width: 40px;
  height: 40px;
  padding: 0;
  border: 1px solid #dbe5f6;
  border-radius: 50%;
  background: #ffffff;
  color: #3b74f2;
  font-size: 18px;
  cursor: pointer;
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.08);
  transition: all 0.15s ease;
}

.page-hero__back:hover {
  color: #ffffff;
  border-color: transparent;
  background: linear-gradient(135deg, #5b9cff 0%, #3b74f2 100%);
}

.page-hero__back:focus-visible {
  outline: 2px solid #3b74f2;
  outline-offset: 2px;
}

.page-hero__icon {
  display: grid;
  place-items: center;
  flex: none;
  width: 60px;
  height: 60px;
  filter: drop-shadow(0 6px 12px rgba(37, 99, 235, 0.25));
}

.page-hero__text {
  min-width: 0;
}

.page-hero__title-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px 12px;
}

.page-hero__meta {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.page-hero__meta :deep(.el-tag) {
  height: 26px;
  padding: 0 12px;
  border-radius: 999px !important;
  font-size: 13px;
  font-weight: 500;
}

.page-hero__title {
  margin: 0;
  font-size: 24px;
  font-weight: 700;
  line-height: 1.3;
  color: #1e293b;
  overflow-wrap: anywhere;
}

.page-hero__title.is-deprecated {
  color: #94a3b8;
  text-decoration: line-through;
}

.page-hero__subtitle {
  margin: 6px 0 0;
  font-size: 14px;
  color: #64748b;
}

.page-hero__actions {
  position: relative;
  z-index: 1;
  display: flex;
  flex-wrap: wrap;
  gap: 14px;
}

.page-hero__extra {
  position: relative;
  z-index: 1;
  flex-basis: 100%;
}

/* 页头按钮: 主操作为白底蓝字胶囊, YAML 添加为蓝色渐变胶囊
   (选择器较长是为了覆盖 global.css 中带 !important 的全局按钮样式) */
.page-hero__actions :deep(.el-button) {
  height: 42px;
  margin: 0 !important;
  padding: 0 22px !important;
  border-radius: 999px !important;
  font-size: 15px;
  font-weight: 500;
}

.page-hero__actions :deep(.el-button .el-icon) {
  font-size: 17px;
}

.page-hero__actions :deep(.el-button.el-button--primary:not(.is-plain):not(.is-text):not(.is-link)) {
  color: #2563eb !important;
  background: #ffffff !important;
  border: 1px solid #d9e5fb !important;
  box-shadow: 0 4px 14px rgba(37, 99, 235, 0.1) !important;
}

.page-hero__actions :deep(.el-button.el-button--primary:not(.is-plain):not(.is-text):not(.is-link):hover) {
  background: #f3f7ff !important;
  border-color: #b9cff8 !important;
}

.page-hero__actions :deep(.el-button.el-button--success:not(.is-plain):not(.is-text):not(.is-link)) {
  color: #ffffff !important;
  background: linear-gradient(135deg, #5b9cff 0%, #3b74f2 100%) !important;
  border: none !important;
  box-shadow: 0 8px 18px rgba(59, 116, 242, 0.3) !important;
}

.page-hero__actions :deep(.el-button.el-button--success:not(.is-plain):not(.is-text):not(.is-link):hover) {
  filter: brightness(1.05);
}

.page-hero__actions :deep(.el-button.el-button--default) {
  color: #334155;
  background: #ffffff;
  border: 1px solid #dfe6f2;
}

.page-hero__actions :deep(.el-button.el-button--default:not(.is-disabled):hover) {
  color: #2563eb;
  background: #f5f9ff;
  border-color: #b9cdf3;
}

/* 删除 / 驱逐类危险操作: 白底彩色描边, 避免页头出现大块红色 */
.page-hero__actions :deep(.el-button.el-button--danger) {
  color: #ef4444 !important;
  background: #ffffff !important;
  border: 1px solid #fca5a5 !important;
  box-shadow: none !important;
}

.page-hero__actions :deep(.el-button.el-button--danger:hover) {
  color: #dc2626 !important;
  background: #fef2f2 !important;
  border-color: #f87171 !important;
}

.page-hero__actions :deep(.el-button.el-button--warning) {
  color: #d97706 !important;
  background: #ffffff !important;
  border: 1px solid #fcd34d !important;
  box-shadow: none !important;
}

.page-hero__actions :deep(.el-button.el-button--warning:hover) {
  color: #b45309 !important;
  background: #fffbeb !important;
  border-color: #fbbf24 !important;
}

@media (max-width: 1200px) {
  .page-hero__deco {
    right: 0;
    opacity: 0.55;
  }
}

@media (max-width: 768px) {
  .page-hero {
    padding: 18px;
  }

  .page-hero__deco {
    display: none;
  }
}
</style>
