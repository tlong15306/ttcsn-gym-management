# Runbook vận hành, migration và khôi phục

## 1. Phân loại artifact

- `09_schema.sql`: bootstrap CSDL mới/rỗng; không chứa DROP; không chạy trên dữ liệu hiện hữu.
- `migrations/NNNN_description.up.sql`: thay đổi production có phiên bản.
- `seeds/dev.sql`, `seeds/test.sql`: dữ liệu giả cho dev/test; không chạy production.
- Reset demo nếu cần phải là script riêng, tên có chữ `reset_demo`, chỉ cho môi trường được xác nhận.

## 2. Quy trình migration

### Trước thay đổi

1. Ghi phiên bản app/schema, chủ sở hữu thay đổi, cửa sổ và tiêu chí thành công.
2. Kiểm tra dung lượng/bất thường dữ liệu; ước lượng lock và thời gian backfill.
3. Tạo backup và thực hiện restore test gần nhất theo mức rủi ro.
4. Chạy migration trên bản sao dữ liệu; lưu log và truy vấn kiểm tra.
5. Xác nhận code trong rollout tương thích schema trước/sau migration.

### Thực hiện

Áp dụng expand → backfill → validate → switch → contract:

1. Expand: thêm bảng/cột nullable/index concurrent nếu phù hợp.
2. Backfill: chạy theo batch, có checkpoint và báo lỗi ngoại lệ.
3. Validate: so sánh count, FK, duplicate, status mapping và query nghiệp vụ.
4. Switch: deploy code đọc/ghi cấu trúc mới, theo dõi metric/log.
5. Contract: chỉ xóa cấu trúc cũ ở release sau khi xác nhận không còn consumer.

Không coi down migration phá dữ liệu là rollback an toàn. Ưu tiên roll-forward; nếu phải phục hồi backup, dừng ghi và ghi rõ RPO bị ảnh hưởng.

## 3. Nâng từ schema v0.2

1. Inventory các bảng/cột/index và xuất thống kê status combinations.
2. Thêm users status/auth_version/full_name, các bảng mới và cột trạng thái mới chưa bắt buộc.
3. Tạo `room_sessions` từ room/date/slot cũ; báo booking không ánh xạ được.
4. Map booking:
   - không PT, active hợp lệ → CONFIRMED;
   - có PT + pt_status confirmed → CONFIRMED;
   - có PT + pending → PENDING_PT và yêu cầu chọn `respond_by` hợp lệ;
   - cancelled/completed theo bằng chứng cũ;
   - tổ hợp khác → staging ngoại lệ, không đoán.
5. Review cũ thiếu booking: chuyển vào bảng/staging legacy hoặc export đối chiếu; không tính rating verified.
6. Không sinh attendance từ booking status nếu thiếu timestamp/người ghi.
7. Làm sạch duplicate/xung đột trước khi bật unique/exclusion/FK NOT NULL.
8. Chạy test MIG-01 và báo cáo count trước/sau.

## 4. Backup và restore

### Backup hằng ngày

- Ghi start/end, exit code, file size, checksum, PostgreSQL/schema version và vị trí.
- Mã hóa/phân quyền nơi lưu; retention do nhóm chốt theo dung lượng.
- Job thất bại phải cảnh báo; file tồn tại không đủ chứng minh backup dùng được.

### Restore drill

1. Tạo DB cô lập rỗng đúng PostgreSQL major version.
2. Restore backup; ghi thời gian bắt đầu/kết thúc và lỗi.
3. Kiểm tra schema migration version, row count, orphan FK, duplicate active booking, review/attendance consistency.
4. Khởi động app ở chế độ test và chạy smoke login/read/create/cancel bằng dữ liệu riêng.
5. Ghi RTO/RPO thực tế, PASS/FAIL và việc khắc phục; xóa môi trường drill theo chính sách.

## 5. Job nền

- Mỗi job có run ID, lease/lock nếu nhiều instance, timeout, retry giới hạn và metric.
- Expire/finalize dùng UPDATE có điều kiện trạng thái nguồn; chạy lại không đổi kết quả.
- Nếu job ngừng: khôi phục scheduler, chạy catch-up theo batch, kiểm tra số bản ghi và cảnh báo; không sửa tay hàng loạt khi chưa backup.

## 6. Sự cố booking

### Dấu hiệu

Room full sai, booking vượt capacity, member/PT trùng ca, nhiều lock wait/deadlock, pending không hết hạn.

### Xử lý

1. Ghi thời điểm, request IDs, session/member/PT IDs; hạn chế tạo booking nếu toàn vẹn đang bị đe dọa.
2. Chụp snapshot truy vấn và log; không xóa bản ghi để “sửa nhanh”.
3. Kiểm tra migration/version, job, trạng thái giữ chỗ, transaction retry và idempotency.
4. Xác định bản ghi hợp lệ theo audit/commit time/quy tắc; chuẩn bị script sửa có preview và backup.
5. Chạy test race/regression trước mở lại; lập post-incident note.

## 7. Sự cố token/quyền

- Khi tài khoản nghi bị lộ: LOCKED + tăng auth_version; thu hồi JTI/session nếu có; kiểm tra audit.
- Khi đổi role/reset password: tăng auth_version và yêu cầu đăng nhập lại.
- Không gửi token/log/DB dump qua kênh không an toàn; không ghi token vào ticket/report.

## 8. Checklist release

- [ ] Decision/SRS/API/schema/migration/changelog cùng version.
- [ ] Migration test + backup/restore evidence.
- [ ] Must tests, race/idempotency/authz pass.
- [ ] Secret/config/CORS/file policy được kiểm tra.
- [ ] Job/health/metrics/log và cảnh báo hoạt động.
- [ ] Kế hoạch roll-forward/rollback, chủ sở hữu và cửa sổ rõ.
- [ ] Giới hạn chưa kiểm chứng được ghi trong release note.
