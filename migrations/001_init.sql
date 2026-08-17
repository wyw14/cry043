CREATE TABLE IF NOT EXISTS specifications(id text PRIMARY KEY,document_id text NOT NULL,version integer NOT NULL,status text NOT NULL,revision bigint NOT NULL,idempotency_key text NOT NULL UNIQUE,payload jsonb NOT NULL,CHECK(version > 0));
CREATE INDEX IF NOT EXISTS specification_status_idx ON specifications(status,document_id,version DESC);
CREATE TABLE IF NOT EXISTS exceptions(id text PRIMARY KEY,specification_id text NOT NULL REFERENCES specifications(id),rule_id text NOT NULL,expires_at timestamptz NOT NULL,payload jsonb NOT NULL);
CREATE TABLE IF NOT EXISTS acknowledgements(id text PRIMARY KEY,specification_id text NOT NULL REFERENCES specifications(id),version integer NOT NULL,user_id text NOT NULL,payload jsonb NOT NULL,UNIQUE(specification_id,version,user_id));
CREATE TABLE IF NOT EXISTS inspections(id text PRIMARY KEY,specification_id text NOT NULL REFERENCES specifications(id),specification_version integer NOT NULL,payload jsonb NOT NULL,performed_at timestamptz NOT NULL);
CREATE TABLE IF NOT EXISTS remediations(id text PRIMARY KEY,status text NOT NULL,revision bigint NOT NULL,due_at timestamptz NOT NULL,payload jsonb NOT NULL);
CREATE INDEX IF NOT EXISTS remediation_due_idx ON remediations(status,due_at);
CREATE TABLE IF NOT EXISTS audit_events(id text PRIMARY KEY,aggregate_id text NOT NULL,previous_hash text NOT NULL,hash text NOT NULL UNIQUE,payload jsonb NOT NULL,created_at timestamptz NOT NULL);
