package parsetypes

import "fmt"

type Location struct {
	FileName string      `json:"fileName"`
	Offset   OffsetRange `json:"offset"`
	Range    Range       `json:"range"`
}

type Position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

func (p Position) Add(line, column int) Position {
	p.Line += line
	p.Column += column
	return p
}

func (p Position) Sub(line, column int) Position {
	p.Line -= line
	p.Column -= column
	return p
}

func (p Position) IsEmpty() bool {
	return p.Line == 0 && p.Column == 0
}

func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

func NewEmptyPosition() Position {
	return Position{
		Line:   1,
		Column: 0,
	}
}

func NewPosition(line int, column int) Position {
	return Position{
		Line:   line,
		Column: column,
	}
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

func (r Range) IsEmpty() bool {
	return r.Start.IsEmpty() && r.End.IsEmpty()
}

// WithStartPosition returns a copy of a range with updated start position.
func (r Range) WithStartPosition(pos Position) Range {
	r.Start = pos
	return r
}

// WithEndPosition returns a copy of a range with updated end position.
func (r Range) WithEndPosition(pos Position) Range {
	r.End = pos
	return r
}

func (r Range) String() string {
	return r.Start.String() + "-" + r.End.String()
}

func NewRange(start, end Position) Range {
	return Range{
		Start: start,
		End:   end,
	}
}

type OffsetRange struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

func NewOffsetRange(offset, count int) OffsetRange {
	return OffsetRange{
		Start: offset,
		End:   offset + count,
	}
}
