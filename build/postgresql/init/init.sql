-- https://gist.github.com/kjmph/5bd772b2c2df145aa645b837da7eca74
-- Generate a custom UUID v8 with microsecond precision
CREATE
    OR REPLACE FUNCTION uuid_generate_v8()
    RETURNS uuid
AS
$$
DECLARE
    timestamp timestamptz;
    microseconds
              INT;
BEGIN
    timestamp = CLOCK_TIMESTAMP();
    microseconds
        = (CAST(EXTRACT(MICROSECONDS FROM timestamp)::INT -
                (FLOOR(EXTRACT(MILLISECONDS FROM timestamp))::INT * 1000) AS DOUBLE PRECISION) * 4.096)::INT;

    -- use random v4 uuid as starting point (which has the same variant we need)
    -- then overlay timestamp
    -- then set version 8 and add microseconds
    RETURN ENCODE(
            SET_BYTE(
                    SET_BYTE(
                            OVERLAY(uuid_send(gen_random_uuid()) PLACING
                                    SUBSTRING(int8send(FLOOR(EXTRACT(EPOCH FROM timestamp) * 1000)::BIGINT) FROM 3)
                                    FROM 1 FOR 6
                            ),
                            6, (b'1000' || (microseconds >> 8)::BIT(4)):: BIT(8):: INT
                    ),
                    7, microseconds::BIT(8)::INT
            ),
            'hex')::uuid;
END
$$
    LANGUAGE plpgsql
    VOLATILE;

CREATE TABLE users
(
    id         uuid PRIMARY KEY      DEFAULT uuid_generate_v8(),
    name       VARCHAR(255) NOT NULL UNIQUE,
    password   TEXT         NOT NULL,
    created_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE weather_stations
(
    city        VARCHAR(255)     NOT NULL,
    temperature DOUBLE PRECISION NOT NULL
);

