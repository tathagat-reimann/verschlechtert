import { auth } from "./firebase";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL as string;

// Calls the Go backend, attaching the current user's Firebase ID token as a Bearer token.
export async function apiFetch(path: string, init: RequestInit = {}) {
  const user = auth.currentUser;
  const headers = new Headers(init.headers);

  if (user) {
    const token = await user.getIdToken();
    headers.set("Authorization", `Bearer ${token}`);
  }

  const response = await fetch(`${API_BASE_URL}${path}`, { ...init, headers });
  if (!response.ok) {
    throw new Error(await response.text());
  }
  return response;
}

export type Report = {
  id: number;
  description: string;
  observedAt?: string;
  status: string;
  createdAt: string;
  product: string;
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
  return response.json() as Promise<Me>;
}

export async function updateLocale(locale: string) {
  const response = await apiFetch("/api/me", {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ locale }),
  });
  return response.json() as Promise<Me>;
}

export type ReportsPage = { reports: Report[]; hasMore: boolean };

export async function getLatestReports(query = "", locale = "de", limit = 24, offset = 0) {
  const params = new URLSearchParams({ locale, limit: String(limit), offset: String(offset) });
  if (query) params.set("q", query);
  const response = await apiFetch(`/api/reports?${params.toString()}`);
  return response.json() as Promise<ReportsPage>;
}

export async function getReportDetail(id: number, locale = "de") {
  const response = await apiFetch(`/api/reports/${id}?${new URLSearchParams({ locale }).toString()}`);
  return response.json() as Promise<ReportDetail>;
}

export async function addReportComment(id: number, body: string) {
  const response = await apiFetch(`/api/reports/${id}/comments`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ body }),
  });
  return response.json() as Promise<Comment>;
}

export async function addReportAlternative(id: number, payload: { brandId: number; sellerId: number; productName: string; productUrl?: string }) {
  const response = await apiFetch(`/api/reports/${id}/alternatives`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  return response.json() as Promise<Alternative>;
}

export async function toggleReportLike(id: number) {
  const response = await apiFetch(`/api/reports/${id}/like`, { method: "POST" });
  return response.json() as Promise<{ liked: boolean; count: number }>;
}

export async function getCatalogOptions(locale = "de") {
  const response = await apiFetch(`/api/catalog/options?${new URLSearchParams({ locale }).toString()}`);
  return response.json() as Promise<CatalogOptions>;
}

export async function getMySubmissions(locale = "de") {
  const response = await apiFetch(`/api/me/submissions?${new URLSearchParams({ locale }).toString()}`);
  return response.json() as Promise<Submission[]>;
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
