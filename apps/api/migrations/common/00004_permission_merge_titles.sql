-- +goose Up

UPDATE permission SET title = 'Очередь спорных объединений' WHERE code = 'identity.queue.read';
UPDATE permission SET title = 'Доля оспоренных объединений (< 1%)' WHERE code = 'metric.disputed_merge_rate.read';
UPDATE permission SET title = 'Размер очереди спорных объединений' WHERE code = 'metric.merge_queue_size.read';

-- +goose Down

UPDATE permission SET title = 'Очередь спорных склеек' WHERE code = 'identity.queue.read';
UPDATE permission SET title = 'Доля оспоренных склеек (< 1%)' WHERE code = 'metric.disputed_merge_rate.read';
UPDATE permission SET title = 'Размер очереди спорных склеек' WHERE code = 'metric.merge_queue_size.read';
