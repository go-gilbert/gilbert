package inputflag

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/go-gilbert/gilbert/internal/v2/log"
	"github.com/go-gilbert/gilbert/internal/v2/manifest"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
	"github.com/spf13/pflag"
)

var (
	_ pflag.Value      = (*listInputFlagBinding)(nil)
	_ pflag.SliceValue = (*listInputFlagBinding)(nil)
)

// listInputFlagBinding binds inputs with array values to command-line flags.
type listInputFlagBinding struct {
	inputBindingBase

	logger          *log.Logger
	listItemDecoder func(val string) (any, error)
}

func newListInputFlagBinding(logger *log.Logger, def *manifest.InputDefinition, flagCtx inputFlagContext) *listInputFlagBinding {
	return &listInputFlagBinding{
		logger:           logger,
		inputBindingBase: newInputBindingBase(def, flagCtx),
	}
}

func (b *listInputFlagBinding) checkItemDecoder() error {
	if b.listItemDecoder != nil {
		return nil
	}

	if b.err != nil {
		// return last decoder init error
		return b.err
	}

	return fmt.Errorf("%T: list item decoder it not initialized", b)
}

func (b *listInputFlagBinding) initDecoder() error {
	t := b.inputDef.Schema
	if t.Type != manifest.ValueTypeList {
		// this should never happen
		return errors.New("listInputFlagBinding should be used only for arrays")
	}

	if t.Items == nil {
		return fmt.Errorf("missing list item type definition for input %q", b.inputDef.Name)
	}

	if t.Items.Type.IsComplex() {
		return fmt.Errorf(
			"bad type in input %q: task and global inputs only support list types with scalar values",
			b.inputDef.Name,
		)
	}

	dec, err := getValueDecoder(*t.Items)
	if err != nil {
		return fmt.Errorf("failed to init decoder for input %q: %w", b.inputDef.Name, err)
	}

	b.listItemDecoder = dec
	return nil
}

func (b *listInputFlagBinding) initDefaultValue(ctx context.Context) {
	if err := b.initDecoder(); err != nil {
		b.err = err
		return
	}

	if err := b.initDefaultFromDef(ctx); err != nil {
		b.err = err
		return
	}

	if binding := b.inputDef.Binding; binding != nil && binding.EnvVarName != "" {
		varVal := b.flagCtx.envVars[binding.EnvVarName]
		if err := b.setValueFromInput(varVal, true); err != nil {
			b.err = err
			return
		}
	}
}

func (b *listInputFlagBinding) setValueFromInput(str string, isDefault bool) error {
	if str == "" {
		return nil
	}

	if isDefault {
		b.dirtyStatus = valueDefault
	} else {
		b.dirtyStatus = valueDirty
	}

	items, err := b.itemsFromString(str)
	if err != nil {
		return err
	}

	b.flagCtx.dstScope.Inputs[b.inputDef.Name] = items
	return nil
}

func (b *listInputFlagBinding) itemsFromString(str string) ([]any, error) {
	delim := b.inputDef.Binding.DelimiterOrDefault()
	chunks := strings.Split(str, delim)
	dst := make([]any, len(chunks))

	for i, item := range chunks {
		val, err := b.listItemDecoder(item)
		if err != nil {
			return nil, err
		}

		dst[i] = val
	}

	return dst, nil
}

func (b *listInputFlagBinding) Set(val string) error {
	if err := b.checkItemDecoder(); err != nil {
		return err
	}

	if err := b.setValueFromInput(val, false); err != nil {
		b.err = err
		return err
	}

	b.dirtyStatus = valueDirty
	return nil
}

func (b *listInputFlagBinding) Append(str string) error {
	if err := b.checkItemDecoder(); err != nil {
		return err
	}

	decVal, err := b.listItemDecoder(str)
	if err != nil {
		return err
	}

	val, ok := b.flagCtx.dstScope.Inputs[b.inputDef.Name]
	if !ok || val == nil {
		b.flagCtx.dstScope.Inputs[b.inputDef.Name] = []any{decVal}
		return nil
	}

	arrVal, err := parsetypes.AnyToList(val)
	if err != nil {
		return err
	}

	arrVal = append(arrVal, decVal)
	b.flagCtx.dstScope.Inputs[b.inputDef.Name] = arrVal
	return nil
}

func (b *listInputFlagBinding) Replace(newItems []string) error {
	if err := b.checkItemDecoder(); err != nil {
		return err
	}

	values := make([]any, len(newItems))
	for i, str := range newItems {
		decVal, err := b.listItemDecoder(str)
		if err != nil {
			return err
		}

		values[i] = decVal
	}

	b.flagCtx.dstScope.Inputs[b.inputDef.Name] = values
	return nil
}

func (b *listInputFlagBinding) GetSlice() []string {
	if !b.inputDef.Schema.Type.IsList() {
		// shouldn't happen
		panic("listInputFlagBinding used on a non-list flag")
	}

	listItemTyp := b.inputDef.Schema.Items
	if listItemTyp.Type.IsComplex() {
		// we can't display this properly
		b.logger.Errorf("list of complex types cannot be used for flags (input %q)", b.inputDef.Name)
		return nil
	}

	val, ok := b.flagCtx.dstScope.Inputs[b.inputDef.Name]
	if !ok || val == nil {
		return nil
	}

	iterator, count, err := parsetypes.IterAny(val)
	if err != nil {
		b.logger.Errorf("value of input %q is not a list: %s", b.inputDef.Name, err)
		return nil
	}

	mapperFn := getStringFormatter(listItemTyp.Type, listItemTyp.DateFormatOrDefault())
	out := make([]string, 0, count)
	for _, v := range iterator {
		out = append(out, mapperFn(v))
	}

	return out
}

func (b *listInputFlagBinding) String() string {
	if !b.inputDef.Schema.Type.IsList() {
		// shouldn't happen
		panic("listInputFlagBinding used on a non-list flag")
	}

	val, ok := b.flagCtx.dstScope.Inputs[b.inputDef.Name]
	if !ok || val == nil {
		return "[]"
	}

	arrVal, err := parsetypes.AnyToList(val)
	if err != nil {
		return jsonEncode(val)
	}

	listItemTyp := b.inputDef.Schema.Items
	if listItemTyp.Type.IsComplex() {
		// we can't display this properly
		return jsonEncode(val)
	}

	mapperFn := getStringFormatter(listItemTyp.Type, listItemTyp.DateFormatOrDefault())
	sb := &strings.Builder{}
	sb.WriteString("[")
	for i, v := range arrVal {
		if i != 0 {
			sb.WriteString(",")
		}

		sb.WriteString(mapperFn(v))
	}
	sb.WriteString("]")
	return sb.String()
}

func jsonEncode(val any) string {
	msg, err := json.Marshal(val)
	if err != nil {
		return fmt.Sprintf("%#v", val)
	}

	return string(msg)
}
