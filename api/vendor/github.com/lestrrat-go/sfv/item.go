package sfv

import (
	"fmt"

	"github.com/lestrrat-go/blackmagic"
)

// BareItemFrom creates a BareItem from the given Go value. This function
// converts common Go types (string, bool, int, int64, float64, float32)
// to their corresponding SFV bare item types. For strings, it creates
// a StringBareItem by default.
func BareItemFrom(value any) (BareItem, error) {
	return bareItemFrom(value, 0)
}

const (
	bareItemStringMode = iota
	bareItemTokenMode
	bareItemDisplayStringMode
)

func bareItemFrom(value any, stringMode int) (BareItem, error) {
	switch v := value.(type) {
	case BareItem:
		return v, nil
	case string:
		switch stringMode {
		case bareItemTokenMode:
			return BareToken(v), nil
		case bareItemDisplayStringMode:
			return BareDisplayString(v), nil
		default:
			return BareString(v), nil
		}
	case bool:
		return BareBoolean(v), nil
	case int:
		return BareInteger(int64(v)), nil
	case int64:
		return BareInteger(v), nil
	case float64:
		return BareDecimal(v), nil
	case float32:
		return BareDecimal(float64(v)), nil
	default:
		return nil, fmt.Errorf("unsupported bare item type %T", v)
	}
}

// This is the actual value, and we're only providing this to avoid
// having to write a lot of boilerplate code for each type.
type uvalue[T any] struct {
	value T
}

func (iv *uvalue[T]) SetValue(value T) error {
	iv.value = value
	return nil
}

func (iv *uvalue[T]) Value() T {
	return iv.value
}

func (iv uvalue[T]) GetValue(dst any) error {
	return blackmagic.AssignIfCompatible(dst, iv.value)
}

// FullItem is a generic container that combines a BareItem with Parameters
// to create a complete SFV Item. It serves as the base implementation for
// all typed item aliases (StringItem, IntegerItem, etc.) in the SFV format.
type FullItem[BT BareItem, UT any] struct {
	bare    BT
	valuefn func() UT
	params  *Parameters
}

func (fi *FullItem[BT, UT]) Parameters() *Parameters {
	return fi.params
}

func (fi *FullItem[BT, UT]) Value() UT {
	return fi.valuefn()
}

func (fi *FullItem[BT, UT]) MarshalSFV() ([]byte, error) {
	bi, err := fi.bare.MarshalSFV()
	if err != nil {
		return nil, fmt.Errorf("error marshaling bare item: %w", err)
	}

	// Add parameters if any
	if fi.params != nil && fi.params.Len() > 0 {
		paramBytes, err := fi.params.MarshalSFV()
		if err != nil {
			return nil, err
		}
		bi = append(bi, paramBytes...)
	}

	return bi, nil
}

func (fi *FullItem[BT, UT]) GetValue(dst any) error {
	return fi.bare.GetValue(dst)
}

func (fi *FullItem[BT, UT]) Type() int {
	return fi.bare.Type()
}

func (fi *FullItem[BT, UT]) Parameter(name string, value any) error {
	bi, err := bareItemFrom(value, bareItemStringMode)
	if err != nil {
		return fmt.Errorf("failed to create bare item for parameter %s: %w", name, err)
	}

	if err := fi.params.Set(name, bi); err != nil {
		return fmt.Errorf("failed to set parameter %s: %v", name, err)
	}
	return nil
}

func (fi *FullItem[BT, UT]) With(params *Parameters) Item {
	return &FullItem[BT, UT]{
		bare:   fi.bare,
		params: params,
	}
}

// CoreItem represents the core API that is shared by both
// Item and BareItem.
type CoreItem interface {
	Marshaler
	Type() int
	// GetValue is a method that assigns the underlying value of the item to dst.
	// It is used to retrieve the value without needing to know the type, or
	// without having to go through type conversion.
	//
	// If you already know the type of the value, you could use the Value() method
	// instead, which returns the value directly.
	GetValue(dst any) error
}

// A BareItem represents a bare item, which is the itemValue plus the item
// type. A bare item cannot carry parameters. However, it _can_ be upgraded
// to a full Item by calling ToItem().
type BareItem interface {
	CoreItem

	// ToItem creates a new Item from this bare item
	ToItem() Item
}

// Item represents a single item in the SFV (Structured Field Value) format.
// It is essentially a bare item with parameters
type Item interface {
	CoreItem

	With(*Parameters) Item
	Parameters() *Parameters
}
