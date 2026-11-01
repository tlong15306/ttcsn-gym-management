# Ma trận phân quyền — v3.0

Đây là nguồn chính sách quyền duy nhất của baseline v0.3. Mặc định mọi quyền là **Deny**. Backend kiểm tra vai trò, phạm vi bản ghi, trạng thái và điều kiện môi trường trên mọi request.

## Quyền khách chưa đăng nhập — bổ sung U1 ngày 31/10/2026

Khách và người đã đăng nhập đều có thể dùng `GET /api/v1/public/rooms`, `/public/rooms/:id`, `/public/rooms/:id/cover-image`, `/public/rooms/:id/sessions`, `/public/trainers`. Chỉ trả projection đọc công khai: phòng AVAILABLE/MAINTENANCE, ảnh tương ứng, HLV APPROVED có tài khoản ACTIVE với role PT, lịch đã công bố/phiên legacy được phép dùng và số chỗ tổng hợp. Lịch giới hạn tối đa 7 ngày mỗi truy vấn, phân trang tối đa 100. Không có mutation trong namespace public; API quản trị, đặt lịch, profile, chứng chỉ và dữ liệu cá nhân vẫn theo ma trận xác thực bên dưới.

## Vai trò đã xác thực

Ký hiệu: `SELF` dữ liệu của mình; `ASSIGNED` booking được gán; `ALL` toàn cơ sở; `OPS` thao tác nghiệp vụ có điều kiện; `—` không có quyền.

| Permission | Hội viên | PT | Lễ tân | Quản lý | Admin | Điều kiện bắt buộc |
|---|---:|---:|---:|---:|---:|---|
| `profile.self.read/update` | SELF | SELF | SELF | SELF | SELF | Không sửa role/status/hạn thẻ/quyết định duyệt. |
| `member.read` | SELF | ASSIGNED tối thiểu | ALL | ALL | ALL | PT chỉ xem dữ liệu cần cho buổi được gán. |
| `member.create/update/deactivate` | — | — | OPS | OPS | OPS | Không hard-delete; thay đổi hạn thẻ có audit. |
| `pt.profile.submit/update` | — | SELF | — | — | — | PT DRAFT/REJECTED được sửa và nộp; PENDING chỉ xem. |
| `pt.approve/reject/suspend` | — | — | — | ALL | ALL | Ghi quyết định, lý do, người thực hiện. |
| `pt.public.read` | ALL | ALL | ALL | ALL | ALL | Chỉ hồ sơ APPROVED; không lộ chứng chỉ riêng tư. |
| `pt.availability.manage` | — | SELF | — | ALL | ALL | Không làm mất lịch đã CONFIRMED. |
| `room/session.read` | ALL | ALL | ALL | ALL | ALL | Chỉ dữ liệu phù hợp giao diện vai trò. |
| `room/session.manage` | — | — | — | ALL | ALL | Tuân thủ BR-06…BR-10 và audit thay đổi nhạy cảm. |
| `booking.create` | SELF | — | OPS đặt hộ | — | OPS hỗ trợ | Lễ tân/Admin ghi người tạo; BR-03…BR-15. |
| `booking.read` | SELF | ASSIGNED | ALL | ALL | ALL | Object-level authorization. |
| `booking.respond_pt` | — | ASSIGNED | — | — | — | Chỉ PENDING_PT, chưa hết hạn. |
| `booking.cancel_self` | SELF | ASSIGNED | — | — | — | Hội viên theo mốc 2 giờ; PT có lý do theo BR-17. |
| `booking.cancel_any` | — | — | OPS | OPS | OPS | Lý do và audit bắt buộc. |
| `attendance.record` | — | — | ALL | ALL | ALL | Người được điểm danh phải thuộc booking. |
| `attendance.read` | SELF | ASSIGNED | ALL | ALL | ALL | Không lộ lịch người khác cho MEMBER/PT. |
| `review.create` | SELF | — | — | — | — | BR-21; backend suy ra member/PT từ booking. |
| `review.read` | ALL hợp lệ | SELF nhận được | ALL | ALL | ALL | Ẩn dữ liệu riêng theo chính sách giao diện. |
| `report.read/export` | — | SELF metrics | — | ALL | ALL | PT chỉ rating/chuyên cần của mình; không có doanh thu. |
| `account.manage` | — | — | — | — | ALL | Tạo, reset, khóa/mở, gán một role; audit. |
| `audit.read` | — | — | — | Phạm vi nghiệp vụ | ALL | Không sửa/xóa audit. |
| `certificate.read` | — | SELF | — | ALL | ALL | Endpoint có xác thực; file không public. |

## Quy tắc thực thi

1. JWT hợp lệ mới xác thực danh tính; quyền luôn được quyết định lại tại backend.
2. ID trong URL/body không chứng minh quyền. Repository/use case phải tải bản ghi và kiểm tra SELF/ASSIGNED.
3. Thay đổi role, khóa tài khoản hoặc reset mật khẩu tăng `auth_version`; token có version cũ bị từ chối.
4. Frontend dùng ma trận để hiện/ẩn thao tác, nhưng phản hồi `403 FORBIDDEN` của server là quyết định cuối.
5. Trả `404 NOT_FOUND` thay cho việc xác nhận sự tồn tại của bản ghi nhạy cảm ngoài phạm vi khi phù hợp.
