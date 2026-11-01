import { Button } from "../../components/ui";
import { localDate, vietnamWeekday } from "./vietnamTime";
export function DateStrip({
  date,
  change,
}: {
  date: string;
  change: (v: string) => void;
}) {
  const selected = new Date(`${date}T12:00:00+07:00`);
  const monday = new Date(selected);
  monday.setUTCDate(monday.getUTCDate() - ((monday.getUTCDay() + 6) % 7));
  function shift(n: number) {
    const d = new Date(selected);
    d.setUTCDate(d.getUTCDate() + n);
    const value = localDate(d);
    change(value < localDate() ? localDate() : value);
  }
  return (
    <div className="date-strip">
      <Button
        variant="secondary"
        aria-label="Tuần trước"
        onClick={() => shift(-7)}
      >
        ←
      </Button>
      <div className="date-strip-days">
        {Array.from({ length: 7 }, (_, i) => {
          const d = new Date(monday);
          d.setUTCDate(d.getUTCDate() + i);
          const value = localDate(d);
          return (
            <button
              key={value}
              disabled={value < localDate()}
              aria-pressed={date === value}
              className={date === value ? "selected" : ""}
              onClick={() => change(value)}
            >
              <span>
                {vietnamWeekday(d)}
              </span>
              <strong>
                {value.slice(-2)}
              </strong>
            </button>
          );
        })}
      </div>
      <Button
        variant="secondary"
        aria-label="Tuần sau"
        onClick={() => shift(7)}
      >
        →
      </Button>
    </div>
  );
}
