-- https://gist.github.com/kjmph/5bd772b2c2df145aa645b837da7eca74
-- Generate a custom UUID v8 with microsecond precision
CREATE
OR replace FUNCTION uuid_generate_v8()
RETURNS uuid
AS $$
DECLARE
TIMESTAMP    timestamptz;
  microseconds
INT;
BEGIN
TIMESTAMP    = clock_timestamp();
  microseconds
= (CAST(EXTRACT(microseconds FROM TIMESTAMP)::INT - (floor(EXTRACT(milliseconds FROM TIMESTAMP))::INT * 1000) AS DOUBLE PRECISION) * 4.096)::INT;

  -- use random v4 uuid as starting point (which has the same variant we need)
  -- then overlay timestamp
  -- then set version 8 and add microseconds
RETURN encode(
        set_byte(
                set_byte(
                        overlay(uuid_send(gen_random_uuid()) placing SUBSTRING(int8send(floor(EXTRACT(epoch FROM TIMESTAMP) * 1000)::bigint) FROM 3)
                FROM 1 FOR 6
                        ),
                        6, (b'1000' || (microseconds >> 8)::BIT(4)):: BIT (8):: INT
                ),
                7, microseconds::BIT(8)::INT
        ),
        'hex')::uuid;
END
$$
LANGUAGE plpgsql
volatile;

CREATE TABLE users
(
    id         uuid PRIMARY KEY      DEFAULT uuid_generate_v8(),
    name VARCHAR(255) NOT NULL UNIQUE,
    password   TEXT         NOT NULL,
    created_at timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP
);