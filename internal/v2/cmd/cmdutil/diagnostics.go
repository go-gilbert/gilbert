package cmdutil

import (
	"bytes"
	"errors"
	"io"
	"iter"
	"os"
	"strconv"
	"strings"
	"unsafe"

	"github.com/valyala/bytebufferpool"

	"github.com/go-gilbert/gilbert/internal/v2/log"
	"github.com/go-gilbert/gilbert/internal/v2/ui/theme"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
)

const sourceLinesCount = 6

type ErrorNote interface {
	Note() string
}

var padBuff = []byte("            ")

// var padBuff = []byte("______________")
func getPad(size int) string {
	if size > len(padBuff) {
		padBuff = bytes.Repeat([]byte(" "), size)
	}

	chunk := padBuff[:size]
	return bytesToString(chunk)
}

type filePool map[string][][]byte

func (p filePool) getFileLines(fname string) ([][]byte, error) {
	// TODO: find more optimal way
	lines, ok := p[fname]
	if ok {
		return lines, nil
	}

	buff, err := os.ReadFile(fname)
	if err != nil {
		return nil, err
	}

	lines = bytes.Split(buff, []byte("\n"))
	p[fname] = lines
	return lines, nil
}

func (p filePool) getDiagLine(diag *parsetypes.Diagnostic) ([]byte, error) {
	lines, err := p.getFileLines(diag.FileName)
	if err != nil {
		return nil, err
	}

	i := diag.Range.Start.Line - 1
	if i >= len(lines) {
		return nil, nil
	}

	return lines[i], nil
}

func (p filePool) iterDiagLines(diag *parsetypes.Diagnostic, lineCount int) (iter.Seq2[int, []byte], error) {
	lines, err := p.getFileLines(diag.FileName)
	if err != nil {
		return nil, err
	}

	i := diag.Range.Start.Line - 1
	if i >= len(lines) || i < 0 {
		return nil, nil
	}

	lineCount = max(1, lineCount)
	half := (lineCount - 1) / 2
	if lineCount%2 == 0 {
		half++
	}

	start := i
	end := i + 1
	if half > 0 {
		start = i - half + 1
		end += half
	} else if lineCount == 2 {
		start = i - 1
	}

	start = max(0, start)
	end = min(len(lines)-1, end)

	return func(yield func(int, []byte) bool) {
		for i := start; i < end; i++ {
			yield(i+1, lines[i])
		}
	}, nil
}

type DiagnosticsRenderer struct {
	logger  *log.Logger
	dst     io.Writer
	opts    BootstrapOpts
	fp      filePool
	palette theme.Palette
}

func NewDiagnosticsRenderer(logger *log.Logger, opts BootstrapOpts) *DiagnosticsRenderer {
	return &DiagnosticsRenderer{
		logger:  logger,
		opts:    opts,
		fp:      make(filePool, 3),
		dst:     os.Stderr,
		palette: theme.NewPalette(opts.NoColor),
	}
}

// SetWriter sets custom output destination for diagnostics.
func (r *DiagnosticsRenderer) SetWriter(w io.Writer) {
	r.dst = w
}

// Reset clears any buffered data
func (r *DiagnosticsRenderer) Reset() {
	r.fp = make(filePool, 3)
}

// RenderDiagnostics renders diagnostics in human-friendly way.
func (r *DiagnosticsRenderer) RenderDiagnostics(diags parsetypes.Diagnostics) {
	if len(diags) == 0 {
		return
	}

	if r.opts.JSON {
		renderDiagnosticsJSON(r.logger, diags)
		return
	}

	// palette := newDiagColorPalette(r.opts.NoColor)
	for _, diag := range diags {
		r.renderDiagnostic(diag)
	}
}

func (r *DiagnosticsRenderer) renderDiagnostic(diag *parsetypes.Diagnostic) {
	// TODO: multiline support
	buff := bytebufferpool.Get()
	defer func() {
		r.palette.Reset.Fprintln(buff)
		_, _ = r.dst.Write(buff.Bytes())
		bytebufferpool.Put(buff)
	}()

	switch diag.Severity {
	case parsetypes.DiagnosticSeverityError:
		r.palette.DiagError.Fprint(buff, "error: ")
	case parsetypes.DiagnosticSeverityWarning:
		r.palette.DiagWarn.Fprint(buff, "warning: ")
	default:
		r.palette.DiagWarn.Fprint(buff, "note: ")
	}

	// Severity
	r.palette.DiagMsg.Fprintln(buff, diag.Err.Error())

	// Error message & filename
	lineNumber := strconv.Itoa(max(diag.Range.Start.Line, diag.Range.End.Line))
	r.palette.Gutter.Fprint(buff, getPad(len(lineNumber)), "--> ")
	r.palette.Reset.Fprintf(buff, "%s:%s\n", diag.FileName, diag.Range.Start)

	if diag.Range.IsEmpty() {
		return
	}

	// Source text
	lines, err := r.fp.iterDiagLines(diag, sourceLinesCount)
	if err != nil {
		return
	}

	startChar := max(0, diag.Range.Start.Column-1)
	highlightLen := diag.Range.End.Column - diag.Range.Start.Column + 1

	errLine := diag.Range.Start.Line
	for lineNo, line := range lines {
		if lineNo != errLine {
			r.palette.Gutter.Fprintf(buff, "%d | ", lineNo)
			r.palette.Reset.Fprintln(buff, bytesToString(line))
			continue
		}

		r.palette.Gutter.Fprintf(buff, "%s | ", lineNumber)
		r.palette.Reset.Fprintln(buff, bytesToString(line))

		// Draw highlight & annotation
		msg := diag.Note
		if msg == "" {
			msg = tryGetErrorReason(diag.Err)
		}

		noteColor := r.palette.GetHighlightColor(diag.Severity)

		r.palette.Gutter.Fprint(buff, getPad(len(lineNumber)), " | ")
		r.palette.Reset.Fprint(buff, getPad(startChar))
		noteColor.Fprint(buff, strings.Repeat("^", highlightLen), " ", msg)
		r.palette.Reset.Fprintln(buff)
	}
}

func renderDiagnosticsJSON(logger *log.Logger, diags parsetypes.Diagnostics) {
	l := logger.Named("diagnostics")
	for _, diag := range diags {
		fields := []log.Field{
			log.NewField("diagnostic", diag),
		}
		if diag.Severity == parsetypes.DiagnosticSeverityWarning {
			l.Warnw(diag.Err.Error(), fields...)
			continue
		}
		l.Errorw(diag.Err.Error(), fields...)
	}
}

func tryGetErrorReason(err error) string {
	if errNote, ok := err.(ErrorNote); ok {
		return errNote.Note()
	}

	unwrapped := errors.Unwrap(err)
	if unwrapped == nil {
		return ""
	}

	// errors.Unwrap doesn't work recursively
	for {
		newUnwrapped := errors.Unwrap(unwrapped)
		if newUnwrapped == nil {
			break
		}

		unwrapped = newUnwrapped
	}

	return unwrapped.Error()
}

// bytesToString casts bytes to string w/o copy, copied from strings.Builder.String()
func bytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}
