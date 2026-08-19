package solver

import (
	"crypto/sha1"
	"fmt"
	"runtime"
	"sort"
	"strings"
	"sync"
)

type Cell struct{ R, C int }

type Piece struct {
	Name  string
	Shape string
}

var Pieces = []Piece{
	{"A", "AA"},
	{"B", "B\nB"},
	{"C", "CCC"},
	{"D", "D\nD\nD"},
	{"E", "E\nEE"},
	{"F", "FF\nF"},
	{"G", "GG\n G"},
	{"H", " H\nHH"},
	{"L", "LLLL"},
	{"M", "M\nM\nM\nM"},
	{"N", " NN\nNN"},
	{"O", " O\nOO\nO"},
}

var nameToText = map[string]string{}

func init() {
	for _, p := range Pieces {
		nameToText[p.Name] = p.Shape
	}
}

func NormalizeGrid(grid string) string {
	lines := []string{}
	for _, line := range strings.Split(strings.ReplaceAll(grid, "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}

func GridHash(grid string) string {
	sum := sha1.Sum([]byte(NormalizeGrid(grid)))
	return fmt.Sprintf("%x", sum[:])
}

func ParseGrid(grid string) (region map[Cell]bool, height, width int, err error) {
	lines := strings.Split(NormalizeGrid(grid), "\n")
	region = map[Cell]bool{}
	width = -1
	for _, line := range lines {
		if width == -1 {
			width = len(line)
		}
		if len(line) != width {
			return nil, 0, 0, fmt.Errorf("每一行长度必须一致")
		}
		for c := 0; c < len(line); c++ {
			ch := line[c]
			if ch != '0' && ch != '1' {
				return nil, 0, 0, fmt.Errorf("只能包含 0 和 1")
			}
			if ch == '1' {
				region[Cell{height, c}] = true
			}
		}
		height++
	}
	if width <= 0 {
		return nil, 0, 0, fmt.Errorf("形状不能为空")
	}
	if len(region) == 0 {
		return nil, 0, 0, fmt.Errorf("形状里没有可填格（1）")
	}
	return region, height, width, nil
}

func PieceSize(name string) int {
	return len(parsePiece(nameToText[name]))
}

func AllPieceNames() []string {
	names := make([]string, 0, len(Pieces))
	for _, p := range Pieces {
		names = append(names, p.Name)
	}
	sort.Strings(names)
	return names
}

func parsePiece(text string) map[Cell]bool {
	rows := strings.Split(text, "\n")
	cells := map[Cell]bool{}
	for r, line := range rows {
		for c := 0; c < len(line); c++ {
			if line[c] != ' ' {
				cells[Cell{r, c}] = true
			}
		}
	}
	return normalizeCells(mapToCells(cells))
}

func mapToCells(m map[Cell]bool) []Cell {
	out := make([]Cell, 0, len(m))
	for c := range m {
		out = append(out, c)
	}
	return out
}

func normalizeCells(cells []Cell) map[Cell]bool {
	minR, minC := int(^uint(0)>>1), int(^uint(0)>>1)
	for _, c := range cells {
		if c.R < minR {
			minR = c.R
		}
		if c.C < minC {
			minC = c.C
		}
	}
	out := map[Cell]bool{}
	for _, c := range cells {
		out[Cell{c.R - minR, c.C - minC}] = true
	}
	return out
}

func sortCells(cells []Cell) {
	sort.Slice(cells, func(i, j int) bool {
		if cells[i].R != cells[j].R {
			return cells[i].R < cells[j].R
		}
		return cells[i].C < cells[j].C
	})
}

func sortedCells(m map[Cell]bool) []Cell {
	out := mapToCells(m)
	sortCells(out)
	return out
}

func placementsForRegion(region map[Cell]bool, name string) [][]Cell {
	base := parsePiece(nameToText[name])
	baseCells := mapToCells(base)
	regionCells := mapToCells(region)
	seen := map[string]bool{}
	var out [][]Cell
	for _, b := range baseCells {
		for _, rc := range regionCells {
			start := Cell{rc.R - b.R, rc.C - b.C}
			cells := make([]Cell, 0, len(baseCells))
			ok := true
			for _, c := range baseCells {
				nc := Cell{start.R + c.R, start.C + c.C}
				if !region[nc] {
					ok = false
					break
				}
				cells = append(cells, nc)
			}
			if !ok {
				continue
			}
			sortCells(cells)
			var sb strings.Builder
			for _, c := range cells {
				fmt.Fprintf(&sb, "%d,%d;", c.R, c.C)
			}
			key := sb.String()
			if !seen[key] {
				seen[key] = true
				out = append(out, cells)
			}
		}
	}
	return out
}

func enumerateCombos(pieces []string, allowRepeat bool, minUsage map[string]int, target int) []map[string]int {
	var combos []map[string]int
	var rec func(idx int, counts map[string]int, total int)
	rec = func(idx int, counts map[string]int, total int) {
		if total > target {
			return
		}
		if idx == len(pieces) {
			if total == target {
				for n, k := range minUsage {
					if counts[n] < k {
						return
					}
				}
				cp := make(map[string]int, len(counts))
				for k, v := range counts {
					cp[k] = v
				}
				combos = append(combos, cp)
			}
			return
		}
		name := pieces[idx]
		size := PieceSize(name)
		maxK := target / size
		if !allowRepeat && maxK > 1 {
			maxK = 1
		}
		for k := 0; k <= maxK; k++ {
			counts[name] = k
			rec(idx+1, counts, total+k*size)
		}
		delete(counts, name)
	}
	rec(0, map[string]int{}, 0)
	return combos
}

type Place struct {
	Name  string
	Cells []Cell
}

const nodeBudget = 2000000

// buildPlacementCache 预计算每个方块在当前区域里的所有摆放位置。
// 摆放位置只依赖区域和方块，与具体组合无关，因此在 SolveOnePerSet 里复用。
func buildPlacementCache(region map[Cell]bool, names []string) map[string][][]Cell {
	cache := make(map[string][][]Cell, len(names))
	for _, name := range names {
		cache[name] = placementsForRegion(region, name)
	}
	return cache
}

// findOneTilingBudget 返回 (摆法, 是否预算超限)。预算超限说明结果不可靠（可能实际可摆但没搜完）。
func findOneTilingBudget(region map[Cell]bool, combo map[string]int, placements map[string][][]Cell, budget int) ([]Place, bool) {
	regionList := sortedCells(region)
	cover := map[Cell][]Place{}
	for name := range combo {
		for _, cells := range placements[name] {
			p := Place{name, cells}
			for _, c := range cells {
				cover[c] = append(cover[c], p)
			}
		}
	}
	used := map[string]int{}
	nodes := 0

	disjoint := func(cells []Cell, covered map[Cell]bool) bool {
		for _, c := range cells {
			if covered[c] {
				return false
			}
		}
		return true
	}

	var dfs func(covered map[Cell]bool, chosen []Place) ([]Place, bool)
	dfs = func(covered map[Cell]bool, chosen []Place) ([]Place, bool) {
		nodes++
		if nodes > budget {
			return nil, true
		}
		if len(covered) == len(regionList) {
			return append([]Place(nil), chosen...), false
		}
		bestCell := Cell{-1, -1}
		bestCnt := -1
		for _, c := range regionList {
			if covered[c] {
				continue
			}
			cnt := 0
			for _, p := range cover[c] {
				if used[p.Name] < combo[p.Name] && disjoint(p.Cells, covered) {
					cnt++
				}
			}
			if cnt == 0 {
				return nil, false
			}
			if bestCnt == -1 || cnt < bestCnt {
				bestCnt = cnt
				bestCell = c
			}
		}
		for _, p := range cover[bestCell] {
			if used[p.Name] >= combo[p.Name] || !disjoint(p.Cells, covered) {
				continue
			}
			used[p.Name]++
			for _, c := range p.Cells {
				covered[c] = true
			}
			res, hit := dfs(covered, append(chosen, p))
			if res != nil {
				return res, false
			}
			if hit {
				return nil, true
			}
			used[p.Name]--
			for _, c := range p.Cells {
				delete(covered, c)
			}
		}
		return nil, false
	}

	return dfs(map[Cell]bool{}, nil)
}

func mapKeys(m map[string]int) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func comboLabel(combo map[string]int) string {
	names := make([]string, 0, len(combo))
	for n := range combo {
		names = append(names, n)
	}
	sort.Strings(names)
	var sb strings.Builder
	for _, n := range names {
		sb.WriteString(strings.Repeat(n, combo[n]))
	}
	return sb.String()
}

// RenderSolution 把一套摆法渲染成文本棋盘
func RenderSolution(sol []Place, height, width int, region map[Cell]bool) string {
	grid := make([][]byte, height)
	for r := range grid {
		grid[r] = []byte(strings.Repeat("0", width))
	}
	for c := range region {
		grid[c.R][c.C] = '1'
	}
	for _, p := range sol {
		for _, c := range p.Cells {
			grid[c.R][c.C] = p.Name[0]
		}
	}
	lines := make([]string, height)
	for r := range grid {
		lines[r] = string(grid[r])
	}
	return strings.Join(lines, "\n")
}

type ComboResult struct {
	Combo string `json:"combo"`
	Grid  string `json:"grid"`
}

// SolveResult 一次求解的完整输出
type SolveResult struct {
	Combos []ComboResult
	// BudgetHits 因节点预算超限未能确认的方块组合标签，结果总数可能偏小
	BudgetHits []string
}

type solveOutcome struct {
	label string
	grid  string
	hit   bool
}

// SolveOnePerSet 每个可行的方块组合返回一套摆法
func SolveOnePerSet(region map[Cell]bool, height, width int, pieces []string, allowRepeat bool, minUsage map[string]int) SolveResult {
	if pieces == nil {
		pieces = AllPieceNames()
	}
	sort.Strings(pieces)
	combos := enumerateCombos(pieces, allowRepeat, minUsage, len(region))
	cache := buildPlacementCache(region, pieces)

	workers := runtime.GOMAXPROCS(0)
	if workers < 1 {
		workers = 1
	}
	jobs := make(chan map[string]int)
	results := make(chan solveOutcome, len(combos))
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for combo := range jobs {
				label := comboLabel(combo)
				sol, hit := findOneTilingBudget(region, combo, cache, nodeBudget)
				if sol == nil {
					if hit {
						results <- solveOutcome{label: label, hit: true}
					}
					continue
				}
				results <- solveOutcome{label: label, grid: RenderSolution(sol, height, width, region)}
			}
		}()
	}
	go func() {
		defer close(jobs)
		for _, combo := range combos {
			jobs <- combo
		}
	}()
	go func() {
		wg.Wait()
		close(results)
	}()

	out := make([]ComboResult, 0, len(combos))
	var hits []string
	for r := range results {
		if r.hit {
			hits = append(hits, r.label)
			continue
		}
		out = append(out, ComboResult{Combo: r.label, Grid: r.grid})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Combo < out[j].Combo })
	sort.Strings(hits)
	return SolveResult{Combos: out, BudgetHits: hits}
}