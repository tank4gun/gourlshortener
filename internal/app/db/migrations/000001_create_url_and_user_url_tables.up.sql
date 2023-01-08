CREATE TABLE IF NOT EXISTS url
(
    id serial PRIMARY KEY,
    value varchar(100)  NOT NULL
);
CREATE TABLE IF NOT EXISTS user_url
(
    id serial PRIMARY KEY,
    user_id int NOT NULL,
    url_id int  NOT NULL,
    FOREIGN KEY (url_id) references url(id)
);
