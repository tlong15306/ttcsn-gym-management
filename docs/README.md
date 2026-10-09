# Bộ tài liệu Hệ thống Quản lý GYM

## Baseline đang có hiệu lực

Baseline phân tích và thiết kế hiện hành là **v0.3**, lập ngày **10/10/2026**. Bộ tài liệu v0.3 dùng để triển khai và kiểm thử; các file v0.1/v0.2 được giữ làm lịch sử và không còn là nguồn quyết định hiện hành.

Nhật ký học và tự phát triển của sinh viên nằm tại `../NHAT_KY_DU_AN.md`; cập nhật file này sau mỗi buổi code.

| Thứ tự | Tài liệu hiện hành | Mục đích |
|---:|---|---|
| 1 | `00_quyet_dinh_va_quy_tac_nghiep_vu.md` | Phạm vi, quyết định D01–D10 và quy tắc BR. |
| 2 | `03_tac_nhan.md` | Mục tiêu, trách nhiệm và phạm vi dữ liệu của 5 tác nhân. |
| 3 | `04_use_case_tong_quat.md` | Danh mục UC-01…UC-16 ổn định. |
| 4 | `06_ma_tran_phan_quyen.md` | Nguồn chính sách quyền duy nhất. |
| 5 | `07_SRS_v0.3.md` | Baseline yêu cầu chức năng và phi chức năng. |
| 6 | `08_use_case_chi_tiet.md` | Luồng chính, ngoại lệ và hậu điều kiện của use case. |
| 7 | `09_ERD_va_CSDL.md` | Mô hình dữ liệu logic/vật lý và quy tắc toàn vẹn. |
| 8 | `09_schema.sql` | Schema PostgreSQL 16+ cho cài đặt mới; không phải migration nâng cấp. |
| 9 | `11_kien_truc_he_thong.md` | Kiến trúc module, transaction, bảo mật và vận hành. |
| 10 | `12_dac_ta_API.md` | Hợp đồng REST API, lỗi, quyền và idempotency. |
| 11 | `13_backlog_v0.3.md` | Ưu tiên, phụ thuộc, trạng thái phân tích/triển khai. |
| 12 | `15_ke_hoach_kiem_thu.md` | Kiểm thử yêu cầu, dữ liệu, đồng thời, bảo mật và NFR. |
| 13 | `16_runbook_van_hanh_nang_cap.md` | Migration, backup/restore, job nền và xử lý sự cố. |

`14_ke_hoach_hoan_thien_tai_lieu.md` là kế hoạch và biên bản rà soát đầu vào. `01_khao_sat_chuc_nang.md`, `05_khung_tien_do_10_tuan.md` và `10_so_sanh_cong_nghe.md` là tài liệu tham khảo; khi có xung đột, baseline v0.3 ở bảng trên được ưu tiên.

## Quy tắc quản trị tài liệu

- Mỗi yêu cầu dùng mã `REQ-*` hoặc `NFR-*`; mỗi use case dùng mã `UC-*`; mỗi quy tắc nghiệp vụ dùng mã `BR-*`.
- Thay đổi phạm vi hoặc quy tắc phải được ghi trong decision log trước khi sửa SRS, ERD và API.
- Quyền được quản lý tại `06_ma_tran_phan_quyen.md`; các tài liệu khác dẫn chiếu, không tạo ma trận quyền cạnh tranh.
- DDL cài đặt mới, migration nâng cấp và seed dữ liệu là ba artifact khác nhau. Không chạy `09_schema.sql` trên CSDL có dữ liệu.
- Trạng thái “Đã triển khai” hoặc “Đã kiểm thử” chỉ được ghi khi có mã nguồn/kết quả kiểm thử tương ứng. Baseline hiện tại mới xác nhận **đã phân tích và thiết kế**.

## Trạng thái xác minh

Bộ tài liệu đã được đối chiếu tĩnh về mã REQ/UC, quyền, trạng thái và trường dữ liệu. Backend hiện đã có config, Gin health endpoints và pool PostgreSQL; sinh viên đã chạy API kết nối DB thành công. Schema/migration và nghiệp vụ GYM, frontend, hiệu năng, tải, backup/restore chưa được xác minh runtime.

Chạy kiểm tra nhất quán sau mỗi lần sửa:

```powershell
powershell -ExecutionPolicy Bypass -File .\docs\check_docs.ps1
```
