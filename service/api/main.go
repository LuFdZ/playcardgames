package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path"
	"playcards/service/api/enum"
	envinit "playcards/service/init"
	"playcards/utils/config"
	gcf "playcards/utils/config"
	"playcards/utils/errors"
	"playcards/utils/log"
	"strconv"
	"strings"
	"time"

	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/metadata"
	"github.com/micro/go-web"
	"github.com/rs/cors"
	"golang.org/x/net/context"
)

func main() {
	envinit.Init()
	log.Info("start %s", enum.APIServiceName)

	address := config.APIHost()
	service := web.NewService(
		web.Name(enum.APIServiceName),
		web.Address(address),
	)
	service.Init()

	c := cors.New(cors.Options{
		AllowedOrigins: gcf.ApiAllowedOrigins(),
		AllowedHeaders: []string{"*"},
		AllowedMethods: []string{"POST", "OPTIONS"},
		Debug:          envinit.Debug,
		MaxAge:         3600,
	})

	service.Handle("/ping", c.Handler(http.HandlerFunc(Ping)))
	service.Handle("/", c.Handler(http.HandlerFunc(API)))
	if err := service.Run(); err != nil {
		log.Err("%v", err)
	}
}

func Ping(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Debug("ping read error: %v", err)
		Response(w, nil, enum.ErrInvalidPing)
		return
	}

	var t float64
	err = json.Unmarshal(body, &t)
	if err != nil {
		log.Info("ping unmarshal error: %v", err)
		Response(w, nil, enum.ErrInvalidPing)
	}

	now := float64(time.Now().UnixNano()) / 1e9
	b, err := json.Marshal([]float64{t, now})
	Response(w, b, err)
}

// URL:  /{service}/{struct}/{function} -> namespace.service::struct.function
// EXMP: /user/UserSrv/Register -> gdc.srv.user::UserSrv.Register
func PathToReceiver(p string) (string, string) {
	ns := enum.Namespace

	p = path.Clean(p)
	p = strings.TrimPrefix(p, "/")
	parts := strings.Split(p, "/")

	// If we've got two or less parts
	// Use first part as service
	// Use all parts as method
	if len(parts) <= 2 {
		service := ns + "." + strings.Join(parts[:len(parts)-1], ".")
		method := strings.Title(strings.Join(parts, "."))
		return service, method
	}

	// Service is everything minus last two parts
	// Method is the last two parts
	service := ns + "." + strings.Join(parts[:len(parts)-2], ".")
	method := strings.Title(strings.Join(parts[len(parts)-2:], "."))
	return service, method
}

func RequestToContext(r *http.Request) context.Context {
	ctx := context.Background()
	md := make(metadata.Metadata)
	for k, v := range r.Header {
		md[k] = strings.Join(v, ",")
	}
	return metadata.NewContext(ctx, md)
}

func GetRPCArgs(r *http.Request) (interface{}, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Info("get args read all failed: %v", err)
		return nil, enum.ErrReadBodyFailed
	}

	var args interface{}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&args); err != nil {
		log.Info("unmarshal args failed: %v", err)
		return nil, enum.ErrDecodeArgsFailed
	}

	return args, nil
}

func API(w http.ResponseWriter, r *http.Request) {
	args, err := GetRPCArgs(r)
	if err != nil {
		Response(w, nil, err)
		return
	}

	service, method := PathToReceiver(r.URL.Path)
	ctx := RequestToContext(r)
	req := client.NewJsonRequest(service, method, args)

	var rpcrsp json.RawMessage
	err = client.Call(ctx, req, &rpcrsp)
	log.Debug("req: %v, rsp: %v, err: %v", req, string(rpcrsp), err)
	Response(w, rpcrsp, err)
}

func Response(w http.ResponseWriter, body json.RawMessage, err error) {
	if err != nil {
		ce := errors.Parse(err.Error())
		if !envinit.Debug {
			ce.Internal = ""
		}
		switch ce.Status {
		case 0:
			w.WriteHeader(500)
		default:
			w.WriteHeader(int(ce.Status))
		}
		w.Write([]byte(ce.Error()))
		return
	}

	b, _ := body.MarshalJSON()
	w.Header().Set("Content-Length", strconv.Itoa(len(b)))
	w.Write(b)
}
