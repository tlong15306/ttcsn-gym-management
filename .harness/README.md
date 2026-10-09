# Harness GYM — bộ thiết kế v1.0

Trạng thái ngày 10/10/2026: **có thiết kế, schema và task khởi đầu; chưa có runner hoặc agent run**.

Đọc [thiết kế đầy đủ](../docs/17_thiet_ke_harness.md) trước khi triển khai. Baseline nghiệp vụ vẫn là [GYM v0.3](../docs/README.md); tài liệu harness không tự thay đổi yêu cầu GYM.

## Artifact đang có

| File | Vai trò |
|---|---|
| `schemas/task.schema.json` | Task spec được version/hash trước khi claim; không chứa runtime status. |
| `schemas/worker-result.schema.json` | Báo cáo từ worker; mọi checks đều là lời khai cần xác minh. |
| `schemas/gate-result.schema.json` | Evidence chỉ verifier được ghi qua API/runtime riêng. |
| `schemas/review.schema.json` | Review gắn đúng candidate SHA và contract hash. |
| `tasks/HAR-001.json` | Task triển khai doctor/config đầu tiên; chưa được chạy/nhận. |

HAR-002…HAR-007 và backlog GYM nằm trong tài liệu thiết kế, chưa phải task spec đủ điều kiện chạy. Prompts, coverage manifest, gate registry, TypeScript runner, SQLite store, sandbox và CI chưa được tạo.

## Quy tắc host cần kiểm tra thêm ngoài JSON Schema

Schema chỉ kiểm tra hình dạng dữ liệu. Runner phải thực thi các invariant sau, bằng code và quyền truy cập runtime:

1. Task ID/revision và acceptance ID không trùng; mọi dependency tồn tại; không tự phụ thuộc hoặc cycle. Gate của mỗi acceptance phải thuộc `gate_ids`, và registry đã có recipe với version/hash. Không có recipe thì `NOT_CONFIGURED`.
2. `allowed_paths` là đường dẫn tương đối trong worktree, hoặc prefix thư mục kết thúc bằng `/`; không hỗ trợ glob ở v1. Reject absolute path, `..`, `.git`, symlink/junction thoát worktree, đường dẫn secret và frozen contract/gate. Kiểm tra lại resolved path sau mỗi thay đổi; prefix không phải sandbox.
3. Worker `candidate_ready` phải trỏ tới commit thật có base/dependency SHAs đúng, không blocker và không file ngoài scope. Host tính changed files từ Git; không tin danh sách agent. Candidate phải được seal trước verifier/reviewer.
4. Gate PASS cần exit 0, không timeout, parse report thành công, required named tests thực sự chạy và artifact hash hợp lệ. Thời điểm, run/attempt, snapshot SHA, contract/recipe hash phải khớp. Required SKIP/NOT_CONFIGURED/ERROR đều chặn nghiệm thu.
5. Review APPROVE chỉ hợp lệ khi tất cả acceptance được đánh giá `satisfied`, có evidence ID tồn tại do verifier sở hữu, không Critical/High mở và cùng candidate/contract hash. Review không phải lệnh merge; integrator còn phải retest kết quả tích hợp.
6. Agent không được tự phát hành gate result hoặc đổi task thành ACCEPTED. `checks_claimed` chỉ hỗ trợ debug. Chữ PASS hoặc review JSON đúng schema không thay thế kiểm thử.
7. Giới hạn task bổ sung cho run budget. Run manifest cần wall-time ceiling và usage/token ceiling tường minh. Telemetry usage thiếu là unknown; policy v1 dừng ở WAITING_INPUT trước agent run khi không thể áp dụng ceiling. Timeout/output cap áp dụng cho subprocess và process tree.

Schema theo JSON Schema draft 2020-12, dùng cho host validation. Nếu provider chỉ hỗ trợ một subset, adapter tạo schema output tương thích đã kiểm thử và vẫn validate bằng schema host gốc; không truyền nguyên file rồi giả định provider hỗ trợ tất cả keyword. Gate-result schema không gửi cho worker làm output format.

## Bắt đầu triển khai

1. Làm `HAR-001` trong `tools/harness/` và template config theo allowed paths của task. Tests doctor dùng fixture offline. Doctor runtime không cài CLI/dependency hoặc gọi model.
2. Chủ trì bootstrap chốt recipe cho HARNESS-UNIT, DOCTOR-FIXTURE và SECRET-SCOPE, lưu hash ngoài snapshot worker trước khi nghiệm thu. Lần đầu chạy kiểm tra độc lập bằng terminal; chưa có runner để tự nghiệm thu bản thân.
3. Đạt doctor rồi triển khai state/DAG, adapters, verifier/worktree, reviewer/integrator, pilot health và crash/retry tests theo HAR-002…007.
4. Chỉ sau pilot và fault tests đạt mới mở task GYM. DB test phải cô lập; không reset DB local `master-golang`.

Các lệnh `harness doctor/run/resume` trong thiết kế là giao diện dự kiến, hiện chưa chạy được. Không có cron, process nền hoặc agent nào được khởi chạy bởi bộ tài liệu này.

## Dữ liệu và Git

Commit spec/schema, source, fixture giả và report đã redaction. Runtime state/log/worktree đặt ngoài OneDrive/repository, ví dụ `%LOCALAPPDATA%/TTCSN-Harness/<project-id>/`. Không commit `.env`, key, cookie, transcript thô, SQLite/WAL hay volume DB.

Nhật ký phải phân biệt agent thực hiện, sinh viên review và phần chưa kiểm chứng. Mỗi task accepted sau này có walkthrough tiếng Việt và lệnh chạy lại để sinh viên đọc hiểu.
