import { useEffect, useRef, useState } from "react";
import { useAuth } from "../../app/session";
import { resource } from "../../api";
import { useRecord, RecordState } from "../../components/shared/record";
import { Button, Field, Input, Modal, Notice } from "../../components/ui";
import type { Booking } from "./Booking";
import { vietnamDateTime } from "../../utils/vietnamTime";

type Participant = { participant_type: string; name: string; attendance: null | { checked_in_at: string; recorder_name: string; override_reason: string } };
type Summary = { participants: Participant[]; server_time: string };
export function AttendancePanel({ booking, onChange }: { booking: Booking; onChange: () => void }) {
  const { session } = useAuth();
  const record = useRecord<Summary>(`bookings/${booking.id}/attendance`);
  useEffect(() => { record.refresh(); }, [booking.server_time]);
  const [selected, setSelected] = useState<Participant>();
  const [reason, setReason] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const submitting = useRef(false);
  const operator = ["ADMIN", "MANAGER", "RECEPTIONIST"].includes(session.user.role);
  const now = new Date(record.data?.server_time || booking.server_time).getTime();
  const early = now < new Date(booking.starts_at).getTime() - 900000;
  const allowed = booking.status === "CONFIRMED" && now < new Date(booking.ends_at).getTime();
  async function submit() {
    if (!selected || submitting.current) return;
    submitting.current = true; setBusy(true); setError("");
    try {
      await resource.save(session.accessToken, `bookings/${booking.id}/attendance`, { participant_type: selected.participant_type, override_reason: reason });
      setNotice(`Đã điểm danh ${selected.name}.`); setSelected(undefined); record.refresh(); onChange();
    } catch (e) { setError((e as Error).message); record.refresh(); onChange(); }
    finally { submitting.current = false; setBusy(false); }
  }
  return <section className="attendance-panel" aria-label="Điểm danh buổi tập">
    <h2>Điểm danh buổi tập</h2>
    <p>Hội viên và PT được ghi nhận riêng. Kết quả buổi tập dựa trên điểm danh Hội viên.</p>
    <Notice>{notice}</Notice>
    {!record.data ? <RecordState {...record} retry={record.refresh} /> : <>
      <Notice>{record.error}</Notice>
      {record.data.participants.map(p => <div className="attendance-participant" key={p.participant_type}>
        <div><strong>{p.name}</strong><p>{p.participant_type === "MEMBER" ? "Hội viên" : "Huấn luyện viên"}</p>
        {p.attendance ? <p>Đã điểm danh lúc {vietnamDateTime(p.attendance.checked_in_at)} · {p.attendance.recorder_name}{p.attendance.override_reason && ` · Ngoại lệ: ${p.attendance.override_reason}`}</p> : <p>Chưa điểm danh</p>}</div>
        {operator && allowed && !p.attendance && <Button disabled={busy} onClick={() => { setSelected(p); setReason(""); setError(""); }}>{early ? "Ghi sớm có lý do" : "Điểm danh"}</Button>}
      </div>)}
      {!allowed && operator && <p>Chỉ điểm danh buổi đã xác nhận, trước giờ kết thúc.</p>}
    </>}
    <Modal open={!!selected} onClose={() => setSelected(undefined)} title={early ? "Xác nhận điểm danh sớm" : "Xác nhận điểm danh"} description={`${selected?.name || ""} · ${booking.room_name} · #${booking.id}`} busy={busy}>
      <form onSubmit={e => { e.preventDefault(); void submit(); }}><Notice>{error}</Notice>
      <Field label={early ? "Lý do ngoại lệ (bắt buộc)" : "Ghi chú ngoại lệ (nếu có)"}><Input required={early} maxLength={1000} value={reason} onChange={e => setReason(e.target.value)} /></Field>
      <Button type="submit" busy={busy}>Xác nhận điểm danh</Button></form>
    </Modal>
  </section>;
}
