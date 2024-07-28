
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

CREATE EXTENSION IF NOT EXISTS "pgsodium" WITH SCHEMA "pgsodium";

COMMENT ON SCHEMA "public" IS 'standard public schema';

CREATE EXTENSION IF NOT EXISTS "pg_graphql" WITH SCHEMA "graphql";

CREATE EXTENSION IF NOT EXISTS "pg_stat_statements" WITH SCHEMA "extensions";

CREATE EXTENSION IF NOT EXISTS "pgcrypto" WITH SCHEMA "extensions";

CREATE EXTENSION IF NOT EXISTS "pgjwt" WITH SCHEMA "extensions";

CREATE EXTENSION IF NOT EXISTS "supabase_vault" WITH SCHEMA "vault";

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA "extensions";

SET default_tablespace = '';

SET default_table_access_method = "heap";

CREATE TABLE IF NOT EXISTS "public"."commutes" (
    "id" integer NOT NULL,
    "user_id" integer NOT NULL,
    "query_time" timestamp without time zone NOT NULL,
    "duration" bigint NOT NULL,
    "route_hash" "text",
    "route" integer NOT NULL,
    "distance" bigint NOT NULL
);

ALTER TABLE "public"."commutes" OWNER TO "postgres";

CREATE SEQUENCE IF NOT EXISTS "public"."commutes_id_seq"
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE "public"."commutes_id_seq" OWNER TO "postgres";

ALTER SEQUENCE "public"."commutes_id_seq" OWNED BY "public"."commutes"."id";

CREATE TABLE IF NOT EXISTS "public"."routes" (
    "id" integer NOT NULL,
    "user_id" integer NOT NULL,
    "start_address" "text",
    "end_address" "text",
    "start_latitude" "text" NOT NULL,
    "end_latitude" "text" NOT NULL,
    "active" boolean NOT NULL,
    "start_date" "date" NOT NULL,
    "end_date" "date",
    "start_longitude" "text" NOT NULL,
    "end_longitude" "text" NOT NULL
);

ALTER TABLE "public"."routes" OWNER TO "postgres";

CREATE SEQUENCE IF NOT EXISTS "public"."routes_id_seq"
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE "public"."routes_id_seq" OWNER TO "postgres";

ALTER SEQUENCE "public"."routes_id_seq" OWNED BY "public"."routes"."id";

CREATE TABLE IF NOT EXISTS "public"."users" (
    "id" integer NOT NULL,
    "username" "text" NOT NULL,
    "date_joined" "date" NOT NULL,
    "name" "text"
);

ALTER TABLE "public"."users" OWNER TO "postgres";

CREATE SEQUENCE IF NOT EXISTS "public"."users_id_seq"
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE "public"."users_id_seq" OWNER TO "postgres";

ALTER SEQUENCE "public"."users_id_seq" OWNED BY "public"."users"."id";

ALTER TABLE ONLY "public"."commutes" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."commutes_id_seq"'::"regclass");

ALTER TABLE ONLY "public"."routes" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."routes_id_seq"'::"regclass");

ALTER TABLE ONLY "public"."users" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."users_id_seq"'::"regclass");

ALTER TABLE ONLY "public"."commutes"
    ADD CONSTRAINT "commutes_pkey" PRIMARY KEY ("id");

ALTER TABLE ONLY "public"."routes"
    ADD CONSTRAINT "routes_pkey" PRIMARY KEY ("id");

ALTER TABLE ONLY "public"."users"
    ADD CONSTRAINT "users_pkey" PRIMARY KEY ("id");

ALTER TABLE ONLY "public"."commutes"
    ADD CONSTRAINT "fk_route" FOREIGN KEY ("route") REFERENCES "public"."routes"("id");

ALTER TABLE ONLY "public"."routes"
    ADD CONSTRAINT "fk_user" FOREIGN KEY ("user_id") REFERENCES "public"."users"("id");

ALTER TABLE ONLY "public"."commutes"
    ADD CONSTRAINT "fk_user" FOREIGN KEY ("user_id") REFERENCES "public"."users"("id");

ALTER TABLE "public"."commutes" ENABLE ROW LEVEL SECURITY;

ALTER TABLE "public"."routes" ENABLE ROW LEVEL SECURITY;

ALTER TABLE "public"."users" ENABLE ROW LEVEL SECURITY;

ALTER PUBLICATION "supabase_realtime" OWNER TO "postgres";

GRANT USAGE ON SCHEMA "public" TO "postgres";
GRANT USAGE ON SCHEMA "public" TO "anon";
GRANT USAGE ON SCHEMA "public" TO "authenticated";
GRANT USAGE ON SCHEMA "public" TO "service_role";

GRANT ALL ON TABLE "public"."commutes" TO "anon";
GRANT ALL ON TABLE "public"."commutes" TO "authenticated";
GRANT ALL ON TABLE "public"."commutes" TO "service_role";

GRANT ALL ON SEQUENCE "public"."commutes_id_seq" TO "anon";
GRANT ALL ON SEQUENCE "public"."commutes_id_seq" TO "authenticated";
GRANT ALL ON SEQUENCE "public"."commutes_id_seq" TO "service_role";

GRANT ALL ON TABLE "public"."routes" TO "anon";
GRANT ALL ON TABLE "public"."routes" TO "authenticated";
GRANT ALL ON TABLE "public"."routes" TO "service_role";

GRANT ALL ON SEQUENCE "public"."routes_id_seq" TO "anon";
GRANT ALL ON SEQUENCE "public"."routes_id_seq" TO "authenticated";
GRANT ALL ON SEQUENCE "public"."routes_id_seq" TO "service_role";

GRANT ALL ON TABLE "public"."users" TO "anon";
GRANT ALL ON TABLE "public"."users" TO "authenticated";
GRANT ALL ON TABLE "public"."users" TO "service_role";

GRANT ALL ON SEQUENCE "public"."users_id_seq" TO "anon";
GRANT ALL ON SEQUENCE "public"."users_id_seq" TO "authenticated";
GRANT ALL ON SEQUENCE "public"."users_id_seq" TO "service_role";

ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON SEQUENCES  TO "postgres";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON SEQUENCES  TO "anon";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON SEQUENCES  TO "authenticated";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON SEQUENCES  TO "service_role";

ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON FUNCTIONS  TO "postgres";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON FUNCTIONS  TO "anon";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON FUNCTIONS  TO "authenticated";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON FUNCTIONS  TO "service_role";

ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON TABLES  TO "postgres";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON TABLES  TO "anon";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON TABLES  TO "authenticated";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON TABLES  TO "service_role";

RESET ALL;
