package download

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const userAgent = "androidshow/0.1 (+https://github.com/ywshow/androidshow)"

func Download(url, dest, checksum string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	tmp := dest + ".part"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("下载失败: %s\n%w", url, err)
	}
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{
		Transport: &http.Transport{
			TLSHandshakeTimeout:   30 * time.Second,
			ResponseHeaderTimeout: 60 * time.Second,
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("下载失败: %s\n%w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败: %s\nHTTP %s", url, resp.Status)
	}

	out, err := os.Create(tmp)
	if err != nil {
		return err
	}
	hasher := sha256.New()
	reader := io.TeeReader(resp.Body, hasher)
	total := resp.ContentLength
	var read int64
	buf := make([]byte, 64*1024)
	for {
		n, readErr := reader.Read(buf)
		if n > 0 {
			if _, err := out.Write(buf[:n]); err != nil {
				out.Close()
				os.Remove(tmp)
				return err
			}
			read += int64(n)
			printProgress(filepath.Base(dest), read, total)
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			out.Close()
			os.Remove(tmp)
			return fmt.Errorf("下载失败: %s\n%w", url, readErr)
		}
	}
	fmt.Fprintln(os.Stderr)
	if err := out.Close(); err != nil {
		os.Remove(tmp)
		return err
	}

	if checksum != "" {
		got := hex.EncodeToString(hasher.Sum(nil))
		if !strings.EqualFold(got, checksum) {
			os.Remove(tmp)
			return fmt.Errorf("%s 校验失败：SHA-256 与目录中记录不一致。", filepath.Base(dest))
		}
	}
	return os.Rename(tmp, dest)
}

func ExtractZip(archive, dest string) error {
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}
	r, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer r.Close()

	destAbs, err := filepath.Abs(dest)
	if err != nil {
		return err
	}
	for _, f := range r.File {
		target := filepath.Join(dest, f.Name)
		targetAbs, err := filepath.Abs(target)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(destAbs, targetAbs)
		if err != nil || strings.HasPrefix(rel, "..") {
			return fmt.Errorf("压缩包包含不安全路径: %s", f.Name)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}
		_, copyErr := io.Copy(out, rc)
		out.Close()
		rc.Close()
		if copyErr != nil {
			return copyErr
		}
	}
	return nil
}

func printProgress(name string, read, total int64) {
	if total > 0 {
		percent := read * 100 / total
		fmt.Fprintf(os.Stderr, "\r下载 %s: %d%% (%s/%s)", name, percent, formatSize(read), formatSize(total))
		return
	}
	fmt.Fprintf(os.Stderr, "\r下载 %s: %s", name, formatSize(read))
}

func formatSize(n int64) string {
	value := float64(n)
	units := []string{"B", "KB", "MB", "GB"}
	for i, unit := range units {
		if value < 1024 || i == len(units)-1 {
			if unit == "B" {
				return fmt.Sprintf("%d %s", n, unit)
			}
			return fmt.Sprintf("%.1f %s", value, unit)
		}
		value /= 1024
	}
	return fmt.Sprintf("%d B", n)
}
