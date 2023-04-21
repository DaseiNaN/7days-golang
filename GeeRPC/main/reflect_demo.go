package main

import (
	"log"
	"reflect"
	"strings"
	"sync"
)

func main() {
	var wg sync.Mutex
	typ := reflect.TypeOf(&wg)

	// (reflect.Type).NumMethod(): 返回类型拥有的方法的数目
	// (reflect.Type).Method(int): 返回第 i 个方法
	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)
		// make(type, length, capacity)
		argv := make([]string, 0, method.Type.NumIn())
		returns := make([]string, 0, method.Type.NumOut())

		// 第 0 个入参为 wg 自身
		for j := 1; j < method.Type.NumIn(); j++ {
			argv = append(argv, method.Type.In(j).Name())
		}

		for j := 0; j < method.Type.NumOut(); j++ {
			returns = append(returns, method.Type.Out(j).Name())
		}

		log.Printf("func (w %s) %s(%s) %s",
			typ.Elem().Name(),
			method.Name,
			strings.Join(argv, ","),
			strings.Join(returns, ","),
		)
	}
}
