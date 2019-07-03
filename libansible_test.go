package libansible

import "testing"

func TestString_UnmarshalJSON(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		s       *String
		args    args
		wantErr bool
	}{
		{"single string", new(String), args{[]byte("\"test\"")}, false},
		{"multiple strings", new(String), args{[]byte("[\"test\" ,\"test2\"]")}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.UnmarshalJSON(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("String.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
