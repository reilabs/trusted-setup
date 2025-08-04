package unique_fifo

import (
	"errors"
	"sync"
)

// Queue defines the contract for a queue.
//
// It allows adding and removing elements and peeking at elements without
// removing them and checking the count of elements.
type Queue[T comparable] interface {
	// Enqueue adds value to the queue. It returns the key of the value, that can be used in PeekByKey.
	Enqueue(value T) int
	// Dequeue removes the next eligible value from the queue and returns it.
	Dequeue() (value T, err error)
	// Peek returns the next eligible value from the queue without removing it.
	Peek() (value T, err error)
	// PeekByKey returns the value by the key without removing it.
	PeekByKey(key int) (value T, err error)
	// PeekAll returns all values from the queue without removing them.
	PeekAll() (value []T, err error)
	// Len returns the number of elements in the queue.
	Len() int
}

var (
	ErrEmpty         = errors.New("fifo empty")
	ErrKeyOutOfRange = errors.New("key out of range")
)

// uniqueFifo is a thread-safe queue that only stores unique values.
//
// The queue can hold any type of values that satisfy the comparable interface.
type uniqueFifo[T comparable] struct {
	lock   sync.RWMutex
	values []T
	seen   map[T]int
}

// New creates a new unique FIFO queue.
//
// Example:
//
//	q := New[int]()
func New[T comparable]() Queue[T] {
	return &uniqueFifo[T]{
		values: make([]T, 0),
		seen:   make(map[T]int),
	}
}

// Enqueue adds a new value to the queue if it is not already present.
//
// If the value is already present, it is not added.
//
// Example:
//
//	k := q.Enqueue(1) // k == 0
//	l := q.Enqueue(2) // l == 1
//	m := q.Enqueue(1) // 1 is not added twice, m == 0
//	n := q.Enqueue(3) // n == 2
//	o := q.Enqueue(2) // 2 is not added twice // o == 1
func (q *uniqueFifo[T]) Enqueue(value T) int {
	q.lock.Lock()
	defer q.lock.Unlock()

	if key, exists := q.seen[value]; exists {
		return key
	}

	q.values = append(q.values, value)
	key := len(q.values) - 1
	q.seen[value] = key
	return key
}

// Dequeue removes the mos recently enqueued value from the queue and returns it.
//
// If the queue is empty, an error is returned.
//
// Example:
//
//	q.Enqueue(1)
//	q.Enqueue(2)
//	q.Enqueue(3)
//	q.Dequeue() // returns 1
//	q.Dequeue() // returns 2
//	q.Dequeue() // returns 3
//	q.Dequeue() // returns error
func (q *uniqueFifo[T]) Dequeue() (value T, err error) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if len(q.values) == 0 {
		err = ErrEmpty
		return
	}
	value = q.values[0]
	q.values = q.values[1:]
	delete(q.seen, value)

	// Update keys
	for k, v := range q.values {
		q.seen[v] = k
	}
	return
}

// Peek returns the most recently enqueued value from the queue without removing it.
//
// If the queue is empty, an error is returned.
//
// Example:
//
//	q.Enqueue(1)
//	q.Enqueue(2)
//	q.Enqueue(3)
//	q.Peek() // returns 1
//	q.Peek() // returns 1
//	q.Dequeue() // returns 1
//	q.Peek() // returns 2
func (q *uniqueFifo[T]) Peek() (value T, err error) {
	return q.PeekByKey(0)
}

// PeekByKey returns the value at the given key from the queue without removing it.
//
// If the queue is empty, an error is returned.
//
// Example:
//
//	q.Enqueue(1)
//	q.Enqueue(2)
//	q.Enqueue(3)
//	q.PeekByKey(0) // returns 1
//	q.PeekByKey(1) // returns 2
//	q.PeekByKey(2) // returns 3
//	q.PeekByKey(3) // returns error
func (q *uniqueFifo[T]) PeekByKey(key int) (value T, err error) {
	q.lock.RLock()
	defer q.lock.RUnlock()

	if len(q.values) == 0 {
		err = ErrEmpty
		return
	} else if key >= len(q.values) {
		err = ErrKeyOutOfRange
		return
	}
	value = q.values[key]
	return
}

// PeekAll returns all the values in the queue without removing them.
//
// If the queue is empty, an error is returned.
//
// Example:
//
//	q.Enqueue(1)
//	q.Enqueue(2)
//	q.Enqueue(3)
//	q.PeekAll() // returns [1, 2, 3]
func (q *uniqueFifo[T]) PeekAll() (value []T, err error) {
	q.lock.RLock()
	defer q.lock.RUnlock()

	if len(q.values) == 0 {
		err = ErrEmpty
		return
	}
	value = q.values
	return
}

// Len returns the number of elements in the queue.
//
// Example:
//
//	q.Enqueue(1)
//	q.Enqueue(2)
//	q.Enqueue(3)
//	q.Len() // returns 3
func (q *uniqueFifo[T]) Len() int {
	q.lock.RLock()
	defer q.lock.RUnlock()

	return len(q.values)
}
