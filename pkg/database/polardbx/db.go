package polardbx

import (
	"database/sql"
	"fmt"
	"time"

	mysqlDriver "gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/ninja0404/meme-signal/pkg/logger"
)

type MysqlWrapper struct {
	db     *gorm.DB
	sqldb  *sql.DB
	config *MysqlConfig
}

type MysqlConfig struct {
	Name     string `yaml:"name" json:"name" toml:"name"`
	User     string `yaml:"user" json:"user" toml:"user"`
	Password string `yaml:"password" json:"password" toml:"password"`
	Host     string `yaml:"host" json:"host" toml:"host"`
	Port     int    `yaml:"port" json:"port" toml:"port"`
	Database string `yaml:"database" json:"database" toml:"database"`

	Timeout string `yaml:"timeout" json:"timeout" toml:"timeout"` // connect timeout

	MaxPoolSize     int           `yaml:"max_pool_size" json:"max_pool_size" toml:"max_pool_size"`
	MaxIdleSize     int           `yaml:"max_idle_size" json:"max_idle_size" toml:"max_idle_size"`
	MaxIdleDuration time.Duration `yaml:"max_idle_ts" json:"max_idle_ts" toml:"max_idle_ts"`
	SqlOpenDebug    bool          `yaml:"open_debug" json:"open_debug" toml:"open_debug"`
	LogLevel        string        `yaml:"log_level" json:"log_level" toml:"log_level"`
}

func createDatabase(srcConf *MysqlConfig) (*MysqlWrapper, error) {
	cnf := validateConfig(srcConf)
	dsn := getDsn(cnf)

	gormConfig := gorm.Config{
		Logger: NewMysqlLogger(logger.DefaultL1().Named("polardbx"), mappingLoggerLevel(cnf.LogLevel, cnf.SqlOpenDebug)),
	}

	db, err := gorm.Open(mysqlDriver.Open(dsn), &gormConfig)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	err = sqlDB.Ping()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(cnf.MaxPoolSize)
	sqlDB.SetMaxIdleConns(cnf.MaxIdleSize)
	sqlDB.SetConnMaxIdleTime(cnf.MaxIdleDuration)
	sqlDB.SetConnMaxLifetime(time.Minute * 30)

	return &MysqlWrapper{
		db:     db,
		sqldb:  sqlDB,
		config: cnf,
	}, nil
}

func (db *MysqlWrapper) DB() *gorm.DB {
	return db.db
}

func (db *MysqlWrapper) Close() error {
	return db.sqldb.Close()
}

func validateConfig(src *MysqlConfig) *MysqlConfig {
	dst := *src

	if src.MaxPoolSize == 0 {
		dst.MaxPoolSize = 20
	}

	if src.MaxIdleSize == 0 {
		dst.MaxIdleSize = 10
	}

	if src.MaxIdleDuration == 0 {
		dst.MaxIdleDuration = 10 * time.Minute
	}
	return &dst
}

func getDsn(cnf *MysqlConfig) string {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=%s", cnf.User, cnf.Password, cnf.Host, cnf.Port, cnf.Database, cnf.Timeout)
	return dsn
}
