import {
  useId,
  useRef,
  cloneElement,
  isValidElement,
  type ButtonHTMLAttributes,
  type ReactNode,
  type InputHTMLAttributes,
  type SelectHTMLAttributes,
  type ChangeEvent,
  type Ref,
} from "react";
import * as Dialog from "@radix-ui/react-dialog";
import * as Dropdown from "@radix-ui/react-dropdown-menu";
import {
  MoreHorizontal,
  X,
  LoaderCircle,
  AlertCircle,
  CheckCircle2,
  Eye,
  EyeOff,
} from "lucide-react";
import { clsx } from "clsx";
import { twMerge } from "tailwind-merge";
import { useState } from "react";

export const cn = (...values: Parameters<typeof clsx>) => twMerge(clsx(values));
export function Button({
  variant = "primary",
  busy,
  className,
  children,
  disabled,
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: "primary" | "secondary" | "danger" | "ghost";
  busy?: boolean;
  ref?: Ref<HTMLButtonElement>;
}) {
  return (
    <button
      {...props}
      disabled={disabled || busy}
      className={cn("btn", `btn-${variant}`, className)}
      aria-busy={busy || undefined}
    >
      {busy && <LoaderCircle size={17} className="spin" />}
      {children}
    </button>
  );
}
export function Field({
  label,
  hint,
  error,
  required,
  children,
}: {
  label: string;
  hint?: string;
  error?: string;
  required?: boolean;
  children: ReactNode;
}) {
  const id = useId();
  const control = isValidElement<{ id?: string; "aria-describedby"?: string }>(
    children,
  )
    ? cloneElement(children, {
        id,
        "aria-describedby": hint || error ? `${id}-help` : undefined,
      })
    : children;
  return (
    <div className="field">
      <span className="field-label">
        <label htmlFor={id}>{label}</label>
        {required && (
          <span aria-hidden="true" className="required">
            {" "}
            *
          </span>
        )}
      </span>
      {control}
      <span id={`${id}-help`}>
        {hint && <span className="field-hint">{hint}</span>}
        {error && (
          <span className="field-error" role="alert">
            {error}
          </span>
        )}
      </span>
    </div>
  );
}
export function Input(props: InputHTMLAttributes<HTMLInputElement>) {
  const [error, setError] = useState("");
  const generated = useId();
  const id = props.id ?? generated;
  function message(input: HTMLInputElement) {
    const v = input.validity;
    if (v.valid) return "";
    if (v.customError) return input.validationMessage;
    if (v.valueMissing) return "Vui lòng điền trường này.";
    if (v.typeMismatch) return "Định dạng chưa đúng. Kiểm tra lại giá trị.";
    if (v.tooShort) return `Nhập ít nhất ${input.minLength} ký tự.`;
    if (v.rangeUnderflow) return `Giá trị tối thiểu là ${input.min}.`;
    if (v.rangeOverflow) return `Giá trị tối đa là ${input.max}.`;
    return "Giá trị chưa hợp lệ.";
  }
  // Native date/time controls can commit their value on blur in WebKit.
  const temporal = ["date", "time", "datetime-local"].includes(
    props.type ?? "",
  );
  return (
    <span className="input-control">
      <input
        {...props}
        id={id}
        className={cn("input", props.className)}
        aria-invalid={!!error || undefined}
        aria-describedby={
          [props["aria-describedby"], error ? `${id}-validation` : ""]
            .filter(Boolean)
            .join(" ") || undefined
        }
        onInput={(e) => {
          props.onInput?.(e);
          if (temporal)
            props.onChange?.(e as unknown as ChangeEvent<HTMLInputElement>);
        }}
        onBlur={(e) => {
          props.onBlur?.(e);
          if (temporal) props.onChange?.(e as ChangeEvent<HTMLInputElement>);
          setError(message(e.currentTarget));
        }}
        onInvalid={(e) => {
          props.onInvalid?.(e);
          setError(message(e.currentTarget));
        }}
        onChange={(e) => {
          props.onChange?.(e);
          if (error) setError(message(e.currentTarget));
        }}
      />
      {error && (
        <span className="field-error" id={`${id}-validation`} role="alert">
          {error}
        </span>
      )}
    </span>
  );
}
export function Select(props: SelectHTMLAttributes<HTMLSelectElement>) {
  return <select {...props} className={cn("input", props.className)} />;
}
export function PasswordInput(props: InputHTMLAttributes<HTMLInputElement>) {
  const [visible, setVisible] = useState(false);
  return (
    <span className="password-input">
      <Input
        {...props}
        type={visible ? "text" : "password"}
        onChange={(e) => {
          if (props.autoComplete === "new-password") {
            const value = e.currentTarget.value;
            e.currentTarget.setCustomValidity(
              new TextEncoder().encode(value).length > 72
                ? "Mật khẩu không được vượt quá 72 byte."
                : value && Array.from(value).length < 8
                  ? "Mật khẩu phải có ít nhất 8 ký tự."
                  : "",
            );
          }
          props.onChange?.(e);
        }}
      />
      <button
        type="button"
        aria-label={visible ? "Ẩn mật khẩu" : "Hiện mật khẩu"}
        aria-pressed={visible}
        onClick={() => setVisible(!visible)}
      >
        {visible ? <EyeOff size={18} /> : <Eye size={18} />}
      </button>
    </span>
  );
}
export function Notice({
  children,
  tone = "error",
}: {
  children?: ReactNode;
  tone?: "error" | "success" | "info";
}) {
  if (!children) return null;
  return (
    <div
      className={`notice notice-${tone}`}
      role={tone === "error" ? "alert" : "status"}
    >
      {tone === "success" ? (
        <CheckCircle2 size={19} />
      ) : (
        <AlertCircle size={19} />
      )}
      <span>{children}</span>
    </div>
  );
}
const statusNames: Record<string, string> = {
  ACTIVE: "Hoạt động",
  LOCKED: "Đã khóa",
  INACTIVE: "Ngừng hoạt động",
  AVAILABLE: "Sẵn sàng",
  MAINTENANCE: "Bảo trì",
  OPEN: "Đang mở",
  CLOSED: "Đã đóng",
  DRAFT: "Bản nháp",
  PENDING: "Chờ duyệt",
  APPROVED: "Đã duyệt",
  REJECTED: "Từ chối",
  SUSPENDED: "Đình chỉ",
};
// A menu item unmounts before a dialog opens; return focus to its persistent trigger.
let pendingActionFocus: HTMLElement | null = null;
export function Badge({ value }: { value: string }) {
  const tone = ["ACTIVE", "AVAILABLE", "OPEN", "APPROVED"].includes(value)
    ? "success"
    : ["PENDING", "DRAFT", "MAINTENANCE"].includes(value)
      ? "warning"
      : "neutral";
  return (
    <span className={`badge badge-${tone}`}>
      <span className="badge-dot" />
      {statusNames[value] ?? value}
    </span>
  );
}
export function Modal({
  title,
  description,
  open,
  onClose,
  children,
  busy,
}: {
  title: string;
  description: string;
  open: boolean;
  onClose: () => void;
  children: ReactNode;
  busy?: boolean;
}) {
  const returnFocus = useRef<HTMLElement | null>(null);
  return (
    <Dialog.Root
      open={open}
      onOpenChange={(value) => {
        if (!value && !busy) onClose();
      }}
    >
      <Dialog.Portal>
        <Dialog.Overlay className="dialog-overlay" />
        <Dialog.Content
          className="dialog-content"
          onOpenAutoFocus={() => {
            returnFocus.current = pendingActionFocus?.isConnected
              ? pendingActionFocus
              : (document.activeElement as HTMLElement);
            pendingActionFocus = null;
          }}
          onCloseAutoFocus={(e) => {
            e.preventDefault();
            returnFocus.current?.focus();
          }}
          onInteractOutside={(e) => e.preventDefault()}
          onEscapeKeyDown={(e) => {
            if (busy) e.preventDefault();
          }}
        >
          <Dialog.Title className="dialog-title">{title}</Dialog.Title>
          <Dialog.Description className="dialog-description">
            {description}
          </Dialog.Description>
          <Dialog.Close asChild>
            <Button
              variant="ghost"
              className="dialog-close"
              disabled={busy}
              aria-label="Đóng"
            >
              <X size={20} />
            </Button>
          </Dialog.Close>
          {children}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
export function ActionMenu({
  name,
  actions,
}: {
  name: string;
  actions: { label: string; run: () => void; danger?: boolean }[];
}) {
  const trigger = useRef<HTMLButtonElement>(null);
  return (
    <Dropdown.Root>
      <Dropdown.Trigger asChild>
        <Button
          ref={trigger}
          variant="ghost"
          className="icon-button"
          aria-label={`Thao tác ${name}`}
        >
          <MoreHorizontal size={20} />
        </Button>
      </Dropdown.Trigger>
      <Dropdown.Portal>
        <Dropdown.Content
          className="action-menu"
          align="end"
          sideOffset={6}
          onCloseAutoFocus={(e) => {
            if (document.querySelector('[role="dialog"]')) e.preventDefault();
          }}
        >
          {actions.map((a) => (
            <Dropdown.Item
              key={a.label}
              className={cn("action-menu-item", a.danger && "text-danger")}
              onSelect={() => {
                pendingActionFocus = trigger.current;
                trigger.current?.focus();
                a.run();
              }}
            >
              {a.label}
            </Dropdown.Item>
          ))}
        </Dropdown.Content>
      </Dropdown.Portal>
    </Dropdown.Root>
  );
}
