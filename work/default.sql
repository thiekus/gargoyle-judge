-- Gargoyle Judgement System main database schema
-- Copyright (C) Thiekus 2019

-- User table
DROP TABLE IF EXISTS %TABLEPREFIX%users;
CREATE TABLE %TABLEPREFIX%users (
    id INTEGER PRIMARY KEY %AUTOINCREMENT%,
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
DROP TABLE IF EXISTS %TABLEPREFIX%roles;
CREATE TABLE %TABLEPREFIX%roles (
    id INTEGER PRIMARY KEY %AUTOINCREMENT%,
    rolename VARCHAR(20) NOT NULL DEFAULT 'undefined',
    access_root INTEGER NOT NULL DEFAULT 0,
    access_jury INTEGER NOT NULL DEFAULT 0,
    access_contestant INTEGER NOT NULL DEFAULT 1
);
-- Default roles
INSERT INTO %TABLEPREFIX%roles (rolename, access_root, access_jury, access_contestant) VALUES
    ('Contestant', 0, 0, 1);
INSERT INTO %TABLEPREFIX%roles (rolename, access_root, access_jury, access_contestant) VALUES
    ('Admin', 1, 1, 0);
INSERT INTO %TABLEPREFIX%roles (rolename, access_root, access_jury, access_contestant) VALUES
    ('Jury', 0, 1, 0);

-- Contestant Group
DROP TABLE IF EXISTS %TABLEPREFIX%groups;
CREATE TABLE %TABLEPREFIX%groups (
    id INTEGER PRIMARY KEY %AUTOINCREMENT%,
    name VARCHAR(50) NOT NULL UNIQUE
);

-- Contestant Group Member relations
DROP TABLE IF EXISTS %TABLEPREFIX%group_members;
CREATE TABLE %TABLEPREFIX%group_members (
    id INTEGER PRIMARY KEY %AUTOINCREMENT%,
    user_id INTEGER NOT NULL,
    group_id INTEGER NOT NULL
);

-- News
DROP TABLE IF EXISTS %TABLEPREFIX%news;
CREATE TABLE %TABLEPREFIX%news (
    id INTEGER PRIMARY KEY %AUTOINCREMENT%,
    author_id INTEGER NOT NULL,
    post_time INTEGER NOT NULL DEFAULT 0,
    title VARCHAR(50) NOT NULL DEFAULT 'Untitled',
    body TEXT NOT NULL DEFAULT ''
);

-- Contest List
DROP TABLE IF EXISTS %TABLEPREFIX%contests;
CREATE TABLE %TABLEPREFIX%contests (
    id INTEGER PRIMARY KEY %AUTOINCREMENT%,
    title VARCHAR(50) NOT NULL DEFAULT 'Untitled',
    description TEXT DEFAULT '',
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
DROP TABLE IF EXISTS %TABLEPREFIX%problems;
CREATE TABLE %TABLEPREFIX%problems (
    id INTEGER PRIMARY KEY %AUTOINCREMENT%,
    contest_id INTEGER NOT NULL,
    problem_name VARCHAR(50) NOT NULL DEFAULT 'Untitled Problem',
    description TEXT NOT NULL DEFAULT '',
    time_limit INTEGER NOT NULL DEFAULT 1000,
    mem_limit INTEGER NOT NULL DEFAULT 32,
    max_attempts INTEGER NOT NULL DEFAULT 0
);

-- Contest Access
DROP TABLE IF EXISTS %TABLEPREFIX%contest_access;
CREATE TABLE %TABLEPREFIX%contest_access (
    id INTEGER PRIMARY KEY %AUTOINCREMENT%,
    id_user INTEGER NOT NULL,
    id_contest INTEGER NOT NULL,
    start_time INTEGER NOT NULL DEFAULT 0,
    end_time INTEGER NOT NULL DEFAULT 0,
    allowed INTEGER NOT NULL DEFAULT 1
);
