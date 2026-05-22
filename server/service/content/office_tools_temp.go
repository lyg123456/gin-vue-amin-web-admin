package content

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
)

// OfficeToolsTempRoot 办公工具统一临时目录（处理中间文件，非用户下载缓存）
func OfficeToolsTempRoot() string {
	d := strings.TrimSpace(global.GVA_CONFIG.OfficeTools.TempDir)
	if d == "" {
		d = filepath.Join(os.TempDir(), "gva-office-tools")
	}
	_ = os.MkdirAll(d, 0o755)
	return d
}

func officeToolsRetention() time.Duration {
	h := global.GVA_CONFIG.OfficeTools.RetentionHours
	if h <= 0 {
		h = 24
	}
	return time.Duration(h) * time.Hour
}

// CleanOfficeToolsExpired 删除超过保留期的临时文件/目录
func CleanOfficeToolsExpired() (removed int, err error) {
	cutoff := time.Now().Add(-officeToolsRetention())
	roots := officeToolsTempRoots()
	for _, root := range roots {
		n, walkErr := cleanOfficePathOlderThan(root, cutoff)
		removed += n
		if walkErr != nil {
			err = walkErr
		}
	}
	if removed > 0 {
		global.GVA_LOG.Info("办公工具临时数据清理完成",
			zap.Int("removed", removed),
			zap.Duration("retention", officeToolsRetention()),
		)
	}
	return removed, err
}

func officeToolsTempRoots() []string {
	seen := map[string]bool{}
	var out []string
	add := func(p string) {
		if p == "" || seen[p] {
			return
		}
		seen[p] = true
		if st, err := os.Stat(p); err == nil && st.IsDir() {
			out = append(out, p)
		} else {
			_ = os.MkdirAll(p, 0o755)
			out = append(out, p)
		}
	}
	add(OfficeToolsTempRoot())
	tmp := os.TempDir()
	add(filepath.Join(tmp, "gva-office-media"))
	add(filepath.Join(tmp, "gva-office-convert"))
	add(filepath.Join(tmp, "gva-office-tools"))
	return out
}

func cleanOfficePathOlderThan(root string, cutoff time.Time) (int, error) {
	st, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	if !st.IsDir() {
		return 0, nil
	}
	removed := 0
	entries, err := os.ReadDir(root)
	if err != nil {
		return 0, err
	}
	for _, ent := range entries {
		path := filepath.Join(root, ent.Name())
		info, err := ent.Info()
		if err != nil {
			continue
		}
		// 跳过正在使用的当日空目录占位可保留；按修改时间删除
		if info.ModTime().Before(cutoff) {
			if err := os.RemoveAll(path); err == nil {
				removed++
			}
		}
	}
	return removed, nil
}

// LogOfficeDataPolicy 启动时说明数据策略
func LogOfficeDataPolicy() {
	global.GVA_LOG.Info("办公工具数据策略：生成/爬取结果以 HTTP 流式返回不落库；服务器临时目录超过保留期自动删除",
		zap.Duration("retention", officeToolsRetention()),
		zap.String("tempRoot", OfficeToolsTempRoot()),
	)
}
