import React from 'react';
import Card from './Card';

const SearchResults = ({
    results,
    isCardInCart,
    addToCart,
    removeFromCart,
    onSearchStore,
    baseUrl
}) => {
    if (results.length === 0) return null;

    return (
        <>
            <div id="resultCount" className="mb-3 text-center bg-warning-subtle text-dark rounded py-2">
                {results.length} result{results.length > 1 ? "s" : ""} found
            </div>
            <div id="result" className="mb-3 text-center">
                <div className="row">
                    {results.map((card, i) => (
                        <Card
                            key={i}
                            card={card}
                            index={i}
                            isCardInCart={isCardInCart}
                            addToCart={addToCart}
                            removeFromCart={removeFromCart}
                            onSearchStore={onSearchStore}
                            baseUrl={baseUrl}
                        />
                    ))}
                </div>
            </div>
        </>
    );
};

export default SearchResults;
