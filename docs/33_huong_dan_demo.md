# Hướng dẫn demo GYM

Dữ liệu demo dùng PostgreSQL `ttcsn_gym_dev`. Xem [hướng dẫn chạy local](17_huong_dan_chay_local.md) để cấu hình và khởi động API/client, sau đó mở `http://127.0.0.1:5173`.

## Tài khoản và dữ liệu

### Xem và chọn thử trước khi đăng nhập (U1–U4)

Mở `/` khi chưa đăng nhập: xem trang chủ, `/explore/rooms` tìm phòng, `/explore/trainers` xem HLV, mở phòng để xem/chọn thử lịch tuần. Trên điện thoại, dùng **Xem lịch** để vuốt ngang đủ 7 ngày, **Chọn giờ** để kéo hoặc chạm ca đầu/ca cuối. Kéo thân khối giữ thời lượng, hai tay nắm đổi biên; có ô Bắt đầu/Kết thúc và phím mũi tên trên tay nắm/thân khối. Escape bỏ thao tác đang kéo; Xóa lựa chọn bỏ dải đã chọn. Bộ lọc sáng/chiều/tối và Tìm ngày còn ca gần nhất giúp tìm giờ phù hợp. Chỉ chọn một dải ca thật liền nhau trong cùng ngày, không tự chia ca nhỏ hơn hoặc kéo qua ca bị chặn.

Chọn Tự tập/Cần huấn luyện viên rồi tiếp tục đăng nhập Hội viên. Bản nháp tối đa **30 phút** chỉ lưu ID/giờ ca và ý định, không giữ chỗ. Khi trở lại, hệ thống kiểm tra lại điều kiện và lịch; bản nháp thuộc tài khoản khác hoặc đã hết chỗ không được dùng. HLV cụ thể phải phù hợp toàn dải; khi cần chọn lại hình thức tập, bấm Tự tập hoặc chọn HLV rõ ràng. Đổi ngày/tuần sẽ bỏ dải cũ kèm thông báo. Xem lại ghi ngày/giờ VN, số ca, thời lượng và điều kiện hủy/HLV; chỉ xác nhận mới tạo lịch. Thành công hiển thị các mã lịch; lỗi mất phản hồi có thể thử lại cùng yêu cầu, không đổi giờ/HLV nếu đang cần nhận lại kết quả cũ. Đăng xuất chủ động xóa bản nháp.

Sau lần hoàn thiện U4, dự án được dừng và máy tắt theo yêu cầu người dùng. Lần sau chạy lại theo hướng dẫn local; không xóa volume DB. Bằng chứng và vị trí backup: [báo cáo U2–U4](reports/public_booking_u2_u4_2026-11-01.md).

| Vai trò / tình huống | Tài khoản | Cách dùng |
|---|---|---|
| Hội viên chính | `demo.member.01` | Có hồ sơ hội viên và hạn thẻ, dùng để chọn ca và đặt lịch. |
| HLV | `demo.pt.strength` | Dùng xem và phản hồi yêu cầu PT khi HLV này được chọn. |
| Lễ tân | `reception.demo` | Xem lịch và điểm danh khi buổi đủ điều kiện. |
| Quản lý | `manager.demo` | Xem lịch, báo cáo và dữ liệu vận hành. |
| Quản trị viên | `admin.demo` | Quản lý dữ liệu theo quyền. |
| Hội viên hết hạn / bị khóa | `demo.member.expired`, `demo.member.locked` | Minh họa các trường hợp không đủ điều kiện. |

Mật khẩu mặc định **cho tài khoản demo mới được seed** là `DemoPass123!`. Seed giữ nguyên mật khẩu tài khoản đã có; nếu từng thay đổi thì dùng mật khẩu hiện tại. `member.demo` là tài khoản đăng nhập được nhưng trong dữ liệu hiện tại chưa có hồ sơ Hội viên, do đó không dùng làm tài khoản đặt lịch chính.

Phòng có tiền tố `DEMO ·` là dữ liệu minh họa. Có lịch đã xác nhận, hoàn thành, hủy, hết hạn, vắng mặt và từ chối; các ca đầy/còn một chỗ để kiểm tra điều kiện. Trạng thái và số chỗ thực tế có thể thay đổi sau khi thao tác.

## Luồng thử nhanh

1. Đăng nhập `demo.member.01`, chọn phòng **DEMO · Strength Arena** và ngày có ca trong tương lai. Kéo qua các ca liền kề; trên điện thoại hoặc bằng bàn phím chọn ca đầu và ca cuối. Kiểm tra ngày/giờ, số ca và số chỗ trong phần tóm tắt.
2. Chọn tự tập để xác nhận ngay hoặc chọn HLV khả dụng cho toàn bộ các ca để gửi yêu cầu. Kiểm tra lại lịch của Hội viên sau khi xác nhận.
3. Với lịch có PT, đăng nhập tài khoản PT tương ứng để xem và phản hồi. Yêu cầu chỉ chờ tối đa 30 phút và phải được phản hồi trước giờ ca bắt đầu.
4. Lễ tân kiểm tra lịch và điểm danh khi ca đủ điều kiện; sau khi ca kết thúc, completion xử lý trạng thái. Hội viên có thể đánh giá buổi đủ điều kiện; quản lý xem báo cáo tương ứng. Các bước chờ giờ thực phụ thuộc thời gian ca, không được thay bằng fixture lịch sử khi trình bày như luồng trực tiếp.
5. Thử riêng các trường hợp ca đầy, trùng lịch, tài khoản hết hạn và HLV không khả dụng. Dữ liệu demo có thể đã thay đổi do thao tác trước đó, nên kiểm tra trạng thái thực tế trước khi trình bày.

## Làm mới yêu cầu PT

Sau khi đã sao lưu dữ liệu dev, chạy từ thư mục `server`:

```powershell
go run ./cmd/seed-demo-scenarios --refresh-pending
```

Lệnh chỉ chấp nhận database local `ttcsn_gym_dev`, giữ các booking đã có và tạo yêu cầu demo mới khi yêu cầu mẫu tương ứng đã hết hạn, từ chối hoặc hủy. Lệnh tạo manifest trong `.artifact_work/demo-manifests`; xem manifest mới nhất để biết ID và thời điểm của từng kịch bản. Yêu cầu mới vẫn hết hạn theo quy tắc 30 phút, nên làm mới gần lúc demo.

Để dừng các tiến trình local đã được script quản lý, chạy `powershell -ExecutionPolicy Bypass -File .\scripts\stop-local.ps1` từ thư mục gốc. Xem hướng dẫn local về `-KeepDatabase` và sao lưu.
