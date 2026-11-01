# Review backend và rating PT — kiểm chứng 14/10/2026

Phạm vi: Review theo REQ-19/UC-05, quyền xem, điều kiện buổi tập, audit, danh sách công khai an toàn, rating PT và hợp đồng API. Không khởi động API/client; không chạy migration hoặc thay đổi dữ liệu dev. Schema `pt_reviews` đã có unique booking constraint nên không cần migration mới.

## Kết quả kiểm tra

| Kiểm tra | Kết quả |
|---|---|
| `go test ./... -count=1` với `TEST_DATABASE_URL` được xác nhận trỏ đúng `ttcsn_gym_test` | PASS. Integration kiểm tra trên PostgreSQL thật; chỉ database test được Goose reset. |
| `TestReviewAPIEligibilityPrivacyAndRatingIntegration` | PASS: chỉ MEMBER tạo; 1–5 sao; reject field giả; member/PT attendance đều bắt buộc; PT phải có; booking COMPLETED và đã kết thúc. Missing MEMBER, missing PT, tự tập, NO_SHOW và buổi chưa kết thúc đều trả `REVIEW_NOT_ELIGIBLE`. |
| Race gửi đồng thời 12 request cho cùng booking | PASS: đúng 1 response 201, 11 response 409 `REVIEW_EXISTS`, một review và một audit cho booking đó. |
| Transaction/audit rollback | PASS: lỗi lưu audit làm review rollback; không tồn tại review mồ côi. |
| Quyền và privacy | PASS: MEMBER SELF, PT ASSIGNED và ba vai trò vận hành đọc được; MEMBER/PT ngoài phạm vi nhận 404. Danh sách chỉ trả sao, comment, ngày và nhãn “Hội viên”, không có ID người dùng/booking hoặc contact. |
| Pagination và rating | PASS: `page_size=1` giữ đúng tổng số; PT profile và booking options có cùng rating trung bình/count tính từ review hợp lệ. |
| `go vet ./...` | PASS. |
| `python .\scripts\generate_api_contracts.py --check` | PASS: 74 operations, route map, schema references, OpenAPI và Postman đồng bộ. |
| `powershell -ExecutionPolicy Bypass -File .\docs\check_docs.ps1` | PASS: 15 file bắt buộc, 20 requirements, 16 use cases, 26 business rules, 18 schema tables. |
| `git diff --check` | PASS. |

Review chỉ tạo một lần cho booking của Hội viên đang đăng nhập; body chỉ nhận `rating` và `comment`, backend suy ra member/PT. Booking phải ở `COMPLETED`, đã hết giờ theo database clock, có PT được gán và đủ hai bản ghi attendance. POST lưu review và audit trong cùng transaction; lỗi audit rollback toàn bộ. GET review theo booking trả `data:null` nếu chưa đánh giá. Đánh giá không có API sửa/xóa.

Danh sách PT công khai áp dụng cho hồ sơ `APPROVED` và tài khoản `ACTIVE`; PT chỉ xem danh sách của mình, còn Lễ tân/Quản lý/Admin xem toàn bộ theo ma trận quyền. Projection không trả ID booking, ID người đánh giá, tên tài khoản, email hay điện thoại. `rating_average` và `review_count` được tính trực tiếp từ cùng tập review hợp lệ; không dùng trường cache hoặc Trust Score.

## Chưa làm trong chặng này

Review UI, Reports API/UI, attendance rate BR-22 và browser QA vẫn thuộc chặng D–F của kế hoạch 25. Không khởi động lại các server. Lượt này không sửa/reset DB dev, Excel, tài khoản hoặc dữ liệu fixture hiện có.

Hợp đồng API nằm trong `openapi_admin_crud.json` và `GYM_Admin_CRUD.postman_collection.json`; hướng dẫn khởi động local hiện có tại [17 — Chạy demo GYM CRUD trên PowerShell](../17_huong_dan_chay_local.md). Kết quả này chỉ xác nhận Review backend trong môi trường test local, không phải nghiệm thu toàn bộ MVP GYM.
