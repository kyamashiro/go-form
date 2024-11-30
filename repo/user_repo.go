package repo

import (
	"database/sql"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id        string
	Name      string
	password  string
	createdAt string
	updatedAt string
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (u *UserRepository) Exists(name string) (bool, error) {
	row := u.db.QueryRow("SELECT id FROM users WHERE name = $1", name)
	user := &User{}
	if err := row.Scan(&user.Id); err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return false, nil
		}
		return false, err
	}
	return user.Id != "", nil
}

func (u *UserRepository) Create(name, password string) (*User, error) {
	// パスワードはハッシュ化して保存する
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	_, err = u.db.Exec("INSERT INTO users (name, password) VALUES ($1, $2)", name, string(hashedPassword))
	if err != nil {
		return nil, err
	}
	user := &User{}
	err = u.db.QueryRow("SELECT id,name FROM users WHERE name = $1", name).Scan(&user.Id, &user.Name)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserRepository) Auth(name, password string) bool {
	user := &User{}
	err := u.db.QueryRow("SELECT id, name, password FROM users WHERE name = $1", name).Scan(&user.Id, &user.Name, &user.password)
	if err != nil {
		return false
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.password), []byte(password)); err != nil {
		return false
	}
	return true
}

func (u *UserRepository) FindByName(name string) *User {
	user := &User{}
	err := u.db.QueryRow("SELECT id, name FROM users WHERE name = $1", name).Scan(&user.Id, &user.Name)
	if err != nil {
		return nil
	}
	return user
}
