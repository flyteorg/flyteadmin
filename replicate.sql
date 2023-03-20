--
-- PostgreSQL database dump
--

-- Dumped from database version 15.1
-- Dumped by pg_dump version 15.1

-- Started on 2023-03-20 14:34:36 PDT

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
-- TOC entry 234 (class 1259 OID 16800)
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
    link text,
    short_description text,
    long_description bytea
);


ALTER TABLE public.description_entities OWNER TO postgres;

--
-- TOC entry 233 (class 1259 OID 16799)
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
-- TOC entry 3481 (class 0 OID 0)
-- Dependencies: 233
-- Name: description_entities_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.description_entities_id_seq OWNED BY public.description_entities.id;


--
-- TOC entry 230 (class 1259 OID 16777)
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
-- TOC entry 229 (class 1259 OID 16776)
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
-- TOC entry 3482 (class 0 OID 0)
-- Dependencies: 229
-- Name: execution_events_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.execution_events_id_seq OWNED BY public.execution_events.id;


--
-- TOC entry 240 (class 1259 OID 16836)
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
-- TOC entry 239 (class 1259 OID 16835)
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
-- TOC entry 3483 (class 0 OID 0)
-- Dependencies: 239
-- Name: executions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.executions_id_seq OWNED BY public.executions.id;


--
-- TOC entry 222 (class 1259 OID 16724)
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
-- TOC entry 221 (class 1259 OID 16723)
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
-- TOC entry 3484 (class 0 OID 0)
-- Dependencies: 221
-- Name: launch_plans_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.launch_plans_id_seq OWNED BY public.launch_plans.id;


--
-- TOC entry 214 (class 1259 OID 16679)
-- Name: migrations; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.migrations (
    id character varying(255) NOT NULL
);


ALTER TABLE public.migrations OWNER TO postgres;

--
-- TOC entry 224 (class 1259 OID 16738)
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
-- TOC entry 223 (class 1259 OID 16737)
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
-- TOC entry 3485 (class 0 OID 0)
-- Dependencies: 223
-- Name: named_entity_metadata_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.named_entity_metadata_id_seq OWNED BY public.named_entity_metadata.id;


--
-- TOC entry 232 (class 1259 OID 16788)
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
-- TOC entry 231 (class 1259 OID 16787)
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
-- TOC entry 3486 (class 0 OID 0)
-- Dependencies: 231
-- Name: node_execution_events_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.node_execution_events_id_seq OWNED BY public.node_execution_events.id;


--
-- TOC entry 228 (class 1259 OID 16764)
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
    internal_data bytea
);


ALTER TABLE public.node_executions OWNER TO postgres;

--
-- TOC entry 227 (class 1259 OID 16763)
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
-- TOC entry 3487 (class 0 OID 0)
-- Dependencies: 227
-- Name: node_executions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.node_executions_id_seq OWNED BY public.node_executions.id;


--
-- TOC entry 216 (class 1259 OID 16685)
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
-- TOC entry 215 (class 1259 OID 16684)
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
-- TOC entry 3488 (class 0 OID 0)
-- Dependencies: 215
-- Name: projects_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.projects_id_seq OWNED BY public.projects.id;


--
-- TOC entry 238 (class 1259 OID 16824)
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
-- TOC entry 237 (class 1259 OID 16823)
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
-- TOC entry 3489 (class 0 OID 0)
-- Dependencies: 237
-- Name: resources_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.resources_id_seq OWNED BY public.resources.id;


--
-- TOC entry 236 (class 1259 OID 16812)
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
-- TOC entry 235 (class 1259 OID 16811)
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
-- TOC entry 3490 (class 0 OID 0)
-- Dependencies: 235
-- Name: signals_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.signals_id_seq OWNED BY public.signals.id;


--
-- TOC entry 226 (class 1259 OID 16751)
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
-- TOC entry 225 (class 1259 OID 16750)
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
-- TOC entry 3491 (class 0 OID 0)
-- Dependencies: 225
-- Name: task_executions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.task_executions_id_seq OWNED BY public.task_executions.id;


--
-- TOC entry 218 (class 1259 OID 16698)
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
-- TOC entry 217 (class 1259 OID 16697)
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
-- TOC entry 3492 (class 0 OID 0)
-- Dependencies: 217
-- Name: tasks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.tasks_id_seq OWNED BY public.tasks.id;


--
-- TOC entry 220 (class 1259 OID 16711)
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
-- TOC entry 219 (class 1259 OID 16710)
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
-- TOC entry 3493 (class 0 OID 0)
-- Dependencies: 219
-- Name: workflows_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.workflows_id_seq OWNED BY public.workflows.id;


--
-- TOC entry 3254 (class 2604 OID 16803)
-- Name: description_entities id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.description_entities ALTER COLUMN id SET DEFAULT nextval('public.description_entities_id_seq'::regclass);


--
-- TOC entry 3252 (class 2604 OID 16780)
-- Name: execution_events id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.execution_events ALTER COLUMN id SET DEFAULT nextval('public.execution_events_id_seq'::regclass);


--
-- TOC entry 3257 (class 2604 OID 16839)
-- Name: executions id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.executions ALTER COLUMN id SET DEFAULT nextval('public.executions_id_seq'::regclass);


--
-- TOC entry 3246 (class 2604 OID 16727)
-- Name: launch_plans id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.launch_plans ALTER COLUMN id SET DEFAULT nextval('public.launch_plans_id_seq'::regclass);


--
-- TOC entry 3248 (class 2604 OID 16741)
-- Name: named_entity_metadata id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.named_entity_metadata ALTER COLUMN id SET DEFAULT nextval('public.named_entity_metadata_id_seq'::regclass);


--
-- TOC entry 3253 (class 2604 OID 16791)
-- Name: node_execution_events id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.node_execution_events ALTER COLUMN id SET DEFAULT nextval('public.node_execution_events_id_seq'::regclass);


--
-- TOC entry 3251 (class 2604 OID 16767)
-- Name: node_executions id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.node_executions ALTER COLUMN id SET DEFAULT nextval('public.node_executions_id_seq'::regclass);


--
-- TOC entry 3242 (class 2604 OID 16688)
-- Name: projects id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.projects ALTER COLUMN id SET DEFAULT nextval('public.projects_id_seq'::regclass);


--
-- TOC entry 3256 (class 2604 OID 16827)
-- Name: resources id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.resources ALTER COLUMN id SET DEFAULT nextval('public.resources_id_seq'::regclass);


--
-- TOC entry 3255 (class 2604 OID 16815)
-- Name: signals id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.signals ALTER COLUMN id SET DEFAULT nextval('public.signals_id_seq'::regclass);


--
-- TOC entry 3250 (class 2604 OID 16754)
-- Name: task_executions id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.task_executions ALTER COLUMN id SET DEFAULT nextval('public.task_executions_id_seq'::regclass);


--
-- TOC entry 3244 (class 2604 OID 16701)
-- Name: tasks id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tasks ALTER COLUMN id SET DEFAULT nextval('public.tasks_id_seq'::regclass);


--
-- TOC entry 3245 (class 2604 OID 16714)
-- Name: workflows id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.workflows ALTER COLUMN id SET DEFAULT nextval('public.workflows_id_seq'::regclass);


--
-- TOC entry 3311 (class 2606 OID 16807)
-- Name: description_entities description_entities_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.description_entities
    ADD CONSTRAINT description_entities_pkey PRIMARY KEY (resource_type, project, domain, name, version);


--
-- TOC entry 3302 (class 2606 OID 16784)
-- Name: execution_events execution_events_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.execution_events
    ADD CONSTRAINT execution_events_pkey PRIMARY KEY (execution_project, execution_domain, execution_name, phase);


--
-- TOC entry 3324 (class 2606 OID 16844)
-- Name: executions executions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.executions
    ADD CONSTRAINT executions_pkey PRIMARY KEY (execution_project, execution_domain, execution_name);


--
-- TOC entry 3282 (class 2606 OID 16732)
-- Name: launch_plans launch_plans_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.launch_plans
    ADD CONSTRAINT launch_plans_pkey PRIMARY KEY (project, domain, name, version);


--
-- TOC entry 3260 (class 2606 OID 16683)
-- Name: migrations migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.migrations
    ADD CONSTRAINT migrations_pkey PRIMARY KEY (id);


--
-- TOC entry 3287 (class 2606 OID 16746)
-- Name: named_entity_metadata named_entity_metadata_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.named_entity_metadata
    ADD CONSTRAINT named_entity_metadata_pkey PRIMARY KEY (resource_type, project, domain, name);


--
-- TOC entry 3309 (class 2606 OID 16795)
-- Name: node_execution_events node_execution_events_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.node_execution_events
    ADD CONSTRAINT node_execution_events_pkey PRIMARY KEY (execution_project, execution_domain, execution_name, node_id, phase);


--
-- TOC entry 3300 (class 2606 OID 16771)
-- Name: node_executions node_executions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.node_executions
    ADD CONSTRAINT node_executions_pkey PRIMARY KEY (execution_project, execution_domain, execution_name, node_id);


--
-- TOC entry 3265 (class 2606 OID 16693)
-- Name: projects projects_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_pkey PRIMARY KEY (identifier);


--
-- TOC entry 3322 (class 2606 OID 16831)
-- Name: resources resources_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.resources
    ADD CONSTRAINT resources_pkey PRIMARY KEY (id);


--
-- TOC entry 3319 (class 2606 OID 16819)
-- Name: signals signals_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.signals
    ADD CONSTRAINT signals_pkey PRIMARY KEY (execution_project, execution_domain, execution_name, signal_id);


--
-- TOC entry 3294 (class 2606 OID 16758)
-- Name: task_executions task_executions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.task_executions
    ADD CONSTRAINT task_executions_pkey PRIMARY KEY (project, domain, name, version, execution_project, execution_domain, execution_name, node_id, retry_attempt);


--
-- TOC entry 3271 (class 2606 OID 16705)
-- Name: tasks tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_pkey PRIMARY KEY (project, domain, name, version);


--
-- TOC entry 3277 (class 2606 OID 16718)
-- Name: workflows workflows_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.workflows
    ADD CONSTRAINT workflows_pkey PRIMARY KEY (project, domain, name, version);


--
-- TOC entry 3312 (class 1259 OID 16808)
-- Name: description_entity_project_domain_name_version_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX description_entity_project_domain_name_version_idx ON public.description_entities USING btree (resource_type, project, domain, name, version);


--
-- TOC entry 3313 (class 1259 OID 16809)
-- Name: idx_description_entities_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_description_entities_deleted_at ON public.description_entities USING btree (deleted_at);


--
-- TOC entry 3314 (class 1259 OID 16810)
-- Name: idx_description_entities_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_description_entities_id ON public.description_entities USING btree (id);


--
-- TOC entry 3303 (class 1259 OID 16785)
-- Name: idx_execution_events_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_execution_events_deleted_at ON public.execution_events USING btree (deleted_at);


--
-- TOC entry 3304 (class 1259 OID 16786)
-- Name: idx_execution_events_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_execution_events_id ON public.execution_events USING btree (id);


--
-- TOC entry 3325 (class 1259 OID 16845)
-- Name: idx_executions_created_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_executions_created_at ON public.executions USING btree (execution_created_at);


--
-- TOC entry 3326 (class 1259 OID 16853)
-- Name: idx_executions_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_executions_deleted_at ON public.executions USING btree (deleted_at);


--
-- TOC entry 3327 (class 1259 OID 16850)
-- Name: idx_executions_error_kind; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_executions_error_kind ON public.executions USING btree (error_kind);


--
-- TOC entry 3328 (class 1259 OID 16847)
-- Name: idx_executions_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_executions_id ON public.executions USING btree (id);


--
-- TOC entry 3329 (class 1259 OID 16846)
-- Name: idx_executions_launch_plan_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_executions_launch_plan_id ON public.executions USING btree (launch_plan_id);


--
-- TOC entry 3330 (class 1259 OID 16848)
-- Name: idx_executions_state; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_executions_state ON public.executions USING btree (state);


--
-- TOC entry 3331 (class 1259 OID 16851)
-- Name: idx_executions_task_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_executions_task_id ON public.executions USING btree (task_id);


--
-- TOC entry 3332 (class 1259 OID 16849)
-- Name: idx_executions_user; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_executions_user ON public.executions USING btree ("user");


--
-- TOC entry 3333 (class 1259 OID 16852)
-- Name: idx_executions_workflow_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_executions_workflow_id ON public.executions USING btree (workflow_id);


--
-- TOC entry 3278 (class 1259 OID 16735)
-- Name: idx_launch_plans_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_launch_plans_deleted_at ON public.launch_plans USING btree (deleted_at);


--
-- TOC entry 3279 (class 1259 OID 16736)
-- Name: idx_launch_plans_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_launch_plans_id ON public.launch_plans USING btree (id);


--
-- TOC entry 3280 (class 1259 OID 16733)
-- Name: idx_launch_plans_workflow_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_launch_plans_workflow_id ON public.launch_plans USING btree (workflow_id);


--
-- TOC entry 3284 (class 1259 OID 16748)
-- Name: idx_named_entity_metadata_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_named_entity_metadata_deleted_at ON public.named_entity_metadata USING btree (deleted_at);


--
-- TOC entry 3285 (class 1259 OID 16749)
-- Name: idx_named_entity_metadata_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_named_entity_metadata_id ON public.named_entity_metadata USING btree (id);


--
-- TOC entry 3305 (class 1259 OID 16797)
-- Name: idx_node_execution_events_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_node_execution_events_deleted_at ON public.node_execution_events USING btree (deleted_at);


--
-- TOC entry 3306 (class 1259 OID 16798)
-- Name: idx_node_execution_events_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_node_execution_events_id ON public.node_execution_events USING btree (id);


--
-- TOC entry 3307 (class 1259 OID 16796)
-- Name: idx_node_execution_events_node_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_node_execution_events_node_id ON public.node_execution_events USING btree (node_id);


--
-- TOC entry 3295 (class 1259 OID 16772)
-- Name: idx_node_executions_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_node_executions_deleted_at ON public.node_executions USING btree (deleted_at);


--
-- TOC entry 3296 (class 1259 OID 16773)
-- Name: idx_node_executions_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_node_executions_id ON public.node_executions USING btree (id);


--
-- TOC entry 3297 (class 1259 OID 16775)
-- Name: idx_node_executions_node_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_node_executions_node_id ON public.node_executions USING btree (node_id);


--
-- TOC entry 3298 (class 1259 OID 16774)
-- Name: idx_node_executions_parent_task_execution_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_node_executions_parent_task_execution_id ON public.node_executions USING btree (parent_task_execution_id);


--
-- TOC entry 3261 (class 1259 OID 16695)
-- Name: idx_projects_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_projects_deleted_at ON public.projects USING btree (deleted_at);


--
-- TOC entry 3262 (class 1259 OID 16696)
-- Name: idx_projects_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_projects_id ON public.projects USING btree (id);


--
-- TOC entry 3263 (class 1259 OID 16694)
-- Name: idx_projects_state; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_projects_state ON public.projects USING btree (state);


--
-- TOC entry 3315 (class 1259 OID 16821)
-- Name: idx_signals_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_signals_deleted_at ON public.signals USING btree (deleted_at);


--
-- TOC entry 3316 (class 1259 OID 16822)
-- Name: idx_signals_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_signals_id ON public.signals USING btree (id);


--
-- TOC entry 3317 (class 1259 OID 16820)
-- Name: idx_signals_signal_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_signals_signal_id ON public.signals USING btree (signal_id);


--
-- TOC entry 3289 (class 1259 OID 16761)
-- Name: idx_task_executions_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_task_executions_deleted_at ON public.task_executions USING btree (deleted_at);


--
-- TOC entry 3290 (class 1259 OID 16760)
-- Name: idx_task_executions_exec; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_task_executions_exec ON public.task_executions USING btree (execution_project, execution_domain, execution_name, node_id);


--
-- TOC entry 3291 (class 1259 OID 16762)
-- Name: idx_task_executions_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_task_executions_id ON public.task_executions USING btree (id);


--
-- TOC entry 3292 (class 1259 OID 16759)
-- Name: idx_task_executions_node_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_task_executions_node_id ON public.task_executions USING btree (node_id);


--
-- TOC entry 3266 (class 1259 OID 16708)
-- Name: idx_tasks_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_tasks_deleted_at ON public.tasks USING btree (deleted_at);


--
-- TOC entry 3267 (class 1259 OID 16709)
-- Name: idx_tasks_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_tasks_id ON public.tasks USING btree (id);


--
-- TOC entry 3272 (class 1259 OID 16721)
-- Name: idx_workflows_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_workflows_deleted_at ON public.workflows USING btree (deleted_at);


--
-- TOC entry 3273 (class 1259 OID 16722)
-- Name: idx_workflows_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_workflows_id ON public.workflows USING btree (id);


--
-- TOC entry 3283 (class 1259 OID 16734)
-- Name: lp_project_domain_name_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX lp_project_domain_name_idx ON public.launch_plans USING btree (project, domain, name);


--
-- TOC entry 3288 (class 1259 OID 16747)
-- Name: named_entity_metadata_type_project_domain_name_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX named_entity_metadata_type_project_domain_name_idx ON public.named_entity_metadata USING btree (resource_type, project, domain, name);


--
-- TOC entry 3320 (class 1259 OID 16832)
-- Name: resource_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX resource_idx ON public.resources USING btree (project, domain, workflow, launch_plan, resource_type);


--
-- TOC entry 3268 (class 1259 OID 16706)
-- Name: task_project_domain_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX task_project_domain_idx ON public.tasks USING btree (project, domain);


--
-- TOC entry 3269 (class 1259 OID 16707)
-- Name: task_project_domain_name_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX task_project_domain_name_idx ON public.tasks USING btree (project, domain, name);


--
-- TOC entry 3274 (class 1259 OID 16719)
-- Name: workflow_project_domain_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX workflow_project_domain_idx ON public.workflows USING btree (project, domain);


--
-- TOC entry 3275 (class 1259 OID 16720)
-- Name: workflow_project_domain_name_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX workflow_project_domain_name_idx ON public.workflows USING btree (project, domain, name);


-- Completed on 2023-03-20 14:34:39 PDT

--
-- PostgreSQL database dump complete
--

