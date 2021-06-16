package compute

import (
	"testing"
)

func TestRawNumberFormat(t *testing.T) {
	type args struct {
		metricValue string
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "109",
			args: args{metricValue: "109"},
			want: 109,
		}, {
			name: "10.39K",
			args: args{metricValue: "10.39K"},
			want: 10390,
		}, {
			name: "145.78K",
			args: args{metricValue: "145.78K"},
			want: 145780,
		}, {
			name: "1M",
			args: args{metricValue: "1M"},
			want: 1000000,
		}, {
			name: "98.87M",
			args: args{metricValue: "98.87M"},
			want: 98870000,
		}, {
			name: "786.96M",
			args: args{metricValue: "786.96M"},
			want: 786960000,
		}, {
			name: "1B",
			args: args{metricValue: "1.00B"},
			want: 1000000000,
		}, {
			name: "12.65B",
			args: args{metricValue: "12.65B"},
			want: 12650000000,
		}, {
			name: "249.06B",
			args: args{metricValue: "249.06B"},
			want: 249060000000,
		}, {
			name: "1 trillion",
			args: args{metricValue: "1T"},
			want: 1000000000000,
		}, {
			name: "653.57 Trillion",
			args: args{metricValue: "653.57T"},
			want: 653570000000000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RawNumberFormat(tt.args.metricValue); got != tt.want {
				t.Errorf("RawNumberFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}
