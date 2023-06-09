package cmd

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/guonaihong/gout"
	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/gjson"
	"hash/crc32"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type UploadToken struct {
	AccessKeyID     string `json:"AccessKeyID"`
	SecretAccessKey string `json:"SecretAccessKey"`
	SessionToken    string `json:"SessionToken"`
}

const (
	amzDateISO8601TimeFormat = "20060102T150405Z"
	shortTimeFormat          = "20060102"
	algorithm                = "AWS4-HMAC-SHA256"
	serviceName              = "imagex"
	serviceID                = "k3u1fbpfcp"
	version                  = "2018-08-01"
	uploadURLFormat          = "https://%s/%s"

	region = "cn-north-1"

	actionApplyImageUpload  = "ApplyImageUpload"
	actionCommitImageUpload = "CommitImageUpload"

	polynomialCRC32 = 0xEDB88320
)

var (
	newLine = []byte{'\n'}

	// if object matches reserved string, no need to encode them
	reservedObjectNames = regexp.MustCompile("^[a-zA-Z0-9-_.~/]+$")
)

type ImageX struct {
	AccessKey string
	SecretKey string
	Region    string
	Client    *http.Client

	Token   string
	Version string
	BaseURL string
}
type Result struct {
	Code int         `json:"err_no"`
	Msg  string      `json:"err_msg"`
	Data interface{} `json:"data"`
}

func JuejinUploadImage(imgPath string) (string, error) {
	uploadToken, err := GetUploadToken()
	if err != nil {
		return "", err
	}
	ix := &ImageX{
		AccessKey: uploadToken.AccessKeyID,
		SecretKey: uploadToken.SecretAccessKey,
		Token:     uploadToken.SessionToken,
		Region:    region,
	}
	//fmt.Printf("%+v\n", uploadToken)
	fmt.Printf("🔑 获得Token: %s \n", uploadToken.AccessKeyID)
	applyRes, err := ix.ApplyImageUpload()
	if err != nil {
		return "", err
	}
	storeInfo := gjson.Get(applyRes, "Result.UploadAddress.StoreInfos.0")
	storeURI := storeInfo.Get("StoreUri").String()
	storeAuth := storeInfo.Get("Auth").String()
	uploadHost := gjson.Get(applyRes, "Result.UploadAddress.UploadHosts.0").String()

	fmt.Printf("🚀 开始上传 : %s \n", storeURI)

	uploadURL := fmt.Sprintf(uploadURLFormat, uploadHost, storeURI)
	if err := ix.Upload(uploadURL, imgPath, storeAuth); err != nil {
		return "", err
	}

	sessionKey := gjson.Get(applyRes, "Result.UploadAddress.SessionKey").String()
	if _, err = ix.CommitImageUpload(sessionKey); err != nil {
		return "", err
	}

	return GetImageURL(storeURI)
}

func GetImageURL(uri string) (string, error) {
	rawurl := fmt.Sprintf("https://api.juejin.cn/imagex/get_img_url?uri=%s",
		uri)
	req, err := http.NewRequest(http.MethodGet, rawurl, nil)
	if err != nil {
		return "", err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	raw := string(b)
	if !gjson.Get(raw, "data.main_url").Exists() {
		return "", fmt.Errorf("获取URL出错: %s", raw)
	}
	main_url := strings.Trim(gjson.Get(raw, "data.main_url").String(), "?")
	fmt.Printf("🌐 获得 Url : %s \n", main_url)
	return main_url, nil
}
func (ix *ImageX) CommitImageUpload(sessionKey string) (string, error) {
	rawurl := fmt.Sprintf("https://imagex.bytedanceapi.com/?Action=%s&Version=%s&SessionKey=%s&ServiceId=%s",
		actionCommitImageUpload, version, sessionKey, serviceID)
	req, err := http.NewRequest(http.MethodPost, rawurl, nil)
	if err != nil {
		return "", err
	}

	if err := ix.signRequest(req); err != nil {
		return "", err
	}

	res, err := ix.getClient().Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	raw := string(b)
	if res.StatusCode != 200 || gjson.Get(raw, "ResponseMetadata.Error").Exists() {
		return "", fmt.Errorf("raw: %s, response: %+v", raw, res)
	}
	return raw, nil
}
func (ix *ImageX) Upload(rawurl, fp, auth string) error {
	crc32, err := hashFileCRC32(fp)
	if err != nil {
		return err
	}
	file, err := os.Open(fp)
	if err != nil {
		return err
	}
	defer file.Close()

	req, err := http.NewRequest(http.MethodPost, rawurl, file)
	if err != nil {
		return err
	}
	req.Header.Add("authorization", auth)
	req.Header.Add("Content-Type", "application/octet-stream")
	req.Header.Add("content-crc32", crc32)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	raw := string(b)
	fmt.Printf("✅  完成上传 : hash（%s） \n", gjson.Get(raw, "payload.hash").String())
	if gjson.Get(raw, "success").Int() != 0 {
		return fmt.Errorf("raw: %s, response: %+v", raw, res)
	}
	return nil
}

// hashFileCRC32 generate CRC32 hash of a file
// Refer https://mrwaggel.be/post/generate-crc32-hash-of-a-file-in-golang-turorial/
func hashFileCRC32(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	tablePolynomial := crc32.MakeTable(polynomialCRC32)
	hash := crc32.New(tablePolynomial)
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
func (ix *ImageX) ApplyImageUpload() (string, error) {
	rawurl := fmt.Sprintf("https://imagex.bytedanceapi.com/?Action=%s&Version=%s&ServiceId=%s",
		actionApplyImageUpload, version, serviceID)
	req, err := http.NewRequest(http.MethodGet, rawurl, nil)
	if err != nil {
		return "", err
	}

	if err := ix.signRequest(req); err != nil {
		return "", err
	}

	res, err := ix.getClient().Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	raw := string(b)
	if res.StatusCode != 200 || gjson.Get(raw, "ResponseMetadata.Error").Exists() {
		return "", fmt.Errorf("raw: %s, response: %+v", raw, res)
	}
	return raw, nil
}
func (ix *ImageX) getClient() *http.Client {
	if ix.Client == nil {
		return http.DefaultClient
	}
	return ix.Client
}
func (ix *ImageX) signRequest(req *http.Request) error {
	t := time.Now().UTC()
	req.Header.Set("x-amz-date", t.Format(amzDateISO8601TimeFormat))

	req.Header.Set("x-amz-security-token", ix.Token)

	k := ix.signKeys(t)
	h := hmac.New(sha256.New, k)

	if err := ix.writeStringToSign(h, t, req); err != nil {
		return err
	}

	auth := bytes.NewBufferString(algorithm)
	auth.Write([]byte(" Credential=" + ix.AccessKey + "/" + ix.creds(t)))
	auth.Write([]byte{',', ' '})
	auth.Write([]byte("SignedHeaders="))
	writeHeaderList(auth, req)
	auth.Write([]byte{',', ' '})
	auth.Write([]byte("Signature=" + fmt.Sprintf("%x", h.Sum(nil))))

	req.Header.Set("authorization", auth.String())
	return nil
}
func (ix *ImageX) signKeys(t time.Time) []byte {
	h := makeHMac([]byte("AWS4"+ix.SecretKey), []byte(t.Format(shortTimeFormat)))
	h = makeHMac(h, []byte(ix.Region))
	h = makeHMac(h, []byte(serviceName))
	h = makeHMac(h, []byte("aws4_request"))
	return h
}
func makeHMac(key []byte, data []byte) []byte {
	hash := hmac.New(sha256.New, key)
	hash.Write(data)
	return hash.Sum(nil)
}
func (ix *ImageX) writeRequest(w io.Writer, r *http.Request) error {
	r.Header.Set("host", r.Host)

	w.Write([]byte(r.Method))
	w.Write(newLine)
	writeURI(w, r)
	w.Write(newLine)
	writeQuery(w, r)
	w.Write(newLine)
	writeHeader(w, r)
	w.Write(newLine)
	w.Write(newLine)
	writeHeaderList(w, r)
	w.Write(newLine)
	return writeBody(w, r)
}
func writeURI(w io.Writer, r *http.Request) {
	path := r.URL.RequestURI()
	if r.URL.RawQuery != "" {
		path = path[:len(path)-len(r.URL.RawQuery)-1]
	}
	slash := strings.HasSuffix(path, "/")
	path = filepath.Clean(path)
	if path != "/" && slash {
		path += "/"
	}
	w.Write([]byte(path))
}
func writeQuery(w io.Writer, r *http.Request) {
	var a []string
	for k, vs := range r.URL.Query() {
		k = url.QueryEscape(k)
		for _, v := range vs {
			if v == "" {
				a = append(a, k)
			} else {
				v = url.QueryEscape(v)
				a = append(a, k+"="+v)
			}
		}
	}
	sort.Strings(a)
	for i, s := range a {
		if i > 0 {
			w.Write([]byte{'&'})
		}
		w.Write([]byte(s))
	}
}

func writeHeader(w io.Writer, r *http.Request) {
	i, a := 0, make([]string, len(r.Header))
	for k, v := range r.Header {
		sort.Strings(v)
		a[i] = strings.ToLower(k) + ":" + strings.Join(v, ",")
		i++
	}
	sort.Strings(a)
	for i, s := range a {
		if i > 0 {
			w.Write(newLine)
		}
		io.WriteString(w, s)
	}
}

func writeHeaderList(w io.Writer, r *http.Request) {
	i, a := 0, make([]string, len(r.Header))
	for k := range r.Header {
		a[i] = strings.ToLower(k)
		i++
	}
	sort.Strings(a)
	for i, s := range a {
		if i > 0 {
			w.Write([]byte{';'})
		}
		w.Write([]byte(s))
	}
}

func writeBody(w io.Writer, r *http.Request) error {
	var (
		b   []byte
		err error
	)
	// If the payload is empty, use the empty string as the input to the SHA256 function
	// http://docs.amazonwebservices.com/general/latest/gr/sigv4-create-canonical-request.html
	if r.Body == nil {
		b = []byte("")
	} else {
		b, err = ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}
		r.Body = ioutil.NopCloser(bytes.NewBuffer(b))
	}

	h := sha256.New()
	h.Write(b)
	fmt.Fprintf(w, "%x", h.Sum(nil))
	return nil
}
func (ix *ImageX) writeStringToSign(w io.Writer, t time.Time, r *http.Request) error {
	w.Write([]byte(algorithm))
	w.Write(newLine)
	w.Write([]byte(t.Format(amzDateISO8601TimeFormat)))
	w.Write(newLine)

	w.Write([]byte(ix.creds(t)))
	w.Write(newLine)

	h := sha256.New()
	if err := ix.writeRequest(h, r); err != nil {
		return err
	}
	fmt.Fprintf(w, "%x", h.Sum(nil))
	return nil
}
func (ix *ImageX) creds(t time.Time) string {
	return t.Format(shortTimeFormat) + "/" + ix.Region + "/" + serviceName + "/aws4_request"
}

func GetUploadToken() (*UploadToken, error) {
	r := Result{}
	s := ""

	params := gout.H{"client": "web"}
	err := gout.GET("https://api.juejin.cn/imagex/gen_token").SetCookies(
		//设置cookie1
		&http.Cookie{
			Name:  "sessionid",
			Value: UploadOpts.Session,
		},
	).SetQuery(params).Debug(false).Callback(func(c *gout.Context) (err error) {

		switch c.Code {
		case 200: //http code为200时，服务端返回的是json 结构
			c.BindJSON(&r)
		default: //http code为404时，服务端返回是html 字符串
			c.BindBody(&s)
		}
		return nil

	}).Do()

	if err != nil {
		return nil, err
	}

	var token *UploadToken

	if r.Code != 0 {
		err := fmt.Errorf("获取Token出错: %s， 请检查`session`", r.Msg)
		return token, err
	}

	//fmt.Printf("%+v\n", r)

	err = mapstructure.Decode(r.Data.(map[string]interface{})["token"], &token)
	if err != nil {
		fmt.Println(err.Error())
	}
	//fmt.Printf("%+v\n", token)
	return token, err
}
