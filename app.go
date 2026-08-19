package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"qudongkuai-gui/solver"
	"qudongkuai-gui/store"
)

type App struct {
	ctx   context.Context
	store *store.Store
}

func NewApp(s *store.Store) *App { return &App{store: s} }

func (a *App) startup(ctx context.Context) { a.ctx = ctx }

type CreateResult struct {
	Config store.Config `json:"config"`
	DupOf  string       `json:"dupOf"`
}

func (a *App) CreateConfig(name, grid string, cartridgeID int64) (CreateResult, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return CreateResult{}, errors.New("配置名称不能为空")
	}
	cfg, err := a.store.CreateConfig(name, grid, cartridgeID)
	if err != nil {
		return CreateResult{}, err
	}
	dupOf := a.findDupName(cfg.ShapeID, cfg.ID)
	return CreateResult{Config: cfg, DupOf: dupOf}, nil
}

func (a *App) findDupName(shapeID, excludeID int64) string {
	var name string
	err := a.store.QueryNameSharingShape(shapeID, excludeID, &name)
	if err == nil {
		return name
	}
	return ""
}

func (a *App) UpdateConfig(id int64, name, grid string, cartridgeID int64) (CreateResult, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return CreateResult{}, errors.New("配置名称不能为空")
	}
	cfg, err := a.store.UpdateConfig(id, name, grid, cartridgeID)
	if err != nil {
		return CreateResult{}, err
	}
	return CreateResult{Config: cfg, DupOf: a.findDupName(cfg.ShapeID, cfg.ID)}, nil
}

func (a *App) ListConfigs() ([]store.Config, error) {
	return a.store.ListConfigs()
}

func (a *App) RenameConfig(id int64, name string) error {
	return a.store.RenameConfig(id, strings.TrimSpace(name))
}

func (a *App) DeleteConfig(id int64) error {
	return a.store.DeleteConfig(id)
}

type ConfigDetail struct {
	store.Config
	Total    int      `json:"total"`
	Cached   bool     `json:"cached"`
	Warnings []string `json:"warnings"`
}

type PieceInfo struct {
	Name  string `json:"name"`
	Shape string `json:"shape"`
}

func (a *App) ListPieces() []PieceInfo {
	out := make([]PieceInfo, 0, len(solver.Pieces))
	for _, p := range solver.Pieces {
		out = append(out, PieceInfo{Name: p.Name, Shape: p.Shape})
	}
	return out
}

func (a *App) CreateCartridge(name, pieces string) (store.Cartridge, error) {
	return a.store.CreateCartridge(name, pieces)
}

func (a *App) UpdateCartridge(id int64, name, pieces string) (store.Cartridge, error) {
	return a.store.UpdateCartridge(id, name, pieces)
}

func (a *App) ListCartridges() ([]store.Cartridge, error) {
	return a.store.ListCartridges()
}

func (a *App) DeleteCartridge(id int64) error {
	return a.store.DeleteCartridge(id)
}

func (a *App) ReorderConfigs(ids []int64) error {
	return a.store.ReorderConfigs(ids)
}

func (a *App) ReorderCartridges(ids []int64) error {
	return a.store.ReorderCartridges(ids)
}

// ensureResult 返回该形状的全量求解结果（缓存优先），需要时现场计算并落库
func (a *App) ensureResult(shapeID int64, grid string) (*store.Result, bool, error) {
	params := "all"
	if cached, err := a.store.GetResult(shapeID, params); err == nil && cached != nil {
		return cached, true, nil
	}
	region, height, width, err := solver.ParseGrid(grid)
	if err != nil {
		return nil, false, err
	}
	res := solver.SolveOnePerSet(region, height, width, nil, true, nil)
	combos := make([]store.ComboResult, len(res.Combos))
	for i, c := range res.Combos {
		combos[i] = store.ComboResult{Combo: c.Combo, Grid: c.Grid}
	}
	warnings := make([]string, 0, len(res.BudgetHits))
	if len(res.BudgetHits) > 0 {
		warnings = append(warnings, fmt.Sprintf("有 %d 种方块组合因计算量过大未能确认结果，总数可能偏小：%s", len(res.BudgetHits), strings.Join(res.BudgetHits, ", ")))
	}
	r := &store.Result{Total: len(combos), Combos: combos, Warnings: warnings}
	_ = a.store.SaveResult(shapeID, params, *r)
	return r, false, nil
}

// GetConfigDetail 计算该形状的全量结果并缓存，仅返回元数据；分页内容由 GetConfigCombos 提供
func (a *App) GetConfigDetail(id int64) (ConfigDetail, error) {
	cfg, err := a.store.GetConfig(id)
	if err != nil {
		return ConfigDetail{}, err
	}
	r, cached, err := a.ensureResult(cfg.ShapeID, cfg.Grid)
	if err != nil {
		return ConfigDetail{}, err
	}
	return ConfigDetail{Config: cfg, Total: r.Total, Cached: cached, Warnings: r.Warnings}, nil
}

type ComboPage struct {
	Total  int                 `json:"total"`
	Combos []store.ComboResult `json:"combos"`
}

// GetConfigCombos 按所需方块筛选后返回一页结果，避免一次全量传输
func (a *App) GetConfigCombos(id int64, filter []string, page, pageSize int) (ComboPage, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	cfg, err := a.store.GetConfig(id)
	if err != nil {
		return ComboPage{}, err
	}
	r, _, err := a.ensureResult(cfg.ShapeID, cfg.Grid)
	if err != nil {
		return ComboPage{}, err
	}
	need := map[string]int{}
	for _, n := range filter {
		if n != "" {
			need[n]++
		}
	}
	filtered := make([]store.ComboResult, 0, len(r.Combos))
	for _, c := range r.Combos {
		if comboSatisfies(c.Combo, need) {
			filtered = append(filtered, c)
		}
	}
	start := (page - 1) * pageSize
	if start > len(filtered) {
		start = len(filtered)
	}
	end := start + pageSize
	if end > len(filtered) {
		end = len(filtered)
	}
	return ComboPage{Total: len(filtered), Combos: filtered[start:end]}, nil
}

func comboSatisfies(combo string, need map[string]int) bool {
	if len(need) == 0 {
		return true
	}
	cnt := map[string]int{}
	for i := 0; i < len(combo); i++ {
		cnt[combo[i:i+1]]++
	}
	for k, v := range need {
		if cnt[k] < v {
			return false
		}
	}
	return true
}
