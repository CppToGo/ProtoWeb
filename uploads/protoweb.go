package main

import (
	"net/http"
	"io"
	"log"
	"os"
	"io/ioutil"
	"html/template"
	"path"
	"runtime/debug"
)

const(
	UPLOAD_DIR = "./uploads"
	TEMPLATE_DIR = "./templates"
)

var templates = make(map[string]*template.Template)

func safeHandler(fn http.HandlerFunc) http.HandlerFunc{ //这样可以截获httphandler的panic
	return func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			if e, ok := recover().(error); ok{
				http.Error(writer , e.Error() , http.StatusInternalServerError)
				log.Println("WARN: panic in %v - %v ", fn, e)
				log.Println(string(debug.Stack()))
			}
		}()
		fn(writer,request)
	}
}

func checkError(err error){
	if err != nil {
		panic(err)
	}
}

func viewHnadler(w http.ResponseWriter , r *http.Request){
	imageId := r.FormValue("id")
	imagePath := UPLOAD_DIR + "/" + imageId
	if exists := isExists(imagePath) ; !exists {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-type", "image")
	http.ServeFile(w, r, imagePath)
}

func isExists(path string)bool{
	_, err := os.Stat(path)
	if err !=nil {
		return os.IsExist(err)
	}
	return true
}

func uploadHandler(w http.ResponseWriter , r * http.Request){
	if r.Method == "GET"{
		err := renderHtml(w, "upload.html", nil)
		checkError(err)
	}
	if r.Method == "POST"{
		f , h  , err := r.FormFile("image")
		checkError(err)
		filename := h.Filename
		defer f.Close()
		t , err := os.Create(UPLOAD_DIR + "/" + filename)
		checkError(err)
		defer t.Close()
		_, err = io.Copy(t, f)
		checkError(err)
		http.Redirect(w , r, "/view?id=" + filename, http.StatusFound)
	}
}


func listHandler(w http.ResponseWriter , r *http.Request){
	fileInfoArr , err := ioutil.ReadDir(UPLOAD_DIR)
	checkError(err)

	locals := make(map[string]interface{})
	images := []string{}
	//var listHtml string
	for _, fileInfo := range fileInfoArr{
		images = append(images, fileInfo.Name())
		//imgid := fileInfo.Name()
		//listHtml += "<li><a href=\"/view?id=" + imgid + "\"> imgid </a></li>"
	}
	locals["images"] = images

	err = renderHtml(w, "list.html" , locals)
	checkError(err)
	//io.WriteString(w , "<html><ol>" + listHtml + "</ol></html>")
}

func renderHtml(w http.ResponseWriter , tmpl string , locals map[string]interface{})(err error){
	return templates[tmpl].Execute(w ,locals)
}

func init(){
	fileInfoArr, err := ioutil.ReadDir(TEMPLATE_DIR)
	if err != nil {
		panic(err)
		return
	}
	var templateName , templatePath string
	for _ , fileInfo := range fileInfoArr{
		templateName = fileInfo.Name()
		if ext := path.Ext(templateName); ext != ".html"{
			continue
		}
		templatePath = TEMPLATE_DIR + "/" + templateName
		log.Println("Loading template: " , templatePath)

		t := template.Must(template.ParseFiles(templatePath))
		templates[templateName] = t
	}
}

func main(){
	http.HandleFunc("/", safeHandler(listHandler))
	http.HandleFunc("/view", safeHandler(viewHnadler))
	http.HandleFunc("/upload", safeHandler(uploadHandler) )
	err := http.ListenAndServe("localhost:8080",nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
}
