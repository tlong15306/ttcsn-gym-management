# Changelog tài liệu

## 2026-10-18 — Newman mutation Identity trên DB test cô lập

- Thêm collection và runner API tạm cổng riêng: tạo account synthetic, lock/unlock, reset/revoke token, RBAC và audit; API tạm dừng và test DB được reset sau chạy.
- Newman chạy 14 request/14 assertion PASS, có 4 audit account mutation. Report ghi rõ phạm vi chỉ Identity, không suy diễn sang browser E2E hay các mutation domain khác.

## 2026-10-18 — Vận hành, idempotency retention và restore local

- Thêm migration v6/index expiry và dọn key idempotency hết hạn cho Booking/Calendar; job dọn ở startup rồi mỗi giờ, có test replay/purge và startup job.
- Bổ sung unit test CORS: chỉ origin cấu hình nhận header allowlist/Vary, origin khác không được cấp quyền.
- Kiểm tra expiration backlog 1.001 bản ghi qua hai lượt; dữ liệu test cuối là `EXPIRED=1001`, `PENDING_PT=0`, audit `booking.EXPIRED=1001`.
- Thêm script tải chỉ đọc và backup/restore an toàn. Burst login/read 100 concurrent trả 400/400 nhưng p95 2.435,10 ms; warm-session read/ramp trả 300/300 với p95 80,92 ms. NFR-01 tổng thể chưa đạt, không đánh dấu PASS. Restore v6 sang DB mới khớp count và cover, nhưng RPO/RTO/bảo mật chuyên sâu vẫn mở.
- Ghi report vận hành, cập nhật runbook/backlog/ma trận; không sửa Excel, không suy diễn hoạt động hay phê duyệt của nhóm.

## 2026-10-18 — Regression local, Newman và browser smoke

- Thêm `start-local.ps1`, `stop-local.ps1`, `run-api-role-smoke.ps1` cùng collection Newman chỉ đọc 5 vai trò; runner dùng Docker để không phụ thuộc Newman cục bộ trên Node 24.
- Chạy lại PostgreSQL/Go/vet/frontend/contracts/docs; Newman 15 request/15 assertion PASS và browser smoke đủ 5 role. Report ghi rõ booking synthetic #4 đã tạo trên account/phòng FLOW DEMO.
- Sửa cảnh báo React key trùng ở chi tiết booking; xác nhận lại Vitest/build và điều hướng browser tới booking #4. Một lỗi `GET /bookings` PT 500 chỉ xuất hiện một lần, không tái hiện qua 20 lần gọi đúng filter tuần; vẫn ghi là giới hạn cần kiểm chứng tải/E2E.
- Ghi rõ chưa hoàn tất browser E2E vòng đời, accessibility/browser matrix, tải, security review hoặc backup/restore; không đổi Excel hoặc suy diễn hoạt động nhóm.

## 2026-10-18 — Đồng bộ trạng thái tuần 1–6 theo bằng chứng

- Thêm ma trận 27 để phân biệt rõ mã nguồn/kiểm thử kỹ thuật đã có với E2E browser, tải, restore và xác nhận hoạt động thực tế của nhóm còn thiếu.
- Cập nhật backlog v0.3: Attendance, Review, Rating và Report đã có API/UI cùng kiểm thử local được dẫn chiếu; không đánh dấu browser E2E hoặc NFR là đạt.
- Cập nhật mục lục tài liệu. Không sửa Excel, không tự ghi phỏng vấn, họp, phê duyệt hoặc ngày làm việc của sinh viên.

## 2026-10-18 — Kế hoạch chốt các phần còn thiếu

- Thêm kế hoạch 26 sau rà tuần 1–6: đồng bộ tiến độ, chuẩn chạy local, Postman Runner, browser E2E và UI QA, job/load/security/restore, hồ sơ bàn giao.
- Giữ riêng chứng cứ kỹ thuật với khảo sát/họp/phê duyệt thực tế của nhóm; chỉ đưa REQ-15 xuất file vào chặng sau các cổng bắt buộc.
- Liên kết từ mục lục và kế hoạch 25. Đây là lượt lập kế hoạch; chưa triển khai các chặng, chưa cập nhật Excel hoặc ghi một kết quả kiểm thử nghiệp vụ mới.

## 2026-10-14 — Giao diện Review và báo cáo

- Thêm `/management/reports` cho Admin/Manager với bộ lọc kỳ/phòng, chỉ số vận hành, trạng thái booking, nguồn hủy và tra cứu hiệu suất PT.
- Thêm `/pt/performance` và `/pt/reviews` trong portal PT riêng; có kỳ báo cáo, N/A cho mẫu số rỗng, review ẩn danh, trạng thái tải/lỗi/rỗng/thử lại và phân trang.
- Thêm 5 test UI; tổng frontend suite 45/45 PASS, production build PASS. Browser visual QA và luồng live UI → PostgreSQL chưa chạy do API/client đang tắt.
- Cập nhật runbook và kế hoạch 25 để mô tả quyền, công thức hiển thị và trình tự demo; chưa ghi nhận demo live đã hoàn thành.

## 2026-10-14 — Metrics và Report API

- Thêm `GET /reports/operations` cho Admin/Manager và `GET /reports/pts/{id}` cho PT xem chính mình, Admin/Manager xem tất cả.
- Cùng một truy vấn snapshot trả kỳ Việt Nam, `as_of`, tử số/mẫu số, tỷ lệ no-show, seat-minute utilization, số chờ completion, rating và nhóm excluded/unclassified. Không tính gộp participant/review gây nhân đôi booking; số 0 mẫu trả N/A.
- `go test ./... -count=1` trên PostgreSQL `ttcsn_gym_test`, `go vet`, contracts và docs checker PASS. Chi tiết, dataset nhỏ và giới hạn tại `reports/report_api_verification_2026-10-14.md`.

## 2026-10-14 — Review backend và rating PT

- Thêm Review API theo REQ-19: tạo một lần cho booking COMPLETED đã qua giờ, có PT và attendance của cả MEMBER/PT; GET theo booking; danh sách công khai phân trang.
- Review cùng audit trong một transaction; khóa theo thứ tự Booking, xử lý race duplicate bằng `REVIEW_EXISTS`; không cho sửa/xóa.
- Response công khai không có định danh/contact người đánh giá, chỉ hiện nhãn “Hội viên”. PT profile và booking options tính `rating_average`/`review_count` từ các review đủ điều kiện.
- Hợp đồng OpenAPI/Postman có 74 route tại thời điểm commit C; kiểm chứng chi tiết tại `reports/review_backend_verification_2026-10-14.md`.

## 2026-10-14 — Review UI cho Hội viên, PT và vận hành

- Thêm đánh giá 1–5 sao và nhận xét tùy chọn tại chi tiết booking đã hoàn thành; nút sao có nhãn rõ, dùng radio bàn phím được, báo lỗi và hiện ngay bản đã lưu.
- PT và vận hành đọc đánh giá trong chi tiết booking; danh sách PT và lựa chọn PT hiển thị rating/count từ API, trường hợp chưa có đánh giá hiện N/A.
- Frontend test/build PASS; browser visual QA chưa thực hiện do API/client vẫn dừng theo yêu cầu.
- Màn hình riêng `/pt/reviews` theo kế hoạch D chưa có; PT hiện đọc review tại booking được gán. Bổ sung vào chặng UI cuối, không tính phạm vi này là đã xong.

## 2026-10-12 — Booking, cover và portal riêng

- Thêm Booking domain/service/repository/handler/job, options toàn phiên, SELF/ASSIGNED/ALL, idempotency, giữ chỗ, PT response, hủy và expiration/audit transaction.
- Migration v4 thêm cover BYTEA/thumbnail/alt/version/demo; Admin/Manager quản lý ảnh, Bearer/ETag khi đọc. Seed ba phòng minh họa riêng, giữ cover/lịch đã có.
- Hội viên chọn ảnh phòng trước ngày/ca/PT; PT có lịch tuần/danh sách và queue riêng; vận hành đặt hộ đúng quyền. Font self-host có OFL.
- 69 operations OpenAPI/Postman; runbook 23 và báo cáo kiểm chứng nêu rõ giới hạn browser. Không sửa Excel hoặc ghi hoạt động sinh viên chưa có bằng chứng.

## 2026-10-11 — Bổ sung ảnh phòng vào luồng Hội viên

- Điều chỉnh kế hoạch 22: xem danh sách phòng có ảnh đại diện → nhấn phòng → chọn ngày/ca/PT → xác nhận, thay tab chọn phòng của bản phác đầu.
- Thêm bản phác catalog/chi tiết desktop và 375px; tạo ba ảnh phòng minh họa bằng `image_gen`, lưu prompt và nguồn minh họa rõ ràng.
- Bổ sung kế hoạch quản trị ảnh cho Admin/Manager, metadata/thumbnail/placeholder/quyền upload/đọc, migration giữ phòng cũ và test ảnh/đúng phòng.
- Chưa triển khai ảnh phòng hoặc Booking vào app/API/DB; giao diện PT vẫn riêng, không đổi quyền tạo booking; giữ Excel và báo cáo thực tế.

## 2026-10-11 — Kế hoạch Booking và giao diện Hội viên/PT riêng

- Ghi lựa chọn cả hai nền sáng, mỗi vai trò một màu nhấn; thiết kế lại sau khi người dùng đánh giá hai mẫu tham khảo đầu chưa hiện đại.
- Thêm kế hoạch 22, MemberLayout/CoachLayout, bản phác desktop/375px và cổng kiểm tra Booking/options/expiration/idempotency/race/object scope.
- Cài và áp dụng skill `frontend-design` của Anthropic; sửa dòng bước tiếp theo còn cũ trong kế hoạch 18, liên kết vào mục lục.
- Chỉ cập nhật tài liệu/bản phác; chưa sửa code ứng dụng, API/migration, DB hoặc Excel; không ghi thiết kế đã được duyệt hay Booking đã hoàn thành.

## 2026-10-11 — Triển khai lịch phòng tuần và ngoại lệ

- Thêm migration v3 và bốn bảng lịch, nâng schema lên 18 bảng; giữ phiên/ID/FK legacy và chuyển đổi gate có chủ đích.
- Thêm chín operations lịch phòng (tổng 58), preview/xác nhận, version/revision, audit transaction, chống holds/overlap và replay idempotency sau restart.
- Thêm grid 7 × 24 kéo chọn/di chuyển/resize, form và bàn phím, undo/redo, ngoại lệ, lịch thực tế/lịch sử và trang chỉ đọc theo quyền; sinh phiên 28 ngày và job startup/00:05 Việt Nam.
- Go/DB/API tests, vet và 26 frontend tests/build PASS; demo tạo/phát hành lịch vào DB thật. Giới hạn browser do hạn mức, Postman/NFR và các bước chưa chạy ghi riêng trong `reports/room_calendar_verification_2026-10-11.md`.
- Cập nhật quyết định/SRS/UC/ERD/API/runbook/backlog; giữ DeepSeek Harness và Excel, không ghi họp/phỏng vấn/phê duyệt hoặc ngày hoàn thành của sinh viên.

## 2026-10-11 — Kế hoạch lịch hoạt động phòng 7 × 24

- Ghi yêu cầu Admin/Manager đăng ký giờ phòng bằng bảng giống Excel; người dùng chọn lịch tuần lặp + ngoại lệ ngày.
- Thêm kế hoạch 20: kéo chọn/di chuyển/resize, form thay thế, preview/xác nhận, version/revision, sinh phiên, bảo vệ booking, chuyển đổi legacy và kiểm thử theo chặng.
- Liên kết vào kế hoạch 18/19 và ghi yêu cầu bổ sung ở decision log. Chưa sửa code ứng dụng, schema/migration/API hoặc dữ liệu DB; không đổi Excel hay báo lịch mới đã hoàn thành.

## 2026-10-11 — Triển khai CRUD quản trị và UI/UX

- Thêm backend/API/UI Hội viên, Phòng, Ca, Phiên phòng và PT; mở rộng chi tiết/sửa tài khoản Admin; giữ phân chia package theo project Gin được yêu cầu.
- Thay UI cũ bằng route/layout/guard, token thiết kế, Radix dialog/action menu, form riêng và bảng/card responsive; thêm xử lý ngày/giờ native, form chưa lưu, submit lặp và phiên hết hạn.
- Thêm constraint chồng giờ PT bằng migration v2; nghiệp vụ/audit trong transaction, kiểm tra booking giữ chỗ, file riêng tư và token revoke.
- Thêm 49 operations OpenAPI/Postman hiện hành, script kiểm tra router và runbook PowerShell 5.1; làm rõ D05 theo UC-12.
- Kiểm thử và giới hạn ghi tại `reports/admin_crud_verification_2026-10-11.md`. Không sửa Excel, không ghi phỏng vấn/họp/phê duyệt nghiệp vụ thực tế.

## 2026-10-11 — Kế hoạch CRUD quản trị và UI/UX

- Thêm `19_ke_hoach_crud_admin_ui_ux.md` theo phạm vi người dùng chọn: Tài khoản, Hội viên, Phòng, Ca/Phiên phòng và PT, chia chặng.
- Đề xuất nền sáng, sidebar xanh đậm, accent cam, component/form/dialog/table thống nhất và responsive; ghi rõ các API còn thiếu, quyền và cổng nghiệm thu.
- Áp dụng skill `ui-ux-pro-max` và `ui-styling` để lập kế hoạch; chưa sửa code ứng dụng hoặc cài thư viện.

## 2026-10-11 — Kế hoạch sau tuần 4

- Thêm `18_ke_hoach_lam_viec_tiep.md`: chốt phần kiểm chứng Identity còn thiếu, dữ liệu nền tuần 5 và lộ trình Booking → Operations → UI → Reports → bàn giao.
- Đề xuất điều chỉnh thứ tự theo phụ thuộc vì API Members/Rooms/TimeSlots/PT chưa có; giữ nguyên lịch và phân công trong Excel khi lập kế hoạch.
- Cập nhật backlog và mục lục theo bằng chứng triển khai; không ghi các module tương lai đã hoàn thành.

## 2026-10-11 — Identity demo tuần 4

- Thêm migration PostgreSQL versioned, seed tài khoản demo và Docker Compose dùng cổng 5433 riêng.
- Thêm Identity/Auth: bcrypt, JWT HS256, `auth_version`, RBAC Admin, audit và API quản lý tài khoản.
- Thêm giao diện React/Vite/Tailwind cho đăng nhập, hồ sơ cá nhân và quản lý tài khoản Admin.
- Thêm OpenAPI, Postman collection và hướng dẫn chạy local.
- Chuẩn hóa backend theo package của project Gin tham chiếu: `app`, `db`, `dto/v1`, `handler/v1`, `middleware`, `repository`, `routes/v1`, `service/v1` và `pkg/auth`.
- Bắt buộc JWT có `exp`; chặn rollback/reset migration ngoài đúng DB `ttcsn_gym_test`; theo dõi template Compose nhưng bỏ qua cấu hình local.
- Kiểm tra trước commit: Go tests/vet, PostgreSQL integration, frontend test/build và checker tài liệu PASS; chi tiết tại `reports/week4_verification_2026-10-11.md`.

## v0.3 — 10/10/2026

### Baseline mới

- Chốt D01–D10 và BR-01…BR-26; xác định rõ phạm vi một cơ sở và các tính năng ngoài MVP.
- Chuẩn hóa 5 tác nhân, 16 use case và một ma trận quyền có object scope.
- Tạo SRS/backlog v0.3; giữ v0.1/v0.2 làm lịch sử.
- Chuẩn hóa vòng đời booking: PENDING_PT, CONFIRMED, REJECTED, EXPIRED, CANCELLED, COMPLETED, NO_SHOW.
- Tách phiên phòng, attendance từng người, review theo booking, idempotency và audit.
- Thay schema reset phá dữ liệu bằng bootstrap schema không có DROP; bổ sung runbook migration/restore.
- Chọn modular monolith và định nghĩa ranh giới handler/use case/repository/transaction.
- Bổ sung hợp đồng API, error code, kế hoạch kiểm thử và cổng nghiệm thu.

### Sửa mâu thuẫn

- Đồng bộ 5 vai trò giữa khảo sát, tác nhân và quyền.
- Sửa ID UC/REQ bằng danh mục UC-01…UC-16 duy nhất.
- Tách trạng thái tài khoản, hồ sơ hội viên, PT và booking.
- Bỏ doanh thu/thanh toán khỏi MVP vì chưa có nguồn giao dịch.
- Không dùng trạng thái booking thay attendance; không cho review thiếu booking hợp lệ.
- Hạ các tuyên bố hiệu năng/tính tuyệt đối thành mục tiêu cần kiểm thử.

### Giới hạn

- Chưa chạy schema trên PostgreSQL 16+.
- Chưa có migration thực thi từ dữ liệu v0.2; runbook mới mô tả quy trình/mapping.
- Chưa có mã nguồn/OpenAPI/Postman/test report/benchmark/restore drill.
