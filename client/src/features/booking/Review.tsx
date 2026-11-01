import { useRef, useState, type FormEvent } from "react";
import { resource } from "../../api";
import { useAuth } from "../../app/session";
import { Button, Field, Notice } from "../../components/ui";
import { useRecord } from "../../components/shared/record";
import type { Booking } from "./Booking";
import { vietnamLongDate } from "../../utils/vietnamTime";

type Review = {
  rating: number;
  comment: string;
  created_at: string;
  reviewer_label: "Hội viên";
};

type AttendanceSummary = {
  participants: {
    participant_type: "MEMBER" | "PT";
    attendance: unknown | null;
  }[];
};

export function RatingSummary({
  average,
  count,
}: {
  average: number | null | undefined;
  count: number;
}) {
  if (average == null || count < 1)
    return (
      <span className="rating-summary rating-empty">Chưa có đánh giá</span>
    );
  return (
    <span
      className="rating-summary"
      aria-label={`${average.toFixed(1)} trên 5 sao, ${count} đánh giá`}
    >
      <span aria-hidden="true">★</span> {average.toFixed(1)}
      <span className="rating-count">{count} đánh giá</span>
    </span>
  );
}

function StarDisplay({ rating }: { rating: number }) {
  return (
    <span className="review-star-display" aria-label={`${rating} trên 5 sao`}>
      <span aria-hidden="true">
        {"★".repeat(rating)}
        {"☆".repeat(5 - rating)}
      </span>
      <strong>{rating}</strong>
      <span className="review-out-of">/ 5</span>
    </span>
  );
}

function ReviewCard({ review }: { review: Review }) {
  const date = vietnamLongDate(review.created_at);
  return (
    <div className="review-readback">
      <div className="review-readback-heading">
        <StarDisplay rating={review.rating} />
        <span className="review-author">
          {review.reviewer_label} · {date}
        </span>
      </div>
      {review.comment && <blockquote>{review.comment}</blockquote>}
    </div>
  );
}

export function ReviewPanel({ booking }: { booking: Booking }) {
  const { session } = useAuth();
  const member = session.user.role === "MEMBER";
  const visible = booking.status === "COMPLETED" && booking.pt_id > 0;
  const record = useRecord<Review | null>(
    visible ? `bookings/${booking.id}/review` : null,
  );
  const attendance = useRecord<AttendanceSummary>(
    visible && member ? `bookings/${booking.id}/attendance` : null,
  );
  const [rating, setRating] = useState(0);
  const [comment, setComment] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const submitting = useRef(false);

  if (!visible) return null;

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (rating < 1 || submitting.current) return;
    submitting.current = true;
    setBusy(true);
    setError("");
    try {
      const result = await resource.save<Review>(
        session.accessToken,
        `bookings/${booking.id}/review`,
        { rating, comment },
      );
      record.setData(result.data);
    } catch (cause) {
      setError((cause as Error).message);
    } finally {
      submitting.current = false;
      setBusy(false);
    }
  }

  const endPassed =
    Date.parse(booking.ends_at) <= Date.parse(booking.server_time);
  const present = new Set(
    attendance.data?.participants
      .filter((participant) => Boolean(participant.attendance))
      .map((participant) => participant.participant_type),
  );
  const eligible = endPassed && present.has("MEMBER") && present.has("PT");

  return (
    <section
      className="review-panel"
      aria-labelledby={`review-title-${booking.id}`}
    >
      <div className="review-panel-heading">
        <div>
          <h3 id={`review-title-${booking.id}`}>Đánh giá huấn luyện viên</h3>
          <p>Buổi tập đã hoàn thành</p>
        </div>
        {record.data && <span className="review-saved">Đã đánh giá</span>}
      </div>

      {record.loading && (
        <p role="status" className="review-loading">
          Đang tải đánh giá…
        </p>
      )}
      {record.error && (
        <>
          <Notice>{record.error}</Notice>
          <Button variant="secondary" onClick={record.refresh}>
            Tải lại
          </Button>
        </>
      )}
      {!record.loading && !record.error && record.data && (
        <ReviewCard review={record.data} />
      )}

      {!record.loading && !record.error && !record.data && member && (
        <>
          {attendance.loading && (
            <p role="status" className="review-loading">
              Đang kiểm tra điều kiện đánh giá…
            </p>
          )}
          {attendance.error && (
            <>
              <Notice>{attendance.error}</Notice>
              <Button variant="secondary" onClick={attendance.refresh}>
                Kiểm tra lại
              </Button>
            </>
          )}
          {!attendance.loading && !attendance.error && !eligible && (
            <p className="review-guidance">
              Để gửi đánh giá, Hội viên và PT cần có mặt trong buổi tập đã kết
              thúc.
            </p>
          )}
          {!attendance.loading && !attendance.error && eligible && (
            <form
              className="review-form"
              onSubmit={(event) => void submit(event)}
            >
              <fieldset className="review-rating-fieldset" disabled={busy}>
                <legend>Buổi tập này thế nào?</legend>
                <div className="review-stars">
                  {[1, 2, 3, 4, 5].map((value) => (
                    <label key={value}>
                      <input
                        type="radio"
                        name={`review-rating-${booking.id}`}
                        value={value}
                        aria-label={`${value} sao`}
                        checked={rating === value}
                        required={value === 1}
                        onChange={() => setRating(value)}
                      />
                      <span aria-hidden="true">
                        {value <= rating ? "★" : "☆"}
                      </span>
                      <span className="sr-only">{value} sao</span>
                    </label>
                  ))}
                  <output aria-live="polite">
                    {rating ? `${rating} / 5` : "Chọn sao"}
                  </output>
                </div>
              </fieldset>
              <Field label="Nhận xét" hint="Không bắt buộc · tối đa 1000 ký tự">
                <textarea
                  className="input review-comment"
                  value={comment}
                  maxLength={1000}
                  rows={3}
                  placeholder="Điều gì trong buổi tập hữu ích với bạn?"
                  onChange={(event) => setComment(event.target.value)}
                />
              </Field>
              <Notice>{error}</Notice>
              <div className="review-form-footer">
                <span>{comment.length}/1000</span>
                <Button busy={busy} disabled={rating < 1}>
                  Gửi đánh giá
                </Button>
              </div>
            </form>
          )}
        </>
      )}

      {!record.loading && !record.error && !record.data && !member && (
        <p className="review-guidance">
          Hội viên chưa gửi đánh giá cho buổi tập này.
        </p>
      )}
    </section>
  );
}
