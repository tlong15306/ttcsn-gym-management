# BÁO CÁO KHẢO SÁT CHỨC NĂNG CÁC HỆ THỐNG QUẢN LÝ PHÒNG GYM THỰC TẾ

> **Trạng thái tài liệu:** Tham khảo khảo sát sản phẩm. Phạm vi MVP hiện hành được chốt tại `00_quyet_dinh_va_quy_tac_nghiep_vu.md` và SRS v0.3. Các nhận định về sản phẩm thương mại cần nguồn trước khi trích dẫn học thuật; tài liệu này không thay thế khảo sát người dùng tại một cơ sở GYM cụ thể.

---

| Thông tin đồ án | Chi tiết |
| :--- | :--- |
| **Tên đề tài** | Hệ thống quản lý GYM (GYM Management System) |
| **Môn học** | Thực tập cơ sở ngành (TTCSN) |
| **Nhóm thực hiện** | Nhóm 1 |
| **Thành viên nhóm** | 1. Nguyễn Quang Đảm<br>2. Nguyễn Thành Long<br>3. Lê Công Quân<br>4. Nguyễn Xuân Mạnh |
| **Phiên bản tài liệu** | v1.0 |
| **Ngày lập** | Tháng 10/2026 |

---

## 1. GIỚI THIỆU VÀ MỤC ĐÍCH KHẢO SÁT

### 1.1. Bối cảnh
Trong thời đại chuyển đổi số và nhu cầu chăm sóc sức khỏe thể chất ngày càng tăng, các trung tâm thể hình (phòng Gym, Fitness Club) đòi hỏi phương thức vận hành hiện đại, chính xác thay thế cho việc ghi chép sổ sách hoặc các bảng tính Excel rời rạc. Một hệ thống quản lý phòng GYM đạt chuẩn không chỉ giúp tối ưu hóa công tác điều phối phòng tập, khung giờ, hạn chế quá tải hoặc trùng lịch, mà còn giúp quản lý hồ sơ hội viên chặt chẽ, hỗ trợ bộ phận lễ tân đón tiếp và ban quản lý nắm bắt hiệu quả vận hành.

### 1.2. Mục đích khảo sát
Tài liệu này tập trung nghiên cứu, đối chiếu và đánh giá **03 hệ thống phần mềm quản lý phòng GYM/Fitness thực tế** đang được sử dụng phổ biến trên phạm vi toàn cầu:
1. **GymMaster** (Trụ sở: New Zealand) – Hệ thống chuyên sâu cho chuỗi phòng tập gym và kiểm soát ra vào 24/7.
2. **Mindbody** (Trụ sở: Hoa Kỳ) – Nền tảng SaaS quy mô lớn trong lĩnh vực thể hình, yoga và chăm sóc sức khỏe.
3. **Perfect Gym** (Trụ sở: Ba Lan / Châu Âu) – Giải pháp quản lý phòng tập thể thao cấp doanh nghiệp (Enterprise Club Management).

**Mục tiêu cụ thể:**
- Phân tích ưu nhược điểm và mô hình chức năng của từng giải pháp thương mại.
- Lập ma trận so sánh 10 tiêu chí kỹ thuật và nghiệp vụ then chốt.
- Rút ra bài học thực tiễn và định hình phạm vi sản phẩm khả thi tối thiểu (**MVP - Minimum Viable Product**) cho đồ án môn học của **Nhóm 1**.

---

## 2. TỔNG QUAN 03 HỆ THỐNG THỰC TẾ ĐƯỢC KHẢO SÁT

```
+-----------------------------------------------------------------------------------+
|                        HỆ THỐNG QUẢN LÝ PHÒNG GYM KHẢO SÁT                        |
+--------------------------+----------------------------+---------------------------+
|      1. GymMaster        |        2. Mindbody         |      3. Perfect Gym       |
|  - Xuất xứ: New Zealand  |  - Xuất xứ: Hoa Kỳ         |  - Xuất xứ: Ba Lan / EU   |
|  - Thế mạnh: Access Gate |  - Thế mạnh: Class Booking |  - Thế mạnh: Multi-branch |
|    & Vận hành Gym 24/7   |    & Marketplace đa kênh   |    & Enterprise Analytics |
+--------------------------+----------------------------+---------------------------+
```

### 2.1. Hệ thống 1: GymMaster (New Zealand)
- **Tổng quan:** GymMaster là một trong những phần mềm quản lý câu lạc bộ thể hình hàng đầu thế giới, phát triển bởi Treshna Enterprises Ltd (New Zealand). Hệ thống được tin dùng tại hơn 90 quốc gia, đặc biệt phổ biến trong mô hình phòng gym tự động 24/7.
- **Mô hình hoạt động:** Cung cấp cả dịch vụ Cloud SaaS lẫn giải pháp phần cứng kết nối (Door Access Controllers, đầu đọc RFID, máy quét vân tay, cổng tay xoay tripod).
- **Điểm mạnh:**
  - Tích hợp kiểm soát vào/ra cực mạnh: Tự động khóa/mở cổng quay dựa trên trạng thái gói tập và lịch đặt phòng của hội viên.
  - Quản lý hội viên chi tiết: Lịch sử gia hạn, công nợ, hồ sơ sức khỏe và nhật ký quét thẻ theo thời gian thực.
  - Khả năng tự động hóa cao: Gửi thông báo nhắc lịch, nhắc gia hạn qua SMS/Email tự động.
- **Điểm hạn chế:**
  - Giao diện người dùng (UI) còn mang phong cách truyền thống, nhiều trường dữ liệu phức tạp gây khó khăn cho nhân viên mới.
  - Chi phí bản quyền và chi phí thiết bị phần cứng rất cao đối với các phòng gym vừa và nhỏ.

### 2.2. Hệ thống 2: Mindbody (Hoa Kỳ)
- **Tổng quan:** Mindbody là nền tảng quản lý trực tuyến và sàn đặt lịch (Marketplace) khổng lồ dành cho ngành fitness, wellness, yoga và thẩm mỹ viện, niêm yết tại Mỹ và phục vụ hàng chục ngàn doanh nghiệp trên toàn thế giới.
- **Mô hình hoạt động:** Nền tảng Cloud tập trung vào trải nghiệm đặt lịch (Scheduling & Booking Engine) đa nền tảng kết hợp mạng lưới hội viên rộng khắp.
- **Điểm mạnh:**
  - Công cụ đặt lịch/phòng (Booking Engine) hoàn thiện bậc nhất: Phân chia lớp học, khung giờ, huấn luyện viên, kiểm tra sức chứa phòng linh hoạt.
  - Ứng dụng di động (Mobile App) dành cho hội viên có trải nghiệm xuất sắc (UX/UI tinh tế, đặt phòng trong vài lượt chạm).
  - Khả năng xử lý thanh toán đa kênh (Thẻ quốc tế, ví điện tử, thanh toán định kỳ recurring billing).
- **Điểm hạn chế:**
  - Tập trung nhiều vào mô hình lớp học (Class/Studio/Yoga) hơn là quản lý cơ sở vật chất phòng gym tạ truyền thống.
  - Hệ thống kiểm soát cửa cơ học (Turnstile/Door Access) không mạnh bằng GymMaster, thường phải thông qua tích hợp phần mềm trung gian của bên thứ ba.
  - Phí hoa hồng và phí duy trì phần mềm định kỳ đắt đỏ.

### 2.3. Hệ thống 3: Perfect Gym (Ba Lan / Toàn cầu)
- **Tổng quan:** Perfect Gym Solutions là giải pháp phần mềm cấp doanh nghiệp (Enterprise Management Platform) thiết kế cho các chuỗi phòng gym có quy mô lớn, nhiều chi nhánh (Multi-location) và hàng trăm ngàn hội viên.
- **Mô hình hoạt động:** Hệ sinh thái mở toàn diện gồm Web Portal quản trị, Kiosk tự phục vụ tại sảnh, Mobile App cho người dùng và Open API mạnh mẽ.
- **Điểm mạnh:**
  - Phân quyền (RBAC) đa cấp độ vượt trội: Phân quyền từ Admin chuỗi, Giám đốc vùng, Quản lý chi nhánh, Lễ tân cho đến Huấn luyện viên và Hội viên.
  - Báo cáo và Business Intelligence (BI): Trực quan hóa dữ liệu theo thời gian thực (lượt check-in, công suất sử dụng phòng, tỷ lệ hủy lịch, doanh thu).
  - Tự phục vụ thông minh: Kiosk cho phép hội viên tự check-in bằng khuôn mặt/mã QR, tự mua vé và tự chọn khung giờ tập.
- **Điểm hạn chế:**
  - Hệ thống quá đồ sộ, quy trình thiết lập ban đầu (Onboarding) kéo dài nhiều tháng.
  - Yêu cầu cấu hình hạ tầng cao và đội ngũ vận hành kỹ thuật chuyên trách.

---

## 3. BẢNG MA TRẬN SO SÁNH TÍNH NĂNG CHI TIẾT

### 3.1. Quy ước đánh giá
- **`V` (Có / Hỗ trợ hoàn chỉnh):** Tính năng có sẵn, hoạt động độc lập và đạt chuẩn công nghiệp.
- **`△` (Hỗ trợ một phần / Phụ thuộc / Hạn chế):** Tính năng có hỗ trợ nhưng cần cấu hình phức tạp, mua thêm module hoặc phụ thuộc đối tác thứ ba.
- **`X` (Không hỗ trợ):** Tính năng không có trong phiên bản tiêu chuẩn của phần mềm.

### 3.2. Bảng so sánh tổng hợp 10 tiêu chí

| STT | Tiêu chí đánh giá | GymMaster (NZ) | Mindbody (Mỹ) | Perfect Gym (EU) | Định hướng Nhóm 1 (MVP) |
| :-: | :--- | :---: | :---: | :---: | :---: |
| **1** | **Quản lý hội viên (Member Management)**<br>- Hồ sơ, thông tin liên lạc, mã hội viên<br>- Lịch sử tập luyện, trạng thái kích hoạt | **V** | **V** | **V** | **V**<br>*(CRUD thông tin hội viên, mã thẻ, trạng thái hoạt động)* |
| **2** | **Đặt lịch & Đặt phòng (Booking & Scheduling)**<br>- Xem danh sách phòng & khung giờ trống<br>- Đặt phòng / Hủy phòng tập<br>- Kiểm tra xung đột & chống trùng lịch | **V** | **V** | **V** | **V**<br>*(Tính năng trọng tâm: Đặt/hủy phòng, kiểm tra trùng lịch triệt để)* |
| **3** | **Điểm danh & Check-in (Check-in System)**<br>- Check-in tại quầy qua mã/ID<br>- Cổng từ RFID / Vân tay / Cổng xoay | **V**<br>*(Hỗ trợ phần cứng chuyên sâu)* | **△**<br>*(Chủ yếu qua App / Quầy)* | **V**<br>*(RFID, QR, Nhận diện khuôn mặt)* | **△**<br>*(Lễ tân thao tác check-in trên Web qua Mã/SĐT)* |
| **4** | **Thanh toán & Gói dịch vụ (Billing & Plans)**<br>- Quản lý gói tập, giá cước<br>- Cổng thanh toán online / Thẻ ngân hàng<br>- Ghi nhận hóa đơn tại quầy | **V** | **V** | **V** | **△**<br>*(Quản lý trạng thái gói cơ bản, thanh toán tiền mặt/chuyển khoản tại quầy)* |
| **5** | **Báo cáo & Thống kê (Reporting & Analytics)**<br>- Thống kê lượt đặt phòng theo ngày/tuần<br>- Thống kê khung giờ cao điểm & tỷ lệ lấp đầy<br>- Báo cáo tăng trưởng hội viên | **V** | **V** | **V**<br>*(Báo cáo BI nâng cao)* | **V**<br>*(Thống kê số lượt đặt, công suất phòng, khung giờ cao điểm)* |
| **6** | **Phân quyền truy cập (RBAC - Role Permissions)**<br>- Phân cấp quyền chặt chẽ theo vai trò<br>- Bảo mật dữ liệu cá nhân theo phân hệ | **V** | **V** | **V** | **V**<br>*(5 vai trò: Hội viên, PT, Lễ tân, Quản lý, Quản trị viên)* |
| **7** | **Quản lý phòng & Khung giờ (Facility & Slots)**<br>- Danh mục phòng tập, sức chứa tối đa<br>- Cấu hình khung giờ cố định (TimeSlots)<br>- Chuyển trạng thái bảo trì/đóng cửa | **V** | **V** | **V** | **V**<br>*(CRUD phòng tập, quản lý sức chứa và danh mục khung giờ)* |
| **8** | **Thông báo (Notification Engine)**<br>- Email / SMS nhắc lịch tập<br>- Thông báo hủy phòng / Xác nhận đặt chỗ | **V** | **V** | **V** | **△**<br>*(Thông báo trực tiếp trên giao diện Dashboard/In-app Alert)* |
| **9** | **Ứng dụng di động (Mobile Application)**<br>- App iOS / Android cho hội viên & HLV | **V** | **V** | **V** | **X**<br>*(Tập trung Web Responsive chuẩn UI Mobile trên trình duyệt)* |
| **10** | **Tích hợp phần cứng & API (Open API & Hardware)**<br>- Cổng tay xoay, camera nhận diện<br>- REST API cho bên thứ ba | **V** | **V** | **V** | **△**<br>*(Thiết kế RESTful API chuẩn mở, sẵn sàng tích hợp tương lai)* |

---

## 4. PHÂN TÍCH CHI TIẾT TỪNG TIÊU CHÍ VÀ THỰC TRẠNG

### 4.1. Tiêu chí 1: Quản lý hội viên
- **Thực tế:** Cả 3 hệ thống đều lưu trữ hồ sơ hội viên rất chi tiết: thông tin nhân khẩu học, liên hệ khẩn cấp, hình ảnh đại diện, lịch sử gói tập và tài liệu y tế. GymMaster và Perfect Gym còn lưu trữ mã định danh vật lý (mã thẻ RFID, mã barcode).
- **Đánh giá:** Đây là phân hệ nền tảng của mọi hệ thống GYM. Việc tách biệt bảng người dùng hệ thống (`Users`) và bảng thông tin hội viên (`Members`) giúp phân biệt rõ ràng giữa tài khoản xác thực và dữ liệu hồ sơ cá nhân.

### 4.2. Tiêu chí 2: Đặt lịch và quản lý khung giờ phòng tập
- **Thực tế:** 
  - Mindbody có cơ chế đặt lịch linh hoạt nhất cho các lớp theo nhóm (Group Fitness) và ca kèm riêng (1-on-1 PT). Hệ thống tự động giới hạn người tham gia theo sức chứa của phòng.
  - Perfect Gym cho phép cấu hình quy tắc hủy phòng (ví dụ: chỉ được hủy trước giờ tập 2 tiếng, nếu trễ tính là vắng mặt).
- **Đánh giá:** Đặt phòng là quy trình xảy ra xung đột dữ liệu cao nhất (Concurrency Issue). Nếu hai hội viên cùng đặt một chỗ cuối cùng trong cùng một khung giờ, hệ thống bắt buộc phải có cơ chế kiểm tra (Validation / Locking) để ngăn chặn tình trạng đặt trùng lịch (Overbooking).

### 4.3. Tiêu chí 3: Điểm danh và Check-in
- **Nhận định cần nguồn:** GymMaster được mô tả là có thế mạnh tích hợp bo mạch điều khiển cửa (Wiegand controllers). Khi hội viên quẹt thẻ, hệ thống có thể điều khiển cổng và ghi nhận bản ghi điểm danh theo cấu hình sản phẩm.
- **Đánh giá:** Đối với dự án môn học, việc đầu tư thiết bị phần cứng cửa xoay là không khả thi. Hướng tiếp cận phù hợp là: **Bộ phận lễ tân sử dụng giao diện Web để tra cứu mã hội viên/số điện thoại và bấm nút Check-in**.

### 4.4. Tiêu chí 4: Phân quyền người dùng (RBAC)
- **Thực tế:** Cả 3 phần mềm đều có ma trận phân quyền chi tiết. Perfect Gym hỗ trợ phân quyền theo cây tổ chức nhiều chi nhánh, trong khi GymMaster phân quyền theo nhóm tính năng của từng nhân viên.
- **Đánh giá cập nhật:** Hệ thống của Nhóm 1 sử dụng 5 vai trò trong baseline v0.3:
  - **Quản trị viên (Admin):** Quản lý tài khoản, phân quyền, cấu hình hệ thống.
  - **Quản lý (Manager):** Quản lý phòng tập, khung giờ, xem báo cáo thống kê.
  - **Lễ tân (Receptionist):** Quản lý thông tin hội viên, thực hiện check-in, hỗ trợ đặt phòng.
  - **Hội viên (Member):** Tự đặt/hủy phòng, tra cứu lịch tập cá nhân.
  - **PT (Personal Trainer):** Nộp hồ sơ chuyên môn, quản lý khả dụng và phản hồi lịch được gán.

### 4.5. Tiêu chí 5: Báo cáo & Thống kê
- **Thực tế:** Mindbody và Perfect Gym cung cấp dashboard đồ thị thời gian thực về: doanh thu, tỷ lệ lấp đầy phòng tập, tỷ lệ hội viên quay lại và khung giờ cao điểm trong tuần.
- **Đánh giá:** Hệ thống của Nhóm 1 cần cung cấp các biểu đồ/báo cáo thiết thực: số lượng đặt phòng theo ngày/tuần, tỷ lệ kín chỗ của từng phòng tập và các khung giờ tập được ưa chuộng nhất nhằm giúp Quản lý đưa ra quyết định điều phối.

---

## 5. BÀI HỌC KINH NGHIỆM VÀ ĐỊNH HƯỚNG CHO HỆ THỐNG CỦA NHÓM 1

Dựa trên việc nghiên cứu chuyên sâu 03 hệ thống thực tế trên, Nhóm 1 rút ra các bài học quan trọng để định hình kiến trúc và phạm vi triển khai cho dự án:

```
+---------------------------------------------------------------------------------------+
|                       BÀI HỌC KINH NGHIỆM TỪ CÁC HỆ THỐNG THỰC TẾ                     |
+---------------------------------------------------------------------------------------+
|  1. Kiến trúc phân quyền 5 vai trò rõ ràng, phân tách trách nhiệm nghiệp vụ (RBAC).   |
|  2. Cơ chế khóa & kiểm tra trùng lịch (Slot Conflict Prevention) là cốt lõi.         |
|  3. Chuẩn hóa quy trình đặt - hủy phòng có ràng buộc thời gian hợp lý.                |
|  4. Tối ưu giao diện lễ tân (Check-in nhanh, thao tác dưới 5 giây).                  |
|  5. Tập trung hoàn thiện phiên bản MVP ổn định, chất lượng cao trong 10 tuần.        |
+---------------------------------------------------------------------------------------+
```

### 5.1. Những ưu điểm cần học hỏi và kế thừa
1. **Kiểm soát tính toàn vẹn của lịch đặt phòng:** Học hỏi từ Mindbody và GymMaster về thuật toán kiểm tra xung đột khung giờ: một phòng trong một khung giờ chỉ được phép tiếp nhận số lượng đặt chỗ không vượt quá sức chứa tối đa (`Capacity`) của phòng đó; một hội viên không thể đặt 2 phòng khác nhau trong cùng một khung giờ.
2. **Quy trình hủy phòng minh bạch:** Thiết lập quy định hủy phòng rõ ràng (ví dụ: hủy trước khi khung giờ bắt đầu ít nhất 1 giờ) để giải phóng slot cho các hội viên khác, học tập cơ chế quản lý quy tắc của Perfect Gym.
3. **Phân tách vai trò Lễ tân và Quản lý:** Lễ tân cần giao diện thao tác nhanh tại quầy (tìm kiếm tức thì, 1-click check-in); trong khi Quản lý cần giao diện tổng quan dữ liệu và điều phối tài nguyên (phòng, ca tập).

### 5.2. Những điểm phức tạp cần tinh giản cho dự án sinh viên (MVP Scope)
1. **Không đầu tư phần cứng chuyên dụng:** Bỏ qua tích hợp cổng tay xoay vật lý hoặc máy quét vân tay phức tạp; thay vào đó số hóa quy trình check-in qua giao diện web của Lễ tân.
2. **Không phát triển ứng dụng di động độc lập (Native App):** Xây dựng giao diện Web trên nền tảng **React + Tailwind CSS** có tính đáp ứng cao (Responsive Design), hiển thị tốt trên cả màn hình máy tính bảng và điện thoại thông minh.
3. **Phương thức thanh toán:** Thanh toán, giao dịch và doanh thu nằm ngoài MVP v0.3. MVP chỉ quản lý thời hạn thẻ; nếu bổ sung thanh toán phải thiết kế gói, đăng ký gói, giao dịch, hoàn tiền và đối soát trước khi báo cáo doanh thu.

---

## 6. KẾT LUẬN

Việc khảo sát 03 hệ thống phần mềm phòng gym thương mại quốc tế (GymMaster, Mindbody, Perfect Gym) đã cung cấp cho Nhóm 1 cái nhìn toàn diện về bức tranh nghiệp vụ thực tế của một trung tâm thể hình hiện đại. 

Hệ thống của **Nhóm 1** sẽ không sao chép toàn bộ sự cồng kềnh của các hệ thống đa quốc gia, mà **chắt lọc chính xác các tính năng hạt nhân (MVP)**:
- Quản lý tài khoản và phân quyền 5 tác nhân.
- Quản lý danh mục phòng tập và khung giờ hoạt động linh hoạt.
- Đặt phòng, hủy phòng với cơ chế kiểm tra trùng lịch có tiêu chí kiểm thử rõ ràng.
- Quy trình tiếp đón và check-in mượt mà dành cho Lễ tân.
- Báo cáo thống kê trực quan hỗ trợ Quản lý ra quyết định.

Đây là cơ sở vững chắc để nhóm tiến hành các bước tiếp theo trong quy trình phân tích yêu cầu phần mềm và thiết kế kiến trúc hệ thống.
