package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"log"
	"time"
)

// 常量定义
const (
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
)

type (
	// 创建一个Token结构体，其中包含了令牌的明文、哈希值、用户ID、过期时间和作用域信息。
	Token struct {
		Plaintext string    `json:"token"`
		Hash      []byte    `json:"-"`
		UserID    int64     `json:"-"`
		Expiry    time.Time `json:"expiry"`
		Scope     string    `json:"-"`
	}

	// TokenModel结构体
	TokenModel struct {
		DB       *sql.DB
		InfoLog  *log.Logger
		ErrorLog *log.Logger
	}
)

// 创建一个TokenModel对象，并初始化其DB、InfoLog和ErrorLog字段。
func (m TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	// 生成一个Token对象，并返回可能的错误。
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	// 将生成的Token对象插入到数据库中，并返回可能的错误。
	err = m.Insert(token)
	return token, err

}

// 插入一个Token对象到数据库中，并返回可能的错误。
func (m TokenModel) Insert(token *Token) error {
	query := `
		INSERT INTO tokens (hash, user_id, expiry, scope)
		VALUES (?, ?, ?, ?)
		`

	// TODO 这里的token.hash存在问题，不能正常添加
	args := []interface{}{string(token.Hash), token.UserID, token.Expiry, token.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 执行对应sql语句操作
	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}

// 删除给定用户ID和作用域的所有令牌记录。
func (m TokenModel) DeleteAllForUser(scope string, userID int64) error {
	query := `
		DELETE FROM tokens
		WHERE scope = ? AND user_id = ?
		`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 执行对应sql语句操作
	_, err := m.DB.ExecContext(ctx, query, scope, userID)
	return err
}

// 生成一个Token对象，其中包含了用户ID、过期时间和作用域信息。
func generateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {
	// 创建一个Token实例，其中包含了用户ID、过期时间和作用域信息。
	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	// 初始化一个长度为16的零值字节切片
	randomBytes := make([]byte, 16)

	// 使用rand.Read()函数从系统随机数生成器中读取随机字节到randomBytes字节切片中。
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	// 使用base32.StdEncoding.WithPadding()函数将随机字节切片转换为base32编码的字符串，并赋值给token.Plaintext字段。
	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	// 生成哈希值，并将结果赋值给token.Hash字段。
	hash := sha256.Sum256([]byte(token.Plaintext))
	// 将哈希值前16个字节赋值给token.Hash字段。
	token.Hash = hash[:]

	return token, nil
}
