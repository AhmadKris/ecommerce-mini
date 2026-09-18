DELETE FROM role_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE code IN ('inventory:read', 'inventory:adjust')
);
DELETE FROM permissions WHERE code IN ('inventory:read', 'inventory:adjust');
