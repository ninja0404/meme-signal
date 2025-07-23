package json

import (
	"encoding/json"

	"github.com/ninja0404/meme-signal/pkg/config/encoder"
)

const ENCODING_NAME string = "json"

type jsonEncoder struct{}

func (j jsonEncoder) Encode(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (j jsonEncoder) Decode(d []byte, v interface{}) error {
	return json.Unmarshal(d, v)
}

func (j jsonEncoder) String() string {
	return ENCODING_NAME
}

func NewJsonEncoder() encoder.Encoder {
	return jsonEncoder{}
}
