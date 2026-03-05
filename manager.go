package payment

import (
	"sync"
)

// 支付管理器
type PaymentManager struct {
	providers map[string]PaymentProvider
	mu        sync.RWMutex
}

func NewPaymentManager() *PaymentManager {
	return &PaymentManager{
		providers: make(map[string]PaymentProvider),
	}
}

// 注册支付渠道
func (m *PaymentManager) Register(provider PaymentProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers[provider.Name()] = provider
}

// 获取支付渠道
func (m *PaymentManager) Get(name string) PaymentProvider {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.providers[name]
}

// GetProvider 获取支付提供者
func (m *PaymentManager) GetProvider(name string) PaymentProvider {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.providers[name]
}

// GetProviders 获取所有支付提供者
func (m *PaymentManager) GetProviders() map[string]PaymentProvider {
	m.mu.RLock()
	defer m.mu.RUnlock()

	providers := make(map[string]PaymentProvider, len(m.providers))
	for name, provider := range m.providers {
		providers[name] = provider
	}
	return providers
}
