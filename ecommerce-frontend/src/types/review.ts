import type { Product } from "./product";

export type ReviewStatus = "pending" | "approved" | "rejected";

export interface Review {
  id: number;
  product_id: number;
  user_id: number;
  order_id: number;
  rating: number;
  title: string;
  body: string;
  status: ReviewStatus;
  created_at: string;
  updated_at: string;
  product?: Product;
}
