package content

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

type xhsSignBridgeReq struct {
	Method  string                 `json:"method"`
	URI     string                 `json:"uri"`
	Cookie  string                 `json:"cookie"`
	Params  map[string]interface{} `json:"params,omitempty"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

type xhsSignBridgeResp struct {
	OK      bool              `json:"ok"`
	Headers map[string]string `json:"headers"`
	Error   string            `json:"error"`
}

func xhsSignRequest(req xhsSignBridgeReq) (map[string]string, error) {
	py := findPythonExecutable()
	if py == "" {
		return nil, fmt.Errorf("未找到 Python 3。请在 config.yaml 配置 office-tools.xhs-python-path（venv 的 python3），或设置环境变量 XHS_PYTHON，并 pip install xhshow")
	}
	script, err := locateXhsSignScript()
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, py, script)
	cmd.Stdin = bytes.NewReader(raw)
	cmd.Env = append(os.Environ(),
		"PYTHONIOENCODING=utf-8",
		"PYTHONUTF8=1",
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("签名脚本失败: %s", msg)
	}
	var resp xhsSignBridgeResp
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		return nil, fmt.Errorf("解析签名结果失败: %w", err)
	}
	if !resp.OK {
		if resp.Error == "" {
			resp.Error = "xhshow 签名失败"
		}
		return nil, fmt.Errorf(resp.Error)
	}
	return resp.Headers, nil
}

func findPythonExecutable() string {
	for _, p := range xhsPythonCandidates() {
		if p == "" {
			continue
		}
		if st, err := os.Stat(p); err != nil || st.IsDir() {
			continue
		}
		if pythonRuns(p) {
			return p
		}
	}
	return ""
}

func xhsPythonCandidates() []string {
	var out []string
	seen := make(map[string]struct{})
	add := func(p string) {
		p = strings.TrimSpace(p)
		if p == "" {
			return
		}
		p = filepath.Clean(p)
		if _, ok := seen[p]; ok {
			return
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}

	if v := strings.TrimSpace(global.GVA_CONFIG.OfficeTools.XhsPythonPath); v != "" {
		add(v)
	}
	if v := os.Getenv("XHS_PYTHON"); v != "" {
		add(v)
	}
	if home := strings.TrimSpace(os.Getenv("XHS_PYTHON_HOME")); home != "" {
		add(filepath.Join(home, "python.exe"))
		add(filepath.Join(home, "bin", "python3"))
	}

	// 宝塔/Linux 部署常用 venv（server/.venv 或项目根 .venv）
	for _, rel := range []string{
		"server/.venv/bin/python3",
		".venv/bin/python3",
		"venv/bin/python3",
	} {
		if wd, err := os.Getwd(); err == nil {
			add(filepath.Join(wd, rel))
		}
	}
	if exe, err := os.Executable(); err == nil {
		root := filepath.Dir(filepath.Dir(exe)) // bin/server -> project or server parent
		for _, rel := range []string{
			filepath.Join("server", ".venv", "bin", "python3"),
			filepath.Join(".venv", "bin", "python3"),
		} {
			add(filepath.Join(root, rel))
		}
	}

	// 常见自定义安装目录（含用户本机 D:\python\py3.10）
	for _, p := range []string{
		`D:\python\py3.10\python.exe`,
		`C:\Python310\python.exe`,
		`C:\Python311\python.exe`,
	} {
		add(p)
	}
	if local := os.Getenv("LOCALAPPDATA"); local != "" {
		for _, ver := range []string{"Python312", "Python311", "Python310"} {
			add(filepath.Join(local, "Programs", "Python", ver, "python.exe"))
		}
	}

	for _, name := range []string{"python3", "python", "py"} {
		p, err := exec.LookPath(name)
		if err != nil || isWindowsStorePythonStub(p) {
			continue
		}
		add(p)
	}
	return out
}

// Windows 未安装 Python 时，PATH 里常有 WindowsApps 占位符，执行会失败。
func isWindowsStorePythonStub(path string) bool {
	return strings.Contains(strings.ToLower(filepath.Clean(path)), `\windowsapps\`)
}

func pythonRuns(py string) bool {
	cmd := exec.Command(py, "--version")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

func locateXhsSignScript() (string, error) {
	seen := make(map[string]struct{})
	var candidates []string
	add := func(p string) {
		p = strings.TrimSpace(p)
		if p == "" {
			return
		}
		p = filepath.Clean(p)
		if _, ok := seen[p]; ok {
			return
		}
		seen[p] = struct{}{}
		candidates = append(candidates, p)
	}

	if v := strings.TrimSpace(os.Getenv("XHS_SCRIPT")); v != "" {
		add(v)
	}
	if v := strings.TrimSpace(global.GVA_CONFIG.OfficeTools.XhsScriptPath); v != "" {
		add(v)
	}

	add(filepath.Join("scripts", "xhs_sign_bridge.py"))
	add(filepath.Join("server", "scripts", "xhs_sign_bridge.py"))

	if wd, err := os.Getwd(); err == nil {
		add(filepath.Join(wd, "scripts", "xhs_sign_bridge.py"))
		add(filepath.Join(wd, "server", "scripts", "xhs_sign_bridge.py"))
		dir := wd
		for i := 0; i < 6; i++ {
			add(filepath.Join(dir, "server", "scripts", "xhs_sign_bridge.py"))
			add(filepath.Join(dir, "scripts", "xhs_sign_bridge.py"))
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	if exe, err := os.Executable(); err == nil {
		if resolved, err := filepath.EvalSymlinks(exe); err == nil {
			exe = resolved
		}
		dir := filepath.Dir(exe)
		add(filepath.Join(dir, "scripts", "xhs_sign_bridge.py"))
		add(filepath.Join(filepath.Dir(dir), "server", "scripts", "xhs_sign_bridge.py"))
		add(filepath.Join(filepath.Dir(dir), "scripts", "xhs_sign_bridge.py"))
	}

	for _, p := range candidates {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, nil
		}
	}
	return "", fmt.Errorf("未找到 xhs_sign_bridge.py，请在 server/config.yaml 配置 office-tools.xhs-script-path 为脚本绝对路径，并确认已上传 server/scripts/xhs_sign_bridge.py")
}
