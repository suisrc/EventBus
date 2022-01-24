package EventBus

import (
	"sort"
	"strings"
)

// BusKind 账户类型
type Kind int

const (
	BusSync      Kind = iota // value -> 0
	BusAsync                 // value -> 1
	BusOnceSync              // value -> 2
	BusOnceAsync             // value -> 3
)

type EventSubscriber interface {
	Subscribe(Bus) (func(), error)
}

type EventHandler interface {
	Subscribe() (kind Kind, topic string, handler interface{})
}

func SearchStrings(strs *[]string, str string) int {
	if strs == nil || len(*strs) == 0 {
		return -1
	}
	if idx := sort.SearchStrings(*strs, str); idx == len(*strs) || (*strs)[idx] != str {
		return -1
	} else {
		return idx
	}
}

func NewMultiError(errs *[]error) *MultiError {
	return &MultiError{
		Errs: *errs,
	}
}

type MultiError struct {
	Errs []error
}

func (e *MultiError) Error() string {
	sbr := strings.Builder{}
	for _, v := range e.Errs {
		sbr.WriteString(v.Error())
		sbr.WriteRune(';')
	}
	return sbr.String()
}
