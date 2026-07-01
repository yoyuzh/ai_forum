# Tech Stack

- Backend: Go, Gin, GORM, Viper, zap, Wire, swaggo/swag, golang-migrate.
- Data/infrastructure: MySQL primary store, Redis cache/limits/hot-score counters/Asynq broker, RabbitMQ domain event bus, Asynq task queue, Elasticsearch search read model.
- Frontend web: React + TypeScript + Vite, Ant Design, TanStack Query, Zustand, React Router, React Virtuoso, Tiptap or mentions input, DOMPurify.
- Admin: React + TypeScript + Refine + Ant Design, TanStack Query, React Router.
- Deployment: Docker Compose for local/v1.0 orchestration; Nginx should not expose `/internal/**`.