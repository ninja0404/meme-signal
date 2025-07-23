package file

import (
	"io/ioutil"
	"os"

	"github.com/ninja0404/meme-signal/pkg/config/source"
)

type file struct {
	path string
	opts source.Options
}

const (
	DEFAULT_CONFIG_FILE_NAME   = "config"
	DEFAULT_CONFIG_FILE_FORMAT = "json"
)

func (f *file) Read() (*source.ChangeSet, error) {
	fh, err := os.Open(f.path)
	if err != nil {
		return nil, err
	}
	defer fh.Close()

	info, err := fh.Stat()
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(fh)
	if err != nil {
		return nil, err
	}

	cs := &source.ChangeSet{
		Format:    f.opts.Format,
		Source:    f.String(),
		Timestamp: info.ModTime(),
		Data:      b,
	}
	cs.Checksum = cs.Sum()

	return cs, nil
}

func (f *file) String() string {
	return "file"
}

func (f *file) Watch() (source.Watcher, error) {
	if _, err := os.Stat(f.path); err != nil {
		return nil, err
	}
	return newWatcher(f)
}

func (f *file) Write(cs *source.ChangeSet) error {
	return nil
}

func NewSource(opts ...source.Option) source.Source {
	options := source.NewOptions(opts...)

	if options.Format == "" {
		options.Format = DEFAULT_CONFIG_FILE_FORMAT
	}

	path, ok := options.Context.Value(filePathKey{}).(string)
	if !ok {
		path = DEFAULT_CONFIG_FILE_NAME + "." + options.Format
	}

	return &file{opts: options, path: path}
}
