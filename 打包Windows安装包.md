# 打包 Windows 64 位安装包（功夫豆GEO）

> 在 **macOS** 上从源码跨平台打出 Windows 64 位（amd64）NSIS 安装包，并放到项目根目录、按日期时分秒命名。

---

## 一、前置条件（只需配一次）

1. **Homebrew**（已装）。
2. **NSIS**（Wails 生成 `.exe` 安装包必需）：
   ```bash
   brew install nsis        # 提供 makensis
   makensis -VERSION        # 验证，应输出 v3.x
   ```
3. **本地 Go 环境 + Wails CLI**：本项目装在 `geo_client2/.go/bin`。
   首次没有就执行：
   ```bash
   cd geo_client2 && make install   # 创建 .go 环境并装 wails、前端依赖
   ```

---

## 二、一键打包

```bash
cd geo_client2
export GOBIN="$PWD/.go/bin"
export PATH="$GOBIN:/opt/homebrew/bin:$PATH"     # 让 wails 和 makensis 都在 PATH

VERSION=$(grep '"version"' frontend/package.json | head -1 | awk -F'"' '{print $4}')

wails build -platform windows/amd64 -nsis -ldflags "-X 'geo_client2/backend.Version=$VERSION'"
```

- `-platform windows/amd64`：64 位 Windows
- `-nsis`：生成 NSIS 安装包（**不加只产出免安装 .exe，不产出安装包**）
- `-ldflags`：把版本号编进二进制（可选）

构建约 10–30 秒（首次稍久）。

---

## 三、产物位置

| 文件 | 说明 |
|---|---|
| `geo_client2/build/bin/功夫豆GEO.exe` | 免安装的绿色版可执行文件（~23MB） |
| `geo_client2/build/bin/功夫豆GEO-amd64-installer.exe` | **NSIS 安装包**（~11MB），双击在 Windows 安装 |

> 安装包名规则：`<wails.json 的 info.productName>-<arch>-installer.exe`。
> 改产品名就改 `geo_client2/wails.json` 里的 `productName` / `outputfilename`。

---

## 四、复制到项目根目录，按「日期_时分秒」命名

```bash
ROOT="/Users/zhanghaoye/Desktop/job/2026job/geo_marketing"   # 项目根目录
SRC="$ROOT/geo_client2/build/bin/功夫豆GEO-amd64-installer.exe"
TS=$(date +"%Y%m%d_%H%M%S")
cp "$SRC" "$ROOT/功夫豆GEO_${TS}.exe"
```

得到例如：`功夫豆GEO_20260612_094629.exe`

---

## 五、注意事项 / 踩过的坑

- **不要在 `wails dev` 运行时执行 `wails build`**：两者都写 `build/bin`，会冲突导致 dev 进程退出。先停掉 `make dev` 再打包。
- **跨平台编译**：macOS 上用 Wails 跨编译到 windows/amd64 是纯 Go，无需 mingw / CGO，开箱即用。
- **macOS TCC 权限**：项目在 `~/Desktop` 下，复制文件偶尔报 `Operation not permitted`，是系统隐私权限（非构建问题），等权限恢复重试 `cp` 即可。
- Windows 上运行需要 **WebView2 运行时**（Win10/11 一般自带；老系统安装包会引导安装）。

---

## 六、（可选）封装成 make 目标

`geo_client2/Makefile` 已有 `make build-win`，但**没带 `-nsis`**（只出 .exe，不出安装包）。
如需一键出安装包，可把该目标的 wails 命令改为带 `-nsis`，或新增一个 `build-win-installer` 目标复用本文第二节命令。
