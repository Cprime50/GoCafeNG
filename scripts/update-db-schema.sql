-- Add company_logo column to jobs table if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'jobs' AND column_name = 'company_logo'
    ) THEN
        ALTER TABLE jobs ADD COLUMN company_logo TEXT;
        RAISE NOTICE 'Added company_logo column to jobs table';
    ELSE
        RAISE NOTICE 'company_logo column already exists in jobs table';
    END IF;
END $$; 