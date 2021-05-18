DROP SCHEMA IF EXISTS dbforum CASCADE;
CREATE EXTENSION IF NOT EXISTS citext;
CREATE SCHEMA dbforum;

CREATE TABLE dbforum.users
(
    id       BIGSERIAL PRIMARY KEY NOT NULL,

    nickname CITEXT UNIQUE         NOT NULL,
    fullname TEXT                  NOT NULL,
    about    TEXT                  NOT NULL,
    email    CITEXT UNIQUE         NOT NULL
);

CREATE TABLE dbforum.forum
(
    id            BIGSERIAL PRIMARY KEY NOT NULL,
    user_nickname CITEXT                NOT NULL,

    title         TEXT                  NOT NULL,
    slug          CITEXT UNIQUE         NOT NULL,
    posts         BIGINT DEFAULT 0      NOT NULL,
    threads       INT    DEFAULT 0      NOT NULL,

    FOREIGN KEY (user_nickname)
        REFERENCES dbforum.users (nickname)
);

CREATE TABLE dbforum.thread
(
    id              BIGSERIAL PRIMARY KEY    NOT NULL,
    forum_slug      CITEXT                   NOT NULL,
    author_nickname CITEXT                   NOT NULL,

    title           TEXT                     NOT NULL,
    message         TEXT                     NOT NULL,
    votes           INT DEFAULT 0            NOT NULL,
    slug            TEXT UNIQUE              NOT NULL,
    created         TIMESTAMP WITH TIME ZONE NOT NULL,

    FOREIGN KEY (forum_slug)
        REFERENCES dbforum.forum (slug),
    FOREIGN KEY (author_nickname)
        REFERENCES dbforum.users (nickname)
);

CREATE TABLE dbforum.votes
(
    nickname  CITEXT        NOT NULL,
    voice     INT DEFAULT 0 NOT NULL,
    thread_id BIGINT        NOT NULL,
    PRIMARY KEY (nickname, thread_id),

    FOREIGN KEY (nickname)
        REFERENCES dbforum.users (nickname),
    FOREIGN KEY (thread_id)
        REFERENCES dbforum.thread (id)
);

CREATE TABLE dbforum.post
(
    id              BIGSERIAL PRIMARY KEY    NOT NULL,
    author_nickname CITEXT                   NOT NULL,
    forum_slug      CITEXT                   NOT NULL,
    thread_id       BIGINT                   NOT NULL,
    message         TEXT                     NOT NULL,

    parent          BIGINT  DEFAULT 0        NOT NULL,
    is_edited       BOOLEAN DEFAULT false    NOT NULL,
    created         TIMESTAMP WITH TIME ZONE NOT NULL,
    tree            TEXT                     NOT NULL,

    FOREIGN KEY (author_nickname)
        REFERENCES dbforum.users (nickname),
    FOREIGN KEY (forum_slug)
        REFERENCES dbforum.forum (slug),
    FOREIGN KEY (thread_id)
        REFERENCES dbforum.thread (id)
);

CREATE TABLE dbforum.forum_users
(
    forum_slug CITEXT        NOT NULL,
    nickname   CITEXT UNIQUE NOT NULL,
    fullname   TEXT          NOT NULL,
    about      TEXT          NOT NULL,
    email      TEXT          NOT NULL,

    FOREIGN KEY (nickname)
        REFERENCES dbforum.users (nickname),
    FOREIGN KEY (forum_slug)
        REFERENCES dbforum.forum (slug),

    PRIMARY KEY (nickname, forum_slug)
);

CREATE OR REPLACE FUNCTION dbforum.insert_forum_user() RETURNS TRIGGER AS
$$
BEGIN
    INSERT INTO dbforum.forum_users(forum_slug, nickname, fullname, about, email)
    SELECT NEW.forum_slug, nickname, fullname, about, email
    FROM dbforum.users
    WHERE nickname = NEW.author_nickname
    ON CONFLICT DO NOTHING;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE OR REPLACE FUNCTION dbforum.update_forum_threads() RETURNS TRIGGER AS
$$
BEGIN
    UPDATE dbforum.forum
    SET threads = threads + 1;
    RETURN NEW;
END
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION dbforum.update_forum_posts() RETURNS TRIGGER AS
$$
BEGIN
    UPDATE dbforum.forum
    SET posts = posts + 1;
    RETURN NEW;
END
$$ LANGUAGE plpgsql;


CREATE TRIGGER thread_insert
    AFTER INSERT
    ON dbforum.thread
    FOR EACH ROW
EXECUTE FUNCTION dbforum.update_forum_threads();

CREATE TRIGGER thread_insert_user_forum
    AFTER INSERT
    ON dbforum.thread
    FOR EACH ROW
EXECUTE FUNCTION dbforum.insert_forum_user();

CREATE TRIGGER post_insert
    AFTER INSERT
    ON dbforum.post
    FOR EACH ROW
EXECUTE FUNCTION dbforum.update_forum_posts();

CREATE TRIGGER post_insert_forum_usert
    AFTER INSERT
    ON dbforum.post
    FOR EACH ROW
EXECUTE FUNCTION dbforum.insert_forum_user();

-- SELECT *
-- FROM dbforum.post
-- WHERE cast(split_part(tree, '.', 1) AS BIGINT) IN
--       (SELECT id FROM dbforum.post WHERE thread_id = $1 AND parent = 0 LIMIT $2)
--   AND CASE WHEN $3 > 0 THEN id > $3 ELSE TRUE END
-- ORDER BY split_part(tree, '.', 1) DESC, tree, id;

--
-- SELECT *
-- FROM dbforum.post
-- WHERE CASE WHEN 1 > 0 THEN id < 228 ELSE TRUE END
-- ORDER BY split_part(tree, '.', 1), tree DESC;
--
-- SELECT *
-- FROM dbforum.post
-- WHERE thread_id = 2
--   and CASE WHEN $1 > 0 THEN id > $2 ELSE TRUE END
-- ORDER BY id
-- LIMIT $3;
--
-- SELECT *
-- FROM dbforum.post
-- WHERE thread_id = 2
--   and CASE WHEN $1 > 0 THEN id < $2 ELSE TRUE END
-- ORDER BY id DESC
-- LIMIT $3;

-- INSERT INTO dbforum.post(author_nickname, forum_slug, thread_id, parent, created, tree)
-- VALUES ('j.sparrow', 'pirate-stories', '2', '0', '2021-03-19T04:18:16.919+03:00',
--         CONCAT('', CAST((SELECT currval(pg_get_serial_sequence('dbforum.post', 'id'))) as text)))
-- RETURNING ID;

-- SELECT * from dbforum.post order by parent

-- SELECT 1
-- FROM dbforum.forum
-- WHERE id = 1
-- LIMIT 1
-- SELECT COUNT(*) FROM dbforum.post

-- INSERT INTO dbforum.forum_users(forum_slug, nickname, fullname, about, email)
-- SELECT 'pirate-stories', nickname, fullname, about, email FROM dbforum.users WHERE nickname = 'j.sparrow'

-- INSERT INTO dbforum.users (nickname, fullname, about, email)
-- VALUES ('kek', 'mem', 'xui', 'kek@mem.ru');
--
-- INSERT INTO dbforum.forum(user_id, title, slug)
-- VALUES (1, 'title', 'uhh');
--
-- INSERT INTO dbforum.forum_users(user_id, forum_slug, nickname, fullname, about, email)
-- values (1, 'uhh', 'aem', 'kess', 'as;ldkjfasdjf', 'alksdjfal;skjf')
--
-- INSERT INTO dbforum.forum_users(user_id, forum_slug, nickname, fullname, about, email)
-- values (3, 'uhh', '0', 'kess', 'as;ldkjfasdjf', 'alksdjfal;skjf')
--
--
-- SELECT fu.nickname, fu.fullname, fu.about, fu.email
-- FROM dbforum.forum_users AS fu
-- WHERE fu.forum_slug = 'uhh'
--   AND fu.nickname > ''
-- ORDER BY fu.nickname DESC
-- LIMIT 2;
--
-- SELECT *
-- FROM dbforum.thread
-- WHERE slug = $1
--   AND created > $2
-- ORDER BY created DESC
-- LIMIT $3

-- INSERT INTO dbforum.users (nickname, fullname, about, email)
-- VALUES ('xit', 'xut', 'xat', 'xit@xit.ru');
-- INSERT INTO dbforum.users (nickname, fullname, about, email)
-- VALUES ('lul', 'zaz', 'tat', 'lul@lul.ru');
--
-- INSERT INTO dbforum.forum(user_id, title, slug)
-- VALUES (1, 'title', 'uhh');
-- INSERT INTO dbforum.forum(user_id, title, slug)
-- VALUES (2, 'kekable', 'xuecable');
--
--
-- INSERT INTO dbforum.thread (forum_id, author_id, title, message, slug, created)
-- VALUES (1, 2, 'thread titel', 'message thread', 'mes slug', '2021-03-19T04:18:16.919+03:00');
-- INSERT INTO dbforum.thread (forum_id, author_id, title, message, slug, created)
-- VALUES (2, 1, 'thread titel 2', 'message thread 2', 'mes slug 2', '2021-03-19T04:18:16.919+04:00');
--
-- INSERT INTO dbforum.post(author_id, forum_id, thread_id, idEdited, created)
-- VALUES (1, 1, 2, false, '2021-03-19T04:18:16.919+05:00');
-- INSERT INTO dbforum.post(author_id, forum_id, thread_id, idEdited, created)
-- VALUES (1, 1, 2, false, '2021-03-19T04:18:16.919+06:00');