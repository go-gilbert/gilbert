package cmdutil

import (
	"bytes"
	"errors"
	"os"
	"strconv"
	"strings"
	"unsafe"

	"github.com/go-gilbert/gilbert/internal/v2/log"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
)

type ErrorNote interface {
	Note() string
}

var padBuff = []byte("            ")

//var padBuff = []byte("______________")

func getPad(size int) string {
	if size > len(padBuff) {
		padBuff = bytes.Repeat([]byte(" "), size)
	}

	// Cast bytes to string w/o copy, copied from strings.Builder.String()
	chunk := padBuff[:size]
	return unsafe.String(unsafe.SliceData(chunk), len(chunk))
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

func RenderDiagnostics(logger *log.Logger, opts BootstrapOpts, diags parsetypes.Diagnostics) {
	if len(diags) == 0 {
		return
	}

	if opts.JSON {
		renderDiagnosticsJSON(*logger, diags)
		return
	}

	fp := make(filePool, 3)

	palette := newDiagColorPalette(opts.NoColor)
	for _, diag := range diags {
		renderDiagnostic(fp, palette, diag)
	}
}

func renderDiagnostic(fp filePool, palette diagColorPalette, diag *parsetypes.Diagnostic) {
	// TODO: multiline support
	defer colorReset.Fprintln(os.Stderr)

	switch diag.Severity {
	case parsetypes.DiagnosticSeverityError:
		palette.diagError.Fprint(os.Stderr, "error: ")
	case parsetypes.DiagnosticSeverityWarning:
		palette.diagWarn.Fprint(os.Stderr, "warning: ")
	default:
		palette.diagWarn.Fprint(os.Stderr, "note: ")
	}

	// Severity
	palette.diagMsg.Fprintln(os.Stderr, diag.Err.Error())

	// Error message & filename
	lineNumber := strconv.Itoa(max(diag.Range.Start.Line, diag.Range.End.Line))
	palette.gutter.Fprint(os.Stderr, getPad(len(lineNumber)), "--> ")
	palette.reset.Fprintf(os.Stderr, "%s:%s\n", diag.FileName, diag.Range.Start)

	// Source text
	line, err := fp.getDiagLine(diag)
	if err != nil {
		return
	}

	startChar := max(0, diag.Range.Start.Column-1)
	highlightLen := diag.Range.End.Column - diag.Range.Start.Column + 1
	//fmt.Printf(
	//	"%d:%d pad=%d\n",
	//	diag.Range.Start.Column,
	//	diag.Range.End.Column,
	//	diag.Range.End.Column-diag.Range.Start.Column+1,
	//)

	palette.gutter.Fprintf(os.Stderr, "%s | ", lineNumber)
	//palette.gutter.Fprintf(os.Stderr, "%s |#", lineNumber)
	palette.reset.Fprintln(os.Stderr, string(line))

	// Draw highlight & annotation
	msg := tryGetErrorReason(diag.Err)

	//palette.gutter.Fprint(os.Stderr, getPad(len(lineNumber)), " |#")
	palette.gutter.Fprint(os.Stderr, getPad(len(lineNumber)), " | ")
	palette.reset.Fprint(os.Stderr, getPad(startChar))
	palette.errMarker.Fprint(os.Stderr, strings.Repeat("^", highlightLen), " ", msg)
	palette.reset.Fprintln(os.Stderr)
}

func renderDiagnosticsJSON(logger log.Logger, diags parsetypes.Diagnostics) {
	logger = logger.Named("diagnostics")
	for _, diag := range diags {
		fields := []log.Field{
			log.NewField("file", diag.FileName),
			log.NewField("range", diag.Range),
			log.NewField("offset", diag.Offset),
		}
		if diag.Severity == parsetypes.DiagnosticSeverityWarning {
			logger.Warnw(diag.Err.Error(), fields...)
			continue
		}
		logger.Errorw(diag.Err.Error(), fields...)
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
