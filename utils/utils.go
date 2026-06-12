package utils

import (
	"context"
	"errors"
	"exp1/utils/log"
	"os"
)

func CheckDirExist(path string) bool{
	info, err := os.Stat(path)
	if err != nil{
		if os.IsNotExist(err){
			return false
		}
	}
	if info.IsDir(){
		return true
	}
	return false
}

func CheckFileExist(path string) bool{
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

func CreateFile(ctx context.Context, cancel context.CancelFunc, path string){
	f, err := os.Create(path)
	if err != nil{
		log.Error(ctx, cancel, err.Error())
	}
	f.Close()
}