package fileupload

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type UploadFileInterface interface {
	UploadFile(*multipart.FileHeader) (string, error)
}

type fileUpload struct{}

// So waht is exposed is Uploader
var _ UploadFileInterface = &fileUpload{}

func NewFileUpload() *fileUpload {
	return &fileUpload{}
}

func (fu *fileUpload) UploadFile(file *multipart.FileHeader) (string, error) {
	f, err := file.Open()
	if err != nil {
		return "", errors.New("cannot open file")
	}
	defer f.Close()

	size := file.Size
	// the image should not be more than 500KB
	fmt.Println("the size: ", size)
	if size > int64(512000) {
		return "", errors.New("sorry, please upload an Image of 500KB or less")
	}

	// only the first 512 bytes ar used to sniff the content type of a file,
	// so, so no need to read the entire bytes of a file
	buffer := make([]byte, size)
	f.Read(buffer)
	fileType := http.DetectContentType(buffer)
	// if the imiage is valid
	if !strings.HasPrefix(fileType, "image") {
		return "", errors.New("please upload a valid image")
	}

	filePath := FormatFile(file.Filename)
	accessKey := os.Getenv("DO_SPACES_KEY")
	secKey := os.Getenv("DO_SPACES_SECRET")
	endpoint := os.Getenv("DO_SPACES_ENDPOINT")
	ssl := true

	// initiate a client using DigitalOcean spaces.
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secKey, ""),
		Secure: ssl,
	})
	if err != nil {
		log.Fatal(err)
	}
	fileBytes := bytes.NewReader(buffer)
	cacheControl := "max-age=31536000"
	// make it public
	userMetadata := map[string]string{"x-amz-acl": "public-read"}
	n, err := client.PutObject(context.Background(), "chodapi", filePath, fileBytes, size, minio.PutObjectOptions{
		ContentType:  fileType,
		CacheControl: cacheControl,
		UserMetadata: userMetadata,
	})
	if err != nil {
		fmt.Println("the error", err)
		return "", errors.New("something went wrong")
	}
	fmt.Println("succes fully uploaded bytes:", n)
	return filePath, nil
}
