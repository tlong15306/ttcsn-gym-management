import { useState } from "react";
import { api } from "../../api";
import { useAuth, roleLabels } from "../../app/session";
import { PageHeader } from "../../components/shared/list";
import { DataForm } from "../../components/shared/form";
import { Field, Input, Badge } from "../../components/ui";

export function ProfilePage() {
  const { session, updateUser } = useAuth();
  const user = session.user;
  const [fullName, setName] = useState(user.full_name);
  const [email, setEmail] = useState(user.email);
  const [phone, setPhone] = useState(user.phone ?? "");
  return (
    <>
      <PageHeader
        title="Hồ sơ của tôi"
        description="Cập nhật thông tin liên hệ của tài khoản đang đăng nhập."
      />
      <div className="profile-layout">
        <aside className="panel profile-summary">
          <span className="profile-avatar">{user.full_name.slice(0, 1)}</span>
          <h2>{user.full_name}</h2>
          <p>@{user.username}</p>
          <span className="role-label">{roleLabels[user.role]}</span>
          <Badge value={user.status} />
        </aside>
        <section className="panel form-panel">
          <div className="section-heading">
            <h2>Thông tin cá nhân</h2>
            <p>Vai trò và trạng thái do quản trị viên quản lý.</p>
          </div>
          <DataForm
            save={async () => {
              const r = await api.updateMe(session.accessToken, {
                full_name: fullName,
                email,
                phone,
              });
              updateUser(r.data);
            }}
          >
            <Field label="Họ và tên" required>
              <Input
                value={fullName}
                onChange={(e) => setName(e.target.value)}
                maxLength={150}
                required
                autoComplete="name"
              />
            </Field>
            <div className="form-grid">
              <Field label="Email" required>
                <Input
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  type="email"
                  maxLength={254}
                  required
                  autoComplete="email"
                />
              </Field>
              <Field label="Số điện thoại">
                <Input
                  value={phone}
                  onChange={(e) => setPhone(e.target.value)}
                  maxLength={20}
                  type="tel"
                  autoComplete="tel"
                />
              </Field>
            </div>
          </DataForm>
        </section>
      </div>
    </>
  );
}
