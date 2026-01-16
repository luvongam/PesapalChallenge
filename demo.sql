CREATE TABLE users (id INT PRIMARY KEY, name STRING NOT NULL, email STRING UNIQUE, age INT)
INSERT INTO users (id name email age) VALUES (1 'Alice' 'alice@example.com' 30)
INSERT INTO users (id name email age) VALUES (2 'Bob' 'bob@example.com' 25)
INSERT INTO users (id name email age) VALUES (3 'Charlie' 'charlie@example.com' 35)
SELECT * FROM users

CREATE TABLE orders (id INT PRIMARY KEY, user_id INT NOT NULL, product STRING NOT NULL, amount FLOAT)
INSERT INTO orders (id user_id product amount) VALUES (1 1 'Laptop' 999.99)
INSERT INTO orders (id user_id product amount) VALUES (2 1 'Mouse' 29.99)
INSERT INTO orders (id user_id product amount) VALUES (3 2 'Keyboard' 79.99)
INSERT INTO orders (id user_id product amount) VALUES (4 3 'Monitor' 299.99)
SELECT * FROM orders
SELECT * FROM orders WHERE user_id = 1

SELECT * FROM users JOIN orders ON users.id = orders.user_id
