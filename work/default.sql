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
    display_name VARCHAR(50) NOT NULL,
    gender VARCHAR(5) NOT NULL,
    address VARCHAR(150) NOT NULL,
    institution VARCHAR(150) NOT NULL,
    country_id VARCHAR(10) NOT NULL,
    avatar VARCHAR(250) NOT NULL,
    role INTEGER DEFAULT 2,
    verified INTEGER DEFAULT 0,
    banned INTEGER DEFAULT 0,
    create_time INTEGER DEFAULT 0
);

-- User roles
DROP TABLE IF EXISTS %TABLEPREFIX%roles;
CREATE TABLE %TABLEPREFIX%roles (
    id INTEGER PRIMARY KEY %AUTOINCREMENT%,
    rolename VARCHAR(20) NOT NULL DEFAULT 'undefined',
    access_root INTEGER DEFAULT 0,
    access_jury INTEGER DEFAULT 0,
    access_contestant INTEGER DEFAULT 1
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
    name VARCHAR(50) NOT NULL
);

-- Contestant Group Member relations
DROP TABLE IF EXISTS %TABLEPREFIX%group_members;
CREATE TABLE %TABLEPREFIX%group_members (
    id INTEGER PRIMARY KEY %AUTOINCREMENT%,
    user_id INTEGER,
    group_id INTEGER
);

-- News
DROP TABLE IF EXISTS %TABLEPREFIX%news;
CREATE TABLE %TABLEPREFIX%news (
    id INTEGER PRIMARY KEY %AUTOINCREMENT%,
    author_id INTEGER,
    post_time INTEGER DEFAULT 0,
    title VARCHAR(50) DEFAULT 'untitled',
    body TEXT DEFAULT ''
);

-- Contest List
DROP TABLE IF EXISTS %TABLEPREFIX%contests;
CREATE TABLE %TABLEPREFIX%contests (
    id INTEGER PRIMARY KEY %AUTOINCREMENT%,
    title VARCHAR(50) NOT NULL DEFAULT 'untitled',
    description TEXT DEFAULT '',
    problem_count INTEGER,
    contest_group_id INTEGER,
    is_unlocked INTEGER,
    is_public INTEGER,
    is_trainer INTEGER,
    must_stream INTEGER,
    start_timestamp INTEGER,
    end_timestamp INTEGER,
    max_runtime INTEGER
);

-- Question List
DROP TABLE IF EXISTS %TABLEPREFIX%problems;
CREATE TABLE %TABLEPREFIX%problems (
    id INTEGER PRIMARY KEY %AUTOINCREMENT%,
    contest_id INTEGER,
    problem_name VARCHAR(50) NOT NULL,
    description TEXT,
    time_limit INTEGER,
    mem_limit INTEGER,
    max_attempts INTEGER
);
