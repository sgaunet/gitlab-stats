-- migrate:up

CREATE TABLE projects (
    id integer PRIMARY KEY NOT NULL,
    project_name character varying(255) NOT NULL
);

CREATE INDEX projects_id_idx       ON projects (id) ;

CREATE TABLE groups (
    id integer PRIMARY KEY NOT NULL,
    group_name character varying(255) NOT NULL
);

CREATE INDEX groups_id_idx       ON groups (id) ;

CREATE TABLE stats (
    id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
    date_exec timestamp NOT NULL,
    total integer NOT NULL,
    closed integer NOT NULL,
    opened integer NOT NULL
);

CREATE INDEX stats_id_idx       ON stats (id) ;

CREATE TABLE stats_projects (
    id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
    projectId integer NOT NULL,
    statsId UUID NOT NULL,
    CONSTRAINT fk_projectid
      FOREIGN KEY(projectId) 
	  REFERENCES projects(id),
    CONSTRAINT fk_statsid
      FOREIGN KEY(statsId) 
	  REFERENCES stats(id)
);

CREATE INDEX stats_projects_id_idx       ON stats_projects (id) ;

CREATE TABLE stats_groups (
    id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
    groupId integer NOT NULL,
    statsId UUID NOT NULL,
    CONSTRAINT fk_groupid
      FOREIGN KEY(groupId) 
	  REFERENCES groups(id),
    CONSTRAINT fk_statsid
      FOREIGN KEY(statsId) 
	  REFERENCES stats(id)
);

CREATE INDEX stats_groups_id_idx       ON stats_groups (id) ;

-- migrate:down

DROP TABLE projects cascade;

DROP TABLE groups cascade;

DROP TABLE stats cascade;

DROP TABLE stats_projects cascade;

DROP TABLE stats_groups cascade;
