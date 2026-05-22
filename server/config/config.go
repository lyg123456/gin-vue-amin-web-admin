package config

type Server struct {
	JWT       JWT     `mapstructure:"jwt" json:"jwt" yaml:"jwt"`
	Zap       Zap     `mapstructure:"zap" json:"zap" yaml:"zap"`
	Redis     Redis   `mapstructure:"redis" json:"redis" yaml:"redis"`
	RedisList []Redis `mapstructure:"redis-list" json:"redis-list" yaml:"redis-list"`
	Mongo     Mongo   `mapstructure:"mongo" json:"mongo" yaml:"mongo"`
	Email     Email   `mapstructure:"email" json:"email" yaml:"email"`
	System    System  `mapstructure:"system" json:"system" yaml:"system"`
	Captcha   Captcha `mapstructure:"captcha" json:"captcha" yaml:"captcha"`
	// auto
	AutoCode Autocode `mapstructure:"autocode" json:"autocode" yaml:"autocode"`
	// gorm
	Mysql  Mysql           `mapstructure:"mysql" json:"mysql" yaml:"mysql"`
	Mssql  Mssql           `mapstructure:"mssql" json:"mssql" yaml:"mssql"`
	Pgsql  Pgsql           `mapstructure:"pgsql" json:"pgsql" yaml:"pgsql"`
	Oracle Oracle          `mapstructure:"oracle" json:"oracle" yaml:"oracle"`
	Sqlite Sqlite          `mapstructure:"sqlite" json:"sqlite" yaml:"sqlite"`
	DBList []SpecializedDB `mapstructure:"db-list" json:"db-list" yaml:"db-list"`
	// oss
	Local        Local        `mapstructure:"local" json:"local" yaml:"local"`
	Qiniu        Qiniu        `mapstructure:"qiniu" json:"qiniu" yaml:"qiniu"`
	AliyunOSS    AliyunOSS    `mapstructure:"aliyun-oss" json:"aliyun-oss" yaml:"aliyun-oss"`
	HuaWeiObs    HuaWeiObs    `mapstructure:"hua-wei-obs" json:"hua-wei-obs" yaml:"hua-wei-obs"`
	TencentCOS   TencentCOS   `mapstructure:"tencent-cos" json:"tencent-cos" yaml:"tencent-cos"`
	AwsS3        AwsS3        `mapstructure:"aws-s3" json:"aws-s3" yaml:"aws-s3"`
	CloudflareR2 CloudflareR2 `mapstructure:"cloudflare-r2" json:"cloudflare-r2" yaml:"cloudflare-r2"`
	Minio        Minio        `mapstructure:"minio" json:"minio" yaml:"minio"`

	Excel Excel `mapstructure:"excel" json:"excel" yaml:"excel"`

	DiskList []DiskList `mapstructure:"disk-list" json:"disk-list" yaml:"disk-list"`

	// 跨域配置
	Cors CORS `mapstructure:"cors" json:"cors" yaml:"cors"`

	// MCP配置
	MCP MCP `mapstructure:"mcp" json:"mcp" yaml:"mcp"`

	// 百度文心（千帆）— 后台 AI 写文章等
	BaiduWenxin BaiduWenxin `mapstructure:"baidu-wenxin" json:"baidu-wenxin" yaml:"baidu-wenxin"`
	// 火山方舟（豆包）— 配置 api-key 时优先于百度用于 AI 写文章
	VolcArk VolcArk `mapstructure:"volc-ark" json:"volc-ark" yaml:"volc-ark"`
	// 火山方舟短视频成片（可选，未配置时仅生成脚本，需手动上传成片）
	VolcArkVideo VolcArkVideo `mapstructure:"volc-ark-video" json:"volc-ark-video" yaml:"volc-ark-video"`
	// 阿里云 DashScope 短视频成片（video-synthesis 异步 API）
	DashScopeVideo DashScopeVideo `mapstructure:"dashscope-video" json:"dashscope-video" yaml:"dashscope-video"`
	// 短视频成片异步 Worker（goroutine + channel + Redis）
	VideoAsync VideoAsync `mapstructure:"video-async" json:"video-async" yaml:"video-async"`
	// 门户办公工具（临时邮箱代理、文件转换）
	OfficeTools OfficeTools `mapstructure:"office-tools" json:"office-tools" yaml:"office-tools"`
}
