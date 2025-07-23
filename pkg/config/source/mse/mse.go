package mse

import (
	"time"

	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/vo"

	"github.com/ninja0404/meme-signal/pkg/config/source"
)

const DEFAULT_GROUP string = "DEFAULT_GROUP"

type mse struct {
	client config_client.IConfigClient
	config *MseConfig
	opts   source.Options
}

func (s *mse) Read() (*source.ChangeSet, error) {
	configContent, err := s.client.GetConfig(vo.ConfigParam{
		Group:  s.config.Group,
		DataId: s.config.DataID,
	})
	if err != nil {
		return nil, err
	}

	cs := &source.ChangeSet{
		Format:    s.opts.Format,
		Source:    s.String(),
		Timestamp: time.Now(),
		Data:      []byte(configContent),
	}
	cs.Checksum = cs.Sum()

	return cs, nil
}

func (s *mse) String() string {
	return "mse"
}

func (s *mse) Watch() (source.Watcher, error) {
	return newWatcher(s)
}

func (s *mse) Write(cs *source.ChangeSet) error {
	return nil
}

func NewSource(opts ...source.Option) source.Source {
	options := source.NewOptions(opts...)
	mseConfig, ok := options.Context.Value(mseConfigKey{}).(*MseConfig)
	if !ok {
		panic("mse config not provided")
	}

	client, err := createClient(mseConfig)
	if err != nil {
		panic(err)
	}

	return &mse{opts: options, client: client, config: mseConfig}
}
