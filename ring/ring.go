package ring

import (
	"fmt"
)

type RingQueue[T any] struct {
	data   []T  // container data of a generic type T
	isFull bool // disambiguate whether the queue is full or empty
	start  int  // start index (inclusive, i.e. first element)
	end    int  // end index (exclusive, i.e. next after last element)
}

func NewRingQueue[T any](capacity int64) *RingQueue[T] {
	return &RingQueue[T]{
		data:   make([]T, capacity),
		isFull: false,
		start:  0,
		end:    0,
	}
}

func (r *RingQueue[T]) String() string {
	return fmt.Sprintf(
		"[RRQ full:%v size:%d start:%d end:%d data:%v]",
		r.isFull,
		len(r.data),
		r.start,
		r.end,
		r.data)
}

func (r *RingQueue[T]) Push(elem T) error {
	if r.isFull {
		return fmt.Errorf("out of bounds push, container is full")
	}

	r.data[r.end] = elem              // place the new element on the available space
	r.end = (r.end + 1) % len(r.data) // move the end forward by modulo of capacity
	r.isFull = r.end == r.start       // check if we're full now

	return nil
}

func (r *RingQueue[T]) Pop() (T, error) {
	var res T // "zero" element (respective of the type)
	if !r.isFull && r.start == r.end {
		return res, fmt.Errorf("empty queue")
	}

	res = r.data[r.start]                 // copy over the first element in the queue
	r.start = (r.start + 1) % len(r.data) // move the start of the queue
	r.isFull = false                      // since we're removing elements, we can never be full

	return res, nil
}

func (r *RingQueue[T]) PushSafe(elem T) error {
	if r.isFull {
		r.Pop()
	}

	return r.Push(elem)
}

func (r *RingQueue[T]) Peek() (T, error) {
	var res T // "zero" element (respective of the type)
	if !r.isFull && r.start == r.end {
		return res, fmt.Errorf("empty queue")
	}

	return r.data[r.start], nil
}

func (r *RingQueue[T]) PeekIdx(idx int) (T, error) {
	var res T
	if idx < 0 || idx >= cap(r.data) {
		return res, fmt.Errorf("index out of bounds")
	}

	index := (r.start + idx) % len(r.data)
	if index >= r.end && index < r.start {
		return res, fmt.Errorf("data not available at index %d yet", idx)
	}

	return r.data[index], nil
}

func (r *RingQueue[T]) PeekSlice(startIdx int) ([]T, error) {
	if startIdx < 0 || startIdx >= cap(r.data) {
		return nil, fmt.Errorf("start index out of bounds")
	}

	sliceSize := r.Size() - startIdx
	slice := make([]T, sliceSize)

	for i := 0; i < sliceSize; i++ {
		slice[i], _ = r.PeekIdx(startIdx + i)
	}

	return slice, nil
}

func (r *RingQueue[T]) Scan(fn func(elem T, idx int) bool) {
	for i := 0; i < r.Size(); i++ {
		index := (r.start + i) % len(r.data)
		stop := fn(r.data[index], i)
		if stop {
			break
		}
	}
}

func (r *RingQueue[T]) Size() int {
	res := r.end - r.start
	if res < 0 || (res == 0 && r.isFull) {
		res = len(r.data) - res
	}

	return res
}

func (r *RingQueue[T]) IsFull() bool {
	return r.isFull
}
