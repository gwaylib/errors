// Error recorder with As
//
// recoding one code, and recording the As caller static with caller, argument information for every As func is called.
//
// the data static format like this:
// ["error code", ["runtime stack of New"], ["runtime stack of As", "args of As"...], ["runtime statick of As"]...]
// the first one is error code, the second is New, the others are func As been called.
//
// # Example
//
// package main
//
// import "github.com/gwaylib/errors"
//
//	func fn1(a int) error {
//	 if a == 1 {
//	     return errors.ErrNoData.As(a)
//	 }
//	 err := errors.New("not implements") // make a error code and record the first stack of caller runtime.
//	 return err.As(a) // make the second stack of caller runtime
//	}
//
//	func fn2(b int) error {
//	  return errors.As(fn1(b)) // make the third stack of caller runtime
//	}
//
//	func main() {
//	  err := fn2(2)
//	  if err != nil {
//	      // errors.ErrNoData == err not necessarily true, so use Equal instead.
//	      if !errors.ErrNoData.Equal(err) {
//	          panic(err)
//	      }
//
//	      fmt.Println(err)
//	  }
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
	// Copy the as stack data for output
	Stack() []interface{}

	// Compare to another error
	// It should be established with err1.Code() == err2.Code().
	Equal(err error) bool
}

// Compare two error are same instances or code are matched.
func Equal(err1 error, err2 error) bool {
	return equal(err1, err2)
}

// Alias name of Equal func, compatible with official errors.Is
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

	// checking the standard package errors
	if errors.Is(err1, err2) {
		return true
	}

	// parse the error and compare the code, net transfer the error would be serial by Error() function.
	eImpl1, eImpl2 := ParseError(err1), ParseError(err2)
	return eImpl1.Code() == eImpl2.Code()
}

// ["error code", ["where stack of first caller ", "As args"...], ["where stack of second caller ", "As args"...]...]
type ErrData []interface{}

type errImpl struct {
	data ErrData // not export the data to keep it read only.
}

// Make a new error with Error type.
func New(code string, args ...interface{}) Error {
	stack := make([]interface{}, len(args)+1)
	stack[0] = caller(2)
	copy(stack[1:], args)
	return &errImpl{[]interface{}{code, stack}}
}

// Parse error from serial string, if it's ErrData format, create an Error of this package defined.
// if src is empty, return a nil Error
func Parse(src string) Error {
	if len(src) == 0 {
		return nil
	}
	return parse(src)
}

// Parse Error from a error instance.
// If the error is the type of interface Error, directly convert to the Error interface of this package.
// Call Parse(err.Error()) in others.
func ParseError(err error) Error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*errImpl); ok {
		return e
	}
	return parse(err.Error())
}

func as(depth int, err error, args ...interface{}) Error {
	if err == nil {
		return nil
	}
	e := ParseError(err).(*errImpl)
	stack := make([]interface{}, len(args)+1)
	stack[0] = caller(depth)
	copy(stack[1:], args)
	data := make([]interface{}, len(e.data)+1)

	copy(data, e.data)
	data[len(data)-1] = stack
	return &errImpl{data: data}
}

// Record a stack of runtime caller and the reason with as.
// return a new error pointer after called.
// return nil if err is nil
func As(err error, args ...interface{}) Error {
	return as(3, err, args...)
}

// Alias name of 'As'
func Wrap(err error, args ...interface{}) Error {
	return as(3, err, args...)
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
	return &errImpl{data: data}
}

// call for domain
func caller(depth int) string {
	at := ""
	pc, file, line, ok := runtime.Caller(depth)
	if !ok {
		at = "caller is false"
	}
	me := runtime.FuncForPC(pc)
	if me == nil {
		at = "pc of caller is not set"
	}

	fileFields := strings.Split(file, "/")
	if len(fileFields) < 1 {
		at = "file of caller is not named"
		return at
	}
	funcFields := strings.Split(me.Name(), "/")
	if len(funcFields) < 1 {
		at = "func of caller is not named"
		return at
	}

	fileName := strings.Join(fileFields[len(fileFields)-1:], "/")
	funcName := strings.Join(funcFields[len(funcFields)-1:], "/")
	return fmt.Sprintf("%s:%d#%s", fileName, line, funcName)
}

// Return the code of New or Parse.
func (e *errImpl) Code() string {
	return e.data[0].(string)
}

// Copy and return the stack array
func (e *errImpl) Stack() []interface{} {
	stack := make([]interface{}, len(e.data)-1)
	copy(stack, e.data[1:])
	return stack
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

// Record caller stack and return a new error interface.
func (e *errImpl) As(args ...interface{}) Error {
	return as(3, e, args...)
}

// Compare to another error
func (e *errImpl) Equal(l error) bool {
	return equal(e, l)
}
