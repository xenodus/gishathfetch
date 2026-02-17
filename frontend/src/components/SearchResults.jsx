import React from 'react';
import Card from './Card';
import SkeletonCard from './SkeletonCard';
import AdComponent from './AdComponent';

const SearchResults = ({
    results,
    isSearching,
    hasSearched,
    isCardInCart,
    addToCart,
    removeFromCart,
    onSearchStore,
    baseUrl
}) => {
    return (
        <>
            {hasSearched && !isSearching && (
                <div id="resultCount" className="mb-3 text-center bg-warning-subtle text-dark rounded py-2">
                    {results?.length || 0} result{(results?.length !== 1) ? "s" : ""} found
                </div>
            )}

            {((results && results.length > 0) || isSearching) && (
                <div id="result" className="mb-3 text-center">
                    <div className="row">
                        {results.map((card, i) => {
                            const showAd = results.length > 8 && ((i + 1) % 8 === 0) && (i + 1 !== results.length);
                            return (
                                <React.Fragment key={i}>
                                    <Card
                                        card={card}
                                        index={i}
                                        isCardInCart={isCardInCart}
                                        addToCart={addToCart}
                                        removeFromCart={removeFromCart}
                                        onSearchStore={onSearchStore}
                                        baseUrl={baseUrl}
                                    />
                                    {showAd && <AdComponent />}
                                </React.Fragment>
                            );
                        })}

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
