package content

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
)

// LogFFmpegStatus 启动时检测 FFmpeg
func LogFFmpegStatus() {
	p := ffmpegBin()
	if p == "" {
		global.GVA_LOG.Warn("FFmpeg 未检测到：Linux 可 apt install ffmpeg 或运行 scripts/install-office-tools-linux.sh；Windows 用 install-ffmpeg.ps1；或配置 office-tools.ffmpeg-path")
		return
	}
	global.GVA_LOG.Info("FFmpeg 已就绪", zap.String("path", p))
}

func ffmpegBin() string {
	if p := strings.TrimSpace(global.GVA_CONFIG.OfficeTools.FFmpegPath); p != "" {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	for _, name := range []string{"ffmpeg", "ffmpeg.exe"} {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}
	for _, p := range discoverFFmpegPaths() {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	return ""
}

func discoverFFmpegPaths() []string {
	var out []string
	seen := map[string]bool{}
	add := func(p string) {
		if p != "" && !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}

	// 项目内置 tools/ffmpeg（install-ffmpeg.ps1 安装位置）
	for _, root := range projectRoots() {
		add(filepath.Join(root, "tools", "ffmpeg", "bin", "ffmpeg.exe"))
		add(filepath.Join(root, "tools", "ffmpeg", "ffmpeg.exe"))
		matches, _ := filepath.Glob(filepath.Join(root, "tools", "ffmpeg", "*", "bin", "ffmpeg.exe"))
		for _, m := range matches {
			add(m)
		}
	}

	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		add("/usr/bin/ffmpeg")
		add("/usr/local/bin/ffmpeg")
	}
	if runtime.GOOS == "windows" {
		add(`C:\ffmpeg\bin\ffmpeg.exe`)
		add(filepath.Join(os.Getenv("ProgramFiles"), "ffmpeg", "bin", "ffmpeg.exe"))
		add(filepath.Join(os.Getenv("ProgramFiles(x86)"), "ffmpeg", "bin", "ffmpeg.exe"))
		if local := os.Getenv("LOCALAPPDATA"); local != "" {
			matches, _ := filepath.Glob(filepath.Join(local, "Microsoft", "WinGet", "Packages", "*ffmpeg*", "*", "bin", "ffmpeg.exe"))
			for _, m := range matches {
				add(m)
			}
			matches2, _ := filepath.Glob(filepath.Join(local, "Microsoft", "WinGet", "Links", "ffmpeg.exe"))
			for _, m := range matches2 {
				add(m)
			}
		}
		if prog := os.Getenv("ProgramData"); prog != "" {
			matches, _ := filepath.Glob(filepath.Join(prog, "chocolatey", "bin", "ffmpeg.exe"))
			for _, m := range matches {
				add(m)
			}
		}
	}
	return out
}

func projectRoots() []string {
	var roots []string
	seen := map[string]bool{}
	put := func(p string) {
		if p == "" || seen[p] {
			return
		}
		seen[p] = true
		roots = append(roots, p)
	}
	if wd, err := os.Getwd(); err == nil {
		put(wd)
		put(filepath.Join(wd, "server"))
		put(filepath.Dir(wd))
	}
	if exe, err := os.Executable(); err == nil {
		d := filepath.Dir(exe)
		put(d)
		put(filepath.Join(d, ".."))
	}
	return roots
}
