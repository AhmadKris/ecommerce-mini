import { zodResolver } from "@hookform/resolvers/zod";
import { useState } from "react";
import { useForm } from "react-hook-form";

import { Button } from "../../components/ui/Button";
import { Input } from "../../components/ui/Input";
import {
  useAddresses,
  useCreateAddress,
  useDeleteAddress,
  useUpdateAddress,
  type AddressInput,
} from "../../hooks/useAddresses";
import { useMe } from "../../hooks/useAuth";
import { addressSchema, type AddressFormValues } from "../../schemas/address";
import { ApiError } from "../../types/api";
import type { Address } from "../../types/address";

function toAddressInput(values: AddressFormValues): AddressInput {
  return {
    recipient_name: values.recipientName,
    phone: values.phone,
    address_line: values.addressLine,
    city: values.city,
    province: values.province,
    postal_code: values.postalCode,
    is_default: values.isDefault,
  };
}

function AddressForm({
  defaultValues,
  onSubmit,
  onCancel,
  isPending,
  submitLabel,
}: {
  defaultValues?: AddressFormValues;
  onSubmit: (values: AddressFormValues) => void;
  onCancel?: () => void;
  isPending: boolean;
  submitLabel: string;
}) {
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<AddressFormValues>({
    resolver: zodResolver(addressSchema),
    defaultValues: defaultValues ?? {
      recipientName: "",
      phone: "",
      addressLine: "",
      city: "",
      province: "",
      postalCode: "",
      isDefault: false,
    },
  });

  return (
    <form
      onSubmit={handleSubmit((values) => {
        onSubmit(values);
        if (!defaultValues) reset();
      })}
      className="grid grid-cols-2 gap-3"
      noValidate
    >
      <Input label="Nama Penerima" error={errors.recipientName?.message} {...register("recipientName")} />
      <Input label="Nomor Telepon" error={errors.phone?.message} {...register("phone")} />
      <div className="col-span-2">
        <Input label="Alamat Lengkap" error={errors.addressLine?.message} {...register("addressLine")} />
      </div>
      <Input label="Kota" error={errors.city?.message} {...register("city")} />
      <Input label="Provinsi" error={errors.province?.message} {...register("province")} />
      <Input label="Kode Pos" error={errors.postalCode?.message} {...register("postalCode")} />
      <label className="text-body-sm text-(--ink-primary) flex items-center gap-2 self-end pb-2.5">
        <input type="checkbox" {...register("isDefault")} />
        Jadikan alamat utama
      </label>
      <div className="col-span-2 flex items-center gap-3">
        <Button type="submit" disabled={isPending}>
          {isPending ? "Menyimpan..." : submitLabel}
        </Button>
        {onCancel && (
          <Button type="button" variant="secondary" onClick={onCancel}>
            Batal
          </Button>
        )}
      </div>
    </form>
  );
}

function AddressCard({ address }: { address: Address }) {
  const [isEditing, setIsEditing] = useState(false);
  const updateAddress = useUpdateAddress();
  const deleteAddress = useDeleteAddress();

  if (isEditing) {
    return (
      <li className="rounded-md border border-(--border-default) bg-(--surface-card) p-4">
        <AddressForm
          defaultValues={{
            recipientName: address.recipient_name,
            phone: address.phone,
            addressLine: address.address_line,
            city: address.city,
            province: address.province,
            postalCode: address.postal_code,
            isDefault: address.is_default,
          }}
          onSubmit={(values) =>
            updateAddress.mutate({ id: address.id, ...toAddressInput(values) }, { onSuccess: () => setIsEditing(false) })
          }
          onCancel={() => setIsEditing(false)}
          isPending={updateAddress.isPending}
          submitLabel="Simpan"
        />
        {updateAddress.isError && (
          <p role="alert" className="text-body-sm text-error-500 mt-2">
            {updateAddress.error instanceof ApiError ? updateAddress.error.message : "Gagal mengubah alamat."}
          </p>
        )}
      </li>
    );
  }

  return (
    <li className="rounded-md border border-(--border-default) bg-(--surface-card) p-4">
      <div className="flex items-start justify-between">
        <div>
          <p className="text-label-md text-(--ink-primary)">
            {address.recipient_name}
            {address.is_default && (
              <span className="text-label-sm ml-2 rounded-full bg-success-100 px-2 py-0.5 text-success-500">
                Utama
              </span>
            )}
          </p>
          <p className="text-body-sm text-(--ink-secondary) mt-1">{address.phone}</p>
          <p className="text-body-sm text-(--ink-secondary) mt-1">
            {address.address_line}, {address.city}, {address.province} {address.postal_code}
          </p>
        </div>
        <div className="flex shrink-0 items-center gap-3">
          <button type="button" className="text-label-sm text-(--ink-link)" onClick={() => setIsEditing(true)}>
            Ubah
          </button>
          <button
            type="button"
            className="text-label-sm text-error-500"
            onClick={() => deleteAddress.mutate(address.id)}
            disabled={deleteAddress.isPending}
          >
            Hapus
          </button>
        </div>
      </div>
      {deleteAddress.isError && deleteAddress.variables === address.id && (
        <p role="alert" className="text-body-sm text-error-500 mt-2">
          {deleteAddress.error instanceof ApiError ? deleteAddress.error.message : "Gagal menghapus alamat."}
        </p>
      )}
    </li>
  );
}

export function Profile() {
  const { data: user, isLoading: isLoadingUser } = useMe();
  const { data: addresses, isLoading: isLoadingAddresses, isError } = useAddresses();
  const createAddress = useCreateAddress();
  const [isAdding, setIsAdding] = useState(false);

  return (
    <main className="mx-auto max-w-2xl px-6 py-16">
      <h1 className="text-heading-xl text-(--ink-primary)">Profil</h1>

      <div className="mt-6 rounded-md border border-(--border-default) bg-(--surface-card) p-4">
        {isLoadingUser ? (
          <p className="text-body-md text-(--ink-secondary)">Memuat profil...</p>
        ) : (
          <>
            <p className="text-label-sm text-(--ink-secondary)">Nama</p>
            <p className="text-body-md text-(--ink-primary) mt-1">{user?.name}</p>
            <p className="text-label-sm text-(--ink-secondary) mt-4">Email</p>
            <p className="text-body-md text-(--ink-primary) mt-1">{user?.email}</p>
          </>
        )}
      </div>

      <div className="mt-10 flex items-center justify-between">
        <h2 className="text-heading-lg text-(--ink-primary)">Buku Alamat</h2>
        {!isAdding && (
          <Button variant="secondary" onClick={() => setIsAdding(true)}>
            Tambah Alamat
          </Button>
        )}
      </div>

      {isAdding && (
        <div className="mt-4 rounded-md border border-(--border-default) bg-(--surface-card) p-4">
          <AddressForm
            onSubmit={(values) =>
              createAddress.mutate(toAddressInput(values), { onSuccess: () => setIsAdding(false) })
            }
            onCancel={() => setIsAdding(false)}
            isPending={createAddress.isPending}
            submitLabel="Simpan Alamat"
          />
          {createAddress.isError && (
            <p role="alert" className="text-body-sm text-error-500 mt-2">
              {createAddress.error instanceof ApiError ? createAddress.error.message : "Gagal menyimpan alamat."}
            </p>
          )}
        </div>
      )}

      <div className="mt-4">
        {isLoadingAddresses && <p className="text-body-md text-(--ink-secondary)">Memuat alamat...</p>}
        {isError && (
          <p role="alert" className="text-body-md text-error-500">
            Gagal memuat buku alamat.
          </p>
        )}
        {addresses && addresses.length === 0 && !isAdding && (
          <p className="text-body-md text-(--ink-secondary)">Belum ada alamat tersimpan.</p>
        )}
        {addresses && addresses.length > 0 && (
          <ul className="flex flex-col gap-3">
            {addresses.map((address) => (
              <AddressCard key={address.id} address={address} />
            ))}
          </ul>
        )}
      </div>
    </main>
  );
}
