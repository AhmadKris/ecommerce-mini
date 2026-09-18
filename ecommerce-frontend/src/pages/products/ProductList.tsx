import { useState } from "react";
import { useSearchParams } from "react-router-dom";

import { Button } from "../../components/ui/Button";
import { Input } from "../../components/ui/Input";
import { ProductCard } from "../../components/product/ProductCard";
import { useDebounce } from "../../hooks/useDebounce";
import { useProducts } from "../../hooks/useProducts";

export function ProductList() {
  const [searchParams, setSearchParams] = useSearchParams();
  const page = Number(searchParams.get("page") ?? "1");
  const category = searchParams.get("category") ?? undefined;
  const [searchInput, setSearchInput] = useState(searchParams.get("search") ?? "");
  const debouncedSearch = useDebounce(searchInput, 300);

  const { data, isLoading, isError } = useProducts({
    page,
    category,
    search: debouncedSearch || undefined,
  });

  function goToPage(nextPage: number) {
    const next = new URLSearchParams(searchParams);
    next.set("page", String(nextPage));
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
