//
// 错误记录器
// errors
//
// 本设计补充并实现了系统的error接口，
// This design complements and implements the error interface of the go package.
//
// 用于发生错误时附带发生错误的原因、位置等信息以便还原现场。
// It can be used to restore the scene with the information of the cause and location of the error when it occurs.
//
// 因本错误设计含有比较大的数据量信息，因此需要注意被调用的频率，以避免影响到系统效率。
// Because this incorrect design contains a large amount of data information, we need to pay attention to the frequency of calls to avoid affecting system efficiency.
//
// 使用例子
// Example
//
//
// package main
//
// import "github.com/gwaylib/errors"
//
// func fn1(a int) error {
//    if a == 1{
//        return errors.ErrNoData.As(a)
//    }
//    return errors.New("not implements").As(a)
// }
//
// func fn2(b int) error {
//    return errors.As(fn1(b))
// }
//
// func main() {
//    err := fn2(2)
//    if err != nil {
//        // errors.ErrNoData == err not necessarily true, so use Equal instead.
//        if !errors.ErrNoData.Equal(err) {
//            panic(err)
//	      }
//
//        // Deals the same error.
//        fmt.Println(err)
//	  }
// }
//
package errors

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
)

var (
	ErrNoData = New("data not found")
)

type Error interface {
	// Return the code of make.
	Code() string

	// Implement the error interface of go package
	Error() string
	// Impelment the json marshal interface of go package.
	MarshalJSON() ([]byte, error)

	// Record the stack when call, and return a new error with new stack.
	As(arg ...interface{}) Error

	// Compare to another error
	// It should be established with err1.Code() == err2.Code().
	Equal(err error) bool
}

//
// 比较两个错误的值是否相等.
// 该比较有两个范围，
// 一个是内存是否相等，常用于同个程序中产生的错误的比较，若内存相等，则两个错误是相等的;
// 一个是值是否相等，常用于跨程序中产生的错误比较，若不是此接口的Error，则比较Error()的值是否相等；
// 若属于此Error接口, 则比较Code()的值是否相等，
func Equal(err1 error, err2 error) bool {
	return equal(err1, err2)
}

func equal(err1 error, err2 error) bool {
	// Memory compare
	if err1 == err2 {
		return true
	}
	if err1 == nil || err2 == nil {
		return false
	}

	eImpl1, eImpl2 := ParseError(err1), ParseError(err2)
	return eImpl1.Code() == eImpl2.Code()
}

type ErrData struct {
	Code *string    `json:"Code"`
	As   [][]string `json:"As"`
}

type errImpl struct {
	data *ErrData
}

//
// 创建一个本包Error接口的实例
//
func New(code string) Error {
	return &errImpl{
		&ErrData{
			Code: &code,
			As:   [][]string{{caller(2), "[init]"}},
		},
	}
}

//
// 解析一个错误文本
// 通常它是一个本包的Error()序列化数据, 该数据是一个json数据，将直接被序列化为本包的Error类型;
// 若非本包的接口的序列化结构，将被直接New一个新的Error出来
//
func Parse(src string) Error {
	if len(src) == 0 {
		return nil
	}
	return parse(src)
}

//
// 将一个标准的错误转为本包的Error接口类型
// 若该错误本已经是本包的Error类型，则直接转为本包的Error并原样返回;
// 若该错误是非本包的Error类型，则调用error.Error()进行值解析创建一个新的本包Error
//
func ParseError(src error) Error {
	if src == nil {
		return nil
	}
	if e, ok := src.(*errImpl); ok {
		return e
	}
	return parse(src.Error())
}

//
// 给一个错误构建错误定位信息
// 解析error时等价于ParseError，并在解析出的Error后构建当前置的错误定位信息。
// 若解析出的是本包的Error类型的实现，将在原实现基础上构建错误定位信息，此时也等价于Error的As方法调用。
//
// 返回增加了定位信息的新Error实现, 因为是新实例返回，因此可以支持并发操作
//
func As(err error, reason ...interface{}) Error {
	if err == nil {
		return nil
	}
	e := ParseError(err).(*errImpl)
	as := []string{caller(2)}
	if len(reason) > 0 {
		as = append(as, fmt.Sprintf("%+v", reason))
	}
	return &errImpl{
		&ErrData{
			Code: e.data.Code,
			As:   append(e.data.As, as),
		},
	}
}

func parse(src string) *errImpl {
	if len(src) == 0 {
		return nil
	}
	if src[:1] != "{" {
		return New(src).(*errImpl)
	}

	data := &ErrData{}
	if err := json.Unmarshal([]byte(src), data); err != nil {
		return New(src).(*errImpl)
	}
	return &errImpl{data}
}

// call for domain
func caller(depth int) string {
	at := ""
	pc, file, line, ok := runtime.Caller(depth)
	if !ok {
		at = "domain of caller is unknown"
	}
	me := runtime.FuncForPC(pc)
	if me == nil {
		at = "domain of call is unnamed"
	}

	fileFields := strings.Split(file, "/")
	if len(fileFields) < 1 {
		at = "domain of file is unnamed"
		return at
	}
	funcFields := strings.Split(me.Name(), "/")
	if len(fileFields) < 1 {
		at = "domain of func is unnamed"
		return at
	}

	fileName := strings.Join(fileFields[len(fileFields)-1:], "/")
	funcName := strings.Join(funcFields[len(funcFields)-1:], "/")

	at = fmt.Sprintf("%s(%s:%d)", funcName, fileName, line)
	return at
}

// Return the code of make.
func (e *errImpl) Code() string {
	return *e.data.Code
}

// Implement the error interface of go package
func (e *errImpl) Error() string {
	data, err := json.Marshal(e.data)
	if err != nil {
		return fmt.Sprintf("%+v", e.data)
	}
	return string(data)
}

// Impelment the json marshal interface of go package.
func (e *errImpl) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.data)
}

// Record the stack when call, and return a new error with new stack.
func (e *errImpl) As(reason ...interface{}) Error {
	as := []string{caller(2)}
	if len(reason) > 0 {
		as = append(as, fmt.Sprintf("%+v", reason))
	}
	return &errImpl{
		&ErrData{
			Code: e.data.Code,
			As:   append(e.data.As, as),
		},
	}
}

// Compare to another error
// It should be established with err1.Code() == err2.Code().
func (e *errImpl) Equal(l error) bool {
	return equal(e, l)
}
