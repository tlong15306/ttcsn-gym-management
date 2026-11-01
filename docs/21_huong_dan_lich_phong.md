# Lịch hoạt động phòng — hướng dẫn demo local

Ngày cập nhật: 11/10/2026. Lịch tuần và ngoại lệ bổ sung cho module Phòng/Ca/Phiên; Booking/Attendance/Review/Report tiếp tục triển khai sau. Không thay lịch/phân công hoặc ghi nhận hoàn thành trong Excel.

## Chạy ứng dụng

Dùng quy trình PowerShell tại [17 — Chạy local](17_huong_dan_chay_local.md): mở Docker Desktop, chạy Compose, vào `server`, `go run .\cmd\migrate up`, `go run .\cmd\seed`, `go run .\cmd\api`; terminal khác vào `client`, `npm.cmd install`, `npm.cmd run dev -- --host 127.0.0.1`. API `127.0.0.1:8080`, UI `127.0.0.1:5173`, PostgreSQL riêng cổng 5433. Migration v3 nâng cấp thêm, không chạy bootstrap schema/reset trên DB dev. Dừng bằng Ctrl+C; không tắt máy.

Đăng nhập `manager.demo` hoặc `admin.demo`, mật khẩu demo `DemoPass123!`. Vào **Phòng tập → phòng → Đăng ký lịch hoạt động**; tạo phòng mới chuyển thẳng sang đăng ký lịch. Hội viên/PT/Lễ tân có **Lịch phòng tập** chỉ đọc. Tải lại trang yêu cầu đăng nhập lại vì token chỉ ở memory.

## Đăng ký lịch tuần

1. Chọn Mở giờ/Đóng giờ, kéo vùng trên bảng 7 ngày × 24 cột; thả chuột chỉ sửa nháp. Ô 07–20 mở 07:00–21:00 (14 giờ). Đóng cả tuần là một lịch hợp lệ, khác phòng chưa đăng ký.
2. Di chuyển/Đổi đầu/Đổi cuối: bắt đầu từ ô đang mở. Có form chọn ngày, giờ đầu/cuối, sao chép ngày, preset cả tuần 07–21 và nút hoàn tác/làm lại. Form qua nửa đêm yêu cầu xác nhận tách hai ngày; Chủ nhật tiếp nối Thứ Hai. Ca booking vẫn có độ dài riêng.
3. Bàn phím: mũi tên/Home/End, Shift + mũi tên chọn vùng, Enter áp dụng, Esc hủy vùng đang chọn. Form là phương án không cần kéo. Trên điện thoại mặc định danh sách 7 ngày; có thể bật bảng cuộn ngang bên trong.
4. Nhập ngày áp dụng và lý do → **Xem trước và xác nhận** → đọc giờ từng ngày/số phiên mới/phiên đóng/mở lại/số chỗ giữ → **Xác nhận đăng ký**. Phòng có lịch chỉ đổi tuần từ ngày mai; thay đổi hôm nay dùng ngoại lệ. Lịch đầu tiên có thể bắt đầu hôm nay nhưng chỉ tạo phiên chưa bắt đầu.
5. Thành công mới ghi DB. Lỗi 409/422/mạng giữ nháp; nếu lịch/ca/booking đổi hoặc preview hết 5 phút, tải lại và xem trước lại. Tải lại dữ liệu khi có nháp yêu cầu xác nhận bỏ nháp. 401 kết thúc phiên, không tự công bố lại.

## Ngoại lệ và lịch thực tế

**Ngoại lệ theo ngày:** nghỉ cả ngày CLOSED hoặc REPLACE toàn bộ giờ của ngày đó bằng một hay nhiều khoảng. Có loại ngày nghỉ/bảo trì/khác và lý do. Chỉnh sửa giữ bản cũ WITHDRAWN, tạo bản mới; rút ngoại lệ phục hồi lịch tuần, giữ lịch sử. Không sửa/rút ngày đã qua. Qua nửa đêm tạo hai ngoại lệ riêng. Phòng MAINTENANCE/INACTIVE luôn đóng kể cả ngoại lệ REPLACE.

**Lịch thực tế:** chọn ngày bắt đầu xem 7 ngày và nguồn giờ mở cửa. Ngoại lệ chỉ áp dụng đúng ngày, còn mẫu tuần lặp qua các tuần. **Lịch sử** giữ versions/ngoại lệ, phân trang 20 dòng. Các role chỉ đọc không nhận lý do quản trị, bản nháp hoặc danh tính booking qua calendar API.

## Phiên và dữ liệu cũ

- Chỉ ca đang hoạt động nằm trọn một khoảng mở mới sinh phiên; không cắt ca tại mép giờ hoặc xuyên qua khoảng nghỉ. Form ca hỗ trợ tùy chọn kết thúc 24:00; bắt đầu phải nhỏ hơn 24:00.
- Công bố/sửa lịch tự bổ sung 28 ngày; preview/confirm cho phép 1–56 ngày. Job chạy lúc API khởi động và 00:05 Việt Nam, bù cửa sổ sau downtime. Nút chuẩn bị 28 ngày chỉ tạo thiếu; job/EXTEND không đổi snapshot hoặc tự mở lại phiên đóng.
- Không giảm/đóng phiên tương lai có PENDING_PT/CONFIRMED; preview liệt kê phiên/giờ/chỗ bị ảnh hưởng và confirm kiểm tra lại trong transaction. Chưa có UI hủy Booking, không tự hủy để công bố được.
- Phiên quá khứ/đang diễn ra và capacity cũ giữ nguyên. Phiên đóng thủ công MANUAL không tự mở lại. Phiên đóng vì lịch có thể mở lại khi lịch tương ứng thay đổi, chưa bắt đầu, không overlap và đủ điều kiện. Phiên do trạng thái phòng đóng mang nguồn ROOM, cần thao tác riêng khi mở lại phòng.
- Phòng có trước migration v3 có `calendar_required=false`: giữ chế độ chuyển đổi, không suy mẫu tuần từ phiên demo. Công bố tuần đầu tiên bật gate và kiểm tra booking/phiên hiện có. Phòng mới có gate ngay, chưa có lịch không tạo phiên. Sau chuyển đổi không có endpoint tắt gate.
- Nguồn phiên CALENDAR, version/exception và audit SYSTEM cho job; không gán thao tác tự động cho sinh viên. Job thất bại chỉ log mã an toàn và báo lỗi chuẩn bị phiên trong trang lịch.

## Hợp đồng API

Tất cả dưới `/api/v1`, envelope `data`/`request_id`; lỗi `error.code/message`/`request_id`. Chín operations mới (tổng hiện tại 58):

| Endpoint | Quyền và dữ liệu |
|---|---|
| GET `/rooms/{id}/weekly-schedule` | Admin/Manager, snapshot phòng/versions/ngoại lệ/ca/phiên tương lai |
| POST `/rooms/{id}/calendar/preview` | Admin/Manager, mutation WEEKLY/EXCEPTION/WITHDRAW/EXTEND |
| PUT `/rooms/{id}/weekly-schedule` | WEEKLY + preview_id + Idempotency-Key |
| GET `/rooms/{id}/calendar?from&to` | Mọi role đã đăng nhập, ngày inclusive tối đa 56 ngày, giờ đã công bố |
| GET `/rooms/{id}/calendar/history` | Admin/Manager, page/page_size mặc định 20, tối đa 100 |
| POST `/rooms/{id}/calendar/exceptions` | EXCEPTION mới |
| PATCH `/rooms/{id}/calendar/exceptions/{exceptionId}` | EXCEPTION thay thế, ID body khớp path, giữ ngày |
| POST `/rooms/{id}/calendar/exceptions/{exceptionId}/withdraw` | WITHDRAW, ID body khớp path |
| POST `/rooms/{id}/calendar/materialize` | EXTEND, chỉ bổ sung thiếu |

Mutation bắt buộc `kind`, `base_revision`, `hours`, `reason`. WEEKLY thêm `effective_from`, weekday 1–7; EXCEPTION thêm `date/mode/closure_type`, weekday 0; WITHDRAW thêm `exception_id`; EXTEND thêm `days` tùy chọn. Giờ là `start_minute/end_minute` bước 60, 0≤start<end≤1440. Khoảng liền kề hợp nhất; overlap bị 422. Không gửi trường không liên quan. Tất cả confirm gửi nguyên mutation đã preview cùng `preview_id` và key dài 8–100 ký tự.

Preview được ký bởi khóa riêng trong lần chạy API, ràng buộc actor/phòng/payload/fingerprint; preview chưa commit cần xem trước lại sau khi khởi động lại API. Confirm khóa phòng → advisory lock danh mục ca → phiên theo ID; cùng key/payload trả kết quả đã commit, khác payload 409. Request đã commit được replay cả sau khi preview hết hạn hoặc API khởi động lại, nhờ bản ghi idempotency trong DB. Audit thất bại rollback lịch/ngoại lệ/phiên/revision/idempotency cùng nhau. Các mã 409 hiện dùng `INVALID_STATE` và thông báo an toàn.

OpenAPI: [openapi_admin_crud.json](openapi_admin_crud.json); collection: [GYM_Admin_CRUD.postman_collection.json](GYM_Admin_CRUD.postman_collection.json). Trong Postman: đăng nhập → GET weekly để lấy revision → đặt `effective_date`/`exception_date` phù hợp → sửa mutation preview → preview lấy ID/key → confirm cùng mutation. Preview mẫu mặc định WEEKLY; ngoại lệ/withdraw/extend phải đổi body preview tương ứng. Retry confirm giữ nguyên key/body; workflow mới lấy revision và preview mới. Collection đã sinh/đối chiếu, chưa chạy bằng Postman.

Backend giữ `handler/v1 → service/v1 → repository`; DTO riêng, models/intervals ở `internal/domain/facility`, job ở `internal/job`. Database có 18 bảng nghiệp vụ, schema cài mới riêng và migration v3 riêng. Không có package model trùng các domain hiện tại.
