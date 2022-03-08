package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	vision "cloud.google.com/go/vision/apiv1"
)

var myImage MyImageSrc

func handleRoot(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Hello Go2!")

	fmt.Println("Method: ", req.Method)
	fmt.Println("Url: ", req.URL)
	fmt.Println("Header: ", req.Header)

	b, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	fmt.Println("Body: ", string(b))

	w.Write([]byte("Post request succeeded."))
}

func handleImage(w http.ResponseWriter, req *http.Request) {
	body, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	myImage.ImageBytes = body
	// json.Unmarshal(body, &myImage)
	extractLabels(body)
}

func main() {
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/GetLabels", handleImage)

	res := http.ListenAndServe(":8080", nil)
	log.Fatal(res)
}

func extractLabels(imgBytes []byte) {
	img, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		log.Fatalln(err)
	}

	fileName := "../outimage.jpg"

	f, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("Failed to create a file outimage.jpg: %v", err)
	}
	defer f.Close()

	opt := jpeg.Options{
		Quality: 90,
	}
	err = jpeg.Encode(f, img, &opt)
	if err != nil {
		log.Fatalf("Failed to encode image to JPG: %v", err)
	}

	ctx := context.Background()
	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}
	defer file.Close()

	visioniamge, err := vision.NewImageFromReader(file)
	if err != nil {
		log.Fatalf("Failed to read image from reader: %v", err)
	}
	labelsSize := 5
	labels, err := client.DetectLabels(ctx, visioniamge, nil, labelsSize)
	if err != nil {
		// TODO : error - failed to detect errors
		log.Fatalf("Failed to detect labels: %v", err)
	}

	myImage.Labels = make([]string, labelsSize)
	for _, label := range labels {
		myImage.Labels = append(myImage.Labels, label.GetDescription())
	}

	// TODO : body 로 하여 반환한다.
	// 이미지 일치하는지 여부를 앱이 수신한 MyImageSrc 의 []byte 일치하는지로 검증.
}

type MyImageSrc struct {
	ImageBytes []byte
	Labels     []string
}
