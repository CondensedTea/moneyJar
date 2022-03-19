-- +goose Up
-- +goose StatementBegin
create table users (
    id int primary key,
    name text not null
);
create table accounts (
    id serial primary key,
    from_user int references users(id),
    to_user int references users(id),
    balance bigint default 0,
    unique (from_user, to_user)
);
create table transactionLog (
    account integer references accounts(id),
    balance_change bigint not null,
    ts timestamp default now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table users cascade;
drop table accounts cascade;
drop table transactionLog cascade;
-- +goose StatementEnd
