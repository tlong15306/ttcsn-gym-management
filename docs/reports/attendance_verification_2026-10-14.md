# Booking, Attendance và Completion — kiểm chứng 14/10/2026

Phạm vi: chốt Booking C0–C3; điểm danh riêng Hội viên/PT; trạng thái `COMPLETED`/`NO_SHOW`; lịch sử và bộ lọc điểm danh; fixture demo; cập nhật hợp đồng, hướng dẫn PowerShell và UI tìm/phân trang phòng. Backend/client không được khởi động lại trong lượt này.

## Kết quả kiểm tra

| Kiểm tra | Kết quả |
|---|---|
| `go test ./...` với `TEST_DATABASE_URL` trỏ chính xác `ttcsn_gym_test` | PASS ngày 14/10, trước chỉnh riêng ngân sách timeout của job. Repository integration suite hoàn tất trong 28 giây; test writer đồng thời và PT accept races PASS. |
| `TestBookingConcurrentEligibilityWriters` và `TestPendingAcceptanceRaces` | PASS trên PostgreSQL test: room/calendar/capacity/membership/availability writers và accept/expire/cancel race. |
| Job sau khi tách timeout | PASS: unit test xác nhận lỗi/timeout expiration không làm bỏ lượt completion; shutdown thoát được. |
| `go vet ./...` sau thay đổi job | PASS. |
| `npm.cmd test` sau tìm kiếm/phân trang phòng | PASS: 35 tests, 5 files. |
| `npm.cmd run build` | PASS TypeScript và Vite. Có cảnh báo chunk chính 503.78 kB sau minify, 158.09 kB gzip. |
| `python .\scripts\generate_api_contracts.py --check` | PASS: 71 API operations và artifacts khớp. |
| `powershell -ExecutionPolicy Bypass -File .\docs\check_docs.ps1` | PASS: 15 file bắt buộc, 20 requirements, 16 use cases, 26 business rules, 18 schema tables. |
| `git diff --check` | PASS. |

Bộ lọc phòng trang điểm danh trước đây chỉ tải trang đầu tối đa 100 phòng. UI hiện tìm theo tên/loại phòng và phân trang kết quả; có test chứng minh chọn được phòng ở trang 2.

Completion job hiện cấp timeout riêng 12 giây cho expiration và completion. Nếu lượt expiration timeout, lượt completion vẫn có ngân sách mới; nếu ứng dụng đang shutdown, context gốc dừng chu kỳ. Mỗi lượt vẫn chạy ngay khi khởi động rồi lặp theo ticker 15 giây.

## Chưa kiểm chứng

Lượt kiểm tra đầu tiên không kết nối được PostgreSQL, nên fixture lúc đó chưa được xác minh. Khi khôi phục được DB local, đã nâng dev từ Goose v4 lên v5 và chạy `go run .\cmd\seed-booking-demo` hai lần; cả hai PASS, giữ mật khẩu hiện hữu và không tự tạo booking/attendance. Live API smoke trong ngày 14/10 vẫn chưa chạy. Full Go suite sau thay đổi job đã PASS trên DB test.

Chưa có browser QA 375/768/1024/1440px, zoom 200%, keyboard/focus và thao tác UI → DB trong lượt này. Browser localhost từng bị automatic approval review chặn trong phiên trước; không dùng cách khác để vượt chặn. Build và component tests không chứng minh giao diện đã render ổn ở mọi kích thước.

Backend API và client vẫn dừng; đã xác nhận không có listener cổng 8080/5173. Không reset DB, không sửa Excel, không thay mật khẩu demo hiện hữu. Chặng A được commit sau khi lập báo cáo: `093469a`.

## Cập nhật sau kiểm chứng ban đầu — 14/10/2026

Sau khi khôi phục Docker, migration v5 được kiểm tra trên `ttcsn_gym_test` bằng full `go test ./... -count=1`; Goose fresh/repeat và test down/up/backfill legacy đều PASS. `go vet ./...` PASS. Dev chỉ được nâng bằng migration `up` từ v4 lên v5; không rollback/reset. Fixture dev chạy hai lượt và PASS. Mã provenance migration/repository thuộc chặng B và được theo dõi riêng tại kế hoạch 25.

## Hướng dẫn chạy

Lệnh khởi động DB/API/client, seed tài khoản và fixture Booking riêng nằm tại [17 — Chạy demo GYM CRUD trên PowerShell](../17_huong_dan_chay_local.md). Fixture chỉ dành cho DB `ttcsn_gym_dev` tại `127.0.0.1:5433`, dữ liệu synthetic ghi nhãn DEMO và không tạo booking/attendance thật.

Kết quả này chỉ chốt phần Booking/Attendance đã kiểm thử; không phải nghiệm thu toàn bộ MVP GYM. Review và report vẫn nằm ở các chặng sau trong [kế hoạch 25](../25_ke_hoach_chot_attendance_review_report.md).
