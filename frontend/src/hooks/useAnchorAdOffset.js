import { useEffect } from "react";

const ANCHOR_AD_SELECTOR = "ins.adsbygoogle.adsbygoogle-noablate";
const CSS_VAR = "--anchor-ad-offset";

function parsePixelValue(value) {
  const parsed = Number.parseFloat(value);
  return Number.isFinite(parsed) ? parsed : 0;
}

function measureAnchorAdOffset() {
  const anchorAds = document.querySelectorAll(ANCHOR_AD_SELECTOR);
  let maxOffset = 0;

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

    const isBottomAnchored =
      computed.position === "fixed" &&
      parsePixelValue(computed.bottom) <= 4 &&
      rect.bottom >= window.innerHeight - 4;

    if (!isBottomAnchored) continue;

    maxOffset = Math.max(maxOffset, Math.ceil(height));
  }

  return maxOffset;
}

function syncAnchorAdOffset() {
  const offset = measureAnchorAdOffset();
  document.documentElement.style.setProperty(CSS_VAR, `${offset}px`);
  return offset;
}

export default function useAnchorAdOffset() {
  useEffect(() => {
    let rafId = 0;

    const scheduleSync = () => {
      if (rafId) {
        cancelAnimationFrame(rafId);
      }
      rafId = requestAnimationFrame(() => {
        rafId = 0;
        syncAnchorAdOffset();
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
      document.documentElement.style.removeProperty(CSS_VAR);
    };
  }, []);
}
