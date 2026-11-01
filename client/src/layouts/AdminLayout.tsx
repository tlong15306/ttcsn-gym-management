import { useState } from "react";
import {
  NavLink,
  Outlet,
  useLocation,
  ScrollRestoration,
} from "react-router-dom";
import * as Dialog from "@radix-ui/react-dialog";
import {
  Dumbbell,
  UsersRound,
  UserRound,
  Menu,
  X,
  LogOut,
  ChevronRight,
  Building2,
  Clock3,
  CalendarDays,
  BadgeCheck,
  BarChart3,
} from "lucide-react";
import { useAuth, roleLabels } from "../app/session";
import { Button, Modal, Notice } from "../components/ui";

export function AdminLayout() {
  const { session, logout } = useAuth();
  const [drawer, setDrawer] = useState(false);
  const [confirm, setConfirm] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const location = useLocation();
  const links = [
    { to: "/management/attendance", label: "Điểm danh", icon: CalendarDays },
    { to: "/management/bookings", label: "Đặt lịch", icon: CalendarDays },
    ...(session.user.role === "ADMIN"
      ? [{ to: "/admin/users", label: "Tài khoản", icon: UsersRound }]
      : []),
    ...(["ADMIN", "MANAGER", "RECEPTIONIST"].includes(session.user.role)
      ? [{ to: "/management/members", label: "Hội viên", icon: UsersRound }]
      : []),
    ...(["ADMIN", "MANAGER"].includes(session.user.role)
      ? [
          { to: "/management/reports", label: "Báo cáo", icon: BarChart3 },
          { to: "/management/rooms", label: "Phòng tập", icon: Building2 },
          { to: "/management/time-slots", label: "Ca tập", icon: Clock3 },
          {
            to: "/management/room-sessions",
            label: "Phiên phòng",
            icon: CalendarDays,
          },
          {
            to: "/management/trainers",
            label: "Huấn luyện viên",
            icon: BadgeCheck,
          },
        ]
      : []),
    ...(session.user.role === "PT"
      ? [{ to: "/pt/profile", label: "Hồ sơ chuyên môn", icon: BadgeCheck }]
      : []),
    ...(!["ADMIN", "MANAGER"].includes(session.user.role)
      ? [{ to: "/rooms", label: "Lịch phòng tập", icon: CalendarDays }]
      : []),
    { to: "/me", label: "Hồ sơ của tôi", icon: UserRound },
  ];
  const current =
    links.find((l) => location.pathname.startsWith(l.to))?.label ?? "Quản lý";
  const navigation = (
    <>
      <div className="brand">
        <span className="brand-symbol">
          <Dumbbell size={24} />
        </span>
        <span>
          GYM<span className="brand-caption">MANAGEMENT</span>
        </span>
      </div>
      <div className="workspace-label">KHÔNG GIAN LÀM VIỆC</div>
      <nav aria-label="Điều hướng chính">
        {links.map((l) => (
          <NavLink
            key={l.to}
            to={l.to}
            onClick={() => setDrawer(false)}
            className={({ isActive }) => `nav-item${isActive ? " active" : ""}`}
          >
            <l.icon size={19} />
            {l.label}
          </NavLink>
        ))}
      </nav>
      <div className="sidebar-footer">
        <span className="sidebar-status-dot" /> Hệ thống quản lý GYM
        <span>TTCSN · Demo local</span>
      </div>
    </>
  );
  return (
    <div className="app-shell">
      <a className="skip-link" href="#main-content">
        Đến nội dung chính
      </a>
      <aside className="sidebar">{navigation}</aside>
      <Dialog.Root open={drawer} onOpenChange={setDrawer}>
        <Dialog.Portal>
          <Dialog.Overlay className="dialog-overlay" />
          <Dialog.Content className="mobile-sidebar">
            <Dialog.Title className="sr-only">Menu điều hướng</Dialog.Title>
            <Dialog.Description className="sr-only">
              Chọn chức năng quản lý hoặc hồ sơ cá nhân.
            </Dialog.Description>
            <Dialog.Close asChild>
              <Button
                className="drawer-close"
                variant="ghost"
                aria-label="Đóng menu"
              >
                <X size={22} />
              </Button>
            </Dialog.Close>
            {navigation}
          </Dialog.Content>
        </Dialog.Portal>
      </Dialog.Root>
      <div className="main-shell">
        <header className="topbar">
          <div className="breadcrumb">
            <Button
              className="menu-button"
              variant="ghost"
              aria-label="Mở menu"
              onClick={() => setDrawer(true)}
            >
              <Menu size={22} />
            </Button>
            <span>Không gian làm việc</span>
            <ChevronRight size={14} />
            <strong>{current}</strong>
          </div>
          <div className="topbar-user">
            <span className="avatar small">
              {session.user.full_name.slice(0, 1)}
            </span>
            <div>
              <strong>{session.user.full_name}</strong>
              <span>{roleLabels[session.user.role]}</span>
            </div>
            <Button
              variant="ghost"
              className="icon-button"
              aria-label="Đăng xuất"
              onClick={() => setConfirm(true)}
            >
              <LogOut size={19} />
            </Button>
          </div>
        </header>
        <main id="main-content" className="main-content">
          <Outlet />
        </main>
        <footer className="app-footer">
          GYM Management <span>Không gian quản lý của bạn</span>
        </footer>
      </div>
      <Modal
        title="Đăng xuất tài khoản?"
        description="Thao tác này thu hồi tất cả phiên đăng nhập của tài khoản."
        open={confirm}
        onClose={() => setConfirm(false)}
        busy={busy}
      >
        <Notice>{error}</Notice>
        <div className="form-actions">
          <Button
            variant="secondary"
            disabled={busy}
            onClick={() => setConfirm(false)}
          >
            Hủy
          </Button>
          <Button
            busy={busy}
            onClick={async () => {
              setBusy(true);
              setError("");
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
        </div>
      </Modal>
      <ScrollRestoration storageKey="gym-scroll-positions" />
    </div>
  );
}
