package affiliatelinks

import "context"

// Store persists Amazon affiliate link entries.
type Store interface {
	ListAll(ctx context.Context) ([]Link, error)
	GetByID(ctx context.Context, id string) (*Link, error)
	Put(ctx context.Context, link Link) error
	Delete(ctx context.Context, id string) error
}
