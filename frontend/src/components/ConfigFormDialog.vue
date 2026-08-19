<template>
  <el-dialog
    :model-value="visible"
    @update:model-value="$emit('update:visible', $event)"
    :title="mode === 'new' ? '新建配置' : '编辑配置'"
    width="500px"
    :close-on-click-modal="false"
  >
    <el-form label-position="top">
      <div class="cfg-top-row">
        <el-form-item label="配置名称" class="cfg-top-item">
          <el-input
            v-model="form.name"
            placeholder="给模型起个名字"
          />
        </el-form-item>
        <el-form-item label="关联卡带" class="cfg-top-item">
          <el-select
            v-model="form.cartridgeId"
            placeholder="选择关联卡带"
            clearable
          >
            <el-option v-for="c in cartridges" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>
      </div>
      <el-form-item label="总模型形状">
        <div class="hist-nav">
          <el-button circle :disabled="!historyItems.length" title="上一个历史形状" @click="stepHistory(-1)">
            <el-icon><ArrowLeftBold /></el-icon>
          </el-button>
          <div class="hist-main">
            <GridEditor v-model="form.grid" />
            <div v-if="historyItems.length" class="hist-status">{{ histLabel }}</div>
          </div>
          <el-button circle :disabled="!historyItems.length" title="下一个历史形状" @click="stepHistory(1)">
            <el-icon><ArrowRightBold /></el-icon>
          </el-button>
        </div>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="onReset">重置</el-button>
      <el-button @click="onCancel">取消</el-button>
      <el-button type="primary" :loading="saving" @click="onSave">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, watch, computed, onBeforeUnmount } from 'vue'
import GridEditor from './GridEditor.vue'
import { DEFAULT_EXAMPLE } from '../palette'

const props = defineProps({
  visible: Boolean,
  mode: { type: String, default: 'new' },
  initialData: { type: Object, default: () => ({ name: '', grid: '', cartridgeId: 0 }) },
  cartridges: { type: Array, default: () => [] },
  saving: { type: Boolean, default: false },
  history: { type: Array, default: () => [] },
})

const emit = defineEmits(['update:visible', 'save', 'cancel', 'reset'])

const form = ref({ name: '', grid: '', cartridgeId: 0 })

// 历史形状 = 已有配置里的形状，按形状去重
const historyItems = computed(() => {
  const seen = new Set()
  const out = []
  for (const c of props.history || []) {
    if (!c.grid || seen.has(c.grid)) continue
    seen.add(c.grid)
    out.push({ name: c.name, grid: c.grid })
  }
  return out
})

const curHist = computed(() => historyItems.value.findIndex((h) => h.grid === form.value.grid))

const histLabel = computed(() => {
  if (!historyItems.value.length) return ''
  const i = curHist.value
  return i === -1 ? '当前形状不在历史中（←/→ 切换）' : `历史形状 ${i + 1} / ${historyItems.value.length}`
})

function stepHistory(dir) {
  const items = historyItems.value
  if (!items.length) return
  const cur = curHist.value
  const next = cur === -1 ? (dir > 0 ? 0 : items.length - 1) : (cur + dir + items.length) % items.length
  form.value.grid = items[next].grid
  form.value.name = items[next].name
}

function onKeydown(e) {
  if (e.key !== 'ArrowLeft' && e.key !== 'ArrowRight') return
  const t = e.target
  if (t && (t.tagName === 'INPUT' || t.tagName === 'TEXTAREA' || t.tagName === 'SELECT' || t.isContentEditable)) return
  if (t instanceof Element && t.closest('.el-select-dropdown')) return
  e.preventDefault()
  stepHistory(e.key === 'ArrowLeft' ? -1 : 1)
}

watch(() => props.visible, (val) => {
  if (val) {
    if (props.mode === 'edit' && props.initialData) {
      form.value = { name: props.initialData.name, grid: props.initialData.grid, cartridgeId: props.initialData.cartridgeId || 0 }
    } else {
      form.value = { name: '', grid: DEFAULT_EXAMPLE, cartridgeId: 0 }
    }
    document.addEventListener('keydown', onKeydown)
  } else {
    document.removeEventListener('keydown', onKeydown)
  }
})

onBeforeUnmount(() => document.removeEventListener('keydown', onKeydown))

function onSave() {
  emit('save', { ...form.value })
}

function onCancel() {
  emit('cancel')
  emit('update:visible', false)
}

function onReset() {
  if (props.mode === 'new') {
    form.value = { name: '', grid: DEFAULT_EXAMPLE, cartridgeId: 0 }
  } else {
    form.value = { name: props.initialData.name, grid: props.initialData.grid, cartridgeId: props.initialData.cartridgeId || 0 }
  }
  emit('reset')
}
</script>

<style scoped>
.cfg-top-row {
  display: flex;
  gap: 12px;
}

.cfg-top-item {
  flex: 1;
  min-width: 0;
}

.hist-nav {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
}

.hist-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.hist-status {
  margin-top: 6px;
  color: #909399;
  font-size: var(--fs-hint);
}
</style>