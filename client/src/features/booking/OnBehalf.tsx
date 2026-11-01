import { Link } from "react-router-dom";
import { useList, SearchBox, Pagination } from "../../components/shared/list";
import { Notice } from "../../components/ui";
type Member = {
  id: number;
  member_code: string;
  user: { full_name: string };
  membership_expiry: string;
  status: string;
};
export function OnBehalf() {
  const list = useList<Member>("members");
  return (
    <>
      <h1>Đặt lịch hộ hội viên</h1>
      <p>
        Chọn hồ sơ hội viên, sau đó chọn phòng và ca tập. Điều kiện thẻ được
        kiểm tra lại khi xác nhận.
      </p>
      <SearchBox
        placeholder="Tìm tên hoặc mã hội viên…"
        value={list.params.get("q") || ""}
        onChange={(v) => list.filter("q", v)}
      />
      <Notice>{list.error}</Notice>
      {list.loading && <p role="status">Đang tải hội viên…</p>}
      <div className="booking-list">
        {list.data.map((m) => (
          <div className="booking-row" key={m.id}>
            <strong>{m.member_code}</strong>
            <div>
              <h2>{m.user.full_name}</h2>
              <p>Hạn thẻ: {m.membership_expiry}</p>
            </div>
            {m.status === "ACTIVE" ? (
              <Link
                className="btn btn-primary"
                to={`/management/bookings/new/rooms?member_id=${m.id}`}
              >
                Chọn hội viên
              </Link>
            ) : (
              <span>Ngừng hoạt động</span>
            )}
          </div>
        ))}
      </div>
      <Pagination meta={list.meta} change={list.filter} />
    </>
  );
}
