import { useEffect } from "react";
import { createMemoryRouter, RouterProvider } from "react-router-dom";
import {
  render,
  screen,
  fireEvent,
  waitFor,
  within,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import { SessionProvider, useSession } from "../../../app/session";
import { resource } from "../../../api";
import { RoomCalendar } from "./RoomCalendar";
import { WeeklyGrid } from "./WeeklyGrid";
import { emptyWeek, paint, toHours, toWeek, moveBlock, blockAt } from "./hours";
afterEach(() => {
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});
describe("Weekly calendar boundaries", () => {
  it("uses exactly 14 cells for 07–21 and preserves midnight", () => {
    const w = paint(emptyWeek(), [0, 7], [0, 20], true);
    expect(w[0].filter(Boolean)).toHaveLength(14);
    expect(w[0][21]).toBe(false);
    expect(toHours(w)).toEqual([
      { weekday: 1, start_minute: 420, end_minute: 1260 },
    ]);
    const late = toWeek([{ weekday: 7, start_minute: 1380, end_minute: 1440 }]);
    expect(toHours(late)).toEqual([
      { weekday: 7, start_minute: 1380, end_minute: 1440 },
    ]);
  });
  it("splits breaks and moves a block between days without changing its duration", () => {
    const w = paint(
      paint(emptyWeek(), [0, 7], [0, 20], true),
      [0, 12],
      [0, 13],
      false,
    );
    expect(toHours(w)).toHaveLength(2);
    const span = blockAt(w, 0, 8)!;
    const moved = moveBlock(w, 0, span, 6, 22);
    expect(moved[0][7]).toBe(false);
    expect(toHours(moved).find((h) => h.weekday === 7)).toEqual({
      weekday: 7,
      start_minute: 1140,
      end_minute: 1440,
    });
  });
  it("keyboard selection stays draft until Enter and Escape cancels", () => {
    const change = vi.fn();
    render(<WeeklyGrid week={emptyWeek()} mode="OPEN" onChange={change} />);
    const cells = screen.getAllByRole("gridcell");
    expect(cells).toHaveLength(168);
    fireEvent.keyDown(cells[7], { key: "ArrowRight", shiftKey: true });
    expect(change).not.toHaveBeenCalled();
    fireEvent.keyDown(cells[8], { key: "Enter" });
    expect(toHours(change.mock.calls[0][0])).toEqual([
      { weekday: 1, start_minute: 420, end_minute: 540 },
    ]);
    change.mockClear();
    fireEvent.keyDown(cells[7], { key: "ArrowDown", shiftKey: true });
    fireEvent.keyDown(cells[31], { key: "Escape" });
    expect(change).not.toHaveBeenCalled();
  });
  it("offers a two-tap alternative to pointer dragging", () => {
    const change = vi.fn();
    render(<WeeklyGrid week={emptyWeek()} mode="OPEN" onChange={change} />);
    fireEvent.click(screen.getByRole("button", { name: "Chọn bằng chạm" }));
    const cells = screen.getAllByRole("gridcell");
    fireEvent.pointerDown(cells[7], { button: 0, pointerId: 1 });
    expect(change).not.toHaveBeenCalled();
    fireEvent.pointerDown(cells[8], { button: 0, pointerId: 2 });
    expect(toHours(change.mock.calls[0][0])).toEqual([
      { weekday: 1, start_minute: 420, end_minute: 540 },
    ]);
  });
  it("moves a block with the two-tap alternative", () => {
    const change = vi.fn();
    const week = paint(emptyWeek(), [0, 7], [0, 8], true);
    render(<WeeklyGrid week={week} mode="MOVE" onChange={change} />);
    fireEvent.click(screen.getByRole("button", { name: "Chọn bằng chạm" }));
    const cells = screen.getAllByRole("gridcell");
    fireEvent.pointerDown(cells[7], { button: 0, pointerId: 1 });
    fireEvent.pointerDown(cells[31], { button: 0, pointerId: 2 });
    expect(toHours(change.mock.calls[0][0])).toEqual([
      { weekday: 2, start_minute: 420, end_minute: 540 },
    ]);
  });
  it("moves and resizes blocks through the keyboard alternative", () => {
    const w = paint(emptyWeek(), [0, 7], [0, 9], true);
    const change = vi.fn();
    const view = render(<WeeklyGrid week={w} mode="MOVE" onChange={change} />);
    fireEvent.keyDown(screen.getAllByRole("gridcell")[7], {
      key: "ArrowDown",
      shiftKey: true,
    });
    fireEvent.keyDown(screen.getAllByRole("gridcell")[31], { key: "Enter" });
    expect(toHours(change.mock.calls[0][0])).toEqual([
      { weekday: 2, start_minute: 420, end_minute: 600 },
    ]);
    view.rerender(<WeeklyGrid week={w} mode="START" onChange={change} />);
    fireEvent.keyDown(screen.getAllByRole("gridcell")[7], {
      key: "ArrowLeft",
      shiftKey: true,
    });
    fireEvent.keyDown(screen.getAllByRole("gridcell")[6], { key: "Enter" });
    expect(toHours(change.mock.calls[1][0])).toEqual([
      { weekday: 1, start_minute: 360, end_minute: 600 },
    ]);
  });
});
const snapshot = {
  room: {
    id: 1,
    name: "Phòng test",
    status: "AVAILABLE",
    calendar_revision: 0,
    calendar_prepared_through: "",
    calendar_job_error: "",
  },
  versions: [],
  exceptions: [],
};
function Bootstrap({ reload }: { reload: () => void }) {
  const auth = useSession();
  useEffect(
    () =>
      auth.login({
        accessToken: "test",
        expiresIn: 900,
        user: {
          id: 1,
          username: "admin.demo",
          full_name: "Demo",
          email: "demo@test.local",
          role: "ADMIN",
          status: "ACTIVE",
          created_at: "",
          updated_at: "",
        },
      }),
    [],
  );
  return auth.session ? (
    <RoomCalendar snapshot={snapshot} reload={reload} />
  ) : null;
}
function mount(reload = vi.fn()) {
  vi.spyOn(resource, "get").mockResolvedValue({ data: [], request_id: "test" });
  const router = createMemoryRouter([
    { path: "/", element: <Bootstrap reload={reload} /> },
  ]);
  render(
    <SessionProvider>
      <RouterProvider router={router} />
    </SessionProvider>,
  );
  return reload;
}
describe("Room calendar publication", () => {
  it("uses preview then an explicit confirm with an idempotency key", async () => {
    const user = userEvent.setup();
    const reload = mount();
    const save = vi.spyOn(resource, "save").mockResolvedValue({
      data: {
        preview_id: "preview",
        expires_at: "",
        creates: 28,
        changes: [],
        conflicts: [],
        prepared_through: "2026-11-10",
      },
      request_id: "test",
    });
    const fetcher = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ data: {}, request_id: "test" }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );
    vi.stubGlobal("fetch", fetcher);
    await user.click(
      screen.getByRole("button", { name: "Cả tuần 07:00–21:00" }),
    );
    expect(save).not.toHaveBeenCalled();
    await user.type(screen.getByLabelText("Lý do công bố"), "Mở cửa tuần này");
    await user.click(
      screen.getByRole("button", { name: "Xem trước và xác nhận" }),
    );
    const dialog = await screen.findByRole("dialog", {
      name: "Xác nhận thay đổi lịch",
    });
    expect(fetcher).not.toHaveBeenCalled();
    expect(save.mock.calls[0][2]).toMatchObject({
      kind: "WEEKLY",
      base_revision: 0,
    });
    await user.click(
      within(dialog).getByRole("button", { name: "Xác nhận đăng ký" }),
    );
    await waitFor(() => expect(reload).toHaveBeenCalled());
    const options = fetcher.mock.calls[0][1];
    expect(options.method).toBe("PUT");
    expect(new Headers(options.headers).get("Idempotency-Key")).toBeTruthy();
    expect(JSON.parse(options.body).preview_id).toBe("preview");
  });
  it("keeps a failed preview editable and blocks confirmation for held bookings", async () => {
    const user = userEvent.setup();
    mount();
    const save = vi
      .spyOn(resource, "save")
      .mockRejectedValueOnce(new Error("Lịch đã thay đổi"));
    await user.type(screen.getByLabelText("Lý do công bố"), "Test");
    await user.click(
      screen.getByRole("button", { name: "Xem trước và xác nhận" }),
    );
    expect(await screen.findByText("Lịch đã thay đổi")).toBeVisible();
    save.mockResolvedValueOnce({
      data: {
        preview_id: "x",
        creates: 0,
        changes: [],
        conflicts: [{ session_id: 1 }],
        prepared_through: "2026-11-10",
      },
      request_id: "t",
    });
    await user.click(
      screen.getByRole("button", { name: "Xem trước và xác nhận" }),
    );
    const dialog = await screen.findByRole("dialog", {
      name: "Xác nhận thay đổi lịch",
    });
    expect(
      within(dialog).getByRole("button", { name: "Xác nhận đăng ký" }),
    ).toBeDisabled();
  });
  it("undoes changes and confirms a Sunday midnight split before editing", async () => {
    const user = userEvent.setup();
    mount();
    await user.click(
      screen.getByRole("button", { name: "Cả tuần 07:00–21:00" }),
    );
    expect(screen.getByText("98 giờ / tuần")).toBeVisible();
    await user.click(screen.getByRole("button", { name: "Hoàn tác" }));
    expect(screen.getByText("0 giờ / tuần")).toBeVisible();
    await user.selectOptions(screen.getByLabelText("Ngày"), "6");
    await user.selectOptions(screen.getByLabelText("Bắt đầu"), "22");
    await user.selectOptions(screen.getByLabelText("Kết thúc"), "2");
    await user.click(screen.getByRole("button", { name: "Mở khoảng giờ" }));
    const dialog = await screen.findByRole("dialog", {
      name: "Khoảng giờ qua nửa đêm",
    });
    expect(screen.getByText("0 giờ / tuần")).toBeVisible();
    await user.click(
      within(dialog).getByRole("button", { name: "Xác nhận tách ngày" }),
    );
    expect(screen.getByText("4 giờ / tuần")).toBeVisible();
  });
});
