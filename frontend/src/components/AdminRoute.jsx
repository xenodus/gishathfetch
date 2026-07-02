import { useEffect, useState } from "react";
import { Navigate, Outlet } from "react-router-dom";
import useAdminAuth from "../hooks/useAdminAuth";

export default function AdminRoute() {
  const { checkSession } = useAdminAuth();
  const [state, setState] = useState({
    loading: true,
    authenticated: false,
    enabled: true,
  });

  useEffect(() => {
    let cancelled = false;

    checkSession()
      .then((result) => {
        if (!cancelled) {
          setState({
            loading: false,
            authenticated: result.authenticated,
            enabled: result.enabled,
          });
        }
      })
      .catch(() => {
        if (!cancelled) {
          setState({
            loading: false,
            authenticated: false,
            enabled: true,
          });
        }
      });

    return () => {
      cancelled = true;
    };
  }, [checkSession]);

  if (state.loading) {
    return (
      <div className="container py-5">
        <p className="text-center text-muted mb-0">Checking admin session…</p>
      </div>
    );
  }

  if (!state.enabled) {
    return (
      <div className="container py-5">
        <div className="alert alert-warning mb-0">
          Admin is not configured yet.
        </div>
      </div>
    );
  }

  if (!state.authenticated) {
    return <Navigate to="/admin/login" replace />;
  }

  return <Outlet />;
}
