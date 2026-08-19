package main

import (
	"embed"
	"os"
	"path/filepath"

	"qudongkuai-gui/store"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// 数据放在程序运行目录下，随程序一起移动；目录不可写时回退到用户配置目录
	dataDir := "."
	if exe, err := os.Executable(); err == nil {
		dataDir = filepath.Dir(exe)
	}
	if !isWritable(dataDir) {
		if cfg, err := os.UserConfigDir(); err == nil {
			alt := filepath.Join(cfg, "异环驱动块")
			_ = os.MkdirAll(alt, 0o755)
			if isWritable(alt) {
				dataDir = alt
			}
		}
	}
	st, err := store.Open(dataDir)
	if err != nil {
		println("store error:", err.Error())
		return
	}
	defer st.Close()

	app := NewApp(st)

	err = wails.Run(&options.App{
		Title:  "异环驱动块计算",
		Width:  1200,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

// isWritable 探测目录是否可写
func isWritable(dir string) bool {
	f, err := os.CreateTemp(dir, ".wtest*")
	if err != nil {
		return false
	}
	f.Close()
	os.Remove(f.Name())
	return true
}