-- ----------------------------------------------------
-- Enable UUID generation defaults (only needed once)
-- ----------------------------------------------------
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

ALTER TABLE orchestrator  ALTER COLUMN id SET DEFAULT gen_random_uuid();
ALTER TABLE site          ALTER COLUMN id SET DEFAULT gen_random_uuid();
ALTER TABLE host          ALTER COLUMN id SET DEFAULT gen_random_uuid();

-- ----------------------------------------------------
-- 1) Insert Central Orchestrator (fixed UUID for reference)
-- ----------------------------------------------------
INSERT INTO orchestrator (id, name, type, region, api_endpoint, created_at, updated_at)
VALUES (
  '550e8400-e29b-41d4-a716-446655440000',  -- fixed UUID
  'Central-Orchestrator',
  'co',
  'India',
  'http://localhost:9000',
  NOW(),
  NOW()
)
ON CONFLICT (id) DO NOTHING;

-- ----------------------------------------------------
-- 2) Insert Sites (UUID auto-generated)
-- ----------------------------------------------------
INSERT INTO site (site_id, name, description, location, orchestrator_id, created_at, updated_at)
VALUES
  ('site-001', 'Chennai-Site', 'Primary site in Chennai', 'Chennai, India',
     '550e8400-e29b-41d4-a716-446655440000', NOW(), NOW()),
  ('site-002', 'Bangalore-Site', 'Secondary site in Bangalore', 'Bangalore, India',
     '550e8400-e29b-41d4-a716-446655440000', NOW(), NOW()),
  ('site-003', 'Tiruvannamalai-Site-1', 'Backup site in Tiruvannamalai', 'Tiruvannamalai, TN, India',
     '550e8400-e29b-41d4-a716-446655440000', NOW(), NOW())
ON CONFLICT (site_id) DO NOTHING;

-- ----------------------------------------------------
-- 3) Insert Hosts using site_id lookup
-- ----------------------------------------------------
INSERT INTO host (
    host_id, site_id, hostname, ip_address, edge_url,
    status, last_seen, metadata, created_at, updated_at
)
VALUES
  -- Hosts for Chennai-Site
  ('host-edge-001', (SELECT id FROM site WHERE site_id = 'site-001'), 'edge1', 'localhost',
        'http://localhost:9105', 'inactive', NOW(),
        '{"role": "edge-node"}', NOW(), NOW()),
  ('host-edge-002', (SELECT id FROM site WHERE site_id = 'site-001'), 'edge2', 'localhost',
        'http://localhost:9106', 'inactive', NOW(),
        '{"role": "edge-node"}', NOW(), NOW()),

  -- Hosts for Bangalore-Site
  ('host-edge-003', (SELECT id FROM site WHERE site_id = 'site-002'), 'edge3', 'localhost',
        'http://localhost:9205', 'inactive', NOW(),
        '{"role": "edge-node"}', NOW(), NOW()),

  -- Hosts for Tiruvannamalai-Site-1
  ('host-edge-004', (SELECT id FROM site WHERE site_id = 'site-003'), 'edge4', 'localhost',
        'http://localhost:9305', 'inactive', NOW(),
        '{"role": "edge-node"}', NOW(), NOW())
ON CONFLICT (host_id) DO NOTHING;
