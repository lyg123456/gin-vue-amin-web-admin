//go:build windows

package content

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

func discoverLibreOfficeFromRegistry() []string {
	var out []string
	keys := []registry.Key{
		registry.LOCAL_MACHINE,
		registry.CURRENT_USER,
	}
	subKeys := []string{
		`SOFTWARE\LibreOffice\LibreOffice`,
		`SOFTWARE\WOW6432Node\LibreOffice\LibreOffice`,
		`SOFTWARE\LibreOffice\UNO\InstallPath`,
	}
	for _, root := range keys {
		for _, sub := range subKeys {
			k, err := registry.OpenKey(root, sub, registry.QUERY_VALUE)
			if err != nil {
				continue
			}
			for _, name := range []string{"Path", "InstallPath", ""} {
				val, _, err := k.GetStringValue(name)
				if err != nil || val == "" {
					continue
				}
				candidates := []string{
					filepath.Join(val, "program", "soffice.exe"),
					filepath.Join(val, "soffice.exe"),
				}
				for _, c := range candidates {
					if st, err := os.Stat(c); err == nil && !st.IsDir() {
						out = append(out, c)
					}
				}
			}
			_ = k.Close()
		}
	}
	return out
}
