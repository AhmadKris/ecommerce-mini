import { Link } from "react-router-dom";

import { AddToCartButton } from "./AddToCartButton";
import { formatCurrency } from "../../lib/format";
import type { Product } from "../../types/product";

interface ProductCardProps {
  product: Product;
}

export function ProductCard({ product }: ProductCardProps) {
  return (
    <div className="overflow-hidden rounded-lg border border-(--border-default) bg-(--surface-card) hover:border-primary-300">
      <Link to={`/products/${product.slug}`} className="block">
        <div className="relative aspect-square bg-neutral-100">
          {product.image_url ? (
            <img src={product.image_url} alt={product.name} className="h-full w-full object-cover" />
          ) : (
            <div className="flex h-full w-full items-center justify-center text-body-sm text-(--ink-secondary)">
              Tidak ada gambar
            </div>
          )}
          {product.stock === 0 && (
            <span className="text-label-sm absolute top-2 left-2 rounded-full border border-(--border-default) bg-(--surface-card) px-2.5 py-1 text-(--ink-secondary)">
              Stok habis
            </span>
          )}
        </div>
        <div className="px-4 pt-4">
          <p className="text-heading-lg text-(--ink-primary) line-clamp-1">{product.name}</p>
          <p className="text-data-md text-(--ink-primary) mt-1">{formatCurrency(product.price)}</p>
        </div>
      </Link>
      <div className="p-4">
        <AddToCartButton product={product} />
      </div>
    </div>
  );
}
