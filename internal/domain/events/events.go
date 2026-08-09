package events

import (
	"sync"
	"time"

	"k8soperation/pkg/logger"
)

// ——— 接口 ———

// DomainEvent 领域事件
type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

// EventHandler 事件处理器
type EventHandler func(event DomainEvent)

// EventPublisher 事件发布器
type EventPublisher interface {
	Publish(event DomainEvent)
	Subscribe(eventName string, handler EventHandler)
}

// ——— 基础实现 ———

// BaseEvent 领域事件基类
type BaseEvent struct {
	name       string
	occurredAt time.Time
}

func NewBaseEvent(name string) BaseEvent {
	return BaseEvent{name: name, occurredAt: time.Now()}
}

func (e BaseEvent) EventName() string  { return e.name }
func (e BaseEvent) OccurredAt() time.Time { return e.occurredAt }

// ——— EventBus ———

type EventBus struct {
	mu       sync.RWMutex
	handlers map[string][]EventHandler
	logger   *logger.Logger
}

func NewEventBus(logger *logger.Logger) *EventBus {
	return &EventBus{
		handlers: make(map[string][]EventHandler),
		logger:   logger,
	}
}

func (b *EventBus) Subscribe(eventName string, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventName] = append(b.handlers[eventName], handler)
}

func (b *EventBus) Publish(event DomainEvent) {
	b.mu.RLock()
	handlers := b.handlers[event.EventName()]
	b.mu.RUnlock()

	for _, h := range handlers {
		func() {
			defer func() {
				if r := recover(); r != nil {
					b.logger.Errorf("[EventBus] handler panic for %s: %v", event.EventName(), r)
				}
			}()
			h(event)
		}()
	}
}
