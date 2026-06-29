import { useEffect } from "react";

const ANCHOR_AD_SELECTOR = "ins.adsbygoogle.adsbygoogle-noablate";
const CSS_VAR = "--anchor-ad-offset";

function measureAnchorAdOffset() {
  const anchorAd = document.querySelector(ANCHOR_AD_SELECTOR);
  if (!anchorAd || !anchorAd.isConnected) {
    return 0;
  }

  const status = anchorAd.getAttribute("data-anchor-status");
  if (status === "dismissed") {
    return 0;
  }

  const rect = anchorAd.getBoundingClientRect();
  if (rect.height <= 0 || rect.bottom < window.innerHeight - 2) {
    return 0;
  }

  return Math.ceil(rect.height);
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
