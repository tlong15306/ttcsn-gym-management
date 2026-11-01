import {
  useEffect,
  useRef,
  useState,
  type FormEvent,
  type ReactNode,
} from "react";
import { useNavigate, useBlocker } from "react-router-dom";
import { Button, Modal, Notice } from "../ui";

export function DataForm({
  children,
  save,
  back,
  submitLabel = "Lưu thay đổi",
  onSuccess,
}: {
  children: ReactNode;
  save: () => Promise<unknown>;
  back?: string;
  submitLabel?: string;
  onSuccess?: () => void;
}) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [dirty, setDirty] = useState(false);
  const dirtyRef = useRef(false);
  const navigate = useNavigate();
  const saving = useRef(false);
  const blocker = useBlocker(
    ({ currentLocation, nextLocation }) =>
      dirtyRef.current && currentLocation.pathname !== nextLocation.pathname,
  );
  useEffect(() => {
    function leave(e: BeforeUnloadEvent) {
      if (dirty) e.preventDefault();
    }
    window.addEventListener("beforeunload", leave);
    return () => window.removeEventListener("beforeunload", leave);
  }, [dirty]);
  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (saving.current) return;
    saving.current = true;
    setBusy(true);
    setError("");
    setSuccess("");
    try {
      await save();
      dirtyRef.current = false;
      setDirty(false);
      setSuccess("Đã lưu thay đổi.");
      onSuccess?.();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Không thể lưu dữ liệu.");
    } finally {
      saving.current = false;
      setBusy(false);
    }
  }
  return (
    <>
      <form
        onSubmit={submit}
        onChange={() => {
          dirtyRef.current = true;
          setDirty(true);
          setSuccess("");
        }}
        className="data-form"
      >
        <Notice>{error}</Notice>
        <Notice tone="success">{success}</Notice>
        <fieldset disabled={busy}>{children}</fieldset>
        <div className="form-actions">
          {back && (
            <Button
              type="button"
              variant="secondary"
              disabled={busy}
              onClick={() => navigate(back)}
            >
              Hủy
            </Button>
          )}
          <Button type="submit" busy={busy}>
            {busy ? "Đang lưu…" : submitLabel}
          </Button>
        </div>
      </form>
      <Modal
        title="Bỏ thay đổi chưa lưu?"
        description="Dữ liệu vừa nhập sẽ không được lưu."
        open={blocker.state === "blocked"}
        onClose={() => blocker.reset?.()}
        busy={busy}
      >
        <div className="form-actions">
          <Button variant="secondary" onClick={() => blocker.reset?.()}>
            Ở lại
          </Button>
          <Button
            variant="danger"
            onClick={() => {
              dirtyRef.current = false;
              setDirty(false);
              blocker.proceed?.();
            }}
          >
            Bỏ thay đổi
          </Button>
        </div>
      </Modal>
    </>
  );
}
