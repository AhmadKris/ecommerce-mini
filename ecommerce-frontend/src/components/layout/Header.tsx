import { Link } from "react-router-dom";

import { useCart } from "../../hooks/useCart";
import { useAuthStore } from "../../store/auth-store";

export function Header() {
  const isAuthenticated = useAuthStore((state) => state.accessToken !== null);
  const isAdmin = useAuthStore((state) => state.roles.includes("admin"));
  const clearSession = useAuthStore((state) => state.clearSession);
  const { data: cart } = useCart();
  const itemCount = cart?.items.length ?? 0;

  return (
    <header className="sticky top-0 z-10 flex h-[72px] items-center justify-between border-b border-(--border-default) bg-(--surface-card) px-20">
      <Link to="/" className="text-heading-sm text-(--ink-primary)">
        Ecommerce Mini
      </Link>

      <nav className="flex items-center gap-8">
        <Link to="/products" className="text-body-md text-(--ink-primary)">
          Produk
        </Link>
        {isAuthenticated && (
          <Link to="/orders" className="text-body-md text-(--ink-primary)">
            Pesanan
          </Link>
        )}
        {isAdmin && (
          <Link to="/admin" className="text-body-md text-(--ink-primary)">
            Admin
          </Link>
        )}
      </nav>

      <div className="flex items-center gap-4">
        {isAuthenticated ? (
          <>
            <Link
              to="/cart"
              aria-label={`Lihat keranjang, ${itemCount} item`}
              className="flex items-center gap-1.5 text-body-md text-(--ink-primary)"
            >
              <CartIcon />
              Keranjang{itemCount > 0 ? ` (${itemCount})` : ""}
            </Link>
            <button type="button" onClick={clearSession} className="text-body-md text-(--ink-secondary)">
              Keluar
            </button>
          </>
        ) : (
          <Link
            to="/login"
            className="text-label-md rounded-md bg-primary-600 px-4 py-2 text-(--ink-inverse)"
          >
            Masuk
          </Link>
        )}
      </div>
    </header>
  );
}

function CartIcon() {
  return (
    <svg
      width="20"
      height="20"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.8"
      aria-hidden="true"
    >
      <path d="M3 3h2l2.4 12.4a2 2 0 0 0 2 1.6h7.2a2 2 0 0 0 2-1.6L21 7H6" />
      <circle cx="9" cy="20" r="1" />
      <circle cx="17" cy="20" r="1" />
    </svg>
  );
}
