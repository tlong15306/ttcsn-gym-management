import {
  createContext,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";
import type { Session, User } from "../types";
import { api } from "../api";
import { clearDraft } from "../features/booking/draft";

type AuthContextValue = {
  session: Session | null;
  notice: string;
  login: (s: Session) => void;
  logout: () => Promise<void>;
  updateUser: (u: User) => void;
  expire: () => void;
};
const AuthContext = createContext<AuthContextValue | null>(null);
export function SessionProvider({ children }: { children: ReactNode }) {
  const [session, setSession] = useState<Session | null>(null);
  const [notice, setNotice] = useState("");
  function expire() {
    setSession(null);
    setNotice("Phiên đã hết hạn hoặc bị thu hồi. Vui lòng đăng nhập lại.");
  }
  useEffect(() => {
    window.addEventListener("session-expired", expire);
    return () => window.removeEventListener("session-expired", expire);
  }, []);
  useEffect(() => {
    if (!session) return;
    const timer = setTimeout(expire, session.expiresIn * 1000);
    return () => clearTimeout(timer);
  }, [session?.accessToken]);
  async function logout() {
    if (session) await api.logout(session.accessToken);
    clearDraft();
    setSession(null);
    setNotice("Đã đăng xuất tất cả phiên của tài khoản.");
  }
  return (
    <AuthContext.Provider
      value={{
        session,
        notice,
        expire,
        login: (s) => {
          setSession(s);
          setNotice("");
        },
        logout,
        updateUser: (u) => setSession((s) => (s ? { ...s, user: u } : s)),
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}
export function useSession() {
  const value = useContext(AuthContext);
  if (!value) throw new Error("Missing SessionProvider");
  return value;
}
export function useAuth() {
  const auth = useSession();
  if (!auth.session) throw new Error("Missing session");
  return { ...auth, session: auth.session };
}
export const roleLabels = {
  ADMIN: "Quản trị viên",
  MANAGER: "Quản lý",
  RECEPTIONIST: "Lễ tân",
  MEMBER: "Hội viên",
  PT: "Huấn luyện viên",
};
