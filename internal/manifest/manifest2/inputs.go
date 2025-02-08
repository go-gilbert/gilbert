package manifest2

type ValueType uint8

const (
	ValueTypeInvalid ValueType = iota
	ValueTypeString
	ValueTypeInt
	ValueTypeFloat
	ValueTypeDuration
	ValueTypeList
)

type InputSchema struct {
	// TODO
}
