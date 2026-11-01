import { Link } from "react-router-dom";
import { SchedulePanel } from "../trainers/Trainers";
import { RecordState, useRecord } from "../../components/shared/record";
export function CoachAvailability() {
  const profile = useRecord<{ id: number }>("pts/me");
  if (profile.missing)
    return (
      <>
        <h1>Giờ làm việc</h1>
        <p>Hoàn thiện hồ sơ PT trước khi đăng ký giờ làm.</p>
        <Link to="/pt/profile">Tạo hồ sơ chuyên môn →</Link>
      </>
    );
  if (!profile.data)
    return <RecordState {...profile} retry={profile.refresh} />;
  return (
    <>
      <h1>Giờ làm việc và ngày nghỉ</h1>
      <p>Lịch rảnh dùng để kiểm tra toàn bộ ca tập khi hội viên chọn PT.</p>
      <SchedulePanel prefix="pts/me" />
    </>
  );
}
