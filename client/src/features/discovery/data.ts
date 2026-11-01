import { useEffect, useState } from "react";
import { request, type Envelope } from "../../api";
import type { PageMeta } from "../../types";

export type PublicRoom = {
  id: number;
  name: string;
  room_type: string;
  default_capacity: number;
  description: string;
  status: "AVAILABLE" | "MAINTENANCE";
  has_cover: boolean;
  cover_alt_text: string;
  cover_version: number;
  cover_is_demo: boolean;
};
export type PublicTrainer = {
  id: number;
  full_name: string;
  specialty: string;
  experience_text: string;
  rating_average: number | null;
  review_count: number;
};
export type PublicSession = {
  id: number;
  room_id: number;
  session_date: string;
  starts_at: string;
  ends_at: string;
  capacity: number;
  remaining: number;
  status: string;
};
export const roomTypes: Record<string, string> = {
  GYM: "Phòng tập",
  STRENGTH: "Sức mạnh",
  CARDIO: "Cardio",
  PT: "Tập cùng HLV",
  FUNCTIONAL: "Chức năng",
  RECOVERY: "Phục hồi",
};
export function roomDescription(room: PublicRoom) {
  if (room.description.startsWith("Dữ liệu DEMO cục bộ do "))
    return "Không gian minh họa cho buổi tập của bạn. Xem lịch đã công bố để chọn giờ phù hợp.";
  return room.description || "Xem thông tin phòng và các ca tập đã công bố.";
}

export function usePublicData<T>(path: string, allPages = false) {
  const [data, setData] = useState<T | null>(null);
  const [meta, setMeta] = useState<PageMeta>();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [version, setVersion] = useState(0);
  useEffect(() => {
    const controller = new AbortController();
    setLoading(true);
    setError("");
    setData(null);
    setMeta(undefined);
    async function load() {
      const first = await request<T>(`/api/v1/public/${path}`, {
        signal: controller.signal,
      });
      if (allPages && Array.isArray(first.data) && first.meta) {
        const values = [...first.data];
        let page = first.meta.page;
        while (page * first.meta.page_size < first.meta.total) {
          const url = new URL(path, "http://local/");
          url.searchParams.set("page", String(++page));
          const next = await request<unknown[]>(
            `/api/v1/public/${url.pathname.slice(1)}${url.search}`,
            { signal: controller.signal },
          );
          if (!next.data.length) break;
          values.push(...next.data);
        }
        return { ...first, data: values as T };
      }
      return first;
    }
    load()
      .then((value: Envelope<T>) => {
        if (!controller.signal.aborted) {
          setData(value.data);
          setMeta(value.meta);
        }
      })
      .catch((err: Error) => {
        if (!controller.signal.aborted)
          setError(
            err.message || "Chưa tải được dữ liệu. Thử lại sau ít phút.",
          );
      })
      .finally(() => {
        if (!controller.signal.aborted) setLoading(false);
      });
    return () => controller.abort();
  }, [path, allPages, version]);
  return { data, meta, loading, error, retry: () => setVersion((v) => v + 1) };
}

export { isCivilDate as validCivilDate } from "../../utils/vietnamTime";
