package toml

import (
	"bytes"

	"github.com/BurntSushi/toml"

	"github.com/ninja0404/meme-signal/pkg/config/encoder"
)

const ENCODING_NAME string = "toml"

type tomlEncoder struct{}

func (t tomlEncoder) Encode(v interface{}) ([]byte, error) {
	b := bytes.NewBuffer(nil)
	defer b.Reset()
	err := toml.NewEncoder(b).Encode(v)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (t tomlEncoder) Decode(d []byte, v interface{}) error {
	return toml.Unmarshal(d, v)
}

func (t tomlEncoder) String() string {
	return ENCODING_NAME
}

func NewTomlEncoder() encoder.Encoder {
	return tomlEncoder{}
}
