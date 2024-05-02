package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const (
	idColumn = "id"
	name     = "name"
	surname  = "surname"
	email    = "email"
	avatar   = "avatar"
	login    = "login"
	password = "password"
	role     = "role"
	weight   = "weight"
	height   = "height"
	locked   = "locked"
)

// User структура представляет пользователя для создания
type User struct {
	Name     string  `json:"name"`
	Surname  string  `json:"surname"`
	Email    string  `json:"email"`
	Avatar   string  `json:"avatar"`
	Login    string  `json:"login"`
	Password string  `json:"password"`
	Role     int32   `json:"role"`
	Weight   float64 `json:"weight"`
	Height   float64 `json:"height"`
	Locked   bool    `json:"locked"`
}

// UserToGet структура представляет пользователя для получения
type UserToGet struct {
	Id       int64   `json:"id"`
	Name     string  `json:"name"`
	Surname  string  `json:"surname"`
	Email    string  `json:"email"`
	Avatar   string  `json:"avatar"`
	Login    string  `json:"login"`
	Password string  `json:"password"`
	Role     int32   `json:"role"`
	Weight   float64 `json:"weight"`
	Height   float64 `json:"height"`
	Locked   bool    `json:"locked"`
}

// CreateResponse структура представляет возвращаемый id
type CreateResponse struct {
	Id int64 `json:"id"`
}

var db *sql.DB

func main() {
	// Загрузка значений из файла .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Подключение к базе данных PostgreSQL
	db, err = sql.Open("postgres", os.Getenv("PG_DSN"))
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Ошибка при проверке соединения с базой данных: %v", err)
	}

	defer db.Close()

	r := gin.Default()

	r.POST("/users", createUser)
	r.GET("/users", getUsers)
	r.GET("/users/:id", getUserById)
	r.PUT("/users/:id", updateUserById)
	r.DELETE("/users/:id", deleteUserByID)

	log.Println("Server is running on: localhost:80808...")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

func createUser(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Println("1")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Println(user)
	var id int64
	query := fmt.Sprintf("INSERT INTO users (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s) VALUES ($1, $2, $3, $4, $5,$6, $7, $8, $9, $10 ) RETURNING id ", name, surname, email, avatar, login, password, role, weight, height, locked)

	err := db.QueryRow(query,
		user.Name,
		user.Surname,
		user.Email,
		user.Avatar,
		user.Login,
		user.Password,
		user.Role,
		user.Weight,
		user.Height,
		user.Locked,
	).
		Scan(&id)
	if err != nil {
		log.Println("2")
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	ID := CreateResponse{
		Id: id,
	}

	c.JSON(http.StatusCreated, ID)
}

func getUsers(c *gin.Context) {
	query := fmt.Sprintf("SELECT %s ,%s, %s, %s, %s, %s, %s, %s, %s, %s, %s FROM users", idColumn, name, surname, email, avatar, login, password, role, weight, height, locked)
	rows, err := db.Query(query)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get users"})
		return
	}
	defer rows.Close()

	users := []UserToGet{}
	for rows.Next() {
		var user UserToGet
		err := rows.Scan(&user.Id, &user.Name, &user.Surname, &user.Email, &user.Avatar, &user.Login, &user.Password,
			&user.Role, &user.Weight, &user.Height, &user.Locked)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get users"})
			return
		}
		users = append(users, user)
	}

	c.JSON(http.StatusOK, users)
}

func getUserById(c *gin.Context) {
	id, err := getId(c)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var user UserToGet

	query := fmt.Sprintf("SELECT %s, %s, %s, %s, %s, %s, %s, %s, %s, %s FROM users WHERE id = $1", name, surname, email, avatar, login, password, role, weight, height, locked)

	err = db.QueryRow(query, id).
		Scan(
			&user.Name,
			&user.Surname,
			&user.Email,
			&user.Avatar,
			&user.Login,
			&user.Password,
			&user.Role,
			&user.Weight,
			&user.Height,
			&user.Locked,
		)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}
	user.Id = id

	c.JSON(http.StatusOK, user)
}

func updateUserById(c *gin.Context) {
	id, err := getId(c)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var updateUser User
	if err := c.ShouldBindJSON(&updateUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := fmt.Sprintf("UPDATE users SET %s=$1, %s=$2, %s=$3, %s=$4, %s=$5, %s=$6, %s=$7, %s=$8, %s=$9, %s=$10 WHERE id=$11", name, surname, email, avatar, login, password, role, weight, height, locked)
	_, err = db.Exec(query,
		updateUser.Name, updateUser.Surname, updateUser.Email, updateUser.Avatar, updateUser.Login, updateUser.Password,
		updateUser.Role, updateUser.Weight, updateUser.Height, updateUser.Locked, id)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.Status(http.StatusOK)
}

func deleteUserByID(c *gin.Context) {
	id, err := getId(c)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	result, err := db.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.Status(http.StatusOK)
}

func getId(c *gin.Context) (int64, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return id, nil
}
