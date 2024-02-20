// Package errors provides errors that have stack-traces.
//
// This is particularly useful when you want to understand the
// state of execution when an error was returned unexpectedly.
//
// It provides the type *Error which implements the standard
// golang error interface, so you can use this library interchangably
// with code that is expecting a normal error return.
//
// For example:
//
//	package crashy
//
//	import "github.com/go-errors/errors"
//
//	var Crashed = errors.Errorf("oh dear")
//
//	func Crash() error {
//	    return errors.New(Crashed)
//	}
//
// This can be called as follows:
//
//	package main
//
//	import (
//	    "crashy"
//	    "fmt"
//	    "github.com/go-errors/errors"
//	)
//
//	func main() {
//	    err := crashy.Crash()
//	    if err != nil {
//	        if errors.Is(err, crashy.Crashed) {
//	            fmt.Println(err.(*errors.Error).ErrorStack())
//	        } else {
//	            panic(err)
//	        }
//	    }
//	}
//
// This package was original written to allow reporting to Bugsnag,
// but after I found similar packages by Facebook and Dropbox, it
// was moved to one canonical location so everyone can benefit.
package errors

import (
	"bytes"
	baseErrors "errors"
	"fmt"
	"reflect"
	"runtime"
)

// The maximum number of stackframes on any error.
var MaxStackDepth = 50

// Error is an error with an attached stacktrace. It can be used
// wherever the builtin error interface is expected.
type Error struct {
	Err    error
	stack  []uintptr
	frames []StackFrame
	prefix string
}

// New makes an Error from the given value. If that value is already an
// error then it will be used directly, if not, it will be passed to
// fmt.Errorf("%v"). The stacktrace will point to the line of code that
// called New.
func New(e interface{}) error {
	var err error

	switch e := e.(type) {
	case error:
		err = e
	default:
		err = fmt.Errorf("%v", e)
	}

	stack := make([]uintptr, MaxStackDepth)
	length := runtime.Callers(2, stack[:])
	return &Error{
		Err:   err,
		stack: stack[:length],
	}
}

// Wrap makes an Error from the given value. If that value is already an *Error
// it will not be wrapped and instead will be returned without modification. If
// that value is already an error then it will be used directly and wrapped.
// Otherwise, the value will be passed to fmt.Errorf("%v") and then wrapped. To
// explicitly wrap an *Error with a new stacktrace use Errorf. The skip
// parameter indicates how far up the stack to start the stacktrace. 0 is from
// the current call, 1 from its caller, etc.
func Wrap(e interface{}, skip int) error {
	if e == nil {
		return nil
	}

	return wrap(e, skip)
}

// internal wrap returning *Error. This should never return nil
func wrap(e interface{}, skip int) *Error {
	var err error

	switch e := e.(type) {
	case *Error:
		return e
	case error:
		err = e
	default:
		err = fmt.Errorf("%v", e)
	}

	stack := make([]uintptr, MaxStackDepth)
	length := runtime.Callers(3+skip, stack[:])
	return &Error{
		Err:   err,
		stack: stack[:length],
	}
}

// WrapPrefix makes an Error from the given value. If that value is already an
// *Error it will not be wrapped and instead will be returned without
// modification. If that value is already an error then it will be used
// directly and wrapped.  Otherwise, the value will be passed to
// fmt.Errorf("%v") and then wrapped. To explicitly wrap an *Error with a new
// stacktrace use Errorf. The prefix parameter is used to add a prefix to the
// error message when calling Error(). The skip parameter indicates how far up
// the stack to start the stacktrace. 0 is from the current call, 1 from its
// caller, etc.
func WrapPrefix(e interface{}, prefix string, skip int) error {
	if e == nil {
		return nil
	}

	err := wrap(e, skip)

	if err.prefix != "" {
		prefix = fmt.Sprintf("%s: %s", prefix, err.prefix)
	}

	return &Error{
		Err:    err.Err,
		stack:  err.stack,
		prefix: prefix,
	}

}

// Errorf creates a new error with the given message. You can use it
// as a drop-in replacement for fmt.Errorf() to provide descriptive
// errors in return values.
func Errorf(format string, a ...interface{}) error {
	return Wrap(fmt.Errorf(format, a...), 1)
}

// Error returns the underlying error's message.
func (err *Error) Error() string {

	msg := err.Err.Error()
	if err.prefix != "" {
		msg = fmt.Sprintf("%s: %s", err.prefix, msg)
	}

	return msg
}

// Stack returns the callstack formatted the same way that go does
// in runtime/debug.Stack()
func (err *Error) Stack() []byte {
	buf := bytes.Buffer{}

	for _, frame := range err.StackFrames() {
		buf.WriteString(frame.String())
	}

	return buf.Bytes()
}

// Callers satisfies the bugsnag ErrorWithCallerS() interface
// so that the stack can be read out.
func (err *Error) Callers() []uintptr {
	return err.stack
}

// ErrorStack returns a string that contains both the
// error message and the callstack.
func (err *Error) ErrorStack() string {
	return err.TypeName() + " " + err.Error() + "\n" + string(err.Stack())
}

// StackFrames returns an array of frames containing information about the
// stack.
func (err *Error) StackFrames() []StackFrame {
	if err.frames == nil {
		err.frames = make([]StackFrame, len(err.stack))

		for i, pc := range err.stack {
			err.frames[i] = NewStackFrame(pc)
		}
	}

	return err.frames
}

// TypeName returns the type this error. e.g. *errors.stringError.
func (err *Error) TypeName() string {
	if _, ok := err.Err.(uncaughtPanic); ok {
		return "panic"
	}
	return reflect.TypeOf(err.Err).String()
}

// Return the wrapped error (implements api for As function).
func (err *Error) Unwrap() error {
	return err.Err
}

// As finds the first error in err's tree that matches target, and if one is found, sets
// target to that error value and returns true. Otherwise, it returns false.
//
// For more information see stdlib errors.As.
func As(err error, target interface{}) bool {
	return baseErrors.As(err, target)
}

// Is detects whether the error is equal to a given error. Errors
// are considered equal by this function if they are matched by errors.Is
// or if their contained errors are matched through errors.Is.
func Is(e error, original error) bool {
	return baseErrors.Is(e, original)
}

// Join returns an error that wraps the given errors.
// Any nil error values are discarded.
// Join returns nil if every value in errs is nil.
// The error formats as the concatenation of the strings obtained
// by calling the Error method of each element of errs, with a newline
// between each string.
//
// A non-nil error returned by Join implements the Unwrap() []error method.
//
// For more information see stdlib errors.Join.
func Join(errs ...error) error {
	return baseErrors.Join(errs...)
}

// Unwrap returns the result of calling the Unwrap method on err, if err's
// type contains an Unwrap method returning error.
// Otherwise, Unwrap returns nil.
//
// Unwrap only calls a method of the form "Unwrap() error".
// In particular Unwrap does not unwrap errors returned by [Join].
//
// For more information see stdlib errors.Unwrap.
func Unwrap(err error) error {
	return baseErrors.Unwrap(err)
}
