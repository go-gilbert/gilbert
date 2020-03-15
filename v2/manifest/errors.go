package manifest

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/hcl/v2"
)

const (
	charLine = "\n"
	charTab  = '\t'

	issueURL            = "https://github.com/go-gilbert/gilbert/issues/new"
	internalErrEpilogue = "This might be internal Gilbert error." +
		"Feel free to submit bug on " + issueURL
)

type InternalError struct {
	msg string
}

func NewInternalError(msg string, args ...interface{}) InternalError {
	return InternalError{msg: fmt.Sprintf(msg, args...)}
}

func (err InternalError) Error() string {
	return err.msg
}

type Error struct {
	fileName    string
	description string
	lines       [][]byte
	diags       hcl.Diagnostics
}

func NewDiagnosticsFromPosition(r hcl.Range, msg string, args ...interface{}) hcl.Diagnostics {
	errMsg := fmt.Sprintf(msg, args...)
	return hcl.Diagnostics{
		{
			Severity: hcl.DiagError,
			Summary:  errMsg,
			Detail:   errMsg,
			Context:  &r,
			Subject:  &r,
		},
	}
}

func NewErrorFromManifest(m *Manifest, description string, diags hcl.Diagnostics) *Error {
	return NewError(m.FileName, description, m.src, diags)
}

func NewError(fileName, desc string, contents []byte, diags hcl.Diagnostics) *Error {
	return &Error{
		fileName:    fileName,
		description: desc,
		lines:       bytes.Split(contents, []byte(charLine)),
		diags:       diags,
	}
}

func (err *Error) PrettyPrint() string {
	sb := &strings.Builder{}
	sb.WriteString(fmt.Sprintf("Failed to load file %q:\n", err.fileName))
	if err.description != "" {
		sb.WriteString("(" + err.description + ")\n")
	}
	sb.WriteString("\n")
	for _, diag := range err.diags {
		formatLineError(err.lines, diag, sb)
	}

	return sb.String()
}

func (err *Error) Error() string {
	return err.diags.Error()
}

func formatLineError(lines [][]byte, d *hcl.Diagnostic, sb *strings.Builder) {
	if d.Context == nil && d.Subject == nil {
		sb.WriteString(d.Detail + "\n")
		return
	}

	pos := d.Subject
	if pos == nil {
		pos = d.Context
	}

	sb.WriteString(diagToString(d))
	sb.WriteRune('\n')

	padding := lineSourcePadding(pos.End.Line)
	sb.WriteString(padding)
	defer sb.WriteString(padding)
	for i := pos.Start.Line; i <= pos.End.Line; i++ {
		if len(lines) < i {
			return
		}

		line := lines[i-1] // lines starts with 1 but arrays at 0
		sb.WriteString(wrapWithLineNumber(i, pos.End.Line, line))
		sb.WriteRune('\n')
	}
}

func diagToString(d *hcl.Diagnostic) string {
	var errType string
	switch d.Severity {
	case hcl.DiagWarning:
		errType = "warning"
	case hcl.DiagError:
		errType = "error"
	default:
		errType = "note"
	}

	return errType + ": " + d.Detail
}

func lineSourcePadding(maxLine int) string {
	paddingLen := len(strconv.Itoa(maxLine)) + 2 // include padding between line num
	return strings.Repeat(" ", paddingLen) + "|\n"
}

func wrapWithLineNumber(line, max int, contents []byte) string {
	maxLineLen := len(strconv.Itoa(max))
	lineStr := strconv.Itoa(line)
	lineLen := len(lineStr)

	padding := (maxLineLen - lineLen) + 1

	// trim left suffix if we have ClRf endings
	contents = bytes.TrimSuffix(contents, []byte("\r"))

	return strings.Repeat(" ", padding) + lineStr + " | " + string(contents)
}
