package ipregion

import (
	"reflect"
	"testing"

	"github.com/lionsoul2014/ip2region/binding/golang/ip2region"
)

func TestIP2Region_Search(t *testing.T) {
	type fields struct {
		region *ip2region.Ip2Region
	}
	type args struct {
		ip string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *IPInfo
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &IP2Region{
				region: tt.fields.region,
			}
			got, err := r.Search(tt.args.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("IP2Region.Search() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("IP2Region.Search() = %v, want %v", got, tt.want)
			}
		})
	}
}
