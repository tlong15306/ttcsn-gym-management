// Vietnam's current civil time is UTC+07:00. Date-only values are formatted
// separately from instants so they cannot be shifted into a different day.
const OFFSET = 7 * 60 * 60 * 1000;
const weekdays = ["CN", "Th 2", "Th 3", "Th 4", "Th 5", "Th 6", "Th 7"];
const pad = (value: number) => String(value).padStart(2, "0");
const shifted = (date: Date) => new Date(date.getTime() + OFFSET);

export function isCivilDate(value: string | null): value is string {
  if (!value || !/^\d{4}-\d{2}-\d{2}$/.test(value)) return false;
  const date = new Date(`${value}T00:00:00Z`);
  return !Number.isNaN(date.getTime()) && date.toISOString().slice(0, 10) === value;
}

export function localDate(date = new Date()) {
  const local = shifted(date);
  return `${local.getUTCFullYear()}-${pad(local.getUTCMonth() + 1)}-${pad(local.getUTCDate())}`;
}
export function civilDate(value: string) {
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(value);
  return match ? `${match[3]}/${match[2]}/${match[1]}` : "—";
}
export function vietnamTime(value: string) {
  const date = shifted(new Date(value));
  return `${pad(date.getUTCHours())}:${pad(date.getUTCMinutes())}`;
}
export function vietnamMinute(value: string) {
  const date = shifted(new Date(value));
  return date.getUTCHours() * 60 + date.getUTCMinutes();
}
export function vietnamTimestampDate(value: string) {
  return civilDate(localDate(new Date(value)));
}
export function vietnamDateTime(value: string) {
  return `${vietnamTime(value)} ${vietnamTimestampDate(value)}`;
}
export function vietnamLongDate(value: string) {
  const date = shifted(new Date(value));
  return `${date.getUTCDate()} tháng ${date.getUTCMonth() + 1}, ${date.getUTCFullYear()}`;
}
export function vietnamDay(value: string) {
  const date = shifted(new Date(value));
  return `${weekdays[date.getUTCDay()]}, ${pad(date.getUTCDate())}/${pad(date.getUTCMonth() + 1)}`;
}
export function vietnamWeekday(date: Date) {
  return weekdays[shifted(date).getUTCDay()];
}
