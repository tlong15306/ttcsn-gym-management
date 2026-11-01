import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { expect, it, vi } from "vitest";
import { TimeCalendar } from "./TimeCalendar";
import type { Booking } from "./Booking";

it("places the PT booking at Vietnam time when Intl is overridden", () => {
  const format = vi.spyOn(Intl, "DateTimeFormat").mockImplementation(() => {
    throw new Error("Unexpected dependence on browser timezone formatting");
  });
  try {
    const booking = {
      id: 1, status: "CONFIRMED", starts_at: "2026-10-31T23:00:00Z",
      ends_at: "2026-11-01T00:00:00Z", member_name: "Hội viên demo",
      room_name: "Phòng demo",
    } as Booking;
    render(<MemoryRouter><TimeCalendar days={["2026-11-01"]} bookings={[booking]} /></MemoryRouter>);
    expect(screen.getByText("CN, 01/11")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /Hội viên demo/ })).toHaveStyle({ top: "264px", height: "44px" });
  } finally {
    format.mockRestore();
  }
});
