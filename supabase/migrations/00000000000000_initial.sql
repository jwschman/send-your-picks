


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


COMMENT ON SCHEMA "public" IS 'standard public schema';



CREATE EXTENSION IF NOT EXISTS "pg_graphql" WITH SCHEMA "graphql";






CREATE EXTENSION IF NOT EXISTS "pg_stat_statements" WITH SCHEMA "extensions";






CREATE EXTENSION IF NOT EXISTS "pgcrypto" WITH SCHEMA "extensions";






CREATE EXTENSION IF NOT EXISTS "supabase_vault" WITH SCHEMA "vault";






CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA "extensions";






CREATE TYPE "public"."game_status" AS ENUM (
    'scheduled',
    'in_progress',
    'final'
);


ALTER TYPE "public"."game_status" OWNER TO "postgres";


CREATE TYPE "public"."user_role" AS ENUM (
    'user',
    'commissioner',
    'admin'
);


ALTER TYPE "public"."user_role" OWNER TO "postgres";


CREATE TYPE "public"."week_status" AS ENUM (
    'draft',
    'active',
    'played',
    'final',
    'games_imported',
    'spreads_set',
    'picks_results_calculated',
    'scored'
);


ALTER TYPE "public"."week_status" OWNER TO "postgres";


CREATE OR REPLACE FUNCTION "public"."custom_access_token_hook"("event" "jsonb") RETURNS "jsonb"
    LANGUAGE "plpgsql" STABLE SECURITY DEFINER
    SET "search_path" TO ''
    AS $$
DECLARE
  claims jsonb;
  user_role text;
BEGIN
  -- Fetch the user role from your profiles table
  SELECT role::text INTO user_role
  FROM public.profiles
  WHERE id = (event->>'user_id')::uuid;

  claims := event->'claims';

  -- Add user_role to the JWT claims
  IF user_role IS NOT NULL THEN
    claims := jsonb_set(claims, '{user_role}', to_jsonb(user_role));
  ELSE
    -- Default to 'user' if no profile exists yet
    claims := jsonb_set(claims, '{user_role}', to_jsonb('user'));
  END IF;

  -- Update the 'claims' object in the event
  event := jsonb_set(event, '{claims}', claims);

  RETURN event;
END;
$$;


ALTER FUNCTION "public"."custom_access_token_hook"("event" "jsonb") OWNER TO "postgres";


CREATE OR REPLACE FUNCTION "public"."handle_new_user"() RETURNS "trigger"
    LANGUAGE "plpgsql" SECURITY DEFINER
    SET "search_path" TO ''
    AS $$
BEGIN
  INSERT INTO public.profiles (id, username, role)
  VALUES (NEW.id, NEW.email, 'user');
  RETURN NEW;
END;
$$;


ALTER FUNCTION "public"."handle_new_user"() OWNER TO "postgres";


CREATE OR REPLACE FUNCTION "public"."update_updated_at_column"() RETURNS "trigger"
    LANGUAGE "plpgsql"
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;


ALTER FUNCTION "public"."update_updated_at_column"() OWNER TO "postgres";

SET default_tablespace = '';

SET default_table_access_method = "heap";


CREATE TABLE IF NOT EXISTS "public"."games" (
    "id" "text" NOT NULL,
    "season_id" "text" NOT NULL,
    "week_id" "text" NOT NULL,
    "home_team_id" "text" NOT NULL,
    "away_team_id" "text" NOT NULL,
    "home_score" integer,
    "away_score" integer,
    "home_spread" numeric(4,1),
    "kickoff_time" timestamp with time zone NOT NULL,
    "neutral_site" boolean DEFAULT false NOT NULL,
    "created_at" timestamp with time zone DEFAULT "now"(),
    "updated_at" timestamp with time zone DEFAULT "now"(),
    "created_by" "uuid" NOT NULL,
    "external_game_id" bigint,
    "home_team_abbr" "text" NOT NULL,
    "away_team_abbr" "text" NOT NULL,
    "status" "public"."game_status" DEFAULT 'scheduled'::"public"."game_status" NOT NULL,
    CONSTRAINT "games_check" CHECK (("home_team_id" <> "away_team_id"))
);


ALTER TABLE "public"."games" OWNER TO "postgres";


COMMENT ON TABLE "public"."games" IS 'Individual games in a week';



COMMENT ON COLUMN "public"."games"."home_spread" IS 'Point spread for home team (negative means home team favored)';



CREATE TABLE IF NOT EXISTS "public"."picks" (
    "id" "text" NOT NULL,
    "user_id" "uuid" NOT NULL,
    "game_id" "text" NOT NULL,
    "week_id" "text" NOT NULL,
    "selected_team_id" "text",
    "created_at" timestamp with time zone DEFAULT "now"(),
    "updated_at" timestamp with time zone DEFAULT "now"(),
    "is_correct" boolean,
    "calculated_at" timestamp with time zone,
    "user_locked_at" timestamp with time zone
);


ALTER TABLE "public"."picks" OWNER TO "postgres";


COMMENT ON TABLE "public"."picks" IS 'User picks for games';



CREATE TABLE IF NOT EXISTS "public"."profiles" (
    "id" "uuid" NOT NULL,
    "role" "public"."user_role" DEFAULT 'user'::"public"."user_role" NOT NULL,
    "username" "text",
    "tagline" "text",
    "avatar_url" "text",
    "created_at" timestamp with time zone DEFAULT "now"(),
    "updated_at" timestamp with time zone DEFAULT "now"()
);


ALTER TABLE "public"."profiles" OWNER TO "postgres";


CREATE TABLE IF NOT EXISTS "public"."season_participants" (
    "season_id" "text" NOT NULL,
    "user_id" "uuid" NOT NULL,
    "joined_at" timestamp with time zone DEFAULT "now"() NOT NULL
);


ALTER TABLE "public"."season_participants" OWNER TO "postgres";


COMMENT ON TABLE "public"."season_participants" IS 'Tracks which users are participating in which seasons. Users must be explicitly added by a commissioner to submit picks and appear in standings.';



COMMENT ON COLUMN "public"."season_participants"."season_id" IS 'The season the user is participating in';



COMMENT ON COLUMN "public"."season_participants"."user_id" IS 'The user participating in the season';



COMMENT ON COLUMN "public"."season_participants"."joined_at" IS 'When the user was added to the season (useful for late-join tracking)';



CREATE TABLE IF NOT EXISTS "public"."season_standings" (
    "id" "text" NOT NULL,
    "user_id" "uuid" NOT NULL,
    "season_id" "text" NOT NULL,
    "week_id" "text" NOT NULL,
    "points" integer NOT NULL,
    "rank" integer,
    "computed_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "updated_at" timestamp with time zone DEFAULT "now"() NOT NULL
);


ALTER TABLE "public"."season_standings" OWNER TO "postgres";


COMMENT ON TABLE "public"."season_standings" IS 'Cumulative season standings per user as of the end of a given week. Derived from week_results and stored as immutable weekly snapshots.';



CREATE TABLE IF NOT EXISTS "public"."seasons" (
    "id" "text" NOT NULL,
    "year" integer NOT NULL,
    "is_active" boolean DEFAULT false NOT NULL,
    "created_at" timestamp with time zone DEFAULT "now"(),
    "updated_at" timestamp with time zone DEFAULT "now"(),
    "created_by" "uuid" NOT NULL,
    "number_of_weeks" integer DEFAULT 18 NOT NULL,
    "is_postseason" boolean DEFAULT false NOT NULL,
    CONSTRAINT "seasons_number_of_weeks_check" CHECK ((("number_of_weeks" > 0) AND ("number_of_weeks" <= 22)))
);


ALTER TABLE "public"."seasons" OWNER TO "postgres";


COMMENT ON TABLE "public"."seasons" IS 'Football seasons by year';



COMMENT ON COLUMN "public"."seasons"."is_active" IS 'Only one season can be active at a time';



CREATE TABLE IF NOT EXISTS "public"."settings" (
    "id" character(26) NOT NULL,
    "pick_cutoff_minutes" integer NOT NULL,
    "allow_pick_edits" boolean NOT NULL,
    "points_per_correct_pick" integer NOT NULL,
    "competition_timezone" "text" NOT NULL,
    "allow_commissioner_overrides" boolean NOT NULL,
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "updated_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "allow_picks_after_kickoff" boolean DEFAULT false NOT NULL,
    "debug_mode" boolean DEFAULT false NOT NULL
);


ALTER TABLE "public"."settings" OWNER TO "postgres";


COMMENT ON TABLE "public"."settings" IS 'Global application settings. This table is intended to contain exactly one row representing default configuration for the entire app.';



CREATE TABLE IF NOT EXISTS "public"."teams" (
    "id" "text" NOT NULL,
    "name" "text" NOT NULL,
    "abbreviation" "text" NOT NULL,
    "city" "text" NOT NULL,
    "is_active" boolean DEFAULT true NOT NULL,
    "logo_url" "text",
    "created_at" timestamp with time zone DEFAULT "now"(),
    "updated_at" timestamp with time zone DEFAULT "now"()
);


ALTER TABLE "public"."teams" OWNER TO "postgres";


COMMENT ON TABLE "public"."teams" IS 'NFL teams';



CREATE TABLE IF NOT EXISTS "public"."week_results" (
    "id" "text" NOT NULL,
    "user_id" "uuid" NOT NULL,
    "week_id" "text" NOT NULL,
    "points" integer DEFAULT 0 NOT NULL,
    "rank" integer,
    "computed_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "updated_at" timestamp with time zone DEFAULT "now"() NOT NULL
);


ALTER TABLE "public"."week_results" OWNER TO "postgres";


COMMENT ON TABLE "public"."week_results" IS 'Stores calculated points and rankings for each user per week';



CREATE TABLE IF NOT EXISTS "public"."weeks" (
    "id" "text" NOT NULL,
    "season_id" "text" NOT NULL,
    "number" integer NOT NULL,
    "created_at" timestamp with time zone DEFAULT "now"(),
    "updated_at" timestamp with time zone DEFAULT "now"(),
    "created_by" "uuid" NOT NULL,
    "activated_at" timestamp with time zone,
    "closed_at" timestamp with time zone,
    "status" "public"."week_status" DEFAULT 'draft'::"public"."week_status" NOT NULL,
    CONSTRAINT "weeks_number_check" CHECK ((("number" > 0) AND ("number" <= 22)))
);


ALTER TABLE "public"."weeks" OWNER TO "postgres";


COMMENT ON TABLE "public"."weeks" IS 'Weeks within a season';



ALTER TABLE ONLY "public"."games"
    ADD CONSTRAINT "games_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."picks"
    ADD CONSTRAINT "picks_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."picks"
    ADD CONSTRAINT "picks_user_id_game_id_key" UNIQUE ("user_id", "game_id");



ALTER TABLE ONLY "public"."profiles"
    ADD CONSTRAINT "profiles_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."season_participants"
    ADD CONSTRAINT "season_participants_pkey" PRIMARY KEY ("season_id", "user_id");



ALTER TABLE ONLY "public"."season_standings"
    ADD CONSTRAINT "season_standings_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."season_standings"
    ADD CONSTRAINT "season_standings_unique_user_week" UNIQUE ("season_id", "week_id", "user_id");



ALTER TABLE ONLY "public"."seasons"
    ADD CONSTRAINT "seasons_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."seasons"
    ADD CONSTRAINT "seasons_year_postseason_key" UNIQUE ("year", "is_postseason");



ALTER TABLE ONLY "public"."settings"
    ADD CONSTRAINT "settings_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."teams"
    ADD CONSTRAINT "teams_abbreviation_key" UNIQUE ("abbreviation");



ALTER TABLE ONLY "public"."teams"
    ADD CONSTRAINT "teams_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."week_results"
    ADD CONSTRAINT "week_results_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."week_results"
    ADD CONSTRAINT "week_results_user_id_week_id_key" UNIQUE ("user_id", "week_id");



ALTER TABLE ONLY "public"."weeks"
    ADD CONSTRAINT "weeks_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."weeks"
    ADD CONSTRAINT "weeks_season_id_number_key" UNIQUE ("season_id", "number");



CREATE UNIQUE INDEX "games_external_game_id_uniq" ON "public"."games" USING "btree" ("external_game_id") WHERE ("external_game_id" IS NOT NULL);



CREATE INDEX "idx_games_away_team_id" ON "public"."games" USING "btree" ("away_team_id");



CREATE INDEX "idx_games_home_team_id" ON "public"."games" USING "btree" ("home_team_id");



CREATE INDEX "idx_games_kickoff_time" ON "public"."games" USING "btree" ("kickoff_time");



CREATE INDEX "idx_games_season_id" ON "public"."games" USING "btree" ("season_id");



CREATE INDEX "idx_games_week_id" ON "public"."games" USING "btree" ("week_id");



CREATE INDEX "idx_games_week_id_kickoff_time" ON "public"."games" USING "btree" ("week_id", "kickoff_time");



CREATE INDEX "idx_games_week_id_status" ON "public"."games" USING "btree" ("week_id", "status");



CREATE INDEX "idx_picks_game_id" ON "public"."picks" USING "btree" ("game_id");



CREATE INDEX "idx_picks_selected_team_id" ON "public"."picks" USING "btree" ("selected_team_id");



CREATE INDEX "idx_picks_user_id" ON "public"."picks" USING "btree" ("user_id");



CREATE INDEX "idx_picks_week_id" ON "public"."picks" USING "btree" ("week_id");



CREATE INDEX "idx_season_participants_user_id" ON "public"."season_participants" USING "btree" ("user_id");



CREATE INDEX "idx_season_standings_computed_at" ON "public"."season_standings" USING "btree" ("computed_at");



CREATE INDEX "idx_season_standings_season_week_rank" ON "public"."season_standings" USING "btree" ("season_id", "week_id", "rank");



CREATE INDEX "idx_seasons_is_active" ON "public"."seasons" USING "btree" ("is_active");



CREATE INDEX "idx_seasons_year" ON "public"."seasons" USING "btree" ("year");



CREATE INDEX "idx_teams_abbreviation" ON "public"."teams" USING "btree" ("abbreviation");



CREATE INDEX "idx_teams_is_active" ON "public"."teams" USING "btree" ("is_active");



CREATE INDEX "idx_week_results_user_id" ON "public"."week_results" USING "btree" ("user_id");



CREATE INDEX "idx_week_results_week_id" ON "public"."week_results" USING "btree" ("week_id");



CREATE INDEX "idx_week_results_week_rank" ON "public"."week_results" USING "btree" ("week_id", "rank");



CREATE INDEX "idx_weeks_number" ON "public"."weeks" USING "btree" ("number");



CREATE INDEX "idx_weeks_season_id" ON "public"."weeks" USING "btree" ("season_id");



CREATE UNIQUE INDEX "settings_single_row_idx" ON "public"."settings" USING "btree" ((true));



CREATE UNIQUE INDEX "unique_active_season" ON "public"."seasons" USING "btree" ("is_active") WHERE ("is_active" = true);



CREATE UNIQUE INDEX "weeks_one_non_final_per_season" ON "public"."weeks" USING "btree" ("season_id") WHERE ("status" <> 'final'::"public"."week_status");



CREATE OR REPLACE TRIGGER "settings_updated_at" BEFORE UPDATE ON "public"."settings" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();



CREATE OR REPLACE TRIGGER "update_games_updated_at" BEFORE UPDATE ON "public"."games" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();



CREATE OR REPLACE TRIGGER "update_picks_updated_at" BEFORE UPDATE ON "public"."picks" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();



CREATE OR REPLACE TRIGGER "update_seasons_updated_at" BEFORE UPDATE ON "public"."seasons" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();



CREATE OR REPLACE TRIGGER "update_teams_updated_at" BEFORE UPDATE ON "public"."teams" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();



CREATE OR REPLACE TRIGGER "update_weeks_updated_at" BEFORE UPDATE ON "public"."weeks" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();



ALTER TABLE ONLY "public"."games"
    ADD CONSTRAINT "games_away_team_id_fkey" FOREIGN KEY ("away_team_id") REFERENCES "public"."teams"("id") ON DELETE RESTRICT;



ALTER TABLE ONLY "public"."games"
    ADD CONSTRAINT "games_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "public"."profiles"("id") ON DELETE RESTRICT;



ALTER TABLE ONLY "public"."games"
    ADD CONSTRAINT "games_home_team_id_fkey" FOREIGN KEY ("home_team_id") REFERENCES "public"."teams"("id") ON DELETE RESTRICT;



ALTER TABLE ONLY "public"."games"
    ADD CONSTRAINT "games_season_id_fkey" FOREIGN KEY ("season_id") REFERENCES "public"."seasons"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."games"
    ADD CONSTRAINT "games_week_id_fkey" FOREIGN KEY ("week_id") REFERENCES "public"."weeks"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."picks"
    ADD CONSTRAINT "picks_game_id_fkey" FOREIGN KEY ("game_id") REFERENCES "public"."games"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."picks"
    ADD CONSTRAINT "picks_selected_team_id_fkey" FOREIGN KEY ("selected_team_id") REFERENCES "public"."teams"("id") ON DELETE RESTRICT;



ALTER TABLE ONLY "public"."picks"
    ADD CONSTRAINT "picks_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."profiles"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."picks"
    ADD CONSTRAINT "picks_week_id_fkey" FOREIGN KEY ("week_id") REFERENCES "public"."weeks"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."profiles"
    ADD CONSTRAINT "profiles_id_fkey" FOREIGN KEY ("id") REFERENCES "auth"."users"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."season_participants"
    ADD CONSTRAINT "season_participants_season_id_fkey" FOREIGN KEY ("season_id") REFERENCES "public"."seasons"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."season_participants"
    ADD CONSTRAINT "season_participants_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."profiles"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."season_standings"
    ADD CONSTRAINT "season_standings_season_id_fkey" FOREIGN KEY ("season_id") REFERENCES "public"."seasons"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."season_standings"
    ADD CONSTRAINT "season_standings_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."profiles"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."season_standings"
    ADD CONSTRAINT "season_standings_week_id_fkey" FOREIGN KEY ("week_id") REFERENCES "public"."weeks"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."seasons"
    ADD CONSTRAINT "seasons_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "public"."profiles"("id") ON DELETE RESTRICT;



ALTER TABLE ONLY "public"."week_results"
    ADD CONSTRAINT "week_results_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."profiles"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."week_results"
    ADD CONSTRAINT "week_results_week_id_fkey" FOREIGN KEY ("week_id") REFERENCES "public"."weeks"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."weeks"
    ADD CONSTRAINT "weeks_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "public"."profiles"("id") ON DELETE RESTRICT;



ALTER TABLE ONLY "public"."weeks"
    ADD CONSTRAINT "weeks_season_id_fkey" FOREIGN KEY ("season_id") REFERENCES "public"."seasons"("id") ON DELETE CASCADE;



CREATE POLICY "Users can insert own profile" ON "public"."profiles" FOR INSERT WITH CHECK (("auth"."uid"() = "id"));



CREATE POLICY "Users can read own profile" ON "public"."profiles" FOR SELECT USING (("auth"."uid"() = "id"));



ALTER TABLE "public"."games" ENABLE ROW LEVEL SECURITY;


ALTER TABLE "public"."picks" ENABLE ROW LEVEL SECURITY;


ALTER TABLE "public"."profiles" ENABLE ROW LEVEL SECURITY;


ALTER TABLE "public"."season_participants" ENABLE ROW LEVEL SECURITY;


ALTER TABLE "public"."season_standings" ENABLE ROW LEVEL SECURITY;


ALTER TABLE "public"."seasons" ENABLE ROW LEVEL SECURITY;


ALTER TABLE "public"."settings" ENABLE ROW LEVEL SECURITY;


ALTER TABLE "public"."teams" ENABLE ROW LEVEL SECURITY;


ALTER TABLE "public"."week_results" ENABLE ROW LEVEL SECURITY;


ALTER TABLE "public"."weeks" ENABLE ROW LEVEL SECURITY;




ALTER PUBLICATION "supabase_realtime" OWNER TO "postgres";


GRANT USAGE ON SCHEMA "public" TO "postgres";
GRANT USAGE ON SCHEMA "public" TO "anon";
GRANT USAGE ON SCHEMA "public" TO "authenticated";
GRANT USAGE ON SCHEMA "public" TO "service_role";
GRANT USAGE ON SCHEMA "public" TO "supabase_auth_admin";

























































































































































REVOKE ALL ON FUNCTION "public"."custom_access_token_hook"("event" "jsonb") FROM PUBLIC;
GRANT ALL ON FUNCTION "public"."custom_access_token_hook"("event" "jsonb") TO "service_role";
GRANT ALL ON FUNCTION "public"."custom_access_token_hook"("event" "jsonb") TO "supabase_auth_admin";



GRANT ALL ON FUNCTION "public"."handle_new_user"() TO "anon";
GRANT ALL ON FUNCTION "public"."handle_new_user"() TO "authenticated";
GRANT ALL ON FUNCTION "public"."handle_new_user"() TO "service_role";



GRANT ALL ON FUNCTION "public"."update_updated_at_column"() TO "anon";
GRANT ALL ON FUNCTION "public"."update_updated_at_column"() TO "authenticated";
GRANT ALL ON FUNCTION "public"."update_updated_at_column"() TO "service_role";


















GRANT ALL ON TABLE "public"."games" TO "anon";
GRANT ALL ON TABLE "public"."games" TO "authenticated";
GRANT ALL ON TABLE "public"."games" TO "service_role";



GRANT ALL ON TABLE "public"."picks" TO "anon";
GRANT ALL ON TABLE "public"."picks" TO "authenticated";
GRANT ALL ON TABLE "public"."picks" TO "service_role";



GRANT ALL ON TABLE "public"."profiles" TO "anon";
GRANT ALL ON TABLE "public"."profiles" TO "authenticated";
GRANT ALL ON TABLE "public"."profiles" TO "service_role";



GRANT ALL ON TABLE "public"."season_participants" TO "anon";
GRANT ALL ON TABLE "public"."season_participants" TO "authenticated";
GRANT ALL ON TABLE "public"."season_participants" TO "service_role";



GRANT ALL ON TABLE "public"."season_standings" TO "anon";
GRANT ALL ON TABLE "public"."season_standings" TO "authenticated";
GRANT ALL ON TABLE "public"."season_standings" TO "service_role";



GRANT ALL ON TABLE "public"."seasons" TO "anon";
GRANT ALL ON TABLE "public"."seasons" TO "authenticated";
GRANT ALL ON TABLE "public"."seasons" TO "service_role";



GRANT ALL ON TABLE "public"."settings" TO "anon";
GRANT ALL ON TABLE "public"."settings" TO "authenticated";
GRANT ALL ON TABLE "public"."settings" TO "service_role";



GRANT ALL ON TABLE "public"."teams" TO "anon";
GRANT ALL ON TABLE "public"."teams" TO "authenticated";
GRANT ALL ON TABLE "public"."teams" TO "service_role";



GRANT ALL ON TABLE "public"."week_results" TO "anon";
GRANT ALL ON TABLE "public"."week_results" TO "authenticated";
GRANT ALL ON TABLE "public"."week_results" TO "service_role";



GRANT ALL ON TABLE "public"."weeks" TO "anon";
GRANT ALL ON TABLE "public"."weeks" TO "authenticated";
GRANT ALL ON TABLE "public"."weeks" TO "service_role";









ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON SEQUENCES TO "postgres";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON SEQUENCES TO "anon";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON SEQUENCES TO "authenticated";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON SEQUENCES TO "service_role";






ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON FUNCTIONS TO "postgres";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON FUNCTIONS TO "anon";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON FUNCTIONS TO "authenticated";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON FUNCTIONS TO "service_role";






ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON TABLES TO "postgres";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON TABLES TO "anon";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON TABLES TO "authenticated";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON TABLES TO "service_role";


-- ============================================================================
-- Items not captured by supabase db dump (non-public schema objects)
-- ============================================================================

-- Enforce only one active week at a time
CREATE UNIQUE INDEX "weeks_single_active" ON "public"."weeks" (("status")) WHERE ("status" = 'active'::"public"."week_status");

-- Trigger to auto-create profile on user signup
CREATE TRIGGER "on_auth_user_created"
  AFTER INSERT ON "auth"."users"
  FOR EACH ROW
  EXECUTE FUNCTION "public"."handle_new_user"();

-- Storage: avatars bucket
INSERT INTO "storage"."buckets" ("id", "name", "public")
VALUES ('avatars', 'avatars', true)
ON CONFLICT ("id") DO NOTHING;

-- Storage RLS: public read access
CREATE POLICY "Public read access to avatars"
  ON "storage"."objects"
  FOR SELECT
  TO "public"
  USING ("bucket_id" = 'avatars');

-- Storage RLS: users can upload their own avatar
CREATE POLICY "Allow users to upload their own avatar"
  ON "storage"."objects"
  FOR INSERT
  TO "authenticated"
  WITH CHECK ("bucket_id" = 'avatars' AND "auth"."uid"()::text = split_part("name", '/', 1));

-- Storage RLS: users can update their own avatar
CREATE POLICY "Allow users to update their own avatar"
  ON "storage"."objects"
  FOR UPDATE
  TO "authenticated"
  USING ("bucket_id" = 'avatars' AND "auth"."uid"()::text = split_part("name", '/', 1));

-- Storage RLS: users can delete their own avatar
CREATE POLICY "Allow users to delete their own avatar"
  ON "storage"."objects"
  FOR DELETE
  TO "authenticated"
  USING ("bucket_id" = 'avatars' AND "auth"."uid"()::text = split_part("name", '/', 1));





























