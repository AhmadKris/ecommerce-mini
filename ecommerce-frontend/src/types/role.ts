// Role entity returned by GET /api/admin/roles.
export interface Role {
  id: number;
  name: string;
  permissions: { id: number; code: string }[];
}
