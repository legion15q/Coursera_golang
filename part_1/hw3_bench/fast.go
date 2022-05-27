package main

import (
	"bufio"
	json "encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

type (
	Json_struct struct {
		Browsers []string `json:"browsers"`
		Company  string   `json:"company"`
		Country  string   `json:"country"`
		Email    string   `json:"email"`
		Job      string   `json:"job"`
		Name     string   `json:"name"`
		Phone    string   `json:"phone"`
	}
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjsonE0340b5dDecodeTestCodegenJson(in *jlexer.Lexer, out *Json_struct) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "browsers":
			if in.IsNull() {
				in.Skip()
				out.Browsers = nil
			} else {
				in.Delim('[')
				if out.Browsers == nil {
					if !in.IsDelim(']') {
						out.Browsers = make([]string, 0, 4)
					} else {
						out.Browsers = []string{}
					}
				} else {
					out.Browsers = (out.Browsers)[:0]
				}
				for !in.IsDelim(']') {
					v1 := string(in.String())
					out.Browsers = append(out.Browsers, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		case "company":
			out.Company = string(in.String())
		case "country":
			out.Country = string(in.String())
		case "email":
			out.Email = string(in.String())
		case "job":
			out.Job = string(in.String())
		case "name":
			out.Name = string(in.String())
		case "phone":
			out.Phone = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonE0340b5dEncodeTestCodegenJson(out *jwriter.Writer, in Json_struct) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"browsers\":"
		out.RawString(prefix[1:])
		if in.Browsers == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v2, v3 := range in.Browsers {
				if v2 > 0 {
					out.RawByte(',')
				}
				out.String(string(v3))
			}
			out.RawByte(']')
		}
	}
	{
		const prefix string = ",\"company\":"
		out.RawString(prefix)
		out.String(string(in.Company))
	}
	{
		const prefix string = ",\"country\":"
		out.RawString(prefix)
		out.String(string(in.Country))
	}
	{
		const prefix string = ",\"email\":"
		out.RawString(prefix)
		out.String(string(in.Email))
	}
	{
		const prefix string = ",\"job\":"
		out.RawString(prefix)
		out.String(string(in.Job))
	}
	{
		const prefix string = ",\"name\":"
		out.RawString(prefix)
		out.String(string(in.Name))
	}
	{
		const prefix string = ",\"phone\":"
		out.RawString(prefix)
		out.String(string(in.Phone))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Json_struct) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonE0340b5dEncodeTestCodegenJson(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Json_struct) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonE0340b5dEncodeTestCodegenJson(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Json_struct) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonE0340b5dDecodeTestCodegenJson(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Json_struct) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonE0340b5dDecodeTestCodegenJson(l, v)
}

var pool = sync.Pool{
	New: func() interface{} {
		return &Json_struct{}
	},
}

// вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	fmt.Fprint(out, "", "found users:\n")
	//foundUsers := ""
	fileScanner := bufio.NewScanner(file)

	uniq_browsers := make(map[string]int, 100)
	i := 0
	for fileScanner.Scan() {
		user := pool.Get().(*Json_struct)
		line := fileScanner.Bytes()
		user.UnmarshalJSON(line)
		if err != nil {
			panic(err)
		}
		isAndroid, isMSIE := false, false
		for _, browser := range user.Browsers {
			if strings.Count(browser, "Android") > 0 {
				uniq_browsers[browser]++
				isAndroid = true
			}
			if strings.Count(browser, "MSIE") > 0 {
				uniq_browsers[browser]++
				isMSIE = true
			}
		}
		if isAndroid && isMSIE {
			email := strings.Replace(user.Email, "@", " [at] ", 1)
			fmt.Fprintf(out, "[%d] %s <%s>\n", i, user.Name, email)
		}
		i++
		pool.Put(user)
	}

	if err := fileScanner.Err(); err != nil {
		log.Fatalf("Error while reading file: %s", err)
	}

	fmt.Fprint(out, "", "\n")
	fmt.Fprintln(out, "Total unique browsers", len(uniq_browsers))

}
