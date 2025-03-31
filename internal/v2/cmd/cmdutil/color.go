package cmdutil

import (
	"fmt"
	"io"

	"github.com/fatih/color"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
)

var (
	colorDiagError  = color.New(color.FgHiRed, color.Bold)
	colorDiagWarn   = color.New(color.FgYellow, color.Bold)
	colorDiagMsg    = color.New(color.FgHiWhite, color.Bold)
	colorErrMarker  = color.New(color.FgRed)
	colorWarnMarker = color.New(color.FgYellow)
	colorNoteMarker = color.New(color.FgHiBlue)
	colorGutter     = color.New(color.ResetBold, color.FgHiBlue)
	colorReset      = color.New(color.Reset)
)

type colorPrinter interface {
	Fprintf(w io.Writer, format string, a ...interface{}) (int, error)
	Fprint(w io.Writer, a ...interface{}) (int, error)
	Fprintln(w io.Writer, a ...interface{}) (int, error)
}

type nopColor struct{}

func (_ nopColor) Fprintf(w io.Writer, format string, a ...interface{}) (int, error) {
	return fmt.Fprintf(w, format, a...)
}

func (_ nopColor) Fprint(w io.Writer, a ...interface{}) (int, error) {
	return fmt.Fprint(w, a...)
}

func (_ nopColor) Fprintln(w io.Writer, a ...interface{}) (int, error) {
	return fmt.Fprintln(w, a...)
}

type diagColorPalette struct {
	diagError  colorPrinter
	diagWarn   colorPrinter
	diagMsg    colorPrinter
	gutter     colorPrinter
	errMarker  colorPrinter
	warnMarker colorPrinter
	noteMarker colorPrinter
	reset      colorPrinter
}

func (pal *diagColorPalette) getHighlightColor(severity parsetypes.DiagnosticSeverity) colorPrinter {
	switch severity {
	case parsetypes.DiagnosticSeverityError:
		return pal.errMarker
	case parsetypes.DiagnosticSeverityWarning:
		return pal.warnMarker
	default:
		return pal.noteMarker
	}
}

func newDiagColorPalette(noColor bool) diagColorPalette {
	if noColor {
		return diagColorPalette{
			diagError:  nopColor{},
			diagWarn:   nopColor{},
			diagMsg:    nopColor{},
			gutter:     nopColor{},
			errMarker:  nopColor{},
			noteMarker: nopColor{},
			reset:      nopColor{},
		}
	}

	return diagColorPalette{
		diagError:  colorDiagError,
		diagWarn:   colorDiagWarn,
		diagMsg:    colorDiagMsg,
		gutter:     colorGutter,
		noteMarker: colorNoteMarker,
		errMarker:  colorErrMarker,
		warnMarker: colorWarnMarker,
		reset:      colorReset,
	}
}
