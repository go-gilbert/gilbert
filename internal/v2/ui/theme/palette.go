package theme

import (
	"github.com/fatih/color"

	"github.com/go-gilbert/gilbert/pkg/parsetypes"
)

var (
	colorDimmed     = color.RGB(66, 66, 66).Add(color.ResetBold)
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

type Palette struct {
	NoColor bool

	TextDimmed  Color
	TextHeading Color
	DiagError   Color
	DiagWarn    Color
	DiagMsg     Color
	Gutter      Color
	ErrMarker   Color
	WarnMarker  Color
	NoteMarker  Color
	Reset       Color
}

func (pal *Palette) GetHighlightColor(severity parsetypes.DiagnosticSeverity) Color {
	switch severity {
	case parsetypes.DiagnosticSeverityError:
		return pal.ErrMarker
	case parsetypes.DiagnosticSeverityWarning:
		return pal.WarnMarker
	default:
		return pal.NoteMarker
	}
}

func NewPalette(noColor bool) Palette {
	if noColor {
		return Palette{
			NoColor:     noColor,
			TextHeading: NoOpColor{},
			TextDimmed:  NoOpColor{},
			DiagError:   NoOpColor{},
			DiagWarn:    NoOpColor{},
			DiagMsg:     NoOpColor{},
			Gutter:      NoOpColor{},
			ErrMarker:   NoOpColor{},
			WarnMarker:  NoOpColor{},
			NoteMarker:  NoOpColor{},
			Reset:       NoOpColor{},
		}
	}

	return Palette{
		TextDimmed:  colorDimmed,
		TextHeading: colorHeading,
		DiagError:   colorDiagError,
		DiagWarn:    colorDiagWarn,
		DiagMsg:     colorDiagMsg,
		Gutter:      colorGutter,
		ErrMarker:   colorErrMarker,
		WarnMarker:  colorWarnMarker,
		NoteMarker:  colorNoteMarker,
		Reset:       colorReset,
	}
}
