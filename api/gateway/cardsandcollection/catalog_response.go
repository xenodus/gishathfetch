package cardsandcollection

// catalogSearchResponse unmarshals only fields needed for in-stock MTG singles.
type catalogSearchResponse struct {
	Hits struct {
		Hits []struct {
			ID     string `json:"_id"`
			Source struct {
				Name           string `json:"name"`
				Img            string `json:"img"`
				SetName        string `json:"setName"`
				QuantityOnSale any    `json:"quantityOnSale"`
				MinPriceSale   any    `json:"minPriceSale"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}