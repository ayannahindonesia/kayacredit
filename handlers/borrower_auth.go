package handlers

import (
	"asira_borrower/asira"
	"asira_borrower/models"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/thedevsaddam/govalidator"
)

type (
	//BorrowerLoginCreds for payload
	BorrowerLoginCreds struct {
		Key      string `json:"key"`
		Password string `json:"password"`
	}
)

//BorrowerLogin borrower login, borrower can choose either login with email / phone
func BorrowerLogin(c echo.Context) error {
	defer c.Request().Body.Close()

	LogTag := "BorrowerLogin"

	var (
		credentials BorrowerLoginCreds
		loginType   string
		borrower    models.Borrower
		validKey    bool
		token       string
		err         error
	)

	rules := govalidator.MapData{
		"key":      []string{"required"},
		"password": []string{"required"},
	}

	validate := validateRequestPayload(c, rules, &credentials)
	if validate != nil {
		NLog("warning", LogTag, map[string]interface{}{
			NLOGMSG: "error authentification",
			NLOGERR: validate}, c.Get("user").(*jwt.Token), "", true, "")

		return returnInvalidResponse(http.StatusBadRequest, validate, "Gagal login")
	}

	emailchecker := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

	if emailchecker.MatchString(credentials.Key) {
		loginType = "email"
	}

	// check if theres record
	switch loginType {
	default: // default login is using phone number
		validKey = asira.App.DB.Where("phone = ? AND agent_referral = 0", credentials.Key).Find(&borrower).RecordNotFound()
		break
	case "email":
		validKey = asira.App.DB.Where("email = ? AND agent_referral = 0", credentials.Key).Find(&borrower).RecordNotFound()
		break
	}
	//check login data exist or not
	user := models.User{}
	err = user.FindbyBorrowerID(borrower.ID)
	if err != nil {
		NLog("warning", LogTag, map[string]interface{}{
			NLOGMSG: "error authentification",
			NLOGERR: err}, c.Get("user").(*jwt.Token), "", true, "")

		return returnInvalidResponse(http.StatusUnprocessableEntity, err, "Borrower tidak memiliki akun personal")
	}

	if !validKey { // check the password

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password))
		if err != nil {
			NLog("error", LogTag, map[string]interface{}{
				NLOGMSG:    "error authentification",
				NLOGERR:    err,
				"username": credentials.Key}, c.Get("user").(*jwt.Token), "", true, "")

			return returnInvalidResponse(http.StatusUnprocessableEntity, err, "Password anda salah")
		}

		tokenrole := "unverified_borrower"
		if borrower.OTPverified {
			tokenrole = "borrower"
		}
		token, err = createJwtToken(strconv.FormatUint(borrower.ID, 10), tokenrole)
		if err != nil {
			NLog("error", LogTag, map[string]interface{}{
				NLOGMSG:    "error generating token",
				NLOGERR:    err,
				"username": credentials.Key}, c.Get("user").(*jwt.Token), "", true, "")

			return returnInvalidResponse(http.StatusInternalServerError, err, "Gagal membuat token")
		}
	} else {

		NLog("error", LogTag, map[string]interface{}{
			NLOGMSG:    "error login ",
			NLOGERR:    err,
			"username": credentials.Key}, c.Get("user").(*jwt.Token), "", true, "")

		return returnInvalidResponse(http.StatusUnprocessableEntity, "", "Gagal Login")
	}

	//logging
	NLog("info", LogTag, map[string]interface{}{
		NLOGMSG:    "Login Success",
		NLOGERR:    err,
		"username": credentials.Key}, c.Get("user").(*jwt.Token), "", true, "")

	jwtConf := asira.App.Config.GetStringMap(fmt.Sprintf("%s.jwt", asira.App.ENV))
	expiration := time.Duration(jwtConf["duration"].(int)) * time.Minute
	return c.JSON(http.StatusOK, map[string]interface{}{
		"token":      token,
		"expires_in": expiration.Seconds(),
	})
}
