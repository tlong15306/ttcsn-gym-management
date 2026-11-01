# Kế hoạch triển khai sau demo Identity tuần 4

**Ngày lập:** 11/10/2026. **Trạng thái:** Kế hoạch, chưa phải biên bản nghiệm thu.

## 1. Điểm xuất phát và nguồn đối chiếu

Backend hiện có config, health, PostgreSQL, Goose migration, seed và Identity/Auth; frontend có đăng nhập, hồ sơ cá nhân và quản lý tài khoản Admin. Đây là demo Identity, chưa phải MVP GYM đầy đủ.

Snapshot mã nguồn: commit `8c4e7de`. Các kiểm tra trước commit và giới hạn được ghi tại `reports/week4_verification_2026-10-11.md`.

Snapshot trên là trước gói CRUD quản trị. Sau yêu cầu triển khai ngày 11/10/2026, các module Hội viên, Phòng, Ca, Phiên phòng và PT đã có backend/UI và kiểm thử local. Kết quả hiện hành tại `reports/admin_crud_verification_2026-10-11.md`; Booking, Attendance, Review và Report vẫn chưa triển khai.

Theo yêu cầu UI/UX mới ngày 11/10/2026, `19_ke_hoach_crud_admin_ui_ux.md` cụ thể hóa CRUD Tài khoản/Hội viên/Phòng/Ca/PT. UI từng module được làm cùng backend tương ứng; các cổng nghiệp vụ và thứ tự phụ thuộc dưới đây vẫn áp dụng. Kế hoạch UI mới đã được người dùng giao triển khai.

Phần bổ sung trước Booking tại [20 — Lịch hoạt động phòng](20_ke_hoach_lich_hoat_dong_phong.md) đã có backend/migration/UI: bảng 7 ngày × 24 giờ, Admin/Manager đăng ký với preview/xác nhận, lặp tuần + ngoại lệ ngày. Kiểm thử và giới hạn tại `reports/room_calendar_verification_2026-10-11.md`; còn kiểm tra trực tiếp UI ngoại lệ, chỉ đọc và responsive sau khi hạn mức công cụ cho phép. Booking tiếp theo phải tuân thủ gate giờ phòng và thứ tự khóa đã chốt; không thay mốc/phân công Excel.

Nguồn kế hoạch:

Cập nhật 12/10/2026: Booking, ảnh phòng và portal Hội viên/PT theo kế hoạch 22 đã có mã; PostgreSQL/API, frontend tests/build và checker PASS trong phạm vi `reports/booking_verification_2026-10-12.md`. Browser QA và các race còn thiếu ghi riêng. Chặng nghiệp vụ tiếp theo là Attendance/Check-in, rồi Review/Report; không tự đổi mốc Excel.

- `../outputs_document/full_10_weeks/Ke_hoach_lam_viec_nhom_GYM.xlsx`, các sheet Tuần 5–10: mốc lịch gốc và công việc nhóm.
- `13_backlog_v0.3.md`: ưu tiên REQ và thứ tự phụ thuộc.
- `00_quyet_dinh_va_quy_tac_nghiep_vu.md`, `06_ma_tran_phan_quyen.md`, `07_SRS_v0.3.md`, `09_ERD_va_CSDL.md`, `12_dac_ta_API.md`: nguồn nghiệp vụ hiện hành.
- `15_ke_hoach_kiem_thu.md`: cổng kiểm thử.

Excel hiện xếp Booking/Check-in/Review ở tuần 5, nhưng các API dữ liệu nền cần cho chúng còn thiếu. Lịch dưới đây **đề xuất điều chỉnh thứ tự triển khai theo phụ thuộc**, giữ mốc tuần để đối chiếu. Không sửa ngày, người thực hiện hoặc trạng thái trong Excel; mọi mốc tương lai vẫn là dự kiến.

Những nội dung cũ trong Excel cần đối chiếu với v0.3 khi triển khai: check-in được lưu cho từng người qua `/bookings/{id}/attendance`, review qua `/bookings/{id}/review`; không thêm trạng thái `Attended` cạnh tranh với vòng đời booking; không triển khai doanh thu/thanh toán hoặc Trust Score. JWT giữ trong memory, hết phiên phải đăng nhập lại; chưa có refresh token.

## 2. Giữ cách phân chia package đã yêu cầu

Tiếp tục dùng cấu trúc của project Gin tham chiếu `lesson12-project-shopping-cart`:

```text
server/internal/
  app/                    # ghép dependency, khởi động và shutdown
  config/
  db/migrations/          # migration mới có phiên bản
  domain/{identity,member,facility,pt,booking,...}/
  dto/v1/                 # request/response HTTP
  handler/v1/             # bind DTO, gọi service, map lỗi
  middleware/             # authentication, request ID, CORS, recovery
  repository/             # interface, SQL, transaction và audit
  routes/v1/              # đăng ký endpoint
  service/v1/             # validation, quyền và nghiệp vụ
server/pkg/auth/
```

Thêm file theo module trong các package hiện có, ví dụ `member_handler.go`, `member_service.go`, `member_repository.go`. Service/domain không import Gin; handler không viết SQL. Một thao tác nhiều bảng phải dùng một transaction ở repository, bao gồm thay đổi nghiệp vụ, audit và idempotency nếu có. Không để mỗi repository tự commit riêng trong cùng thao tác.

Giữ DeepSeek Harness local hiện tại. Codex triển khai trực tiếp trong project; không bổ sung harness khác. PostgreSQL riêng tiếp tục dùng cổng 5433 với `ttcsn_gym_dev` và `ttcsn_gym_test`; reset/rollback chỉ trên DB test.

## 3. Chốt phần kiểm chứng Identity còn thiếu — dự kiến 11–14/10

| ID | Công việc | Điều kiện hoàn thành |
|---|---|---|
| NEXT-01 | Mở rộng bộ test JWT/config/health và API Identity | Có case thiếu từng claim bắt buộc, cấu hình sai/default và timeout readiness; response không lộ secret. |
| NEXT-02 | Kiểm thử PostgreSQL cho account và audit | Duplicate email/username; lỗi audit làm rollback thay đổi; đổi role có/không hồ sơ; thao tác đồng thời không làm mất Admin hoạt động cuối cùng. |
| NEXT-03 | Tự động hóa luồng API 5 vai trò và thu hồi token | Non-Admin 403; `/me` không đổi role/status; token cũ bị từ chối sau lock, reset mật khẩu, đổi role và logout. Seed chạy lặp giữ mật khẩu cũ. |
| NEXT-04 | Hoàn thiện hợp đồng Identity | OpenAPI có schema request/success/error/pagination và ví dụ; Postman đủ mọi endpoint đã triển khai; 204 không mô tả body JSON. |
| NEXT-05 | Kiểm tra demo bằng browser và mở rộng frontend test | Login/profile/Admin actions, lỗi API, 401, empty/loading, tìm kiếm/lọc/phân trang; kiểm tra desktop và 375px; kiểm tra truy cập trực tiếp trang quản trị sau khi có routing. |

Chỉ ghi nghiệm thu đầy đủ tuần 4 khi có bằng chứng tương ứng. Các test đang có và lần chạy API trước là bằng chứng một phần; không suy diễn thành toàn bộ ma trận bảo mật hoặc UI đã PASS.

## 4. Chặng tiếp theo: dữ liệu nền — tuần 5, dự kiến 15–21/10

Mục tiêu: dữ liệu thật và API có đủ điều kiện đầu vào để Booking hoạt động; chưa làm Booking trước khi chặng này đạt cổng nghiệm thu.

| ID | Hạng mục và REQ | Phạm vi triển khai | Nghiệm thu |
|---|---|---|---|
| W5-01 | Hội viên — REQ-03/11 | Domain, repo, service, handler; tạo account + profile atomically; đọc/sửa hồ sơ, membership, tìm kiếm, lọc và phân trang. | Account/profile/audit rollback cùng nhau; email/username trùng có lỗi ổn định; quyền SELF/ASSIGNED/ALL và dữ liệu tối thiểu cho PT đúng ma trận. |
| W5-02 | Phòng và ca — REQ-04/05 | Danh mục phòng; capacity/status; ca toàn cơ sở không chồng giờ; khóa sửa giờ ca đã dùng, cho ngừng hoạt động. | Ca liền kề hợp lệ, ca chồng bị từ chối kể cả concurrent requests; Manager/Admin quản lý, các role chỉ đọc đúng quyền. |
| W5-03 | Phiên phòng — REQ-05 | Phiên theo room/date/slot; snapshot giờ và capacity; lọc lịch; cập nhật capacity/status kèm lý do/audit. | Một phiên duy nhất cho room/date/slot; timezone đúng; không giảm capacity dưới số booking giữ chỗ; không phá lịch đã có. |
| W5-04 | Hồ sơ PT và duyệt — REQ-16/17 | PT tạo/sửa/nộp hồ sơ, Manager/Admin duyệt/từ chối/đình chỉ; lưu lịch sử quyết định; certificate riêng tư. | Transition đúng baseline/schema; MEMBER chỉ thấy PT APPROVED; tải certificate kiểm tra quyền mỗi lần; kiểm tra MIME/nội dung/kích thước, tên file và đường dẫn an toàn. |
| W5-05 | PT availability — REQ-16 | Lịch tuần và khoảng nghỉ; kiểm tra active account, trạng thái hồ sơ, ca và lịch bị ảnh hưởng. | Không chồng khoảng; không cho sửa availability làm mất hiệu lực booking đang giữ chỗ mà không xử lý theo quy tắc. |
| W5-06 | Dữ liệu demo và hợp đồng | Fixture riêng cho test; mở rộng OpenAPI/Postman và hướng dẫn PowerShell; chuẩn bị màn hình nhập dữ liệu nền. | Seed lặp an toàn; từ DB mới tạo được Member, PT được duyệt và room session bằng API thật. |

Thứ tự code: W5-01 → W5-02 → W5-03 → W5-04 → W5-05 → W5-06. Trước mỗi module đối chiếu field/constraint trong migration với SRS/API; nếu cần sửa DB đã có dữ liệu thì thêm migration mới, không sửa migration v1 đã áp dụng.

Avatar và certificate dùng chính sách file khác nhau: avatar chỉ công khai sau validation; certificate chỉ tải qua endpoint có quyền. Không dùng URL certificate công khai để thay kiểm tra quyền.

**Cổng tuần 5:** unit và PostgreSQL integration tests PASS; kiểm tra quyền theo object; dữ liệu tạo qua API tồn tại trong DB; account MEMBER/PT và hồ sơ tương ứng liên kết đúng; Go tests/vet, docs checker PASS. UI nghiệp vụ đầy đủ tiếp tục ở tuần 7; thiếu test hoặc file policy sẽ được ghi rõ thay vì coi đã hoàn tất.

## 5. Lộ trình tuần 6–10 sau khi điều chỉnh phụ thuộc

| Mốc dự kiến | Công việc chính | Cổng nghiệm thu |
|---|---|---|
| Tuần 6: 22–28/10 | Booking, PT phản hồi, hủy, lịch/polling; expiration job; tiếp tục hoàn thiện routing/layout frontend đã có. REQ-06/07/08/13/14/18. | Transaction giữ chỗ, row lock, unique constraints, audit và idempotency đúng; chạy 20 request vào chỗ cuối chỉ có một booking giữ chỗ; member/PT không trùng ca; retry không tạo bản ghi mới. |
| Tuần 7: 29/10–04/11 | Attendance, kết thúc buổi/no-show, review, rating/chuyên cần; UI hội viên/phòng/ca/phiên/PT. REQ-09/19/20 cùng UI REQ-03/04/05/16/17. | Điểm danh từng người; completed/no-show chỉ sau khi hết buổi; chỉ booking đủ điều kiện được review một lần; rating và chuyên cần đối chiếu dữ liệu nguồn; UI ghi vào DB thật. |
| Tuần 8: 05–11/11 | UI lịch, chọn PT, đặt/hủy, PT nhận/từ chối, check-in và review; hoàn thiện demo 5 vai trò. | E2E đủ luồng bằng browser; forbidden/expired/loading/empty states rõ; responsive 375px/desktop; backend vẫn chặn quyền khi gọi API trực tiếp. |
| Tuần 9: 12–18/11 | Báo cáo vận hành, tổng hợp PT, E2E/security/load, backup/restore và sửa lỗi. REQ-10; REQ-15 chỉ nếu Must/Should đã ổn định. | Báo cáo khớp booking/attendance; benchmark công bố workload, p95 và concurrency; restore dữ liệu sang DB riêng thành công; không còn lỗi Critical/High mở. |
| Tuần 10: 19–20/11 | Freeze bản đạt nghiệm thu, hướng dẫn 5 vai trò, báo cáo/slides, kịch bản demo và tổng duyệt. | Chạy từ hướng dẫn trên môi trường sạch; ghi giới hạn thật; tag release sau nghiệm thu; không ghi đã họp, bảo vệ hoặc được phê duyệt khi chưa diễn ra. |

Tuần 6 không tạo lại frontend từ đầu: giữ React/TypeScript/Vite/Tailwind và lockfile hiện có; tách `pages/components/services` khi thêm màn hình. Không tự đổi thư viện chỉ để khớp tên Axios/React Router trong lịch tham khảo; chọn phần cần cho routing/API theo code và contract đang dùng.

## 6. Các quy tắc cần test trước khi nghiệm thu Booking

- Chỉ member có account ACTIVE, profile hợp lệ và membership chưa hết hạn tại ngày tập mới đặt được. PT phải APPROVED, account ACTIVE, có availability và không trùng lịch.
- PENDING_PT và CONFIRMED đều giữ chỗ. Hạn phản hồi là mốc sớm hơn giữa 30 phút và bắt đầu phiên; từ chối/hết hạn giải phóng chỗ/PT, không tự chuyển sang booking không PT.
- Idempotency-Key có scope và hash request; cùng key/cùng payload trả kết quả cũ, khác payload trả conflict. Booking, hold, audit và idempotency commit/rollback cùng nhau.
- Khóa theo thứ tự ổn định; unique constraints bảo vệ xung đột member/PT. Test race phải kiểm tra dữ liệu cuối và số chỗ, không chỉ HTTP status.
- Member tự hủy trước giờ bắt đầu ít nhất 2 giờ, test cả biên bằng 2 giờ; override vận hành phải đúng quyền, có lý do và audit.
- Expiration/completion jobs chạy lặp an toàn. Test tranh chấp giữa accept/cancel/expire, giữa tạo booking và giảm capacity/khóa phiên.
- Mọi API áp dụng quyền theo bản ghi; không tin member_id/pt_id/created_by gửi từ client.

## 7. Cách bàn giao từng chặng

Mỗi chặng tạo một phần mã nguồn chạy được, migration/fixture nếu cần, contract API, hướng dẫn demo và báo cáo test ngắn. Chạy kiểm tra phù hợp thay đổi: Go unit/vet, PostgreSQL integration trên DB test, API test, frontend test/build và checker tài liệu; E2E/browser kiểm tra cho màn hình được bổ sung.

Commit theo chức năng sau khi test liên quan PASS. Trong báo cáo ghi rõ test đã chạy, case bỏ qua, lỗi còn mở và giới hạn. Nhật ký chỉ ghi hành động đã thực hiện; không tự ghi việc sinh viên đã hiểu bài hoặc nhóm đã phê duyệt.

Nếu dữ liệu nền/Booking chưa đạt cổng, ưu tiên hoàn thành Must và kiểm tra tính đúng; báo cáo rủi ro lịch và điều chỉnh dự kiến. REQ-15 xuất file là Could, có thể hoãn trước các Must; không lược bỏ quyền, transaction hoặc test race để giữ ngày dự kiến.

**Bước code tiếp theo (kế hoạch 12/10/2026):** C0 hoàn thiện kiểm chứng/fixture Booking, rồi Attendance/Completion và UI điểm danh theo `24_ke_hoach_attendance_review_report.md`. Booking và portal riêng theo kế hoạch 22 đã có code; các case kiểm chứng thiếu tiếp tục theo dõi. Review/metrics/report làm sau dữ liệu attendance. Giữ `handler/v1 → service/v1 → repository`, transaction/audit và PostgreSQL thật; lịch Excel không đổi.
