package kafka

const (
	_  = iota
	KB = 1 << (10 * iota)
	MB = 1 << (10 * iota)
	GB = 1 << (10 * iota)
)

const (
	DefaultVersionRequest  = "true"
	DefaultMessageMaxBytes = 10 * MB
)

type KafkaProducerConfig struct {
	VersionRequest    string `json:"version_request" yaml:"version_request"`
	MessageMaxBytes   int    `json:"message_max_bytes" yaml:"message_max_bytes"`
	LingerMs          int    `json:"linger_ms" yaml:"linger_ms"`
	PartitionLingerMs int    `json:"partition_linger_ms" yaml:"partition_linger_ms"`
	RetryBackoffMs    int    `json:"retry_backoff_ms" yaml:"retry_backoff_ms"`
	RequiredAcks      int    `json:"required_acks" yaml:"required_acks"`
	EnableAsync       bool   `json:"enable_async" yaml:"enable_async"`
	ClientID          string `json:"client_id" yaml:"client_id"`

	SecurityProtocol string `json:"security_protocol" yaml:"security_protocol"`
	SaslUsername     string `json:"sasl_username" yaml:"sasl_username"`
	SaslPassword     string `json:"sasl_password" yaml:"sasl_password"`
	SaslMechanism    string `json:"sasl_mechanism" yaml:"sasl_mechanism"`

	SslCaLocation          string `json:"ssl_ca_location" yaml:"ssl_ca_location"`
	SslCertificateLocation string `json:"ssl_certificate_location" yaml:"ssl_certificate_location"`
	SslKeyLocation         string `json:"ssl_key_location" yaml:"ssl_key_location"`
}
