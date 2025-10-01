DROP INDEX IF EXISTS idx_schema_versions_created_at;
DROP INDEX IF EXISTS idx_schema_versions_service_id;
DROP INDEX IF EXISTS idx_schema_versions_app_id;
DROP INDEX IF EXISTS idx_services_app_id;
DROP INDEX IF EXISTS idx_applications_name;

DROP TABLE IF EXISTS schema_versions;
DROP TABLE IF EXISTS services;
DROP TABLE IF EXISTS applications;