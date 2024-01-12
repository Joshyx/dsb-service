package main

import (
	"bytes"
	"image"
	"image/draw"
	"image/jpeg"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/irgendwr/dsb-go"
)

func main() {
	account := dsb.NewAccount("236661", "GOmobile")
	content, err := account.GetContent()

	if err != nil {
		// exit on error
		log.Printf("Error: %s", err)
		os.Exit(1)
	}

	// get timetables
	timetables := content.GetTimetables()
	if len(timetables) == 0 {
		log.Println("no timetables found")
		return
	}
    todays := Filter(timetables, func(mi dsb.MenuItem) bool {
        return strings.Contains(mi.Title, "heute")
    })
    tomorrows := Filter(timetables, func(mi dsb.MenuItem) bool {
        return strings.Contains(mi.Title, "morgen")
    })
	todaysImages := Map(todays, func(item dsb.MenuItem) image.Image {
		return loadImageFromURL(item.GetURL())
	})
    tomorrowsImages := Map(tomorrows, func(item dsb.MenuItem) image.Image {
		return loadImageFromURL(item.GetURL())
	})
    img1 := mergeVertical(todaysImages)
    img2 := mergeVertical(tomorrowsImages)
    
	buf := new(bytes.Buffer)
	jpeg.Encode(buf, mergeTwoHorizontal(img1, img2), nil)
	data := buf.Bytes()

	timetableHandler := func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Length", strconv.Itoa(len(data)))
		w.Write(data)
	}

	http.HandleFunc("/timetables", timetableHandler)
	log.Println("Listing for requests at http://localhost:8000/timetables")
	log.Fatal(http.ListenAndServe(":8000", nil))

}

func loadImageFromURL(URL string) image.Image {
	//Get the response bytes from the url
	response, err := http.Get(URL)
	if err != nil {
		return nil
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil
	}

	image, _, _ := image.Decode(response.Body)
	return image
}

func mergeHorizontal(imgs []image.Image) image.Image {
	if len(imgs) < 1 {
		log.Println("No images found")
		return nil
	}
	img := imgs[0]
	for _, newImg := range imgs[1:] {
		img = mergeTwoHorizontal(img, newImg)
	}
	return img
}

func mergeTwoHorizontal(img1, img2 image.Image) image.Image {
	sp2 := image.Point{img1.Bounds().Dx(), 0}

	r2 := image.Rectangle{sp2, sp2.Add(img2.Bounds().Size())}
	r := image.Rectangle{image.Point{0, 0}, r2.Max}
	rgba := image.NewRGBA(r)

	draw.Draw(rgba, img1.Bounds(), img1, image.Point{0, 0}, draw.Src)
	draw.Draw(rgba, r2, img2, image.Point{0, 0}, draw.Src)

	return rgba
}

func mergeVertical(imgs []image.Image) image.Image {
	if len(imgs) < 1 {
		log.Println("No images found")
		return nil
	}
	img := imgs[0]
	for _, newImg := range imgs[1:] {
		img = mergeTwoVertical(img, newImg)
	}
	return img
}

func mergeTwoVertical(img1, img2 image.Image) image.Image {
	sp2 := image.Point{0, img1.Bounds().Dy()}

	r2 := image.Rectangle{sp2, sp2.Add(img2.Bounds().Size())}
	r := image.Rectangle{image.Point{0, 0}, r2.Max}
	rgba := image.NewRGBA(r)

	draw.Draw(rgba, img1.Bounds(), img1, image.Point{0, 0}, draw.Src)
	draw.Draw(rgba, r2, img2, image.Point{0, 0}, draw.Src)

	return rgba
}
func Map[T interface{}, U interface{}](vs []T, f func(T) U) []U {
	vsm := make([]U, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}
func Filter[T interface{}](vs []T, f func(T) bool) []T {
    vsm := make([]T, 0)
	for _, v := range vs {
		if f(v) {
            vsm = append(vsm, v)
        }
	}
	return vsm
}
