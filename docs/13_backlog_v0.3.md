# Product Backlog v0.3

Trạng thái baseline: **Đã phân tích**. “Đã phân tích” không đồng nghĩa “Đã triển khai” hoặc “Đã kiểm thử”.

## 1. Backlog chức năng

| ID | Hạng mục | Ưu tiên | UC | Phụ thuộc chính | Trạng thái |
|---|---|---:|---|---|---|
| REQ-01 | Xác thực phiên | Must | UC-07 | users, auth_version, NFR-03 | Đã phân tích |
| REQ-02 | Phân quyền role + object scope | Must | Mọi UC | REQ-01, matrix quyền | Đã phân tích |
| REQ-03 | Quản lý hội viên | Must | UC-09 | REQ-01/02 | Đã phân tích |
| REQ-04 | Quản lý phòng | Must | UC-10 | REQ-01/02 | Đã phân tích |
| REQ-05 | Quản lý ca và phiên phòng | Must | UC-11 | REQ-04, BR-06…10 | Đã phân tích |
| REQ-06 | Tạo booking/PT tùy chọn | Must | UC-02 | REQ-03/05/08/16/17 | Đã phân tích |
| REQ-07 | Hủy booking | Must | UC-06 | REQ-06/08 | Đã phân tích |
| REQ-08 | Chống xung đột/quá tải | Must | Include | REQ-05/06, transaction | Đã phân tích |
| REQ-09 | Điểm danh từng người | Must | UC-04 | REQ-06/07 | Đã phân tích |
| REQ-10 | Báo cáo vận hành | Should | UC-14 | REQ-06/09/19/20 | Đã phân tích |
| REQ-11 | Tìm kiếm/lọc hội viên | Should | UC-09 | REQ-03 | Đã phân tích |
| REQ-12 | Quản lý tài khoản | Must | UC-08 | REQ-01/02 | Đã phân tích |
| REQ-13 | Xem lịch | Should | UC-13 | REQ-06/07/18 | Đã phân tích |
| REQ-14 | Phản hồi và polling yêu cầu | Should | UC-16 | REQ-06/18 | Đã phân tích |
| REQ-15 | Xuất PDF/Excel | Could | UC-15 | REQ-10 | Đã phân tích, có thể hoãn |
| REQ-16 | Nộp hồ sơ PT | Must | UC-12 | REQ-01/12, file policy | Đã phân tích |
| REQ-17 | Duyệt/đình chỉ PT | Must | UC-01 | REQ-16/02 | Đã phân tích |
| REQ-18 | PT phản hồi lịch | Must | UC-03 | REQ-06/14/17 | Đã phân tích |
| REQ-19 | Đánh giá PT | Should | UC-05 | REQ-09/18 | Đã phân tích |
| REQ-20 | Rating/chuyên cần PT | Should | UC-05/14 | REQ-09/19 | Đã phân tích |

## 2. Thứ tự triển khai theo phụ thuộc

1. Identity: REQ-01, 02, 12 và audit nền tảng.
2. Master data: REQ-03, 04, 05, 11, 16, 17.
3. Booking core: REQ-06, 08, 07, 18, 13, 14.
4. Operations: REQ-09, 19, 20.
5. Reporting: REQ-10, sau đó REQ-15 nếu còn thời gian.

Không triển khai REQ-06 trước khi schema phiên phòng, trạng thái booking, unique index và giao thức transaction được kiểm thử. REQ-09 là Must vì REQ-19/20 phụ thuộc dữ liệu attendance.

## 3. Backlog phi chức năng

| ID | Hạng mục | Cổng nghiệm thu | Trạng thái |
|---|---|---|---|
| NFR-01 | p95 API <2s theo workload công bố | Báo cáo benchmark có cấu hình/dữ liệu/percentile | Chưa chạy |
| NFR-02 | 100 concurrent users + tính đúng race | Load report và race assertions | Chưa chạy |
| NFR-03 | Bảo mật auth/authz/upload/log | Test quyền, token revoke, file, secret scan | Chưa chạy |
| NFR-04 | Responsive/usability/accessibility cơ bản | Browser matrix + checklist thao tác | Chưa chạy |
| NFR-05 | Backup/restore RPO/RTO | Restore drill có timestamp và kiểm tra dữ liệu | Chưa chạy |

## 4. Ngoài phạm vi / phiên bản sau

Thanh toán và doanh thu; nhiều cơ sở; đa vai trò/tài khoản; lớp học nhóm; lịch ca linh hoạt chồng giờ; app native; QR/RFID; email/SMS/WebSocket; notification lưu bền; quyền động; Trust Score tổng hợp.

Mỗi mục chỉ chuyển vào MVP qua decision log, cập nhật SRS/ERD/API/test và ước lượng lại tiến độ.
