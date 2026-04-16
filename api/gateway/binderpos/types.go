package binderpos

// CardInfo is the structure of product variant information scraped from Binderpos stores.
type CardInfo struct {
	ID                     int64    `json:"id"`
	Title                  string   `json:"title"`
	Option1                string   `json:"option1"`
	Option2                any      `json:"option2"`
	Option3                any      `json:"option3"`
	Sku                    string   `json:"sku"`
	RequiresShipping       bool     `json:"requires_shipping"`
	Taxable                bool     `json:"taxable"`
	FeaturedImage          any      `json:"featured_image"`
	Available              bool     `json:"available"`
	Name                   string   `json:"name"`
	PublicTitle            string   `json:"public_title"`
	Options                []string `json:"options"`
	Price                  int      `json:"price"`
	Weight                 int      `json:"weight"`
	CompareAtPrice         any      `json:"compare_at_price"`
	InventoryManagement    string   `json:"inventory_management"`
	Barcode                any      `json:"barcode"`
	RequiresSellingPlan    bool     `json:"requires_selling_plan"`
	SellingPlanAllocations []any    `json:"selling_plan_allocations"`
}
