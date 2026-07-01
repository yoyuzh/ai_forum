## ADDED Requirements

### Requirement: Config struct mirrors architecture §14.2
The `Config` struct SHALL expose fields for `Server{Port,Mode}`, `MySQL{Host,Port,Username,Password,Database}`, `Redis{Addr,Password,DB}`, `RabbitMQ{URL}`, `Elasticsearch{Addresses}`, `JWT{Secret,ExpireHours}`, `InternalAPI{Token}`, `AI{Provider,Model,APIKey,MaxConcurrency,RequestPerSecond,Burst}`, `Worker{AiReplyConcurrency,TaggingConcurrency,SearchIndexConcurrency,NotificationConcurrency}`, `HotScore{RefreshIntervalSeconds,BatchSize}`, and `Log{Level,Encoding}`.

#### Scenario: Dev config file parses
- **WHEN** `Load("config/config.dev.yaml")` is called with all `${VAR}` placeholders resolved via environment
- **THEN** the returned `*Config` populates every field above with non-zero values

### Requirement: Config precedence is env > file > default
The loader SHALL apply defaults first, overlay values from the YAML file, then overlay environment variables on top. Only an explicit allowlist of env keys (e.g. `MYSQL_PASSWORD` → `mysql.password`, `INTERNAL_API_TOKEN` → `internal_api.token`, `JWT_SECRET` → `jwt.secret`) SHALL override file values.

#### Scenario: Env overrides file value
- **WHEN** the YAML file sets `mysql.password: filevalue` and the env var `MYSQL_PASSWORD=envvalue` is set
- **THEN** `Load` returns `Config.MySQL.Password == "envvalue"`

#### Scenario: Default applied when absent
- **WHEN** neither the YAML file nor the environment sets `server.port`
- **THEN** `Load` returns `Config.Server.Port == 8080`

### Requirement: Required secrets validated at startup
`Validate(*Config)` SHALL return a non-nil error when `JWT.Secret` or `InternalAPI.Token` is empty, regardless of mode. In `server.mode != debug`, it SHALL also fail when `MySQL.Password` is empty. The error SHALL list every missing key in one aggregate message.

#### Scenario: Missing required secrets fails loud
- **WHEN** `Validate` is called on a config with empty `JWT.Secret` and empty `InternalAPI.Token`
- **THEN** it returns an error whose message contains both `jwt.secret` and `internal_api.token`

#### Scenario: Debug mode relaxes DB password
- **WHEN** `Server.Mode == "debug"` and `MySQL.Password` is empty but `JWT.Secret` and `InternalAPI.Token` are set
- **THEN** `Validate` returns nil

### Requirement: No literal secrets in committed config
The committed `config/config.dev.yaml` SHALL contain `${VAR}` placeholders for every secret field and SHALL NOT contain any literal secret value.

#### Scenario: Repo config has no literal secrets
- **WHEN** `config/config.dev.yaml` is inspected
- **THEN** the `password`, `secret`, `token`, and `api_key` fields contain `${...}` placeholders, not real values
