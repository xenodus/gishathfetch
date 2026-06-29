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
}

export default function useBottomChromeLayout() {
  useEffect(() => {
    let rafId = 0;

    const scheduleSync = () => {
      if (rafId) {
        cancelAnimationFrame(rafId);
      }
      rafId = requestAnimationFrame(() => {
        rafId = 0;
        syncBottomChromeLayout();
      });
    };

    scheduleSync();

    const observer = new MutationObserver(scheduleSync);
    observer.observe(document.body, {
      childList: true,
      subtree: true,
      attributes: true,
      attributeFilter: ["style", "class", "data-anchor-status"],
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
      document.documentElement.style.removeProperty(FOOTER_NAV_HEIGHT_VAR);
      document.documentElement.style.removeProperty(ANCHOR_AD_HEIGHT_VAR);
    };
  }, []);
}
