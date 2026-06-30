import { useEffect, useId, useState } from "react";
import { ChevronDown, TrendingUp } from "react-feather";
import {
  BASE_URL,
  DESKTOP_MIN_WIDTH_MEDIA_QUERY,
  TOP_SEARCH_KEYWORDS_DISPLAY_LIMIT,
  TOP_SEARCH_KEYWORDS_MOBILE_DISPLAY_LIMIT,
} from "../constants";
import useMediaQuery from "../hooks/useMediaQuery";
import { buildPopularSearchUrl } from "../utils/searchUrl";

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
  { id: "last6Months", label: "6 months" },
  { id: "last1Year", label: "1 year" },
];

// Stable selector for AdSense "Excluded areas" and google-anno-skip for ad intents.
const SECTION_CLASS_NAME = "popular-searches-section google-anno-skip mb-3";

function hasAnyKeywords(keywordsByPeriod) {
  return PERIOD_OPTIONS.some(
    (option) => (keywordsByPeriod?.[option.id]?.length ?? 0) > 0,
  );
}

function PeriodToggle({ period, onPeriodChange, disabled }) {
  return (
    <fieldset className="popular-search-period-toggle border-0 p-0 m-0">
      <legend className="visually-hidden">Popular search time range</legend>
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
    </fieldset>
  );
}

function PopularSearchesToggle({ isExpanded, collapsible, panelId, onToggle }) {
  const label = isExpanded ? "Popular searches" : "Show popular searches";

  if (!collapsible) {
    return (
      <div className="popular-searches-header-static">
        <TrendingUp
          size={15}
          aria-hidden="true"
          className="popular-searches-icon"
        />
        <span className="popular-searches-title">{label}</span>
      </div>
    );
  }

  return (
    <button
      type="button"
      className="popular-searches-toggle"
      aria-expanded={isExpanded}
      aria-controls={panelId}
      onClick={onToggle}
    >
      <TrendingUp
        size={15}
        aria-hidden="true"
        className="popular-searches-icon"
      />
      <span className="popular-searches-title">{label}</span>
      <ChevronDown
        size={16}
        aria-hidden="true"
        className={`popular-searches-chevron${isExpanded ? " is-expanded" : ""}`}
      />
    </button>
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
  collapseOnSearch = false,
  defaultExpanded = false,
}) {
  const [period, setPeriod] = useState("last24Hours");
  const [isExpanded, setIsExpanded] = useState(
    () => !collapsible || defaultExpanded,
  );
  const isDesktop = useMediaQuery(DESKTOP_MIN_WIDTH_MEDIA_QUERY);
  const displayLimit = getDisplayLimit(isDesktop);
  const panelId = useId();

  useEffect(() => {
    if (defaultExpanded) {
      setIsExpanded(true);
    }
  }, [defaultExpanded]);

  useEffect(() => {
    if (collapseOnSearch) {
      setIsExpanded(false);
    }
  }, [collapseOnSearch]);

  if (!isLoading && !hasAnyKeywords(keywordsByPeriod)) {
    return null;
  }

  const keywords = (keywordsByPeriod?.[period] ?? []).slice(0, displayLimit);
  const selectedPeriodLabel =
    PERIOD_OPTIONS.find((option) => option.id === period)?.label ?? "";
  const showContent = !collapsible || isExpanded;

  const handleToggle = () => {
    setIsExpanded((expanded) => !expanded);
  };

  const handlePopularSearchClick = (keyword) => {
    if (window.gtag) {
      window.gtag("event", "popular_search_click", {
        search_term: keyword,
        popular_search_period: period,
      });
    }
  };

  return (
    <div
      className={`${SECTION_CLASS_NAME}${
        showContent ? " is-expanded" : " is-collapsed"
      }`}
    >
      <PopularSearchesToggle
        isExpanded={isExpanded}
        collapsible={collapsible}
        panelId={panelId}
        onToggle={handleToggle}
      />

      {showContent && (
        <div className="popular-searches-panel" id={panelId}>
          <div className="popular-searches-controls">
            <PeriodToggle
              period={period}
              onPeriodChange={setPeriod}
              disabled={isLoading}
            />
          </div>

          {isLoading ? (
            <div className="popular-searches-pills">
              {LOADING_SKELETON_KEYS.slice(0, displayLimit).map((key) => (
                <span
                  key={key}
                  className="placeholder rounded-pill popular-search-pill-skeleton"
                />
              ))}
            </div>
          ) : keywords.length > 0 ? (
            <div className="popular-searches-pills">
              {keywords.map((keyword) => (
                <a
                  key={keyword}
                  href={buildPopularSearchUrl(BASE_URL, keyword, period)}
                  className="btn btn-sm popular-search-pill text-decoration-none"
                  aria-label={`Search for ${keyword}`}
                  onClick={() => handlePopularSearchClick(keyword)}
                >
                  {keyword}
                </a>
              ))}
            </div>
          ) : (
            <p className="popular-searches-empty small text-muted mb-0">
              No popular searches in the last {selectedPeriodLabel}.
            </p>
          )}
        </div>
      )}
    </div>
  );
}
