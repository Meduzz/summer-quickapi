package qa

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Meduzz/helper/http/client"
	"github.com/Meduzz/helper/http/herror"
	"github.com/Meduzz/summer-quickapi/api"
)

type (
	QaHttpClient struct {
		base   string
		entity string
	}
)

var (
	preload = "preload"
	sort    = "sort"
	where   = "where"
)

func NewHttpClient(base, entity string) *QaHttpClient {
	return &QaHttpClient{
		base:   base,
		entity: entity,
	}
}

func (q *QaHttpClient) Create(data *api.Create) (any, error) {
	req, err := client.POST(q.path(""), data.Entity)

	if err != nil {
		return nil, err
	}

	res, err := req.DoDefault()

	if err != nil {
		return nil, err
	}

	err = herror.IsError(res.Code())

	if err != nil {
		return nil, err
	}

	bs, err := res.AsBytes()

	if err != nil {
		return nil, err
	}

	return json.RawMessage(bs), nil
}

func (q *QaHttpClient) Read(data *api.Read) (any, error) {
	url := q.path(data.ID)
	preload := mapToQuery(preload, data.Preload)

	if preload != "" {
		url = fmt.Sprintf("%s?%s", url, preload)
	}

	req, err := client.GET(url)

	if err != nil {
		return nil, err
	}

	res, err := req.DoDefault()

	if err != nil {
		return nil, err
	}

	err = herror.IsError(res.Code())

	if err != nil {
		return nil, err
	}

	bs, err := res.AsBytes()

	if err != nil {
		return nil, err
	}

	return json.RawMessage(bs), nil
}

func (q *QaHttpClient) Update(data *api.Update) (any, error) {
	req, err := client.PUT(q.path(data.ID), data.Entity)

	if err != nil {
		return nil, err
	}

	res, err := req.DoDefault()

	if err != nil {
		return nil, err
	}

	err = herror.IsError(res.Code())

	if err != nil {
		return nil, err
	}

	bs, err := res.AsBytes()

	if err != nil {
		return nil, err
	}

	return json.RawMessage(bs), nil
}

func (q *QaHttpClient) Delete(data *api.Delete) error {
	req, err := client.DELETE(q.path(data.ID), nil)

	if err != nil {
		return err
	}

	res, err := req.DoDefault()

	if err != nil {
		return err
	}

	err = herror.IsError(res.Code())

	if err != nil {
		return err
	}

	return nil
}

func (q *QaHttpClient) Search(data *api.Search) (any, error) {
	if data.Take == 0 {
		data.Take = 25
	}

	where := mapToQuery(where, data.Where)
	sort := mapToQuery(sort, data.Sort)
	preload := mapToQuery(preload, data.Preload)
	skip := fmt.Sprintf("skip=%d", data.Skip)
	take := fmt.Sprintf("take=%d", data.Take)

	query := []string{skip, take}

	if where != "" {
		query = append(query, where)
	}

	if sort != "" {
		query = append(query, sort)
	}

	if preload != "" {
		query = append(query, preload)
	}

	url := fmt.Sprintf("%s?%s", q.path(""), strings.Join(query, "&"))

	req, err := client.GET(url)

	if err != nil {
		return nil, err
	}

	res, err := req.DoDefault()

	if err != nil {
		return nil, err
	}

	err = herror.IsError(res.Code())

	if err != nil {
		return nil, err
	}

	bs, err := res.AsBytes()

	if err != nil {
		return nil, err
	}

	return json.RawMessage(bs), nil
}

func (q *QaHttpClient) Patch(data *api.Patch) (any, error) {
	url := q.path("")
	preload := mapToQuery(preload, data.Preload)

	if preload != "" {
		url = fmt.Sprintf("%s?%s", url, preload)
	}

	bs, err := json.Marshal(data.Data)

	if err != nil {
		return nil, err
	}

	req, err := client.NewRequest("PATCH", url, bs, "application/json")

	if err != nil {
		return nil, err
	}

	res, err := req.DoDefault()

	if err != nil {
		return nil, err
	}

	err = herror.IsError(res.Code())

	if err != nil {
		return nil, err
	}

	bs, err = res.AsBytes()

	if err != nil {
		return nil, err
	}

	return json.RawMessage(bs), nil
}

func (q *QaHttpClient) path(id string) string {
	if id != "" {
		return fmt.Sprintf("%s/%s/%s", q.base, q.entity, id)
	}

	return fmt.Sprintf("%s/%s/", q.base, q.entity)
}

func mapToQuery(key string, in map[string]string) string {
	pairs := make([]string, 0)
	for k, v := range in {
		pairs = append(pairs, fmt.Sprintf("%s[%s]=%s", key, k, v))
	}

	return strings.Join(pairs, "&")
}

func filtersToQuery(in map[string]map[string]string) string {
	pairs := make([]string, 0)

	for key, filters := range in {
		pairs = append(pairs, mapToQuery(key, filters))
	}

	return strings.Join(pairs, "&")
}
