
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
  StudentMeResponse,
  Subject,
  Student,
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

export function getReadableErrorMessage(error: unknown): string {
  if (error instanceof ApiError) {
    const message = error.message.toLowerCase();
    const status = error.status;
    
    if (status === 401 || message.includes('invalid credentials') || message.includes('unauthorized')) {
      return 'Неверный email или пароль. Проверьте правильность ввода.';
    }
    
    if (message.includes('fetch') || message.includes('network') || message.includes('failed to fetch')) {
      return 'Нет соединения с сервером. Проверьте интернет и попробуйте снова.';
    }
    
    if (message.includes('not found') && message.includes('user')) {
      return 'Пользователь с таким email не найден. Зарегистрируйтесь или проверьте email.';
    }
    
    if (message.includes('already exists') || message.includes('already taken')) {
      return 'Пользователь с таким email уже зарегистрирован. Используйте другой email или войдите.';
    }
    
    if (status === 400) {
      if (message.includes('password')) {
        return 'Пароль слишком простой. Используйте минимум 6 символов.';
      }
      if (message.includes('email')) {
        return 'Введите корректный email адрес.';
      }
      return 'Проверьте правильность заполнения формы.';
    }
    
    if (status === 403) {
      return 'Доступ запрещен. У вас недостаточно прав.';
    }
    
    if (status === 404) {
      return 'Запрашиваемые данные не найдены.';
    }
    
    if (status >= 500) {
      return 'Сервер временно недоступен. Попробуйте позже.';
    }
    
    if (error.message && error.message.length < 100 && !error.message.includes('stack')) {
      return error.message;
    }
    
    return 'Что-то пошло не так. Попробуйте обновить страницу.';
  }
  
  if (error instanceof Error) {
    if (error.message.includes('NetworkError') || error.message.includes('Failed to fetch')) {
      return 'Ошибка сети. Проверьте соединение с интернетом.';
    }
    return error.message;
  }
  
  return 'Произошла неизвестная ошибка. Попробуйте позже.';
}

// Старая функция для обратной совместимости
export const readableError = getReadableErrorMessage;

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
  
  let body;
  try {
    body = isJson ? await response.json() : await response.text();
  } catch {
    body = null;
  }

  if (!response.ok) {
    let errorMessage = "Запрос не выполнен";
    
    if (typeof body === "object" && body) {
      errorMessage = body.error || body.message || body.detail || String(body);
    } else if (typeof body === "string" && body) {
      errorMessage = body;
    }
    
    if (response.status === 401) {
      errorMessage = "Invalid credentials";
    } else if (response.status === 403) {
      errorMessage = "Доступ запрещен";
    } else if (response.status === 404 && errorMessage === "Запрос не выполнен") {
      errorMessage = "Ресурс не найден";
    }
    
    throw new ApiError(errorMessage, response.status);
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


  studentMe: () => request<StudentMeResponse>("/students/me"),
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
  teacherGroupStudents: (groupId: number) =>
    requestArray<Student>(`/teachers/groups/${groupId}/students`),
  pendingAchievements: () =>
    requestArray<Achievement>("/teachers/achievements/pending"),
  confirmAchievement: (achievement_id: number) =>
    post<MessageResponse>("/teachers/achievements/confirm", { achievement_id }),
  denyAchievement: (achievement_id: number) =>
    post<MessageResponse>("/teachers/achievements/deny", { achievement_id }),
  awardTokens: (student_id: number, amount: number) =>
    post<MessageResponse>("/teachers/tokens/award", { student_id, amount }),
};
  
 



