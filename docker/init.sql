SELECT 'CREATE DATABASE dvijise'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'dvijise')\gexec
