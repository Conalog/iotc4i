package iotc4i

import (
	"errors"
)

type CircularQueue[T any] struct {
	data     []T
	head     int
	tail     int
	size     int
	capacity int
}

// NewCircularQueue creates a new circular queue with a specified capacity
func NewCircularQueue[T any](capacity int) *CircularQueue[T] {
	return &CircularQueue[T]{
		data:     make([]T, capacity),
		head:     0,
		tail:     0,
		size:     0,
		capacity: capacity,
	}
}

// Enqueue adds an element to the queue
func (q *CircularQueue[T]) Enqueue(value T) error {
	if q.IsFull() {
		return errors.New("queue is full")
	}
	q.data[q.tail] = value
	q.tail = (q.tail + 1) % q.capacity
	q.size++
	return nil
}

// Dequeue removes an element from the queue
func (q *CircularQueue[T]) Dequeue() (T, error) {
	var zeroValue T
	if q.IsEmpty() {
		return zeroValue, errors.New("queue is empty")
	}
	value := q.data[q.head]
	q.head = (q.head + 1) % q.capacity
	q.size--
	return value, nil
}

// IsEmpty checks if the queue is empty
func (q *CircularQueue[T]) IsEmpty() bool {
	return q.size == 0
}

// IsFull checks if the queue is full
func (q *CircularQueue[T]) IsFull() bool {
	return q.size == q.capacity
}

// Size returns the current size of the queue
func (q *CircularQueue[T]) Size() int {
	return q.size
}

// Capacity returns the capacity of the queue
func (q *CircularQueue[T]) Capacity() int {
	return q.capacity
}

// Clear removes all elements from the queue
func (q *CircularQueue[T]) Clear() {
	q.head = 0
	q.tail = 0
	q.size = 0
}

// GetAllData returns all the data currently stored in the queue
func (q *CircularQueue[T]) GetAllData() []T {
	result := make([]T, q.size)
	if q.size == 0 {
		return result
	}

	if q.head < q.tail {
		copy(result, q.data[q.head:q.tail])
	} else {
		copy(result, q.data[q.head:])
		copy(result[q.capacity-q.head:], q.data[:q.tail])
	}

	return result
}
