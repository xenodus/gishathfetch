import React, { useRef, useEffect } from 'react';
import StoreSelector from './StoreSelector';

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
    onCloseSuggestions
}) => {
    const wrapperRef = useRef(null);

    useEffect(() => {
        function handleClickOutside(event) {
            if (wrapperRef.current && !wrapperRef.current.contains(event.target)) {
                onCloseSuggestions && onCloseSuggestions();
            }
        }
        document.addEventListener("mousedown", handleClickOutside);
        return () => {
            document.removeEventListener("mousedown", handleClickOutside);
        };
    }, [onCloseSuggestions]);

    return (
        <div ref={wrapperRef}>
            <form id="searchForm" onSubmit={onSearchSubmit}>
                <div className="mb-3 position-relative">
                    <div className="form-floating">
                        <input
                            type="search"
                            className="form-control"
                            id="search"
                            placeholder="lightning bolt"
                            value={searchQuery}
                            onChange={onQueryChange}
                            autoComplete="off"
                            onFocus={onFocus}
                        />
                        <label htmlFor="search">Card Name</label>
                    </div>

                    {showSuggestions && suggestions.length > 0 && (
                        <div id="suggestions" className="suggestions d-block">
                            {suggestions.map((s) => {
                                // Escape query for regex and split suggestion into parts
                                const escapedQuery = searchQuery.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
                                const parts = s.split(new RegExp(`(${escapedQuery})`, 'gi'));

                                return (
                                    <div
                                        key={s}
                                        className="suggestion-item"
                                        onClick={() => onSuggestionClick(s)}
                                    >
                                        {parts.map((part, index) =>
                                            part.toLowerCase() === searchQuery.toLowerCase() ? (
                                                <b key={`${index}-${part}`}>{part}</b>
                                            ) : (
                                                <span key={`${index}-${part}`}>{part}</span>
                                            )
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

                <div className="mb-3 d-grid">
                    <button
                        id="searchBtn"
                        type="submit"
                        className="btn btn-primary"
                        disabled={isSearching}
                    >
                        {isSearching ? searchProgress : "Search"}
                    </button>
                </div>
            </form>
        </div>
    );
};

export default SearchForm;
