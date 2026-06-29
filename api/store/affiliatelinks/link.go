package affiliatelinks

import (
	"strings"
	"time"
)

const (
	StatusActive   = "active"
	StatusInactive = "inactive"

	PlatformAmazon = "amazon"
	PlatformShopee = "shopee"
)

// Link is a curated affiliate product entry.
type Link struct {
	ID         string `json:"id" dynamodbav:"id"`
	Platform   string `json:"platform" dynamodbav:"platform"`
	Title      string `json:"title,omitempty" dynamodbav:"title,omitempty"`
	ImageURL   string `json:"imageUrl" dynamodbav:"imageUrl"`
	Price      string `json:"price" dynamodbav:"price"`
	Link       string `json:"link" dynamodbav:"link"`
	ExpiryDate string `json:"expiryDate,omitempty" dynamodbav:"expiryDate,omitempty"`
	Status     string `json:"status" dynamodbav:"status"`
	CreatedAt  string `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt  string `json:"updatedAt" dynamodbav:"updatedAt"`
}

// IsSupportedPlatform reports whether platform is a known affiliate provider.
func IsSupportedPlatform(platform string) bool {
	switch strings.ToLower(strings.TrimSpace(platform)) {
	case PlatformAmazon, PlatformShopee:
		return true
	default:
		return false
	}
}

// IsActive reports whether the link should be shown on the public site.
func (l Link) IsActive(now time.Time) bool {
	if l.Status != StatusActive {
		return false
	}
	if l.ExpiryDate == "" {
		return true
	}
	expiry, err := time.Parse("2006-01-02", l.ExpiryDate)
	if err != nil {
		return false
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return !expiry.Before(today)
}
