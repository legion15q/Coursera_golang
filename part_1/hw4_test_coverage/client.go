package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	orderAsc = iota
	orderDesc
)

var (
	errTest = errors.New("testing")
	//http клиент через который отправляется запрос. По типу http.Post
	client = &http.Client{Timeout: time.Second}
)

//структура определяющая формат полученный данных
type User struct {
	Id int
	//name складывается из first_name + " " + last_name
	Name   string
	Age    int
	About  string
	Gender string
}

//результат который мы должны получить
type SearchResponse struct {
	Users    []User
	NextPage bool
}

type SearchErrorResponse struct {
	Error string
}

/*ASC | DESC Указывает порядок сортировки значений в указанном столбце — по возрастанию или по убыванию. Значение ASC сортирует от низких значений к высоким.*/
const (
	OrderByAsc  = -1
	OrderByAsIs = 0
	OrderByDesc = 1

	ErrorBadOrderField = `OrderField invalid`
)

//стрктура определяющая параметры запроса и то, что нам нужно найти
type SearchRequest struct {
	Limit  int    //количество записей, которое требуется найти (вернет сервер)
	Offset int    // Сколько записей пропустить/не учитывать от полученных значений (после сортировки)
	Query  string //Сам запрос -- то есть подстроки, которые нам нужно найти. Ищет по полям `Name` и `About`.  `Name` - это first_name + " " + last_name из xml.
	//Если `query` пустой, то делаем только сортировку, т.е. возвращаем все записи
	OrderField string //Параметр сортировки. То есть сортировать по полям `Id`, `Age` или `Name`. Если пустой - то сортируем по `Name`.
	// Если что-то другое - SearchServer ругается ошибкой.
	OrderBy int // -1 -- сортировка по убыванию, 0 -- без сортировки, 1 -- сортировка по возрастанию
}

//структура определяющая куда мы отправляем запрос и задает уникальный токен клиента (это уже делает курсера)
type SearchClient struct {
	// токен, по которому происходит авторизация на внешней системе, уходит туда через хедер
	AccessToken string
	// урл внешней системы, куда идти
	URL string
}

// FindUsers отправляет запрос во внешнюю систему, которая непосредственно ищет пользоваталей
func (srv *SearchClient) FindUsers(req SearchRequest) (*SearchResponse, error) {
	//создаем структуру из пакета url для заполнения параметров http запроса. Важно понимать чем параметры отличаются от хедеров
	/*Хедеры содержат метаинформацию, а параметры содержат фактические данные.
	HTTP-серверы автоматически отменяют/декодируют имена/значения параметров. Это не относится к именам/значениям хедеров.
	Имена / значения хедеров должны быть экранированы / закодированы вручную на стороне клиента и вручную разэкранированы / декодированы
	на стороне сервера. Часто используется кодировка Base64 или процентное экранирование.
	Параметры могут быть видны конечным пользователям (в URL), но заголовки скрыты для конечных пользователей.
	*/
	searcherParams := url.Values{}

	if req.Limit < 0 {
		return nil, fmt.Errorf("limit must be > 0")
	}
	if req.Limit > 25 {
		req.Limit = 25
	}
	if req.Offset < 0 {
		return nil, fmt.Errorf("offset must be > 0")
	}

	//нужно для получения следующей записи, на основе которой мы скажем - можно показать переключатель следующей страницы или нет
	//Грубо говоря если у нас на html странице вмещается n записей, и мы отправляем запрос на n записей, то и стрелочку для переключения страницы нет смысла отображать.
	// А если существует n+1 записей, то уже нужно
	req.Limit++
	//заполняем параметры запроса
	searcherParams.Add("limit", strconv.Itoa(req.Limit))
	searcherParams.Add("offset", strconv.Itoa(req.Offset))
	searcherParams.Add("query", req.Query)
	searcherParams.Add("order_field", req.OrderField)
	searcherParams.Add("order_by", strconv.Itoa(req.OrderBy))
	//задаем тип запроса и добавляем к нему параметры
	searcherReq, err := http.NewRequest("GET", srv.URL+"?"+searcherParams.Encode(), nil)
	//добавляем хедер к get запросу, а именно accesstoken
	searcherReq.Header.Add("AccessToken", srv.AccessToken)
	//client.do -- отправляет запрос. Принимает ответ сервера resp и ошибку err
	resp, err := client.Do(searcherReq)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return nil, fmt.Errorf("timeout for %s", searcherParams.Encode())
		}
		return nil, fmt.Errorf("unknown error %s", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("Bad AccessToken")
	case http.StatusInternalServerError:
		return nil, fmt.Errorf("SearchServer fatal error")
	case http.StatusBadRequest:
		errResp := SearchErrorResponse{}
		err = json.Unmarshal(body, &errResp)
		if err != nil {
			return nil, fmt.Errorf("cant unpack error json: %s", err)
		}
		if errResp.Error == "ErrorBadOrderField" {
			return nil, fmt.Errorf("OrderFeld %s invalid", req.OrderField)
		}
		return nil, fmt.Errorf("unknown bad request error: %s", errResp.Error)
	}
	//создаем слайс юзеров (не один юзер)
	data := []User{}
	//распаковываем полученный json в массив (слайс) пользователей data
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("cant unpack result json: %s", err)
	}

	result := SearchResponse{}
	if len(data) == req.Limit {
		result.NextPage = true
		result.Users = data[0 : len(data)-1]
	} else {
		result.Users = data[0:len(data)]
		//result.NextPage по умолчанию false
	}

	return &result, err
}
