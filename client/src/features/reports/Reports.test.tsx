import { useEffect, type ReactNode } from "react";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { createMemoryRouter, RouterProvider } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";
import { resource } from "../../api";
import { SessionProvider, useSession } from "../../app/session";
import type { Role } from "../../types";
import {
  AdminReportsPage,
  CoachPerformancePage,
  CoachReviewsPage,
} from "./Reports";

const period = {
  from: "2026-10-01",
  to: "2026-10-14",
  timezone: "Asia/Ho_Chi_Minh",
  as_of: "2026-10-14T09:00:00+07:00",
};

const operations = {
  ...period,
  total_bookings: 2,
  status_counts: { CONFIRMED: 1, COMPLETED: 1 },
  member_attendance_count: 1,
  pt_attendance_count: 2,
  no_show: { numerator: 0, denominator: 1, rate_percent: 0 },
  pending_completion_count: 0,
  utilization: { booked: 60, capacity: 120, rate_percent: 50 },
  cancellation_source_counts: { MEMBER: 1 },
  unclassified_booking_count: 0,
};

const trainerReport = {
  ...period,
  pt_id: 7,
  rating_average: null,
  review_count: 0,
  attendance: { numerator: 0, denominator: 0, rate_percent: null },
  excluded_count: 1,
  unclassified_count: 0,
  pending_completion_count: 0,
};

function Bootstrap({ children, role }: { children: ReactNode; role: Role }) {
  const auth = useSession();
  useEffect(() => {
    auth.login({
      accessToken: `${role.toLowerCase()}-token`,
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
    });
  }, []);
  return auth.session ? children : null;
}

function mount(page: ReactNode, role: Role, path: string) {
  const router = createMemoryRouter([{ path: "*", element: page }], {
    initialEntries: [path],
  });
  return render(
    <SessionProvider>
      <Bootstrap role={role}>
        <RouterProvider router={router} />
      </Bootstrap>
    </SessionProvider>,
  );
}

afterEach(() => vi.restoreAllMocks());

describe("report pages", () => {
  it("filters the admin report by the selected period and room", async () => {
    const list = vi
      .spyOn(resource, "list")
      .mockImplementation(async (_token, path) => ({
        data: (path === "rooms"
          ? [{ id: 2, name: "Strength Studio", status: "AVAILABLE" }]
          : []) as never,
        meta: { page: 1, page_size: 100, total: path === "rooms" ? 1 : 0 },
        request_id: "test",
      }));
    const get = vi
      .spyOn(resource, "get")
      .mockImplementation(async (_token, path) => ({
        data: (path.startsWith("reports/operations")
          ? operations
          : trainerReport) as never,
        request_id: "test",
      }));

    mount(
      <AdminReportsPage />,
      "ADMIN",
      "/management/reports?from=2026-10-01&to=2026-10-14&room_id=2",
    );

    expect(
      await screen.findByText("50.00%", { exact: false }),
    ).toBeInTheDocument();
    expect(screen.getByText("1 HV · 2 PT")).toBeInTheDocument();
    await waitFor(() =>
      expect(
        get.mock.calls.some((call) =>
          call[1].includes(
            "reports/operations?from=2026-10-01&to=2026-10-14&room_id=2",
          ),
        ),
      ).toBe(true),
    );
    expect(
      await screen.findByRole("option", { name: "Strength Studio" }),
    ).toBeInTheDocument();
    expect(list).toHaveBeenCalledWith(
      "admin-token",
      "rooms",
      expect.objectContaining({ get: expect.any(Function) }),
      expect.any(AbortSignal),
    );
  });

  it("rejects an inverted date range before changing the report query", async () => {
    vi.spyOn(resource, "list").mockResolvedValue({
      data: [],
      meta: { page: 1, page_size: 100, total: 0 },
      request_id: "test",
    });
    const get = vi.spyOn(resource, "get").mockResolvedValue({
      data: operations as never,
      request_id: "test",
    });
    mount(
      <AdminReportsPage />,
      "MANAGER",
      "/management/reports?from=2026-10-01&to=2026-10-14",
    );

    await screen.findByText("Báo cáo vận hành");
    fireEvent.change(screen.getByLabelText("Từ ngày"), {
      target: { value: "2026-10-15" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Áp dụng" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "từ ngày không sau đến ngày",
    );
    await waitFor(() => expect(get).toHaveBeenCalledTimes(1));
  });

  it("shows an API failure and retries the same report period", async () => {
    vi.spyOn(resource, "list").mockResolvedValue({
      data: [],
      meta: { page: 1, page_size: 100, total: 0 },
      request_id: "test",
    });
    let reportCalls = 0;
    const get = vi
      .spyOn(resource, "get")
      .mockImplementation(async (_token, path) => {
        if (path.startsWith("reports/operations") && reportCalls++ === 0) {
          throw new Error("Máy chủ tạm thời không phản hồi.");
        }
        return { data: operations as never, request_id: "test" };
      });
    mount(
      <AdminReportsPage />,
      "ADMIN",
      "/management/reports?from=2026-10-01&to=2026-10-14",
    );

    expect(
      await screen.findByText("Máy chủ tạm thời không phản hồi."),
    ).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Thử tải lại" }));

    expect(await screen.findByText("50.00%")).toBeInTheDocument();
    expect(get).toHaveBeenCalledTimes(2);
  });

  it("loads a PT's own report and displays empty-denominator values as N/A", async () => {
    const get = vi
      .spyOn(resource, "get")
      .mockImplementation(async (_token, path) => ({
        data: (path === "pts/me"
          ? { id: 7, specialty: "Strength", approval_status: "APPROVED" }
          : trainerReport) as never,
        request_id: "test",
      }));
    mount(
      <CoachPerformancePage />,
      "PT",
      "/pt/performance?from=2026-10-01&to=2026-10-14",
    );

    expect(await screen.findByText("Đánh giá trung bình")).toBeInTheDocument();
    expect(screen.getAllByText("N/A")).toHaveLength(2);
    expect(
      get.mock.calls.some((call) =>
        call[1].startsWith("reports/pts/7?from=2026-10-01&to=2026-10-14"),
      ),
    ).toBe(true);
  });

  it("shows anonymized PT reviews and requests the next page", async () => {
    vi.spyOn(resource, "get").mockResolvedValue({
      data: {
        id: 7,
        specialty: "Strength",
        approval_status: "APPROVED",
      } as never,
      request_id: "test",
    });
    const list = vi
      .spyOn(resource, "list")
      .mockImplementation(async (_token, path, query) => ({
        data: (path === "pts/7/reviews" && query.get("page") === "1"
          ? [
              {
                rating: 5,
                comment: "Hướng dẫn rõ ràng.",
                created_at: "2026-10-12T10:30:00+07:00",
                reviewer_label: "Hội viên",
              },
            ]
          : []) as never,
        meta: { page: Number(query.get("page")), page_size: 20, total: 21 },
        request_id: "test",
      }));
    mount(<CoachReviewsPage />, "PT", "/pt/reviews");

    expect(await screen.findByText("Hướng dẫn rõ ràng.")).toBeInTheDocument();
    expect(screen.getByText("Hội viên")).toBeInTheDocument();
    expect(screen.queryByText("Member Private Name")).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Trang sau" }));
    await waitFor(() =>
      expect(
        list.mock.calls.some(
          (call) => call[1] === "pts/7/reviews" && call[2].get("page") === "2",
        ),
      ).toBe(true),
    );
  });
});
