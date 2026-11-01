import { useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import type { User } from "../../types";
import { ApiError, request, resource } from "../../api";
import { useAuth } from "../../app/session";
import { vietnamDateTime } from "../../utils/vietnamTime";
import {
  Button,
  Field,
  Input,
  Select,
  Badge,
  ActionMenu,
  Notice,
} from "../../components/ui";
import { DataForm } from "../../components/shared/form";
import { OperationDialog } from "../../components/shared/operation";
import {
  PageHeader,
  Pagination,
  ListState,
  SearchBox,
  useList,
} from "../../components/shared/list";
import {
  DataTable,
  Details,
  RecordState,
  useRecord,
} from "../../components/shared/record";
import {
  AvailabilityGrid,
  type AvailabilityHour,
} from "./AvailabilityGrid";

type Profile = {
  id: number;
  user_id: number;
  user: User;
  specialty: string;
  experience_text: string;
  has_certificate: boolean;
  approval_status: string;
  decision_reason: string;
  decided_at: string | null;
};
type WorkingHour = AvailabilityHour & {
  id?: number;
};
type Absence = {
  id: number;
  starts_at: string;
  ends_at: string;
  reason: string;
};
const root = "/management/trainers";
const dayNames = [
  "",
  "Thứ hai",
  "Thứ ba",
  "Thứ tư",
  "Thứ năm",
  "Thứ sáu",
  "Thứ bảy",
  "Chủ nhật",
];
export function TrainersPage() {
  const list = useList<Profile>("pts");
  const navigate = useNavigate();
  const [target, setTarget] = useState<{
    profile: Profile;
    action: string;
  } | null>(null);
  const [notice, setNotice] = useState("");
  const filtered = !!(list.params.get("q") || list.params.get("status"));
  return (
    <>
      <PageHeader
        title="Huấn luyện viên"
        description="Xem hồ sơ chuyên môn, duyệt PT và điều phối lịch làm việc."
      />
      <Notice tone="info">
        PT tự tạo và nộp hồ sơ từ tài khoản của mình. Tài khoản PT mới được
        Admin tạo ở mục Tài khoản.
      </Notice>
      <Notice tone="success">{notice}</Notice>
      <section className="panel">
        <div className="filter-bar">
          <SearchBox
            value={list.params.get("q") ?? ""}
            onChange={(v) => list.filter("q", v)}
            placeholder="Tìm tên hoặc chuyên môn…"
          />
          <Select
            aria-label="Lọc trạng thái hồ sơ"
            value={list.params.get("status") ?? ""}
            onChange={(e) => list.filter("status", e.target.value)}
          >
            <option value="">Tất cả trạng thái</option>
            <option value="DRAFT">Bản nháp</option>
            <option value="PENDING">Chờ duyệt</option>
            <option value="APPROVED">Đã duyệt</option>
            <option value="REJECTED">Từ chối</option>
            <option value="SUSPENDED">Đình chỉ</option>
          </Select>
          {filtered && (
            <Button variant="ghost" onClick={list.clear}>
              Xóa bộ lọc
            </Button>
          )}
        </div>
        <div className="list-caption">
          <span>Hồ sơ PT</span>
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
                label: "Huấn luyện viên",
                render: (p) => (
                  <div>
                    <Link className="row-title" to={`${root}/${p.id}`}>
                      {p.user.full_name}
                    </Link>
                    <span className="cell-subtitle">@{p.user.username}</span>
                  </div>
                ),
              },
              {
                label: "Chuyên môn",
                render: (p) => p.specialty || "Chưa cung cấp",
              },
              {
                label: "Chứng chỉ",
                render: (p) => (p.has_certificate ? "Đã tải lên" : "Chưa có"),
              },
              {
                label: "Hồ sơ / tài khoản",
                render: (p) => (
                  <div className="badge-group">
                    <Badge value={p.approval_status} />
                    <Badge value={p.user.status} />
                  </div>
                ),
              },
            ]}
            actions={(p) => (
              <ActionMenu
                name={p.user.full_name}
                actions={[
                  {
                    label: "Xem hồ sơ / lịch",
                    run: () => navigate(`${root}/${p.id}`),
                  },
                  ...(p.approval_status === "PENDING"
                    ? [
                        {
                          label: "Duyệt hồ sơ",
                          run: () =>
                            setTarget({ profile: p, action: "APPROVE" }),
                        },
                        {
                          label: "Từ chối hồ sơ",
                          danger: true,
                          run: () =>
                            setTarget({ profile: p, action: "REJECT" }),
                        },
                      ]
                    : []),
                  ...(p.approval_status === "APPROVED"
                    ? [
                        {
                          label: "Đình chỉ PT",
                          danger: true,
                          run: () =>
                            setTarget({ profile: p, action: "SUSPEND" }),
                        },
                      ]
                    : []),
                ]}
              />
            )}
          />
        </ListState>
        <Pagination meta={list.meta} change={list.filter} />
      </section>
      {target && (
        <DecisionDialog
          target={target}
          close={() => setTarget(null)}
          done={() => {
            setTarget(null);
            setNotice("Đã lưu quyết định hồ sơ PT.");
            list.refresh();
          }}
        />
      )}
    </>
  );
}
function DecisionDialog({
  target,
  close,
  done,
}: {
  target: { profile: Profile; action: string };
  close: () => void;
  done: () => void;
}) {
  const { session } = useAuth();
  const [reason, setReason] = useState("");
  const title =
    target.action === "APPROVE"
      ? "Duyệt hồ sơ"
      : target.action === "REJECT"
        ? "Từ chối hồ sơ"
        : "Đình chỉ PT";
  return (
    <OperationDialog
      title={title}
      description={target.profile.user.full_name}
      warning="Quyết định và lý do được lưu vào nhật ký. Đình chỉ bị chặn khi còn booking đang giữ chỗ."
      danger={target.action !== "APPROVE"}
      close={close}
      done={done}
      save={() =>
        resource.save(
          session.accessToken,
          `pts/${target.profile.id}/decision`,
          { action: target.action, reason },
        )
      }
    >
      <Field label="Lý do quyết định" required>
        <Input
          value={reason}
          onChange={(e) => setReason(e.target.value)}
          required
        />
      </Field>
    </OperationDialog>
  );
}
function CertificateButton({ profile }: { profile: Profile }) {
  const { session } = useAuth();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  return (
    <>
      <Notice>{error}</Notice>
      <Button
        variant="secondary"
        busy={busy}
        disabled={!profile.has_certificate}
        onClick={async () => {
          if (busy) return;
          setBusy(true);
          setError("");
          try {
            const r = await fetch(`/api/v1/pts/${profile.id}/certificate`, {
              headers: { Authorization: `Bearer ${session.accessToken}` },
            });
            if (!r.ok) {
              if (r.status === 401)
                window.dispatchEvent(new Event("session-expired"));
              const b = await r.json();
              throw new ApiError(
                r.status,
                b.error?.code,
                b.error?.message ?? "Không thể tải chứng chỉ.",
              );
            }
            const blob = await r.blob();
            const url = URL.createObjectURL(blob);
            const link = document.createElement("a");
            link.href = url;
            const extension = blob.type.includes("pdf")
              ? "pdf"
              : blob.type.includes("png")
                ? "png"
                : "jpg";
            link.download = `chung-chi-pt-${profile.id}.${extension}`;
            link.click();
            setTimeout(() => URL.revokeObjectURL(url), 1000);
          } catch (e) {
            setError((e as Error).message);
          } finally {
            setBusy(false);
          }
        }}
      >
        Tải chứng chỉ
      </Button>
    </>
  );
}
export function TrainerDetailPage() {
  const { id } = useParams();
  const record = useRecord<Profile>(`pts/${id}`);
  const [action, setAction] = useState<string | null>(null);
  const p = record.data;
  if (!p) return <RecordState {...record} retry={record.refresh} />;
  return (
    <>
      <PageHeader
        title={p.user.full_name}
        description={`Hồ sơ PT · @${p.user.username}`}
        back={root}
      />
      <section className="panel detail-panel">
        <Details
          items={[
            {
              label: "Trạng thái hồ sơ",
              value: <Badge value={p.approval_status} />,
            },
            { label: "Tài khoản", value: <Badge value={p.user.status} /> },
            { label: "Chuyên môn", value: p.specialty },
            { label: "Kinh nghiệm", value: p.experience_text },
            { label: "Lý do quyết định", value: p.decision_reason },
            {
              label: "Ngày quyết định",
              value: p.decided_at
                ? vietnamDateTime(p.decided_at)
                : "Chưa có",
            },
          ]}
        />
        <div className="form-actions">
          <CertificateButton profile={p} />
          {p.approval_status === "PENDING" && (
            <>
              <Button variant="danger" onClick={() => setAction("REJECT")}>
                Từ chối hồ sơ
              </Button>
              <Button onClick={() => setAction("APPROVE")}>Duyệt hồ sơ</Button>
            </>
          )}
          {p.approval_status === "APPROVED" && (
            <Button variant="danger" onClick={() => setAction("SUSPEND")}>
              Đình chỉ PT
            </Button>
          )}
        </div>
      </section>
      <SchedulePanel prefix={`pts/${id}`} />
      {action && (
        <DecisionDialog
          target={{ profile: p, action }}
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
export function OwnTrainerPage() {
  const record = useRecord<Profile>("pts/me");
  if (record.loading || (record.error && !record.missing))
    return <RecordState {...record} retry={record.refresh} />;
  return (
    <OwnProfileEditor
      key={record.data?.id ?? "new"}
      profile={record.data}
      update={record.setData}
    />
  );
}
function OwnProfileEditor({
  profile,
  update,
}: {
  profile: Profile | null;
  update: (p: Profile) => void;
}) {
  const { session } = useAuth();
  const [specialty, setSpecialty] = useState(profile?.specialty ?? "");
  const [experience, setExperience] = useState(profile?.experience_text ?? "");
  const [file, setFile] = useState<File | null>(null);
  const [uploadBusy, setUploadBusy] = useState(false);
  const [error, setError] = useState("");
  const [submit, setSubmit] = useState(false);
  const editable =
    !profile || ["DRAFT", "REJECTED"].includes(profile.approval_status);
  return (
    <>
      <PageHeader
        title="Hồ sơ chuyên môn của tôi"
        description="Lưu chuyên môn, tải chứng chỉ và gửi hồ sơ để quản lý xét duyệt."
      />
      <section className="panel form-panel">
        <div className="section-heading">
          <h2>Thông tin chuyên môn</h2>
          {profile && (
            <div className="badge-group">
              <Badge value={profile.approval_status} />
            </div>
          )}
        </div>
        {profile?.decision_reason && (
          <Notice tone="info">
            Lý do quyết định: {profile.decision_reason}
          </Notice>
        )}
        {editable ? (
          <DataForm
            save={async () => {
              const r = await resource.save<Profile>(
                session.accessToken,
                "pts/me/profile",
                { specialty, experience_text: experience },
                "PUT",
              );
              update(r.data);
            }}
          >
            <Field label="Chuyên môn" required>
              <Input
                value={specialty}
                onChange={(e) => setSpecialty(e.target.value)}
                required
                maxLength={200}
                placeholder="Ví dụ: Strength training, Yoga…"
              />
            </Field>
            <Field label="Kinh nghiệm" required>
              <textarea
                value={experience}
                onChange={(e) => setExperience(e.target.value)}
                required
                maxLength={10000}
                placeholder="Mô tả kinh nghiệm và quá trình đào tạo"
              />
            </Field>
          </DataForm>
        ) : (
          <Details
            items={[
              { label: "Chuyên môn", value: profile?.specialty },
              { label: "Kinh nghiệm", value: profile?.experience_text },
            ]}
          />
        )}
        {profile && (
          <div className="certificate-section">
            <h2>Chứng chỉ chuyên môn</h2>
            <p className="subtle">
              PDF, PNG hoặc JPEG · tối đa 5 MB · chỉ bạn và quản lý/Admin được
              tải.
            </p>
            <Notice>{error}</Notice>
            <div className="form-actions">
              <CertificateButton profile={profile} />
            </div>
            {editable && (
              <form
                onSubmit={async (e) => {
                  e.preventDefault();
                  if (!file || uploadBusy) return;
                  setUploadBusy(true);
                  setError("");
                  try {
                    if (file.size > 5 * 1024 * 1024)
                      throw new Error("Chứng chỉ không được vượt quá 5 MB.");
                    const body = new FormData();
                    body.set("file", file);
                    const r = await request<Profile>(
                      "/api/v1/pts/me/certificates",
                      { method: "POST", body },
                      session.accessToken,
                    );
                    update(r.data);
                    setFile(null);
                  } catch (e) {
                    setError((e as Error).message);
                  } finally {
                    setUploadBusy(false);
                  }
                }}
              >
                <Field
                  label="Chọn chứng chỉ"
                  hint="Chọn file minh chứng chuyên môn để gửi cùng hồ sơ."
                >
                  <Input
                    type="file"
                    accept=".pdf,.png,.jpg,.jpeg"
                    disabled={uploadBusy}
                    onChange={(e) => setFile(e.target.files?.[0] ?? null)}
                  />
                </Field>
                <Button
                  type="submit"
                  variant="secondary"
                  busy={uploadBusy}
                  disabled={!file}
                >
                  Tải chứng chỉ lên
                </Button>
              </form>
            )}
            {editable && (
              <div className="form-actions">
                <Button
                  disabled={!profile.has_certificate}
                  onClick={() => setSubmit(true)}
                >
                  Gửi hồ sơ xét duyệt
                </Button>
              </div>
            )}
          </div>
        )}
      </section>
      {profile && <SchedulePanel prefix="pts/me" />}
      {submit && (
        <OperationDialog
          title="Gửi hồ sơ xét duyệt"
          description="Hồ sơ sẽ chuyển sang Chờ duyệt và không thể chỉnh sửa trong khi xét duyệt."
          close={() => setSubmit(false)}
          done={() => setSubmit(false)}
          save={async () => {
            const r = await resource.save<Profile>(
              session.accessToken,
              "pts/me/submit",
              {},
            );
            update(r.data);
          }}
        />
      )}
    </>
  );
}
export function SchedulePanel({ prefix }: { prefix: string }) {
  const hours = useRecord<WorkingHour[]>(`${prefix}/availability`);
  const absences = useList<Absence>(`${prefix}/unavailability`);
  const [editing, setEditing] = useState(false);
  const [adding, setAdding] = useState(false);
  const [remove, setRemove] = useState<Absence | null>(null);
  const [notice, setNotice] = useState("");
  const { session } = useAuth();
  return (
    <section className="panel detail-panel schedule-panel">
      <div className="section-heading">
        <h2>Lịch làm việc</h2>
        <p>
          Múi giờ Việt Nam. Mọi thay đổi được kiểm tra với booking đang giữ chỗ.
        </p>
      </div>
      <Notice tone="success">{notice}</Notice>
      {hours.data ? (
        <>
          <div className="schedule-heading">
            <h3>Khung giờ hàng tuần</h3>
            <Button variant="secondary" onClick={() => setEditing(true)}>
              Chỉnh sửa lịch
            </Button>
          </div>
          {hours.data.length ? (
            <div className="hours-grid">
              {hours.data.map((h, i) => (
                <div className="hour-card" key={i}>
                  <strong>{dayNames[h.day_of_week]}</strong>
                  <span>
                    {h.start_time}–{h.end_time}
                  </span>
                </div>
              ))}
            </div>
          ) : (
            <p className="subtle">Chưa thiết lập lịch hàng tuần.</p>
          )}
        </>
      ) : (
        <RecordState {...hours} retry={hours.refresh} />
      )}
      <div className="schedule-heading">
        <h3>Khoảng nghỉ</h3>
        <Button variant="secondary" onClick={() => setAdding(true)}>
          Thêm khoảng nghỉ
        </Button>
      </div>
      {!absences.loading && !absences.error ? (
        absences.data.length ? (
          <ul className="absence-list">
            {absences.data.map((a) => (
              <li key={a.id}>
                <div>
                  <strong>
                    {vietnamDateTime(a.starts_at)}{" "}
                    →{" "}
                    {vietnamDateTime(a.ends_at)}
                  </strong>
                  <span>{a.reason}</span>
                </div>
                <Button variant="ghost" onClick={() => setRemove(a)}>
                  Hủy khoảng nghỉ
                </Button>
              </li>
            ))}
          </ul>
        ) : (
          <p className="subtle">Chưa có khoảng nghỉ.</p>
        )
      ) : (
        <ListState
          {...absences}
          empty={false}
          filtered={false}
          retry={absences.refresh}
        >
          {null}
        </ListState>
      )}
      <Pagination meta={absences.meta} change={absences.filter} />
      {editing && (
        <HoursDialog
          hours={hours.data ?? []}
          prefix={prefix}
          close={() => setEditing(false)}
          done={() => {
            setEditing(false);
            setNotice("Đã cập nhật lịch làm việc.");
            hours.refresh();
          }}
        />
      )}
      {adding && (
        <AbsenceDialog
          prefix={prefix}
          close={() => setAdding(false)}
          done={() => {
            setAdding(false);
            setNotice("Đã thêm khoảng nghỉ.");
            absences.refresh();
          }}
        />
      )}
      {remove && (
        <OperationDialog
          title="Hủy khoảng nghỉ"
          description={remove.reason}
          warning="PT sẽ có thể nhận lịch trong khoảng này nếu lịch hàng tuần cho phép. Nhật ký hủy được giữ lại."
          close={() => setRemove(null)}
          done={() => {
            setRemove(null);
            absences.refresh();
            setNotice("Đã hủy khoảng nghỉ.");
          }}
          save={() =>
            request(
              `/api/v1/${prefix}/unavailability/${remove.id}`,
              { method: "DELETE" },
              session.accessToken,
            )
          }
        />
      )}
    </section>
  );
}
function HoursDialog({
  hours,
  prefix,
  close,
  done,
}: {
  hours: WorkingHour[];
  prefix: string;
  close: () => void;
  done: () => void;
}) {
  const { session } = useAuth();
  const [rows, setRows] = useState<WorkingHour[]>(
    hours.map((h) => ({
      day_of_week: h.day_of_week,
      start_time: h.start_time,
      end_time: h.end_time,
    })),
  );
  function change(
    index: number,
    key: keyof WorkingHour,
    value: string | number,
  ) {
    setRows((prev) =>
      prev.map((r, i) => (i === index ? { ...r, [key]: value } : r)),
    );
  }
  return (
    <OperationDialog
      title="Lưu lịch hàng tuần"
      description="Kéo để thêm hoặc bỏ cả dải giờ. Khung giờ cùng ngày không chồng nhau; các trường chi tiết giữ được cả mốc phút lẻ."
      close={close}
      done={done}
      save={() =>
        resource.save(
          session.accessToken,
          `${prefix}/availability`,
          { hours: rows },
          "PUT",
        )
      }
    >
      <AvailabilityGrid
        hours={rows}
        onChange={(next) =>
          setRows(
            next.map((hour) => ({
              day_of_week: hour.day_of_week,
              start_time: hour.start_time,
              end_time: hour.end_time,
            })),
          )
        }
      />
      <div className="hours-editor-heading">
        <h3>Chỉnh chính xác từng khung giờ</h3>
        <p>
          Dùng các trường này cho mốc phút lẻ hoặc khi cần chỉnh từng khoảng.
        </p>
      </div>
      {rows.map((h, i) => (
        <div className="hours-editor-row" key={i}>
          <Field label={`Ngày ${i + 1}`}>
            <Select
              value={h.day_of_week}
              onChange={(e) => change(i, "day_of_week", Number(e.target.value))}
            >
              {dayNames.slice(1).map((d, j) => (
                <option key={d} value={j + 1}>
                  {d}
                </option>
              ))}
            </Select>
          </Field>
          <Field label={`Bắt đầu ${i + 1}`}>
            <Input
              type="time"
              required
              value={h.start_time}
              onChange={(e) => change(i, "start_time", e.target.value)}
            />
          </Field>
          <Field label={`Kết thúc ${i + 1}`}>
            <Input
              type="time"
              required
              value={h.end_time}
              onChange={(e) => change(i, "end_time", e.target.value)}
            />
          </Field>
          <Button
            type="button"
            variant="ghost"
            aria-label={`Bỏ khung giờ ${i + 1}`}
            onClick={() => setRows((p) => p.filter((_, j) => j !== i))}
          >
            Bỏ
          </Button>
        </div>
      ))}
      <Button
        type="button"
        variant="secondary"
        disabled={rows.length >= 50}
        onClick={() =>
          setRows((p) => [
            ...p,
            { day_of_week: 1, start_time: "08:00", end_time: "17:00" },
          ])
        }
      >
        Thêm khung giờ
      </Button>
    </OperationDialog>
  );
}
function AbsenceDialog({
  prefix,
  close,
  done,
}: {
  prefix: string;
  close: () => void;
  done: () => void;
}) {
  const { session } = useAuth();
  const [start, setStart] = useState("");
  const [end, setEnd] = useState("");
  const [reason, setReason] = useState("");
  return (
    <OperationDialog
      title="Thêm khoảng nghỉ"
      description="Nhập ngày/giờ Việt Nam. Khoảng nghỉ không được trùng nhau hoặc ảnh hưởng booking đang giữ chỗ."
      close={close}
      done={done}
      save={() =>
        resource.save(session.accessToken, `${prefix}/unavailability`, {
          starts_at: `${start}:00+07:00`,
          ends_at: `${end}:00+07:00`,
          reason,
        })
      }
    >
      <Field label="Bắt đầu nghỉ" required>
        <Input
          type="datetime-local"
          required
          value={start}
          onChange={(e) => setStart(e.target.value)}
        />
      </Field>
      <Field label="Kết thúc nghỉ" required>
        <Input
          type="datetime-local"
          required
          min={start}
          value={end}
          onChange={(e) => setEnd(e.target.value)}
        />
      </Field>
      <Field label="Lý do nghỉ" required>
        <Input
          value={reason}
          onChange={(e) => setReason(e.target.value)}
          required
        />
      </Field>
    </OperationDialog>
  );
}
