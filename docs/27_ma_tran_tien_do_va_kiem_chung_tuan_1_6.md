# Ma trận tiến độ và kiểm chứng tuần 1–6

**Đối chiếu hiện hành 31/10/2026:** [30 — Tiến độ và kế hoạch hoàn thiện](30_tien_do_va_ke_hoach_hoan_thien_2026_10_31.md). Ma trận dưới đây giữ snapshot 18/10; không dùng kết quả cũ thay cho nghiệm thu đợt demo/grid/batch mới.

**Ngày đối chiếu:** 18/10/2026. **Phạm vi:** source, migration, test và báo cáo nằm trong repository. Đây là bảng trạng thái kỹ thuật; không chứng minh phỏng vấn, họp, review chéo, phân công hay phê duyệt thực tế của nhóm.

## Cách đọc trạng thái

- **Có artifact**: tài liệu hoặc source đã tồn tại và có thể đối chiếu.
- **PASS kỹ thuật local**: có báo cáo lệnh/test trên môi trường local hoặc PostgreSQL test; không suy ra browser hay production.
- **Còn cổng nghiệm thu**: còn thiếu bằng chứng chạy lại, browser E2E, benchmark, restore hoặc xác nhận nhóm tương ứng.

## Ma trận theo chặng trong kế hoạch 10 tuần

| Tuần | Hạng mục đã đối chiếu | Trạng thái có bằng chứng | Bằng chứng chính | Cổng còn thiếu / giới hạn |
|---|---|---|---|---|
| 1 | Khảo sát sản phẩm, tác nhân, phạm vi và quy tắc nghiệp vụ | Có artifact | `01_khao_sat_chuc_nang.md`, `03_tac_nhan.md`, `00_quyet_dinh_va_quy_tac_nghiep_vu.md` | Không có bằng chứng phỏng vấn cơ sở hoặc xác nhận nhóm; chỉ coi là khảo sát/tổng hợp tài liệu |
| 1 | Use case, quyền và backlog MVP | Có artifact, checker tài liệu PASS trong snapshot 14/10 | `04_use_case_tong_quat.md`, `06_ma_tran_phan_quyen.md`, `13_backlog_v0.3.md`, `reports/week4_verification_2026-10-11.md` | Cần review chéo thật nếu học phần yêu cầu |
| 2 | SRS v0.3, ERD/schema, API và kiến trúc | Có artifact, đối chiếu tĩnh PASS trong snapshot | `07_SRS_v0.3.md`, `08_use_case_chi_tiet.md`, `09_ERD_va_CSDL.md`, `11_kien_truc_he_thong.md`, `12_dac_ta_API.md` | Không có biên bản chốt/duyệt thực tế |
| 2 | Kế hoạch kiểm thử và runbook | Có artifact | `15_ke_hoach_kiem_thu.md`, `16_runbook_van_hanh_nang_cap.md`, `17_huong_dan_chay_local.md` | Cần chạy lại từng cổng và lưu log/report mới |
| 3 | Config, health, PostgreSQL riêng, migration và seed | PASS kỹ thuật local đã chạy lại | `server/internal/config`, `server/internal/db`, `cmd/migrate`, `cmd/seed`, `reports/local_regression_2026-10-18.md` | Không thay cho vận hành production |
| 3 | Request ID, recovery, error mapping, log và shutdown | Có source và test local | `server/internal/middleware`, `server/internal/handler/v1`, `server/internal/app` | Chưa có load/chaos hoặc production observability |
| 4 | Identity: bcrypt, JWT, login/logout, revoke và RBAC | PASS kỹ thuật local; Newman 5 role và browser login đã chạy lại | `server/pkg/auth`, `domain/identity`, `GYM_Role_Smoke.postman_collection.json`, `reports/local_regression_2026-10-18.md` | Chưa có browser matrix đầy đủ về 401 giữa form/TTL |
| 4 | Quản trị tài khoản và audit transaction | PASS kỹ thuật local trong snapshot | `handler/v1/identity_handler.go`, `service/v1/service.go`, `repository/identity_repository.go`, `reports/admin_crud_verification_2026-10-11.md` | Cần thao tác browser với database thật và kiểm tra lại token bị thu hồi |
| 5 | Hội viên, phòng, ca, phiên phòng, hồ sơ PT và lịch PT | PASS PostgreSQL/API/frontend local trong snapshot | `reports/admin_crud_verification_2026-10-11.md`, `reports/room_calendar_verification_2026-10-11.md` | Lịch ngoại lệ/chỉ đọc và responsive browser chưa đủ ma trận |
| 5 | Lịch phòng tuần 7 × 24, ngoại lệ, preview và sinh phiên | PASS kỹ thuật local trong snapshot | migration `00003_room_calendar.sql`, `calendar_*`, `reports/room_calendar_verification_2026-10-11.md` | Browser grid, keyboard, mobile và job theo đồng hồ thực còn cần kiểm chứng |
| 6 | Booking, PT response, hủy, hold, idempotency và portal Member/PT riêng | PASS PostgreSQL/API/frontend; browser tạo và đọc lại booking self-training synthetic #4 | `reports/booking_verification_2026-10-12.md`, `reports/local_regression_2026-10-18.md` | PT response/cancel/retry mạng và các race tổ hợp browser còn thiếu |
| 6 | Attendance, completion/no-show | PASS PostgreSQL/API/frontend local trong snapshot | `reports/attendance_verification_2026-10-14.md`, `attendance_*`, `job/booking.go` | Browser E2E, backlog job >1.000 và ca chờ clock/job thật còn thiếu |
| 6 | Review, rating và privacy | PASS PostgreSQL/API/frontend local trong snapshot | `reports/review_backend_verification_2026-10-14.md`, `reports/review_ui_verification_2026-10-14.md` | E2E theo ca hoàn thành thật và keyboard/browser cần chạy |
| 6 | Report vận hành và hiệu suất PT | PASS PostgreSQL/API/frontend local trong snapshot | `reports/report_api_verification_2026-10-14.md`, `reports/report_ui_verification_2026-10-14.md` | Browser E2E và benchmark p95 chưa chạy |

## Trạng thái yêu cầu chức năng hiện tại

| Nhóm | REQ | Trạng thái kỹ thuật | Chưa được coi là đạt |
|---|---|---|---|
| Identity/quản trị | REQ-01, 02, 12 | Có API/UI, RBAC, revoke, audit và test local | E2E browser/API Runner mới trên bản hiện tại |
| Master data | REQ-03, 04, 05, 11, 16, 17 | Có API/UI/migration và test local | Browser matrix đầy đủ và review nhóm |
| Booking | REQ-06, 07, 08, 13, 14, 18 | Có API/UI/job/test transaction local | Luồng 5 vai trò trên browser, retry mạng và tải thực |
| Operations | REQ-09, 19, 20 | Có API/UI/test PostgreSQL local | Luồng kết thúc buổi–điểm danh–review trên browser/DB thật |
| Reports | REQ-10 | Có API/UI/test công thức và scope local | E2E và benchmark theo NFR-01/NFR-02 |
| Xuất file | REQ-15 | Chưa triển khai; phạm vi Could | Chỉ làm sau các cổng Must/Should |

## Cổng còn mở và thứ tự xử lý

1. **R1:** đã có script và smoke local tại `reports/local_regression_2026-10-18.md`; tiếp tục giữ lệnh chạy lặp lại được.
2. **R2:** Newman role smoke 15/15 PASS; mutation smoke Identity cô lập DB test 14/14 PASS, gồm create, lock/unlock, reset/revoke, RBAC và 4 audit. Collection mutation cho Room/Calendar/Booking/Attendance/Review/Report còn mở.
3. **R3:** đã browser smoke 5 vai trò và viewport một số màn hình; tiếp tục E2E vòng đời, keyboard/accessibility và browser matrix.
4. **R4:** retention idempotency, backlog expiration 1.001, CORS/login smoke và một backup/restore drill đã có tại `reports/operations_nfr_restore_2026-10-18.md`. Burst login/read 100 concurrent trả 400/400 nhưng p95 2.435,10 ms; warm-session read có ramp trả 300/300 với p95 80,92 ms. NFR-01 tổng thể vẫn chưa đạt; security review mở rộng và workload hỗn hợp/metrics host-DB còn mở.
5. **R5:** tài liệu demo, slide và các phần nhóm xác nhận thật.

Kế hoạch chi tiết và điều kiện an toàn của các chặng là [26 — Kế hoạch chốt phần còn thiếu](26_ke_hoach_chot_phan_con_thieu.md). Khi một chặng chạy xong, bổ sung report có commit, lệnh, môi trường, thời gian, kết quả và giới hạn; không thay đổi trạng thái bằng lời khẳng định không có bằng chứng.
