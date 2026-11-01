import { Link } from "react-router-dom";
import { useList, SearchBox, Pagination } from "../../components/shared/list";
import { Notice } from "../../components/ui";
import { RatingSummary } from "./Review";
type Trainer = {
  id: number;
  user: { full_name: string };
  specialty: string;
  experience_text: string;
  rating_average: number | null;
  review_count: number;
};
export function TrainerDirectory() {
  const list = useList<Trainer>("pts?approved=true");
  return (
    <>
      <span className="portal-kicker">HUẤN LUYỆN CÁ NHÂN</span>
      <h1>Tìm người đồng hành.</h1>
      <p>
        Chọn PT trong trang phòng sau khi chọn giờ tập để kiểm tra lịch rảnh
        thực tế.
      </p>
      <SearchBox
        value={list.params.get("q") || ""}
        onChange={(v) => list.filter("q", v)}
        placeholder="Tìm tên hoặc chuyên môn PT…"
      />
      <Notice>{list.error}</Notice>
      {list.loading && <p role="status">Đang tải PT…</p>}
      {!list.loading && !list.data.length && (
        <p className="portal-empty">Chưa có PT đã duyệt phù hợp.</p>
      )}
      <div className="room-grid">
        {list.data.map((p) => (
          <article key={p.id} className="room-card">
            <div className="room-card-copy">
              <span className="profile-avatar">
                {p.user.full_name.slice(0, 1)}
              </span>
              <h2>{p.user.full_name}</h2>
              <strong>{p.specialty}</strong>
              <p>{p.experience_text}</p>
              <RatingSummary average={p.rating_average} count={p.review_count} />
              <Link to="/member/rooms" className="btn btn-secondary">
                Chọn phòng và giờ tập
              </Link>
            </div>
          </article>
        ))}
      </div>
      <Pagination meta={list.meta} change={list.filter} />
    </>
  );
}
