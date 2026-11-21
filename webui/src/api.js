import CONFIG from "../config/config.js";

function authHeaders() {
  const token = localStorage.getItem("admin_token");
  return token ? { Authorization: `Bearer ${token}` } : {};
}

async function request(method, path, body = null, isFormData = false) {
  const headers = { ...authHeaders() };
  if (body && !isFormData) headers["Content-Type"] = "application/json";

  const res = await fetch(`${CONFIG.BASE_URL}${path}`, {
    method,
    headers,
    body: body
      ? isFormData
        ? body
        : JSON.stringify(body)
      : undefined,
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }));
    const err = new Error(body.error || `HTTP ${res.status}`);
    err.status = res.status;
    throw err;
  }

  // Binary response (media files)
  if (res.headers.get("Content-Type")?.startsWith("image/") ||
      res.headers.get("Content-Type") === "application/pdf") {
    return res.blob();
  }

  return res.json();
}

// ── Projects ──────────────────────────────────────────────────────────────────
export const api = {
  projects: {
    list: () => request("GET", "/projects"),
    get: (id) => request("GET", `/projects/${id}`),
    // size: "small" or "large" for a web copy (800 / 1600 px wide), omitted for the original.
    mediaUrl: (projectId, mediaId, size) =>
      `${CONFIG.BASE_URL}/projects/${projectId}/media/${mediaId}${size ? `?size=${size}` : ""}`,

    // Admin
    create: (formData) => request("POST", "/admin/projects", formData, true),
    update: (id, data) => request("PUT", `/admin/projects/${id}`, data),
    delete: (id) => request("DELETE", `/admin/projects/${id}`),
    addMedia: (id, formData) =>
      request("POST", `/admin/projects/${id}/media`, formData, true),
    deleteMedia: (projectId, mediaId) =>
      request("DELETE", `/admin/projects/${projectId}/media/${mediaId}`),
    reorder: (category, projectIds) =>
      request("PUT", "/admin/projects/order", { category, project_ids: projectIds }),
    reorderMedia: (projectId, mediaIds) =>
      request("PUT", `/admin/projects/${projectId}/media/order`, { media_ids: mediaIds }),
  },

  requests: {
    create: (formData) => request("POST", "/requests", formData, true),

    // Admin
    list: () => request("GET", "/admin/requests"),
    get: (id) => request("GET", `/admin/requests/${id}`),
    attachmentUrl: (requestId, attachmentId) =>
      `${CONFIG.BASE_URL}/admin/requests/${requestId}/attachments/${attachmentId}`,
  },

  contacts: {
    get: () => request("GET", "/contacts"),
    photoUrl: () => `${CONFIG.BASE_URL}/about/photo`,
    update: (data) => request("PUT", "/admin/contacts", data),
  },

  faq: {
    list: () => request("GET", "/faq"),
  },

  auth: {
    requestOTP: (email) => request("POST", "/auth/request-otp", { email }),
    verifyOTP: (email, code) =>
      request("POST", "/auth/verify-otp", { email, code }),
  },
};
