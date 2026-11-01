# Kiểm tra bản phác Hội viên / PT — 11/10/2026

## Phạm vi thực hiện

- Cài skill `frontend-design` của Anthropic vào thư mục skill Codex, đọc hướng dẫn; dùng cùng hai skill UI local đã có.
- Đối chiếu routing thực tế, cấu trúc backend và quyền/Booking rules trong tài liệu hiện hành; lập kế hoạch 22 và sửa bước tiếp theo còn cũ của kế hoạch 18.
- Tạo hai bản phác desktop 1440 × 900 và hai bản phác mobile 375 × 812, PNG + SVG có font nhúng; thêm bảng xem chung và giấy phép font.
- Nội dung người/ca/phòng trong bản phác là giả lập. Không dùng thông tin cá nhân từ ảnh người dùng; không đọc/ghi database để tạo ảnh.

## Kiểm tra đã thực hiện

| Kiểm tra | Kết quả / ranh giới |
|---|---|
| Xem ảnh local | Đã xem cả bốn PNG và bảng tổng hợp; không thấy chữ bị cắt/tràn ở ảnh đã xem |
| Rà lại thiết kế | Đổi nhãn danh sách ca để không mâu thuẫn với phiên đã đầy; thêm trạng thái bằng chữ trên ô lịch PT; tăng nút nhận/từ chối và nút tiếp tục; giãn caption khỏi nội dung |
| Biên chữ trên canvas | 219 text elements trong bốn màn hình; không có bounding box ra ngoài canvas; không phải kiểm tra DOM/overlap tự động |
| Tương phản màu | Tám cặp token đã kiểm tra đạt 4.68:1–16.12:1; chi tiết trong `assets/booking-portals/artifact-checks.json`; chưa phải audit accessibility toàn ứng dụng |
| SVG/font/giấy phép | Bốn SVG có font nhúng; kèm OFL cho Barlow Condensed và Be Vietnam Pro; không cần cài font toàn máy |
| XML/kích thước artifact | PASS: parse bốn SVG, kích thước PNG khớp SVG, xác nhận font nhúng và giấy phép tồn tại |
| Checker tài liệu | PASS: 15 file bắt buộc, 20 REQ, 16 UC, 26 BR, 18 bảng; không có DROP TABLE |
| Whitespace diff | `git diff --check` PASS cho các file đã theo dõi được sửa |

Không chạy lại Go/frontend tests vì chưa sửa mã ứng dụng. Lượt này chưa commit hoặc push các tài liệu/bản phác mới; skill đã được cài ngoài repository, không đưa repo skill vào mã dự án.

## Giới hạn

Đây là ảnh/SVG thiết kế, không phải component React chạy thật. Chưa kiểm tra tương tác, keyboard/screen reader, focus/dialog, reduced motion, zoom, API hoặc responsive của hai layout trong browser. Chưa triển khai Booking, job hết hạn, transaction/idempotency hay luồng demo mới. Kiểm thử browser lịch phòng còn thiếu trong báo cáo trước vẫn còn hiệu lực; lượt này không thử lại thao tác bị giới hạn.

Không sửa app, router/OpenAPI/Postman hiện hành, schema/migration, DB, DeepSeek Harness hoặc Excel. Chưa ghi thiết kế được người dùng duyệt, không ghi nhận nghiệm thu module Booking hoặc toàn bộ MVP.

## Bổ sung cùng ngày: xem ảnh phòng trước khi đặt

Theo yêu cầu người dùng sau bản phác đầu, đã cập nhật kế hoạch 22: trang danh sách có một ảnh đại diện mỗi phòng; click ảnh/card mới vào phòng và chọn ngày/ca/PT. Không tự tạo booking hoặc hold khi xem ảnh. Phòng bảo trì vẫn xem được thông tin, không nhận booking. PT giữ layout/quyền riêng.

Đã dùng **built-in `image_gen`**, không dùng CLI, tạo và xem ba ảnh phòng giả lập: khu sức mạnh, cardio và phòng PT riêng. Ảnh gốc được copy vào workspace và giữ nguyên trong `assets/booking-portals/photos/`; prompt cuối ở `photos/prompts.json`. Ảnh được đặt vào viewport card/chi tiết khi dựng giao diện; không được giới thiệu như phòng thật hoặc chứng cứ trang thiết bị. Có nhãn “Ảnh minh họa” trên bản phác.

Đã tạo/xem bốn bản phác mới: catalog desktop 1440 × 900, catalog mobile trang dài 375 × 1634, phòng đã chọn desktop 1440 × 1040 và mobile 375 × 940. Không thấy chữ bị cắt/tràn trong ảnh đã xem. 162 text elements không có bounding box ra ngoài canvas; bốn cặp token kiểm tra đạt 5.83:1–16.12:1. Kết quả nằm ở `room-first-artifact-checks.json`; SVG cần folder ảnh tương đối đi kèm, PNG có ảnh ghép sẵn.

Metadata/ảnh chưa có trong `facility.Room` hiện hành, nên đã thêm chặng kế hoạch upload/thumbnail/quyền đọc-sửa/placeholder/migration giữ phòng cũ vào B1. Chưa sửa code backend/frontend, contract hiện hành hoặc dữ liệu.

Sau bổ sung: checker tài liệu PASS (15 file, 20 REQ, 16 UC, 26 BR, 18 bảng); `git diff --check` PASS cho diff đang theo dõi. Artifact check PASS cho bốn cặp SVG/PNG, đường dẫn ảnh, font nhúng, 162 bounding box và bốn cặp tương phản. SHA-256 của ba ảnh workspace khớp output gốc `image_gen`, xác nhận không thay file ảnh nguồn; có đủ ba prompt. Không suy diễn những kiểm tra ảnh tĩnh này thành test app.
