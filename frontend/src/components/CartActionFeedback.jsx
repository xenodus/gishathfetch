const CartActionFeedback = ({ message }) => {
  if (!message) {
    return null;
  }

  return (
    <output
      className="cart-action-feedback"
      aria-live="polite"
      aria-atomic="true"
    >
      {message}
    </output>
  );
};

export default CartActionFeedback;
