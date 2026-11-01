# Review và báo cáo UI — kiểm chứng 14/10/2026

Phạm vi chặng F của [kế hoạch 25](../25_ke_hoach_chot_attendance_review_report.md): báo cáo vận hành Admin/Manager, hiệu suất PT cá nhân, trang review riêng của PT, điều hướng theo role và runbook. Tận dụng API đã kiểm chứng ở [báo cáo Report API](report_api_verification_2026-10-14.md).

Commit triển khai F: `49b423a`.

## Kết quả

| Kiểm tra | Kết quả |
|---|---|
| `npm.cmd test -- --run` | PASS: 45 tests / 7 files, gồm 5 test mới cho lọc kỳ/phòng, validation kỳ, lỗi/thử tải lại, giá trị N/A, review PT ẩn danh và phân trang. |
| `npm.cmd run build` | PASS: TypeScript + Vite production build. Chunk trang `Reports` 18.56 kB minified / 5.82 kB gzip. |
| `git diff --check` | PASS. |
| `python .\scripts\generate_api_contracts.py --check` | PASS; F không đổi API. |
| `powershell -ExecutionPolicy Bypass -File .\docs\check_docs.ps1` | PASS sau cập nhật kế hoạch, mục lục và runbook. |

Build còn cảnh báo chunk chính lớn hơn ngưỡng Vite: 509.46 kB minified / 159.47 kB gzip. Trang báo cáo tải tách riêng bằng dynamic import.

## Hành vi đã triển khai

- `/management/reports` nằm sau route guard ADMIN/MANAGER; lọc ngày và phòng, hiển thị tổng booking, no-show, seat-minute, attendance, phân bố trạng thái, hủy theo nguồn và dữ liệu unclassified. Tra cứu PT dùng chung kỳ và hiển thị tử số/mẫu số, excluded, unclassified và pending completion.
- `/pt/performance` và `/pt/reviews` nằm trong portal PT riêng. PT report chỉ gọi ID từ hồ sơ `pts/me`; trang review đọc endpoint phân trang hiện có, chỉ hiển thị trường public và nhãn reviewer “Hội viên”.
- Bộ lọc ngày giới hạn tối đa 366 ngày; lỗi range hiện inline. Màn hình phân biệt loading, lỗi/thử lại, không có dữ liệu và `N/A` cho denominator/rating rỗng. Trạng thái booking được giải thích là trạng thái hiện tại tại `as_of`; utilization được giải thích là chỗ đặt seat-minute, không phải thời gian tập hoặc doanh thu.
- Chart trạng thái dùng `<progress>` có accessible label, đồng thời hiển thị tên trạng thái và giá trị số để không phụ thuộc màu sắc.

## Chưa kiểm chứng

Không chạy browser visual QA ở 375px/desktop, keyboard/focus thực, hoặc luồng live UI → API → PostgreSQL. API và client đang tắt theo yêu cầu dừng trước đó; test UI dùng mock API. Báo cáo API riêng xác nhận service/repository/PostgreSQL; hai lớp kiểm chứng này không thay thế demo end-to-end. Runbook có trình tự demo cho lần chủ động chạy local tiếp theo.
