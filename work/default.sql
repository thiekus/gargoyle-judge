-- Gargoyle Judgement System main database schema
-- Copyright (C) Thiekus 2019

-- User table
DROP TABLE IF EXISTS gy_users;
CREATE TABLE gy_users (
    id INTEGER PRIMARY KEY %AUTOINCREMENT%,
    username VARCHAR(50) NOT NULL UNIQUE,
    password VARCHAR(64) NOT NULL,
    email VARCHAR(50) NOT NULL UNIQUE,
    iguser VARCHAR(50),
    display_name VARCHAR(50) NOT NULL,
    address VARCHAR(150) NOT NULL,
    avatar VARCHAR(150) NOT NULL,
    role INTEGER,
    verified INTEGER,
    banned INTEGER,
    create_time INTEGER
);

-- User roles
DROP TABLE IF EXISTS gy_roles;
CREATE TABLE gy_roles (
    id INTEGER PRIMARY KEY %AUTOINCREMENT%,
    rolename VARCHAR(20) NOT NULL,
    access_root INTEGER,
    access_jury INTEGER,
    access_user INTEGER
);
-- Default roles
INSERT INTO gy_roles (rolename, access_root, access_jury, access_user) VALUES
    ('Peserta', 0, 0, 1);
INSERT INTO gy_roles (rolename, access_root, access_jury, access_user) VALUES
    ('Admin', 1, 1, 1);
INSERT INTO gy_roles (rolename, access_root, access_jury, access_user) VALUES
    ('Juri', 0, 1, 0);

-- News
DROP TABLE IF EXISTS gy_news;
CREATE TABLE gy_news (
    id INTEGER PRIMARY KEY %AUTOINCREMENT%,
    author_id INTEGER,
    post_time INTEGER,
    title VARCHAR(50),
    body TEXT
);

-- Contest List
DROP TABLE IF EXISTS gy_contests;
CREATE TABLE gy_contests (
    id INTEGER PRIMARY KEY %AUTOINCREMENT%,
    title VARCHAR(50) NOT NULL,
    description TEXT,
    quest_count INTEGER,
    is_unlocked INTEGER,
    is_private INTEGER,
    is_trainer INTEGER,
    start_timestamp INTEGER,
    end_timestamp INTEGER,
    max_runtime INTEGER
);

-- Question List
DROP TABLE IF EXISTS gy_quests;
CREATE TABLE gy_quests (
    id INTEGER PRIMARY KEY %AUTOINCREMENT%,
    contest_id INTEGER,
    quest_name VARCHAR(50) NOT NULL,
    description TEXT,
    time_limit INTEGER,
    mem_limit INTEGER,
    max_attempts INTEGER
);
