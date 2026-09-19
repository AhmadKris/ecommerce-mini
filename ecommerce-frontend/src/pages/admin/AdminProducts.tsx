import { useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";

import { Button } from "../../components/ui/Button";
import { Input } from "../../components/ui/Input";
import { useCategories } from "../../hooks/useCategories";
import { useDebounce } from "../../hooks/useDebounce";
import { useProducts } from "../../hooks/useProducts";
import { formatCurrency } from "../../lib/format";
import type { Product } from "../../types/product";

/** Builds a CSV file from the currently loaded (filtered, paginated) rows
 * and triggers a browser download — client-side only, no export endpoint on
 * the backend since there's nothing else that needs this data server-side. */
function exportToCsv(products: Product[]) {
  const header = ["Nama", "SKU", "Kategori", "Harga", "Stok"];
  const rows = products.map((product) => [
    product.name,
    product.sku,
    product.category?.name ?? "",
    String(product.price),
    String(product.stock),
  ]);
  const csv = [header, ...rows]
    .map((row) => row.map((cell) => `"${cell.replace(/"/g, '""')}"`).join(","))
    .join("\n");

  const blob = new Blob([csv], { type: "text/csv;charset=utf-8;" });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = "produk.csv";
  link.click();
  URL.revokeObjectURL(url);
}

export function AdminProducts() {
  const [searchParams, setSearchParams] = useSearchParams();
  const page = Number(searchParams.get("page") ?? "1");
  const category = searchParams.get("category") ?? undefined;
  const [searchInput, setSearchInput] = useState(searchParams.get("search") ?? "");
  const debouncedSearch = useDebounce(searchInput, 300);

  const { data: categories } = useCategories();
  const { data, isLoading, isError } = useProducts({
    page,
    category,
    search: debouncedSearch || undefined,
  });
  const navigate = useNavigate();

  function goToPage(nextPage: number) {
    const next = new URLSearchParams(searchParams);
    next.set("page", String(nextPage));
    setSearchParams(next);
  }

  function handleSearchChange(value: string) {
    setSearchInput(value);
    const next = new URLSearchParams(searchParams);
    if (value) {
      next.set("search", value);
    } else {
      next.delete("search");
    }
    next.delete("page");
    setSearchParams(next);
  }

  function handleCategoryChange(value: string) {
    const next = new URLSearchParams(searchParams);
    if (value) {
      next.set("category", value);
    } else {
      next.delete("category");
    }
    next.delete("page");
    setSearchParams(next);
  }

  return (
    <div className="mx-auto max-w-4xl">
      <div className="flex items-center justify-between">
        <h1 className="text-heading-xl text-(--ink-primary)">Kelola Produk</h1>
        <Link to="/admin/products/new">
          <Button>Tambah Produk</Button>
        </Link>
      </div>

      <div className="mt-4 flex flex-wrap items-end justify-between gap-3">
        <div className="flex flex-wrap items-end gap-3">
          <div className="max-w-xs">
            <Input
              label="Cari produk"
              value={searchInput}
              onChange={(event) => handleSearchChange(event.target.value)}
              placeholder="Nama produk atau SKU..."
            />
          </div>
          <label className="text-body-sm text-(--ink-secondary) flex flex-col gap-1">
            Kategori
            <select
              value={category ?? ""}
              onChange={(event) => handleCategoryChange(event.target.value)}
              className="text-body-sm rounded-md border border-(--border-default) px-3 py-2.5"
            >
              <option value="">Semua kategori</option>
              {categories?.map((cat) => (
                <option key={cat.id} value={cat.slug}>
                  {cat.name}
                </option>
              ))}
            </select>
          </label>
        </div>
        <Button
          variant="secondary"
          onClick={() => data && exportToCsv(data.items)}
          disabled={!data || data.items.length === 0}
        >
          Export
        </Button>
      </div>

      {isLoading && <p className="text-body-md text-(--ink-secondary) mt-8">Memuat produk...</p>}
      {isError && (
        <p role="alert" className="text-body-md text-error-500 mt-8">
          Gagal memuat produk.
        </p>
      )}
      {data && data.items.length === 0 && (
        <p className="text-body-md text-(--ink-secondary) mt-8">Belum ada produk.</p>
      )}

      {data && data.items.length > 0 && (
        <>
          <div className="mt-6 overflow-x-auto rounded-md border border-(--border-default)">
            <table className="w-full text-left">
              <thead className="bg-neutral-50">
                <tr>
                  <th className="text-label-sm text-(--ink-secondary) px-4 py-3">Nama</th>
                  <th className="text-label-sm text-(--ink-secondary) px-4 py-3">SKU</th>
                  <th className="text-label-sm text-(--ink-secondary) px-4 py-3">Kategori</th>
                  <th className="text-label-sm text-(--ink-secondary) px-4 py-3">Harga</th>
                  <th className="text-label-sm text-(--ink-secondary) px-4 py-3">Stok</th>
                  <th className="text-label-sm text-(--ink-secondary) px-4 py-3" />
                </tr>
              </thead>
              <tbody>
                {data.items.map((product) => (
                  <tr key={product.id} className="border-t border-(--border-default)">
                    <td className="text-body-sm text-(--ink-primary) px-4 py-3">{product.name}</td>
                    <td className="text-data-md text-(--ink-secondary) px-4 py-3">{product.sku}</td>
                    <td className="text-body-sm text-(--ink-secondary) px-4 py-3">
                      {product.category?.name ?? "-"}
                    </td>
                    <td className="text-data-md text-(--ink-primary) px-4 py-3">
                      {formatCurrency(product.price)}
                    </td>
                    <td className="text-body-sm text-(--ink-primary) px-4 py-3">
                      {product.stock === 0 ? (
                        <span className="text-error-500">Habis</span>
                      ) : (
                        product.stock
                      )}
                    </td>
                    <td className="px-4 py-3 text-right">
                      <button
                        type="button"
                        className="text-label-sm text-(--ink-link)"
                        onClick={() =>
                          navigate(`/admin/products/${product.id}/edit`, { state: { product } })
                        }
                      >
                        Ubah
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div className="mt-6 flex items-center gap-3">
            <Button variant="secondary" onClick={() => goToPage(page - 1)} disabled={page <= 1}>
              Sebelumnya
            </Button>
            <span className="text-body-sm text-(--ink-secondary)">
              Halaman {data.meta.page} dari {data.meta.total_pages}
            </span>
            <Button
              variant="secondary"
              onClick={() => goToPage(page + 1)}
              disabled={page >= data.meta.total_pages}
            >
              Berikutnya
            </Button>
          </div>
        </>
      )}
    </div>
  );
}
