# Đặc tả yêu cầu phần mềm — SRS v0.3

**Hệ thống:** Quản lý phòng tập GYM · **Nhóm:** Nhóm 1 · **Ngày baseline:** 10/10/2026  
**Trạng thái:** Đã hoàn thiện phân tích; chưa xác nhận triển khai/kiểm thử ứng dụng.

## 1. Mục đích và phạm vi

SRS v0.3 là nguồn yêu cầu hiện hành, thay thế v0.2 cho công việc tiếp theo. Phạm vi, thuật ngữ và quyết định tuân theo `00_quyet_dinh_va_quy_tac_nghiep_vu.md`; quyền theo `06_ma_tran_phan_quyen.md`; API theo `12_dac_ta_API.md`.

MVP là web responsive cho một cơ sở, dùng React/Tailwind, Go/Gin và PostgreSQL 16+. MVP không gồm thanh toán/doanh thu, nhiều chi nhánh, app native, cổng phần cứng, SMS/email hoặc lớp học nhóm riêng.

## 2. Tác nhân và thuật ngữ

Năm tác nhân: MEMBER, PT, RECEPTIONIST, MANAGER, ADMIN. `booking` là một chỗ của một hội viên trong một `room_session`; `room_session` là phiên mở phòng theo ngày và ca; attendance là bản ghi có mặt của từng người; trạng thái tài khoản, hội viên, PT và booking là các miền độc lập.

## 3. Yêu cầu chức năng

### REQ-01 — Xác thực phiên (Must)

- UC: UC-07. Tất cả vai trò đăng nhập bằng username/password.
- Hệ thống dùng thông báo lỗi chung cho sai thông tin, từ chối tài khoản LOCKED, và cấp token có `sub`, `role`, `auth_version`, `jti`, `exp` khi hợp lệ.
- Logout phía client xóa token; khóa/reset/đổi role phải làm token cũ mất hiệu lực theo BR-02.
- Nghiệm thu: mật khẩu băm; token hết hạn/version cũ bị từ chối; không tiết lộ tài khoản tồn tại.

### REQ-02 — Phân quyền theo vai trò và bản ghi (Must)

- UC: mọi UC. Mặc định deny; kiểm tra role + SELF/ASSIGNED/ALL + trạng thái tại server.
- Frontend không được xem là ranh giới bảo mật; ID do client gửi không chứng minh quyền.
- Nghiệm thu: bộ test cho từng permission, truy cập ngang, tài khoản khóa và trạng thái không hợp lệ.

### REQ-03 — Quản lý hội viên (Must)

- UC: UC-09. Lễ tân/Quản lý/Admin tạo, xem, cập nhật, gia hạn và vô hiệu hóa.
- Dữ liệu tối thiểu: mã hội viên duy nhất, họ tên, ngày sinh, giới tính tùy chọn, SĐT, email tùy chọn, địa chỉ, ngày tham gia, ngày hết hạn, trạng thái.
- Không hard-delete; thay đổi trạng thái/hạn thẻ có audit khi ảnh hưởng quyền sử dụng.
- Nghiệm thu: validate ngày, mã duy nhất; lịch sử booking còn nguyên sau vô hiệu hóa.

### REQ-04 — Quản lý phòng (Must)

- UC: UC-10. Quản lý/Admin tạo, sửa mô tả/loại/sức chứa mặc định và trạng thái.
- Tên phòng duy nhất không phân biệt hoa thường; phòng đã tham chiếu chỉ vô hiệu hóa/bảo trì.
- Chuyển bảo trì phải hiển thị phiên/booking tương lai bị ảnh hưởng, không tự xóa.

### REQ-05 — Quản lý ca và phiên phòng (Must)

- UC: UC-11. Quản lý/Admin tạo ca chung không chồng giờ và mở phiên phòng theo ngày.
- Ca có tên, giờ bắt đầu/kết thúc, trạng thái; phiên phòng có timestamps, capacity và OPEN/CLOSED/MAINTENANCE.
- Không sửa giờ ca đã dùng; không giảm capacity phiên dưới số booking giữ chỗ.

### REQ-06 — Tạo booking, PT tùy chọn (Must)

- UC: UC-02. Hội viên đặt cho mình; Lễ tân/Admin đặt hộ.
- Hệ thống thực thi BR-03…BR-15 trong một transaction, dùng `Idempotency-Key`.
- Không PT → CONFIRMED; có PT → PENDING_PT với `respond_by`.
- Nghiệm thu: không vượt capacity, không xung đột member/PT, retry không tạo trùng, không trả thành công trước commit.

### REQ-07 — Hủy booking (Must)

- UC: UC-06. Hội viên hủy theo mốc 2 giờ; PT xử lý lịch được gán; nhân viên hủy ngoại lệ có lý do.
- Chỉ trạng thái cho phép mới chuyển CANCELLED; thao tác lặp không giải phóng hai lần.
- Nghiệm thu: test biên đúng 2 giờ ±1 giây, quyền sở hữu, đồng thời với check-in/job.

### REQ-08 — Chống xung đột và quá tải (Must)

- UC: include trong UC-02/03/06/11.
- UNIQUE có điều kiện bảo vệ lịch member/PT; transaction khóa `room_sessions` để kiểm tra capacity; ca chung không chồng giờ.
- Có chiến lược timeout/deadlock/serialization retry giới hạn và mã lỗi ổn định.

### REQ-09 — Điểm danh (Must)

- UC: UC-04. Lễ tân/Quản lý/Admin ghi attendance riêng cho MEMBER/PT thuộc booking CONFIRMED.
- Mỗi participant có tối đa một attendance/booking; ghi người thực hiện và thời gian.
- Job kết thúc buổi dùng attendance MEMBER để chọn COMPLETED/NO_SHOW.

### REQ-10 — Báo cáo vận hành (Should)

- UC: UC-14. Quản lý/Admin xem booking theo trạng thái, attendance, occupancy, no-show, rating, chuyên cần; PT xem chỉ số của mình.
- Mỗi chỉ số ghi công thức, khoảng ngày, timezone và bộ lọc; không có doanh thu.

### REQ-11 — Tìm kiếm/lọc hội viên (Should)

- UC: UC-09. Tìm theo mã/tên/SĐT/email; lọc trạng thái và hạn thẻ; phân trang bắt buộc.
- Mục tiêu p95 API < 2 giây chỉ được công bố đạt sau khi chạy NFR-01 với dữ liệu đại diện.

### REQ-12 — Quản lý tài khoản (Must)

- UC: UC-08. Admin tạo tài khoản một role, reset password, khóa/mở và đổi role có kiểm tra.
- Username/email chuẩn hóa duy nhất; SĐT không bắt buộc duy nhất; thao tác nhạy cảm có audit và tăng `auth_version` khi cần.

### REQ-13 — Xem lịch (Should)

- UC: UC-13. MEMBER xem của mình; PT xem được gán; RECEPTIONIST/MANAGER/ADMIN xem toàn cơ sở.
- Lọc khoảng ngày, phòng, PT, trạng thái; giới hạn khoảng và phân trang để bảo vệ hiệu năng.

### REQ-14 — Phản hồi và danh sách yêu cầu (Should)

- UC: UC-16. Request hiện tại có toast/alert; PT lấy PENDING_PT và hội viên lấy trạng thái lịch từ server bằng polling.
- Trạng thái trong CSDL là nguồn thật; không cam kết email/SMS/push trong MVP.

### REQ-15 — Xuất PDF/Excel (Could)

- UC: UC-15. Dùng cùng nguồn/công thức/bộ lọc với báo cáo; ghi thời gian, timezone, người xuất.
- Không được làm chậm API tương tác; giới hạn khoảng/dung lượng và quyền dữ liệu.

### REQ-16 — Nộp hồ sơ PT (Must)

- UC: UC-12. PT tạo/cập nhật hồ sơ và nộp từ DRAFT/REJECTED sang PENDING.
- Chứng chỉ lưu riêng tư, validate loại/kích thước; lý do quyết định cũ được giữ qua audit.

### REQ-17 — Duyệt/cấp phép PT (Must)

- UC: UC-01. Quản lý/Admin duyệt, từ chối hoặc đình chỉ, ghi người/thời gian/lý do.
- Duyệt không tự mở tài khoản LOCKED; đình chỉ không xóa lịch tương lai mà trả danh sách cần xử lý.

### REQ-18 — PT phản hồi lịch 1-1 (Must)

- UC: UC-03. PT chỉ phản hồi booking PENDING_PT được gán, trước respond_by.
- Chấp nhận → CONFIRMED; từ chối → REJECTED; job quá hạn → EXPIRED.

### REQ-19 — Đánh giá PT (Should)

- UC: UC-05. Hội viên đánh giá 1–5 sao cho booking đủ BR-21; một review/booking.
- Backend suy ra member/PT từ booking; không tin `member_id`/`pt_id` trong body.

### REQ-20 — Rating và chuyên cần PT (Should)

- UC: UC-05/14. Rating là trung bình review hợp lệ; attendance rate theo BR-22 và hiển thị N/A khi mẫu số 0.
- Hai chỉ số độc lập; không tạo Trust Score tổng hợp khi chưa có quyết định nghiệp vụ mới.

## 4. Yêu cầu phi chức năng

### NFR-01 — Hiệu năng

- Mục tiêu: p95 API tương tác < 2 giây với dữ liệu đại diện; không đặt mục tiêu “API trung bình < 50 ms” khi chưa đo.
- Báo cáo phải nêu cấu hình máy, số dòng, mix request, thời lượng, p50/p95/p99, lỗi và lock wait.

### NFR-02 — Tải và tính đúng đồng thời

- Mục tiêu: 100 người dùng đồng thời trong kịch bản đã định nghĩa, không có 5xx ngoài lỗi tiêm chủ động và không vi phạm capacity/unique.
- Bài race chỗ cuối là test tính đúng riêng; throughput không thay thế tính đúng.

### NFR-03 — Bảo mật

- Bcrypt cost ≥10; TLS khi triển khai; truy vấn tham số hóa; validate đầu vào; secret qua môi trường/secret store.
- Authorization phía server; token cũ bị thu hồi khi khóa/reset/đổi role; upload chứng chỉ private, tên ngẫu nhiên, allowlist type/size.
- Log/audit không chứa secret; CORS/XSS/CSRF được cấu hình theo chính sách token thực tế.

### NFR-04 — Khả dụng và dễ dùng

- Responsive từ 375px; thao tác đặt lịch mục tiêu ≤3 bước sau khi chọn lịch; lỗi tiếng Việt kèm mã ổn định.
- Keyboard/focus/label cơ bản cho form; timezone hiển thị rõ trong lịch/báo cáo.

### NFR-05 — Tin cậy, sao lưu và khôi phục

- Mục tiêu RPO <24 giờ, RTO <2 giờ chỉ được ghi “đạt” sau restore drill.
- Backup hằng ngày, theo dõi thất bại, kiểm tra khả năng đọc và định kỳ phục hồi sang môi trường cô lập.
- Transaction xử lý lỗi Begin/query/Commit; job nền idempotent và có retry giới hạn.

## 5. Giao diện ngoài

- REST JSON dưới `/api/v1`, UTF-8, HTTPS khi triển khai; timestamp ISO 8601 có offset.
- Envelope thành công `{data, meta?, request_id}`; lỗi `{error:{code,message,details?},request_id}`.
- Upload avatar/chứng chỉ qua multipart; chứng chỉ chỉ tải qua endpoint có quyền.
- PostgreSQL 16+; timezone vận hành `Asia/Ho_Chi_Minh`.

## 6. Ma trận truy vết

| REQ | UC | BR chính | Dữ liệu chính | Test chính |
|---|---|---|---|---|
| 01 | 07 | 01–02 | users | AUTH, SEC |
| 02 | mọi UC | 01–05, 25 | users, audit_events | AUTHZ |
| 03,11 | 09 | 03,24–25 | members | MEMBER |
| 04 | 10 | 09–10,24 | rooms | FACILITY |
| 05 | 11 | 06–10 | time_slots, room_sessions | SLOT/SESSION |
| 06,08 | 02 | 03–16 | bookings, idempotency_keys | BOOK/RACE |
| 07 | 06 | 17–18 | bookings, audit_events | CANCEL |
| 09 | 04 | 19–20 | booking_attendance | ATTEND |
| 10 | 14 | 20–23 | bookings, attendance, reviews | REPORT |
| 12 | 08 | 01–02,25 | users, roles | ACCOUNT |
| 13 | 13 | quyền | bookings, sessions | SCHEDULE |
| 14 | 16 | 12,16 | bookings | POLL |
| 15 | 15 | quyền,23 | báo cáo | EXPORT |
| 16 | 12 | 04–05,26 | personal_trainers, files | PT-SUBMIT |
| 17 | 01 | 04–05,25 | personal_trainers, audit | PT-APPROVE |
| 18 | 03 | 12,16,18 | bookings | PT-RESPOND |
| 19,20 | 05,14 | 19–22 | attendance, pt_reviews | REVIEW/METRIC |

## 7. Phê duyệt và thay đổi

Tài liệu này chưa tự tuyên bố chữ ký. Khi nhóm rà soát, ghi tên, ngày, phiên bản và vấn đề còn mở trong biên bản thực tế. Thay đổi BR hoặc phạm vi cần cập nhật decision log, SRS, use case, schema, API, backlog và test liên quan.
