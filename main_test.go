package main

import "testing"

func TestGetConfig(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		want    *config
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := GetConfig()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetConfig() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetConfig() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("GetConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_main(t *testing.T) {
	tests := []struct {
		name string // description of this test case
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			main()
		})
	}
}
