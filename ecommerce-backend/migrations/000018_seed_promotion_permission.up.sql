INSERT INTO permissions (code) VALUES ('promotion:manage');

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'admin' AND p.code = 'promotion:manage';
