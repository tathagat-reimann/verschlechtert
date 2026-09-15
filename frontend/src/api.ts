import { auth } from "./firebase";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL as string;

// Calls the Go backend, attaching the current user's Firebase ID token as a Bearer token.
type ApiEnvelope<T> = { data: T };
type ApiErrorBody = { code?: string; message?: string; correlationId?: string; error?: ApiErrorBody };

export async function apiFetch(path: string, init: RequestInit = {}) {
  const user = auth.currentUser;
  const headers = new Headers(init.headers);

  if (user) {
    const token = await user.getIdToken();
    headers.set("Authorization", `Bearer ${token}`);
  }

  const response = await fetch(`${API_BASE_URL}${path}`, { ...init, headers });
  if (!response.ok) {
    const payload = await response.json().catch(() => null) as ApiErrorBody | null;
    const errorPayload = payload?.error ?? payload;
    throw new Error(errorPayload?.message ?? response.statusText ?? "Request failed");
  }
  return response;
}

async function unwrapData<T>(response: Response): Promise<T> {
  const payload = await response.json() as ApiEnvelope<T> | T;
  return (payload as ApiEnvelope<T>).data ?? (payload as T);
}

export type Report = {
  id: number;
  description: string;
  observedAt?: string;
  status: string;
  createdAt: string;
  productName: string;
  brand: string;
  category: string;
  seller: string;
  likeCount: number;
  likedByMe: boolean;
};

export type Submission = {
  id: number;
  description: string;
  observedAt?: string;
  status: string;
  createdAt: string;
  product: string;
  brand: string;
  category: string;
  seller: string;
  brandId: number;
  categoryId: number;
  sellerId: number;
  likeCount: number;
  likedByMe: boolean;
};

export type Comment = {
  id: number;
  body: string;
  authorName: string;
  createdAt: string;
};

export type Alternative = {
  id: number;
  productName: string;
  brand: string;
  seller: string;
  productUrl?: string;
  authorName: string;
  createdAt: string;
  suggestedByMe: boolean;
};

export type ReportImage = { imageUrl: string; sortOrder: number };

export type ReportDetail = {
  id: number;
  description: string;
  observedAt?: string;
  status: string;
  createdAt: string;
  product: string;
  brand: string;
  category: string;
  seller: string;
  productUrl?: string;
  images: ReportImage[];
  comments: Comment[];
  alternatives: Alternative[];
  likeCount: number;
  likedByMe: boolean;
  isOwner: boolean;
  hasAlternative: boolean;
};

export type CatalogOption = { id: number; name: string; slug?: string };

export type CatalogOptions = {
  brands: CatalogOption[];
  categories: CatalogOption[];
  sellers: CatalogOption[];
  unspecifiedBrand: CatalogOption;
  unspecifiedCategory: CatalogOption;
  unspecifiedSeller: CatalogOption;
};

export type Me = {
  id: number;
  firebaseUid: string;
  displayName: string;
  photoUrl: string;
  email: string;
  locale: string;
};

export async function getMe() {
  const response = await apiFetch("/api/me");
  return unwrapData<Me>(response);
}

export async function updateLocale(locale: string) {
  const response = await apiFetch("/api/me", {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ locale }),
  });
  return unwrapData<Me>(response);
}

export type ReportsPage = { reports: Report[]; hasMore: boolean };

export async function getLatestReports(query = "", locale = "de", limit = 24, offset = 0) {
  const params = new URLSearchParams({ locale, limit: String(limit), offset: String(offset) });
  if (query) params.set("q", query);
  const response = await apiFetch(`/api/reports?${params.toString()}`);
  return unwrapData<ReportsPage>(response);
}

export async function getReportDetail(id: number, _locale = "de") {
  const response = await apiFetch(`/api/reports/${id}`);
  return unwrapData<ReportDetail>(response);
}

export async function addReportComment(id: number, body: string) {
  const response = await apiFetch(`/api/reports/${id}/comments`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ body }),
  });
  return unwrapData<Comment>(response);
}

export async function addReportAlternative(id: number, payload: { brandId: number; sellerId: number; productName: string; productUrl?: string }) {
  const response = await apiFetch(`/api/reports/${id}/alternatives`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  return unwrapData<Alternative>(response);
}

export async function toggleReportLike(id: number) {
  const response = await apiFetch(`/api/reports/${id}/like`, { method: "POST" });
  return unwrapData<{ liked: boolean; count: number }>(response);
}

export async function getCatalogOptions(_locale = "de") {
  const response = await apiFetch(`/api/catalog/options`);
  return unwrapData<CatalogOptions>(response);
}

export async function getMySubmissions(_locale = "de") {
  const response = await apiFetch(`/api/me/submissions`);
  return unwrapData<Submission[]>(response);
}

export async function createSubmission(payload: Record<string, unknown>) {
  return apiFetch("/api/submissions", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
}

export async function updateSubmission(id: number, payload: Record<string, unknown>) {
  return apiFetch(`/api/me/submissions/${id}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
}
