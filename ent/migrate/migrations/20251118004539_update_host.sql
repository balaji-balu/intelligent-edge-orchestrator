-- Rename a column from "last_heartbeat" to "last_seen"
ALTER TABLE "host" RENAME COLUMN "last_heartbeat" TO "last_seen";
-- Modify "host" table
ALTER TABLE "host" ADD COLUMN "misses" bigint NULL;
