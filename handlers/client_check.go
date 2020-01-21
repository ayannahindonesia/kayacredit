package handlers

import (
	"asira_borrower/asira"
	"net/http"
	"strings"

	"github.com/labstack/echo"
)

type (
	Borrower struct {
		IdCardNumber string `json:"idcard_number" gorm:"column:idcard_number;type:varchar(255);unique;not null"`
		TaxIDnumber  string `json:"taxid_number" gorm:"column:taxid_number;type:varchar(255)"`
		Email        string `json:"email" gorm:"column:email;type:varchar(255);unique"`
		Phone        string `json:"phone" gorm:"column:phone;type:varchar(255);unique;not null"`
	}
)

func CheckData(c echo.Context) error {
	defer c.Request().Body.Close()
	var (
		borrower Borrower
	)
	var values []string

	if email := c.QueryParam("email"); email != "" && !asira.App.DB.Where("email = ? AND agent_referral = 0", email).Find(&borrower).RecordNotFound() {
		values = append(values, EnglishToIndonesiaFieldsUnderscored["email"])
	}
	if phone := c.QueryParam("phone"); phone != "" && !asira.App.DB.Where("phone = ? AND agent_referral = 0", phone).Find(&borrower).RecordNotFound() {
		values = append(values, EnglishToIndonesiaFieldsUnderscored["phone"])
	}
	if idcard_number := c.QueryParam("idcard_number"); idcard_number != "" && !asira.App.DB.Where("idcard_number = ? AND agent_referral = 0", idcard_number).Find(&borrower).RecordNotFound() {
		values = append(values, EnglishToIndonesiaFieldsUnderscored["idcard_number"])
	}
	if taxid_number := c.QueryParam("taxid_number"); taxid_number != "" && !asira.App.DB.Where("taxid_number = ? AND agent_referral = 0", taxid_number).Find(&borrower).RecordNotFound() {
		values = append(values, EnglishToIndonesiaFieldsUnderscored["taxid_number"])
	}
	if len(values) < 1 {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":  true,
			"message": "Ok",
		})
	}
	result := "Field : " + strings.Join(values, " , ") + " Telah Digunakan"
	return returnInvalidResponse(http.StatusUnprocessableEntity, "", result)

}
