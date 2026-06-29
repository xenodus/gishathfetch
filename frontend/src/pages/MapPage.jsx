import { useCallback, useEffect, useState } from "react";
import { ArrowLeft } from "react-feather";
import { Link } from "react-router-dom";
import LazyMapIframe from "../components/LazyMapIframe";
import StoreLocationMap from "../components/StoreLocationMap";
import { LGS_MAP } from "../constants";
import { applyMapSeo } from "../utils/seo";

const THEME_STORAGE_KEY = "gishathfetch-theme";

const MapPage = () => {
  const [activeStoreId, setActiveStoreId] = useState(null);

  useEffect(() => {
    applyMapSeo();
  }, []);

  useEffect(() => {
    if (typeof window === "undefined") {
      return;
    }

    try {
      const savedTheme = localStorage.getItem(THEME_STORAGE_KEY);
      if (savedTheme === "light" || savedTheme === "dark") {
        document.documentElement.setAttribute("data-bs-theme", savedTheme);
      }
    } catch {
      // Ignore storage access issues.
    }
  }, []);

  const handleStoreLinkClick = useCallback((storeId) => {
    setActiveStoreId(storeId);
  }, []);

  const handleMarkerClick = useCallback((storeId) => {
    setActiveStoreId(storeId);
    const element = document.getElementById(storeId);
    element?.scrollIntoView({ behavior: "smooth", block: "start" });
  }, []);

  return (
    <div className="container-xl my-3 px-3 pb-5">
      <nav className="mb-3" aria-label="Map page navigation">
        <Link
          to="/"
          className="btn btn-outline-primary d-inline-flex align-items-center gap-2"
        >
          <ArrowLeft size={16} aria-hidden="true" />
          Back to search
        </Link>
      </nav>

      <header className="mb-4">
        <h1 className="h3 mb-2" id="map-list">
          Where are the shops?
        </h1>
        <p className="text-muted mb-0">
          Singapore MTG store locations on the Google map below. Tap a pin or
          shop name for details.
        </p>
      </header>

      <section className="mb-4" aria-label="Singapore store map">
        <StoreLocationMap
          stores={LGS_MAP}
          activeStoreId={activeStoreId}
          onMarkerClick={handleMarkerClick}
        />
      </section>

      <section className="mb-4" aria-label="Store list">
        <h2 className="h5 mb-3">All stores</h2>
        <ul className="store-map-list">
          {LGS_MAP.map((shop) => (
            <li key={shop.id}>
              <a
                href={`#${shop.id}`}
                className="link-offset-2"
                onClick={() => handleStoreLinkClick(shop.id)}
              >
                {shop.name}
              </a>
            </li>
          ))}
        </ul>
      </section>

      {LGS_MAP.map((shop) => (
        <section
          id={shop.id}
          key={shop.id}
          className="mb-4 map-item"
          aria-label={`${shop.name} location details`}
        >
          <h2 className="h5">{shop.name}</h2>
          <div className="mb-2">{shop.address}</div>
          <div className="mb-2">
            <a href={shop.website} target="_blank" rel="noreferrer">
              {shop.website}
            </a>
          </div>
          <div className="mb-2">
            <Link
              to={`/?src=${encodeURIComponent(shop.searchStore)}`}
              className="link-offset-2"
            >
              Search this store on Gishath Fetch
            </Link>
          </div>
          <LazyMapIframe src={shop.iframe} title={shop.name} isActive />
          <div>
            <a href="#map-list" className="btn btn-primary">
              Back to top
            </a>
          </div>
        </section>
      ))}

      <footer className="text-muted small">
        &copy; 2023 gishathfetch.com by{" "}
        <a href="https://github.com/xenodus" target="_blank" rel="noreferrer">
          xenodus
        </a>
      </footer>
    </div>
  );
};

export default MapPage;
