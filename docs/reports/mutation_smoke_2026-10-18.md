# Newman mutation smoke cô lập — 18/10/2026

## Mục tiêu và cô lập môi trường

Collection `GYM_Mutation_Smoke.postman_collection.json` kiểm tra một vòng đời tài khoản tổng hợp: Admin tạo tài khoản Lễ tân synthetic → Lễ tân đăng nhập → Admin khóa/mở → token cũ bị thu hồi → reset mật khẩu → token/mật khẩu cũ bị từ chối → mật khẩu mới đăng nhập được → Lễ tân bị 403 khi quản lý tài khoản.

Lệnh chạy:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\run-api-mutation-smoke.ps1
```

Runner chỉ dùng `ttcsn_gym_test`: reset/migrate trước khi chạy, seed 5 demo account, build API tạm trên `127.0.0.1:18080`, chạy Newman Docker qua `host.docker.internal`, thu audit rồi tắt API tạm và reset DB test lần nữa. Script từ chối cổng đang dùng; không gọi server dev `8080`, client `5173` hay database development.

## Kết quả

| Hạng mục | Kết quả |
|---|---|
| Newman request | 14/14 hoàn tất |
| Newman assertions | 14/14 PASS |
| Tạo account | `201`, không trả `password_hash` |
| Lock/unlock | `200`; token trước lock và trước unlock đều `401` |
| Reset mật khẩu | `200`; token trước reset và mật khẩu cũ đều `401`; mật khẩu mới `200` |
| RBAC | Lễ tân `GET /users` trả `403` |
| Audit | 4 event: tạo account, khóa, mở, reset password |
| Cleanup | API tạm đã dừng; `ttcsn_gym_test` còn migration v6, `users=0` |

Summary không có token: `.artifact_work/api-mutation-smoke/mutation-smoke-20261018-123109.summary.json`. Newman JSON có trong cùng thư mục artifact để xem khi cần, không commit.

## Giới hạn

Đây là smoke Identity mutation. Nó chưa thay collection cho Room/Calendar/Booking/Attendance/Review/Report, browser E2E, race tải hay kiểm thử production. Test database bị reset theo chủ đích; không dùng nó làm lịch sử demo.
