import { useMemo, useState } from "react";
import { Button, Form, Offcanvas } from "react-bootstrap";
import AdComponent from "./AdComponent";
import Card from "./Card";

const CartOffcanvas = ({
  show,
  onHide,
  cart,
  isCardInCart,
  removeFromCart,
  onSearchStore,
  onSearchWithFavouriteStores,
  hasFavourites,
  onClearCart,
  baseUrl,
}) => {
  const [sortOption, setSortOption] = useState("default");

  const handleClearCart = () => {
    if (window.confirm("Are you sure you want to remove all saved cards?")) {
      onClearCart();
    }
  };

  const displayedCart = useMemo(() => {
    const withIndex = cart.map((card, index) => ({
      ...card,
      originalIndex: index,
    }));

    if (sortOption === "name-asc") {
      return [...withIndex].sort((a, b) =>
        (a.name || "").localeCompare(b.name || "", undefined, {
          sensitivity: "base",
        }),
      );
    }

    if (sortOption === "name-desc") {
      return [...withIndex].sort((a, b) =>
        (b.name || "").localeCompare(a.name || "", undefined, {
          sensitivity: "base",
        }),
      );
    }

    return withIndex;
  }, [cart, sortOption]);

  const groupedCart = useMemo(() => {
    if (sortOption !== "store") return null;

    const groups = {};
    cart.forEach((card, index) => {
      const storeName = card.src || "Unknown Store";
      if (!groups[storeName]) {
        groups[storeName] = { cards: [], total: 0 };
      }
      groups[storeName].cards.push({ ...card, originalIndex: index });
      groups[storeName].total += card.price || 0;
    });
    return groups;
  }, [cart, sortOption]);

  return (
    <Offcanvas show={show} onHide={onHide} placement="end">
      <Offcanvas.Header closeButton>
        <Offcanvas.Title>Saved Cards</Offcanvas.Title>
      </Offcanvas.Header>
      <Offcanvas.Body>
        <div className="mb-3 small text-muted">
          When a card is saved, a snapshot of it from that point in time is
          taken. If there is any change in its price or availability, it will
          not be updated automatically.
        </div>

        {cart.length > 0 && (
          <Form.Group className="mb-4">
            <Form.Label className="small fw-bold text-uppercase mb-1">
              Sort By
            </Form.Label>
            <Form.Select
              value={sortOption}
              onChange={(e) => setSortOption(e.target.value)}
              size="sm"
            >
              <option value="default">Saved Order</option>
              <option value="name-asc">Card Name Asc</option>
              <option value="name-desc">Card Name Desc</option>
              <option value="store">Store</option>
            </Form.Select>
          </Form.Group>
        )}

        {cart.length > 0 ? (
          <>
            {(sortOption === "default" ||
              sortOption === "name-asc" ||
              sortOption === "name-desc") && (
              <div className="row">
                {displayedCart.map((card) => (
                  <Card
                    key={card.originalIndex}
                    card={card}
                    index={card.originalIndex}
                    isCart={true}
                    isCardInCart={isCardInCart}
                    removeFromCart={removeFromCart}
                    onSearchStore={onSearchStore}
                    onSearchWithFavouriteStores={onSearchWithFavouriteStores}
                    hasFavourites={hasFavourites}
                    baseUrl={baseUrl}
                  />
                ))}
              </div>
            )}

            {sortOption === "store" &&
              Object.entries(groupedCart).map(([storeName, data]) => (
                <div key={storeName} className="mb-4">
                  <h5 className="border-bottom pb-2 mb-3">
                    {storeName} - S$ {data.total.toFixed(2)}
                  </h5>
                  <div className="row">
                    {data.cards.map((card) => (
                      <Card
                        key={card.originalIndex}
                        card={card}
                        index={card.originalIndex}
                        isCart={true}
                        isCardInCart={isCardInCart}
                        removeFromCart={removeFromCart}
                        onSearchStore={onSearchStore}
                        onSearchWithFavouriteStores={
                          onSearchWithFavouriteStores
                        }
                        hasFavourites={hasFavourites}
                        baseUrl={baseUrl}
                      />
                    ))}
                  </div>
                </div>
              ))}

            <div className="mt-4">
              <AdComponent lazyLoad />
            </div>

            {cart.length >= 2 && (
              <div className="mt-5">
                <Button
                  variant="danger"
                  className="w-100 text-uppercase"
                  onClick={handleClearCart}
                >
                  Remove all saved cards
                </Button>
              </div>
            )}
          </>
        ) : (
          <strong>No cards saved yet.</strong>
        )}
      </Offcanvas.Body>
    </Offcanvas>
  );
};

export default CartOffcanvas;
