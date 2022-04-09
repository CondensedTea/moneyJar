-- +goose Up
-- +goose StatementBegin
alter table transactionlog
    add column comment text default '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table transactionlog
    drop column comment;
-- +goose StatementEnd
