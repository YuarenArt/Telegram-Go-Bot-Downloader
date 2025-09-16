package repository

const (
	// Table creation
	createTableSubscriptions = `
CREATE TABLE IF NOT EXISTS subscriptions (
    id SERIAL PRIMARY KEY,
    subscription_status TEXT NOT NULL DEFAULT 'inactive',
    duration TEXT NOT NULL DEFAULT 'month',
    start_subscription TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    end_subscription TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);`

	createTableUsers = `
		CREATE TABLE IF NOT EXISTS users (
			username TEXT PRIMARY KEY,
			subscription_id INTEGER NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
			traffic DOUBLE PRECISION DEFAULT 0,
			chat_id BIGINT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);`

	// Indexes
	createUserChatIDIdx         = `CREATE INDEX IF NOT EXISTS idx_users_chat_id ON users(chat_id);`
	createUserSubscriptionIDIdx = `CREATE INDEX IF NOT EXISTS idx_users_subscription_id ON users(subscription_id);`
	createSubscriptionStatusIdx = `CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(subscription_status);`
	createSubscriptionDatesIdx  = `CREATE INDEX IF NOT EXISTS idx_subscriptions_dates ON subscriptions(start_subscription, end_subscription);`

	// Queries
	selectUserSQL = `
		SELECT u.username, u.traffic, u.chat_id, u.created_at, u.updated_at,
			   s.id, s.subscription_status, s.duration, s.start_subscription, s.end_subscription, s.created_at, s.updated_at
		FROM users u
		JOIN subscriptions s ON u.subscription_id = s.id
		WHERE u.username = $1;`

	// Insert subscription and return id
	insertSubscriptionSQL = `
		INSERT INTO subscriptions (subscription_status, duration, start_subscription, end_subscription)
		VALUES ($1, $2, $3, $4)
		RETURNING id;`

	// Insert user (used inside transaction)
	insertUserSQL = `
		INSERT INTO users (username, subscription_id, chat_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (username) DO NOTHING;`

	// Update subscription linked to a user
	updateSubscriptionByUsernameSQL = `
		UPDATE subscriptions
		SET subscription_status = $1, duration = $2, start_subscription = $3, end_subscription = $4, updated_at = NOW()
		FROM users
		WHERE users.username = $5 AND subscriptions.id = users.subscription_id;`

	// Delete user and cascade delete subscription due to FK with ON DELETE CASCADE.
	deleteUserSQL = `DELETE FROM users WHERE username = $1;`

	// Check user existence
	userExistsSQL = `SELECT EXISTS (SELECT 1 FROM users WHERE username = $1);`

	// Update traffic
	updateUserTrafficSQL = `UPDATE users SET traffic = $1, updated_at = NOW() WHERE username = $2;`

	// Get subscription status
	userSubscriptionStatusSQL = `
		SELECT s.subscription_status
		FROM users u
		JOIN subscriptions s ON u.subscription_id = s.id
		WHERE u.username = $1;`

	// Get all usernames
	allUsernamesSQL = `SELECT username FROM users ORDER BY created_at DESC;`

	// Cleanup unused subscriptions: deletes subscriptions not referenced by any user.
	cleanupUnusedSubscriptionsSQL = `
		DELETE FROM subscriptions s
		WHERE NOT EXISTS (
		  SELECT 1 FROM users u WHERE u.subscription_id = s.id
		);`
)
