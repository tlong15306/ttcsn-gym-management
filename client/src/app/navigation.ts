import type { Role } from "../types";

export function roleHome(role: Role) {
  return role === "MEMBER"
    ? "/member/rooms"
    : role === "PT"
      ? "/pt/schedule"
      : role === "ADMIN"
        ? "/admin/users"
        : role === "MANAGER"
          ? "/management/rooms"
          : "/management/members";
}

// Return only to an internal route belonging to the authenticated role.
export function loginDestination(role: Role, from: unknown) {
  if (
    typeof from !== "string" ||
    !from.startsWith("/") ||
    from.startsWith("//") ||
    from.includes("\\")
  )
    return roleHome(role);
  const path = from.split(/[?#]/)[0];
  const allowed =
    role === "MEMBER"
      ? /^\/member\/(rooms(?:\/\d+)?|bookings(?:\/\d+)?|trainers|profile)$/
      : role === "PT"
        ? /^\/pt\//
        : role === "ADMIN"
          ? /^\/(admin|management)\//
          : /^\/management\//;
  return allowed.test(path) ? from : roleHome(role);
}
