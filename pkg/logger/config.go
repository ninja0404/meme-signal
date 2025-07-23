package logger

import (
	"fmt"
	"time"

	pconfig "github.com/ninja0404/meme-signal/pkg/config"
)

type Config struct {
	// Dir 输出方式：stdout、file、discard
	OUTPUT string `yaml:"output" json:"output" toml:"output"`
	// Dir 日志目录
	Dir string `yaml:"dir" json:"dir" toml:"dir"`
	// Name 日志文件名
	Name string `yaml:"name" json:"name" toml:"name"`
	// Level 日志等级
	Level string `yaml:"level" json:"level" toml:"level"`
	// 是否添加调用者信息
	AddCaller bool `yaml:"add_caller" json:"add_caller" toml:"add_caller"`
	// 单文件最大长度(单位: mb)
	MaxSize int `yaml:"max_size" json:"max_size" toml:"max_size"`
	// 日志文件最大保留时间(单位: 天)
	MaxAge int `yaml:"max_age" json:"max_age" toml:"max_age"`
	// 日志副本数
	MaxBackup int `yaml:"max_backup" json:"max_backup" toml:"max_backup"`
	// 日志磁盘刷盘间隔
	Interval time.Duration `yaml:"interval" json:"interval" toml:"interval"`
	// 日志调用者层级
	CallerSkip int `yaml:"caller_skip" json:"caller_skip" toml:"caller_skip"`
	// 异步日志输出(暂不支持)
	Async bool `yaml:"async" json:"async" toml:"async"`
	//FlushBufferSize
	FlushBufferSize int `yaml:"flush_buffer_size" json:"flush_buffer_size" toml:"flush_buffer_size"`
	//FlushInterval
	FlushInterval time.Duration `yaml:"flush_interval" json:"flush_interval" toml:"flush_interval"`
	// 是否是调试状态
	Debug bool `yaml:"debug" json:"debug" toml:"debug"`
	// 日志是否丢弃
	Discard bool `yaml:"discard" json:"discard" toml:"discard"`
	// 禁用Sentry
	DisableSentry bool `yaml:"disable_sentry" json:"disable_sentry" toml:"disable_sentry"`
	// 发送sentry的等级
	SentryLevel string `yaml:"sentry_level" json:"sentry_level" toml:"sentry_level"`
}

func (c *Config) Filename() string {
	return fmt.Sprintf("%s/%s", c.Dir, c.Name)
}

func (c *Config) Build() *Logger {

	logger := newLogger(c)

	return logger
}

func FromConfig(key string) *Config {
	var conf *Config = defaultConfig()
	if err := pconfig.Get(key).Scan(conf); err != nil {
		panic(err)
	}
	return conf
}

func defaultConfig() *Config {
	return &Config{
		Name:            "log",
		OUTPUT:          "stdout",
		Dir:             "./logs/",
		Level:           "info",
		MaxSize:         1000, // 1000M
		MaxAge:          1,    // 1 day
		MaxBackup:       10,   // 10 backup
		Interval:        24 * time.Hour,
		CallerSkip:      0,
		AddCaller:       true,
		Async:           false,
		FlushBufferSize: 256 * 1024,
		FlushInterval:   5 * time.Second,
		SentryLevel:     "error",
	}
}
