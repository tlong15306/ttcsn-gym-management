# Vận hành, NFR và restore local — 18/10/2026

## Phạm vi và môi trường

- Môi trường local Windows, PostgreSQL 17 trong `ttcsn-gym-postgres` tại `127.0.0.1:5433`, API Gin tại `127.0.0.1:8080`.
- Database phát triển là `ttcsn_gym_dev`; kiểm thử repository dùng riêng `ttcsn_gym_test`. Không reset hay restore đè database phát triển.
- Kết quả dưới đây là kiểm chứng local tại thời điểm chạy; không phải benchmark production, chứng nhận an toàn hay xác nhận hoạt động của nhóm.

## Idempotency và job booking

Migration Goose `00006_idempotency_retention.sql` thêm index `idx_idempotency_keys_expires_at`; database development đã nâng từ v5 lên v6.

Quy tắc replay giữ nguyên: key cùng actor/scope/nội dung được replay trong 24 giờ. Khi key hết hạn, Booking và Calendar xóa bản ghi đã hết hạn dưới transaction/lock trước khi xử lý yêu cầu mới; Calendar không replay bản ghi quá hạn. Job booking chạy dọn dẹp một lần khi khởi động và mỗi giờ, với timeout 12 giây. Không log hash, mật khẩu hoặc token.

| Kiểm tra | Lệnh/phạm vi | Kết quả |
|---|---|---|
| Retention/replay | `go test .\\internal\\repository -run 'TestExpiredIdempotencyKeysAreNotReplayedAndArePurged|TestBookingAndCoverIntegration' -count=1` | PASS: Calendar không replay/commit conflict key hết hạn, Booking xử lý lại key hết hạn thay vì trả lại `201`; key đang hiệu lực còn lại. |
| Startup job | `go test .\\internal\\job -run TestBookingStartupPrunesExpiredIdempotencyKeys -count=1` | PASS, repository hỗ trợ purge được gọi khi job bắt đầu. |
| Backlog expiration | Fixture 1.001 booking `PENDING_PT` quá hạn trên `ttcsn_gym_test` | Lượt một giới hạn 1.000: còn 1; lượt hai: `EXPIRED=1001`, `PENDING_PT=0`; `audit_events` có `booking.EXPIRED=1001`. |

Kiểm thử backlog giữ fixture trong DB test có thể bị reset bởi lần chạy integration kế tiếp. Nó không đụng vào DB development.

## Tải local 100 concurrent

Lệnh đã chạy:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\invoke-api-load.ps1 -Concurrency 100 -RequestsPerUser 3
```

Script tạo 100 worker đồng thời, chia đều trên năm tài khoản seed, mỗi worker login đúng một lần rồi thực hiện ba request chỉ đọc (`/me`, room/booking hoặc report theo quyền). Nó đo 100 client đồng thời nhưng không giả lập 100 danh tính riêng. Không ghi dữ liệu và report không lưu token/mật khẩu. Kết quả chi tiết: `.artifact_work/load/api-load-20261018_120637.json`.

| Chỉ số | Kết quả |
|---|---:|
| Tổng request | 400 |
| 2xx | 400 |
| 4xx | 0 |
| 5xx/lỗi mạng | 0 |
| Thời gian | 4,99 s |
| Throughput | 80,15 request/s |
| p50 | 646,48 ms |
| p95 | 2.435,10 ms |
| p99 | 3.158,14 ms |

Để tách chi phí bcrypt của simultaneous login khỏi read path, script đã được bổ sung `-WarmSessions`, `-RampSeconds` và percentile theo phase. Lượt đo thứ hai dùng 100 worker, 100 token được tạo tuần tự ngoài thời gian đo, 20 giây ramp và 3 request đọc/worker:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\invoke-api-load.ps1 -Concurrency 100 -RequestsPerUser 3 -WarmSessions -RampSeconds 20
```

Kết quả `.artifact_work/load/api-load-20261018_122431.json`: 300/300 request 2xx, 4xx/5xx/network bằng 0, p50 15,81 ms, p95 **80,92 ms**, p99 167,87 ms, throughput 13,06 request/s. Đây là **read path đạt mục tiêu 2 giây trong điều kiện warm-session/ramp này**, không thay kết quả login burst.

**NFR-01 chưa đạt cho workload tổng thể** vì burst login/read vẫn p95 vượt 2 giây. Hai lượt mới chỉ dùng năm tài khoản seed lặp lại, không có mutation, dataset production, hold dài, CPU/RAM/connection/lock wait; cần profiling bcrypt/login và workload hỗn hợp đã chốt trước khi công bố NFR-01/NFR-02 đạt.

## Security smoke local

- Preflight từ `http://127.0.0.1:5173` tới `/api/v1/rooms` trả `204` với origin allowlist chính xác; unit test giữ lại allowlist và `Vary: Origin`.
- Preflight từ origin không tin cậy trả `204` nhưng không có `Access-Control-Allow-Origin`; unit test giữ hành vi này.
- Login với account không tồn tại trả `401`; kiểm tra response không có chuỗi SQL, password hash, JWT hoặc token.

Smoke này chỉ bao phủ CORS và lỗi login tổng quát. Rà object scope, upload, log, dependency và kiểm thử bảo mật chuyên sâu vẫn là việc mở.

## Backup và restore drill

Lệnh đã chạy:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\backup-restore-drill.ps1
```

Script chỉ nhận source `ttcsn_gym_dev` hoặc `ttcsn_gym_test`; restore target bắt buộc có tiền tố `ttcsn_gym_restore_` và phải chưa tồn tại. Nó dump PostgreSQL dạng custom, tạo manifest SHA-256, nén upload riêng tư nếu có, restore vào DB mới và đối chiếu dữ liệu. Ảnh phòng nằm trong PostgreSQL BYTEA nên thuộc archive DB.

| Mục | Source `ttcsn_gym_dev` | Restore `ttcsn_gym_restore_20261018_005904` |
|---|---:|---:|
| Schema Goose | 6 | 6 |
| Users | 10 | 10 |
| Rooms | 8 | 8 |
| Bookings | 4 | 4 |
| Audit events | 77 | 77 |
| Rooms có cover | 7 | 7 |

Archive `.artifact_work/backup-drills/ttcsn_gym_dev_20261018_005904.dump` có 2.851.316 bytes, SHA-256 `AFC57836E334A530127333FE6A9F5ECF30339211D42799F95FAC9090DDEC0DEE`. Manifest cùng thư mục ghi timestamp, source/target và đường dẫn archive upload private. Restore target được giữ để có thể kiểm tra độc lập; không ghi đè `ttcsn_gym_dev`, `ttcsn_gym_test` hoặc database ngoài project.

Một restore drill chứng minh quy trình local này hoạt động ở snapshot trên. Nó chưa chứng minh backup hằng ngày, cảnh báo lỗi, RPO <24h hay RTO <2h.

## Regression sau thay đổi

- `go test ./... -count=1` với `TEST_DATABASE_URL` trỏ đúng `ttcsn_gym_test`: PASS sau các thay đổi cuối; package repository mất 131,417 giây do reset/migration và fixture integration.
- `TestBookingAndCoverIntegration` và `TestExpiredIdempotencyKeysAreNotReplayedAndArePurged` cũng đã chạy riêng PASS với migration v6 từ database trống.
- `go vet ./...`, `npm.cmd test -- --run` (7 files, 45 tests), `npm.cmd run build`, contract checker và docs checker: PASS. Build vẫn cảnh báo main chunk 509,45 kB; đây là mục tối ưu sau cổng lỗi/NFR.

## Việc còn mở

1. Chốt workload có ramp/hold, dataset, metrics host/DB và sửa/đo lại p95 dưới mục tiêu.
2. Thực hiện security review có scope upload, object authorization, logging, dependency và lỗi input.
3. Bổ sung scenario Postman mutation cô lập trên DB test, browser E2E vòng đời đầy đủ, accessibility và browser matrix.
4. Chỉ lập slide/tổng duyệt sau khi các cổng kỹ thuật bắt buộc có bằng chứng tương ứng.
