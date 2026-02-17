import { useState, useCallback } from 'react';

export default function useCart() {
    const [cart, setCart] = useState(() => {
        const storedCart = localStorage.getItem('cart');
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
            const exists = prev.some((item) => JSON.stringify(item) === JSON.stringify(card));

            if (exists) return prev;

            const newCart = [card, ...prev];
            try {
                localStorage.setItem('cart', JSON.stringify(newCart));
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
                localStorage.setItem('cart', JSON.stringify(newCart));
            } catch (err) {
                console.error("Failed to save cart to localStorage:", err);
            }
            return newCart;
        });
    }, []);

    const clearCart = useCallback(() => {
        setCart([]);
        try {
            localStorage.removeItem('cart');
        } catch (err) {
            console.error("Failed to clear cart from localStorage:", err);
        }
    }, []);

    const isCardInCart = useCallback((card) => {
        return cart.some((item) => JSON.stringify(item) === JSON.stringify(card));
    }, [cart]);

    return {
        cart,
        showCart,
        setShowCart,
        addToCart,
        removeFromCart,
        clearCart,
        isCardInCart
    };
}
