# Chạy demo GYM CRUD quản trị trên Windows PowerShell

## 1. Chuẩn bị

Docker Desktop phải chạy Linux engine. Cần Go 1.26.5, Node.js 24, npm và các cổng 5433/8080/5173. PostgreSQL riêng của project có hai DB `ttcsn_gym_dev` và `ttcsn_gym_test`; database `master-golang` cổng 5432 được giữ nguyên.

Các lệnh dưới đây tương thích PowerShell 5.1. Codex chạy trực tiếp project; không cần mở DeepSeek Harness để chạy ứng dụng.

## 2. Database và cấu hình local (lần đầu)

```powershell
Set-Location C:\Users\choqu\OneDrive\Desktop\TTCSN
if (-not (Test-Path .\infra\.env.compose)) {
  Copy-Item .\infra\.env.compose.example .\infra\.env.compose
  notepad .\infra\.env.compose
}
docker compose --env-file .\infra\.env.compose up -d --wait
docker compose --env-file .\infra\.env.compose ps
```

Điền mật khẩu local trong `infra/.env.compose` trước khi chạy Compose. Tạo `server/.env.local` một lần nếu chưa có:

```powershell
if (-not (Test-Path .\server\.env.local)) {
  $composeText = Get-Content .\infra\.env.compose -Raw
  $composePassword = ($composeText -split '=', 2)[1].Trim()
  $encodedPassword = [Uri]::EscapeDataString($composePassword)
  $bytes = [byte[]]::new(32)
  $rng = [System.Security.Cryptography.RandomNumberGenerator]::Create()
  try { $rng.GetBytes($bytes) } finally { $rng.Dispose() }
  $jwtSecret = -join ($bytes | ForEach-Object { $_.ToString('x2') })
  @"
APP_ENV=development
HTTP_ADDR=127.0.0.1:8080
DATABASE_URL=postgres://gym_app:$encodedPassword@127.0.0.1:5433/ttcsn_gym_dev?sslmode=disable
JWT_SECRET=$jwtSecret
JWT_TTL=15m
BCRYPT_COST=10
UPLOAD_DIR=../uploads/private
CORS_ORIGINS=http://127.0.0.1:5173,http://localhost:5173
"@ | Set-Content .\server\.env.local -NoNewline
}
```

Hai file cấu hình, upload và cache local bị Git bỏ qua. JWT secret ngẫu nhiên dùng ký token; `BCRYPT_COST=10` là tham số hash mật khẩu. API ưu tiên biến môi trường PowerShell, rồi `.env.local`, rồi `.env`. Nếu shell còn `DATABASE_URL/JWT_SECRET` cũ, mở shell mới hoặc xóa hai biến trong shell trước khi chạy API.

## 3. Chạy nhanh bằng script project

Sau khi đã tạo `infra/.env.compose`, `server/.env.local` và chạy `npm.cmd ci` một lần, dùng một PowerShell tại thư mục gốc:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\start-local.ps1
```

Script chỉ dùng Docker PostgreSQL của TTCSN, chạy migration lặp an toàn, seed các account còn thiếu, build API rồi chờ `/health/ready`; sau đó chạy Vite và chờ `/login`. Nó từ chối ghi đè listener 8080/5173 và lưu PID riêng trong `.artifact_work/runtime`. Muốn thêm đúng fixture phòng/lịch demo đã bảo vệ, thêm `-WithBookingFixtures`; không dùng tùy chọn này cho dữ liệu trình diễn đang được chỉnh tay.

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\start-local.ps1 -WithBookingFixtures
```

Để dừng đúng API/client do script này tạo và dừng PostgreSQL nhưng **giữ volume dữ liệu**:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\stop-local.ps1
```

`stop-local.ps1` chỉ dừng PID có metadata/command line của TTCSN; không quét hay tắt tiến trình Node/Go khác. Dùng `-KeepDatabase` nếu chỉ muốn khởi động lại API/client.

Nếu API đã chạy và chỉ cần khởi động lại Vite sau khi sửa giao diện, dùng `-ClientOnly`; script vẫn từ chối nếu cổng 5173 đang có listener.

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\start-local.ps1 -ClientOnly
```

## 4. Mỗi lần chạy thủ công: API

Trong PowerShell thứ nhất:

```powershell
Set-Location C:\Users\choqu\OneDrive\Desktop\TTCSN
docker compose --env-file .\infra\.env.compose up -d --wait
Set-Location .\server
$env:GOCACHE = Join-Path (Get-Location).Path '..\.artifact_work\go-cache'
go run .\cmd\migrate up
if ($LASTEXITCODE -ne 0) { throw 'Migration thất bại' }
go run .\cmd\seed
if ($LASTEXITCODE -ne 0) { throw 'Seed thất bại' }
go run .\cmd\api
```

Goose áp dụng tuần tự toàn bộ migration hiện có đến v6 (schema nền, giờ PT, lịch phòng, ảnh cover, provenance Booking và index retention idempotency); chạy `up` lặp an toàn. Seed chỉ tạo tài khoản còn thiếu, không ghi đè mật khẩu đã có. Phòng mới cần đăng ký giờ mở cửa trước khi tạo phiên; phòng legacy giữ chế độ chuyển đổi đến lần công bố lịch đầu tiên. Hướng dẫn grid tuần/ngoại lệ: [21 — Lịch phòng](21_huong_dan_lich_phong.md).

Fixture Booking cho buổi demo là thao tác riêng, chỉ chạy khi đã seed Admin và muốn thêm hai tài khoản synthetic cùng ba phòng/ảnh/lịch DEMO. Lệnh yêu cầu database dev local đúng host/tên; không thay mật khẩu tài khoản tồn tại, không tự đặt booking hay điểm danh. Từ thư mục `server`:

```powershell
$localConfig = Get-Content .\.env.local
$databaseLine = $localConfig | Where-Object { $_ -match '^DATABASE_URL=' } | Select-Object -First 1
if (-not $databaseLine) { throw 'DATABASE_URL chưa có trong .env.local' }
$localDatabaseUrl = ($databaseLine -replace '^DATABASE_URL=', '').Trim('"')
if ($localDatabaseUrl -notmatch '^postgres://.*@127\.0\.0\.1:5433/ttcsn_gym_dev\?') {
  throw 'Từ chối chạy fixture: chỉ chấp nhận database dev local của project'
}
$previousDatabaseUrl = $env:DATABASE_URL
try {
  $env:DATABASE_URL = $localDatabaseUrl
  go run .\cmd\seed-booking-demo
  if ($LASTEXITCODE -ne 0) { throw 'Tạo fixture Booking thất bại' }
} finally {
  if ($null -eq $previousDatabaseUrl) { Remove-Item Env:DATABASE_URL -ErrorAction SilentlyContinue }
  else { $env:DATABASE_URL = $previousDatabaseUrl }
}
```

Tài khoản mới dùng mật khẩu `DemoPass123!`; fixture kiểm tra role/trạng thái/profile hiện hữu và dừng nếu dữ liệu không khớp, không sửa tay thay người dùng.

## 5. Mỗi lần chạy thủ công: giao diện

Trong PowerShell thứ hai:

```powershell
Set-Location C:\Users\choqu\OneDrive\Desktop\TTCSN\client
npm.cmd ci
if ($LASTEXITCODE -ne 0) { throw 'Cài frontend thất bại' }
npm.cmd run dev -- --host 127.0.0.1 --port 5173 --strictPort
```

Mở [ứng dụng GYM](http://127.0.0.1:5173). Kiểm tra API từ shell khác:

```powershell
Invoke-RestMethod http://127.0.0.1:8080/health/ready
```

Mật khẩu của tài khoản seed mới: `DemoPass123!`.

| Tài khoản | Quyền demo |
|---|---|
| `admin.demo` | Tài khoản, hội viên, phòng/ca/phiên và xét duyệt/lịch PT |
| `manager.demo` | Hội viên, phòng/ca/phiên và xét duyệt/lịch PT |
| `reception.demo` | Quản lý hội viên |
| `pt.demo` | Hồ sơ cá nhân, chuyên môn/chứng chỉ và lịch của mình |
| `member.demo` | Hồ sơ cá nhân; API đọc hội viên SELF khi đã có profile |

Token chỉ giữ trong memory; reload cần đăng nhập lại. TTL 15 phút. Khóa/reset/đổi role/logout thu hồi token cũ; logout thu hồi **mọi phiên** của tài khoản.

## 6. Các luồng CRUD đã triển khai

- Admin: tạo tài khoản → mở chi tiết → sửa liên hệ → khóa/mở; đổi role và reset dùng dialog riêng. Username giữ cố định; Admin cuối không được khóa/hạ quyền.
- Hội viên: tạo tài khoản MEMBER mới hoặc liên kết MEMBER chưa có hồ sơ. Trạng thái thẻ, hạn thẻ và trạng thái đăng nhập độc lập.
- Phòng/ca/phiên: tạo phòng và ca → tạo phiên theo ngày. Sức chứa 0 khi tạo lấy mặc định phòng; phiên chốt thời gian/sức chứa. Ca đã dùng chỉ được ngừng, không sửa giờ. Đổi phòng sang bảo trì/ngừng sẽ đóng phiên OPEN tương lai nếu không còn booking giữ chỗ; mở lại phòng không tự mở lại phiên.
- PT: tự lưu hồ sơ → chọn file PDF/PNG/JPEG tối đa 5 MiB → tải lên → gửi xét duyệt. Manager/Admin duyệt/từ chối; đình chỉ bị chặn khi còn booking giữ chỗ. DRAFT/REJECTED sửa được, PENDING/APPROVED/SUSPENDED chỉ xem. Duyệt không tự mở khóa tài khoản.
- Lịch PT: khung giờ tuần ISO thứ 1–7, khoảng nghỉ có offset +07:00, phân trang. Sửa lịch/nghỉ không được phá booking đã giữ. File chứng chỉ chỉ tải qua endpoint xác thực, không có URL public.

Booking và điểm danh có API/UI: vận hành ghi riêng attendance Hội viên/PT, Hội viên/PT chỉ đọc đúng scope, completion tự đưa buổi đã hết giờ thành `COMPLETED` hoặc `NO_SHOW`. Review cho Hội viên và trang đọc review của PT nằm trong portal riêng. PT xem `/pt/reviews` và `/pt/performance`; Admin/Manager xem `/management/reports` để lọc kỳ/phòng, xem chỉ số và tra cứu PT. Member/Lễ tân không có quyền báo cáo. Công thức, N/A, dữ liệu chưa phân loại và giới hạn được ghi ở [25 — kế hoạch chốt Attendance, Review và Report](25_ke_hoach_chot_attendance_review_report.md) cùng [báo cáo UI](reports/report_ui_verification_2026-10-14.md). Browser smoke local đủ 5 vai trò và một booking synthetic đã chạy; E2E vòng đời đầy đủ, accessibility và browser matrix vẫn chưa hoàn tất, xem [regression 18/10](reports/local_regression_2026-10-18.md).

## 7. Kiểm thử có DB thật

Integration test reset/rollback dữ liệu của **đúng DB test**, nên không chạy khi người khác đang dùng DB test.

```powershell
Set-Location C:\Users\choqu\OneDrive\Desktop\TTCSN\server
$env:GOCACHE = Join-Path (Get-Location).Path '..\.artifact_work\go-cache'
$composeText = Get-Content ..\infra\.env.compose -Raw
$composePassword = ($composeText -split '=', 2)[1].Trim()
$encodedPassword = [Uri]::EscapeDataString($composePassword)
$env:TEST_DATABASE_URL = "postgres://gym_app:$encodedPassword@127.0.0.1:5433/ttcsn_gym_test?sslmode=disable"
go test ./... -count=1
go vet ./...
Remove-Item Env:TEST_DATABASE_URL
Set-Location ..\client
npm.cmd test
npm.cmd run build
Set-Location ..
python .\scripts\generate_api_contracts.py --check
powershell -ExecutionPolicy Bypass -File .\docs\check_docs.ps1
```

`cmd/reset-test`, migration `down/reset` từ chối database ngoài đúng `ttcsn_gym_test`. Không chạy schema bootstrap hay rollback trên DB phát triển có dữ liệu.

Hợp đồng hiện hành: `docs/openapi_admin_crud.json`, `docs/GYM_Admin_CRUD.postman_collection.json`. Import collection, login đúng role rồi tự đặt các biến ID; không chạy cả collection một lượt vì có các thao tác đổi trạng thái/reset/logout. Collection đã sinh và đối chiếu router, chưa chạy bằng Postman runner. Các file tên Identity_Week4 giữ làm snapshot lịch sử.

Collection chỉ đọc `docs/GYM_Role_Smoke.postman_collection.json` kiểm tra login và `/me` của 5 vai trò, quyền Report của Admin/Manager và 403 cho Lễ tân/Hội viên. Nó chạy bằng image `postman/newman:alpine` trong Docker để không phụ thuộc Newman cục bộ với Node 24:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\run-api-role-smoke.ps1
```

Kết quả JSON nằm trong `.artifact_work/api-smoke/`. Collection này không tạo/sửa dữ liệu; mutation, transaction và race tiếp tục dùng PostgreSQL test suite.

Collection mutation Identity chạy riêng trên API tạm cổng `18080` và **chỉ** `ttcsn_gym_test`. Nó tạo một tài khoản Lễ tân synthetic, kiểm tra lock/unlock/reset và RBAC, thu audit rồi reset DB test. Không chạy khi một người khác đang dùng DB test:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\run-api-mutation-smoke.ps1
```

Report, giới hạn và đường dẫn artifact: [mutation smoke 18/10](reports/mutation_smoke_2026-10-18.md). Script không động tới instance dev đang chạy cổng 8080/5173.

## 7.1 Kiểm tra vận hành và backup local

Migration hiện tại là v6. Các key idempotency Booking/Calendar được replay tối đa 24 giờ và job dọn bản hết hạn khi khởi động rồi mỗi giờ. Chạy tải burst 100 concurrent, security smoke và restore drill bằng các script sau:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\invoke-api-load.ps1 -Concurrency 100 -RequestsPerUser 3
powershell -ExecutionPolicy Bypass -File .\scripts\invoke-api-load.ps1 -Concurrency 100 -RequestsPerUser 3 -WarmSessions -RampSeconds 20
powershell -ExecutionPolicy Bypass -File .\scripts\backup-restore-drill.ps1
```

Load script chỉ login/đọc dữ liệu và lưu report không chứa token trong `.artifact_work/load/`. Lệnh đầu là burst login + read; lệnh thứ hai tạo token tuần tự, ramp worker và chỉ đo read path. Không dùng lệnh thứ hai để che p95 của login burst. Backup script không restore đè development: target phải có tiền tố `ttcsn_gym_restore_` và chưa tồn tại. Kết quả local, p95 và giới hạn NFR/security/restore được ghi tại [báo cáo vận hành 18/10](reports/operations_nfr_restore_2026-10-18.md); không coi lệnh chạy thành NFR đạt nếu p95 vượt mục tiêu.

## 8. Dừng và xử lý lỗi

Dùng Ctrl+C trong hai shell backend/frontend, hoặc dùng `scripts/stop-local.ps1` nếu đã khởi động bằng script. Nếu chạy thủ công, từ thư mục gốc:

```powershell
docker compose --env-file .\infra\.env.compose stop
```

Volume được giữ. Không cần `down -v` để dừng demo.

| Lỗi | Cách xử lý |
|---|---|
| `dockerDesktopLinuxEngine ... cannot find` | Mở Docker Desktop, chờ Linux engine rồi chạy lại Compose. |
| Cổng đã có listener | Nếu demo đang chạy, mở URL hiện có; dừng đúng tiến trình trước khi chạy instance khác. |
| Thiếu secret/DB không kết nối | Kiểm tra cấu hình local, ưu tiên env shell, healthy của Compose và cổng 5433. |
| Bị 401 | Đăng nhập lại; token hết TTL hoặc đã bị thu hồi. |
| 409 khi sửa phòng/phiên/ca/lịch | Đọc lý do: lịch đã dùng, có chỗ đang giữ hoặc chồng giờ. |
| Upload 422 | Chọn file đúng nội dung PDF/PNG/JPEG, không chỉ đổi đuôi; tối đa 5 MiB. |
| `npm.ps1` bị chặn | Dùng `npm.cmd`. |
| Go báo lỗi cache ở project khác | Đặt GOCACHE project local như lệnh trên. |
