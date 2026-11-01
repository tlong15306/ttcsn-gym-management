import {
  useMemo,
  useRef,
  useState,
  type KeyboardEvent,
  type PointerEvent,
} from "react";
import { weekdays } from "../facilities/calendar/hours";

export type AvailabilityHour = {
  day_of_week: number;
  start_time: string;
  end_time: string;
};

const slotMinutes = 30;
const slotsPerDay = (24 * 60) / slotMinutes;

type SlotWeek = boolean[][];
type Point = [number, number];
type Drag = { from: Point; to: Point };

export type AvailabilityGridProps = {
  hours: AvailabilityHour[];
  onChange: (hours: AvailabilityHour[]) => void;
  disabled?: boolean;
};

function emptyWeek(): SlotWeek {
  return Array.from({ length: 7 }, () => Array(slotsPerDay).fill(false));
}

function minute(value: string): number | null {
  const match = /^(\d{2}):(\d{2})$/.exec(value);
  if (!match) return null;
  const hours = Number(match[1]);
  const minutes = Number(match[2]);
  if (hours === 24 && minutes === 0) return 24 * 60;
  if (hours > 23 || minutes > 59) return null;
  return hours * 60 + minutes;
}

function time(value: number): string {
  return `${String(Math.floor(value / 60)).padStart(2, "0")}:${String(
    value % 60,
  ).padStart(2, "0")}`;
}

function isGridAligned(hour: AvailabilityHour): boolean {
  const start = minute(hour.start_time);
  const end = minute(hour.end_time);
  return (
    start !== null &&
    end !== null &&
    start < end &&
    start % slotMinutes === 0 &&
    end % slotMinutes === 0
  );
}

function normalizedHour(hour: AvailabilityHour): AvailabilityHour {
  return {
    day_of_week: hour.day_of_week,
    start_time: hour.start_time,
    end_time: hour.end_time,
  };
}

export function splitAvailabilityHours(hours: AvailabilityHour[]) {
  const preciseDays = new Set<number>();
  for (const hour of hours)
    if (!isGridAligned(hour)) preciseDays.add(hour.day_of_week);

  const editable = hours.filter(
    (hour) =>
      hour.day_of_week >= 1 &&
      hour.day_of_week <= 7 &&
      !preciseDays.has(hour.day_of_week),
  );
  const preserved = hours.filter(
    (hour) =>
      hour.day_of_week < 1 ||
      hour.day_of_week > 7 ||
      preciseDays.has(hour.day_of_week),
  );
  const week = emptyWeek();
  for (const hour of editable) {
    const start = minute(hour.start_time)! / slotMinutes;
    const end = minute(hour.end_time)! / slotMinutes;
    for (let slot = start; slot < end; slot++) week[hour.day_of_week - 1][slot] = true;
  }
  return { week, preserved, preciseDays };
}

export function availabilityHoursFromWeek(week: SlotWeek): AvailabilityHour[] {
  const hours: AvailabilityHour[] = [];
  week.forEach((day, index) => {
    let start = -1;
    for (let slot = 0; slot <= slotsPerDay; slot++) {
      if (day[slot] && start < 0) start = slot;
      if (!day[slot] && start >= 0) {
        hours.push({
          day_of_week: index + 1,
          start_time: time(start * slotMinutes),
          end_time: time(slot * slotMinutes),
        });
        start = -1;
      }
    }
  });
  return hours;
}

function cloneWeek(week: SlotWeek): SlotWeek {
  return week.map((day) => [...day]);
}

function paint(week: SlotWeek, from: Point, to: Point, open: boolean): SlotWeek {
  const next = cloneWeek(week);
  for (let day = Math.min(from[0], to[0]); day <= Math.max(from[0], to[0]); day++)
    for (
      let slot = Math.min(from[1], to[1]);
      slot <= Math.max(from[1], to[1]);
      slot++
    )
      next[day][slot] = open;
  return next;
}

function label(day: number, slot: number, open: boolean) {
  const start = slot * slotMinutes;
  return `${weekdays[day]}, ${time(start)}–${time(start + slotMinutes)}, ${
    open ? "đã đăng ký làm việc" : "chưa đăng ký"
  }`;
}

export function AvailabilityGrid({
  hours,
  onChange,
  disabled,
}: AvailabilityGridProps) {
  const { week, preserved, preciseDays } = useMemo(
    () => splitAvailabilityHours(hours),
    [hours],
  );
  const [mode, setMode] = useState<"OPEN" | "CLOSE">("OPEN");
  const [focus, setFocus] = useState(0);
  const [draft, setDraft] = useState<SlotWeek | null>(null);
  const [tapMode, setTapMode] = useState(false);
  const [tapAnchor, setTapAnchor] = useState<Point | null>(null);
  const [keyAnchor, setKeyAnchor] = useState<Point | null>(null);
  const drag = useRef<Drag | null>(null);
  const grid = useRef<HTMLDivElement>(null);
  const shown = draft ?? week;

  function locked(day: number) {
    return preciseDays.has(day + 1);
  }

  function crossesLockedDay(fromDay: number, toDay: number) {
    for (let day = Math.min(fromDay, toDay); day <= Math.max(fromDay, toDay); day++)
      if (locked(day)) return true;
    return false;
  }

  function commit(next: SlotWeek) {
    onChange([
      ...availabilityHoursFromWeek(next),
      ...preserved.map(normalizedHour),
    ]);
  }

  function clearSelection() {
    drag.current = null;
    setDraft(null);
    setTapAnchor(null);
    setKeyAnchor(null);
  }

  function tap(day: number, slot: number) {
    const point: Point = [day, slot];
    if (!tapAnchor) {
      setTapAnchor(point);
      setDraft(null);
      return;
    }
    if (crossesLockedDay(tapAnchor[0], point[0])) return;
    commit(paint(week, tapAnchor, point, mode === "OPEN"));
    setTapAnchor(null);
    setDraft(null);
  }

  function down(
    event: PointerEvent<HTMLButtonElement>,
    day: number,
    slot: number,
  ) {
    if (disabled || event.button !== 0 || locked(day)) return;
    event.preventDefault();
    event.currentTarget.focus();
    setFocus(day * slotsPerDay + slot);
    setKeyAnchor(null);
    if (tapMode) {
      tap(day, slot);
      return;
    }
    drag.current = { from: [day, slot], to: [day, slot] };
    grid.current?.setPointerCapture(event.pointerId);
    setDraft(paint(week, [day, slot], [day, slot], mode === "OPEN"));
  }

  function move(event: PointerEvent<HTMLDivElement>) {
    if (!drag.current) return;
    const cell = document
      .elementFromPoint(event.clientX, event.clientY)
      ?.closest<HTMLElement>("[data-availability-cell]");
    if (!cell || !grid.current?.contains(cell)) return;
    const target: Point = [Number(cell.dataset.day), Number(cell.dataset.slot)];
    if (crossesLockedDay(drag.current.from[0], target[0])) return;
    drag.current.to = target;
    setDraft(paint(week, drag.current.from, target, mode === "OPEN"));
  }

  function finish(event: PointerEvent<HTMLDivElement>) {
    if (!drag.current) return;
    const next = paint(
      week,
      drag.current.from,
      drag.current.to,
      mode === "OPEN",
    );
    drag.current = null;
    setDraft(null);
    commit(next);
    if (grid.current?.hasPointerCapture(event.pointerId))
      grid.current.releasePointerCapture(event.pointerId);
  }

  function key(
    event: KeyboardEvent<HTMLButtonElement>,
    day: number,
    slot: number,
  ) {
    if (disabled || locked(day)) return;
    if (event.key === "Escape") {
      event.preventDefault();
      clearSelection();
      return;
    }
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      if (!keyAnchor) {
        setKeyAnchor([day, slot]);
        return;
      }
      commit(paint(week, keyAnchor, [day, slot], mode === "OPEN"));
      setKeyAnchor(null);
      setDraft(null);
      return;
    }
    const moves: Record<string, Point> = {
      ArrowLeft: [day, Math.max(0, slot - 1)],
      ArrowRight: [day, Math.min(slotsPerDay - 1, slot + 1)],
      ArrowUp: [Math.max(0, day - 1), slot],
      ArrowDown: [Math.min(6, day + 1), slot],
      Home: [day, 0],
      End: [day, slotsPerDay - 1],
    };
    const next = moves[event.key];
    if (!next || locked(next[0])) return;
    event.preventDefault();
    if (event.shiftKey) {
      const anchor = keyAnchor ?? [day, slot];
      setKeyAnchor(anchor);
      setDraft(paint(week, anchor, next, mode === "OPEN"));
    } else {
      setKeyAnchor(null);
      setDraft(null);
    }
    setFocus(next[0] * slotsPerDay + next[1]);
    grid.current
      ?.querySelector<HTMLButtonElement>(
        `[data-day="${next[0]}"][data-slot="${next[1]}"]`,
      )
      ?.focus();
  }

  return (
    <fieldset className="availability-grid-fieldset" disabled={disabled}>
      <legend>Lịch tuần kéo chọn theo nửa giờ</legend>
      <div className="availability-grid-toolbar">
        <div className="availability-grid-tools" aria-label="Công cụ lịch làm việc">
          <button
            type="button"
            className={`availability-grid-mode ${mode === "OPEN" ? "is-active" : ""}`}
            aria-pressed={mode === "OPEN"}
            onClick={() => setMode("OPEN")}
          >
            Thêm giờ
          </button>
          <button
            type="button"
            className={`availability-grid-mode ${mode === "CLOSE" ? "is-active" : ""}`}
            aria-pressed={mode === "CLOSE"}
            onClick={() => setMode("CLOSE")}
          >
            Bỏ giờ
          </button>
          <button
            type="button"
            className={`availability-grid-mode ${tapMode ? "is-active" : ""}`}
            aria-pressed={tapMode}
            onClick={() => {
              setTapMode((value) => !value);
              setTapAnchor(null);
              setDraft(null);
            }}
          >
            Chọn bằng chạm
          </button>
        </div>
        <p className="availability-grid-status" aria-live="polite">
          {tapAnchor
            ? `Đã chọn ${weekdays[tapAnchor[0]]} ${time(tapAnchor[1] * slotMinutes)}. Chạm điểm kết thúc.`
            : tapMode
              ? "Chạm điểm bắt đầu rồi chạm điểm kết thúc."
              : "Kéo bằng chuột hoặc trackpad; trên cảm ứng dùng Chọn bằng chạm. Giữ Shift và dùng phím mũi tên để chọn bằng bàn phím."}
        </p>
      </div>
      {preciseDays.size > 0 && (
        <p className="availability-grid-warning" role="status">
          {Array.from(preciseDays, (day) => weekdays[day - 1]).join(", ")} có
          mốc phút lẻ. Các ngày này giữ nguyên và chỉnh bằng trường giờ chi
          tiết bên dưới.
        </p>
      )}
      <div className="availability-grid-scroll" tabIndex={0} aria-label="Cuộn ngang lịch làm việc">
        <div
          ref={grid}
          className="availability-grid"
          role="grid"
          aria-label="Lịch làm việc, 7 ngày và từng nửa giờ"
          aria-rowcount={8}
          aria-colcount={slotsPerDay + 1}
          onPointerMove={move}
          onPointerUp={finish}
          onPointerCancel={clearSelection}
        >
          <div role="row" className="availability-grid-row availability-grid-head">
            <span role="columnheader">Ngày / giờ</span>
            {Array.from({ length: slotsPerDay }, (_, slot) => (
              <span role="columnheader" key={slot}>
                {slot % 2 === 0 ? String(slot / 2).padStart(2, "0") : "·"}
              </span>
            ))}
          </div>
          {weekdays.map((dayName, day) => (
            <div role="row" className="availability-grid-row" key={dayName}>
              <span role="rowheader" className="availability-day-label">
                {dayName}
                {locked(day) && <small>Mốc phút lẻ</small>}
              </span>
              {shown[day].map((open, slot) => (
                <button
                  type="button"
                  role="gridcell"
                  key={slot}
                  data-availability-cell
                  data-day={day}
                  data-slot={slot}
                  tabIndex={focus === day * slotsPerDay + slot ? 0 : -1}
                  aria-selected={open}
                  aria-label={label(day, slot, open)}
                  disabled={disabled || locked(day)}
                  className={`availability-cell ${open ? "is-open" : ""} ${
                    draft && open !== week[day][slot] ? "is-draft" : ""
                  } ${
                    tapAnchor?.[0] === day && tapAnchor[1] === slot
                      ? "is-anchor"
                      : ""
                  }`}
                  onPointerDown={(event) => down(event, day, slot)}
                  onKeyDown={(event) => key(event, day, slot)}
                  onFocus={() => setFocus(day * slotsPerDay + slot)}
                >
                  <span aria-hidden="true">
                    {open && slot % 2 === 0 ? String(slot / 2).padStart(2, "0") : ""}
                  </span>
                </button>
              ))}
            </div>
          ))}
        </div>
      </div>
      <p className="availability-grid-help">
        Mỗi ô là 30 phút. Kéo chỉ thay đổi bản nháp trong hộp thoại này; bấm
        “Lưu thay đổi” để gửi lịch mới.
      </p>
    </fieldset>
  );
}
