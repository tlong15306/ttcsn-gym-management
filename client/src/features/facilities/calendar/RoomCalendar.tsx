import { useEffect, useMemo, useRef, useState } from "react";
import { Link, useBlocker, useParams } from "react-router-dom";
import { vietnamTime } from "../../../utils/vietnamTime";
import {
  CalendarDays,
  Clock3,
  Undo2,
  Redo2,
  Copy,
  Check,
  Grid3X3,
  List,
  ArrowRight,
  History,
} from "lucide-react";
import { request, resource } from "../../../api";
import { useAuth } from "../../../app/session";
import {
  Button,
  Field,
  Input,
  Select,
  Modal,
  Notice,
} from "../../../components/ui";
import { PageHeader } from "../../../components/shared/list";
import { useRecord, RecordState } from "../../../components/shared/record";
import { WeeklyGrid, type GridMode } from "./WeeklyGrid";
import {
  emptyWeek,
  toWeek,
  toHours,
  paint,
  clock,
  localDate,
  weekdays,
  type Week,
  type Hours,
} from "./hours";
type Version = {
  id: number;
  effective_from: string;
  effective_until: string;
  hours: Hours[];
  reason: string;
  status: string;
  created_at: string;
};
type Exception = {
  id: number;
  date: string;
  mode: string;
  closure_type: string;
  reason: string;
  status: string;
  hours: Hours[];
};
type Snapshot = {
  room: {
    id: number;
    name: string;
    status: string;
    calendar_revision: number;
    calendar_required?: boolean;
    calendar_prepared_through: string;
    calendar_job_error: string;
  };
  versions: Version[];
  exceptions: Exception[];
};
type Change = {
  session_id: number;
  date: string;
  starts_at: string;
  ends_at: string;
  active_holds: number;
  action: string;
};
type Preview = {
  preview_id: string;
  expires_at: string;
  creates: number;
  changes: Change[];
  conflicts: Change[];
  prepared_through: string;
};
type Mutation = {
  kind: string;
  base_revision: number;
  effective_from?: string;
  date?: string;
  mode?: string;
  closure_type?: string;
  exception_id?: number;
  hours: Hours[];
  reason: string;
  days?: number;
};
type Day = { date: string; source: string; hours: Hours[] };
type Pending = { mutation: Mutation; preview: Preview; key: string };
const sourceLabel: Record<string, string> = {
  WEEKLY: "Lịch tuần",
  CLOSED: "Nghỉ cả ngày",
  REPLACE: "Giờ đặc biệt",
  UNCONFIGURED: "Chưa đăng ký",
  MAINTENANCE: "Phòng bảo trì",
  INACTIVE: "Phòng ngừng hoạt động",
};
export function RoomCalendarPage() {
  const { id } = useParams();
  const [feedback, setFeedback] = useState("");
  const record = useRecord<Snapshot>(`rooms/${id}/weekly-schedule`);
  if (!record.data || record.loading)
    return <RecordState {...record} retry={record.refresh} />;
  return (
    <>
      <Notice tone="success">{feedback}</Notice>
      <Notice>{record.error}</Notice>
      <RoomCalendar
        key={`${id}-${record.data.room.calendar_revision}`}
        snapshot={record.data}
        reload={record.refresh}
        onSaved={() => setFeedback("Đã xác nhận và lưu lịch vào database.")}
      />
    </>
  );
}
export function RoomCalendar({
  snapshot: s,
  reload,
  onSaved,
}: {
  snapshot: Snapshot;
  reload: () => void;
  onSaved?: () => void;
}) {
  const { session } = useAuth(),
    room = s.room,
    path = `rooms/${room.id}`;
  const initialDate = s.versions.length ? localDate(1) : localDate();
  function versionFor(date: string) {
    return s.versions.find(
      (v) =>
        v.status === "ACTIVE" &&
        v.effective_from <= date &&
        (!v.effective_until || v.effective_until > date),
    );
  }
  const initial = useMemo(
    () => toWeek(versionFor(initialDate)?.hours ?? []),
    [],
  );
  const [states, setStates] = useState<Week[]>([initial]),
    [index, setIndex] = useState(0);
  const week = states[index];
  const [effective, setEffective] = useState(initialDate),
    [reason, setReason] = useState("");
  const [mode, setMode] = useState<GridMode>("OPEN"),
    [gridView, setGridView] = useState(() => window.innerWidth >= 768);
  const [day, setDay] = useState(0),
    [start, setStart] = useState(7),
    [end, setEnd] = useState(21),
    [copyTo, setCopyTo] = useState(1),
    [overnight, setOvernight] = useState(false);
  const [busy, setBusy] = useState(false),
    [error, setError] = useState(""),
    [success, setSuccess] = useState("");
  const [pending, setPending] = useState<Pending | null>(null),
    [previewError, setPreviewError] = useState("");
  const [tab, setTab] = useState("WEEKLY"),
    [exception, setException] = useState<Exception | null>(null),
    [exceptionOpen, setExceptionOpen] = useState(false);
  const [exDate, setExDate] = useState(localDate()),
    [exMode, setExMode] = useState("CLOSED"),
    [exType, setExType] = useState("HOLIDAY"),
    [exReason, setExReason] = useState("");
  const [exWeek, setExWeek] = useState<Week>(emptyWeek()),
    [exStart, setExStart] = useState(7),
    [exEnd, setExEnd] = useState(21);
  const [calendarDate, setCalendarDate] = useState(localDate()),
    [days, setDays] = useState<Day[]>([]),
    [loadingDays, setLoadingDays] = useState(false),
    [daysError, setDaysError] = useState("");
  const [historyPage, setHistoryPage] = useState(1);
  const [discardReload, setDiscardReload] = useState(false);
  const dirty =
    JSON.stringify(week) !== JSON.stringify(initial) ||
    effective !== initialDate ||
    !!reason;
  const dirtyRef = useRef(dirty);
  dirtyRef.current = dirty;
  const saving = useRef(false);
  const blocker = useBlocker(
    ({ currentLocation, nextLocation }) =>
      dirtyRef.current && currentLocation.pathname !== nextLocation.pathname,
  );
  useEffect(() => {
    const leave = (e: BeforeUnloadEvent) => {
      if (dirtyRef.current) e.preventDefault();
    };
    window.addEventListener("beforeunload", leave);
    return () => window.removeEventListener("beforeunload", leave);
  }, []);
  useEffect(() => {
    const abort = new AbortController();
    setLoadingDays(true);
    setDaysError("");
    const until = new Date(`${calendarDate}T00:00:00Z`);
    if (Number.isNaN(until.getTime())) {
      setLoadingDays(false);
      return;
    }
    until.setUTCDate(until.getUTCDate() + 6);
    resource
      .get<Day[]>(
        session.accessToken,
        `${path}/calendar?from=${calendarDate}&to=${until.toISOString().slice(0, 10)}`,
        abort.signal,
      )
      .then((r) => {
        if (!abort.signal.aborted) setDays(r.data);
      })
      .catch((e) => {
        if (!abort.signal.aborted) setDaysError(e.message);
      })
      .finally(() => {
        if (!abort.signal.aborted) setLoadingDays(false);
      });
    return () => abort.abort();
  }, [calendarDate, room.calendar_revision, session.accessToken]);
  function change(next: Week) {
    if (JSON.stringify(next) === JSON.stringify(week)) return;
    setStates([...states.slice(0, index + 1), next]);
    setIndex(index + 1);
    setSuccess("");
  }
  function range(confirmed = false) {
    if (start === end) {
      setError("Giờ bắt đầu và kết thúc phải khác nhau.");
      return;
    }
    if (end < start && !confirmed) {
      setOvernight(true);
      return;
    }
    let next = week;
    if (end < start) {
      next = paint(next, [day, start], [day, 23], mode !== "CLOSE");
      if (end > 0)
        next = paint(
          next,
          [(day + 1) % 7, 0],
          [(day + 1) % 7, end - 1],
          mode !== "CLOSE",
        );
    } else next = paint(next, [day, start], [day, end - 1], mode !== "CLOSE");
    change(next);
    setOvernight(false);
    setError("");
  }
  async function preview(mutation: Mutation) {
    if (saving.current) return;
    saving.current = true;
    setBusy(true);
    setError("");
    setPreviewError("");
    try {
      const r = await resource.save<Preview>(
        session.accessToken,
        `${path}/calendar/preview`,
        mutation,
      );
      setPending({ mutation, preview: r.data, key: crypto.randomUUID() });
      setExceptionOpen(false);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Không thể xem trước.");
    } finally {
      saving.current = false;
      setBusy(false);
    }
  }
  async function confirm() {
    if (!pending || saving.current) return;
    saving.current = true;
    setBusy(true);
    setPreviewError("");
    let target = `${path}/weekly-schedule`,
      method = "PUT";
    const m = pending.mutation;
    if (m.kind === "EXCEPTION") {
      target = `${path}/calendar/exceptions${m.exception_id ? `/${m.exception_id}` : ""}`;
      method = m.exception_id ? "PATCH" : "POST";
    }
    if (m.kind === "WITHDRAW") {
      target = `${path}/calendar/exceptions/${m.exception_id}/withdraw`;
      method = "POST";
    }
    if (m.kind === "EXTEND") {
      target = `${path}/calendar/materialize`;
      method = "POST";
    }
    try {
      await request(
        `/api/v1/${target}`,
        {
          method,
          body: JSON.stringify({
            ...m,
            preview_id: pending.preview.preview_id,
          }),
          headers: { "Idempotency-Key": pending.key },
        },
        session.accessToken,
      );
      dirtyRef.current = false;
      setPending(null);
      setSuccess("Đã xác nhận và lưu lịch vào database.");
      onSaved?.();
      reload();
    } catch (e) {
      setPreviewError(e instanceof Error ? e.message : "Không thể lưu lịch.");
    } finally {
      saving.current = false;
      setBusy(false);
    }
  }
  function openException(item: Exception | null) {
    setException(item);
    setExDate(item?.date ?? localDate());
    setExMode(item?.mode ?? "CLOSED");
    setExType(item?.closure_type ?? "HOLIDAY");
    setExReason("");
    setExWeek(toWeek((item?.hours ?? []).map((h) => ({ ...h, weekday: 1 }))));
    setExceptionOpen(true);
  }
  const hours = toHours(week),
    total = week.flat().filter(Boolean).length,
    activeExceptions = s.exceptions.filter((e) => e.status === "ACTIVE");
  const history = [
    ...s.versions.map((v) => ({
      key: `v${v.id}`,
      title: `Lịch tuần · từ ${v.effective_from}`,
      status: v.status,
      reason: v.reason,
    })),
    ...s.exceptions.map((e) => ({
      key: `e${e.id}`,
      title: `Ngoại lệ · ${e.date}`,
      status: e.status,
      reason: e.reason,
    })),
  ].reverse();
  return (
    <>
      <PageHeader
        title={`Lịch hoạt động · ${room.name}`}
        description="Đăng ký giờ mở cửa theo tuần, điều chỉnh ngày đặc biệt và chuẩn bị phiên đặt chỗ."
        back={`/management/rooms/${room.id}`}
      />
      <Notice>{error}</Notice>
      <Notice tone="success">{success}</Notice>
      <Notice>{room.calendar_job_error}</Notice>
      {room.calendar_required === false && (
        <Notice tone="info">
          Phòng cũ đang chờ chuyển sang lịch tuần. Phiên hiện có được giữ
          nguyên. Lần công bố lịch đầu tiên sẽ bật kiểm tra giờ mở cửa cho phòng
          này.
        </Notice>
      )}
      <Button
        variant="ghost"
        disabled={busy}
        onClick={() => (dirty ? setDiscardReload(true) : reload())}
      >
        Tải lại dữ liệu
      </Button>
      {room.status !== "AVAILABLE" && (
        <Notice tone="info">
          Phòng đang{" "}
          {room.status === "MAINTENANCE" ? "bảo trì" : "ngừng hoạt động"}. Giờ
          mở cửa đã đăng ký được giữ, nhưng không có phiên mới khi phòng chưa
          sẵn sàng.
        </Notice>
      )}
      <div className="calendar-summary">
        <div>
          <CalendarDays size={20} />
          <span>
            <strong>
              {s.versions.length ? "Đã có lịch tuần" : "Chưa đăng ký lịch"}
            </strong>
            <small>
              {s.versions.length
                ? `Phiên bản ${room.calendar_revision}`
                : "Công bố lịch trước khi tạo phiên"}
            </small>
          </span>
        </div>
        <div>
          <Clock3 size={20} />
          <span>
            <strong>{total} giờ / tuần</strong>
            <small>Giờ trong bản đang chỉnh</small>
          </span>
        </div>
        <div>
          <Check size={20} />
          <span>
            <strong>
              {room.calendar_prepared_through || "Chưa chuẩn bị phiên"}
            </strong>
            <small>Ngày cuối đã chuẩn bị · cửa sổ 28 ngày</small>
          </span>
        </div>
      </div>
      <div className="calendar-tabs" role="tablist" aria-label="Quản lý lịch">
        {[
          ["WEEKLY", "Lịch tuần"],
          ["EXCEPTIONS", "Ngoại lệ theo ngày"],
          ["EFFECTIVE", "Lịch thực tế"],
          ["HISTORY", "Lịch sử"],
        ].map(([key, label]) => (
          <button
            type="button"
            role="tab"
            aria-selected={tab === key}
            aria-controls={`calendar-${key}`}
            id={`tab-${key}`}
            tabIndex={tab === key ? 0 : -1}
            key={key}
            onClick={() => setTab(key)}
            onKeyDown={(e) => {
              const keys = ["WEEKLY", "EXCEPTIONS", "EFFECTIVE", "HISTORY"];
              const index = keys.indexOf(key);
              let next = index;
              if (e.key === "ArrowRight") next = (index + 1) % 4;
              else if (e.key === "ArrowLeft") next = (index + 3) % 4;
              else if (e.key === "Home") next = 0;
              else if (e.key === "End") next = 3;
              else return;
              e.preventDefault();
              setTab(keys[next]);
              document.getElementById(`tab-${keys[next]}`)?.focus();
            }}
          >
            {label}
            {key === "EXCEPTIONS" && <span>{activeExceptions.length}</span>}
          </button>
        ))}
      </div>
      {tab === "WEEKLY" && (
        <section
          className="panel calendar-panel"
          role="tabpanel"
          id="calendar-WEEKLY"
          aria-labelledby="tab-WEEKLY"
        >
          <div className="calendar-panel-heading">
            <div>
              <h2>Giờ mở cửa hằng tuần</h2>
              <p>
                Kéo chọn vùng để mở/đóng giờ. Thả chuột chỉ sửa bản nháp; chọn
                “Xem trước” để công bố.
              </p>
            </div>
            <span className={`calendar-draft-label ${dirty ? "dirty" : ""}`}>
              {dirty ? "Có thay đổi chưa lưu" : "Bản đang xem"}
            </span>
          </div>
          <div className="calendar-toolbar">
            <div className="calendar-tool-group">
              {(
                [
                  ["OPEN", "Mở giờ"],
                  ["CLOSE", "Đóng giờ"],
                  ["MOVE", "Di chuyển"],
                  ["START", "Đổi đầu"],
                  ["END", "Đổi cuối"],
                ] as [GridMode, string][]
              ).map(([key, label]) => (
                <Button
                  key={key}
                  variant={mode === key ? "primary" : "secondary"}
                  aria-pressed={mode === key}
                  onClick={() => setMode(key)}
                  disabled={busy}
                >
                  {label}
                </Button>
              ))}
            </div>
            <div className="calendar-tool-group">
              <Button
                variant="ghost"
                aria-label="Hoàn tác"
                disabled={index === 0 || busy}
                onClick={() => setIndex(index - 1)}
              >
                <Undo2 size={18} />
              </Button>
              <Button
                variant="ghost"
                aria-label="Làm lại"
                disabled={index === states.length - 1 || busy}
                onClick={() => setIndex(index + 1)}
              >
                <Redo2 size={18} />
              </Button>
              <Button
                variant="secondary"
                onClick={() => setGridView(!gridView)}
              >
                {gridView ? <List size={17} /> : <Grid3X3 size={17} />}
                {gridView ? "Xem danh sách" : "Xem bảng 7 × 24"}
              </Button>
            </div>
          </div>
          {gridView ? (
            <WeeklyGrid
              week={week}
              onChange={change}
              mode={mode}
              disabled={busy}
            />
          ) : (
            <div className="calendar-day-list">
              {weekdays.map((label, d) => (
                <div key={label}>
                  <strong>{label}</strong>
                  <span>
                    {hours.filter((h) => h.weekday === d + 1).length ? (
                      hours
                        .filter((h) => h.weekday === d + 1)
                        .map((h) => (
                          <span
                            className="calendar-hour-pill"
                            key={h.start_minute}
                          >
                            {clock(h.start_minute)}–{clock(h.end_minute)}
                          </span>
                        ))
                    ) : (
                      <small>Đóng cả ngày</small>
                    )}
                  </span>
                  <Button
                    variant="ghost"
                    onClick={() => {
                      setDay(d);
                      document
                        .getElementById("calendar-range")
                        ?.scrollIntoView({
                          block: "center",
                          behavior: "smooth",
                        });
                    }}
                  >
                    Sửa giờ
                  </Button>
                </div>
              ))}
            </div>
          )}
          <p className="calendar-help">
            Ô 07–20 tương ứng 07:00–21:00. Bàn phím: mũi tên để di chuyển, Shift
            + mũi tên để chọn vùng, Enter áp dụng, Esc hủy chọn. Với “Di
            chuyển/Đổi đầu/Đổi cuối”, bắt đầu từ một ô đang mở.
          </p>
          <div className="calendar-range" id="calendar-range">
            <h3>Đặt giờ bằng biểu mẫu</h3>
            <div className="calendar-range-fields">
              <Field label="Ngày">
                <Select
                  value={day}
                  onChange={(e) => setDay(Number(e.target.value))}
                >
                  {weekdays.map((v, i) => (
                    <option value={i} key={v}>
                      {v}
                    </option>
                  ))}
                </Select>
              </Field>
              <Field label="Bắt đầu">
                <Select
                  value={start}
                  onChange={(e) => setStart(Number(e.target.value))}
                >
                  {Array.from({ length: 24 }, (_, i) => (
                    <option key={i} value={i}>
                      {clock(i * 60)}
                    </option>
                  ))}
                </Select>
              </Field>
              <Field label="Kết thúc">
                <Select
                  value={end}
                  onChange={(e) => setEnd(Number(e.target.value))}
                >
                  {Array.from({ length: 25 }, (_, i) => (
                    <option key={i} value={i}>
                      {clock(i * 60)}
                    </option>
                  ))}
                </Select>
              </Field>
              <Button
                variant="secondary"
                onClick={() => range()}
                disabled={busy}
              >
                {mode === "CLOSE" ? "Đóng khoảng giờ" : "Mở khoảng giờ"}
              </Button>
            </div>
            <div className="calendar-copy">
              <Field label="Sao chép ngày đang chọn sang">
                <Select
                  value={copyTo}
                  onChange={(e) => setCopyTo(Number(e.target.value))}
                >
                  {weekdays.map((v, i) => (
                    <option value={i} key={v}>
                      {v}
                    </option>
                  ))}
                </Select>
              </Field>
              <Button
                variant="secondary"
                disabled={busy}
                onClick={() => {
                  const next = week.map((r) => [...r]);
                  next[copyTo] = [...week[day]];
                  change(next);
                }}
              >
                <Copy size={16} />
                Sao chép ngày
              </Button>
              <Button
                variant="secondary"
                disabled={busy}
                onClick={() =>
                  change(
                    toWeek(
                      weekdays.map((_, i) => ({
                        weekday: i + 1,
                        start_minute: 420,
                        end_minute: 1260,
                      })),
                    ),
                  )
                }
              >
                Cả tuần 07:00–21:00
              </Button>
              <Button
                variant="ghost"
                disabled={busy}
                onClick={() => change(emptyWeek())}
              >
                Đóng cả tuần
              </Button>
            </div>
          </div>
          <form
            className="calendar-publish"
            onSubmit={(e) => {
              e.preventDefault();
              preview({
                kind: "WEEKLY",
                base_revision: room.calendar_revision,
                effective_from: effective,
                hours,
                reason,
              });
            }}
          >
            <div className="form-grid">
              <Field
                label="Áp dụng từ ngày"
                hint={
                  s.versions.length
                    ? "Đổi lịch tuần từ ngày mai. Ngoại lệ dùng cho thay đổi hôm nay."
                    : "Phòng mới có thể bắt đầu hôm nay."
                }
                required
              >
                <Input
                  type="date"
                  min={s.versions.length ? localDate(1) : localDate()}
                  max={localDate(56)}
                  required
                  value={effective}
                  onChange={(e) => setEffective(e.target.value)}
                />
              </Field>
              <Field label="Lý do công bố" required>
                <Input
                  value={reason}
                  maxLength={1000}
                  required
                  onChange={(e) => setReason(e.target.value)}
                  placeholder="Ví dụ: đăng ký giờ hoạt động tháng này"
                />
              </Field>
            </div>
            <div className="form-actions">
              <Button
                type="button"
                variant="secondary"
                disabled={busy}
                onClick={() =>
                  change(toWeek(versionFor(effective)?.hours ?? []))
                }
              >
                Nạp lịch đã công bố của ngày này
              </Button>
              <Button type="submit" busy={busy}>
                <ArrowRight size={17} />
                Xem trước và xác nhận
              </Button>
            </div>
          </form>
        </section>
      )}
      {tab === "EXCEPTIONS" && (
        <section
          className="panel calendar-panel"
          role="tabpanel"
          id="calendar-EXCEPTIONS"
          aria-labelledby="tab-EXCEPTIONS"
        >
          <div className="calendar-panel-heading">
            <div>
              <h2>Ngày nghỉ và giờ đặc biệt</h2>
              <p>
                Ngoại lệ thay thế toàn bộ giờ của ngày đó. Đổi lịch tuần vẫn giữ
                các ngoại lệ đã đăng ký.
              </p>
            </div>
            <Button onClick={() => openException(null)}>Thêm ngoại lệ</Button>
          </div>
          {activeExceptions.length === 0 ? (
            <div className="empty-state">
              <CalendarDays />
              <h3>Chưa có ngoại lệ</h3>
              <p>
                Lịch tuần áp dụng đều đặn. Thêm ngày nghỉ, bảo trì hoặc giờ đặc
                biệt khi cần.
              </p>
            </div>
          ) : (
            <div className="calendar-day-list">
              {activeExceptions.map((e) => (
                <div key={e.id}>
                  <span>
                    <strong>{e.date}</strong>
                    <small>
                      {e.mode === "CLOSED"
                        ? "Nghỉ cả ngày"
                        : e.hours
                            .map(
                              (h) =>
                                `${clock(h.start_minute)}–${clock(h.end_minute)}`,
                            )
                            .join(", ")}
                    </small>
                    <small>{e.reason}</small>
                  </span>
                  <div className="calendar-tool-group">
                    <Button
                      variant="secondary"
                      disabled={e.date < localDate() || busy}
                      onClick={() => openException(e)}
                    >
                      Chỉnh sửa
                    </Button>
                    <Button
                      variant="ghost"
                      disabled={e.date < localDate() || busy}
                      onClick={() =>
                        preview({
                          kind: "WITHDRAW",
                          base_revision: room.calendar_revision,
                          exception_id: e.id,
                          reason: "Rút ngoại lệ, khôi phục lịch tuần",
                          hours: [],
                        })
                      }
                    >
                      Rút ngoại lệ
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </section>
      )}
      {tab === "EFFECTIVE" && (
        <section
          className="panel calendar-panel"
          role="tabpanel"
          id="calendar-EFFECTIVE"
          aria-labelledby="tab-EFFECTIVE"
        >
          <h2>Lịch thực tế của phòng</h2>
          <p>
            Trạng thái phòng → ngoại lệ ngày → lịch tuần. Các phiên đã bắt đầu
            giữ snapshot riêng.
          </p>
          <Field label="Xem từ ngày">
            <Input
              type="date"
              value={calendarDate}
              required
              onChange={(e) => setCalendarDate(e.target.value)}
            />
          </Field>
          <Notice>{daysError}</Notice>
          {loadingDays ? (
            <p role="status">Đang tải lịch…</p>
          ) : (
            <div className="calendar-day-list">
              {days.map((d) => (
                <div key={d.date}>
                  <span>
                    <strong>{d.date}</strong>
                    <small>{sourceLabel[d.source] ?? d.source}</small>
                  </span>
                  <span>
                    {d.hours.length
                      ? d.hours.map((h) => (
                          <span
                            className="calendar-hour-pill"
                            key={h.start_minute}
                          >
                            {clock(h.start_minute)}–{clock(h.end_minute)}
                          </span>
                        ))
                      : "Đóng cửa"}
                  </span>
                </div>
              ))}
            </div>
          )}
          <div className="form-actions">
            <Link
              className="btn btn-secondary"
              to={`/management/room-sessions?room_id=${room.id}`}
            >
              Xem phiên đã tạo
            </Link>
            <Button
              busy={busy}
              onClick={() =>
                preview({
                  kind: "EXTEND",
                  base_revision: room.calendar_revision,
                  days: 28,
                  reason: "Bổ sung cửa sổ 28 ngày",
                  hours: [],
                })
              }
            >
              Chuẩn bị 28 ngày tiếp theo
            </Button>
          </div>
        </section>
      )}
      {tab === "HISTORY" && (
        <section
          className="panel calendar-panel"
          role="tabpanel"
          id="calendar-HISTORY"
          aria-labelledby="tab-HISTORY"
        >
          <h2>
            <History size={20} /> Lịch sử đăng ký
          </h2>
          <div className="calendar-day-list">
            {history
              .slice((historyPage - 1) * 20, historyPage * 20)
              .map((h) => (
                <div key={h.key}>
                  <span>
                    <strong>{h.title}</strong>
                    <small>{h.reason}</small>
                  </span>
                  <span>
                    {
                      (
                        {
                          ACTIVE: "Đang có hiệu lực",
                          SUPERSEDED: "Đã thay thế",
                          WITHDRAWN: "Đã rút",
                        } as Record<string, string>
                      )[h.status]
                    }
                  </span>
                </div>
              ))}
          </div>
          {history.length === 0 && <p>Chưa công bố lịch.</p>}
          <div className="form-actions">
            <Button
              variant="secondary"
              disabled={historyPage === 1}
              onClick={() => setHistoryPage(historyPage - 1)}
            >
              Trước
            </Button>
            <span>Trang {historyPage}</span>
            <Button
              variant="secondary"
              disabled={historyPage * 20 >= history.length}
              onClick={() => setHistoryPage(historyPage + 1)}
            >
              Sau
            </Button>
          </div>
        </section>
      )}
      <Modal
        title="Khoảng giờ qua nửa đêm"
        description={`Tách thành hai ngày: ${weekdays[day]} ${clock(start * 60)}–24:00 và ${weekdays[(day + 1) % 7]} 00:00–${clock(end * 60)}.`}
        open={overnight}
        onClose={() => setOvernight(false)}
      >
        <div className="form-actions">
          <Button variant="secondary" onClick={() => setOvernight(false)}>
            Hủy
          </Button>
          <Button onClick={() => range(true)}>Xác nhận tách ngày</Button>
        </div>
      </Modal>
      <Modal
        title={exception ? "Chỉnh sửa ngoại lệ" : "Thêm ngoại lệ theo ngày"}
        description="Giờ đặc biệt thay thế cả ngày, gồm nhiều khoảng không liền nhau."
        open={exceptionOpen}
        onClose={() => setExceptionOpen(false)}
        busy={busy}
      >
        <Notice>{error}</Notice>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            preview({
              kind: "EXCEPTION",
              base_revision: room.calendar_revision,
              exception_id: exception?.id ?? 0,
              date: exDate,
              mode: exMode,
              closure_type: exType,
              reason: exReason,
              hours:
                exMode === "CLOSED"
                  ? []
                  : toHours(exWeek)
                      .filter((h) => h.weekday === 1)
                      .map((h) => ({ ...h, weekday: 0 })),
            });
          }}
        >
          <div className="form-grid">
            <Field label="Ngày" required>
              <Input
                type="date"
                min={localDate()}
                max={localDate(56)}
                required
                value={exDate}
                disabled={!!exception}
                onChange={(e) => setExDate(e.target.value)}
              />
            </Field>
            <Field label="Cách áp dụng">
              <Select
                value={exMode}
                onChange={(e) => setExMode(e.target.value)}
              >
                <option value="CLOSED">Nghỉ cả ngày</option>
                <option value="REPLACE">Thay giờ hoạt động</option>
              </Select>
            </Field>
            <Field label="Loại ngoại lệ">
              <Select
                value={exType}
                onChange={(e) => setExType(e.target.value)}
              >
                <option value="HOLIDAY">Ngày nghỉ</option>
                <option value="MAINTENANCE">Bảo trì</option>
                <option value="OTHER">Khác</option>
              </Select>
            </Field>
          </div>
          {exMode === "REPLACE" && (
            <>
              <div className="calendar-hour-pills">
                {toHours(exWeek).map((h) => (
                  <button
                    type="button"
                    className="calendar-hour-pill"
                    key={h.start_minute}
                    onClick={() =>
                      setExWeek(
                        paint(
                          exWeek,
                          [0, h.start_minute / 60],
                          [0, h.end_minute / 60 - 1],
                          false,
                        ),
                      )
                    }
                    aria-label={`Bỏ khoảng ${clock(h.start_minute)}–${clock(h.end_minute)}`}
                  >
                    {clock(h.start_minute)}–{clock(h.end_minute)} ×
                  </button>
                ))}
              </div>
              <div className="form-grid">
                <Field label="Từ giờ">
                  <Select
                    value={exStart}
                    onChange={(e) => setExStart(Number(e.target.value))}
                  >
                    {Array.from({ length: 24 }, (_, i) => (
                      <option key={i} value={i}>
                        {clock(i * 60)}
                      </option>
                    ))}
                  </Select>
                </Field>
                <Field label="Đến giờ">
                  <Select
                    value={exEnd}
                    onChange={(e) => setExEnd(Number(e.target.value))}
                  >
                    {Array.from({ length: 24 }, (_, i) => (
                      <option key={i} value={i + 1}>
                        {clock((i + 1) * 60)}
                      </option>
                    ))}
                  </Select>
                </Field>
              </div>
              <Button
                type="button"
                variant="secondary"
                disabled={exStart >= exEnd}
                onClick={() =>
                  setExWeek(paint(exWeek, [0, exStart], [0, exEnd - 1], true))
                }
              >
                Thêm khoảng giờ
              </Button>
              <p className="field-hint">
                Qua nửa đêm: tạo ngoại lệ riêng cho hai ngày.
              </p>
            </>
          )}
          <Field label="Lý do" required>
            <Input
              required
              maxLength={1000}
              value={exReason}
              onChange={(e) => setExReason(e.target.value)}
            />
          </Field>
          <div className="form-actions">
            <Button
              type="button"
              variant="secondary"
              onClick={() => setExceptionOpen(false)}
            >
              Hủy
            </Button>
            <Button
              type="submit"
              busy={busy}
              disabled={exMode === "REPLACE" && !exWeek[0].some(Boolean)}
            >
              Xem trước
            </Button>
          </div>
        </form>
      </Modal>
      <Modal
        title="Xác nhận thay đổi lịch"
        description="Kiểm tra tác động trước khi ghi vào database. Preview có hiệu lực 5 phút; dữ liệu được kiểm tra lại khi lưu."
        open={!!pending}
        onClose={() => setPending(null)}
        busy={busy}
      >
        <Notice>{previewError}</Notice>
        {pending && (
          <>
            <div className="calendar-preview-stats">
              <span>
                <strong>{pending.preview.creates}</strong> phiên mới
              </span>
              <span>
                <strong>
                  {
                    pending.preview.changes.filter((c) => c.action === "CLOSE")
                      .length
                  }
                </strong>{" "}
                phiên đóng
              </span>
              <span>
                <strong>
                  {
                    pending.preview.changes.filter((c) => c.action === "REOPEN")
                      .length
                  }
                </strong>{" "}
                phiên mở lại
              </span>
            </div>
            <p>Lý do: {pending.mutation.reason}</p>
            {pending.mutation.kind === "WEEKLY" && (
              <>
                <p>
                  Áp dụng từ {pending.mutation.effective_from}, lặp hằng tuần:
                </p>
                <div className="calendar-preview-list">
                  {weekdays.map((label, d) => (
                    <div key={label}>
                      <strong>{label}</strong>
                      <span>
                        {pending.mutation.hours
                          .filter((h) => h.weekday === d + 1)
                          .map(
                            (h) =>
                              `${clock(h.start_minute)}–${clock(h.end_minute)}`,
                          )
                          .join(", ") || "Đóng cả ngày"}
                      </span>
                    </div>
                  ))}
                </div>
              </>
            )}
            {pending.mutation.kind === "EXCEPTION" && (
              <p>
                Ngày {pending.mutation.date}:{" "}
                {pending.mutation.mode === "CLOSED"
                  ? "nghỉ cả ngày"
                  : pending.mutation.hours
                      .map(
                        (h) =>
                          `${clock(h.start_minute)}–${clock(h.end_minute)}`,
                      )
                      .join(", ")}
                .
              </p>
            )}
            <p>
              Chuẩn bị đến {pending.preview.prepared_through}. Phiên đã bắt đầu
              và phiên đóng thủ công giữ nguyên.
            </p>
            {pending.preview.conflicts.length > 0 && (
              <Notice>
                Có {pending.preview.conflicts.length} phiên chứa booking bị ảnh
                hưởng. Cần xử lý booking trước khi công bố.
              </Notice>
            )}
            <div className="calendar-preview-list">
              {pending.preview.changes.map((c) => (
                <div key={c.session_id}>
                  <span>
                    #{c.session_id} · {c.date}
                    <small>
                      {vietnamTime(c.starts_at)}
                      –
                      {vietnamTime(c.ends_at)}
                    </small>
                  </span>
                  <span>
                    {c.action === "CLOSE" ? "Đóng" : "Mở lại"} ·{" "}
                    {c.active_holds} chỗ giữ
                  </span>
                </div>
              ))}
            </div>
            <div className="form-actions">
              <Button
                variant="secondary"
                disabled={busy}
                onClick={() => setPending(null)}
              >
                Quay lại chỉnh
              </Button>
              {previewError && (
                <Button
                  variant="secondary"
                  disabled={busy}
                  onClick={() => {
                    setPending(null);
                    setDiscardReload(true);
                  }}
                >
                  Tải lại lịch
                </Button>
              )}
              <Button
                busy={busy}
                disabled={pending.preview.conflicts.length > 0}
                onClick={confirm}
              >
                Xác nhận đăng ký
              </Button>
            </div>
          </>
        )}
      </Modal>
      <Modal
        title="Tải lại lịch từ database?"
        description="Bản nháp hiện tại sẽ bị bỏ. Hãy sao lại các giờ cần giữ trước khi tiếp tục."
        open={discardReload}
        onClose={() => setDiscardReload(false)}
      >
        <div className="form-actions">
          <Button variant="secondary" onClick={() => setDiscardReload(false)}>
            Ở lại
          </Button>
          <Button
            onClick={() => {
              dirtyRef.current = false;
              reload();
            }}
          >
            Tải lại
          </Button>
        </div>
      </Modal>
      <Modal
        title="Bỏ lịch chưa lưu?"
        description="Các thay đổi trong bản nháp sẽ bị bỏ."
        open={blocker.state === "blocked"}
        onClose={() => blocker.reset?.()}
      >
        <div className="form-actions">
          <Button variant="secondary" onClick={() => blocker.reset?.()}>
            Ở lại
          </Button>
          <Button
            variant="danger"
            onClick={() => {
              dirtyRef.current = false;
              blocker.proceed?.();
            }}
          >
            Bỏ thay đổi
          </Button>
        </div>
      </Modal>
    </>
  );
}
