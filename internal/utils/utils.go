package utils

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"net"
	"net/url"
	"sort"
	"strconv"
	"strings"

	dao "github.com/tkeel-io/rule-manager/internal/dao"
	"github.com/tkeel-io/rule-manager/internal/dao/action_sink"
	daorequest "github.com/tkeel-io/rule-manager/internal/dao/utils"
	uuid "github.com/satori/go.uuid"
)

func DeepCopy(src, dst interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}

const (
	RuleIdPrefix   = "rule"
	ActionIdPrefix = "ac"
)

var templateID = "iot-%s-%s"

// 将[]string定义为MyStringList类型
type MyStringList []string

// 实现sort.Interface接口的获取元素数量方法
func (m MyStringList) Len() int {
	return len(m)
}

// 实现sort.Interface接口的比较元素方法
func (m MyStringList) Less(i, j int) bool {
	return m[i] < m[j]
}

// 实现sort.Interface接口的交换元素方法
func (m MyStringList) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

//sort slice.
func SortStringSlice(s []string) []string {
	ss := MyStringList(s)
	sort.Sort(ss)
	return []string(ss)
}

func GenerateID(ctx context.Context, userId, prefix string) (string, error) {

	baseLen := 14
	for {
		uuid := uuid.NewV4().String()
		uuid = strings.ReplaceAll(uuid, "-", "")
		for i := 0; i < len(uuid)-baseLen; i++ {
			subUuid := fmt.Sprintf(templateID, prefix, uuid[i:i+baseLen])

			//query.
			switch prefix {
			case RuleIdPrefix:
				rule := &dao.Rule{}
				rs, err := rule.Query(ctx, &daorequest.RuleQueryReq{
					Id:           &subUuid,
					UserId:       userId,
					FlagQueryBan: true,
				})
				if nil != err {
					return "", err
				}
				if len(rs) == 0 {
					return subUuid, nil
				}
			case ActionIdPrefix:
				action := &dao.Action{}
				rs, err := action.Query(ctx, &daorequest.ActionQueryReq{
					Id:           &subUuid,
					UserId:       userId,
					FlagQueryBan: true,
				})
				if nil != err {
					return "", err
				}
				if len(rs) == 0 {
					return subUuid, nil
				}
			default:
				//never.
			}
		}
	}

}

func ContainString(slice []string, elem string) bool {
	for _, e := range slice {
		if strings.HasPrefix(elem, e) {
			return true
		}
	}
	return false
}

func ContainFieldType(fieldTypes []action_sink.BaseFieldType, ele string) bool {
	lowerEle := strings.ToLower(ele)
	for _, fieldType := range fieldTypes {
		if strings.HasPrefix(lowerEle, strings.ToLower(fieldType.Name)) {
			return true
		}
	}
	return false
}

func GenerateUrlChronusDB(hosts []string, database string) []string {

	endpoints := make([]string, 0)
	for _, host := range hosts {

		// u := url.URL{
		// 	Scheme: "http",
		// 	//Opaque      string    // encoded opaque data
		// 	//User: url.UserPassword(user, passwd),
		// 	Host: host,
		// 	//Path        string    // path (relative paths may omit leading slash)
		// 	//RawPath     string    // encoded path hint (see EscapedPath method); added in Go 1.5
		// 	//ForceQuery  bool      // append a query ('?') even if RawQuery is empty; added in Go 1.7
		// 	//RawQuery    string    // encoded query values, without '?'
		// 	//Fragment    string    // fragment for references, without '#'
		// 	//RawFragment string    // encoded fragment hint (see EscapedFragment method); added in Go 1.15
		// }
		// q := u.Query()
		// q.Set("database", database)
		// u.RawQuery = q.Encode()
		// endpoints = append(endpoints, u.String())

		endpoints = append(endpoints, fmt.Sprintf("http://%s?database=%s", host, database))
	}
	return endpoints
}

func GenerateUrlsChronusDB(hosts []string, user, password, database string) []string {

	endpoints := make([]string, 0)
	for _, host := range hosts {

		u := url.URL{
			Scheme: "http",
			//Opaque      string    // encoded opaque data
			//User: url.UserPassword(user, passwd),
			Host: host,
			User: url.UserPassword(user, password),
			Path: url.PathEscape(database), //        string    // path (relative paths may omit leading slash)
			//RawPath     string    // encoded path hint (see EscapedPath method); added in Go 1.5
			//ForceQuery  bool      // append a query ('?') even if RawQuery is empty; added in Go 1.7
			//RawQuery    string    // encoded query values, without '?'
			//Fragment    string    // fragment for references, without '#'
			//RawFragment string    // encoded fragment hint (see EscapedFragment method); added in Go 1.15
		}
		//q := u.Query()
		//q.Set("database", database)
		//u.RawQuery = q.Encode()
		endpoints = append(endpoints, u.String())
	}
	return endpoints
}

func GenerateUrlKafka(host, user, passwd, topic string) string {

	return fmt.Sprintf("kafka://%s/%s/qingcloud", host, url.PathEscape(topic))
}

func GenerateUrlMysql(endpoints []string, user, passwd, db string) []string {
	urls := []string{}
	for _, endpoint := range endpoints {
		urls = append(urls, fmt.Sprintf("%s:%s@tcp(%s)/%s?%s", user, passwd, endpoint, db, "charset=utf8&parseTime=True&loc=Asia%2FShanghai"))
	}
	return urls
}

func GenerateUrlPostgresql(endpoints []string, user, passwd, db string) []string {
	urls := []string{}
	for _, endpoint := range endpoints {
		str := strings.Split(endpoint, ":")
		if len(str) != 2 {
			continue
		}
		urls = append(urls, fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", str[0], str[1], user, passwd, db))
	}
	return urls
}

func GenerateUrlRedis(endpoints []string, user, passwd, db string) []string {
	urls := []string{}
	for _, endpoint := range endpoints {
		url := url.URL{
			Scheme: "redis",
			Host:   endpoint,
			User:   url.UserPassword(user, passwd),
			Path:   url.PathEscape(db),
		}
		urls = append(urls, url.String())
	}
	return urls
}

func CheckHost(hosts []string) bool {
	for _, host := range hosts {
		p := strings.Split(host, ":")
		if len(p) != 2 {
			return false
		}
		//check ip
		if nil == net.ParseIP(p[0]) {
			return false
		}
		//check port
		if port, err := strconv.ParseInt(p[1], 10, 63); nil != err {
			return false
		} else if port >= 65535 {
			return false
		}
	}
	return true
}

func MapCat(m1, m2 map[string]interface{}) map[string]interface{} {

	if nil == m1 {
		m1 = make(map[string]interface{})
	}
	for key, value := range m2 {
		m1[key] = value
	}
	return m1
}
