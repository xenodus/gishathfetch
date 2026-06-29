import { useEffect, useRef, useState } from "react";

const AdComponent = () => {
  const adInitialized = useRef(false);
  const insRef = useRef(null);
  const [collapsed, setCollapsed] = useState(false);

  useEffect(() => {
    if (adInitialized.current) return;
    adInitialized.current = true;

    try {
      // biome-ignore lint/suspicious/noAssignInExpressions: Legacy Google Ads code
      (window.adsbygoogle = window.adsbygoogle || []).push({});

      // Parity fix: use setTimeout to set z-index, exactly as in legacy index.js
      setTimeout(() => {
        const ads = document.querySelectorAll("ins.adsbygoogle");
        ads.forEach((ad) => {
          ad.style.zIndex = "1000";
        });
      }, 1000);
    } catch (e) {
      console.error("AdSense error:", e);
    }
  }, []);

  useEffect(() => {
    const insEl = insRef.current;
    if (!insEl) return;

    let cancelled = false;

    const hasFilledAd = () => {
      if (!insEl.isConnected) return false;
      if (insEl.querySelector("iframe")) return true;
      // Some fills won't use an iframe immediately; a non-trivial height is a good proxy.
      return insEl.offsetHeight >= 50;
    };

    // Give AdSense a moment to render; if it doesn't fill, collapse this block.
    const timeoutId = window.setTimeout(() => {
      if (cancelled) return;
      if (!hasFilledAd()) setCollapsed(true);
    }, 2500);

    // If it does fill later, un-collapse.
    const observer = new MutationObserver(() => {
      if (cancelled) return;
      if (hasFilledAd()) setCollapsed(false);
    });
    observer.observe(insEl, {
      childList: true,
      subtree: true,
      attributes: true,
    });

    return () => {
      cancelled = true;
      window.clearTimeout(timeoutId);
      observer.disconnect();
    };
  }, []);

  if (collapsed) return null;

  return (
    <div
      className="ad-large text-center d-print-none d-block d-sm-block w-100"
      style={{ overflow: "hidden" }}
    >
      <div className="text-center mb-2" style={{ fontSize: "11px" }}>
        <a
          href="https://www.patreon.com/GishathFetch"
          target="_blank"
          rel="noreferrer"
        >
          Follow / Support Gishath Fetch on Patreon
        </a>
      </div>
      <div style={{ minHeight: "90px", overflow: "hidden" }}>
        <ins
          ref={insRef}
          className="adsbygoogle"
          style={{ display: "block", width: "100%", maxWidth: "100%" }}
          data-ad-client="ca-pub-2393161407259792"
          data-ad-slot="6707964257"
          data-ad-format="auto"
          data-full-width-responsive="true"
        ></ins>
      </div>
      <div className="text-secondary" style={{ fontSize: "11px" }}>
        Advertisement
      </div>
    </div>
  );
};

export default AdComponent;
