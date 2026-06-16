# 打包 macOS 安装包（DMG，功夫豆GEO）

> 在 **macOS** 上从源码打出 `.app` 并制成 **DMG**（含拖拽到 Applications 的快捷方式），放到项目根目录、按日期时分秒命名。
> 制 DMG 用系统自带的 `hdiutil`，**无需额外安装任何工具**。

---

## 一、前置条件（只需配一次）

- **本地 Go 环境 + Wails CLI**：本项目装在 `geo_client2/.go/bin`。
  首次没有就执行：
  ```bash
  cd geo_client2 && make install   # 创建 .go 环境并装 wails、前端依赖
  ```
- `hdiutil`：macOS 系统自带，无需安装。

---

## 二、第 1 步：构建 macOS 应用（.app）

```bash
cd geo_client2
export GOBIN="$PWD/.go/bin"
export PATH="$GOBIN:/opt/homebrew/bin:$PATH"

VERSION=$(grep '"version"' frontend/package.json | head -1 | awk -F'"' '{print $4}')

# universal = 同时兼容 Intel(amd64) 与 Apple Silicon(arm64)；只要本机用可换 darwin/arm64 更快
wails build -platform darwin/universal -ldflags "-X 'geo_client2/backend.Version=$VERSION'"
```

产物：`geo_client2/build/bin/功夫豆GEO.app`（universal，已 self-sign 自签名）。

---

## 三、第 2 步：用 hdiutil 制成 DMG

```bash
ROOT="/Users/zhanghaoye/Desktop/job/2026job/geo_marketing"   # 项目根目录
APP="$ROOT/geo_client2/build/bin/功夫豆GEO.app"
STAGE=/tmp/dmgstage

rm -rf "$STAGE" /tmp/功夫豆GEO.dmg
mkdir -p "$STAGE"
cp -R "$APP" "$STAGE/"                 # 把 .app 放进暂存目录
ln -s /Applications "$STAGE/Applications"   # 加“拖到 Applications 安装”的快捷方式

hdiutil create -volname "功夫豆GEO" -srcfolder "$STAGE" -ov -format UDZO /tmp/功夫豆GEO.dmg
```

- `-format UDZO`：压缩的只读 DMG（体积小，~20MB）
- `-volname`：挂载后显示的卷名

---

## 四、第 3 步：复制到项目根目录，按「日期_时分秒」命名

```bash
TS=$(date +"%Y%m%d_%H%M%S")
cp /tmp/功夫豆GEO.dmg "$ROOT/功夫豆GEO_${TS}.dmg"
rm -rf /tmp/dmgstage /tmp/功夫豆GEO.dmg     # 清理暂存
```

得到例如：`功夫豆GEO_20260612_140614.dmg`

---

## 五、注意事项 / 踩过的坑

- **未做 Apple 公证（notarize）**：当前是 self-sign 自签名，DMG 在**别的 Mac** 上打开会被 Gatekeeper 拦“无法验证开发者”。让用户：
  - 右键 `.app` →「打开」→ 确认；或
  - 终端执行 `xattr -cr /Applications/功夫豆GEO.app` 清除隔离属性后再开。
  - 要彻底免提示需 Apple 开发者证书做签名 + 公证（另配）。
- **universal vs arm64**：`darwin/universal` 兼容性最好但编译慢；只在本机用可改 `darwin/arm64`。
- **不要在 `wails dev` 运行时执行 `wails build`**：都写 `build/bin` 会冲突。先停 `make dev`。
- **macOS TCC 权限**：项目在 `~/Desktop` 下，复制 `.app`/`.dmg` 偶尔报 `Operation not permitted`，是系统隐私权限（非构建问题），等权限恢复重试即可。

---

## 六、一句话流程（已配好环境后）

```bash
cd geo_client2 && export PATH="$PWD/.go/bin:/opt/homebrew/bin:$PATH" && \
wails build -platform darwin/universal && \
ROOT=$(cd .. && pwd) && rm -rf /tmp/dmgstage && mkdir /tmp/dmgstage && \
cp -R build/bin/功夫豆GEO.app /tmp/dmgstage/ && ln -s /Applications /tmp/dmgstage/Applications && \
hdiutil create -volname 功夫豆GEO -srcfolder /tmp/dmgstage -ov -format UDZO /tmp/功夫豆GEO.dmg && \
cp /tmp/功夫豆GEO.dmg "$ROOT/功夫豆GEO_$(date +%Y%m%d_%H%M%S).dmg"
```
