/**
 * @license
 * Copyright Google Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

var gmail_service *gmail.Service

const GMAIL_TOKEN_FILE = "../gmail_token.json"
const GMAIL_CREDENTIALS_FILE = "../gmail_credentials.json"
const RESET_PASSWORD_TEMPLATE_FILE = "../reset_password.html"

func GetTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

func GetClient(config *oauth2.Config) *http.Client {
	token, err := GetToken()
	if err != nil {
		token = GetTokenFromWeb(config)
		SaveToken(token)
	}
	return config.Client(context.Background(), token)
}

func GetToken() (*oauth2.Token, error) {
	f, err := os.Open(GMAIL_TOKEN_FILE)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	return token, err
}

func SaveToken(token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", GMAIL_TOKEN_FILE)
	file, err := os.OpenFile(GMAIL_TOKEN_FILE, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer file.Close()
	json.NewEncoder(file).Encode(token)
}

func GetGmailConfig() *oauth2.Config {
	credential_bytes, err := os.ReadFile(GMAIL_CREDENTIALS_FILE)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}
	config, err := google.ConfigFromJSON(credential_bytes, gmail.GmailSendScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	return config
}

func InitGmailService() {
	ctx := context.Background()
	config := GetGmailConfig()
	client := GetClient(config)

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatal("failed to receive gmail client", err.Error())
	}

	gmail_service = srv
}

var reset_password_template string

func InitPasswordResetTemplate() {
	bytes, err := os.ReadFile(RESET_PASSWORD_TEMPLATE_FILE)
	if err != nil {
		log.Fatal("Unable to read in reset_password.html")
	}
	reset_password_template = string(bytes)
}

const PWD_RESET_URL_ROUTE = "/password-reset-request?token="

func GetResetPasswordMessageBody(user User) (string, error) {
	otp := strconv.Itoa(rand.Int() % 1000)
	otp_id_map[otp] = user.ID

	replacer := strings.NewReplacer("{{name}}", user.Username, "{{otp}}", otp)
	final_msg := replacer.Replace(reset_password_template)
	return final_msg, nil
}

func SendEmail(user User, msg_body string) error {
	from := "From: " + "gocharge.group@gmail.com" + "\r\n"
	to := "To: " + user.Email + "\r\n"
	subject := "Subject: " + "Password Reset Request" + "\r\n"
	mime := "MIME-Version: " + "1.0" + "\r\n"
	content_type := "Content-Type: " + "text/html; charset=\"utf-8\"" + "\r\n\r\n"
	body := msg_body + "\r\n"

	msgString := from + to + subject + mime + content_type + body
	msg := []byte(msgString)

	message := gmail.Message{
		Raw: base64.URLEncoding.EncodeToString([]byte(msg)),
	}

	_, err := gmail_service.Users.Messages.Send("me", &message).Do()
	return err
}
