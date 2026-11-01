# Kế hoạch chi tiết: chốt điểm danh → đánh giá PT → báo cáo

Ngày cập nhật: **14/10/2026, Asia/Ho_Chi_Minh**. Tiếp nối kế hoạch 24 được lập ngày 12/10.

**Kế hoạch tiếp nối ngày 18/10/2026:** [26 — Chốt các phần còn thiếu sau rà tuần 1–6](26_ke_hoach_chot_phan_con_thieu.md). Các trạng thái có ngày bên dưới được giữ làm snapshot; phần kiểm chứng E2E/NFR và bàn giao tiếp tục theo kế hoạch 26.

**Trạng thái hiện tại (14/10/2026):** A `093469a`; B `f35775d`; C Review backend `a58c359`; D Review UI `a67db62`; E Reports API `92cb96f`; F UI/report `49b423a`. F đã thêm dashboard `/management/reports`, trang `/pt/performance`, trang `/pt/reviews`, test frontend và runbook. Frontend 45/45 tests, production build, contract checker và docs checker PASS. Báo cáo: [Report API verification](reports/report_api_verification_2026-10-14.md), [Report UI verification](reports/report_ui_verification_2026-10-14.md). API/client vẫn tắt; PostgreSQL test đã dùng đúng `ttcsn_gym_test`. Chưa chạy browser visual QA hoặc demo UI → DB; kiểm tra luồng live còn lại.

## 1. Điểm xuất phát đã đối chiếu trước kế hoạch 25

| Hạng mục | Hiện tại | Còn lại |
|---|---|---|
| Git | HEAD `4e60eda`; thay đổi Booking/Attendance chưa commit | Rà diff, kiểm chứng và commit theo chặng |
| Backend | Có `domain/attendance`, DTO/service/repository/handler và wiring | Chốt quyền/scope/replay/transaction/biên giờ ở bản cuối |
| Job | Startup + mỗi 15 giây; CONFIRMED → COMPLETED/NO_SHOW | Rà ngân sách job, backlog và shutdown |
| UI | `/management/attendance`, chi tiết Member/PT và `/pt/history` | QA responsive/keyboard, filters và error states |
| Fixture | `cmd/seed-booking-demo`, member.flow.demo/pt.flow.demo, ba phòng FLOW DEMO | Kiểm tra lại điều kiện theo ngày chạy; không ghi đè dữ liệu |
| Hợp đồng/schema | Working tree có 71 API operations; Goose v4, 18 bảng nghiệp vụ | Đồng bộ tài liệu, kiểm tra contract |
| Review/report | Ảnh chụp ban đầu: Review backend/UI đã có; Metrics/Report chưa có | E hiện đã thêm Reports API; F còn dashboard/bàn giao và trang riêng `/pt/reviews` |

Phiên 12/10 đã có kết quả Go/DB tests, **34 frontend tests + build**, checker và live API demo 5 vai trò PASS ở các lượt tương ứng. Ngày 14/10, Go suite trỏ đúng `ttcsn_gym_test` và các race writer/pending PASS; sau khi tách timeout job, job unit tests và `go vet` PASS. Frontend sau cập nhật bộ lọc phòng có 35 tests và production build PASS. Contract checker xác nhận 71 operations; docs checker PASS. Lệnh fixture dev bị dừng ở kết nối trước khi ghi dữ liệu vì Docker engine không truy cập được; không được tính là PASS. Browser QA trực tiếp còn thiếu. Port API/client 8080/5173 đã được kiểm tra không có listener và vẫn dừng theo yêu cầu.

Nguồn nghiệp vụ: decision log BR-14/17–26, SRS REQ-08/09/10/19/20, UC-03/04/05/14, ma trận quyền và schema v0.3. Giữ Excel/phân công/ngày thực tế, DeepSeek Harness, dữ liệu dev và file người dùng. Không tự ghi họp, phỏng vấn, phê duyệt chuyên môn hoặc mức hiểu bài của sinh viên.

## 2. Thứ tự và cổng hoàn thành

| Chặng | Phạm vi | Cổng bàn giao |
|---|---|---|
| A — Chốt C0–C3 | Booking, Attendance, completion, fixture, UI và kiểm thử cuối | Report + runbook + commit; giới hạn QA ghi riêng |
| B — Chuẩn dữ liệu | Thời điểm xác nhận, nguồn hủy, legacy và công thức | Decision log/trace matrix + migration mới + tests |
| C — Review backend | Tạo/đọc đánh giá, quyền, điều kiện và privacy | API/contracts + concurrent/rollback/scope tests PASS |
| D — Review UI | Hội viên đánh giá; PT/vận hành đọc | Form/readback, quyền và accessibility tests/build PASS |
| E — Metrics/report API | Rating, chuyên cần và báo cáo vận hành | SQL chuẩn + API đúng scope + số liệu khớp fixture độc lập |
| F — UI báo cáo/bàn giao | PT/Admin/Manager, demo tổng và hướng dẫn | Bộ kiểm chứng cuối + runbook + commits |

Mỗi chặng: triển khai → test → chẩn đoán → sửa → test lại → commit. Không gắn ngày hoàn thành vào Excel khi chưa có bằng chứng. Không bắt đầu Review/report khi backend điểm danh/completion còn lỗi bắt buộc.

## 3. Chặng A — Chốt phần đang dở

### A1. Booking và fixture

1. Kiểm tra Git/diff, router/contracts/dependency; giữ `.agents/`, file `s`, Excel và thay đổi người dùng. Không chỉ đếm endpoint.
2. Chạy lại `TestBookingConcurrentEligibilityWriters`: create-vs-room/calendar/capacity/membership/availability. Fixture dùng ngày Việt Nam; assert trạng thái cuối, holds/capacity, audit và rollback.
3. Chạy `TestPendingAcceptanceRaces`, cùng key, chỗ cuối và overlap. PT nhận sau deadline bị từ chối; hủy/expire giải phóng hold đúng một lần.
4. Khi có deadlock/serialization thực sự: sửa thứ tự khóa trước; retry lỗi tạm thời cho toàn transaction với giới hạn. Không retry validation/conflict hoặc ghi một phần.
5. Giữ chính sách đã ghi: key Booking đã commit chưa tự xóa kể cả sau expires_at; replay giữ kết quả tạo ban đầu. Retention/xóa key là thay đổi riêng.
6. Chạy fixture hai lần, kiểm tra membership/role/status/PT approval/hours/absence, cover và lịch phiên. Không tự sửa mật khẩu/hồ sơ/lịch cũ. Synthetic phải ghi rõ DEMO, không coi là phê duyệt/điểm danh thực tế.

### A2. Attendance và completion

- Chỉ Admin/Manager/Lễ tân ghi; Member/PT đọc SELF/ASSIGNED. Server suy ra participant/recorder, reject field thừa.
- Cửa sổ đã chốt: `starts_at - 15 phút <= DB now < ends_at`. Ghi sớm cần lý do; không thêm attendance sau kết thúc hoặc cho booking không CONFIRMED.
- Ghi hai người riêng; tự tập không có participant PT. Lần đầu 201; replay 200 giữ thời gian/người ghi/lý do, không thêm audit, kể cả sau kết thúc.
- Attendance + audit cùng transaction. Không hủy thường sau attendance hoặc sau ends_at, dù job chưa chạy.
- CONFIRMED hết giờ: MEMBER có mặt → COMPLETED, còn lại NO_SHOW; PT attendance độc lập. Không mở endpoint public tự hoàn thành.
- Khóa room → member → PT → gate → session → booking; test cancel-vs-check-in, completion-vs-check-in, rollback và cả bốn tổ hợp có mặt.
- Hiện Expire và Complete dùng chung context 30 giây: rà/tách ngân sách để expiration timeout không bỏ qua completion cả lượt. Batch/thời gian có giới hạn, restart xử lý tiếp, một booking lỗi không chặn các booking khác. Kiểm tra backlog >100, lặp và shutdown.

### A3. UI và kiểm chứng cuối

- Vận hành: lọc ngày/phòng/trạng thái, tìm mã booking/mã Hội viên/tên/SĐT, phân trang; chi tiết ghi riêng hai người và hiện thời gian/người ghi. SĐT chỉ cho vận hành, không đưa contact vào payload Member/PT.
- Bộ lọc phòng hiện lấy tối đa 100: rà tìm kiếm/tải thêm để phòng ngoài trang đầu không bị mất khỏi lựa chọn.
- PT history chỉ buổi đã qua/trạng thái kết thúc, tách lịch tuần/yêu cầu. Member/PT xem kết quả, không tự điểm danh.
- Rà đổi phòng/ngày xóa checkout cũ, double submit, refresh sau lỗi, 401/loading/empty/retry, filter URL, focus dialog và lý do chỉ có khoảng trắng.
- QA 375/768/1024/1440px, zoom 200%, keyboard, vùng chạm khoảng 44px. Dùng browser khi quyền/hạn mức cho phép; bị chặn thì ghi rõ phần thiếu, không vượt chặn qua kênh khác.
- Chạy Go tests/vet, PostgreSQL/API integration, frontend tests/build, contract/docs checker. Completion dùng DB test/clock kiểm soát, không sửa snapshot dev để giả buổi đã diễn ra.

**Đầu ra A:** report ngày chạy thực tế, runbook PowerShell và commit `feat: finalize attendance and booking completion` với đủ wiring/fixture/contracts/docs. Browser QA chưa thực hiện phải ghi riêng.

## 4. Chặng B — Dữ liệu và công thức trước chỉ số

### B1. Migration và writer

| Trường dự kiến trên bookings | Ý nghĩa |
|---|---|
| confirmed_at nullable TIMESTAMPTZ | Tự tập: lúc tạo CONFIRMED; cùng PT: lúc nhận. Giữ khi hủy/kết thúc |
| cancellation_source nullable | MEMBER/PT/FACILITY/UNKNOWN; snapshot lúc hủy |
| cancelled_actor_role nullable | Vai trò actor tại lúc hủy, hỗ trợ audit |

- Đề xuất bám BR-22: Admin/Manager/Lễ tân hủy là FACILITY; Member/PT hủy theo actor. Không suy ra từ role hiện tại hoặc chuỗi lý do.
- Cập nhật Create/Act và mọi writer hủy, audit/DTO/contracts trong cùng thay đổi. Giữ khóa/transaction; không tạo lại attendance/review tables.
- Không backfill bằng updated_at/ngày migration hoặc timestamp đoán. Chỉ khôi phục khi có chứng cứ rõ và lưu nguồn; thiếu thì null/UNKNOWN, thống kê riêng.
- Không sửa/xóa dữ liệu cũ; fresh/repeat/up-down-up trên DB test. Rà tương thích API trước apply dev, thêm index khi query/EXPLAIN cho thấy cần.

### B2. Công thức đề xuất

Lọc kỳ theo booking.session_date Việt Nam, hai đầu bao gồm; chỉ tính buổi đã hết giờ theo DB. Response có as_of/timezone/range/tử số/mẫu số và lượng thiếu chứng cứ.

| Trường hợp buổi có PT đã hết giờ | Mẫu số chuyên cần PT | Tử số |
|---|---|---|
| COMPLETED/NO_SHOW, có chứng cứ từng xác nhận | Có | Có nếu attendance PT |
| MEMBER/PT hủy sau xác nhận | Có theo BR-22 | Theo attendance PT hợp lệ; hiện hủy sau attendance bị chặn |
| FACILITY hủy | Không | Không |
| REJECTED/EXPIRED/hủy khi pending | Không | Không |
| Chưa hết giờ | Không | Không |
| Legacy thiếu chứng cứ xác nhận/nguồn hủy | Không suy đoán | Báo unclassified_count riêng |

`attendance_rate = 100 × present_count / eligible_count`; mẫu số 0 → null/N/A. NO_SHOW có mặt PT vẫn tăng tử số PT. Nếu muốn loại cả Hội viên hủy khỏi mẫu số, phải thay BR-22 rõ ràng trước code, không đổi ngầm.

Rating = AVG(review hợp lệ), kèm review_count; không review → null/N/A. Dùng SQL aggregate làm nguồn chuẩn giai đoạn này, không cập nhật cache bằng công thức khác. Rating trong kỳ dựa trên ngày buổi tập; review gửi sau được tính ở lần đọc hiện tại, có as_of.

**Đầu ra B:** decision log/REQ → UC → permission → entity → test + migration và legacy tests PASS. Commit `feat: track booking confirmation and cancellation provenance`.

## 5. Chặng C — Review backend

### API dự kiến dưới /api/v1

| Endpoint | Quyền/kết quả |
|---|---|
| POST /bookings/{id}/review | MEMBER SELF; tạo 201 |
| GET /bookings/{id}/review — bổ sung readback | SELF/ASSIGNED/ALL; 200 review hoặc data=null khi chưa có; ngoài scope 404 |
| GET /pts/{id}/reviews | Member public của PT được phép hiển thị; PT SELF; vận hành ALL; phân trang |

- Input chỉ rating nguyên 1–5, comment tùy chọn tối đa 1000 ký tự sau trim. Không nhận member_id/pt_id/status.
- Booking thuộc Hội viên đăng nhập, COMPLETED, đã hết giờ, có PT và đủ attendance hai người. Tự tập/NO_SHOW/CANCELLED/pending/thiếu attendance → 422 REVIEW_NOT_ELIGIBLE.
- Unique booking_id hiện có bảo vệ concurrent: một review commit; lần hai/request còn lại 409 REVIEW_EXISTS, một audit. Không thêm sửa/xóa review.
- Review + audit cùng transaction; kiểm tra eligibility bên trong bằng quy tắc khóa chung. Audit lỗi phải rollback review.
- Public projection chỉ nội dung/rating/ngày và nhãn người đánh giá tối thiểu (đề xuất “Hội viên”); không member/user ID, username, email, SĐT, địa chỉ, chứng chỉ hoặc override reason. Readback giữ object scope; operator chỉ nhận trường cần tra cứu.
- Review cũ không bị xóa vì tài khoản sau đó khóa/đổi trạng thái. Quyền public của PT theo chính sách profile hiện hành.
- Thêm rating/review_count public từ cùng query nguồn vào projection PT; không đọc cache cũ cho card chọn PT.

**Đầu ra C:** service/unit + API 5 vai trò/scope + validation/concurrent/rollback/privacy/pagination PASS; OpenAPI/Postman khớp router. Commit `feat: add scoped trainer reviews`.

## 6. Chặng D — UI Review riêng từng vai trò

Giữ frontend-design và identity hiện có: nền sáng; Hội viên xanh lá/phòng có ảnh; PT xanh dương/lịch làm việc. Bổ sung vào portal hiện có, không dùng chung shell vận hành hoặc thêm KPI/hero trang trí.

- **Hội viên:** vùng đánh giá trong chi tiết buổi đủ điều kiện; radio 1–5 sao dùng keyboard, comment đếm ký tự, xác nhận và khóa double submit. Đã đánh giá thì đọc lại nội dung/ngày/rating; không đủ điều kiện có lý do.
- **Chọn PT:** điểm trung bình + số đánh giá thật, không có thì “Chưa có đánh giá”; danh sách review phân trang, không contact Hội viên.
- **PT `/pt/reviews`:** đọc đánh giá nhận được, loading/empty/error/retry/page; không chọn PT khác hoặc viết/sửa review.
- **Vận hành:** tra cứu ở chi tiết booking/PT theo quyền; không giả Hội viên viết review.
- 409 tải lại review; 422 giữ comment đang soạn và giải thích điều kiện; 401 xóa phiên theo cơ chế hiện hành.

**Đầu ra D:** form/readback/eligibility/duplicate/401/403/409/422/keyboard tests + build; mobile/desktop QA có bằng chứng hoặc ghi thiếu. Commit `feat: add member reviews and coach review page`.

## 7. Chặng E — Metrics và report API

| Endpoint | Quyền | Dữ liệu |
|---|---|---|
| GET /reports/operations | ADMIN/MANAGER | from,to,room_id; booking/status/attendance/no-show/utilization và công thức |
| GET /reports/pts/{id} | PT SELF hoặc ADMIN/MANAGER ALL | from,to; rating/count, present/eligible/excluded/unclassified và tỷ lệ |

Đề xuất default tháng hiện tại tới hôm nay theo Việt Nam; from <= to, tối đa 366 ngày. ID/filter hợp lệ; 403 sai quyền, 404 ngoài scope, 422 sai input. Member/Lễ tân không đọc report dù có thể đọc review.

### Báo cáo vận hành

1. Tổng booking/phân bố theo trạng thái hiện tại tại as_of trong kỳ ngày buổi tập; không gọi là số sự kiện phát sinh trong kỳ.
2. Đếm MEMBER/PT attendance riêng bằng booking/participant duy nhất. Không JOIN attendance và reviews rồi COUNT toàn hàng.
3. `no_show_rate = NO_SHOW / (COMPLETED + NO_SHOW)` của buổi đã hết giờ; mẫu số 0 → null. CONFIRMED hết giờ đang chờ job báo riêng.
4. Sử dụng theo đặt chỗ = seat-minutes của booking CONFIRMED/COMPLETED/NO_SHOW trên phiên OPEN đã hết giờ / tổng capacity × duration của chính tập phiên đó. Pending/cancelled không tăng tử số; Member và PT không thành hai chỗ. Đây không phải thời gian tập thực tế.
5. Trả nhóm bị loại/legacy chưa phân loại, kỳ/timezone/công thức. Không doanh thu, Trust Score, forecast hoặc số giả.

### Query và dữ liệu kiểm chứng

- Fixture: tự tập, đủ hai attendance, PT-only, Member-only, không ai đến, từng nguồn hủy, rejected/expired, pending/future, qua nửa đêm và UNKNOWN.
- Đối chiếu API với SQL độc lập và kỳ vọng tính trước. Nhiều attendance/reviews không nhân đôi booking hoặc capacity.
- Rà EXPLAIN trên dataset local có kích thước ghi rõ; không tuyên bố đạt NFR tải/p95 production từ dataset nhỏ.
- Rating public/trang PT/report cùng nguồn; khi kỳ khác nhau phải ghi rõ “toàn bộ” hoặc range.

**Đầu ra E:** scope/filters/timezone/legacy/0 mẫu và công thức tests PASS; API khớp số nguồn. Commit `feat: add verified trainer metrics and operations reports`.

## 8. Chặng F — UI báo cáo và bàn giao

- **PT `/pt/performance`:** rating/count; chuyên cần với tử số/mẫu số; lọc kỳ, giải thích nhóm bị loại/UNKNOWN. Không chọn PT khác.
- **Admin/Manager `/management/reports`:** kỳ/phòng, tổng booking/trạng thái/sử dụng; chart phụ trợ kèm bảng số/nhãn, drill-down dùng cùng filters khi hỗ trợ.
- **Admin/Manager tra cứu PT:** bộ metrics giống PT SELF, có lượng thiếu chứng cứ. Empty khác lỗi; giữ filter/loading/retry; N/A khác 0%.
- Member/Lễ tân vào URL hoặc gọi API report đều bị từ chối. 375px không overflow toàn trang, keyboard/focus/reduced motion theo UI hiện có.
- Export Excel/PDF là Could riêng. Chưa làm correction sau buổi, QR/RFID, thanh toán, production hoặc đổi harness.

### Demo tổng và kỹ thuật

1. Hội viên chọn phòng có ảnh → ca/PT đủ điều kiện → đặt; PT nhận; Lễ tân điểm danh riêng hai người; completion; Hội viên review.
2. PT xem review/rating/chuyên cần; Manager đối chiếu report với booking nguồn.
3. Luồng phụ: PT đến/Member vắng → NO_SHOW nhưng PT có mặt; tự tập đủ Member → COMPLETED không review PT; FACILITY cancel/legacy UNKNOWN không tính sai.
4. Không sửa snapshot/attendance dev để giả thời gian. Test dùng DB test/clock; demo dùng dữ liệu synthetic và mốc giờ hợp lệ, ghi rõ phần chưa chạy UI thật.

- Backend giữ `domain/{attendance,review,report}`, DTO/service/repository/handler/routes/app/job; không tạo internal/model trùng domain. Gin ở HTTP, SQL/transaction/audit ở repository.
- Runbook PowerShell: Docker/DB → migrate → seed/fixture opt-in → API/client → tài khoản → dừng. Secrets giữ file local bị Git bỏ qua.
- OpenAPI/Postman, decision log, SRS/ma trận/trace matrix, backlog và reports đồng bộ. Checker không thay test nghiệp vụ.
- Commit bằng danh sách file cụ thể, giữ file người dùng. Không tự push/deploy/tắt máy. Không khởi động lại app trong lượt chỉ lập kế hoạch này.
- Báo cáo có build/commit/migration, ngày/môi trường/dataset, lệnh, PASS/FAIL và browser QA còn thiếu; không ghi nghiệm thu toàn MVP.

## 9. Checklist nghiệm thu

| Nhóm | Bắt buộc |
|---|---|
| Booking | Writer/race, last seat/overlap, replay/key conflict, deadline/holds/audit |
| Attendance/job | Scope 5 vai trò, biên giờ, reason/replay, rollback/races, backlog/restart/shutdown |
| Migration | Fresh/repeat/test rollback; giữ dev; legacy timestamp/source có chứng cứ |
| Review | Rating/comment/eligibility/SELF, duplicate/concurrent, audit rollback/privacy/page |
| Metrics | Từng nguồn hủy, PT-only/NO_SHOW, future/UNKNOWN/0 mẫu, cùng nguồn rating |
| Reports | SELF/ALL/403/404, kỳ Việt Nam/midnight, không double-count, SQL độc lập |
| UI | Form/readback/errors/retry/401/403, keyboard/focus/375px và layout riêng |
| Tổng | Go tests/vet + DB/API + frontend tests/build + contract/docs checker + runbook/commit |

**Bước tiếp theo:** browser QA ở 375px/desktop và demo tổng qua API/client đang chạy; đối chiếu booking, attendance, review và số report trên dữ liệu synthetic. Code UI F, runbook và kiểm thử component đã xong. Không chỉnh timestamp/attendance để giả thời gian; nếu dữ liệu demo không đáp ứng vòng đời, ghi rõ luồng nào chưa chạy. Hiện không chạy API/client.
