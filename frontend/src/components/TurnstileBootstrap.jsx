import { useEffect, useRef } from "react";
import {
  isTurnstileConfigured,
  prepareTurnstileWidget,
  teardownTurnstileWidget,
} from "../utils/turnstileSession";

/**
 * Invisible Turnstile widget used to obtain tokens for GET /session on the API host.
 */
export default function TurnstileBootstrap() {
  const containerRef = useRef(null);

  useEffect(() => {
    if (!isTurnstileConfigured()) {
      return undefined;
    }

    let cancelled = false;
    const abort = new AbortController();
    prepareTurnstileWidget(containerRef.current, {
      signal: abort.signal,
    }).catch(() => {
      if (!cancelled) {
        // Session bootstrap will surface errors on the next search.
      }
    });

    return () => {
      cancelled = true;
      abort.abort();
      teardownTurnstileWidget();
    };
  }, []);

  if (!isTurnstileConfigured()) {
    return null;
  }

  return (
    <div
      ref={containerRef}
      className="turnstile-bootstrap"
      aria-hidden="true"
    />
  );
}
