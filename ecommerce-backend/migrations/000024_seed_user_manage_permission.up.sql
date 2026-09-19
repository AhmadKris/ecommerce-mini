-- Seed the user:manage permission and grant it to the admin role.
-- Permission code was documented in CLAUDE.md RBAC examples from the start
-- but never actually inserted into the permissions table — same pattern as
-- product:delete (000012) and the other permission seed migrations.

INSERT INTO permissions (code) VALUES ('user:manage');

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'admin' AND p.code = 'user:manage';
