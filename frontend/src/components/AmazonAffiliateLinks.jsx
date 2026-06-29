import useAffiliateLinks from "../hooks/useAffiliateLinks";

export default function AmazonAffiliateLinks() {
  const { links, isLoading, error } = useAffiliateLinks(true);

  if (isLoading || links.length === 0) {
    return null;
  }

  return (
    <section className="my-4" aria-label="Featured Amazon products">
      <div className="d-flex align-items-center justify-content-between mb-2">
        <h2 className="h5 mb-0">Featured Products</h2>
        <span className="text-secondary small">Amazon affiliate links</span>
      </div>
      {error ? (
        <div className="alert alert-warning py-2 mb-0">{error}</div>
      ) : (
        <div className="row g-3">
          {links.map((item) => (
            <div key={item.id} className="col-12 col-sm-6 col-lg-4">
              <a
                href={item.link}
                target="_blank"
                rel="noopener noreferrer sponsored"
                className="card h-100 text-decoration-none text-reset affiliate-link-card"
              >
                <img
                  src={item.imageUrl}
                  alt={item.title || "Amazon product"}
                  className="card-img-top affiliate-link-image"
                  loading="lazy"
                />
                <div className="card-body">
                  <div className="fw-semibold">
                    {item.title || "View on Amazon"}
                  </div>
                  <div className="text-primary mt-1">{item.price}</div>
                </div>
              </a>
            </div>
          ))}
        </div>
      )}
    </section>
  );
}
