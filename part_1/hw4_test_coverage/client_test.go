package main

// код писать тут

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

type User_xml struct {
	ID         int    `xml:"id"`
	First_name string `xml:"first_name"`
	Last_name  string `xml:"last_name"`
	Age        int    `xml:"age"`
	About      string `xml:"about"`
	Gender     string `xml:"gender"`
}

type Users struct {
	Version    string     `xml:"version,attr"`
	Users_list []User_xml `xml:"row"`
}

func read_xml(file_name string) (*Users, error) {
	users := new(Users)
	file, _ := os.Open(file_name)
	defer file.Close()
	xmlData, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Printf("error: %v", err)
		return users, err
	}
	err = xml.Unmarshal(xmlData, &users)
	if err != nil {
		fmt.Printf("error: %v", err)
		return users, err
	}
	return users, nil

}

type SortBy []User
type SortByName struct {
	SortBy
}
type SortById struct {
	SortBy
}
type SortByAge struct {
	SortBy
}

func (obj SortBy) Len() int {
	return len(obj)
}
func (obj SortBy) Swap(i, j int) {
	temp := obj[j]
	obj[j] = obj[i]
	obj[i] = temp
}
func (obj SortByName) Less(i, j int) bool {
	res := strings.Compare(obj.SortBy[i].Name, obj.SortBy[j].Name)
	return res < 0
}
func (obj SortById) Less(i, j int) bool {
	return obj.SortBy[i].Id < obj.SortBy[j].Id
}
func (obj SortByAge) Less(i, j int) bool {
	return obj.SortBy[i].Age < obj.SortBy[j].Age
}

func remove(s []User, i int) []User {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func find_users(users_data_xml *Users, r *http.Request) ([]byte, error) {
	data := []User{}
	limit, _ := strconv.Atoi(r.FormValue("limit"))
	offset, _ := strconv.Atoi(r.FormValue("offset"))
	query := r.FormValue("query")
	order_field := r.FormValue("order_field")
	order_by, _ := strconv.Atoi(r.FormValue("order_by"))
	for _, user := range users_data_xml.Users_list {
		if limit == 0 {
			break
		}
		name := user.First_name + " " + user.Last_name
		if strings.Contains(name, query) || strings.Contains(user.About, query) || query == "" {
			user_src := User{user.ID, name, user.Age, user.About, user.Gender}
			data = append(data, user_src)
			if query != "" {
				limit--
			}
		}
	}
	if order_field == "" {
		order_field = "Name"
	}
	order_sort := func(data_ sort.Interface, order int) {
		if order == 1 {
			sort.Sort(data_)
		} else {
			sort.Sort(sort.Reverse(data_))
		}
	}

	if order_by != 0 {
		switch order_field {
		case "Id":
			order_sort(SortById{data}, order_by)
		case "Age":
			order_sort(SortByAge{data}, order_by)
		case "Name":
			order_sort(SortByName{data}, order_by)
		default:
			return nil, fmt.Errorf("ErrorBadOrderField")
		}
	} else if order_by != 1 && order_by != -1 {
		return nil, fmt.Errorf("ErrorBadOrderBy")
	}
	for i := 0; i < offset; i++ {
		data = remove(data, i)
	}
	result, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("error, %v", err)
		return nil, err
	}

	return result, err
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	acces_token := r.Header.Get("AccessToken")
	if acces_token != "test_token" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	users, _ := read_xml("dataset.xml")
	req, err := find_users(users, r)
	if err != nil {
		if err.Error() == "ErrorBadOrderField" {
			w.WriteHeader(http.StatusBadRequest)
			err_ := SearchErrorResponse{"ErrorBadOrderField"}
			json_err, _ := json.Marshal(err_)
			w.Write(json_err)
		}
		if err.Error() == "ErrorBadOrderBy" {
			w.WriteHeader(http.StatusBadRequest)
			err_ := SearchErrorResponse{"ErrorBadOrderBy"}
			json_err, _ := json.Marshal(err_)
			w.Write(json_err)
		}

		return
	}
	w.Write(req)

	//fmt.Println(req)

}

func TestBrokenServersFunc(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()
	access_token := "test_token"
	TestRequest := SearchRequest{
		Limit:      10,
		Offset:     1,
		Query:      "J",
		OrderField: "Age",
		OrderBy:    1,
	}
	sc := &SearchClient{access_token, ts.URL}
	sc.FindUsers(TestRequest)
	test_handler_timeout := func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(client.Timeout)
		SearchServer(w, r)
	}
	ts2 := httptest.NewServer(http.HandlerFunc(test_handler_timeout))
	defer ts2.Close()
	sc2 := &SearchClient{access_token, ts2.URL}
	sc2.FindUsers(TestRequest)
	test_handler_error_json := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		err_ := "error_json"
		json_err, _ := json.Marshal(err_)
		w.Write(json_err)
		SearchServer(w, r)
	}
	ts3 := httptest.NewServer(http.HandlerFunc(test_handler_error_json))
	defer ts3.Close()
	sc3 := &SearchClient{access_token, ts3.URL}
	sc3.FindUsers(TestRequest)
	test_handler_fatal_error := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		SearchServer(w, r)
	}
	ts4 := httptest.NewServer(http.HandlerFunc(test_handler_fatal_error))
	defer ts4.Close()
	sc4 := &SearchClient{access_token, ts4.URL}
	sc4.FindUsers(TestRequest)
	test_header_unckown_error := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(302)
		w.Header().Set("Location", "")
	}
	ts5 := httptest.NewServer(http.HandlerFunc(test_header_unckown_error))
	defer ts5.Close()
	sc5 := &SearchClient{access_token, ts5.URL}
	sc5.FindUsers(TestRequest)
}

func TestFindUsers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	sc := &SearchClient{"mistake_token", ts.URL}
	sc.FindUsers(SearchRequest{})
	access_token := "test_token"
	sc = &SearchClient{access_token, ts.URL}

	TestRequest := []SearchRequest{
		{
			Limit:      10,
			Offset:     1,
			Query:      "J",
			OrderField: "Age",
			OrderBy:    1,
		},
		{
			Limit: -1,
		},
		{
			Limit: 26,
		},
		{
			Offset: -1,
		},
		{
			Limit:      10,
			Offset:     1,
			Query:      "J",
			OrderField: "mistake",
			OrderBy:    1,
		},
		{
			Limit:      10,
			Offset:     1,
			Query:      "J",
			OrderField: "Age",
			OrderBy:    -10,
		},
		{
			Limit:      1,
			Offset:     0,
			Query:      "J",
			OrderField: "Age",
			OrderBy:    1,
		},

		{
			Limit:      10,
			Offset:     1,
			Query:      "10",
			OrderField: "10",
			OrderBy:    -10,
		},
	}
	for _, test_req := range TestRequest {
		sc.FindUsers(test_req)
	}

}
