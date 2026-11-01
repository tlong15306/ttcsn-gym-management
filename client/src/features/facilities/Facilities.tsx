import { useState, useEffect, useRef } from "react";
import { Link, useParams, useNavigate } from "react-router-dom";
import { resource } from "../../api";
import { civilDate, vietnamDateTime, vietnamTime } from "../../utils/vietnamTime";
import { useAuth } from "../../app/session";
import {
  Field,
  Input,
  Select,
  Button,
  Badge,
  ActionMenu,
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
  Pagination,
  ListState,
  SearchBox,
  useList,
} from "../../components/shared/list";
import { OperationDialog } from "../../components/shared/operation";
import { RoomCoverManager } from '../booking/RoomImage';

export type Room = {
  has_cover?: boolean;
  cover_alt_text?: string;
  cover_version?: number;
  cover_is_demo?: boolean;
  id: number;
  name: string;
  room_type: string;
  default_capacity: number;
  description: string;
  status: string;
};
export type TimeSlot = {
  id: number;
  name: string;
  start_time: string;
  end_time: string;
  is_active: boolean;
};
export type RoomSession = {
  id: number;
  room_id: number;
  room_name: string;
  time_slot_id: number;
  slot_name: string;
  session_date: string;
  starts_at: string;
  ends_at: string;
  capacity: number;
  active_holds: number;
  status: string;
  change_reason: string;
};
type Impact = { future_sessions: number; active_holds: number };
const root = "/management";
function StatusFilter({
  values,
  value,
  change,
}: {
  values: [string, string][];
  value: string;
  change: (v: string) => void;
}) {
  return (
    <Select
      aria-label="Lọc trạng thái"
      value={value}
      onChange={(e) => change(e.target.value)}
    >
      <option value="">Tất cả trạng thái</option>
      {values.map(([v, l]) => (
        <option key={v} value={v}>
          {l}
        </option>
      ))}
    </Select>
  );
}
export function RoomsPage() {
  const list = useList<Room>("rooms");
  const navigate = useNavigate();
  const filtered = !!(list.params.get("q") || list.params.get("status"));
  return (
    <>
      <PageHeader
        title="Phòng tập"
        description="Quản lý không gian tập, sức chứa mặc định và trạng thái sử dụng."
        action={{ label: "Thêm phòng", to: `${root}/rooms/new` }}
      />
      <section className="panel">
        <div className="filter-bar">
          <SearchBox
            value={list.params.get("q") ?? ""}
            onChange={(v) => list.filter("q", v)}
            placeholder="Tìm tên hoặc loại phòng…"
          />
          <StatusFilter
            value={list.params.get("status") ?? ""}
            change={(v) => list.filter("status", v)}
            values={[
              ["AVAILABLE", "Sẵn sàng"],
              ["MAINTENANCE", "Bảo trì"],
              ["INACTIVE", "Ngừng hoạt động"],
            ]}
          />
          {filtered && (
            <Button variant="ghost" onClick={list.clear}>
              Xóa bộ lọc
            </Button>
          )}
        </div>
        <div className="list-caption">
          <span>Danh sách phòng tập</span>
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
                label: "Phòng tập",
                render: (r) => (
                  <Link className="row-title" to={`${root}/rooms/${r.id}`}>
                    {r.name}
                  </Link>
                ),
              },
              { label: "Loại phòng", render: (r) => r.room_type },
              {
                label: "Sức chứa mặc định",
                render: (r) => `${r.default_capacity} chỗ`,
              },
              {
                label: "Trạng thái",
                render: (r) => <Badge value={r.status} />,
              },
            ]}
            actions={(r) => (
              <ActionMenu
                name={r.name}
                actions={[
                  {
                    label: "Xem chi tiết",
                    run: () => navigate(`${root}/rooms/${r.id}`),
                  },
                  {
                    label: "Chỉnh sửa / trạng thái",
                    run: () => navigate(`${root}/rooms/${r.id}/edit`),
                  },
                ]}
              />
            )}
          />
        </ListState>
        <Pagination meta={list.meta} change={list.filter} />
      </section>
    </>
  );
}
export function RoomFormPage() {
  const { id } = useParams();
  const record = useRecord<Room>(id ? `rooms/${id}` : null);
  if (id && !record.data)
    return <RecordState {...record} retry={record.refresh} />;
  return <RoomForm key={id ?? "new"} room={record.data} />;
}
function RoomForm({ room }: { room: Room | null }) {
  const { session } = useAuth();
  const navigate = useNavigate();
  const createdID = useRef<number | null>(null);
  const [v, setV] = useState({
    name: room?.name ?? "",
    room_type: room?.room_type ?? "",
    default_capacity: room?.default_capacity ?? 20,
    description: room?.description ?? "",
    status: room?.status ?? "AVAILABLE",
    reason: "",
  });
  const impact = useRecord<Impact>(room ? `rooms/${room.id}/impact` : null);
  const change = (key: keyof typeof v, value: string | number) =>
    setV((p) => ({ ...p, [key]: value }));
  const statusChanged = !!room && v.status !== room.status;
  return (
    <>
      <PageHeader
        title={room ? "Chỉnh sửa phòng tập" : "Thêm phòng tập"}
        description="Sức chứa mặc định chỉ áp dụng cho phiên mới. Phiên đã tạo giữ snapshot riêng."
        back={`${root}/rooms`}
      />
      <section className="panel form-panel">
        <DataForm
          back={`${root}/rooms`}
          submitLabel={room ? "Lưu thay đổi" : "Tạo phòng"}
          onSuccess={() =>
            navigate(
              room
                ? `${root}/rooms`
                : `${root}/rooms/${createdID.current}/calendar`,
            )
          }
          save={async () => {
            const result = await resource.save<Room>(
              session.accessToken,
              room ? `rooms/${room.id}` : "rooms",
              v,
              room ? "PATCH" : "POST",
            );
            createdID.current = result.data.id;
            return result;
          }}
        >
          <div className="form-grid">
            <Field label="Tên phòng" required>
              <Input
                value={v.name}
                onChange={(e) => change("name", e.target.value)}
                required
                maxLength={100}
              />
            </Field>
            <Field label="Loại phòng" required hint="Ví dụ: Gym, Yoga, Cardio.">
              <Input
                value={v.room_type}
                onChange={(e) => change("room_type", e.target.value)}
                required
                maxLength={50}
              />
            </Field>
            <Field label="Sức chứa mặc định" required>
              <Input
                type="number"
                min={1}
                max={10000}
                required
                value={v.default_capacity}
                onChange={(e) =>
                  change("default_capacity", Number(e.target.value))
                }
              />
            </Field>
            <Field label="Trạng thái">
              <Select
                value={v.status}
                onChange={(e) => change("status", e.target.value)}
              >
                <option value="AVAILABLE">Sẵn sàng</option>
                <option value="MAINTENANCE">Bảo trì</option>
                <option value="INACTIVE">Ngừng hoạt động</option>
              </Select>
            </Field>
          </div>
          <Field label="Mô tả">
            <textarea
              value={v.description}
              onChange={(e) => change("description", e.target.value)}
              maxLength={4000}
            />
          </Field>
          {room && (
            <>
              <Field label="Lý do thay đổi" required>
                <Input
                  value={v.reason}
                  onChange={(e) => change("reason", e.target.value)}
                  required
                />
              </Field>
              {statusChanged && (
                <>
                  <Notice tone="info">
                    {impact.loading
                      ? "Đang kiểm tra lịch liên quan…"
                      : impact.error ||
                        `${impact.data?.future_sessions ?? 0} phiên chưa kết thúc · ${impact.data?.active_holds ?? 0} booking đang giữ chỗ. Bảo trì/ngừng hoạt động sẽ đóng các phiên OPEN sắp tới. Nếu có booking, hệ thống yêu cầu xử lý lịch trước. Mở lại phòng không tự mở lại các phiên.`}
                  </Notice>
                  <label className="check-field">
                    <input
                      type="checkbox"
                      required
                      disabled={impact.loading || !!impact.error}
                    />
                    Tôi đã đọc tác động của việc đổi trạng thái.
                  </label>
                </>
              )}
            </>
          )}
        </DataForm>
      </section>
    </>
  );
}
export function RoomDetailPage() {
  const { id } = useParams();
  const record = useRecord<Room>(`rooms/${id}`);
  const r = record.data;
  if (!r) return <RecordState {...record} retry={record.refresh} />;
  return (
    <>
      <PageHeader
        title={r.name}
        description={r.room_type}
        back={`${root}/rooms`}
        action={{ label: "Chỉnh sửa", to: `${root}/rooms/${id}/edit` }}
      />
      <section className="panel detail-panel">
        <Details
          items={[
            { label: "Sức chứa mặc định", value: `${r.default_capacity} chỗ` },
            { label: "Trạng thái", value: <Badge value={r.status} /> },
            { label: "Mô tả", value: r.description },
          ]}
        />
        <Link
          className="btn btn-secondary"
          to={`${root}/room-sessions?room_id=${r.id}`}
        >
          Xem phiên phòng
        </Link>
        <Link className="btn btn-primary" to={`${root}/rooms/${r.id}/calendar`}>
          Đăng ký lịch hoạt động
        </Link>
      </section>
      <RoomCoverManager room={r} refresh={record.refresh} />
    </>
  );
}
export function TimeSlotsPage() {
  const list = useList<TimeSlot>("time-slots");
  const { session } = useAuth();
  const navigate = useNavigate();
  const [target, setTarget] = useState<TimeSlot | null>(null);
  const [reason, setReason] = useState("");
  const [notice, setNotice] = useState("");
  const filtered = !!(list.params.get("q") || list.params.get("status"));
  return (
    <>
      <PageHeader
        title="Ca tập"
        description="Các khung giờ dùng chung. Ca liền kề được phép; ca đang hoạt động không được chồng giờ."
        action={{ label: "Thêm ca", to: `${root}/time-slots/new` }}
      />
      <Notice tone="success">{notice}</Notice>
      <section className="panel">
        <div className="filter-bar">
          <SearchBox
            value={list.params.get("q") ?? ""}
            onChange={(v) => list.filter("q", v)}
            placeholder="Tìm tên ca…"
          />
          <StatusFilter
            value={list.params.get("status") ?? ""}
            change={(v) => list.filter("status", v)}
            values={[
              ["ACTIVE", "Hoạt động"],
              ["INACTIVE", "Ngừng hoạt động"],
            ]}
          />
          {filtered && (
            <Button variant="ghost" onClick={list.clear}>
              Xóa bộ lọc
            </Button>
          )}
        </div>
        <div className="list-caption">
          <span>Danh sách ca</span>
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
                label: "Ca tập",
                render: (t) => <span className="row-title">{t.name}</span>,
              },
              { label: "Bắt đầu", render: (t) => t.start_time },
              { label: "Kết thúc", render: (t) => t.end_time },
              {
                label: "Trạng thái",
                render: (t) => (
                  <Badge value={t.is_active ? "ACTIVE" : "INACTIVE"} />
                ),
              },
            ]}
            actions={(t) => (
              <ActionMenu
                name={t.name}
                actions={[
                  {
                    label: "Chỉnh sửa ca",
                    run: () => navigate(`${root}/time-slots/${t.id}/edit`),
                  },
                  {
                    label: t.is_active ? "Ngừng hoạt động" : "Kích hoạt ca",
                    danger: t.is_active,
                    run: () => {
                      setTarget(t);
                      setReason("");
                    },
                  },
                ]}
              />
            )}
          />
        </ListState>
        <Pagination meta={list.meta} change={list.filter} />
      </section>
      {target && (
        <OperationDialog
          title={target.is_active ? "Ngừng hoạt động" : "Kích hoạt ca"}
          description={target.name}
          warning="Phiên đã tạo giữ nguyên thời gian. Ngừng ca chỉ ngăn tạo phiên mới; kích hoạt sẽ kiểm tra chồng giờ."
          close={() => setTarget(null)}
          done={() => {
            setTarget(null);
            setNotice("Đã cập nhật ca tập.");
            list.refresh();
          }}
          save={() =>
            resource.save(
              session.accessToken,
              `time-slots/${target.id}/status`,
              { is_active: !target.is_active, reason },
              "PATCH",
            )
          }
        >
          <Field label="Lý do" required>
            <Input
              required
              value={reason}
              onChange={(e) => setReason(e.target.value)}
            />
          </Field>
        </OperationDialog>
      )}
    </>
  );
}
export function TimeSlotFormPage() {
  const { id } = useParams();
  const record = useRecord<TimeSlot>(id ? `time-slots/${id}` : null);
  if (id && !record.data)
    return <RecordState {...record} retry={record.refresh} />;
  return <TimeSlotForm key={id ?? "new"} slot={record.data} />;
}
function TimeSlotForm({ slot }: { slot: TimeSlot | null }) {
  const { session } = useAuth();
  const navigate = useNavigate();
  const [name, setName] = useState(slot?.name ?? "");
  const [start, setStart] = useState(slot?.start_time ?? "");
  const [end, setEnd] = useState(slot?.end_time ?? "");
  const [reason, setReason] = useState("");
  return (
    <>
      <PageHeader
        title={slot ? "Chỉnh sửa ca" : "Thêm ca tập"}
        description="Ca đã được dùng trong phiên chỉ có thể ngừng hoạt động. Giữ giờ lịch sử cho mọi phiên."
        back={`${root}/time-slots`}
      />
      <section className="panel form-panel">
        <DataForm
          back={`${root}/time-slots`}
          submitLabel={slot ? "Lưu thay đổi" : "Tạo ca"}
          onSuccess={() => navigate(`${root}/time-slots`)}
          save={() => {
            if (end <= start)
              throw new Error("Giờ kết thúc phải sau giờ bắt đầu.");
            return resource.save(
              session.accessToken,
              slot ? `time-slots/${slot.id}` : "time-slots",
              {
                name,
                start_time: start,
                end_time: end,
                ...(slot ? { reason } : {}),
              },
              slot ? "PATCH" : "POST",
            );
          }}
        >
          <Field label="Tên ca" required>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              maxLength={80}
            />
          </Field>
          <div className="form-grid">
            <Field label="Giờ bắt đầu" required>
              <Input
                type="time"
                value={start}
                onChange={(e) => setStart(e.target.value)}
                required
              />
            </Field>
            <Field label="Giờ kết thúc" required>
              <Input
                type="time"
                value={end === "24:00" ? "" : end}
                onChange={(e) => setEnd(e.target.value)}
                required={end !== "24:00"}
                disabled={end === "24:00"}
              />
            </Field>
          </div>
          <label className="check-field">
            <input
              type="checkbox"
              checked={end === "24:00"}
              onChange={(e) => setEnd(e.target.checked ? "24:00" : "23:00")}
            />{" "}
            Kết thúc lúc 24:00 (cuối ngày)
          </label>
          {slot && (
            <Field label="Lý do thay đổi" required>
              <Input
                value={reason}
                onChange={(e) => setReason(e.target.value)}
                required
              />
            </Field>
          )}
        </DataForm>
      </section>
    </>
  );
}
export function RoomSessionsPage() {
  const list = useList<RoomSession>("room-sessions");
  const navigate = useNavigate();
  const filtered = !!(
    list.params.get("q") ||
    list.params.get("status") ||
    list.params.get("date") ||
    list.params.get("room_id")
  );
  return (
    <>
      <PageHeader
        title="Phiên phòng"
        description="Lịch phòng theo ngày và ca. Sức chứa, thời gian được chốt khi tạo phiên."
        action={{ label: "Thêm phiên", to: `${root}/room-sessions/new` }}
      />
      <section className="panel">
        <div className="filter-bar">
          <SearchBox
            value={list.params.get("q") ?? ""}
            onChange={(v) => list.filter("q", v)}
            placeholder="Tìm phòng hoặc ca…"
          />
          <Input
            className="date-filter"
            aria-label="Lọc ngày"
            type="date"
            value={list.params.get("date") ?? ""}
            onChange={(e) => list.filter("date", e.target.value)}
          />
          <StatusFilter
            value={list.params.get("status") ?? ""}
            change={(v) => list.filter("status", v)}
            values={[
              ["OPEN", "Đang mở"],
              ["CLOSED", "Đã đóng"],
              ["MAINTENANCE", "Bảo trì"],
            ]}
          />
          {filtered && (
            <Button variant="ghost" onClick={list.clear}>
              Xóa bộ lọc
            </Button>
          )}
        </div>
        <div className="list-caption">
          <span>Lịch phiên phòng</span>
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
                label: "Phòng / ca",
                render: (s) => (
                  <div>
                    <Link
                      className="row-title"
                      to={`${root}/room-sessions/${s.id}`}
                    >
                      {s.room_name}
                    </Link>
                    <span className="cell-subtitle">{s.slot_name}</span>
                  </div>
                ),
              },
              {
                label: "Ngày / giờ",
                render: (s) => (
                  <div>
                    {civilDate(s.session_date)}
                    <span className="cell-subtitle">
                      {vietnamTime(s.starts_at)}
                      –
                      {vietnamTime(s.ends_at)}
                    </span>
                  </div>
                ),
              },
              {
                label: "Đã giữ / sức chứa",
                render: (s) => `${s.active_holds} / ${s.capacity}`,
              },
              {
                label: "Trạng thái",
                render: (s) => <Badge value={s.status} />,
              },
            ]}
            actions={(s) => (
              <ActionMenu
                name={`${s.room_name} ${s.session_date}`}
                actions={[
                  {
                    label: "Xem chi tiết",
                    run: () => navigate(`${root}/room-sessions/${s.id}`),
                  },
                  {
                    label: "Sức chứa / trạng thái",
                    run: () => navigate(`${root}/room-sessions/${s.id}/edit`),
                  },
                ]}
              />
            )}
          />
        </ListState>
        <Pagination meta={list.meta} change={list.filter} />
      </section>
    </>
  );
}
export function SessionFormPage() {
  const { id } = useParams();
  const record = useRecord<RoomSession>(id ? `room-sessions/${id}` : null);
  if (id && !record.data)
    return <RecordState {...record} retry={record.refresh} />;
  return <SessionForm key={id ?? "new"} item={record.data} />;
}
function SessionForm({ item }: { item: RoomSession | null }) {
  const { session } = useAuth();
  const navigate = useNavigate();
  const [rooms, setRooms] = useState<Room[]>([]);
  const [slots, setSlots] = useState<TimeSlot[]>([]);
  const [lookupError, setLookupError] = useState("");
  const [roomQuery, setRoomQuery] = useState("");
  const [slotQuery, setSlotQuery] = useState("");
  const [roomID, setRoomID] = useState(String(item?.room_id ?? ""));
  const [slotID, setSlotID] = useState(String(item?.time_slot_id ?? ""));
  const [date, setDate] = useState(item?.session_date ?? "");
  const [capacity, setCapacity] = useState(item?.capacity ?? 0);
  const [status, setStatus] = useState(item?.status ?? "OPEN");
  const [reason, setReason] = useState("");
  useEffect(() => {
    if (item) return;
    const abort = new AbortController();
    const timer = setTimeout(() => {
      setLookupError("");
      Promise.all([
        resource.list<Room>(
          session.accessToken,
          "rooms",
          new URLSearchParams({
            status: "AVAILABLE",
            q: roomQuery,
            page_size: "100",
          }),
          abort.signal,
        ),
        resource.list<TimeSlot>(
          session.accessToken,
          "time-slots",
          new URLSearchParams({
            status: "ACTIVE",
            q: slotQuery,
            page_size: "100",
          }),
          abort.signal,
        ),
      ])
        .then(([r, t]) => {
          setRooms(r.data);
          setSlots(t.data);
        })
        .catch((e) => {
          if (!abort.signal.aborted) setLookupError(e.message);
        });
    }, 300);
    return () => {
      clearTimeout(timer);
      abort.abort();
    };
  }, [item, session.accessToken, roomQuery, slotQuery]);
  return (
    <>
      <PageHeader
        title={item ? "Cập nhật phiên phòng" : "Thêm phiên phòng"}
        description={
          item
            ? `${item.room_name} · ${item.slot_name} · ${item.session_date}`
            : "Chọn phòng và ca đang hoạt động. Giờ được chuyển theo múi giờ Việt Nam."
        }
        back={`${root}/room-sessions`}
      />
      <section className="panel form-panel">
        <Notice>{lookupError}</Notice>
        <DataForm
          back={`${root}/room-sessions`}
          submitLabel={item ? "Lưu thay đổi" : "Tạo phiên"}
          onSuccess={() => navigate(`${root}/room-sessions`)}
          save={() =>
            resource.save(
              session.accessToken,
              item ? `room-sessions/${item.id}` : "room-sessions",
              item
                ? { capacity, status, reason }
                : {
                    room_id: Number(roomID),
                    time_slot_id: Number(slotID),
                    session_date: date,
                    capacity,
                  },
              item ? "PATCH" : "POST",
            )
          }
        >
          {!item && (
            <>
              <div className="form-grid">
                <Field label="Tìm phòng">
                  <Input
                    value={roomQuery}
                    onChange={(e) => setRoomQuery(e.target.value)}
                    placeholder="Thu hẹp danh sách phòng"
                  />
                </Field>
                <Field label="Tìm ca">
                  <Input
                    value={slotQuery}
                    onChange={(e) => setSlotQuery(e.target.value)}
                    placeholder="Thu hẹp danh sách ca"
                  />
                </Field>
                <Field label="Phòng" required>
                  <Select
                    required
                    value={roomID}
                    onChange={(e) => setRoomID(e.target.value)}
                  >
                    <option value="">Chọn phòng</option>
                    {rooms.map((r) => (
                      <option key={r.id} value={r.id}>
                        {r.name} · {r.default_capacity} chỗ
                      </option>
                    ))}
                  </Select>
                </Field>
                <Field label="Ca" required>
                  <Select
                    required
                    value={slotID}
                    onChange={(e) => setSlotID(e.target.value)}
                  >
                    <option value="">Chọn ca</option>
                    {slots.map((t) => (
                      <option key={t.id} value={t.id}>
                        {t.name} · {t.start_time}–{t.end_time}
                      </option>
                    ))}
                  </Select>
                </Field>
              </div>
              <Field label="Ngày tập" required>
                <Input
                  type="date"
                  required
                  value={date}
                  onChange={(e) => setDate(e.target.value)}
                />
              </Field>
            </>
          )}
          <Field
            label="Sức chứa"
            required
            hint={
              item
                ? `Đang giữ ${item.active_holds} chỗ; không được giảm thấp hơn.`
                : "Nhập 0 để lấy sức chứa mặc định của phòng."
            }
          >
            <Input
              type="number"
              min={item ? Math.max(1, item.active_holds) : 0}
              max={10000}
              required
              value={capacity}
              onChange={(e) => setCapacity(Number(e.target.value))}
            />
          </Field>
          {item && (
            <>
              <Field label="Trạng thái">
                <Select
                  value={status}
                  onChange={(e) => setStatus(e.target.value)}
                >
                  <option value="OPEN">Đang mở</option>
                  <option value="CLOSED">Đã đóng</option>
                  <option value="MAINTENANCE">Bảo trì</option>
                </Select>
              </Field>
              <Notice tone="info">
                Đóng/bảo trì phiên đang có booking sẽ bị từ chối. Phòng phải sẵn
                sàng trước khi mở phiên.
              </Notice>
              <Field label="Lý do thay đổi" required>
                <Input
                  value={reason}
                  onChange={(e) => setReason(e.target.value)}
                  required
                />
              </Field>
              {status !== item.status && (
                <label className="check-field">
                  <input type="checkbox" required />
                  Tôi xác nhận đổi trạng thái phiên này.
                </label>
              )}
            </>
          )}
        </DataForm>
      </section>
    </>
  );
}
export function SessionDetailPage() {
  const { id } = useParams();
  const record = useRecord<RoomSession>(`room-sessions/${id}`);
  const s = record.data;
  if (!s) return <RecordState {...record} retry={record.refresh} />;
  return (
    <>
      <PageHeader
        title={s.room_name}
        description={`${s.slot_name} · ${s.session_date}`}
        back={`${root}/room-sessions`}
        action={{ label: "Chỉnh sửa", to: `${root}/room-sessions/${id}/edit` }}
      />
      <section className="panel detail-panel">
        <Details
          items={[
            {
              label: "Bắt đầu (Việt Nam)",
              value: vietnamDateTime(s.starts_at),
            },
            {
              label: "Kết thúc (Việt Nam)",
              value: vietnamDateTime(s.ends_at),
            },
            {
              label: "Đang giữ / sức chứa",
              value: `${s.active_holds} / ${s.capacity}`,
            },
            { label: "Trạng thái", value: <Badge value={s.status} /> },
            { label: "Lý do cập nhật", value: s.change_reason },
          ]}
        />
      </section>
    </>
  );
}
