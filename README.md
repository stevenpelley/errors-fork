This is a fork of go-errors/errors for my personal use and where I want to
consider breaking changes for V2

Breaking changes:
- Functions return `error` instead of `*Error`.  This allows callers to use `nil` as expected.  See https://go.dev/doc/faq#nil_error
- Requires Go 1.20.  No attempt made for backwards compatibility.  The purpose is to match go-errors to standard Go error patterns.
- `(*Error) Is` no longer inspects the contained error.  Go's doc for `Is` specifically says "An Is method should only shallowly compare err and the target and not call Unwrap on either.".  The previous implementation was inconsistent with `As` -- an error could be `Is` another error, but would not then provide an error using `As`.  It in-essense unwrapped an error by accessing `(*Error).Err`, which is exactly what `Unwrap()` does.  This is also much simpler.

TODO:
- reconcile how stacks should be exposed to the caller (method names).  See https://github.com/golang/go/issues/60873 and https://github.com/golang/go/issues/63358 for a Go discussion.
- figure out use cases and helper methods for `Is` and `As`. I expect users want to do things like "give me the stack-containing `*Error` whose immediate child `Is` some type of error", which is what the previous implementation of `Is` provided in a way.  I just think this should have a different function so as to not confuse the expected golang `errors.Is` semantics.
- simplify the tests.  Consider testify for requires/assertions. Table-based tests where appropriate.
- reconsider whether this is a "drop in replacement for Go's errors".  With the above `Is` change this merely passes through calls to `Is`, `As`, `Join`, and `Unwrap`.  It has different semantics for `New` (accepts interface{} vs string).  `Wrap` and `WrapPrefix` are new functions with no equivalent in `errors`.  `Errorf` is equivalent to `fmt.Errorf` but that's not in the `errors` package.  I think it makes more sense to say this is a module "to supplement Go's existing `errors` module, additionally replacing some functions."  If this is no longer a drop-in replacement for Go errors then it should have a different name.  Something like stackfulerrors, stackerrors, etc


go-errors/errors
================

[![Build Status](https://travis-ci.org/go-errors/errors.svg?branch=master)](https://travis-ci.org/go-errors/errors)

Package errors adds stacktrace support to errors in go.

This is particularly useful when you want to understand the state of execution
when an error was returned unexpectedly.

It provides the type \*Error which implements the standard golang error
interface, so you can use this library interchangeably with code that is
expecting a normal error return.

Usage
-----

Full documentation is available on
[godoc](https://godoc.org/github.com/go-errors/errors), but here's a simple
example:

```go
package crashy

import "github.com/go-errors/errors"

var Crashed = errors.Errorf("oh dear")

func Crash() error {
    return errors.New(Crashed)
}
```

This can be called as follows:

```go
package main

import (
    "crashy"
    "fmt"
    "github.com/go-errors/errors"
)

func main() {
    err := crashy.Crash()
    if err != nil {
        if errors.Is(err, crashy.Crashed) {
            fmt.Println(err.(*errors.Error).ErrorStack())
        } else {
            panic(err)
        }
    }
}
```

Meta-fu
-------

This package was original written to allow reporting to
[Bugsnag](https://bugsnag.com/) from
[bugsnag-go](https://github.com/bugsnag/bugsnag-go), but after I found similar
packages by Facebook and Dropbox, it was moved to one canonical location so
everyone can benefit.

This package is licensed under the MIT license, see LICENSE.MIT for details.


## Changelog
* v1.1.0 updated to use go1.13's standard-library errors.Is method instead of == in errors.Is
* v1.2.0 added `errors.As` from the standard library.
* v1.3.0 *BREAKING* updated error methods to return `error` instead of `*Error`.
>  Code that needs access to the underlying `*Error` can use the new errors.AsError(e)
> ```
>   // before
>   errors.New(err).ErrorStack()
>   // after
>.  errors.AsError(errors.Wrap(err)).ErrorStack()
> ```
* v1.4.0 *BREAKING* v1.4.0 reverted all changes from v1.3.0 and is identical to v1.2.0
* v1.4.1 no code change, but now without an unnecessary cover.out file.
* v1.4.2 performance improvement to ErrorStack() to avoid unnecessary work https://github.com/go-errors/errors/pull/40
* v1.5.0 add errors.Join() and errors.Unwrap() copying the stdlib https://github.com/go-errors/errors/pull/40
* v1.5.1 fix build on go1.13..go1.19 (broken by adding Join and Unwrap with wrong build constraints)
