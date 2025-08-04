package unique_fifo_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/reilabs/trusted-setup/online/server/contributors_manager/unique_fifo"
)

func TestFifo(t *testing.T) {
	q := unique_fifo.NewUniqueFifo[int]()

	// Initially queue is empty
	assert.Equal(t, 0, q.Len(), "initial len should be zero")

	// Enqueue some values
	key0 := q.Enqueue(10)
	key1 := q.Enqueue(20)
	key2 := q.Enqueue(30)

	// Check keys are in order
	assert.Equal(t, 0, key0)
	assert.Equal(t, 1, key1)
	assert.Equal(t, 2, key2)

	// Check length after push
	assert.Equal(t, 3, q.Len(), "len after pushing 3 elements")

	// Peek at the first element
	val, err := q.Peek()
	assert.NoError(t, err, "peek should not error")
	assert.Equal(t, 10, val, "peek should return first element")

	// Length should not change after peek
	assert.Equal(t, 3, q.Len(), "len after peek should remain unchanged")

	// PeekByKey
	val, err = q.PeekByKey(key0)
	assert.NoError(t, err)
	assert.Equal(t, 10, val)
	val, err = q.PeekByKey(key1)
	assert.NoError(t, err)
	assert.Equal(t, 20, val)
	val, err = q.PeekByKey(key2)
	assert.NoError(t, err)
	assert.Equal(t, 30, val)

	// Length should not change after peek
	assert.Equal(t, 3, q.Len(), "len after peek should remain unchanged")

	// PeekAll
	vals, err := q.PeekAll()
	assert.NoError(t, err)
	assert.Equal(t, []int{10, 20, 30}, vals)

	// Length should not change after peek
	assert.Equal(t, 3, q.Len(), "len after peek should remain unchanged")

	// Dequeue all the elements one by one and check length each time
	id, err := q.Dequeue()
	assert.NoError(t, err)
	assert.Equal(t, 10, id)
	assert.Equal(t, 2, q.Len())

	id, err = q.Dequeue()
	assert.NoError(t, err)
	assert.Equal(t, 20, id)
	assert.Equal(t, 1, q.Len())

	id, err = q.Dequeue()
	assert.NoError(t, err)
	assert.Equal(t, 30, id)
	assert.Equal(t, 0, q.Len())

	// Now queue is empty, pop should error
	_, err = q.Dequeue()
	assert.ErrorIs(t, err, unique_fifo.ErrEmpty, "pop from empty queue should error")

	// Ensure peek on empty queue errors
	_, err = q.Peek()
	assert.ErrorIs(t, err, unique_fifo.ErrEmpty, "peek on empty queue should error")

	// Ensure peekKey on empty queue errors
	_, err = q.PeekByKey(key0)
	assert.ErrorIs(t, err, unique_fifo.ErrEmpty, "peekKey on empty queue should error")
}

func TestFifoUnique(t *testing.T) {
	q := unique_fifo.NewUniqueFifo[int]()
	key0 := q.Enqueue(10)
	key1 := q.Enqueue(10)
	assert.Equal(t, 1, q.Len())
	assert.True(t, key0 == key1)
	v, err := q.Dequeue()
	assert.NoError(t, err)
	assert.Equal(t, 10, v)
	assert.Equal(t, 0, q.Len())
}

func TestFifoConcurrent(t *testing.T) {
	q := unique_fifo.NewUniqueFifo[int]()
	const numGoroutines = 10
	const numPerGoroutine = 1000

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2) // for pushers and poppers

	// Pushing goroutines
	for i := 0; i < numGoroutines; i++ {
		go func(base int) {
			defer wg.Done()
			for j := 0; j < numPerGoroutine; j++ {
				_ = q.Enqueue(base*numPerGoroutine + j)
			}
		}(i)
	}

	// Popping goroutines
	popped := make(chan int, numGoroutines*numPerGoroutine)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			count := 0
			for count < numPerGoroutine {
				val, err := q.Dequeue()
				if err == nil {
					popped <- val
					count++
				}
			}
		}()
	}

	wg.Wait()
	close(popped)

	// Check we got all expected values
	got := make(map[int]bool)
	for val := range popped {
		got[val] = true
	}

	expectedTotal := numGoroutines * numPerGoroutine
	if len(got) != expectedTotal {
		t.Fatalf("expected %d unique popped values but got %d", expectedTotal, len(got))
	}
}
