package util

import "strings"

func MapQuality(quality string) string {
	switch strings.ToUpper(strings.TrimSpace(quality)) {
	case "NM", "NM/M":
		return "Near Mint"
	case "LP":
		return "Lightly Played"
	case "MP":
		return "Moderately Played"
	case "HP":
		return "Heavily Played"
	case "DM":
		return "Damaged"
	case "EX/EX+":
		return "Excellent"
	default:
		return quality
	}
}
