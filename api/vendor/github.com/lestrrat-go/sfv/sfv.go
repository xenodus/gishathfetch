package sfv

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/lestrrat-go/sfv/internal/tokens"
)

// Value is the top-level interface for all SFV (Structured Field Value) types.
// It represents any value that can be marshaled according to the SFV specification.
// This includes Items, Lists, Dictionaries, and any custom types that implement
// the Marshaler interface.
type Value interface { //nolint:iface
	Marshaler
}

// RFC 9651 Section 3.3.1: Integers have at most 15 decimal digits
const (
	maxIntegerDigits = 15
	maxSFVInteger    = 999999999999999
)

const (
	parseModeDefault = 0

	parseModeList = iota // parseModeDefault == parseModelist
	parseModeDictionary
	parseModeItem
)

type parseContext struct {
	idx   int // current index in the data
	size  int // size of the data
	mode  int
	data  []byte
	value any // the parsed value, if any
}

func Parse(data []byte) (any, error) {
	return parse(data, parseModeDefault)
}

func parse(data []byte, mode int) (any, error) {
	var pctx parseContext
	pctx.init(data, mode)
	if err := pctx.do(); err != nil {
		return nil, err
	}
	return pctx.value, nil
}

func ParseDictionary(data []byte) (*Dictionary, error) {
	v, err := parse(data, parseModeDictionary)
	if err != nil {
		return nil, err
	}
	dict, ok := v.(*Dictionary)
	if !ok {
		return nil, fmt.Errorf("expected *Dictionary, got %T", v)
	}
	return dict, nil
}

func ParseItem(data []byte) (Item, error) {
	v, err := parse(data, parseModeItem)
	if err != nil {
		return nil, err
	}
	item, ok := v.(Item)
	if !ok {
		return nil, fmt.Errorf("expected Item, got %T", v)
	}
	return item, nil
}

func (pctx *parseContext) init(data []byte, mode int) {
	pctx.data = data
	pctx.size = len(data)
	pctx.idx = 0
	pctx.mode = mode
}

func (pctx *parseContext) eof() bool {
	return pctx.idx >= pctx.size
}

func (pctx *parseContext) current() byte {
	if pctx.eof() {
		return 0 // EOF
	}
	return pctx.data[pctx.idx]
}

func (pctx *parseContext) advance() {
	if pctx.eof() {
		return
	}
	pctx.idx++
}

func (pctx *parseContext) stripWhitespace() {
	for !pctx.eof() && unicode.IsSpace(rune(pctx.data[pctx.idx])) {
		pctx.advance()
	}
}

// isDictionary checks if the input looks like a dictionary by looking for key=value patterns
func (pctx *parseContext) isDictionary() bool {
	// Save current position
	savedIdx := pctx.idx
	defer func() { pctx.idx = savedIdx }()

	pctx.stripWhitespace()
	if pctx.eof() {
		return false
	}

	// Look for key=value pattern
	// First, try to find a token (key)
	if !isAlpha(pctx.current()) && pctx.current() != '*' {
		return false
	}

	// Skip token characters
	for !pctx.eof() && (isAlpha(pctx.current()) || isDigit(pctx.current()) ||
		pctx.current() == '_' || pctx.current() == '-' || pctx.current() == '.' ||
		pctx.current() == ':' || pctx.current() == '/' || pctx.current() == '*') {
		pctx.advance()
	}

	pctx.stripWhitespace()

	// Check if we have '=' which indicates dictionary
	return !pctx.eof() && pctx.current() == '='
}

func (pctx *parseContext) do() error {
	// RFC 9651 Section 4.2: Parsing Structured Fields algorithm

	// 1. Convert input_bytes into an ASCII string input_string; if conversion fails, fail parsing.
	// (This is already done in init() since we're working with []byte)

	// 2. Discard any leading SP characters from input_string.
	pctx.stripWhitespace()

	// Check if this looks like a dictionary or a list
	var output any
	var err error

	switch pctx.mode {
	case parseModeDictionary:
		output, err = pctx.parseDictionary()
		if err != nil {
			return fmt.Errorf("sfv: failed to parse dictionary: %w", err)
		}
	case parseModeList:
		output, err = pctx.parseList()
		if err != nil {
			return fmt.Errorf("sfv: failed to parse list: %w", err)
		}
	case parseModeItem:
		output, err = pctx.parseItem()
		if err != nil {
			return fmt.Errorf("sfv: failed to parse item: %w", err)
		}

	default:
		if pctx.isDictionary() {
			// 3. Parse as sf-dictionary
			output, err = pctx.parseDictionary()
			if err != nil {
				return fmt.Errorf("sfv: failed to parse dictionary: %w", err)
			}
		} else {
			// 3. Parse as sf-list (the primary structured field type)
			output, err = pctx.parseList()
			if err != nil {
				return fmt.Errorf("sfv: failed to parse list: %w", err)
			}
		}
	}

	// 6. Discard any leading SP characters from input_string.
	pctx.stripWhitespace()

	// 7. If input_string is not empty, fail parsing.
	if !pctx.eof() {
		return fmt.Errorf("sfv: unexpected trailing characters")
	}

	// 8. Otherwise, return output.
	pctx.value = output
	return nil
}

// parseList implements the List parsing algorithm from RFC 9651 Section 4.2.1
func (pctx *parseContext) parseList() (*List, error) {
	var members []any

	for !pctx.eof() {
		// Parse an Item or Inner List - check first character to determine which
		var item any
		var err error

		if pctx.current() == tokens.OpenParen {
			// Parse Inner List
			item, err = pctx.parseInnerList()
			if err != nil {
				return nil, fmt.Errorf("sfv: parse list: expected inner list: %w", err)
			}
		} else {
			// Parse Item
			item, err = pctx.parseItem()
			if err != nil {
				return nil, fmt.Errorf("sfv: parse list: expected item: %w", err)
			}
		}

		members = append(members, item)

		// Discard any leading OWS characters (optional whitespace)
		pctx.stripWhitespace()

		// If input is empty, return the list
		if pctx.eof() {
			return &List{values: members}, nil
		}

		// Consume comma; if not comma, fail parsing
		if pctx.current() != tokens.Comma {
			return nil, fmt.Errorf("sfv: parse list: expected comma, got '%c'", pctx.current())
		}
		pctx.advance() // consume comma

		// Discard any leading OWS characters
		pctx.stripWhitespace()

		// If input is empty after comma, there is a trailing comma; fail parsing
		if pctx.eof() {
			return nil, fmt.Errorf("sfv: parse list: trailing comma")
		}
	}

	// No structured data has been found; return empty list
	return &List{values: members}, nil
}

// parseDictionary implements the Dictionary parsing algorithm from RFC 9651 Section 4.2.2
func (pctx *parseContext) parseDictionary() (*Dictionary, error) {
	dict := NewDictionary()
	for !pctx.eof() {
		// Parse the key (must be a token)
		key, err := pctx.parseKey()
		if err != nil {
			return nil, fmt.Errorf("sfv: parse dictionary: %w", err)
		}

		var value any

		// Check for '=' to see if there's a value
		if !pctx.eof() && pctx.current() == '=' {
			pctx.advance() // consume '='

			// Parse the value (Item or Inner List)
			if pctx.current() == tokens.OpenParen {
				// Parse Inner List
				value, err = pctx.parseInnerList()
				if err != nil {
					return nil, fmt.Errorf("sfv: parse dictionary value: %w", err)
				}
			} else {
				// Parse Item
				value, err = pctx.parseItem()
				if err != nil {
					return nil, fmt.Errorf("sfv: parse dictionary value: %w", err)
				}
			}
		} else {
			// No value specified, create a boolean Item with true value
			value = True()
		}

		// Parse parameters for the dictionary member
		params, err := pctx.parseParameters()
		if err != nil {
			return nil, fmt.Errorf("sfv: parse dictionary parameters: %w", err)
		}

		// If the value has parameters, ensure it's an Item
		if params.Len() > 0 {
			switch v := value.(type) {
			case Item:
				v.With(params)
			case BareItem:
				// Convert BareItem to Item when parameters are present
				value = v.ToItem().With(params)
			}
		}

		dict.keys = append(dict.keys, key)
		dict.values[key] = value

		// Discard any leading OWS characters
		pctx.stripWhitespace()

		// If input is empty, return the dictionary
		if pctx.eof() {
			return dict, nil
		}

		// Consume comma; if not comma, fail parsing
		if pctx.current() != tokens.Comma {
			return nil, fmt.Errorf("sfv: parse dictionary: expected comma, got '%c'", pctx.current())
		}
		pctx.advance() // consume comma

		// Discard any leading OWS characters
		pctx.stripWhitespace()

		// If input is empty after comma, there is a trailing comma; fail parsing
		if pctx.eof() {
			return nil, fmt.Errorf("sfv: parse dictionary: trailing comma")
		}
	}

	return dict, nil
}

func (pctx *parseContext) parseInnerList() (*InnerList, error) {
	pctx.stripWhitespace()
	if pctx.current() != tokens.OpenParen {
		return nil, fmt.Errorf(`sfv: parse inner list: expected '%c', got '%c'`, tokens.OpenParen, pctx.current())
	}
	pctx.advance() // consume opening parenthesis

	var list InnerList
	for !pctx.eof() {
		pctx.stripWhitespace()
		if pctx.current() == tokens.CloseParen {
			// done with this list, consume this character
			pctx.advance()
			params, err := pctx.parseParameters()
			if err != nil {
				return nil, fmt.Errorf("sfv: parse inner list: %w", err)
			}

			if params.Len() > 0 {
				list.params = params
			}
			return &list, nil
		}

		// otherwise, parse an Item
		item, err := pctx.parseItem()
		if err != nil {
			return nil, fmt.Errorf("sfv: parse inner list: %w", err)
		}
		list.values = append(list.values, item)

		// This must be followed by a space or a close paren
		if !pctx.eof() {
			if c := pctx.current(); !unicode.IsSpace(rune(c)) && c != tokens.CloseParen {
				return nil, fmt.Errorf("sfv: parse inner list: expected space or '%c' after item, got '%c'", tokens.CloseParen, c)
			}
		}
	}
	// If we reach here, we've reached EOF without finding a closing paren
	return nil, fmt.Errorf("sfv: parse inner list: unexpected end of input, expected closing paren")
}

// parseKey implements the Key parsing algorithm from RFC 9651 Section 4.2.3.3
func (pctx *parseContext) parseKey() (string, error) {
	// 1. If the first character of input_string is not lcalpha or "*", fail parsing.
	if pctx.eof() {
		return "", fmt.Errorf("sfv: unexpected end of input while parsing key")
	}

	c := pctx.current()
	if !isLowerAlpha(c) && c != tokens.Asterisk {
		return "", fmt.Errorf("sfv: key must start with lowercase letter or asterisk, got '%c'", c)
	}

	// 2. Let output_string be an empty string.
	var sb strings.Builder

	// 3. While input_string is not empty:
	for !pctx.eof() {
		c := pctx.current()

		// 3.1. If the first character of input_string is not one of lcalpha, DIGIT, "_", "-", ".", or "*", return output_string.
		if !isLowerAlpha(c) && !isDigit(c) && c != tokens.Underscore && c != tokens.Dash && c != tokens.Period && c != tokens.Asterisk {
			break
		}

		// 3.2. Let char be the result of consuming the first character of input_string.
		pctx.advance()

		// 3.3. Append char to output_string.
		sb.WriteByte(c)
	}

	// 4. Return output_string.
	result := sb.String()
	if result == "" {
		return "", fmt.Errorf("sfv: empty key")
	}
	return result, nil
}

func isLowerAlpha(c byte) bool {
	return c >= 'a' && c <= 'z'
}

func (pctx *parseContext) parseParameters() (*Parameters, error) {
	// RFC 9651 Section 4.2.3.2: Parsing Parameters
	var keys []string
	var values map[string]BareItem

	for !pctx.eof() {
		// 1. If the first character of input_string is not ";", exit the loop.
		if pctx.current() != tokens.Semicolon {
			break
		}

		// 2. Consume the ";" character from the beginning of input_string.
		pctx.advance()

		// 3. Discard any leading SP characters from input_string.
		pctx.stripWhitespace()

		// 4. Let param_key be the result of running Parsing a Key with input_string.
		paramKey, err := pctx.parseKey()
		if err != nil {
			return nil, fmt.Errorf("sfv: failed to parse parameter key: %w", err)
		}

		// 5. Let param_value be Boolean true.
		var paramValue BareItem = True()

		// 6. If the first character of input_string is "=":
		if !pctx.eof() && pctx.current() == tokens.Equals {
			// 6.1. Consume the "=" character at the beginning of input_string.
			pctx.advance()

			// 6.2. Let param_value be the result of running Parsing a Bare Item with input_string.
			bareItem, err := pctx.parseBareItem()
			if err != nil {
				return nil, fmt.Errorf("sfv: failed to parse parameter value: %w", err)
			}
			paramValue = bareItem
		}

		// Initialize maps on first parameter
		if values == nil {
			values = make(map[string]BareItem)
		}

		// 7. If parameters already contains a key param_key (comparing character for character),
		//    overwrite its value with param_value.
		// 8. Otherwise, append key param_key with value param_value to parameters.
		if _, exists := values[paramKey]; !exists {
			// Only add to keys slice if it's a new key
			keys = append(keys, paramKey)
		}
		values[paramKey] = paramValue
	}

	// Only create Parameters object if we actually have parameters
	if len(keys) == 0 {
		return &Parameters{Values: make(map[string]BareItem)}, nil
	}

	return &Parameters{
		keys:   keys,
		Values: values,
	}, nil
}

const (
	InvalidType = iota
	IntegerType
	DecimalType
	StringType
	TokenType
	ByteSequenceType
	BooleanType
	DateType
	DisplayStringType
)

func (pctx *parseContext) parseItem() (Item, error) {
	bareItem, err := pctx.parseBareItem()
	if err != nil {
		return nil, fmt.Errorf("sfv: failed to parse bare item: %w", err)
	}

	params, err := pctx.parseParameters()
	if err != nil {
		return nil, fmt.Errorf("sfv: failed to parse parameters: %w", err)
	}

	return bareItem.ToItem().With(params), nil
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func (pctx *parseContext) parseBareItem() (BareItem, error) {
	pctx.stripWhitespace()
	switch c := pctx.current(); {
	case c == '-' || isDigit(c):
		v, err := pctx.parseDecimal()
		if err != nil {
			return nil, fmt.Errorf(`sfv: failed to parse bare item (decimal): %w`, err)
		}
		return v, nil
	case c == tokens.DoubleQuote:
		v, err := pctx.parseString()
		if err != nil {
			return nil, fmt.Errorf(`sfv: failed to parse bare item (quoted string): %w`, err)
		}
		return v, nil
	case c == tokens.Asterisk || isAlpha(c):
		v, err := pctx.parseToken()
		if err != nil {
			return nil, fmt.Errorf(`sfv: failed to parse bare item (token): %w`, err)
		}
		return v, nil
	case c == tokens.Colon:
		v, err := pctx.parseByteSequence()
		if err != nil {
			return nil, fmt.Errorf(`sfv: failed to parse bare item (byte sequence): %w`, err)
		}
		return v, nil
	case c == tokens.QuestionMark:
		v, err := pctx.parseBoolean()
		if err != nil {
			return nil, fmt.Errorf(`sfv: failed to parse bare item (boolean): %w`, err)
		}
		return v, nil
	case c == tokens.AtMark:
		v, err := pctx.parseDate()
		if err != nil {
			return nil, fmt.Errorf(`sfv: failed to parse bare item (date): %w`, err)
		}
		return v, nil
	case c == tokens.Percent:
		v, err := pctx.parseDisplayString()
		if err != nil {
			return nil, fmt.Errorf(`sfv: failed to parse bare item (display string): %w`, err)
		}
		return v, nil
	default:
		return nil, fmt.Errorf(`sfv: unrecognized character while parsing bare item: %c`, c)
	}
}

func (pctx *parseContext) parseDecimal() (BareItem, error) {
	var decimal bool
	sign := 1

	if c := pctx.current(); c == tokens.Dash {
		pctx.advance()
		sign = -1
	}

	if pctx.eof() {
		return nil, fmt.Errorf(`sfv: failed to parse numeric value: expected digit`)
	}

	var sb strings.Builder
LOOP:
	for !pctx.eof() {
		c := pctx.current()

		if sb.Len() == 0 && !isDigit(c) {
			return nil, fmt.Errorf(`sfv: failed to parse numeric value: expected digit at the start`)
		}

		switch {
		case c == tokens.Period:
			if decimal {
				// If we already have a decimal point we consider this
				// the end of the number
				break LOOP
			}

			// 12 digits of precision is all we can do
			if sb.Len() > 12 {
				return nil, fmt.Errorf(`sfv: failed to parse numeric value: too many (%d) digits for decimal number`, sb.Len())
			}
			decimal = true
		case !isDigit(c):
			// End of number - break out of loop
			break LOOP
		default:
		}

		pctx.advance()
		sb.WriteByte(c)
	}

	if decimal {
		if sb.Len() > 16 {
			return nil, fmt.Errorf(`sfv: failed to parse numeric value: too many (%d) digits for decimal number`, sb.Len())
		}

		s := sb.String()
		if s[sb.Len()-1] == tokens.Period {
			return nil, fmt.Errorf(`sfv: failed to parse numeric value: expected digit after decimal point`)
		}
		i := strings.IndexByte(s, tokens.Period)
		if sb.Len()-i > 4 { // decimal point + max 3 fractional digits
			return nil, fmt.Errorf(`sfv: failed to parse numeric value: too many (%d) digits after decimal point`, sb.Len()-i-1)
		}

		v, err := strconv.ParseFloat(sb.String(), 64)
		if err != nil {
			return nil, fmt.Errorf(`sfv: failed to parse numeric value as float: %w`, err)
		}
		return BareDecimal(v * float64(sign)), nil
	}

	if sb.Len() > maxIntegerDigits {
		return nil, fmt.Errorf(`sfv: failed to parse numeric value: too many (%d) digits for integer number`, sb.Len())
	}

	v, err := strconv.Atoi(sb.String())
	if err != nil {
		return nil, fmt.Errorf(`sfv: failed to parse numeric value as integer: %w`, err)
	}
	return BareInteger(int64(v * sign)), nil
}

// parseString parses a quoted string according to RFC 9651 Section 4.2.5
func (pctx *parseContext) parseString() (BareItem, error) {
	if pctx.current() != tokens.DoubleQuote {
		return nil, fmt.Errorf("sfv: expected quote at start of string")
	}
	pctx.advance() // consume opening quote

	var sb strings.Builder
	for !pctx.eof() {
		c := pctx.current()
		pctx.advance()

		if c == tokens.Backslash {
			if pctx.eof() {
				return nil, fmt.Errorf("sfv: unexpected end of input after backslash")
			}
			next := pctx.current()
			if next != tokens.DoubleQuote && next != tokens.Backslash {
				return nil, fmt.Errorf("sfv: invalid escape sequence \\%c", next)
			}
			pctx.advance()
			sb.WriteByte(next)
		} else if c == tokens.DoubleQuote {
			return BareString(sb.String()), nil
		} else if c <= 0x1f || c >= 0x7f {
			return nil, fmt.Errorf("sfv: invalid character in string: %c", c)
		} else {
			sb.WriteByte(c)
		}
	}
	return nil, fmt.Errorf("sfv: unexpected end of input, expected closing quote")
}

// parseToken parses a token according to RFC 9651 Section 4.2.6
func (pctx *parseContext) parseToken() (*TokenBareItem, error) {
	// token = (ALPHA / "*") *tchar
	// tchar = "!" / "#" / "$" / "%" / "&" / "'" / "*" / "+" / "-" / "." / "^" / "_" / "`" / "|" / "~" / DIGIT / ALPHA
	c := pctx.current()
	if !isAlpha(c) && c != tokens.Asterisk {
		return nil, fmt.Errorf("sfv: token must start with alpha or asterisk")
	}

	var sb strings.Builder
OUTER:
	for !pctx.eof() {
		c := pctx.current()

		switch {
		case isAlpha(c):
		case isDigit(c):
		default:
			switch c {
			case tokens.Ampersand, tokens.Asterisk,
				tokens.Backtick, tokens.Caret,
				tokens.Colon, tokens.Dash,
				tokens.Dollar, tokens.Exclamation,
				tokens.Hash, tokens.Percent,
				tokens.Period, tokens.Pipe,
				tokens.Plus, tokens.SingleQuote,
				tokens.Slash, tokens.Tilde,
				tokens.Underscore:
			default:
				break OUTER
			}
		}
		sb.WriteByte(c)
		pctx.advance()
	}

	if sb.Len() == 0 {
		return nil, fmt.Errorf("sfv: empty token")
	}

	stok := sb.String()

	return BareToken(stok), nil
}

// parseByteSequence parses a byte sequence according to RFC 9651 Section 4.2.7
func (pctx *parseContext) parseByteSequence() (*ByteSequenceBareItem, error) {
	if pctx.current() != tokens.Colon {
		return nil, fmt.Errorf("sfv: expected colon at start of byte sequence")
	}
	pctx.advance() // consume opening colon

	var sb strings.Builder
	foundClosingColon := false
	for !pctx.eof() {
		c := pctx.current()
		if c == tokens.Colon {
			pctx.advance() // consume closing colon
			foundClosingColon = true
			break
		}
		// Valid base64 characters
		if isAlpha(c) || isDigit(c) || c == tokens.Plus || c == tokens.Slash || c == tokens.Equals {
			sb.WriteByte(c)
			pctx.advance()
		} else {
			return nil, fmt.Errorf("sfv: invalid character in byte sequence: %c", c)
		}
	}

	if !foundClosingColon {
		return nil, fmt.Errorf("sfv: expected closing colon in byte sequence")
	}

	// Decode base64
	decoded, err := base64.StdEncoding.DecodeString(sb.String())
	if err != nil {
		return nil, fmt.Errorf("sfv: failed to decode base64: %w", err)
	}
	return BareByteSequence(decoded), nil
}

// parseBoolean parses a boolean according to RFC 9651 Section 4.2.8
func (pctx *parseContext) parseBoolean() (BooleanBareItem, error) {
	if pctx.current() != tokens.QuestionMark {
		return False(), fmt.Errorf("sfv: expected question mark at start of boolean")
	}
	pctx.advance() // consume question mark

	if pctx.eof() {
		return False(), fmt.Errorf("sfv: unexpected end of input, expected boolean value")
	}

	c := pctx.current()
	pctx.advance()

	switch c {
	case tokens.One:
		return True(), nil
	case tokens.Zero:
		return False(), nil
	default:
		return False(), fmt.Errorf("sfv: invalid boolean value, expected '0' or '1', got %c", c)
	}
}

// parseDate parses a date according to RFC 9651 Section 4.2.9
func (pctx *parseContext) parseDate() (*DateBareItem, error) {
	if pctx.current() != tokens.AtMark {
		return nil, fmt.Errorf("sfv: expected @ at start of date")
	}
	pctx.advance() // consume @ mark

	// Parse the integer value
	value, err := pctx.parseDecimal()
	if err != nil {
		return nil, fmt.Errorf("sfv: failed to parse date integer: %w", err)
	}

	// Date must be an integer, not a decimal
	if value.Type() != IntegerType {
		return nil, fmt.Errorf("sfv: date must be an integer")
	}

	var intValue int64
	if err := value.GetValue(&intValue); err != nil {
		return nil, fmt.Errorf("sfv: failed to convert date value to int64: %w", err)
	}

	return BareDate(intValue), nil
}

// parseDisplayString parses a display string according to RFC 9651 Section 4.2.10
func (pctx *parseContext) parseDisplayString() (*DisplayStringBareItem, error) {
	// Expect %"
	if pctx.current() != tokens.Percent {
		return nil, fmt.Errorf("sfv: expected %% at start of display string")
	}
	pctx.advance()

	if pctx.eof() || pctx.current() != tokens.DoubleQuote {
		return nil, fmt.Errorf("sfv: expected quote after %% in display string")
	}
	pctx.advance() // consume quote

	var byteArray []byte
	for !pctx.eof() {
		c := pctx.current()
		pctx.advance()

		if c <= 0x1f || c >= 0x7f {
			return nil, fmt.Errorf("sfv: invalid character in display string: %c", c)
		}

		if c == tokens.Percent {
			// Percent-encoded byte
			if pctx.eof() {
				return nil, fmt.Errorf("sfv: unexpected end after %% in display string")
			}
			hex1 := pctx.current()
			pctx.advance()
			if pctx.eof() {
				return nil, fmt.Errorf("sfv: incomplete hex sequence in display string")
			}
			hex2 := pctx.current()
			pctx.advance()

			// Decode hex - ParseUint will validate the hex characters for us
			hexStr := string([]byte{hex1, hex2})
			val, err := strconv.ParseUint(hexStr, 16, 8)
			if err != nil {
				return nil, fmt.Errorf("sfv: invalid hex sequence %%%c%c in display string: %w", hex1, hex2, err)
			}
			byteArray = append(byteArray, byte(val))
		} else if c == tokens.DoubleQuote {
			// End of display string
			// Decode as UTF-8
			return BareDisplayString(string(byteArray)), nil
		} else {
			// Regular ASCII character
			byteArray = append(byteArray, c)
		}
	}
	return nil, fmt.Errorf("sfv: unexpected end of input, expected closing quote in display string")
}
