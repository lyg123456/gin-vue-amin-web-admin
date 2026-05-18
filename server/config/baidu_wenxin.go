package config

// BaiduWenxin 百度千帆文心：优先 V2 OpenAPI（Bearer）；未配置 v2-api-key 时回退 OAuth + aip 旧版 chat URL
type BaiduWenxin struct {
	// V2APIKey 千帆控制台「API Key」（Authorization: Bearer），与旧版 api-key/secret-key 二选一；同时配置时优先 V2
	V2APIKey string `mapstructure:"v2-api-key" json:"v2-api-key" yaml:"v2-api-key"`
	// V2ChatURL 默认 https://qianfan.baidubce.com/v2/chat/completions
	V2ChatURL string `mapstructure:"v2-chat-url" json:"v2-chat-url" yaml:"v2-chat-url"`
	// V2Model 请求体 model 字段，如 ernie-speed-128k
	V2Model string `mapstructure:"v2-model" json:"v2-model" yaml:"v2-model"`

	ApiKey    string `mapstructure:"api-key" json:"api-key" yaml:"api-key"`          // 旧版 OAuth client_id
	SecretKey string `mapstructure:"secret-key" json:"secret-key" yaml:"secret-key"` // 旧版 OAuth client_secret
	// ModelEndpoint 旧版：aip.baidubce.com .../wenxinworkshop/chat/服务名
	ModelEndpoint string `mapstructure:"model-endpoint" json:"model-endpoint" yaml:"model-endpoint"`
	Timeout       int    `mapstructure:"timeout" json:"timeout" yaml:"timeout"` // 单次请求超时（秒），默认 120
}
