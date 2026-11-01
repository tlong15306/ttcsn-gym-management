import {
  useEffect,
  useMemo,
  useState,
  type FormEvent,
  type ReactNode,
} from "react";
import { useSearchParams } from "react-router-dom";
import {
  BarChart3,
  ChevronLeft,
  ChevronRight,
  RefreshCw,
  Star,
} from "lucide-react";
import { resource } from "../../api";
import { useAuth } from "../../app/session";
import { PageHeader } from "../../components/shared/list";
import { useRecord } from "../../components/shared/record";
import { Button, Field, Input, Notice, Select } from "../../components/ui";
import type { PageMeta } from "../../types";
import { localDate, vietnamDateTime } from "../../utils/vietnamTime";

type Rate = {
  numerator: number;
  denominator: number;
  rate_percent: number | null;
};
type SeatMinutes = {
  booked: number;
  capacity: number;
  rate_percent: number | null;
};
type OperationsReport = {
  from: string;
  to: string;
  timezone: string;
  as_of: string;
  total_bookings: number;
  status_counts: Record<string, number>;
  member_attendance_count: number;
  pt_attendance_count: number;
  no_show: Rate;
  pending_completion_count: number;
  utilization: SeatMinutes;
  cancellation_source_counts: Record<string, number>;
  unclassified_booking_count: number;
};
type TrainerReport = {
  from: string;
  to: string;
  timezone: string;
  as_of: string;
  pt_id: number;
  rating_average: number | null;
  review_count: number;
  attendance: Rate;
  excluded_count: number;
  unclassified_count: number;
  pending_completion_count: number;
};
type RoomOption = { id: number; name: string; status: string };
type PTOption = { id: number; user: { full_name: string }; specialty: string };
type PTProfile = { id: number; specialty: string; approval_status: string };
type Review = {
  rating: number;
  comment: string;
  created_at: string;
  reviewer_label: string;
};

const roles = [
  ["PENDING_PT", "Chờ PT phản hồi"],
  ["CONFIRMED", "Đã xác nhận"],
  ["REJECTED", "PT từ chối"],
  ["EXPIRED", "Hết hạn phản hồi"],
  ["CANCELLED", "Đã hủy"],
  ["COMPLETED", "Hoàn thành"],
  ["NO_SHOW", "Vắng mặt"],
] as const;
const cancellationLabels: Record<string, string> = {
  MEMBER: "Hội viên",
  PT: "Huấn luyện viên",
  FACILITY: "Cơ sở",
  UNKNOWN: "Chưa rõ nguồn",
};

function defaultRange() {
  const today = localDate();
  return { from: `${today.slice(0, 7)}-01`, to: today };
}

function reportPath(resourcePath: string, from: string, to: string) {
  const query = new URLSearchParams({ from, to });
  return `${resourcePath}?${query.toString()}`;
}

function formatDateTime(value: string) {
  return vietnamDateTime(value);
}

function formatRate(value: number | null) {
  return value == null ? "N/A" : `${value.toFixed(2)}%`;
}

function useReport<T>(path: string | null) {
  const { session } = useAuth();
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(!!path);
  const [error, setError] = useState("");
  const [revision, setRevision] = useState(0);
  useEffect(() => {
    if (!path) {
      setData(null);
      setLoading(false);
      setError("");
      return;
    }
    const abort = new AbortController();
    setData(null);
    setLoading(true);
    setError("");
    resource
      .get<T>(session.accessToken, path, abort.signal)
      .then((result) => {
        if (!abort.signal.aborted) setData(result.data);
      })
      .catch((cause) => {
        if (!abort.signal.aborted) setError((cause as Error).message);
      })
      .finally(() => {
        if (!abort.signal.aborted) setLoading(false);
      });
    return () => abort.abort();
  }, [path, revision, session.accessToken]);
  return { data, loading, error, retry: () => setRevision((v) => v + 1) };
}

function validatePeriod(from: string, to: string) {
  if (!from || !to || from > to)
    return "Chọn khoảng ngày hợp lệ, từ ngày không sau đến ngày.";
  const days =
    (Date.parse(`${to}T00:00:00Z`) - Date.parse(`${from}T00:00:00Z`)) /
    86_400_000;
  if (!Number.isFinite(days) || days > 365)
    return "Khoảng báo cáo tối đa là 366 ngày.";
  return "";
}

function PeriodFields({
  from,
  to,
  setFrom,
  setTo,
}: {
  from: string;
  to: string;
  setFrom: (value: string) => void;
  setTo: (value: string) => void;
}) {
  return (
    <>
      <Field label="Từ ngày">
        <Input
          aria-label="Từ ngày"
          type="date"
          value={from}
          onChange={(event) => setFrom(event.target.value)}
          required
        />
      </Field>
      <Field label="Đến ngày">
        <Input
          aria-label="Đến ngày"
          type="date"
          value={to}
          onChange={(event) => setTo(event.target.value)}
          required
        />
      </Field>
    </>
  );
}

function useRangeParams() {
  const defaults = useMemo(defaultRange, []);
  const [params, setParams] = useSearchParams();
  const [from, setFrom] = useState(params.get("from") || defaults.from);
  const [to, setTo] = useState(params.get("to") || defaults.to);
  useEffect(() => {
    setFrom(params.get("from") || defaults.from);
    setTo(params.get("to") || defaults.to);
  }, [params, defaults.from, defaults.to]);
  function applyRange(
    event: FormEvent<HTMLFormElement>,
    extra: Record<string, string> = {},
  ) {
    event.preventDefault();
    const error = validatePeriod(from, to);
    if (error) return error;
    setParams((current) => {
      const next = new URLSearchParams(current);
      next.set("from", from);
      next.set("to", to);
      for (const [key, value] of Object.entries(extra)) {
        if (value) next.set(key, value);
        else next.delete(key);
      }
      return next;
    });
    return "";
  }
  function setParam(key: string, value: string) {
    setParams((current) => {
      const next = new URLSearchParams(current);
      if (value) next.set(key, value);
      else next.delete(key);
      return next;
    });
  }
  return { params, from, to, setFrom, setTo, applyRange, setParam };
}

function PeriodCaption({
  report,
}: {
  report: { from: string; to: string; as_of: string };
}) {
  return (
    <p className="report-period-caption">
      {report.from} → {report.to} · Cập nhật {formatDateTime(report.as_of)}
    </p>
  );
}

function Metric({
  label,
  value,
  detail,
  icon,
}: {
  label: string;
  value: string;
  detail: string;
  icon?: ReactNode;
}) {
  return (
    <article className="report-metric">
      <div className="report-metric-top">
        <span>{label}</span>
        {icon}
      </div>
      <strong>{value}</strong>
      <small>{detail}</small>
    </article>
  );
}

function ReportState<T>({
  state,
  children,
}: {
  state: { data: T | null; loading: boolean; error: string; retry: () => void };
  children: (value: T) => ReactNode;
}) {
  if (state.loading)
    return (
      <div className="report-loading" role="status">
        Đang tải báo cáo…
      </div>
    );
  if (state.error)
    return (
      <div className="report-state">
        <Notice>{state.error}</Notice>
        <Button variant="secondary" onClick={state.retry}>
          <RefreshCw size={16} /> Thử tải lại
        </Button>
      </div>
    );
  if (!state.data)
    return (
      <div className="report-state report-empty">
        <BarChart3 size={22} />
        <p>Chưa có dữ liệu trong khoảng thời gian này.</p>
      </div>
    );
  return <>{children(state.data)}</>;
}

function StatusBreakdown({ report }: { report: OperationsReport }) {
  const rows = roles
    .map(([key, label]) => ({
      key,
      label,
      value: report.status_counts[key] ?? 0,
    }))
    .sort(
      (a, b) =>
        b.value - a.value ||
        roles.findIndex(([key]) => key === a.key) -
          roles.findIndex(([key]) => key === b.key),
    );
  return (
    <section
      className="panel report-section"
      aria-labelledby="report-status-title"
    >
      <div className="report-section-heading">
        <div>
          <h2 id="report-status-title">Trạng thái đặt lịch</h2>
          <p>Trạng thái hiện tại của các booking trong kỳ.</p>
        </div>
        <strong>{report.total_bookings} booking</strong>
      </div>
      {report.total_bookings === 0 ? (
        <div className="report-inline-empty">
          Chưa có booking trong khoảng ngày đã chọn.
        </div>
      ) : (
        <div className="report-status-list">
          {rows.map((row) => (
            <div className="report-status-row" key={row.key}>
              <span>{row.label}</span>
              <progress
                max={report.total_bookings}
                value={row.value}
                aria-label={`${row.label}: ${row.value} booking`}
              />
              <strong>{row.value}</strong>
            </div>
          ))}
        </div>
      )}
    </section>
  );
}

function OperationsMetrics({ report }: { report: OperationsReport }) {
  const cancellations = Object.entries(report.cancellation_source_counts).sort(
    ([a], [b]) => a.localeCompare(b),
  );
  return (
    <>
      <PeriodCaption report={report} />
      <section className="report-metrics" aria-label="Chỉ số vận hành">
        <Metric
          label="Tổng booking"
          value={String(report.total_bookings)}
          detail="Theo trạng thái hiện tại"
          icon={<BarChart3 size={17} aria-hidden="true" />}
        />
        <Metric
          label="Tỷ lệ vắng"
          value={formatRate(report.no_show.rate_percent)}
          detail={`${report.no_show.numerator} / ${report.no_show.denominator} buổi đã hoàn tất`}
        />
        <Metric
          label="Sử dụng chỗ đặt"
          value={formatRate(report.utilization.rate_percent)}
          detail={`${report.utilization.booked} / ${report.utilization.capacity} seat-phút`}
        />
        <Metric
          label="Điểm danh"
          value={`${report.member_attendance_count} HV · ${report.pt_attendance_count} PT`}
          detail="Số booking có mặt, tính riêng từng người"
        />
      </section>
      <StatusBreakdown report={report} />
      <section className="report-detail-grid">
        <article className="panel report-section">
          <div className="report-section-heading">
            <div>
              <h2>Đang chờ xử lý</h2>
              <p>Buổi đã hết giờ nhưng job chưa đổi trạng thái.</p>
            </div>
          </div>
          <div className="report-detail-number">
            {report.pending_completion_count}
          </div>
          <small>booking đang chờ hoàn tất</small>
        </article>
        <article className="panel report-section">
          <div className="report-section-heading">
            <div>
              <h2>Hủy theo nguồn</h2>
              <p>Nguồn được lưu tại thời điểm hủy.</p>
            </div>
          </div>
          {cancellations.length ? (
            <dl className="report-source-list">
              {cancellations.map(([key, amount]) => (
                <div key={key}>
                  <dt>{cancellationLabels[key] ?? key}</dt>
                  <dd>{amount}</dd>
                </div>
              ))}
            </dl>
          ) : (
            <p className="report-muted">Không có booking bị hủy trong kỳ.</p>
          )}
        </article>
      </section>
      {report.unclassified_booking_count > 0 && (
        <p className="report-data-note">
          <strong>{report.unclassified_booking_count}</strong> booking thiếu
          bằng chứng xác nhận hoặc nguồn hủy; hệ thống tách riêng thay vì tự suy
          đoán.
        </p>
      )}
      <p className="report-method-note">
        Tỷ lệ vắng tính trên COMPLETED + NO_SHOW đã kết thúc. Mức sử dụng tính
        seat-minute đã đặt trên phiên OPEN đã kết thúc; đây không phải thời gian
        tập được đo.
      </p>
    </>
  );
}

function TrainerMetrics({ report }: { report: TrainerReport }) {
  return (
    <>
      <PeriodCaption report={report} />
      <section
        className="report-metrics report-metrics-two"
        aria-label="Hiệu suất huấn luyện viên"
      >
        <Metric
          label="Đánh giá trung bình"
          value={
            report.rating_average == null
              ? "N/A"
              : `${report.rating_average.toFixed(1)} / 5`
          }
          detail={`${report.review_count} đánh giá hợp lệ trong kỳ`}
          icon={<Star size={17} aria-hidden="true" />}
        />
        <Metric
          label="Chuyên cần PT"
          value={formatRate(report.attendance.rate_percent)}
          detail={`${report.attendance.numerator} / ${report.attendance.denominator} buổi đủ điều kiện`}
        />
      </section>
      <section className="report-detail-grid">
        <article className="panel report-section">
          <div className="report-section-heading">
            <div>
              <h2>Phân loại buổi tập</h2>
              <p>Để đối chiếu tỷ lệ chuyên cần.</p>
            </div>
          </div>
          <dl className="report-source-list">
            <div>
              <dt>Đủ điều kiện, PT có mặt</dt>
              <dd>{report.attendance.numerator}</dd>
            </div>
            <div>
              <dt>Mẫu số theo BR-22</dt>
              <dd>{report.attendance.denominator}</dd>
            </div>
            <div>
              <dt>Bị loại khỏi mẫu số</dt>
              <dd>{report.excluded_count}</dd>
            </div>
            <div>
              <dt>Thiếu chứng cứ / UNKNOWN</dt>
              <dd>{report.unclassified_count}</dd>
            </div>
            <div>
              <dt>Chờ job hoàn tất</dt>
              <dd>{report.pending_completion_count}</dd>
            </div>
          </dl>
        </article>
        <article className="panel report-section report-explainer">
          <h2>Cách tính</h2>
          <p>
            Chuyên cần dùng buổi COMPLETED/NO_SHOW đã xác nhận và buổi Hội
            viên/PT hủy sau xác nhận. Hủy bởi cơ sở, từ chối, hết hạn và yêu cầu
            đang chờ được loại.
          </p>
          <p>
            Review chỉ tính khi booking hoàn thành, đã kết thúc và cả Hội viên
            lẫn PT có attendance. Không có mẫu số hoặc review sẽ hiện N/A.
          </p>
        </article>
      </section>
    </>
  );
}

function DateFilterError({ children }: { children: string }) {
  return children ? (
    <p className="report-filter-error" role="alert">
      {children}
    </p>
  ) : null;
}

export function AdminReportsPage() {
  const { session } = useAuth();
  const range = useRangeParams();
  const selectedRoom = range.params.get("room_id") || "";
  const [roomDraft, setRoomDraft] = useState(selectedRoom);
  const [filterError, setFilterError] = useState("");
  const [rooms, setRooms] = useState<RoomOption[]>([]);
  const [roomsLoading, setRoomsLoading] = useState(true);
  const [roomsError, setRoomsError] = useState("");
  const [roomsRevision, setRoomsRevision] = useState(0);
  const [ptQuery, setPTQuery] = useState("");
  const [pts, setPTs] = useState<PTOption[]>([]);
  const [ptsLoading, setPTsLoading] = useState(false);
  const [ptsError, setPTsError] = useState("");
  const selectedPT = range.params.get("pt_id") || "";
  useEffect(() => setRoomDraft(selectedRoom), [selectedRoom]);
  useEffect(() => {
    const abort = new AbortController();
    setRoomsLoading(true);
    setRoomsError("");
    resource
      .list<RoomOption>(
        session.accessToken,
        "rooms",
        new URLSearchParams({ page: "1", page_size: "100" }),
        abort.signal,
      )
      .then((result) => {
        if (!abort.signal.aborted) setRooms(result.data);
      })
      .catch((cause) => {
        if (!abort.signal.aborted) setRoomsError((cause as Error).message);
      })
      .finally(() => {
        if (!abort.signal.aborted) setRoomsLoading(false);
      });
    return () => abort.abort();
  }, [session.accessToken, roomsRevision]);
  useEffect(() => {
    const abort = new AbortController();
    const timer = setTimeout(() => {
      setPTsLoading(true);
      setPTsError("");
      resource
        .list<PTOption>(
          session.accessToken,
          "pts",
          new URLSearchParams({ q: ptQuery, page: "1", page_size: "20" }),
          abort.signal,
        )
        .then((result) => {
          if (!abort.signal.aborted) setPTs(result.data);
        })
        .catch((cause) => {
          if (!abort.signal.aborted) setPTsError((cause as Error).message);
        })
        .finally(() => {
          if (!abort.signal.aborted) setPTsLoading(false);
        });
    }, 200);
    return () => {
      clearTimeout(timer);
      abort.abort();
    };
  }, [session.accessToken, ptQuery]);
  const operationsPath =
    reportPath(
      "reports/operations",
      range.params.get("from") || range.from,
      range.params.get("to") || range.to,
    ) + (selectedRoom ? `&room_id=${encodeURIComponent(selectedRoom)}` : "");
  const operations = useReport<OperationsReport>(operationsPath);
  const ptReportPath = selectedPT
    ? `${reportPath(`reports/pts/${encodeURIComponent(selectedPT)}`, range.params.get("from") || range.from, range.params.get("to") || range.to)}`
    : null;
  const ptReport = useReport<TrainerReport>(ptReportPath);

  function apply(event: FormEvent<HTMLFormElement>) {
    const error = range.applyRange(event, {
      room_id: roomDraft,
      pt_id: selectedPT,
    });
    setFilterError(error);
  }
  function choosePT(value: string) {
    range.setParam("pt_id", value);
  }

  return (
    <div className="report-page">
      <PageHeader
        title="Báo cáo vận hành"
        description="Số liệu theo ngày buổi tập tại Việt Nam, có công thức và thời điểm chụp rõ ràng."
      />
      <form className="panel report-filter-panel" onSubmit={apply}>
        <div className="report-filter-fields">
          <PeriodFields
            from={range.from}
            to={range.to}
            setFrom={range.setFrom}
            setTo={range.setTo}
          />
          <Field label="Phòng tập">
            <Select
              aria-label="Lọc phòng tập"
              value={roomDraft}
              disabled={roomsLoading}
              onChange={(event) => setRoomDraft(event.target.value)}
            >
              <option value="">Toàn bộ phòng</option>
              {rooms.map((room) => (
                <option value={String(room.id)} key={room.id}>
                  {room.name}
                </option>
              ))}
            </Select>
          </Field>
          <Button type="submit">Áp dụng</Button>
        </div>
        <DateFilterError>{filterError}</DateFilterError>
        {roomsError && (
          <div className="report-filter-feedback">
            <Notice>{roomsError}</Notice>
            <Button
              variant="secondary"
              onClick={() => setRoomsRevision((n) => n + 1)}
            >
              Tải lại danh sách phòng
            </Button>
          </div>
        )}
      </form>
      <ReportState state={operations}>
        {(report) => <OperationsMetrics report={report} />}
      </ReportState>
      <section
        className="panel report-pt-lookup"
        aria-labelledby="pt-report-title"
      >
        <div className="report-section-heading">
          <div>
            <h2 id="pt-report-title">Tra cứu hiệu suất PT</h2>
            <p>Dùng cùng kỳ ngày đã chọn ở trên.</p>
          </div>
        </div>
        <div className="report-filter-fields report-pt-fields">
          <Field label="Tìm huấn luyện viên">
            <Input
              value={ptQuery}
              onChange={(event) => setPTQuery(event.target.value)}
              placeholder="Nhập tên hoặc chuyên môn"
            />
          </Field>
          <Field label="Huấn luyện viên">
            <Select
              aria-label="Chọn huấn luyện viên"
              value={selectedPT}
              onChange={(event) => choosePT(event.target.value)}
              disabled={ptsLoading}
            >
              <option value="">Chọn một PT</option>
              {pts.map((pt) => (
                <option value={pt.id} key={pt.id}>
                  {pt.user.full_name}
                  {pt.specialty ? ` · ${pt.specialty}` : ""}
                </option>
              ))}
            </Select>
          </Field>
        </div>
        {ptsLoading && (
          <p role="status" className="report-muted">
            Đang tìm huấn luyện viên…
          </p>
        )}
        {ptsError && <Notice>{ptsError}</Notice>}
        {selectedPT && (
          <ReportState state={ptReport}>
            {(report) => <TrainerMetrics report={report} />}
          </ReportState>
        )}
      </section>
      <p className="report-method-note">
        Tổng booking là trạng thái hiện tại tại as_of; số buổi tương lai trong
        kỳ có thể xuất hiện. Tỷ lệ chỉ tính buổi đã kết thúc. N/A nghĩa là mẫu
        số bằng 0 hoặc chưa có review.
      </p>
    </div>
  );
}

export function CoachPerformancePage() {
  const profile = useRecord<PTProfile>("pts/me");
  const range = useRangeParams();
  const [filterError, setFilterError] = useState("");
  const path = profile.data?.id
    ? reportPath(
        `reports/pts/${profile.data.id}`,
        range.params.get("from") || range.from,
        range.params.get("to") || range.to,
      )
    : null;
  const report = useReport<TrainerReport>(path);
  function apply(event: FormEvent<HTMLFormElement>) {
    setFilterError(range.applyRange(event));
  }
  return (
    <div className="report-page coach-report-page">
      <PageHeader
        title="Hiệu suất của tôi"
        description="Đánh giá và chuyên cần theo kỳ, kèm số buổi đủ điều kiện để tự đối chiếu."
      />
      <form className="panel report-filter-panel" onSubmit={apply}>
        <div className="report-filter-fields">
          <PeriodFields
            from={range.from}
            to={range.to}
            setFrom={range.setFrom}
            setTo={range.setTo}
          />
          <Button type="submit">Áp dụng</Button>
        </div>
        <DateFilterError>{filterError}</DateFilterError>
      </form>
      {profile.loading && (
        <div className="report-loading" role="status">
          Đang tải hồ sơ PT…
        </div>
      )}
      {profile.error && (
        <div className="report-state">
          <Notice>{profile.error}</Notice>
          <Button variant="secondary" onClick={profile.refresh}>
            Thử lại
          </Button>
        </div>
      )}
      {profile.data && (
        <ReportState state={report}>
          {(data) => <TrainerMetrics report={data} />}
        </ReportState>
      )}
    </div>
  );
}

function ReviewEntry({ review }: { review: Review }) {
  return (
    <article className="pt-review-entry">
      <div className="pt-review-entry-head">
        <span
          className="pt-review-stars"
          aria-label={`${review.rating} trên 5 sao`}
        >
          <span aria-hidden="true">
            {"★".repeat(review.rating)}
            {"☆".repeat(5 - review.rating)}
          </span>
          <strong>{review.rating}/5</strong>
        </span>
        <time dateTime={review.created_at}>
          {formatDateTime(review.created_at)}
        </time>
      </div>
      <strong className="pt-review-author">
        {review.reviewer_label || "Hội viên"}
      </strong>
      {review.comment ? (
        <p>{review.comment}</p>
      ) : (
        <p className="report-muted">Chỉ chấm sao, không để lại nhận xét.</p>
      )}
    </article>
  );
}

export function CoachReviewsPage() {
  const { session } = useAuth();
  const profile = useRecord<PTProfile>("pts/me");
  const [page, setPage] = useState(1);
  const [rows, setRows] = useState<Review[]>([]);
  const [meta, setMeta] = useState<PageMeta>({
    page: 1,
    page_size: 20,
    total: 0,
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [revision, setRevision] = useState(0);
  useEffect(() => {
    if (!profile.data?.id) return;
    const abort = new AbortController();
    setLoading(true);
    setError("");
    setRows([]);
    const query = new URLSearchParams({ page: String(page), page_size: "20" });
    resource
      .list<Review>(
        session.accessToken,
        `pts/${profile.data.id}/reviews`,
        query,
        abort.signal,
      )
      .then((result) => {
        if (!abort.signal.aborted) {
          setRows(result.data);
          if (result.meta) setMeta(result.meta);
        }
      })
      .catch((cause) => {
        if (!abort.signal.aborted) setError((cause as Error).message);
      })
      .finally(() => {
        if (!abort.signal.aborted) setLoading(false);
      });
    return () => abort.abort();
  }, [profile.data?.id, page, revision, session.accessToken]);
  const pages = Math.max(1, Math.ceil(meta.total / meta.page_size));
  return (
    <div className="report-page coach-report-page">
      <PageHeader
        title="Đánh giá của hội viên"
        description="Phản hồi đã gửi về các buổi tập của bạn. Thông tin hội viên được ẩn danh."
      />
      {profile.loading && (
        <div className="report-loading" role="status">
          Đang tải hồ sơ PT…
        </div>
      )}
      {profile.error && (
        <div className="report-state">
          <Notice>{profile.error}</Notice>
          <Button variant="secondary" onClick={profile.refresh}>
            Thử lại
          </Button>
        </div>
      )}
      {profile.data && (
        <section
          className="panel pt-review-list"
          aria-label="Danh sách đánh giá"
        >
          {loading ? (
            <div className="report-loading" role="status">
              Đang tải đánh giá…
            </div>
          ) : error ? (
            <div className="report-state">
              <Notice>{error}</Notice>
              <Button
                variant="secondary"
                onClick={() => setRevision((n) => n + 1)}
              >
                Thử lại
              </Button>
            </div>
          ) : rows.length === 0 ? (
            <div className="report-state report-empty">
              <Star size={22} />
              <p>Chưa có đánh giá trong hồ sơ của bạn.</p>
            </div>
          ) : (
            <div className="pt-review-list-items">
              {rows.map((review, index) => (
                <ReviewEntry
                  key={`${review.created_at}-${index}`}
                  review={review}
                />
              ))}
            </div>
          )}
          {!error && meta.total > 0 && (
            <nav className="report-pagination" aria-label="Phân trang đánh giá">
              <Button
                variant="secondary"
                disabled={loading || page <= 1}
                onClick={() => setPage((n) => Math.max(1, n - 1))}
              >
                <ChevronLeft size={17} /> Trang trước
              </Button>
              <span>
                Trang {page} / {pages} · {meta.total} đánh giá
              </span>
              <Button
                variant="secondary"
                disabled={loading || page >= pages}
                onClick={() => setPage((n) => Math.min(pages, n + 1))}
              >
                Trang sau <ChevronRight size={17} />
              </Button>
            </nav>
          )}
        </section>
      )}
    </div>
  );
}
