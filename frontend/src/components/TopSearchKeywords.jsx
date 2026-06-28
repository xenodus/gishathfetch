import { useState } from "react";

const LOADING_SKELETON_KEYS = [
  "top-search-keyword-skeleton-a",
  "top-search-keyword-skeleton-b",
  "top-search-keyword-skeleton-c",
  "top-search-keyword-skeleton-d",
  "top-search-keyword-skeleton-e",
];

const PERIOD_OPTIONS = [
  { id: "last24Hours", label: "24 hours" },
  { id: "last30Days", label: "30 days" },
];

function hasAnyKeywords(keywordsByPeriod) {
  return PERIOD_OPTIONS.some(
    (option) => (keywordsByPeriod?.[option.id]?.length ?? 0) > 0,
  );
}

function PeriodToggle({ period, onPeriodChange, disabled }) {
  return (
    <div className="popular-search-period-toggle">
      {PERIOD_OPTIONS.map((option) => (
        <button
          key={option.id}
          type="button"
          className={`btn btn-sm popular-search-period-btn${
            period === option.id ? " is-active" : ""
          }`}
          disabled={disabled}
          aria-pressed={period === option.id}
          onClick={() => onPeriodChange(option.id)}
        >
          {option.label}
        </button>
      ))}
    </div>
  );
}

function PopularSearchHeader({ period, onPeriodChange, disabled }) {
  return (
    <fieldset className="popular-search-header border-0 p-0 m-0 mb-2">
      <div className="d-inline-flex align-items-center justify-content-center gap-2 flex-wrap">
        <legend className="popular-search-legend small text-muted mb-0">
          Popular searches:
        </legend>
        <PeriodToggle
          period={period}
          onPeriodChange={onPeriodChange}
          disabled={disabled}
        />
      </div>
    </fieldset>
  );
}

export default function TopSearchKeywords({
  keywordsByPeriod,
  isLoading,
  onKeywordClick,
  disabled = false,
}) {
  const [period, setPeriod] = useState("last24Hours");

  if (!isLoading && !hasAnyKeywords(keywordsByPeriod)) {
    return null;
  }

  const keywords = keywordsByPeriod?.[period] ?? [];
  const selectedPeriodLabel =
    PERIOD_OPTIONS.find((option) => option.id === period)?.label ?? "";

  if (isLoading) {
    return (
      <div className="mb-3 text-center">
        <PopularSearchHeader
          period={period}
          onPeriodChange={setPeriod}
          disabled
        />
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

  return (
    <div className="mb-3 text-center">
      <PopularSearchHeader
        period={period}
        onPeriodChange={setPeriod}
        disabled={disabled}
      />
      {keywords.length > 0 ? (
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
      ) : (
        <div className="small text-muted">
          No popular searches in the last {selectedPeriodLabel}.
        </div>
      )}
    </div>
  );
}
