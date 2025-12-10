package queue

// структура для очереди (дженерики)
type FIFOQueue[T any] struct {
	items chan T
}

// конструктор для очереди
func NewFIFOQueue[T any](capacity int) *FIFOQueue[T] {
	return &FIFOQueue[T]{
		items: make(chan T, capacity),
	}
}
