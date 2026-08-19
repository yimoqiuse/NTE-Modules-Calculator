const API = window.go.main.App

export function listConfigs() {
  return API.ListConfigs()
}

export function listPieces() {
  return API.ListPieces()
}

export function getConfigDetail(id) {
  return API.GetConfigDetail(id)
}

export function getConfigCombos(id, filter, page, pageSize) {
  return API.GetConfigCombos(id, filter, page, pageSize)
}

export function createConfig(name, grid, cartridgeId) {
  return API.CreateConfig(name, grid, cartridgeId)
}

export function updateConfig(id, name, grid, cartridgeId) {
  return API.UpdateConfig(id, name, grid, cartridgeId)
}

export function deleteConfig(id) {
  return API.DeleteConfig(id)
}

export function listCartridges() {
  return API.ListCartridges()
}

export function createCartridge(name, pieces) {
  return API.CreateCartridge(name, pieces)
}

export function updateCartridge(id, name, pieces) {
  return API.UpdateCartridge(id, name, pieces)
}

export function deleteCartridge(id) {
  return API.DeleteCartridge(id)
}

export function reorderConfigs(ids) {
  return API.ReorderConfigs(ids)
}

export function reorderCartridges(ids) {
  return API.ReorderCartridges(ids)
}