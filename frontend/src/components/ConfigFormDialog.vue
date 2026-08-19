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
        <GridEditor v-model="form.grid" />
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
import { ref, watch } from 'vue'
import GridEditor from './GridEditor.vue'
import { DEFAULT_EXAMPLE } from '../palette'

const props = defineProps({
  visible: Boolean,
  mode: { type: String, default: 'new' },
  initialData: { type: Object, default: () => ({ name: '', grid: '', cartridgeId: 0 }) },
  cartridges: { type: Array, default: () => [] },
  saving: { type: Boolean, default: false },
})

const emit = defineEmits(['update:visible', 'save', 'cancel', 'reset'])

const form = ref({ name: '', grid: '', cartridgeId: 0 })

watch(() => props.visible, (val) => {
  if (val) {
    if (props.mode === 'edit' && props.initialData) {
      form.value = { name: props.initialData.name, grid: props.initialData.grid, cartridgeId: props.initialData.cartridgeId || 0 }
    } else {
      form.value = { name: '', grid: DEFAULT_EXAMPLE, cartridgeId: 0 }
    }
  }
})

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
</style>