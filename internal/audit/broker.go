package audit

import "context"

// Observer получает события аудита (паттерн «Наблюдатель»).
type Observer interface {
	Notify(ctx context.Context, event Event)
}

// Broker рассылает событие всем подписанным наблюдателям.
type Broker struct {
	observers []Observer
}

func NewBroker() *Broker {
	return &Broker{}
}

func (b *Broker) Subscribe(o Observer) {
	if o == nil {
		return
	}
	b.observers = append(b.observers, o)
}

func (b *Broker) Publish(ctx context.Context, event Event) {
	if b == nil {
		return
	}
	for _, o := range b.observers {
		o.Notify(ctx, event)
	}
}
