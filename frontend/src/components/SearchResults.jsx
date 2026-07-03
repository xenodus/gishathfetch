import React, { useEffect, useMemo, useRef } from "react";
import { DESKTOP_MIN_WIDTH_MEDIA_QUERY } from "../constants";
import useMediaQuery from "../hooks/useMediaQuery";
import useResultFilters from "../hooks/useResultFilters";
import AdComponent from "./AdComponent";
import Card from "./Card";
import CardKingdomPrice from "./CardKingdomPrice";
import ResultFilters from "./ResultFilters";
import SkeletonCard from "./SkeletonCard";
import StoreErrorsBanner from "./StoreErrorsBanner";

// Display ad after every N results (fewer on mobile where cards are 2-wide)
const AD_DISPLAY_INTERVAL_DESKTOP = 8;
const AD_DISPLAY_INTERVAL_MOBILE = 4;
const MAX_IN_FEED_ADS_DESKTOP = 3;
const MAX_IN_FEED_ADS_MOBILE = 2;

const EmptySearchState = () => (
  <div className="mb-3 text-center py-4 px-3">
    <div className="fw-semibold mb-2">No results found</div>
    <p className="small text-muted mb-0">
      Try picking a card from the auto-suggest, using the full card name, or
      selecting fewer stores for a faster, more accurate search.
    </p>
  </div>
);

const EmptyFilteredState = ({ onClearFilters }) => (
  <div className="mb-3 text-center py-4 px-3">
    <div className="fw-semibold mb-2">No results match your filters</div>
    <p className="small text-muted mb-3">
      Try a different condition, turning off foil only or cheapest per store, or
      clear your filters.
    </p>
    <button
      type="button"
      className="btn btn-outline-primary btn-sm"
      onClick={onClearFilters}
    >
      Clear filters
    </button>
  </div>
);

const SearchResults = ({
  results,
  searchQuery,
  isSearching,
  hasSearched,
  searchError,
  searchStoreErrors,
  onDismissStoreErrors,
  onRetrySearch,
  isCardInCart,
  addToCart,
  removeFromCart,
  removeFromCartByCard,
  onSearchStore,
  cardKingdomPrice,
  baseUrl,
}) => {
  const {
    filteredResults,
    sortBy,
    setSortBy,
    qualityFilter,
    setQualityFilter,
    availableQualities,
    foilOnly,
    setFoilOnly,
    cheapestPerStore,
    setCheapestPerStore,
    hasActiveFilters,
    clearFilters,
  } = useResultFilters(results, searchQuery);

  const isDesktop = useMediaQuery(DESKTOP_MIN_WIDTH_MEDIA_QUERY);
  const adDisplayInterval = isDesktop
    ? AD_DISPLAY_INTERVAL_DESKTOP
    : AD_DISPLAY_INTERVAL_MOBILE;
  const maxInFeedAds = isDesktop
    ? MAX_IN_FEED_ADS_DESKTOP
    : MAX_IN_FEED_ADS_MOBILE;

  const inFeedAdSlotIndices = useMemo(() => {
    if (filteredResults.length <= adDisplayInterval) {
      return new Set();
    }

    const slots = [];
    for (let i = 0; i < filteredResults.length; i++) {
      const position = i + 1;
      if (
        position % adDisplayInterval !== 0 ||
        position === filteredResults.length
      ) {
        continue;
      }
      if (slots.length >= maxInFeedAds) break;
      slots.push(i);
    }
    return new Set(slots);
  }, [filteredResults.length, adDisplayInterval, maxInFeedAds]);

  const resultsAnchorRef = useRef(null);
  const wasSearchingRef = useRef(false);

  useEffect(() => {
    if (wasSearchingRef.current && !isSearching && hasSearched) {
      resultsAnchorRef.current?.scrollIntoView({
        behavior: "smooth",
        block: "start",
      });
    }
    wasSearchingRef.current = isSearching;
  }, [isSearching, hasSearched]);

  const resultCountLabel = hasActiveFilters
    ? `Showing ${filteredResults.length} of ${results.length} result${results.length !== 1 ? "s" : ""}`
    : `${results.length} result${results.length !== 1 ? "s" : ""} found`;

  return (
    <>
      <div ref={resultsAnchorRef} className="scroll-margin-top" />
      {hasSearched &&
        !isSearching &&
        (searchError ? (
          <div
            className="mb-3 text-center bg-danger-subtle text-dark rounded py-3 px-3"
            role="alert"
            aria-live="assertive"
          >
            <strong>Error:</strong> {searchError}
            {onRetrySearch && (
              <div className="mt-2">
                <button
                  type="button"
                  className="btn btn-outline-danger btn-sm"
                  onClick={onRetrySearch}
                >
                  Try again
                </button>
              </div>
            )}
          </div>
        ) : results.length === 0 && searchStoreErrors.length === 0 ? (
          <>
            <CardKingdomPrice price={cardKingdomPrice} />
            <EmptySearchState />
          </>
        ) : (
          <>
            {searchStoreErrors.length > 0 && (
              <StoreErrorsBanner
                storeErrors={searchStoreErrors}
                onDismiss={onDismissStoreErrors}
              />
            )}

            {results.length === 0 ? (
              <>
                <CardKingdomPrice price={cardKingdomPrice} />
                <EmptySearchState />
              </>
            ) : (
              <>
                <div
                  id="resultCount"
                  className="mb-3 text-center bg-warning-subtle text-warning-emphasis rounded py-2"
                  aria-live="polite"
                >
                  {resultCountLabel}
                </div>

                <CardKingdomPrice price={cardKingdomPrice} />

                <ResultFilters
                  sortBy={sortBy}
                  onSortChange={setSortBy}
                  qualityFilter={qualityFilter}
                  onQualityFilterChange={setQualityFilter}
                  availableQualities={availableQualities}
                  foilOnly={foilOnly}
                  onFoilOnlyChange={setFoilOnly}
                  cheapestPerStore={cheapestPerStore}
                  onCheapestPerStoreChange={setCheapestPerStore}
                  hasActiveFilters={hasActiveFilters}
                  onClearFilters={clearFilters}
                />

                {filteredResults.length === 0 ? (
                  <EmptyFilteredState onClearFilters={clearFilters} />
                ) : (
                  <div id="result" className="mb-3 text-center">
                    <div className="row">
                      {filteredResults.map((card, i) => (
                        <React.Fragment
                          key={`${card.src}-${card.url}-${card.price}-${card.quality}`}
                        >
                          <Card
                            card={card}
                            index={i}
                            isCardInCart={isCardInCart}
                            addToCart={addToCart}
                            removeFromCart={removeFromCart}
                            removeFromCartByCard={removeFromCartByCard}
                            onSearchStore={onSearchStore}
                            baseUrl={baseUrl}
                          />
                          {inFeedAdSlotIndices.has(i) && (
                            <div className="col-12 mb-4">
                              <AdComponent lazyLoad />
                            </div>
                          )}
                        </React.Fragment>
                      ))}
                    </div>
                  </div>
                )}
              </>
            )}
          </>
        ))}

      {isSearching && (
        <div id="result" className="mb-3 text-center">
          <div className="row">
            {[...Array(4)].map((_, i) => (
              // biome-ignore lint/suspicious/noArrayIndexKey: Skeleton loaders are static
              <SkeletonCard key={`skeleton-${i}`} />
            ))}
          </div>
        </div>
      )}
    </>
  );
};

export default SearchResults;
