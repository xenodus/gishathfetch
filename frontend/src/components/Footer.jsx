import React, { useEffect } from "react";
import { ArrowUp, FolderPlus, HelpCircle, Map as MapIcon } from "react-feather";

const Footer = ({ cartCount, onShowCart, onShowMap, onShowFaq }) => {
  const adInitialized = React.useRef(false);

  useEffect(() => {
    if (adInitialized.current) return;
    adInitialized.current = true;

    try {
      // biome-ignore lint/suspicious/noAssignInExpressions: Legacy Google Ads code
      (window.adsbygoogle = window.adsbygoogle || []).push({});

      // Parity fix: use setTimeout to set z-index for footer ads, exactly as in legacy index.js
      setTimeout(() => {
        const ads = document.querySelectorAll("ins.adsbygoogle");
        ads.forEach((ad) => {
          ad.style.zIndex = "1000";
        });
      }, 1000);
    } catch (e) {
      console.error("Footer AdSense error:", e);
    }
  }, []);
  return (
    <>
      {/* Footer / Ads */}
      <div className="ad-large mt-4 pb-5 text-center d-print-none d-block d-sm-block">
        <div className="text-secondary mb-2" style={{ fontSize: "11px" }}>
          Advertisement
        </div>
        <div style={{ minHeight: "90px" }}>
          {/* AdSense slot placeholder */}
          <ins
            className="adsbygoogle"
            style={{ display: "inline-block", width: "728px", height: "90px" }}
            data-ad-client="ca-pub-2393161407259792"
            data-ad-slot="6707964257"
          ></ins>
        </div>
        <div className="text-center mt-2" style={{ fontSize: "11px" }}>
          <a
            href="https://www.patreon.com/GishathFetch"
            target="_blank"
            rel="noreferrer"
          >
            Follow / Support Gishath Fetch on Patreon
          </a>
        </div>
      </div>

      {/* Fixed Bottom Navigation */}
      <div className="fixed-bottom bg-primary text-light text-center">
        <div className="d-flex flex-row align-items-center justify-content-center">
          <button
            type="button"
            aria-label={`View saved cards${cartCount > 0 ? ` (${cartCount} items)` : ""}`}
            className="btn btn-link py-1 link-light link-offset-2 link-underline-opacity-0 text-decoration-none border-0"
            onClick={onShowCart}
          >
            <div className="px-2 py-1">
              <FolderPlus size={14} className="me-1 mb-1" /> Saved{" "}
              {cartCount > 0 && `(${cartCount})`}
            </div>
          </button>
          <button
            type="button"
            aria-label="View store locations map"
            className="btn btn-link py-1 link-light link-offset-2 link-underline-opacity-0 text-decoration-none border-0"
            onClick={onShowMap}
          >
            <div className="px-2 py-1">
              <MapIcon size={14} className="me-1" /> Map
            </div>
          </button>
          <button
            type="button"
            aria-label="View frequently asked questions"
            className="btn btn-link py-1 link-light link-offset-2 link-underline-opacity-0 text-decoration-none border-0"
            onClick={onShowFaq}
          >
            <div className="px-2 py-1">
              <HelpCircle size={14} className="me-1 mb-1" /> FAQs
            </div>
          </button>
          <a
            href="#top"
            aria-label="Scroll to top of page"
            className="py-1 link-light link-offset-2 link-underline-opacity-0"
          >
            <div className="px-2 py-1">
              <ArrowUp size={14} className="me-1" /> Top
            </div>
          </a>
        </div>
      </div>
    </>
  );
};

export default Footer;
