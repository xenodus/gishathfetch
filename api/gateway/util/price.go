package util

import (
	"strconv"
	"strings"
)

func ParsePrice(price string) (float64, error) {
	priceStr := strings.TrimSpace(price)
	priceStr = strings.Replace(priceStr, "From", "", -1)
	priceStr = strings.Replace(priceStr, "S$", "", -1)
	priceStr = strings.Replace(priceStr, "$", "", -1)
	priceStr = strings.Replace(priceStr, ",", "", -1)
	priceStr = strings.Replace(priceStr, "SGD", "", -1)
	priceStr = strings.TrimSpace(priceStr)
	return strconv.ParseFloat(priceStr, 64)
}
