# Đề xuất UI/UX: trang chủ công khai và chọn lịch Hội viên

Ngày lập: 31/10/2026. Trạng thái: **đề xuất để người dùng duyệt**. Ưu tiên đợt này trước các phần vận hành còn lại của [kế hoạch 32](32_ke_hoach_xu_ly_gio_va_on_dinh_demo.md). Chưa triển khai ứng dụng theo tài liệu này.

**Mockup U0 đã có:** [mở bản mẫu tương tác](mockups/u0/index.html) và [hướng dẫn duyệt](mockups/u0/README.md). Đây là bản mẫu độc lập, có dữ liệu minh họa và chưa thay đổi luồng của ứng dụng thật.

## 1. Yêu cầu đã làm rõ

Xây dựng trải nghiệm GYM theo luồng **khám phá → chọn lịch → đăng nhập khi cần → xác nhận**. Người chưa đăng nhập vào được trang chủ, tìm hiểu phòng và HLV, xem lịch trống và thử chọn khoảng giờ. Hội viên đặt lịch bằng lịch trực quan có thao tác kéo chọn và chỉnh khoảng giờ, thay cho các ô ca rời rạc hiện nay.

Trong đề xuất này, “đăng ký” là đăng ký lịch tập. Việc cấp tài khoản/hồ sơ Hội viên tiếp tục theo luồng hiện có. Trang chủ có nút đăng nhập và hướng dẫn liên hệ lễ tân khi chưa có tài khoản; không trình bày nút tự đăng ký tài khoản khi hệ thống chưa có chức năng đó.

### Cơ sở từ mã hiện tại

- Route `/` đang nằm trong Guard và chuyển người chưa đăng nhập sang `/login`.
- API phòng, ảnh phòng, HLV và lịch hiện yêu cầu xác thực. `RoomImage`, `RoomCatalog` và các hook dữ liệu cũng phụ thuộc session; cần tách dữ liệu hiển thị công khai.
- Hội viên đã có kéo qua các nút ca nhưng chưa có lịch trục giờ và khối chọn có tay nắm chỉnh hai đầu. Grid admin hỗ trợ mở/đóng, di chuyển và thay đổi biên lịch; Hội viên cần cùng nguyên tắc thao tác với mục đích chọn buổi tập.
- Batch booking hiện yêu cầu các ca cùng phòng, cùng ngày, liền kề và tối đa 24 ca. Giao diện phải bám ranh giới session thực tế, không cho chọn phút tùy ý ngoài contract.
- Các hàm ngày/giờ Việt Nam đã được chuẩn hóa và phải được dùng tiếp trên các màn mới.

## 2. Phạm vi khách và Hội viên

| Tính năng | Khách chưa đăng nhập | Hội viên đăng nhập |
|---|---|---|
| Trang chủ, giới thiệu, hướng dẫn tập | Xem | Xem |
| Tìm/lọc phòng, ảnh và thông tin phòng | Xem dữ liệu công khai | Xem |
| Danh sách HLV đã được duyệt | Xem tên, chuyên môn, giới thiệu và điểm đánh giá tổng hợp được phép công khai | Xem và kiểm tra HLV phù hợp khi chọn giờ |
| Lịch mở cửa/ca nhận đặt, tình trạng còn chỗ | Xem, thử chọn khoảng giờ | Xem và chọn khoảng giờ |
| Tạo/xác nhận đặt lịch hoặc gửi yêu cầu PT | Yêu cầu đăng nhập | Cho phép khi hồ sơ, hạn thẻ và điều kiện nghiệp vụ hợp lệ |
| Lịch của tôi, hủy, đánh giá, hồ sơ cá nhân | Yêu cầu đăng nhập | Theo quyền và trạng thái buổi tập |

Lựa chọn thử của khách là bản nháp, chưa giữ chỗ. Đăng nhập chỉ chứng minh tài khoản hợp lệ; hồ sơ Hội viên và hạn thẻ vẫn cần được kiểm tra riêng. PT/admin/quản lý/lễ tân dùng không gian theo vai trò, không được gửi yêu cầu đặt lịch dưới danh nghĩa Hội viên chỉ vì quay lại từ trang công khai.

## 3. Trang chủ và điều hướng

### Bố cục đề xuất

1. **Header gọn:** thương hiệu, Trang chủ, Phòng tập, Huấn luyện viên, Hướng dẫn; nút Đăng nhập hoặc menu tài khoản theo phiên hiện tại.
2. **Khu vực mở đầu:** một thông điệp chính và ảnh phòng tập phù hợp, CTA “Khám phá phòng tập”, hành động phụ “Xem lịch trống”. Không làm hero quá cao khiến phòng/lịch bị đẩy xuống sâu.
3. **Khám phá phòng:** ảnh, loại phòng, mô tả ngắn, sức chứa và nút “Xem lịch”. Có tìm kiếm/lọc loại phòng, trạng thái hiển thị rõ; chi tiết bảo trì dùng thông báo phù hợp cho khách.
4. **HLV nổi bật/khám phá HLV:** chỉ dữ liệu đã duyệt, giới thiệu và đánh giá tổng hợp; đường dẫn về chọn phòng/giờ có ngữ cảnh.
5. **Cách đặt lịch:** chọn phòng và giờ → chọn tự tập/HLV → xác nhận. Nội dung ngắn, không đẩy người dùng vào quy trình đăng nhập trước khi họ hiểu sản phẩm.
6. **Footer:** thông tin liên hệ và giờ hoạt động đã được xác nhận trong dữ liệu/nội dung cấu hình. Không tự tạo số liệu thành viên, giá, nhận xét hoặc thông tin liên hệ giả.

### Phong cách đề xuất

- Giữ nhận diện xanh đậm hiện có, nền trắng/xám nhạt, một màu nhấn cho hành động chính. Thiết kế thể thao, gọn, hình phòng tập rõ, khoảng cách và cấp độ chữ nhất quán.
- Đưa thông tin cần quyết định lên trước: phòng, ngày, giờ, số chỗ và HLV. Giảm những khối văn bản và thẻ cao chiếm nhiều màn hình.
- Dùng cùng hệ thống header, thẻ phòng, trạng thái, button và khoảng cách giữa phần công khai với phần Hội viên.
- Chọn font hỗ trợ tiếng Việt và độ tương phản dễ đọc; giữ đường viền focus và giảm chuyển động theo tùy chọn hệ thống.

## 4. Màn chọn lịch Hội viên

### Bố cục desktop

```text
Ảnh nhỏ + tên phòng + loại phòng + sức chứa
[Tuần trước] [Tuần hiện tại / chọn ngày] [Tuần sau] [Hôm nay]

LỊCH THEO GIỜ (phần chính)        TÓM TẮT BUỔI TẬP (cố định khi cuộn)
Giờ | T2 | T3 | T4 | T5 | T6 ...  Phòng + ngày
06  |    |    |    |    |         07:00–09:00 · 2 ca · 120 phút
07  |    | [Khối lựa chọn]       Tự tập / HLV đã chọn
08  |    | [chỉnh hai đầu]       Thông báo điều kiện nếu có
09  |    |    |    |    |         [Tiếp tục / Xem lại đặt lịch]

Tự tập hoặc chọn HLV phù hợp với toàn khoảng giờ
```

Lịch tuần là chế độ mặc định trên cả máy tính và điện thoại; có nút chuyển **Tuần / Ngày**. Trên máy tính đủ rộng hiển thị 7 cột. Trên điện thoại vẫn có đủ 7 ngày trong cùng lịch tuần: vuốt ngang bên trong khung lịch để xem các ngày, cuộn dọc để xem giờ; không ép 7 cột thành các ô quá nhỏ. Trục giờ cố định bên trái và tiêu đề ngày bám theo khi cuộn để luôn biết đang xem ngày/giờ nào. Thanh ngày cho biết toàn tuần, ngày hôm nay và ngày đang chọn; chạm ngày đưa tới cột tương ứng. Chuyển tuần dùng nút trước/sau riêng, không nhầm với cuộn giữa các cột. Chỉ khung lịch cuộn ngang, toàn trang không bị tràn ngang. Phần tóm tắt thu gọn ở cuối màn hình, có khoảng đệm để không che nội dung/focus. Không ép người dùng thu nhỏ trình duyệt để thao tác.

### Quy tắc kéo chọn và chỉnh giờ

- Kéo từ ca bắt đầu đến ca cuối trong cùng ngày để tạo **một khối khoảng giờ**; trong lúc kéo hiển thị giờ bắt đầu/kết thúc, số ca và tổng thời lượng.
- Có tay nắm trên/dưới để đổi biên; có thể kéo thân khối để chuyển thời điểm trong cùng ngày và giữ thời lượng nếu toàn vùng đích hợp lệ. Phần đặt lịch không có công cụ mở/đóng lịch quản trị.
- Mọi chuyển động bám mốc các session thật. Ví dụ hệ thống có ca 60 phút thì chỉ chọn được theo ca 60 phút; thay đổi thiết kế không tự mở ra khả năng đặt theo 15/30 phút.
- Bản đầu chỉ chọn một dải liên tục trong một ngày. Chuyển ngày/tuần/phòng phải phản hồi rõ khi xóa hoặc đổi bản nháp, không âm thầm gửi các session cũ.
- Ca đã qua, hết chỗ, đóng/bảo trì, hoặc không đủ điều kiện phải có trạng thái và lý do. Ca còn ít chỗ nổi bật vừa phải. Dải kéo qua phần không hợp lệ hiển thị lỗi và giữ lựa chọn hợp lệ trước đó.
- Hỗ trợ chọn ca đầu/ca cuối bằng hai lần chạm hoặc click, bàn phím và điều khiển giờ bắt đầu/kết thúc để thay thế kéo. Escape hủy thao tác đang kéo. Có nút xóa lựa chọn, phản hồi khi thả ngoài và tự cuộn khi kéo gần mép.
- Trên mobile phân biệt rõ xem/cuộn lịch và chọn giờ, có điều khiển **Xem lịch / Chọn giờ** với trạng thái nhìn thấy được. Khi xem, vuốt ngang/dọc để di chuyển lịch; khi chọn, kéo tạo/chỉnh khoảng giờ trong ngày hoặc dùng hai lần chạm. Tay nắm có vùng chạm đủ lớn, không đòi hỏi kéo chính xác từng pixel. Không chiếm cử chỉ cuộn trang ngoài khung lịch; cho phép hủy chọn và trở lại chế độ xem dễ dàng.
- Phân biệt trạng thái bằng nhãn/biểu tượng và màu, có chú giải gọn. Tải dữ liệu có skeleton, không giật bố cục; hết lịch/lỗi mạng có giải thích và nút thử lại.
- Có hành động “Tìm ngày còn ca gần nhất” khi hôm nay đã hết giờ hoặc không có lịch, và bộ lọc sáng/chiều/tối để giảm phần lịch phải dò. Hiển thị rõ ngày được chuyển đến; không tự thay đổi bản nháp đã chọn.

### Chọn tự tập/HLV và xác nhận

- Sau khi chọn giờ, hiển thị “Tự tập” cùng danh sách HLV khả dụng cho **toàn dải**, dưới dạng thẻ gọn có tên, chuyên môn, điểm đánh giá và trạng thái đang chọn. Tránh danh sách thẻ cao kéo dài như hiện tại.
- Khi đổi giờ, tải lại khả dụng PT, không giữ PT cũ như thể vẫn hợp lệ. Đổi/sắp hết chỗ trong khi thao tác phải được kiểm tra lại trước xác nhận.
- Màn xem lại ghi phòng, ngày/giờ Việt Nam, số ca, thời lượng, tự tập/HLV và điều kiện phản hồi/hủy liên quan. Chỉ một hành động xác nhận; khóa gửi trùng và tiếp tục dùng idempotency hiện có.
- Thành công có mã lịch, trạng thái “Đã xác nhận” hoặc “Chờ PT”, nút “Xem lịch của tôi”. Nếu hết chỗ/trùng lịch thì giữ ngữ cảnh, cập nhật lịch và để người dùng chọn lại, không tự đổi sang ca hoặc HLV khác.

## 5. Đăng nhập đúng thời điểm và giữ ngữ cảnh

Luồng đề xuất:

```text
Trang chủ → Phòng → Xem lịch → Kéo chọn thử
  → Tiếp tục đặt lịch
      → Khách: Đăng nhập → Khôi phục bản nháp → Kiểm tra lại điều kiện
      → Hội viên: Kiểm tra điều kiện
  → Chọn/xác minh HLV → Xem lại → Xác nhận → Kết quả
```

- Không bật hộp đăng nhập khi khách chỉ cuộn trang, lọc phòng hoặc xem lịch. Khách được thử thao tác chọn giờ trước.
- Khách có thể chọn ý định tự tập/cần HLV; kiểm tra PT khả dụng cụ thể cần session Hội viên. Thông báo điều này tại đúng vị trí, không trình bày danh sách khả dụng giả.
- Lưu bản nháp gồm phòng, ngày và ID ca trong trạng thái phiên ngắn hạn, không lưu token/mật khẩu vào bản nháp. Thời hạn bản nháp đề xuất 30 phút; khôi phục chỉ sau khi kiểm tra dữ liệu hiện tại, không giữ chỗ.
- Sau đăng nhập không tự tạo booking: quay về bản nháp để người dùng xem lại và xác nhận. Link quay lại chỉ chấp nhận đường dẫn nội bộ phù hợp vai trò.
- Trường hợp thiếu hồ sơ/hết thẻ/khóa hiển thị một thông báo rõ cùng hướng xử lý liên hệ lễ tân; không chỉ làm mờ toàn bộ lịch mà không giải thích.
- Đăng nhập bằng vai trò khác đưa tới không gian đúng quyền, không tạo booking Hội viên. Tính năng cá nhân vẫn cần xác thực dù truy cập trực tiếp URL/API.

## 6. Các phần kỹ thuật cần làm

| Nhóm | Công việc và đầu ra |
|---|---|
| Phạm vi/nội dung | Chốt các mục công khai, thông điệp, ảnh, contact thực tế; ma trận khách/Hội viên/vai trò khác. |
| Thiết kế | Wireframe desktop/mobile cho trang chủ, danh sách/chi tiết phòng, lịch kéo chọn, đăng nhập tiếp tục và kết quả. Mockup có đủ trạng thái lỗi, rỗng, đang tải, đã chọn và ca khóa. Duyệt mẫu trước khi code ứng dụng. |
| API công khai | Endpoint đọc riêng hoặc projection công khai được giới hạn rõ cho phòng, ảnh, HLV đã duyệt và lịch/ca còn chỗ đã công bố. Phân trang, giới hạn khoảng ngày truy vấn, chỉ trả trường cần hiển thị. Không mở nguyên response quản trị hoặc trả tên người đặt, lịch cá nhân, thông tin liên hệ riêng, chứng chỉ hay đường dẫn file private. Đổi quyền công khai của ảnh phải xét đúng ảnh phòng được phép hiển thị. |
| Route/session | Tách layout công khai và layout cá nhân; `/` truy cập được khi chưa login. Chỉnh Guard và luồng login quay lại đúng chỗ; quản lý bản nháp, đúng vai trò và hết hạn. |
| Component lịch | Chọn dải theo trục thời gian, preview, resize hai đầu, move trong ngày, snap theo session, điều khiển thay thế kéo, trạng thái và tóm tắt responsive. Tái sử dụng logic hợp lệ/transaction; không ghép nguyên grid sửa lịch của admin vào Hội viên. |
| UI phòng/HLV | Tách thành phần đọc công khai khỏi hook bắt buộc session; ảnh và thẻ thống nhất, bố cục PT gọn, tải dữ liệu không nhảy. |
| Kiểm chứng | Guest xem được và không gọi được mutation; login khôi phục ngữ cảnh nhưng vẫn chờ xác nhận; membership/role đúng; ngày giờ khớp; chọn nhiều ca atomic, replay và hết chỗ đúng. Mobile phải truy cập đủ 7 ngày, chuyển Tuần/Ngày, cuộn hai hướng và kéo chọn không xung đột; trục giờ/tiêu đề ngày và tóm tắt không che điều khiển. Kiểm tra browser thật, mobile/touch/keyboard, rồi regression/build/contract/docs theo phần thay đổi. |
| Bàn giao | Hướng dẫn demo khách → login → đặt lịch; báo cáo bằng chứng và danh sách giới hạn; giữ dữ liệu demo dev và test reset riêng. |

API dự kiến có namespace như `/api/v1/public/rooms`, `/public/rooms/:id`, ảnh phòng, lịch phiên và `/public/trainers`; tên/chữ ký cuối chốt trong thiết kế contract trước triển khai. Lịch công khai chỉ thể hiện khả dụng chung; quyền tạo booking và tính khả dụng theo Hội viên vẫn do backend quyết định.

Dữ liệu public đề xuất chỉ gồm phòng đang phục vụ hoặc thông báo bảo trì phù hợp, không trả phòng `INACTIVE`; HLV chỉ `APPROVED`. Mỗi truy vấn lịch giới hạn tối đa 7 ngày trong khoảng lịch đã công bố; danh sách có phân trang. Số chỗ là ảnh chụp tại thời điểm tải, không phải cam kết giữ chỗ.

## 7. Thứ tự và tiêu chí duyệt

1. **U0 — Chốt yêu cầu và mẫu:** duyệt tài liệu này, sau đó dựng mockup trang chủ và chọn lịch desktop/mobile để duyệt hình thức cụ thể.
2. **U1 — Luồng công khai:** layout, dữ liệu public, route và giới hạn quyền; khách truy cập được trang chủ/phòng/HLV/lịch.
3. **U2 — Lịch chọn giờ:** triển khai grid và khối kéo chỉnh, tóm tắt, chọn PT và trạng thái đủ desktop/mobile.
4. **U3 — Nối luồng:** bản nháp → login → kiểm tra lại → xác nhận; dữ liệu/trạng thái sau thành công và ngoại lệ.
5. **U4 — Tổng duyệt:** browser end-to-end khách/Hội viên, kiểm tra quyền API, kéo/chạm/bàn phím và regression; cập nhật hướng dẫn.

### Tiến độ U1 — 31/10/2026

**Cập nhật 01/11/2026:** đã thực hiện tiếp U2–U4 theo yêu cầu người dùng. Lịch kéo/chỉnh dùng session thật đã tích hợp vào trang công khai và Hội viên, có tuần/ngày, cuộn tuần trên điện thoại, hai chế độ xem/chọn, thao tác chạm/bàn phím, bản nháp 30 phút và kiểm tra lại trước xác nhận. Browser QA đã chạy cả tự tập nhiều ca, gửi lại khi mất phản hồi, HLV nhận lịch, hết hạn thẻ, hết chỗ và khôi phục sau hết phiên. Báo cáo hiện hành: [U2–U4](reports/public_booking_u2_u4_2026-11-01.md). Đoạn U1 bên dưới là mốc lịch sử, không còn là giới hạn giao diện hiện tại.

Đã triển khai trang chủ công khai `/`, danh sách phòng `/explore/rooms`, chi tiết phòng/lịch 7 ngày `/explore/rooms/:id`, danh sách HLV `/explore/trainers`. Dữ liệu đọc từ năm endpoint public riêng với projection/visibility được kiểm tra ở backend; API đặt lịch và dữ liệu cá nhân tiếp tục yêu cầu xác thực. Khách bấm đặt được chuyển tới đăng nhập; Hội viên quay lại đúng phòng/ngày, role khác về khu vực tương ứng. Trên điện thoại vẫn đủ 7 ngày bằng cuộn ngang trong khung lịch. U1 là lịch đọc, chưa thay grid chọn giờ cũ bằng khối kéo chỉnh U2. Bằng chứng: `reports/public_discovery_u1_2026-10-31.md`.

Chốt mốc khi: khách xem được không cần tài khoản; thao tác cá nhân/đặt lịch vẫn yêu cầu login đúng quyền; Hội viên chọn 07:00–09:00 bằng một dải và chỉnh lại được; giờ/số ca/thời lượng nhất quán; chuyển qua login không làm mất bản nháp hợp lệ; không tạo booking trước xác nhận; API không rò dữ liệu riêng. Dữ liệu hiện có được giữ nguyên.

**Cấu hình đề xuất để duyệt:** trang chủ công khai + thử chọn lịch cho khách; login tại bước tiếp tục đặt; phong cách nền sáng/xanh đậm; lịch tuần mặc định trên desktop và mobile, có chuyển sang một ngày; mobile cuộn ngang trong khung để truy cập đủ 7 ngày; một dải cùng ngày, có hai tay nắm và kéo chuyển thời điểm trong ngày.
