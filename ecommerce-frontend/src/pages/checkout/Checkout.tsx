import { zodResolver } from "@hookform/resolvers/zod";
import { useRef, useState } from "react";
import { useForm } from "react-hook-form";
import { Link } from "react-router-dom";

import { Button } from "../../components/ui/Button";
import { Input } from "../../components/ui/Input";
import { useCart } from "../../hooks/useCart";
import { useCheckout } from "../../hooks/useOrders";
import { formatCurrency } from "../../lib/format";
import { checkoutSchema, type CheckoutFormValues } from "../../schemas/checkout";
import { ApiError } from "../../types/api";
import type { Order } from "../../types/order";

// Pre-submit estimate only, shown before the order exists — must match the
// backend's flatShippingCost (internal/repository/order_repository.go) so
// this estimate lines up with the confirmation's real total_amount. The
// confirmation screen below never uses this constant; it displays the
// order's own shipping_cost/total_amount, which is the source of truth.
const ESTIMATED_SHIPPING_COST = 25000;

export function Checkout() {
  const { data: cart, isLoading } = useCart();
  const checkout = useCheckout();
  const [completedOrder, setCompletedOrder] = useState<Order | null>(null);
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<CheckoutFormValues>({ resolver: zodResolver(checkoutSchema) });

  // One key per checkout attempt, reused across retries (e.g. clicking
  // "Buat Pesanan" again after the first submit's connection drops) so the
  // backend replays the original order instead of creating a duplicate.
  const idempotencyKeyRef = useRef(crypto.randomUUID());

  function onSubmit(values: CheckoutFormValues) {
    checkout.mutate(
      { shippingAddress: values.shippingAddress, idempotencyKey: idempotencyKeyRef.current },
      { onSuccess: (order) => setCompletedOrder(order) },
    );
  }

  if (completedOrder) {
    return (
      <main className="mx-auto max-w-xl px-6 py-16">
        <h1 className="text-heading-xl text-(--ink-primary)">Pesanan berhasil dibuat</h1>
        <p className="text-body-md text-(--ink-secondary) mt-2">
          Order #{completedOrder.id} — total {formatCurrency(completedOrder.total_amount)}
          {" "}(termasuk ongkir {formatCurrency(completedOrder.shipping_cost)})
        </p>
        <Link to="/orders" className="text-body-sm text-(--ink-link) mt-4 inline-block">
          Lihat riwayat pesanan
        </Link>
      </main>
    );
  }

  if (isLoading) {
    return (
      <main className="mx-auto max-w-xl px-6 py-16">
        <p className="text-body-md text-(--ink-secondary)">Memuat...</p>
      </main>
    );
  }

  if (!cart || cart.items.length === 0) {
    return (
      <main className="mx-auto max-w-xl px-6 py-16">
        <h1 className="text-heading-xl text-(--ink-primary)">Keranjang Anda kosong</h1>
        <Link to="/products" className="text-body-sm text-(--ink-link) mt-2 inline-block">
          Lihat produk
        </Link>
      </main>
    );
  }

  const estimatedTotal = cart.total + ESTIMATED_SHIPPING_COST;

  return (
    <main className="mx-auto max-w-xl px-6 py-16">
      <h1 className="text-heading-xl text-(--ink-primary)">Checkout</h1>

      <div className="mt-6 flex flex-col gap-2 rounded-md border border-(--border-default) bg-(--surface-card) p-4">
        {cart.items.map((item) => (
          <div key={item.id} className="text-body-sm flex justify-between text-(--ink-secondary)">
            <span>
              {item.product.name} × {item.quantity}
            </span>
            <span className="text-data-md text-(--ink-primary)">{formatCurrency(item.subtotal)}</span>
          </div>
        ))}
        <div className="text-body-sm flex justify-between text-(--ink-secondary)">
          <span>Ongkos kirim</span>
          <span className="text-data-md text-(--ink-primary)">{formatCurrency(ESTIMATED_SHIPPING_COST)}</span>
        </div>
        <div className="text-heading-sm text-(--ink-primary) flex justify-between border-t border-(--border-default) pt-2">
          <span>Total</span>
          <span className="text-data-md">{formatCurrency(estimatedTotal)}</span>
        </div>
      </div>

      <form onSubmit={handleSubmit(onSubmit)} className="mt-6 flex flex-col gap-4" noValidate>
        <Input
          label="Alamat Pengiriman"
          error={errors.shippingAddress?.message}
          {...register("shippingAddress")}
        />
        {checkout.isError && (
          <p role="alert" className="text-body-sm text-error-500">
            {checkout.error instanceof ApiError ? checkout.error.message : "Gagal membuat pesanan."}
          </p>
        )}
        <Button type="submit" disabled={checkout.isPending}>
          {checkout.isPending ? "Memproses..." : "Buat Pesanan"}
        </Button>
      </form>
    </main>
  );
}
