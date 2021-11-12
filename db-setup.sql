CREATE TABLE IF NOT EXISTS servers
(
    id   smallserial PRIMARY KEY,
    name varchar(20) UNIQUE NOT NULL,

    CHECK (LENGTH(name) > 0)
);

INSERT INTO servers
VALUES (-1, 'Network')
ON CONFLICT DO NOTHING;

INSERT INTO servers
VALUES (-2, 'Website')
ON CONFLICT DO NOTHING;


CREATE TABLE IF NOT EXISTS ranks
(
    id           smallserial PRIMARY KEY,
    level        smallint UNIQUE    NOT NULL,
    name         varchar(30) UNIQUE NOT NULL,
    display_name varchar(75)        NOT NULL,
    chat_format  varchar(200)       NOT NULL,

    CHECK (level > 0),
    CHECK (LENGTH(name) > 0),
    CHECK (LENGTH(display_name) > 0),
    CHECK (LENGTH(chat_format) > 0)
);

INSERT INTO ranks
VALUES (-1, 32767, 'Owner', '&4Owner', '[{{RANK}}] {{NICK}} : {{MSG}}')
ON CONFLICT DO NOTHING;

INSERT INTO ranks
VALUES (-2, 1000, 'Player', '&7Player', '[{{RANK}}] {{NICK}} : {{MSG}}')
ON CONFLICT DO NOTHING;


CREATE TABLE IF NOT EXISTS players
(
    id       serial PRIMARY KEY,
    nick     varchar(16) UNIQUE        NOT NULL,
    password varchar                   NOT NULL,
    email    varchar(100) DEFAULT NULL UNIQUE,
    rank     smallint     DEFAULT 1000 NOT NULL,

    CHECK (nick SIMILAR TO '\w{3,16}'),
    CHECK (email SIMILAR TO '%\S+@\S+.\S+%'),

    FOREIGN KEY (rank) REFERENCES ranks (id)
        ON DELETE SET DEFAULT
);

INSERT INTO players
VALUES (-1, 'CONSOLE', '-', DEFAULT, -1)
ON CONFLICT DO NOTHING;


CREATE TABLE IF NOT EXISTS permissions
(
    rank    smallint     NOT NULL,
    server  smallint     NOT NULL,
    value   varchar(100) NOT NULL,
    negated boolean      NOT NULL,

    PRIMARY KEY (rank, server, value),
    CHECK (LENGTH(value) > 0 AND value SIMILAR TO '[A-Za-z0-9*]+(.[A-Za-z0-9*]+)*'),
    FOREIGN KEY (rank) REFERENCES ranks (id)
        ON DELETE CASCADE
        ON UPDATE CASCADE,
    FOREIGN KEY (server) REFERENCES servers (id)
        ON DELETE CASCADE
);

INSERT INTO permissions
VALUES (-1, -2, 'rank.view', false)
ON CONFLICT DO NOTHING;

INSERT INTO permissions
VALUES (-1, -2, 'rank.modify', false)
ON CONFLICT DO NOTHING;


CREATE TABLE IF NOT EXISTS bans
(
    id         serial PRIMARY KEY,
    server     smallint  NOT NULL,
    recipient  int       NOT NULL,
    giver      int       NOT NULL,
    start      timestamp NOT NULL,
    expiration timestamp NOT NULL,
    reason     varchar   NOT NULL,
    CHECK (start < expiration),
    CHECK (LENGTH(reason) > 0),
    FOREIGN KEY (server) REFERENCES servers (id)
        ON DELETE CASCADE,
    FOREIGN KEY (recipient) REFERENCES players (id)
        ON DELETE CASCADE,
    FOREIGN KEY (giver) REFERENCES players (id)
        ON DELETE CASCADE
);


CREATE TABLE IF NOT EXISTS active_bans
(
    id int PRIMARY KEY,
    FOREIGN KEY (id) REFERENCES bans (id)
        ON DELETE CASCADE
);


CREATE TABLE IF NOT EXISTS old_bans
(
    id                  int PRIMARY KEY,
    actual_expiration   timestamp NOT NULL,
    old_type            char      NOT NULL,
    new_ban             int,
    modder              int DEFAULT -1,
    modification_reason varchar,
    CHECK (old_type IN ('E', 'U', 'M')),
    CHECK (modification_reason IS NULL OR LENGTH(modification_reason) > 0),
    CHECK (
            (old_type = 'E' AND new_ban IS NULL AND modder IS NULL AND modification_reason IS NULL) OR
            (old_type = 'U' AND new_ban IS NULL AND modder IS NOT NULL AND modification_reason IS NOT NULL) OR
            (old_type = 'M' AND new_ban IS NOT NULL AND modder IS NOT NULL AND modification_reason IS NOT NULL)
        ),
    FOREIGN KEY (id) REFERENCES bans (id)
        ON DELETE CASCADE,
    FOREIGN KEY (new_ban) REFERENCES bans (id)
        ON DELETE CASCADE,
    FOREIGN KEY (modder) REFERENCES players (id)
        ON DELETE SET DEFAULT
);


CREATE TYPE permission AS
(
    server  smallint,
    value   varchar(100),
    negated boolean
);


CREATE FUNCTION get_servers()
    RETURNS TABLE
            (
                _id_   smallint,
                _name_ varchar(20)
            )
AS
$$
BEGIN
    RETURN QUERY
        SELECT id, name
        FROM servers;
END
$$ LANGUAGE plpgsql;


CREATE FUNCTION get_ranks()
    RETURNS TABLE
            (
                _id_           smallint,
                _level_        smallint,
                _name_         varchar(30),
                _display_name_ varchar(75),
                _chat_format_  varchar(200),
                _permissions_  permission[]
            )
AS
$$
BEGIN
    RETURN QUERY
        SELECT id,
               level,
               name,
               display_name,
               chat_format,
               ARRAY(
                       SELECT server, value, negated
                       FROM permissions
                       WHERE rank = ranks.id
                   )
        FROM ranks
        ORDER BY level;
END
$$ LANGUAGE plpgsql;


CREATE FUNCTION get_server_ranks(_server smallint)
    RETURNS TABLE
            (
                _id_                  smallint,
                _level_               smallint,
                _name_                varchar(30),
                _display_name_        varchar(75),
                _chat_format_         varchar(200),
                _permissions_         varchar(100)[],
                _negated_permissions_ varchar(100)[]
            )
AS
$$
BEGIN
    RETURN QUERY
        SELECT id,
               level,
               name,
               display_name,
               chat_format,
               ARRAY(
                       SELECT value
                       FROM permissions
                       WHERE rank = ranks.id
                         AND server = _server
                         AND negated = false
                   ),
               ARRAY(
                       SELECT value
                       FROM permissions
                       WHERE rank = ranks.id
                         AND server = _server
                         AND negated = true
                   )
        FROM ranks
        ORDER BY level;
END
$$ LANGUAGE plpgsql;


CREATE FUNCTION get_permissions(_rank smallint)
    RETURNS TABLE
            (
                _server_     smallint,
                _permission_ varchar(100),
                _negated_    boolean
            )
AS
$$
BEGIN
    RETURN QUERY
        SELECT server, value, negated
        FROM permissions
        WHERE rank = _rank;
END
$$ LANGUAGE plpgsql;


CREATE FUNCTION create_rank(_level smallint, _name varchar(30), _display_name varchar(75), _chat_format varchar(200),
                            _permissions permission[])
    RETURNS smallint AS
$$
DECLARE
    _id smallint;
BEGIN
    INSERT INTO ranks
    VALUES (DEFAULT, _level, _name, _display_name, _chat_format)
    RETURNING id INTO _id;

    INSERT INTO permissions
    SELECT _id, server, value, negated
    FROM UNNEST(_permissions) AS x(server, value, negated);

    RETURN _id;
END
$$ LANGUAGE plpgsql;


CREATE FUNCTION remove_rank(_id smallint)
    RETURNS void AS
$$
DECLARE
    _removed bool;
BEGIN
    DELETE
    FROM ranks
    WHERE id = _id
    RETURNING true INTO _removed;

    IF _removed IS NULL THEN
        RAISE no_data_found;
    END IF;
END
$$ LANGUAGE plpgsql;


CREATE FUNCTION modify_rank(_id smallint, _level smallint, _name varchar(30), _display_name varchar(75),
                            _chat_format varchar(200), _removed_permissions permission[],
                            _added_permissions permission[])
    RETURNS void AS
$$
DECLARE
    _modified bool;
BEGIN
    UPDATE ranks
    SET level        = COALESCE(_level, level),
        name         = COALESCE(_name, name),
        display_name = COALESCE(_display_name, display_name),
        chat_format  = COALESCE(_chat_format, chat_format)
    WHERE id = _id
    RETURNING true INTO _modified;

    IF _modified IS NULL THEN
        RAISE no_data_found;
    END IF;

    DELETE
    FROM permissions
    WHERE rank = _id
      AND (server, value, negated) IN
          (SELECT server, value, negated FROM UNNEST(_removed_permissions) AS x(server, value, negated));

    INSERT INTO permissions
    SELECT _id, server, value, negated
    FROM UNNEST(_added_permissions) AS x(server, value, negated);
END
$$ LANGUAGE plpgsql;


CREATE FUNCTION get_auth_data(_nick varchar)
    RETURNS TABLE (_id_ int, _password_ varchar, _rank_ smallint) AS $$
BEGIN
    RETURN QUERY
        SELECT id, password, rank
        FROM players
        WHERE nick = _nick;
END
$$ LANGUAGE plpgsql;
