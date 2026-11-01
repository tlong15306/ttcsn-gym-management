# Tiến độ và kế hoạch hoàn thiện — 31/10/2026

Ngày đối chiếu: **31/10/2026, Asia/Ho_Chi_Minh**. Phạm vi lượt này: kiểm tra source, thay đổi Git, log, trạng thái dịch vụ và các kiểm tra không cần DB; viết kế hoạch. Chưa sửa nghiệp vụ, seed/reset DB, khởi động dịch vụ hay commit thay người dùng.

Tài liệu này là điểm theo dõi hiện hành, tiếp nối [kế hoạch 29](29_ke_hoach_demo_va_keo_chon_gio.md) và [kế hoạch 28](28_ke_hoach_lam_viec_tiep_2026_10_25.md). Các báo cáo 14–25/10 là snapshot lịch sử, chỉ nâng trạng thái khi có bằng chứng chạy mới.

## 1. Kết luận tiến độ

Các module chức năng chính đã có API/UI: Identity và phân quyền, Hội viên, HLV, phòng/ca/phiên, lịch phòng, Booking, Attendance/Completion, Review và Report. Đợt 25/10 đã bổ sung mã nguồn seed DEMO, batch booking và kéo chọn giờ. Dự án đang ở giai đoạn **hoàn thiện và kiểm chứng để bàn giao**, chưa đủ bằng chứng để gọi là bản demo hoàn tất.

Không quy đổi số endpoint hoặc test thành phần trăm nghiệm thu. `REQ-15` xuất PDF/Excel vẫn là Could và chưa triển khai; các cổng Must/Should, dữ liệu demo thật, browser E2E và NFR còn phải đóng.

| Hạng mục | Đã có | Phần còn mở |
|---|---|---|
| Core nghiệp vụ | API/UI và test cho các module chính; migrations đến v6 trong source | Regression PostgreSQL trên working tree hiện tại, luồng liên thông 5 vai trò |
| Batch booking | Hai endpoint batch, transaction, idempotency, validation cùng phòng/ngày/liền kề; có integration test | Chạy DB thật; bổ sung race batch–batch/batch–single, PT/Hội viên trùng ở một ca và rollback audit/idempotency |
| Chọn giờ | `SessionRangePicker`, grid phòng, grid HLV 30 phút, cách chạm/bàn phím | Sửa lệch contract giờ HLV; kiểm tra Escape, cuộn/cảm ứng, thả ngoài vùng, đọc lại sau lưu |
| Dữ liệu DEMO | Command và catalog dự kiến 6 phòng, 4 HLV, 12 Hội viên, 24 booking | Chưa xác minh số lượng trong DB; seed có khoảng trống về refresh, deadline, nghiệp vụ và manifest |
| Tài liệu/contract | Contract checker khớp 78 operations; docs checker PASS | Đồng bộ kết quả runtime, account/scenario và hướng dẫn demo theo bản bàn giao |
| Môi trường | Có script start/stop và log lần API/client đã chạy 25/10 | Hiện không thấy listener 5433/8080/5173/18080/15173; Docker engine không truy cập được |
| NFR/vận hành | Có report tải, expiration backlog, restore ngày 18/10 | NFR-01/02 chưa đạt toàn bộ; security, completion backlog và RPO/RTO cần kiểm chứng tiếp |
| Git/bàn giao | HEAD `1226740`; các thay đổi demo/grid/batch vẫn có trong working tree | Chưa được commit thành mốc bàn giao; giữ nguyên thay đổi có trước và các file ngoài phạm vi |

### Kiểm tra chạy lại trong lượt này

| Kiểm tra | Kết quả |
|---|---|
| `python scripts/generate_api_contracts.py --check` | PASS, 78 operations |
| `powershell -ExecutionPolicy Bypass -File docs/check_docs.ps1` | PASS, 20 REQ / 16 UC / 26 BR / 18 bảng |
| `git diff --check` trước cập nhật kế hoạch | PASS, chỉ có cảnh báo chuyển LF/CRLF |
| `go test ./... -count=1` và `go vet ./...` | PASS; không bật integration DB |
| `npm.cmd run build` | PASS, còn cảnh báo main chunk 513,59 kB >500 kB |
| `npm.cmd test` | FAIL: 27 PASS / 25 FAIL, phần lớn timeout 5 giây và một lỗi chờ option; chạy cùng lượt Go, chưa kết luận nguyên nhân |
| `npm.cmd test -- --maxWorkers=1 --no-file-parallelism` | FAIL: 51 PASS / 1 FAIL; test bảo vệ ngày có phút lẻ trong `AvailabilityGrid.test.tsx` timeout 5 giây. Giữ nguyên timeout. |
| PostgreSQL integration, seed và browser | Chưa chạy trong lượt này; Docker engine không sẵn sàng |

Lượt Go test đặt `TEST_DATABASE_URL` rỗng để không chạy các test reset DB. Vì vậy kết quả Go thành công cũng không thay bằng chứng integration PostgreSQL. Frontend cải thiện khi giảm worker nhưng chưa PASS toàn suite; test/build từng chồng thời gian ở một phần lượt kiểm tra, nên chưa kết luận nguyên nhân timeout. Kết quả 52/52 ngày 25/10 không được dùng để thay kết quả hiện tại.

Log cũ `.artifact_work/runtime/api.stderr.log` cho thấy API ready ngày 25/10, sau đó có nhiều lỗi `BOOKING_JOB_FAILED` cho expiration/completion, kéo dài tới 26/10. Log chỉ có mã chung, chưa đủ kết luận nguyên nhân là DB, timeout hay nghiệp vụ. Không dùng lỗi socket Docker của lần trước làm chẩn đoán hiện tại: lần kiểm tra này chỉ xác nhận engine không truy cập được.

## 2. Các khoảng trống cụ thể cần sửa/kiểm chứng

Các nhận xét dưới đây dựa trên source hiện tại; chưa phải lỗi đã tái hiện qua browser/DB trong lượt này.

| Mã | Ưu tiên | Phát hiện | Công việc cần làm |
|---|---|---|---|
| F01 | P0 | `ensureBookingFixtures` dừng nếu trạng thái booking DEMO khác fixture ban đầu. Job hoặc thao tác demo hợp lệ sẽ làm lần seed sau thất bại. Mã fixture cố định cũng không tạo bộ tương lai mới khi qua ngày. | Giữ booking đã thao tác, quản lý scenario theo đợt/ngày; phân biệt seed lặp với refresh; báo cáo ID đã giữ/tạo mới, không reset dữ liệu. |
| F02 | P0 | `createFixtureBooking` chèn SQL cho cả booking tương lai; `PENDING_PT` lấy deadline `starts_at - 30 phút`, trong khi `booking.Deadline` là tối đa 30 phút từ lúc tạo. | Tạo booking live qua service/repository nghiệp vụ và deadline hiện hành; fixture lịch sử có nhãn/audit và timestamp nhất quán. |
| F03 | P0 | `ensureUsableSlots` chấp nhận chỉ một ca trong 06–21h; không đảm bảo có hai ca liền kề để demo kéo. Seed chưa có manifest và scenario đầy/ít chỗ được kiểm chứng. | Chuẩn bị ca phù hợp không chồng lịch hiện hữu; xuất manifest account/phòng/session/booking/trạng thái; bổ sung ca đầy, còn một chỗ, PT bận/nghỉ. |
| F04 | P0 | Grid HLV có thể xuất `end_time="24:00"`; `validHours` dùng `time.Parse("15:04")` nên từ chối mốc này. | Chốt một contract UI/API; chặn vùng không hợp lệ hoặc hỗ trợ 24:00 có kiểm tra end-only; test lưu/đọc lại ranh giới ngày. |
| F05 | P0 | `SessionRangePicker` chưa có handler Escape, cơ chế tự cuộn khi kéo; lần chọn range lỗi có nhánh đổi thành một ca mới. | Làm rõ thao tác hủy và giữ/xóa lựa chọn; phản hồi lỗi dải giờ; kiểm tra pointer cancel/thả ngoài, touch cuộn và keyboard theo kế hoạch 29. |
| F06 | P1 | `useCommonPTs` lấy trang 1, tối đa 100 HLV cho từng ca rồi giao danh sách; không tải các trang tiếp theo, dữ liệu cũ chưa bị xóa ngay khi đổi dải. | Không bỏ sót PT sau trang đầu; cân nhắc API khả dụng cho cả dải hoặc tải đủ trang; tránh chọn PT cũ khi đang tải/lỗi. |
| F07 | P0 | Test batch hiện có success/replay/validation/quyền/ca đầy, nhưng chưa có test cạnh tranh nhiều request cho batch mới. | Bổ sung concurrency và rollback assertions; chạy cả batch và single trên cùng tài nguyên, giữ lock order nhất quán. |
| F08 | P0 | Log job có lỗi lặp, chưa có nguyên nhân và bằng chứng hồi phục sau mất DB/restart. | Tái hiện trên DB test; thêm chẩn đoán có che dữ liệu nhạy cảm nếu cần; kiểm tra phục hồi, audit và không xử lý trùng. |
| F09 | P0 | Frontend regression ngày 31/10 có 25 test thất bại ở lượt mặc định; lượt 1 worker còn 1 timeout grid HLV, khác kết quả 52/52 ngày 25/10. | Đo/kiểm tra môi trường và chạy từng file/worker; xác định lỗi test hay ứng dụng trước khi sửa, không coi build PASS là regression PASS. |

Nguồn chính: `server/cmd/seed-demo-scenarios/main.go`, `server/internal/domain/booking/booking.go`, `server/internal/service/v1/trainer_service.go`, `server/internal/repository/booking_batch_integration_test.go`, `client/src/features/booking/{Booking,SessionRangePicker}.tsx`, `client/src/features/trainers/AvailabilityGrid.tsx` và log runtime.

## 3. Kế hoạch hoàn thiện theo thứ tự ưu tiên

| Chặng | Ưu tiên / khối lượng dự kiến | Công việc | Cổng hoàn thành |
|---|---|---|---|
| H0 — Môi trường và hồi phục | P0 · 1 buổi | Xác minh Docker/DB/schema; chẩn đoán F08/F09; tạo backup DB dev; chuẩn hóa API/client test tách dev và test. | API ready, client truy cập được; job hồi phục sau gián đoạn; frontend regression chạy lặp có kết quả rõ nguyên nhân; có tên DB và backup đối chiếu. |
| H1 — Dữ liệu demo dùng được | P0 · 1–2 buổi | Sửa F01–F03; seed/refresh rõ ràng; account/scenario manifest; lịch tương lai theo ngày chạy. | Đọc được 6 phòng/4 HLV/12 Hội viên và ít nhất 24 booking minh họa; có 2 ca liền kề, ca đầy/ít chỗ; seed lặp và refresh sau thay đổi không nhân bản hay ghi đè thao tác. |
| H2 — Chốt batch và chọn giờ | P0 · 1–2 buổi | Sửa F04/F05; xử lý F06 trong phạm vi khả dụng PT; bổ sung F07, ca PT/Hội viên bận, retry và rollback audit. | Kéo 2+ ca → một lần xác nhận; thành công đủ hoặc không tạo ca nào; replay không tăng booking/audit; mốc giờ UI gửi lên hợp lệ; lưu/đọc lại chính xác. |
| H3 — E2E và ngoại lệ | P0 · 1–2 buổi | Browser Hội viên → PT → Lễ tân → completion → review → report; accept/reject/expire/cancel; kiểm tra quyền, 401/mất mạng; responsive, touch, keyboard và focus. | Có booking ID/request ID/ảnh/SQL đối chiếu; không còn lỗi P0/P1 cản luồng; tách fixture lịch sử và completion chờ giờ thật. |
| H4 — Vận hành/NFR/bảo mật | P1 · 1–2 buổi | Workload 100 danh tính có mutation và metrics host/DB; profile login; completion backlog >1.000, timeout/restart; object scope/upload/log/dependency; restore DB + file private. | Số liệu p95 và tính đúng theo workload công bố; báo đạt/chưa đạt NFR riêng; restore sang DB mới, khớp dữ liệu/file và đo thời gian. |
| H5 — Bàn giao | P1 · 1 buổi | Regression cuối; cập nhật backlog/ma trận/report; hướng dẫn tài khoản và demo 5 vai trò; commit theo nhóm thay đổi, checklist chạy/tắt và các giới hạn còn lại. | Người dùng mở app, thử luồng từ dữ liệu sẵn, restart vẫn đọc được; report gắn revision; tài liệu khớp bản bàn giao. |

Ước lượng **6–10 buổi tập trung**, phụ thuộc môi trường, lỗi tìm được và thời gian chờ ca thật; không phải cam kết ngày. Có thể dùng H0–H3 làm mốc demo chức năng trước, nhưng chỉ gọi hoàn thiện kỹ thuật khi H4–H5 có bằng chứng tương ứng. Không hoãn lỗi sai dữ liệu hoặc phân quyền sang sau demo.

### Thứ tự kiểm chứng và ranh giới dữ liệu

1. `ttcsn_gym_dev`: dữ liệu demo lưu bền; backup trước seed đầu/refresh; kiểm tra tên DB từ kết nối thực.
2. `ttcsn_gym_test`: integration/race/Newman/browser mutation. Các suite reset DB phải chạy lần lượt, không chạy song song trên cùng DB.
3. API test dùng `18080`; client test dự kiến `15173`, phải cấu hình proxy/CORS đúng. `vite.config.ts` hiện trỏ API dev `8080`, chưa thể chỉ đổi port client rồi coi là môi trường test.
4. Test/frontend build nên chạy riêng khi máy bị tải; ghi cả lượt thất bại và lượt kiểm tra lại, không tăng timeout để che lỗi chưa chẩn đoán.
5. Giữ DeepSeek Harness chính thức là harness AI của dự án. Dùng script kiểm chứng hiện có; không đưa thêm hệ harness AI khác vào phạm vi này.

## 4. Phạm vi bản hoàn thiện

- Đợt này: đóng luồng nghiệp vụ hiện có, dữ liệu DEMO sử dụng được, chọn giờ, kiểm chứng UI/API/DB, vận hành và tài liệu bàn giao.
- Sau các cổng chính mới cân nhắc `REQ-15` xuất PDF/Excel; cần ưu tiên thêm nếu học phần bắt buộc.
- Chưa mở rộng thanh toán, doanh thu, nhiều cơ sở, app native, notification/WebSocket hay QR/RFID.
- Không tự điền tiến độ/ngày thực tế của nhóm vào Excel, không suy diễn việc source có sẵn thành hoạt động hoặc nghiệm thu học phần.

**Bước tiếp theo đề xuất:** bắt đầu H0, sau đó sửa seed DEMO ở H1 trước khi tổng duyệt kéo chọn/batch. Lượt này bàn giao đánh giá và kế hoạch; các sửa ứng dụng được liệt kê vẫn chưa thực hiện.
