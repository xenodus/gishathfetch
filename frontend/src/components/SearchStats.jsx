import { useEffect, useState } from "react";
import { Form } from "react-bootstrap";

const SEARCH_STATS_STORAGE_KEY = "gishathfetch-show-search-stats";

function formatDurationMs(durationMs) {
  if (!Number.isFinite(durationMs) || durationMs < 0) {
    return "—";
  }
  if (durationMs >= 1000) {
    return `${(durationMs / 1000).toFixed(2)}s`;
  }
  return `${Math.round(durationMs)}ms`;
}

function readShowStatsPreference() {
  try {
    return localStorage.getItem(SEARCH_STATS_STORAGE_KEY) === "1";
  } catch {
    return false;
  }
}

const SearchStats = ({
  stats,
  totalDurationMs,
  sessionTurnstileDurationMs,
  sessionMintDurationMs,
  searchResponseDurationMs,
  hasSearched,
  isSearching,
}) => {
  const [showStats, setShowStats] = useState(readShowStatsPreference);

  useEffect(() => {
    try {
      localStorage.setItem(SEARCH_STATS_STORAGE_KEY, showStats ? "1" : "0");
    } catch {
      // Ignore storage access issues and keep in-memory preference only.
    }
  }, [showStats]);

  if (!hasSearched || isSearching) {
    return null;
  }

  const storeStats = Array.isArray(stats) ? stats : [];
  const hasTotal = Number.isFinite(totalDurationMs) && totalDurationMs >= 0;
  const hasTurnstileTiming = sessionTurnstileDurationMs !== null;
  const hasSessionMintTiming =
    Number.isFinite(sessionMintDurationMs) && sessionMintDurationMs >= 0;
  const hasSearchResponseTiming =
    Number.isFinite(searchResponseDurationMs) && searchResponseDurationMs >= 0;
  const hasClientTiming =
    hasTurnstileTiming || hasSessionMintTiming || hasSearchResponseTiming;

  return (
    <div className="search-stats mt-3 rounded py-2 px-3">
      <Form.Check
        type="switch"
        id="show-search-stats"
        label="Show search stats"
        checked={showStats}
        onChange={(e) => setShowStats(e.target.checked)}
        className="mb-0 search-stats-toggle"
      />

      {showStats && (
        <div className="search-stats-panel mt-2">
          {storeStats.length === 0 && !hasTotal && !hasClientTiming ? (
            <p className="small text-muted mb-0">
              No store timing data for this search.
            </p>
          ) : (
            <>
              {hasTotal && (
                <p className="small text-muted mb-2 search-stats-total">
                  Total search time:{" "}
                  <span className="text-body fw-semibold">
                    {formatDurationMs(totalDurationMs)}
                  </span>
                  {storeStats.length > 1
                    ? " (waits for all selected stores and Card Kingdom enrichment; per-store times below)"
                    : " (includes Card Kingdom enrichment wait; per-store time below)"}
                </p>
              )}
              {hasClientTiming && (
                <div className="table-responsive mb-2">
                  <table className="table table-sm table-borderless mb-0 search-stats-table">
                    <thead>
                      <tr>
                        <th scope="col">Client timing</th>
                        <th scope="col" className="text-end">
                          Time
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      {hasTurnstileTiming && (
                        <tr>
                          <td>Turnstile</td>
                          <td className="text-end text-nowrap">
                            {formatDurationMs(sessionTurnstileDurationMs)}
                          </td>
                        </tr>
                      )}
                      {hasSessionMintTiming && (
                        <tr>
                          <td>Session mint</td>
                          <td className="text-end text-nowrap">
                            {formatDurationMs(sessionMintDurationMs)}
                          </td>
                        </tr>
                      )}
                      {hasSearchResponseTiming && (
                        <tr>
                          <td>Search response</td>
                          <td className="text-end text-nowrap">
                            {formatDurationMs(searchResponseDurationMs)}
                          </td>
                        </tr>
                      )}
                    </tbody>
                  </table>
                </div>
              )}
              {storeStats.length > 0 && (
                <div className="table-responsive">
                  <table className="table table-sm table-borderless mb-0 search-stats-table">
                    <thead>
                      <tr>
                        <th scope="col">Store</th>
                        <th scope="col" className="text-end">
                          Items
                        </th>
                        <th scope="col" className="text-end">
                          Time
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      {storeStats.map((stat) => (
                        <tr key={stat.store}>
                          <td>{stat.store}</td>
                          <td className="text-end">{stat.itemCount}</td>
                          <td className="text-end text-nowrap">
                            {formatDurationMs(stat.durationMs)}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </>
          )}
        </div>
      )}
    </div>
  );
};

export default SearchStats;
