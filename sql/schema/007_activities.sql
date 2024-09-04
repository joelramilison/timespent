-- +goose Up

ALTER TABLE sessions
DROP CONSTRAINT sessions_activity_id_fkey;

ALTER TABLE sessions
ADD CONSTRAINT fk_activity_of_session
FOREIGN KEY (activity_id)
REFERENCES activities (id)
ON DELETE CASCADE;


ALTER TABLE activities
ADD CONSTRAINT fk_user_of_activity
FOREIGN KEY (user_id)
REFERENCES users (id)
ON DELETE CASCADE;



-- +goose Down

ALTER TABLE activities
DROP CONSTRAINT fk_user_of_activity;

ALTER TABLE sessions
DROP CONSTRAINT fk_activity_of_session;

ALTER TABLE sessions
ADD CONSTRAINT sessions_activity_id_fkey
FOREIGN KEY (activity_id)
REFERENCES activities (id);

