import type { SessionRangeOption } from "./SessionRangePicker";
import { localDate } from "../../utils/vietnamTime";

export type WeekSession = SessionRangeOption & { session_date?: string };
export function addDays(date: string, amount: number) {
  const value = new Date(`${date}T12:00:00Z`);
  value.setUTCDate(value.getUTCDate() + amount);
  return value.toISOString().slice(0, 10);
}
export function weekStart(date: string) {
  return addDays(date, -((new Date(`${date}T12:00:00Z`).getUTCDay() + 6) % 7));
}
export function sessionDay(
  session: Pick<WeekSession, "starts_at" | "session_date">,
) {
  return session.session_date || localDate(new Date(session.starts_at));
}
export function minute(session: WeekSession) {
  const date = new Date(session.starts_at);
  return ((date.getUTCHours() + 7) % 24) * 60 + date.getUTCMinutes();
}
export function duration(session: WeekSession) {
  return (Date.parse(session.ends_at) - Date.parse(session.starts_at)) / 60000;
}
export function validateRange<T extends WeekSession>(values: T[]) {
  const sorted = [...values].sort(
    (a, b) => Date.parse(a.starts_at) - Date.parse(b.starts_at),
  );
  if (!sorted.length || sorted.length > 24)
    return "Chọn từ 1 đến 24 ca liền nhau.";
  if (sorted.some((item) => item.unavailable_reason))
    return "Khoảng giờ đi qua ca không thể đặt. Lựa chọn trước vẫn được giữ.";
  if (sorted.some((item) => sessionDay(item) !== sessionDay(sorted[0])))
    return "Chỉ chọn một khoảng giờ trong cùng ngày.";
  for (let i = 1; i < sorted.length; i++)
    if (
      sorted[i - 1].ends_at !== sorted[i].starts_at &&
      Date.parse(sorted[i - 1].ends_at) !== Date.parse(sorted[i].starts_at)
    )
      return "Các ca phải liền nhau, không có khoảng trống.";
  return "";
}
export function rangeBetween<T extends WeekSession>(
  sessions: T[],
  first: number,
  last: number,
): { values: T[]; error: string } {
  const a = sessions.find((s) => s.id === first),
    b = sessions.find((s) => s.id === last);
  if (!a || !b || sessionDay(a) !== sessionDay(b))
    return { values: [], error: "Chỉ chọn một khoảng giờ trong cùng ngày." };
  const day = sessions
    .filter((s) => sessionDay(s) === sessionDay(a))
    .sort((a, b) => Date.parse(a.starts_at) - Date.parse(b.starts_at));
  const left = day.findIndex((s) => s.id === first),
    right = day.findIndex((s) => s.id === last);
  const values = day.slice(Math.min(left, right), Math.max(left, right) + 1);
  return { values, error: validateRange(values) };
}
