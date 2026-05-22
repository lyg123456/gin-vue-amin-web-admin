//go:build !windows

package content

func discoverLibreOfficeFromRegistry() []string {
	return nil
}
