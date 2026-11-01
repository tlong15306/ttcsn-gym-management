# GYM Management — TTCSN

Ứng dụng quản lý phòng gym với frontend React/TypeScript, backend Go/Gin và PostgreSQL. Hỗ trợ 5 vai trò: quản trị viên, quản lý, lễ tân, huấn luyện viên và hội viên.

## Chức năng

- Trang chủ, danh sách phòng/HLV và lịch trống công khai.
- Quản lý tài khoản, phân quyền, hội viên, HLV, phòng và lịch hoạt động.
- Đặt lịch một hoặc nhiều ca liền kề, kéo chọn/chỉnh giờ, tự tập hoặc yêu cầu HLV; giữ bản nháp qua đăng nhập.
- HLV phản hồi yêu cầu, hủy lịch theo điều kiện, điểm danh, hoàn thành/vắng mặt, đánh giá và báo cáo.

## Cấu trúc

| Thư mục | Nội dung |
|---|---|
| `client/` | Frontend, ảnh/font tĩnh, frontend tests và lockfile npm |
| `server/` | API, nghiệp vụ, migration, seed, unit/integration tests, Go modules |
| `infra/` | Mẫu cấu hình Compose và khởi tạo DB test |
| `scripts/` | Chạy/dừng local, kiểm tra contract, smoke test và backup/restore |
| `docs/` | SRS, ERD, API, hướng dẫn chạy/demo và báo cáo kiểm chứng |

## Chạy local trên Windows

Cần Docker Desktop (Linux engine), Go 1.26.5 và Node.js 24/npm. Các cổng mặc định: PostgreSQL `5433`, API `8080`, frontend `5173`.

```powershell
git clone https://github.com/tlong15306/ttcsn-gym-management.git
Set-Location ttcsn-gym-management
Copy-Item infra/.env.compose.example infra/.env.compose
Copy-Item server/.env.local.example server/.env.local
```

Điền cấu hình riêng trước khi chạy:

1. Đặt `POSTGRES_PASSWORD` trong `infra/.env.compose`.
2. Đặt cùng mật khẩu trong `DATABASE_URL` của `server/.env.local`; URL-encode nếu có ký tự đặc biệt. Giữ DB `ttcsn_gym_dev`, địa chỉ `127.0.0.1:5433`.
3. Thay `JWT_SECRET` bằng secret ngẫu nhiên riêng, ít nhất 32 ký tự. Không đưa cấu hình thật lên Git.

```powershell
Push-Location client
npm.cmd ci
Pop-Location
powershell -ExecutionPolicy Bypass -File ./scripts/start-local.ps1
```

Mở <http://127.0.0.1:5173/>. Script chạy migration và tạo tài khoản nền còn thiếu. Để tạo thêm phòng/HLV/Hội viên và các tình huống demo trên DB local mới:

```powershell
Push-Location server
go run ./cmd/seed-demo-scenarios
Pop-Location
```

Tài khoản đặt lịch chính: `demo.member.01`. Mật khẩu demo mặc định cho tài khoản mới: `DemoPass123!`; seed không đổi mật khẩu tài khoản đã tồn tại. Đây là dữ liệu minh họa local, không dùng tài khoản/mật khẩu demo cho production. Xem [hướng dẫn demo](docs/33_huong_dan_demo.md) để chọn tài khoản và luồng phù hợp.

Mở lại với dữ liệu đang có và dừng mà giữ volume:

```powershell
powershell -ExecutionPolicy Bypass -File ./scripts/start-local.ps1 -SkipSeed
powershell -ExecutionPolicy Bypass -File ./scripts/stop-local.ps1
```

Hướng dẫn cấu hình chi tiết: [chạy local](docs/17_huong_dan_chay_local.md).

## Kiểm tra

```powershell
Push-Location client
npm.cmd test -- --maxWorkers=1 --no-file-parallelism --testTimeout=20000
npm.cmd run build
Pop-Location
Push-Location server
go test -p 1 ./... -count=1 -timeout 300s
go vet ./...
Pop-Location
python scripts/generate_api_contracts.py --check
powershell -ExecutionPolicy Bypass -File docs/check_docs.ps1
```

Integration test chỉ chạy khi có `TEST_DATABASE_URL`; dùng **`ttcsn_gym_test` riêng** vì suite có reset/ghi dữ liệu. Không trỏ test vào `ttcsn_gym_dev`. Lệnh Go không có DB test không thay thế bằng chứng integration.

## Phạm vi bàn giao

Source, test, migration, lockfile, cấu hình mẫu, tài nguyên giao diện và tài liệu được quản lý bằng Git. `.env` thật, secret, database/backup, upload riêng, dependencies, build/cache/log, công cụ AI local và các bản xuất bài tập trong `outputs_document/` được bỏ qua; các file local vẫn được giữ trên máy.

Đây là bản phát triển/demo local. [Báo cáo U2–U4 ngày 01/11/2026](docs/reports/public_booking_u2_u4_2026-11-01.md) ghi kết quả kiểm chứng và giới hạn; hiệu năng/NFR và nghiệm thu production còn mở. Bản đồ tài liệu: [docs/README.md](docs/README.md).
