import { useEffect, useRef } from "react";
import {
  isTurnstileConfigured,
  prepareTurnstileWidget,
} from "../utils/turnstileSession";

/**
 * Invisible Turnstile widget used to obtain tokens for GET /api/session.
 */
export default function TurnstileBootstrap() {
  const containerRef = useRef(null);

  useEffect(() => {
    if (!isTurnstileConfigured()) {
      return undefined;
    }

    let cancelled = false;
    prepareTurnstileWidget(containerRef.current).catch(() => {
      if (!cancelled) {
        // Session bootstrap will surface errors on the next search.
      }
    });

    return () => {
      cancelled = true;
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
