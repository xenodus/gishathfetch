import { useEffect } from "react";

const FOOTER_NAV_SELECTOR = ".site-bottom-nav";
const ANCHOR_AD_SELECTOR = "ins.adsbygoogle.adsbygoogle-noablate";
const FOOTER_NAV_HEIGHT_VAR = "--footer-nav-height";
const ANCHOR_AD_HEIGHT_VAR = "--anchor-ad-height";

function parsePixelValue(value) {
  const parsed = Number.parseFloat(value);
  return Number.isFinite(parsed) ? parsed : 0;
}

function measureFooterNavHeight() {
  const footerNav = document.querySelector(FOOTER_NAV_SELECTOR);
  if (!footerNav) {
    return 0;
  }

  return Math.ceil(footerNav.getBoundingClientRect().height);
}

function measureAnchorAdHeight() {
  const anchorAds = document.querySelectorAll(ANCHOR_AD_SELECTOR);
  let maxHeight = 0;

  for (const anchorAd of anchorAds) {
    if (!anchorAd.isConnected) continue;

    const status = anchorAd.getAttribute("data-anchor-status");
    if (status === "dismissed") continue;

    const rect = anchorAd.getBoundingClientRect();
    const computed = window.getComputedStyle(anchorAd);
    const height = Math.max(
      rect.height,
      anchorAd.offsetHeight,
      parsePixelValue(computed.height),
    );
    if (height <= 0) continue;
    if (!isBottomAnchorAd(anchorAd)) continue;

    maxHeight = Math.max(maxHeight, Math.ceil(height));
  }

  return maxHeight;
}

function syncBottomChromeLayout() {
  const footerNavHeight = measureFooterNavHeight();
  const anchorAdHeight = measureAnchorAdHeight();

  document.documentElement.style.setProperty(
    FOOTER_NAV_HEIGHT_VAR,
    `${footerNavHeight}px`,
  );
  document.documentElement.style.setProperty(
    ANCHOR_AD_HEIGHT_VAR,
    `${anchorAdHeight}px`,
  );
  applyAnchorAdOffsets(footerNavHeight);
}

function isBottomAnchorAd(anchorAd) {
  const computed = window.getComputedStyle(anchorAd);
  if (computed.position !== "fixed") {
    return false;
  }

  const bottom = parsePixelValue(computed.bottom);
  if (bottom > 0 && bottom <= 16) {
    return true;
  }

  const rect = anchorAd.getBoundingClientRect();
  return rect.height > 0 && rect.bottom >= window.innerHeight - 16;
}

function applyAnchorAdOffsets(footerNavHeight) {
  const anchorAds = document.querySelectorAll(ANCHOR_AD_SELECTOR);

  for (const anchorAd of anchorAds) {
    if (!anchorAd.isConnected) continue;

    const status = anchorAd.getAttribute("data-anchor-status");
    if (status === "dismissed") {
      anchorAd.style.removeProperty("margin-bottom");
      anchorAd.style.removeProperty("z-index");
      continue;
    }

    if (!isBottomAnchorAd(anchorAd)) {
      continue;
    }

    anchorAd.style.setProperty(
      "margin-bottom",
      `${footerNavHeight}px`,
      "important",
    );
    anchorAd.style.setProperty("z-index", "1020", "important");
  }
}

export default function useBottomChromeLayout() {
  useEffect(() => {
    let rafId = 0;
    const resizeObservers = new Map();

    const unobserveAnchorAd = (anchorAd) => {
      const observer = resizeObservers.get(anchorAd);
      if (!observer) return;
      observer.disconnect();
      resizeObservers.delete(anchorAd);
    };

    const observeAnchorAd = (anchorAd) => {
      if (resizeObservers.has(anchorAd)) return;

      const resizeObserver = new ResizeObserver(scheduleSync);
      resizeObserver.observe(anchorAd);
      resizeObservers.set(anchorAd, resizeObserver);
    };

    const syncAnchorObservers = () => {
      const anchorAds = document.querySelectorAll(ANCHOR_AD_SELECTOR);
      for (const anchorAd of anchorAds) {
        observeAnchorAd(anchorAd);
      }
      for (const anchorAd of resizeObservers.keys()) {
        if (!anchorAd.isConnected) {
          unobserveAnchorAd(anchorAd);
        }
      }
    };

    const scheduleSync = () => {
      if (rafId) {
        cancelAnimationFrame(rafId);
      }
      rafId = requestAnimationFrame(() => {
        rafId = 0;
        syncAnchorObservers();
        syncBottomChromeLayout();
      });
    };

    scheduleSync();

    const observer = new MutationObserver(scheduleSync);
    observer.observe(document.body, {
      childList: true,
      subtree: true,
      attributes: true,
      attributeFilter: ["class", "data-anchor-status"],
    });

    window.addEventListener("resize", scheduleSync);

    const pollId = window.setInterval(scheduleSync, 2000);
    const stopPollId = window.setTimeout(() => {
      window.clearInterval(pollId);
    }, 30000);

    return () => {
      if (rafId) {
        cancelAnimationFrame(rafId);
      }
      observer.disconnect();
      window.removeEventListener("resize", scheduleSync);
      window.clearInterval(pollId);
      window.clearTimeout(stopPollId);
      for (const anchorAd of [...resizeObservers.keys()]) {
        unobserveAnchorAd(anchorAd);
      }
      document.documentElement.style.removeProperty(FOOTER_NAV_HEIGHT_VAR);
      document.documentElement.style.removeProperty(ANCHOR_AD_HEIGHT_VAR);
    };
  }, []);
}
