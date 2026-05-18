package config

// DashScopeVideo 阿里云百炼 / DashScope 异步视频生成（video-synthesis）
type DashScopeVideo struct {
	ApiKey          string `mapstructure:"api-key" json:"api-key" yaml:"api-key"` // DASHSCOPE_API_KEY
	SynthesisURL    string `mapstructure:"synthesis-url" json:"synthesis-url" yaml:"synthesis-url"`
	TaskURL         string `mapstructure:"task-url" json:"task-url" yaml:"task-url"` // 查询任务，后接 /{task_id}
	Model           string `mapstructure:"model" json:"model" yaml:"model"`           // 有素材图时用 i2v
	T2VModel        string `mapstructure:"t2v-model" json:"t2v-model" yaml:"t2v-model"` // 无图时用文生视频
	Resolution      string `mapstructure:"resolution" json:"resolution" yaml:"resolution"`
	Watermark       bool   `mapstructure:"watermark" json:"watermark" yaml:"watermark"`
	Timeout         int    `mapstructure:"timeout" json:"timeout" yaml:"timeout"` // 秒，含轮询
	PollIntervalSec int    `mapstructure:"poll-interval-sec" json:"poll-interval-sec" yaml:"poll-interval-sec"`
	// PublicFileBaseURL 本地素材对外访问前缀（HTTPS 公网），如 https://your-domain.com；未配置时对本地文件自动转 Base64
	PublicFileBaseURL string `mapstructure:"public-file-base-url" json:"public-file-base-url" yaml:"public-file-base-url"`
}
