import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes, useLocation } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";
import { SessionProvider } from "../../app/session";
import * as api from "../../api";
import { DiscoveryHome, DiscoveryLayout, DiscoveryRoom } from "./Discovery";
import { loginDestination } from "../../app/navigation";

const room = {
  id: 9,
  name: "Public Strength",
  room_type: "STRENGTH",
  default_capacity: 18,
  description: "Thông tin công khai",
  status: "AVAILABLE",
  has_cover: false,
  cover_alt_text: "",
  cover_version: 0,
  cover_is_demo: false,
};
const slot = {
  id: 17,
  room_id: 9,
  session_date: "2027-01-01",
  starts_at: "2026-12-31T23:00:00Z",
  ends_at: "2027-01-01T00:00:00Z",
  capacity: 18,
  remaining: 4,
  status: "OPEN",
};
function LoginTarget() {
  const location = useLocation();
  return <p>{(location.state as { from: string })?.from}</p>;
}
function mount(path = "/") {
  return render(
    <SessionProvider>
      <MemoryRouter initialEntries={[path]}>
        <Routes>
          <Route element={<DiscoveryLayout />}>
            <Route path="/" element={<DiscoveryHome />} />
            <Route path="/explore/rooms/:id" element={<DiscoveryRoom />} />
          </Route>
          <Route path="/login" element={<LoginTarget />} />
        </Routes>
      </MemoryRouter>
    </SessionProvider>,
  );
}
afterEach(() => {
  vi.restoreAllMocks();
  sessionStorage.clear();
});
function mockPublicData() {
  return vi
    .spyOn(api, "request")
    .mockImplementation(async <T,>(path: string) => {
      const data = path.includes("/sessions?")
        ? [slot]
        : /\/rooms\/9$/.test(path)
          ? room
          : path.includes("/rooms?")
            ? [room]
            : [];
      return {
        data: data as T,
        meta: {
          page: 1,
          page_size: 100,
          total: Array.isArray(data) ? data.length : 1,
        },
        request_id: "public-test",
      };
    });
}
describe("Public discovery", () => {
  it("loads the home without authentication and uses only public reads", async () => {
    const request = mockPublicData();
    mount();
    expect(
      screen.getByRole("heading", { name: /Tập theo nhịp/ }),
    ).toBeInTheDocument();
    expect(
      await screen.findByRole("heading", { name: "Public Strength" }),
    ).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Đăng nhập" })).toBeInTheDocument();
    expect(
      request.mock.calls.every(
        (call) =>
          call[0].startsWith("/api/v1/public/") && call[2] === undefined,
      ),
    ).toBe(true);
  });
  it("shows real weekly availability and sends guests to login with their room/date", async () => {
    mockPublicData();
    mount("/explore/rooms/9?from=2027-01-01");
    expect(await screen.findByText("06:00 – 07:00")).toBeInTheDocument();
    expect(screen.getByText("Còn 4 chỗ")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: /^06:00 – 07:00/ }));
    fireEvent.click(
      screen.getByRole("link", { name: "Đăng nhập để đặt lịch" }),
    );
    expect(
      await screen.findByText("/member/rooms/9?date=2027-01-01"),
    ).toBeInTheDocument();
  });
  it("loads all pages for the seven-day availability view", async () => {
    const request = vi
      .spyOn(api, "request")
      .mockImplementation(async <T,>(path: string) => {
        if (!path.includes("/sessions?"))
          return { data: room as T, request_id: "test" };
        const second =
          new URL(path, "http://local").searchParams.get("page") === "2";
        return {
          data: [
            second
              ? {
                  ...slot,
                  id: 18,
                  starts_at: "2027-01-01T00:00:00Z",
                  ends_at: "2027-01-01T01:00:00Z",
                }
              : slot,
          ] as T,
          meta: { page: second ? 2 : 1, page_size: 1, total: 2 },
          request_id: "test",
        };
      });
    mount("/explore/rooms/9?from=2027-01-01");
    expect(await screen.findByText("07:00 – 08:00")).toBeInTheDocument();
    await waitFor(() =>
      expect(
        request.mock.calls.some((call) => call[0].includes("page=2")),
      ).toBe(true),
    );
  });
  it("offers retry when anonymous reads fail rather than showing fabricated rooms", async () => {
    vi.spyOn(api, "request").mockRejectedValue(new Error("Mất kết nối"));
    mount();
    expect(await screen.findAllByRole("alert")).toHaveLength(2);
    expect(screen.getAllByRole("button", { name: "Thử lại" })).toHaveLength(2);
    expect(screen.queryByText("Public Strength")).not.toBeInTheDocument();
  });
});
describe("Login return destination", () => {
  it("keeps a member's internal booking destination and rejects external or wrong-role targets", () => {
    expect(loginDestination("MEMBER", "/member/rooms/9?date=2027-01-01")).toBe(
      "/member/rooms/9?date=2027-01-01",
    );
    for (const from of [
      "https://example.com",
      "//example.com",
      "/\\example.com",
      "/admin/users",
    ])
      expect(loginDestination("MEMBER", from)).toBe("/member/rooms");
    expect(loginDestination("PT", "/member/rooms/9")).toBe("/pt/schedule");
  });
});
