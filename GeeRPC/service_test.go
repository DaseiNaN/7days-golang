package GeeRPC

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"reflect"
	"testing"
)

type Foo int

type Args struct {
	Num1 int
	Num2 int
}

func (f Foo) Sum(args Args, replyv *int) error {
	*replyv = args.Num1 + args.Num2
	return nil
}

func (f Foo) sum(args Args, replyv *int) error {
	*replyv = args.Num1 + args.Num2
	return nil
}

func _assert(condition bool, msg string, v ...interface{}) {
	if !condition {
		panic(fmt.Sprintf("Assertion failed:"+msg, v))
	}
}

func TestNewService(t *testing.T) {
	var foo Foo
	s := newService(&foo)
	Convey("New Service 测试", t, func() {
		So(len(s.method), ShouldEqual, 1)
		So(s.method["Sum"], ShouldNotBeNil)
	})
}

func TestMethodType_Call(t *testing.T) {
	var foo Foo
	s := newService(&foo)
	mType := s.method["Sum"]

	Convey("Call 测试", t, func() {
		argv := mType.newArgv()
		replyv := mType.newReplyv()

		argv.Set(reflect.ValueOf(Args{Num1: 1, Num2: 3}))
		err := s.call(mType, argv, replyv)
		So(err, ShouldBeNil)
		So(mType.NumCalls(), ShouldEqual, 1)
		So(*replyv.Interface().(*int), ShouldEqual, 4)
	})
}
