package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

const (
	msgFmt           = "==== %s ====\n"
	uploadServerPath = "uploads"
)

func getValueFromParams(data map[string]interface{}, key string) interface{} {
	keys := strings.Split(key, ".")
	for _, k := range keys {
		if v, ok := data[k].(map[string]interface{}); ok {
			data = v
		} else {
			return data[k]
		}
	}
	return nil
}

func stringToFloat32Slice(str string) ([]float32, error) {
	str1 := strings.Replace(str, "[", "", -1)
	str2 := strings.Replace(str1, "]", "", -1)
	var result []float32
	parts := strings.Split(str2, " ")
	for _, part := range parts {
		if f64, err := strconv.ParseFloat(part, 64); err == nil {
			result = append(result, float32(f64))
		} else {
			return nil, err
		}
	}
	return result, nil
}

type ParamImgInfo struct {
	Url    string `json:"url"`
	Apikey string `json:"api_key"`
}
type ParamTextInfo struct {
	Data   string `json:"data"`
	Apikey string `json:"api_key"`
}
type RespInfo struct {
	Embedding string `json:"embedding"`
}

func get_img_vec(embed_server_url string, url string, apikey string) ([]float32, error) {
	var paramBytes []byte = nil
	if strings.HasPrefix(url, uploadServerPath) {
		paramBytes, _ = json.Marshal(ParamImgInfo{Url: url, Apikey: apikey})
	} else {
		index := strings.Index(url, uploadServerPath)
		if index != -1 {
			url_fix := url[index+len(url):]
			paramBytes, _ = json.Marshal(ParamImgInfo{Url: url_fix, Apikey: apikey})
		} else {
			return []float32{0}, errors.New("url path err")
		}
	}
	resp, err := http.Post(embed_server_url+"/get_img_vec", "application/json", bytes.NewBuffer(paramBytes))
	if err != nil {
		return []float32{0}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []float32{0}, err
	}
	var respInfo RespInfo
	err = json.Unmarshal(body, &respInfo)
	if err != nil {
		return []float32{0}, err
	}

	vec, err := stringToFloat32Slice(respInfo.Embedding)
	if err != nil {
		return []float32{0}, err
	}
	return vec, nil
}

func get_text_vec(embed_server_url string, data string, apikey string) ([]float32, error) {
	paramBytes, err := json.Marshal(ParamTextInfo{Data: data, Apikey: apikey})
	resp, err := http.Post(embed_server_url+"/get_txt_vec", "application/json", bytes.NewBuffer(paramBytes))
	if err != nil {
		return []float32{0}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []float32{0}, err
	}
	var respInfo RespInfo
	err = json.Unmarshal(body, &respInfo)
	if err != nil {
		return []float32{0}, err
	}

	vec, err := stringToFloat32Slice(respInfo.Embedding)
	if err != nil {
		return []float32{0}, err
	}
	return vec, nil
}

func get_milvus_client(ctx context.Context, milvus_server string, milvus_port string, milvus_username string, milvus_pass string) (client.Client, error) {
	milvusAddr := milvus_server + `:` + milvus_port
	log.Printf(msgFmt, "start connecting to Milvus: "+milvusAddr)
	c, err := client.NewClient(ctx, client.Config{
		Address:  milvusAddr,
		Username: milvus_username,
		Password: milvus_pass,
		DialOptions: []grpc.DialOption{
			grpc.WithBlock(),
			grpc.WithTimeout(time.Duration((10000) * 1000 * 1000)),
		},
	})
	return c, err
}

func instanceCreate(gincontext *gin.Context) {
	var jsonParams map[string]interface{}
	if err := gincontext.BindJSON(&jsonParams); err != nil {
		gincontext.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	milvus_server := getValueFromParams(jsonParams, "milvus_server").(string)
	milvus_port := getValueFromParams(jsonParams, "milvus_port").(string)
	milvus_username := getValueFromParams(jsonParams, "milvus_username").(string)
	milvus_pass := getValueFromParams(jsonParams, "milvus_pass").(string)
	collection_name := getValueFromParams(jsonParams, "collection_name").(string)
	index_name := getValueFromParams(jsonParams, "index_name").(string)
	metric_type := getValueFromParams(jsonParams, "metric_type").(string)
	dimstr := getValueFromParams(jsonParams, "collection_dim").(string)
	dim, _ := strconv.ParseInt(dimstr, 10, 64)

	ctx := context.Background()
	c, err := get_milvus_client(ctx, milvus_server, milvus_port, milvus_username, milvus_pass)
	if err != nil {
		log.Println("failed to connect to milvus, err: ", err.Error())
		gincontext.JSON(http.StatusBadRequest, gin.H{"error": "get_milvus_client failed"})
		return
	} else {
		defer c.Close()
	}

	log.Printf(msgFmt, fmt.Sprintf("create collection, `%s`", collection_name))
	schema := entity.NewSchema().WithName(collection_name).WithDescription("milvus_image_search").
		WithField(entity.NewField().WithName("id").WithDataType(entity.FieldTypeInt64).WithIsPrimaryKey(true).WithIsAutoID(true)).
		WithField(entity.NewField().WithName("vec").WithDataType(entity.FieldTypeFloatVector).WithDim(dim)).
		WithField(entity.NewField().WithName("url").WithDataType(entity.FieldTypeVarChar).WithMaxLength(500))

	if err := c.CreateCollection(ctx, schema, entity.DefaultShardNumber); err != nil {
		log.Println("create collection failed, err: ", err.Error())
		gincontext.JSON(http.StatusBadRequest, gin.H{"create collection failed, err: ": err.Error()})
		return
	}

	if index_name == "HNSW" {
		log.Printf(msgFmt, "start creating index HNSW")
		idx, err := entity.NewIndexHNSW(entity.MetricType(metric_type), 12, 50)
		if err != nil {
			log.Println("failed to create HNSW index, err: ", err.Error())
			gincontext.JSON(http.StatusBadRequest, gin.H{"failed to create HNSW index, err:": err.Error()})
			return
		}
		if err := c.CreateIndex(ctx, collection_name, "vec", idx, false); err != nil {
			log.Println("failed to create index, err: ", err.Error())
			gincontext.JSON(http.StatusBadRequest, gin.H{"failed to create index, err:": err.Error()})
			return
		}
	} else if index_name == "IVF_FLAT" {
		log.Printf(msgFmt, "start creating index IVF_FLAT")
		idx, err := entity.NewIndexIvfFlat(entity.MetricType(metric_type), 12)
		if err != nil {
			log.Println("failed to create IVF_FLAT index, err: ", err.Error())
			gincontext.JSON(http.StatusBadRequest, gin.H{"failed to create IVF_FLAT index, err:": err.Error()})
			return
		}
		if err := c.CreateIndex(ctx, collection_name, "vec", idx, false); err != nil {
			log.Println("failed to create index, err: ", err.Error())
			gincontext.JSON(http.StatusBadRequest, gin.H{"failed to create index, err: ": err.Error()})
			return
		}
	} else if index_name == "IVF_SQ8" {
		log.Printf(msgFmt, "start creating index IVF_SQ8")
		idx, err := entity.NewIndexIvfSQ8(entity.MetricType(metric_type), 12)
		if err != nil {
			log.Println("failed to create IVF_SQ8 index, err: ", err.Error())
			gincontext.JSON(http.StatusBadRequest, gin.H{"failed to create IVF_SQ8 index, err: ": err.Error()})
			return
		}
		if err := c.CreateIndex(ctx, collection_name, "vec", idx, false); err != nil {
			log.Println("failed to create index, err: ", err.Error())
			gincontext.JSON(http.StatusBadRequest, gin.H{"failed to create index, err:": err.Error()})
			return
		}
	} else if index_name == "SCANN" {
		log.Printf(msgFmt, "start creating index SCANN")
		idx, err := entity.NewIndexSCANN(entity.MetricType(metric_type), 12, true)
		if err != nil {
			log.Println("failed to create SCANN index, err: ", err.Error())
			gincontext.JSON(http.StatusBadRequest, gin.H{"failed to create SCANN index, err: ": err.Error()})
			return
		}
		if err := c.CreateIndex(ctx, collection_name, "vec", idx, false); err != nil {
			log.Println("failed to create index, err: ", err.Error())
			gincontext.JSON(http.StatusBadRequest, gin.H{"failed to create index, err: ": err.Error()})
			return
		}
	}

	log.Printf(msgFmt, "start loading collection")
	err = c.LoadCollection(ctx, collection_name, false)
	if err != nil {
		gincontext.JSON(http.StatusBadRequest, gin.H{"failed to load collection, err: ": err.Error()})
		return
	}

	gincontext.JSON(http.StatusOK, gin.H{"message": "success"})
}

func uploadImageFiles(c *gin.Context) {
	form, err := c.MultipartForm()
	collectionName := c.PostForm("collectionName")

	savePath := uploadServerPath + "/" + collectionName
	err = os.MkdirAll(savePath, os.ModePerm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建目录失败"})
		return
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	files := form.File["files"]

	imgFileName := ""
	for _, file := range files {
		imgFileName = file.Filename
		dst := filepath.Join(savePath, file.Filename)
		src, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "打开文件失败"})
			return
		}
		defer src.Close()

		out, err := os.Create(dst)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建文件失败"})
			return
		}
		defer out.Close()

		_, err = io.Copy(out, src)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "复制文件失败"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "Files uploaded successfully", "url": savePath + "/" + imgFileName})
}

func onPicImport(gincontext *gin.Context) {
	var jsonParams map[string]interface{}
	if err := gincontext.BindJSON(&jsonParams); err != nil {
		gincontext.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	milvus_server := getValueFromParams(jsonParams, "milvus_server").(string)
	milvus_port := getValueFromParams(jsonParams, "milvus_port").(string)
	milvus_username := getValueFromParams(jsonParams, "milvus_username").(string)
	milvus_pass := getValueFromParams(jsonParams, "milvus_pass").(string)
	collection_name := getValueFromParams(jsonParams, "collection_name").(string)
	embed_server_url := getValueFromParams(jsonParams, "embed_server_url").(string)
	embed_server_apikey := getValueFromParams(jsonParams, "embed_server_apikey").(string)

	ctx := context.Background()
	c, err := get_milvus_client(ctx, milvus_server, milvus_port, milvus_username, milvus_pass)
	if err != nil {
		log.Println("failed to connect to milvus, err: ", err.Error())
		gincontext.JSON(http.StatusBadRequest, gin.H{"error": "get_milvus_client failed"})
		return
	} else {
		defer c.Close()
	}

	log.Printf(msgFmt, "start inserting images vectors")

	type Row struct {
		Vec []float32 `json:"vec" milvus:"name:vec"`
		Url string    `json:"url" milvus:"name:url"`
	}

	savePath := uploadServerPath + "/" + collection_name
	_, err = os.Stat(savePath)
	if err != nil {
		log.Println("failed to load images path, err: ", err.Error())
		gincontext.JSON(http.StatusBadRequest, gin.H{"failed to load images path, err: ": err.Error()})
		return
	}

	fileCount := 0
	entries, _ := os.ReadDir(savePath)
	for _, entry := range entries {
		if !entry.IsDir() {
			fileCount++
		}
	}
	rows := make([]interface{}, 0, fileCount)
	err = filepath.Walk(savePath, func(path string, resinfo os.FileInfo, errWalk error) error {
		if errWalk != nil {
			log.Println("遍历文件时出错, path=%s, err: ", path, errWalk.Error())
			return errWalk
		}
		if !resinfo.IsDir() {
			vec, err := get_img_vec(embed_server_url, path, embed_server_apikey)
			if err != nil {
				log.Println("get vector error, path="+path+", err: ", err.Error())
				return err
			}
			row := Row{
				Vec: vec,
				Url: path,
			}
			rows = append(rows, &row)
		}
		return nil
	})
	if err != nil {
		gincontext.JSON(http.StatusBadRequest, gin.H{"error": "get vector error"})
		return
	}

	_, errInsert := c.InsertRows(ctx, collection_name, "", rows)
	if errInsert != nil {
		log.Println("failed to insert rows: "+savePath, errInsert.Error())
		gincontext.JSON(http.StatusBadRequest, gin.H{"failed to insert rows: ": errInsert.Error()})
		return
	} else {
		log.Printf(msgFmt, "insert succeed: "+savePath)
	}
	gincontext.JSON(http.StatusOK, gin.H{"message": "insert successfully"})
}

type SearchRepos struct {
	Url      string  `json:"url"`
	Score    float32 `json:"score"`
	Filename string  `json:"filename"`
}

func picSearchByText(gincontext *gin.Context) {
	var jsonParams map[string]interface{}
	if err := gincontext.BindJSON(&jsonParams); err != nil {
		gincontext.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	milvus_server := getValueFromParams(jsonParams, "milvus_server").(string)
	milvus_port := getValueFromParams(jsonParams, "milvus_port").(string)
	milvus_username := getValueFromParams(jsonParams, "milvus_username").(string)
	milvus_pass := getValueFromParams(jsonParams, "milvus_pass").(string)
	collection_name := getValueFromParams(jsonParams, "collection_name").(string)
	index_name := getValueFromParams(jsonParams, "index_name").(string)
	metric_type := getValueFromParams(jsonParams, "metric_type").(string)
	embed_server_url := getValueFromParams(jsonParams, "embed_server_url").(string)
	embed_server_apikey := getValueFromParams(jsonParams, "embed_server_apikey").(string)
	search_text := getValueFromParams(jsonParams, "search_text").(string)
	search_topkstr := getValueFromParams(jsonParams, "search_topk").(string)
	search_topk, _ := strconv.Atoi(search_topkstr)

	ctx := context.Background()
	c, err := get_milvus_client(ctx, milvus_server, milvus_port, milvus_username, milvus_pass)
	if err != nil {
		log.Println("failed to connect to milvus, err: ", err.Error())
		gincontext.JSON(http.StatusBadRequest, gin.H{"error": "get_milvus_client failed"})
		return
	} else {
		defer c.Close()
	}

	log.Printf(msgFmt, "start searcching based on vector similarity")
	vecList := make([][]float32, 0, 1)
	var vec = []float32{0}

	log.Println("search by text: " + search_text + "==================")
	vec, err = get_text_vec(embed_server_url, search_text, embed_server_apikey)
	if err != nil {
		log.Println("failed to search, err: ", err.Error())
		gincontext.JSON(http.StatusBadRequest, gin.H{"failed to search, err: ": err.Error()})
		return
	}

	vecList = append(vecList, vec)
	log.Println("search_vec: ==================")
	for _, row := range vecList {
		fmt.Print("[")
		for i, value := range row {
			fmt.Print(value)
			if i != len(row)-1 {
				fmt.Print(", ")
			}
		}
		fmt.Println("]")
	}
	log.Println("==================")

	vec2search := []entity.Vector{
		entity.FloatVector(vecList[len(vecList)-1]),
	}
	var resdata []SearchRepos
	if index_name == "HNSW" {
		sp, _ := entity.NewIndexHNSWSearchParam(10)
		begin := time.Now()
		sRet, err := c.Search(ctx, collection_name, nil, "", []string{"url"}, vec2search,
			"vec", entity.MetricType(metric_type), search_topk, sp)
		end := time.Now()
		if err != nil {
			log.Println("failed to search collection, err: ", err.Error())
			gincontext.JSON(http.StatusBadRequest, gin.H{"failed to search, err: ": err.Error()})
			return
		}
		log.Println("results:")
		fmt.Println("url\tscore")
		for i := 0; i < search_topk; i++ {
			for _, res := range sRet {
				value1, _ := res.Fields.GetColumn("url").GetAsString(i)
				fmt.Print(value1)
				fmt.Print("\t")
				fmt.Print(res.Scores[i])
				fmt.Println()
				resdata = append(resdata, SearchRepos{Url: value1, Score: res.Scores[i], Filename: filepath.Base(value1)})
			}
		}
		log.Printf("\tsearch latency: %dms\n", end.Sub(begin)/time.Millisecond)
	} else if index_name == "IVF_FLAT" {
		sp, _ := entity.NewIndexIvfFlatSearchParam(10)
		begin := time.Now()
		sRet, err := c.Search(ctx, collection_name, nil, "", []string{"url"}, vec2search,
			"vec", entity.MetricType(metric_type), search_topk, sp)
		end := time.Now()
		if err != nil {
			log.Println("failed to search collection, err: ", err.Error())
			gincontext.JSON(http.StatusBadRequest, gin.H{"failed to search, err: ": err.Error()})
			return
		}

		log.Println("results:")
		fmt.Println("url\tscore")
		for i := 0; i < search_topk; i++ {
			for _, res := range sRet {
				value1, _ := res.Fields.GetColumn("url").GetAsString(i)
				fmt.Print(value1)
				fmt.Print("\t")
				fmt.Print(res.Scores[i])
				fmt.Println()
				resdata = append(resdata, SearchRepos{Url: value1, Score: res.Scores[i], Filename: filepath.Base(value1)})
			}
		}
		log.Printf("\tsearch latency: %dms\n", end.Sub(begin)/time.Millisecond)
	} else if index_name == "IVF_SQ8" {
		sp, _ := entity.NewIndexIvfSQ8SearchParam(10)
		begin := time.Now()
		sRet, err := c.Search(ctx, collection_name, nil, "", []string{"url"}, vec2search,
			"vec", entity.MetricType(metric_type), search_topk, sp)
		end := time.Now()
		if err != nil {
			log.Println("failed to search collection, err: ", err.Error())
			gincontext.JSON(http.StatusBadRequest, gin.H{"failed to search, err: ": err.Error()})
			return
		}

		log.Println("results:")
		fmt.Println("url\tscore")
		for i := 0; i < search_topk; i++ {
			for _, res := range sRet {
				value1, _ := res.Fields.GetColumn("url").GetAsString(i)
				fmt.Print(value1)
				fmt.Print("\t")
				fmt.Print(res.Scores[i])
				fmt.Println()
				resdata = append(resdata, SearchRepos{Url: value1, Score: res.Scores[i], Filename: filepath.Base(value1)})
			}
		}
		log.Printf("\tsearch latency: %dms\n", end.Sub(begin)/time.Millisecond)
	} else if index_name == "SCANN" {
		sp, _ := entity.NewIndexSCANNSearchParam(10, search_topk)
		begin := time.Now()
		sRet, err := c.Search(ctx, collection_name, nil, "", []string{"url"}, vec2search,
			"vec", entity.MetricType(metric_type), search_topk, sp)
		end := time.Now()
		if err != nil {
			log.Println("failed to search collection, err: ", err.Error())
			gincontext.JSON(http.StatusBadRequest, gin.H{"failed to search, err: ": err.Error()})
			return
		}

		log.Println("results:")
		fmt.Println("url\tscore")
		for i := 0; i < search_topk; i++ {
			for _, res := range sRet {
				value1, _ := res.Fields.GetColumn("url").GetAsString(i)
				fmt.Print(value1)
				fmt.Print("\t")
				fmt.Print(res.Scores[i])
				fmt.Println()
				resdata = append(resdata, SearchRepos{Url: value1, Score: res.Scores[i], Filename: filepath.Base(value1)})
			}
		}
		log.Printf("\tsearch latency: %dms\n", end.Sub(begin)/time.Millisecond)
	}
	resJsonData, _ := json.Marshal(resdata)
	gincontext.JSON(http.StatusOK, gin.H{"message": "search successfully", "data": string(resJsonData)})
}

func picSearchByImg(gincontext *gin.Context) {
	var jsonParams map[string]interface{}
	if err := gincontext.BindJSON(&jsonParams); err != nil {
		gincontext.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	milvus_server := getValueFromParams(jsonParams, "milvus_server").(string)
	milvus_port := getValueFromParams(jsonParams, "milvus_port").(string)
	milvus_username := getValueFromParams(jsonParams, "milvus_username").(string)
	milvus_pass := getValueFromParams(jsonParams, "milvus_pass").(string)
	collection_name := getValueFromParams(jsonParams, "collection_name").(string)
	index_name := getValueFromParams(jsonParams, "index_name").(string)
	metric_type := getValueFromParams(jsonParams, "metric_type").(string)
	embed_server_url := getValueFromParams(jsonParams, "embed_server_url").(string)
	embed_server_apikey := getValueFromParams(jsonParams, "embed_server_apikey").(string)
	search_img := getValueFromParams(jsonParams, "search_img").(string)
	search_topkstr := getValueFromParams(jsonParams, "search_topk").(string)
	search_topk, _ := strconv.Atoi(search_topkstr)

	ctx := context.Background()
	c, err := get_milvus_client(ctx, milvus_server, milvus_port, milvus_username, milvus_pass)
	if err != nil {
		log.Println("failed to connect to milvus, err: ", err.Error())
		gincontext.JSON(http.StatusBadRequest, gin.H{"error": "get_milvus_client failed"})
		return
	} else {
		defer c.Close()
	}

	log.Printf(msgFmt, "start searcching based on vector similarity")
	vecList := make([][]float32, 0, 1)
	var vec = []float32{0}

	log.Println("search by img: " + search_img + "==================")
	vec, err = get_img_vec(embed_server_url, uploadServerPath+"/"+collection_name+"/"+search_img, embed_server_apikey)
	if err != nil {
		log.Println("failed to get_img_vec, err: ", err.Error())
		gincontext.JSON(http.StatusBadRequest, gin.H{"failed to get_img_vec, err: ": err.Error()})
		return
	}

	vecList = append(vecList, vec)
	log.Println("search_vec: ==================")
	for _, row := range vecList {
		fmt.Print("[")
		for i, value := range row {
			fmt.Print(value)
			if i != len(row)-1 {
				fmt.Print(", ")
			}
		}
		fmt.Println("]")
	}
	log.Println("==================")

	vec2search := []entity.Vector{
		entity.FloatVector(vecList[len(vecList)-1]),
	}
	var resdata []SearchRepos
	if index_name == "HNSW" {
		sp, _ := entity.NewIndexHNSWSearchParam(10)
		begin := time.Now()
		sRet, err := c.Search(ctx, collection_name, nil, "", []string{"url"}, vec2search,
			"vec", entity.MetricType(metric_type), search_topk, sp)
		end := time.Now()
		if err != nil {
			log.Println("failed to search collection, err: ", err.Error())
			gincontext.JSON(http.StatusBadRequest, gin.H{"failed to search, err: ": err.Error()})
			return
		}
		log.Println("results:")
		fmt.Println("url\tscore")
		for i := 0; i < search_topk; i++ {
			for _, res := range sRet {
				value1, _ := res.Fields.GetColumn("url").GetAsString(i)
				fmt.Print(value1)
				fmt.Print("\t")
				fmt.Print(res.Scores[i])
				fmt.Println()
				resdata = append(resdata, SearchRepos{Url: value1, Score: res.Scores[i], Filename: filepath.Base(value1)})
			}
		}
		log.Printf("\tsearch latency: %dms\n", end.Sub(begin)/time.Millisecond)
	} else if index_name == "IVF_FLAT" {
		sp, _ := entity.NewIndexIvfFlatSearchParam(10)
		begin := time.Now()
		sRet, err := c.Search(ctx, collection_name, nil, "", []string{"url"}, vec2search,
			"vec", entity.MetricType(metric_type), search_topk, sp)
		end := time.Now()
		if err != nil {
			log.Println("failed to search collection, err: ", err.Error())
			gincontext.JSON(http.StatusBadRequest, gin.H{"failed to search, err: ": err.Error()})
			return
		}

		log.Println("results:")
		fmt.Println("url\tscore")
		for i := 0; i < search_topk; i++ {
			for _, res := range sRet {
				value1, _ := res.Fields.GetColumn("url").GetAsString(i)
				fmt.Print(value1)
				fmt.Print("\t")
				fmt.Print(res.Scores[i])
				fmt.Println()
				resdata = append(resdata, SearchRepos{Url: value1, Score: res.Scores[i], Filename: filepath.Base(value1)})
			}
		}
		log.Printf("\tsearch latency: %dms\n", end.Sub(begin)/time.Millisecond)
	} else if index_name == "IVF_SQ8" {
		sp, _ := entity.NewIndexIvfSQ8SearchParam(10)
		begin := time.Now()
		sRet, err := c.Search(ctx, collection_name, nil, "", []string{"url"}, vec2search,
			"vec", entity.MetricType(metric_type), search_topk, sp)
		end := time.Now()
		if err != nil {
			log.Println("failed to search collection, err: ", err.Error())
			gincontext.JSON(http.StatusBadRequest, gin.H{"failed to search, err: ": err.Error()})
			return
		}

		log.Println("results:")
		fmt.Println("url\tscore")
		for i := 0; i < search_topk; i++ {
			for _, res := range sRet {
				value1, _ := res.Fields.GetColumn("url").GetAsString(i)
				fmt.Print(value1)
				fmt.Print("\t")
				fmt.Print(res.Scores[i])
				fmt.Println()
				resdata = append(resdata, SearchRepos{Url: value1, Score: res.Scores[i], Filename: filepath.Base(value1)})
			}
		}
		log.Printf("\tsearch latency: %dms\n", end.Sub(begin)/time.Millisecond)
	} else if index_name == "SCANN" {
		sp, _ := entity.NewIndexSCANNSearchParam(10, search_topk)
		begin := time.Now()
		sRet, err := c.Search(ctx, collection_name, nil, "", []string{"url"}, vec2search,
			"vec", entity.MetricType(metric_type), search_topk, sp)
		end := time.Now()
		if err != nil {
			log.Println("failed to search collection, err: ", err.Error())
			gincontext.JSON(http.StatusBadRequest, gin.H{"failed to search, err: ": err.Error()})
			return
		}

		log.Println("results:")
		fmt.Println("url\tscore")
		for i := 0; i < search_topk; i++ {
			for _, res := range sRet {
				value1, _ := res.Fields.GetColumn("url").GetAsString(i)
				fmt.Print(value1)
				fmt.Print("\t")
				fmt.Print(res.Scores[i])
				fmt.Println()
				resdata = append(resdata, SearchRepos{Url: value1, Score: res.Scores[i], Filename: filepath.Base(value1)})
			}
		}
		log.Printf("\tsearch latency: %dms\n", end.Sub(begin)/time.Millisecond)
	}
	resJsonData, _ := json.Marshal(resdata)
	gincontext.JSON(http.StatusOK, gin.H{"message": "search successfully", "data": string(resJsonData)})
}

func instanceDelete(gincontext *gin.Context) {
	var jsonParams map[string]interface{}
	if err := gincontext.BindJSON(&jsonParams); err != nil {
		gincontext.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	milvus_server := getValueFromParams(jsonParams, "milvus_server").(string)
	milvus_port := getValueFromParams(jsonParams, "milvus_port").(string)
	milvus_username := getValueFromParams(jsonParams, "milvus_username").(string)
	milvus_pass := getValueFromParams(jsonParams, "milvus_pass").(string)
	collection_name := getValueFromParams(jsonParams, "collection_name").(string)

	ctx := context.Background()
	c, err := get_milvus_client(ctx, milvus_server, milvus_port, milvus_username, milvus_pass)
	if err != nil {
		log.Println("failed to connect to milvus, err: ", err.Error())
		gincontext.JSON(http.StatusBadRequest, gin.H{"error": "get_milvus_client failed"})
		return
	} else {
		defer c.Close()
	}

	has, err := c.HasCollection(ctx, collection_name)
	if err != nil {
		log.Println("failed to check collection exists, err: ", err.Error())
		gincontext.JSON(http.StatusBadRequest, gin.H{"failed to check collection exists: ": err.Error()})
		return
	}
	if has {
		c.DropCollection(ctx, collection_name)
		os.RemoveAll(uploadServerPath + "/" + collection_name)
	}
	gincontext.JSON(http.StatusOK, gin.H{"message": "success"})
}

func main() {
	serverport := flag.String("port", "8081", "port")
	flag.Parse()

	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}

	router.Use(cors.New(config))

	router.POST("/api/instanceCreate", instanceCreate)
	router.POST("/api/uploadImageFiles", uploadImageFiles)
	router.POST("/api/onPicImport", onPicImport)
	router.POST("/api/picSearchByText", picSearchByText)
	router.POST("/api/picSearchByImg", picSearchByImg)
	router.POST("/api/instanceDelete", instanceDelete)

	router.Static("/assets", "./dist/assets")
	router.Static(uploadServerPath, uploadServerPath)
	router.StaticFile("/favicon.ico", "./dist/favicon.ico")
	router.StaticFile("/", "./dist/index.html")

	router.Run(":" + *serverport)
}
