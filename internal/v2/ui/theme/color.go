package theme

import (
	"fmt"
	"io"
	"os"
)

// Color is abstract interface to implement colors.
type Color interface {
	Fprintf(w io.Writer, format string, a ...any) (int, error)
	Fprint(w io.Writer, a ...any) (int, error)
	Fprintln(w io.Writer, a ...any) (int, error)
}

type NoOpColor struct{}

func (NoOpColor) Fprintf(w io.Writer, format string, a ...any) (int, error) {
	return fmt.Fprintf(w, format, a...)
}

func (NoOpColor) Fprint(w io.Writer, a ...any) (int, error) {
	return fmt.Fprint(w, a...)
}

func (NoOpColor) Fprintln(w io.Writer, a ...any) (int, error) {
	return fmt.Fprintln(w, a...)
}

func Print(p Color, args ...any) {
	_, _ = p.Fprint(os.Stdout, args...)
}

func Printf(p Color, format string, args ...any) {
	_, _ = p.Fprintf(os.Stdout, format, args...)
}

func Println(p Color, args ...any) {
	_, _ = p.Fprintln(os.Stdout, args...)
}

func Eprintln(p Color, args ...any) {
	_, _ = p.Fprintln(os.Stderr, args...)
}

func Eprintf(p Color, format string, args ...any) {
	_, _ = p.Fprintf(os.Stderr, format, args...)
}
