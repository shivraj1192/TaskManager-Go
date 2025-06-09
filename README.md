# Task Manager API

A robust backend API for managing users, teams, tasks, comments, labels, attachments, and notifications. Built with **Go**, **Gin**, and **GORM**.

---

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Project Structure](#project-structure)
- [Prerequisites](#prerequisites)
- [Getting Started](#getting-started)
  - [1. Clone the Repository](#1-clone-the-repository)
  - [2. Install Go Dependencies](#2-install-go-dependencies)
  - [3. Configure Environment Variables](#3-configure-environment-variables)
  - [4. Set Up PostgreSQL Database](#4-set-up-postgresql-database)
  - [5. Run Database Migrations](#5-run-database-migrations)
  - [6. Start the Server](#6-start-the-server)
- [API Structure](#api-structure)
- [Database Models](#database-models)
- [Development Notes](#development-notes)
- [Troubleshooting](#troubleshooting)
- [Further Reading](#further-reading)
- [Manual SQL & Foreign Key Troubleshooting](#manual-sql--foreign-key-troubleshooting-optional)

---

## Overview

Task Manager API is a backend service for collaborative task management. It supports user authentication, team collaboration, task assignment, threaded comments, labels, file attachments, and notifications.

---

## Features

- **User Management:** Registration, login, JWT authentication, profile, and role-based access (Admin/Member)
- **Team Management:** Create teams, add/remove members, transfer ownership
- **Task Management:** Create tasks, assign users, track status, parent/subtask hierarchy
- **Comments:** Threaded replies on tasks
- **Labels:** Create, assign, and manage labels for tasks
- **Attachments:** Upload and manage files for tasks
- **Notifications:** Receive notifications for user actions
- **Admin Endpoints:** Advanced management for admins
- **Docker Compose:** Easy PostgreSQL setup

---

## Project Structure

```
.
├── .env
├── docker-compose.yaml
├── go.mod
├── go.sum
├── README.md
├── cmd/
│   └── main.go
├── config/
│   └── connect.go
├── controller/
│   ├── Home.go
│   ├── attachmentController/
│   ├── commentController/
│   ├── labelController/
│   ├── notificationController/
│   ├── taskController/
│   ├── teamController/
│   └── userController/
├── middleware/
│   └── AuthMiddleware.go
├── model/
│   ├── User.go
│   ├── Team.go
│   ├── Task.go
│   ├── Comment.go
│   ├── Notification.go
│   ├── Label.go
│   └── Attachment.go
├── routes/
│   └── setUpRoutes.go
└── static/
    └── file/
```

---

## Prerequisites

- [Go](https://go.dev/doc/install) (v1.16 or later)
- [Git](https://git-scm.com/downloads)
- [PostgreSQL](https://www.postgresql.org/download/) (or use Docker)
- [Docker](https://docs.docker.com/get-docker/) & [Docker Compose](https://docs.docker.com/compose/) (optional, recommended)

---

## Getting Started

### 1. Clone the Repository

```sh
git clone https://github.com/shivraj1192/TaskManager-Go.git
cd task-manager
```

### 2. Install Go Dependencies

run:

```sh
go mod tidy
```

**Otherwise**, If you are starting from scratch (for learning or troubleshooting), you can delete `go.mod` and `go.sum` and re-initialize:

```sh
# Delete old files
rm go.mod go.sum         # Linux/Mac
del go.mod go.sum        # Windows

# Initialize a new Go module
go mod init task-manager

# Download dependencies
go mod tidy
``` 

### 3. Configure Environment Variables

Copy the example `.env` file or create your own:

```env
PORT=8080
SECRET_KEY=your_secret_key
```

- `PORT`: The port your API will run on (default: 8080)
- `SECRET_KEY`: Used for JWT signing (choose a strong, random value)

### 4. Set Up PostgreSQL Database

#### Option A: Using Docker (`Recommended`)

Start PostgreSQL using Docker Compose:

```sh
docker-compose up -d
```

This uses the configuration in [docker-compose.yaml](docker-compose.yaml).

**Otherwise,** run:
```sh
docker run --name postgres-taskmanager -e POSTGRES_USER=user -e POSTGRES_PASSWORD=user123 -e POSTGRES_DB=taskmanager -p 5432:5432 -d postgres
```

#### Option B: Local PostgreSQL

- Install PostgreSQL and create a database named `taskmanager`.
- Update credentials in `.env` and [config/connect.go](config/connect.go) if needed.

##### **OS-specific PostgreSQL Installation**

- **Windows:** [Windows Installer](https://www.postgresql.org/download/windows/)
- **Linux (Ubuntu):**
    ```sh
    sudo apt update
    sudo apt install postgresql postgresql-contrib
    ```
- **Mac:** [Postgres.app](https://postgresapp.com/) or Homebrew:
    ```sh
    brew install postgresql
    ```

##### **Create Database**

```sh
# Login to psql (may require 'sudo -u postgres psql' on Linux)
psql -U postgres
CREATE DATABASE taskmanager;
\q
```

### 5. Run Database Migrations

Tables are auto-created on server start using GORM's auto-migrate feature.

> **Note:**  
> If you prefer manual SQL setup or want to inspect the schema, refer to  
> [`config/connect.go`](config/connect.go) for auto-migrate logic.


### 6. Start the Server

run:
```sh
cd ./cmd/
air
```

**Otherwise,** run
```sh
go run ./cmd/main.go
```

The API will be available at [http://localhost:8080/](http://localhost:8080/).

---

## API Structure

- **Base URL:** `http://localhost:8080/`
- **Authentication:** JWT Bearer token in `Authorization` header

### Main Endpoints

- `POST /api/register` – Register a new user ([userController/Register.go](controller/userController/Register.go))
- `POST /api/login` – Login and receive JWT token ([userController/Login.go](controller/userController/Login.go))

#### Protected Endpoints (require JWT)

- `/api/users` – User details, update, delete, change password, admin can manage all users ([userController/](controller/userController/))
- `/api/teams` – Team management ([teamController/](controller/teamController/))
- `/api/tasks` – Task management ([taskController/](controller/taskController/))
- `/api/comments` – Manage comments ([commentController/](controller/commentController/))
- `/api/labels` – Admin label management ([labelController/](controller/labelController/))
- `/api/attachments` – Manage task attachments ([attachmentController/](controller/attachmentController/))
- `/api/notifications` – View notifications (if implemented)

See [routes/setUpRoutes.go](routes/setUpRoutes.go) for the full API routing.

---

## Database Models

- **User:** [model/User.go](model/User.go)
- **Team:** [model/Team.go](model/Team.go)
- **Task:** [model/Task.go](model/Task.go)
- **Comment:** [model/Comment.go](model/Comment.go)
- **Notification:** [model/Notification.go](model/Notification.go)
- **Label:** [model/Label.go](model/Label.go)
- **Attachment:** [model/Attachment.go](model/Attachment.go)

---

## Development Notes

- Environment variables loaded from `.env` ([cmd/main.go](cmd/main.go))
- Auto-migration on startup ([config/connect.go](config/connect.go))
- Uses [Gin](https://gin-gonic.com/docs/) for HTTP routing and middleware
- Uses [GORM](https://gorm.io/docs/) for ORM/database access
- Attachments are stored in `static/file/` directory
- Notifications are created for most user actions

---

## Troubleshooting

- **Port Already in Use:** Change the `PORT` in `.env`.
- **Database Connection Issues:** Check credentials in `.env` and [config/connect.go](config/connect.go).
- **JWT Errors:** Ensure `SECRET_KEY` is set and matches in both `.env` and your environment.
- **Auto-migration Issues:** Check [config/connect.go](config/connect.go).

---

## Further Reading

- [Gin Documentation](https://gin-gonic.com/docs/)
- [GORM Documentation](https://gorm.io/docs/)
- [JWT Introduction](https://jwt.io/introduction)
- [Go Modules](https://blog.golang.org/using-go-modules)
- [Docker Compose Overview](https://docs.docker.com/compose/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)

---

## Manual SQL & Foreign Key Troubleshooting (Optional)

If you encounter **foreign key constraint errors** (for example, after changing your models or relationships), you may need to manually drop and re-add foreign key constraints in PostgreSQL.

### Step 1: Open PostgreSQL Command Line

#### Windows

- Open **pgAdmin** and use the Query Tool, or
- Open **Command Prompt** and run:
  ```sh
  psql -U your_username -d taskmanager
  ```
  (Replace `your_username` with your PostgreSQL username.)

#### Linux

- Open Terminal and run:
  ```sh
  sudo -u postgres psql -d taskmanager
  ```
  or, if you have a user:
  ```sh
  psql -U your_username -d taskmanager
  ```

#### Mac

- If using **Postgres.app**, open the app and click the database, then open a new SQL window.
- If using Homebrew:
  ```sh
  psql -U your_username -d taskmanager
  ```

### Step 2: Run Foreign Key Fix Queries

Paste and run these queries **one by one** in your SQL prompt to drop and re-add foreign key constraints for all main models:

```sql
-- For tasks table (team_id foreign key)
ALTER TABLE tasks DROP CONSTRAINT IF EXISTS fk_teams_tasks;
ALTER TABLE tasks
  ADD CONSTRAINT fk_teams_tasks FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE;

-- For comments table (task_id foreign key)
ALTER TABLE comments DROP CONSTRAINT IF EXISTS fk_tasks_comments;
ALTER TABLE comments
  ADD CONSTRAINT fk_tasks_comments FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE;

-- For attachments table (task_id foreign key)
ALTER TABLE attachments DROP CONSTRAINT IF EXISTS fk_tasks_attachments;
ALTER TABLE attachments
  ADD CONSTRAINT fk_tasks_attachments FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE;

-- For notifications table (user_id foreign key)
ALTER TABLE notifications DROP CONSTRAINT IF EXISTS fk_users_notifications;
ALTER TABLE notifications
  ADD CONSTRAINT fk_users_notifications FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- For labels table (task_id foreign key, if exists)
ALTER TABLE labels DROP CONSTRAINT IF EXISTS fk_tasks_labels;
ALTER TABLE labels
  ADD CONSTRAINT fk_tasks_labels FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE;
```

> **Tip:** If you get an error about a missing constraint, that's OK—`DROP CONSTRAINT IF EXISTS` will skip it.