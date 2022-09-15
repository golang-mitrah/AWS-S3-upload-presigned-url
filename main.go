package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

var (
	routerPort   = ":8081"
	Region       string
	Bucket       string
	Access_Key   string
	Secret_Key   string
	Content_Type = "multipart/form-data"
)

func init() {
	err := godotenv.Load("config.env")
	if err != nil {
		return
	}
	Region, Bucket, Access_Key, Secret_Key = os.Getenv("region"), os.Getenv("bucket"), os.Getenv("access_key"), os.Getenv("secret_key")
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/mitrahsoft", mitrahsoft).Methods(http.MethodPost)
	http.ListenAndServe(routerPort, router)
}

func mitrahsoft(w http.ResponseWriter, r *http.Request) {
	// Local file
	byteFile, err := ioutil.ReadFile("test.txt")
	if ErrorCheck(w, err) {
		return
	}
	signedURL, err := NewUploader(r, byteFile, "./test_file/test.txt")
	if ErrorCheck(w, err) {
		return
	}
	fmt.Println("url:", signedURL)
	ResponseJson(w, signedURL)
}

func NewUploader(r *http.Request, fileBytes []byte, filename string) (string, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(Region),
		Credentials: credentials.NewStaticCredentials(Access_Key, Secret_Key, ""),
	})
	fmt.Println("1.", err)
	if err != nil {
		return "", err
	}
	uploader := s3manager.NewUploader(sess)

	_, err = uploader.UploadWithContext(r.Context(), &s3manager.UploadInput{
		Bucket:      aws.String(Bucket),             // bucket's name
		Key:         aws.String("file/" + filename), // files destination location
		Body:        bytes.NewReader(fileBytes),     // content of the file
		ContentType: aws.String(Content_Type),       // content type
	})
	fmt.Println("2.", err)
	if err != nil {
		return "", err
	}
	signedUrl, err := GetPresignedURL(sess, &Bucket, aws.String("file/"+filename))
	fmt.Println("4.", err)

	if err != nil {
		return "", err
	}
	return signedUrl, nil
}

func GetPresignedURL(sess *session.Session, bucket, key *string) (string, error) {
	svc := s3.New(sess)

	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: bucket,
		Key:    key,
	})
	urlStr, err := req.Presign(15 * time.Minute) // URL valid in 15 minutes
	fmt.Println("3.", err)
	if err != nil {
		return "", err
	}
	return urlStr, nil
}

func ResponseJson(w http.ResponseWriter, url string) {
	jsn := map[string]string{
		"url": url,
	}
	jsnByte, _ := json.Marshal(jsn)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(jsnByte)
}

// This function is used to check error
func ErrorCheck(w http.ResponseWriter, err error) bool {
	// Error handling
	if err != nil {
		jsn := map[string]string{
			"data":  "Nil",
			"error": err.Error(),
		}
		response, _ := json.Marshal(jsn)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(response)
		return true
	}
	return false
}
