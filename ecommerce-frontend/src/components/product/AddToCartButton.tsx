import { useState } from "react";
import { useNavigate } from "react-router-dom";

import { Button } from "../ui/Button";
import { useAddCartItem } from "../../hooks/useCart";
import { useAuthStore } from "../../store/auth-store";
import { ApiError } from "../../types/api";
import type { Product } from "../../types/product";

interface AddToCartButtonProps {
  product: Product;
}

/**
 * Adds one unit of `product` to the cart. Redirects to /login instead of
 * calling the API when the viewer isn't authenticated — the backend would
 * reject it as 401 anyway, so this avoids a round-trip just to fail.
 */
export function AddToCartButton({ product }: AddToCartButtonProps) {
  const navigate = useNavigate();
  const isAuthenticated = useAuthStore((state) => state.accessToken !== null);
  const addCartItem = useAddCartItem();
  const [justAdded, setJustAdded] = useState(false);

  if (product.stock === 0) {
    return (
      <button
        type="button"
        disabled
        className="text-label-md w-full cursor-not-allowed rounded-md bg-neutral-200 px-4 py-2.5 text-neutral-400"
      >
        Stok habis
      </button>
    );
  }

  function handleClick() {
    if (!isAuthenticated) {
      navigate("/login");
      return;
    }
    addCartItem.mutate(
      { productId: product.id, quantity: 1 },
      {
        onSuccess: () => {
          setJustAdded(true);
          setTimeout(() => setJustAdded(false), 1500);
        },
      },
    );
  }

  return (
    <div>
      <Button type="button" className="w-full" onClick={handleClick} disabled={addCartItem.isPending}>
        {justAdded ? "Ditambahkan" : addCartItem.isPending ? "Menambahkan..." : "Tambah ke Keranjang"}
      </Button>
      {addCartItem.isError && (
        <p role="alert" className="text-body-sm text-error-500 mt-1">
          {addCartItem.error instanceof ApiError ? addCartItem.error.message : "Gagal menambahkan ke cart."}
        </p>
      )}
    </div>
  );
}
