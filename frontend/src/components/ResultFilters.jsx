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
  hasActiveFilters,
  onClearFilters,
}) => {
  return (
    <div className="mb-3 text-start result-filters">
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
    </div>
  );
};

export default ResultFilters;
