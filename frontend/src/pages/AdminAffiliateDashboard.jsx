import { useCallback, useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { getApiBaseUrl } from "../constants";
import {
  clearStoredAdminApiKey,
  createAffiliateLink,
  deleteAffiliateLink,
  fetchAdminAffiliateLinks,
  fileToImagePayload,
  loadStoredAdminApiKey,
  storeAdminApiKey,
  updateAffiliateLink,
} from "../utils/affiliateAdminApi";

const EMPTY_FORM = {
  title: "",
  price: "",
  link: "",
  expiryDate: "",
  status: "active",
  imageFile: null,
  imageUrl: "",
};

function formatDate(value) {
  if (!value) {
    return "No expiry";
  }
  return value;
}

function LinkFormModal({
  show,
  initialValues,
  onClose,
  onSubmit,
  isSaving,
  error,
}) {
  const [form, setForm] = useState(EMPTY_FORM);

  useEffect(() => {
    if (!show) {
      return;
    }
    setForm({
      title: initialValues?.title || "",
      price: initialValues?.price || "",
      link: initialValues?.link || "",
      expiryDate: initialValues?.expiryDate || "",
      status: initialValues?.status || "active",
      imageFile: null,
      imageUrl: initialValues?.imageUrl || "",
    });
  }, [show, initialValues]);

  if (!show) {
    return null;
  }

  const handleChange = (event) => {
    const { name, value, files } = event.target;
    if (name === "imageFile") {
      setForm((current) => ({ ...current, imageFile: files?.[0] || null }));
      return;
    }
    setForm((current) => ({ ...current, [name]: value }));
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    await onSubmit(form);
  };

  return (
    <div className="modal show d-block" tabIndex="-1" role="dialog">
      <div className="modal-dialog modal-lg modal-dialog-scrollable">
        <form className="modal-content" onSubmit={handleSubmit}>
          <div className="modal-header">
            <h2 className="modal-title h5">
              {initialValues?.id ? "Edit Affiliate Link" : "Add Affiliate Link"}
            </h2>
            <button type="button" className="btn-close" onClick={onClose} />
          </div>
          <div className="modal-body">
            {error ? <div className="alert alert-danger">{error}</div> : null}
            <div className="row g-3">
              <div className="col-md-6">
                <label className="form-label" htmlFor="affiliate-title">
                  Title
                </label>
                <input
                  id="affiliate-title"
                  name="title"
                  className="form-control"
                  value={form.title}
                  onChange={handleChange}
                  placeholder="Deck box, sleeves, etc."
                />
              </div>
              <div className="col-md-6">
                <label className="form-label" htmlFor="affiliate-price">
                  Price
                </label>
                <input
                  id="affiliate-price"
                  name="price"
                  className="form-control"
                  value={form.price}
                  onChange={handleChange}
                  placeholder="S$24.90"
                  required
                />
              </div>
              <div className="col-12">
                <label className="form-label" htmlFor="affiliate-link">
                  Amazon link
                </label>
                <input
                  id="affiliate-link"
                  name="link"
                  type="url"
                  className="form-control"
                  value={form.link}
                  onChange={handleChange}
                  placeholder="https://www.amazon.sg/..."
                  required
                />
              </div>
              <div className="col-md-6">
                <label className="form-label" htmlFor="affiliate-expiry">
                  Expiry date
                </label>
                <input
                  id="affiliate-expiry"
                  name="expiryDate"
                  type="date"
                  className="form-control"
                  value={form.expiryDate}
                  onChange={handleChange}
                />
              </div>
              <div className="col-md-6">
                <label className="form-label" htmlFor="affiliate-status">
                  Status
                </label>
                <select
                  id="affiliate-status"
                  name="status"
                  className="form-select"
                  value={form.status}
                  onChange={handleChange}
                >
                  <option value="active">Active</option>
                  <option value="inactive">Inactive</option>
                </select>
              </div>
              <div className="col-md-6">
                <label className="form-label" htmlFor="affiliate-image-file">
                  Upload image
                </label>
                <input
                  id="affiliate-image-file"
                  name="imageFile"
                  type="file"
                  accept="image/jpeg,image/png,image/webp,image/gif"
                  className="form-control"
                  onChange={handleChange}
                />
              </div>
              <div className="col-md-6">
                <label className="form-label" htmlFor="affiliate-image-url">
                  Or image URL
                </label>
                <input
                  id="affiliate-image-url"
                  name="imageUrl"
                  type="url"
                  className="form-control"
                  value={form.imageUrl}
                  onChange={handleChange}
                  placeholder="https://..."
                />
              </div>
              {form.imageUrl ? (
                <div className="col-12">
                  <img
                    src={form.imageUrl}
                    alt="Current product"
                    className="affiliate-admin-preview"
                  />
                </div>
              ) : null}
            </div>
          </div>
          <div className="modal-footer">
            <button
              type="button"
              className="btn btn-outline-secondary"
              onClick={onClose}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="btn btn-primary"
              disabled={isSaving}
            >
              {isSaving ? "Saving..." : "Save"}
            </button>
          </div>
        </form>
      </div>
      <div className="modal-backdrop show" />
    </div>
  );
}

export default function AdminAffiliateDashboard() {
  const [apiKey, setApiKey] = useState(loadStoredAdminApiKey);
  const [apiKeyInput, setApiKeyInput] = useState("");
  const [links, setLinks] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState("");
  const [formError, setFormError] = useState("");
  const [editingLink, setEditingLink] = useState(null);
  const [showForm, setShowForm] = useState(false);

  const isAuthenticated = useMemo(() => apiKey.trim().length > 0, [apiKey]);

  const loadLinks = useCallback(async () => {
    if (!apiKey.trim()) {
      return;
    }
    setIsLoading(true);
    setError("");
    try {
      const data = await fetchAdminAffiliateLinks(getApiBaseUrl(), apiKey);
      setLinks(data);
    } catch (err) {
      setError(err.message || "Failed to load affiliate links.");
      if (
        String(err.message || "")
          .toLowerCase()
          .includes("unauthorized")
      ) {
        clearStoredAdminApiKey();
        setApiKey("");
      }
    } finally {
      setIsLoading(false);
    }
  }, [apiKey]);

  useEffect(() => {
    if (isAuthenticated) {
      loadLinks();
    }
  }, [isAuthenticated, loadLinks]);

  const handleLogin = (event) => {
    event.preventDefault();
    const trimmed = apiKeyInput.trim();
    if (!trimmed) {
      setError("API key is required.");
      return;
    }
    storeAdminApiKey(trimmed);
    setApiKey(trimmed);
    setApiKeyInput("");
    setError("");
  };

  const handleLogout = () => {
    clearStoredAdminApiKey();
    setApiKey("");
    setLinks([]);
  };

  const buildPayload = async (form) => {
    const payload = {
      title: form.title.trim(),
      price: form.price.trim(),
      link: form.link.trim(),
      expiryDate: form.expiryDate,
      status: form.status,
    };

    if (form.imageFile) {
      Object.assign(payload, await fileToImagePayload(form.imageFile));
    } else if (form.imageUrl.trim()) {
      payload.imageUrl = form.imageUrl.trim();
    } else if (!editingLink) {
      throw new Error("Image is required.");
    }

    return payload;
  };

  const handleCreate = async (form) => {
    setIsSaving(true);
    setFormError("");
    try {
      const payload = await buildPayload(form);
      await createAffiliateLink(getApiBaseUrl(), apiKey, payload);
      setShowForm(false);
      await loadLinks();
    } catch (err) {
      setFormError(err.message || "Failed to create affiliate link.");
    } finally {
      setIsSaving(false);
    }
  };

  const handleUpdate = async (form) => {
    setIsSaving(true);
    setFormError("");
    try {
      const payload = await buildPayload(form);
      if (!payload.imageUrl && !payload.imageData && editingLink?.imageUrl) {
        payload.imageUrl = editingLink.imageUrl;
      }
      await updateAffiliateLink(
        getApiBaseUrl(),
        apiKey,
        editingLink.id,
        payload,
      );
      setEditingLink(null);
      setShowForm(false);
      await loadLinks();
    } catch (err) {
      setFormError(err.message || "Failed to update affiliate link.");
    } finally {
      setIsSaving(false);
    }
  };

  const handleDelete = async (link) => {
    if (!window.confirm(`Delete "${link.title || link.id}"?`)) {
      return;
    }
    setError("");
    try {
      await deleteAffiliateLink(getApiBaseUrl(), apiKey, link.id);
      await loadLinks();
    } catch (err) {
      setError(err.message || "Failed to delete affiliate link.");
    }
  };

  if (!isAuthenticated) {
    return (
      <div className="container py-5" style={{ maxWidth: "480px" }}>
        <h1 className="h3 mb-3">Affiliate Links Admin</h1>
        <p className="text-secondary">
          Enter the admin API key configured in the Lambda environment.
        </p>
        {error ? <div className="alert alert-danger">{error}</div> : null}
        <form onSubmit={handleLogin} className="card card-body shadow-sm">
          <label className="form-label" htmlFor="admin-api-key">
            API key
          </label>
          <input
            id="admin-api-key"
            type="password"
            className="form-control mb-3"
            value={apiKeyInput}
            onChange={(event) => setApiKeyInput(event.target.value)}
            autoComplete="current-password"
          />
          <button type="submit" className="btn btn-primary">
            Sign in
          </button>
        </form>
        <div className="mt-3">
          <Link to="/">Back to search</Link>
        </div>
      </div>
    );
  }

  return (
    <div className="container-xl py-4">
      <div className="d-flex flex-wrap justify-content-between align-items-center gap-2 mb-4">
        <div>
          <h1 className="h3 mb-1">Affiliate Links Admin</h1>
          <p className="text-secondary mb-0">
            Manage Amazon affiliate products shown on the homepage.
          </p>
        </div>
        <div className="d-flex gap-2">
          <Link to="/" className="btn btn-outline-secondary">
            Back to search
          </Link>
          <button
            type="button"
            className="btn btn-outline-danger"
            onClick={handleLogout}
          >
            Sign out
          </button>
          <button
            type="button"
            className="btn btn-primary"
            onClick={() => {
              setEditingLink(null);
              setFormError("");
              setShowForm(true);
            }}
          >
            Add link
          </button>
        </div>
      </div>

      {error ? <div className="alert alert-danger">{error}</div> : null}

      <div className="table-responsive card shadow-sm">
        <table className="table table-striped mb-0 align-middle">
          <thead>
            <tr>
              <th>Image</th>
              <th>Title</th>
              <th>Price</th>
              <th>Status</th>
              <th>Expiry</th>
              <th>Link</th>
              <th className="text-end">Actions</th>
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              <tr>
                <td colSpan={7} className="text-center py-4">
                  Loading...
                </td>
              </tr>
            ) : links.length === 0 ? (
              <tr>
                <td colSpan={7} className="text-center py-4 text-secondary">
                  No affiliate links yet.
                </td>
              </tr>
            ) : (
              links.map((link) => (
                <tr key={link.id}>
                  <td>
                    <img
                      src={link.imageUrl}
                      alt={link.title || "Product"}
                      className="affiliate-admin-thumb"
                    />
                  </td>
                  <td>{link.title || "—"}</td>
                  <td>{link.price}</td>
                  <td>
                    <span
                      className={`badge ${link.status === "active" ? "text-bg-success" : "text-bg-secondary"}`}
                    >
                      {link.status}
                    </span>
                  </td>
                  <td>{formatDate(link.expiryDate)}</td>
                  <td className="text-truncate" style={{ maxWidth: "220px" }}>
                    <a href={link.link} target="_blank" rel="noreferrer">
                      {link.link}
                    </a>
                  </td>
                  <td className="text-end">
                    <div className="btn-group btn-group-sm">
                      <button
                        type="button"
                        className="btn btn-outline-primary"
                        onClick={() => {
                          setEditingLink(link);
                          setFormError("");
                          setShowForm(true);
                        }}
                      >
                        Edit
                      </button>
                      <button
                        type="button"
                        className="btn btn-outline-danger"
                        onClick={() => handleDelete(link)}
                      >
                        Delete
                      </button>
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      <LinkFormModal
        show={showForm}
        initialValues={editingLink}
        onClose={() => {
          setShowForm(false);
          setEditingLink(null);
          setFormError("");
        }}
        onSubmit={editingLink ? handleUpdate : handleCreate}
        isSaving={isSaving}
        error={formError}
      />
    </div>
  );
}
