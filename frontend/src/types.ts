export type Role = "student" | "teacher";

export interface SessionResponse {
  user_id: number;
  role: Role;
}

export interface LoginResponse {
  message: string;
  role: Role;
}

export interface MessageResponse {
  message: string;
}

export interface Group {
  id: number;
  name: string;
}

export interface Student {
  id: number;
  name: string;
  email?: string;
  group_id: number;
  tokens?: number;
  user_id?: number;
}

export interface GradeView {
  id: number;
  student_id: number;
  student_name: string;
  subject_id: number;
  subject_name: string;
  value: number;
  lesson_date?: string;
  created_at?: string;
}

export interface Grade {
  id: number;
  student_id: number;
  subject_id: number;
  value: number;
  lesson_date?: string;
  created_at?: string;
}

export interface Subject {
  id: number;
  name: string;
}

export interface Achievement {
  id: number;
  student_id: number;
  title: string;
  description?: string;
  status: "pending" | "confirmed" | "denied";
  confirmed: boolean;
  confirmed_by_teacher_id?: number;
}

export interface PendingAchievementView {
  id: number;
  student_id: number;
  student_name: string;
  group_id: number;
  group_name: string;
  title: string;
  description?: string;
  status: "pending" | "confirmed" | "denied";
  confirmed: boolean;
}

export interface Merch {
  id: number;
  title: string;
  description?: string;
  price: number;
}

export interface Purchase {
  id: number;
  student_id: number;
  merch_id: number;
  title: string;
  price: number;
  created_at: string;
}

export interface PurchaseResult {
  message: string;
  merch: Merch;
  purchase_id: number;
  new_balance: number;
}

export interface StudentGroupResponse {
  group: Group | null;
  students: Student[];
}

export interface StudentMeResponse {
  student_id: number;
  name: string;
  email: string;
  group_id: number;
  tokens: number;
  wallet_address: string;
}

export interface SubjectRecommendation {
  subject: string;
  score: number;
  recommendation: string;
}

export interface AIRecommendation {
  student_id: string;
  student_name: string;
  student_surname: string;
  strengths: string[];
  weaknesses: string[];
  recommendations: SubjectRecommendation[];
  general_advice: string;
}

export interface StoredRecommendation {
  id: number;
  student_id: number;
  payload: AIRecommendation;
  created_at: string;
}

export interface TokenOperation {
  id: number;
  student_id: number;
  student_name?: string;
  teacher_id?: number;
  teacher_name?: string;
  amount: number;
  operation_type: "achievement_reward" | "manual_award" | "purchase";
  reason?: string;
  created_at: string;
}
