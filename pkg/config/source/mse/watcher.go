package mse

import (
	"time"

	"github.com/ninja0404/meme-signal/pkg/config/source"

	"github.com/nacos-group/nacos-sdk-go/vo"
)

type watcher struct {
	mseClient   *mse
	contentChan chan string
	exit        chan bool
}

func newWatcher(mseInstance *mse) (source.Watcher, error) {
	// listening config update
	contentChan := make(chan string)
	err := mseInstance.client.ListenConfig(vo.ConfigParam{
		DataId: mseInstance.config.DataID,
		Group:  mseInstance.config.Group,
		OnChange: func(namespace, group, dataId, data string) {
			contentChan <- data
		},
	})
	if err != nil {
		return nil, err
	}

	return &watcher{
		mseClient:   mseInstance,
		contentChan: contentChan,
		exit:        make(chan bool),
	}, nil
}

func (w *watcher) Next() (*source.ChangeSet, error) {
	// is it closed?
	select {
	case <-w.exit:
		return nil, source.ErrWatcherStopped
	default:
	}

	// try get the event
	select {
	case data := <-w.contentChan:
		cs := &source.ChangeSet{
			Format:    w.mseClient.opts.Format,
			Source:    w.mseClient.String(),
			Timestamp: time.Now(),
			Data:      []byte(data),
		}
		cs.Checksum = cs.Sum()
		return cs, nil
	case <-w.exit:
		return nil, source.ErrWatcherStopped
	}
}

func (w *watcher) Stop() error {
	w.mseClient.client.CancelListenConfig(vo.ConfigParam{
		DataId: w.mseClient.config.DataID,
		Group:  w.mseClient.config.Group,
	})
	return nil
}
