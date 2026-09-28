CREATE SEQUENCE IF NOT EXISTS user_id_seq START WITH 1 INCREMENT BY 1;

SELECT setval(
    'user_id_seq',
    COALESCE((SELECT MAX(user_id) FROM urls), 0)
);
