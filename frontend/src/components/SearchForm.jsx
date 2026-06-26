import { useEffect, useRef, useState } from "react";
import { MAX_SEARCH_LENGTH, MIN_SEARCH_LENGTH } from "../constants";
import StoreSelector from "./StoreSelector";

const TIP_DISMISSED_STORAGE_KEY = "search-form-tip-dismissed";

const SearchForm = ({
  searchQuery,
  onQueryChange,
  onSearchSubmit,
  suggestions,
  showSuggestions,
  onSuggestionClick,
  onFocus,
  isSearching,
  searchProgress,
  lgsOptions,
  selectedStores,
  onStoreToggle,
  onSelectAll,
  onSelectNone,
  onCloseSuggestions,
  searchError,
  storesWarning,
  onCancelSearch,
}) => {
  const wrapperRef = useRef(null);
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

  const handleKeyDown = (e) => {
    if (!showSuggestions || suggestions.length === 0) {
      // Reset index when suggestions are not shown
      if (selectedIndex !== -1) {
        setSelectedIndex(-1);
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

  return (
    <div ref={wrapperRef}>
      <form id="searchForm" onSubmit={onSearchSubmit}>
        <div className="mb-3 position-relative">
          <div className="form-floating">
            <input
              type="search"
              className="form-control"
              id="search"
              role="combobox"
              placeholder="lightning bolt"
              value={searchQuery}
              onChange={onQueryChange}
              onKeyDown={handleKeyDown}
              autoComplete="off"
              onFocus={onFocus}
              maxLength={MAX_SEARCH_LENGTH}
              aria-autocomplete="list"
              aria-controls="suggestions"
              aria-expanded={showSuggestions && suggestions.length > 0}
              aria-activedescendant={
                selectedIndex >= 0 ? `suggestion-${selectedIndex}` : undefined
              }
            />
            <label htmlFor="search">Card Name</label>
          </div>

          {(queryTooShort || searchError) && (
            <div className="form-text text-danger" role="alert">
              {searchError ||
                `Enter at least ${MIN_SEARCH_LENGTH} characters to search.`}
            </div>
          )}

          {showSuggestions && suggestions.length > 0 && (
            <div
              id="suggestions"
              className="suggestions d-block"
              role="listbox"
              aria-label="Card name suggestions"
            >
              {suggestions.map((s, suggestionIndex) => {
                // Escape query for regex and split suggestion into parts
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
                        <span key={`${suggestionIndex}-${index}`}>{part}</span>
                      ),
                    )}
                  </div>
                );
              })}
            </div>
          )}
        </div>

        <StoreSelector
          options={lgsOptions}
          selectedStores={selectedStores}
          onToggle={onStoreToggle}
          onSelectAll={onSelectAll}
          onSelectNone={onSelectNone}
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

        <div className="mb-3 d-grid">
          {isSearching ? (
            <div className="btn-group">
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
            </div>
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
