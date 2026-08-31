package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"androidshow/internal/apkinfo"
	"androidshow/internal/catalog"
	"androidshow/internal/java"
	"androidshow/internal/manager"
	"androidshow/internal/paths"
	"androidshow/internal/runner"
)

var version = "0.1.0"

const help = `安卓常用工具箱（Windows / Linux / macOS）

用法:
  ashow doctor                 检查 Java 与已安装工具
  ashow list                   列出可安装工具
  ashow install [name|--all]   下载并安装工具
  ashow apktool [参数...]      调用 apktool
  ashow jadx [参数...]         调用 jadx
  ashow bundletool [参数...]   调用 bundletool
  ashow aapt [参数...]         调用 aapt
  ashow info <apk>             查看 APK 包名、版本、启动 Activity
  ashow decode <apk>           等价于 apktool d
  ashow build <dir>            等价于 apktool b
`

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printHelp()
		return 0
	}
	if args[0] == "-V" || args[0] == "--version" {
		fmt.Printf("androidshow %s\n", version)
		return 0
	}

	switch args[0] {
	case "doctor":
		return cmdDoctor()
	case "list":
		return cmdList()
	case "install":
		return cmdInstall(args[1:])
	case "info":
		return cmdInfo(args[1:])
	case "decode":
		return cmdRun("apktool", append([]string{"d"}, args[1:]...))
	case "build":
		return cmdRun("apktool", append([]string{"b"}, args[1:]...))
	default:
		if _, err := catalog.Get(args[0]); err == nil {
			return cmdRun(args[0], args[1:])
		}
		fmt.Fprintf(os.Stderr, "未知命令: %s\n\n%s", args[0], help)
		return 2
	}
}

func printHelp() {
	fmt.Print(`ashow - 跨平台安卓工具箱

`)
	fmt.Print(help)
	fmt.Print("选项:\n  -h, --help      显示帮助\n  -V, --version   显示版本\n")
}

func cmdList() int {
	tools, err := catalog.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Println("可安装工具：")
	for _, spec := range tools {
		printToolRow(spec)
	}
	fmt.Printf("\n数据目录: %s\n", paths.DataDir())
	return 0
}

func cmdDoctor() int {
	exe, _ := os.Executable()
	fmt.Printf("androidshow %s\n", version)
	fmt.Printf("系统: %s (%s)\n", paths.OSName(), paths.ArchName())
	fmt.Printf("二进制: %s (%s/%s)\n", exe, runtime.GOOS, runtime.GOARCH)
	fmt.Printf("数据目录: %s\n", paths.DataDir())

	info, err := java.Find(11)
	if err != nil {
		fmt.Printf("Java: %s\n", err)
	} else {
		fmt.Printf("Java: %d  %s\n", info.Version, info.Path)
		fmt.Printf("      %s\n", info.Detail)
	}

	tools, loadErr := catalog.Load()
	if loadErr != nil {
		fmt.Fprintln(os.Stderr, loadErr)
		return 1
	}
	fmt.Println("\n工具状态：")
	for _, spec := range tools {
		printToolRow(spec)
	}
	return 0
}

func cmdInstall(args []string) int {
	force := false
	all := false
	var names []string
	for _, arg := range args {
		switch arg {
		case "--force":
			force = true
		case "--all":
			all = true
		case "-h", "--help":
			fmt.Print("用法: ashow install <工具名|--all> [--force]\n")
			return 0
		default:
			if strings.HasPrefix(arg, "-") {
				fmt.Fprintf(os.Stderr, "未知参数: %s\n", arg)
				return 2
			}
			names = append(names, arg)
		}
	}
	if all || (len(names) == 1 && names[0] == "all") {
		names = catalog.Names()
	}
	if len(names) == 0 {
		fmt.Fprintln(os.Stderr, "请指定工具名，例如: ashow install apktool\n或安装全部: ashow install --all")
		return 2
	}

	for _, name := range names {
		spec, err := catalog.Get(name)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		if manager.Installed(spec) && !force {
			fmt.Printf("%s %s 已安装，跳过。需要重装请加 --force\n", spec.Name, spec.Version)
			continue
		}
		fmt.Printf("正在安装 %s %s ...\n", spec.Name, spec.Version)
		if _, err := manager.Install(name, force); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		fmt.Printf("%s 安装完成。\n", spec.Name)
	}
	return 0
}

func cmdInfo(args []string) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print("用法: ashow info <apk>\n查看 APK 包名、版本名、启动 Activity 等信息。\n")
		if len(args) == 0 {
			return 2
		}
		return 0
	}
	apk := args[0]
	if _, err := os.Stat(apk); err != nil {
		fmt.Fprintf(os.Stderr, "找不到文件: %s\n", apk)
		return 1
	}
	spec, err := catalog.Get("aapt")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if !manager.Installed(spec) {
		fmt.Fprintln(os.Stderr, "未安装 aapt，请先执行: ashow install aapt")
		return 1
	}
	out, err := runner.Output(spec, []string{"dump", "badging", apk})
	text := string(out)
	if err != nil && !strings.Contains(text, "package:") {
		fmt.Fprintln(os.Stderr, strings.TrimSpace(text))
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	info := apkinfo.ParseBadging(text)
	if info.Package == "" {
		fmt.Fprintln(os.Stderr, "未能从 APK 解析出包名。原始输出：")
		fmt.Fprintln(os.Stderr, text)
		return 1
	}
	fmt.Printf("包名:     %s\n", info.Package)
	if info.Label != "" {
		fmt.Printf("应用名:   %s\n", info.Label)
	}
	if info.VersionName != "" || info.VersionCode != "" {
		fmt.Printf("版本:     %s (%s)\n", emptyDash(info.VersionName), emptyDash(info.VersionCode))
	}
	if info.Launch != "" {
		fmt.Printf("启动:     %s\n", info.Launch)
	}
	if info.MinSDK != "" || info.TargetSDK != "" {
		fmt.Printf("SDK:      min %s / target %s\n", emptyDash(info.MinSDK), emptyDash(info.TargetSDK))
	}
	return 0
}

func emptyDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func cmdRun(name string, args []string) int {
	spec, err := catalog.Get(name)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return runner.Run(spec, args)
}

func printToolRow(spec catalog.ToolSpec) {
	status := "未安装"
	if manager.Installed(spec) {
		status = "已安装"
	}
	fmt.Printf("  %-12s %-8s %-6s  %s\n", spec.Name, spec.Version, status, spec.Description)
}
