import { NavLink, Outlet, useLocation } from "react-router-dom";

import { usePermission } from "../../hooks/usePermission";
import { useAuthStore } from "../../store/auth-store";

const navItems = [
  { to: "/admin", label: "Dashboard", end: true },
  { to: "/admin/products", label: "Produk" },
  { to: "/admin/categories", label: "Kategori" },
  { to: "/admin/orders", label: "Pesanan" },
  { to: "/admin/customers", label: "Pelanggan", permission: "customer:read" },
  { to: "/admin/inventory", label: "Stok" },
  { to: "/admin/promotions", label: "Promo" },
  { to: "/admin/reviews", label: "Ulasan" },
];

// Maps a path prefix to the breadcrumb section label shown in the top bar —
// checked longest-prefix-first so /admin/products/new doesn't match the
// bare /admin/products entry by accident.
const breadcrumbSections: { prefix: string; label: string }[] = [
  { prefix: "/admin/products", label: "Produk" },
  { prefix: "/admin/categories", label: "Kategori" },
  { prefix: "/admin/orders", label: "Pesanan" },
  { prefix: "/admin/customers", label: "Pelanggan" },
  { prefix: "/admin/inventory", label: "Stok" },
  { prefix: "/admin/promotions", label: "Promo" },
  { prefix: "/admin/reviews", label: "Ulasan" },
  { prefix: "/admin", label: "Dashboard" },
];

function currentSectionLabel(pathname: string): string {
  const match = breadcrumbSections.find((section) => pathname.startsWith(section.prefix));
  return match?.label ?? "";
}

/**
 * Shell for every /admin/* page — dark sidebar nav, separate from the
 * storefront's Header, so the admin panel reads as its own surface instead
 * of a customer page with extra links bolted on. Uses the --surface-admin*
 * design tokens that existed in index.css since the Design System was
 * transcribed but were never actually applied anywhere until now.
 */
export function AdminLayout() {
  const clearSession = useAuthStore((state) => state.clearSession);
  const location = useLocation();
  const canReadCustomers = usePermission("customer:read");
  const visibleNavItems = navItems.filter((item) => item.permission !== "customer:read" || canReadCustomers);

  return (
    <div className="flex min-h-screen">
      <aside className="w-60 shrink-0 bg-(--surface-admin-nav) px-4 py-6">
        <a href="/" className="text-heading-sm text-(--ink-inverse)">
          Ecommerce Mini
        </a>
        <nav className="mt-8 flex flex-col gap-1">
          {visibleNavItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.end}
              className={({ isActive }) =>
                `text-body-sm rounded-md px-3 py-2 text-(--ink-inverse) ${
                  isActive ? "bg-(--nav-active-bg)" : "hover:bg-white/10"
                }`
              }
            >
              {item.label}
            </NavLink>
          ))}
        </nav>
      </aside>

      <div className="flex-1 bg-(--surface-admin)">
        <header className="flex h-16 items-center justify-between border-b border-(--border-default) bg-(--surface-card) px-6">
          <span className="text-body-sm text-(--ink-secondary)">
            Admin / {currentSectionLabel(location.pathname)}
          </span>
          <button type="button" onClick={clearSession} className="text-body-sm text-(--ink-secondary)">
            Keluar
          </button>
        </header>
        <main className="p-6">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
