package interfaces

// Интерфейс с дженериком для FIFO очереди
type FIFOQueueInterface[T any] interface {
	Enqueue(item T) bool
	Dequeue() (T, bool)
	Size() int
}
