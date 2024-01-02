/*
Package stackerr annotates errors with their source filename and line number.

This package differs from other error packages in two ways:

  - Errors provide the filename and line numbers that create or wrap errors.
    They do not record a stack trace.
  - It is not a replacement for the standard errors package.

The Errorf function is the same as fmt.Errorf except that it also records the
filename and line number of where it's called. When an Errorf error is formatted
with a '+' flag ("%+s", "%+v", or "%+q") it prints its source file location:

	err := Errorf("wrong password")
	fmt.Printf("error: %+v\n", err)
	// Output:
	// error: wrong password
	// /go/src/login.go:14 wrong password

If it wraps another error, it also prints source information for other Errorf
errors in its tree:

	user := "gcurtis"
	wrapped := Errorf("wrong password")
	err := Errorf("login %q: %w", user, wrapped)
	fmt.Printf("error: %+v\n", err)

	// Output:
	// error: login "gcurtis": wrong password
	// /go/src/handler.go:176 login "gcurtis": wrong password
	// /go/src/login.go:14 wrong password

Note that the output is not a stack trace. It provides the location of the call
to Errorf, not where the error returns up the stack. In some cases, seeing the
lines that build the error chain is more accurate than a stack trace, but it
also means that errors from other packages will not have location information.

When using the [log] or [log/slog] packages, consider setting the
[log.Llongfile] flag or [log/slog.HandlerOptions.AddSource] field, respectively,
instead of using this package.
*/
package stackerr

import (
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// Errorf is like fmt.Errorf, except it also records the source file and line
// number that calls Errorf. Formatting the error with the fmt '+' flag prints
// its source location. See the package documentation for details.
func Errorf(format string, a ...any) error {
	err := fmt.Errorf(format, a...)
	switch t := err.(type) {
	case interface{ Unwrap() error }:
		werr := &wrapError{
			srcError: srcError{msg: err.Error()},
			err:      t.Unwrap(),
		}
		runtime.Callers(2, werr.pc[:])
		return werr
	case interface{ Unwrap() []error }:
		werr := &wrapErrors{
			srcError: srcError{msg: err.Error()},
			errs:     t.Unwrap(),
		}
		runtime.Callers(2, werr.pc[:])
		return werr
	default:
		serr := &srcError{msg: err.Error()}
		runtime.Callers(2, serr.pc[:])
		return serr
	}
}

type srcError struct {
	msg string
	pc  [1]uintptr
}

func (e *srcError) Error() string                 { return e.msg }
func (e *srcError) Frame() runtime.Frame          { fr, _ := runtime.CallersFrames(e.pc[:]).Next(); return fr }
func (e *srcError) Format(f fmt.State, verb rune) { format(f, verb, e) }

type wrapError struct {
	srcError
	err error
}

func (e *wrapError) Unwrap() error                 { return e.err }
func (e *wrapError) Format(f fmt.State, verb rune) { format(f, verb, e) }

type wrapErrors struct {
	srcError
	errs []error
}

func (e *wrapErrors) Unwrap() []error               { return e.errs }
func (e *wrapErrors) Format(f fmt.State, verb rune) { format(f, verb, e) }

func format(f fmt.State, verb rune, err error) {
	fmt.Fprintf(f, fmt.FormatString(f, verb), err.Error())
	if f.Flag('+') {
		io.WriteString(f, "\n")
		printChain(f, err, "")
	}
}

func printChain(w io.Writer, err error, indent string) {
	printFileLine(w, err, "")
	for {
		switch uw := err.(type) {
		case interface{ Unwrap() error }:
			err = uw.Unwrap()
			if err == nil {
				return
			}
		case interface{ Unwrap() []error }:
			joined := uw.Unwrap()
			if len(joined) == 0 {
				return
			}
			width := len(strconv.Itoa(len(joined)))
			indent := "\t" + strings.Repeat(" ", width+3)
			for i, err := range uw.Unwrap() {
				if err == nil {
					continue
				}
				fmt.Fprintf(w, "\n\t[%*d] ", width, i)
				printChain(w, err, indent)
			}
			return
		default:
			return
		}
		printFileLine(w, err, "\n"+indent)
	}
}

var basePath = ""

func printFileLine(w io.Writer, err error, prefix string) {
	var fr runtime.Frame
	if err, ok := err.(interface{ Frame() runtime.Frame }); ok {
		fr = err.Frame()
	}
	if fr.Line == 0 {
		return
	}
	if basePath != "" {
		if rel, err := filepath.Rel(basePath, fr.File); err == nil {
			fr.File = rel
		}
	}
	msg := err.Error()
	if !strconv.CanBackquote(strings.ReplaceAll(msg, "`", "\"")) {
		msg = strconv.Quote(msg)
	}
	io.WriteString(w, prefix+fr.File+":"+strconv.Itoa(fr.Line)+" "+msg)
}
