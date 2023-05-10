package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	dblayer "net_disk/db"
	"net_disk/meta"
	"net_disk/store/oss"
	"net_disk/util"
	"os"
	"strconv"
	"time"
)

// UploadHandler 上传文件
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		//返回上传html页面
		data, err := os.ReadFile("./static/view/upload.html")
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
		fileMeta.FileSha1 = util.FileSha1(newFile)
		//同时将文件写入ceph存储
		newFile.Seek(0, 0)
		//ceph存储
		//data, _ := io.ReadAll(newFile)
		//bucket := ceph.GetCephBucket("userfile")
		//cephPath := "/ceph/" + fileMeta.FileSha1
		//_ = bucket.Put(cephPath, data, "octet-stream", s3.PublicRead)
		//fileMeta.Location = cephPath
		//OSS存储
		ossPath := "oss/" + fileMeta.FileSha1
		err = oss.Bucket().PutObject(ossPath, newFile)
		if err != nil {
			fmt.Println(err.Error())
			w.Write([]byte("Upload Failed"))
			return
		}
		fileMeta.Location = ossPath
		_ = meta.UpdateFileMetaDB(fileMeta)
		//更新用户文件表记录
		r.ParseForm()
		username := r.Form.Get("username")
		suc := dblayer.OnUserFileUploadFinished(username, fileMeta.FileSha1, fileMeta.FileName, fileMeta.FileSize)
		if suc {
			http.Redirect(w, r, "/static/view/home.html", http.StatusFound)
		} else {
			w.Write([]byte("Upload Failed."))
		}
	}
}

// UploadSucHandler 上传已完成

func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "喜来提示：文件已成功上传！Upload finished!")

}

// GetFileMetaHandler 获取文件信息
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

// FileQueryHandler 查询批量的文件元信息
func FileQueryHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	limitCnt, _ := strconv.Atoi(r.Form.Get("limit"))
	username := r.Form.Get("username")
	userFiles, err := dblayer.QueryUserFileMetas(username, limitCnt)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(userFiles)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// DownloadHandler 文件下载
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

// FileMetaUpdateHandler 更新元信息接口(重命名)
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

// FileDeleteHandler 删除文件
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

// TryFastUploadHandler 秒传借口
func TryFastUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	//1.解析请求参数
	username := r.Form.Get("username")
	fileHash := r.Form.Get("filehash")
	filename := r.Form.Get("filename")
	filesize, err := strconv.Atoi(r.Form.Get("filesize"))
	//2.从文件表中查询相同hash的文件记录
	fileMeta, err := meta.GetFileMetaDB(fileHash)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
	//3.查不到记录则返回秒传失败
	if fileMeta == nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请访问普通上传接口",
		}
		w.Write(resp.JSONBytes())
		return
	}
	//4.上传过则将文件信息写入用户文件表，返回成功
	suc := dblayer.OnUserFileUploadFinished(username, fileHash, filename, int64(filesize))
	if suc {
		resp := util.RespMsg{
			Code: 0,
			Msg:  "秒传成功",
		}
		w.Write(resp.JSONBytes())
	} else {
		resp := util.RespMsg{
			Code: -2,
			Msg:  "秒传失败,请稍后重试",
		}
		w.Write(resp.JSONBytes())
		return
	}
}

// DownloadURLHandler 生成文件的下载地址
func DownloadURLHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	filehash := r.Form.Get("filehash")
	//从文件表查找记录
	row, _ := dblayer.GetFileMeta(filehash)

	//TODO:判断文件存在oss还是Ceph
	signedURL := oss.DownloadURL(row.FileAddr.String)
	w.Write([]byte(signedURL))
}
