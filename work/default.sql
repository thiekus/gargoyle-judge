-- Gargoyle Judgement System main database schema
-- Copyright (C) Thiekus 2019

-- User table
DROP TABLE IF EXISTS %TABLEPREFIX%users;
CREATE TABLE %TABLEPREFIX%users (
    id INTEGER PRIMARY KEY %AUTOINCREMENT%,
    username VARCHAR(50) NOT NULL UNIQUE,
    password VARCHAR(64) NOT NULL,
    email VARCHAR(50) NOT NULL UNIQUE,
    iguser VARCHAR(50) DEFAULT '',
    display_name VARCHAR(50) NOT NULL,
    address VARCHAR(150) NOT NULL,
    avatar VARCHAR(150) NOT NULL,
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
    access_user INTEGER DEFAULT 1
);
-- Default roles
INSERT INTO %TABLEPREFIX%roles (rolename, access_root, access_jury, access_user) VALUES
    ('Peserta', 0, 0, 1);
INSERT INTO %TABLEPREFIX%roles (rolename, access_root, access_jury, access_user) VALUES
    ('Admin', 1, 1, 1);
INSERT INTO %TABLEPREFIX%roles (rolename, access_root, access_jury, access_user) VALUES
    ('Juri', 0, 1, 0);

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
    quest_count INTEGER,
    is_unlocked INTEGER,
    is_private INTEGER,
    is_trainer INTEGER,
    start_timestamp INTEGER,
    end_timestamp INTEGER,
    max_runtime INTEGER
);

-- Question List
DROP TABLE IF EXISTS %TABLEPREFIX%quests;
CREATE TABLE %TABLEPREFIX%quests (
    id INTEGER PRIMARY KEY %AUTOINCREMENT%,
    contest_id INTEGER,
    quest_name VARCHAR(50) NOT NULL,
    description TEXT,
    time_limit INTEGER,
    mem_limit INTEGER,
    max_attempts INTEGER
);
