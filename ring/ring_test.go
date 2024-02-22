package ring

import (
	"fmt"
	"testing"
)

func eqSlices[T comparable](a []T, b []T) bool {
	if len(a) != len(b) {
		return false
	}

	for idx := 0; idx < len(a); idx++ {
		if a[idx] != b[idx] {
			return false
		}
	}

	return true
}

func TestToString(t *testing.T) {
	obj := NewRingQueue[int](10)
	expected := "[RRQ full:false size:10 start:0 end:0 data:[0 0 0 0 0 0 0 0 0 0]]"
	actual := fmt.Sprint(obj)

	if actual != expected {
		t.Fatalf("Mismatch, expected:%s, found:%s", expected, actual)
	}
}

func TestRingQueue_PeekIdx(t *testing.T) {
	// Create a RingQueue with capacity 5
	r := NewRingQueue[int](6)

	// Push elements into the RingQueue
	for i := 1; i <= 5; i++ {
		err := r.Push(i)
		if err != nil {
			t.Errorf("Error pushing element: %v", err)
		}
	}

	// Test cases for PeekIdx
	tests := []struct {
		index    int
		expected int
		err      bool
	}{
		{0, 1, false}, // Element at index 0 is 1
		{2, 3, false}, // Element at index 2 is 3
		{4, 5, false}, // Element at index 4 is 5
		{-1, 0, true}, // Index out of bounds
		{6, 0, true},  // Data not available at index 6 yet
	}

	for _, test := range tests {
		result, err := r.PeekIdx(test.index)

		if test.err {
			if err == nil {
				t.Errorf("Expected error for index %d, but got none", test.index)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for index %d: %v", test.index, err)
			}

			if result != test.expected {
				t.Errorf("For index %d, expected %d, but got %d", test.index, test.expected, result)
			}
		}
	}
}

func TestRingQueue_PeekIdxWithOverflow(t *testing.T) {
	// Create a RingQueue with capacity 5
	r := NewRingQueue[int](10)

	// Push elements into the RingQueue
	for i := 1; i <= 15; i++ {
		err := r.PushSafe(i)
		if err != nil {
			t.Errorf("Error pushing element: %v", err)
		}
	}

	// Test cases for PeekIdx
	tests := []struct {
		index    int
		expected int
		err      bool
	}{
		{0, 6, false},  // Element at index 0 is 6
		{5, 11, false}, // Element at index 5 is 11
		{9, 15, false}, // Element at index 9 is 15
	}

	for _, test := range tests {
		result, err := r.PeekIdx(test.index)

		if test.err {
			if err == nil {
				t.Errorf("Expected error for index %d, but got none", test.index)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for index %d: %v", test.index, err)
			}

			if result != test.expected {
				t.Errorf("For index %d, expected %d, but got %d", test.index, test.expected, result)
			}
		}
	}
}

func TestRingQueue_PeekSlice(t *testing.T) {
	// Create a RingQueue with capacity 5
	r := NewRingQueue[int](6)

	// Push elements into the RingQueue
	for i := 1; i <= 5; i++ {
		err := r.Push(i)
		if err != nil {
			t.Errorf("Error pushing element: %v", err)
		}
	}

	// Test cases for PeekIdx
	tests := []struct {
		index       int
		expectedLen int
		err         bool
	}{
		{0, 5, false}, // Get 5 elements starting at idx 0
		{2, 3, false}, // Get 3 elements starting at idx 2
		{5, 0, false}, // Get 0 elements starting at idx 5
	}

	for _, test := range tests {
		result, err := r.PeekSlice(test.index)

		if test.err {
			if err == nil {
				t.Errorf("Expected error for index %d, but got none", test.index)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for index %d: %v", test.index, err)
			}

			if len(result) != test.expectedLen {
				t.Errorf("For index %d, expected %d, but got %d", test.index, test.expectedLen, len(result))
			}
		}
	}
}

func TestRingQueue_PeekSliceWithOverflow(t *testing.T) {
	// Create a RingQueue with capacity 5
	r := NewRingQueue[int](10)

	// Push elements into the RingQueue
	for i := 1; i <= 15; i++ {
		err := r.PushSafe(i)
		if err != nil {
			t.Errorf("Error pushing element: %v", err)
		}
	}

	// Test cases for PeekIdx
	tests := []struct {
		index       int
		expectedLen int
		err         bool
	}{
		{0, 10, false}, // Get 5 elements starting at idx 0
		{5, 5, false},  // Get 5 elements starting at idx 0
		{9, 1, false},  // Get 5 elements starting at idx 0
	}

	for _, test := range tests {
		result, err := r.PeekSlice(test.index)

		if test.err {
			if err == nil {
				t.Errorf("Expected error for index %d, but got none", test.index)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for index %d: %v", test.index, err)
			}

			if len(result) != test.expectedLen {
				t.Errorf("For index %d, expected %d, but got %d", test.index, test.expectedLen, len(result))
			}
		}
	}
}

func TestPushEnough(t *testing.T) {
	obj := NewRingQueue[int](10)
	for idx := 0; idx < 10; idx++ {
		err := obj.Push(idx)
		if err != nil {
			t.Fatalf("Unexpected error in adding an element with index %d", idx)
		}
	}

	expected := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	if !eqSlices(obj.data, expected) {
		t.Fatalf("Container data mismatch, expected:%v, found:%v", expected, obj.data)
	}
}

func TestScan(t *testing.T) {
	obj := NewRingQueue[int](10)
	for idx := 0; idx < 15; idx++ {
		err := obj.PushSafe(idx)
		if err != nil {
			t.Fatalf("Unexpected error in adding an element with index %d", idx)
		}
	}

	expected := []int{5, 6, 7, 8, 9, 10, 11, 12, 13, 14}
	got := []int{}
	obj.Scan(func(elem int, idx int) bool {
		got = append(got, elem)
		return false
	})

	if !eqSlices(got, expected) {
		t.Fatalf("Container data mismatch, expected:%v, found:%v", expected, got)
	}
}

func TestPushSafe(t *testing.T) {
	obj := NewRingQueue[int](5)
	for idx := 0; idx < 10; idx++ {
		err := obj.PushSafe(idx)
		if err != nil {
			t.Fatalf("Unexpected error in adding an element with index %d", idx)
		}
	}

	expected := []int{5, 6, 7, 8, 9}

	if !eqSlices(obj.data, expected) {
		t.Fatalf("Container data mismatch, expected:%v, found:%v", expected, obj.data)
	}
}

func TestPushOver(t *testing.T) {
	obj := NewRingQueue[int](10)
	for idx := 0; idx < 10; idx++ {
		err := obj.Push(idx)
		if err != nil {
			t.Fatalf("Unexpected error in adding an element with index %d", idx)
		}
	}

	err := obj.Push(100)
	if err == nil {
		t.Fatalf("Expected overflow error")
	}

	expected := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	if !eqSlices(obj.data, expected) {
		t.Fatalf("Container data mismatch, expected:%v, found:%v", expected, obj.data)
	}
}

func TestPushPop(t *testing.T) {
	obj := NewRingQueue[int](10)
	for idx := 0; idx < 8; idx++ {
		obj.Push(idx)
	}
	for idx := 0; idx < 5; idx++ {
		e, err := obj.Pop()
		if err != nil || e != idx {
			t.Fatalf("inconsistent behavior")
		}
	}
	for idx := 0; idx < 7; idx++ {
		obj.Push(100 + idx)
	}

	expected := []int{102, 103, 104, 105, 106, 5, 6, 7, 100, 101}

	if !eqSlices(obj.data, expected) {
		t.Fatalf("Container data mismatch, expected:%v, found:%v", expected, obj.data)
	}

	if obj.Size() != 10 {
		t.Fatalf("inconsistent size: %d", obj.Size())
	}

	for idx := 0; idx < 10; idx++ {
		e, _ := obj.Pop()
		if e != expected[(5+idx)%10] {
			t.Fatalf("inconsistent behavior")
		}
	}
}
