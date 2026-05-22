package content

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
)

// LogLibreOfficeStatus 启动时输出 LibreOffice 探测结果
func LogLibreOfficeStatus() {
	s := &OfficeFileConvertService{}
	p := s.libreOfficeBin()
	if p == "" {
		global.GVA_LOG.Warn("LibreOffice 未检测到：Linux 可 apt install libreoffice 或 scripts/install-office-tools-linux.sh；Windows 用 install-libreoffice.ps1；或配置 libreoffice-path；亦可上传 docx/xlsx/pptx 走 Go 简易转换")
		return
	}
	global.GVA_LOG.Info("LibreOffice 已就绪，Office 高保真转 PDF 可用", zap.String("path", p))
}

func (s *OfficeFileConvertService) libreOfficeBin() string {
	cfg := global.GVA_CONFIG.OfficeTools
	if p := strings.TrimSpace(cfg.LibreOfficePath); p != "" {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	if !cfg.EnableLibreOffice {
		return ""
	}
	for _, name := range []string{"soffice", "libreoffice", "soffice.exe", "libreoffice.exe"} {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}
	for _, p := range discoverLibreOfficeAll() {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func discoverLibreOfficeAll() []string {
	var out []string
	seen := map[string]bool{}
	add := func(p string) {
		if p == "" || seen[p] {
			return
		}
		seen[p] = true
		out = append(out, p)
	}
	for _, p := range discoverLibreOfficeProjectLocal() {
		add(p)
	}
	for _, p := range discoverLibreOfficeOnWindows() {
		add(p)
	}
	for _, p := range discoverLibreOfficeOnUnix() {
		add(p)
	}
	return out
}

func discoverLibreOfficeOnUnix() []string {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		return nil
	}
	return []string{
		"/usr/bin/soffice",
		"/usr/local/bin/soffice",
		"/usr/bin/libreoffice",
		"/snap/bin/libreoffice",
	}
}

// discoverLibreOfficeProjectLocal 项目内或程序目录旁的便携/本地安装
func discoverLibreOfficeProjectLocal() []string {
	var roots []string
	if wd, err := os.Getwd(); err == nil {
		roots = append(roots, wd)
	}
	if exe, err := os.Executable(); err == nil {
		roots = append(roots, filepath.Dir(exe))
	}
	if env := os.Getenv("GVA_LIBREOFFICE_HOME"); env != "" {
		roots = append(roots, env)
	}
	var candidates []string
	for _, root := range roots {
		candidates = append(candidates,
			filepath.Join(root, "libreoffice", "program", "soffice.exe"),
			filepath.Join(root, "LibreOffice", "program", "soffice.exe"),
			filepath.Join(root, "server", "libreoffice", "program", "soffice.exe"),
			filepath.Join(root, "third_party", "libreoffice", "program", "soffice.exe"),
			filepath.Join(root, "..", "libreoffice", "program", "soffice.exe"),
			filepath.Join(root, "libreoffice", "program", "soffice"),
			filepath.Join(root, "server", "libreoffice", "program", "soffice"),
		)
	}
	return candidates
}

// discoverLibreOfficeOnWindows 常见安装路径（未加入 PATH 时也能找到）
func discoverLibreOfficeOnWindows() []string {
	if runtime.GOOS != "windows" {
		return nil
	}
	var roots []string
	for _, env := range []string{"ProgramFiles", "ProgramFiles(x86)", "LOCALAPPDATA"} {
		if v := os.Getenv(env); v != "" {
			roots = append(roots, v)
		}
	}
	var out []string
	seen := map[string]bool{}
	add := func(p string) {
		if p != "" && !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	for _, root := range roots {
		matches, _ := filepath.Glob(filepath.Join(root, "LibreOffice*", "program", "soffice.exe"))
		for _, m := range matches {
			add(m)
		}
		add(filepath.Join(root, "LibreOffice", "program", "soffice.exe"))
	}
	for _, p := range discoverLibreOfficeFromRegistry() {
		add(p)
	}
	return out
}

func (s *OfficeFileConvertService) convertOfficeWithLibreOffice(file *multipart.FileHeader, lo string) ([]byte, string, string, error) {
	workDir, err := os.MkdirTemp(s.tempBase(), "office-*")
	if err != nil {
		return nil, "", "", err
	}
	defer os.RemoveAll(workDir)

	srcPath := filepath.Join(workDir, filepath.Base(file.Filename))
	if err := saveUpload(file, srcPath); err != nil {
		return nil, "", "", err
	}
	outDir := filepath.Join(workDir, "out")
	_ = os.MkdirAll(outDir, 0o755)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, lo, "--headless", "--nologo", "--nofirststartwizard", "--convert-to", "pdf", "--outdir", outDir, srcPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, "", "", fmt.Errorf("LibreOffice 转换失败: %v %s", err, string(out))
	}
	base := strings.TrimSuffix(filepath.Base(file.Filename), filepath.Ext(file.Filename))
	outFile := filepath.Join(outDir, base+".pdf")
	data, err := os.ReadFile(outFile)
	if err != nil {
		return nil, "", "", fmt.Errorf("未找到输出 PDF: %w", err)
	}
	return data, base + ".pdf", "application/pdf", nil
}
