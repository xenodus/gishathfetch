import { useEffect, useRef, useState } from "react";

const UNFILLED_COLLAPSE_MS = 2500;
const LAZY_LOAD_ROOT_MARGIN = "200px";

function hasFilledAd(insEl) {
  if (!insEl?.isConnected) return false;
  if (insEl.querySelector("iframe")) return true;
  // Some fills won't use an iframe immediately; a non-trivial height is a good proxy.
  return insEl.offsetHeight >= 50;
}

const AdComponent = ({ lazyLoad = false, collapseWhenUnfilled = true }) => {
  const containerRef = useRef(null);
  const adInitialized = useRef(false);
  const insRef = useRef(null);
  const [collapsed, setCollapsed] = useState(false);
  const [isNearViewport, setIsNearViewport] = useState(!lazyLoad);

  useEffect(() => {
    if (!lazyLoad) return;

    const containerEl = containerRef.current;
    if (!containerEl) return;

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIsNearViewport(true);
          observer.disconnect();
        }
      },
      { rootMargin: LAZY_LOAD_ROOT_MARGIN },
    );

    observer.observe(containerEl);

    return () => observer.disconnect();
  }, [lazyLoad]);

  useEffect(() => {
    if (!isNearViewport || adInitialized.current) return;

    adInitialized.current = true;

    try {
      // biome-ignore lint/suspicious/noAssignInExpressions: Legacy Google Ads code
      (window.adsbygoogle = window.adsbygoogle || []).push({});
    } catch (e) {
      console.error("AdSense error:", e);
    }
  }, [isNearViewport]);

  useEffect(() => {
    if (!collapseWhenUnfilled || !isNearViewport) return;

    const insEl = insRef.current;
    if (!insEl) return;

    let cancelled = false;

    const timeoutId = window.setTimeout(() => {
      if (cancelled) return;
      if (!hasFilledAd(insEl)) setCollapsed(true);
    }, UNFILLED_COLLAPSE_MS);

    const observer = new MutationObserver(() => {
      if (cancelled) return;
      if (hasFilledAd(insEl)) setCollapsed(false);
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
  }, [collapseWhenUnfilled, isNearViewport]);

  if (collapsed) return null;

  return (
    <div
      ref={containerRef}
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
