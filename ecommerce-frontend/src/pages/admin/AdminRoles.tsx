import { useState } from "react";
import { Navigate, useSearchParams } from "react-router-dom";

import { Button } from "../../components/ui/Button";
import { Input } from "../../components/ui/Input";
import { useAdminCustomers } from "../../hooks/useAdminCustomers";
import { useDebounce } from "../../hooks/useDebounce";
import { usePermission } from "../../hooks/usePermission";
import { useRoles, useUpdateUserRoles } from "../../hooks/useRoles";
import { useAuthStore } from "../../store/auth-store";
import type { Customer } from "../../types/customer";

export function AdminRoles() {
  const canManageRoles = usePermission("user:manage");
  const currentUserId = useAuthStore((state) => state.userId);

  const { data: roles, isLoading: isLoadingRoles, isError: isErrorRoles } = useRoles();

  const [searchParams, setSearchParams] = useSearchParams();
  const page = Number(searchParams.get("page") ?? "1");
  const [searchInput, setSearchInput] = useState(searchParams.get("search") ?? "");
  const debouncedSearch = useDebounce(searchInput, 300);

  const {
    data: customersData,
    isLoading: isLoadingCustomers,
    isError: isErrorCustomers,
  } = useAdminCustomers(page, debouncedSearch || undefined);

  // Selected user for role management modal/panel
  const [selectedUser, setSelectedUser] = useState<Customer | null>(null);
  const [selectedRoleNames, setSelectedRoleNames] = useState<string[]>([]);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  const updateRolesMutation = useUpdateUserRoles();

  if (!canManageRoles) {
    return <Navigate to="/admin" replace />;
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

  function goToPage(nextPage: number) {
    const next = new URLSearchParams(searchParams);
    next.set("page", String(nextPage));
    setSearchParams(next);
  }

  function openEditModal(user: Customer) {
    setSelectedUser(user);
    setSelectedRoleNames(user.roles.map((r) => r.name));
    setSuccessMessage(null);
  }

  function closeModal() {
    setSelectedUser(null);
    setSuccessMessage(null);
  }

  function toggleRole(roleName: string) {
    if (!selectedUser) return;
    const isSelf = selectedUser.id === currentUserId;
    if (isSelf && roleName === "admin" && selectedRoleNames.includes("admin")) {
      return; // Self-lockout protection in UI
    }

    if (selectedRoleNames.includes(roleName)) {
      setSelectedRoleNames(selectedRoleNames.filter((r) => r !== roleName));
    } else {
      setSelectedRoleNames([...selectedRoleNames, roleName]);
    }
  }

  function handleSaveRoles() {
    if (!selectedUser) return;
    setSuccessMessage(null);

    updateRolesMutation.mutate(
      {
        userId: selectedUser.id,
        roleNames: selectedRoleNames,
      },
      {
        onSuccess: (updatedCustomer) => {
          setSuccessMessage(`Role untuk ${updatedCustomer.name} berhasil diperbarui.`);
          // Update selected user local reference
          setSelectedUser((prev) => (prev ? { ...prev, roles: updatedCustomer.roles } : null));
        },
      },
    );
  }

  return (
    <div className="mx-auto max-w-5xl">
      <h1 className="text-heading-xl text-(--ink-primary) mb-2">Manajemen Role & Akses</h1>
      <p className="text-body-md text-(--ink-secondary) mb-8">
        Kelola hak akses sistem, daftar role yang tersedia, serta alokasi role untuk setiap pengguna.
      </p>

      {/* Section 1: Daftar Role Sistem */}
      <section className="mb-10">
        <h2 className="text-heading-md text-(--ink-primary) mb-4">Daftar Role & Permission Sistem</h2>

        {isLoadingRoles && <p className="text-body-sm text-(--ink-secondary)">Memuat daftar role...</p>}
        {isErrorRoles && (
          <p role="alert" className="text-body-sm text-error-500">
            Gagal memuat daftar role sistem.
          </p>
        )}

        {roles && (
          <div className="grid gap-4 sm:grid-cols-2">
            {roles.map((role) => (
              <div
                key={role.id}
                className="rounded-md border border-(--border-default) bg-(--surface-card) p-4 shadow-2xs"
              >
                <div className="flex items-center justify-between mb-2">
                  <span className="text-label-md text-(--ink-primary) font-semibold uppercase tracking-wider">
                    {role.name}
                  </span>
                  <span className="text-label-xs rounded-full bg-neutral-100 px-2 py-0.5 text-(--ink-tertiary)">
                    ID: {role.id}
                  </span>
                </div>
                <p className="text-label-xs text-(--ink-tertiary) mb-2">Permissions ({role.permissions.length}):</p>
                <div className="flex flex-wrap gap-1.5">
                  {role.permissions.map((perm) => (
                    <span
                      key={perm.id}
                      className="text-label-xs rounded-md border border-(--border-default) bg-neutral-50 px-2 py-0.5 text-(--ink-secondary)"
                    >
                      {perm.code}
                    </span>
                  ))}
                  {role.permissions.length === 0 && (
                    <span className="text-body-xs text-(--ink-tertiary)">Tidak ada permission</span>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </section>

      {/* Section 2: Alokasi Akses Pengguna */}
      <section>
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-4">
          <h2 className="text-heading-md text-(--ink-primary)">Kelola Akses Pengguna</h2>
          <div className="w-full sm:w-72">
            <Input
              label=""
              value={searchInput}
              onChange={(e) => handleSearchChange(e.target.value)}
              placeholder="Cari pengguna..."
            />
          </div>
        </div>

        {isLoadingCustomers && <p className="text-body-sm text-(--ink-secondary)">Memuat data pengguna...</p>}
        {isErrorCustomers && (
          <p role="alert" className="text-body-sm text-error-500">
            Gagal memuat daftar pengguna.
          </p>
        )}

        {customersData && customersData.items.length === 0 && (
          <p className="text-body-sm text-(--ink-secondary)">Pengguna tidak ditemukan.</p>
        )}

        {customersData && customersData.items.length > 0 && (
          <>
            <div className="overflow-x-auto rounded-md border border-(--border-default) bg-(--surface-card)">
              <table className="w-full text-left border-collapse">
                <thead className="bg-neutral-50 border-b border-(--border-default)">
                  <tr>
                    <th className="text-label-sm text-(--ink-secondary) px-4 py-3">Nama</th>
                    <th className="text-label-sm text-(--ink-secondary) px-4 py-3">Email</th>
                    <th className="text-label-sm text-(--ink-secondary) px-4 py-3">Role Saat Ini</th>
                    <th className="text-label-sm text-(--ink-secondary) px-4 py-3 text-right">Aksi</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-(--border-default)">
                  {customersData.items.map((user) => (
                    <tr key={user.id} className="hover:bg-neutral-50/50 transition-colors">
                      <td className="text-body-sm font-medium text-(--ink-primary) px-4 py-3">
                        {user.name}
                        {user.id === currentUserId && (
                          <span className="ml-2 text-label-xs rounded bg-primary-100 px-1.5 py-0.5 text-primary-700">
                            (Anda)
                          </span>
                        )}
                      </td>
                      <td className="text-body-sm text-(--ink-secondary) px-4 py-3">{user.email}</td>
                      <td className="text-body-sm px-4 py-3">
                        <div className="flex flex-wrap gap-1">
                          {user.roles.map((r) => (
                            <span
                              key={r.id}
                              className={`text-label-xs rounded-full px-2 py-0.5 ${
                                r.name === "admin"
                                  ? "bg-warning-100 text-warning-700"
                                  : "bg-neutral-100 text-(--ink-secondary)"
                              }`}
                            >
                              {r.name}
                            </span>
                          ))}
                          {user.roles.length === 0 && (
                            <span className="text-body-xs text-(--ink-tertiary)">Tanpa role</span>
                          )}
                        </div>
                      </td>
                      <td className="text-body-sm px-4 py-3 text-right">
                        <Button variant="secondary" onClick={() => openEditModal(user)}>
                          Ubah Role
                        </Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <div className="mt-4 flex items-center justify-between">
              <span className="text-body-sm text-(--ink-secondary)">
                Halaman {customersData.meta.page} dari {customersData.meta.total_pages}
              </span>
              <div className="flex items-center gap-2">
                <Button variant="secondary" onClick={() => goToPage(page - 1)} disabled={page <= 1}>
                  Sebelumnya
                </Button>
                <Button
                  variant="secondary"
                  onClick={() => goToPage(page + 1)}
                  disabled={page >= customersData.meta.total_pages}
                >
                  Berikutnya
                </Button>
              </div>
            </div>
          </>
        )}
      </section>

      {/* Modal / Dialog Edit Role User */}
      {selectedUser && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
          <div className="w-full max-w-lg rounded-lg border border-(--border-default) bg-(--surface-card) p-6 shadow-xl">
            <h3 className="text-heading-md text-(--ink-primary) mb-1">Kelola Role — {selectedUser.name}</h3>
            <p className="text-body-sm text-(--ink-secondary) mb-4">{selectedUser.email}</p>

            {successMessage && (
              <p role="status" className="text-body-sm text-success-600 mb-4 font-medium">
                {successMessage}
              </p>
            )}

            {updateRolesMutation.isError && (
              <p role="alert" className="text-body-sm text-error-500 mb-4 font-medium">
                {updateRolesMutation.error.message || "Gagal memperbarui role."}
              </p>
            )}

            {selectedUser.id === currentUserId && (
              <p className="text-body-sm text-warning-500 mb-3 bg-warning-50 p-2.5 rounded border border-warning-200">
                Anda tidak dapat menghapus role <strong>admin</strong> dari akun Anda sendiri.
              </p>
            )}

            <div className="flex flex-col gap-3 mb-6">
              {roles?.map((role) => {
                const isChecked = selectedRoleNames.includes(role.name);
                const isSelfAdmin = selectedUser.id === currentUserId && role.name === "admin";

                return (
                  <label
                    key={role.id}
                    className={`flex items-start gap-3 rounded-md border p-3 ${
                      isSelfAdmin
                        ? "border-neutral-200 bg-neutral-50 cursor-not-allowed opacity-75"
                        : isChecked
                          ? "border-primary-500 bg-primary-50/20 cursor-pointer"
                          : "border-(--border-default) cursor-pointer hover:bg-neutral-50"
                    }`}
                  >
                    <input
                      type="checkbox"
                      checked={isChecked}
                      disabled={isSelfAdmin}
                      onChange={() => toggleRole(role.name)}
                      className="mt-1 h-4 w-4 accent-primary-600 cursor-pointer disabled:cursor-not-allowed"
                    />
                    <div>
                      <span className="text-label-md text-(--ink-primary) font-semibold uppercase">{role.name}</span>
                      <p className="text-body-xs text-(--ink-secondary) mt-0.5">
                        {role.permissions.map((p) => p.code).join(", ") || "Tidak ada permission khusus"}
                      </p>
                    </div>
                  </label>
                );
              })}
            </div>

            <div className="flex justify-end gap-3 border-t border-(--border-default) pt-4">
              <Button variant="secondary" onClick={closeModal}>
                Tutup
              </Button>
              <Button onClick={handleSaveRoles} disabled={updateRolesMutation.isPending}>
                {updateRolesMutation.isPending ? "Menyimpan..." : "Simpan Role"}
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
