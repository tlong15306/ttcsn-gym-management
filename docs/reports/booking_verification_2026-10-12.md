# Booking và portal — kiểm chứng 12/10/2026

Phạm vi mã nguồn: Booking API/options, giữ chỗ và deadline PT, idempotency, audit atomic, hủy, job expiration; room cover, seed ảnh minh họa; giao diện Hội viên/PT riêng và đặt hộ vận hành. Backend tiếp tục chia domain/dto/handler/service/repository/routes/app/job theo Gin project tham chiếu. DeepSeek Harness và Excel giữ nguyên.

## Bằng chứng đã chạy

- `go test ./... -count=1` với TEST_DATABASE_URL trỏ đúng `ttcsn_gym_test`: toàn bộ package PASS, gồm migration down/up/repeat, Identity/management/calendar và Booking. `go vet ./...` PASS. Test reset chỉ DB test, không dùng dev.
- Booking test thực hiện login đủ năm vai trò; options sessions/PT; PT không tạo booking, MEMBER không phản hồi thay PT; pending → PT nhận → MEMBER hủy; request khác expired và audit đúng một lần; trạng thái đã xử lý trả 409; PT không đọc booking tự tập không được gán.
- 12 request đồng thời cùng actor/key/payload trả cùng ID. 20 hội viên khác nhau tranh capacity=1: một thành công, các request khác CAPACITY_FULL, DB còn đúng một hold.
- Hai phòng khác nhau, snapshot thời gian chồng dù khác slot ID: MEMBER_OVERLAP và PT_UNAVAILABLE. Mốc hủy đúng 2 giờ và deadline min(30 phút,start) có test biên.
- Trigger gây lỗi audit cancellation: API 500 và booking CONFIRMED/hold giữ nguyên sau rollback. Expiration gọi lặp không thêm audit trùng.
- Upload PNG thật bằng Admin → thumbnail JPEG đọc bằng Member; Member không xóa được; Admin xóa rồi GET 404. Unit cover kiểm tra resize 1600/480, file giả/lớn và ba vai trò không có quyền ghi.
- Frontend 30 tests PASS (26 regression + 4 booking/portal/image tests), TypeScript và Vite production build PASS. Test mới kiểm tra card đúng room_id không tạo hold, review trước POST, retry giữ key khi mất response, PT navigation riêng và authenticated object URL.
- Contract generator/checker PASS: 69 operations khớp router, schema references và artifact hiện hành. `docs/check_docs.ps1` PASS: 20 REQ, 16 UC, 26 BR, 18 bảng.
- Dev migration v4 chạy thành công, không reset dữ liệu; API tại 127.0.0.1:8080 readiness HTTP 200. Seed ba phòng DEMO riêng đã chạy; HTTP thật login Admin → GET metadata → GET thumbnail của cả ba phòng đều 200 image/jpeg. Không ghi đè cover/lịch/phê duyệt đã có.

## Giới hạn kiểm chứng

Chưa chạy browser QA các viewport 375/768/1024/1440, zoom 200%, bàn phím/focus và toàn bộ tương tác UI → DB. Automatic approval review trước đó từ chối thao tác browser localhost do hạn mức; không dùng công cụ khác để vượt chặn. Ảnh phác trong assets là minh họa thiết kế, không phải screenshot ứng dụng.

Test race hiện chứng minh last-seat/same-key và guard khoảng giờ; chưa chứng minh mọi tổ hợp tranh chấp với calendar/Identity/membership/availability, accept-vs-expire/cancel hoặc backlog >100. Chưa thực thi collection trong Postman, đo NFR hoặc production deployment. Không coi build, health hoặc API integration là nghiệm thu toàn bộ MVP.

## Chi tiết vận hành

Ảnh re-encode JPEG, metadata/bytes/audit trong PostgreSQL; DB backup bao gồm ảnh. Resize stdlib hiện dùng nearest-neighbor, có thể nâng chất lượng nội suy sau khi kiểm tra ảnh thật. Idempotency creation lưu response gốc; replay có thể chứa trạng thái tại lúc tạo, GET detail cho trạng thái mới nhất. Keys có expires_at 24 giờ nhưng chưa có job dọn key; replay hiện giữ được lâu hơn 24 giờ. Expiration tối đa 100 pending/lần mỗi 15 giây, chạy ngay khi API start. Countdown dùng server_time, server quyết định deadline.

Hướng dẫn sử dụng và chuẩn bị profile/giờ mở phòng: `../23_huong_dan_booking_portal.md`. Attendance/Review/Report và chuyển COMPLETED/NO_SHOW thuộc chặng sau.
