-- khan
-- https://github.com/topfreegames/khan
--
-- Licensed under the MIT license:
-- http://www.opensource.org/licenses/mit-license
-- Copyright © 2016 Top Free Games <backend@tfgco.com>

REVOKE ALL ON SCHEMA public FROM khan;
DROP DATABASE IF EXISTS khan;

DROP ROLE khan;

CREATE ROLE khan LOGIN
  SUPERUSER INHERIT CREATEDB CREATEROLE;

CREATE DATABASE khan
  WITH OWNER = khan
       ENCODING = 'UTF8'
       TABLESPACE = pg_default
       TEMPLATE = template0;

GRANT ALL ON SCHEMA public TO khan;
