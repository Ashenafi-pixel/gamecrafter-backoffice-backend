-- Migration: Initial Schema
-- Generated from existing database schema
-- This migration creates the complete database schema

-- PostgreSQL database dump

SET statement_timeout = 0;
SET lock_timeout = 0;

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';

CREATE TYPE public.alert_currency_code AS ENUM (
    'USD',
    'BTC',
    'ETH',
    'SOL',
    'USDT',
    'USDC'
);

CREATE TYPE public.alert_status AS ENUM (
    'active',
    'inactive',
    'triggered'
);

CREATE TYPE public.alert_type AS ENUM (
    'bets_count_less',
    'bets_count_more',
    'bets_amount_less',
    'bets_amount_more',
    'deposits_total_less',
    'deposits_total_more',
    'deposits_type_less',
    'deposits_type_more',
    'withdrawals_total_less',
    'withdrawals_total_more',
    'withdrawals_type_less',
    'withdrawals_type_more',
    'ggr_total_less',
    'ggr_total_more',
    'ggr_single_more'
);

CREATE TYPE public.bet_status AS ENUM (
    'open',
    'in_progress',
    'closed',
    'failed'
);

CREATE TYPE public.components AS ENUM (
    'real_money',
    'bonus_money',
    'points',
    'withdrawal'
);

CREATE TYPE public.conversion_type AS ENUM (
    'deposit',
    'withdrawal',
    'exchange'
);

CREATE TYPE public.currency_type AS ENUM (
    'fiat',
    'crypto'
);

CREATE TYPE public.deposit_session_status AS ENUM (
    'pending',
    'processing',
    'completed',
    'failed',
    'cancelled',
    'expired'
);

CREATE TYPE public.notification_type AS ENUM (
    'promotional',
    'kyc',
    'bonus',
    'welcome',
    'system',
    'alert',
    'payments',
    'security',
    'general',
    'tip'
);

CREATE TYPE public.processortype AS ENUM (
    'internal',
    'pdm'
);

CREATE TYPE public.spinningwheelmysterytypes AS ENUM (
    'point',
    'internet_package_in_gb',
    'better',
    'other'
);

CREATE TYPE public.spinningwheeltypes AS ENUM (
    'point',
    'internet_package_in_gb',
    'better',
    'mystery',
    'free spin'
);

CREATE TYPE public.to_cold_storage_movement_status AS ENUM (
    'notmoved',
    'moved'
);

CREATE TYPE public.wallet_transaction_status AS ENUM (
    'pending',
    'confirmed',
    'failed'
);

CREATE TYPE public.withdrawal_status AS ENUM (
    'pending',
    'processing',
    'completed',
    'failed',
    'cancelled',
    'awaiting_admin_review'
);

CREATE FUNCTION public.check_and_update_user_level(p_user_id uuid) RETURNS void
    LANGUAGE plpgsql
    AS $$
DECLARE
    current_tier RECORD;
    new_tier RECORD;
    user_stats RECORD;
BEGIN
    SELECT ul.*, ct.tier_name, ct.cashback_percentage
    INTO user_stats
    FROM user_levels ul
    LEFT JOIN cashback_tiers ct ON ul.current_tier_id = ct.id
    WHERE ul.user_id = p_user_id;
    
    SELECT ct.*
    INTO new_tier
    FROM cashback_tiers ct
    WHERE ct.is_active = true 
    AND ct.min_ggr_required <= user_stats.total_ggr
    ORDER BY ct.tier_level DESC
    LIMIT 1;
    
    IF new_tier.id IS NOT NULL AND (user_stats.current_tier_id IS NULL OR new_tier.tier_level > user_stats.current_level) THEN
        UPDATE user_levels 
        SET 
            current_level = new_tier.tier_level,
            current_tier_id = new_tier.id,
            level_progress = CASE 
                WHEN new_tier.tier_level = 5 THEN 100.00 -- Diamond is max level
                ELSE LEAST(100.00, (user_stats.total_ggr / (
                    SELECT min_ggr_required 
                    FROM cashback_tiers 
                    WHERE tier_level = new_tier.tier_level + 1 
                    AND is_active = true
                )) * 100.00)
            END,
            last_level_up = NOW(),
            updated_at = NOW()
        WHERE user_id = p_user_id;
    END IF;
END;
$$;

CREATE FUNCTION public.check_kyc_requirement(p_user_id uuid, p_transaction_type character varying, p_amount_usd numeric) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
DECLARE
    user_kyc_status VARCHAR(20);
    thresholds JSONB;
    threshold_value DECIMAL;
BEGIN
    SELECT kyc_status INTO user_kyc_status
    FROM users WHERE id = p_user_id;
    
    SELECT setting_value INTO thresholds
    FROM kyc_settings WHERE setting_key = 'kyc_thresholds';
    
    IF p_transaction_type = 'DEPOSIT' THEN
        threshold_value := (thresholds->>'deposit_threshold_usd')::DECIMAL;
    ELSIF p_transaction_type = 'WITHDRAWAL' THEN
        threshold_value := (thresholds->>'withdrawal_threshold_usd')::DECIMAL;
    ELSE
        RETURN TRUE; -- No KYC required for other transaction types
    END IF;
    
    IF p_amount_usd >= threshold_value AND user_kyc_status NOT IN ('ID_VERIFIED', 'ID_SOF_VERIFIED') THEN
        RETURN FALSE;
    END IF;
    
    RETURN TRUE;
END;
$$;

COMMENT ON FUNCTION public.check_kyc_requirement(p_user_id uuid, p_transaction_type character varying, p_amount_usd numeric) IS 'Checks if user meets KYC requirements for transaction amount';

CREATE FUNCTION public.clean_expired_crypto_wallet_challenges() RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
    DELETE FROM crypto_wallet_challenges WHERE expires_at < NOW();
END;
$$;

CREATE FUNCTION public.cleanup_expired_groove_sessions() RETURNS integer
    LANGUAGE plpgsql
    AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    UPDATE groove_game_sessions 
    WHERE status = 'active' 
    AND expires_at < NOW();
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$;

CREATE FUNCTION public.create_user_balances() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    INSERT INTO balances (user_id, currency_code, amount_cents, amount_units, reserved_cents, reserved_units, updated_at)
    SELECT 
        NEW.id, 
        currency_code, 
        0, 
        0, 
        0, 
        0, 
        CURRENT_TIMESTAMP
    FROM currency_config;
    
    RETURN NEW;
END;
$$;

CREATE FUNCTION public.delete_account_block_on_active() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF NEW.status = 'ACTIVE' AND OLD.status IN ('SUSPENDED','INACTIVE','PENDING') THEN
        DELETE FROM account_block WHERE user_id = NEW.id;
    END IF;
    RETURN NEW;
END;
$$;

CREATE FUNCTION public.generate_groove_session_id() RETURNS text
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN '3818_' || gen_random_uuid()::text;
END;
$$;

CREATE FUNCTION public.get_groove_account_summary(p_account_id character varying) RETURNS TABLE(account_id character varying, balance numeric, currency character varying, status character varying, total_transactions bigint, last_transaction_at timestamp with time zone)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT 
        ga.account_id,
        ga.balance,
        ga.currency,
        ga.status,
        COUNT(gt.transaction_id) as total_transactions,
        MAX(gt.created_at) as last_transaction_at
    FROM groove_accounts ga
    LEFT JOIN groove_transactions gt ON ga.account_id = gt.account_id
    WHERE ga.account_id = p_account_id
    GROUP BY ga.account_id, ga.balance, ga.currency, ga.status;
END;
$$;

CREATE FUNCTION public.get_user_wallets(user_uuid uuid) RETURNS TABLE(wallet_type character varying, wallet_address character varying, wallet_chain character varying, wallet_name character varying, is_verified boolean, last_used_at timestamp with time zone)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT 
        cwc.wallet_type,
        cwc.wallet_address,
        cwc.wallet_chain,
        cwc.wallet_name,
        cwc.is_verified,
        cwc.last_used_at
    FROM crypto_wallet_connections cwc
    WHERE cwc.user_id = user_uuid
    ORDER BY cwc.last_used_at DESC;
END;
$$;

CREATE FUNCTION public.log_kyc_status_change() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF OLD.kyc_status != NEW.kyc_status THEN
        INSERT INTO kyc_status_changes (user_id, old_status, new_status, changed_by, change_reason)
        VALUES (NEW.id, OLD.kyc_status, NEW.kyc_status, NEW.id, 'Status updated');
    END IF;
    RETURN NEW;
END;
$$;

CREATE FUNCTION public.update_admin_activity_logs_updated_at() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;

CREATE FUNCTION public.update_alert_email_groups_updated_at() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;

CREATE FUNCTION public.update_crypto_wallet_connections_updated_at() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;

CREATE FUNCTION public.update_game_session_activity() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.last_activity = NOW();
    RETURN NEW;
END;
$$;

CREATE FUNCTION public.update_groove_accounts_updated_at() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;

CREATE FUNCTION public.update_tip_transactions_updated_at() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$;

CREATE FUNCTION public.update_updated_at_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;

CREATE FUNCTION public.update_user_level_stats() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    UPDATE user_levels 
    SET 
        total_ggr = total_ggr + NEW.ggr_amount,
        updated_at = NOW()
    WHERE user_id = NEW.user_id;
    
    PERFORM check_and_update_user_level(NEW.user_id);
    
    RETURN NEW;
END;
$$;

CREATE TABLE public.account_block (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    blocked_by uuid NOT NULL,
    duration character varying NOT NULL,
    type character varying NOT NULL,
    blocked_from timestamp with time zone,
    blocked_to timestamp with time zone,
    unblocked_at timestamp with time zone,
    reason character varying,
    note character varying,
    created_at timestamp with time zone
);

CREATE TABLE public.active_game_sessions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    game_id character varying(50) NOT NULL,
    session_token character varying(255) NOT NULL,
    device_fingerprint character varying(255),
    ip_address inet,
    user_agent text,
    started_at timestamp with time zone DEFAULT now(),
    last_activity timestamp with time zone DEFAULT now(),
    expires_at timestamp with time zone NOT NULL,
    is_active boolean DEFAULT true
);

CREATE TABLE public.adds_services (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    service_url character varying(255) NOT NULL,
    name character varying(255) NOT NULL,
    service_id character varying(255) NOT NULL,
    service_secret character varying(255) NOT NULL,
    description text,
    status character varying(255) NOT NULL,
    created_by uuid NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    deleted_at timestamp without time zone
);

CREATE TABLE public.admin_activity_actions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(100) NOT NULL,
    description text,
    category_id uuid,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT now()
);

CREATE TABLE public.admin_activity_categories (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(50) NOT NULL,
    description text,
    color character varying(7),
    icon character varying(50),
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT now()
);

CREATE TABLE public.admin_activity_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    admin_user_id uuid NOT NULL,
    action character varying(100) NOT NULL,
    resource_type character varying(50) NOT NULL,
    resource_id uuid,
    description text NOT NULL,
    details jsonb,
    ip_address inet,
    user_agent text,
    session_id character varying(255),
    severity character varying(20) DEFAULT 'info'::character varying,
    category character varying(50) NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT admin_activity_logs_severity_check CHECK (((severity)::text = ANY ((ARRAY['low'::character varying, 'info'::character varying, 'warning'::character varying, 'error'::character varying, 'critical'::character varying])::text[])))
);

CREATE TABLE public.admin_fund_movements (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    admin_id uuid NOT NULL,
    from_address character varying(255) NOT NULL,
    to_address character varying(255) NOT NULL,
    chain_id text NOT NULL,
    currency_code character varying(10) NOT NULL,
    protocol character varying(10) NOT NULL,
    amount numeric(36,18) NOT NULL,
    tx_hash character varying(100),
    status character varying(20) DEFAULT 'pending'::character varying NOT NULL,
    movement_type character varying(20) NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT admin_fund_movements_amount_check CHECK ((amount > (0)::numeric)),
    CONSTRAINT admin_fund_movements_from_address_check CHECK (((from_address)::text <> ''::text)),
    CONSTRAINT admin_fund_movements_movement_type_check CHECK (((movement_type)::text = ANY (ARRAY[('to_cold_storage'::character varying)::text, ('to_hot_wallet'::character varying)::text, ('rebalance'::character varying)::text]))),
    CONSTRAINT admin_fund_movements_status_check CHECK (((status)::text = ANY (ARRAY[('pending'::character varying)::text, ('completed'::character varying)::text, ('failed'::character varying)::text]))),
    CONSTRAINT admin_fund_movements_to_address_check CHECK (((to_address)::text <> ''::text))
);

CREATE TABLE public.affiliate_referal_track (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    affiliate_code character varying(255) NOT NULL,
    user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    deleted_at timestamp with time zone,
    has_first_deposit boolean DEFAULT false,
    registration_postback_status boolean DEFAULT false,
    deposit_postback_status boolean DEFAULT false,
    s1 character varying(255),
    s2 character varying(255),
    s3 character varying(255),
    s4 character varying(255),
    s5 character varying(255)
);

CREATE TABLE public.agent_providers (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    client_id character varying NOT NULL,
    client_secret text NOT NULL,
    status character varying(20) DEFAULT 'active'::character varying NOT NULL,
    name text NOT NULL,
    description text,
    callback_url character varying,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

CREATE TABLE public.agent_referrals (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    request_id character varying(255) NOT NULL,
    callback_url text NOT NULL,
    user_id uuid,
    conversion_type character varying(100),
    amount numeric(20,8) DEFAULT 0,
    msisdn character varying(20),
    converted_at timestamp with time zone DEFAULT now(),
    callback_sent boolean DEFAULT false,
    callback_attempts integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);

CREATE TABLE public.airtime_transactions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    transaction_id character varying NOT NULL,
    cashout numeric NOT NULL,
    billername character varying NOT NULL,
    utilitypackageid integer NOT NULL,
    packagename character varying NOT NULL,
    amount numeric NOT NULL,
    status character varying NOT NULL,
    "timestamp" timestamp without time zone NOT NULL
);

CREATE TABLE public.airtime_utilities (
    local_id uuid DEFAULT gen_random_uuid() NOT NULL,
    id integer NOT NULL,
    productname character varying NOT NULL,
    billername character varying NOT NULL,
    amount character varying NOT NULL,
    isamountfixed boolean NOT NULL,
    price numeric,
    status character varying NOT NULL,
    "timestamp" timestamp without time zone NOT NULL
);

CREATE TABLE public.alchemy_webhooks (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    chain_id text NOT NULL,
    webhook_id character varying(255) NOT NULL,
    webhook_url text NOT NULL,
    network character varying(50) NOT NULL,
    address_count integer DEFAULT 0 NOT NULL,
    signing_key text DEFAULT ''::text NOT NULL,
    max_addresses integer DEFAULT 100000 NOT NULL,
    version character varying(50) DEFAULT 'V2'::character varying NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    deactivation_reason text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT alchemy_webhooks_address_count_check CHECK ((address_count >= 0)),
    CONSTRAINT alchemy_webhooks_max_addresses_check CHECK ((max_addresses > 0))
);

CREATE TABLE public.alert_configurations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    alert_type public.alert_type NOT NULL,
    status public.alert_status DEFAULT 'active'::public.alert_status,
    threshold_amount numeric(20,8) NOT NULL,
    time_window_minutes integer NOT NULL,
    currency_code public.alert_currency_code,
    email_notifications boolean DEFAULT false,
    webhook_url text,
    created_by uuid,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    updated_by uuid,
    email_group_ids uuid[] DEFAULT '{}'::uuid[]
);

CREATE TABLE public.alert_email_group_members (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    group_id uuid NOT NULL,
    email character varying(255) NOT NULL,
    created_at timestamp with time zone DEFAULT now()
);

CREATE TABLE public.alert_email_groups (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    created_by uuid,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    updated_by uuid
);

CREATE TABLE public.alert_rules (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    condition jsonb NOT NULL,
    severity character varying(20) NOT NULL,
    is_active boolean DEFAULT true,
    channels text[] NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT alert_rules_severity_check CHECK (((severity)::text = ANY ((ARRAY['low'::character varying, 'medium'::character varying, 'high'::character varying, 'critical'::character varying])::text[])))
);

CREATE TABLE public.alert_triggers (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    alert_configuration_id uuid NOT NULL,
    triggered_at timestamp with time zone DEFAULT now(),
    trigger_value numeric(20,8) NOT NULL,
    threshold_value numeric(20,8) NOT NULL,
    user_id uuid,
    transaction_id character varying(255),
    amount_usd numeric(20,8),
    currency_code public.alert_currency_code,
    context_data jsonb,
    acknowledged boolean DEFAULT false,
    acknowledged_by uuid,
    acknowledged_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now()
);

CREATE TABLE public.audit_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    action character varying(50) NOT NULL,
    entity_type character varying(50) NOT NULL,
    entity_id character varying(255) NOT NULL,
    admin_id uuid,
    old_values jsonb,
    new_values jsonb,
    ip_address inet,
    user_agent text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT audit_logs_action_check CHECK (((action)::text <> ''::text)),
    CONSTRAINT audit_logs_entity_id_check CHECK (((entity_id)::text <> ''::text)),
    CONSTRAINT audit_logs_entity_type_check CHECK (((entity_type)::text <> ''::text))
);

CREATE TABLE public.balance_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    component character varying(50) NOT NULL,
    operational_group_id uuid,
    operational_type_id uuid,
    description text,
    "timestamp" timestamp without time zone,
    transaction_id character varying DEFAULT ''::character varying NOT NULL,
    status character varying DEFAULT 'COMPLETED'::character varying,
    currency_code character varying(10),
    change_cents bigint DEFAULT 0,
    change_units numeric(36,18) DEFAULT 0,
    balance_after_cents bigint,
    balance_after_units numeric(36,18),
    brand_id uuid DEFAULT '00000000-0000-0000-0000-000000000001'::uuid NOT NULL,
    CONSTRAINT check_non_zero_change CHECK (((change_cents <> 0) OR (change_units <> (0)::numeric)))
);

CREATE TABLE public.balances (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    currency_code character varying(10) NOT NULL,
    amount_cents bigint DEFAULT 0,
    amount_units numeric(36,18) DEFAULT 0,
    reserved_cents bigint DEFAULT 0,
    reserved_units numeric(36,18) DEFAULT 0,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    brand_id uuid DEFAULT '00000000-0000-0000-0000-000000000002'::uuid NOT NULL,
    CONSTRAINT balances_amount_cents_check CHECK ((amount_cents >= 0)),
    CONSTRAINT balances_amount_units_check CHECK ((amount_units >= (0)::numeric)),
    CONSTRAINT balances_reserved_cents_check CHECK ((reserved_cents >= 0)),
    CONSTRAINT balances_reserved_units_check CHECK ((reserved_units >= (0)::numeric))
);

CREATE TABLE public.banners (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    page character varying(50) NOT NULL,
    page_url text NOT NULL,
    image_url text NOT NULL,
    headline character varying(255) NOT NULL,
    tagline character varying(500),
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);

CREATE TABLE public.bets (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    round_id uuid NOT NULL,
    amount numeric(10,2) NOT NULL,
    currency character varying(3) NOT NULL,
    client_transaction_id character varying(255) NOT NULL,
    cash_out_multiplier numeric(10,2),
    payout numeric(10,2),
    "timestamp" timestamp without time zone,
    status character varying DEFAULT 'ACTIVE'::character varying,
    house_edge numeric(5,4) DEFAULT 0.0200,
    is_test_transaction boolean DEFAULT true,
    brand_id uuid DEFAULT '00000000-0000-0000-0000-000000000001'::uuid NOT NULL
);

CREATE TABLE public.brands (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL,
    code character varying(50) NOT NULL,
    domain character varying(255),
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);

CREATE TABLE public.campaign_recipients (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    campaign_id uuid NOT NULL,
    user_id uuid NOT NULL,
    notification_id uuid,
    status character varying(20) DEFAULT 'pending'::character varying NOT NULL,
    sent_at timestamp with time zone,
    delivered_at timestamp with time zone,
    read_at timestamp with time zone,
    error_message text,
    created_at timestamp with time zone DEFAULT now(),
    CONSTRAINT campaign_recipients_status_check CHECK (((status)::text = ANY ((ARRAY['pending'::character varying, 'sent'::character varying, 'delivered'::character varying, 'read'::character varying, 'failed'::character varying])::text[])))
);

CREATE TABLE public.casbin_rule (
    id integer NOT NULL,
    ptype character varying(100),
    v0 character varying(100),
    v1 character varying(100),
    v2 character varying(100),
    v3 character varying(100),
    v4 character varying(100),
    v5 character varying(100)
);

CREATE SEQUENCE public.casbin_rule_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.casbin_rule_id_seq OWNED BY public.casbin_rule.id;

CREATE TABLE public.cashback_claims (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    claim_amount numeric(20,8) NOT NULL,
    currency_code character varying(3) DEFAULT 'USD'::character varying NOT NULL,
    status character varying(20) DEFAULT 'pending'::character varying,
    transaction_id uuid,
    processing_fee numeric(20,8) DEFAULT 0,
    net_amount numeric(20,8) NOT NULL,
    claimed_earnings jsonb NOT NULL,
    admin_notes text,
    processed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);

CREATE TABLE public.cashback_earnings (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    tier_id uuid NOT NULL,
    earning_type character varying(20) NOT NULL,
    source_bet_id uuid,
    ggr_amount numeric(18,6) NOT NULL,
    cashback_rate numeric(5,2) NOT NULL,
    earned_amount numeric(18,6) NOT NULL,
    claimed_amount numeric(18,6) DEFAULT 0,
    available_amount numeric(18,6) NOT NULL,
    status character varying(20) DEFAULT 'pending'::character varying,
    expires_at timestamp with time zone DEFAULT (now() + '30 days'::interval),
    claimed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    brand_id uuid DEFAULT '00000000-0000-0000-0000-000000000001'::uuid NOT NULL
);

CREATE TABLE public.cashback_tiers (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tier_name character varying(50) NOT NULL,
    tier_level integer NOT NULL,
    min_ggr_required numeric(20,8) NOT NULL,
    cashback_percentage numeric(5,2) NOT NULL,
    bonus_multiplier numeric(3,2) DEFAULT 1.00,
    daily_cashback_limit numeric(20,8) DEFAULT NULL::numeric,
    weekly_cashback_limit numeric(20,8) DEFAULT NULL::numeric,
    monthly_cashback_limit numeric(20,8) DEFAULT NULL::numeric,
    special_benefits jsonb DEFAULT '{}'::jsonb,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);

CREATE TABLE public.chain_currencies (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    chain_id text NOT NULL,
    currency_code character varying(10) NOT NULL,
    contract_address character varying(255),
    is_native boolean DEFAULT false NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE public.chain_processing_state (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    chain_id text NOT NULL,
    last_processed_block bigint DEFAULT 0 NOT NULL,
    last_processed_block_hash text,
    last_processed_timestamp timestamp with time zone,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE public.clubs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    club_name character varying NOT NULL,
    status character varying DEFAULT 'ACTIVE'::character varying,
    "timestamp" timestamp without time zone DEFAULT now() NOT NULL
);

CREATE TABLE public.company (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    site_name character varying(255) NOT NULL,
    support_email character varying(255) NOT NULL,
    support_phone character varying(50) NOT NULL,
    maintenance_mode boolean DEFAULT false,
    maximum_login_attempt integer,
    password_expiry integer,
    lockout_duration integer,
    require_two_factor_authentication boolean DEFAULT false,
    ip_list inet[],
    created_by uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT company_lockout_duration_check CHECK ((lockout_duration > 0)),
    CONSTRAINT company_maximum_login_attempt_check CHECK ((maximum_login_attempt > 0)),
    CONSTRAINT company_password_expiry_check CHECK ((password_expiry > 0)),
    CONSTRAINT company_support_email_check CHECK (((support_email)::text ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'::text))
);

CREATE TABLE public.configs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(20) NOT NULL,
    value character varying NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);

CREATE TABLE public.conversion_remainders (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    transaction_id uuid NOT NULL,
    original_amount numeric(36,18) NOT NULL,
    converted_amount bigint NOT NULL,
    remainder_amount numeric(10,8) NOT NULL,
    currency_code character varying(10) NOT NULL,
    conversion_type public.conversion_type NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE public.crypto_kings (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    status character varying DEFAULT 'ACTIVE'::character varying NOT NULL,
    bet_amount numeric NOT NULL,
    won_amount numeric,
    start_crypto_value numeric NOT NULL,
    end_crypto_value numeric NOT NULL,
    selected_end_second integer,
    selected_start_value numeric,
    selected_end_value numeric,
    won_status character varying NOT NULL,
    type character varying NOT NULL,
    "timestamp" timestamp without time zone NOT NULL
);

CREATE TABLE public.crypto_wallet_auth_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    wallet_address character varying(255) NOT NULL,
    wallet_type character varying(50) NOT NULL,
    action character varying(50) NOT NULL,
    ip_address inet,
    user_agent text,
    success boolean NOT NULL,
    error_message text,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now(),
    CONSTRAINT crypto_wallet_auth_logs_action_check CHECK (((action)::text = ANY (ARRAY[('connect'::character varying)::text, ('disconnect'::character varying)::text, ('login'::character varying)::text, ('verify'::character varying)::text, ('challenge'::character varying)::text])))
);

CREATE TABLE public.crypto_wallet_challenges (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    wallet_address character varying(255) NOT NULL,
    wallet_type character varying(50) NOT NULL,
    challenge_message text NOT NULL,
    challenge_nonce character varying(255) NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    is_used boolean DEFAULT false,
    created_at timestamp with time zone DEFAULT now()
);

CREATE TABLE public.crypto_wallet_connections (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    wallet_type character varying(50) NOT NULL,
    wallet_address character varying(255) NOT NULL,
    wallet_chain character varying(50) DEFAULT 'ethereum'::character varying NOT NULL,
    wallet_name character varying(255),
    wallet_icon_url text,
    is_verified boolean DEFAULT false,
    verification_signature text,
    verification_message text,
    verification_timestamp timestamp with time zone,
    last_used_at timestamp with time zone DEFAULT now(),
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT crypto_wallet_connections_wallet_type_check CHECK (((wallet_type)::text = ANY (ARRAY[('metamask'::character varying)::text, ('walletconnect'::character varying)::text, ('coinbase'::character varying)::text, ('phantom'::character varying)::text, ('trust'::character varying)::text, ('ledger'::character varying)::text])))
);

CREATE TABLE public.currencies (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying NOT NULL,
    status character varying DEFAULT 'ACTIVE'::character varying NOT NULL,
    "timestamp" timestamp without time zone DEFAULT now() NOT NULL
);

CREATE TABLE public.currency_config (
    currency_code character varying(10) NOT NULL,
    currency_name character varying(50) NOT NULL,
    currency_type public.currency_type NOT NULL,
    decimal_places integer NOT NULL,
    smallest_unit_name character varying(20),
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT currency_config_decimal_places_check CHECK (((decimal_places >= 0) AND (decimal_places <= 18)))
);

CREATE TABLE public.departements_users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    department_id uuid NOT NULL,
    created_at timestamp with time zone
);

CREATE TABLE public.departments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying NOT NULL,
    notifications text[],
    created_at timestamp without time zone
);

CREATE TABLE public.deposit_event_records (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    chain_id character varying(100) NOT NULL,
    tx_hash character varying(100) NOT NULL,
    block_number bigint NOT NULL,
    from_address character varying(100) NOT NULL,
    to_address character varying(100) NOT NULL,
    currency_code character varying(10) NOT NULL,
    amount character varying(50) NOT NULL,
    amount_units numeric(36,18) DEFAULT 0,
    usd_amount_cents bigint NOT NULL,
    exchange_rate double precision NOT NULL,
    status character varying(50) NOT NULL,
    error_message text,
    retry_count integer DEFAULT 0 NOT NULL,
    last_retry_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT deposit_event_records_amount_check CHECK (((amount)::text <> ''::text)),
    CONSTRAINT deposit_event_records_amount_units_check CHECK ((amount_units >= (0)::numeric)),
    CONSTRAINT deposit_event_records_chain_id_check CHECK (((chain_id)::text <> ''::text)),
    CONSTRAINT deposit_event_records_currency_code_check CHECK (((currency_code)::text <> ''::text)),
    CONSTRAINT deposit_event_records_from_address_check CHECK (((from_address)::text <> ''::text)),
    CONSTRAINT deposit_event_records_status_check CHECK (((status)::text <> ''::text)),
    CONSTRAINT deposit_event_records_to_address_check CHECK (((to_address)::text <> ''::text)),
    CONSTRAINT deposit_event_records_tx_hash_check CHECK (((tx_hash)::text <> ''::text))
);

CREATE TABLE public.deposit_sessions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    session_id character varying(255) NOT NULL,
    user_id uuid NOT NULL,
    wallet_id uuid NOT NULL,
    chain_id text NOT NULL,
    currency_code character varying(10) NOT NULL,
    protocol character varying(10) NOT NULL,
    expected_amount numeric(36,18) NOT NULL,
    usd_amount_cents bigint,
    exchange_rate numeric(15,6),
    status public.deposit_session_status DEFAULT 'pending'::public.deposit_session_status NOT NULL,
    to_cold_storage_movement_status public.to_cold_storage_movement_status DEFAULT 'notmoved'::public.to_cold_storage_movement_status NOT NULL,
    qr_code_data text,
    payment_link text,
    metadata jsonb,
    error_message text,
    verification_attempts integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    first_verification_attempt_at timestamp with time zone,
    CONSTRAINT deposit_sessions_expected_amount_check CHECK ((expected_amount > (0)::numeric)),
    CONSTRAINT deposit_sessions_session_id_check CHECK (((session_id)::text <> ''::text))
);

CREATE TABLE public.emergency_access_overrides (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    game_id character varying(50),
    override_type character varying(20) NOT NULL,
    reason text NOT NULL,
    created_by uuid NOT NULL,
    expires_at timestamp without time zone,
    is_active boolean DEFAULT true,
    created_at timestamp without time zone DEFAULT now(),
    CONSTRAINT emergency_access_overrides_override_type_check CHECK (((override_type)::text = ANY ((ARRAY['allow_all'::character varying, 'deny_all'::character varying, 'specific_game'::character varying])::text[])))
);

CREATE TABLE public.exchange_rates (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    from_currency character varying(10) NOT NULL,
    to_currency character varying(10) NOT NULL,
    rate numeric(15,6) NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT exchange_rates_rate_check CHECK ((rate > (0)::numeric))
);

CREATE TABLE public.failed_bet_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    round_id uuid NOT NULL,
    bet_id uuid NOT NULL,
    manual boolean DEFAULT true NOT NULL,
    admin_id uuid,
    status character varying DEFAULT 'IN_PROGRESS'::character varying NOT NULL,
    created_at timestamp without time zone NOT NULL,
    transaction_id uuid NOT NULL
);

CREATE TABLE public.falcon_liquidity_messages (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    message_id character varying(255) NOT NULL,
    transaction_id character varying(255) NOT NULL,
    user_id uuid NOT NULL,
    message_type character varying(50) NOT NULL,
    message_data jsonb NOT NULL,
    bet_amount numeric(20,8) NOT NULL,
    payout_amount numeric(20,8) NOT NULL,
    currency character varying(10) DEFAULT 'USD'::character varying NOT NULL,
    game_name character varying(255),
    game_id character varying(100),
    house_edge numeric(8,6),
    falcon_routing_key character varying(255) NOT NULL,
    falcon_exchange character varying(255) NOT NULL,
    falcon_queue character varying(255) NOT NULL,
    status character varying(50) DEFAULT 'pending'::character varying NOT NULL,
    retry_count integer DEFAULT 0,
    last_retry_at timestamp without time zone,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    sent_at timestamp without time zone,
    acknowledged_at timestamp without time zone,
    error_message text,
    error_code character varying(100),
    falcon_response jsonb,
    reconciliation_status character varying(50) DEFAULT 'pending'::character varying,
    reconciliation_notes text,
    CONSTRAINT falcon_messages_reconciliation_check CHECK (((reconciliation_status)::text = ANY ((ARRAY['pending'::character varying, 'reconciled'::character varying, 'disputed'::character varying])::text[]))),
    CONSTRAINT falcon_messages_status_check CHECK (((status)::text = ANY ((ARRAY['pending'::character varying, 'sent'::character varying, 'failed'::character varying, 'acknowledged'::character varying])::text[]))),
    CONSTRAINT falcon_messages_type_check CHECK (((message_type)::text = ANY ((ARRAY['casino'::character varying, 'sport'::character varying])::text[])))
);

COMMENT ON TABLE public.falcon_liquidity_messages IS 'Stores all messages sent to Falcon Liquidity for reconciliation and dispute resolution';

COMMENT ON COLUMN public.falcon_liquidity_messages.message_id IS 'Unique identifier for the message sent to Falcon';

COMMENT ON COLUMN public.falcon_liquidity_messages.transaction_id IS 'Original transaction ID from our system';

COMMENT ON COLUMN public.falcon_liquidity_messages.message_data IS 'Complete message data sent to Falcon (JSONB)';

COMMENT ON COLUMN public.falcon_liquidity_messages.status IS 'Current status of the message: pending, sent, failed, acknowledged';

COMMENT ON COLUMN public.falcon_liquidity_messages.falcon_response IS 'Response received from Falcon (if any)';

COMMENT ON COLUMN public.falcon_liquidity_messages.reconciliation_status IS 'Reconciliation status: pending, reconciled, disputed';

COMMENT ON COLUMN public.falcon_liquidity_messages.reconciliation_notes IS 'Notes for dispute resolution and reconciliation';

CREATE TABLE public.football_match_rounds (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    status character varying,
    "timestamp" timestamp without time zone DEFAULT now() NOT NULL
);

CREATE TABLE public.football_matchs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    round_id uuid NOT NULL,
    league character varying NOT NULL,
    date timestamp without time zone NOT NULL,
    home_team character varying NOT NULL,
    away_team character varying,
    status character varying DEFAULT 'ACTIVE'::character varying,
    won character varying,
    "timestamp" timestamp without time zone DEFAULT now() NOT NULL,
    home_score integer DEFAULT 0,
    away_score integer DEFAULT 0
);

CREATE TABLE public.game_access_config (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    config_key character varying(100) NOT NULL,
    config_value jsonb NOT NULL,
    description text,
    is_active boolean DEFAULT true,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);

CREATE TABLE public.game_access_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    game_id character varying(50) NOT NULL,
    action character varying(50) NOT NULL,
    access_granted boolean NOT NULL,
    reason text,
    ip_address inet,
    user_agent text,
    created_at timestamp without time zone DEFAULT now(),
    CONSTRAINT game_access_logs_action_check CHECK (((action)::text = ANY ((ARRAY['launch'::character varying, 'wager'::character varying, 'result'::character varying, 'denied'::character varying])::text[])))
);

CREATE TABLE public.game_access_templates (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    template_name character varying(100) NOT NULL,
    description text,
    access_rules jsonb NOT NULL,
    is_active boolean DEFAULT true,
    created_by uuid,
    created_at timestamp without time zone DEFAULT now()
);

CREATE TABLE public.game_cashback_rates (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    game_type character varying(50) NOT NULL,
    game_variant character varying(100),
    game_id character varying(50),
    cashback_rate numeric(5,2) NOT NULL,
    is_active boolean DEFAULT true,
    effective_from timestamp with time zone DEFAULT now(),
    effective_until timestamp with time zone,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);

CREATE TABLE public.game_house_edges (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    game_type character varying(50) NOT NULL,
    game_variant character varying(50),
    house_edge numeric(10,2) NOT NULL,
    min_bet numeric(20,8) DEFAULT 0,
    max_bet numeric(20,8) DEFAULT NULL::numeric,
    is_active boolean DEFAULT true,
    effective_from timestamp with time zone DEFAULT now(),
    effective_until timestamp with time zone,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    game_id character varying(50),
    CONSTRAINT game_house_edges_house_edge_check CHECK (((house_edge >= (0)::numeric) AND (house_edge <= (100)::numeric))),
    CONSTRAINT game_house_edges_max_bet_check CHECK (((max_bet IS NULL) OR (max_bet > min_bet))),
    CONSTRAINT game_house_edges_min_bet_check CHECK ((min_bet >= (0)::numeric))
);

COMMENT ON TABLE public.game_house_edges IS 'House edge configuration per game and provider';

CREATE TABLE public.game_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    round_id uuid NOT NULL,
    action character varying(255) NOT NULL,
    detail json,
    "timestamp" timestamp without time zone
);

CREATE TABLE public.game_permissions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    game_id character varying(50) NOT NULL,
    permission_name character varying(100) NOT NULL,
    description text,
    is_active boolean DEFAULT true,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);

CREATE TABLE public.game_sessions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    session_id character varying(64) DEFAULT public.generate_groove_session_id(),
    game_id character varying(20) NOT NULL,
    device_type character varying(20) NOT NULL,
    game_mode character varying(10) NOT NULL,
    groove_url text,
    home_url text,
    exit_url text,
    history_url text,
    license_type character varying(20) DEFAULT 'Curacao'::character varying,
    is_test_account boolean DEFAULT false,
    reality_check_elapsed integer DEFAULT 0,
    reality_check_interval integer DEFAULT 60,
    created_at timestamp without time zone DEFAULT now(),
    expires_at timestamp without time zone DEFAULT (now() + '02:00:00'::interval),
    is_active boolean DEFAULT true,
    last_activity timestamp without time zone DEFAULT now(),
    CONSTRAINT game_sessions_device_type_check CHECK (((device_type)::text = ANY (ARRAY[('desktop'::character varying)::text, ('mobile'::character varying)::text]))),
    CONSTRAINT game_sessions_game_mode_check CHECK (((game_mode)::text = ANY (ARRAY[('demo'::character varying)::text, ('real'::character varying)::text])))
);

CREATE TABLE public.games (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying NOT NULL,
    status character varying DEFAULT 'ACTIVE'::character varying NOT NULL,
    "timestamp" timestamp without time zone DEFAULT now() NOT NULL,
    photo character varying,
    price character varying,
    enabled boolean DEFAULT false,
    game_id character varying(255),
    internal_name character varying(255),
    provider character varying(255),
    integration_partner character varying(255),
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON COLUMN public.games.game_id IS 'Unique game identifier from provider';

COMMENT ON COLUMN public.games.internal_name IS 'Internal game name used by the system';

COMMENT ON COLUMN public.games.provider IS 'Game provider (e.g., Pragmatic Play, Evolution)';

COMMENT ON COLUMN public.games.integration_partner IS 'Integration partner (e.g., groovetech)';

CREATE TABLE public.geographic_restrictions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    role_id uuid,
    game_id character varying(50) NOT NULL,
    restriction_type character varying(20) NOT NULL,
    allowed_countries jsonb,
    blocked_countries jsonb,
    allowed_regions jsonb,
    ip_whitelist jsonb,
    ip_blacklist jsonb,
    vpn_detection boolean DEFAULT false,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT geographic_restrictions_restriction_type_check CHECK (((restriction_type)::text = ANY ((ARRAY['user'::character varying, 'role'::character varying, 'global'::character varying])::text[])))
);

CREATE TABLE public.global_rakeback_override (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    is_active boolean DEFAULT false,
    rakeback_percentage numeric(5,2) DEFAULT 0.00,
    start_time timestamp with time zone,
    end_time timestamp with time zone,
    created_by uuid,
    created_at timestamp with time zone DEFAULT now(),
    updated_by uuid,
    updated_at timestamp with time zone DEFAULT now()
);

COMMENT ON TABLE public.global_rakeback_override IS 'Global rakeback override configuration for Happy Hour promotions. Allows setting time-based rakeback overrides.';

COMMENT ON COLUMN public.global_rakeback_override.is_active IS 'When true, rakeback_percentage applies to all users regardless of VIP tier during the time window';

COMMENT ON COLUMN public.global_rakeback_override.rakeback_percentage IS 'Global rakeback percentage (0.00-100.00) that overrides VIP tier rates when active';

COMMENT ON COLUMN public.global_rakeback_override.start_time IS 'Start timestamp for the override period';

COMMENT ON COLUMN public.global_rakeback_override.end_time IS 'End timestamp for the override period';

COMMENT ON COLUMN public.global_rakeback_override.created_by IS 'Admin user ID who created the override';

COMMENT ON COLUMN public.global_rakeback_override.updated_by IS 'Admin user ID who last updated the override';

CREATE TABLE public.groove_accounts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    account_id character varying(255) NOT NULL,
    session_id character varying(255),
    balance numeric(20,8) DEFAULT 0 NOT NULL,
    currency character varying(10) DEFAULT 'USD'::character varying NOT NULL,
    status character varying(50) DEFAULT 'active'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    last_activity timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    bonus_balance numeric(20,8) DEFAULT 0,
    real_balance numeric(20,8) DEFAULT 0,
    game_mode integer DEFAULT 1,
    brand_id uuid DEFAULT '00000000-0000-0000-0000-000000000001'::uuid NOT NULL
);

CREATE TABLE public.groove_game_sessions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    session_id character varying(255) NOT NULL,
    account_id character varying(255) NOT NULL,
    game_id character varying(255) NOT NULL,
    balance numeric(20,8) NOT NULL,
    currency character varying(10) DEFAULT 'USD'::character varying NOT NULL,
    status character varying(50) DEFAULT 'active'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    last_activity timestamp with time zone DEFAULT now() NOT NULL,
    is_test_game_session boolean DEFAULT true,
    user_id uuid,
    device_type character varying(50),
    game_mode character varying(50) DEFAULT 'real'::character varying,
    groove_url text,
    home_url text,
    exit_url text,
    history_url text,
    license_type character varying(50),
    reality_check_elapsed integer DEFAULT 0,
    reality_check_interval integer DEFAULT 0,
    is_active boolean DEFAULT true,
    brand_id uuid DEFAULT '00000000-0000-0000-0000-000000000001'::uuid NOT NULL
);

CREATE TABLE public.groove_transactions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    transaction_id character varying(255) NOT NULL,
    account_id character varying(255) NOT NULL,
    session_id character varying(255),
    amount numeric(20,8) NOT NULL,
    currency character varying(10) DEFAULT 'USD'::character varying NOT NULL,
    type character varying(50) NOT NULL,
    status character varying(50) DEFAULT 'completed'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    metadata jsonb,
    is_test_transaction boolean DEFAULT true,
    account_transaction_id character varying(255),
    game_session_id character varying(255),
    round_id character varying(255),
    game_id character varying(255),
    bet_amount numeric(20,8),
    device character varying(50),
    frbid character varying(255),
    user_id uuid,
    wallet_tx character varying(255),
    bonus_win numeric(20,8) DEFAULT 0,
    real_money_win numeric(20,8) DEFAULT 0,
    bonus_money_bet numeric(20,8) DEFAULT 0,
    real_money_bet numeric(20,8) DEFAULT 0,
    game_mode integer DEFAULT 1,
    order_type character varying(50),
    api_version character varying(10) DEFAULT '1.2'::character varying,
    balance_before numeric(20,8) DEFAULT 0,
    balance_after numeric(20,8) DEFAULT 0,
    win_amount numeric(20,8) DEFAULT 0,
    net_result numeric(20,8) DEFAULT 0,
    brand_id uuid DEFAULT '00000000-0000-0000-0000-000000000001'::uuid NOT NULL
);

COMMENT ON COLUMN public.groove_transactions.balance_before IS 'User balance before transaction';

COMMENT ON COLUMN public.groove_transactions.balance_after IS 'User balance after transaction';

CREATE TABLE public.ip_filters (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_by uuid NOT NULL,
    start_ip character varying NOT NULL,
    end_ip character varying NOT NULL,
    type character varying NOT NULL,
    created_at timestamp with time zone,
    description character varying DEFAULT ''::character varying NOT NULL,
    hits integer DEFAULT 0 NOT NULL,
    last_hit timestamp without time zone
);

CREATE TABLE public.kyc_documents (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    document_type character varying(50) NOT NULL,
    file_name character varying(255) NOT NULL,
    upload_date timestamp with time zone DEFAULT now(),
    status character varying(20) DEFAULT 'PENDING'::character varying,
    reviewed_by uuid,
    rejection_reason text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    file_url text NOT NULL,
    review_date timestamp with time zone
);

COMMENT ON TABLE public.kyc_documents IS 'Stores uploaded KYC documents with review status';

CREATE TABLE public.kyc_settings (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    setting_key character varying(100) NOT NULL,
    setting_value jsonb NOT NULL,
    description text,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);

COMMENT ON TABLE public.kyc_settings IS 'Configurable KYC rules and thresholds';

CREATE TABLE public.kyc_status_changes (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    old_status character varying(20),
    new_status character varying(20) NOT NULL,
    changed_by uuid NOT NULL,
    change_reason text,
    admin_notes text,
    changed_at timestamp with time zone DEFAULT now()
);

COMMENT ON TABLE public.kyc_status_changes IS 'Audit trail for KYC status changes';

CREATE TABLE public.kyc_submissions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    submission_type character varying(20) NOT NULL,
    status character varying(20) NOT NULL,
    submitted_at timestamp with time zone DEFAULT now(),
    reviewed_by uuid,
    reviewed_at timestamp with time zone,
    admin_notes text,
    auto_triggered boolean DEFAULT false,
    trigger_reason character varying(100),
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);

COMMENT ON TABLE public.kyc_submissions IS 'Tracks KYC application submissions and their status';

CREATE TABLE public.leagues (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    league_name character varying NOT NULL,
    status character varying DEFAULT 'ACTIVE'::character varying,
    "timestamp" timestamp without time zone DEFAULT now() NOT NULL
);

CREATE TABLE public.level_requirements (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    level_id uuid NOT NULL,
    type character varying NOT NULL,
    value character varying NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at timestamp without time zone,
    created_by uuid NOT NULL
);

CREATE TABLE public.levels (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    level numeric NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at timestamp without time zone,
    created_by uuid NOT NULL,
    type character varying(50) DEFAULT 'players'::character varying NOT NULL,
    CONSTRAINT levels_type_check CHECK (((type)::text = ANY (ARRAY[('players'::character varying)::text, ('squads'::character varying)::text])))
);

CREATE TABLE public.login_attempts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    ip_address character varying(100) NOT NULL,
    success boolean NOT NULL,
    attempt_time timestamp without time zone,
    user_agent character varying(500),
    brand_id uuid DEFAULT '00000000-0000-0000-0000-000000000001'::uuid NOT NULL
);

CREATE TABLE public.logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    module character varying(255) NOT NULL,
    detail json,
    ip_address character varying(46),
    "timestamp" timestamp without time zone
);

CREATE TABLE public.loot_box (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    type text NOT NULL,
    prizeamount numeric NOT NULL,
    weight numeric NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE public.loot_box_place_bets (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    user_selection uuid,
    loot_box jsonb NOT NULL,
    wonstatus character varying(10) DEFAULT 'pending'::character varying NOT NULL,
    status text DEFAULT 'pending'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE public.lotteries (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    price numeric NOT NULL,
    min_selectable integer NOT NULL,
    max_selectable integer NOT NULL,
    draw_frequency text NOT NULL,
    number_of_balls integer NOT NULL,
    description text NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE public.lottery_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    lottery_id uuid NOT NULL,
    lottery_reward_id uuid NOT NULL,
    prize numeric NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    uniq_identifier uuid DEFAULT gen_random_uuid() NOT NULL,
    draw_numbers integer[]
);

CREATE TABLE public.lottery_services (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    client_id character varying NOT NULL,
    client_secret text NOT NULL,
    status character varying(20) DEFAULT 'active'::character varying NOT NULL,
    name text NOT NULL,
    description text,
    callback_url character varying,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

CREATE TABLE public.lottery_winners_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    lottery_id uuid NOT NULL,
    user_id uuid NOT NULL,
    reward_id uuid NOT NULL,
    won_amount numeric NOT NULL,
    currency text DEFAULT 'P'::text NOT NULL,
    number_of_tickets integer NOT NULL,
    ticket_number text NOT NULL,
    status text DEFAULT 'closed'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE public.manual_funds (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    admin_id uuid NOT NULL,
    transaction_id character varying NOT NULL,
    type character varying NOT NULL,
    note character varying NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    reason character varying DEFAULT 'system_restart'::character varying NOT NULL,
    filename character varying(255),
    amount_cents bigint DEFAULT 0 NOT NULL,
    currency_code character varying(10) NOT NULL,
    CONSTRAINT manual_funds_amount_cents_check CHECK ((amount_cents > 0))
);

CREATE TABLE public.message_campaigns (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    title character varying(255) NOT NULL,
    subject character varying(255) NOT NULL,
    content text NOT NULL,
    created_by uuid NOT NULL,
    status character varying(20) DEFAULT 'draft'::character varying NOT NULL,
    scheduled_at timestamp with time zone,
    sent_at timestamp with time zone,
    total_recipients integer DEFAULT 0,
    delivered_count integer DEFAULT 0,
    read_count integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    message_type public.notification_type NOT NULL,
    CONSTRAINT message_campaigns_status_check CHECK (((status)::text = ANY ((ARRAY['draft'::character varying, 'scheduled'::character varying, 'sending'::character varying, 'sent'::character varying, 'cancelled'::character varying])::text[])))
);

CREATE TABLE public.message_segments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    campaign_id uuid NOT NULL,
    segment_type character varying(50) NOT NULL,
    segment_name character varying(255),
    criteria jsonb,
    csv_data text,
    user_count integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now(),
    CONSTRAINT message_segments_segment_type_check CHECK (((segment_type)::text = ANY ((ARRAY['criteria'::character varying, 'csv'::character varying, 'all_users'::character varying])::text[])))
);

CREATE TABLE public.monitoring_reports (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    report_type character varying(100) NOT NULL,
    status character varying(20) DEFAULT 'pending'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    completed_at timestamp with time zone,
    download_url text,
    error text,
    filters jsonb,
    created_by uuid,
    CONSTRAINT monitoring_reports_status_check CHECK (((status)::text = ANY ((ARRAY['pending'::character varying, 'processing'::character varying, 'completed'::character varying, 'failed'::character varying])::text[])))
);

CREATE TABLE public.operational_groups (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(50),
    description text,
    created_at timestamp without time zone
);

CREATE TABLE public.operational_types (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    group_id uuid NOT NULL,
    name character varying(50),
    description text,
    created_at timestamp without time zone
);

CREATE TABLE public.otps (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    email character varying(255) NOT NULL,
    otp_code character varying(10) NOT NULL,
    type character varying(50) DEFAULT 'email_verification'::character varying NOT NULL,
    status character varying(20) DEFAULT 'pending'::character varying NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    verified_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);

COMMENT ON TABLE public.otps IS 'OTP table for email verification and password reset functionality';

COMMENT ON COLUMN public.otps.email IS 'Email address for OTP delivery';

COMMENT ON COLUMN public.otps.otp_code IS '6-digit OTP code';

COMMENT ON COLUMN public.otps.type IS 'Type of OTP: email_verification, password_reset, login';

COMMENT ON COLUMN public.otps.status IS 'Status: pending, verified, expired, used';

COMMENT ON COLUMN public.otps.expires_at IS 'When the OTP expires';

COMMENT ON COLUMN public.otps.verified_at IS 'When the OTP was verified';

CREATE TABLE public.pages (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    path character varying(255) NOT NULL,
    label character varying(255) NOT NULL,
    parent_id uuid,
    icon character varying(100),
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

COMMENT ON TABLE public.pages IS 'Stores all available pages/routes in the system';

COMMENT ON COLUMN public.pages.path IS 'Unique route path (e.g., /dashboard, /players)';

COMMENT ON COLUMN public.pages.label IS 'Display name for the page (e.g., Dashboard, Player Management)';

COMMENT ON COLUMN public.pages.parent_id IS 'Reference to parent page (for hierarchical structure - sidebar items as parents)';

CREATE TABLE public.passkey_credentials (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    credential_id text NOT NULL,
    raw_id bytea NOT NULL,
    public_key bytea NOT NULL,
    attestation_object bytea NOT NULL,
    client_data_json bytea NOT NULL,
    counter bigint DEFAULT 0 NOT NULL,
    name character varying(255) DEFAULT 'Passkey'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    last_used_at timestamp with time zone,
    is_active boolean DEFAULT true
);

COMMENT ON TABLE public.passkey_credentials IS 'Stores WebAuthn passkey credentials for 2FA authentication';

COMMENT ON COLUMN public.passkey_credentials.credential_id IS 'Base64-encoded credential ID from WebAuthn';

COMMENT ON COLUMN public.passkey_credentials.raw_id IS 'Raw credential ID bytes';

COMMENT ON COLUMN public.passkey_credentials.public_key IS 'Public key from WebAuthn attestation';

COMMENT ON COLUMN public.passkey_credentials.attestation_object IS 'WebAuthn attestation object';

COMMENT ON COLUMN public.passkey_credentials.client_data_json IS 'WebAuthn client data JSON';

COMMENT ON COLUMN public.passkey_credentials.counter IS 'WebAuthn signature counter';

COMMENT ON COLUMN public.passkey_credentials.name IS 'User-friendly name for the passkey';

COMMENT ON COLUMN public.passkey_credentials.last_used_at IS 'When the passkey was last used for authentication';

COMMENT ON COLUMN public.passkey_credentials.is_active IS 'Whether the passkey is active and can be used';

CREATE TABLE public.permissions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying NOT NULL,
    description character varying
);

CREATE TABLE public.player_deposit_tracking (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    deposit_id uuid,
    amount numeric(20,8) NOT NULL,
    currency character varying(10) DEFAULT 'USD'::character varying NOT NULL,
    period_type character varying(20) NOT NULL,
    period_start timestamp without time zone NOT NULL,
    period_end timestamp without time zone NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    CONSTRAINT valid_period_type CHECK (((period_type)::text = ANY (ARRAY[('daily'::character varying)::text, ('weekly'::character varying)::text, ('monthly'::character varying)::text])))
);

CREATE TABLE public.player_excluded_games (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    game_id character varying(255) NOT NULL,
    exclusion_type character varying(20) DEFAULT 'temporary'::character varying,
    start_date timestamp without time zone DEFAULT now(),
    end_date timestamp without time zone,
    reason text,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    CONSTRAINT valid_exclusion_dates CHECK (((end_date IS NULL) OR ((start_date IS NOT NULL) AND (end_date > start_date))))
);

CREATE TABLE public.player_gaming_time_tracking (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    session_id uuid,
    minutes_played integer DEFAULT 0 NOT NULL,
    period_type character varying(20) NOT NULL,
    period_start timestamp without time zone NOT NULL,
    period_end timestamp without time zone NOT NULL,
    session_start timestamp without time zone NOT NULL,
    session_end timestamp without time zone,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    CONSTRAINT valid_period_type_time CHECK (((period_type)::text = ANY (ARRAY[('daily'::character varying)::text, ('weekly'::character varying)::text, ('monthly'::character varying)::text])))
);

CREATE TABLE public.player_self_protection_activity_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    activity_type character varying(50) NOT NULL,
    action character varying(50) NOT NULL,
    old_value jsonb,
    new_value jsonb,
    ip_address character varying(46),
    user_agent character varying(255),
    created_at timestamp without time zone DEFAULT now(),
    CONSTRAINT valid_action CHECK (((action)::text = ANY (ARRAY[('enabled'::character varying)::text, ('disabled'::character varying)::text, ('updated'::character varying)::text, ('added'::character varying)::text, ('removed'::character varying)::text]))),
    CONSTRAINT valid_activity_type CHECK (((activity_type)::text = ANY (ARRAY[('deposit_limit'::character varying)::text, ('time_limit'::character varying)::text, ('self_exclusion'::character varying)::text, ('reality_check'::character varying)::text])))
);

CREATE TABLE public.player_self_protection_settings (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    deposit_limit_enabled boolean DEFAULT false,
    deposit_limit_daily numeric(20,8) DEFAULT 0,
    deposit_limit_weekly numeric(20,8) DEFAULT 0,
    deposit_limit_monthly numeric(20,8) DEFAULT 0,
    deposit_limit_currency character varying(10) DEFAULT 'USD'::character varying,
    time_limit_enabled boolean DEFAULT false,
    time_limit_daily_minutes integer DEFAULT 0,
    time_limit_weekly_minutes integer DEFAULT 0,
    time_limit_monthly_minutes integer DEFAULT 0,
    self_exclusion_enabled boolean DEFAULT false,
    self_exclusion_type character varying(20) DEFAULT 'temporary'::character varying,
    self_exclusion_start_date timestamp without time zone,
    self_exclusion_end_date timestamp without time zone,
    self_exclusion_reason text,
    reality_check_enabled boolean DEFAULT false,
    reality_check_interval_minutes integer DEFAULT 60,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    CONSTRAINT valid_deposit_limits CHECK (((deposit_limit_daily >= (0)::numeric) AND (deposit_limit_weekly >= (0)::numeric) AND (deposit_limit_monthly >= (0)::numeric))),
    CONSTRAINT valid_reality_check_interval CHECK ((reality_check_interval_minutes > 0)),
    CONSTRAINT valid_self_exclusion_dates CHECK (((self_exclusion_end_date IS NULL) OR ((self_exclusion_start_date IS NOT NULL) AND (self_exclusion_end_date > self_exclusion_start_date)))),
    CONSTRAINT valid_time_limits CHECK (((time_limit_daily_minutes >= 0) AND (time_limit_weekly_minutes >= 0) AND (time_limit_monthly_minutes >= 0)))
);

CREATE TABLE public.plinko (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    bet_amount numeric NOT NULL,
    drop_path character varying NOT NULL,
    multiplier numeric,
    win_amount numeric,
    finalposition numeric,
    "timestamp" timestamp without time zone DEFAULT now() NOT NULL
);

CREATE TABLE public.public_winner_notifications (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    username character varying(255) NOT NULL,
    game_name character varying(255) NOT NULL,
    win_amount numeric(20,8) NOT NULL,
    net_winnings numeric(20,8) NOT NULL,
    currency character varying(10) NOT NULL,
    provider character varying(255) NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    expires_at timestamp with time zone NOT NULL
);

CREATE TABLE public.quick_hustles (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    status character varying DEFAULT 'ACTIVE'::character varying NOT NULL,
    bet_amount numeric NOT NULL,
    won_status character varying,
    user_guessed character varying,
    first_card character varying NOT NULL,
    second_card character varying,
    "timestamp" timestamp without time zone NOT NULL,
    won_amount numeric
);

CREATE TABLE public.rakeback_schedules (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    start_time timestamp with time zone NOT NULL,
    end_time timestamp with time zone NOT NULL,
    percentage numeric(5,2) NOT NULL,
    scope_type character varying(50) DEFAULT 'all'::character varying NOT NULL,
    scope_value text,
    status character varying(20) DEFAULT 'scheduled'::character varying NOT NULL,
    created_by uuid,
    activated_at timestamp with time zone,
    deactivated_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT valid_percentage CHECK (((percentage >= 0.00) AND (percentage <= 100.00))),
    CONSTRAINT valid_scope_type CHECK (((scope_type)::text = ANY ((ARRAY['all'::character varying, 'provider'::character varying, 'game'::character varying])::text[]))),
    CONSTRAINT valid_status CHECK (((status)::text = ANY ((ARRAY['scheduled'::character varying, 'active'::character varying, 'completed'::character varying, 'cancelled'::character varying])::text[]))),
    CONSTRAINT valid_time_range CHECK ((end_time > start_time))
);

COMMENT ON TABLE public.rakeback_schedules IS 'Scheduled rakeback events with automatic activation/deactivation based on time windows';

COMMENT ON COLUMN public.rakeback_schedules.name IS 'Display name for the schedule (e.g., "Weekend Boost", "Happy Hour")';

COMMENT ON COLUMN public.rakeback_schedules.start_time IS 'When the rakeback override should automatically activate';

COMMENT ON COLUMN public.rakeback_schedules.end_time IS 'When the rakeback override should automatically deactivate';

COMMENT ON COLUMN public.rakeback_schedules.percentage IS 'Rakeback percentage (0-100%) to apply during the scheduled window';

COMMENT ON COLUMN public.rakeback_schedules.scope_type IS 'Scope of the rakeback: all (all games), provider (specific provider), game (specific game)';

COMMENT ON COLUMN public.rakeback_schedules.scope_value IS 'Provider name or game ID when scope_type is provider or game';

COMMENT ON COLUMN public.rakeback_schedules.status IS 'Current status: scheduled (pending), active (running now), completed (finished), cancelled (manually cancelled)';

CREATE TABLE public.restriction_violations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    game_id character varying(50) NOT NULL,
    violation_type character varying(50) NOT NULL,
    violation_details jsonb NOT NULL,
    action_taken character varying(50) NOT NULL,
    created_at timestamp with time zone DEFAULT now()
);

CREATE TABLE public.retryable_operations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    type character varying(100) NOT NULL,
    user_id uuid NOT NULL,
    data jsonb NOT NULL,
    attempts integer DEFAULT 0 NOT NULL,
    last_error text,
    next_retry_at timestamp with time zone,
    status character varying(50) DEFAULT 'pending'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE public.risk_settings (
    id smallint DEFAULT 1 NOT NULL,
    system_limits_enabled boolean DEFAULT false NOT NULL,
    system_max_daily_airtime_conversion bigint DEFAULT 0 NOT NULL,
    system_max_weekly_airtime_conversion bigint DEFAULT 0 NOT NULL,
    system_max_monthly_airtime_conversion bigint DEFAULT 0 NOT NULL,
    player_limits_enabled boolean DEFAULT false NOT NULL,
    player_max_daily_airtime_conversion integer DEFAULT 0 NOT NULL,
    player_max_weekly_airtime_conversion integer DEFAULT 0 NOT NULL,
    player_max_monthly_airtime_conversion integer DEFAULT 0 NOT NULL,
    player_min_airtime_conversion_amount integer DEFAULT 0 NOT NULL,
    player_conversion_cooldown_hours integer DEFAULT 0 NOT NULL,
    kyc_required_above_amount integer DEFAULT 0 NOT NULL,
    kyc_verification_timeout_hours integer DEFAULT 0 NOT NULL,
    kyc_allow_partial boolean DEFAULT false NOT NULL,
    fraud_max_login_attempts smallint DEFAULT 5 NOT NULL,
    fraud_login_lockout_duration_minutes integer DEFAULT 0 NOT NULL,
    alert_admins_on_trigger boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT singleton_check CHECK ((id = 1))
);

CREATE TABLE public.role_game_access (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    role_id uuid NOT NULL,
    game_id character varying(50) NOT NULL,
    access_type character varying(20) NOT NULL,
    restrictions jsonb,
    is_active boolean DEFAULT true,
    created_at timestamp without time zone DEFAULT now(),
    CONSTRAINT role_game_access_access_type_check CHECK (((access_type)::text = ANY ((ARRAY['allow'::character varying, 'deny'::character varying, 'restricted'::character varying])::text[])))
);

CREATE TABLE public.role_permissions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    role_id uuid NOT NULL,
    permission_id uuid NOT NULL,
    value numeric(20,8)
);

CREATE TABLE public.roles (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying NOT NULL,
    description character varying
);

CREATE TABLE public.roll_da_dice (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    status character varying DEFAULT 'ACTIVE'::character varying NOT NULL,
    bet_amount numeric NOT NULL,
    won_status character varying,
    crash_point numeric NOT NULL,
    "timestamp" timestamp without time zone NOT NULL,
    won_amount numeric,
    user_guessed_start_point numeric,
    user_guessed_end_point numeric,
    multiplier numeric
);

CREATE TABLE public.rounds (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    status public.bet_status NOT NULL,
    crash_point numeric(10,2) NOT NULL,
    created_at timestamp without time zone NOT NULL,
    closed_at timestamp without time zone
);

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);

CREATE TABLE public.scratch_cards (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    status character varying DEFAULT 'ACTIVE'::character varying NOT NULL,
    bet_amount numeric NOT NULL,
    won_status character varying,
    "timestamp" timestamp without time zone NOT NULL,
    won_amount numeric
);

CREATE TABLE public.service_api_keys (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    issuer_service text NOT NULL,
    receiver_service text NOT NULL,
    key text NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT service_api_keys_key_check CHECK ((key <> ''::text))
);

CREATE TABLE public.session_limits (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    role_id uuid,
    game_id character varying(50) NOT NULL,
    limit_type character varying(20) NOT NULL,
    max_concurrent_sessions integer DEFAULT 1,
    session_timeout_minutes integer DEFAULT 30,
    device_tracking boolean DEFAULT true,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT session_limits_limit_type_check CHECK (((limit_type)::text = ANY ((ARRAY['user'::character varying, 'role'::character varying, 'global'::character varying])::text[])))
);

CREATE TABLE public.session_tracking (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    game_id character varying(50) NOT NULL,
    session_id character varying(255) NOT NULL,
    session_start timestamp with time zone DEFAULT now(),
    session_end timestamp with time zone,
    is_active boolean DEFAULT true,
    device_info jsonb,
    ip_address inet,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);

CREATE TABLE public.spending_limits (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    role_id uuid,
    game_id character varying(50) NOT NULL,
    limit_type character varying(20) NOT NULL,
    period_type character varying(20) NOT NULL,
    limit_amount numeric(20,8) NOT NULL,
    currency_code character varying(3) DEFAULT 'USD'::character varying,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT spending_limits_limit_type_check CHECK (((limit_type)::text = ANY ((ARRAY['user'::character varying, 'role'::character varying, 'global'::character varying])::text[]))),
    CONSTRAINT spending_limits_period_type_check CHECK (((period_type)::text = ANY ((ARRAY['daily'::character varying, 'weekly'::character varying, 'monthly'::character varying])::text[])))
);

CREATE TABLE public.spending_tracking (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    game_id character varying(50) NOT NULL,
    amount numeric(20,8) NOT NULL,
    currency_code character varying(3) DEFAULT 'USD'::character varying,
    period_start timestamp with time zone NOT NULL,
    period_end timestamp with time zone NOT NULL,
    period_type character varying(20) NOT NULL,
    transaction_id uuid,
    created_at timestamp with time zone DEFAULT now()
);

CREATE TABLE public.spinning_wheel_configs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying NOT NULL,
    amount numeric NOT NULL,
    type public.spinningwheeltypes NOT NULL,
    status character varying DEFAULT 'ACTIVE'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid NOT NULL,
    deleted_at timestamp with time zone,
    frequency integer DEFAULT 1 NOT NULL,
    icon character varying DEFAULT ''::character varying NOT NULL,
    color text DEFAULT 'blue'::text NOT NULL
);

CREATE TABLE public.spinning_wheel_mysteries (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying NOT NULL,
    amount numeric NOT NULL,
    type public.spinningwheelmysterytypes NOT NULL,
    status character varying DEFAULT 'ACTIVE'::character varying NOT NULL,
    frequency integer DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid NOT NULL,
    deleted_at timestamp with time zone,
    icon character varying DEFAULT ''::character varying NOT NULL
);

CREATE TABLE public.spinning_wheel_rewards (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    round_id uuid NOT NULL,
    name character varying(255) NOT NULL,
    amount numeric NOT NULL,
    type character varying(255) NOT NULL,
    status character varying(255) NOT NULL,
    claim_status character varying(255) NOT NULL,
    transaction_id character varying,
    user_id uuid NOT NULL
);

CREATE TABLE public.spinning_wheels (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    status character varying DEFAULT 'ACTIVE'::character varying NOT NULL,
    bet_amount character varying NOT NULL,
    "timestamp" timestamp without time zone NOT NULL,
    won_amount character varying,
    won_status character varying,
    type character varying(255)
);

CREATE TABLE public.sport_bets (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    transaction_id character varying(255) NOT NULL,
    bet_amount numeric(10,2) NOT NULL,
    bet_reference_num character varying(255) NOT NULL,
    game_reference character varying(255) NOT NULL,
    bet_mode character varying(50) NOT NULL,
    description text,
    user_id uuid NOT NULL,
    frontend_type character varying(50),
    status character varying(50),
    sport_ids text,
    site_id character varying(255) NOT NULL,
    client_ip character varying(255),
    affiliate_user_id character varying(255),
    autorecharge character varying(10),
    bet_details jsonb NOT NULL,
    currency character varying(3) DEFAULT 'NGN'::character varying NOT NULL,
    potential_win numeric(10,2),
    actual_win numeric(10,2),
    odds numeric(10,4),
    placed_at timestamp without time zone DEFAULT now() NOT NULL,
    settled_at timestamp without time zone,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    brand_id uuid DEFAULT '00000000-0000-0000-0000-000000000001'::uuid NOT NULL
);

CREATE TABLE public.squads (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    handle character varying NOT NULL,
    owner uuid NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at timestamp without time zone,
    type character varying(50) DEFAULT 'Open'::character varying NOT NULL
);

CREATE TABLE public.squads_earns (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    squad_id uuid NOT NULL,
    user_id uuid NOT NULL,
    currency character varying(3) DEFAULT 'P'::character varying NOT NULL,
    earned numeric DEFAULT 0 NOT NULL,
    game_id uuid NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at timestamp without time zone
);

CREATE TABLE public.squads_memebers (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    squad_id uuid NOT NULL,
    user_id uuid NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at timestamp without time zone
);

CREATE TABLE public.street_kings (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    status character varying DEFAULT 'ACTIVE'::character varying NOT NULL,
    version character varying(50) NOT NULL,
    bet_amount numeric NOT NULL,
    won_amount numeric,
    crash_point numeric NOT NULL,
    cash_out_point numeric,
    "timestamp" timestamp without time zone NOT NULL
);

CREATE TABLE public.supported_chains (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    chain_id text NOT NULL,
    chain_name text NOT NULL,
    protocol text NOT NULL,
    is_testnet boolean DEFAULT false NOT NULL,
    native_currency character varying(10) NOT NULL,
    processor public.processortype DEFAULT 'internal'::public.processortype NOT NULL,
    status character varying(20) DEFAULT 'active'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT supported_chains_chain_id_check CHECK ((chain_id <> ''::text)),
    CONSTRAINT supported_chains_chain_name_check CHECK ((chain_name <> ''::text)),
    CONSTRAINT supported_chains_protocol_check CHECK ((protocol <> ''::text)),
    CONSTRAINT supported_chains_status_check CHECK (((status)::text = ANY (ARRAY[('active'::character varying)::text, ('inactive'::character varying)::text])))
);

CREATE TABLE public.system_config (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    config_key character varying(100) NOT NULL,
    config_value jsonb NOT NULL,
    description text,
    updated_by uuid,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    brand_id uuid,
    CONSTRAINT system_config_config_key_check CHECK (((config_key)::text <> ''::text))
);

COMMENT ON COLUMN public.system_config.brand_id IS 'NULL = global config (applies to all brands), UUID = brand-specific config';

CREATE TABLE public.temp (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    data jsonb NOT NULL
);

CREATE TABLE public.time_restrictions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    role_id uuid,
    game_id character varying(50) NOT NULL,
    restriction_type character varying(20) NOT NULL,
    timezone character varying(50) DEFAULT 'UTC'::character varying,
    allowed_hours jsonb NOT NULL,
    allowed_days jsonb,
    recurring_pattern character varying(20) DEFAULT 'daily'::character varying,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT time_restrictions_restriction_type_check CHECK (((restriction_type)::text = ANY ((ARRAY['user'::character varying, 'role'::character varying, 'global'::character varying])::text[])))
);

CREATE TABLE public.tip_transactions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    sender_id uuid NOT NULL,
    receiver_id uuid NOT NULL,
    transaction_id uuid DEFAULT gen_random_uuid() NOT NULL,
    amount numeric(36,18) NOT NULL,
    fee numeric(36,18) DEFAULT 0 NOT NULL,
    fee_paid_by character varying(10) NOT NULL,
    currency_code character varying(10) DEFAULT 'USD'::character varying NOT NULL,
    message text,
    sender_balance_before numeric(36,18) NOT NULL,
    sender_balance_after numeric(36,18) NOT NULL,
    receiver_balance_before numeric(36,18) NOT NULL,
    receiver_balance_after numeric(36,18) NOT NULL,
    ip_address inet NOT NULL,
    user_agent text,
    device_type character varying(50),
    device_name character varying(255),
    sender_transaction_log_id uuid,
    receiver_transaction_log_id uuid,
    status character varying(20) DEFAULT 'completed'::character varying NOT NULL,
    error_message text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    brand_id uuid DEFAULT '00000000-0000-0000-0000-000000000001'::uuid NOT NULL,
    CONSTRAINT tip_transactions_amount_check CHECK ((amount > (0)::numeric)),
    CONSTRAINT tip_transactions_fee_check CHECK ((fee >= (0)::numeric)),
    CONSTRAINT tip_transactions_fee_paid_by_check CHECK (((fee_paid_by)::text = ANY ((ARRAY['sender'::character varying, 'receiver'::character varying])::text[]))),
    CONSTRAINT tip_transactions_status_check CHECK (((status)::text = ANY ((ARRAY['completed'::character varying, 'failed'::character varying, 'pending'::character varying])::text[])))
);

CREATE TABLE public.tournaments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    rank text NOT NULL,
    level integer NOT NULL,
    cumulative_points integer NOT NULL,
    rewards jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

CREATE TABLE public.tournaments_claims (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tournament_id uuid NOT NULL,
    squad_id uuid NOT NULL,
    claimed_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE public.transactions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    deposit_session_id character varying(255),
    withdrawal_id character varying(255),
    chain_id text NOT NULL,
    currency_code character varying(10) NOT NULL,
    protocol character varying(10) NOT NULL,
    tx_hash character varying(100) NOT NULL,
    from_address character varying(100) NOT NULL,
    to_address character varying(100) NOT NULL,
    amount numeric(36,18) DEFAULT 0 NOT NULL,
    usd_amount_cents bigint,
    exchange_rate numeric(15,6),
    fee numeric(36,18) DEFAULT 0,
    block_number bigint,
    block_hash character varying(100),
    status character varying(20) DEFAULT 'pending'::character varying NOT NULL,
    confirmations integer DEFAULT 0 NOT NULL,
    "timestamp" timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    verified_at timestamp with time zone,
    processor public.processortype DEFAULT 'internal'::public.processortype NOT NULL,
    transaction_type character varying(20) DEFAULT 'withdrawal'::character varying NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    is_test_transaction boolean DEFAULT true,
    balance_before_cents bigint DEFAULT 0,
    balance_after_cents bigint DEFAULT 0,
    CONSTRAINT transactions_balance_after_cents_check CHECK ((balance_after_cents >= 0)),
    CONSTRAINT transactions_balance_before_cents_check CHECK ((balance_before_cents >= 0)),
    CONSTRAINT transactions_confirmations_check CHECK ((confirmations >= 0)),
    CONSTRAINT transactions_from_address_check CHECK (((from_address)::text <> ''::text)),
    CONSTRAINT transactions_status_check CHECK (((status)::text = ANY (ARRAY[('pending'::character varying)::text, ('verified'::character varying)::text, ('failed'::character varying)::text, ('processing'::character varying)::text]))),
    CONSTRAINT transactions_to_address_check CHECK (((to_address)::text <> ''::text)),
    CONSTRAINT transactions_transaction_type_check CHECK (((transaction_type)::text = ANY (ARRAY[('deposit'::character varying)::text, ('withdrawal'::character varying)::text])))
);

CREATE TABLE public.tucan_game_permissions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    game_id character varying(50) NOT NULL,
    permission_name character varying(100) NOT NULL,
    description text,
    is_active boolean DEFAULT true,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);

CREATE TABLE public.user_2fa_attempts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    attempt_type character varying(50) NOT NULL,
    is_successful boolean NOT NULL,
    ip_address inet,
    user_agent text,
    created_at timestamp without time zone DEFAULT now()
);

COMMENT ON TABLE public.user_2fa_attempts IS 'Audit log for 2FA attempts and rate limiting';

COMMENT ON COLUMN public.user_2fa_attempts.attempt_type IS 'Type of attempt: setup, verify, disable';

CREATE TABLE public.user_2fa_methods (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    method character varying(50) NOT NULL,
    enabled_at timestamp without time zone DEFAULT now(),
    disabled_at timestamp without time zone,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);

COMMENT ON TABLE public.user_2fa_methods IS 'Stores enabled 2FA methods for each user';

COMMENT ON COLUMN public.user_2fa_methods.method IS '2FA method type: totp, email_otp, sms_otp, biometric, backup_codes';

COMMENT ON COLUMN public.user_2fa_methods.enabled_at IS 'When the method was enabled (NULL if disabled)';

COMMENT ON COLUMN public.user_2fa_methods.disabled_at IS 'When the method was disabled (NULL if enabled)';

CREATE TABLE public.user_2fa_otps (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    method character varying(50) NOT NULL,
    otp_code character varying(10) NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now()
);

COMMENT ON TABLE public.user_2fa_otps IS 'Stores temporary OTP codes for 2FA email and SMS methods';

COMMENT ON COLUMN public.user_2fa_otps.method IS '2FA method: email_otp or sms_otp';

COMMENT ON COLUMN public.user_2fa_otps.otp_code IS 'The actual OTP code sent to user';

COMMENT ON COLUMN public.user_2fa_otps.expires_at IS 'When the OTP expires (typically 10 minutes)';

CREATE TABLE public.user_2fa_settings (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    secret_key character varying(255) NOT NULL,
    backup_codes text[],
    is_enabled boolean DEFAULT false,
    enabled_at timestamp without time zone,
    last_used_at timestamp without time zone,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);

COMMENT ON TABLE public.user_2fa_settings IS 'Stores 2FA settings and secrets for users';

COMMENT ON COLUMN public.user_2fa_settings.secret_key IS 'Base32 encoded TOTP secret key';

COMMENT ON COLUMN public.user_2fa_settings.backup_codes IS 'Array of single-use backup codes';

CREATE TABLE public.user_activity_log (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    activity_type character varying(50) NOT NULL,
    activity_data jsonb,
    created_at timestamp with time zone DEFAULT now()
);

CREATE TABLE public.user_allowed_pages (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    page_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

COMMENT ON TABLE public.user_allowed_pages IS 'Junction table mapping users to their allowed pages';

CREATE TABLE public.user_game_access (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    game_id character varying(50) NOT NULL,
    access_type character varying(20) NOT NULL,
    restrictions jsonb,
    expires_at timestamp without time zone,
    created_by uuid,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    CONSTRAINT user_game_access_access_type_check CHECK (((access_type)::text = ANY ((ARRAY['allow'::character varying, 'deny'::character varying, 'restricted'::character varying])::text[])))
);

CREATE TABLE public.user_levels (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    current_level integer DEFAULT 1 NOT NULL,
    total_ggr numeric(20,8) DEFAULT 0 NOT NULL,
    total_bets numeric(20,8) DEFAULT 0 NOT NULL,
    total_wins numeric(20,8) DEFAULT 0 NOT NULL,
    level_progress numeric(5,2) DEFAULT 0 NOT NULL,
    current_tier_id uuid,
    last_level_up timestamp with time zone,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    brand_id uuid DEFAULT '00000000-0000-0000-0000-000000000001'::uuid NOT NULL
);

COMMENT ON COLUMN public.user_levels.brand_id IS 'Brand identifier for multi-brand support - links to brands table';

CREATE TABLE public.user_limits (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    limit_type character varying(50) NOT NULL,
    daily_limit_cents bigint,
    all_time_limit_cents bigint,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    withdrawal_limit_enabled boolean DEFAULT true NOT NULL,
    CONSTRAINT user_limits_limit_type_check CHECK (((limit_type)::text = ANY ((ARRAY['deposit'::character varying, 'withdrawal'::character varying])::text[])))
);

COMMENT ON COLUMN public.user_limits.withdrawal_limit_enabled IS 'Whether the withdrawal limit is currently enabled for this user';

CREATE TABLE public.user_notifications (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    title text NOT NULL,
    content text NOT NULL,
    metadata jsonb,
    read boolean DEFAULT false NOT NULL,
    delivered boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    read_at timestamp with time zone,
    created_by uuid,
    type text NOT NULL,
    CONSTRAINT check_notification_type CHECK ((type = ANY (ARRAY['Promotional'::text, 'KYC'::text, 'Bonus'::text, 'Welcome'::text, 'System'::text, 'Alert'::text, 'payments'::text, 'security'::text, 'general'::text, 'tip'::text])))
);

COMMENT ON CONSTRAINT check_notification_type ON public.user_notifications IS 'Ensures notification type is one of the allowed values including the new tip type';

CREATE TABLE public.user_preferences (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    site_language character varying(10) DEFAULT 'en'::character varying,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);

CREATE TABLE public.user_roles (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    role_id uuid NOT NULL
);

CREATE TABLE public.user_sessions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    token text NOT NULL,
    expires_at timestamp without time zone NOT NULL,
    ip_address character varying(46),
    user_agent character varying(255),
    created_at timestamp without time zone,
    refresh_token text,
    refresh_token_expires_at timestamp without time zone,
    brand_id uuid DEFAULT '00000000-0000-0000-0000-000000000001'::uuid NOT NULL
);

CREATE TABLE public.user_withdrawal_usage (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    volume_period character varying(20) NOT NULL,
    usage_cents bigint DEFAULT 0 NOT NULL,
    period_start timestamp with time zone NOT NULL,
    period_end timestamp with time zone NOT NULL,
    last_reset_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT user_withdrawal_usage_usage_cents_check CHECK ((usage_cents >= 0)),
    CONSTRAINT user_withdrawal_usage_volume_period_check CHECK (((volume_period)::text = ANY ((ARRAY['hourly'::character varying, 'daily'::character varying, 'weekly'::character varying, 'monthly'::character varying, 'yearly'::character varying])::text[])))
);

CREATE TABLE public.users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    username character varying(20),
    phone_number character varying(15),
    password text NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    default_currency character varying(3),
    profile character varying DEFAULT ''::character varying,
    email character varying DEFAULT ''::character varying,
    first_name character varying DEFAULT ''::character varying,
    last_name character varying DEFAULT ''::character varying,
    date_of_birth character varying DEFAULT ''::character varying,
    source character varying DEFAULT ''::character varying,
    is_email_verified boolean DEFAULT false,
    referal_code character varying DEFAULT ''::character varying,
    street_address character varying DEFAULT ''::character varying NOT NULL,
    country character varying DEFAULT ''::character varying NOT NULL,
    state character varying DEFAULT ''::character varying NOT NULL,
    city character varying DEFAULT ''::character varying NOT NULL,
    postal_code character varying DEFAULT ''::character varying NOT NULL,
    kyc_status character varying DEFAULT 'PENDING'::character varying NOT NULL,
    created_by uuid,
    is_admin boolean,
    status character varying DEFAULT 'ACTIVE'::character varying,
    referal_type character varying(255),
    refered_by_code character varying(255),
    user_type character varying(255) DEFAULT 'PLAYER'::character varying,
    primary_wallet_address character varying(255),
    wallet_verification_status character varying(50) DEFAULT 'none'::character varying,
    is_test_account boolean DEFAULT true,
    two_factor_enabled boolean DEFAULT false,
    two_factor_setup_at timestamp without time zone,
    withdrawal_restricted boolean DEFAULT false,
    kyc_required_for_withdrawal boolean DEFAULT true,
    brand_id uuid DEFAULT '00000000-0000-0000-0000-000000000001'::uuid NOT NULL,
    CONSTRAINT users_wallet_verification_status_check CHECK (((wallet_verification_status)::text = ANY (ARRAY[('none'::character varying)::text, ('pending'::character varying)::text, ('verified'::character varying)::text, ('failed'::character varying)::text])))
);

COMMENT ON COLUMN public.users.two_factor_enabled IS 'Whether user has 2FA enabled';

COMMENT ON COLUMN public.users.two_factor_setup_at IS 'When user completed 2FA setup';

COMMENT ON COLUMN public.users.brand_id IS 'Brand identifier - links user to their brand';

CREATE TABLE public.users_football_matche_rounds (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    status character varying DEFAULT 'ACTIVE'::character varying NOT NULL,
    won_status character varying DEFAULT 'PENDING'::character varying,
    user_id uuid NOT NULL,
    football_round_id uuid,
    bet_amount numeric,
    won_amount numeric NOT NULL,
    "timestamp" timestamp without time zone,
    currency character varying DEFAULT 'USD'::character varying NOT NULL
);

CREATE TABLE public.users_football_matches (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    status character varying DEFAULT 'PENDING'::character varying NOT NULL,
    match_id uuid NOT NULL,
    selection character varying DEFAULT 'DRAW'::character varying NOT NULL,
    users_football_matche_round_id uuid
);

CREATE TABLE public.users_otp (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    otp character varying NOT NULL,
    created_at timestamp without time zone NOT NULL
);

CREATE TABLE public.violation_alerts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    game_id character varying(50) NOT NULL,
    violation_type character varying(100) NOT NULL,
    severity character varying(20) NOT NULL,
    message text NOT NULL,
    details jsonb,
    created_at timestamp with time zone DEFAULT now(),
    is_resolved boolean DEFAULT false,
    resolved_at timestamp with time zone,
    resolved_by uuid,
    action_taken text,
    CONSTRAINT violation_alerts_severity_check CHECK (((severity)::text = ANY ((ARRAY['low'::character varying, 'medium'::character varying, 'high'::character varying, 'critical'::character varying])::text[])))
);

CREATE TABLE public.vpn_detection_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    ip_address inet NOT NULL,
    country character varying(100),
    country_code character varying(10),
    is_vpn boolean DEFAULT false,
    vpn_provider character varying(255),
    risk_score double precision DEFAULT 0.0,
    created_at timestamp with time zone DEFAULT now()
);

CREATE TABLE public.waiting_squad_members (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    squad_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE public.wallet_balances (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    wallet_id uuid NOT NULL,
    currency_code character varying(10) NOT NULL,
    balance_amount numeric(36,18) DEFAULT 0 NOT NULL,
    last_deposit_amount numeric(36,18) DEFAULT 0,
    last_deposit_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE public.wallet_transactions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    wallet_id uuid NOT NULL,
    currency_code character varying(10) NOT NULL,
    amount numeric(36,18) NOT NULL,
    balance_before_cents numeric(36,18) NOT NULL,
    balance_after_cents numeric(36,18) NOT NULL,
    tx_hash character varying(100) NOT NULL,
    from_address character varying(255),
    to_address character varying(255),
    block_number bigint,
    confirmations integer DEFAULT 0,
    status public.wallet_transaction_status DEFAULT 'pending'::public.wallet_transaction_status NOT NULL,
    confirmed_at timestamp with time zone,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT wallet_transactions_amount_check CHECK ((amount <> (0)::numeric)),
    CONSTRAINT wallet_transactions_tx_hash_check CHECK (((tx_hash)::text <> ''::text))
);

CREATE TABLE public.wallets (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    chain_id text NOT NULL,
    address text NOT NULL,
    vault_key_path character varying(500) NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    last_used timestamp with time zone,
    is_in_alchemy_webhook boolean DEFAULT false NOT NULL,
    CONSTRAINT wallets_address_check CHECK ((address <> ''::text)),
    CONSTRAINT wallets_vault_key_path_check CHECK (((vault_key_path)::text <> ''::text))
);

CREATE TABLE public.withdrawals (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    admin_id uuid,
    chain_id text NOT NULL,
    currency_code character varying(10) NOT NULL,
    protocol character varying(10) NOT NULL,
    withdrawal_id character varying(255) NOT NULL,
    usd_amount_cents bigint NOT NULL,
    crypto_amount numeric(36,18) NOT NULL,
    exchange_rate numeric(15,6) NOT NULL,
    fee_cents bigint DEFAULT 0 NOT NULL,
    source_wallet_address character varying(255) NOT NULL,
    to_address character varying(255) NOT NULL,
    tx_hash character varying(100),
    status public.withdrawal_status DEFAULT 'pending'::public.withdrawal_status NOT NULL,
    requires_admin_review boolean DEFAULT false NOT NULL,
    admin_review_deadline timestamp with time zone,
    processed_by_system boolean DEFAULT false,
    amount_reserved_cents bigint NOT NULL,
    reservation_released boolean DEFAULT false,
    reservation_released_at timestamp with time zone,
    metadata jsonb,
    error_message text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    first_verification_attempt_at timestamp with time zone,
    verification_attempts integer DEFAULT 0,
    CONSTRAINT withdrawals_amount_reserved_cents_check CHECK ((amount_reserved_cents > 0)),
    CONSTRAINT withdrawals_crypto_amount_check CHECK ((crypto_amount > (0)::numeric)),
    CONSTRAINT withdrawals_exchange_rate_check CHECK ((exchange_rate > (0)::numeric)),
    CONSTRAINT withdrawals_fee_cents_check CHECK ((fee_cents >= 0)),
    CONSTRAINT withdrawals_source_wallet_address_check CHECK (((source_wallet_address)::text <> ''::text)),
    CONSTRAINT withdrawals_to_address_check CHECK (((to_address)::text <> ''::text)),
    CONSTRAINT withdrawals_usd_amount_cents_check CHECK ((usd_amount_cents > 0)),
    CONSTRAINT withdrawals_withdrawal_id_check CHECK (((withdrawal_id)::text <> ''::text))
);

ALTER TABLE ONLY public.casbin_rule ALTER COLUMN id SET DEFAULT nextval('public.casbin_rule_id_seq'::regclass);

ALTER TABLE ONLY public.account_block
    ADD CONSTRAINT account_block_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.active_game_sessions
    ADD CONSTRAINT active_game_sessions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.adds_services
    ADD CONSTRAINT adds_services_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.admin_activity_actions
    ADD CONSTRAINT admin_activity_actions_name_key UNIQUE (name);

ALTER TABLE ONLY public.admin_activity_actions
    ADD CONSTRAINT admin_activity_actions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.admin_activity_categories
    ADD CONSTRAINT admin_activity_categories_name_key UNIQUE (name);

ALTER TABLE ONLY public.admin_activity_categories
    ADD CONSTRAINT admin_activity_categories_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.admin_activity_logs
    ADD CONSTRAINT admin_activity_logs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.admin_fund_movements
    ADD CONSTRAINT admin_fund_movements_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.affiliate_referal_track
    ADD CONSTRAINT affiliate_referal_track_affiliate_code_user_id_key UNIQUE (affiliate_code, user_id);

ALTER TABLE ONLY public.affiliate_referal_track
    ADD CONSTRAINT affiliate_referal_track_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.agent_providers
    ADD CONSTRAINT agent_providers_client_id_key UNIQUE (client_id);

ALTER TABLE ONLY public.agent_providers
    ADD CONSTRAINT agent_providers_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.agent_referrals
    ADD CONSTRAINT agent_referrals_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.agent_referrals
    ADD CONSTRAINT agent_referrals_request_id_key UNIQUE (request_id);

ALTER TABLE ONLY public.airtime_transactions
    ADD CONSTRAINT airtime_transactions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.airtime_utilities
    ADD CONSTRAINT airtime_utilities_pkey PRIMARY KEY (local_id);

ALTER TABLE ONLY public.alchemy_webhooks
    ADD CONSTRAINT alchemy_webhooks_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.alchemy_webhooks
    ADD CONSTRAINT alchemy_webhooks_webhook_id_key UNIQUE (webhook_id);

ALTER TABLE ONLY public.alert_configurations
    ADD CONSTRAINT alert_configurations_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.alert_email_group_members
    ADD CONSTRAINT alert_email_group_members_group_id_email_key UNIQUE (group_id, email);

ALTER TABLE ONLY public.alert_email_group_members
    ADD CONSTRAINT alert_email_group_members_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.alert_email_groups
    ADD CONSTRAINT alert_email_groups_name_key UNIQUE (name);

ALTER TABLE ONLY public.alert_email_groups
    ADD CONSTRAINT alert_email_groups_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.alert_rules
    ADD CONSTRAINT alert_rules_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.alert_triggers
    ADD CONSTRAINT alert_triggers_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.balance_logs
    ADD CONSTRAINT balance_logs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.balances
    ADD CONSTRAINT balances_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.banners
    ADD CONSTRAINT banners_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.bets
    ADD CONSTRAINT bets_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.brands
    ADD CONSTRAINT brands_code_key UNIQUE (code);

ALTER TABLE ONLY public.brands
    ADD CONSTRAINT brands_name_key UNIQUE (name);

ALTER TABLE ONLY public.brands
    ADD CONSTRAINT brands_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.campaign_recipients
    ADD CONSTRAINT campaign_recipients_campaign_id_user_id_key UNIQUE (campaign_id, user_id);

ALTER TABLE ONLY public.campaign_recipients
    ADD CONSTRAINT campaign_recipients_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.casbin_rule
    ADD CONSTRAINT casbin_rule_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.cashback_claims
    ADD CONSTRAINT cashback_claims_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.cashback_earnings
    ADD CONSTRAINT cashback_earnings_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.cashback_tiers
    ADD CONSTRAINT cashback_tiers_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.cashback_tiers
    ADD CONSTRAINT cashback_tiers_tier_level_key UNIQUE (tier_level);

ALTER TABLE ONLY public.cashback_tiers
    ADD CONSTRAINT cashback_tiers_tier_name_key UNIQUE (tier_name);

ALTER TABLE ONLY public.chain_currencies
    ADD CONSTRAINT chain_currencies_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.chain_processing_state
    ADD CONSTRAINT chain_processing_state_chain_id_key UNIQUE (chain_id);

ALTER TABLE ONLY public.chain_processing_state
    ADD CONSTRAINT chain_processing_state_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.clubs
    ADD CONSTRAINT clubs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.company
    ADD CONSTRAINT company_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.configs
    ADD CONSTRAINT configs_name_key UNIQUE (name);

ALTER TABLE ONLY public.configs
    ADD CONSTRAINT configs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.conversion_remainders
    ADD CONSTRAINT conversion_remainders_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.crypto_kings
    ADD CONSTRAINT crypto_kings_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.crypto_wallet_auth_logs
    ADD CONSTRAINT crypto_wallet_auth_logs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.crypto_wallet_challenges
    ADD CONSTRAINT crypto_wallet_challenges_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.crypto_wallet_connections
    ADD CONSTRAINT crypto_wallet_connections_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.crypto_wallet_connections
    ADD CONSTRAINT crypto_wallet_connections_user_id_wallet_address_key UNIQUE (user_id, wallet_address);

ALTER TABLE ONLY public.crypto_wallet_connections
    ADD CONSTRAINT crypto_wallet_connections_wallet_address_wallet_type_key UNIQUE (wallet_address, wallet_type);

ALTER TABLE ONLY public.currencies
    ADD CONSTRAINT currencies_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.currency_config
    ADD CONSTRAINT currency_config_pkey PRIMARY KEY (currency_code);

ALTER TABLE ONLY public.departements_users
    ADD CONSTRAINT departements_users_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.departments
    ADD CONSTRAINT departments_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.deposit_event_records
    ADD CONSTRAINT deposit_event_records_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.deposit_sessions
    ADD CONSTRAINT deposit_sessions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.deposit_sessions
    ADD CONSTRAINT deposit_sessions_session_id_key UNIQUE (session_id);

ALTER TABLE ONLY public.emergency_access_overrides
    ADD CONSTRAINT emergency_access_overrides_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.exchange_rates
    ADD CONSTRAINT exchange_rates_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.failed_bet_logs
    ADD CONSTRAINT failed_bet_logs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.falcon_liquidity_messages
    ADD CONSTRAINT falcon_liquidity_messages_message_id_key UNIQUE (message_id);

ALTER TABLE ONLY public.falcon_liquidity_messages
    ADD CONSTRAINT falcon_liquidity_messages_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.football_match_rounds
    ADD CONSTRAINT football_match_rounds_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.football_matchs
    ADD CONSTRAINT football_matchs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.game_access_config
    ADD CONSTRAINT game_access_config_config_key_key UNIQUE (config_key);

ALTER TABLE ONLY public.game_access_config
    ADD CONSTRAINT game_access_config_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.game_access_logs
    ADD CONSTRAINT game_access_logs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.game_access_templates
    ADD CONSTRAINT game_access_templates_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.game_cashback_rates
    ADD CONSTRAINT game_cashback_rates_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.game_house_edges
    ADD CONSTRAINT game_house_edges_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.game_logs
    ADD CONSTRAINT game_logs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.game_permissions
    ADD CONSTRAINT game_permissions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.game_sessions
    ADD CONSTRAINT game_sessions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.game_sessions
    ADD CONSTRAINT game_sessions_session_id_key UNIQUE (session_id);

ALTER TABLE ONLY public.games
    ADD CONSTRAINT games_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.geographic_restrictions
    ADD CONSTRAINT geographic_restrictions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.global_rakeback_override
    ADD CONSTRAINT global_rakeback_override_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.groove_accounts
    ADD CONSTRAINT groove_accounts_account_id_key UNIQUE (account_id);

ALTER TABLE ONLY public.groove_accounts
    ADD CONSTRAINT groove_accounts_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.groove_game_sessions
    ADD CONSTRAINT groove_game_sessions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.groove_game_sessions
    ADD CONSTRAINT groove_game_sessions_session_id_key UNIQUE (session_id);

ALTER TABLE ONLY public.groove_transactions
    ADD CONSTRAINT groove_transactions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.groove_transactions
    ADD CONSTRAINT groove_transactions_transaction_id_key UNIQUE (transaction_id);

ALTER TABLE ONLY public.ip_filters
    ADD CONSTRAINT ip_filters_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.kyc_documents
    ADD CONSTRAINT kyc_documents_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.kyc_settings
    ADD CONSTRAINT kyc_settings_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.kyc_settings
    ADD CONSTRAINT kyc_settings_setting_key_key UNIQUE (setting_key);

ALTER TABLE ONLY public.kyc_status_changes
    ADD CONSTRAINT kyc_status_changes_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.kyc_submissions
    ADD CONSTRAINT kyc_submissions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.leagues
    ADD CONSTRAINT leagues_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.level_requirements
    ADD CONSTRAINT level_requirements_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.levels
    ADD CONSTRAINT levels_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.login_attempts
    ADD CONSTRAINT login_attempts_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.logs
    ADD CONSTRAINT logs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.loot_box
    ADD CONSTRAINT loot_box_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.loot_box_place_bets
    ADD CONSTRAINT loot_box_place_bets_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.lotteries
    ADD CONSTRAINT lotteries_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.lottery_logs
    ADD CONSTRAINT lottery_logs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.lottery_services
    ADD CONSTRAINT lottery_services_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.lottery_winners_logs
    ADD CONSTRAINT lottery_winners_logs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.manual_funds
    ADD CONSTRAINT manual_funds_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.message_campaigns
    ADD CONSTRAINT message_campaigns_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.message_segments
    ADD CONSTRAINT message_segments_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.monitoring_reports
    ADD CONSTRAINT monitoring_reports_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.operational_groups
    ADD CONSTRAINT operational_groups_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.operational_types
    ADD CONSTRAINT operational_types_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.otps
    ADD CONSTRAINT otps_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.pages
    ADD CONSTRAINT pages_path_key UNIQUE (path);

ALTER TABLE ONLY public.pages
    ADD CONSTRAINT pages_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.passkey_credentials
    ADD CONSTRAINT passkey_credentials_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.passkey_credentials
    ADD CONSTRAINT passkey_credentials_user_id_credential_id_key UNIQUE (user_id, credential_id);

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.plinko
    ADD CONSTRAINT plinko_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.public_winner_notifications
    ADD CONSTRAINT public_winner_notifications_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.quick_hustles
    ADD CONSTRAINT quick_hustles_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.rakeback_schedules
    ADD CONSTRAINT rakeback_schedules_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.restriction_violations
    ADD CONSTRAINT restriction_violations_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.retryable_operations
    ADD CONSTRAINT retryable_operations_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.risk_settings
    ADD CONSTRAINT risk_settings_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.role_game_access
    ADD CONSTRAINT role_game_access_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.roll_da_dice
    ADD CONSTRAINT roll_da_dice_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.rounds
    ADD CONSTRAINT rounds_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);

ALTER TABLE ONLY public.scratch_cards
    ADD CONSTRAINT scratch_cards_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.service_api_keys
    ADD CONSTRAINT service_api_keys_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.session_limits
    ADD CONSTRAINT session_limits_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.spending_limits
    ADD CONSTRAINT spending_limits_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.spending_tracking
    ADD CONSTRAINT spending_tracking_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.spinning_wheel_configs
    ADD CONSTRAINT spinning_wheel_configs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.spinning_wheel_mysteries
    ADD CONSTRAINT spinning_wheel_mysteries_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.spinning_wheel_rewards
    ADD CONSTRAINT spinning_wheel_rewards_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.spinning_wheels
    ADD CONSTRAINT spinning_wheels_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.sport_bets
    ADD CONSTRAINT sport_bets_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.sport_bets
    ADD CONSTRAINT sport_bets_transaction_id_key UNIQUE (transaction_id);

ALTER TABLE ONLY public.squads_earns
    ADD CONSTRAINT squads_earns_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.squads_memebers
    ADD CONSTRAINT squads_memebers_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.squads
    ADD CONSTRAINT squads_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.street_kings
    ADD CONSTRAINT street_kings_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.supported_chains
    ADD CONSTRAINT supported_chains_chain_id_key UNIQUE (chain_id);

ALTER TABLE ONLY public.supported_chains
    ADD CONSTRAINT supported_chains_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.system_config
    ADD CONSTRAINT system_config_brand_id_config_key_unique UNIQUE (brand_id, config_key);

ALTER TABLE ONLY public.system_config
    ADD CONSTRAINT system_config_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.temp
    ADD CONSTRAINT temp_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.time_restrictions
    ADD CONSTRAINT time_restrictions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.tip_transactions
    ADD CONSTRAINT tip_transactions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.tip_transactions
    ADD CONSTRAINT tip_transactions_transaction_id_key UNIQUE (transaction_id);

ALTER TABLE ONLY public.tournaments_claims
    ADD CONSTRAINT tournaments_claims_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.tournaments
    ADD CONSTRAINT tournaments_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.tucan_game_permissions
    ADD CONSTRAINT tucan_game_permissions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.wallets
    ADD CONSTRAINT unique_chain_address UNIQUE (chain_id, address);

ALTER TABLE ONLY public.chain_currencies
    ADD CONSTRAINT unique_chain_currency UNIQUE (chain_id, currency_code);

ALTER TABLE ONLY public.alchemy_webhooks
    ADD CONSTRAINT unique_chain_webhook UNIQUE (chain_id, webhook_id);

ALTER TABLE ONLY public.exchange_rates
    ADD CONSTRAINT unique_currency_pair UNIQUE (from_currency, to_currency);

ALTER TABLE ONLY public.banners
    ADD CONSTRAINT unique_page UNIQUE (page);

ALTER TABLE ONLY public.service_api_keys
    ADD CONSTRAINT unique_service_key UNIQUE (issuer_service, key);

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT unique_tx_hash UNIQUE (tx_hash);

ALTER TABLE ONLY public.balances
    ADD CONSTRAINT unique_user_currency UNIQUE (user_id, currency_code);

ALTER TABLE ONLY public.player_excluded_games
    ADD CONSTRAINT unique_user_game_exclusion UNIQUE (user_id, game_id);

ALTER TABLE ONLY public.wallet_balances
    ADD CONSTRAINT unique_wallet_currency UNIQUE (wallet_id, currency_code);

ALTER TABLE ONLY public.user_allowed_pages
    ADD CONSTRAINT uq_user_allowed_pages UNIQUE (user_id, page_id);

ALTER TABLE ONLY public.user_2fa_attempts
    ADD CONSTRAINT user_2fa_attempts_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.user_2fa_methods
    ADD CONSTRAINT user_2fa_methods_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.user_2fa_methods
    ADD CONSTRAINT user_2fa_methods_user_id_method_key UNIQUE (user_id, method);

ALTER TABLE ONLY public.user_2fa_otps
    ADD CONSTRAINT user_2fa_otps_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.user_2fa_otps
    ADD CONSTRAINT user_2fa_otps_user_id_method_key UNIQUE (user_id, method);

ALTER TABLE ONLY public.user_2fa_settings
    ADD CONSTRAINT user_2fa_settings_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.user_2fa_settings
    ADD CONSTRAINT user_2fa_settings_user_id_key UNIQUE (user_id);

ALTER TABLE ONLY public.user_activity_log
    ADD CONSTRAINT user_activity_log_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.user_allowed_pages
    ADD CONSTRAINT user_allowed_pages_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.user_game_access
    ADD CONSTRAINT user_game_access_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.user_levels
    ADD CONSTRAINT user_levels_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.user_levels
    ADD CONSTRAINT user_levels_user_id_key UNIQUE (user_id);

ALTER TABLE ONLY public.user_limits
    ADD CONSTRAINT user_limits_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.user_limits
    ADD CONSTRAINT user_limits_user_id_limit_type_key UNIQUE (user_id, limit_type);

ALTER TABLE ONLY public.user_notifications
    ADD CONSTRAINT user_notifications_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.user_preferences
    ADD CONSTRAINT user_preferences_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.user_preferences
    ADD CONSTRAINT user_preferences_user_id_key UNIQUE (user_id);

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.user_sessions
    ADD CONSTRAINT user_sessions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.user_withdrawal_usage
    ADD CONSTRAINT user_withdrawal_usage_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.user_withdrawal_usage
    ADD CONSTRAINT user_withdrawal_usage_user_id_volume_period_period_start_key UNIQUE (user_id, volume_period, period_start);

ALTER TABLE ONLY public.users_football_matche_rounds
    ADD CONSTRAINT users_football_matche_rounds_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.users_football_matches
    ADD CONSTRAINT users_football_matches_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.users_otp
    ADD CONSTRAINT users_otp_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_phone_number_key UNIQUE (phone_number);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_username_key UNIQUE (username);

ALTER TABLE ONLY public.violation_alerts
    ADD CONSTRAINT violation_alerts_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.waiting_squad_members
    ADD CONSTRAINT waiting_squad_members_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.wallet_balances
    ADD CONSTRAINT wallet_balances_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.wallet_transactions
    ADD CONSTRAINT wallet_transactions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.wallets
    ADD CONSTRAINT wallets_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.withdrawals
    ADD CONSTRAINT withdrawals_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.withdrawals
    ADD CONSTRAINT withdrawals_withdrawal_id_key UNIQUE (withdrawal_id);

CREATE UNIQUE INDEX game_house_edges_game_type_game_variant_game_id_effective_from_ ON public.game_house_edges USING btree (game_type, game_variant, game_id, effective_from);

CREATE INDEX idx_active_sessions_token ON public.active_game_sessions USING btree (session_token);

CREATE INDEX idx_active_sessions_user ON public.active_game_sessions USING btree (user_id);

CREATE INDEX idx_adds_services_created_at ON public.adds_services USING btree (created_at);

CREATE INDEX idx_adds_services_status ON public.adds_services USING btree (status);

CREATE INDEX idx_admin_activity_logs_action ON public.admin_activity_logs USING btree (action);

CREATE INDEX idx_admin_activity_logs_admin_user_id ON public.admin_activity_logs USING btree (admin_user_id);

CREATE INDEX idx_admin_activity_logs_category ON public.admin_activity_logs USING btree (category);

CREATE INDEX idx_admin_activity_logs_composite ON public.admin_activity_logs USING btree (admin_user_id, created_at DESC);

CREATE INDEX idx_admin_activity_logs_created_at ON public.admin_activity_logs USING btree (created_at);

CREATE INDEX idx_admin_activity_logs_resource_id ON public.admin_activity_logs USING btree (resource_id);

CREATE INDEX idx_admin_activity_logs_resource_type ON public.admin_activity_logs USING btree (resource_type);

CREATE INDEX idx_admin_activity_logs_severity ON public.admin_activity_logs USING btree (severity);

CREATE INDEX idx_admin_fund_movements_admin_id ON public.admin_fund_movements USING btree (admin_id);

CREATE INDEX idx_admin_fund_movements_from_address ON public.admin_fund_movements USING btree (from_address);

CREATE INDEX idx_admin_fund_movements_status ON public.admin_fund_movements USING btree (status);

CREATE INDEX idx_affiliate_referal_track_affiliate_code ON public.affiliate_referal_track USING btree (affiliate_code);

CREATE INDEX idx_affiliate_referal_track_created_at ON public.affiliate_referal_track USING btree (created_at);

CREATE INDEX idx_affiliate_referal_track_deleted_at ON public.affiliate_referal_track USING btree (deleted_at) WHERE (deleted_at IS NULL);

CREATE INDEX idx_affiliate_referal_track_user_id ON public.affiliate_referal_track USING btree (user_id);

CREATE INDEX idx_agent_referrals_callback_attempts ON public.agent_referrals USING btree (callback_attempts);

CREATE INDEX idx_agent_referrals_callback_sent ON public.agent_referrals USING btree (callback_sent);

CREATE INDEX idx_agent_referrals_converted_at ON public.agent_referrals USING btree (converted_at);

CREATE INDEX idx_agent_referrals_request_id ON public.agent_referrals USING btree (request_id);

CREATE INDEX idx_agent_referrals_user_id ON public.agent_referrals USING btree (user_id);

CREATE INDEX idx_alert_configurations_email_group_ids ON public.alert_configurations USING gin (email_group_ids);

CREATE INDEX idx_alert_configurations_status ON public.alert_configurations USING btree (status);

CREATE INDEX idx_alert_configurations_type ON public.alert_configurations USING btree (alert_type);

CREATE UNIQUE INDEX idx_alert_configurations_type_unique_active ON public.alert_configurations USING btree (alert_type) WHERE (status = 'active'::public.alert_status);

COMMENT ON INDEX public.idx_alert_configurations_type_unique_active IS 'Ensures only one active alert configuration exists per alert_type';

CREATE INDEX idx_alert_email_group_members_email ON public.alert_email_group_members USING btree (email);

CREATE INDEX idx_alert_email_group_members_group_id ON public.alert_email_group_members USING btree (group_id);

CREATE INDEX idx_alert_email_groups_created_by ON public.alert_email_groups USING btree (created_by);

CREATE INDEX idx_alert_email_groups_name ON public.alert_email_groups USING btree (name);

CREATE INDEX idx_alert_triggers_acknowledged ON public.alert_triggers USING btree (acknowledged);

CREATE INDEX idx_alert_triggers_config_id ON public.alert_triggers USING btree (alert_configuration_id);

CREATE INDEX idx_alert_triggers_triggered_at ON public.alert_triggers USING btree (triggered_at);

CREATE INDEX idx_alert_triggers_user_id ON public.alert_triggers USING btree (user_id);

CREATE INDEX idx_audit_logs_action ON public.audit_logs USING btree (action);

CREATE INDEX idx_audit_logs_created_at ON public.audit_logs USING btree (created_at DESC);

CREATE INDEX idx_audit_logs_entity_type ON public.audit_logs USING btree (entity_type);

CREATE INDEX idx_balance_logs_brand_id ON public.balance_logs USING btree (brand_id);

CREATE INDEX idx_balance_logs_brand_user ON public.balance_logs USING btree (brand_id, user_id);

CREATE INDEX idx_balances_brand_id ON public.balances USING btree (brand_id);

CREATE INDEX idx_balances_brand_user ON public.balances USING btree (brand_id, user_id);

CREATE INDEX idx_bets_brand_id ON public.bets USING btree (brand_id);

CREATE INDEX idx_bets_brand_user ON public.bets USING btree (brand_id, user_id);

CREATE INDEX idx_brands_code ON public.brands USING btree (code);

CREATE INDEX idx_brands_is_active ON public.brands USING btree (is_active);

CREATE INDEX idx_campaign_recipients_campaign_id ON public.campaign_recipients USING btree (campaign_id);

CREATE INDEX idx_campaign_recipients_notification_id ON public.campaign_recipients USING btree (notification_id);

CREATE INDEX idx_campaign_recipients_status ON public.campaign_recipients USING btree (status);

CREATE INDEX idx_campaign_recipients_user_id ON public.campaign_recipients USING btree (user_id);

CREATE UNIQUE INDEX idx_casbin_rule ON public.casbin_rule USING btree (ptype, v0, v1, v2, v3, v4, v5);

CREATE INDEX idx_cashback_claims_status ON public.cashback_claims USING btree (status);

CREATE INDEX idx_cashback_claims_user_id ON public.cashback_claims USING btree (user_id);

CREATE INDEX idx_cashback_earnings_brand_id ON public.cashback_earnings USING btree (brand_id);

CREATE INDEX idx_cashback_earnings_brand_user ON public.cashback_earnings USING btree (brand_id, user_id);

CREATE INDEX idx_cashback_earnings_created_at ON public.cashback_earnings USING btree (created_at);

CREATE INDEX idx_cashback_earnings_expires_at ON public.cashback_earnings USING btree (expires_at);

CREATE INDEX idx_cashback_earnings_status ON public.cashback_earnings USING btree (status);

CREATE INDEX idx_cashback_earnings_user_id ON public.cashback_earnings USING btree (user_id);

CREATE INDEX idx_chain_currencies_chain_id ON public.chain_currencies USING btree (chain_id);

CREATE INDEX idx_chain_currencies_currency ON public.chain_currencies USING btree (currency_code);

CREATE INDEX idx_conversion_remainders_transaction_id ON public.conversion_remainders USING btree (transaction_id);

CREATE INDEX idx_crypto_wallet_auth_logs_created_at ON public.crypto_wallet_auth_logs USING btree (created_at);

CREATE INDEX idx_crypto_wallet_auth_logs_wallet_address ON public.crypto_wallet_auth_logs USING btree (wallet_address);

CREATE INDEX idx_crypto_wallet_challenges_expires_at ON public.crypto_wallet_challenges USING btree (expires_at);

CREATE INDEX idx_crypto_wallet_challenges_wallet_address ON public.crypto_wallet_challenges USING btree (wallet_address);

CREATE INDEX idx_crypto_wallet_connections_user_id ON public.crypto_wallet_connections USING btree (user_id);

CREATE INDEX idx_crypto_wallet_connections_wallet_address ON public.crypto_wallet_connections USING btree (wallet_address);

CREATE INDEX idx_crypto_wallet_connections_wallet_type ON public.crypto_wallet_connections USING btree (wallet_type);

CREATE INDEX idx_deposit_sessions_session_id ON public.deposit_sessions USING btree (session_id);

CREATE INDEX idx_deposit_sessions_status ON public.deposit_sessions USING btree (status);

CREATE INDEX idx_deposit_sessions_status_created_at ON public.deposit_sessions USING btree (status, created_at) WHERE (status = ANY (ARRAY['pending'::public.deposit_session_status, 'processing'::public.deposit_session_status]));

CREATE INDEX idx_deposit_sessions_user_id ON public.deposit_sessions USING btree (user_id);

CREATE INDEX idx_emergency_access_overrides_is_active ON public.emergency_access_overrides USING btree (is_active);

CREATE INDEX idx_emergency_access_overrides_user_id ON public.emergency_access_overrides USING btree (user_id);

CREATE INDEX idx_falcon_messages_created_at ON public.falcon_liquidity_messages USING btree (created_at);

CREATE INDEX idx_falcon_messages_message_id ON public.falcon_liquidity_messages USING btree (message_id);

CREATE INDEX idx_falcon_messages_message_type ON public.falcon_liquidity_messages USING btree (message_type);

CREATE INDEX idx_falcon_messages_reconciliation_status ON public.falcon_liquidity_messages USING btree (reconciliation_status);

CREATE INDEX idx_falcon_messages_status ON public.falcon_liquidity_messages USING btree (status);

CREATE INDEX idx_falcon_messages_transaction_id ON public.falcon_liquidity_messages USING btree (transaction_id);

CREATE INDEX idx_falcon_messages_transaction_status ON public.falcon_liquidity_messages USING btree (transaction_id, status);

CREATE INDEX idx_falcon_messages_user_id ON public.falcon_liquidity_messages USING btree (user_id);

CREATE INDEX idx_falcon_messages_user_status ON public.falcon_liquidity_messages USING btree (user_id, status);

CREATE INDEX idx_game_access_logs_created_at ON public.game_access_logs USING btree (created_at);

CREATE INDEX idx_game_access_logs_game_id ON public.game_access_logs USING btree (game_id);

CREATE INDEX idx_game_access_logs_user_id ON public.game_access_logs USING btree (user_id);

CREATE INDEX idx_game_cashback_rates_active ON public.game_cashback_rates USING btree (is_active);

CREATE INDEX idx_game_cashback_rates_game_id ON public.game_cashback_rates USING btree (game_id);

CREATE INDEX idx_game_cashback_rates_game_type ON public.game_cashback_rates USING btree (game_type);

CREATE INDEX idx_game_house_edges_game_id ON public.game_house_edges USING btree (game_id);

CREATE INDEX idx_game_house_edges_game_type ON public.game_house_edges USING btree (game_type);

CREATE INDEX idx_game_house_edges_is_active ON public.game_house_edges USING btree (is_active);

CREATE INDEX idx_game_permissions_game_id ON public.game_permissions USING btree (game_id);

CREATE INDEX idx_game_permissions_is_active ON public.game_permissions USING btree (is_active);

CREATE INDEX idx_game_sessions_active ON public.game_sessions USING btree (is_active);

CREATE INDEX idx_game_sessions_created_at ON public.game_sessions USING btree (created_at);

CREATE INDEX idx_game_sessions_game_id ON public.game_sessions USING btree (game_id);

CREATE INDEX idx_game_sessions_session_id ON public.game_sessions USING btree (session_id);

CREATE INDEX idx_game_sessions_user_id ON public.game_sessions USING btree (user_id);

CREATE INDEX idx_games_created_at ON public.games USING btree (created_at);

CREATE INDEX idx_games_game_id ON public.games USING btree (game_id);

CREATE INDEX idx_geographic_restrictions_role_game ON public.geographic_restrictions USING btree (role_id, game_id) WHERE ((restriction_type)::text = 'role'::text);

CREATE INDEX idx_geographic_restrictions_user_game ON public.geographic_restrictions USING btree (user_id, game_id) WHERE ((restriction_type)::text = 'user'::text);

CREATE INDEX idx_global_rakeback_override_active ON public.global_rakeback_override USING btree (is_active) WHERE (is_active = true);

CREATE INDEX idx_groove_accounts_brand_id ON public.groove_accounts USING btree (brand_id);

CREATE INDEX idx_groove_accounts_brand_user ON public.groove_accounts USING btree (brand_id, user_id);

CREATE INDEX idx_groove_accounts_status ON public.groove_accounts USING btree (status);

CREATE INDEX idx_groove_accounts_user_id ON public.groove_accounts USING btree (user_id);

CREATE INDEX idx_groove_game_sessions_brand_id ON public.groove_game_sessions USING btree (brand_id);

CREATE INDEX idx_groove_game_sessions_is_test ON public.groove_game_sessions USING btree (is_test_game_session);

CREATE INDEX idx_groove_game_sessions_user_id ON public.groove_game_sessions USING btree (user_id);

CREATE INDEX idx_groove_transactions_account_created ON public.groove_transactions USING btree (account_id, created_at);

CREATE INDEX idx_groove_transactions_account_id ON public.groove_transactions USING btree (account_id);

CREATE INDEX idx_groove_transactions_brand_created ON public.groove_transactions USING btree (brand_id, created_at);

CREATE INDEX idx_groove_transactions_brand_id ON public.groove_transactions USING btree (brand_id);

CREATE INDEX idx_groove_transactions_brand_user ON public.groove_transactions USING btree (brand_id, user_id);

CREATE INDEX idx_groove_transactions_created_at ON public.groove_transactions USING btree (created_at);

CREATE INDEX idx_groove_transactions_game_id ON public.groove_transactions USING btree (game_id);

CREATE INDEX idx_groove_transactions_lookup ON public.groove_transactions USING btree (transaction_id, type, session_id, round_id) WHERE ((type)::text = 'wager'::text);

COMMENT ON INDEX public.idx_groove_transactions_lookup IS 'Composite index for fast wager lookup during result processing. Optimized for high-throughput casino operations.';

CREATE INDEX idx_groove_transactions_round_id ON public.groove_transactions USING btree (round_id);

CREATE INDEX idx_groove_transactions_session_id ON public.groove_transactions USING btree (session_id);

CREATE INDEX idx_groove_transactions_transaction_id ON public.groove_transactions USING btree (transaction_id);

CREATE INDEX idx_groove_transactions_user_id ON public.groove_transactions USING btree (user_id);

CREATE INDEX idx_kyc_documents_status ON public.kyc_documents USING btree (status);

CREATE INDEX idx_kyc_documents_type ON public.kyc_documents USING btree (document_type);

CREATE INDEX idx_kyc_documents_user_id ON public.kyc_documents USING btree (user_id);

CREATE INDEX idx_kyc_status_changes_user_id ON public.kyc_status_changes USING btree (user_id);

CREATE INDEX idx_kyc_submissions_status ON public.kyc_submissions USING btree (status);

CREATE INDEX idx_kyc_submissions_user_id ON public.kyc_submissions USING btree (user_id);

CREATE INDEX idx_login_attempts_brand_id ON public.login_attempts USING btree (brand_id);

CREATE INDEX idx_login_attempts_brand_user ON public.login_attempts USING btree (brand_id, user_id);

CREATE INDEX idx_message_campaigns_created_by ON public.message_campaigns USING btree (created_by);

CREATE INDEX idx_message_campaigns_scheduled_at ON public.message_campaigns USING btree (scheduled_at);

CREATE INDEX idx_message_campaigns_status ON public.message_campaigns USING btree (status);

CREATE INDEX idx_message_segments_campaign_id ON public.message_segments USING btree (campaign_id);

CREATE INDEX idx_message_segments_segment_type ON public.message_segments USING btree (segment_type);

CREATE INDEX idx_otps_created_at ON public.otps USING btree (created_at);

CREATE INDEX idx_otps_email ON public.otps USING btree (email);

CREATE INDEX idx_otps_email_type_created ON public.otps USING btree (email, type, created_at DESC);

CREATE INDEX idx_otps_expires_at ON public.otps USING btree (expires_at);

CREATE INDEX idx_otps_status ON public.otps USING btree (status);

CREATE INDEX idx_otps_type ON public.otps USING btree (type);

CREATE INDEX idx_pages_parent_id ON public.pages USING btree (parent_id);

CREATE INDEX idx_pages_path ON public.pages USING btree (path);

CREATE INDEX idx_passkey_credentials_active ON public.passkey_credentials USING btree (user_id, is_active) WHERE (is_active = true);

CREATE INDEX idx_passkey_credentials_credential_id ON public.passkey_credentials USING btree (credential_id);

CREATE INDEX idx_passkey_credentials_user_id ON public.passkey_credentials USING btree (user_id);

CREATE INDEX idx_player_deposit_tracking_user_period ON public.player_deposit_tracking USING btree (user_id, period_type, period_start, period_end);

CREATE INDEX idx_player_excluded_games_game_id ON public.player_excluded_games USING btree (game_id);

CREATE INDEX idx_player_excluded_games_user_game ON public.player_excluded_games USING btree (user_id, game_id);

CREATE INDEX idx_player_excluded_games_user_id ON public.player_excluded_games USING btree (user_id);

CREATE INDEX idx_player_gaming_time_session ON public.player_gaming_time_tracking USING btree (session_id) WHERE (session_id IS NOT NULL);

CREATE INDEX idx_player_gaming_time_user_period ON public.player_gaming_time_tracking USING btree (user_id, period_type, period_start, period_end);

CREATE INDEX idx_player_self_protection_activity_logs_created_at ON public.player_self_protection_activity_logs USING btree (created_at DESC);

CREATE INDEX idx_player_self_protection_activity_logs_user_created ON public.player_self_protection_activity_logs USING btree (user_id, created_at DESC);

CREATE INDEX idx_player_self_protection_activity_logs_user_id ON public.player_self_protection_activity_logs USING btree (user_id);

CREATE INDEX idx_player_self_protection_self_exclusion ON public.player_self_protection_settings USING btree (self_exclusion_enabled, self_exclusion_end_date) WHERE (self_exclusion_enabled = true);

CREATE INDEX idx_player_self_protection_user_id ON public.player_self_protection_settings USING btree (user_id);

CREATE INDEX idx_public_winner_notifications_expires_at ON public.public_winner_notifications USING btree (expires_at);

CREATE INDEX idx_public_winner_notifications_win_amount ON public.public_winner_notifications USING btree (win_amount DESC);

CREATE INDEX idx_rakeback_schedules_created_by ON public.rakeback_schedules USING btree (created_by);

CREATE INDEX idx_rakeback_schedules_end_time ON public.rakeback_schedules USING btree (end_time);

CREATE INDEX idx_rakeback_schedules_scheduler ON public.rakeback_schedules USING btree (status, start_time, end_time);

CREATE INDEX idx_rakeback_schedules_scope_type ON public.rakeback_schedules USING btree (scope_type);

CREATE INDEX idx_rakeback_schedules_start_time ON public.rakeback_schedules USING btree (start_time);

CREATE INDEX idx_rakeback_schedules_status ON public.rakeback_schedules USING btree (status);

CREATE INDEX idx_role_game_access_game_id ON public.role_game_access USING btree (game_id);

CREATE INDEX idx_role_game_access_role_id ON public.role_game_access USING btree (role_id);

CREATE INDEX idx_role_permissions_value ON public.role_permissions USING btree (value) WHERE (value IS NOT NULL);

CREATE INDEX idx_session_limits_role_game ON public.session_limits USING btree (role_id, game_id) WHERE ((limit_type)::text = 'role'::text);

CREATE INDEX idx_session_limits_user_game ON public.session_limits USING btree (user_id, game_id) WHERE ((limit_type)::text = 'user'::text);

CREATE INDEX idx_spending_limits_role_game ON public.spending_limits USING btree (role_id, game_id) WHERE ((limit_type)::text = 'role'::text);

CREATE INDEX idx_spending_limits_user_game ON public.spending_limits USING btree (user_id, game_id) WHERE ((limit_type)::text = 'user'::text);

CREATE INDEX idx_spending_tracking_user_period ON public.spending_tracking USING btree (user_id, game_id, period_start, period_end);

CREATE INDEX idx_sport_bets_bet_status ON public.sport_bets USING btree (status);

CREATE INDEX idx_sport_bets_brand_created ON public.sport_bets USING btree (brand_id, created_at);

CREATE INDEX idx_sport_bets_brand_id ON public.sport_bets USING btree (brand_id);

CREATE INDEX idx_sport_bets_brand_user ON public.sport_bets USING btree (brand_id, user_id);

CREATE INDEX idx_sport_bets_placed_at ON public.sport_bets USING btree (placed_at);

CREATE INDEX idx_sport_bets_transaction_id ON public.sport_bets USING btree (transaction_id);

CREATE INDEX idx_sport_bets_user_id ON public.sport_bets USING btree (user_id);

CREATE INDEX idx_squads_type ON public.squads USING btree (type);

CREATE INDEX idx_supported_chains_chain_id ON public.supported_chains USING btree (chain_id);

CREATE INDEX idx_supported_chains_status ON public.supported_chains USING btree (status);

CREATE INDEX idx_system_config_brand_id ON public.system_config USING btree (brand_id);

CREATE INDEX idx_time_restrictions_role_game ON public.time_restrictions USING btree (role_id, game_id) WHERE ((restriction_type)::text = 'role'::text);

CREATE INDEX idx_time_restrictions_user_game ON public.time_restrictions USING btree (user_id, game_id) WHERE ((restriction_type)::text = 'user'::text);

CREATE INDEX idx_tip_transactions_brand_created ON public.tip_transactions USING btree (brand_id, created_at);

CREATE INDEX idx_tip_transactions_brand_id ON public.tip_transactions USING btree (brand_id);

CREATE INDEX idx_tip_transactions_brand_sender ON public.tip_transactions USING btree (brand_id, sender_id);

CREATE INDEX idx_tip_transactions_created_at ON public.tip_transactions USING btree (created_at DESC);

CREATE INDEX idx_tip_transactions_receiver_id ON public.tip_transactions USING btree (receiver_id);

CREATE INDEX idx_tip_transactions_sender_id ON public.tip_transactions USING btree (sender_id);

CREATE INDEX idx_tip_transactions_sender_receiver ON public.tip_transactions USING btree (sender_id, receiver_id);

CREATE INDEX idx_tip_transactions_status ON public.tip_transactions USING btree (status);

CREATE INDEX idx_tip_transactions_transaction_id ON public.tip_transactions USING btree (transaction_id);

CREATE INDEX idx_tournaments_rank ON public.tournaments USING btree (rank);

CREATE INDEX idx_transactions_currency_code ON public.transactions USING btree (currency_code);

CREATE INDEX idx_transactions_status ON public.transactions USING btree (status);

CREATE INDEX idx_transactions_tx_hash ON public.transactions USING btree (tx_hash);

CREATE INDEX idx_user_2fa_attempts_created_at ON public.user_2fa_attempts USING btree (created_at);

CREATE INDEX idx_user_2fa_attempts_type ON public.user_2fa_attempts USING btree (attempt_type);

CREATE INDEX idx_user_2fa_attempts_user_id ON public.user_2fa_attempts USING btree (user_id);

CREATE INDEX idx_user_2fa_methods_enabled_at ON public.user_2fa_methods USING btree (enabled_at);

CREATE INDEX idx_user_2fa_methods_method ON public.user_2fa_methods USING btree (method);

CREATE INDEX idx_user_2fa_methods_user_id ON public.user_2fa_methods USING btree (user_id);

CREATE INDEX idx_user_2fa_otps_expires_at ON public.user_2fa_otps USING btree (expires_at);

CREATE INDEX idx_user_2fa_otps_user_method ON public.user_2fa_otps USING btree (user_id, method);

CREATE INDEX idx_user_2fa_settings_enabled ON public.user_2fa_settings USING btree (is_enabled);

CREATE INDEX idx_user_2fa_settings_user_id ON public.user_2fa_settings USING btree (user_id);

CREATE INDEX idx_user_activity_log_activity_type ON public.user_activity_log USING btree (activity_type);

CREATE INDEX idx_user_activity_log_created_at ON public.user_activity_log USING btree (created_at);

CREATE INDEX idx_user_activity_log_user_id ON public.user_activity_log USING btree (user_id);

CREATE INDEX idx_user_allowed_pages_page_id ON public.user_allowed_pages USING btree (page_id);

CREATE INDEX idx_user_allowed_pages_user_id ON public.user_allowed_pages USING btree (user_id);

CREATE INDEX idx_user_game_access_access_type ON public.user_game_access USING btree (access_type);

CREATE INDEX idx_user_game_access_game_id ON public.user_game_access USING btree (game_id);

CREATE INDEX idx_user_game_access_user_id ON public.user_game_access USING btree (user_id);

CREATE INDEX idx_user_levels_brand_id ON public.user_levels USING btree (brand_id);

CREATE INDEX idx_user_levels_brand_user ON public.user_levels USING btree (brand_id, user_id);

CREATE INDEX idx_user_levels_current_level ON public.user_levels USING btree (current_level);

CREATE INDEX idx_user_levels_user_id ON public.user_levels USING btree (user_id);

CREATE INDEX idx_user_preferences_user_id ON public.user_preferences USING btree (user_id);

CREATE INDEX idx_user_sessions_brand_id ON public.user_sessions USING btree (brand_id);

CREATE INDEX idx_user_sessions_brand_user ON public.user_sessions USING btree (brand_id, user_id);

CREATE INDEX idx_users_brand_id ON public.users USING btree (brand_id);

CREATE INDEX idx_users_brand_status ON public.users USING btree (brand_id, status);

CREATE INDEX idx_users_kyc_status ON public.users USING btree (kyc_status);

CREATE INDEX idx_users_two_factor_enabled ON public.users USING btree (two_factor_enabled);

CREATE UNIQUE INDEX idx_users_username_lower_unique ON public.users USING btree (lower((username)::text)) WHERE (username IS NOT NULL);

COMMENT ON INDEX public.idx_users_username_lower_unique IS 'Case-insensitive unique index on username. Ensures usernames like "darren", "Darren", and "DARREN" are treated as duplicates.';

CREATE INDEX idx_violations_type ON public.restriction_violations USING btree (violation_type);

CREATE INDEX idx_violations_user_game ON public.restriction_violations USING btree (user_id, game_id);

CREATE UNIQUE INDEX idx_waiting_squad_members_user_squad ON public.waiting_squad_members USING btree (user_id, squad_id);

CREATE INDEX idx_wallet_balances_currency ON public.wallet_balances USING btree (currency_code);

CREATE INDEX idx_wallet_balances_wallet_id ON public.wallet_balances USING btree (wallet_id);

CREATE INDEX idx_wallet_transactions_currency ON public.wallet_transactions USING btree (currency_code);

CREATE INDEX idx_wallet_transactions_pending ON public.wallet_transactions USING btree (status, created_at) WHERE (status = 'pending'::public.wallet_transaction_status);

CREATE INDEX idx_wallet_transactions_status ON public.wallet_transactions USING btree (status);

CREATE INDEX idx_wallet_transactions_tx_hash ON public.wallet_transactions USING btree (tx_hash);

CREATE INDEX idx_wallet_transactions_wallet_id ON public.wallet_transactions USING btree (wallet_id);

CREATE INDEX idx_wallets_address ON public.wallets USING btree (address);

CREATE INDEX idx_wallets_chain_id ON public.wallets USING btree (chain_id);

CREATE INDEX idx_wallets_chain_id_not_in_webhook ON public.wallets USING btree (chain_id, is_in_alchemy_webhook) WHERE ((is_active = true) AND (is_in_alchemy_webhook = false));

CREATE INDEX idx_wallets_user_id ON public.wallets USING btree (user_id);

CREATE INDEX idx_withdrawals_requires_admin_review ON public.withdrawals USING btree (requires_admin_review);

CREATE INDEX idx_withdrawals_status ON public.withdrawals USING btree (status);

CREATE INDEX idx_withdrawals_status_tx_hash ON public.withdrawals USING btree (status, tx_hash) WHERE ((status = ANY (ARRAY['pending'::public.withdrawal_status, 'processing'::public.withdrawal_status])) AND (tx_hash IS NOT NULL));

CREATE INDEX idx_withdrawals_user_id ON public.withdrawals USING btree (user_id);

CREATE INDEX idx_withdrawals_withdrawal_id ON public.withdrawals USING btree (withdrawal_id);

CREATE UNIQUE INDEX uniq_adds_services_service_id_active ON public.adds_services USING btree (service_id) WHERE (deleted_at IS NULL);

CREATE UNIQUE INDEX uniq_adds_services_service_id_deleted ON public.adds_services USING btree (service_id, deleted_at);

CREATE UNIQUE INDEX uniq_company_support_phone_active ON public.company USING btree (support_phone) WHERE (deleted_at IS NULL);

CREATE UNIQUE INDEX uniq_company_support_phone_deleted ON public.company USING btree (support_phone, deleted_at);

CREATE UNIQUE INDEX uniq_levels_deleted ON public.levels USING btree (level, deleted_at);

CREATE UNIQUE INDEX uniq_lottery_client_id_active ON public.lottery_services USING btree (client_id) WHERE (deleted_at IS NULL);

CREATE UNIQUE INDEX uniq_lottery_client_id_deleted ON public.lottery_services USING btree (client_id, deleted_at);

CREATE UNIQUE INDEX uniq_squads_active ON public.squads USING btree (handle) WHERE (deleted_at IS NULL);

CREATE UNIQUE INDEX uniq_squads_deleted ON public.squads USING btree (handle, deleted_at);

CREATE TRIGGER trg_delete_block_when_active AFTER UPDATE ON public.users FOR EACH ROW EXECUTE FUNCTION public.delete_account_block_on_active();

CREATE TRIGGER trigger_kyc_status_change AFTER UPDATE ON public.users FOR EACH ROW EXECUTE FUNCTION public.log_kyc_status_change();

CREATE TRIGGER trigger_update_admin_activity_logs_updated_at BEFORE UPDATE ON public.admin_activity_logs FOR EACH ROW EXECUTE FUNCTION public.update_admin_activity_logs_updated_at();

CREATE TRIGGER trigger_update_alert_email_groups_updated_at BEFORE UPDATE ON public.alert_email_groups FOR EACH ROW EXECUTE FUNCTION public.update_alert_email_groups_updated_at();

CREATE TRIGGER trigger_update_crypto_wallet_connections_updated_at BEFORE UPDATE ON public.crypto_wallet_connections FOR EACH ROW EXECUTE FUNCTION public.update_crypto_wallet_connections_updated_at();

CREATE TRIGGER trigger_update_game_session_activity BEFORE UPDATE ON public.game_sessions FOR EACH ROW EXECUTE FUNCTION public.update_game_session_activity();

CREATE TRIGGER trigger_update_groove_accounts_updated_at BEFORE UPDATE ON public.groove_accounts FOR EACH ROW EXECUTE FUNCTION public.update_groove_accounts_updated_at();

CREATE TRIGGER update_message_campaigns_updated_at BEFORE UPDATE ON public.message_campaigns FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();

CREATE TRIGGER update_otps_updated_at BEFORE UPDATE ON public.otps FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();

CREATE TRIGGER update_tip_transactions_updated_at BEFORE UPDATE ON public.tip_transactions FOR EACH ROW EXECUTE FUNCTION public.update_tip_transactions_updated_at();

ALTER TABLE ONLY public.account_block
    ADD CONSTRAINT account_block_blocked_by_fkey FOREIGN KEY (blocked_by) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.account_block
    ADD CONSTRAINT account_block_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.active_game_sessions
    ADD CONSTRAINT active_game_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.adds_services
    ADD CONSTRAINT adds_services_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.admin_activity_actions
    ADD CONSTRAINT admin_activity_actions_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.admin_activity_categories(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.admin_activity_logs
    ADD CONSTRAINT admin_activity_logs_admin_user_id_fkey FOREIGN KEY (admin_user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.admin_fund_movements
    ADD CONSTRAINT admin_fund_movements_admin_id_fkey FOREIGN KEY (admin_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.admin_fund_movements
    ADD CONSTRAINT admin_fund_movements_chain_id_fkey FOREIGN KEY (chain_id) REFERENCES public.supported_chains(chain_id);

ALTER TABLE ONLY public.admin_fund_movements
    ADD CONSTRAINT admin_fund_movements_currency_code_fkey FOREIGN KEY (currency_code) REFERENCES public.currency_config(currency_code);

ALTER TABLE ONLY public.affiliate_referal_track
    ADD CONSTRAINT affiliate_referal_track_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.airtime_transactions
    ADD CONSTRAINT airtime_transactions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.alchemy_webhooks
    ADD CONSTRAINT alchemy_webhooks_chain_id_fkey FOREIGN KEY (chain_id) REFERENCES public.supported_chains(chain_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.alert_configurations
    ADD CONSTRAINT alert_configurations_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.alert_configurations
    ADD CONSTRAINT alert_configurations_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.alert_email_group_members
    ADD CONSTRAINT alert_email_group_members_group_id_fkey FOREIGN KEY (group_id) REFERENCES public.alert_email_groups(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.alert_email_groups
    ADD CONSTRAINT alert_email_groups_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.alert_email_groups
    ADD CONSTRAINT alert_email_groups_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.alert_triggers
    ADD CONSTRAINT alert_triggers_acknowledged_by_fkey FOREIGN KEY (acknowledged_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.alert_triggers
    ADD CONSTRAINT alert_triggers_alert_configuration_id_fkey FOREIGN KEY (alert_configuration_id) REFERENCES public.alert_configurations(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.alert_triggers
    ADD CONSTRAINT alert_triggers_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_admin_id_fkey FOREIGN KEY (admin_id) REFERENCES public.users(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.balance_logs
    ADD CONSTRAINT balance_logs_brand_id_fkey FOREIGN KEY (brand_id) REFERENCES public.brands(id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.balance_logs
    ADD CONSTRAINT balance_logs_operational_group_id_fkey FOREIGN KEY (operational_group_id) REFERENCES public.operational_groups(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.balance_logs
    ADD CONSTRAINT balance_logs_operational_type_id_fkey FOREIGN KEY (operational_type_id) REFERENCES public.operational_types(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.balance_logs
    ADD CONSTRAINT balance_logs_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.balances
    ADD CONSTRAINT balances_brand_id_fkey FOREIGN KEY (brand_id) REFERENCES public.brands(id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.balances
    ADD CONSTRAINT balances_currency_code_fkey FOREIGN KEY (currency_code) REFERENCES public.currency_config(currency_code);

ALTER TABLE ONLY public.balances
    ADD CONSTRAINT balances_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.bets
    ADD CONSTRAINT bets_brand_id_fkey FOREIGN KEY (brand_id) REFERENCES public.brands(id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.bets
    ADD CONSTRAINT bets_round_id_fkey FOREIGN KEY (round_id) REFERENCES public.rounds(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.bets
    ADD CONSTRAINT bets_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.campaign_recipients
    ADD CONSTRAINT campaign_recipients_campaign_id_fkey FOREIGN KEY (campaign_id) REFERENCES public.message_campaigns(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.campaign_recipients
    ADD CONSTRAINT campaign_recipients_notification_id_fkey FOREIGN KEY (notification_id) REFERENCES public.user_notifications(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.campaign_recipients
    ADD CONSTRAINT campaign_recipients_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.cashback_claims
    ADD CONSTRAINT cashback_claims_transaction_id_fkey FOREIGN KEY (transaction_id) REFERENCES public.balance_logs(id);

ALTER TABLE ONLY public.cashback_claims
    ADD CONSTRAINT cashback_claims_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.cashback_earnings
    ADD CONSTRAINT cashback_earnings_brand_id_fkey FOREIGN KEY (brand_id) REFERENCES public.brands(id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.cashback_earnings
    ADD CONSTRAINT cashback_earnings_source_bet_id_fkey FOREIGN KEY (source_bet_id) REFERENCES public.bets(id);

ALTER TABLE ONLY public.cashback_earnings
    ADD CONSTRAINT cashback_earnings_tier_id_fkey FOREIGN KEY (tier_id) REFERENCES public.cashback_tiers(id);

ALTER TABLE ONLY public.cashback_earnings
    ADD CONSTRAINT cashback_earnings_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.chain_currencies
    ADD CONSTRAINT chain_currencies_chain_id_fkey FOREIGN KEY (chain_id) REFERENCES public.supported_chains(chain_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.chain_currencies
    ADD CONSTRAINT chain_currencies_currency_code_fkey FOREIGN KEY (currency_code) REFERENCES public.currency_config(currency_code) ON DELETE CASCADE;

ALTER TABLE ONLY public.chain_processing_state
    ADD CONSTRAINT chain_processing_state_chain_id_fkey FOREIGN KEY (chain_id) REFERENCES public.supported_chains(chain_id);

ALTER TABLE ONLY public.company
    ADD CONSTRAINT company_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.conversion_remainders
    ADD CONSTRAINT conversion_remainders_currency_code_fkey FOREIGN KEY (currency_code) REFERENCES public.currency_config(currency_code);

ALTER TABLE ONLY public.crypto_kings
    ADD CONSTRAINT crypto_kings_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.crypto_wallet_connections
    ADD CONSTRAINT crypto_wallet_connections_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.departements_users
    ADD CONSTRAINT departements_users_department_id_fkey FOREIGN KEY (department_id) REFERENCES public.departments(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.departements_users
    ADD CONSTRAINT departements_users_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.deposit_sessions
    ADD CONSTRAINT deposit_sessions_chain_id_fkey FOREIGN KEY (chain_id) REFERENCES public.supported_chains(chain_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.deposit_sessions
    ADD CONSTRAINT deposit_sessions_currency_code_fkey FOREIGN KEY (currency_code) REFERENCES public.currency_config(currency_code);

ALTER TABLE ONLY public.deposit_sessions
    ADD CONSTRAINT deposit_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.deposit_sessions
    ADD CONSTRAINT deposit_sessions_wallet_id_fkey FOREIGN KEY (wallet_id) REFERENCES public.wallets(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.emergency_access_overrides
    ADD CONSTRAINT emergency_access_overrides_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.emergency_access_overrides
    ADD CONSTRAINT emergency_access_overrides_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);

ALTER TABLE ONLY public.exchange_rates
    ADD CONSTRAINT exchange_rates_from_currency_fkey FOREIGN KEY (from_currency) REFERENCES public.currency_config(currency_code);

ALTER TABLE ONLY public.exchange_rates
    ADD CONSTRAINT exchange_rates_to_currency_fkey FOREIGN KEY (to_currency) REFERENCES public.currency_config(currency_code);

ALTER TABLE ONLY public.failed_bet_logs
    ADD CONSTRAINT failed_bet_logs_bet_id_fkey FOREIGN KEY (bet_id) REFERENCES public.bets(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.failed_bet_logs
    ADD CONSTRAINT failed_bet_logs_round_id_fkey FOREIGN KEY (round_id) REFERENCES public.rounds(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.failed_bet_logs
    ADD CONSTRAINT failed_bet_logs_transaction_id_fkey FOREIGN KEY (transaction_id) REFERENCES public.balance_logs(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.failed_bet_logs
    ADD CONSTRAINT failed_bet_logs_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.falcon_liquidity_messages
    ADD CONSTRAINT falcon_liquidity_messages_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);

ALTER TABLE ONLY public.agent_referrals
    ADD CONSTRAINT fk_agent_referrals_user_id FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.balances
    ADD CONSTRAINT fk_balances_currency_code FOREIGN KEY (currency_code) REFERENCES public.currency_config(currency_code);

ALTER TABLE ONLY public.levels
    ADD CONSTRAINT fk_created_by FOREIGN KEY (created_by) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.level_requirements
    ADD CONSTRAINT fk_created_by FOREIGN KEY (created_by) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.level_requirements
    ADD CONSTRAINT fk_level_id FOREIGN KEY (level_id) REFERENCES public.levels(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.pages
    ADD CONSTRAINT fk_pages_parent FOREIGN KEY (parent_id) REFERENCES public.pages(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.squads_earns
    ADD CONSTRAINT fk_squad_earns_squad FOREIGN KEY (squad_id) REFERENCES public.squads(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.squads_earns
    ADD CONSTRAINT fk_squad_earns_user FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.squads
    ADD CONSTRAINT fk_squad_owner FOREIGN KEY (owner) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.squads_memebers
    ADD CONSTRAINT fk_squad_user_by FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.logs
    ADD CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES public.users(id);

ALTER TABLE ONLY public.affiliate_referal_track
    ADD CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.user_allowed_pages
    ADD CONSTRAINT fk_user_allowed_pages_page FOREIGN KEY (page_id) REFERENCES public.pages(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.user_allowed_pages
    ADD CONSTRAINT fk_user_allowed_pages_user FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.football_matchs
    ADD CONSTRAINT football_matchs_round_id_fkey FOREIGN KEY (round_id) REFERENCES public.football_match_rounds(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.game_access_logs
    ADD CONSTRAINT game_access_logs_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);

ALTER TABLE ONLY public.game_access_templates
    ADD CONSTRAINT game_access_templates_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.game_logs
    ADD CONSTRAINT game_logs_round_id_fkey FOREIGN KEY (round_id) REFERENCES public.rounds(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.game_sessions
    ADD CONSTRAINT game_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.geographic_restrictions
    ADD CONSTRAINT geographic_restrictions_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.roles(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.geographic_restrictions
    ADD CONSTRAINT geographic_restrictions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.global_rakeback_override
    ADD CONSTRAINT global_rakeback_override_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.global_rakeback_override
    ADD CONSTRAINT global_rakeback_override_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.groove_accounts
    ADD CONSTRAINT groove_accounts_brand_id_fkey FOREIGN KEY (brand_id) REFERENCES public.brands(id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.groove_accounts
    ADD CONSTRAINT groove_accounts_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.groove_game_sessions
    ADD CONSTRAINT groove_game_sessions_account_id_fkey FOREIGN KEY (account_id) REFERENCES public.groove_accounts(account_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.groove_game_sessions
    ADD CONSTRAINT groove_game_sessions_brand_id_fkey FOREIGN KEY (brand_id) REFERENCES public.brands(id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.groove_game_sessions
    ADD CONSTRAINT groove_game_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);

ALTER TABLE ONLY public.groove_transactions
    ADD CONSTRAINT groove_transactions_account_id_fkey FOREIGN KEY (account_id) REFERENCES public.groove_accounts(account_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.groove_transactions
    ADD CONSTRAINT groove_transactions_brand_id_fkey FOREIGN KEY (brand_id) REFERENCES public.brands(id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.groove_transactions
    ADD CONSTRAINT groove_transactions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);

ALTER TABLE ONLY public.ip_filters
    ADD CONSTRAINT ip_filters_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.kyc_documents
    ADD CONSTRAINT kyc_documents_reviewed_by_fkey FOREIGN KEY (reviewed_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.kyc_documents
    ADD CONSTRAINT kyc_documents_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.kyc_status_changes
    ADD CONSTRAINT kyc_status_changes_changed_by_fkey FOREIGN KEY (changed_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.kyc_status_changes
    ADD CONSTRAINT kyc_status_changes_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.kyc_submissions
    ADD CONSTRAINT kyc_submissions_reviewed_by_fkey FOREIGN KEY (reviewed_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.kyc_submissions
    ADD CONSTRAINT kyc_submissions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.login_attempts
    ADD CONSTRAINT login_attempts_brand_id_fkey FOREIGN KEY (brand_id) REFERENCES public.brands(id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.login_attempts
    ADD CONSTRAINT login_attempts_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.lottery_winners_logs
    ADD CONSTRAINT lottery_winners_logs_lottery_id_fkey FOREIGN KEY (lottery_id) REFERENCES public.lotteries(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.lottery_winners_logs
    ADD CONSTRAINT lottery_winners_logs_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.manual_funds
    ADD CONSTRAINT manual_funds_admin_id_fkey FOREIGN KEY (admin_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.manual_funds
    ADD CONSTRAINT manual_funds_currency_code_fkey FOREIGN KEY (currency_code) REFERENCES public.currency_config(currency_code);

ALTER TABLE ONLY public.manual_funds
    ADD CONSTRAINT manual_funds_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.message_campaigns
    ADD CONSTRAINT message_campaigns_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.message_segments
    ADD CONSTRAINT message_segments_campaign_id_fkey FOREIGN KEY (campaign_id) REFERENCES public.message_campaigns(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.operational_types
    ADD CONSTRAINT operational_types_group_id_fkey FOREIGN KEY (group_id) REFERENCES public.operational_groups(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.passkey_credentials
    ADD CONSTRAINT passkey_credentials_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.player_deposit_tracking
    ADD CONSTRAINT player_deposit_tracking_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.player_excluded_games
    ADD CONSTRAINT player_excluded_games_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.plinko
    ADD CONSTRAINT plinko_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.quick_hustles
    ADD CONSTRAINT quick_hustles_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.rakeback_schedules
    ADD CONSTRAINT rakeback_schedules_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.restriction_violations
    ADD CONSTRAINT restriction_violations_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.role_game_access
    ADD CONSTRAINT role_game_access_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.roles(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_permission_id_fkey FOREIGN KEY (permission_id) REFERENCES public.permissions(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.roles(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.roll_da_dice
    ADD CONSTRAINT roll_da_dice_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.scratch_cards
    ADD CONSTRAINT scratch_cards_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.session_limits
    ADD CONSTRAINT session_limits_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.roles(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.session_limits
    ADD CONSTRAINT session_limits_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.spending_limits
    ADD CONSTRAINT spending_limits_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.roles(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.spending_limits
    ADD CONSTRAINT spending_limits_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.spending_tracking
    ADD CONSTRAINT spending_tracking_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.spinning_wheel_configs
    ADD CONSTRAINT spinning_wheel_configs_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.spinning_wheel_mysteries
    ADD CONSTRAINT spinning_wheel_mysteries_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.spinning_wheel_rewards
    ADD CONSTRAINT spinning_wheel_rewards_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);

ALTER TABLE ONLY public.spinning_wheels
    ADD CONSTRAINT spinning_wheels_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.sport_bets
    ADD CONSTRAINT sport_bets_brand_id_fkey FOREIGN KEY (brand_id) REFERENCES public.brands(id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.sport_bets
    ADD CONSTRAINT sport_bets_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.street_kings
    ADD CONSTRAINT street_kings_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.supported_chains
    ADD CONSTRAINT supported_chains_native_currency_fkey FOREIGN KEY (native_currency) REFERENCES public.currency_config(currency_code);

ALTER TABLE ONLY public.system_config
    ADD CONSTRAINT system_config_brand_id_fkey FOREIGN KEY (brand_id) REFERENCES public.brands(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.system_config
    ADD CONSTRAINT system_config_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES public.users(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.time_restrictions
    ADD CONSTRAINT time_restrictions_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.roles(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.time_restrictions
    ADD CONSTRAINT time_restrictions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.tip_transactions
    ADD CONSTRAINT tip_transactions_brand_id_fkey FOREIGN KEY (brand_id) REFERENCES public.brands(id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.tip_transactions
    ADD CONSTRAINT tip_transactions_receiver_id_fkey FOREIGN KEY (receiver_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.tip_transactions
    ADD CONSTRAINT tip_transactions_receiver_transaction_log_id_fkey FOREIGN KEY (receiver_transaction_log_id) REFERENCES public.balance_logs(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.tip_transactions
    ADD CONSTRAINT tip_transactions_sender_id_fkey FOREIGN KEY (sender_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.tip_transactions
    ADD CONSTRAINT tip_transactions_sender_transaction_log_id_fkey FOREIGN KEY (sender_transaction_log_id) REFERENCES public.balance_logs(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.tournaments_claims
    ADD CONSTRAINT tournaments_claims_squad_id_fkey FOREIGN KEY (squad_id) REFERENCES public.squads(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.tournaments_claims
    ADD CONSTRAINT tournaments_claims_tournament_id_fkey FOREIGN KEY (tournament_id) REFERENCES public.tournaments(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_chain_id_fkey FOREIGN KEY (chain_id) REFERENCES public.supported_chains(chain_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_currency_code_fkey FOREIGN KEY (currency_code) REFERENCES public.currency_config(currency_code);

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_withdrawal_id_fkey FOREIGN KEY (withdrawal_id) REFERENCES public.withdrawals(withdrawal_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.user_2fa_attempts
    ADD CONSTRAINT user_2fa_attempts_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.user_2fa_methods
    ADD CONSTRAINT user_2fa_methods_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.user_2fa_settings
    ADD CONSTRAINT user_2fa_settings_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.user_activity_log
    ADD CONSTRAINT user_activity_log_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.user_game_access
    ADD CONSTRAINT user_game_access_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.user_game_access
    ADD CONSTRAINT user_game_access_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.user_levels
    ADD CONSTRAINT user_levels_brand_id_fkey FOREIGN KEY (brand_id) REFERENCES public.brands(id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.user_levels
    ADD CONSTRAINT user_levels_current_tier_id_fkey FOREIGN KEY (current_tier_id) REFERENCES public.cashback_tiers(id);

ALTER TABLE ONLY public.user_levels
    ADD CONSTRAINT user_levels_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.user_limits
    ADD CONSTRAINT user_limits_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.user_notifications
    ADD CONSTRAINT user_notifications_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.user_preferences
    ADD CONSTRAINT user_preferences_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.roles(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.user_sessions
    ADD CONSTRAINT user_sessions_brand_id_fkey FOREIGN KEY (brand_id) REFERENCES public.brands(id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.user_sessions
    ADD CONSTRAINT user_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);

ALTER TABLE ONLY public.user_withdrawal_usage
    ADD CONSTRAINT user_withdrawal_usage_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_brand_id_fkey FOREIGN KEY (brand_id) REFERENCES public.brands(id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.users_football_matche_rounds
    ADD CONSTRAINT users_football_matche_rounds_football_round_id_fkey FOREIGN KEY (football_round_id) REFERENCES public.football_match_rounds(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.users_football_matche_rounds
    ADD CONSTRAINT users_football_matche_rounds_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.users_football_matches
    ADD CONSTRAINT users_football_matches_match_id_fkey FOREIGN KEY (match_id) REFERENCES public.football_matchs(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.waiting_squad_members
    ADD CONSTRAINT waiting_squad_members_squad_id_fkey FOREIGN KEY (squad_id) REFERENCES public.squads(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.waiting_squad_members
    ADD CONSTRAINT waiting_squad_members_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.wallet_balances
    ADD CONSTRAINT wallet_balances_currency_code_fkey FOREIGN KEY (currency_code) REFERENCES public.currency_config(currency_code);

ALTER TABLE ONLY public.wallet_balances
    ADD CONSTRAINT wallet_balances_wallet_id_fkey FOREIGN KEY (wallet_id) REFERENCES public.wallets(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.wallet_transactions
    ADD CONSTRAINT wallet_transactions_currency_code_fkey FOREIGN KEY (currency_code) REFERENCES public.currency_config(currency_code);

ALTER TABLE ONLY public.wallet_transactions
    ADD CONSTRAINT wallet_transactions_wallet_id_fkey FOREIGN KEY (wallet_id) REFERENCES public.wallets(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.wallets
    ADD CONSTRAINT wallets_chain_id_fkey FOREIGN KEY (chain_id) REFERENCES public.supported_chains(chain_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.wallets
    ADD CONSTRAINT wallets_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.withdrawals
    ADD CONSTRAINT withdrawals_admin_id_fkey FOREIGN KEY (admin_id) REFERENCES public.users(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.withdrawals
    ADD CONSTRAINT withdrawals_chain_id_fkey FOREIGN KEY (chain_id) REFERENCES public.supported_chains(chain_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.withdrawals
    ADD CONSTRAINT withdrawals_currency_code_fkey FOREIGN KEY (currency_code) REFERENCES public.currency_config(currency_code);

ALTER TABLE ONLY public.withdrawals
    ADD CONSTRAINT withdrawals_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

-- PostgreSQL database dump complete
