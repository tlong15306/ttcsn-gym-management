# Đặc tả REST API — v1.0 cho baseline SRS v0.3

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
| `POST /auth/logout` | Authenticated | none | 204; thu hồi JTI nếu bật session store | 01/UC-07 |
| `GET /me` | Authenticated | none | hồ sơ tài khoản + role/status phù hợp | 01/UC-07 |
| `PATCH /me` | `profile.self.update` | trường hồ sơ cho phép | không nhận role/status | 02/UC-07 |
| `GET /users` | `account.manage` | filter, page | danh sách phân trang | 12/UC-08 |
| `POST /users` | `account.manage` | username, password, email, role | 201 user | 12/UC-08 |
| `PATCH /users/{id}/status` | `account.manage` | ACTIVE/LOCKED, reason | tăng auth_version khi khóa | 12/UC-08 |
| `PATCH /users/{id}/role` | `account.manage` | role, reason | kiểm tra hồ sơ phụ thuộc | 12/UC-08 |
| `POST /users/{id}/reset-password` | `account.manage` | mật khẩu mới hoặc flow reset | tăng auth_version | 12/UC-08 |

### Members

| Method & path | Permission | Request chính | Ghi chú | REQ/UC |
|---|---|---|---|---|
| `GET /members` | `member.read ALL` | q,status,expiry,page | tìm mã/tên/SĐT/email | 03,11/UC-09 |
| `POST /members` | `member.create` | account + profile | tạo atomically hoặc rollback | 03/UC-09 |
| `GET /members/{id}` | SELF/ASSIGNED/ALL | — | field-level filtering cho PT | 03/UC-09 |
| `PATCH /members/{id}` | `member.update` | trường hồ sơ | không hard-delete | 03/UC-09 |
| `PATCH /members/{id}/membership` | `member.update` | expiry,status,reason | audit | 03/UC-09 |

### PT

| Method & path | Permission | Request chính | Ghi chú | REQ/UC |
|---|---|---|---|---|
| `GET /pts` | Authenticated | approved,available_at,q,page | member chỉ thấy APPROVED | 16,17/UC-12 |
| `GET /pts/me` | PT SELF | — | hồ sơ + quyết định gần nhất | 16/UC-12 |
| `PUT /pts/me/profile` | `pt.profile.update` | specialty,experience | DRAFT/REJECTED có thể sửa | 16/UC-12 |
| `POST /pts/me/submit` | `pt.profile.submit` | — | → PENDING | 16/UC-12 |
| `POST /pts/me/certificates` | PT SELF | multipart | private object, allowlist | 16/UC-12 |
| `GET /pts/{id}/certificate` | `certificate.read` | — | kiểm tra quyền mỗi lần tải | 16,17/UC-01 |
| `POST /pts/{id}/decision` | `pt.approve/reject/suspend` | action,reason | state transition + audit | 17/UC-01 |
| `GET/PUT /pts/me/availability` | `pt.availability.manage` | weekly schedule | validate không phá lịch | 16/UC-12 |
| `GET/POST/DELETE /pts/me/unavailability` | SELF | from,to,reason | kiểm tra booking bị ảnh hưởng | 16/UC-12 |

### Facilities & Sessions

| Method & path | Permission | Request chính | Ghi chú | REQ/UC |
|---|---|---|---|---|
| `GET /rooms` | `room/session.read` | status,page | danh sách | 04/UC-10 |
| `POST /rooms` | `room/session.manage` | name,type,default_capacity | tên unique CI | 04/UC-10 |
| `PATCH /rooms/{id}` | `room/session.manage` | trường cho phép | trạng thái có impact preview | 04/UC-10 |
| `GET /time-slots` | Authenticated | active | danh mục ca | 05/UC-11 |
| `POST /time-slots` | `room/session.manage` | name,start,end | chặn overlap | 05/UC-11 |
| `PATCH /time-slots/{id}/status` | `room/session.manage` | ACTIVE/INACTIVE | không sửa giờ ca đã dùng | 05/UC-11 |
| `GET /room-sessions` | Authenticated | from,to,room,status | số chỗ tính từ DB | 05/UC-11 |
| `POST /room-sessions` | `room/session.manage` | room,date,slot,capacity | 201 phiên duy nhất | 05/UC-11 |
| `PATCH /room-sessions/{id}` | `room/session.manage` | capacity,status,reason | lock + BR-09/10 | 05/UC-11 |

### Bookings, Attendance & Reviews

| Method & path | Permission | Request chính | Ghi chú | REQ/UC |
|---|---|---|---|---|
| `POST /bookings` | MEMBER SELF | room_session_id,pt_id? | bắt buộc Idempotency-Key | 06,08/UC-02 |
| `POST /members/{id}/bookings` | đặt hộ | room_session_id,pt_id? | created_by từ token | 06,08/UC-02 |
| `GET /bookings` | theo scope | from,to,status,room,pt,page | backend áp scope | 13/UC-13 |
| `GET /bookings/{id}` | SELF/ASSIGNED/ALL | — | object authorization | 13/UC-13 |
| `POST /bookings/{id}/pt-response` | PT ASSIGNED | ACCEPT/REJECT,reason? | lock + respond_by | 18/UC-03 |
| `POST /bookings/{id}/cancel` | theo scope | reason? | lý do bắt buộc cho override/PT confirmed | 07/UC-06 |
| `POST /bookings/{id}/attendance` | `attendance.record` | participant_type | member/PT phải thuộc booking | 09/UC-04 |
| `GET /bookings/{id}/attendance` | theo scope | — | danh sách từng người | 09/UC-04 |
| `POST /bookings/{id}/review` | MEMBER SELF | rating,comment | không nhận member_id/pt_id | 19/UC-05 |
| `GET /pts/{id}/reviews` | Authenticated | page | review hợp lệ, dữ liệu công khai | 19/UC-05 |

### Reports

| Method & path | Permission | Request | Ghi chú | REQ/UC |
|---|---|---|---|---|
| `GET /reports/operations` | `report.read` | from,to,room | kèm công thức/meta timezone | 10/UC-14 |
| `GET /reports/pts/{id}` | SELF/ALL | from,to | rating + attendance rate riêng | 20/UC-14 |
| `POST /reports/exports` | `report.export` | type,format,filters | Could; sync với giới hạn nhỏ | 15/UC-15 |

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

## 5. Transaction và retry

- Tạo/hủy/phản hồi/điểm danh/review/chỉnh capacity phải kiểm tra trạng thái nguồn và commit trước response.
- Deadlock/serialization failure có thể retry tối đa số lần cấu hình với backoff nhỏ; request timeout hoặc retry exhausted trả mã ổn định.
- Unique violation được ánh xạ sang conflict nghiệp vụ; không trả stack/SQL.
- Idempotency lưu actor, route, key, hash payload, response status/body và thời hạn; chỉ trả lại kết quả sau transaction gốc đã commit.

## 6. OpenAPI và Postman

Khi triển khai, tạo OpenAPI 3.x từ hợp đồng này. Postman Collection được sinh/đối chiếu với OpenAPI, không trở thành nguồn schema độc lập. Mọi thay đổi endpoint phải cập nhật REQ/UC/test tương ứng.
