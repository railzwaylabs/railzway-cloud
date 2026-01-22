CREATE TABLE IF NOT EXISTS instances (
    id BIGSERIAL PRIMARY KEY,
    org_id BIGINT NOT NULL,
    nomad_job_id VARCHAR(255) NOT NULL,
    desired_version VARCHAR(50) NOT NULL,
    current_version VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT fk_instances_org FOREIGN KEY (org_id) REFERENCES organizations(id),
    CONSTRAINT uniq_org_instance UNIQUE (org_id)
);
