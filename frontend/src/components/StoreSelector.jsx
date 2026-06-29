import { memo, useEffect, useId, useMemo, useState } from "react";
import { ChevronDown, MapPin, Search, X } from "react-feather";

const SECTION_CLASS_NAME = "store-selector-section google-anno-skip mb-3";

function getSelectionSummary(selectedStores, totalCount) {
  const selectedCount = selectedStores.length;

  if (selectedCount === 0) {
    return "No stores selected";
  }

  if (selectedCount === totalCount) {
    return `All ${totalCount} stores`;
  }

  if (selectedCount <= 3) {
    return selectedStores.join(", ");
  }

  return `${selectedCount} of ${totalCount} stores`;
}

const StoreSelector = memo(
  ({
    options,
    selectedStores,
    onToggle,
    onSelectAll,
    onSelectNone,
    collapsible = true,
    collapseOnSearch = false,
  }) => {
    const [isExpanded, setIsExpanded] = useState(true);
    const [filterQuery, setFilterQuery] = useState("");
    const panelId = useId();
    const filterInputId = useId();

    useEffect(() => {
      if (collapseOnSearch) {
        setIsExpanded(false);
      }
    }, [collapseOnSearch]);

    const selectedCount = selectedStores.length;
    const totalCount = options.length;
    const allSelected = selectedCount === totalCount;
    const noneSelected = selectedCount === 0;
    const showContent = !collapsible || isExpanded;

    const summaryText = useMemo(
      () => getSelectionSummary(selectedStores, totalCount),
      [selectedStores, totalCount],
    );

    const filteredOptions = useMemo(() => {
      const query = filterQuery.trim().toLowerCase();
      if (!query) {
        return options;
      }

      return options.filter((store) => store.toLowerCase().includes(query));
    }, [filterQuery, options]);

    const handleToggleSection = () => {
      setIsExpanded((expanded) => !expanded);
    };

    const handleClearFilter = () => {
      setFilterQuery("");
    };

    return (
      <div
        className={`${SECTION_CLASS_NAME}${
          showContent ? " is-expanded" : " is-collapsed"
        }`}
      >
        {collapsible ? (
          <button
            type="button"
            className="store-selector-toggle"
            aria-expanded={isExpanded}
            aria-controls={panelId}
            onClick={handleToggleSection}
          >
            <MapPin
              size={15}
              aria-hidden="true"
              className="store-selector-icon"
            />
            <span className="store-selector-title">
              {isExpanded ? "Stores" : summaryText}
            </span>
            {!isExpanded && !allSelected && !noneSelected && (
              <span className="store-selector-badge" aria-hidden="true">
                {selectedCount}
              </span>
            )}
            <ChevronDown
              size={16}
              aria-hidden="true"
              className={`store-selector-chevron${
                isExpanded ? " is-expanded" : ""
              }`}
            />
          </button>
        ) : (
          <div className="store-selector-header-static">
            <MapPin
              size={15}
              aria-hidden="true"
              className="store-selector-icon"
            />
            <span className="store-selector-title">Stores</span>
            <span className="store-selector-count text-muted small">
              {selectedCount} of {totalCount} selected
            </span>
          </div>
        )}

        {showContent && (
          <div className="store-selector-panel" id={panelId}>
            <div className="store-selector-controls">
              <div className="store-selector-bulk-actions">
                <fieldset className="store-selector-bulk-toggle border-0 p-0 m-0">
                  <legend className="visually-hidden">
                    Select all or no stores
                  </legend>
                  <button
                    type="button"
                    className={`btn btn-sm store-selector-bulk-btn${
                      allSelected ? " is-active" : ""
                    }`}
                    aria-pressed={allSelected}
                    onClick={onSelectAll}
                  >
                    All
                  </button>
                  <button
                    type="button"
                    className={`btn btn-sm store-selector-bulk-btn${
                      noneSelected ? " is-active" : ""
                    }`}
                    aria-pressed={noneSelected}
                    onClick={onSelectNone}
                  >
                    None
                  </button>
                </fieldset>
                <span
                  className="store-selector-count text-muted small"
                  aria-live="polite"
                >
                  {selectedCount} of {totalCount} selected
                </span>
              </div>

              <div className="store-selector-filter">
                <label className="visually-hidden" htmlFor={filterInputId}>
                  Filter stores
                </label>
                <div className="store-selector-filter-input-wrap">
                  <Search
                    size={14}
                    aria-hidden="true"
                    className="store-selector-filter-icon"
                  />
                  <input
                    type="search"
                    id={filterInputId}
                    className="form-control form-control-sm store-selector-filter-input"
                    placeholder="Filter stores..."
                    value={filterQuery}
                    onChange={(event) => setFilterQuery(event.target.value)}
                    autoComplete="off"
                  />
                  {filterQuery && (
                    <button
                      type="button"
                      className="btn btn-sm store-selector-filter-clear"
                      aria-label="Clear store filter"
                      onClick={handleClearFilter}
                    >
                      <X size={14} aria-hidden="true" />
                    </button>
                  )}
                </div>
              </div>
            </div>

            {filteredOptions.length > 0 ? (
              <fieldset className="store-selector-pills border-0 p-0 m-0">
                <legend className="visually-hidden">Local game stores</legend>
                {filteredOptions.map((store) => {
                  const isSelected = selectedStores.includes(store);

                  return (
                    <button
                      key={store}
                      type="button"
                      className={`btn btn-sm store-selector-pill${
                        isSelected ? " is-selected" : ""
                      }`}
                      aria-pressed={isSelected}
                      onClick={() => onToggle(store)}
                    >
                      {store}
                    </button>
                  );
                })}
              </fieldset>
            ) : (
              <p className="store-selector-empty small text-muted mb-0">
                No stores match &ldquo;{filterQuery.trim()}&rdquo;.
              </p>
            )}
          </div>
        )}
      </div>
    );
  },
);

export default StoreSelector;
