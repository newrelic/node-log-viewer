package common

type Stack[V any] struct {
	elements []V
}

func NewStack[V any]() Stack[V] {
	return Stack[V]{
		elements: make([]V, 0),
	}
}

func (s *Stack[V]) Pop() (V, bool) {
	if len(s.elements) == 0 {
		var result V
		return result, false
	}
	result := s.elements[len(s.elements)-1]
	s.elements = s.elements[:len(s.elements)-1]
	return result, true
}

func (s *Stack[V]) Push(e ...V) {
	s.elements = append(s.elements, e...)
}

func (s *Stack[V]) Size() int {
	return len(s.elements)
}
