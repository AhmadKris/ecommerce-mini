import { useState } from "react";
import { Link, Navigate, useSearchParams } from "react-router-dom";

import { Input } from "../../components/ui/Input";
import { Button } from "../../components/ui/Button";
import { useAdminCustomers } from "../../hooks/useAdminCustomers";
import { usePermission } from "../../hooks/usePermission";
import { useDebounce } from "../../hooks/useDebounce";

export function AdminCustomers() {
  const canReadCustomers = usePermission("customer:read");
  const [searchParams, setSearchParams] = useSearchParams();
  const page = Number(searchParams.get("page") ?? "1");
  const [searchInput, setSearchInput] = useState(searchParams.get("search") ?? "");
  const debouncedSearch = useDebounce(searchInput, 300);
  const { data, isLoading, isError } = useAdminCustomers(page, debouncedSearch || undefined);

  if (!canReadCustomers) {
    return <Navigate to="/admin" replace />;
  }

  function goToPage(nextPage: number) {
    const next = new URLSearchParams(searchParams);
    next.set("page", String(nextPage));
    setSearchParams(next);
  }

  function handleSearchChange(value: string) {
    setSearchInput(value);
    const next = new URLSearchParams(searchParams);
    if (value) {
      next.set("search", value);
    } else {
      next.delete("search");
    }
    next.delete("page");
    setSearchParams(next);
  }

  return (
    <div className="mx-auto max-w-4xl">
      <h1 className="text-heading-xl text-(--ink-primary) mb-6">Pelanggan</h1>

      <div className="mb-4 max-w-xs">
        <Input
          label="Cari pelanggan"
          value={searchInput}
          onChange={(event) => handleSearchChange(event.target.value)}
          placeholder="Nama atau email..."
        />
      </div>

      {isLoading && <p className="text-body-md text-(--ink-secondary)">Memuat pelanggan...</p>}
      {isError && (
        <p role="alert" className="text-body-md text-error-500">
          Gagal memuat daftar pelanggan.
        </p>
      )}
      {data && data.items.length === 0 && (
        <p className="text-body-md text-(--ink-secondary)">Belum ada pelanggan.</p>
      )}

      {data && data.items.length > 0 && (
        <>
          <div className="overflow-x-auto rounded-md border border-(--border-default)">
            <table className="w-full text-left">
              <thead className="bg-neutral-50">
                <tr>
                  <th className="text-label-sm text-(--ink-secondary) px-4 py-3">Nama</th>
                  <th className="text-label-sm text-(--ink-secondary) px-4 py-3">Email</th>
                  <th className="text-label-sm text-(--ink-secondary) px-4 py-3">Role</th>
                  <th className="text-label-sm text-(--ink-secondary) px-4 py-3">Terdaftar</th>
                </tr>
              </thead>
              <tbody>
                {data.items.map((customer) => (
                  <tr key={customer.id} className="border-t border-(--border-default)">
                    <td className="text-body-sm px-4 py-3">
                      <Link to={`/admin/customers/${customer.id}`} className="text-(--ink-link)">
                        {customer.name}
                      </Link>
                    </td>
                    <td className="text-body-sm text-(--ink-secondary) px-4 py-3">{customer.email}</td>
                    <td className="text-body-sm text-(--ink-secondary) px-4 py-3">
                      {customer.roles.map((role) => role.name).join(", ") || "—"}
                    </td>
                    <td className="text-body-sm text-(--ink-secondary) px-4 py-3">
                      {new Date(customer.created_at).toLocaleDateString("id-ID", { dateStyle: "medium" })}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div className="mt-6 flex items-center gap-3">
            <Button variant="secondary" onClick={() => goToPage(page - 1)} disabled={page <= 1}>
              Sebelumnya
            </Button>
            <span className="text-body-sm text-(--ink-secondary)">
              Halaman {data.meta.page} dari {data.meta.total_pages}
            </span>
            <Button
              variant="secondary"
              onClick={() => goToPage(page + 1)}
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
