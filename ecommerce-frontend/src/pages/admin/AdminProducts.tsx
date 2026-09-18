import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";

import { Button } from "../../components/ui/Button";
import { useProducts } from "../../hooks/useProducts";
import { formatCurrency } from "../../lib/format";

export function AdminProducts() {
  const [page, setPage] = useState(1);
  const { data, isLoading, isError } = useProducts({ page });
  const navigate = useNavigate();

  return (
    <main className="mx-auto max-w-4xl px-6 py-12">
      <div className="flex items-center justify-between">
        <h1 className="text-heading-xl text-(--ink-primary)">Kelola Produk</h1>
        <Link to="/admin/products/new">
          <Button>Tambah Produk</Button>
        </Link>
      </div>

      {isLoading && <p className="text-body-md text-(--ink-secondary) mt-8">Memuat produk...</p>}
      {isError && (
        <p role="alert" className="text-body-md text-error-500 mt-8">
          Gagal memuat produk.
        </p>
      )}

      {data && data.items.length > 0 && (
        <>
          <div className="mt-8 overflow-x-auto rounded-md border border-(--border-default)">
            <table className="w-full text-left">
              <thead className="bg-neutral-50">
                <tr>
                  <th className="text-label-sm text-(--ink-secondary) px-4 py-3">Nama</th>
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
            <Button variant="secondary" onClick={() => setPage((p) => p - 1)} disabled={page <= 1}>
              Sebelumnya
            </Button>
            <span className="text-body-sm text-(--ink-secondary)">
              Halaman {data.meta.page} dari {data.meta.total_pages}
            </span>
            <Button
              variant="secondary"
              onClick={() => setPage((p) => p + 1)}
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
