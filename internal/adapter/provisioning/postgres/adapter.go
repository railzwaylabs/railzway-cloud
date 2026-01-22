package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Adapter struct {
	adminConnString string
}

func NewAdapter(adminConnString string) *Adapter {
	return &Adapter{
		adminConnString: adminConnString,
	}
}

// Provision implements provisioning.DatabaseProvisioner
func (a *Adapter) Provision(ctx context.Context, orgID int64, password string) error {
	conn, err := pgx.Connect(ctx, a.adminConnString)
	if err != nil {
		return fmt.Errorf("failed to connect to admin db: %w", err)
	}
	defer conn.Close(ctx)

	userName := fmt.Sprintf("railzway_user_%d", orgID)
	dbName := fmt.Sprintf("railzway_org_%d", orgID)

	// 1. Create User (Idempotent)
	// Check if user exists
	var exists bool
	err = conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_roles WHERE rolname=$1)", userName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}

	if !exists {
		// Create User
		// NOTE: Parameterized queries for identifiers (like username) are not supported in standard SQL
		// We securely format the string since we control the inputs (orgID is int)
		query := fmt.Sprintf("CREATE USER %q WITH PASSWORD '%s'", userName, password)
		if _, err := conn.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
	} else {
		// Rotate Password
		query := fmt.Sprintf("ALTER USER %q WITH PASSWORD '%s'", userName, password)
		if _, err := conn.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to rotate password: %w", err)
		}
	}

	// 2. Create Database (Idempotent)
	// Check if db exists
	err = conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)", dbName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check db existence: %w", err)
	}

	if !exists {
		// Create Database and assign owner
		query := fmt.Sprintf("CREATE DATABASE %q OWNER %q", dbName, userName)
		if _, err := conn.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
	} else {
		// Ensure owner is correct (Optional but good for drift)
		query := fmt.Sprintf("ALTER DATABASE %q OWNER TO %q", dbName, userName)
		if _, err := conn.Exec(ctx, query); err != nil {
			// Warn but don't fail, maybe? Or fail. Let's fail for correctness.
			return fmt.Errorf("failed to set database owner: %w", err)
		}
	}

	// 3. Grant Privileges (Ensure user has access to public schema in their DB)
	// Note: We need to connect to the NEW database to grant schema privileges inside it.
	// However, OWNER usually has full rights.
	// For standard public schema usage, OWNER is sufficient.
	// If we needed specific grants, we'd loop in a new connection here.

	return nil
}
