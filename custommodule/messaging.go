package custommodule

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo"
)

type (
	Messaging struct {
		Key         string
		Secret      string
		Token       string
		Expire      time.Time
		URL         string
		AdminKey    string
		AdminSecret string
		Endpoints   MessagingEndpoints
	}
	MessagingEndpoints struct {
		ClientAuth       string
		SMS              string
		PushNotification string
		ListNotification string
		AdminAuth        string
	}
)

// SetConfig func
func (model *Messaging) SetConfig(key string, secret string, adminKey string, adminSecret string, URL string, Endpoints MessagingEndpoints) {
	model.Key = key
	model.Secret = secret
	model.URL = URL
	model.Endpoints = Endpoints
	model.AdminKey = adminKey
	model.AdminSecret = adminSecret
}

// ClientAuth func
func (model *Messaging) ClientAuth() (err error) {

	fmt.Println("model.Key ==> ", model.Key, model.Secret)
	basicToken := base64.StdEncoding.EncodeToString([]byte(model.Key + ":" + model.Secret))
	request, _ := http.NewRequest("GET", model.URL+model.Endpoints.ClientAuth, nil)
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", fmt.Sprintf("Basic %s", basicToken))

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	buffer, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	var parseresponse map[string]interface{}
	json.Unmarshal([]byte(buffer), &parseresponse)

	expires := 1200
	if parseresponse["expires_in"] != nil {
		expires = int(parseresponse["expires_in"].(float64))
	}
	model.Expire = time.Now().Local().Add(time.Second * time.Duration(expires))

	if parseresponse["token"] != nil {
		model.Token = parseresponse["token"].(string)
	} else {
		return fmt.Errorf("no token detected")
	}

	return nil
}

//AdminAuth func
func (model *Messaging) AdminAuth() (err error) {
	var parseresponse map[string]interface{}

	basicToken := base64.StdEncoding.EncodeToString([]byte(model.Key + ":" + model.Secret))
	request, _ := http.NewRequest("GET", model.URL+model.Endpoints.ClientAuth, nil)
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", fmt.Sprintf("Basic %s", basicToken))

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	buffer, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	//parse client login
	json.Unmarshal([]byte(buffer), &parseresponse)

	//set up payload and do login
	payload, _ := json.Marshal(map[string]interface{}{
		"key":      model.AdminKey,
		"password": model.AdminSecret,
	})

	// fmt.Println("parseresponse[token] ==> ", parseresponse["token"], model.AdminKey, model.AdminSecret, model.Endpoints.AdminAuth)
	request, _ = http.NewRequest("POST", model.URL+model.Endpoints.AdminAuth, bytes.NewBuffer(payload))
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", parseresponse["token"]))

	response, err = http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	buffer, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	//parse client login
	json.Unmarshal([]byte(buffer), &parseresponse)

	expires := 1200
	if parseresponse["expires_in"] != nil {
		expires = int(parseresponse["expires_in"].(float64))
	}
	model.Expire = time.Now().Local().Add(time.Second * time.Duration(expires))

	if parseresponse["token"] != nil {
		model.Token = parseresponse["token"].(string)
	} else {
		return fmt.Errorf("no token detected")
	}

	return nil
}

// RefreshToken func
func (model *Messaging) RefreshToken(auth_type string) (err error) {
	if time.Now().After(model.Expire) {
		// if auth_type == "client" {
		err = model.ClientAuth()
		// } else if auth_type == "admin" {
		// 	err = model.AdminAuth()
		// }
		if err != nil {
			return err
		}
	}

	return nil
}

// SendSMS func
func (model *Messaging) SendSMS(number string, message string) (err error) {
	err = model.RefreshToken("client")
	if err != nil {
		return err
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"phone_number": number,
		"message":      message,
	})

	request, _ := http.NewRequest("POST", model.URL+model.Endpoints.SMS, bytes.NewBuffer(payload))
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", model.Token))

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(response.Body)
		log.Printf("Failed sending sms : %s", string(body))

		return fmt.Errorf("Failed sending SMS")
	}

	return nil
}

func (model *Messaging) SendNotificationByToken(title string, message_body string, map_data map[string]string, firebase_token string) (err error) {

	//bug cycling call dependency
	//topics := asira.App.Config.GetStringMap(fmt.Sprintf("%s.messaging.push_notification", asira.App.ENV))

	err = model.RefreshToken("client")
	if err != nil {
		return err
	}
	if firebase_token == "" {
		firebase_token = model.Endpoints.PushNotification
	}
	payload, _ := json.Marshal(map[string]interface{}{
		"title":          title,
		"message_body":   message_body,
		"firebase_token": firebase_token,
		"data":           map_data,
	})

	request, _ := http.NewRequest("POST", model.URL+model.Endpoints.PushNotification, bytes.NewBuffer(payload))
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", model.Token))

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	log.Println("PUSH NOTIF : ", response)
	if response.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(response.Body)
		log.Printf("Failed sending notification : %s", string(body))

		return fmt.Errorf("Failed sending notification")
	}

	return nil
}

//TODO: GetNotificationBySenderId
//NOTE: get data from Messaging microservice
func (model *Messaging) GetNotificationByToken(token string, c echo.Context) (string, error) {

	//err := model.RefreshToken("admin")//BUG: gak bisa keselect
	err := model.AdminAuth()
	if err != nil {
		return "", err
	}

	//+"?token="+token
	request, _ := http.NewRequest("GET", model.URL+model.Endpoints.ListNotification, nil)
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", model.Token))

	//create url query
	q := request.URL.Query()

	// pagination parameters
	q.Add("rows", c.QueryParam("rows"))
	q.Add("page", c.QueryParam("page"))
	q.Add("orderby", c.QueryParam("orderby"))
	q.Add("sort", c.QueryParam("sort"))
	// filters
	q.Add("id", c.QueryParam("id"))
	q.Add("client_id", c.QueryParam("client_id"))
	q.Add("title", c.QueryParam("title"))
	//NOTE: token diambil dari DB (berdasarkan ID Borrower) bukan dari parameter
	q.Add("token", token)
	q.Add("topic", c.QueryParam("topic"))
	q.Add("send_time", c.QueryParam("send_time"))
	request.URL.RawQuery = q.Encode()
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	log.Println("GET NOTIF : ", response)
	if response.StatusCode != http.StatusOK {

		log.Printf("Failed get notification : %s", string(body))

		return string(body), fmt.Errorf("Failed get notification")
	}

	return string(body), nil
}
