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
  clientDurationMs,
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
  const hasClient = Number.isFinite(clientDurationMs) && clientDurationMs >= 0;
  const clientGapMs =
    hasClient && hasTotal ? Math.max(0, clientDurationMs - totalDurationMs) : 0;
  const showClientGap = hasClient && clientGapMs >= 500;

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
          {storeStats.length === 0 && !hasTotal && !hasClient ? (
            <p className="small text-muted mb-0">
              No store timing data for this search.
            </p>
          ) : (
            <>
              {hasClient && (
                <p className="small text-muted mb-2 search-stats-client">
                  Time to results:{" "}
                  <span className="text-body fw-semibold">
                    {formatDurationMs(clientDurationMs)}
                  </span>
                  {" (browser wall clock: session check, network, and API)"}
                </p>
              )}
              {hasTotal && (
                <p className="small text-muted mb-2 search-stats-total">
                  API search time:{" "}
                  <span className="text-body fw-semibold">
                    {formatDurationMs(totalDurationMs)}
                  </span>
                  {storeStats.length > 1
                    ? " (waits for all selected stores and Card Kingdom enrichment; per-store times below)"
                    : " (includes Card Kingdom enrichment wait; per-store time below)"}
                </p>
              )}
              {showClientGap && (
                <p className="small text-muted mb-2 search-stats-gap">
                  Outside API search:{" "}
                  <span className="text-body fw-semibold">
                    {formatDurationMs(clientGapMs)}
                  </span>
                  {
                    " — usually session/Turnstile, Lambda cold start, or network; not counted in API search time or per-store times"
                  }
                </p>
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
