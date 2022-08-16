-- Gargoyle Judgement System main database schema
-- Copyright (C) Thiekus 2019

-- Programming Languages
-- {{if eq .Driver "sqlserver"}}
IF OBJECT_ID('{{.TablePrefix}}languages', 'U') IS NOT NULL DROP TABLE {{.TablePrefix}}languages;
-- {{else}}
DROP TABLE IF EXISTS {{.TablePrefix}}languages;
-- {{end}}
CREATE TABLE {{.TablePrefix}}languages (
    id INTEGER PRIMARY KEY {{.AutoIncrement}},
    ext_name VARCHAR(20) NOT NULL DEFAULT 'c',
    display_name VARCHAR(50) NOT NULL DEFAULT 'Unknown',
    enabled INTEGER NOT NULL DEFAULT 1,
    syntax_name VARCHAR(50) NOT NULL DEFAULT 'c_cpp',
    source_name VARCHAR(50) NOT NULL DEFAULT 'appmain.c',
    exe_name VARCHAR(50) NOT NULL DEFAULT 'appmain',
    compile_cmd VARCHAR(200) NOT NULL DEFAULT 'gcc \{\{.WorkPath\}\}/\{\{.SourceName\}\}',
    exec_cmd VARCHAR(200) NOT NULL DEFAULT '\{\{.WorkPath\}\}/appmain',
    enable_sandbox INTEGER NOT NULL DEFAULT 1,
    limit_memory INTEGER NOT NULL DEFAULT 0,
    limit_syscall INTEGER NOT NULL DEFAULT 0,
    preg_replace_from VARCHAR(200) NOT NULL DEFAULT '',
    preg_replace_to VARCHAR(200) NOT NULL DEFAULT '',
    forbidden_keys TEXT NOT NULL DEFAULT ''
);
-- Default languages
-- Pure C using GCC
INSERT INTO {{.TablePrefix}}languages (ext_name, display_name, enabled, syntax_name, source_name, exe_name, compile_cmd, exec_cmd, enable_sandbox, limit_memory, limit_syscall, preg_replace_from, preg_replace_to, forbidden_keys)
    VALUES ('c', 'C', 1, 'c_cpp', 'appmain.c', 'appmain', 'gcc -Wall -o \{\{.WorkPath\}\}/\{\{.ExeName\}\} -O2 -std=gnu99 -lm \{\{.WorkPath\}\}/\{\{.SourceName\}\}', '\{\{.WorkPath\}\}/\{\{.ExeName\}\}', 1, 1, 1, '', '', '');
-- C++ using GCC
INSERT INTO {{.TablePrefix}}languages (ext_name, display_name, enabled, syntax_name, source_name, exe_name, compile_cmd, exec_cmd, enable_sandbox, limit_memory, limit_syscall, preg_replace_from, preg_replace_to, forbidden_keys)
    VALUES ('cpp', 'C++', 1, 'c_cpp', 'appmain.cpp', 'appmain', 'g++ -Wall -o \{\{.WorkPath\}\}/\{\{.ExeName\}\} -O2 -std=gnu++14 -lm \{\{.WorkPath\}\}/\{\{.SourceName\}\}', '\{\{.WorkPath\}\}/\{\{.ExeName\}\}', 1, 1, 1, '', '', '');
-- Pascal using FPC
INSERT INTO {{.TablePrefix}}languages (ext_name, display_name, enabled, syntax_name, source_name, exe_name, compile_cmd, exec_cmd, enable_sandbox, limit_memory, limit_syscall, preg_replace_from, preg_replace_to, forbidden_keys)
    VALUES ('pas', 'Pascal', 1, 'pascal', 'appmain.pas', 'appmain', 'fpc -O2 -XS -Sg \{\{.WorkPath\}\}/\{\{.SourceName\}\}', '\{\{.WorkPath\}\}/\{\{.ExeName\}\}', 1, 1, 1, '', '', '');
-- Java
INSERT INTO {{.TablePrefix}}languages (ext_name, display_name, enabled, syntax_name, source_name, exe_name, compile_cmd, exec_cmd, enable_sandbox, limit_memory, limit_syscall, preg_replace_from, preg_replace_to, forbidden_keys)
    VALUES ('java', 'Java', 1, 'java', 'PandoraApp.java', 'PandoraApp.class', 'javac -encoding UTF-8 -d . \{\{.WorkPath\}\}/\{\{.SourceName\}\}', 'java -Xmx\{\{.MemLimit\}\}m -cp \{\{.WorkPath\}\} PandoraApp', 0, 0, 1, 'class .*\\{|class .*\\s{', 'class PandoraApp {', '');
-- Golang
INSERT INTO {{.TablePrefix}}languages (ext_name, display_name, enabled, syntax_name, source_name, exe_name, compile_cmd, exec_cmd, enable_sandbox, limit_memory, limit_syscall, preg_replace_from, preg_replace_to, forbidden_keys)
    VALUES ('go', 'Go', 1, 'golang', 'appmain.go', 'appmain', 'go build -o \{\{.WorkPath\}\}/\{\{.ExeName\}\} \{\{.WorkPath\}\}/', '\{\{.WorkPath\}\}/\{\{.ExeName\}\}', 1, 1, 1, '', '', '');

-- User table
-- {{if eq .Driver "sqlserver"}}
IF OBJECT_ID('{{.TablePrefix}}users', 'U') IS NOT NULL DROP TABLE {{.TablePrefix}}users;
-- {{else}}
DROP TABLE IF EXISTS {{.TablePrefix}}users;
-- {{end}}
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
    syntax_theme VARCHAR(50) NOT NULL DEFAULT 'eclipse',
    role INTEGER NOT NULL DEFAULT 2,
    active INTEGER NOT NULL DEFAULT 1,
    banned INTEGER NOT NULL DEFAULT 0,
    create_time INTEGER NOT NULL DEFAULT 0,
    lastaccess_time INTEGER NOT NULL DEFAULT 0
);

-- User roles
-- {{if eq .Driver "sqlserver"}}
IF OBJECT_ID('{{.TablePrefix}}roles', 'U') IS NOT NULL DROP TABLE {{.TablePrefix}}roles;
-- {{else}}
DROP TABLE IF EXISTS {{.TablePrefix}}roles;
-- {{end}}
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
-- {{if eq .Driver "sqlserver"}}
IF OBJECT_ID('{{.TablePrefix}}groups', 'U') IS NOT NULL DROP TABLE {{.TablePrefix}}groups;
-- {{else}}
DROP TABLE IF EXISTS {{.TablePrefix}}groups;
-- {{end}}
CREATE TABLE {{.TablePrefix}}groups (
    id INTEGER PRIMARY KEY {{.AutoIncrement}},
    name VARCHAR(50) NOT NULL UNIQUE
);

-- Contestant Group Member relations
-- {{if eq .Driver "sqlserver"}}
IF OBJECT_ID('{{.TablePrefix}}group_members', 'U') IS NOT NULL DROP TABLE {{.TablePrefix}}group_members;
-- {{else}}
DROP TABLE IF EXISTS {{.TablePrefix}}group_members;
-- {{end}}
CREATE TABLE {{.TablePrefix}}group_members (
    id INTEGER PRIMARY KEY {{.AutoIncrement}},
    user_id INTEGER NOT NULL,
    group_id INTEGER NOT NULL
);

-- News
-- {{if eq .Driver "sqlserver"}}
IF OBJECT_ID('{{.TablePrefix}}news', 'U') IS NOT NULL DROP TABLE {{.TablePrefix}}news;
-- {{else}}
DROP TABLE IF EXISTS {{.TablePrefix}}news;
-- {{end}}
CREATE TABLE {{.TablePrefix}}news (
    id INTEGER PRIMARY KEY {{.AutoIncrement}},
    author_id INTEGER NOT NULL,
    post_time INTEGER NOT NULL DEFAULT 0,
    title VARCHAR(50) NOT NULL DEFAULT 'Untitled',
    body TEXT NOT NULL DEFAULT ''
);

-- Contest List
-- {{if eq .Driver "sqlserver"}}
IF OBJECT_ID('{{.TablePrefix}}contests', 'U') IS NOT NULL DROP TABLE {{.TablePrefix}}contests;
-- {{else}}
DROP TABLE IF EXISTS {{.TablePrefix}}contests;
-- {{end}}
CREATE TABLE {{.TablePrefix}}contests (
    id INTEGER PRIMARY KEY {{.AutoIncrement}},
    title VARCHAR(50) NOT NULL DEFAULT 'Untitled',
    description TEXT NOT NULL DEFAULT '',
    style VARCHAR(20) NOT NULL DEFAULT 'ICPC',
    allowed_lang VARCHAR(200) NOT NULL DEFAULT '1,2,4',
    problem_count INTEGER NOT NULL DEFAULT 0,
    contest_group_id INTEGER NOT NULL DEFAULT 0,
    enable_freeze INTEGER NOT NULL DEFAULT 0,
    active INTEGER NOT NULL DEFAULT 1,
    allow_public INTEGER NOT NULL DEFAULT 1,
    must_stream INTEGER NOT NULL DEFAULT 0,
    start_timestamp INTEGER NOT NULL DEFAULT 0,
    end_timestamp INTEGER NOT NULL DEFAULT 0,
    freeze_timestamp INTEGER NOT NULL DEFAULT 0,
    unfreeze_timestamp INTEGER NOT NULL DEFAULT 0,
    max_runtime INTEGER NOT NULL DEFAULT 0,
    penalty_time INTEGER NOT NULL DEFAULT 0
);

-- Problem List
-- {{if eq .Driver "sqlserver"}}
IF OBJECT_ID('{{.TablePrefix}}problems', 'U') IS NOT NULL DROP TABLE {{.TablePrefix}}problems;
-- {{else}}
DROP TABLE IF EXISTS {{.TablePrefix}}problems;
-- {{end}}
CREATE TABLE {{.TablePrefix}}problems (
    id INTEGER PRIMARY KEY {{.AutoIncrement}},
    contest_id INTEGER NOT NULL,
    problem_name VARCHAR(50) NOT NULL DEFAULT 'Untitled Problem',
    problem_shortname VARCHAR(20) NOT NULL DEFAULT 'Untitled',
    description TEXT NOT NULL DEFAULT '',
    time_limit INTEGER NOT NULL DEFAULT 1,
    mem_limit INTEGER NOT NULL DEFAULT 32,
    max_attempts INTEGER NOT NULL DEFAULT 0
);

-- Contest Access
-- {{if eq .Driver "sqlserver"}}
IF OBJECT_ID('{{.TablePrefix}}contest_access', 'U') IS NOT NULL DROP TABLE {{.TablePrefix}}contest_access;
-- {{else}}
DROP TABLE IF EXISTS {{.TablePrefix}}contest_access;
-- {{end}}
CREATE TABLE {{.TablePrefix}}contest_access (
    id_user INTEGER NOT NULL,
    id_contest INTEGER NOT NULL,
    start_time INTEGER NOT NULL DEFAULT 0,
    end_time INTEGER NOT NULL DEFAULT 0,
    allowed INTEGER NOT NULL DEFAULT 1,
    PRIMARY KEY (id_user, id_contest)
);

-- Contest Submissions
-- {{if eq .Driver "sqlserver"}}
IF OBJECT_ID('{{.TablePrefix}}submissions', 'U') IS NOT NULL DROP TABLE {{.TablePrefix}}submissions;
-- {{else}}
DROP TABLE IF EXISTS {{.TablePrefix}}submissions;
-- {{end}}
CREATE TABLE {{.TablePrefix}}submissions (
    id INTEGER PRIMARY KEY {{.AutoIncrement}},
    id_problem INTEGER NOT NULL,
    id_user INTEGER NOT NULL,
    id_lang INTEGER NOT NULL DEFAULT 1,
    code TEXT NOT NULL DEFAULT '',
    verdict VARCHAR(4) NOT NULL DEFAULT 'QU',
    details TEXT NOT NULL DEFAULT '',
    score INTEGER NOT NULL DEFAULT 0,
    submit_time INTEGER NOT NULL DEFAULT 0,
    compile_time REAL NOT NULL DEFAULT 0,
    compile_stdout TEXT NOT NULL DEFAULT '',
    compile_stderr TEXT NOT NULL DEFAULT ''
);

-- Contest problem testcase
-- {{if eq .Driver "sqlserver"}}
IF OBJECT_ID('{{.TablePrefix}}testcases', 'U') IS NOT NULL DROP TABLE {{.TablePrefix}}testcases;
-- {{else}}
DROP TABLE IF EXISTS {{.TablePrefix}}testcases;
-- {{end}}
CREATE TABLE {{.TablePrefix}}testcases (
    id INTEGER PRIMARY KEY {{.AutoIncrement}},
    id_problem INTEGER NOT NULL,
    test_no INTEGER NOT NULL,
    input TEXT NOT NULL DEFAULT '',
    output TEXT NOT NULL DEFAULT ''
);

-- Contest problem testresults
-- {{if eq .Driver "sqlserver"}}
IF OBJECT_ID('{{.TablePrefix}}testresults', 'U') IS NOT NULL DROP TABLE {{.TablePrefix}}testresults;
-- {{else}}
DROP TABLE IF EXISTS {{.TablePrefix}}testresults;
-- {{end}}
CREATE TABLE {{.TablePrefix}}testresults (
    id INTEGER PRIMARY KEY {{.AutoIncrement}},
    id_problem INTEGER NOT NULL,
    id_submission INTEGER NOT NULL,
    test_no INTEGER NOT NULL,
    verdict VARCHAR(4) NOT NULL DEFAULT 'QU',
    time_elapsed REAL NOT NULL DEFAULT 0,
    memory_used INTEGER NOT NULL DEFAULT 0,
    score REAL NOT NULL DEFAULT 0
);

-- Contest score for internal beholder (Admin and Jury)
-- {{if eq .Driver "sqlserver"}}
IF OBJECT_ID('{{.TablePrefix}}scores_private', 'U') IS NOT NULL DROP TABLE {{.TablePrefix}}scores_private;
-- {{else}}
DROP TABLE IF EXISTS {{.TablePrefix}}scores_private;
-- {{end}}
CREATE TABLE {{.TablePrefix}}scores_private (
    id_contest INTEGER NOT NULL,
    id_problem INTEGER NOT NULL,
    id_user INTEGER NOT NULL,
    score INTEGER NOT NULL DEFAULT 0,
    accepted_time INTEGER NOT NULL DEFAULT 0,
    penalty_time INTEGER NOT NULL DEFAULT 0,
    submission_count INTEGER NOT NULL DEFAULT 0,
    one_hit INTEGER NOT NULL DEFAULT 0,
    regraded INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (id_contest, id_problem, id_user)
);

-- Contest score for public
-- {{if eq .Driver "sqlserver"}}
IF OBJECT_ID('{{.TablePrefix}}scores_public', 'U') IS NOT NULL DROP TABLE {{.TablePrefix}}scores_public;
-- {{else}}
DROP TABLE IF EXISTS {{.TablePrefix}}scores_public;
-- {{end}}
CREATE TABLE {{.TablePrefix}}scores_public (
    id_contest INTEGER NOT NULL,
    id_problem INTEGER NOT NULL,
    id_user INTEGER NOT NULL,
    score INTEGER NOT NULL DEFAULT 0,
    accepted_time INTEGER NOT NULL DEFAULT 0,
    penalty_time INTEGER NOT NULL DEFAULT 0,
    submission_count INTEGER NOT NULL DEFAULT 0,
    one_hit INTEGER NOT NULL DEFAULT 0,
    regraded INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (id_contest, id_problem, id_user)
);

-- Slaves list
-- {{if eq .Driver "sqlserver"}}
IF OBJECT_ID('{{.TablePrefix}}slaves', 'U') IS NOT NULL DROP TABLE {{.TablePrefix}}slaves;
-- {{else}}
DROP TABLE IF EXISTS {{.TablePrefix}}slaves;
-- {{end}}
CREATE TABLE {{.TablePrefix}}slaves (
    id INTEGER PRIMARY KEY {{.AutoIncrement}},
    name VARCHAR(200) NOT NULL DEFAULT 'Unnamed',
    address VARCHAR(200) NOT NULL,
    enable INTEGER NOT NULL DEFAULT 1
);
-- Insert default slave
INSERT INTO {{.TablePrefix}}slaves (name, address, enable)
    VALUES ('Localhost Slave', 'localhost:28499', 1);

-- Notifications
-- {{if eq .Driver "sqlserver"}}
IF OBJECT_ID('{{.TablePrefix}}notifications', 'U') IS NOT NULL DROP TABLE {{.TablePrefix}}notifications;
-- {{else}}
DROP TABLE IF EXISTS {{.TablePrefix}}notifications;
-- {{end}}
CREATE TABLE {{.TablePrefix}}notifications (
    id INTEGER PRIMARY KEY {{.AutoIncrement}},
    id_user INTEGER NOT NULL,
    id_user_from INTEGER NOT NULL,
    received_time INTEGER NOT NULL,
    has_read INTEGER NOT NULL DEFAULT 0,
    description VARCHAR(200) NOT NULL DEFAULT 'Empty Notification',
    link VARCHAR(100) NOT NULL DEFAULT 'dashboard'
);

-- Login tokens
-- {{if eq .Driver "sqlserver"}}
IF OBJECT_ID('{{.TablePrefix}}tokens', 'U') IS NOT NULL DROP TABLE {{.TablePrefix}}tokens;
-- {{else}}
DROP TABLE IF EXISTS {{.TablePrefix}}tokens;
-- {{end}}
CREATE TABLE {{.TablePrefix}}tokens (
    token VARCHAR(64) PRIMARY KEY NOT NULL,
    id_user INTEGER NOT NULL,
    login_time INTEGER NOT NULL DEFAULT 0
);
