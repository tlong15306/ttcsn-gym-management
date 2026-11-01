# Kế hoạch chốt các phần còn thiếu sau rà soát tuần 1–6

**Ngày lập:** 18/10/2026, Asia/Ho_Chi_Minh. **Trạng thái:** Kế hoạch triển khai, chưa phải kết quả nghiệm thu.

**Kế hoạch tiếp nối ngày 25/10/2026:** [28 — Kế hoạch làm việc tiếp theo](28_ke_hoach_lam_viec_tiep_2026_10_25.md), đã đối chiếu thêm các kết quả regression, Identity mutation, vận hành/NFR/restore ngày 18/10. Thứ tự công việc còn lại theo kế hoạch 28; nội dung dưới đây giữ mốc lập kế hoạch và phạm vi chi tiết ban đầu.

Mục tiêu là chốt những khoảng trống có bằng chứng sau đợt rà tuần 1–6, sau đó hoàn thành kiểm thử tích hợp và bàn giao. Phần chức năng đã có tiếp tục được dùng; ưu tiên tiếp theo là kiểm chứng, sửa lỗi và đồng bộ tài liệu.

## 1. Điểm xuất phát và cách hiểu tiến độ

- HEAD khi lập kế hoạch: `540016f`. Code đã có Identity, CRUD nền, lịch phòng tuần/ngoại lệ, ảnh phòng, Booking, Attendance/Completion, Review và Report API/UI.
- Bằng chứng gần nhất được lưu trong repo là các báo cáo ngày 11–14/10/2026 tại `reports/`. Chúng ghi Go/PostgreSQL tests, `go vet`, 45 frontend tests/build và contracts/docs checker PASS trong phạm vi tương ứng. Đây là kết quả đã ghi nhận, không phải lượt chạy mới ngày 18/10.
- Kết quả khởi động local hoặc API smoke trong cuộc trao đổi trước chưa thay cho báo cáo browser E2E đầy đủ. Trạng thái tiến trình phải kiểm tra lại khi bắt đầu triển khai.
- Excel gốc `../outputs_document/full_10_weeks/Ke_hoach_lam_viec_nhom_GYM.xlsx` vẫn giữ các sheet tuần 1–6 ở trạng thái Kế hoạch và ngày thực tế trống theo lần đọc trước. Phần chức năng đã code không tự chuyển thành ngày hoàn thành hoặc hoạt động thực tế của sinh viên.
- Tuần 6 theo Excel là nền frontend; lộ trình kỹ thuật điều chỉnh trong kế hoạch 18 đưa Booking vào tuần 6. Ma trận kiểm chứng phải ghi rõ đang đối chiếu nguồn nào.

| Mốc được rà | Phần đã có | Khoảng trống cần chốt |
|---|---|---|
| Tuần 1 | Khảo sát 3 sản phẩm, phạm vi, actors, backlog/use case và công nghệ | Rà nguồn cho các nhận định còn thiếu; chuẩn bị bộ xác nhận giả định. Phỏng vấn/họp/review nhóm chỉ ghi khi nhóm thực hiện và cung cấp chứng cứ. |
| Tuần 2 | SRS v0.3, quyền, ERD/schema, kiến trúc, API và kế hoạch test | Đồng bộ trạng thái tài liệu với code/báo cáo mới; rà trace REQ → UC → BR → quyền → test; ghi riêng việc phê duyệt thực tế còn thiếu. |
| Tuần 3–4 | Nền backend và Identity đã có báo cáo kiểm thử | Chạy lại regression trên bản cuối khi có sửa code; kiểm tra quy trình khởi động từ hướng dẫn. |
| Tuần 5 | Booking, race, attendance và review đã có code/tests; collection đã sinh | Thực thi các kịch bản Postman Runner có kiểm soát, lưu kết quả và chứng cứ DB. |
| Tuần 6 | React/Vite/Tailwind/Router, auth context, Bearer/401, login, guards/layout và báo cáo vận hành | Browser kiểm tra kết nối thật, phân quyền, trạng thái lỗi và responsive; chốt khác biệt giữa checklist cũ và cách triển khai hiện hành. |
| Phần vượt mốc tuần 6 | Portal Hội viên/PT riêng, review và reports | Demo đầy đủ UI → API → PostgreSQL, đo tải, security checks và backup/restore drill. |

Giữ Go `handler/v1 → service/v1 → repository`, model nghiệp vụ ở `domain`, DTO ở `dto/v1`; DeepSeek Harness hiện có; PostgreSQL local cổng 5433 và dữ liệu dev. Frontend tiếp tục dùng fetch wrapper, React Router hiện có, token trong memory và yêu cầu đăng nhập lại khi hết phiên. Dashboard Admin/Manager dùng `/management/reports`; Hội viên/PT giữ portal riêng theo yêu cầu đã chốt. Các lựa chọn này cần được mô tả đúng trong checklist tuần 6, không coi khác tên thư viện là lỗi chức năng.

## 2. Thứ tự triển khai và đầu ra

| Chặng | Ưu tiên | Công việc | Cổng hoàn thành |
|---|---|---|---|
| R0 — Chốt nguồn tiến độ | P0 | Ma trận thiếu sót và đồng bộ trạng thái tài liệu | Mỗi việc có nguồn yêu cầu, bằng chứng hiện có, phần chưa kiểm chứng và bước xử lý |
| R1 — Chạy local lặp lại được | P0 | Chuẩn PowerShell 5.1, startup/stop và regression nền | Chạy từ hướng dẫn thành công, dùng đúng DB, không lộ secret hoặc ảnh hưởng tiến trình khác |
| R2 — Kiểm chứng API/collection | P0 | Kịch bản Postman Runner cho 5 vai trò và luồng quan trọng | Assertions PASS, ID/token tự lấy, số liệu DB/audit đúng, có report |
| R3 — Browser E2E và UI/UX | P0 | Luồng nghiệp vụ đủ vòng đời, responsive, keyboard và sửa lỗi | UI ghi/đọc DB thật, đúng quyền, không còn lỗi Critical/High mở |
| R4 — Vận hành và NFR | P1 | Job/backlog, idempotency policy, tải, security, backup/restore | Báo cáo có workload/dataset/số đo và restore thực tế; giới hạn ghi rõ |
| R5 — Bàn giao | P1 | Hướng dẫn 5 vai trò, báo cáo, slide, demo và tổng hợp tiến độ | Hướng dẫn chạy lại được; hồ sơ khớp bản kiểm chứng; release sau nghiệm thu |

Thực hiện `R0 → R1 → R2 → R3 → R4 → R5`. Trong mỗi chặng: kiểm tra → ghi lỗi → sửa → regression liên quan → lưu bằng chứng → commit theo phạm vi. Phần xác nhận nghiệp vụ của nhóm ở mục 8 có thể tiến hành trong khi Codex triển khai các chặng kỹ thuật.

## 3. R0 — Đồng bộ tài liệu và danh sách thiếu sót

1. Lập ma trận tuần 1–6: mã task Excel, REQ/UC/BR liên quan, file/code/test/report, commit, trạng thái `đã có code` / `đã kiểm thử` / `chưa kiểm chứng` / `cần nhóm xác nhận`.
2. Cập nhật `13_backlog_v0.3.md`: Attendance, Review, Metrics/Report hiện vẫn có những dòng ghi Đã phân tích dù đã có code và báo cáo 14/10. Giữ riêng giới hạn E2E/NFR còn thiếu.
3. Rà `docs/README.md`, kế hoạch 18/24/25, đầu mục tiến độ trong SRS và nhật ký. Giữ báo cáo cũ là snapshot có ngày; thêm dẫn chiếu bản mới thay vì viết lại kết quả quá khứ. Nhật ký chỉ bổ sung sự kiện kỹ thuật có thật, không viết thay mức độ hiểu bài của sinh viên.
4. Rà khảo sát 3 sản phẩm: nhận định nào chưa có nguồn phù hợp thì bổ sung nguồn hoặc chuyển thành nhận định/giả định có nhãn. Tách nghiên cứu sản phẩm với phỏng vấn người dùng.
5. Rà contract generator và trace matrix theo 78 operations hiện có; endpoint chưa triển khai phải được ghi rõ.

**Đầu ra:** ma trận kiểm chứng hiện hành, backlog/mục lục thống nhất, danh sách lỗi có mức độ và cách tái hiện. `docs/check_docs.ps1` và contract checker PASS chỉ xác nhận phần chúng thực sự kiểm tra.

## 4. R1 — Quy trình chạy dự án từ PowerShell

- Chuẩn hóa startup/stop thành script dự án nếu thao tác thủ công còn dễ sai: đọc cấu hình theo tên biến, URL-encode mật khẩu DB, tạo secret bằng API tương thích PowerShell 5.1 và lưu log local bị Git bỏ qua.
- Kiểm tra Docker engine, PostgreSQL healthy, migration up, seed chạy lặp giữ mật khẩu cũ, API readiness và frontend/proxy. Tái sử dụng instance dự án đã healthy; kiểm tra PID/đường dẫn trước khi quản lý tiến trình. Helper nền chạy ẩn.
- Lệnh dừng chỉ tác động instance dự án đã xác minh; PostgreSQL dừng bằng Compose và giữ volume. Kịch bản reset/rollback chỉ nhận đúng `ttcsn_gym_test`.
- Rà phiên bản thực Go/Node/npm/Docker/PostgreSQL, biến môi trường bắt buộc và hướng dẫn xử lý các lỗi Docker daemon, cổng bận, cache Go, `npm.ps1` và 401.
- Chạy baseline cuối: Go tests với PostgreSQL test thật, `go vet`, frontend tests/build, contract/docs checker. Nếu source không đổi, dùng kết quả này làm mốc trước các sửa lỗi tiếp theo.

**Đầu ra:** script/hướng dẫn PowerShell chạy lại được, log startup và report phiên bản/migration. Smoke khởi động được ghi riêng với luồng nghiệp vụ.

## 5. R2 — Postman Runner và luồng API

Collection tổng hiện có các thao tác khóa/reset/logout; tạo thư mục kịch bản kiểm chứng có thứ tự, tự lấy token/ID và assertions thay vì chạy mù toàn bộ collection.

| Kịch bản | Kiểm tra bắt buộc |
|---|---|
| Identity 5 vai trò | Login/me, hồ sơ SELF, non-Admin 403, `/me` không đổi role/status, logout và revoke sau lock/reset/role |
| Phòng/lịch/PT | Cover đúng phòng, giờ tuần + ngoại lệ, preview/xác nhận, ca/phiên/capacity, PT hợp lệ và file chứng chỉ đúng quyền |
| Booking | Tự tập/có PT, pending → accept/reject/expire, hủy đúng quyền/mốc giờ, retry cùng key không tạo booking mới |
| Attendance/Review | Điểm danh từng người, completion/no-show, review đúng điều kiện/một lần, projection công khai không lộ contact |
| Reports | Admin/Manager ALL, PT SELF, Member/Lễ tân 403, filter/kỳ Việt Nam, N/A và số liệu đối chiếu SQL |

- Dữ liệu ghi dùng namespace synthetic theo từng lần chạy, trên DB test của dự án. Xác minh DB đích trước khi setup/reset; không chạy đồng thời suite khác đang reset cùng DB.
- Tách scenario destructive của Identity với actor vận hành chính để khóa/logout không làm hỏng các bước tiếp theo. Token còn hạn tại bước sử dụng; log/report không chứa token/mật khẩu.
- Chạy test concurrency bằng Go/PostgreSQL hiện có; Postman Runner bổ sung bằng chứng collection thực thi, không thay cho test race.

**Đầu ra:** report runner/assertions, danh sách ID synthetic để đối chiếu, truy vấn kiểm tra hold/audit và report lỗi nếu có. Cập nhật runbook theo collection thực sự chạy được.

## 6. R3 — Browser E2E, giao diện và sửa lỗi

### Luồng chính cần chứng minh

1. Admin/Manager tạo phòng có ảnh → khai báo giờ tuần/ngoại lệ → preview/xác nhận → kiểm tra phiên được sinh.
2. Hội viên chọn ảnh phòng → ngày/ca → PT hợp lệ → xác nhận đặt. Chọn ảnh hoặc xem options chưa tạo hold; đổi phòng xóa lựa chọn ca/PT cũ.
3. PT xem yêu cầu, nhận lịch; Hội viên/PT xem đúng booking của mình. Kiểm tra thêm từ chối, hết hạn và hủy trước mốc quy định.
4. Lễ tân tìm booking và ghi attendance riêng Hội viên/PT; cả hai đọc lại kết quả. Member/PT truy cập trực tiếp chức năng vận hành vẫn bị chặn.
5. Sau khi ca hết giờ, job chuyển trạng thái; Hội viên đánh giá buổi đủ điều kiện; PT đọc review và hiệu suất; Manager đối chiếu báo cáo với booking/attendance/review nguồn.
6. Luồng phụ: tự tập, Member vắng/PT đến, ca đầy, PT bận, hồ sơ/thẻ không hợp lệ, tài khoản bị khóa, 401 giữa thao tác và retry sau lỗi mạng.

Chuẩn bị dữ liệu riêng trên DB test và API/client kiểm chứng dùng cổng/proxy rõ ràng để tách khỏi instance dev. Tạo ca synthetic ngắn sắp bắt đầu bằng API nếu quy tắc hiện hành cho phép; dùng giờ thật và chờ job xử lý trước khi review. Không sửa timestamp/attendance của dev để giả hoàn thành. Ca biên thời gian được kiểm tra bằng fixtures backend riêng; chỉ ghi full browser E2E PASS khi tất cả bước UI của luồng đã thực sự chạy.

### QA giao diện

- Viewport 375/768/1024/1440px, desktop zoom 200%; kiểm tra các trình duyệt nêu trong NFR-04 theo khả năng môi trường và ghi rõ browser/version đã dùng.
- Đăng nhập và landing/điều hướng đủ 5 vai trò; token memory, reload, TTL/401; chặn URL trực tiếp ngoài quyền.
- Header/sidebar/drawer, ảnh thật/placeholder, lịch phòng 7 × 24, chọn ca/PT, bảng/card/phân trang, form/dialog và trang báo cáo. Giới hạn cuộn ngang trong vùng lịch khi cần; toàn trang không tràn ở 375px.
- Keyboard/Tab/Enter/Escape, focus sau đóng dialog, label/error, vùng chạm, loading/empty/error/retry và xác nhận thao tác nhạy cảm.
- Dùng frontend-design và hệ thống UI hiện có khi cần chỉnh màn hình; giữ portal Hội viên xanh lá, PT xanh dương và ảnh phòng đã chốt. Chỉ thêm dashboard/chỉ số khi có nhu cầu và nguồn dữ liệu thật.

**Đầu ra:** ma trận browser/viewport, ảnh chụp app thật và kết quả từng kịch bản; lỗi Critical/High được sửa và kiểm tra lại trước bàn giao. Thêm regression test khi lỗi có nguy cơ lặp lại; component mocks được ghi riêng với browser/live DB.

## 7. R4 — Vận hành, tải, bảo mật và khôi phục

### Job và dữ liệu retry

- Hiện có test backlog 105 bản ghi và giới hạn 1.000 ứng viên mỗi lượt ở expiration/completion. Kiểm tra backlog vượt 1.000, nhiều chu kỳ, restart/catch-up, lỗi một bản ghi, timeout và shutdown; kiểm tra dữ liệu cuối, hold và số audit.
- Rà quy tắc `expires_at` của idempotency: hiện ghi 24 giờ nhưng chưa có cleanup. Chốt ý nghĩa replay/retention cho Booking và lịch phòng trước khi thêm xóa dữ liệu; test retry cũ không âm thầm tạo thao tác trùng. Giới hạn giữ key hiện có được ghi trong runbook cho đến khi có chính sách và kiểm chứng thay thế.

### Tải và security

- Tạo dataset test có kích thước công bố; chốt mix login/read/search/create/cancel/report, ramp-up, thời gian giữ tải và điều kiện nghiệp vụ trước khi chạy.
- Theo SRS: đo workload 100 người dùng đồng thời, mục tiêu p95 API tương tác <2 giây. Thu p50/p95/p99, throughput, 4xx nghiệp vụ, 5xx, CPU/RAM, DB connections/lock wait. Đối chiếu capacity/unique/audit sau tải.
- Nếu chưa đạt, dùng kết quả truy vấn/bundle để chọn chỗ sửa; cảnh báo main chunk >500 kB là việc tối ưu sau lỗi chức năng, phải đo lại sau chỉnh.
- Rà token/role/object scope, CORS allowlist, upload MIME/size/private certificate, input không mong đợi và log/error/secret. Kết quả local không suy thành chứng nhận bảo mật production.

### Backup/restore

- Backup PostgreSQL và các file upload riêng tư cần cho chứng chỉ; ảnh phòng BYTEA nằm trong DB. Có manifest, hash/size/time và phiên bản schema.
- Restore vào database/thư mục mới được đặt tên, xác minh độc lập, rồi kiểm tra FK/count/dữ liệu nguồn, ảnh phòng và tải chứng chỉ đúng quyền. Scripts phải kiểm tra target trước khi ghi; không restore đè dev hoặc `master-golang`.
- Đo thời gian phục hồi và tuổi snapshot theo mục tiêu NFR-05 (RPO <24h, RTO <2h); báo rõ môi trường và phạm vi drill. Việc backup hằng ngày và phát hiện thất bại phải có hướng dẫn/kiểm chứng riêng nếu công bố đáp ứng mục tiêu dài hạn.

**Đầu ra:** báo cáo job/load/security/restore có dataset, command, thời điểm và số đo; các giới hạn không đạt được giữ trong danh sách mở.

## 8. R5 — Hồ sơ và phần nhóm phải xác nhận

### Codex chuẩn bị và hoàn thiện

- Bộ câu hỏi khảo sát thực tế, bảng giả định nghiệp vụ và checklist review tuần 1–2 để nhóm sử dụng; nguồn so sánh sản phẩm được dẫn rõ.
- Hướng dẫn 5 vai trò bằng ảnh app thật, lệnh PowerShell, demo accounts, tình huống 401/403/409, cách dừng và xử lý lỗi thường gặp.
- Báo cáo tổng kết kỹ thuật, kết quả đạt/chưa đạt; slide và kịch bản demo đủ vòng đời; đối chiếu cuối REQ/UC/BR → test/report/commit.
- Tổng hợp tiến độ theo bằng chứng. Khi cập nhật Excel, giữ layout/lịch/phân công; ngày thực tế của sinh viên chỉ điền từ thông tin có chứng cứ của nhóm. Không thay ngày làm việc nhóm bằng ngày Codex chạy test.
- Chạy bộ kiểm tra cuối phù hợp thay đổi; commit theo chặng. Đề xuất release/tag khi nghiệm thu local đạt; đích hiện hành là demo local.

### Nhóm thực hiện và cung cấp kết quả thật

- Khảo sát/phỏng vấn tại cơ sở nếu học phần yêu cầu; xác nhận giả định, người trả lời và thời điểm thực tế.
- Review chéo, xác nhận phạm vi/tài liệu và kết quả bàn giao; ghi biên bản, người tham gia và ý kiến có thật.
- Xác nhận phần việc/ngày thực tế của từng thành viên và tổng duyệt bảo vệ. Các mục này có trạng thái độc lập với việc phần mềm đã chạy/test PASS.

### Phần ưu tiên sau các cổng bắt buộc

REQ-15 xuất Excel/PDF là Could: chỉ triển khai sau khi Must/Should và các cổng trên ổn định. Dữ liệu xuất phải giữ filter/kỳ/scope và không đưa contact trái quyền hoặc doanh thu ngoài phạm vi vào báo cáo. Thanh toán, QR/RFID, nhiều chi nhánh và production tiếp tục là phạm vi khác.

## 9. Điều kiện kết thúc kế hoạch

- [ ] Backlog và ma trận tiến độ khớp source/test/report; các hoạt động thực tế của nhóm được ghi đúng chứng cứ.
- [ ] Hướng dẫn/script local chạy lại được; collection kịch bản có kết quả Runner.
- [ ] Full browser flow 5 vai trò và QA mobile/desktop có chứng cứ; không còn lỗi Critical/High mở trong phạm vi nghiệm thu.
- [ ] Go/DB tests, vet, frontend tests/build, contracts và docs checker PASS trên bản cuối.
- [ ] Có báo cáo tải/security/job và backup/restore; mục chưa đạt được công bố rõ.
- [ ] Hướng dẫn, báo cáo, slide và demo khớp bản bàn giao; ghi riêng các Could bị hoãn.

**Chặng bắt đầu khi triển khai:** R0 đồng bộ tiến độ và ma trận thiếu sót, tiếp theo R1 chuẩn chạy local và R2 collection. Lượt lập kế hoạch này chỉ thêm tài liệu kế hoạch; kết quả thực thi sẽ được ghi trong report mới theo ngày chạy thật.
