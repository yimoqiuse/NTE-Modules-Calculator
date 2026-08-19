<template>
  <div class="g-editor">
    <div class="editor-tools">
      <span class="dim">行</span>
      <el-input-number v-model="rows" :min="1" :max="20" size="small" style="width: 84px" />
      <span class="dim">列</span>
      <el-input-number v-model="cols" :min="1" :max="20" size="small" style="width: 84px" />
      <el-button size="small" @click="applySize">重设大小</el-button>
      <el-button size="small" @click="loadExample">载入示例</el-button>
      <el-button size="small" @click="clear">清空</el-button>
    </div>
    <table class="edit-grid">
      <tr v-for="(row, r) in cells" :key="r">
        <td v-for="(v, c) in row" :key="c" :class="{ on: v }" @click="toggle(r, c)"></td>
      </tr>
    </table>
    <div class="hint">点击格子填色 = 总模型的形状（可填区域），至少要有 1 个格子。</div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { DEFAULT_EXAMPLE } from '../palette'

const props = defineProps({
  modelValue: { type: String, default: '' },
})
const emit = defineEmits(['update:modelValue'])

const rows = ref(7)
const cols = ref(7)
const cells = ref([])

function fromText(text) {
  const lines = String(text || '')
    .split('\n')
    .filter((l) => l.trim() !== '')
  const r = lines.length || 7
  const c = lines.length ? Math.max(...lines.map((l) => l.length)) : 7
  const uneven = lines.some((l) => l.length !== c)
  cells.value = Array.from({ length: r }, (_, i) =>
    Array.from({ length: c }, (_, j) => (lines[i] && lines[i][j] === '1' ? 1 : 0)),
  )
  rows.value = r
  cols.value = c
  if (uneven && lines.length) {
    ElMessage.warning('粘贴的各行列数不一致，已按最长行补齐为 0')
  }
}

function toText() {
  return cells.value.map((row) => row.map((v) => (v ? '1' : '0')).join('')).join('\n')
}

function emitText() {
  emit('update:modelValue', toText())
}

function toggle(r, c) {
  cells.value[r][c] = cells.value[r][c] ? 0 : 1
  emitText()
}

function applySize() {
  const r = rows.value
  const c = cols.value
  cells.value = Array.from({ length: r }, (_, i) =>
    Array.from({ length: c }, (_, j) => (cells.value[i] && cells.value[i][j] ? 1 : 0)),
  )
  emitText()
}

function loadExample() {
  fromText(DEFAULT_EXAMPLE)
  emitText()
}

function clear() {
  cells.value = Array.from({ length: rows.value }, () => Array(cols.value).fill(0))
  emitText()
}

watch(
  () => props.modelValue,
  (v) => {
    if (v !== toText()) fromText(v)
  },
  { immediate: true },
)
</script>