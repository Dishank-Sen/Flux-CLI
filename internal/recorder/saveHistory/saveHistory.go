package savehistory

import (
	"encoding/json"
	"exp1/internal/types"
	"os"
	"path/filepath"
)

func Save(data types.Write) error{
	title := data.Timestamp.Format("20060102_150405") + ".json"
	filePath := filepath.Join(".rec", "history", title)

	jsonData, err := json.Marshal(data)
	if err != nil{
		return err
	}

	return  os.WriteFile(filePath, jsonData, 0644)
}