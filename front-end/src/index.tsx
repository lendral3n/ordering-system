// src/types/index.ts
export interface User {
  id: number;
  username: string;
  email: string;
  full_name: string;
  role: "admin" | "cashier" | "waiter" | "kitchen";
}

export interface Table {
  id: number;
  table_number: string;
  qr_code: string;
  status: "available" | "occupied" | "reserved";
  capacity: number;
  created_at: string;
  updated_at: string;
}

export interface MenuCategory {
  id: number;
  name: string;
  description?: string;
  display_order: number;
  is_active: boolean;
  menu_items?: MenuItem[];
}

export interface MenuItem {
  id: number;
  category_id: number;
  name: string;
  description?: string;
  price: number;
  image_url?: string;
  image_360_url?: string;
  video_url?: string;
  is_available: boolean;
  preparation_time?: number;
  stock_quantity?: number;
  category?: MenuCategory;
  media_files?: MediaFile[];
}

export interface Order {
  id: number;
  order_number: string;
  session_id: number;
  table_id: number;
  status: OrderStatus;
  total_amount: number;
  tax_amount: number;
  service_charge: number;
  grand_total: number;
  payment_status: PaymentStatus;
  payment_method?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
  order_items?: OrderItem[];
  table?: Table;
  payment?: Payment;
}

export interface OrderItem {
  id: number;
  order_id: number;
  menu_item_id: number;
  quantity: number;
  unit_price: number;
  subtotal: number;
  notes?: string;
  status: OrderItemStatus;
  menu_item?: MenuItem;
}

export interface Payment {
  id: number;
  order_id: number;
  midtrans_order_id: string;
  midtrans_transaction_id?: string;
  payment_type?: string;
  transaction_status?: string;
  transaction_time?: string;
  gross_amount: number;
  currency: string;
  va_number?: string;
  bank?: string;
  fraud_status?: string;
  status_message?: string;
}

export interface StaffNotification {
  id: number;
  staff_id?: number;
  order_id?: number;
  type: "new_order" | "payment_received" | "order_ready" | "assistance_request";
  message: string;
  is_read: boolean;
  read_at?: string;
  created_at: string;
  order?: Order;
}

export interface MediaFile {
  id: number;
  file_type: "image" | "video" | "image_360";
  file_url: string;
  thumbnail_url?: string;
  file_size: number;
  mime_type: string;
  menu_item_id?: number;
  uploaded_by: number;
}

export type OrderStatus =
  | "pending"
  | "confirmed"
  | "preparing"
  | "ready"
  | "served"
  | "completed"
  | "cancelled";

export type OrderItemStatus =
  | "pending"
  | "preparing"
  | "ready"
  | "served"
  | "cancelled";

export type PaymentStatus =
  | "unpaid"
  | "pending"
  | "paid"
  | "failed"
  | "refunded";

// API Response types
export interface ApiResponse<T> {
  success: boolean;
  message: string;
  data?: T;
  error?: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}

export interface OrderFilter {
  status?: OrderStatus;
  payment_status?: PaymentStatus;
  table_id?: number;
  date_from?: string;
  date_to?: string;
  limit?: number;
  offset?: number;
}

export interface PaymentFilter {
  status?: string;
  date_from?: string;
  date_to?: string;
  limit?: number;
  offset?: number;
}

export interface SalesAnalytics {
  total_revenue: number;
  total_orders: number;
  completed_orders: number;
  cancelled_orders: number;
  average_order_value: number;
  top_selling_items: Array<{
    name: string;
    quantity: number;
  }>;
  hourly_distribution: Record<number, number>;
  daily_revenue: Record<string, number>;
}
