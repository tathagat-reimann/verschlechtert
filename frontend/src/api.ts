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

  return fetch(`${API_BASE_URL}${path}`, { ...init, headers });
}
