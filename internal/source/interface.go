package source

import (
	"context"

	"github.com/ninja0404/meme-signal/internal/model"
)

// TransactionSource 交易数据源接口
type TransactionSource interface {
	// Start 启动数据源
	Start(ctx context.Context) error

	// Stop 停止数据源
	Stop() error

	// Subscribe 订阅交易数据流
	Subscribe() <-chan *model.Transaction

	// Errors 错误通道
	Errors() <-chan error

	// String 数据源名称
	String() string

	// IsInitialDataLoaded 检查初始数据是否已加载完成
	IsInitialDataLoaded() bool
}

// Manager 数据源管理器
type Manager struct {
	sources   []TransactionSource
	txChan    chan *model.Transaction
	errorChan chan error
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewManager 创建数据源管理器
func NewManager() *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		sources:   make([]TransactionSource, 0),
		txChan:    make(chan *model.Transaction, 100_000), // 缓冲通道
		errorChan: make(chan error, 100),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// AddSource 添加数据源
func (m *Manager) AddSource(source TransactionSource) {
	m.sources = append(m.sources, source)
}

// Start 启动所有数据源
func (m *Manager) Start() error {
	for _, source := range m.sources {
		if err := source.Start(m.ctx); err != nil {
			return err
		}

		// 启动协程监听每个数据源
		go m.listenSource(source)
	}

	return nil
}

// Stop 停止所有数据源
func (m *Manager) Stop() error {
	// 取消上下文
	m.cancel()

	// 停止所有数据源
	for _, source := range m.sources {
		if err := source.Stop(); err != nil {
			return err
		}
	}

	// 关闭通道
	close(m.txChan)
	close(m.errorChan)

	return nil
}

// Transactions 获取交易数据流
func (m *Manager) Transactions() <-chan *model.Transaction {
	return m.txChan
}

// Errors 获取错误流
func (m *Manager) Errors() <-chan error {
	return m.errorChan
}

// IsInitialDataLoaded 检查所有数据源的初始数据是否已加载完成
func (m *Manager) IsInitialDataLoaded() bool {
	for _, source := range m.sources {
		if !source.IsInitialDataLoaded() {
			return false
		}
	}
	return len(m.sources) > 0 // 确保至少有一个数据源
}

// listenSource 监听单个数据源
func (m *Manager) listenSource(source TransactionSource) {
	txChan := source.Subscribe()
	errChan := source.Errors()

	for {
		select {
		case <-m.ctx.Done():
			return
		case tx, ok := <-txChan:
			if !ok {
				return
			}
			select {
			case m.txChan <- tx:
			case <-m.ctx.Done():
				return
			}
		case err, ok := <-errChan:
			if !ok {
				return
			}
			select {
			case m.errorChan <- err:
			case <-m.ctx.Done():
				return
			}
		}
	}
}
