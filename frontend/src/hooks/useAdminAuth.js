import { useCallback, useState } from "react";
import {
  ADMIN_LOGIN_URL,
  ADMIN_LOGOUT_URL,
  ADMIN_SESSION_URL,
} from "../constants";

async function parseJsonResponse(response) {
  try {
    return await response.json();
  } catch {
    return null;
  }
}

export default function useAdminAuth() {
  const [isSubmitting, setIsSubmitting] = useState(false);

  const checkSession = useCallback(async () => {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 5000);

    try {
      const response = await fetch(ADMIN_SESSION_URL, {
        method: "GET",
        credentials: "include",
        signal: controller.signal,
      });
      const payload = await parseJsonResponse(response);
      return {
        authenticated: Boolean(payload?.authenticated),
        enabled: payload?.enabled !== false,
        status: response.status,
      };
    } catch {
      return {
        authenticated: false,
        enabled: true,
        status: 0,
      };
    } finally {
      clearTimeout(timeoutId);
    }
  }, []);

  const login = useCallback(async (username, password) => {
    setIsSubmitting(true);
    try {
      const response = await fetch(ADMIN_LOGIN_URL, {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ username, password }),
      });
      const payload = await parseJsonResponse(response);

      if (response.status === 503 && payload?.enabled === false) {
        return { ok: false, error: "Admin is not configured yet." };
      }
      if (response.status === 429) {
        return {
          ok: false,
          error: "Too many login attempts. Try again later.",
        };
      }
      if (!response.ok) {
        return { ok: false, error: "Invalid username or password." };
      }

      return { ok: true };
    } catch {
      return { ok: false, error: "Unable to reach the admin login service." };
    } finally {
      setIsSubmitting(false);
    }
  }, []);

  const logout = useCallback(async () => {
    try {
      await fetch(ADMIN_LOGOUT_URL, {
        method: "POST",
        credentials: "include",
      });
    } catch {
      // Ignore network errors; the client should still return to the login page.
    }
  }, []);

  return {
    checkSession,
    login,
    logout,
    isSubmitting,
  };
}
