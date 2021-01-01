## GO Task Manager

Go task manager with redis for authentication and PostgresQL as storage database. Users and have multiple tasks and tasks can have multiple users.

## Requirements

- [Go](https://golang.org) - v1.11 above
- Redis
- PostgreSQL

### Endpoints

- `POST` `/signup` - Signup
- `GET` `/signin` - Signin
- `POST` `/logout` - Logout of a session
- `GET` `/tokens` - Generate new access and refresh tokens
- `POST` `/tasks` - Create a task
- `GET` `/tasks` - Get all tasks
- `GET` `/tasks/{taskID}` - Get a single task
- `POST` `/tasks/{taskID}/{userID}` - Add user to task
- `DELETE` `/tasks/{taskID}/{userID}` - Remove user from task
- `PATCH` `/tasks/{id}` - Update task
- `DELETE` `/tasks/{id}` - Delete task
