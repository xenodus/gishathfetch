import { useEffect, useRef } from "react";
import {
  isTurnstileConfigured,
  prepareTurnstileWidget,
  teardownTurnstileWidget,
} from "../utils/turnstileSession";

const PREPARE_RETRY_DELAY_MS = 1500;
const PREPARE_MAX_ATTEMPTS = 3;

/**
 * Invisible Turnstile widget used to obtain tokens for GET /session on the API host.
 * Retries prepare a few times — script/challenge load is flaky under private browsing.
 */
export default function TurnstileBootstrap() {
  const containerRef = useRef(null);

  useEffect(() => {
    if (!isTurnstileConfigured()) {
      return undefined;
    }

    let cancelled = false;
    const abort = new AbortController();
    let retryTimer = null;

    const prepareWithRetry = async (attempt) => {
      try {
        await prepareTurnstileWidget(containerRef.current, {
          signal: abort.signal,
        });
      } catch {
        if (cancelled || abort.signal.aborted) {
          return;
        }
        if (attempt < PREPARE_MAX_ATTEMPTS) {
          retryTimer = setTimeout(() => {
            prepareWithRetry(attempt + 1);
          }, PREPARE_RETRY_DELAY_MS * attempt);
          return;
        }
        // Session bootstrap will surface errors on the next search.
      }
    };

    prepareWithRetry(1);

    return () => {
      cancelled = true;
      abort.abort();
      if (retryTimer != null) {
        clearTimeout(retryTimer);
      }
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
