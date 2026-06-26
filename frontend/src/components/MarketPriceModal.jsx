import { useMemo, useState } from "react";
import { Button, Modal, Spinner } from "react-bootstrap";
import { ExternalLink, TrendingDown, TrendingUp } from "react-feather";
import {
  computeTrendPercent,
  filterHistoryByRange,
} from "../utils/marketPrices";
import MarketPriceChart from "./MarketPriceChart";

const RANGES = [
  { id: "1w", label: "1W" },
  { id: "1m", label: "1M" },
  { id: "3m", label: "3M" },
  { id: "all", label: "ALL" },
];

const formatUsd = (value) =>
  value == null
    ? "—"
    : new Intl.NumberFormat("en-US", {
        style: "currency",
        currency: "USD",
      }).format(value);

const formatSgd = (value) =>
  value == null
    ? "—"
    : new Intl.NumberFormat("en-SG", {
        style: "currency",
        currency: "SGD",
      }).format(value);

const ReferenceRow = ({ label, usd, sgd, url }) => (
  <tr>
    <td>
      {url ? (
        <a
          href={url}
          target="_blank"
          rel="noreferrer"
          className="link-offset-2"
        >
          {label}
        </a>
      ) : (
        label
      )}
    </td>
    <td>{formatSgd(sgd)}</td>
    <td>{formatUsd(usd)}</td>
  </tr>
);

const MarketPriceModal = ({
  show,
  onHide,
  card,
  data,
  error,
  status,
  isLoading,
}) => {
  const [range, setRange] = useState("3m");

  const filteredHistory = useMemo(() => {
    if (!data?.history) {
      return { cardkingdom: [], tcgplayer: [] };
    }
    return {
      cardkingdom: filterHistoryByRange(data.history.cardkingdom, range),
      tcgplayer: filterHistoryByRange(data.history.tcgplayer, range),
    };
  }, [data, range]);

  const anchorSeries =
    filteredHistory.cardkingdom.length > 0
      ? filteredHistory.cardkingdom
      : filteredHistory.tcgplayer;
  const trend = computeTrendPercent(anchorSeries);
  const trendLabel =
    trend == null
      ? null
      : `${trend >= 0 ? "▲" : "▼"} ${Math.abs(trend).toFixed(1)}%`;

  const tcgHigh =
    filteredHistory.tcgplayer.length > 0
      ? Math.max(...filteredHistory.tcgplayer.map((point) => point.price))
      : data?.references?.tcgplayer?.usd;
  const tcgLow =
    filteredHistory.tcgplayer.length > 0
      ? Math.min(...filteredHistory.tcgplayer.map((point) => point.price))
      : data?.references?.tcgplayer?.usd;

  return (
    <Modal show={show} onHide={onHide} size="lg" centered scrollable>
      <Modal.Header closeButton>
        <Modal.Title>Global market prices</Modal.Title>
      </Modal.Header>
      <Modal.Body>
        {card && (
          <div className="d-flex gap-3 align-items-start mb-3">
            {card.img && (
              <img
                src={data?.image ?? card.img}
                alt={card.name}
                className="market-modal-image rounded"
                loading="lazy"
              />
            )}
            <div>
              <div className="fw-bold">{data?.cardName ?? card.name}</div>
              {data?.setName && (
                <div className="small text-muted">
                  {data.setName}
                  {data.collectorNumber ? ` #${data.collectorNumber}` : ""}
                  {data.isFoil ? " · Foil" : " · Non-foil"}
                </div>
              )}
              <div className="small text-muted mt-1">
                Global market data (SGD) — not Singapore shop prices.
              </div>
            </div>
          </div>
        )}

        {isLoading && (
          <div className="text-center py-4">
            <Spinner animation="border" size="sm" className="me-2" />
            <span className="small text-muted">{status || "Loading…"}</span>
          </div>
        )}

        {error && !isLoading && (
          <div className="alert alert-danger mb-0" role="alert">
            {error}
          </div>
        )}

        {data && !isLoading && (
          <>
            <div className="market-hero rounded p-3 mb-3">
              <div className="small text-muted mb-1">
                Card Kingdom
                {data.references.cardkingdom.source === "live" ? " · live" : ""}
              </div>
              <div className="market-hero-price">
                {formatSgd(data.references.cardkingdom.sgd)}
              </div>
              <div className="small text-muted">
                {formatUsd(data.references.cardkingdom.usd)} USD
                {data.references.cardkingdom.quantity != null
                  ? ` · ${data.references.cardkingdom.quantity} in stock`
                  : ""}
              </div>
            </div>

            <div className="d-flex flex-wrap justify-content-between align-items-center gap-2 mb-2">
              <div>
                <div className="fw-semibold">Price history</div>
                {trendLabel && (
                  <div
                    className={`small ${trend >= 0 ? "text-success" : "text-danger"}`}
                  >
                    {trend >= 0 ? (
                      <TrendingUp size={14} className="me-1" />
                    ) : (
                      <TrendingDown size={14} className="me-1" />
                    )}
                    {trendLabel} past{" "}
                    {range === "all" ? "all data" : range.toUpperCase()}
                  </div>
                )}
              </div>
              <fieldset className="btn-group btn-group-sm border-0 p-0 m-0">
                <legend className="visually-hidden">Chart range</legend>
                {RANGES.map((item) => (
                  <button
                    key={item.id}
                    type="button"
                    className={`btn ${range === item.id ? "btn-primary" : "btn-outline-primary"}`}
                    onClick={() => setRange(item.id)}
                  >
                    {item.label}
                  </button>
                ))}
              </fieldset>
            </div>

            <MarketPriceChart
              cardkingdom={filteredHistory.cardkingdom}
              tcgplayer={filteredHistory.tcgplayer}
              usdToSgd={data.usdToSgd}
            />

            <div className="small text-muted mt-2 mb-3">
              Card Kingdom anchor
              {tcgHigh != null && tcgLow != null
                ? ` · TCGplayer High ${formatUsd(tcgHigh)} · Low ${formatUsd(tcgLow)}`
                : ""}
            </div>

            <h6 className="mb-2">Market references</h6>
            <div className="table-responsive">
              <table className="table table-sm align-middle mb-3">
                <thead>
                  <tr>
                    <th>Reference</th>
                    <th>SGD</th>
                    <th>USD</th>
                  </tr>
                </thead>
                <tbody>
                  <ReferenceRow
                    label="Card Kingdom"
                    usd={data.references.cardkingdom.usd}
                    sgd={data.references.cardkingdom.sgd}
                    url={data.references.cardkingdom.url}
                  />
                  <ReferenceRow
                    label="TCGplayer"
                    usd={data.references.tcgplayer.usd}
                    sgd={data.references.tcgplayer.sgd}
                    url={data.references.tcgplayer.url}
                  />
                </tbody>
              </table>
            </div>

            <div className="d-flex flex-wrap gap-2">
              {data.references.cardkingdom.url && (
                <Button
                  as="a"
                  href={data.references.cardkingdom.url}
                  target="_blank"
                  rel="noreferrer"
                  variant="outline-primary"
                  size="sm"
                >
                  <ExternalLink size={14} className="me-1" />
                  Card Kingdom
                </Button>
              )}
              {data.references.tcgplayer.url && (
                <Button
                  as="a"
                  href={data.references.tcgplayer.url}
                  target="_blank"
                  rel="noreferrer"
                  variant="outline-primary"
                  size="sm"
                >
                  <ExternalLink size={14} className="me-1" />
                  TCGplayer
                </Button>
              )}
            </div>
          </>
        )}
      </Modal.Body>
    </Modal>
  );
};

export default MarketPriceModal;
