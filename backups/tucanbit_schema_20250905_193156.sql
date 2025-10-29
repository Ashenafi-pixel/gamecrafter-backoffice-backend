--
-- PostgreSQL database dump
--

-- Dumped from database version 13.21 (Debian 13.21-1.pgdg120+1)
-- Dumped by pg_dump version 13.21 (Debian 13.21-1.pgdg120+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: bet_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.bet_status AS ENUM (
    'open',
    'in_progress',
    'closed',
    'failed'
);


--
-- Name: components; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.components AS ENUM (
    'real_money',
    'bonus_money',
    'points'
);


--
-- Name: spinningwheelmysterytypes; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.spinningwheelmysterytypes AS ENUM (
    'point',
    'internet_package_in_gb',
    'better',
    'other'
);


--
-- Name: spinningwheeltypes; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.spinningwheeltypes AS ENUM (
    'point',
    'internet_package_in_gb',
    'better',
    'mystery',
    'free spin'
);


--
-- Name: clean_expired_crypto_wallet_challenges(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.clean_expired_crypto_wallet_challenges() RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
    DELETE FROM crypto_wallet_challenges WHERE expires_at < NOW();
END;
$$;


--
-- Name: get_user_wallets(uuid); Type: FUNCTION; Schema: public; Owner: -
--

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


--
-- Name: update_crypto_wallet_connections_updated_at(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_crypto_wallet_connections_updated_at() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;


--
-- Name: update_updated_at_column(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_updated_at_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: account_block; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: adds_services; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: agent_providers; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: agent_referrals; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: airtime_transactions; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: airtime_utilities; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: balance_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.balance_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    component public.components NOT NULL,
    currency character varying(3),
    change_amount numeric,
    operational_group_id uuid,
    operational_type_id uuid,
    description text,
    "timestamp" timestamp without time zone,
    balance_after_update numeric DEFAULT 0.0 NOT NULL,
    transaction_id character varying DEFAULT ''::character varying NOT NULL,
    status character varying DEFAULT 'COMPLETED'::character varying
);


--
-- Name: balances; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.balances (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    currency character varying(3) NOT NULL,
    real_money numeric,
    bonus_money numeric,
    points integer,
    updated_at timestamp without time zone
);


--
-- Name: banners; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: bets; Type: TABLE; Schema: public; Owner: -
--

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
    status character varying DEFAULT 'ACTIVE'::character varying
);


--
-- Name: casbin_rule; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: casbin_rule_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.casbin_rule_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: casbin_rule_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.casbin_rule_id_seq OWNED BY public.casbin_rule.id;


--
-- Name: clubs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.clubs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    club_name character varying NOT NULL,
    status character varying DEFAULT 'ACTIVE'::character varying,
    "timestamp" timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: company; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: configs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.configs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(20) NOT NULL,
    value character varying NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: crypto_kings; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: crypto_wallet_auth_logs; Type: TABLE; Schema: public; Owner: -
--

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
    CONSTRAINT crypto_wallet_auth_logs_action_check CHECK (((action)::text = ANY ((ARRAY['connect'::character varying, 'disconnect'::character varying, 'login'::character varying, 'verify'::character varying, 'challenge'::character varying])::text[])))
);


--
-- Name: crypto_wallet_challenges; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: crypto_wallet_connections; Type: TABLE; Schema: public; Owner: -
--

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
    CONSTRAINT crypto_wallet_connections_wallet_type_check CHECK (((wallet_type)::text = ANY ((ARRAY['metamask'::character varying, 'walletconnect'::character varying, 'coinbase'::character varying, 'phantom'::character varying, 'trust'::character varying, 'ledger'::character varying])::text[])))
);


--
-- Name: currencies; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.currencies (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying NOT NULL,
    status character varying DEFAULT 'ACTIVE'::character varying NOT NULL,
    "timestamp" timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: departements_users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.departements_users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    department_id uuid NOT NULL,
    created_at timestamp with time zone
);


--
-- Name: departments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.departments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying NOT NULL,
    notifications text[],
    created_at timestamp without time zone
);


--
-- Name: exchange_rates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.exchange_rates (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    currency_from character varying(3),
    currency_to character varying(3),
    rate numeric,
    updated_at timestamp without time zone
);


--
-- Name: failed_bet_logs; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: football_match_rounds; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.football_match_rounds (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    status character varying,
    "timestamp" timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: football_matchs; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: game_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.game_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    round_id uuid NOT NULL,
    action character varying(255) NOT NULL,
    detail json,
    "timestamp" timestamp without time zone
);


--
-- Name: games; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.games (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying NOT NULL,
    status character varying DEFAULT 'ACTIVE'::character varying NOT NULL,
    "timestamp" timestamp without time zone DEFAULT now() NOT NULL,
    photo character varying,
    price character varying,
    enabled boolean DEFAULT false
);


--
-- Name: ip_filters; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: leagues; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.leagues (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    league_name character varying NOT NULL,
    status character varying DEFAULT 'ACTIVE'::character varying,
    "timestamp" timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: level_requirements; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: levels; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.levels (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    level numeric NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at timestamp without time zone,
    created_by uuid NOT NULL,
    type character varying(50) DEFAULT 'players'::character varying NOT NULL,
    CONSTRAINT levels_type_check CHECK (((type)::text = ANY ((ARRAY['players'::character varying, 'squads'::character varying])::text[])))
);


--
-- Name: login_attempts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.login_attempts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    ip_address character varying(50) NOT NULL,
    success boolean NOT NULL,
    attempt_time timestamp without time zone,
    user_agent character varying(50)
);


--
-- Name: logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    module character varying(255) NOT NULL,
    detail json,
    ip_address character varying(46),
    "timestamp" timestamp without time zone
);


--
-- Name: loot_box; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.loot_box (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    type text NOT NULL,
    prizeamount numeric NOT NULL,
    weight numeric NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: loot_box_place_bets; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: lotteries; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: lottery_logs; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: lottery_services; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: lottery_winners_logs; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: manual_funds; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.manual_funds (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    admin_id uuid NOT NULL,
    transaction_id character varying NOT NULL,
    type character varying NOT NULL,
    amount numeric NOT NULL,
    currency character varying(3) NOT NULL,
    note character varying NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    reason character varying DEFAULT 'system_restart'::character varying NOT NULL
);


--
-- Name: operational_groups; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.operational_groups (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(50),
    description text,
    created_at timestamp without time zone
);


--
-- Name: operational_types; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.operational_types (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    group_id uuid NOT NULL,
    name character varying(50),
    description text,
    created_at timestamp without time zone
);


--
-- Name: otps; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: TABLE otps; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON TABLE public.otps IS 'OTP table for email verification and password reset functionality';


--
-- Name: COLUMN otps.email; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.otps.email IS 'Email address for OTP delivery';


--
-- Name: COLUMN otps.otp_code; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.otps.otp_code IS '6-digit OTP code';


--
-- Name: COLUMN otps.type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.otps.type IS 'Type of OTP: email_verification, password_reset, login';


--
-- Name: COLUMN otps.status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.otps.status IS 'Status: pending, verified, expired, used';


--
-- Name: COLUMN otps.expires_at; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.otps.expires_at IS 'When the OTP expires';


--
-- Name: COLUMN otps.verified_at; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.otps.verified_at IS 'When the OTP was verified';


--
-- Name: permissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.permissions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying NOT NULL,
    description character varying
);


--
-- Name: plinko; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: quick_hustles; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: risk_settings; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: role_permissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.role_permissions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    role_id uuid NOT NULL,
    permission_id uuid NOT NULL
);


--
-- Name: roles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.roles (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying NOT NULL,
    description character varying
);


--
-- Name: roll_da_dice; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: rounds; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.rounds (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    status public.bet_status NOT NULL,
    crash_point numeric(10,2) NOT NULL,
    created_at timestamp without time zone NOT NULL,
    closed_at timestamp without time zone
);


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);


--
-- Name: scratch_cards; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.scratch_cards (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    status character varying DEFAULT 'ACTIVE'::character varying NOT NULL,
    bet_amount numeric NOT NULL,
    won_status character varying,
    "timestamp" timestamp without time zone NOT NULL,
    won_amount numeric
);


--
-- Name: spinning_wheel_configs; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: spinning_wheel_mysteries; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: spinning_wheel_rewards; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: spinning_wheels; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: sport_bets; Type: TABLE; Schema: public; Owner: -
--

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
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: squads; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.squads (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    handle character varying NOT NULL,
    owner uuid NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at timestamp without time zone,
    type character varying(50) DEFAULT 'Open'::character varying NOT NULL
);


--
-- Name: squads_earns; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: squads_memebers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.squads_memebers (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    squad_id uuid NOT NULL,
    user_id uuid NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at timestamp without time zone
);


--
-- Name: street_kings; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: temp; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.temp (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    data jsonb NOT NULL
);


--
-- Name: tournaments; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: tournaments_claims; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tournaments_claims (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tournament_id uuid NOT NULL,
    squad_id uuid NOT NULL,
    claimed_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: user_notifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_notifications (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    title text NOT NULL,
    content text NOT NULL,
    type text NOT NULL,
    metadata jsonb,
    read boolean DEFAULT false NOT NULL,
    delivered boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    read_at timestamp with time zone,
    created_by uuid
);


--
-- Name: user_roles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_roles (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    role_id uuid NOT NULL
);


--
-- Name: user_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_sessions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    token text NOT NULL,
    expires_at timestamp without time zone NOT NULL,
    ip_address character varying(46),
    user_agent character varying(255),
    created_at timestamp without time zone,
    refresh_token text,
    refresh_token_expires_at timestamp without time zone
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

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
    CONSTRAINT users_wallet_verification_status_check CHECK (((wallet_verification_status)::text = ANY ((ARRAY['none'::character varying, 'pending'::character varying, 'verified'::character varying, 'failed'::character varying])::text[])))
);


--
-- Name: users_football_matche_rounds; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: users_football_matches; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users_football_matches (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    status character varying DEFAULT 'PENDING'::character varying NOT NULL,
    match_id uuid NOT NULL,
    selection character varying DEFAULT 'DRAW'::character varying NOT NULL,
    users_football_matche_round_id uuid
);


--
-- Name: users_otp; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users_otp (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    otp character varying NOT NULL,
    created_at timestamp without time zone NOT NULL
);


--
-- Name: waiting_squad_members; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.waiting_squad_members (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    squad_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: casbin_rule id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.casbin_rule ALTER COLUMN id SET DEFAULT nextval('public.casbin_rule_id_seq'::regclass);


--
-- Name: account_block account_block_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.account_block
    ADD CONSTRAINT account_block_pkey PRIMARY KEY (id);


--
-- Name: adds_services adds_services_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.adds_services
    ADD CONSTRAINT adds_services_pkey PRIMARY KEY (id);


--
-- Name: agent_providers agent_providers_client_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_providers
    ADD CONSTRAINT agent_providers_client_id_key UNIQUE (client_id);


--
-- Name: agent_providers agent_providers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_providers
    ADD CONSTRAINT agent_providers_pkey PRIMARY KEY (id);


--
-- Name: agent_referrals agent_referrals_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_referrals
    ADD CONSTRAINT agent_referrals_pkey PRIMARY KEY (id);


--
-- Name: agent_referrals agent_referrals_request_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_referrals
    ADD CONSTRAINT agent_referrals_request_id_key UNIQUE (request_id);


--
-- Name: airtime_transactions airtime_transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.airtime_transactions
    ADD CONSTRAINT airtime_transactions_pkey PRIMARY KEY (id);


--
-- Name: airtime_utilities airtime_utilities_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.airtime_utilities
    ADD CONSTRAINT airtime_utilities_pkey PRIMARY KEY (local_id);


--
-- Name: balance_logs balance_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.balance_logs
    ADD CONSTRAINT balance_logs_pkey PRIMARY KEY (id);


--
-- Name: balances balances_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.balances
    ADD CONSTRAINT balances_pkey PRIMARY KEY (id);


--
-- Name: banners banners_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.banners
    ADD CONSTRAINT banners_pkey PRIMARY KEY (id);


--
-- Name: bets bets_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.bets
    ADD CONSTRAINT bets_pkey PRIMARY KEY (id);


--
-- Name: casbin_rule casbin_rule_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.casbin_rule
    ADD CONSTRAINT casbin_rule_pkey PRIMARY KEY (id);


--
-- Name: clubs clubs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.clubs
    ADD CONSTRAINT clubs_pkey PRIMARY KEY (id);


--
-- Name: company company_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.company
    ADD CONSTRAINT company_pkey PRIMARY KEY (id);


--
-- Name: configs configs_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.configs
    ADD CONSTRAINT configs_name_key UNIQUE (name);


--
-- Name: configs configs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.configs
    ADD CONSTRAINT configs_pkey PRIMARY KEY (id);


--
-- Name: crypto_kings crypto_kings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.crypto_kings
    ADD CONSTRAINT crypto_kings_pkey PRIMARY KEY (id);


--
-- Name: crypto_wallet_auth_logs crypto_wallet_auth_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.crypto_wallet_auth_logs
    ADD CONSTRAINT crypto_wallet_auth_logs_pkey PRIMARY KEY (id);


--
-- Name: crypto_wallet_challenges crypto_wallet_challenges_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.crypto_wallet_challenges
    ADD CONSTRAINT crypto_wallet_challenges_pkey PRIMARY KEY (id);


--
-- Name: crypto_wallet_connections crypto_wallet_connections_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.crypto_wallet_connections
    ADD CONSTRAINT crypto_wallet_connections_pkey PRIMARY KEY (id);


--
-- Name: crypto_wallet_connections crypto_wallet_connections_user_id_wallet_address_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.crypto_wallet_connections
    ADD CONSTRAINT crypto_wallet_connections_user_id_wallet_address_key UNIQUE (user_id, wallet_address);


--
-- Name: crypto_wallet_connections crypto_wallet_connections_wallet_address_wallet_type_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.crypto_wallet_connections
    ADD CONSTRAINT crypto_wallet_connections_wallet_address_wallet_type_key UNIQUE (wallet_address, wallet_type);


--
-- Name: currencies currencies_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.currencies
    ADD CONSTRAINT currencies_pkey PRIMARY KEY (id);


--
-- Name: departements_users departements_users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.departements_users
    ADD CONSTRAINT departements_users_pkey PRIMARY KEY (id);


--
-- Name: departments departments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.departments
    ADD CONSTRAINT departments_pkey PRIMARY KEY (id);


--
-- Name: exchange_rates exchange_rates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exchange_rates
    ADD CONSTRAINT exchange_rates_pkey PRIMARY KEY (id);


--
-- Name: failed_bet_logs failed_bet_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.failed_bet_logs
    ADD CONSTRAINT failed_bet_logs_pkey PRIMARY KEY (id);


--
-- Name: football_match_rounds football_match_rounds_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.football_match_rounds
    ADD CONSTRAINT football_match_rounds_pkey PRIMARY KEY (id);


--
-- Name: football_matchs football_matchs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.football_matchs
    ADD CONSTRAINT football_matchs_pkey PRIMARY KEY (id);


--
-- Name: game_logs game_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.game_logs
    ADD CONSTRAINT game_logs_pkey PRIMARY KEY (id);


--
-- Name: games games_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.games
    ADD CONSTRAINT games_pkey PRIMARY KEY (id);


--
-- Name: ip_filters ip_filters_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ip_filters
    ADD CONSTRAINT ip_filters_pkey PRIMARY KEY (id);


--
-- Name: leagues leagues_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.leagues
    ADD CONSTRAINT leagues_pkey PRIMARY KEY (id);


--
-- Name: level_requirements level_requirements_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.level_requirements
    ADD CONSTRAINT level_requirements_pkey PRIMARY KEY (id);


--
-- Name: levels levels_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.levels
    ADD CONSTRAINT levels_pkey PRIMARY KEY (id);


--
-- Name: login_attempts login_attempts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.login_attempts
    ADD CONSTRAINT login_attempts_pkey PRIMARY KEY (id);


--
-- Name: logs logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.logs
    ADD CONSTRAINT logs_pkey PRIMARY KEY (id);


--
-- Name: loot_box loot_box_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.loot_box
    ADD CONSTRAINT loot_box_pkey PRIMARY KEY (id);


--
-- Name: loot_box_place_bets loot_box_place_bets_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.loot_box_place_bets
    ADD CONSTRAINT loot_box_place_bets_pkey PRIMARY KEY (id);


--
-- Name: lotteries lotteries_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lotteries
    ADD CONSTRAINT lotteries_pkey PRIMARY KEY (id);


--
-- Name: lottery_logs lottery_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lottery_logs
    ADD CONSTRAINT lottery_logs_pkey PRIMARY KEY (id);


--
-- Name: lottery_services lottery_services_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lottery_services
    ADD CONSTRAINT lottery_services_pkey PRIMARY KEY (id);


--
-- Name: lottery_winners_logs lottery_winners_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lottery_winners_logs
    ADD CONSTRAINT lottery_winners_logs_pkey PRIMARY KEY (id);


--
-- Name: manual_funds manual_funds_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.manual_funds
    ADD CONSTRAINT manual_funds_pkey PRIMARY KEY (id);


--
-- Name: operational_groups operational_groups_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.operational_groups
    ADD CONSTRAINT operational_groups_pkey PRIMARY KEY (id);


--
-- Name: operational_types operational_types_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.operational_types
    ADD CONSTRAINT operational_types_pkey PRIMARY KEY (id);


--
-- Name: otps otps_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.otps
    ADD CONSTRAINT otps_pkey PRIMARY KEY (id);


--
-- Name: permissions permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_pkey PRIMARY KEY (id);


--
-- Name: plinko plinko_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.plinko
    ADD CONSTRAINT plinko_pkey PRIMARY KEY (id);


--
-- Name: quick_hustles quick_hustles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.quick_hustles
    ADD CONSTRAINT quick_hustles_pkey PRIMARY KEY (id);


--
-- Name: risk_settings risk_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risk_settings
    ADD CONSTRAINT risk_settings_pkey PRIMARY KEY (id);


--
-- Name: role_permissions role_permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_pkey PRIMARY KEY (id);


--
-- Name: roles roles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id);


--
-- Name: roll_da_dice roll_da_dice_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.roll_da_dice
    ADD CONSTRAINT roll_da_dice_pkey PRIMARY KEY (id);


--
-- Name: rounds rounds_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.rounds
    ADD CONSTRAINT rounds_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: scratch_cards scratch_cards_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.scratch_cards
    ADD CONSTRAINT scratch_cards_pkey PRIMARY KEY (id);


--
-- Name: spinning_wheel_configs spinning_wheel_configs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.spinning_wheel_configs
    ADD CONSTRAINT spinning_wheel_configs_pkey PRIMARY KEY (id);


--
-- Name: spinning_wheel_mysteries spinning_wheel_mysteries_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.spinning_wheel_mysteries
    ADD CONSTRAINT spinning_wheel_mysteries_pkey PRIMARY KEY (id);


--
-- Name: spinning_wheel_rewards spinning_wheel_rewards_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.spinning_wheel_rewards
    ADD CONSTRAINT spinning_wheel_rewards_pkey PRIMARY KEY (id);


--
-- Name: spinning_wheels spinning_wheels_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.spinning_wheels
    ADD CONSTRAINT spinning_wheels_pkey PRIMARY KEY (id);


--
-- Name: sport_bets sport_bets_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sport_bets
    ADD CONSTRAINT sport_bets_pkey PRIMARY KEY (id);


--
-- Name: sport_bets sport_bets_transaction_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sport_bets
    ADD CONSTRAINT sport_bets_transaction_id_key UNIQUE (transaction_id);


--
-- Name: squads_earns squads_earns_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.squads_earns
    ADD CONSTRAINT squads_earns_pkey PRIMARY KEY (id);


--
-- Name: squads_memebers squads_memebers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.squads_memebers
    ADD CONSTRAINT squads_memebers_pkey PRIMARY KEY (id);


--
-- Name: squads squads_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.squads
    ADD CONSTRAINT squads_pkey PRIMARY KEY (id);


--
-- Name: street_kings street_kings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.street_kings
    ADD CONSTRAINT street_kings_pkey PRIMARY KEY (id);


--
-- Name: temp temp_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.temp
    ADD CONSTRAINT temp_pkey PRIMARY KEY (id);


--
-- Name: tournaments_claims tournaments_claims_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tournaments_claims
    ADD CONSTRAINT tournaments_claims_pkey PRIMARY KEY (id);


--
-- Name: tournaments tournaments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tournaments
    ADD CONSTRAINT tournaments_pkey PRIMARY KEY (id);


--
-- Name: banners unique_page; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.banners
    ADD CONSTRAINT unique_page UNIQUE (page);


--
-- Name: user_notifications user_notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_notifications
    ADD CONSTRAINT user_notifications_pkey PRIMARY KEY (id);


--
-- Name: user_roles user_roles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_pkey PRIMARY KEY (id);


--
-- Name: user_sessions user_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_sessions
    ADD CONSTRAINT user_sessions_pkey PRIMARY KEY (id);


--
-- Name: users_football_matche_rounds users_football_matche_rounds_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users_football_matche_rounds
    ADD CONSTRAINT users_football_matche_rounds_pkey PRIMARY KEY (id);


--
-- Name: users_football_matches users_football_matches_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users_football_matches
    ADD CONSTRAINT users_football_matches_pkey PRIMARY KEY (id);


--
-- Name: users_otp users_otp_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users_otp
    ADD CONSTRAINT users_otp_pkey PRIMARY KEY (id);


--
-- Name: users users_phone_number_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_phone_number_key UNIQUE (phone_number);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: users users_username_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_username_key UNIQUE (username);


--
-- Name: waiting_squad_members waiting_squad_members_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.waiting_squad_members
    ADD CONSTRAINT waiting_squad_members_pkey PRIMARY KEY (id);


--
-- Name: idx_adds_services_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_adds_services_created_at ON public.adds_services USING btree (created_at);


--
-- Name: idx_adds_services_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_adds_services_status ON public.adds_services USING btree (status);


--
-- Name: idx_agent_referrals_callback_attempts; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_agent_referrals_callback_attempts ON public.agent_referrals USING btree (callback_attempts);


--
-- Name: idx_agent_referrals_callback_sent; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_agent_referrals_callback_sent ON public.agent_referrals USING btree (callback_sent);


--
-- Name: idx_agent_referrals_converted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_agent_referrals_converted_at ON public.agent_referrals USING btree (converted_at);


--
-- Name: idx_agent_referrals_request_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_agent_referrals_request_id ON public.agent_referrals USING btree (request_id);


--
-- Name: idx_agent_referrals_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_agent_referrals_user_id ON public.agent_referrals USING btree (user_id);


--
-- Name: idx_casbin_rule; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_casbin_rule ON public.casbin_rule USING btree (ptype, v0, v1, v2, v3, v4, v5);


--
-- Name: idx_crypto_wallet_auth_logs_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_crypto_wallet_auth_logs_created_at ON public.crypto_wallet_auth_logs USING btree (created_at);


--
-- Name: idx_crypto_wallet_auth_logs_wallet_address; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_crypto_wallet_auth_logs_wallet_address ON public.crypto_wallet_auth_logs USING btree (wallet_address);


--
-- Name: idx_crypto_wallet_challenges_expires_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_crypto_wallet_challenges_expires_at ON public.crypto_wallet_challenges USING btree (expires_at);


--
-- Name: idx_crypto_wallet_challenges_wallet_address; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_crypto_wallet_challenges_wallet_address ON public.crypto_wallet_challenges USING btree (wallet_address);


--
-- Name: idx_crypto_wallet_connections_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_crypto_wallet_connections_user_id ON public.crypto_wallet_connections USING btree (user_id);


--
-- Name: idx_crypto_wallet_connections_wallet_address; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_crypto_wallet_connections_wallet_address ON public.crypto_wallet_connections USING btree (wallet_address);


--
-- Name: idx_crypto_wallet_connections_wallet_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_crypto_wallet_connections_wallet_type ON public.crypto_wallet_connections USING btree (wallet_type);


--
-- Name: idx_otps_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_otps_created_at ON public.otps USING btree (created_at);


--
-- Name: idx_otps_email; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_otps_email ON public.otps USING btree (email);


--
-- Name: idx_otps_email_type_created; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_otps_email_type_created ON public.otps USING btree (email, type, created_at DESC);


--
-- Name: idx_otps_expires_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_otps_expires_at ON public.otps USING btree (expires_at);


--
-- Name: idx_otps_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_otps_status ON public.otps USING btree (status);


--
-- Name: idx_otps_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_otps_type ON public.otps USING btree (type);


--
-- Name: idx_sport_bets_bet_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sport_bets_bet_status ON public.sport_bets USING btree (status);


--
-- Name: idx_sport_bets_placed_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sport_bets_placed_at ON public.sport_bets USING btree (placed_at);


--
-- Name: idx_sport_bets_transaction_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sport_bets_transaction_id ON public.sport_bets USING btree (transaction_id);


--
-- Name: idx_sport_bets_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sport_bets_user_id ON public.sport_bets USING btree (user_id);


--
-- Name: idx_squads_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_squads_type ON public.squads USING btree (type);


--
-- Name: idx_tournaments_rank; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tournaments_rank ON public.tournaments USING btree (rank);


--
-- Name: idx_waiting_squad_members_user_squad; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_waiting_squad_members_user_squad ON public.waiting_squad_members USING btree (user_id, squad_id);


--
-- Name: uniq_adds_services_service_id_active; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uniq_adds_services_service_id_active ON public.adds_services USING btree (service_id) WHERE (deleted_at IS NULL);


--
-- Name: uniq_adds_services_service_id_deleted; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uniq_adds_services_service_id_deleted ON public.adds_services USING btree (service_id, deleted_at);


--
-- Name: uniq_company_support_phone_active; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uniq_company_support_phone_active ON public.company USING btree (support_phone) WHERE (deleted_at IS NULL);


--
-- Name: uniq_company_support_phone_deleted; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uniq_company_support_phone_deleted ON public.company USING btree (support_phone, deleted_at);


--
-- Name: uniq_levels_deleted; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uniq_levels_deleted ON public.levels USING btree (level, deleted_at);


--
-- Name: uniq_lottery_client_id_active; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uniq_lottery_client_id_active ON public.lottery_services USING btree (client_id) WHERE (deleted_at IS NULL);


--
-- Name: uniq_lottery_client_id_deleted; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uniq_lottery_client_id_deleted ON public.lottery_services USING btree (client_id, deleted_at);


--
-- Name: uniq_squads_active; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uniq_squads_active ON public.squads USING btree (handle) WHERE (deleted_at IS NULL);


--
-- Name: uniq_squads_deleted; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uniq_squads_deleted ON public.squads USING btree (handle, deleted_at);


--
-- Name: crypto_wallet_connections trigger_update_crypto_wallet_connections_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trigger_update_crypto_wallet_connections_updated_at BEFORE UPDATE ON public.crypto_wallet_connections FOR EACH ROW EXECUTE FUNCTION public.update_crypto_wallet_connections_updated_at();


--
-- Name: otps update_otps_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_otps_updated_at BEFORE UPDATE ON public.otps FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: account_block account_block_blocked_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.account_block
    ADD CONSTRAINT account_block_blocked_by_fkey FOREIGN KEY (blocked_by) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: account_block account_block_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.account_block
    ADD CONSTRAINT account_block_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: adds_services adds_services_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.adds_services
    ADD CONSTRAINT adds_services_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: airtime_transactions airtime_transactions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.airtime_transactions
    ADD CONSTRAINT airtime_transactions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: balance_logs balance_logs_operational_group_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.balance_logs
    ADD CONSTRAINT balance_logs_operational_group_id_fkey FOREIGN KEY (operational_group_id) REFERENCES public.operational_groups(id) ON DELETE CASCADE;


--
-- Name: balance_logs balance_logs_operational_type_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.balance_logs
    ADD CONSTRAINT balance_logs_operational_type_id_fkey FOREIGN KEY (operational_type_id) REFERENCES public.operational_types(id) ON DELETE CASCADE;


--
-- Name: balance_logs balance_logs_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.balance_logs
    ADD CONSTRAINT balance_logs_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: balances balances_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.balances
    ADD CONSTRAINT balances_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: bets bets_round_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.bets
    ADD CONSTRAINT bets_round_id_fkey FOREIGN KEY (round_id) REFERENCES public.rounds(id) ON DELETE CASCADE;


--
-- Name: bets bets_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.bets
    ADD CONSTRAINT bets_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: company company_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.company
    ADD CONSTRAINT company_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: crypto_kings crypto_kings_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.crypto_kings
    ADD CONSTRAINT crypto_kings_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: crypto_wallet_connections crypto_wallet_connections_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.crypto_wallet_connections
    ADD CONSTRAINT crypto_wallet_connections_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: departements_users departements_users_department_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.departements_users
    ADD CONSTRAINT departements_users_department_id_fkey FOREIGN KEY (department_id) REFERENCES public.departments(id) ON DELETE CASCADE;


--
-- Name: departements_users departements_users_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.departements_users
    ADD CONSTRAINT departements_users_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: failed_bet_logs failed_bet_logs_bet_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.failed_bet_logs
    ADD CONSTRAINT failed_bet_logs_bet_id_fkey FOREIGN KEY (bet_id) REFERENCES public.bets(id) ON DELETE CASCADE;


--
-- Name: failed_bet_logs failed_bet_logs_round_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.failed_bet_logs
    ADD CONSTRAINT failed_bet_logs_round_id_fkey FOREIGN KEY (round_id) REFERENCES public.rounds(id) ON DELETE CASCADE;


--
-- Name: failed_bet_logs failed_bet_logs_transaction_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.failed_bet_logs
    ADD CONSTRAINT failed_bet_logs_transaction_id_fkey FOREIGN KEY (transaction_id) REFERENCES public.balance_logs(id) ON DELETE CASCADE;


--
-- Name: failed_bet_logs failed_bet_logs_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.failed_bet_logs
    ADD CONSTRAINT failed_bet_logs_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: agent_referrals fk_agent_referrals_user_id; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_referrals
    ADD CONSTRAINT fk_agent_referrals_user_id FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: levels fk_created_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.levels
    ADD CONSTRAINT fk_created_by FOREIGN KEY (created_by) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: level_requirements fk_created_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.level_requirements
    ADD CONSTRAINT fk_created_by FOREIGN KEY (created_by) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: level_requirements fk_level_id; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.level_requirements
    ADD CONSTRAINT fk_level_id FOREIGN KEY (level_id) REFERENCES public.levels(id) ON DELETE CASCADE;


--
-- Name: squads_earns fk_squad_earns_squad; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.squads_earns
    ADD CONSTRAINT fk_squad_earns_squad FOREIGN KEY (squad_id) REFERENCES public.squads(id) ON DELETE CASCADE;


--
-- Name: squads_earns fk_squad_earns_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.squads_earns
    ADD CONSTRAINT fk_squad_earns_user FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: squads fk_squad_owner; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.squads
    ADD CONSTRAINT fk_squad_owner FOREIGN KEY (owner) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: squads_memebers fk_squad_user_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.squads_memebers
    ADD CONSTRAINT fk_squad_user_by FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: logs fk_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.logs
    ADD CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: football_matchs football_matchs_round_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.football_matchs
    ADD CONSTRAINT football_matchs_round_id_fkey FOREIGN KEY (round_id) REFERENCES public.football_match_rounds(id) ON DELETE CASCADE;


--
-- Name: game_logs game_logs_round_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.game_logs
    ADD CONSTRAINT game_logs_round_id_fkey FOREIGN KEY (round_id) REFERENCES public.rounds(id) ON DELETE CASCADE;


--
-- Name: ip_filters ip_filters_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ip_filters
    ADD CONSTRAINT ip_filters_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: login_attempts login_attempts_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.login_attempts
    ADD CONSTRAINT login_attempts_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: lottery_winners_logs lottery_winners_logs_lottery_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lottery_winners_logs
    ADD CONSTRAINT lottery_winners_logs_lottery_id_fkey FOREIGN KEY (lottery_id) REFERENCES public.lotteries(id) ON DELETE CASCADE;


--
-- Name: lottery_winners_logs lottery_winners_logs_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lottery_winners_logs
    ADD CONSTRAINT lottery_winners_logs_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: manual_funds manual_funds_admin_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.manual_funds
    ADD CONSTRAINT manual_funds_admin_id_fkey FOREIGN KEY (admin_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: manual_funds manual_funds_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.manual_funds
    ADD CONSTRAINT manual_funds_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: operational_types operational_types_group_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.operational_types
    ADD CONSTRAINT operational_types_group_id_fkey FOREIGN KEY (group_id) REFERENCES public.operational_groups(id) ON DELETE CASCADE;


--
-- Name: plinko plinko_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.plinko
    ADD CONSTRAINT plinko_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: quick_hustles quick_hustles_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.quick_hustles
    ADD CONSTRAINT quick_hustles_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: role_permissions role_permissions_permission_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_permission_id_fkey FOREIGN KEY (permission_id) REFERENCES public.permissions(id) ON DELETE CASCADE;


--
-- Name: role_permissions role_permissions_role_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.roles(id) ON DELETE CASCADE;


--
-- Name: roll_da_dice roll_da_dice_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.roll_da_dice
    ADD CONSTRAINT roll_da_dice_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: scratch_cards scratch_cards_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.scratch_cards
    ADD CONSTRAINT scratch_cards_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: spinning_wheel_configs spinning_wheel_configs_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.spinning_wheel_configs
    ADD CONSTRAINT spinning_wheel_configs_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: spinning_wheel_mysteries spinning_wheel_mysteries_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.spinning_wheel_mysteries
    ADD CONSTRAINT spinning_wheel_mysteries_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: spinning_wheel_rewards spinning_wheel_rewards_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.spinning_wheel_rewards
    ADD CONSTRAINT spinning_wheel_rewards_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: spinning_wheels spinning_wheels_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.spinning_wheels
    ADD CONSTRAINT spinning_wheels_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: sport_bets sport_bets_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sport_bets
    ADD CONSTRAINT sport_bets_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: street_kings street_kings_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.street_kings
    ADD CONSTRAINT street_kings_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: tournaments_claims tournaments_claims_squad_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tournaments_claims
    ADD CONSTRAINT tournaments_claims_squad_id_fkey FOREIGN KEY (squad_id) REFERENCES public.squads(id) ON DELETE CASCADE;


--
-- Name: tournaments_claims tournaments_claims_tournament_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tournaments_claims
    ADD CONSTRAINT tournaments_claims_tournament_id_fkey FOREIGN KEY (tournament_id) REFERENCES public.tournaments(id) ON DELETE CASCADE;


--
-- Name: user_notifications user_notifications_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_notifications
    ADD CONSTRAINT user_notifications_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: user_roles user_roles_role_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.roles(id) ON DELETE CASCADE;


--
-- Name: user_roles user_roles_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_sessions user_sessions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_sessions
    ADD CONSTRAINT user_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: users_football_matche_rounds users_football_matche_rounds_football_round_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users_football_matche_rounds
    ADD CONSTRAINT users_football_matche_rounds_football_round_id_fkey FOREIGN KEY (football_round_id) REFERENCES public.football_match_rounds(id) ON DELETE CASCADE;


--
-- Name: users_football_matche_rounds users_football_matche_rounds_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users_football_matche_rounds
    ADD CONSTRAINT users_football_matche_rounds_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: users_football_matches users_football_matches_match_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users_football_matches
    ADD CONSTRAINT users_football_matches_match_id_fkey FOREIGN KEY (match_id) REFERENCES public.football_matchs(id) ON DELETE CASCADE;


--
-- Name: waiting_squad_members waiting_squad_members_squad_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.waiting_squad_members
    ADD CONSTRAINT waiting_squad_members_squad_id_fkey FOREIGN KEY (squad_id) REFERENCES public.squads(id) ON DELETE CASCADE;


--
-- Name: waiting_squad_members waiting_squad_members_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.waiting_squad_members
    ADD CONSTRAINT waiting_squad_members_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

