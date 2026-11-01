import { useEffect, useRef, useState, type PointerEvent } from "react";

export type SessionRangeOption = {
  id: number;
  starts_at: string;
  ends_at: string;
  remaining: number;
  unavailable_reason: string;
};

type RangeResult =
  | { sessions: SessionRangeOption[]; error: "" }
  | { sessions: []; error: string };

function ordered(sessions: SessionRangeOption[]) {
  return [...sessions].sort(
    (left, right) =>
      new Date(left.starts_at).getTime() - new Date(right.starts_at).getTime(),
  );
}

function contiguousRange(
  sessions: SessionRangeOption[],
  fromID: number,
  toID: number,
): RangeResult {
  const sorted = ordered(sessions);
  const from = sorted.findIndex((session) => session.id === fromID);
  const to = sorted.findIndex((session) => session.id === toID);
  if (from < 0 || to < 0) return { sessions: [], error: "Ca tập không còn trong lịch này." };
  const range = sorted.slice(Math.min(from, to), Math.max(from, to) + 1);
  if (range.length > 24) {
    return { sessions: [], error: "Mỗi lần chỉ có thể đăng ký tối đa 24 ca." };
  }
  if (range.some((session) => session.unavailable_reason)) {
    return {
      sessions: [],
      error: "Khoảng kéo đi qua ca không còn chỗ hoặc không thể đặt.",
    };
  }
  for (let index = 1; index < range.length; index += 1) {
    if (
      new Date(range[index - 1].ends_at).getTime() !==
      new Date(range[index].starts_at).getTime()
    ) {
      return {
        sessions: [],
        error: "Chỉ có thể chọn các ca liền kề, không có khoảng trống.",
      };
    }
  }
  return { sessions: range, error: "" };
}

type Props = {
  allowRange?: boolean;
  formatTime: (value: string) => string;
  onRangeError: (message: string) => void;
  onSelect: (sessions: SessionRangeOption[]) => void;
  selectedIDs: number[];
  sessions: SessionRangeOption[];
};

/**
 * A drag-friendly, but fully button-based, range picker. A regular click (or
 * Enter/Space) selects the start then the end, so range selection does not
 * require drag gestures on touch devices or with a keyboard.
 */
export function SessionRangePicker({
  allowRange = true,
  formatTime,
  onRangeError,
  onSelect,
  selectedIDs,
  sessions,
}: Props) {
  const [anchorID, setAnchorID] = useState<number>();
  const [previewIDs, setPreviewIDs] = useState<number[]>([]);
  const pointer = useRef<
    { fromID: number; lastID: number; moved: boolean } | undefined
  >(undefined);
  const suppressClick = useRef(false);
  const scrollFrame = useRef<number | null>(null);
  const scrollVelocity = useRef(0);
  const sorted = ordered(sessions);
  const selected = new Set(selectedIDs);
  const preview = new Set(previewIDs);

  useEffect(() => {
    setAnchorID(undefined);
    setPreviewIDs([]);
    pointer.current = undefined;
  }, [sessions.map((session) => session.id).join(",")]);

  useEffect(() => () => {
    if (scrollFrame.current !== null) window.cancelAnimationFrame(scrollFrame.current);
  }, []);

  function stopAutoScroll() {
    scrollVelocity.current = 0;
    if (scrollFrame.current !== null) window.cancelAnimationFrame(scrollFrame.current);
    scrollFrame.current = null;
  }

  function updateAutoScroll(event: PointerEvent<HTMLButtonElement>) {
    if (event.pointerType !== "mouse") return;
    scrollVelocity.current = event.clientY < 48 ? -12 : event.clientY > window.innerHeight - 48 ? 12 : 0;
    if (!scrollVelocity.current) { stopAutoScroll(); return; }
    if (scrollFrame.current !== null) return;
    const scroll = () => {
      if (!pointer.current || !scrollVelocity.current) { scrollFrame.current = null; return; }
      window.scrollBy(0, scrollVelocity.current);
      scrollFrame.current = window.requestAnimationFrame(scroll);
    };
    scrollFrame.current = window.requestAnimationFrame(scroll);
  }

  function choose(fromID: number, toID: number) {
    const result = contiguousRange(sorted, fromID, toID);
    if (result.error) {
      onRangeError(result.error);
      return false;
    }
    onRangeError("");
    onSelect(result.sessions);
    return true;
  }

  function selectByClick(session: SessionRangeOption) {
    if (!allowRange) {
      onRangeError("");
      onSelect([session]);
      return;
    }
    if (anchorID === undefined) {
      onRangeError("");
      onSelect([session]);
      setAnchorID(session.id);
      return;
    }
    if (choose(anchorID, session.id)) setAnchorID(undefined);
  }

  function optionAt(event: PointerEvent<HTMLButtonElement>) {
    const target = document.elementFromPoint(event.clientX, event.clientY);
    const node = target?.closest<HTMLButtonElement>("[data-session-id]");
    const id = Number(node?.dataset.sessionId);
    return sorted.find((session) => session.id === id);
  }

  function beginDrag(event: PointerEvent<HTMLButtonElement>, session: SessionRangeOption) {
    if (!allowRange || session.unavailable_reason || (event.pointerType === "mouse" && event.button !== 0)) return;
    event.currentTarget.setPointerCapture?.(event.pointerId);
    pointer.current = { fromID: session.id, lastID: session.id, moved: false };
    setPreviewIDs([]);
  }

  function previewDrag(event: PointerEvent<HTMLButtonElement>) {
    const active = pointer.current;
    if (!active) return;
    updateAutoScroll(event);
    const target = optionAt(event);
    if (!target || target.id === active.lastID) return;
    active.lastID = target.id;
    active.moved = true;
    const result = contiguousRange(sorted, active.fromID, target.id);
    setPreviewIDs(result.sessions.map((session) => session.id));
  }

  function finishDrag(event: PointerEvent<HTMLButtonElement>) {
    const active = pointer.current;
    pointer.current = undefined;
    stopAutoScroll();
    setPreviewIDs([]);
    if (!active || !active.moved) return;
    event.currentTarget.releasePointerCapture?.(event.pointerId);
    suppressClick.current = true;
    window.setTimeout(() => {
      suppressClick.current = false;
    }, 0);
    const target = optionAt(event);
    if (target && choose(active.fromID, target.id)) setAnchorID(undefined);
  }

  const selectedText = selected.size
    ? `Đã chọn ${selected.size} ca${selected.size > 1 ? " liền kề" : ""}.`
    : "Chưa chọn ca tập.";

  return (
    <div className="session-range-picker" onKeyDown={(event) => {
      if (event.key !== "Escape") return;
      pointer.current = undefined;
      stopAutoScroll();
      setPreviewIDs([]);
      setAnchorID(undefined);
      onRangeError("");
    }}>
      <p className="session-range-help" id="session-range-help">
        {allowRange
          ? "Kéo từ ca bắt đầu tới ca kết thúc để chọn giờ liên tiếp. Trên điện thoại hoặc bàn phím, chọn ca đầu rồi chọn ca cuối; nhấn Escape để bỏ thao tác đang chọn."
          : "Chọn một ca tập cho hội viên."}
      </p>
      <p className="sr-only" aria-live="polite">
        {selectedText}
      </p>
      <div className="session-grid" role="group" aria-describedby="session-range-help" aria-label="Các ca tập">
        {sorted.map((session) => {
          const chosen = selected.has(session.id);
          const isPreview = preview.has(session.id);
          return (
            <button
              key={session.id}
              type="button"
              data-session-id={session.id}
              className={`session-option ${chosen ? "selected" : ""} ${isPreview ? "range-preview" : ""}`}
              disabled={!!session.unavailable_reason}
              aria-pressed={chosen}
              onPointerDown={(event) => beginDrag(event, session)}
              onPointerMove={previewDrag}
              onPointerUp={finishDrag}
              onPointerCancel={() => {
                pointer.current = undefined;
                stopAutoScroll();
                setPreviewIDs([]);
              }}
              onClick={() => {
                if (suppressClick.current) {
                  suppressClick.current = false;
                  return;
                }
                selectByClick(session);
              }}
            >
              <strong>
                {formatTime(session.starts_at)} – {formatTime(session.ends_at)}
              </strong>
              <span>
                {session.unavailable_reason || `${session.remaining} chỗ còn lại`}
              </span>
            </button>
          );
        })}
      </div>
    </div>
  );
}
