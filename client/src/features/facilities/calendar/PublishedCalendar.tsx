import { useState } from "react";
import { Link, useParams } from "react-router-dom";
import {
  PageHeader,
  useList,
  ListState,
  Pagination,
  SearchBox,
} from "../../../components/shared/list";
import { useRecord, RecordState } from "../../../components/shared/record";
import { Field, Input } from "../../../components/ui";
import { localDate, clock, type Hours } from "./hours";
type Room = { id: number; name: string; room_type: string; status: string };
type Day = { date: string; source: string; hours: Hours[] };
const sources: Record<string, string> = {
  WEEKLY: "Lịch tuần",
  CLOSED: "Nghỉ cả ngày",
  REPLACE: "Giờ đặc biệt",
  UNCONFIGURED: "Chưa đăng ký lịch",
  MAINTENANCE: "Phòng bảo trì",
  INACTIVE: "Phòng ngừng hoạt động",
};
export function PublishedRoomsPage() {
  const list = useList<Room>("rooms");
  return (
    <>
      <PageHeader
        title="Lịch mở cửa phòng tập"
        description="Xem giờ hoạt động đã công bố và các ngày nghỉ của phòng."
      />
      <section className="panel calendar-panel">
        <SearchBox
          value={list.params.get("q") ?? ""}
          onChange={(v) => list.filter("q", v)}
          placeholder="Tìm phòng tập"
        />
        <ListState
          {...list}
          empty={list.data.length === 0}
          filtered={!!list.params.get("q")}
          retry={list.refresh}
        >
          {list.data.map((r) => (
            <div key={r.id} className="calendar-public-room">
              <span>
                <strong>{r.name}</strong>
                <small>{r.room_type}</small>
              </span>
              <Link
                className="btn btn-secondary"
                to={`/rooms/${r.id}/calendar`}
              >
                Xem lịch mở cửa
              </Link>
            </div>
          ))}
        </ListState>
        <Pagination meta={list.meta} change={list.filter} />
      </section>
    </>
  );
}
export function PublishedCalendarPage() {
  const { id } = useParams();
  const room = useRecord<Room>(`rooms/${id}`);
  const [from, setFrom] = useState(localDate());
  const until = new Date(`${from}T00:00:00Z`);
  until.setUTCDate(until.getUTCDate() + 6);
  const to = Number.isNaN(until.getTime())
    ? from
    : until.toISOString().slice(0, 10);
  const days = useRecord<Day[]>(`rooms/${id}/calendar?from=${from}&to=${to}`);
  return (
    <>
      <PageHeader
        title={`Lịch mở cửa · ${room.data?.name ?? "Phòng tập"}`}
        description="Lịch đã công bố. Giờ đặc biệt thay thế giờ mở cửa thường lệ của đúng ngày đó."
        back="/rooms"
      />
      <section className="panel calendar-panel">
        <Field label="Xem từ ngày">
          <Input
            type="date"
            value={from}
            required
            onChange={(e) => setFrom(e.target.value)}
          />
        </Field>
        {days.loading || days.error || !days.data ? (
          <RecordState {...days} retry={days.refresh} />
        ) : (
          <div className="calendar-day-list">
            {days.data.map((d) => (
              <div key={d.date}>
                <span>
                  <strong>{d.date}</strong>
                  <small>{sources[d.source]}</small>
                </span>
                <span>
                  {d.hours.length
                    ? d.hours.map((h) => (
                        <span
                          key={h.start_minute}
                          className="calendar-hour-pill"
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
      </section>
    </>
  );
}
