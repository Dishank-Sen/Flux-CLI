package roottimeline

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

func Save(data any) error{
	fileName := getFileName()
	filePath := getFilePath(fileName)

	jsonData, err := json.MarshalIndent(data, "", " ")
	if err != nil{
		log.Fatal("error (root-timeline.go): ",err)
		return err
	}

	// fmt.Println("json data:", string(jsonData))

	if err = os.WriteFile(filePath, jsonData, 0644); err != nil{
		// log.Fatal("error (root-timeline.go): ",err)
		return err
	}
	
	return nil
}

func getFileName() string{
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s.json", timestamp)
	return filename
}

func getFilePath(fileName string) string{
	filePath := filepath.Join(".rec", "root-timeline", fileName)
	return filePath
}