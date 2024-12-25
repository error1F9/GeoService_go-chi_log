package main

import (
	"encoding/json"
	"github.com/go-chi/jwtauth/v5"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var tokenAuth *jwtauth.JWTAuth

var tokenString = make(map[string]string)

var logPass = make(map[string]string)

func init() {
	tokenAuth = jwtauth.New("HS256", []byte("123456"), nil)
}

// @Summary Регистрация пользователя
// @Tags Authentication
// @Description Регистрирует нового пользователя.
// @ID register
// @Accept json
// @Produce json
// @Param input body User true "Данные пользователя для регистрации"
// @Success 200 {string} string "Регистрация прошла успешно"
// @Failure 400 {string} string "Некорректный запрос или пользователь уже существует"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /register [post]
func Register(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if user.Username == "" || user.Password == "" {
		http.Error(w, "username or password is empty", http.StatusBadRequest)
		return
	}

	if _, ok := logPass[user.Username]; ok {
		http.Error(w, "Username already registered", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user.Password = string(hashedPassword)

	logPass[user.Username] = string(hashedPassword)

	_, tokenString[user.Username], _ = tokenAuth.Encode(map[string]interface{}{
		"username": user.Username,
		"password": user.Password,
	})
	w.WriteHeader(http.StatusOK)

}

// @Summary Вход пользователя
// @Tags Authentication
// @Description Аутентифицирует пользователя и возвращает JWT токен.
// @ID login
// @Accept json
// @Produce json
// @Param input body User true "Данные пользователя для входа"
// @Success 200 {object} map[string]string "Успешный вход, токен в поле 'token'"
// @Failure 400 {string} string "Некорректный запрос"
// @Failure 200 {string} string "Пользователь не существует или неверный пароль"
// @Router /login [post]
func Login(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, ok := logPass[user.Username]; !ok {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("User does not exist"))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(logPass[user.Username]), []byte(user.Password)); err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Wrong password"))
		return
	}

	_, token, _ := tokenAuth.Encode(map[string]interface{}{"user_id": user.Username})
	tokenString[user.Username] = token

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": token})

}
