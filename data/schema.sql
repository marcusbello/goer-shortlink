-- creating urls table
CREATE TABLE links (
                      id bigserial primary key,
                      short text NOT NULL ,
                      url text NOT NULL )