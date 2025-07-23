package mse

import (
	"context"

	"github.com/ninja0404/meme-signal/pkg/config/source"
)

type mseConfigKey struct{}

type MseConfig struct {
	ServerAddr  string `env:"MSE_SERVER_ADDR,required"`
	NamespaceID string `env:"MSE_NAMESPACE,required"`
	AccessKey   string `env:"MSE_ACCESSKEY,required"`
	SecretKey   string `env:"MSE_SECRETKEY,required"`
	Group       string `env:"MSE_GROUP,required"`
	DataID      string `env:"MSE_DATAID,required"`
	LogDir      string `env:"MSE_LOG_DIR,empty"`
	CacheDir    string `env:"MSE_CACHE_DIR,empty"`
}

func WithMseConfig(conf *MseConfig) source.Option {
	return func(o *source.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, mseConfigKey{}, conf)
	}
}
