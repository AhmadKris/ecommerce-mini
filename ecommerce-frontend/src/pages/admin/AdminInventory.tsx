import { zodResolver } from "@hookform/resolvers/zod";
import { useState } from "react";
import { useForm } from "react-hook-form";

import { Button } from "../../components/ui/Button";
import { Input } from "../../components/ui/Input";
import { useAdjustInventory, useInventoryMovements } from "../../hooks/useInventory";
import { useProducts } from "../../hooks/useProducts";
import {
  inventoryAdjustmentSchema,
  type InventoryAdjustmentFormInput,
  type InventoryAdjustmentFormValues,
} from "../../schemas/inventory";
import { ApiError } from "../../types/api";

const typeLabels: Record<string, string> = {
  in: "Masuk (+)",
  out: "Keluar (-)",
  correction: "Koreksi (nilai absolut)",
};

function AdjustmentForm({ onAdjusted }: { onAdjusted: (productId: number) => void }) {
  const { data: products } = useProducts({ limit: 100 });
  const adjustInventory = useAdjustInventory();
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<InventoryAdjustmentFormInput, unknown, InventoryAdjustmentFormValues>({
    resolver: zodResolver(inventoryAdjustmentSchema),
    defaultValues: { type: "in" },
  });

  function onSubmit(values: InventoryAdjustmentFormValues) {
    adjustInventory.mutate(
      {
        product_id: values.productId,
        type: values.type,
        quantity: values.quantity,
        reason: values.reason,
        reference: values.reference,
      },
      {
        onSuccess: (movement) => {
          onAdjusted(movement.product_id);
          reset({ type: values.type, productId: values.productId });
        },
      },
    );
  }

  return (
    <form
      onSubmit={handleSubmit(onSubmit)}
      className="grid grid-cols-2 gap-4 rounded-md border border-(--border-default) bg-(--surface-card) p-4"
      noValidate
    >
      <div className="flex flex-col gap-1">
        <label className="text-label-md text-(--ink-primary)" htmlFor="inventory-product">
          Produk
        </label>
        <select
          id="inventory-product"
          className="text-body-md rounded-md border border-(--border-default) px-3 py-2.5 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/30"
          {...register("productId")}
        >
          <option value="">Pilih produk</option>
          {products?.items.map((product) => (
            <option key={product.id} value={product.id}>
              {product.name} (stok saat ini: {product.stock})
            </option>
          ))}
        </select>
        {errors.productId && <p className="text-body-sm text-error-500">{errors.productId.message}</p>}
      </div>

      <div className="flex flex-col gap-1">
        <label className="text-label-md text-(--ink-primary)" htmlFor="inventory-type">
          Tipe Penyesuaian
        </label>
        <select
          id="inventory-type"
          className="text-body-md rounded-md border border-(--border-default) px-3 py-2.5 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/30"
          {...register("type")}
        >
          {Object.entries(typeLabels).map(([value, label]) => (
            <option key={value} value={value}>
              {label}
            </option>
          ))}
        </select>
      </div>

      <Input label="Kuantitas" type="number" error={errors.quantity?.message} {...register("quantity")} />
      <Input label="Referensi (opsional)" error={errors.reference?.message} {...register("reference")} />

      <div className="col-span-2">
        <Input label="Alasan" error={errors.reason?.message} {...register("reason")} />
      </div>

      <div className="col-span-2 flex items-center gap-3">
        <Button type="submit" disabled={adjustInventory.isPending}>
          {adjustInventory.isPending ? "Menyimpan..." : "Terapkan Penyesuaian"}
        </Button>
        {adjustInventory.isError && (
          <p role="alert" className="text-body-sm text-error-500">
            {adjustInventory.error instanceof ApiError ? adjustInventory.error.message : "Gagal menyesuaikan stok."}
          </p>
        )}
      </div>
    </form>
  );
}

function MovementHistory({ productId }: { productId: number }) {
  const [page, setPage] = useState(1);
  const { data, isLoading, isError } = useInventoryMovements(productId, page);

  if (!productId) {
    return <p className="text-body-md text-(--ink-secondary)">Pilih produk untuk melihat riwayat pergerakan stok.</p>;
  }

  return (
    <div>
      {isLoading && <p className="text-body-md text-(--ink-secondary)">Memuat riwayat...</p>}
      {isError && (
        <p role="alert" className="text-body-md text-error-500">
          Gagal memuat riwayat pergerakan stok.
        </p>
      )}
      {data && data.items.length === 0 && (
        <p className="text-body-md text-(--ink-secondary)">Belum ada riwayat pergerakan untuk produk ini.</p>
      )}
      {data && data.items.length > 0 && (
        <>
          <div className="overflow-x-auto rounded-md border border-(--border-default)">
            <table className="w-full text-left">
              <thead className="bg-neutral-50">
                <tr>
                  <th className="text-label-sm text-(--ink-secondary) px-4 py-3">Tanggal</th>
                  <th className="text-label-sm text-(--ink-secondary) px-4 py-3">Tipe</th>
                  <th className="text-label-sm text-(--ink-secondary) px-4 py-3">Sebelum</th>
                  <th className="text-label-sm text-(--ink-secondary) px-4 py-3">Sesudah</th>
                  <th className="text-label-sm text-(--ink-secondary) px-4 py-3">Alasan</th>
                </tr>
              </thead>
              <tbody>
                {data.items.map((movement) => (
                  <tr key={movement.id} className="border-t border-(--border-default)">
                    <td className="text-body-sm text-(--ink-secondary) px-4 py-3">
                      {new Date(movement.created_at).toLocaleString("id-ID", { dateStyle: "medium", timeStyle: "short" })}
                    </td>
                    <td className="text-body-sm text-(--ink-primary) px-4 py-3">{typeLabels[movement.type]}</td>
                    <td className="text-data-md text-(--ink-secondary) px-4 py-3">{movement.before_quantity}</td>
                    <td className="text-data-md text-(--ink-primary) px-4 py-3">{movement.after_quantity}</td>
                    <td className="text-body-sm text-(--ink-secondary) px-4 py-3">{movement.reason}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div className="mt-4 flex items-center gap-3">
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
    </div>
  );
}

export function AdminInventory() {
  const { data: products } = useProducts({ limit: 100 });
  const [selectedProductId, setSelectedProductId] = useState(0);

  return (
    <div className="mx-auto max-w-4xl">
      <h1 className="text-heading-xl text-(--ink-primary) mb-6">Penyesuaian Stok</h1>

      <AdjustmentForm onAdjusted={setSelectedProductId} />

      <div className="mt-10 flex items-center justify-between">
        <h2 className="text-heading-lg text-(--ink-primary)">Riwayat Pergerakan Stok</h2>
        <select
          aria-label="Lihat riwayat produk"
          className="text-body-md rounded-md border border-(--border-default) px-3 py-2 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/30"
          value={selectedProductId || ""}
          onChange={(e) => setSelectedProductId(Number(e.target.value))}
        >
          <option value="">Pilih produk</option>
          {products?.items.map((product) => (
            <option key={product.id} value={product.id}>
              {product.name}
            </option>
          ))}
        </select>
      </div>
      <div className="mt-4">
        <MovementHistory productId={selectedProductId} />
      </div>
    </div>
  );
}
