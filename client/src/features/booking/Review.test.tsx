import { useEffect, type ReactNode } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { ApiError, resource } from "../../api";
import { SessionProvider, useSession } from "../../app/session";
import type { Role } from "../../types";
import type { Booking } from "./Booking";
import { RatingSummary, ReviewPanel } from "./Review";

const booking: Booking = {
  id: 44,
  room_id: 2,
  room_name: "Strength Studio",
  member_name: "Member Private Name",
  pt_name: "Coach Demo",
  pt_id: 7,
  status: "COMPLETED",
  starts_at: "2026-10-12T09:00:00+07:00",
  ends_at: "2026-10-12T10:00:00+07:00",
  server_time: "2026-10-12T10:30:00+07:00",
  cancellation_reason: "",
};

const present = {
  participants: [
    {
      participant_type: "MEMBER" as const,
      attendance: { checked_in_at: "2026-10-12T09:00:00+07:00" },
    },
    {
      participant_type: "PT" as const,
      attendance: { checked_in_at: "2026-10-12T09:00:00+07:00" },
    },
  ],
  server_time: booking.server_time,
};

function Auth({ children, role }: { children: ReactNode; role: Role }) {
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

function mount(role: Role = "MEMBER") {
  return render(
    <SessionProvider>
      <Auth role={role}>
        <ReviewPanel booking={booking} />
      </Auth>
    </SessionProvider>,
  );
}

afterEach(() => vi.restoreAllMocks());

describe("PT review", () => {
  it("lets an eligible member submit a keyboard-accessible rating and reads back the saved review", async () => {
    vi.spyOn(resource, "get").mockImplementation(async (_token, path) => ({
      data: (path.endsWith("/attendance") ? present : null) as never,
      request_id: "test",
    }));
    const save = vi.spyOn(resource, "save").mockResolvedValue({
      data: {
        rating: 5,
        comment: "Huấn luyện viên hướng dẫn rõ ràng.",
        created_at: "2026-10-12T10:30:00+07:00",
        reviewer_label: "Hội viên",
      },
      request_id: "test",
    });

    mount();
    fireEvent.click(await screen.findByLabelText("5 sao"));
    fireEvent.change(screen.getByRole("textbox", { name: "Nhận xét" }), {
      target: { value: "Huấn luyện viên hướng dẫn rõ ràng." },
    });
    fireEvent.click(screen.getByRole("button", { name: "Gửi đánh giá" }));

    expect(await screen.findByText("Đã đánh giá")).toBeInTheDocument();
    expect(save).toHaveBeenCalledWith("member-token", "bookings/44/review", {
      rating: 5,
      comment: "Huấn luyện viên hướng dẫn rõ ràng.",
    });
    expect(
      screen.getByText("Huấn luyện viên hướng dẫn rõ ràng."),
    ).toBeInTheDocument();
    expect(screen.getByText("Hội viên", { exact: false })).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Gửi đánh giá" }),
    ).not.toBeInTheDocument();
  });

  it("does not offer a form if either participant attendance is missing", async () => {
    vi.spyOn(resource, "get").mockImplementation(async (_token, path) => ({
      data: (path.endsWith("/attendance")
        ? {
            ...present,
            participants: [
              present.participants[0],
              { ...present.participants[1], attendance: null },
            ],
          }
        : null) as never,
      request_id: "test",
    }));

    mount();

    expect(
      await screen.findByText(/Hội viên và PT cần có mặt/),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Gửi đánh giá" }),
    ).not.toBeInTheDocument();
  });

  it("shows an already submitted review to the assigned PT without exposing an edit action", async () => {
    vi.spyOn(resource, "get").mockResolvedValue({
      data: {
        rating: 4,
        comment: "Buổi tập có nhịp độ tốt.",
        created_at: "2026-10-12T10:30:00+07:00",
        reviewer_label: "Hội viên",
      } as never,
      request_id: "test",
    });

    mount("PT");

    expect(
      await screen.findByText("Buổi tập có nhịp độ tốt."),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Gửi đánh giá" }),
    ).not.toBeInTheDocument();
    expect(screen.queryByText("Member Private Name")).not.toBeInTheDocument();
  });

  it("shows the server error and keeps the form editable after a failed submit", async () => {
    vi.spyOn(resource, "get").mockImplementation(async (_token, path) => ({
      data: (path.endsWith("/attendance") ? present : null) as never,
      request_id: "test",
    }));
    vi.spyOn(resource, "save").mockRejectedValue(
      new ApiError(422, "REVIEW_NOT_ELIGIBLE", "Chưa đủ điều kiện đánh giá."),
    );

    mount();
    fireEvent.click(await screen.findByLabelText("4 sao"));
    fireEvent.click(screen.getByRole("button", { name: "Gửi đánh giá" }));

    expect(
      await screen.findByText("Chưa đủ điều kiện đánh giá."),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Gửi đánh giá" })).toBeEnabled();
  });

  it("uses N/A for PTs without reviews and a compact average otherwise", () => {
    const { rerender } = render(<RatingSummary average={null} count={0} />);
    expect(screen.getByText("Chưa có đánh giá")).toBeInTheDocument();
    rerender(<RatingSummary average={4.25} count={12} />);
    expect(
      screen.getByLabelText("4.3 trên 5 sao, 12 đánh giá"),
    ).toBeInTheDocument();
  });
});
