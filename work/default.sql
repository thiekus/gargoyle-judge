-- Gargoyle Judgement System main database schema
-- Copyright (C) Thiekus 2019

-- User table
{{if eq .Driver "sqlserver"}}
IF OBJECT_ID('{{.TablePrefix}}users', 'U') IS NOT NULL DROP TABLE {{.TablePrefix}}users; 
{{else}}
DROP TABLE IF EXISTS {{.TablePrefix}}users;
{{end}}
CREATE TABLE {{.TablePrefix}}users (
    id INTEGER PRIMARY KEY {{.AutoIncrement}},
    username VARCHAR(50) NOT NULL UNIQUE,
    password VARCHAR(64) NOT NULL,
    salt VARCHAR(32) NOT NULL,
    email VARCHAR(50) NOT NULL UNIQUE,
    display_name VARCHAR(50) NOT NULL DEFAULT 'Somebody',
    gender VARCHAR(5) NOT NULL DEFAULT 'M',
    address VARCHAR(150) NOT NULL DEFAULT 'Somewhere',
    institution VARCHAR(150) NOT NULL DEFAULT 'Any Organization',
    country_id VARCHAR(10) NOT NULL DEFAULT 'id',
    avatar VARCHAR(250) NOT NULL DEFAULT '',
    role INTEGER NOT NULL DEFAULT 2,
    verified INTEGER NOT NULL DEFAULT 0,
    banned INTEGER NOT NULL DEFAULT 0,
    create_time INTEGER NOT NULL DEFAULT 0
);

-- User roles
{{if eq .Driver "sqlserver"}}
IF OBJECT_ID('{{.TablePrefix}}roles', 'U') IS NOT NULL DROP TABLE {{.TablePrefix}}roles; 
{{else}}
DROP TABLE IF EXISTS {{.TablePrefix}}roles;
{{end}}
CREATE TABLE {{.TablePrefix}}roles (
    id INTEGER PRIMARY KEY {{.AutoIncrement}},
    rolename VARCHAR(20) NOT NULL DEFAULT 'undefined',
    access_root INTEGER NOT NULL DEFAULT 0,
    access_jury INTEGER NOT NULL DEFAULT 0,
    access_contestant INTEGER NOT NULL DEFAULT 1
);
-- Default roles
INSERT INTO {{.TablePrefix}}roles (rolename, access_root, access_jury, access_contestant) VALUES
    ('Contestant', 0, 0, 1);
INSERT INTO {{.TablePrefix}}roles (rolename, access_root, access_jury, access_contestant) VALUES
    ('Administrator', 1, 1, 0);
INSERT INTO {{.TablePrefix}}roles (rolename, access_root, access_jury, access_contestant) VALUES
    ('Jury', 0, 1, 0);

-- Contestant Group
{{if eq .Driver "sqlserver"}}
IF OBJECT_ID('{{.TablePrefix}}groups', 'U') IS NOT NULL DROP TABLE {{.TablePrefix}}groups; 
{{else}}
DROP TABLE IF EXISTS {{.TablePrefix}}groups;
{{end}}
CREATE TABLE {{.TablePrefix}}groups (
    id INTEGER PRIMARY KEY {{.AutoIncrement}},
    name VARCHAR(50) NOT NULL UNIQUE
);

-- Contestant Group Member relations
{{if eq .Driver "sqlserver"}}
IF OBJECT_ID('{{.TablePrefix}}group_members', 'U') IS NOT NULL DROP TABLE {{.TablePrefix}}group_members; 
{{else}}
DROP TABLE IF EXISTS {{.TablePrefix}}group_members;
{{end}}
CREATE TABLE {{.TablePrefix}}group_members (
    id INTEGER PRIMARY KEY {{.AutoIncrement}},
    user_id INTEGER NOT NULL,
    group_id INTEGER NOT NULL
);

-- News
{{if eq .Driver "sqlserver"}}
IF OBJECT_ID('{{.TablePrefix}}news', 'U') IS NOT NULL DROP TABLE {{.TablePrefix}}news; 
{{else}}
DROP TABLE IF EXISTS {{.TablePrefix}}news;
{{end}}
CREATE TABLE {{.TablePrefix}}news (
    id INTEGER PRIMARY KEY {{.AutoIncrement}},
    author_id INTEGER NOT NULL,
    post_time INTEGER NOT NULL DEFAULT 0,
    title VARCHAR(50) NOT NULL DEFAULT 'Untitled',
    body TEXT NOT NULL DEFAULT ''
);

-- Contest List
{{if eq .Driver "sqlserver"}}
IF OBJECT_ID('{{.TablePrefix}}contests', 'U') IS NOT NULL DROP TABLE {{.TablePrefix}}contests; 
{{else}}
DROP TABLE IF EXISTS {{.TablePrefix}}contests;
{{end}}
CREATE TABLE {{.TablePrefix}}contests (
    id INTEGER PRIMARY KEY {{.AutoIncrement}},
    title VARCHAR(50) NOT NULL DEFAULT 'Untitled',
    description TEXT NOT NULL DEFAULT '',
    problem_count INTEGER NOT NULL DEFAULT 0,
    contest_group_id INTEGER NOT NULL DEFAULT 0,
    is_unlocked INTEGER NOT NULL DEFAULT 1,
    is_public INTEGER NOT NULL DEFAULT 1,
    must_stream INTEGER NOT NULL DEFAULT 0,
    start_timestamp INTEGER NOT NULL DEFAULT 0,
    end_timestamp INTEGER NOT NULL DEFAULT 0,
    max_runtime INTEGER NOT NULL DEFAULT 0
);

-- Problem List
{{if eq .Driver "sqlserver"}}
IF OBJECT_ID('{{.TablePrefix}}problems', 'U') IS NOT NULL DROP TABLE {{.TablePrefix}}problems; 
{{else}}
DROP TABLE IF EXISTS {{.TablePrefix}}problems;
{{end}}
CREATE TABLE {{.TablePrefix}}problems (
    id INTEGER PRIMARY KEY {{.AutoIncrement}},
    contest_id INTEGER NOT NULL,
    problem_name VARCHAR(50) NOT NULL DEFAULT 'Untitled Problem',
    description TEXT NOT NULL DEFAULT '',
    time_limit INTEGER NOT NULL DEFAULT 1000,
    mem_limit INTEGER NOT NULL DEFAULT 32,
    max_attempts INTEGER NOT NULL DEFAULT 0
);

-- Contest Access
{{if eq .Driver "sqlserver"}}
IF OBJECT_ID('{{.TablePrefix}}contest_access', 'U') IS NOT NULL DROP TABLE {{.TablePrefix}}contest_access; 
{{else}}
DROP TABLE IF EXISTS {{.TablePrefix}}contest_access;
{{end}}
CREATE TABLE {{.TablePrefix}}contest_access (
    id INTEGER PRIMARY KEY {{.AutoIncrement}},
    id_user INTEGER NOT NULL,
    id_contest INTEGER NOT NULL,
    start_time INTEGER NOT NULL DEFAULT 0,
    end_time INTEGER NOT NULL DEFAULT 0,
    allowed INTEGER NOT NULL DEFAULT 1
);

-- Contest Submissions
{{if eq .Driver "sqlserver"}}
IF OBJECT_ID('{{.TablePrefix}}submissions', 'U') IS NOT NULL DROP TABLE {{.TablePrefix}}submissions; 
{{else}}
DROP TABLE IF EXISTS {{.TablePrefix}}submissions;
{{end}}
CREATE TABLE {{.TablePrefix}}submissions (
    id INTEGER PRIMARY KEY {{.AutoIncrement}},
    id_problem INTEGER NOT NULL,
    id_user INTEGER NOT NULL,
    id_cache VARCHAR(64) NOT NULL UNIQUE,
    lang VARCHAR(50) NOT NULL DEFAULT 'c',
    code TEXT NOT NULL DEFAULT '',
    status VARCHAR(4) NOT NULL DEFAULT 'QU',
    details TEXT NOT NULL DEFAULT '',
    score INTEGER NOT NULL DEFAULT 0,
    submit_time INTEGER NOT NULL DEFAULT 0,
    compile_time REAL NOT NULL DEFAULT 0,
    compile_stdout TEXT NOT NULL DEFAULT '',
    compile_stderr TEXT NOT NULL DEFAULT ''
);
