import { useState, type FormEvent } from "react";
import { Dumbbell, ArrowRight, ShieldCheck } from "lucide-react";
import { api } from "../../api";
import type { Session } from "../../types";
import {
  Button,
  Field,
  Input,
  Notice,
  PasswordInput,
} from "../../components/ui";

export function LoginPanel({
  notice,
  onLogin,
}: {
  notice: string;
  onLogin: (s: Session) => void;
}) {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  async function submit(e: FormEvent) {
    e.preventDefault();
    if (busy) return;
    setBusy(true);
    setError("");
    try {
      onLogin(await api.login(username, password));
    } catch (e) {
      setError(e instanceof Error ? e.message : "Không thể đăng nhập.");
    } finally {
      setBusy(false);
    }
  }
  return (
    <main className="login-page">
      <section className="login-story">
        <div className="brand">
          <span className="brand-symbol">
            <Dumbbell size={24} />
          </span>
          <span>
            GYM<span className="brand-caption">MANAGEMENT</span>
          </span>
        </div>
        <div className="login-story-content">
          <span className="eyebrow">KHÔNG GIAN QUẢN LÝ</span>
          <h1>
            Vận hành phòng tập.
            <br />
            Gọn gàng mỗi ngày.
          </h1>
          <p>
            Tài khoản, hội viên, phòng tập và huấn luyện viên trong cùng một nơi
            làm việc.
          </p>
          <div className="story-line" />
          <div className="security-note">
            <ShieldCheck size={22} />
            <span>
              Quyền truy cập phù hợp
              <br />
              <small>với từng vai trò của bạn</small>
            </span>
          </div>
        </div>
        <span className="login-footnote">TTCSN · Hệ thống quản lý GYM</span>
      </section>
      <section className="login-form-area">
        <div className="login-card">
          <a className="back-link" href="/">← Về trang chủ</a>
          <span className="eyebrow">CHÀO MỪNG TRỞ LẠI</span>
          <h2>Đăng nhập</h2>
          <p className="subtle">Nhập tài khoản để vào không gian của bạn.</p>
          <Notice tone="info">{notice}</Notice>
          <Notice>{error}</Notice>
          <form onSubmit={submit}>
            <Field label="Tên đăng nhập" required>
              <Input
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                autoComplete="username"
                placeholder="Nhập tên đăng nhập"
                required
                disabled={busy}
              />
            </Field>
            <Field label="Mật khẩu" required>
              <PasswordInput
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                autoComplete="current-password"
                placeholder="Nhập mật khẩu"
                required
                disabled={busy}
              />
            </Field>
            <Button className="login-submit" busy={busy} type="submit">
              {busy ? "Đang đăng nhập…" : "Đăng nhập"}
              {!busy && <ArrowRight size={18} />}
            </Button>
          </form>
          <p className="login-help">
            Quên mật khẩu? Liên hệ quản trị viên để được hỗ trợ.
          </p>
        </div>
        <p className="login-session-note">
          Phiên chỉ được giữ trong lần mở trang này.
        </p>
      </section>
    </main>
  );
}
