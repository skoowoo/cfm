package cfm

import (
	"container/list"
)

type stack struct {
	l *list.List
}

func newStack() (s *stack) {
	s = new(stack)
	s.l = list.New()
	return
}

func (s *stack) push(ctx *Context) {
	s.l.PushBack(ctx)
}

func (s *stack) pop() (ctx *Context) {
	var ok bool

	if e := s.l.Back(); e != nil {
		v := s.l.Remove(e)
		if ctx, ok = v.(*Context); ok {
			return
		}
	}

	return nil
}

func (s *stack) top() (ctx *Context) {
	var ok bool

	if e := s.l.Back(); e != nil {
		if ctx, ok = e.Value.(*Context); ok {
			return
		}
		ctx = nil
	}
	return
}

func skip(c byte) bool {
	return isBlank(c)
}

func isBlank(c byte) bool {
	if c == ' ' || c == '\t' || c == '\r' || c == '\n' {
		return true
	}
	return false
}

// delete one '"' of string's left and right
func trimString(s string) string {
	if len(s) != 0 && s[0] == '"' {
		s = s[1:]
	}

	if l := len(s); l != 0 && s[l-1] == '"' {
		s = s[:l-1]
	}

	return s
}
