# Review UI — kiểm chứng 14/10/2026

Phạm vi: form đánh giá Hội viên trong chi tiết booking; điều kiện hiển thị; readback cho Hội viên, PT và vận hành; rating/count trong danh sách PT và lúc chọn PT. Kế thừa backend Review ở [báo cáo chặng C](review_backend_verification_2026-10-14.md).

## Kết quả kiểm tra

| Kiểm tra | Kết quả |
|---|---|
| `npm.cmd test -- --reporter=dot` | PASS: 40 tests, 6 files. Gồm luồng gửi/readback, chặn form khi thiếu điểm danh một bên, PT đọc review không lộ tên hội viên, lỗi API giữ form và hiển thị N/A khi PT chưa có review. |
| `npm.cmd run build` | PASS: TypeScript và Vite. Main chunk 508.44 kB minified / 159.23 kB gzip; Vite giữ cảnh báo chunk vượt 500 kB. |
| `git diff --check` | PASS. |

Hội viên chỉ thấy form khi booking đã `COMPLETED`, có PT, thời điểm kết thúc đã qua theo thời gian server và API attendance xác nhận cả hai người có mặt. Backend vẫn quyết định eligibility cuối cùng. Chọn sao dùng nhóm radio 1–5 với accessible name, bàn phím và focus; nhận xét tối đa 1000 ký tự. Lưu thành công chuyển ngay sang trạng thái readback; lỗi từ API giữ nội dung và cho phép thử lại.

PT được gán và các vai trò vận hành đọc review trong chi tiết booking qua endpoint đã phân quyền. Màn hình chỉ hiển thị sao, nhận xét, thời gian và nhãn chung “Hội viên”; danh sách PT và lựa chọn PT hiển thị `rating_average`/`review_count`, hoặc “Chưa có đánh giá”. Hội viên vẫn dùng sắc xanh lá, PT dùng xanh dương.

## Chưa kiểm chứng

Không chạy browser visual QA ở 375px/desktop, keyboard thực, zoom hoặc thao tác UI → PostgreSQL; API/client vẫn dừng theo yêu cầu. Vitest kiểm tra component bằng API mock; tính đúng nghiệp vụ và DB được xác nhận riêng trong báo cáo backend. Build warning của main bundle còn tồn tại và không ảnh hưởng kết quả build.
