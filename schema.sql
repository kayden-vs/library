CREATE TABLE users (
    id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    hashed_password CHAR(60) NOT NULL,
    created DATETIME NOT NULL,
    role ENUM('student', 'librarian', 'admin') NOT NULL DEFAULT 'student',
    CONSTRAINT users_uc_email UNIQUE (email)
);

CREATE TABLE sessions (
    token CHAR(43) PRIMARY KEY,
    data BLOB NOT NULL,
    expiry TIMESTAMP(6) NOT NULL
);

CREATE INDEX sessions_expiry_idx ON sessions (expiry);

CREATE TABLE books (
    id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(255) NOT NULL,
    author VARCHAR(255) NOT NULL,
    isbn VARCHAR(20) NOT NULL,
    total_copies INTEGER NOT NULL DEFAULT 1,
    available_copies INTEGER NOT NULL DEFAULT 1,
    created DATETIME NOT NULL,
    CONSTRAINT books_uc_isbn UNIQUE (isbn)
);

-- tracks book issues
CREATE TABLE issues (
    id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    book_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    issued_at DATETIME NOT NULL,
    due_date DATETIME NOT NULL,
    returned_at DATETIME,
    FOREIGN KEY (book_id) REFERENCES books(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- alter existing users table to add role (run this if table already exists)
-- ALTER TABLE users ADD COLUMN role ENUM('student', 'librarian', 'admin') NOT NULL DEFAULT 'student';
