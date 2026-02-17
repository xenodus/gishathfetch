import { useState, useEffect } from 'react';

export default function useCart() {
    const [cart, setCart] = useState([]);
    const [showCart, setShowCart] = useState(false);

    // Initial load from LocalStorage
    useEffect(() => {
        const storedCart = localStorage.getItem('cart');
        if (storedCart) {
            try {
                setCart(JSON.parse(storedCart));
            } catch (err) {
                console.error("Failed to parse cart from localStorage:", err);
            }
        }
    }, []);

    const addToCart = (card) => {
        setCart((prev) => {
            const newCart = [card, ...prev];
            localStorage.setItem('cart', JSON.stringify(newCart));
            return newCart;
        });
    };

    const removeFromCart = (index) => {
        setCart((prev) => {
            const newCart = prev.filter((_, i) => i !== index);
            localStorage.setItem('cart', JSON.stringify(newCart));
            return newCart;
        });
    };

    const clearCart = () => {
        setCart([]);
        localStorage.removeItem('cart');
    };

    const isCardInCart = (card) => {
        return cart.some((item) => JSON.stringify(item) === JSON.stringify(card));
    };

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
