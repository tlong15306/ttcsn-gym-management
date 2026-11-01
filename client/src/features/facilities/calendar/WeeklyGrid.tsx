import { useRef, useState, type KeyboardEvent, type PointerEvent } from "react";
import {
  blockAt,
  moveBlock,
  paint,
  cloneWeek,
  weekdays,
  type Week,
} from "./hours";
export type GridMode = "OPEN" | "CLOSE" | "MOVE" | "START" | "END";
type Drag = { from: number[]; to: number[]; span: [number, number] | null };
export function WeeklyGrid({
  week,
  onChange,
  mode,
  disabled,
}: {
  week: Week;
  onChange: (week: Week) => void;
  mode: GridMode;
  disabled?: boolean;
}) {
  const [focus, setFocus] = useState(0),
    [draft, setDraft] = useState<Week | null>(null),
    [tapMode, setTapMode] = useState(false),
    [tapAnchor, setTapAnchor] = useState<[number, number] | null>(null);
  const drag = useRef<Drag | null>(null),
    anchor = useRef<number[] | null>(null);
  const grid = useRef<HTMLDivElement>(null);
  function apply(
    from: number[],
    to: number[],
    span: [number, number] | null,
  ): Week {
    if (mode === "MOVE" && span)
      return moveBlock(week, from[0], span, to[0], span[0] + to[1] - from[1]);
    if ((mode === "START" || mode === "END") && span) {
      const copy = cloneWeek(week);
      for (let h = span[0]; h < span[1]; h++) copy[from[0]][h] = false;
      const start = mode === "START" ? Math.min(to[1], span[1] - 1) : span[0];
      const end = mode === "END" ? Math.max(to[1] + 1, span[0] + 1) : span[1];
      for (let h = start; h < end; h++) copy[from[0]][h] = true;
      return copy;
    }
    if (mode !== "OPEN" && mode !== "CLOSE") return week;
    return paint(week, from, to, mode === "OPEN");
  }
  function down(e: PointerEvent<HTMLButtonElement>, d: number, h: number) {
    if (disabled || e.button !== 0) return;
    const span = blockAt(week, d, h);
    if (mode !== "OPEN" && mode !== "CLOSE" && !span && !(tapMode && tapAnchor))
      return;
    e.preventDefault();
    e.currentTarget.focus();
    setFocus(d * 24 + h);
    anchor.current = null;
    if (tapMode) {
      if (tapAnchor) {
        onChange(apply(tapAnchor, [d, h], blockAt(week, tapAnchor[0], tapAnchor[1])));
        setTapAnchor(null);
      } else {
        setTapAnchor([d, h]);
      }
      setDraft(null);
      return;
    }
    drag.current = { from: [d, h], to: [d, h], span };
    grid.current?.setPointerCapture(e.pointerId);
    setDraft(apply([d, h], [d, h], span));
  }
  function move(e: PointerEvent<HTMLDivElement>) {
    if (!drag.current) return;
    const cell = document
      .elementFromPoint(e.clientX, e.clientY)
      ?.closest<HTMLElement>("[data-calendar-cell]");
    if (!cell || !grid.current?.contains(cell)) return;
    const to = [Number(cell.dataset.day), Number(cell.dataset.hour)];
    drag.current.to = to;
    setDraft(apply(drag.current.from, to, drag.current.span));
  }
  function finish(e: PointerEvent<HTMLDivElement>) {
    if (!drag.current) return;
    const v = apply(drag.current.from, drag.current.to, drag.current.span);
    drag.current = null;
    setDraft(null);
    onChange(v);
    if (grid.current?.hasPointerCapture(e.pointerId))
      grid.current.releasePointerCapture(e.pointerId);
  }
  function key(e: KeyboardEvent<HTMLButtonElement>, d: number, h: number) {
    if (disabled) return;
    if (e.key === "Escape") {
      drag.current = null;
      anchor.current = null;
      setDraft(null);
      return;
    }
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      const from = anchor.current ?? [d, h];
      onChange(apply(from, [d, h], blockAt(week, from[0], from[1])));
      anchor.current = null;
      setDraft(null);
      return;
    }
    const shifts: Record<string, number[]> = {
      ArrowLeft: [d, Math.max(0, h - 1)],
      ArrowRight: [d, Math.min(23, h + 1)],
      ArrowUp: [Math.max(0, d - 1), h],
      ArrowDown: [Math.min(6, d + 1), h],
      Home: [d, 0],
      End: [d, 23],
    };
    const next = shifts[e.key];
    if (!next) return;
    e.preventDefault();
    if (e.shiftKey) {
      anchor.current ??= [d, h];
      setDraft(
        apply(
          anchor.current,
          next,
          blockAt(week, anchor.current[0], anchor.current[1]),
        ),
      );
    } else {
      anchor.current = null;
      setDraft(null);
    }
    setFocus(next[0] * 24 + next[1]);
    grid.current
      ?.querySelector<HTMLButtonElement>(
        `[data-day="${next[0]}"][data-hour="${next[1]}"]`,
      )
      ?.focus();
  }
  const shown = draft ?? week;
  return (
    <div
      className="calendar-grid-scroll"
      tabIndex={0}
      aria-label="Cuộn ngang bảng giờ mở cửa"
    >
      <div className="calendar-grid-controls">
        <button
          type="button"
          className={`calendar-grid-tap-mode ${tapMode ? "is-active" : ""}`}
          aria-pressed={tapMode}
          disabled={disabled}
          onClick={() => {
            setTapMode((value) => !value);
            setTapAnchor(null);
            setDraft(null);
          }}
        >
          Chọn bằng chạm
        </button>
        <span aria-live="polite">
          {tapAnchor
            ? `Đã chọn ${weekdays[tapAnchor[0]]} ${String(tapAnchor[1]).padStart(2, "0")}:00. Chạm điểm kết thúc.`
            : tapMode
              ? "Chạm điểm bắt đầu rồi chạm điểm kết thúc."
              : "Kéo bằng chuột hoặc trackpad; trên cảm ứng dùng Chọn bằng chạm. Bàn phím: giữ Shift, dùng mũi tên rồi Enter."}
        </span>
      </div>
      <div
        ref={grid}
        className="calendar-grid"
        role="grid"
        aria-label="Lịch tuần, 7 ngày và 24 giờ"
        aria-rowcount={8}
        aria-colcount={25}
        onPointerMove={move}
        onPointerUp={finish}
        onPointerCancel={() => {
          drag.current = null;
          setTapAnchor(null);
          setDraft(null);
        }}
        onKeyDown={(e) => {
          if (e.key === "Escape") {
            drag.current = null;
            setDraft(null);
          }
        }}
      >
        <div role="row" className="calendar-grid-row calendar-grid-head">
          <span role="columnheader">Ngày / giờ</span>
          {Array.from({ length: 24 }, (_, h) => (
            <span role="columnheader" key={h}>
              {String(h).padStart(2, "0")}
            </span>
          ))}
        </div>
        {weekdays.map((day, d) => (
          <div role="row" className="calendar-grid-row" key={day}>
            <span role="rowheader" className="calendar-day-label">
              {day}
            </span>
            {shown[d].map((open, h) => (
              <button
                type="button"
                role="gridcell"
                key={h}
                data-calendar-cell
                data-day={d}
                data-hour={h}
                tabIndex={focus === d * 24 + h ? 0 : -1}
                aria-selected={open}
                disabled={disabled}
                aria-label={`${day}, ${h}:00–${h + 1}:00, ${open ? "mở cửa" : "đóng cửa"}`}
                className={`calendar-cell ${open ? "is-open" : ""} ${draft && open !== week[d][h] ? "is-draft" : ""} ${tapAnchor?.[0] === d && tapAnchor[1] === h ? "is-anchor" : ""}`}
                onPointerDown={(e) => down(e, d, h)}
                onKeyDown={(e) => key(e, d, h)}
                onFocus={() => setFocus(d * 24 + h)}
              >
                {open ? (
                  <span aria-hidden="true">
                    {h === 0 || !shown[d][h - 1]
                      ? `${String(h).padStart(2, "0")}`
                      : "·"}
                  </span>
                ) : (
                  <span aria-hidden="true">—</span>
                )}
              </button>
            ))}
          </div>
        ))}
      </div>
      <p className="calendar-grid-end">
        24:00 là điểm kết thúc ngày · mỗi ô tương đương 1 giờ
      </p>
    </div>
  );
}
