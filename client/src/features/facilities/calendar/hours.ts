export type Hours = {
  weekday: number;
  start_minute: number;
  end_minute: number;
};
export type Week = boolean[][];
export const weekdays = [
  "Thứ Hai",
  "Thứ Ba",
  "Thứ Tư",
  "Thứ Năm",
  "Thứ Sáu",
  "Thứ Bảy",
  "Chủ nhật",
];
export const emptyWeek = (): Week =>
  Array.from({ length: 7 }, () => Array(24).fill(false));
export const cloneWeek = (week: Week): Week => week.map((row) => [...row]);
export function toWeek(hours: Hours[]): Week {
  const week = emptyWeek();
  for (const h of hours)
    for (let c = h.start_minute / 60; c < h.end_minute / 60; c++)
      if (h.weekday >= 1 && h.weekday <= 7) week[h.weekday - 1][c] = true;
  return week;
}
export function toHours(week: Week): Hours[] {
  const hours: Hours[] = [];
  week.forEach((row, day) => {
    let start = -1;
    for (let h = 0; h <= 24; h++) {
      if (row[h] && start < 0) start = h;
      if (!row[h] && start >= 0) {
        hours.push({
          weekday: day + 1,
          start_minute: start * 60,
          end_minute: h * 60,
        });
        start = -1;
      }
    }
  });
  return hours;
}
export function paint(
  week: Week,
  from: number[],
  to: number[],
  open: boolean,
): Week {
  const copy = cloneWeek(week);
  for (let d = Math.min(from[0], to[0]); d <= Math.max(from[0], to[0]); d++)
    for (let h = Math.min(from[1], to[1]); h <= Math.max(from[1], to[1]); h++)
      copy[d][h] = open;
  return copy;
}
export function blockAt(
  week: Week,
  day: number,
  hour: number,
): [number, number] | null {
  if (!week[day][hour]) return null;
  let start = hour,
    end = hour + 1;
  while (start > 0 && week[day][start - 1]) start--;
  while (end < 24 && week[day][end]) end++;
  return [start, end];
}
export function moveBlock(
  week: Week,
  fromDay: number,
  span: [number, number],
  day: number,
  start: number,
): Week {
  const copy = cloneWeek(week),
    width = span[1] - span[0];
  const dest = Math.min(24 - width, Math.max(0, start));
  for (let h = span[0]; h < span[1]; h++) copy[fromDay][h] = false;
  for (let h = dest; h < dest + width; h++) copy[day][h] = true;
  return copy;
}
export const clock = (minute: number) =>
  `${String(Math.floor(minute / 60)).padStart(2, "0")}:${String(minute % 60).padStart(2, "0")}`;
export function localDate(offset = 0): string {
  const date = new Date(Date.now() + 7 * 3600000 + offset * 86400000);
  return date.toISOString().slice(0, 10);
}
