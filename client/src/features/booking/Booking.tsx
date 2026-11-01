import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  Link,
  useNavigate,
  useParams,
  useSearchParams,
  useLocation,
} from "react-router-dom";
import { useAuth } from "../../app/session";
import { ApiError, request, resource } from "../../api";
import {
  Button,
  Field,
  Input,
  Modal,
  Notice,
  Select,
} from "../../components/ui";
import { Pagination, SearchBox, useList } from "../../components/shared/list";
import { RecordState, useRecord } from "../../components/shared/record";
import type { Room, RoomSession } from "../facilities/Facilities";
import type { PageMeta } from "../../types";
import { RoomImage } from "./RoomImage";
import { AttendancePanel } from "./Attendance";
import { DateStrip } from "./DateStrip";
import { RatingSummary, ReviewPanel } from "./Review";
import { SessionRangePicker } from "./SessionRangePicker";
import {
  localDate,
  vietnamDay as day,
  vietnamTime as time,
} from "./vietnamTime";
import { civilDate, isCivilDate } from "../../utils/vietnamTime";
import { WeekCalendar } from "./WeekCalendar";
import { fetchDay, useWeeklyOptions } from "./useWeeklyOptions";
import { addDays, sessionDay, validateRange, weekStart } from "./weekly";
import { clearDraft, readDraft, restoreDraft, saveDraft } from "./draft";

export { localDate } from "./vietnamTime";

export type Booking = {
  has_attendance?: boolean;
  id: number;
  room_id: number;
  room_name: string;
  member_name: string;
  pt_name: string;
  pt_id: number;
  status: string;
  starts_at: string;
  ends_at: string;
  respond_by?: string;
  server_time: string;
  cancellation_reason: string;
};
type Option = RoomSession & { remaining: number; unavailable_reason: string };
type PT = {
  id: number;
  full_name: string;
  specialty: string;
  rating_average: number | null;
  review_count: number;
};
const labels: Record<string, string> = {
  PENDING_PT: "Chờ PT phản hồi",
  CONFIRMED: "Đã xác nhận",
  CANCELLED: "Đã hủy",
  REJECTED: "PT từ chối",
  EXPIRED: "Hết thời gian phản hồi",
  COMPLETED: "Hoàn thành",
  NO_SHOW: "Vắng mặt",
};
function useOptions<T>(path: string | null, query: string, revision = 0) {
  const { session } = useAuth();
  const [data, setData] = useState<T[]>([]);
  const [meta, setMeta] = useState<PageMeta>({
    page: 1,
    page_size: 20,
    total: 0,
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  useEffect(() => {
    const abort = new AbortController();
    setData([]);
    setError("");
    if (!path) {
      setLoading(false);
      return;
    }
    setLoading(true);
    resource
      .list<T>(
        session.accessToken,
        path,
        new URLSearchParams(query),
        abort.signal,
      )
      .then((r) => {
        if (!abort.signal.aborted) {
          setData(r.data);
          setMeta(r.meta!);
        }
      })
      .catch((e) => {
        if (!abort.signal.aborted) setError(e.message);
      })
      .finally(() => {
        if (!abort.signal.aborted) setLoading(false);
      });
    return () => abort.abort();
  }, [path, query, revision, session.accessToken]);
  return { data, meta, loading, error };
}
function useCommonPTs(
  sessions: Option[],
  page: number,
  pageSize: number,
  revision = 0,
) {
  const { session } = useAuth();
  const [data, setData] = useState<PT[]>([]);
  const [meta, setMeta] = useState<PageMeta>({
    page: 1,
    page_size: pageSize,
    total: 0,
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const sessionIDs = sessions
    .map((item) => item.id)
    .sort((a, b) => a - b)
    .join(",");
  const [forIDs, setForIDs] = useState("");
  const [allPTs, setAllPTs] = useState<PT[]>([]);
  useEffect(() => {
    const abort = new AbortController();
    setError("");
    setData([]);
    setForIDs("");
    setAllPTs([]);
    if (!sessions.length) {
      setData([]);
      setMeta({ page: 1, page_size: pageSize, total: 0 });
      setLoading(false);
      return;
    }
    setLoading(true);
    async function loadAllPTs(item: Option): Promise<PT[]> {
      const all: PT[] = [];
      let currentPage = 1;
      for (;;) {
        const response = await resource.list<PT>(
          session.accessToken,
          "booking-options/pts",
          new URLSearchParams({
            room_session_id: String(item.id),
            page: String(currentPage),
            page_size: "100",
          }),
          abort.signal,
        );
        all.push(...response.data);
        if (
          !response.data.length ||
          !response.meta ||
          currentPage * response.meta.page_size >= response.meta.total
        )
          return all;
        currentPage += 1;
      }
    }
    Promise.all(sessions.map(loadAllPTs))
      .then((responses) => {
        if (abort.signal.aborted) return;
        const shared = (responses[0] || []).filter((candidate) =>
          responses.every((response) =>
            response.some((item) => item.id === candidate.id),
          ),
        );
        const offset = (page - 1) * pageSize;
        setData(shared.slice(offset, offset + pageSize));
        setForIDs(sessionIDs);
        setAllPTs(shared);
        setMeta({ page, page_size: pageSize, total: shared.length });
      })
      .catch((e) => {
        if (!abort.signal.aborted) setError((e as Error).message);
      })
      .finally(() => {
        if (!abort.signal.aborted) setLoading(false);
      });
    return () => abort.abort();
  }, [page, pageSize, revision, session.accessToken, sessionIDs]);
  return { data, meta, loading, error, forIDs, allPTs };
}
export function RoomCatalog({ operations = false }: { operations?: boolean }) {
  const list = useList<Room>("rooms");
  const memberID = list.params.get("member_id") || "";
  return (
    <>
      <div className="portal-kicker">KHÔNG GIAN CỦA BẠN</div>
      <div className="portal-heading">
        <h1>
          Tìm nơi cho
          <br />
          buổi tập tiếp theo.
        </h1>
        <p>Khám phá phòng tập, xem giờ còn chỗ và chọn lịch phù hợp với bạn.</p>
      </div>
      <div className="catalog-toolbar">
        <SearchBox
          value={list.params.get("q") || ""}
          onChange={(v) => list.filter("q", v)}
          placeholder="Tìm phòng tập…"
        />
        <Select
          aria-label="Trạng thái phòng"
          value={list.params.get("status") || ""}
          onChange={(e) => list.filter("status", e.target.value)}
        >
          <option value="">Tất cả phòng</option>
          <option value="AVAILABLE">Sẵn sàng</option>
          <option value="MAINTENANCE">Bảo trì</option>
        </Select>
      </div>
      <Notice>{list.error}</Notice>
      {list.loading && <p role="status">Đang tải phòng tập…</p>}
      {!list.loading && !list.data.length && (
        <div className="portal-empty">
          Chưa có phòng phù hợp. Thử đổi từ khóa hoặc quay lại sau.
        </div>
      )}
      <div className="room-grid">
        {list.data.map((r) => (
          <Link
            key={r.id}
            className="room-card"
            to={
              operations
                ? `/management/bookings/new/rooms/${r.id}?member_id=${memberID}`
                : `/member/rooms/${r.id}`
            }
          >
            <RoomImage room={r} />
            <div className="room-card-copy">
              <span className="portal-kicker">{r.room_type}</span>
              <h2>
                {r.name}
                <span aria-hidden="true">↗</span>
              </h2>
              <p>{r.description || "Khám phá không gian và các ca tập."}</p>
              <div className="room-card-footer">
                <span>{r.default_capacity} chỗ / ca</span>
                <span>
                  {r.status === "AVAILABLE"
                    ? "Xem lịch trống"
                    : "Tạm ngừng đặt"}
                </span>
              </div>
            </div>
          </Link>
        ))}
      </div>
      <Pagination meta={list.meta} change={list.filter} />
    </>
  );
}
export function RoomBooking({ operations = false }: { operations?: boolean }) {
  const { id } = useParams();
  const [params, setParams] = useSearchParams();
  const memberID = Number(params.get("member_id"));
  const { session } = useAuth();
  const navigate = useNavigate();
  const room = useRecord<Room>(`rooms/${id}`);
  const requestedDate = params.get("date");
  const savedDraft = useMemo(() => readDraft(), [id, session.user.id]);
  const draftDate =
    !operations &&
    savedDraft?.roomID === Number(id) &&
    (savedDraft.owner === null || savedDraft.owner === session.user.id)
      ? savedDraft.date
      : null;
  const initialDate =
    isCivilDate(requestedDate) && requestedDate >= localDate()
      ? requestedDate
      : draftDate && draftDate >= localDate()
        ? draftDate
        : localDate();
  const [date, setDate] = useState(initialDate);
  const [page, setPage] = useState(1);
  const [ptPage, setPtPage] = useState(1);
  const [pageSize, setPageSize] = useState(100);
  const [ptPageSize, setPtPageSize] = useState(100);
  const [selected, setSelected] = useState<Option[]>([]);
  const [pt, setPT] = useState<PT>();
  const [revision, setRevision] = useState(0);
  const dailyOptions = useOptions<Option>(
    operations ? "booking-options/sessions" : null,
    `room_id=${id}&date=${date}&page=${page}&page_size=${pageSize}`,
    revision,
  );
  const weeklyOptions = useWeeklyOptions<Option>(
    session.accessToken,
    id!,
    date,
    revision,
    !operations,
  );
  const options = operations ? dailyOptions : weeklyOptions;
  const pts = useCommonPTs(selected, ptPage, ptPageSize, revision);
  const [review, setReview] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [rangeError, setRangeError] = useState("");
  const [draftNotice, setDraftNotice] = useState("");
  const [smallScreen, setSmallScreen] = useState(
    () =>
      typeof matchMedia !== "undefined" &&
      matchMedia("(max-width:760px)").matches,
  );
  useEffect(() => {
    if (typeof matchMedia === "undefined") return;
    const media = matchMedia("(max-width:760px)");
    const update = () => setSmallScreen(media.matches);
    media.addEventListener("change", update);
    return () => media.removeEventListener("change", update);
  }, []);
  const [needsPTChoice, setNeedsPTChoice] = useState(false);
  const restored = useRef("");
  const restorePT = useRef(0);
  const key = useRef("");
  const submitting = useRef(false);
  useEffect(() => {
    setSelected([]);
    setPT(undefined);
    key.current = "";
    setError("");
    setRangeError("");
    setDate(initialDate);
    setPage(1);
    setReview(false);
    restored.current = "";
    restorePT.current = 0;
    setNeedsPTChoice(false);
  }, [id, memberID, session.user.id]);
  useEffect(() => {
    if (
      isCivilDate(requestedDate) &&
      requestedDate >= localDate() &&
      requestedDate !== date
    ) {
      setSelected([]);
      setPT(undefined);
      key.current = "";
      setDate(requestedDate);
      setReview(false);
      restorePT.current = 0;
      if (!operations) {
        clearDraft(Number(id));
        setDraftNotice(
          "Đã chuyển ngày theo đường dẫn. Hãy chọn lại giờ trong lịch đang xem.",
        );
      }
    }
  }, [requestedDate, id]);
  function syncDate(next: string) {
    setDate(next);
    if (!operations && params.get("date") !== next) {
      const query = new URLSearchParams(params);
      query.set("date", next);
      setParams(query, { replace: true });
    }
  }
  useEffect(() => {
    if (
      operations ||
      options.loading ||
      options.error ||
      !weeklyOptions.loadedKey
    )
      return;
    const marker = `${id}:${session.user.id}`;
    if (restored.current !== marker) {
      restored.current = marker;
      const draft = readDraft();
      if (draft?.roomID === Number(id)) {
        if (draft.date !== date) {
          clearDraft(Number(id));
          setDraftNotice(
            "Bản nháp ở ngày khác đã được bỏ để giữ ngày đang xem.",
          );
          return;
        }
        const result = restoreDraft(
          draft,
          Number(id),
          session.user.id,
          options.data,
        );
        if (result.error) {
          clearDraft(Number(id));
          setDraftNotice(result.error);
        } else if (result.values.length) {
          setSelected(result.values);
          syncDate(draft.date);
          restorePT.current = draft.ptID;
          setNeedsPTChoice(draft.wantsPT || draft.ptID > 0);
          setDraftNotice(
            `Đã khôi phục ${result.values.length} ca và kiểm tra lại điều kiện. ${draft.wantsPT ? "Hãy xác minh lại HLV cho cả khoảng giờ." : "Chưa giữ chỗ; kiểm tra lại trước xác nhận."}`,
          );
          saveDraft(
            Number(id),
            result.values,
            session.user.id,
            draft.ptID,
            draft.wantsPT,
          );
        }
      }
    } else if (selected.length) {
      const refreshed = selected
        .map((old) =>
          options.data.find(
            (s) =>
              s.id === old.id &&
              s.starts_at === old.starts_at &&
              s.ends_at === old.ends_at,
          ),
        )
        .filter((s): s is Option => !!s);
      if (refreshed.length !== selected.length || validateRange(refreshed)) {
        choose();
        clearDraft(Number(id));
        setReview(false);
        setDraftNotice(
          "Lịch hoặc điều kiện đặt đã thay đổi. Hãy chọn lại giờ còn phù hợp.",
        );
      }
    }
  }, [
    weeklyOptions.loadedKey,
    options.loading,
    options.error,
    id,
    session.user.id,
  ]);
  useEffect(() => {
    if (
      !restorePT.current ||
      pts.loading ||
      !selected.length ||
      pts.forIDs !==
        selected
          .map((s) => s.id)
          .sort((a, b) => a - b)
          .join(",")
    )
      return;
    const restoredPT = pts.allPTs.find((p) => p.id === restorePT.current);
    restorePT.current = 0;
    if (restoredPT) {
      setPT(restoredPT);
      setNeedsPTChoice(false);
    } else
      setDraftNotice(
        "Đã khôi phục giờ tập. HLV trước đó không còn trong danh sách phù hợp; hãy chọn lại hình thức tập.",
      );
  }, [pts.loading, pts.allPTs, pts.forIDs, selected.length]);
  useEffect(() => {
    if (
      pt &&
      !pts.loading &&
      pts.forIDs ===
        selected
          .map((s) => s.id)
          .sort((a, b) => a - b)
          .join(",") &&
      !pts.allPTs.some((p) => p.id === pt.id)
    ) {
      setPT(undefined);
      setNeedsPTChoice(true);
      key.current = "";
      setReview(false);
      setDraftNotice(
        "HLV không còn phù hợp toàn khoảng giờ. Hãy xác minh lại hình thức tập.",
      );
    }
  }, [pt, pts.loading, pts.forIDs, pts.allPTs]);
  useEffect(() => {
    if (!operations && selected.length)
      saveDraft(
        Number(id),
        selected,
        session.user.id,
        pt?.id || restorePT.current,
        !!pt || needsPTChoice,
      );
  }, [selected, pt, needsPTChoice, id, operations, session.user.id]);
  useEffect(() => {
    if (operations) return;
    const refresh = () => {
      if (
        document.visibilityState === "visible" &&
        !submitting.current &&
        !review
      )
        setRevision((v) => v + 1);
    };
    window.addEventListener("focus", refresh);
    const timer = setInterval(refresh, 60000);
    return () => {
      window.removeEventListener("focus", refresh);
      clearInterval(timer);
    };
  }, [review, operations]);
  function choose(next: Option[] = []) {
    if (pt) setNeedsPTChoice(true);
    setSelected(
      [...next].sort(
        (left, right) =>
          new Date(left.starts_at).getTime() -
          new Date(right.starts_at).getTime(),
      ),
    );
    setPT(undefined);
    setPtPage(1);
    key.current = "";
    setError("");
    setRangeError("");
    restorePT.current = 0;
    if (!operations) {
      if (next.length) {
        syncDate(sessionDay(next[0]));
        saveDraft(Number(id), next, session.user.id);
      } else clearDraft(Number(id));
    }
  }
  function changeDate(next: string) {
    if (next === date) return;
    if (selected.length)
      setDraftNotice(
        "Đã xóa khoảng giờ cũ khi chuyển ngày/tuần. Hãy chọn giờ trong lịch đang xem.",
      );
    choose();
    syncDate(next);
    setPage(1);
    setReview(false);
  }
  async function findNext() {
    const start = date < localDate() ? localDate() : date;
    const current = options.data.find(
      (s) =>
        sessionDay(s) >= start &&
        !s.unavailable_reason &&
        Date.parse(s.starts_at) > Date.now(),
    );
    if (current) {
      changeDate(sessionDay(current));
      setDraftNotice(
        `Đã chuyển tới ngày ${sessionDay(current)} còn ca phù hợp.`,
      );
      return;
    }
    for (let offset = 1; offset <= 28; offset += 7) {
      const days = await Promise.all(
        Array.from({ length: 7 }, (_, i) =>
          fetchDay<Option>(
            session.accessToken,
            id!,
            addDays(start, offset + i),
          ),
        ),
      );
      const found = days
        .flat()
        .filter(
          (s) => !s.unavailable_reason && Date.parse(s.starts_at) > Date.now(),
        )
        .sort((a, b) => Date.parse(a.starts_at) - Date.parse(b.starts_at))[0];
      if (found) {
        changeDate(sessionDay(found));
        setDraftNotice(
          `Đã chuyển tới ngày ${sessionDay(found)} còn ca phù hợp.`,
        );
        return;
      }
    }
    setDraftNotice(
      "Chưa tìm thấy ca còn phù hợp trong 28 ngày tới. Liên hệ lễ tân hoặc chọn phòng khác.",
    );
  }
  const checking = useRef(false);
  async function openReview() {
    if (operations) {
      setReview(true);
      return;
    }
    if (checking.current || !selected.length || needsPTChoice) return;
    checking.current = true;
    setBusy(true);
    setError("");
    try {
      const fresh = await fetchDay<Option>(
        session.accessToken,
        id!,
        sessionDay(selected[0]),
      );
      const resolved = selected
        .map((old) =>
          fresh.find(
            (s) =>
              s.id === old.id &&
              s.starts_at === old.starts_at &&
              s.ends_at === old.ends_at,
          ),
        )
        .filter((s): s is Option => !!s);
      if (resolved.length !== selected.length || validateRange(resolved)) {
        choose();
        setRevision((v) => v + 1);
        throw new Error(
          "Giờ đã chọn không còn phù hợp. Lịch đã được cập nhật; hãy chọn lại.",
        );
      }
      if (pt) {
        for (const s of resolved) {
          let page = 1,
            found = false;
          for (;;) {
            const r = await resource.list<PT>(
              session.accessToken,
              "booking-options/pts",
              new URLSearchParams({
                room_session_id: String(s.id),
                page: String(page),
                page_size: "100",
              }),
            );
            if (r.data.some((p) => p.id === pt.id)) {
              found = true;
              break;
            }
            if (
              !r.data.length ||
              !r.meta ||
              page * r.meta.page_size >= r.meta.total
            )
              break;
            page++;
          }
          if (!found) {
            setPT(undefined);
            setNeedsPTChoice(true);
            key.current = "";
            setRevision((v) => v + 1);
            throw new Error(
              "HLV vừa thay đổi lịch. Hãy xác minh lại hình thức tập trước khi tiếp tục.",
            );
          }
        }
      }
      setSelected(resolved);
      setReview(true);
    } catch (e) {
      setError((e as Error).message);
    } finally {
      checking.current = false;
      setBusy(false);
    }
  }
  async function book() {
    if (
      !selected.length ||
      submitting.current ||
      (operations && (!Number.isSafeInteger(memberID) || memberID < 1))
    )
      return;
    submitting.current = true;
    setBusy(true);
    setError("");
    if (!key.current) key.current = crypto.randomUUID();
    try {
      const batch = selected.length > 1;
      const r = await request<Booking | Booking[]>(
        batch
          ? operations
            ? `/api/v1/members/${memberID}/bookings/batch`
            : "/api/v1/bookings/batch"
          : operations
            ? `/api/v1/members/${memberID}/bookings`
            : "/api/v1/bookings",
        {
          method: "POST",
          headers: { "Idempotency-Key": key.current },
          body: JSON.stringify(
            batch
              ? {
                  room_session_ids: selected.map((item) => item.id),
                  pt_id: pt?.id || 0,
                }
              : { room_session_id: selected[0].id, pt_id: pt?.id || 0 },
          ),
        },
        session.accessToken,
      );
      const created = Array.isArray(r.data) ? r.data : [r.data];
      if (!operations) clearDraft(Number(id));
      navigate(
        `${operations ? "/management" : "/member"}/bookings/${created[0].id}`,
        {
          state: {
            bookingResult: {
              ids: created.map((b) => b.id),
              pending: created.some((b) => b.status === "PENDING_PT"),
              starts_at: selected[0].starts_at,
              ends_at: selected.at(-1)!.ends_at,
            },
          },
        },
      );
    } catch (e) {
      setError((e as Error).message);
      if (e instanceof ApiError && e.status < 500 && e.status !== 408) {
        setReview(false);
        setRevision((v) => v + 1);
        key.current = "";
      }
    } finally {
      submitting.current = false;
      setBusy(false);
    }
  }
  if (!room.data || room.data.id !== Number(id))
    return <RecordState {...room} retry={room.refresh} />;
  const r = room.data;
  return (
    <>
      <Link
        className="back-link"
        to={
          operations
            ? `/management/bookings/new/rooms?member_id=${memberID}`
            : "/member/rooms"
        }
      >
        ← Khám phá phòng
      </Link>
      <div
        className={`room-detail-hero ${operations ? "" : "booking-week-hero"}`}
      >
        <RoomImage room={r} detail />
        <div>
          <span className="portal-kicker">
            {r.room_type} · {r.default_capacity} CHỖ
          </span>
          <h1>{r.name}</h1>
          <p>
            {r.description.startsWith("Dữ liệu DEMO cục bộ do ")
              ? "Không gian minh họa cho buổi tập của bạn."
              : r.description}
          </p>
        </div>
      </div>
      <div
        className={`booking-layout ${operations ? "" : "booking-week-layout"}`}
      >
        {operations && (!Number.isSafeInteger(memberID) || memberID < 1) && (
          <Notice>
            Chọn hội viên từ màn hình đặt lịch hộ trước khi xác nhận.
          </Notice>
        )}
        <section className="booking-picker">
          <h2>
            <span className="step-number">01</span> Chọn buổi tập
          </h2>
          <Field label="Ngày tập">
            <Input
              type="date"
              min={localDate()}
              value={date}
              onChange={(e) => {
                if (!e.target.value) return;
                changeDate(e.target.value);
              }}
            />
          </Field>
          <DateStrip
            date={date}
            change={(v) => {
              changeDate(v);
            }}
          />
          <Notice>{options.error}</Notice>
          {options.error && (
            <Button
              variant="secondary"
              onClick={() => setRevision((v) => v + 1)}
            >
              Thử tải lại lịch
            </Button>
          )}
          {draftNotice && (
            <p className="booking-draft-note" role="status">
              {draftNotice}
            </p>
          )}
          <Notice>{!review ? error : ""}</Notice>
          {!operations && (
            <p
              className="booking-refresh-status"
              role="status"
              aria-live="polite"
            >
              {options.loading && weeklyOptions.hasSnapshot
                ? "Đang cập nhật chỗ và điều kiện đặt…"
                : ""}
            </p>
          )}
          {options.loading && (operations || !weeklyOptions.hasSnapshot) ? (
            <p role="status">Đang kiểm tra lịch trống…</p>
          ) : (
            <>
              {!options.data.length && (
                <p className="portal-empty">
                  Ngày này chưa có ca tập. Hãy chọn ngày khác.
                </p>
              )}
              {operations ? (
                <SessionRangePicker
                  allowRange={!operations}
                  formatTime={time}
                  sessions={options.data}
                  selectedIDs={selected.map((item) => item.id)}
                  onRangeError={setRangeError}
                  onSelect={(sessions) => {
                    const chosen = new Set(sessions.map((item) => item.id));
                    choose(options.data.filter((item) => chosen.has(item.id)));
                  }}
                />
              ) : (
                <WeekCalendar
                  sessions={options.data}
                  selectedIDs={selected.map((s) => s.id)}
                  date={date}
                  onDate={changeDate}
                  onSelect={choose}
                  onError={setRangeError}
                  onFindNext={findNext}
                />
              )}
              <Notice>{rangeError}</Notice>
              {operations && (
                <Pagination
                  meta={options.meta}
                  change={(k, v) => {
                    if (k === "page") {
                      setPage(Number(v));
                      choose();
                    }
                    if (k === "page_size") {
                      setPageSize(Number(v));
                      setPage(1);
                      choose();
                    }
                  }}
                />
              )}
            </>
          )}
          <h2>
            <span className="step-number">02</span> Tập theo cách của bạn
          </h2>
          <button
            className={`pt-option ${!pt && !needsPTChoice ? "selected" : ""}`}
            onClick={() => {
              setPT(undefined);
              setNeedsPTChoice(false);
              key.current = "";
            }}
            aria-pressed={!pt && !needsPTChoice}
          >
            Tự tập <span>Không cần PT xác nhận</span>
          </button>
          {needsPTChoice && (
            <p className="booking-draft-note" role="status">
              Hãy chọn Tự tập hoặc một HLV phù hợp để xác nhận hình thức tập.
            </p>
          )}
          {!!selected.length && (
            <>
              <Notice>{pts.error}</Notice>
              {pts.error && (
                <Button
                  variant="secondary"
                  onClick={() => setRevision((v) => v + 1)}
                >
                  Thử tải lại HLV
                </Button>
              )}
              {pts.loading && <p role="status">Đang tìm PT rảnh…</p>}
              {!pts.loading && !pts.error && !pts.data.length && (
                <p>
                  Không có PT phù hợp toàn bộ thời gian đã chọn. Bạn vẫn có thể
                  tự tập.
                </p>
              )}
              <div className="booking-pt-grid">
                {pts.data.map((p) => (
                  <button
                    key={p.id}
                    className={`pt-option ${pt?.id === p.id ? "selected" : ""}`}
                    aria-pressed={pt?.id === p.id}
                    onClick={() => {
                      setPT(p);
                      setNeedsPTChoice(false);
                      key.current = "";
                    }}
                  >
                    <strong>{p.full_name}</strong>
                    <span>{p.specialty || "Huấn luyện cá nhân"}</span>
                    <RatingSummary
                      average={p.rating_average}
                      count={p.review_count}
                    />
                  </button>
                ))}
              </div>
              <Pagination
                meta={pts.meta}
                change={(k, v) => {
                  if (k === "page") setPtPage(Number(v));
                  if (k === "page_size") {
                    setPtPageSize(Number(v));
                    setPtPage(1);
                  }
                }}
              />
            </>
          )}
        </section>
        <aside className="booking-summary" hidden={!operations && smallScreen}>
          <span className="portal-kicker">BUỔI TẬP CỦA BẠN</span>
          <h2>{r.name}</h2>
          <p>{day(`${date}T12:00:00+07:00`)}</p>
          <strong>
            {selected.length
              ? `${time(selected[0].starts_at)} – ${time(selected[selected.length - 1].ends_at)}${selected.length > 1 ? ` · ${selected.length} ca` : ""}`
              : "Chưa chọn giờ"}
          </strong>
          <p>
            {pt
              ? `Cùng ${pt.full_name}`
              : needsPTChoice
                ? "Cần chọn lại hình thức tập"
                : "Tự tập"}
          </p>
          <p className="booking-note">
            {pt
              ? "Giữ chỗ trong khi chờ PT. PT có tối đa 30 phút, và phải phản hồi trước lúc ca bắt đầu."
              : "Lịch tự tập được xác nhận ngay khi còn chỗ."}
          </p>
          <Button
            busy={busy}
            disabled={
              !selected.length ||
              needsPTChoice ||
              !!options.error ||
              options.loading ||
              (!!pt && (pts.loading || !!pts.error))
            }
            onClick={() => void openReview()}
          >
            Kiểm tra đặt lịch →
          </Button>
          <small>Có thể hủy trước giờ tập ít nhất 2 tiếng.</small>
        </aside>
        {!operations && (
          <div className="booking-mobile-bar" hidden={!smallScreen}>
            <div>
              <strong>
                {selected.length
                  ? `${time(selected[0].starts_at)} – ${time(selected.at(-1)!.ends_at)} · ${selected.length} ca`
                  : "Chọn giờ trong lịch"}
              </strong>
              <span>
                {day(`${date}T12:00:00+07:00`)} ·{" "}
                {pt?.full_name ||
                  (needsPTChoice ? "Chọn hình thức tập" : "Tự tập")}
              </span>
            </div>
            <Button
              busy={busy}
              disabled={
                !selected.length ||
                needsPTChoice ||
                !!options.error ||
                options.loading ||
                (!!pt && (pts.loading || !!pts.error))
              }
              onClick={() => void openReview()}
            >
              Kiểm tra đặt lịch →
            </Button>
          </div>
        )}
      </div>
      <Modal
        open={review}
        onClose={() => setReview(false)}
        busy={busy}
        title="Xác nhận đặt lịch"
        description={`${r.name} · ${civilDate(date)} · ${selected.length ? `${time(selected[0].starts_at)} – ${time(selected[selected.length - 1].ends_at)} · ${selected.length} ca · ${Math.round((Date.parse(selected.at(-1)!.ends_at) - Date.parse(selected[0].starts_at)) / 60000)} phút` : ""} · ${pt?.full_name || "Tự tập"}`}
      >
        <p>
          {pt
            ? "Sau xác nhận, chỗ được giữ trong khi chờ HLV. HLV phản hồi trong tối đa 30 phút và trước lúc ca bắt đầu."
            : "Lịch tự tập được xác nhận khi gửi thành công và còn chỗ."}
        </p>
        <p>
          Có thể hủy trước giờ tập ít nhất 2 tiếng. Đang xem lại chưa giữ chỗ.
        </p>
        <Notice>{error}</Notice>
        <Button busy={busy} onClick={() => void book()}>
          Xác nhận đặt lịch
        </Button>
      </Modal>
    </>
  );
}
function Deadline({ booking }: { booking: Booking }) {
  const [seconds, setSeconds] = useState(0);
  useEffect(() => {
    const offset = new Date(booking.server_time).getTime() - Date.now();
    const update = () =>
      setSeconds(
        Math.max(
          0,
          Math.ceil(
            (new Date(booking.respond_by || "").getTime() -
              Date.now() -
              offset) /
              1000,
          ),
        ),
      );
    update();
    const timer = setInterval(update, 1000);
    return () => clearInterval(timer);
  }, [booking.server_time, booking.respond_by]);
  return booking.status === "PENDING_PT" ? (
    <span className="deadline">
      {seconds > 0
        ? `Còn ${Math.floor(seconds / 60)}:${String(seconds % 60).padStart(2, "0")} để phản hồi`
        : "Đang cập nhật kết quả…"}
    </span>
  ) : null;
}
export function BookingList({
  requests = false,
  history = false,
  attendance = false,
}: {
  requests?: boolean;
  history?: boolean;
  attendance?: boolean;
}) {
  const [queryParams, setQueryParams] = useSearchParams();
  useEffect(() => {
    if (attendance && !queryParams.has("from") && !queryParams.has("to")) {
      const next = new URLSearchParams(queryParams);
      next.set("from", localDate());
      next.set("to", localDate());
      setQueryParams(next, { replace: true });
    }
  }, [attendance]);
  const { session } = useAuth();
  const [roomQuery, setRoomQuery] = useState("");
  const [roomPage, setRoomPage] = useState(1);
  const [roomPageSize, setRoomPageSize] = useState(100);
  const updateRoomQuery = useCallback((value: string) => {
    setRoomQuery(value);
    setRoomPage(1);
  }, []);
  const list = useList<Booking>(
    requests
      ? "bookings?status=PENDING_PT"
      : history
        ? "bookings?history=true"
        : "bookings",
  );
  const rooms = useOptions<Room>(
    attendance ? "rooms" : null,
    new URLSearchParams({
      q: roomQuery,
      page: String(roomPage),
      page_size: String(roomPageSize),
    }).toString(),
  );
  const base =
    session.user.role === "MEMBER"
      ? "/member"
      : session.user.role === "PT"
        ? "/pt"
        : "/management";
  useEffect(() => {
    const timer = setInterval(() => {
      if (document.visibilityState === "visible") list.refresh();
    }, 15000);
    return () => clearInterval(timer);
  }, []);
  return (
    <>
      <span className="portal-kicker">
        {session.user.role === "PT" ? "COACH WORKSPACE" : "LỊCH TẬP"}
      </span>
      <h1>
        {attendance
          ? "Điểm danh hôm nay"
          : history
            ? "Lịch sử buổi tập"
            : requests
              ? "Yêu cầu đang chờ bạn"
              : session.user.role === "PT"
                ? "Lịch làm việc của bạn"
                : "Lịch đặt của bạn"}
      </h1>
      {["ADMIN", "RECEPTIONIST"].includes(session.user.role) && (
        <Link className="btn btn-primary" to="/management/bookings/new">
          Đặt lịch hộ hội viên
        </Link>
      )}
      <p>
        {requests
          ? "Kiểm tra thời gian và phản hồi trước hạn để xác nhận buổi tập."
          : "Các buổi tập và trạng thái được cập nhật từ hệ thống."}
      </p>
      {!requests && (
        <div className="catalog-toolbar">
          <SearchBox
            value={list.params.get("q") || ""}
            onChange={(v) => list.filter("q", v)}
            placeholder={
              attendance
                ? "Tìm mã booking, mã hội viên, tên hoặc SĐT…"
                : "Tìm mã booking hoặc hội viên…"
            }
          />
          {attendance && (
            <div className="attendance-room-filter">
              <SearchBox
                value={roomQuery}
                onChange={updateRoomQuery}
                placeholder="Tìm tên hoặc loại phòng…"
              />
              <Field label="Phòng tập">
                <Select
                  value={list.params.get("room_id") || ""}
                  onChange={(e) => list.filter("room_id", e.target.value)}
                >
                  <option value="">Tất cả phòng</option>
                  {rooms.data.map((r) => (
                    <option key={r.id} value={r.id}>
                      {r.name}
                    </option>
                  ))}
                </Select>
              </Field>
              {rooms.meta.total > rooms.meta.page_size && (
                <Pagination
                  meta={rooms.meta}
                  change={(key, value) => {
                    if (key === "page") setRoomPage(Number(value));
                    if (key === "page_size") {
                      setRoomPage(1);
                      setRoomPageSize(Number(value));
                    }
                  }}
                />
              )}
            </div>
          )}
          <Select
            aria-label="Lọc trạng thái lịch"
            value={list.params.get("status") || ""}
            onChange={(e) => list.filter("status", e.target.value)}
          >
            <option value="">Tất cả trạng thái</option>
            {Object.entries(labels).map(([v, l]) => (
              <option key={v} value={v}>
                {l}
              </option>
            ))}
          </Select>
          <Field label="Từ ngày">
            <Input
              type="date"
              value={list.params.get("from") || ""}
              onChange={(e) => list.filter("from", e.target.value)}
            />
          </Field>
          <Field label="Đến ngày">
            <Input
              type="date"
              value={list.params.get("to") || ""}
              onChange={(e) => list.filter("to", e.target.value)}
            />
          </Field>
        </div>
      )}
      <Notice>{list.error}</Notice>
      {list.error && (
        <Button variant="secondary" onClick={list.refresh}>
          Thử tải lại
        </Button>
      )}
      {list.loading && <p role="status">Đang tải lịch…</p>}
      {!list.loading && !list.data.length && (
        <div className="portal-empty">Chưa có lịch phù hợp.</div>
      )}
      <div className="booking-list">
        {list.data.map((b) => (
          <Link
            key={b.id}
            to={`${base}/bookings/${b.id}`}
            className="booking-row"
          >
            <div className="booking-date">
              {day(b.starts_at)}
              <strong>{time(b.starts_at)}</strong>
            </div>
            <div>
              <h2>{b.room_name}</h2>
              <p>
                #{b.id} · {b.member_name} · {b.pt_name || "Tự tập"}
              </p>
              <Deadline booking={b} />
            </div>
            <span className={`booking-status status-${b.status}`}>
              {labels[b.status]}
            </span>
            <span aria-hidden="true">↗</span>
          </Link>
        ))}
      </div>
      <Pagination meta={list.meta} change={list.filter} />
    </>
  );
}
export function BookingDetail() {
  const { id } = useParams();
  const { session } = useAuth();
  const location = useLocation();
  const result = (
    location.state as {
      bookingResult?: {
        ids: number[];
        pending: boolean;
        starts_at: string;
        ends_at: string;
      };
    } | null
  )?.bookingResult;
  const record = useRecord<Booking>(`bookings/${id}`);
  const [action, setAction] = useState("");
  const [reason, setReason] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  useEffect(() => {
    const timer = setInterval(() => {
      if (document.visibilityState === "visible") record.refresh();
    }, 15000);
    return () => clearInterval(timer);
  }, []);
  async function perform() {
    setBusy(true);
    setError("");
    try {
      await resource.save(
        session.accessToken,
        `bookings/${id}/${action === "CANCEL" ? "cancel" : "pt-response"}`,
        { action: action === "CANCEL" ? "" : action, reason },
      );
      setAction("");
      record.refresh();
    } catch (e) {
      setError((e as Error).message);
      record.refresh();
    } finally {
      setBusy(false);
    }
  }
  if (!record.data) return <RecordState {...record} retry={record.refresh} />;
  const b = record.data;
  const member = session.user.role === "MEMBER";
  const pt = session.user.role === "PT";
  const active = ["CONFIRMED", "PENDING_PT"].includes(b.status);
  const cancellable =
    active &&
    !b.has_attendance &&
    new Date(b.server_time).getTime() < new Date(b.ends_at).getTime() &&
    (!member ||
      new Date(b.starts_at).getTime() - new Date(b.server_time).getTime() >=
        7200000) &&
    !(pt && b.status === "PENDING_PT");
  return (
    <>
      {result?.ids?.includes(Number(id)) && (
        <section className="booking-draft-note" role="status">
          <h2>
            {result.pending
              ? "Đã gửi yêu cầu — Chờ HLV"
              : "Đặt lịch thành công"}
          </h2>
          <p>
            {day(result.starts_at)} · {time(result.starts_at)} –{" "}
            {time(result.ends_at)} · {result.ids.length} ca
          </p>
          <p>Mã lịch: {result.ids.map((n) => `#${n}`).join(", ")}</p>
          <Link to={member ? "/member/bookings" : "/management/bookings"}>
            Xem lịch của tôi →
          </Link>
        </section>
      )}
      <span className="portal-kicker">ĐẶT LỊCH #{b.id}</span>
      <h1>{b.room_name}</h1>
      <div className="booking-detail">
        <span className={`booking-status status-${b.status}`}>
          {labels[b.status]}
        </span>
        <h2>
          {day(b.starts_at)} · {time(b.starts_at)} – {time(b.ends_at)}
        </h2>
        <p>Hội viên: {b.member_name}</p>
        <p>Huấn luyện viên: {b.pt_name || "Tự tập"}</p>
        <AttendancePanel booking={b} onChange={record.refresh} />
        <ReviewPanel booking={b} />
        <Deadline booking={b} />
        {b.cancellation_reason && <p>Lý do: {b.cancellation_reason}</p>}
        <div className="form-actions">
          {pt && b.status === "PENDING_PT" && (
            <>
              <Button onClick={() => setAction("ACCEPT")}>Nhận buổi tập</Button>
              <Button variant="secondary" onClick={() => setAction("REJECT")}>
                Từ chối
              </Button>
            </>
          )}
          {cancellable && (
            <Button variant="danger" onClick={() => setAction("CANCEL")}>
              Hủy lịch
            </Button>
          )}
        </div>
        {member && active && !cancellable && (
          <p>
            Đã qua hạn hủy trước giờ tập 2 tiếng. Liên hệ lễ tân nếu cần hỗ trợ.
          </p>
        )}
        <Link
          to={
            member
              ? "/member/bookings"
              : pt
                ? "/pt/schedule"
                : "/management/bookings"
          }
        >
          ← Về danh sách lịch
        </Link>
      </div>
      <Modal
        open={!!action}
        onClose={() => {
          setAction("");
          setError("");
        }}
        title={
          action === "CANCEL"
            ? "Hủy buổi tập?"
            : action === "ACCEPT"
              ? "Nhận buổi tập?"
              : "Từ chối yêu cầu?"
        }
        description={`${b.room_name} · ${day(b.starts_at)} ${time(b.starts_at)}`}
        busy={busy}
      >
        <form
          onSubmit={(e) => {
            e.preventDefault();
            void perform();
          }}
        >
          <Notice>{error}</Notice>
          {action !== "ACCEPT" && (
            <Field label="Lý do" required={!member && action === "CANCEL"}>
              <Input
                required={!member && action === "CANCEL"}
                value={reason}
                maxLength={1000}
                onChange={(e) => setReason(e.target.value)}
              />
            </Field>
          )}
          <Button
            busy={busy}
            variant={action === "CANCEL" ? "danger" : "primary"}
          >
            Xác nhận
          </Button>
        </form>
      </Modal>
    </>
  );
}
