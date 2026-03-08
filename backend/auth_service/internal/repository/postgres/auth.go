package postgres

//TODO:
type Repo interface {
	// Create(user) — INSERT INTO users ...
	// GetByEmail(email) — SELECT * FROM users WHERE email = $1
	// GetByID(id) — SELECT * FROM users WHERE id = $1
	// UpdatePassword(id, hash) — UPDATE users SET password_hash = ...
	// SetEmailVerified(id) — UPDATE users SET is_verified = true
}