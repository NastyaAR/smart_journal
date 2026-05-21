import type {
  Achievement,
  Grade,
  GradeView,
  Group,
  LoginResponse,
  Merch,
  MessageResponse,
  Purchase,
  PurchaseResult,
  SessionResponse,
  StudentGroupResponse,
  Subject,
} from "./types";

const API_BASE = import.meta.env.VITE_API_BASE || "/api";

export class ApiError extends Error {
  status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      ...(options.body ? { "Content-Type": "application/json" } : {}),
      ...options.headers,
    },
    credentials: "include",
  });

  const contentType = response.headers.get("content-type") || "";
  const isJson = contentType.includes("application/json");
  const body = isJson ? await response.json() : await response.text();

  if (!response.ok) {
    const message =
      typeof body === "object" && body && "error" in body
        ? String(body.error)
        : typeof body === "string" && body
          ? body
          : "Запрос не выполнен";
    throw new ApiError(message, response.status);
  }

  return body as T;
}

const post = <T>(path: string, payload?: unknown) =>
  request<T>(path, {
    method: "POST",
    body: payload === undefined ? undefined : JSON.stringify(payload),
  });

const requestArray = async <T>(path: string): Promise<T[]> => {
  const body = await request<T[] | null>(path);
  return Array.isArray(body) ? body : [];
};

const requestStudentGroup = async (): Promise<StudentGroupResponse> => {
  const body = await request<StudentGroupResponse | null>("/students/group");
  return {
    group: body?.group ?? null,
    students: Array.isArray(body?.students) ? body.students : [],
  };
};

export const api = {
  health: () => request<string>("/health"),
  login: (login: string, password: string) =>
    post<LoginResponse>("/auth/login", { login, password }),
  logout: () => post<MessageResponse>("/auth/logout"),
  session: () => request<SessionResponse>("/auth/session"),
  register: (payload: {
    name: string;
    email: string;
    password: string;
    group_id?: number;
  }) => post<{ message: string; id: number }>("/register", payload),

  studentGroup: requestStudentGroup,
  studentGrades: () => requestArray<GradeView>("/students/grades"),
  studentBalance: () => request<{ balance: string }>("/students/balance"),
  studentMerch: () => requestArray<Merch>("/students/merch"),
  studentPurchases: () => requestArray<Purchase>("/students/purchases"),
  studentAchievements: () => requestArray<Achievement>("/students/achievements"),
  createAchievement: (payload: { title: string; description?: string }) =>
    post<Achievement>("/students/achievements", payload),
  buyMerch: (merch_id: number) =>
    post<PurchaseResult>("/students/merch/buy", { merch_id }),

  teacherGroups: () => requestArray<Group>("/teachers/groups"),
  createGroup: (name: string) => post<Group>("/teachers/groups", { name }),
  attachGroup: (group_id: number) =>
    post<MessageResponse>("/teachers/groups/attach", { group_id }),
  createSubject: (name: string) =>
    post<Subject>("/teachers/subjects", { name }),
  attachSubject: (subject_id: number) =>
    post<MessageResponse>("/teachers/subjects/attach", { subject_id }),
  addStudentToGroup: (student_id: number, group_id: number) =>
    post<MessageResponse>("/teachers/groups/add-student", {
      student_id,
      group_id,
    }),
  setGrade: (payload: { student_id: number; subject_id: number; value: number }) =>
    post<Grade>("/teachers/grades", payload),
  teacherGroupGrades: (groupId: number) =>
    requestArray<GradeView>(`/teachers/groups/${groupId}/grades`),
  pendingAchievements: () =>
    requestArray<Achievement>("/teachers/achievements/pending"),
  confirmAchievement: (achievement_id: number) =>
    post<MessageResponse>("/teachers/achievements/confirm", { achievement_id }),
  denyAchievement: (achievement_id: number) =>
    post<MessageResponse>("/teachers/achievements/deny", { achievement_id }),
  awardTokens: (student_id: number, amount: number) =>
    post<MessageResponse>("/teachers/tokens/award", { student_id, amount }),
};
