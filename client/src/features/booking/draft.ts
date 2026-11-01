import { isCivilDate } from "../../utils/vietnamTime";
import { sessionDay, validateRange, type WeekSession } from "./weekly";
const key = "gym.booking-draft.v1";
export const DRAFT_TTL = 30 * 60 * 1000;
export type BookingDraft = {
  version: 1;
  roomID: number;
  date: string;
  sessions: { id: number; starts_at: string; ends_at: string }[];
  ptID: number;
  wantsPT: boolean;
  owner: number | null;
  savedAt: number;
};
export function readDraft(): BookingDraft | null {
  try {
    const d = JSON.parse(
      sessionStorage.getItem(key) || "null",
    ) as BookingDraft | null;
    if (!d) return null;
    if (
      d.version !== 1 ||
      !Number.isSafeInteger(d.roomID) ||
      d.roomID < 1 ||
      !isCivilDate(d.date) ||
      !Array.isArray(d.sessions) ||
      !d.sessions.length ||
      d.sessions.length > 24 ||
      !d.sessions.every(
        (s) =>
          Number.isSafeInteger(s.id) &&
          s.id > 0 &&
          Number.isFinite(Date.parse(s.starts_at)) &&
          Number.isFinite(Date.parse(s.ends_at)),
      ) ||
      !Number.isFinite(d.savedAt) ||
      Date.now() - d.savedAt > DRAFT_TTL ||
      d.savedAt > Date.now() + 60000 ||
      !(d.owner === null || Number.isSafeInteger(d.owner)) ||
      !Number.isSafeInteger(d.ptID) ||
      typeof d.wantsPT !== "boolean"
    ) {
      clearDraft();
      return null;
    }
    return d;
  } catch {
    clearDraft();
    return null;
  }
}
export function saveDraft(
  roomID: number,
  values: WeekSession[],
  owner: number | null,
  ptID = 0,
  wantsPT = false,
) {
  if (!values.length) {
    clearDraft(roomID);
    return;
  }
  const d: BookingDraft = {
    version: 1,
    roomID,
    date: sessionDay(values[0]),
    sessions: values.map(({ id, starts_at, ends_at }) => ({
      id,
      starts_at,
      ends_at,
    })),
    ptID,
    wantsPT,
    owner,
    savedAt: Date.now(),
  };
  try {
    sessionStorage.setItem(key, JSON.stringify(d));
  } catch {
    /* Booking remains usable when browser storage is disabled. */
  }
}
export function clearDraft(roomID?: number) {
  try {
    if (roomID !== undefined) {
      const d = JSON.parse(sessionStorage.getItem(key) || "null");
      if (d?.roomID !== roomID) return;
    }
    sessionStorage.removeItem(key);
  } catch {
    /* no persistent credentials */
  }
}
export function restoreDraft<T extends WeekSession>(
  draft: BookingDraft,
  roomID: number,
  userID: number,
  sessions: T[],
) {
  if (draft.roomID !== roomID) return { values: [] as T[], error: "" };
  if (draft.owner !== null && draft.owner !== userID)
    return {
      values: [] as T[],
      error:
        "Bản nháp thuộc tài khoản khác và đã được bỏ. Hãy chọn giờ cho tài khoản hiện tại.",
    };
  const values = draft.sessions
    .map((old) =>
      sessions.find(
        (s) =>
          s.id === old.id &&
          Date.parse(s.starts_at) === Date.parse(old.starts_at) &&
          Date.parse(s.ends_at) === Date.parse(old.ends_at),
      ),
    )
    .filter((s): s is T => !!s);
  if (values.length !== draft.sessions.length)
    return {
      values: [] as T[],
      error: "Lịch trong bản nháp đã thay đổi. Hãy chọn lại giờ phù hợp.",
    };
  const error = validateRange(values);
  return {
    values: error ? [] : values,
    error: error
      ? "Bản nháp không còn đủ điều kiện hoặc chỗ. Hãy chọn lại giờ."
      : "",
  };
}
