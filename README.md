# ChirpyGo API

Chirpy is a lightweight Twitter-style microblogging API server written in Go. It allows users to register, log in, post short messages (called “chirps”), and interact with a secure, token-based authentication system. Built with simplicity, clarity, and testability in mind, Chirpy is ideal for learning backend development with Go.

## Motivation

Chirpy was born out of the desire to build a real-world API with Go while applying clean architecture principles and best practices such as modular package organization, environment-based configuration, and secure JWT-based authentication.

The goal was to create a modern, minimal, and extensible backend that could serve as a foundation for any microblogging or social media app, while remaining small enough to understand and improve upon as a developer.

## Getting Started

### Installing

Ensure you have a working Go environment set up. If not, refer to the official Go installation guide.

### Run the Server
Set environment variables (you can use a .env file) and then run:


## Features

### User Management

- `POST /api/users` – Create a new user
- `PUT /api/users` – Update an existing user (requires auth)
- `POST /api/login` – Login and receive a JWT access token

### Chirps (Posts)

- `POST /api/chirps` – Post a new chirp (short message, max 140 characters)
- `GET /api/chirps` – Retrieve all chirps or by user
- `DELETE /api/chirps/{id}` – Delete a chirp (must be owner)

### Admin & Metrics

- `GET /admin/metrics` – View server usage stats
- `POST /admin/reset` – Reset usage counters

### Webhooks

- `POST /api/polka/webhooks` – Accept external events from Polka (e.g. subscription status updates)

### JWT-based Auth

All protected routes require a valid Authorization: Bearer <token> header. JWTs include user ID in the subject and expiration info, and are validated on every request.

## Improvement Ideas

- Add pagination to GET /chirps
- Introduce likes or replies to chirps
- Web UI frontend (React, Svelte, etc.)
- Token refresh mechanism
- Rate limiting for API abuse prevention
- Admin dashboard for moderation
