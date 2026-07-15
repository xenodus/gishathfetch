package fivemana

const soldOutProductHTML = `
<ul class="grid product-grid">
<li class="grid__item">
<div class="card-wrapper product-card-wrapper">
  <div class="card card--standard card--media">
    <div class="card__inner">
      <div class="card__media">
        <img src="//5-mana.sg/cdn/shop/files/ten-rings.png" alt="The Ten Rings [Marvel Super Heroes]">
      </div>
      <div class="card__content">
        <div class="card__badge bottom left"><span class="badge badge--bottom-left">Sold out</span></div>
      </div>
    </div>
    <div class="card__content">
      <h3 class="card__heading h5">
        <a href="/products/the-ten-rings-marvel-super-heroes">The Ten Rings [Marvel Super Heroes]</a>
      </h3>
      <div class="price price--sold-out">
        <div class="price__sale">
          <span class="price-item price-item--sale price-item--last">$9.90 SGD</span>
        </div>
      </div>
    </div>
  </div>
</div>
</li>
</ul>
`

const inStockProductHTML = `
<ul class="grid product-grid">
<li class="grid__item">
<div class="card-wrapper product-card-wrapper">
  <div class="card card--standard card--media">
    <div class="card__inner">
      <div class="card__media">
        <img src="//5-mana.sg/cdn/shop/files/abrade.png" alt="Abrade [Foundations]">
      </div>
    </div>
    <div class="card__content">
      <h3 class="card__heading h5">
        <a href="/products/abrade-foundations">Abrade [Foundations]</a>
      </h3>
      <div class="price">
        <div class="price__sale">
          <span class="price-item price-item--sale price-item--last">$0.40 SGD</span>
        </div>
      </div>
    </div>
  </div>
</div>
</li>
</ul>
`

const suggestResponseFixture = `{
  "resources": {
    "results": {
      "products": [
        {
          "available": true,
          "handle": "abrade-foundations",
          "image": "https://cdn.shopify.com/files/abrade.png",
          "price": "0.40",
          "tags": ["Foundations", "Foundations Non-Foil", "Red", "Uncommon"],
          "title": "Abrade [Foundations]",
          "url": "/products/abrade-foundations?_pos=2"
        },
        {
          "available": false,
          "handle": "the-ten-rings-marvel-super-heroes",
          "image": "https://cdn.shopify.com/files/ten-rings.png",
          "price": "9.90",
          "tags": ["Marvel Super Heroes"],
          "title": "The Ten Rings [Marvel Super Heroes]",
          "url": "/products/the-ten-rings-marvel-super-heroes"
        },
        {
          "available": true,
          "handle": "lightning-bolt-foil",
          "image": "https://cdn.shopify.com/files/bolt.png",
          "price": "12.00",
          "tags": ["Foil", "Red"],
          "title": "Lightning Bolt [Alpha] [Foil]",
          "url": "/products/lightning-bolt-foil"
        }
      ]
    }
  }
}`
