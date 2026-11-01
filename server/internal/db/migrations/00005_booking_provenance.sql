-- +goose Up
ALTER TABLE bookings
    ADD COLUMN confirmed_at TIMESTAMPTZ,
    ADD COLUMN cancellation_source VARCHAR(16),
    ADD COLUMN cancelled_actor_role VARCHAR(20);

-- Historical cancellations have no trustworthy role snapshot. Mark the
-- source unknown instead of inferring it from the current user role or reason.
UPDATE bookings
SET cancellation_source = 'UNKNOWN'
WHERE status = 'CANCELLED';

ALTER TABLE bookings
    ADD CONSTRAINT ck_bookings_cancellation_source
        CHECK (cancellation_source IS NULL OR cancellation_source IN ('MEMBER','PT','FACILITY','UNKNOWN')),
    ADD CONSTRAINT ck_bookings_cancelled_actor_role
        CHECK (cancelled_actor_role IS NULL OR cancelled_actor_role IN ('ADMIN','MANAGER','RECEPTIONIST','MEMBER','PT')),
    ADD CONSTRAINT ck_bookings_cancellation_provenance
        CHECK (status <> 'CANCELLED' OR COALESCE(
            (cancellation_source = 'UNKNOWN' AND cancelled_actor_role IS NULL)
            OR (cancellation_source = 'MEMBER' AND cancelled_actor_role = 'MEMBER')
            OR (cancellation_source = 'PT' AND cancelled_actor_role = 'PT')
            OR (cancellation_source = 'FACILITY' AND cancelled_actor_role IN ('ADMIN','MANAGER','RECEPTIONIST')),
            FALSE
        ));

-- +goose Down
ALTER TABLE bookings
    DROP CONSTRAINT ck_bookings_cancellation_provenance,
    DROP CONSTRAINT ck_bookings_cancelled_actor_role,
    DROP CONSTRAINT ck_bookings_cancellation_source,
    DROP COLUMN cancelled_actor_role,
    DROP COLUMN cancellation_source,
    DROP COLUMN confirmed_at;
