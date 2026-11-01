import { useState } from "react";
import { Link } from "react-router-dom";
import { Button, Field, Input, Notice } from "../../components/ui";
import { useAuth } from "../../app/session";
import { useEffect } from "react";
import { resource } from "../../api";
import { localDate, type Booking } from "./Booking";
import { TimeCalendar } from "./TimeCalendar";
import { useRecord } from "../../components/shared/record";
import { vietnamDay, vietnamTime } from "../../utils/vietnamTime";
export function CoachSchedule() {
  const [start, setStart] = useState(localDate());
  const [bookings, setBookings] = useState<Booking[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [version, setVersion] = useState(0);
  const { session } = useAuth();
  const profile = useRecord<{ approval_status: string }>("pts/me");
  const days = Array.from({ length: 7 }, (_, i) => {
    const d = new Date(`${start}T12:00:00+07:00`);
    d.setUTCDate(d.getUTCDate() + i);
    return localDate(d);
  });
  useEffect(() => {
    const abort = new AbortController();
    setLoading(true);
    setError("");
    async function load() {
      const all: Booking[] = [];
      let page = 1;
      while (!abort.signal.aborted) {
        const r = await resource.list<Booking>(
          session.accessToken,
          "bookings",
          new URLSearchParams({
            from: start,
            to: days[6],
            page: String(page),
            page_size: "100",
          }),
          abort.signal,
        );
        all.push(...r.data);
        if (page * r.meta!.page_size >= r.meta!.total) break;
        page++;
      }
      if (!abort.signal.aborted) setBookings(all);
    }
    load()
      .catch((e) => {
        if (!abort.signal.aborted) setError(e.message);
      })
      .finally(() => {
        if (!abort.signal.aborted) setLoading(false);
      });
    return () => abort.abort();
  }, [start, version, session.accessToken]);
  useEffect(() => {
    const timer = setInterval(() => {
      if (document.visibilityState === "visible") setVersion((v) => v + 1);
    }, 15000);
    return () => clearInterval(timer);
  }, []);
  const fmt = vietnamTime;
  function shift(n: number) {
    const d = new Date(`${start}T12:00:00+07:00`);
    d.setUTCDate(d.getUTCDate() + n);
    setStart(localDate(d));
  }
  return (
    <>
      <span className="portal-kicker">COACH WORKSPACE</span>
      <h1>Lịch làm việc của bạn</h1>
      <p>
        Các buổi tập được gán cho bạn. Yêu cầu chờ phản hồi nằm ở trang riêng.
      </p>
      <div className="catalog-toolbar">
        <Button variant="secondary" onClick={() => shift(-7)}>
          ← Tuần trước
        </Button>
        <Field label="Bắt đầu từ">
          <Input
            type="date"
            value={start}
            onChange={(e) => {
              if (e.target.value) setStart(e.target.value);
            }}
          />
        </Field>
        <Button variant="secondary" onClick={() => shift(7)}>
          Tuần sau →
        </Button>
        <Link to="/pt/requests" className="btn btn-primary">
          Xem yêu cầu
        </Link>
      </div>
      <Notice>{error}</Notice>
      {(profile.missing ||
        (profile.data && profile.data.approval_status !== "APPROVED")) && (
        <Notice tone="info">
          Hồ sơ PT chưa đủ điều kiện nhận buổi tập.{" "}
          <Link to="/pt/profile">Hoàn thiện hồ sơ chuyên môn</Link>.
        </Notice>
      )}
      <Link className="back-link" to="/pt/availability">
        Quản lý giờ làm và ngày nghỉ →
      </Link>
      {loading && <p role="status">Đang cập nhật lịch…</p>}
      <TimeCalendar days={days} bookings={bookings} />
      <div className="coach-week">
        {days.map((d) => {
          const entries = bookings
            .filter(
              (b) =>
                localDate(new Date(b.starts_at)) === d &&
                b.status === "CONFIRMED",
            )
            .sort((a, b) => a.starts_at.localeCompare(b.starts_at));
          return (
            <section key={d} className="coach-day">
              <h2>
                {vietnamDay(`${d}T12:00:00+07:00`)}
              </h2>
              {entries.length ? (
                entries.map((b) => (
                  <Link
                    className="coach-event"
                    key={b.id}
                    to={`/pt/bookings/${b.id}`}
                  >
                    <strong>
                      {fmt(b.starts_at)}–{fmt(b.ends_at)}
                    </strong>
                    <span>{b.member_name}</span>
                    <small>{b.room_name}</small>
                  </Link>
                ))
              ) : (
                <p className="coach-day-empty">Chưa có buổi tập</p>
              )}
            </section>
          );
        })}
      </div>
      <Link to="/pt/history" className="back-link">
        Xem tất cả trạng thái đặt lịch →
      </Link>
    </>
  );
}
