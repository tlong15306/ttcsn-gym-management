# Runbook vận hành, migration và khôi phục

## 1. Phân loại artifact

- `09_schema.sql`: bootstrap CSDL mới/rỗng; không chứa DROP; không chạy trên dữ liệu hiện hữu.
- `server/internal/db/migrations/NNNN_description.sql`: thay đổi production có phiên bản.
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

## 9. Lịch phòng local — migration v3 và sinh phiên

Trên DB dev chỉ chạy `go run .\cmd\migrate up`; v3 thêm bảng/constraints, giữ ID/FK và từ chối dữ liệu phiên OPEN overlap. Không xóa phiên legacy để migration qua; kiểm tra các cặp phòng/giờ xung đột và xử lý có lý do/audit trước khi thử lại. Rollback/reset chỉ đúng `ttcsn_gym_test`, không dùng để nâng cấp hoặc dừng demo.

Sau upgrade, phòng legacy chưa có gate lịch; quản lý công bố tuần đầu tiên để chuyển có chủ đích. Phòng mới có gate ngay. API startup bù phiên còn thiếu, sau đó chạy lúc 00:05 Việt Nam, cửa sổ 28 ngày. Kiểm tra `calendar_prepared_through`/`calendar_job_error` trên trang lịch; xử lý lỗi an toàn trong API log rồi xem trước/xác nhận chuẩn bị lại. Không tự đổi capacity/snapshot hoặc mở lại phiên MANUAL/ROOM. Chi tiết retry, preview sau restart và tài khoản demo tại `21_huong_dan_lich_phong.md`.

## 10. Review và báo cáo vận hành

- Admin/Manager xem **Báo cáo** tại `/management/reports`, lọc theo kỳ và phòng, rồi tra cứu một PT. PT chỉ xem `/pt/performance` của mình và `/pt/reviews`; Member/Lễ tân không có quyền đọc báo cáo.
- Kỳ mặc định là tháng hiện tại theo Asia/Ho_Chi_Minh; cho phép tối đa 366 ngày. `as_of` cho biết thời điểm chụp. Trạng thái booking là trạng thái hiện tại của các buổi trong kỳ, không phải số sự kiện thay đổi trạng thái trong kỳ.
- `N/A` nghĩa là mẫu số bằng 0 hoặc chưa có đánh giá; không đổi thành 0%. Chỉ số sử dụng là seat-minute đã đặt trên phiên OPEN đã kết thúc, không phải thời gian tập đo được và không phải doanh thu.
- PT review dùng nhãn ẩn danh “Hội viên”; response không chứa định danh hay thông tin liên hệ người đánh giá. API report và các phép tính nguồn được kiểm thử bằng integration test PostgreSQL; frontend tests dùng API mock.

### Trình tự demo nghiệp vụ khi chủ động chạy local

1. Theo các bước trong `17_huong_dan_chay_local.md`, chạy PostgreSQL, migrate/seed, API và client. Nếu cần phòng/ảnh/lịch synthetic, chỉ chạy fixture opt-in đã mô tả trong hướng dẫn local.
2. Admin/Manager kiểm tra phòng và lịch được công bố. Member chọn một phòng, xem trang chi tiết/ảnh, chọn ca và PT nếu muốn, rồi xác nhận booking.
3. PT đăng nhập và xử lý yêu cầu. Trong buổi hợp lệ, vận hành ghi attendance Hội viên và PT riêng. Chờ buổi qua giờ và job completion xử lý; không sửa timestamp hoặc trạng thái để giả lập buổi đã hoàn thành.
4. Hội viên mở chi tiết booking `COMPLETED` đủ điều kiện và gửi review. PT đối chiếu review tại `/pt/reviews`, hiệu suất tại `/pt/performance`; Admin/Manager lọc đúng kỳ/phòng tại `/management/reports` và đối chiếu booking nguồn.
5. Nếu không có buổi trong kỳ đã kết thúc phù hợp, phần completion/review/report chỉ có thể kiểm tra bằng test DB/API; không ghi nhận demo end-to-end UI đã đạt. Kết thúc thì Ctrl+C hai terminal ứng dụng; chỉ dừng Compose khi muốn dừng cả PostgreSQL.
