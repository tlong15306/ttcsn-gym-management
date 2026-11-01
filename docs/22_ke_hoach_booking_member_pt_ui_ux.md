# Kế hoạch Booking và hai giao diện Hội viên / PT

Lập ngày 11/10/2026; cập nhật triển khai 12/10/2026 theo yêu cầu người dùng. Booking API, ảnh phòng và hai layout riêng đã có mã nguồn. Kết quả kiểm chứng và phần còn thiếu nằm tại `reports/booking_verification_2026-10-12.md`; không coi ảnh phác hoặc build là nghiệm thu browser. Phạm vi: Hội viên và PT nền sáng, màu nhấn riêng; xem phòng bằng ảnh đại diện trước khi chọn lịch.

## 1. Điểm xuất phát và mục tiêu

Identity, Hội viên, Phòng, Ca/Phiên phòng, PT và lịch phòng tuần/ngoại lệ đã có mã nguồn. Hiện `client/src/App.tsx` vẫn bọc các vai trò bằng `AdminLayout`; đổi màu menu sẽ không đáp ứng yêu cầu tách trải nghiệm. Booking, Attendance, Review và Report chưa triển khai. Kết quả cũ và giới hạn browser nằm tại `reports/room_calendar_verification_2026-10-11.md`.

Chặng tiếp theo làm Booking và hai portal trên cùng React/TypeScript/Vite/Tailwind, cùng đăng nhập, API và database hiện có. Giữ cấu trúc Go theo project Gin tham chiếu, DeepSeek Harness và dữ liệu đang có. Không đổi lịch/phân công/ngày thực hiện trong Excel; không coi bản phác là sản phẩm đã chạy.

| Vai trò | Việc chính | Cách tổ chức giao diện |
|---|---|---|
| Hội viên | Xem ảnh và chọn phòng → chọn ngày, ca → chọn tập tự do hoặc PT → xác nhận; xem/hủy lịch của mình | Điều hướng trên cùng, danh sách phòng có ảnh, trang phòng với bộ chọn ngày/ca và tóm tắt đặt lịch |
| PT | Xem lịch được gán, nhận/từ chối yêu cầu, quản lý giờ rảnh/ngày nghỉ | Điều hướng bên trái gọn, lịch tuần làm trung tâm, hàng đợi yêu cầu riêng |
| Admin / Quản lý / Lễ tân | Quản lý dữ liệu và xử lý lịch theo quyền | Tiếp tục giao diện quản trị; thêm màn hình vận hành Booking đúng scope |

PT không có quyền tạo booking theo ma trận hiện hành. Tên trang của PT là **Lịch làm việc**, không dùng màn hình đặt lịch của Hội viên. Grid phòng 7 ngày × 24 giờ vẫn dành cho Admin/Manager thiết lập giờ mở cửa; không đưa grid này thành form đặt lịch của Hội viên.

## 2. Thiết kế lần 1: token và bố cục

Sáu màu nền tảng:

| Token | Giá trị | Dùng cho |
|---|---|---|
| `surface` | `#FFFFFF` | Nội dung và form |
| `canvas` | `#F5F7FA` | Nền ngoài và vùng điều hướng phụ |
| `ink` | `#182322` | Chữ chính |
| `muted` | `#566362` | Chữ phụ, nhãn giờ |
| `member-action` | `#166534` | Nút chính và trạng thái chọn của Hội viên |
| `pt-action` | `#2563EB` | Nút chính và trạng thái chọn của PT |

Màu trạng thái bổ sung có semantic token riêng: warning `#92400E` trên `#FFFBEB`, danger `#B91C1C` trên `#FEF2F2`, success `#166534` trên `#F0FDF4`; border `#DEE5E4`. Màu không thay thế nhãn trạng thái. Các cặp chữ/nền phải kiểm tra tương phản trước bàn giao.

Typography: **Barlow Condensed Bold** cho tiêu đề đặt lịch Hội viên, tạo nét thể thao; **Be Vietnam Pro** cho nội dung, số giờ và giao diện PT. H1 Hội viên 48–56px desktop / 36–40px mobile; PT 28–32px; body 15–16px, nhãn 13–14px. Font hỗ trợ dấu tiếng Việt, self-host và kèm OFL khi đưa vào app. Chữ số thời gian dùng tabular numbers. Không đặt chữ quan trọng trong ảnh.

Hội viên: bố cục căn trái, rộng tối đa khoảng 1280px. Sau bổ sung mục 3a, ảnh phòng là điểm nổi bật của trang đầu; chọn ca và tóm tắt khoảng 340px bên phải nằm trong trang phòng đã chọn. Wireframe dưới giữ làm lịch sử bố cục chọn ca ban đầu, không phải landing mới.

```text
Logo    Đặt lịch    Lịch của tôi    PT                       Hồ sơ
Đặt lịch tập.
Ngày đang chọn                         Tháng / chuyển tuần
[T2 12] [T3 13] [T4 14] ... [CN 18]
Phòng / thời điểm              | Ca đã chọn
07–08  còn chỗ                 | Phòng, ngày, giờ
08–09  đang chọn               | Tập tự do / cùng PT
09–10  đã đầy                  | Kiểm tra và xác nhận
```

PT: sidebar sáng khoảng 200px, lịch và yêu cầu căn trái; tập trung vào thời gian và việc cần phản hồi.

```text
Logo          | Lịch làm việc                     Tuần / Danh sách
Lịch          | 12–18/10/2026
Yêu cầu (2)   | Giờ | T2 | T3 | T4 | T5 | T6 | T7 | CN | Chờ phản hồi
Giờ rảnh      | 08h | buổi tập / giờ rảnh             | hạn, hội viên
Hồ sơ PT      | 09h |                                 | nhận / từ chối
Tài khoản     | ...                                   | yêu cầu tiếp
```

Nguyên tắc: một điểm nổi bật mỗi màn hình; khoảng trắng phục vụ đọc lịch, không làm màn hình trống; nút mô tả đúng hành động. Dùng icon Lucide hiện có, không thêm thư viện lịch chỉ vì hình thức. Chuyển động 100–180ms khi mở panel/chọn ca/xác nhận, tôn trọng `prefers-reduced-motion`.

## 3. Rà soát thiết kế trước khi dựng bản phác

Đã bỏ phương án hero marketing, ba thẻ KPI và sidebar tối dùng chung vì chúng không giúp chọn ca hoặc xử lý yêu cầu. Hội viên có điểm nhấn ở chữ và bộ chọn ca; PT có điểm nhấn ở lịch, không lặp lại hero Hội viên. Giảm khung viền/shadow; ngày chọn, phiên chọn và thông báo dùng hình thức theo chức năng, không chia mọi nội dung thành các card giống nhau. Không dùng tên/email/SĐT trong ảnh người dùng làm dữ liệu demo.

Tham khảo cách chọn ngày/giờ của [Cal.com](https://cal.com/) và sự nhất quán giữa điều hướng, tiêu đề, bộ chọn chế độ trong [Linear UI refresh ngày 12/03/2026](https://linear.app/changelog/2026-03-12-ui-refresh). Đây là nguồn tham khảo bố cục; không nhập tính năng, thư viện hoặc quy tắc đặt lịch của các sản phẩm này vào GYM.

Bản phác đầu: [Hội viên desktop](reports/assets/booking-portals/member-desktop.png), [PT desktop](reports/assets/booking-portals/pt-desktop.png), [Hội viên 375px](reports/assets/booking-portals/member-mobile.png), [PT 375px](reports/assets/booking-portals/pt-mobile.png), [bảng xem chung](reports/assets/booking-portals/direction-board.png). Phần Hội viên được thay bằng luồng mục 3a; PT giữ bố cục riêng như bản phác đầu. SVG tương ứng có cùng tên. **Dữ liệu giả lập, ảnh tĩnh chưa nối API; chưa chứng minh tương tác, accessibility hoặc responsive của app.**

Mobile: Hội viên chuyển thành một cột, thanh tóm tắt/nút tiếp tục có khoảng đệm để không che nội dung; PT mặc định danh sách theo ngày và tab Yêu cầu, không ép lịch 7 cột vào 375px. Thanh điều hướng dưới tối đa bốn mục. Form hồ sơ dùng cùng logic trường nhưng nằm trong layout tương ứng, chiều rộng khoảng 720–800px desktop.

## 3a. Bổ sung: xem ảnh phòng trước khi đặt lịch

Luồng đã được người dùng yêu cầu: **Danh sách phòng có ảnh → nhấn chọn phòng → xem thông tin và chọn ngày/ca trong phòng đó → chọn PT nếu cần → xem lại/xác nhận**. Nhấn vào ảnh/card chỉ điều hướng, không tạo booking hay giữ chỗ. Ngày và ca không xuất hiện trên trang danh sách đầu; người dùng đã biết mình chọn không gian nào khi bắt đầu đặt.

Thiết kế cập nhật vẫn dùng sáu màu và hai font ở mục 2. Ảnh phòng trở thành điểm nổi bật: H1 khoảng 48px desktop, không thêm ảnh hero, gradient hoặc các KPI phía trên. Card ảnh tỷ lệ 16:9, tên/loại phòng/mô tả ngắn ở bên dưới trên nền trắng; desktop ba cột, tablet hai cột, mobile một cột. Một ảnh đại diện cho mỗi phòng trong chặng này; album nhiều ảnh, video, tour 360° để sau.

```text
Logo    Phòng tập    Lịch của tôi    PT                      Hồ sơ
Chọn phòng tập.
Tìm theo tên phòng          Tất cả / Tập chung / Phòng PT
[Ảnh phòng A]          [Ảnh phòng B]          [Ảnh phòng PT]
Tên / mô tả            Tên / mô tả            Tên / mô tả
Loại / sức chứa        Loại / sức chứa        Trạng thái
Xem phòng và chọn ca   Xem phòng và chọn ca   Xem thông tin

Nhấn phòng A:
Quay lại danh sách phòng
[Ảnh phòng A]   Tên / mô tả / sức chứa / trạng thái
Chọn ngày và ca của phòng A             | Tóm tắt + PT
[T2] [T3] ... [CN]                      | Xem lại / xác nhận
```

Rà lại theo brief trước khi dựng: bỏ hàng tab Gym A/Gym B/Phòng PT ở form đặt lịch; thay bằng danh sách ảnh có điều hướng. Không hiển thị “còn N chỗ” hoặc “đang mở” trên card nếu chưa có ngày/ca hoặc lịch thực tế để tính. Sức chứa trên card ghi là mặc định, còn chỗ trên form lấy từ phiên cụ thể. Phòng MAINTENANCE/CLOSED vẫn xem thông tin/ảnh được, có nhãn trạng thái và thông báo không nhận đặt; không vô hiệu hóa toàn bộ card làm người dùng không xem được ảnh.

Bản phác cập nhật: [Danh sách phòng desktop](reports/assets/booking-portals/room-catalog-desktop.png), [danh sách phòng 375px](reports/assets/booking-portals/room-catalog-mobile.png), [phòng đã chọn và đặt ca desktop](reports/assets/booking-portals/room-booking-desktop.png), [phòng đã chọn 375px](reports/assets/booking-portals/room-booking-mobile.png). Ảnh phòng là minh họa AI, không phải phòng thật; ba ảnh và prompt lưu trong `reports/assets/booking-portals/photos/`. Bản phác ghi rõ nhãn minh họa; khi Admin/Manager tải ảnh thật thì dùng ảnh thật của đúng phòng.

Card là một link có tên truy cập được, hỗ trợ click ảnh hoặc phần thông tin và Enter; không lồng button trong link. Alt text mô tả không gian, không chỉ “ảnh 1”; thông tin quan trọng luôn có chữ ngoài ảnh. Có placeholder gọn khi chưa có ảnh/ảnh lỗi, giữ nguyên kích thước để tránh nhảy bố cục. Không thay ảnh phòng thiếu bằng ảnh của phòng khác. Ảnh đầu màn hình ưu tiên tải, ảnh phía dưới lazy-load; có kích thước thumbnail, skeleton và giới hạn mô tả.

Trang phòng giữ ảnh đã chọn, tên và breadcrumb/quay lại danh sách, bên dưới là lịch mở cửa thực tế và ca có thể đặt. Đường dẫn có room_id để mở trực tiếp/refresh xác định đúng phòng; quay lại giữ tìm kiếm/lọc. Chọn phòng khác phải xóa lựa chọn phiên/PT và nội dung checkout cũ; phòng từ room_session_id được backend xác định lại khi tạo booking, không tin room_id gửi kèm để đặt nhầm phòng. PT có thể xem ảnh phòng trong chi tiết buổi được gán; vẫn dùng lịch làm việc riêng, không mở quyền tạo booking.

### Phần quản trị ảnh phòng cần bổ sung

Tại thời điểm lập kế hoạch chưa có ảnh phòng; triển khai 12/10/2026 đã bổ sung cover metadata vào `facility.Room`. Admin và Manager (quyền quản lý phòng hiện hành) tải/thay/xóa **một ảnh đại diện** tại chi tiết phòng, xem preview và nhập mô tả ảnh; các vai trò có quyền đọc phòng xem ảnh. Lễ tân/Hội viên/PT không sửa ảnh.

Chốt kỹ thuật local: migration v4 lưu JPEG detail và thumbnail bằng BYTEA trên `rooms`, cùng alt text, version và cờ minh họa. Bytes/metadata/audit commit trong cùng transaction; không cần protocol dọn file mồ côi. DB backup bao gồm ảnh nên dung lượng tăng; chứng chỉ PT giữ cơ chế riêng. API phòng chỉ trả metadata, không trả bytes hoặc filesystem path. Phòng cũ thiếu ảnh dùng placeholder. Seed ảnh là lệnh riêng, không ghi đè cover đã có.

API đã triển khai dưới `/api/v1`: POST/DELETE/GET `/rooms/{id}/cover-image`, GET nhận `size=card|detail`. GET phòng có cover metadata. Route ảnh dùng Bearer và object URL được revoke khi đổi ảnh/unmount; không đưa token vào query string. Lookup chỉ theo room_id, không nhận đường dẫn tùy ý.

Nhận JPEG/PNG tối đa 5 MiB, mỗi chiều tối đa 4096 và tổng tối đa 16 triệu pixel; kiểm tra nội dung decode, re-encode JPEG bỏ metadata, detail tối đa 1600px và thumbnail 480px. UI preview trước gửi; backend trả lỗi có request_id. Thay ảnh tăng version riêng, không đổi lịch hoặc booking; audit không chứa bytes.

## 4. Routing, package và ranh giới tái sử dụng

| Khu vực | Route đề xuất | Quyền |
|---|---|---|
| Hội viên | `/member/rooms`, `/member/rooms/:roomId`, `/member/bookings`, `/member/bookings/:id`, `/member/trainers`, `/member/profile` | MEMBER; danh sách/chi tiết phòng đọc theo quyền, booking SELF |
| PT | `/pt/schedule`, `/pt/requests`, `/pt/bookings/:id`, `/pt/availability` | PT, dữ liệu ASSIGNED/SELF |
| PT hiện có | `/pt/profile`, thêm `/pt/account` | Hồ sơ nghề nghiệp / thông tin tài khoản riêng |
| Vận hành | `/management/bookings` | Admin, Manager, Reception theo quyền từng hành động |

Tạo `MemberLayout` và `CoachLayout`, giữ `AdminLayout` cho vận hành. Dùng chung auth/session, API transport, button/dialog/form primitives, formatter thời gian và component hồ sơ; tách navigation, trang và thành phần nghiệp vụ. Các token được scope theo layout, tránh CSS của một vai trò ghi đè vai trò khác.

Sau đăng nhập: MEMBER tới `/member/rooms`, PT tới lịch làm việc, ADMIN tới tài khoản, MANAGER tới phòng, RECEPTION tới hội viên. `/`, `/member/book` cũ và đường dẫn `/me` chuyển theo vai trò; đường dẫn lịch phòng cũ chuyển tới trang đọc phù hợp. Link trực tiếp sai vai trò hiển thị từ chối truy cập; backend vẫn kiểm tra quyền từng bản ghi. PT chưa có hồ sơ hoặc chưa được duyệt có hướng dẫn hoàn thiện hồ sơ, không hiển thị nhận lịch như đã đủ điều kiện.

Backend bổ sung theo tổ chức hiện có:

```text
internal/domain/booking/
internal/dto/v1/booking.go
internal/handler/v1/booking_handler.go
internal/service/v1/booking_service.go
internal/repository/booking_repository.go
internal/job/booking_expiration.go
internal/routes/v1/   và   internal/app/
```

Luồng `handler/v1 → service/v1 → repository`; domain/service không phụ thuộc Gin. Domain là model nghiệp vụ, DTO là hợp đồng vào/ra; không tạo thêm `internal/model` trùng định nghĩa chỉ để đổi tên package.

## 5. Hợp đồng Booking cần triển khai

Các endpoint dưới `/api/v1`, đều là **dự kiến**, chưa thêm vào snapshot 58 operations đã triển khai:

| Endpoint | Mục đích và điều kiện |
|---|---|
| `GET /booking-options/sessions` | Trang phòng đã chọn gửi room_id + ngày; phiên tương lai đúng phòng, giờ bắt đầu/kết thúc, chỗ còn lại và lý do không đặt được; phân trang |
| `GET /booking-options/pts?room_session_id=...` | PT APPROVED/ACTIVE rảnh toàn bộ khoảng giờ của phiên, không chỉ một thời điểm |
| `POST /bookings` | MEMBER tạo cho mình, bắt buộc `Idempotency-Key` |
| `POST /members/{id}/bookings` | Lễ tân/Admin đặt hộ, người tạo lấy từ token |
| `GET /bookings` | Bộ lọc from/to/status/room/pt, backend áp SELF/ASSIGNED/ALL |
| `GET /bookings/{id}` | Chi tiết theo object scope, không trả dữ liệu riêng tư dư thừa |
| `POST /bookings/{id}/pt-response` | PT được gán nhận/từ chối yêu cầu chưa hết hạn |
| `POST /bookings/{id}/cancel` | Hủy đúng mốc/quyền, lý do/audit theo baseline |

Chuẩn envelope/error và `request_id` hiện có; page mặc định 20, page_size tối đa 100. Bộ lọc và sort ổn định, options không được diễn giải thành giữ chỗ: phải kiểm tra lại trong transaction khi xác nhận. Account chưa đủ điều kiện vẫn xem phòng/ảnh theo quyền đọc, nhưng không tạo được booking. Danh sách PT công khai không trả chứng chỉ riêng tư, hash/token hoặc danh sách khách của PT. Account MEMBER chưa có hồ sơ/membership hợp lệ phải được hướng dẫn liên hệ lễ tân; account PT không đồng nghĩa PT đã APPROVED.

Tuân theo `06_ma_tran_phan_quyen.md`, BR-03…BR-17 và UC-02/03/06/13:

- Không PT → CONFIRMED; có PT → PENDING_PT. Hai trạng thái đều giữ chỗ và giữ lịch. Hạn phản hồi `min(created_at + 30 phút, starts_at)`.
- Từ chối/hết hạn → giải phóng chỗ và lịch; không tự chuyển sang tập tự do. PT chỉ nhận/từ chối yêu cầu của mình, trước hạn.
- Hội viên tự hủy khi `now <= starts_at - 2 giờ`, kể cả đúng biên 2 giờ. PT hủy sau xác nhận và vận hành hủy ngoại lệ phải có lý do/audit.
- Chỉ phiên OPEN, phòng sẵn sàng, membership hợp lệ tại ngày tập mới cho đặt. PT phải đủ điều kiện, availability phủ cả phiên, không nằm trong thời gian nghỉ hoặc booking giữ lịch.
- Không trùng khoảng thời gian thực của Hội viên/PT, kể cả khác phòng hoặc khác time_slot_id. Chỉ kiểm tra UNIQUE theo ID ca là chưa đủ với phiên legacy và các ca khác nhau.
- Booking, audit, giữ chỗ và idempotency commit/rollback cùng transaction. Cùng key/cùng payload trả kết quả cũ; khác payload trả conflict. UI giữ key khi retry kết quả chưa rõ, không sinh key mới sau timeout.

Giờ hiển thị theo Việt Nam (`Asia/Ho_Chi_Minh`), không theo timezone thiết bị. Countdown dùng `respond_by` và giờ server; dữ liệu giả lập trong bản phác không được hard-code vào app.

## 6. An toàn dữ liệu và job hết hạn

Cổng bắt buộc trước khi viết Booking: rà soát thứ tự khóa của calendar, room/session, member, Identity và PT; cập nhật decision log/BR-14 nếu bổ sung khóa Hội viên. Protocol hiện hành yêu cầu phòng → PT nếu có → phiên theo ID tăng dần. Cần thêm cơ chế tuần tự hóa theo Hội viên để ngăn đặt trùng giờ ở hai phòng; không tự chèn khóa hàng user/member ở một vị trí tùy ý vì các writer hiện có có thể lấy khóa theo chiều ngược lại.

Triển khai hiện dùng phòng → Hội viên → PT (nếu có) → advisory gate ca chung → phiên → booking, đọc user không khóa hàng; cập nhật BR-14 theo mã nguồn. Đã thử last-seat/same-key và guard trùng giờ cross-room/slot; chưa chứng minh mọi tổ hợp writer/race, xem báo cáo kiểm chứng thay vì suy diễn tất cả trường hợp đã PASS.

Tạo booking phải dùng chung khóa phòng/gate liên quan với thay đổi lịch phòng, kiểm tra lại phiên, interval overlap và capacity dưới khóa. Hủy/nhận/từ chối/expire cũng phải tuân protocol thống nhất. Job expire lấy danh sách ứng viên rồi xử lý từng transaction theo protocol, đọc lại trạng thái và deadline; không khóa booking trước rồi quay lại khóa phòng. Mọi lần chạy lặp và restart phải an toàn, job bắt kịp yêu cầu đã quá hạn khi ứng dụng dừng.

UI làm mới khoảng 15 giây khi đang xem màn hình có yêu cầu pending, đồng thời invalidate sau mutation. Dừng polling khi tab ẩn/unmount/401; không gửi request mỗi giây cho countdown. Chưa thêm WebSocket vào phạm vi này. Timer trên client không tự quyết định hết hạn; backend là nguồn trạng thái.

## 7. Chia chặng và cổng nghiệm thu

Mỗi chặng: triển khai → test → phân tích lỗi → sửa → test lại, bàn giao một commit theo chức năng khi đạt kiểm tra liên quan. Không ấn định ngày hoàn thành thực tế trước khi làm.

| Chặng | Thực hiện | Cổng nghiệm thu |
|---|---|---|
| B0 — Chốt đầu vào | Rà contract/quyền, protocol khóa và lưu ảnh; kiểm tra các luồng lịch phòng còn thiếu khi browser cho phép; đối chiếu bản phác với yêu cầu | Có danh sách case còn thiếu, protocol nhất quán, scope Booking/ảnh phòng; không coi UI QA bị giới hạn là đã PASS |
| B1 — Tách trải nghiệm và ảnh phòng | Token/font, MemberLayout/CoachLayout, landing/redirect/guard; cover metadata/upload/thumbnail, quản trị ảnh; catalog phòng và hồ sơ đúng layout | 5 vai trò tới đúng trang; card ảnh/placeholder, quyền đọc/sửa ảnh, upload/thay/gỡ/lỗi và migration phòng cũ PASS; frontend tests/build và regression CRUD PASS |
| B2 — Booking thật | Options, domain/service/repository, transaction/idempotency/audit, cancel/PT response, expiration job | PostgreSQL/API tests và race tests PASS, dữ liệu cuối đúng, không vượt chỗ/trùng giờ; cập nhật OpenAPI/Postman theo router thật |
| B3 — Hội viên | Nhấn card phòng → chi tiết có ảnh → chọn ngày/ca đúng phòng → PT → review/xác nhận; lịch của tôi/chi tiết/hủy | UI ghi DB thật; không đặt nhầm phòng; chọn phòng khác xóa checkout cũ; chỗ vừa hết/PT vừa bận có lỗi phục hồi; retry không nhân đôi; đúng điều kiện membership và mốc hủy |
| B4 — PT | Lịch tuần/danh sách, pending queue, nhận/từ chối, hủy có lý do; đưa giờ rảnh/ngày nghỉ hiện có vào layout | Chỉ dữ liệu ASSIGNED/SELF; race accept/expire/cancel đúng; mobile thao tác được không kéo lịch; không có nút tạo booking |
| B5 — Vận hành và bàn giao | Đặt hộ Lễ tân/Admin, xem/hủy đúng quyền; demo nhiều vai trò, runbook và kiểm chứng | Luồng đặt → PT nhận/từ chối/hết hạn → hủy chạy thật; API/object scope/race và browser responsive PASS hoặc ghi rõ phần chưa chạy |

Attendance, check-in, completed/no-show, đánh giá và báo cáo là chặng sau Booking. Có thể hiển thị trạng thái từ API nhưng không tạo nút điểm danh/đánh giá hoặc số rating giả để lấp giao diện. Không đưa đổi lịch bằng kéo thả vào chặng PT này; đổi lịch cần quy tắc nghiệp vụ riêng.

## 8. UX và kiểm thử bàn giao

Hội viên: luồng xem ảnh/chọn phòng → chọn ngày/ca/PT → xem lại → xác nhận; sau thành công chuyển chi tiết booking và hiển thị “Đã xác nhận” hoặc “Chờ PT phản hồi đến …”. Nút tạo phải phân biệt tập tự do với gửi yêu cầu PT. PT nhận ngay từ queue nhưng từ chối/hủy có xác nhận và lý do theo nghiệp vụ. Sau thành công bỏ đúng yêu cầu khỏi queue, cập nhật lịch và focus; lỗi không làm mất nội dung đang nhập.

Trạng thái cần thiết: loading/skeleton, chưa có phiên, hết chỗ, chưa có PT phù hợp, lỗi mạng + thử lại, phiên vừa bị đóng, membership hết hạn, hồ sơ chưa đủ, yêu cầu đã hết hạn hoặc đã được xử lý nơi khác. Không chỉ disable mà bỏ lý do; trường lỗi gắn đúng input, thông báo thành công/lỗi qua live region, không chỉ toast biến mất nhanh. Khi 401 xóa token memory và về đăng nhập; tải lại trang vẫn đăng nhập lại như hiện có.

- Go unit/vet: chuyển trạng thái, mốc giờ, options và quyền; cùng contract lỗi hiện hành.
- PostgreSQL/API: 5 vai trò, SELF/ASSIGNED/ALL, đoán booking_id, membership/account/PT không hợp lệ, profile/account chưa liên kết, options phủ cả interval và timezone biên ngày.
- Race: 20 yêu cầu vào chỗ cuối chỉ một hold; cùng Hội viên/PT ở hai phòng/ca khác ID; create đối đầu calendar/exception/room status/capacity; accept/cancel/expire; audit/idempotency rollback và retry sau restart.
- Frontend: form, chọn/rà/xác nhận/hủy, PT queue, guards/redirect, API error/timeout/401, key retry, polling dừng đúng lúc và countdown theo server; giữ các test CRUD/lịch phòng cũ.
- Ảnh phòng: Admin/Manager sửa được; Lễ tân/MEMBER/PT bị chặn ghi; đúng ảnh theo room_id; MIME giả, file lớn/kích thước bất thường và tên đường dẫn bị từ chối; upload/metadata/audit rollback; thumbnail, cache sau thay ảnh, thiếu/lỗi ảnh, alt text và object URL cleanup. Click card không tạo hold; room maintenance xem được nhưng không đặt được; trực tiếp URL không bypass guard/điều kiện phiên.
- Browser: 375/768/1024/1440px, zoom 200%, bàn phím, focus sau dialog, vùng chạm khoảng 44–48px, giảm chuyển động; không cuộn ngang cả trang. PT lịch tuần chỉ dùng khi đủ chiều rộng.
- Demo DB thật: Hội viên đặt có PT → PT nhận → Hội viên xem CONFIRMED; request khác từ chối/hết hạn giải phóng; tập tự do CONFIRMED; hai người tranh chỗ cuối; Admin/Lễ tân đặt hộ và hủy có lý do. Ghi cả người thao tác và audit, không dùng ảnh tĩnh làm chứng cứ.

Hướng dẫn vùng chạm trong sản phẩm là 44–48px; [WCAG 2.2 Target Size Minimum](https://www.w3.org/WAI/WCAG22/Understanding/target-size-minimum.html) quy định tối thiểu 24 × 24 CSS px kèm ngoại lệ, không diễn giải 44px thành yêu cầu tối thiểu của tiêu chí này.

## 9. Skill và bằng chứng của lượt lập kế hoạch

Đã cài `frontend-design` từ [repo Anthropic](https://github.com/anthropics/skills/blob/main/skills/frontend-design/SKILL.md) vào `C:/Users/choqu/.codex/skills/frontend-design`, đọc hướng dẫn và dùng quy trình token → wireframe → rà lại → bản phác. Có dùng thêm skill local `ui-ux-pro-max` và `ui-styling` cho responsive/form/trạng thái; không lấy nguyên các gợi ý landing page không phù hợp từ kết quả tìm kiếm skill.

Lượt này chỉ tạo kế hoạch, ảnh/SVG thiết kế và cập nhật liên kết tài liệu. Bản phác là định hướng cần triển khai bằng component có ngữ nghĩa, không phải ảnh để nhúng thay giao diện. Báo cáo kiểm tra artifact: `reports/booking_portals_design_2026-10-11.md`.
