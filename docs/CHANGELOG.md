# Changelog tài liệu

## v0.3 — 10/10/2026

### Baseline mới

- Chốt D01–D10 và BR-01…BR-26; xác định rõ phạm vi một cơ sở và các tính năng ngoài MVP.
- Chuẩn hóa 5 tác nhân, 16 use case và một ma trận quyền có object scope.
- Tạo SRS/backlog v0.3; giữ v0.1/v0.2 làm lịch sử.
- Chuẩn hóa vòng đời booking: PENDING_PT, CONFIRMED, REJECTED, EXPIRED, CANCELLED, COMPLETED, NO_SHOW.
- Tách phiên phòng, attendance từng người, review theo booking, idempotency và audit.
- Thay schema reset phá dữ liệu bằng bootstrap schema không có DROP; bổ sung runbook migration/restore.
- Chọn modular monolith và định nghĩa ranh giới handler/use case/repository/transaction.
- Bổ sung hợp đồng API, error code, kế hoạch kiểm thử và cổng nghiệm thu.

### Sửa mâu thuẫn

- Đồng bộ 5 vai trò giữa khảo sát, tác nhân và quyền.
- Sửa ID UC/REQ bằng danh mục UC-01…UC-16 duy nhất.
- Tách trạng thái tài khoản, hồ sơ hội viên, PT và booking.
- Bỏ doanh thu/thanh toán khỏi MVP vì chưa có nguồn giao dịch.
- Không dùng trạng thái booking thay attendance; không cho review thiếu booking hợp lệ.
- Hạ các tuyên bố hiệu năng/tính tuyệt đối thành mục tiêu cần kiểm thử.

### Giới hạn

- Chưa chạy schema trên PostgreSQL 16+.
- Chưa có migration thực thi từ dữ liệu v0.2; runbook mới mô tả quy trình/mapping.
- Chưa có mã nguồn/OpenAPI/Postman/test report/benchmark/restore drill.
