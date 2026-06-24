package sfv

import (
	"bytes"
	"encoding/base64"
)

// ByteSequenceItem represents a base64-encoded byte sequence value,
// with optional parameters.
//
// ByteSequenceItem implements the Item interface.
type ByteSequenceItem = FullItem[*ByteSequenceBareItem, []byte]

var _ Item = (*ByteSequenceItem)(nil)

// ByteSequenceBareItem is a bare item that represents a base64-encoded byte sequence.
// Bare items cannot have parameters. Some constructs
// may require a bare item instead of a full byte sequence item
// (e.g. dictionary values).
type ByteSequenceBareItem struct {
	uvalue[[]byte]
}

var _ BareItem = (*ByteSequenceBareItem)(nil)

// ByteSequence creates a new ByteSequence (ByteSequenceItem) with the
// given byte slice. This function does NOT validate the bytes
// to ensure they form a valid byte sequence (Validation only happens
// when the item is marshaled/parsed).
//
// If you need a bare byte sequence item, use BareByteSequence() instead.
func ByteSequence(b []byte) *ByteSequenceItem {
	return BareByteSequence(b).toItem()
}

func (b *ByteSequenceBareItem) toItem() *ByteSequenceItem {
	return &ByteSequenceItem{
		bare:   b,
		params: NewParameters(),
	}
}

// BareByteSequence creates a new ByteSequenceBareItem with the given byte slice.
// This function does NOT validate the bytes to ensure they form a
// valid byte sequence (Validation only happens when the item is
// marshaled/parsed).
//
// If you need a full byte sequence item (with parameters), use ByteSequence() instead.
func BareByteSequence(b []byte) *ByteSequenceBareItem {
	var v ByteSequenceBareItem
	_ = v.SetValue(b)
	return &v
}

// ToItem converts the ByteSequenceBareItem to a full Item.
func (b *ByteSequenceBareItem) ToItem() Item {
	return b.toItem()
}

// MarshalSFV implements the Marshaler interface for ByteSequenceBareItem.
func (b ByteSequenceBareItem) MarshalSFV() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(':')
	buf.WriteString(base64.StdEncoding.EncodeToString(b.value))
	buf.WriteByte(':')
	return buf.Bytes(), nil
}

// Type returns the type of the ByteSequenceBareItem, useful when
// you have a list of BareItems and need to know the type
// of each item.
func (b ByteSequenceBareItem) Type() int {
	return ByteSequenceType
}
