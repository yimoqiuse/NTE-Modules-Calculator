<template>
  <el-container class="app-root">
    <el-aside width="240px" class="aside">
      <el-scrollbar class="menu-scroll">
        <el-menu
          :default-active="activeMenu"
          class="cfg-menu"
          background-color="#141c28"
          text-color="#b7c0cd"
          active-text-color="#ffffff"
          @select="onSelect"
        >
          <el-menu-item index="new">
            <span>配置管理</span>
          </el-menu-item>
          <el-menu-item index="cartridge">
            <span>卡带管理</span>
          </el-menu-item>
          <el-menu-item index="divider" :disabled="true" class="menu-divider"></el-menu-item>
          <div class="cfg-search">
            <el-input
              v-model="searchQuery"
              placeholder="搜索配置名称"
              clearable
              size="large"
            >
              <template #prefix><el-icon><Search /></el-icon></template>
            </el-input>
          </div>
          <el-menu-item v-for="c in filteredConfigs" :key="c.id" :index="String(c.id)">
            <span class="cfg-name">{{ c.name }}</span>
          </el-menu-item>
        </el-menu>
      </el-scrollbar>
      <el-empty
        v-if="!configs.length"
        description="还没有配置"
        :image-size="60"
        class="aside-empty"
      />
      <el-empty
        v-else-if="!filteredConfigs.length"
        description="没有匹配的配置"
        :image-size="50"
        class="aside-empty"
      />
    </el-aside>

    <el-main class="main">
      <!-- 空状态 -->
      <div v-if="view === 'empty'" class="welcome">
        <el-empty description="点击左侧「配置管理」管理配置，或直接创建第一个模型">
          <el-button type="primary" @click="openNew">
            <el-icon><Plus /></el-icon>&nbsp;新增配置
          </el-button>
        </el-empty>
      </div>

      <!-- 配置列表 -->
      <div v-else-if="view === 'form'" class="config-page">

        <el-card shadow="never" class="list-card">
          <template #header>
            <span style="display: flex; align-items: center; justify-content: space-between;">
              <b>已有配置</b>
              <el-button type="primary" @click="openNew">
                <el-icon><Plus /></el-icon>&nbsp;新增配置
              </el-button>
            </span>
          </template>
          <div v-if="!configs.length" class="list-empty">
            <el-empty description="还没有配置" :image-size="60" />
          </div>
          <div v-else class="sortable-list">
            <div
              v-for="(c, i) in configs"
              :key="c.id"
              class="sortable-item"
              :class="{ 'drag-over': dragOverIndex === i && dragType === 'config' }"
              draggable="true"
              @dragstart="onDragStart($event, i, 'config')"
              @dragover.prevent="onDragOver($event, i)"
              @dragenter.prevent
              @drop="onDrop($event, i, 'config')"
              @dragend="onDragEnd"
              @click="selectConfig(c.id)"
            >
              <el-icon class="drag-handle"><Rank /></el-icon>
              <div class="sortable-item-info">
                <div class="sortable-item-name">{{ c.name }}</div>
                <div v-if="cartridgeName(c.cartridgeId)" class="sortable-item-tags">
                  <span class="piece-tag">卡带：{{ cartridgeName(c.cartridgeId) }}</span>
                </div>
              </div>
              <div class="sortable-item-actions" @click.stop>
                <el-button size="small" text @click="editConfigFromList(c)">编辑</el-button>
                <el-button size="small" type="danger" plain @click="remove(c.id)">删除</el-button>
              </div>
            </div>
          </div>
        </el-card>
      </div>

      <!-- 卡带管理 -->
      <div v-else-if="view === 'cartridge'" class="cartridge-page">
        <el-card shadow="never">
          <template #header><b>{{ cartFormMode === 'new' ? '添加卡带' : '编辑卡带' }}</b></template>
          <div class="cartridge-add">
            <el-input v-model="cartForm.name" placeholder="卡带名称" style="max-width: 240px" />
            <div class="cartridge-piece-select">
              <span class="cartridge-piece-label">预选方块（卡带将只筛选这些方块）：</span>
              <div class="cartridge-piece-chips">
                <button
                  v-for="p in pieces"
                  :key="p.name"
                  type="button"
                  class="piece-chip"
                  :class="{ active: cartForm.pieces.includes(p.name) }"
                  @click="toggleCartPiece(p.name)"
                >
                  <GridDisplay :text="p.shape" :size="10" />
                  <span class="piece-name">{{ p.name }}</span>
                </button>
              </div>
            </div>
            <div>
              <el-button type="primary" :loading="saving" @click="saveCartridge">{{ cartFormMode === 'new' ? '添加卡带' : '保存' }}</el-button>
              <el-button v-if="cartFormMode === 'edit'" @click="resetCartForm">取消</el-button>
            </div>
          </div>
        </el-card>

        <el-card shadow="never" class="list-card">
          <template #header><b>已有卡带</b></template>
          <div v-if="!cartridges.length" class="list-empty">
            <el-empty description="还没有卡带" :image-size="60" />
          </div>
          <div v-else class="sortable-list">
            <div
              v-for="(c, i) in cartridges"
              :key="c.id"
              class="sortable-item"
              :class="{ 'drag-over': dragOverIndex === i && dragType === 'cartridge' }"
              draggable="true"
              @dragstart="onDragStart($event, i, 'cartridge')"
              @dragover.prevent="onDragOver($event, i)"
              @dragenter.prevent
              @drop="onDrop($event, i, 'cartridge')"
              @dragend="onDragEnd"
            >
              <el-icon class="drag-handle"><Rank /></el-icon>
              <div class="sortable-item-info">
                <div class="sortable-item-name">{{ c.name }}</div>
                <div class="sortable-item-tags">
                  <span v-for="pn in parseCartridgePieces(c.pieces)" :key="pn" class="piece-tag">{{ pn }}</span>
                  <span v-if="!parseCartridgePieces(c.pieces).length" class="no-tags">未限制方块</span>
                </div>
              </div>
              <div class="sortable-item-actions" @click.stop>
                <el-button size="small" text @click="editCartridgeFromList(c)">编辑</el-button>
                <el-button size="small" type="danger" plain @click="removeCartridge(c.id)">删除</el-button>
              </div>
            </div>
          </div>
        </el-card>
      </div>

      <!-- 详情：方法列表 -->
      <div v-else-if="view === 'detail' && detail" v-loading="loading" class="detail">
        <div class="detail-head">
          <div class="title-row">
            <h2>{{ detail.name }}</h2>
            <div class="head-actions">
              <el-button size="small" @click="editCurrentConfig">编辑</el-button>
              <el-button size="small" type="danger" plain @click="remove(detail.id)">
                删除
              </el-button>
            </div>
          </div>
        </div>

        <el-alert
          v-if="detail.warnings && detail.warnings.length"
          class="detail-warning"
          type="warning"
          :closable="false"
          show-icon
          :title="detail.warnings.join(' ')"
        />

        <div class="detail-tools">
          <el-card class="piece-card" shadow="never">
          <div class="piece-library">
            <span class="piece-label">方块库（点击添加筛选，右键取消1个，无需重新计算）</span>
            <div class="piece-chips">
              <button
                v-for="p in pieces"
                :key="p.name"
                type="button"
                class="piece-chip"
                :class="{ active: selFilter.includes(p.name) }"
                :title="'点击添加选中的方块（当前已选 ' + pieceCount(p.name) + ' 次）'"
                @click="addToFilter(p.name)"
                @contextmenu.prevent="removeFromFilter(p.name)"
              >
                <GridDisplay :text="p.shape" :size="10" />
                <span class="piece-name">{{ p.name }}</span>
                <span v-if="pieceCount(p.name) > 0" class="chip-count">{{ pieceCount(p.name) }}</span>
              </button>
            </div>
          </div>
          </el-card>
          <div class="shape-preview">
            <GridDisplay :text="detail.grid" :size="18" />
            <div class="cartridge-select-wrap">
              <el-select v-model="cartridgeFilter" placeholder="选择卡带一键筛选" clearable size="small" @change="onCartridgeFilterChange">
                <el-option v-for="c in cartridges" :key="c.id" :label="c.name" :value="c.id" />
              </el-select>
            </div>
          </div>
        </div>

        <div class="summary">
          <template v-if="selFilter.length">
            筛选后 <b>{{ pageTotal }}</b> 种 / 全部 {{ detail.total }} 种可行方块组合
            <el-button link type="danger" size="small" @click="clearFilter">清除筛选</el-button>
          </template>
          <template v-else>
            共找到 <b>{{ detail.total }}</b> 种可行方块组合
          </template>
        </div>

        <el-empty
          v-if="!pageCombos.length"
          description="当前筛选下没有可行的方块组合"
        />

        <template v-else>
          <div class="method-grid">
            <el-card v-for="(m, i) in pageCombos" :key="'c' + i" shadow="hover" class="method-card">
              <div class="method-label">{{ m.combo }}</div>
              <div class="method-grid-wrap">
                <GridDisplay :text="m.grid" />
              </div>
            </el-card>
          </div>
          <el-pagination
            class="method-pager"
            v-model:current-page="page"
            v-model:page-size="pageSize"
            :page-sizes="[4, 8, 20, 40]"
            :total="pageTotal"
            layout="total, sizes, prev, pager, next"
            @size-change="onPageSizeChange"
            @current-change="onPageChange"
          />
        </template>
      </div>
    </el-main>
  </el-container>

  <ConfigFormDialog
    v-model:visible="cfgDialogVisible"
    :mode="cfgFormMode"
    :initial-data="cfgEditData"
    :cartridges="cartridges"
    :history="configs"
    :saving="saving"
    @save="handleDialogSave"
    @cancel="handleDialogCancel"
    @reset="handleDialogReset"
  />
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { pinyin } from 'pinyin-pro'
import GridDisplay from './components/GridDisplay.vue'
import ConfigFormDialog from './components/ConfigFormDialog.vue'
import {
  listConfigs, listPieces, getConfigDetail, getConfigCombos,
  createConfig, updateConfig, deleteConfig,
  listCartridges, createCartridge, updateCartridge, deleteCartridge,
  reorderConfigs, reorderCartridges,
} from './api'
import { errMsg } from './palette'

const configs = ref([])
const pieces = ref([])
const selFilter = ref([])
const selectedId = ref(null)
const view = ref('empty')
const saving = ref(false)
const loading = ref(false)
const detail = ref(null)

// 侧边栏搜索：支持名称模糊匹配 + 拼音/首字母
const searchQuery = ref('')
const pyCache = new Map()
function pinyinText(name) {
  if (pyCache.has(name)) return pyCache.get(name)
  const first = pinyin(name, { pattern: 'first', toneType: 'none', nonZh: 'consecutive' }).replace(/\s+/g, '').toLowerCase()
  const full = pinyin(name, { toneType: 'none', nonZh: 'consecutive' }).replace(/\s+/g, '').toLowerCase()
  const v = { first, full }
  pyCache.set(name, v)
  return v
}
function matchConfigName(name, q) {
  if (!q) return true
  if (name.toLowerCase().includes(q)) return true
  const p = pinyinText(name)
  return p.first.includes(q) || p.full.includes(q)
}
const filteredConfigs = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return configs.value
  return configs.value.filter((c) => matchConfigName(c.name, q))
})

// 配置管理表单
const cfgFormMode = ref('new')
const cfgDialogVisible = ref(false)
const cfgEditData = ref({ name: '', grid: '', cartridgeId: 0 })
const editingCfgId = ref(null)

// 卡带管理
const cartridges = ref([])
const cartridgeFilter = ref(null)
const cartFormMode = ref('new')
const cartForm = ref({ name: '', pieces: [] })
const editingCartId = ref(null)

// 分页：结果由后端按页返回，前端只持有当前页
const page = ref(1)
const pageSize = ref(Number(localStorage.getItem('pageSize') || 4))
const pageTotal = ref(0)
const pageCombos = ref([])
let loadToken = 0

// 拖拽
const dragType = ref(null)  // 'config' | 'cartridge'
const dragIndex = ref(-1)
const dragOverIndex = ref(-1)

async function reloadCombos() {
  const token = ++loadToken
  if (!detail.value) return
  loading.value = true
  try {
    const res = await getConfigCombos(selectedId.value, selFilter.value, page.value, pageSize.value)
    if (token !== loadToken) return // 已有更新的请求，丢弃过期结果
    pageCombos.value = res.combos
    pageTotal.value = res.total
  } catch (e) {
    if (token === loadToken) ElMessage.error(errMsg(e))
  } finally {
    if (token === loadToken) loading.value = false
  }
}

function onPageChange() {
  reloadCombos()
}

function onPageSizeChange(val) {
  localStorage.setItem('pageSize', String(val))
  page.value = 1
  reloadCombos()
}

function clearFilter() {
  selFilter.value = []
  cartridgeFilter.value = null
  page.value = 1
  reloadCombos()
}

const activeMenu = computed(() => (selectedId.value ? String(selectedId.value) : ''))

async function loadConfigs() {
  try {
    configs.value = await listConfigs()
  } catch (e) {
    ElMessage.error(errMsg(e))
  }
}

function onSelect(index) {
  if (index === 'new') {
    openConfigManager()
  } else if (index === 'cartridge') {
    openCartridge()
  } else if (index === 'divider') {
    return
  } else {
    selectConfig(Number(index))
  }
}

async function selectConfig(id) {
  selectedId.value = id
  await showDetail()
}

async function showDetail() {
  if (!selectedId.value) return
  loading.value = true
  try {
    detail.value = await getConfigDetail(selectedId.value)
    if (detail.value.cartridgeId) {
      applyCartridgeFilter(detail.value.cartridgeId)
    } else {
      selFilter.value = []
      cartridgeFilter.value = null
    }
    page.value = 1
    view.value = 'detail'
    await reloadCombos()
  } catch (e) {
    ElMessage.error(errMsg(e))
  } finally {
    loading.value = false
  }
}

function addToFilter(name) {
  selFilter.value.push(name)
  cartridgeFilter.value = null
  page.value = 1
  reloadCombos()
}

function removeFromFilter(name) {
  const i = selFilter.value.indexOf(name)
  if (i !== -1) selFilter.value.splice(i, 1)
  cartridgeFilter.value = null
  page.value = 1
  reloadCombos()
}

function pieceCount(name) {
  return selFilter.value.filter((n) => n === name).length
}

// ---- 配置管理 ----

function openConfigManager() {
  view.value = 'form'
}

function openNew() {
  cfgFormMode.value = 'new'
  editingCfgId.value = null
  cfgEditData.value = { name: '', grid: '', cartridgeId: 0 }
  cfgDialogVisible.value = true
  view.value = 'form'
}

function editCurrentConfig() {
  const c = configs.value.find((x) => x.id === selectedId.value)
  if (!c) return
  cfgFormMode.value = 'edit'
  editingCfgId.value = c.id
  cfgEditData.value = { name: c.name, grid: c.grid, cartridgeId: c.cartridgeId || 0 }
  cfgDialogVisible.value = true
}

function editConfigFromList(c) {
  cfgFormMode.value = 'edit'
  editingCfgId.value = c.id
  cfgEditData.value = { name: c.name, grid: c.grid, cartridgeId: c.cartridgeId || 0 }
  cfgDialogVisible.value = true
}

async function handleDialogSave(formData) {
  const name = formData.name.trim()
  if (!name) {
    ElMessage.warning('请填写配置名称')
    return
  }
  if (!formData.grid.split('\n').some((l) => l.includes('1'))) {
    ElMessage.warning('请先填入总模型的形状')
    return
  }
  saving.value = true
  try {
    const cartridgeId = formData.cartridgeId || 0
    if (cfgFormMode.value === 'edit') {
      const res = await updateConfig(editingCfgId.value, name, formData.grid, cartridgeId)
      await loadConfigs()
      if (res.dupOf) ElMessage.info(`该形状与「${res.dupOf}」相同，将共享计算结果`)
      else ElMessage.success('已保存')
      if (view.value === 'detail' && selectedId.value === editingCfgId.value) {
        await showDetail()
      }
    } else {
      const res = await createConfig(name, formData.grid, cartridgeId)
      await loadConfigs()
      if (res.dupOf) ElMessage.info(`该形状与「${res.dupOf}」相同，将共享计算结果`)
      else ElMessage.success('已创建')
    }
    cfgDialogVisible.value = false
  } catch (e) {
    ElMessage.error(errMsg(e))
  } finally {
    saving.value = false
  }
}

function handleDialogCancel() {
  cfgDialogVisible.value = false
}

function handleDialogReset() {
  // 对话框内部已处理重置逻辑
}

async function remove(id) {
  try {
    await ElMessageBox.confirm('确定删除这个配置？', '提示', { type: 'warning' })
  } catch {
    return
  }
  try {
    await deleteConfig(id)
    if (selectedId.value === id) {
      selectedId.value = null
      detail.value = null
      view.value = 'empty'
    }
    await loadConfigs()
    ElMessage.success('已删除')
  } catch (e) {
    ElMessage.error(errMsg(e))
  }
}

// ---- 卡带管理 ----

function openCartridge() {
  resetCartForm()
  view.value = 'cartridge'
  loadCartridges()
}

function resetCartForm() {
  cartFormMode.value = 'new'
  cartForm.value = { name: '', pieces: [] }
}

function editCartridgeFromList(c) {
  cartFormMode.value = 'edit'
  editingCartId.value = c.id
  cartForm.value = { name: c.name, pieces: parseCartridgePieces(c.pieces) }
}

function parseCartridgePieces(piecesStr) {
  try {
    return JSON.parse(piecesStr)
  } catch {
    return []
  }
}

function toggleCartPiece(name) {
  const idx = cartForm.value.pieces.indexOf(name)
  if (idx === -1) cartForm.value.pieces.push(name)
  else cartForm.value.pieces.splice(idx, 1)
}

async function saveCartridge() {
  const name = cartForm.value.name.trim()
  if (!name) {
    ElMessage.warning('请填写卡带名称')
    return
  }
  saving.value = true
  try {
    const piecesStr = JSON.stringify(cartForm.value.pieces)
    if (cartFormMode.value === 'edit' && editingCartId.value !== null) {
      await updateCartridge(editingCartId.value, name, piecesStr)
      ElMessage.success('已保存')
      resetCartForm()
      await loadCartridges()
    } else {
      await createCartridge(name, piecesStr)
      ElMessage.success('已添加卡带')
      resetCartForm()
      await loadCartridges()
    }
  } catch (e) {
    ElMessage.error(errMsg(e))
  } finally {
    saving.value = false
  }
}

async function loadCartridges() {
  try {
    cartridges.value = await listCartridges()
  } catch (e) {
    ElMessage.error(errMsg(e))
  }
}

async function removeCartridge(id) {
  try {
    await ElMessageBox.confirm('确定删除这个卡带？', '提示', { type: 'warning' })
  } catch {
    return
  }
  try {
    await deleteCartridge(id)
    if (cartridgeFilter.value === id) cartridgeFilter.value = null
    await loadCartridges()
    ElMessage.success('已删除')
  } catch (e) {
    ElMessage.error(errMsg(e))
  }
}

function onCartridgeFilterChange(id) {
  applyCartridgeFilter(id)
  page.value = 1
  reloadCombos()
}

function applyCartridgeFilter(id) {
  cartridgeFilter.value = id || null
  if (!id) {
    selFilter.value = []
    return
  }
  const c = cartridges.value.find((x) => x.id === id)
  if (!c) return
  const pieces = parseCartridgePieces(c.pieces)
  if (!pieces.length) {
    selFilter.value = []
    return
  }
  selFilter.value = [...pieces]
}

function cartridgeName(id) {
  const c = cartridges.value.find((x) => x.id === id)
  return c ? c.name : ''
}

// ---- 拖拽排序 ----

function onDragStart(e, index, type) {
  dragType.value = type
  dragIndex.value = index
  dragOverIndex.value = -1
  e.dataTransfer.effectAllowed = 'move'
}

function onDragOver(e, index) {
  dragOverIndex.value = index
  e.dataTransfer.dropEffect = 'move'
}

function onDrop(e, index, type) {
  if (dragType.value !== type) return
  const from = dragIndex.value
  const to = index
  if (from === to) {
    onDragEnd()
    return
  }
  const list = type === 'config' ? configs : cartridges
  const arr = [...list.value]
  const [item] = arr.splice(from, 1)
  arr.splice(to, 0, item)
  list.value = arr

  // 持久化排序
  const ids = arr.map((x) => x.id)
  const fn = type === 'config' ? reorderConfigs : reorderCartridges
  fn(ids).catch((e) => {
    ElMessage.error(errMsg(e))
    loadConfigs()
    loadCartridges()
  })
  onDragEnd()
}

function onDragEnd() {
  dragType.value = null
  dragIndex.value = -1
  dragOverIndex.value = -1
}

async function loadPieces() {
  try {
    pieces.value = await listPieces()
    selFilter.value = []
  } catch (e) {
    ElMessage.error(errMsg(e))
  }
}

onMounted(() => {
  loadConfigs()
  loadPieces()
  loadCartridges()
})
</script>