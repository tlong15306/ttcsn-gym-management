import { useEffect, type ReactNode } from "react";
import {
  act,
  render,
  screen,
  fireEvent,
  waitFor,
} from "@testing-library/react";
import { createMemoryRouter, RouterProvider } from "react-router-dom";
import { afterEach, describe, it, expect, vi } from "vitest";
import { SessionProvider, useSession } from "../../app/session";
import { resource } from "../../api";
import { BookingList, RoomCatalog, RoomBooking } from "./Booking";
import { PortalLayout } from "../../layouts/PortalLayout";
import { RoomImage } from "./RoomImage";
import { saveDraft } from "./draft";
import type { Role } from "../../types";
const room = {
  id: 3,
  name: "Strength Studio",
  room_type: "GYM",
  default_capacity: 12,
  description: "Không gian tập sức mạnh",
  status: "AVAILABLE",
  has_cover: false,
};
function Bootstrap({
  children,
  role = "MEMBER",
}: {
  children: ReactNode;
  role?: Role;
}) {
  const auth = useSession();
  useEffect(
    () =>
      auth.login({
        accessToken: "member-token",
        expiresIn: 900,
        user: {
          id: 1,
          username: "tester",
          full_name: "Tester",
          email: "test@example.local",
          role,
          status: "ACTIVE",
          created_at: "",
          updated_at: "",
        },
      }),
    [],
  );
  return auth.session ? children : null;
}
function mount(element: ReactNode, path = "/", role: Role = "MEMBER") {
  if (path === "/member/rooms/3") path += "?date=2026-12-01";
  const router = createMemoryRouter(
    [
      { path: "/", element },
      { path: "/member/rooms/:id", element },
      { path: "/member/bookings/:id", element: <h1>Đặt lịch thành công</h1> },
    ],
    { initialEntries: [path] },
  );
  return render(
    <SessionProvider>
      <Bootstrap role={role}>
        <RouterProvider router={router} />
      </Bootstrap>
    </SessionProvider>,
  );
}
afterEach(() => {
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
  sessionStorage.clear();
});
describe("Room-first booking", () => {
  it("keeps the calendar mounted and selectable during a same-week focus refresh", async () => {
    const value = {
      id: 8,
      starts_at: "2026-12-01T08:00:00+07:00",
      ends_at: "2026-12-01T09:00:00+07:00",
      remaining: 4,
      unavailable_reason: "",
    };
    let refresh = false;
    const waiting: (() => void)[] = [];
    vi.spyOn(resource, "get").mockResolvedValue({
      data: room,
      request_id: "test",
    });
    vi.spyOn(resource, "list").mockImplementation(async (_token, path) => {
      if (path.includes("/pts"))
        return {
          data: [],
          meta: { page: 1, page_size: 100, total: 0 },
          request_id: "test",
        };
      if (refresh) await new Promise<void>((resolve) => waiting.push(resolve));
      return {
        data: [value],
        meta: { page: 1, page_size: 100, total: 1 },
        request_id: "test",
      };
    });
    mount(<RoomBooking />, "/member/rooms/3");
    const slot = await screen.findByRole("button", { name: /^08:00.*09:00/ });
    refresh = true;
    fireEvent(window, new Event("focus"));
    await screen.findByText("Đang cập nhật chỗ và điều kiện đặt…");
    expect(document.querySelector('[data-session-id="8"]')).toBe(slot);
    fireEvent.click(slot);
    expect(slot).toHaveAttribute("aria-pressed", "true");
    expect(
      screen.getByRole("button", { name: "Kiểm tra đặt lịch →" }),
    ).toBeDisabled();
    await waitFor(() => expect(waiting.length).toBe(7));
    await act(async () => {
      waiting.splice(0).forEach((resolve) => resolve());
    });
    await waitFor(() =>
      expect(
        screen.getByRole("button", { name: "Kiểm tra đặt lịch →" }),
      ).toBeEnabled(),
    );
    expect(document.querySelector('[data-session-id="8"]')).toBe(slot);
    expect(slot).toHaveAttribute("aria-pressed", "true");
  });
  it("keeps a changed week in the draft and restores it when returning without a date query", async () => {
    vi.spyOn(resource, "get").mockResolvedValue({
      data: room,
      request_id: "test",
    });
    vi.spyOn(resource, "list").mockImplementation(
      async (_token, path, params) => {
        const date = params.get("date");
        const data =
          path.includes("/pts") ||
          !date ||
          !["2026-12-01", "2026-12-08"].includes(date)
            ? []
            : [
                {
                  id: date === "2026-12-01" ? 8 : 18,
                  starts_at: `${date}T08:00:00+07:00`,
                  ends_at: `${date}T09:00:00+07:00`,
                  remaining: 4,
                  unavailable_reason: "",
                },
              ];
        return {
          data,
          meta: { page: 1, page_size: 100, total: data.length },
          request_id: "test",
        };
      },
    );
    const view = mount(<RoomBooking />, "/member/rooms/3");
    fireEvent.click(
      await screen.findByRole("button", { name: /^08:00.*09:00/ }),
    );
    fireEvent.change(screen.getByLabelText("Ngày tập"), {
      target: { value: "2026-12-08" },
    });
    await waitFor(() =>
      expect(document.querySelector('[data-session-id="18"]')).not.toBeNull(),
    );
    fireEvent.click(screen.getByRole("button", { name: /^08:00.*09:00/ }));
    expect(document.querySelector('[data-session-id="18"]')).toHaveAttribute(
      "aria-pressed",
      "true",
    );
    fireEvent.blur(screen.getByLabelText("Ngày tập"));
    expect(document.querySelector('[data-session-id="18"]')).toHaveAttribute(
      "aria-pressed",
      "true",
    );
    view.unmount();
    mount(<RoomBooking />, "/member/rooms/3?");
    await waitFor(() =>
      expect(document.querySelector('[data-session-id="18"]')).toHaveAttribute(
        "aria-pressed",
        "true",
      ),
    );
    expect(screen.getByLabelText("Ngày tập")).toHaveValue("2026-12-08");
  });
  it("restores an anonymous range but requires an explicit PT choice", async () => {
    const values = [
      {
        id: 8,
        starts_at: "2026-12-01T08:00:00+07:00",
        ends_at: "2026-12-01T09:00:00+07:00",
        remaining: 4,
        unavailable_reason: "",
      },
    ];
    saveDraft(3, values, null, 0, true);
    vi.spyOn(resource, "get").mockResolvedValue({
      data: room,
      request_id: "test",
    });
    vi.spyOn(resource, "list").mockImplementation(async (_token, path) => ({
      data: path.includes("/pts") ? [] : values,
      meta: { page: 1, page_size: 100, total: path.includes("/pts") ? 0 : 1 },
      request_id: "test",
    }));
    mount(<RoomBooking />, "/member/rooms/3");
    const slot = await screen.findByRole("button", { name: /^08:00.*09:00/ });
    await waitFor(() => expect(slot).toHaveAttribute("aria-pressed", "true"));
    expect(
      screen.getByRole("button", { name: "Kiểm tra đặt lịch →" }),
    ).toBeDisabled();
    fireEvent.click(screen.getByRole("button", { name: /^Tự tập/ }));
    expect(
      screen.getByRole("button", { name: "Kiểm tra đặt lịch →" }),
    ).toBeEnabled();
  });
  it("rechecks current eligibility before opening review and makes no mutation if it changed", async () => {
    const value = {
      id: 8,
      starts_at: "2026-12-01T08:00:00+07:00",
      ends_at: "2026-12-01T09:00:00+07:00",
      remaining: 4,
      unavailable_reason: "",
    };
    let reads = 0;
    vi.spyOn(resource, "get").mockResolvedValue({
      data: room,
      request_id: "test",
    });
    vi.spyOn(resource, "list").mockImplementation(
      async (_token, path, params) => {
        if (path.includes("/pts"))
          return {
            data: [],
            meta: { page: 1, page_size: 100, total: 0 },
            request_id: "test",
          };
        if (params.get("date") === "2026-12-01") reads++;
        return {
          data: [
            reads > 1 ? { ...value, unavailable_reason: "Hết chỗ" } : value,
          ],
          meta: { page: 1, page_size: 100, total: 1 },
          request_id: "test",
        };
      },
    );
    const fetch = vi.spyOn(globalThis, "fetch");
    mount(<RoomBooking />, "/member/rooms/3");
    fireEvent.click(
      await screen.findByRole("button", { name: /^08:00.*09:00/ }),
    );
    fireEvent.click(
      screen.getByRole("button", { name: "Kiểm tra đặt lịch →" }),
    );
    expect(
      await screen.findByText(/Giờ đã chọn không còn phù hợp/),
    ).toBeInTheDocument();
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(fetch).not.toHaveBeenCalled();
  });
  it("keeps the valid day selected before login when entering booking", async () => {
    vi.spyOn(resource, "get").mockResolvedValue({
      data: room,
      request_id: "test",
    });
    const list = vi.spyOn(resource, "list").mockResolvedValue({
      data: [],
      meta: { page: 1, page_size: 100, total: 0 },
      request_id: "test",
    });
    mount(<RoomBooking />, "/member/rooms/3?date=2027-01-01");
    await waitFor(() =>
      expect(
        list.mock.calls.some(
          (call) =>
            call[1] === "booking-options/sessions" &&
            call[2].get("date") === "2027-01-01",
        ),
      ).toBe(true),
    );
    expect(screen.getByLabelText("Ngày tập")).toHaveValue("2027-01-01");
  });
  it("lets operations search and paginate the room filter beyond the first 100", async () => {
    const list = vi
      .spyOn(resource, "list")
      .mockImplementation(async (_token, path, params) => {
        if (path === "rooms") {
          const page = Number(params.get("page"));
          return {
            data: [page === 1 ? room : { ...room, id: 101, name: "Room 101" }],
            meta: {
              page,
              page_size: Number(params.get("page_size")),
              total: 101,
            },
            request_id: "test",
          };
        }
        return {
          data: [],
          meta: { page: 1, page_size: 20, total: 0 },
          request_id: "test",
        };
      });
    mount(<BookingList attendance />, "/", "MANAGER");
    expect(
      await screen.findByRole("option", { name: "Strength Studio" }),
    ).toBeInTheDocument();
    fireEvent.click(screen.getAllByRole("button", { name: "Trang sau" })[0]);
    expect(
      await screen.findByRole("option", { name: "Room 101" }),
    ).toBeInTheDocument();
    fireEvent.change(screen.getByPlaceholderText("Tìm tên hoặc loại phòng…"), {
      target: { value: "cardio" },
    });
    await waitFor(() =>
      expect(
        list.mock.calls.some(
          (call) =>
            call[1] === "rooms" &&
            call[2].get("q") === "cardio" &&
            call[2].get("page") === "1",
        ),
      ).toBe(true),
    );
  });
  it("links each room to its detail without creating a booking", async () => {
    vi.spyOn(resource, "list").mockResolvedValue({
      data: [room],
      meta: { page: 1, page_size: 20, total: 1 },
      request_id: "test",
    });
    const save = vi.spyOn(resource, "save");
    mount(<RoomCatalog />);
    expect(
      await screen.findByRole("link", { name: /Strength Studio/ }),
    ).toHaveAttribute("href", "/member/rooms/3");
    expect(screen.getByText("Phòng chưa có ảnh")).toBeInTheDocument();
    expect(save).not.toHaveBeenCalled();
  });
  it("keeps the idempotency key after a lost response and confirms only on review", async () => {
    vi.spyOn(resource, "get").mockResolvedValue({
      data: room,
      request_id: "test",
    });
    vi.spyOn(resource, "list").mockImplementation(async (_token, path) => ({
      data: path.includes("/pts")
        ? []
        : [
            {
              id: 8,
              room_id: 3,
              room_name: room.name,
              starts_at: "2026-12-01T09:00:00+07:00",
              ends_at: "2026-12-01T10:00:00+07:00",
              remaining: 4,
              unavailable_reason: "",
            },
          ],
      meta: { page: 1, page_size: 20, total: path.includes("/pts") ? 0 : 1 },
      request_id: "test",
    }));
    const calls: RequestInit[] = [];
    const fetchMock = vi
      .spyOn(globalThis, "fetch")
      .mockImplementation(async (_url, init) => {
        calls.push(init!);
        if (calls.length === 1) throw new Error("Mất kết nối");
        return new Response(
          JSON.stringify({ data: { id: 22 }, request_id: "test" }),
          { status: 201, headers: { "Content-Type": "application/json" } },
        );
      });
    mount(<RoomBooking />, "/member/rooms/3");
    fireEvent.click(
      await screen.findByRole("button", { name: /09:00.*10:00/ }),
    );
    fireEvent.click(
      screen.getByRole("button", { name: "Kiểm tra đặt lịch →" }),
    );
    expect(fetchMock).not.toHaveBeenCalled();
    fireEvent.click(
      await screen.findByRole("button", { name: "Xác nhận đặt lịch" }),
    );
    expect(await screen.findByText("Mất kết nối")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Xác nhận đặt lịch" }));
    expect(
      await screen.findByRole("heading", { name: "Đặt lịch thành công" }),
    ).toBeInTheDocument();
    expect(new Headers(calls[0].headers).get("Idempotency-Key")).toBe(
      new Headers(calls[1].headers).get("Idempotency-Key"),
    );
  });
  it("selects a dragged consecutive range and submits one atomic batch", async () => {
    vi.stubGlobal("PointerEvent", MouseEvent);
    vi.spyOn(resource, "get").mockResolvedValue({
      data: room,
      request_id: "test",
    });
    vi.spyOn(resource, "list").mockImplementation(async (_token, path) => ({
      data: path.includes("/pts")
        ? []
        : [
            {
              id: 8,
              room_id: 3,
              room_name: room.name,
              starts_at: "2026-12-01T08:00:00+07:00",
              ends_at: "2026-12-01T09:00:00+07:00",
              remaining: 4,
              unavailable_reason: "",
            },
            {
              id: 9,
              room_id: 3,
              room_name: room.name,
              starts_at: "2026-12-01T09:00:00+07:00",
              ends_at: "2026-12-01T10:00:00+07:00",
              remaining: 4,
              unavailable_reason: "",
            },
          ],
      meta: { page: 1, page_size: 100, total: path.includes("/pts") ? 0 : 2 },
      request_id: "test",
    }));
    const calls: Array<{ url: string; init: RequestInit }> = [];
    vi.spyOn(globalThis, "fetch").mockImplementation(async (url, init) => {
      calls.push({ url: String(url), init: init! });
      return new Response(
        JSON.stringify({ data: [{ id: 22 }, { id: 23 }], request_id: "test" }),
        { status: 201, headers: { "Content-Type": "application/json" } },
      );
    });
    mount(<RoomBooking />, "/member/rooms/3");
    const start = await screen.findByRole("button", { name: /08:00.*09:00/ });
    const end = screen.getByRole("button", { name: /09:00.*10:00/ });
    Object.defineProperty(document, "elementFromPoint", {
      configurable: true,
      value: () => end,
    });
    fireEvent.pointerDown(start, { pointerId: 1 });
    fireEvent.pointerMove(start, { pointerId: 1, clientX: 10, clientY: 10 });
    fireEvent.pointerUp(start, { pointerId: 1 });
    expect(screen.getByText(/2 ca · 120 phút/)).toBeInTheDocument();
    fireEvent.click(
      screen.getByRole("button", { name: "Kiểm tra đặt lịch →" }),
    );
    fireEvent.click(
      await screen.findByRole("button", { name: "Xác nhận đặt lịch" }),
    );
    expect(
      await screen.findByRole("heading", { name: "Đặt lịch thành công" }),
    ).toBeInTheDocument();
    expect(calls).toHaveLength(1);
    expect(calls[0].url).toContain("/api/v1/bookings/batch");
    expect(JSON.parse(String(calls[0].init.body))).toEqual({
      room_session_ids: [8, 9],
      pt_id: 0,
    });
  });
  it("finds a PT available throughout a range even beyond the first page", async () => {
    vi.spyOn(resource, "get").mockResolvedValue({
      data: room,
      request_id: "test",
    });
    const shared = {
      id: 999,
      full_name: "PT Shared",
      specialty: "Strength",
      rating_average: null,
      review_count: 0,
    };
    const list = vi
      .spyOn(resource, "list")
      .mockImplementation(async (_token, path, params) => {
        if (!path.includes("/pts"))
          return {
            data: [
              {
                id: 8,
                room_id: 3,
                room_name: room.name,
                starts_at: "2026-12-01T08:00:00+07:00",
                ends_at: "2026-12-01T09:00:00+07:00",
                remaining: 4,
                unavailable_reason: "",
              },
              {
                id: 9,
                room_id: 3,
                room_name: room.name,
                starts_at: "2026-12-01T09:00:00+07:00",
                ends_at: "2026-12-01T10:00:00+07:00",
                remaining: 4,
                unavailable_reason: "",
              },
            ],
            meta: { page: 1, page_size: 100, total: 2 },
            request_id: "test",
          };
        const page = Number(params.get("page"));
        const sessionID = Number(params.get("room_session_id"));
        const data =
          page === 1
            ? Array.from({ length: 100 }, (_, index) => ({
                id: sessionID * 1000 + index,
                full_name: `PT ${sessionID}-${index}`,
                specialty: "",
                rating_average: null,
                review_count: 0,
              }))
            : [shared];
        return {
          data,
          meta: { page, page_size: 100, total: 101 },
          request_id: "test",
        };
      });
    mount(<RoomBooking />, "/member/rooms/3");
    fireEvent.click(
      await screen.findByRole("button", { name: /08:00.*09:00/ }),
    );
    fireEvent.click(screen.getByRole("button", { name: /09:00.*10:00/ }));
    expect(
      await screen.findByRole("button", { name: /PT Shared/ }),
    ).toBeInTheDocument();
    expect(
      list.mock.calls.some(
        (call) =>
          call[1].includes("/pts") &&
          call[2].get("page") === "2" &&
          call[2].get("room_session_id") === "9",
      ),
    ).toBe(true);
  });
  it("uses an independent PT navigation without a create action", async () => {
    mount(<PortalLayout coach />, "/", "PT");
    expect(
      await screen.findByRole("link", { name: "Lịch làm việc" }),
    ).toHaveAttribute("href", "/pt/schedule");
    expect(screen.queryByText("Khám phá phòng")).not.toBeInTheDocument();
    expect(screen.queryByText("Đặt lịch")).not.toBeInTheDocument();
  });
  it("loads protected room images and revokes their object URL", async () => {
    const create = vi.fn(() => "blob:room");
    const revoke = vi.fn();
    Object.defineProperty(URL, "createObjectURL", {
      configurable: true,
      value: create,
    });
    Object.defineProperty(URL, "revokeObjectURL", {
      configurable: true,
      value: revoke,
    });
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(new Blob(["jpeg"]), { status: 200 }),
    );
    const view = mount(
      <RoomImage
        room={{
          ...room,
          has_cover: true,
          cover_alt_text: "Ảnh phòng strength",
        }}
      />,
    );
    expect(
      await screen.findByRole("img", { name: "Ảnh phòng strength" }),
    ).toHaveAttribute("src", "blob:room");
    await waitFor(() => expect(create).toHaveBeenCalled());
    view.unmount();
    expect(revoke).toHaveBeenCalledWith("blob:room");
  });
});
