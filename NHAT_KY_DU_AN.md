# NHẬT KÝ HỌC VÀ PHÁT TRIỂN DỰ ÁN GYM

> Đây là nhật ký cá nhân ghi lại quá trình tôi học và tự xây dựng **Hệ thống Quản lý Phòng tập GYM**. Mục tiêu của file là giúp tôi nhớ mình đã làm gì, tại sao làm như vậy, đã hiểu phần nào, gặp lỗi gì và giải quyết ra sao.

## 1. Thông tin dự án

| Nội dung | Thông tin |
|---|---|
| Tên dự án | Hệ thống Quản lý Phòng tập GYM |
| Học phần | Thực tập cơ sở ngành — TTCSN |
| Nhóm | Nhóm 1 |
| Công nghệ dự kiến | React + Tailwind CSS, Go + Gin, PostgreSQL 16+ |
| Thời điểm bắt đầu ghi nhật ký | 10/10/2026 |
| Trạng thái hiện tại | Hoàn thiện baseline tài liệu v0.3, chưa bắt đầu viết mã nguồn ứng dụng |

## 2. Cách tôi sẽ học và làm dự án

Tôi không muốn nhận một dự án đã được viết hoàn chỉnh mà không hiểu bên trong. Tôi sẽ làm theo từng bước nhỏ:

1. Hiểu mục tiêu của bước hiện tại.
2. Tự gõ đoạn code được hướng dẫn.
3. Chạy chương trình và quan sát kết quả.
4. Ghi lại lỗi, nguyên nhân và cách sửa.
5. Tự giải thích lại đoạn code bằng lời của mình.
6. Chỉ chuyển sang bước mới khi đã hiểu phần đang làm.

Codex đóng vai trò hướng dẫn, giải thích, review và hỗ trợ tìm lỗi. Tôi là người trực tiếp viết và chạy code để học.

### Nền tảng đã học và cách điều chỉnh hướng dẫn — 10/10/2026

- Theo thông tin tôi cung cấp, tôi đã học Go cơ bản, syntax và CRUD cơ bản trong project `C:\Users\choqu\OneDrive\Desktop\GO\ginrr`.
- Khung tham khảo backend là `C:\Users\choqu\OneDrive\Desktop\GO\khoa-hoc-lap-trinh-restful-api-voi-gin-framework\lesson12-project-shopping-cart`.
- Codex đã đọc các file khởi động, lắp ghép module, handler, service và repository của khung tham khảo; chưa chạy project đó trong buổi này.
- Hướng dẫn tiếp theo tập trung vào cấu trúc backend, nối dependency và áp dụng nghiệp vụ GYM. Không cần học lại chương trình Hello World hoặc syntax từ đầu.
- Giữ cách chia trách nhiệm Handler → Service/Usecase → Repository. Theo kiến trúc GYM đã chốt, use case nhận `context.Context`, không nhận `gin.Context`; dependency DB được truyền vào constructor thay vì truy cập DB global.
- Tạo dần các phần khi cần dùng. Redis, RabbitMQ và mail trong khung mẫu chưa được chọn làm phụ thuộc bắt buộc cho bước khởi tạo GYM.
- Trạng thái mã nguồn GYM: đã tạo Go module và khung API Gin tối thiểu; PostgreSQL và nghiệp vụ GYM chưa được kết nối.

## 3. Bối cảnh trước khi bắt đầu code

Các mục dưới đây được tổng hợp từ tài liệu trong workspace. Chúng mô tả quá trình chuẩn bị dự án; không tự động chứng minh phần mềm đã được triển khai.

### Giai đoạn định hướng đề tài

- Chọn đề tài Hệ thống Quản lý Phòng tập GYM.
- Xác định mục tiêu quản lý hội viên, PT, phòng tập, lịch đặt, điểm danh và báo cáo.
- Tham khảo các phần mềm GYM thương mại để hình dung chức năng.
- Chọn hướng xây dựng web responsive cho một cơ sở GYM.

### Giai đoạn lập kế hoạch và phân tích ban đầu

- Tạo kế hoạch làm việc nhóm và khung tiến độ 10 tuần.
- Xây dựng backlog yêu cầu chức năng.
- Xác định 5 tác nhân: Hội viên, PT, Lễ tân, Quản lý và Admin.
- Viết SRS, use case, ma trận phân quyền, ERD, schema SQL và tài liệu kiến trúc.
- Chọn React/Tailwind cho frontend, Go/Gin cho backend và PostgreSQL cho database.

### Giai đoạn rà soát tài liệu — 10/10/2026

Tôi yêu cầu Codex đọc toàn bộ package tài liệu và lập kế hoạch hoàn thiện vì phần phân tích hệ thống, tác nhân và database còn khó bảo trì/nâng cấp.

Kết quả rà soát phát hiện các vấn đề chính:

- Quyền của các vai trò chưa thống nhất giữa các tài liệu.
- Mã use case trong SRS và tài liệu use case bị lệch nhau.
- Trạng thái booking trong nghiệp vụ và schema không đồng nhất.
- Chưa tách điểm danh Hội viên và PT.
- Review PT chưa liên kết với booking đã hoàn thành.
- Chưa mô hình hóa lịch làm việc/nghỉ của PT.
- Schema cũ dùng lệnh `DROP TABLE`, chưa phù hợp để nâng cấp và giữ dữ liệu.
- Một số tuyên bố hiệu năng chưa có kết quả kiểm thử.

### Hoàn thiện baseline tài liệu v0.3 — 10/10/2026

Sau khi xem kế hoạch, tôi đồng ý sử dụng các quyết định D01–D10 để viết lại package tài liệu.

Các kết quả đã có:

- Chốt phạm vi MVP một cơ sở GYM và 5 vai trò.
- Chuẩn hóa 20 yêu cầu, 16 use case và 26 quy tắc nghiệp vụ.
- Tạo một ma trận phân quyền chính thức.
- Chuẩn hóa vòng đời booking:
  `PENDING_PT`, `CONFIRMED`, `REJECTED`, `EXPIRED`, `CANCELLED`, `COMPLETED`, `NO_SHOW`.
- Tách `room_sessions`, attendance từng người, review theo booking, idempotency và audit.
- Viết lại schema PostgreSQL gồm 14 bảng và không còn `DROP TABLE`.
- Bổ sung tài liệu API, kiến trúc, backlog, kế hoạch kiểm thử và runbook migration/restore.
- Viết script `docs/check_docs.ps1` để kiểm tra tính nhất quán tài liệu.

Kết quả kiểm tra tài liệu:

```text
Result              : PASS
RequiredFiles       : 15
Requirements        : 20
UseCases            : 16
BusinessRules       : 26
SchemaTables        : 14
DropTableStatements : 0
```

Giới hạn hiện tại:

- Chưa có source code backend/frontend trong workspace.
- Chưa chạy schema trên PostgreSQL 16+.
- Chưa có migration thực thi từ schema cũ.
- Chưa chạy unit test, API test, E2E, load test hoặc restore drill.

### Khởi tạo Git baseline — 10/10/2026

- Khởi tạo Git repository local với nhánh chính `main`.
- Thiết lập `.gitignore` để không đưa `node_modules`, `.env`, khóa bảo mật, dữ liệu PostgreSQL cục bộ và artifact của công cụ vào repository.
- Tạo commit đầu tiên với thông điệp `docs: establish Gym project baseline v0.3`.
- Commit này chứa package tài liệu v0.3, nhật ký dự án, script kiểm tra tài liệu và các file Excel đầu ra hiện có.
- Đã publish lên repository riêng tư: `https://github.com/tlong15306/ttcsn-gym-management`.
- Nhánh local `main` đã theo dõi nhánh `origin/main` trên GitHub.

## 4. Điểm bắt đầu học code

Thứ tự học và phát triển đã thống nhất:

1. Kiểm tra môi trường Go, Node.js, npm, PostgreSQL và Git.
2. Khởi tạo Go module và entry point `cmd/api`, dựa trên nền tảng Go đã học.
3. Cài Gin, tách router/handler và tự viết endpoint `/health/live` đầu tiên.
4. Hiểu cách nối Handler → Usecase → Repository qua chức năng GYM nhỏ.
5. Khởi tạo PostgreSQL test và kiểm tra schema v0.3.
6. Kết nối Go với PostgreSQL.
7. Làm Identity: đăng nhập, tài khoản và phân quyền.
8. Làm dữ liệu nền: Hội viên, PT, Phòng, Ca và Phiên phòng.
9. Làm Booking core: đặt, hủy, PT phản hồi và chống race condition.
10. Làm Attendance, Review và Report.
11. Khởi tạo frontend React và tích hợp từng API.
12. Chạy kiểm thử, sửa lỗi và chuẩn bị demo.

## 5. Nhật ký theo buổi

### Buổi 00 — Chuẩn bị môi trường

**Ngày:** Chưa thực hiện  
**Mục tiêu:** Kiểm tra máy đã sẵn sàng để bắt đầu code.

Các lệnh sẽ chạy trong PowerShell:

```powershell
go version
node --version
npm --version
psql --version
git --version
```

**Kết quả thực tế:**

- Node.js: `v24.19.0` — đã có.
- npm: `11.17.0` — đã có.
- Git: `2.55.0.windows.3` — đã có.
- Go: `go1.26.5 windows/386` — đã có; sẽ dùng để học backend. `386` nghĩa là Go đang chạy bản 32-bit, vẫn dùng được cho project này.
- Docker Engine: `29.8.1` — đã có.
- Docker Compose: `v5.5.1` — đã có.
- `psql` trên Windows host: chưa có trong `PATH`.
- PostgreSQL Docker: container `postgres-db`, image `postgres:17.11-alpine3.24`, đang chạy và map cổng host `5432` vào cổng PostgreSQL trong container.

**Điều tôi đã hiểu:** `psql` là command-line client để kết nối PostgreSQL. Khi PostgreSQL chạy trong Docker, có thể dùng `psql` nằm sẵn bên trong container; không bắt buộc cài `psql` trực tiếp trên Windows host ở giai đoạn này. Image PostgreSQL `17.11` vẫn đáp ứng tài liệu thiết kế yêu cầu PostgreSQL 16 trở lên.  
**Lỗi gặp phải:** Lệnh `docker ps --format ...` báo Docker Desktop Linux Engine chưa chạy; lệnh `docker ps` ngay sau đó thành công và cho thấy container vừa khởi động.  
Lệnh kiểm tra PostgreSQL bằng user `postgres` báo `FATAL: role "postgres" does not exist`.  
**Cách giải quyết:** Đợi/khởi động Docker Desktop rồi chạy lại lệnh Docker. Với PostgreSQL, không tạo role mới vội; kiểm tra biến `POSTGRES_USER` và `POSTGRES_DB` của container để dùng đúng user/database đã khởi tạo.  
Kết quả kiểm tra: `POSTGRES_USER=root`, `POSTGRES_DB=master-golang`.  
Kết quả xác nhận server: PostgreSQL `17.11`.  
**Trạng thái Buổi 00:** Hoàn thành.  
**Bước tiếp theo:** Khởi tạo Go module trong thư mục `server` và hiểu file `go.mod`.

### Buổi 01 — Khởi tạo API Gin và liveness endpoint

**Ngày:** 10/10/2026
**Mục tiêu:** Tạo module backend GYM, tách entry point khỏi router/handler và xác nhận API chạy được.

**Việc đã làm:**

- Tạo module `github.com/tlong15306/ttcsn-gym-management/server` tại `server/go.mod` bằng Go `1.26.5`.
- Cài Gin và tạo `cmd/api/main.go` làm entry point khởi động HTTP server.
- Tạo `internal/delivery/http/router/router.go` với `gin.New()`, `gin.Logger()`, `gin.Recovery()` và `SetTrustedProxies(nil)`.
- Tạo `internal/delivery/http/handler/health_handler.go`; handler trả JSON `{ "status": "ok" }` cho `GET /health/live`.
- Phát hiện file placeholder rỗng `internal/app/app.go` khiến `go test ./...` lỗi vì thiếu khai báo `package`; file này đã được xóa vì chưa thuộc cấu trúc đang dùng.

**Kết quả kiểm tra:**

- API chạy tại `http://localhost:8080`.
- `GET /health/live` đã được kiểm tra bằng Postman và Gin log ghi HTTP `200`.
- `go test ./...` chạy thành công cho `cmd/api`, `handler` và `router`.

**Điều tôi đã hiểu:** `main` chỉ khởi động chương trình; router đăng ký URL và middleware; handler đọc/trả HTTP response. Liveness kiểm tra process API còn phản hồi, không kiểm tra database. Endpoint readiness kiểm tra PostgreSQL sẽ được thêm sau khi có cấu hình/kết nối DB.

**Trạng thái Buổi 01:** Hoàn thành.
**Bước tiếp theo:** Thêm cấu hình môi trường không chứa secret và khởi tạo kết nối PostgreSQL cho endpoint readiness.

## 6. Mẫu ghi cho mỗi buổi học/code

Sao chép mẫu này xuống cuối file sau mỗi buổi:

````markdown
### Buổi XX — Tên nội dung

**Ngày:** DD/MM/YYYY  
**Thời gian học:** ...  
**Mục tiêu:** ...

#### Kiến thức đã học

- ...

#### Việc tôi đã tự làm

- ...

#### File đã tạo hoặc sửa

- `đường/dẫn/file`: nội dung thay đổi.

#### Lệnh đã chạy

```text
lệnh ở đây
```

#### Kết quả kiểm tra

- [ ] Chạy được chương trình.
- [ ] Hiểu luồng xử lý chính.
- [ ] Tự giải thích lại được đoạn code.
- [ ] Đã ghi lại lỗi và cách sửa.

#### Lỗi gặp phải và cách sửa

| Lỗi/triệu chứng | Nguyên nhân | Cách sửa | Tôi học được gì |
|---|---|---|---|
| ... | ... | ... | ... |

#### Điều tôi còn chưa hiểu

- ...

#### Bước tiếp theo

- ...
````

## 7. Bảng tiến độ học tập

| Mốc | Trạng thái | Bằng chứng | Điều tôi cần giải thích được |
|---|---|---|---|
| Môi trường phát triển | Hoàn thành | Node 24.19.0, npm 11.17.0, Go 1.26.5, Git 2.55.0, Docker 29.8.1, PostgreSQL 17.11 | Vai trò của Go, Node, PostgreSQL, Docker và Git |
| Go cơ bản + module | Hoàn thành cho GYM | `server/go.mod`; module `github.com/tlong15306/ttcsn-gym-management/server` | Module GYM, entry point và import giữa các package |
| Gin `/health/live` | Hoàn thành | Postman và Gin log HTTP 200; `go test ./...` pass | Cách nối router/handler và ý nghĩa liveness |
| Cấu trúc backend | Đang bắt đầu | `main` → `router` → `HealthHandler` | Handler, Usecase, Repository khác nhau thế nào |
| PostgreSQL/schema | Chưa bắt đầu | DDL chạy trên DB test | bảng, PK, FK, CHECK, UNIQUE, index |
| Auth/RBAC | Chưa bắt đầu | test login/quyền | hash, token, authentication, authorization |
| Master data | Chưa bắt đầu | API Member/PT/Room/Session | validate và vòng đời dữ liệu |
| Booking core | Chưa bắt đầu | test race/conflict | transaction, lock, idempotency |
| Attendance/Review | Chưa bắt đầu | test eligibility | dữ liệu nguồn và dữ liệu dẫn xuất |
| Frontend React | Chưa bắt đầu | UI gọi API | component, state, route, API client |
| Hoàn thiện/demo | Chưa bắt đầu | test report + demo | hệ thống hoạt động end-to-end ra sao |

## 8. Quy tắc giữ nhật ký hữu ích

- Ghi điều đã làm thật; kế hoạch luôn để trạng thái “Chưa thực hiện”.
- Không ghi mật khẩu, token, chuỗi kết nối hoặc secret vào nhật ký.
- Khi gặp lỗi, giữ lại thông báo lỗi quan trọng và lệnh đã chạy.
- Không chỉ ghi “đã làm xong”; phải ghi mình hiểu được gì.
- Sau mỗi chức năng, liên kết tới REQ/UC/BR tương ứng trong `docs/`.
- Khi Codex sửa hộ một lỗi, tôi đọc diff và tự giải thích lại trước khi tiếp tục.
- Nếu kết quả chưa được chạy thử, ghi rõ “chưa kiểm chứng”.

## 9. Tài liệu cần đọc lại

- `docs/README.md`: bản đồ package tài liệu.
- `docs/00_quyet_dinh_va_quy_tac_nghiep_vu.md`: phạm vi và quy tắc.
- `docs/07_SRS_v0.3.md`: hệ thống cần làm gì.
- `docs/11_kien_truc_he_thong.md`: code được chia thành các phần nào.
- `docs/12_dac_ta_API.md`: frontend/backend giao tiếp ra sao.
- `docs/15_ke_hoach_kiem_thu.md`: chứng minh code đúng bằng cách nào.
