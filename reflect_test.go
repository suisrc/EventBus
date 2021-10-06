package EventBus_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/suisrc/EventBus"
)

type TA struct {
}

type TB struct {
	TA
}

type TC interface {
}

func Test001(t *testing.T) {
	b := &TB{}
	fmt.Println(reflect.TypeOf(b).AssignableTo(reflect.TypeOf(t01).In(0)))
	fmt.Println(reflect.TypeOf(b).AssignableTo(reflect.TypeOf(t02).In(0)))
	fmt.Println(reflect.TypeOf(*b).AssignableTo(reflect.TypeOf(t03).In(0)))
	fmt.Println(reflect.TypeOf(*b).AssignableTo(reflect.TypeOf(t04).In(0)))

	fmt.Println(reflect.TypeOf(b).ConvertibleTo(reflect.TypeOf(t01).In(0)))
	fmt.Println(reflect.TypeOf(b).ConvertibleTo(reflect.TypeOf(t02).In(0)))
	fmt.Println(reflect.TypeOf(*b).ConvertibleTo(reflect.TypeOf(t03).In(0)))
	fmt.Println(reflect.TypeOf(*b).ConvertibleTo(reflect.TypeOf(t04).In(0)))

	fmt.Println(reflect.TypeOf(b).ConvertibleTo(reflect.TypeOf(t05).In(0)))

	t.Fail()
}

func Test002(t *testing.T) {
	b := &TB{}
	fmt.Println(reflect.TypeOf(t05).IsVariadic())
	fmt.Println(reflect.TypeOf(b).ConvertibleTo(reflect.TypeOf(t05).In(1).Elem()))
	fmt.Println(reflect.TypeOf(t05))

	t.Fail()
}

func Test003(t *testing.T) {
	b := &TB{}
	_, ok := new(EventBus.EventBus).PassedArguments(reflect.TypeOf(t05), b, nil, b)

	fmt.Println(ok)
	t.Fail()

	t05(b)
}

func t01(tc TC)                    {}
func t02(tc *TB)                   {}
func t03(tc TB)                    {}
func t04(tc TA)                    {}
func t05(t1 TC, t2 ...interface{}) {}
