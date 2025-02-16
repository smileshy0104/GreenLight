package data

import (
	"DesignMode/GreenLight/internal/validator"
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"log"
	"time"
)

var (
	// ErrDuplicateEmail 邮箱重复
	ErrDuplicateEmail = errors.New("duplicate email")
)

// UserModel 结构体
type UserModel struct {
	DB       *sql.DB
	InfoLog  *log.Logger
	ErrorLog *log.Logger
}

// AnonymousUser 匿名用户
var AnonymousUser = &User{}

// IsAnonymous 判断用户是否为匿名用户
func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

// User 结构体
type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

// password 结构体
type password struct {
	plaintext *string
	hash      []byte
}

// Set 设置密码
func (p *password) Set(plaintextPassword string) error {
	// 使用bcrypt对密码进行加密
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}
	p.plaintext = &plaintextPassword
	p.hash = hash
	return nil
}

// Matches 判断密码是否匹配
func (p *password) Matches(plaintextPassword string) (bool, error) {
	// 使用bcrypt对密码进行比较
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

// Insert 插入用户
func (m UserModel) Insert(user *User) error {
	query := `
		INSERT INTO users (created_at, name, email, password_hash, activated)
		VALUES (?, ?, ?, ?, ?)
		`

	args := []interface{}{time.Now().Unix(), user.Name, user.Email, user.Password.hash, user.Activated}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 执行插入
	err := m.DB.QueryRowContext(ctx, query, args...).Err()
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

// GetByEmail 通过邮箱获取用户
func (m UserModel) GetByEmail(email string) (*User, error) {
	query := `
		SELECT id, created_at, name, email, password_hash, activated, version
		FROM users
		WHERE email = ?
		`
	// 声明user
	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 执行查询
	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

// Update 更新用户
func (m UserModel) Update(user *User) error {
	query := `
		UPDATE users
		SET name = ?, email = ?, password_hash = ?, activated = ?, version = version + 1
		WHERE id = ? AND version = ?
		`

	args := []interface{}{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activated,
		user.ID,
		user.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// 执行更新
	err := m.DB.QueryRowContext(ctx, query, args...).Err()
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

// GetForToken 通过token获取用户
func (m UserModel) GetForToken(tokenScope, tokenPlaintext string) (*User, error) {
	// 创建一个sha256哈希
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	query := `
		SELECT 
			users.id, users.created_at, users.name, users.email, 
			users.password_hash, users.activated, users.version
		FROM       users
        INNER JOIN tokens
			ON users.id = tokens.user_id
        WHERE tokens.hash = ?  
			AND tokens.scope = ?
			AND tokens.expiry > ?
		`

	args := []interface{}{tokenHash[:], tokenScope, time.Now()}

	// 声明user
	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 执行查询
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	// 返回匹配的user
	return &user, nil
}

// ValidateEmail 验证邮箱
func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

// ValidatePasswordPlaintext 验证密码
func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

// ValidateUser 验证用户
func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")
	// 调用 ValidateEmail() helper。
	ValidateEmail(v, user.Email)
	// 调用 ValidatePasswordPlaintext() helper。
	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}
	// 如果密码为空，则返回。
	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}
