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
  email: string;
  group_id: number;
  tokens: number;
  user_id: number;
}

export interface GradeView {
  id: number;
  student_id: number;
  student_name: string;
  subject_id: number;
  subject_name: string;
  value: number;
}

export interface Grade {
  id: number;
  student_id: number;
  subject_id: number;
  value: number;
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
