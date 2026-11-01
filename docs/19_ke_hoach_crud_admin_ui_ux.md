# Kế hoạch CRUD trang quản trị và thiết kế lại UI/UX

**Ngày lập:** 11/10/2026. **Trạng thái:** Người dùng đã yêu cầu triển khai; backend và UI các chặng 1–4 đã có. Giới hạn kiểm chứng được ghi trong báo cáo bên dưới.

**Phạm vi đã được người dùng chọn:** Tài khoản + hội viên + phòng + ca/phiên phòng + PT, chia chặng. Kế hoạch này cụ thể hóa và ưu tiên phần CRUD/UI trong `18_ke_hoach_lam_viec_tiep.md`. Sau mỗi module backend sẽ làm giao diện tương ứng, thay vì chờ gom toàn bộ UI nghiệp vụ vào tuần 7.

## 1. Mục tiêu bàn giao

Một khu vực quản trị tiếng Việt, có điều hướng bằng URL, bảng dữ liệu dễ tra cứu, form thêm/sửa rõ ràng, thao tác nhạy cảm có xác nhận và giao diện dùng được ở 375px. Thao tác được nghiệm thu phải lưu vào PostgreSQL thật và tuân thủ quyền v0.3.

CRUD theo vòng đời nghiệp vụ: tạo → xem → sửa → khóa/ngừng hoạt động/đình chỉ. Dữ liệu liên quan lịch sử được giữ theo FK RESTRICT và chính sách v0.3; không bổ sung API xóa cứng tài khoản/hội viên/phòng/PT chỉ để đủ bốn nút CRUD.

Booking, Attendance, Review và Reports tiếp tục theo kế hoạch chặng sau. Gói này chuẩn bị dữ liệu nền và UI dùng lại được cho chúng.

## 2. Hiện trạng đã rà soát trong code

| Vấn đề | Bằng chứng hiện tại | Cách sửa dự kiến |
|---|---|---|
| Thao tác quản trị rời rạc | `UserRow` dùng nhiều `window.prompt/confirm`; mật khẩu mới nhập trong prompt dạng text. | Dialog theo từng nghiệp vụ; mật khẩu che mặc định, có nút hiện/ẩn; lý do và cảnh báo nằm cùng form. |
| Bảng khó mở rộng | Mỗi dòng chứa ba nút nhỏ; table cuộn ngang trên màn hẹp. | Một action menu rõ nhãn; desktop dùng table, mobile dùng danh sách card có đủ dữ liệu/thao tác. |
| Điều hướng chỉ ở state | `page` chỉ có profile/users; không có URL từng màn hình. | Route, layout chung, guard theo quyền, trang 403/404 và breadcrumb khi cần. |
| Form và phản hồi chưa thống nhất | Form tạo user chưa có trạng thái submitting; lỗi thường nằm chung một chỗ; notice đăng xuất và lỗi đăng nhập cùng kiểu đỏ. | Validation theo field, giữ dữ liệu khi lỗi, chống submit lặp, phân biệt thông báo thành công/cảnh báo/lỗi. |
| Visual chưa có hệ thống | Màu/class lặp trong `App.tsx`; stylesheet chỉ có font và focus toàn cục. | Token màu/khoảng cách/font; component dùng lại cho button, badge, form, dialog, table và page header. |
| CRUD tài khoản còn thiếu API | Router hiện có list/create/status/role/reset, chưa có `GET /users/{id}` hoặc `PATCH /users/{id}` cho Admin. | Bổ sung đọc chi tiết và sửa thông tin liên hệ theo allowlist, quyền Admin và audit. |
| Các module khác chưa chạy | Router hiện chỉ đăng ký Identity/health dù schema đã có bảng nghiệp vụ. | Triển khai backend từng module rồi kết nối UI; menu của module được đưa vào bản chạy khi chặng tương ứng hoạt động. |

Đây là rà soát tĩnh code. Chưa có kiểm tra hình ảnh/bố cục thực tế trong browser trong lần lập kế hoạch này. Test frontend hiện có một case gửi thông tin đăng nhập qua API mock và gọi callback, chưa phủ CRUD hoặc responsive.

## 3. Hướng thiết kế đề xuất

**Phong cách:** giao diện công cụ quản trị tối giản, bố cục rõ, nền sáng; sidebar xanh đậm và màu cam tiết chế cho hành động chính. Nội dung bảng và form là trọng tâm; icon SVG một bộ, nhãn tiếng Việt cụ thể.

### Màu và typography

| Token | Giá trị đề xuất | Sử dụng |
|---|---|---|
| Background / surface | `#F5F6F8` / `#FFFFFF` | Nền trang và vùng dữ liệu. |
| Text / muted text | `#0F172A` / `#475569` | Nội dung chính và thông tin phụ. |
| Sidebar / sidebar text | `#0F172A` / `#CBD5E1` | Điều hướng, nhóm menu và tên người dùng. |
| Primary / on-primary | `#C2410C` / `#FFFFFF` | Một nút hành động chính mỗi màn hình. |
| Border | `#E2E8F0` | Phân chia vùng/table; control có border/focus đủ nhận biết. |
| Control border / focus | `#64748B` / `#C2410C` | Viền input và focus rõ hơn đường phân chia trang. |
| Success | `#166534` trên `#DCFCE7` | Hoạt động/đã duyệt, kèm chữ hoặc icon. |
| Warning | `#92400E` trên `#FEF3C7` | Chờ duyệt/sắp hết hạn, kèm chữ. |
| Danger | `#991B1B` trên `#FEE2E2` | Lỗi/đã khóa; nút nhạy cảm có style riêng. |

Các cặp chữ/nền trên đã được tính tỷ lệ tương phản từ mã màu: primary 5.18:1, muted text 7.58:1, sidebar text 12.02:1; success/warning/danger đều trên 6:1. Đây là kiểm tra palette đề xuất, chưa chứng minh UI sau triển khai đạt accessibility.

Font đề xuất: `Segoe UI, Inter, system-ui, sans-serif`, ưu tiên hỗ trợ tiếng Việt và chạy demo local ổn định. Nội dung/form 16px, bảng desktop 14px, tiêu đề trang 24–28px. Một thang spacing 4/8/12/16/24/32px; radius 8–12px, shadow nhẹ; chuyển trạng thái 150–200ms và tôn trọng reduced motion.

### Bố cục và component

- Desktop từ 1024px: sidebar khoảng 240px, topbar khoảng 64px, nội dung có padding 24–32px và giới hạn chiều rộng khoảng 1440px. Sidebar có nhóm Quản trị / Dữ liệu nền / Cá nhân, menu active dễ nhận biết.
- Tablet/mobile: menu mở bằng drawer, padding 16px; filter xuống dòng, table chuyển card; form dài dùng trang riêng có nút quay lại. Control thao tác chính cao khoảng 44px; giữ chữ đọc được và không tràn ngang toàn trang.
- Mỗi màn hình có PageHeader gồm tên, mô tả ngắn và nút `Thêm …`; dưới là search/filter, danh sách và pagination. Chỉ số tổng lấy từ API và ghi đúng ngữ cảnh filter; dashboard thống kê nằm ở chặng Reports.
- Form tài khoản/hội viên/phòng dài dùng route tạo/sửa; thao tác khóa, đổi role, reset hoặc duyệt dùng Dialog/AlertDialog. Chi tiết ngắn có thể dùng Sheet để giữ ngữ cảnh danh sách. Dialog phải quản lý focus, Esc và trả focus về nút mở.
- Dùng React/TypeScript/Vite/Tailwind hiện có, bổ sung các component shadcn/ui cần thiết với bộ primitive Radix nhất quán. Token màu đặt trong CSS variables và Tailwind `@theme`; chọn một bộ icon SVG, ưu tiên Lucide để đồng bộ với component.

Wireframe desktop cho màn danh sách:

```text
┌───────────────────┬───────────────────────────────────────────────────────┐
│ GYM               │ Quản trị / Tài khoản                      Người dùng │
│                   ├───────────────────────────────────────────────────────┤
│ Tài khoản         │ Tài khoản hệ thống                  [ + Thêm mới ]   │
│ Hội viên          │ Quản lý quyền và trạng thái tài khoản                │
│ Phòng tập         │                                                       │
│ Ca & phiên phòng  │ [Tìm tên / email...] [Vai trò ▾] [Trạng thái ▾]       │
│ PT                │ Bộ lọc đang dùng ...                    [Xóa bộ lọc] │
│                   │                                                       │
│ Cá nhân           │ Tài khoản         Vai trò       Trạng thái   Thao tác │
│ Hồ sơ của tôi     │ Họ tên + liên hệ  Vai trò       Badge chữ       [⋮] │
│                   │ ...                                                   │
│                   │ 1–20 / tổng từ API     [20/trang ▾] [Trước] [Sau]    │
└───────────────────┴───────────────────────────────────────────────────────┘
```

Login được thiết kế cùng hệ thống: form gọn, label rõ, hiện/ẩn mật khẩu, hỗ trợ autofill/paste, trạng thái đang đăng nhập. Không điền sẵn mật khẩu; thông tin tài khoản demo nằm trong hướng dẫn local. Trang hồ sơ cá nhân dùng cùng form và thông báo với khu vực quản trị.

## 4. Phạm vi CRUD và quyền từng module

| Module | Tạo / xem / sửa | Vòng đời thay cho xóa | Quyền |
|---|---|---|---|
| Tài khoản | List, search/filter/page size, chi tiết, tạo; sửa họ tên/email/SĐT; role/reset tách riêng. Username giữ read-only sau tạo. | Khóa/mở; cảnh báo thu hồi token; bảo vệ Admin hoạt động cuối cùng tại backend. | Admin; mọi role có `/me` cho thông tin của mình. |
| Hội viên | Tạo tài khoản + profile trong một transaction; liên kết tài khoản MEMBER đã có khi đủ điều kiện; xem/sửa profile, ngày tham gia/hạn thẻ, tìm mã/tên/SĐT/email. | ACTIVE/INACTIVE cho profile; trạng thái tài khoản và hết hạn thẻ hiển thị riêng. | Admin/Manager/Receptionist quản lý; Member SELF, PT ASSIGNED với dữ liệu tối thiểu. |
| Phòng | Tên, loại, mô tả, capacity mặc định; danh sách/chi tiết/sửa. | AVAILABLE/MAINTENANCE/INACTIVE; preview tác động trước đổi trạng thái có lịch liên quan. | Admin/Manager quản lý; các role đọc dữ liệu phù hợp. |
| Ca / phiên phòng | Ca mẫu có giờ bắt đầu/kết thúc; tạo phiên room/date/slot và capacity snapshot; xem lịch phiên, sửa capacity/status hợp lệ. | Ca đã dùng chỉ ngừng hoạt động; phiên OPEN/CLOSED/MAINTENANCE kèm reason. | Admin/Manager quản lý; các role đọc dữ liệu phù hợp. |
| PT | Admin tạo tài khoản PT; PT tự tạo/sửa/nộp hồ sơ chuyên môn và chứng chỉ; Admin/Manager xem hồ sơ và quản lý quyết định/lịch theo quyền. | Duyệt/từ chối/đình chỉ và lưu lịch sử; trạng thái profile độc lập account. | Profile update/submit là PT SELF; duyệt và lịch vận hành theo ma trận, certificate riêng tư. |

PT được quản lý theo workflow v0.3, không gán quyền sửa hồ sơ chuyên môn của người khác cho Admin qua một form CRUD chung. Thông tin liên hệ tài khoản PT được Admin sửa ở module Tài khoản.

## 5. Chặng triển khai và điều kiện hoàn thành

| Chặng | Công việc | Đầu ra và cổng nghiệm thu |
|---|---|---|
| 0 — Chốt thiết kế | Dựng preview Login, danh sách Tài khoản, form thêm/sửa và dialog khóa ở desktop/mobile; review palette, density và cách dùng form. | Preview dùng dữ liệu minh họa có nhãn; hình/bố cục được kiểm tra ở 375px và desktop; chốt token trước khi áp dụng đồng loạt. |
| 1 — Nền UI + Tài khoản | Tách `App.tsx`, route/layout/auth guard, component chung; thay prompt/confirm; bổ sung API chi tiết/sửa Admin; hoàn thiện search/filter/pagination và profile. | Admin tạo → xem → sửa → khóa/mở → đổi role/reset bằng UI lưu DB thật; non-Admin 403; 401 hết phiên; form lỗi giữ dữ liệu, submit không lặp. |
| 2 — Hội viên | Domain/repository/service/handler, API và UI danh sách/chi tiết/form/membership; tạo mới hoặc liên kết MEMBER chưa có profile. | Transaction account/profile/audit; kiểm tra trùng, quyền object và ngày thẻ; card mobile đủ thao tác; frontend không gán role tùy ý qua member.create. |
| 3 — Phòng, ca, phiên | Backend và UI cho phòng, time slot, room session; form thời gian/ngày/capacity, preview tác động thay đổi. | Ca không chồng kể cả request đồng thời; ca liền kề hợp lệ; snapshot phiên đúng timezone; không giảm capacity dưới số chỗ giữ hoặc phá lịch. |
| 4 — PT | Portal PT nộp hồ sơ, file policy; giao diện Admin/Manager duyệt/đình chỉ; availability/unavailability và quyết định có lý do. | Certificate kiểm tra quyền mỗi lần tải; transition/audit đúng; lịch sửa không phá booking; account PT đơn thuần chưa được coi là hồ sơ APPROVED. |
| 5 — Kiểm thử và bàn giao | Browser/E2E, responsive, keyboard, regression, dữ liệu demo, OpenAPI/Postman và runbook. | Test/build và DB/API suite PASS; màn hình nối DB thật; báo cáo lỗi/giới hạn rõ, commit theo chặng. |

Ưu tiên chạy trọn Chặng 1 trước rồi làm tuần tự Chặng 2–4. Sau Chặng 1 cập nhật ước lượng lịch theo kết quả và lỗi thực tế; các cổng dữ liệu nền trong kế hoạch tuần 5 vẫn là điều kiện trước Booking.

## 6. Những bổ sung backend/contract cần làm trước khi nối UI

1. **Tài khoản:** đề xuất `GET /api/v1/users/{id}` và `PATCH /api/v1/users/{id}` chỉ Admin; PATCH allowlist `full_name`, `email`, `phone`, không sửa username/role/status/password qua payload này. Update + audit commit cùng transaction; không trả password hash/token. Tìm kiếm/filter/pagination chạy ở server, page size mặc định 20, tối đa 100. Sort chỉ bật khi API có contract và allowlist trường rõ ràng.
2. **Hội viên:** chốt payload `POST /members` cho đúng một trong hai cách: tạo account MEMBER mới hoặc `user_id` của account MEMBER chưa có profile. Giữ quyền member.create cho nghiệp vụ này, không mở account.manage cho Receptionist/Manager. Đồng bộ hồ sơ/contact về nguồn dữ liệu đúng và audit thay đổi; không tự ghi đè mật khẩu account có sẵn.
3. **Phòng/ca/phiên:** triển khai endpoint theo `12_dac_ta_API.md`; thống nhất `status` với `is_active`, field `room_type`/payload và snapshot thời gian từ schema trước khi code form. Chốt response preview impact nếu UI cần trước khi cập nhật phòng/phiên có lịch.
4. **PT:** triển khai list/profile/submit/certificate/decision/availability; chốt contract tìm và xem mọi trạng thái dành cho Manager/Admin cùng object scope. Mapping state DRAFT/PENDING/APPROVED/REJECTED/SUSPENDED theo schema. Profile creation/upload thuộc PT SELF, quyết định duyệt thuộc vận hành.
5. Bổ sung contract còn thiếu cho đọc chi tiết room/PT và các input/response/error cần cho form trước khi UI gọi endpoint. Cập nhật API/OpenAPI/Postman trong cùng chặng, ghi decision log nếu thay đổi chính sách quyền hoặc nghiệp vụ.

Giữ backend theo `handler/v1 → service/v1 → repository`, domain độc lập Gin, `routes/v1`, `dto/v1`, `app`, `db/migrations` đúng project Gin tham chiếu đã yêu cầu. Nếu thay schema DB đã áp dụng, thêm migration mới; không sửa migration v1 hoặc reset DB dev. DeepSeek Harness local hiện tại được giữ nguyên.

## 7. Quy tắc UX khi thao tác dữ liệu thật

- URL giữ trạng thái search/filter/page/page size; đổi filter quay về trang 1. Search debounce khoảng 300ms; bỏ qua/hủy request cũ để kết quả trả chậm không ghi đè kết quả mới. Không sort riêng trang hiện tại rồi trình bày như toàn bộ danh sách đã sort.
- List có skeleton khi tải đầu, phản hồi khi refresh, lỗi kèm `Thử lại`, trạng thái chưa có dữ liệu và không có kết quả lọc khác nhau. Sau mutation cập nhật/refetch đúng query; nếu trang cuối rỗng do thay trạng thái thì điều chỉnh page hợp lệ.
- Field có label, required marker, helper và lỗi ngay dưới; email dùng `type=email`, SĐT `type=tel`; validate trên blur/submit. Nhiều lỗi có summary đưa focus đúng nơi; lỗi server giữ nguyên input. Nếu API chưa có field error, dùng code/message ổn định và error banner, không đoán từ SQL.
- Button lưu có submitting và chống click lặp. Đóng form có dữ liệu chưa lưu phải có lựa chọn ở lại/bỏ thay đổi; lỗi không được đóng form như đã thành công.
- Action menu có `Xem`, `Chỉnh sửa` và nhóm thao tác nhạy cảm riêng, đúng quyền thực tế. Nhãn giao diện dùng `Đặt lại mật khẩu`, `Khóa tài khoản`, `Ngừng hoạt động` thay vì mã kỹ thuật. Thao tác bị từ chối có lý do và cách xử lý rõ; kết quả 403/409 từ backend được giữ nguyên ý nghĩa.
- Dialog khóa/đổi role/đình chỉ ghi tên đối tượng, lý do bắt buộc và hậu quả; nút `Hủy` dễ thấy. Reset password dùng input password + xác nhận nhập lại, hỗ trợ paste; sau revoke token của chính người dùng phải giải thích cần đăng nhập lại. Mở khóa không làm token cũ sống lại.
- Thành công dùng thông báo ngắn; lỗi cần giữ lại để xử lý. Thông báo không giành focus; màu luôn kèm chữ/icon. Khi hết phiên, xóa token memory và về login, tránh tự gửi lại mutation nhạy cảm sau đăng nhập.

## 8. Tổ chức frontend dự kiến

```text
client/src/
  app/          # router, providers, guards
  layouts/      # AdminLayout, AuthLayout
  components/
    ui/         # primitive shadcn/ui đã chọn
    shared/     # PageHeader, Filters, Pagination, Error/Empty state
  features/
    auth/
    profile/
    users/      # pages, forms, hooks, types
    members/
    facilities/
    trainers/
  services/     # fetch client, ApiError, request_id, AbortSignal
  styles/       # tokens và global styles
```

Route đề xuất: `/login`, `/me`, `/admin/users`, `/admin/users/new`, `/admin/users/:id`, `/admin/users/:id/edit`; các trang dữ liệu nền dùng nhóm `/management/members`, `/management/rooms`, `/management/time-slots`, `/management/room-sessions`, `/management/trainers` với guard theo từng permission. Receptionist chỉ vào phần hội viên phù hợp; Manager không có quyền tài khoản; PT có `/pt/profile` và lịch của mình.

Giữ fetch client typed hiện có và token memory; router hỗ trợ deep link/back, guard 401/403; page chi tiết dùng route thay vì phụ thuộc danh sách đã tải. Khi thêm thư viện UI/router/form, chỉ thêm phần cần và cập nhật lockfile bằng `npm.cmd`, chốt phiên bản tương thích toolchain hiện có. Các màn mới dùng cùng token/component; không viết tiếp toàn bộ trong một `App.tsx`.

## 9. Nghiệm thu theo chặng

- **Backend/DB:** Go unit/vet; duplicate username/email/member code/room name; transaction + audit rollback; RBAC/object scope; admin cuối; role/profile; constraint thời gian/capacity; upload và certificate access. Có case âm, boundary và concurrent cho invariants liên quan.
- **Frontend:** render đúng quyền; nhập/sửa/lưu/hủy; validation, lỗi server, loading và submit lặp; search/filter/page/reset; stale request; menu/dialog bằng keyboard; 401/403/409; dirty form và token revoke. Kiểm tra hành vi thay vì chỉ snapshot/class name.
- **Browser:** 375/768/1024/1440px, zoom 200%, nội dung tiếng Việt dài, keyboard Tab/Enter/Esc, focus không bị che, reduced motion; kiểm tra tương phản cặp màu thực tế, không chỉ palette. Chụp màn hình các màn được bàn giao để review visual.
- **Luồng DB thật:** Admin tạo account → sửa contact → khóa/mở; tạo hoặc liên kết Member → sửa hạn thẻ; tạo phòng/ca/phiên; PT nộp profile/certificate → Manager/Admin duyệt → quản lý lịch. Reload/read lại xác nhận dữ liệu đã lưu.
- Cập nhật OpenAPI/Postman, hướng dẫn và báo cáo test từng chặng; chỉ đánh dấu hoàn thành phần đã chạy và có bằng chứng.

## 10. Skill và nguồn áp dụng

Đã đọc `ui-ux-pro-max` và `ui-styling`, cùng hướng dẫn dialog accessibility/responsive. Search thiết kế tổng thể trả về pattern landing/marketing sau một lần thu hẹp query, nên không dùng/persist pattern đó cho admin CRUD. Kế hoạch này tổng hợp phần style tối giản phù hợp dashboard, hướng dẫn UX đã đối chiếu và hiện trạng code; palette sáng và bố cục CRUD là đề xuất cụ thể của Codex.

Tài liệu kỹ thuật chính thức đã đối chiếu khi lập kế hoạch:

- [shadcn/ui trên Vite](https://ui.shadcn.com/docs/installation/vite).
- [Radix Dialog](https://ui.shadcn.com/docs/components/radix/dialog), [Sheet](https://ui.shadcn.com/docs/components/radix/sheet), [Dropdown Menu](https://ui.shadcn.com/docs/components/radix/dropdown-menu).
- [Tailwind theme variables](https://tailwindcss.com/docs/theme).

## 11. Kết quả thực hiện

Các mục hiện trạng ở mục 2 là snapshot trước triển khai. Bản hiện tại có route/layout/guard, token giao diện, dialog và action menu Radix, bảng desktop/card mobile, form thêm/sửa riêng, dữ liệu nền và workflow PT nối DB thật. Backend giữ `handler/v1 → service/v1 → repository` theo project Gin tham chiếu.

Snapshot bàn giao CRUD có 49 operations; sau bổ sung lịch phòng, `openapi_admin_crud.json` và `GYM_Admin_CRUD.postman_collection.json` hiện có 58 operations, đối chiếu router bằng `scripts/generate_api_contracts.py --check`. Hướng dẫn: `17_huong_dan_chay_local.md`. Bằng chứng CRUD: `reports/admin_crud_verification_2026-10-11.md`; lịch phòng: `reports/room_calendar_verification_2026-10-11.md`. Booking/Attendance/Review/Reports vẫn chưa triển khai; không coi gói CRUD này là MVP GYM hoàn chỉnh.

## 12. Bổ sung lịch hoạt động phòng theo yêu cầu mới

Người dùng yêu cầu Admin/Manager tạo phòng và đăng ký giờ hoạt động bằng bảng giống Excel 7 hàng ngày × 24 cột giờ, kéo chọn/kéo thả và xác nhận; đã chọn **lặp hằng tuần + ngoại lệ theo ngày**. Kế hoạch chi tiết tại [20 — Lịch hoạt động phòng](20_ke_hoach_lich_hoat_dong_phong.md) đã được triển khai local; hướng dẫn tại tài liệu 21 và giới hạn kiểm chứng tại báo cáo lịch phòng. Kết quả CRUD trước đó vẫn giữ làm snapshot.

Lịch mở cửa riêng từng phòng bổ sung cho ca chung/phiên ngày hiện có. Chọn 07:00–21:00 tương ứng 14 ô; phiên chỉ sinh từ ca nằm trọn giờ mở. Lưu version, preview/xác nhận, giữ snapshot/lịch sử và chặn thay đổi ảnh hưởng booking giữ chỗ. Migration v3 thêm bốn bảng; phòng legacy chỉ bật gate khi quản lý chủ động công bố lịch đầu tiên. Phòng mới có gate ngay.
