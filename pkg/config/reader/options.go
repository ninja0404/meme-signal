package reader

import (
	"github.com/ninja0404/meme-signal/pkg/config/encoder"
	"github.com/ninja0404/meme-signal/pkg/config/encoder/json"
	"github.com/ninja0404/meme-signal/pkg/config/encoder/toml"
	"github.com/ninja0404/meme-signal/pkg/config/encoder/yaml"
)

type Options struct {
	Encoding map[string]encoder.Encoder
}

type Option func(o *Options)

func NewOptions(opts ...Option) Options {
	options := Options{
		Encoding: map[string]encoder.Encoder{
			"json": json.NewJsonEncoder(),
			"yaml": yaml.NewYamlEncoder(),
			"toml": toml.NewTomlEncoder(),
		},
	}
	for _, o := range opts {
		o(&options)
	}
	return options
}

func WithEncoder(e encoder.Encoder) Option {
	return func(o *Options) {
		if o.Encoding == nil {
			o.Encoding = make(map[string]encoder.Encoder)
		}
		o.Encoding[e.String()] = e
	}
}
