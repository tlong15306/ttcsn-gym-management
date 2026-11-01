# Bộ tài liệu Hệ thống Quản lý GYM

## Baseline đang có hiệu lực

Baseline phân tích và thiết kế hiện hành là **v0.3**, lập ngày **10/10/2026**. Bộ tài liệu v0.3 dùng để triển khai và kiểm thử; các file v0.1/v0.2 được giữ làm lịch sử và không còn là nguồn quyết định hiện hành.

Nhật ký học và tự phát triển của sinh viên nằm tại `../NHAT_KY_DU_AN.md`; cập nhật file này sau mỗi buổi code.

| Thứ tự | Tài liệu hiện hành | Mục đích |
|---:|---|---|
| 1 | `00_quyet_dinh_va_quy_tac_nghiep_vu.md` | Phạm vi, quyết định D01–D10 và quy tắc BR. |
| 2 | `03_tac_nhan.md` | Mục tiêu, trách nhiệm và phạm vi dữ liệu của 5 tác nhân. |
| 3 | `04_use_case_tong_quat.md` | Danh mục UC-01…UC-16 ổn định. |
| 4 | `06_ma_tran_phan_quyen.md` | Nguồn chính sách quyền duy nhất. |
| 5 | `07_SRS_v0.3.md` | Baseline yêu cầu chức năng và phi chức năng. |
| 6 | `08_use_case_chi_tiet.md` | Luồng chính, ngoại lệ và hậu điều kiện của use case. |
| 7 | `09_ERD_va_CSDL.md` | Mô hình dữ liệu logic/vật lý và quy tắc toàn vẹn. |
| 8 | `09_schema.sql` | Schema PostgreSQL 16+ cho cài đặt mới; không phải migration nâng cấp. |
| 9 | `11_kien_truc_he_thong.md` | Kiến trúc module, transaction, bảo mật và vận hành. |
| 10 | `12_dac_ta_API.md` | Hợp đồng REST API, lỗi, quyền và idempotency. |
| 11 | `13_backlog_v0.3.md` | Ưu tiên, phụ thuộc, trạng thái phân tích/triển khai. |
| 12 | `15_ke_hoach_kiem_thu.md` | Kiểm thử yêu cầu, dữ liệu, đồng thời, bảo mật và NFR. |
| 13 | `16_runbook_van_hanh_nang_cap.md` | Migration, backup/restore, job nền và xử lý sự cố. |
| 14 | `17_huong_dan_chay_local.md` | Chạy database, API và CRUD quản trị bằng PowerShell. |
| 15 | `18_ke_hoach_lam_viec_tiep.md` | Phần kiểm chứng còn thiếu và lộ trình triển khai sau tuần 4. |
| 16 | `reports/week4_verification_2026-10-11.md` | Kết quả kiểm tra trước commit và giới hạn kiểm chứng. |
| 17 | `19_ke_hoach_crud_admin_ui_ux.md` | CRUD quản trị và thiết kế lại UI/UX theo từng module. |
| 18 | `reports/admin_crud_verification_2026-10-11.md` | Kết quả backend/DB/frontend/browser của gói CRUD. |
| 19 | `openapi_admin_crud.json` / `GYM_Admin_CRUD.postman_collection.json` | Hợp đồng sinh từ router hiện có; 78 operations. |
| 20 | `20_ke_hoach_lich_hoat_dong_phong.md` | Contract và chặng triển khai lịch phòng 7 ngày × 24 giờ. |
| 21 | `21_huong_dan_lich_phong.md` | Sử dụng grid, ngoại lệ, preview, sinh phiên và chuyển đổi phòng cũ. |
| 22 | `reports/room_calendar_verification_2026-10-11.md` | Bằng chứng kiểm thử lịch phòng và giới hạn đã kiểm chứng. |
| 23 | `22_ke_hoach_booking_member_pt_ui_ux.md` | Booking, Hội viên chọn phòng qua ảnh rồi đặt ca, layout PT riêng; bản phác desktop/mobile, chưa triển khai. |
| 24 | `reports/booking_portals_design_2026-10-11.md` | Kiểm tra bản phác và ranh giới giữa thiết kế với app hoạt động. |
| 25 | `25_ke_hoach_chot_attendance_review_report.md` | Các chặng chốt Attendance, Review, Report và phần UI/bàn giao còn lại. |
| 26 | `reports/review_backend_verification_2026-10-14.md` | Review API, eligibility, quyền, privacy và PostgreSQL tests. |
| 27 | `reports/review_ui_verification_2026-10-14.md` | Form/readback Review, tests/build và giới hạn UI còn lại. |
| 28 | `reports/report_api_verification_2026-10-14.md` | Công thức Metrics, scope, nguồn legacy và PostgreSQL verification. |
| 29 | `reports/report_ui_verification_2026-10-14.md` | Giao diện Reports/PT, 45 frontend tests, build và giới hạn browser/demo live. |
| 30 | `26_ke_hoach_chot_phan_con_thieu.md` | Kế hoạch sau rà tuần 1–6: đồng bộ tiến độ, Postman Runner, browser E2E, UI QA, NFR/restore và bàn giao. |
| 31 | `27_ma_tran_tien_do_va_kiem_chung_tuan_1_6.md` | Trạng thái code, kiểm thử và các cổng còn thiếu theo bằng chứng; không thay cho xác nhận hoạt động nhóm. |
| 32 | `reports/local_regression_2026-10-18.md` | Chạy lại local: script, Newman 15/15, Go/DB/frontend regression và browser smoke 5 vai trò. |
| 33 | `reports/operations_nfr_restore_2026-10-18.md` | Idempotency retention, backlog job, tải 100 concurrent, CORS/login smoke và backup/restore local; ghi rõ p95 chưa đạt. |
| 34 | `reports/mutation_smoke_2026-10-18.md` | Newman mutation Identity trên API/DB test cô lập: create, lock/unlock, reset/revoke, RBAC, audit và cleanup. |
| 35 | `28_ke_hoach_lam_viec_tiep_2026_10_25.md` | Kế hoạch tiếp nối 25/10: API nghiệp vụ, browser đủ vòng đời PT, UI QA, tải/job, bảo mật/restore và chốt demo local. |
| 36 | `29_ke_hoach_demo_va_keo_chon_gio.md` | Kế hoạch và kết quả đợt 25/10: seed demo an toàn, kéo chọn giờ phòng/HLV, Hội viên đặt nhiều ca theo transaction; ghi rõ phần live còn chờ môi trường. |
| 37 | `30_tien_do_va_ke_hoach_hoan_thien_2026_10_31.md` | Tiến độ ngày 31/10, khoảng trống demo/grid/batch/job và kế hoạch H0–H5 với cổng nghiệm thu. |

`14_ke_hoach_hoan_thien_tai_lieu.md` là kế hoạch và biên bản rà soát đầu vào. `01_khao_sat_chuc_nang.md`, `05_khung_tien_do_10_tuan.md` và `10_so_sanh_cong_nghe.md` là tài liệu tham khảo; khi có xung đột, baseline v0.3 ở bảng trên được ưu tiên.

## Quy tắc quản trị tài liệu

- Mỗi yêu cầu dùng mã `REQ-*` hoặc `NFR-*`; mỗi use case dùng mã `UC-*`; mỗi quy tắc nghiệp vụ dùng mã `BR-*`.
- Thay đổi phạm vi hoặc quy tắc phải được ghi trong decision log trước khi sửa SRS, ERD và API.
- Quyền được quản lý tại `06_ma_tran_phan_quyen.md`; các tài liệu khác dẫn chiếu, không tạo ma trận quyền cạnh tranh.
- DDL cài đặt mới, migration nâng cấp và seed dữ liệu là ba artifact khác nhau. Không chạy `09_schema.sql` trên CSDL có dữ liệu.
- Trạng thái “Đã triển khai” hoặc “Đã kiểm thử” chỉ được ghi khi có mã nguồn/kết quả kiểm thử tương ứng. Baseline v0.3 mô tả phân tích/thiết kế; tiến độ thực thi nằm trong backlog và báo cáo kiểm tra, không suy diễn từ tài liệu thành nghiệm thu.

## Trạng thái xác minh

UI/UX mới: [35 — Trang chủ công khai và chọn lịch Hội viên](35_de_xuat_ui_ux_trang_chu_va_dat_lich.md). U1–U4 đã triển khai và kiểm tra local; kết quả mới nhất: [U2–U4, 01/11/2026](reports/public_booking_u2_u4_2026-11-01.md).

Mockup U0: [Trang chủ và lịch tuần desktop/mobile](mockups/u0/index.html), kèm [hướng dẫn duyệt](mockups/u0/README.md).

Kết quả thực hiện mới: [Kiểm chứng giờ và demo ngày 31/10](reports/time_demo_verification_2026-10-31.md).

U1: [Trang chủ, phòng, HLV và lịch công khai](reports/public_discovery_u1_2026-10-31.md).

Hướng dẫn thao tác demo: [33 — Hướng dẫn demo GYM](33_huong_dan_demo.md).

Kế hoạch đang triển khai: [32 — Xử lý giờ và ổn định demo](32_ke_hoach_xu_ly_gio_va_on_dinh_demo.md).

Kết quả mới nhất sau khi thực hiện kế hoạch: [31 — Báo cáo demo ngày 31/10](31_bao_cao_demo_2026_10_31.md). Báo cáo ghi rõ phần đã kiểm chứng và các cổng còn mở.

Snapshot đánh giá và kế hoạch ngày **31/10/2026**: [30 — Tiến độ và kế hoạch hoàn thiện](30_tien_do_va_ke_hoach_hoan_thien_2026_10_31.md). Tại thời điểm lập kế hoạch, Docker không truy cập được và chưa kiểm chứng DB; trạng thái sau đó được cập nhật trong báo cáo 31. Log job cũ có lỗi lặp đến 26/10, chưa kết luận nguyên nhân.

Snapshot đợt **25/10/2026** theo [29 — Dữ liệu demo và kéo chọn giờ](29_ke_hoach_demo_va_keo_chon_gio.md) ghi phần code và kiểm tra tĩnh đã xong, nhưng seed/browser/DB chưa kiểm chứng. [28 — Kế hoạch làm việc tiếp theo](28_ke_hoach_lam_viec_tiep_2026_10_25.md), [26 — Chốt các phần còn thiếu](26_ke_hoach_chot_phan_con_thieu.md) và [27 — Ma trận tuần 1–6](27_ma_tran_tien_do_va_kiem_chung_tuan_1_6.md) giữ mốc lịch sử; kế hoạch thực thi tiếp theo theo H0–H5 trong tài liệu 30.

Snapshot kế hoạch ngày 14/10/2026: [25 — Chốt Attendance, Review và Report](25_ke_hoach_chot_attendance_review_report.md), tiếp nối [kế hoạch 24](24_ke_hoach_attendance_review_report.md). A `093469a`, B `f35775d`, C Review backend `a58c359`, D Review UI `a67db62`, E Metrics/Report API `92cb96f`, F UI/report `49b423a`; 45 frontend tests và production build PASS. Backend PostgreSQL tests, `go vet`, contract và docs checker PASS. Báo cáo: [Review backend](reports/review_backend_verification_2026-10-14.md), [Review UI](reports/review_ui_verification_2026-10-14.md), [Report API](reports/report_api_verification_2026-10-14.md), [Report UI](reports/report_ui_verification_2026-10-14.md). Browser smoke local 5 vai trò và demo UI → DB đã chạy lại; bằng chứng, booking synthetic và giới hạn E2E/browser matrix/tải ở [regression 18/10](reports/local_regression_2026-10-18.md).

Bộ tài liệu đã được đối chiếu tĩnh về mã REQ/UC, quyền, trạng thái và trường dữ liệu. Backend có Identity, Members, Rooms, TimeSlots, RoomSessions, lịch phòng tuần/ngoại lệ và PT (hồ sơ/chứng chỉ/duyệt/lịch), Goose migrations đến v6 và PostgreSQL riêng. Giao diện React tại `../client` có route theo quyền, form/dialog, bảng desktop/card mobile và lịch phòng 7 × 24. Hướng dẫn local: `17_huong_dan_chay_local.md`, `21_huong_dan_lich_phong.md`; kiểm chứng CRUD cũ: `reports/admin_crud_verification_2026-10-11.md`; kiểm chứng lịch phòng và giới hạn browser: `reports/room_calendar_verification_2026-10-11.md`.

Booking và ảnh phòng theo kế hoạch 22 đã có mã nguồn, cùng portal Hội viên/PT riêng. Hướng dẫn: [23_huong_dan_booking_portal.md](23_huong_dan_booking_portal.md); kiểm chứng trước đó: [reports/booking_verification_2026-10-12.md](reports/booking_verification_2026-10-12.md). Attendance/Completion, Review và Metrics/Report API có báo cáo kiểm chứng tương ứng. Excel giữ nguyên; không suy diễn việc Codex triển khai thành ngày hoàn thành hay mức hiểu bài của sinh viên. Chưa phải MVP GYM đầy đủ.

Chạy kiểm tra nhất quán sau mỗi lần sửa:

```powershell
powershell -ExecutionPolicy Bypass -File .\docs\check_docs.ps1
```
