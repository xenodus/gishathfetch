const formatUsd = (value) =>
  new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
  }).format(value);

const formatDataDate = (value) => {
  if (!value) {
    return null;
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return null;
  }

  const day = String(date.getUTCDate()).padStart(2, "0");
  const month = new Intl.DateTimeFormat("en-US", {
    month: "short",
    timeZone: "UTC",
  }).format(date);
  const year = date.getUTCFullYear();

  return `${day} ${month} ${year}`;
};

const CardKingdomPrice = ({ price }) => {
  if (!price) {
    return null;
  }

  const dataDate = formatDataDate(price.updatedAt);

  return (
    <div
      className="mb-3 text-center bg-body-secondary rounded py-2 px-3"
      aria-live="polite"
    >
      <span className="small">
        Card Kingdom from {formatUsd(price.priceUsd)}
        {price.edition ? ` · ${price.edition}` : ""}
        {price.isFoil ? " · Foil" : ""}
        {dataDate ? ` · as of ${dataDate}` : ""}
        {" · "}
        <a
          href={price.url}
          target="_blank"
          rel="noreferrer"
          className="link-offset-2"
        >
          View listing
        </a>
      </span>
    </div>
  );
};

export default CardKingdomPrice;
