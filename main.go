package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Item float64

type Stack struct {
	items []Item
	rwLock sync.RWMutex
}

func (stack *Stack) Push(t Item) {
	if stack.items == nil {
		stack.items = []Item{}
	}
	stack.rwLock.Lock()
	stack.items = append(stack.items, t)
	stack.rwLock.Unlock()
}

func (stack *Stack) Pop() *Item {
	if len(stack.items) == 0 {
		return nil
	}
	stack.rwLock.Lock()
	item := stack.items[len(stack.items)-1]
	stack.items = stack.items[0 : len(stack.items)-1]
	stack.rwLock.Unlock()
	return &item
}

func (stack *Stack) Size() int {
	stack.rwLock.RLock()
	defer stack.rwLock.RUnlock()
	return len(stack.items)
}

func (stack *Stack) All() []Item {
	stack.rwLock.RLock()
	defer stack.rwLock.RUnlock()
	return stack.items
}

func (stack *Stack) IsEmpty() bool {
	stack.rwLock.RLock()
	defer stack.rwLock.RUnlock()
	return len(stack.items) == 0
}

func main() {
	http.HandleFunc("/", home)
	http.HandleFunc("/calculate", calculate)

	fmt.Println("server started at localhost:9090")
	http.ListenAndServe(":9090", nil)
}

func home(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		var tmpl = template.Must(template.New("form").ParseFiles("view.html"))
		var err = tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	http.Error(w, "", http.StatusBadRequest)
}

func calculate(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var tmpl = template.Must(template.New("result").ParseFiles("view.html"))
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var data = map[string]interface{}{}
		data["success"] = false

		var exp = r.FormValue("exp")
		elements := strings.Split(exp, " ")
		if len(elements) < 2 {
			if !isNum(elements[0]) {
				data["success"] = false
				data["hasil"] = "HARUS BERUPA ANGKA"
				if err := tmpl.Execute(w, data); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}

			if toFloat(elements[0]) == 0 {
				data["success"] = false
				data["hasil"] = "SELAIN 0"
				if err := tmpl.Execute(w, data); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}

			data["success"] = true
			data["hasil"] = elements[0]
			if err := tmpl.Execute(w, data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		if !isNum(elements[0]) {
			data["hasil"] = "DIGIT PERTAMA & KEDUA HARUS ANGKA"
			if err := tmpl.Execute(w, data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		if !isNum(elements[1]) {
			data["hasil"] = "DIGIT PERTAMA & KEDUA HARUS ANGKA"
			if err := tmpl.Execute(w, data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		if !isOperator(elements[len(elements)-1]) {
			data["hasil"] = "ELEMEN TERAKHIR HARUS OPERATOR"
			if err := tmpl.Execute(w, data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		if isContainsAlphabet(exp) {
			data["hasil"] = "TIDAK BOLEH ADA HURUF"
			if err := tmpl.Execute(w, data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		if !count(elements) {
			data["hasil"] = "EKSPRESI SALAH"
			if err := tmpl.Execute(w, data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		stack := Stack{}
		for _, v := range elements {
			if isNum(v) {
				stack.Push(Item(toFloat(v)))
			}else {
				if stack.IsEmpty() {
					data["success"] = false
					data["hasil"] =  "Ups Stacknya Kosong"
					if err := tmpl.Execute(w, data); err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
					return
				}
				val1 := stack.Pop()

				if stack.IsEmpty() {
					data["success"] = false
					data["hasil"] =  "stack kosong"
					if err := tmpl.Execute(w, data); err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
					return
				}
				val2 := stack.Pop()
				switch v {
					case "+":
						stack.Push((*val2) + (*val1))
					case "-":
						stack.Push((*val2) - (*val1))
					case "*":
						stack.Push((*val2) * (*val1))
					case "/":
						stack.Push((*val2) / (*val1))
				}
			}
		}

		data["success"] = true
		data["hasil"] =  fmt.Sprint(*stack.Pop())
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	http.Error(w, "", http.StatusBadRequest)
}

func toFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func isNum(s string) bool {
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return true
	}
	return false
}

func isOperator(s string) bool {
	return s == "+" || s == "-" || s == "*" || s == "/"
}

func isContainsAlphabet(s string) bool {
	s = strings.TrimSpace(s)
	const alpha = "abcdefghijklmnopqrstuvwxyz"
	for _, c := range s {
		if strings.Contains(alpha, strings.ToLower(string(c))) {
			return true
		}
	}
	return false
}

func count(s []string) bool {
	nAngka := 0
	nOperator := 0

	for _, v := range s {
		if isNum(v) {
			nAngka++
		}
		if isOperator(v) {
			nOperator++
		}
	}

	return nAngka - nOperator == 1
}