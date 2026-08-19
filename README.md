# 异环驱动块计算

一个用于《异环》驱动块摆放的桌面计算工具：给定总模型的形状，穷举所有可行的「驱动块方块组合」，并为每种组合给出一种完整的摆放方案。支持按方块筛选、关联卡带、结果缓存与多核并行求解。

基于 [Wails v2](https://wails.io)（Go + Vue 3 + Element Plus）构建。

## 功能特性

- **形状编辑**：点击格子或直接粘贴 `0/1` 文本绘制总模型形状（行、列最大 20），编辑器内居中显示。
- **历史形状快速切换**：新增/编辑配置时，可用左右箭头按钮（或键盘 `←`/`→`）直接切换已有的历史形状，免去重复绘制。
- **配置搜索**：侧边栏按名称搜索配置，支持模糊匹配与拼音/首字母匹配（如「残虹星盘」可搜 `chxp` 或 `canhong`）。
- **穷举求解**：自动枚举所有可行的方块组合，并为每种组合提供一种完整摆法（多核并行）。
- **方块筛选**：点击方块图标即可筛选包含该方块的组合，可多次点击表示「至少用到 N 个」，无需重新计算。
- **卡带管理**：把一组「预选方块」保存为卡带，配置关联卡带后可一键筛选。
- **结果缓存**：求解结果缓存在本地 SQLite 中，同一形状重复查看秒开；求解逻辑升级时自动清空旧缓存。
- **多配置管理**：同一形状的多个配置共享计算结果，支持拖拽排序、重命名、删除。

## 环境要求

- [Go](https://go.dev/dl/) 1.25+
- [Node.js](https://nodejs.org/)（含 npm）
- [Wails CLI](https://wails.io/docs/gettingstarted/installation) v2.13+，安装方式：

  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  ```

  Windows 下还需安装 [WebView2](https://developer.microsoft.com/microsoft-edge/webview2/) 运行时（Win10/11 一般已内置）。

## 使用说明

### 快速上手

1. 运行程序（或直接使用 `build/bin/异环驱动块计算.exe`）。
2. 配置多了以后，可在左侧**搜索框**按名称快速定位（支持模糊 / 拼音首字母 / 全拼，如输入 `ch`、`canhong`）。
3. 点击左侧「配置管理」→「新增配置」。
4. 在弹窗中：
   - 填写**配置名称**；
   - 用格子编辑器绘制总模型的形状（也可粘贴 `0/1` 文本，`1` 表示可摆放区域）；
   - 若形状与已有配置重复，可用旁边的 **`←`/`→` 按钮或键盘方向键**切换到历史形状直接使用，无需重绘；
   - 可选：关联一个卡带，用于自动筛选；
   - 点击**保存**。
5. 保存后点击左侧该配置进入详情页，程序会一次性算出该形状的全部可行方块组合。
6. 在详情页中：
   - 点击方块库里的方块图标，筛选包含该方块的组合（右键可减少 1 个，数字角标表示当前要求的最少数量）；
   - 用下方分页浏览每种组合的完整摆法，字母即方块名；
   - 点击「清除筛选」恢复全部结果。

### 方块说明

12 种驱动块方块（字母为方块名）：

| 字母 | 形状 | 大小 |
| ---- | ---- | ---- |
| A | `AA` | 2 |
| B | `B/B` | 2 |
| C | `CCC` | 3 |
| D | `D/D/D` | 3 |
| E | `E/EE` | 3 |
| F | `FF/F` | 3 |
| G | `GG/·G` | 3 |
| H | `·H/HH` | 3 |
| L | `LLLL` | 4 |
| M | `M/M/M/M` | 4 |
| N | `·NN/NN` | 4 |
| O | `·O/OO/O` | 4 |

> 方块总数（12 种 × 原样摆放）必须恰好填满整个形状区域才会被列为可行组合。每种方块在一种组合中最多使用一次。

### 数据存储

- 程序数据存放在 **exe 所在目录**的 `data.db`（SQLite），随程序一起移动即可备份/迁移。
- 若程序目录不可写（如安装在 `C:\Program Files`），则自动回退到用户配置目录 `%AppData%\异环驱动块\`。
- 删配置会级联清理其形状与缓存结果。

## 开发（Live Development）

```bash
wails dev
```

- 后端改动会触发自动重编译重启；
- 前端由 Vite 提供热更新；
- 浏览器开发模式：Wails 会启动一个 `http://localhost:34115` 的前端服务，可在浏览器中调用 Go 方法（适合调试样式）。

## 构建文档

### 生产构建

```bash
wails build
```

- 产物输出到 `build/bin/异环驱动块计算.exe`（名称由 `wails.json` 的 `outputfilename` 决定）。
- 前端会被 `npm run build` 打包后以 `//go:embed` 嵌入进可执行文件，**产物为单文件，可直接分发**。
- 图标与 Windows 版本信息来自 `build/windows/`（见下文）。

常用参数：

```bash
wails build -clean           # 构建前清理缓存
wails build -upx             # 使用 UPX 压缩体积（需安装 upx）
wails build -platform windows/amd64   # 指定平台
wails build -trimpath        # 去除构建路径信息
wails build -nsis            # 同时生成 Windows NSIS 安装包
```

### 仅构建前端 / 后端

```bash
# 前端产物（输出到 frontend/dist，供 go:embed 使用）
cd frontend && npm install && npm run build

# 后端编译（需要先有 frontend/dist）
go build -o build/bin/异环驱动块计算.exe .
```

### Windows 打包文件（build/windows）

- `icon.ico`：程序图标。替换后重新 `wails build` 生效；若缺失会由 `build/appicon.png` 自动生成。
- `info.json`：可执行文件的版本信息（右键 exe → 属性 → 详细信息）。
- `installer/`：NSIS 安装包脚本（使用 `wails build -nsis` 时生效）。
- `wails.exe.manifest`：应用程序清单。

> `build/bin` 与 `frontend/dist` 已被 `.gitignore` 忽略，不会进入版本库。

## 项目结构

```
.
├── main.go            # 入口：数据目录探测、Wails 启动
├── app.go             # 暴露给前端的 Go 方法（Wails Bind）
├── wails.json         # Wails 项目配置（产物名、前端构建命令等）
├── solver/            # 求解核心：方块定义、组合枚举、平铺搜索（多核并行）
├── store/             # SQLite 存储层：形状/配置/卡带/结果缓存与迁移
├── build/             # 构建资源（图标、Windows 安装包、产物目录）
└── frontend/          # Vue 3 + Element Plus 前端
    ├── src/api.js     # 调用 Go 后端方法的封装
    ├── src/App.vue    # 主界面：配置/卡带管理、详情与筛选
    └── src/components/  # GridEditor（形状编辑）、GridDisplay（摆法展示）、ConfigFormDialog
```

## 求解机制（简要）

1. 将形状解析为网格区域，统计格子总数。
2. 枚举所有「方块用量组合」，使各方块大小之和恰等于格子总数（每方块 0~1 次）。
3. 为每个组合做**精确覆盖搜索**，找到一套摆法即停；搜不出即判定不可行。
4. 结果按方块组合排序，缓存入库。计算量极大的组合可能因节点预算超限未搜完，界面会给出警告。

## 技术栈

- 后端：Go 1.25+、Wails v2、modernc.org/sqlite（纯 Go，免 CGO）
- 前端：Vue 3、Element Plus、Vite、pinyin-pro（配置搜索的拼音匹配）