import { useEffect, useRef, useState, type PointerEvent } from "react";
import {
  civilDate,
  localDate,
  vietnamTime,
  vietnamWeekday,
} from "../../utils/vietnamTime";
import {
  addDays,
  duration,
  minute,
  rangeBetween,
  sessionDay,
  validateRange,
  weekStart,
  type WeekSession,
} from "./weekly";
import "./week-calendar.css";

type Gesture = {
  kind: "create" | "move" | "start" | "end";
  anchor: number;
  last: number;
  x: number;
  y: number;
  moved: boolean;
};
export function WeekCalendar<T extends WeekSession>({
  sessions,
  selectedIDs,
  date,
  onDate,
  onSelect,
  onError,
  loading = false,
  onFindNext,
}: {
  sessions: T[];
  selectedIDs: number[];
  date: string;
  onDate: (date: string) => void;
  onSelect: (values: T[]) => void;
  onError: (message: string) => void;
  loading?: boolean;
  onFindNext?: () => Promise<void>;
}) {
  const [layout, setLayout] = useState<"week" | "day">("week");
  const [wide, setWide] = useState(
    () =>
      typeof matchMedia !== "undefined" &&
      matchMedia("(min-width:1100px)").matches,
  );
  useEffect(() => {
    if (typeof matchMedia === "undefined") return;
    const media = matchMedia("(min-width:1100px)");
    const update = () => setWide(media.matches);
    media.addEventListener("change", update);
    return () => media.removeEventListener("change", update);
  }, []);
  const [viewport, setViewport] = useState(0);
  const column = wide ? Math.max(96, (viewport - 56) / 7) : 184;
  const [mode, setMode] = useState<"view" | "select">(() =>
    typeof matchMedia !== "undefined" &&
    matchMedia("(max-width: 760px)").matches
      ? "view"
      : "select",
  );
  const [filter, setFilter] = useState("all"),
    [preview, setPreview] = useState<T[] | null>(null),
    [message, setMessage] = useState(""),
    [finding, setFinding] = useState(false);
  const scroll = useRef<HTMLDivElement>(null),
    gesture = useRef<Gesture | null>(null),
    frame = useRef<number | null>(null),
    suppress = useRef(false),
    anchor = useRef<number | null>(null);
  useEffect(() => {
    const element = scroll.current;
    if (!element || typeof ResizeObserver === "undefined") return;
    const update = () => setViewport(element.clientWidth);
    const observer = new ResizeObserver(update);
    observer.observe(element);
    update();
    return () => observer.disconnect();
  }, []);
  const from = weekStart(date),
    dates = Array.from({ length: 7 }, (_, i) => addDays(from, i));
  const all = [...sessions].sort(
    (a, b) => Date.parse(a.starts_at) - Date.parse(b.starts_at),
  );
  const selected = all.filter((s) => selectedIDs.includes(s.id)),
    shown = preview || selected;
  const inFilter = (s: T) =>
    filter === "all" ||
    (filter === "morning"
      ? minute(s) < 720
      : filter === "afternoon"
        ? minute(s) >= 720 && minute(s) < 1080
        : minute(s) >= 1080);
  const visible = all.filter(
    (s) => (layout === "week" || sessionDay(s) === date) && inFilter(s),
  );
  const min = visible.length
    ? Math.floor(Math.min(...visible.map(minute)) / 60) * 60
    : 360;
  const max = visible.length
    ? Math.max(
        min + 240,
        Math.ceil(
          Math.max(...visible.map((s) => minute(s) + duration(s))) / 60,
        ) * 60,
      )
    : 600;
  const scale = Math.max(
      1.5,
      ...visible.map((s) => 44 / Math.max(1, duration(s))),
    ),
    height = (max - min) * scale;
  const days = layout === "week" ? dates : [date];
  const currentDay = selected.length ? sessionDay(selected[0]) : date,
    daySessions = all.filter((s) => sessionDay(s) === currentDay);
  const chosenFirst = selected[0],
    chosenLast = selected.at(-1);
  function stop() {
    gesture.current = null;
    if (frame.current !== null) cancelAnimationFrame(frame.current);
    frame.current = null;
    setPreview(null);
  }
  useEffect(() => {
    if (gesture.current)
      setMessage(
        "Lịch đã thay đổi; thao tác kéo được hủy và giữ lựa chọn trước.",
      );
    stop();
    anchor.current = null;
  }, [
    from,
    layout,
    filter,
    sessions
      .map((s) => `${s.id}:${s.starts_at}:${s.ends_at}:${s.unavailable_reason}`)
      .join("|"),
  ]);
  useEffect(
    () => () => {
      if (frame.current !== null) cancelAnimationFrame(frame.current);
    },
    [],
  );
  useEffect(() => {
    if (layout !== "week" || !scroll.current) return;
    const col = scroll.current.querySelector<HTMLElement>(
      `[data-week-day="${date}"]`,
    );
    if (col) scroll.current.scrollLeft = Math.max(0, col.offsetLeft - 56);
  }, [date, layout, from, column]);
  function report(error: string) {
    setMessage(error);
    onError(error);
  }
  function commit(values: T[]) {
    const error = validateRange(values);
    if (error) {
      report(error);
      return;
    }
    report("");
    onSelect(values);
  }
  function at(x: number, y: number): T | undefined {
    const node = document
      .elementFromPoint?.(x, y)
      ?.closest<HTMLElement>("[data-session-id], [data-week-day]");
    const direct = Number(node?.dataset.sessionId);
    if (direct) return all.find((s) => s.id === direct);
    const col = node?.closest<HTMLElement>(".bk-col");
    if (!col) return;
    const rect = col.getBoundingClientRect();
    if (x < rect.left || x > rect.right || y < rect.top || y > rect.bottom)
      return;
    const m = min + (y - rect.top) / scale;
    return all.find(
      (s) =>
        sessionDay(s) === col.dataset.weekDay &&
        minute(s) <= m &&
        minute(s) + duration(s) > m,
    );
  }
  function movedRange(target: T, active: Gesture) {
    if (!chosenFirst || !chosenLast)
      return { values: [] as T[], error: "Chọn giờ trước khi chỉnh." };
    if (
      (active.kind === "start" &&
        Date.parse(target.starts_at) > Date.parse(chosenLast.starts_at)) ||
      (active.kind === "end" &&
        Date.parse(target.starts_at) < Date.parse(chosenFirst.starts_at))
    )
      return { values: [] as T[], error: "Giờ kết thúc phải sau giờ bắt đầu." };
    if (active.kind !== "move")
      return rangeBetween(
        all,
        active.kind === "start" ? target.id : chosenFirst.id,
        active.kind === "end" ? target.id : chosenLast.id,
      );
    const origin = all.find((s) => s.id === active.anchor);
    if (!origin || sessionDay(target) !== currentDay)
      return { values: [] as T[], error: "Chỉ di chuyển trong cùng ngày." };
    const start =
        Date.parse(chosenFirst.starts_at) +
        Date.parse(target.starts_at) -
        Date.parse(origin.starts_at),
      end =
        start +
        Date.parse(chosenLast.ends_at) -
        Date.parse(chosenFirst.starts_at);
    const first = all.find(
        (s) =>
          sessionDay(s) === currentDay && Date.parse(s.starts_at) === start,
      ),
      last = all.find(
        (s) => sessionDay(s) === currentDay && Date.parse(s.ends_at) === end,
      );
    return first && last
      ? rangeBetween(all, first.id, last.id)
      : {
          values: [] as T[],
          error: "Không có dải giờ liền nhau cùng thời lượng ở vị trí này.",
        };
  }
  function proposal(active: Gesture, target: T) {
    return active.kind === "create"
      ? rangeBetween(all, active.anchor, target.id)
      : movedRange(target, active);
  }
  function update(active: Gesture) {
    const target = at(active.x, active.y);
    if (!target) {
      setPreview(null);
      return;
    }
    const result = proposal(active, target);
    setPreview(result.error ? null : result.values);
  }
  function autoScroll() {
    if (frame.current !== null) return;
    const tick = () => {
      const active = gesture.current,
        e = scroll.current;
      if (!active || !e) {
        frame.current = null;
        return;
      }
      const rect = e.getBoundingClientRect();
      const delta =
        active.y < rect.top + 75 ? -10 : active.y > rect.bottom - 35 ? 10 : 0;
      if (delta) {
        e.scrollTop += delta;
        update(active);
      }
      frame.current = requestAnimationFrame(tick);
    };
    frame.current = requestAnimationFrame(tick);
  }
  function begin(
    event: PointerEvent<HTMLElement>,
    session: T,
    kind: Gesture["kind"] = "create",
  ) {
    if (
      mode !== "select" ||
      event.button !== 0 ||
      session.unavailable_reason ||
      loading
    )
      return;
    if (event.pointerType === "touch") event.preventDefault();
    event.stopPropagation();
    event.currentTarget.setPointerCapture?.(event.pointerId);
    const target =
      kind === "move" ? at(event.clientX, event.clientY) || session : session;
    gesture.current = {
      kind,
      anchor: target.id,
      last: target.id,
      x: event.clientX,
      y: event.clientY,
      moved: false,
    };
  }
  function move(event: PointerEvent<HTMLElement>) {
    const active = gesture.current;
    if (!active) return;
    if (
      Math.abs(active.x - event.clientX) + Math.abs(active.y - event.clientY) >
      3
    )
      active.moved = true;
    active.x = event.clientX;
    active.y = event.clientY;
    update(active);
    autoScroll();
  }
  function finish(event: PointerEvent<HTMLElement>) {
    const active = gesture.current;
    if (!active) return;
    const target = at(event.clientX, event.clientY);
    stop();
    if (!active.moved) return;
    suppress.current = true;
    setTimeout(() => {
      suppress.current = false;
    }, 0);
    anchor.current = null;
    if (!target) {
      report("Đã bỏ thao tác ngoài lịch; lựa chọn trước vẫn được giữ.");
      return;
    }
    const result = proposal(active, target);
    if (result.error) report(result.error);
    else commit(result.values);
  }
  function click(session: T) {
    if (suppress.current || loading) return;
    if (mode !== "select") {
      setMessage("Chuyển sang Chọn giờ để tạo khoảng tập.");
      return;
    }
    if (anchor.current === null) {
      commit([session]);
      anchor.current = session.id;
    } else {
      const result = rangeBetween(all, anchor.current, session.id);
      if (result.error) report(result.error);
      else {
        commit(result.values);
        anchor.current = null;
      }
    }
  }
  function keyAdjust(kind: Gesture["kind"], delta: number) {
    if (!chosenFirst || !chosenLast) return;
    const base = kind === "end" ? chosenLast : chosenFirst,
      index = daySessions.findIndex((s) => s.id === base.id),
      target = daySessions[index + delta];
    if (!target) {
      report("Đã tới biên của lịch trong ngày.");
      return;
    }
    const result = movedRange(target, {
      kind,
      anchor: base.id,
      last: base.id,
      x: 0,
      y: 0,
      moved: true,
    });
    if (result.error) report(result.error);
    else commit(result.values);
  }
  async function nearest() {
    setFinding(true);
    try {
      if (onFindNext) await onFindNext();
      else {
        const item = all.find(
          (s) => !s.unavailable_reason && Date.parse(s.starts_at) > Date.now(),
        );
        if (item) onDate(sessionDay(item));
        else
          setMessage(
            "Tuần này chưa có ca còn chỗ. Hãy chuyển sang tuần tiếp theo.",
          );
      }
    } catch (e) {
      report((e as Error).message);
    } finally {
      setFinding(false);
    }
  }
  const summary = shown.length
    ? `${vietnamTime(shown[0].starts_at)} – ${vietnamTime(shown.at(-1)!.ends_at)} · ${shown.length} ca · ${Math.round((Date.parse(shown.at(-1)!.ends_at) - Date.parse(shown[0].starts_at)) / 60000)} phút`
    : "Chưa chọn giờ";
  return (
    <div
      className={`bk-calendar bk-${mode} ${wide && viewport < 1000 ? "bk-compact" : ""}`}
      onKeyDown={(event) => {
        if (event.key === "Escape") {
          stop();
          anchor.current = null;
          report("Đã hủy thao tác đang kéo; giữ lựa chọn trước.");
        }
      }}
    >
      <div className="bk-toolbar">
        <div role="group" aria-label="Chế độ lịch">
          {(["week", "day"] as const).map((v) => (
            <button
              key={v}
              type="button"
              aria-pressed={layout === v}
              onClick={() => setLayout(v)}
            >
              {v === "week" ? "Tuần" : "Ngày"}
            </button>
          ))}
        </div>
        <div role="group" aria-label="Thao tác lịch">
          {(["view", "select"] as const).map((v) => (
            <button
              key={v}
              type="button"
              aria-pressed={mode === v}
              onClick={() => {
                stop();
                anchor.current = null;
                setMode(v);
              }}
            >
              {v === "view" ? "Xem lịch" : "Chọn giờ"}
            </button>
          ))}
        </div>
      </div>
      <div className="bk-tools">
        <label>
          Buổi tập
          <select value={filter} onChange={(e) => setFilter(e.target.value)}>
            <option value="all">Cả ngày</option>
            <option value="morning">Sáng · trước 12h</option>
            <option value="afternoon">Chiều · 12–18h</option>
            <option value="evening">Tối · sau 18h</option>
          </select>
        </label>
        <button
          type="button"
          disabled={finding || loading}
          onClick={() => void nearest()}
        >
          {finding ? "Đang tìm ngày…" : "Tìm ngày còn ca gần nhất"}
        </button>
      </div>
      <p className="bk-help">
        {mode === "view"
          ? "Vuốt hoặc cuộn trong lịch để xem đủ 7 ngày. Bật Chọn giờ khi muốn chọn buổi tập."
          : "Kéo qua các ca liền nhau hoặc chọn ca đầu rồi ca cuối. Kéo khối để di chuyển, hai tay nắm để đổi giờ; Escape hủy thao tác."}
      </p>
      <div
        ref={scroll}
        className="bk-scroll"
        role="region"
        aria-label="Lịch tập tuần, cuộn ngang và dọc"
        tabIndex={0}
        onPointerMove={move}
        onPointerUp={finish}
        onPointerCancel={() => {
          stop();
          setMessage("Thao tác đã hủy; giữ lựa chọn trước.");
        }}
      >
        <div
          className="bk-canvas"
          style={{ width: 56 + days.length * column, minWidth: "100%" }}
        >
          <div
            className="bk-head"
            style={{
              gridTemplateColumns: `56px repeat(${days.length},minmax(${column}px,1fr))`,
            }}
          >
            <div className="bk-corner">Giờ VN</div>
            {days.map((d) => (
              <button
                key={d}
                className={date === d ? "active" : ""}
                aria-pressed={date === d}
                onClick={() => onDate(d)}
              >
                <span>
                  {vietnamWeekday(new Date(`${d}T12:00:00+07:00`))}
                  {d === localDate() ? " · Hôm nay" : ""}
                </span>
                <strong>{civilDate(d).slice(0, 5)}</strong>
              </button>
            ))}
          </div>
          <div className="bk-body" style={{ height }}>
            <div className="bk-axis">
              {Array.from({ length: Math.ceil((max - min) / 60) }, (_, i) => (
                <span key={i} style={{ top: i * 60 * scale }}>
                  {String(Math.floor((min + i * 60) / 60) % 24).padStart(
                    2,
                    "0",
                  )}
                  :00
                </span>
              ))}
            </div>
            <div
              className="bk-days"
              style={{
                gridTemplateColumns: `repeat(${days.length},minmax(${column}px,1fr))`,
              }}
            >
              {days.map((d) => {
                const slots = visible.filter((s) => sessionDay(s) === d),
                  block = shown.filter((s) => sessionDay(s) === d),
                  first = block[0],
                  last = block.at(-1),
                  blockVisible =
                    first &&
                    last &&
                    minute(first) >= min &&
                    minute(last) + duration(last) <= max &&
                    inFilter(first) &&
                    inFilter(last);
                return (
                  <section
                    key={d}
                    className="bk-col"
                    data-week-day={d}
                    style={{ backgroundSize: `100% ${60 * scale}px` }}
                    aria-label={civilDate(d)}
                  >
                    {!slots.length && (
                      <p className="bk-empty">
                        {loading
                          ? "Đang tải lịch…"
                          : "Chưa có ca trong khung này."}
                      </p>
                    )}
                    {slots.map((s) => (
                      <button
                        type="button"
                        key={s.id}
                        className={`bk-slot ${s.unavailable_reason ? "unavailable" : s.remaining <= 2 ? "low" : ""}`}
                        data-session-id={s.id}
                        aria-label={`${vietnamTime(s.starts_at)} – ${vietnamTime(s.ends_at)} · ${civilDate(d)} · ${s.unavailable_reason || `Còn ${s.remaining} chỗ`}`}
                        aria-pressed={selectedIDs.includes(s.id)}
                        disabled={!!s.unavailable_reason || loading}
                        tabIndex={mode === "view" ? -1 : 0}
                        style={{
                          top: (minute(s) - min) * scale,
                          height: duration(s) * scale - 3,
                        }}
                        onPointerDown={(e) => begin(e, s)}
                        onClick={() => click(s)}
                      >
                        <strong>
                          {vietnamTime(s.starts_at)} – {vietnamTime(s.ends_at)}
                        </strong>
                        <span>
                          {s.unavailable_reason || `Còn ${s.remaining} chỗ`}
                        </span>
                      </button>
                    ))}
                    {blockVisible && (
                      <div
                        className={`bk-selection ${preview ? "preview" : ""}`}
                        style={{
                          top: (minute(first) - min) * scale,
                          height:
                            (minute(last) + duration(last) - minute(first)) *
                              scale -
                            3,
                        }}
                      >
                        <button
                          type="button"
                          className="bk-handle start"
                          aria-label="Đổi giờ bắt đầu; dùng mũi tên lên xuống"
                          onPointerDown={(e) => begin(e, first, "start")}
                          onKeyDown={(e) => {
                            if (e.key === "ArrowUp" || e.key === "ArrowDown") {
                              e.preventDefault();
                              keyAdjust("start", e.key === "ArrowUp" ? -1 : 1);
                            }
                          }}
                        >
                          <span />
                        </button>
                        <button
                          type="button"
                          className="bk-block-body"
                          aria-label={`Di chuyển khoảng giờ ${summary}; dùng mũi tên lên xuống`}
                          onPointerDown={(e) => begin(e, first, "move")}
                          onKeyDown={(e) => {
                            if (e.key === "ArrowUp" || e.key === "ArrowDown") {
                              e.preventDefault();
                              keyAdjust("move", e.key === "ArrowUp" ? -1 : 1);
                            }
                          }}
                        >
                          <strong>
                            {vietnamTime(first.starts_at)} –{" "}
                            {vietnamTime(last.ends_at)}
                          </strong>
                          <span>{block.length} ca · Đã chọn</span>
                        </button>
                        <button
                          type="button"
                          className="bk-handle end"
                          aria-label="Đổi giờ kết thúc; dùng mũi tên lên xuống"
                          onPointerDown={(e) => begin(e, last, "end")}
                          onKeyDown={(e) => {
                            if (e.key === "ArrowUp" || e.key === "ArrowDown") {
                              e.preventDefault();
                              keyAdjust("end", e.key === "ArrowUp" ? -1 : 1);
                            }
                          }}
                        >
                          <span />
                        </button>
                      </div>
                    )}
                  </section>
                );
              })}
            </div>
          </div>
        </div>
      </div>
      <div className="bk-legend">
        <span>● Còn chỗ</span>
        <span>◐ Sắp đầy</span>
        <span>▧ Không thể đặt</span>
        <span>■ Khoảng đã chọn</span>
      </div>
      <p
        className="bk-live"
        role="status"
        aria-live="polite"
        aria-atomic="true"
      >
        {summary}
        {message ? ` · ${message}` : ""}
      </p>
      {chosenFirst && chosenLast && (
        <div className="bk-adjust">
          <label>
            Bắt đầu
            <select
              value={chosenFirst.id}
              onChange={(e) => {
                const s = daySessions.find(
                  (s) => s.id === Number(e.target.value),
                );
                if (!s) return;
                if (
                  Date.parse(s.starts_at) > Date.parse(chosenLast.starts_at)
                ) {
                  report("Giờ bắt đầu phải trước giờ kết thúc.");
                  return;
                }
                const r = rangeBetween(all, s.id, chosenLast.id);
                r.error ? report(r.error) : commit(r.values);
              }}
            >
              {daySessions.map((s) => (
                <option
                  key={s.id}
                  value={s.id}
                  disabled={!!s.unavailable_reason}
                >
                  {vietnamTime(s.starts_at)}
                </option>
              ))}
            </select>
          </label>
          <label>
            Kết thúc
            <select
              value={chosenLast.id}
              onChange={(e) => {
                const s = daySessions.find(
                  (s) => s.id === Number(e.target.value),
                );
                if (!s) return;
                if (
                  Date.parse(s.starts_at) < Date.parse(chosenFirst.starts_at)
                ) {
                  report("Giờ kết thúc phải sau giờ bắt đầu.");
                  return;
                }
                const r = rangeBetween(all, chosenFirst.id, s.id);
                r.error ? report(r.error) : commit(r.values);
              }}
            >
              {daySessions.map((s) => (
                <option
                  key={s.id}
                  value={s.id}
                  disabled={!!s.unavailable_reason}
                >
                  {vietnamTime(s.ends_at)}
                </option>
              ))}
            </select>
          </label>
          <button
            type="button"
            onClick={() => {
              stop();
              anchor.current = null;
              report("");
              onSelect([]);
            }}
          >
            Xóa lựa chọn
          </button>
        </div>
      )}
      {selected.length > 0 && !selected.some(inFilter) && (
        <p className="bk-help">
          Khoảng đã chọn nằm ngoài bộ lọc đang xem và vẫn được giữ.
        </p>
      )}
    </div>
  );
}
