# Đặc tả REST API — v1.0 cho baseline SRS v0.3

**Phân biệt thiết kế và triển khai:** Identity, Members, PT, Facilities/Sessions, lịch phòng, ảnh phòng, Booking, Attendance, Review, hai Report API và Public Discovery U1 đã có hợp đồng local theo `openapi_admin_crud.json` (83 operations). Response Booking có `confirmed_at`, `cancellation_source`, `cancelled_actor_role`; dữ liệu cũ thiếu chứng cứ không được đoán. OpenAPI là hợp đồng của mã nguồn hiện hành; payload/lỗi ở baseline dưới đây chưa có trong OpenAPI không được coi là đã bàn giao. Kiểm tra router bằng `scripts/generate_api_contracts.py --check`; collection là `GYM_Admin_CRUD.postman_collection.json`. Hướng dẫn hiện hành: `17_huong_dan_chay_local.md`, `21_huong_dan_lich_phong.md`, `23_huong_dan_booking_portal.md`.

### API công khai U1

U2–U4 (01/11/2026) không thêm endpoint: thử chọn ở public chỉ lưu ý định phía trình duyệt; sau login lấy lại `booking-options/sessions`/`booking-options/pts` có Bearer rồi dùng API booking đơn/batch hiện hữu. Server tiếp tục kiểm tra atomically và xử lý idempotency, không tin điều kiện/ID trong bản nháp. Số operations vẫn là 83.

`GET /public/rooms` (q, room_type, page, page_size), `/public/rooms/:id`, `/public/rooms/:id/cover-image` (size=card/detail), `/public/rooms/:id/sessions` (from/to bắt buộc, tối đa 7 ngày bao gồm hai đầu, có phân trang), `/public/trainers` (q, page, page_size). Không yêu cầu Bearer; chỉ trả trường công khai. Ảnh trả JPEG/ETag, kiểm tra lại visibility trước 304. Session chỉ trả số chỗ tổng hợp và trạng thái, không quyết định điều kiện của từng Hội viên và không giữ chỗ. Mutation và endpoint hiện hữu vẫn cần xác thực như trước.

## 1. Quy ước chung

- Base path: `/api/v1`; JSON UTF-8; timestamp ISO 8601 có offset.
- Header xác thực: `Authorization: Bearer <token>`.
- Tạo booking bắt buộc `Idempotency-Key` (UUID hoặc chuỗi ngẫu nhiên đủ mạnh, tối đa 100 ký tự).
- Phân trang: `page` bắt đầu từ 1, `page_size` mặc định 20, tối đa 100.
- Thành công: `{"data": ..., "meta": ..., "request_id": "..."}`.
- Lỗi: `{"error":{"code":"...","message":"...","details":...},"request_id":"..."}`.
- Client không gửi `member_id` cho thao tác SELF; server suy ra từ token. Endpoint đặt hộ mới nhận `member_id`.

## 2. Mã lỗi ổn định

| HTTP | Code | Ý nghĩa |
|---:|---|---|
| 400 | INVALID_REQUEST | JSON/query/header sai định dạng. |
| 401 | UNAUTHENTICATED | Thiếu/hết hạn/sai token hoặc auth_version cũ. |
| 403 | FORBIDDEN | Có danh tính nhưng không có quyền. |
| 404 | NOT_FOUND | Không tồn tại hoặc bị che vì ngoài phạm vi. |
| 409 | INVALID_STATE | Trạng thái nguồn không cho phép thao tác. |
| 409 | MEMBER_CONFLICT / PT_CONFLICT | Trùng lịch giữ chỗ. |
| 409 | ROOM_FULL | Phiên đã đủ chỗ. |
| 409 | IDEMPOTENCY_CONFLICT | Cùng key nhưng payload khác. |
| 409 | REVIEW_EXISTS | Booking đã có đánh giá. |
| 422 | MEMBERSHIP_INACTIVE | Tài khoản/hồ sơ/hạn thẻ không hợp lệ. |
| 422 | CANCELLATION_WINDOW_CLOSED | Hội viên tự hủy sau mốc cho phép. |
| 422 | REVIEW_NOT_ELIGIBLE | Chưa hoàn tất/thiếu attendance/không có PT. |
| 422 | VALIDATION_FAILED | Dữ liệu hợp lệ cú pháp nhưng sai quy tắc. |
| 429 | RATE_LIMITED | Vượt giới hạn request. |
| 500 | INTERNAL_ERROR | Lỗi không dự kiến; không lộ SQL/stack trace. |
| 503 | TEMPORARILY_UNAVAILABLE | Phụ thuộc tạm lỗi hoặc retry transaction đã hết. |

## 3. Endpoint theo module

### Identity & Access

| Method & path | Permission | Request chính | Response/ghi chú | REQ/UC |
|---|---|---|---|---|
| `POST /auth/login` | Public | username, password | token + user; lỗi chung | 01/UC-07 |
| `POST /auth/logout` | Authenticated | none | 204; tăng `auth_version`, thu hồi toàn bộ token hiện có của tài khoản | 01/UC-07 |
| `GET /me` | Authenticated | none | hồ sơ tài khoản + role/status phù hợp | 01/UC-07 |
| `PATCH /me` | `profile.self.update` | trường hồ sơ cho phép | không nhận role/status | 02/UC-07 |
| `GET /users` | `account.manage` | filter, page | danh sách phân trang | 12/UC-08 |
| `POST /users` | `account.manage` | username, password, email, role | 201 user | 12/UC-08 |
| `GET /users/{id}` | `account.manage` | — | chi tiết public, không có hash | 12/UC-08 |
| `PATCH /users/{id}` | `account.manage` | full_name,email,phone | update contact + audit | 12/UC-08 |
| `PATCH /users/{id}/status` | `account.manage` | ACTIVE/LOCKED, reason | tăng auth_version khi khóa | 12/UC-08 |
| `PATCH /users/{id}/role` | `account.manage` | role, reason | kiểm tra hồ sơ phụ thuộc | 12/UC-08 |
| `POST /users/{id}/reset-password` | `account.manage` | mật khẩu mới hoặc flow reset | tăng auth_version | 12/UC-08 |

### Members

| Method & path | Permission | Request chính | Ghi chú | REQ/UC |
|---|---|---|---|---|
| `GET /members` | SELF/ASSIGNED/ALL | q,status,page,page_size | tìm mã/tên/SĐT/email theo scope; chưa có filter expiry | 03,11/UC-09 |
| `GET /members/eligible-accounts` | `member.create` | q,page,page_size | MEMBER chưa có profile; không đòi account.manage | 03/UC-09 |
| `POST /members` | `member.create` | account hoặc user_id + profile | tạo atomically; account mới buộc MEMBER | 03/UC-09 |
| `GET /members/{id}` | SELF/ASSIGNED/ALL | — | field-level filtering cho PT | 03/UC-09 |
| `PATCH /members/{id}` | `member.update` | trường hồ sơ | không hard-delete | 03/UC-09 |
| `PATCH /members/{id}/membership` | `member.update` | expiry,status,reason | audit | 03/UC-09 |

### PT

| Method & path | Permission | Request chính | Ghi chú | REQ/UC |
|---|---|---|---|---|
| `GET /pts` | Authenticated | approved,available_at,q,page | member chỉ thấy APPROVED | 16,17/UC-12 |
| `GET /pts/me` | PT SELF | — | hồ sơ + quyết định gần nhất | 16/UC-12 |
| `GET /pts/{id}` | scope PT | — | người khác chỉ xem PT active/approved, không có lý do quyết định nội bộ; trả `rating_average` và `review_count` tính từ review hợp lệ | 16,17/UC-12,20/UC-05 |
| `PUT /pts/me/profile` | `pt.profile.update` | specialty,experience | DRAFT/REJECTED có thể sửa | 16/UC-12 |
| `POST /pts/me/submit` | `pt.profile.submit` | — | → PENDING | 16/UC-12 |
| `POST /pts/me/certificates` | PT SELF | multipart | private object, allowlist | 16/UC-12 |
| `GET /pts/{id}/certificate` | `certificate.read` | — | kiểm tra quyền mỗi lần tải | 16,17/UC-01 |
| `POST /pts/{id}/decision` | `pt.approve/reject/suspend` | action,reason | state transition + audit | 17/UC-01 |
| `GET/PUT /pts/me/availability` | `pt.availability.manage` | weekly schedule | validate không phá lịch | 16/UC-12 |
| `GET/POST /pts/me/unavailability` | SELF | starts_at,ends_at,reason; GET page/page_size | kiểm tra booking; GET phân trang | 16/UC-12 |
| `DELETE /pts/me/unavailability/{absenceId}` | SELF | — | hủy khoảng nghỉ, giữ audit | 16/UC-12 |

Admin/Manager có các đường tương ứng `/pts/{id}/availability` (GET/PUT), `/pts/{id}/unavailability` (GET/POST) và `/pts/{id}/unavailability/{absenceId}` (DELETE); PT chỉ dùng ID thuộc mình. `available_at` kiểm tra một thời điểm; không thay kiểm tra toàn bộ interval khi triển khai Booking. Lịch tuần ISO thứ 1–7 tối đa 50 khung; khoảng nghỉ phân trang 20–100. Certificate upload `file` là PDF/PNG/JPEG tối đa 5 MiB, nội dung được kiểm tra; download là binary qua Bearer, không public path.

### Facilities & Sessions

| Method & path | Permission | Request chính | Ghi chú | REQ/UC |
|---|---|---|---|---|
| `GET /rooms` | `room/session.read` | status,page | danh sách | 04/UC-10 |
| `POST /rooms` | `room/session.manage` | name,room_type,default_capacity,status? | tên unique CI | 04/UC-10 |
| `GET /rooms/{id}` | `room/session.read` | — | chi tiết | 04/UC-10 |
| `GET /rooms/{id}/impact` | `room/session.manage` | — | số phiên tương lai và booking giữ chỗ | 04/UC-10 |
| `PATCH /rooms/{id}` | `room/session.manage` | trường cho phép | trạng thái có impact preview | 04/UC-10 |
| `GET /time-slots` | Authenticated | status=ACTIVE/INACTIVE,q,page,page_size | danh mục ca | 05/UC-11 |
| `GET /time-slots/{id}` | Authenticated | — | chi tiết ca | 05/UC-11 |
| `POST /time-slots` | `room/session.manage` | name,start_time,end_time | chặn overlap | 05/UC-11 |
| `PATCH /time-slots/{id}` | `room/session.manage` | name,start_time,end_time,reason | chỉ khi chưa dùng | 05/UC-11 |
| `PATCH /time-slots/{id}/status` | `room/session.manage` | is_active,reason | giữ nguyên snapshot lịch sử | 05/UC-11 |
| `GET /room-sessions` | Authenticated | date,room_id,status,q,page,page_size | số chỗ tính từ DB; chưa có from/to | 05/UC-11 |
| `GET /room-sessions/{id}` | Authenticated | — | chi tiết phiên | 05/UC-11 |
| `POST /room-sessions` | `room/session.manage` | room_id,session_date,time_slot_id,capacity? | 201 phiên duy nhất | 05/UC-11 |
| `PATCH /room-sessions/{id}` | `room/session.manage` | capacity,status,reason | lock + BR-09/10 | 05/UC-11 |

### Bookings, Attendance & Reviews

| Method & path | Permission | Request chính | Ghi chú | REQ/UC |
|---|---|---|---|---|
| `POST /bookings` | MEMBER SELF | room_session_id,pt_id? | bắt buộc Idempotency-Key | 06,08/UC-02 |
| `POST /members/{id}/bookings` | đặt hộ | room_session_id,pt_id? | created_by từ token | 06,08/UC-02 |
| `POST /bookings/batch` | MEMBER SELF | room_session_ids (1–24),pt_id? | bắt buộc Idempotency-Key; cùng phòng/ngày, ca liền kề; all-or-none, trả mảng theo thời gian | 06,08/UC-02 |
| `POST /members/{id}/bookings/batch` | đặt hộ | room_session_ids (1–24),pt_id? | cùng hợp đồng batch, created_by từ token | 06,08/UC-02 |
| `GET /bookings` | theo scope | from,to,status,room,pt,page | backend áp scope | 13/UC-13 |
| `GET /bookings/{id}` | SELF/ASSIGNED/ALL | — | object authorization | 13/UC-13 |
| `POST /bookings/{id}/pt-response` | PT ASSIGNED | ACCEPT/REJECT,reason? | lock + respond_by | 18/UC-03 |
| `POST /bookings/{id}/cancel` | theo scope | reason? | lý do bắt buộc cho override/PT confirmed | 07/UC-06 |
| `POST /bookings/{id}/attendance` | `attendance.record` | participant_type | member/PT phải thuộc booking | 09/UC-04 |
| `GET /bookings/{id}/attendance` | theo scope | — | danh sách từng người | 09/UC-04 |
| `POST /bookings/{id}/review` | MEMBER SELF | rating 1–5, comment tùy chọn ≤1000 ký tự | Chỉ buổi COMPLETED đã kết thúc, có PT và điểm danh của cả hai; một review/booking; audit transaction; không sửa/xóa | 19/UC-05 |
| `GET /bookings/{id}/review` | SELF/ASSIGNED/ALL | — | trả `data:null` khi chưa đánh giá; ngoài scope 404 | 19/UC-05 |
| `GET /pts/{id}/reviews` | Member: PT active/approved; PT SELF; vận hành ALL | page,page_size | Chỉ sao, nhận xét, ngày và nhãn “Hội viên”; không có ID booking/người dùng hay liên hệ | 19/UC-05 |

Hồ sơ PT và lựa chọn PT khi đặt lịch trả `rating_average` (null khi chưa có review) và `review_count`; cùng tính trực tiếp từ review đủ điều kiện, không dùng Trust Score.

### Reports

Hai endpoint dưới đây đã triển khai và kiểm thử PostgreSQL thật. Kỳ ngày `session_date` bao gồm cả hai đầu, dùng `Asia/Ho_Chi_Minh`, mặc định tháng hiện tại đến hôm nay và tối đa 366 ngày. Response chứa kỳ và `as_of`; bộ lọc/ngày không hợp lệ trả 422.

| Method & path | Permission | Request | Ghi chú | REQ/UC |
|---|---|---|---|---|
| `GET /reports/operations` | ADMIN, MANAGER | from,to,room_id? | Tổng trạng thái tại `as_of`; chuyên cần, no-show, pending-completion, hủy và sử dụng có tử số/mẫu số | 10/UC-14 |
| `GET /reports/pts/{id}` | PT SELF; ADMIN/MANAGER ALL | from,to | Rating và chuyên cần; nhóm loại trừ, thiếu chứng cứ và chờ completion riêng | 20/UC-14 |
| `POST /reports/exports` | `report.export` | type,format,filters | Could; sync với giới hạn nhỏ | 15/UC-15 |

`no_show_rate = NO_SHOW / (COMPLETED + NO_SHOW)` trên buổi đã kết thúc; mẫu số 0 trả null. `utilization` là seat-minutes đã đặt (một booking tính một chỗ) trên capacity-minutes của phiên OPEN đã kết thúc, không phải thời gian tập thực tế. Chuyên cần PT tính các buổi COMPLETED/NO_SHOW có `confirmed_at` và hủy bởi MEMBER/PT sau xác nhận theo BR-22; hủy FACILITY, REJECTED, EXPIRED và yêu cầu đang chờ bị loại. Booking lịch sử thiếu xác nhận hoặc nguồn hủy UNKNOWN được đếm riêng, không suy diễn. Rating trong kỳ lấy từ cùng predicate review hợp lệ với PT profile.

## 4. Trạng thái và payload chuẩn

Booking status: `PENDING_PT`, `CONFIRMED`, `REJECTED`, `EXPIRED`, `CANCELLED`, `COMPLETED`, `NO_SHOW`. PT status: `DRAFT`, `PENDING`, `APPROVED`, `REJECTED`, `SUSPENDED`. Room session: `OPEN`, `CLOSED`, `MAINTENANCE`. User: `ACTIVE`, `LOCKED`.

Ví dụ tạo booking:

```http
POST /api/v1/bookings
Authorization: Bearer <token>
Idempotency-Key: 5e7fd6a9-4cd5-49e1-931f-fdb019920b3b
Content-Type: application/json

{"room_session_id":321,"pt_id":18}
```

```json
{
  "data": {
    "id": 845,
    "status": "PENDING_PT",
    "room_session_id": 321,
    "pt_id": 18,
    "respond_by": "2026-10-20T17:30:00+07:00"
  },
  "request_id": "req_..."
}
```

Ví dụ tạo một dải ca liên tiếp (commit toàn bộ hoặc không tạo ca nào):

```http
POST /api/v1/bookings/batch
Authorization: Bearer <token>
Idempotency-Key: 838bedf7-19d1-4437-9d3c-ecb702a34868
Content-Type: application/json

{"room_session_ids":[321,322],"pt_id":18}
```

`data` là mảng Booking được sắp theo `starts_at`. Ca khác phòng/ngày hoặc không liền kề trả `422`; hết chỗ, trùng lịch, PT không rảnh hoặc dùng lại khóa với payload khác trả `409`. Cả dải được kiểm tra và ghi trong cùng transaction.

## 5. Transaction và retry

- Tạo/hủy/phản hồi/điểm danh/review/chỉnh capacity phải kiểm tra trạng thái nguồn và commit trước response.
- Deadlock/serialization failure có thể retry tối đa số lần cấu hình với backoff nhỏ; request timeout hoặc retry exhausted trả mã ổn định.
- Unique violation được ánh xạ sang conflict nghiệp vụ; không trả stack/SQL.
- Idempotency lưu actor, route, key, hash payload, response status/body và thời hạn; chỉ trả lại kết quả sau transaction gốc đã commit.

## 6. OpenAPI và Postman

Khi triển khai, tạo OpenAPI 3.x từ hợp đồng này. Postman Collection được sinh/đối chiếu với OpenAPI, không trở thành nguồn schema độc lập. Mọi thay đổi endpoint phải cập nhật REQ/UC/test tương ứng.
