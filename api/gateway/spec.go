package gateway

type Card struct {
	Name       string  `json:"name"`
	Url        string  `json:"url"`
	Img        string  `json:"img"`
	Price      float64 `json:"price"`
	InStock    bool    `json:"inStock"`
	Source     string  `json:"src"`
	Quality    string  `json:"quality"`
	IsScrapped bool    `json:"-"`
}

type LGS interface {
	Search(searchStr string) ([]Card, error)
}
