# BÁO CÁO ĐẶC TẢ YÊU CẦU DỰ ÁN - PRODUCT BACKLOG v0.1

> **Lịch sử:** Phiên bản này được giữ để truy vết. Backlog hiện hành là `13_backlog_v0.3.md`.

**Học phần:** Thực tập Cơ sở ngành (TTCSN)  
**Tên dự án:** Hệ thống Quản lý Phòng tập GYM (GYM Management System)  
**Nhóm thực hiện:** Nhóm 1  
**Thành viên nhóm:**
- Nguyễn Quang Đảm
- Nguyễn Thành Long
- Lê Công Quân
- Nguyễn Xuân Mạnh

---

## 1. Giới thiệu tổng quan

Product Backlog v0.1 định nghĩa danh sách các yêu cầu chức năng (Functional Requirements) ban đầu cho hệ thống Quản lý GYM. Tài liệu này đóng vai trò là cơ sở để nhóm phát triển xác định phạm vi sản phẩm khả thi tối thiểu (MVP - Minimum Viable Product), phân chia độ ưu tiên theo phương pháp MoSCoW, và lập kế hoạch phân bổ các đợt phát triển (Sprint Planning).

### Quy ước mức độ ưu tiên (MoSCoW):
- **Must (Bắt buộc):** Các tính năng cốt lõi bắt buộc phải có để hệ thống có thể vận hành tối thiểu.
- **Should (Nên có):** Các tính năng quan trọng gia tăng trải nghiệm người dùng và hiệu quả vận hành nhưng không làm tê liệt hệ thống nếu chưa có ngay.
- **Could (Có thể có):** Các tính năng tiện ích bổ sung, có thể thực hiện khi còn dư dả thời gian và nguồn lực.
- **Won't (Chưa phát triển):** Các tính năng chưa xem xét trong giai đoạn hiện tại.

---

## 2. Bảng Product Backlog v0.1 (Yêu cầu chức năng)

| ID | Yêu cầu | Mô tả chi tiết | Độ ưu tiên | Tác nhân | Sprint dự kiến |
| :--- | :--- | :--- | :---: | :--- | :---: |
| **REQ-01** | Đăng nhập hệ thống | Cho phép người dùng xác thực danh tính vào hệ thống bằng tên đăng nhập/email và mật khẩu. Hệ thống xác thực thông tin, cấp phiên làm việc an toàn (JWT Token) và điều hướng về trang chủ tương ứng với vai trò. Nếu thông tin không chính xác hoặc tài khoản đang bị vô hiệu hóa, hệ thống sẽ đưa ra thông báo lỗi phù hợp. | **Must** | Tất cả | Sprint 1 |
| **REQ-02** | Phân quyền theo vai trò | Hệ thống thiết lập cơ chế phân quyền chặt chẽ dựa trên 5 vai trò chính: Hội viên, PT, Lễ tân, Quản lý và Quản trị viên (Admin). Từng vai trò chỉ được phép truy cập và thực thi các chức năng thuộc phạm vi quyền hạn được cấp, chặn mọi hành vi truy cập trái phép qua URL hoặc API. Đảm bảo an toàn dữ liệu và tính độc lập giữa các bộ phận tác nghiệp. | **Must** | Quản trị viên | Sprint 1 |
| **REQ-03** | Quản lý thông tin hội viên (CRUD) | Cung cấp đầy đủ các thao tác Thêm, Xem, Sửa, Xóa/Khóa đối với hồ sơ hội viên phòng tập. Thông tin quản lý gồm họ tên, ngày sinh, giới tính, số điện thoại, email, địa chỉ, ngày tham gia và gói thẻ tập. Dữ liệu hội viên được lưu trữ tập trung, liên kết với dữ liệu đặt phòng và lịch sử ra vào câu lạc bộ. | **Must** | Lễ tân, Quản lý | Sprint 2 |
| **REQ-04** | Quản lý phòng tập | Cho phép tạo lập và cấu hình danh mục các phòng tập (phòng Gym, phòng Yoga, phòng Aerobic, phòng Boxing). Cập nhật thông tin chi tiết bao gồm tên phòng, sức chứa tối đa, loại hình tập luyện, trang thiết bị và trạng thái hoạt động (sẵn sàng, đang bảo trì, tạm dừng). Làm căn cứ dữ liệu để hệ thống phân bổ các khung giờ đặt phòng hợp lệ. | **Must** | Quản lý | Sprint 2 |
| **REQ-05** | Quản lý khung giờ | Thiết lập danh sách các khung giờ hoạt động tiêu chuẩn trong ngày tại từng phòng tập cụ thể (ví dụ: 06:00 - 08:00, 17:30 - 19:30). Cho phép tùy chỉnh số lượng học viên/hội viên tối đa được phép tham gia trong mỗi khung giờ nhằm tránh quá tải. Cung cấp chức năng bật/tắt hoặc điều chỉnh khung giờ linh hoạt theo kế hoạch vận hành. | **Must** | Quản lý | Sprint 2 |
| **REQ-06** | Đặt phòng | Cho phép Hội viên tự đặt phòng trực tuyến hoặc Lễ tân tạo lịch đặt phòng hỗ trợ cho khách tại quầy. Người dùng lựa chọn phòng tập, ngày tập, khung giờ và PT (nếu có) để tiến hành giữ chỗ. Hệ thống lưu lại phiếu đặt với trạng thái 'Đã xác nhận' (hoặc 'Chờ PT xác nhận') và tự động giảm số chỗ trống khả dụng của ca tập đó. | **Must** | Hội viên, Lễ tân | Sprint 3 |
| **REQ-07** | Hủy phòng | Cho phép Hội viên hoặc Lễ tân thực hiện hủy lịch đặt phòng trước giờ bắt đầu của ca tập theo quy chế phòng gym (ví dụ: trước 2 giờ). Khi lệnh hủy được chấp thuận, hệ thống chuyển trạng thái phiếu đặt sang 'Đã hủy'. Đồng thời, hệ thống tự động hoàn lại số lượng chỗ trống khả dụng cho khung giờ đó để người khác có thể đăng ký. | **Must** | Hội viên, Lễ tân | Sprint 3 |
| **REQ-08** | Kiểm tra trùng lịch | Cơ chế tự động kiểm tra xung đột thời gian ngay tại thời điểm người dùng gửi yêu cầu đặt phòng. Hệ thống ngăn chặn việc cùng một hội viên đặt trùng hai ca tập trong cùng một khoảng thời gian hoặc đặt khi phòng đã hết sức chứa. Trường hợp phát hiện trùng lặp hoặc hết chỗ, hệ thống lập tức thông báo lỗi và không tạo bản ghi đặt chỗ mới. | **Must** | Hệ thống | Sprint 3 |
| **REQ-09** | Check-in hội viên | Hỗ trợ nhân viên Lễ tân ghi nhận sự hiện diện thực tế của hội viên khi đến phòng tập thông qua mã hội viên, số điện thoại hoặc mã đặt phòng. Hệ thống kiểm tra tính hợp lệ của lịch đặt trong ngày và cập nhật trạng thái phiếu đặt thành 'Đã check-in'. Ghi nhận mốc thời gian thực tế hỗ trợ quản lý tần suất tập luyện của hội viên. | **Should** | Lễ tân | Sprint 4 |
| **REQ-10** | Báo cáo thống kê cơ bản | Cung cấp giao diện bảng điều khiển trực quan (Dashboard) cho cấp Quản lý và Admin. Hệ thống tổng hợp các chỉ số quan trọng như tổng số hội viên mới, số lượt đặt phòng theo tuần/tháng, tỷ lệ lấp đầy của từng phòng tập và số lượng check-in thực tế. Giúp ban điều hành nắm bắt nhanh hiện trạng khai thác cơ sở vật chất. | **Should** | Quản lý, Quản trị viên | Sprint 4 |
| **REQ-11** | Tìm kiếm / lọc hội viên | Trang bị bộ công cụ tìm kiếm nhanh và lọc dữ liệu đa tiêu chí cho danh sách hội viên. Người dùng có thể tìm kiếm theo họ tên, số điện thoại, mã định danh, hoặc lọc theo trạng thái thẻ tập và thời gian đăng ký. Hỗ trợ hiển thị phân trang kết quả tìm kiếm rõ ràng, giúp thao tác tra cứu tại quầy lễ tân diễn ra nhanh chóng. | **Should** | Lễ tân, Quản lý | Sprint 2 |
| **REQ-12** | Quản lý tài khoản người dùng | Cho phép Quản trị viên hệ thống (Admin) tạo tài khoản mới cho nhân viên/PT, đặt lại mật khẩu và gán quyền hạn truy cập tương ứng. Hỗ trợ thao tác khóa hoặc kích hoạt lại tài khoản của bất kỳ người dùng nào khi phát hiện bất thường hoặc khi nhân viên nghỉ việc. Lưu vết lịch sử chỉnh sửa tài khoản để phục vụ công tác giám sát an toàn. | **Must** | Quản trị viên | Sprint 1 |
| **REQ-13** | Xem lịch đặt phòng | Hiển thị giao diện lịch biểu trực quan (dạng ngày/tuần/tháng) về tình trạng các ca tập đã được đặt. Hội viên xem được danh sách lịch sử và các ca tập sắp tới của riêng mình; Lễ tân và Quản lý theo dõi được lịch trình hoạt động chung của toàn bộ phòng tập. Cung cấp bộ lọc theo phòng và theo trạng thái đặt chỗ giúp theo dõi dễ dàng. | **Should** | Tất cả | Sprint 3 |
| **REQ-14** | Thông báo đặt phòng thành công | Hệ thống phát sinh thông báo tức thời (Toast/Alert) trên giao diện ngay khi giao dịch đặt hoặc hủy phòng hoàn tất thành công. Hiển thị tóm tắt thông tin lịch tập bao gồm tên phòng, ngày tập, khung giờ và lưu ý cho hội viên. Định hướng mở rộng trong tương lai có thể gửi email xác nhận tự động đến hòm thư của hội viên. | **Could** | Hội viên, Lễ tân | Sprint 4 |
| **REQ-15** | Xuất báo cáo PDF/Excel | Cho phép Quản lý và Admin trích xuất danh sách hội viên, lịch sử đặt phòng và số liệu thống kê ra file định dạng Excel (.xlsx) hoặc tài liệu PDF. Dữ liệu xuất ra được định dạng chuẩn mực với tiêu đề, ngày xuất và chữ ký xác nhận phục vụ báo cáo định kỳ. Giúp lưu trữ tài liệu kế toán và đối soát ngoại tuyến thuận tiện. | **Could** | Quản lý, Quản trị viên | Sprint 5 |
| **REQ-16** | Đăng ký hồ sơ PT | Cho phép PT nộp hồ sơ đăng ký tài năng, chuyên môn và bằng cấp để chờ quản lý xét duyệt. | **Must** | PT | Sprint 2 |
| **REQ-17** | Duyệt/Cấp phép PT | Cung cấp công cụ để Quản lý xem xét (TRUST), đánh giá và cấp phép cho hồ sơ PT hoạt động. | **Must** | Quản lý | Sprint 2 |
| **REQ-18** | Xác nhận lịch tập 1-1 | Khi Hội viên đặt phòng có chọn PT, PT sẽ nhận được thông báo để chủ động Xác nhận hoặc Từ chối lịch tập. | **Must** | PT | Sprint 3 |
| **REQ-19** | Đánh giá PT | Cho phép Hội viên đánh giá sao và nhận xét PT sau buổi tập hoàn thành. | **Should** | Hội viên | Sprint 4 |
| **REQ-20** | Tính điểm uy tín PT | Hệ thống cập nhật tự động điểm rating trung bình và điểm chuyên cần của PT dựa trên các buổi tập. | **Should** | Hệ thống | Sprint 4 |

---

## 3. Thống kê và Phân tích sơ bộ

### 3.1. Phân bổ theo mức độ ưu tiên
- **Must Have (Bắt buộc):** 12 yêu cầu (REQ-01 đến REQ-08, REQ-12, REQ-16 đến REQ-18) - *Chiếm 60.0%*, hình thành bộ khung tính năng MVP cốt lõi.
- **Should Have (Nên có):** 6 yêu cầu (REQ-09, REQ-10, REQ-11, REQ-13, REQ-19, REQ-20) - *Chiếm 30.0%*, tối ưu hóa tiện ích tác nghiệp hằng ngày.
- **Could Have (Có thể có):** 2 yêu cầu (REQ-14, REQ-15) - *Chiếm 10.0%*, gia tăng tiện ích bổ trợ.
- **Tổng cộng:** 20 yêu cầu chức năng.

### 3.2. Phân bổ theo đợt phát triển (Sprint)
- **Sprint 1 (Khởi tạo & Quản trị nền tảng):** REQ-01, REQ-02, REQ-12 (Tập trung xác thực, phân quyền và tài khoản).
- **Sprint 2 (Quản lý thực thể cốt lõi):** REQ-03, REQ-04, REQ-05, REQ-11, REQ-16, REQ-17 (Hội viên, phòng tập, khung giờ, tra cứu, quản lý PT).
- **Sprint 3 (Nghiệp vụ cốt lõi - Đặt/Hủy phòng):** REQ-06, REQ-07, REQ-08, REQ-13, REQ-18 (Quy trình booking có PT và kiểm tra trùng lịch).
- **Sprint 4 (Tác nghiệp & Báo cáo cơ bản):** REQ-09, REQ-10, REQ-14, REQ-19, REQ-20 (Check-in, dashboard thống kê, thông báo và đánh giá PT).
- **Sprint 5 (Hoàn thiện & Mở rộng):** REQ-15 (Xuất báo cáo tài liệu, kiểm thử tích hợp toàn diện và nghiệm thu).
