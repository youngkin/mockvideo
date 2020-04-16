DROP DATABASE IF EXISTS mockvideo;
CREATE DATABASE mockvideo;

CREATE USER 'admin'@'%' IDENTIFIED BY 'admin';
GRANT ALL   ON *.*   TO 'admin'@'%'   WITH GRANT OPTION;

USE mockvideo;

# user represents an authorized individual on an account
DROP TABLE IF EXISTS user;
CREATE TABLE user (
    accountID INT,
    id INT AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    # 
    # role: 1 - admin, 2 - unrestricted, 3 - restricted
    role INT,
    password VARCHAR(255),
    PRIMARY KEY (id),
    UNIQUE KEY (email)
);

# account is the high level information about a customer
DROP TABLE IF EXISTS account;
CREATE TABLE account (
    id INT AUTO_INCREMENT,
    accountHolderName VARCHAR(255) NOT NULL,
    nickName VARCHAR(255),
    serviceAddress VARCHAR(255) NOT NULL,
    billingAddress VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(10) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY (email)
);

# bundle represents a group of one or more products. 
DROP TABLE IF EXISTS bundle;
CREATE TABLE bundle (
    id INT AUTO_INCREMENT,
    # name - e.g., basic, internet+tv, internet+tv+voice, internet+tv+voice+home+mobile
    name VARCHAR(255) NOT NULL,
    price int NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY (name)
);

# product represents a single purchasable item such as a premium
# television channel like HBO, sports package, internet package (e.g., Gigabit, 300Mbps), 
# mobile, or a home security package
DROP TABLE IF EXISTS product;
CREATE TABLE product (
    id INT AUTO_INCREMENT,
    # prodType: 0 - internet, 1 - phone, 2 - TV, 3 - Home security, 4 - mobile
    prodType INT,

    # name - e.g., Internet (1000mbps, 300mbps, 600mbps), TV (e.g., A&E, Telemundo, HBO), voice (e.g., voice), mobile (e.g., mobile)
    name VARCHAR(255) NOT NULL,
    price DECIMAL(5,2) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY (prodType, name)
);

# accountUser represents the association between an account and
# the authorized users on that account.
DROP TABLE IF EXISTS accountUser;
CREATE TABLE accountUser (
    accountID INT NOT NULL,
    userID INT NOT NULL,
    PRIMARY KEY (accountID, userID)
);

# bundleProduct represents the association between a bundle
# and the products it contains.
DROP TABLE IF EXISTS bundleProduct;
CREATE TABLE bundleProduct (
    bundleID INT NOT NULL,
    productID INT NOT NULL,
    PRIMARY KEY (bundleID, productID)
);