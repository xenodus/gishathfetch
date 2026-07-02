import { useEffect, useState } from "react";
import { Alert, Button, Card, Form } from "react-bootstrap";
import { Link, useNavigate } from "react-router-dom";
import useAdminAuth from "../hooks/useAdminAuth";

export default function AdminLogin() {
  const navigate = useNavigate();
  const { login, checkSession, isSubmitting } = useAdminAuth();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [adminEnabled, setAdminEnabled] = useState(true);

  useEffect(() => {
    let cancelled = false;

    checkSession()
      .then((result) => {
        if (cancelled) {
          return;
        }
        if (result.authenticated) {
          navigate("/admin", { replace: true });
          return;
        }
        setAdminEnabled(result.enabled);
      })
      .catch(() => {
        if (!cancelled) {
          setAdminEnabled(true);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [checkSession, navigate]);

  const handleSubmit = async (event) => {
    event.preventDefault();
    setError("");

    const result = await login(username, password);
    if (!result.ok) {
      setError(result.error);
      return;
    }

    navigate("/admin", { replace: true });
  };

  return (
    <div className="container py-5" style={{ maxWidth: "480px" }}>
      <Card>
        <Card.Body className="p-4">
          <h1 className="h4 mb-3">Admin login</h1>
          <p className="text-muted">
            Sign in to access the Gishath Fetch admin area.
          </p>

          {!adminEnabled ? (
            <Alert variant="warning">Admin is not configured yet.</Alert>
          ) : null}
          {error ? <Alert variant="danger">{error}</Alert> : null}

          <Form onSubmit={handleSubmit}>
            <Form.Group className="mb-3" controlId="admin-username">
              <Form.Label>Username</Form.Label>
              <Form.Control
                autoComplete="username"
                value={username}
                onChange={(event) => setUsername(event.target.value)}
                required
                disabled={!adminEnabled}
              />
            </Form.Group>

            <Form.Group className="mb-4" controlId="admin-password">
              <Form.Label>Password</Form.Label>
              <Form.Control
                type="password"
                autoComplete="current-password"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                required
                disabled={!adminEnabled}
              />
            </Form.Group>

            <div className="d-grid gap-2">
              <Button type="submit" disabled={isSubmitting || !adminEnabled}>
                {isSubmitting ? "Signing in…" : "Sign in"}
              </Button>
              <Button as={Link} to="/" variant="outline-secondary">
                Back to search
              </Button>
            </div>
          </Form>
        </Card.Body>
      </Card>
    </div>
  );
}
