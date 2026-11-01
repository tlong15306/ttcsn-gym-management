import { useEffect, useState, type ReactNode } from "react";
import { createMemoryRouter, RouterProvider, Link } from "react-router-dom";
import {
  render,
  screen,
  waitFor,
  within,
  fireEvent,
  act,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import { SessionProvider, useSession } from "./app/session";
import { LoginPanel } from "./features/auth/LoginPanel";
import { Guard } from "./App";
import { UserFormPage, UserActionDialog } from "./features/users/Users";
import { DataForm } from "./components/shared/form";
import { Field, Input, ActionMenu, Modal } from "./components/ui";
import { api, ApiError, request, resource } from "./api";
import type { Session, User, Role } from "./types";
import { MemberFormPage } from "./features/members/Members";
import { TimeSlotFormPage } from "./features/facilities/Facilities";
import { OwnTrainerPage } from "./features/trainers/Trainers";
import { useList } from "./components/shared/list";

const user: User = {
  id: 1,
  username: "admin.demo",
  full_name: "Admin Demo",
  email: "admin@test.local",
  role: "ADMIN",
  status: "ACTIVE",
  created_at: "2026-10-11",
  updated_at: "2026-10-11",
};
const session: Session = { accessToken: "test-token", expiresIn: 900, user };
function Bootstrap({
  children,
  role = "ADMIN",
}: {
  children: ReactNode;
  role?: Role;
}) {
  const auth = useSession();
  useEffect(() => auth.login({ ...session, user: { ...user, role } }), []);
  return auth.session ? children : null;
}
function mount(
  element: ReactNode,
  options: { path?: string; role?: Role } = {},
) {
  const path = options.path ?? "/";
  const router = createMemoryRouter(
    [
      { path, element },
      { path: "/admin/users", element: <h1>Danh sách đã cập nhật</h1> },
      { path: "/management/members", element: <h1>Hội viên đã cập nhật</h1> },
      { path: "/other", element: <h1>Trang khác</h1> },
    ],
    { initialEntries: [path] },
  );
  render(
    <SessionProvider>
      <Bootstrap role={options.role}>
        <RouterProvider router={router} />
      </Bootstrap>
    </SessionProvider>,
  );
  return router;
}
afterEach(() => {
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});
describe("Login states", () => {
  it("keeps passwords blank, shows failures and preserves input", async () => {
    vi.spyOn(api, "login").mockRejectedValue(
      new ApiError(
        401,
        "UNAUTHENTICATED",
        "Tên đăng nhập hoặc mật khẩu không đúng.",
      ),
    );
    const events = userEvent.setup();
    render(<LoginPanel notice="" onLogin={vi.fn()} />);
    expect(screen.getByLabelText("Mật khẩu")).toHaveValue("");
    await events.type(screen.getByLabelText("Tên đăng nhập"), "wrong");
    await events.type(screen.getByLabelText("Mật khẩu"), "Password123!");
    await events.click(screen.getByRole("button", { name: "Đăng nhập" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("không đúng");
    expect(screen.getByLabelText("Tên đăng nhập")).toHaveValue("wrong");
  });
  it("disables duplicate submissions while login is pending", async () => {
    const login = vi
      .spyOn(api, "login")
      .mockImplementation(() => new Promise(() => {}));
    const events = userEvent.setup();
    render(<LoginPanel notice="" onLogin={vi.fn()} />);
    await events.type(screen.getByLabelText("Tên đăng nhập"), "admin");
    await events.type(screen.getByLabelText("Mật khẩu"), "Password123!");
    await events.click(screen.getByRole("button", { name: "Đăng nhập" }));
    expect(
      screen.getByRole("button", { name: /Đang đăng nhập/ }),
    ).toBeDisabled();
    expect(login).toHaveBeenCalledTimes(1);
  });
});
describe("Permissions and session expiry", () => {
  for (const role of ["MEMBER", "PT", "RECEPTIONIST", "MANAGER"] as Role[])
    it(`blocks account page for ${role}`, async () => {
      const router = createMemoryRouter(
        [
          {
            element: <Guard roles={["ADMIN"]} />,
            children: [
              { path: "/admin/users", element: <h1>Private account data</h1> },
            ],
          },
        ],
        { initialEntries: ["/admin/users"] },
      );
      render(
        <SessionProvider>
          <Bootstrap role={role}>
            <RouterProvider router={router} />
          </Bootstrap>
        </SessionProvider>,
      );
      expect(
        await screen.findByRole("heading", {
          name: "Bạn không có quyền truy cập",
        }),
      ).toBeInTheDocument();
      expect(
        screen.queryByText("Private account data"),
      ).not.toBeInTheDocument();
    });
  it("sends a single expiry event on protected 401", async () => {
    const expired = vi.fn();
    window.addEventListener("session-expired", expired);
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            error: { code: "UNAUTHENTICATED", message: "Hết phiên" },
          }),
          { status: 401 },
        ),
      ),
    );
    await expect(request("/api/v1/me", {}, "expired")).rejects.toMatchObject({
      status: 401,
    });
    expect(expired).toHaveBeenCalledTimes(1);
    window.removeEventListener("session-expired", expired);
  });
});
describe("Account forms and confirmations", () => {
  it("retains create form on conflict, then saves the entered values", async () => {
    const create = vi
      .spyOn(api, "createUser")
      .mockRejectedValueOnce(
        new ApiError(409, "INVALID_STATE", "Email đã được sử dụng."),
      )
      .mockResolvedValueOnce({ data: { ...user, id: 2 }, request_id: "test" });
    const events = userEvent.setup();
    mount(<UserFormPage />, { path: "/admin/users/new" });
    await events.type(await screen.findByLabelText("Họ và tên"), "Test New");
    await events.type(screen.getByLabelText("Tên đăng nhập"), "test.new");
    await events.type(screen.getByLabelText("Email"), "new@test.local");
    await events.type(
      screen.getByLabelText("Mật khẩu ban đầu"),
      "Password123!",
    );
    await events.click(screen.getByRole("button", { name: "Tạo tài khoản" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("Email đã");
    expect(screen.getByLabelText("Họ và tên")).toHaveValue("Test New");
    await events.click(screen.getByRole("button", { name: "Tạo tài khoản" }));
    expect(
      await screen.findByRole("heading", { name: "Danh sách đã cập nhật" }),
    ).toBeInTheDocument();
    expect(create).toHaveBeenLastCalledWith(
      "test-token",
      expect.objectContaining({
        username: "test.new",
        full_name: "Test New",
        role: "MEMBER",
      }),
    );
  });
  it("requires matching reset password and waits for API success", async () => {
    const reset = vi
      .spyOn(api, "resetPassword")
      .mockResolvedValue({ data: user, request_id: "test" });
    const done = vi.fn();
    const events = userEvent.setup();
    mount(
      <UserActionDialog
        action={{ user: { ...user, id: 2 }, kind: "password" }}
        close={vi.fn()}
        done={done}
      />,
    );
    await events.type(
      await screen.findByLabelText("Mật khẩu mới"),
      "Password123!",
    );
    await events.type(
      screen.getByLabelText("Nhập lại mật khẩu"),
      "Different123!",
    );
    await events.click(
      screen.getByRole("button", { name: "Đặt lại mật khẩu" }),
    );
    expect(await screen.findByRole("alert")).toHaveTextContent("chưa khớp");
    expect(reset).not.toHaveBeenCalled();
    await events.clear(screen.getByLabelText("Nhập lại mật khẩu"));
    await events.type(
      screen.getByLabelText("Nhập lại mật khẩu"),
      "Password123!",
    );
    await events.click(
      screen.getByRole("button", { name: "Đặt lại mật khẩu" }),
    );
    await waitFor(() => expect(done).toHaveBeenCalledTimes(1));
    expect(reset).toHaveBeenCalledWith("test-token", 2, "Password123!");
  });
  it("uses the reason and selected role for account changes", async () => {
    const change = vi
      .spyOn(api, "changeRole")
      .mockResolvedValue({ data: user, request_id: "test" });
    const events = userEvent.setup();
    mount(
      <UserActionDialog
        action={{ user: { ...user, id: 2 }, kind: "role" }}
        close={vi.fn()}
        done={vi.fn()}
      />,
    );
    await events.selectOptions(
      await screen.findByLabelText("Vai trò mới"),
      "MANAGER",
    );
    await events.type(
      screen.getByLabelText("Lý do thay đổi"),
      "Phân công quản lý",
    );
    await events.click(screen.getByRole("button", { name: "Đổi vai trò" }));
    await waitFor(() =>
      expect(change).toHaveBeenCalledWith(
        "test-token",
        2,
        "MANAGER",
        "Phân công quản lý",
      ),
    );
  });
  it("confirms dirty navigation and can keep entered data", async () => {
    const events = userEvent.setup();
    mount(
      <>
        <Link to="/other">Chuyển trang</Link>
        <DataForm save={async () => {}}>
          <Field label="Họ tên">
            <Input />
          </Field>
        </DataForm>
      </>,
    );
    await events.type(await screen.findByLabelText("Họ tên"), "Unsaved");
    await events.click(screen.getByRole("link", { name: "Chuyển trang" }));
    const dialog = await screen.findByRole("dialog");
    expect(dialog).toHaveTextContent("Bỏ thay đổi chưa lưu?");
    await events.click(within(dialog).getByRole("button", { name: "Ở lại" }));
    expect(screen.getByLabelText("Họ tên")).toHaveValue("Unsaved");
    await events.click(screen.getByRole("link", { name: "Chuyển trang" }));
    await events.click(
      within(await screen.findByRole("dialog")).getByRole("button", {
        name: "Bỏ thay đổi",
      }),
    );
    expect(
      await screen.findByRole("heading", { name: "Trang khác" }),
    ).toBeInTheDocument();
  });
});
describe("Requests", () => {
  it("commits native date input events before form submission", async () => {
    function DateField() {
      const [date, setDate] = useState("");
      return (
        <>
          <Field label="Ngày">
            <Input
              type="date"
              value={date}
              onChange={(e) => setDate(e.target.value)}
            />
          </Field>
          <output>{date}</output>
        </>
      );
    }
    render(<DateField />);
    fireEvent.input(screen.getByLabelText("Ngày"), {
      target: { value: "2027-10-11" },
    });
    expect(screen.getByRole("status")).toHaveTextContent("2027-10-11");
  });
  it("ignores a slow earlier list response after changing filters", async () => {
    let resolveOld: ((v: any) => void) | undefined;
    vi.spyOn(resource, "list").mockImplementation((_token, _path, params) =>
      params.get("status")
        ? Promise.resolve({
            data: [{ id: 2, name: "Current" }],
            meta: { page: 1, page_size: 20, total: 1 },
            request_id: "new",
          })
        : new Promise((resolve) => {
            resolveOld = resolve;
          }),
    );
    function List() {
      const list = useList<{ id: number; name: string }>("rooms");
      return (
        <>
          <button onClick={() => list.filter("status", "AVAILABLE")}>
            Filter
          </button>
          {list.data.map((row) => (
            <span key={row.id}>{row.name}</span>
          ))}
        </>
      );
    }
    mount(<List />);
    const events = userEvent.setup();
    await events.click(await screen.findByRole("button", { name: "Filter" }));
    expect(await screen.findByText("Current")).toBeInTheDocument();
    await act(async () => {
      resolveOld?.({
        data: [{ id: 1, name: "Stale" }],
        meta: { page: 1, page_size: 20, total: 1 },
        request_id: "old",
      });
    });
    expect(screen.queryByText("Stale")).not.toBeInTheDocument();
  });
  it("passes abort signals and server filters unchanged", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: [],
          meta: { page: 2, page_size: 50, total: 0 },
          request_id: "test",
        }),
      ),
    );
    vi.stubGlobal("fetch", fetchMock);
    const abort = new AbortController();
    await resource.list(
      "token",
      "members",
      new URLSearchParams({ q: "name", page: "2", page_size: "50" }),
      abort.signal,
    );
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/members?q=name&page=2&page_size=50",
      expect.objectContaining({ signal: abort.signal }),
    );
  });
});

describe("Management modules", () => {
  it("returns dialog focus to the persistent action menu trigger", async () => {
    function MenuDialog() {
      const [open, setOpen] = useState(false);
      return (
        <>
          <ActionMenu
            name="Demo"
            actions={[{ label: "Open dialog", run: () => setOpen(true) }]}
          />
          {open && (
            <Modal
              open
              title="Dialog"
              description="Demo"
              onClose={() => setOpen(false)}
            >
              <p>Body</p>
            </Modal>
          )}
        </>
      );
    }
    render(<MenuDialog />);
    const events = userEvent.setup();
    const trigger = screen.getByRole("button", { name: "Thao tác Demo" });
    await events.click(trigger);
    await events.click(
      await screen.findByRole("menuitem", { name: "Open dialog" }),
    );
    await screen.findByRole("dialog");
    await events.keyboard("{Escape}");
    await waitFor(() => expect(trigger).toHaveFocus());
  });
  it("links an eligible MEMBER without submitting a selectable role", async () => {
    vi.spyOn(resource, "list").mockResolvedValue({
      data: [{ ...user, id: 9, role: "MEMBER" }],
      meta: { page: 1, page_size: 20, total: 1 },
      request_id: "lookup",
    });
    const save = vi
      .spyOn(resource, "save")
      .mockResolvedValue({ data: { id: 1 }, request_id: "saved" });
    mount(<MemberFormPage />, {
      path: "/management/members/new",
      role: "RECEPTIONIST",
    });
    const events = userEvent.setup();
    await events.selectOptions(
      await screen.findByLabelText("Cách tạo"),
      "existing",
    );
    await screen.findByRole("option", { name: /Admin Demo/ });
    await events.selectOptions(
      screen.getByLabelText("Tài khoản MEMBER chưa có hồ sơ"),
      "9",
    );
    await events.type(screen.getByLabelText("Mã hội viên"), "HV-TEST");
    fireEvent.input(screen.getByLabelText("Ngày tham gia"), {
      target: { value: "2026-10-11" },
    });
    fireEvent.input(screen.getByLabelText("Hạn thẻ"), {
      target: { value: "2027-10-11" },
    });
    await events.click(screen.getByRole("button", { name: "Tạo hội viên" }));
    await waitFor(() =>
      expect(save).toHaveBeenCalledWith(
        "test-token",
        "members",
        expect.objectContaining({
          user_id: 9,
          membership_expiry: "2027-10-11",
        }),
      ),
    );
    expect(save.mock.calls[0][2]).not.toHaveProperty("role");
    expect(save.mock.calls[0][2]).not.toHaveProperty("account");
    expect(
      await screen.findByRole("heading", { name: "Hội viên đã cập nhật" }),
    ).toBeInTheDocument();
  });
  it("rejects reversed slot times while preserving the form", async () => {
    const save = vi.spyOn(resource, "save");
    mount(<TimeSlotFormPage />, { path: "/management/time-slots/new" });
    const events = userEvent.setup();
    await events.type(await screen.findByLabelText("Tên ca"), "Morning");
    fireEvent.input(screen.getByLabelText("Giờ bắt đầu"), {
      target: { value: "10:00" },
    });
    fireEvent.input(screen.getByLabelText("Giờ kết thúc"), {
      target: { value: "09:00" },
    });
    await events.click(screen.getByRole("button", { name: "Tạo ca" }));
    expect(await screen.findByRole("alert")).toHaveTextContent(
      "Giờ kết thúc phải sau",
    );
    expect(save).not.toHaveBeenCalled();
    expect(screen.getByLabelText("Tên ca")).toHaveValue("Morning");
  });
  it("keeps a pending PT profile read-only and hides upload/submit", async () => {
    vi.spyOn(resource, "get").mockImplementation((_token, path) =>
      Promise.resolve({
        data:
          path === "pts/me"
            ? {
                id: 3,
                user_id: 1,
                user: { ...user, role: "PT" },
                specialty: "Gym",
                experience_text: "Demo",
                has_certificate: true,
                approval_status: "PENDING",
                decision_reason: "",
                decided_at: null,
              }
            : [],
        request_id: "pt",
      }),
    );
    mount(<OwnTrainerPage />, { path: "/pt/profile", role: "PT" });
    expect(await screen.findByText("Chờ duyệt")).toBeInTheDocument();
    expect(screen.queryByLabelText("Chuyên môn")).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Gửi hồ sơ xét duyệt" }),
    ).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Tải chứng chỉ" })).toBeEnabled();
  });
});
