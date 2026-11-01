import { useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { resource, api } from "../../api";
import type { Role, User } from "../../types";
import { roleLabels, useAuth } from "../../app/session";
import { vietnamTimestampDate } from "../../utils/vietnamTime";
import {
  Button,
  Input,
  Field,
  PasswordInput,
  Select,
  Badge,
  ActionMenu,
  Modal,
  Notice,
} from "../../components/ui";
import {
  PageHeader,
  Pagination,
  ListState,
  SearchBox,
  useList,
} from "../../components/shared/list";
import { DataForm } from "../../components/shared/form";
import {
  DataTable,
  Details,
  RecordState,
  useRecord,
} from "../../components/shared/record";

const root = "/admin/users";
type UserAction = { user: User; kind: "status" | "role" | "password" };
export function UsersPage() {
  const list = useList<User>("users");
  const navigate = useNavigate();
  const [action, setAction] = useState<UserAction | null>(null);
  const [notice, setNotice] = useState("");
  const filtered = !!(
    list.params.get("q") ||
    list.params.get("role") ||
    list.params.get("status")
  );
  return (
    <>
      <PageHeader
        title="Tài khoản hệ thống"
        description="Quản lý danh tính, quyền truy cập và trạng thái đăng nhập."
        action={{ label: "Thêm tài khoản", to: `${root}/new` }}
      />
      <Notice tone="success">{notice}</Notice>
      <section className="panel">
        <div className="filter-bar">
          <SearchBox
            value={list.params.get("q") ?? ""}
            onChange={(v) => list.filter("q", v)}
            placeholder="Tìm tên, tài khoản hoặc email…"
          />
          <Select
            aria-label="Lọc vai trò"
            value={list.params.get("role") ?? ""}
            onChange={(e) => list.filter("role", e.target.value)}
          >
            <option value="">Tất cả vai trò</option>
            {Object.entries(roleLabels).map(([value, label]) => (
              <option key={value} value={value}>
                {label}
              </option>
            ))}
          </Select>
          <Select
            aria-label="Lọc trạng thái"
            value={list.params.get("status") ?? ""}
            onChange={(e) => list.filter("status", e.target.value)}
          >
            <option value="">Tất cả trạng thái</option>
            <option value="ACTIVE">Hoạt động</option>
            <option value="LOCKED">Đã khóa</option>
          </Select>
          {filtered && (
            <Button variant="ghost" onClick={list.clear}>
              Xóa bộ lọc
            </Button>
          )}
        </div>
        <div className="list-caption">
          <span>Danh sách tài khoản</span>
          <span>
            {list.meta.total} kết quả{filtered ? " phù hợp bộ lọc" : ""}
          </span>
        </div>
        <ListState
          {...list}
          empty={!list.data.length}
          filtered={filtered}
          retry={list.refresh}
        >
          <DataTable
            rows={list.data}
            columns={[
              {
                label: "Tài khoản",
                render: (u) => (
                  <div className="person-cell">
                    <span className="avatar">{u.full_name.slice(0, 1)}</span>
                    <div>
                      <Link className="row-title" to={`${root}/${u.id}`}>
                        {u.full_name}
                      </Link>
                      <span className="cell-subtitle">@{u.username}</span>
                    </div>
                  </div>
                ),
              },
              {
                label: "Liên hệ",
                render: (u) => (
                  <div>
                    <span>{u.email}</span>
                    <span className="cell-subtitle">
                      {u.phone || "Chưa có SĐT"}
                    </span>
                  </div>
                ),
              },
              {
                label: "Vai trò",
                render: (u) => (
                  <span className="role-label">{roleLabels[u.role]}</span>
                ),
              },
              {
                label: "Trạng thái",
                render: (u) => <Badge value={u.status} />,
              },
            ]}
            actions={(u) => (
              <ActionMenu
                name={u.username}
                actions={[
                  {
                    label: "Xem hồ sơ",
                    run: () => navigate(`${root}/${u.id}`),
                  },
                  {
                    label: "Chỉnh sửa",
                    run: () => navigate(`${root}/${u.id}/edit`),
                  },
                  {
                    label:
                      u.status === "ACTIVE"
                        ? "Khóa tài khoản"
                        : "Mở khóa tài khoản",
                    danger: u.status === "ACTIVE",
                    run: () => setAction({ user: u, kind: "status" }),
                  },
                  {
                    label: "Đổi vai trò",
                    run: () => setAction({ user: u, kind: "role" }),
                  },
                  {
                    label: "Đặt lại mật khẩu",
                    danger: true,
                    run: () => setAction({ user: u, kind: "password" }),
                  },
                ]}
              />
            )}
          />
        </ListState>
        <Pagination meta={list.meta} change={list.filter} />
      </section>
      {action && (
        <UserActionDialog
          action={action}
          close={() => setAction(null)}
          done={() => {
            setAction(null);
            setNotice("Đã cập nhật tài khoản.");
            list.refresh();
          }}
        />
      )}
    </>
  );
}
export function UserActionDialog({
  action,
  close,
  done,
}: {
  action: UserAction;
  close: () => void;
  done: () => void;
}) {
  const { session, expire } = useAuth();
  const { user, kind } = action;
  const [reason, setReason] = useState("");
  const [role, setRole] = useState<Role>(user.role);
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const title =
    kind === "password"
      ? "Đặt lại mật khẩu"
      : kind === "role"
        ? "Đổi vai trò"
        : user.status === "ACTIVE"
          ? "Khóa tài khoản"
          : "Mở khóa tài khoản";
  async function save() {
    if (busy) return;
    if (kind === "password" && password !== confirmPassword) {
      setError("Mật khẩu xác nhận chưa khớp.");
      return;
    }
    setBusy(true);
    setError("");
    try {
      if (kind === "password")
        await api.resetPassword(session.accessToken, user.id, password);
      else if (kind === "role")
        await api.changeRole(session.accessToken, user.id, role, reason);
      else
        await api.changeStatus(
          session.accessToken,
          user.id,
          user.status === "ACTIVE" ? "LOCKED" : "ACTIVE",
          reason,
        );
      done();
      if (user.id === session.user.id) expire();
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  }
  return (
    <Modal
      title={title}
      description={`${user.full_name} · @${user.username}`}
      open
      onClose={close}
      busy={busy}
    >
      <Notice tone="info">
        Token cũ sẽ mất hiệu lực. Mở khóa yêu cầu đăng nhập lại. Không thể khóa
        hoặc hạ quyền Admin hoạt động cuối cùng.
      </Notice>
      <Notice>{error}</Notice>
      <form
        onSubmit={(e) => {
          e.preventDefault();
          void save();
        }}
      >
        <fieldset disabled={busy}>
          {kind === "role" && (
            <Field
              label="Vai trò mới"
              hint="Đổi sang Hội viên/PT cần hồ sơ nghiệp vụ tương ứng."
            >
              <Select
                value={role}
                onChange={(e) => setRole(e.target.value as Role)}
              >
                {Object.entries(roleLabels).map(([v, l]) => (
                  <option key={v} value={v}>
                    {l}
                  </option>
                ))}
              </Select>
            </Field>
          )}
          {kind === "password" ? (
            <>
              <Field
                label="Mật khẩu mới"
                required
                hint="8 ký tự trở lên, tối đa 72 byte."
              >
                <PasswordInput
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  minLength={8}
                  maxLength={72}
                  required
                  autoComplete="new-password"
                />
              </Field>
              <Field label="Nhập lại mật khẩu" required>
                <PasswordInput
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  required
                  autoComplete="new-password"
                />
              </Field>
            </>
          ) : (
            <Field label="Lý do thay đổi" required>
              <Input
                value={reason}
                onChange={(e) => setReason(e.target.value)}
                required
                maxLength={1000}
                placeholder="Nhập lý do để lưu vào nhật ký"
              />
            </Field>
          )}
        </fieldset>
        <div className="form-actions">
          <Button
            variant="secondary"
            type="button"
            disabled={busy}
            onClick={close}
          >
            Hủy
          </Button>
          <Button variant={kind === "role" ? "primary" : "danger"} busy={busy}>
            {title}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
export function UserFormPage() {
  const { id } = useParams();
  const auth = useAuth();
  const navigate = useNavigate();
  const record = useRecord<User>(id ? `users/${id}` : null);
  if (id && !record.data)
    return <RecordState {...record} retry={record.refresh} />;
  return (
    <UserForm
      key={id ?? "new"}
      user={record.data}
      token={auth.session.accessToken}
      done={() => navigate(id ? `${root}/${id}` : root)}
    />
  );
}
function UserForm({
  user,
  token,
  done,
}: {
  user: User | null;
  token: string;
  done: () => void;
}) {
  const [values, setValues] = useState({
    username: user?.username ?? "",
    full_name: user?.full_name ?? "",
    email: user?.email ?? "",
    phone: user?.phone ?? "",
    password: "",
    role: user?.role ?? ("MEMBER" as Role),
  });
  function change(key: keyof typeof values, value: string) {
    setValues((v) => ({ ...v, [key]: value }));
  }
  return (
    <>
      <PageHeader
        title={user ? "Chỉnh sửa tài khoản" : "Thêm tài khoản"}
        description={
          user
            ? "Chỉnh sửa thông tin liên hệ. Quyền và trạng thái được thay đổi riêng."
            : "Tài khoản mới có trạng thái hoạt động. Hồ sơ hội viên/PT được tạo ở module tương ứng."
        }
        back={root}
      />
      <section className="panel form-panel">
        <div className="section-heading">
          <h2>Thông tin tài khoản</h2>
          <p>Các trường có dấu * là bắt buộc.</p>
        </div>
        <DataForm
          back={root}
          submitLabel={user ? "Lưu thay đổi" : "Tạo tài khoản"}
          onSuccess={done}
          save={() =>
            user
              ? resource.save(
                  token,
                  `users/${user.id}`,
                  {
                    full_name: values.full_name,
                    email: values.email,
                    phone: values.phone,
                  },
                  "PATCH",
                )
              : api.createUser(token, values)
          }
        >
          <div className="form-grid">
            <Field label="Họ và tên" required>
              <Input
                value={values.full_name}
                onChange={(e) => change("full_name", e.target.value)}
                required
                maxLength={150}
                autoComplete="name"
              />
            </Field>
            <Field
              label="Tên đăng nhập"
              required
              hint={
                user ? "Tên đăng nhập được giữ nguyên." : "Từ 3 đến 50 ký tự."
              }
            >
              <Input
                value={values.username}
                onChange={(e) => change("username", e.target.value)}
                readOnly={!!user}
                required
                minLength={3}
                maxLength={50}
                autoComplete="off"
              />
            </Field>
            <Field label="Email" required>
              <Input
                value={values.email}
                onChange={(e) => change("email", e.target.value)}
                type="email"
                required
                maxLength={254}
                autoComplete="email"
              />
            </Field>
            <Field label="Số điện thoại">
              <Input
                value={values.phone}
                onChange={(e) => change("phone", e.target.value)}
                type="tel"
                maxLength={20}
                autoComplete="tel"
              />
            </Field>
            {!user && (
              <>
                <Field label="Vai trò" required>
                  <Select
                    value={values.role}
                    onChange={(e) => change("role", e.target.value)}
                  >
                    {Object.entries(roleLabels).map(([v, l]) => (
                      <option key={v} value={v}>
                        {l}
                      </option>
                    ))}
                  </Select>
                </Field>
                <Field
                  label="Mật khẩu ban đầu"
                  required
                  hint="Ít nhất 8 ký tự, tối đa 72 byte."
                >
                  <PasswordInput
                    value={values.password}
                    onChange={(e) => change("password", e.target.value)}
                    required
                    minLength={8}
                    maxLength={72}
                    autoComplete="new-password"
                  />
                </Field>
              </>
            )}
          </div>
        </DataForm>
      </section>
    </>
  );
}
export function UserDetailPage() {
  const { id } = useParams();
  const record = useRecord<User>(`users/${id}`);
  const [action, setAction] = useState<UserAction | null>(null);
  const user = record.data;
  if (!user) return <RecordState {...record} retry={record.refresh} />;
  return (
    <>
      <PageHeader
        title={user.full_name}
        description={`Tài khoản @${user.username}`}
        back={root}
        action={{ label: "Chỉnh sửa", to: `${root}/${id}/edit` }}
      />
      <section className="panel detail-panel">
        <Details
          items={[
            { label: "Vai trò", value: roleLabels[user.role] },
            { label: "Trạng thái", value: <Badge value={user.status} /> },
            { label: "Email", value: user.email },
            { label: "Số điện thoại", value: user.phone },
            {
              label: "Ngày tạo",
              value: vietnamTimestampDate(user.created_at),
            },
          ]}
        />
        <div className="form-actions">
          <Button
            variant="secondary"
            onClick={() => setAction({ user, kind: "role" })}
          >
            Đổi vai trò
          </Button>
          <Button
            variant="secondary"
            onClick={() => setAction({ user, kind: "password" })}
          >
            Đặt lại mật khẩu
          </Button>
          <Button
            variant="danger"
            onClick={() => setAction({ user, kind: "status" })}
          >
            {user.status === "ACTIVE" ? "Khóa tài khoản" : "Mở khóa tài khoản"}
          </Button>
        </div>
      </section>
      {action && (
        <UserActionDialog
          action={action}
          close={() => setAction(null)}
          done={() => {
            setAction(null);
            record.refresh();
          }}
        />
      )}
    </>
  );
}
