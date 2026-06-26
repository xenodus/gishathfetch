package sfv

import (
	"bytes"
)

// TokenItem represents a token, an unquoted string value,
// with optional parameters.
//
// TokenItem implements the Item interface.
type TokenItem = FullItem[*TokenBareItem, string]

var _ Item = (*TokenItem)(nil)

// TokenBareItem is a bare item that represents a token.
// Bare items cannot have parameters. Some constructs
// may require a bare item instead of a full token item
// (e.g. dictionary values).
type TokenBareItem struct {
	uvalue[string]
}

var _ BareItem = (*TokenBareItem)(nil)

// Token creates a new Token (TokenItem) with the
// given string. This function does NOT validate the string
// to ensure it is a valid token (Validation only happens
// when the item is marshaled/parsed).
//
// If you need a bare token item, use BareToken() instead.
func Token(s string) *TokenItem {
	return BareToken(s).toItem()
}

func (t *TokenBareItem) toItem() *TokenItem {
	return &TokenItem{
		bare:   t,
		params: NewParameters(),
	}
}

// BareToken creates a new TokenBareItem with the given string.
// This function does NOT validate the string to ensure it is a
// valid token (Validation only happens when the item is
// marshaled/parsed).
//
// If you need a full token item (with parameters), use Token() instead.
func BareToken(s string) *TokenBareItem {
	var v TokenBareItem
	_ = v.SetValue(s)
	return &v
}

// ToItem converts the TokenBareItem to a full Item.
func (t *TokenBareItem) ToItem() Item {
	return t.toItem()
}

// MarshalSFV implements the Marshaler interface for TokenBareItem.
func (t TokenBareItem) MarshalSFV() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(t.value)
	return buf.Bytes(), nil
}

// Type returns the type of the TokenBareItem, useful when
// you have a list of BareItems and need to know the type
// of each item.
func (t TokenBareItem) Type() int {
	return TokenType
}
