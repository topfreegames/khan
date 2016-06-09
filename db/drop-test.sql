-- khan
-- https://github.com/topfreegames/khan
--
-- Licensed under the MIT license:
-- http://www.opensource.org/licenses/mit-license
-- Copyright © 2016 Top Free Games <backend@tfgco.com>

REVOKE ALL ON SCHEMA public FROM khan_test;
DROP DATABASE IF EXISTS khan_test;
DROP ROLE khan_test;

CREATE ROLE khan_test LOGIN
  SUPERUSER INHERIT CREATEDB CREATEROLE;

CREATE DATABASE khan_test
  WITH OWNER = khan_test
       ENCODING = 'UTF8'
       TABLESPACE = pg_default
       TEMPLATE = template0;

GRANT ALL ON SCHEMA public TO khan_test;
