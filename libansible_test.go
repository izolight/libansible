package libansible

import (
	"reflect"
	"testing"
)

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

func TestState_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		s       State
		want    []byte
		wantErr bool
	}{
		{"present", true, []byte("\"present\""), false},
		{"absent", false, []byte("\"absent\""), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("State.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("State.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestState_UnmarshalJSON(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		s       *State
		args    args
		wantErr bool
	}{
		{"present", new(State), args{[]byte("\"present\"")}, false},
		{"absent", new(State), args{[]byte("\"absent\"")}, false},
		{"invalid input", new(State), args{[]byte("\"gugus\"")}, true},
		{"no input", new(State), args{[]byte("")}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.UnmarshalJSON(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("State.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
