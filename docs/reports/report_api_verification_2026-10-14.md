# Metrics và Report API — kiểm chứng 14/10/2026

Phạm vi: `GET /api/v1/reports/operations` và `GET /api/v1/reports/pts/{id}`. Không có migration; truy vấn chỉ đọc Booking, RoomSession, Attendance, Review và provenance ở `bookings`.

## Hợp đồng và công thức

- Kỳ lọc theo `room_sessions.session_date` tại `Asia/Ho_Chi_Minh`, tính cả `from` và `to`; mặc định ngày đầu tháng đến hôm nay; tối đa 366 ngày. Mỗi response trả `as_of` từ PostgreSQL.
- Tổng Booking và phân bố trạng thái là trạng thái hiện tại tại `as_of` trong kỳ, có thể gồm ca tương lai nằm trong kỳ. Attendance, no-show, utilization và chuyên cần chỉ xét buổi đã kết thúc.
- `no_show_rate = NO_SHOW / (COMPLETED + NO_SHOW)`; mẫu số 0 trả `null`. Attendance đếm booking/participant duy nhất.
- Utilization = seat-minutes đặt bởi Booking `CONFIRMED`, `COMPLETED`, `NO_SHOW` trên phiên `OPEN` đã kết thúc / capacity-minutes của cùng các phiên đó. PT không tạo thêm chỗ; đây là thời gian đã đặt, không phải thời gian tập được đo.
- PT attendance mẫu số gồm `COMPLETED`/`NO_SHOW` có `confirmed_at`, và Booking `CANCELLED` đã xác nhận bởi MEMBER/PT theo BR-22. Tử số là các Booking đó có PT attendance. FACILITY cancel, `REJECTED`, `EXPIRED`, `PENDING_PT` bị loại. Booking `CONFIRMED` đã hết giờ nhưng job chưa hoàn tất được báo riêng.
- Booking kết thúc thiếu `confirmed_at`, hoặc đã hủy mà thiếu bằng chứng/source là `UNKNOWN`, được báo riêng. Không suy ra lịch sử từ `created_at`, `updated_at`, role hiện tại hoặc lý do cũ.
- Rating trong kỳ dùng chung predicate review hợp lệ với danh sách review và PT profile: `COMPLETED`, đã kết thúc, có PT và attendance của cả hai người. Không có review/mẫu số thì trả `null`, không biến thành 0.

## Kết quả kiểm tra

| Kiểm tra | Kết quả |
|---|---|
| `go test ./internal/repository -run '^TestReportAPIUsesVerifiedPeriodMetricsAndScopes$' -count=1` | PASS trên PostgreSQL thật. Kiểm tra totals/status, hai kiểu attendance, 40% no-show, utilization 360/7800 = 4.62%, rating 5/1, PT attendance 3/5 = 60%, các nguồn hủy, UNKNOWN, chờ completion, ngày tương lai, mẫu số 0, room filter và quyền 403/404/422. |
| `go test ./... -count=1` | PASS với `TEST_DATABASE_URL` trỏ đúng `ttcsn_gym_test`; migration và các PostgreSQL integration tests chạy hoàn tất. Chỉ DB test được reset. |
| Dataset nhỏ khi đo `EXPLAIN (ANALYZE, BUFFERS)` | 15 bookings, 15 room sessions, 5 attendance, 1 review. Query tổng hợp đại diện chạy 0.290 ms; planner chọn sequential scan với dữ liệu rất nhỏ. Không kết luận hiệu năng production/NFR từ dữ liệu này. |
| Scope/filter | MEMBER và RECEPTIONIST nhận 403; PT chỉ được xem profile ID của mình; ID ngoài scope hoặc room/PT không tồn tại trả 404; ngày/range/room_id sai trả 422. |

Không chạy API process, Vite process hoặc browser QA; HTTP integration gọi Gin router trong test và truy cập PostgreSQL test thật. API 8080/client 5173 vẫn tắt. Dashboard, `/pt/reviews`, runbook demo tổng và browser visual QA thuộc chặng F.
