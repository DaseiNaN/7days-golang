package GeeRPC

import (
	"go/ast"
	"log"
	"reflect"
	"sync/atomic"
)

type methodType struct {
	method    reflect.Method
	ArgType   reflect.Type
	ReplyType reflect.Type
	numCalls  uint64
}

func (m *methodType) NumCalls() uint64 {
	return atomic.LoadUint64(&m.numCalls)
}

func (m *methodType) newArgv() reflect.Value {
	var argv reflect.Value

	// argv 可能是指针类型, 也可能是值类型
	if m.ArgType.Kind() == reflect.Ptr {
		// 指针类型 -> 取到实例类型 -> 获得零值指针
		argv = reflect.New(m.ArgType.Elem())
	} else {
		// 值类型 -> 获得零值指针 -> 取到零值实例本体
		argv = reflect.New(m.ArgType).Elem()
	}

	return argv
}

func (m *methodType) newReplyv() reflect.Value {
	replyv := reflect.New(m.ReplyType.Elem())
	switch m.ReplyType.Elem().Kind() {
	case reflect.Map:
		replyv.Elem().Set(reflect.MakeMap(m.ReplyType.Elem()))
	case reflect.Slice:
		replyv.Elem().Set(reflect.MakeSlice(m.ReplyType.Elem(), 0, 0))
	}
	return replyv
}

type service struct {
	name     string                 // 映射的结构体的名称
	typ      reflect.Type           // 映射的结构体的类型
	receiver reflect.Value          // 结构体的实例
	method   map[string]*methodType // 结构方法映射表
}

func newService(receiver interface{}) *service {
	// receiver: 任意需要映射为服务的结构体实例
	s := new(service)

	s.receiver = reflect.ValueOf(receiver)
	s.name = reflect.Indirect(s.receiver).Type().Name()
	s.typ = reflect.TypeOf(receiver)

	if !ast.IsExported(s.name) {
		log.Fatalf("RPC Server: %s is not a valid service name", s.name)
	}
	s.registerMethods()
	return s
}

func (s *service) registerMethods() {
	s.method = make(map[string]*methodType)
	for i := 0; i < s.typ.NumMethod(); i++ {
		method := s.typ.Method(i)
		mtype := method.Type

		// 三个入参(包括实例本身), 一个出参(error)
		if mtype.NumIn() != 3 || mtype.NumOut() != 1 {
			continue
		}
		// 出参的类型必须为 error
		if mtype.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}

		// 必须为 exported 或 builtin 类型
		argType, replyType := mtype.In(1), mtype.In(2)
		if !isExportedOrBuiltinType(argType) || !isExportedOrBuiltinType(replyType) {
			continue
		}

		s.method[method.Name] = &methodType{
			method:    method,
			ArgType:   argType,
			ReplyType: replyType,
		}
		log.Printf("RPC Server: register %s.%s\n", s.name, method.Name)
	}
}

func isExportedOrBuiltinType(t reflect.Type) bool {
	return ast.IsExported(t.Name()) || t.PkgPath() == ""
}

func (s *service) call(m *methodType, argv reflect.Value, replyv reflect.Value) error {
	atomic.AddUint64(&m.numCalls, 1)
	f := m.method.Func
	returnValues := f.Call([]reflect.Value{s.receiver, argv, replyv})

	if errInter := returnValues[0].Interface(); errInter != nil {
		return errInter.(error)
	}
	return nil
}
