package util

import "strings"

var qualityByKey = map[string]string{
	"NM":     "Near Mint",
	"NM/M":   "Near Mint",
	"LP":     "Lightly Played",
	"MP":     "Moderately Played",
	"HP":     "Heavily Played",
	"DM":     "Damaged",
	"EX/EX+": "Excellent",
}

func normalizedQualityKey(quality string) string {
	return strings.ToUpper(strings.TrimSpace(quality))
}

func MapQuality(quality string) string {
	if mapped, ok := qualityByKey[normalizedQualityKey(quality)]; ok {
		return mapped
	}
	return quality
}

func IsKnownQuality(quality string) bool {
	_, ok := qualityByKey[normalizedQualityKey(quality)]
	return ok
}
