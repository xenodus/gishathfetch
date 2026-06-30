import { useCallback, useState } from "react";
import { cardsExactMatch, dedupeCartItems } from "../utils/cardIdentity";

const loadCartFromStorage = () => {
  const storedCart = localStorage.getItem("cart");
  if (!storedCart) {
    return [];
  }

  try {
    const parsed = JSON.parse(storedCart);
    if (!Array.isArray(parsed)) {
      return [];
    }

    const deduped = dedupeCartItems(parsed);
    if (deduped.length !== parsed.length) {
      localStorage.setItem("cart", JSON.stringify(deduped));
    }
    return deduped;
  } catch (err) {
    console.error("Failed to parse cart from localStorage:", err);
    return [];
  }
};

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
  const [cart, setCart] = useState(loadCartFromStorage);
  const [showCart, setShowCart] = useState(false);

  const addToCart = useCallback((card) => {
    setCart((prev) => {
      const withoutExactMatch = prev.filter(
        (item) => !cardsExactMatch(item, card),
      );
      const newCart = [{ ...card, savedAt: Date.now() }, ...withoutExactMatch];
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
      const newCart = prev.filter((item) => !cardsExactMatch(item, card));
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
      return cart.some((item) => cardsExactMatch(item, card));
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
