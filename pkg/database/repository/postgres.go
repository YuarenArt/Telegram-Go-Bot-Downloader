package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"youtube_downloader/pkg/database"
	"youtube_downloader/pkg/database/models"

	_ "github.com/lib/pq"
)

// Database is a thin wrapper around sql.DB providing application operations.
type Database struct {
	db     *sql.DB
	config *database.Config
}

// Close closes underlying DB connection.
func (d *Database) Close() error {
	if d == nil || d.db == nil {
		return nil
	}
	return d.db.Close()
}

// NewDatabase opens connection to Postgres, sets pool params, runs migrations and returns Database.
// It will attempt multiple retries with backoff to increase resiliency on startup.
func NewDatabase(ctx context.Context, cfg *database.Config) (*Database, error) {
	if cfg == nil {
		cfg = database.DefaultConfig()
	}

	// Try connecting with simple retry/backoff (for containerized environments).
	var db *sql.DB
	var err error
	maxAttempts := 5
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		db, err = sql.Open("postgres", cfg.DSN())
		if err == nil {
			// configure pool
			db.SetMaxOpenConns(cfg.MaxOpenConns)
			db.SetMaxIdleConns(cfg.MaxIdleConns)
			db.SetConnMaxLifetime(5 * time.Minute)

			// ping with timeout
			pctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err = db.PingContext(pctx)
			cancel()
			if err == nil {
				break
			}
		}

		// close if opened
		if db != nil {
			_ = db.Close()
		}

		wait := time.Duration(attempt*500) * time.Millisecond
		log.Printf("database connection attempt %d/%d failed: %v; retrying in %s", attempt, maxAttempts, err, wait)
		select {
		case <-time.After(wait):
			// continue retry
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled while connecting to database: %w", ctx.Err())
		}
	}
	if err != nil {
		return nil, fmt.Errorf("could not connect to database: %w", err)
	}

	repo := &Database{
		db:     db,
		config: cfg,
	}

	// Initialize schema (create tables and indexes).
	if err := repo.initSchema(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	// Cleanup unused subscriptions
	if _, err := db.ExecContext(context.Background(), cleanupUnusedSubscriptionsSQL); err != nil {
		// log but don't fail startup: cleanup is best-effort
		log.Printf("warning: cleanupUnusedSubscriptions failed: %v", err)
	}

	log.Println("Database connection established successfully.")
	return repo, nil
}

// initSchema creates tables and indexes if not exist.
func (d *Database) initSchema(ctx context.Context) error {
	statements := []string{
		createTableSubscriptions,
		createTableUsers,
		createUserChatIDIdx,
		createUserSubscriptionIDIdx,
		createSubscriptionEndIdx,
	}
	for _, s := range statements {
		if _, err := d.db.ExecContext(ctx, s); err != nil {
			return fmt.Errorf("schema exec failed: %w", err)
		}
	}
	return nil
}

// CreateUser creates a subscription and a user in a single transaction.
// Returns error when username invalid or DB operation failed.
func (d *Database) CreateUser(ctx context.Context, user *models.User) error {
	if user == nil {
		return errors.New("user is nil")
	}
	if user.Username == "" {
		return errors.New("username is required")
	}
	// validate subscription dates
	start := user.Subscription.StartSubscription
	if start.IsZero() {
		start = time.Now().UTC()
	}
	end := user.Subscription.EndSubscription
	if end.IsZero() {
		end = start.AddDate(0, 1, 0) // default 1 month
	}
	if !end.After(start) {
		return errors.New("end_subscription must be after start_subscription")
	}

	tx, err := d.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Printf("tx rollback error: %v", err)
		}
	}()

	var subscriptionID int64
	if err := tx.QueryRowContext(ctx, insertSubscriptionSQL,
		defaultString(user.Subscription.Duration, "month"),
		start,
		end,
	).Scan(&subscriptionID); err != nil {
		return fmt.Errorf("insert subscription: %w", err)
	}

	// try to insert user and detect conflict
	var insertedUsername string
	err = tx.QueryRowContext(ctx, insertUserSQL, user.Username, subscriptionID, user.ChatID).Scan(&insertedUsername)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// ON CONFLICT DO NOTHING returned no row -> conflict happened
			return fmt.Errorf("username %s already exists", user.Username)
		}
		return fmt.Errorf("insert user: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// User retrieves user + subscription by username.
func (d *Database) User(ctx context.Context, username string) (*models.User, error) {
	if username == "" {
		return nil, errors.New("username is required")
	}

	row := d.db.QueryRowContext(ctx, selectUserSQL, username)

	var u models.User
	var s models.Subscription
	err := row.Scan(
		&u.Username,
		&u.Traffic,
		&u.ChatID,
		&u.CreatedAt,
		&u.UpdatedAt,
		&s.ID,
		&s.Duration,
		&s.StartSubscription,
		&s.EndSubscription,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("scan user: %w", err)
	}
	u.Subscription = s
	return &u, nil
}

// UpdateUserSubscription updates subscription fields for the given username.
func (d *Database) UpdateUserSubscription(ctx context.Context, username string, newSub models.Subscription) error {
	if username == "" {
		return errors.New("username is required")
	}

	// Check user exists
	exists, err := d.IsUserExists(ctx, username)
	if err != nil {
		return fmt.Errorf("check exists: %w", err)
	}
	if !exists {
		return sql.ErrNoRows
	}

	// Execute update (affects subscription linked to user)
	res, err := d.db.ExecContext(ctx, updateSubscriptionByUsernameSQL,
		defaultString(newSub.Duration, "month"),
		nullableTime(newSub.StartSubscription),
		nullableTime(newSub.EndSubscription),
		username,
	)
	if err != nil {
		return fmt.Errorf("update subscription: %w", err)
	}
	// Optional: verify rows affected > 0
	if n, _ := res.RowsAffected(); n == 0 {
		// nothing was updated; could be because values equal previous ones.
		log.Printf("update subscription: no rows affected for user %s", username)
	}
	return nil
}

// DeleteUser deletes user (and subscription will cascade-delete).
func (d *Database) DeleteUser(ctx context.Context, username string) error {
	if username == "" {
		return errors.New("username is required")
	}
	_, err := d.db.ExecContext(ctx, deleteUserSQL, username)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

// IsUserExists checks existence.
func (d *Database) IsUserExists(ctx context.Context, username string) (bool, error) {
	var exists bool
	if err := d.db.QueryRowContext(ctx, userExistsSQL, username).Scan(&exists); err != nil {
		return false, fmt.Errorf("exists query: %w", err)
	}
	return exists, nil
}

// SubscriptionStatus returns user's subscription status.
func (d *Database) SubscriptionStatus(ctx context.Context, username string) (string, error) {
	var status string
	if err := d.db.QueryRowContext(ctx, userSubscriptionStatusSQL, username).Scan(&status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", sql.ErrNoRows
		}
		return "", fmt.Errorf("subscription status query: %w", err)
	}
	return status, nil
}

// UpdateUserTraffic sets traffic value for a user.
func (d *Database) UpdateUserTraffic(ctx context.Context, username string, traffic float64) error {
	if username == "" {
		return errors.New("username is required")
	}
	_, err := d.db.ExecContext(ctx, updateUserTrafficSQL, traffic, username)
	if err != nil {
		return fmt.Errorf("update traffic: %w", err)
	}
	return nil
}

// ResetUserTraffic resets traffic to 0.
func (d *Database) ResetUserTraffic(ctx context.Context, username string) error {
	return d.UpdateUserTraffic(ctx, username, 0)
}

// AllUsernames returns all usernames.
func (d *Database) AllUsernames(ctx context.Context) ([]string, error) {
	rows, err := d.db.QueryContext(ctx, allUsernamesSQL)
	if err != nil {
		return nil, fmt.Errorf("all usernames query: %w", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, fmt.Errorf("scan username: %w", err)
		}
		names = append(names, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	return names, nil
}

func (d *Database) IsSubscriptionActive(ctx context.Context, username string) (bool, error) {
	if username == "" {
		return false, errors.New("username is required")
	}
	var active sql.NullBool
	if err := d.db.QueryRowContext(ctx, userSubscriptionActiveSQL, username).Scan(&active); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, sql.ErrNoRows
		}
		return false, fmt.Errorf("subscription active query: %w", err)
	}
	return active.Valid && active.Bool, nil
}

// Helper: default string if empty.
func defaultString(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

// Helper: nullableTime returns zero time as NOW if zero; but we pass time values directly to SQL.
func nullableTime(t time.Time) interface{} {
	if t.IsZero() {
		return time.Now()
	}
	return t
}
