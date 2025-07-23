package polardbx

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/hashicorp/go-multierror"

	cnf "github.com/ninja0404/meme-signal/pkg/config"
	"github.com/ninja0404/meme-signal/pkg/logger"
)

const (
	DEFAULT_DB     = "default"
	DEFAULT_CONFIG = "polarx"
)

var dbs map[string]*gorm.DB

func init() {
	dbs = make(map[string]*gorm.DB)
}

func SetupDatabaseFromDefaultConfig() error {
	return setupDatabaseFromConfig(DEFAULT_DB, DEFAULT_CONFIG)
}

func setupDatabaseFromConfig(name string, configKey string) (err error) {
	var config MysqlConfig
	err = cnf.Get(configKey).Scan(&config)
	if err != nil {
		return err
	}
	newDB, err := createDatabase(&config)
	if err != nil {
		return err
	}
	dbs[name] = newDB.db
	logger.Info(
		"mysql database connected",
		logger.String("name", name),
		logger.String("host", config.Host),
		logger.Int("port", config.Port),
		logger.String("database", config.Database),
	)
	return nil
}

func Stop() error {
	var merr error
	for dname, db := range dbs {
		realDB, err := db.DB()
		if err != nil {
			merr = multierror.Append(merr, err)
		}
		err = realDB.Close()
		if err != nil {
			merr = multierror.Append(merr, err)
		}
		logger.Info(
			"mysql database closed",
			logger.String("name", dname),
		)
	}
	return merr
}

func GetDb() (*gorm.DB, error) {
	return GetDbWithName(DEFAULT_DB)
}

func GetDbWithName(name string) (*gorm.DB, error) {
	db, ok := dbs[name]
	if !ok {
		return nil, errors.New("database does not initialized")
	}
	return db, nil
}
