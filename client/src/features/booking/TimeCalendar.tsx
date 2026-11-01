import { Link } from "react-router-dom";
import { useEffect, useRef } from "react";
import { localDate, type Booking } from "./Booking";
import { vietnamDay, vietnamMinute as minute } from "../../utils/vietnamTime";
export function TimeCalendar({
  days,
  bookings,
}: {
  days: string[];
  bookings: Booking[];
}) {
  const scroll = useRef<HTMLDivElement>(null);
  useEffect(() => {
    if (scroll.current) scroll.current.scrollTop = 7 * 44;
  }, []);
  return (
    <div className="time-calendar" aria-label="Lịch tuần theo giờ">
      <div className="time-calendar-head">
        <span>Giờ</span>
        {days.map((d) => (
          <strong key={d}>
            {vietnamDay(`${d}T12:00:00+07:00`)}
          </strong>
        ))}
      </div>
      <div className="time-calendar-scroll" ref={scroll}>
        <div className="time-calendar-body">
          <div className="time-axis">
            {Array.from({ length: 24 }, (_, h) => (
              <span key={h} style={{ top: h * 44 }}>
                {String(h).padStart(2, "0")}:00
              </span>
            ))}
          </div>
          {days.map((d) => (
            <div className="time-day" key={d}>
              {bookings
                .filter(
                  (b) =>
                    b.status === "CONFIRMED" &&
                    localDate(new Date(b.starts_at)) === d,
                )
                .map((b) => (
                  <Link
                    key={b.id}
                    className="time-event"
                    to={`/pt/bookings/${b.id}`}
                    style={{
                      top: (minute(b.starts_at) / 60) * 44,
                      height: Math.max(
                        28,
                        ((new Date(b.ends_at).getTime() -
                          new Date(b.starts_at).getTime()) /
                          3600000) *
                          44,
                      ),
                    }}
                  >
                    <strong>{b.member_name}</strong>
                    <span>{b.room_name}</span>
                  </Link>
                ))}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
