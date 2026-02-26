package gateway

import "context"

type Card struct {
	Name      string
	Url       string
	Img       string
	Price     float64
	InStock   bool
	IsFoil    bool
	Source    string
	Quality   string
	ExtraInfo []string
}

type LGS interface {
	Search(ctx context.Context, searchStr string) ([]Card, error)
}
