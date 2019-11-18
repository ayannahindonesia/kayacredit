package handlers

import (
	"asira_borrower/asira"
	"fmt"
	"net/http"
	"strconv"

	"github.com/dgrijalva/jwt-go"
	"gitlab.com/asira-ayannah/basemodel"

	"github.com/labstack/echo"
)

type BankList struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func AgentAllBank(c echo.Context) error {
	defer c.Request().Body.Close()
	var (
		banklist  []BankList
		totalRows int
	)
	user := c.Get("user")
	token := user.(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	agentID, err := strconv.Atoi(claims["jti"].(string))

	if err != nil {
		return returnInvalidResponse(http.StatusForbidden, err, "Akun tidak ditemukan")
	}
	//custom query
	db := asira.App.DB
	db = db.Table("banks b").
		Select("b.id, b.name").
		Joins("LEFT JOIN agents a ON b.id = ANY(a.banks)").
		Where("a.id = ?", agentID)
	//query
	err = db.Find(&banklist).Error
	if err != nil {
		fmt.Println(err)
	}

	//create custom response
	tempDB := db
	tempDB.Count(&totalRows)
	result := basemodel.PagedFindResult{
		TotalData: totalRows,
		Rows:      totalRows,
		Data:      banklist,
	}
	return c.JSON(http.StatusOK, result)
}
