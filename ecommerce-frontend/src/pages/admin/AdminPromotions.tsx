import { zodResolver } from "@hookform/resolvers/zod";
import { useState } from "react";
import { useForm } from "react-hook-form";

import { Button } from "../../components/ui/Button";
import { Input } from "../../components/ui/Input";
import {
  useCreatePromotion,
  useDeletePromotion,
  usePromotions,
  useUpdatePromotion,
} from "../../hooks/usePromotions";
import { formatCurrency } from "../../lib/format";
import {
  promotionSchema,
  promotionUpdateSchema,
  type PromotionFormInput,
  type PromotionFormValues,
  type PromotionUpdateFormInput,
  type PromotionUpdateFormValues,
} from "../../schemas/promotion";
import { ApiError } from "../../types/api";
import type { Promotion } from "../../types/promotion";

/** datetime-local inputs need "YYYY-MM-DDTHH:mm", not a full ISO timestamp. */
function toDatetimeLocal(isoString: string): string {
  return isoString.slice(0, 16);
}

function CreatePromotionForm() {
  const createPromotion = useCreatePromotion();
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<PromotionFormInput, unknown, PromotionFormValues>({ resolver: zodResolver(promotionSchema) });

  function onSubmit(values: PromotionFormValues) {
    createPromotion.mutate(
      { ...values, starts_at: new Date(values.starts_at).toISOString(), ends_at: new Date(values.ends_at).toISOString() },
      { onSuccess: () => reset() },
    );
  }

  return (
    <form
      onSubmit={handleSubmit(onSubmit)}
      className="grid grid-cols-2 gap-4 rounded-md border border-(--border-default) bg-(--surface-card) p-4 sm:grid-cols-3"
      noValidate
    >
      <Input label="Kode" error={errors.code?.message} {...register("code")} />
      <div className="flex flex-col gap-1">
        <label className="text-label-md text-(--ink-primary)" htmlFor="promo-type">
          Tipe
        </label>
        <select
          id="promo-type"
          className="text-body-md rounded-md border border-(--border-default) px-3 py-2.5 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/30"
          {...register("type")}
        >
          <option value="percentage">Persentase</option>
          <option value="fixed">Potongan Tetap</option>
        </select>
      </div>
      <Input label="Nilai" type="number" step="any" error={errors.value?.message} {...register("value")} />
      <Input
        label="Minimum Pembelian"
        type="number"
        step="any"
        error={errors.minimum_purchase?.message}
        {...register("minimum_purchase")}
      />
      <Input
        label="Batas Pemakaian"
        type="number"
        error={errors.usage_limit?.message}
        {...register("usage_limit")}
      />
      <Input label="Mulai" type="datetime-local" error={errors.starts_at?.message} {...register("starts_at")} />
      <Input label="Berakhir" type="datetime-local" error={errors.ends_at?.message} {...register("ends_at")} />
      <div className="col-span-full flex items-center gap-3">
        <Button type="submit" disabled={createPromotion.isPending}>
          {createPromotion.isPending ? "Menyimpan..." : "Buat Promo"}
        </Button>
        {createPromotion.isError && (
          <p role="alert" className="text-body-sm text-error-500">
            {createPromotion.error instanceof ApiError ? createPromotion.error.message : "Gagal membuat promo."}
          </p>
        )}
      </div>
    </form>
  );
}

function EditPromotionForm({ promotion, onDone }: { promotion: Promotion; onDone: () => void }) {
  const updatePromotion = useUpdatePromotion();
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<PromotionUpdateFormInput, unknown, PromotionUpdateFormValues>({
    resolver: zodResolver(promotionUpdateSchema),
    defaultValues: {
      type: promotion.type,
      value: promotion.value,
      minimum_purchase: promotion.minimum_purchase,
      usage_limit: promotion.usage_limit,
      starts_at: toDatetimeLocal(promotion.starts_at),
      ends_at: toDatetimeLocal(promotion.ends_at),
      status: promotion.status,
    },
  });

  function onSubmit(values: PromotionUpdateFormValues) {
    updatePromotion.mutate(
      {
        id: promotion.id,
        ...values,
        starts_at: new Date(values.starts_at).toISOString(),
        ends_at: new Date(values.ends_at).toISOString(),
      },
      { onSuccess: onDone },
    );
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="grid grid-cols-2 gap-3 sm:grid-cols-3" noValidate>
      <div className="flex flex-col gap-1">
        <label className="text-label-md text-(--ink-primary)" htmlFor={`type-${promotion.id}`}>
          Tipe
        </label>
        <select
          id={`type-${promotion.id}`}
          className="text-body-md rounded-md border border-(--border-default) px-3 py-2.5 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/30"
          {...register("type")}
        >
          <option value="percentage">Persentase</option>
          <option value="fixed">Potongan Tetap</option>
        </select>
      </div>
      <Input label="Nilai" type="number" step="any" error={errors.value?.message} {...register("value")} />
      <Input
        label="Minimum Pembelian"
        type="number"
        step="any"
        error={errors.minimum_purchase?.message}
        {...register("minimum_purchase")}
      />
      <Input
        label="Batas Pemakaian"
        type="number"
        error={errors.usage_limit?.message}
        {...register("usage_limit")}
      />
      <Input label="Mulai" type="datetime-local" error={errors.starts_at?.message} {...register("starts_at")} />
      <Input label="Berakhir" type="datetime-local" error={errors.ends_at?.message} {...register("ends_at")} />
      <div className="flex flex-col gap-1">
        <label className="text-label-md text-(--ink-primary)" htmlFor={`status-${promotion.id}`}>
          Status
        </label>
        <select
          id={`status-${promotion.id}`}
          className="text-body-md rounded-md border border-(--border-default) px-3 py-2.5 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/30"
          {...register("status")}
        >
          <option value="active">Aktif</option>
          <option value="inactive">Nonaktif</option>
        </select>
      </div>
      <div className="col-span-full flex items-center gap-3">
        <Button type="submit" disabled={updatePromotion.isPending}>
          Simpan
        </Button>
        <Button type="button" variant="secondary" onClick={onDone}>
          Batal
        </Button>
        {updatePromotion.isError && (
          <p role="alert" className="text-body-sm text-error-500">
            {updatePromotion.error instanceof ApiError ? updatePromotion.error.message : "Gagal mengubah promo."}
          </p>
        )}
      </div>
    </form>
  );
}

function PromotionRow({ promotion }: { promotion: Promotion }) {
  const [isEditing, setIsEditing] = useState(false);
  const deletePromotion = useDeletePromotion();

  if (isEditing) {
    return (
      <tr className="border-t border-(--border-default)">
        <td colSpan={6} className="px-4 py-3">
          <EditPromotionForm promotion={promotion} onDone={() => setIsEditing(false)} />
        </td>
      </tr>
    );
  }

  return (
    <tr className="border-t border-(--border-default)">
      <td className="text-data-md text-(--ink-primary) px-4 py-3">{promotion.code}</td>
      <td className="text-body-sm text-(--ink-secondary) px-4 py-3">
        {promotion.type === "percentage" ? `${promotion.value}%` : formatCurrency(promotion.value)}
      </td>
      <td className="text-body-sm text-(--ink-secondary) px-4 py-3">
        {promotion.used_count} / {promotion.usage_limit}
      </td>
      <td className="px-4 py-3">
        <span
          className={`text-label-sm rounded-full px-2.5 py-1 ${
            promotion.status === "active" ? "bg-success-100 text-success-500" : "bg-neutral-100 text-(--ink-secondary)"
          }`}
        >
          {promotion.status === "active" ? "Aktif" : "Nonaktif"}
        </span>
      </td>
      <td className="text-body-sm text-(--ink-secondary) px-4 py-3">
        {new Date(promotion.ends_at).toLocaleDateString("id-ID", { dateStyle: "medium" })}
      </td>
      <td className="px-4 py-3">
        <div className="flex items-center gap-3">
          <button type="button" className="text-label-sm text-(--ink-link)" onClick={() => setIsEditing(true)}>
            Ubah
          </button>
          <button
            type="button"
            className="text-label-sm text-error-500"
            onClick={() => deletePromotion.mutate(promotion.id)}
            disabled={deletePromotion.isPending}
          >
            Hapus
          </button>
        </div>
        {deletePromotion.isError && deletePromotion.variables === promotion.id && (
          <p role="alert" className="text-body-sm text-error-500 mt-1">
            {deletePromotion.error instanceof ApiError ? deletePromotion.error.message : "Gagal menghapus promo."}
          </p>
        )}
      </td>
    </tr>
  );
}

export function AdminPromotions() {
  const { data: promotions, isLoading, isError } = usePromotions();

  return (
    <div className="mx-auto max-w-4xl">
      <h1 className="text-heading-xl text-(--ink-primary) mb-6">Kelola Promo</h1>

      <CreatePromotionForm />

      <div className="mt-8">
        {isLoading && <p className="text-body-md text-(--ink-secondary)">Memuat promo...</p>}
        {isError && (
          <p role="alert" className="text-body-md text-error-500">
            Gagal memuat daftar promo.
          </p>
        )}
        {promotions && promotions.length === 0 && (
          <p className="text-body-md text-(--ink-secondary)">Belum ada promo.</p>
        )}
        {promotions && promotions.length > 0 && (
          <div className="overflow-x-auto rounded-md border border-(--border-default)">
            <table className="w-full text-left">
              <thead className="bg-neutral-50">
                <tr>
                  <th className="text-label-sm text-(--ink-secondary) px-4 py-3">Kode</th>
                  <th className="text-label-sm text-(--ink-secondary) px-4 py-3">Diskon</th>
                  <th className="text-label-sm text-(--ink-secondary) px-4 py-3">Pemakaian</th>
                  <th className="text-label-sm text-(--ink-secondary) px-4 py-3">Status</th>
                  <th className="text-label-sm text-(--ink-secondary) px-4 py-3">Berakhir</th>
                  <th className="text-label-sm text-(--ink-secondary) px-4 py-3">Aksi</th>
                </tr>
              </thead>
              <tbody>
                {promotions.map((promotion) => (
                  <PromotionRow key={promotion.id} promotion={promotion} />
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
