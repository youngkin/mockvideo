INSERT INTO account (accountHolderName, nickName, serviceAddress, billingAddress, email, phone) 
VALUES ("mickey dolenz", "mickey", "123 Laurel Canyon Drive", "123 Laurel Canyon Drive", "mickeyd@gmail.com", "7132224512");

INSERT INTO account (accountHolderName, nickName, serviceAddress, billingAddress, email, phone) 
VALUES ("cass elliot", "mama cass", "1023 Laurel Canyon Drive", "1023 Laurel Canyon Drive", "mama@gmail.com", "7132224512");

INSERT INTO user (accountID, name, email, role, password) VALUES (1, "mickey dolenz", "mickeyd@gmail.com", 1, "alksdf98423)*(&#");
INSERT INTO user (accountID, name, email, role, password) VALUES (1, "peter tork", "petertd@gmail.com", 3, "alksdf98423)*(&#");
INSERT INTO user (accountID, name, email, role, password) VALUES (1, "davy jones", "djonesI@gmail.com", 3, "alksdf98423)*(&#");
INSERT INTO user (accountID, name, email, role, password) VALUES (1, "michael nesmith", "joanne@gmail.com", 2, "alksdf98423)*(&#");

INSERT INTO user (accountID, name, email, role, password) VALUES (2, "mama cass", "mama@gmail.com", 1, "alksdf98423)*(&#");

INSERT INTO accountUser (accountID, userID) VALUES (1, 1);
INSERT INTO accountUser (accountID, userID) VALUES (1, 2);
INSERT INTO accountUser (accountID, userID) VALUES (1, 3);
INSERT INTO accountUser (accountID, userID) VALUES (1, 4);
INSERT INTO accountUser (accountID, userID) VALUES (2, 5);

# Example query:
# SELECT user.name, user.role 
#   FROM user, accountUser 
#   WHERE user.accountID = accountUser.accountID 
#   AND user.id = accountUser.userID;



INSERT INTO product (prodType, name, price) VALUES (0, "1000mbps", 129.99);
INSERT INTO product (prodType, name, price) VALUES (0, "300mbps", 59.99);
INSERT INTO product (prodType, name, price) VALUES (0, "100mbps", 29.99);
INSERT INTO product (prodType, name, price) VALUES (1, "voice", 29.99);
INSERT INTO product (prodType, name, price) VALUES (2, "ABC", 29.99);
INSERT INTO product (prodType, name, price) VALUES (2, "HBO", 29.99);
INSERT INTO product (prodType, name, price) VALUES (2, "A&E", 29.99);
INSERT INTO product (prodType, name, price) VALUES (2, "telemundo", 29.99);

INSERT INTO bundle (name, price) VALUES ("basic+internet+voice+tv", 99.99);
INSERT INTO bundle (name, price) VALUES ("pro+internet+voice+tv", 199.99);
INSERT INTO bundle (name, price) VALUES ("premium+internet+voice+tv", 299.99);

# basic
INSERT INTO bundleProduct (bundleID, productID) VALUES (1, 3);
INSERT INTO bundleProduct (bundleID, productID) VALUES (1, 4);
INSERT INTO bundleProduct (bundleID, productID) VALUES (1, 5);

# pro
INSERT INTO bundleProduct (bundleID, productID) VALUES (2, 2);
INSERT INTO bundleProduct (bundleID, productID) VALUES (2, 4);
INSERT INTO bundleProduct (bundleID, productID) VALUES (2, 5);
INSERT INTO bundleProduct (bundleID, productID) VALUES (2, 6);
INSERT INTO bundleProduct (bundleID, productID) VALUES (2, 7);

# premium
INSERT INTO bundleProduct (bundleID, productID) VALUES (3, 1);
INSERT INTO bundleProduct (bundleID, productID) VALUES (3, 4);
INSERT INTO bundleProduct (bundleID, productID) VALUES (3, 5);
INSERT INTO bundleProduct (bundleID, productID) VALUES (3, 6);
INSERT INTO bundleProduct (bundleID, productID) VALUES (3, 7);
INSERT INTO bundleProduct (bundleID, productID) VALUES (3, 8);

# Example query:
# SELECT bundle.name AS Bundle, bundle.price AS Price, product.name AS Product 
#   FROM bundle, product, bundleProduct 
#   WHERE bundle.id=bundleProduct.bundleID 
#   AND product.id=bundleProduct.productID