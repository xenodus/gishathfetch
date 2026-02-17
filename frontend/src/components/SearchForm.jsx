import React from 'react';
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
    onSelectNone
}) => {
    return (
        <div>
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
                            {suggestions.map((s, i) => {
                                const boldedSuggestion = s.replace(
                                    new RegExp(searchQuery, 'gi'),
                                    (match) => `<b>${match}</b>`
                                );
                                return (
                                    <div
                                        key={i}
                                        className="suggestion-item"
                                        onClick={() => onSuggestionClick(s)}
                                        dangerouslySetInnerHTML={{ __html: boldedSuggestion }}
                                    />
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
