# Kế hoạch tiếp theo: hoàn thiện Booking → điểm danh → đánh giá → báo cáo

Ngày lập: **12/10/2026**. Đây là kế hoạch gốc trước lượt triển khai C0–C3. C0–C3 hiện đã có code chưa commit, còn kiểm chứng cuối; C4–C5 chưa triển khai. Kế hoạch chi tiết và trạng thái đối chiếu ngày 14/10 nằm tại [25 — Chốt điểm danh, Review và Report](25_ke_hoach_chot_attendance_review_report.md). Không thay mốc, người thực hiện hoặc ngày thực tế trong Excel.

## 1. Điểm xuất phát ngày 12/10, trước lượt triển khai

- HEAD `4e60eda`: Identity, CRUD nền, lịch phòng tuần/ngoại lệ, cover ảnh, Booking và portal riêng Hội viên/PT đã có code. API hiện hành có 69 operations; migration v4, 18 bảng nghiệp vụ.
- Báo cáo `reports/booking_verification_2026-10-12.md` ghi Go/DB/vet, 30 frontend tests/build và checker PASS tại lần kiểm tra trước. Lập kế hoạch lần này không chạy lại hoặc coi đây là kết quả mới.
- Chưa có API/UI attendance, completion job, review hoặc reports. Bảng `booking_attendance` và `pt_reviews` đã tồn tại; không tạo lại hoặc reset DB dev.
- Còn thiếu browser QA, một số race giữa các writer, backlog expiration >100, chính sách dọn idempotency. Demo accounts chưa tự có membership/profile PT hợp lệ; ảnh DEMO không đồng nghĩa đã có lịch mở phòng.
- Giữ React/TypeScript/Vite/Tailwind, font/ảnh và hai portal hiện có; giữ cấu trúc Go theo project Gin tham chiếu. Giữ DeepSeek Harness, DB riêng 5433 và các file người dùng chưa theo dõi.

Nguồn nghiệp vụ: `00_quyet_dinh_va_quy_tac_nghiep_vu.md` BR-17…22, `06_ma_tran_phan_quyen.md`, SRS REQ-09/10/19/20, UC-04/05/14 và backlog v0.3. Mọi quy tắc mới bên dưới là đề xuất local, cần chốt trong decision log trước khi sửa code/schema/contract.

## 2. Các chặng

| Chặng | Công việc | Cổng hoàn thành |
|---|---|---|
| C0 — Booking dễ test và ổn định | Fixtures hợp lệ, rà UI/lịch sử, xử lý lỗi, mở rộng race/job tests | Luồng 5 vai trò chạy DB thật; các test mới PASS; browser kiểm được thì có bằng chứng, bị giới hạn thì ghi riêng |
| C1 — Attendance backend | Ghi/đọc điểm danh từng MEMBER/PT, quyền, cửa sổ giờ, audit/transaction | Đúng người/booking; ghi lặp không nhân đôi; rollback và quyền PASS |
| C2 — Kết thúc buổi | Job CONFIRMED → COMPLETED/NO_SHOW, startup catch-up, xử lý tranh chấp | Dùng attendance MEMBER; chạy lặp/restart và race với hủy/check-in PASS |
| C3 — UI điểm danh và kết quả | Màn hình vận hành hôm nay; Member/PT xem kết quả riêng | Lễ tân ghi từng người qua UI → DB; mobile/keyboard/error states có kiểm tra |
| C4 — Review và chỉ số PT | Review 1–5 sao, một review/booking; rating/chuyên cần có mẫu số rõ | Đủ điều kiện mới review; concurrent/rollback/privacy và công thức PASS |
| C5 — Reports và bàn giao | Báo cáo vận hành, metrics PT SELF, lọc ngày/phòng và demo tổng | Số liệu khớp SQL nguồn; 5 vai trò đúng quyền; hợp đồng/runbook/kiểm chứng khớp code |

Mỗi chặng thực hiện: triển khai → test → phân tích lỗi → sửa → test lại → commit theo chức năng. Không ghi hoàn thành chỉ vì build hoặc endpoint trả 200.

## 3. C0: làm cho Booking test được từ đầu đến cuối

1. Thêm lệnh **fixture demo nghiệp vụ riêng**, opt-in và chỉ nhận DB local dev/test của project. Tạo các tài khoản/hồ sơ demo chuyên biệt thay vì tự duyệt hoặc gia hạn hồ sơ người dùng đang có. Không thay mật khẩu, cover, lịch hoặc booking đã tồn tại.
2. Fixture có Hội viên còn hạn, PT synthetic đã duyệt với lý do ghi rõ demo, giờ làm/ngày nghỉ, ba phòng có ảnh minh họa và ca trong lịch đã công bố. Mốc thời gian tương đối theo Asia/Ho_Chi_Minh; lần chạy lặp không tạo bản ghi trùng. Kiểm tra điều kiện của cả fixture đã có trước khi báo ready.
3. Rà `/pt/history`: tên **Lịch sử buổi tập**, mặc định các buổi đã qua/trạng thái kết thúc; `/pt/schedule` phục vụ lịch làm việc, `/pt/requests` phục vụ yêu cầu. Không dùng cùng tiêu đề khiến người dùng khó phân biệt.
4. Rà Member booking: đổi phòng phải xóa ca/PT/checkout cũ; hạn thẻ/hồ sơ thiếu hiển thị hướng dẫn; lỗi ca đầy hoặc PT bận có cách chọn lại; retry không nhân đôi. Cải thiện empty/error/loading, cỡ chữ, vùng chạm và focus; không dựng lại toàn bộ UI đã có.
5. Bổ sung test accept-vs-expire/cancel, create-vs-calendar/room/capacity/membership/availability; kiểm tra trạng thái cuối, hold và audit. Khi gặp deadlock/timeout, chỉ retry lỗi tạm thời có giới hạn; không retry lỗi nghiệp vụ và không trả SQL trong response.
6. Kiểm tra expiration backlog >100 và khởi động lại; đề xuất quét nhiều batch trong ngân sách thời gian, tránh một lỗi booking chặn toàn backlog. Chốt retention idempotency và replay; không dọn key chỉ dựa vào expires_at mà chưa định nghĩa retry sau hạn lưu.
7. Browser QA khi công cụ cho phép: 375/768/1024/1440px, zoom 200%, keyboard, dialog focus và 401. Tôn trọng giới hạn công cụ hiện có; không dùng kênh khác để vượt chặn. Test API/UI unit vẫn tiếp tục độc lập.

## 4. C1: Attendance — REQ-09 / UC-04

### Quyền và hành vi

- Lễ tân/Manager/Admin được ghi; Member/PT chỉ đọc theo SELF/ASSIGNED. Member/PT không có nút tự check-in, không thêm QR/RFID vào chặng này.
- Input chọn `participant_type = MEMBER | PT`; server suy ra participant_user_id từ booking. Không tin ID người được điểm danh hoặc người ghi do client cung cấp.
- Chỉ booking CONFIRMED và người thực sự thuộc booking. Booking tự tập không có hàng điểm danh PT; pending/rejected/expired/cancelled không được check-in.
- Điểm danh MEMBER và PT là hai thao tác riêng. Có mặt của PT không thay thế có mặt của Hội viên và không tự chuyển booking sang COMPLETED.
- Unique `(booking_id, participant_type)` hiện có bảo vệ ghi trùng. Chốt trả bản ghi hiện có với HTTP 200 khi ghi lại cùng người; lần đầu 201. Không thay checked_in_at/người ghi hoặc tạo thêm audit khi replay.
- Ghi attendance và audit trong một transaction; audit lỗi thì rollback toàn bộ. Không hard-delete hoặc sửa thời gian điểm danh bằng form thường.

### Mốc giờ đề xuất cho demo local

Cửa sổ thường: **`starts_at - 15 phút <= now < ends_at`**, dùng đồng hồ server/DB và timezone Việt Nam. Đây là giả định mới vì baseline chưa nêu số phút; ghi decision log và test cả hai biên trước khi triển khai.

Override ngoài cửa sổ chỉ áp dụng **trước khi buổi kết thúc**, đúng quyền attendance.record và có lý do; UI xác nhận rõ người/buổi/lý do. Không cho thao tác điểm danh thường vượt trạng thái kết thúc hoặc hủy. Ghi bổ sung/sửa sai sau buổi cần một quy trình correction riêng với audit, chưa nằm trong C1; tránh để completion job và thao tác đến muộn sửa số liệu ngầm.

### API dự kiến dưới /api/v1

| Endpoint | Input/kết quả | Quyền |
|---|---|---|
| `POST /bookings/{id}/attendance` | participant_type, override_reason tùy điều kiện; trả attendance | Lễ tân/Manager/Admin |
| `GET /bookings/{id}/attendance` | danh sách attendance thực tế, trạng thái từng participant và server_time | SELF/ASSIGNED/ALL |

Giữ envelope/request_id và lỗi ổn định: 403 sai vai trò, 404 ngoài object scope, 409 trạng thái không cho phép, 422 input/participant/cửa sổ sai. Schema OpenAPI phải tách attendance đã ghi khỏi trạng thái dự kiến của participant; không tạo bản ghi DB giả cho người chưa đến.

### Tìm kiếm vận hành

Mở rộng GET /bookings hiện có theo quyền: ngày/from/to, status, room_id và từ khóa booking ID/mã/tên Hội viên. SĐT chỉ được tìm bởi ba vai trò vận hành; không đưa SĐT hoặc bộ lọc đó sang payload public của Member/PT. Phân trang mặc định 20, tối đa 100; không thêm API danh sách cạnh tranh.

## 5. C2: Completion job và tranh chấp

- Chỉ xử lý booking CONFIRMED khi `now >= ends_at`: có attendance MEMBER → COMPLETED, không có → NO_SHOW. Trạng thái PT attendance được giữ riêng cho chuyên cần.
- Pending hết hạn vẫn thuộc expiration job; cancelled/rejected/expired không trở thành no-show. Completion không thay snapshot phiên hoặc dữ liệu attendance.
- Job chạy startup và định kỳ, quét batch có giới hạn, tiếp tục backlog ở lượt sau; chạy lặp/restart không nhân đôi audit. Không mở endpoint public cho người dùng tự đánh dấu COMPLETED/NO_SHOW.
- Audit trạng thái và cập nhật booking cùng transaction. Rà BR-14 và toàn bộ writer trước khi bổ sung: cùng thứ tự room → member → PT nếu có → gate ca → session → booking; attendance không lấy lock theo chiều ngược lại.
- Hủy và check-in sau ends_at phải bị kiểm soát theo server time ngay cả khi job chưa kịp chạy. Không dựa vào trạng thái CONFIRMED đang trễ để cho hủy một buổi đã kết thúc. Chốt cách xử lý buổi đã có attendance trước khi cho hủy ngoại lệ; đề xuất chặn hủy thường khi đã điểm danh, correction riêng nếu thực tế ghi sai.
- Thử cancel-vs-check-in và completion-vs-check-in: một trạng thái cuối hợp lệ, không có attendance cho booking bị hủy do request cạnh tranh, không mất attendance đã commit hợp lệ.

## 6. C3: UI theo công việc và vai trò

Áp dụng skill frontend-design đã dùng, giữ nền sáng và identity hiện có. Điểm nổi bật của màn hình vận hành là danh sách người sắp đến và hành động điểm danh; không thêm hero ảnh hoặc KPI trang trí vào bàn làm việc Lễ tân.

```text
Điểm danh hôm nay                     Ngày / Phòng / Trạng thái
[Tìm mã booking, mã Hội viên, tên hoặc SĐT]
Giờ      Hội viên         Phòng       MEMBER          PT
09–10    Tên / mã HV      Strength    [Điểm danh]     [Điểm danh]
10–11    Tên / mã HV      Cardio      Đã đến 09:52    Tự tập
```

- Admin/Manager/Lễ tân: `/management/attendance`, tìm nhanh, mở chi tiết, hai hàng participant với nút rõ ràng. Ghi thành công cập nhật đúng hàng, thời gian/người ghi và focus; request đang chạy khóa thao tác lặp. Override dùng dialog nhập lý do.
- Mobile: mỗi booking một khối gồm giờ/phòng/người, hai participant riêng; không ép bảng rộng thành cuộn ngang toàn trang. Vùng chạm tối thiểu khoảng 44px.
- Member: lịch của tôi/chi tiết hiển thị “Đã điểm danh”, “Buổi hoàn thành” hoặc “Vắng mặt” từ API; không có nút điểm danh.
- PT: lịch/chi tiết được gán có kết quả MEMBER/PT riêng. Lịch sử dùng ngày, trạng thái và tên buổi rõ ràng; không hiện số liệu giả hoặc quyền tạo booking.
- Giữ Be Vietnam Pro cho form/danh sách; Member giữ ảnh phòng và xanh lá, PT giữ xanh dương. Nội dung căn trái, cỡ chữ dễ đọc; trạng thái có chữ, không chỉ dùng màu. Loading, empty và error đều có hướng hành động và thử lại.

## 7. C4–C5: Review, metrics và báo cáo sau attendance

### Review — REQ-19 / UC-05

- `POST /bookings/{id}/review`: MEMBER SELF, rating 1–5, comment giới hạn đề xuất 1000 ký tự. Chỉ COMPLETED, có PT và attendance của cả hai; server suy ra member/PT.
- Một review/booking bằng unique constraint; concurrent requests chỉ tạo một review. Đã có review trả 409 REVIEW_EXISTS; chưa đủ điều kiện 422 REVIEW_NOT_ELIGIBLE.
- `GET /pts/{id}/reviews` phân trang và chính sách quyền theo ma trận; không trả username/contact/chứng chỉ hoặc dữ liệu người đánh giá ngoài nhu cầu. PT đọc review của mình, Member đọc review công khai của PT được phép hiển thị; vận hành theo ALL. Chốt projection an toàn trước khi triển khai.
- UI đánh giá tại chi tiết buổi đã hoàn tất; sao thao tác được bằng bàn phím, có nhãn, thông báo thành công và trạng thái đã đánh giá. Không thêm sửa/xóa review nếu chưa có quy tắc.
- Review, audit và cập nhật cache rating nếu dùng cache phải cùng transaction. Chọn một nguồn tính chuẩn; không để field cache và report query có công thức khác nhau.

### Rating và chuyên cần — REQ-20

Rating = AVG(review hợp lệ), kèm số review; không có review hiển thị N/A. Chuyên cần PT theo BR-22, chỉ tính buổi đã kết thúc; từ chối/hết hạn không đưa vào mẫu số, buổi cơ sở hủy loại trừ; các trường hợp PT/Hội viên hủy cần chốt rõ dựa trên việc đã xác nhận và thời điểm hủy.

Schema hiện không có confirmed_at/provenance vai trò lúc hủy đầy đủ. Trước khi tính chuyên cần, chốt lưu snapshot xác nhận và loại nguồn hủy trong migration mới; không suy ra lịch sử từ vai trò hiện tại của actor hoặc chuỗi lý do tự do. Bản ghi legacy chỉ backfill khi có chứng cứ audit; không tự bịa thời gian xác nhận hoặc phân loại người hủy. Báo cáo phải chỉ rõ bản ghi thiếu dữ kiện/chưa đưa vào mẫu số.

### Reports — REQ-10 / UC-14

- `GET /reports/operations`: Admin/Manager, ngày/phòng/trạng thái; số booking và phân bố trạng thái, attendance, no-show và mức sử dụng phiên với công thức/mẫu số công khai.
- `GET /reports/pts/{id}`: PT SELF hoặc Admin/Manager ALL; rating/chuyên cần/số mẫu riêng, không tạo Trust Score hoặc doanh thu.
- Định nghĩa rõ báo cáo trạng thái hiện tại và đếm sự kiện trong kỳ; không trộn hai cách. Khoảng ngày theo Asia/Ho_Chi_Minh; tỷ lệ mẫu số 0 là N/A, không phải 0%.
- Các aggregate phải tránh nhân đôi do JOIN attendance/reviews; đối chiếu fixture và SQL độc lập trước khi làm chart. Không gắn rating/occupancy giả vào dashboard để lấp giao diện.
- Export Excel/PDF là Could, chỉ sau Must/Should và công thức ổn định. Backup/restore, NFR và production là cổng riêng; không coi build/test unit là đã đáp ứng tải 100 người hoặc p95.

## 8. Tổ chức backend, kiểm thử và bàn giao

Thêm `domain/attendance`, `dto/v1/attendance.go`, `handler/v1/attendance_handler.go`, `service/v1/attendance_service.go`, `repository/attendance_repository.go`, routes/app wiring và completion job. Review/report làm tương tự sau C3. Không tạo `internal/model` trùng với domain; service/domain không import Gin, handler không viết SQL.

| Nhóm kiểm tra | Trường hợp bắt buộc |
|---|---|
| Quyền | 5 vai trò; Member/PT POST attendance 403; GET SELF/ASSIGNED, đoán ID ngoài scope; PT tự tập không có participant PT |
| Attendance | Đúng từng người; biên starts-15 phút/ends_at; override lý do; input ID giả/field thừa; replay/concurrent unique |
| Transaction | Attendance/audit rollback; completion/audit rollback; không để ghi partial hoặc token/hash/SQL vào log/response |
| Job/race | Cả MEMBER có/không có mặt, PT có/không có; restart/backlog/lặp; check-in/cancel/completion cạnh tranh, dữ liệu cuối và audit |
| Frontend | Tìm/lọc/page, form/confirm/duplicate click, loading/empty/error, 401, 403, focus, layout riêng và thời gian server |
| Review/metrics | Điều kiện hai attendance, rating ngoài 1–5, review trùng/concurrent/rollback, privacy, mẫu số 0, cancellation provenance/legacy |
| Reports | Quyền SELF/ALL; timezone/range; không double-count; aggregate so với dữ liệu nguồn; ghi rõ dataset/filters/công thức |

Demo DB thật: Hội viên đặt cùng PT → PT nhận → Lễ tân điểm danh riêng cả hai → sau ends_at job chuyển COMPLETED → Hội viên review → PT xem metrics → Manager xem báo cáo. Luồng phụ: không có attendance MEMBER → NO_SHOW; PT đến nhưng Member vắng vẫn NO_SHOW; buổi tự tập có Member → COMPLETED nhưng không review PT.

Không chờ thời gian thực bằng sleep dài trong test: unit dùng clock kiểm soát; integration fixtures giữ snapshot/constraint hợp lệ và chỉ thao tác DB test. Khi demo app thật, hiển thị thời gian demo rõ ràng, không giả dữ liệu buổi thật đã diễn ra.

Mỗi chặng bàn giao code, migration nếu cần, OpenAPI/Postman khớp router, hướng dẫn PowerShell, fixture riêng, báo cáo test có phần chưa kiểm chứng. Chỉ reset/rollback `ttcsn_gym_test`; DB dev giữ nguyên. Không sửa Excel, tự ghi họp/phỏng vấn/phê duyệt hoặc tắt máy.

**Bước triển khai đầu tiên khi kế hoạch được giao thực hiện:** C0 — fixture nghiệp vụ, kiểm tra luồng demo và khép các lỗi Booking hiện có; sau đó C1 Attendance và C2 Completion. C4–C5 không chạy trước khi kết quả C1–C3 có bằng chứng.
