# github.com/lestrrat-go/sfv [![CI](https://github.com/lestrrat-go/sfv/actions/workflows/ci.yml/badge.svg)](https://github.com/lestrrat-go/sfv/actions/workflows/ci.yml) [![Go Reference](https://pkg.go.dev/badge/github.com/lestrrat-go/sfv.svg)](https://pkg.go.dev/github.com/lestrrat-go/sfv) [![codecov.io](https://codecov.io/github/lestrrat-go/sfv/coverage.svg?branch=v1)](https://codecov.io/github/lestrrat-go/sfv?branch=v1)

Go module implementing RFC 9651 Structured Field Values for HTTP.

# Features

* Complete implementation of RFC 9651 Structured Field Values specification
  * Parse and serialize Items, Lists, and Dictionaries
  * Support for all SFV data types: Integers, Decimals, Strings, Tokens, Byte Sequences, Booleans, and Dates
  * Full parameter support for Items and Inner Lists
* Clean, type-safe API

# SYNOPSIS

<!-- INCLUDE(examples/sfv_readme_example_test.go) -->
```go
package examples_test

import (
  "fmt"
  "log"

  "github.com/lestrrat-go/sfv"
)

func Example() {
  // Parse various SFV types
  
  // Parse a simple string item
  item, err := sfv.ParseItem([]byte(`"hello world"`))
  if err != nil {
    log.Fatal(err)
  }
  itemSerialized, _ := sfv.Marshal(item)
  fmt.Printf("Parsed string item: %s\n", string(itemSerialized))

  // Parse a list with mixed types
  list, err := sfv.Parse([]byte(`"text", 42, ?1, @1659578233`))
  if err != nil {
    log.Fatal(err)
  }
  listSerialized, _ := sfv.Marshal(list)
  fmt.Printf("Parsed list: %s\n", string(listSerialized))

  // Parse a dictionary
  dict, err := sfv.ParseDictionary([]byte(`key1="value1", key2=42, flag`))
  if err != nil {
    log.Fatal(err)
  }
  dictSerialized, _ := sfv.Marshal(dict)
  fmt.Printf("Parsed dictionary: %s\n", string(dictSerialized))

  // Create and serialize SFV values programmatically

  // Create a dictionary with various data types
  newDict := sfv.NewDictionary()
  newDict.Set("name", sfv.String("John Doe"))
  newDict.Set("age", sfv.Integer(30))
  newDict.Set("active", sfv.Boolean(true))
  newDict.Set("score", sfv.Decimal(98.5))
  newDict.Set("data", sfv.ByteSequence([]byte("hello")))

  // Serialize the dictionary
  serialized, err := sfv.Marshal(newDict)
  if err != nil {
    log.Fatal(err)
  }
  fmt.Printf("Serialized dictionary: %s\n", string(serialized))

  // Work with parameters on items
  itemWithParams := sfv.String("cached-resource")
  itemWithParams.Parameter("max-age", 3600)
  itemWithParams.Parameter("public", true)

  paramSerialized, err := sfv.Marshal(itemWithParams)
  if err != nil {
    log.Fatal(err)
  }
  fmt.Printf("Item with parameters: %s\n", string(paramSerialized))

  // Create a list with inner lists
  mainList := &sfv.List{}

  // Add simple items
  mainList.Add(sfv.String("item1"))
  mainList.Add(sfv.Integer(42))

  // Add an inner list with parameters
  innerList := &sfv.InnerList{}
  innerList.Add(sfv.String("inner1"))
  innerList.Add(sfv.String("inner2"))
  // Inner lists can have parameters too

  mainList.Add(innerList)

  complexSerialized, err := sfv.Marshal(mainList)
  if err != nil {
    log.Fatal(err)
  }
  fmt.Printf("Complex list: %s\n", string(complexSerialized))

  // OUTPUT:
  // Parsed string item: "hello world"
  // Parsed list: "text", 42, ?1, @1659578233
  // Parsed dictionary: key1="value1", key2=42, flag
  // Serialized dictionary: name="John Doe", age=30, active, score=98.5, data=:aGVsbG8=:
  // Item with parameters: "cached-resource"; max-age=3600; public
  // Complex list: "item1", 42, ("inner1" "inner2")
}
```
source: [examples/sfv_readme_example_test.go](https://github.com/lestrrat-go/sfv/blob/main/examples/sfv_readme_example_test.go)
<!-- END INCLUDE -->

# How-to Documentation

* [API documentation](https://pkg.go.dev/github.com/lestrrat-go/sfv)
* [RFC 9651 Specification](https://tools.ietf.org/html/rfc9651)

# Description

This Go module implements [RFC 9651 Structured Field Values](https://tools.ietf.org/html/rfc9651), a specification for parsing and serializing structured data in HTTP header fields and other contexts.

## Example Use Cases

### HTTP Cache-Control Header
```go
// Parse: "max-age=3600, must-revalidate, private"
dict, _ := sfv.ParseDictionary([]byte(`max-age=3600, must-revalidate, private`))

var maxAge int64
dict.GetValue("max-age", &maxAge) // maxAge = 3600
```

### HTTP Message Signatures
```go
// Create signature parameters for HTTP Message Signatures
params := sfv.NewDictionary()
params.Set("created", sfv.Date(time.Now().Unix()))
params.Set("keyid", sfv.String("key-123"))

// Serialize for Signature-Input header
sigInput, _ := sfv.Marshal(params)
```

# Contributions

## Issues

For bug reports and feature requests, please try to follow the issue templates as much as possible.
For either bug reports or feature requests, failing tests are even better.

## Pull Requests

Please make sure to include tests that exercise the changes you made.

## Discussions / Usage

Please try [discussions](https://github.com/lestrrat-go/sfv/discussions) first.

# License

MIT License. See LICENSE file for details.