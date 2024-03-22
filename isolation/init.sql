CREATE DATABASE isolation_demo;
USE isolation_demo;

CREATE TABLE test_data
(
    id    INT PRIMARY KEY,
    value VARCHAR(50)
);

INSERT INTO test_data (id, value)
VALUES (1, 'Initial Value');
