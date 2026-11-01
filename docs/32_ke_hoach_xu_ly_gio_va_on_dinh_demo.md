# Kế hoạch xử lý giờ và ổn định demo — 31/10/2026

Trạng thái: **đã thực hiện một phần; cập nhật kế hoạch còn lại để duyệt ngày 31/10/2026**. Kết quả mới nằm trong [báo cáo kiểm chứng giờ và demo](reports/time_demo_verification_2026-10-31.md); giữ kế hoạch H0–H5 ở [tài liệu 30](30_tien_do_va_ke_hoach_hoan_thien_2026_10_31.md) cho phần hoàn thiện kỹ thuật toàn dự án.

## Tiến độ mới nhất và kế hoạch làm tiếp

Đã đồng bộ hàm ngày/giờ Việt Nam trên các màn chính; frontend **59/59 PASS**, backend PostgreSQL regression và `go vet` PASS. Phép thử HTTP loopback với 100 danh tính: 80 booking thành công, 20 lỗi hết chỗ đúng dự kiến; không vượt capacity 80. p95 login **2,84 giây**, đọc **0,35 giây**, booking **1,22 giây**. Burst test ngắn này chưa đủ để nghiệm thu NFR; profiling lần riêng cho thấy bcrypt chiếm phần lớn CPU. Bộ điều khiển trình duyệt dừng vì không xác định chắc URL, nên chưa hoàn tất browser E2E.

| Đợt tiếp theo | Việc cụ thể | Kết quả cần bàn giao |
|---|---|---|
| A · Chốt demo trên giao diện (P0) | Chuẩn bị API/client test riêng, proxy và CORS đúng; dùng DB test cho các thao tác thử. Kiểm tra Hội viên kéo 2–3 ca → xác nhận một lần → PT phản hồi → lễ tân điểm danh → completion → review → báo cáo. Kiểm tra lưu/mở lại lịch phòng, lịch HLV, mốc 24:00 và giờ giữa các vai trò. Bổ sung tình huống đầy chỗ, trùng lịch, hủy/từ chối/hết hạn, mất phiên/mạng, bàn phím và cảm ứng. Xác minh URL của công cụ trước thao tác; nếu công cụ vẫn chặn, ghi rõ phần chưa xác minh và tiếp tục phần API/DB độc lập. | Checklist theo vai trò, ID booking/request và bằng chứng browser/API/DB khớp. Sửa các lỗi chức năng tìm được. Phân biệt test điều chỉnh thời gian trong DB với completion chờ ca thật. |
| B · Ổn định môi trường và job (P1) | Chẩn đoán cập nhật Vite; đối chiếu trình duyệt sạch với tab có tiện ích đổi múi giờ. Kiểm tra job sau gián đoạn kết nối trên môi trường cô lập; bổ sung chẩn đoán lỗi đã che dữ liệu nhạy cảm và đo thời gian xử lý tồn đọng. | Có cách chạy demo nhận đúng bản mã; job tiếp tục xử lý sau gián đoạn và không tạo audit trùng. Không đổi tiện ích người dùng khi chưa được phép. |
| C · Đóng hiệu năng và bảo mật (P1) | Đo tải HTTP có ramp và giữ tải, đủ 100 danh tính với login/read/booking mutation; thu p50/p95/p99, throughput, lỗi, CPU/RAM và pool/lock-wait. Phân tích login theo số đo bcrypt và tài nguyên, thử cấu hình hoặc tối ưu có căn cứ rồi so lại. Rà quyền theo đối tượng, upload/file private, log và dependency. | Báo riêng tính đúng và hiệu năng; p95 tương tác theo mục tiêu <2 giây hoặc ghi rõ chưa đạt và nguyên nhân đo được. Giữ yêu cầu bảo mật mật khẩu, không hạ bcrypt chỉ để vượt benchmark. |
| D · Kiểm tra cuối và bàn giao (P1) | Chạy các kiểm tra chịu ảnh hưởng, full regression và các load test opt-in; cập nhật tài khoản/scenario/hướng dẫn chạy/tắt/refresh. Kiểm tra restart giữ dữ liệu, backup/restore DB và file private. Dọn artifact do lượt kiểm thử tạo, gồm binary profiling, và phân nhóm thay đổi Git để review. | Hướng dẫn demo dùng được, báo cáo khớp revision và kết quả; mọi mục chưa đóng được ghi rõ. Giữ thay đổi và dữ liệu có trước; không tự xóa file chưa rõ nguồn hoặc ghi nhận commit khi chưa tạo. |

Ưu tiên thực hiện **A → B → C → D**. Có thể đóng mốc demo chức năng sau A; chỉ kết luận hoàn thiện kỹ thuật khi các cổng B–D có bằng chứng. Khối lượng còn lại dự kiến **3–5 buổi tập trung**, phụ thuộc lỗi tìm được và khả năng kiểm tra browser; không phải cam kết ngày.

Các mục dưới đây giữ lại nội dung kế hoạch gốc để đối chiếu phạm vi đã duyệt.

## Cơ sở lập kế hoạch

- Màn đặt lịch đã hiển thị đúng 06:00–09:00 sau khi chuyển phép tính ngày/giờ Việt Nam sang UTC+7 trực tiếp. Đã xác minh trên browser; 11 test liên quan và build PASS.
- Trong tab bị lỗi, phép định dạng yêu cầu `Asia/Ho_Chi_Minh` lại trả múi giờ `UTC`. Chưa xác định thành phần nào can thiệp; chưa được kết luận là extension hay cấu hình Chrome cụ thể.
- Nhiều màn còn định dạng ngày/giờ riêng: lịch Hội viên/HLV, điểm danh, đánh giá, quản lý phòng/lịch phòng, HLV, báo cáo và thông tin Hội viên/tài khoản. Có chỗ dùng ngày của máy người dùng để điền mặc định hoặc lọc.
- Tài khoản `member.demo` có quyền MEMBER nhưng chưa có hồ sơ Hội viên. `demo.member.01` có hồ sơ và hạn thẻ phù hợp. Đăng nhập thành công chưa chứng minh đủ điều kiện đặt lịch.
- Regression frontend gần nhất chưa PASS toàn bộ: 53/55, hai timeout; chạy lại riêng có timeout đổi vị trí. Chưa xác định nguyên nhân.

## Thứ tự thực hiện

| Bước | Công việc | Điều kiện hoàn thành |
|---|---|---|
| 1 · Chuẩn hóa thời gian (P0) | Chuyển tiện ích ngày/giờ Việt Nam thành module dùng chung; tách rõ ngày lịch `YYYY-MM-DD`, giờ trong ngày và timestamp có múi giờ. Dùng nhất quán ở lịch Hội viên/HLV, điểm danh, đánh giá, phòng/lịch phòng, HLV và báo cáo; rà ngày mặc định và bộ lọc. Tránh chuyển múi giờ hai lần. | Cùng một ca hiện cùng ngày/giờ ở mọi màn và mọi vai trò; ranh giới 00:00, qua ngày/tháng/năm và mốc kết thúc 24:00 đúng. Timestamp gửi API, lưu DB và hạn phản hồi vẫn theo hợp đồng nghiệp vụ. |
| 2 · Xác định can thiệp trong tab (P1) | So sánh tab đang lỗi với hồ sơ trình duyệt thử nghiệm sạch; kiểm tra việc ép múi giờ và kết nối cập nhật Vite. Chỉ kết luận theo phép đo. Tách chẩn đoán khỏi giao diện sản phẩm và dọn mã tạm sau kiểm tra. | Xác định được thành phần gây lỗi hoặc ghi rõ phạm vi chưa biết; có cách chạy demo nhận đúng bản mã mới. Mọi thay đổi extension/cài đặt người dùng chỉ thực hiện khi được phép. |
| 3 · Chốt demo và luồng chọn giờ (P0) | Chọn `demo.member.01` làm tài khoản hướng dẫn chính; ghi rõ tài khoản thiếu hồ sơ, khóa, hết thẻ là scenario riêng. Kiểm tra kéo 2–3 ca, hai lần chạm, bàn phím, Escape, cuộn và thả ngoài cho Hội viên; kéo/lưu/đọc lại lịch phòng và lịch HLV. Kiểm tra PT khả dụng cho toàn dải, ca đầy/còn một chỗ, PT bận/nghỉ, Hội viên trùng lịch. | Tài khoản demo chính thao tác được; chọn giờ và tóm tắt khớp; lưu rồi mở lại không đổi giờ. Lỗi lựa chọn có phản hồi rõ ràng. |
| 4 · Kiểm chứng liên thông và regression (P0) | Kiểm tra chọn nhiều ca → xác nhận một lần → PT phản hồi → điểm danh → completion → đánh giá → báo cáo. Kiểm tra batch đủ hoặc không tạo ca nào, replay, quyền và hết phiên. Chạy test DB tuần tự; chẩn đoán timeout frontend bằng đo thời gian và phạm vi render/tương tác rồi sửa nguyên nhân. | Có ID booking/request và đối chiếu API/DB/browser; các lỗi bận, đầy, hủy, từ chối, hết hạn không làm sai dữ liệu. Frontend regression toàn bộ PASS, không chỉ từng file; build và các kiểm tra backend cần thiết PASS. |
| 5 · Chốt hướng dẫn và phần còn mở (P1) | Cập nhật báo cáo, tài khoản/scenario, cách chạy/tắt và refresh yêu cầu PT demo. Ghi rõ browser nào đã thử, kết quả test, giới hạn còn lại. Tiếp tục H4 theo kế hoạch 30: tải, hồi phục job, bảo mật và restore file private. | Có bản demo lặp lại được và danh sách công việc còn mở có bằng chứng; không gọi đạt NFR hoặc hoàn thiện toàn dự án khi chưa kiểm chứng. |

## Nguyên tắc triển khai

- Dữ liệu demo lưu bền ở `ttcsn_gym_dev`; kiểm thử có reset/mutation dùng `ttcsn_gym_test`, không chạy các suite reset song song. Backup trước thay đổi dữ liệu dev.
- Không tự sửa hồ sơ/quyền/mật khẩu tài khoản cũ để làm test qua; không ghi đè booking đã được thao tác. Giữ các thay đổi Git có trước.
- Dùng UTC+7 cho lịch vận hành hiện tại tại Việt Nam; ngày lịch thuần túy không được coi nhầm là timestamp UTC. Backend quyết định deadline và điều kiện nghiệp vụ theo giờ server.
- Bổ sung test cho sai ngày/giờ và hành vi thực tế; không tăng timeout để che lỗi chưa chẩn đoán. Phân biệt E2E có điều chỉnh thời gian trong DB test với completion chờ thời gian thực.
- Không đưa lưu token lâu dài, thay đổi đăng nhập hoặc đổi thiết kế toàn ứng dụng vào phạm vi sửa giờ nếu chưa có nhu cầu riêng.

## Mốc đề xuất

Ưu tiên đóng **bước 1 → 3 → 4** để có demo ổn định. Bước 2 thực hiện xen kẽ để xử lý môi trường phát triển và ghi nhận nguyên nhân; bước 5 chốt tài liệu rồi tiếp tục các cổng vận hành H4/H5. Mỗi bước báo rõ đã kiểm tra gì và còn thiếu gì.
