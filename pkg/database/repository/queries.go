package repository

// SQL statements - consistent with models (no subscription_status)
const (
	createTableSubscriptions = `
		CREATE TABLE IF NOT EXISTS subscriptions (
			id SERIAL PRIMARY KEY,
			duration TEXT NOT NULL DEFAULT 'month',
			start_subscription TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			end_subscription TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			-- enforce end > start at DB level
			CHECK (end_subscription > start_subscription)
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

	createUserChatIDIdx         = `CREATE INDEX IF NOT EXISTS idx_users_chat_id ON users(chat_id);`
	createUserSubscriptionIDIdx = `CREATE INDEX IF NOT EXISTS idx_users_subscription_id ON users(subscription_id);`
	createSubscriptionEndIdx    = `CREATE INDEX IF NOT EXISTS idx_subscriptions_end ON subscriptions(end_subscription);`

	selectUserSQL = `
		SELECT u.username, u.traffic, u.chat_id, u.created_at, u.updated_at,
			   s.id, s.duration, s.start_subscription, s.end_subscription, s.created_at, s.updated_at
		FROM users u
		JOIN subscriptions s ON u.subscription_id = s.id
		WHERE u.username = $1;`

	insertSubscriptionSQL = `
		INSERT INTO subscriptions (duration, start_subscription, end_subscription)
		VALUES ($1, $2, $3)
		RETURNING id;`

	insertUserSQL = `
		INSERT INTO users (username, subscription_id, chat_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (username) DO NOTHING
		RETURNING username;`

	updateSubscriptionByUsernameSQL = `
		UPDATE subscriptions
		SET duration = $1, start_subscription = $2, end_subscription = $3, updated_at = NOW()
		FROM users
		WHERE users.username = $4 AND subscriptions.id = users.subscription_id;`

	deleteUserSQL = `DELETE FROM users WHERE username = $1;`

	userExistsSQL = `SELECT EXISTS (SELECT 1 FROM users WHERE username = $1);`

	updateUserTrafficSQL = `UPDATE users SET traffic = $1, updated_at = NOW() WHERE username = $2;`

	userSubscriptionActiveSQL = `
		SELECT (s.end_subscription > NOW()) AS active
		FROM users u
		JOIN subscriptions s ON u.subscription_id = s.id
		WHERE u.username = $1;`

	allUsernamesSQL = `SELECT username FROM users ORDER BY created_at DESC;`

	cleanupUnusedSubscriptionsSQL = `
		DELETE FROM subscriptions s
		WHERE NOT EXISTS (
		  SELECT 1 FROM users u WHERE u.subscription_id = s.id
		);`

	userSubscriptionStatusSQL = `
		SELECT CASE WHEN s.end_subscription > NOW() THEN 'active' ELSE 'inactive' END AS status
		FROM users u
		JOIN subscriptions s ON u.subscription_id = s.id
		WHERE u.username = $1
		LIMIT 1;`
)
