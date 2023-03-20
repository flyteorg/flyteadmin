--
-- PostgreSQL database dump
--

-- Dumped from database version 15.1
-- Dumped by pg_dump version 15.1

-- Started on 2023-03-20 14:15:17 PDT

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
-- TOC entry 236 (class 1259 OID 16540)
-- Name: artifact_data; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.artifact_data (
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    dataset_project text NOT NULL,
    dataset_name text NOT NULL,
    dataset_domain text NOT NULL,
    dataset_version text NOT NULL,
    artifact_id text NOT NULL,
    name text NOT NULL,
    location text
);


ALTER TABLE public.artifact_data OWNER TO postgres;

--
-- TOC entry 233 (class 1259 OID 16516)
-- Name: artifacts; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.artifacts (
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    dataset_project text NOT NULL,
    dataset_name text NOT NULL,
    dataset_domain text NOT NULL,
    dataset_version text NOT NULL,
    artifact_id text NOT NULL,
    dataset_uuid uuid,
    serialized_metadata bytea
);


ALTER TABLE public.artifacts OWNER TO postgres;

--
-- TOC entry 215 (class 1259 OID 16389)
-- Name: datasets; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.datasets (
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    project text NOT NULL,
    name text NOT NULL,
    domain text NOT NULL,
    version text NOT NULL,
    uuid uuid,
    serialized_metadata bytea
);


ALTER TABLE public.datasets OWNER TO postgres;

--
-- TOC entry 249 (class 1259 OID 16646)
-- Name: description_entities; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.description_entities (
    resource_type integer NOT NULL,
    project text NOT NULL,
    domain text NOT NULL,
    name text NOT NULL,
    version text NOT NULL,
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    short_description text,
    long_description bytea,
    link text
);


ALTER TABLE public.description_entities OWNER TO postgres;

--
-- TOC entry 248 (class 1259 OID 16645)
-- Name: description_entities_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.description_entities_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.description_entities_id_seq OWNER TO postgres;

--
-- TOC entry 3553 (class 0 OID 0)
-- Dependencies: 248
-- Name: description_entities_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.description_entities_id_seq OWNED BY public.description_entities.id;


--
-- TOC entry 228 (class 1259 OID 16484)
-- Name: execution_events; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.execution_events (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    execution_project text NOT NULL,
    execution_domain text NOT NULL,
    execution_name text NOT NULL,
    request_id text,
    occurred_at timestamp with time zone,
    phase text NOT NULL
);


ALTER TABLE public.execution_events OWNER TO postgres;

--
-- TOC entry 227 (class 1259 OID 16483)
-- Name: execution_events_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.execution_events_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.execution_events_id_seq OWNER TO postgres;

--
-- TOC entry 3554 (class 0 OID 0)
-- Dependencies: 227
-- Name: execution_events_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.execution_events_id_seq OWNED BY public.execution_events.id;


--
-- TOC entry 226 (class 1259 OID 16465)
-- Name: executions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.executions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    execution_project text NOT NULL,
    execution_domain text NOT NULL,
    execution_name text NOT NULL,
    launch_plan_id bigint,
    workflow_id bigint,
    task_id bigint,
    phase text,
    closure bytea,
    spec bytea NOT NULL,
    started_at timestamp with time zone,
    execution_created_at timestamp with time zone,
    execution_updated_at timestamp with time zone,
    duration bigint,
    abort_cause text,
    mode integer,
    source_execution_id bigint,
    parent_node_execution_id bigint,
    cluster text,
    inputs_uri text,
    user_inputs_uri text,
    error_kind text,
    error_code text,
    "user" text,
    state integer DEFAULT 0,
    launch_entity text
);


ALTER TABLE public.executions OWNER TO postgres;

--
-- TOC entry 225 (class 1259 OID 16464)
-- Name: executions_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.executions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.executions_id_seq OWNER TO postgres;

--
-- TOC entry 3555 (class 0 OID 0)
-- Dependencies: 225
-- Name: executions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.executions_id_seq OWNED BY public.executions.id;


--
-- TOC entry 224 (class 1259 OID 16451)
-- Name: launch_plans; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.launch_plans (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    project text NOT NULL,
    domain text NOT NULL,
    name text NOT NULL,
    version text NOT NULL,
    spec bytea NOT NULL,
    workflow_id bigint,
    closure bytea NOT NULL,
    state integer DEFAULT 0,
    digest bytea,
    schedule_type text
);


ALTER TABLE public.launch_plans OWNER TO postgres;

--
-- TOC entry 223 (class 1259 OID 16450)
-- Name: launch_plans_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.launch_plans_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.launch_plans_id_seq OWNER TO postgres;

--
-- TOC entry 3556 (class 0 OID 0)
-- Dependencies: 223
-- Name: launch_plans_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.launch_plans_id_seq OWNED BY public.launch_plans.id;


--
-- TOC entry 214 (class 1259 OID 16386)
-- Name: migrations; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.migrations (
    id character varying(255) NOT NULL
);


ALTER TABLE public.migrations OWNER TO postgres;

--
-- TOC entry 241 (class 1259 OID 16570)
-- Name: named_entity_metadata; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.named_entity_metadata (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    resource_type integer NOT NULL,
    project text NOT NULL,
    domain text NOT NULL,
    name text NOT NULL,
    description character varying(300),
    state integer DEFAULT 0
);


ALTER TABLE public.named_entity_metadata OWNER TO postgres;

--
-- TOC entry 240 (class 1259 OID 16569)
-- Name: named_entity_metadata_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.named_entity_metadata_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.named_entity_metadata_id_seq OWNER TO postgres;

--
-- TOC entry 3557 (class 0 OID 0)
-- Dependencies: 240
-- Name: named_entity_metadata_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.named_entity_metadata_id_seq OWNED BY public.named_entity_metadata.id;


--
-- TOC entry 232 (class 1259 OID 16508)
-- Name: node_execution_events; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.node_execution_events (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    execution_project text NOT NULL,
    execution_domain text NOT NULL,
    execution_name text NOT NULL,
    node_id text NOT NULL,
    request_id text,
    occurred_at timestamp with time zone,
    phase text NOT NULL
);


ALTER TABLE public.node_execution_events OWNER TO postgres;

--
-- TOC entry 231 (class 1259 OID 16507)
-- Name: node_execution_events_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.node_execution_events_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.node_execution_events_id_seq OWNER TO postgres;

--
-- TOC entry 3558 (class 0 OID 0)
-- Dependencies: 231
-- Name: node_execution_events_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.node_execution_events_id_seq OWNED BY public.node_execution_events.id;


--
-- TOC entry 230 (class 1259 OID 16495)
-- Name: node_executions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.node_executions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    execution_project text NOT NULL,
    execution_domain text NOT NULL,
    execution_name text NOT NULL,
    node_id text NOT NULL,
    phase text,
    input_uri text,
    closure bytea,
    started_at timestamp with time zone,
    node_execution_created_at timestamp with time zone,
    node_execution_updated_at timestamp with time zone,
    duration bigint,
    parent_task_execution_id bigint,
    dynamic_workflow_remote_closure_reference text,
    internal_data bytea,
    node_execution_metadata bytea,
    parent_id bigint,
    cache_status text,
    error_kind text,
    error_code text
);


ALTER TABLE public.node_executions OWNER TO postgres;

--
-- TOC entry 229 (class 1259 OID 16494)
-- Name: node_executions_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.node_executions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.node_executions_id_seq OWNER TO postgres;

--
-- TOC entry 3559 (class 0 OID 0)
-- Dependencies: 229
-- Name: node_executions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.node_executions_id_seq OWNED BY public.node_executions.id;


--
-- TOC entry 237 (class 1259 OID 16547)
-- Name: partition_keys; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.partition_keys (
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    dataset_uuid uuid NOT NULL,
    name text NOT NULL
);


ALTER TABLE public.partition_keys OWNER TO postgres;

--
-- TOC entry 238 (class 1259 OID 16554)
-- Name: partitions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.partitions (
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    dataset_uuid uuid NOT NULL,
    key text NOT NULL,
    value text NOT NULL,
    artifact_id text NOT NULL
);


ALTER TABLE public.partitions OWNER TO postgres;

--
-- TOC entry 217 (class 1259 OID 16399)
-- Name: projects; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.projects (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    identifier text NOT NULL,
    name text,
    description character varying(300),
    labels bytea,
    state integer DEFAULT 0
);


ALTER TABLE public.projects OWNER TO postgres;

--
-- TOC entry 216 (class 1259 OID 16397)
-- Name: projects_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.projects_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.projects_id_seq OWNER TO postgres;

--
-- TOC entry 3560 (class 0 OID 0)
-- Dependencies: 216
-- Name: projects_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.projects_id_seq OWNED BY public.projects.id;


--
-- TOC entry 239 (class 1259 OID 16562)
-- Name: reservations; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.reservations (
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    dataset_project text NOT NULL,
    dataset_name text NOT NULL,
    dataset_domain text NOT NULL,
    dataset_version text NOT NULL,
    tag_name text NOT NULL,
    owner_id text,
    expires_at timestamp with time zone,
    serialized_metadata bytea
);


ALTER TABLE public.reservations OWNER TO postgres;

--
-- TOC entry 243 (class 1259 OID 16583)
-- Name: resources; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.resources (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    project text,
    domain text,
    workflow text,
    launch_plan text,
    resource_type text,
    priority integer,
    attributes bytea
);


ALTER TABLE public.resources OWNER TO postgres;

--
-- TOC entry 242 (class 1259 OID 16582)
-- Name: resources_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.resources_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.resources_id_seq OWNER TO postgres;

--
-- TOC entry 3561 (class 0 OID 0)
-- Dependencies: 242
-- Name: resources_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.resources_id_seq OWNED BY public.resources.id;


--
-- TOC entry 245 (class 1259 OID 16595)
-- Name: schedulable_entities; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.schedulable_entities (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    project text NOT NULL,
    domain text NOT NULL,
    name text NOT NULL,
    version text NOT NULL,
    cron_expression text,
    fixed_rate_value bigint,
    unit integer,
    kickoff_time_input_arg text,
    active boolean
);


ALTER TABLE public.schedulable_entities OWNER TO postgres;

--
-- TOC entry 244 (class 1259 OID 16594)
-- Name: schedulable_entities_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.schedulable_entities_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.schedulable_entities_id_seq OWNER TO postgres;

--
-- TOC entry 3562 (class 0 OID 0)
-- Dependencies: 244
-- Name: schedulable_entities_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.schedulable_entities_id_seq OWNED BY public.schedulable_entities.id;


--
-- TOC entry 247 (class 1259 OID 16606)
-- Name: schedule_entities_snapshots; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.schedule_entities_snapshots (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    snapshot bytea
);


ALTER TABLE public.schedule_entities_snapshots OWNER TO postgres;

--
-- TOC entry 246 (class 1259 OID 16605)
-- Name: schedule_entities_snapshots_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.schedule_entities_snapshots_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.schedule_entities_snapshots_id_seq OWNER TO postgres;

--
-- TOC entry 3563 (class 0 OID 0)
-- Dependencies: 246
-- Name: schedule_entities_snapshots_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.schedule_entities_snapshots_id_seq OWNED BY public.schedule_entities_snapshots.id;


--
-- TOC entry 251 (class 1259 OID 16658)
-- Name: signals; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.signals (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    execution_project text NOT NULL,
    execution_domain text NOT NULL,
    execution_name text NOT NULL,
    signal_id text NOT NULL,
    type bytea NOT NULL,
    value bytea
);


ALTER TABLE public.signals OWNER TO postgres;

--
-- TOC entry 250 (class 1259 OID 16657)
-- Name: signals_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.signals_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.signals_id_seq OWNER TO postgres;

--
-- TOC entry 3564 (class 0 OID 0)
-- Dependencies: 250
-- Name: signals_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.signals_id_seq OWNED BY public.signals.id;


--
-- TOC entry 218 (class 1259 OID 16416)
-- Name: tags; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.tags (
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    dataset_project text NOT NULL,
    dataset_name text NOT NULL,
    dataset_domain text NOT NULL,
    dataset_version text NOT NULL,
    tag_name text NOT NULL,
    artifact_id text,
    dataset_uuid uuid
);


ALTER TABLE public.tags OWNER TO postgres;

--
-- TOC entry 235 (class 1259 OID 16528)
-- Name: task_executions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.task_executions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    project text NOT NULL,
    domain text NOT NULL,
    name text NOT NULL,
    version text NOT NULL,
    execution_project text NOT NULL,
    execution_domain text NOT NULL,
    execution_name text NOT NULL,
    node_id text NOT NULL,
    retry_attempt bigint NOT NULL,
    phase text,
    phase_version bigint,
    input_uri text,
    closure bytea,
    started_at timestamp with time zone,
    task_execution_created_at timestamp with time zone,
    task_execution_updated_at timestamp with time zone,
    duration bigint
);


ALTER TABLE public.task_executions OWNER TO postgres;

--
-- TOC entry 234 (class 1259 OID 16527)
-- Name: task_executions_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.task_executions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.task_executions_id_seq OWNER TO postgres;

--
-- TOC entry 3565 (class 0 OID 0)
-- Dependencies: 234
-- Name: task_executions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.task_executions_id_seq OWNED BY public.task_executions.id;


--
-- TOC entry 220 (class 1259 OID 16425)
-- Name: tasks; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.tasks (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    project text NOT NULL,
    domain text NOT NULL,
    name text NOT NULL,
    version text NOT NULL,
    closure bytea NOT NULL,
    digest bytea,
    type text,
    short_description text
);


ALTER TABLE public.tasks OWNER TO postgres;

--
-- TOC entry 219 (class 1259 OID 16423)
-- Name: tasks_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.tasks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.tasks_id_seq OWNER TO postgres;

--
-- TOC entry 3566 (class 0 OID 0)
-- Dependencies: 219
-- Name: tasks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.tasks_id_seq OWNED BY public.tasks.id;


--
-- TOC entry 222 (class 1259 OID 16438)
-- Name: workflows; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.workflows (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    project text NOT NULL,
    domain text NOT NULL,
    name text NOT NULL,
    version text NOT NULL,
    typed_interface bytea,
    remote_closure_identifier text NOT NULL,
    digest bytea,
    short_description text
);


ALTER TABLE public.workflows OWNER TO postgres;

--
-- TOC entry 221 (class 1259 OID 16437)
-- Name: workflows_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.workflows_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.workflows_id_seq OWNER TO postgres;

--
-- TOC entry 3567 (class 0 OID 0)
-- Dependencies: 221
-- Name: workflows_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.workflows_id_seq OWNED BY public.workflows.id;


--
-- TOC entry 3297 (class 2604 OID 16649)
-- Name: description_entities id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.description_entities ALTER COLUMN id SET DEFAULT nextval('public.description_entities_id_seq'::regclass);


--
-- TOC entry 3288 (class 2604 OID 16616)
-- Name: execution_events id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.execution_events ALTER COLUMN id SET DEFAULT nextval('public.execution_events_id_seq'::regclass);


--
-- TOC entry 3286 (class 2604 OID 16618)
-- Name: executions id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.executions ALTER COLUMN id SET DEFAULT nextval('public.executions_id_seq'::regclass);


--
-- TOC entry 3284 (class 2604 OID 16620)
-- Name: launch_plans id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.launch_plans ALTER COLUMN id SET DEFAULT nextval('public.launch_plans_id_seq'::regclass);


--
-- TOC entry 3292 (class 2604 OID 16622)
-- Name: named_entity_metadata id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.named_entity_metadata ALTER COLUMN id SET DEFAULT nextval('public.named_entity_metadata_id_seq'::regclass);


--
-- TOC entry 3290 (class 2604 OID 16624)
-- Name: node_execution_events id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.node_execution_events ALTER COLUMN id SET DEFAULT nextval('public.node_execution_events_id_seq'::regclass);


--
-- TOC entry 3289 (class 2604 OID 16626)
-- Name: node_executions id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.node_executions ALTER COLUMN id SET DEFAULT nextval('public.node_executions_id_seq'::regclass);


--
-- TOC entry 3280 (class 2604 OID 16628)
-- Name: projects id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.projects ALTER COLUMN id SET DEFAULT nextval('public.projects_id_seq'::regclass);


--
-- TOC entry 3294 (class 2604 OID 16630)
-- Name: resources id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.resources ALTER COLUMN id SET DEFAULT nextval('public.resources_id_seq'::regclass);


--
-- TOC entry 3295 (class 2604 OID 16633)
-- Name: schedulable_entities id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.schedulable_entities ALTER COLUMN id SET DEFAULT nextval('public.schedulable_entities_id_seq'::regclass);


--
-- TOC entry 3296 (class 2604 OID 16635)
-- Name: schedule_entities_snapshots id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.schedule_entities_snapshots ALTER COLUMN id SET DEFAULT nextval('public.schedule_entities_snapshots_id_seq'::regclass);


--
-- TOC entry 3298 (class 2604 OID 16661)
-- Name: signals id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.signals ALTER COLUMN id SET DEFAULT nextval('public.signals_id_seq'::regclass);


--
-- TOC entry 3291 (class 2604 OID 16639)
-- Name: task_executions id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.task_executions ALTER COLUMN id SET DEFAULT nextval('public.task_executions_id_seq'::regclass);


--
-- TOC entry 3282 (class 2604 OID 16641)
-- Name: tasks id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tasks ALTER COLUMN id SET DEFAULT nextval('public.tasks_id_seq'::regclass);


--
-- TOC entry 3283 (class 2604 OID 16643)
-- Name: workflows id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.workflows ALTER COLUMN id SET DEFAULT nextval('public.workflows_id_seq'::regclass);


--
-- TOC entry 3372 (class 2606 OID 16546)
-- Name: artifact_data artifact_data_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.artifact_data
    ADD CONSTRAINT artifact_data_pkey PRIMARY KEY (dataset_project, dataset_name, dataset_domain, dataset_version, artifact_id, name);


--
-- TOC entry 3364 (class 2606 OID 16523)
-- Name: artifacts artifacts_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.artifacts
    ADD CONSTRAINT artifacts_pkey PRIMARY KEY (dataset_project, dataset_name, dataset_domain, dataset_version, artifact_id);


--
-- TOC entry 3305 (class 2606 OID 16398)
-- Name: datasets datasets_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.datasets
    ADD CONSTRAINT datasets_pkey PRIMARY KEY (project, name, domain, version);


--
-- TOC entry 3307 (class 2606 OID 16404)
-- Name: datasets datasets_uuid_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.datasets
    ADD CONSTRAINT datasets_uuid_key UNIQUE (uuid);


--
-- TOC entry 3397 (class 2606 OID 16653)
-- Name: description_entities description_entities_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.description_entities
    ADD CONSTRAINT description_entities_pkey PRIMARY KEY (resource_type, project, domain, name, version);


--
-- TOC entry 3346 (class 2606 OID 16491)
-- Name: execution_events execution_events_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.execution_events
    ADD CONSTRAINT execution_events_pkey PRIMARY KEY (execution_project, execution_domain, execution_name, phase);


--
-- TOC entry 3335 (class 2606 OID 16473)
-- Name: executions executions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.executions
    ADD CONSTRAINT executions_pkey PRIMARY KEY (execution_project, execution_domain, execution_name);


--
-- TOC entry 3332 (class 2606 OID 16459)
-- Name: launch_plans launch_plans_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.launch_plans
    ADD CONSTRAINT launch_plans_pkey PRIMARY KEY (project, domain, name, version);


--
-- TOC entry 3300 (class 2606 OID 16393)
-- Name: migrations migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.migrations
    ADD CONSTRAINT migrations_pkey PRIMARY KEY (id);


--
-- TOC entry 3383 (class 2606 OID 16578)
-- Name: named_entity_metadata named_entity_metadata_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.named_entity_metadata
    ADD CONSTRAINT named_entity_metadata_pkey PRIMARY KEY (resource_type, project, domain, name);


--
-- TOC entry 3361 (class 2606 OID 16515)
-- Name: node_execution_events node_execution_events_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.node_execution_events
    ADD CONSTRAINT node_execution_events_pkey PRIMARY KEY (execution_project, execution_domain, execution_name, node_id, phase);


--
-- TOC entry 3356 (class 2606 OID 16502)
-- Name: node_executions node_executions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.node_executions
    ADD CONSTRAINT node_executions_pkey PRIMARY KEY (execution_project, execution_domain, execution_name, node_id);


--
-- TOC entry 3374 (class 2606 OID 16553)
-- Name: partition_keys partition_keys_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.partition_keys
    ADD CONSTRAINT partition_keys_pkey PRIMARY KEY (dataset_uuid, name);


--
-- TOC entry 3377 (class 2606 OID 16560)
-- Name: partitions partitions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.partitions
    ADD CONSTRAINT partitions_pkey PRIMARY KEY (dataset_uuid, key, value, artifact_id);


--
-- TOC entry 3312 (class 2606 OID 16410)
-- Name: projects projects_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_pkey PRIMARY KEY (identifier);


--
-- TOC entry 3379 (class 2606 OID 16568)
-- Name: reservations reservations_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.reservations
    ADD CONSTRAINT reservations_pkey PRIMARY KEY (dataset_project, dataset_name, dataset_domain, dataset_version, tag_name);


--
-- TOC entry 3387 (class 2606 OID 16632)
-- Name: resources resources_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.resources
    ADD CONSTRAINT resources_pkey PRIMARY KEY (id);


--
-- TOC entry 3391 (class 2606 OID 16602)
-- Name: schedulable_entities schedulable_entities_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.schedulable_entities
    ADD CONSTRAINT schedulable_entities_pkey PRIMARY KEY (project, domain, name, version);


--
-- TOC entry 3395 (class 2606 OID 16637)
-- Name: schedule_entities_snapshots schedule_entities_snapshots_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.schedule_entities_snapshots
    ADD CONSTRAINT schedule_entities_snapshots_pkey PRIMARY KEY (id);


--
-- TOC entry 3405 (class 2606 OID 16665)
-- Name: signals signals_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.signals
    ADD CONSTRAINT signals_pkey PRIMARY KEY (execution_project, execution_domain, execution_name, signal_id);


--
-- TOC entry 3315 (class 2606 OID 16422)
-- Name: tags tags_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tags
    ADD CONSTRAINT tags_pkey PRIMARY KEY (dataset_project, dataset_name, dataset_domain, dataset_version, tag_name);


--
-- TOC entry 3370 (class 2606 OID 16535)
-- Name: task_executions task_executions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.task_executions
    ADD CONSTRAINT task_executions_pkey PRIMARY KEY (project, domain, name, version, execution_project, execution_domain, execution_name, node_id, retry_attempt);


--
-- TOC entry 3321 (class 2606 OID 16432)
-- Name: tasks tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_pkey PRIMARY KEY (project, domain, name, version);


--
-- TOC entry 3327 (class 2606 OID 16445)
-- Name: workflows workflows_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.workflows
    ADD CONSTRAINT workflows_pkey PRIMARY KEY (project, domain, name, version);


--
-- TOC entry 3362 (class 1259 OID 16526)
-- Name: artifacts_dataset_uuid_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX artifacts_dataset_uuid_idx ON public.artifacts USING btree (dataset_uuid);


--
-- TOC entry 3301 (class 1259 OID 16411)
-- Name: dataset_domain_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX dataset_domain_idx ON public.datasets USING btree (domain);


--
-- TOC entry 3302 (class 1259 OID 16412)
-- Name: dataset_name_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX dataset_name_idx ON public.datasets USING btree (name);


--
-- TOC entry 3303 (class 1259 OID 16408)
-- Name: dataset_version_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX dataset_version_idx ON public.datasets USING btree (version);


--
-- TOC entry 3398 (class 1259 OID 16656)
-- Name: description_entity_project_domain_name_version_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX description_entity_project_domain_name_version_idx ON public.description_entities USING btree (resource_type, project, domain, name, version);


--
-- TOC entry 3399 (class 1259 OID 16654)
-- Name: idx_description_entities_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_description_entities_deleted_at ON public.description_entities USING btree (deleted_at);


--
-- TOC entry 3400 (class 1259 OID 16655)
-- Name: idx_description_entities_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_description_entities_id ON public.description_entities USING btree (id);


--
-- TOC entry 3347 (class 1259 OID 16492)
-- Name: idx_execution_events_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_execution_events_deleted_at ON public.execution_events USING btree (deleted_at);


--
-- TOC entry 3348 (class 1259 OID 16617)
-- Name: idx_execution_events_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_execution_events_id ON public.execution_events USING btree (id);


--
-- TOC entry 3336 (class 1259 OID 16481)
-- Name: idx_executions_created_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_executions_created_at ON public.executions USING btree (execution_created_at);


--
-- TOC entry 3337 (class 1259 OID 16478)
-- Name: idx_executions_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_executions_deleted_at ON public.executions USING btree (deleted_at);


--
-- TOC entry 3338 (class 1259 OID 16480)
-- Name: idx_executions_error_kind; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_executions_error_kind ON public.executions USING btree (error_kind);


--
-- TOC entry 3339 (class 1259 OID 16619)
-- Name: idx_executions_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_executions_id ON public.executions USING btree (id);


--
-- TOC entry 3340 (class 1259 OID 16474)
-- Name: idx_executions_launch_plan_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_executions_launch_plan_id ON public.executions USING btree (launch_plan_id);


--
-- TOC entry 3341 (class 1259 OID 16476)
-- Name: idx_executions_state; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_executions_state ON public.executions USING btree (state);


--
-- TOC entry 3342 (class 1259 OID 16482)
-- Name: idx_executions_task_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_executions_task_id ON public.executions USING btree (task_id);


--
-- TOC entry 3343 (class 1259 OID 16479)
-- Name: idx_executions_user; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_executions_user ON public.executions USING btree ("user");


--
-- TOC entry 3344 (class 1259 OID 16477)
-- Name: idx_executions_workflow_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_executions_workflow_id ON public.executions USING btree (workflow_id);


--
-- TOC entry 3328 (class 1259 OID 16462)
-- Name: idx_launch_plans_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_launch_plans_deleted_at ON public.launch_plans USING btree (deleted_at);


--
-- TOC entry 3329 (class 1259 OID 16621)
-- Name: idx_launch_plans_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_launch_plans_id ON public.launch_plans USING btree (id);


--
-- TOC entry 3330 (class 1259 OID 16460)
-- Name: idx_launch_plans_workflow_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_launch_plans_workflow_id ON public.launch_plans USING btree (workflow_id);


--
-- TOC entry 3380 (class 1259 OID 16581)
-- Name: idx_named_entity_metadata_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_named_entity_metadata_deleted_at ON public.named_entity_metadata USING btree (deleted_at);


--
-- TOC entry 3381 (class 1259 OID 16623)
-- Name: idx_named_entity_metadata_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_named_entity_metadata_id ON public.named_entity_metadata USING btree (id);


--
-- TOC entry 3357 (class 1259 OID 16524)
-- Name: idx_node_execution_events_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_node_execution_events_deleted_at ON public.node_execution_events USING btree (deleted_at);


--
-- TOC entry 3358 (class 1259 OID 16625)
-- Name: idx_node_execution_events_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_node_execution_events_id ON public.node_execution_events USING btree (id);


--
-- TOC entry 3359 (class 1259 OID 16521)
-- Name: idx_node_execution_events_node_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_node_execution_events_node_id ON public.node_execution_events USING btree (node_id);


--
-- TOC entry 3349 (class 1259 OID 16505)
-- Name: idx_node_executions_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_node_executions_deleted_at ON public.node_executions USING btree (deleted_at);


--
-- TOC entry 3350 (class 1259 OID 16593)
-- Name: idx_node_executions_error_kind; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_node_executions_error_kind ON public.node_executions USING btree (error_kind);


--
-- TOC entry 3351 (class 1259 OID 16627)
-- Name: idx_node_executions_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_node_executions_id ON public.node_executions USING btree (id);


--
-- TOC entry 3352 (class 1259 OID 16504)
-- Name: idx_node_executions_node_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_node_executions_node_id ON public.node_executions USING btree (node_id);


--
-- TOC entry 3353 (class 1259 OID 16592)
-- Name: idx_node_executions_parent_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_node_executions_parent_id ON public.node_executions USING btree (parent_id);


--
-- TOC entry 3354 (class 1259 OID 16503)
-- Name: idx_node_executions_parent_task_execution_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_node_executions_parent_task_execution_id ON public.node_executions USING btree (parent_task_execution_id);


--
-- TOC entry 3375 (class 1259 OID 16561)
-- Name: idx_partitions_artifact_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_partitions_artifact_id ON public.partitions USING btree (artifact_id);


--
-- TOC entry 3308 (class 1259 OID 16414)
-- Name: idx_projects_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_projects_deleted_at ON public.projects USING btree (deleted_at);


--
-- TOC entry 3309 (class 1259 OID 16629)
-- Name: idx_projects_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_projects_id ON public.projects USING btree (id);


--
-- TOC entry 3310 (class 1259 OID 16413)
-- Name: idx_projects_state; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_projects_state ON public.projects USING btree (state);


--
-- TOC entry 3388 (class 1259 OID 16603)
-- Name: idx_schedulable_entities_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_schedulable_entities_deleted_at ON public.schedulable_entities USING btree (deleted_at);


--
-- TOC entry 3389 (class 1259 OID 16634)
-- Name: idx_schedulable_entities_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_schedulable_entities_id ON public.schedulable_entities USING btree (id);


--
-- TOC entry 3392 (class 1259 OID 16614)
-- Name: idx_schedule_entities_snapshots_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_schedule_entities_snapshots_deleted_at ON public.schedule_entities_snapshots USING btree (deleted_at);


--
-- TOC entry 3393 (class 1259 OID 16638)
-- Name: idx_schedule_entities_snapshots_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_schedule_entities_snapshots_id ON public.schedule_entities_snapshots USING btree (id);


--
-- TOC entry 3401 (class 1259 OID 16667)
-- Name: idx_signals_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_signals_deleted_at ON public.signals USING btree (deleted_at);


--
-- TOC entry 3402 (class 1259 OID 16668)
-- Name: idx_signals_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_signals_id ON public.signals USING btree (id);


--
-- TOC entry 3403 (class 1259 OID 16666)
-- Name: idx_signals_signal_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_signals_signal_id ON public.signals USING btree (signal_id);


--
-- TOC entry 3365 (class 1259 OID 16538)
-- Name: idx_task_executions_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_task_executions_deleted_at ON public.task_executions USING btree (deleted_at);


--
-- TOC entry 3366 (class 1259 OID 16537)
-- Name: idx_task_executions_exec; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_task_executions_exec ON public.task_executions USING btree (execution_project, execution_domain, execution_name, node_id);


--
-- TOC entry 3367 (class 1259 OID 16640)
-- Name: idx_task_executions_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_task_executions_id ON public.task_executions USING btree (id);


--
-- TOC entry 3368 (class 1259 OID 16536)
-- Name: idx_task_executions_node_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_task_executions_node_id ON public.task_executions USING btree (node_id);


--
-- TOC entry 3316 (class 1259 OID 16434)
-- Name: idx_tasks_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_tasks_deleted_at ON public.tasks USING btree (deleted_at);


--
-- TOC entry 3317 (class 1259 OID 16642)
-- Name: idx_tasks_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_tasks_id ON public.tasks USING btree (id);


--
-- TOC entry 3322 (class 1259 OID 16448)
-- Name: idx_workflows_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_workflows_deleted_at ON public.workflows USING btree (deleted_at);


--
-- TOC entry 3323 (class 1259 OID 16644)
-- Name: idx_workflows_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_workflows_id ON public.workflows USING btree (id);


--
-- TOC entry 3333 (class 1259 OID 16461)
-- Name: lp_project_domain_name_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX lp_project_domain_name_idx ON public.launch_plans USING btree (project, domain, name);


--
-- TOC entry 3384 (class 1259 OID 16580)
-- Name: named_entity_metadata_type_project_domain_name_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX named_entity_metadata_type_project_domain_name_idx ON public.named_entity_metadata USING btree (resource_type, project, domain, name);


--
-- TOC entry 3385 (class 1259 OID 16591)
-- Name: resource_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX resource_idx ON public.resources USING btree (project, domain, workflow, launch_plan, resource_type);


--
-- TOC entry 3313 (class 1259 OID 16424)
-- Name: tags_dataset_uuid_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX tags_dataset_uuid_idx ON public.tags USING btree (dataset_uuid);


--
-- TOC entry 3318 (class 1259 OID 16436)
-- Name: task_project_domain_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX task_project_domain_idx ON public.tasks USING btree (project, domain);


--
-- TOC entry 3319 (class 1259 OID 16433)
-- Name: task_project_domain_name_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX task_project_domain_name_idx ON public.tasks USING btree (project, domain, name);


--
-- TOC entry 3324 (class 1259 OID 16446)
-- Name: workflow_project_domain_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX workflow_project_domain_idx ON public.workflows USING btree (project, domain);


--
-- TOC entry 3325 (class 1259 OID 16447)
-- Name: workflow_project_domain_name_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX workflow_project_domain_name_idx ON public.workflows USING btree (project, domain, name);


-- Completed on 2023-03-20 14:15:25 PDT

--
-- PostgreSQL database dump complete
--

