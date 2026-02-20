import { Button, Offcanvas } from "react-bootstrap";
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
  const handleClearCart = () => {
    if (window.confirm("Are you sure you want to remove all saved cards?")) {
      onClearCart();
    }
  };

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
        {cart.length > 0 ? (
          <>
            <div className="row">
              {cart.map((card, i) => (
                <Card
                  // biome-ignore lint/suspicious/noArrayIndexKey: Cart items do not have unique IDs
                  key={i}
                  card={card}
                  index={i}
                  isCart={true}
                  isCardInCart={isCardInCart}
                  removeFromCart={removeFromCart}
                  onSearchStore={onSearchStore}
                  baseUrl={baseUrl}
                />
              ))}
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
