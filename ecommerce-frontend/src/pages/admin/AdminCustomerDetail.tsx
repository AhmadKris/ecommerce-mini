import { Link, Navigate, useParams } from "react-router-dom";

import { useAdminCustomer } from "../../hooks/useAdminCustomers";
import { usePermission } from "../../hooks/usePermission";

export function AdminCustomerDetail() {
  const canReadCustomers = usePermission("customer:read");
  const { id = "" } = useParams<{ id: string }>();
  const customerId = Number(id);
  const { data: customer, isLoading, isError } = useAdminCustomer(customerId);

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

      <div className="mt-4 rounded-md border border-(--border-default) bg-(--surface-card) p-4">
        <p className="text-label-sm text-(--ink-secondary) mb-3">Role & Permission</p>
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
    </div>
  );
}
