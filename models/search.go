package models

import (
	"encoding/json"

	"github.com/mattbaird/elastigo/lib"
)

// Search is a search struct
type Search struct {
	Index         string
	Type          string
	Elasticsearch *elastigo.Conn
}

// Query searches elasticsearch
func (s *Search) Query(query string, result *EmberMultiData) error {

	out, err := s.Elasticsearch.Search(s.Index, s.Type, nil, query)
	if err != nil {
		return err
	}
	var results []*EmberDataObj
	for _, item := range out.Hits.Hits {
		var dat map[string]interface{}
		if err := json.Unmarshal(*item.Source, &dat); err != nil {
			return err
		}
		if val, ok := dat["Id"]; ok {
			res := &EmberDataObj{
				Type:       "organization",
				ID:         val.(string),
				Attributes: *item.Source,
			}
			results = append(results, res)
		}

	}
	result.Data = results
	return nil
}
