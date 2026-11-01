# Mockup U0 — Trang chủ và chọn lịch GYM

Mở **[index.html](index.html)** bằng Chrome hoặc Edge. Bản mẫu hoạt động bằng file cục bộ, không cần backend hay đăng nhập thật. Giữ nguyên vị trí thư mục trong repo để ảnh/font tương đối được tải đúng. Có thể mở [app.html](app.html) để xem riêng giao diện không có thanh điều khiển duyệt mẫu.

## Cách duyệt

- Chuyển **Trang chủ / Chọn lịch** trên thanh điều khiển phía trên.
- Chọn **Máy tính / Điện thoại** để thử hai bề rộng. Điện thoại mặc định vẫn xem lịch tuần đủ 7 ngày; vuốt ngang trong lịch hoặc chạm ngày để di chuyển, có chế độ xem một ngày.
- Chuyển **Khách / Hội viên mẫu / Thẻ hết hạn** và các tình huống **Lịch bình thường / Đang tải / Không có ca / Lỗi kết nối**.
- Trên màn lịch, bật **Chọn giờ**, kéo chọn một dải trong ngày; kéo thân khối để di chuyển, kéo hai tay nắm để đổi biên. Chạm ca đầu/ca cuối hoặc dùng phần chọn giờ bên dưới lịch để thay thế kéo. Ca đầy/đóng không cho chọn, dải lỗi giữ lựa chọn cũ.
- Khách bấm **Tiếp tục** sẽ thấy lời mời đăng nhập. Nút **Tiếp tục với hội viên mẫu** mô phỏng đăng nhập và giữ khoảng giờ; vẫn phải xem lại và xác nhận riêng.
- Chọn tự tập hoặc HLV mẫu, mở màn xem lại rồi xác nhận để xem kết quả mô phỏng. Không có request tạo booking hoặc tài khoản nào gửi vào dự án thật.

## Hướng thiết kế

| Thành phần | Lựa chọn |
|---|---|
| Màu | Xanh chữ `#163e34`, xanh hành động `#175540`, nền `#f7f9f7`, trắng `#ffffff`, viền `#dbe3dc`, cảnh báo `#8b5e16` |
| Chữ | Barlow Condensed cho tiêu đề thể thao; Be Vietnam Pro cho nội dung/điều khiển, dùng font sẵn trong repo |
| Trang chủ | Hero ảnh phòng tập + thông điệp ngắn, phòng có bộ lọc, HLV và hướng dẫn ba bước |
| Chọn lịch | Lịch tuần là vùng chính, tóm tắt bên phải trên desktop; trục giờ và tiêu đề ngày sticky, thanh tóm tắt cuối màn hình trên mobile |
| Tương tác | Tách Xem lịch/Chọn giờ trên cảm ứng, snap theo ca 60 phút trong dữ liệu mẫu, một dải cùng ngày; không cho chọn phút ngoài session |

Ảnh minh họa tái sử dụng từ `docs/reports/assets/booking-portals/photos`. Font và giấy phép sẵn trong `client/public/fonts`. Nội dung phòng, HLV, số chỗ và kết quả đều là dữ liệu minh họa của mockup.

## Đã kiểm tra

- Kiểm tra JavaScript và luồng bằng jsdom: lọc phòng, lịch Tuần/Ngày, giữ bản nháp qua đăng nhập mẫu, thẻ hết hạn, di chuyển/chỉnh biên, bảo toàn lựa chọn khi lỗi và xác nhận.
- Chromium headless với hồ sơ thử nghiệm riêng: kéo chuột trên desktop, kéo cảm ứng ở 390 px, truy cập đủ 7 ngày trên mobile, không tràn ngang toàn trang ở 320/390 px, đăng nhập/xem lại mẫu và điều khiển trong trang duyệt iframe.
- Đã xem ảnh render desktop và mobile, sửa bố cục nhỏ và cách phân biệt cuộn/chọn. Đây là kiểm chứng mockup, chưa phải nghiệm thu ứng dụng thật hoặc thử trên điện thoại vật lý.

Các điểm cần người dùng duyệt: hướng màu/chữ, bố cục trang chủ, mật độ lịch, cách chọn Tuần/Ngày, hai chế độ Xem/Chọn giờ và thời điểm yêu cầu đăng nhập.
