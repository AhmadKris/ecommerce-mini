import { Link } from "react-router-dom";

const links = [
  { to: "/admin/products", label: "Produk", description: "Tambah dan ubah produk katalog." },
  { to: "/admin/categories", label: "Kategori", description: "Kelola kategori produk." },
  { to: "/admin/orders", label: "Pesanan", description: "Lihat semua pesanan dari seluruh customer." },
  { to: "/admin/promotions", label: "Promo", description: "Kelola kode diskon dan masa berlakunya." },
  { to: "/admin/reviews", label: "Ulasan", description: "Moderasi ulasan produk dari customer." },
];

export function AdminDashboard() {
  return (
    <div className="mx-auto max-w-3xl">
      <h1 className="text-heading-xl text-(--ink-primary) mb-6">Admin</h1>
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        {links.map((link) => (
          <Link
            key={link.to}
            to={link.to}
            className="rounded-md border border-(--border-default) bg-(--surface-card) p-4 hover:bg-neutral-50"
          >
            <span className="text-heading-sm text-(--ink-primary)">{link.label}</span>
            <p className="text-body-sm text-(--ink-secondary) mt-1">{link.description}</p>
          </Link>
        ))}
      </div>
    </div>
  );
}
