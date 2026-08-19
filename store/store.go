package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"qudongkuai-gui/solver"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

type Config struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	ShapeID     int64  `json:"shapeId"`
	CartridgeID int64  `json:"cartridgeId"`
	Hash        string `json:"hash"`
	Grid        string `json:"grid"`
	SortOrder   int64  `json:"sortOrder"`
}

type Cartridge struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Pieces    string `json:"pieces"` // JSON array of piece names, e.g. ["A","B","C"]
	SortOrder int64  `json:"sortOrder"`
}

type ComboResult struct {
	Combo string `json:"combo"`
	Grid  string `json:"grid"`
}

type Result struct {
	Total    int           `json:"total"`
	Combos   []ComboResult `json:"combos"`
	Warnings []string      `json:"warnings"`
}

func Open(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", filepath.Join(dir, "data.db"))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	// SQLite 默认关闭外键，必须显式开启，否则 ON DELETE CASCADE 不生效
	if _, err := db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		db.Close()
		return nil, err
	}
	s := &Store{db: db}
	if err := s.init(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }

const solverVersion = "6"

func (s *Store) init() error {
	if _, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS shapes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    hash TEXT NOT NULL UNIQUE,
    grid TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS configs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    shape_id INTEGER NOT NULL REFERENCES shapes(id) ON DELETE CASCADE,
    cartridge_id INTEGER NOT NULL DEFAULT 0,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE TABLE IF NOT EXISTS results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    shape_id INTEGER NOT NULL REFERENCES shapes(id) ON DELETE CASCADE,
    params TEXT NOT NULL,
    combos TEXT NOT NULL,
    total INTEGER NOT NULL,
    computed_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(shape_id, params)
);
CREATE TABLE IF NOT EXISTS cartridges (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    pieces TEXT NOT NULL DEFAULT '[]',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE TABLE IF NOT EXISTS meta (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
`); err != nil {
		return err
	}
	if err := s.migrateSchema(); err != nil {
		return err
	}
	return s.checkVersion()
}

// migrateSchema 为旧数据库补充新列
func (s *Store) migrateSchema() error {
	_, _ = s.db.Exec(`ALTER TABLE configs ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0`)
	_, _ = s.db.Exec(`ALTER TABLE cartridges ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0`)
	_, _ = s.db.Exec(`ALTER TABLE configs ADD COLUMN cartridge_id INTEGER NOT NULL DEFAULT 0`)
	return nil
}

// checkVersion 求解逻辑变更时清空旧的结果缓存
func (s *Store) checkVersion() error {
	var v string
	err := s.db.QueryRow(`SELECT value FROM meta WHERE key = 'solver_version'`).Scan(&v)
	if err == sql.ErrNoRows {
		_, err = s.db.Exec(`INSERT INTO meta (key, value) VALUES ('solver_version', ?)`, solverVersion)
		return err
	}
	if err != nil {
		return err
	}
	if v != solverVersion {
		if _, err := s.db.Exec(`DELETE FROM results`); err != nil {
			return err
		}
		_, err = s.db.Exec(`UPDATE meta SET value = ? WHERE key = 'solver_version'`, solverVersion)
		return err
	}
	return nil
}

func (s *Store) getOrCreateShape(hash, grid string) (int64, error) {
	var id int64
	err := s.db.QueryRow(`SELECT id FROM shapes WHERE hash = ?`, hash).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}
	res, err := s.db.Exec(`INSERT INTO shapes (hash, grid) VALUES (?, ?)`, hash, grid)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) CreateConfig(name, grid string, cartridgeID int64) (Config, error) {
	norm := solver.NormalizeGrid(grid)
	if norm == "" {
		return Config{}, fmt.Errorf("形状不能为空")
	}
	hash := solver.GridHash(norm)
	shapeID, err := s.getOrCreateShape(hash, norm)
	if err != nil {
		return Config{}, err
	}
	// 新配置排在最后
	var maxOrder int64
	_ = s.db.QueryRow(`SELECT COALESCE(MAX(sort_order), 0) FROM configs`).Scan(&maxOrder)
	res, err := s.db.Exec(`INSERT INTO configs (name, shape_id, cartridge_id, sort_order) VALUES (?, ?, ?, ?)`, name, shapeID, cartridgeID, maxOrder+1)
	if err != nil {
		return Config{}, err
	}
	id, _ := res.LastInsertId()
	return Config{ID: id, Name: name, ShapeID: shapeID, CartridgeID: cartridgeID, Hash: hash, Grid: norm, SortOrder: maxOrder + 1}, nil
}

func (s *Store) UpdateConfig(id int64, name, grid string, cartridgeID int64) (Config, error) {
	norm := solver.NormalizeGrid(grid)
	if norm == "" {
		return Config{}, fmt.Errorf("形状不能为空")
	}
	hash := solver.GridHash(norm)
	shapeID, err := s.getOrCreateShape(hash, norm)
	if err != nil {
		return Config{}, err
	}
	if _, err := s.db.Exec(`UPDATE configs SET name = ?, shape_id = ?, cartridge_id = ? WHERE id = ?`, name, shapeID, cartridgeID, id); err != nil {
		return Config{}, err
	}
	return Config{ID: id, Name: name, ShapeID: shapeID, CartridgeID: cartridgeID, Hash: hash, Grid: norm}, nil
}

func (s *Store) scanConfig(row *sql.Row) (Config, error) {
	var c Config
	err := row.Scan(&c.ID, &c.Name, &c.ShapeID, &c.CartridgeID, &c.Hash, &c.Grid, &c.SortOrder)
	return c, err
}

func (s *Store) ListConfigs() ([]Config, error) {
	rows, err := s.db.Query(`
		SELECT c.id, c.name, c.shape_id, c.cartridge_id, sh.hash, sh.grid, c.sort_order
		FROM configs c JOIN shapes sh ON sh.id = c.shape_id
		ORDER BY c.sort_order, c.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Config
	for rows.Next() {
		var c Config
		if err := rows.Scan(&c.ID, &c.Name, &c.ShapeID, &c.CartridgeID, &c.Hash, &c.Grid, &c.SortOrder); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) GetConfig(id int64) (Config, error) {
	return s.scanConfig(s.db.QueryRow(`
		SELECT c.id, c.name, c.shape_id, c.cartridge_id, sh.hash, sh.grid, c.sort_order
		FROM configs c JOIN shapes sh ON sh.id = c.shape_id
		WHERE c.id = ?`, id))
}

func (s *Store) QueryNameSharingShape(shapeID, excludeID int64, name *string) error {
	return s.db.QueryRow(
		`SELECT name FROM configs WHERE shape_id = ? AND id != ? LIMIT 1`,
		shapeID, excludeID).Scan(name)
}

func (s *Store) RenameConfig(id int64, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("配置名称不能为空")
	}
	_, err := s.db.Exec(`UPDATE configs SET name = ? WHERE id = ?`, name, id)
	return err
}

func (s *Store) DeleteConfig(id int64) error {
	if _, err := s.db.Exec(`DELETE FROM configs WHERE id = ?`, id); err != nil {
		return err
	}
	_, err := s.db.Exec(`DELETE FROM shapes WHERE id NOT IN (SELECT shape_id FROM configs)`)
	return err
}

func (s *Store) GetResult(shapeID int64, params string) (*Result, error) {
	var raw string
	err := s.db.QueryRow(`SELECT combos FROM results WHERE shape_id = ? AND params = ?`, shapeID, params).Scan(&raw)
	if err != nil {
		return nil, err
	}
	var r Result
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *Store) ReorderConfigs(ids []int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(`UPDATE configs SET sort_order = ? WHERE id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for i, id := range ids {
		if _, err := stmt.Exec(int64(i), id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) CreateCartridge(name, pieces string) (Cartridge, error) {
	var maxOrder int64
	_ = s.db.QueryRow(`SELECT COALESCE(MAX(sort_order), 0) FROM cartridges`).Scan(&maxOrder)
	res, err := s.db.Exec(`INSERT INTO cartridges (name, pieces, sort_order) VALUES (?, ?, ?)`, name, pieces, maxOrder+1)
	if err != nil {
		return Cartridge{}, err
	}
	id, _ := res.LastInsertId()
	return Cartridge{ID: id, Name: name, Pieces: pieces, SortOrder: maxOrder + 1}, nil
}

func (s *Store) ListCartridges() ([]Cartridge, error) {
	rows, err := s.db.Query(`SELECT id, name, pieces, sort_order FROM cartridges ORDER BY sort_order, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Cartridge
	for rows.Next() {
		var c Cartridge
		if err := rows.Scan(&c.ID, &c.Name, &c.Pieces, &c.SortOrder); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) UpdateCartridge(id int64, name, pieces string) (Cartridge, error) {
	if _, err := s.db.Exec(`UPDATE cartridges SET name = ?, pieces = ? WHERE id = ?`, name, pieces, id); err != nil {
		return Cartridge{}, err
	}
	return Cartridge{ID: id, Name: name, Pieces: pieces}, nil
}

func (s *Store) DeleteCartridge(id int64) error {
	if _, err := s.db.Exec(`UPDATE configs SET cartridge_id = 0 WHERE cartridge_id = ?`, id); err != nil {
		return err
	}
	_, err := s.db.Exec(`DELETE FROM cartridges WHERE id = ?`, id)
	return err
}

func (s *Store) ReorderCartridges(ids []int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(`UPDATE cartridges SET sort_order = ? WHERE id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for i, id := range ids {
		if _, err := stmt.Exec(int64(i), id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) SaveResult(shapeID int64, params string, r Result) error {
	raw, err := json.Marshal(r)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`
		INSERT INTO results (shape_id, params, combos, total) VALUES (?, ?, ?, ?)
		ON CONFLICT(shape_id, params) DO UPDATE SET combos = excluded.combos, total = excluded.total, computed_at = datetime('now')`,
		shapeID, params, string(raw), r.Total)
	return err
}
