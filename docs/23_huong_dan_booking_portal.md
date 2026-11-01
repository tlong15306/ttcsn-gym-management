# Chạy Booking và hai portal — 12/10/2026

Khởi động Docker Desktop, dùng cấu hình local đã tạo theo `17_huong_dan_chay_local.md`. Không reset DB dev; migration chỉ bổ sung cột ảnh.

PowerShell thứ nhất:

```powershell
Set-Location C:\Users\choqu\OneDrive\Desktop\TTCSN
docker compose --env-file .\infra\.env.compose up -d --wait
Set-Location .\server
$env:GOCACHE = Join-Path (Get-Location).Path '..\.artifact_work\go-cache'
go run .\cmd\migrate up
go run .\cmd\seed
# Tùy chọn: tạo ba phòng DEMO và ảnh minh họa, không thay ảnh/lịch đã có
go run .\cmd\seed-room-demo
go run .\cmd\api
```

PowerShell thứ hai:

```powershell
Set-Location C:\Users\choqu\OneDrive\Desktop\TTCSN\client
npm.cmd ci
npm.cmd run dev -- --host 127.0.0.1
```

Mở http://127.0.0.1:5173. Các tài khoản mới do seed tạo có mật khẩu `DemoPass123!`: `admin.demo`, `manager.demo`, `reception.demo`, `member.demo`, `pt.demo`. Seed không ghi đè mật khẩu đã đổi.

- Admin/Manager: chi tiết phòng → tải ảnh JPEG/PNG, nhập alt text → lưu. Chọn “ảnh minh họa” cho ảnh giả lập. DEMO rooms mới cần công bố lịch hoạt động trước khi có ca đặt; thao tác theo `21_huong_dan_lich_phong.md`.
- Hội viên: `/member/rooms` → nhấn phòng → chọn ngày/ca → tự tập hoặc PT rảnh → kiểm tra → xác nhận. Tập tự do xác nhận ngay; cùng PT giữ chỗ chờ tối đa 30 phút và không vượt giờ bắt đầu.
- PT: `/pt/schedule` lịch tuần desktop, danh sách theo ngày màn nhỏ; `/pt/requests` nhận/từ chối; `/pt/profile` quản lý chuyên môn, giờ làm và ngày nghỉ. Chỉ PT đã duyệt, tài khoản hoạt động và rảnh toàn bộ ca mới xuất hiện trong lựa chọn.
- Admin/Lễ tân: `/management/bookings` → đặt hộ → chọn hồ sơ → chọn phòng/ca/PT. Manager xem/hủy có lý do, không đặt hộ.
- Hội viên tự hủy ít nhất 2 giờ trước ca. PT hủy lịch đã nhận và vận hành hủy đều phải nhập lý do. Pending PT dùng từ chối.

Account demo MEMBER/PT không tự có hồ sơ nghiệp vụ hoặc phê duyệt. Admin/Lễ tân cần tạo/liên kết Hội viên, cập nhật thẻ; PT tạo hồ sơ/chứng chỉ/giờ làm, Admin/Manager xét duyệt. Không có KPI, rating hoặc trạng thái phê duyệt giả.

Token chỉ trong memory: reload cần đăng nhập lại; 401 chuyển về đăng nhập. Retry cùng nội dung dùng lại Idempotency-Key, không tạo booking mới. Options chỉ là thông tin, backend kiểm tra lại lúc xác nhận; có thể báo ca đầy hoặc PT vừa bận.

Ảnh thiếu/lỗi có placeholder. Upload giới hạn 5 MiB, chiều 4096px, tổng 16 triệu pixel. Thumbnail/detail là JPEG 480/1600px lưu DB; backup PostgreSQL bao gồm ảnh. Chạy seed-room-demo lần nữa giữ cover đã có, nhưng sẽ bổ sung lại cover đang trống cho đúng ba phòng DEMO.

Job expiration chạy khi API khởi động và mỗi 15 giây, xử lý tối đa 100 yêu cầu/lần. API dừng thì job dừng; khởi động lại xử lý backlog theo từng batch. Dừng ứng dụng bằng Ctrl+C ở hai terminal; `docker compose --env-file .\infra\.env.compose stop` dừng DB và giữ volume.

API mới: `openapi_admin_crud.json` và `GYM_Admin_CRUD.postman_collection.json`, tổng 69 operations. Điền booking_date/session/member/PT/booking IDs và booking_key trước khi gửi; giữ cùng key khi retry, đổi key khi đổi nội dung. Không Run All mutations trên dữ liệu thật.
