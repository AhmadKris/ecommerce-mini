export type PromotionType = "percentage" | "fixed";
export type PromotionStatus = "active" | "inactive";

export interface Promotion {
  id: number;
  code: string;
  type: PromotionType;
  value: number;
  minimum_purchase: number;
  usage_limit: number;
  used_count: number;
  starts_at: string;
  ends_at: string;
  status: PromotionStatus;
  created_at: string;
  updated_at: string;
}
