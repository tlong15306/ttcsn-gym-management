import type { PageMeta, Role, Session, Status, User } from "./types";

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    public readonly code: string,
    message: string,
  ) {
    super(message);
  }
}

export type Envelope<T> = { data: T; meta?: PageMeta; request_id: string };

export async function request<T>(
  path: string,
  init: RequestInit = {},
  token?: string,
): Promise<Envelope<T>> {
  const headers = new Headers(init.headers);
  if (!(init.body instanceof FormData))
    headers.set("Content-Type", "application/json");
  if (token) headers.set("Authorization", `Bearer ${token}`);

  const response = await fetch(path, { ...init, headers });
  if (response.status === 204)
    return {
      data: undefined as T,
      request_id: response.headers.get("X-Request-ID") ?? "",
    };
  const body = await response.json().catch(() => null);
  if (!response.ok) {
    if (response.status === 401 && token)
      window.dispatchEvent(new Event("session-expired"));
    throw new ApiError(
      response.status,
      body?.error?.code ?? "NETWORK_ERROR",
      body?.error?.message ?? "Không thể kết nối tới máy chủ.",
    );
  }
  return body as Envelope<T>;
}

export const resource = {
  list<T>(
    token: string,
    path: string,
    query: URLSearchParams,
    signal?: AbortSignal,
  ) {
    return request<T[]>(`/api/v1/${path}${path.includes('?') ? '&' : '?'}${query}`, { signal }, token);
  },
  get<T>(token: string, path: string, signal?: AbortSignal) {
    return request<T>(`/api/v1/${path}`, { signal }, token);
  },
  save<T>(token: string, path: string, input: unknown, method = "POST") {
    return request<T>(
      `/api/v1/${path}`,
      { method, body: JSON.stringify(input) },
      token,
    );
  },
};

export const api = {
  async login(username: string, password: string): Promise<Session> {
    const result = await request<{
      access_token: string;
      expires_in: number;
      user: User;
    }>("/api/v1/auth/login", {
      method: "POST",
      body: JSON.stringify({ username, password }),
    });
    return {
      accessToken: result.data.access_token,
      expiresIn: result.data.expires_in,
      user: result.data.user,
    };
  },
  logout(token: string) {
    return request<void>("/api/v1/auth/logout", { method: "POST" }, token);
  },
  me(token: string) {
    return request<User>("/api/v1/me", {}, token);
  },
  updateMe(token: string, input: Pick<User, "full_name" | "email" | "phone">) {
    return request<User>(
      "/api/v1/me",
      { method: "PATCH", body: JSON.stringify(input) },
      token,
    );
  },
  users(
    token: string,
    filters: {
      q?: string;
      role?: Role | "";
      status?: Status | "";
      page?: number;
    } = {},
  ) {
    const params = new URLSearchParams();
    if (filters.q) params.set("q", filters.q);
    if (filters.role) params.set("role", filters.role);
    if (filters.status) params.set("status", filters.status);
    params.set("page", String(filters.page ?? 1));
    return request<User[]>(`/api/v1/users?${params.toString()}`, {}, token);
  },
  createUser(
    token: string,
    input: {
      username: string;
      password: string;
      email: string;
      phone: string;
      full_name: string;
      role: Role;
    },
  ) {
    return request<User>(
      "/api/v1/users",
      { method: "POST", body: JSON.stringify(input) },
      token,
    );
  },
  changeStatus(token: string, id: number, status: Status, reason: string) {
    return request<User>(
      `/api/v1/users/${id}/status`,
      { method: "PATCH", body: JSON.stringify({ status, reason }) },
      token,
    );
  },
  changeRole(token: string, id: number, role: Role, reason: string) {
    return request<User>(
      `/api/v1/users/${id}/role`,
      { method: "PATCH", body: JSON.stringify({ role, reason }) },
      token,
    );
  },
  resetPassword(token: string, id: number, password: string) {
    return request<User>(
      `/api/v1/users/${id}/reset-password`,
      { method: "POST", body: JSON.stringify({ password }) },
      token,
    );
  },
};
