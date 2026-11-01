# Kiểm chứng giờ và demo — 31/10/2026

Đây là snapshot cục bộ, không phải nghiệm thu toàn bộ [kế hoạch 32](../32_ke_hoach_xu_ly_gio_va_on_dinh_demo.md). Các test có ghi/reset dùng `ttcsn_gym_test`; dữ liệu demo lưu ở `ttcsn_gym_dev`.

## Đã thực hiện

- Chuẩn hóa hàm hiển thị giờ Việt Nam cho booking, lịch PT, điểm danh, review, phòng, lịch phòng, HLV, báo cáo và ngày của Hội viên/tài khoản. Ngày lịch thuần `YYYY-MM-DD` được định dạng riêng, timestamp được chuyển sang UTC+7 một lần. Lịch tuần PT dùng phút Việt Nam để đặt đúng vị trí buổi tập, không phụ thuộc múi giờ local của trình duyệt.
- Browser trước đó đã xác nhận màn booking hiển thị đúng 06:00–09:00 sau sửa. Các màn mới chuyển sang hàm chung đã được kiểm tra bằng test/build, chưa xác nhận từng vai trò trên browser sau thay đổi này.
- Tìm thấy tiện ích Chrome **Change Timezone for Google Chrome** đã cài trong hồ sơ Default. Mã của tiện ích chèn script vào trang và ghi đè tùy chọn `timeZone` của `Intl.DateTimeFormat`. Cơ chế này khớp kết quả đo ở tab GYM (yêu cầu Asia/Ho_Chi_Minh nhưng nhận UTC); chưa làm thử nghiệm bật/tắt để khẳng định quan hệ nhân quả. Không thay đổi thiết lập tiện ích. Kết nối cập nhật trang Vite chưa có kết luận.
- Frontend regression: **58/58 PASS**, 11 files, một worker, 81,81 giây; production build PASS, còn cảnh báo chunk >500 kB. Test lịch PT kiểm tra booking lúc 06:00 ở đúng vị trí khi `Intl.DateTimeFormat` bị ghi đè. Đã kiểm tra ranh giới ngày/năm, timestamp và ngày lịch.
- Kiểm tra role API: **15/15 assertions PASS** trên API local. Kiểm thử PostgreSQL test gồm tranh chỗ batch–batch/batch–single, lưu lịch PT đến 24:00 và vòng đời booking → PT → điểm danh → completion → review → báo cáo: PASS. Bài vòng đời điều chỉnh thời gian ca trong test DB; không phải demo chờ ca thực tế.
- Completion backlog 1.001 booking trên test DB: sau các lượt giới hạn 12 giây, số hoàn thành lần lượt 450, 987, 1.001; lượt tiếp theo không tạo audit trùng. Phân loại hoàn thành/vắng mặt khớp điểm danh. Hai lượt đầu chạm deadline: job có tiến triển và có thể cần nhiều chu kỳ; không kết luận nguyên nhân log lỗi lịch sử chỉ từ phép thử này.
- Tải cô lập qua router trong tiến trình, giới hạn `GOMAXPROCS=4`: 100 tài khoản khác nhau, 100 login + 100 đọc `/me` + 100 sửa hồ sơ, đều thành công; DB test xác nhận 100 cập nhật và 100 audit. Tổng 8,38 giây; p95 login 6,94 giây, đọc 2,28 giây, sửa 1,81 giây. Đây là burst test ngắn trên máy hiện tại, không phải thử tải bền qua mạng với dataset đại diện hoặc đo đầy đủ host/DB; chưa công bố đạt NFR-01/NFR-02. Login/read p95 của phép thử này vượt mục tiêu 2 giây.
- Backup dev `ttcsn_gym_dev_20261031_043654.dump` đã restore sang `ttcsn_gym_restore_20261031_043654`; tổng quan DB khớp, 1/1 file private được so khớp hash. Sau backup, lệnh `--refresh-pending` giữ dữ liệu cũ và tạo 6 booking PT chờ mới. Các yêu cầu hết hạn tối đa sau 30 phút; manifest `.artifact_work/demo-manifests/demo-20261031-043807.307737200.json` là snapshot, cần kiểm tra trạng thái hiện tại trước demo.
- Đã thêm [hướng dẫn demo](../33_huong_dan_demo.md), giải thích tài khoản đủ điều kiện, luồng, refresh và giới hạn thời gian.

## Còn mở

### Cập nhật kiểm tra buổi tối 31/10

- Sau khi khởi động lại môi trường, API/client trả HTTP 200; dev còn 26 users, 14 rooms và 45 booking trước refresh. Backend full regression PostgreSQL và `go vet` PASS; frontend full regression **59/59 PASS**, 11 files, 43,80 giây.
- Backup `ttcsn_gym_dev_20261031_211855.dump` restore sang `ttcsn_gym_restore_20261031_211855`, số booking khớp 45/45 và file private khớp hash 1/1. Refresh lúc 21:19 tạo 6 yêu cầu PT mới, giữ lịch sử; manifest `demo-20261031-211909.388250100.json` chỉ là snapshot có hạn phản hồi theo nghiệp vụ.
- Tải router với 100 danh tính chạy lại: p95 login 2,82 giây, đọc 1,43 giây, sửa hồ sơ 0,62 giây. CPU profile ở `.artifact_work/login-load-20261031-2119.cpu`: hàm Blowfish encryptBlock chiếm khoảng 96% mẫu CPU; không hạ bcrypt cost để đạt số đo.
- Phép thử mới qua HTTP loopback: 100 hội viên cạnh tranh session capacity 80; **80 created, 20 CAPACITY_FULL**, DB xác nhận 80 booking và 80 audit. p95 login 2,84 giây, đọc 0,35 giây, booking 1,22 giây; tổng burst khoảng 4,50 giây. Log `.artifact_work/network-booking-load-20261031-2121.log`. Đã có booking mutation qua HTTP; chưa có tải bền/dataset đại diện/đo lock-wait nên NFR vẫn mở.
- Điều khiển browser dừng vì không xác định đủ chắc URL hiện tại; chưa xác nhận thêm các màn hoặc luồng trực tiếp trong lượt tối. Không coi phần này PASS.

### Công việc còn lại

- Kiểm tra trên browser từng màn/role sau thay đổi, gồm kéo/chạm/bàn phím, lưu và đọc lại lịch phòng/HLV; chứng minh luồng Hội viên → PT → lễ tân → completion → review → báo cáo thực tế. Kiểm tra tiếp từ chối/hủy/hết hạn, mất mạng và responsive.
- Cần phép thử có kiểm soát để xác nhận tiện ích là nguyên nhân trực tiếp ở tab GYM; chưa chẩn đoán được HMR. Không tự thay đổi tiện ích của người dùng.
- NFR còn thiếu tải bền qua mạng, dữ liệu đại diện, throughput, số đo CPU/RAM và pool/lock-wait trong thời gian giữ tải; chưa chứng minh job hồi phục sau DB gián đoạn hoặc khẳng định nguyên nhân log cũ. Cần kiểm tra bảo mật/object scope/upload/log/dependency và hoàn thiện bàn giao theo kế hoạch 30.
