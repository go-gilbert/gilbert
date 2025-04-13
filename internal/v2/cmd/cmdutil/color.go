package cmdutil

import (
	"fmt"
	"io"
	"strings"

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
	colorHeading    = color.New(color.FgHiWhite, color.Bold)
	colorReset      = color.New(color.Reset)
)

type ColorPrinter interface {
	Fprintf(w io.Writer, format string, a ...interface{}) (int, error)
	Fprint(w io.Writer, a ...interface{}) (int, error)
	Fprintln(w io.Writer, a ...interface{}) (int, error)
}

type nopColor struct{}

func (_ nopColor) Fprintf(w io.Writer, format string, a ...any) (int, error) {
	return fmt.Fprintf(w, format, a...)
}

func (_ nopColor) Fprint(w io.Writer, a ...any) (int, error) {
	return fmt.Fprint(w, a...)
}

func (_ nopColor) Fprintln(w io.Writer, a ...any) (int, error) {
	return fmt.Fprintln(w, a...)
}

type UsageColorPalette struct {
	NoColor bool
	Heading ColorPrinter
	Reset   ColorPrinter
}

func (p UsageColorPalette) RenderHeading(str string) string {
	if p.NoColor {
		return str
	}

	sb := &strings.Builder{}
	sb.Grow(len(str) + 8)
	p.Heading.Fprint(sb, str)
	p.Reset.Fprint(sb)
	return sb.String()
}

func NewUsageColorPalette(noColor bool) UsageColorPalette {
	if noColor {
		return UsageColorPalette{
			NoColor: noColor,
			Heading: nopColor{},
			Reset:   nopColor{},
		}
	}

	return UsageColorPalette{
		NoColor: noColor,
		Heading: colorHeading,
		Reset:   colorReset,
	}
}

type diagColorPalette struct {
	noColor    bool
	diagError  ColorPrinter
	diagWarn   ColorPrinter
	diagMsg    ColorPrinter
	gutter     ColorPrinter
	errMarker  ColorPrinter
	warnMarker ColorPrinter
	noteMarker ColorPrinter
	reset      ColorPrinter
}

func (pal *diagColorPalette) getHighlightColor(severity parsetypes.DiagnosticSeverity) ColorPrinter {
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
			noColor:    noColor,
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
		noColor:    noColor,
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
