package memory

import (
	"github.com/ninja0404/meme-signal/pkg/config/loader"
	"github.com/ninja0404/meme-signal/pkg/config/reader"
	"github.com/ninja0404/meme-signal/pkg/config/source"
)

func WithSource(s source.Source) loader.Option {
	return func(o *loader.Options) {
		o.Source = append(o.Source, s)
	}
}

// WithReader sets the config reader
func WithReader(r reader.Reader) loader.Option {
	return func(o *loader.Options) {
		o.Reader = r
	}
}
