/** Role as embedded in the customer list — permissions omitted there
 * (backend only preloads Roles, not Roles.Permissions, for the list query). */
export interface CustomerRole {
  id: number;
  name: string;
  permissions?: { id: number; code: string }[];
}

export interface Customer {
  id: number;
  name: string;
  email: string;
  created_at: string;
  roles: CustomerRole[];
}
