import { useEffect, useRef, useState } from "react";
import { ADSENSE_CLIENT, ADSENSE_DISPLAY_AD_SLOT } from "../constants";

const UNFILLED_COLLAPSE_MS = 2500;
const LAZY_UNFILLED_COLLAPSE_MS = 5000;
const LAZY_LOAD_ROOT_MARGIN = "200px";
const ADSENSE_SCRIPT_WAIT_MS = 10000;

function hasFilledAd(insEl) {
  if (!insEl?.isConnected) return false;

  const adStatus = insEl.getAttribute("data-ad-status");
  if (adStatus === "filled") return true;
  if (adStatus === "unfilled") return false;

  if (insEl.querySelector("iframe")) return true;
  // Some fills won't use an iframe immediately; a non-trivial height is a good proxy.
  return insEl.offsetHeight >= 50;
}

function isAdSenseProcessed(insEl) {
  return insEl?.getAttribute("data-adsbygoogle-status") === "done";
}

function pushAdSlot(insEl) {
  if (!insEl || isAdSenseProcessed(insEl)) return;

  try {
    // biome-ignore lint/suspicious/noAssignInExpressions: Legacy Google Ads code
    (window.adsbygoogle = window.adsbygoogle || []).push({});
  } catch (e) {
    console.error("AdSense error:", e);
  }
}

function waitForAdSenseScript(callback) {
  if (window.adsbygoogle?.loaded || window.adsbygoogle?.push) {
    callback();
    return () => {};
  }

  let cancelled = false;
  const startedAt = Date.now();

  const intervalId = window.setInterval(() => {
    if (cancelled) return;

    if (window.adsbygoogle?.loaded || window.adsbygoogle?.push) {
      window.clearInterval(intervalId);
      callback();
      return;
    }

    if (Date.now() - startedAt >= ADSENSE_SCRIPT_WAIT_MS) {
      window.clearInterval(intervalId);
      callback();
    }
  }, 100);

  return () => {
    cancelled = true;
    window.clearInterval(intervalId);
  };
}

const AdComponent = ({
  lazyLoad = false,
  collapseWhenUnfilled = true,
  slot = ADSENSE_DISPLAY_AD_SLOT,
  layoutKey,
}) => {
  const containerRef = useRef(null);
  const adInitialized = useRef(false);
  const insRef = useRef(null);
  const [collapsed, setCollapsed] = useState(false);
  const [isNearViewport, setIsNearViewport] = useState(!lazyLoad);
  const isInFeedFormat = Boolean(layoutKey);

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

    const insEl = insRef.current;
    if (!insEl) return;

    adInitialized.current = true;

    return waitForAdSenseScript(() => {
      pushAdSlot(insEl);
    });
  }, [isNearViewport]);

  useEffect(() => {
    if (!collapseWhenUnfilled || !isNearViewport) return;

    const insEl = insRef.current;
    if (!insEl) return;

    let cancelled = false;
    const collapseMs = lazyLoad
      ? LAZY_UNFILLED_COLLAPSE_MS
      : UNFILLED_COLLAPSE_MS;

    const maybeCollapse = () => {
      if (cancelled) return;
      if (insEl.getAttribute("data-ad-status") === "unfilled") {
        setCollapsed(true);
        return;
      }
      if (hasFilledAd(insEl)) setCollapsed(false);
    };

    const timeoutId = window.setTimeout(() => {
      if (cancelled) return;
      if (!hasFilledAd(insEl)) setCollapsed(true);
    }, collapseMs);

    const observer = new MutationObserver(maybeCollapse);
    observer.observe(insEl, {
      childList: true,
      subtree: true,
      attributes: true,
      attributeFilter: ["data-ad-status", "data-adsbygoogle-status"],
    });

    return () => {
      cancelled = true;
      window.clearTimeout(timeoutId);
      observer.disconnect();
    };
  }, [collapseWhenUnfilled, isNearViewport, lazyLoad]);

  if (collapsed) return null;

  if (lazyLoad && !isNearViewport) {
    return (
      <div
        ref={containerRef}
        className="ad-large d-print-none w-100"
        style={{ height: 1, overflow: "hidden" }}
        aria-hidden="true"
      />
    );
  }

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
          data-ad-client={ADSENSE_CLIENT}
          data-ad-slot={slot}
          data-ad-format={isInFeedFormat ? "fluid" : "auto"}
          data-full-width-responsive={isInFeedFormat ? undefined : "true"}
          {...(isInFeedFormat ? { "data-ad-layout-key": layoutKey } : {})}
        ></ins>
      </div>
      <div className="text-secondary" style={{ fontSize: "11px" }}>
        Advertisement
      </div>
    </div>
  );
};

export default AdComponent;
