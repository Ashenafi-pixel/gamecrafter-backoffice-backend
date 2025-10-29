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

SET default_tablespace = '';

SET default_table_access_method = heap;

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
-- Name: currencies; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.currencies (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying NOT NULL,
    status character varying DEFAULT 'ACTIVE'::character varying NOT NULL,
    "timestamp" timestamp without time zone DEFAULT now() NOT NULL
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
-- Data for Name: company; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.company (id, site_name, support_email, support_phone, maintenance_mode, maximum_login_attempt, password_expiry, lockout_duration, require_two_factor_authentication, ip_list, created_by, created_at, updated_at, deleted_at) FROM stdin;
\.


--
-- Data for Name: currencies; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.currencies (id, name, status, "timestamp") FROM stdin;
\.


--
-- Data for Name: games; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.games (id, name, status, "timestamp", photo, price, enabled) FROM stdin;
cfb2c688-0d30-46ea-ba7e-6ee2b29a8443	TucanBIT	ACTIVE	2025-08-29 21:02:46.355312	\N	\N	f
843495fe-c0b7-451f-b1d2-e68b86d06008	Plinko	ACTIVE	2025-08-29 21:02:46.356095	\N	\N	f
e567e3b0-a432-4062-84e5-dca294aa2479	Crypto_kings	ACTIVE	2025-08-29 21:02:46.356566	\N	\N	f
f144c263-911f-4bf6-b1d9-90b8efb92a9d	Football fixtures	ACTIVE	2025-08-29 21:02:46.357236	\N	\N	f
8d2ca9f6-1c0e-46ca-975f-72ca3afed060	Quick hustle	ACTIVE	2025-08-29 21:02:46.3577	\N	\N	f
b2a8cd89-83a0-40b5-8803-46a079ac245b	Roll Da Dice	ACTIVE	2025-08-29 21:02:46.359971	\N	\N	f
22ab4676-6657-410b-b030-4344d0ee1937	Scratch Card	ACTIVE	2025-08-29 21:02:46.360503	\N	\N	f
66f1020e-d9c8-4152-94c0-fc7b812b0016	Spinning Wheel	ACTIVE	2025-08-29 21:02:46.361111	\N	\N	f
5a729589-987d-4862-b1a5-3c32831da50d	Street Kings	ACTIVE	2025-08-29 21:02:46.361552	\N	\N	f
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.users (id, username, phone_number, password, created_at, default_currency, profile, email, first_name, last_name, date_of_birth, source, is_email_verified, referal_code, street_address, country, state, city, postal_code, kyc_status, created_by, is_admin, status, referal_type, refered_by_code, user_type, primary_wallet_address, wallet_verification_status) FROM stdin;
e3588cc3-0578-4939-ab9e-597d1216f2a5	player_decf5b93	\N		2025-09-03 23:49:31.81907	USD		\N				wallet	f	HaJPhpfib2Ip						PENDING	\N	f	active	PLAYER	\N	PLAYER	\N	none
99d03999-0f02-4c97-9bc3-d7d3bb8fd885	player_5a517fd1	\N		2025-09-03 23:51:13.955767	USD		\N				wallet	f	2dRDYTVv1fle						PENDING	\N	f	active	PLAYER	\N	PLAYER	\N	none
a6147b2f-bd0b-4ff5-b94d-fd431d1aa074		\N		2025-08-31 19:07:38.597788			test@example.com					f	I6E2P3E1						PENDING	\N	f	\N	\N	\N	PLAYER	\N	none
a5e168fb-168e-4183-84c5-d49038ce00b5	P-uBKtmkyl5LPo	+251912308971	$2a$10$kCLcXu5lrI3/Qee.oEXfiu1iv1EYj8/XVSVOEhbljzQedaWo7Te/O	2025-08-31 20:25:57.127123	\N		ashenafialemu9898@gmail.com	Ashenafi	Alemu			t	ouWiLgN0RZDB						PENDING	\N	\N	ACTIVE	\N	\N	PLAYER	\N	none
47781512-f533-45f5-8e6e-14d239e890d5	P-f2zhknBSxyfh	+25191122014	$2a$10$5WTIfXdwnb8HZPwMKru1WOY3Qr6olelnbo5sw2fo2n/yiXqxRZU6y	2025-08-31 20:30:37.351669	P		ashenafi.mlm@gmail.com	Ashenafi	Alemu	1992-05-15	facebook	t	P-f2zhknBSxyfh						PENDING	\N	f	\N	PLAYER	REF999	PLAYER	\N	none
21e96e96-f70d-4eb8-a5eb-0ff6c2bc33eb	player_87879fe4	\N		2025-09-03 23:53:40.235362	USD		\N				wallet	f	p0iceouHqsPK						PENDING	\N	f	active	PLAYER	\N	PLAYER	\N	none
3fbc49ac-db45-40c4-a949-ade55082662e	player_72ee1fc4	\N		2025-09-03 23:47:58.477659	USD		\N				wallet	f	lc2zFJy4HGxx						PENDING	\N	f	active	PLAYER	\N	PLAYER	0xa9ce394f87cf36f3b7ddfd85824c2218270d25c1	verified
29ff053e-cc63-4894-bc2f-a38aaeca3244	player_de82421d	\N		2025-09-03 23:48:36.782038	USD		\N				wallet	f	wddKUHPkBvhj						PENDING	\N	f	active	PLAYER	\N	PLAYER	\N	none
f12e2768-0c41-40af-9c12-0d264a76d5ca	player_099d98f1	\N		2025-09-03 23:48:43.230372	USD		\N				wallet	f	wddKUHPkBvhj						PENDING	\N	f	active	PLAYER	\N	PLAYER	0x1234567890abcdef1234567890abcdef12345678	verified
\.


--
-- Name: company company_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.company
    ADD CONSTRAINT company_pkey PRIMARY KEY (id);


--
-- Name: currencies currencies_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.currencies
    ADD CONSTRAINT currencies_pkey PRIMARY KEY (id);


--
-- Name: games games_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.games
    ADD CONSTRAINT games_pkey PRIMARY KEY (id);


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
-- Name: uniq_company_support_phone_active; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uniq_company_support_phone_active ON public.company USING btree (support_phone) WHERE (deleted_at IS NULL);


--
-- Name: uniq_company_support_phone_deleted; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uniq_company_support_phone_deleted ON public.company USING btree (support_phone, deleted_at);


--
-- Name: company company_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.company
    ADD CONSTRAINT company_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- PostgreSQL database dump complete
--

