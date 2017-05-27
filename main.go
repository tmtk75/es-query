package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"
)

var (
	Version     = ""
	VersionLong = ""
)

type QueryBuilder struct {
	Start, End time.Time
	Size       int
	Query      string
	Aggr       map[string]interface{}
}

func (r QueryBuilder) Build() io.Reader {
	t := template.Must(template.New("_").Parse(`
{
  "size": {{.size}},
  "query": {
    "bool": {
      "must": [
        {
          "query_string": {
            "analyze_wildcard": true,
    	    "query": "{{.query}}"
          }
        },
        {
          "range": {
            "@timestamp": {
              "gte": {{.gte}},
              "lte": {{.lte}},
              "format": "epoch_millis",
              "time_zone": "UTC"
            }
          }
        }
      ]
    }
  }
}`))
	b := bytes.NewBufferString("")
	t.Execute(b, map[string]interface{}{
		"size":  r.Size,
		"query": r.Query,
		"gte":   r.Start.UnixNano() / int64(time.Millisecond),
		"lte":   r.End.UnixNano() / int64(time.Millisecond),
	})
	if r.Aggr == nil {
		return b
	}

	var i map[string]interface{}
	err := json.Unmarshal(b.Bytes(), &i)
	if err != nil {
		log.Fatalf("%s", err)
	}

	i["aggs"] = r.Aggr
	by, err := json.Marshal(i)
	if err != nil {
		log.Fatalf("%s", err)
	}

	return bytes.NewBuffer(by)
}

type TPS struct {
	Interval string
}

func (a TPS) Build() map[string]interface{} {
	return map[string]interface{}{
		"tps": map[string]interface{}{
			"date_histogram": map[string]interface{}{
				"field":         "@timestamp",
				"interval":      a.Interval,
				"time_zone":     "UTC",
				"min_doc_count": 1,
			},
		},
	}
}

func main() {
	ed := time.Now()
	st := ed.Add(-time.Duration(15 * time.Minute))
	var (
		ver   = flag.Bool("version", false, "Print vesion")
		sr    = flag.String("start", st.Format(time.RFC3339), "")
		en    = flag.String("end", ed.Format(time.RFC3339), "")
		dr    = flag.Duration("span", time.Duration(0), "")
		qr    = flag.String("query", "*", "Elasticsearch query string")
		sz    = flag.Int("size", 3, "Document size")
		tp    = flag.String("interval", "", "Aggr interval e.g) 1s, 1m, 1h")
		ui    = flag.Bool("ui", true, "Show UI")
		width = flag.Int("width", 96, "UI width")
	)
	flag.Parse()
	if *ver {
		fmt.Println(Version)
		os.Exit(0)
	}

	start, err := time.Parse(time.RFC3339, *sr)
	if err != nil {
		log.Fatalln(err)
	}
	end, err := time.Parse(time.RFC3339, *en)
	if err != nil {
		log.Fatalln(err)
	}
	if *dr != 0 {
		if *dr > 0 {
			end = start.Add(*dr)
		} else {
			start = end.Add(*dr)
		}
	}

	esHost, f := os.LookupEnv("ES_HOST")
	if !f || esHost == "" {
		log.Fatalln("ES_HOST is empty")
	}

	prefix, f := os.LookupEnv("ES_INDEX_PREFIX")
	if !f || esHost == "" {
		log.Fatalln("ES_INDEX_PREFIX is empty")
	}

	client := new(http.Client)
	qb := QueryBuilder{Query: *qr, Start: start, End: end, Size: *sz}
	if *tp != "" {
		qb.Aggr = TPS{Interval: *tp}.Build()
	}
	q := qb.Build()

	in := indexNames(prefix, start, end)
	res, err := client.Post("https://"+esHost+"/"+in+"/_search", "application/json", q)
	defer res.Body.Close()
	if err != nil {
		log.Fatalln(err)
	}

	//fmt.Println(res)
	a, _ := ioutil.ReadAll(res.Body)
	if *ui {
		opts := &UIOpts{Width: *width}
		opts.Render(string(a), *qr, start, end)
	} else {
		fmt.Println(string(a))
	}
}

func indexNames(prefix string, start, end time.Time) string {
	y, w := start.ISOWeek()
	p, q := end.ISOWeek()
	s := make([]string, 0)
	for i := y; i <= p; i++ {
		for j := w; j <= q; j++ {
			s = append(s, fmt.Sprintf("%s-%04d%02d", prefix, i, j))
		}
	}
	return strings.Join(s, ",")
}
