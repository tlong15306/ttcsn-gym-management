# Kế hoạch và kết quả: dữ liệu demo và kéo chọn giờ

**Cập nhật tiếp nối 31/10/2026:** [30 — Tiến độ và kế hoạch hoàn thiện](30_tien_do_va_ke_hoach_hoan_thien_2026_10_31.md). Nội dung dưới đây giữ snapshot 25/10; các khoảng trống seed/giờ HLV/batch và trạng thái môi trường hiện tại được đối chiếu ở tài liệu 30.

Ngày cập nhật: **25/10/2026, Asia/Ho_Chi_Minh**. **Trạng thái: đã được duyệt; phần mã nguồn D0–D2 đã triển khai, còn kiểm chứng live D0/D3/D4 chờ Docker local.**

Đây là bản sửa của [kế hoạch 28](28_ke_hoach_lam_viec_tiep_2026_10_25.md) theo yêu cầu mới: thêm phòng, huấn luyện viên và các lịch đã đăng ký để người dùng demo/test; sửa thao tác chọn giờ ở **cả lịch đăng ký giờ phòng/HLV và màn Hội viên đặt ca**. Ưu tiên đợt này là có dữ liệu sẵn và giao diện dễ thao tác để người dùng thử luồng trực tiếp. Kết quả thực thi và giới hạn môi trường ở mục 9.

## 1. Các điểm đã đối chiếu trong code

- `seed-room-demo` có thể thêm ba phòng minh họa và ảnh; `seed-booking-demo` có ba phòng FLOW DEMO, một Hội viên và một PT. Seed booking hiện chuẩn bị điều kiện đặt, chưa tạo một bộ booking phong phú theo trạng thái.
- `WeeklyGrid.tsx` đã có pointer drag, đổi đầu/cuối, di chuyển và keyboard. Lịch mở cửa phòng dùng ô một giờ; màn nhỏ mặc định chuyển sang danh sách. Việc tiếp theo là kiểm tra và cải thiện thao tác đang có trên UI thật.
- Lịch làm việc HLV trong `Trainers.tsx` đang thêm từng dòng ngày/giờ bắt đầu/kết thúc.
- Hội viên trong `Booking.tsx` đang bấm chọn từng ca. API tạo booking hiện nhận một `room_session_id`, nên kéo chọn nhiều ca và xác nhận một lần cần bổ sung xử lý backend.
- Các báo cáo kiểm chứng gần nhất là ngày 18/10. Lần PT tải lịch từng gặp 500 và p95 login/read vượt 2 giây vẫn được giữ trong danh sách cần theo dõi.

Lúc lập kế hoạch, lượt này chỉ đối chiếu source. Sau khi được duyệt, mã nguồn và regression được thực hiện theo mục 9; không seed DB dev khi Docker chưa lên được.

## 2. Bộ dữ liệu demo đề xuất

| Nhóm | Quy mô ban đầu | Nội dung để demo |
|---|---|---|
| Phòng | 6 phòng | Hai phòng sức mạnh, hai cardio, hai PT; có ảnh, mô tả, sức chứa; năm phòng hoạt động và một phòng bảo trì |
| HLV | 4 HLV | Ba hồ sơ đã duyệt với chuyên môn/lịch làm khác nhau, một hồ sơ chờ duyệt để thử phân quyền; có nghỉ và giờ bận |
| Hội viên | 12 Hội viên | Mười tài khoản/thẻ hợp lệ, một thẻ hết hạn, một tài khoản bị khóa để thử thông báo lỗi |
| Vai trò vận hành | Admin, Manager, Lễ tân | Tái sử dụng demo accounts còn phù hợp; tạo thêm account DEMO nếu cần |
| Lịch mở và phiên | Từ ngày chạy đến 14 ngày tới | Có ngày trống, giờ trống, giờ đã đăng ký, phiên còn ít chỗ và phiên đầy; HLV rảnh/bận theo ca |
| Booking đã đăng ký | Tối thiểu 24 booking | Mốc ban đầu: 6 chờ PT, 8 xác nhận, 4 hoàn thành, 2 vắng, 2 đã hủy, 1 bị từ chối, 1 hết hạn |
| Attendance/Review | Theo các booking đủ điều kiện | Có trường hợp đủ hai người, thiếu một người, đã review/chưa review; báo cáo có dữ liệu để đối chiếu |

Các số lượng trên là đề xuất cho bộ mới; khi triển khai cần kiểm đếm phần DEMO đang có và tái sử dụng phù hợp. Không cam kết trạng thái đứng yên: booking chờ PT sẽ hết hạn theo job nếu không phản hồi đúng cửa sổ.

### Cách tạo và giữ dữ liệu để người dùng tự thử

1. Bổ sung dữ liệu có nhãn/mã `DEMO` vào DB local phục vụ demo, dự kiến `ttcsn_gym_dev` đang dùng. Dữ liệu này lưu bền sau khi đóng/mở app; test tự động ghi/reset chỉ trên `ttcsn_gym_test`.
2. Chụp backup trước đợt seed đầu; xác minh DB đích bằng kết nối thực. Tái sử dụng seed và service hiện có, giữ kiểm tra quyền, capacity, lịch PT và audit.
3. Seed có mã định danh/manifest để chạy lại không nhân bản. Khi trùng dữ liệu đã sửa thủ công, báo rõ thay vì tự ghi đè mật khẩu, lịch hay booking của người dùng.
4. Tính ngày/giờ theo `Asia/Ho_Chi_Minh` tại lần chạy. Có lệnh bổ sung lịch tương lai và scenario mới khi dữ liệu demo cũ đã qua ngày; giữ những booking người dùng đã thử.
5. Booking tương lai và các scenario đang chờ xử lý được tạo qua nghiệp vụ hiện có. Dữ liệu lịch sử để hiển thị report/review là fixture có nhãn riêng, đảm bảo invariant và audit; ghi rõ đây là dữ liệu minh họa.
6. Tạo manifest gồm ID/phòng/HLV/Hội viên/ngày/ca/trạng thái và bảng account theo vai trò. Chuẩn bị hướng dẫn ngắn: đăng nhập account nào, mở trang nào, booking nào dùng để nhận lịch, điểm danh, hủy hoặc review.

**Đầu ra:** dữ liệu thật đọc được từ UI/API/DB, danh sách tài khoản và scenario sẵn dùng. Không chỉ điền dữ liệu giả ở frontend.

## 3. Lịch đăng ký giờ phòng và giờ HLV

### Thao tác mong muốn

- Hiển thị lịch tuần: hàng là ngày, cột là giờ. Phần giờ và tên ngày giữ rõ khi cuộn trong vùng lịch.
- Nhấn giữ ở giờ bắt đầu, kéo sang giờ kết thúc, thả để có vùng chọn. Ví dụ Thứ Hai `07:00 → 11:00` được chọn trong một lần kéo.
- Có thể kéo qua nhiều hàng ngày để áp dụng cùng dải giờ; có thao tác sao chép ngày sang ngày khác.
- Vùng đang chọn hiển thị ngay giờ bắt đầu/kết thúc; khối đã chọn có tay nắm kéo đầu/cuối và có thể di chuyển.
- Có chế độ thêm/bỏ khung giờ, hoàn tác, xóa lựa chọn và xem lại thay đổi trước khi lưu.
- Kéo chỉ sửa bản nháp. Lịch phòng tiếp tục qua bước xem trước/xác nhận để công bố và sinh phiên; lịch HLV chỉ lưu khi người dùng bấm xác nhận.

### Theo từng màn

| Màn | Phần thay đổi | Quy tắc cần giữ |
|---|---|---|
| Giờ mở cửa phòng | Hoàn thiện kéo quét, chỉnh đầu/cuối và khả năng dùng trên màn nhỏ của grid hiện có | Ô một giờ theo hợp đồng lịch phòng hiện hành; bảo vệ phiên/booking bị ảnh hưởng, preview/revision/idempotency |
| Giờ làm việc HLV | Thay thao tác thêm từng dòng bằng grid kéo chọn; giữ phần nhập giờ chính xác như lựa chọn phụ | Không chồng khung cùng ngày; giữ khoảng nghỉ, lịch đã nhận và dữ liệu có phút lẻ; không làm tròn âm thầm khi mở/lưu |
| HLV tự đăng ký và vận hành chỉnh lịch HLV | Tái sử dụng trải nghiệm grid theo quyền ở hai ngữ cảnh | Giữ chính sách SELF và quyền Admin/Manager theo ma trận hiện có |

Độ chia khi kéo lịch HLV dự kiến 30 phút; dữ liệu có mốc khác phải được hiển thị/giữ chính xác và chỉnh được bằng trường giờ phụ. Chỉ chốt bước chia sau khi đối chiếu validation backend; thay đổi độ chia lịch phòng là phạm vi khác cần ghi riêng.

## 4. Hội viên kéo chọn ca để đặt phòng

### Luồng chính đề xuất

1. Chọn phòng bằng ảnh → chọn ngày.
2. Kéo trên thanh giờ/lịch ngày để chọn **một dải ca liên tiếp trong cùng phòng và ngày**. Ví dụ nếu có ca `07:00–08:00` và `08:00–09:00`, kéo `07:00 → 09:00` chọn cả hai.
3. Hiển thị giờ đã chọn, số ca, số chỗ và các HLV rảnh xuyên suốt tất cả ca; có lựa chọn tự tập.
4. Xem tóm tắt toàn bộ dải giờ rồi bấm **Xác nhận đăng ký** một lần.
5. Đọc lại các booking đã tạo trong Lịch đã đăng ký; hiển thị đúng phòng, giờ, HLV, trạng thái và hành động hợp lệ.

### Các quy tắc cần duyệt cùng giao diện

- Vùng chọn bám theo các ca/phiên hiện có. Chưa thêm loại booking với thời lượng tùy ý hoặc kéo xuyên nhiều ngày/phòng.
- Một ca hết chỗ, đóng cửa, đã qua giờ, trùng lịch hoặc không hợp lệ làm dải đó không thể xác nhận. UI chỉ rõ ca và lý do; không tự bỏ qua ca lỗi để đặt những ca khác.
- Một HLV phải đủ điều kiện và rảnh trên toàn bộ dải đã chọn. Nếu không có HLV phù hợp thì đề nghị đổi giờ hoặc tự tập.
- Đổi ngày/phòng/dải giờ phải làm mới khả dụng và xóa HLV không còn phù hợp. Kết quả UI là tham khảo; server kiểm tra lại lúc xác nhận.
- Mỗi ca vẫn tạo **một booking riêng** theo mô hình hiện hành. PT nhận/từ chối, attendance, hủy và review tiếp tục theo từng booking; tóm tắt phải ghi rõ số booking sắp tạo.

### Phần backend cần bổ sung

Đề xuất API tạo một nhóm booking từ danh sách session ID, dùng chung phòng/ngày/Hội viên và HLV tùy chọn. Xác nhận theo nguyên tắc **tất cả thành công hoặc không tạo booking nào**:

- Kiểm tra đủ quyền, tính liên tiếp, điều kiện Hội viên/HLV và capacity cho mọi ca trong transaction.
- Giữ thứ tự khóa nhất quán với Booking hiện có; kiểm tra race cạnh tranh chỗ cuối, lịch HLV và lịch Hội viên.
- Dùng idempotency cho cả nhóm; double submit/retry không tạo trùng hoặc làm sai số hold/audit.
- Trả danh sách booking hoặc danh sách ca bị conflict để UI cập nhật lại. Không triển khai bằng vòng lặp gọi API từng ca ở frontend rồi chấp nhận kết quả dở dang.
- Giữ API tạo một booking đang có; cập nhật contract, decision log và test cho hành vi nhóm được duyệt.

## 5. Cảm ứng, bàn phím và cách trình bày

- Desktop: kéo bằng chuột, thay đổi vùng chọn thấy ngay; Escape/pointer cancel bỏ thao tác nháp; kéo ngược chiều vẫn cho dải đúng.
- Điện thoại: mặc định xem theo ngày, có chế độ **Chọn giờ** để kéo mà không lẫn với vuốt cuộn. Ngoài chế độ này, cuộn hoạt động bình thường.
- Có cách không cần kéo: chạm điểm bắt đầu rồi chạm điểm kết thúc; bàn phím dùng phím mũi tên/Shift và xác nhận. Các cách này chọn cả dải, không buộc chọn từng ô.
- Khi kéo sát mép vùng lịch có tự cuộn trong vùng; thả ra ngoài không tự đặt lịch. Không gọi API mỗi lần con trỏ đi qua một ô.
- Các trạng thái có màu kèm nhãn: còn chỗ, đang chọn, đã đăng ký, hết chỗ/không khả dụng. Phòng tập chung có người đặt vẫn có thể còn chỗ; chỉ phần chỗ đầy hoặc booking trùng của Hội viên mới làm không thể chọn theo quy tắc tương ứng.
- Giữ kiểu chữ và màu hiện có: vận hành Segoe UI, portal Be Vietnam Pro/Barlow Condensed; Hội viên xanh lá, PT xanh dương. Dùng nền `#f5f6f8`, mặt `#ffffff`, chữ `#0f172a`, viền `#e2e8f0`, nhấn Hội viên `#18684d`, nhấn PT `#275eb0` theo ngữ cảnh; tái sử dụng token hiện có.

Phác bố cục màn Hội viên:

```text
Phòng DEMO · Strength       [Ngày tập]

Giờ       07      08      09      10      11
Ca        [  đang chọn  ] [đầy] [còn chỗ]
          Kéo hoặc chạm điểm đầu / điểm cuối

Đã chọn: 07:00–09:00 · 2 ca
HLV: [Tự tập] [HLV rảnh toàn bộ khung giờ]

[Xóa lựa chọn]             [Kiểm tra đăng ký]
```

Đây là phác tương tác để duyệt. Chi tiết kích thước và ảnh app thật được kiểm tra lúc triển khai. Hai skill thiết kế được dùng để giữ phản hồi tức thì, phân biệt cuộn/chọn và có lựa chọn bàn phím/chạm thay cho kéo.

## 6. Thứ tự triển khai sau khi duyệt

| Chặng | Công việc | Điều kiện bàn giao |
|---|---|---|
| D0 — Môi trường và dữ liệu demo | Kiểm tra local, backup, mở rộng seed, lịch theo ngày chạy, manifest/account/scenario | Đăng nhập thấy phòng/HLV/booking sẵn; dữ liệu còn sau restart; chạy seed lại không nhân bản |
| D1 — Grid phòng/HLV | Hoàn thiện grid phòng; thêm grid HLV, cảm ứng, bàn phím và preview | Kéo chọn 07–11 và nhiều ngày được; lưu/đọc lại đúng; lịch phút lẻ của HLV được giữ |
| D2 — Đăng ký nhiều ca | Chốt contract nhóm, transaction/idempotency, giao diện kéo dải và xác nhận | Chọn hai ca bằng một lần kéo; server tạo đủ hoặc rollback toàn bộ; không trùng khi retry |
| D3 — Test luồng và sửa UI | Tự tập/có PT, chờ/nhận/từ chối/hủy, attendance/completion/review/report; mobile/desktop | Có bằng chứng UI → API → DB; lỗi tìm được sửa và test lại; kiểm tra booking mới có thể thao tác trong demo |
| D4 — Bàn giao để người dùng thử | Hướng dẫn ngắn, account, danh sách scenario, lệnh refresh demo và report | Người dùng mở app, đăng nhập và tự thử được các luồng mà không chuẩn bị lại dữ liệu |

Trong mỗi chặng: triển khai → kiểm tra → sửa → lưu bằng chứng. Đợt đầu này ưu tiên D0–D4. API/security/load/job/restore mở rộng và hồ sơ slide trong kế hoạch 28 tiếp tục sau đó; lỗi dữ liệu/phân quyền cản demo phải xử lý ngay.

## 7. Kiểm chứng bắt buộc

- Seed: chạy hai lần không tăng bản ghi ngoài dự kiến; khởi động lại vẫn thấy dữ liệu; ngày/ca còn hợp lệ; dữ liệu ngoài DEMO và mật khẩu cũ được giữ.
- Grid: kéo xuôi/ngược, nhiều hàng, kéo đầu/cuối, Escape/cancel, cuộn, cảm ứng, nhập phụ và bàn phím; 375/768/1024/1440px; lưu xong tải lại đúng.
- Batch booking: thành công nhiều ca; một ca conflict rollback toàn bộ; hai người tranh chỗ cuối; PT/Hội viên bận ở một ca; retry cùng key; kiểm tra hold, audit và rollback.
- Luồng demo: Hội viên đặt → PT nhận → Lễ tân điểm danh → job hoàn thành → review → report; scenario lịch sử có nhãn fixture, luồng live có thời gian thực và bằng chứng riêng.
- Regression phù hợp phạm vi: Go/DB tests, vet, frontend tests/build, contract/docs checker; chạy tuần tự các suite có reset DB test. Không dùng DB demo lưu bền làm DB reset.

## 8. Các điểm người dùng duyệt

1. Quy mô dữ liệu: 6 phòng, 4 HLV, 12 Hội viên, tối thiểu 24 booking nhiều trạng thái, lưu bền cho demo.
2. Áp dụng kéo chọn cho cả lịch giờ phòng/HLV và Hội viên đặt ca; có cảm ứng và cách chạm hai đầu/bàn phím.
3. Hội viên chọn nhiều ca liên tiếp rồi xác nhận một lần; mỗi ca vẫn là một booking; tạo nhóm theo transaction tất cả hoặc không có gì.
4. Thứ tự D0 dữ liệu → D1 lịch phòng/HLV → D2 đặt nhiều ca → D3 test/sửa → D4 bàn giao để tự thử.

## 9. Kết quả thực thi sau khi được duyệt

| Chặng | Kết quả | Trạng thái kiểm chứng |
|---|---|---|
| D0 — seed DEMO | Thêm `server/cmd/seed-demo-scenarios`: chỉ nhận PostgreSQL local `ttcsn_gym_dev`, từ chối DB test/remote, chạy lặp không nhân bản/ghi đè; có 12 Hội viên, 4 HLV, 6 phòng, cover/lịch và 24 booking theo trạng thái kế hoạch. | Static PASS; **chưa chạy vào DB dev** vì Docker Desktop không khởi engine. |
| D1 — lịch phòng/HLV | Lịch phòng có chế độ chạm hai đầu bên cạnh kéo/bàn phím. HLV có grid tuần 30 phút kéo/chạm/bàn phím; ngày có mốc phút lẻ bị khóa trên grid và giữ qua form chi tiết. | Component tests và production build PASS. |
| D2 — dải ca Hội viên | Thêm `POST /bookings/batch` và endpoint đặt hộ; transaction all-or-none, idempotency riêng, kiểm tra liên tiếp/capacity/trùng lịch/PT. Màn Hội viên kéo hoặc chọn hai đầu dải ca; PT là giao của các ca; một ca vẫn dùng API cũ. | Go/unit + frontend test/build PASS; PostgreSQL integration test compile nhưng skip khi `TEST_DATABASE_URL` chưa có. |
| D3/D4 — demo live | Có lệnh seed và account demo để tiếp tục. | **Chưa xác minh** UI → API → DB, restart dữ liệu, browser responsive hay flow PT/attendance/review/report vì engine Docker bị chặn trước khi DB chạy. |

Đã chạy sau khi ghép thay đổi: `go test ./...`, `go vet ./...`, `npm.cmd test` (8 file, 52 test), `npm.cmd run build`, `python scripts/generate_api_contracts.py --check` (78 operations) và `docs/check_docs.ps1`. Các lệnh không yêu cầu DB đều PASS.

Docker Desktop 4.93.0 dừng trước khi mở engine do socket tạm `AppData/Local/docker-secrets-engine/engine.sock` bị hệ thống khóa. Không xóa database, image hay volume, và không factory-reset Docker. Khi Docker đã hoạt động lại, tiếp tục từ thư mục `server` theo thứ tự:

```powershell
go run .\cmd\migrate up
go run .\cmd\seed
go run .\cmd\seed-demo-scenarios
```

Sau đó mới kiểm chứng hai lượt seed, DB `ttcsn_gym_dev`, browser và integration chỉ dùng `ttcsn_gym_test`/port `18080`.
