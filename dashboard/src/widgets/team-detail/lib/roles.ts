export const READONLY_ROLES = new Set(["observer"]);

export function isReadOnlyRole(role: string | undefined | null): boolean {
  if (!role) return false;
  return READONLY_ROLES.has(role);
}

export function canMutate(role: string | undefined | null): boolean {
  return !isReadOnlyRole(role);
}
