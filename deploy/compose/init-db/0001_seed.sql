INSERT INTO environments (id, name, description)
VALUES
  ('11111111-1111-1111-1111-111111111111', 'dev', 'Development environment'),
  ('22222222-2222-2222-2222-222222222222', 'staging', 'Pre-production testing'),
  ('33333333-3333-3333-3333-333333333333', 'prod', 'Production environment')
ON CONFLICT (name) DO NOTHING;

INSERT INTO retention_policies (id, signal_type, retention_days)
VALUES
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'metrics', 30),
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'logs', 14),
  ('cccccccc-cccc-cccc-cccc-cccccccccccc', 'traces', 7)
ON CONFLICT (signal_type) DO NOTHING;