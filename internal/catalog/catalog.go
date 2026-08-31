package catalog

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"androidshow/internal/paths"
)

//go:embed catalog.json
var raw []byte

type PlatformArtifact struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	SHA256   string `json:"sha256,omitempty"`
	Bin      string `json:"bin,omitempty"`
}

type ToolSpec struct {
	Name         string                      `json:"name"`
	Description  string                      `json:"description"`
	Version      string                      `json:"version"`
	Kind         string                      `json:"kind"`
	RequiresJava int                         `json:"requires_java"`
	URL          string                      `json:"url"`
	Filename     string                      `json:"filename"`
	SHA256       string                      `json:"sha256,omitempty"`
	BinUnix      string                      `json:"bin_unix,omitempty"`
	BinWindows   string                      `json:"bin_windows,omitempty"`
	Platforms    map[string]PlatformArtifact `json:"platforms,omitempty"`
}

func (s ToolSpec) CurrentPlatform() (PlatformArtifact, error) {
	if len(s.Platforms) == 0 {
		bin := s.BinUnix
		if paths.IsWindows() {
			bin = s.BinWindows
		}
		return PlatformArtifact{
			URL:      s.URL,
			Filename: s.Filename,
			SHA256:   s.SHA256,
			Bin:      bin,
		}, nil
	}
	osName := paths.OSName()
	art, ok := s.Platforms[osName]
	if !ok {
		return PlatformArtifact{}, fmt.Errorf("%s 没有 %s 平台的发行包", s.Name, osName)
	}
	return art, nil
}

func Load() ([]ToolSpec, error) {
	var tools []ToolSpec
	if err := json.Unmarshal(raw, &tools); err != nil {
		return nil, fmt.Errorf("解析工具目录失败: %w", err)
	}
	return tools, nil
}

func Get(name string) (ToolSpec, error) {
	tools, err := Load()
	if err != nil {
		return ToolSpec{}, err
	}
	for _, spec := range tools {
		if spec.Name == name {
			return spec, nil
		}
	}
	names := make([]string, 0, len(tools))
	for _, spec := range tools {
		names = append(names, spec.Name)
	}
	return ToolSpec{}, fmt.Errorf("未知工具: %s。当前支持: %s", name, strings.Join(names, ", "))
}

func Names() []string {
	tools, err := Load()
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(tools))
	for _, spec := range tools {
		names = append(names, spec.Name)
	}
	return names
}
