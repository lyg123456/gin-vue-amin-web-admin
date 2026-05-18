package config

// VolcArkVideo 火山方舟短视频生成（成片 API，AK 后续在 config.yaml 配置）
type VolcArkVideo struct {
	ApiKey  string `mapstructure:"api-key" json:"api-key" yaml:"api-key"`
	BaseURL string `mapstructure:"base-url" json:"base-url" yaml:"base-url"`
	Model   string `mapstructure:"model" json:"model" yaml:"model"`
	Timeout int    `mapstructure:"timeout" json:"timeout" yaml:"timeout"`
}
