package config

// VolcArk 火山引擎方舟（豆包等）OpenAPI chat/completions，Bearer 鉴权
type VolcArk struct {
	ApiKey   string `mapstructure:"api-key" json:"api-key" yaml:"api-key"`       // Authorization: Bearer
	ChatURL  string `mapstructure:"chat-url" json:"chat-url" yaml:"chat-url"`    // 默认北京区 v3 chat
	Model    string `mapstructure:"model" json:"model" yaml:"model"`             // 必填：推理接入点 Endpoint ID，如 ep-20xxxxxxxx（非商品名 Doubao-lite-4k）
	Timeout  int    `mapstructure:"timeout" json:"timeout" yaml:"timeout"`       // 秒，默认 120
}
