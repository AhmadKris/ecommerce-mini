import { zodResolver } from "@hookform/resolvers/zod";
import { useState } from "react";
import { useForm } from "react-hook-form";

import { Button } from "../../components/ui/Button";
import { Input } from "../../components/ui/Input";
import { useCategories, useCreateCategory, useDeleteCategory, useUpdateCategory } from "../../hooks/useCategories";
import { categorySchema, type CategoryFormValues } from "../../schemas/category";
import { ApiError } from "../../types/api";
import type { Category } from "../../types/product";

function CreateCategoryForm() {
  const createCategory = useCreateCategory();
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CategoryFormValues>({ resolver: zodResolver(categorySchema) });

  function onSubmit(values: CategoryFormValues) {
    createCategory.mutate(values.name, { onSuccess: () => reset() });
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="flex items-end gap-3" noValidate>
      <div className="w-64">
        <Input label="Kategori baru" error={errors.name?.message} {...register("name")} />
      </div>
      <Button type="submit" disabled={createCategory.isPending}>
        {createCategory.isPending ? "Menyimpan..." : "Tambah"}
      </Button>
      {createCategory.isError && (
        <p role="alert" className="text-body-sm text-error-500">
          {createCategory.error instanceof ApiError ? createCategory.error.message : "Gagal membuat kategori."}
        </p>
      )}
    </form>
  );
}

function CategoryRow({ category }: { category: Category }) {
  const [isEditing, setIsEditing] = useState(false);
  const updateCategory = useUpdateCategory();
  const deleteCategory = useDeleteCategory();
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<CategoryFormValues>({
    resolver: zodResolver(categorySchema),
    defaultValues: { name: category.name },
  });

  function onSubmit(values: CategoryFormValues) {
    updateCategory.mutate({ id: category.id, name: values.name }, { onSuccess: () => setIsEditing(false) });
  }

  if (isEditing) {
    return (
      <tr className="border-t border-(--border-default)">
        <td colSpan={2} className="px-4 py-3">
          <form onSubmit={handleSubmit(onSubmit)} className="flex items-center gap-2" noValidate>
            <Input label="" aria-label="Nama kategori" error={errors.name?.message} {...register("name")} />
            <Button type="submit" disabled={updateCategory.isPending}>
              Simpan
            </Button>
            <Button type="button" variant="secondary" onClick={() => setIsEditing(false)}>
              Batal
            </Button>
          </form>
          {updateCategory.isError && (
            <p role="alert" className="text-body-sm text-error-500 mt-1">
              {updateCategory.error instanceof ApiError ? updateCategory.error.message : "Gagal mengubah kategori."}
            </p>
          )}
        </td>
      </tr>
    );
  }

  return (
    <tr className="border-t border-(--border-default)">
      <td className="text-body-sm text-(--ink-primary) px-4 py-3">{category.name}</td>
      <td className="px-4 py-3">
        <div className="flex items-center gap-3">
          <button
            type="button"
            className="text-label-sm text-(--ink-link)"
            onClick={() => setIsEditing(true)}
          >
            Ubah
          </button>
          <button
            type="button"
            className="text-label-sm text-error-500"
            onClick={() => deleteCategory.mutate(category.id)}
            disabled={deleteCategory.isPending}
          >
            Hapus
          </button>
        </div>
        {deleteCategory.isError && deleteCategory.variables === category.id && (
          <p role="alert" className="text-body-sm text-error-500 mt-1">
            {deleteCategory.error instanceof ApiError ? deleteCategory.error.message : "Gagal menghapus kategori."}
          </p>
        )}
      </td>
    </tr>
  );
}

export function AdminCategories() {
  const { data: categories, isLoading, isError } = useCategories();

  return (
    <div className="mx-auto max-w-3xl">
      <h1 className="text-heading-xl text-(--ink-primary) mb-6">Kelola Kategori</h1>

      <CreateCategoryForm />

      <div className="mt-8">
        {isLoading && <p className="text-body-md text-(--ink-secondary)">Memuat kategori...</p>}
        {isError && (
          <p role="alert" className="text-body-md text-error-500">
            Gagal memuat kategori.
          </p>
        )}
        {categories && categories.length > 0 && (
          <table className="w-full text-left">
            <tbody>
              {categories.map((category) => (
                <CategoryRow key={category.id} category={category} />
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
