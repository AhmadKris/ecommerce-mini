DELETE FROM role_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE code IN ('category:create', 'category:update', 'category:delete')
);
DELETE FROM permissions WHERE code IN ('category:create', 'category:update', 'category:delete');
