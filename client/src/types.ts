export type Role = "MEMBER" | "PT" | "RECEPTIONIST" | "MANAGER" | "ADMIN";
export type Status = "ACTIVE" | "LOCKED";

export interface User {
  id: number;
  username: string;
  email: string;
  phone?: string;
  full_name: string;
  role: Role;
  status: Status;
  created_at: string;
  updated_at: string;
}

export interface Session {
  accessToken: string;
  expiresIn: number;
  user: User;
}

export interface PageMeta {
  page: number;
  page_size: number;
  total: number;
}
