package affiliatelinks

import (
	"context"
	"testing"
	"time"

	store "mtg-price-checker-sg/store/affiliatelinks"

	"github.com/stretchr/testify/require"
)

type memoryStore struct {
	links map[string]store.Link
}

func (m *memoryStore) ListAll(ctx context.Context) ([]store.Link, error) {
	links := make([]store.Link, 0, len(m.links))
	for _, link := range m.links {
		links = append(links, link)
	}
	return links, nil
}

func (m *memoryStore) GetByID(ctx context.Context, id string) (*store.Link, error) {
	link, ok := m.links[id]
	if !ok {
		return nil, nil
	}
	return &link, nil
}

func (m *memoryStore) Put(ctx context.Context, link store.Link) error {
	m.links[link.ID] = link
	return nil
}

func (m *memoryStore) Delete(ctx context.Context, id string) error {
	delete(m.links, id)
	return nil
}

func TestServiceListActiveFiltersExpiredAndInactive(t *testing.T) {
	fixedNow := time.Date(2026, 6, 29, 12, 0, 0, 0, time.UTC)
	service := NewService(&memoryStore{
		links: map[string]store.Link{
			"1": {ID: "1", Platform: store.PlatformAmazon, Status: store.StatusActive, ExpiryDate: "2099-01-01", ImageURL: "img", Price: "1", Link: "link"},
			"2": {ID: "2", Platform: store.PlatformAmazon, Status: store.StatusInactive, ExpiryDate: "2099-01-01", ImageURL: "img", Price: "1", Link: "link"},
			"3": {ID: "3", Platform: store.PlatformAmazon, Status: store.StatusActive, ExpiryDate: "2026-06-28", ImageURL: "img", Price: "1", Link: "link"},
		},
	}, nil)
	service.now = func() time.Time { return fixedNow }

	links, err := service.ListActive(context.Background(), "")
	require.NoError(t, err)
	require.Len(t, links, 1)
	require.Equal(t, "1", links[0].ID)
}

func TestServiceCreateRequiresFields(t *testing.T) {
	service := NewService(&memoryStore{links: map[string]store.Link{}}, nil)
	service.now = func() time.Time { return time.Date(2026, 6, 29, 0, 0, 0, 0, time.UTC) }
	service.newID = func() string { return "abc123" }

	_, err := service.Create(context.Background(), CreateInput{
		Price: "S$10",
		Link:  "https://amazon.sg/example",
	})
	require.Error(t, err)

	link, err := service.Create(context.Background(), CreateInput{
		Platform: store.PlatformShopee,
		ImageURL: "https://example.com/image.jpg",
		Price:    "S$10",
		Link:     "https://shopee.sg/example",
		Title:    "Deck Box",
		Status:   store.StatusActive,
	})
	require.NoError(t, err)
	require.Equal(t, "abc123", link.ID)
	require.Equal(t, "Deck Box", link.Title)
	require.Equal(t, store.PlatformShopee, link.Platform)
}

func TestServiceListActiveFiltersByPlatform(t *testing.T) {
	fixedNow := time.Date(2026, 6, 29, 12, 0, 0, 0, time.UTC)
	service := NewService(&memoryStore{
		links: map[string]store.Link{
			"1": {ID: "1", Platform: store.PlatformAmazon, Status: store.StatusActive, ImageURL: "img", Price: "1", Link: "link"},
			"2": {ID: "2", Platform: store.PlatformShopee, Status: store.StatusActive, ImageURL: "img", Price: "1", Link: "link"},
		},
	}, nil)
	service.now = func() time.Time { return fixedNow }

	links, err := service.ListActive(context.Background(), store.PlatformShopee)
	require.NoError(t, err)
	require.Len(t, links, 1)
	require.Equal(t, store.PlatformShopee, links[0].Platform)
}
