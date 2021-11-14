package rql

import (
	"fmt"
	"testing"

	"github.com/jinzhu/gorm"
)

func TestMapping(t *testing.T) {
	tests := []struct {
		name    string
		conf    Config
		input   []byte
		wantErr bool
		wantOut *Params
	}{
		{
			name: "test simple mapping of column names",
			conf: Config{
				ColumnFn: func(columnName string) string {
					return fmt.Sprintf("person.%s", gorm.ToDBName(columnName))
				},
				Model: new(struct {
					Age     int    `rql:"filter, sort"`
					Name    string `rql:"filter"`
					Address string `rql:"filter"`
				}),
				DefaultSort:  []string{"+age"},
				DefaultLimit: 25,
			},
			input: []byte(`{
				"filter": {
					"name": "foo",
					"age": 12,
					"$or": [
						{ "address": "DC" },
						{ "address": "Marvel" }
					],
					"$and": [
						{ "age": { "$neq": 10} },
						{ "age": { "$neq": 20} },
						{ "$or": [{ "age": 11 }, {"age": 10}] }
					]
				}
			}`),
			wantOut: &Params{
				Limit:      25,
				FilterExp:  "person.name = ? AND person.age = ? AND (person.address = ? OR person.address = ?) AND (person.age <> ? AND person.age <> ? AND (person.age = ? OR person.age = ?))",
				FilterArgs: []interface{}{"foo", 12, "DC", "Marvel", 10, 20, 11, 10},
				Sort:       "person.age asc",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.conf.Log = t.Logf
			p, err := NewParser(tt.conf)
			if err != nil {
				t.Fatalf("failed to build parser: %v", err)
			}
			out, err := p.Parse(tt.input)
			if tt.wantErr != (err != nil) {
				t.Fatalf("want: %v\ngot:%v\nerr: %v", tt.wantErr, err != nil, err)
			}
			assertParams(t, out, tt.wantOut)
		})
	}
}
