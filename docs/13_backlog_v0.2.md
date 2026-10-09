# BÁO CÁO ĐẶC TẢ YÊU CẦU DỰ ÁN - PRODUCT BACKLOG v0.2

> **Lịch sử:** Phiên bản này được giữ để truy vết. Backlog hiện hành là `13_backlog_v0.3.md`.
## KẾ HOẠCH KHUNG 10 TUẦN (5 SPRINT x 2 TUẦN)

**Học phần:** Thực tập Cơ sở ngành (TTCSN)  
**Tên dự án:** Hệ thống Quản lý Phòng tập GYM (GYM Management System)  
**Nhóm thực hiện:** Nhóm 1  
**Thành viên nhóm:**
- Nguyễn Quang Đảm
- Nguyễn Thành Long
- Lê Công Quân
- Nguyễn Xuân Mạnh

---

## 1. Mục tiêu cập nhật phiên bản v0.2

Bản cập nhật **Product Backlog v0.2** được xây dựng nhằm hoàn thiện và chuẩn hóa danh mục yêu cầu sau giai đoạn phân tích Use Case tổng quát và khảo sát quy trình thực tế. Phiên bản này bao gồm các cải tiến trọng tâm sau:
1. **Bổ sung cột Trạng thái phân tích:** Đánh giá tiến độ phân tích chi tiết của từng yêu cầu chức năng (Đã phân tích / Chưa phân tích) nhằm phục vụ công tác thiết kế CSDL và kiến trúc phần mềm.
2. **Bổ sung Danh mục Yêu cầu Phi chức năng (Non-Functional Requirements - NFR):** Bổ sung 5 tiêu chuẩn kỹ thuật thiết yếu về hiệu năng, độ co giãn, an toàn bảo mật, trải nghiệm người dùng và tính toàn vẹn dữ liệu.
3. **Chuẩn hóa Kế hoạch phân bổ Sprint theo Khung 10 tuần:** Điều chỉnh lộ trình phát triển thành 5 Sprint rõ ràng (mỗi Sprint 2 tuần) tương ứng với tiến độ 10 tuần học của học phần Thực tập Cơ sở ngành.

---

## 2. Kế hoạch tổng thể Khung phát triển 10 tuần (Sprint Planning)

Lộ trình dự án được chia làm 5 Sprint (mỗi Sprint kéo dài 2 tuần) với các mục tiêu trọng tâm tương ứng:

| Sprint | Thời gian (Tuần) | Mục tiêu chính | Yêu cầu phụ trách |
| :---: | :---: | :--- | :--- |
| **Sprint 1** | **Tuần 1 - 2** | Khởi tạo dự án, khảo sát thực tế, thiết kế CSDL PostgreSQL, hoàn thiện xác thực JWT, RBAC và quản lý tài khoản. | REQ-01, REQ-02, REQ-12, NFR-03 |
| **Sprint 2** | **Tuần 3 - 4** | Xây dựng API thực thể cốt lõi: Hồ sơ hội viên (CRUD), danh mục phòng tập, khung giờ và đăng ký/xét duyệt hồ sơ PT. | REQ-03, REQ-04, REQ-05, REQ-11, REQ-16, REQ-17, NFR-04 |
| **Sprint 3** | **Tuần 5 - 6** | Nghiệp vụ cốt lõi: Động cơ đặt phòng (Transaction & Row Locking chống Race condition), đặt PT, hủy phòng, khởi tạo Frontend React. | REQ-06, REQ-07, REQ-08, REQ-13, REQ-18, NFR-01 |
| **Sprint 4** | **Tuần 7 - 8** | Giao diện tác nghiệp tại quầy & Đặt lịch: Calendar Grid, Check-in nhanh quầy lễ tân, đánh giá PT, tính điểm uy tín và hệ thống thông báo. | REQ-09, REQ-10, REQ-14, REQ-19, REQ-20, NFR-02 |
| **Sprint 5** | **Tuần 9 - 10** | Tích hợp toàn diện End-to-End, Báo cáo thống kê & Biểu đồ, xuất PDF/Excel, kiểm thử nghiệm thu, User Guide và Bảo vệ đồ án. | REQ-15, NFR-05, Hoàn thiện tài liệu & Bảo vệ |

---

## 3. Product Backlog v0.2 - Yêu cầu Chức năng (Functional Requirements)

| ID | Yêu cầu | Mô tả chi tiết | Độ ưu tiên | Tác nhân | Sprint dự kiến | Trạng thái |
| :--- | :--- | :--- | :---: | :--- | :---: | :---: |
| **REQ-01** | Đăng nhập hệ thống | Cho phép người dùng xác thực danh tính vào hệ thống bằng tên đăng nhập/email và mật khẩu. Hệ thống xác thực thông tin, cấp phiên làm việc an toàn (JWT Token) và điều hướng về trang chủ tương ứng với vai trò. Nếu thông tin không chính xác hoặc tài khoản đang bị vô hiệu hóa, hệ thống sẽ đưa ra thông báo lỗi phù hợp. | **Must** | Tất cả | Sprint 1 (Tuần 1-2) | **Đã phân tích** |
| **REQ-02** | Phân quyền theo vai trò | Hệ thống thiết lập cơ chế phân quyền chặt chẽ dựa trên 5 vai trò chính: Hội viên, PT, Lễ tân, Quản lý và Quản trị viên (Admin). Từng vai trò chỉ được phép truy cập và thực thi các chức năng thuộc phạm vi quyền hạn được cấp, chặn mọi hành vi truy cập trái phép qua URL hoặc API. Đảm bảo an toàn dữ liệu và tính độc lập giữa các bộ phận tác nghiệp. | **Must** | Quản trị viên | Sprint 1 (Tuần 1-2) | **Đã phân tích** |
| **REQ-03** | Quản lý thông tin hội viên (CRUD) | Cung cấp đầy đủ các thao tác Thêm, Xem, Sửa, Xóa/Khóa đối với hồ sơ hội viên phòng tập. Thông tin quản lý gồm họ tên, ngày sinh, giới tính, số điện thoại, email, địa chỉ, ngày tham gia và gói thẻ tập. Dữ liệu hội viên được lưu trữ tập trung, liên kết với dữ liệu đặt phòng và lịch sử ra vào câu lạc bộ. | **Must** | Lễ tân, Quản lý | Sprint 2 (Tuần 3-4) | **Đã phân tích** |
| **REQ-04** | Quản lý phòng tập | Cho phép tạo lập và cấu hình danh mục các phòng tập (phòng Gym, phòng Yoga, phòng Aerobic, phòng Boxing). Cập nhật thông tin chi tiết bao gồm tên phòng, sức chứa tối đa, loại hình tập luyện, trang thiết bị và trạng thái hoạt động (sẵn sàng, đang bảo trì, tạm dừng). Làm căn cứ dữ liệu để hệ thống phân bổ các khung giờ đặt phòng hợp lệ. | **Must** | Quản lý | Sprint 2 (Tuần 3-4) | **Đã phân tích** |
| **REQ-05** | Quản lý khung giờ | Thiết lập danh sách các khung giờ hoạt động tiêu chuẩn trong ngày tại từng phòng tập cụ thể (ví dụ: 06:00 - 08:00, 17:30 - 19:30). Cho phép tùy chỉnh số lượng học viên/hội viên tối đa được phép tham gia trong mỗi khung giờ nhằm tránh quá tải. Cung cấp chức năng bật/tắt hoặc điều chỉnh khung giờ linh hoạt theo kế hoạch vận hành. | **Must** | Quản lý | Sprint 2 (Tuần 3-4) | **Đã phân tích** |
| **REQ-06** | Đặt phòng & Đặt PT | Cho phép Hội viên tự đặt phòng trực tuyến hoặc Lễ tân tạo lịch đặt phòng hỗ trợ cho khách tại quầy. Người dùng lựa chọn phòng tập, ngày tập, khung giờ và PT (nếu muốn) để tiến hành giữ chỗ. Hệ thống lưu lại phiếu đặt với trạng thái 'Đã xác nhận' hoặc 'Chờ PT xác nhận' và tự động giảm số chỗ trống khả dụng của ca tập đó. | **Must** | Hội viên, Lễ tân | Sprint 3 (Tuần 5-6) | **Đã phân tích** |
| **REQ-07** | Hủy phòng | Cho phép Hội viên hoặc Lễ tân thực hiện hủy lịch đặt phòng trước giờ bắt đầu của ca tập theo quy chế phòng gym (ví dụ: trước 2 giờ). Khi lệnh hủy được chấp thuận, hệ thống chuyển trạng thái phiếu đặt sang 'Đã hủy'. Đồng thời, hệ thống tự động hoàn lại số lượng chỗ trống khả dụng cho khung giờ đó để người khác có thể đăng ký. | **Must** | Hội viên, Lễ tân | Sprint 3 (Tuần 5-6) | **Đã phân tích** |
| **REQ-08** | Kiểm tra trùng lịch | Cơ chế tự động kiểm tra xung đột thời gian ngay tại thời điểm người dùng gửi yêu cầu đặt phòng. Hệ thống ngăn chặn việc cùng một hội viên hoặc một PT bị trùng hai ca tập trong cùng một khoảng thời gian, hoặc đặt khi phòng đã hết sức chứa. Trường hợp phát hiện trùng lặp, hệ thống lập tức thông báo lỗi. | **Must** | Hệ thống | Sprint 3 (Tuần 5-6) | **Đã phân tích** |
| **REQ-09** | Check-in hội viên | Hỗ trợ nhân viên Lễ tân ghi nhận sự hiện diện thực tế của hội viên khi đến phòng tập thông qua mã hội viên, số điện thoại hoặc mã đặt phòng. Hệ thống kiểm tra tính hợp lệ của lịch đặt trong ngày và cập nhật trạng thái phiếu đặt thành 'Đã check-in'. Ghi nhận mốc thời gian thực tế hỗ trợ quản lý tần suất tập luyện của hội viên. | **Should** | Lễ tân | Sprint 4 (Tuần 7-8) | **Đã phân tích** |
| **REQ-10** | Báo cáo thống kê cơ bản | Cung cấp giao diện bảng điều khiển trực quan (Dashboard) cho cấp Quản lý và Admin. Hệ thống tổng hợp các chỉ số quan trọng như tổng số hội viên mới, số lượt đặt phòng theo tuần/tháng, tỷ lệ lấp đầy của từng phòng tập và số lượng check-in thực tế. Giúp ban điều hành nắm bắt nhanh hiện trạng khai thác cơ sở vật chất. | **Should** | Quản lý, Quản trị viên | Sprint 4 (Tuần 7-8) | **Đã phân tích** |
| **REQ-11** | Tìm kiếm / lọc hội viên | Trang bị bộ công cụ tìm kiếm nhanh và lọc dữ liệu đa tiêu chí cho danh sách hội viên. Người dùng có thể tìm kiếm theo họ tên, số điện thoại, mã định danh, hoặc lọc theo trạng thái thẻ tập và thời gian đăng ký. Hỗ trợ hiển thị phân trang kết quả tìm kiếm rõ ràng, giúp thao tác tra cứu tại quầy lễ tân diễn ra nhanh chóng. | **Should** | Lễ tân, Quản lý | Sprint 2 (Tuần 3-4) | **Đã phân tích** |
| **REQ-12** | Quản lý tài khoản người dùng | Cho phép Quản trị viên hệ thống (Admin) tạo tài khoản mới cho nhân viên/PT, đặt lại mật khẩu và gán quyền hạn truy cập tương ứng. Hỗ trợ thao tác khóa hoặc kích hoạt lại tài khoản của bất kỳ người dùng nào khi phát hiện bất thường hoặc khi nhân viên nghỉ việc. Lưu vết lịch sử chỉnh sửa tài khoản để phục vụ công tác giám sát an toàn. | **Must** | Quản trị viên | Sprint 1 (Tuần 1-2) | **Đã phân tích** |
| **REQ-13** | Xem lịch đặt phòng | Hiển thị giao diện lịch biểu trực quan (dạng ngày/tuần/tháng) về tình trạng các ca tập đã được đặt. Hội viên xem được danh sách lịch sử và các ca tập sắp tới của riêng mình; PT xem được lịch dạy; Lễ tân và Quản lý theo dõi được lịch trình hoạt động chung. Cung cấp bộ lọc theo phòng, theo PT và trạng thái đặt chỗ giúp theo dõi dễ dàng. | **Should** | Tất cả | Sprint 3 (Tuần 5-6) | **Đã phân tích** |
| **REQ-14** | Thông báo đặt phòng thành công | Hệ thống phát sinh thông báo tức thời (Toast/Alert) trên giao diện ngay khi giao dịch đặt hoặc hủy phòng hoàn tất thành công. Hiển thị tóm tắt thông tin lịch tập bao gồm tên phòng, ngày tập, khung giờ và lưu ý cho hội viên. Định hướng mở rộng trong tương lai có thể gửi email xác nhận tự động đến hòm thư của hội viên. | **Could** | Hội viên, Lễ tân | Sprint 4 (Tuần 7-8) | **Chưa phân tích** |
| **REQ-15** | Xuất báo cáo PDF/Excel | Cho phép Quản lý và Admin trích xuất danh sách hội viên, lịch sử đặt phòng và số liệu thống kê ra file định dạng Excel (.xlsx) hoặc tài liệu PDF. Dữ liệu xuất ra được định dạng chuẩn mực với tiêu đề, ngày xuất và chữ ký xác nhận phục vụ báo cáo định kỳ. Giúp lưu trữ tài liệu kế toán và đối soát ngoại tuyến thuận tiện. | **Could** | Quản lý, Quản trị viên | Sprint 5 (Tuần 9-10) | **Chưa phân tích** |
| **REQ-16** | Đăng ký hồ sơ PT | PT nộp hồ sơ đăng ký tài năng, chuyên môn và bằng cấp trên hệ thống để chờ duyệt. | **Must** | PT | Sprint 2 (Tuần 3-4) | **Chưa phân tích** |
| **REQ-17** | Duyệt/Cấp phép PT | Quản lý xem xét (TRUST) và cấp phép cho PT hoạt động, thiết lập cấp độ của PT. | **Must** | Quản lý | Sprint 2 (Tuần 3-4) | **Chưa phân tích** |
| **REQ-18** | Xác nhận lịch tập 1-1 | PT nhận thông báo và Xác nhận/Từ chối khi Hội viên book lịch tập kèm PT. | **Must** | PT | Sprint 3 (Tuần 5-6) | **Chưa phân tích** |
| **REQ-19** | Đánh giá PT | Hội viên đánh giá điểm sao (1-5) và nhận xét PT sau khi hoàn thành buổi tập. | **Should** | Hội viên | Sprint 4 (Tuần 7-8) | **Chưa phân tích** |
| **REQ-20** | Tính điểm uy tín PT | Hệ thống cập nhật tự động điểm uy tín rating và chuyên cần của PT. | **Should** | Hệ thống | Sprint 4 (Tuần 7-8) | **Chưa phân tích** |

---

## 4. Danh mục Yêu cầu Phi chức năng (Non-Functional Requirements - NFR)

Nhằm đảm bảo chất lượng kỹ thuật, tính ổn định và độ tin cậy trong môi trường vận hành thực tế, nhóm bổ sung 5 yêu cầu phi chức năng then chốt sau:

| ID | Phân loại | Tên yêu cầu | Tiêu chí & Mô tả chi tiết kỹ thuật | Sprint triển khai | Trạng thái |
| :--- | :--- | :--- | :--- | :---: | :---: |
| **NFR-01** | **Performance** *(Hiệu năng)* | Thời gian phản hồi < 2 giây (Golang API < 50ms) | Thời gian phản hồi của hệ thống (Response Time) đối với 95% các thao tác người dùng (truy vấn danh sách, tìm kiếm, đặt chỗ) phải đạt dưới **2.0 giây** trong điều kiện mạng tiêu chuẩn. Các truy vấn cơ sở dữ liệu nặng phải được tối ưu bằng chỉ mục (Index) và giới hạn phân trang dữ liệu. | Sprint 3 (Tuần 5-6) | **Đã phân tích** |
| **NFR-02** | **Scalability** *(Khả năng mở rộng)* | Tối thiểu 100 người dùng đồng thời | Hệ thống máy chủ ứng dụng Golang và cơ sở dữ liệu PostgreSQL phải chịu tải ổn định với tối thiểu **100 người dùng đồng thời (concurrent users)** mà không phát sinh lỗi sập hệ thống (5xx) hoặc rò rỉ bộ nhớ. Sử dụng cơ chế Connection Pooling và kiến trúc Stateless API. | Sprint 4 (Tuần 7-8) | **Đã phân tích** |
| **NFR-03** | **Security** *(Bảo mật)* | Mã hóa mật khẩu bằng bcrypt | Toàn bộ mật khẩu của người dùng tuyệt đối không lưu dạng plain text mà phải được băm một chiều bằng thuật toán **bcrypt** với Salt Rounds tối thiểu là **10**. Hệ thống sử dụng JSON Web Token (JWT) có thời hạn hiệu lực và cơ chế kiểm soát mã độc XSS/CSRF. | Sprint 1 (Tuần 1-2) | **Đã phân tích** |
| **NFR-04** | **Usability** *(Tính khả dụng)* | Giao diện tương thích (Responsive) | Giao diện người dùng được thiết kế chuẩn Responsive tương thích trên nhiều kích thước màn hình (Desktop, Tablet, Mobile) bằng Tailwind CSS. Tối ưu hóa trải nghiệm vuốt chạm trên thiết bị di động cho hội viên khi tra cứu và đặt phòng. | Sprint 2 (Tuần 3-4) | **Đã phân tích** |
| **NFR-05** | **Reliability** *(Độ tin cậy)* | Sao lưu dữ liệu tự động hằng ngày | Hệ thống thiết lập quy trình tự động sao lưu dữ liệu (Database Backup) định kỳ hằng ngày vào khung giờ thấp điểm (02:00 AM) bằng `pg_dump`. Bản sao lưu được nén và lưu trữ an toàn, đảm bảo khả năng phục hồi dữ liệu (RTO < 2 giờ, RPO < 24 giờ) khi xảy ra sự cố phần cứng. | Sprint 5 (Tuần 9-10) | **Đã phân tích** |

---

## 5. Bảng tổng kết thay đổi giữa v0.1 và v0.2 (Changelog)

1. **Bổ sung cột Trạng thái:** Toàn bộ 20 yêu cầu chức năng được cập nhật trạng thái phân tích:
   - 13 yêu cầu đã hoàn thiện phân tích chi tiết.
   - 7 yêu cầu mới/mở rộng (REQ-14 đến REQ-20) giữ trạng thái *Chưa phân tích* để chuẩn bị cho giai đoạn phát triển tiếp theo.
2. **Bổ sung Yêu cầu phi chức năng:** Tích hợp 5 yêu cầu NFR từ NFR-01 đến NFR-05 nhằm định hình các tiêu chuẩn kỹ thuật kiểm thử và thiết kế hệ thống.
3. **Chi tiết hóa kế hoạch 10 tuần:** Phân bổ cụ thể từng yêu cầu theo khung **5 Sprint x 2 tuần**, gắn liền với các mốc báo cáo định kỳ của học phần Thực tập Cơ sở ngành.
