import { Link } from "react-router-dom";

import { useCart, useRemoveCartItem, useUpdateCartItem } from "../../hooks/useCart";
import { formatCurrency } from "../../lib/format";

// Pre-checkout estimate only — must match the backend's flatShippingCost
// (internal/repository/order_repository.go) so this preview lines up with
// the order actually created at checkout, which computes its own
// shipping_cost server-side.
const SHIPPING_COST = 25000;

export function Cart() {
  const { data: cart, isLoading, isError } = useCart();
  const updateItem = useUpdateCartItem();
  const removeItem = useRemoveCartItem();

  if (isLoading) {
    return (
      <main className="mx-auto max-w-5xl px-6 py-12">
        <p className="text-body-md text-(--ink-secondary)">Memuat cart...</p>
      </main>
    );
  }

  if (isError || !cart) {
    return (
      <main className="mx-auto max-w-5xl px-6 py-12">
        <p role="alert" className="text-body-md text-error-500">
          Gagal memuat cart.
        </p>
      </main>
    );
  }

  if (cart.items.length === 0) {
    return (
      <main className="mx-auto max-w-5xl px-6 py-12">
        <h1 className="text-heading-xl text-(--ink-primary)">Keranjang Anda kosong</h1>
        <Link to="/products" className="text-body-sm text-(--ink-link) mt-2 inline-block">
          Lihat produk
        </Link>
      </main>
    );
  }

  const grandTotal = cart.total + SHIPPING_COST;

  return (
    <main className="mx-auto max-w-5xl px-6 py-12">
      <h1 className="text-heading-xl text-(--ink-primary) mb-6">Keranjang Anda ({cart.items.length} item)</h1>

      <div className="flex items-start gap-8">
        <div className="flex flex-grow flex-col">
          {cart.items.map((item) => (
            <div
              key={item.id}
              className="flex items-center gap-4 border-b border-(--border-default) py-5"
            >
              <div className="h-20 w-20 flex-shrink-0 overflow-hidden rounded-md bg-neutral-100">
                {item.product.image_url ? (
                  <img
                    src={item.product.image_url}
                    alt={item.product.name}
                    className="h-full w-full object-cover"
                  />
                ) : (
                  <div className="flex h-full w-full items-center justify-center text-body-sm text-(--ink-secondary)">
                    IMG
                  </div>
                )}
              </div>

              <div className="flex flex-grow flex-col gap-1">
                <Link to={`/products/${item.product.slug}`} className="text-heading-sm text-(--ink-primary)">
                  {item.product.name}
                </Link>
                <button
                  type="button"
                  onClick={() => removeItem.mutate(item.id)}
                  disabled={removeItem.isPending}
                  className="self-start text-body-sm text-error-500"
                >
                  Hapus
                </button>
              </div>

              <div className="flex items-center gap-1 rounded-md border border-(--border-default)">
                <button
                  type="button"
                  aria-label={`Kurangi jumlah ${item.product.name}`}
                  onClick={() => updateItem.mutate({ itemId: item.id, quantity: item.quantity - 1 })}
                  disabled={updateItem.isPending || item.quantity <= 1}
                  className="h-8 w-8 text-heading-sm text-(--ink-primary) disabled:opacity-30"
                >
                  −
                </button>
                <span className="text-data-md text-(--ink-primary) w-6 text-center">{item.quantity}</span>
                <button
                  type="button"
                  aria-label={`Tambah jumlah ${item.product.name}`}
                  onClick={() => updateItem.mutate({ itemId: item.id, quantity: item.quantity + 1 })}
                  disabled={updateItem.isPending || item.quantity >= item.product.stock}
                  className="h-8 w-8 text-heading-sm text-(--ink-primary) disabled:opacity-30"
                >
                  +
                </button>
              </div>

              <div className="text-data-md text-(--ink-primary) w-32 text-right">
                {formatCurrency(item.subtotal)}
              </div>
            </div>
          ))}
        </div>

        <div className="flex w-[360px] flex-shrink-0 flex-col gap-4 rounded-md border border-(--border-default) bg-(--surface-card) p-6">
          <p className="text-heading-md text-(--ink-primary)">Ringkasan Pesanan</p>
          <div className="text-body-md text-(--ink-secondary) flex justify-between">
            <span>Subtotal</span>
            <span className="text-data-md text-(--ink-primary)">{formatCurrency(cart.total)}</span>
          </div>
          <div className="text-body-md text-(--ink-secondary) flex justify-between">
            <span>Ongkos kirim</span>
            <span className="text-data-md text-(--ink-primary)">{formatCurrency(SHIPPING_COST)}</span>
          </div>
          <div className="text-heading-sm text-(--ink-primary) flex justify-between border-t border-(--border-default) pt-4">
            <span>Total</span>
            <span className="text-data-md">{formatCurrency(grandTotal)}</span>
          </div>
          <Link
            to="/checkout"
            className="text-label-md bg-primary-600 text-(--ink-inverse) hover:bg-primary-700 rounded-md px-5 py-2.5 text-center transition-colors"
          >
            Checkout
          </Link>
        </div>
      </div>
    </main>
  );
}
