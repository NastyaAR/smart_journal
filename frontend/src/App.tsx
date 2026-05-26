import {
  BadgeCheck,
  BookOpen,
  Check,
  ClipboardCheck,
  Coins,
  Copy,
  GraduationCap,
  ListChecks,
  Loader2,
  LogIn,
  LogOut,
  Medal,
  Plus,
  RefreshCw,
  School,
  ShoppingBag,
  Sparkles,
  User,
  UserPlus,
  Users,
  Wallet,
  X,
} from "lucide-react";
import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import { api, ApiError, getReadableErrorMessage } from "./api";
import type {
  Achievement,
  GradeView,
  Group,
  LoginResponse,
  Merch,
  PendingAchievementView,
  Purchase,
  Role,
  SessionResponse,
  Student,
  StudentGroupResponse,
  StudentMeResponse,
  StoredRecommendation,
  Subject,
  TokenOperation,
} from "./types";

type StatusKind = "success" | "error" | "info";

interface Notice {
  kind: StatusKind;
  text: string;
}

type AuthMode = "teacher" | "student";
type StudentTab = "overview" | "wallet" | "grades" | "recommendations" | "achievements" | "market" | "purchases";
type TeacherTab = "overview" | "groups" | "grades" | "requests" | "tokens";

interface TokenEvent {
  id: string;
  title: string;
  detail: string;
  amount: number;
  kind: "income" | "expense";
  date?: string;
}

interface StudentOption {
  id: number;
  name: string;
}

interface SubjectOption {
  id: number;
  name: string;
}

const demoCredentials: Record<AuthMode, { login: string; password: string }> = {
  teacher: {
    login: "anna.ivanova@school.edu",
    password: "password",
  },
  student: {
    login: "student.demo@example.com",
    password: "password",
  },
};

const statusText: Record<Achievement["status"], string> = {
  pending: "На проверке",
  confirmed: "Подтверждено",
  denied: "Отклонено",
};

function readableError(error: unknown): string {
  if (error instanceof ApiError) {
    const msg = error.message.toLowerCase();
    
    if (error.status === 401 || msg.includes('invalid credentials')) 
      return 'Неверный email или пароль. Проверьте правильность ввода.';
    
    if (msg.includes('fetch') || msg.includes('network')) 
      return 'Нет соединения с сервером. Проверьте интернет.';
    
    if (msg.includes('already exists')) 
      return 'Пользователь с таким email уже существует.';
    
    if (error.status === 403) 
      return 'Доступ запрещен. Недостаточно прав.';
    
    if (error.status >= 500) 
      return 'Сервер временно недоступен. Попробуйте позже.';
    
    return error.message;
  }
  
  if (error instanceof Error) {
    if (error.message.includes('Failed to fetch')) 
      return 'Ошибка сети. Проверьте соединение.';
    return error.message;
  }
  
  return 'Что-то пошло не так. Попробуйте позже.';
}

function formatDate(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

function averageGrade(grades: GradeView[]): string {
  if (!grades.length) return "0.0";
  const total = grades.reduce((sum, grade) => sum + grade.value, 0);
  return (total / grades.length).toFixed(1);
}

function studentWalletAddress(studentId: number | null): string | null {
  if (!studentId) return null;
  return `0x${String(studentId).padStart(40, "0")}`;
}

function buildTokenEvents(achievements: Achievement[], purchases: Purchase[]): TokenEvent[] {
  const income = achievements
    .filter((item) => item.status === "confirmed")
    .map<TokenEvent>((item) => ({
      id: `achievement-${item.id}`,
      title: item.title,
      detail: "Подтвержденная активность",
      amount: 10,
      kind: "income",
    }));

  const expenses = purchases.map<TokenEvent>((item) => ({
    id: `purchase-${item.id}`,
    title: item.title,
    detail: `Покупка #${item.id}`,
    amount: -item.price,
    kind: "expense",
    date: item.created_at,
  }));

  return [...income, ...expenses];
}

const tokenOperationTitle: Record<TokenOperation["operation_type"], string> = {
  achievement_reward: "Награда за активность",
  manual_award: "Ручное начисление",
  purchase: "Покупка",
};

function buildTokenEventsFromOperations(operations: TokenOperation[]): TokenEvent[] {
  return operations.map<TokenEvent>((operation) => ({
    id: `operation-${operation.id}`,
    title: operation.reason || tokenOperationTitle[operation.operation_type],
    detail: [
      tokenOperationTitle[operation.operation_type],
      operation.teacher_name ? `Преподаватель: ${operation.teacher_name}` : "",
    ].filter(Boolean).join(" · "),
    amount: operation.amount,
    kind: operation.amount >= 0 ? "income" : "expense",
    date: operation.created_at,
  }));
}

function studentOptionsFromStudents(students: Student[]): StudentOption[] {
  return students
    .map((student) => ({ id: student.id, name: student.name || `Студент #${student.id}` }))
    .sort((left, right) => left.name.localeCompare(right.name, "ru"));
}

function knownStudentsFrom(grades: GradeView[], achievements: Achievement[]): StudentOption[] {
  const byId = new Map<number, string>();

  grades.forEach((grade) => {
    byId.set(grade.student_id, grade.student_name || `Студент #${grade.student_id}`);
  });

  achievements.forEach((achievement) => {
    if (!byId.has(achievement.student_id)) {
      byId.set(achievement.student_id, `Студент #${achievement.student_id}`);
    }
  });

  return Array.from(byId, ([id, name]) => ({ id, name })).sort((left, right) =>
    left.name.localeCompare(right.name, "ru"),
  );
}

function knownSubjectsFrom(grades: GradeView[], createdSubjects: Subject[]): SubjectOption[] {
  const byId = new Map<number, string>();

  createdSubjects.forEach((subject) => {
    byId.set(subject.id, subject.name);
  });

  grades.forEach((grade) => {
    byId.set(grade.subject_id, grade.subject_name || `Предмет #${grade.subject_id}`);
  });

  return Array.from(byId, ([id, name]) => ({ id, name })).sort((left, right) =>
    left.name.localeCompare(right.name, "ru"),
  );
}

function filterGrades(
  grades: GradeView[],
  studentFilter: string,
  subjectFilter: string,
): GradeView[] {
  return grades.filter((grade) => {
    const matchesStudent = studentFilter === "all" || grade.student_id === Number(studentFilter);
    const matchesSubject = subjectFilter === "all" || grade.subject_id === Number(subjectFilter);
    return matchesStudent && matchesSubject;
  });
}

function filterAchievements(achievements: Achievement[], query: string): Achievement[];
function filterAchievements(achievements: PendingAchievementView[], query: string): PendingAchievementView[];
function filterAchievements(achievements: (Achievement | PendingAchievementView)[], query: string): (Achievement | PendingAchievementView)[] {
  const value = query.trim().toLowerCase();
  if (!value) return achievements;

  return achievements.filter((achievement) => {
    const searchable = [
      String(achievement.student_id),
      achievement.title,
      achievement.description || "",
      "student_name" in achievement ? achievement.student_name : "",
      "group_name" in achievement ? achievement.group_name : "",
    ]
      .join(" ")
      .toLowerCase();

    return searchable.includes(value);
  });
}

function filterAchievementsByGroup(
  achievements: PendingAchievementView[],
  groupFilter: string,
): PendingAchievementView[] {
  if (groupFilter === "all") return achievements;
  return achievements.filter((achievement) => achievement.group_id === Number(groupFilter));
}

function useNotice() {
  const [notice, setNotice] = useState<Notice | null>(null);

  const showNotice = useCallback((kind: StatusKind, text: string) => {
    setNotice({ kind, text });
  }, []);

  const clearNotice = useCallback(() => setNotice(null), []);

  return { notice, showNotice, clearNotice };
}

function App() {
  const [session, setSession] = useState<SessionResponse | null>(null);
  const [checkingSession, setCheckingSession] = useState(true);
  const { notice, showNotice, clearNotice } = useNotice();

  const refreshSession = useCallback(async () => {
    setCheckingSession(true);
    try {
      const activeSession = await api.session();
      setSession(activeSession);
    } catch {
      setSession(null);
    } finally {
      setCheckingSession(false);
    }
  }, []);

  useEffect(() => {
    refreshSession();
  }, [refreshSession]);

  const handleLogin = (response: LoginResponse) => {
    setSession({ user_id: 0, role: response.role });
    refreshSession();
  };

  const handleLogout = async () => {
    try {
      await api.logout();
      setSession(null);
      showNotice("success", "Вы вышли из аккаунта");
    } catch (error) {
      showNotice("error", getReadableErrorMessage(error));
    }
  };

  return (
    <main className="app-shell">
      <div className="ambient ambient-one" />
      <div className="ambient ambient-two" />

      <header className="topbar" aria-label="Главная навигация">
        <div className="brand">
          <div className="brand-mark">
            <GraduationCap size={24} aria-hidden="true" />
          </div>
          <div>
            <p>Smart Journal</p>
            <span>учебный кабинет AMT</span>
          </div>
        </div>

        {session && (
          <button
            className="ghost-button"
            type="button"
            title="Выйти из аккаунта"
            onClick={handleLogout}
          >
            <LogOut size={18} aria-hidden="true" />
            Выйти
          </button>
        )}
      </header>

      {notice && (
        <div className={`notice ${notice.kind}`} role="status">
          <span>{notice.text}</span>
          <button
            type="button"
            className="icon-button"
            title="Закрыть уведомление"
            onClick={clearNotice}
          >
            <X size={16} aria-hidden="true" />
          </button>
        </div>
      )}

      {checkingSession ? (
        <CenteredState icon={<Loader2 className="spin" size={28} />} title="Проверяем сессию" />
      ) : session ? (
        <Workspace
          session={session}
          role={session.role}
          onNotice={showNotice}
          onLogout={handleLogout}
        />
      ) : (
        <AuthPanel onLogin={handleLogin} onNotice={showNotice} />
      )}
    </main>
  );
}

function Workspace({
  session,
  role,
  onNotice,
  onLogout,
}: {
  session: SessionResponse;
  role: Role;
  onNotice: (kind: StatusKind, text: string) => void;
  onLogout: () => void;
}) {
  return (
    <section className="workspace">
      {role === "student" ? (
        <StudentWorkspace session={session} onNotice={onNotice} />
      ) : (
        <TeacherWorkspace onNotice={onNotice} />
      )}
      <button
        className="mobile-logout"
        type="button"
        title="Выйти из аккаунта"
        onClick={onLogout}
      >
        <LogOut size={18} aria-hidden="true" />
        Выйти
      </button>
    </section>
  );
}

function AuthPanel({
  onLogin,
  onNotice,
}: {
  onLogin: (response: LoginResponse) => void;
  onNotice: (kind: StatusKind, text: string) => void;
}) {
  const [mode, setMode] = useState<AuthMode>("teacher");
  const [login, setLogin] = useState(demoCredentials.teacher.login);
  const [password, setPassword] = useState(demoCredentials.teacher.password);
  const [name, setName] = useState("Иван Петров");
  const [email, setEmail] = useState("ivan.petrov@example.com");
  const [studentPassword, setStudentPassword] = useState("secret");
  const [groupId, setGroupId] = useState("");
  const [loading, setLoading] = useState(false);
  const [registeredId, setRegisteredId] = useState<number | null>(null);
  const [isFormVisible, setIsFormVisible] = useState(false);
  const [selectedCardType, setSelectedCardType] = useState<"teacher" | "student" | "register" | null>(null);
  
  // Состояния для ошибок
  const [loginError, setLoginError] = useState<string | null>(null);
  const [registerError, setRegisterError] = useState<string | null>(null);
  const [emailFocused, setEmailFocused] = useState(false);
  const [passwordFocused, setPasswordFocused] = useState(false);

  const chooseMode = (nextMode: AuthMode) => {
    setMode(nextMode);
    setLogin(demoCredentials[nextMode].login);
    setPassword(demoCredentials[nextMode].password);
    setLoginError(null); // Сбрасываем ошибку при смене режима
  };

  const submitLogin = async (event: FormEvent) => {
    event.preventDefault();
    setLoading(true);
    setLoginError(null);
    
    // Простая валидация на фронте
    if (!login.trim()) {
      setLoginError("Введите email");
      setLoading(false);
      return;
    }
    if (!password) {
      setLoginError("Введите пароль");
      setLoading(false);
      return;
    }
    
    try {
      const response = await api.login(login, password);
      onNotice("success", response.role === "teacher" ? "Вход выполнен" : "Добро пожаловать");
      onLogin(response);
    } catch (error) {
      const errorMessage = getReadableErrorMessage(error);
      setLoginError(errorMessage);
      // Не показываем глобальное уведомление, т.к. ошибка уже видна в форме
    } finally {
      setLoading(false);
    }
  };

  const submitRegister = async (event: FormEvent) => {
    event.preventDefault();
    setLoading(true);
    setRegisterError(null);
    
    if (!name.trim()) {
      setRegisterError("Введите имя");
      setLoading(false);
      return;
    }
    if (!email.trim() || !email.includes('@')) {
      setRegisterError("Введите корректный email");
      setLoading(false);
      return;
    }
    if (studentPassword.length < 4) {
      setRegisterError("Пароль должен быть не менее 4 символов");
      setLoading(false);
      return;
    }
    
    try {
      const response = await api.register({
        name,
        email,
        password: studentPassword,
        group_id: groupId.trim() ? Number(groupId) : 0,
      });
      setRegisteredId(response.id);
      setLogin(email);
      setPassword(studentPassword);
      setMode("student");
      onNotice("success", `Ученик зарегистрирован. ID: ${response.id}`);
    } catch (error) {
      setRegisterError(getReadableErrorMessage(error));
    } finally {
      setLoading(false);
    }
  };

  const copyRegisteredId = async () => {
    if (!registeredId) return;
    try {
      await navigator.clipboard.writeText(String(registeredId));
      onNotice("success", "ID скопирован");
    } catch {
      onNotice("error", "Не удалось скопировать");
    }
  };

  const handleCardClick = (type: "teacher" | "student" | "register") => {
    if (type === "teacher") chooseMode("teacher");
    if (type === "student") chooseMode("student");
    setSelectedCardType(type);
    setIsFormVisible(true);
    setLoginError(null);
    setRegisterError(null);
  };

  const handleBack = () => {
    setIsFormVisible(false);
    setSelectedCardType(null);
    setLoginError(null);
    setRegisterError(null);
  };

  return (
    <div className="new-auth-wrapper">
      <div className="new-auth-content">
        <div className="new-auth-hero">
          <h1 className="new-auth-title">
            Журнал, где достижения<br />
            превращаются в токены
          </h1>
          <p className="new-auth-description">
            Учитель ведет группы, предметы, оценки и подтверждает активности.<br />
            Ученик видит прогресс, копит AMT и тратит их на мерч.
          </p>
          <div className={`new-auth-stats ${isFormVisible ? 'stats-hidden' : ''}`}>
            <div className="new-stat-card"><div className="new-stat-value">2</div><div className="new-stat-label">Роли</div><div className="new-stat-caption">учитель и ученик</div></div>
            <div className="new-stat-card"><div className="new-stat-value">10</div><div className="new-stat-label">AMT</div><div className="new-stat-caption">за подтверждение</div></div>
            <div className="new-stat-card"><div className="new-stat-value">7</div><div className="new-stat-label">Товаров</div><div className="new-stat-caption">в магазине</div></div>
          </div>
        </div>

        <div className="new-auth-cards-container">
          <div className={`auth-cards-section ${isFormVisible ? 'form-visible' : ''}`}>
            <div className="new-auth-cards">
              <button className="new-auth-card teacher" type="button" onClick={() => handleCardClick("teacher")}>
                <div className="new-auth-card-icon"><ClipboardCheck size={32} /></div>
                <h3>Учитель</h3><p>Вход в кабинет учителя</p><span className="new-auth-card-arrow">→</span>
              </button>
              <button className="new-auth-card student" type="button" onClick={() => handleCardClick("student")}>
                <div className="new-auth-card-icon"><User size={32} /></div>
                <h3>Ученик</h3><p>Вход в кабинет ученика</p><span className="new-auth-card-arrow">→</span>
              </button>
              <button className="new-auth-card register" type="button" onClick={() => handleCardClick("register")}>
                <div className="new-auth-card-icon"><UserPlus size={32} /></div>
                <h3>Регистрация</h3><p>Создать аккаунт ученика</p><span className="new-auth-card-arrow">→</span>
              </button>
            </div>

            <div className={`new-auth-form-wrapper ${isFormVisible ? 'visible' : ''}`}>
              <button className="new-auth-back-btn" type="button" onClick={handleBack}>← Назад к выбору</button>

              {/* Форма входа */}
              {selectedCardType !== "register" && (
                <form className="new-auth-form" onSubmit={submitLogin} noValidate>
                  <div className="new-auth-form-header">
                    <div className="new-auth-form-icon">{mode === "teacher" ? <ClipboardCheck size={24} /> : <User size={24} />}</div>
                    <div><p>Вход в систему</p><h2>{mode === "teacher" ? "Кабинет учителя" : "Кабинет ученика"}</h2></div>
                  </div>

                  <div className={`new-auth-form-group ${loginError && !emailFocused ? 'error' : ''}`}>
                    <label>Email</label>
                    <input
                      type="email"
                      value={login}
                      onChange={(e) => { setLogin(e.target.value); setLoginError(null); }}
                      onFocus={() => setEmailFocused(true)}
                      onBlur={() => setEmailFocused(false)}
                      placeholder="example@school.edu"
                      autoComplete="email"
                      required
                    />
                  </div>

                  <div className={`new-auth-form-group ${loginError && !passwordFocused ? 'error' : ''}`}>
                    <label>Пароль</label>
                    <input
                      type="password"
                      value={password}
                      onChange={(e) => { setPassword(e.target.value); setLoginError(null); }}
                      onFocus={() => setPasswordFocused(true)}
                      onBlur={() => setPasswordFocused(false)}
                      placeholder="••••••••"
                      autoComplete="current-password"
                      required
                    />
                  </div>

                  {/* Ошибка входа — прямо над кнопкой */}
                  {loginError && (
                    <div className="auth-error-message" role="alert">
                      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                        <circle cx="12" cy="12" r="10"/>
                        <line x1="12" y1="8" x2="12" y2="12"/>
                        <line x1="12" y1="16" x2="12.01" y2="16"/>
                      </svg>
                      <span>{loginError}</span>
                    </div>
                  )}

                  <button className="new-auth-submit-btn" type="submit" disabled={loading}>
                    {loading ? <Loader2 className="spin" size={18} /> : <LogIn size={18} />}
                    Войти
                  </button>

                  <div className="auth-hint">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                      <circle cx="12" cy="12" r="10"/>
                      <path d="M12 16v-4M12 8h.01"/>
                    </svg>
                    <span>
                      Демо-доступ: <strong>{demoCredentials[mode].login}</strong> /{" "}
                      <strong>{demoCredentials[mode].password}</strong>
                    </span>
                  </div>
                </form>
              )}

              {/* Форма регистрации */}
              {selectedCardType === "register" && (
                <form className="new-auth-form" onSubmit={submitRegister} noValidate>
                  <div className="new-auth-form-header">
                    <div className="new-auth-form-icon"><UserPlus size={24} /></div>
                    <div><p>Добро пожаловать</p><h2>Регистрация ученика</h2></div>
                  </div>

                  <div className="new-auth-form-group">
                    <label>Имя</label>
                    <input value={name} onChange={(e) => setName(e.target.value)} placeholder="Иван Петров" required />
                  </div>

                  <div className="new-auth-form-group">
                    <label>Email</label>
                    <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} placeholder="ivan@example.com" required />
                  </div>

                  <div className="new-auth-form-row">
                    <div className="new-auth-form-group">
                      <label>ID группы (опционально)</label>
                      <input type="number" min="0" value={groupId} placeholder="0" onChange={(e) => setGroupId(e.target.value)} />
                    </div>
                    <div className="new-auth-form-group">
                      <label>Пароль</label>
                      <input type="password" value={studentPassword} onChange={(e) => setStudentPassword(e.target.value)} placeholder="минимум 4 символа" required />
                    </div>
                  </div>

                  {registerError && (
                    <div className="auth-error-message" role="alert">
                      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                        <circle cx="12" cy="12" r="10"/>
                        <line x1="12" y1="8" x2="12" y2="12"/>
                        <line x1="12" y1="16" x2="12.01" y2="16"/>
                      </svg>
                      <span>{registerError}</span>
                    </div>
                  )}

                  <button className="new-auth-submit-btn" type="submit" disabled={loading}>
                    {loading ? <Loader2 className="spin" size={18} /> : <UserPlus size={18} />}
                    Зарегистрироваться
                  </button>

                  {registeredId && (
                    <div className="auth-success-message">
                      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                        <path d="M20 6L9 17l-5-5"/>
                      </svg>
                      <span>Ученик зарегистрирован! ID: {registeredId}</span>
                      <button type="button" onClick={copyRegisteredId} className="copy-id-btn">
                        <Copy size={12} /> Копировать
                      </button>
                    </div>
                  )}
                </form>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function StudentWorkspace({
  session,
  onNotice,
}: {
  session: SessionResponse;
  onNotice: (kind: StatusKind, text: string) => void;
}) {
  const [tab, setTab] = useState<StudentTab>("overview");
  const [profile, setProfile] = useState<StudentMeResponse | null>(null);
  const [groupData, setGroupData] = useState<StudentGroupResponse | null>(null);
  const [grades, setGrades] = useState<GradeView[]>([]);
  const [balance, setBalance] = useState(0);
  const [merch, setMerch] = useState<Merch[]>([]);
  const [purchases, setPurchases] = useState<Purchase[]>([]);
  const [achievements, setAchievements] = useState<Achievement[]>([]);
  const [tokenOperations, setTokenOperations] = useState<TokenOperation[]>([]);
  const [recommendation, setRecommendation] = useState<StoredRecommendation | null>(null);
  const [recommendationLoading, setRecommendationLoading] = useState(false);
  const [recommendationError, setRecommendationError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [achievementTitle, setAchievementTitle] = useState("");
  const [achievementDescription, setAchievementDescription] = useState("");
  const [studentSubjectFilter, setStudentSubjectFilter] = useState("all");

  const loadLatestRecommendation = useCallback(async (showNotice = false) => {
    try {
      const latest = await api.latestRecommendation();
      setRecommendation(latest);
      setRecommendationError(null);
    } catch (error) {
      setRecommendation(null);
      if (error instanceof ApiError && error.status === 404) {
        setRecommendationError(null);
        return;
      }

      const message = getReadableErrorMessage(error);
      setRecommendationError(message);
      if (showNotice) {
        onNotice("error", message);
      }
    }
  }, [onNotice]);

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const [studentProfile, group, gradeList, balanceValue, merchList, purchaseList, achievementList, tokenHistory] =
        await Promise.all([
          api.studentMe(),
          api.studentGroup(),
          api.studentGrades(),
          api.studentBalance(),
          api.studentMerch(),
          api.studentPurchases(),
          api.studentAchievements(),
          api.studentTokenOperations(),
        ]);
      setProfile(studentProfile);
      setGroupData(group);
      setGrades(gradeList);
      setBalance(Number(balanceValue.balance || 0));
      setMerch(merchList);
      setPurchases(purchaseList);
      setAchievements(achievementList);
      setTokenOperations(tokenHistory);
      await loadLatestRecommendation(false);
    } catch (error) {
      onNotice("error", getReadableErrorMessage(error));
    } finally {
      setLoading(false);
    }
  }, [loadLatestRecommendation, onNotice]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const submitAchievement = async (event: FormEvent) => {
    event.preventDefault();
    setSaving(true);
    try {
      await api.createAchievement({
        title: achievementTitle,
        description: achievementDescription,
      });
      setAchievementTitle("");
      setAchievementDescription("");
      onNotice("success", "Заявка отправлена учителю");
      await loadData();
    } catch (error) {
      onNotice("error", getReadableErrorMessage(error));
    } finally {
      setSaving(false);
    }
  };

  const buyMerch = async (item: Merch) => {
    const confirmed = window.confirm(`Купить "${item.title}" за ${item.price} AMT?`);
    if (!confirmed) return;

    setSaving(true);
    try {
      const result = await api.buyMerch(item.id);
      setBalance(result.new_balance);
      onNotice("success", `Покупка оформлена: ${item.title}`);
      await loadData();
    } catch (error) {
      onNotice("error", getReadableErrorMessage(error));
    } finally {
      setSaving(false);
    }
  };

  const generateRecommendation = async () => {
    setRecommendationLoading(true);
    setRecommendationError(null);
    try {
      const result = await api.generateRecommendation();
      setRecommendation(result);
      onNotice("success", "AI-рекомендации обновлены");
    } catch (error) {
      let message = getReadableErrorMessage(error);
      if (error instanceof ApiError && error.status === 400) {
        message = "Для рекомендаций нужны оценки в журнале. Попросите преподавателя добавить хотя бы одну оценку.";
      }
      if (error instanceof ApiError && error.status === 502) {
        message = "Сервис рекомендаций временно недоступен.";
      }
      setRecommendationError(message);
      onNotice("error", message);
    } finally {
      setRecommendationLoading(false);
    }
  };

  const pendingCount = achievements.filter((item) => item.status === "pending").length;
  const confirmedCount = achievements.filter((item) => item.status === "confirmed").length;
  const currentStudentId = profile?.student_id || null;
  const currentStudent = profile
    ? {
        id: profile.student_id,
        name: profile.name,
        email: profile.email,
        group_id: profile.group_id,
        tokens: profile.tokens,
        user_id: session.user_id,
      }
    : null;
  const ownGrades = currentStudentId
    ? grades.filter((grade) => grade.student_id === currentStudentId)
    : grades;
  const classmates = groupData?.students || [];
  const walletAddress = profile?.wallet_address || studentWalletAddress(currentStudentId);
  const tokenEvents = tokenOperations.length
    ? buildTokenEventsFromOperations(tokenOperations)
    : buildTokenEvents(achievements, purchases);
  const subjects = Array.from(new Set(ownGrades.map((grade) => grade.subject_name))).filter(Boolean);
  const subjectOptions = knownSubjectsFrom(ownGrades, []);
  const filteredOwnGrades = filterGrades(ownGrades, "all", studentSubjectFilter);

  return (
    <div className="role-layout">
      <RoleHero
        icon={<Wallet size={28} />}
        eyebrow="Кабинет ученика"
        title="Мой прогресс и AMT"
        text="Оценки, активности, баланс токенов и покупки в одном рабочем пространстве."
        actions={
          <button className="ghost-button light" type="button" title="Обновить данные" onClick={loadData}>
            <RefreshCw size={18} aria-hidden="true" />
            Обновить
          </button>
        }
      />

      {!loading && groupData?.group && (
        <TabBar<StudentTab>
          active={tab}
          onChange={setTab}
          items={[
            { id: "overview", label: "Обзор", icon: <School size={18} /> },
            { id: "wallet", label: "Кошелек", icon: <Wallet size={18} /> },
            { id: "grades", label: "Оценки", icon: <BookOpen size={18} /> },
            { id: "recommendations", label: "Рекомендации", icon: <Sparkles size={18} /> },
            { id: "achievements", label: "Активности", icon: <Medal size={18} /> },
            { id: "market", label: "Маркет", icon: <ShoppingBag size={18} /> },
            { id: "purchases", label: "Покупки", icon: <ListChecks size={18} /> },
          ]}
        />
      )}

      {loading ? (
        <CenteredState icon={<Loader2 className="spin" size={28} />} title="Загружаем кабинет" />
      ) : !groupData?.group ? (
        <PendingGroupState
          balance={balance}
          pendingCount={pendingCount}
          achievementCount={achievements.length}
          onRefresh={loadData}
        />
      ) : (
        <>
          {tab === "overview" && (
            <div className="content-grid">
              <MetricCard label="Баланс" value={`${balance}`} caption="AMT" accent="green" />
              <MetricCard label="Средний балл" value={averageGrade(ownGrades)} caption={`${ownGrades.length} моих оценок`} />
              <MetricCard label="Заявки" value={`${pendingCount}`} caption="на проверке" accent="amber" />
              <MetricCard label="Курсы" value={`${subjects.length}`} caption={`${confirmedCount} активностей`} accent="blue" />

              <section className="surface wide">
                <div className="section-heading">
                  <div>
                    <p>{currentStudent?.name || "Профиль ученика"}</p>
                    <h2>{groupData?.group?.name || "Группа не назначена"}</h2>
                  </div>
                  <Users size={20} aria-hidden="true" />
                </div>
                <PeopleList students={classmates} currentStudentId={currentStudentId} />
              </section>

              <section className="surface">
                <div className="section-heading">
                  <div>
                    <p>AMT</p>
                    <h2>Кошелек</h2>
                  </div>
                  <Wallet size={20} aria-hidden="true" />
                </div>
                <WalletSummary
                  balance={balance}
                  walletAddress={walletAddress}
                  tokenEvents={tokenEvents.slice(0, 4)}
                />
              </section>
            </div>
          )}

          {tab === "wallet" && (
            <section className="surface readable">
              <div className="section-heading">
                <div>
                  <p>AMT</p>
                  <h2>Блокчейн-кошелек</h2>
                </div>
                <Wallet size={20} aria-hidden="true" />
              </div>
              <WalletSummary
                balance={balance}
                walletAddress={walletAddress}
                tokenEvents={tokenEvents}
              />
            </section>
          )}

          {tab === "grades" && (
            <section className="surface">
              <div className="section-heading">
                <div>
                  <p>Личный журнал</p>
                  <h2>Мой журнал успеваемости</h2>
                </div>
                <BookOpen size={20} aria-hidden="true" />
              </div>
              <div className="toolbar filter-toolbar">
                <select
                  value={studentSubjectFilter}
                  onChange={(event) => setStudentSubjectFilter(event.target.value)}
                >
                  <option value="all">Все предметы</option>
                  {subjectOptions.map((subject) => (
                    <option key={subject.id} value={subject.id}>
                      {subject.name}
                    </option>
                  ))}
                </select>
              </div>
              <GradeSummary grades={filteredOwnGrades} mode="subjects" />
              <GradeTable grades={filteredOwnGrades} showStudent={false} />
            </section>
          )}

          {tab === "recommendations" && (
            <RecommendationPanel
              grades={ownGrades}
              recommendation={recommendation}
              loading={recommendationLoading}
              error={recommendationError}
              onGenerate={generateRecommendation}
            />
          )}

          {tab === "achievements" && (
            <div className="two-column">
              <form className="surface form-surface" onSubmit={submitAchievement}>
                <div className="section-heading">
                  <div>
                    <p>Новая заявка</p>
                    <h2>Активность</h2>
                  </div>
                  <Plus size={20} aria-hidden="true" />
                </div>
                <label>
                  Название
                  <input
                    value={achievementTitle}
                    onChange={(event) => setAchievementTitle(event.target.value)}
                    placeholder="Участие в олимпиаде"
                    required
                  />
                </label>
                <label>
                  Описание
                  <textarea
                    value={achievementDescription}
                    onChange={(event) => setAchievementDescription(event.target.value)}
                    placeholder="Коротко опишите результат"
                    rows={5}
                  />
                </label>
                <button className="primary-button" type="submit" title="Отправить заявку" disabled={saving}>
                  {saving ? <Loader2 className="spin" size={18} /> : <Plus size={18} />}
                  Отправить
                </button>
              </form>

              <section className="surface">
                <div className="section-heading">
                  <div>
                    <p>История</p>
                    <h2>Мои активности</h2>
                  </div>
                  <Medal size={20} aria-hidden="true" />
                </div>
                <AchievementList achievements={achievements} />
              </section>
            </div>
          )}

          {tab === "market" && (
            <section className="surface">
              <div className="section-heading">
                <div>
                  <p>Баланс: {balance} AMT</p>
                  <h2>Маркет мерча</h2>
                </div>
                <ShoppingBag size={20} aria-hidden="true" />
              </div>
              {merch.length ? (
                <div className="item-grid">
                  {merch.map((item) => (
                    <article className="item-card" key={item.id}>
                      <div>
                        <span className={`item-status ${balance >= item.price ? "available" : "locked"}`}>
                          {balance >= item.price ? "Доступно" : `Не хватает ${item.price - balance} AMT`}
                        </span>
                        <h3>{item.title}</h3>
                        <p>{item.description || "Описание появится позже"}</p>
                      </div>
                      <div className="item-footer">
                        <strong>{item.price} AMT</strong>
                        <button
                          className="secondary-button small"
                          type="button"
                          title="Купить мерч"
                          disabled={saving || balance < item.price}
                          onClick={() => buyMerch(item)}
                        >
                          <ShoppingBag size={14} aria-hidden="true" />
                          Купить
                        </button>
                      </div>
                    </article>
                  ))}
                </div>
              ) : (
                <EmptyState title="Мерч пока не добавлен" text="Товары появятся здесь после заполнения каталога." />
              )}
            </section>
          )}

          {tab === "purchases" && (
            <section className="surface">
              <div className="section-heading">
                <div>
                  <p>История</p>
                  <h2>Покупки мерча</h2>
                </div>
                <ListChecks size={20} aria-hidden="true" />
              </div>
              <PurchasesTable purchases={purchases} />
            </section>
          )}
        </>
      )}
    </div>
  );
}

function TeacherWorkspace({ onNotice }: { onNotice: (kind: StatusKind, text: string) => void }) {
  const [tab, setTab] = useState<TeacherTab>("overview");
  const [groups, setGroups] = useState<Group[]>([]);
  const [pendingAchievements, setPendingAchievements] = useState<PendingAchievementView[]>([]);
  const [selectedGroupId, setSelectedGroupId] = useState<number | null>(null);
  const [groupGrades, setGroupGrades] = useState<GradeView[]>([]);
  const [groupStudents, setGroupStudents] = useState<Student[]>([]);
  const [tokenOperations, setTokenOperations] = useState<TokenOperation[]>([]);
  const [createdSubjects, setCreatedSubjects] = useState<Subject[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  const [newGroup, setNewGroup] = useState("БПИ-231");
  const [attachGroupId, setAttachGroupId] = useState("");
  const [newSubject, setNewSubject] = useState("Математика");
  const [attachSubjectId, setAttachSubjectId] = useState("");
  const [studentId, setStudentId] = useState("");
  const [studentGroupId, setStudentGroupId] = useState("");
  const [gradeStudentId, setGradeStudentId] = useState("");
  const [gradeSubjectId, setGradeSubjectId] = useState("");
  const [gradeValue, setGradeValue] = useState("5");
  const [gradeLessonDate, setGradeLessonDate] = useState(() => new Date().toISOString().slice(0, 10));
  const [awardStudentId, setAwardStudentId] = useState("");
  const [awardAmount, setAwardAmount] = useState("10");
  const [gradeStudentFilter, setGradeStudentFilter] = useState("all");
  const [gradeSubjectFilter, setGradeSubjectFilter] = useState("all");
  const [requestGroupFilter, setRequestGroupFilter] = useState("all");
  const [requestSearch, setRequestSearch] = useState("");

  const loadTeacherData = useCallback(async () => {
    setLoading(true);
    try {
      const [groupList, achievements, subjectList] = await Promise.all([
        api.teacherGroups(),
        api.pendingAchievements(),
        api.teacherSubjects(),
      ]);
      setGroups(groupList);
      setPendingAchievements(achievements);
      setCreatedSubjects(subjectList);
      const nextGroup = selectedGroupId || groupList[0]?.id || null;
      setSelectedGroupId(nextGroup);
      if (nextGroup) {
        const [grades, students, operations] = await Promise.all([
          api.teacherGroupGrades(nextGroup),
          api.teacherGroupStudents(nextGroup),
          api.teacherGroupTokenOperations(nextGroup),
        ]);
        setGroupGrades(grades);
        setGroupStudents(students);
        setTokenOperations(operations);
      } else {
        setGroupGrades([]);
        setGroupStudents([]);
        setTokenOperations([]);
      }
    } catch (error) {
      onNotice("error", getReadableErrorMessage(error));
    } finally {
      setLoading(false);
    }
  }, [onNotice, selectedGroupId]);

  useEffect(() => {
    loadTeacherData();
  }, [loadTeacherData]);

  useEffect(() => {
    if (selectedGroupId && !studentGroupId) {
      setStudentGroupId(String(selectedGroupId));
    }
  }, [selectedGroupId, studentGroupId]);

  const loadGradesForGroup = async (groupId: number) => {
    setSelectedGroupId(groupId);
    setStudentGroupId(String(groupId));
    setGradeStudentFilter("all");
    setGradeSubjectFilter("all");
    setAwardStudentId("");
    try {
      const [grades, students, operations] = await Promise.all([
        api.teacherGroupGrades(groupId),
        api.teacherGroupStudents(groupId),
        api.teacherGroupTokenOperations(groupId),
      ]);
      setGroupGrades(grades);
      setGroupStudents(students);
      setTokenOperations(operations);
    } catch (error) {
      onNotice("error", getReadableErrorMessage(error));
    }
  };

  const handleCreateGroup = async (event: FormEvent) => {
    event.preventDefault();
    await runTeacherAction(async () => {
      const group = await api.createGroup(newGroup);
      onNotice("success", `Группа создана: ${group.name}`);
      setNewGroup("");
      await loadTeacherData();
    });
  };

  const handleAttachGroup = async (event: FormEvent) => {
    event.preventDefault();
    await runTeacherAction(async () => {
      await api.attachGroup(Number(attachGroupId));
      onNotice("success", `Группа ${attachGroupId} привязана`);
      setAttachGroupId("");
      await loadTeacherData();
    });
  };

  const handleCreateSubject = async (event: FormEvent) => {
    event.preventDefault();
    await runTeacherAction(async () => {
      await api.createSubject(newSubject);
      onNotice("success", `Предмет создан: ${newSubject}`);
      setNewSubject("");
      await loadTeacherData();
    });
  };

  const handleAttachSubject = async (event: FormEvent) => {
    event.preventDefault();
    await runTeacherAction(async () => {
      await api.attachSubject(Number(attachSubjectId));
      onNotice("success", `Предмет ${attachSubjectId} привязан`);
      setAttachSubjectId("");
      await loadTeacherData();
    });
  };

  const handleAddStudent = async (event: FormEvent) => {
    event.preventDefault();
    await runTeacherAction(async () => {
      await api.addStudentToGroup(Number(studentId), Number(studentGroupId));
      onNotice("success", `Ученик ${studentId} добавлен в группу`);
      setStudentId("");
      await loadTeacherData();
    });
  };

  const handleSetGrade = async (event: FormEvent) => {
    event.preventDefault();
    await runTeacherAction(async () => {
      await api.setGrade({
        student_id: Number(gradeStudentId),
        subject_id: Number(gradeSubjectId),
        value: Number(gradeValue),
        lesson_date: gradeLessonDate,
      });
      onNotice("success", "Оценка добавлена");
      if (selectedGroupId) await loadGradesForGroup(selectedGroupId);
    });
  };

  const handleAwardTokens = async (event: FormEvent) => {
    event.preventDefault();
    const amount = Number(awardAmount);
    const studentID = Number(awardStudentId);
    const studentName =
      knownStudents.find((student) => student.id === studentID)?.name || `ID ${studentID}`;
    const confirmed = window.confirm(
      `Начислить ${amount} AMT студенту ${studentName}? Это действие сразу изменит баланс.`,
    );
    if (!confirmed) return;

    await runTeacherAction(async () => {
      await api.awardTokens(studentID, amount);
      onNotice("success", `${amount} AMT начислено ученику ${awardStudentId}`);
      setAwardStudentId("");
      if (selectedGroupId) await loadGradesForGroup(selectedGroupId);
    });
  };

  const handleAchievementAction = async (achievementId: number, action: "confirm" | "deny") => {
    await runTeacherAction(async () => {
      if (action === "confirm") {
        await api.confirmAchievement(achievementId);
        onNotice("success", "Активность подтверждена, AMT начислены");
      } else {
        await api.denyAchievement(achievementId);
        onNotice("success", "Активность отклонена");
      }
      await loadTeacherData();
    });
  };

  const runTeacherAction = async (action: () => Promise<void>) => {
    setSaving(true);
    try {
      await action();
    } catch (error) {
      onNotice("error", getReadableErrorMessage(error));
    } finally {
      setSaving(false);
    }
  };

  const selectedGroup = useMemo(
    () => groups.find((group) => group.id === selectedGroupId),
    [groups, selectedGroupId],
  );
  const knownStudents = useMemo(
    () => studentOptionsFromStudents(groupStudents),
    [groupStudents],
  );
  const knownSubjects = useMemo(
    () => knownSubjectsFrom(groupGrades, createdSubjects),
    [groupGrades, createdSubjects],
  );
  const journalSubjects = useMemo(
    () => knownSubjectsFrom(groupGrades, []),
    [groupGrades],
  );
  const filteredGroupGrades = useMemo(
    () => filterGrades(groupGrades, gradeStudentFilter, gradeSubjectFilter),
    [groupGrades, gradeStudentFilter, gradeSubjectFilter],
  );
  const filteredPendingAchievements = useMemo(
    () => filterAchievements(filterAchievementsByGroup(pendingAchievements, requestGroupFilter), requestSearch),
    [pendingAchievements, requestGroupFilter, requestSearch],
  );

  return (
    <div className="role-layout">
      <RoleHero
        icon={<ClipboardCheck size={28} />}
        eyebrow="Кабинет учителя"
        title="Группы, оценки и AMT-награды"
        text="Рабочее место для ведения учебного журнала и подтверждения достижений студентов."
        actions={
          <button
            className="ghost-button light"
            type="button"
            title="Обновить данные"
            onClick={loadTeacherData}
          >
            <RefreshCw size={18} aria-hidden="true" />
            Обновить
          </button>
        }
      />

      <TabBar<TeacherTab>
        active={tab}
        onChange={setTab}
        items={[
          { id: "overview", label: "Обзор", icon: <School size={18} /> },
          { id: "groups", label: "Группы", icon: <Users size={18} /> },
          { id: "grades", label: "Оценки", icon: <BookOpen size={18} /> },
          { id: "requests", label: "Заявки", icon: <Medal size={18} /> },
          { id: "tokens", label: "Токены", icon: <Coins size={18} /> },
        ]}
      />

      {loading ? (
        <CenteredState icon={<Loader2 className="spin" size={28} />} title="Загружаем кабинет" />
      ) : (
        <>
          {tab === "overview" && (
            <div className="content-grid">
              <div className="grid-caption full-row">Общее по кабинету</div>
              <MetricCard label="Мои группы" value={`${groups.length}`} caption="доступны" />
              <MetricCard label="Заявки" value={`${pendingAchievements.length}`} caption="ожидают решения" accent="amber" />
              <div className="grid-caption full-row">Выбранная группа</div>
              <MetricCard label="Оценки группы" value={`${groupGrades.length}`} caption={selectedGroup?.name || "нет группы"} accent="blue" />
              <MetricCard label="Средний балл" value={averageGrade(groupGrades)} caption="по выбранной группе" accent="green" />
              <MetricCard label="Студенты" value={`${knownStudents.length}`} caption="есть в журнале" />
              <MetricCard label="Предметы" value={`${journalSubjects.length}`} caption="есть в журнале" accent="blue" />

              <section className="surface wide">
                <div className="section-heading">
                  <div>
                    <p>Группы</p>
                    <h2>Мои учебные потоки</h2>
                  </div>
                  <Users size={20} aria-hidden="true" />
                </div>
                <GroupList
                  groups={groups}
                  selectedGroupId={selectedGroupId}
                  onSelect={loadGradesForGroup}
                />
              </section>

              <section className="surface wide">
                <div className="section-heading">
                  <div>
                    <p>{selectedGroup?.name || "Группа не выбрана"}</p>
                    <h2>Последние оценки</h2>
                  </div>
                  <BookOpen size={20} aria-hidden="true" />
                </div>
                <GradeTable grades={groupGrades.slice(0, 5)} compact />
              </section>
            </div>
          )}

          {tab === "groups" && (
            <div className="two-column">
              <section className="surface form-surface">
                <div className="section-heading">
                  <div>
                    <p>Управление</p>
                    <h2>Группы и предметы</h2>
                  </div>
                  <School size={20} aria-hidden="true" />
                </div>

                <form className="inline-form" onSubmit={handleCreateGroup}>
                  <label>
                    Новая группа
                    <input value={newGroup} onChange={(event) => setNewGroup(event.target.value)} required />
                  </label>
                  <button className="primary-button" type="submit" title="Создать группу" disabled={saving}>
                    <Plus size={16} aria-hidden="true" />
                    Создать
                  </button>
                </form>

                <form className="inline-form" onSubmit={handleCreateSubject}>
                  <label>
                    Новый предмет
                    <input value={newSubject} onChange={(event) => setNewSubject(event.target.value)} required />
                  </label>
                  <button className="primary-button" type="submit" title="Создать предмет" disabled={saving}>
                    <Plus size={16} aria-hidden="true" />
                    Создать
                  </button>
                </form>

                <details className="advanced-actions">
                  <summary>Привязать существующие</summary>
                  <form className="inline-form" onSubmit={handleAttachGroup}>
                    <label>
                      ID группы
                      <input
                        type="number"
                        min="1"
                        value={attachGroupId}
                        onChange={(event) => setAttachGroupId(event.target.value)}
                        required
                      />
                    </label>
                    <button className="secondary-button" type="submit" title="Привязать группу" disabled={saving}>
                      <Check size={16} aria-hidden="true" />
                      Привязать группу
                    </button>
                  </form>

                  <form className="inline-form" onSubmit={handleAttachSubject}>
                    <label>
                      ID предмета
                      <input
                        type="number"
                        min="1"
                        value={attachSubjectId}
                        onChange={(event) => setAttachSubjectId(event.target.value)}
                        required
                      />
                    </label>
                    <button className="secondary-button" type="submit" title="Привязать предмет" disabled={saving}>
                      <Check size={16} aria-hidden="true" />
                      Привязать предмет
                    </button>
                  </form>
                </details>
              </section>

              <section className="surface">
                <div className="section-heading">
                  <div>
                    <p>Студенты</p>
                    <h2>Добавить в группу</h2>
                  </div>
                  <Users size={20} aria-hidden="true" />
                </div>
                <form className="form-stack" onSubmit={handleAddStudent}>
                  <label>
                    ID ученика
                    <input
                      type="number"
                      min="1"
                      value={studentId}
                      onChange={(event) => setStudentId(event.target.value)}
                      required
                    />
                  </label>
                  <p className="field-note">
                    ID ученика появляется после регистрации на главной странице.
                  </p>
                  <label>
                    Группа
                    <select
                      value={studentGroupId}
                      onChange={(event) => setStudentGroupId(event.target.value)}
                      required
                    >
                      <option value="">Выберите группу</option>
                      {groups.map((group) => (
                        <option key={group.id} value={group.id.toString()}>
                          {group.name} · ID {group.id}
                        </option>
                      ))}
                    </select>
                  </label>
                  <button className="primary-button" type="submit" title="Добавить ученика" disabled={saving}>
                    <UserPlus size={16} aria-hidden="true" />
                    Добавить
                  </button>
                </form>

                {!!createdSubjects.length && (
                  <div className="local-list">
                    <p>Созданные предметы в этой сессии</p>
                    {createdSubjects.map((subject) => (
                      <span key={subject.id}>
                        {subject.name} · ID {subject.id}
                      </span>
                    ))}
                  </div>
                )}
              </section>
            </div>
          )}

          {tab === "grades" && (
            <div className="two-column wide-left">
              <section className="surface">
                <div className="section-heading">
                  <div>
                    <p>{selectedGroup?.name || "Группа не выбрана"}</p>
                    <h2>Журнал оценок</h2>
                  </div>
                  <BookOpen size={20} aria-hidden="true" />
                </div>
                <div className="toolbar">
                  <select
                    value={selectedGroupId?.toString() || ""}
                    onChange={(event) => loadGradesForGroup(Number(event.target.value))}
                  >
                    {groups.map((group) => (
                      <option key={group.id} value={group.id.toString()}>
                        {group.name}
                      </option>
                    ))}
                  </select>
                  <select
                    value={gradeStudentFilter}
                    onChange={(event) => setGradeStudentFilter(event.target.value)}
                  >
                    <option value="all">Все студенты</option>
                    {knownStudents.map((student) => (
                      <option key={student.id} value={student.id}>
                        {student.name}
                      </option>
                    ))}
                  </select>
                  <select
                    value={gradeSubjectFilter}
                    onChange={(event) => setGradeSubjectFilter(event.target.value)}
                  >
                    <option value="all">Все предметы</option>
                    {journalSubjects.map((subject) => (
                      <option key={subject.id} value={subject.id}>
                        {subject.name}
                      </option>
                    ))}
                  </select>
                </div>
                <GradeSummary grades={filteredGroupGrades} mode="mixed" />
                <GradeTable grades={filteredGroupGrades} />
              </section>

              <form className="surface form-surface" onSubmit={handleSetGrade}>
                <div className="section-heading">
                  <div>
                    <p>Новая оценка</p>
                    <h2>Выставление</h2>
                  </div>
                  <Plus size={20} aria-hidden="true" />
                </div>
                <label>
                  ID ученика
                  <input
                    type="number"
                    list="known-students-for-grade"
                    min="1"
                    value={gradeStudentId}
                    onChange={(event) => setGradeStudentId(event.target.value)}
                    required
                  />
                  <datalist id="known-students-for-grade">
                    {knownStudents.map((student) => (
                      <option key={student.id} value={student.id}>
                        {student.name}
                      </option>
                    ))}
                  </datalist>
                </label>
                <label>
                  ID предмета
                  <input
                    type="number"
                    list="known-subjects-for-grade"
                    min="1"
                    value={gradeSubjectId}
                    onChange={(event) => setGradeSubjectId(event.target.value)}
                    required
                  />
                  <datalist id="known-subjects-for-grade">
                    {knownSubjects.map((subject) => (
                      <option key={subject.id} value={subject.id}>
                        {subject.name}
                      </option>
                    ))}
                  </datalist>
                </label>
                <p className="field-note">
                  Подсказки берутся из уже загруженного журнала и предметов, созданных в этой сессии.
                </p>
                <label>
                  Оценка
                  <select value={gradeValue} onChange={(event) => setGradeValue(event.target.value)}>
                    {[5, 4, 3, 2, 1].map((value) => (
                      <option key={value} value={value}>
                        {value}
                      </option>
                    ))}
                  </select>
                </label>
                <label>
                  Дата занятия
                  <input
                    type="date"
                    value={gradeLessonDate}
                    onChange={(event) => setGradeLessonDate(event.target.value)}
                    required
                  />
                </label>
                <button className="primary-button" type="submit" title="Поставить оценку" disabled={saving}>
                  <Check size={16} aria-hidden="true" />
                  Поставить
                </button>
              </form>
            </div>
          )}

          {tab === "requests" && (
            <section className="surface">
              <div className="section-heading">
                <div>
                  <p>Подтверждение</p>
                  <h2>Заявки на активности</h2>
                </div>
                <Medal size={20} aria-hidden="true" />
              </div>
              <div className="toolbar filter-toolbar">
                <select
                  value={requestGroupFilter}
                  onChange={(event) => setRequestGroupFilter(event.target.value)}
                >
                  <option value="all">Все группы</option>
                  {groups.map((group) => (
                    <option key={group.id} value={group.id}>
                      {group.name}
                    </option>
                  ))}
                </select>
                <input
                  value={requestSearch}
                  onChange={(event) => setRequestSearch(event.target.value)}
                  placeholder="Поиск по ID, названию или описанию"
                />
              </div>
              <TeacherAchievementQueue
                achievements={filteredPendingAchievements}
                saving={saving}
                onAction={handleAchievementAction}
              />
            </section>
          )}

          {tab === "tokens" && (
            <>
              <div className="two-column">
                <section className="surface form-surface">
                <div className="section-heading">
                  <div>
                    <p>AMT</p>
                    <h2>Ручное начисление</h2>
                  </div>
                  <Coins size={20} aria-hidden="true" />
                </div>
                <div className="toolbar filter-toolbar">
                  <select
                    value={selectedGroupId?.toString() || ""}
                    onChange={(event) => loadGradesForGroup(Number(event.target.value))}
                  >
                    {groups.map((group) => (
                      <option key={group.id} value={group.id.toString()}>
                        {group.name}
                      </option>
                    ))}
                  </select>
                </div>
                <form className="form-stack" onSubmit={handleAwardTokens}>
                  <label>
                    ID ученика
                    <input
                      type="number"
                      min="1"
                      value={awardStudentId}
                      onChange={(event) => setAwardStudentId(event.target.value)}
                      required
                    />
                  </label>
                  <label>
                    Количество AMT
                    <input
                      type="number"
                      min="1"
                      value={awardAmount}
                      onChange={(event) => setAwardAmount(event.target.value)}
                      required
                    />
                  </label>
                  <button className="primary-button" type="submit" title="Начислить токены" disabled={saving}>
                    <Coins size={16} aria-hidden="true" />
                    Начислить
                  </button>
                </form>
                </section>

                <section className="surface">
                  <div className="section-heading">
                    <div>
                      <p>{selectedGroup?.name || "Группа не выбрана"}</p>
                      <h2>Студенты группы</h2>
                    </div>
                    <Users size={20} aria-hidden="true" />
                  </div>
                  <StudentShortcutList
                    students={knownStudents}
                    onPick={(id) => setAwardStudentId(String(id))}
                  />
                </section>
              </div>

              <section className="surface">
                <div className="section-heading">
                  <div>
                    <p>{selectedGroup?.name || "Группа не выбрана"}</p>
                    <h2>История операций AMT</h2>
                  </div>
                  <ListChecks size={20} aria-hidden="true" />
                </div>
                <TokenOperationTable operations={tokenOperations} showStudent />
              </section>
            </>
          )}
        </>
      )}
    </div>
  );
}

function RoleHero({
  icon,
  eyebrow,
  title,
  text,
  actions,
}: {
  icon: React.ReactNode;
  eyebrow: string;
  title: string;
  text: string;
  actions?: React.ReactNode;
}) {
  return (
    <section className="role-hero">
      <div className="role-icon">{icon}</div>
      <div>
        <p>{eyebrow}</p>
        <h1>{title}</h1>
        <span>{text}</span>
      </div>
      <div className="role-actions">{actions}</div>
    </section>
  );
}

function TabBar<T extends string>({
  active,
  items,
  onChange,
}: {
  active: T;
  items: { id: T; label: string; icon: React.ReactNode }[];
  onChange: (id: T) => void;
}) {
  return (
    <nav className="tabbar" aria-label="Разделы кабинета">
      {items.map((item) => (
        <button
          key={item.id}
          type="button"
          className={active === item.id ? "active" : ""}
          title={item.label}
          onClick={() => onChange(item.id)}
        >
          {item.icon}
          {item.label}
        </button>
      ))}
    </nav>
  );
}

function MetricCard({
  label,
  value,
  caption,
  accent = "default",
}: {
  label: string;
  value: string;
  caption: string;
  accent?: "default" | "green" | "amber" | "blue";
}) {
  return (
    <article className={`metric ${accent}`}>
      <span>{label}</span>
      <strong>{value}</strong>
      <p>{caption}</p>
    </article>
  );
}

function PendingGroupState({
  balance,
  pendingCount,
  achievementCount,
  onRefresh,
}: {
  balance: number;
  pendingCount: number;
  achievementCount: number;
  onRefresh: () => void;
}) {
  return (
    <section className="surface pending-state">
      <div className="pending-icon">
        <School size={26} aria-hidden="true" />
      </div>
      <div>
        <p>Профиль ожидает группу</p>
        <h2>Преподаватель еще не добавил вас в учебную группу</h2>
        <span>
          После добавления откроются журнал, активности, покупки и остальные разделы кабинета.
        </span>
      </div>
      <div className="pending-metrics">
        <span>
          <strong>{balance}</strong>
          Баланс AMT
        </span>
        <span>
          <strong>{pendingCount}</strong>
          Заявок на проверке
        </span>
        <span>
          <strong>{achievementCount}</strong>
          Активностей
        </span>
      </div>
      <button className="primary-button" type="button" onClick={onRefresh}>
        <RefreshCw size={16} aria-hidden="true" />
        Проверить снова
      </button>
    </section>
  );
}

function RecommendationPanel({
  grades,
  recommendation,
  loading,
  error,
  onGenerate,
}: {
  grades: GradeView[];
  recommendation: StoredRecommendation | null;
  loading: boolean;
  error: string | null;
  onGenerate: () => void;
}) {
  const payload = recommendation?.payload;
  const hasGrades = grades.length > 0;
  const strengths = payload?.strengths || [];
  const weaknesses = payload?.weaknesses || [];
  const subjectRecommendations = payload?.recommendations || [];
  
  const recommendationServiceAvailable = true;

  return (
    <section className="surface recommendations-panel">
      <div className="section-heading recommendations-heading">
        <div>
          <p>AI-помощник</p>
          <h2>Рекомендации по курсам</h2>
        </div>
        <button
          className="primary-button"
          type="button"
          title={recommendationServiceAvailable ? "Сгенерировать персональные рекомендации" : "Сервис временно недоступен"}
          disabled={!recommendationServiceAvailable || loading || !hasGrades}
          onClick={recommendationServiceAvailable ? onGenerate : undefined}
        >
          {loading ? <Loader2 className="spin" size={18} /> : <Sparkles size={18} />}
          {loading ? "Генерируем..." : "Сгенерировать рекомендации"}
        </button>
      </div>

      {!recommendationServiceAvailable && (
        <div className="recommendation-alert">
          <Sparkles size={18} aria-hidden="true" />
          <span>Сервис рекомендаций временно недоступен.</span>
        </div>
      )}

      {recommendationServiceAvailable && !hasGrades && (
        <EmptyState
          title="Нет оценок для анализа"
          text="Рекомендации появятся после того, как преподаватель добавит оценки в журнал."
        />
      )}

      {recommendationServiceAvailable && hasGrades && error && (
        <div className="recommendation-alert">
          <Sparkles size={18} aria-hidden="true" />
          <span>{error}</span>
        </div>
      )}

      {recommendationServiceAvailable && hasGrades && !payload && !error && !loading && (
        <EmptyState
          title="Рекомендаций пока нет"
          text="Нажмите кнопку, чтобы отправить текущие оценки в AI-сервис и сохранить результат."
        />
      )}

      {recommendationServiceAvailable && payload && (
        <div className="recommendations-layout">
          <div className="recommendation-overview">
            <div>
              <span>Последнее обновление</span>
              <strong>{recommendation ? formatDate(recommendation.created_at) : "нет данных"}</strong>
            </div>
            <div>
              <span>Оценок в анализе</span>
              <strong>{grades.length}</strong>
            </div>
            <div>
              <span>Предметов</span>
              <strong>{new Set(grades.map((grade) => grade.subject_id)).size}</strong>
            </div>
          </div>

          <div className="recommendation-columns">
            <RecommendationList title="Сильные стороны" items={strengths} variant="strength" />
            <RecommendationList title="Что подтянуть" items={weaknesses} variant="weakness" />
          </div>

          {payload.general_advice && (
            <article className="general-advice">
              <span>Общий совет</span>
              <p>{payload.general_advice}</p>
            </article>
          )}

          <div className="recommendation-card-grid">
            {subjectRecommendations.map((item) => (
              <article className="recommendation-card" key={`${item.subject}-${item.score}`}>
                <div className="recommendation-card-head">
                  <div>
                    <span>Курс</span>
                    <h3>{item.subject}</h3>
                  </div>
                  <strong>{item.score}</strong>
                </div>
                <p>{item.recommendation}</p>
              </article>
            ))}
          </div>
        </div>
      )}
    </section>
  );
}

function RecommendationList({
  title,
  items,
  variant,
}: {
  title: string;
  items: string[];
  variant: "strength" | "weakness";
}) {
  return (
    <div className={`recommendation-list ${variant}`}>
      <h3>{title}</h3>
      {items.length ? (
        <ul>
          {items.map((item) => (
            <li key={item}>{item}</li>
          ))}
        </ul>
      ) : (
        <p>Данных пока нет.</p>
      )}
    </div>
  );
}

function GradeSummary({
  grades,
  mode,
}: {
  grades: GradeView[];
  mode: "subjects" | "mixed";
}) {
  if (!grades.length) return null;

  const buildStats = (items: Array<{ key: string; label: string; value: number }>) => {
    const byKey = new Map<string, { label: string; total: number; count: number }>();

    items.forEach((item) => {
      const current = byKey.get(item.key) || { label: item.label, total: 0, count: 0 };
      current.total += item.value;
      current.count += 1;
      byKey.set(item.key, current);
    });

    return Array.from(byKey, ([key, item]) => ({
      key,
      label: item.label,
      average: (item.total / item.count).toFixed(1),
      count: item.count,
    })).sort((left, right) => left.label.localeCompare(right.label, "ru"));
  };

  const subjectStats = buildStats(
    grades.map((grade) => ({
      key: String(grade.subject_id),
      label: grade.subject_name,
      value: grade.value,
    })),
  );
  const studentStats =
    mode === "mixed"
      ? buildStats(
          grades.map((grade) => ({
            key: String(grade.student_id),
            label: grade.student_name,
            value: grade.value,
          })),
        )
      : [];

  return (
    <div className={`grade-summary ${mode === "subjects" ? "single" : ""}`}>
      <SummaryBlock title="По предметам" items={subjectStats} />
      {mode === "mixed" && <SummaryBlock title="По студентам" items={studentStats} />}
    </div>
  );
}

function SummaryBlock({
  title,
  items,
}: {
  title: string;
  items: Array<{ key: string; label: string; average: string; count: number }>;
}) {
  return (
    <div className="summary-block">
      <p>{title}</p>
      <div>
        {items.map((item) => (
          <span key={item.key}>
            <strong>{item.label}</strong>
            {item.average} · {item.count}
          </span>
        ))}
      </div>
    </div>
  );
}

function GradeTable({
  grades,
  compact = false,
  showStudent = true,
}: {
  grades: GradeView[];
  compact?: boolean;
  showStudent?: boolean;
}) {
  if (!grades.length) {
    return <EmptyState title="Оценок пока нет" text="Данные появятся после выставления оценки учителем." />;
  }

  return (
    <div className="table-wrap">
      <table className={compact ? "compact" : ""}>
        <thead>
          <tr>
            {showStudent && <th>Ученик</th>}
            <th>Предмет</th>
            <th>Оценка</th>
            <th>Дата</th>
          </tr>
        </thead>
        <tbody>
          {grades.map((grade) => (
            <tr key={grade.id}>
              {showStudent && (
                <td>
                  <strong>{grade.student_name}</strong>
                  <span>ID {grade.student_id}</span>
                </td>
              )}
              <td>{grade.subject_name}</td>
              <td>
                <span className={`grade-pill grade-${grade.value}`}>{grade.value}</span>
              </td>
              <td>{grade.lesson_date || (grade.created_at ? formatDate(grade.created_at) : "—")}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function StudentShortcutList({
  students,
  onPick,
}: {
  students: StudentOption[];
  onPick: (studentId: number) => void;
}) {
  if (!students.length) {
    return <EmptyState title="ID пока нет" text="Студенты появятся здесь после первых оценок в выбранной группе." />;
  }

  return (
    <div className="shortcut-list">
      {students.map((student) => (
        <button key={student.id} type="button" onClick={() => onPick(student.id)}>
          <User size={14} aria-hidden="true" />
          <span>{student.name}</span>
          <strong>ID {student.id}</strong>
        </button>
      ))}
    </div>
  );
}

function PeopleList({
  students,
  currentStudentId,
}: {
  students: Student[];
  currentStudentId: number | null;
}) {
  if (!students.length) {
    return <EmptyState title="Список пуст" text="Учитель добавит учеников после регистрации." />;
  }

  return (
    <div className="people-list">
      {students.map((student) => (
        <div className="person-row" key={student.id}>
          <div className="avatar">{student.name.slice(0, 1).toUpperCase()}</div>
          <div>
            <strong>{student.name}</strong>
            <span>{student.id === currentStudentId ? "Это вы" : "Одногруппник"}</span>
          </div>
        </div>
      ))}
    </div>
  );
}

function AchievementList({ achievements }: { achievements: Achievement[] }) {
  if (!achievements.length) {
    return <EmptyState title="Заявок пока нет" text="Отправьте первую активность на проверку." />;
  }

  return (
    <div className="list-stack">
      {achievements.map((achievement) => (
        <article className="line-item" key={achievement.id}>
          <div>
            <h3>{achievement.title}</h3>
            <p>{achievement.description || "Без описания"}</p>
          </div>
          <span className={`status ${achievement.status}`}>{statusText[achievement.status]}</span>
        </article>
      ))}
    </div>
  );
}

function WalletSummary({
  balance,
  walletAddress,
  tokenEvents,
}: {
  balance: number;
  walletAddress: string | null;
  tokenEvents: TokenEvent[];
}) {
  return (
    <div className="wallet-summary">
      <div className="wallet-balance">
        <span>Баланс</span>
        <strong>{balance} AMT</strong>
      </div>
      <div className="wallet-address">
        <span>Адрес</span>
        <code>{walletAddress || "недоступен без student_id"}</code>
      </div>
      <TokenHistory events={tokenEvents} />
    </div>
  );
}

function TokenHistory({ events }: { events: TokenEvent[] }) {
  if (!events.length) {
    return <EmptyState title="Операций пока нет" text="История появится после подтверждения активностей или покупок." />;
  }

  return (
    <div className="token-history">
      {events.map((event) => (
        <article className="token-event" key={event.id}>
          <div>
            <h3>{event.title}</h3>
            <p>{event.date ? formatDate(event.date) : event.detail}</p>
          </div>
          <strong className={event.kind}>
            {event.amount > 0 ? "+" : ""}
            {event.amount} AMT
          </strong>
        </article>
      ))}
    </div>
  );
}

function TokenOperationTable({
  operations,
  showStudent = false,
}: {
  operations: TokenOperation[];
  showStudent?: boolean;
}) {
  if (!operations.length) {
    return <EmptyState title="Операций пока нет" text="Начисления и покупки выбранной группы появятся здесь после первых действий с AMT." />;
  }

  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            {showStudent && <th>Студент</th>}
            <th>Операция</th>
            <th>Сумма</th>
            <th>Дата</th>
            <th>Преподаватель</th>
          </tr>
        </thead>
        <tbody>
          {operations.map((operation) => (
            <tr key={operation.id}>
              {showStudent && (
                <td>
                  <strong>{operation.student_name || `Студент #${operation.student_id}`}</strong>
                  <span>ID {operation.student_id}</span>
                </td>
              )}
              <td>
                <strong>{operation.reason || tokenOperationTitle[operation.operation_type]}</strong>
                <span>{tokenOperationTitle[operation.operation_type]}</span>
              </td>
              <td>
                <span className={`amount-pill ${operation.amount >= 0 ? "income" : "expense"}`}>
                  {operation.amount > 0 ? "+" : ""}
                  {operation.amount} AMT
                </span>
              </td>
              <td>{formatDate(operation.created_at)}</td>
              <td>{operation.teacher_name || "—"}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function SubjectList({ subjects }: { subjects: string[] }) {
  if (!subjects.length) {
    return <EmptyState title="Курсов пока нет" text="Предметы появятся после первых оценок в журнале." />;
  }

  return (
    <div className="subject-list">
      {subjects.map((subject) => (
        <span key={subject}>
          <BookOpen size={14} aria-hidden="true" />
          {subject}
        </span>
      ))}
    </div>
  );
}

function TeacherAchievementQueue({
  achievements,
  saving,
  onAction,
}: {
  achievements: PendingAchievementView[];
  saving: boolean;
  onAction: (achievementId: number, action: "confirm" | "deny") => void;
}) {
  if (!achievements.length) {
    return <EmptyState title="Заявок нет" text="Новые активности появятся здесь после отправки учениками." />;
  }

  return (
    <div className="list-stack">
      {achievements.map((achievement) => (
        <article className="line-item action-row" key={achievement.id}>
          <div>
            <h3>{achievement.title}</h3>
            <p>{achievement.description || "Без описания"}</p>
            <span>{achievement.student_name || `Ученик #${achievement.student_id}`}{achievement.group_name ? ` · ${achievement.group_name}` : ""}</span>
          </div>
          <div className="row-actions">
            <button
              className="icon-action approve"
              type="button"
              title="Подтвердить активность"
              disabled={saving}
              onClick={() => onAction(achievement.id, "confirm")}
            >
              <Check size={16} aria-hidden="true" />
            </button>
            <button
              className="icon-action deny"
              type="button"
              title="Отклонить активность"
              disabled={saving}
              onClick={() => onAction(achievement.id, "deny")}
            >
              <X size={16} aria-hidden="true" />
            </button>
          </div>
        </article>
      ))}
    </div>
  );
}

function PurchasesTable({ purchases }: { purchases: Purchase[] }) {
  if (!purchases.length) {
    return <EmptyState title="Покупок пока нет" text="История появится после обмена AMT на мерч." />;
  }

  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Товар</th>
            <th>Цена</th>
            <th>Дата</th>
          </tr>
        </thead>
        <tbody>
          {purchases.map((purchase) => (
            <tr key={purchase.id}>
              <td>
                <strong>{purchase.title}</strong>
                <span>Покупка #{purchase.id}</span>
              </td>
              <td>{purchase.price} AMT</td>
              <td>{formatDate(purchase.created_at)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function GroupList({
  groups,
  selectedGroupId,
  onSelect,
}: {
  groups: Group[];
  selectedGroupId: number | null;
  onSelect: (groupId: number) => void;
}) {
  if (!groups.length) {
    return <EmptyState title="Групп пока нет" text="Создайте группу или привяжите существующую." />;
  }

  return (
    <div className="group-list">
      {groups.map((group) => (
        <button
          key={group.id}
          type="button"
          className={group.id === selectedGroupId ? "selected" : ""}
          title="Выбрать группу"
          onClick={() => onSelect(group.id)}
        >
          <School size={16} aria-hidden="true" />
          <span>{group.name}</span>
          <small>ID {group.id}</small>
        </button>
      ))}
    </div>
  );
}

function EmptyState({ title, text }: { title: string; text: string }) {
  return (
    <div className="empty-state">
      <BadgeCheck size={20} aria-hidden="true" />
      <strong>{title}</strong>
      <p>{text}</p>
    </div>
  );
}

function CenteredState({ icon, title }: { icon: React.ReactNode; title: string }) {
  return (
    <div className="centered-state">
      {icon}
      <p>{title}</p>
    </div>
  );
}

export default App;
