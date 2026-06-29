import { useCallback, useRef, useState } from "react";

export default function useAdminSession() {
  const credentialRef = useRef("");
  const loginInputRef = useRef(null);
  const [isAuthenticated, setIsAuthenticated] = useState(false);

  const signIn = useCallback(() => {
    const value = loginInputRef.current?.value?.trim() ?? "";
    if (!value) {
      return false;
    }

    credentialRef.current = value;
    if (loginInputRef.current) {
      loginInputRef.current.value = "";
    }
    setIsAuthenticated(true);
    return true;
  }, []);

  const signOut = useCallback(() => {
    credentialRef.current = "";
    setIsAuthenticated(false);
  }, []);

  const getAuthorizationHeader = useCallback(
    () => `Bearer ${credentialRef.current}`,
    [],
  );

  return {
    loginInputRef,
    isAuthenticated,
    signIn,
    signOut,
    getAuthorizationHeader,
  };
}
