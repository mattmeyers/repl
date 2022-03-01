package repl

// History holds a list of previous commands in a ring buffer structure. If the
// capacity of the buffer is reached, then the next item appended will wrap around
// and overwrite the first element in the buffer.
type History struct {
	Values []string
	Head   int
	Cap    int
}

// NewHistory creates a new History object with the provided size. If the size is
// negatvie, then this function will panic.
func NewHistory(size int) *History {
	if size < 0 {
		panic("history cannot have negative size")
	}

	return &History{
		Values: make([]string, size),
		Head:   size,
		Cap:    size - 1,
	}
}

// Get retrieves the element at the provided offset from the head of the buffer. If
// the offset is set to 0, then the head (i.e. the most recently appended element)
// will be returned. Any postive offset will move back through the buffer, wrapping
// around to the end if necessary.
func (h *History) Get(offset uint) string {
	return h.Values[((h.Head-int(offset))%h.Cap+h.Cap)%h.Cap]
}

// Append adds a new element to the end of the history buffer. If the head pointer is
// at the end of the buffer, then the pointer will wrap around to the beginning and
// overwrite the first element.
func (h *History) Append(s string) {
	h.Head = (h.Head + 1) % h.Cap
	h.Values[h.Head] = s
}
