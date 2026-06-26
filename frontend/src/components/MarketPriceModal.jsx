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
    if (!data?.history?.cardkingdom) {
      return [];
    }
    return filterHistoryByRange(data.history.cardkingdom, range);
  }, [data, range]);

  const trend = computeTrendPercent(filteredHistory);
  const trendLabel =
    trend == null
      ? null
      : `${trend >= 0 ? "▲" : "▼"} ${Math.abs(trend).toFixed(1)}%`;

  return (
    <Modal show={show} onHide={onHide} size="lg" centered scrollable>
      <Modal.Header closeButton>
        <Modal.Title>Card Kingdom market price</Modal.Title>
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
                  {data.isFoil ? " · Foil" : " · Non-foil"}
                </div>
              )}
              <div className="small text-muted mt-1">
                Card Kingdom reference price (USD) — not a Singapore shop price.
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
                Card Kingdom · {data.priceListDate}
              </div>
              <div className="market-hero-price">
                {formatUsd(data.references.cardkingdom.usd)}
              </div>
              {data.references.cardkingdom.quantity != null && (
                <div className="small text-muted">
                  {data.references.cardkingdom.quantity} in stock
                </div>
              )}
            </div>

            <div className="d-flex flex-wrap justify-content-between align-items-center gap-2 mb-2">
              <div>
                <div className="fw-semibold">Price trend</div>
                {trendLabel ? (
                  <div
                    className={`small ${trend >= 0 ? "text-success" : "text-danger"}`}
                  >
                    {trend >= 0 ? (
                      <TrendingUp size={14} className="me-1" />
                    ) : (
                      <TrendingDown size={14} className="me-1" />
                    )}
                    {trendLabel} past{" "}
                    {range === "all" ? "recorded updates" : range.toUpperCase()}
                  </div>
                ) : (
                  <div className="small text-muted">
                    Trend appears after Card Kingdom price updates on future
                    visits.
                  </div>
                )}
              </div>
              <fieldset className="btn-group btn-group-sm border-0 p-0 m-0">
                <legend className="visually-hidden">Trend range</legend>
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

            <MarketPriceChart cardkingdom={filteredHistory} />

            <div className="small text-muted mt-2 mb-3">
              Trend is built from Card Kingdom price list snapshots saved on
              this device.
            </div>

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
                View on Card Kingdom
              </Button>
            )}
          </>
        )}
      </Modal.Body>
    </Modal>
  );
};

export default MarketPriceModal;
