INSERT INTO permissions (code) VALUES ('product:delete');

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'admin' AND p.code = 'product:delete';
