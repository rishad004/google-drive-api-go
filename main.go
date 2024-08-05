package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

var (
	config *oauth2.Config
	tok    *oauth2.Token
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("------------------Error to Load ENV-------------------------")
	}
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		log.Fatalf("GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET environment variables must be set")
	}

	config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{drive.DriveFileScope},
		Endpoint:     google.Endpoint,
		RedirectURL:  "http://localhost:8080/callback",
	}

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		url := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
		c.HTML(http.StatusOK, "index.html", gin.H{
			"AuthURL": url,
		})
	})

	r.GET("/callback", func(c *gin.Context) {
		code := c.Query("code")
		tok, err := config.Exchange(context.Background(), code)
		if err != nil {
			c.String(http.StatusBadRequest, "Unable to retrieve token from web: %v", err)
			return
		}
		saveToken("token.json", tok)
		c.Redirect(307, "http://localhost:8080/")
	})

	r.POST("/create_folder", func(c *gin.Context) {
		parentFolderID := c.PostForm("parent_folder_id")
		folderName := c.PostForm("folder_name")

		client := getClient(config)
		srv, err := drive.NewService(context.Background(), option.WithHTTPClient(client))
		if err != nil {
			c.String(http.StatusInternalServerError, "Unable to retrieve Drive client: %v", err)
			return
		}

		folder := &drive.File{
			Name:     folderName,
			MimeType: "application/vnd.google-apps.folder",
			Parents:  []string{parentFolderID},
		}
		folder, err = srv.Files.Create(folder).Do()
		if err != nil {
			c.String(http.StatusInternalServerError, "Unable to create folder: %v", err)
			return
		}
		c.String(http.StatusOK, "Folder created: %s\n", folder.Id)
	})

	r.POST("/upload_file", func(c *gin.Context) {
		folderID := c.PostForm("folder_id")
		file, err := c.FormFile("file")
		if err != nil {
			c.String(http.StatusBadRequest, "File not found in request: %v", err)
			return
		}

		client := getClient(config)
		srv, err := drive.NewService(context.Background(), option.WithHTTPClient(client))
		if err != nil {
			c.String(http.StatusInternalServerError, "Unable to retrieve Drive client: %v", err)
			return
		}

		f, err := file.Open()
		if err != nil {
			c.String(http.StatusInternalServerError, "Unable to open file: %v", err)
			return
		}
		defer f.Close()

		driveFile := &drive.File{
			Name:    file.Filename,
			Parents: []string{folderID},
		}

		driveFile, err = srv.Files.Create(driveFile).Media(f).Do()
		if err != nil {
			c.String(http.StatusInternalServerError, "Unable to upload file: %v", err)
			return
		}
		c.String(http.StatusOK, "File uploaded: %s\n", driveFile.Id)
	})

	r.LoadHTMLFiles("templates/index.html")
	r.Run(":8080")
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func getClient(config *oauth2.Config) *http.Client {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		log.Fatalf("Unable to read token from file: %v", err)
	}
	return config.Client(context.Background(), tok)
}
