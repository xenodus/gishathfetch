import { useCallback, useEffect, useRef, useState } from "react";
import { cardsExactMatch, dedupeCartItems } from "../utils/cardIdentity";

const CART_FEEDBACK_DURATION_MS = 2500;

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
  const [cartActionFeedback, setCartActionFeedback] = useState(null);
  const feedbackTimeoutRef = useRef(null);

  const showCartActionFeedback = useCallback((message) => {
    if (feedbackTimeoutRef.current) {
      clearTimeout(feedbackTimeoutRef.current);
    }

    setCartActionFeedback(message);
    feedbackTimeoutRef.current = setTimeout(() => {
      setCartActionFeedback(null);
      feedbackTimeoutRef.current = null;
    }, CART_FEEDBACK_DURATION_MS);
  }, []);

  useEffect(() => {
    return () => {
      if (feedbackTimeoutRef.current) {
        clearTimeout(feedbackTimeoutRef.current);
      }
    };
  }, []);

  const addToCart = useCallback(
    (card) => {
      let feedbackMessage = "Card saved";

      setCart((prev) => {
        const wasInCart = prev.some((item) => cardsExactMatch(item, card));
        if (wasInCart) {
          feedbackMessage = "Card updated";
        }

        const withoutExactMatch = prev.filter(
          (item) => !cardsExactMatch(item, card),
        );
        const newCart = [
          { ...card, savedAt: Date.now() },
          ...withoutExactMatch,
        ];
        try {
          localStorage.setItem("cart", JSON.stringify(newCart));
        } catch (err) {
          console.error("Failed to save cart to localStorage:", err);
        }
        return newCart;
      });

      showCartActionFeedback(feedbackMessage);
    },
    [showCartActionFeedback],
  );

  const removeFromCart = useCallback(
    (index) => {
      setCart((prev) => {
        const newCart = prev.filter((_, i) => i !== index);
        try {
          localStorage.setItem("cart", JSON.stringify(newCart));
        } catch (err) {
          console.error("Failed to save cart to localStorage:", err);
        }
        return newCart;
      });

      showCartActionFeedback("Card removed");
    },
    [showCartActionFeedback],
  );

  const clearCart = useCallback(() => {
    setCart([]);
    try {
      localStorage.removeItem("cart");
    } catch (err) {
      console.error("Failed to clear cart from localStorage:", err);
    }

    showCartActionFeedback("All saved cards removed");
  }, [showCartActionFeedback]);

  const removeFromCartByCard = useCallback(
    (card) => {
      let removed = false;

      setCart((prev) => {
        const newCart = prev.filter((item) => !cardsExactMatch(item, card));
        if (newCart.length === prev.length) return prev;

        removed = true;
        try {
          localStorage.setItem("cart", JSON.stringify(newCart));
        } catch (err) {
          console.error("Failed to save cart to localStorage:", err);
        }
        return newCart;
      });

      if (removed) {
        showCartActionFeedback("Card removed");
      }
    },
    [showCartActionFeedback],
  );

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
    cartActionFeedback,
    addToCart,
    removeFromCart,
    removeFromCartByCard,
    clearCart,
    isCardInCart,
  };
}
