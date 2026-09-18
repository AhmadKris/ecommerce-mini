import type { Product } from "./product";

export interface OrderItem {
  id: number;
  order_id: number;
  product_id: number;
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
