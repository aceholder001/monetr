CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE EXTENSION IF NOT EXISTS "citext";

CREATE TABLE IF NOT EXISTS "logins"
(
    "login_id"          bigserial NOT NULL,
    "email"             text      NOT NULL UNIQUE,
    "password_hash"     text      NOT NULL,
    "phone_number"      text,
    "is_enabled"        boolean   NOT NULL,
    "is_email_verified" boolean   NOT NULL,
    "is_phone_verified" boolean   NOT NULL,
    PRIMARY KEY ("login_id"),
    UNIQUE ("email")
);

CREATE TABLE IF NOT EXISTS "registrations"
(
    "registration_id" uuid        NOT NULL DEFAULT uuid_generate_v4(),
    "login_id"        bigint      NOT NULL,
    "is_complete"     boolean     NOT NULL,
    "date_created"    timestamptz NOT NULL,
    "date_expires"    timestamptz NOT NULL,
    PRIMARY KEY ("registration_id"),
    FOREIGN KEY ("login_id") REFERENCES "logins" ("login_id") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "email_verifications"
(
    "email_verification_id" bigserial   NOT NULL,
    "login_id"              bigint      NOT NULL,
    "email_address"         text        NOT NULL,
    "is_verified"           boolean     NOT NULL,
    "created_at"            timestamptz NOT NULL DEFAULT now(),
    "expires_at"            timestamptz NOT NULL,
    "verified_at"           timestamptz,
    PRIMARY KEY ("email_verification_id"),
    UNIQUE ("login_id", "email_address"),
    FOREIGN KEY ("login_id") REFERENCES "logins" ("login_id") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "phone_verifications"
(
    "phone_verification_id" bigserial   NOT NULL,
    "login_id"              bigint      NOT NULL,
    "code"                  text        NOT NULL,
    "phone_number"          text        NOT NULL,
    "is_verified"           boolean     NOT NULL,
    "created_at"            timestamptz NOT NULL DEFAULT now(),
    "expires_at"            timestamptz NOT NULL,
    "verified_at"           timestamptz,
    PRIMARY KEY ("phone_verification_id"),
    UNIQUE ("login_id", "code"),
    UNIQUE ("login_id", "phone_number"),
    FOREIGN KEY ("login_id") REFERENCES "logins" ("login_id") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "accounts"
(
    "account_id" bigserial NOT NULL,
    "timezone"   text      NOT NULL DEFAULT 'UTC',
    PRIMARY KEY ("account_id")
);

CREATE TABLE IF NOT EXISTS "users"
(
    "user_id"    bigserial NOT NULL,
    "login_id"   bigint    NOT NULL,
    "account_id" bigint    NOT NULL,
    "first_name" text      NOT NULL,
    "last_name"  text,
    PRIMARY KEY ("user_id"),
    UNIQUE ("login_id", "account_id"),
    FOREIGN KEY ("login_id") REFERENCES "logins" ("login_id") ON DELETE CASCADE,
    FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "plaid_links"
(
    "plaid_link_id"    bigserial NOT NULL,
    "item_id"          text      NOT NULL,
    "access_token"     text      NOT NULL,
    "products"         text[],
    "webhook_url"      text,
    "institution_id"   text,
    "institution_name" text,
    PRIMARY KEY ("plaid_link_id")
);

CREATE TABLE IF NOT EXISTS "links"
(
    "link_id"                 bigserial   NOT NULL,
    "account_id"              bigint      NOT NULL,
    "link_type"               smallint    NOT NULL,
    "plaid_link_id"           bigint,
    "institution_name"        text,
    "custom_institution_name" text,
    "created_at"              timestamptz NOT NULL,
    "created_by_user_id"      bigint      NOT NULL,
    "updated_at"              timestamptz NOT NULL,
    "updated_by_user_id"      bigint,
    PRIMARY KEY ("link_id", "account_id"),
    FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id") ON DELETE CASCADE,
    FOREIGN KEY ("plaid_link_id") REFERENCES "plaid_links" ("plaid_link_id") ON DELETE SET NULL,
    FOREIGN KEY ("created_by_user_id") REFERENCES "users" ("user_id") ON DELETE CASCADE,
    FOREIGN KEY ("updated_by_user_id") REFERENCES "users" ("user_id") ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS "bank_accounts"
(
    "bank_account_id"     bigserial NOT NULL,
    "account_id"          bigserial NOT NULL,
    "link_id"             bigint    NOT NULL,
    "plaid_account_id"    text,
    "available_balance"   bigint    NOT NULL,
    "current_balance"     bigint    NOT NULL,
    "mask"                text,
    "name"                text      NOT NULL,
    "plaid_name"          text,
    "plaid_official_name" text,
    "account_type"        text,
    "account_sub_type"    text,
    PRIMARY KEY ("bank_account_id", "account_id"),
    FOREIGN KEY ("link_id", "account_id") REFERENCES "links" ("link_id", "account_id") ON DELETE CASCADE,
    FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "funding_schedules"
(
    "funding_schedule_id" bigserial NOT NULL,
    "account_id"          bigint    NOT NULL,
    "bank_account_id"     bigint    NOT NULL,
    "name"                text      NOT NULL,
    "description"         text,
    "rule"                text      NOT NULL,
    "last_occurrence"     date,
    "next_occurrence"     date,
    PRIMARY KEY ("funding_schedule_id", "account_id", "bank_account_id"),
    UNIQUE ("bank_account_id", "name"),
    FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id") ON DELETE CASCADE,
    FOREIGN KEY ("bank_account_id", "account_id") REFERENCES "bank_accounts" ("bank_account_id", "account_id") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "expenses"
(
    "expense_id"               bigserial NOT NULL,
    "account_id"               bigint    NOT NULL,
    "bank_account_id"          bigint    NOT NULL,
    "funding_schedule_id"      bigint,
    "name"                     text      NOT NULL,
    "description"              text,
    "target_amount"            bigint    NOT NULL,
    "current_amount"           bigint    NOT NULL,
    "recurrence_rule"          text      NOT NULL,
    "last_recurrence"          date,
    "next_recurrence"          date      NOT NULL,
    "next_contribution_amount" bigint    NOT NULL,
    "is_behind"                boolean   NOT NULL,
    PRIMARY KEY ("expense_id", "account_id", "bank_account_id"),
    UNIQUE ("bank_account_id", "name"),
    FOREIGN KEY ("funding_schedule_id", "account_id", "bank_account_id") REFERENCES "funding_schedules" ("funding_schedule_id", "account_id", "bank_account_id") ON DELETE SET NULL,
    FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id") ON DELETE CASCADE,
    FOREIGN KEY ("bank_account_id", "account_id") REFERENCES "bank_accounts" ("bank_account_id", "account_id") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "transactions"
(
    "transaction_id"         bigserial   NOT NULL,
    "account_id"             bigint      NOT NULL,
    "bank_account_id"        bigint      NOT NULL,
    "plaid_transaction_id"   text,
    "amount"                 bigint      NOT NULL,
    "expense_id"             bigint,
    "categories"             text[],
    "original_categories"    text[],
    "date"                   date        NOT NULL,
    "authorized_date"        date,
    "name"                   text,
    "original_name"          text        NOT NULL,
    "merchant_name"          text,
    "original_merchant_name" text,
    "is_pending"             boolean     NOT NULL,
    "created_at"             timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY ("transaction_id", "account_id", "bank_account_id"),
    UNIQUE ("bank_account_id", "plaid_transaction_id"),
    FOREIGN KEY ("bank_account_id", "account_id") REFERENCES "bank_accounts" ("bank_account_id", "account_id") ON DELETE CASCADE,
    FOREIGN KEY ("expense_id", "account_id", "bank_account_id") REFERENCES "expenses" ("expense_id", "account_id", "bank_account_id") ON DELETE SET NULL,
    FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id") ON DELETE CASCADE
);

INSERT INTO "logins" ("login_id", "email", "password_hash", "phone_number", "is_enabled", "is_email_verified",
                      "is_phone_verified")
VALUES (-1, 'support@harderthanitneedstobe.com', '', DEFAULT, FALSE, FALSE, FALSE)
RETURNING "phone_number";
INSERT INTO "accounts" ("account_id", "timezone")
VALUES (-1, 'UTC');
INSERT INTO "users" ("user_id", "login_id", "account_id", "first_name", "last_name")
VALUES (-1, -1, -1, 'System', 'Bot');
