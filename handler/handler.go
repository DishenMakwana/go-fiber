package handler

import (
	"net/mail"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/dishenmakwana/go-fiber/database"
	"github.com/dishenmakwana/go-fiber/helper"
	"github.com/dishenmakwana/go-fiber/middleware"
	"github.com/dishenmakwana/go-fiber/model"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func valid(email string) bool {
	_, err := mail.ParseAddress(email)

	return err == nil
}

func CreateUser(c *fiber.Ctx) error {
	db := database.DB.Db
	user := new(model.User)

	if err := c.BodyParser(user); err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Something's wrong with your input", "data": err})
	}

	if err := db.Where("username= ?", &user.Username).First(&user).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Username already taken", "data": err})
	}

	if err := db.Where("email= ?", &user.Email).First(&user).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Email already taken", "data": err})
	}

	hashed, err := helper.HashPassword(user.Password)
	user.Password = hashed

	err = db.Create(&user).Error
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Could not create user", "data": err})
	}

	return c.Status(201).JSON(fiber.Map{"status": "success", "message": "User has created", "data": user})
}

func GetAllUsers(c *fiber.Ctx) error {
	db := database.DB.Db
	var users []model.User

	db.Select([]string{"id", "username", "email"}).Find(&users).Order("created_at desc")

	if len(users) == 0 {
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "Users not found", "data": nil})
	}

	return c.Status(200).JSON(fiber.Map{"status": "success", "message": "Users Found", "data": users})
}

func GetSingleUser(c *fiber.Ctx) error {

	db := database.DB.Db
	id := c.Params("id")

	var user model.User
	db.Select([]string{"id", "username", "email"}).Find(&user, "id = ?", id)

	if user.ID == uuid.Nil {
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "User not found", "data": nil})
	}

	return c.Status(200).JSON(fiber.Map{"status": "success", "message": "User Found", "data": user})
}

func UpdateUser(c *fiber.Ctx) error {

	type updateUser struct {
		Username string `json:"username"`
	}

	db := database.DB.Db
	var user model.User

	id := c.Params("id")

	db.Find(&user, "id = ?", id)

	if user.ID == uuid.Nil {
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "User not found", "data": nil})
	}

	var updateUserData updateUser
	err := c.BodyParser(&updateUserData)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Something's wrong with your input", "data": err})
	}

	user.Username = updateUserData.Username
	db.Save(&user)

	return c.Status(200).JSON(fiber.Map{"status": "success", "message": "users Found", "data": user})

}

func DeleteUserByID(c *fiber.Ctx) error {
	db := database.DB.Db
	var user model.User

	id := c.Params("id")

	db.Find(&user, "id = ?", id)

	if user.ID == uuid.Nil {
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "User not found", "data": nil})
	}

	err := db.Delete(&user, "id = ?", id).Error

	if err != nil {
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "Failed to delete user", "data": nil})
	}

	return c.Status(200).JSON(fiber.Map{"status": "success", "message": "User deleted"})
}

func Login(c *fiber.Ctx) error {
	db := database.DB.Db
	var user model.User

	var input middleware.LoginInput

	// binding user input to a struct
	if err := c.BodyParser(&input); err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// set a variable depending on the condition
	var query string
	if valid(input.Identity) {
		query = "email= ?"
	} else {
		query = "username= ?"
	}

	if err := db.Where(query, input.Identity).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status": "error", "message": "User does not exists",
		})
	}

	identity := input.Identity
	pass := input.Password

	if !helper.ValidatePassword(pass, user.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status": "error", "mess_validation": "Password incorrect",
		})
	}

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["identity"] = identity
	claims["admin"] = true
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	t, err := token.SignedString([]byte("secret"))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{"status": "success", "message": "Success login", "token": t})

}
