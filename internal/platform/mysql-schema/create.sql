DROP DATABASE IF EXISTS mockvideo;
CREATE DATABASE mockvideo;

USE mockvideo;

DROP TABLE IF EXISTS customer;
CREATE TABLE customer (
    id INT AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    streetAddress VARCHAR(255),
    city VARCHAR(255),
    state CHAR(2), 
    country VARCHAR(50),
    PRIMARY KEY (id)
)