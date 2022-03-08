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

func getImage(imageBytes []byte) {
	extractLabels(imageBytes)
}

func extractLabels(imgBytes []byte) {
	img, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		log.Fatalln(err)
	}

	f, err := os.Create("../outimage.jpg")
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	opt := jpeg.Options{
		Quality: 90,
	}
	err = jpeg.Encode(f, img, &opt)
	if err != nil {
	}

	ctx := context.Background()
	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
	}
	defer client.close()

	visionImage, err := vision.NewImageFromReader(img)
	if err != nil {
	}
	labels, err := client.DetectLabels(ctx, visionImage, nil, 5)
	if err != nil {
	}
	fmt.Println("Labels:", labels)

	// TODO
	len := 5 // d세팅한 라벨 수
	myImage.Labels = make([]string, len)
	for _, label := range labels {
		myImage.Labels = append(myImage.Labels, label)
	}
	// TODO : body 로 하여 반환한다.
	// 이미지 일치하는지 여부를 앱이 수신한 MyImageSrc 의 []byte 일치하는지로 검증.
}

type MyImageSrc struct {
	ImageBytes []byte
	Labels     []string
}
