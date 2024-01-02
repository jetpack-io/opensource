package stackerr

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
)

// Set the base path to the current working directory so that file paths in
// the output are relative and predictable.
func init() { basePath, _ = os.Getwd() }

func TestNoWrap(t *testing.T) {
	t.Run("%v", func(t *testing.T) {
		msg := "test"
		got := fmt.Sprintf("%v", Errorf("test"))
		if got != msg {
			t.Errorf("got %q, want %q", got, msg)
		}
	})
	t.Run("%+v", func(t *testing.T) {
		got := fmt.Sprintf("%+s", Errorf("test error"))
		match, err := regexp.MatchString(strings.TrimSpace(`
test error
stackerr_test.go:\d+ test error
`), got)
		if err != nil {
			t.Fatal(err)
		}
		if !match {
			t.Errorf("regexp doesn't match output:\n%s", got)
		}
	})
	t.Run("%s", func(t *testing.T) {
		msg := "test"
		got := fmt.Sprintf("%s", Errorf("test"))
		if got != msg {
			t.Errorf("got %q, want %q", got, msg)
		}
	})
	t.Run("%+s", func(t *testing.T) {
		got := fmt.Sprintf("%+s", Errorf("test error"))
		match, err := regexp.MatchString(strings.TrimSpace(`
test error
stackerr_test.go:\d+ test error
`), got)
		if err != nil {
			t.Fatal(err)
		}
		if !match {
			t.Errorf("regexp doesn't match output:\n%s", got)
		}
	})
	t.Run("%q", func(t *testing.T) {
		got := fmt.Sprintf("%q", Errorf("test error"))
		want := `"test error"`
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
	t.Run("%+q", func(t *testing.T) {
		got := fmt.Sprintf("%+q", Errorf("test error"))
		match, err := regexp.MatchString(strings.TrimSpace(`
"test error"
stackerr_test.go:\d+ test error
`), got)
		if err != nil {
			t.Fatal(err)
		}
		if !match {
			t.Errorf("regexp doesn't match output:\n%s", got)
		}
	})
}

func TestWrapped(t *testing.T) {
	t.Run("%v", func(t *testing.T) {
		wrapped := Errorf("wrapped")
		got := fmt.Sprintf("%v", Errorf("test error: %w", wrapped))
		want := "test error: wrapped"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
	t.Run("%+v", func(t *testing.T) {
		wrapped := Errorf("wrapped")
		got := fmt.Sprintf("%+v", Errorf("test error: %w", wrapped))
		match, err := regexp.MatchString(strings.TrimSpace(`
test error: wrapped
stackerr_test.go:\d+ test error: wrapped
stackerr_test.go:\d+ wrapped
`), got)
		if err != nil {
			t.Fatal(err)
		}
		if !match {
			t.Errorf("regexp doesn't match output:\n%s", got)
		}
	})
	t.Run("%v", func(t *testing.T) {
		wrapped := Errorf("wrapped")
		got := fmt.Sprintf("%v", Errorf("test error: %w", wrapped))
		want := "test error: wrapped"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
	t.Run("%+v", func(t *testing.T) {
		wrapped := Errorf("wrapped")
		got := fmt.Sprintf("%+v", Errorf("test error: %w", wrapped))
		match, err := regexp.MatchString(strings.TrimSpace(`
test error: wrapped
stackerr_test.go:\d+ test error: wrapped
stackerr_test.go:\d+ wrapped
`), got)
		if err != nil {
			t.Fatal(err)
		}
		if !match {
			t.Errorf("regexp doesn't match output:\n%s", got)
		}
	})
	t.Run("%q", func(t *testing.T) {
		wrapped := Errorf("wrapped")
		got := fmt.Sprintf("%q", Errorf("test error: %w", wrapped))
		want := `"test error: wrapped"`
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
	t.Run("%+q", func(t *testing.T) {
		wrapped := Errorf("wrapped")
		got := fmt.Sprintf("%+q", Errorf("test error: %w", wrapped))
		match, err := regexp.MatchString(strings.TrimSpace(`
"test error: wrapped"
stackerr_test.go:\d+ test error: wrapped
stackerr_test.go:\d+ wrapped
`), got)
		if err != nil {
			t.Fatal(err)
		}
		if !match {
			t.Errorf("regexp doesn't match output:\n%s", got)
		}
	})
}

func TestJoined(t *testing.T) {
	t.Run("%v", func(t *testing.T) {
		err1 := Errorf("err1")
		err2 := Errorf("err2")
		err3 := Errorf("err3")
		got := fmt.Sprintf("%v", errors.Join(err1, err2, err3))
		want := "err1\nerr2\nerr3"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
	t.Run("%+v", func(t *testing.T) {
		err1 := Errorf("err1")
		err2 := Errorf("err2")
		err3 := Errorf("err3")
		got := fmt.Sprintf("%+v", Errorf("joined:\n%w", errors.Join(err1, err2, err3)))
		match, err := regexp.MatchString(strings.TrimSpace(`
joined:
err1
err2
err3
stackerr_test.go:\d+ "joined:\\nerr1\\nerr2\\nerr3"
	\[0\] stackerr_test.go:\d+ err1
	\[1\] stackerr_test.go:\d+ err2
	\[2\] stackerr_test.go:\d+ err3
`), got)
		if err != nil {
			t.Fatal(err)
		}
		if !match {
			t.Errorf("regexp doesn't match output:\n%s", got)
		}
	})
	t.Run("%s", func(t *testing.T) {
		err1 := Errorf("err1")
		err2 := Errorf("err2")
		err3 := Errorf("err3")
		got := fmt.Sprintf("%s", errors.Join(err1, err2, err3))
		want := "err1\nerr2\nerr3"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
	t.Run("%+s", func(t *testing.T) {
		err1 := Errorf("err1")
		err2 := Errorf("err2")
		err3 := Errorf("err3")
		got := fmt.Sprintf("%+s", Errorf("joined:\n%w", errors.Join(err1, err2, err3)))
		match, err := regexp.MatchString(strings.TrimSpace(`
joined:
err1
err2
err3
stackerr_test.go:\d+ "joined:\\nerr1\\nerr2\\nerr3"
	\[0\] stackerr_test.go:\d+ err1
	\[1\] stackerr_test.go:\d+ err2
	\[2\] stackerr_test.go:\d+ err3
`), got)
		if err != nil {
			t.Fatal(err)
		}
		if !match {
			t.Errorf("regexp doesn't match output:\n%s", got)
		}
	})
	t.Run("%q", func(t *testing.T) {
		err1 := Errorf("err1")
		err2 := Errorf("err2")
		err3 := Errorf("err3")
		got := fmt.Sprintf("%q", Errorf("joined:\n%w", errors.Join(err1, err2, err3)))
		want := `"joined:\nerr1\nerr2\nerr3"`
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
	t.Run("%+q", func(t *testing.T) {
		err1 := Errorf("err1")
		err2 := Errorf("err2")
		err3 := Errorf("err3")
		got := fmt.Sprintf("%+q", Errorf("joined:\n%w", errors.Join(err1, err2, err3)))
		match, err := regexp.MatchString(strings.TrimSpace(`
"joined:\\nerr1\\nerr2\\nerr3"
stackerr_test.go:\d+ "joined:\\nerr1\\nerr2\\nerr3"
	\[0\] stackerr_test.go:\d+ err1
	\[1\] stackerr_test.go:\d+ err2
	\[2\] stackerr_test.go:\d+ err3
`), got)
		if err != nil {
			t.Fatal(err)
		}
		if !match {
			t.Errorf("regexp doesn't match output:\n%s", got)
		}
	})
}

func TestIs(t *testing.T) {
	wrapped := Errorf("wrapped")
	err := Errorf("error: %w", wrapped)
	if !errors.Is(err, wrapped) {
		t.Errorf("error %q doesn't unwrap to %q", err, wrapped)
	}

	wrapped = os.ErrNotExist
	err = Errorf("error: %w", wrapped)
	if !errors.Is(err, wrapped) {
		t.Errorf("error %q doesn't unwrap to %q", err, wrapped)
	}
}

func TestAs(t *testing.T) {
	var unwrapped *os.PathError
	wrapped := &os.PathError{Op: "test", Path: "/test/path", Err: fmt.Errorf("error")}
	err := Errorf("error: %w", wrapped)
	if !errors.As(err, &unwrapped) {
		t.Errorf("error %q doesn't unwrap as %T", err, unwrapped)
	}
}
