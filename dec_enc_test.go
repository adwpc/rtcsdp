package sdp

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"testing"
)

func TestDecode(t *testing.T) {
	files, _ := ioutil.ReadDir("dec")
	for _, fi := range files {
		if !fi.IsDir() {
			fmt.Println("\n=================", filepath.Join("dec/"+fi.Name()), "===================")
			b, _ := ioutil.ReadFile(filepath.Join("dec/" + fi.Name()))
			str := string(b)
			fmt.Print()
			fmt.Println("=========================decoding============================")
			var sdp SDP
			if err := sdp.Decode(str); err != nil {
				log.Println(err.Error())
			}
			fmt.Println("=============================================================")
			fmt.Println("=========================encoding============================")
			enc := sdp.Encode()
			fmt.Print(enc)
			if err := ioutil.WriteFile(filepath.Join("enc/"+fi.Name()), []byte(enc), 0666); err != nil {
				log.Println(err)
			}
			fmt.Println("=============================================================")
		}
	}
}
