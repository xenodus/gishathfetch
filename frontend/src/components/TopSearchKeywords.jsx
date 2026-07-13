import { useEffect, useId, useState } from "react";
import { ChevronDown, TrendingUp } from "react-feather";
import {
  BASE_URL,
  CK_PRICE_CHANGES_DISPLAY_LIMIT,
  DESKTOP_MIN_WIDTH_MEDIA_QUERY,
  TOP_SEARCH_KEYWORDS_DISPLAY_LIMIT,
  TOP_SEARCH_KEYWORDS_MOBILE_DISPLAY_LIMIT,
} from "../constants";
import useCKPriceIncreases from "../hooks/useCKPriceIncreases";
import useMediaQuery from "../hooks/useMediaQuery";
import { buildPopularSearchUrl, buildSearchQueryUrl } from "../utils/searchUrl";

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
  "top-search-keyword-skeleton-k",
  "top-search-keyword-skeleton-l",
  "top-search-keyword-skeleton-m",
  "top-search-keyword-skeleton-n",
  "top-search-keyword-skeleton-o",
  "top-search-keyword-skeleton-p",
  "top-search-keyword-skeleton-q",
  "top-search-keyword-skeleton-r",
  "top-search-keyword-skeleton-s",
  "top-search-keyword-skeleton-t",
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
      <legend className="visually-hidden">Trending search time range</legend>
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

function TrendingSectionToggle({ isExpanded, collapsible, panelId, onToggle }) {
  const label = isExpanded ? "Trending" : "Show trending";

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

function CKPriceIncreasesPanel({
  isVisible,
  isLoading,
  error,
  priceIncreases,
  searchQuery,
  panelId,
}) {
  if (!isVisible) {
    return null;
  }

  if (isLoading) {
    return (
      <div className="trending-price-increases" id={panelId}>
        <div className="popular-searches-pills">
          {LOADING_SKELETON_KEYS.slice(0, CK_PRICE_CHANGES_DISPLAY_LIMIT).map(
            (key) => (
              <span
                key={key}
                className="placeholder rounded-pill popular-search-pill-skeleton"
              />
            ),
          )}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="trending-price-increases" id={panelId}>
        <p className="popular-searches-empty small text-muted mb-0">{error}</p>
      </div>
    );
  }

  if (priceIncreases.length === 0) {
    return (
      <div className="trending-price-increases" id={panelId}>
        <p className="popular-searches-empty small text-muted mb-0">
          No CK price increases available.
        </p>
      </div>
    );
  }

  return (
    <div className="trending-price-increases" id={panelId}>
      <div className="popular-searches-pills">
        {priceIncreases.map((item) => {
          const isActive = isMatchingTrendingSearch(item.cardName, searchQuery);

          return (
            <a
              key={item.cardName}
              href={buildSearchQueryUrl(BASE_URL, item.cardName)}
              className={`btn btn-sm popular-search-pill text-decoration-none${
                isActive ? " is-active" : ""
              }`}
              aria-label={`Search for ${item.cardName}`}
              aria-current={isActive ? "page" : undefined}
            >
              {item.cardName}
            </a>
          );
        })}
      </div>
    </div>
  );
}

function getDisplayLimit(isDesktop, showAllKeywords) {
  if (isDesktop) {
    return TOP_SEARCH_KEYWORDS_DISPLAY_LIMIT;
  }

  return showAllKeywords
    ? TOP_SEARCH_KEYWORDS_DISPLAY_LIMIT
    : TOP_SEARCH_KEYWORDS_MOBILE_DISPLAY_LIMIT;
}

function isMatchingTrendingSearch(keyword, searchQuery) {
  const normalizedKeyword = String(keyword ?? "")
    .trim()
    .toLowerCase();
  const normalizedQuery = String(searchQuery ?? "")
    .trim()
    .toLowerCase();

  return (
    normalizedKeyword.length > 0 &&
    normalizedQuery.length > 0 &&
    normalizedKeyword === normalizedQuery
  );
}

export default function TopSearchKeywords({
  keywordsByPeriod,
  isLoading,
  searchQuery = "",
  collapsible = false,
  collapseOnSearch = false,
  defaultExpanded = false,
}) {
  const [period, setPeriod] = useState("last24Hours");
  const [isExpanded, setIsExpanded] = useState(
    () => !collapsible || defaultExpanded,
  );
  const [showAllKeywords, setShowAllKeywords] = useState(false);
  const [showPriceIncreases, setShowPriceIncreases] = useState(false);
  const {
    priceIncreases,
    isLoading: isLoadingPriceIncreases,
    error: priceIncreasesError,
    hasLoaded: hasLoadedPriceIncreases,
    loadPriceIncreases,
  } = useCKPriceIncreases();
  const isDesktop = useMediaQuery(DESKTOP_MIN_WIDTH_MEDIA_QUERY);
  const displayLimit = getDisplayLimit(isDesktop, showAllKeywords);
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

  const handlePeriodChange = (nextPeriod) => {
    setPeriod(nextPeriod);
    setShowAllKeywords(false);
  };

  if (!isLoading && !hasAnyKeywords(keywordsByPeriod)) {
    return null;
  }

  const allKeywords = keywordsByPeriod?.[period] ?? [];
  const keywords = allKeywords.slice(0, displayLimit);
  const hasMoreKeywords =
    !isDesktop && allKeywords.length > TOP_SEARCH_KEYWORDS_MOBILE_DISPLAY_LIMIT;
  const selectedPeriodLabel =
    PERIOD_OPTIONS.find((option) => option.id === period)?.label ?? "";
  const showContent = !collapsible || isExpanded;

  const handleToggle = () => {
    setIsExpanded((expanded) => !expanded);
  };

  const handleTrendingSearchClick = (keyword) => {
    if (window.gtag) {
      window.gtag("event", "popular_search_click", {
        search_term: keyword,
        popular_search_period: period,
      });
    }
  };

  const handlePriceIncreasesToggle = async () => {
    const nextVisible = !showPriceIncreases;
    setShowPriceIncreases(nextVisible);

    if (nextVisible && !hasLoadedPriceIncreases && !isLoadingPriceIncreases) {
      await loadPriceIncreases();
    }
  };

  return (
    <div
      className={`${SECTION_CLASS_NAME}${
        showContent ? " is-expanded" : " is-collapsed"
      }`}
    >
      <TrendingSectionToggle
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
              onPeriodChange={handlePeriodChange}
              disabled={isLoading}
            />
            <button
              type="button"
              className={`btn btn-sm trending-price-increases-btn${
                showPriceIncreases ? " is-active" : ""
              }`}
              aria-expanded={showPriceIncreases}
              aria-controls={`${panelId}-price-increases`}
              disabled={isLoadingPriceIncreases}
              onClick={handlePriceIncreasesToggle}
            >
              {showPriceIncreases
                ? "Hide CK price increases"
                : "Show CK price increases"}
            </button>
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
            <>
              <div className="popular-searches-pills">
                {keywords.map((keyword) => {
                  const isActive = isMatchingTrendingSearch(
                    keyword,
                    searchQuery,
                  );

                  return (
                    <a
                      key={keyword}
                      href={buildPopularSearchUrl(BASE_URL, keyword, period)}
                      className={`btn btn-sm popular-search-pill text-decoration-none${
                        isActive ? " is-active" : ""
                      }`}
                      aria-label={`Search for ${keyword}`}
                      aria-current={isActive ? "page" : undefined}
                      onClick={() => handleTrendingSearchClick(keyword)}
                    >
                      {keyword}
                    </a>
                  );
                })}
              </div>
              {hasMoreKeywords && (
                <button
                  type="button"
                  className="popular-searches-show-more"
                  aria-expanded={showAllKeywords}
                  onClick={() => setShowAllKeywords((expanded) => !expanded)}
                >
                  {showAllKeywords ? "Show less" : "Show more"}
                </button>
              )}
            </>
          ) : (
            <p className="popular-searches-empty small text-muted mb-0">
              No trending searches in the last {selectedPeriodLabel}.
            </p>
          )}

          <CKPriceIncreasesPanel
            isVisible={showPriceIncreases}
            isLoading={isLoadingPriceIncreases}
            error={priceIncreasesError}
            priceIncreases={priceIncreases}
            searchQuery={searchQuery}
            panelId={`${panelId}-price-increases`}
          />
        </div>
      )}
    </div>
  );
}
