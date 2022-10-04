package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
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

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func backup(srv *drive.Service, fileId string) (string, error) {
	progressUpdater := func(current, total int64) { log.Printf("%dB / %dB total\n", current, total) }

	f, err := os.Open("folder.zip")
	if err != nil {
		log.Fatalf("Unable to open test.txt: %v", err)
		return "", err
	}

	fInf, err := f.Stat()
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	var up *drive.File
	if fileId == "" {
		fmt.Println("Creating Backup...")
		up, err = srv.Files.Create(&drive.File{Name: "test_folder.zip", MimeType: "application/zip"}).ResumableMedia(context.Background(), f, fInf.Size(), "application/zip").ProgressUpdater(progressUpdater).Do()
		if err != nil {
			log.Fatalf("Unable to create file: %v", err)
			return "", err
		}
		// log.Printf("file: %+v", up)
		fmt.Println("Backup complete!")
		log.Printf("file: %s", up.Id)
	} else {
		fmt.Println("Updating Backup...")
		up, err = srv.Files.Update("1_Xyv51Qb508OHXb5h3NGnUco-bvHd_PS", &drive.File{MimeType: "application/zip"}).ResumableMedia(context.Background(), f, fInf.Size(), "application/zip").ProgressUpdater(progressUpdater).Do()
		if err != nil {
			log.Fatal(err)
			return "", err
		}
		fmt.Println("Update complete!")
	}

	// r, err := srv.Files.List().PageSize(20).
	// 	Fields("nextPageToken, files(id, name)").Do()
	// if err != nil {
	// 	log.Fatalf("Unable to retrieve files: %v", err)
	// 	return "", err
	// }
	// fmt.Println("Files:")
	// if len(r.Files) == 0 {
	// 	fmt.Println("No files found.")
	// } else {
	// 	for _, i := range r.Files {
	// 		fmt.Printf("%s (%s)\n", i.Name, i.Id)
	// 	}
	// }

	return up.Id, nil
}

// func update(srv *drive.Service) {
// 	progressUpdater := func(current, total int64) { log.Printf("%dB / %dB total\n", current, total) }

// 	f, err := os.Open("folder.zip")
// 	if err != nil {
// 		log.Fatalf("Unable to open test.txt: %v", err)
// 		// return "", err
// 	}

// 	fInf, err := f.Stat()
// 	if err != nil {
// 		log.Fatal(err)
// 		// return "", err
// 	}

// 	up, err := srv.Files.Update("1uN_Gqv3xZHjLTgGp_r2EPrl2mpgog8JL", &drive.File{MimeType: "application/zip"}).ResumableMedia(context.Background(), f, fInf.Size(), "application/zip").ProgressUpdater(progressUpdater).Do()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	fmt.Println(up.Id)
// }

func main() {
	ctx := context.Background()
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	currentFileId := "1_Xyv51Qb508OHXb5h3NGnUco-bvHd_PS"

	r := gin.Default()
	r.GET("/backup", func(c *gin.Context) {
		fileId, err := backup(srv, currentFileId)
		if err != nil {
			log.Fatal(err)
			c.JSON(500, gin.H{
				"message": "File backup failed",
			})
		}

		currentFileId = fileId
		// fmt.Println(currentFileId)

		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080
}

// package main

// import (
// 	"fmt"
// 	"io/ioutil"
// 	"net/http"

// 	"github.com/gin-gonic/gin"
// )

// // Retrieves a token from a local file.
// func tokenFromFile(file string) (*oauth2.Token, error) {
// 	f, err := os.Open(file)
// 	if err != nil {
// 			return nil, err
// 	}
// 	defer f.Close()
// 	tok := &oauth2.Token{}
// 	err = json.NewDecoder(f).Decode(tok)
// 	return tok, err
// }

// func initUpload() {
// 	url := "https://www.googleapis.com/upload/drive/v3/files?uploadType=resumable"
// 	method := "POST"

// 	client := &http.Client{}
// 	req, err := http.NewRequest(method, url, nil)

// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	req.Header.Add("Authorization", "Bearer ya29.a0Aa4xrXOWzqEXQUcrYQc_Ck9Z3st6V0TOOedDxi7jmm2_IhZHE3ceAJBPf_ANOXYc4qwd2S2_LTvy3fhi0XuCFp4IYjWmNA0chLiiAToj95N1vZVCOLR1wk3D4J2dQtX7auXk2jU3ApNTIwG9KSY408_SHHALkAaCgYKATASARESFQEjDvL9IQzrDFYoBMp6WtHqrb8Xpg0165")

// 	res, err := client.Do(req)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	defer res.Body.Close()

// 	body, err := ioutil.ReadAll(res.Body)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	fmt.Println(string(body))
// }

// func main() {
// 	r := gin.Default()
// 	r.GET("/upload", func(c *gin.Context) {
// 		c.JSON(200, gin.H{
// 			"message": "pong",
// 		})
// 	})
// 	r.Run() // listen and serve on 0.0.0.0:8080
// }
