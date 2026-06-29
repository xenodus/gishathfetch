package affiliatelinks

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	store "mtg-price-checker-sg/store/affiliatelinks"
)

var nowFunc = time.Now

// Service coordinates affiliate link persistence and image uploads.
type Service struct {
	store  store.Store
	images *store.ImageUploader
	now    func() time.Time
	newID  func() string
}

func NewService(linkStore store.Store, imageUploader *store.ImageUploader) *Service {
	return &Service{
		store:  linkStore,
		images: imageUploader,
		now:    nowFunc,
		newID:  newLinkID,
	}
}

// CreateInput is the payload for creating a new affiliate link.
type CreateInput struct {
	Title            string `json:"title"`
	ImageURL         string `json:"imageUrl"`
	ImageData        string `json:"imageData"`
	ImageContentType string `json:"imageContentType"`
	Price            string `json:"price"`
	Link             string `json:"link"`
	ExpiryDate       string `json:"expiryDate"`
	Status           string `json:"status"`
}

// UpdateInput is the payload for updating an existing affiliate link.
type UpdateInput struct {
	Title            string `json:"title"`
	ImageURL         string `json:"imageUrl"`
	ImageData        string `json:"imageData"`
	ImageContentType string `json:"imageContentType"`
	Price            string `json:"price"`
	Link             string `json:"link"`
	ExpiryDate       string `json:"expiryDate"`
	Status           string `json:"status"`
}

func (s *Service) ListAll(ctx context.Context) ([]store.Link, error) {
	return s.store.ListAll(ctx)
}

func (s *Service) ListActive(ctx context.Context) ([]store.Link, error) {
	links, err := s.store.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	now := s.now().UTC()
	active := make([]store.Link, 0, len(links))
	for _, link := range links {
		if link.IsActive(now) {
			active = append(active, link)
		}
	}
	return active, nil
}

func (s *Service) Create(ctx context.Context, input CreateInput) (store.Link, error) {
	imageURL, err := s.resolveImageURL(ctx, strings.TrimSpace(input.ImageURL), input.ImageData, input.ImageContentType)
	if err != nil {
		return store.Link{}, err
	}

	link, err := s.buildLink("", input.Title, imageURL, input.Price, input.Link, input.ExpiryDate, input.Status, "", "")
	if err != nil {
		return store.Link{}, err
	}

	if err := s.store.Put(ctx, link); err != nil {
		return store.Link{}, err
	}
	return link, nil
}

func (s *Service) Update(ctx context.Context, id string, input UpdateInput) (store.Link, error) {
	existing, err := s.store.GetByID(ctx, id)
	if err != nil {
		return store.Link{}, err
	}
	if existing == nil {
		return store.Link{}, fmt.Errorf("affiliate link not found")
	}

	imageURL := strings.TrimSpace(input.ImageURL)
	if input.ImageData != "" {
		imageURL, err = s.resolveImageURL(ctx, "", input.ImageData, input.ImageContentType)
		if err != nil {
			return store.Link{}, err
		}
	} else if imageURL == "" {
		imageURL = existing.ImageURL
	}

	link, err := s.buildLink(id, input.Title, imageURL, input.Price, input.Link, input.ExpiryDate, input.Status, existing.CreatedAt, existing.UpdatedAt)
	if err != nil {
		return store.Link{}, err
	}

	if err := s.store.Put(ctx, link); err != nil {
		return store.Link{}, err
	}
	return link, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	existing, err := s.store.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("affiliate link not found")
	}
	return s.store.Delete(ctx, id)
}

func (s *Service) buildLink(id, title, imageURL, price, linkURL, expiryDate, status, createdAt, updatedAt string) (store.Link, error) {
	title = strings.TrimSpace(title)
	imageURL = strings.TrimSpace(imageURL)
	price = strings.TrimSpace(price)
	linkURL = strings.TrimSpace(linkURL)
	expiryDate = strings.TrimSpace(expiryDate)
	status = normalizeStatus(status)

	if imageURL == "" {
		return store.Link{}, fmt.Errorf("image is required")
	}
	if price == "" {
		return store.Link{}, fmt.Errorf("price is required")
	}
	if linkURL == "" {
		return store.Link{}, fmt.Errorf("link is required")
	}
	if err := validateExpiryDate(expiryDate); err != nil {
		return store.Link{}, err
	}

	now := s.now().UTC().Format(time.RFC3339)
	if id == "" {
		id = s.newID()
		createdAt = now
	}
	if createdAt == "" {
		createdAt = now
	}

	return store.Link{
		ID:         id,
		Title:      title,
		ImageURL:   imageURL,
		Price:      price,
		Link:       linkURL,
		ExpiryDate: expiryDate,
		Status:     status,
		CreatedAt:  createdAt,
		UpdatedAt:  now,
	}, nil
}

func (s *Service) resolveImageURL(ctx context.Context, imageURL, imageData, contentType string) (string, error) {
	imageURL = strings.TrimSpace(imageURL)
	imageData = strings.TrimSpace(imageData)
	contentType = strings.TrimSpace(contentType)

	if imageData != "" {
		if s.images == nil {
			return "", fmt.Errorf("image upload is not configured")
		}
		raw, err := base64.StdEncoding.DecodeString(imageData)
		if err != nil {
			return "", fmt.Errorf("invalid image data")
		}
		return s.images.Upload(ctx, raw, contentType)
	}
	if imageURL == "" {
		return "", fmt.Errorf("image is required")
	}
	return imageURL, nil
}

func normalizeStatus(status string) string {
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "" {
		return store.StatusActive
	}
	if status != store.StatusActive && status != store.StatusInactive {
		return store.StatusActive
	}
	return status
}

func validateExpiryDate(expiryDate string) error {
	if expiryDate == "" {
		return nil
	}
	_, err := time.Parse("2006-01-02", expiryDate)
	if err != nil {
		return fmt.Errorf("expiry date must use YYYY-MM-DD format")
	}
	return nil
}

func newLinkID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("link-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}
