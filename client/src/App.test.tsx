import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { LoginPanel } from "./App";
import { api } from "./api";

describe("LoginPanel", () => {
  it("sends the entered credentials", async () => {
    const login = vi.spyOn(api, "login").mockResolvedValue({
      accessToken: "token",
      expiresIn: 900,
      user: {
        id: 1,
        username: "admin",
        email: "a@example.com",
        full_name: "Admin",
        role: "ADMIN",
        status: "ACTIVE",
        created_at: "",
        updated_at: "",
      },
    });
    const onLogin = vi.fn();
    render(<LoginPanel notice="" onLogin={onLogin} />);
    fireEvent.change(screen.getByLabelText("Tên đăng nhập"), {
      target: { value: "tester" },
    });
    fireEvent.change(screen.getByLabelText("Mật khẩu"), {
      target: { value: "Password123!" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Đăng nhập" }));
    await vi.waitFor(() =>
      expect(login).toHaveBeenCalledWith("tester", "Password123!"),
    );
    expect(onLogin).toHaveBeenCalledTimes(1);
  });
});
