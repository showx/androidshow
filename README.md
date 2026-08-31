# androidshow

跨平台安卓常用工具箱。用 Go 编译成单一二进制，在 Windows、Linux、macOS 上用同一套命令调用 apktool、jadx、bundletool 等工具，并统一负责下载和版本管理。

底层工具仍使用官方发行包；本仓库提供统一入口，不重新实现它们。

## 环境要求

- 编译：Go 1.21+
- 运行：对应平台的 `ashow` 二进制
- JDK 11+（apktool / jadx / bundletool 都依赖 Java）
  - 可通过 `JAVA_HOME` 或 `ANDROIDSHOW_JAVA` 指定 `java` 可执行文件

## 编译

当前平台：

```bash
go build -o bin/ashow.exe ./cmd/ashow   # Windows
go build -o bin/ashow ./cmd/ashow       # Linux / macOS
```

一次打出三端二进制：

```powershell
# Windows
.\scripts\build.ps1
```

```bash
# Linux / macOS
chmod +x scripts/build.sh
./scripts/build.sh
```

输出：

| 文件 | 平台 |
|------|------|
| `bin/ashow(.exe)` | 当前机器 |
| `dist/ashow-windows-amd64.exe` | Windows x64 |
| `dist/ashow-windows-arm64.exe` | Windows ARM64 |
| `dist/ashow-linux-amd64` | Linux x64 |
| `dist/ashow-linux-arm64` | Linux ARM64 |
| `dist/ashow-darwin-amd64` | macOS Intel |
| `dist/ashow-darwin-arm64` | macOS Apple Silicon |

## 使用

```bash
# Windows
.\bin\ashow.exe doctor
.\bin\ashow.exe install apktool
.\bin\ashow.exe apktool d app.apk

# Linux / macOS
./bin/ashow doctor
./bin/ashow install apktool
./bin/ashow apktool d app.apk
```

## 命令

| 命令 | 说明 |
|------|------|
| `ashow doctor` | 检查系统、二进制、Java、工具安装状态 |
| `ashow list` | 列出可安装工具 |
| `ashow install apktool` | 安装指定工具 |
| `ashow install --all` | 安装目录中的全部工具 |
| `ashow install --force apktool` | 强制重新下载 |
| `ashow apktool ...` | 原样转发给 apktool |
| `ashow jadx ...` | 原样转发给 jadx |
| `ashow bundletool ...` | 原样转发给 bundletool |
| `ashow aapt ...` | 原样转发给 aapt |
| `ashow info app.apk` | 查看 APK 包名、版本、启动 Activity |
| `ashow decode app.apk` | 等价于 `apktool d app.apk` |
| `ashow build out -o app.apk` | 等价于 `apktool b out -o app.apk` |

## 当前封装的工具

| 工具 | 版本 | 用途 |
|------|------|------|
| [apktool](https://apktool.org/) | 3.0.3 | APK 反编译 / 回编译 |
| [jadx](https://github.com/skylot/jadx) | 1.5.6 | Java 反编译 |
| [bundletool](https://github.com/google/bundletool) | 1.18.3 | AAB 拆包、转 APK |
| [aapt](https://developer.android.com/tools/releases/build-tools) | 35.0.0 | 查看 APK 包名 / 资源（按平台下载官方 Build-Tools） |

查看包名：

```bash
ashow install aapt
ashow info app.apk
ashow aapt dump badging app.apk
```

后续计划：`platform-tools`（adb / fastboot）、`apksigner` / `zipalign`。这些同样按 Windows / Linux / macOS 选择官方包。

## 数据目录

下载的工具不会放进仓库，而是写到本机数据目录（可用 `ANDROIDSHOW_HOME` 覆盖）：

| 系统 | 默认路径 |
|------|----------|
| Windows | `%LOCALAPPDATA%\androidshow` |
| macOS | `~/Library/Application Support/androidshow` |
| Linux | `${XDG_DATA_HOME:-~/.local/share}/androidshow` |

## 设计

- 入口是 Go 静态二进制（`CGO_ENABLED=0`），三端交叉编译。
- 工具清单在 `internal/catalog/catalog.json`，编译时嵌入二进制。
- `jar` 类工具（apktool、bundletool）在三端都走 `java -jar`。
- `zip` 类工具（jadx、aapt）解压后按平台选择启动文件。
- 像 aapt 这类官方按系统分发的包，在 `catalog.json` 的 `platforms` 里分别写 Windows / Linux / macOS 下载地址。
