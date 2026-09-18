import { useState } from "react";
import { useSearchParams } from "react-router-dom";

import { Button } from "../../components/ui/Button";
import { Input } from "../../components/ui/Input";
import { ProductCard } from "../../components/product/ProductCard";
import { useCategories } from "../../hooks/useCategories";
import { useDebounce } from "../../hooks/useDebounce";
import { useProducts, type ProductFilters } from "../../hooks/useProducts";

const sortOptions: { value: NonNullable<ProductFilters["sort"]>; label: string }[] = [
  { value: "newest", label: "Terbaru" },
  { value: "price_asc", label: "Harga: Rendah ke Tinggi" },
  { value: "price_desc", label: "Harga: Tinggi ke Rendah" },
];

export function ProductList() {
  const [searchParams, setSearchParams] = useSearchParams();
  const page = Number(searchParams.get("page") ?? "1");
  const category = searchParams.get("category") ?? undefined;
  const sort = (searchParams.get("sort") as ProductFilters["sort"]) ?? undefined;
  const [searchInput, setSearchInput] = useState(searchParams.get("search") ?? "");
  const debouncedSearch = useDebounce(searchInput, 300);

  const { data: categories } = useCategories();
  const { data, isLoading, isError } = useProducts({
    page,
    category,
    sort,
    search: debouncedSearch || undefined,
  });

  function goToPage(nextPage: number) {
    const next = new URLSearchParams(searchParams);
    next.set("page", String(nextPage));
    setSearchParams(next);
  }

  // Any filter change resets to page 1 — staying on, say, page 3 of a
  // now-different result set would silently show the wrong slice.
  function setFilter(key: "category" | "sort", value: string | undefined) {
    const next = new URLSearchParams(searchParams);
    if (value) {
      next.set(key, value);
    } else {
      next.delete(key);
    }
    next.delete("page");
    setSearchParams(next);
  }

  return (
    <main className="mx-auto max-w-5xl px-6 py-12">
      <h1 className="text-heading-xl text-(--ink-primary)">Produk</h1>

      <div className="mt-4 max-w-xs">
        <Input
          label="Cari produk"
          value={searchInput}
          onChange={(event) => setSearchInput(event.target.value)}
          placeholder="Nama produk..."
        />
      </div>

      <div className="mt-4 flex flex-wrap items-center justify-between gap-3">
        <div className="flex flex-wrap gap-2">
          <button
            type="button"
            onClick={() => setFilter("category", undefined)}
            className={`text-label-sm rounded-full border px-3 py-1.5 ${
              category === undefined
                ? "border-primary-600 text-primary-600"
                : "border-(--border-default) text-(--ink-secondary)"
            }`}
          >
            Semua
          </button>
          {categories?.map((cat) => (
            <button
              key={cat.id}
              type="button"
              onClick={() => setFilter("category", cat.slug)}
              className={`text-label-sm rounded-full border px-3 py-1.5 ${
                category === cat.slug
                  ? "border-primary-600 text-primary-600"
                  : "border-(--border-default) text-(--ink-secondary)"
              }`}
            >
              {cat.name}
            </button>
          ))}
        </div>

        <label className="text-body-sm text-(--ink-secondary) flex items-center gap-2">
          Urutkan:
          <select
            value={sort ?? "newest"}
            onChange={(event) => setFilter("sort", event.target.value)}
            className="text-body-sm rounded-md border border-(--border-default) px-2 py-1.5"
          >
            {sortOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
      </div>

      {isLoading && <p className="text-body-md text-(--ink-secondary) mt-8">Memuat produk...</p>}
      {isError && (
        <p role="alert" className="text-body-md text-error-500 mt-8">
          Gagal memuat produk. Coba muat ulang halaman.
        </p>
      )}
      {data && data.items.length === 0 && (
        <p className="text-body-md text-(--ink-secondary) mt-8">Belum ada produk.</p>
      )}

      {data && data.items.length > 0 && (
        <>
          <div className="mt-8 grid grid-cols-2 gap-6 sm:grid-cols-3">
            {data.items.map((product) => (
              <ProductCard key={product.id} product={product} />
            ))}
          </div>

          <div className="mt-8 flex items-center gap-3">
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
    </main>
  );
}
