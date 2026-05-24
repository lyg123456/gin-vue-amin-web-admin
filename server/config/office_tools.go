package config

// OfficeTools 门户办公工具（临时邮箱代理、文件转换）
type OfficeTools struct {
	// 临时邮箱：mailtm（默认，api.mail.tm）| 1secmail（易被 403）
	TempEmailProvider  string `mapstructure:"temp-email-provider" json:"temp-email-provider" yaml:"temp-email-provider"`
	MailTmAPI          string `mapstructure:"mailtm-api" json:"mailtm-api" yaml:"mailtm-api"`
	TempEmailAPI       string `mapstructure:"temp-email-api" json:"temp-email-api" yaml:"temp-email-api"`
	MaxUploadMB        int    `mapstructure:"max-upload-mb" json:"max-upload-mb" yaml:"max-upload-mb"`
	LibreOfficePath    string `mapstructure:"libreoffice-path" json:"libreoffice-path" yaml:"libreoffice-path"` // 空则自动探测 soffice/libreoffice
	EnableLibreOffice  bool   `mapstructure:"enable-libreoffice" json:"enable-libreoffice" yaml:"enable-libreoffice"`
	TempDir            string `mapstructure:"temp-dir" json:"temp-dir" yaml:"temp-dir"`
	RetentionHours     int    `mapstructure:"retention-hours" json:"retention-hours" yaml:"retention-hours"` // 临时文件保留小时数，默认 24
	FFmpegPath         string `mapstructure:"ffmpeg-path" json:"ffmpeg-path" yaml:"ffmpeg-path"`
	// 小红书签名 Python（venv 内 python3 绝对路径）；也可用环境变量 XHS_PYTHON
	XhsPythonPath string `mapstructure:"xhs-python-path" json:"xhs-python-path" yaml:"xhs-python-path"`
	// 签名脚本绝对路径（部署时若找不到 scripts 可显式填写）
	XhsScriptPath string `mapstructure:"xhs-script-path" json:"xhs-script-path" yaml:"xhs-script-path"`
	MirrorMaxPages     int    `mapstructure:"mirror-max-pages" json:"mirror-max-pages" yaml:"mirror-max-pages"` // 整站下载最多页面数
	MirrorMaxDepth     int    `mapstructure:"mirror-max-depth" json:"mirror-max-depth" yaml:"mirror-max-depth"`
	MirrorMaxMB        int    `mapstructure:"mirror-max-mb" json:"mirror-max-mb" yaml:"mirror-max-mb"`
}
