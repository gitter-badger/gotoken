package gotoken

type circularBuffer struct {
	buffer []int
	shift  int
	size   int
}

func makeCircularBuffer(size int) circularBuffer {
	var cb circularBuffer
	cb.buffer = make([]int, size)
	cb.shift = 0
	cb.size = 0
	return cb
}

func (cb *circularBuffer) full() bool {
	return cb.size == len(cb.buffer)
}

func (cb *circularBuffer) empty() bool {
	return cb.size == 0
}

func (cb *circularBuffer) pop() {
	if !cb.empty() {
		cb.size = cb.size - 1
		cb.shift = (cb.shift + 1) % len(cb.buffer)
	}
}

func (cb *circularBuffer) push(value int) {
	if cb.full() {
		cb.pop()
	}
	cb.buffer[(cb.size+cb.shift)%len(cb.buffer)] = value
	cb.size = cb.size + 1
}

func (cb *circularBuffer) extract() (int, []int) {
	left := cb.buffer[cb.shift]
	right := make([]int, cb.size-1)
	for i := 1; i < cb.size; i++ {
		right[i-1] = cb.buffer[(cb.shift+i)%len(cb.buffer)]
	}
	return left, right
}
