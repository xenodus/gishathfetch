import { memo, useEffect, useId, useMemo, useState } from "react";
import { ChevronDown, MapPin } from "react-feather";

const SECTION_CLASS_NAME = "store-selector-section google-anno-skip mb-3";

function getSelectionSummary(selectedCount, totalCount) {
  return `${selectedCount} of ${totalCount} stores selected`;
}

const StoreSelector = memo(
  ({
    options,
    selectedStores,
    onToggle,
    onSelectAll,
    onSelectNone,
    onLoadFavourites,
    onSaveFavourites,
    hasFavourites = false,
    favouritesMatchSelection = false,
    collapsible = true,
    collapseOnSearch = false,
    defaultExpanded = true,
  }) => {
    const [isExpanded, setIsExpanded] = useState(defaultExpanded);
    const panelId = useId();

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
      () => getSelectionSummary(selectedCount, totalCount),
      [selectedCount, totalCount],
    );

    const handleToggleSection = () => {
      setIsExpanded((expanded) => !expanded);
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
              <fieldset className="store-selector-bulk-toggle border-0 p-0 m-0">
                <legend className="visually-hidden">
                  Select all, no, or favourite stores
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
                <span
                  className="store-selector-bulk-divider"
                  aria-hidden="true"
                />
                <button
                  type="button"
                  className={`btn btn-sm store-selector-bulk-btn store-selector-fav-btn${
                    favouritesMatchSelection ? " is-active" : ""
                  }`}
                  aria-pressed={favouritesMatchSelection}
                  aria-label="Load favourite stores"
                  disabled={!hasFavourites}
                  onClick={onLoadFavourites}
                >
                  Load Fav. Stores
                </button>
                <button
                  type="button"
                  className="btn btn-sm store-selector-bulk-btn store-selector-fav-btn"
                  aria-label="Save current selection as favourite stores"
                  onClick={onSaveFavourites}
                >
                  Save Fav. Stores
                </button>
              </fieldset>
              <span
                className="store-selector-count text-muted small"
                aria-live="polite"
              >
                {selectedCount} of {totalCount} selected
              </span>
            </div>

            <fieldset className="store-selector-pills border-0 p-0 m-0">
              <legend className="visually-hidden">Local game stores</legend>
              {options.map((store) => {
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
          </div>
        )}
      </div>
    );
  },
);

export default StoreSelector;
