import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import {
  AvailabilityGrid,
  availabilityHoursFromWeek,
  splitAvailabilityHours,
} from "./AvailabilityGrid";

describe("AvailabilityGrid", () => {
  it("keeps minute-precise availability outside the half-hour grid", () => {
    const result = splitAvailabilityHours([
      { day_of_week: 1, start_time: "07:00", end_time: "09:00" },
      { day_of_week: 2, start_time: "07:15", end_time: "09:45" },
    ]);

    expect(result.preciseDays).toEqual(new Set([2]));
    expect(result.preserved).toEqual([
      { day_of_week: 2, start_time: "07:15", end_time: "09:45" },
    ]);
    expect(result.week[0].slice(14, 18)).toEqual([true, true, true, true]);
    expect(result.week[1].some(Boolean)).toBe(false);
  });

  it("commits a half-hour range through the two-tap alternative", () => {
    const change = vi.fn();
    render(<AvailabilityGrid hours={[]} onChange={change} />);
    fireEvent.click(screen.getByRole("button", { name: "Chọn bằng chạm" }));
    const cells = screen.getAllByRole("gridcell");
    fireEvent.pointerDown(cells[14], { button: 0, pointerId: 1 });
    fireEvent.pointerDown(cells[15], { button: 0, pointerId: 2 });

    expect(change).toHaveBeenCalledWith([
      { day_of_week: 1, start_time: "07:00", end_time: "08:00" },
    ]);
  });

  it("does not drag or tap through a day with minute-precise availability", () => {
    const change = vi.fn();
    render(
      <AvailabilityGrid
        hours={[{ day_of_week: 2, start_time: "07:15", end_time: "09:45" }]}
        onChange={change}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Chọn bằng chạm" }));
    const cells = screen.getAllByRole("gridcell");
    fireEvent.pointerDown(cells[14], { button: 0, pointerId: 1 });
    fireEvent.pointerDown(cells[48 * 2 + 14], { button: 0, pointerId: 2 });

    expect(change).not.toHaveBeenCalled();
  });

  it("exports contiguous grid cells as exact intervals", () => {
    const week = Array.from({ length: 7 }, () => Array(48).fill(false));
    week[0][14] = true;
    week[0][15] = true;
    week[0][18] = true;

    expect(availabilityHoursFromWeek(week)).toEqual([
      { day_of_week: 1, start_time: "07:00", end_time: "08:00" },
      { day_of_week: 1, start_time: "09:00", end_time: "09:30" },
    ]);
  });
});
