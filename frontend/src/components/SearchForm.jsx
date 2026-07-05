import { useEffect, useRef, useState } from "react";
import { Loader, X } from "react-feather";
import { MAX_SEARCH_LENGTH, MIN_SEARCH_LENGTH } from "../constants";
import StoreSelector from "./StoreSelector";

const TIP_DISMISSED_STORAGE_KEY = "search-form-tip-dismissed";

const SearchForm = ({
  searchQuery,
  onQueryChange,
  onClearQuery,
  onSearchSubmit,
  suggestions,
  showSuggestions,
  showEmptySuggestions,
  isLoadingSuggestions,
  onSuggestionClick,
  onFocus,
  isSearching,
  searchProgress,
  lgsOptions,
  selectedStores,
  onStoreToggle,
  onSelectAll,
  onSelectNone,
  onLoadFavourites,
  onSaveFavourites,
  hasFavourites,
  favouritesMatchSelection,
  onCloseSuggestions,
  searchError,
  storesWarning,
  onCancelSearch,
  popularSearchesSlot,
}) => {
  const wrapperRef = useRef(null);
  const inputRef = useRef(null);
  const [selectedIndex, setSelectedIndex] = useState(-1);
  const [showTip, setShowTip] = useState(() => {
    if (typeof window === "undefined") {
      return true;
    }

    try {
      return localStorage.getItem(TIP_DISMISSED_STORAGE_KEY) !== "true";
    } catch {
      return true;
    }
  });

  useEffect(() => {
    function handleClickOutside(event) {
      if (wrapperRef.current && !wrapperRef.current.contains(event.target)) {
        onCloseSuggestions?.();
        setSelectedIndex(-1);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [onCloseSuggestions]);

  useEffect(() => {
    if (selectedIndex < 0) {
      return;
    }

    document
      .getElementById(`suggestion-${selectedIndex}`)
      ?.scrollIntoView({ block: "nearest" });
  }, [selectedIndex]);

  useEffect(() => {
    const handleGlobalKeyDown = (event) => {
      if (event.key !== "/" || event.ctrlKey || event.metaKey || event.altKey) {
        return;
      }

      const target = event.target;
      if (target instanceof HTMLElement) {
        const tag = target.tagName;
        if (
          tag === "INPUT" ||
          tag === "TEXTAREA" ||
          tag === "SELECT" ||
          target.isContentEditable
        ) {
          return;
        }
      }

      event.preventDefault();
      inputRef.current?.focus();
    };

    document.addEventListener("keydown", handleGlobalKeyDown);
    return () => {
      document.removeEventListener("keydown", handleGlobalKeyDown);
    };
  }, []);

  const handleInputChange = (event) => {
    setSelectedIndex(-1);
    onQueryChange(event);
  };

  const handleKeyDown = (e) => {
    if (!showSuggestions || suggestions.length === 0) {
      if (e.key === "Escape") {
        onCloseSuggestions?.();
      }
      return;
    }

    switch (e.key) {
      case "ArrowDown":
        e.preventDefault();
        setSelectedIndex((prev) =>
          prev < suggestions.length - 1 ? prev + 1 : prev,
        );
        break;
      case "ArrowUp":
        e.preventDefault();
        setSelectedIndex((prev) => (prev > 0 ? prev - 1 : -1));
        break;
      case "Enter":
        if (selectedIndex >= 0 && selectedIndex < suggestions.length) {
          e.preventDefault();
          onSuggestionClick(suggestions[selectedIndex]);
          setSelectedIndex(-1);
        }
        break;
      case "Escape":
        e.preventDefault();
        onCloseSuggestions();
        setSelectedIndex(-1);
        break;
      default:
        break;
    }
  };

  const handleDismissTip = () => {
    setShowTip(false);

    try {
      localStorage.setItem(TIP_DISMISSED_STORAGE_KEY, "true");
    } catch {
      // Ignore storage access failures and keep in-memory state only.
    }
  };

  const queryTooShort =
    searchQuery.length > 0 && searchQuery.length < MIN_SEARCH_LENGTH;
  const showSuggestionDropdown =
    (showSuggestions && suggestions.length > 0) ||
    isLoadingSuggestions ||
    showEmptySuggestions;
  const showClearButton = searchQuery.length > 0 && !isSearching;

  return (
    <div ref={wrapperRef}>
      <form id="searchForm" onSubmit={onSearchSubmit}>
        <div className="mb-3 position-relative">
          <div className="form-floating search-input-wrapper">
            <input
              ref={inputRef}
              type="text"
              inputMode="search"
              enterKeyHint="search"
              className="form-control search-input"
              id="search"
              role="combobox"
              placeholder="lightning bolt"
              value={searchQuery}
              onChange={handleInputChange}
              onKeyDown={handleKeyDown}
              autoComplete="off"
              onFocus={onFocus}
              maxLength={MAX_SEARCH_LENGTH}
              aria-autocomplete="list"
              aria-controls="suggestions"
              aria-expanded={showSuggestionDropdown}
              aria-activedescendant={
                selectedIndex >= 0 ? `suggestion-${selectedIndex}` : undefined
              }
              aria-describedby={
                queryTooShort || searchError ? "search-error" : undefined
              }
            />
            <label htmlFor="search">Card Name</label>
            <div className="search-input-actions">
              {isLoadingSuggestions && (
                <Loader
                  size={16}
                  className="search-input-spinner"
                  aria-hidden="true"
                />
              )}
              {showClearButton && (
                <button
                  type="button"
                  className="search-input-clear"
                  onClick={() => {
                    onClearQuery();
                    inputRef.current?.focus();
                  }}
                  aria-label="Clear search"
                >
                  <X size={16} aria-hidden="true" />
                </button>
              )}
            </div>
          </div>

          {(queryTooShort || searchError) && (
            <div
              className="form-text text-danger"
              id="search-error"
              role="alert"
            >
              {searchError ||
                `Enter at least ${MIN_SEARCH_LENGTH} characters to search.`}
            </div>
          )}

          {showSuggestionDropdown && (
            <div
              id="suggestions"
              className="suggestions d-block"
              role="listbox"
              aria-label="Card name suggestions"
            >
              {isLoadingSuggestions && (
                <output className="suggestion-status">
                  <Loader size={14} className="search-input-spinner me-2" />
                  Loading suggestions…
                </output>
              )}

              {!isLoadingSuggestions && showEmptySuggestions && (
                <output className="suggestion-status">
                  No matching cards found.
                </output>
              )}

              {!isLoadingSuggestions &&
                showSuggestions &&
                suggestions.map((s, suggestionIndex) => {
                  const escapedQuery = searchQuery.replace(
                    /[.*+?^${}()|[\]\\]/g,
                    "\\$&",
                  );
                  const parts = s.split(new RegExp(`(${escapedQuery})`, "gi"));

                  return (
                    // biome-ignore lint/a11y/useFocusableInteractive: Focus is managed by input
                    // biome-ignore lint/a11y/useKeyWithClickEvents: Keyboard navigation is handled by input
                    <div
                      key={s}
                      id={`suggestion-${suggestionIndex}`}
                      className={`suggestion-item${selectedIndex === suggestionIndex ? " selected" : ""}`}
                      onClick={() => {
                        onSuggestionClick(s);
                        setSelectedIndex(-1);
                      }}
                      role="option"
                      aria-selected={selectedIndex === suggestionIndex}
                    >
                      {parts.map((part, index) =>
                        part.toLowerCase() === searchQuery.toLowerCase() ? (
                          // biome-ignore lint/suspicious/noArrayIndexKey: Order of regex parts is stable
                          <b key={`${suggestionIndex}-${index}`}>{part}</b>
                        ) : (
                          // biome-ignore lint/suspicious/noArrayIndexKey: Order of regex parts is stable
                          <span key={`${suggestionIndex}-${index}`}>
                            {part}
                          </span>
                        ),
                      )}
                    </div>
                  );
                })}
            </div>
          )}
        </div>

        {popularSearchesSlot}

        <StoreSelector
          options={lgsOptions}
          selectedStores={selectedStores}
          onToggle={onStoreToggle}
          onSelectAll={onSelectAll}
          onSelectNone={onSelectNone}
          onLoadFavourites={onLoadFavourites}
          onSaveFavourites={onSaveFavourites}
          hasFavourites={hasFavourites}
          favouritesMatchSelection={favouritesMatchSelection}
          collapsible
          collapseOnSearch={isSearching}
          defaultExpanded={false}
        />

        {showTip && (
          <div
            className="alert bg-info-subtle mb-3 px-2 py-1 small d-flex align-items-center justify-content-between"
            role="note"
          >
            <div>
              <span className="text-info-emphasis me-1 fw-semibold">Tip:</span>
              Selecting fewer stores usually helps GishathFetch finish searching
              faster and keeps operational costs down.
            </div>
            <button
              type="button"
              className="btn-close ms-2"
              aria-label="Dismiss tip"
              onClick={handleDismissTip}
            />
          </div>
        )}

        {storesWarning && (
          <div className="alert alert-warning py-2 px-3 mb-3 small">
            {storesWarning}
          </div>
        )}

        <div className="mb-3 d-grid gap-2">
          {isSearching ? (
            <>
              <button
                id="searchBtn"
                type="button"
                className="btn btn-primary"
                disabled
              >
                {searchProgress}
              </button>
              <button
                type="button"
                className="btn btn-outline-secondary"
                onClick={onCancelSearch}
              >
                Cancel
              </button>
            </>
          ) : (
            <button
              id="searchBtn"
              type="submit"
              className="btn btn-primary"
              disabled={queryTooShort}
            >
              Search
            </button>
          )}
        </div>
      </form>
    </div>
  );
};

export default SearchForm;
