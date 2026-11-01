import { NavLink, Outlet } from "react-router-dom";
import { useState } from "react";
import { useAuth } from "../app/session";
import { Button, Modal, Notice } from "../components/ui";
import { Dumbbell, LogOut } from "lucide-react";
export function PortalLayout({ coach = false }: { coach?: boolean }) {
  const { session, logout } = useAuth();
  const [open, setOpen] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const links = coach
    ? [
        ["/pt/schedule", "Lịch làm việc"],
        ["/pt/requests", "Yêu cầu"],
        ["/pt/reviews", "Đánh giá"],
        ["/pt/performance", "Hiệu suất"],
        ["/pt/profile", "Chuyên môn"],
        ["/pt/account", "Tài khoản"],
      ]
    : [
        ["/member/rooms", "Khám phá phòng"],
        ["/member/bookings", "Lịch của tôi"],
        ["/member/trainers", "Huấn luyện viên"],
        ["/member/profile", "Tài khoản"],
      ];
  return (
    <div className={`portal-shell ${coach ? "coach-portal" : "member-portal"}`}>
      <a className="skip-link" href="#portal-main">
        Đến nội dung chính
      </a>
      <header className="portal-header">
        <NavLink className="portal-brand" to={links[0][0]}>
          <Dumbbell size={24} /> GYM<span>{coach ? "COACH" : "CLUB"}</span>
        </NavLink>
        <nav aria-label="Điều hướng">
          {links.map(([to, label]) => (
            <NavLink key={to} to={to}>
              {label}
            </NavLink>
          ))}
        </nav>
        <div className="portal-user">
          <span>{session.user.full_name}</span>
          <Button
            variant="ghost"
            aria-label="Đăng xuất"
            onClick={() => setOpen(true)}
          >
            <LogOut size={18} />
          </Button>
        </div>
      </header>
      <main id="portal-main" className="portal-main">
        <Outlet />
      </main>
      <Modal
        title="Đăng xuất?"
        description="Thao tác này kết thúc tất cả phiên của tài khoản."
        open={open}
        onClose={() => setOpen(false)}
        busy={busy}
      >
        <Notice>{error}</Notice>
        <Button
          busy={busy}
          onClick={async () => {
            setBusy(true);
            try {
              await logout();
            } catch (e) {
              setError((e as Error).message);
            } finally {
              setBusy(false);
            }
          }}
        >
          Đăng xuất
        </Button>
      </Modal>
    </div>
  );
}
