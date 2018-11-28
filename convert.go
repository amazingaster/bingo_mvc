package bingo_mvc

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	utils "github.com/aosfather/bingo_utils"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

/*
默认返回转换器
1、根据返回类型来进行转换
2、ModelView-> 走template转换
3、其它类型->走json
4、文件流的支持？
5、xml的支持?
6、图片?
*/
type defaultResponseConverter struct {
	templateDir string
	engine      *TemplateEngine
}

func (this *defaultResponseConverter) setTemplateDir(dir string) {
	if this.engine == nil {
		this.engine = new(TemplateEngine)

	}
	this.engine.RootPath = dir

}

func (this *defaultResponseConverter) Convert(writer http.ResponseWriter, obj interface{}, req *http.Request) {
	if mv, ok := obj.(ModelView); ok {
		writer.Header().Add(_CONTENT_TYPE, _CONTENT_HTML+";charset=utf-8")
		this.writeWithTemplate(writer, mv.View, mv.Model)
	} else if rv, ok := obj.(StaticView); ok { //静态资源处理
		writeUseFile(writer, rv)

	} else if rv, ok := obj.(string); ok {
		writer.Write([]byte(rv))
	} else if rv, ok := obj.(RedirectEntity); ok { //处理跳转
		for _, cookie := range rv.Cookies { //设置cookie
			http.SetCookie(writer, cookie)
		}

		http.Redirect(writer, req, rv.Url, rv.Code)
	} else {

		writeUseJson(writer, obj)
	}
}

func (this *defaultResponseConverter) writeWithTemplate(writer http.ResponseWriter, templateName string, obj interface{}) {
	if this.engine == nil {
		this.setTemplateDir("")
	}
	this.engine.Render(writer, templateName, obj)
}
func writeUseFile(writer http.ResponseWriter, rv StaticView) {
	writer.Header().Add(_CONTENT_TYPE, rv.Media)
	writer.Header().Add("Cache-Control", "max-age=2592000")
	//	writer.Header().Add("Content-Disposition", "attachment;fileName="+rv.Name)

	defer rv.Reader.Close()
	length, err := io.Copy(writer, rv.Reader)

	if err != nil {
		writer.WriteHeader(500)
		writer.Write([]byte(err.Error()))
	}

	writer.Header().Add("Content-Length", strconv.Itoa(int(length)))

}

func writeUseJson(writer http.ResponseWriter, obj interface{}) {

	if utils.HasFieldofStruct(obj, "XMLName") {
		writer.Header().Add(_CONTENT_TYPE, _CONTENT_XML)
		result, err := xml.Marshal(obj)
		if err == nil {
			writer.Write(result)
		}
	} else {
		writer.Header().Add(_CONTENT_TYPE, _CONTENT_JSON)
		//		result, err := json.Marshal(obj)
		result, err := toJson(obj)
		if err == nil {
			writer.Write(result)
		}
	}

}

func toJson(obj interface{}) ([]byte, error) {
	result := bytes.Buffer{}
	encoder := json.NewEncoder(&result)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(obj)
	return result.Bytes(), err

}

func writeUseTemplate(writer http.ResponseWriter, templateName, content string, obj interface{}) {
	tmpl, err := template.New(templateName).Parse(content)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(writer, obj)
	if err != nil {
		panic(err)
	}
}

func addParamsToForm(values url.Values, p Params) {
	if len(p) > 0 {
		for _, p1 := range p {
			values[p1.Key] = []string{p1.Value}
		}
	}
}
