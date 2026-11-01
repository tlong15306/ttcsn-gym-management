import { describe, expect, it, vi } from "vitest";
import { localDate, vietnamDay, vietnamTime, vietnamWeekday } from "./vietnamTime";
import { civilDate, vietnamDateTime, vietnamMinute } from "../../utils/vietnamTime";

describe("Vietnam booking time", () => {
  it("uses Vietnam's civil day and time even if browser formatting is forced to UTC", () => {
    const dateTimeFormat = vi.spyOn(Intl, "DateTimeFormat").mockImplementation(() => {
      throw new Error("Booking must not depend on an overridden formatter");
    });
    const instant = new Date("2026-10-31T23:00:00Z");
    expect(localDate(instant)).toBe("2026-11-01");
    expect(vietnamTime(instant.toISOString())).toBe("06:00");
    expect(vietnamDay(instant.toISOString())).toBe("CN, 01/11");
    expect(vietnamWeekday(instant)).toBe("CN");
    dateTimeFormat.mockRestore();
  });

  it("formats midnight across a year boundary", () => {
    expect(localDate(new Date("2026-12-31T17:00:00Z"))).toBe("2027-01-01");
    expect(vietnamTime("2026-12-31T17:00:00Z")).toBe("00:00");
    expect(vietnamDateTime("2026-12-31T17:00:00Z")).toBe("00:00 01/01/2027");
    expect(vietnamMinute("2026-12-31T17:00:00Z")).toBe(0);
    expect(civilDate("2026-12-31")).toBe("31/12/2026");
  });

  it("ignores overridden local Date getters", () => {
    const hour = vi.spyOn(Date.prototype, "getHours").mockReturnValue(23);
    const day = vi.spyOn(Date.prototype, "getDate").mockReturnValue(31);
    try {
      expect(localDate(new Date("2026-10-31T23:00:00Z"))).toBe("2026-11-01");
      expect(vietnamTime("2026-10-31T23:00:00Z")).toBe("06:00");
    } finally {
      hour.mockRestore();
      day.mockRestore();
    }
  });
});
