package yaml

import (
	"github.com/ghodss/yaml"

	"github.com/ninja0404/meme-signal/pkg/config/encoder"
)

const ENCODING_NAME string = "yaml"

type yamlEncoder struct{}

func (y yamlEncoder) Encode(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}

func (y yamlEncoder) Decode(d []byte, v interface{}) error {
	return yaml.Unmarshal(d, v)
}

func (y yamlEncoder) String() string {
	return ENCODING_NAME
}

func NewYamlEncoder() encoder.Encoder {
	return yamlEncoder{}
}
