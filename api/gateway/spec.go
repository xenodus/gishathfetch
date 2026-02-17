package gateway

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
	Search(searchStr string) ([]Card, error)
}
