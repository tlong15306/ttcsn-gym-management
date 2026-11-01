# Kiểm tra trước commit demo Identity

Ngày ghi: **11/10/2026**. Phạm vi: nền tảng và Identity hiện có, trước khi triển khai nghiệp vụ GYM tiếp theo.

| Kiểm tra thực hiện | Kết quả | Giới hạn |
|---|---|---|
| `go test ./... -count=1` với `TEST_DATABASE_URL` trỏ đúng `ttcsn_gym_test` | PASS | Chạy các case hiện có; không đồng nghĩa mọi case trong kế hoạch đã được viết. |
| `go vet ./...` | PASS | Kiểm tra tĩnh Go. |
| Repository PostgreSQL integration, nằm trong Go suite trên | PASS | Migration rollback/up trên DB test; tạo user, duplicate username, login, lock/auth_version, chặn khóa Admin cuối cùng và audit count. Chưa có đầy đủ case race/audit failure/reset/role. |
| JWT thiếu `exp` và migration command safety | PASS | Unit regression mới; `down`/`reset` ngoài DB test bị từ chối, command không hỗ trợ bị từ chối. |
| `npm.cmd test` | PASS: 1 file, 1 test | Test form đăng nhập; chưa kiểm chứng toàn bộ thao tác Admin hoặc responsive. |
| `npm.cmd run build` | PASS | TypeScript và Vite build; không chứng minh thao tác qua browser. |
| `powershell -ExecutionPolicy Bypass -File .\docs\check_docs.ps1` | PASS | 15 file bắt buộc, 20 REQ, 16 UC, 26 BR, 14 bảng bootstrap; không kiểm chứng toàn bộ hành vi runtime. |

Go được chạy với cache riêng tại `.artifact_work/go-cache` thay cho cache cũ trỏ sang project khác bị lỗi quyền. Vitest lần đầu gặp lỗi quyền rename cache tạm trong sandbox; chạy lại ngoài sandbox đã PASS. PostgreSQL riêng của TTCSN dùng cổng 5433; integration không reset DB phát triển hoặc `master-golang`.

Không chạy lại toàn bộ luồng HTTP 5 vai trò hoặc thao tác UI trong lần kiểm tra này. Các cổng còn thiếu được đưa vào kế hoạch tiếp theo; không dùng báo cáo này để tuyên bố nghiệm thu toàn bộ MVP.
