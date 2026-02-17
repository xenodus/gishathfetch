import React, { useRef, useEffect, useState } from 'react';
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
    const [selectedIndex, setSelectedIndex] = useState(-1);

    useEffect(() => {
        function handleClickOutside(event) {
            if (wrapperRef.current && !wrapperRef.current.contains(event.target)) {
                onCloseSuggestions && onCloseSuggestions();
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
            case 'ArrowDown':
                e.preventDefault();
                setSelectedIndex(prev =>
                    prev < suggestions.length - 1 ? prev + 1 : prev
                );
                break;
            case 'ArrowUp':
                e.preventDefault();
                setSelectedIndex(prev => prev > 0 ? prev - 1 : -1);
                break;
            case 'Enter':
                if (selectedIndex >= 0 && selectedIndex < suggestions.length) {
                    e.preventDefault();
                    onSuggestionClick(suggestions[selectedIndex]);
                    setSelectedIndex(-1);
                }
                break;
            case 'Escape':
                e.preventDefault();
                onCloseSuggestions();
                setSelectedIndex(-1);
                break;
            default:
                break;
        }
    };

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
                            onKeyDown={handleKeyDown}
                            autoComplete="off"
                            onFocus={onFocus}
                            aria-autocomplete="list"
                            aria-controls="suggestions"
                            aria-expanded={showSuggestions && suggestions.length > 0}
                        />
                        <label htmlFor="search">Card Name</label>
                    </div>

                    {showSuggestions && suggestions.length > 0 && (
                        <div
                            id="suggestions"
                            className="suggestions d-block"
                            role="listbox"
                            aria-label="Card name suggestions"
                        >
                            {suggestions.map((s, suggestionIndex) => {
                                // Escape query for regex and split suggestion into parts
                                const escapedQuery = searchQuery.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
                                const parts = s.split(new RegExp(`(${escapedQuery})`, 'gi'));

                                return (
                                    <div
                                        key={s}
                                        className={`suggestion-item${selectedIndex === suggestionIndex ? ' selected' : ''}`}
                                        onClick={() => {
                                            onSuggestionClick(s);
                                            setSelectedIndex(-1);
                                        }}
                                        role="option"
                                        aria-selected={selectedIndex === suggestionIndex}
                                    >
                                        {parts.map((part, index) =>
                                            part.toLowerCase() === searchQuery.toLowerCase() ? (
                                                <b key={`${suggestionIndex}-${index}`}>{part}</b>
                                            ) : (
                                                <span key={`${suggestionIndex}-${index}`}>{part}</span>
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
