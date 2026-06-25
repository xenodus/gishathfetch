import { Form } from "react-bootstrap";

const ResultFilters = ({
  sortBy,
  onSortChange,
  foilOnly,
  onFoilOnlyChange,
  hasActiveFilters,
  onClearFilters,
}) => {
  return (
    <div className="mb-3 text-start result-filters">
      <div className="row g-2 align-items-end">
        <div className="col-12 col-md-6">
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
        <div className="col-12 col-md-6 d-flex flex-wrap align-items-center gap-3">
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
    </div>
  );
};

export default ResultFilters;
