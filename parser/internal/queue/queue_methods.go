package queue

// метод для добавления нового элемента в очередь
func (q *FIFOQueue[T]) Enqueue(item T) bool {
	select {
	case q.items <- item:
		return true
	default:
		return false // очередь переполнена
	}
}

// метод для получения элемента из очереди
func (q *FIFOQueue[T]) Dequeue() (T, bool) {
	select {
	case item := <-q.items:
		return item, true
	default:
		var zeroVal T
		return zeroVal, false // очередь пуста
	}
}

// метод для получения размера очереди в данный момент
func (q *FIFOQueue[T]) Size() int {
	return len(q.items)
}
