import axios from "axios";

const API_BASE = process.env.REACT_APP_API_URL || "http://localhost:8080/api/v1";

const api = axios.create({
  baseURL: API_BASE,
  headers: { "Content-Type": "application/json" },
});

export interface PendingItem {
  data_id: string;
  labels: Record<string, unknown>;
  confidence: number;
}

export interface ValidationPayload {
  data_id: string;
  validator_id: string;
  action: "accept" | "modify" | "reject";
  labels?: Record<string, unknown>;
  confidence?: number;
  notes?: string;
}

export interface FinalLabel {
  final_label: Record<string, unknown>;
  confidence: number;
  source: string[];
}

export const fetchPending = () =>
  api.get<PendingItem[]>("/validation/pending").then((r) => r.data);

export const submitValidation = (payload: ValidationPayload) =>
  api.post("/validation/submit", payload).then((r) => r.data);

export const fetchLabel = (dataId: string) =>
  api.get<FinalLabel>(`/label/${dataId}`).then((r) => r.data);

export const uploadData = (type: string, data: Record<string, unknown>) =>
  api.post("/data/upload", { type, data }).then((r) => r.data);
