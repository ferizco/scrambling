package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"sync"
)

type Contact struct {
	Name  string `json:"nama"`
	Phone string `json:"phone"`
}

var contacts []Contact
var mu sync.Mutex

const jsonFile = "contacts.json"

var substitutionMap = map[rune]rune{
	'0': 'z', '1': 'a', '2': 'f', '3': 'l',
	'4': 'h', '5': 'm', '6': 'r', '7': 'q',
	'8': 's', '9': 'v',
}

func scramblePhone(phone string) string {
	var scrambled []rune

	for _, char := range phone {
		if newChar, found := substitutionMap[char]; found {
			scrambled = append(scrambled, newChar)
		} else {
			scrambled = append(scrambled, char)
		}
	}
	return string(scrambled)
}

func unscramblePhone(scrambled string) string {
	var unscrambled []rune

	for _, char := range scrambled {
		originalChar := char
		for k, v := range substitutionMap {
			if v == char {
				originalChar = k
				break
			}
		}
		unscrambled = append(unscrambled, originalChar)
	}
	return string(unscrambled)
}

func saveContactsToFile() error {
	data, err := json.MarshalIndent(contacts, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(jsonFile, data, 0644)
}

func loadContactsFromFile() error {
	data, err := os.ReadFile(jsonFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(data, &contacts)
}

func mainPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		phone := r.FormValue("phone")

		scrambledPhone := scramblePhone(phone)

		mu.Lock()
		contacts = append(contacts, Contact{Name: name, Phone: scrambledPhone})
		mu.Unlock()

		if err := saveContactsToFile(); err != nil {
			http.Error(w, "Failed to save contacts", http.StatusInternalServerError)
			return
		}
	}

	mu.Lock()
	defer mu.Unlock()

	for i := range contacts {
		contacts[i].Phone = unscramblePhone(contacts[i].Phone)
	}

	tmpl := template.Must(template.ParseFiles("index.html"))
	tmpl.Execute(w, contacts)
}

func main() {
	if err := loadContactsFromFile(); err != nil {
		panic(err)
	}

	http.HandleFunc("/", mainPage)
	http.ListenAndServe(":8080", nil)
}
