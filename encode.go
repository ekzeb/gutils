package util

import (
	"log"
	"os"
	"io/ioutil"
	"encoding/json"
	"bytes"
	"encoding/gob"
)


// store gob data
func StoreGob(data interface{}, filename string, fileMode os.FileMode) (err error) {
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)
	err = encoder.Encode(data)

	if err != nil {
		log.Println("Error Encode GOB data:", err)
	}

	err = ioutil.WriteFile(filename, buffer.Bytes(), fileMode)

	if err != nil {
		log.Println("Error Store GOB data:", err)
	}

	return
}
// store json data
func StoreJson(data interface{}, filename string, fileMode os.FileMode) (err error)  {

	b, err := json.Marshal(data)
	if err != nil {
		log.Println("Error Marshal JSON data:", err)

	}

	if err := ioutil.WriteFile(filename, b, fileMode); err != nil {
		log.Println("Error write JSON data:", err)

	}

	return
}
// load gob data
func LoadGob(data interface{}, filename string) (err error) {
	raw, err := ioutil.ReadFile(filename)

	if err != nil {
		log.Println("Error Load GOB data:", err)
	}

	buffer := bytes.NewBuffer(raw)

	dec := gob.NewDecoder(buffer)

	err = dec.Decode(data)
	if err != nil {
		log.Println("Error Decode GOD data:", err)
	}

	return
}
// load json data
func LoadJson(data interface{}, filename string) (err error) {
	jsonFile, err := os.Open(filename)
	if err != nil {
		log.Println("Error opening JSON file:", err)
	}
	defer jsonFile.Close()

	jsonData, err := ioutil.ReadAll(jsonFile)

	if err != nil {
		log.Println("Error reading JSON data:", err)
	}

	if err := json.Unmarshal(jsonData, data); err != nil {
		log.Println("Error Unmarshal JSON data:", err)
	}

	return
}
