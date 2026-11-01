# Kiểm chứng lịch hoạt động phòng — 11/10/2026

Phạm vi: phần lịch phòng bổ sung sau gói CRUD quản trị; môi trường demo Windows local, PostgreSQL 17 riêng cổng 5433, API 8080, React/Vite 5173. Các test có reset/rollback chỉ dùng đúng `ttcsn_gym_test`. DB dev chỉ nâng cấp bằng migration up; database `master-golang`, tài khoản cá nhân, Harness và Excel được giữ nguyên.

## Mã nguồn và hợp đồng

Snapshot commit local: backend `dfbab72`, frontend `b346a85`; tài liệu/contract bổ sung được commit riêng. Không push remote trong lần này.

- Goose v3: bốn bảng mới, tổng 18 bảng nghiệp vụ; versions/giờ tuần, ngoại lệ/giờ ngoại lệ, revision và nguồn phiên. Migration v1/v2 giữ nguyên.
- Package theo chuẩn đã yêu cầu: `domain/facility`, `dto/v1`, `handler/v1 → service/v1 → repository`, `routes/v1`, `app`, `job`.
- Chín operations calendar, tổng 58 operations trong router/OpenAPI/Postman hiện hành; không có API Booking/Attendance/Review/Report.
- UI quản lý tuần/ngoại lệ/lịch thực tế/lịch sử; trang giờ đã công bố cho ba role chỉ đọc. Grid 168 ô có kéo chọn, di chuyển/resize, bàn phím/form, undo/redo và guard nháp.
- Hướng dẫn thao tác và chuyển đổi legacy: [21 — Lịch phòng](../21_huong_dan_lich_phong.md).

## Kiểm tra tự động đã chạy

| Kiểm tra | Kết quả và giới hạn |
|---|---|
| `go test ./... -count=1` với TEST_DATABASE_URL | PASS tất cả package; repository integration chạy PostgreSQL thật, không bị skip vì thiếu DB |
| `go vet ./...` | PASS |
| `go build -o ... ./cmd/api` | PASS; bản executable nằm trong `.artifact_work` bị Git bỏ qua |
| Migration/fixtures trên DB test | PASS; khởi tạo, chạy lặp và down/up được kiểm tra trong test; dev chỉ up v3. Không chạy lại kiểm thử seed mật khẩu trong gói lịch này |
| Interval/effective calendar | PASS: 07–21, khoảng liền kề/nghỉ giữa giờ, overlap/bước sai, 24:00 cuối, timezone và ưu tiên ngoại lệ/trạng thái phòng |
| Preview/confirm/RBAC | PASS: actor/phòng/payload/signature/TTL, revision/fingerprint ca đổi, role không quản lý bị từ chối, gate phòng mới chưa đăng ký |
| PostgreSQL nghiệp vụ | PASS: CLOSED/REPLACE/withdraw, 24:00 sinh sang ngày kế, giữ snapshot/phiên đã bắt đầu/phạm vi không liên quan, giữ phiên đóng MANUAL |
| Test bổ sung PATCH ngoại lệ | PASS `TestCalendarExceptionReplacementKeepsHistory`: giữ bản cũ WITHDRAWN, tạo ACTIVE mới, đúng giờ REPLACE; từ chối đổi ngày/rút lại bản cũ, revision không đổi khi bị từ chối |
| Transaction và idempotency | PASS: CONFIRMED/PENDING_PT giữ chỗ, EXPIRED giải phóng, hold thêm sau preview bị chặn; audit lỗi rollback; hai confirm cạnh tranh một 200/một 409; cùng key replay, gồm sau API restart |
| Job và EXTEND | PASS: chạy trực tiếp/lặp không tạo trùng, chỉ bổ sung thiếu; không tự mở phiên MANUAL |
| `npm.cmd test` | 26/26 PASS, ba file; gồm grid keyboard/move/resize, interval/form qua nửa đêm Chủ nhật, undo, preview/xác nhận/idempotency và lỗi/conflict giữ nháp |
| `npm.cmd run build` | PASS TypeScript + Vite; không coi build là kiểm chứng render trên mọi kích thước |
| Generator API `--check` | 58 operations khớp router, schema references và artifacts hiện hành |
| SwaggerParser validator | OpenAPI 3.0.3 VALID |
| `docs/check_docs.ps1` | PASS: 20 REQ, 16 UC, 26 BR, 18 bảng, không có DROP TABLE trong schema cài mới |

Log local để chẩn đoán: `.artifact_work/calendar-go.log`, `calendar-vet.log`, `calendar-exception-edit-test.log`, `calendar-client-tests.log`, `calendar-build.log`; không đưa secret/token/SQL lỗi vào báo cáo. Bản API mới đã khởi động, `/health/ready` trả `ready`.

## Luồng browser và DB thật đã kiểm chứng

Manager demo đăng nhập → tạo phòng synthetic **Gym A · Lịch demo**, capacity 20 → chuyển thẳng trang lịch → preset/đóng tuần → kéo vùng Thứ hai đến Thứ tư 07–21 thành 42 ô → kéo block sang ngày khác → undo → kéo mép cuối → undo. Những thao tác này chỉ thay nháp trước xác nhận.

Sau đó đặt Thứ hai–Thứ bảy 07–21 và Chủ nhật 08–12, tổng **88 giờ/tuần**. Preview báo **27 phiên mới**; xác nhận lưu version 1 vào PostgreSQL dev. Truy vấn DB xác nhận phòng mới ID 2, revision 1, gate true, 27 phiên CALENDAR, chuẩn bị đến 07/11/2026. Phòng legacy ID 1 còn gate false và phiên MANUAL ID 1/capacity 20 vẫn được giữ. Không suy lịch tuần của phòng cũ từ một phiên demo.

Đây là dữ liệu demo synthetic, không phải phỏng vấn/họp/phê duyệt thực tế hoặc ngày hoàn thành của sinh viên. Ảnh ở kế hoạch 20 là mockup trước triển khai; không dùng làm bằng chứng UI sản phẩm hiện tại.

## Phần chưa kiểm chứng và công việc còn lại

- Auto-review chặn thao tác browser mở tab **Ngoại lệ theo ngày** vì tài khoản chạm hạn mức. Thao tác đó chưa chạy; chưa kiểm tra trực tiếp thao tác ngoại lệ, trang chỉ đọc và responsive/touch/zoom/focus của lịch mới. API/DB tests không thay cho các bước browser này.
- Chưa chờ đồng hồ chạy job đúng 00:05; startup/direct materializer và khả năng chạy lặp đã có bằng chứng. Chưa mô phỏng mọi sự cố pool/timeout hoặc đo lock wait khi nhiều phòng.
- Chưa chạy collection bằng Postman. Generator/validator chỉ xác minh hợp đồng/artifact, không tự chứng minh toàn bộ workflow collection.
- Chưa có Booking API; các holds/race được kiểm bằng fixtures và calendar API. Race Booking-vs-publish phải bổ sung khi làm Booking, theo cùng thứ tự khóa.
- Chưa chạy load 100 người dùng, benchmark p95, restore drill, accessibility audit toàn diện hoặc production deployment. Không đánh dấu toàn MVP/nghiệm thu NFR đạt từ gói này.

Lộ trình tiếp: hoàn tất browser QA bị chặn, rồi Booking theo `18_ke_hoach_lam_viec_tiep.md`. Giữ PowerShell 5.1, không yêu cầu chạy Harness để mở ứng dụng.
