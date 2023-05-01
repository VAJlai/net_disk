package handler

import (
	"encoding/json"
	"fmt"
	"gocloud/meta"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		//返回上传html页面
		data, err := os.ReadFile("./static/index.html")
		if err != nil {
			io.WriteString(w, "internel server error")
		}
		w.Write(data)
		//io.WriteString(w, string(data))
	} else if r.Method == "POST" {
		//接收文件流及存储到本地目录
		file, head, err := r.FormFile("file")
		defer file.Close()
		if err != nil {
			fmt.Println("Failed to get data,err:", err.Error())
			return
		}

		fileMeta := meta.FileMeta{
			FileName: head.Filename,
			Location: "D:/goapp/net_disk/tmp/" + head.Filename,
			UploadAt: time.Now().Format("2006-01-02 15:04:05"),
		}

		newFile, err := os.Create(fileMeta.Location)
		defer newFile.Close()
		if err != nil {
			fmt.Println("Failed to create file,err:", err.Error())
			return
		}
		fileMeta.FileSize, err = io.Copy(newFile, file)
		if err != nil {
			fmt.Println("Failed to save data into file,err:", err.Error())
			return
		}

		newFile.Seek(0, 0)
		//hash不方便测试，暂时用名字替代
		fileMeta.FileSha1 = fileMeta.FileName + "1"
		//fileMeta.FileSha1 = util.FileSha1(newFile)
		//meta.UpdateFileMeta(fileMeta)
		meta.UpdateFileMetaDB(fileMeta)

		http.Redirect(w, r, "/file/upload/suc", http.StatusFound)
	}
}

// 上传已完成
func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "喜来提示：文件已成功上传！Upload finished!")

}

// 获取文件信息
func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	filehash := r.Form["filehash"][0]
	//fMeta := meta.GetFileMeta(filehash)
	fMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(fMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// 获取所有文件信息
func FileQueryHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	countStr := r.Form["count"][0]
	count, err := strconv.Atoi(countStr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fMetaArray := meta.GetLastFileMetas(count)
	data, err := json.Marshal(fMetaArray)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// 文件下载
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fsha1 := r.Form.Get("filehash")
	fm := meta.GetFileMeta(fsha1)
	f, err := os.Open(fm.Location)
	defer f.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//文件小才可用此读取否则用流
	data, err := io.ReadAll(f)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment;filename=\""+fm.FileName+"\"")
	w.Write(data)
}

// 更新元信息接口(重命名)
func FileMetaUpdateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	opType := r.Form.Get("op")
	fileSha1 := r.Form.Get("filehash")
	newFileName := r.Form.Get("filename")

	if opType != "0" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	curFileMeta := meta.GetFileMeta(fileSha1)
	curFileMeta.FileName = newFileName
	meta.UpdateFileMeta(curFileMeta)

	data, err := json.Marshal(curFileMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func FileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fileSha1 := r.Form.Get("filehash")

	fMeta := meta.GetFileMeta(fileSha1)
	err := os.Remove(fMeta.Location)
	if err != nil {
		fmt.Println(fMeta.Location + "源文件删除失败")
	}
	meta.RemoveFileMeta(fileSha1)
	w.WriteHeader(http.StatusOK)
}
