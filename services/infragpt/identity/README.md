# identity

This implements identity functionality with user authentication and authorization.

## Security aspects:

- Password hashing with bcrypt.

- JWT signed with RS256.

- Refresh tokens stored with versioning to allow revocation.

- Access tokens short-lived (15 minutes) and refresh token long-lived(30 days).

- Rate limiting on login and signup endpoints.

- Input validation (email format, password strength).

- Use secure random for JWT IDs in refresh tokens.

- Set appropriate headers (CORS, Content-Type, etc.).

- Error messages don't reveal sensitive info (e.g. don't say "password is incorrect").