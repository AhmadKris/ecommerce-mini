/** Reports whether permissions contains the exact permission code. */
export function hasPermission(permissions: string[], code: string): boolean {
  return permissions.includes(code);
}
