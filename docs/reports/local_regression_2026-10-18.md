# Regression local và smoke E2E — 18/10/2026

**Môi trường:** Windows local, Docker Desktop Linux engine, PostgreSQL 17 container `ttcsn-gym-postgres` tại `127.0.0.1:5433`, Go API `127.0.0.1:8080`, Vite `127.0.0.1:5173`. Database phát triển là `ttcsn_gym_dev`; integration test chỉ nhận `ttcsn_gym_test`.

## Chuẩn chạy local

- Docker Desktop ban đầu chưa chạy (`dockerDesktopLinuxEngine` không tồn tại). Sau khi khởi động Desktop, container PostgreSQL healthy.
- `scripts/start-local.ps1 -SkipSeed -NoClient`: PASS. Goose báo database ở version 5, build API và `/health/ready` thành công.
- `scripts/start-local.ps1 -ClientOnly`: PASS, Vite sẵn sàng tại `/login`.
- Script mới giữ PID trong `.artifact_work/runtime`, từ chối cổng 8080/5173 đang được dùng và `stop-local.ps1` chỉ dừng process có metadata TTCSN. Volume PostgreSQL không bị xóa.
- Smoke HTTP trực tiếp: readiness `ready`, client HTTP 200; đăng nhập và `GET /me` của `admin.demo`, `manager.demo`, `reception.demo`, `pt.demo`, `member.demo` đều PASS. `member.demo` gọi Report nhận 403.

## Newman role smoke

Collection mới: `../GYM_Role_Smoke.postman_collection.json`. Lệnh:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\run-api-role-smoke.ps1
```

Runner dùng `postman/newman:alpine` 6.1.3 trong Docker và gọi API local qua `host.docker.internal`; tránh lỗi tương thích của Newman cục bộ với Node 24. Kết quả:

| Mục | Kết quả |
|---|---:|
| Request | 15 / 15 PASS |
| Assertion | 15 / 15 PASS |
| Login + `/me` | 5 role PASS |
| Report | Admin/Manager 200; Lễ tân/Hội viên 403 |
| Login tài khoản không tồn tại | 401 |
| Thời gian phản hồi trung bình | 81 ms |

JSON report cục bộ: `.artifact_work/api-smoke/newman-role-smoke-20261018-004817.json`. Collection chỉ đọc, không tạo hay sửa dữ liệu.

## Regression kỹ thuật

| Lệnh | Kết quả |
|---|---|
| `go test ./... -count=1` với `TEST_DATABASE_URL=ttcsn_gym_test` | PASS |
| `go vet ./...` | PASS |
| `npm.cmd test -- --run` | 7 files, 45 tests PASS |
| `npm.cmd run build` | PASS; có cảnh báo main chunk 509.46 kB, chưa coi là benchmark/tối ưu hoàn tất |
| `python .\scripts\generate_api_contracts.py --check` | PASS, 76 operations |
| `docs\check_docs.ps1` | PASS, 20 REQ / 16 UC / 26 BR / 18 bảng |
| `git diff --check` | PASS |

Lượt Vitest đầu chạy cùng lúc với `go vet` và build đã timeout 5 case sau 5 giây (40/45 PASS). Chạy lại **độc lập** ngay sau đó PASS 45/45; chỉ kết quả độc lập được dùng cho regression này.

## Browser E2E local

Browser thật tại `http://127.0.0.1:5173` đã kiểm tra các state sau:

| Vai trò | Luồng thấy được | Kết quả |
|---|---|---|
| Admin | Login → `/admin/users`, danh sách 10 account; logout dialog thu hồi phiên | PASS |
| Member (`member.flow.demo`) | Catalog ảnh phòng → FLOW DEMO · Strength → chọn ngày/ca → review dialog → xác nhận | PASS; booking synthetic **#4**, `CONFIRMED` |
| Member | `/member/bookings` đọc lại #4 cùng ba booking fixture với trạng thái COMPLETED/CANCELLED/EXPIRED | PASS |
| PT (`pt.flow.demo`) | Login → `/pt/schedule`; lịch tuần 7 ngày × 24 giờ và route requests/reviews/performance hiển thị | PASS |
| Manager | Login → quản lý phòng → `/management/reports`; report hiện số liệu snapshot, filter kỳ/phòng/PT | PASS |
| Receptionist | Login → `/management/members`; danh sách hội viên, search/filter/action hiển thị ở layout mobile | PASS |

Các direct route ngoài quyền (`PT → /member/rooms`, `Manager → /admin/users`, `Receptionist → /management/reports`) đều bị chặn ở client và đưa về `/login`; API runner xác nhận Report là 403 cho hai role bị cấm. Token không được đưa vào report/log.

Responsive đã đo scroll ngang trang thực:

| Trang / viewport | `scrollWidth > innerWidth` |
|---|---|
| Manager Report, 375 px | Không |
| Member catalog, 375 px | Không |
| Member catalog, 768 / 1024 / 1440 px | Không / Không / Không |

Viewport browser đã được reset sau kiểm tra. Booking #4 là mutation duy nhất tạo trong lượt browser và dùng account/phòng `FLOW DEMO` chuyên cho demo; không sửa tài khoản, phòng, lịch hay dữ liệu người dùng thật.

## Phát hiện và xử lý trong lượt chạy

- Console ghi nhận một cảnh báo React ở lúc mở chi tiết booking #4: `AttendancePanel` và `ReviewPanel` là hai sibling cùng mang `key={b.id}`. Đã bỏ hai `key` dư thừa, chạy lại Vitest 45/45 và production build. Sau đó đi từ danh sách tới chi tiết #4 trong browser; không xuất hiện cảnh báo khóa trùng mới (console vẫn lưu bản ghi lịch sử lúc 00:38:44 trước khi sửa).
- Có một lần `GET /bookings` của PT trả 500 ở lần tải đầu lịch tuần. API không log SQL/chi tiết nhạy cảm theo chính sách log; lần tải lại ngay sau đó trả 200. Kiểm tra lại bằng đúng account PT và filter tuần `2026-10-18` đến `2026-10-24`: **20/20** request liên tiếp trả 200. Đây là quan sát không tái hiện được, vì vậy vẫn cần giữ cổng E2E/PT và tải ở phần giới hạn, không coi là đã chứng minh độ ổn định tải.
- Để loại trừ lỗi dữ liệu ảnh ở trạng thái phiên hợp lệ, đã gọi trực tiếp đủ 7 phòng có cover bằng account `member.flow.demo`; **7/7** trả HTTP 200, `image/jpeg` và payload khác 0 byte. Các 401/500 cover trước đó xuất hiện khi chuyển route trực tiếp làm phiên memory bị xóa và các request ảnh cũ bị hủy, không lặp lại với token hợp lệ.

## Giới hạn còn mở

- Chưa có full browser vòng đời PT accept/reject → attendance → completion theo đồng hồ thật → review; các quy tắc đó hiện có Go/PostgreSQL tests.
- Keyboard/focus audit toàn bộ màn hình, browser matrix Chrome/Edge/Firefox, E2E lỗi mạng/401 giữa form và viewport mọi trang chưa hoàn thành.
- Chưa chạy NFR workload 100 concurrent, backlog job >1.000, retention idempotency, security review đầy đủ hoặc backup/restore drill. Cảnh báo bundle >500 kB cần đo/ưu tiên sau các cổng chức năng.
- Lượt này không phải xác nhận khảo sát, họp, phê duyệt hoặc mức độ hoàn thành của từng thành viên nhóm.
