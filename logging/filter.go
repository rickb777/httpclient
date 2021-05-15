package logging

import (
	"net/http"
	"sync"
)

// A Filter determines the logging level, possibly based on each request.
type Filter interface {
	Level(*http.Request) Level
}

// FilterFunc adapts a function to be a Filter.
type FilterFunc func(*http.Request) Level

func (f FilterFunc) Level(req *http.Request) Level {
	return f(req)
}

//-------------------------------------------------------------------------------------------------

// FixedLevel is a FilterFunc that always return a specified Level.
func FixedLevel(level Level) FilterFunc {
	return func(_ *http.Request) Level {
		return level
	}
}

//-------------------------------------------------------------------------------------------------

// VariableFilter is a Filter that is controlled by a predicate.
// This predicate can be altered by any goroutine at any time.
type VariableFilter struct {
	predicate FilterFunc
	mu        sync.RWMutex
}

// NewVariableFilter is a Filter that initially has a fixed level. However,
// its predicate can be changed later.
func NewVariableFilter(initial Level) *VariableFilter {
	return NewVariablePredicate(FixedLevel(initial))
}

// NewVariablePredicate is a Filter with a predicate that can be changed later.
func NewVariablePredicate(initial FilterFunc) *VariableFilter {
	return &VariableFilter{
		predicate: initial,
		mu:        sync.RWMutex{},
	}
}

func (vf *VariableFilter) Level(req *http.Request) Level {
	vf.mu.RLock()
	defer vf.mu.RUnlock()
	return vf.predicate(req)
}

// SetLevel allows the predicate to be changed. This can be called from any goroutine.
func (vf *VariableFilter) SetLevel(newLevel FilterFunc) {
	vf.mu.Lock()
	defer vf.mu.Unlock()
	vf.predicate = newLevel
}
