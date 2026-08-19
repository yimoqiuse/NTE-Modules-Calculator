export const PALETTE = {
  A: '#e91e63',
  B: '#9c27b0',
  C: '#673ab7',
  D: '#3f51b5',
  E: '#2196f3',
  F: '#00bcd4',
  G: '#009688',
  H: '#4caf50',
  L: '#cddc39',
  M: '#ff9800',
  N: '#ff5722',
  O: '#795548',
}

export const DEFAULT_EXAMPLE = [
  '0000000',
  '0111110',
  '0111110',
  '0111110',
  '0111110',
  '0000000',
  '0000000',
].join('\n')

export function errMsg(e) {
  if (!e) return '未知错误'
  if (typeof e === 'string') return e
  if (e.message) return e.message
  return JSON.stringify(e)
}