package apis

import (
	"reflect"
	"testing"

	"github.com/joeyscat/qim"
	"github.com/joeyscat/qim/naming"
	"github.com/joeyscat/qim/services/router/conf"
)

func Test_selectIdc(t *testing.T) {
	testRegions := &conf.Region{
		Idcs:  []conf.IDC{{ID: "SH_ALI"}, {ID: "HZ_ALI"}, {ID: "SH_TENCENT"}},
		Slots: []byte{0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 2, 2, 2, 2},
	}
	type args struct {
		token  string
		region *conf.Region
	}
	tests := []struct {
		name string
		args args
		want *conf.IDC
	}{
		{
			"token1 hit SH_ALI",
			args{"token1", testRegions},
			&conf.IDC{ID: "SH_ALI"},
		},
		{
			"token2 hit HZ_ALI",
			args{"token2", testRegions},
			&conf.IDC{ID: "HZ_ALI"},
		},
		{
			"token3 hit SH_TENCENT",
			args{"token3", testRegions},
			&conf.IDC{ID: "SH_TENCENT"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := selectIdc(tt.args.token, tt.args.region); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("selectIdc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_selectGateways(t *testing.T) {
	gateways := []qim.ServiceRegistration{
		&naming.DefaultService{ID: "g1"},
		&naming.DefaultService{ID: "g2"},
		&naming.DefaultService{ID: "g3"},
		&naming.DefaultService{ID: "g4"},
		&naming.DefaultService{ID: "g5"},
		&naming.DefaultService{ID: "g6"},
	}
	type args struct {
		token    string
		gateways []qim.ServiceRegistration
		num      int
	}
	tests := []struct {
		name string
		args args
		want []qim.ServiceRegistration
	}{
		{
			"ok",
			args{"token1", gateways, 3},
			[]qim.ServiceRegistration{
				&naming.DefaultService{ID: "g4"},
				&naming.DefaultService{ID: "g5"},
				&naming.DefaultService{ID: "g6"},
			},
		},
		{
			"ok",
			args{"token2", gateways, 3},
			[]qim.ServiceRegistration{
				&naming.DefaultService{ID: "g6"},
				&naming.DefaultService{ID: "g1"},
				&naming.DefaultService{ID: "g2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := selectGateways(tt.args.token, tt.args.gateways, tt.args.num); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("selectGateways() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_hashcode(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"ok", args{"token1"}, 847786290},
		{"ok", args{"token2"}, 2877382792},
		{"ok", args{"token3"}, 3699789854},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hashcode(tt.args.s); got != tt.want {
				t.Errorf("hashcode() = %v, want %v", got, tt.want)
			}
		})
	}
}
