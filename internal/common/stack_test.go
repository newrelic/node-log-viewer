package common

import (
	"reflect"
	"testing"

	"github.com/gookit/goutil/testutil/assert"
)

func Test_Stack(t *testing.T) {
	stack := NewStack[int]()
	assert.IsKind(t, reflect.Struct, stack)
	assert.Equal(t, 0, stack.Size())

	stack.Push(1, 2, 3)
	assert.Equal(t, 3, stack.Size())

	e, ok := stack.Pop()
	assert.Equal(t, 3, e)
	assert.Equal(t, true, ok)
	e, ok = stack.Pop()
	assert.Equal(t, 2, e)
	assert.Equal(t, true, ok)
	e, ok = stack.Pop()
	assert.Equal(t, 1, e)
	assert.Equal(t, true, ok)
	e, ok = stack.Pop()
	assert.Empty(t, e)
	assert.Equal(t, false, ok)
}
