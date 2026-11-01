# Kế hoạch làm việc tiếp theo — 25/10/2026

Ngày lập: **25/10/2026, Asia/Ho_Chi_Minh**. Trạng thái: **đề xuất triển khai, chưa thực thi các chặng dưới đây**.

**Bản sửa đang trình duyệt:** [29 — Dữ liệu demo và kéo chọn giờ](29_ke_hoach_demo_va_keo_chon_gio.md). Theo yêu cầu bổ sung ngày 25/10, ưu tiên mới là dữ liệu demo lưu bền và kéo chọn giờ ở cả lịch phòng/HLV lẫn Hội viên đặt ca. Thứ tự đợt đầu theo D0–D4 trong bản 29; các phần kiểm chứng mở rộng bên dưới tiếp tục sau đợt đó.

Mục tiêu đợt tiếp theo là hoàn thành kiểm chứng vòng đời GYM trên môi trường thật, sửa các lỗi tìm được và chuẩn bị bản demo local có bằng chứng. Phạm vi chi tiết tiếp tục theo [kế hoạch 26](26_ke_hoach_chot_phan_con_thieu.md); thứ tự công việc còn lại được cập nhật trong tài liệu này.

## 1. Điểm xuất phát đã đối chiếu

HEAD khi lập kế hoạch là `1226740` trên `main`. Có thay đổi cục bộ ở kế hoạch 25, thư mục `.agents/` và file `s`; giữ nguyên các phần đó khi triển khai.

| Hạng mục | Bằng chứng hiện có | Phần cần tiếp tục |
|---|---|---|
| Chức năng | Identity, CRUD, lịch phòng, Booking, Attendance/Completion, Review và Report đã có API/UI | Kiểm chứng luồng liên thông và các ngoại lệ trên browser |
| Regression | Báo cáo 18/10 ghi Go/PostgreSQL tests, vet, 45 frontend tests, build, 76 API operations và docs checker PASS | Chạy theo thay đổi và chốt lại trên bản bàn giao |
| API Runner | Role smoke 15/15; Identity mutation 14/14 và 4 audit PASS | Room/Calendar/Booking/Attendance/Review/Report mutation chưa có collection kiểm chứng đầy đủ |
| Browser | Smoke 5 vai trò; đã tạo/đọc lại booking tự tập synthetic; đo responsive một số trang | Luồng PT phản hồi → điểm danh → completion theo giờ thật → review → báo cáo; ma trận browser/keyboard/lỗi mạng |
| Hiệu năng | Burst login/read 100 worker: 400/400 thành công, p95 **2.435,10 ms**; warm-session read/ramp: 300/300, p95 **80,92 ms** | NFR-01 tổng thể chưa đạt; thiếu 100 danh tính, workload hỗn hợp giữ tải và số đo host/DB |
| Vận hành | Retention idempotency v6; expiration backlog 1.001 qua hai lượt; một backup/restore local khớp dữ liệu | Completion backlog, restart/timeout/shutdown; security mở rộng; phục hồi file private và số đo RPO/RTO |

Lượt kiểm tra ngày 25/10 không thấy listener tại `5433/8080/5173/18080/15173`; Docker Linux engine hiện không truy cập được. Đây là điều kiện môi trường cần xử lý khi bắt đầu chạy, chưa phải lỗi ứng dụng đã chẩn đoán.

Nguồn: [regression 18/10](reports/local_regression_2026-10-18.md), [vận hành/NFR/restore 18/10](reports/operations_nfr_restore_2026-10-18.md), [Identity mutation 18/10](reports/mutation_smoke_2026-10-18.md), [ma trận tiến độ](27_ma_tran_tien_do_va_kiem_chung_tuan_1_6.md). Các kết quả này là snapshot cũ đã đọc lại; lượt lập kế hoạch không chạy lại test ứng dụng.

## 2. Thứ tự công việc

| Chặng | Ưu tiên | Công việc chính | Đầu ra và cổng hoàn thành |
|---|---|---|---|
| S0 — Sẵn sàng môi trường | P0 | Khôi phục môi trường local, xác nhận DB/port, chuẩn bị dữ liệu synthetic và log | API/client test chạy được; DB đích và phiên bản schema được xác minh; tắt/khởi động lại đúng instance |
| S1 — API nghiệp vụ | P0 | Bổ sung collection theo vòng đời và đối chiếu DB/audit | Kịch bản chính và lỗi nghiệp vụ PASS; báo cáo Runner, trạng thái DB và danh sách lỗi |
| S2 — Browser đủ vòng đời | P0 | Hội viên đặt với PT → PT nhận → Lễ tân điểm danh → job hoàn thành → Hội viên review → Manager/PT đọc báo cáo | Thao tác UI thật, booking/attendance/review/report liên thông và khớp DB; có ảnh và ID đối chiếu |
| S3 — Ngoại lệ và UI QA | P0 | Từ chối/hết hạn/hủy, sai quyền, ca đầy, 401/lỗi mạng, responsive và keyboard | Ma trận kiểm tra có kết quả; không còn lỗi Critical/High mở trong phạm vi nghiệm thu |
| S4 — Tải và job | P1 | Đo nguyên nhân login chậm; workload 100 danh tính; completion backlog/restart | Báo cáo số đo và tính đúng; NFR chỉ được ghi đạt khi đủ tiêu chí |
| S5 — Bảo mật và khôi phục | P1 | Scope/upload/log/dependency; restore DB và file private, đo thời gian | Checklist bảo mật có chứng cứ, restore độc lập đọc được dữ liệu/file, công bố RPO/RTO đạt hoặc chưa đạt |
| S6 — Chốt bản demo | P1 | Regression cuối, đồng bộ tài liệu, hướng dẫn 5 vai trò và hồ sơ demo | Bản demo local chạy lại được, hướng dẫn/slide/báo cáo khớp commit và giới hạn thực tế |

Thực hiện **S0 → S1 → S2 → S3 → S4 → S5 → S6**. Sau mỗi chặng: ghi phát hiện → sửa lỗi trong phạm vi → chạy kiểm tra liên quan → lưu bằng chứng → commit theo chặng. Một lỗi làm sai dữ liệu hoặc phân quyền phải được xử lý trước khi dùng luồng đó để kiểm chứng bước sau.

## 3. S0 — Chuẩn bị để các lượt kiểm chứng chạy lặp lại được

1. Kiểm tra Docker Desktop, PostgreSQL, Go/Node/npm, migration và các dependency sẵn có. Dùng script dự án hiện có làm điểm bắt đầu; rà script trước khi chạy.
2. Chuẩn hóa cách chạy API test trên `127.0.0.1:18080`, client test dự kiến `127.0.0.1:15173`, DB `ttcsn_gym_test`. Cổng client phải được kiểm tra còn trống và proxy phải trỏ đúng API test. Script startup hiện chủ yếu phục vụ dev: bổ sung lựa chọn cấu hình test nếu cần, không thay mặc định dev hoặc ghi đè cấu hình riêng.
3. Xác minh tên DB từ kết nối thật trước seed/reset. Mỗi lượt dùng namespace synthetic; Go integration, Newman và browser dùng chung DB test phải chạy lần lượt để tránh reset dữ liệu của nhau.
4. Xác định dữ liệu cần cho luồng chính: 5 vai trò, PT được duyệt, hội viên/thẻ hợp lệ, phòng có ảnh, ca/phiên phù hợp và lịch PT. Chuẩn bị thêm dữ liệu riêng cho ca đầy, PT bận và hồ sơ/thẻ không hợp lệ.
5. Lưu request ID và lỗi đã che dữ liệu nhạy cảm. Theo dõi lại lần tải lịch PT từng trả 500 trong báo cáo 18/10; ghi account synthetic, filter và request ID nếu tái diễn. Lần cũ không tái hiện được nên chưa gán nguyên nhân.

Hoàn thành khi có lệnh khởi động/dừng kiểm chứng, readiness, schema version và xác nhận DB đích. Kiểm tra lại các dependency đã có trước khi cài thêm. Giữ DeepSeek Harness chính thức là harness AI duy nhất của dự án; kiểm tra giao diện/model probe nếu dùng DSH trong đợt này.

## 4. S1 — Mở rộng API Runner từ Identity sang nghiệp vụ

Tái sử dụng runner/collection sẵn có; thêm scenario có thứ tự, tự lấy token/ID và dùng dữ liệu synthetic trên DB test.

| Nhóm scenario | Assertions và đối chiếu bắt buộc |
|---|---|
| Room/Calendar | Tạo phòng và cover; lịch tuần/ngoại lệ; preview/confirm; kiểm tra phiên được sinh và audit |
| Booking/PT | Tự tập/có PT; accept/reject/expire; hold được giữ/giải phóng đúng; hủy theo quyền và mốc giờ |
| Idempotency | Replay trong thời hạn không tạo booking/audit mới; key hết hạn tuân theo hành vi v6 và quy tắc nghiệp vụ đã ghi |
| Attendance/Completion | Ghi MEMBER/PT riêng; replay; chặn sớm/muộn/sai quyền; trạng thái cuối và audit đúng |
| Review/Report | Đủ điều kiện, chỉ đánh giá một lần; projection không lộ contact; scope Admin/Manager/PT và số liệu khớp SQL |

Đối chiếu source v6 với decision log/API/runbook về key hết hạn trước khi ghi kỳ vọng vào test. Nếu tài liệu lệch code thì ghi rõ và đồng bộ theo quyết định nghiệp vụ có căn cứ.

Runner giữ assertions và trạng thái DB trước cleanup trong báo cáo đã che secret. Test race Go/PostgreSQL hiện có tiếp tục kiểm tra tính đúng đồng thời. Các ca biên thời gian chạy bằng fixture/backend test; chúng được ghi riêng với luồng dùng giờ thật ở S2.

## 5. S2 — Chứng minh một vòng đời trên browser

Luồng bắt buộc dùng cùng booking và dữ liệu DB xuyên suốt:

1. Admin/Manager chuẩn bị phòng, ảnh, lịch mở và phiên; Hội viên thấy lựa chọn tương ứng.
2. Hội viên chọn phòng → ca → PT và gửi yêu cầu; PT đọc và nhận lịch.
3. Hội viên/PT đọc cùng booking đã xác nhận; Lễ tân ghi attendance riêng cho hai người.
4. Chờ ca kết thúc theo đồng hồ thật, quan sát job chuyển `COMPLETED` và các bên đọc lại.
5. Hội viên đánh giá; PT đọc review/hiệu suất; Manager đối chiếu báo cáo với dữ liệu nguồn.

Chọn ca synthetic ngắn hợp lệ theo quy tắc hiện hành. Khảo sát trước độ dài ca, deadline PT và cửa sổ attendance để đặt lịch chạy phù hợp. Nếu không tạo được ca ngắn hợp lệ, bố trí giờ kiểm chứng theo ca hiện có; không đổi đồng hồ, quy tắc nghiệp vụ hoặc timestamp để ghi full browser PASS.

Đầu ra: report theo ngày chạy, browser/version, ID synthetic, ảnh các bước, kết quả DB/audit và thời điểm job. Ghi riêng bước dùng fixture/API hỗ trợ; chỉ công bố full browser flow cho các bước đã thao tác trên UI thật.

## 6. S3 — Đóng các nhánh lỗi và kiểm tra sử dụng

- Nghiệp vụ: PT từ chối/hết hạn, hủy đúng và sai mốc, tự tập, Member vắng/PT có mặt, ca đầy/PT bận, thẻ hết hạn và PT chưa được duyệt.
- Phiên/quyền: 401 khi đang mở form, revoke sau khóa/reset/đổi role, URL trực tiếp ngoài quyền, truy cập object người khác.
- Khả năng phục hồi: double submit, retry mạng và idempotency, loading/empty/error/retry; đổi phòng/ngày phải xóa lựa chọn cũ.
- Giao diện: màn hình chính của 5 vai trò tại 375/768/1024/1440px, zoom 200%, Tab/Enter/Escape, focus dialog và label/error. Chạy luồng trọng yếu trên Chrome/Edge/Firefox theo khả năng môi trường; chỗ chưa chạy giữ trạng thái chưa kiểm chứng.

Giữ giao diện hiện có và chỉ sửa các vấn đề tìm được. Giữ portal Hội viên xanh lá, PT xanh dương, ảnh phòng và cách điều hướng đã chốt. Mỗi lỗi có mức độ, bước tái hiện, bằng chứng sửa và kiểm tra lại.

## 7. S4 — Đo tải có dữ liệu đại diện và kiểm tra job

1. Xác định trước workload/dataset, không đổi tiêu chí sau khi nhìn kết quả. Dùng 100 tài khoản synthetic có điều kiện nghiệp vụ hợp lệ thay cho năm tài khoản seed dùng lặp.
2. Công bố số users/members/PT/rooms/sessions/bookings/reviews, cấu hình máy và tỷ lệ login/read/search/create/cancel/report. Cấu hình khởi điểm để thử: ramp 30 giây, giữ tải 5 phút; điều chỉnh có ghi lý do trước lượt đo chính.
3. Tách báo cáo login burst, warm-session read và workload hỗn hợp. Đo p50/p95/p99 theo nhóm endpoint, throughput, 4xx nghiệp vụ, 5xx, CPU/RAM, DB connections/lock wait và trạng thái dữ liệu sau tải.
4. Profile login/bcrypt, CPU, pool DB và truy vấn khi có số đo. Chọn thay đổi theo nguyên nhân; giữ chính sách bảo vệ mật khẩu. Cảnh báo bundle >500 kB xử lý khi có số đo cho thấy ảnh hưởng UI.
5. Kiểm tra completion backlog >1.000 qua nhiều lượt, restart/catch-up, timeout, lỗi một booking và shutdown; giữ fixture expiration 1.001 đã có làm regression. Đối chiếu capacity, hold, unique và số audit.

Cổng NFR-01/NFR-02: p95 API tương tác **<2 giây** với workload/dataset công bố, 100 người dùng đồng thời, không có 5xx ngoài lỗi tiêm chủ động và không vi phạm capacity/unique. Nếu chưa đạt thì lưu số đo, nguyên nhân đã chứng minh và việc còn mở; warm-session read không đóng được toàn bộ cổng.

## 8. S5 — Bảo mật và phục hồi

- Rà role/object scope, token revoke, trường input thừa, lỗi/log/secret và dependency. Với chứng chỉ PT: kiểm tra MIME/size, đường dẫn và quyền đọc file; ảnh phòng kiểm tra đúng room/scope.
- Tái sử dụng script backup/restore hiện có. Restore vào DB/thư mục mới và kiểm tra target trước khi ghi.
- Đối chiếu schema, FK/count/dữ liệu nguồn, ảnh phòng BYTEA và file chứng chỉ private; đọc file qua API với role được phép và bị cấm.
- Ghi timestamp snapshot, hash/size, thời điểm bắt đầu/kết thúc restore, thời gian tới khi API đọc dữ liệu thành công. Đánh giá riêng RPO <24h và RTO <2h; kiểm tra quy trình backup định kỳ/cảnh báo lỗi nếu công bố đáp ứng dài hạn.

Đầu ra là báo cáo có kết quả từng phép kiểm tra và rủi ro còn mở. Một restore DB thành công trước đây được dùng làm nền để mở rộng kiểm chứng.

## 9. S6 — Chốt hồ sơ và demo local

1. Chạy regression cuối trên bản source bàn giao: Go/PostgreSQL tests và vet; frontend tests/build; contract/docs checker; các scenario Runner và browser trọng yếu bị ảnh hưởng.
2. Đồng bộ backlog, ma trận REQ/UC/BR → test/report/commit và hướng dẫn local. Giữ các report cũ như snapshot có ngày.
3. Hoàn thiện hướng dẫn 5 vai trò bằng ảnh app thật, demo accounts synthetic, kịch bản demo đủ vòng đời, báo cáo tổng kết và slide.
4. Tổng duyệt từ lệnh khởi động đến luồng nghiệp vụ và dừng dự án; ghi các giới hạn còn lại trước đề xuất bản bàn giao/release.

Các hoạt động phỏng vấn, họp, review chéo, phê duyệt và ngày thực tế của thành viên do nhóm cung cấp bằng chứng. Khi cập nhật Excel, dùng thông tin đó và giữ form/phân công; không dùng ngày chạy test làm ngày hoàn thành của sinh viên.

## 10. Phạm vi đợt đầu và điều kiện kết thúc

**Đợt triển khai đầu tiên tập trung S0 → S1 → S2:** môi trường test sẵn sàng, collection nghiệp vụ chạy được và một vòng đời PT trên browser có chứng cứ. Những lỗi cản luồng được xử lý ngay trong đợt này. Kết quả đợt đầu là cơ sở ước lượng phần QA/tải còn lại.

REQ-15 xuất Excel/PDF tiếp tục ở mức Could, xem xét sau các cổng Must/Should. Thanh toán, QR/RFID, nhiều chi nhánh và triển khai production nằm ngoài đợt này.

- [ ] Môi trường test khởi động/dừng lặp lại được, đúng DB/port và có dữ liệu synthetic.
- [ ] API Runner nghiệp vụ có assertions và đối chiếu DB/audit.
- [ ] Browser đủ vòng đời PT → attendance → completion → review → report có bằng chứng.
- [ ] Ma trận UI/ngoại lệ/quyền có kết quả; không còn lỗi Critical/High mở trong phạm vi nghiệm thu.
- [ ] Load/job/security/restore có report; tiêu chí đạt và mục chưa đạt được công bố rõ.
- [ ] Regression cuối, hướng dẫn, báo cáo, slide và kịch bản demo khớp bản bàn giao.

Không gán thời hạn hoàn thành cố định trước khi chốt môi trường và các lỗi ở S1/S2. Tiến độ được cập nhật theo cổng hoàn thành và bằng chứng thực tế.
