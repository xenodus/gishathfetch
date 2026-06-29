const ADMIN_API_KEY_STORAGE_KEY = "gishathfetch-affiliate-admin-api-key";

function getHeaders(apiKey) {
  return {
    "Content-Type": "application/json",
    Authorization: `Bearer ${apiKey}`,
  };
}

async function parseError(response) {
  try {
    const payload = await response.json();
    if (typeof payload?.error === "string" && payload.error.trim()) {
      return payload.error;
    }
  } catch {
    // Ignore JSON parse failures and fall back to status text.
  }
  return `Request failed (${response.status})`;
}

export function loadStoredAdminApiKey() {
  try {
    return sessionStorage.getItem(ADMIN_API_KEY_STORAGE_KEY) || "";
  } catch {
    return "";
  }
}

export function storeAdminApiKey(apiKey) {
  try {
    sessionStorage.setItem(ADMIN_API_KEY_STORAGE_KEY, apiKey);
  } catch {
    // Ignore storage failures; caller can keep the key in memory only.
  }
}

export function clearStoredAdminApiKey() {
  try {
    sessionStorage.removeItem(ADMIN_API_KEY_STORAGE_KEY);
  } catch {
    // Ignore storage failures.
  }
}

export async function fetchAdminAffiliateLinks(apiBaseUrl, apiKey) {
  const response = await fetch(`${apiBaseUrl}admin/affiliate-links`, {
    headers: getHeaders(apiKey),
  });
  if (!response.ok) {
    throw new Error(await parseError(response));
  }
  const payload = await response.json();
  return Array.isArray(payload?.data) ? payload.data : [];
}

export async function createAffiliateLink(apiBaseUrl, apiKey, body) {
  const response = await fetch(`${apiBaseUrl}admin/affiliate-links`, {
    method: "POST",
    headers: getHeaders(apiKey),
    body: JSON.stringify(body),
  });
  if (!response.ok) {
    throw new Error(await parseError(response));
  }
  return response.json();
}

export async function updateAffiliateLink(apiBaseUrl, apiKey, id, body) {
  const response = await fetch(`${apiBaseUrl}admin/affiliate-links/${id}`, {
    method: "PUT",
    headers: getHeaders(apiKey),
    body: JSON.stringify(body),
  });
  if (!response.ok) {
    throw new Error(await parseError(response));
  }
  return response.json();
}

export async function deleteAffiliateLink(apiBaseUrl, apiKey, id) {
  const response = await fetch(`${apiBaseUrl}admin/affiliate-links/${id}`, {
    method: "DELETE",
    headers: getHeaders(apiKey),
  });
  if (!response.ok) {
    throw new Error(await parseError(response));
  }
}

export async function fileToImagePayload(file) {
  if (!file) {
    return {};
  }

  const allowedTypes = ["image/jpeg", "image/png", "image/webp", "image/gif"];
  if (!allowedTypes.includes(file.type)) {
    throw new Error("Image must be JPEG, PNG, WebP, or GIF.");
  }
  if (file.size > 2 * 1024 * 1024) {
    throw new Error("Image must be 2 MB or smaller.");
  }

  const buffer = await file.arrayBuffer();
  const bytes = new Uint8Array(buffer);
  let binary = "";
  for (const byte of bytes) {
    binary += String.fromCharCode(byte);
  }

  return {
    imageData: btoa(binary),
    imageContentType: file.type,
  };
}
