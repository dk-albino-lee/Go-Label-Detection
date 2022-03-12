package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"os"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/db"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"

	vision "cloud.google.com/go/vision/apiv1"
)

var myImage MyImageSrc
var imgRef *db.Ref

func main() {
	ctx := context.Background()
	conf := &firebase.Config{
		DatabaseURL: "https://story-of-your-things-default-rtdb.asia-southeast1.firebasedatabase.app/",
	}
	opt := option.WithCredentialsFile("~/Users/dongkeunalbinolee/Downloads/robust-chess-337506-5fabd188701b.json")

	app, err := firebase.NewApp(ctx, conf, opt)
	if err != nil {
		log.Fatalf("error initialising app: %v\n", err)
	}

	databaseClient, err := app.Database(ctx)
	if err != nil {
		log.Fatalln("Error initialising database client:", err)
	}

	ref := databaseClient.NewRef("restricted_access/secret_document")
	var data map[string]interface{}
	if err := ref.Get(ctx, &data); err != nil {
		log.Fatal("Error reading from database:", err)
	}
	fmt.Println(data)

	imgRef := ref.Child("images")

	http.HandleFunc("/GetLabels", handleImage)

	res := http.ListenAndServe(":8080", nil)
	log.Fatal(res)
}

func handleImageWithGin(context *gin.Context) {

}

func handleImage(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Url: ", req.URL)
	fmt.Println("Body: ", req.Body)
	defer req.Body.Close()

	decoder := json.NewDecoder(req.Body)
	var ra ReceivedArgument
	err := decoder.Decode(&ra)
	if err != nil {
		panic(err)
	}

	myImage = extractLabels(ra.ImageBytes)
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	resp, err := json.Marshal(myImage)
	errorHandling(err, "Error happened in JSON marshal")
	w.Write(resp)
}

// func handleImage(w http.ResponseWriter, req *http.Request) {
// 	body, _ := ioutil.ReadAll(req.Body)
// 	defer req.Body.Close()
// 	myImage.ImageBytes = body
// 	// json.Unmarshal(body, &myImage)
// 	extractLabels(body)
// }

func extractLabels(imgBytes []byte) MyImageSrc {
	var result MyImageSrc
	result.ImageBytes = imgBytes

	err := imgRef.Set(context.Background(), map[string]*ReceivedArgument{
		"Image": {
			ImageBytes: imgBytes,
		},
	})
	if err != nil {
		log.Fatalln("error setting value:", err)
	}

	img, _, err := image.Decode(bytes.NewReader(imgBytes))
	errorHandling(err, "Failed to decode bytes to image")

	fileName := "../outimage.jpg"

	f, err := os.Create(fileName)
	errorHandling(err, "Failed to create a file outimage.jpg")
	defer f.Close()

	opt := jpeg.Options{
		Quality: 90,
	}
	err = jpeg.Encode(f, img, &opt)
	errorHandling(err, "Failed to encode image to JPG")

	ctx := context.Background()
	client, err := vision.NewImageAnnotatorClient(ctx)
	errorHandling(err, "Failed to create client")
	defer client.Close()

	file, err := os.Open(fileName)
	errorHandling(err, "Failed to read file")
	defer file.Close()

	visioniamge, err := vision.NewImageFromReader(file)
	errorHandling(err, "Failed to read image from reader")

	labelsSize := 5
	labels, err := client.DetectLabels(ctx, visioniamge, nil, labelsSize)
	errorHandling(err, "Failed to detect labels")

	result.Labels = make([]string, labelsSize)
	for _, label := range labels {
		result.Labels = append(result.Labels, label.GetDescription())
	}

	// 현재로서는 vision 에서 코드 내 이미지 읽는 방법을 모르겠음.
	// 이미지 읽어들인 후 바로 삭제하는 방식으로 한다.
	resultErr := os.Remove(fileName)
	errorHandling(resultErr, "Failed to delete image file")

	return result
}

func errorHandling(err error, message string) {
	if err != nil {
		log.Fatalf(message+": %v", err)
	}
}

type ReceivedArgument struct {
	ImageBytes []byte `json:"ImageBytes"`
}

type MyImageSrc struct {
	ImageBytes []byte
	Labels     []string
}
