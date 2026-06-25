import { Form } from "react-bootstrap";

const ResultFilters = ({
  sortBy,
  onSortChange,
  foilOnly,
  onFoilOnlyChange,
  qualityFilter,
  onQualityFilterChange,
  availableQualities,
  priceMin,
  onPriceMinChange,
  priceMax,
  onPriceMaxChange,
  availableStores,
  isStoreSelected,
  onToggleStore,
  hasActiveFilters,
  onClearFilters,
}) => {
  return (
    <div className="mb-3 text-start result-filters">
      <div className="row g-2 align-items-end mb-2">
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
            <option value="store-asc">Store name</option>
          </Form.Select>
        </div>
        <div className="col-6 col-md-2">
          <Form.Label htmlFor="price-min" className="small fw-semibold mb-1">
            Min price
          </Form.Label>
          <Form.Control
            id="price-min"
            type="number"
            size="sm"
            min="0"
            step="0.01"
            placeholder="S$"
            value={priceMin}
            onChange={(e) => onPriceMinChange(e.target.value)}
          />
        </div>
        <div className="col-6 col-md-2">
          <Form.Label htmlFor="price-max" className="small fw-semibold mb-1">
            Max price
          </Form.Label>
          <Form.Control
            id="price-max"
            type="number"
            size="sm"
            min="0"
            step="0.01"
            placeholder="S$"
            value={priceMax}
            onChange={(e) => onPriceMaxChange(e.target.value)}
          />
        </div>
        <div className="col-12 col-md-4 d-flex flex-wrap align-items-center gap-3">
          <Form.Check
            type="checkbox"
            id="foil-only"
            label="Foil only"
            checked={foilOnly}
            onChange={(e) => onFoilOnlyChange(e.target.checked)}
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

      {availableQualities.length > 0 && (
        <div className="mb-2">
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

      {availableStores.length > 1 && (
        <div>
          <div className="small fw-semibold mb-1">Stores</div>
          <div className="d-flex flex-wrap gap-1">
            {availableStores.map((store) => {
              const selected = isStoreSelected(store);
              return (
                <button
                  key={store}
                  type="button"
                  className={`btn btn-sm ${selected ? "btn-primary" : "btn-outline-secondary"}`}
                  onClick={() => onToggleStore(store)}
                  aria-pressed={selected}
                >
                  {store}
                </button>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
};

export default ResultFilters;
