import { useMemo } from "react";

const CheapestByStore = ({ results }) => {
  const cheapestByStore = useMemo(() => {
    const byStore = new Map();

    for (const card of results) {
      const existing = byStore.get(card.src);
      if (!existing || card.price < existing.price) {
        byStore.set(card.src, card);
      }
    }

    return [...byStore.values()].sort((a, b) => a.price - b.price);
  }, [results]);

  if (cheapestByStore.length < 2) {
    return null;
  }

  return (
    <div className="mb-4">
      <h6 className="text-center mb-2 fw-semibold">Cheapest per store</h6>
      <div className="table-responsive">
        <table className="table table-sm table-bordered align-middle mb-0 cheapest-by-store-table">
          <thead className="table-light">
            <tr>
              <th scope="col">Store</th>
              <th scope="col" className="text-nowrap">
                Price
              </th>
              <th scope="col">Listing</th>
            </tr>
          </thead>
          <tbody>
            {cheapestByStore.map((card) => {
              const details = [
                card.quality,
                card.isFoil ? "Foil" : null,
                card.extraInfo,
              ].filter(Boolean);

              return (
                <tr key={card.src}>
                  <td>
                    <a
                      href={card.url}
                      target="_blank"
                      rel="noreferrer"
                      className="link-offset-2"
                    >
                      {card.src}
                    </a>
                  </td>
                  <td className="text-nowrap fw-semibold">
                    S$ {card.price.toFixed(2)}
                  </td>
                  <td className="small text-muted">
                    {details.length > 0 ? details.join(" · ") : "—"}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default CheapestByStore;
