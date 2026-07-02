import { Button } from "react-bootstrap";
import { Link, useNavigate } from "react-router-dom";
import useAdminAuth from "../hooks/useAdminAuth";

export default function AdminDashboard() {
  const navigate = useNavigate();
  const { logout } = useAdminAuth();

  const handleLogout = async () => {
    await logout();
    navigate("/admin/login", { replace: true });
  };

  return (
    <div className="container py-5">
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h1 className="h3 mb-0">Admin</h1>
        <div className="d-flex gap-2">
          <Button as={Link} to="/" variant="outline-secondary">
            Back to search
          </Button>
          <Button variant="primary" onClick={handleLogout}>
            Log out
          </Button>
        </div>
      </div>

      <p className="text-muted mb-0">
        Admin landing page. Content will be added here later.
      </p>
    </div>
  );
}
