package reconstruct

// import (
// 	"encoding/json"
// 	"exp1/internal/types"
// 	"fmt"
// 	"log"
// 	"os"

// 	"github.com/sergi/go-diff/diffmatchpatch"
// )

// func Reconstruct(path string){
// 	correctPath := "../../.rec/history" + path
// 	entries, err := os.ReadDir(correctPath)
// 	if err != nil{
// 		log.Fatal("error reading the dir (reconstruct.go):",err)
// 	}

// 	fmt.Println("entries:",entries)
// 	var content string

// 	dmp := diffmatchpatch.New()

// 	for _, entry := range entries{
// 		fileContent := string(ReadFile(entry.Name()))
// 		var fileRecord types.FileRecord

// 		err := json.Unmarshal([]byte(fileContent), &fileRecord)
// 		if err != nil{
// 			log.Fatal("error while unmarshel (reconstruct.go):",err)
// 		}
// 		fmt.Println(fileRecord.Type)
// 		fmt.Println(fileRecord.Action)
// 		if fileRecord.Type == "snapshot" && fileRecord.Action == "write"{
// 			fmt.Println(fileRecord.Blob)
// 			blobPath := "../../" + fileRecord.Blob
// 			blobByte, err := os.ReadFile(blobPath)
// 			if err != nil{
// 				log.Fatal("error while reading blob:", err)
// 			}
// 			blobText := string(blobByte)
// 			content = blobText
// 		}else if fileRecord.Type == "delta" && fileRecord.Action == "write"{
// 			blobPath := "../../" + fileRecord.Blob
// 			blobByte, err := os.ReadFile(blobPath)
// 			if err != nil{
// 				log.Fatal("error while reading blob:", err)
// 			}
// 			blobText := string(blobByte)

// 			patch, err := dmp.PatchFromText(blobText)
// 			if err != nil{
// 				log.Fatal("error while converting text to patch:",err)
// 			}

// 			newText, result := dmp.PatchApply(patch, content)
// 			fmt.Println("result for ", entry.Name(), " ",result)
// 			fmt.Println("new text for ", entry.Name(), " ", newText)
// 			content = newText
// 		}else{
// 			continue
// 		}
// 	}

// 	fmt.Println("final content: ", content)
// }

// func ReadFile(name string) []byte{
// 	path := "../../.rec/history/sample/d1/t2.txt/" + name
// 	data, err := os.ReadFile(path)
// 	if err != nil{
// 		log.Fatal(err)
// 	}

// 	return data
// }