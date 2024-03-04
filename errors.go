// 错误记录器
// error recorder
//
// 本设计补充并实现了系统的error接口。
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
// package main
//
// import "github.com/gwaylib/errors"
//
//	func fn1(a int) error {
//	   if a == 1{
//	       return errors.ErrNoData.As(a)
//	   }
//	   return errors.New("not implements").As(a)
//	}
//
//	func fn2(b int) error {
//	   return errors.As(fn1(b))
//	}
//
//	func main() {
//	   err := fn2(2)
//	   if err != nil {
//	       // errors.ErrNoData == err not necessarily true, so use Equal instead.
//	       if !errors.ErrNoData.Equal(err) {
//	           panic(err)
//	       }
//
//	       // Deals the same error.
//	       fmt.Println(err)
//	   }
//	}
package errors

import (
	"encoding/json"
	"errors"
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

// Compare two error is same instance or code is match.
func Equal(err1 error, err2 error) bool {
	return equal(err1, err2)
}

// Compatible with official errors.Is
func Is(err1 error, err2 error) bool {
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
	if errors.Is(err1, err2) {
		return true
	}

	eImpl1, eImpl2 := ParseError(err1), ParseError(err2)
	return eImpl1.Code() == eImpl2.Code()
}

type ErrData []interface{} // [0] is code, [1][0] is where, [1][1] is reason args

type errImpl struct {
	data ErrData // not export the data to keep is only readed
}

// Make a new error with Error type.
func New(code string) Error {
	return &errImpl{append(ErrData{code}, ErrData{caller(2)})}
}

// Parse from a Error serial.
func Parse(src string) Error {
	if len(src) == 0 {
		return nil
	}
	return parse(src)
}

// Parse Error from a error instance.
// If the error is type of Error, it will be return directly.
func ParseError(src error) Error {
	if src == nil {
		return nil
	}
	if e, ok := src.(*errImpl); ok {
		return e
	}
	return parse(src.Error())
}

// Record the reason with as, and return a new error with new stack of reason.
// It would be safe for concurrency.
func As(err error, reason ...interface{}) Error {
	if err == nil {
		return nil
	}
	e := ParseError(err).(*errImpl)

	// this code is same as e.As(reason...), but the caller(2) need call at here.
	as := ErrData{caller(2)}
	if len(reason) > 0 {
		as = append(as, reason...)
	}
	return &errImpl{
		append(e.data, as),
	}
}

// Same as 'As', just implement the errors system package
func Wrap(err error, arg ...interface{}) Error {
	return As(err, arg...)
}

func parse(src string) *errImpl {
	if len(src) == 0 {
		return nil
	}
	if src[0] != '[' {
		return New(src).(*errImpl)
	}

	data := ErrData{}
	if err := json.Unmarshal([]byte(src), &data); err != nil {
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
	return fmt.Sprintf("%s:%d#%s", fileName, line, funcName)
}

// Return the code of make.
func (e *errImpl) Code() string {
	return e.data[0].(string)
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
	as := ErrData{caller(2)}
	if len(reason) > 0 {
		as = append(as, reason...)
	}
	return &errImpl{
		append(e.data, as),
	}
}

// Compare to another error
// It should be established with err1.Code() == err2.Code().
func (e *errImpl) Equal(l error) bool {
	return equal(e, l)
}
