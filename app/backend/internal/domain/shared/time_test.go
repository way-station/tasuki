package shared

import "testing"

func TestParsePace(t *testing.T) {
	tests := []struct {
		input   string
		want    int
		wantErr bool
	}{
		{input: "3:28", want: 208},
		{input: "03:28", want: 208},
		{input: "12:05", want: 725},
		{input: "0:00", wantErr: true},
		{input: "3:60", wantErr: true},
		{input: "3:5", wantErr: true},
		{input: "100:00", wantErr: true},
		{input: "", wantErr: true},
	}
	for _, testCase := range tests {
		t.Run(testCase.input, func(t *testing.T) {
			got, err := ParsePace(testCase.input)
			if testCase.wantErr {
				if err == nil {
					t.Fatalf("ParsePace(%q) = %v, want an error", testCase.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParsePace(%q): %v", testCase.input, err)
			}
			if got.Seconds() != testCase.want {
				t.Errorf("ParsePace(%q) = %d seconds, want %d", testCase.input, got.Seconds(), testCase.want)
			}
		})
	}
}

func TestPaceString(t *testing.T) {
	pace, err := PaceFromSeconds(208)
	if err != nil {
		t.Fatal(err)
	}
	if got := pace.String(); got != "3:28" {
		t.Errorf("String() = %q, want %q", got, "3:28")
	}
}

func TestTotalTime(t *testing.T) {
	got, err := ParseTotalTime("2:45:30")
	if err != nil {
		t.Fatal(err)
	}
	if got.Seconds() != 9930 {
		t.Errorf("Seconds() = %d, want 9930", got.Seconds())
	}
	if got.String() != "2:45:30" {
		t.Errorf("String() = %q, want %q", got.String(), "2:45:30")
	}
	for _, input := range []string{"0:00:00", "2:60:00", "45:30", ""} {
		if _, err := ParseTotalTime(input); err == nil {
			t.Errorf("ParseTotalTime(%q): want an error", input)
		}
	}
}
