import { useState } from "react";
import {
  BASE_URL,
  DESKTOP_MIN_WIDTH_MEDIA_QUERY,
  TOP_SEARCH_KEYWORDS_DISPLAY_LIMIT,
  TOP_SEARCH_KEYWORDS_MOBILE_DISPLAY_LIMIT,
} from "../constants";
import useMediaQuery from "../hooks/useMediaQuery";
import { buildSearchQueryUrl } from "../utils/searchUrl";

const LOADING_SKELETON_KEYS = [
  "top-search-keyword-skeleton-a",
  "top-search-keyword-skeleton-b",
  "top-search-keyword-skeleton-c",
  "top-search-keyword-skeleton-d",
  "top-search-keyword-skeleton-e",
  "top-search-keyword-skeleton-f",
  "top-search-keyword-skeleton-g",
  "top-search-keyword-skeleton-h",
  "top-search-keyword-skeleton-i",
  "top-search-keyword-skeleton-j",
];

const PERIOD_OPTIONS = [
  { id: "last24Hours", label: "24 hours" },
  { id: "last30Days", label: "30 days" },
];

// Stable selector for AdSense "Excluded areas" and google-anno-skip for ad intents.
const SECTION_CLASS_NAME =
  "popular-searches-section google-anno-skip mb-3";

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
      <div className="d-flex align-items-center justify-content-start gap-2 flex-wrap">
        <legend className="popular-search-legend small mb-0">
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

function getDisplayLimit(isDesktop) {
  return isDesktop
    ? TOP_SEARCH_KEYWORDS_DISPLAY_LIMIT
    : TOP_SEARCH_KEYWORDS_MOBILE_DISPLAY_LIMIT;
}

export default function TopSearchKeywords({
  keywordsByPeriod,
  isLoading,
  collapsible = false,
}) {
  const [period, setPeriod] = useState("last24Hours");
  const [isExpanded, setIsExpanded] = useState(!collapsible);
  const isDesktop = useMediaQuery(DESKTOP_MIN_WIDTH_MEDIA_QUERY);
  const displayLimit = getDisplayLimit(isDesktop);
  const panelId = "popular-searches-panel";

  if (!isLoading && !hasAnyKeywords(keywordsByPeriod)) {
    return null;
  }

  const keywords = (keywordsByPeriod?.[period] ?? []).slice(0, displayLimit);
  const selectedPeriodLabel =
    PERIOD_OPTIONS.find((option) => option.id === period)?.label ?? "";

  if (collapsible && !isExpanded) {
    return (
      <div className={`${SECTION_CLASS_NAME} popular-searches-collapsed`}>
        <button
          type="button"
          className="btn btn-link btn-sm p-0 text-decoration-none popular-searches-toggle"
          aria-expanded="false"
          aria-controls={panelId}
          onClick={() => setIsExpanded(true)}
        >
          Show popular searches
        </button>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className={SECTION_CLASS_NAME} id={panelId}>
        {collapsible && (
          <div className="mb-2">
            <button
              type="button"
              className="btn btn-link btn-sm p-0 text-decoration-none popular-searches-toggle"
              aria-expanded="true"
              aria-controls={panelId}
              onClick={() => setIsExpanded(false)}
            >
              Hide popular searches
            </button>
          </div>
        )}
        <PopularSearchHeader
          period={period}
          onPeriodChange={setPeriod}
          disabled
        />
        <div className="d-flex flex-wrap justify-content-start gap-2">
          {LOADING_SKELETON_KEYS.slice(0, displayLimit).map((key) => (
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
    <div className={SECTION_CLASS_NAME} id={panelId}>
      {collapsible && (
        <div className="mb-2">
          <button
            type="button"
            className="btn btn-link btn-sm p-0 text-decoration-none popular-searches-toggle"
            aria-expanded="true"
            aria-controls={panelId}
            onClick={() => setIsExpanded(false)}
          >
            Hide popular searches
          </button>
        </div>
      )}
      <PopularSearchHeader period={period} onPeriodChange={setPeriod} />
      {keywords.length > 0 ? (
        <div className="d-flex flex-wrap justify-content-start gap-2">
          {keywords.map((keyword) => (
            <a
              key={keyword}
              href={buildSearchQueryUrl(BASE_URL, keyword)}
              className="btn btn-sm popular-search-pill text-decoration-none"
              aria-label={`Search for ${keyword}`}
            >
              {keyword}
            </a>
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
