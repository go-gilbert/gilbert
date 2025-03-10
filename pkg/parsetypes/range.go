package parsetypes

import "fmt"

type Position struct {
	Line   int
	Column int
}

func (p Position) Add(line int, column int) Position {
	p.Line += line
	p.Column += column
	return p
}

func (p Position) IsEmpty() bool {
	return p.Line == 0 && p.Column == 0
}

func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

func NewPosition(line int, column int) Position {
	return Position{
		Line:   line,
		Column: column,
	}
}

type Range struct {
	Start Position
	End   Position
}

func (r Range) IsEmpty() bool {
	return r.Start.IsEmpty() && r.End.IsEmpty()
}

func NewRange(start, end Position) Range {
	return Range{
		Start: start,
		End:   end,
	}
}

type OffsetRange struct {
	Start int
	End   int
}

func NewOffsetRange(offset, count int) OffsetRange {
	return OffsetRange{
		Start: offset,
		End:   offset + count,
	}
}
