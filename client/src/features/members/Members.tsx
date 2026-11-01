import { useState, useEffect } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import type { User } from "../../types";
import { resource } from "../../api";
import { useAuth } from "../../app/session";
import { civilDate, localDate } from "../../utils/vietnamTime";
import {
  Field,
  Input,
  Select,
  Badge,
  ActionMenu,
  Button,
  Modal,
  PasswordInput,
  Notice,
} from "../../components/ui";
import { DataForm } from "../../components/shared/form";
import {
  DataTable,
  Details,
  RecordState,
  useRecord,
} from "../../components/shared/record";
import {
  PageHeader,
  ListState,
  Pagination,
  SearchBox,
  useList,
} from "../../components/shared/list";

export type Member = {
  id: number;
  user_id: number;
  user: User;
  member_code: string;
  date_of_birth: string;
  gender: string;
  address: string;
  join_date: string;
  membership_expiry: string;
  status: string;
  expired: boolean;
};
const root = "/management/members";
const dateLabel = (s: string) =>
  s ? civilDate(s) : "—";
export function MembersPage() {
  const list = useList<Member>("members");
  const navigate = useNavigate();
  const [target, setTarget] = useState<Member | null>(null);
  const [notice, setNotice] = useState("");
  const filtered = !!(list.params.get("q") || list.params.get("status"));
  return (
    <>
      <PageHeader
        title="Hội viên"
        description="Theo dõi hồ sơ, hạn thẻ và trạng thái hội viên."
        action={{ label: "Thêm hội viên", to: `${root}/new` }}
      />
      <Notice tone="success">{notice}</Notice>
      <section className="panel">
        <div className="filter-bar">
          <SearchBox
            value={list.params.get("q") ?? ""}
            onChange={(v) => list.filter("q", v)}
            placeholder="Tìm mã, họ tên, email hoặc SĐT…"
          />
          <Select
            aria-label="Lọc trạng thái hội viên"
            value={list.params.get("status") ?? ""}
            onChange={(e) => list.filter("status", e.target.value)}
          >
            <option value="">Tất cả trạng thái</option>
            <option value="ACTIVE">Hoạt động</option>
            <option value="INACTIVE">Ngừng hoạt động</option>
          </Select>
          {filtered && (
            <Button variant="ghost" onClick={list.clear}>
              Xóa bộ lọc
            </Button>
          )}
        </div>
        <div className="list-caption">
          <span>Danh sách hội viên</span>
          <span>{list.meta.total} kết quả</span>
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
                label: "Hội viên",
                render: (m) => (
                  <div>
                    <Link className="row-title" to={`${root}/${m.id}`}>
                      {m.user.full_name}
                    </Link>
                    <span className="cell-subtitle">{m.member_code}</span>
                  </div>
                ),
              },
              {
                label: "Liên hệ",
                render: (m) => (
                  <div>
                    {m.user.email}
                    <span className="cell-subtitle">
                      {m.user.phone || "Chưa có SĐT"}
                    </span>
                  </div>
                ),
              },
              {
                label: "Hạn thẻ",
                render: (m) => (
                  <div>
                    {dateLabel(m.membership_expiry)}
                    <span
                      className={
                        m.expired
                          ? "cell-subtitle text-danger"
                          : "cell-subtitle"
                      }
                    >
                      {m.expired ? "Đã hết hạn" : "Còn thời hạn"}
                    </span>
                  </div>
                ),
              },
              {
                label: "Hội viên / tài khoản",
                render: (m) => (
                  <div className="badge-group">
                    <Badge value={m.status} />
                    <Badge value={m.user.status} />
                  </div>
                ),
              },
            ]}
            actions={(m) => (
              <ActionMenu
                name={m.member_code}
                actions={[
                  {
                    label: "Xem hồ sơ",
                    run: () => navigate(`${root}/${m.id}`),
                  },
                  {
                    label: "Chỉnh sửa",
                    run: () => navigate(`${root}/${m.id}/edit`),
                  },
                  {
                    label: "Cập nhật thẻ / trạng thái",
                    run: () => setTarget(m),
                  },
                ]}
              />
            )}
          />
        </ListState>
        <Pagination meta={list.meta} change={list.filter} />
      </section>
      {target && (
        <MembershipDialog
          member={target}
          close={() => setTarget(null)}
          done={() => {
            setTarget(null);
            setNotice("Đã cập nhật thẻ hội viên.");
            list.refresh();
          }}
        />
      )}
    </>
  );
}
function MembershipDialog({
  member,
  close,
  done,
}: {
  member: Member;
  close: () => void;
  done: () => void;
}) {
  const { session } = useAuth();
  const [expiry, setExpiry] = useState(member.membership_expiry);
  const [status, setStatus] = useState(member.status);
  const [reason, setReason] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  return (
    <Modal
      title="Cập nhật thẻ hội viên"
      description={`${member.user.full_name} · ${member.member_code}`}
      open
      onClose={close}
      busy={busy}
    >
      <Notice tone="info">
        Trạng thái hội viên độc lập với trạng thái đăng nhập của tài khoản.
      </Notice>
      <Notice>{error}</Notice>
      <form
        onSubmit={async (e) => {
          e.preventDefault();
          if (busy) return;
          setBusy(true);
          setError("");
          try {
            await resource.save(
              session.accessToken,
              `members/${member.id}/membership`,
              { membership_expiry: expiry, status, reason },
              "PATCH",
            );
            done();
          } catch (e) {
            setError((e as Error).message);
          } finally {
            setBusy(false);
          }
        }}
      >
        <fieldset disabled={busy}>
          <Field label="Hạn thẻ mới" required>
            <Input
              type="date"
              min={member.join_date}
              required
              value={expiry}
              onChange={(e) => setExpiry(e.target.value)}
            />
          </Field>
          <Field label="Trạng thái hội viên">
            <Select value={status} onChange={(e) => setStatus(e.target.value)}>
              <option value="ACTIVE">Hoạt động</option>
              <option value="INACTIVE">Ngừng hoạt động</option>
            </Select>
          </Field>
          <Field label="Lý do thay đổi" required>
            <Input
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              required
            />
          </Field>
        </fieldset>
        <div className="form-actions">
          <Button
            type="button"
            variant="secondary"
            disabled={busy}
            onClick={close}
          >
            Hủy
          </Button>
          <Button busy={busy}>Lưu thay đổi</Button>
        </div>
      </form>
    </Modal>
  );
}
export function MemberFormPage() {
  const { id } = useParams();
  const record = useRecord<Member>(id ? `members/${id}` : null);
  if (id && !record.data)
    return <RecordState {...record} retry={record.refresh} />;
  return <MemberForm key={id ?? "new"} member={record.data} />;
}
function MemberForm({ member }: { member: Member | null }) {
  const { session } = useAuth();
  const navigate = useNavigate();
  const [mode, setMode] = useState("new");
  const [query, setQuery] = useState("");
  const [eligible, setEligible] = useState<User[]>([]);
  const [lookupError, setLookupError] = useState("");
  const [userID, setUserID] = useState("");
  const [v, setV] = useState({
    member_code: member?.member_code ?? "",
    date_of_birth: member?.date_of_birth ?? "",
    gender: member?.gender ?? "",
    address: member?.address ?? "",
    join_date: member?.join_date ?? localDate(),
    membership_expiry: member?.membership_expiry ?? "",
    full_name: member?.user.full_name ?? "",
    email: member?.user.email ?? "",
    phone: member?.user.phone ?? "",
    username: "",
    password: "",
  });
  const change = (key: keyof typeof v, value: string) =>
    setV((p) => ({ ...p, [key]: value }));
  useEffect(() => {
    if (member || mode !== "existing") return;
    const abort = new AbortController();
    const timer = setTimeout(() => {
      setLookupError("");
      resource
        .list<User>(
          session.accessToken,
          "members/eligible-accounts",
          new URLSearchParams({ q: query }),
          abort.signal,
        )
        .then((r) => setEligible(r.data))
        .catch((e) => {
          if (!abort.signal.aborted) setLookupError(e.message);
        });
    }, 300);
    return () => {
      clearTimeout(timer);
      abort.abort();
    };
  }, [mode, query, session.accessToken, member]);
  async function save() {
    if (member)
      return resource.save(
        session.accessToken,
        `members/${member.id}`,
        {
          full_name: v.full_name,
          email: v.email,
          phone: v.phone,
          date_of_birth: v.date_of_birth,
          gender: v.gender,
          address: v.address,
        },
        "PATCH",
      );
    const common = {
      member_code: v.member_code,
      date_of_birth: v.date_of_birth,
      gender: v.gender,
      address: v.address,
      join_date: v.join_date,
      membership_expiry: v.membership_expiry,
    };
    return resource.save(
      session.accessToken,
      "members",
      mode === "existing"
        ? { ...common, user_id: Number(userID) }
        : {
            ...common,
            account: {
              username: v.username,
              password: v.password,
              full_name: v.full_name,
              email: v.email,
              phone: v.phone,
            },
          },
    );
  }
  return (
    <>
      <PageHeader
        title={member ? "Chỉnh sửa hội viên" : "Thêm hội viên"}
        description="Thông tin liên hệ được đồng bộ với tài khoản. Ngày thẻ được cập nhật bằng thao tác riêng."
        back={root}
      />
      <section className="panel form-panel">
        <DataForm
          save={save}
          back={root}
          submitLabel={member ? "Lưu thay đổi" : "Tạo hội viên"}
          onSuccess={() => navigate(member ? `${root}/${member.id}` : root)}
        >
          {!member && (
            <>
              <div className="section-heading">
                <h2>Liên kết tài khoản</h2>
                <p>Tài khoản mới được gán vai trò Hội viên.</p>
              </div>
              <Field label="Cách tạo">
                <Select value={mode} onChange={(e) => setMode(e.target.value)}>
                  <option value="new">Tạo tài khoản MEMBER mới</option>
                  <option value="existing">
                    Liên kết tài khoản MEMBER đã có
                  </option>
                </Select>
              </Field>
              {mode === "existing" && (
                <>
                  <Field label="Tìm tài khoản">
                    <Input
                      value={query}
                      onChange={(e) => setQuery(e.target.value)}
                      placeholder="Tên đăng nhập, họ tên hoặc email…"
                    />
                  </Field>
                  <Notice>{lookupError}</Notice>
                  <Field
                    label="Tài khoản MEMBER chưa có hồ sơ"
                    required
                    hint="Chỉ hiển thị tài khoản phù hợp, tối đa 20 kết quả. Nhập từ khóa để thu hẹp."
                  >
                    <Select
                      required
                      value={userID}
                      onChange={(e) => setUserID(e.target.value)}
                    >
                      <option value="">Chọn tài khoản</option>
                      {eligible.map((u) => (
                        <option key={u.id} value={u.id}>
                          {u.full_name} · @{u.username}
                        </option>
                      ))}
                    </Select>
                  </Field>
                </>
              )}
            </>
          )}
          {(member || mode === "new") && (
            <>
              <div className="form-grid">
                <Field label="Họ và tên" required>
                  <Input
                    value={v.full_name}
                    onChange={(e) => change("full_name", e.target.value)}
                    required
                    maxLength={150}
                  />
                </Field>
                <Field label="Email" required>
                  <Input
                    value={v.email}
                    onChange={(e) => change("email", e.target.value)}
                    type="email"
                    required
                    maxLength={254}
                  />
                </Field>
                <Field label="Số điện thoại">
                  <Input
                    value={v.phone}
                    onChange={(e) => change("phone", e.target.value)}
                    type="tel"
                    maxLength={20}
                  />
                </Field>
                {!member && (
                  <>
                    <Field label="Tên đăng nhập" required>
                      <Input
                        value={v.username}
                        onChange={(e) => change("username", e.target.value)}
                        required
                        minLength={3}
                        maxLength={50}
                      />
                    </Field>
                    <Field
                      label="Mật khẩu ban đầu"
                      required
                      hint="Từ 8 ký tự, tối đa 72 byte."
                    >
                      <PasswordInput
                        value={v.password}
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
            </>
          )}
          <div className="section-heading">
            <h2>Hồ sơ hội viên</h2>
          </div>
          <div className="form-grid">
            <Field label="Mã hội viên" required>
              <Input
                value={v.member_code}
                onChange={(e) => change("member_code", e.target.value)}
                readOnly={!!member}
                required
                maxLength={30}
              />
            </Field>
            <Field label="Ngày sinh">
              <Input
                type="date"
                value={v.date_of_birth}
                onChange={(e) => change("date_of_birth", e.target.value)}
              />
            </Field>
            <Field label="Giới tính">
              <Select
                value={v.gender}
                onChange={(e) => change("gender", e.target.value)}
              >
                <option value="">Chưa cung cấp</option>
                <option value="MALE">Nam</option>
                <option value="FEMALE">Nữ</option>
                <option value="OTHER">Khác</option>
              </Select>
            </Field>
            {!member && (
              <>
                <Field label="Ngày tham gia" required>
                  <Input
                    type="date"
                    value={v.join_date}
                    onChange={(e) => change("join_date", e.target.value)}
                    required
                  />
                </Field>
                <Field label="Hạn thẻ" required>
                  <Input
                    type="date"
                    value={v.membership_expiry}
                    min={v.join_date}
                    onChange={(e) =>
                      change("membership_expiry", e.target.value)
                    }
                    required
                  />
                </Field>
              </>
            )}
          </div>
          <Field label="Địa chỉ">
            <textarea
              value={v.address}
              onChange={(e) => change("address", e.target.value)}
              maxLength={2000}
            />
          </Field>
        </DataForm>
      </section>
    </>
  );
}
export function MemberDetailPage() {
  const { id } = useParams();
  const record = useRecord<Member>(`members/${id}`);
  const [target, setTarget] = useState(false);
  const m = record.data;
  if (!m) return <RecordState {...record} retry={record.refresh} />;
  return (
    <>
      <PageHeader
        title={m.user.full_name}
        description={`Mã hội viên ${m.member_code} · @${m.user.username}`}
        back={root}
        action={{ label: "Chỉnh sửa", to: `${root}/${id}/edit` }}
      />
      <section className="panel detail-panel">
        <Details
          items={[
            { label: "Email", value: m.user.email },
            { label: "Số điện thoại", value: m.user.phone },
            { label: "Ngày sinh", value: dateLabel(m.date_of_birth) },
            {
              label: "Giới tính",
              value: (
                { MALE: "Nam", FEMALE: "Nữ", OTHER: "Khác" } as Record<
                  string,
                  string
                >
              )[m.gender],
            },
            { label: "Địa chỉ", value: m.address },
            { label: "Ngày tham gia", value: dateLabel(m.join_date) },
            {
              label: "Hạn thẻ",
              value: `${dateLabel(m.membership_expiry)} · ${m.expired ? "Đã hết hạn" : "Còn thời hạn"}`,
            },
            { label: "Trạng thái hội viên", value: <Badge value={m.status} /> },
            {
              label: "Trạng thái đăng nhập",
              value: <Badge value={m.user.status} />,
            },
          ]}
        />
        <div className="form-actions">
          <Button onClick={() => setTarget(true)}>
            Cập nhật thẻ / trạng thái
          </Button>
        </div>
      </section>
      {target && (
        <MembershipDialog
          member={m}
          close={() => setTarget(false)}
          done={() => {
            setTarget(false);
            record.refresh();
          }}
        />
      )}
    </>
  );
}
