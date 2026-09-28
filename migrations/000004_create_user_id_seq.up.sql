CREATE SEQUENCE IF NOT EXISTS user_id_seq START WITH 1 INCREMENT BY 1;

DO $$
DECLARE
    max_uid BIGINT;
BEGIN
    SELECT COALESCE(MAX(user_id), 0) INTO max_uid FROM urls;
    IF max_uid > 0 THEN
        PERFORM setval('user_id_seq', max_uid);
    END IF;
END $$;
