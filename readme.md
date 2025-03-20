# Bad Jokes Web Application

A web application for sharing and interacting with jokes. Users can post jokes, comment, vote, and interact with content through a modern web interface.

## Technology Stack

- **Backend**: Go
- **Frontend**: React
- **Database**: PostgreSQL,SQLlite
- **Containerization**: Docker and Docker Compose
- **Authentication**: JWT, Google OAuth, GitHub OAuth

## Project Structure

```
.
├── api/                # Go backend service
├── frontend/           # React frontend
├── migrations/         # SQL database migrations
├── create_env.sh       # Environment setup script
├── docker-compose.yml  # Docker services configuration
└── .env                # Environment variables (created by setup script)
```

## Setup Instructions

### Prerequisites

- Docker and Docker Compose
- Git

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/russianssamovar/bad-jokes.git
   cd bad-jokes
   ```

2. Create environment file:
   ```bash
   chmod +x create_env.sh
   ./create_env.sh
   ```
   Follow the prompts to configure your environment variables.

3. Start the application:
   ```bash
   docker-compose up --build -d
   ```

4. Access the application:
    - Frontend: http://localhost:8080
    - API: http://localhost:9999

## Database Migrations

The application uses SQL migrations for database setup. Migrations are executed in numerical order (e.g., 001_create_tables.sql executes before 1000_seed_database.sql).

## API Endpoints

The API is available at `/api/` on the frontend server or directly at port 9999.

## Development

For local development:

1. Set up environment variables using `create_env.sh`
2. Run `docker-compose up -d db` to start only the database
3. Run backend and frontend separately in development mode

## Authentication

The application supports authentication via:
- Email/Password
- Google OAuth
- GitHub OAuth

## License

Apache 2.0
