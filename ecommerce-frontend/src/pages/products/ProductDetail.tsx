import { Link, useParams } from "react-router-dom";

import { AddToCartButton } from "../../components/product/AddToCartButton";
import { ProductReviews } from "../../components/product/ProductReviews";
import { formatCurrency } from "../../lib/format";
import { useProduct } from "../../hooks/useProducts";

export function ProductDetail() {
  const { slug = "" } = useParams<{ slug: string }>();
  const { data: product, isLoading, isError } = useProduct(slug);

  if (isLoading) {
    return (
      <main className="mx-auto max-w-3xl px-6 py-12">
        <p className="text-body-md text-(--ink-secondary)">Memuat produk...</p>
      </main>
    );
  }

  if (isError || !product) {
    return (
      <main className="mx-auto max-w-3xl px-6 py-12">
        <p role="alert" className="text-body-md text-error-500">
          Produk tidak ditemukan.
        </p>
        <Link to="/products" className="text-body-sm text-(--ink-link) mt-2 inline-block">
          Kembali ke daftar produk
        </Link>
      </main>
    );
  }

  return (
    <main className="mx-auto max-w-3xl px-6 py-12">
      <Link to="/products" className="text-body-sm text-(--ink-link)">
        ← Kembali
      </Link>

      <div className="mt-4 grid gap-8 sm:grid-cols-2">
        <div className="aspect-square overflow-hidden rounded-lg bg-neutral-100">
          {product.image_url ? (
            <img
              src={product.image_url}
              alt={product.name}
              className="h-full w-full object-cover"
            />
          ) : (
            <div className="flex h-full w-full items-center justify-center text-body-sm text-(--ink-secondary)">
              Tidak ada gambar
            </div>
          )}
        </div>

        <div>
          {product.category && (
            <p className="text-label-md text-(--ink-link)">{product.category.name}</p>
          )}
          <h1 className="text-heading-xl text-(--ink-primary) mt-1">{product.name}</h1>
          <p className="text-data-md text-(--ink-primary) mt-2">{formatCurrency(product.price)}</p>
          <p className="text-body-sm text-(--ink-secondary) mt-1">
            {product.stock > 0 ? `Stok: ${product.stock}` : "Stok habis"}
          </p>
          <p className="text-body-lg text-(--ink-primary) mt-6">
            {product.description || "Tidak ada deskripsi."}
          </p>
          <div className="mt-6 max-w-xs">
            <AddToCartButton product={product} />
          </div>
        </div>
      </div>

      <ProductReviews slug={product.slug} />
    </main>
  );
}
