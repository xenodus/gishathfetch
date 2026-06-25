import { useCallback, useState } from "react";

const stripSavedAt = (item) => {
  const { savedAt: _savedAt, ...card } = item;
  return card;
};

const cardsMatch = (a, b) =>
  JSON.stringify(stripSavedAt(a)) === JSON.stringify(stripSavedAt(b));

const formatSavedAt = (savedAt) => {
  if (!savedAt) return null;

  const savedDate = new Date(savedAt);
  const now = new Date();
  const startOfToday = new Date(
    now.getFullYear(),
    now.getMonth(),
    now.getDate(),
  );
  const startOfSavedDay = new Date(
    savedDate.getFullYear(),
    savedDate.getMonth(),
    savedDate.getDate(),
  );
  const dayDiff = Math.floor(
    (startOfToday - startOfSavedDay) / (1000 * 60 * 60 * 24),
  );

  if (dayDiff === 0) return "Saved today";
  if (dayDiff === 1) return "Saved yesterday";
  if (dayDiff < 7) return `Saved ${dayDiff} days ago`;

  return `Saved on ${savedDate.toLocaleDateString()}`;
};

export { formatSavedAt };

export default function useCart() {
  const [cart, setCart] = useState(() => {
    const storedCart = localStorage.getItem("cart");
    if (storedCart) {
      try {
        return JSON.parse(storedCart);
      } catch (err) {
        console.error("Failed to parse cart from localStorage:", err);
        return [];
      }
    }
    return [];
  });
  const [showCart, setShowCart] = useState(false);

  const addToCart = useCallback((card) => {
    setCart((prev) => {
      const exists = prev.some((item) => cardsMatch(item, card));

      if (exists) return prev;

      const newCart = [{ ...card, savedAt: Date.now() }, ...prev];
      try {
        localStorage.setItem("cart", JSON.stringify(newCart));
      } catch (err) {
        console.error("Failed to save cart to localStorage:", err);
      }
      return newCart;
    });
  }, []);

  const removeFromCart = useCallback((index) => {
    setCart((prev) => {
      const newCart = prev.filter((_, i) => i !== index);
      try {
        localStorage.setItem("cart", JSON.stringify(newCart));
      } catch (err) {
        console.error("Failed to save cart to localStorage:", err);
      }
      return newCart;
    });
  }, []);

  const clearCart = useCallback(() => {
    setCart([]);
    try {
      localStorage.removeItem("cart");
    } catch (err) {
      console.error("Failed to clear cart from localStorage:", err);
    }
  }, []);

  const removeFromCartByCard = useCallback((card) => {
    setCart((prev) => {
      const newCart = prev.filter((item) => !cardsMatch(item, card));
      if (newCart.length === prev.length) return prev;

      try {
        localStorage.setItem("cart", JSON.stringify(newCart));
      } catch (err) {
        console.error("Failed to save cart to localStorage:", err);
      }
      return newCart;
    });
  }, []);

  const isCardInCart = useCallback(
    (card) => {
      return cart.some((item) => cardsMatch(item, card));
    },
    [cart],
  );

  return {
    cart,
    showCart,
    setShowCart,
    addToCart,
    removeFromCart,
    removeFromCartByCard,
    clearCart,
    isCardInCart,
  };
}
