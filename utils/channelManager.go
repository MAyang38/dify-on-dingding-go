package utils

import (
	"sync"
)

type ChannelManager struct {
	DataCh    chan string
	CloseCh   chan struct{}
	CardStart chan struct{}
	closeOnce sync.Once
	mutex     sync.Mutex
	isClosed  bool
}

func NewChannelManager() *ChannelManager {
	return &ChannelManager{
		DataCh:    make(chan string, 1),
		CloseCh:   make(chan struct{}),
		CardStart: make(chan struct{}),
	}
}

func (cm *ChannelManager) CloseChannel() {
	cm.closeOnce.Do(func() {
		close(cm.CloseCh)
		close(cm.DataCh)
		close(cm.CardStart)
		cm.mutex.Lock()
		defer cm.mutex.Unlock()
		cm.isClosed = true
	})
}

func (cm *ChannelManager) IsClosed() bool {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	return cm.isClosed
}
