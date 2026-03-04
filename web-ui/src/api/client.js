import axios from "axios";

const BASE = import.meta.env.VITE_API_URL || "http://localhost:4000";

const api = axios.create({ baseURL: BASE, timeout: 10000 });

export const agents = {
  create: (data) => api.post("/agents", data),
  list: () => api.get("/agents"),
  get: (id) => api.get(`/agents/${id}`),
};

export const scans = {
  start: (agentId, sourcePath) => api.post(`/agents/${agentId}/scan`, { sourcePath }),
  get: (scanId) => api.get(`/scans/${scanId}`),
  tree: (scanId) => api.get(`/scans/${scanId}/tree`),
};

export default api;
