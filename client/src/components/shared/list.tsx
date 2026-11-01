import { useEffect, useRef, useState, type ReactNode } from "react";
import { useSearchParams } from "react-router-dom";
import {
  Search,
  ChevronLeft,
  ChevronRight,
  SearchX,
  Plus,
  ArrowLeft,
} from "lucide-react";
import { Link } from "react-router-dom";
import { resource } from "../../api";
import type { PageMeta } from "../../types";
import { Button, Notice, Select, Input } from "../ui";
import { useAuth } from "../../app/session";

export function PageHeader({
  title,
  description,
  action,
  back,
}: {
  title: string;
  description: string;
  action?: { label: string; to: string };
  back?: string;
}) {
  return (
    <div className="page-header">
      <div>
        {back && (
          <Link to={back} className="back-link">
            <ArrowLeft size={16} />
            Quay lại danh sách
          </Link>
        )}
        <h1>{title}</h1>
        <p>{description}</p>
      </div>
      {action && (
        <Link to={action.to} className="btn btn-primary">
          <Plus size={18} />
          {action.label}
        </Link>
      )}
    </div>
  );
}
export function useList<T>(path: string) {
  const { session } = useAuth();
  const [params, setParams] = useSearchParams();
  const [data, setData] = useState<T[]>([]);
  const [meta, setMeta] = useState<PageMeta>({
    page: 1,
    page_size: 20,
    total: 0,
  });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [version, setVersion] = useState(0);
  const sequence = useRef(0);
  const query = params.toString();
  useEffect(() => {
    const abort = new AbortController();
    const current = ++sequence.current;
    setLoading(true);
    setError("");
    resource
      .list<T>(
        session.accessToken,
        path,
        new URLSearchParams(query),
        abort.signal,
      )
      .then((result) => {
        if (current !== sequence.current) return;
        const next = result.meta!;
        if (
          next.total > 0 &&
          next.page > Math.ceil(next.total / next.page_size)
        ) {
          setParams(
            (p) => {
              p.set("page", String(Math.ceil(next.total / next.page_size)));
              return p;
            },
            { replace: true },
          );
          return;
        }
        setData(result.data);
        setMeta(next);
      })
      .catch((e) => {
        if (!abort.signal.aborted && current === sequence.current)
          setError(e.message || "Không thể tải dữ liệu.");
      })
      .finally(() => {
        if (!abort.signal.aborted && current === sequence.current)
          setLoading(false);
      });
    return () => abort.abort();
  }, [session.accessToken, path, query, version]);
  function filter(key: string, value: string) {
    setParams((p) => {
      if (value) p.set(key, value);
      else p.delete(key);
      if (key !== "page") p.set("page", "1");
      return p;
    });
  }
  return {
    data,
    meta,
    loading,
    error,
    params,
    filter,
    refresh: () => setVersion((v) => v + 1),
    clear: () => setParams({}),
  };
}
export function SearchBox({
  value,
  onChange,
  placeholder,
}: {
  value: string;
  onChange: (s: string) => void;
  placeholder: string;
}) {
  const [draft, setDraft] = useState(value);
  useEffect(() => setDraft(value), [value]);
  useEffect(() => {
    if (draft === value) return;
    const timer = setTimeout(() => onChange(draft), 300);
    return () => clearTimeout(timer);
  }, [draft, value, onChange]);
  return (
    <span className="search-box">
      <Search size={18} />
      <Input
        aria-label="Tìm kiếm"
        placeholder={placeholder}
        value={draft}
        onChange={(e) => setDraft(e.target.value)}
      />
    </span>
  );
}
export function ListState({
  loading,
  error,
  empty,
  filtered,
  retry,
  children,
}: {
  loading: boolean;
  error: string;
  empty: boolean;
  filtered: boolean;
  retry: () => void;
  children: ReactNode;
}) {
  if (error)
    return (
      <div className="state-panel">
        <Notice>{error}</Notice>
        <Button variant="secondary" onClick={retry}>
          Thử lại
        </Button>
      </div>
    );
  if (loading && empty)
    return (
      <div
        className="skeleton-list"
        aria-label="Đang tải danh sách"
        role="status"
      >
        {[1, 2, 3, 4].map((n) => (
          <div className="skeleton-row" key={n} />
        ))}
      </div>
    );
  if (empty)
    return (
      <div className="empty-state">
        <SearchX size={36} />
        <h2>{filtered ? "Không có kết quả phù hợp" : "Chưa có dữ liệu"}</h2>
        <p>
          {filtered
            ? "Thử thay đổi từ khóa hoặc xóa bộ lọc."
            : "Thêm dữ liệu đầu tiên để bắt đầu quản lý."}
        </p>
      </div>
    );
  return (
    <div aria-busy={loading}>
      {loading && (
        <div className="refresh-status" role="status">
          Đang cập nhật…
        </div>
      )}
      {children}
    </div>
  );
}
export function Pagination({
  meta,
  change,
}: {
  meta: PageMeta;
  change: (key: string, value: string) => void;
}) {
  const pages = Math.max(1, Math.ceil(meta.total / meta.page_size));
  return (
    <div className="pagination">
      <span>
        {meta.total
          ? `${(meta.page - 1) * meta.page_size + 1}–${Math.min(meta.page * meta.page_size, meta.total)}`
          : "0"}{" "}
        / {meta.total} kết quả
      </span>
      <div>
        <Select
          aria-label="Số dòng mỗi trang"
          value={meta.page_size}
          onChange={(e) => change("page_size", e.target.value)}
        >
          {[20, 50, 100].map((n) => (
            <option key={n} value={n}>
              {n} / trang
            </option>
          ))}
        </Select>
        <Button
          variant="secondary"
          aria-label="Trang trước"
          disabled={meta.page <= 1}
          onClick={() => change("page", String(meta.page - 1))}
        >
          <ChevronLeft size={17} />
        </Button>
        <span>
          {meta.page} / {pages}
        </span>
        <Button
          variant="secondary"
          aria-label="Trang sau"
          disabled={meta.page >= pages}
          onClick={() => change("page", String(meta.page + 1))}
        >
          <ChevronRight size={17} />
        </Button>
      </div>
    </div>
  );
}
