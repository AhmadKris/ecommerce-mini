import { useState } from "react";
import { Link, Navigate, useParams } from "react-router-dom";

import { Button } from "../../components/ui/Button";
import { useAdminCustomer } from "../../hooks/useAdminCustomers";
import { usePermission } from "../../hooks/usePermission";
import { useRoles, useUpdateUserRoles } from "../../hooks/useRoles";
import { useAuthStore } from "../../store/auth-store";
import { ApiError } from "../../types/api";

export function AdminCustomerDetail() {
  const canReadCustomers = usePermission("customer:read");
  const canManageRoles = usePermission("user:manage");
  const { id = "" } = useParams<{ id: string }>();
  const customerId = Number(id);
  const { data: customer, isLoading, isError } = useAdminCustomer(customerId);
  const { data: availableRoles } = useRoles();
  const updateRoles = useUpdateUserRoles(customerId);
  const currentUserId = useAuthStore((s) => s.userId);

  // Track which role names are checked in the management form. Initialised
  // lazily from the customer's current roles when they first load.
  const [selectedRoles, setSelectedRoles] = useState<Set<string> | null>(null);
  const [saveError, setSaveError] = useState<string | null>(null);
  const [saveSuccess, setSaveSuccess] = useState(false);

  // Initialise the selection once customer data arrives (only once — the user
  // owns the checkbox state from that point on).
  if (customer && selectedRoles === null) {
    setSelectedRoles(new Set(customer.roles.map((r) => r.name)));
  }

  const isSelf = currentUserId === customerId;

  function handleToggleRole(roleName: string) {
    setSelectedRoles((prev) => {
      if (!prev) return prev;
      const next = new Set(prev);
      if (next.has(roleName)) {
        next.delete(roleName);
      } else {
        next.add(roleName);
      }
      return next;
    });
    setSaveError(null);
    setSaveSuccess(false);
  }

  function handleSaveRoles() {
    if (!selectedRoles) return;
    setSaveError(null);
    setSaveSuccess(false);
    updateRoles.mutate(Array.from(selectedRoles), {
      onSuccess: () => setSaveSuccess(true),
      onError: (err) => {
        setSaveError(err instanceof ApiError ? err.message : "Gagal memperbarui role.");
      },
    });
  }

  if (!canReadCustomers) {
    return <Navigate to="/admin" replace />;
  }

  if (isLoading) {
    return (
      <div className="mx-auto max-w-3xl">
        <p className="text-body-md text-(--ink-secondary)">Memuat pelanggan...</p>
      </div>
    );
  }

  if (isError || !customer) {
    return (
      <div className="mx-auto max-w-3xl">
        <p role="alert" className="text-body-md text-error-500">
          Pelanggan tidak ditemukan.
        </p>
        <Link to="/admin/customers" className="text-body-sm text-(--ink-link) mt-2 inline-block">
          Kembali ke daftar pelanggan
        </Link>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-3xl">
      <Link to="/admin/customers" className="text-body-sm text-(--ink-link)">
        ← Kembali ke daftar pelanggan
      </Link>

      <h1 className="text-heading-xl text-(--ink-primary) mt-4">{customer.name}</h1>

      <div className="mt-6 grid gap-4 sm:grid-cols-2">
        <div className="rounded-md border border-(--border-default) bg-(--surface-card) p-4">
          <p className="text-label-sm text-(--ink-secondary)">Email</p>
          <p className="text-body-md text-(--ink-primary) mt-1">{customer.email}</p>
        </div>
        <div className="rounded-md border border-(--border-default) bg-(--surface-card) p-4">
          <p className="text-label-sm text-(--ink-secondary)">Terdaftar</p>
          <p className="text-body-md text-(--ink-primary) mt-1">
            {new Date(customer.created_at).toLocaleDateString("id-ID", { dateStyle: "long" })}
          </p>
        </div>
      </div>

      {/* Role & Permission view (always visible to customer:read) */}
      <div className="mt-4 rounded-md border border-(--border-default) bg-(--surface-card) p-4">
        <p className="text-label-sm text-(--ink-secondary) mb-3">Role &amp; Permission</p>
        {customer.roles.length === 0 ? (
          <p className="text-body-sm text-(--ink-secondary)">Tidak ada role.</p>
        ) : (
          <div className="flex flex-col gap-3">
            {customer.roles.map((role) => (
              <div key={role.id}>
                <span className="text-label-sm rounded-full bg-warning-100 px-2.5 py-1 text-warning-500">
                  {role.name}
                </span>
                {role.permissions && role.permissions.length > 0 && (
                  <div className="text-body-sm text-(--ink-secondary) mt-2 flex flex-wrap gap-2">
                    {role.permissions.map((permission) => (
                      <span
                        key={permission.id}
                        className="text-label-sm rounded-full border border-(--border-default) px-2.5 py-1"
                      >
                        {permission.code}
                      </span>
                    ))}
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Role management section — only shown if caller has user:manage */}
      {canManageRoles && availableRoles && selectedRoles !== null && (
        <div className="mt-4 rounded-md border border-(--border-default) bg-(--surface-card) p-4">
          <p className="text-label-sm text-(--ink-secondary) mb-3">Manajemen Role</p>

          {isSelf && (
            <p className="text-body-sm text-warning-500 mb-3">
              Anda tidak dapat menghapus role <strong>admin</strong> dari akun Anda sendiri.
            </p>
          )}

          <div className="flex flex-col gap-2">
            {availableRoles.map((role) => {
              // Prevent self-lockout: disable the admin checkbox for own account.
              const isDisabled = isSelf && role.name === "admin";
              return (
                <label
                  key={role.id}
                  className={`flex items-center gap-3 ${isDisabled ? "cursor-not-allowed opacity-50" : "cursor-pointer"}`}
                >
                  <input
                    type="checkbox"
                    id={`role-${role.id}`}
                    checked={selectedRoles.has(role.name)}
                    disabled={isDisabled}
                    onChange={() => handleToggleRole(role.name)}
                    className="h-4 w-4 accent-(--color-primary-600)"
                  />
                  <span className="text-body-md text-(--ink-primary)">{role.name}</span>
                  {role.permissions && role.permissions.length > 0 && (
                    <span className="text-body-sm text-(--ink-tertiary)">
                      ({role.permissions.length} permission)
                    </span>
                  )}
                </label>
              );
            })}
          </div>

          {saveError && (
            <p role="alert" className="text-body-sm text-error-500 mt-3">
              {saveError}
            </p>
          )}
          {saveSuccess && (
            <p role="status" className="text-body-sm text-success-600 mt-3">
              Role berhasil diperbarui.
            </p>
          )}

          <Button
            className="mt-4"
            disabled={updateRoles.isPending || selectedRoles.size === 0}
            onClick={handleSaveRoles}
          >
            {updateRoles.isPending ? "Menyimpan..." : "Simpan Role"}
          </Button>
        </div>
      )}
    </div>
  );
}
