import {
  createBrowserRouter,
  RouterProvider,
  Navigate,
  Outlet,
  Link,
  useNavigate,
  useLocation,
  useParams,
} from "react-router-dom";
import type { Role } from "./types";
import { lazy, Suspense } from "react";
import { SessionProvider, useSession } from "./app/session";
import { LoginPanel } from "./features/auth/LoginPanel";
import { AdminLayout } from "./layouts/AdminLayout";
import { ProfilePage } from "./features/profile/Profile";
import { PortalLayout } from "./layouts/PortalLayout";
import { CoachSchedule } from "./features/booking/CoachSchedule";
import { OnBehalf } from "./features/booking/OnBehalf";
import { roleHome as landing, loginDestination } from "./app/navigation";
const DiscoveryLayout = lazy(async () => ({
  default: (await import("./features/discovery/Discovery")).DiscoveryLayout,
}));
const DiscoveryHome = lazy(async () => ({
  default: (await import("./features/discovery/Discovery")).DiscoveryHome,
}));
const DiscoveryRooms = lazy(async () => ({
  default: (await import("./features/discovery/Discovery")).DiscoveryRooms,
}));
const DiscoveryRoom = lazy(async () => ({
  default: (await import("./features/discovery/Discovery")).DiscoveryRoom,
}));
const DiscoveryTrainers = lazy(async () => ({
  default: (await import("./features/discovery/Discovery")).DiscoveryTrainers,
}));
const TrainerDirectory = lazy(async () => ({
  default: (await import("./features/booking/TrainerDirectory"))
    .TrainerDirectory,
}));
const CoachAvailability = lazy(async () => ({
  default: (await import("./features/booking/CoachAvailability"))
    .CoachAvailability,
}));
import {
  RoomCatalog,
  RoomBooking,
  BookingList,
  BookingDetail,
} from "./features/booking/Booking";
import "./portal.css";
function LegacyRoomRedirect() {
  const { session } = useSession();
  const { id } = useParams();
  const role = session?.user.role;
  return (
    <Navigate
      replace
      to={
        role === "MEMBER"
          ? `/member/rooms${id ? "/" + id : ""}`
          : role === "PT"
            ? id
              ? `/pt/rooms/${id}/calendar`
              : "/pt/schedule"
            : `/management/read-rooms${id ? "/" + id + "/calendar" : ""}`
      }
    />
  );
}
function HomeRedirect({ profile = false }: { profile?: boolean }) {
  const { session } = useSession();
  return (
    <Navigate
      replace
      to={
        session
          ? profile
            ? session.user.role === "MEMBER"
              ? "/member/profile"
              : session.user.role === "PT"
                ? "/pt/account"
                : "/account"
            : landing(session.user.role)
          : "/login"
      }
    />
  );
}
const UsersPage = lazy(async () => ({
  default: (await import("./features/users/Users")).UsersPage,
}));
const UserFormPage = lazy(async () => ({
  default: (await import("./features/users/Users")).UserFormPage,
}));
const UserDetailPage = lazy(async () => ({
  default: (await import("./features/users/Users")).UserDetailPage,
}));
const MembersPage = lazy(async () => ({
  default: (await import("./features/members/Members")).MembersPage,
}));
const MemberFormPage = lazy(async () => ({
  default: (await import("./features/members/Members")).MemberFormPage,
}));
const MemberDetailPage = lazy(async () => ({
  default: (await import("./features/members/Members")).MemberDetailPage,
}));
const RoomsPage = lazy(async () => ({
  default: (await import("./features/facilities/Facilities")).RoomsPage,
}));
const RoomFormPage = lazy(async () => ({
  default: (await import("./features/facilities/Facilities")).RoomFormPage,
}));
const RoomDetailPage = lazy(async () => ({
  default: (await import("./features/facilities/Facilities")).RoomDetailPage,
}));
const RoomCalendarPage = lazy(async () => ({
  default: (await import("./features/facilities/calendar/RoomCalendar"))
    .RoomCalendarPage,
}));
const PublishedRoomsPage = lazy(async () => ({
  default: (await import("./features/facilities/calendar/PublishedCalendar"))
    .PublishedRoomsPage,
}));
const PublishedCalendarPage = lazy(async () => ({
  default: (await import("./features/facilities/calendar/PublishedCalendar"))
    .PublishedCalendarPage,
}));
const TimeSlotsPage = lazy(async () => ({
  default: (await import("./features/facilities/Facilities")).TimeSlotsPage,
}));
const TimeSlotFormPage = lazy(async () => ({
  default: (await import("./features/facilities/Facilities")).TimeSlotFormPage,
}));
const RoomSessionsPage = lazy(async () => ({
  default: (await import("./features/facilities/Facilities")).RoomSessionsPage,
}));
const SessionFormPage = lazy(async () => ({
  default: (await import("./features/facilities/Facilities")).SessionFormPage,
}));
const SessionDetailPage = lazy(async () => ({
  default: (await import("./features/facilities/Facilities")).SessionDetailPage,
}));
const TrainersPage = lazy(async () => ({
  default: (await import("./features/trainers/Trainers")).TrainersPage,
}));
const TrainerDetailPage = lazy(async () => ({
  default: (await import("./features/trainers/Trainers")).TrainerDetailPage,
}));
const OwnTrainerPage = lazy(async () => ({
  default: (await import("./features/trainers/Trainers")).OwnTrainerPage,
}));
const AdminReportsPage = lazy(async () => ({
  default: (await import("./features/reports/Reports")).AdminReportsPage,
}));
const CoachPerformancePage = lazy(async () => ({
  default: (await import("./features/reports/Reports")).CoachPerformancePage,
}));
const CoachReviewsPage = lazy(async () => ({
  default: (await import("./features/reports/Reports")).CoachReviewsPage,
}));
export { LoginPanel } from "./features/auth/LoginPanel";

export function Guard({ roles }: { roles?: Role[] }) {
  const { session } = useSession();
  const location = useLocation();
  if (!session)
    return (
      <Navigate
        to="/login"
        state={{ from: location.pathname + location.search }}
        replace
      />
    );
  if (roles && !roles.includes(session.user.role))
    return (
      <StatusPage
        title="Bạn không có quyền truy cập"
        description="Liên hệ quản trị viên nếu bạn cần sử dụng chức năng này."
      />
    );
  return <Outlet />;
}
function LoginPage() {
  const auth = useSession();
  const navigate = useNavigate();
  const location = useLocation();
  const from = (location.state as { from?: unknown } | null)?.from;
  if (auth.session)
    return (
      <Navigate to={loginDestination(auth.session.user.role, from)} replace />
    );
  return (
    <LoginPanel
      notice={auth.notice}
      onLogin={(s) => {
        auth.login(s);
        navigate(loginDestination(s.user.role, from), {
          replace: true,
        });
      }}
    />
  );
}
export function StatusPage({
  title,
  description,
}: {
  title: string;
  description: string;
}) {
  return (
    <div className="empty-state">
      <h1>{title}</h1>
      <p>{description}</p>
      <Link to="/me" className="btn btn-secondary">
        Về hồ sơ của tôi
      </Link>
    </div>
  );
}
const router = createBrowserRouter([
  { path: "/login", element: <LoginPage /> },
  {
    element: <DiscoveryLayout />,
    children: [
      { path: "/", element: <DiscoveryHome /> },
      { path: "/explore/rooms", element: <DiscoveryRooms /> },
      { path: "/explore/rooms/:id", element: <DiscoveryRoom /> },
      { path: "/explore/trainers", element: <DiscoveryTrainers /> },
    ],
  },
  {
    element: <Guard />,
    children: [
      { path: "/rooms", element: <LegacyRoomRedirect /> },
      { path: "/rooms/:id/calendar", element: <LegacyRoomRedirect /> },
      { path: "/member/book", element: <HomeRedirect /> },
      { path: "/me", element: <HomeRedirect profile /> },
      {
        element: <Guard roles={["MEMBER"]} />,
        children: [
          {
            element: <PortalLayout />,
            children: [
              { path: "/member/rooms", element: <RoomCatalog /> },
              { path: "/member/rooms/:id", element: <RoomBooking /> },
              { path: "/member/bookings", element: <BookingList /> },
              { path: "/member/trainers", element: <TrainerDirectory /> },
              { path: "/member/bookings/:id", element: <BookingDetail /> },
              { path: "/member/profile", element: <ProfilePage /> },
            ],
          },
        ],
      },
      {
        element: <Guard roles={["PT"]} />,
        children: [
          {
            element: <PortalLayout coach />,
            children: [
              { path: "/pt/schedule", element: <CoachSchedule /> },
              { path: "/pt/reviews", element: <CoachReviewsPage /> },
              { path: "/pt/performance", element: <CoachPerformancePage /> },
              { path: "/pt/history", element: <BookingList history /> },
              { path: "/pt/requests", element: <BookingList requests /> },
              { path: "/pt/bookings/:id", element: <BookingDetail /> },
              { path: "/pt/profile", element: <OwnTrainerPage /> },
              {
                path: "/pt/rooms/:id/calendar",
                element: <PublishedCalendarPage />,
              },
              { path: "/pt/availability", element: <CoachAvailability /> },
              { path: "/pt/account", element: <ProfilePage /> },
            ],
          },
        ],
      },
      {
        element: <Guard roles={["ADMIN", "MANAGER", "RECEPTIONIST"]} />,
        children: [
          {
            element: <AdminLayout />,
            children: [
              { path: "/account", element: <ProfilePage /> },
              {
                path: "/management/attendance",
                element: <BookingList attendance />,
              },
              { path: "/management/bookings", element: <BookingList /> },
              { path: "/management/bookings/:id", element: <BookingDetail /> },
              {
                element: <Guard roles={["ADMIN", "RECEPTIONIST"]} />,
                children: [
                  { path: "/management/bookings/new", element: <OnBehalf /> },
                  {
                    path: "/management/bookings/new/rooms",
                    element: <RoomCatalog operations />,
                  },
                  {
                    path: "/management/bookings/new/rooms/:id",
                    element: <RoomBooking operations />,
                  },
                ],
              },
              {
                path: "/management/read-rooms",
                element: <PublishedRoomsPage />,
              },
              {
                path: "/management/read-rooms/:id/calendar",
                element: <PublishedCalendarPage />,
              },
              {
                element: <Guard roles={["ADMIN"]} />,
                children: [
                  { path: "/admin/users", element: <UsersPage /> },
                  { path: "/admin/users/new", element: <UserFormPage /> },
                  { path: "/admin/users/:id", element: <UserDetailPage /> },
                  { path: "/admin/users/:id/edit", element: <UserFormPage /> },
                ],
              },
              {
                element: <Guard roles={["ADMIN", "MANAGER", "RECEPTIONIST"]} />,
                children: [
                  { path: "/management/members", element: <MembersPage /> },
                  {
                    path: "/management/members/new",
                    element: <MemberFormPage />,
                  },
                  {
                    path: "/management/members/:id",
                    element: <MemberDetailPage />,
                  },
                  {
                    path: "/management/members/:id/edit",
                    element: <MemberFormPage />,
                  },
                ],
              },
              {
                element: <Guard roles={["ADMIN", "MANAGER"]} />,
                children: [
                  {
                    path: "/management/reports",
                    element: <AdminReportsPage />,
                  },
                  { path: "/management/trainers", element: <TrainersPage /> },
                  {
                    path: "/management/trainers/:id",
                    element: <TrainerDetailPage />,
                  },
                  { path: "/management/rooms", element: <RoomsPage /> },
                  { path: "/management/rooms/new", element: <RoomFormPage /> },
                  {
                    path: "/management/rooms/:id",
                    element: <RoomDetailPage />,
                  },
                  {
                    path: "/management/rooms/:id/edit",
                    element: <RoomFormPage />,
                  },
                  {
                    path: "/management/rooms/:id/calendar",
                    element: <RoomCalendarPage />,
                  },
                  {
                    path: "/management/time-slots",
                    element: <TimeSlotsPage />,
                  },
                  {
                    path: "/management/time-slots/new",
                    element: <TimeSlotFormPage />,
                  },
                  {
                    path: "/management/time-slots/:id/edit",
                    element: <TimeSlotFormPage />,
                  },
                  {
                    path: "/management/room-sessions",
                    element: <RoomSessionsPage />,
                  },
                  {
                    path: "/management/room-sessions/new",
                    element: <SessionFormPage />,
                  },
                  {
                    path: "/management/room-sessions/:id",
                    element: <SessionDetailPage />,
                  },
                  {
                    path: "/management/room-sessions/:id/edit",
                    element: <SessionFormPage />,
                  },
                ],
              },
              {
                path: "*",
                element: (
                  <StatusPage
                    title="Không tìm thấy trang"
                    description="Đường dẫn này không tồn tại. Bạn có thể trở về hồ sơ của mình."
                  />
                ),
              },
            ],
          },
        ],
      },
    ],
  },
]);
export function App() {
  return (
    <SessionProvider>
      <Suspense
        fallback={
          <div className="state-panel" role="status">
            Đang tải trang…
          </div>
        }
      >
        <RouterProvider router={router} />
      </Suspense>
    </SessionProvider>
  );
}
