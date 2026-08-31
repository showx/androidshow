package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"androidshow/internal/catalog"
	"androidshow/internal/java"
	"androidshow/internal/manager"
)

func CommandFor(spec catalog.ToolSpec) ([]string, error) {
	if spec.Kind == "jar" {
		if !manager.Installed(spec) {
			return nil, fmt.Errorf("未安装 %s，请先执行: ashow install %s", spec.Name, spec.Name)
		}
		info, err := java.Find(spec.RequiresJava)
		if err != nil {
			return nil, err
		}
		return []string{info.Path, "-jar", manager.ArtifactPath(spec)}, nil
	}

	binary, err := manager.ResolveBin(spec)
	if err != nil {
		return nil, err
	}
	if filepath.Ext(binary) != ".exe" && filepath.Ext(binary) != ".bat" {
		_ = os.Chmod(binary, 0o755)
	}
	if filepath.Ext(binary) == ".bat" {
		return []string{"cmd", "/c", binary}, nil
	}
	return []string{binary}, nil
}

func Run(spec catalog.ToolSpec, args []string) int {
	command, err := CommandFor(spec)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	cmd := exec.Command(command[0], append(command[1:], args...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			return exit.ExitCode()
		}
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func Output(spec catalog.ToolSpec, args []string) ([]byte, error) {
	command, err := CommandFor(spec)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(command[0], append(command[1:], args...)...)
	return cmd.CombinedOutput()
}
