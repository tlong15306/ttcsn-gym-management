# Quyết định và quy tắc nghiệp vụ — Baseline v0.3

Ngày chốt baseline: **10/10/2026**. Căn cứ: kế hoạch hoàn thiện tại `14_ke_hoach_hoan_thien_tai_lieu.md` đã được chấp thuận để triển khai.

## 1. Phạm vi đã chốt

**Bổ sung ngày 31/10/2026 — Khám phá công khai (U1)**

Theo yêu cầu UI/UX và mockup U0 đã được người dùng cho phép triển khai bước tiếp theo, khách chưa đăng nhập được xem trang chủ, phòng đang phục vụ/bảo trì, ảnh phòng, HLV APPROVED với tài khoản PT ACTIVE và lịch phiên công khai có số chỗ tổng hợp. Dùng projection và endpoint public riêng; không mở nguyên dữ liệu quản trị. Không công khai phòng INACTIVE, danh tính người đặt, lịch cá nhân, hồ sơ Hội viên, thông tin liên hệ tài khoản, chứng chỉ hoặc lý do xử lý nội bộ. Mọi thao tác đặt lịch và dữ liệu cá nhân tiếp tục yêu cầu xác thực/quyền hiện hành. U1 chỉ đọc lịch; kéo chỉnh lịch theo mẫu được thực hiện ở U2, bản nháp dải giờ đầy đủ ở U3.

MVP phục vụ **một cơ sở GYM** và 5 vai trò: Hội viên, PT, Lễ tân, Quản lý và Admin. Hệ thống quản lý tài khoản, hồ sơ hội viên/PT, phòng, ca, phiên phòng theo ngày, đặt chỗ, PT 1-1 tùy chọn, điểm danh, đánh giá và báo cáo vận hành cơ bản.

Không thuộc MVP: thanh toán/doanh thu, nhiều chi nhánh, lớp học nhóm như một sản phẩm riêng, thuê độc quyền phòng, ứng dụng native, cổng vật lý, SMS/email/WebSocket và cấu hình quyền động. Các mục này chỉ được bổ sung sau một quyết định phạm vi mới.

## 2. Decision log

| Mã | Quyết định được chấp thuận | Hệ quả thiết kế |
|---|---|---|
| D01 | Một booking giữ một chỗ cho một hội viên trong một phiên phòng; có thể kèm một PT 1-1. | `bookings` tham chiếu `room_sessions`, `member_id`, PT tùy chọn. |
| D02 | MVP dùng danh mục ca chung không chồng giờ và phiên phòng theo ngày. | `time_slots` bất biến khi đã dùng; `room_sessions` chốt thời gian/sức chứa. |
| D03 | PT từ chối → `REJECTED`; không phản hồi đúng hạn → `EXPIRED`; cả hai giải phóng tài nguyên. | Booking có `respond_by`; job nền hết hạn yêu cầu; không tự đổi sang booking không PT. |
| D04 | MVP mỗi tài khoản có đúng một vai trò. | Giữ `users.role_id`; đa vai trò là thay đổi schema tương lai. |
| D05 | Admin tạo tài khoản PT; PT DRAFT/REJECTED được sửa và nộp/nộp lại hồ sơ; PENDING được đăng nhập để xem nhưng không sửa hoặc nhận lịch. | Trạng thái tài khoản và trạng thái duyệt PT độc lập; thống nhất với UC-12. |
| D06 | Hội viên hủy khi `now <= starts_at - 2 giờ`; điểm danh từng người; chỉ đánh giá sau khi buổi kết thúc và cả hội viên/PT có mặt. | Có `booking_attendance`, `pt_reviews.booking_id` duy nhất, audit hủy ngoại lệ. |
| D07 | Kiểm tra tài khoản, hồ sơ và hạn thẻ tại ngày tập; ca đã dùng không sửa giờ; không giảm sức chứa dưới số chỗ đang giữ. | Vô hiệu hóa thay vì xóa/sửa lịch sử; validate trong transaction. |
| D08 | Gói tập chỉ được biểu diễn bằng thời hạn thẻ trong MVP; không báo cáo doanh thu. | Chưa có bảng payment/transaction; dashboard chỉ dùng dữ liệu vận hành. |
| D09 | PT rảnh khi đã duyệt, tài khoản hoạt động, trong giờ làm, không nghỉ và không có booking giữ lịch. Rating và tỷ lệ chuyên cần là hai chỉ số riêng. | Có lịch làm/ngoại lệ; không tạo “trust score” tổng hợp. |
| D10 | Một cơ sở; thông báo MVP là toast cho người thao tác và danh sách yêu cầu/lịch lấy lại từ server bằng polling. | Không cần message broker; trạng thái CSDL là nguồn dữ liệu. |

### Yêu cầu giao diện/ảnh phòng bổ sung — triển khai local 12/10/2026

Ngày 11/10/2026 người dùng yêu cầu Hội viên/PT dùng giao diện riêng nền sáng, chọn phòng bằng ảnh trước khi đặt lịch, và sau đó yêu cầu triển khai theo kế hoạch 22. Ngày 12/10/2026 đã có mã Booking, cover API, migration v4 và portal riêng. Chốt ảnh local BYTEA để bytes/metadata/audit atomic; thumbnail 480px, detail 1600px, input JPEG/PNG tối đa 5 MiB/4096px/16 triệu pixel. 69 API, 18 bảng nghiệp vụ. Không đổi trạng thái giữ chỗ/hạn hủy/quyền PT; không ghi phê duyệt nghiệp vụ ngoài thực tế. Kiểm chứng và giới hạn ghi tại `reports/booking_verification_2026-10-12.md`.

## 3. Quy tắc nghiệp vụ

### Yêu cầu lịch phòng bổ sung — đã triển khai demo local

Ngày 11/10/2026, người dùng yêu cầu Admin/Manager tạo phòng và đăng ký giờ hoạt động trên bảng giống Excel 7 ngày × 24 giờ, có kéo chọn/kéo thả và xác nhận; lựa chọn **lặp hằng tuần + ngoại lệ theo ngày** đã được người dùng xác nhận. Chi tiết đề xuất ở `20_ke_hoach_lich_hoat_dong_phong.md`.

Lịch mở phòng riêng bổ sung D02 và REQ-04/05, UC-10/11; ca chung vẫn quyết định độ dài phiên. Amendment ngày 11/10/2026 có 18 bảng nghiệp vụ, migration v3 và 58 API. Lịch tuần có phiên bản theo ngày hiệu lực; ngoại lệ thay cả ngày; trạng thái phòng có ưu tiên cao nhất. Preview 5 phút và xác nhận dùng revision/idempotency/transaction/audit; chặn thay đổi ảnh hưởng booking giữ chỗ, giữ phiên đã bắt đầu và đóng thủ công. Phòng legacy giữ chế độ chuyển đổi cho đến lần công bố lịch đầu tiên; phòng mới cần lịch. Hướng dẫn: `21_huong_dan_lich_phong.md`; bằng chứng: `reports/room_calendar_verification_2026-10-11.md`.

### Ghi nhận triển khai CRUD ngày 11/10/2026

Theo yêu cầu thực hiện kế hoạch 19 của người dùng, Codex triển khai vòng đời dữ liệu thay cho xóa cứng. Làm rõ D05 theo UC-12: chỉ DRAFT/REJECTED sửa/nộp, PENDING chỉ xem. Đổi phòng sang không sẵn sàng đóng các phiên OPEN tương lai khi không còn booking giữ chỗ; nếu còn giữ chỗ thì từ chối toàn bộ transaction. Mở lại phòng không tự mở lại phiên đã đóng. Ca đã dùng bất biến; sức chứa phòng chỉ áp dụng khi tạo snapshot mới. Lịch/ngoại lệ PT và audit commit cùng transaction; chứng chỉ lưu riêng tư. Đây là ghi nhận kỹ thuật theo kế hoạch được giao, không phải biên bản phỏng vấn/phê duyệt nghiệp vụ ngoài thực tế.

### Tài khoản, vai trò và hồ sơ

- **BR-01:** Mỗi tài khoản có một vai trò và một trạng thái `ACTIVE` hoặc `LOCKED`. Vai trò không thay thế quyền sở hữu/phân công trên từng bản ghi.
- **BR-02:** Khóa tài khoản chặn yêu cầu mới và làm token cũ mất hiệu lực qua `auth_version`; không xóa lịch sử nghiệp vụ.
- **BR-03:** Hội viên hợp lệ để đặt khi tài khoản `ACTIVE`, hồ sơ `ACTIVE`, và `membership_expiry >= session_date`.
- **BR-04:** PT chỉ xuất hiện trong danh sách đặt khi hồ sơ `APPROVED`, tài khoản `ACTIVE`, nằm trong lịch làm việc và không có khoảng nghỉ/xung đột.
- **BR-05:** Duyệt hồ sơ PT không tự mở khóa tài khoản; khóa tài khoản cũng không xóa quyết định duyệt.

### Phòng, ca và phiên phòng

- **BR-06:** Các ca chung trong `time_slots` không được chồng nhau; khoảng thời gian dùng quy ước nửa mở `[start_time, end_time)`.
- **BR-07:** Ca đã được tham chiếu không sửa giờ hoặc xóa; tạo ca mới và vô hiệu hóa ca cũ.
- **BR-08:** Một phòng/ngày/ca có tối đa một `room_session`. Phiên phòng chốt `starts_at`, `ends_at`, `capacity` và trạng thái.
- **BR-09:** Phòng/phiên `MAINTENANCE` hoặc `CLOSED` không nhận booking mới. Thay đổi ảnh hưởng booking tương lai phải được xử lý có lý do và audit.
- **BR-10:** Không giảm `room_sessions.capacity` xuống dưới số booking đang giữ chỗ.

### Đặt chỗ và đồng thời

- **BR-11:** Các trạng thái giữ chỗ là `PENDING_PT` và `CONFIRMED`; các trạng thái `REJECTED`, `EXPIRED`, `CANCELLED`, `COMPLETED`, `NO_SHOW` không giữ chỗ cho booking mới.
- **BR-12:** Booking không PT được tạo ở `CONFIRMED`; booking có PT được tạo ở `PENDING_PT` với `respond_by = min(created_at + 30 phút, starts_at)`.
- **BR-13:** Trong một ca, hội viên không có hơn một booking giữ lịch; PT không có hơn một booking giữ lịch.
- **BR-14:** Tổng booking giữ chỗ không vượt `capacity`. Booking khóa phòng → Hội viên → PT nếu có → advisory gate ca chung → phiên → booking; kiểm tra lại trạng thái, sức chứa và khoảng giờ thực trong transaction. Booking đọc hàng user không lấy row lock, tránh đảo thứ tự với các writer Identity/profile hiện có. Calendar giữ protocol phòng → gate → các phiên theo ID tăng dần.
- **BR-15:** Mỗi yêu cầu tạo booking có `Idempotency-Key`; cùng key và cùng payload trả cùng kết quả, khác payload trả `409 IDEMPOTENCY_CONFLICT`.
- **BR-16:** PT chỉ chấp nhận/từ chối booking đang `PENDING_PT`, được gán cho mình và chưa hết `respond_by`.
- **BR-17:** Hội viên không tự hủy sau mốc 2 giờ. Lễ tân/Quản lý/Admin được hủy ngoại lệ nhưng phải ghi lý do; PT từ chối khi pending, còn hủy sau xác nhận phải ghi lý do và có audit.
- **BR-18:** Mọi chuyển trạng thái kiểm tra trạng thái nguồn trong transaction; gọi lặp không giải phóng tài nguyên hai lần.

### Điểm danh, hoàn thành và đánh giá

- **BR-19:** Điểm danh được ghi theo từng người tham gia booking. Lễ tân/Quản lý/Admin là người ghi; mỗi người chỉ có một bản ghi điểm danh cho một booking.
- **BR-20:** Sau `ends_at`, booking thành `COMPLETED` nếu hội viên có mặt; ngược lại thành `NO_SHOW`. Sự có mặt của PT được dùng cho chuyên cần, không thay thế sự có mặt của hội viên.
- **BR-21:** Hội viên chỉ đánh giá booking `COMPLETED`, có PT, có điểm danh của cả hội viên và PT; mỗi booking có tối đa một đánh giá.
- **BR-22:** Rating PT là trung bình các đánh giá hợp lệ. Tỷ lệ chuyên cần PT là số buổi PT có mặt chia số buổi đã xác nhận và không bị cơ sở hủy; mẫu số bằng 0 hiển thị N/A.

### Dữ liệu, thời gian và audit

- **BR-23:** Thời điểm lưu bằng `TIMESTAMPTZ`; ngày vận hành và hiển thị theo `Asia/Ho_Chi_Minh`. API gửi/nhận timestamp ISO 8601 có offset.
- **BR-24:** Không hard-delete người dùng, hội viên, PT, phòng, ca, phiên phòng, booking, attendance, review và audit đã được tham chiếu.
- **BR-25:** Đặt hộ, hủy ngoại lệ, khóa tài khoản, đổi vai trò, duyệt/từ chối/đình chỉ PT và thay đổi phiên phòng phải có audit event.
- **BR-26:** Audit/log không lưu mật khẩu, token, nội dung file chứng chỉ hoặc dữ liệu bí mật khác.

## 4. Các giả định cần xác minh thực tế

Nhóm chưa cung cấp biên bản phỏng vấn một cơ sở GYM cụ thể. Mốc chờ PT 30 phút, cửa sổ điểm danh và lịch làm việc mặc định là giả định thiết kế. Trước khi triển khai production, nhóm phải xác nhận với người chịu trách nhiệm nghiệp vụ và ghi quyết định thay đổi nếu cần.


## Quyết định triển khai Attendance local — 12/10/2026

Theo kế hoạch 24 được giao triển khai: cửa sổ thường starts_at - 15 phút <= DB now < ends_at. Chỉ Admin/Manager/Lễ tân ghi, có lý do khi ghi sớm; không bổ sung sau ends_at. Replay giữ bản ghi gốc. Không hủy thường sau bất kỳ attendance nào. Job chỉ chuyển CONFIRMED đã hết giờ sang COMPLETED khi có MEMBER attendance, còn lại NO_SHOW; PT attendance độc lập. Correction và Review/report chưa triển khai trong chặng này. Idempotency Booking hiện giữ key đã commit, kể cả sau expires_at, không tự xóa; replay giữ nguyên kết quả tạo ban đầu.

## Amendment provenance Booking — 14/10/2026

Theo chặng B của kế hoạch 25, `confirmed_at` được ghi lúc tạo booking tự tập và lúc PT nhận booking có PT. Booking lịch sử không backfill từ `created_at`, `updated_at` hay ngày migration. Khi hủy, Member được ghi nguồn `MEMBER`, PT ghi `PT`, Admin/Manager/Lễ tân ghi `FACILITY`; lưu thêm role của actor ngay tại thao tác. Booking đã hủy trước migration được gắn `UNKNOWN`, role snapshot để null vì role hiện tại/lý do cũ không chứng minh được actor role lúc hủy. Không đổi công thức BR-22: hủy bởi Member/PT sau xác nhận vẫn thuộc mẫu số; hủy bởi cơ sở bị loại. Dữ liệu thiếu `confirmed_at` hoặc nguồn hủy được báo unclassified, không suy đoán.
