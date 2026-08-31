package java

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"androidshow/internal/paths"
)

type Error struct {
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

type Info struct {
	Path    string
	Version int
	Detail  string
}

func exeName() string {
	if paths.IsWindows() {
		return "java.exe"
	}
	return "java"
}

func Candidates() []string {
	seen := map[string]struct{}{}
	var result []string
	add := func(p string) {
		if p == "" {
			return
		}
		key := p
		if paths.IsWindows() {
			key = strings.ToLower(p)
		}
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		result = append(result, p)
	}

	add(os.Getenv("ANDROIDSHOW_JAVA"))
	if home := os.Getenv("JAVA_HOME"); home != "" {
		add(filepath.Join(home, "bin", exeName()))
	}
	add(exeName())
	return result
}

var versionRe = regexp.MustCompile(`version\s+"(\d+)(?:\.(\d+))?`)

func ParseVersion(text string) (int, bool) {
	match := versionRe.FindStringSubmatch(text)
	if match == nil {
		return 0, false
	}
	major, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, false
	}
	if major == 1 && match[2] != "" {
		minor, err := strconv.Atoi(match[2])
		if err != nil {
			return 0, false
		}
		return minor, true
	}
	return major, true
}

func Probe(javaPath string) (*Info, error) {
	cmd := exec.Command(javaPath, "-version")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case <-time.After(15 * time.Second):
		_ = cmd.Process.Kill()
		return nil, fmt.Errorf("java -version 超时")
	case err := <-done:
		if err != nil && stdout.Len() == 0 && stderr.Len() == 0 {
			return nil, err
		}
	}
	output := stderr.String() + stdout.String()
	version, ok := ParseVersion(output)
	if !ok {
		return nil, fmt.Errorf("无法解析 Java 版本")
	}
	detail := ""
	for _, line := range strings.Split(output, "\n") {
		if s := strings.TrimSpace(line); s != "" {
			detail = s
			break
		}
	}
	return &Info{Path: javaPath, Version: version, Detail: detail}, nil
}

func Find(minVersion int) (*Info, error) {
	var last *Info
	for _, candidate := range Candidates() {
		info, err := Probe(candidate)
		if err != nil {
			continue
		}
		last = info
		if info.Version >= minVersion {
			return info, nil
		}
	}
	if last != nil {
		return nil, &Error{Message: fmt.Sprintf(
			"检测到 Java %d（%s），但当前工具需要 Java %d+。请安装更高版本 JDK，或设置 JAVA_HOME / ANDROIDSHOW_JAVA。",
			last.Version, last.Path, minVersion,
		)}
	}
	return nil, &Error{Message: fmt.Sprintf(
		"未检测到 Java。请安装 JDK %d+ 并加入 PATH，或设置 JAVA_HOME / ANDROIDSHOW_JAVA。",
		minVersion,
	)}
}
