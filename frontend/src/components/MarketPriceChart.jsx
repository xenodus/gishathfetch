const CHART_WIDTH = 640;
const CHART_HEIGHT = 220;
const PADDING = { top: 16, right: 16, bottom: 28, left: 44 };

const formatMoney = (value, currency = "USD") =>
  new Intl.NumberFormat("en-US", {
    style: "currency",
    currency,
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(value);

const buildPath = (points, xScale, yScale) => {
  if (points.length === 0) {
    return "";
  }
  return points
    .map((point, index) => {
      const command = index === 0 ? "M" : "L";
      return `${command}${xScale(point.date)},${yScale(point.price)}`;
    })
    .join(" ");
};

const MarketPriceChart = ({ cardkingdom = [] }) => {
  if (cardkingdom.length === 0) {
    return (
      <div className="market-chart-empty text-muted small text-center py-4">
        No Card Kingdom trend data yet.
      </div>
    );
  }

  if (cardkingdom.length === 1) {
    return (
      <div className="market-chart-empty text-muted small text-center py-4">
        Trend chart appears after Card Kingdom price updates on future visits.
      </div>
    );
  }

  const dates = cardkingdom.map((point) => point.date);
  const prices = cardkingdom.map((point) => point.price);
  const minPrice = Math.min(...prices);
  const maxPrice = Math.max(...prices);
  const pricePadding =
    maxPrice === minPrice ? 0.5 : (maxPrice - minPrice) * 0.1;
  const yMin = Math.max(0, minPrice - pricePadding);
  const yMax = maxPrice + pricePadding;

  const innerWidth = CHART_WIDTH - PADDING.left - PADDING.right;
  const innerHeight = CHART_HEIGHT - PADDING.top - PADDING.bottom;

  const xScale = (date) => {
    const index = dates.indexOf(date);
    if (dates.length === 1) {
      return PADDING.left + innerWidth / 2;
    }
    return PADDING.left + (index / (dates.length - 1)) * innerWidth;
  };

  const yScale = (price) =>
    PADDING.top +
    innerHeight -
    ((price - yMin) / (yMax - yMin || 1)) * innerHeight;

  const yTicks = [yMin, (yMin + yMax) / 2, yMax];
  const xTickIndexes =
    dates.length <= 4
      ? dates.map((_, index) => index)
      : [0, Math.floor(dates.length / 2), dates.length - 1];

  return (
    <div className="market-chart-wrap">
      <svg
        viewBox={`0 0 ${CHART_WIDTH} ${CHART_HEIGHT}`}
        className="market-chart"
        role="img"
        aria-label="Card Kingdom price trend chart"
      >
        {yTicks.map((tick) => (
          <g key={tick}>
            <line
              x1={PADDING.left}
              x2={CHART_WIDTH - PADDING.right}
              y1={yScale(tick)}
              y2={yScale(tick)}
              className="market-chart-grid"
            />
            <text
              x={PADDING.left - 8}
              y={yScale(tick) + 4}
              textAnchor="end"
              className="market-chart-axis"
            >
              {formatMoney(tick, "USD")}
            </text>
          </g>
        ))}

        <path
          d={buildPath(cardkingdom, xScale, yScale)}
          className="market-chart-line market-chart-line-ck"
          fill="none"
        />

        {xTickIndexes.map((index) => (
          <text
            key={dates[index]}
            x={xScale(dates[index])}
            y={CHART_HEIGHT - 8}
            textAnchor="middle"
            className="market-chart-axis"
          >
            {dates[index].slice(5)}
          </text>
        ))}
      </svg>

      <div className="market-chart-legend d-flex justify-content-center gap-3 small">
        <span>
          <span className="market-chart-swatch market-chart-swatch-ck" />
          Card Kingdom
        </span>
      </div>
    </div>
  );
};

export default MarketPriceChart;
