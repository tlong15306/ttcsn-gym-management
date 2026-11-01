import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { resource } from "../../api";
import { AttendancePanel } from "./Attendance";
const state = vi.hoisted(() => ({ role: "RECEPTIONIST" }));
vi.mock("../../app/session", () => ({ useAuth: () => ({ session: { accessToken: "test", user: { role: state.role } } }) }));
const booking = { id: 1, room_id: 1, room_name: "Strength", member_name: "Member Demo", pt_name: "PT Demo", pt_id: 2, status: "CONFIRMED", starts_at: "2026-10-12T09:00:00+07:00", ends_at: "2026-10-12T10:00:00+07:00", server_time: "2026-10-12T09:01:00+07:00", cancellation_reason: "" };
let now = booking.server_time;
beforeEach(() => {
 state.role = "RECEPTIONIST"; now = booking.server_time;
 vi.spyOn(resource,"get").mockImplementation(async () => ({ data: { participants: [{ participant_type: "MEMBER", name: "Member Demo", attendance: null }, { participant_type: "PT", name: "PT Demo", attendance: null }], server_time: now }, request_id: "test" }) as never);
});
afterEach(() => vi.restoreAllMocks());
describe("Attendance", () => {
 it("shows read-only participants for members and PTs", async () => {
  for (const role of ["MEMBER","PT"]) {
   state.role=role;const view=render(<AttendancePanel booking={booking} onChange={vi.fn()} />);
   await screen.findByText("Member Demo"); expect(screen.queryByRole("button",{name:"Điểm danh"})).toBeNull();view.unmount();
  }
 });
 it("confirms an individual participant and submits once while busy", async () => {
  let resolve!: () => void;
  const save=vi.spyOn(resource,"save").mockImplementation(() => new Promise(r=>{ resolve=()=>r({data:{},request_id:"test"} as never); }));
  const changed=vi.fn();render(<AttendancePanel booking={booking} onChange={changed} />);
  fireEvent.click((await screen.findAllByRole("button",{name:"Điểm danh"}))[0]);
  const confirm=screen.getByRole("button",{name:"Xác nhận điểm danh"});fireEvent.click(confirm);fireEvent.click(confirm);
  expect(save).toHaveBeenCalledTimes(1);expect(save).toHaveBeenCalledWith("test","bookings/1/attendance",{participant_type:"MEMBER",override_reason:""});
  resolve();await waitFor(()=>expect(changed).toHaveBeenCalled());expect(await screen.findByText("Đã điểm danh Member Demo.")).toBeInTheDocument();
 });
 it("requires an override reason early and displays API failure", async () => {
  now="2026-10-12T08:00:00+07:00";vi.spyOn(resource,"save").mockRejectedValue(new Error("Buổi đã kết thúc"));
  render(<AttendancePanel booking={booking} onChange={vi.fn()} />);
  fireEvent.click((await screen.findAllByRole("button",{name:"Ghi sớm có lý do"}))[0]);
  const reason=screen.getByRole("textbox");expect(reason).toBeRequired();fireEvent.change(reason,{target:{value:"Synthetic early arrival"}});
  fireEvent.click(screen.getByRole("button",{name:"Xác nhận điểm danh"}));expect(await screen.findByText("Buổi đã kết thúc")).toBeInTheDocument();
 });
 it("does not allow new records after the server end time", async () => {
  now=booking.ends_at;render(<AttendancePanel booking={booking} onChange={vi.fn()} />);await screen.findByText("Member Demo");expect(screen.queryByRole("button",{name:"Điểm danh"})).toBeNull();
 });
});
