-- BackSaaS Test Database Initialization
-- This script sets up the test database with necessary schemas and test data
-- Note: The backsaas_test database is already created by Docker environment variables

-- Create test schemas
CREATE SCHEMA IF NOT EXISTS platform;
CREATE SCHEMA IF NOT EXISTS tenants;
CREATE SCHEMA IF NOT EXISTS testing;

-- Create test user with appropriate permissions
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'test_user') THEN
        CREATE ROLE test_user WITH LOGIN PASSWORD 'test_password';
    END IF;
END
$$;

-- Grant permissions to test user
GRANT ALL PRIVILEGES ON DATABASE backsaas_test TO test_user;
GRANT ALL PRIVILEGES ON SCHEMA platform TO test_user;
GRANT ALL PRIVILEGES ON SCHEMA tenants TO test_user;
GRANT ALL PRIVILEGES ON SCHEMA testing TO test_user;

-- Create basic test tables for integration tests
CREATE TABLE IF NOT EXISTS testing.test_runs (
    id SERIAL PRIMARY KEY,
    run_id VARCHAR(255) UNIQUE NOT NULL,
    service VARCHAR(100) NOT NULL,
    test_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL,
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    details JSONB
);

-- Insert initial test data
INSERT INTO testing.test_runs (run_id, service, test_type, status, details) 
VALUES ('init-test', 'system', 'setup', 'completed', '{"message": "Test database initialized"}')
ON CONFLICT (run_id) DO NOTHING;

-- Create function to clean test data between runs
CREATE OR REPLACE FUNCTION testing.clean_test_data()
RETURNS void AS $$
BEGIN
    -- Clean up any test-specific tables
    DELETE FROM testing.test_runs WHERE run_id LIKE 'test-%';
    
    -- Reset sequences
    -- Add more cleanup logic as needed
    
    RAISE NOTICE 'Test data cleaned successfully';
END;
$$ LANGUAGE plpgsql;
