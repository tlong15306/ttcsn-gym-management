import { useState, useRef, type ReactNode } from "react";
import { Button, Modal, Notice } from "../ui";

export function OperationDialog({
  title,
  description,
  children,
  save,
  close,
  done,
  warning,
  danger,
}: {
  title: string;
  description: string;
  children?: ReactNode;
  save: () => Promise<unknown>;
  close: () => void;
  done: () => void;
  warning?: string;
  danger?: boolean;
}) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const saving = useRef(false);
  return (
    <Modal
      title={title}
      description={description}
      open
      onClose={close}
      busy={busy}
    >
      <Notice tone="info">{warning}</Notice>
      <Notice>{error}</Notice>
      <form
        onSubmit={async (e) => {
          e.preventDefault();
          if (saving.current) return;
          saving.current = true;
          setBusy(true);
          setError("");
          try {
            await save();
            done();
          } catch (e) {
            setError((e as Error).message);
          } finally {
            saving.current = false;
            setBusy(false);
          }
        }}
      >
        <fieldset disabled={busy}>{children}</fieldset>
        <div className="form-actions">
          <Button
            type="button"
            variant="secondary"
            onClick={close}
            disabled={busy}
          >
            Hủy
          </Button>
          <Button busy={busy} variant={danger ? "danger" : "primary"}>
            {title}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
