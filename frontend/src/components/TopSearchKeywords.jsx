const LOADING_SKELETON_KEYS = [
  "top-search-keyword-skeleton-a",
  "top-search-keyword-skeleton-b",
  "top-search-keyword-skeleton-c",
  "top-search-keyword-skeleton-d",
  "top-search-keyword-skeleton-e",
];

export default function TopSearchKeywords({
  keywords,
  isLoading,
  onKeywordClick,
  disabled = false,
}) {
  if (isLoading) {
    return (
      <div className="mb-3 text-center">
        <div className="small text-muted mb-2">
          Popular searches (last 24 hours)
        </div>
        <div className="d-flex flex-wrap justify-content-center gap-2">
          {LOADING_SKELETON_KEYS.map((key) => (
            <span
              key={key}
              className="placeholder col-3 rounded-pill"
              style={{ height: "31px" }}
            />
          ))}
        </div>
      </div>
    );
  }

  if (keywords.length === 0) {
    return null;
  }

  return (
    <div className="mb-3 text-center">
      <div className="small text-muted mb-2">
        Popular searches (last 24 hours)
      </div>
      <div className="d-flex flex-wrap justify-content-center gap-2">
        {keywords.map((keyword) => (
          <button
            key={keyword}
            type="button"
            className="btn btn-outline-secondary btn-sm rounded-pill"
            disabled={disabled}
            aria-label={`Search for ${keyword}`}
            onClick={() => onKeywordClick(keyword)}
          >
            {keyword}
          </button>
        ))}
      </div>
    </div>
  );
}
