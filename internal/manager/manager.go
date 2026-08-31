package manager

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"androidshow/internal/catalog"
	"androidshow/internal/download"
	"androidshow/internal/paths"
)

func ArtifactPath(spec catalog.ToolSpec) string {
	return filepath.Join(paths.ToolHome(spec.Name), spec.Filename)
}

func Installed(spec catalog.ToolSpec) bool {
	switch spec.Kind {
	case "jar":
		info, err := os.Stat(ArtifactPath(spec))
		return err == nil && !info.IsDir()
	case "zip":
		_, err := ResolveBin(spec)
		return err == nil
	default:
		return false
	}
}

func ResolveBin(spec catalog.ToolSpec) (string, error) {
	plat, err := spec.CurrentPlatform()
	if err != nil {
		return "", err
	}
	if plat.Bin == "" {
		return "", fmt.Errorf("%s 未配置当前平台的启动文件。", spec.Name)
	}
	found, err := FindBinary(paths.ToolHome(spec.Name), plat.Bin)
	if err != nil {
		return "", fmt.Errorf("未安装 %s，请先执行: ashow install %s", spec.Name, spec.Name)
	}
	return found, nil
}

func FindBinary(home, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("未指定可执行文件")
	}
	if strings.ContainsAny(name, `/\`) {
		p := filepath.Join(home, filepath.FromSlash(name))
		if fileExists(p) {
			return p, nil
		}
		return "", fmt.Errorf("未找到 %s", p)
	}

	var found string
	walkErr := filepath.WalkDir(home, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.EqualFold(d.Name(), name) {
			found = path
			return fs.SkipAll
		}
		return nil
	})
	if walkErr != nil && found == "" {
		return "", walkErr
	}
	if found == "" {
		return "", fmt.Errorf("未找到 %s", name)
	}
	return found, nil
}

func Install(name string, force bool) (catalog.ToolSpec, error) {
	spec, err := catalog.Get(name)
	if err != nil {
		return spec, err
	}
	home := paths.ToolHome(spec.Name)
	if Installed(spec) && !force {
		return spec, nil
	}
	if force {
		if err := os.RemoveAll(home); err != nil {
			return spec, err
		}
	}
	if err := os.MkdirAll(home, 0o755); err != nil {
		return spec, err
	}

	plat, err := spec.CurrentPlatform()
	if err != nil {
		return spec, err
	}
	filename := plat.Filename
	if filename == "" {
		filename = spec.Filename
	}
	cached := filepath.Join(paths.CacheDir(), spec.Name, paths.OSName(), filename)
	if err := download.Download(plat.URL, cached, plat.SHA256); err != nil {
		return spec, err
	}
	switch spec.Kind {
	case "jar":
		target := ArtifactPath(spec)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return spec, err
		}
		return spec, copyFile(cached, target)
	case "zip":
		return spec, download.ExtractZip(cached, home)
	default:
		return spec, fmt.Errorf("不支持的工具类型: %s", spec.Kind)
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}
