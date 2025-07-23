package mse

import (
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

func createClient(conf *MseConfig) (config_client.IConfigClient, error) {
	serverCfg := []constant.ServerConfig{
		{
			IpAddr: conf.ServerAddr,
			Port:   8848,
		},
	}
	clientCfg := constant.ClientConfig{
		NamespaceId:         conf.NamespaceID,
		AccessKey:           conf.AccessKey,
		SecretKey:           conf.SecretKey,
		TimeoutMs:           5 * 1000, // 5 seconds
		NotLoadCacheAtStart: true,
		//LogDir:              opts.LogDir,
		//CacheDir:            opts.CacheDir,
	}
	configClient, err := clients.NewConfigClient(vo.NacosClientParam{
		ClientConfig:  &clientCfg,
		ServerConfigs: serverCfg,
	})
	if err != nil {
		return nil, err
	}

	return configClient, nil
}
