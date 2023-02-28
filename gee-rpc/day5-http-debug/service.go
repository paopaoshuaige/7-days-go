package geerpc

import (
	"go/ast"
	"log"
	"reflect"
	"sync/atomic"
)

// methodType 通过反射获取方法
type methodType struct {
	method    reflect.Method // 方法
	ArgType   reflect.Type   // 参数类型
	ReplyType reflect.Type   // 第二个参数类型
	numCalls  uint64         // 统计方法调用次数
}

// NumCalls 返回加载到 numCalls 的值
func (m *methodType) NumCalls() uint64 {
	return atomic.LoadUint64(&m.numCalls)
}

// newArgv 初始化 Argv 参数，如果是指针类型就新建实例，否则就新建类型之后再建实例。
func (m *methodType) newArgv() reflect.Value {
	var argv reflect.Value // argv可以是指针或者值类型
	if m.ArgType.Kind() == reflect.Ptr {
		argv = reflect.New(m.ArgType.Elem())
	} else {
		argv = reflect.New(m.ArgType).Elem()
	}
	return argv
}

// newReplyv 回复必须是指针类型
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

// service 方法服务
type service struct {
	name   string                 // 名字
	typ    reflect.Type           // 类型
	rcvr   reflect.Value          // 实例
	method map[string]*methodType // 存储所有符合条件的方法
}

// newService 解析当前方法
func newService(rcvr interface{}) *service {
	s := new(service)
	s.rcvr = reflect.ValueOf(rcvr)                  // 获取实例
	s.name = reflect.Indirect(s.rcvr).Type().Name() // 获取名字
	s.typ = reflect.TypeOf(rcvr)                    // 获取类型
	if !ast.IsExported(s.name) {
		log.Fatalf("rpc server: %s is not a valid service name", s.name)
	}
	s.registerMethods()
	return s
}

// isExportedOrBuiltinType 判断入参是否正确
func isExportedOrBuiltinType(t reflect.Type) bool {
	return ast.IsExported(t.Name()) || t.PkgPath() == ""
}

// registerMethods 注册方法
func (s *service) registerMethods() {
	s.method = make(map[string]*methodType)
	for i := 0; i < s.typ.NumMethod(); i++ {
		method := s.typ.Method(i)
		mType := method.Type
		// 如果不是两个导出或内置类型的入参，或者返回值不止一个（反射时候算三个，0是自身）
		if mType.NumIn() != 3 || mType.NumOut() != 1 {
			continue
		}
		// 如果mtype自身和反射获取到的不一样
		if mType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}
		// 两个入参
		argType, replyType := mType.In(1), mType.In(2)
		if !isExportedOrBuiltinType(argType) || !isExportedOrBuiltinType(replyType) {
			continue
		}
		s.method[method.Name] = &methodType{
			method:    method,
			ArgType:   argType,
			ReplyType: replyType,
		}
		log.Printf("rpc server: register %s.%s\n", s.name, method.Name)
	}
}

// call 通过反射值调用方法
func (s *service) call(m *methodType, argv, replyv reflect.Value) error {
	atomic.AddUint64(&m.numCalls, 1)
	f := m.method.Func
	returnValues := f.Call([]reflect.Value{s.rcvr, argv, replyv})
	if errInter := returnValues[0].Interface(); errInter != nil {
		return errInter.(error)
	}
	return nil
}
