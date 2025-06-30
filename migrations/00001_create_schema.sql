DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'uint256') THEN
            CREATE DOMAIN UINT256 AS NUMERIC
                CHECK (VALUE >= 0 AND VALUE < POWER(CAST(2 AS NUMERIC), CAST(256 AS NUMERIC)) AND SCALE(VALUE) = 0);
        ELSE
            ALTER DOMAIN UINT256 DROP CONSTRAINT uint256_check;
            ALTER DOMAIN UINT256 ADD
                CHECK (VALUE >= 0 AND VALUE < POWER(CAST(2 AS NUMERIC), CAST(256 AS NUMERIC)) AND SCALE(VALUE) = 0);
        END IF;
    END
$$;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp" cascade;

create table if not exists gas_fee(
    guid                   TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    token_address          VARCHAR,
    chain_id               UINT256,
    gas_fee                UINT256  default 0,
    timestamp              INTEGER
);
CREATE INDEX IF NOT EXISTS gas_fee_chain_id ON gas_fee(chain_id);

