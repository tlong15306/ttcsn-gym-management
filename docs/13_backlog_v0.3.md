# Product Backlog v0.3

**Tiến độ và thứ tự hoàn thiện hiện hành:** [30 — Đối chiếu ngày 31/10/2026](30_tien_do_va_ke_hoach_hoan_thien_2026_10_31.md). Các trạng thái chi tiết dưới đây giữ mốc cập nhật 18/10; phần demo/grid/batch bổ sung và kết quả kiểm tra mới được theo dõi ở tài liệu 30.

Baseline phân tích: **v0.3**. Tiến độ triển khai cập nhật **18/10/2026** theo code và báo cáo kiểm tra. “Đã phân tích” không đồng nghĩa “Đã triển khai”; test một phần không đồng nghĩa nghiệm thu đầy đủ. Ma trận trạng thái theo bằng chứng: `27_ma_tran_tien_do_va_kiem_chung_tuan_1_6.md`. Bằng chứng Booking/ảnh/portal: `reports/booking_verification_2026-10-12.md`; browser QA còn chưa kiểm chứng.

## 1. Backlog chức năng

| ID | Hạng mục | Ưu tiên | UC | Phụ thuộc chính | Trạng thái |
|---|---|---:|---|---|---|
| REQ-01 | Xác thực phiên | Must | UC-07 | users, auth_version, NFR-03 | Đã triển khai; kiểm thử token/revoke local PASS |
| REQ-02 | Phân quyền role + object scope | Must | Mọi UC | REQ-01, matrix quyền | Identity, Member/PT, Booking, Attendance, Review và Report scope đã có; API/DB tests local PASS, browser E2E còn kiểm chứng |
| REQ-03 | Quản lý hội viên | Must | UC-09 | REQ-01/02 | API/UI đã triển khai; transaction/scope local PASS |
| REQ-04 | Quản lý phòng | Must | UC-10 | REQ-01/02 | API/UI + lịch tuần/ngoại lệ đã có; DB/API PASS, UI ngoại lệ/responsive còn kiểm chứng |
| REQ-05 | Quản lý ca và phiên phòng | Must | UC-11 | REQ-04, BR-06…10 | API/UI + sinh phiên theo lịch đã có; overlap/holds/job lặp/snapshot local PASS |
| REQ-06 | Tạo booking/PT tùy chọn | Must | UC-02 | REQ-03/05/08/16/17 | API/UI đã có; PostgreSQL/idempotency/đặt hộ PASS; browser còn kiểm chứng |
| REQ-07 | Hủy booking | Must | UC-06 | REQ-06/08 | API/UI đã có; mốc 2h/lý do/audit rollback PASS |
| REQ-08 | Chống xung đột/quá tải | Must | Include | REQ-05/06, transaction | Last-seat 20 người và cross-room/slot guard PASS; toàn bộ tổ hợp race còn kiểm chứng |
| REQ-09 | Điểm danh từng người | Must | UC-04 | REQ-06/07 | API/UI đã triển khai; attendance riêng MEMBER/PT, completion/no-show, scope và PostgreSQL tests local PASS; browser E2E còn kiểm chứng |
| REQ-10 | Báo cáo vận hành | Should | UC-14 | REQ-06/09/19/20 | API/UI đã triển khai; metric nguồn Booking/Attendance/Review, scope và tests local PASS; browser E2E và benchmark còn kiểm chứng |
| REQ-11 | Tìm kiếm/lọc hội viên | Should | UC-09 | REQ-03 | API/UI search/filter/page đã triển khai |
| REQ-12 | Quản lý tài khoản | Must | UC-08 | REQ-01/02 | API/UI đã triển khai; Admin/RBAC/revoke và UI local PASS |
| REQ-13 | Xem lịch | Should | UC-13 | REQ-06/07/18 | API scope và UI Member/PT tuần/danh sách đã có; browser còn kiểm chứng |
| REQ-14 | Phản hồi và polling yêu cầu | Should | UC-16 | REQ-06/18 | UI polling 15s/countdown server_time đã có; browser còn kiểm chứng |
| REQ-15 | Xuất PDF/Excel | Could | UC-15 | REQ-10 | Đã phân tích, có thể hoãn |
| REQ-16 | Nộp hồ sơ PT | Must | UC-12 | REQ-01/12, file policy | API/UI profile/upload/submit đã triển khai; file policy local PASS |
| REQ-17 | Duyệt/đình chỉ PT | Must | UC-01 | REQ-16/02 | API/UI decision/schedule đã triển khai; state/audit local PASS |
| REQ-18 | PT phản hồi lịch | Must | UC-03 | REQ-06/14/17 | API/UI đã có; accept, trạng thái cuối, scope/expiry PASS |
| REQ-19 | Đánh giá PT | Should | UC-05 | REQ-09/18 | API/UI đã triển khai; eligibility, privacy, scope và PostgreSQL tests local PASS; browser E2E còn kiểm chứng |
| REQ-20 | Rating/chuyên cần PT | Should | UC-05/14 | REQ-09/19 | API/UI đã triển khai; rating và metric PT/report với numerator/denominator local PASS; browser E2E và benchmark còn kiểm chứng |

## 2. Thứ tự triển khai theo phụ thuộc

1. Identity: REQ-01, 02, 12 và audit nền tảng.
2. Master data: REQ-03, 04, 05, 11, 16, 17.
3. Booking core: REQ-06, 08, 07, 18, 13, 14.
4. Operations: REQ-09, 19, 20.
5. Reporting: REQ-10, sau đó REQ-15 nếu còn thời gian.

Không triển khai REQ-06 trước khi schema phiên phòng, trạng thái booking, unique index và giao thức transaction được kiểm thử. REQ-09 là Must vì REQ-19/20 phụ thuộc dữ liệu attendance.

Kết quả hiện hành và phần còn thiếu: `27_ma_tran_tien_do_va_kiem_chung_tuan_1_6.md`; snapshot tuần 4 được giữ trong báo cáo cũ. Kế hoạch thực thi kế tiếp: `26_ke_hoach_chot_phan_con_thieu.md`. Những kết quả local không thay cho benchmark NFR, kiểm tra mọi browser hoặc nghiệm thu production.

## 3. Backlog phi chức năng

| ID | Hạng mục | Cổng nghiệm thu | Trạng thái |
|---|---|---|---|
| NFR-01 | p95 API <2s theo workload công bố | Báo cáo benchmark có cấu hình/dữ liệu/percentile | Login/read burst 100 concurrent p95 2.435,10 ms, chưa đạt; warm-session read/ramp p95 80,92 ms nhưng không thay kết quả workload tổng thể. |
| NFR-02 | 100 concurrent users + tính đúng race | Load report và race assertions | Burst 100 concurrent chỉ đọc trả 400/400; warm-session read/ramp trả 300/300. Race correctness có test PostgreSQL nhưng chưa cùng workload, nên chưa đạt toàn bộ. |
| NFR-03 | Bảo mật auth/authz/upload/log | Test quyền, token revoke, file, secret scan | Auth/revoke/scope/upload local PASS; chưa đánh giá production |
| NFR-04 | Responsive/usability/accessibility cơ bản | Browser matrix + checklist thao tác | Component tests/build và ảnh phác UI có; browser matrix 5 vai trò, keyboard và accessibility toàn diện chưa chạy |
| NFR-05 | Backup/restore RPO/RTO | Restore drill có timestamp và kiểm tra dữ liệu | Một restore local sang DB mới khớp schema/count/cover; chưa chứng minh backup hằng ngày, RPO/RTO. |

## 4. Ngoài phạm vi / phiên bản sau

Thanh toán và doanh thu; nhiều cơ sở; đa vai trò/tài khoản; lớp học nhóm; lịch ca linh hoạt chồng giờ; app native; QR/RFID; email/SMS/WebSocket; notification lưu bền; quyền động; Trust Score tổng hợp.

Mỗi mục chỉ chuyển vào MVP qua decision log, cập nhật SRS/ERD/API/test và ước lượng lại tiến độ.

## 5. Bổ sung UI/UX ngày 31/10/2026

Theo decision log và kế hoạch 35: U0 mockup đã có; U1 trang chủ/khám phá/lịch 7 ngày công khai đã triển khai bằng projection API riêng, giữ xác thực đặt lịch và dữ liệu riêng, trả lại phòng/ngày sau login. U2 grid kéo chỉnh, U3 bản nháp/xác nhận và U4 tổng duyệt vẫn là các chặng sau. Bằng chứng U1: `reports/public_discovery_u1_2026-10-31.md`.
