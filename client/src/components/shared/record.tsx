import { useEffect, useState } from "react";
import { resource } from "../../api";
import { useAuth } from "../../app/session";
import { Button, Notice } from "../ui";
import type { ReactNode } from "react";

export function useRecord<T>(path: string | null) {
  const { session } = useAuth();
  const [data, setData] = useState<T | null>(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(!!path);
  const [revision, setRevision] = useState(0);
  const [missing, setMissing] = useState(false);
  useEffect(() => {
    if (!path) {
      setLoading(false);
      return;
    }
    const abort = new AbortController();
    setLoading(true);
    setError("");
    setMissing(false);
    resource
      .get<T>(session.accessToken, path, abort.signal)
      .then((r) => {
        if (!abort.signal.aborted) setData(r.data);
      })
      .catch((e) => {
        if (!abort.signal.aborted) {
          setError(e.message);
          setMissing(e.status === 404);
        }
      })
      .finally(() => {
        if (!abort.signal.aborted) setLoading(false);
      });
    return () => abort.abort();
  }, [path, session.accessToken, revision]);
  return {
    data,
    setData,
    error,
    loading,
    missing,
    refresh: () => setRevision((v) => v + 1),
  };
}
export function RecordState({
  loading,
  error,
  retry,
}: {
  loading: boolean;
  error: string;
  retry: () => void;
}) {
  if (loading)
    return (
      <div className="skeleton-list" role="status" aria-label="Đang tải hồ sơ">
        <div className="skeleton-row" />
        <div className="skeleton-row" />
      </div>
    );
  return (
    <div className="state-panel">
      <Notice>{error || "Không tìm thấy hồ sơ."}</Notice>
      <Button variant="secondary" onClick={retry}>
        Thử lại
      </Button>
    </div>
  );
}
export function DataTable<T extends { id: number }>({
  rows,
  columns,
  actions,
}: {
  rows: T[];
  columns: { label: string; render: (row: T) => ReactNode }[];
  actions: (row: T) => ReactNode;
}) {
  return (
    <>
      <div className="desktop-table">
        <table>
          <thead>
            <tr>
              {columns.map((c) => (
                <th key={c.label}>{c.label}</th>
              ))}
              <th className="actions-column">
                <span className="sr-only">Thao tác</span>
              </th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.id}>
                {columns.map((c) => (
                  <td key={c.label}>{c.render(row)}</td>
                ))}
                <td className="actions-column">{actions(row)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <div className="mobile-cards">
        {rows.map((row) => (
          <article className="record-card" key={row.id}>
            <div className="record-card-header">
              {columns[0].render(row)}
              {actions(row)}
            </div>
            <dl>
              {columns.slice(1).map((c) => (
                <div key={c.label}>
                  <dt>{c.label}</dt>
                  <dd>{c.render(row)}</dd>
                </div>
              ))}
            </dl>
          </article>
        ))}
      </div>
    </>
  );
}
export function Details({
  items,
}: {
  items: { label: string; value: ReactNode }[];
}) {
  return (
    <dl className="details-grid">
      {items.map((i) => (
        <div key={i.label}>
          <dt>{i.label}</dt>
          <dd>{i.value || "Chưa có thông tin"}</dd>
        </div>
      ))}
    </dl>
  );
}
