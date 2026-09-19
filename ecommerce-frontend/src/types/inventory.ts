import type { Product } from "./product";

export type InventoryMovementType = "in" | "out" | "correction";

export interface InventoryMovement {
  id: number;
  product_id: number;
  type: InventoryMovementType;
  quantity: number;
  before_quantity: number;
  after_quantity: number;
  reason: string;
  reference: string;
  created_by: number;
  created_at: string;
  product?: Product;
}
