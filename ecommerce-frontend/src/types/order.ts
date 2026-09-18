import type { Product } from "./product";

export interface OrderItem {
  id: number;
  order_id: number;
  product_id: number;
  /** Snapshotted at checkout — always display this, not `product.name`,
   * which reflects the product's *current* name and can drift after a
   * rename. */
  product_name: string;
  quantity: number;
  price_at_purchase: number;
  product: Product;
}

export interface Order {
  id: number;
  user_id: number;
  status: string;
  total_amount: number;
  shipping_cost: number;
  shipping_address: string;
  created_at: string;
  items: OrderItem[];
}
