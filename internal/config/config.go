package config

import (
	"github.com/ninja0404/meme-signal/pkg/config"
	"github.com/ninja0404/meme-signal/pkg/config/source"
	"github.com/ninja0404/meme-signal/pkg/config/source/file"
	"github.com/ninja0404/meme-signal/pkg/database/polardbx"
	"github.com/ninja0404/meme-signal/pkg/logger"
)

// AppConfig 应用配置结构
type AppConfig struct {
	Logger LoggerConfig         `yaml:"logger"`
	PolarX polardbx.MysqlConfig `yaml:"polarx"`
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Output     string `yaml:"output"`
	Debug      bool   `yaml:"debug"`
	Level      string `yaml:"level"`
	AddCaller  bool   `yaml:"add_caller"`
	CallerSkip int    `yaml:"caller_skip"`
}

// DatabaseConfig 数据库数据源配置
type DatabaseConfig struct {
	QueryInterval     int `yaml:"query_interval"`      // 查询间隔（秒）
	InitWindowMinutes int `yaml:"init_window_minutes"` // 初始查询窗口（分钟）
	BatchSize         int `yaml:"batch_size"`          // 批量查询大小
}

// Manager 配置管理器
type Manager struct {
	config *AppConfig
}

// NewManager 创建配置管理器
func NewManager() *Manager {
	return &Manager{}
}

// Load 加载配置文件
func (m *Manager) Load(configPath string) error {
	// 加载配置文件
	err := config.Load(file.NewSource(
		file.WithPath(configPath),
		source.WithFormat("yaml"),
	))
	if err != nil {
		return err
	}

	// 解析配置
	var appConfig AppConfig
	err = config.Scan(&appConfig)
	if err != nil {
		return err
	}

	m.config = &appConfig
	return nil
}

// GetAppConfig 获取应用配置
func (m *Manager) GetAppConfig() *AppConfig {
	return m.config
}

// GetLoggerConfig 获取日志配置
func (m *Manager) GetLoggerConfig() LoggerConfig {
	return m.config.Logger
}

// GetDatabaseConfig 获取数据库配置
func (m *Manager) GetDatabaseConfig() polardbx.MysqlConfig {
	return m.config.PolarX
}

// InitLogger 初始化日志系统
func (m *Manager) InitLogger() error {
	loggerConfig := logger.FromConfig("logger")
	loggerInstance := loggerConfig.Build()
	logger.SetDefault(loggerInstance)
	logger.SetDefaultL1(loggerInstance)
	return nil
}
