CREATE TABLE IF NOT EXISTS application_versions (
    application_name VARCHAR(100) NOT NULL,
    version VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL,
    release_date TIMESTAMP WITH TIME ZONE NOT NULL,
    is_default BOOLEAN DEFAULT false,
    min_tier VARCHAR(50),
    docker_image TEXT NOT NULL,
    changelog_url TEXT,
    release_notes TEXT,
    breaking_changes BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (application_name, version),
    CONSTRAINT chk_status CHECK (status IN ('stable', 'beta', 'rc', 'deprecated', 'eol'))
);

CREATE INDEX idx_application_versions_app_status ON application_versions(application_name, status);
CREATE INDEX idx_application_versions_app_default ON application_versions(application_name, is_default);
CREATE INDEX idx_application_versions_release_date ON application_versions(release_date DESC);

-- Ensure only one default version per application
CREATE UNIQUE INDEX idx_application_versions_unique_default 
ON application_versions(application_name, is_default) 
WHERE is_default = true;

-- Insert initial versions for Railzway OSS based on actual git tags
INSERT INTO application_versions (
    application_name, version, status, release_date, is_default, min_tier, docker_image, changelog_url, release_notes, breaking_changes
) VALUES
('railzway', 'v1.6.0', 'stable', '2026-01-10 00:00:00+00', true, 'FREE',
 'ghcr.io/smallbiznis/railzway:v1.6.0',
 'https://github.com/smallbiznis/railzway/releases/tag/v1.6.0',
 'Latest stable release with significant performance improvements and new billing features.', false),

('railzway', 'v1.5.5', 'stable', '2025-12-25 00:00:00+00', false, 'FREE',
 'ghcr.io/smallbiznis/railzway:v1.5.5',
 'https://github.com/smallbiznis/railzway/releases/tag/v1.5.5',
 'Stable release including critical bug fixes.', false),

('railzway', 'v1.5.0', 'deprecated', '2025-11-15 00:00:00+00', false, 'FREE',
 'ghcr.io/smallbiznis/railzway:v1.5.0',
 'https://github.com/smallbiznis/railzway/releases/tag/v1.5.0',
 'Deprecated version. Please upgrade to v1.5.5 or later.', false);
