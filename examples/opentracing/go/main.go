package main

import (
	"flag"
	"fmt"
	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	"io/ioutil"
	"log"
	"net/http"
)

var envMark, url string

func init() {
	flag.StringVar(&envMark, "envMark", "dev", "")
	flag.StringVar(&url, "url", "none", "")
}

func printOpenTracingText(w http.ResponseWriter, r *http.Request) {
	tracer, closer := jaeger.NewTracer("demo", jaeger.NewConstSampler(false), jaeger.NewNullReporter())
	defer closer.Close()
	ctx, err := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	var reqEnvMark, requestText string = "undefined", ""
	if err == nil {
		ctx.ForeachBaggageItem(func(k, v string) bool {
			if k == "envMark" {
				reqEnvMark = v
			}
			return false
		})
	}

	span := tracer.StartSpan("root", opentracing.ChildOf(ctx))

	hdr := opentracing.HTTPHeadersCarrier{}
	err = tracer.Inject(span.Context(), opentracing.HTTPHeaders, hdr)

	if url != "" && url != "none" {

		httpReq, _ := http.NewRequest("GET", url, nil)
		err = hdr.ForeachKey(func(key, val string) error {
			httpReq.Header.Add(key, val)
			return nil
		})
		resp, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			requestText = "call " + url + " failed"
		} else {
			defer resp.Body.Close()
			if err != nil {
				requestText = "call " + url + " failed"
			} else {
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					requestText = "call " + url + " failed"
				} else {
					requestText = string(body)
				}
			}
		}
	}

	fmt.Fprintf(w, requestText+"\n"+"[go][request env mark is "+reqEnvMark+"][my env mark is "+envMark+"]")
}

func main() {
	flag.Parse() //暂停获取参数

	log.Printf("envMark:" + envMark)
	log.Printf("url:" + url)

	http.HandleFunc("/demo", printOpenTracingText)
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
