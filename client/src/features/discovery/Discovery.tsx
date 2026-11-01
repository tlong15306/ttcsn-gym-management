import { useEffect, useMemo, useRef, useState, type ReactNode } from "react";
import {
  Link,
  NavLink,
  Outlet,
  useLocation,
  useParams,
  useSearchParams,
} from "react-router-dom";
import {
  ArrowLeft,
  ArrowRight,
  ArrowUpRight,
  CalendarDays,
  ChevronLeft,
  ChevronRight,
  Dumbbell,
  Menu,
  Search,
  SlidersHorizontal,
  Users,
  X,
} from "lucide-react";
import { useSession } from "../../app/session";
import { roleHome } from "../../app/navigation";
import {
  civilDate,
  localDate,
  vietnamTime,
  vietnamWeekday,
} from "../../utils/vietnamTime";
import {
  roomDescription,
  roomTypes,
  usePublicData,
  validCivilDate,
  type PublicRoom,
  type PublicTrainer,
  type PublicSession,
} from "./data";
import type { PageMeta } from "../../types";
import { request } from "../../api";
import { WeekCalendar } from "../booking/WeekCalendar";
import {
  addDays,
  sessionDay,
  validateRange,
  weekStart,
  type WeekSession,
} from "../booking/weekly";
import {
  clearDraft,
  readDraft,
  restoreDraft,
  saveDraft,
} from "../booking/draft";
import "./discovery.css";

export function DiscoveryLayout() {
  const { session } = useSession();
  const [menu, setMenu] = useState(false);
  return (
    <div className="discovery">
      <a className="skip-link" href="#discovery-main">
        Đến nội dung chính
      </a>
      <header className="dc-header">
        <Link className="dc-brand" to="/" onClick={() => setMenu(false)}>
          <Dumbbell aria-hidden="true" />
          <span>
            GYM<span>CLUB</span>
          </span>
        </Link>
        <nav
          aria-label="Điều hướng khám phá"
          className={menu ? "open" : ""}
          onClick={() => setMenu(false)}
        >
          <NavLink end to="/">
            Trang chủ
          </NavLink>
          <NavLink to="/explore/rooms">Phòng tập</NavLink>
          <NavLink to="/explore/trainers">Huấn luyện viên</NavLink>
          <Link to="/#guide">Hướng dẫn</Link>
        </nav>
        <Link
          className="dc-button dc-button-small"
          to={session ? roleHome(session.user.role) : "/login"}
        >
          {session ? "Không gian của tôi" : "Đăng nhập"}
        </Link>
        <button
          className="dc-icon dc-menu"
          aria-label={menu ? "Đóng điều hướng" : "Mở điều hướng"}
          aria-expanded={menu}
          onClick={() => setMenu(!menu)}
        >
          {menu ? <X /> : <Menu />}
        </button>
      </header>
      <main id="discovery-main">
        <Outlet />
      </main>
      <footer className="dc-footer dc-width">
        <Link className="dc-brand" to="/">
          GYM CLUB
        </Link>
        <span>Chủ động thời gian. Tập theo cách của bạn.</span>
        <small>Cần tài khoản? Liên hệ lễ tân để được hỗ trợ.</small>
      </footer>
    </div>
  );
}

function PublicImage({
  room,
  hero = false,
}: {
  room: PublicRoom;
  hero?: boolean;
}) {
  const [failed, setFailed] = useState(false);
  useEffect(() => setFailed(false), [room.id, room.cover_version]);
  return (
    <div className="dc-photo">
      {room.has_cover && !failed ? (
        <img
          src={`/api/v1/public/rooms/${room.id}/cover-image?size=${hero ? "detail" : "card"}&v=${room.cover_version}`}
          alt={room.cover_alt_text || room.name}
          loading={hero ? "eager" : "lazy"}
          onError={() => setFailed(true)}
        />
      ) : (
        <div className="dc-photo-empty">
          <Dumbbell aria-hidden="true" />
          <span>
            {failed ? "Chưa tải được ảnh phòng" : "Phòng chưa có ảnh"}
          </span>
        </div>
      )}
      {room.cover_is_demo && room.has_cover && !failed && (
        <span className="dc-photo-label">Ảnh minh họa</span>
      )}
    </div>
  );
}
function LoadState({
  loading,
  error,
  retry,
  children,
}: {
  loading: boolean;
  error: string;
  retry: () => void;
  children: ReactNode;
}) {
  if (loading)
    return (
      <div className="dc-skeletons" role="status" aria-label="Đang tải dữ liệu">
        <div />
        <div />
        <div />
      </div>
    );
  if (error)
    return (
      <div className="dc-empty" role="alert">
        <h2>Chưa tải được dữ liệu.</h2>
        <p>{error}</p>
        <button className="dc-button dc-button-small" onClick={retry}>
          Thử lại
        </button>
      </div>
    );
  return <>{children}</>;
}
function RoomCards({ rooms }: { rooms: PublicRoom[] }) {
  return (
    <div className="dc-room-grid">
      {rooms.map((room) => (
        <article key={room.id} className="dc-room-card">
          <Link
            to={`/explore/rooms/${room.id}`}
            aria-label={`Xem ${room.name}`}
          >
            <PublicImage room={room} />
          </Link>
          <div className="dc-card-body">
            <div className="dc-card-meta">
              <span>{roomTypes[room.room_type] || room.room_type}</span>
              <span>{room.default_capacity} chỗ / ca</span>
            </div>
            <h3>
              <Link to={`/explore/rooms/${room.id}`}>{room.name}</Link>
            </h3>
            <p>{roomDescription(room)}</p>
            <div className="dc-card-bottom">
              <span
                className={room.status === "MAINTENANCE" ? "dc-warning" : ""}
              >
                {room.status === "AVAILABLE"
                  ? "Sẵn sàng đón bạn"
                  : "Đang bảo trì"}
              </span>
              <Link to={`/explore/rooms/${room.id}`}>
                Xem lịch <ArrowUpRight size={16} aria-hidden="true" />
              </Link>
            </div>
          </div>
        </article>
      ))}
    </div>
  );
}
function TrainerCards({ trainers }: { trainers: PublicTrainer[] }) {
  return (
    <div className="dc-trainer-grid">
      {trainers.map((trainer) => (
        <article key={trainer.id} className="dc-trainer">
          <span className="dc-avatar" aria-hidden="true">
            {trainer.full_name
              .split(/\s+/)
              .filter(Boolean)
              .slice(-2)
              .map((s) => s[0])
              .join("")}
          </span>
          <div>
            <h3>{trainer.full_name}</h3>
            <strong>{trainer.specialty || "Huấn luyện cá nhân"}</strong>
            <p>
              {trainer.experience_text ||
                "Chọn phòng và giờ tập để kiểm tra lịch phù hợp."}
            </p>
            <small>
              {trainer.rating_average == null || !trainer.review_count
                ? "Chưa có đánh giá"
                : `${trainer.rating_average.toFixed(1)}/5 · ${trainer.review_count} đánh giá`}
            </small>
            <Link className="dc-text-link" to="/explore/rooms">
              Chọn phòng và giờ tập <ArrowRight size={15} aria-hidden="true" />
            </Link>
          </div>
        </article>
      ))}
    </div>
  );
}
export function DiscoveryHome() {
  const location = useLocation();
  const rooms = usePublicData<PublicRoom[]>("rooms?page_size=3");
  const trainers = usePublicData<PublicTrainer[]>("trainers?page_size=2");
  const hero =
    rooms.data?.find((room) => room.status === "AVAILABLE" && room.has_cover) ||
    rooms.data?.[0];
  useEffect(() => {
    if (location.hash === "#guide")
      document.getElementById("guide")?.scrollIntoView();
  }, [location.hash]);
  return (
    <>
      <section className="dc-hero dc-width">
        <div className="dc-hero-copy">
          <p className="dc-intro">
            <span />
            Không gian tập luyện của bạn
          </p>
          <h1>
            Tập theo nhịp
            <br />
            của bạn.
          </h1>
          <p className="dc-hero-description">
            Chọn không gian, tìm giờ phù hợp.
            <br />
            Tự tập hoặc đồng hành cùng huấn luyện viên.
          </p>
          <div className="dc-actions">
            <Link to="/explore/rooms" className="dc-button">
              Khám phá phòng tập <ArrowUpRight size={18} aria-hidden="true" />
            </Link>
            <Link className="dc-text-link" to="/explore/rooms">
              Xem lịch tuần <ArrowRight size={17} aria-hidden="true" />
            </Link>
          </div>
          <p className="dc-footnote">
            Cứ khám phá trước. Đăng nhập khi bạn sẵn sàng đặt lịch.
          </p>
        </div>
        <div className="dc-hero-media">
          <>
            <div className="dc-photo hero">
              <img
                src="/images/gym-home-demo.jpg"
                alt="Không gian tập sức mạnh minh họa"
                fetchPriority="high"
              />
              <span className="dc-photo-label">Ảnh minh họa</span>
            </div>
            <div className="dc-hero-caption">
              <div>
                <strong>Không gian cho buổi tập của bạn.</strong>
                <span>Khám phá phòng và lịch tập đã công bố.</span>
              </div>
              <Link to="/explore/rooms" aria-label="Khám phá phòng tập">
                <ArrowUpRight aria-hidden="true" />
              </Link>
            </div>
          </>
        </div>
      </section>
      <section className="dc-benefits dc-width">
        <div>
          <CalendarDays aria-hidden="true" />
          <span>
            <strong>Xem lịch trước</strong>
            <small>Khám phá mà không cần tài khoản</small>
          </span>
        </div>
        <div>
          <SlidersHorizontal aria-hidden="true" />
          <span>
            <strong>Chủ động giờ tập</strong>
            <small>Tìm khoảng giờ phù hợp với bạn</small>
          </span>
        </div>
        <div>
          <Users aria-hidden="true" />
          <span>
            <strong>Tập theo cách của bạn</strong>
            <small>Tự tập hoặc có HLV đồng hành</small>
          </span>
        </div>
      </section>
      <section className="dc-section dc-width">
        <div className="dc-section-title">
          <div>
            <p className="dc-label">Phòng tập</p>
            <h2>Tìm đúng không gian.</h2>
          </div>
          <Link className="dc-text-link" to="/explore/rooms">
            Xem tất cả phòng <ArrowRight size={18} aria-hidden="true" />
          </Link>
        </div>
        <LoadState {...rooms}>
          {rooms.data?.length ? (
            <RoomCards rooms={rooms.data} />
          ) : (
            <div className="dc-empty">
              <p>Chưa có phòng được công bố. Vui lòng quay lại sau.</p>
            </div>
          )}
        </LoadState>
      </section>
      <section className="dc-trainer-band">
        <div className="dc-width">
          <div className="dc-section-title">
            <div>
              <p className="dc-label">Huấn luyện viên</p>
              <h2>
                Có người đồng hành.
                <br />
                Có thêm động lực.
              </h2>
            </div>
            <Link to="/explore/trainers" className="dc-text-link">
              Khám phá HLV <ArrowRight size={18} aria-hidden="true" />
            </Link>
          </div>
          <LoadState {...trainers}>
            {trainers.data?.length ? (
              <TrainerCards trainers={trainers.data} />
            ) : (
              <p>
                Chưa có HLV được công bố. Bạn vẫn có thể khám phá phòng và lịch
                tập.
              </p>
            )}
          </LoadState>
        </div>
      </section>
      <section className="dc-section dc-width" id="guide">
        <div className="dc-section-title">
          <h2>Một buổi tập, ba bước.</h2>
        </div>
        <div className="dc-guide">
          {[
            [
              "Chọn phòng và giờ",
              "Xem lịch các phòng, tìm ngày và giờ phù hợp.",
            ],
            ["Chọn cách tập", "Tự tập hoặc chọn HLV phù hợp với cả buổi."],
            [
              "Xem lại và xác nhận",
              "Đăng nhập, kiểm tra điều kiện và xác nhận lịch của bạn.",
            ],
          ].map(([title, body], index) => (
            <article key={title}>
              <span>{index + 1}</span>
              <h3>{title}</h3>
              <p>{body}</p>
            </article>
          ))}
        </div>
      </section>
    </>
  );
}
function PublicPagination({
  meta,
  change,
}: {
  meta?: PageMeta;
  change: (page: number) => void;
}) {
  if (!meta || meta.total <= meta.page_size) return null;
  return (
    <nav className="dc-pagination" aria-label="Phân trang">
      <button
        className="dc-icon"
        aria-label="Trang trước"
        disabled={meta.page <= 1}
        onClick={() => change(meta.page - 1)}
      >
        <ChevronLeft />
      </button>
      <span>
        Trang {meta.page} / {Math.ceil(meta.total / meta.page_size)}
      </span>
      <button
        className="dc-icon"
        aria-label="Trang sau"
        disabled={meta.page * meta.page_size >= meta.total}
        onClick={() => change(meta.page + 1)}
      >
        <ChevronRight />
      </button>
    </nav>
  );
}
export function DiscoveryRooms() {
  const [params, setParams] = useSearchParams();
  const query = new URLSearchParams({
    q: params.get("q") || "",
    room_type: params.get("room_type") || "",
    page: params.get("page") || "1",
    page_size: "6",
  });
  const rooms = usePublicData<PublicRoom[]>(`rooms?${query}`);
  function filter(key: string, value: string) {
    const next = new URLSearchParams(params);
    value ? next.set(key, value) : next.delete(key);
    if (key !== "page") next.delete("page");
    setParams(next, { replace: true });
  }
  return (
    <section className="dc-section dc-width">
      <div className="dc-page-heading">
        <p className="dc-label">Khám phá phòng tập</p>
        <h1>Tìm không gian cho buổi tập.</h1>
        <p>Xem phòng và lịch đã công bố. Đăng nhập khi bạn muốn đặt lịch.</p>
      </div>
      <div className="dc-filters">
        <label className="dc-search">
          <Search size={19} aria-hidden="true" />
          <input
            aria-label="Tìm phòng tập"
            placeholder="Tìm tên hoặc loại phòng…"
            value={query.get("q") || ""}
            onChange={(event) => filter("q", event.target.value)}
            maxLength={120}
          />
        </label>
        <label className="dc-select">
          Loại phòng
          <select
            value={query.get("room_type") || ""}
            onChange={(event) => filter("room_type", event.target.value)}
          >
            <option value="">Tất cả loại phòng</option>
            {Object.entries(roomTypes).map(([value, label]) => (
              <option key={value} value={value}>
                {label}
              </option>
            ))}
          </select>
        </label>
      </div>
      <LoadState {...rooms}>
        {rooms.data?.length ? (
          <RoomCards rooms={rooms.data} />
        ) : (
          <div className="dc-empty">
            <h2>Chưa có phòng phù hợp.</h2>
            <p>Thử đổi từ khóa hoặc loại phòng.</p>
            <button
              className="dc-button dc-button-small"
              onClick={() => setParams({})}
            >
              Xóa bộ lọc
            </button>
          </div>
        )}
      </LoadState>
      <PublicPagination
        meta={rooms.meta}
        change={(page) => filter("page", String(page))}
      />
    </section>
  );
}
export function DiscoveryTrainers() {
  const [params, setParams] = useSearchParams();
  const q = params.get("q") || "";
  const trainers = usePublicData<PublicTrainer[]>(
    `trainers?${new URLSearchParams({ q, page: params.get("page") || "1", page_size: "8" })}`,
  );
  return (
    <section className="dc-section dc-width">
      <div className="dc-page-heading">
        <p className="dc-label">Huấn luyện cá nhân</p>
        <h1>Tìm người đồng hành.</h1>
        <p>
          Các HLV đã được duyệt chuyên môn. Chọn phòng và giờ để kiểm tra lịch
          phù hợp trước khi đặt.
        </p>
      </div>
      <div className="dc-filters">
        <label className="dc-search">
          <Search size={19} aria-hidden="true" />
          <input
            aria-label="Tìm huấn luyện viên"
            placeholder="Tìm tên hoặc chuyên môn…"
            value={q}
            maxLength={120}
            onChange={(event) =>
              setParams(event.target.value ? { q: event.target.value } : {}, {
                replace: true,
              })
            }
          />
        </label>
      </div>
      <LoadState {...trainers}>
        {trainers.data?.length ? (
          <TrainerCards trainers={trainers.data} />
        ) : (
          <div className="dc-empty">
            <h2>Chưa có HLV phù hợp.</h2>
            <p>Thử đổi từ khóa hoặc quay lại sau.</p>
          </div>
        )}
      </LoadState>
      <PublicPagination
        meta={trainers.meta}
        change={(page) => setParams({ q, page: String(page) })}
      />
    </section>
  );
}
function BookingLink({
  id,
  date,
  disabled = false,
}: {
  id: number;
  date: string;
  disabled?: boolean;
}) {
  const { session } = useSession();
  const target = `/member/rooms/${id}?date=${date}`;
  if (disabled)
    return <span className="dc-unavailable">Chưa nhận đặt lịch</span>;
  if (session && session.user.role !== "MEMBER")
    return (
      <Link className="dc-text-link" to={roleHome(session.user.role)}>
        Vào không gian của tôi <ArrowRight size={15} aria-hidden="true" />
      </Link>
    );
  return (
    <Link
      className="dc-button dc-button-small"
      to={session ? target : "/login"}
      state={session ? undefined : { from: target }}
    >
      {session ? "Chọn giờ và đặt lịch" : "Đăng nhập để đặt lịch"}
      <ArrowRight size={15} aria-hidden="true" />
    </Link>
  );
}
export function DiscoveryRoom() {
  const { id } = useParams();
  const { session } = useSession();
  const [params, setParams] = useSearchParams();
  const date = validCivilDate(params.get("from"))
    ? params.get("from")!
    : localDate();
  const from = weekStart(date);
  const [selected, setSelected] = useState<WeekSession[]>([]),
    [wantsPT, setWantsPT] = useState(false),
    [selectionNote, setSelectionNote] = useState("");
  const restored = useRef("");
  const day = (offset: number) => {
    const date = new Date(`${from}T00:00:00Z`);
    date.setUTCDate(date.getUTCDate() + offset);
    return date.toISOString().slice(0, 10);
  };
  const room = usePublicData<PublicRoom>(`rooms/${id}`);
  const sessions = usePublicData<PublicSession[]>(
    `rooms/${id}/sessions?${new URLSearchParams({ from, to: day(6), page_size: "100" })}`,
    true,
  );
  const options = useMemo<WeekSession[]>(
    () =>
      (sessions.data || []).map((s) => ({
        ...s,
        unavailable_reason:
          room.data?.status === "MAINTENANCE"
            ? "Phòng đang bảo trì"
            : s.status !== "OPEN"
              ? "Ca chưa nhận đặt"
              : Date.parse(s.starts_at) <= Date.now()
                ? "Ca đã bắt đầu"
                : s.remaining < 1
                  ? "Đã hết chỗ"
                  : "",
      })),
    [sessions.data, room.data?.status],
  );
  useEffect(() => {
    if (sessions.loading || sessions.error) return;
    const marker = `${id}:${from}`;
    if (restored.current !== marker) {
      restored.current = marker;
      setSelected([]);
      const draft = readDraft();
      if (
        draft?.roomID === Number(id) &&
        weekStart(draft.date) === from &&
        (draft.owner === null || draft.owner === session?.user.id)
      ) {
        const result = restoreDraft(
          draft,
          Number(id),
          session?.user.id || 0,
          options,
        );
        if (result.values.length) {
          setSelected(result.values);
          setWantsPT(draft.wantsPT);
          setSelectionNote(
            "Đã khôi phục ý định chọn giờ. Chưa giữ chỗ; điều kiện Hội viên sẽ được kiểm tra sau đăng nhập.",
          );
        } else if (result.error) {
          clearDraft(Number(id));
          setSelectionNote(result.error);
        }
      }
    } else if (selected.length) {
      const current = selected
        .map((old) => options.find((s) => s.id === old.id))
        .filter((s): s is WeekSession => !!s);
      if (current.length !== selected.length || validateRange(current)) {
        setSelected([]);
        clearDraft(Number(id));
        setSelectionNote("Lịch vừa thay đổi. Hãy chọn lại các ca còn phù hợp.");
      }
    }
  }, [id, from, sessions.loading, sessions.error, options]);
  function choose(values: WeekSession[]) {
    setSelected(values);
    setSelectionNote("");
    if (values.length) {
      setParams({ from: sessionDay(values[0]) }, { replace: true });
      saveDraft(Number(id), values, session?.user.id || null, 0, wantsPT);
    } else clearDraft(Number(id));
  }
  function changeDate(next: string) {
    if (next === date) return;
    if (selected.length)
      setSelectionNote(
        "Đã xóa khoảng giờ cũ khi chuyển ngày/tuần. Hãy chọn lại trong lịch đang xem.",
      );
    setSelected([]);
    clearDraft(Number(id));
    setParams({ from: next });
  }
  function shift(offset: number) {
    changeDate(day(offset));
  }
  async function findNext() {
    const start = date < localDate() ? localDate() : date;
    const current = options
      .filter((s) => sessionDay(s) >= start && !s.unavailable_reason)
      .sort((a, b) => Date.parse(a.starts_at) - Date.parse(b.starts_at))[0];
    if (current) {
      changeDate(sessionDay(current));
      setSelectionNote(`Đã chuyển tới ngày ${sessionDay(current)} còn ca.`);
      return;
    }
    for (let offset = 1; offset <= 28; offset += 7) {
      const begin = addDays(start, offset),
        end = addDays(begin, 6);
      let page = 1;
      const all: PublicSession[] = [];
      for (;;) {
        const r = await request<PublicSession[]>(
          `/api/v1/public/rooms/${id}/sessions?${new URLSearchParams({ from: begin, to: end, page: String(page), page_size: "100" })}`,
        );
        all.push(...r.data);
        if (
          !r.data.length ||
          !r.meta ||
          page * r.meta.page_size >= r.meta.total
        )
          break;
        page++;
      }
      const found = all
        .filter(
          (s) =>
            s.status === "OPEN" &&
            s.remaining > 0 &&
            Date.parse(s.starts_at) > Date.now(),
        )
        .sort((a, b) => Date.parse(a.starts_at) - Date.parse(b.starts_at))[0];
      if (found) {
        changeDate(sessionDay(found));
        setSelectionNote(`Đã chuyển tới ngày ${sessionDay(found)} còn ca.`);
        return;
      }
    }
    setSelectionNote(
      "Chưa tìm thấy ca còn chỗ trong 28 ngày tới. Hãy thử phòng khác hoặc liên hệ lễ tân.",
    );
  }
  return (
    <section className="dc-section dc-width">
      <Link className="dc-back" to="/explore/rooms">
        <ArrowLeft size={15} aria-hidden="true" />
        Khám phá phòng tập
      </Link>
      <LoadState {...room}>
        {room.data && (
          <>
            <div className="dc-room-detail">
              <PublicImage room={room.data} hero />
              <div>
                <p className="dc-label">
                  {roomTypes[room.data.room_type] || room.data.room_type}
                </p>
                <h1>{room.data.name}</h1>
                <p>{roomDescription(room.data)}</p>
                <div className="dc-detail-meta">
                  <span>{room.data.default_capacity} chỗ / ca</span>
                  <span>
                    {room.data.status === "AVAILABLE"
                      ? "Đang phục vụ"
                      : "Đang bảo trì"}
                  </span>
                </div>
              </div>
            </div>
            {room.data.status === "MAINTENANCE" && (
              <p className="dc-notice">
                Phòng đang bảo trì và chưa nhận đặt lịch. Bạn có thể khám phá
                các phòng khác.
              </p>
            )}
            <div className="dc-section-title dc-calendar-title">
              <div>
                <h2>Lịch nhận đặt</h2>
                <p>
                  Giờ Việt Nam · Số chỗ cập nhật khi tải lịch, chưa giữ chỗ.
                </p>
              </div>
              <div className="dc-week-nav">
                <button
                  className="dc-icon"
                  aria-label="7 ngày trước"
                  onClick={() => shift(-7)}
                >
                  <ChevronLeft />
                </button>
                <strong>
                  {civilDate(from)} – {civilDate(day(6))}
                </strong>
                <button
                  className="dc-icon"
                  aria-label="7 ngày tiếp"
                  onClick={() => shift(7)}
                >
                  <ChevronRight />
                </button>
                <button
                  className="dc-button-quiet"
                  onClick={() => changeDate(localDate())}
                >
                  Từ hôm nay
                </button>
              </div>
            </div>
            <LoadState {...sessions}>
              <WeekCalendar
                sessions={options}
                selectedIDs={selected.map((s) => s.id)}
                date={date}
                onDate={changeDate}
                onSelect={choose}
                onError={setSelectionNote}
                onFindNext={findNext}
              />
              <div
                className="dc-intent"
                role="group"
                aria-label="Hình thức tập dự kiến"
              >
                <button
                  aria-pressed={!wantsPT}
                  onClick={() => {
                    setWantsPT(false);
                    if (selected.length)
                      saveDraft(
                        Number(id),
                        selected,
                        session?.user.id || null,
                        0,
                        false,
                      );
                  }}
                >
                  Tự tập
                </button>
                <button
                  aria-pressed={wantsPT}
                  onClick={() => {
                    setWantsPT(true);
                    if (selected.length)
                      saveDraft(
                        Number(id),
                        selected,
                        session?.user.id || null,
                        0,
                        true,
                      );
                  }}
                >
                  Cần huấn luyện viên
                </button>
              </div>
              {wantsPT && (
                <p className="dc-calendar-hint">
                  Bạn sẽ chọn HLV khả dụng cho toàn khoảng giờ sau khi đăng nhập
                  Hội viên.
                </p>
              )}
              {selectionNote && (
                <p className="booking-draft-note" role="status">
                  {selectionNote}
                </p>
              )}
              <div className="dc-selection-summary">
                <div>
                  <strong>
                    {selected.length
                      ? `${vietnamTime(selected[0].starts_at)} – ${vietnamTime(selected.at(-1)!.ends_at)} · ${selected.length} ca`
                      : "Chọn giờ để tiếp tục"}
                  </strong>
                  <p>
                    {selected.length
                      ? civilDate(sessionDay(selected[0]))
                      : "Kéo hoặc chọn hai đầu của khoảng tập."}{" "}
                    · Chưa giữ chỗ.
                  </p>
                </div>
                {selected.length ? (
                  <BookingLink
                    id={room.data.id}
                    date={sessionDay(selected[0])}
                    disabled={room.data.status !== "AVAILABLE"}
                  />
                ) : (
                  <button className="dc-button dc-button-small" disabled>
                    Chọn giờ trước khi đặt
                  </button>
                )}
              </div>
            </LoadState>
          </>
        )}
      </LoadState>
    </section>
  );
}
