# Kế hoạch kiểm thử baseline v0.3

Trạng thái: **Kế hoạch kiểm thử toàn MVP**. Kết quả đã chạy từng module nằm trong `reports/`; các cổng còn thiếu không tự coi là đạt. PASS chỉ được ghi khi có môi trường, phiên bản code/schema, log và kết quả tương ứng.

## 1. Mức kiểm thử

| Mức | Mục tiêu | Công cụ gợi ý |
|---|---|---|
| Static/doc | REQ–UC–BR–permission–table–API khớp | script kiểm tra mã/liên kết, review chéo |
| Unit | Quy tắc/trạng thái/công thức | `go test ./...` |
| DB integration | Constraint, transaction, migration | PostgreSQL 16 container/local test DB |
| API integration | Contract, authn/authz, error code | Go HTTP tests/Postman/Newman |
| Concurrency | Capacity và unique dưới race | nhiều connection/goroutine, barrier |
| E2E | Luồng 5 vai trò | browser automation |
| NFR | tải, security, backup/restore | load tool, security checks, restore drill |

## 2. Điều kiện môi trường và bằng chứng

Mỗi lần chạy ghi: commit/build ID, migration version, PostgreSQL version, cấu hình máy, timezone, seed size, lệnh, thời gian, PASS/FAIL và đường dẫn log/report. Dữ liệu test tách môi trường; không dùng dữ liệu cá nhân thật.

## 3. Kịch bản bắt buộc

| ID | Kịch bản | Kết quả mong đợi | Mức |
|---|---|---|---|
| AUTHZ-01 | Member A gửi ID B; PT A phản hồi lịch PT B; nhân viên sửa role. | 403/404; không lộ/sửa dữ liệu. | API |
| PT-01 | PT PENDING nộp, thử nhận lịch, bị từ chối/nộp lại, bị đình chỉ. | Transition/quyền đúng; lịch tương lai được báo để xử lý. | Unit/API |
| RACE-01 | Hai member cùng đặt chỗ cuối. | Một thành công, một ROOM_FULL; count không vượt capacity. | DB concurrency |
| RACE-02 | Cùng member hoặc PT đặt hai phòng đồng thời cùng ca. | Unique/conflict; tối đa một booking giữ lịch. | DB concurrency |
| SLOT-01 | Tạo 06–08 và 07–09; tạo 08–10. | Ca chồng bị chặn; ca liền nhau hợp lệ. | DB |
| STATE-01 | PT accept đúng lúc job expire; hủy đồng thời check-in; giảm capacity đồng thời đặt. | Một transition hợp lệ; không hồi sinh/trừ chỗ hai lần. | DB concurrency |
| CANCEL-01 | Hủy đúng 2 giờ, trước/sau 1 giây; override; hủy lặp. | Biên đúng BR-17; audit; idempotent. | Unit/API |
| ATTEND-01 | Chỉ member, chỉ PT, cả hai hoặc không ai có mặt. | Attendance riêng; COMPLETED/NO_SHOW và metric đúng. | Integration |
| REVIEW-01 | Review trước kết thúc, thiếu attendance, đúng điều kiện, lần hai. | 422, 422, success, 409. | API/DB |
| HISTORY-01 | Vô hiệu hóa room/user/PT/slot đã dùng. | Lịch sử còn; booking mới bị chặn phù hợp. | DB/API |
| IDEMP-01 | Timeout rồi retry cùng key; payload khác cùng key; commit lỗi. | Cùng kết quả; conflict; không trả success/không bản ghi dở. | Integration |
| MIG-01 | Dữ liệu v0.2 có status/review mơ hồ. | Báo ngoại lệ; không bịa attendance/booking link. | Migration |
| FILE-01 | Tải chứng chỉ PT khác; upload sai type/quá lớn/path name. | Bị chặn; không path traversal/public access. | Security/API |
| REPORT-01 | Dataset gồm open/closed, pending/cancel/no-show/unverified. | Công thức/mẫu số đúng; không có doanh thu. | Integration |
| TOKEN-01 | Khóa/reset/đổi role sau khi cấp token. | Token version cũ 401; dữ liệu không đổi ngoài thao tác. | Security/API |

## 4. Kiểm thử phi chức năng

### NFR-01/NFR-02

- Seed tối thiểu được công bố: số users/members/PT/rooms/sessions/bookings/reviews.
- Workload nêu tỷ lệ login/read/search/create/cancel/report; ramp-up, steady duration và concurrent users.
- Thu p50/p95/p99, throughput, 4xx theo nghiệp vụ, 5xx, CPU/RAM, DB connections/lock wait/retry.
- Tiêu chí: p95 API tương tác <2 giây và không vi phạm constraint trong workload 100 concurrent users. Báo riêng kết quả race chỗ cuối.

### NFR-03

- Bcrypt hash/cost; generic login error; token expiry/revoke; object authz; input validation; upload private; CORS; secret/log scan.
- Không tuyên bố chống XSS/CSRF chỉ dựa vào React/CORS; test theo cơ chế lưu token thực tế.

### NFR-04

- Chrome/Edge/Firefox hiện hành; viewport 375, 768, ≥1200px.
- Keyboard focus, label/error, contrast cơ bản; lịch hiển thị timezone; flow đặt ≤3 bước sau chọn lịch.

### NFR-05

- Tạo backup, ghi hash/size/time; restore sang DB rỗng cô lập; chạy integrity queries và smoke API.
- Đo RPO/RTO thực; kiểm tra cảnh báo khi backup thất bại. Mục tiêu chỉ PASS nếu RPO <24h và RTO <2h trong drill.

## 5. Cổng nghiệm thu

- G1: 0 mâu thuẫn REQ/UC/permission/status/schema/API mức P0/P1.
- G2: Unit + DB integration + API Must đều pass; không lỗi dữ liệu mở.
- G3: RACE/STATE/IDEMP pass trên PostgreSQL thật.
- G4: E2E 5 vai trò pass; 0 lỗi Critical/High chưa có xử lý.
- G5: NFR report và restore drill có bằng chứng; giới hạn chưa đạt được ghi rõ.

## 6. Amendment kiểm thử lịch phòng — 11/10/2026

Các cổng phía trên vẫn là mục tiêu cho toàn MVP. Lịch phòng có unit interval/timezone/priority, PostgreSQL/API revision/preview/holds/rollback/idempotency/restart/job, cùng frontend grid/form/xác nhận/giữ nháp. Case sửa ngoại lệ phải giữ WITHDRAWN cũ và ACTIVE mới, không cho đổi ngày hoặc rút lại bản cũ. Báo cáo kết quả từng phạm vi tại `reports/room_calendar_verification_2026-10-11.md`.

Còn phải kiểm tra trực tiếp UI ngoại lệ/chỉ đọc, 375/768/1024/1440px, zoom 200%, focus/touch; browser bị auto-review chặn do hạn mức trong lần này. Job startup/lặp đã chạy, chưa chờ đồng hồ 00:05; chưa có race Booking API vì module đó chưa triển khai. Không đánh dấu G4/G5 toàn MVP đạt chỉ từ test lịch phòng.
