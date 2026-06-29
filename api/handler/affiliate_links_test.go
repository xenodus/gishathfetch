package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"

	affiliatecontroller "mtg-price-checker-sg/controller/affiliatelinks"
	"mtg-price-checker-sg/pkg/config"
	"mtg-price-checker-sg/store/affiliatelinks"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/require"
)

type stubAffiliateService struct {
	active []affiliatelinks.Link
	all    []affiliatelinks.Link
}

func (s *stubAffiliateService) ListActive(ctx context.Context, platform string) ([]affiliatelinks.Link, error) {
	return s.active, nil
}

func (s *stubAffiliateService) ListAll(ctx context.Context) ([]affiliatelinks.Link, error) {
	return s.all, nil
}

func (s *stubAffiliateService) Create(ctx context.Context, input affiliatecontroller.CreateInput) (affiliatelinks.Link, error) {
	return affiliatelinks.Link{ID: "new", Title: input.Title}, nil
}

func (s *stubAffiliateService) Update(ctx context.Context, id string, input affiliatecontroller.UpdateInput) (affiliatelinks.Link, error) {
	return affiliatelinks.Link{ID: id, Title: input.Title}, nil
}

func (s *stubAffiliateService) Delete(ctx context.Context, id string) error {
	return nil
}

func TestNormalizeAPIPath(t *testing.T) {
	require.Equal(t, "/affiliate-links", normalizeAPIPath("/prod/affiliate-links"))
	require.Equal(t, "/admin/affiliate-links", normalizeAPIPath("/prod/admin/affiliate-links"))
	require.Equal(t, "/admin/affiliate-links/abc", normalizeAPIPath("/admin/affiliate-links/abc"))
}

func TestIsAffiliateAdminAuthorized(t *testing.T) {
	t.Setenv(config.AffiliateAdminAPIKeyEnv, "secret-key")

	require.True(t, isAffiliateAdminAuthorized(map[string]string{
		"authorization": "Bearer secret-key",
	}))
	require.True(t, isAffiliateAdminAuthorized(map[string]string{
		"x-admin-api-key": "secret-key",
	}))
	require.False(t, isAffiliateAdminAuthorized(map[string]string{
		"authorization": "Bearer wrong",
	}))
}

func TestAffiliateLinksPublicGET(t *testing.T) {
	original := newAffiliateService
	defer func() { newAffiliateService = original }()

	newAffiliateService = func(ctx context.Context) (*affiliatecontroller.Service, error) {
		return affiliatecontroller.NewService(&memoryAffiliateStore{
			links: []affiliatelinks.Link{
				{ID: "1", Status: affiliatelinks.StatusActive, ImageURL: "img", Price: "1", Link: "link"},
			},
		}, nil), nil
	}

	resp, err := AffiliateLinks(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodGet,
		Path:       "/affiliate-links",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var payload affiliateLinksResponse
	require.NoError(t, json.Unmarshal([]byte(resp.Body), &payload))
	require.Len(t, payload.Data, 1)
}

func TestAdminAffiliateLinksUnauthorized(t *testing.T) {
	t.Setenv(config.AffiliateAdminAPIKeyEnv, "secret-key")

	resp, err := AdminAffiliateLinks(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodGet,
		Path:       "/admin/affiliate-links",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAdminAffiliateLinksAuthorizedGET(t *testing.T) {
	t.Setenv(config.AffiliateAdminAPIKeyEnv, "secret-key")

	original := newAffiliateService
	defer func() { newAffiliateService = original }()

	newAffiliateService = func(ctx context.Context) (*affiliatecontroller.Service, error) {
		return affiliatecontroller.NewService(&memoryAffiliateStore{
			links: []affiliatelinks.Link{
				{ID: "1", Title: "Deck Box", Status: affiliatelinks.StatusActive, ImageURL: "img", Price: "1", Link: "link"},
			},
		}, nil), nil
	}

	resp, err := AdminAffiliateLinks(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodGet,
		Path:       "/admin/affiliate-links",
		Headers: map[string]string{
			"x-admin-api-key": "secret-key",
		},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

type memoryAffiliateStore struct {
	links []affiliatelinks.Link
}

func (m *memoryAffiliateStore) ListAll(ctx context.Context) ([]affiliatelinks.Link, error) {
	return m.links, nil
}

func (m *memoryAffiliateStore) GetByID(ctx context.Context, id string) (*affiliatelinks.Link, error) {
	for _, link := range m.links {
		if link.ID == id {
			copy := link
			return &copy, nil
		}
	}
	return nil, nil
}

func (m *memoryAffiliateStore) Put(ctx context.Context, link affiliatelinks.Link) error {
	for i, existing := range m.links {
		if existing.ID == link.ID {
			m.links[i] = link
			return nil
		}
	}
	m.links = append(m.links, link)
	return nil
}

func (m *memoryAffiliateStore) Delete(ctx context.Context, id string) error {
	filtered := m.links[:0]
	for _, link := range m.links {
		if link.ID != id {
			filtered = append(filtered, link)
		}
	}
	m.links = filtered
	return nil
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
