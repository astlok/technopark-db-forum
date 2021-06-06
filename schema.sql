ALTER USER postgres WITH ENCRYPTED PASSWORD 'admin';
ALTER SYSTEM SET checkpoint_completion_target = '0.9';
ALTER SYSTEM SET wal_buffers = '6912kB';
ALTER SYSTEM SET default_statistics_target = '100';
ALTER SYSTEM SET random_page_cost = '1.1';
ALTER SYSTEM SET effective_io_concurrency = '200';
ALTER SYSTEM SET seq_page_cost = '0.1';
ALTER SYSTEM SET random_page_cost = '0.1';
ALTER SYSTEM SET max_worker_processes = '4';
ALTER SYSTEM SET max_parallel_workers_per_gather = '2';
ALTER SYSTEM SET max_parallel_workers = '4';
ALTER SYSTEM SET max_parallel_maintenance_workers = '2';
DROP SCHEMA IF EXISTS dbforum CASCADE;
CREATE EXTENSION IF NOT EXISTS citext;
CREATE SCHEMA dbforum;

CREATE UNLOGGED TABLE dbforum.users
(
    id       BIGSERIAL PRIMARY KEY NOT NULL,

    nickname CITEXT UNIQUE         NOT NULL,
    fullname TEXT                  NOT NULL,
    about    TEXT                  NOT NULL,
    email    CITEXT UNIQUE         NOT NULL
);

create index user_nickname_idx on dbforum.users (nickname);
create index user_email_idx on dbforum.users (email);

CREATE UNLOGGED TABLE dbforum.forum
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

create index forum_slug_idx on dbforum.forum (slug);

CREATE UNLOGGED TABLE dbforum.thread
(
    id              BIGSERIAL PRIMARY KEY    NOT NULL,
    forum_slug      CITEXT                   NOT NULL,
    author_nickname CITEXT                   NOT NULL,

    title           TEXT                     NOT NULL,
    message         TEXT                     NOT NULL,
    votes           INT DEFAULT 0            NOT NULL,
    slug            citext UNIQUE,
    created         TIMESTAMP WITH TIME ZONE NOT NULL,

    FOREIGN KEY (forum_slug)
        REFERENCES dbforum.forum (slug),
    FOREIGN KEY (author_nickname)
        REFERENCES dbforum.users (nickname)
);
create index thread_forum_slug_idx on dbforum.thread (forum_slug);
create index thread_slug_id_forum_slug_idx on dbforum.thread (slug, id, forum_slug);
create index thread_slug_idx on dbforum.thread (slug);
create index thread_created_idx on dbforum.thread (created);

CREATE UNLOGGED TABLE dbforum.votes
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

create index votes_thread_id_nickname_voice_idx on dbforum.votes (thread_id, nickname, voice);

CREATE UNLOGGED TABLE dbforum.post
(
    id              BIGSERIAL PRIMARY KEY               NOT NULL,
    author_nickname CITEXT                              NOT NULL,
    forum_slug      CITEXT                              NOT NULL,
    thread_id       BIGINT                              NOT NULL,
    message         TEXT                                NOT NULL,

    parent          BIGINT   DEFAULT 0                  NOT NULL,
    is_edited       BOOLEAN  DEFAULT false              NOT NULL,
    created         TIMESTAMP WITH TIME ZONE            NOT NULL,
    tree            BIGINT[] DEFAULT ARRAY []::BIGINT[] NOT NULL,

    FOREIGN KEY (author_nickname)
        REFERENCES dbforum.users (nickname),
    FOREIGN KEY (forum_slug)
        REFERENCES dbforum.forum (slug),
    FOREIGN KEY (thread_id)
        REFERENCES dbforum.thread (id)
);

create index posts_thread_id_parent_idx on dbforum.post (thread_id, parent);
create index posts_tree_1_id_idx on dbforum.post ((tree[1]), id);
create index posts_tree_1_desc_tree_id_idx on dbforum.post ((tree[1]) DESC, tree, id);
create index posts_tree_id_idx on dbforum.post (tree, id);
create index posts_tree_idx on dbforum.post using gin (tree);

CREATE UNLOGGED TABLE dbforum.forum_users
(
    forum_slug CITEXT NOT NULL,
    nickname   CITEXT NOT NULL,
    fullname   TEXT   NOT NULL,
    about      TEXT   NOT NULL,
    email      TEXT   NOT NULL,

    FOREIGN KEY (nickname)
        REFERENCES dbforum.users (nickname),
    FOREIGN KEY (forum_slug)
        REFERENCES dbforum.forum (slug),

    PRIMARY KEY (nickname, forum_slug)
);
create index forum_users_forum_slug_idx on dbforum.forum_users (forum_slug);

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
    SET threads = threads + 1
    WHERE slug = NEW.forum_slug;
    RETURN NEW;
END
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION dbforum.update_forum_posts() RETURNS TRIGGER AS
$$
BEGIN
    NEW.tree = (SELECT tree FROM dbforum.post WHERE id = NEW.parent LIMIT 1) || NEW.ID;
    UPDATE dbforum.forum
    SET posts = posts + 1
    WHERE slug = NEW.forum_slug;
    RETURN NEW;
END
$$ LANGUAGE plpgsql;


CREATE OR REPLACE FUNCTION dbforum.insert_thread_vote() RETURNS TRIGGER AS
$$
BEGIN
    UPDATE dbforum.thread SET votes=(votes + NEW.voice) WHERE id = NEW.thread_id;
    RETURN NEW;
END
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION dbforum.update_thread_vote() RETURNS TRIGGER AS
$$
BEGIN
    IF NEW.voice > 0 THEN
        UPDATE dbforum.thread SET votes=(votes + 2) WHERE id = NEW.thread_id;
    ELSE
        UPDATE dbforum.thread SET votes=(votes - 2) WHERE id = NEW.thread_id;
    END IF;
    RETURN NEW;
END
$$ LANGUAGE plpgsql;

CREATE TRIGGER insert_voice
    AFTER INSERT
    ON dbforum.votes
    FOR EACH ROW
EXECUTE FUNCTION dbforum.insert_thread_vote();


CREATE TRIGGER update_voice
    AFTER UPDATE
    ON dbforum.votes
    FOR EACH ROW
EXECUTE FUNCTION dbforum.update_thread_vote();


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
    BEFORE INSERT
    ON dbforum.post
    FOR EACH ROW
EXECUTE FUNCTION dbforum.update_forum_posts();

CREATE TRIGGER post_insert_forum_usert
    AFTER INSERT
    ON dbforum.post
    FOR EACH ROW
EXECUTE FUNCTION dbforum.insert_forum_user();

