package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"sync"

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
	wg     sync.WaitGroup
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
		form, err := c.MultipartForm()
		if err != nil {
			c.String(http.StatusBadRequest, "File not found in request: %v", err)
			return
		}

		files := form.File["files"]
		if len(files) == 0 {
			c.String(http.StatusBadRequest, "No files found in request")
			return
		}

		client := getClient(config)
		srv, err := drive.NewService(context.Background(), option.WithHTTPClient(client))
		if err != nil {
			c.String(http.StatusInternalServerError, "Unable to retrieve Drive client: %v", err)
			return
		}

		var fileIds []string

		go func() {
			defer wg.Wait()
		}()

		for _, file := range files {
			wg.Add(1)
			go uploadFile(c, srv, file, folderID)
		}
		c.String(http.StatusOK, "File uploaded: %s\n", fileIds)
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

func uploadFile(c *gin.Context, srv *drive.Service, file *multipart.FileHeader, folderID string) {
	defer wg.Done()

	f, err := file.Open()
	if err != nil {
		c.String(http.StatusInternalServerError, "Unable to open file: %v", err)
		return
	}
	defer f.Close()

	cm := imageCompression(c, f)

	driveFile := &drive.File{
		Name:    file.Filename,
		Parents: []string{folderID},
	}

	_, err = srv.Files.Create(driveFile).Media(cm).Do()
	if err != nil {
		c.String(http.StatusInternalServerError, "Unable to upload file: %v", err)
		return
	}
}

func imageCompression(c *gin.Context, f multipart.File) *bytes.Reader {
	img, _, err := image.Decode(f)
	if err != nil {
		c.String(http.StatusInternalServerError, "Unable to decode image: %v", err)
		return nil
	}

	var buf bytes.Buffer

	quality := 90
	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
	if err != nil {
		c.String(http.StatusInternalServerError, "Unable to compress image: %v", err)
		return nil
	}

	compressedImage := bytes.NewReader(buf.Bytes())

	return compressedImage
}
