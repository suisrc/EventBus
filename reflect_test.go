package EventBus_test

import (
	"fmt"
	"reflect"
	"testing"
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

	t.Fail()
}

func t01(tc TC)  {}
func t02(tc *TB) {}
func t03(tc TB)  {}
func t04(tc TA)  {}
