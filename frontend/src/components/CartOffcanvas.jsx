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
  onClearCart,
  baseUrl,
}) => {
  const [sortOption, setSortOption] = useState("default");

  const renderCardsWithAds = (cards) => {
    const elements = [];
    for (let i = 0; i < cards.length; i++) {
      const card = cards[i];
      elements.push(
        <Card
          key={`card-${i}`}
          card={card}
          index={i}
          isCart={true}
          isCardInCart={isCardInCart}
          removeFromCart={removeFromCart}
          onSearchStore={onSearchStore}
          baseUrl={baseUrl}
        />,
      );

      const isEndOfGroup = (i + 1) % 4 === 0;
      const hasMoreCards = i + 1 < cards.length;
      if (isEndOfGroup && hasMoreCards) {
        elements.push(
          <div key={`ad-${i}`} className="col-12">
            <AdComponent variant="responsive" />
          </div>,
        );
      }
    }
    return elements;
  };

  const handleClearCart = () => {
    if (window.confirm("Are you sure you want to remove all saved cards?")) {
      onClearCart();
    }
  };

  const groupedCart = useMemo(() => {
    if (sortOption === "default") return null;

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
              <option value="store">Store</option>
            </Form.Select>
          </Form.Group>
        )}

        {cart.length > 0 ? (
          <>
            {sortOption === "default" && (
              <div className="row">{renderCardsWithAds(cart)}</div>
            )}

            {sortOption === "store" &&
              Object.entries(groupedCart).map(([storeName, data]) => (
                <div key={storeName} className="mb-4">
                  <h5 className="border-bottom pb-2 mb-3">
                    {storeName} - S$ {data.total.toFixed(2)}
                  </h5>
                  <div className="row">
                    {(() => {
                      const elements = [];
                      for (let i = 0; i < data.cards.length; i++) {
                        const card = data.cards[i];
                        elements.push(
                          <Card
                            key={`card-${storeName}-${card.originalIndex}`}
                            card={card}
                            index={card.originalIndex}
                            isCart={true}
                            isCardInCart={isCardInCart}
                            removeFromCart={removeFromCart}
                            onSearchStore={onSearchStore}
                            baseUrl={baseUrl}
                          />,
                        );

                        const isEndOfGroup = (i + 1) % 4 === 0;
                        const hasMoreCards = i + 1 < data.cards.length;
                        if (isEndOfGroup && hasMoreCards) {
                          elements.push(
                            <div
                              key={`ad-${storeName}-${i}`}
                              className="col-12"
                            >
                              <AdComponent variant="responsive" />
                            </div>,
                          );
                        }
                      }
                      return elements;
                    })()}
                  </div>
                </div>
              ))}
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
