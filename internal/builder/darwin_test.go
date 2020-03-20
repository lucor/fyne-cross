package builder

import "testing"

func TestDarwin_Output(t *testing.T) {
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
			want: "test",
		},
		{
			name: "386",
			fields: fields{
				opts: Options{
					Arch:   "386",
					Output: "test",
				},
			},
			want: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewDarwin(tt.fields.opts)
			if got := b.Output(); got != tt.want {
				t.Errorf("Darwin.Output() = %v, want %v", got, tt.want)
			}
		})
	}
}
