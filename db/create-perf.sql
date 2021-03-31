-- khan
-- https://github.com/topfreegames/khan
--
-- Licensed under the MIT license:
-- http://www.opensource.org/licenses/mit-license
-- Copyright © 2016 Top Free Games <backend@tfgco.com>

CREATE ROLE khan_perf LOGIN
  SUPERUSER INHERIT CREATEDB CREATEROLE;

CREATE DATABASE khan_perf
  WITH OWNER = khan_perf
       ENCODING = 'UTF8'
       TABLESPACE = pg_default
       TEMPLATE = template0;

GRANT ALL ON SCHEMA public TO khan_perf;
