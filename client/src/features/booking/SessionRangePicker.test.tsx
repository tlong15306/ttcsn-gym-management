import { useState } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { SessionRangePicker, type SessionRangeOption } from "./SessionRangePicker";

const sessions: SessionRangeOption[] = [
  { id: 1, starts_at: "2026-12-01T08:00:00+07:00", ends_at: "2026-12-01T09:00:00+07:00", remaining: 2, unavailable_reason: "" },
  { id: 2, starts_at: "2026-12-01T09:00:00+07:00", ends_at: "2026-12-01T10:00:00+07:00", remaining: 0, unavailable_reason: "Hết chỗ" },
  { id: 3, starts_at: "2026-12-01T10:00:00+07:00", ends_at: "2026-12-01T11:00:00+07:00", remaining: 2, unavailable_reason: "" },
];

function Picker() {
  const [selected, setSelected] = useState<SessionRangeOption[]>([]);
  const [error, setError] = useState("");
  return <>
    <SessionRangePicker sessions={sessions} selectedIDs={selected.map((item) => item.id)} onSelect={setSelected} onRangeError={setError} formatTime={(value) => value.slice(11, 16)} />
    <div role="alert">{error}</div>
  </>;
}

describe("session range selection", () => {
  it("keeps the original selection when a range crosses an unavailable session", () => {
    render(<Picker />);
    const first = screen.getByRole("button", { name: /08:00.*09:00/ });
    const last = screen.getByRole("button", { name: /10:00.*11:00/ });
    fireEvent.click(first);
    fireEvent.click(last);
    expect(screen.getByRole("alert")).toHaveTextContent("không còn chỗ");
    expect(first).toHaveAttribute("aria-pressed", "true");
    expect(last).toHaveAttribute("aria-pressed", "false");
  });

  it("cancels the unfinished endpoint with Escape", () => {
    render(<Picker />);
    const first = screen.getByRole("button", { name: /08:00.*09:00/ });
    const last = screen.getByRole("button", { name: /10:00.*11:00/ });
    fireEvent.click(first);
    fireEvent.keyDown(first, { key: "Escape" });
    fireEvent.click(last);
    expect(first).toHaveAttribute("aria-pressed", "false");
    expect(last).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByRole("alert")).toBeEmptyDOMElement();
  });
});
