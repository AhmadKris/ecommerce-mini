INSERT INTO roles (name) VALUES ('admin'), ('customer');

-- Only permissions actually referenced by the Fase 1 endpoints in .claude/CLAUDE.md.
-- Add more (e.g. product:delete, user:manage) once an endpoint needs them.
INSERT INTO permissions (code) VALUES
    ('product:create'),
    ('product:update'),
    ('order:read_own'),
    ('order:read_all'),
    ('order:update_status'),
    ('cart:manage');

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p WHERE r.name = 'admin';

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'customer' AND p.code IN ('order:read_own', 'cart:manage');
