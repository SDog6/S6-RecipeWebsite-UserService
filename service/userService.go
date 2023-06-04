package service

import (
	"fmt"
	"os"
	"time"
	"userMicroService/dbaccess"
	"userMicroService/model"

	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	cache "github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

var usersCache *cache.Cache

func init() {
	// Initialize the cache with a default expiration time of 5 minutes
	usersCache = cache.New(5*time.Minute, 10*time.Minute)
}

func RegisterUser(c *gin.Context) {
	db := dbaccess.ConnectToDb()
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})

	file, err := os.OpenFile("RegisterUserLog.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Error("Error opening file:", err)
		log.Fatal("Error opening file:", err)
	}
	defer file.Close()

	// Set the output of the logger to the file
	logger.SetOutput(file)

	// Get client IP
	clientIP := c.ClientIP()

	var userCarrier model.User
	err = c.BindJSON(&userCarrier)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"IP":     clientIP,
			"Status": "Error",
		}).Error("(RegisterUser) c.BindJSON", err)
		log.Fatal("(RegisterUser) c.BindJSON", err)
	}

	query1 := `SELECT email FROM user WHERE email = ?`
	rows, err := db.Query(query1, userCarrier.Email)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"IP":     clientIP,
			"Status": "Error",
		}).Error(err)
		log.Fatal(err)
	}

	// Process the result
	for rows.Next() {
		var email string
		err := rows.Scan(&email)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"IP":     clientIP,
				"Status": "Error",
			}).Error(err)
			log.Fatal(err)
		}
		if email != "" {
			logger.WithFields(logrus.Fields{
				"IP":     clientIP,
				"Status": "Error",
			}).Error("Email already exists")
			message := []byte("Email already exists")
			n, err := file.Write(message)
			if err != nil {
				logger.WithFields(logrus.Fields{
					"IP":     clientIP,
					"Status": "Error",
				}).Error("Error writing to file:", err)
				fmt.Println("Error writing to file:", err)
				return
			}
			logger.WithFields(logrus.Fields{
				"IP":           clientIP,
				"Status":       "Error",
				"BytesWritten": n,
			}).Info("Bytes written to file")
			fmt.Printf("Bytes written: %d\n", n)
			c.JSON(http.StatusOK, "Email already exists")
			return
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(userCarrier.Password), 10)

	if err != nil {
		logger.WithFields(logrus.Fields{
			"IP":     clientIP,
			"Status": "Error",
		}).Error(err)
		log.Fatal(err)
	}

	query := `INSERT INTO user (first_name, last_name, email,password) VALUES (?, ?,?,?)`
	res, err := db.Exec(query, userCarrier.Email, hash, userCarrier.Role)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"IP":     clientIP,
			"Status": "Error",
		}).Error("(RegisterUser) db.Exec", err)
		log.Fatal("(RegisterUser) db.Exec", err)
	}
	userCarrier.ID, err = res.LastInsertId()
	if err != nil {
		logger.WithFields(logrus.Fields{
			"IP":     clientIP,
			"Status": "Error",
			"Query":  query,
		}).Error("(CreateProduct) res.LastInsertId", err)
		log.Fatal("(CreateProduct) res.LastInsertId", err)
	}

	// Log the successful registration with client IP
	logger.WithFields(logrus.Fields{
		"IP":     clientIP,
		"Status": "Success",
	}).Info("User registered successfully")

	c.JSON(http.StatusOK, userCarrier)
}

func GetUsers(c *gin.Context) {
	type UsersResponse struct {
		Users []model.User `json:"users"`
	}

	// Check if the users are already cached
	if cachedUsers, found := usersCache.Get("users"); found {
		// If the users are cached, return the cached data
		c.JSON(http.StatusOK, UsersResponse{Users: cachedUsers.([]model.User)})
		return
	}

	db := dbaccess.ConnectToDb()
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})

	file, err := os.OpenFile("GetUsersLog.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Error("Error opening file:", err)
		log.Fatal("Error opening file:", err)
	}
	defer file.Close()

	// Set the output of the logger to the file
	logger.SetOutput(file)

	// Get client IP
	clientIP := c.ClientIP()

	query := "SELECT * FROM user"
	res, err := db.Query(query)
	defer res.Close()
	if err != nil {
		logger.WithFields(logrus.Fields{
			"IP":     clientIP,
			"Status": "Error",
		}).Error("(GetProducts) db.Query", err)
		log.Fatal("(GetProducts) db.Query", err)
	}

	users := []model.User{}
	for res.Next() {
		var user model.User
		err := res.Scan(&user.ID, &user.Email, &user.Password, &user.Role)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"IP":     clientIP,
				"Status": "Error",
			}).Error("(GetUsers) res.Scan", err)
			log.Fatal("(GetUsers) res.Scan", err)
		}
		users = append(users, user)
	}

	// Store the retrieved users in the cache
	usersCache.Set("users", users, cache.DefaultExpiration)

	// Log the result to the file with client IP
	logger.WithFields(logrus.Fields{
		"IP":     clientIP,
		"Status": "Success",
	}).Info("Users retrieved successfully")

	// Wrap the users array within an object
	response := UsersResponse{Users: users}

	c.JSON(http.StatusOK, response)
}

func Login(c *gin.Context) {
	db := dbaccess.ConnectToDb()
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	file, err := os.OpenFile("LoginAttemptLog.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Error("Error opening file:", err)
		log.Fatal("Error opening file:", err)
	}
	defer file.Close()

	// Set the output of the logger to the file
	logger.SetOutput(file)

	// Output logs as JSON format
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Get client IP
	clientIP := c.ClientIP()

	var attempt model.LoginAttept
	if c.BindJSON(&attempt) != nil {
		logger.Error("Failed to read the body")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read the body",
		})
		return
	}

	query := `SELECT * FROM user WHERE email = ?`
	rows, err := db.Query(query, attempt.Email)
	if err != nil {
		logger.Error(err)
		log.Fatal(err)
	}
	defer rows.Close()

	found := false
	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.ID, &user.Email, &user.Password, &user.Role); err != nil {
			logger.Error(err)
			log.Fatal(err)
		}

		err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(attempt.Password))
		if err != nil {
			logger.WithFields(logrus.Fields{
				"IP":     clientIP,
				"Email":  attempt.Email,
				"Status": "Failed",
			}).Error("Wrong email or password")

			c.JSON(http.StatusBadRequest, "Wrong email or password")
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"id":     user.ID,
			"expire": time.Now().Add(time.Hour * 24 * 30).Unix(),
		})

		tokenString, err := token.SignedString([]byte(os.Getenv("SECRET")))
		if err != nil {
			logger.WithFields(logrus.Fields{
				"IP":     clientIP,
				"Email":  attempt.Email,
				"Status": "Failed",
			}).Error("Failed to generate token")

			c.JSON(http.StatusBadRequest, "Failed to generate token")
			return
		}

		logger.WithFields(logrus.Fields{
			"IP":     clientIP,
			"Email":  attempt.Email,
			"Status": "Success",
		}).Info("Login successful")

		c.JSON(http.StatusOK, gin.H{
			"token": tokenString,
		})

		found = true
		break
	}

	if err := rows.Err(); err != nil {
		logger.Error(err)
		log.Fatal(err)
	}

	if !found {
		logger.WithFields(logrus.Fields{
			"IP":     clientIP,
			"Email":  attempt.Email,
			"Status": "Failed",
		}).Info("Login failed: User not found")

		c.JSON(http.StatusBadRequest, "User not found")
	}
}
