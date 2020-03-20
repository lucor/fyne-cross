package builder

import "testing"

func TestWindows_Output(t *testing.T) {
	type fields struct {
		os   string
		opts Options
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "amd64",
			fields: fields{
				opts: Options{
					Arch:   "amd64",
					Output: "test",
				},
			},
			want: "test.exe",
		},
		{
			name: "386",
			fields: fields{
				opts: Options{
					Arch:   "386",
					Output: "test",
				},
			},
			want: "test.exe",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewWindows(tt.fields.opts)
			if got := b.Output(); got != tt.want {
				t.Errorf("Windows.Output() = %v, want %v", got, tt.want)
			}
		})
	}
}
