import React from 'react';
import Card from './Card';
import SkeletonCard from './SkeletonCard';

const SearchResults = ({
    results,
    isSearching,
    isCardInCart,
    addToCart,
    removeFromCart,
    onSearchStore,
    baseUrl
}) => {
    return (
        <>
            {results.length > 0 && (
                <div id="resultCount" className="mb-3 text-center bg-warning-subtle text-dark rounded py-2">
                    {results.length} result{results.length > 1 ? "s" : ""} found
                </div>
            )}

            {(results.length > 0 || isSearching) && (
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

                        {isSearching && (
                            <>
                                {[...Array(4)].map((_, i) => (
                                    <SkeletonCard key={`skeleton-${i}`} />
                                ))}
                            </>
                        )}
                    </div>
                </div>
            )}
        </>
    );
};

export default SearchResults;

export default SearchResults;
