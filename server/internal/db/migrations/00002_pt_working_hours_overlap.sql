-- +goose Up
-- Weekly schedules use the same half-open intervals as time slots.
ALTER TABLE pt_working_hours ADD CONSTRAINT pt_working_hours_no_overlap
EXCLUDE USING gist (pt_id WITH =, day_of_week WITH =, int4range(EXTRACT(EPOCH FROM start_time)::INTEGER, EXTRACT(EPOCH FROM end_time)::INTEGER, '[)') WITH &&) WHERE (is_active);

-- +goose Down
ALTER TABLE pt_working_hours DROP CONSTRAINT pt_working_hours_no_overlap;
