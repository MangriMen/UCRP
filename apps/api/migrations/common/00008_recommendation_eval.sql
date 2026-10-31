-- +goose Up
CREATE TABLE recommendation_eval (
  id                        TEXT         PRIMARY KEY,
  guest_id                  TEXT         NOT NULL REFERENCES guest(id),
  probabilities             TEXT         NOT NULL,
  ranking                   TEXT         NOT NULL,
  confidence                NUMERIC(6,5),
  recommendation_computed_at TIMESTAMP   NOT NULL,
  evaluated_at              TIMESTAMP    NOT NULL
);
CREATE INDEX ix_reco_eval_guest ON recommendation_eval(guest_id, evaluated_at);

-- +goose Down
DROP TABLE IF EXISTS recommendation_eval;
