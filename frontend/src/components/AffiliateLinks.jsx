import { AFFILIATE_PLATFORMS } from "../constants";

function platformLabel(platform) {
  return AFFILIATE_PLATFORMS[platform]?.label || platform;
}

function platformLinkLabel(platform) {
  return AFFILIATE_PLATFORMS[platform]?.linkLabel || "View product";
}

function groupLinksByPlatform(links) {
  const grouped = new Map();
  for (const link of links) {
    const platform = link.platform || "amazon";
    if (!grouped.has(platform)) {
      grouped.set(platform, []);
    }
    grouped.get(platform).push(link);
  }
  return grouped;
}

export default function AffiliateLinks({ links, error }) {
  if (links.length === 0) {
    return null;
  }

  const groupedLinks = groupLinksByPlatform(links);

  return (
    <section className="my-4" aria-label="Featured affiliate products">
      <div className="d-flex align-items-center justify-content-between mb-2">
        <h2 className="h5 mb-0">Featured Products</h2>
        <span className="text-secondary small">Affiliate links</span>
      </div>
      {error ? (
        <div className="alert alert-warning py-2 mb-0">{error}</div>
      ) : (
        [...groupedLinks.entries()].map(([platform, platformLinks]) => (
          <div key={platform} className="mb-3">
            <h3 className="h6 text-secondary mb-2">
              {platformLabel(platform)}
            </h3>
            <div className="row g-3">
              {platformLinks.map((item) => (
                <div key={item.id} className="col-12 col-sm-6 col-lg-4">
                  <a
                    href={item.link}
                    target="_blank"
                    rel="noopener noreferrer sponsored"
                    className="card h-100 text-decoration-none text-reset affiliate-link-card"
                  >
                    <img
                      src={item.imageUrl}
                      alt={item.title || platformLabel(item.platform)}
                      className="card-img-top affiliate-link-image"
                      loading="lazy"
                    />
                    <div className="card-body">
                      <div className="fw-semibold">
                        {item.title || platformLinkLabel(item.platform)}
                      </div>
                      <div className="text-primary mt-1">{item.price}</div>
                    </div>
                  </a>
                </div>
              ))}
            </div>
          </div>
        ))
      )}
    </section>
  );
}
