package writer

import (
    "os"
    "log"
    _ "fmt"
    "encoding/csv"
    s "github.com/mohakkataria/kfit-scraper/scraper"
    _ "github.com/gocarina/gocsv"

)

var file *os.File = nil

func createFile() *os.File {
    file, err := os.Create("result.csv")
    checkError("Cannot create file", err)
    return file
}

func Write(data []s.Partner) {
    
    if (file == nil) {
        file = createFile()
    }

    writer := csv.NewWriter(file)
    defer writer.Flush()

    for _,b := range data {
        err := writer.Write(b.GetSerializedData())
        checkError("Cannot write to file", err)
    }
    
}

func checkError(message string, err error) {
    if err != nil {
        log.Fatal(message, err)
    }
}