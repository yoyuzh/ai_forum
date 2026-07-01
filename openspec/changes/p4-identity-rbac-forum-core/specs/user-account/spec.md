## ADDED Requirements

### Requirement: User registration
The system SHALL allow a user to register with a unique username and a password. Passwords SHALL be stored bcrypt-hashed, never plaintext.

#### Scenario: Successful registration
- **WHEN** a user registers with a new username and a valid password
- **THEN** a user row is created with a bcrypt hash and the request returns 201

#### Scenario: Duplicate username rejected
- **WHEN** a user registers with an existing username
- **THEN** the request returns 409 and no new row is created

### Requirement: Login issues JWT
The system SHALL issue a JWT (secret from config, expiry per §14.2) on valid credentials. Incorrect credentials return 401.

#### Scenario: Valid login
- **WHEN** a user submits correct username and password
- **THEN** a signed JWT is returned

#### Scenario: Invalid login
- **WHEN** a user submits a wrong password
- **THEN** the request returns 401 and no token is issued
