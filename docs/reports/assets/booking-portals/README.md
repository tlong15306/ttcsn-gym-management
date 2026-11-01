# Bản phác giao diện Booking GYM

Ảnh tĩnh, dữ liệu giả lập, chưa nối API. Kế hoạch: `../../../22_ke_hoach_booking_member_pt_ui_ux.md`.

## Luồng Hội viên cập nhật: ảnh phòng trước, đặt lịch sau

- `room-catalog-desktop.png` / `.svg`: danh sách ba phòng có ảnh đại diện; 1440 × 900.
- `room-catalog-mobile.png` / `.svg`: danh sách một cột 375px; bản phác trang dài 375 × 1634, không phải bằng chứng viewport browser.
- `room-booking-desktop.png` / `.svg`: phòng đã chọn có ảnh, chọn ngày/ca/PT; 1440 × 1040.
- `room-booking-mobile.png` / `.svg`: phòng đã chọn, 375 × 940.
- `room-first-artifact-checks.json`: kiểm tra của bốn bản phác cập nhật.
- `photos/strength-demo.png`, `photos/cardio-demo.png`, `photos/pt-demo.png`: ảnh minh họa AI bằng built-in `image_gen`, không phải ảnh phòng thật; prompt cuối trong `photos/prompts.json`.

SVG cập nhật tham chiếu ảnh bằng đường dẫn tương đối `photos/`, cần giữ folder này bên cạnh SVG. PNG có ảnh ghép sẵn. File ảnh gốc được giữ nguyên; hình được đặt vào viewport của bản phác theo tỷ lệ ảnh đại diện. Khi triển khai phải dùng ảnh thật do Admin/Manager tải lên hoặc ghi rõ minh họa; không suy diễn ảnh demo thành trang thiết bị/phòng thật.

## Bản phác trước bổ sung ảnh phòng

- `member-desktop.png` / `.svg`: chọn ngày, phòng, ca và PT; 1440 × 900.
- `pt-desktop.png` / `.svg`: lịch làm việc và yêu cầu được gán; 1440 × 900.
- `member-mobile.png` / `.svg`, `pt-mobile.png` / `.svg`: bố cục riêng ở 375 × 812.
- `direction-board.png`: bảng xem chung; 1568 × 1830.
- `artifact-checks.json`: bounding box chữ và tám cặp tương phản token, không phải kết quả test app.

Phần Hội viên trong bản phác trước được giữ để đối chiếu, không còn là landing đề xuất. PT tiếp tục dùng lịch làm việc riêng.

SVG nhúng font để mở offline. Barlow Condensed Bold và Be Vietnam Pro Regular/SemiBold lấy từ [Google Fonts repository](https://github.com/google/fonts); giấy phép tương ứng trong `BarlowCondensed-OFL.txt` và `BeVietnamPro-OFL.txt`. Không cần cài font hệ thống. Khi triển khai, dùng text/component thật thay vì nhúng ảnh làm trang.

Bản phác không chứa ảnh screenshot hoặc thông tin cá nhân được người dùng gửi. Tên người, phòng, ca, chỗ trống và countdown chỉ để minh họa; không phản ánh dữ liệu local hiện có.
