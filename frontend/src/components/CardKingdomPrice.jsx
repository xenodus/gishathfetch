const formatUsd = (value) =>
  new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
  }).format(value);

const CardKingdomPrice = ({ price }) => {
  if (!price) {
    return null;
  }

  return (
    <div
      className="mb-3 text-center bg-body-secondary rounded py-2 px-3"
      aria-live="polite"
    >
      <span className="small">
        Card Kingdom from {formatUsd(price.priceUsd)}
        {price.edition ? ` · ${price.edition}` : ""}
        {price.isFoil ? " · Foil" : ""}
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
