import { Form } from "react-bootstrap";

const ResultFilters = ({
  sortBy,
  onSortChange,
  qualityFilter,
  onQualityFilterChange,
  availableQualities,
  foilOnly,
  onFoilOnlyChange,
  cheapestPerStore,
  onCheapestPerStoreChange,
  availableStores,
  selectedStores,
  onToggleStore,
  onSelectAllStores,
  onSelectNoStores,
  hasActiveFilters,
  onClearFilters,
}) => {
  const allStoresSelected =
    availableStores.length > 0 &&
    selectedStores.length === availableStores.length;
  const noStoresSelected = selectedStores.length === 0;

  return (
    <div className="mb-3 text-start result-filters rounded py-3 px-3">
      <div className="row g-2 align-items-end">
        <div className="col-12 col-md-4">
          <Form.Label htmlFor="result-sort" className="small fw-semibold mb-1">
            Sort by
          </Form.Label>
          <Form.Select
            id="result-sort"
            size="sm"
            value={sortBy}
            onChange={(e) => onSortChange(e.target.value)}
          >
            <option value="price-asc">Price (low to high)</option>
            <option value="price-desc">Price (high to low)</option>
          </Form.Select>
        </div>
        {availableQualities.length > 0 && (
          <div className="col-12 col-md-4">
            <Form.Label
              htmlFor="quality-filter"
              className="small fw-semibold mb-1"
            >
              Condition
            </Form.Label>
            <Form.Select
              id="quality-filter"
              size="sm"
              value={qualityFilter}
              onChange={(e) => onQualityFilterChange(e.target.value)}
            >
              <option value="all">All conditions</option>
              {availableQualities.map((quality) => (
                <option key={quality} value={quality}>
                  {quality}
                </option>
              ))}
            </Form.Select>
          </div>
        )}
        <div className="col-12 col-md-4 d-flex flex-wrap align-items-center gap-3">
          <Form.Check
            type="checkbox"
            id="foil-only"
            label="Foil only"
            checked={foilOnly}
            onChange={(e) => onFoilOnlyChange(e.target.checked)}
            className="mb-0"
          />
          <Form.Check
            type="checkbox"
            id="cheapest-per-store"
            label="Cheapest per store"
            checked={cheapestPerStore}
            onChange={(e) => onCheapestPerStoreChange(e.target.checked)}
            className="mb-0"
          />
          {hasActiveFilters && (
            <button
              type="button"
              className="btn btn-link btn-sm p-0 text-decoration-none"
              onClick={onClearFilters}
            >
              Clear filters
            </button>
          )}
        </div>
      </div>

      {availableStores.length > 0 && (
        <div className="mt-3 pt-3 border-top border-secondary-subtle">
          <div className="d-flex flex-wrap align-items-center justify-content-between gap-2 mb-2">
            <Form.Label className="small fw-semibold mb-0">Stores</Form.Label>
            <div className="d-flex flex-wrap align-items-center gap-2">
              <fieldset className="store-selector-bulk-toggle border-0 p-0 m-0">
                <legend className="visually-hidden">
                  Select all or no stores in results
                </legend>
                <button
                  type="button"
                  className={`btn btn-sm store-selector-bulk-btn${
                    allStoresSelected ? " is-active" : ""
                  }`}
                  aria-pressed={allStoresSelected}
                  onClick={onSelectAllStores}
                >
                  All
                </button>
                <button
                  type="button"
                  className={`btn btn-sm store-selector-bulk-btn${
                    noStoresSelected ? " is-active" : ""
                  }`}
                  aria-pressed={noStoresSelected}
                  onClick={onSelectNoStores}
                >
                  None
                </button>
              </fieldset>
              <span
                className="store-selector-count text-muted small"
                aria-live="polite"
              >
                {selectedStores.length} of {availableStores.length} selected
              </span>
            </div>
          </div>

          <fieldset className="store-selector-pills border-0 p-0 m-0">
            <legend className="visually-hidden">Filter results by store</legend>
            {availableStores.map((store) => {
              const isSelected = selectedStores.includes(store);

              return (
                <button
                  key={store}
                  type="button"
                  className={`btn btn-sm store-selector-pill${
                    isSelected ? " is-selected" : ""
                  }`}
                  aria-pressed={isSelected}
                  onClick={() => onToggleStore(store)}
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
};

export default ResultFilters;
