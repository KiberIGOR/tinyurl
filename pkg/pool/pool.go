// Package pool предоставляет generic-пул объектов с методом Reset().
package pool

import "sync"

// Resetter — тип, умеющий сбрасывать своё состояние (генератор reset в cmd/reset).
type Resetter interface {
	Reset()
}

// Pool хранит и переиспользует объекты типа P.
// P — указатель на T, для которого определён Reset().
type Pool[T any, P resettablePtr[T]] struct {
	pool sync.Pool
}

type resettablePtr[T any] interface {
	*T
	Resetter
}

// New создаёт пустой пул. Новые элементы выделяются через new(T).
func New[T any, P resettablePtr[T]]() *Pool[T, P] {
	return &Pool[T, P]{
		pool: sync.Pool{
			New: func() any {
				return P(new(T))
			},
		},
	}
}

// Get возвращает объект из пула или новый экземпляр.
func (p *Pool[T, P]) Get() P {
	v := p.pool.Get()
	if v == nil {
		return P(new(T))
	}
	return v.(P)
}

// Put возвращает объект в пул после сброса состояния.
func (p *Pool[T, P]) Put(v P) {
	if v == nil {
		return
	}
	v.Reset()
	p.pool.Put(v)
}
