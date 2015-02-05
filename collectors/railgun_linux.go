package collectors

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	// "time"

	"github.com/oliveagle/go-collectors/datapoint"
	"github.com/oliveagle/go-collectors/metadata"
	"github.com/oliveagle/go-collectors/util"
)

var (
	rgListenRE = regexp.MustCompile(`^stats.listen\s+?=\s+?([0-9.:]+)`)
	rgURL      string
)

func parseRailURL() string {
	var config string
	var url string
	util.ReadCommand(func(line string) error {
		fields := strings.Fields(line)
		if len(fields) == 0 || !strings.Contains(fields[0], "rg-listener") {
			return nil
		}
		for i, s := range fields {
			if s == "-config" && len(fields) > i {
				config = fields[i+1]
			}
		}
		return nil
	}, "ps", "-e", "-o", "args")
	if config == "" {
		return config
	}
	util.ReadLine(config, func(s string) error {
		if m := rgListenRE.FindStringSubmatch(s); len(m) > 0 {
			url = "http://" + m[1]
		}
		return nil
	})
	return url
}

func enableRailgun() bool {
	rgURL = parseRailURL()
	return enableURL(rgURL)()
}

func c_railgun() (datapoint.MultiDataPoint, error) {
	var md datapoint.MultiDataPoint
	res, err := http.Get(rgURL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var r map[string]interface{}
	j := json.NewDecoder(res.Body)
	if err := j.Decode(&r); err != nil {
		return nil, err
	}
	for k, v := range r {
		if _, ok := v.(float64); ok {
			Add(&md, "railgun."+k, v, nil, metadata.Unknown, metadata.None, "")
		}
	}
	return md, nil
}
