<template>
  <table class="g-grid" :style="{ '--cell': size + 'px' }">
    <tr v-for="(row, r) in rows" :key="r">
      <td
        v-for="(ch, c) in row"
        :key="c"
        :class="cellClass(ch)"
        :style="cellStyle(ch)"
      >{{ letterOf(ch) }}</td>
    </tr>
  </table>
</template>

<script setup>
import { computed } from 'vue'
import { PALETTE } from '../palette'

const props = defineProps({
  text: { type: String, required: true },
  size: { type: Number, default: 26 },
})

const rows = computed(() => String(props.text).split('\n').map((l) => l.split('')))

function cellClass(ch) {
  if (ch === '0' || ch === ' ') return 'cell-0'
  if (ch === '1') return 'cell-1'
  return 'cell-letter'
}

function letterOf(ch) {
  return ch === '0' || ch === '1' || ch === ' ' ? '' : ch
}

function cellStyle(ch) {
  if (ch === '0' || ch === '1' || ch === ' ') return {}
  return { background: PALETTE[ch] || '#8a8a8a' }
}
</script>