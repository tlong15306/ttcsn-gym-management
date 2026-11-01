import { useEffect, useState } from "react";
import { resource } from "../../api";
import { addDays, sessionDay, weekStart, type WeekSession } from "./weekly";

export async function fetchDay<T>(
  token: string,
  roomID: string,
  date: string,
  signal?: AbortSignal,
): Promise<T[]> {
  const values: T[] = [];
  let page = 1;
  for (;;) {
    const response = await resource.list<T>(
      token,
      "booking-options/sessions",
      new URLSearchParams({
        room_id: roomID,
        date,
        page: String(page),
        page_size: "100",
      }),
      signal,
    );
    values.push(...response.data);
    if (
      !response.data.length ||
      !response.meta ||
      page * response.meta.page_size >= response.meta.total
    )
      return values;
    page++;
  }
}
export function useWeeklyOptions<T extends WeekSession>(
  token: string,
  roomID: string,
  date: string,
  revision: number,
  enabled: boolean,
) {
  const from = weekStart(date),
    [state, setState] = useState<{
      data: T[];
      loading: boolean;
      error: string;
      loadedKey: string;
      scope: string;
      ready: boolean;
    }>({
      data: [],
      loading: enabled,
      error: "",
      loadedKey: "",
      scope: "",
      ready: false,
    });
  const scope = `${token}:${roomID}:${from}`;
  const key = `${token}:${roomID}:${from}:${revision}`;
  useEffect(() => {
    if (!enabled) return;
    const controller = new AbortController();
    setState((previous) => ({
      data: previous.scope === scope ? previous.data : [],
      loading: true,
      error: "",
      loadedKey: "",
      scope,
      ready: previous.scope === scope && previous.ready,
    }));
    Promise.all(
      Array.from({ length: 7 }, (_, i) =>
        fetchDay<T>(token, roomID, addDays(from, i), controller.signal),
      ),
    )
      .then((days) => {
        if (controller.signal.aborted) return;
        const map = new Map(
          days
            .flat()
            .filter(
              (s) => sessionDay(s) >= from && sessionDay(s) <= addDays(from, 6),
            )
            .map((s) => [s.id, s]),
        );
        setState({
          data: [...map.values()],
          loading: false,
          error: "",
          loadedKey: key,
          scope,
          ready: true,
        });
      })
      .catch((error: Error) => {
        if (!controller.signal.aborted)
          setState((previous) => ({
            data: previous.scope === scope ? previous.data : [],
            loading: false,
            error: error.message,
            loadedKey: key,
            scope,
            ready: previous.scope === scope && previous.ready,
          }));
      });
    return () => controller.abort();
  }, [key, enabled]);
  return {
    ...state,
    data: state.scope === scope ? state.data : [],
    hasSnapshot: state.scope === scope && state.ready,
    loading: enabled && (state.loading || state.loadedKey !== key),
    meta: {
      page: 1,
      page_size: state.data.length || 100,
      total: state.data.length,
    },
  };
}
