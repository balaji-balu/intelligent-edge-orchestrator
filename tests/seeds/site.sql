INSERT INTO orchestrator (id, name, type, region, api_endpoint, created_at, updated_at)
VALUES
  ('550e8400-e29b-41d4-a716-446655440000', 'Central-Orchestrator', 'co', 'India', 'http://localhost:9000', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

INSERT INTO site (id, site_id, name, description, location, orchestrator_id, created_at, updated_at)
VALUES
  (gen_random_uuid(), 'site-001', 'Chennai-Site', 'Primary site in Chennai', 'Chennai, India', '550e8400-e29b-41d4-a716-446655440000', NOW(), NOW()),
  (gen_random_uuid(),'site-002', 'Bangalore-Site', 'Secondary site in Bangalore', 'Bangalore, India', '550e8400-e29b-41d4-a716-446655440000', NOW(), NOW())
ON CONFLICT (site_id) DO NOTHING;

INSERT INTO site (
    id,
    site_id,
    name,
    description,
    location,
    orchestrator_id,
    created_at,
    updated_at
) VALUES (
    gen_random_uuid(),
    'site-003',
    'Tiruvannamalai-Site-1',
    'Backup site in Tiruvannamalai',
    'Tiruvannamalai, TN, India',
    '550e8400-e29b-41d4-a716-446655440000',
    NOW(),
    NOW()
);

-- insert after site. 
--   ensure that site id matches with existing site id
INSERT INTO host (id, host_id, site_id, hostname, ip_address, edge_url, status, last_seen, metadata, created_at, updated_at)
VALUES
  (gen_random_uuid(),'host-edge-001', '321459e0-fdaf-47b1-85c2-5df0af395177', 'edge1', 'localhost', 'http://localhost:9105', 'inactive', NOW(), '{"role": "edge-node"}', NOW(), NOW()),
  (gen_random_uuid(),'host-edge-002', '321459e0-fdaf-47b1-85c2-5df0af395177', 'edge2', 'localhost', 'http://localhost:910681', 'inactive', NOW(), '{"role": "edge-node"}', NOW(), NOW())
ON CONFLICT (host_id) DO NOTHING;