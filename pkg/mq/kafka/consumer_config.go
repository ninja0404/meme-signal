package kafka

import (
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaConsumerConfig struct {
	Topics       []string `json:"topics" yaml:"topics"`
	Version      string   `json:"version" yaml:"version"`
	Assignor     string   `json:"assignor" yaml:"assignor"`
	OffsetIntial string   `json:"offset_initial" yaml:"offset_initial"`
	GroupId      string   `json:"group_id" yaml:"group_id"`
	ClientID     string   `json:"client_id" yaml:"client_id"`

	EnableAutoCommit   bool `json:"enable_auto_commit" yaml:"enable_auto_commit"`
	EnableAutoStore    bool `json:"enable_auto_store" yaml:"enable_auto_store"`
	AutoCommitInterval int  `json:"auto_commit_interval" yaml:"auto_commit_interval"`
	MaxPollInterval    int  `json:"max_poll_interval" yaml:"max_poll_interval"`
	SessionTimeout     int  `json:"session_timeout" yaml:"session_timeout"`
	HeartbeatInterval  int  `json:"heartbeat_interval" yaml:"heartbeat_interval"`
	ReadTimeout        int  `json:"read_timeout" yaml:"read_timeout"`

	SecurityProtocol string `json:"security_protocol" yaml:"security_protocol"`
	SaslUsername     string `json:"sasl_username" yaml:"sasl_username"`
	SaslPassword     string `json:"sasl_password" yaml:"sasl_password"`
	SaslMechanism    string `json:"sasl_mechanism" yaml:"sasl_mechanism"`

	SslCaLocation          string `json:"ssl_ca_location" yaml:"ssl_ca_location"`
	SslCertificateLocation string `json:"ssl_certificate_location" yaml:"ssl_certificate_location"`
	SslKeyLocation         string `json:"ssl_key_location" yaml:"ssl_key_location"`
}

func newConsumerConfig(brokers []string, cfg KafkaConsumerConfig) *kafka.ConfigMap {
	//common arguments
	var kafkaconf = &kafka.ConfigMap{
		"api.version.request":      "true",
		"auto.offset.reset":        "latest",
		"enable.auto.commit":       cfg.EnableAutoCommit, // 是否启用自动提交位点。设置为true，表示启用自动提交位点。
		"enable.auto.offset.store": cfg.EnableAutoStore,  // 是否启用自动 Store，默认是true
		"auto.commit.interval.ms":  5000,                 // 自动提交位点的间隔时间。设置为5000毫秒（即5秒），表示每5秒自动提交一次位点。
		"max.poll.interval.ms":     300000,               // Consumer在一次poll操作中最长的等待时间。设置为300000毫秒（即5分钟），表示Consumer在一次poll操作中最多等待5分钟
		"session.timeout.ms":       10000,                // 指定Consumer与broker之间的会话超时时间,设置10秒
		"heartbeat.interval.ms":    3000,                 // 指定Consumer发送心跳消息的间隔时间。设置为3000毫秒（即3秒）
	}

	kafkaconf.SetKey("group.id", cfg.GroupId)
	if cfg.ClientID != "" {
		kafkaconf.SetKey("client.id", cfg.ClientID+"_"+getClientID())
	}
	if cfg.OffsetIntial != "" {
		kafkaconf.SetKey("auto.offset.reset", cfg.OffsetIntial)
	}
	if cfg.AutoCommitInterval > 0 {
		kafkaconf.SetKey("auto.commit.interval.ms", cfg.AutoCommitInterval)
	}

	bootstrapServers := strings.Join(brokers, ",")
	kafkaconf.SetKey("bootstrap.servers", bootstrapServers)

	switch cfg.SecurityProtocol {
	case "PLAINTEXT", "":
		kafkaconf.SetKey("security.protocol", "plaintext")
	case "SASL_SSL":
		kafkaconf.SetKey("security.protocol", "sasl_ssl")
		//kafkaconf.SetKey("sasl.mechanism", cfg.SaslMechanism)
		kafkaconf.SetKey("sasl.username", cfg.SaslUsername)
		kafkaconf.SetKey("sasl.password", cfg.SaslPassword)
		kafkaconf.SetKey("ssl.ca.location", cfg.SslCaLocation)
		kafkaconf.SetKey("ssl.certificate.location", cfg.SslCertificateLocation)
		kafkaconf.SetKey("ssl.key.location", cfg.SslKeyLocation)
		// hostname校验改成空
		kafkaconf.SetKey("ssl.endpoint.identification.algorithm", "None")
		kafkaconf.SetKey("enable.ssl.certificate.verification", "false")
	case "SSL":
		kafkaconf.SetKey("security.protocol", "ssl")
		//kafkaconf.SetKey("sasl.mechanism", cfg.SaslMechanism)
		kafkaconf.SetKey("ssl.ca.location", cfg.SslCaLocation)
		kafkaconf.SetKey("ssl.certificate.location", cfg.SslCertificateLocation)
		kafkaconf.SetKey("ssl.key.location", cfg.SslKeyLocation)
		// hostname校验改成空
		//kafkaconf.SetKey("ssl.endpoint.identification.algorithm", "None")
		kafkaconf.SetKey("enable.ssl.certificate.verification", "false")
	case "SASL_PLAINTEXT":
		kafkaconf.SetKey("security.protocol", "SASL_PLAINTEXT")
		kafkaconf.SetKey("sasl.username", cfg.SaslUsername)
		kafkaconf.SetKey("sasl.password", cfg.SaslPassword)
		kafkaconf.SetKey("sasl.mechanism", cfg.SaslMechanism)
	default:
		panic(kafka.NewError(kafka.ErrUnknownProtocol, "unknown protocol:"+cfg.SecurityProtocol, true))
	}

	return kafkaconf
}
