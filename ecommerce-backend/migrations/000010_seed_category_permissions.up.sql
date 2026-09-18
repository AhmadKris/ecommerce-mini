INSERT INTO permissions (code) VALUES
    ('category:create'),
    ('category:update'),
    ('category:delete');

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'admin' AND p.code IN ('category:create', 'category:update', 'category:delete');
