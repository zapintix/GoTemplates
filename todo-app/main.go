package main

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type User struct {
	ID       int
	Username string
	Password string
}

var db *sql.DB

func main() {
	var err error
	connStr := "user=postgres password=admin dbname=users sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(50) UNIQUE NOT NULL,
		password VARCHAR(100) NOT NULL
	)`)
	if err != nil {
		panic(err)
	}

	r := gin.Default()

	r.POST("/login", loginHandler)
	r.POST("/register", registerHandler)

	r.Run(":8080")
}
func loginHandler(c *gin.Context) {
	var user User
	// Привязываем данные запроса к структуре User.
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем наличие пользователя в базе данных.
	err := db.QueryRow("SELECT id FROM users WHERE username=$1 AND password=$2", user.Username, user.Password).Scan(&user.ID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Пользователь успешно аутентифицирован.
	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "userID": user.ID})
}

// Обработчик для регистрации нового пользователя.
func registerHandler(c *gin.Context) {
	var user User
	// Привязываем данные запроса к структуре User.
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем, не существует ли уже пользователь с таким именем.
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username=$1", user.Username).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	// Добавляем пользователя в базу данных.
	_, err = db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", user.Username, user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	// Пользователь успешно зарегистрирован.
	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}
