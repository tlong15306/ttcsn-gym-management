import { useState } from "react";
import { render, screen, fireEvent } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { WeekCalendar } from "./WeekCalendar";
import { rangeBetween, type WeekSession } from "./weekly";
import {
  clearDraft,
  DRAFT_TTL,
  readDraft,
  restoreDraft,
  saveDraft,
} from "./draft";

const slots: WeekSession[] = Array.from({ length: 4 }, (_, i) => ({
  id: i + 1,
  starts_at: `2027-01-01T${String(6 + i).padStart(2, "0")}:00:00+07:00`,
  ends_at: `2027-01-01T${String(7 + i).padStart(2, "0")}:00:00+07:00`,
  remaining: 5,
  unavailable_reason: "",
}));
function Calendar({ values = slots }: { values?: WeekSession[] }) {
  const [chosen, setChosen] = useState<WeekSession[]>([]),
    [date, setDate] = useState("2027-01-01"),
    [error, setError] = useState("");
  return (
    <>
      <WeekCalendar
        sessions={values}
        selectedIDs={chosen.map((s) => s.id)}
        date={date}
        onDate={setDate}
        onSelect={setChosen}
        onError={setError}
      />
      <p data-testid="selected">{chosen.map((s) => s.id).join(",")}</p>
      <p role="alert">{error}</p>
    </>
  );
}
afterEach(() => {
  clearDraft();
  vi.restoreAllMocks();
  vi.useRealTimers();
});
describe("Weekly selection", () => {
  it("keeps seven days and supports two endpoints, resizing and moving with a keyboard", () => {
    render(<Calendar />);
    expect(document.querySelectorAll(".bk-head>button")).toHaveLength(7);
    fireEvent.click(screen.getByRole("button", { name: /^06:00 – 07:00/ }));
    fireEvent.click(screen.getByRole("button", { name: /^08:00 – 09:00/ }));
    expect(screen.getByTestId("selected")).toHaveTextContent("1,2,3");
    fireEvent.keyDown(
      screen.getByRole("button", { name: /Đổi giờ kết thúc/ }),
      { key: "ArrowUp" },
    );
    expect(screen.getByTestId("selected")).toHaveTextContent("1,2");
    fireEvent.keyDown(
      screen.getByRole("button", { name: /Di chuyển khoảng giờ/ }),
      { key: "ArrowDown" },
    );
    expect(screen.getByTestId("selected")).toHaveTextContent("2,3");
    fireEvent.change(screen.getByLabelText("Bắt đầu"), {
      target: { value: "3" },
    });
    expect(screen.getByTestId("selected")).toHaveTextContent("3");
  });
  it("rejects holes and blocked ranges and keeps the previous valid choice", () => {
    render(
      <Calendar
        values={slots.map((s) =>
          s.id === 2 ? { ...s, unavailable_reason: "Hết chỗ" } : s,
        )}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: /^06:00 – 07:00/ }));
    fireEvent.click(screen.getByRole("button", { name: /^09:00 – 10:00/ }));
    expect(screen.getByTestId("selected")).toHaveTextContent("1");
    expect(screen.getByRole("alert")).toHaveTextContent("không thể đặt");
    expect(rangeBetween([slots[0], slots[2]], 1, 3).error).toContain(
      "khoảng trống",
    );
  });
  it("view mode allows navigation without changing the chosen range", () => {
    render(<Calendar />);
    fireEvent.click(screen.getByRole("button", { name: "Xem lịch" }));
    fireEvent.click(screen.getByRole("button", { name: /^06:00 – 07:00/ }));
    expect(screen.getByTestId("selected")).toBeEmptyDOMElement();
    fireEvent.click(screen.getByRole("button", { name: "Ngày" }));
    expect(document.querySelectorAll(".bk-head>button")).toHaveLength(1);
  });
  it("does not cross days or change the selected duration when the destination has a gap", () => {
    expect(
      rangeBetween(
        [
          slots[0],
          {
            ...slots[1],
            starts_at: "2027-01-02T07:00:00+07:00",
            ends_at: "2027-01-02T08:00:00+07:00",
          },
        ],
        1,
        2,
      ).error,
    ).toContain("cùng ngày");
    render(<Calendar values={[slots[0], slots[1], slots[3]]} />);
    fireEvent.click(screen.getByRole("button", { name: /^06:00 – 07:00/ }));
    fireEvent.click(screen.getByRole("button", { name: /^07:00 – 08:00/ }));
    fireEvent.keyDown(
      screen.getByRole("button", { name: /Di chuyển khoảng giờ/ }),
      { key: "ArrowDown" },
    );
    expect(screen.getByTestId("selected")).toHaveTextContent("1,2");
  });
});
describe("Draft after login", () => {
  it("stores no credentials and only restores against fresh eligible sessions", () => {
    saveDraft(9, slots.slice(0, 2), null, 0, true);
    const draft = readDraft()!;
    expect(draft.owner).toBeNull();
    expect(restoreDraft(draft, 9, 7, slots).values).toHaveLength(2);
    expect(
      restoreDraft(
        draft,
        9,
        7,
        slots.map((s) =>
          s.id === 2 ? { ...s, unavailable_reason: "Không đủ hạn thẻ" } : s,
        ),
      ).values,
    ).toHaveLength(0);
    expect(JSON.stringify(draft)).not.toMatch(/token|password|email/i);
  });
  it("expires in thirty minutes and rejects another member or a changed session", () => {
    vi.useFakeTimers();
    saveDraft(9, [slots[0]], 12, 3, true);
    const draft = readDraft()!;
    expect(restoreDraft(draft, 9, 13, slots).error).toContain("tài khoản khác");
    expect(
      restoreDraft(draft, 9, 12, [{ ...slots[0], ends_at: slots[1].ends_at }])
        .values,
    ).toHaveLength(0);
    vi.advanceTimersByTime(DRAFT_TTL + 1);
    expect(readDraft()).toBeNull();
  });
  it("does not restore another room and removes malformed storage", () => {
    saveDraft(9, [slots[0]], null);
    expect(restoreDraft(readDraft()!, 10, 1, slots).values).toHaveLength(0);
    sessionStorage.setItem("gym.booking-draft.v1", '{"version":99}');
    expect(readDraft()).toBeNull();
  });
});
