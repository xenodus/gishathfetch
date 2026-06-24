package shopifysuggest

import (
	"fmt"
	"net/url"
)

const (
	fyendalMTGSingleProductType = "MTG Single Cards"
)

// BinderposQueryValues returns the suggest query parameters used by BinderPOS
// storefront search UIs: plain q=, unavailable products last, and a minimal
// field projection.
func BinderposQueryValues(searchStr string) url.Values {
	return url.Values{
		"q": {searchStr},
		"resources[type]": {"product"},
		"resources[options][unavailable_products]": {"last"},
		"resources[options][fields]":               {"title,variants.title,product_type"},
	}
}

// FyendalQueryValues scopes predictive search to MTG single cards and caps
// results at Shopify's 10-product limit.
func FyendalQueryValues(searchStr string) url.Values {
	return url.Values{
		"q":                {searchStr},
		"resources[type]":  {"product"},
		"resources[limit]": {predictiveSearchLimit},
	}
}

// PlainQuery builds the q= parameter unchanged.
func PlainQuery(searchStr string) string {
	return searchStr
}

// FyendalQuery scopes the search to Fyendal Hobby's MTG single product type.
func FyendalQuery(searchStr string) string {
	return fmt.Sprintf("product_type:%q AND %s", fyendalMTGSingleProductType, searchStr)
}
