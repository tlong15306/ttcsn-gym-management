import { useEffect, useState } from "react";
import { useAuth } from "../../app/session";
import type { Room } from "../facilities/Facilities";
import { ApiError, request } from "../../api";
import { Button, Field, Input, Modal, Notice } from "../../components/ui";

export function RoomImage({
  room,
  detail = false,
}: {
  room: Room;
  detail?: boolean;
}) {
  const { session } = useAuth();
  const [url, setUrl] = useState("");
  const [error, setError] = useState("");
  useEffect(() => {
    const abort = new AbortController();
    let objectURL = "";
    setUrl("");
    setError("");
    if (!room.has_cover) return;
    fetch(
      `/api/v1/rooms/${room.id}/cover-image?size=${detail ? "detail" : "card"}`,
      {
        headers: { Authorization: `Bearer ${session.accessToken}` },
        signal: abort.signal,
      },
    )
      .then(async (r) => {
        if (!r.ok) {
          if (r.status === 401)
            window.dispatchEvent(new Event("session-expired"));
          throw new ApiError(
            r.status,
            "IMAGE_ERROR",
            "Không tải được ảnh phòng.",
          );
        }
        return r.blob();
      })
      .then((blob) => {
        if (!abort.signal.aborted) {
          objectURL = URL.createObjectURL(blob);
          setUrl(objectURL);
        }
      })
      .catch((e) => {
        if (!abort.signal.aborted) setError(e.message);
      });
    return () => {
      abort.abort();
      if (objectURL) URL.revokeObjectURL(objectURL);
    };
  }, [
    room.id,
    room.cover_version,
    room.has_cover,
    detail,
    session.accessToken,
  ]);
  return (
    <div className="room-photo">
      {url ? (
        <img
          src={url}
          alt={room.cover_alt_text || room.name}
          loading={detail ? "eager" : "lazy"}
        />
      ) : (
        <div className="photo-placeholder" role="status">
          {error || (room.has_cover ? "Đang tải ảnh…" : "Phòng chưa có ảnh")}
        </div>
      )}
      {room.cover_is_demo && (
        <span className="demo-photo-label">Ảnh minh họa</span>
      )}
    </div>
  );
}

export function RoomCoverManager({
  room,
  refresh,
}: {
  room: Room;
  refresh: () => void;
}) {
  const { session } = useAuth();
  const [file, setFile] = useState<File>();
  const [preview, setPreview] = useState("");
  useEffect(() => {
    if (!file) {
      setPreview("");
      return;
    }
    const url = URL.createObjectURL(file);
    setPreview(url);
    return () => URL.revokeObjectURL(url);
  }, [file]);
  const [alt, setAlt] = useState(room.cover_alt_text || "");
  const [demo, setDemo] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [confirm, setConfirm] = useState(false);
  async function save(remove = false) {
    setBusy(true);
    setError("");
    try {
      const form = new FormData();
      if (file) form.set("file", file);
      form.set("alt_text", alt);
      form.set("is_demo", String(demo));
      await request(
        `/api/v1/rooms/${room.id}/cover-image`,
        { method: remove ? "DELETE" : "POST", body: remove ? undefined : form },
        session.accessToken,
      );
      setFile(undefined);
      setConfirm(false);
      refresh();
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  }
  return (
    <section className="panel cover-manager">
      <h2>Ảnh đại diện phòng</h2>
      <RoomImage room={room} />
      {preview && (
        <div className="room-photo">
          <img src={preview} alt="Ảnh đang chọn, chưa lưu" />
        </div>
      )}
      <Notice>{error}</Notice>
      <form
        onSubmit={(e) => {
          e.preventDefault();
          if (file && file.size <= 5 * 1024 * 1024) void save();
          else setError("Chọn ảnh JPEG/PNG tối đa 5 MB.");
        }}
      >
        <Field label="Ảnh JPEG hoặc PNG" required>
          <Input
            type="file"
            accept="image/jpeg,image/png"
            required
            onChange={(e) => setFile(e.target.files?.[0])}
          />
        </Field>
        <Field label="Mô tả ảnh" required>
          <Input
            value={alt}
            maxLength={300}
            required
            onChange={(e) => setAlt(e.target.value)}
          />
        </Field>
        <label>
          <input
            type="checkbox"
            checked={demo}
            onChange={(e) => setDemo(e.target.checked)}
          />{" "}
          Đây là ảnh minh họa
        </label>
        <div className="form-actions">
          <Button busy={busy}>Lưu ảnh</Button>
          {room.has_cover && (
            <Button
              type="button"
              variant="danger"
              onClick={() => setConfirm(true)}
            >
              Xóa ảnh
            </Button>
          )}
        </div>
      </form>
      <Modal
        open={confirm}
        onClose={() => setConfirm(false)}
        busy={busy}
        title="Xóa ảnh phòng?"
        description="Phòng vẫn giữ nguyên lịch hoạt động và các đặt lịch."
      >
        <Button variant="danger" busy={busy} onClick={() => void save(true)}>
          Xác nhận xóa ảnh
        </Button>
      </Modal>
    </section>
  );
}
