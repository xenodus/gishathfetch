package fivemana

const unavailableSuggestJSON = `{
  "resources": {
    "results": {
      "products": [
        {
          "available": false,
          "title": "The Ten Rings [Marvel Super Heroes]",
          "price": "9.90",
          "image": "https://cdn.shopify.com/example.png",
          "url": "/products/the-ten-rings-marvel-super-heroes"
        }
      ]
    }
  }
}`

const inStockSuggestJSON = `{
  "resources": {
    "results": {
      "products": [
        {
          "available": true,
          "title": "Abrade [Foundations]",
          "price": "0.40",
          "image": "https://cdn.shopify.com/abrade.png",
          "url": "/products/abrade-foundations",
          "tags": ["Foundations", "Foundations Non-Foil", "Red", "Uncommon"]
        }
      ]
    }
  }
}`

const foilSuggestJSON = `{
  "resources": {
    "results": {
      "products": [
        {
          "available": true,
          "title": "Rhystic Study [Foil]",
          "price": "120.00",
          "image": "https://cdn.shopify.com/foil.png",
          "url": "/products/rhystic-study-foil",
          "tags": ["Foil"]
        }
      ]
    }
  }
}`
