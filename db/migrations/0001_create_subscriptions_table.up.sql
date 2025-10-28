CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE IF NOT EXISTS subscriptions (
                                             id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    service_name varchar(200) NOT NULL,
    price integer NOT NULL,
    user_id uuid NOT NULL,
    start_date date NOT NULL,
    end_date date NULL,
    created_at timestamp with time zone DEFAULT now()
    );
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions (user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_service_name ON subscriptions (service_name);
