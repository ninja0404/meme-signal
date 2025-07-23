package source

import "context"

type Options struct {
	// 数据格式
	Format string

	// 自定义的选项
	Context context.Context
}

type Option func(o *Options)

func NewOptions(opts ...Option) Options {
	options := Options{
		Context: context.Background(),
	}

	for _, o := range opts {
		o(&options)
	}

	return options
}

// WithEncoder sets the source encoder
func WithFormat(f string) Option {
	return func(o *Options) {
		if len(f) > 0 && f[0] == '.' {
			o.Format = f[1:]
		} else {
			o.Format = f
		}
	}
}
